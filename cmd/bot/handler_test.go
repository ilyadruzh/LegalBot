package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"log/slog"

	"legalbot/internal/db"
)

type mockTelegram struct {
	chatID   int64
	text     string
	messages []string
	err      error
}

func (m *mockTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	m.chatID = chatID
	m.text = text
	m.messages = append(m.messages, text)
	return m.err
}

type mockOpenRouter struct {
	prompt string
	resp   string
	err    error
}

func (m *mockOpenRouter) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	m.prompt = prompt
	return m.resp, m.err
}

type mockRepo struct {
	chatID  int64
	data    string
	id      int64
	results []db.Result
	err     error
}

type mockLimiter struct{ ok bool }

func (m *mockLimiter) Allow(id int64) bool { return m.ok }

func (m *mockRepo) SaveResult(ctx context.Context, chatID int64, data string) (int64, error) {
	m.chatID = chatID
	m.data = data
	return m.id, m.err
}

func (m *mockRepo) RecentResults(ctx context.Context, chatID int64, limit int) ([]db.Result, error) {
	m.chatID = chatID
	return m.results, m.err
}

func (m *mockRepo) DeleteHistory(ctx context.Context, chatID int64) error {
	m.chatID = chatID
	return m.err
}

func TestHandleClaimSuccess(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{resp: "ok"}
	repo := &mockRepo{id: 1}
	lim := &mockLimiter{ok: true}
	ctx := context.Background()
	if err := handleClaim(ctx, tg, or, repo, lim, 123, "hi"); err != nil {
		t.Fatal(err)
	}
	if or.prompt != "hi" {
		t.Errorf("unexpected prompt %s", or.prompt)
	}
	if repo.chatID != 123 || repo.data != "ok" {
		t.Errorf("repo got %d %s", repo.chatID, repo.data)
	}
	if tg.chatID != 123 || tg.text != "ok" {
		t.Errorf("telegram got %d %s", tg.chatID, tg.text)
	}
}

func TestHandleClaimOpenRouterError(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{err: errors.New("boom")}
	repo := &mockRepo{}
	lim := &mockLimiter{ok: true}
	if err := handleClaim(context.Background(), tg, or, repo, lim, 1, "hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.data != "" {
		t.Errorf("repo should not be called")
	}
	if tg.text != "temporary error, please try again later" {
		t.Errorf("unexpected message %s", tg.text)
	}
}

func TestHandleClaimRepoError(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{resp: "x"}
	repo := &mockRepo{err: errors.New("db")}
	lim := &mockLimiter{ok: true}
	if err := handleClaim(context.Background(), tg, or, repo, lim, 1, "hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.chatID != 1 || repo.data != "x" {
		t.Errorf("repo not called correctly")
	}
	if tg.text != "temporary error, please try again later" {
		t.Errorf("unexpected message %s", tg.text)
	}
}

func TestHandleClaimTelegramError(t *testing.T) {
	tg := &mockTelegram{err: errors.New("tg")}
	or := &mockOpenRouter{resp: "x"}
	repo := &mockRepo{}
	lim := &mockLimiter{ok: true}
	if err := handleClaim(context.Background(), tg, or, repo, lim, 1, "hi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleLang(t *testing.T) {
	handleLang(1, "ru")
	if langFor(1) != "ru" {
		t.Fatalf("expected ru, got %s", langFor(1))
	}
}

func TestLangForDefault(t *testing.T) {
	if langFor(99) != "en" {
		t.Fatalf("expected default en")
	}
}

func TestHandleRecent(t *testing.T) {
	tg := &mockTelegram{}
	repo := &mockRepo{results: []db.Result{{ID: 1}, {ID: 2}}}
	docsBaseURL = "http://d"
	if err := handleRecent(context.Background(), tg, repo, 10); err != nil {
		t.Fatal(err)
	}
	if len(tg.messages) != 2 {
		t.Fatalf("expected 2 messages")
	}
	if tg.messages[0] != "http://d/1" {
		t.Fatalf("unexpected message %s", tg.messages[0])
	}
}

func TestHandleRecentRepoError(t *testing.T) {
	tg := &mockTelegram{}
	repo := &mockRepo{err: errors.New("db")}
	if err := handleRecent(context.Background(), tg, repo, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tg.text != "temporary error, please try again later" {
		t.Errorf("unexpected message %s", tg.text)
	}
}

func TestHandleDelete(t *testing.T) {
	tg := &mockTelegram{}
	repo := &mockRepo{}
	if err := handleDelete(context.Background(), tg, repo, 20); err != nil {
		t.Fatal(err)
	}
	if tg.text != "history deleted" {
		t.Fatalf("unexpected text %s", tg.text)
	}
	if repo.chatID != 20 {
		t.Fatalf("repo not called")
	}
}

func TestHandleDeleteRepoError(t *testing.T) {
	tg := &mockTelegram{}
	repo := &mockRepo{err: errors.New("db")}
	if err := handleDelete(context.Background(), tg, repo, 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tg.text != "temporary error, please try again later" {
		t.Errorf("unexpected message %s", tg.text)
	}
}

func TestCheckSecretTokenMatch(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("X-Telegram-Bot-Api-Secret-Token", "a")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if !checkSecretToken(r, "a", logger) {
		t.Fatalf("expected token match")
	}
}

func TestCheckSecretTokenMismatch(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("X-Telegram-Bot-Api-Secret-Token", "bad")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if checkSecretToken(r, "good", logger) {
		t.Fatalf("expected mismatch")
	}
}

func TestHandleClaimTooLong(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{}
	repo := &mockRepo{}
	lim := &mockLimiter{ok: true}
	long := strings.Repeat("a", 8001)
	if err := handleClaim(context.Background(), tg, or, repo, lim, 1, long); err == nil {
		t.Fatalf("expected length error")
	}
}

func TestHandleClaimRateLimit(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{}
	repo := &mockRepo{}
	lim := &mockLimiter{ok: false}
	if err := handleClaim(context.Background(), tg, or, repo, lim, 1, "hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tg.text != "rate limit exceeded, try again later" {
		t.Fatalf("unexpected message %s", tg.text)
	}
	if or.prompt != "" {
		t.Fatalf("openrouter should not be called")
	}
}
