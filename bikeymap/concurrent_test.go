package bikeymap

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleNewConcurrent() {
	bm := NewConcurrent[string, int, string]()
	_ = bm.Set("keyA1", 1, "value1")
	value, exists := bm.GetByKeyA("keyA1")
	fmt.Printf("[Key A] value: %v, exists: %v\n", value, exists)
	value, exists = bm.GetByKeyB(1)
	fmt.Printf("[Key B] value: %v, exists: %v\n", value, exists)
	// Output:
	// [Key A] value: value1, exists: true
	// [Key B] value: value1, exists: true
}

func TestConcurrentBiKeyMap_SetAndGet(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	value, exists := bm.GetByKeyA("keyA1")
	if !exists || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}

	value, exists = bm.GetByKeyB(1)
	if !exists || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}
}

func TestConcurrentBiKeyMap_SetDuplicateKeys(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	// keyB 1 is already paired with keyA1.
	err = bm.Set("keyA2", 1, "value2")
	require.ErrorIs(t, err, ErrKeyBConflict)
	require.EqualError(t, err, "bikeymap: keyB is paired with a different keyA")

	// keyA1 is already paired with keyB 1.
	err = bm.Set("keyA1", 2, "value2")
	require.ErrorIs(t, err, ErrKeyAConflict)
	require.EqualError(t, err, "bikeymap: keyA is paired with a different keyB")

	// Both keys conflict: the keyB conflict is reported.
	require.NoError(t, bm.Set("keyA3", 3, "value3"))
	err = bm.Set("keyA1", 3, "value4")
	require.ErrorIs(t, err, ErrKeyBConflict)

	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA3": "value3"}, maps.Collect(bm.All()))
}

func TestConcurrentBiKeyMap_DeleteByKeyA(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	bm.DeleteByKeyA("keyA1")

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be removed")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be removed")
	}
}

func TestConcurrentBiKeyMap_DeleteByKeyB(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)
	bm.DeleteByKeyB(1)

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be removed")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be removed")
	}
}

func TestConcurrentBiKeyMap_String(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	_ = bm.Set("keyA1", 1, "value1")
	expected := "ConcurrentBiKeyMap: map[keyA1:value1]"
	if bm.String() != expected {
		t.Errorf("expected %s, got %s", expected, bm.String())
	}
}

func TestConcurrentBiKeyMap_DeleteByKeyA_NotFound(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	bm.DeleteByKeyA("nonExistentKey")
	assert.Equal(t, 1, bm.Len())
}

func TestConcurrentBiKeyMap_DeleteByKeyB_NotFound(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	bm.DeleteByKeyB(999)
	assert.Equal(t, 1, bm.Len())
}
func TestConcurrentBiKeyMap_Len(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	assert.Equal(t, 0, bm.Len())
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	assert.Equal(t, 1, bm.Len())
}

func TestConcurrentBiKeyMap_Clear(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	bm.Clear()
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_Values(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.ElementsMatch(t, []string{"value1", "value2"}, slices.Collect(bm.Values()))
}

func TestConcurrentBiKeyMap_ValuesStopsEarly(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	count := 0
	for range bm.Values() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestConcurrentBiKeyMap_ValuesAllowsWritesInLoop(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	keyB := 100
	for v := range bm.Values() {
		bm.DeleteByKeyA("keyA" + strings.TrimPrefix(v, "value"))
		keyB++
		require.NoError(t, bm.Set("new"+v, keyB, v))
	}
	assert.ElementsMatch(t, []string{"value1", "value2"}, slices.Collect(bm.Values()))
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
}

func TestConcurrentBiKeyMap_All(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2"}, maps.Collect(bm.All()))
}

func TestConcurrentBiKeyMap_Keys(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.ElementsMatch(t, []string{"keyA1", "keyA2"}, slices.Collect(bm.KeysA()))
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(bm.KeysB()))
}

func TestConcurrentBiKeyMap_IteratorsStopEarly(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	count := 0
	for range bm.All() {
		count++
		break
	}
	for range bm.KeysA() {
		count++
		break
	}
	for range bm.KeysB() {
		count++
		break
	}
	assert.Equal(t, 3, count)
}

func TestConcurrentBiKeyMap_IteratorsAllowWritesInLoop(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	var errs []error
	requireNoDeadlock(t, func() {
		for keyA, v := range bm.All() {
			bm.DeleteByKeyA(keyA)
			errs = append(errs, bm.Set("all", 2, v))
		}
		for keyA := range bm.KeysA() {
			bm.DeleteByKeyA(keyA)
			errs = append(errs, bm.Set("keysA", 3, "value3"))
		}
		for keyB := range bm.KeysB() {
			bm.DeleteByKeyB(keyB)
			errs = append(errs, bm.Set("keysB", 4, "value4"))
		}
		for v := range bm.Values() {
			errs = append(errs, bm.Set("keysB", 4, v+"!"))
		}
	})
	require.NoError(t, errors.Join(errs...))
	assert.Equal(t, map[string]string{"keysB": "value4!"}, maps.Collect(bm.All()))
}

// requireNoDeadlock fails the test if f does not return within a second.
func requireNoDeadlock(t *testing.T, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deadlock: loop body could not write to the map")
	}
}

func TestConcurrentBiKeyMap_KeyAAndKeyB(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	keyA, ok := bm.KeyA(1)
	assert.True(t, ok)
	assert.Equal(t, "keyA1", keyA)
	keyB, ok := bm.KeyB("keyA1")
	assert.True(t, ok)
	assert.Equal(t, 1, keyB)

	keyA, ok = bm.KeyA(2)
	assert.False(t, ok)
	assert.Empty(t, keyA)
	keyB, ok = bm.KeyB("missing")
	assert.False(t, ok)
	assert.Zero(t, keyB)
}

func TestConcurrentBiKeyMap_ConcurrentAccess(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrently set values
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentSet(t, i, &wg, bm)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently get values
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentGet(t, i, &wg, bm)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently remove values
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentDelete(t, i, &wg, bm)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func TestConcurrentBiKeyMap_LenDuringSet(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := range numGoroutines {
		wg.Go(func() {
			assert.NoError(t, bm.Set(strconv.Itoa(i), i, "value"))
		})
		wg.Go(func() {
			_ = bm.Len()
		})
	}
	wg.Wait()

	assert.Equal(t, numGoroutines, bm.Len())
}

func concurrentSet(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	value := fmt.Sprintf("value%d", i)
	if err := bm.Set(keyA, keyB, value); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func concurrentGet(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	if value, exists := bm.GetByKeyA(keyA); !exists || value != fmt.Sprintf("value%d", i) {
		t.Errorf("expected value%d, got %v", i, value)
	}
	if value, exists := bm.GetByKeyB(keyB); !exists || value != fmt.Sprintf("value%d", i) {
		t.Errorf("expected value%d, got %v", i, value)
	}
}

func concurrentDelete(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	bm.DeleteByKeyA(keyA)
	if _, exists := bm.GetByKeyA(keyA); exists {
		t.Errorf("expected keyA%d to be removed", i)
	}
	if _, exists := bm.GetByKeyB(keyB); exists {
		t.Errorf("expected keyB%d to be removed", i)
	}
}

func TestConcurrentBiKeyMap_ZeroValueReads(t *testing.T) {
	var bm ConcurrentBiKeyMap[string, int, string]
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
	_, ok = bm.GetByKeyB(1)
	assert.False(t, ok)
	assert.Equal(t, 0, bm.Len())
	assert.Empty(t, slices.Collect(bm.Values()))
	assert.Equal(t, "ConcurrentBiKeyMap: map[]", bm.String())
	bm.DeleteByKeyA("keyA1")
	bm.DeleteByKeyB(1)
	bm.Clear()
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_ZeroValueWrites(t *testing.T) {
	var bm ConcurrentBiKeyMap[string, int, string]
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	value, ok := bm.GetByKeyB(1)
	assert.True(t, ok)
	assert.Equal(t, "value1", value)
	bm.Clear()
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.Equal(t, 1, bm.Len())
}

// Benchmarks

var benchmarkConcurrentSizes = []struct {
	size int
}{
	{size: 100},
	{size: 1000},
	{size: 10_000},
	{size: 100_000},
}

func BenchmarkConcurrentBiKeyMapGet(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			for n := range v.size {
				_ = m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
			}
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.GetByKeyA(strconv.Itoa(n))
					m.GetByKeyB(n)
				}
			}
		})
	}
}

func BenchmarkConcurrentBiKeyMapSet(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					_ = m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkConcurrentBiKeyMapDelete(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			for n := range v.size {
				_ = m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
			}
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.DeleteByKeyB(n)
				}
			}
		})
	}
}
