# BigMap
[![GoReport](https://goreportcard.com/badge/github.com/worldOneo/bigmap)](https://goreportcard.com/report/github.com/worldOneo/bigmap)  
## Fast - Scaling - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Fast

Most operations are done in **about 0.2μs** and can therefore be done **5 Million times / second**.  
And all this **per thread**. This is achieved by storing the objects in one single byte-slice and having a Zero-Allocation, Share-Nothing oriented design.  
Resulting in **minimimal GC pressure** and **maximum performance**.

## Concurrent

The map has **no global lock**.  
It is split into **multiple shards** which are locked individual.  
As the benchmarks show bigmap **gains from concurrent access**.  
With preallocations and items having a max size it is **faster than the standard map**.

## Scaling

Each shard can store gigabytes of data without loosing performance, so it is good for storing tons of tons of normalized data.
If you have more concurrent accesses, you can always increase the shard count.  
As always: only benchmarking **your usecase** will reveal the optimal settings.  

## Benchmarks

The benchmarks are done on a machine with an i7-8750H CPU (6c/12t 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
The key size is ~24 bytes and the value size is 100 bytes. All settings are default.
We can see I reach up to ~16 million OPs per second in the 10% Write 10% Delete 80% Read parallel benchmark on my machine.
The MB/s can must be seen as OP/s and are 1/100th of the real throughput.

```sh
go version
go version go1.19.1 windows/amd64

go.exe test -benchmem -run=^$ -bench "BenchmarkGenKey.*|BenchmarkBigMap.*" github.com/worldOneo/bigmap --benchtime=20000000x
goos: windows
goarch: amd64
pkg: github.com/worldOneo/bigmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkGenKey-12                              20000000               174.5 ns/op            40 B/op          3 allocs/op
BenchmarkBigMap_Put-12                          20000000               317.1 ns/op         3.15 MB/s         483 B/op          0 allocs/op
BenchmarkBigMap_Get-12                          20000000               213.9 ns/op         4.67 MB/s         112 B/op          1 allocs/op
BenchmarkBigMap_GetInto-12                      20000000               169.2 ns/op         5.91 MB/s           0 B/op          0 allocs/op
BenchmarkBigMap_Delete-12                       20000000               151.1 ns/op         6.62 MB/s          26 B/op          0 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                20000000                37.80 ns/op       26.46 MB/s           0 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12              20000000               181.0 ns/op         5.52 MB/s         140 B/op          0 allocs/op
# Parallel benchmarks have allocations because of the key generation
# which makes them slightly slower than in a perfect real world application.
BenchmarkBigMap_Put_Parallel-12                 20000000               152.5 ns/op         6.56 MB/s         539 B/op          2 allocs/op
BenchmarkBigMap_Get_Parallel-12                 20000000                71.12 ns/op       14.06 MB/s         112 B/op          1 allocs/op
BenchmarkBigMap_GetInto_Parallel-12             20000000                41.21 ns/op       24.27 MB/s           0 B/op          0 allocs/op
BenchmarkBigMap_Delete_Parallel-12              20000000                82.70 ns/op       12.09 MB/s          66 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       20000000               104.6 ns/op         9.56 MB/s         190 B/op          2 allocs/op
BenchmarkBigMap_10_10_80_Parallel-12            20000000                62.17 ns/op       16.08 MB/s          87 B/op          2 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     20000000                79.32 ns/op       12.61 MB/s         162 B/op          2 allocs/op
```

## Attention
The map scales as more data is added but, to enable high performance, doesn't schrink.
To enable the fast accessess free heap is held "hot" to be ready to use.
This means the map might grow once realy big, which might seeme like a memory leak at first glance because it doesn shrink, but then never grows again.  
