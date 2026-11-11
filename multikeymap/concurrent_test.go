package multikeymap

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"testing"

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
			for n := range v.size {
				m.Set(strconv.Itoa(n), n)
			}
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.Delete(strconv.Itoa(n))
				}
			}
		})
	}
}
