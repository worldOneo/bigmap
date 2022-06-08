package bigmap

// ExpirationService is the interface used for expiring items within a shard
type ExpirationService interface {
	// BeforeLock is called before the shard was accessed (put or get
	// and before the shard is locked with the key which is about to
	// be accessed.
	// Accessing the shard from this method can not cause
	// any deadlock between the shard and the ExpirationService.
	BeforeLock(key uint64, shard *Shard)
	// Lock is called before the shard was accessed (put or get
	// and after the shard is locked with the key which is about to
	// be accessed.
	//
	//Accessing the shard from this method might cause a deadlock.
	Lock(key uint64, shard *Shard)
	// Access is called after something was the was inserted into the shard
	// (put) and before it is unlocked.
	//
	// Accessing the shard from this method
	// might cause a deadlock.
	//
	// If an item should be removed in the lock shard.UnsafeDelete
	// is safe inside this function call.
	Access(key uint64, shard *Shard)
	// AfterAccess is called after the shard was accessed (put or get) after
	// unlocking it.
	// Accessing the shard from this method can not cause
	// any deadlock between the shard and the ExpirationService.
	AfterAccess(key uint64, shard *Shard)
	// Remove is called before the shard is changed (the item associated
	// with the key will be removed) and after it was locked.
	// Accessing the shard from this method might cause a deadlock.
	Remove(key uint64, shard *Shard)
}
