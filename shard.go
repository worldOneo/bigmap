package bigmap

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"

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

type storage struct {
	array []byte
	lock  []commoncollections.OptLock
}

// Shard is a fraction of a bigmap.
// A bigmap is made up of shards which are
// individuall locked to increase performance.
// A shard locks itself while Put/Delete
// and RLocks itself while Get
type Shard struct {
	lock      commoncollections.OptLock
	ptrs      intmap.IntMap
	storage   *storage
	freePtrs  PointerQueue
	size      uint64
	entrysize uint64
	expSrv    ExpirationService
}

// NewShard initializes a new shard.
// The capacity is the initial capacity of the shard.
// The entrysize defines the size each entry takes.
// Smaller entries are no problem, but bigger will result in an error.
// Expires defines the time after items can be removed.
// If expires is smaller or equals 0 it will be ignored and
// items wont be removed automatically.
func NewShard(elementCount, entrysize uint64, expSrv ExpirationService) *Shard {
	data := make([]byte, elementCount*(entrysize+LengthBytes))
	storage := &storage{
		array: data,
		lock:  make([]commoncollections.OptLock, elementCount),
	}
	shrd := &Shard{
		ptrs:      intmap.New(64),
		freePtrs:  NewPointerQueue(),
		size:      0,
		entrysize: entrysize,
		storage:   storage,
		expSrv:    expSrv,
	}
	return shrd
}

func (S *Shard) loadArray(ptr uint64) ([]byte, *commoncollections.OptLock) {
	data := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&S.storage)))
	storage := (*storage)(data)
	return storage.array, &storage.lock[ptr]
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
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	S.hitExpirationService(key, ExpirationService.Lock)
retry:
	verify, ok := S.lock.RLock()
	if !ok {
		runtime.Gosched()
		goto retry
	}
	ptr, ok := S.ptrs.Get(key)
	if !ok {
		ptr, ok = S.freePtrs.Dequeue()
		if !ok {
			ptr = atomic.AddUint64(&S.size, 1)
			S.sizeCheck(ptr)
		}
		S.ptrs.Put(key, ptr)
	}
	array, lock := S.loadArray(ptr)
	ptr *= (S.entrysize + LengthBytes)
	dataIndex := ptr + LengthBytes
	lock.Lock()
	binary.LittleEndian.PutUint64(array[ptr:dataIndex], dataLength)
	copy(array[dataIndex:dataIndex+dataLength], val)
	lock.Unlock()
	if !S.lock.RVerify(verify) {
		runtime.Gosched()
		goto retry
	}
	return nil
}

// Get retrieves an item from the shards internal byte-array.
// It returns a slice representing the item.
// and a boolean if the items was contained if the boolean
// is false the slice will be nil.
//
// The return value is mapped to the underlying byte-array.
// If you want to change it use GetCopy.
func (S *Shard) Get(key uint64) ([]byte, bool) {
	S.hitExpirationService(key, ExpirationService.BeforeLock)
	defer func() {
		S.hitExpirationService(key, ExpirationService.AfterAccess)
	}()
	for {
		S.hitExpirationService(key, ExpirationService.Lock)
		ptr, ok := S.ptrs.Get(key)
		if !ok {
			return nil, false
		}
		array, lock := S.loadArray(ptr)
		ptr *= (S.entrysize + LengthBytes)
		spin := spinner(0)
		var check uint32
		for {
			check, ok = lock.RLock()
			if ok {
				break
			}
			spin.spin()
		}
		dataIndex := ptr + LengthBytes
		dataLength := binary.LittleEndian.Uint64(array[ptr:])
		if !lock.RVerify(check) {
			continue // avoid allocation
		}
		dst := make([]byte, dataLength)
		copy(dst, array[dataIndex:dataIndex+dataLength])
		if lock.RVerify(check) {
			return dst, true
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

func (S *Shard) sizeCheck(ptr uint64) {
	array, _ := S.loadArray(0)
	l := uint64(len(array))
	desiredLength := ptr * (S.entrysize + LengthBytes)
	if desiredLength >= l {
		var b []byte
		S.lock.Lock()
		defer S.lock.Unlock()
		l := uint64(len(array))
		for desiredLength >= l {
			l *= 2
			b = make([]byte, l)
			copy(b, array)
		}
		locks := make([]commoncollections.OptLock, l/(S.entrysize+LengthBytes))
		storage := &storage{
			array: b,
			lock:  locks,
		}
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&S.storage)), unsafe.Pointer(storage))
	}
}

func (S *Shard) hitExpirationService(key uint64, hit func(ExpirationService, uint64, *Shard)) {
	if S.expSrv != nil {
		hit(S.expSrv, key, S)
	}
}
