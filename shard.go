package bigmap

import (
	"encoding/binary"
	"fmt"
	"runtime"

	commoncollections "github.com/worldOneo/CommonCollections"
	"github.com/worldOneo/bigmap/intmap"
)

const maxSpins = 16

type spinner uint8

func (spin *spinner) spin() {
	*spin++
	cnt := *spin
	if cnt > maxSpins {
		cnt = maxSpins
	}
	for i := 0; i < int(cnt); i++ {
		runtime.Gosched()
	}
}

// Shard is a fraction of a bigmap.
// A bigmap is made up of shards which are
// individuall locked to increase performance.
// A shard locks itself while Put/Delete
// and RLocks itself while Get
type Shard struct {
	lock      commoncollections.OptLock
	ptrs      intmap.IntMap
	freePtrs  PointerQueue
	size      uint64
	entrysize uint64
	array     []byte
	expSrv    ExpirationService
}

// NewShard initializes a new shard.
// The capacity is the initial capacity of the shard.
// The entrysize defines the size each entry takes.
// Smaller entries are no problem, but bigger will result in an error.
// Expires defines the time after items can be removed.
// If expires is smaller or equals 0 it will be ignored and
// items wont be removed automatically.
func NewShard(capacity, entrysize uint64, expSrv ExpirationService) *Shard {
	shrd := &Shard{
		lock:      *commoncollections.NewOptLock(),
		ptrs:      intmap.New(),
		freePtrs:  NewPointerQueue(),
		size:      0,
		entrysize: entrysize,
		array:     make([]byte, capacity),
		expSrv:    expSrv,
	}
	return shrd
}

// Put adds or overwrites an item in(to) the shards internal byte-array.
func (S *Shard) Put(key uint64, val []byte) error {
	dataLength := uint64(len(val))
	if dataLength > S.entrysize {
		_lval := dataLength
		maxSize := S.entrysize
		return fmt.Errorf("shard put: value size to long (%d > %d)", _lval, maxSize)
	}
	S.hitExpirationService(key, ExpirationService.BeforeLock)
	defer func() {
		S.hitExpirationService(key, ExpirationService.Access)
		S.lock.Unlock()
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	S.lock.Lock()
	S.hitExpirationService(key, ExpirationService.Lock)
	ptr, ok := S.ptrs.Get(key)
	if !ok {
		ptr, ok = S.freePtrs.Dequeue()
		if !ok {
			ptr = S.size
			S.sizeCheck(S.entrysize + LengthBytes)
		}
		S.ptrs.Put(key, ptr)
	}
	dataIndex := ptr + LengthBytes
	binary.LittleEndian.PutUint64(S.array[ptr:dataIndex], dataLength)
	copy(S.array[dataIndex:dataIndex+dataLength], val)
	S.size += LengthBytes
	S.size += S.entrysize
	return nil
}

// Get retrieves an item from the shards internal byte-array.
// It returns a slice representing the item.
// and a boolean if the items was contained if the boolean
// is false the slice will be nil.
func (S *Shard) Get(key uint64) ([]byte, bool) {
	S.hitExpirationService(key, ExpirationService.BeforeLock)
	defer func() {
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	for {
		spin := spinner(0)
		var check uint32
		var ok bool
		for {
			check, ok = S.lock.RLock()
			if ok {
				break
			}
			spin.spin()
		}
		S.hitExpirationService(key, ExpirationService.Lock)
		ptr, ok := S.ptrs.Get(key)
		if !ok {
			return nil, false
		}
		dataIndex := ptr + LengthBytes
		dataLength := binary.LittleEndian.Uint64(S.array[ptr:])
		if !S.lock.RVerify(check) {
			continue // avoid allocation
		}
		dst := make([]byte, dataLength)
		copy(dst, S.array[dataIndex:dataIndex+dataLength])
		if S.lock.RVerify(check) {
			return dst, true
		}
		runtime.Gosched()
	}
}

// GetInto retrieves an item from the shards internal byte-array
// and writes it into buffer.
// It returns the size, true if the item was contained and 0, false otherwise.
func (S *Shard) GetInto(key uint64, buffer []byte) (uint64, bool) {
	// sync with Get
	S.hitExpirationService(key, ExpirationService.BeforeLock)
	defer func() {
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	for {
		spin := spinner(0)
		var check uint32
		var ok bool
		for {
			check, ok = S.lock.RLock()
			if ok {
				break
			}
			spin.spin()
		}
		S.hitExpirationService(key, ExpirationService.Lock)
		ptr, ok := S.ptrs.Get(key)
		if !ok {
			return 0, false
		}
		dataIndex := ptr + LengthBytes
		dataLength := binary.LittleEndian.Uint64(S.array[ptr:])
		if !S.lock.RVerify(check) {
			continue
		}
		copy(buffer, S.array[dataIndex:dataIndex+dataLength])
		if S.lock.RVerify(check) {
			return dataLength, true
		}
		runtime.Gosched()
	}
}

// Delete removes an item from the shard.
// And returns true if an item was deleted and
// false if the key didn't exist in the shard.
// Delete doesnt shrink the size of the byte-array
// nor of the shard.
// It only enables the space to be reused.
func (S *Shard) Delete(key uint64) bool {
	S.lock.Lock()
	defer S.lock.Unlock()
	S.hitExpirationService(key, ExpirationService.Remove)
	return S.UnsafeDelete(key)
}

// UnsafeDelete deletes an object without locking the shard.
// If no manual locking is provided data races may occur.
func (S *Shard) UnsafeDelete(key uint64) bool {
	ptr, ok := S.ptrs.Delete(key)
	if ok {
		S.freePtrs.Enqueue(ptr)
	}
	return ok
}

func (S *Shard) sizeCheck(add uint64) {
	l := uint64(len(S.array))
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
