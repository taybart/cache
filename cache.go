package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("item not found")
)

const (
	TTLNeverExpire = -1
)

var (
	defaultConfig = Config{
		TTL:           time.Hour * 24,
		PruneRate:     time.Minute * 5,
		SleepDuration: time.Millisecond * 100,
	}

	sharedCache = Cache{
		items: make(map[string]Item),
	}
)

type Config struct {
	TTL           time.Duration
	PruneRate     time.Duration
	SleepDuration time.Duration
}

type Cache struct {
	init   bool
	config Config
	Ctx    context.Context
	cancel context.CancelFunc
	items  map[string]Item
	mu     sync.RWMutex
}

type Item struct {
	CreatedAt time.Time
	TTL       time.Duration
	Data      []byte
}

func Default() Config {
	return defaultConfig
}

func (c *Config) Init() {
	// make sure defaults are preserved
	if c.TTL == 0 {
		c.TTL = defaultConfig.TTL
	}
	if c.PruneRate == 0 {
		c.PruneRate = defaultConfig.PruneRate
	}
	if c.SleepDuration == 0 {
		c.SleepDuration = defaultConfig.SleepDuration
	}
}

func New(config Config) *Cache {
	config.Init()

	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache{
		init:   true,
		config: config,
		Ctx:    ctx,
		cancel: cancel,
		items:  make(map[string]Item),
	}
	go c.Prune()
	return c
}

func NewShared(config Config) *Cache {
	if !sharedCache.init {
		config.Init()
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

func (c *Cache) SetConfig(config Config) {
	c.config = config
}

// DEPRECATED: should just use config?
func (c *Cache) SetTTL(ttl time.Duration) {
	c.config.TTL = ttl
}

// DEPRECATED: should just use config?
func (c *Cache) SetPruneRate(pr time.Duration) {
	c.config.PruneRate = pr
	c.cancel()
	c.Ctx, c.cancel = context.WithCancel(context.Background())
	go c.Prune()
}

// DEPRECATED: should just use config?
func (c *Cache) SetSleepDuration(sd time.Duration) {
	c.config.SleepDuration = sd
	c.cancel()
	c.Ctx, c.cancel = context.WithCancel(context.Background())
	go c.Prune()
}

func (c *Cache) Set(key string, data interface{}) error {
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
	return nil
}

func (c *Cache) SetWithTTL(key string, data interface{}, ttl time.Duration) error {
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

func (c *Cache) Get(key string, data interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if i, ok := c.items[key]; ok {
		return gob.NewDecoder(bytes.NewReader(i.Data)).Decode(data)
	}
	return ErrNotFound
}

func (c *Cache) Prune() {
	if c.config.PruneRate == 0 {
		return
	}
	lastPrune := time.Now()
pruner:
	for {
		select {
		case <-c.Ctx.Done():
			break pruner
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
