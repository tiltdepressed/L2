package mirror

import (
	"bytes"
	"net/url"
	"path"
	"strings"
	"sync"
)

// Минимальная поддержка robots.txt для User-agent: *.
// Реализована логика "самого длинного совпадения"; при равной длине Allow побеждает Disallow.
type robotRule struct {
	allow    bool
	prefix   string
	prefixLC string
}

type robotsHostRules struct {
	loaded bool
	rules  []robotRule
}

type robotsManager struct {
	mu    sync.Mutex
	cache map[string]*robotsHostRules
}

func newRobotsManager() *robotsManager {
	return &robotsManager{
		cache: make(map[string]*robotsHostRules),
	}
}

func (rm *robotsManager) allowed(u *url.URL, fetch fetchText) bool {
	host := strings.ToLower(u.Host)
	rm.mu.Lock()
	entry, ok := rm.cache[host]
	rm.mu.Unlock()

	if !ok || !entry.loaded {
		// загрузим/распарсим
		rules := rm.fetchAndParse(u, fetch)
		rm.mu.Lock()
		rm.cache[host] = &robotsHostRules{loaded: true, rules: rules}
		entry = rm.cache[host]
		rm.mu.Unlock()
	}
	return matchAllowed(u, entry.rules)
}

type fetchText func(robotsURL string) ([]byte, error)

func (rm *robotsManager) fetchAndParse(u *url.URL, fetch fetchText) []robotRule {
	rURL := *u
	rURL.Path = path.Join("/", "robots.txt")
	rURL.RawQuery = ""
	rURL.Fragment = ""
	body, err := fetch(rURL.String())
	if err != nil || len(body) == 0 {
		return nil
	}
	return parseRobots(body)
}

func parseRobots(b []byte) []robotRule {
	lines := bytes.Split(b, []byte{'\n'})
	var (
		activeUAStar bool
		rules        []robotRule
	)
	for _, ln := range lines {
		line := string(ln)
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.TrimSpace(val)
		switch key {
		case "user-agent":
			ua := strings.ToLower(val)
			// начинаем секцию для * если встречаем ее; если лучшие секции (конкретнее) тоже встретятся — в этой минимальной реализации мы используем только '*'
			activeUAStar = (ua == "*")
		case "allow":
			if activeUAStar {
				p := val
				rules = append(rules, robotRule{allow: true, prefix: p, prefixLC: strings.ToLower(p)})
			}
		case "disallow":
			if activeUAStar {
				p := val
				rules = append(rules, robotRule{allow: false, prefix: p, prefixLC: strings.ToLower(p)})
			}
		}
	}
	return rules
}

func matchAllowed(u *url.URL, rules []robotRule) bool {
	if len(rules) == 0 {
		return true
	}
	target := strings.ToLower(u.EscapedPath())
	var (
		bestLen   = -1
		bestAllow = true // по умолчанию allow
	)
	for _, r := range rules {
		p := r.prefixLC
		if p == "" {
			// Disallow:  (пусто) — означает разрешено всё; Allow:  — тоже.
			continue
		}
		if strings.HasPrefix(target, p) {
			if ll := len(p); ll > bestLen || (ll == bestLen && r.allow) {
				bestLen = ll
				bestAllow = r.allow
			}
		}
	}
	return bestAllow
}
