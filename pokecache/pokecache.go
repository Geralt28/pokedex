package pokecache

import (
	"sync"
	"time"
)

const interval = 10

//var cache map[string]cacheEntry // Declare globally
//var cacheMu sync.RWMutex

type Cache struct {
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

//type cache_interface interface {
//	Add()
//	Get()
//	reapLoop()
//}

// Create a cache.Add() method that adds a new entry to the cache. It should take a key (a string) and a val (a []byte).
func (c *Cache) Add(key string, val []byte) {

	c.mu.Lock() // Lock before modifying the map
	defer c.mu.Unlock()

	c.cache[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) (val []byte, ok bool) {

	c.mu.RLock() // Lock before modifying the map
	defer c.mu.RUnlock()

	entry, ok := c.cache[key]
	if ok {
		val = entry.val
	}

	return val, ok
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		czas := time.Now()

		c.mu.Lock() // Lock before modifying the map
		for key, i := range c.cache {
			if czas.Sub(i.createdAt).Seconds() > interval {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

func NewCache() *Cache {
	c := &Cache{
		cache: make(map[string]cacheEntry),
	}
	go c.reapLoop()
	return c
}
