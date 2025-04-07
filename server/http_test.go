package server

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jattoabdul/minervacache/cache"
)

// MockCache implements cache.Cache for testing purposes
type MockCache struct {
	GetFunc    func(bucket, key string, opts cache.Options) ([]byte, error)
	SetFunc    func(bucket, key string, value []byte, opts cache.Options) error
	DeleteFunc func(bucket, key string) error
	StopFunc   func()
}

func (m *MockCache) Get(bucket, key string, opts cache.Options) ([]byte, error) {
	return m.GetFunc(bucket, key, opts)
}

func (m *MockCache) Set(bucket, key string, value []byte, opts cache.Options) error {
	return m.SetFunc(bucket, key, value, opts)
}

func (m *MockCache) Delete(bucket, key string) error {
	return m.DeleteFunc(bucket, key)
}

func (m *MockCache) Stop() {
	m.StopFunc()
}

// MockMetrics implements cache.MetricsExporter for testing
type MockMetrics struct{}

func (m *MockMetrics) RecordHit()      {}
func (m *MockMetrics) RecordMiss()     {}
func (m *MockMetrics) RecordEviction() {}
func (m *MockMetrics) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}
func (m *MockMetrics) CollectMetrics() map[string]float64 { return nil }

func TestHandleGet(t *testing.T) {
	mockCache := &MockCache{
		GetFunc: func(bucket, key string, opts cache.Options) ([]byte, error) {
			if bucket == "test-bucket" && key == "test-key" {
				return []byte("test-value"), nil
			}
			return nil, cache.ErrKeyNotFound
		},
	}

	server := NewHTTPServer(mockCache, &MockMetrics{}).(*httpServer)

	// Test successful get
	_ = httptest.NewRequest("GET", "/cache/test-bucket/test-key", nil)
	_ = httptest.NewRecorder()

	// You'd need to extract the handler logic and test it directly
	// or refactor your middleware to be more testable
	result, err := server.handleGet("test-bucket", "test-key", nil, cache.Options{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-value"), result)

	// Test key not found
	result, err = server.handleGet("test-bucket", "non-existent", nil, cache.Options{})
	assert.Error(t, err, "Expected error for non-existent key")

	if !errors.Is(err, cache.ErrKeyNotFound) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}
