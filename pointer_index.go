package bigmap

// PointerIndex data structure for storing
// the indecies for a shard.
// We can take advantage of the fact that we don't need to shrink it again
type PointerIndex struct {
	data   []uint64
	mask   uint64
	size   uint64
	length int
}

// NewPointerIndex creates a new PointerIndex
func NewPointerIndex() *PointerIndex {
	return &PointerIndex{
		data: make([]uint64, 1024),
		size: 1024,
		mask: 512,
	}
}

// Put stores an pointer in the index.
// This might cause a resize of the index if it is full.
func (P *PointerIndex) Put(key uint64, ptr uint32) {
	index := P.index(key)
	elem := P.data[index]
	if elem != 0 && elem != key {
		P.expand()
		P.Put(key, ptr)
		return
	}
	P.length++
	P.data[index+1] = uint64(ptr)
}

// Get returns the pointer associated to the key and true
// or an invalid value and false if the key doesnt exist in
// the index
func (P *PointerIndex) Get(key uint64) (uint32, bool) {
	index := P.index(key)
	elem := P.data[index]
	return uint32(P.data[index+1]), elem == key
}

// Delete removes an item from the index returns
// if the key existed in the index
func (P *PointerIndex) Delete(key uint64) (uint32, bool) {
	index := P.index(key)
	elem := P.data[index]
	val := P.data[index+1]
	if elem != 0 && elem != key {
		return 0, false
	}
	P.data[index] = 0
	return uint32(val), true
}

func (P *PointerIndex) expand() {
	P.mask = P.size
	P.size *= 2
	data := make([]uint64, P.size)

	oldData := P.data
	P.data = data
	for i := 0; i < len(oldData); i += 2 {
		P.Put(oldData[i], uint32(oldData[i+1]))
	}
}

func (P *PointerIndex) index(key uint64) uint64 {
	return uint64(P.mask-1) & key
}
