// Package multikeymap provides a map with one primary key and any number of
// secondary string keys, grouped by name.
//
// Each value is stored once under its primary key; secondary keys point to the
// primary key, so a lookup by either key costs O(1).
// MultiKeyMap is not safe for concurrent use; ConcurrentMultiKeyMap is.
// The zero value of both types is an empty map ready to use.
package multikeymap

import (
	"fmt"
	"iter"
	"maps"
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

// DeleteSecondaryKey removes the secondary key in group.
// The primary key it points to and its value stay.
// If the secondary key does not exist, DeleteSecondaryKey is a no-op.
func (m *MultiKeyMap[K, V]) DeleteSecondaryKey(group string, key string) {
	primaryKey, exists := m.secondary[group][key]
	if !exists {
		return
	}
	delete(m.secondary[group], key)
	if len(m.secondary[group]) == 0 {
		delete(m.secondary, group)
	}
	m.removeReverseEntry(primaryKey, group, key)
}

// DeleteGroup removes group and all its secondary keys.
// The primary keys and their values stay.
// If group does not exist, DeleteGroup is a no-op.
func (m *MultiKeyMap[K, V]) DeleteGroup(group string) {
	for _, primaryKey := range m.secondary[group] {
		groups := m.secondaryTo[primaryKey]
		delete(groups, group)
		if len(groups) == 0 {
			delete(m.secondaryTo, primaryKey)
		}
	}
	delete(m.secondary, group)
}

// DeleteBySecondaryKey removes the primary key that the secondary key in group points to,
// its value and all its secondary keys, like Delete.
// If the secondary key does not exist, DeleteBySecondaryKey is a no-op.
func (m *MultiKeyMap[K, V]) DeleteBySecondaryKey(group string, key string) {
	if primaryKey, exists := m.secondary[group][key]; exists {
		m.Delete(primaryKey)
	}
}

// LoadOrStore returns the value of primaryKey if it exists.
// Otherwise, it stores value under primaryKey and returns value.
// The loaded result is true if the value was loaded, false if stored.
func (m *MultiKeyMap[K, V]) LoadOrStore(primaryKey K, value V) (actual V, loaded bool) {
	if actual, loaded := m.primary[primaryKey]; loaded {
		return actual, true
	}
	m.Set(primaryKey, value)
	return value, false
}

// LoadAndDelete removes primaryKey, its value and all its secondary keys
// and returns the previous value, if any.
// The loaded result reports whether primaryKey existed.
func (m *MultiKeyMap[K, V]) LoadAndDelete(primaryKey K) (value V, loaded bool) {
	value, loaded = m.primary[primaryKey]
	if loaded {
		m.Delete(primaryKey)
	}
	return value, loaded
}

// LoadAndDeleteBySecondaryKey removes the primary key that the secondary key in group points to,
// its value and all its secondary keys, like DeleteBySecondaryKey,
// and returns the previous value, if any.
// The loaded result reports whether the secondary key existed.
func (m *MultiKeyMap[K, V]) LoadAndDeleteBySecondaryKey(group string, key string) (value V, loaded bool) {
	primaryKey, exists := m.secondary[group][key]
	if !exists {
		return value, false
	}
	return m.LoadAndDelete(primaryKey)
}

// Update calls fn with the current value of primaryKey and whether it exists.
// If fn returns true, Update stores the returned value under primaryKey;
// it creates primaryKey if needed and keeps its existing secondary keys.
// If fn returns false, Update deletes primaryKey and its secondary keys like Delete;
// if primaryKey does not exist, nothing happens.
func (m *MultiKeyMap[K, V]) Update(primaryKey K, fn func(old V, exists bool) (V, bool)) {
	old, exists := m.primary[primaryKey]
	value, keep := fn(old, exists)
	switch {
	case keep:
		m.Set(primaryKey, value)
	case exists:
		m.Delete(primaryKey)
	}
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

// DeleteFunc removes every primary key for which del returns true,
// together with its value and all its secondary keys.
func (m *MultiKeyMap[K, V]) DeleteFunc(del func(K, V) bool) {
	for k, v := range m.primary {
		if del(k, v) {
			m.Delete(k)
		}
	}
}

// Clone returns a copy of the map with its own primary and secondary keys.
// The values are copied with assignment, like maps.Clone does, so this is a shallow clone.
func (m *MultiKeyMap[K, V]) Clone() *MultiKeyMap[K, V] {
	c := &MultiKeyMap[K, V]{primary: maps.Clone(m.primary)}
	if m.secondary == nil {
		return c
	}
	c.secondary = make(map[string]map[string]K, len(m.secondary))
	for group, keys := range m.secondary {
		c.secondary[group] = maps.Clone(keys)
	}
	c.secondaryTo = make(map[K]map[string]map[string]struct{}, len(m.secondaryTo))
	for primaryKey, groups := range m.secondaryTo {
		cg := make(map[string]map[string]struct{}, len(groups))
		for group, keys := range groups {
			cg[group] = maps.Clone(keys)
		}
		c.secondaryTo[primaryKey] = cg
	}
	return c
}

// Clear removes all elements from the map.
// Like clear on a Go map, it keeps the allocated memory,
// and an iteration that is running yields no further entries.
func (m *MultiKeyMap[K, V]) Clear() {
	// clear instead of resetting the struct: a running range keeps the old
	// maps and would go on yielding their entries.
	// The nested maps are cleared first, because an iterator such as
	// SecondaryKeys may already range over one of them.
	for _, groups := range m.secondaryTo {
		for _, keys := range groups {
			clear(keys)
		}
		clear(groups)
	}
	for _, keys := range m.secondary {
		clear(keys)
	}
	clear(m.primary)
	clear(m.secondary)
	clear(m.secondaryTo)
}

// String returns a string representation of the map.
func (m *MultiKeyMap[K, V]) String() string {
	return fmt.Sprintf("MultiKeyMap: %v", m.primary)
}
