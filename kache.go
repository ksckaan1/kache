package kache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type valueWithTTL[K comparable, V any] struct {
	key        K
	value      V
	hitCount   uint
	expireTime time.Time
}

type Kache[K comparable, V any] struct {
	elems               *list.List // front of list == greater risk of deletion <---------list---------> back of list == less risk of deletion
	store               map[K]*list.Element
	mut                 *sync.Mutex
	maxRecordThreshold  int
	cleanNum            int
	replacementStrategy ReplacementStrategy
	pollInterval        time.Duration
	pollCancel          chan struct{}
}

type Config struct {
	ReplacementStrategy ReplacementStrategy // default: ReplacementStrategyNone
	MaxRecordTreshold   int                 // This parameter is used to control the maximum number of records in the cache. If the number of records exceeds this threshold, records will be deleted according to the replacement strategy.
	CleanNum            int                 // This parameter is used to control the number of records to be deleted.
	PollInterval        time.Duration       // This parameter is used to control the polling interval. If value is 0, uses default = 1 second.
}

func New[K comparable, V any](cfg Config) *Kache[K, V] {
	pollInterval := time.Second
	if cfg.PollInterval > 0 {
		pollInterval = cfg.PollInterval
	}

	k := &Kache[K, V]{
		elems:               list.New(),
		store:               make(map[K]*list.Element),
		mut:                 new(sync.Mutex),
		maxRecordThreshold:  cfg.MaxRecordTreshold,
		cleanNum:            cfg.CleanNum,
		replacementStrategy: cfg.ReplacementStrategy,
		pollInterval:        pollInterval,
		pollCancel:          make(chan struct{}),
	}

	go k.poll()

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

	value := item.Value.(*valueWithTTL[K, V])

	if k.replacementStrategy == ReplacementStrategyMFU || k.replacementStrategy == ReplacementStrategyLFU {
		value.hitCount++
	}

	switch k.replacementStrategy {
	case ReplacementStrategyLRU:
		k.elems.MoveToBack(item)
	case ReplacementStrategyMRU:
		k.elems.MoveToFront(item)
	case ReplacementStrategyMFU, ReplacementStrategyLFU:
		k.moveByHits(item)
	}

	return value.value, true
}

// Delete deletes a value from the cache.
func (k *Kache[K, V]) Delete(key K) {
	defer k.lock()()
	k.elems.Remove(k.store[key])
	delete(k.store, key)
}

// Flush deletes all values from the cache.
func (k *Kache[K, V]) Flush() {
	defer k.lock()()
	k.elems.Init()
	clear(k.store)
}

// Keys returns all keys in the cache.
func (k *Kache[K, V]) Keys() []K {
	defer k.lock()()
	keys := make([]K, 0, len(k.store))
	for e := k.elems.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*valueWithTTL[K, V]).key)
	}
	return keys
}

// Count returns the number of values in the cache.
func (k *Kache[K, V]) Count() int {
	defer k.lock()()
	return k.elems.Len()
}

func (k *Kache[K, V]) Close() {
	close(k.pollCancel)
}

// Poll deletes expired values from the cache with the given poll interval. If context is cancelled, the polling stops.
func (k *Kache[K, V]) poll() {
	fmt.Println("polling", k.pollInterval)
	ticker := time.NewTicker(k.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.pollCancel:
			fmt.Println("poll canceled")
			return
		case <-ticker.C:
			fmt.Println("polling")
			k.mut.Lock()
			now := time.Now()
			for key := range k.store {
				if k.store[key].Value.(*valueWithTTL[K, V]).expireTime.IsZero() {
					continue
				}
				if elem := k.store[key]; elem.Value.(*valueWithTTL[K, V]).expireTime.After(now) {
					k.elems.Remove(elem)
					delete(k.store, key)
				}
			}
			k.mut.Unlock()
		}
	}

	fmt.Println("poll stopped")
}

func (k *Kache[K, V]) lock() func() {
	k.mut.Lock()
	return k.mut.Unlock
}

func (k *Kache[K, V]) set(key K, v V, ttl time.Duration) {
	defer k.lock()()
	if k.maxRecordThreshold > 0 && k.cleanNum > 0 && k.replacementStrategy > ReplacementStrategyNone && len(k.store) >= k.maxRecordThreshold {
		k.clean()
	}

	value := &valueWithTTL[K, V]{
		key:        key,
		value:      v,
		expireTime: time.Now().Add(ttl),
	}

	// if exists
	if oldElem, ok := k.store[key]; ok {
		oldElem.Value = value
		switch k.replacementStrategy {
		case ReplacementStrategyLRU:
			k.elems.MoveToBack(oldElem)
		case ReplacementStrategyMRU:
			k.elems.MoveToFront(oldElem)
		}
		return
	}

	// if not exists
	switch k.replacementStrategy {
	case ReplacementStrategyFIFO, ReplacementStrategyLRU, ReplacementStrategyLFU, ReplacementStrategyMFU:
		k.store[key] = k.elems.PushBack(value)
	case ReplacementStrategyLIFO, ReplacementStrategyMRU:
		k.store[key] = k.elems.PushFront(value)
	}
}

func (k *Kache[K, V]) moveByHits(elem *list.Element) {
	prev := elem.Prev()
	next := elem.Next()
	switch k.replacementStrategy {
	case ReplacementStrategyLFU:
		if prev != nil && prev.Value.(*valueWithTTL[K, V]).hitCount > elem.Value.(*valueWithTTL[K, V]).hitCount {
			k.elems.MoveBefore(elem, prev)
			k.moveByHits(elem)
			return
		}

		if next != nil && next.Value.(*valueWithTTL[K, V]).hitCount < elem.Value.(*valueWithTTL[K, V]).hitCount {
			k.elems.MoveAfter(elem, next)
			k.moveByHits(elem)
		}
	case ReplacementStrategyMFU:
		if prev != nil && prev.Value.(*valueWithTTL[K, V]).hitCount < elem.Value.(*valueWithTTL[K, V]).hitCount {
			k.elems.MoveBefore(elem, prev)
			k.moveByHits(elem)
			return
		}

		if next != nil && next.Value.(*valueWithTTL[K, V]).hitCount > elem.Value.(*valueWithTTL[K, V]).hitCount {
			k.elems.MoveAfter(elem, next)
			k.moveByHits(elem)
		}
	}
}
