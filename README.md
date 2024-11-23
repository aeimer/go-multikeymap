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

Benchmark results (`task gotb`):

```
BenchmarkMultiKeyMapGet100-12          	  770486	      1551 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapGet1000-12         	   32757	     36681 ns/op	    2880 B/op	     900 allocs/op
BenchmarkMultiKeyMapGet10000-12        	    2816	    434499 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkMultiKeyMapGet100000-12       	     241	   4755312 ns/op	  518884 B/op	   99900 allocs/op

BenchmarkMultiKeyMapPut100-12          	  450188	      2608 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapPut1000-12         	   30291	     39424 ns/op	    2884 B/op	     900 allocs/op
BenchmarkMultiKeyMapPut10000-12        	    2382	    455831 ns/op	   39291 B/op	    9900 allocs/op
BenchmarkMultiKeyMapPut100000-12       	     220	   5090152 ns/op	  555726 B/op	   99918 allocs/op

BenchmarkMultiKeyMapRemove100-12       	  507541	      2374 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapRemove1000-12      	   42507	     28215 ns/op	    2880 B/op	     900 allocs/op
BenchmarkMultiKeyMapRemove10000-12     	    4044	    289343 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkMultiKeyMapRemove100000-12    	     414	   2850928 ns/op	  518897 B/op	   99900 allocs/op
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

Benchmark results (`task gotb`):

```
BenchmarkBiKeyMapGet100-12          	  348357	      3399 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapGet1000-12         	   20005	     58957 ns/op	    2880 B/op	     900 allocs/op
BenchmarkBiKeyMapGet10000-12        	    1480	    816056 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkBiKeyMapGet100000-12       	     128	   9118454 ns/op	  518881 B/op	   99900 allocs/op

BenchmarkBiKeyMapPut100-12          	  211629	      5452 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapPut1000-12         	   12333	     97718 ns/op	    5792 B/op	    1800 allocs/op
BenchmarkBiKeyMapPut10000-12        	     907	   1313887 ns/op	   81296 B/op	   19800 allocs/op
BenchmarkBiKeyMapPut100000-12       	      78	  20215060 ns/op	 1381083 B/op	  199951 allocs/op

BenchmarkBiKeyMapRemove100-12       	  344187	      3304 ns/op	    1600 B/op	      99 allocs/op
BenchmarkBiKeyMapRemove1000-12      	   36328	     33004 ns/op	   15999 B/op	     999 allocs/op
BenchmarkBiKeyMapRemove10000-12     	    3446	    334289 ns/op	  159953 B/op	    9997 allocs/op
BenchmarkBiKeyMapRemove100000-12    	     351	   3368186 ns/op	 1595457 B/op	   99715 allocs/op
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
