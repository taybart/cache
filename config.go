package cache

import (
	"time"
)

var defaultConfig = Config{
	TTL:           time.Hour * 24,
	PruneRate:     time.Minute * 5,
	SleepDuration: time.Millisecond * 100,
}

type Config struct {
	TTL           time.Duration
	PruneRate     time.Duration
	SleepDuration time.Duration
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
