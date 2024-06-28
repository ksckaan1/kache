package main

import (
	"fmt"

	"github.com/ksckaan1/kache"
)

func main() {
	k := kache.New[string, string]().
		WithCleanStrategyFIFO(10, 5)

	k.Set("key1", "value")
	k.Set("key2", "value")
	k.Set("key3", "value")
	k.Set("key4", "value")
	k.Set("key5", "value")
	k.Set("key6", "value")
	k.Set("key7", "value")
	k.Set("key8", "value")
	k.Set("key9", "value")
	k.Set("key10", "value")

	fmt.Println(k.Get("key1"))
	fmt.Println(k.Get("key2"))

	k.Set("key11", "value")

	fmt.Println(k.Keys())
}
