package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Royal17x/lenta-parser/internal/config"
	"github.com/Royal17x/lenta-parser/internal/metrics"
)

type Client struct {
	httpClient  *http.Client
	proxies     []*url.URL
	cfg         *config.HTTPConfig
	baseHeaders map[string]string
}

func New(cfg *config.HTTPConfig) (*Client, error) {
	proxies, err := parseProxies(cfg.Proxies)
	if err != nil {
		return nil, fmt.Errorf("parse proxies: %w", err)
	}

	c := &Client{
		cfg:     cfg,
		proxies: proxies,
		baseHeaders: map[string]string{
			"User-Agent":            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Accept-Language":       "ru-RU,ru;q=0.9",
			"Accept":                "text/html,application/xhtml+xml,*/*",
			"client":                "angular_web_0.0.2",
			"x-delivery-mode":       "pickup",
			"x-domain":              "moscow",
			"x-platform":            "omniweb",
			"x-retail-brand":        "lo",
			"x-device-os":           "Web",
			"x-device-web-platform": "desktop_web",
		},
	}

	c.httpClient = c.buildHTTPClient(nil)
	return c, nil
}

func (c *Client) buildHTTPClient(proxy *url.URL) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{
		Timeout:   c.cfg.Timeout,
		Transport: transport,
	}
}

func (c *Client) Get(ctx context.Context, targetURL string, category string, cookies []*http.Cookie) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < c.cfg.RetryCount; attempt++ {
		if attempt > 0 {
			wait := c.backoffDuration(attempt)
			slog.Debug("retrying request", "url", targetURL, "attempt", attempt, "wait", wait)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		body, err := c.doRequest(ctx, targetURL, category, cookies)
		if err == nil {
			return body, nil
		}
		lastErr = err
		slog.Warn("request failed", "url", targetURL, "attempt", attempt, "error", err)
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", c.cfg.RetryCount, lastErr)
}

func (c *Client) doRequest(ctx context.Context, targetURL string, category string, cookies []*http.Cookie) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range c.baseHeaders {
		req.Header.Set(k, v)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	if len(c.proxies) > 0 {
		proxy := c.proxies[rand.Intn(len(c.proxies))]
		c.httpClient = c.buildHTTPClient(proxy)
	}

	metrics.ActiveRequests.Inc()
	start := time.Now()

	resp, err := c.httpClient.Do(req)

	metrics.ActiveRequests.Dec()
	metrics.RequestDuration.WithLabelValues(category).Observe(time.Since(start).Seconds())

	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	time.Sleep(c.cfg.RateLimitDelay)

	return body, nil
}

func (c *Client) backoffDuration(attempt int) time.Duration {
	base := float64(c.cfg.RetryWaitBase)
	exp := base * float64(int(1)<<attempt)
	jitter := rand.Float64() * base * 0.5
	return time.Duration(exp + jitter)
}

func parseProxies(raw string) ([]*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	result := make([]*url.URL, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy %q: %w", p, err)
		}
		result = append(result, u)
	}
	return result, nil
}
