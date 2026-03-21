package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt time.Time
}

// Cache is a thread-safe in-memory cache with TTL and background eviction.
type Cache struct {
	mu    sync.RWMutex
	items map[string]entry
	ttl   time.Duration
	stop  chan struct{}
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]entry),
		ttl:   ttl,
		stop:  make(chan struct{}),
	}
	go c.evictLoop()
	return c
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(c.ttl)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Stop terminates the background eviction goroutine.
func (c *Cache) Stop() {
	close(c.stop)
}

func (c *Cache) evictLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.evict()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache) evict() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}
