package intmap

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	// Free is the key value of free fields
	Free = 0
	// Phi to scramble keys to prevent bad hashes
	Phi = 0x9E3779B9
)

// KeyType is the type of the keys for this map
type KeyType = uint64

// ValType is the type of the values for this map
type ValType = uint64

// IntMap to store uint64->uint32 relations
type IntMap struct {
	current *smallMap
	next    *IntMap
}

// NewIntMap instanciates an new IntMap
func New(itemCount KeyType) IntMap {
	return IntMap{
		current: newSmallMap(KeyType(itemCount)),
		next:    nil,
	}
}

func (intMap *IntMap) Put(key KeyType, val ValType) {
	intMap.putPreassured(key, val, 0)
}

func (intMap *IntMap) putPreassured(key KeyType, val ValType, preassure uint64) {
	smallMap := (*smallMap)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&intMap.current))))
	step, bang := smallMap.put(key, val)
	next := (*IntMap)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&intMap.next))))
	if next != nil {
		next.Put(key, val)
	} else {
		if step {
			intMap.expandStep(smallMap, key, val, preassure)
		}
		if bang {
			intMap.expandBang(smallMap, key, val, preassure)
		}
	}
}

func (intMap *IntMap) expandStep(m *smallMap, key KeyType, val ValType, preassure uint64) {
	intMap.expand(m, 1+preassure, key, val)
}

func (intMap *IntMap) expandBang(m *smallMap, key KeyType, val ValType, preassure uint64) {
	intMap.expand(m, 2+preassure, key, val)
}

func (intMap *IntMap) expand(m *smallMap, f uint64, key KeyType, val ValType) {
	nextPtr := (*unsafe.Pointer)(unsafe.Pointer(&intMap.next))
	next := atomic.LoadPointer(nextPtr)
	if next != nil {
		runtime.Gosched()
		intMap.Put(key, val)
	}
	nextMap := New(KeyType(m.dataSize) * f)
	smap := (unsafe.Pointer)(unsafe.Pointer(&nextMap))
	if !atomic.CompareAndSwapPointer(nextPtr, next, smap) {
		runtime.Gosched()
		intMap.Put(key, val)
	}

	for i := int64(0); i < int64(m.dataSize); i += 2 {
		key := atomic.LoadUint64(&m.data[i])
		if key == Free {
			continue
		}
		val := atomic.SwapUint64(&m.data[i+1], transient)
		if val == tombstone {
			continue
		}
		nextMap.putPreassured(key, val, 1)
	}
	nextMap.current.freeSet = m.freeSet
	nextMap.current.freeVal = m.freeVal
	nextMap.putPreassured(key, val, 1)
	atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&intMap.current)), unsafe.Pointer(nextMap.current))
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&intMap.next)), nil)
}

func (intMap *IntMap) Get(key KeyType) (val ValType, ok bool) {
	smallMap := (*smallMap)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&intMap.current))))
	val, ok = smallMap.get(key)
	return
}

func (intMap *IntMap) Delete(key KeyType) (val ValType, ok bool) {
	val, ok, moved := intMap.current.delete(key)
	if moved {
		runtime.Gosched()
		return intMap.Delete(key)
	}
	return
}

// new instanciates an new IntMap
func newSmallMap(itemCount KeyType) *smallMap {
	return &smallMap{
		data:        make([]KeyType, itemCount),
		dataSize:    itemCount,
		capacity:    itemCount / 2,
		maxCapacity: int64(itemCount)/2 - int64(itemCount)/8,
		dataMask:    itemCount - 1,
		capMask:     itemCount/2 - 1,
		size:        int64(0),
	}
}

const maxSteps = 64

// Put adds an item to the int map
func (small *smallMap) put(key KeyType, val ValType) (overstep bool, sizebang bool) {
	if key == Free {
		small.freeVal = val
		small.freeSet = true
		return false, false
	}
	index := small.index(key)
	for steps := 0; steps < maxSteps; steps++ {
		definedKey := atomic.LoadUint64(&small.data[index])
		if key != definedKey && definedKey != Free {
			index = small.next(index)
			continue
		}
		if definedKey == Free {
			if atomic.CompareAndSwapUint64(&small.data[index], Free, key) {
				atomic.StoreUint64(&small.data[index+1], val)
				return false, atomic.AddInt64(&small.size, 1) > small.maxCapacity
			} else {
				index = small.next(index)
				continue
			}
		}
		if atomic.CompareAndSwapUint64(&small.data[index], definedKey, uint64(val)) {
			atomic.StoreUint64(&small.data[index+1], key)
			return false, atomic.AddInt64(&small.size, 1) > small.maxCapacity
		} else {
			index = small.next(index)
			continue
		}
	}
	overstep = true
	sizebang = false
	return
}

const tombstone = KeyType(^uint64(0))
const transient = KeyType(^uint64(0) - 1)

// Get retrieves an item from the intmap and returns value, true or
// 0, false if the item isn't in this map.
func (small *smallMap) get(key KeyType) (ValType, bool) {
	if key == Free {
		if small.freeSet {
			return small.freeVal, true
		}
		return 0, false
	}
	index := small.index(key)
	for {
		if index >= uint64(len(small.data)) {
			// check for optimistic concurrency
			// if the map is accessed while it is modified
			// it yields invalid results instead of panicking
			return 0, false
		}
		definedKey := atomic.LoadUint64(&small.data[index])
		if definedKey == Free {
			return 0, false
		}
		if key != definedKey {
			index = small.next(index)
			continue
		}
		val := atomic.LoadUint64(&small.data[index+1])
		if val == tombstone || val == transient {
			return 0, false
		}
		return ValType(val), true
	}
}

// Delete removes a value from this map returns value,true or
// 0, false if the key wasnt in this map
func (small *smallMap) delete(key KeyType) (ValType, bool, bool) {
	if key == Free {
		if small.freeSet {
			small.freeSet = false
			return small.freeVal, true, false
		}
		return 0, false, false
	}
	index := small.index(key)
	for {
		definedKey := atomic.LoadUint64(&small.data[index])
		if definedKey == Free {
			return 0, false, false
		}
		if key != definedKey {
			index = small.next(index)
			continue
		}
		data := atomic.LoadUint64(&small.data[index+1])
		if data == tombstone {
			return 0, false, false
		}
		if data == transient {
			return 0, false, true
		}

		atomic.StoreUint64(&small.data[index+1], tombstone)
		atomic.AddInt64(&small.size, -1)
		return ValType(data), true, false
	}
}

func scramble(key KeyType) KeyType {
	hash := key * Phi
	return hash ^ (hash >> 16)
}

func (small *smallMap) next(index KeyType) KeyType {
	return (index + 2) & small.dataMask
}

func (small *smallMap) index(key KeyType) KeyType {
	return (scramble(key) & small.capMask) << 1
}

type smallMap struct {
	data        []KeyType
	dataSize    KeyType
	capacity    KeyType
	dataMask    KeyType
	capMask     KeyType
	maxCapacity int64
	freeSet     bool
	freeVal     ValType
	size        int64
}
