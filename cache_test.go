package cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/taybart/cache"
)

func TestGetSet(t *testing.T) {
	is := is.New(t)

	c := cache.New(cache.Default())
	defer c.Finish()

	type testStruct struct {
		Message string
	}

	c.Set("test", "test")
	c.Set("test-gob", testStruct{"test"})

	var item string
	err := c.Get("test", &item)
	is.NoErr(err)
	is.Equal(item, "test")
	var gobItem testStruct
	err = c.Get("test-gob", &gobItem)
	is.NoErr(err)
	is.Equal(gobItem.Message, "test")
}

func TestPubSub(t *testing.T) {
	is := is.New(t)

	c := cache.New(cache.Default())
	defer c.Finish()

	type testStruct struct {
		Message string
	}

	c.Set("test-gob", testStruct{"test"})

	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	ch := c.Subscribe("test-gob")
	go func(ctx context.Context, ch chan any) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			is.NoErr(ctx.Err())
		case item := <-ch:
			is.True(item.(testStruct).Message == "woohoo")
			return
		}
	}(ctx, ch)

	ch = c.Subscribe("test-gob")
	go func(ctx context.Context, ch chan any) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			is.NoErr(ctx.Err())
		case item := <-ch:
			is.True(item.(testStruct).Message == "woohoo")
			return
		}
	}(ctx, ch)
	c.Set("test-gob", testStruct{"woohoo"})

	wg.Wait()
}

func TestPrune(t *testing.T) {
	is := is.New(t)

	c := cache.New(cache.Config{
		TTL:           time.Millisecond,
		PruneRate:     time.Millisecond * 5,
		SleepDuration: time.Millisecond,
	})
	defer c.Finish()

	c.Set("test", "test")
	time.Sleep(7 * time.Millisecond)

	var item string
	err := c.Get("test", &item)
	is.True(errors.Is(err, cache.ErrNotFound))
}
