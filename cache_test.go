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

	c := cache.New()
	c.Set("test", "test")

	var item string
	err := c.Get("test", &item)
	is.NoErr(err)
	is.Equal(item, "test")
	c.Finish()
}

func TestPrune(t *testing.T) {
	is := is.New(t)

	c := cache.New()
	c.SetTTL(time.Millisecond)
	c.SetPruneRate(5 * time.Millisecond)
	c.Set("test", "test")
	time.Sleep(15 * time.Millisecond)

	var item string
	err := c.Get("test", &item)
	is.True(errors.Is(err, cache.ErrNotFound))
	c.Finish()
}
