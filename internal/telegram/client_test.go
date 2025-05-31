package telegram

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	}))
	defer srv.Close()

	old := apiURL
	apiURL = srv.URL
	defer func() { apiURL = old }()

	c := New("TOKEN")
	if err := c.SendMessage(chatID, text); err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
}

func TestSendMessageError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusTeapot)
	}))
	defer srv.Close()

	old := apiURL
	apiURL = srv.URL
	defer func() { apiURL = old }()

	c := New("TOKEN")
	err := c.SendMessage(1, "hi")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error containing boom, got %v", err)
	}
}
