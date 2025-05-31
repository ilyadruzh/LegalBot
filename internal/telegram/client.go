package telegram

import "net/http"

// Client is a minimal Telegram Bot API client.
type Client struct {
Token string
HTTP  *http.Client
}

// New creates a new client.
func New(token string) *Client {
return &Client{Token: token, HTTP: http.DefaultClient}
}

// SendMessage sends a text message.
func (c *Client) SendMessage(chatID int64, text string) error {
// TODO: implement actual API call
return nil
}
