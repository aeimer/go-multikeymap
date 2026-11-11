package multikeymap

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleNewConcurrent() {
	mm := NewConcurrent[string, int]()
	mm.Set("keyA1", 1)
	mm.AddSecondaryKeys("keyA1", "group1", "key1", "key2")
	mm.AddSecondaryKeys("keyA1", "group2", "key3", "key4")
	value, exists := mm.Get("keyA1")
	fmt.Printf("[Key A1] value: %v, exists: %v\n", value, exists)
	value, exists = mm.GetBySecondaryKey("group1", "key2")
	fmt.Printf("[Secondary group1 key2] value: %v, exists: %v\n", value, exists)

	// Output:
	// [Key A1] value: 1, exists: true
	// [Secondary group1 key2] value: 1, exists: true
}

func TestConcurrentMultiKeyMap_SetAndGet(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	value, exists := mm.Get("key1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestConcurrentMultiKeyMap_SetSecondaryKeys(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	value, exists := mm.GetBySecondaryKey("group1", "secKey1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestConcurrentMultiKeyMap_AddSecondaryKeysReportsMissingPrimaryKey(t *testing.T) {
	mm := NewConcurrent[string, int]()
	assert.False(t, mm.AddSecondaryKeys("missing", "group1", "secKey1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())

	mm.Set("key1", 1)
	assert.True(t, mm.AddSecondaryKeys("key1", "group1", "secKey1"))
	assert.True(t, mm.HasSecondaryKey("group1", "secKey1"))
}

func TestConcurrentMultiKeyMap_Has(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	if !mm.Has("key1") {
		t.Error("expected primary key 'key1' to exist")
	}
}

func TestConcurrentMultiKeyMap_HasSecondaryKey(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	if !mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to exist")
	}
}

func TestConcurrentMultiKeyMap_Delete(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.Delete("key1")
	if mm.Has("key1") {
		t.Error("expected primary key 'key1' to be removed")
	}
	if mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to be removed")
	}
}

func TestConcurrentMultiKeyMap_KeyGroups(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key2", "group2", "secKey3")
	allGroups := mm.KeyGroups()
	if len(allGroups) != 2 || len(allGroups["group1"]) != 2 || len(allGroups["group2"]) != 1 {
		t.Errorf("expected allGroups to contain 'group1' and 'group2' with correct keys, got %v", allGroups)
	}
}

func TestConcurrentMultiKeyMap_GetBySecondaryKey_NotFound(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	if _, exists := mm.GetBySecondaryKey("group1", "nonExistentKey"); exists {
		t.Error("expected 'nonExistentKey' to not be found")
	}
}

func TestConcurrentMultiKeyMap_String(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	expected := "ConcurrentMultiKeyMap: map[key1:1]"
	if mm.String() != expected {
		t.Errorf("expected %s, got %s", expected, mm.String())
	}
}

func TestConcurrentMultiKeyMap_Len(t *testing.T) {
	mm := NewConcurrent[string, int]()
	assert.Equal(t, 0, mm.Len())
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.Equal(t, 2, mm.Len())
}

func TestConcurrentMultiKeyMap_Values(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(mm.Values()))
}

func TestConcurrentMultiKeyMap_ValuesStopsEarly(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	count := 0
	for range mm.Values() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestConcurrentMultiKeyMap_ValuesAllowsWritesInLoop(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	for v := range mm.Values() {
		mm.Set("key"+strconv.Itoa(v+10), v)
		mm.Delete("key" + strconv.Itoa(v))
	}
	assert.Equal(t, 2, mm.Len())
	assert.False(t, mm.Has("key1"))
	assert.True(t, mm.Has("key11"))
}

func TestConcurrentMultiKeyMap_Clear(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Clear()
	if mm.Len() != 0 {
		t.Error("expected map to be empty after clear")
	}
}

func TestConcurrentMultiKeyMap_All(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.Equal(t, map[string]int{"key1": 1, "key2": 2}, maps.Collect(mm.All()))
}

func TestConcurrentMultiKeyMap_Keys(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.ElementsMatch(t, []string{"key1", "key2"}, slices.Collect(mm.Keys()))
}

func TestConcurrentMultiKeyMap_IteratorsStopEarly(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	count := 0
	for range mm.All() {
		count++
		break
	}
	for range mm.Keys() {
		count++
		break
	}
	assert.Equal(t, 2, count)
}

func TestConcurrentMultiKeyMap_IteratorsAllowWritesInLoop(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	requireNoDeadlock(t, func() {
		for k, v := range mm.All() {
			mm.Delete(k)
			mm.Set("all"+k, v)
		}
		for k := range mm.Keys() {
			mm.Delete(k)
			mm.Set("keys"+k, 0)
		}
		for v := range mm.Values() {
			mm.Set("values"+strconv.Itoa(v), v)
		}
	})
	assert.ElementsMatch(t, []string{"keysallkey1", "keysallkey2", "values0"}, slices.Collect(mm.Keys()))
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

func TestConcurrentMultiKeyMap_PrimaryKey(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")

	pk, ok := mm.PrimaryKey("group1", "secKey1")
	assert.True(t, ok)
	assert.Equal(t, "key1", pk)

	pk, ok = mm.PrimaryKey("group1", "missing")
	assert.False(t, ok)
	assert.Empty(t, pk)
	_, ok = mm.PrimaryKey("missing", "secKey1")
	assert.False(t, ok)
}

func TestConcurrentMultiKeyMap_SecondaryKeys(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey3")
	mm.AddSecondaryKeys("key2", "group1", "secKey4")

	type groupKey struct{ group, key string }
	var got []groupKey
	for group, key := range mm.SecondaryKeys("key1") {
		got = append(got, groupKey{group, key})
	}
	assert.ElementsMatch(t, []groupKey{{"group1", "secKey1"}, {"group1", "secKey2"}, {"group2", "secKey3"}}, got)
	for group, key := range mm.SecondaryKeys("missing") {
		t.Errorf("unexpected secondary key %s/%s", group, key)
	}
}

func TestConcurrentMultiKeyMap_SecondaryKeysStopsEarly(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey3")
	count := 0
	for range mm.SecondaryKeys("key1") {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestConcurrentMultiKeyMap_SecondaryKeysAllowsWritesInLoop(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	requireNoDeadlock(t, func() {
		for group, key := range mm.SecondaryKeys("key1") {
			mm.AddSecondaryKeys("key1", group, "new"+key)
		}
	})
	assert.Len(t, mm.KeyGroups()["group1"], 4)
}

func TestConcurrentMultiKeyMap_ConcurrentAccess(t *testing.T) {
	mm := NewConcurrent[string, int]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrently add values
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentAdd(t, i, &wg, mm)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently get values
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentGet(t, i, &wg, mm)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently remove values
	for i := range numGoroutines {
		wg.Add(1)
		go func(i int) {
			concurrentDelete(t, i, &wg, mm)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func concurrentAdd(t *testing.T, i int, wg *sync.WaitGroup, mm *ConcurrentMultiKeyMap[string, int]) {
	t.Helper()
	defer wg.Done()
	key := fmt.Sprintf("key%d", i)
	mm.Set(key, i)
}

func concurrentGet(t *testing.T, i int, wg *sync.WaitGroup, mm *ConcurrentMultiKeyMap[string, int]) {
	t.Helper()
	defer wg.Done()
	key := fmt.Sprintf("key%d", i)
	if value, exists := mm.Get(key); !exists || value != i {
		t.Errorf("expected %d, got %v", i, value)
	}
}

func concurrentDelete(t *testing.T, i int, wg *sync.WaitGroup, mm *ConcurrentMultiKeyMap[string, int]) {
	t.Helper()
	defer wg.Done()
	key := fmt.Sprintf("key%d", i)
	mm.Delete(key)
	if _, exists := mm.Get(key); exists {
		t.Errorf("expected key%d to be removed", i)
	}
}

func TestConcurrentMultiKeyMap_ZeroValueReads(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	_, ok := mm.Get("key1")
	assert.False(t, ok)
	_, ok = mm.GetBySecondaryKey("group1", "secKey1")
	assert.False(t, ok)
	assert.False(t, mm.Has("key1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())
	assert.Equal(t, 0, mm.Len())
	assert.Empty(t, slices.Collect(mm.Values()))
	assert.Equal(t, "ConcurrentMultiKeyMap: map[]", mm.String())
	assert.False(t, mm.AddSecondaryKeys("key1", "group1", "secKey1"))
	mm.Delete("key1")
	mm.Clear()
	assert.Equal(t, 0, mm.Len())
}

func TestConcurrentMultiKeyMap_ZeroValueWrites(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	mm.Set("key1", 1)
	assert.True(t, mm.AddSecondaryKeys("key1", "group1", "secKey1"))
	value, ok := mm.GetBySecondaryKey("group1", "secKey1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
	mm.Clear()
	mm.Set("key2", 2)
	assert.True(t, mm.AddSecondaryKeys("key2", "group1", "secKey2"))
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey2": "key2"}}, mm.KeyGroups())
}

func TestConcurrentMultiKeyMap_LoadOrStore(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	actual, loaded := mm.LoadOrStore("key1", 1)
	assert.False(t, loaded)
	assert.Equal(t, 1, actual)

	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	actual, loaded = mm.LoadOrStore("key1", 2)
	assert.True(t, loaded)
	assert.Equal(t, 1, actual)
	value, _ := mm.GetBySecondaryKey("group1", "secKey1")
	assert.Equal(t, 1, value)
}

func TestConcurrentMultiKeyMap_LoadAndDelete(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	_, loaded := mm.LoadAndDelete("key1")
	assert.False(t, loaded)

	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	value, loaded := mm.LoadAndDelete("key1")
	assert.True(t, loaded)
	assert.Equal(t, 1, value)
	assert.False(t, mm.Has("key1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())

	_, loaded = mm.LoadAndDelete("key1")
	assert.False(t, loaded)
}

func TestConcurrentMultiKeyMap_UpdateCreatesMissingKey(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	mm.Update("key1", func(old int, exists bool) (int, bool) {
		assert.False(t, exists)
		assert.Equal(t, 0, old)
		return 1, true
	})
	value, ok := mm.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
}

func TestConcurrentMultiKeyMap_UpdateReplacesValueAndKeepsSecondaryKeys(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.Update("key1", func(old int, exists bool) (int, bool) {
		assert.True(t, exists)
		return old + 1, true
	})
	value, ok := mm.GetBySecondaryKey("group1", "secKey1")
	assert.True(t, ok)
	assert.Equal(t, 2, value)
}

func TestConcurrentMultiKeyMap_UpdateDeletesExistingKey(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.Update("key1", func(int, bool) (int, bool) { return 0, false })
	assert.False(t, mm.Has("key1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())
}

func TestConcurrentMultiKeyMap_UpdateWithoutKeepOnMissingKeyIsNoOp(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	mm.Update("key1", func(int, bool) (int, bool) { return 1, false })
	assert.False(t, mm.Has("key1"))
	assert.Equal(t, 0, mm.Len())
}

func TestConcurrentMultiKeyMap_LoadOrStoreStoresOnce(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	var wg sync.WaitGroup
	var stored atomic.Int32
	for i := range 100 {
		wg.Go(func() {
			if _, loaded := mm.LoadOrStore("key1", i); !loaded {
				stored.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, int32(1), stored.Load())
}

func TestConcurrentMultiKeyMap_LoadAndDeleteLoadsOnce(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	mm.Set("key1", 1)
	var wg sync.WaitGroup
	var loadedCount atomic.Int32
	for range 100 {
		wg.Go(func() {
			if _, loaded := mm.LoadAndDelete("key1"); loaded {
				loadedCount.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, int32(1), loadedCount.Load())
}

func TestConcurrentMultiKeyMap_UpdateIsAtomic(t *testing.T) {
	var mm ConcurrentMultiKeyMap[string, int]
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			mm.Update("counter", func(old int, _ bool) (int, bool) { return old + 1, true })
		})
	}
	wg.Wait()
	value, _ := mm.Get("counter")
	assert.Equal(t, 100, value)
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

func BenchmarkConcurrentMultiKeyMapGet(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int]()
			for n := range v.size {
				m.Set(strconv.Itoa(n), n)
			}
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.Get(strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkConcurrentMultiKeyMapSet(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.Set(strconv.Itoa(n), n)
				}
			}
		})
	}
}

func BenchmarkConcurrentMultiKeyMapDelete(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int]()
			for b.Loop() {
				// Refill outside the timer so that every iteration deletes existing keys
				// and their secondary keys.
				b.StopTimer()
				for n := range v.size {
					key := strconv.Itoa(n)
					m.Set(key, n)
					m.AddSecondaryKeys(key, "group", key)
				}
				b.StartTimer()
				for n := range v.size {
					m.Delete(strconv.Itoa(n))
				}
			}
		})
	}
}
