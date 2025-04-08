package bikeymap

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aeimer/go-multikeymap/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleNew() {
	bm := New[string, int, string]()
	_ = bm.Put("keyA1", 1, "value1")
	value, exists := bm.GetByKeyA("keyA1")
	fmt.Printf("[Key A] value: %v, exists: %v\n", value, exists)
	value, exists = bm.GetByKeyB(1)
	fmt.Printf("[Key B] value: %v, exists: %v\n", value, exists)
	// Output:
	// [Key A] value: value1, exists: true
	// [Key B] value: value1, exists: true
}

func TestBiKeyMap_ImplementsContainerInterface(t *testing.T) {
	instance := New[int, int, int]()
	if _, ok := any(instance).(container.Container[int]); !ok {
		t.Error("BiKeyMap does not implement the Container interface")
	}
}

func TestBiKeyMap_SetAndGet(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
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

func TestBiKeyMap_SetDuplicateKeys(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)

	err = bm.Put("keyA2", 1, "value2")
	require.Error(t, err)

	err = bm.Put("keyA1", 2, "value2")
	require.Error(t, err)
}

func TestBiKeyMap_RemoveByKeyA(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)
	err = bm.RemoveByKeyA("keyA1")
	require.NoError(t, err)

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be removed")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be removed")
	}
}

func TestBiKeyMap_RemoveByKeyB(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)
	err = bm.RemoveByKeyB(1)
	require.NoError(t, err)

	_, exists := bm.GetByKeyA("keyA1")
	if exists {
		t.Error("expected keyA1 to be removed")
	}

	_, exists = bm.GetByKeyB(1)
	if exists {
		t.Error("expected keyB 1 to be removed")
	}
}

func TestBiKeyMap_String(t *testing.T) {
	bm := New[string, int, string]()
	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)

	expected := "BiKeyMap: map[keyA1:value1]"
	assert.Equal(t, expected, bm.String())
}

func TestBiKeyMap_RemoveByKeyA_NotFound(t *testing.T) {
	bm := New[string, int, string]()
	err := bm.RemoveByKeyA("nonExistentKey")
	require.Error(t, err)
}

func TestBiKeyMap_RemoveByKeyB_NotFound(t *testing.T) {
	bm := New[string, int, string]()
	err := bm.RemoveByKeyB(999)
	require.Error(t, err)
}
func TestBiKeyMap_EmptyAndSize(t *testing.T) {
	bm := New[string, int, string]()

	if !bm.Empty() {
		t.Error("expected map to be empty")
	}

	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)
	if bm.Empty() {
		t.Error("expected map to not be empty")
	}

	assert.Equal(t, 1, bm.Size())
}

func TestBiKeyMap_Clear(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)
	err = bm.Put("keyA2", 2, "value2")
	require.NoError(t, err)
	bm.Clear()

	if !bm.Empty() {
		t.Error("expected map to be empty after clear")
	}

	assert.Equal(t, 0, bm.Size())
}

func TestBiKeyMap_Values(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Put("keyA1", 1, "value1")
	require.NoError(t, err)
	err = bm.Put("keyA2", 2, "value2")
	require.NoError(t, err)

	values := bm.Values()
	assert.Len(t, values, 2)

	expectedValues := map[string]bool{"value1": true, "value2": true}
	for _, value := range values {
		assert.Contains(t, expectedValues, value)
	}
}

// Benchmarks

var benchmarkSizes = []struct {
	size int
}{
	{size: 100},
	{size: 1000},
	{size: 10_000},
	{size: 100_000},
}

func BenchmarkBiKeyMapGet(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int, string]()
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

func BenchmarkBiKeyMapPut(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int, string]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					_ = m.Put(strconv.Itoa(n), n, strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkBiKeyMapRemove(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int, string]()
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
