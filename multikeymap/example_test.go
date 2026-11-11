package multikeymap_test

import (
	"fmt"
	"maps"
	"slices"

	"github.com/aeimer/go-multikeymap/multikeymap"
)

func Example_zeroValue() {
	// The zero value is ready to use; New is not required.
	var m multikeymap.MultiKeyMap[string, int]
	m.Set("berlin", 3_500_000)
	m.AddSecondaryKeys("berlin", "postcode", "10115")

	population, ok := m.GetBySecondaryKey("postcode", "10115")
	fmt.Println(population, ok)

	// The same holds for the concurrent variant.
	var c multikeymap.ConcurrentMultiKeyMap[string, int]
	c.Set("hamburg", 1_800_000)
	fmt.Println(c.Len())

	// Output:
	// 3500000 true
	// 1
}

func ExampleMultiKeyMap_All() {
	m := multikeymap.New[string, int]()
	m.Set("berlin", 3_500_000)
	m.Set("hamburg", 1_800_000)

	// The iteration order is unspecified; collect and sort for stable output.
	populations := maps.Collect(m.All())
	for _, city := range slices.Sorted(maps.Keys(populations)) {
		fmt.Println(city, populations[city])
	}

	// Output:
	// berlin 3500000
	// hamburg 1800000
}

func ExampleMultiKeyMap_PrimaryKey() {
	m := multikeymap.New[string, int]()
	m.Set("berlin", 3_500_000)
	m.AddSecondaryKeys("berlin", "postcode", "10115", "10117")

	city, ok := m.PrimaryKey("postcode", "10117")
	fmt.Println(city, ok)

	_, ok = m.PrimaryKey("postcode", "20095")
	fmt.Println(ok)

	// Output:
	// berlin true
	// false
}

func ExampleMultiKeyMap_SecondaryKeys() {
	m := multikeymap.New[string, int]()
	m.Set("berlin", 3_500_000)
	m.AddSecondaryKeys("berlin", "postcode", "10115")
	m.AddSecondaryKeys("berlin", "plate", "B")

	var keys []string
	for group, key := range m.SecondaryKeys("berlin") {
		keys = append(keys, group+"="+key)
	}
	slices.Sort(keys)
	fmt.Println(keys)

	// Output:
	// [plate=B postcode=10115]
}

func ExampleConcurrentMultiKeyMap_All() {
	var m multikeymap.ConcurrentMultiKeyMap[string, int]
	m.Set("berlin", 3_500_000)

	// The loop sees a snapshot and holds no lock, so it may write to the map.
	for city, population := range m.All() {
		m.Set(city+"-metro", population*2)
	}
	fmt.Println(m.Len())

	// Output:
	// 2
}

func ExampleMultiKeyMap_LoadOrStore() {
	var m multikeymap.MultiKeyMap[string, int]

	population, loaded := m.LoadOrStore("berlin", 3_500_000)
	fmt.Println(population, loaded)

	// The second call keeps the stored value.
	population, loaded = m.LoadOrStore("berlin", 0)
	fmt.Println(population, loaded)

	// Output:
	// 3500000 false
	// 3500000 true
}

func ExampleMultiKeyMap_Update() {
	var visits multikeymap.MultiKeyMap[string, int]
	visits.Set("berlin", 0)
	visits.AddSecondaryKeys("berlin", "postcode", "10115")

	// Count visits; a missing key starts at 0.
	for _, city := range []string{"berlin", "hamburg", "berlin"} {
		visits.Update(city, func(old int, _ bool) (int, bool) {
			return old + 1, true
		})
	}
	count, _ := visits.GetBySecondaryKey("postcode", "10115")
	fmt.Println(count)

	// Returning false deletes the entry.
	visits.Update("hamburg", func(int, bool) (int, bool) { return 0, false })
	fmt.Println(visits.Has("hamburg"))

	// Output:
	// 2
	// false
}

func ExampleMultiKeyMap_DeleteBySecondaryKey() {
	m := multikeymap.New[string, int]()
	m.Set("berlin", 3_500_000)
	m.AddSecondaryKeys("berlin", "postcode", "10115")
	m.AddSecondaryKeys("berlin", "plate", "B")

	// Deletes berlin with all its secondary keys.
	m.DeleteBySecondaryKey("plate", "B")
	fmt.Println(m.Has("berlin"), m.HasSecondaryKey("postcode", "10115"))

	// Output:
	// false false
}

func ExampleMultiKeyMap_LoadAndDeleteBySecondaryKey() {
	m := multikeymap.New[string, int]()
	m.Set("berlin", 3_500_000)
	m.AddSecondaryKeys("berlin", "plate", "B")

	// Deletes berlin with all its secondary keys and returns its value.
	population, loaded := m.LoadAndDeleteBySecondaryKey("plate", "B")
	fmt.Println(population, loaded, m.Len())

	// Output:
	// 3500000 true 0
}

func ExampleMultiKeyMap_Clone() {
	m := multikeymap.New[string, int]()
	m.Set("berlin", 3_500_000)
	m.AddSecondaryKeys("berlin", "postcode", "10115")

	c := m.Clone()
	c.Set("berlin", 3_600_000)
	c.DeleteSecondaryKey("postcode", "10115")

	population, _ := m.Get("berlin")
	fmt.Println(population, m.HasSecondaryKey("postcode", "10115"))
	population, _ = c.Get("berlin")
	fmt.Println(population, c.HasSecondaryKey("postcode", "10115"))

	// Output:
	// 3500000 true
	// 3600000 false
}
