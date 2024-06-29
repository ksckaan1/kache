# kache

In-memory cache with TTL and generics support.

## Features
- TTL support
- generics support
- Cache Replacement Strategy: LRU, FIFO, None

## Installation

```bash
go get github.com/ksckaan1/kache
```

## Example

```go
package main

import (
  "context"
  "fmt"
  "time"

  "github.com/ksckaan1/kache"
)

func main() {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  // key = string, value = string
  k := kache.New[string, string]().
    WithCleanStrategyLRU(10000, 1000)
  // if reaches 10000, it will delete 1000 records
  // by LRU (Least Recently Used)

  go k.Poll(ctx, time.Second) // make sense if storing value with ttl

  k.Set("get_user_response/user_id:1", "user1")
  k.SetWithTTL("user_token:1", "token1", 10*time.Minute)

  fmt.Println("count:", k.Count()) // 2
  fmt.Println("keys:", k.Keys())   // [get_user_response/user_id:1 user_token:1]

  k.Delete("get_user_response/user_id:1") // deletes record with the given key

  k.Flush() // deletes all records
  fmt.Println("count:", k.Count()) // 0
}
```