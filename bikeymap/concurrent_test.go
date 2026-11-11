package bikeymap

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleNewConcurrent() {
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	value, exists := bm.GetByKeyA("keyA1")
	fmt.Printf("[Key A] value: %v, exists: %v\n", value, exists)
	value, exists = bm.GetByKeyB(1)
	fmt.Printf("[Key B] value: %v, exists: %v\n", value, exists)
	// Output:
	// [Key A] value: value1, exists: true
	// [Key B] value: value1, exists: true
}

func TestConcurrentBiKeyMap_SetAndGet(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()

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

func TestConcurrentBiKeyMap_Set_Overwrite(t *testing.T) {
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
			bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_DeleteByKeyA(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_DeleteByKeyB(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_String(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	expected := "ConcurrentBiKeyMap: map[keyA1:value1]"

	// when / then
	if bm.String() != expected {
		t.Errorf("expected %s, got %s", expected, bm.String())
	}
}

func TestConcurrentBiKeyMap_DeleteByKeyA_NotFound(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")

	// when
	bm.DeleteByKeyA("nonExistentKey")

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestConcurrentBiKeyMap_DeleteByKeyB_NotFound(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")

	// when
	bm.DeleteByKeyB(999)

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestConcurrentBiKeyMap_Len(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()

	// when / then
	assert.Equal(t, 0, bm.Len())

	// when
	bm.Set("keyA1", 1, "value1")

	// then
	assert.Equal(t, 1, bm.Len())
}

func TestConcurrentBiKeyMap_Clear(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when
	bm.Clear()

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_Values(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when / then
	assert.ElementsMatch(t, []string{"value1", "value2"}, slices.Collect(bm.Values()))
}

func TestConcurrentBiKeyMap_ValuesStopsEarly(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_ValuesAllowsWritesInLoop(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")
	keyB := 100

	// when
	for v := range bm.Values() {
		bm.DeleteByKeyA("keyA" + strings.TrimPrefix(v, "value"))
		keyB++
		bm.Set("new"+v, keyB, v)
	}

	// then
	assert.ElementsMatch(t, []string{"value1", "value2"}, slices.Collect(bm.Values()))
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
}

func TestConcurrentBiKeyMap_All(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when / then
	assert.Equal(t, map[string]string{"keyA1": "value1", "keyA2": "value2"}, maps.Collect(bm.All()))
}

func TestConcurrentBiKeyMap_Keys(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when / then
	assert.ElementsMatch(t, []string{"keyA1", "keyA2"}, slices.Collect(bm.KeysA()))
	assert.ElementsMatch(t, []int{1, 2}, slices.Collect(bm.KeysB()))
}

func TestConcurrentBiKeyMap_IteratorsStopEarly(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_IteratorsAllowWritesInLoop(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")

	// when
	requireNoDeadlock(t, func() {
		for keyA, v := range bm.All() {
			bm.DeleteByKeyA(keyA)
			bm.Set("all", 2, v)
		}
		for keyA := range bm.KeysA() {
			bm.DeleteByKeyA(keyA)
			bm.Set("keysA", 3, "value3")
		}
		for keyB := range bm.KeysB() {
			bm.DeleteByKeyB(keyB)
			bm.Set("keysB", 4, "value4")
		}
		for v := range bm.Values() {
			bm.Set("keysB", 4, v+"!")
		}
	})

	// then
	assert.Equal(t, map[string]string{"keysB": "value4!"}, maps.Collect(bm.All()))
}

// requireNoDeadlock fails the test if f does not return within a second.
func requireNoDeadlock(t *testing.T, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deadlock: loop body could not write to the map")
	}
}

func TestConcurrentBiKeyMap_KeyAAndKeyB(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_ConcurrentAccess(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// when / then
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentSet(t, i, &wg, bm)
	}
	wg.Wait()

	// when / then
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentGet(t, i, &wg, bm)
	}
	wg.Wait()

	// when / then
	for i := range numGoroutines {
		wg.Add(1)
		go concurrentDelete(t, i, &wg, bm)
	}
	wg.Wait()
}

func TestConcurrentBiKeyMap_LenDuringSet(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// when
	for i := range numGoroutines {
		wg.Go(func() {
			bm.Set(strconv.Itoa(i), i, "value")
		})
		wg.Go(func() {
			_ = bm.Len()
		})
	}
	wg.Wait()

	// then
	assert.Equal(t, numGoroutines, bm.Len())
}

func concurrentSet(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	value := fmt.Sprintf("value%d", i)
	bm.Set(keyA, keyB, value)
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

func concurrentDelete(t *testing.T, i int, wg *sync.WaitGroup, bm *ConcurrentBiKeyMap[string, int, string]) {
	t.Helper()
	defer wg.Done()
	keyA := fmt.Sprintf("keyA%d", i)
	keyB := i
	bm.DeleteByKeyA(keyA)
	if _, exists := bm.GetByKeyA(keyA); exists {
		t.Errorf("expected keyA%d to be removed", i)
	}
	if _, exists := bm.GetByKeyB(keyB); exists {
		t.Errorf("expected keyB%d to be removed", i)
	}
}

func TestConcurrentBiKeyMap_ZeroValueReads(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]

	// when / then
	_, ok := bm.GetByKeyA("keyA1")
	assert.False(t, ok)
	_, ok = bm.GetByKeyB(1)
	assert.False(t, ok)
	assert.Equal(t, 0, bm.Len())
	assert.Empty(t, slices.Collect(bm.Values()))
	assert.Equal(t, "ConcurrentBiKeyMap: map[]", bm.String())

	// when
	bm.DeleteByKeyA("keyA1")
	bm.DeleteByKeyB(1)
	bm.Clear()

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_ZeroValueWrites(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]

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

func TestConcurrentBiKeyMap_LoadAndDeleteByKeyA(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]

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

func TestConcurrentBiKeyMap_LoadAndDeleteByKeyB(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]

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

func TestConcurrentBiKeyMap_LoadAndDeleteLoadsOnce(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")
	var wg sync.WaitGroup
	var loadedA, loadedB atomic.Int32

	// when
	for range 100 {
		wg.Go(func() {
			if _, loaded := bm.LoadAndDeleteByKeyA("keyA1"); loaded {
				loadedA.Add(1)
			}
		})
		wg.Go(func() {
			if _, loaded := bm.LoadAndDeleteByKeyB(2); loaded {
				loadedB.Add(1)
			}
		})
	}
	wg.Wait()

	// then
	assert.Equal(t, int32(1), loadedA.Load())
	assert.Equal(t, int32(1), loadedB.Load())
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_Clone(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_CloneCopiesValuesShallowly(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, *int]()
	value := 1
	bm.Set("keyA1", 1, &value)

	// when
	clone := bm.Clone()

	// then
	got, _ := clone.GetByKeyA("keyA1")
	assert.Same(t, &value, got)
}

func TestConcurrentBiKeyMap_CloneZeroValue(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]

	// when
	clone := bm.Clone()

	// then
	assert.Equal(t, 0, clone.Len())

	// when
	clone.Set("keyA1", 1, "value1")

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_DeleteFunc(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
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

func TestConcurrentBiKeyMap_DeleteFuncZeroValue(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]

	// when
	bm.DeleteFunc(func(string, int, string) bool { return true })

	// then
	assert.Equal(t, 0, bm.Len())
}

func TestConcurrentBiKeyMap_CloneHasOwnLock(t *testing.T) {
	// given
	var bm ConcurrentBiKeyMap[string, int, string]
	bm.Set("keyA1", 1, "value1")
	clone := bm.Clone()

	// when
	requireNoDeadlock(t, func() {
		bm.DeleteFunc(func(string, int, string) bool {
			clone.Set("keyA2", 2, "value2")
			return false
		})
	})

	// then
	assert.Equal(t, 2, clone.Len())
	assert.Equal(t, 1, bm.Len())
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

func BenchmarkConcurrentBiKeyMapSet(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
			b.ResetTimer()
			for range b.N {
				for n := range v.size {
					m.Set(strconv.Itoa(n), n, strconv.Itoa(n))
				}
			}
		})
	}
}

func BenchmarkConcurrentBiKeyMapDelete(b *testing.B) {
	for _, v := range benchmarkConcurrentSizes {
		b.Run(fmt.Sprintf("size_%d", v.size), func(b *testing.B) {
			m := NewConcurrent[string, int, string]()
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

// reentrantValue calls back into its map from String,
// while a writer waits for the lock.
type reentrantValue struct {
	m *ConcurrentBiKeyMap[string, int, reentrantValue]
}

func (v reentrantValue) String() string {
	go v.m.Set("writer", 2, reentrantValue{})
	// Give the writer time to block on the lock, if String holds it.
	time.Sleep(20 * time.Millisecond)
	return strconv.Itoa(v.m.Len())
}

func TestConcurrentBiKeyMap_String_ValueCallsBack(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, reentrantValue]()
	bm.Set("self", 1, reentrantValue{m: bm})

	// when
	done := make(chan string)
	go func() { done <- bm.String() }()

	// then
	select {
	case s := <-done:
		assert.Contains(t, s, "ConcurrentBiKeyMap: map[self:")
	case <-time.After(2 * time.Second):
		t.Fatal("String deadlocked")
	}
}

func TestConcurrentBiKeyMap_Pairs(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when
	pairs := maps.Collect(bm.Pairs())

	// then
	assert.Equal(t, map[string]int{"keyA1": 1, "keyA2": 2}, pairs)
}

func TestConcurrentBiKeyMap_PairsStopsEarly(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")
	bm.Set("keyA2", 2, "value2")

	// when
	count := 0
	for range bm.Pairs() {
		count++
		break
	}

	// then
	assert.Equal(t, 1, count)
}

func TestConcurrentBiKeyMap_PairsAllowsWritesInLoop(t *testing.T) {
	// given
	bm := NewConcurrent[string, int, string]()
	bm.Set("keyA1", 1, "value1")

	// when
	requireNoDeadlock(t, func() {
		for keyA, keyB := range bm.Pairs() {
			bm.DeleteByKeyA(keyA)
			bm.Set("pairs", keyB+1, "value2")
		}
	})

	// then
	assert.Equal(t, map[string]int{"pairs": 2}, maps.Collect(bm.Pairs()))
}
