package bigmap

import (
	"fmt"
	"math/rand"
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
	for i := 0; i < b.N; i++ {
		shard.Put(FNV64(GenKey(i)), GenVal())
	}
}

func BenchmarkShard_Get(b *testing.B) {
	shard := NewShard(1024, 100)
	for i := 0; i < b.N; i++ {
		shard.Put(FNV64(GenKey(i)), GenVal())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shard.Get(FNV64(GenKey(i)))
	}
}

func BenchmarkShard_Mix_Ballanced(b *testing.B) {
	shard := NewShard(1024, 100)
	for i := 0; i < b.N; i++ {
		k := FNV64(GenKey(i))
		shard.Put(k, GenVal())
		shard.Get(k)
		shard.Delete(k)
	}
}

func BenchmarkBigMap_Put(b *testing.B) {
	bigmap := New(100)
	for i := 0; i < b.N; i++ {
		bigmap.Put(GenKey(i), GenVal())
	}
}

func BenchmarkBigMap_Get(b *testing.B) {
	bigmap := New(100)
	for i := 0; i < b.N; i++ {
		bigmap.Put(GenKey(i), GenVal())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Get(GenKey(i))
	}
}

func BenchmarkBigMap_Mix_Ballanced(b *testing.B) {
	bigmap := New(100)
	for i := 0; i < b.N; i++ {
		k := GenKey(i)
		bigmap.Put(k, GenVal())
		bigmap.Get(k)
		bigmap.Delete(k)
	}
}

func BenchmarkBigMap_Get_Parallel(b *testing.B) {
	bigmap := New(100)
	for i := 0; i < b.N; i++ {
		bigmap.Put(GenKey(i), GenVal())
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		i := 0
		for p.Next() {
			bigmap.Get(GenKey(i))
			i++
		}
	})
}

func BenchmarkBigMap_Put_Parallel(b *testing.B) {
	bigmap := New(100)
	rand.Seed(time.Now().UnixNano())

	b.RunParallel(func(p *testing.PB) {
		worker := rand.Intn(1024)
		i := 0
		for p.Next() {
			bigmap.Put(GenSafeKey(worker, i), GenVal())
			i++
		}
	})
}

func BenchmarkBigMap_Mix_Ballanced_Parallel(b *testing.B) {
	bigmap := New(100)

	rand.Seed(time.Now().UnixNano())

	b.RunParallel(func(p *testing.PB) {
		worker := rand.Intn(1024)
		i := 0
		for p.Next() {
			k := GenSafeKey(worker, i)
			bigmap.Put(k, GenVal())
			bigmap.Get(k)
			bigmap.Delete(k)
		}
	})
}

func GenSafeKey(a, b int) string {
	return fmt.Sprintf("gen-safe-%d-%d", a, b)
}

func GenKey(i int) string {
	return fmt.Sprintf("gen-%d", i)
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
