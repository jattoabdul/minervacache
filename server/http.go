package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jattoabdul/minervacache/cache"
)

type httpServer struct {
	cache   cache.Cache
	metrics cache.MetricsExporter
	server  *http.Server
}

// Start starts the HTTP server on the given address and port.
// It initializes the server and registers the routes.
func (s *httpServer) Start(addr string, port int) error {
	mux := http.NewServeMux()
	// Register routes with middleware
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /cache/{bucket}/{key}", requireBucketAndKey(s.handleGet)) // takes ?policy=lru&ttl=60s
	mux.HandleFunc("PUT /cache/{bucket}/{key}", requireBucketAndKey(s.handleSet))
	mux.HandleFunc("DELETE /cache/{bucket}/{key}", requireBucketAndKey(s.handleDelete))
	mux.Handle("GET /stats", s.metrics.HTTPHandler())

	addr = fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	s.server = server

	log.Printf("Starting HTTP server on %s", addr)
	return server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server.
func (s *httpServer) Stop() error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(context.Background())
}

// HTTP Middlewares decorator functions that wrap handlers to perform common tasks

// kvHandler is a type for handlers that operate on key-value pairs.
type kvHandler func(bucket, key string, body []byte, opts cache.Options) ([]byte, error)

// requireBucketAndKey is a middleware that ensures the request has valid bucket and key parameters.
func requireBucketAndKey(handler kvHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.PathValue("key")
		if bucket == "" || key == "" {
			http.Error(w, "bucket and key are required", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse options like ttl and policy from the request
		opts, err := cache.ParseOptionsFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid options: %v", err), http.StatusBadRequest)
			return
		}

		result, err := handler(bucket, key, body, opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("operation failed: %v", err), http.StatusInternalServerError)
			return
		}

		// TODO: handle response marshalling to json, setting content type, formatting and status codes based on the operation separately.
		w.Write(result)
	}
}

// HTTP Handlers for cache operations

// handleGet retrieves the value associated with the given key in the bucket.
func (s *httpServer) handleGet(bucket, key string, body []byte, opts cache.Options) ([]byte, error) {
	return s.cache.Get(bucket, key, opts)
}

// handleSet sets the value to the provided key in the given bucket.
func (s *httpServer) handleSet(bucket, key string, body []byte, opts cache.Options) ([]byte, error) {
	return nil, s.cache.Set(bucket, key, body, opts)
}

// handleDelete removes the key and value from the bucket.
func (s *httpServer) handleDelete(bucket, key string, body []byte, opts cache.Options) ([]byte, error) {
	return nil, s.cache.Delete(bucket, key)
}

// handleHealth checks the health of the cache server.
func (s *httpServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//TODO: handle response marshalling to json, setting content type, formatting and correct status code separately.
	w.Write([]byte("OK"))
}
