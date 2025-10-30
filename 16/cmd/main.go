package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"site-mirror/internal/mirror"
)

func main() {
	var (
		rawURL        string
		outDir        string
		depth         int
		concurrency   int
		timeout       time.Duration
		userAgent     string
		respectRobots bool
		sameHostOnly  bool
	)

	flag.StringVar(&rawURL, "url", "", "Стартовый URL (обязательный)")
	flag.StringVar(&outDir, "out", "mirror_output", "Каталог, куда сохранить зеркало")
	flag.IntVar(&depth, "depth", 2, "Глубина рекурсии (кол-во уровней ссылок)")
	flag.IntVar(&concurrency, "concurrency", 8, "Макс. число одновременных загрузок")
	flag.DurationVar(&timeout, "timeout", 20*time.Second, "Таймаут одного HTTP-запроса")
	flag.StringVar(&userAgent, "user-agent", "site-mirror/1.0 (+https://example.local)", "User-Agent")
	flag.BoolVar(&respectRobots, "respect-robots", true, "Учитывать robots.txt")
	flag.BoolVar(&sameHostOnly, "same-host-only", true, "Скачивать только с того же хоста")
	flag.Parse()

	if rawURL == "" {
		fmt.Println("Укажите стартовый URL с помощью флага -url")
		flag.Usage()
		os.Exit(2)
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		log.Fatalf("Некорректный URL: %v", err)
	}

	absOut, err := filepath.Abs(outDir)
	if err != nil {
		log.Fatalf("Не удалось получить абсолютный путь к out: %v", err)
	}
	if err := os.MkdirAll(absOut, 0o755); err != nil {
		log.Fatalf("Не удалось создать каталог %s: %v", absOut, err)
	}

	cfg := mirror.Config{
		BaseURL:        u,
		OutputDir:      absOut,
		MaxDepth:       depth,
		Concurrency:    concurrency,
		RequestTimeout: timeout,
		UserAgent:      userAgent,
		RespectRobots:  respectRobots,
		SameHostOnly:   sameHostOnly,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c, err := mirror.NewCrawler(cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации: %v", err)
	}
	if err := c.Run(ctx); err != nil {
		log.Fatalf("Завершено с ошибкой: %v", err)
	}

	log.Println("Готово.")
}
