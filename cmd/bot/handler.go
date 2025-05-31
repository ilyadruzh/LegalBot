package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"

	"legalbot/internal/db"
)

// langPref stores user language preferences.
var langPref = map[int64]string{}

// handleLang changes the language preference for a chat.
func handleLang(chatID int64, lang string) {
	langPref[chatID] = lang
}

// langFor returns language preference or default "en".
func langFor(chatID int64) string {
	if l, ok := langPref[chatID]; ok {
		return l
	}
	return "en"
}

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type OpenRouterClient interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

type ResultSaver interface {
	SaveResult(ctx context.Context, chatID int64, data string) (int64, error)
}

type ResultFetcher interface {
	RecentResults(ctx context.Context, chatID int64, limit int) ([]db.Result, error)
}

type HistoryDeleter interface {
	DeleteHistory(ctx context.Context, chatID int64) error
}

var docsBaseURL = "https://example.com/docs"

// checkSecretToken validates the Telegram secret token header.
// It returns true if the header matches the expected token.
func checkSecretToken(r *http.Request, expected string, l *slog.Logger) bool {
	token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
		if l != nil {
			l.Warn("invalid secret token", "remote", r.RemoteAddr)
		}
		return false
	}
	return true
}

// handleClaim processes user claim: sends prompt to OpenRouter, saves the result and sends it back to Telegram.
func handleClaim(ctx context.Context, tg TelegramSender, or OpenRouterClient, repo ResultSaver, chatID int64, prompt string) error {
	resp, err := or.ChatCompletion(ctx, prompt)
	if err != nil {
		return err
	}
	if _, err := repo.SaveResult(ctx, chatID, resp); err != nil {
		return err
	}
	if err := tg.SendMessage(ctx, chatID, resp); err != nil {
		return err
	}
	return nil
}

// handleRecent sends links to recent documents for a chat.
func handleRecent(ctx context.Context, tg TelegramSender, repo ResultFetcher, chatID int64) error {
	res, err := repo.RecentResults(ctx, chatID, 5)
	if err != nil {
		return err
	}
	for _, r := range res {
		link := fmt.Sprintf("%s/%d", docsBaseURL, r.ID)
		if err := tg.SendMessage(ctx, chatID, link); err != nil {
			return err
		}
	}
	return nil
}

// handleDelete removes chat history.
func handleDelete(ctx context.Context, tg TelegramSender, repo HistoryDeleter, chatID int64) error {
	if err := repo.DeleteHistory(ctx, chatID); err != nil {
		return err
	}
	return tg.SendMessage(ctx, chatID, "history deleted")
}
