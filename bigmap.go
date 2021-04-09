package bigmap

import (
	"time"
)

const (
	// DefaultCapacity is the default initial capacity of an shard in bytes
	DefaultCapacity uint32 = 1024
	// DefaultShards is the default amount of shards in a BigMap
	DefaultShards int = 16
	// LengthBytes is the amount of bytes required to define the length
	LengthBytes uint32 = 4
	// Offset64 is the offset for FNV64
	Offset64 = 14695981039346656037
	// Prime64 is the prime for FNV64
	Prime64 = 1099511628211
)

// BigMap is a distributed byte-array based map.
// It is made up of multiple shards therefore
// enabling efficient parallel accesses.
//
// The BigMap doesn't has a global lock.
// Shards are only locked individually.
//
// By default is it split into 16 shards
// each shard holding a 1KB (by default) byte-array
// the shards will double there size if they run out of space
// and will never shrink again. This enables
// the map to stay fast even with many accesses.
type BigMap struct {
	shards []*Shard
}

// Config defines values for a BigMap.
// Values which are 0 will become the default values.
//
// Shards is the amount of shards which
// the map holds. The default is 16 and is
// sufficient most of the time. Try to benchmark
// your application to find out the shard amount
// that fits you best.
//
// Capacity defines the initial capacity of shard.
// This doesn't make a big difference in the long run
// as shards just scale up and stop changing at some
// time. But if you know your max capacity you can safe
// some (if only very little) time avoiding the
// resizing of shards.
//
// Expires defines the time (in ns) after items *can* be
// removed. There is no guarantee that the item will be
// removed exactly as the cleaner is only called after a
// multiple of expires. For example: If an items expire after 2
// seconds an item added at second 1 it will be removed in the
// cleaning cycle of second 4
type Config struct {
	Shards   int
	Capacity uint32
	Expires  time.Duration
}

// New creates a new BigMap and populates its shards.
//
// The entrysize defines the maximum size of the items added.
// Smaller items are no problem, bigger will return an error.
//
// A config may be provided to tune the map as needed
// and/or enable expiration of items.
// See bigmap.Config
func New(entrysize uint32, config ...Config) BigMap {
	cap := DefaultCapacity
	shardcnt := DefaultShards
	var expires time.Duration = 0
	if len(config) != 0 {
		conf := config[0]
		ncap := conf.Capacity
		if ncap != 0 {
			cap = ncap
		}
		nshards := conf.Shards
		if nshards != 0 {
			shardcnt = nshards
		}
		nexpires := conf.Expires
		if nexpires != 0 {
			expires = nexpires
		}
	}

	bm := BigMap{shards: make([]*Shard, shardcnt)}

	for i := 0; i < shardcnt; i++ {
		bm.shards[i] = NewShard(cap, entrysize, expires)
	}
	return bm
}

// FNV64 hashes the byte-array with the FNV64 algorithm.
//
// This function is very performant and takes for a key size
// of 10 often only around 5 ns on modern hardware.
func FNV64(key []byte) uint64 {
	var hash uint64 = Offset64
	l := len(key)
	for i := 0; i < l; i++ {
		hash ^= uint64(key[i])
		hash *= Prime64
	}
	return hash
}

// Put puts an item into the map by putting
// it into corresponding shard of the key.
//
// An error is returned if the shard corresponding
// to the key returns an error. This happens if
// the item is to big.
func (B *BigMap) Put(key []byte, val []byte) error {
	s, h := B.SelectShard(key)
	return s.Put(h, val)
}

// Get retrieves an item from the corresponding shard for the key.
// It returns a copy of the corresponding byte slice
// and a boolean if the items was contained. If the boolean
// is false the slice will be nil.
func (B *BigMap) Get(key []byte) ([]byte, bool) {
	s, h := B.SelectShard(key)
	return s.Get(h)
}

// Delete removes an item from the corresponding shard for the key.
// Delete doesnt shrink the memory size of the map.
// It only enables the space to be reused.
func (B *BigMap) Delete(key []byte) bool {
	s, h := B.SelectShard(key)
	return s.Delete(h)
}

// SelectShard return the corresponding shard to the given key.
func (B *BigMap) SelectShard(key []byte) (*Shard, uint64) {
	h := FNV64(key)
	return B.shards[h%uint64(len(B.shards))], h
}
