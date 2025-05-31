package openrouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChatCompletionSuccess(t *testing.T) {
	wantBody := "{\"ok\":true}"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Write([]byte(wantBody))
	}))
	defer srv.Close()

	c := NewWithOptions("Bearer test", WithEndpoint(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	body, err := c.ChatCompletion(ctx, "{}")
	if err != nil {
		t.Fatalf("ChatCompletion returned error: %v", err)
	}
	if body != wantBody {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestChatCompletionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewWithOptions("key", WithEndpoint(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := c.ChatCompletion(ctx, "{}")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error containing 'boom', got %v", err)
	}
}
