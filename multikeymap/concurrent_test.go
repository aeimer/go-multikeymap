package multikeymap

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/aeimer/go-multikeymap/container"
)

func ExampleNewConcurrent() {
	mm := NewConcurrent[string, int]()
	mm.Put("keyA1", 1)
	mm.PutSecondaryKeys("keyA1", "group1", "key1", "key2")
	mm.PutSecondaryKeys("keyA1", "group2", "key3", "key4")
	value, exists := mm.Get("keyA1")
	fmt.Printf("[Key A1] value: %v, exists: %v\n", value, exists)
	value, exists = mm.GetBySecondaryKey("group1", "key2")
	fmt.Printf("[Secondary group1 key2] value: %v, exists: %v\n", value, exists)

	// Output:
	// [Key A1] value: 1, exists: true
	// [Secondary group1 key2] value: 1, exists: true
}

func TestConcurrentMultiKeyMap_ImplementsContainerInterface(t *testing.T) {
	instance := NewConcurrent[int, int]()
	if _, ok := any(instance).(container.Container[int]); !ok {
		t.Error("MultiKeyMap does not implement the Container interface")
	}
}

func TestConcurrentMultiKeyMap_SetAndGet(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	value, exists := mm.Get("key1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestConcurrentMultiKeyMap_SetSecondaryKeys(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.PutSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	value, exists := mm.GetBySecondaryKey("group1", "secKey1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestConcurrentMultiKeyMap_HasPrimaryKey(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	if !mm.HasPrimaryKey("key1") {
		t.Error("expected primary key 'key1' to exist")
	}
}

func TestConcurrentMultiKeyMap_HasSecondaryKey(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.PutSecondaryKeys("key1", "group1", "secKey1")
	if !mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to exist")
	}
}

func TestConcurrentMultiKeyMap_Remove(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.PutSecondaryKeys("key1", "group1", "secKey1")
	mm.Remove("key1")
	if mm.HasPrimaryKey("key1") {
		t.Error("expected primary key 'key1' to be removed")
	}
	if mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to be removed")
	}
}

func TestConcurrentMultiKeyMap_GetAllKeyGroups(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.PutSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.Put("key2", 2)
	mm.PutSecondaryKeys("key2", "group2", "secKey3")
	allGroups := mm.GetAllKeyGroups()
	if len(allGroups) != 2 || len(allGroups["group1"]) != 2 || len(allGroups["group2"]) != 1 {
		t.Errorf("expected allGroups to contain 'group1' and 'group2' with correct keys, got %v", allGroups)
	}
}

func TestConcurrentMultiKeyMap_GetBySecondaryKey_NotFound(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.PutSecondaryKeys("key1", "group1", "secKey1")
	if _, exists := mm.GetBySecondaryKey("group1", "nonExistentKey"); exists {
		t.Error("expected 'nonExistentKey' to not be found")
	}
}

func TestConcurrentMultiKeyMap_String(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	expected := "ConcurrentMultiKeyMap: map[key1:1]"
	if mm.String() != expected {
		t.Errorf("expected %s, got %s", expected, mm.String())
	}
}

func TestConcurrentMultiKeyMap_Size(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.Put("key2", 2)
	if mm.Size() != 2 {
		t.Errorf("expected size 2, got %d", mm.Size())
	}
}

func TestConcurrentMultiKeyMap_Empty(t *testing.T) {
	mm := NewConcurrent[string, int]()
	if !mm.Empty() {
		t.Error("expected map to be empty")
	}
	mm.Put("key1", 1)
	if mm.Empty() {
		t.Error("expected map to not be empty")
	}
}

func TestConcurrentMultiKeyMap_Values(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.Put("key2", 2)
	values := mm.Values()
	expected := []int{1, 2}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("expected value %d, got %d", expected[i], v)
		}
	}
}

func TestConcurrentMultiKeyMap_Clear(t *testing.T) {
	mm := NewConcurrent[string, int]()
	mm.Put("key1", 1)
	mm.Clear()
	if !mm.Empty() {
		t.Error("expected map to be empty after clear")
	}
}

func TestConcurrentMultiKeyMap_ConcurrentAccess(t *testing.T) {
	mm := NewConcurrent[string, int]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrently add values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			mm.Put(key, i)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently get values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			if value, exists := mm.Get(key); !exists || value != i {
				t.Errorf("expected %d, got %v", i, value)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently remove values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			mm.Remove(key)
			if _, exists := mm.Get(key); exists {
				t.Errorf("expected key%d to be removed", i)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
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
			for n := 0; n < v.size; n++ {
				m.Put(strconv.Itoa(n), n)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for n := 0; n < v.size; n++ {
					m.Get(strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkConcurrentMultiKeyMapPut(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int]()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for n := 0; n < v.size; n++ {
					m.Put(strconv.Itoa(n), n)
				}
			}
		})
	}
}

func BenchmarkConcurrentMultiKeyMapRemove(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int]()
			for n := 0; n < v.size; n++ {
				m.Put(strconv.Itoa(n), n)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for n := 0; n < v.size; n++ {
					m.Remove(strconv.Itoa(n))
				}
			}
		})
	}
}
