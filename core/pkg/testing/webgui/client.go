package webgui

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Client provides testing utilities for the Web GUI stack.
type Client struct {
	httpClient *http.Client
	baseURL    string
	timeout    time.Duration
}

// NewClient creates a new Web GUI test client.
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		timeout: 30 * time.Second,
	}
}

// WaitForReady waits for the service to be healthy.
func (c *Client) WaitForReady(ctx context.Context) error {
	deadline := time.Now().Add(c.timeout)
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/health", nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			// Retry
		}
	}
	return fmt.Errorf("service not ready after %v", c.timeout)
}

// Get performs a GET request and returns the response.
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.httpClient.Do(req)
}

// CheckHealth checks if the health endpoint returns 200.
func (c *Client) CheckHealth(ctx context.Context) error {
	resp, err := c.Get(ctx, "/api/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
