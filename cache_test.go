package cache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/taybart/cache"
)

func TestGetSet(t *testing.T) {
	is := is.New(t)

	c := cache.New(cache.Default())
	c.Set("test", "test")

	var item string
	err := c.Get("test", &item)
	is.NoErr(err)
	is.Equal(item, "test")
	c.Finish()
}

func TestPrune(t *testing.T) {
	is := is.New(t)

	c := cache.New(cache.Config{
		TTL:           time.Millisecond,
		PruneRate:     time.Millisecond * 5,
		SleepDuration: time.Millisecond,
	})
	c.Set("test", "test")
	time.Sleep(7 * time.Millisecond)

	var item string
	err := c.Get("test", &item)
	is.True(errors.Is(err, cache.ErrNotFound))
	c.Finish()
}
