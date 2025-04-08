package bikeymap

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/aeimer/go-multikeymap/container"
)

func ExampleNewConcurrent() {
	bm := NewConcurrent[string, int, string]()
	_ = bm.Put("keyA1", 1, "value1")
	value, exists := bm.GetByKeyA("keyA1")
	fmt.Printf("[Key A] value: %v, exists: %v\n", value, exists)
	value, exists = bm.GetByKeyB(1)
	fmt.Printf("[Key B] value: %v, exists: %v\n", value, exists)
	// Output:
	// [Key A] value: value1, exists: true
	// [Key B] value: value1, exists: true
}

func TestConcurrentBiKeyMap_ImplementsContainerInterface(t *testing.T) {
	instance := NewConcurrent[int, int, int]()
	if _, ok := any(instance).(container.Container[int]); !ok {
		t.Error("ConcurrentBiKeyMap does not implement the Container interface")
	}
}

func TestConcurrentBiKeyMap_SetAndGet(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
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

func TestConcurrentBiKeyMap_SetDuplicateKeys(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = bm.Put("keyA2", 1, "value2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	err = bm.Put("keyA1", 2, "value2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConcurrentBiKeyMap_RemoveByKeyA(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	_ = bm.Put("keyA1", 1, "value1")
	err := bm.RemoveByKeyA("keyA1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be removed")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be removed")
	}
}

func TestConcurrentBiKeyMap_RemoveByKeyB(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	_ = bm.Put("keyA1", 1, "value1")
	err := bm.RemoveByKeyB(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	_ = bm.Put("keyA1", 1, "value1")
	expected := "ConcurrentBiKeyMap: map[keyA1:value1]"
	if bm.String() != expected {
		t.Errorf("expected %s, got %s", expected, bm.String())
	}
}

func TestConcurrentBiKeyMap_RemoveByKeyA_NotFound(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	err := bm.RemoveByKeyA("nonExistentKey")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestConcurrentBiKeyMap_RemoveByKeyB_NotFound(t *testing.T) {
	bm := NewConcurrent[string, int, string]()
	err := bm.RemoveByKeyB(999)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
func TestConcurrentBiKeyMap_EmptyAndSize(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	if !bm.Empty() {
		t.Error("expected map to be empty")
	}

	_ = bm.Put("keyA1", 1, "value1")
	if bm.Empty() {
		t.Error("expected map to not be empty")
	}

	if bm.Size() != 1 {
		t.Errorf("expected size 1, got %d", bm.Size())
	}
}

func TestConcurrentBiKeyMap_Clear(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	_ = bm.Put("keyA1", 1, "value1")
	_ = bm.Put("keyA2", 2, "value2")
	bm.Clear()

	if !bm.Empty() {
		t.Error("expected map to be empty after clear")
	}

	if bm.Size() != 0 {
		t.Errorf("expected size 0, got %d", bm.Size())
	}
}

func TestConcurrentBiKeyMap_Values(t *testing.T) {
	bm := NewConcurrent[string, int, string]()

	_ = bm.Put("keyA1", 1, "value1")
	_ = bm.Put("keyA2", 2, "value2")

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
		go concurrentRemove(t, i, &wg, bm)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func concurrentSet(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	value := fmt.Sprintf("value%d", i)
	if err := bm.Put(keyA, keyB, value); err != nil {
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

func concurrentRemove(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	if err := bm.RemoveByKeyA(keyA); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, exists := bm.GetByKeyA(keyA); exists {
		t.Errorf("expected keyA%d to be removed", i)
	}
	if _, exists := bm.GetByKeyB(keyB); exists {
		t.Errorf("expected keyB%d to be removed", i)
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

func BenchmarkConcurrentBiKeyMapGet(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			for n := range v.size {
				_ = m.Put(strconv.Itoa(n), n, strconv.Itoa(n))
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

func BenchmarkConcurrentBiKeyMapPut(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					_ = m.Put(strconv.Itoa(n), n, strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkConcurrentBiKeyMapRemove(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			for n := range v.size {
				_ = m.Put(strconv.Itoa(n), n, strconv.Itoa(n))
			}
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					_ = m.RemoveByKeyB(n)
				}
			}
		})
	}
}
