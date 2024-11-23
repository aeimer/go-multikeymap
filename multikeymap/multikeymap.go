package multikeymap

import (
	"fmt"
	"sync"
)

// MultiKeyMap is a generic in-memory map with a primary key and multiple secondary keys.
// It implements container/Container.
type MultiKeyMap[K comparable, V any] struct {
	mu          sync.RWMutex
	primary     map[K]V
	secondary   map[string]map[string]K // Group -> SecondaryKey -> PrimaryKey
	secondaryTo map[K]map[string]string // PrimaryKey -> Group -> SecondaryKey
}

// New creates a new MultiKeyMap instance.
func New[K comparable, V any]() *MultiKeyMap[K, V] {
	return &MultiKeyMap[K, V]{
		primary:     make(map[K]V),
		secondary:   make(map[string]map[string]K),
		secondaryTo: make(map[K]map[string]string),
	}
}

// Put inserts a value with a primary key.
func (m *MultiKeyMap[K, V]) Put(primaryKey K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.primary[primaryKey] = value
}

// PutSecondaryKeys adds secondary keys under a group for a primary key.
func (m *MultiKeyMap[K, V]) PutSecondaryKeys(primaryKey K, group string, keys ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.secondary[group] == nil {
		m.secondary[group] = make(map[string]K)
	}
	if m.secondaryTo[primaryKey] == nil {
		m.secondaryTo[primaryKey] = make(map[string]string)
	}
	for _, key := range keys {
		m.secondary[group][key] = primaryKey
		m.secondaryTo[primaryKey][group] = key
	}
}

// HasPrimaryKey checks if a primary key exists.
func (m *MultiKeyMap[K, V]) HasPrimaryKey(primaryKey K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.primary[primaryKey]
	return exists
}

// HasSecondaryKey checks if a secondary key exists in a specific group.
func (m *MultiKeyMap[K, V]) HasSecondaryKey(group string, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if groupKeys, exists := m.secondary[group]; exists {
		_, exists := groupKeys[key]
		return exists
	}
	return false
}

// GetAllKeyGroups returns all key groups and their secondary keys.
func (m *MultiKeyMap[K, V]) GetAllKeyGroups() map[string]map[string]K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Create a copy of the key groups to avoid concurrency issues
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
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.primary, primaryKey)
	if groups, exists := m.secondaryTo[primaryKey]; exists {
		for group, key := range groups {
			delete(m.secondary[group], key)
			if len(m.secondary[group]) == 0 {
				delete(m.secondary, group)
			}
		}
		delete(m.secondaryTo, primaryKey)
	}
}

// Get returns a value by primary key.
func (m *MultiKeyMap[K, V]) Get(primaryKey K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.primary[primaryKey]
	return value, exists
}

// GetBySecondaryKey returns a primary key by secondary key and group.
func (m *MultiKeyMap[K, V]) GetBySecondaryKey(group string, key string) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.primary)
}

// Empty checks if the map is empty.
func (m *MultiKeyMap[K, V]) Empty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.primary) == 0
}

// Values returns a slice of all values in the map.
func (m *MultiKeyMap[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.primary))
	for _, value := range m.primary {
		values = append(values, value)
	}
	return values
}

// Clear removes all elements from the map.
func (m *MultiKeyMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.primary = make(map[K]V)
	m.secondary = make(map[string]map[string]K)
	m.secondaryTo = make(map[K]map[string]string)
}

// String returns a string representation of the map.
func (m *MultiKeyMap[K, V]) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return fmt.Sprintf("MultiKeyMap: %v", m.primary)
}
