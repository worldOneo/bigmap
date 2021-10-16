# BigMap
[![GoReport](https://goreportcard.com/badge/github.com/worldOneo/bigmap)](https://goreportcard.com/report/github.com/worldOneo/bigmap)  
## Fast - Scaling - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Fast

Most operations are done in **under 0.3μs** and can therefore be done over **3 Million times / second**.  
And all this **per thread**. This is achieved by storing the objects in one single byte-slice and having a Zero-Allocation, Share-Nothing oriented design.  
Resulting in **minimimal GC pressure**.

## Concurrent

The map has **no global lock**.  
It is split into **multiple shards** which are locked individual.  
As the benchmarks show bigmap **gains from concurrent access**.  
With preallocations and items having a max size it is **faster than the standard map**.

## Scaling

If you have more concurrent accesses, you can always increase the shard count.  
As always: only benchmarking **your usecase** will reveal the optimal settings.  
But as shown, with the default 16 shards, you still get a good access speed even with half a million routines.  
Each shard can store gigabytes of data without loosing performance, so it is good for storing tons of tons of normalized data.

## Comparison with std-lib
### Sync
| | Time | diff | GC Pause total | GC Diff |
| --- | --- |--- | --- | --- |
| **BigMap** | 2m6s | 0% | 8ms | 0% |
| **StdMap** | 3m31s | +67% | 18.6ms | +132% |
| **Syncmap** | 9m54s | +371% | 18.3ms | +128% |
 
### Async
| | Time | diff |
| --- | --- |--- |
| **BigMap** | 38.1s | 0% |
| **StdMap** (synced with RWMutex) | 4m32s | +456% |
| **Syncmap** | 10m35.6s | +1.568% |

## Benchmarks

The benchmarks are done on a machine with an i7-8750H CPU (12x 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
```sh
go version
go version go1.17.2 windows/amd64

go test -benchmem -run=^$ -bench .* github.com/worldOneo/bigmap -benchtime=2s

BenchmarkGenKey-12                              20750223               113.5 ns/op            24 B/op          2 allocs/op
BenchmarkFNV64-12                               472440758                5.106 ns/op           0 B/op          0 allocs/op
BenchmarkBigMap_Put-12                           7318198               350.4 ns/op           368 B/op          0 allocs/op
BenchmarkBigMap_Get-12                          11111563               233.0 ns/op           112 B/op          1 allocs/op
BenchmarkBigMap_Delete-12                       16521562               168.9 ns/op             8 B/op          0 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                40484532                59.73 ns/op           37 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12              13838244               207.8 ns/op           139 B/op          0 allocs/op
BenchmarkShard_Put-12                            8765626               290.1 ns/op           308 B/op          0 allocs/op
BenchmarkShard_Put_Stretched-12                 10666585               249.1 ns/op           228 B/op          0 allocs/op
BenchmarkShard_Get-12                           17142880               160.7 ns/op           112 B/op          1 allocs/op
BenchmarkShard_Delete-12                        20168134               155.1 ns/op            13 B/op          0 allocs/op
BenchmarkShard_Mix_Ballanced-12                 49782512                46.73 ns/op           37 B/op          0 allocs/op
BenchmarkShard_Mix_Unballanced-12               17025538               184.8 ns/op           184 B/op          0 allocs/op
# Parallel benchmarks have allocations because of the key generation (113.5ns; 2 allocs/op).
# That also slows them down a little but this is required for the parallel test.
BenchmarkBigMap_Put_Parallel-12                 14182803               175.8 ns/op           436 B/op          3 allocs/op
BenchmarkBigMap_Get_Parallel-12                 35819558                71.89 ns/op          115 B/op          3 allocs/op
BenchmarkBigMap_Delete_Parallel-12              29999961                84.62 ns/op           48 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       16380764               153.9 ns/op           204 B/op          3 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     13714394               155.8 ns/op           145 B/op          3 allocs/op


go test -benchmem -bench BenchmarkBigMap_Goroutines -benchtime=2s
goos: windows
goarch: amd64
pkg: github.com/worldOneo/bigmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
 === 
Benchmarking with 480000 routines
 ===
BenchmarkBigMap_Goroutines/BigMap_Put_Parallel-12                8461524               290.9 ns/op           374 B/op          2 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Get_Parallel-12               32871289                96.97 ns/op           41 B/op          3 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Delete_Parallel-12            27436130                90.82 ns/op           38 B/op          2 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Mix_Parallel-12               10230573               225.8 ns/op           181 B/op          2 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Mix_Unbalanced_Parallel-12     8242996               281.3 ns/op           200 B/op          2 allocs/op
```

## Attention
The map scales as more data is added but, to enable high performance, doesn't schrink.
To enable the fast accessess free heap is held "hot" to be ready to use.
This means the map might grow once realy big, which might seeme like a memory leak at first glance because it doesn shrink, but then never grows again.
