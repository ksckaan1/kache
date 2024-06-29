package kache

type cleanStrategy uint

const (
	cleanStrategyNone cleanStrategy = iota
	cleanStrategyLRU
	cleanStrategyFIFO
)

// WithCleanStrategyLRU sets the clean strategy to LRU and Deletes the specified number of records, if possible, when the maximum number of records is reached
//
// LRU: Least Recently Used
//
// maxRecordCount: maximum number of records
//
// recordsToClean: number of records to clean
//
// example: WithCleanStrategyLRU(10000, 1000), if record number reaches 10000, it will delete 1000 records
func (k *Kache[K, V]) WithCleanStrategyLRU(maxRecordCount int, recordsToClean int) *Kache[K, V] {
	k.cleanStrategy = cleanStrategyLRU
	k.maxRecordNum = maxRecordCount
	k.cleanNum = recordsToClean
	return k
}

// WithCleanStrategyFIFO sets the clean strategy to FIFO and Deletes the specified number of records, if possible, when the maximum number of records is reached
//
// FIFO: First In First Out
//
// maxRecordCount: maximum number of records
//
// recordsToClean: number of records to clean
//
// example: WithCleanStrategyFIFO(10000, 1000), if record number reaches 10000, it will delete 1000 records
func (k *Kache[K, V]) WithCleanStrategyFIFO(maxRecordCount int, recordsToClean int) *Kache[K, V] {
	k.cleanStrategy = cleanStrategyFIFO
	k.maxRecordNum = maxRecordCount
	k.cleanNum = recordsToClean
	return k
}

// WithCleanStrategyNone sets the clean strategy to none, so no records will be deleted,
// except for those who use ttl
func (k *Kache[K, V]) WithCleanStrategyNone() *Kache[K, V] {
	k.cleanStrategy = cleanStrategyNone
	k.maxRecordNum = -1
	k.cleanNum = -1
	return k
}

func (k *Kache[K, V]) clean() {
	currentElem := k.elems.Front()
	deletedCount := 0
	for {
		if deletedCount >= k.cleanNum || currentElem == nil {
			break
		}
		delete(k.store, currentElem.Value.(*valueWithTTL[K, V]).key)
		nextElem := currentElem.Next()
		k.elems.Remove(currentElem)
		deletedCount++
		currentElem = nextElem
	}
}
