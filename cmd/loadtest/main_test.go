package main

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestMockTestClientConnect tests the mock client connection behavior.
func TestMockTestClientConnect(t *testing.T) {
	tests := []struct {
		name          string
		targetLatency time.Duration
		wantErr       bool
	}{
		{
			name:          "low latency connection",
			targetLatency: 50 * time.Millisecond,
			wantErr:       false,
		},
		{
			name:          "medium latency connection",
			targetLatency: 500 * time.Millisecond,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockTestClient{
				id:            1,
				targetLatency: tt.targetLatency,
			}

			err := client.Connect()
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !client.connected {
				t.Error("Connect() succeeded but client not marked as connected")
			}
		})
	}
}

// TestMockTestClientSendUpdate tests the send update behavior.
func TestMockTestClientSendUpdate(t *testing.T) {
	client := &mockTestClient{
		id:            1,
		targetLatency: 100 * time.Millisecond,
	}

	// Should fail when not connected
	err := client.SendUpdate()
	if err == nil {
		t.Error("SendUpdate() should fail when not connected")
	}

	// Connect and try again
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	// Should succeed when connected (most of the time, given low failure rate)
	successCount := 0
	for i := 0; i < 100; i++ {
		if err := client.SendUpdate(); err == nil {
			successCount++
		}
	}

	// With low latency, should have high success rate (>95%)
	if successCount < 95 {
		t.Errorf("SendUpdate() success rate too low: %d/100", successCount)
	}
}

// TestMockTestClientDisconnect tests disconnection behavior.
func TestMockTestClientDisconnect(t *testing.T) {
	client := &mockTestClient{
		id:            1,
		targetLatency: 100 * time.Millisecond,
	}

	// Connect first
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	if !client.connected {
		t.Fatal("Client should be connected")
	}

	// Disconnect
	client.Disconnect()

	if client.connected {
		t.Error("Client should be disconnected")
	}

	// Should not be able to send after disconnect
	err := client.SendUpdate()
	if err == nil {
		t.Error("SendUpdate() should fail after disconnect")
	}
}

// TestTestClientRecordMessageSent tests message sent recording.
func TestTestClientRecordMessageSent(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tc := createTestClient(1, "localhost:8080", 100*time.Millisecond, logger)

	// Record some messages
	for i := 0; i < 10; i++ {
		tc.recordMessageSent()
	}

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.messagesSent != 10 {
		t.Errorf("Expected 10 messages sent, got %d", tc.messagesSent)
	}
}

// TestTestClientRecordMessageReceived tests message received recording.
func TestTestClientRecordMessageReceived(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tc := createTestClient(1, "localhost:8080", 100*time.Millisecond, logger)

	// Record some messages
	for i := 0; i < 15; i++ {
		tc.recordMessageReceived()
	}

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.messagesRecv != 15 {
		t.Errorf("Expected 15 messages received, got %d", tc.messagesRecv)
	}
}

// TestTestClientRecordError tests error recording.
func TestTestClientRecordError(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tc := createTestClient(1, "localhost:8080", 100*time.Millisecond, logger)

	// Record some errors
	tc.recordError(context.Canceled)
	tc.recordError(context.DeadlineExceeded)

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if len(tc.errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(tc.errors))
	}
}

// TestAggregateResults tests the results aggregation logic.
func TestAggregateResults(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	startTime := time.Now()

	// Create test clients with different scenarios
	clients := []*TestClient{
		// Client 0: Successful (connected full duration)
		{
			id:             0,
			connectTime:    startTime,
			disconnectTime: startTime.Add(10 * time.Second),
			connected:      false,
			messagesSent:   100,
			messagesRecv:   100,
			reconnectCount: 0,
			errors:         []error{},
			logger:         logger,
		},
		// Client 1: Successful with reconnects
		{
			id:             1,
			connectTime:    startTime,
			disconnectTime: startTime.Add(10 * time.Second),
			connected:      false,
			messagesSent:   95,
			messagesRecv:   95,
			reconnectCount: 2,
			errors:         []error{context.Canceled},
			logger:         logger,
		},
		// Client 2: Failed (disconnected early)
		{
			id:             2,
			connectTime:    startTime,
			disconnectTime: startTime.Add(5 * time.Second),
			connected:      false,
			messagesSent:   50,
			messagesRecv:   50,
			reconnectCount: 0,
			errors:         []error{context.DeadlineExceeded},
			logger:         logger,
		},
		// Client 3: Successful (still connected at end)
		{
			id:             3,
			connectTime:    startTime,
			disconnectTime: time.Time{}, // Not disconnected
			connected:      true,
			messagesSent:   105,
			messagesRecv:   105,
			reconnectCount: 1,
			errors:         []error{},
			logger:         logger,
		},
	}

	endTime := startTime.Add(10 * time.Second)
	results := aggregateResults(clients, startTime, endTime)

	// Verify basic counts
	if results.TotalClients != 4 {
		t.Errorf("Expected 4 total clients, got %d", results.TotalClients)
	}

	// Client 0, 1, and 3 should be successful (90%+ uptime)
	if results.SuccessfulClients != 3 {
		t.Errorf("Expected 3 successful clients, got %d", results.SuccessfulClients)
	}

	// Client 2 should be counted as premature disconnect
	if results.TotalDisconnects != 1 {
		t.Errorf("Expected 1 premature disconnect, got %d", results.TotalDisconnects)
	}

	// Check total reconnects
	expectedReconnects := 0 + 2 + 0 + 1
	if results.TotalReconnects != expectedReconnects {
		t.Errorf("Expected %d reconnects, got %d", expectedReconnects, results.TotalReconnects)
	}

	// Check total messages
	expectedSent := uint64(100 + 95 + 50 + 105)
	if results.MessagesSent != expectedSent {
		t.Errorf("Expected %d messages sent, got %d", expectedSent, results.MessagesSent)
	}

	expectedRecv := uint64(100 + 95 + 50 + 105)
	if results.MessagesReceived != expectedRecv {
		t.Errorf("Expected %d messages received, got %d", expectedRecv, results.MessagesReceived)
	}

	// Check errors
	if results.Errors != 2 {
		t.Errorf("Expected 2 errors, got %d", results.Errors)
	}
}

// TestCreateTestClient tests test client creation.
func TestCreateTestClient(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name          string
		id            int
		serverAddr    string
		targetLatency time.Duration
	}{
		{
			name:          "low latency client",
			id:            1,
			serverAddr:    "localhost:8080",
			targetLatency: 100 * time.Millisecond,
		},
		{
			name:          "high latency client",
			id:            2,
			serverAddr:    "remote.server:9000",
			targetLatency: 2000 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient(tt.id, tt.serverAddr, tt.targetLatency, logger)

			if client.id != tt.id {
				t.Errorf("Expected id %d, got %d", tt.id, client.id)
			}

			if client.targetLatency != tt.targetLatency {
				t.Errorf("Expected latency %v, got %v", tt.targetLatency, client.targetLatency)
			}

			if client.logger == nil {
				t.Error("Logger should not be nil")
			}

			if client.errors == nil {
				t.Error("Errors slice should be initialized")
			}
		})
	}
}

// TestRunTestClientShortDuration tests running a client for a short duration.
func TestRunTestClientShortDuration(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel) // Reduce noise

	tc := createTestClient(1, "localhost:8080", 50*time.Millisecond, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run client
	runTestClient(ctx, tc)

	// Verify client ran and recorded activity
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.messagesSent == 0 {
		t.Error("Expected some messages to be sent")
	}

	if tc.messagesRecv == 0 {
		t.Error("Expected some messages to be received")
	}

	if tc.connectTime.IsZero() {
		t.Error("Connect time should be set")
	}

	if tc.disconnectTime.IsZero() {
		t.Error("Disconnect time should be set")
	}
}

// BenchmarkMockTestClientSendUpdate benchmarks the send update operation.
func BenchmarkMockTestClientSendUpdate(b *testing.B) {
	client := &mockTestClient{
		id:            1,
		targetLatency: 10 * time.Millisecond, // Low latency for benchmarking
	}

	if err := client.Connect(); err != nil {
		b.Fatalf("Connect() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SendUpdate()
	}
}

// BenchmarkTestClientRecordMessageSent benchmarks message recording.
func BenchmarkTestClientRecordMessageSent(b *testing.B) {
	logger := logrus.NewEntry(logrus.New())
	tc := createTestClient(1, "localhost:8080", 100*time.Millisecond, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.recordMessageSent()
	}
}

// TestPrintProgress tests the progress printing function.
func TestPrintProgress(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	startTime := time.Now()

	clients := []*TestClient{
		{
			id:             0,
			connected:      true,
			reconnectCount: 0,
			errors:         []error{},
			logger:         logger,
		},
		{
			id:             1,
			connected:      false,
			reconnectCount: 2,
			errors:         []error{context.Canceled},
			logger:         logger,
		},
	}

	// Should not panic
	printProgress(clients, startTime)
}

// TestPrintResults tests the results printing function.
func TestPrintResults(t *testing.T) {
	results := LoadTestResults{
		TotalClients:      4,
		Duration:          10 * time.Second,
		SuccessfulClients: 3,
		TotalReconnects:   2,
		TotalDisconnects:  1,
		MessagesSent:      1000,
		MessagesReceived:  1000,
		Errors:            1,
		StartTime:         time.Now().Add(-10 * time.Second),
		EndTime:           time.Now(),
	}

	// Should not panic
	printResults(results)
}

// TestMockTestClientReconnect tests the reconnection behavior.
func TestMockTestClientReconnect(t *testing.T) {
	client := &mockTestClient{
		id:            1,
		targetLatency: 100 * time.Millisecond,
	}

	// Connect first
	if err := client.Connect(); err != nil {
		t.Fatalf("Initial connect failed: %v", err)
	}

	// Simulate disconnection
	client.connected = false

	// Reconnect
	err := client.Reconnect()
	if err != nil {
		// Reconnect can fail due to random chance, that's ok
		return
	}

	if !client.connected {
		t.Error("Client should be connected after successful reconnect")
	}
}
