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
		cap = config[0].Capacity
		shardcnt = config[0].Shards
	}

	bm := BigMap{shards: make([]*Shard, shardcnt)}

	for i := 0; i < shardcnt; i++ {
		bm.shards[i] = NewShard(cap, entrysize)
	}
	return bm
}

func FNV64(key string) uint64 {
	var hash uint64 = Offset64
	for _, i := range key {
		hash ^= uint64(i)
		hash *= Prime64
	}
	return hash
}

func (B *BigMap) Put(key string, val []byte) error {
	s, h := B.selectShard(key)
	return s.Put(h, val)
}

func (B *BigMap) Get(key string) ([]byte, bool) {
	s, h := B.selectShard(key)
	return s.Get(h)
}

func (B *BigMap) Delete(key string) bool {
	s, h := B.selectShard(key)
	return s.Delete(h)
}

func (B *BigMap) selectShard(key string) (*Shard, uint64) {
	h := FNV64(key)
	return B.shards[h%uint64(len(B.shards))], h
}
