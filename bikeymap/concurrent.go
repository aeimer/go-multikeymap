package bikeymap

import (
	"errors"
	"fmt"
	"sync"
)

// ConcurrentBiKeyMap is the same as BiKeyMap, but it is safe for concurrent use.
// It uses a RWMutex to protect the map from concurrent reads and writes.
// Therefore, it is slower than BiKeyMap, but it is safe for concurrent use.
type ConcurrentBiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	mu sync.RWMutex
	BiKeyMap[KeyA, KeyB, V]
}

// NewConcurrent creates a new instance of ConcurrentBiKeyMap.
func NewConcurrent[KeyA comparable, KeyB comparable, V any]() *ConcurrentBiKeyMap[KeyA, KeyB, V] {
	return &ConcurrentBiKeyMap[KeyA, KeyB, V]{
		BiKeyMap: BiKeyMap[KeyA, KeyB, V]{
			dataByKeyA: make(map[KeyA]V),
			keyAByKeyB: make(map[KeyB]KeyA),
			keyBByKeyA: make(map[KeyA]KeyB),
		},
	}
}

// Put stores a value with two keys. It only fails if one of the keys is already set without the other.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) Put(keyA KeyA, keyB KeyB, value V) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if one key is set without the other, or if they do not point to each other.
	if existingKeyA, keyBExists := m.keyAByKeyB[keyB]; keyBExists && existingKeyA != keyA {
		return errors.New("keyB is already set with a different keyA")
	}
	if existingKeyB, keyAExists := m.keyBByKeyA[keyA]; keyAExists && existingKeyB != keyB {
		return errors.New("keyA is already set with a different keyB")
	}

	// Put the new values for both keys.
	m.dataByKeyA[keyA] = value
	m.keyAByKeyB[keyB] = keyA
	m.keyBByKeyA[keyA] = keyB
	return nil
}

// GetByKeyA retrieves a value using the first key.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) GetByKeyA(keyA KeyA) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.dataByKeyA[keyA]
	return value, exists
}

// GetByKeyB retrieves a value using the second key.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) GetByKeyB(keyB KeyB) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyA, exists := m.keyAByKeyB[keyB]
	if !exists {
		var zero V
		return zero, false
	}

	value, exists := m.dataByKeyA[keyA]
	return value, exists
}

// RemoveByKeyA removes a value using the first key, ensuring the corresponding second key is also deleted.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) RemoveByKeyA(keyA KeyA) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify that keyA exists and retrieve the associated keyB.
	keyB, ok := m.keyBByKeyA[keyA]
	if !ok {
		return errors.New("keyA does not exist")
	}

	// Remove keyA, keyB, and the associated value.
	delete(m.dataByKeyA, keyA)
	delete(m.keyAByKeyB, keyB)
	delete(m.keyBByKeyA, keyA)

	return nil
}

// RemoveByKeyB removes a value using the second key, ensuring the corresponding first key is also deleted.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) RemoveByKeyB(keyB KeyB) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify that keyB exists and retrieve the associated keyA.
	keyA, ok := m.keyAByKeyB[keyB]
	if !ok {
		return errors.New("keyB does not exist")
	}

	// Remove keyA, keyB, and the associated value.
	delete(m.dataByKeyA, keyA)
	delete(m.keyAByKeyB, keyB)
	delete(m.keyBByKeyA, keyA)

	return nil
}

// Empty checks if the map is empty.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) Empty() bool {
	return len(m.dataByKeyA) == 0
}

// Size returns the number of elements in the map.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) Size() int {
	return len(m.dataByKeyA)
}

// Values returns a slice of all values in the map.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := make([]V, 0, len(m.dataByKeyA))
	for _, value := range m.dataByKeyA {
		values = append(values, value)
	}
	return values
}

// Clear removes all elements from the map.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dataByKeyA = make(map[KeyA]V)
	m.keyAByKeyB = make(map[KeyB]KeyA)
	m.keyBByKeyA = make(map[KeyA]KeyB)
}

// String returns a string representation of the map.
func (m *ConcurrentBiKeyMap[KeyA, KeyB, V]) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("ConcurrentBiKeyMap: %v", m.dataByKeyA)
}
