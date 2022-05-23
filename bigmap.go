package bigmap

const (
	// DefaultCapacity is the default initial capacity of an shard in bytes
	DefaultCapacity uint64 = 1024
	// DefaultShards is the default amount of shards in a BigMap
	DefaultShards int = 32
	// LengthBytes is the amount of bytes required to define the length
	LengthBytes uint64 = 8
	// Offset64 is the offset for FNV64
	Offset64 = 14695981039346656037
	// Prime64 is the prime for FNV64
	Prime64 = 1099511628211
)

// BigMap is a distributed byte-array based map.
// It is made up of multiple shards therefore
// enabling efficient parallel accesses.
//
// The BigMap doesn't have a global lock.
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
type Config struct {
	// Shards is the amount of shards which
	// the map holds. The default is 16 and is
	// sufficient most of the time. Try to benchmark
	// your application to find out the shard amount
	// that fits you best.
	//
	// Default: 16
	Shards int
	// Capacity defines the initial capacity of shard.
	// This doesn't make a big difference in the long run
	// as shards just scale up and stop changing at some
	// time. But if you know your max capacity you can safe
	// some (if only very little) time avoiding the
	// resizing of shards.
	//
	// Default: 1024
	Capacity uint64
	// ExpirationFactory is used to create expirationServices
	// for expiring items. An value of nil will result in no
	// expiration of items.
	//
	// Default: nil
	ExpirationFactory ExpirationFactory
}

// New creates a new BigMap and populates its shards.
//
// The entrysize defines the maximum size of the items added.
// Smaller items are no problem, bigger will return an error.
//
// A config may be provided to tune the map as needed
// and/or enable expiration of items.
// See bigmap.Config
func New(entrysize uint64, config ...Config) BigMap {
	conf := Config{
		Shards:            DefaultShards,
		Capacity:          DefaultCapacity,
		ExpirationFactory: nil,
	}
	if len(config) != 0 {
		firstConf := config[0]
		if firstConf.Capacity != 0 {
			conf.Capacity = firstConf.Capacity
		}
		if firstConf.Shards != 0 {
			conf.Shards = firstConf.Shards
		}
		conf.ExpirationFactory = firstConf.ExpirationFactory
	}

	bm := BigMap{shards: make([]*Shard, conf.Shards)}

	for i := 0; i < conf.Shards; i++ {
		var expirationService ExpirationService = nil
		if conf.ExpirationFactory != nil {
			expirationService = conf.ExpirationFactory(i)
		}
		bm.shards[i] = NewShard(conf.Capacity, entrysize, expirationService)
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

// Get retrieves an item for the key.
// It returns a copy of the corresponding byte slice
// and a boolean if the item was contained. If the boolean
// is false the slice will be nil.
func (B *BigMap) Get(key []byte) ([]byte, bool) {
	s, h := B.SelectShard(key)
	return s.Get(h)
}

// GetCopy returns a copy of the corresponding byte slice
// and a boolean if the item was contained. If the boolean
// is false the slice will be nil.
// The returned byte slice is a copy of the original one.
func (B *BigMap) GetCopy(key []byte) ([]byte, bool) {
	s, h := B.SelectShard(key)
	return s.GetCopy(h)
}

// Delete removes an item from the map.
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
