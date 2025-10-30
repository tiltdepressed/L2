package mirror

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func normalize(u *url.URL) *url.URL {
	cp := *u
	cp.Fragment = ""
	// Lowercase host, scheme
	cp.Scheme = strings.ToLower(cp.Scheme)
	cp.Host = strings.ToLower(cp.Host)
	// Remove default ports
	host, port, hasPort := strings.Cut(cp.Host, ":")
	if hasPort {
		if (cp.Scheme == "http" && port == "80") || (cp.Scheme == "https" && port == "443") {
			cp.Host = host
		}
	}
	return &cp
}

func urlKey(u *url.URL) string {
	n := normalize(u)
	return n.String()
}

func sameHost(a, b *url.URL) bool {
	return strings.EqualFold(a.Hostname(), b.Hostname())
}

func sanitizeFilename(name string) string {
	// remove characters invalid on some FS
	invalid := regexp.MustCompile(`[<>:"\\|?*\x00-\x1F]`)
	name = invalid.ReplaceAllString(name, "_")
	// trim spaces and dots
	name = strings.Trim(name, " .")
	if name == "" {
		name = "file"
	}
	return name
}

func isHTMLLikeByPath(u *url.URL) bool {
	p := u.Path
	if strings.HasSuffix(p, "/") || p == "" {
		return true
	}
	ext := strings.ToLower(path.Ext(p))
	switch ext {
	case ".html", ".htm", ".xhtml":
		return true
	case "":
		return true
	default:
		return false
	}
}

// localPathForURL возвращает относительный к корню зеркала путь для данного URL.
// Стратегия маппинга: out/<host>/<path> и для "страниц" без расширения — index.html.
func localPathForURL(u *url.URL) string {
	hostDir := sanitizeFilename(strings.ToLower(u.Host))
	p := u.EscapedPath()
	if p == "" {
		p = "/"
	}
	// Добавим уникальность для query, чтобы не было коллизий
	var querySuffix string
	if u.RawQuery != "" {
		h := sha1.Sum([]byte(u.RawQuery))
		querySuffix = "_q_" + hex.EncodeToString(h[:4])
	}

	ext := strings.ToLower(path.Ext(p))
	isHTML := isHTMLLikeByPath(u)

	// Нормализуем базовый путь
	if strings.HasSuffix(p, "/") || p == "/" {
		if isHTML {
			p = path.Join(p, "index.html")
		} else {
			p = path.Join(p, "index")
		}
	} else if ext == "" {
		if isHTML {
			p = path.Join(p, "index.html")
		} else {
			// для ассетов без расширения оставляем как есть, но добавим .bin
			p = p + ".bin"
		}
	}

	// Вставим querySuffix перед расширением, если есть
	if querySuffix != "" {
		if ext == "" {
			p = p + querySuffix
		} else {
			base := strings.TrimSuffix(p, ext)
			p = base + querySuffix + ext
		}
	}

	// Санитизация компонентов
	parts := strings.Split(p, "/")
	for i, s := range parts {
		if s == "" {
			continue
		}
		parts[i] = sanitizeFilename(s)
	}
	p = strings.Join(parts, "/")
	return filepath.FromSlash(filepath.Join(hostDir, p))
}

func relativeLink(fromLocalPath, toLocalPath string) string {
	fromDir := filepath.Dir(fromLocalPath)
	rel, err := filepath.Rel(fromDir, toLocalPath)
	if err != nil {
		return toLocalPath
	}
	rel = filepath.ToSlash(rel)
	// Приведем "./file" к "file"
	rel = strings.TrimPrefix(rel, "./")
	return rel
}
