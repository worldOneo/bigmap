package bigmap

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
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
	expSrv    *expirationService
}

// NewShard initializes a new shard.
// The capacity is the initial capacity of the shard.
// The entrysize defines the size each entry takes.
// Smaller entries are no problem, but bigger will result in an error.
// Expires defines the time after items can be removed.
// If expires is >= 0 it will be ignored and items wont be removed.
func NewShard(capacity, entrysize uint32, expires time.Duration) *Shard {
	shrd := &Shard{
		ptrs:      make(map[uint64]uint32),
		free:      make([]uint32, 1024),
		size:      0,
		entrysize: entrysize,
		array:     make([]byte, capacity),
		buff:      make([]byte, LengthBytes),
	}
	if expires > 0 {
		shrd.expSrv = newExpirationService(shrd, expires)
	}
	return shrd
}

// Put adds or overwrites an item in(to) the shards internal byte-array.
func (S *Shard) Put(key uint64, val []byte) error {
	lval := uint32(len(val))
	if lval > S.entrysize {
		_lval := lval
		return fmt.Errorf("shard put: value size to long (%d > %d)", _lval, S.entrysize)
	}
	S.hitIfExpires(key)
	S.Lock()
	defer S.Unlock()
	ptr, ok := S.ptrs[key]
	if !ok {
		if S.freecdx < S.freeidx {
			ptr = S.free[S.freecdx]
			S.freecdx++
		} else {
			ptr = S.size
			S.sizeCheck(lval + LengthBytes)
		}
		S.ptrs[key] = ptr
	}
	pp := ptr + LengthBytes
	binary.LittleEndian.PutUint32(S.buff, lval)
	copy(S.array[ptr:pp], S.buff)
	copy(S.array[pp:pp+lval], val)
	S.size += LengthBytes
	S.size += S.entrysize
	return nil
}

// Get retrieves an item from the shards internal byte-array.
// It returns a copy of the corresponding byte slice
// and a boolean if the items was contained if the boolean
// is false the slice will be nil.
func (S *Shard) Get(key uint64) ([]byte, bool) {
	S.hitIfExpires(key)
	S.RLock()
	defer S.RUnlock()
	ptr, ok := S.ptrs[key]
	if !ok {
		return nil, false
	}
	ppl := ptr + LengthBytes
	l := binary.LittleEndian.Uint32(S.array[ptr:])
	dst := make([]byte, l)
	copy(dst, S.array[ppl:ppl+l])
	return dst, true
}

// Delete removes an item from the shard.
// Delete doesnt shrink the size of the byte-array
// nor of the shard.
// It only enables the space to be reused.
func (S *Shard) Delete(key uint64) bool {
	if S.expSrv != nil {
		S.expSrv.remove(key)
	}
	S.Lock()
	defer S.Unlock()
	return S.unsafeDelete(key)
}

// unsafeDelete deltes an object but requires the shard to be locked.
// It will not be locked automagically by this function.
func (S *Shard) unsafeDelete(key uint64) bool {
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

func (S *Shard) hitIfExpires(key uint64) {
	if S.expSrv != nil {
		S.expSrv.hit(key)
	}
}
