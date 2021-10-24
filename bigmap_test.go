package bigmap

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func RandomString(n int) []byte {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return s
}

func BenchmarkGenKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenKey(i)
	}
}

func BenchmarkFNV64(b *testing.B) {
	// This benchmark is to fast to iterate over b.N items
	// therefore we need to limit the amount
	b.StopTimer()
	k := make([][]byte, 2_000_000)
	for i := 0; i < len(k); i++ {
		k[i] = GenKey(i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		FNV64(k[i%len(k)])
	}
}

func BenchmarkBigMap_Put(b *testing.B) {
	bigmap := New(100)
	val := GenVal()
	keys := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = GenKey(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Put(keys[i], val)
	}
}

func BenchmarkBigMap_Get(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Get(keys[i])
	}
}

func BenchmarkBigMap_Delete(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Delete(keys[i])
	}
}

func BenchmarkBigMap_Mix_Ballanced(b *testing.B) {
	bigmap := New(100)
	keys := GenMapKeys(b.N)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N/3; i++ {
		k := keys[i]
		bigmap.Put(k, GenVal())
		bigmap.Get(k)
		bigmap.Delete(k)
	}
}

func BenchmarkBigMap_Mix_Unballanced(b *testing.B) {
	bigmap := New(100)
	N := b.N/3 + 1
	keys := GenMapKeys(N)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < N; i++ {
		bigmap.Put(keys[i], GenVal())
	}
	for i := 0; i < N; i++ {
		bigmap.Get(keys[i])
	}
	for i := 0; i < N; i++ {
		bigmap.Delete(keys[i])
	}
}

func BenchmarkBigMap_Put_Parallel(b *testing.B) {
	bigmap := New(100)
	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(rand.Int())
		i := 0
		for p.Next() {
			bigmap.Put(GenSafeKey(worker, i), GenVal())
			i++
		}
	})
}

func BenchmarkBigMap_Get_Parallel(b *testing.B) {
	bigmap := New(100)
	PopulateMapParallel(b, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	w := int32(-1)
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(int(atomic.AddInt32(&w, 1)))
		i := 0
		for p.Next() {
			bigmap.Get(GenSafeKey(worker, i))
			i++
		}
	})
}

func BenchmarkBigMap_Delete_Parallel(b *testing.B) {
	bigmap := New(100)
	PopulateMapParallel(b, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	w := int32(0)
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(int(atomic.AddInt32(&w, 1)))
		i := 0
		for p.Next() {
			bigmap.Delete(GenSafeKey(worker, i))
			i++
		}
	})
}

func BenchmarkBigMap_Mix_Ballanced_Parallel(b *testing.B) {
	bigmap := New(100)

	rand.Seed(time.Now().UnixNano())

	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(rand.Int())
		i := 0
		for p.Next() {
			k := GenSafeKey(worker, i)
			if i%3 == 0 {
				bigmap.Put(k, GenVal())
			} else if i%3 == 1 {
				bigmap.Get(k)
			} else {
				bigmap.Delete(k)
			}
			i++
		}
	})
}

// Hash by Thomas Wang (https://burtleburtle.net/bob/hash/integer.html)
func hash(a uint32) uint32 {
	a = (a ^ 61) ^ (a >> 16)
	a = a + (a << 3)
	a = a ^ (a >> 4)
	a = a * 0x27d4eb2d
	a = a ^ (a >> 15)
	return a
}

func BenchmarkBigMap_Mix_Unballanced_Parallel(b *testing.B) {
	bigmap := New(100)

	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(rand.Int())
		r := rand.Uint32()
		i := 0
		s := 0
		for p.Next() {
			i++
			k := GenSafeKey(worker, i)
			switch s {
			case 0:
				bigmap.Put(k, GenVal())
			case 1:
				bigmap.Delete(k)
			case 2:
				bigmap.Get(k)
			}
			r = hash(r)
			if r&0xFFFF == 0xFFFF { // Switch randomly
				i = 0
				s++
				s %= 3
			}
		}
	})
}

func BenchmarkBigMap_Goroutines(b *testing.B) {
	BenchParallel(b, 10)
	BenchParallel(b, 1000)
	BenchParallel(b, 10000)
}

func BenchParallel(b *testing.B, n int) {
	fmt.Printf(" === \nBenchmarking with %d routines\n === \n", runtime.GOMAXPROCS(0)*n)
	b.Run("BigMap Put Parallel", wrap(BenchmarkBigMap_Put_Parallel, n))
	b.Run("BigMap Get Parallel", wrap(BenchmarkBigMap_Get_Parallel, n))
	b.Run("BigMap Delete Parallel", wrap(BenchmarkBigMap_Delete_Parallel, n))
	b.Run("BigMap Mix Parallel", wrap(BenchmarkBigMap_Mix_Ballanced_Parallel, n))
	b.Run("BigMap Mix Unbalanced Parallel", wrap(BenchmarkBigMap_Mix_Ballanced_Parallel, n))
}

func wrap(a func(b *testing.B), i int) func(b *testing.B) {
	return func(b *testing.B) {
		b.SetParallelism(i)
		a(b)
	}
}

func GenSafeKey(a string, b int) []byte {
	return []byte(fmt.Sprintf("gen-%s-%d", a, b))
}

func GenKey(i int) []byte {
	return []byte(fmt.Sprintf("gen-%d", i))
}

func GenMapKeys(n int) [][]byte {
	keys := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = GenKey(i)
	}
	return keys
}

func PopulateMap(n int, bm *BigMap) [][]byte {
	keys := make([][]byte, n)
	val := GenVal()
	for i := 0; i < n; i++ {
		keys[i] = GenKey(i)
		bm.Put(keys[i], val)
	}
	return keys
}

func PopulateMapParallel(b *testing.B, bm *BigMap) {
	n := int(math.Min(float64(b.N), 2000000))
	//b.Logf("Populating map with %d items", n)
	wg := sync.WaitGroup{}
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		w := strconv.Itoa(i)
		go func(worker string) {
			for i := 0; i < n; i++ {
				bm.Put(GenSafeKey(worker, i), GenVal())
			}
			wg.Done()
		}(w)
	}
	wg.Wait()
}

func GenVal() []byte {
	return make([]byte, 100)
}

func TestBigMap(t *testing.T) {
	keys := make([][]byte, 4096*8)
	vals := make([][]byte, 4096*8)
	for i := range keys {
		keys[i] = RandomString(10)
		vals[i] = RandomString(100)
	}
	bigmap := New(100)
	for i, key := range keys {
		err := bigmap.Put(key, vals[i])
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for i, key := range keys {
		val, _ := bigmap.Get(key)

		if string(val) != string(vals[i]) {
			t.Fatalf("val expected: '%s' != '%s' ", string(val), vals[i])
		}
	}

	for _, key := range keys {
		ok := bigmap.Delete(key)

		if !ok {
			t.Fatalf("delete expected")
		}
	}

	for i, key := range keys {
		err := bigmap.Put(key, vals[i])
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

}


func TestBigMap_New_config(t *testing.T) {
	bigmap := New(100, Config{
		Shards: 3,
		Capacity: 128,
		ExpirationFactory: Expires(time.Hour, ExpirationPolicyPassive),
	})
	if len(bigmap.shards) != 3 {
		t.Fatalf("Failed to configure shards got %d, want %d", len(bigmap.shards), 3)
	}
	if len(bigmap.shards[0].array) != 128 {
		t.Fatalf("Failed to configure capacity got %d, want %d", len(bigmap.shards[0].array), 128)
	}
	if bigmap.shards[0].expSrv == nil {
		t.Fatalf("Failed to configure expiration got nil, want !nil")
	}
}