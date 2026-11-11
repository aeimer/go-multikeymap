package multikeymap

import (
	"fmt"
	"iter"
)

// MultiKeyMap is a generic in-memory map with a primary key and multiple secondary keys.
// Secondary keys are strings, organised in named groups.
// The zero value is an empty map ready to use.
// A MultiKeyMap is not safe for concurrent use; use ConcurrentMultiKeyMap instead.
type MultiKeyMap[K comparable, V any] struct {
	primary     map[K]V
	secondary   map[string]map[string]K              // Group -> SecondaryKey -> PrimaryKey
	secondaryTo map[K]map[string]map[string]struct{} // PrimaryKey -> Group -> set of SecondaryKeys
}

// New returns a new empty MultiKeyMap.
// It is equivalent to new(MultiKeyMap[K, V]).
func New[K comparable, V any]() *MultiKeyMap[K, V] {
	return new(MultiKeyMap[K, V])
}

// Set stores value under primaryKey.
// If primaryKey exists, Set replaces its value and keeps its secondary keys.
func (m *MultiKeyMap[K, V]) Set(primaryKey K, value V) {
	if m.primary == nil {
		m.primary = make(map[K]V)
	}
	m.primary[primaryKey] = value
}

// AddSecondaryKeys adds secondary keys under a group for a primary key.
// It reports false and changes nothing if the primary key does not exist.
// A secondary key that already belongs to another primary key in the same group
// moves to primaryKey; the other primary key loses it.
func (m *MultiKeyMap[K, V]) AddSecondaryKeys(primaryKey K, group string, keys ...string) bool {
	if _, exists := m.primary[primaryKey]; !exists {
		return false
	}
	if len(keys) == 0 {
		return true
	}
	if m.secondary == nil {
		m.secondary = make(map[string]map[string]K)
		m.secondaryTo = make(map[K]map[string]map[string]struct{})
	}
	if m.secondary[group] == nil {
		m.secondary[group] = make(map[string]K)
	}
	groups := m.secondaryTo[primaryKey]
	if groups == nil {
		groups = make(map[string]map[string]struct{})
		m.secondaryTo[primaryKey] = groups
	}
	if groups[group] == nil {
		groups[group] = make(map[string]struct{})
	}
	for _, key := range keys {
		if owner, exists := m.secondary[group][key]; exists && owner != primaryKey {
			m.removeReverseEntry(owner, group, key)
		}
		m.secondary[group][key] = primaryKey
		groups[group][key] = struct{}{}
	}
	return true
}

// removeReverseEntry drops key from the reverse index of primaryKey
// and removes the group and primary key entries that become empty.
func (m *MultiKeyMap[K, V]) removeReverseEntry(primaryKey K, group string, key string) {
	groups := m.secondaryTo[primaryKey]
	delete(groups[group], key)
	if len(groups[group]) == 0 {
		delete(groups, group)
	}
	if len(groups) == 0 {
		delete(m.secondaryTo, primaryKey)
	}
}

// Has reports whether primaryKey exists.
func (m *MultiKeyMap[K, V]) Has(primaryKey K) bool {
	_, exists := m.primary[primaryKey]
	return exists
}

// HasSecondaryKey checks if a secondary key exists in a specific group.
func (m *MultiKeyMap[K, V]) HasSecondaryKey(group string, key string) bool {
	if groupKeys, exists := m.secondary[group]; exists {
		_, exists := groupKeys[key]
		return exists
	}
	return false
}

// PrimaryKey returns the primary key that the secondary key in group points to.
// The bool reports whether the secondary key exists.
func (m *MultiKeyMap[K, V]) PrimaryKey(group string, key string) (K, bool) {
	primaryKey, exists := m.secondary[group][key]
	return primaryKey, exists
}

// SecondaryKeys returns an iterator over the secondary keys of primaryKey
// as pairs of group and key.
// The iteration order is unspecified, like for a Go map.
// If primaryKey has no secondary keys, the iterator yields nothing.
func (m *MultiKeyMap[K, V]) SecondaryKeys(primaryKey K) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for group, keys := range m.secondaryTo[primaryKey] {
			for key := range keys {
				if !yield(group, key) {
					return
				}
			}
		}
	}
}

// KeyGroups returns a copy of all key groups,
// mapping group to secondary key to primary key.
// Changes to the result do not affect the map.
func (m *MultiKeyMap[K, V]) KeyGroups() map[string]map[string]K {
	result := make(map[string]map[string]K)
	for group, keys := range m.secondary {
		result[group] = make(map[string]K)
		for key, primary := range keys {
			result[group][key] = primary
		}
	}
	return result
}

// Delete removes primaryKey, its value and all its secondary keys.
// If primaryKey does not exist, Delete is a no-op.
func (m *MultiKeyMap[K, V]) Delete(primaryKey K) {
	delete(m.primary, primaryKey)
	groups, exists := m.secondaryTo[primaryKey]
	if !exists {
		return
	}
	for group, keys := range groups {
		for key := range keys {
			delete(m.secondary[group], key)
		}
		if len(m.secondary[group]) == 0 {
			delete(m.secondary, group)
		}
	}
	delete(m.secondaryTo, primaryKey)
}

// Get returns a value by primary key.
func (m *MultiKeyMap[K, V]) Get(primaryKey K) (V, bool) {
	value, exists := m.primary[primaryKey]
	return value, exists
}

// GetBySecondaryKey returns the value of the primary key that the secondary key
// in group points to. The bool reports whether such a value exists.
func (m *MultiKeyMap[K, V]) GetBySecondaryKey(group string, key string) (V, bool) {
	if groupKeys, exists := m.secondary[group]; exists {
		primaryKey, exists := groupKeys[key]
		if exists {
			value, exists := m.primary[primaryKey]
			return value, exists
		}
	}
	return *new(V), false
}

// Len returns the number of primary keys in the map.
func (m *MultiKeyMap[K, V]) Len() int {
	return len(m.primary)
}

// All returns an iterator over all primary keys and their values.
// The iteration order is unspecified, like for a Go map.
func (m *MultiKeyMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m.primary {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys returns an iterator over all primary keys.
// The iteration order is unspecified, like for a Go map.
func (m *MultiKeyMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.primary {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
func (m *MultiKeyMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m.primary {
			if !yield(v) {
				return
			}
		}
	}
}

// Clear removes all elements from the map.
func (m *MultiKeyMap[K, V]) Clear() {
	*m = MultiKeyMap[K, V]{}
}

// String returns a string representation of the map.
func (m *MultiKeyMap[K, V]) String() string {
	return fmt.Sprintf("MultiKeyMap: %v", m.primary)
}
