package cache

import (
	"sync"
	"time"
)

// CountsCache holds cached category counts with expiration
type CountsCache struct {
	counts    map[string]int64
	expiresAt time.Time
	mu        sync.RWMutex
	TTL       time.Duration
}

// NewCountsCache creates a new counts cache with the specified TTL
func NewCountsCache(ttl time.Duration) *CountsCache {
	return &CountsCache{
		counts:    make(map[string]int64),
		expiresAt: time.Now(),
		TTL:       ttl,
	}
}

// Get retrieves cached counts if not expired
func (c *CountsCache) Get() (map[string]int64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Now().Before(c.expiresAt) && len(c.counts) > 0 {
		cached := make(map[string]int64, len(c.counts))
		for k, v := range c.counts {
			cached[k] = v
		}
		return cached, true
	}

	return nil, false
}

// Set stores counts in the cache with TTL expiration
// Creates a defensive copy to prevent external modifications
func (c *CountsCache) Set(counts map[string]int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create defensive copy
	c.counts = make(map[string]int64, len(counts))
	for k, v := range counts {
		c.counts[k] = v
	}
	c.expiresAt = time.Now().Add(c.TTL)
}

// Invalidate clears the cache by expiring it immediately
func (c *CountsCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.expiresAt = time.Now().Add(-1 * time.Second)
}
