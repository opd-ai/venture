package webrtc

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewRelayNode(t *testing.T) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 100)

	if node == nil {
		t.Fatal("NewRelayNode returned nil")
	}
	if node.ID != "relay1" {
		t.Errorf("ID = %s, want relay1", node.ID)
	}
	if node.URL != "turn:relay.example.com:3478" {
		t.Errorf("URL = %s, want turn:relay.example.com:3478", node.URL)
	}
	if node.MaxConnections != 100 {
		t.Errorf("MaxConnections = %d, want 100", node.MaxConnections)
	}
	if !node.healthy {
		t.Error("new relay should be healthy")
	}
}

func TestRelayNodeAcquireRelease(t *testing.T) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 2)

	// Acquire first connection
	if err := node.Acquire(); err != nil {
		t.Fatalf("Acquire() failed: %v", err)
	}

	stats := node.GetStats()
	if stats.ActiveConnections != 1 {
		t.Errorf("ActiveConnections = %d, want 1", stats.ActiveConnections)
	}

	// Acquire second connection
	if err := node.Acquire(); err != nil {
		t.Fatalf("Acquire() second failed: %v", err)
	}

	stats = node.GetStats()
	if stats.ActiveConnections != 2 {
		t.Errorf("ActiveConnections = %d, want 2", stats.ActiveConnections)
	}

	// Third acquire should fail (at capacity)
	if err := node.Acquire(); err != ErrRelayFull {
		t.Errorf("Acquire() at capacity = %v, want ErrRelayFull", err)
	}

	// Release one connection
	node.Release()
	stats = node.GetStats()
	if stats.ActiveConnections != 1 {
		t.Errorf("ActiveConnections after release = %d, want 1", stats.ActiveConnections)
	}

	// Should be able to acquire again
	if err := node.Acquire(); err != nil {
		t.Fatalf("Acquire() after release failed: %v", err)
	}
}

func TestRelayNodeUnhealthy(t *testing.T) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 10)

	node.SetHealthy(false)

	if node.IsAvailable() {
		t.Error("unhealthy relay should not be available")
	}

	if err := node.Acquire(); err == nil {
		t.Error("Acquire() on unhealthy relay should fail")
	}
}

func TestRelayNodeStats(t *testing.T) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 10)

	node.UpdateLatency(50 * time.Millisecond)
	node.UpdateBandwidth(1024 * 1024) // 1 MB/s
	node.AddBytesRelayed(1024)

	stats := node.GetStats()
	if stats.Latency != 50*time.Millisecond {
		t.Errorf("Latency = %v, want 50ms", stats.Latency)
	}
	if stats.Bandwidth != 1024*1024 {
		t.Errorf("Bandwidth = %d, want 1048576", stats.Bandwidth)
	}
	if stats.BytesRelayed != 1024 {
		t.Errorf("BytesRelayed = %d, want 1024", stats.BytesRelayed)
	}
	if !stats.Healthy {
		t.Error("relay should be healthy")
	}
}

func TestRelayNodeUtilization(t *testing.T) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 10)

	// 0% utilization
	stats := node.GetStats()
	if stats.Utilization != 0.0 {
		t.Errorf("Utilization = %f, want 0.0", stats.Utilization)
	}

	// 50% utilization
	for i := 0; i < 5; i++ {
		node.Acquire()
	}
	stats = node.GetStats()
	if stats.Utilization != 0.5 {
		t.Errorf("Utilization = %f, want 0.5", stats.Utilization)
	}

	// 100% utilization
	for i := 0; i < 5; i++ {
		node.Acquire()
	}
	stats = node.GetStats()
	if stats.Utilization != 1.0 {
		t.Errorf("Utilization = %f, want 1.0", stats.Utilization)
	}
}

func TestNewRelayManager(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	if rm == nil {
		t.Fatal("NewRelayManager returned nil")
	}
	if rm.strategy != StrategyLowestLatency {
		t.Errorf("strategy = %v, want StrategyLowestLatency", rm.strategy)
	}
}

func TestRelayManagerAddRemove(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)

	// Add relays
	if err := rm.AddRelay(node1); err != nil {
		t.Fatalf("AddRelay(node1) failed: %v", err)
	}
	if err := rm.AddRelay(node2); err != nil {
		t.Fatalf("AddRelay(node2) failed: %v", err)
	}

	// Duplicate add should fail
	if err := rm.AddRelay(node1); err == nil {
		t.Error("AddRelay(duplicate) should fail")
	}

	// Remove relay
	if err := rm.RemoveRelay("relay1"); err != nil {
		t.Fatalf("RemoveRelay(relay1) failed: %v", err)
	}

	// Remove non-existent should fail
	if err := rm.RemoveRelay("relay1"); err == nil {
		t.Error("RemoveRelay(non-existent) should fail")
	}
}

func TestRelayManagerSelectLowestLatency(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	node1.UpdateLatency(100 * time.Millisecond)

	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)
	node2.UpdateLatency(50 * time.Millisecond)

	rm.AddRelay(node1)
	rm.AddRelay(node2)

	selected, err := rm.SelectRelay()
	if err != nil {
		t.Fatalf("SelectRelay() failed: %v", err)
	}

	if selected.ID != "relay2" {
		t.Errorf("selected relay = %s, want relay2 (lowest latency)", selected.ID)
	}
}

func TestRelayManagerSelectHighestBandwidth(t *testing.T) {
	rm := NewRelayManager(StrategyHighestBandwidth)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	node1.UpdateBandwidth(1024 * 1024) // 1 MB/s

	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)
	node2.UpdateBandwidth(5 * 1024 * 1024) // 5 MB/s

	rm.AddRelay(node1)
	rm.AddRelay(node2)

	selected, err := rm.SelectRelay()
	if err != nil {
		t.Fatalf("SelectRelay() failed: %v", err)
	}

	if selected.ID != "relay2" {
		t.Errorf("selected relay = %s, want relay2 (highest bandwidth)", selected.ID)
	}
}

func TestRelayManagerSelectLowestUtilization(t *testing.T) {
	rm := NewRelayManager(StrategyLowestUtilization)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	for i := 0; i < 8; i++ {
		node1.Acquire()
	}

	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)
	for i := 0; i < 2; i++ {
		node2.Acquire()
	}

	rm.AddRelay(node1)
	rm.AddRelay(node2)

	selected, err := rm.SelectRelay()
	if err != nil {
		t.Fatalf("SelectRelay() failed: %v", err)
	}

	if selected.ID != "relay2" {
		t.Errorf("selected relay = %s, want relay2 (lowest utilization)", selected.ID)
	}
}

func TestRelayManagerSelectRoundRobin(t *testing.T) {
	rm := NewRelayManager(StrategyRoundRobin)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)
	node3 := NewRelayNode("relay3", "turn:relay3.example.com:3478", "user", "pass", "eu-central", 10)

	rm.AddRelay(node1)
	rm.AddRelay(node2)
	rm.AddRelay(node3)

	// Track selections
	selections := make(map[string]int)
	for i := 0; i < 9; i++ {
		selected, err := rm.SelectRelay()
		if err != nil {
			t.Fatalf("SelectRelay() failed: %v", err)
		}
		selections[selected.ID]++
	}

	// Each relay should be selected at least once (round-robin distributes evenly)
	for id, count := range selections {
		if count == 0 {
			t.Errorf("relay %s never selected", id)
		}
	}

	// Total selections should equal attempts
	totalSelections := 0
	for _, count := range selections {
		totalSelections += count
	}
	if totalSelections != 9 {
		t.Errorf("total selections = %d, want 9", totalSelections)
	}
}

func TestRelayManagerNoAvailableRelay(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	// No relays added
	_, err := rm.SelectRelay()
	if err != ErrNoRelayAvailable {
		t.Errorf("SelectRelay() with no relays = %v, want ErrNoRelayAvailable", err)
	}

	// Add relay but make it unhealthy
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 10)
	node.SetHealthy(false)
	rm.AddRelay(node)

	_, err = rm.SelectRelay()
	if err != ErrRelayFull {
		t.Errorf("SelectRelay() with unhealthy relay = %v, want ErrRelayFull", err)
	}
}

func TestRelayManagerGetStats(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)

	rm.AddRelay(node1)
	rm.AddRelay(node2)

	stats := rm.GetRelayStats()
	if len(stats) != 2 {
		t.Errorf("GetRelayStats() returned %d relays, want 2", len(stats))
	}
}

func TestSelectionStrategyString(t *testing.T) {
	tests := []struct {
		strategy SelectionStrategy
		want     string
	}{
		{StrategyLowestLatency, "LowestLatency"},
		{StrategyHighestBandwidth, "HighestBandwidth"},
		{StrategyLowestUtilization, "LowestUtilization"},
		{StrategyRoundRobin, "RoundRobin"},
		{SelectionStrategy(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.strategy.String(); got != tt.want {
			t.Errorf("SelectionStrategy(%d).String() = %s, want %s", tt.strategy, got, tt.want)
		}
	}
}

func BenchmarkRelayNodeAcquire(b *testing.B) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 1000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Acquire()
	}
}

func BenchmarkRelayNodeRelease(b *testing.B) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 1000000)
	for i := 0; i < 100000; i++ {
		node.Acquire()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Release()
	}
}

func BenchmarkRelayNodeGetStats(b *testing.B) {
	node := NewRelayNode("relay1", "turn:relay.example.com:3478", "user", "pass", "us-east", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.GetStats()
	}
}

func BenchmarkRelayManagerSelectRelay(b *testing.B) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	for i := 0; i < 10; i++ {
		node := NewRelayNode(
			fmt.Sprintf("relay%d", i),
			fmt.Sprintf("turn:relay%d.example.com:3478", i),
			"user", "pass", "us-east", 100,
		)
		node.UpdateLatency(time.Duration(i*10) * time.Millisecond)
		rm.AddRelay(node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.SelectRelay()
	}
}

func TestRelayManagerRoundRobinOverflowProtection(t *testing.T) {
	rm := NewRelayManager(StrategyRoundRobin)
	defer rm.Close()

	node1 := NewRelayNode("relay1", "turn:relay1.example.com:3478", "user", "pass", "us-east", 10)
	node2 := NewRelayNode("relay2", "turn:relay2.example.com:3478", "user", "pass", "us-west", 10)
	rm.AddRelay(node1)
	rm.AddRelay(node2)

	// Simulate counter near max int to verify overflow reset
	rm.mu.Lock()
	rm.rrCounter = int(^uint(0) >> 1) // max int
	rm.mu.Unlock()

	// Should not panic or produce negative index
	for i := 0; i < 5; i++ {
		selected, err := rm.SelectRelay()
		if err != nil {
			t.Fatalf("SelectRelay() after overflow failed: %v", err)
		}
		if selected == nil {
			t.Fatal("SelectRelay() returned nil")
		}
	}

	// Verify counter was reset after overflow
	rm.mu.RLock()
	if rm.rrCounter < 0 {
		t.Errorf("rrCounter = %d, should be non-negative after overflow protection", rm.rrCounter)
	}
	rm.mu.RUnlock()
}

func TestRelayManagerSelectionSnapshotLocking(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	defer rm.Close()

	// Create nodes with different latencies
	for i := 0; i < 5; i++ {
		node := NewRelayNode(
			fmt.Sprintf("relay%d", i),
			fmt.Sprintf("turn:relay%d.example.com:3478", i),
			"user", "pass", "us-east", 100,
		)
		node.UpdateLatency(time.Duration((i+1)*10) * time.Millisecond)
		rm.AddRelay(node)
	}

	// Concurrent selection and updates should not deadlock
	done := make(chan bool, 10)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				rm.SelectRelay()
			}
			done <- true
		}()
		go func(idx int) {
			for j := 0; j < 100; j++ {
				rm.mu.RLock()
				for _, node := range rm.nodes {
					node.UpdateLatency(time.Duration(idx*10+j) * time.Millisecond)
				}
				rm.mu.RUnlock()
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("deadlock detected: concurrent selection and updates timed out")
		}
	}
}

// TestPingRelayInvalidURLs tests pingRelay with invalid URL formats to ensure no panic.
func TestPingRelayInvalidURLs(t *testing.T) {
	rm := NewRelayManager(StrategyLowestLatency)
	ctx := context.Background()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "empty URL",
			url:  "",
			want: false,
		},
		{
			name: "too short URL",
			url:  "turn",
			want: false,
		},
		{
			name: "missing prefix",
			url:  "relay.example.com:3478",
			want: false,
		},
		{
			name: "invalid prefix",
			url:  "http:relay.example.com:3478",
			want: false,
		},
		{
			name: "valid turn prefix but missing port",
			url:  "turn:relay.example.com",
			want: false,
		},
		{
			name: "valid turns prefix but missing port",
			url:  "turns:relay.example.com",
			want: false,
		},
		{
			name: "turn with malformed host",
			url:  "turn:::3478",
			want: false,
		},
		{
			name: "turns with malformed host",
			url:  "turns:::3478",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			got := rm.pingRelay(ctx, tt.url)
			if got != tt.want {
				t.Errorf("pingRelay() = %v, want %v", got, tt.want)
			}
		})
	}
}
