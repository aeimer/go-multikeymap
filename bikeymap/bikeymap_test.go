package bikeymap

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleNew() {
	bm := New[string, int, string]()
	_ = bm.Set("keyA1", 1, "value1")
	value, exists := bm.GetByKeyA("keyA1")
	fmt.Printf("[Key A] value: %v, exists: %v\n", value, exists)
	value, exists = bm.GetByKeyB(1)
	fmt.Printf("[Key B] value: %v, exists: %v\n", value, exists)
	// Output:
	// [Key A] value: value1, exists: true
	// [Key B] value: value1, exists: true
}

func TestBiKeyMap_SetAndGet(t *testing.T) {
	bm := New[string, int, string]()

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

func TestBiKeyMap_SetDuplicateKeys(t *testing.T) {
	bm := New[string, int, string]()

	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	err = bm.Set("keyA2", 1, "value2")
	require.Error(t, err)

	err = bm.Set("keyA1", 2, "value2")
	require.Error(t, err)
}

func TestBiKeyMap_DeleteByKeyA(t *testing.T) {
	bm := New[string, int, string]()

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

func TestBiKeyMap_DeleteByKeyB(t *testing.T) {
	bm := New[string, int, string]()

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

func TestBiKeyMap_String(t *testing.T) {
	bm := New[string, int, string]()
	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	expected := "BiKeyMap: map[keyA1:value1]"
	assert.Equal(t, expected, bm.String())
}

func TestBiKeyMap_DeleteByKeyA_NotFound(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	bm.DeleteByKeyA("nonExistentKey")
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_DeleteByKeyB_NotFound(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	bm.DeleteByKeyB(999)
	assert.Equal(t, 1, bm.Len())
}
func TestBiKeyMap_Len(t *testing.T) {
	bm := New[string, int, string]()
	assert.Equal(t, 0, bm.Len())
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_Clear(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	bm.Clear()
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_ValuesReadsMapWhenIterationStarts(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	values := bm.Values()
	bm.Clear()
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.Equal(t, []string{"value2"}, slices.Collect(values))
}

func TestBiKeyMap_ValuesStopsEarly(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	count := 0
	for range bm.Values() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestBiKeyMap_Values(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.ElementsMatch(t, []string{"value1", "value2"}, slices.Collect(bm.Values()))
}

func TestBiKeyMap_ZeroValueReads(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
	_, ok = bm.GetByKeyB(1)
	assert.False(t, ok)
	assert.Equal(t, 0, bm.Len())
	assert.Empty(t, slices.Collect(bm.Values()))
	assert.Equal(t, "BiKeyMap: map[]", bm.String())
	bm.DeleteByKeyA("keyA1")
	bm.DeleteByKeyB(1)
	bm.Clear()
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_ZeroValueWrites(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	value, ok := bm.GetByKeyB(1)
	assert.True(t, ok)
	assert.Equal(t, "value1", value)
	bm.Clear()
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.Equal(t, 1, bm.Len())
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

func BenchmarkBiKeyMapSet(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int, string]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					_ = m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkBiKeyMapDelete(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int, string]()
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
