package multikeymap

import (
	"fmt"
	"sync"
	"testing"

	"github.com/aeimer/go-multikeymap/container"
)

func ExampleNewMultiKeyMap() {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("keyA1", 1)
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

func TestMultiKeyMap_ImplementsContainerInterface(t *testing.T) {
	instance := NewMultiKeyMap[int, int]()
	if _, ok := any(instance).(container.Container[int]); !ok {
		t.Error("MultiKeyMap does not implement the Container interface")
	}
}

func TestMultiKeyMap_AddAndGet(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	value, exists := mm.Get("key1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestMultiKeyMap_AddSecondaryKeys(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	value, exists := mm.GetBySecondaryKey("group1", "secKey1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestMultiKeyMap_HasPrimaryKey(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	if !mm.HasPrimaryKey("key1") {
		t.Error("expected primary key 'key1' to exist")
	}
}

func TestMultiKeyMap_HasSecondaryKey(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	if !mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to exist")
	}
}

func TestMultiKeyMap_Delete(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.Delete("key1")
	if mm.HasPrimaryKey("key1") {
		t.Error("expected primary key 'key1' to be deleted")
	}
	if mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to be deleted")
	}
}

func TestMultiKeyMap_GetAllKeyGroups(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.Add("key2", 2)
	mm.AddSecondaryKeys("key2", "group2", "secKey3")
	allGroups := mm.GetAllKeyGroups()
	if len(allGroups) != 2 || len(allGroups["group1"]) != 2 || len(allGroups["group2"]) != 1 {
		t.Errorf("expected allGroups to contain 'group1' and 'group2' with correct keys, got %v", allGroups)
	}
}

func TestMultiKeyMap_GetBySecondaryKey_NotFound(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	if _, exists := mm.GetBySecondaryKey("group1", "nonExistentKey"); exists {
		t.Error("expected 'nonExistentKey' to not be found")
	}
}

func TestMultiKeyMap_String(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	expected := "MultiKeyMap: map[key1:1]"
	if mm.String() != expected {
		t.Errorf("expected %s, got %s", expected, mm.String())
	}
}

func TestMultiKeyMap_Size(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.Add("key2", 2)
	if mm.Size() != 2 {
		t.Errorf("expected size 2, got %d", mm.Size())
	}
}

func TestMultiKeyMap_Empty(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	if !mm.Empty() {
		t.Error("expected map to be empty")
	}
	mm.Add("key1", 1)
	if mm.Empty() {
		t.Error("expected map to not be empty")
	}
}

func TestMultiKeyMap_Values(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.Add("key2", 2)
	values := mm.Values()
	expected := []int{1, 2}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("expected value %d, got %d", expected[i], v)
		}
	}
}

func TestMultiKeyMap_Clear(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	mm.Add("key1", 1)
	mm.Clear()
	if !mm.Empty() {
		t.Error("expected map to be empty after clear")
	}
}

func TestMultiKeyMap_ConcurrentAccess(t *testing.T) {
	mm := NewMultiKeyMap[string, int]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrently add values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			mm.Add(key, i)
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

	// Concurrently delete values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			mm.Delete(key)
			if _, exists := mm.Get(key); exists {
				t.Errorf("expected key%d to be deleted", i)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
