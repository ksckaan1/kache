package kache

import (
	"slices"
)

type cleanStrategy uint

const (
	cleanStrategyLRU cleanStrategy = iota
	cleanStrategyFIFO
)

func (k *Kache[K, V]) WithCleanStrategyLRU(maxRecordCount int, recordsToClean int) *Kache[K, V] {
	k.cleanStrategy = cleanStrategyLRU
	k.maxRecordNum = maxRecordCount
	k.cleanNum = recordsToClean
	return k
}

func (k *Kache[K, V]) WithCleanStrategyFIFO(maxRecordCount int, recordsToClean int) *Kache[K, V] {
	k.cleanStrategy = cleanStrategyFIFO
	k.maxRecordNum = maxRecordCount
	k.cleanNum = recordsToClean
	return k
}

func (k *Kache[K, V]) cleanLRU() {
	list := make([]*valueWithTTL[K, V], 0, len(k.store))
	for key := range k.store {
		list = append(list, k.store[key])
	}
	slices.SortFunc(list, func(a, b *valueWithTTL[K, V]) int {
		if a.lastHitTime.Before(b.lastHitTime) {
			return -1
		}
		return 1
	})
	for i := 0; i < k.cleanNum && i < len(list); i++ {
		delete(k.store, list[i].key)
	}
}

func (k *Kache[K, V]) cleanFIFO() {
	list := make([]*valueWithTTL[K, V], 0, len(k.store))
	for key := range k.store {
		list = append(list, k.store[key])
	}
	slices.SortFunc(list, func(a, b *valueWithTTL[K, V]) int {
		if a.setTime.Before(b.setTime) {
			return -1
		}
		return 1
	})
	for i := 0; i < k.cleanNum && i < len(list); i++ {
		delete(k.store, list[i].key)
	}
}
