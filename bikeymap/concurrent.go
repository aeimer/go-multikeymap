package bikeymap

import (
	"fmt"
	"sync"
)

// ConcurrentBiKeyMap is the same as BiKeyMap, but it is safe for concurrent use.
// It uses a RWMutex to protect the map from concurrent reads and writes.
// Therefore, it is slower than BiKeyMap, but it is safe for concurrent use.
type ConcurrentBiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	mu sync.RWMutex
	m  BiKeyMap[KeyA, KeyB, V]
}

// NewConcurrent creates a new instance of ConcurrentBiKeyMap.
func NewConcurrent[KeyA comparable, KeyB comparable, V any]() *ConcurrentBiKeyMap[KeyA, KeyB, V] {
	return &ConcurrentBiKeyMap[KeyA, KeyB, V]{m: *New[KeyA, KeyB, V]()}
}

// Put stores a value with two keys.
// It returns an error and changes nothing if keyA or keyB is already paired with a different key.
// If both keys are already paired with each other, Put replaces the value.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Put(keyA KeyA, keyB KeyB, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.Put(keyA, keyB, value)
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

// RemoveByKeyA removes a value using the first key, ensuring the corresponding second key is also deleted.
// It returns an error if keyA does not exist.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) RemoveByKeyA(keyA KeyA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.RemoveByKeyA(keyA)
}

// RemoveByKeyB removes a value using the second key, ensuring the corresponding first key is also deleted.
// It returns an error if keyB does not exist.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) RemoveByKeyB(keyB KeyB) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.RemoveByKeyB(keyB)
}

// Empty checks if the map is empty.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Empty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Empty()
}

// Size returns the number of elements in the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Size()
}

// Values returns a slice of all values in the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Values()
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
