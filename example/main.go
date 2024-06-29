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

	k := kache.New[string, string]().
		WithCleanStrategyLRU(10000, 1000)

	go k.Poll(ctx, time.Second) // make sense if used value with ttl

	k.Set("get_user_response/user_id:1", "user1")
	k.SetWithTTL("user_token:1", "token1", 10*time.Minute)

	fmt.Println("count:", k.Count()) // 2
	fmt.Println("keys:", k.Keys())   // [get_user_response/user_id:1 user_token:1]

	k.Delete("get_user_response/user_id:1")

	k.Flush()
	fmt.Println("count:", k.Count()) // 0
}
