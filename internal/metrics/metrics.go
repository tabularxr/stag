package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "stag_http_request_duration_seconds",
			Help: "Duration of HTTP requests in seconds",
		},
		[]string{"method", "path", "status"},
	)

	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stag_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	DatabaseOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stag_database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "collection", "status"},
	)

	DatabaseDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "stag_database_operation_duration_seconds",
			Help: "Duration of database operations in seconds",
		},
		[]string{"operation", "collection"},
	)

	ActiveWebsocketConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stag_websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
		[]string{"session_id"},
	)

	MeshProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "stag_mesh_processing_duration_seconds",
			Help: "Duration of mesh processing operations in seconds",
		},
	)

	StorageUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stag_storage_usage_bytes",
			Help: "Storage usage in bytes",
		},
		[]string{"type"},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stag_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)

	CompressionRatio = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "stag_compression_ratio",
			Help:    "Compression ratio achieved for mesh data",
			Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 0.99},
		},
		[]string{"compression_level"},
	)

	CompressionOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stag_compression_operations_total",
			Help: "Total number of compression/decompression operations",
		},
		[]string{"operation", "status"},
	)

	MeshSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "stag_mesh_size_bytes",
			Help:    "Size of mesh data in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 2, 20),
		},
		[]string{"type"},
	)
)