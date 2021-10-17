package bigmap

const freeElem = 0

// FastMap for fast non synchronized mapping.
// We can take advantage of the fact that we don't need to shrink it again
type FastMap struct {
	data       []uint64
	isFreeElem bool
	freeKey    uint64
	freeValue  uint64
	mask       uint64
	size       uint64
	length     int
}

// NewFastMap creates a new FastMap
func NewFastMap() *FastMap {
	return &FastMap{
		data: make([]uint64, 1024),
		size: 1024,
		mask: 512,
	}
}

// Put stores an pointer in the index.
// This might cause a resize of the index if it is full.
func (P *FastMap) Put(key uint64, ptr uint32) {
	index := P.index(key)
	if index == freeElem {
		if key != P.freeKey {
			P.expand()
			P.Put(key, ptr)
			return
		}
		P.freeKey = key
		P.isFreeElem = true
		P.freeValue = uint64(ptr)
		return
	}
	elem := P.data[index]
	if elem != 0 && elem != key {
		P.expand()
		P.Put(key, ptr)
		return
	}
	P.data[index] = key
	P.data[index+1] = uint64(ptr)
}

// Get returns the pointer associated to the key and true
// or an invalid value and false if the key doesnt exist in
// the index
func (P *FastMap) Get(key uint64) (uint32, bool) {
	index := P.index(key)
	if index == freeElem {
		if P.isFreeElem && P.freeKey == key {
			return uint32(P.freeValue), true
		}
		return 0, false
	}
	elem := P.data[index]
	return uint32(P.data[index+1]), elem == key
}

// Delete removes an item from the index returns
// if the key existed in the index
func (P *FastMap) Delete(key uint64) (uint32, bool) {
	index := P.index(key)
	if index == freeElem {
		if P.isFreeElem && P.freeKey == key {
			P.isFreeElem = false
			return uint32(P.freeValue), true
		}
		return 0, false
	}
	elem := P.data[index]
	if elem == 0 || elem != key {
		return 0, false
	}
	P.data[index] = 0
	return uint32(P.data[index+1]), true
}

func (P *FastMap) expand() {
	P.mask = P.size
	P.size *= 2
	oldData := P.data
	P.data = make([]uint64, P.size)
	for i := 0; i < len(oldData); i += 2 {
		P.Put(oldData[i], uint32(oldData[i+1]))
	}
}

func (P *FastMap) index(key uint64) uint64 {
	return (uint64(P.mask-1) & key) * 2
}
