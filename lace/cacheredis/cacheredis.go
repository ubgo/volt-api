package cacheredis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const version = "0.02"

// Cache is a Redis-backed cache. The Redis client is injected as a dependency.
type Cache struct {
	client *redis.Client
}

// Version returns the cache implementation version.
func (c *Cache) Version() string {
	return version
}

// New builds a Cache that uses the given Redis client.
func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Put stores val at key as JSON with ttl seconds. Returns false on marshal or Redis error.
func (c *Cache) Put(key string, val interface{}, ttl int) (bool, error) {
	ctx := context.Background()
	p, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	expiration := time.Duration(ttl) * time.Second
	err = c.client.Set(ctx, key, p, expiration).Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

// Get returns the value at key, unmarshalled into an interface{}. Returns nil if missing or on error.
func (c *Cache) Get(key string) interface{} {
	ctx := context.Background()
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return data
	}
	return v
}

// Del removes the key from the cache.
func (c *Cache) Del(key string) {
	ctx := context.Background()
	_ = c.client.Del(ctx, key).Err()
}

// Flush removes all keys in the current Redis DB.
func (c *Cache) Flush() {
	ctx := context.Background()
	_ = c.client.FlushDB(ctx).Err()
}
