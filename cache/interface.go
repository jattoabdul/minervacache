package cache

import (
	"errors"
	"net/http"
	"time"
)

var (
	ErrCacheFull      = errors.New("cache is full")
	ErrKeyNotFound    = errors.New("key not found")
	ErrKeyExpired     = errors.New("key expired")
	ErrBucketNotFound = errors.New("bucket not found")
	ErrInvalidPolicy  = errors.New("invalid eviction policy")
)

type EvictionPolicy int

const (
	NoEvictionPolicy EvictionPolicy = iota
	OldestEvictionPolicy
	NewestEvictionPolicy
	LRUEvictionPolicy
	MRUEvictionPolicy

	MaxCacheSize = 255 // Maximum number of keys the cache can hold

	DefaultTTL = "0" // Default TTL ("0" means no expiration)
	//DefaultCleanupInterval = 30 * time.Second // Default Cleanup Interval (should we have this?)
)

type Options struct {
	TTL            time.Duration  // Time to live for the cache entries. Default is 0 (no expiration).
	EvictionPolicy EvictionPolicy // Controls how keys should be removed from cache. Options are: Oldest, Newest, LRU(default), MRU
}

// Option function type as specified in the problem
type Option func(o *Options) error

// TODO: Since this is what is in the problem statement, but it seems to be wrong as it doesn't allow modifying the default option.
// 	Clarify if it is to be used, and if so, it should be a pointer to Options or return a new Options.
// type Option func(o Options) error

// OriginalCache interface as specified in the problem.
// TODO: Clarify if this is not a mistake and maybe discuss why I changed the approach to use Options struct
// 	instead of separate options per method.
//type OriginalCache interface {
//	// Set sets the value to the provided key in the given bucket.
//	// Applying any provided options during the operation.
//	// An error is returned if operation fails.
//	Set(bucket string, key string, value []byte, opts ...Option) error
//
//	// Get returns the value associated with the given key in the bucket.
//	// Applying any provided options during the operation.
//	// An error is returned if operation fails.
//	Get(bucket, key string, opts ...Option) ([]byte, error)
//
//	// Delete removes the key and value from the bucket.
//	// Applying any provided options during the operation.
//	// An error is returned if operation fails.
//	Delete(bucket, key string, opts ...Option) error
//}

// Cache interface used in my solution. It is a simplified version of the original one which is implemented by the MinervaCache.
type Cache interface {
	// Set sets the value to the provided key in the given bucket.
	// An error is returned if operation fails.
	Set(bucket string, key string, value []byte, opts Options) error
	// Get returns the value associated with the given key in the bucket.
	// An error is returned if operation fails.
	Get(bucket, key string, opts Options) ([]byte, error)
	// Delete removes the key and value from the bucket. (Do we need the extra opts Options argument here?)
	// An error is returned if operation fails.
	Delete(bucket, key string) error

	// Stats returns statistics about the cache. Should I do this or use prometheus to get performance metrics?
	// Or maybe this just stores the stats in the cache, and then we can use prometheus to get them?
	// Do I even need to add this to the interface?
	//Stats() Stats

	// Stop terminates any background processes and cleans up resources. Should I add this to the interface?
	//Stop()
}

func ParseOptionsFromRequest(r *http.Request) (Options, error) {
	ttl := r.URL.Query().Get("ttl")
	if ttl == "" {
		ttl = DefaultTTL
	}

	ttlCleanupInterval, err := time.ParseDuration(ttl) // See func doc for formats.
	if err != nil {
		return Options{}, err
	}
	if ttlCleanupInterval < 0 {
		return Options{}, errors.New("ttl cannot be negative: " + ttl)
	}

	policy := r.URL.Query().Get("policy")
	if policy == "" {
		policy = "lru" // Default to LRU
	}

	var evictionPolicy EvictionPolicy
	switch policy {
	case "lru":
		evictionPolicy = LRUEvictionPolicy
	case "mru":
		evictionPolicy = MRUEvictionPolicy
	case "oldest":
		evictionPolicy = OldestEvictionPolicy
	case "newest":
		evictionPolicy = NewestEvictionPolicy
	default:
		return Options{}, errors.New("invalid policy: " + policy)
	}

	return Options{
		TTL:            ttlCleanupInterval,
		EvictionPolicy: evictionPolicy,
	}, nil
}
