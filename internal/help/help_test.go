package help

import "testing"

func TestMessage(t *testing.T) {
	msg := Message()
	if len(msg) == 0 || msg[0] != 'A' {
		t.Errorf("unexpected message: %q", msg)
	}
}
