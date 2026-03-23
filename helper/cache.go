package helper

import (
	"sync"
	"time"
)

// CacheItem adalah item dalam cache dengan expiry time
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// Cache adalah sistem cache full dengan TTL support
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewCache membuat cache baru
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Set menambahkan item ke cache dengan TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Get mengambil item dari cache
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

// GetWithDefault mengambil item dari cache dengan default value
func (c *Cache) GetWithDefault(key string, defaultValue interface{}) interface{} {
	value, found := c.Get(key)
	if !found {
		return defaultValue
	}
	return value
}

// Delete menghapus item dari cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear membersihkan semua item dari cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]CacheItem)
}

// Count mendapatkan jumlah item dalam cache
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

// cleanupLoop membersihkan expired items secara berkala
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup membersihkan expired items
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

// RateLimiter adalah rate limiter berbasis cache
type RateLimiter struct {
	cache     *Cache
	limit     int
	windowSec time.Duration
}

// NewRateLimiter membuat rate limiter baru
func NewRateLimiter(limit int, windowSec time.Duration) *RateLimiter {
	return &RateLimiter{
		cache:     NewCache(),
		limit:     limit,
		windowSec: windowSec,
	}
}

// Allow memeriksa apakah request diizinkan
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
