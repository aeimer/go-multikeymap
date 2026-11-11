package bikeymap

import (
	"fmt"
	"iter"
	"slices"
	"sync"
)

// ConcurrentBiKeyMap is the same as BiKeyMap, but it is safe for concurrent use.
// It uses a RWMutex to protect the map from concurrent reads and writes.
// Therefore, it is slower than BiKeyMap, but it is safe for concurrent use.
// The zero value is an empty map ready to use.
// A ConcurrentBiKeyMap must not be copied after first use.
type ConcurrentBiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	mu sync.RWMutex
	m  BiKeyMap[KeyA, KeyB, V]
}

// NewConcurrent returns a new empty ConcurrentBiKeyMap.
// It is equivalent to new(ConcurrentBiKeyMap[KeyA, KeyB, V]).
func NewConcurrent[KeyA comparable, KeyB comparable, V any]() *ConcurrentBiKeyMap[KeyA, KeyB, V] {
	return new(ConcurrentBiKeyMap[KeyA, KeyB, V])
}

// Set stores value under keyA and keyB and pairs the two keys.
// It returns an error and changes nothing if keyA or keyB is already paired with a different key.
// If both keys are already paired with each other, Set replaces the value.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Set(keyA KeyA, keyB KeyB, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.Set(keyA, keyB, value)
}

// GetByKeyA retrieves a value using the first key.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) GetByKeyA(keyA KeyA) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetByKeyA(keyA)
}

// GetByKeyB retrieves a value using the second key.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) GetByKeyB(keyB KeyB) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetByKeyB(keyB)
}

// DeleteByKeyA removes the value stored under keyA and the paired keyB.
// If keyA does not exist, DeleteByKeyA is a no-op.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) DeleteByKeyA(keyA KeyA) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteByKeyA(keyA)
}

// DeleteByKeyB removes the value stored under keyB and the paired keyA.
// If keyB does not exist, DeleteByKeyB is a no-op.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) DeleteByKeyB(keyB KeyB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteByKeyB(keyB)
}

// Len returns the number of values in the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Len()
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		c.mu.RLock()
		values := slices.Collect(c.m.Values())
		c.mu.RUnlock()
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// Clear removes all elements from the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Clear()
}

// String returns a string representation of the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("ConcurrentBiKeyMap: %v", c.m.dataByKeyA)
}
