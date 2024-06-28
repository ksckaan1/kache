package kache

import (
	"context"
	"sync"
	"time"
)

type valueWithTTL[K comparable, V any] struct {
	key         K
	value       V
	setTime     time.Time
	ttl         time.Duration
	lastHitTime time.Time
}

type Kache[K comparable, V any] struct {
	store         map[K]*valueWithTTL[K, V]
	mut           *sync.Mutex
	maxRecordNum  int
	cleanNum      int
	cleanStrategy cleanStrategy
}

func New[K comparable, V any]() *Kache[K, V] {
	k := &Kache[K, V]{
		store:        make(map[K]*valueWithTTL[K, V]),
		mut:          new(sync.Mutex),
		maxRecordNum: -1,
	}
	return k
}

// Set sets a value in the cache.
func (k *Kache[K, V]) Set(key K, v V) {
	k.set(key, v, 0)
}

// SetWithTTL sets a value in the cache with a TTL. If the TTL is 0, the value will not expire.
func (k *Kache[K, V]) SetWithTTL(key K, v V, ttl time.Duration) {
	k.set(key, v, ttl)
}

// Get gets a value from the cache. Returns false in second value if the key does not exist.
func (k *Kache[K, V]) Get(key K) (V, bool) {
	defer k.lock()()
	item, ok := k.store[key]
	if !ok {
		return *new(V), false
	}
	item.lastHitTime = time.Now()
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
			for key := range k.store {
				if k.store[key].ttl == 0 {
					continue
				}
				if k.store[key].setTime.Add(k.store[key].ttl).After(now) {
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

func (k *Kache[K, V]) set(key K, v V, ttl time.Duration) {
	defer k.lock()()
	if k.maxRecordNum > 0 && len(k.store) >= k.maxRecordNum {
		k.clean()
	}
	now := time.Now()
	k.store[key] = &valueWithTTL[K, V]{
		key:         key,
		value:       v,
		ttl:         ttl,
		lastHitTime: now,
		setTime:     now,
	}
}

func (k *Kache[K, V]) clean() {
	switch k.cleanStrategy {
	case cleanStrategyLRU:
		k.cleanLRU()
	case cleanStrategyFIFO:
		k.cleanFIFO()
	}
}
