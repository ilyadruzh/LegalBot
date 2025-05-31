package openrouter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client calls the OpenRouter API.
type Client struct {
	APIKey string
	HTTP   *http.Client
}

// New creates a new OpenRouter client using the provided API key. Timeout can be
// customized via the OPENROUTER_TIMEOUT environment variable (e.g. "20s").
// Default timeout is 15 seconds.
func New(apiKey string) *Client {
	timeout := 15 * time.Second
	if v := os.Getenv("OPENROUTER_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}
	return &Client{APIKey: apiKey, HTTP: &http.Client{Timeout: timeout}}
}

// WithTimeout allows customizing HTTP client timeout when creating a new
// client.
func WithTimeout(d time.Duration) func(*Client) {
	return func(c *Client) { c.HTTP.Timeout = d }
}

// NewWithOptions creates a new client and applies given options.
func NewWithOptions(apiKey string, opts ...func(*Client)) *Client {
	c := New(apiKey)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

var endpoint = "https://openrouter.ai/v1/chat/completions"

// ChatCompletion sends a prompt and returns the response.
func (c *Client) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(prompt))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("openrouter: status %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
