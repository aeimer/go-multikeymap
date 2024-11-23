package bikeymap

import (
	"errors"
	"fmt"
	"sync"
)

// BiKeyMap is a generic in-memory map with two independent keys for each value.
// It implements container/Container.
type BiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	mu         sync.RWMutex
	dataByKeyA map[KeyA]V
	keyAByKeyB map[KeyB]KeyA
	keyBByKeyA map[KeyA]KeyB
}

// NewBiKeyMap creates a new instance of BiKeyMap.
func NewBiKeyMap[KeyA comparable, KeyB comparable, V any]() *BiKeyMap[KeyA, KeyB, V] {
	return &BiKeyMap[KeyA, KeyB, V]{
		dataByKeyA: make(map[KeyA]V),
		keyAByKeyB: make(map[KeyB]KeyA),
		keyBByKeyA: make(map[KeyA]KeyB),
	}
}

// Put stores a value with two keys. It only fails if one of the keys is already set without the other.
func (c *BiKeyMap[KeyA, KeyB, V]) Put(keyA KeyA, keyB KeyB, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if one key is set without the other, or if they do not point to each other.
	if existingKeyA, keyBExists := c.keyAByKeyB[keyB]; keyBExists && existingKeyA != keyA {
		return errors.New("keyB is already set with a different keyA")
	}
	if existingKeyB, keyAExists := c.keyBByKeyA[keyA]; keyAExists && existingKeyB != keyB {
		return errors.New("keyA is already set with a different keyB")
	}

	// Put the new values for both keys.
	c.dataByKeyA[keyA] = value
	c.keyAByKeyB[keyB] = keyA
	c.keyBByKeyA[keyA] = keyB
	return nil
}

// GetByKeyA retrieves a value using the first key.
func (c *BiKeyMap[KeyA, KeyB, V]) GetByKeyA(keyA KeyA) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.dataByKeyA[keyA]
	return value, exists
}

// GetByKeyB retrieves a value using the second key.
func (c *BiKeyMap[KeyA, KeyB, V]) GetByKeyB(keyB KeyB) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keyA, exists := c.keyAByKeyB[keyB]
	if !exists {
		var zero V
		return zero, false
	}

	value, exists := c.dataByKeyA[keyA]
	return value, exists
}

// RemoveByKeyA removes a value using the first key, ensuring the corresponding second key is also deleted.
func (c *BiKeyMap[KeyA, KeyB, V]) RemoveByKeyA(keyA KeyA) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify that keyA exists and retrieve the associated keyB.
	keyB, ok := c.keyBByKeyA[keyA]
	if !ok {
		return errors.New("keyA does not exist")
	}

	// Remove keyA, keyB, and the associated value.
	delete(c.dataByKeyA, keyA)
	delete(c.keyAByKeyB, keyB)
	delete(c.keyBByKeyA, keyA)

	return nil
}

// RemoveByKeyB removes a value using the second key, ensuring the corresponding first key is also deleted.
func (c *BiKeyMap[KeyA, KeyB, V]) RemoveByKeyB(keyB KeyB) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify that keyB exists and retrieve the associated keyA.
	keyA, ok := c.keyAByKeyB[keyB]
	if !ok {
		return errors.New("keyB does not exist")
	}

	// Remove keyA, keyB, and the associated value.
	delete(c.dataByKeyA, keyA)
	delete(c.keyAByKeyB, keyB)
	delete(c.keyBByKeyA, keyA)

	return nil
}

// Empty checks if the map is empty.
func (c *BiKeyMap[KeyA, KeyB, V]) Empty() bool {
	return len(c.dataByKeyA) == 0
}

// Size returns the number of elements in the map.
func (c *BiKeyMap[KeyA, KeyB, V]) Size() int {
	return len(c.dataByKeyA)
}

// Values returns a slice of all values in the map.
func (c *BiKeyMap[KeyA, KeyB, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.dataByKeyA))
	for _, value := range c.dataByKeyA {
		values = append(values, value)
	}
	return values
}

// Clear removes all elements from the map.
func (c *BiKeyMap[KeyA, KeyB, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dataByKeyA = make(map[KeyA]V)
	c.keyAByKeyB = make(map[KeyB]KeyA)
	c.keyBByKeyA = make(map[KeyA]KeyB)
}

// String returns a string representation of the map.
func (c *BiKeyMap[KeyA, KeyB, V]) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("BiKeyMap: %v", c.dataByKeyA)
}
