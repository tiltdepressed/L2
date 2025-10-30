package mirror

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Crawler struct {
	cfg Config

	httpc *HttpClient

	robots *robotsManager

	base *url.URL

	visitedMu sync.Mutex
	visited   map[string]struct{}

	queue chan task
	wg    sync.WaitGroup
}

func NewCrawler(cfg Config) (*Crawler, error) {
	if cfg.BaseURL == nil || cfg.BaseURL.Scheme == "" || cfg.BaseURL.Host == "" {
		return nil, errors.New("BaseURL не задан")
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.MaxDepth < 0 {
		cfg.MaxDepth = 0
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 20_000_000_000
	}
	return &Crawler{
		cfg:     cfg,
		httpc:   NewHttpClient(cfg.RequestTimeout, cfg.UserAgent),
		robots:  newRobotsManager(),
		base:    cfg.BaseURL,
		visited: make(map[string]struct{}),
		queue:   make(chan task, cfg.Concurrency*4),
	}, nil
}

func (c *Crawler) Run(ctx context.Context) error {
	startKind := ResourcePage
	startTask := task{
		URL:       c.base,
		DepthLeft: c.cfg.MaxDepth,
		Kind:      startKind,
	}
	c.markVisited(startTask.URL)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.queue <- startTask
	}()

	// Старт воркеров
	for i := 0; i < c.cfg.Concurrency; i++ {
		c.wg.Add(1)
		go func(id int) {
			defer c.wg.Done()
			c.worker(ctx, id)
		}(i + 1)
	}

	// Ожидание завершения
	go func() {
		c.wg.Wait()
		close(c.queue) // чтобы воркеры исчерпали
	}()

	// Дожидаемся опустошения канала
	for range c.queue {
		// worker сам вернет задачи, мы здесь просто поддерживаем блокировку до завершения
	}
	return nil
}

func (c *Crawler) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-c.queue:
			if !ok {
				return
			}
			if err := c.processTask(ctx, t); err != nil {
				log.Printf("[worker %d] Ошибка обработки %s: %v", id, t.URL, err)
			}
		}
	}
}

func (c *Crawler) processTask(ctx context.Context, t task) error {
	// robots.txt
	if c.cfg.RespectRobots {
		if !c.robots.allowed(t.URL, func(robotsURL string) ([]byte, error) {
			res, err := c.httpc.GetText(ctx, robotsURL, 1<<20)
			if err != nil {
				return nil, err
			}
			return res.Body, nil
		}) {
			log.Printf("robots.txt запретил: %s", t.URL)
			return nil
		}
	}

	// same-host-only
	if c.cfg.SameHostOnly && !sameHost(c.base, t.URL) {
		return nil
	}

	res, err := c.httpc.Get(ctx, t.URL.String())
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}

	finalURL, err := url.Parse(res.FinalURL)
	if err == nil {
		t.URL = finalURL
	}

	// Вычисляем локальный путь (относительный к OutputDir)
	localPath := localPathForURL(t.URL)
	absPath := filepath.Join(c.cfg.OutputDir, localPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}

	ct := strings.ToLower(res.ContentType)

	// HTML
	if t.Kind == ResourcePage || strings.Contains(ct, "text/html") {
		rewritten, links, err := rewriteHTMLAndDiscover(t.URL, localPath, res.Body, c.cfg.SameHostOnly)
		if err != nil {
			// на всякий случай сохраняем оригинал
			if writeErr := os.WriteFile(absPath, res.Body, 0o644); writeErr != nil {
				log.Printf("Ошибка сохранения %s: %v", absPath, writeErr)
			}
			return err
		}
		if err := os.WriteFile(absPath, rewritten, 0o644); err != nil {
			return err
		}

		for _, dl := range links {
			c.enqueueIfNew(dl, t.DepthLeft)
		}
		return nil
	}

	// CSS: переписываем url(...)
	if strings.Contains(ct, "text/css") || strings.HasSuffix(strings.ToLower(t.URL.Path), ".css") {
		rewritten, cssLinks := rewriteCSSAndDiscover(t.URL, localPath, res.Body, c.cfg.SameHostOnly)
		if err := os.WriteFile(absPath, rewritten, 0o644); err != nil {
			return err
		}
		for _, dl := range cssLinks {
			c.enqueueIfNew(dl, t.DepthLeft)
		}
		return nil
	}

	// Бинарные ассеты: сохраняем как есть
	if err := os.WriteFile(absPath, res.Body, 0o644); err != nil {
		return err
	}
	return nil
}

func (c *Crawler) enqueueIfNew(dl discoveredLink, parentDepth int) {
	// Глубина: для страниц уменьшаем, для ассетов — нет
	depthLeft := parentDepth
	if dl.Kind == ResourcePage {
		depthLeft = parentDepth - 1
	}
	if depthLeft < 0 {
		return
	}
	// same-host-only
	if c.cfg.SameHostOnly && !sameHost(c.base, dl.URL) {
		return
	}
	// dedup
	if !c.markVisited(dl.URL) {
		return
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		select {
		case c.queue <- task{URL: dl.URL, DepthLeft: depthLeft, Kind: dl.Kind}:
		default:
			// если очередь заполнена, блокирующая отправка
			c.queue <- task{URL: dl.URL, DepthLeft: depthLeft, Kind: dl.Kind}
		}
	}()
}

func (c *Crawler) markVisited(u *url.URL) bool {
	key := urlKey(u)
	c.visitedMu.Lock()
	defer c.visitedMu.Unlock()
	if _, ok := c.visited[key]; ok {
		return false
	}
	c.visited[key] = struct{}{}
	return true
}
