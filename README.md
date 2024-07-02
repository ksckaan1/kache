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