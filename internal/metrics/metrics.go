package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	CacheOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hpcs_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"},
	)

	CacheSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpcs_cache_size_bytes",
			Help: "Current cache size in bytes",
		},
		[]string{"type"},
	)

	CacheEntries = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hpcs_cache_entries_total",
			Help: "Total number of cache entries",
		},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hpcs_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hpcs_active_connections",
			Help: "Number of active connections",
		},
	)

	TotalConnections = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "hpcs_connections_total",
			Help: "Total number of connections handled",
		},
	)
)

func init() {
	prometheus.MustRegister(
		CacheOperations,
		CacheSize,
		CacheEntries,
		RequestDuration,
		ActiveConnections,
		TotalConnections,
	)
}

func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	return http.ListenAndServe(addr, nil)
}

func RecordCacheOperation(operation, status string) {
	CacheOperations.WithLabelValues(operation, status).Inc()
}

func UpdateCacheSize(size int64) {
	CacheSize.WithLabelValues("used").Set(float64(size))
}

func UpdateCacheEntries(count int) {
	CacheEntries.Set(float64(count))
}

func RecordRequestDuration(operation string, duration time.Duration) {
	RequestDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

func IncrementActiveConnections() {
	ActiveConnections.Inc()
}

func DecrementActiveConnections() {
	ActiveConnections.Dec()
}

func IncrementTotalConnections() {
	TotalConnections.Inc()
}