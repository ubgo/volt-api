package cache

import "strings"

const Version = "0.0.1"

type Store interface {
	Get(key string) interface{}
	Put(key string, val interface{}, ttl int) (bool, error)
	Del(key string)
	Flush()
}

// Cache struct
type Cache struct {
	Config
}

type Config struct {
	Store     Store
	GlobalKey string `wire:"-"`
}

func New(config Config) Cache {
	if len(strings.TrimSpace(config.GlobalKey)) == 0 {
		config.GlobalKey = "jcache"
	}

	cache := Cache{Config: config}
	return cache
}

func (a *Cache) getKey(key string) string {
	key = a.Config.GlobalKey + ":" + key
	return key
}

func (a *Cache) Get(key string) interface{} {
	key = a.getKey(key)
	return a.Config.Store.Get(key)
}

func (a *Cache) Put(key string, val interface{}, ttl int) (bool, error) {
	key = a.getKey(key)
	return a.Config.Store.Put(key, val, ttl)
}

func (a *Cache) Del(key string) {
	key = a.getKey(key)
	a.Config.Store.Del(key)
}

func (a *Cache) Flush() {
	a.Config.Store.Flush()
}
