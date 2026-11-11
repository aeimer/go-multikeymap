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
