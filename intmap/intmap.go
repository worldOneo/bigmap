package intmap

const (
	// Free is the key value of free fields
	Free = 0
	// Phi to scramble keys to prevent bad hashes
	Phi = 0x9E3779B9
)

// KeyType is the type of the keys for this map
type KeyType = uint64

// ValType is the type of the values for this map
type ValType = uint32

// IntMap to store uint64->uint32 relations
type IntMap struct {
	data        []KeyType
	dataSize    KeyType
	capacity    KeyType
	dataMask    KeyType
	capMask     KeyType
	maxCapacity KeyType
	freeSet     bool
	freeVal     ValType
	size        KeyType
}

// New instanciates an new IntMap
func New() *IntMap {
	return &IntMap{
		data:        make([]KeyType, 64),
		dataSize:    64,
		capacity:    32,
		maxCapacity: 24,
		dataMask:    63,
		capMask:     31,
	}
}

// Put adds an item to the int map
func (I *IntMap) Put(key KeyType, val ValType) {
	if key == Free {
		I.freeVal = val
		I.freeSet = true
		return
	}
	index := I.index(key)
	for {
		definedKey := I.data[index]
		if key == definedKey || definedKey == Free {
			if definedKey == Free {
				I.size++
				I.data[index] = key
			}
			I.data[index+1] = KeyType(val)
			break
		}
		index = I.next(index)
	}
	I.expand()
}

// Get retrieves an item from the intmap and returns value, true or
// 0, false if the item isn't in this map.
func (I *IntMap) Get(key KeyType) (ValType, bool) {
	if key == Free {
		if I.freeSet {
			return I.freeVal, true
		}
		return 0, false
	}
	index := I.index(key)
	for {
		definedKey := I.data[index]
		if definedKey == Free {
			return 0, false
		}
		if key == definedKey {
			return ValType(I.data[index+1]), true
		}
		index = I.next(index)
	}
}

// Delete removes a value from this map returns value,true or
// 0, false if the key wasnt in this map
func (I *IntMap) Delete(key KeyType) (ValType, bool) {
	if key == Free {
		if I.freeSet {
			I.freeSet = false
			return I.freeVal, true
		}
		return 0, false
	}
	index := I.index(key)
	for {
		definedKey := I.data[index]
		if definedKey == Free {
			return 0, false
		}
		if key == definedKey {
			data := I.data[index+1]
			I.unshift(index)
			I.size--
			return ValType(data), true
		}
		index = I.next(index)
	}
}

func (I *IntMap) unshift(current KeyType) {
	var key KeyType
	for {
		last := current
		current = I.next(current)
		for {
			key = I.data[current]
			if key == Free {
				I.data[last] = Free
				return
			}
			slot := I.index(key)
			if last <= current {
				if last >= slot || slot > current {
					break
				}
			} else if last >= slot && slot > current {
				break
			}
			current = I.next(current)
		}
		I.data[last] = key
		I.data[last+1] = I.data[current+1]
	}
}

func (I *IntMap) expand() {
	if I.size < I.maxCapacity {
		return
	}

	oldLen := I.dataSize
	oldData := I.data
	I.capacity = I.dataSize
	I.capMask = I.capacity - 1
	I.dataSize *= 2
	I.maxCapacity *= 2
	I.dataMask = I.dataSize - 1
	I.size = 0
	I.data = make([]uint64, I.dataSize)
	for i := KeyType(0); i < oldLen; i += 2 {
		key := oldData[i]
		if key != Free {
			I.Put(key, ValType(oldData[i+1]))
		}
	}
}

func scramble(key KeyType) KeyType {
	hash := key * Phi
	return hash ^ (hash >> 16)
}

func (I *IntMap) next(index KeyType) KeyType {
	return (index + 2) & I.dataMask
}

func (I *IntMap) index(key KeyType) KeyType {
	return (scramble(key) & I.capMask) << 1
}
