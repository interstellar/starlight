package limit

import (
	"sync"
)

// A ConcurrencyLimiter limits how many concurrent requests
// are allowed to happen per key.
type ConcurrencyLimiter struct {
	limit int

	mu     sync.Mutex
	counts map[string]int
}

func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		limit:  limit,
		counts: make(map[string]int),
	}
}

// Acquire consumes an available slot for key,
// reporting whether one was available. If acquired,
// the caller is responsible for releasing the slot
// when the request is complete.
func (c *ConcurrencyLimiter) Acquire(key string) (ok bool) {
	c.mu.Lock()
	v := c.counts[key]
	if v < c.limit {
		c.counts[key] = v + 1
		ok = true
	}
	c.mu.Unlock()
	return ok
}

// Release returns an acquired slot back to the limiter.
func (c *ConcurrencyLimiter) Release(key string) {
	c.mu.Lock()
	c.counts[key] -= 1
	c.mu.Unlock()
}
