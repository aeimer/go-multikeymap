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
// The zero value is an empty map ready to use.
// A ConcurrentMultiKeyMap must not be copied after first use.
type ConcurrentMultiKeyMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  MultiKeyMap[K, V]
}

// NewConcurrent returns a new empty ConcurrentMultiKeyMap.
// It is equivalent to new(ConcurrentMultiKeyMap[K, V]).
func NewConcurrent[K comparable, V any]() *ConcurrentMultiKeyMap[K, V] {
	return new(ConcurrentMultiKeyMap[K, V])
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

// PrimaryKey returns the primary key that the secondary key in group points to.
// The bool reports whether the secondary key exists.
func (c *ConcurrentMultiKeyMap[K, V]) PrimaryKey(group string, key string) (K, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.PrimaryKey(group, key)
}

// SecondaryKeys returns an iterator over the secondary keys of primaryKey
// as pairs of group and key.
// If primaryKey has no secondary keys, the iterator yields nothing.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentMultiKeyMap[K, V]) SecondaryKeys(primaryKey K) iter.Seq2[string, string] {
	return snapshot2(&c.mu, func() iter.Seq2[string, string] { return c.m.SecondaryKeys(primaryKey) })
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

// All returns an iterator over all primary keys and their values.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentMultiKeyMap[K, V]) All() iter.Seq2[K, V] {
	return snapshot2(&c.mu, c.m.All)
}

// Keys returns an iterator over all primary keys.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentMultiKeyMap[K, V]) Keys() iter.Seq[K] {
	return snapshot(&c.mu, c.m.Keys)
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentMultiKeyMap[K, V]) Values() iter.Seq[V] {
	return snapshot(&c.mu, c.m.Values)
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

// snapshot returns an iterator that, when iteration starts, collects seq() under the read lock
// and then yields the copy without holding the lock.
func snapshot[T any](mu *sync.RWMutex, seq func() iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		mu.RLock()
		items := slices.Collect(seq())
		mu.RUnlock()
		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// snapshot2 is snapshot for iter.Seq2.
func snapshot2[K, V any](mu *sync.RWMutex, seq func() iter.Seq2[K, V]) iter.Seq2[K, V] {
	type pair struct {
		k K
		v V
	}
	return func(yield func(K, V) bool) {
		mu.RLock()
		var items []pair
		for k, v := range seq() {
			items = append(items, pair{k, v})
		}
		mu.RUnlock()
		for _, p := range items {
			if !yield(p.k, p.v) {
				return
			}
		}
	}
}
