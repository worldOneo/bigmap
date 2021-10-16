package bigmap

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// Shard is a fraction of a bigmap.
// A bigmap is made up of shards which are
// individuall locked to increase performance.
// A shard locks itself while Put/Delete
// and RLocks itself while Get
type Shard struct {
	sync.RWMutex
	ptrs      map[uint64]uint32
	free      []uint32
	freeidx   int // Position of insertion of free pointers
	freecdx   int // Position of claiming of free pointers
	size      uint32
	entrysize uint32
	array     []byte
	buff      []byte
	expSrv    ExpirationService
}

// NewShard initializes a new shard.
// The capacity is the initial capacity of the shard.
// The entrysize defines the size each entry takes.
// Smaller entries are no problem, but bigger will result in an error.
// Expires defines the time after items can be removed.
// If expires is smaller or equals 0 it will be ignored and
// items wont be removed automatically.
func NewShard(capacity, entrysize uint32, expSrv ExpirationService) *Shard {
	shrd := &Shard{
		ptrs:      make(map[uint64]uint32),
		free:      make([]uint32, 1024),
		size:      0,
		entrysize: entrysize,
		array:     make([]byte, capacity),
		buff:      make([]byte, LengthBytes),
		expSrv:    expSrv,
	}
	return shrd
}

// Put adds or overwrites an item in(to) the shards internal byte-array.
func (S *Shard) Put(key uint64, val []byte) error {
	dataLength := uint32(len(val))
	if dataLength > S.entrysize {
		_lval := dataLength
		return fmt.Errorf("shard put: value size to long (%d > %d)", _lval, S.entrysize)
	}
	S.hitExpirationService(key, ExpirationService.BeforeLock)
	defer func() {
		S.hitExpirationService(key, ExpirationService.Access)
		S.Unlock()
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	S.Lock()
	S.hitExpirationService(key, ExpirationService.Lock)
	ptr, ok := S.ptrs[key]
	if !ok {
		if S.freecdx < S.freeidx {
			ptr = S.free[S.freecdx]
			S.freecdx++
		} else {
			ptr = S.size
			S.sizeCheck(dataLength + LengthBytes)
		}
		S.ptrs[key] = ptr
	}
	dataIndex := ptr + LengthBytes
	binary.LittleEndian.PutUint32(S.buff, dataLength)
	copy(S.array[ptr:dataIndex], S.buff)
	copy(S.array[dataIndex:dataIndex+dataLength], val)
	S.size += LengthBytes
	S.size += S.entrysize
	return nil
}

// Get retrieves an item from the shards internal byte-array.
// It returns a copy of the corresponding byte slice
// and a boolean if the items was contained if the boolean
// is false the slice will be nil.
func (S *Shard) Get(key uint64) ([]byte, bool) {
	S.hitExpirationService(key, ExpirationService.BeforeLock)
	S.RLock()
	defer func() {
		S.hitExpirationService(key, ExpirationService.Access)
		S.RUnlock()
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	S.hitExpirationService(key, ExpirationService.Lock)
	ptr, ok := S.ptrs[key]
	if !ok {
		return nil, false
	}
	dataIndex := ptr + LengthBytes
	dataLength := binary.LittleEndian.Uint32(S.array[ptr:])
	dst := make([]byte, dataLength)
	copy(dst, S.array[dataIndex:dataIndex+dataLength])
	return dst, true
}

// Delete removes an item from the shard.
// And returns true if an item was deleted and
// false if the key didn't exist in the shard.
// Delete doesnt shrink the size of the byte-array
// nor of the shard.
// It only enables the space to be reused.
func (S *Shard) Delete(key uint64) bool {
	S.Lock()
	defer S.Unlock()
	S.hitExpirationService(key, ExpirationService.Remove)
	return S.UnsafeDelete(key)
}

// UnsafeDelete deletes an object without locking the shard.
// If no manual locking is provided data races may occur.
func (S *Shard) UnsafeDelete(key uint64) bool {
	ptr, ok := S.ptrs[key]
	if ok {
		delete(S.ptrs, key)
		S.free[S.freeidx] = ptr
		S.freeidx++
		lfree := len(S.free)
		if S.freeidx >= lfree {
			if len(S.free)-S.freecdx < lfree/2 {
				copy(S.free, S.free[S.freecdx:])
			} else {
				a := make([]uint32, len(S.free)*2)
				copy(a, S.free[S.freecdx:])
				S.free = a
			}
			S.freeidx -= S.freecdx
			S.freecdx = 0
		}
	}
	return ok
}

func (S *Shard) sizeCheck(add uint32) {
	l := uint32(len(S.array))
	for l < S.size+add {
		l *= 2
		b := make([]byte, l)
		copy(b, S.array)
		S.array = b
	}
}

func (S *Shard) hitExpirationService(key uint64, hit func(ExpirationService, uint64, *Shard)) {
	if S.expSrv != nil {
		hit(S.expSrv, key, S)
	}
}
