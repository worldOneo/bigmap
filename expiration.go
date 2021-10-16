package bigmap

import "time"

// ExpirationPolicy determins the way expiration
// is treated within the a shard.
type ExpirationPolicy uint64

const (
	// ExpirationPolicyPassive checks an items
	// expiration on access and if the item is
	// expired nil, false is returned and
	// the item is removed.
	//
	// This policy might be better in terms of
	// performance but an removal of an item is
	// not guaranteed and could therefore lead
	// to a memory leak like behaviour if keys
	// are unique and expired items never removed.
	ExpirationPolicyPassive ExpirationPolicy = iota
	// ExpirationPolicySweep checks for any expired items when
	// the map is accessed and removes any expired
	// item if one is detected.
	//
	// This policy might be better in terms
	// of memory usage as items are guaranteed
	// to be removed after they expired. But it
	// could have a major impact in performance
	// as each item must is checked on access
	// and the shard must be writelocked while
	// beeing cleaned.
	ExpirationPolicySweep
)

// ExpirationFactory is a function which creates can
// create a new ExpirationService given the index of
// the shard
type ExpirationFactory func(shardIndex int) ExpirationService

// Expires creates a new ExpirationFactory based on the
// provided ExpirationPolicy.
func Expires(duration time.Duration, policy ExpirationPolicy) ExpirationFactory {
	return func(shardIndex int) ExpirationService {
		switch policy {
		case ExpirationPolicySweep:
			return NewSweepExpirationService(duration)
		}
		return NewPassiveExpirationService(duration)
	}
}
