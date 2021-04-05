# BigMap
## Scaling - Fast - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Benchmarks
The benchmarks are done on a machine with an i7-8750H CPU (12x 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
```
go version
go version go1.16.2 windows/amd64

go test -benchmem -run=^$ -bench .* github.com/worldOneo/bigmap -benchtime=2s
goos: windows
goarch: amd64
pkg: github.com/worldOneo/bigmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkShard_Put-12                            6885738               372.2 ns/op           399 B/op          1 allocs/op
BenchmarkShard_Get-12                           11115006               280.2 ns/op           119 B/op          1 allocs/op
BenchmarkShard_Delete-12                        11428695               228.1 ns/op            19 B/op          0 allocs/op
BenchmarkShard_Mix_Ballanced-12                 31176806                72.91 ns/op           44 B/op          0 allocs/op
BenchmarkShard_Mix_Unballanced-12               11389477               278.0 ns/op           166 B/op          1 allocs/op
BenchmarkBigMap_Put-12                           6826752               359.1 ns/op           402 B/op          1 allocs/op
BenchmarkBigMap_Get-12                          10797139               250.3 ns/op            20 B/op          1 allocs/op
BenchmarkBigMap_Delete-12                        9357001               277.7 ns/op           119 B/op          1 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                26223202                89.08 ns/op           45 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12               9896658               281.8 ns/op           172 B/op          1 allocs/op
BenchmarkBigMap_Put_Parallel-12                 13252675               156.1 ns/op           403 B/op          2 allocs/op
BenchmarkBigMap_Get_Parallel-12                 55826805                65.49 ns/op           38 B/op          2 allocs/op
BenchmarkBigMap_Delete_Parallel-12              35703021                74.55 ns/op           36 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       16510561               152.3 ns/op           182 B/op          2 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     10160574               224.3 ns/op            18 B/op          1 allocs/op
PASS
ok      github.com/worldOneo/bigmap     105.974s

go test -benchmem -run=^$ -bench BenchmarkBigMap_Goroutines github.com/worldOneo/bigmap -benchtime=3s
 === 
Benchmarking with 480000 routines
 ===
goos: windows
goarch: amd64
pkg: github.com/worldOneo/bigmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkBigMap_Goroutines/BigMap_Put_Parallel-12               11916488       282.2 ns/op      414 B/op     0 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Get_Parallel-12               59722618       150.2 ns/op       47 B/op     1 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Delete_Parallel-12            42789036       121.6 ns/op       16 B/op     1 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Mix_Parallel-12               15347977       204.7 ns/op       92 B/op     0 allocs/op
BenchmarkBigMap_Goroutines/BigMap_Mix_Unbalanced_Parallel-12    16880276       214.4 ns/op      148 B/op     0 allocs/op
PASS
ok      github.com/worldOneo/bigmap     136.896s
```

## Fast
As you can see, most operations are done in **under 0.3μs** and can therefore be accessed over **3 Million times / second**.
It also **avoids GC** checks. This is achieved by storing the objects in one byte-slice.

## Concurrent
The map has **no global lock**.
It is split into **16 Shards** which are locked individual. As the benchmarks show **it gains from concurrent access**

As the benchmarks already show, the BigMap gains from concurrency.
As it claims a big chunk of memory and items have a max size it is faster than the standart map.
### Sync
| | Time | diff | GC Pause total | GC Diff |
| --- | --- |--- | --- | --- |
| **BigMap** | 2m1.5s | 0% | 6.6ms | 0% |
| **StdMap** | 3m21s | +66% | 14.7ms | +22% |
| **Syncmap** | 9m46.9s | +383% | 16.6ms | +51% |
 
### Async
| | Time | diff |
| --- | --- |--- |
| **BigMap** | 38.1s | 0% |
| **StdMap** (synced with RWMutex) | 4m32s | +456% |
| **Syncmap** | 10m35.6s | +1.568% |

## Scaling
If you have more concurrent accesses, you always increase the shard count.
As always: only benchmarking **your usecase** will reveal the required settings.