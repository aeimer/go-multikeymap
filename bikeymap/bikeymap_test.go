package bikeymap

import (
	"fmt"
	"github.com/aeimer/go-multikeymap/container"
	"sync"
	"testing"
)

func ExampleNewBiKeyMap() {
	bm := NewBiKeyMap[string, int, string]()
	_ = bm.Set("keyA1", 1, "value1")
	value, exists := bm.GetByKeyA("keyA1")
	fmt.Printf("[Key A] value: %v, exists: %v\n", value, exists)
	value, exists = bm.GetByKeyB(1)
	fmt.Printf("[Key B] value: %v, exists: %v\n", value, exists)
	// Output:
	// [Key A] value: value1, exists: true
	// [Key B] value: value1, exists: true
}

func TestBiKeyMap_ImplementsContainerInterface(t *testing.T) {
	instance := NewBiKeyMap[int, int, int]()
	if _, ok := any(instance).(container.Container[int]); !ok {
		t.Error("BiKeyMap does not implement the Container interface")
	}
}

func TestBiKeyMap_SetAndGet(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	value, exists := bm.GetByKeyA("keyA1")
	if !exists || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}

	value, exists = bm.GetByKeyB(1)
	if !exists || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}
}

func TestBiKeyMap_SetDuplicateKeys(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = bm.Set("keyA2", 1, "value2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	err = bm.Set("keyA1", 2, "value2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBiKeyMap_DeleteByKeyA(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	bm.Set("keyA1", 1, "value1")
	err := bm.DeleteByKeyA("keyA1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be deleted")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be deleted")
	}
}

func TestBiKeyMap_DeleteByKeyB(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	bm.Set("keyA1", 1, "value1")
	err := bm.DeleteByKeyB(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be deleted")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be deleted")
	}
}

func TestBiKeyMap_EmptyAndSize(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	if !bm.Empty() {
		t.Error("expected map to be empty")
	}

	bm.Set("keyA1", 1, "value1")
	if bm.Empty() {
		t.Error("expected map to not be empty")
	}

	if bm.Size() != 1 {
		t.Errorf("expected size 1, got %d", bm.Size())
	}
}

func TestBiKeyMap_Clear(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")
	bm.Clear()

	if !bm.Empty() {
		t.Error("expected map to be empty after clear")
	}

	if bm.Size() != 0 {
		t.Errorf("expected size 0, got %d", bm.Size())
	}
}

func TestBiKeyMap_Values(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()

	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	values := bm.Values()
	if len(values) != 2 {
		t.Errorf("expected 2 values, got %d", len(values))
	}

	expectedValues := map[string]bool{"value1": true, "value2": true}
	for _, value := range values {
		if !expectedValues[value] {
			t.Errorf("unexpected value: %v", value)
		}
	}
}

func TestBiKeyMap_ConcurrentAccess(t *testing.T) {
	bm := NewBiKeyMap[string, int, string]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrently set values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			keyA := fmt.Sprintf("keyA%d", i)
			keyB := i
			value := fmt.Sprintf("value%d", i)
			if err := bm.Set(keyA, keyB, value); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Concurrently get values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			keyA := fmt.Sprintf("keyA%d", i)
			keyB := i
			if value, exists := bm.GetByKeyA(keyA); !exists || value != fmt.Sprintf("value%d", i) {
				t.Errorf("expected value%d, got %v", i, value)
			}
			if value, exists := bm.GetByKeyB(keyB); !exists || value != fmt.Sprintf("value%d", i) {
				t.Errorf("expected value%d, got %v", i, value)
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
			keyA := fmt.Sprintf("keyA%d", i)
			keyB := i
			if err := bm.DeleteByKeyA(keyA); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if _, exists := bm.GetByKeyA(keyA); exists {
				t.Errorf("expected keyA%d to be deleted", i)
			}
			if _, exists := bm.GetByKeyB(keyB); exists {
				t.Errorf("expected keyB%d to be deleted", i)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}