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
	ErrNotFound = errors.New("Item not found")
)

var (
	defaultDuration  = time.Hour * 24
	defaultPruneRate = time.Minute * 5
	pruningShared    = false
	sharedInit       = false
	sharedCache      = Cache{
		PruneRate: defaultPruneRate,
		TTL:       defaultDuration,
		Items:     make(map[string]Item),
	}
)

type Item struct {
	CreatedAt time.Time
	Data      []byte
}

type Cache struct {
	Ctx       context.Context
	cancel    context.CancelFunc
	TTL       time.Duration
	Items     map[string]Item
	PruneRate time.Duration
	Mu        sync.Mutex
}

func New() *Cache {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache{
		Ctx:       ctx,
		cancel:    cancel,
		PruneRate: defaultPruneRate,
		Items:     make(map[string]Item),
		TTL:       defaultDuration,
	}
	go c.Prune()
	return c
}

func NewShared() *Cache {
	if !sharedInit {
		sharedCache.Ctx = context.Background()
		go sharedCache.Prune()
	}
	return &sharedCache
}

func (c *Cache) Finish() {
	c.cancel()
}

func (c *Cache) SetTTL(ttl time.Duration) {
	c.TTL = ttl
}

func (c *Cache) SetPruneRate(pr time.Duration) {
	c.PruneRate = pr
}

func (c *Cache) Set(key string, data interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}
	(*c).Items[key] = Item{
		CreatedAt: time.Now(),
		Data:      buf.Bytes(),
	}
	return nil
}

func (c *Cache) Get(key string, data interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if i, ok := c.Items[key]; ok {
		return gob.NewDecoder(bytes.NewReader(i.Data)).Decode(data)
	}
	return ErrNotFound
}

func (c *Cache) Prune() {
	for {
		select {
		case <-c.Ctx.Done():
			break
		default:
			if len(c.Items) > 0 {
				now := time.Now()
				for d, i := range c.Items {
					if now.Sub(i.CreatedAt) > c.TTL {
						c.Mu.Lock()
						delete(c.Items, d)
						c.Mu.Unlock()
					}
				}
			}
		}
	}
}
