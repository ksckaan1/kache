# kache

In-memory cache with TTL and generics support.

## Features
- TTL support
- Generics support
- Supported Cache Replacement Strategies
  - LRU (Least Recently Used)
  - MRU (Most Recently Used)
  - LFU (Least Frequently Used)
  - MFU (Most Frequently Used)
  - FIFO (First In First Out)
  - LIFO (Last In First Out)
  - NONE (no replacement)

## Installation

```bash
go get github.com/ksckaan1/kache
```

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/ksckaan1/kache"
)

func main() {
	k := kache.New[string, string](kache.Config{
		ReplacementStrategy: kache.ReplacementStrategyLRU,
		MaxRecordTreshold:   1000,
		CleanNum:            100,
	})
	defer k.Close()

	// Set with TTL
	k.SetWithTTL("token/user_id:1", "eyJhbGciOiJ...", 30*time.Minute)

	// Set without TTL
	k.Set("get_user_response/user_id:1", "John Doe")
	k.Set("get_user_response/user_id:2", "Jane Doe")
	k.Set("get_user_response/user_id:3", "Walter White")
	k.Set("get_user_response/user_id:4", "Jesse Pinkman")

	k.Delete("get_user_response/user_id:1")

	fmt.Println(k.Get("token/user_id:1"))             // eyJhbGciOiJ..., true
	fmt.Println(k.Get("get_user_response/user_id:1")) // "", false

	fmt.Println("keys", k.Keys())   // List of keys
	fmt.Println("count", k.Count()) // Number of keys

	k.Flush() // Deletes all keys
}

```

## Benchmark Tests

### Set With TTL / Set Without TTL
```bash
goos: darwin
goarch: arm64
pkg: github.com/ksckaan1/kache
BenchmarkKacheSetWithTTL
BenchmarkKacheSetWithTTL-8   	 4838650	       236.8 ns/op	     129 B/op	       4 allocs/op
PASS
ok  	github.com/ksckaan1/kache	1.842s
```

### Get
```bash
goos: darwin
goarch: arm64
pkg: github.com/ksckaan1/kache
BenchmarkKacheGet
BenchmarkKacheGet-8   	83910825	        13.98 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/ksckaan1/kache	2.094s
```

