package openrouter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// Client calls the OpenRouter API.
type Client struct {
	APIKey string
}

// New creates a new OpenRouter client using the provided API key.
func New(apiKey string) *Client {
	return &Client{APIKey: apiKey}
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

	resp, err := http.DefaultClient.Do(req)
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
