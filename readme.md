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
The key size is ~24 bytes and the value size is 100 bytes. All settings are default.
We can see I reach up to ~15 million OPs per second in the 10% Write 10% Delete 80% Read parallel benchmark on my machine.

```sh
go version  
go version go1.18.1 windows/amd64

go.exe test -benchmem -run=^$ -bench "BenchmarkGenKey.*|BenchmarkBigMap.*" github.com/worldOneo/bigmap --benchtime=2s
goos: windows
goarch: amd64
pkg: github.com/worldOneo/bigmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkGenKey-12                              20777691               116.3 ns/op            24 B/op          2 allocs/op
BenchmarkBigMap_Put-12                           9230487               287.3 ns/op           290 B/op          0 allocs/op
BenchmarkBigMap_Get-12                          12443001               246.9 ns/op           112 B/op          1 allocs/op
BenchmarkBigMap_Delete-12                       17242245               180.2 ns/op            31 B/op          0 allocs/op
BenchmarkBigMap_Mix_Ballanced-12                49623379                52.27 ns/op           37 B/op          0 allocs/op
BenchmarkBigMap_Mix_Unballanced-12              14191022               198.2 ns/op           141 B/op          0 allocs/op
# Parallel benchmarks have allocations because of the key generation (116.3 ns/op; 24 B/op; 2 allocs/op)
# which also makes them slightly slower.
BenchmarkBigMap_10_10_80_Parallel-12            31258323                75.20 ns/op           92 B/op          2 allocs/op
BenchmarkBigMap_Put_Parallel-12                 15186710               164.1 ns/op           409 B/op          2 allocs/op
BenchmarkBigMap_Get_Parallel-12                 30183781                86.25 ns/op          129 B/op          3 allocs/op
BenchmarkBigMap_Delete_Parallel-12              27003519                84.30 ns/op           59 B/op          2 allocs/op
BenchmarkBigMap_Mix_Ballanced_Parallel-12       20456823               114.1 ns/op           187 B/op          2 allocs/op
BenchmarkBigMap_Mix_Unballanced_Parallel-12     24044000                86.45 ns/op          172 B/op          2 allocs/op
```

## Attention
The map scales as more data is added but, to enable high performance, doesn't schrink.
To enable the fast accessess free heap is held "hot" to be ready to use.
This means the map might grow once realy big, which might seeme like a memory leak at first glance because it doesn shrink, but then never grows again.  
