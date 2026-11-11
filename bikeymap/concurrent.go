package bikeymap

import (
	"iter"
	"slices"
	"sync"
)

// ConcurrentBiKeyMap is the same as BiKeyMap, but it is safe for concurrent use.
// It uses a RWMutex to protect the map from concurrent reads and writes.
// Therefore, it is slower than BiKeyMap, but it is safe for concurrent use.
// The zero value is an empty map ready to use.
// A ConcurrentBiKeyMap must not be copied after first use.
type ConcurrentBiKeyMap[KeyA comparable, KeyB comparable, V any] struct {
	mu sync.RWMutex
	m  BiKeyMap[KeyA, KeyB, V]
}

// NewConcurrent returns a new empty ConcurrentBiKeyMap.
// It is equivalent to new(ConcurrentBiKeyMap[KeyA, KeyB, V]).
func NewConcurrent[KeyA comparable, KeyB comparable, V any]() *ConcurrentBiKeyMap[KeyA, KeyB, V] {
	return new(ConcurrentBiKeyMap[KeyA, KeyB, V])
}

// Set stores value under keyA and keyB and pairs the two keys.
// If keyB is already paired with a different keyA, Set returns ErrKeyBConflict.
// Otherwise, if keyA is already paired with a different keyB, Set returns ErrKeyAConflict.
// On an error Set changes nothing.
// If both keys are already paired with each other, Set replaces the value.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Set(keyA KeyA, keyB KeyB, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.Set(keyA, keyB, value)
}

// GetByKeyA retrieves a value using the first key.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) GetByKeyA(keyA KeyA) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetByKeyA(keyA)
}

// GetByKeyB retrieves a value using the second key.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) GetByKeyB(keyB KeyB) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetByKeyB(keyB)
}

// KeyA returns the keyA paired with keyB.
// The bool reports whether keyB exists.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) KeyA(keyB KeyB) (KeyA, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.KeyA(keyB)
}

// KeyB returns the keyB paired with keyA.
// The bool reports whether keyA exists.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) KeyB(keyA KeyA) (KeyB, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.KeyB(keyA)
}

// DeleteByKeyA removes the value stored under keyA and the paired keyB.
// If keyA does not exist, DeleteByKeyA is a no-op.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) DeleteByKeyA(keyA KeyA) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteByKeyA(keyA)
}

// DeleteByKeyB removes the value stored under keyB and the paired keyA.
// If keyB does not exist, DeleteByKeyB is a no-op.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) DeleteByKeyB(keyB KeyB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteByKeyB(keyB)
}

// LoadAndDeleteByKeyA removes the value stored under keyA and the paired keyB
// and returns the previous value, if any.
// The bool reports whether keyA existed.
// The lookup and the delete happen atomically.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) LoadAndDeleteByKeyA(keyA KeyA) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.LoadAndDeleteByKeyA(keyA)
}

// LoadAndDeleteByKeyB removes the value stored under keyB and the paired keyA
// and returns the previous value, if any.
// The bool reports whether keyB existed.
// The lookup and the delete happen atomically.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) LoadAndDeleteByKeyB(keyB KeyB) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.LoadAndDeleteByKeyB(keyB)
}

// Len returns the number of values in the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Len()
}

// All returns an iterator over all keyA keys and their values.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) All() iter.Seq2[KeyA, V] {
	return snapshot2(&c.mu, c.m.All)
}

// KeysA returns an iterator over all keyA keys.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) KeysA() iter.Seq[KeyA] {
	return snapshot(&c.mu, c.m.KeysA)
}

// KeysB returns an iterator over all keyB keys.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) KeysB() iter.Seq[KeyB] {
	return snapshot(&c.mu, c.m.KeysB)
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Values() iter.Seq[V] {
	return snapshot(&c.mu, c.m.Values)
}

// DeleteFunc removes every value for which del returns true, together with its keyA and keyB.
// del runs while DeleteFunc holds the write lock,
// so del must not call methods of the same map, or it deadlocks.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) DeleteFunc(del func(KeyA, KeyB, V) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteFunc(del)
}

// Clone returns a copy of the map with its own keys and its own lock.
// The values are copied with assignment, like maps.Clone does, so this is a shallow clone.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Clone() *ConcurrentBiKeyMap[KeyA, KeyB, V] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &ConcurrentBiKeyMap[KeyA, KeyB, V]{m: *c.m.Clone()}
}

// Clear removes all elements from the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Clear()
}

// String returns a string representation of the map.
func (c *ConcurrentBiKeyMap[KeyA, KeyB, V]) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return "ConcurrentBiKeyMap: " + c.m.format()
}

// snapshot returns an iterator that, when iteration starts, collects seq() under the read lock
// and then yields the copy without holding the lock.
func snapshot[T any](mu *sync.RWMutex, seq func() iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		mu.RLock()
		items := slices.Collect(seq())
		mu.RUnlock()
		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// snapshot2 is snapshot for iter.Seq2.
func snapshot2[K, V any](mu *sync.RWMutex, seq func() iter.Seq2[K, V]) iter.Seq2[K, V] {
	type pair struct {
		k K
		v V
	}
	return func(yield func(K, V) bool) {
		mu.RLock()
		var items []pair
		for k, v := range seq() {
			items = append(items, pair{k, v})
		}
		mu.RUnlock()
		for _, p := range items {
			if !yield(p.k, p.v) {
				return
			}
		}
	}
}
