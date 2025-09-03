package cache

import (
	"sync"
	"time"
)

type Entry struct {
	Key       string
	Value     []byte
	ExpiresAt int64
	CreatedAt int64
	LastUsed  int64
	UseCount  int
}

type Cache struct {
	entries map[string]*Entry
	mu      sync.RWMutex
	maxSize int64
	size    int64
}

func New(maxSize int64) *Cache {
	return &Cache{
		entries: make(map[string]*Entry),
		maxSize: maxSize,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	
	if entry.ExpiresAt > 0 && time.Now().UnixNano() > entry.ExpiresAt {
		delete(c.entries, key)
		c.size -= int64(len(entry.Value))
		return nil, false
	}
	
	entry.LastUsed = time.Now().UnixNano()
	entry.UseCount++
	
	return entry.Value, true
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now().UnixNano()
	var expiresAt int64
	if ttl > 0 {
		expiresAt = now + int64(ttl.Nanoseconds())
	}
	
	if existing, exists := c.entries[key]; exists {
		c.size -= int64(len(existing.Value))
	}
	
	entry := &Entry{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		LastUsed:  now,
		UseCount:  1,
	}
	
	c.entries[key] = entry
	c.size += int64(len(value))
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return false
	}
	
	delete(c.entries, key)
	c.size -= int64(len(entry.Value))
	return true
}

func (c *Cache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*Entry)
	c.size = 0
}

func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}
	return keys
}

func (c *Cache) GetEntries() map[string]*Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entries := make(map[string]*Entry, len(c.entries))
	for k, v := range c.entries {
		entries[k] = v
	}
	return entries
}