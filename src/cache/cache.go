package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotExists = errors.New("cache value does not exists")
	ErrExpired   = errors.New("cache value has been expired")
)

var itemsPool = sync.Pool{
	New: func() interface{} { return new(Item) },
}

type Item struct {
	value interface{}
	ttl   time.Time
}

type Cache struct {
	items map[string]*Item
	mx    sync.RWMutex
}

var cache *Cache

var once sync.Once

func GetCache() *Cache {
	once.Do(func() {
		cache = new(Cache)
		cache.items = make(map[string]*Item, 0)
	})
	return cache
}

func (c *Cache) Set(name string, value interface{}, ttl time.Duration) *Cache {
	defer c.mx.Unlock()
	c.mx.Lock()
	if _, ok := c.items[name]; !ok {
		c.items[name] = itemsPool.Get().(*Item)
	}
	c.items[name].value = value
	c.items[name].ttl = time.Now().Add(ttl)

	return c
}

func (c *Cache) Delete(name string) error {
	defer c.mx.Unlock()
	c.mx.Lock()
	if _, ok := c.items[name]; !ok {
		return ErrNotExists
	}
	c.items[name].value = nil
	itemsPool.Put(c.items[name])
	c.items[name] = nil

	return nil
}

func (c *Cache) Exists(name string) bool {
	defer c.mx.RUnlock()
	c.mx.RLock()
	_, ok := c.items[name]
	return ok
}

func (c *Cache) RLock() {
	c.mx.RLock()
}

func (c *Cache) RUnlock() {
	c.mx.RUnlock()
}

func (c *Cache) Get(name string) (interface{}, error) {
	var item *Item
	var ok bool
	defer c.mx.RUnlock()
	c.mx.RLock()
	if item, ok = c.items[name]; !ok {
		return nil, ErrNotExists
	}
	if item.ttl.Before(time.Now()) {
		return nil, ErrExpired
	}

	return c.items[name].value, nil
}

func (c *Cache) GetRaw(name string) (interface{}, error) {
	var item *Item
	var ok bool

	if item, ok = c.items[name]; !ok {
		return nil, ErrNotExists
	}
	if item.ttl.Before(time.Now()) {
		return nil, ErrExpired
	}

	return c.items[name].value, nil
}

func (c *Cache) StartBackground(ctx context.Context) {
	go c._garbageCollector(ctx)
}

func (c *Cache) _garbageCollector(ctx context.Context) {
	for {
		select {
		case <-time.After(5 * time.Second):
			c.mx.Lock()
			for name, item := range c.items {
				if item.ttl.Before(time.Now()) {
					c.items[name] = nil
					delete(c.items, name)
				}
			}
			c.mx.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
