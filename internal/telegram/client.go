package telegram

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a minimal Telegram Bot API client.
type Client struct {
	Token  string
	HTTP   *http.Client
	Logger *slog.Logger
}

// New creates a new client.
func New(token string) *Client {
	return &Client{Token: token, HTTP: &http.Client{Timeout: 10 * time.Second}, Logger: slog.Default()}
}

// WithLogger sets a custom logger when creating a new client.
func WithLogger(l *slog.Logger) func(*Client) {
	return func(c *Client) { c.Logger = l }
}

var apiURL = "https://api.telegram.org"

// SendMessage sends a text message.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	u := fmt.Sprintf("%s/bot%s/sendMessage", apiURL, c.Token)
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("text", text)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if c.Logger != nil {
		c.Logger.Info("send telegram message", "chat_id", chatID)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram: status %d: %s", resp.StatusCode, string(b))
	}

	if c.Logger != nil {
		c.Logger.Info("telegram message sent", "chat_id", chatID)
	}

	return nil
}
