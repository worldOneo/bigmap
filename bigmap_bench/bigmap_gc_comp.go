package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	"github.com/worldOneo/bigmap"
)

const (
	iterations = 5
	items      = 20_000_000
)

var totallAccesses uint64

func main() {
	debug.SetGCPercent(10)
	bench("StdMap_Put", benchmarkStdMapPut)
	bench("StdMap_Get", benchmarkStdMapGet)
	bench("StdMap_Del ", benchmarkStdMapDelete)
	bench("SyncMap_Put ", benchmarkSyncMapPut)
	bench("SyncMap_Get ", benchmarkSyncMapGet)
	bench("SyncMap_Del ", benchmarkSyncMapDelete)

	fmt.Println("==  Sync ==")
	fmt.Println("Stdmap:")
	benchmarkGCPressure(mapGC)
	fmt.Println("Bigmap:")
	benchmarkGCPressure(bigmapGC)
	fmt.Println("Syncmap:")
	benchmarkGCPressure(syncmapGC)

	fmt.Println("== Async ==")
	fmt.Println("Bigmap:")
	benchmarkGCPressure(bigmapGCAsync)
	fmt.Println("Stdmap (locked)")
	benchmarkGCPressure(mapGCAsync)
	fmt.Println("Syncmap:")
	benchmarkGCPressure(syncGCAsync)
	fmt.Printf("Done! Totally generated %d keys\n", totallAccesses)
}

func bench(n string, f func(b *testing.B)) {
	r := testing.Benchmark(f)
	fmt.Printf("%s %s %s\n", n, r.String(), r.MemString())
}

func benchmarkGCPressure(a func(i int)) {
	runtime.GC()
	var stats debug.GCStats
	var mstats runtime.MemStats
	runtime.ReadMemStats(&mstats)
	debug.ReadGCStats(&stats)
	start := time.Now()
	before := stats.PauseTotal
	alloc := mstats.TotalAlloc
	for i := 0; i < iterations; i++ {
		a(items)
	}
	debug.ReadGCStats(&stats)
	runtime.ReadMemStats(&mstats)
	total := stats.PauseTotal - before
	stop := time.Since(start)
	memdiff := mstats.TotalAlloc - alloc
	fmt.Println("Alloc (MiB): ", (float64(memdiff)/1024)/1024)
	fmt.Println("Time: ", stop)
	fmt.Println("Total GC Pause: ", total)
}

func bigmapGC(n int) {
	bigmap := bigmap.New(64, bigmap.Config{
		Shards: 128,
	})
	for i := 0; i < n; i++ {
		k := []byte(key(i))
		bigmap.Put(k, val())
	}
	for i := 0; i < n; i++ {
		k := []byte(key(i))
		_, _ = bigmap.Get(k)
	}
	for i := 0; i < n; i++ {
		k := []byte(key(i))
		_ = bigmap.Delete(k)
	}
}

func mapGC(n int) {
	mp := make(map[string][]byte)
	for i := 0; i < n; i++ {
		k := key(i)
		mp[k] = val()
	}
	for i := 0; i < n; i++ {
		k := key(i)
		_ = mp[k]
	}
	for i := 0; i < n; i++ {
		k := key(i)
		delete(mp, k)
	}
}

func syncmapGC(n int) {
	mp := sync.Map{}
	for i := 0; i < n; i++ {
		k := key(i)
		mp.Store(k, val())
	}
	for i := 0; i < n; i++ {
		k := key(i)
		_, _ = mp.Load(k)
	}
	for i := 0; i < n; i++ {
		k := key(i)
		mp.Delete(k)
	}
}

func bigmapGCAsync(o int) {
	mp := bigmap.New(64, bigmap.Config{
		Shards: 128,
	})

	wg := sync.WaitGroup{}
	n := o / runtime.GOMAXPROCS(0)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(worker int) {
			for i := 0; i < n; i++ {
				k := []byte(keySafe(worker, i))
				mp.Put(k, val())
			}
			for i := 0; i < n; i++ {
				k := []byte(keySafe(worker, i))
				mp.Get(k)
			}
			for i := 0; i < n; i++ {
				k := []byte(keySafe(worker, i))
				mp.Delete(k)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func mapGCAsync(o int) {
	mp := make(map[string][]byte)
	l := sync.RWMutex{}

	wg := sync.WaitGroup{}
	n := o / runtime.GOMAXPROCS(0)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(worker int) {
			for i := 0; i < n; i++ {
				k := keySafe(worker, i)
				l.Lock()
				mp[k] = val()
				l.Unlock()
			}
			for i := 0; i < n; i++ {
				k := keySafe(worker, i)
				l.RLock()
				_ = mp[k]
				l.RUnlock()
			}
			for i := 0; i < n; i++ {
				k := keySafe(worker, i)
				l.Lock()
				delete(mp, k)
				l.Unlock()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func syncGCAsync(o int) {
	mp := &sync.Map{}

	wg := sync.WaitGroup{}
	n := o / runtime.GOMAXPROCS(0)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(worker int) {
			for i := 0; i < n; i++ {
				k := keySafe(worker, i)
				mp.Store(k, val())
			}
			for i := 0; i < n; i++ {
				k := keySafe(worker, i)
				mp.Load(k)
			}
			for i := 0; i < n; i++ {
				k := keySafe(worker, i)
				mp.Delete(k)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func benchmarkStdMapPut(b *testing.B) {
	mp := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		mp[key(i)] = val()
	}
}

func benchmarkStdMapGet(b *testing.B) {
	mp := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		mp[key(i)] = val()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mp[key(i)]
	}
}

func benchmarkStdMapDelete(b *testing.B) {
	mp := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		mp[key(i)] = val()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delete(mp, key(i))
	}
}

func benchmarkSyncMapPut(b *testing.B) {
	mp := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		mp[key(i)] = val()
	}
}

func benchmarkSyncMapGet(b *testing.B) {
	mp := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		mp[key(i)] = val()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mp[key(i)]
	}
}

func benchmarkSyncMapDelete(b *testing.B) {
	mp := sync.Map{}
	for i := 0; i < b.N; i++ {
		mp.Store(key(i), val())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.Delete(key(i))
	}
}

func keySafe(j, i int) string {
	totallAccesses++
	return fmt.Sprintf("gen-%d-%d", j, i)
}

func key(i int) string {
	totallAccesses++
	return fmt.Sprintf("gen-%d", i)
}

func val() []byte {
	return make([]byte, 64)
}
