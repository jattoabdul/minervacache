package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMinervaCache(t *testing.T) {
	mc := NewMinervaCache(10, 0)
	defer mc.Stop()

	assert.NotNil(t, mc, "expected cache to be non-nil")
}

func TestMinervaCache_Set(t *testing.T) {
	// Test the Set method of MinervaCache.
	mc := NewMinervaCache(10, 0)
	defer mc.Stop()

	// Set a value in the cache.
	err := mc.Set("bkt1", "key1", []byte("val1"), Options{})
	assert.NoError(t, err, "expected no error on Set")

	// Check if the value is set correctly by checking the Cache buckets and keys.
	bucket, ok := mc.buckets["bkt1"]
	assert.True(t, ok, "expected bucket to exist")
	assert.Equal(t, 1, len(bucket), "expected bucket to have 1 key")
}

func TestMinervaCache_Get(t *testing.T) {
	// Test the Get method of MinervaCache.
	mc := NewMinervaCache(10, 0)
	defer mc.Stop()

	// Set a value in the cache.
	err := mc.Set("bkt1", "key1", []byte("val1"), Options{})
	assert.NoError(t, err, "expected no error on Set")

	// Get the value from the cache.
	value, err := mc.Get("bkt1", "key1", Options{})
	assert.NoError(t, err, "expected no error on Get")
	assert.Equal(t, []byte("val1"), value, "expected value to be 'val1'")
}

func TestMinervaCache_Delete(t *testing.T) {
	// Test the Delete method of MinervaCache.
	mc := NewMinervaCache(10, 0)
	defer mc.Stop()

	// Set a value in the cache.
	err := mc.Set("bkt1", "key1", []byte("val1"), Options{})
	assert.NoError(t, err, "expected no error on Set")
	err = mc.Set("bkt1", "key2", []byte("val2"), Options{})
	assert.NoError(t, err, "expected no error on Set")

	// Delete the first value from the cache.
	err = mc.Delete("bkt1", "key1")
	assert.NoError(t, err, "expected no error on Delete")
	// Try to get the deleted value.
	_, err = mc.Get("bkt1", "key1", Options{})
	assert.Error(t, err, "expected error on Get after Delete")

	// Check if the bucket is empty after deletion.
	bucket, ok := mc.buckets["bkt1"]
	assert.True(t, ok, "expected bucket to exist")
	assert.Equal(t, 1, len(bucket), "expected bucket to have 1 key")

	// Delete the second value from the cache.
	err = mc.Delete("bkt1", "key2")
	assert.NoError(t, err, "expected no error on Delete")
	// Try to get the deleted value.
	_, err = mc.Get("bkt1", "key2", Options{})
	assert.Error(t, err, "expected error on Get after Delete")
	// Ensure the bucket was deleted as well because it is empty.
	_, ok = mc.buckets["bkt1"]
	assert.False(t, ok, "expected bucket to be deleted")
}

func TestNoTTL(t *testing.T) {
	mc := NewMinervaCache(10, 0)
	defer mc.Stop()

	err := mc.Set("bkt1", "key1", []byte("val1"), Options{})
	assert.NoError(t, err)

	value, err := mc.Get("bkt1", "key1", Options{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("val1"), value)

	_, err = mc.Get("bkt1", "key2", Options{})
	assert.Error(t, err)
}

func TestTTL(t *testing.T) {
	ttlCheckInterval := 10 * time.Millisecond
	mc := NewMinervaCache(10, ttlCheckInterval)
	defer mc.Stop()

	err := mc.Set("bkt1", "key1", []byte("val1"), Options{
		TTL: 300 * time.Millisecond,
	})
	assert.NoError(t, err)

	time.Sleep(150 * time.Millisecond) // Half of the TTL duration.

	val, err := mc.Get("bkt1", "key1", Options{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("val1"), val)

	time.Sleep(300 * time.Millisecond) // Wait for the TTL to expire.

	_, err = mc.Get("bkt1", "key1", Options{})
	assert.Error(t, err, "expected error after TTL expiration")
}

func TestCapacity(t *testing.T) {
	mc := NewMinervaCache(3, 0)
	defer mc.Stop()
	err := mc.Set("bkt1", "key1", []byte("val1"), Options{})
	assert.NoError(t, err)

	err = mc.Set("bkt1", "key2", []byte("val2"), Options{})
	assert.NoError(t, err)

	err = mc.Set("bkt1", "key3", []byte("val3"), Options{})
	assert.NoError(t, err)

	_, gErr := mc.Get("bkt1", "key1", Options{})
	assert.Error(t, gErr)

	val, gErr := mc.Get("bkt1", "key2", Options{})
	assert.NoError(t, gErr)
	assert.Equal(t, []byte("val2"), val)
}

func TestEviction(t *testing.T) {
	// Test the default eviction policy by filling the cache and checking if the least recently used entry is evicted.
	mc := NewMinervaCache(3, 0)
	defer mc.Stop()

	mc.Set("bkt1", "key1", []byte("val1"), Options{})
	mc.Set("bkt1", "key2", []byte("val2"), Options{})
	mc.Set("bkt1", "key3", []byte("val3"), Options{})

	mc.Set("bkt1", "key4", []byte("val4"), Options{}) // This should evict "key1" as the least recently used key.

	_, err := mc.Get("bkt1", "key1", Options{})
	assert.Error(t, err, "expected error for evicted key")

	val, err := mc.Get("bkt1", "key2", Options{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("val2"), val, "expected key2 to be available")

	mc.Set("bkt1", "key5", []byte("val5"), Options{}) // This should evict "key2" as the least recently used key.

	_, err = mc.Get("bkt1", "key2", Options{})
	assert.Error(t, err, "expected error for evicted key")
}

// TODO: Add more tests for different eviction policies and edge cases.
