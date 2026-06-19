package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached response with its expiration time
type CacheEntry struct {
	Data       []byte
	ExpiresAt  time.Time
	Headers    map[string][]string
	StatusCode int
}

// Cache is a thread-safe in-memory cache
type Cache struct {
	mu    sync.RWMutex
	store map[string]CacheEntry
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	return &Cache{
		store: make(map[string]CacheEntry),
	}
}

// Get retrieves a value from the cache if it exists and hasn't expired
func (c *Cache) Get(key string) ([]byte, map[string][]string, int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.store[key]
	if !exists {
		return nil, nil, 0, false
	}

	// Check if the entry has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, nil, 0, false
	}

	return entry.Data, entry.Headers, entry.StatusCode, true
}

// Set stores a value in the cache with an expiration time
func (c *Cache) Set(key string, data []byte, headers map[string][]string, statusCode int, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = CacheEntry{
		Data:       data,
		Headers:    headers,
		StatusCode: statusCode,
		ExpiresAt:  time.Now().Add(ttl),
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]CacheEntry)
}

// CleanupExpired removes expired entries from the cache
func (c *Cache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.store {
		if now.After(entry.ExpiresAt) {
			delete(c.store, key)
		}
	}
}
