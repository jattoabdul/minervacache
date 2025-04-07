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
	// metrics is used for tracking cache actions. Only hit, miss and size for now.
	metrics MetricsHandler
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

func NewMinervaCache(capacity int, ttlCheckInterval time.Duration, metrics MetricsHandler) *MinervaCache {
	mc := &MinervaCache{
		capacity:         capacity,
		ttlCheckInterval: ttlCheckInterval,
		stop:             make(chan struct{}),
		buckets:          make(map[string]map[string]*list.Element),
		order:            list.New(),
		metrics:          metrics,
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

		// Update the access time for LRU/MRU policies.
		if opts.EvictionPolicy == LRUEvictionPolicy || opts.EvictionPolicy == MRUEvictionPolicy {
			mc.order.MoveToBack(el) // Move the element to the back of the list since it was accessed.
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
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// The Get method is expected to use the Oldest eviction policy if the cache is full.
	// TODO: Should we really be overriding the eviction policy in the options here when the capacity is full?
	if mc.order.Len() >= mc.capacity {
		mc.evict(OldestEvictionPolicy)
	}

	// Check if the bucket exists
	mcb, ok := mc.buckets[bucket]
	if !ok {
		mc.metrics.AddMiss() // TODO: maybe add a key notFound metric for this specifically?
		return nil, ErrBucketNotFound
	}

	// Check if the key exists in the bucket
	el, ok := mcb[key]
	if !ok {
		mc.metrics.AddMiss() // TODO: maybe add a key notFound metric for this specifically?
		return nil, ErrKeyNotFound
	}

	// Check if the item is expired. This is an inline check for expired items. Always check for expired items in Get.
	item := el.Value.(*cacheItem)
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		mc.deleteAndRemoveFromInsertOrder(el)
		mc.metrics.AddMiss() // TODO: maybe add a key expired metric for this specifically?
		return nil, ErrKeyExpired
	}

	// Update the last access time for LRU/MRU policies.
	if opts.EvictionPolicy == LRUEvictionPolicy || opts.EvictionPolicy == MRUEvictionPolicy {
		mc.order.MoveToBack(el) // Move the element to the back of the list since it was accessed.
	}

	mc.metrics.AddHit() // Track the hit action for metrics.
	return item.value, nil
}

// Delete removes the key and value from the specified bucket. If the bucket is empty, it is deleted.
// An error is returned if the operation fails. (Do we need the extra opts Options argument here?)
func (mc *MinervaCache) Delete(bucket string, key string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Check if the bucket exists
	mcb, ok := mc.buckets[bucket]
	if !ok {
		mc.metrics.AddMiss() // TODO: maybe add a bucket notFound metric for this specifically?
		return ErrBucketNotFound
	}

	// Check if the key exists in the bucket
	el, ok := mcb[key]
	if ok {
		// TODO: Track the delete action for metrics. Need to add this to the metrics handler.

		// Remove the key from the bucket and update insertion order list. Remove bucket if empty as well.
		mc.deleteAndRemoveFromInsertOrder(el)

		return nil
	}

	mc.metrics.AddMiss() // TODO: maybe add a key notFound metric for this specifically?
	return ErrKeyNotFound
}

// evict removes the oldest or newest or lru or mru item from the cache based on the eviction policy.
// It is called when the cache reaches its capacity and needs to evict an item.
// The eviction policy is passed as an argument to determine which item to evict.
// No locking is needed here, as the caller already locks the mutex.
func (mc *MinervaCache) evict(policy EvictionPolicy) {
	var el *list.Element

	switch policy {
	case MRUEvictionPolicy, NewestEvictionPolicy:
		el = mc.order.Back() // MRU or Newest item
	default:
		el = mc.order.Front() // LRU or Oldest item or When no policy is set (None).
	}

	mc.deleteAndRemoveFromInsertOrder(el)
	// TODO: Track the eviction action for metrics. Need to add this to the metrics handler.
}

// deleteAndRemoveFromInsertOrder removes the key from the bucket and updates the insertion order list.
// Used in Delete and evict and must be called with the mutex locked in the caller.
func (mc *MinervaCache) deleteAndRemoveFromInsertOrder(el *list.Element) {
	mc.order.Remove(el)

	item := el.Value.(*cacheItem)
	mcb := mc.buckets[item.bucket]
	delete(mcb, item.key)

	// Check if the bucket is empty after deletion
	if len(mcb) == 0 {
		delete(mc.buckets, item.bucket)
	}
}

// startTTLCheck starts a goroutine that periodically checks for expired items in the cache.
func (mc *MinervaCache) startTTLCheck() {
	if mc.ttlCheckInterval <= 0 {
		return // No TTL check needed
	}

	ticker := time.NewTicker(mc.ttlCheckInterval) // Maybe put this on the mc struct as a pointer to avoid creating a new one every time?
	go func() {
		for {
			select {
			case <-ticker.C:
				mc.metrics.SetSize(mc.order.Len()) // Update the size metric
				mc.checkExpiredItems()
			case <-mc.stop:
				ticker.Stop() // TODO: should I defer this at the top of the routine?
				return
			}
		}
	}()
}

// Stop terminates the TTL check goroutine and cleans up resources. NB: Get action always checks for expired items anyway.
func (mc *MinervaCache) Stop() {
	close(mc.stop) // Stop the TTL check goroutine

	// TODO: Do I really want to do all this below cleanups? Maybe just stop the goroutine and let it clean up?
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Clean up buckets and order list
	for _, mcb := range mc.buckets {
		for _, el := range mcb {
			mc.order.Remove(el)
		}
	}
	mc.buckets = make(map[string]map[string]*list.Element)
	mc.order.Init() // Reset the order list
}

// checkExpiredItems checks for expired items in the cache and removes them.
func (mc *MinervaCache) checkExpiredItems() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Iterate over all buckets and check for expired items.
	// Although this is ran in a separate goroutine, it is still O(b*i). TODO: How to optimize this?
	for _, mcb := range mc.buckets {
		for _, el := range mcb {
			item := el.Value.(*cacheItem)
			if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
				// Item is expired, remove it.
				mc.deleteAndRemoveFromInsertOrder(el)
			}
		}
	}
}

// getBucket returns the bucket for the given key. If the bucket doesn't exist, it creates a new one.
func (mc *MinervaCache) getBucket(bucket string) map[string]*list.Element {
	mcb, ok := mc.buckets[bucket]
	if !ok {
		mcb = make(map[string]*list.Element)
		mc.buckets[bucket] = mcb
	}
	return mcb
}
