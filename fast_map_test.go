package bigmap

import (
	"testing"
)

func BenchmarkStdMap_Set(b *testing.B) {
	mp := make(map[uint64]uint32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp[uint64(i)] = uint32(i)
	}
}

func fullMap(n int) *FastMap {
	mp := NewFastMap()
	for i := 0; i < n; i++ {
		mp.Put(uint64(i), uint32(i))
	}
	return mp
}

func BenchmarkFastMap_Put(b *testing.B) {
	fullMap(b.N)
}

func BenchmarkStdMap_Get(b *testing.B) {
	mp := make(map[uint64]uint32)
	for i := 0; i < b.N; i++ {
		mp[uint64(i)] = uint32(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mp[uint64(i)]
	}
}

func BenchmarkFastMap_Get(b *testing.B) {
	mp := fullMap(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mp.Get(uint64(i))
	}
}

func BenchmarkStdMap_Delete(b *testing.B) {
	mp := make(map[uint64]uint32)
	for i := 0; i < b.N; i++ {
		mp[uint64(i)] = uint32(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delete(mp, uint64(i))
	}
}

func BenchmarkFastMap_Delete(b *testing.B) {
	mp := fullMap(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mp.Delete(uint64(i))
	}
}

func TestFastMap_Put(t *testing.T) {
	fullMap(3000)
}

func TestFastMap_Get(t *testing.T) {
	mp := fullMap(3000)
	for i := 0; i < 3000; i++ {
		v, ok := mp.Get(uint64(i))
		if v != uint32(i) || !ok {
			t.Fatalf("Failed to read key %d %d ok=%t", i, v, ok)
		}
	}
	_, ok := mp.Get(uint64(3001))
	if ok {
		t.Fatalf("Falsely returned key existence")
	}
}

func TestFastMap_Delete(t *testing.T) {
	mp := fullMap(3000)
	for i := 0; i < 3000; i++ {
		v, ok := mp.Delete(uint64(i))
		if v != uint32(i) || !ok {
			t.Fatalf("Deleted key didnt exist in index %d", i)
		}
	}
	_, ok := mp.Get(uint64(1))
	if ok {
		t.Fatalf("Key wasnt deleted")
	}
	_, ok = mp.Delete(uint64(3001))
	if ok {
		t.Fatalf("Falsely returned key existence")
	}
}
