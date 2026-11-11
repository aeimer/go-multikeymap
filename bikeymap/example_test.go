package bikeymap_test

import (
	"fmt"
	"maps"
	"slices"

	"github.com/aeimer/go-multikeymap/bikeymap"
)

func Example_zeroValue() {
	// The zero value is ready to use; New is not required.
	var m bikeymap.BiKeyMap[string, int, string]
	m.Set("berlin", 10115, "Berlin")

	name, ok := m.GetByKeyB(10115)
	fmt.Println(name, ok)

	// The same holds for the concurrent variant.
	var c bikeymap.ConcurrentBiKeyMap[string, int, string]
	c.Set("hamburg", 20095, "Hamburg")
	fmt.Println(c.Len())

	// Output:
	// Berlin true
	// 1
}

func ExampleBiKeyMap_All() {
	m := bikeymap.New[string, int, string]()
	m.Set("berlin", 10115, "Berlin")
	m.Set("hamburg", 20095, "Hamburg")

	// The iteration order is unspecified; collect and sort for stable output.
	names := maps.Collect(m.All())
	for _, keyA := range slices.Sorted(maps.Keys(names)) {
		fmt.Println(keyA, names[keyA])
	}

	// Output:
	// berlin Berlin
	// hamburg Hamburg
}

func ExampleBiKeyMap_KeyA() {
	m := bikeymap.New[string, int, string]()
	m.Set("berlin", 10115, "Berlin")

	keyA, ok := m.KeyA(10115)
	fmt.Println(keyA, ok)

	// Output:
	// berlin true
}

func ExampleBiKeyMap_KeyB() {
	m := bikeymap.New[string, int, string]()
	m.Set("berlin", 10115, "Berlin")

	keyB, ok := m.KeyB("berlin")
	fmt.Println(keyB, ok)

	_, ok = m.KeyB("hamburg")
	fmt.Println(ok)

	// Output:
	// 10115 true
	// false
}

func ExampleBiKeyMap_Set() {
	m := bikeymap.New[string, int, string]()
	m.Set("berlin", 10115, "Berlin")
	m.Set("hamburg", 20095, "Hamburg")

	// 10115 is paired with "berlin": Set removes "berlin".
	m.Set("munich", 10115, "Munich")
	fmt.Println(m.Len())
	_, ok := m.GetByKeyA("berlin")
	fmt.Println(ok)

	// "munich" is paired with 10115 and 20095 with "hamburg":
	// Set removes 10115 and "hamburg", so Len shrinks by one.
	m.Set("munich", 20095, "Munich")
	fmt.Println(m.Len())
	keyA, _ := m.KeyA(20095)
	fmt.Println(keyA)

	// Output:
	// 2
	// false
	// 1
	// munich
}

func ExampleBiKeyMap_Clone() {
	m := bikeymap.New[string, int, string]()
	m.Set("berlin", 10115, "capital")

	c := m.Clone()
	c.Set("berlin", 10115, "city")
	c.DeleteByKeyB(10115)

	value, ok := m.GetByKeyB(10115)
	fmt.Println(value, ok)
	fmt.Println(c.Len())

	// Output:
	// capital true
	// 0
}
