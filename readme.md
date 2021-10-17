# BigMap
[![GoReport](https://goreportcard.com/badge/github.com/worldOneo/bigmap)](https://goreportcard.com/report/github.com/worldOneo/bigmap)  
## Fast - Scaling - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Fast

The average operation is done in **under 0.1μs** and can therefore be done over **10 Million times / second**.  
And all this **per thread**. This is achieved by storing the objects in one single byte-slice and having a Zero-Allocation, Share-Nothing oriented design.  
Resulting in **minimimal GC pressure** and **maximal performance**.

## Concurrent

The map has **no global lock**.  
It is split into **multiple shards** which are locked individual.  
As the benchmarks show bigmap **gains from concurrent access**.  
With preallocations and items having a max size it is significant **faster than the standard map**.

## Scaling

If you have more concurrent accesses, you can always increase the shard count.  
As always: only benchmarking **your usecase** will reveal the optimal settings.  
But as shown, with the default 16 shards, you still get a good access speed even with half a million routines.  
Each shard can store gigabytes of data without loosing performance, so it is good for storing tons of tons of normalized data.

## Comparison with std-lib
### Sync
| | Time | diff | GC Pause total | GC Diff |
| --- | --- |--- | --- | --- |
| **BigMap** | 53s | +0% | 4ms | 0% |
| **StdMap** | 2m36s | +294% | 20.2ms | +505% |
| **Syncmap** | 7m24s | +853% | 17ms | +425% |
 
### Async
| | Time | diff |
| --- | --- |--- |
| **BigMap** | 24.9s | +0% |
| **StdMap** (synced with RWMutex) | 3m41s | +887% |
| **Syncmap** | 7m32.9s | +1.875% |

## Benchmarks

The benchmarks are done on a machine with an i7-8750H CPU (12x 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
_I can't do any bigger benchmarks becaus that would mean I needed to populate maps bigger than my memory_
```sh
go version
go version go1.17.2 windows/amd64

go test -benchmem -run=^$ -bench BenchmarkBigMap_.* github.com/worldOneo/bigmap -benchtime 1s

BenchmarkGenKey-12                              20750223               113.5 ns/op            24 B/op          2 allocs/op
BenchmarkFNV64-12                               472440758                5.106 ns/op           0 B/op          0 allocs/op
BenchmarkBigMap_Put-12                           7841419               149.1 ns/op           273 B/op          0 allocs/op
BenchmarkBigMap_Get-12                          41903976                34.56 ns/op            0 B/op          0 allocs/op
BenchmarkBigMap_Delete-12                       26086956                44.28 ns/op           10 B/op          0 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                29894546                40.24 ns/op            0 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12              19918322                66.17 ns/op          111 B/op          0 allocs/op
# Parallel benchmarks have allocations because of the key generation (113.5ns; 2 allocs/op).
# That also slows them down a little but this is required for the parallel test.
BenchmarkBigMap_Put_Parallel-12                 10720135               145.4 ns/op           456 B/op          2 allocs/op
BenchmarkBigMap_Get_Parallel-12                 30789361                45.67 ns/op           39 B/op          2 allocs/op
BenchmarkBigMap_Delete_Parallel-12              18111045                66.87 ns/op           54 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       10256593               108.7 ns/op           107 B/op          2 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     10000141               107.8 ns/op           111 B/op          2 allocs/op
```

## Attention
The map scales as more data is added but, to enable high performance, doesn't schrink.
To enable the fast accessess free heap is held "hot" to be ready to use.
This means the map might grow once realy big, which might seeme like a memory leak at first glance because it doesn shrink, but then never grows again.
