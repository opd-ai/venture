// Package main demonstrates the health, readiness, and status endpoints.
// This example shows how to set up the observability metrics exporter
// with custom readiness checks for production monitoring.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/venture/pkg/observability"
	"github.com/sirupsen/logrus"
)

// MockReadinessChecker implements a simple readiness check.
type MockReadinessChecker struct {
	componentName string
	isReady       bool
}

func (m *MockReadinessChecker) Check() (string, error) {
	if !m.isReady {
		return m.componentName, fmt.Errorf("component not ready")
	}
	return m.componentName, nil
}

// MockPerformanceMonitor provides fake performance metrics.
type MockPerformanceMonitor struct{}

func (m *MockPerformanceMonitor) GetFPS() float64            { return 60.0 }
func (m *MockPerformanceMonitor) GetFrameTime() float64     { return 16.67 }
func (m *MockPerformanceMonitor) GetMemoryUsageMB() uint64  { return 120 }

// MockNetworkServer provides fake network metrics.
type MockNetworkServer struct {
	clients int
}

func (m *MockNetworkServer) GetConnectedClients() int      { return m.clients }
func (m *MockNetworkServer) GetTotalBytesSent() uint64     { return 1024000 }
func (m *MockNetworkServer) GetTotalBytesReceived() uint64 { return 2048000 }
func (m *MockNetworkServer) GetPacketsSent() uint64        { return 10000 }
func (m *MockNetworkServer) GetPacketsReceived() uint64    { return 20000 }

// MockWorld provides fake game state metrics.
type MockWorld struct{}

func (m *MockWorld) GetEntityCount() int      { return 1500 }
func (m *MockWorld) GetActiveQuestCount() int { return 25 }
func (m *MockWorld) GetTradeVolume() uint64   { return 50000 }

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create metrics exporter
	exporter := observability.NewMetricsExporterWithLogger(":9090", logger)

	// Register metrics sources
	exporter.RegisterPerformanceMonitor(&MockPerformanceMonitor{})
	exporter.RegisterNetworkServer(&MockNetworkServer{clients: 4})
	exporter.RegisterWorld(&MockWorld{})

	// Register readiness checkers
	databaseChecker := &MockReadinessChecker{
		componentName: "database",
		isReady:       true, // Simulating ready database
	}
	federationChecker := &MockReadinessChecker{
		componentName: "federation",
		isReady:       true, // Simulating ready federation
	}

	exporter.RegisterReadinessChecker(databaseChecker)
	exporter.RegisterReadinessChecker(federationChecker)

	// Start metrics server
	if err := exporter.Start(); err != nil {
		log.Fatalf("Failed to start metrics exporter: %v", err)
	}

	logger.Info("Health endpoints demo started")
	logger.Info("Endpoints available:")
	logger.Info("  GET http://localhost:9090/health  - Basic liveness check")
	logger.Info("  GET http://localhost:9090/ready   - Readiness check with component validation")
	logger.Info("  GET http://localhost:9090/status  - Detailed status with metrics")
	logger.Info("  GET http://localhost:9090/metrics - Prometheus metrics")
	logger.Info("")
	logger.Info("Demonstrating endpoints in 2 seconds...")

	// Wait a moment for server to start
	time.Sleep(2 * time.Second)

	// Demonstrate each endpoint
	demonstrateEndpoints()

	// Simulate federation becoming unavailable after 5 seconds
	go func() {
		time.Sleep(5 * time.Second)
		logger.Warn("Simulating federation failure...")
		federationChecker.isReady = false
		logger.Info("Try GET http://localhost:9090/ready now to see the failure")
	}()

	logger.Info("")
	logger.Info("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
	if err := exporter.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping exporter")
	}
	logger.Info("Shutdown complete")
}

func demonstrateEndpoints() {
	endpoints := []string{
		"http://localhost:9090/health",
		"http://localhost:9090/ready",
		"http://localhost:9090/status",
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\n=== %s ===\n", endpoint)
		resp, err := http.Get(endpoint)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
			continue
		}

		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
		fmt.Printf("Body:\n%s\n", string(body))
	}
}
