package mirror

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// В Go нет бэкреференсов, поэтому три альтернативы:
// 1) url('...') 2) url("...") 3) url(unquoted)
var cssURLRe = regexp.MustCompile(`(?i)url\(\s*(?:'([^']*)'|"([^"]*)"|([^'")\s]+))\s*\)`)

func rewriteHTMLAndDiscover(baseURL *url.URL, localPathForBase string, htmlBytes []byte, sameHostOnly bool) ([]byte, []discoveredLink, error) {
	doc, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, nil, err
	}

	var found []discoveredLink
	visitNode := func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		switch strings.ToLower(n.Data) {
		case "a":
			// только ссылочные страницы
			for i := range n.Attr {
				if strings.EqualFold(n.Attr[i].Key, "href") {
					val := strings.TrimSpace(n.Attr[i].Val)
					abs, ok := resolveURL(baseURL, val)
					if !ok {
						continue
					}
					if sameHostOnly && !sameHost(baseURL, abs) {
						continue
					}
					// предполагаем HTML-страницу
					kind := ResourcePage
					found = append(found, discoveredLink{URL: abs, Kind: kind})
					toLocal := localPathForURL(abs)
					n.Attr[i].Val = relativeLink(localPathForBase, toLocal)
				}
			}
		case "link":
			var isStylesheet bool
			for _, a := range n.Attr {
				if strings.EqualFold(a.Key, "rel") && strings.Contains(strings.ToLower(a.Val), "stylesheet") {
					isStylesheet = true
					break
				}
			}
			for i := range n.Attr {
				if strings.EqualFold(n.Attr[i].Key, "href") {
					val := strings.TrimSpace(n.Attr[i].Val)
					abs, ok := resolveURL(baseURL, val)
					if !ok {
						continue
					}
					if sameHostOnly && !sameHost(baseURL, abs) {
						continue
					}
					kind := ResourceAsset
					if isStylesheet {
						kind = ResourceAsset
					}
					found = append(found, discoveredLink{URL: abs, Kind: kind})
					toLocal := localPathForURL(abs)
					n.Attr[i].Val = relativeLink(localPathForBase, toLocal)
				}
			}
		case "script", "img", "source", "video", "audio", "iframe":
			for i := range n.Attr {
				if strings.EqualFold(n.Attr[i].Key, "src") {
					val := strings.TrimSpace(n.Attr[i].Val)
					abs, ok := resolveURL(baseURL, val)
					if !ok {
						continue
					}
					if sameHostOnly && !sameHost(baseURL, abs) {
						continue
					}
					found = append(found, discoveredLink{URL: abs, Kind: ResourceAsset})
					toLocal := localPathForURL(abs)
					n.Attr[i].Val = relativeLink(localPathForBase, toLocal)
				}
				if strings.EqualFold(n.Attr[i].Key, "srcset") {
					// srcset может содержать несколько URL
					parts := strings.Split(n.Attr[i].Val, ",")
					var out []string
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p == "" {
							continue
						}
						sub := strings.Fields(p)
						if len(sub) == 0 {
							continue
						}
						link := sub[0]
						abs, ok := resolveURL(baseURL, link)
						if !ok {
							out = append(out, p)
							continue
						}
						if sameHostOnly && !sameHost(baseURL, abs) {
							out = append(out, p)
							continue
						}
						found = append(found, discoveredLink{URL: abs, Kind: ResourceAsset})
						toLocal := localPathForURL(abs)
						rel := relativeLink(localPathForBase, toLocal)
						if len(sub) > 1 {
							out = append(out, strings.Join(append([]string{rel}, sub[1:]...), " "))
						} else {
							out = append(out, rel)
						}
					}
					n.Attr[i].Val = strings.Join(out, ", ")
				}
			}
		}
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		visitNode(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), found, nil
}

func resolveURL(base *url.URL, ref string) (*url.URL, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "data:") || strings.HasPrefix(ref, "javascript:") || strings.HasPrefix(ref, "mailto:") || strings.HasPrefix(ref, "#") {
		return nil, false
	}
	u, err := url.Parse(ref)
	if err != nil {
		return nil, false
	}
	return base.ResolveReference(u), true
}

func rewriteCSSAndDiscover(baseURL *url.URL, localPathForBase string, css []byte, sameHostOnly bool) ([]byte, []discoveredLink) {
	rewritten := cssURLRe.ReplaceAllFunc(css, func(m []byte) []byte {
		sub := cssURLRe.FindSubmatch(m)
		// sub[0] — полное совпадение, дальше группы:
		// 1 — одинарные кавычки, 2 — двойные, 3 — без кавычек
		var raw, quote string
		if len(sub) >= 2 && len(sub[1]) > 0 {
			raw = string(sub[1])
			quote = "'"
		} else if len(sub) >= 3 && len(sub[2]) > 0 {
			raw = string(sub[2])
			quote = `"`
		} else if len(sub) >= 4 && len(sub[3]) > 0 {
			raw = string(sub[3])
			quote = ""
		} else {
			return m
		}
		abs, ok := resolveURL(baseURL, raw)
		if !ok {
			return m
		}
		if sameHostOnly && !sameHost(baseURL, abs) {
			return m
		}
		local := localPathForURL(abs)
		rel := relativeLink(localPathForBase, local)

		// Собрать обратно: сохраняем исходный тип кавычек
		if quote == "" {
			return []byte("url(" + rel + ")")
		}
		return []byte("url(" + quote + rel + quote + ")")
	})
	return rewritten, discoverCSSURLs(baseURL, rewritten, sameHostOnly)
}

func discoverCSSURLs(base *url.URL, css []byte, sameHostOnly bool) []discoveredLink {
	matches := cssURLRe.FindAllSubmatch(css, -1)
	var out []discoveredLink
	for _, m := range matches {
		var raw string
		if len(m) >= 2 && len(m[1]) > 0 {
			raw = string(m[1])
		} else if len(m) >= 3 && len(m[2]) > 0 {
			raw = string(m[2])
		} else if len(m) >= 4 && len(m[3]) > 0 {
			raw = string(m[3])
		} else {
			continue
		}
		abs, ok := resolveURL(base, raw)
		if !ok {
			continue
		}
		if sameHostOnly && !sameHost(base, abs) {
			continue
		}
		out = append(out, discoveredLink{URL: abs, Kind: ResourceAsset})
	}
	return out
}
