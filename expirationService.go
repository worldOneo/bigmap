package bigmap

import (
	"sync"
	"time"
)

type expirationService struct {
	sync.Mutex
	accesses  map[uint64]int64
	shard     *Shard
	lastCheck int64
	Expires   int64
}

func newExpirationService(shard *Shard, expires time.Duration) *expirationService {
	expSrv := &expirationService{
		accesses: make(map[uint64]int64),
		shard:    shard,
		Expires:  int64(expires),
	}
	return expSrv
}

func (E *expirationService) expirationCheck() {
	now := time.Now().UnixNano()
	if now-E.lastCheck < E.Expires {
		return
	}
	locked := false
	for k, v := range E.accesses {
		if now-v > E.Expires {
			if !locked {
				locked = true
				E.shard.Lock()
				defer E.shard.Unlock()
			}
			E.shard.unsafeDelete(k)
			delete(E.accesses, k)
		}
	}
	E.lastCheck = now
}

func (E *expirationService) hit(key uint64) {
	E.Lock()
	defer E.Unlock()
	E.expirationCheck()
	E.accesses[key] = time.Now().UnixNano()
}

func (E *expirationService) remove(key uint64) {
	delete(E.accesses, key)
}
