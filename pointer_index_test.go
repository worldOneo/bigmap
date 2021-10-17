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

func BenchmarkPointerIndex_Put(b *testing.B) {
	mp := NewPointerIndex()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.Put(uint64(i), uint32(i))
	}
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

func BenchmarkPointerIndex_Get(b *testing.B) {
	mp := NewPointerIndex()
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

func TestPointerIndex_Fill(t *testing.T) {
	mp := NewPointerIndex()
	for i := 0; i < 4096*8; i++ {
		mp.Put(uint64(i), uint32(i))
	}
}

func BenchmarkPointerIndex_Delete(b *testing.B) {
	mp := NewPointerIndex()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mp.Delete(uint64(i))
	}
}

func TestPointerIndex_Put(t *testing.T) {
	mp := NewPointerIndex()
	for i := 0; i < 3000; i++ {
		mp.Put(uint64(i), uint32(i))
	}
}

func TestPointerIndex_Get(t *testing.T) {
	mp := NewPointerIndex()
	for i := 0; i < 3000; i++ {
		mp.Put(uint64(i), uint32(i))
	}
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

func TestPointerIndex_Delete(t *testing.T) {
	mp := NewPointerIndex()
	for i := 0; i < 3000; i++ {
		mp.Put(uint64(i), uint32(i))
	}
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
