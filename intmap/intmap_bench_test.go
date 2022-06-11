package intmap

import (
	"testing"
)

func filledStd(n KeyType) map[uint64]uint64 {
	m := make(map[uint64]uint64)
	for i := KeyType(0); i < n; i++ {
		m[i] = i
	}
	return m
}

func BenchmarkStdMap_Put(b *testing.B) {
	filledStd(uint64(b.N))
}

func BenchmarkIntMap_Put(b *testing.B) {
	filled(uint64(b.N))
}

func BenchmarkStdMap_Get(b *testing.B) {
	m := filledStd(uint64(b.N))
	b.ResetTimer()
	for i := KeyType(0); i < KeyType(b.N); i++ {
		_, _ = m[i]
	}
}

func BenchmarkIntMap_Get(b *testing.B) {
	m := filled(uint64(b.N))
	b.ResetTimer()
	for i := KeyType(0); i < KeyType(b.N); i++ {
		_, _ = m.Get(i)
	}
}

func BenchmarkStdMap_Delete(b *testing.B) {
	m := filledStd(uint64(b.N))
	b.ResetTimer()
	for i := KeyType(0); i < KeyType(b.N); i++ {
		delete(m, i)
	}
}

func BenchmarkIntMap_Delete(b *testing.B) {
	m := filled(uint64(b.N))
	b.ResetTimer()
	for i := KeyType(0); i < KeyType(b.N); i++ {
		m.Delete(i)
	}
}

func BenchmarkIntMap_10_10_80(b *testing.B) {
	m := New(16)
	keys := [128]KeyType{}
	for i := 0; i < len(keys); i++ {
		keys[i] = KeyType(i)
	}
	b.RunParallel(func(pb *testing.PB) {
		i := KeyType(0)
		for pb.Next() {
			if i%10 == 0 {
				r := i / 10
				keys[r%128] = r
				m.Put(r, r)
			} else if i%10 == 5 {
				r := (i - 5) / 10
				m.Delete(keys[r%128])
			} else {
				m.Get(keys[i%128])
			}
			i++
		}
	})
}

func BenchmarkIntMap_Put_Parallel(b *testing.B) {
	m := New(16)
	b.RunParallel(func(pb *testing.PB) {
		i := KeyType(0)
		for pb.Next() {
			i++
			m.Put(i, i)
		}
	})
}

func BenchmarkIntMap_Get_Parallel(b *testing.B) {
	m := New(16)
	for i := KeyType(0); i < KeyType(b.N); i++ {
		m.Put(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := KeyType(0)
		for pb.Next() {
			i++
			m.Get(i)
		}
	})
}

func BenchmarkIntMap_Delete_Parallel(b *testing.B) {
	m := New(16)
	for i := KeyType(0); i < KeyType(b.N); i++ {
		m.Put(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := KeyType(0)
		for pb.Next() {
			i++
			m.Delete(i)
		}
	})

}
