# BigMap
[![GoReport](https://goreportcard.com/badge/github.com/worldOneo/bigmap)](https://goreportcard.com/report/github.com/worldOneo/bigmap)  
## Fast - Scaling - Concurrent map for serializeable data

Inspired by [allegro/bigcache](https://github.com/allegro/bigcache/)

## Fast

Most operations are done in **about 0.2μs** and can therefore be done **5 Million times / second**.  
And all this **per thread**. This is achieved by storing the objects in one single byte-slice and having a Zero-Allocation, Share-Nothing oriented design.  
Resulting in **minimimal GC pressure** and **maximal performance**.

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

## Benchmarks

The benchmarks are done on a machine with an i7-8750H CPU (12x 2.20 - 4GHz), 16GB  RAM (2666 MHz), Windows 10 machine
```sh
go version
go version go1.17.2 windows/amd64

go test -benchmem -run=^$ -bench BenchmarkBigMap.* github.com/worldOneo/bigmap -benchtime=2s
goos: windows
goarch: amd64
pkg: github.com/worldOneo/bigmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz

BenchmarkGenKey-12                              20750223               113.5 ns/op            24 B/op          2 allocs/op
BenchmarkFNV64-12                               472440758                5.106 ns/op           0 B/op          0 allocs/op
BenchmarkBigMap_Put-12                           4457443               272.7 ns/op           301 B/op          0 allocs/op
BenchmarkBigMap_Get-12                           6747880               217.0 ns/op           112 B/op          1 allocs/op
BenchmarkBigMap_Delete-12                        9018730               154.9 ns/op            14 B/op          0 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                19041000                60.56 ns/op           37 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12               6570567               189.2 ns/op           144 B/op          0 allocs/op
# Parallel benchmarks have allocations because of the key generation (113.5ns; 2 allocs/op).
# That also slows them down a little but this is required for the parallel test.
BenchmarkBigMap_Put_Parallel-12                  5863574               207.6 ns/op           468 B/op          2 allocs/op
BenchmarkBigMap_Get_Parallel-12                 12765644               103.1 ns/op           151 B/op          3 allocs/op
BenchmarkBigMap_Delete_Parallel-12              11317700               107.2 ns/op            51 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12        6312416               167.6 ns/op           162 B/op          2 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12      7689206               190.4 ns/op           217 B/op          3 allocs/op
```

## Attention
The map scales as more data is added but, to enable high performance, doesn't schrink.
To enable the fast accessess free heap is held "hot" to be ready to use.
This means the map might grow once realy big, which might seeme like a memory leak at first glance because it doesn shrink, but then never grows again.  
