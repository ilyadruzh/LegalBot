package help

import (
	"strings"
	"testing"
)

func TestMessageEN(t *testing.T) {
	msg := Message("en")
	if !strings.HasPrefix(msg, "Available") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestMessageRU(t *testing.T) {
	msg := Message("ru")
	if !strings.HasPrefix(msg, "Доступные") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestMessageDefault(t *testing.T) {
	msg := Message("unknown")
	if !strings.HasPrefix(msg, "Available") {
		t.Errorf("fallback not used: %q", msg)
	}
}
