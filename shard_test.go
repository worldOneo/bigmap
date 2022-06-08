package bigmap

import (
	"testing"
)

func GenShardKeys(n int) []uint64 {
	keys := make([]uint64, n)
	for i := 0; i < n; i++ {
		keys[i] = FNV64(GenKey(i))
	}
	return keys
}

func WithShard(b *testing.B, populate bool, bench func(shard *Shard, keys []uint64)) {
	shard := NewShard(1024, 100, nil)
	keys := GenShardKeys(b.N)
	if populate {
		for i := 0; i < b.N; i++ {
			shard.Put(keys[i], GenVal())
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	bench(shard, keys)
}

func BenchmarkShard_Put(b *testing.B) {
	val := GenVal()
	WithShard(b, false, func(shard *Shard, keys []uint64) {
		for i := 0; i < b.N; i++ {
			shard.Put(keys[i], val)
		}
	})
}

func BenchmarkShard_Put_Stretched(b *testing.B) {
	shard := NewShard(1024, 100, nil)
	for i := 0; i < b.N/2; i++ {
		shard.Put(FNV64(GenSafeKey("singly", i)), GenVal())
	}
	for i := 0; i < b.N/2; i++ {
		shard.Delete(FNV64(GenSafeKey("singly", i)))
	}

	keys := GenShardKeys(b.N)
	val := GenVal()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shard.Put(keys[i], val)
	}
}

func BenchmarkShard_Get(b *testing.B) {
	WithShard(b, true, func(shard *Shard, keys []uint64) {
		for i := 0; i < b.N; i++ {
			shard.Get(keys[i])
		}
	})
}

func BenchmarkShard_Delete(b *testing.B) {
	WithShard(b, true, func(shard *Shard, keys []uint64) {
		for i := 0; i < b.N; i++ {
			shard.Delete(keys[i])
		}
	})
}

func BenchmarkShard_Mix_Ballanced(b *testing.B) {
	shard := NewShard(1024, 100, nil)
	keys := GenShardKeys(b.N)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N/3; i++ {
		shard.Put(keys[i], GenVal())
		shard.Get(keys[i])
		shard.Delete(keys[i])
	}
}

func BenchmarkShard_Mix_Unballanced(b *testing.B) {
	shard := NewShard(1024, 100, nil)
	keys := GenShardKeys(b.N)
	b.ReportAllocs()
	b.ResetTimer()
	N := b.N/3 + 1
	for i := 0; i < N; i++ {
		shard.Put(keys[i], GenVal())
	}
	for i := 0; i < N; i++ {
		shard.Get(keys[i])
	}
	for i := 0; i < N; i++ {
		shard.Delete(keys[i])
	}
}

func TestShard(t *testing.T) {
	keys := make([]uint64, 4096)
	vals := make([][]byte, 4096)
	for i := range keys {
		keys[i] = FNV64(RandomString(10))
		vals[i] = RandomString(100)
	}
	shard := NewShard(1024, 100, nil)
	for i, key := range keys {
		err := shard.Put(key, vals[i])
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for i, key := range keys {
		val, ok := shard.Get(key)

		if !ok || string(val) != string(vals[i]) {
			t.Fatalf("val expected: '%s' != '%s' ", string(val), vals[i])
		}
	}

	for _, key := range keys {
		ok := shard.Delete(key)

		if !ok {
			t.Fatalf("delete expected")
		}
	}
}

func TestShard_Put_insertToBig(t *testing.T) {
	shard := NewShard(1024, 100, nil)
	err := shard.Put(123, []byte(RandomString(111)))
	if err == nil {
		t.Fatal("To big insert got nil, want err")
	}
}
