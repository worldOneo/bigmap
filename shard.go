package bigmap

import (
	"encoding/binary"
	"fmt"
	"sync"
)

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
}

func NewShard(capacity, entrysize uint32) *Shard {
	return &Shard{
		ptrs:      make(map[uint64]uint32),
		free:      make([]uint32, 1024),
		size:      0,
		entrysize: entrysize,
		array:     make([]byte, capacity),
		buff:      make([]byte, LengthBytes),
	}
}

func (S *Shard) Put(key uint64, val []byte) error {
	lval := uint32(len(val))
	if lval > S.entrysize {
		return fmt.Errorf("shard put: value size to long (%d > %d)", lval, S.entrysize)
	}
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
	}
	pp := ptr + LengthBytes
	binary.LittleEndian.PutUint32(S.buff, lval)
	copy(S.array[ptr:pp], S.buff)
	copy(S.array[pp:pp+lval], val)
	S.ptrs[key] = ptr
	S.size += LengthBytes
	S.size += S.entrysize
	return nil
}

func (S *Shard) Get(key uint64) ([]byte, bool) {
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

func (S *Shard) Delete(key uint64) bool {
	S.Lock()
	defer S.Unlock()
	ptr, ok := S.ptrs[key]
	if ok {
		delete(S.ptrs, key)
		S.free[S.freeidx] = ptr
		S.freeidx++
		lfree := len(S.free)
		if S.freeidx >= lfree {
			var a []uint32
			if len(S.free)-S.freecdx < lfree/2 {
				a = make([]uint32, len(S.free))
			} else {
				a = make([]uint32, len(S.free)*2)
			}
			copy(a, S.free[S.freecdx:])
			S.freeidx -= S.freecdx
			S.freecdx = 0
			S.free = a
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
