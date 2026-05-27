package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Store is a thin wrapper around go-cache for in-memory storage with TTL support.
type Store struct {
	c *gocache.Cache
}

// New creates a Store.
// defaultTTL is used for items stored without an explicit TTL.
// cleanupInterval controls how often expired items are purged from memory.
func New(defaultTTL, cleanupInterval time.Duration) *Store {
	return &Store{
		c: gocache.New(defaultTTL, cleanupInterval),
	}
}

// Set stores a value under the given key with a specific TTL.
// Pass gocache.NoExpiration (0) for items that should never expire.
func (s *Store) Set(key string, value any, ttl time.Duration) {
	s.c.Set(key, value, ttl)
}

// Get retrieves a value from the cache. Returns (value, true) on hit.
func (s *Store) Get(key string) (any, bool) {
	return s.c.Get(key)
}

// Delete removes an item from the cache.
func (s *Store) Delete(key string) {
	s.c.Delete(key)
}
