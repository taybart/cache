package cache

import (
	"time"
)

var defaultConfig = Config{
	TTL:           time.Hour * 24,
	PruneRate:     time.Minute * 5,
	SleepDuration: time.Millisecond * 100,
}

// Config for the cache
// TTL: the time to live for an item
// PruneRate: how often to prune the cache
// SleepDuration: how long to sleep between prunes
type Config struct {
	TTL           time.Duration
	PruneRate     time.Duration
	SleepDuration time.Duration
}

// Default returns the default config
// TTL: 24 hours
// PruneRate: 5 minutes
// SleepDuration: 100 milliseconds
func Default() Config {
	return defaultConfig
}

func (c *Config) init() {
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
