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
	// given
	bm := New[string, int, string]()

	// when
	err := bm.Set("keyA1", 1, "value1")

	// then
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
	// given
	bm := New[string, int, string]()
	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	// when
	// keyB 1 is already paired with keyA1.
	err = bm.Set("keyA2", 1, "value2")

	// then
	require.ErrorIs(t, err, ErrKeyBConflict)
	require.EqualError(t, err, "bikeymap: keyB is paired with a different keyA")

	// when
	// keyA1 is already paired with keyB 1.
	err = bm.Set("keyA1", 2, "value2")

	// then
	require.ErrorIs(t, err, ErrKeyAConflict)
	require.EqualError(t, err, "bikeymap: keyA is paired with a different keyB")

	// given
	require.NoError(t, bm.Set("keyA3", 3, "value3"))

	// when
	// Both keys conflict: the keyB conflict is reported.
	err = bm.Set("keyA1", 3, "value4")

	// then
	require.ErrorIs(t, err, ErrKeyBConflict)
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA3": "value3"}, maps.Collect(bm.All()))
}

func TestBiKeyMap_DeleteByKeyA(t *testing.T) {
	// given
	bm := New[string, int, string]()
	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	// when
	bm.DeleteByKeyA("keyA1")

	// then
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
	// given
	bm := New[string, int, string]()
	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)

	// when
	bm.DeleteByKeyB(1)

	// then
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
	// given
	bm := New[string, int, string]()
	err := bm.Set("keyA1", 1, "value1")
	require.NoError(t, err)
	expected := "BiKeyMap: map[keyA1:value1]"

	// when / then
	assert.Equal(t, expected, bm.String())
}

func TestBiKeyMap_DeleteByKeyA_NotFound(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// when
	bm.DeleteByKeyA("nonExistentKey")

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_DeleteByKeyB_NotFound(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// when
	bm.DeleteByKeyB(999)

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_Len(t *testing.T) {
	// given
	bm := New[string, int, string]()

	// when / then
	assert.Equal(t, 0, bm.Len())

	// when
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_Clear(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when
	bm.Clear()

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_ValuesReadsMapWhenIterationStarts(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// when
	values := bm.Values()
	bm.Clear()
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// then
	assert.Equal(t, []string{"value2"}, slices.Collect(values))
}

func TestBiKeyMap_ValuesStopsEarly(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when
	count := 0
	for range bm.Values() {
		count++
		break
	}

	// then
	assert.Equal(t, 1, count)
}

func TestBiKeyMap_Values(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when / then
	assert.ElementsMatch(t, []string{"value1", "value2"}, slices.Collect(bm.Values()))
}

func TestBiKeyMap_ZeroValueReads(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when / then
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
	_, ok = bm.GetByKeyB(1)
	assert.False(t, ok)
	assert.Equal(t, 0, bm.Len())
	assert.Empty(t, slices.Collect(bm.Values()))
	assert.Equal(t, "BiKeyMap: map[]", bm.String())

	// when
	bm.DeleteByKeyA("keyA1")
	bm.DeleteByKeyB(1)
	bm.Clear()

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_ZeroValueWrites(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// then
	value, ok := bm.GetByKeyB(1)
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// when
	bm.Clear()
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_All(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when / then
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2"}, maps.Collect(bm.All()))
}

func TestBiKeyMap_Keys(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when / then
	assert.ElementsMatch(t, []string{"keyA1", "keyA2"}, slices.Collect(bm.KeysA()))
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(bm.KeysB()))
}

func TestBiKeyMap_IteratorsStopEarly(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when
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

	// then
	assert.Equal(t, 3, count)
}

func TestBiKeyMap_IteratorsReadMapWhenIterationStarts(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when
	all, keysA, keysB := bm.All(), bm.KeysA(), bm.KeysB()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// then
	assert.Equal(t, map[string]string{"keyA1": "value1"}, maps.Collect(all))
	assert.Equal(t, []string{"keyA1"}, slices.Collect(keysA))
	assert.Equal(t, []int{1}, slices.Collect(keysB))
}

func TestBiKeyMap_KeyAAndKeyB(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// when
	keyA, ok := bm.KeyA(1)

	// then
	assert.True(t, ok)
	assert.Equal(t, "keyA1", keyA)

	// when
	keyB, ok := bm.KeyB("keyA1")

	// then
	assert.True(t, ok)
	assert.Equal(t, 1, keyB)

	// when
	keyA, ok = bm.KeyA(2)

	// then
	assert.False(t, ok)
	assert.Empty(t, keyA)

	// when
	keyB, ok = bm.KeyB("missing")

	// then
	assert.False(t, ok)
	assert.Zero(t, keyB)
}

func TestBiKeyMap_LoadAndDeleteByKeyA(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when
	_, loaded := bm.LoadAndDeleteByKeyA("keyA1")

	// then
	assert.False(t, loaded)

	// given
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when
	value, loaded := bm.LoadAndDeleteByKeyA("keyA1")

	// then
	assert.True(t, loaded)
	assert.Equal(t, "value1", value)
	_, ok := bm.GetByKeyB(1)
	assert.False(t, ok)
	assert.Equal(t, 1, bm.Len())

	// when
	_, loaded = bm.LoadAndDeleteByKeyA("keyA1")

	// then
	assert.False(t, loaded)
}

func TestBiKeyMap_LoadAndDeleteByKeyB(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when
	_, loaded := bm.LoadAndDeleteByKeyB(1)

	// then
	assert.False(t, loaded)

	// given
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when
	value, loaded := bm.LoadAndDeleteByKeyB(1)

	// then
	assert.True(t, loaded)
	assert.Equal(t, "value1", value)
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
	assert.Equal(t, 1, bm.Len())

	// when
	_, loaded = bm.LoadAndDeleteByKeyB(1)

	// then
	assert.False(t, loaded)
}

func TestBiKeyMap_Clone(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))

	// when
	clone := bm.Clone()

	// then
	assert.Equal(t, maps.Collect(bm.All()), maps.Collect(clone.All()))

	// when
	// Changes to the clone and to the original stay apart.
	require.NoError(t, clone.Set("keyA1", 1, "changed"))
	require.NoError(t, clone.Set("keyA3", 3, "value3"))
	clone.DeleteByKeyB(2)
	require.NoError(t, bm.Set("keyA4", 4, "value4"))

	// then
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2", "keyA4": "value4"}, maps.Collect(bm.All()))
	assert.ElementsMatch(t, []int{1, 2, 4}, slices.Collect(bm.KeysB()))
	assert.Equal(t, map[string]string{"keyA1": "changed", "keyA3": "value3"}, maps.Collect(clone.All()))
	assert.ElementsMatch(t, []int{1, 3}, slices.Collect(clone.KeysB()))
}

func TestBiKeyMap_CloneCopiesValuesShallowly(t *testing.T) {
	// given
	bm := New[string, int, *int]()
	value := 1
	require.NoError(t, bm.Set("keyA1", 1, &value))

	// when
	clone := bm.Clone()

	// then
	got, _ := clone.GetByKeyA("keyA1")
	assert.Same(t, &value, got)
}

func TestBiKeyMap_CloneZeroValue(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when
	clone := bm.Clone()

	// then
	assert.Equal(t, 0, clone.Len())

	// when
	require.NoError(t, clone.Set("keyA1", 1, "value1"))

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_DeleteFunc(t *testing.T) {
	// given
	bm := New[string, int, string]()
	for i := range 4 {
		require.NoError(t, bm.Set("keyA"+strconv.Itoa(i), i, "value"+strconv.Itoa(i)))
	}

	// when
	bm.DeleteFunc(func(keyA string, keyB int, value string) bool {
		assert.Equal(t, "keyA"+strconv.Itoa(keyB), keyA)
		assert.Equal(t, "value"+strconv.Itoa(keyB), value)
		return keyB%2 == 1
	})

	// then
	assert.Equal(t, map[string]string{"keyA0": "value0", "keyA2": "value2"}, maps.Collect(bm.All()))
	assert.ElementsMatch(t, []int{0, 2}, slices.Collect(bm.KeysB()))
}

func TestBiKeyMap_DeleteFuncZeroValue(t *testing.T) {
	// given
	var bm BiKeyMap[string, int, string]

	// when
	bm.DeleteFunc(func(string, int, string) bool { return true })

	// then
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

func TestBiKeyMap_Clear_DuringRange(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))
	require.NoError(t, bm.Set("keyA2", 2, "value2"))
	require.NoError(t, bm.Set("keyA3", 3, "value3"))

	// when
	yielded := 0
	for range bm.All() {
		yielded++
		bm.Clear()
	}

	// then
	assert.Equal(t, 1, yielded)
	assert.Equal(t, 0, bm.Len())
	_, ok := bm.KeyA(1)
	assert.False(t, ok)

	// when
	require.NoError(t, bm.Set("keyA4", 1, "value4"))

	// then
	value, ok := bm.GetByKeyB(1)
	assert.True(t, ok)
	assert.Equal(t, "value4", value)
}
