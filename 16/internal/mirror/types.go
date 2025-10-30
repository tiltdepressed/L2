package mirror

import (
	"net/url"
	"time"
)

type Config struct {
	BaseURL        *url.URL
	OutputDir      string
	MaxDepth       int
	Concurrency    int
	RequestTimeout time.Duration
	UserAgent      string
	RespectRobots  bool
	SameHostOnly   bool
}

type task struct {
	URL       *url.URL
	DepthLeft int
	Kind      ResourceKind // Page or Asset
	From      *url.URL     // откуда обнаружена (для диагностики)
}

type ResourceKind int

const (
	ResourcePage ResourceKind = iota
	ResourceAsset
)

type discoveredLink struct {
	URL  *url.URL
	Kind ResourceKind
}
