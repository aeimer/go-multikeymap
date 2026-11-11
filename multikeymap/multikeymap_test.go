package multikeymap

import (
	"fmt"
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
			for n := range v.size {
				m.Set(strconv.Itoa(n), n)
			}
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.Delete(strconv.Itoa(n))
				}
			}
		})
	}
}
