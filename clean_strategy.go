package kache

type ReplacementStrategy uint

const (
	ReplacementStrategyNone ReplacementStrategy = iota
	ReplacementStrategyLRU                      // Least Recently Used
	ReplacementStrategyMRU                      // Most Recently Used
	ReplacementStrategyFIFO                     // First In First Out
	ReplacementStrategyLIFO                     // Last In First Out
	ReplacementStrategyLFU                      // Least Frequently Used
	ReplacementStrategyMFU                      // Most Frequently Used
)

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
