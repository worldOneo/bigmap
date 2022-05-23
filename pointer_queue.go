package bigmap

// PointerQueue is an Unbound queue to store free pointers in a shard.
// It is unsafe to access it parallel.
type PointerQueue struct {
	pointers   []uint64
	readIndex  int // Position of reading free pointers
	writeIndex int // Position of writing free pointers
	length     int
}

// NewPointerQueue iniitates a new PointerQueue
func NewPointerQueue() PointerQueue {
	return PointerQueue{
		pointers:   make([]uint64, 128),
		readIndex:  0,
		writeIndex: 0,
		length:     128,
	}
}

// Dequeue returns the next pointer if available and true
// or 0 and false
func (P *PointerQueue) Dequeue() (v uint64, ok bool) {
	if P.readIndex == P.writeIndex {
		return 0, false
	}
	v, ok = P.pointers[P.readIndex], true
	P.readIndex++
	if P.readIndex == P.length {
		P.readIndex = 0
	}
	return
}

// Enqueue pushes a new Pointer to the queue
func (P *PointerQueue) Enqueue(ptr uint64) {
	P.pointers[P.writeIndex] = ptr
	P.writeIndex++
	if P.writeIndex == P.length {
		P.writeIndex = 0
	}
	if P.writeIndex == P.readIndex {
		nLength := P.length * 2
		a := make([]uint64, nLength)

		fromWdxToEnd := P.pointers[P.writeIndex:P.length]
		copy(a[P.length+P.writeIndex:nLength], fromWdxToEnd)

		fromBegToWdx := P.pointers[0:P.writeIndex]
		copy(a, fromBegToWdx)
		P.readIndex += P.length
		P.length = nLength
		P.pointers = a
	}
}
