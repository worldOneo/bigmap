package bigmap

import "time"

type passiveExpirationService struct {
	accesses  map[uint64]int64
	Expires   int64
}

// NewPassiveExpirationService creates a new expiration service
// which is working according to ExpirationPolicyPassive.
func NewPassiveExpirationService(expires time.Duration) ExpirationService {
	expSrv := &passiveExpirationService{
		accesses: make(map[uint64]int64),
		Expires:  int64(expires),
	}
	return expSrv
}

func (p *passiveExpirationService) BeforeLock(key uint64, shard *Shard) {
}

func (p *passiveExpirationService) Lock(key uint64, shard *Shard) {
	now := time.Now().UnixNano()
	if now-p.accesses[key] < p.Expires {
		p.accesses[key] = now
		return
	}
	shard.UnsafeDelete(key)
	delete(p.accesses, key)
}

func (p *passiveExpirationService) Access(key uint64, shard *Shard) {
	p.accesses[key] = time.Now().UnixNano()
}

func (p *passiveExpirationService) AfterAccess(key uint64, shard *Shard) {
}

func (p *passiveExpirationService) Remove(key uint64, shard *Shard) {
	delete(p.accesses, key)
}
