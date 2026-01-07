package observability

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// Mock implementations for testing

type mockPerformanceMonitor struct {
	fps       float64
	frameTime float64
	memoryMB  uint64
}

func (m *mockPerformanceMonitor) GetFPS() float64            { return m.fps }
func (m *mockPerformanceMonitor) GetFrameTime() float64     { return m.frameTime }
func (m *mockPerformanceMonitor) GetMemoryUsageMB() uint64  { return m.memoryMB }

type mockNetworkServer struct {
	clients       int
	bytesSent     uint64
	bytesReceived uint64
	packetsSent   uint64
	packetsRecv   uint64
}

func (m *mockNetworkServer) GetConnectedClients() int      { return m.clients }
func (m *mockNetworkServer) GetTotalBytesSent() uint64     { return m.bytesSent }
func (m *mockNetworkServer) GetTotalBytesReceived() uint64 { return m.bytesReceived }
func (m *mockNetworkServer) GetPacketsSent() uint64        { return m.packetsSent }
func (m *mockNetworkServer) GetPacketsReceived() uint64    { return m.packetsRecv }

type mockWorld struct {
	entityCount int
	questCount  int
	tradeVolume uint64
}

func (m *mockWorld) GetEntityCount() int       { return m.entityCount }
func (m *mockWorld) GetActiveQuestCount() int  { return m.questCount }
func (m *mockWorld) GetTradeVolume() uint64    { return m.tradeVolume }

// Tests

func TestNewMetricsExporter(t *testing.T) {
	exporter := NewMetricsExporter(":9090")
	
	if exporter == nil {
		t.Fatal("Expected non-nil exporter")
	}
	
	if exporter.addr != ":9090" {
		t.Errorf("Expected address :9090, got %s", exporter.addr)
	}
	
	if exporter.logger == nil {
		t.Error("Expected non-nil logger")
	}
}

func TestNewMetricsExporterWithLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	
	exporter := NewMetricsExporterWithLogger(":9091", logger)
	
	if exporter == nil {
		t.Fatal("Expected non-nil exporter")
	}
	
	if exporter.logger != logger {
		t.Error("Expected custom logger to be set")
	}
}

func TestRegisterPerformanceMonitor(t *testing.T) {
	exporter := NewMetricsExporter(":9092")
	mock := &mockPerformanceMonitor{fps: 60.0, frameTime: 16.67, memoryMB: 100}
	
	exporter.RegisterPerformanceMonitor(mock)
	
	if exporter.perfMonitor != mock {
		t.Error("Performance monitor not registered")
	}
}

func TestRegisterNetworkServer(t *testing.T) {
	exporter := NewMetricsExporter(":9093")
	mock := &mockNetworkServer{clients: 4, bytesSent: 1024, bytesReceived: 2048}
	
	exporter.RegisterNetworkServer(mock)
	
	if exporter.networkServer != mock {
		t.Error("Network server not registered")
	}
}

func TestRegisterWorld(t *testing.T) {
	exporter := NewMetricsExporter(":9094")
	mock := &mockWorld{entityCount: 1000, questCount: 10, tradeVolume: 5000}
	
	exporter.RegisterWorld(mock)
	
	if exporter.world != mock {
		t.Error("World not registered")
	}
}

func TestStartStop(t *testing.T) {
	exporter := NewMetricsExporter(":19090")
	
	// Start server
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop server
	if err := exporter.Stop(); err != nil {
		t.Fatalf("Failed to stop exporter: %v", err)
	}
}

func TestStartAlreadyStarted(t *testing.T) {
	exporter := NewMetricsExporter(":19091")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	// Try to start again
	err := exporter.Start()
	if err == nil {
		t.Error("Expected error when starting already started exporter")
	}
	if !strings.Contains(err.Error(), "already started") {
		t.Errorf("Expected 'already started' error, got: %v", err)
	}
}

func TestStopNotStarted(t *testing.T) {
	exporter := NewMetricsExporter(":19092")
	
	// Stop without starting should not error
	if err := exporter.Stop(); err != nil {
		t.Errorf("Unexpected error when stopping not-started exporter: %v", err)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	exporter := NewMetricsExporter(":19093")
	
	// Register mock sources
	exporter.RegisterPerformanceMonitor(&mockPerformanceMonitor{
		fps:       60.0,
		frameTime: 16.67,
		memoryMB:  120,
	})
	exporter.RegisterNetworkServer(&mockNetworkServer{
		clients:       4,
		bytesSent:     102400,
		bytesReceived: 204800,
		packetsSent:   1000,
		packetsRecv:   2000,
	})
	exporter.RegisterWorld(&mockWorld{
		entityCount: 2000,
		questCount:  15,
		tradeVolume: 10000,
	})
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	// Make request to metrics endpoint
	resp, err := http.Get("http://localhost:19093/metrics")
	if err != nil {
		t.Fatalf("Failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	
	// Verify metrics are present
	tests := []struct {
		name   string
		metric string
	}{
		{"uptime", "venture_uptime_seconds"},
		{"fps", "venture_fps 60.00"},
		{"frame_time", "venture_frame_time_ms 16.67"},
		{"memory", "venture_memory_usage_mb 120"},
		{"players", "venture_players_connected 4"},
		{"bytes_sent", "venture_network_bytes_sent_total 102400"},
		{"bytes_recv", "venture_network_bytes_received_total 204800"},
		{"packets_sent", "venture_network_packets_sent_total 1000"},
		{"packets_recv", "venture_network_packets_received_total 2000"},
		{"entities", "venture_entities_total 2000"},
		{"quests", "venture_quests_active 15"},
		{"trade", "venture_trade_volume_total 10000"},
		{"goroutines", "venture_go_goroutines"},
		{"heap", "venture_go_heap_alloc_bytes"},
		{"gc", "venture_go_gc_runs_total"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.metric) {
				t.Errorf("Expected metric %q in output, not found", tt.metric)
			}
		})
	}
	
	// Verify Prometheus format
	if !strings.Contains(output, "# HELP") {
		t.Error("Expected Prometheus HELP comments")
	}
	if !strings.Contains(output, "# TYPE") {
		t.Error("Expected Prometheus TYPE comments")
	}
}

func TestMetricsEndpointNoSources(t *testing.T) {
	exporter := NewMetricsExporter(":19094")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	// Make request without registered sources
	resp, err := http.Get("http://localhost:19094/metrics")
	if err != nil {
		t.Fatalf("Failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	
	// Should still have uptime and runtime metrics
	if !strings.Contains(output, "venture_uptime_seconds") {
		t.Error("Expected uptime metric even without sources")
	}
	if !strings.Contains(output, "venture_go_goroutines") {
		t.Error("Expected runtime metrics even without sources")
	}
}

func TestHealthEndpoint(t *testing.T) {
	exporter := NewMetricsExporter(":19095")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:19095/health")
	if err != nil {
		t.Fatalf("Failed to GET /health: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	if string(body) != "OK\n" {
		t.Errorf("Expected 'OK', got %q", string(body))
	}
}

func TestStopWithTimeout(t *testing.T) {
	exporter := NewMetricsExporter(":19096")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Stop with custom timeout
	if err := exporter.StopWithTimeout(5 * time.Second); err != nil {
		t.Fatalf("Failed to stop exporter: %v", err)
	}
	
	// Verify server is stopped
	_, err := http.Get("http://localhost:19096/metrics")
	if err == nil {
		t.Error("Expected error when connecting to stopped server")
	}
}

func TestContentType(t *testing.T) {
	exporter := NewMetricsExporter(":19097")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:19097/metrics")
	if err != nil {
		t.Fatalf("Failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	
	contentType := resp.Header.Get("Content-Type")
	expectedType := "text/plain; version=0.0.4; charset=utf-8"
	
	if contentType != expectedType {
		t.Errorf("Expected Content-Type %q, got %q", expectedType, contentType)
	}
}

func TestConcurrentAccess(t *testing.T) {
	exporter := NewMetricsExporter(":19098")
	
	mock := &mockPerformanceMonitor{fps: 60.0, frameTime: 16.67, memoryMB: 100}
	exporter.RegisterPerformanceMonitor(mock)
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	// Make concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := http.Get("http://localhost:19098/metrics")
			if err != nil {
				t.Errorf("Failed to GET /metrics: %v", err)
				done <- false
				return
			}
			resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
				done <- false
				return
			}
			done <- true
		}()
	}
	
	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Mock readiness checker for testing

type mockReadinessChecker struct {
	name string
	err  error
}

func (m *mockReadinessChecker) Check() (string, error) {
	return m.name, m.err
}

func TestRegisterReadinessChecker(t *testing.T) {
	exporter := NewMetricsExporter(":29090")
	checker := &mockReadinessChecker{name: "test", err: nil}
	
	exporter.RegisterReadinessChecker(checker)
	
	if len(exporter.readinessCheckers) != 1 {
		t.Errorf("Expected 1 readiness checker, got %d", len(exporter.readinessCheckers))
	}
	if exporter.readinessCheckers[0] != checker {
		t.Error("Readiness checker not registered correctly")
	}
}

func TestReadyEndpointAllChecksPass(t *testing.T) {
	exporter := NewMetricsExporter(":29091")
	
	// Register checkers that pass
	exporter.RegisterReadinessChecker(&mockReadinessChecker{name: "database", err: nil})
	exporter.RegisterReadinessChecker(&mockReadinessChecker{name: "federation", err: nil})
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:29091/ready")
	if err != nil {
		t.Fatalf("Failed to GET /ready: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	if !strings.Contains(output, `"status":"ready"`) {
		t.Errorf("Expected ready status in output, got: %s", output)
	}
}

func TestReadyEndpointCheckFails(t *testing.T) {
	exporter := NewMetricsExporter(":29092")
	
	// Register checkers with one that fails
	exporter.RegisterReadinessChecker(&mockReadinessChecker{name: "database", err: nil})
	exporter.RegisterReadinessChecker(&mockReadinessChecker{name: "federation", err: fmt.Errorf("connection timeout")})
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:29092/ready")
	if err != nil {
		t.Fatalf("Failed to GET /ready: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	if !strings.Contains(output, `"status":"not_ready"`) {
		t.Errorf("Expected not_ready status in output, got: %s", output)
	}
	if !strings.Contains(output, "federation") {
		t.Errorf("Expected federation in failed checks, got: %s", output)
	}
	if !strings.Contains(output, "connection timeout") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestReadyEndpointNoCheckers(t *testing.T) {
	exporter := NewMetricsExporter(":29093")
	
	// Don't register any checkers
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:29093/ready")
	if err != nil {
		t.Fatalf("Failed to GET /ready: %v", err)
	}
	defer resp.Body.Close()
	
	// Should be ready if no checks are registered
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	if !strings.Contains(output, `"status":"ready"`) {
		t.Errorf("Expected ready status in output, got: %s", output)
	}
}

func TestStatusEndpoint(t *testing.T) {
	exporter := NewMetricsExporter(":29094")
	
	// Register mock sources
	exporter.RegisterPerformanceMonitor(&mockPerformanceMonitor{
		fps:       60.0,
		frameTime: 16.67,
		memoryMB:  120,
	})
	exporter.RegisterNetworkServer(&mockNetworkServer{
		clients:       4,
		bytesSent:     102400,
		bytesReceived: 204800,
		packetsSent:   1000,
		packetsRecv:   2000,
	})
	exporter.RegisterWorld(&mockWorld{
		entityCount: 2000,
		questCount:  15,
		tradeVolume: 10000,
	})
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:29094/status")
	if err != nil {
		t.Fatalf("Failed to GET /status: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	
	// Verify JSON structure and key fields
	tests := []struct {
		name  string
		field string
	}{
		{"status", `"status":"ok"`},
		{"uptime", `"uptime_seconds"`},
		{"started_at", `"started_at"`},
		{"fps", `"fps":60.00`},
		{"frame_time", `"frame_time_ms":16.67`},
		{"memory", `"memory_mb":120`},
		{"connected_players", `"connected_players":4`},
		{"bytes_sent", `"bytes_sent":102400`},
		{"bytes_received", `"bytes_received":204800`},
		{"entity_count", `"entity_count":2000`},
		{"active_quests", `"active_quests":15`},
		{"trade_volume", `"trade_volume":10000`},
		{"goroutines", `"goroutines"`},
		{"heap_alloc", `"heap_alloc_bytes"`},
		{"gc_runs", `"gc_runs"`},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.field) {
				t.Errorf("Expected field %q in status output, not found. Output: %s", tt.field, output)
			}
		})
	}
}

func TestStatusEndpointNoSources(t *testing.T) {
	exporter := NewMetricsExporter(":29095")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:29095/status")
	if err != nil {
		t.Fatalf("Failed to GET /status: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	output := string(body)
	
	// Should still have basic fields even without sources
	if !strings.Contains(output, `"status":"ok"`) {
		t.Error("Expected status ok even without sources")
	}
	if !strings.Contains(output, `"uptime_seconds"`) {
		t.Error("Expected uptime even without sources")
	}
	if !strings.Contains(output, `"runtime"`) {
		t.Error("Expected runtime metrics even without sources")
	}
}

func TestReadyEndpointContentType(t *testing.T) {
	exporter := NewMetricsExporter(":29096")
	
	if err := exporter.Start(); err != nil {
		t.Fatalf("Failed to start exporter: %v", err)
	}
	defer exporter.Stop()
	
	time.Sleep(100 * time.Millisecond)
	
	resp, err := http.Get("http://localhost:29096/ready")
	if err != nil {
		t.Fatalf("Failed to GET /ready: %v", err)
	}
	defer resp.Body.Close()
	
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

