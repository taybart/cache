package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrNotFound item is not found
	ErrNotFound = errors.New("item not found")
)

const (
	// TTLNeverExpire keep forever
	TTLNeverExpire = -1
)

var sharedCache = Cache{
	items: make(map[string]Item),
}

// Cache is a simple key value store
type Cache struct {
	init   bool
	config Config
	Ctx    context.Context
	cancel context.CancelFunc
	items  map[string]Item
	mu     sync.RWMutex
	subs   map[string][]chan any
}

// Item is a cache item
type Item struct {
	CreatedAt time.Time
	TTL       time.Duration
	Data      []byte
}

// New creates a new cache,
// default config can be set with cache.New(cache.Default())
func New(config Config) *Cache {
	config.init()

	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache{
		init:   true,
		config: config,
		Ctx:    ctx,
		cancel: cancel,
		items:  make(map[string]Item),
		subs:   make(map[string][]chan any),
	}
	go c.Prune()
	return c
}

func NewShared(config Config) *Cache {
	if !sharedCache.init {
		config.init()
		sharedCache.config = config
		sharedCache.init = true
		sharedCache.Ctx = context.Background()
		go sharedCache.Prune()
	}
	return &sharedCache
}

func (c *Cache) Finish() {
	c.cancel()
}

func (c *Cache) Set(key string, data any) error {
	c.isInit()
	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}
	(*c).items[key] = Item{
		CreatedAt: time.Now(),
		Data:      buf.Bytes(),
		TTL:       c.config.TTL,
	}
	c.updateSubs(key, data)
	return nil
}

func (c *Cache) SetWithTTL(key string, data any, ttl time.Duration) error {
	c.isInit()
	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}
	(*c).items[key] = Item{
		CreatedAt: time.Now(),
		Data:      buf.Bytes(),
		TTL:       ttl,
	}
	return nil
}

// Get will get the data if it exists, returns cache.ErrNotFound on error
func (c *Cache) Get(key string, data any) error {
	c.isInit()
	c.mu.RLock()
	defer c.mu.RUnlock()
	if i, ok := c.items[key]; ok {
		return gob.NewDecoder(bytes.NewReader(i.Data)).Decode(data)
	}
	return ErrNotFound
}

// GetFallback will get the data if it exists
// otherwise it will set the data and return it
func (c *Cache) GetFallback(key string, get *any, set func() (any, error)) error {
	c.isInit()
	c.mu.RLock()
	defer c.mu.RUnlock()

	if i, ok := c.items[key]; ok {
		return gob.NewDecoder(bytes.NewReader(i.Data)).Decode(get)
	}

	var buf bytes.Buffer
	s, err := set()
	if err != nil {
		return err
	}

	*get = s

	if err := gob.NewEncoder(&buf).Encode(s); err != nil {
		return err
	}
	update := Item{
		CreatedAt: time.Now(),
		Data:      buf.Bytes(),
		TTL:       c.config.TTL,
	}
	(*c).items[key] = update

	c.updateSubs(key, update)
	if i, ok := c.items[key]; ok {
		return gob.NewDecoder(bytes.NewReader(i.Data)).Decode(get)
	}
	return ErrNotFound
}

// GetFallbackWithTTL will get the data if it exists
// otherwise it will set the data and return it
func (c *Cache) GetFallbackWithTTL(key string, get *any, ttl time.Duration, set func() (any, error)) error {
	c.isInit()
	c.mu.RLock()
	defer c.mu.RUnlock()

	if i, ok := c.items[key]; ok {
		return gob.NewDecoder(bytes.NewReader(i.Data)).Decode(get)
	}
	var buf bytes.Buffer
	s, err := set()
	if err != nil {
		return err
	}

	*get = s

	if err = gob.NewEncoder(&buf).Encode(s); err != nil {
		return err
	}
	item := Item{
		CreatedAt: time.Now(),
		Data:      buf.Bytes(),
		TTL:       ttl,
	}
	(*c).items[key] = item
	c.updateSubs(key, item)
	return nil
}

func (c *Cache) Prune() {
	c.isInit()
	if c.config.PruneRate == 0 {
		return
	}
	lastPrune := time.Now()
prune:
	for {
		select {
		case <-c.Ctx.Done():
			break prune
		default:
			if time.Since(lastPrune) > c.config.PruneRate {
				lastPrune = time.Now()
				if len(c.items) > 0 {
					c.mu.RLock()
					now := time.Now()
					for d, i := range c.items {
						if i.TTL == TTLNeverExpire {
							continue
						}
						if now.Sub(i.CreatedAt) > i.TTL {
							c.mu.RUnlock()
							c.mu.Lock()
							delete(c.items, d)
							c.mu.Unlock()
							c.mu.RLock()
						}
					}
					c.mu.RUnlock()
				}
			}
			time.Sleep(c.config.SleepDuration) // chill for a sec
		}
	}
}

func (c *Cache) Subscribe(key string) chan any {
	c.isInit()
	ch := make(chan any)
	if len(c.subs[key]) == 0 {
		c.subs[key] = make([]chan any, 0)
	}

	c.subs[key] = append(c.subs[key], ch)

	return ch
}

func (c *Cache) updateSubs(key string, update any) {
	// FIXME: work on sub locks
	for _, ch := range c.subs[key] {
		ch <- update
	}
}

func (c *Cache) isInit() {
	if !c.init {
		panic(fmt.Errorf("cache not started, use something like cache.New(cache.Default())"))
	}
}
