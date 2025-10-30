package mirror

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	client    *http.Client
	userAgent string
	timeout   time.Duration
}

func NewHttpClient(timeout time.Duration, userAgent string) *HttpClient {
	tr := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DisableCompression:  false,
		MaxIdleConnsPerHost: 8,
	}
	return &HttpClient{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		userAgent: userAgent,
		timeout:   timeout,
	}
}

type FetchResult struct {
	StatusCode  int
	FinalURL    string
	ContentType string
	Body        []byte
}

func (hc *HttpClient) Get(ctx context.Context, url string) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if hc.userAgent != "" {
		req.Header.Set("User-Agent", hc.userAgent)
	}
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Limit body to avoid memory explosion on huge files (e.g. 50MB)
	const maxBody = 50 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, err
	}

	ct := resp.Header.Get("Content-Type")
	return &FetchResult{
		StatusCode:  resp.StatusCode,
		FinalURL:    resp.Request.URL.String(),
		ContentType: ct,
		Body:        body,
	}, nil
}

func (hc *HttpClient) GetText(ctx context.Context, url string, max int64) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if hc.userAgent != "" {
		req.Header.Set("User-Agent", hc.userAgent)
	}
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if max <= 0 {
		max = 2 << 20
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, max))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	return &FetchResult{
		StatusCode:  resp.StatusCode,
		FinalURL:    resp.Request.URL.String(),
		ContentType: ct,
		Body:        body,
	}, nil
}
