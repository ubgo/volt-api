package cacheredis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const testKeyPrefix = "cacheredis_test_"

// testRedisClient creates a Redis client for tests. Skips t if Redis is not reachable.
func testRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	opts := &redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use DB 1 to avoid clashing with other data
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s: %v (start with e.g. task dev)", opts.Addr, err)
	}
	return client
}

func TestNew(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	if cache == nil {
		t.Fatal("New returned nil")
	}
	if cache.client != client {
		t.Error("cache.client != client")
	}
}

func TestVersion(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	v := cache.Version()
	if v == "" {
		t.Error("Version() returned empty string")
	}
	t.Logf("Version: %s", v)
}

func TestPutAndGet(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	ctx := context.Background()
	defer client.FlushDB(ctx)

	key := testKeyPrefix + "put_get"
	val := map[string]interface{}{"foo": "bar", "n": float64(42)}
	ttl := 60

	ok, err := cache.Put(key, val, ttl)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if !ok {
		t.Fatal("Put returned false")
	}

	got := cache.Get(key)
	if got == nil {
		t.Fatal("Get returned nil")
	}
	m, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("Get: expected map, got %T", got)
	}
	if m["foo"] != "bar" || m["n"] != float64(42) {
		t.Errorf("Get: got %v", m)
	}
}

func TestGetMissing(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	ctx := context.Background()
	defer client.FlushDB(ctx)

	got := cache.Get(testKeyPrefix + "missing")
	if got != nil {
		t.Errorf("Get(missing): expected nil, got %v", got)
	}
}

func TestDel(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	ctx := context.Background()
	defer client.FlushDB(ctx)

	key := testKeyPrefix + "del"
	_, _ = cache.Put(key, "value", 60)
	cache.Del(key)
	got := cache.Get(key)
	if got != nil {
		t.Errorf("Get after Del: expected nil, got %v", got)
	}
}

func TestFlush(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	ctx := context.Background()
	defer client.FlushDB(ctx)

	key := testKeyPrefix + "flush"
	_, _ = cache.Put(key, "value", 60)
	cache.Flush()
	got := cache.Get(key)
	if got != nil {
		t.Errorf("Get after Flush: expected nil, got %v", got)
	}
}

func TestPutInvalidJSON(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	// Channel cannot be marshalled to JSON
	ch := make(chan int)
	ok, err := cache.Put(testKeyPrefix+"invalid", ch, 60)
	if err == nil {
		t.Error("Put(channel): expected error")
	}
	if ok {
		t.Error("Put(channel): expected false")
	}
}

func TestPutGetStringValue(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	cache := New(client)
	ctx := context.Background()
	defer client.FlushDB(ctx)

	key := testKeyPrefix + "string"
	_, _ = cache.Put(key, "plain", 60)
	got := cache.Get(key)
	if got == nil {
		t.Fatal("Get: expected non-nil")
	}
	if s, ok := got.(string); !ok || s != "plain" {
		t.Errorf("Get: expected \"plain\", got %v", got)
	}
}
