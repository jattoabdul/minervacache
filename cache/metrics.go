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
}

type MetricsExporter interface {
	HTTPHandler() http.Handler
}

// mockMetrics is a no-op implementation of the MetricsHandler interface. For testing purpose.
type mockMetrics struct{}

func (n *mockMetrics) SetSize(size int) {}
func (n *mockMetrics) AddHit()          {}
func (n *mockMetrics) AddMiss()         {}

// PmMetrics is a Prometheus implementation of the MetricsHandler interface.
type PmMetrics struct {
	size *prometheus.GaugeVec
	hit  *prometheus.CounterVec
	miss *prometheus.CounterVec // Can be broken down into more granular metrics. TODO: Add more metrics if time permits.
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

// HTTPHandler returns an HTTP handler for exposing the metrics.
func (pm *PmMetrics) HTTPHandler() http.Handler {
	return promhttp.Handler()
}
