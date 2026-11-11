package bikeymap

import (
	"errors"
	"fmt"
)

// BiKeyMap is a generic in-memory map with two independent keys for each value.
// A BiKeyMap is not safe for concurrent use; use ConcurrentBiKeyMap instead.
type BiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	dataByKeyA map[KeyA]V
	keyAByKeyB map[KeyB]KeyA
	keyBByKeyA map[KeyA]KeyB
}

// New creates a new instance of BiKeyMap.
func New[KeyA comparable, KeyB comparable, V any]() *BiKeyMap[KeyA, KeyB, V] {
	return &BiKeyMap[KeyA, KeyB, V]{
		dataByKeyA: make(map[KeyA]V),
		keyAByKeyB: make(map[KeyB]KeyA),
		keyBByKeyA: make(map[KeyA]KeyB),
	}
}

// Put stores a value with two keys.
// It returns an error and changes nothing if keyA or keyB is already paired with a different key.
// If both keys are already paired with each other, Put replaces the value.
func (m *BiKeyMap[KeyA, KeyB, V]) Put(keyA KeyA, keyB KeyB, value V) error {
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
func (m *BiKeyMap[KeyA, KeyB, V]) GetByKeyA(keyA KeyA) (V, bool) {
	value, exists := m.dataByKeyA[keyA]
	return value, exists
}

// GetByKeyB retrieves a value using the second key.
func (m *BiKeyMap[KeyA, KeyB, V]) GetByKeyB(keyB KeyB) (V, bool) {
	keyA, exists := m.keyAByKeyB[keyB]
	if !exists {
		var zero V
		return zero, false
	}

	value, exists := m.dataByKeyA[keyA]
	return value, exists
}

// RemoveByKeyA removes a value using the first key, ensuring the corresponding second key is also deleted.
// It returns an error if keyA does not exist.
func (m *BiKeyMap[KeyA, KeyB, V]) RemoveByKeyA(keyA KeyA) error {
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
// It returns an error if keyB does not exist.
func (m *BiKeyMap[KeyA, KeyB, V]) RemoveByKeyB(keyB KeyB) error {
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
func (m *BiKeyMap[KeyA, KeyB, V]) Empty() bool {
	return len(m.dataByKeyA) == 0
}

// Size returns the number of elements in the map.
func (m *BiKeyMap[KeyA, KeyB, V]) Size() int {
	return len(m.dataByKeyA)
}

// Values returns a slice of all values in the map.
func (m *BiKeyMap[KeyA, KeyB, V]) Values() []V {
	values := make([]V, 0, len(m.dataByKeyA))
	for _, value := range m.dataByKeyA {
		values = append(values, value)
	}
	return values
}

// Clear removes all elements from the map.
func (m *BiKeyMap[KeyA, KeyB, V]) Clear() {
	m.dataByKeyA = make(map[KeyA]V)
	m.keyAByKeyB = make(map[KeyB]KeyA)
	m.keyBByKeyA = make(map[KeyA]KeyB)
}

// String returns a string representation of the map.
func (m *BiKeyMap[KeyA, KeyB, V]) String() string {
	return fmt.Sprintf("BiKeyMap: %v", m.dataByKeyA)
}
