package multikeymap

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleNew() {
	mm := New[string, int]()
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

func TestMultiKeyMap_SetAndGet(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	value, exists := mm.Get("key1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestMultiKeyMap_SetSecondaryKeys(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	value, exists := mm.GetBySecondaryKey("group1", "secKey1")
	if !exists || value != 1 {
		t.Errorf("expected value 1, got %v, exists: %v", value, exists)
	}
}

func TestMultiKeyMap_AddSecondaryKeysReportsMissingPrimaryKey(t *testing.T) {
	mm := New[string, int]()
	assert.False(t, mm.AddSecondaryKeys("missing", "group1", "secKey1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())

	mm.Set("key1", 1)
	assert.True(t, mm.AddSecondaryKeys("key1", "group1", "secKey1"))
	assert.True(t, mm.HasSecondaryKey("group1", "secKey1"))
}

func TestMultiKeyMap_Has(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	if !mm.Has("key1") {
		t.Error("expected primary key 'key1' to exist")
	}
}

func TestMultiKeyMap_HasSecondaryKey(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	if !mm.HasSecondaryKey("group1", "secKey1") {
		t.Error("expected secondary key 'secKey1' in group 'group1' to exist")
	}
}

func TestMultiKeyMap_Delete(t *testing.T) {
	mm := New[string, int]()
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

func TestMultiKeyMap_DeleteMissingKeyIsNoOp(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Delete("missing")
	assert.Equal(t, 1, mm.Len())
}

func TestMultiKeyMap_DeleteDeletesAllSecondaryKeysOfGroup(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey3")

	mm.Delete("key1")

	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey2"))
	assert.False(t, mm.HasSecondaryKey("group2", "secKey3"))
	assert.Empty(t, mm.KeyGroups())
	assert.Empty(t, mm.secondaryTo)
}

func TestMultiKeyMap_DeleteKeepsOtherPrimaryKeysInGroup(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.AddSecondaryKeys("key2", "group1", "secKey2")

	mm.Delete("key1")

	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	value, exists := mm.GetBySecondaryKey("group1", "secKey2")
	assert.True(t, exists)
	assert.Equal(t, 2, value)
}

func TestMultiKeyMap_AddSecondaryKeysMovesKeyToNewPrimaryKey(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey2", "secKey1")
	mm.AddSecondaryKeys("key2", "group1", "secKey1")

	value, exists := mm.GetBySecondaryKey("group1", "secKey1")
	assert.True(t, exists)
	assert.Equal(t, 2, value)

	// Removing the old owner must not delete the moved key.
	mm.Delete("key1")
	value, exists = mm.GetBySecondaryKey("group1", "secKey1")
	assert.True(t, exists)
	assert.Equal(t, 2, value)
	assert.False(t, mm.HasSecondaryKey("group1", "secKey2"))

	mm.Delete("key2")
	assert.Empty(t, mm.KeyGroups())
	assert.Empty(t, mm.secondaryTo)
}

func TestMultiKeyMap_AddSecondaryKeysMoveDropsEmptyReverseEntry(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.AddSecondaryKeys("key2", "group1", "secKey1")

	assert.NotContains(t, mm.secondaryTo, "key1")
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey1": "key2"}}, mm.KeyGroups())
}

func TestMultiKeyMap_AddSecondaryKeysWithoutKeys(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1")

	assert.Empty(t, mm.KeyGroups())
	assert.Empty(t, mm.secondaryTo)
}

func TestMultiKeyMap_KeyGroups(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key2", "group2", "secKey3")
	allGroups := mm.KeyGroups()
	if len(allGroups) != 2 || len(allGroups["group1"]) != 2 || len(allGroups["group2"]) != 1 {
		t.Errorf("expected allGroups to contain 'group1' and 'group2' with correct keys, got %v", allGroups)
	}
}

func TestMultiKeyMap_GetBySecondaryKey_NotFound(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	if _, exists := mm.GetBySecondaryKey("group1", "nonExistentKey"); exists {
		t.Error("expected 'nonExistentKey' to not be found")
	}
}

func TestMultiKeyMap_String(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)

	expected := "MultiKeyMap: map[key1:1]"
	assert.Equal(t, expected, mm.String())
}

func TestMultiKeyMap_Len(t *testing.T) {
	mm := New[string, int]()
	assert.Equal(t, 0, mm.Len())
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.Equal(t, 2, mm.Len())
}

func TestMultiKeyMap_Values(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(mm.Values()))
}

func TestMultiKeyMap_ValuesReadsMapWhenIterationStarts(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	values := mm.Values()
	mm.Clear()
	mm.Set("key2", 2)
	assert.Equal(t, []int{2}, slices.Collect(values))
}

func TestMultiKeyMap_ValuesStopsEarly(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	count := 0
	for range mm.Values() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestMultiKeyMap_Clear(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Clear()
	if mm.Len() != 0 {
		t.Error("expected map to be empty after clear")
	}
}

func TestMultiKeyMap_ZeroValueReads(t *testing.T) {
	var mm MultiKeyMap[string, int]
	_, ok := mm.Get("key1")
	assert.False(t, ok)
	_, ok = mm.GetBySecondaryKey("group1", "secKey1")
	assert.False(t, ok)
	assert.False(t, mm.Has("key1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())
	assert.Equal(t, 0, mm.Len())
	assert.Empty(t, slices.Collect(mm.Values()))
	assert.Equal(t, "MultiKeyMap: map[]", mm.String())
	assert.False(t, mm.AddSecondaryKeys("key1", "group1", "secKey1"))
	mm.Delete("key1")
	mm.Clear()
	assert.Equal(t, 0, mm.Len())
}

func TestMultiKeyMap_ZeroValueWrites(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.Set("key1", 1)
	assert.True(t, mm.AddSecondaryKeys("key1", "group1", "secKey1"))
	value, ok := mm.GetBySecondaryKey("group1", "secKey1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
	mm.Clear()
	mm.Set("key2", 2)
	assert.True(t, mm.AddSecondaryKeys("key2", "group1", "secKey2"))
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey2": "key2"}}, mm.KeyGroups())
}

func TestMultiKeyMap_All(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.Equal(t, map[string]int{"key1": 1, "key2": 2}, maps.Collect(mm.All()))
}

func TestMultiKeyMap_Keys(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	assert.ElementsMatch(t, []string{"key1", "key2"}, slices.Collect(mm.Keys()))
}

func TestMultiKeyMap_IteratorsStopEarly(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	count := 0
	for range mm.All() {
		count++
		break
	}
	for range mm.Keys() {
		count++
		break
	}
	assert.Equal(t, 2, count)
}

func TestMultiKeyMap_IteratorsReadMapWhenIterationStarts(t *testing.T) {
	var mm MultiKeyMap[string, int]
	all, keys := mm.All(), mm.Keys()
	mm.Set("key1", 1)
	assert.Equal(t, map[string]int{"key1": 1}, maps.Collect(all))
	assert.Equal(t, []string{"key1"}, slices.Collect(keys))
}

func TestMultiKeyMap_PrimaryKey(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")

	pk, ok := mm.PrimaryKey("group1", "secKey1")
	assert.True(t, ok)
	assert.Equal(t, "key1", pk)

	pk, ok = mm.PrimaryKey("group1", "missing")
	assert.False(t, ok)
	assert.Empty(t, pk)
	_, ok = mm.PrimaryKey("missing", "secKey1")
	assert.False(t, ok)
}

func TestMultiKeyMap_SecondaryKeys(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey3")
	mm.AddSecondaryKeys("key2", "group1", "secKey4")

	type groupKey struct{ group, key string }
	var got []groupKey
	for group, key := range mm.SecondaryKeys("key1") {
		got = append(got, groupKey{group, key})
	}
	assert.ElementsMatch(t, []groupKey{{"group1", "secKey1"}, {"group1", "secKey2"}, {"group2", "secKey3"}}, got)
	for group, key := range mm.SecondaryKeys("missing") {
		t.Errorf("unexpected secondary key %s/%s", group, key)
	}
}

func TestMultiKeyMap_SecondaryKeysStopsEarly(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey3")
	count := 0
	for range mm.SecondaryKeys("key1") {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestMultiKeyMap_LoadOrStore(t *testing.T) {
	var mm MultiKeyMap[string, int]
	actual, loaded := mm.LoadOrStore("key1", 1)
	assert.False(t, loaded)
	assert.Equal(t, 1, actual)

	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	actual, loaded = mm.LoadOrStore("key1", 2)
	assert.True(t, loaded)
	assert.Equal(t, 1, actual)
	value, _ := mm.GetBySecondaryKey("group1", "secKey1")
	assert.Equal(t, 1, value)
}

func TestMultiKeyMap_LoadAndDelete(t *testing.T) {
	var mm MultiKeyMap[string, int]
	_, loaded := mm.LoadAndDelete("key1")
	assert.False(t, loaded)

	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	value, loaded := mm.LoadAndDelete("key1")
	assert.True(t, loaded)
	assert.Equal(t, 1, value)
	assert.False(t, mm.Has("key1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())

	_, loaded = mm.LoadAndDelete("key1")
	assert.False(t, loaded)
}

func TestMultiKeyMap_UpdateCreatesMissingKey(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.Update("key1", func(old int, exists bool) (int, bool) {
		assert.False(t, exists)
		assert.Equal(t, 0, old)
		return 1, true
	})
	value, ok := mm.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
}

func TestMultiKeyMap_UpdateReplacesValueAndKeepsSecondaryKeys(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.Update("key1", func(old int, exists bool) (int, bool) {
		assert.True(t, exists)
		return old + 1, true
	})
	value, ok := mm.GetBySecondaryKey("group1", "secKey1")
	assert.True(t, ok)
	assert.Equal(t, 2, value)
}

func TestMultiKeyMap_UpdateDeletesExistingKey(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.Update("key1", func(int, bool) (int, bool) { return 0, false })
	assert.False(t, mm.Has("key1"))
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.Empty(t, mm.KeyGroups())
}

func TestMultiKeyMap_UpdateWithoutKeepOnMissingKeyIsNoOp(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.Update("key1", func(int, bool) (int, bool) { return 1, false })
	assert.False(t, mm.Has("key1"))
	assert.Equal(t, 0, mm.Len())
}

func TestMultiKeyMap_DeleteSecondaryKey(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key2", "group1", "secKey3")

	mm.DeleteSecondaryKey("group1", "secKey1")
	assert.False(t, mm.HasSecondaryKey("group1", "secKey1"))
	assert.True(t, mm.Has("key1"))
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey2": "key1", "secKey3": "key2"}}, mm.KeyGroups())

	mm.DeleteSecondaryKey("group1", "secKey2")
	assert.Empty(t, maps.Collect(mm.SecondaryKeys("key1")))
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey3": "key2"}}, mm.KeyGroups())

	mm.DeleteSecondaryKey("group1", "secKey3")
	assert.Empty(t, mm.KeyGroups())
	assert.Empty(t, mm.secondaryTo)
	assert.Equal(t, 2, mm.Len())
}

func TestMultiKeyMap_DeleteSecondaryKeyMissingIsNoOp(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.DeleteSecondaryKey("group1", "secKey1")
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.DeleteSecondaryKey("group1", "missing")
	mm.DeleteSecondaryKey("missing", "secKey1")
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey1": "key1"}}, mm.KeyGroups())
}

func TestMultiKeyMap_DeleteGroup(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey1")
	mm.AddSecondaryKeys("key2", "group1", "secKey3")

	mm.DeleteGroup("group1")
	assert.Equal(t, map[string]map[string]string{"group2": {"secKey1": "key1"}}, mm.KeyGroups())
	assert.Equal(t, map[string]string{"group2": "secKey1"}, maps.Collect(mm.SecondaryKeys("key1")))
	assert.Empty(t, maps.Collect(mm.SecondaryKeys("key2")))
	assert.NotContains(t, mm.secondaryTo, "key2")
	assert.Equal(t, 2, mm.Len())

	mm.DeleteGroup("group2")
	assert.Empty(t, mm.KeyGroups())
	assert.Empty(t, mm.secondaryTo)
}

func TestMultiKeyMap_DeleteGroupMissingIsNoOp(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.DeleteGroup("group1")
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.DeleteGroup("missing")
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey1": "key1"}}, mm.KeyGroups())
}

func TestMultiKeyMap_DeleteBySecondaryKey(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key1", "group2", "secKey1")
	mm.AddSecondaryKeys("key2", "group1", "secKey3")

	mm.DeleteBySecondaryKey("group1", "secKey2")
	assert.False(t, mm.Has("key1"))
	assert.True(t, mm.Has("key2"))
	assert.Equal(t, map[string]map[string]string{"group1": {"secKey3": "key2"}}, mm.KeyGroups())
}

func TestMultiKeyMap_DeleteBySecondaryKeyMissingIsNoOp(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.DeleteBySecondaryKey("group1", "secKey1")
	mm.Set("key1", 1)
	mm.AddSecondaryKeys("key1", "group1", "secKey1")
	mm.DeleteBySecondaryKey("group1", "missing")
	mm.DeleteBySecondaryKey("missing", "secKey1")
	assert.True(t, mm.Has("key1"))
	assert.True(t, mm.HasSecondaryKey("group1", "secKey1"))
}

func TestMultiKeyMap_Clone(t *testing.T) {
	mm := New[string, int]()
	mm.Set("key1", 1)
	mm.Set("key2", 2)
	mm.AddSecondaryKeys("key1", "group1", "secKey1", "secKey2")
	mm.AddSecondaryKeys("key2", "group2", "secKey3")

	clone := mm.Clone()
	assert.Equal(t, maps.Collect(mm.All()), maps.Collect(clone.All()))
	assert.Equal(t, mm.KeyGroups(), clone.KeyGroups())

	// Changes to the clone and to the original stay apart.
	clone.Set("key1", 10)
	clone.Set("key3", 3)
	clone.AddSecondaryKeys("key3", "group1", "secKey4")
	clone.DeleteSecondaryKey("group1", "secKey1")
	clone.DeleteGroup("group2")
	mm.AddSecondaryKeys("key2", "group2", "secKey5")

	assert.Equal(t, map[string]int{"key1": 1, "key2": 2}, maps.Collect(mm.All()))
	assert.Equal(t, map[string]map[string]string{
		"group1": {"secKey1": "key1", "secKey2": "key1"},
		"group2": {"secKey3": "key2", "secKey5": "key2"},
	}, mm.KeyGroups())
	assert.Equal(t, map[string]int{"key1": 10, "key2": 2, "key3": 3}, maps.Collect(clone.All()))
	assert.Equal(t, map[string]map[string]string{
		"group1": {"secKey2": "key1", "secKey4": "key3"},
	}, clone.KeyGroups())

	// The reverse index of the original is not shared either.
	mm.Delete("key1")
	assert.Equal(t, map[string]string{"group1": "secKey2"}, maps.Collect(clone.SecondaryKeys("key1")))
	assert.Empty(t, maps.Collect(clone.SecondaryKeys("key2")))
}

func TestMultiKeyMap_CloneCopiesValuesShallowly(t *testing.T) {
	mm := New[string, *int]()
	value := 1
	mm.Set("key1", &value)
	clone := mm.Clone()
	got, _ := clone.Get("key1")
	assert.Same(t, &value, got)
}

func TestMultiKeyMap_CloneZeroValue(t *testing.T) {
	var mm MultiKeyMap[string, int]
	clone := mm.Clone()
	assert.Equal(t, 0, clone.Len())
	clone.Set("key1", 1)
	assert.True(t, clone.AddSecondaryKeys("key1", "group1", "secKey1"))
	assert.Equal(t, 0, mm.Len())
	assert.Empty(t, mm.KeyGroups())
}

func TestMultiKeyMap_DeleteFunc(t *testing.T) {
	mm := New[string, int]()
	for i := range 4 {
		key := "key" + strconv.Itoa(i)
		mm.Set(key, i)
		mm.AddSecondaryKeys(key, "group1", "secKey"+strconv.Itoa(i))
	}
	mm.DeleteFunc(func(key string, value int) bool { return value%2 == 1 })
	assert.Equal(t, map[string]int{"key0": 0, "key2": 2}, maps.Collect(mm.All()))
	assert.Equal(t, map[string]map[string]string{
		"group1": {"secKey0": "key0", "secKey2": "key2"},
	}, mm.KeyGroups())
	assert.NotContains(t, mm.secondaryTo, "key1")
	assert.NotContains(t, mm.secondaryTo, "key3")
}

func TestMultiKeyMap_DeleteFuncZeroValue(t *testing.T) {
	var mm MultiKeyMap[string, int]
	mm.DeleteFunc(func(string, int) bool { return true })
	assert.Equal(t, 0, mm.Len())
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

func BenchmarkMultiKeyMapGet(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int]()
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

func BenchmarkMultiKeyMapSet(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.Set(strconv.Itoa(n), n)
				}
			}
		})
	}
}

func BenchmarkMultiKeyMapDelete(b *testing.B) {
	for _, v := range benchmarkSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := New[string, int]()
			for b.Loop() {
				// Refill outside the timer so that every iteration deletes existing keys
				// and their secondary keys.
				b.StopTimer()
				for n := range v.size {
					key := strconv.Itoa(n)
					m.Set(key, n)
					m.AddSecondaryKeys(key, "group", key)
				}
				b.StartTimer()
				for n := range v.size {
					m.Delete(strconv.Itoa(n))
				}
			}
		})
	}
}
