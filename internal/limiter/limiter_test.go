package limiter

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	now := time.Now()
	rl := New(10, time.Minute, WithNow(func() time.Time { return now }))
	for i := 0; i < 10; i++ {
		if !rl.Allow(1) {
			t.Fatalf("unexpected deny at %d", i)
		}
	}
	if rl.Allow(1) {
		t.Fatalf("expected deny after limit")
	}
	now = now.Add(time.Minute)
	if !rl.Allow(1) {
		t.Fatalf("expected allow after window")
	}
}
