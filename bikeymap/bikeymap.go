package bikeymap

import (
	"errors"
	"fmt"
	"iter"
)

// ErrKeyAConflict is returned by Set when keyA is already paired with a different keyB.
var ErrKeyAConflict = errors.New("bikeymap: keyA is paired with a different keyB")

// ErrKeyBConflict is returned by Set when keyB is already paired with a different keyA.
var ErrKeyBConflict = errors.New("bikeymap: keyB is paired with a different keyA")

// BiKeyMap is a generic in-memory map with two independent keys for each value.
// The zero value is an empty map ready to use.
// A BiKeyMap is not safe for concurrent use; use ConcurrentBiKeyMap instead.
type BiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	byKeyA map[KeyA]entry[KeyB, V]
	byKeyB map[KeyB]KeyA
}

// entry holds the value of a keyA together with its paired keyB.
type entry[KeyB comparable, V any] struct {
	keyB  KeyB
	value V
}

// New returns a new empty BiKeyMap.
// It is equivalent to new(BiKeyMap[KeyA, KeyB, V]).
func New[KeyA comparable, KeyB comparable, V any]() *BiKeyMap[KeyA, KeyB, V] {
	return new(BiKeyMap[KeyA, KeyB, V])
}

// Set stores value under keyA and keyB and pairs the two keys.
// If keyB is already paired with a different keyA, Set returns ErrKeyBConflict.
// Otherwise, if keyA is already paired with a different keyB, Set returns ErrKeyAConflict.
// On an error Set changes nothing.
// If both keys are already paired with each other, Set replaces the value.
func (m *BiKeyMap[KeyA, KeyB, V]) Set(keyA KeyA, keyB KeyB, value V) error {
	// Check if one key is set without the other, or if they do not point to each other.
	if existingKeyA, keyBExists := m.byKeyB[keyB]; keyBExists && existingKeyA != keyA {
		return ErrKeyBConflict
	}
	if e, keyAExists := m.byKeyA[keyA]; keyAExists && e.keyB != keyB {
		return ErrKeyAConflict
	}

	// Store the value and pair both keys.
	if m.byKeyA == nil {
		m.byKeyA = make(map[KeyA]entry[KeyB, V])
		m.byKeyB = make(map[KeyB]KeyA)
	}
	m.byKeyA[keyA] = entry[KeyB, V]{keyB: keyB, value: value}
	m.byKeyB[keyB] = keyA
	return nil
}

// GetByKeyA retrieves a value using the first key.
func (m *BiKeyMap[KeyA, KeyB, V]) GetByKeyA(keyA KeyA) (V, bool) {
	e, exists := m.byKeyA[keyA]
	return e.value, exists
}

// GetByKeyB retrieves a value using the second key.
func (m *BiKeyMap[KeyA, KeyB, V]) GetByKeyB(keyB KeyB) (V, bool) {
	keyA, exists := m.byKeyB[keyB]
	if !exists {
		var zero V
		return zero, false
	}

	e, exists := m.byKeyA[keyA]
	return e.value, exists
}

// KeyA returns the keyA paired with keyB.
// The bool reports whether keyB exists.
func (m *BiKeyMap[KeyA, KeyB, V]) KeyA(keyB KeyB) (KeyA, bool) {
	keyA, exists := m.byKeyB[keyB]
	return keyA, exists
}

// KeyB returns the keyB paired with keyA.
// The bool reports whether keyA exists.
func (m *BiKeyMap[KeyA, KeyB, V]) KeyB(keyA KeyA) (KeyB, bool) {
	e, exists := m.byKeyA[keyA]
	return e.keyB, exists
}

// DeleteByKeyA removes the value stored under keyA and the paired keyB.
// If keyA does not exist, DeleteByKeyA is a no-op.
func (m *BiKeyMap[KeyA, KeyB, V]) DeleteByKeyA(keyA KeyA) {
	e, ok := m.byKeyA[keyA]
	if !ok {
		return
	}
	delete(m.byKeyA, keyA)
	delete(m.byKeyB, e.keyB)
}

// DeleteByKeyB removes the value stored under keyB and the paired keyA.
// If keyB does not exist, DeleteByKeyB is a no-op.
func (m *BiKeyMap[KeyA, KeyB, V]) DeleteByKeyB(keyB KeyB) {
	keyA, ok := m.byKeyB[keyB]
	if !ok {
		return
	}
	delete(m.byKeyA, keyA)
	delete(m.byKeyB, keyB)
}

// LoadAndDeleteByKeyA removes the value stored under keyA and the paired keyB
// and returns the previous value, if any.
// The bool reports whether keyA existed.
func (m *BiKeyMap[KeyA, KeyB, V]) LoadAndDeleteByKeyA(keyA KeyA) (V, bool) {
	e, ok := m.byKeyA[keyA]
	if ok {
		delete(m.byKeyA, keyA)
		delete(m.byKeyB, e.keyB)
	}
	return e.value, ok
}

// LoadAndDeleteByKeyB removes the value stored under keyB and the paired keyA
// and returns the previous value, if any.
// The bool reports whether keyB existed.
func (m *BiKeyMap[KeyA, KeyB, V]) LoadAndDeleteByKeyB(keyB KeyB) (V, bool) {
	keyA, ok := m.byKeyB[keyB]
	if !ok {
		var zero V
		return zero, false
	}
	e := m.byKeyA[keyA]
	delete(m.byKeyA, keyA)
	delete(m.byKeyB, keyB)
	return e.value, true
}

// Len returns the number of values in the map.
func (m *BiKeyMap[KeyA, KeyB, V]) Len() int {
	return len(m.byKeyA)
}

// All returns an iterator over all keyA keys and their values.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[KeyA, KeyB, V]) All() iter.Seq2[KeyA, V] {
	return func(yield func(KeyA, V) bool) {
		for keyA, e := range m.byKeyA {
			if !yield(keyA, e.value) {
				return
			}
		}
	}
}

// KeysA returns an iterator over all keyA keys.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[KeyA, KeyB, V]) KeysA() iter.Seq[KeyA] {
	return func(yield func(KeyA) bool) {
		for keyA := range m.byKeyA {
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
		for keyB := range m.byKeyB {
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
		for _, e := range m.byKeyA {
			if !yield(e.value) {
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
	return "BiKeyMap: " + m.format()
}

// format formats keyA and values like a map[KeyA]V.
func (m *BiKeyMap[KeyA, KeyB, V]) format() string {
	values := make(map[KeyA]V, len(m.byKeyA))
	for keyA, e := range m.byKeyA {
		values[keyA] = e.value
	}
	return fmt.Sprint(values)
}
