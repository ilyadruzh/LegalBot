package main

import (
	"context"
	"errors"
	"testing"
)

type mockTelegram struct {
	chatID int64
	text   string
	err    error
}

func (m *mockTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	m.chatID = chatID
	m.text = text
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
	chatID int64
	data   string
	id     int64
	err    error
}

func (m *mockRepo) SaveResult(ctx context.Context, chatID int64, data string) (int64, error) {
	m.chatID = chatID
	m.data = data
	return m.id, m.err
}

func TestHandleClaimSuccess(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{resp: "ok"}
	repo := &mockRepo{id: 1}
	ctx := context.Background()
	if err := handleClaim(ctx, tg, or, repo, 123, "hi"); err != nil {
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
	if err := handleClaim(context.Background(), tg, or, repo, 1, "hi"); err == nil {
		t.Fatal("expected error")
	}
	if repo.data != "" || tg.text != "" {
		t.Errorf("unexpected calls")
	}
}

func TestHandleClaimRepoError(t *testing.T) {
	tg := &mockTelegram{}
	or := &mockOpenRouter{resp: "x"}
	repo := &mockRepo{err: errors.New("db")}
	if err := handleClaim(context.Background(), tg, or, repo, 1, "hi"); err == nil {
		t.Fatal("expected error")
	}
	if tg.text != "" {
		t.Errorf("telegram should not be called")
	}
}

func TestHandleClaimTelegramError(t *testing.T) {
	tg := &mockTelegram{err: errors.New("tg")}
	or := &mockOpenRouter{resp: "x"}
	repo := &mockRepo{}
	if err := handleClaim(context.Background(), tg, or, repo, 1, "hi"); err == nil {
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
