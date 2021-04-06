package bigmap

const (
	DefaultCapacity uint32 = 1024
	DefaultShards   int    = 16
	LengthBytes     uint32 = 4
	Offset64               = 14695981039346656037
	Prime64                = 1099511628211
)

type BigMap struct {
	shards []*Shard
}

type Config struct {
	Shards   int
	Capacity uint32
}

func New(entrysize uint32, config ...Config) BigMap {
	cap := DefaultCapacity
	shardcnt := DefaultShards
	if len(config) != 0 {
		ncap := config[0].Capacity
		if ncap != 0 {
			cap = ncap
		}
		nshards := config[0].Shards
		if nshards != 0 {
			shardcnt = nshards
		}
	}

	bm := BigMap{shards: make([]*Shard, shardcnt)}

	for i := 0; i < shardcnt; i++ {
		bm.shards[i] = NewShard(cap, entrysize)
	}
	return bm
}

func FNV64(key []byte) uint64 {
	var hash uint64 = Offset64
	l := len(key)
	for i := 0; i < l; i++ {
		hash ^= uint64(key[i])
		hash *= Prime64
	}
	return hash
}

func (B *BigMap) Put(key []byte, val []byte) error {
	s, h := B.SelectShard(key)
	return s.Put(h, val)
}

func (B *BigMap) Get(key []byte) ([]byte, bool) {
	s, h := B.SelectShard(key)
	return s.Get(h)
}

func (B *BigMap) Delete(key []byte) bool {
	s, h := B.SelectShard(key)
	return s.Delete(h)
}

func (B *BigMap) SelectShard(key []byte) (*Shard, uint64) {
	h := FNV64(key)
	return B.shards[h%uint64(len(B.shards))], h
}
