package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a minimal Telegram Bot API client.
type Client struct {
	Token string
	HTTP  *http.Client
}

// New creates a new client.
func New(token string) *Client {
	return &Client{Token: token, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

var apiURL = "https://api.telegram.org"

// SendMessage sends a text message.
func (c *Client) SendMessage(chatID int64, text string) error {
	u := fmt.Sprintf("%s/bot%s/sendMessage", apiURL, c.Token)
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("text", text)

	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram: status %d: %s", resp.StatusCode, string(b))
	}

	return nil
}
