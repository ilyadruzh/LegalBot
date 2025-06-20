package openrouter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Client calls the OpenRouter API.
type Client struct {
	APIKey   string
	Endpoint string
	HTTP     *http.Client
	Logger   *slog.Logger
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
	endpoint := defaultEndpoint
	if v := os.Getenv("OPENROUTER_ENDPOINT"); v != "" {
		endpoint = v
	}
	return &Client{APIKey: apiKey, Endpoint: endpoint, HTTP: &http.Client{Timeout: timeout}, Logger: slog.Default()}
}

// WithTimeout allows customizing HTTP client timeout when creating a new
// client.
func WithTimeout(d time.Duration) func(*Client) {
	return func(c *Client) { c.HTTP.Timeout = d }
}

// WithEndpoint allows customizing API endpoint when creating a new client.
func WithEndpoint(u string) func(*Client) {
	return func(c *Client) { c.Endpoint = u }
}

// WithLogger allows setting a custom logger when creating a new client.
func WithLogger(l *slog.Logger) func(*Client) {
	return func(c *Client) { c.Logger = l }
}

// NewWithOptions creates a new client and applies given options.
func NewWithOptions(apiKey string, opts ...func(*Client)) *Client {
	c := New(apiKey)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

var defaultEndpoint = "https://openrouter.ai/v1/chat/completions"

// ChatCompletion sends a prompt and returns the response.
func (c *Client) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewBufferString(prompt))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	if c.Logger != nil {
		if reqID, _ := ctx.Value("request_id").(string); reqID != "" {
			c.Logger.Info("openrouter request", "request_id", reqID)
		}
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("openrouter: status %d: %s", resp.StatusCode, string(body))
	}

	if c.Logger != nil {
		if reqID, _ := ctx.Value("request_id").(string); reqID != "" {
			c.Logger.Info("openrouter response", "request_id", reqID)
		}
	}

	return string(body), nil
}
