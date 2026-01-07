package observability

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsExporter provides Prometheus-compatible metrics export via HTTP.
type MetricsExporter struct {
	mu     sync.RWMutex
	addr   string
	server *http.Server
	logger *logrus.Logger

	// Registered metrics sources
	perfMonitor   PerformanceMonitor
	networkServer NetworkServer
	world         World

	// Readiness checkers
	readinessCheckers []ReadinessChecker

	// Internal state
	startTime time.Time
}

// PerformanceMonitor provides performance metrics.
type PerformanceMonitor interface {
	GetFPS() float64
	GetFrameTime() float64
	GetMemoryUsageMB() uint64
}

// NetworkServer provides network metrics.
type NetworkServer interface {
	GetConnectedClients() int
	GetTotalBytesSent() uint64
	GetTotalBytesReceived() uint64
	GetPacketsSent() uint64
	GetPacketsReceived() uint64
}

// World provides game state metrics.
type World interface {
	GetEntityCount() int
	GetActiveQuestCount() int
	GetTradeVolume() uint64
}

// ReadinessChecker provides readiness check functionality.
// Implement this interface to add custom readiness checks.
type ReadinessChecker interface {
	// Check returns an error if the component is not ready, nil otherwise.
	// The component name is used for logging and status reporting.
	Check() (componentName string, err error)
}

// NewMetricsExporter creates a new metrics exporter listening on the given address.
// Address should be in format ":port" or "host:port".
func NewMetricsExporter(addr string) *MetricsExporter {
	return NewMetricsExporterWithLogger(addr, logrus.New())
}

// NewMetricsExporterWithLogger creates a new metrics exporter with a custom logger.
func NewMetricsExporterWithLogger(addr string, logger *logrus.Logger) *MetricsExporter {
	return &MetricsExporter{
		addr:      addr,
		logger:    logger,
		startTime: time.Now(),
	}
}

// RegisterPerformanceMonitor registers a performance monitor for metrics export.
func (m *MetricsExporter) RegisterPerformanceMonitor(pm PerformanceMonitor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.perfMonitor = pm
}

// RegisterNetworkServer registers a network server for metrics export.
func (m *MetricsExporter) RegisterNetworkServer(ns NetworkServer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.networkServer = ns
}

// RegisterWorld registers a game world for metrics export.
func (m *MetricsExporter) RegisterWorld(w World) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.world = w
}

// RegisterReadinessChecker registers a readiness checker for the /ready endpoint.
// Multiple checkers can be registered and all will be executed during readiness checks.
func (m *MetricsExporter) RegisterReadinessChecker(checker ReadinessChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readinessCheckers = append(m.readinessCheckers, checker)
}

// Start begins serving metrics on the configured HTTP endpoint.
func (m *MetricsExporter) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil {
		return fmt.Errorf("metrics exporter already started")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", m.handleMetrics)
	mux.HandleFunc("/health", m.handleHealth)
	mux.HandleFunc("/ready", m.handleReady)
	mux.HandleFunc("/status", m.handleStatus)

	m.server = &http.Server{
		Addr:         m.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		m.logger.WithField("address", m.addr).Info("Starting metrics HTTP server")
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.WithError(err).Error("Metrics server error")
		}
	}()

	return nil
}

// Stop gracefully shuts down the metrics HTTP server.
func (m *MetricsExporter) Stop() error {
	return m.StopWithTimeout(30 * time.Second)
}

// StopWithTimeout gracefully shuts down the metrics HTTP server with a custom timeout.
func (m *MetricsExporter) StopWithTimeout(timeout time.Duration) error {
	m.mu.Lock()
	server := m.server
	m.mu.Unlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	m.logger.Info("Stopping metrics HTTP server")
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown metrics server: %w", err)
	}

	m.mu.Lock()
	m.server = nil
	m.mu.Unlock()

	return nil
}

// handleMetrics serves Prometheus-compatible metrics in text format.
func (m *MetricsExporter) handleMetrics(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// Write metrics in Prometheus exposition format
	fmt.Fprintf(w, "# HELP venture_uptime_seconds Server uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE venture_uptime_seconds counter\n")
	fmt.Fprintf(w, "venture_uptime_seconds %.2f\n", time.Since(m.startTime).Seconds())

	// Performance metrics
	if m.perfMonitor != nil {
		fps := m.perfMonitor.GetFPS()
		frameTime := m.perfMonitor.GetFrameTime()
		memoryMB := m.perfMonitor.GetMemoryUsageMB()

		fmt.Fprintf(w, "# HELP venture_fps Current frames per second\n")
		fmt.Fprintf(w, "# TYPE venture_fps gauge\n")
		fmt.Fprintf(w, "venture_fps %.2f\n", fps)

		fmt.Fprintf(w, "# HELP venture_frame_time_ms Frame time in milliseconds\n")
		fmt.Fprintf(w, "# TYPE venture_frame_time_ms gauge\n")
		fmt.Fprintf(w, "venture_frame_time_ms %.2f\n", frameTime)

		fmt.Fprintf(w, "# HELP venture_memory_usage_mb Memory usage in megabytes\n")
		fmt.Fprintf(w, "# TYPE venture_memory_usage_mb gauge\n")
		fmt.Fprintf(w, "venture_memory_usage_mb %d\n", memoryMB)
	}

	// Network metrics
	if m.networkServer != nil {
		clients := m.networkServer.GetConnectedClients()
		bytesSent := m.networkServer.GetTotalBytesSent()
		bytesRecv := m.networkServer.GetTotalBytesReceived()
		packetsSent := m.networkServer.GetPacketsSent()
		packetsRecv := m.networkServer.GetPacketsReceived()

		fmt.Fprintf(w, "# HELP venture_players_connected Number of connected players\n")
		fmt.Fprintf(w, "# TYPE venture_players_connected gauge\n")
		fmt.Fprintf(w, "venture_players_connected %d\n", clients)

		fmt.Fprintf(w, "# HELP venture_network_bytes_sent_total Total bytes sent\n")
		fmt.Fprintf(w, "# TYPE venture_network_bytes_sent_total counter\n")
		fmt.Fprintf(w, "venture_network_bytes_sent_total %d\n", bytesSent)

		fmt.Fprintf(w, "# HELP venture_network_bytes_received_total Total bytes received\n")
		fmt.Fprintf(w, "# TYPE venture_network_bytes_received_total counter\n")
		fmt.Fprintf(w, "venture_network_bytes_received_total %d\n", bytesRecv)

		fmt.Fprintf(w, "# HELP venture_network_packets_sent_total Total packets sent\n")
		fmt.Fprintf(w, "# TYPE venture_network_packets_sent_total counter\n")
		fmt.Fprintf(w, "venture_network_packets_sent_total %d\n", packetsSent)

		fmt.Fprintf(w, "# HELP venture_network_packets_received_total Total packets received\n")
		fmt.Fprintf(w, "# TYPE venture_network_packets_received_total counter\n")
		fmt.Fprintf(w, "venture_network_packets_received_total %d\n", packetsRecv)
	}

	// Game state metrics
	if m.world != nil {
		entityCount := m.world.GetEntityCount()
		questCount := m.world.GetActiveQuestCount()
		tradeVolume := m.world.GetTradeVolume()

		fmt.Fprintf(w, "# HELP venture_entities_total Total number of entities in game world\n")
		fmt.Fprintf(w, "# TYPE venture_entities_total gauge\n")
		fmt.Fprintf(w, "venture_entities_total %d\n", entityCount)

		fmt.Fprintf(w, "# HELP venture_quests_active Number of active quests\n")
		fmt.Fprintf(w, "# TYPE venture_quests_active gauge\n")
		fmt.Fprintf(w, "venture_quests_active %d\n", questCount)

		fmt.Fprintf(w, "# HELP venture_trade_volume_total Total trade volume\n")
		fmt.Fprintf(w, "# TYPE venture_trade_volume_total counter\n")
		fmt.Fprintf(w, "venture_trade_volume_total %d\n", tradeVolume)
	}

	// Runtime metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Fprintf(w, "# HELP venture_go_goroutines Number of active goroutines\n")
	fmt.Fprintf(w, "# TYPE venture_go_goroutines gauge\n")
	fmt.Fprintf(w, "venture_go_goroutines %d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP venture_go_heap_alloc_bytes Bytes allocated in heap\n")
	fmt.Fprintf(w, "# TYPE venture_go_heap_alloc_bytes gauge\n")
	fmt.Fprintf(w, "venture_go_heap_alloc_bytes %d\n", memStats.HeapAlloc)

	fmt.Fprintf(w, "# HELP venture_go_gc_runs_total Total number of GC runs\n")
	fmt.Fprintf(w, "# TYPE venture_go_gc_runs_total counter\n")
	fmt.Fprintf(w, "venture_go_gc_runs_total %d\n", memStats.NumGC)
}

// handleHealth serves a simple health check endpoint.
func (m *MetricsExporter) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}

// handleReady serves a readiness check endpoint.
// Returns 200 OK if all registered readiness checks pass, 503 Service Unavailable otherwise.
func (m *MetricsExporter) handleReady(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	checkers := m.readinessCheckers
	m.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	// Execute all readiness checks
	var failedChecks []string
	for _, checker := range checkers {
		componentName, err := checker.Check()
		if err != nil {
			failedChecks = append(failedChecks, fmt.Sprintf("%s: %v", componentName, err))
			m.logger.WithError(err).WithField("component", componentName).Warn("Readiness check failed")
		}
	}

	// Return status based on checks
	if len(failedChecks) > 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","failed_checks":[`)
		for i, check := range failedChecks {
			if i > 0 {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, `"%s"`, check)
		}
		fmt.Fprintf(w, `]}`)
		fmt.Fprintf(w, "\n")
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready"}`)
		fmt.Fprintf(w, "\n")
	}
}

// handleStatus serves a detailed status endpoint with operational metrics.
// Returns comprehensive server status including uptime, player count, and resource usage.
func (m *MetricsExporter) handleStatus(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Calculate uptime
	uptime := time.Since(m.startTime)

	// Gather runtime metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Fprintf(w, `{"status":"ok",`)
	fmt.Fprintf(w, `"uptime_seconds":%.2f,`, uptime.Seconds())
	fmt.Fprintf(w, `"started_at":"%s",`, m.startTime.Format(time.RFC3339))

	// Performance metrics
	if m.perfMonitor != nil {
		fmt.Fprintf(w, `"performance":{`)
		fmt.Fprintf(w, `"fps":%.2f,`, m.perfMonitor.GetFPS())
		fmt.Fprintf(w, `"frame_time_ms":%.2f,`, m.perfMonitor.GetFrameTime())
		fmt.Fprintf(w, `"memory_mb":%d`, m.perfMonitor.GetMemoryUsageMB())
		fmt.Fprintf(w, `},`)
	}

	// Network metrics
	if m.networkServer != nil {
		fmt.Fprintf(w, `"network":{`)
		fmt.Fprintf(w, `"connected_players":%d,`, m.networkServer.GetConnectedClients())
		fmt.Fprintf(w, `"bytes_sent":%d,`, m.networkServer.GetTotalBytesSent())
		fmt.Fprintf(w, `"bytes_received":%d,`, m.networkServer.GetTotalBytesReceived())
		fmt.Fprintf(w, `"packets_sent":%d,`, m.networkServer.GetPacketsSent())
		fmt.Fprintf(w, `"packets_received":%d`, m.networkServer.GetPacketsReceived())
		fmt.Fprintf(w, `},`)
	}

	// Game state metrics
	if m.world != nil {
		fmt.Fprintf(w, `"game_state":{`)
		fmt.Fprintf(w, `"entity_count":%d,`, m.world.GetEntityCount())
		fmt.Fprintf(w, `"active_quests":%d,`, m.world.GetActiveQuestCount())
		fmt.Fprintf(w, `"trade_volume":%d`, m.world.GetTradeVolume())
		fmt.Fprintf(w, `},`)
	}

	// Runtime metrics
	fmt.Fprintf(w, `"runtime":{`)
	fmt.Fprintf(w, `"goroutines":%d,`, runtime.NumGoroutine())
	fmt.Fprintf(w, `"heap_alloc_bytes":%d,`, memStats.HeapAlloc)
	fmt.Fprintf(w, `"gc_runs":%d`, memStats.NumGC)
	fmt.Fprintf(w, `}`)

	fmt.Fprintf(w, `}`)
	fmt.Fprintf(w, "\n")
}

