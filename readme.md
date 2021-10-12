# BigMap
## Scaling - Fast - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Benchmarks
The benchmarks are done on a machine with an i7-8750H CPU (12x 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
```sh
go version
go version go1.16.2 windows/amd64

go test -benchmem -run=^$ -bench .* github.com/worldOneo/bigmap -benchtime=2s

BenchmarkGenKey-12                              20893110               114.0 ns/op            24 B/op          2 allocs/op
BenchmarkFNV64-12                               458367789                5.318 ns/op           0 B/op          0 allocs/op
BenchmarkBigMap_Put-12                           7318198               350.4 ns/op           368 B/op          0 allocs/op
BenchmarkBigMap_Get-12                          12140280               267.0 ns/op           112 B/op          1 allocs/op
BenchmarkBigMap_Delete-12                       15841792               180.8 ns/op             8 B/op          0 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                38933733                63.11 ns/op           37 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12              13309575               222.6 ns/op           143 B/op          0 allocs/op
BenchmarkShard_Put-12                           10402132               324.0 ns/op           466 B/op          0 allocs/op
BenchmarkShard_Put_Stretched-12                 10161430               240.0 ns/op           345 B/op          0 allocs/op
BenchmarkShard_Get-12                           15021987               340.0 ns/op           112 B/op          1 allocs/op
BenchmarkShard_Delete-12                        16946853               152.8 ns/op            15 B/op          0 allocs/op
BenchmarkShard_Mix_Ballanced-12                 46158284                48.51 ns/op           37 B/op          0 allocs/op
BenchmarkShard_Mix_Unballanced-12               16336854               199.1 ns/op           190 B/op          0 allocs/op
# Parallel benchmarks have allocations because of the key generation.
# That also slows them down a little but this is required for the parallel test.
BenchmarkBigMap_Put_Parallel-12                 13968252               172.8 ns/op           442 B/op          3 allocs/op
BenchmarkBigMap_Get_Parallel-12                 32586734               105.4 ns/op           122 B/op          3 allocs/op
BenchmarkBigMap_Delete_Parallel-12              26624924               109.5 ns/op            50 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       12507817               182.9 ns/op           163 B/op          3 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     12417795               162.6 ns/op           158 B/op          3 allocs/op


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

## Fast
As you can see, most operations are done in **under 0.3μs** and can therefore be done over **3 Million times / second**.
It also **avoids GC** checks.
This is achieved by storing the objects in one single byte-slice and having a Zero-Allocation oriented design.

## Concurrent
The map has **no global lock**.
It is split into **16 Shards** which are locked individual. As the benchmarks show **it gains from concurrent access**

As the benchmarks already show, the BigMap gains from concurrency.
As it claims a big chunk of memory and items have a max size it is faster than the standard map.
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

