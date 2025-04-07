package cache

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ MetricsHandler = &mockMetrics{}

// MetricsHandler allows MinervaCache to track and report metrics for monitoring.
// We would use Prometheus for actual implementation and do nothing for testing by using the mockMetrics.
type MetricsHandler interface {
	SetSize(size int)
	AddHit()
	AddMiss()
	AddSet()
	AddSetExists()
	AddDelete()
	AddEvict()
	AddExpire(inlineCheck bool)
	AddNotFound()
}

type MetricsExporter interface {
	HTTPHandler() http.Handler
}

// mockMetrics is a no-op implementation of the MetricsHandler interface. For testing purpose.
type mockMetrics struct{}

func (n *mockMetrics) SetSize(size int)           {}
func (n *mockMetrics) AddHit()                    {}
func (n *mockMetrics) AddMiss()                   {}
func (n *mockMetrics) AddSet()                    {}
func (n *mockMetrics) AddSetExists()              {}
func (n *mockMetrics) AddDelete()                 {}
func (n *mockMetrics) AddEvict()                  {}
func (n *mockMetrics) AddExpire(inlineCheck bool) {}
func (n *mockMetrics) AddNotFound()               {}

// PmMetrics is a Prometheus implementation of the MetricsHandler interface.
type PmMetrics struct {
	size      *prometheus.GaugeVec
	hit       *prometheus.CounterVec
	miss      *prometheus.CounterVec // Can be broken down into more granular metrics. Broken down below.
	set       *prometheus.CounterVec
	setExists *prometheus.CounterVec
	delete    *prometheus.CounterVec
	evict     *prometheus.CounterVec
	expire    *prometheus.CounterVec
	notFound  *prometheus.CounterVec
}

// NewPmMetrics creates a new instance of pmMetrics with Prometheus metrics.
// It registers the metrics with the Prometheus registry.
func NewPmMetrics() *PmMetrics {
	pm := &PmMetrics{
		size: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cache_size",
				Help: "Size of the cache",
			},
			nil,
		),
		hit: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hit",
				Help: "Number of cache hits",
			},
			nil,
		),
		miss: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_miss",
				Help: "Number of cache misses",
			},
			nil,
		),
		set: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_set",
				Help: "Number of cache sets",
			},
			nil,
		),
		setExists: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_set_exists",
				Help: "Number of cache sets that already exist",
			},
			nil,
		),
		delete: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_delete",
				Help: "Number of cache deletes",
			},
			nil,
		),
		evict: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_evict",
				Help: "Number of cache evictions",
			},
			nil,
		),
		expire: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_expire",
				Help: "Number of cache expirations",
			},
			[]string{"inline"},
		),
		notFound: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_not_found",
				Help: "Number of cache not found",
			},
			nil,
		),
	}

	prometheus.MustRegister(pm.size, pm.hit, pm.miss)
	return pm
}

// SetSize sets the size of the cache.
func (pm *PmMetrics) SetSize(size int) {
	pm.size.WithLabelValues().Set(float64(size))
}

// AddHit increments the hit counter for the cache.
func (pm *PmMetrics) AddHit() {
	pm.hit.WithLabelValues().Inc()
}

// AddMiss increments the miss counter for the cache.
func (pm *PmMetrics) AddMiss() {
	pm.miss.WithLabelValues().Inc()
}

// AddSet increments the set counter for the cache.
func (pm *PmMetrics) AddSet() {
	pm.size.WithLabelValues().Inc()
}

// AddSetExists increments the set exists counter for the cache.
func (pm *PmMetrics) AddSetExists() {
	pm.size.WithLabelValues().Inc()
}

// AddDelete increments the delete counter for the cache.
func (pm *PmMetrics) AddDelete() {
	pm.size.WithLabelValues().Dec()
}

// AddEvict increments the evict counter for the cache.
func (pm *PmMetrics) AddEvict() {
	pm.size.WithLabelValues().Dec()
}

// AddExpire increments the expire counter for the cache.
func (pm *PmMetrics) AddExpire(lazy bool) {
	pm.size.WithLabelValues().Dec()
}

// AddNotFound increments the not found counter for the cache.
func (pm *PmMetrics) AddNotFound() {
	pm.miss.WithLabelValues().Inc()
}

// HTTPHandler returns an HTTP handler for exposing the metrics.
func (pm *PmMetrics) HTTPHandler() http.Handler {
	return promhttp.Handler()
}
