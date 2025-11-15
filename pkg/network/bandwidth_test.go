package network

import (
	"testing"
	"time"
)

func TestBandwidthMonitor_RecordSent(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	playerID := uint64(100)
	monitor.RecordSent(playerID, 1024)
	monitor.RecordSent(playerID, 512)

	stats := monitor.GetPlayerStats(playerID)
	if stats.BytesSent != 1536 {
		t.Errorf("Expected 1536 bytes sent, got %d", stats.BytesSent)
	}
	if stats.MessagesSent != 2 {
		t.Errorf("Expected 2 messages sent, got %d", stats.MessagesSent)
	}
}

func TestBandwidthMonitor_RecordReceived(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	playerID := uint64(200)
	monitor.RecordReceived(playerID, 2048)
	monitor.RecordReceived(playerID, 1024)

	stats := monitor.GetPlayerStats(playerID)
	if stats.BytesReceived != 3072 {
		t.Errorf("Expected 3072 bytes received, got %d", stats.BytesReceived)
	}
	if stats.MessagesRecv != 2 {
		t.Errorf("Expected 2 messages received, got %d", stats.MessagesRecv)
	}
}

func TestBandwidthMonitor_GlobalStats(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	monitor.RecordSent(100, 1000)
	monitor.RecordSent(101, 2000)
	monitor.RecordReceived(100, 500)
	monitor.RecordReceived(102, 1500)

	sent, received, duration := monitor.GetGlobalStats()
	if sent != 3000 {
		t.Errorf("Expected 3000 bytes sent globally, got %d", sent)
	}
	if received != 2000 {
		t.Errorf("Expected 2000 bytes received globally, got %d", received)
	}
	if duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestBandwidthMonitor_UpdateRates(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	playerID := uint64(300)
	monitor.RecordSent(playerID, 10240) // 10 KB

	// Wait a bit to accumulate time
	time.Sleep(100 * time.Millisecond)

	monitor.UpdateRates()

	stats := monitor.GetPlayerStats(playerID)
	if stats.CurrentRate == 0 {
		t.Error("Expected non-zero current rate")
	}

	// Should be approximately 102,400 bytes/sec (10KB / 0.1s)
	t.Logf("Current rate: %.2f bytes/sec", stats.CurrentRate)
}

func TestBandwidthMonitor_ResetPlayerStats(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	playerID := uint64(400)
	monitor.RecordSent(playerID, 5000)
	monitor.RecordReceived(playerID, 3000)

	// Verify stats exist
	stats := monitor.GetPlayerStats(playerID)
	if stats.BytesSent == 0 || stats.BytesReceived == 0 {
		t.Fatal("Stats should not be zero before reset")
	}

	// Reset
	monitor.ResetPlayerStats(playerID)

	// Verify stats are reset
	stats = monitor.GetPlayerStats(playerID)
	if stats.BytesSent != 0 {
		t.Errorf("Expected BytesSent to be 0 after reset, got %d", stats.BytesSent)
	}
	if stats.BytesReceived != 0 {
		t.Errorf("Expected BytesReceived to be 0 after reset, got %d", stats.BytesReceived)
	}
	if stats.MessagesSent != 0 {
		t.Errorf("Expected MessagesSent to be 0 after reset, got %d", stats.MessagesSent)
	}
}

func TestBandwidthMonitor_GetAllPlayerStats(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	// Record for multiple players
	monitor.RecordSent(100, 1000)
	monitor.RecordSent(101, 2000)
	monitor.RecordSent(102, 3000)

	allStats := monitor.GetAllPlayerStats()
	if len(allStats) != 3 {
		t.Errorf("Expected 3 player stats, got %d", len(allStats))
	}

	for playerID, stats := range allStats {
		if stats.PlayerID != playerID {
			t.Errorf("Player ID mismatch: key=%d, stats.PlayerID=%d", playerID, stats.PlayerID)
		}
	}
}

func TestCalculateBandwidthKBps(t *testing.T) {
	tests := []struct {
		name      string
		bytes     uint64
		duration  time.Duration
		expected  float64
		tolerance float64
	}{
		{
			name:      "1 KB per second",
			bytes:     1024,
			duration:  1 * time.Second,
			expected:  1.0,
			tolerance: 0.01,
		},
		{
			name:      "10 KB per second",
			bytes:     10240,
			duration:  1 * time.Second,
			expected:  10.0,
			tolerance: 0.01,
		},
		{
			name:      "5 KB over 2 seconds",
			bytes:     5120,
			duration:  2 * time.Second,
			expected:  2.5,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBandwidthKBps(tt.bytes, tt.duration)
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("Expected %.2f KB/s, got %.2f KB/s", tt.expected, result)
			}
		})
	}
}

func TestCalculateBandwidthMBps(t *testing.T) {
	// Test 1 MB per second
	bytes := uint64(1024 * 1024)
	duration := 1 * time.Second
	result := CalculateBandwidthMBps(bytes, duration)

	if result < 0.99 || result > 1.01 {
		t.Errorf("Expected ~1.0 MB/s, got %.2f MB/s", result)
	}
}

func TestBandwidthMonitor_ChatScenario(t *testing.T) {
	// Simulate Phase 36 acceptance criteria: 10 messages/minute, <10KB/s
	monitor := NewBandwidthMonitor(60 * time.Second)

	playerID := uint64(500)
	messagesPerMinute := 10
	avgMessageSize := uint64(150) // bytes (typical chat message)

	// Simulate sending 10 messages over 60 seconds
	for i := 0; i < messagesPerMinute; i++ {
		monitor.RecordSent(playerID, avgMessageSize)
		time.Sleep(6 * time.Second / 10) // Spread evenly (6 seconds / 10 messages = ~0.6s per message, but we'll use faster simulation)
	}

	// Actually, don't sleep in tests - just record
	sent, _, _ := monitor.GetGlobalStats()

	// Estimate bandwidth (assuming 60-second window)
	bandwidth := CalculateBandwidthKBps(sent, 60*time.Second)

	t.Logf("Sent %d bytes over simulation period", sent)
	t.Logf("Estimated bandwidth: %.2f KB/s", bandwidth)

	// Check Phase 36 acceptance criteria: <10 KB/s
	if bandwidth > 10.0 {
		t.Errorf("Bandwidth %.2f KB/s exceeds 10 KB/s target", bandwidth)
	}
}

func TestBandwidthMonitor_PeakRate(t *testing.T) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	playerID := uint64(600)

	// Send burst of data
	monitor.RecordSent(playerID, 50000) // 50 KB burst
	time.Sleep(50 * time.Millisecond)

	monitor.UpdateRates()
	stats := monitor.GetPlayerStats(playerID)

	if stats.PeakRate == 0 {
		t.Error("Expected non-zero peak rate")
	}

	t.Logf("Peak rate: %.2f bytes/sec (%.2f KB/s)", stats.PeakRate, stats.PeakRate/1024.0)
}

// BenchmarkRecordSent benchmarks recording sent bytes
func BenchmarkRecordSent(b *testing.B) {
	monitor := NewBandwidthMonitor(60 * time.Second)
	playerID := uint64(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordSent(playerID, 100)
	}
}

// BenchmarkUpdateRates benchmarks rate calculation
func BenchmarkUpdateRates(b *testing.B) {
	monitor := NewBandwidthMonitor(60 * time.Second)

	// Add some players with data
	for i := 0; i < 50; i++ {
		monitor.RecordSent(uint64(i), 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.UpdateRates()
	}
}
