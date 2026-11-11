package multikeymap

import (
	"fmt"
)

// MultiKeyMap is a generic in-memory map with a primary key and multiple secondary keys.
// Secondary keys are strings, organised in named groups.
// A MultiKeyMap is not safe for concurrent use; use ConcurrentMultiKeyMap instead.
type MultiKeyMap[K comparable, V any] struct {
	primary     map[K]V
	secondary   map[string]map[string]K              // Group -> SecondaryKey -> PrimaryKey
	secondaryTo map[K]map[string]map[string]struct{} // PrimaryKey -> Group -> set of SecondaryKeys
}

// New creates a new MultiKeyMap instance.
func New[K comparable, V any]() *MultiKeyMap[K, V] {
	return &MultiKeyMap[K, V]{
		primary:     make(map[K]V),
		secondary:   make(map[string]map[string]K),
		secondaryTo: make(map[K]map[string]map[string]struct{}),
	}
}

// Put inserts a value with a primary key.
func (m *MultiKeyMap[K, V]) Put(primaryKey K, value V) {
	m.primary[primaryKey] = value
}

// PutSecondaryKeys adds secondary keys under a group for a primary key.
// It reports false and changes nothing if the primary key does not exist.
// A secondary key that already belongs to another primary key in the same group
// moves to primaryKey; the other primary key loses it.
func (m *MultiKeyMap[K, V]) PutSecondaryKeys(primaryKey K, group string, keys ...string) bool {
	if _, exists := m.primary[primaryKey]; !exists {
		return false
	}
	if len(keys) == 0 {
		return true
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

// HasPrimaryKey checks if a primary key exists.
func (m *MultiKeyMap[K, V]) HasPrimaryKey(primaryKey K) bool {
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

// GetAllKeyGroups returns a copy of all key groups,
// mapping group to secondary key to primary key.
// Changes to the result do not affect the map.
func (m *MultiKeyMap[K, V]) GetAllKeyGroups() map[string]map[string]K {
	result := make(map[string]map[string]K)
	for group, keys := range m.secondary {
		result[group] = make(map[string]K)
		for key, primary := range keys {
			result[group][key] = primary
		}
	}
	return result
}

// Remove removes a primary key and its associated secondary keys.
func (m *MultiKeyMap[K, V]) Remove(primaryKey K) {
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

// Size returns the number of primary keys in the map.
func (m *MultiKeyMap[K, V]) Size() int {
	return len(m.primary)
}

// Empty checks if the map is empty.
func (m *MultiKeyMap[K, V]) Empty() bool {
	return len(m.primary) == 0
}

// Values returns a slice of all values in the map.
func (m *MultiKeyMap[K, V]) Values() []V {
	values := make([]V, 0, len(m.primary))
	for _, value := range m.primary {
		values = append(values, value)
	}
	return values
}

// Clear removes all elements from the map.
func (m *MultiKeyMap[K, V]) Clear() {
	m.primary = make(map[K]V)
	m.secondary = make(map[string]map[string]K)
	m.secondaryTo = make(map[K]map[string]map[string]struct{})
}

// String returns a string representation of the map.
func (m *MultiKeyMap[K, V]) String() string {
	return fmt.Sprintf("MultiKeyMap: %v", m.primary)
}
