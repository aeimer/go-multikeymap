# multikeymap

See docs: https://pkg.go.dev/github.com/aeimer/go-multikeymap

DISCLAIMER: Until version 1 is reached, the API may change.

A go lib which handles maps with multiple keys.
Both data-structures are available in go routine safe (concurrent) and a non-concurrent version.

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
	mm := multikeymap.New[string, City]()
	// or: mm := multikeymap.NewConcurrent[string, City]()
	mm.Put("Berlin", City{"Berlin", 3_500_000})
	mm.PutSecondaryKeys("Berlin", "postcode", "10115", "10117", "10119")
	mm.Get("Berlin")                          // City{"Berlin", 3_500_000}
	mm.GetBySecondaryKey("postcode", "10115") // City{"Berlin", 3_500_000}
}
```

Benchmark results (`task gotb`):

```
# => Non-Concurrent <=
BenchmarkMultiKeyMapGet/size_100-12                   	 1430344	       824.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapGet/size_1000-12                  	   36643	     33912 ns/op	    2880 B/op	     900 allocs/op
BenchmarkMultiKeyMapGet/size_10000-12                 	    3036	    390264 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkMultiKeyMapGet/size_100000-12                	     261	   4601110 ns/op	  518882 B/op	   99900 allocs/op

BenchmarkMultiKeyMapPut/size_100-12                   	  888855	      1412 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapPut/size_1000-12                  	   36153	     33815 ns/op	    2883 B/op	     900 allocs/op
BenchmarkMultiKeyMapPut/size_10000-12                 	    3026	    394762 ns/op	   39202 B/op	    9900 allocs/op
BenchmarkMultiKeyMapPut/size_100000-12                	     248	   4571614 ns/op	  551453 B/op	   99915 allocs/op

BenchmarkMultiKeyMapRemove/size_100-12                	 2821974	       425.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapRemove/size_1000-12               	   78538	     15281 ns/op	    2880 B/op	     900 allocs/op
BenchmarkMultiKeyMapRemove/size_10000-12              	    6922	    169382 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkMultiKeyMapRemove/size_100000-12             	     670	   1764847 ns/op	  518884 B/op	   99900 allocs/op

# => Concurrent <=
BenchmarkConcurrentMultiKeyMapGet/size_100-12         	  735793	      1578 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMultiKeyMapGet/size_1000-12        	   32385	     37066 ns/op	    2880 B/op	     900 allocs/op
BenchmarkConcurrentMultiKeyMapGet/size_10000-12       	    2756	    433572 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkConcurrentMultiKeyMapGet/size_100000-12      	     240	   4950545 ns/op	  518883 B/op	   99900 allocs/op

BenchmarkConcurrentMultiKeyMapPut/size_100-12         	  454376	      2609 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMultiKeyMapPut/size_1000-12        	   29721	     40303 ns/op	    2884 B/op	     900 allocs/op
BenchmarkConcurrentMultiKeyMapPut/size_10000-12       	    2592	    459977 ns/op	   39256 B/op	    9900 allocs/op
BenchmarkConcurrentMultiKeyMapPut/size_100000-12      	     224	   5369616 ns/op	  555038 B/op	   99918 allocs/op

BenchmarkConcurrentMultiKeyMapRemove/size_100-12      	  490444	      2412 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMultiKeyMapRemove/size_1000-12     	   41236	     29073 ns/op	    2880 B/op	     900 allocs/op
BenchmarkConcurrentMultiKeyMapRemove/size_10000-12    	    4030	    298559 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkConcurrentMultiKeyMapRemove/size_100000-12   	     400	   2968791 ns/op	  518884 B/op	   99900 allocs/op
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
	bm := bikeymap.New[string, int, City]()
	// or: bm := bikeymap.NewConcurrent[string, int, City]()
	bm.Put("Berlin", 3_500_000, City{"Berlin", 3_500_000})
	bm.Put("Hamburg", 1_800_000, City{"Hamburg", 1_800_000})
	bm.GetByKeyA("Berlin")  // City{"Berlin", 3_500_000}
	bm.GetByKeyB(1_800_000) // City{"Hamburg", 1_800_000}
}
```

Benchmark results (`task gotb`):

```
# => Non-Concurrent <=
BenchmarkBiKeyMapGet/size_100-12        	  531668	      2094 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapGet/size_1000-12       	   24627	     49464 ns/op	    2880 B/op	     900 allocs/op
BenchmarkBiKeyMapGet/size_10000-12      	    1644	    729522 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkBiKeyMapGet/size_100000-12     	     139	   8544642 ns/op	  518882 B/op	   99900 allocs/op

BenchmarkBiKeyMapPut/size_100-12        	  254956	      4653 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapPut/size_1000-12       	   12880	     92922 ns/op	    5791 B/op	    1800 allocs/op
BenchmarkBiKeyMapPut/size_10000-12      	     914	   1289536 ns/op	   81248 B/op	   19800 allocs/op
BenchmarkBiKeyMapPut/size_100000-12     	      67	  17404694 ns/op	 1436872 B/op	  199973 allocs/op

BenchmarkBiKeyMapRemove/size_100-12     	 5802380	       206.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapRemove/size_1000-12    	  612367	      1984 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapRemove/size_10000-12   	   60577	     19730 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapRemove/size_100000-12  	    5990	    199194 ns/op	       0 B/op	       0 allocs/op

# => Concurrent <=
BenchmarkConcurrentBiKeyMapGet/size_100-12         	  325862	      3638 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentBiKeyMapGet/size_1000-12        	   19471	     61879 ns/op	    2880 B/op	     900 allocs/op
BenchmarkConcurrentBiKeyMapGet/size_10000-12       	    1428	    832351 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkConcurrentBiKeyMapGet/size_100000-12      	     100	  10026768 ns/op	  518882 B/op	   99900 allocs/op

BenchmarkConcurrentBiKeyMapPut/size_100-12         	  207386	      5683 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentBiKeyMapPut/size_1000-12        	   12157	     99885 ns/op	    5793 B/op	    1800 allocs/op
BenchmarkConcurrentBiKeyMapPut/size_10000-12       	     870	   1371878 ns/op	   81438 B/op	   19800 allocs/op
BenchmarkConcurrentBiKeyMapPut/size_100000-12      	      62	  19618138 ns/op	 1469608 B/op	  199989 allocs/op

BenchmarkConcurrentBiKeyMapRemove/size_100-12      	  353018	      3667 ns/op	    1599 B/op	      99 allocs/op
BenchmarkConcurrentBiKeyMapRemove/size_1000-12     	   34488	     34653 ns/op	   15999 B/op	     999 allocs/op
BenchmarkConcurrentBiKeyMapRemove/size_10000-12    	    3363	    364001 ns/op	  159952 B/op	    9997 allocs/op
BenchmarkConcurrentBiKeyMapRemove/size_100000-12   	     320	   3551128 ns/op	 1595034 B/op	   99687 allocs/op
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
