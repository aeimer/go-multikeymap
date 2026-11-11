package bikeymap_test

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/aeimer/go-multikeymap/bikeymap"
)

func Example_zeroValue() {
	// The zero value is ready to use; New is not required.
	var m bikeymap.BiKeyMap[string, int, string]
	_ = m.Set("berlin", 10115, "Berlin")

	name, ok := m.GetByKeyB(10115)
	fmt.Println(name, ok)

	// The same holds for the concurrent variant.
	var c bikeymap.ConcurrentBiKeyMap[string, int, string]
	_ = c.Set("hamburg", 20095, "Hamburg")
	fmt.Println(c.Len())

	// Output:
	// Berlin true
	// 1
}

func ExampleBiKeyMap_All() {
	m := bikeymap.New[string, int, string]()
	_ = m.Set("berlin", 10115, "Berlin")
	_ = m.Set("hamburg", 20095, "Hamburg")

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
	_ = m.Set("berlin", 10115, "Berlin")

	keyA, ok := m.KeyA(10115)
	fmt.Println(keyA, ok)

	// Output:
	// berlin true
}

func ExampleBiKeyMap_KeyB() {
	m := bikeymap.New[string, int, string]()
	_ = m.Set("berlin", 10115, "Berlin")

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
	_ = m.Set("berlin", 10115, "Berlin")

	// 10115 is already paired with "berlin".
	err := m.Set("hamburg", 10115, "Hamburg")
	fmt.Println(errors.Is(err, bikeymap.ErrKeyBConflict))

	// Output:
	// true
}

func ExampleBiKeyMap_Clone() {
	m := bikeymap.New[string, int, string]()
	_ = m.Set("berlin", 10115, "capital")

	c := m.Clone()
	_ = c.Set("berlin", 10115, "city")
	c.DeleteByKeyB(10115)

	value, ok := m.GetByKeyB(10115)
	fmt.Println(value, ok)
	fmt.Println(c.Len())

	// Output:
	// capital true
	// 0
}
