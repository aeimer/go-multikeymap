package bikeymap

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleNew() {
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")
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
	bm.Set("keyA1", 1, "value1")

	// then
	value, exists := bm.GetByKeyA("keyA1")
	if !exists || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}
	value, exists = bm.GetByKeyB(1)
	if !exists || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}
}

func TestBiKeyMap_Set_Overwrite(t *testing.T) {
	tests := []struct {
		name      string
		givenKeyA string
		givenKeyB int
		wantAll   map[string]string
		wantPairs map[string]int
	}{
		{
			name:      "new pair",
			givenKeyA: "keyA3", givenKeyB: 3,
			wantAll:   map[string]string{"keyA1": "value1", "keyA2": "value2", "keyA3": "new"},
			wantPairs: map[string]int{"keyA1": 1, "keyA2": 2, "keyA3": 3},
		},
		{
			name:      "same pair replaces the value",
			givenKeyA: "keyA1", givenKeyB: 1,
			wantAll:   map[string]string{"keyA1": "new", "keyA2": "value2"},
			wantPairs: map[string]int{"keyA1": 1, "keyA2": 2},
		},
		{
			name:      "keyA paired with another keyB",
			givenKeyA: "keyA1", givenKeyB: 3,
			wantAll:   map[string]string{"keyA1": "new", "keyA2": "value2"},
			wantPairs: map[string]int{"keyA1": 3, "keyA2": 2},
		},
		{
			name:      "keyB paired with another keyA",
			givenKeyA: "keyA3", givenKeyB: 1,
			wantAll:   map[string]string{"keyA2": "value2", "keyA3": "new"},
			wantPairs: map[string]int{"keyA2": 2, "keyA3": 1},
		},
		{
			name:      "both keys paired with others",
			givenKeyA: "keyA1", givenKeyB: 2,
			wantAll:   map[string]string{"keyA1": "new"},
			wantPairs: map[string]int{"keyA1": 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			bm := New[string, int, string]()
			bm.Set("keyA1", 1, "value1")
			bm.Set("keyA2", 2, "value2")

			// when
			bm.Set(tt.givenKeyA, tt.givenKeyB, "new")

			// then
			assert.Equal(t, tt.wantAll, maps.Collect(bm.All()))
			assert.Equal(t, len(tt.wantAll), bm.Len())
			for keyA, keyB := range tt.wantPairs {
				gotKeyB, ok := bm.KeyB(keyA)
				assert.True(t, ok)
				assert.Equal(t, keyB, gotKeyB)
				gotKeyA, ok := bm.KeyA(keyB)
				assert.True(t, ok)
				assert.Equal(t, keyA, gotKeyA)
			}
			// Both indexes hold exactly the same pairs.
			assert.ElementsMatch(t, slices.Collect(maps.Values(tt.wantPairs)), slices.Collect(bm.KeysB()))
		})
	}
}

func TestBiKeyMap_DeleteByKeyA(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")

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
	bm.Set("keyA1", 1, "value1")

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
	bm.Set("keyA1", 1, "value1")
	expected := "BiKeyMap: map[keyA1:value1]"

	// when / then
	assert.Equal(t, expected, bm.String())
}

func TestBiKeyMap_DeleteByKeyA_NotFound(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")

	// when
	bm.DeleteByKeyA("nonExistentKey")

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_DeleteByKeyB_NotFound(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")

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
	bm.Set("keyA1", 1, "value1")

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_Clear(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when
	bm.Clear()

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_ValuesReadsMapWhenIterationStarts(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")

	// when
	values := bm.Values()
	bm.Clear()
	bm.Set("keyA2", 2, "value2")

	// then
	assert.Equal(t, []string{"value2"}, slices.Collect(values))
}

func TestBiKeyMap_ValuesStopsEarly(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

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
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

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
	bm.Set("keyA1", 1, "value1")

	// then
	value, ok := bm.GetByKeyB(1)
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// when
	bm.Clear()
	bm.Set("keyA2", 2, "value2")

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestBiKeyMap_All(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when / then
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2"}, maps.Collect(bm.All()))
}

func TestBiKeyMap_Keys(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when / then
	assert.ElementsMatch(t, []string{"keyA1", "keyA2"}, slices.Collect(bm.KeysA()))
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(bm.KeysB()))
}

func TestBiKeyMap_IteratorsStopEarly(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

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
	bm.Set("keyA1", 1, "value1")

	// then
	assert.Equal(t, map[string]string{"keyA1": "value1"}, maps.Collect(all))
	assert.Equal(t, []string{"keyA1"}, slices.Collect(keysA))
	assert.Equal(t, []int{1}, slices.Collect(keysB))
}

func TestBiKeyMap_KeyAAndKeyB(t *testing.T) {
	// given
	bm := New[string, int, string]()
	bm.Set("keyA1", 1, "value1")

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
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

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
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

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
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when
	clone := bm.Clone()

	// then
	assert.Equal(t, maps.Collect(bm.All()), maps.Collect(clone.All()))

	// when
	// Changes to the clone and to the original stay apart.
	clone.Set("keyA1", 1, "changed")
	clone.Set("keyA3", 3, "value3")
	clone.DeleteByKeyB(2)
	bm.Set("keyA4", 4, "value4")

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
	bm.Set("keyA1", 1, &value)

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
	clone.Set("keyA1", 1, "value1")

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestBiKeyMap_DeleteFunc(t *testing.T) {
	// given
	bm := New[string, int, string]()
	for i := range 4 {
		bm.Set("keyA"+strconv.Itoa(i), i, "value"+strconv.Itoa(i))
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
				m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
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
					m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
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
					m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
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
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")
	bm.Set("keyA3", 3, "value3")

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
	bm.Set("keyA4", 1, "value4")

	// then
	value, ok := bm.GetByKeyB(1)
	assert.True(t, ok)
	assert.Equal(t, "value4", value)
}
