package bikeymap

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"sync"
)

// ConcurrentBiKeyMap is the same as BiKeyMap, but it is safe for concurrent use.
// It uses a RWMutex to protect the map from concurrent reads and writes.
// Therefore, it is slower than BiKeyMap, but it is safe for concurrent use.
// The zero value is an empty map ready to use.
// A ConcurrentBiKeyMap must not be copied after first use.
type ConcurrentBiKeyMap[A comparable, B comparable, V any] struct {
	mu sync.RWMutex
	m  BiKeyMap[A, B, V]
}

// NewConcurrent returns a new empty ConcurrentBiKeyMap.
// It is equivalent to new(ConcurrentBiKeyMap[A, B, V]).
func NewConcurrent[A comparable, B comparable, V any]() *ConcurrentBiKeyMap[A, B, V] {
	return new(ConcurrentBiKeyMap[A, B, V])
}

// Set stores value under keyA and keyB and pairs the two keys.
// If both keys are already paired with each other, Set replaces the value.
// If keyA is paired with a different keyB, Set removes that old keyB;
// keyA keeps its entry, now with keyB and value.
// If keyB is paired with a different keyA, Set removes that old keyA and its value.
// No key is ever left without its pair.
// Len grows by one if both keys are new, shrinks by one if both cases above apply,
// and stays the same otherwise.
// The removals and the store happen atomically.
func (c *ConcurrentBiKeyMap[A, B, V]) Set(keyA A, keyB B, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.Set(keyA, keyB, value)
}

// GetByKeyA retrieves a value using the first key.
func (c *ConcurrentBiKeyMap[A, B, V]) GetByKeyA(keyA A) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetByKeyA(keyA)
}

// GetByKeyB retrieves a value using the second key.
func (c *ConcurrentBiKeyMap[A, B, V]) GetByKeyB(keyB B) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.GetByKeyB(keyB)
}

// KeyA returns the keyA paired with keyB.
// The bool reports whether keyB exists.
func (c *ConcurrentBiKeyMap[A, B, V]) KeyA(keyB B) (A, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.KeyA(keyB)
}

// KeyB returns the keyB paired with keyA.
// The bool reports whether keyA exists.
func (c *ConcurrentBiKeyMap[A, B, V]) KeyB(keyA A) (B, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.KeyB(keyA)
}

// DeleteByKeyA removes the value stored under keyA and the paired keyB.
// If keyA does not exist, DeleteByKeyA is a no-op.
func (c *ConcurrentBiKeyMap[A, B, V]) DeleteByKeyA(keyA A) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteByKeyA(keyA)
}

// DeleteByKeyB removes the value stored under keyB and the paired keyA.
// If keyB does not exist, DeleteByKeyB is a no-op.
func (c *ConcurrentBiKeyMap[A, B, V]) DeleteByKeyB(keyB B) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteByKeyB(keyB)
}

// LoadAndDeleteByKeyA removes the value stored under keyA and the paired keyB
// and returns the previous value, if any.
// The bool reports whether keyA existed.
// The lookup and the delete happen atomically.
func (c *ConcurrentBiKeyMap[A, B, V]) LoadAndDeleteByKeyA(keyA A) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.LoadAndDeleteByKeyA(keyA)
}

// LoadAndDeleteByKeyB removes the value stored under keyB and the paired keyA
// and returns the previous value, if any.
// The bool reports whether keyB existed.
// The lookup and the delete happen atomically.
func (c *ConcurrentBiKeyMap[A, B, V]) LoadAndDeleteByKeyB(keyB B) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m.LoadAndDeleteByKeyB(keyB)
}

// Len returns the number of values in the map.
func (c *ConcurrentBiKeyMap[A, B, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m.Len()
}

// All returns an iterator over all keyA keys and their values.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[A, B, V]) All() iter.Seq2[A, V] {
	return snapshot2(&c.mu, c.m.All)
}

// Pairs returns an iterator over all pairs of keyA and keyB.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[A, B, V]) Pairs() iter.Seq2[A, B] {
	return snapshot2(&c.mu, c.m.Pairs)
}

// KeysA returns an iterator over all keyA keys.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[A, B, V]) KeysA() iter.Seq[A] {
	return snapshot(&c.mu, c.m.KeysA)
}

// KeysB returns an iterator over all keyB keys.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[A, B, V]) KeysB() iter.Seq[B] {
	return snapshot(&c.mu, c.m.KeysB)
}

// Values returns an iterator over all values in the map.
// The iteration order is unspecified, like for a Go map.
// Each iteration sees a snapshot taken when it starts and holds no lock while it runs,
// so the loop body may modify the map.
func (c *ConcurrentBiKeyMap[A, B, V]) Values() iter.Seq[V] {
	return snapshot(&c.mu, c.m.Values)
}

// DeleteFunc removes every value for which del returns true, together with its keyA and keyB.
// del runs while DeleteFunc holds the write lock,
// so del must not call methods of the same map, or it deadlocks.
func (c *ConcurrentBiKeyMap[A, B, V]) DeleteFunc(del func(A, B, V) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m.DeleteFunc(del)
}

// Clone returns a copy of the map with its own keys and its own lock.
// The values are copied with assignment, like maps.Clone does, so this is a shallow clone.
func (c *ConcurrentBiKeyMap[A, B, V]) Clone() *ConcurrentBiKeyMap[A, B, V] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &ConcurrentBiKeyMap[A, B, V]{m: *c.m.Clone()}
}

// Clear removes all elements from the map and releases their memory.
// Iterations that are running are not affected, because they work on a snapshot.
func (c *ConcurrentBiKeyMap[A, B, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Reset instead of clear: the iterators work on snapshots, so clearing
	// in place would only cost time under the lock and keep the memory.
	c.m = BiKeyMap[A, B, V]{}
}

// String returns a string representation of the map.
func (c *ConcurrentBiKeyMap[A, B, V]) String() string {
	// All copies the entries under the read lock and yields them without it,
	// so a value's String method may call back into the map without a deadlock.
	// Collecting into a map lets fmt sort the keys like the plain String does.
	return fmt.Sprint("ConcurrentBiKeyMap: ", maps.Collect(c.All()))
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
