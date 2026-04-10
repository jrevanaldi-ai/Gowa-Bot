package helper

import (
	"sync"
	"time"
)


type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}


type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}


func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}


	go cache.cleanupLoop()

	return cache
}


func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}


func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().After(item.Expiration) {
		return nil, false
	}

	return item.Value, true
}


func (c *Cache) GetWithDefault(key string, defaultValue interface{}) interface{} {
	value, found := c.Get(key)
	if !found {
		return defaultValue
	}
	return value
}


func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}


func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]CacheItem)
}


func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, item := range c.items {
		if time.Now().Before(item.Expiration) {
			count++
		}
	}
	return count
}


func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}


func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}


type RateLimiter struct {
	cache     *Cache
	limit     int
	windowSec time.Duration
}


func NewRateLimiter(limit int, windowSec time.Duration) *RateLimiter {
	return &RateLimiter{
		cache:     NewCache(),
		limit:     limit,
		windowSec: windowSec,
	}
}


func (rl *RateLimiter) Allow(key string) bool {
	rl.cache.mu.Lock()
	defer rl.cache.mu.Unlock()

	count, _ := rl.cache.items[key].Value.(int)
	if count >= rl.limit {
		return false
	}

	rl.cache.items[key] = CacheItem{
		Value:      count + 1,
		Expiration: time.Now().Add(rl.windowSec),
	}

	return true
}
