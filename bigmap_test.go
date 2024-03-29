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
		GenSafeKey("00", i)
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
	b.SetBytes(1)
}

func BenchmarkBigMap_Get(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Get(keys[i])
	}
	b.SetBytes(1)
}

func BenchmarkBigMap_GetInto(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMap(b.N, &bigmap)
	buff := make([]byte, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.GetInto(keys[i], buff)
	}
	b.SetBytes(1)
}

func BenchmarkBigMap_Delete(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMap(b.N, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bigmap.Delete(keys[i])
	}
	b.SetBytes(1)
}

func BenchmarkBigMap_Mix_Ballanced(b *testing.B) {
	bigmap := New(100)
	keys := GenMapKeys(b.N)
	buff := make([]byte, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N/3; i++ {
		k := keys[i]
		bigmap.Put(k, GenVal())
		bigmap.GetInto(k, buff)
		bigmap.Delete(k)
	}
	b.SetBytes(1)
}

func BenchmarkBigMap_Mix_Unballanced(b *testing.B) {
	bigmap := New(100)
	N := b.N/3 + 1
	keys := GenMapKeys(N)
	buff := make([]byte, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < N; i++ {
		bigmap.Put(keys[i], GenVal())
	}
	for i := 0; i < N; i++ {
		bigmap.GetInto(keys[i], buff)
	}
	for i := 0; i < N; i++ {
		bigmap.Delete(keys[i])
	}
	b.SetBytes(1)
}

func BenchmarkBigMap_Put_Parallel(b *testing.B) {
	bigmap := New(100)
	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		val := make([]byte, 100)
		worker := strconv.Itoa(rand.Int())
		i := 0
		for p.Next() {
			bigmap.Put(GenSafeKey(worker, i), val)
			i++
		}
	})
	b.SetBytes(1)
}

func BenchmarkBigMap_Get_Parallel(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMapParallel(b, &bigmap)
	PopulateMapParallel(b, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	w := int32(-1)
	b.RunParallel(func(p *testing.PB) {
		id := int(atomic.AddInt32(&w, 1))
		len := len(keys[id])
		i := 0
		for p.Next() {
			bigmap.Get(keys[id][i%len])
			i++
		}
	})
	b.SetBytes(1)
}

func BenchmarkBigMap_GetInto_Parallel(b *testing.B) {
	bigmap := New(100)
	keys := PopulateMapParallel(b, &bigmap)

	b.ReportAllocs()
	b.ResetTimer()
	w := int32(-1)
	b.RunParallel(func(p *testing.PB) {
		buff := make([]byte, 100)
		id := int(atomic.AddInt32(&w, 1))
		len := len(keys[id])
		i := 0
		for p.Next() {
			bigmap.GetInto(keys[id][i%len], buff)
			i++
		}
	})
	b.SetBytes(1)
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
	b.SetBytes(1)
}

func BenchmarkBigMap_Mix_Ballanced_Parallel(b *testing.B) {
	bigmap := New(100)
	rand.Seed(time.Now().UnixNano())

	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		buff := make([]byte, 100)
		worker := strconv.Itoa(rand.Int())
		i := 0
		for p.Next() {
			k := GenSafeKey(worker, i)
			if i%3 == 0 {
				bigmap.Put(k, buff)
			} else if i%3 == 1 {
				bigmap.GetInto(k, buff)
			} else {
				bigmap.Delete(k)
			}
			i++
		}
	})
	b.SetBytes(1)
}

func BenchmarkBigMap_10_10_80_Parallel(b *testing.B) {
	bigmap := New(100)
	rand.Seed(time.Now().UnixNano())
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		buff := make([]byte, 100)
		worker := strconv.Itoa(rand.Int())
		inserted := [24][]byte{}
		for i := 0; i < 24; i++ {
			inserted[i] = GenSafeKey(worker, i)
		}
		i := 0
		for p.Next() {
			if i%10 == 0 {
				inserted[i%24] = GenSafeKey(worker, i)
				bigmap.Put(inserted[i%24], buff)
			} else if i%10 == 1 {
				bigmap.Delete(inserted[i%24])
			} else {
				bigmap.GetInto(GenSafeKey(worker, i), buff)
			}
			i++
		}
	})
	b.SetBytes(1)
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
		buff := make([]byte, 100)
		worker := strconv.Itoa(rand.Int())
		inserted := [24][]byte{}
		for i := 0; i < 24; i++ {
			inserted[i] = GenSafeKey(worker, i)
		}
		r := rand.Uint32()
		i := 0
		s := 0
		for p.Next() {
			i++
			switch s {
			case 0:
				k := GenSafeKey(worker, i)
				inserted[i%24] = k
				bigmap.Put(k, buff)
			case 1:
				k := inserted[i%24]
				bigmap.Delete(k)
			case 2:
				k := GenSafeKey(worker, i)
				bigmap.GetInto(k, buff)
			}
			r = hash(r)
			if r&0xFFFF == 0xFFFF { // Switch randomly
				i = 0
				s++
				s %= 3
			}
		}
	})
	b.SetBytes(1)
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

func PopulateMapParallel(b *testing.B, bm *BigMap) [][][]byte {
	n := int(math.Min(float64(b.N), 2000000))
	keys := make([][][]byte, runtime.GOMAXPROCS(0))
	//b.Logf("Populating map with %d items", n)
	wg := sync.WaitGroup{}
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		keys[i] = make([][]byte, n)
		wg.Add(1)
		w := strconv.Itoa(i)
		go func(num int, worker string) {
			for i := 0; i < n; i++ {
				keys[num][i] = GenSafeKey(worker, i)
				bm.Put(keys[num][i], make([]byte, bm.shards[0].entrysize))
			}
			wg.Done()
		}(i, w)
	}
	wg.Wait()
	return keys
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
		Shards:            3,
		Capacity:          128,
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
