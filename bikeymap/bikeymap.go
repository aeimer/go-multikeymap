package bikeymap

import (
	"errors"
	"fmt"
	"iter"
)

// BiKeyMap is a generic in-memory map with two independent keys for each value.
// The zero value is an empty map ready to use.
// A BiKeyMap is not safe for concurrent use; use ConcurrentBiKeyMap instead.
type BiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	dataByKeyA map[KeyA]V
	keyAByKeyB map[KeyB]KeyA
	keyBByKeyA map[KeyA]KeyB
}

// New returns a new empty BiKeyMap.
// It is equivalent to new(BiKeyMap[KeyA, KeyB, V]).
func New[KeyA comparable, KeyB comparable, V any]() *BiKeyMap[KeyA, KeyB, V] {
	return new(BiKeyMap[KeyA, KeyB, V])
}

// Set stores value under keyA and keyB and pairs the two keys.
// It returns an error and changes nothing if keyA or keyB is already paired with a different key.
// If both keys are already paired with each other, Set replaces the value.
func (m *BiKeyMap[KeyA, KeyB, V]) Set(keyA KeyA, keyB KeyB, value V) error {
	// Check if one key is set without the other, or if they do not point to each other.
	if existingKeyA, keyBExists := m.keyAByKeyB[keyB]; keyBExists && existingKeyA != keyA {
		return errors.New("keyB is already set with a different keyA")
	}
	if existingKeyB, keyAExists := m.keyBByKeyA[keyA]; keyAExists && existingKeyB != keyB {
		return errors.New("keyA is already set with a different keyB")
	}

	// Store the value and pair both keys.
	if m.dataByKeyA == nil {
		m.dataByKeyA = make(map[KeyA]V)
		m.keyAByKeyB = make(map[KeyB]KeyA)
		m.keyBByKeyA = make(map[KeyA]KeyB)
	}
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

// KeyA returns the keyA paired with keyB.
// The bool reports whether keyB exists.
func (m *BiKeyMap[KeyA, KeyB, V]) KeyA(keyB KeyB) (KeyA, bool) {
	keyA, exists := m.keyAByKeyB[keyB]
	return keyA, exists
}

// KeyB returns the keyB paired with keyA.
// The bool reports whether keyA exists.
func (m *BiKeyMap[KeyA, KeyB, V]) KeyB(keyA KeyA) (KeyB, bool) {
	keyB, exists := m.keyBByKeyA[keyA]
	return keyB, exists
}

// DeleteByKeyA removes the value stored under keyA and the paired keyB.
// If keyA does not exist, DeleteByKeyA is a no-op.
func (m *BiKeyMap[KeyA, KeyB, V]) DeleteByKeyA(keyA KeyA) {
	keyB, ok := m.keyBByKeyA[keyA]
	if !ok {
		return
	}
	delete(m.dataByKeyA, keyA)
	delete(m.keyAByKeyB, keyB)
	delete(m.keyBByKeyA, keyA)
}

// DeleteByKeyB removes the value stored under keyB and the paired keyA.
// If keyB does not exist, DeleteByKeyB is a no-op.
func (m *BiKeyMap[KeyA, KeyB, V]) DeleteByKeyB(keyB KeyB) {
	keyA, ok := m.keyAByKeyB[keyB]
	if !ok {
		return
	}
	delete(m.dataByKeyA, keyA)
	delete(m.keyAByKeyB, keyB)
	delete(m.keyBByKeyA, keyA)
}

// Len returns the number of values in the map.
func (m *BiKeyMap[KeyA, KeyB, V]) Len() int {
	return len(m.dataByKeyA)
}

// All returns an iterator over all keyA keys and their values.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[KeyA, KeyB, V]) All() iter.Seq2[KeyA, V] {
	return func(yield func(KeyA, V) bool) {
		for keyA, v := range m.dataByKeyA {
			if !yield(keyA, v) {
				return
			}
		}
	}
}

// KeysA returns an iterator over all keyA keys.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[KeyA, KeyB, V]) KeysA() iter.Seq[KeyA] {
	return func(yield func(KeyA) bool) {
		for keyA := range m.dataByKeyA {
			if !yield(keyA) {
				return
			}
		}
	}
}

// KeysB returns an iterator over all keyB keys.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[KeyA, KeyB, V]) KeysB() iter.Seq[KeyB] {
	return func(yield func(KeyB) bool) {
		for keyB := range m.keyAByKeyB {
			if !yield(keyB) {
				return
			}
		}
	}
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[KeyA, KeyB, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m.dataByKeyA {
			if !yield(v) {
				return
			}
		}
	}
}

// Clear removes all elements from the map.
func (m *BiKeyMap[KeyA, KeyB, V]) Clear() {
	*m = BiKeyMap[KeyA, KeyB, V]{}
}

// String returns a string representation of the map.
func (m *BiKeyMap[KeyA, KeyB, V]) String() string {
	return fmt.Sprintf("BiKeyMap: %v", m.dataByKeyA)
}
