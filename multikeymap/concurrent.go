package multikeymap

import (
	"fmt"
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

// Put inserts a value with a primary key.
func (c *ConcurrentMultiKeyMap[K, V]) Put(primaryKey K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Put(primaryKey, value)
}

// PutSecondaryKeys adds secondary keys under a group for a primary key.
func (c *ConcurrentMultiKeyMap[K, V]) PutSecondaryKeys(primaryKey K, group string, keys ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.PutSecondaryKeys(primaryKey, group, keys...)
}

// HasPrimaryKey checks if a primary key exists.
func (c *ConcurrentMultiKeyMap[K, V]) HasPrimaryKey(primaryKey K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.HasPrimaryKey(primaryKey)
}

// HasSecondaryKey checks if a secondary key exists in a specific group.
func (c *ConcurrentMultiKeyMap[K, V]) HasSecondaryKey(group string, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.HasSecondaryKey(group, key)
}

// GetAllKeyGroups returns all key groups and their secondary keys.
func (c *ConcurrentMultiKeyMap[K, V]) GetAllKeyGroups() map[string]map[string]K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetAllKeyGroups()
}

// Remove removes a primary key and its associated secondary keys.
func (c *ConcurrentMultiKeyMap[K, V]) Remove(primaryKey K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Remove(primaryKey)
}

// Get returns a value by primary key.
func (c *ConcurrentMultiKeyMap[K, V]) Get(primaryKey K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Get(primaryKey)
}

// GetBySecondaryKey returns a primary key by secondary key and group.
func (c *ConcurrentMultiKeyMap[K, V]) GetBySecondaryKey(group string, key string) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetBySecondaryKey(group, key)
}

// Size returns the number of primary keys in the map.
func (c *ConcurrentMultiKeyMap[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Size()
}

// Empty checks if the map is empty.
func (c *ConcurrentMultiKeyMap[K, V]) Empty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Empty()
}

// Values returns a slice of all values in the map.
func (c *ConcurrentMultiKeyMap[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Values()
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
