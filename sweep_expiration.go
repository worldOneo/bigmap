package bigmap

import (
	"time"
)

type sweepExpirationService struct {
	accesses  map[uint64]int64
	lastCheck int64
	Expires   int64
}

// NewSweepExpirationService creates a new expiration service
// which is working according to ExpirationPolicySweep.
func NewSweepExpirationService(expires time.Duration) ExpirationService {
	expSrv := &sweepExpirationService{
		accesses: make(map[uint64]int64),
		Expires:  int64(expires),
	}
	return expSrv
}

func (p *sweepExpirationService) BeforeLock(key uint64, shard *Shard) {
}

func (p *sweepExpirationService) Lock(key uint64, shard *Shard) {
	now := time.Now().UnixNano()
	if now-p.lastCheck < p.Expires {
		return
	}
	if now-p.lastCheck < p.Expires {
		return
	}
	for itemKey, itemAccessed := range p.accesses {
		if now-itemAccessed > p.Expires {
			shard.UnsafeDelete(itemKey)
			delete(p.accesses, itemKey)
		}
	}
	p.lastCheck = now
}

func (p *sweepExpirationService) Access(key uint64, shard *Shard) {
	p.accesses[key] = time.Now().UnixNano()
}

func (p *sweepExpirationService) AfterAccess(key uint64, shard *Shard) {
}

func (p *sweepExpirationService) Remove(key uint64, shard *Shard) {
	delete(p.accesses, key)
}
