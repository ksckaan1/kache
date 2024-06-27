package main

import (
	"fmt"

	"github.com/ksckaan1/kache"
)

func main() {
	k := kache.New[string, string]()

	k.Set("foo1", "bar")
	k.Set("foo2", "bar")

	fmt.Println("key count:", k.Count())
	fmt.Println("keys:", k.Keys())

	fmt.Println(k.Get("foo1"))

	k.Delete("foo1")

	fmt.Println(k.Get("foo"))

	k.Flush()

	fmt.Println(k.Get("foo2"))
}
