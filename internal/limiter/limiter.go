package limiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	users  map[int64][]time.Time
}

func New(limit int, window time.Duration, opts ...func(*RateLimiter)) *RateLimiter {
	rl := &RateLimiter{
		limit:  limit,
		window: window,
		now:    time.Now,
		users:  make(map[int64][]time.Time),
	}
	for _, opt := range opts {
		opt(rl)
	}
	return rl
}

func WithNow(f func() time.Time) func(*RateLimiter) {
	return func(rl *RateLimiter) { rl.now = f }
}

func (rl *RateLimiter) allow(t time.Time, user int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket := rl.users[user]
	cutoff := t.Add(-rl.window)
	i := 0
	for i < len(bucket) && !bucket[i].After(cutoff) {
		i++
	}
	bucket = bucket[i:]
	if len(bucket) >= rl.limit {
		rl.users[user] = bucket
		return false
	}
	bucket = append(bucket, t)
	rl.users[user] = bucket
	return true
}

func (rl *RateLimiter) Allow(user int64) bool {
	return rl.allow(rl.now(), user)
}
