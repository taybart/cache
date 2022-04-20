# cache

![Test](https://github.com/taybart/cache/workflows/Test/badge.svg)

## Usage

#### Single

```go
package main

import (
  "fmt"
  "github.com/taybart/cache"
)

func main() {
  c := cache.New(cache.Default())
  defer c.Finish()

  // set item
  c.Set("test", "test")
  c.SetWithTTL("user", "me", cache.TTLNeverExpire)

  var item string
  err := c.Get("test", &item)
  fmt.Println(item) // output: test
}
```

#### Shared

```go
package main

import (
  "fmt"
  "github.com/taybart/cache"
)

func main() {
  // Will access the same cache across calls
  c := cache.NewShared(cache.Default())
  defer c.Finish()

  // set item
  c.Set("test", "test")

  var item string
  err := c.Get("test", &item)
  fmt.Println(item) // output: test
}
```

## Config

#### Pruning 

Defaults:
TTL: 24 hours
PruneRate: 5 Minutes
SleepDuration: 100 Milliseconds

```go
package main

import (
  "fmt"
  "github.com/taybart/cache"
)

func main() {
  c := cache.New(cache.Config{
    TTL: time.Nanosecond, // we really don't care about data here
    PruneRate: 3*time.Nanosecond, // extend the prune rate to some stuff might live
    SleepDuration: time.Hour, // wait an hour between checks
  })
  defer c.Finish()

  c.SetPruneRate(0) // actually, don't prune at all
}
```
