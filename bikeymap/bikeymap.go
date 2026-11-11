// Package bikeymap provides a map with two unique keys per value,
// a bidirectional map that also carries a value.
//
// Each keyA is paired with exactly one keyB, and a value can be looked up by
// either key in O(1).
// BiKeyMap is not safe for concurrent use; ConcurrentBiKeyMap is.
// The zero value of both types is an empty map ready to use.
package bikeymap

import (
	"fmt"
	"iter"
	"maps"
)

// BiKeyMap is a generic in-memory map with two independent keys for each value.
// A is the type of keyA, B the type of keyB.
// The zero value is an empty map ready to use.
// A BiKeyMap is not safe for concurrent use; use ConcurrentBiKeyMap instead.
type BiKeyMap[A comparable, B comparable, V any] struct {
	byKeyA map[A]entry[B, V]
	byKeyB map[B]A
}

// entry holds the value of a keyA together with its paired keyB.
type entry[B comparable, V any] struct {
	keyB  B
	value V
}

// New returns a new empty BiKeyMap.
// It is equivalent to new(BiKeyMap[A, B, V]).
func New[A comparable, B comparable, V any]() *BiKeyMap[A, B, V] {
	return new(BiKeyMap[A, B, V])
}

// Set stores value under keyA and keyB and pairs the two keys.
// If both keys are already paired with each other, Set replaces the value.
// If keyA is paired with a different keyB, Set removes that old keyB;
// keyA keeps its entry, now with keyB and value.
// If keyB is paired with a different keyA, Set removes that old keyA and its value.
// No key is ever left without its pair.
// Len grows by one if both keys are new, shrinks by one if both cases above apply,
// and stays the same otherwise.
func (m *BiKeyMap[A, B, V]) Set(keyA A, keyB B, value V) {
	if m.byKeyA == nil {
		m.byKeyA = make(map[A]entry[B, V])
		m.byKeyB = make(map[B]A)
	}
	// Drop the other half of each conflicting pair; the half that shares a key
	// with the new pair is overwritten below.
	if e, exists := m.byKeyA[keyA]; exists && e.keyB != keyB {
		delete(m.byKeyB, e.keyB)
	}
	if oldKeyA, exists := m.byKeyB[keyB]; exists && oldKeyA != keyA {
		delete(m.byKeyA, oldKeyA)
	}
	m.byKeyA[keyA] = entry[B, V]{keyB: keyB, value: value}
	m.byKeyB[keyB] = keyA
}

// GetByKeyA retrieves a value using the first key.
func (m *BiKeyMap[A, B, V]) GetByKeyA(keyA A) (V, bool) {
	e, exists := m.byKeyA[keyA]
	return e.value, exists
}

// GetByKeyB retrieves a value using the second key.
func (m *BiKeyMap[A, B, V]) GetByKeyB(keyB B) (V, bool) {
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
func (m *BiKeyMap[A, B, V]) KeyA(keyB B) (A, bool) {
	keyA, exists := m.byKeyB[keyB]
	return keyA, exists
}

// KeyB returns the keyB paired with keyA.
// The bool reports whether keyA exists.
func (m *BiKeyMap[A, B, V]) KeyB(keyA A) (B, bool) {
	e, exists := m.byKeyA[keyA]
	return e.keyB, exists
}

// DeleteByKeyA removes the value stored under keyA and the paired keyB.
// If keyA does not exist, DeleteByKeyA is a no-op.
func (m *BiKeyMap[A, B, V]) DeleteByKeyA(keyA A) {
	e, ok := m.byKeyA[keyA]
	if !ok {
		return
	}
	delete(m.byKeyA, keyA)
	delete(m.byKeyB, e.keyB)
}

// DeleteByKeyB removes the value stored under keyB and the paired keyA.
// If keyB does not exist, DeleteByKeyB is a no-op.
func (m *BiKeyMap[A, B, V]) DeleteByKeyB(keyB B) {
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
func (m *BiKeyMap[A, B, V]) LoadAndDeleteByKeyA(keyA A) (V, bool) {
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
func (m *BiKeyMap[A, B, V]) LoadAndDeleteByKeyB(keyB B) (V, bool) {
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
func (m *BiKeyMap[A, B, V]) Len() int {
	return len(m.byKeyA)
}

// All returns an iterator over all keyA keys and their values.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[A, B, V]) All() iter.Seq2[A, V] {
	return func(yield func(A, V) bool) {
		for keyA, e := range m.byKeyA {
			if !yield(keyA, e.value) {
				return
			}
		}
	}
}

// Pairs returns an iterator over all pairs of keyA and keyB.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[A, B, V]) Pairs() iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		for keyA, e := range m.byKeyA {
			if !yield(keyA, e.keyB) {
				return
			}
		}
	}
}

// KeysA returns an iterator over all keyA keys.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[A, B, V]) KeysA() iter.Seq[A] {
	return func(yield func(A) bool) {
		for keyA := range m.byKeyA {
			if !yield(keyA) {
				return
			}
		}
	}
}

// KeysB returns an iterator over all keyB keys.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[A, B, V]) KeysB() iter.Seq[B] {
	return func(yield func(B) bool) {
		for keyB := range m.byKeyB {
			if !yield(keyB) {
				return
			}
		}
	}
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
func (m *BiKeyMap[A, B, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, e := range m.byKeyA {
			if !yield(e.value) {
				return
			}
		}
	}
}

// DeleteFunc removes every value for which del returns true, together with its keyA and keyB.
func (m *BiKeyMap[A, B, V]) DeleteFunc(del func(A, B, V) bool) {
	for keyA, e := range m.byKeyA {
		if del(keyA, e.keyB, e.value) {
			delete(m.byKeyA, keyA)
			delete(m.byKeyB, e.keyB)
		}
	}
}

// Clone returns a copy of the map with its own keys.
// The values are copied with assignment, like maps.Clone does, so this is a shallow clone.
func (m *BiKeyMap[A, B, V]) Clone() *BiKeyMap[A, B, V] {
	return &BiKeyMap[A, B, V]{
		byKeyA: maps.Clone(m.byKeyA),
		byKeyB: maps.Clone(m.byKeyB),
	}
}

// Clear removes all elements from the map.
// Like clear on a Go map, it keeps the allocated memory,
// and an iteration that is running yields no further entries.
func (m *BiKeyMap[A, B, V]) Clear() {
	// clear instead of resetting the struct: a running range keeps the old
	// maps and would go on yielding their entries.
	clear(m.byKeyA)
	clear(m.byKeyB)
}

// String returns a string representation of the map.
func (m *BiKeyMap[A, B, V]) String() string {
	return fmt.Sprint("BiKeyMap: ", m.values())
}

// values returns a new map[A]V, which fmt prints sorted by key.
func (m *BiKeyMap[A, B, V]) values() map[A]V {
	values := make(map[A]V, len(m.byKeyA))
	for keyA, e := range m.byKeyA {
		values[keyA] = e.value
	}
	return values
}
