package multikeymap

import (
	"fmt"
	"iter"
	"slices"
	"sync"
)

// ConcurrentMultiKeyMap is the same as MultiKeyMap, but it is safe for concurrent use.
// It uses a RWMutex to protect the map from concurrent reads and writes.
// Therefore, it is slower than MultiKeyMap, but it is safe for concurrent use.
type ConcurrentMultiKeyMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  MultiKeyMap[K, V]
}

// NewConcurrent creates a new ConcurrentMultiKeyMap instance.
func NewConcurrent[K comparable, V any]() *ConcurrentMultiKeyMap[K, V] {
	return &ConcurrentMultiKeyMap[K, V]{m: *New[K, V]()}
}

// Set stores value under primaryKey.
// If primaryKey exists, Set replaces its value and keeps its secondary keys.
func (c *ConcurrentMultiKeyMap[K, V]) Set(primaryKey K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Set(primaryKey, value)
}

// AddSecondaryKeys adds secondary keys under a group for a primary key.
// It reports false and changes nothing if the primary key does not exist.
// A secondary key that already belongs to another primary key in the same group
// moves to primaryKey; the other primary key loses it.
func (c *ConcurrentMultiKeyMap[K, V]) AddSecondaryKeys(primaryKey K, group string, keys ...string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.AddSecondaryKeys(primaryKey, group, keys...)
}

// Has reports whether primaryKey exists.
func (c *ConcurrentMultiKeyMap[K, V]) Has(primaryKey K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Has(primaryKey)
}

// HasSecondaryKey checks if a secondary key exists in a specific group.
func (c *ConcurrentMultiKeyMap[K, V]) HasSecondaryKey(group string, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.HasSecondaryKey(group, key)
}

// KeyGroups returns a copy of all key groups,
// mapping group to secondary key to primary key.
// Changes to the result do not affect the map.
func (c *ConcurrentMultiKeyMap[K, V]) KeyGroups() map[string]map[string]K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.KeyGroups()
}

// Delete removes primaryKey, its value and all its secondary keys.
// If primaryKey does not exist, Delete is a no-op.
func (c *ConcurrentMultiKeyMap[K, V]) Delete(primaryKey K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Delete(primaryKey)
}

// Get returns a value by primary key.
func (c *ConcurrentMultiKeyMap[K, V]) Get(primaryKey K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Get(primaryKey)
}

// GetBySecondaryKey returns the value of the primary key that the secondary key
// in group points to. The bool reports whether such a value exists.
func (c *ConcurrentMultiKeyMap[K, V]) GetBySecondaryKey(group string, key string) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetBySecondaryKey(group, key)
}

// Len returns the number of primary keys in the map.
func (c *ConcurrentMultiKeyMap[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Len()
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentMultiKeyMap[K, V]) Values() iter.Seq[V] {
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
func (c *ConcurrentMultiKeyMap[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Clear()
}

// String returns a string representation of the map.
func (c *ConcurrentMultiKeyMap[K, V]) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("ConcurrentMultiKeyMap: %v", c.m.primary)
}
