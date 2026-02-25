package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient   *http.Client
	retryCount   int
	retryBackoff time.Duration
	ticker       *time.Ticker
}

func New(timeout time.Duration, retryCount int, retryBackoff time.Duration, rateLimitPerSec int) *Client {
	if rateLimitPerSec <= 0 {
		rateLimitPerSec = 1
	}
	return &Client{
		httpClient:   &http.Client{Timeout: timeout},
		retryCount:   retryCount,
		retryBackoff: retryBackoff,
		ticker:       time.NewTicker(time.Second / time.Duration(rateLimitPerSec)),
	}
}

func (c *Client) Close() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

func (c *Client) Get(ctx context.Context, reqURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retryCount; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.ticker.C:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("User-Agent", "recru-net/1.0 (+https://local)")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.retryCount {
				time.Sleep(c.retryBackoff)
				continue
			}
			return nil, fmt.Errorf("http get: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt < c.retryCount {
				time.Sleep(c.retryBackoff)
				continue
			}
			return nil, fmt.Errorf("read body: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close body: %w", closeErr)
		}

		if resp.StatusCode >= 500 && attempt < c.retryCount {
			lastErr = fmt.Errorf("server status %d", resp.StatusCode)
			time.Sleep(c.retryBackoff)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		return body, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request failed")
}
