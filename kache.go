package kache

import (
	"context"
	"sync"
	"time"
)

type valueWithTTL[V any] struct {
	value V
	ttl   time.Time
}

type Kache[K comparable, V any] struct {
	store        map[K]valueWithTTL[V]
	mut          *sync.Mutex
	maxRecordNum int
}

func New[K comparable, V any]() *Kache[K, V] {
	k := &Kache[K, V]{
		store:        make(map[K]valueWithTTL[V]),
		mut:          new(sync.Mutex),
		maxRecordNum: -1,
	}
	return k
}

// WithMaxRecordNum sets the maximum number of records to keep in the cache.
func (k *Kache[K, V]) WithMaxRecordNum(num int) *Kache[K, V] {
	k.maxRecordNum = num
	return k
}

// Set sets a value in the cache.
func (k *Kache[K, V]) Set(key K, v V) {
	defer k.lock()()
	k.store[key] = valueWithTTL[V]{
		value: v,
	}
}

// SetWithTTL sets a value in the cache with a TTL. If the TTL is 0, the value will not expire.
func (k *Kache[K, V]) SetWithTTL(key K, v V, ttl time.Duration) {
	defer k.lock()()
	var ttlVal time.Time
	if ttl > 0 {
		ttlVal = time.Now().Add(ttl)
	}
	k.store[key] = valueWithTTL[V]{
		value: v,
		ttl:   ttlVal,
	}
}

// Get gets a value from the cache. Returns false in second value if the key does not exist.
func (k *Kache[K, V]) Get(key K) (V, bool) {
	defer k.lock()()
	item, ok := k.store[key]
	return item.value, ok
}

// Delete deletes a value from the cache.
func (k *Kache[K, V]) Delete(key K) {
	defer k.lock()()
	delete(k.store, key)
}

// Flush deletes all values from the cache.
func (k *Kache[K, V]) Flush() {
	defer k.lock()()
	clear(k.store)
}

// Keys returns all keys in the cache.
func (k *Kache[K, V]) Keys() []K {
	defer k.lock()()
	keys := make([]K, 0, len(k.store))
	for k := range k.store {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of values in the cache.
func (k *Kache[K, V]) Count() int {
	defer k.lock()()
	return len(k.store)
}

// Poll deletes expired values from the cache with the given poll interval. If context is cancelled, the polling stops.
func (k *Kache[K, V]) Poll(ctx context.Context, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			for key, value := range k.store {
				if value.ttl.IsZero() {
					continue
				}
				if value.ttl.After(now) {
					k.mut.Lock()
					delete(k.store, key)
					k.mut.Unlock()
				}
			}
		}
	}
}

func (k *Kache[K, V]) lock() func() {
	k.mut.Lock()
	return k.mut.Unlock
}
