package bigmap

import (
	"math/rand"
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
	N := b.N / 3
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
	N := b.N / 3
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

	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		worker := rand.Intn(1024)
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
		worker := rand.Intn(1024)
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
		worker := rand.Intn(1024)
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
		worker := rand.Intn(1024)
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

func BenchmarkBigMap_Mix_Unballanced_Parallel(b *testing.B) {
	bigmap := New(100)

	rand.Seed(time.Now().UnixNano())

	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		worker := rand.Intn(1024)
		i := 0
		s := 0
		for p.Next() {
			k := GenSafeKey(worker, i)
			switch s {
			case 0:
				bigmap.Put(k, GenVal())
			case 1:
				bigmap.Delete(k)
			case 2:
				bigmap.Get(k)
			}
			if rand.Float32() < float32(1)/100_000 {
				s++
				s %= 3
			}
		}
	})
}

func GenSafeKey(a, b int) string {
	return "gen-safe-" + strconv.Itoa(a) + "-" + strconv.Itoa(b)
}

func GenKey(i int) string {
	return "gen-" + strconv.Itoa(i)
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
	keys := make([]string, 4096)
	vals := make([]string, 4096)
	for i := range keys {
		keys[i] = RandomString(10)
		vals[i] = RandomString(100)
	}
	bigmap := New(100)
	for i, key := range keys {
		err := bigmap.Put(key, []byte(vals[i]))
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for i, key := range keys {
		val, ok := bigmap.Get(key)

		if !ok || string(val) != vals[i] {
			t.Fatalf("val expected: '%s' != '%s' ", string(val), vals[i])
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
		err := shard.Put(FNV64(key), []byte(vals[i]))
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for i, key := range keys {
		val, ok := shard.Get(FNV64(key))

		if !ok || string(val) != vals[i] {
			t.Fatalf("val expected: '%s' != '%s' ", string(val), vals[i])
		}
	}
}
