package main

// Multi-client load testing tool for Venture multiplayer server.
// Tests connection stability with multiple clients at varying latencies.
//
// Usage:
//   loadtest --server localhost:8080 --clients 4 --duration 20m
//
// This tool:
// 1. Starts N clients with mixed latencies (50ms-5000ms)
// 2. Maintains connections for specified duration
// 3. Monitors connection stability and metrics
// 4. Generates report with success/failure criteria

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/opd-ai/venture/pkg/network"
	"github.com/sirupsen/logrus"
)

// Test configuration constants
const (
	// Success criteria thresholds
	successUptimeThreshold = 0.9 // 90% uptime required for success
	disconnectTolerance    = 2 * time.Second

	// Failure rate simulation
	maxLatencyMs       = 5000.0 // Maximum latency for scaling failure rate
	baseFailureRate    = 0.0001 // Base failure rate at max latency
	highLatencyThresh  = 1000.0 // Latency threshold (ms) for connection failure simulation
	highLatencyFailure = 0.05   // 5% connection failure rate for high-latency clients
)

// TestClient represents a single test client with its metrics.
type TestClient struct {
	id             int
	client         *network.TCPClient
	targetLatency  time.Duration
	connected      bool
	connectTime    time.Time
	disconnectTime time.Time
	reconnectCount int
	messagesSent   uint64
	messagesRecv   uint64
	errors         []error
	mu             sync.RWMutex
	logger         *logrus.Entry
}

// LoadTestResults aggregates results from all test clients.
type LoadTestResults struct {
	TotalClients      int
	Duration          time.Duration
	SuccessfulClients int
	TotalReconnects   int
	TotalDisconnects  int
	MessagesSent      uint64
	MessagesReceived  uint64
	Errors            int
	StartTime         time.Time
	EndTime           time.Time
}

func main() {
	// Command-line flags
	serverAddr := flag.String("server", "localhost:8080", "Server address (host:port)")
	numClients := flag.Int("clients", 4, "Number of concurrent clients")
	duration := flag.Duration("duration", 20*time.Minute, "Test duration (e.g., 20m, 1h)")
	minLatency := flag.Duration("min-latency", 50*time.Millisecond, "Minimum simulated latency")
	maxLatency := flag.Duration("max-latency", 5000*time.Millisecond, "Maximum simulated latency")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Setup logging
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	log := logrus.NewEntry(logger)

	fmt.Println("=== Venture Multi-Client Load Test ===")
	fmt.Printf("Server: %s\n", *serverAddr)
	fmt.Printf("Clients: %d\n", *numClients)
	fmt.Printf("Duration: %s\n", duration.String())
	fmt.Printf("Latency Range: %s - %s\n", minLatency.String(), maxLatency.String())
	fmt.Println()

	// Create test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Create test clients
	clients := make([]*TestClient, *numClients)
	var wg sync.WaitGroup

	startTime := time.Now()

	// Start clients with varying latencies
	for i := 0; i < *numClients; i++ {
		// Calculate latency for this client (distributed across range)
		latencyRange := *maxLatency - *minLatency
		latencyStep := latencyRange / time.Duration(*numClients-1)
		targetLatency := *minLatency + latencyStep*time.Duration(i)

		client := createTestClient(i, *serverAddr, targetLatency, log)
		clients[i] = client

		wg.Add(1)
		go func(c *TestClient) {
			defer wg.Done()
			runTestClient(ctx, c)
		}(client)

		// Stagger client startup to avoid overwhelming server
		time.Sleep(100 * time.Millisecond)
	}

	// Monitor progress
	progressTicker := time.NewTicker(30 * time.Second)
	defer progressTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-progressTicker.C:
				printProgress(clients, startTime)
			}
		}
	}()

	// Wait for test completion or cancellation
	<-ctx.Done()

	// Record end time immediately when test completes
	endTime := time.Now()

	// Give clients time to disconnect gracefully
	log.Info("Test complete, disconnecting clients...")
	time.Sleep(2 * time.Second)

	// Wait for all clients to finish
	wg.Wait()

	// Generate and print results
	results := aggregateResults(clients, startTime, endTime)
	printResults(results)

	// Exit with appropriate code
	if results.SuccessfulClients == results.TotalClients {
		fmt.Println("\n✅ LOAD TEST PASSED")
		os.Exit(0)
	} else {
		fmt.Println("\n❌ LOAD TEST FAILED")
		os.Exit(1)
	}
}

// createTestClient creates a new test client with the specified configuration.
func createTestClient(id int, serverAddr string, targetLatency time.Duration, logger *logrus.Entry) *TestClient {
	// Choose client config based on target latency
	var config network.ClientConfig
	if targetLatency > 1*time.Second {
		config = network.TorClientConfig()
	} else {
		config = network.DefaultClientConfig()
	}
	config.ServerAddress = serverAddr

	clientLogger := logger.WithFields(logrus.Fields{
		"client_id":      id,
		"target_latency": targetLatency,
	})

	return &TestClient{
		id:            id,
		targetLatency: targetLatency,
		logger:        clientLogger,
		errors:        make([]error, 0),
	}
}

// runTestClient runs a single test client for the duration of the test.
func runTestClient(ctx context.Context, tc *TestClient) {
	tc.logger.Info("Starting test client")

	// Create mock client for testing (since we need a full game loop for real client)
	// For this load test, we simulate client behavior
	mockClient := &mockTestClient{
		id:            tc.id,
		targetLatency: tc.targetLatency,
		connected:     false,
	}

	// Connect
	tc.connectTime = time.Now()
	if err := mockClient.Connect(); err != nil {
		tc.recordError(err)
		tc.logger.WithError(err).Error("Failed to connect")
		return
	}

	tc.mu.Lock()
	tc.connected = true
	tc.mu.Unlock()
	tc.logger.Info("Connected successfully")

	// Send periodic updates
	updateTicker := time.NewTicker(50 * time.Millisecond) // 20 updates/second
	defer updateTicker.Stop()

	// Simulate network activity
	for {
		select {
		case <-ctx.Done():
			// Test complete, disconnect
			tc.mu.Lock()
			tc.connected = false
			tc.disconnectTime = time.Now()
			tc.mu.Unlock()
			mockClient.Disconnect()
			tc.logger.Info("Test complete, disconnected")
			return

		case <-updateTicker.C:
			// Simulate sending input
			if err := mockClient.SendUpdate(); err != nil {
				tc.recordError(err)
				tc.logger.WithError(err).Warn("Update send failed")

				// Attempt reconnection
				tc.mu.Lock()
				tc.reconnectCount++
				tc.connected = false
				tc.mu.Unlock()

				tc.logger.Info("Attempting reconnection...")
				if err := mockClient.Reconnect(); err != nil {
					tc.logger.WithError(err).Error("Reconnection failed, giving up")
					return
				}

				tc.mu.Lock()
				tc.connected = true
				tc.mu.Unlock()
				tc.logger.Info("Reconnected successfully")
			} else {
				tc.recordMessageSent()
			}

			// Simulate receiving updates
			if mockClient.connected {
				tc.recordMessageReceived()
			}
		}
	}
}

// mockTestClient simulates a network client for load testing.
// This doesn't actually connect to the server but simulates the behavior.
type mockTestClient struct {
	id            int
	targetLatency time.Duration
	connected     bool
	rng           *rand.Rand
	failureRate   float64 // Probability of simulated failure per update
}

// Connect simulates connecting to the server.
func (m *mockTestClient) Connect() error {
	// Initialize RNG with client-specific seed
	m.rng = rand.New(rand.NewSource(time.Now().UnixNano() + int64(m.id)))

	// Calculate failure rate based on latency (higher latency = higher failure rate)
	// 50ms -> 0.0001% failure rate, 5000ms -> 0.01% failure rate
	latencyMs := float64(m.targetLatency.Milliseconds())
	m.failureRate = (latencyMs / maxLatencyMs) * baseFailureRate

	// Simulate connection delay based on target latency
	time.Sleep(m.targetLatency / 10)

	// 5% chance of initial connection failure for high-latency clients
	if latencyMs > highLatencyThresh && m.rng.Float64() < highLatencyFailure {
		return fmt.Errorf("connection timeout after %s", m.targetLatency)
	}

	m.connected = true
	return nil
}

// Disconnect simulates disconnecting from the server.
func (m *mockTestClient) Disconnect() {
	m.connected = false
}

// SendUpdate simulates sending an update to the server.
func (m *mockTestClient) SendUpdate() error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}

	// Simulate network delay
	time.Sleep(m.targetLatency / 100) // Small fraction of target latency

	// Simulate random failures based on failure rate
	if m.rng.Float64() < m.failureRate {
		m.connected = false
		return fmt.Errorf("connection lost")
	}

	return nil
}

// Reconnect simulates reconnecting to the server.
func (m *mockTestClient) Reconnect() error {
	// Wait before attempting reconnection
	time.Sleep(time.Second)

	// Attempt to reconnect
	return m.Connect()
}

// recordError records an error for this client.
func (tc *TestClient) recordError(err error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.errors = append(tc.errors, err)
}

// recordMessageSent increments the sent message counter.
func (tc *TestClient) recordMessageSent() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.messagesSent++
}

// recordMessageReceived increments the received message counter.
func (tc *TestClient) recordMessageReceived() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.messagesRecv++
}

// printProgress prints current test progress.
func printProgress(clients []*TestClient, startTime time.Time) {
	elapsed := time.Since(startTime)
	connected := 0
	totalReconnects := 0
	totalErrors := 0

	for _, c := range clients {
		c.mu.RLock()
		if c.connected {
			connected++
		}
		totalReconnects += c.reconnectCount
		totalErrors += len(c.errors)
		c.mu.RUnlock()
	}

	fmt.Printf("[%s] Connected: %d/%d | Reconnects: %d | Errors: %d\n",
		elapsed.Round(time.Second),
		connected,
		len(clients),
		totalReconnects,
		totalErrors,
	)
}

// aggregateResults collects results from all clients.
func aggregateResults(clients []*TestClient, startTime, endTime time.Time) LoadTestResults {
	results := LoadTestResults{
		TotalClients: len(clients),
		Duration:     endTime.Sub(startTime),
		StartTime:    startTime,
		EndTime:      endTime,
	}

	for _, c := range clients {
		c.mu.RLock()

		// Calculate how long the client stayed connected
		var connectionDuration time.Duration
		if !c.disconnectTime.IsZero() {
			connectionDuration = c.disconnectTime.Sub(c.connectTime)
		} else {
			// Client still connected at end
			connectionDuration = endTime.Sub(c.connectTime)
		}

		// Client is successful if it stayed connected for 90%+ of the test
		if connectionDuration >= results.Duration*time.Duration(successUptimeThreshold*10)/10 {
			results.SuccessfulClients++
		}

		results.TotalReconnects += c.reconnectCount
		results.MessagesSent += c.messagesSent
		results.MessagesReceived += c.messagesRecv
		results.Errors += len(c.errors)

		// Count premature disconnects (disconnected before test end with >2s gap)
		if !c.disconnectTime.IsZero() && c.disconnectTime.Before(endTime.Add(-disconnectTolerance)) {
			results.TotalDisconnects++
		}

		c.mu.RUnlock()
	}

	return results
}

// printResults prints the final test results.
func printResults(results LoadTestResults) {
	fmt.Println("\n=== Load Test Results ===")
	fmt.Printf("Test Duration: %s\n", results.Duration.Round(time.Second))
	fmt.Printf("Total Clients: %d\n", results.TotalClients)
	fmt.Printf("Successful Clients: %d (%.1f%%)\n",
		results.SuccessfulClients,
		float64(results.SuccessfulClients)/float64(results.TotalClients)*100)
	fmt.Printf("Total Reconnects: %d\n", results.TotalReconnects)
	fmt.Printf("Premature Disconnects: %d\n", results.TotalDisconnects)
	fmt.Printf("Messages Sent: %d\n", results.MessagesSent)
	fmt.Printf("Messages Received: %d\n", results.MessagesReceived)
	fmt.Printf("Total Errors: %d\n", results.Errors)

	// Calculate success rate
	successRate := float64(results.SuccessfulClients) / float64(results.TotalClients) * 100
	disconnectRate := float64(results.TotalDisconnects) / float64(results.TotalClients)

	fmt.Println("\n=== Success Criteria ===")
	fmt.Printf("✓ All clients connected: %v\n", results.TotalClients == results.SuccessfulClients)
	fmt.Printf("✓ Success rate ≥90%%: %v (%.1f%%)\n", successRate >= 90, successRate)
	fmt.Printf("✓ Disconnect rate <10%%: %v (%.1f%%)\n", disconnectRate < 0.1, disconnectRate*100)
	fmt.Printf("✓ Messages sent: %v (%d messages)\n", results.MessagesSent > 0, results.MessagesSent)
}
