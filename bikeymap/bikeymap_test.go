package bikeymap

import (
	"fmt"
	"maps"
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

func TestBiKeyMap_All(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2"}, maps.Collect(bm.All()))
}

func TestBiKeyMap_Keys(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	assert.ElementsMatch(t, []string{"keyA1", "keyA2"}, slices.Collect(bm.KeysA()))
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(bm.KeysB()))
}

func TestBiKeyMap_IteratorsStopEarly(t *testing.T) {
	bm := New[string, int, string]()
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

func TestBiKeyMap_IteratorsReadMapWhenIterationStarts(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	all, keysA, keysB := bm.All(), bm.KeysA(), bm.KeysB()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	assert.Equal(t, map[string]string{"keyA1": "value1"}, maps.Collect(all))
	assert.Equal(t, []string{"keyA1"}, slices.Collect(keysA))
	assert.Equal(t, []int{1}, slices.Collect(keysB))
}

func TestBiKeyMap_KeyAAndKeyB(t *testing.T) {
	bm := New[string, int, string]()
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

func TestBiKeyMap_LoadAndDeleteByKeyA(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	_, loaded := bm.LoadAndDeleteByKeyA("keyA1")
	assert.False(t, loaded)

	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	value, loaded := bm.LoadAndDeleteByKeyA("keyA1")
	assert.True(t, loaded)
	assert.Equal(t, "value1", value)
	_, ok := bm.GetByKeyB(1)
	assert.False(t, ok)
	assert.Equal(t, 1, bm.Len())

	_, loaded = bm.LoadAndDeleteByKeyA("keyA1")
	assert.False(t, loaded)
}

func TestBiKeyMap_LoadAndDeleteByKeyB(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	_, loaded := bm.LoadAndDeleteByKeyB(1)
	assert.False(t, loaded)

	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	value, loaded := bm.LoadAndDeleteByKeyB(1)
	assert.True(t, loaded)
	assert.Equal(t, "value1", value)
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
	assert.Equal(t, 1, bm.Len())

	_, loaded = bm.LoadAndDeleteByKeyB(1)
	assert.False(t, loaded)
}

func TestBiKeyMap_Clone(t *testing.T) {
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	clone := bm.Clone()
	assert.Equal(t, maps.Collect(bm.All()), maps.Collect(clone.All()))

	// Changes to the clone and to the original stay apart.
	require.NoError(t, clone.Set("keyA1", 1, "changed"))
	require.NoError(t, clone.Set("keyA3", 3, "value3"))
	clone.DeleteByKeyB(2)
	require.NoError(t, bm.Set("keyA4", 4, "value4"))

	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2", "keyA4": "value4"}, maps.Collect(bm.All()))
	assert.ElementsMatch(t, []int{1, 2, 4}, slices.Collect(bm.KeysB()))
	assert.Equal(t, map[string]string{"keyA1": "changed", "keyA3": "value3"}, maps.Collect(clone.All()))
	assert.ElementsMatch(t, []int{1, 3}, slices.Collect(clone.KeysB()))
}

func TestBiKeyMap_CloneCopiesValuesShallowly(t *testing.T) {
	bm := New[string, int, *int]()
	value := 1
	require.NoError(t, bm.Set("keyA1", 1, &value))
	clone := bm.Clone()
	got, _ := clone.GetByKeyA("keyA1")
	assert.Same(t, &value, got)
}

func TestBiKeyMap_CloneZeroValue(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	clone := bm.Clone()
	assert.Equal(t, 0, clone.Len())
	require.NoError(t, clone.Set("keyA1", 1, "value1"))
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_DeleteFunc(t *testing.T) {
	bm := New[string, int, string]()
	for i := range 4 {
		require.NoError(t, bm.Set("keyA"+strconv.Itoa(i), i, "value"+strconv.Itoa(i)))
	}
	bm.DeleteFunc(func(keyA string, keyB int, value string) bool {
		assert.Equal(t, "keyA"+strconv.Itoa(keyB), keyA)
		assert.Equal(t, "value"+strconv.Itoa(keyB), value)
		return keyB%2 == 1
	})
	assert.Equal(t, map[string]string{"keyA0": "value0", "keyA2": "value2"}, maps.Collect(bm.All()))
	assert.ElementsMatch(t, []int{0, 2}, slices.Collect(bm.KeysB()))
}

func TestBiKeyMap_DeleteFuncZeroValue(t *testing.T) {
	var bm BiKeyMap[string, int, string]
	bm.DeleteFunc(func(string, int, string) bool { return true })
	assert.Equal(t, 0, bm.Len())
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
			for b.Loop() {
				// Refill outside the timer so that every iteration deletes existing keys.
				b.StopTimer()
				for n := range v.size {
					_ = m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
				}
				b.StartTimer()
				for n := range v.size {
					m.DeleteByKeyB(n)
				}
			}
		})
	}
}
