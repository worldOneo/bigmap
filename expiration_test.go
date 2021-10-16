package bigmap

import (
	"math/rand"
	"testing"
	"time"
)


func TestShardSweepExpiration(t *testing.T) {
	ShardExpiration(t, Expires(time.Second, ExpirationPolicySweep))
}

func TestShardPassiveExpiration(t *testing.T) {
	ShardExpiration(t, Expires(time.Second, ExpirationPolicyPassive))
}

func TestMapSweepExpiration(t *testing.T) {
	ShardExpiration(t, Expires(time.Second, ExpirationPolicySweep))
}

func TestMapPassiveExpiration(t *testing.T) {
	ShardExpiration(t, Expires(time.Second, ExpirationPolicyPassive))
}

func MapExpiration(t *testing.T, factory ExpirationFactory) {
	rand.Seed(time.Now().UnixNano())
	keys := make([][]byte, 4096*8)
	vals := make([][]byte, 4096*8)
	a := GenVal()
	b := GenKey(1)
	for i := range keys {
		keys[i] = b
		vals[i] = a
	}
	bigmap := New(1024, Config{
		ExpirationFactory: factory,
	})
	for i, key := range keys {
		err := bigmap.Put(key, vals[i])
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for _, key := range keys {
		_, ok := bigmap.Get(key)

		if !ok {
			t.Fatalf("Expiration service swooped to early")
		}
	}
	time.Sleep(time.Second * 2)
	for i, key := range keys {
		_, ok := bigmap.Get(key)

		if ok {
			t.Fatalf("Expiration service didn't swoop well enough for key %s (idx: %d)", key, i)
		}
	}
}

func ShardExpiration(t *testing.T, factory ExpirationFactory) {
	rand.Seed(time.Now().UnixNano())
	keys := make([][]byte, 4096*8)
	vals := make([][]byte, 4096*8)
	a := GenVal()
	b := GenKey(1)
	for i := range keys {
		keys[i] = b
		vals[i] = a
	}
	shard := NewShard(1024, 1024, factory(0))
	for i, key := range keys {
		err := shard.Put(FNV64(key), vals[i])
		if err != nil {
			t.Fatalf("shard put: %v", err)
		}
	}

	for _, key := range keys {
		_, ok := shard.Get(FNV64(key))

		if !ok {
			t.Fatalf("Expiration service swooped to early")
		}
	}
	time.Sleep(time.Second * 2)
	for i, key := range keys {
		_, ok := shard.Get(FNV64(key))

		if ok {
			t.Fatalf("Expiration service didn't swoop well enough for key %s (idx: %d)", key, i)
		}
	}
}

