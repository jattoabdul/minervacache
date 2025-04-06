package cache

import (
	"container/list"
	"sync"
	"time"
)

var _ Cache = &MinervaCache{} // MinervaCache implements Cache interface. This is called a compile-time assertion.

// MinervaCache implements a key-value cache with various eviction policies [EvictionPolicy] and TTL support.
// It is designed to be used in a distributed system where multiple processes can access the cache.
type MinervaCache struct {
	capacity         int
	ttlCheckInterval time.Duration
	stop             chan struct{}
	// TODO: stats/metrics tracking.
	// mutex locks all the buckets and the order list in the cache.
	// We could use a RWMutex, but since we are using a single mutex for all operations,
	// we don't need to worry about read/write locks. Especially since we perform write update operations like eviction
	// and usage/insertion order updates in Get operations as well. It would be over-complicated to use a RWMutex
	// and have to Rlock, RUnlock, Lock, and Unlock for every operation that needs both read and writes.
	mutex sync.Mutex
	// buckets is a map of buckets where each bucket is a map of key-value pairs.
	// The value is set in a Value field of a list.Element and stored in the bucket as a pointer to the element in the insertion order list.
	buckets map[string]map[string]*list.Element
	// order is a doubly linked list that maintains the order of keys based on the eviction policy.
	// The order is by default the insertion order. Used to evict the oldest or newest keys.
	// For [EvictionPolicyLRU] or [EvictionPolicyMRU] policies, the order is also updated during Get and Set operations manually.
	order *list.List
}

type cacheItem struct {
	bucket    string
	key       string
	value     []byte
	expiresAt time.Time
}

func NewMinervaCache(capacity int, ttlCheckInterval time.Duration) *MinervaCache {
	mc := &MinervaCache{
		capacity:         capacity,
		ttlCheckInterval: ttlCheckInterval,
		stop:             make(chan struct{}),
		buckets:          make(map[string]map[string]*list.Element),
		order:            list.New(),
		//TODO: metrics handler
	}
	// Start the TTL check (maybe in a separate goroutine?)
	mc.startTTLCheck()

	return mc
}

// Set sets the value for the given key in the specified bucket.
// An error is returned if the operation fails.
func (mc *MinervaCache) Set(bucket string, key string, value []byte, opts Options) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// NB: If we were using options per method, maybe we should apply the options here and use some default values?
	//options := Options{ EvictionPolicy: LRUEvictionPolicy }
	//for _, opt := range opts {
	//	if err := opt(&options); err != nil { return err }
	//}

	// Get or Create bucket if it doesn't exist
	mcb := mc.getBucket(bucket)

	// Create a new bucket item
	expiresAt := time.Time{}
	if opts.TTL > 0 { // If TTL is set, calculate the expiration time.
		expiresAt = time.Now().Add(opts.TTL)
	}

	item := &cacheItem{
		bucket:    bucket,
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}

	// Check if the key already exists
	if el, ok := mcb[key]; ok {
		// Update existing key
		el.Value = item

		if opts.EvictionPolicy == LRUEvictionPolicy || opts.EvictionPolicy == MRUEvictionPolicy {
			mc.order.MoveToBack(el) // Move the element to the back of the list
		}

		// TODO: Track update item action for metrics.

		return nil
	}

	// Evict before inserting new key if the cache is full
	if mc.order.Len() >= mc.capacity {
		// Evict based on policy
		mc.evict(opts.EvictionPolicy)
	}

	// Add the new item to the bucket and update insertion order list
	el := mc.order.PushBack(item)
	mcb[key] = el // Store the element in the bucket map

	return nil
}

// Get retrieves the value for the given key in the specified bucket.
// An error is returned if the operation fails.
func (mc *MinervaCache) Get(bucket string, key string, opts Options) ([]byte, error) {
	return nil, ErrKeyNotFound
}

// Delete removes the key and value from the specified bucket. If the bucket is empty, it is deleted.
// An error is returned if the operation fails. (Do we need the extra opts Options argument here?)
func (mc *MinervaCache) Delete(bucket string, key string) error {
	return ErrKeyNotFound
}

// evict removes the oldest or newest or lru or mru item from the cache based on the eviction policy.
func (mc *MinervaCache) evict(policy EvictionPolicy) {}

// removeFromInsertOrder (del) removes the key from the bucket and updates the insertion order list.
// Used in Delete and evict and must be called with the mutex locked in the caller.
func (mc *MinervaCache) removeFromInsertOrder(el *list.Element) {}

// startTTLCheck starts a goroutine that periodically checks for expired items in the cache.
func (mc *MinervaCache) startTTLCheck() {}

// Stop terminates the TTL check goroutine and cleans up resources. NB: Get action always checks for expired items.
func (mc *MinervaCache) Stop() {}

// checkExpiredItems checks for expired items in the cache and removes them.
func (mc *MinervaCache) checkExpiredItems() {}

// getBucket returns the bucket for the given key. If the bucket doesn't exist, it creates a new one.
func (mc *MinervaCache) getBucket(bucket string) map[string]*list.Element {
	mcb, ok := mc.buckets[bucket]
	if !ok {
		mcb = make(map[string]*list.Element)
		mc.buckets[bucket] = mcb
	}
	return mcb
}
