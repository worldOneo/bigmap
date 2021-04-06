# BigMap
## Scaling - Fast - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Benchmarks
The benchmarks are done on a machine with an i7-8750H CPU (12x 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
```
go version
go version go1.16.2 windows/amd64

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
BenchmarkBigMap_Goroutines/BigMap_Mix_Unbalanced_Parallel-12             8242996               281.3 ns/op           200 B/op          2 allocs/op

go test -benchmem -run=^$ -bench .* github.com/worldOneo/bigmap -benchtime=2s

BenchmarkGenKey-12                              20473971               118.6 ns/op            24 B/op          2 allocs/op
BenchmarkFNV64-12                               19664120               122.4 ns/op            24 B/op          2 allocs/op
BenchmarkShard_Put-12                            5706768               441.8 ns/op           450 B/op          2 allocs/op
BenchmarkShard_Put_Stretched-12                  5874186               437.5 ns/op           232 B/op          2 allocs/op
BenchmarkShard_Get-12                            7291815               435.3 ns/op           135 B/op          2 allocs/op
BenchmarkShard_Delete-12                         8860858               288.0 ns/op            39 B/op          1 allocs/op
BenchmarkShard_Mix_Ballanced-12                 23935660                99.46 ns/op           45 B/op          1 allocs/op
BenchmarkShard_Mix_Unballanced-12                8144883               329.3 ns/op           214 B/op          2 allocs/op
BenchmarkBigMap_Put-12                           5727324               420.0 ns/op           448 B/op          2 allocs/op
BenchmarkBigMap_Get-12                           8583420               303.6 ns/op            39 B/op          1 allocs/op
BenchmarkBigMap_Delete-12                        6568816               385.2 ns/op           135 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                21140475               116.6 ns/op            45 B/op          1 allocs/op
BenchmarkBigMap_Mix_Unballanced-12               7501790               357.3 ns/op           155 B/op          2 allocs/op
BenchmarkBigMap_Put_Parallel-12                 13660968               161.4 ns/op           450 B/op          3 allocs/op
BenchmarkBigMap_Get_Parallel-12                 32206766                82.10 ns/op          123 B/op          3 allocs/op
BenchmarkBigMap_Delete_Parallel-12              27272757                94.66 ns/op           49 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       14540809               150.7 ns/op           148 B/op          3 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     15997461               155.1 ns/op           202 B/op          3 allocs/op
[...]
PASS
```

## Fast
As you can see, most operations are done in **under 0.3μs** and can therefore be done over **3 Million times / second**.
It also **avoids GC** checks. This is achieved by storing the objects in one byte-slice.

## Concurrent
The map has **no global lock**.
It is split into **16 Shards** which are locked individual. As the benchmarks show **it gains from concurrent access**

As the benchmarks already show, the BigMap gains from concurrency.
As it claims a big chunk of memory and items have a max size it is faster than the standart map.
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

## Scaling
If you have more concurrent accesses, you can always increase the shard count.
As always: only benchmarking **your usecase** will reveal the optimal settings.
But as shown, with the default 16 shards, you don't get a bad access speed even with half a million routines.
Each shard can store gigabytes of data without loosing performance, so it is good for storing tons of tons of normalized data.

## Attention
The map scales as more data is added but, to enable high performance, doesn't schrink.
To enable the fast accessess free heap is held "hot" to be ready to use.
This means the map might grow once realy big but then never grows again.

