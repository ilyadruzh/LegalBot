package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSendMessageSuccess(t *testing.T) {
	chatID := int64(123)
	text := "hi"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/botTOKEN/sendMessage" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if r.Form.Get("chat_id") != "123" || r.Form.Get("text") != text {
			t.Errorf("unexpected form: %v", r.Form)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	old := apiURL
	apiURL = srv.URL
	defer func() { apiURL = old }()

	c := New("TOKEN")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.SendMessage(ctx, chatID, text); err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
}

func TestSendMessageHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusTeapot)
	}))
	defer srv.Close()

	old := apiURL
	apiURL = srv.URL
	defer func() { apiURL = old }()

	c := New("TOKEN")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.SendMessage(ctx, 1, "hi")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error containing boom, got %v", err)
	}
}

func TestSendMessageAPIFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":false,"description":"fail"}`))
	}))
	defer srv.Close()

	old := apiURL
	apiURL = srv.URL
	defer func() { apiURL = old }()

	c := New("TOKEN")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.SendMessage(ctx, 1, "hi")
	if err == nil || !strings.Contains(err.Error(), "fail") {
		t.Fatalf("expected error containing fail, got %v", err)
	}
}
