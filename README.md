# multikeymap

See docs: https://pkg.go.dev/github.com/aeimer/go-multikeymap

DISCLAIMER: Until version 1 is reached, the API may change.

A go lib which handles maps with multiple keys.
Both data-structures are go routine safe.

* **MultiKeyMap** is a data structure based on go native maps.
It has a primary key which directly maps to the values.
The secondary keys are mapping to the primary key.
Therefore, the memory consumption is a bit higher than a native map.
The access is O(1+1+1) => O(1) due to the underlying hashmap.

* **BiKeyMap** is a stricter version of MultiKeyMap.
It has KeyA and KeyB, both need to be unique.
The access is O(1+1) => O(1) due to the underlying hashmap.

## MultiKeyMap

This map has a generic primary key and multiple string secondary keys.
You can use it like this:

```go
package main

import "github.com/aeimer/go-multikeymap/multikeymap"

func main() {
	type City struct {
		Name       string
		Population int
	}
	mm := multikeymap.NewMultiKeyMap[string, City]()
	mm.Put("Berlin", City{"Berlin", 3_500_000})
	mm.PutSecondaryKeys("Berlin", "postcode", "10115", "10117", "10119")
	mm.Get("Berlin")                          // City{"Berlin", 3_500_000}
	mm.GetBySecondaryKey("postcode", "10115") // City{"Berlin", 3_500_000}
}
```

# BiKeyMap

This map has two generic keys, both need to be unique.
You can use it like this:

```go
package main

import "github.com/aeimer/go-multikeymap/bikeymap"

func main() {
	type City struct {
		Name       string
		Population int
	}
	// keyA: Cityname, keyB: Population
	bm := bikeymap.NewBiKeyMap[string, int, City]()
	bm.Put("Berlin", 3_500_000, City{"Berlin", 3_500_000})
	bm.Put("Hamburg", 1_800_000, City{"Hamburg", 1_800_000})
	bm.GetByKeyA("Berlin")  // City{"Berlin", 3_500_000}
	bm.GetByKeyB(1_800_000) // City{"Hamburg", 1_800_000}
}
```

## Contribution

Feel free to contribute by opening issues or pull requests.
To set up the project, you need to have go installed.
Then you can run the following commands:

```bash
# List Taskfile tasks
task

# Install dependencies
task tools

# Run tests
task go-test

# Run benchmarks
task go-test-bench
```

## Star history

[![Star History](https://api.star-history.com/svg?repos=aeimer/go-multikeymap&type=Date)](https://star-history.com/#aeimer/go-multikeymap&Date)
