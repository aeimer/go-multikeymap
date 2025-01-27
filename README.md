# multikeymap

See docs: https://pkg.go.dev/github.com/aeimer/go-multikeymap

DISCLAIMER: Until version 1 is reached, the API may change.

A go lib which handles maps with multiple keys.
Both data-structures are available in go routine safe (concurrent) and non-concurrent version.

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
# => Concurrent <=
BenchmarkConcurrentMultiKeyMapGet100-12          	  785500	      1474 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMultiKeyMapGet1000-12         	   34244	     34785 ns/op	    2880 B/op	     900 allocs/op
BenchmarkConcurrentMultiKeyMapGet10000-12        	    2864	    408021 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkConcurrentMultiKeyMapGet100000-12       	     260	   4617505 ns/op	  518883 B/op	   99900 allocs/op

BenchmarkConcurrentMultiKeyMapPut100-12          	  478063	      2462 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMultiKeyMapPut1000-12         	   30484	     38430 ns/op	    2884 B/op	     900 allocs/op
BenchmarkConcurrentMultiKeyMapPut10000-12        	    2683	    437981 ns/op	   39243 B/op	    9900 allocs/op
BenchmarkConcurrentMultiKeyMapPut100000-12       	     247	   4853266 ns/op	  551621 B/op	   99916 allocs/op

BenchmarkConcurrentMultiKeyMapRemove100-12       	  507627	      2330 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentMultiKeyMapRemove1000-12      	   43503	     27669 ns/op	    2880 B/op	     900 allocs/op
BenchmarkConcurrentMultiKeyMapRemove10000-12     	    4153	    279454 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkConcurrentMultiKeyMapRemove100000-12    	     435	   2789207 ns/op	  518896 B/op	   99900 allocs/op

# => Non-Concurrent <=
BenchmarkMultiKeyMapGet100-12                    	 1462314	       791.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapGet1000-12                   	   38054	     31372 ns/op	    2880 B/op	     900 allocs/op
BenchmarkMultiKeyMapGet10000-12                  	    3162	    371122 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkMultiKeyMapGet100000-12                 	     278	   4291932 ns/op	  518882 B/op	   99900 allocs/op

BenchmarkMultiKeyMapPut100-12                    	  912810	      1324 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapPut1000-12                   	   37822	     31526 ns/op	    2883 B/op	     900 allocs/op
BenchmarkMultiKeyMapPut10000-12                  	    3148	    372532 ns/op	   39189 B/op	    9900 allocs/op
BenchmarkMultiKeyMapPut100000-12                 	     285	   4176945 ns/op	  547220 B/op	   99913 allocs/op

BenchmarkMultiKeyMapRemove100-12                 	 2957056	       404.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiKeyMapRemove1000-12                	   81163	     14679 ns/op	    2880 B/op	     900 allocs/op
BenchmarkMultiKeyMapRemove10000-12               	    7063	    163110 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkMultiKeyMapRemove100000-12              	     702	   1674200 ns/op	  518892 B/op	   99900 allocs/op
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
# => Concurrent <=
BenchmarkBiKeyMapGet100-12                    	  596592	      2038 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapGet1000-12                   	   25460	     45825 ns/op	    2880 B/op	     900 allocs/op
BenchmarkBiKeyMapGet10000-12                  	    1664	    692854 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkBiKeyMapGet100000-12                 	     148	   8296314 ns/op	  518881 B/op	   99900 allocs/op

BenchmarkBiKeyMapPut100-12                    	  259808	      4586 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapPut1000-12                   	   13803	     93115 ns/op	    5789 B/op	    1800 allocs/op
BenchmarkBiKeyMapPut10000-12                  	     931	   1233748 ns/op	   81198 B/op	   19800 allocs/op
BenchmarkBiKeyMapPut100000-12                 	      84	  16569381 ns/op	 1356669 B/op	  199940 allocs/op

BenchmarkBiKeyMapRemove100-12                 	 6005941	       198.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapRemove1000-12                	  626977	      1897 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapRemove10000-12               	   63441	     18883 ns/op	       0 B/op	       0 allocs/op
BenchmarkBiKeyMapRemove100000-12              	    6264	    190622 ns/op	       0 B/op	       0 allocs/op

# => Non-Concurrent <=
BenchmarkConcurrentBiKeyMapGet100-12          	  364874	      3497 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentBiKeyMapGet1000-12         	   20120	     58907 ns/op	    2880 B/op	     900 allocs/op
BenchmarkConcurrentBiKeyMapGet10000-12        	    1513	    786043 ns/op	   38880 B/op	    9900 allocs/op
BenchmarkConcurrentBiKeyMapGet100000-12       	     133	   8940005 ns/op	  518881 B/op	   99900 allocs/op

BenchmarkConcurrentBiKeyMapPut100-12          	  211080	      5466 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentBiKeyMapPut1000-12         	   12590	     93671 ns/op	    5792 B/op	    1800 allocs/op
BenchmarkConcurrentBiKeyMapPut10000-12        	     932	   1278751 ns/op	   81198 B/op	   19800 allocs/op
BenchmarkConcurrentBiKeyMapPut100000-12       	      79	  14791849 ns/op	 1376440 B/op	  199947 allocs/op

BenchmarkConcurrentBiKeyMapRemove100-12       	  359648	      3230 ns/op	    1600 B/op	      99 allocs/op
BenchmarkConcurrentBiKeyMapRemove1000-12      	   36781	     32997 ns/op	   15999 B/op	     999 allocs/op
BenchmarkConcurrentBiKeyMapRemove10000-12     	    3426	    334445 ns/op	  159953 B/op	    9997 allocs/op
BenchmarkConcurrentBiKeyMapRemove100000-12    	     357	   3275911 ns/op	 1595519 B/op	   99719 allocs/op
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
