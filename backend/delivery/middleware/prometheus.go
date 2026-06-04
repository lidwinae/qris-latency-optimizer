package middleware

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Registry = prometheus.NewRegistry()

	appMetrics = promauto.With(Registry)

	httpRequestsTotal = appMetrics.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "network_mode"},
	)

	httpRequestDuration = appMetrics.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "network_mode"},
	)

	clientRequestDuration = appMetrics.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "client_request_duration_seconds",
			Help:    "Duration of HTTP requests from the client's perspective",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "network_mode"},
	)

	transactionsCreatedTotal = appMetrics.NewCounter(
		prometheus.CounterOpts{
			Name: "transactions_created_total",
			Help: "Total number of created transactions",
		},
	)

	paymentConfirmationsTotal = appMetrics.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_confirmations_total",
			Help: "Total number of payment confirmation requests",
		},
		[]string{"mode", "result"},
	)

	paymentConfirmationDuration = appMetrics.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_confirmation_duration_seconds",
			Help:    "Duration of payment confirmation requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"mode"},
	)

	paymentWorkerProcessedTotal = appMetrics.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_worker_processed_total",
			Help: "Total number of payment confirmation messages processed by the worker",
		},
		[]string{"result"},
	)

	paymentWorkerDuration = appMetrics.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "payment_worker_duration_seconds",
			Help:    "Duration of payment worker processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	cacheLookupTotal = appMetrics.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_lookup_total",
			Help: "Total number of cache lookups",
		},
		[]string{"type", "result"},
	)

	cacheWriteTotal = appMetrics.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_write_total",
			Help: "Total number of cache writes",
		},
		[]string{"type", "result"},
	)
)

func init() {
	// Register default Go process and runtime collectors so that
	// process_cpu_seconds_total, process_resident_memory_bytes,
	// go_goroutines, go_gc_duration_seconds, etc. are exposed.
	Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	Registry.MustRegister(prometheus.NewGoCollector())
}

func NetworkModeFromHost(host string) string {
	if host == "" {
		return "unknown"
	}

	_, port, err := net.SplitHostPort(host)
	if err != nil {
		parts := strings.Split(host, ":")
		if len(parts) > 1 {
			port = parts[len(parts)-1]
		}
	}

	switch port {
	case "8080":
		return "normal"
	case "8081":
		return "rural"
	default:
		return "unknown"
	}
}

func RecordClientLatency(method, path, networkMode string, durationSeconds float64) {
	clientRequestDuration.WithLabelValues(method, path, networkMode).Observe(durationSeconds)
}

func RecordTransactionCreated() {
	transactionsCreatedTotal.Inc()
}

func RecordPaymentConfirmation(mode, result string, durationSeconds float64) {
	paymentConfirmationsTotal.WithLabelValues(mode, result).Inc()
	paymentConfirmationDuration.WithLabelValues(mode).Observe(durationSeconds)
}

func RecordPaymentWorkerProcessed(result string, durationSeconds float64) {
	paymentWorkerProcessedTotal.WithLabelValues(result).Inc()
	paymentWorkerDuration.Observe(durationSeconds)
}

func RecordCacheLookup(cacheType, result string) {
	cacheLookupTotal.WithLabelValues(cacheType, result).Inc()
}

func RecordCacheWrite(cacheType, result string) {
	cacheWriteTotal.WithLabelValues(cacheType, result).Inc()
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/metrics" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		// Process request
		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Use c.FullPath() so that dynamic routes like /api/transactions/:id are grouped
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		networkMode := NetworkModeFromHost(c.Request.Host)

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status, networkMode).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path, networkMode).Observe(duration)
	}
}
