package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"legalbot/internal/db"
)

// langPrefs stores user language preferences using a mutex for safe concurrent access.
type langPrefs struct {
	mu sync.RWMutex
	m  map[int64]string
}

func (l *langPrefs) set(id int64, lang string) {
	l.mu.Lock()
	l.m[id] = lang
	l.mu.Unlock()
}

func (l *langPrefs) get(id int64) string {
	l.mu.RLock()
	v, ok := l.m[id]
	l.mu.RUnlock()
	if ok {
		return v
	}
	return "en"
}

var langPref = langPrefs{m: map[int64]string{}}

// handleLang changes the language preference for a chat.
func handleLang(chatID int64, lang string) {
	langPref.set(chatID, lang)
}

// langFor returns language preference or default "en".
func langFor(chatID int64) string {
	return langPref.get(chatID)
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

type RateLimiter interface {
	Allow(chatID int64) bool
}

// loadDocsBaseURL returns the documentation base URL using the DOCS_BASE_URL
// environment variable if set, otherwise falling back to the default value.
func loadDocsBaseURL() string {
	if v := os.Getenv("DOCS_BASE_URL"); v != "" {
		return v
	}
	return "https://example.com/docs"
}

var docsBaseURL = loadDocsBaseURL()

const temporaryErrorMsg = "temporary error, please try again later"

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
func handleClaim(ctx context.Context, tg TelegramSender, or OpenRouterClient, repo ResultSaver, limiter RateLimiter, chatID int64, prompt string) error {
	if len(prompt) > 8000 {
		return fmt.Errorf("message too long: %d characters", len(prompt))
	}
	if !limiter.Allow(chatID) {
		if err := tg.SendMessage(ctx, chatID, "rate limit exceeded, try again later"); err != nil {
			return err
		}
		return nil
	}
	resp, err := or.ChatCompletion(ctx, prompt)
	if err != nil {
		slog.Error("openrouter", "err", err)
		if sendErr := tg.SendMessage(ctx, chatID, temporaryErrorMsg); sendErr != nil {
			return sendErr
		}
		return nil
	}
	if _, err := repo.SaveResult(ctx, chatID, resp); err != nil {
		slog.Error("db save", "err", err)
		if sendErr := tg.SendMessage(ctx, chatID, temporaryErrorMsg); sendErr != nil {
			return sendErr
		}
		return nil
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
		slog.Error("db recent", "err", err)
		if sendErr := tg.SendMessage(ctx, chatID, temporaryErrorMsg); sendErr != nil {
			return sendErr
		}
		return nil
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
		slog.Error("db delete", "err", err)
		if sendErr := tg.SendMessage(ctx, chatID, temporaryErrorMsg); sendErr != nil {
			return sendErr
		}
		return nil
	}
	return tg.SendMessage(ctx, chatID, "history deleted")
}
