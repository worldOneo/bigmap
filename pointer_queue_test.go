package bigmap

import "testing"

func BenchmarkPointerQueue_Enqueue(b *testing.B) {
	queue := NewPointerQueue()
	for i := uint32(0); i < uint32(b.N); i++ {
		queue.Enqueue(i)
	}
}

func BenchmarkPointerQueue_Dequeue(b *testing.B) {
	queue := NewPointerQueue()
	for i := uint32(0); i < uint32(b.N); i++ {
		queue.Enqueue(i)
	}
	b.ResetTimer()
	for i := uint32(0); i < uint32(b.N); i++ {
		queue.Dequeue()
	}
}

func TestPointerQueue_Enqueue(t *testing.T) {
	queue := NewPointerQueue()
	for i := uint32(0); i < 3000; i++ {
		queue.Enqueue(i)
	}
}

func TestPointerQueue_Dequeue(t *testing.T) {
	queue := NewPointerQueue()
	for i := uint32(0); i < 3000; i++ {
		queue.Enqueue(i)
	}
	for i := uint32(0); i < 3000; i++ {
		k, ok := queue.Dequeue()
		if k != i || !ok {
			t.Fatalf("Item %d couldn't be dequeued ok=%t v=%d", i, ok, k)
		}
	}
	ptr, ok := queue.Dequeue()
	if ok {
		t.Fatalf("Dequeued empty Queue (%d, %t)", ptr, ok)
	}
}
