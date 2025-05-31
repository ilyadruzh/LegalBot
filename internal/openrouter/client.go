package openrouter

import "context"

// Client calls the OpenRouter API.
type Client struct{
    APIKey string
}

func New(apiKey string) *Client {
    return &Client{APIKey: apiKey}
}

// ChatCompletion sends a prompt and returns the response.
func (c *Client) ChatCompletion(ctx context.Context, prompt string) (string, error) {
    // TODO: implement HTTP call to OpenRouter
    return "", nil
}
