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
  c := cache.New()
  defer c.Finish()

  // set item
  c.Set("test", "test")

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
  c := cache.NewShared() // Will access the same cache across calls
  defer c.Finish()

  // set item
  c.Set("test", "test")

  var item string
  err := c.Get("test", &item)
  fmt.Println(item) // output: test
}
```

## Config

Defaults:
TTL: 24 hours
PruneRate: 5 Minutes

```go
package main

import (
  "fmt"
  "github.com/taybart/cache"
)

func main() {
  c := cache.New() // Will access the same cache across calls
  defer c.Finish()

  c.SetTTL(time.Nanosecond) // we really don't care about data here
  c.SetPruneRate(3*time.Nanosecond) // extend the prune rate to some stuff might live
}
```
