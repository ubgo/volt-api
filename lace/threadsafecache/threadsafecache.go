package threadsafecache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type CacheEntry[T any] struct {
	Value     T
	CreatedAt time.Time
}

type Cache[T any] struct {
	storage *lru.Cache[string, CacheEntry[T]]
	mu      sync.RWMutex
	ttl     time.Duration
}

func NewCache[T any](size int, ttl time.Duration) (*Cache[T], error) {
	storage, err := lru.New[string, CacheEntry[T]](size)
	if err != nil {
		return nil, err
	}

	return &Cache[T]{
		storage: storage,
		ttl:     ttl,
	}, nil
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage.Add(key, CacheEntry[T]{
		Value:     value,
		CreatedAt: time.Now(),
	})
}

func (c *Cache[T]) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage.Remove(key)
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.storage.Get(key)
	if !exists {
		var zero T
		return zero, false
	}

	if time.Since(entry.CreatedAt) > c.ttl {
		c.storage.Remove(key)
		var zero T
		return zero, false
	}

	return entry.Value, true
}
