package bigmap

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func RandomString(n int) string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func BenchmarkShard_Put(b *testing.B) {
	shard := NewShard(1024, 100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		shard.Put(FNV64(GenKey(i)), GenVal())
	}
}

func BenchmarkShard_Get(b *testing.B) {
	shard := NewShard(1024, 100)
	for i := 0; i < b.N; i++ {
		shard.Put(FNV64(GenKey(i)), GenVal())
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shard.Get(FNV64(GenKey(i)))
	}
}

func BenchmarkShard_Delete(b *testing.B) {
	shard := NewShard(1024, 100)
	for i := 0; i < b.N; i++ {
		shard.Put(FNV64(GenKey(i)), GenVal())
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shard.Delete(FNV64(GenKey(i)))
	}
}

func BenchmarkShard_Mix_Ballanced(b *testing.B) {
	shard := NewShard(1024, 100)
	b.ReportAllocs()
	for i := 0; i < b.N/3; i++ {
		k := FNV64(GenKey(i))
		shard.Put(k, GenVal())
		shard.Get(k)
		shard.Delete(k)
	}
}

func BenchmarkShard_Mix_Unballanced(b *testing.B) {
	shard := NewShard(1024, 100)
	b.ReportAllocs()
	N := b.N/3 + 1
	for i := 0; i < N; i++ {
		k := FNV64(GenKey(i))
		shard.Put(k, GenVal())
	}
	for i := 0; i < N; i++ {
		k := FNV64(GenKey(i))
		shard.Get(k)
	}
	for i := 0; i < N; i++ {
		k := FNV64(GenKey(i))
		shard.Delete(k)
	}
}

func BenchmarkBigMap_Put(b *testing.B) {
	bigmap := New(100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bigmap.Put(GenKey(i), GenVal())
	}
}

func BenchmarkBigMap_Get(b *testing.B) {
	bigmap := New(100)
	PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Delete(GenKey(i))
	}
}

func BenchmarkBigMap_Delete(b *testing.B) {
	bigmap := New(100)
	PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Get(GenKey(i))
	}
}

func BenchmarkBigMap_Mix_Ballanced(b *testing.B) {
	bigmap := New(100)
	b.ReportAllocs()
	for i := 0; i < b.N/3; i++ {
		k := GenKey(i)
		bigmap.Put(k, GenVal())
		bigmap.Get(k)
		bigmap.Delete(k)
	}
}

func BenchmarkBigMap_Mix_Unballanced(b *testing.B) {
	bigmap := New(100)
	N := b.N/3 + 1
	b.ReportAllocs()
	for i := 0; i < N; i++ {
		k := GenKey(i)
		bigmap.Put(k, GenVal())
	}
	for i := 0; i < N; i++ {
		k := GenKey(i)
		bigmap.Get(k)
	}
	for i := 0; i < N; i++ {
		k := GenKey(i)
		bigmap.Delete(k)
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
	PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(rand.Int())
		i := 0
		for p.Next() {
			bigmap.Get(GenSafeKey(worker, i))
			i++
		}
	})
}

func BenchmarkBigMap_Delete_Parallel(b *testing.B) {
	bigmap := New(100)
	PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		worker := strconv.Itoa(rand.Int())
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
	BenchParallel(b, 40000)
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

func PopulateMap(n int, bm *BigMap) {
	for i := 0; i < n; i++ {
		bm.Put(GenKey(i), GenVal())
	}
}

func GenVal() []byte {
	return make([]byte, 100)
}

func TestBigMap(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	keys := make([]string, 4096*8)
	vals := make([]string, 4096*8)
	for i := range keys {
		keys[i] = RandomString(10)
		vals[i] = RandomString(100)
	}
	bigmap := New(100)
	for i, key := range keys {
		err := bigmap.Put([]byte(key), []byte(vals[i]))
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for i, key := range keys {
		val, _ := bigmap.Get([]byte(key))

		if string(val) != vals[i] {
			t.Fatalf("val expected: '%s' != '%s' ", string(val), vals[i])
		}
	}

	for _, key := range keys {
		ok := bigmap.Delete([]byte(key))

		if !ok {
			t.Fatalf("delete expected")
		}
	}

	for i, key := range keys {
		err := bigmap.Put([]byte(key), []byte(vals[i]))
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

}

func TestShard(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	keys := make([]string, 4096)
	vals := make([]string, 4096)
	for i := range keys {
		keys[i] = RandomString(10)
		vals[i] = RandomString(100)
	}
	shard := NewShard(1024, 1024)
	for i, key := range keys {
		err := shard.Put(FNV64([]byte(key)), []byte(vals[i]))
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for i, key := range keys {
		val, ok := shard.Get(FNV64([]byte(key)))

		if !ok || string(val) != vals[i] {
			t.Fatalf("val expected: '%s' != '%s' ", string(val), vals[i])
		}
	}

	for _, key := range keys {
		ok := shard.Delete(FNV64([]byte(key)))

		if !ok {
			t.Fatalf("delete expected")
		}
	}
}
