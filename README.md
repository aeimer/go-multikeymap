# go-multikeymap

A go lib which handles maps with multiple keys.

* **MuliKeyMap** is a data structure based on go native maps.
It has a primary key which directly maps to the values.
The secondary keys are mapping to the primary key.
Therefore, the memory consumption is a bit higher than a native map.
The access is O(1+1+1) => O(1) due to the underlying hashmap.

* **BiKeyMap** is a stricter version of MultiKeyMap.
It has KeyA and KeyB, both need to be unique.
The access is O(1+1) => O(1) due to the underlying hashmap.
