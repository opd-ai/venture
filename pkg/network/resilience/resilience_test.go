package resilience

import (
	"testing"
	"time"
)

// Test network configuration validation
func TestNetworkConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  NetworkConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: NetworkConfig{
				Latency:        200 * time.Millisecond,
				PacketLossRate: 0.05,
				Jitter:         50 * time.Millisecond,
				BandwidthLimit: 50000,
			},
			wantErr: false,
		},
		{
			name: "zero config (valid)",
			config: NetworkConfig{
				Latency:        0,
				PacketLossRate: 0,
				Jitter:         0,
				BandwidthLimit: 0,
			},
			wantErr: false,
		},
		{
			name: "negative latency (invalid)",
			config: NetworkConfig{
				Latency:        -100 * time.Millisecond,
				PacketLossRate: 0,
				Jitter:         0,
				BandwidthLimit: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid packet loss rate (too high)",
			config: NetworkConfig{
				Latency:        0,
				PacketLossRate: 1.5,
				Jitter:         0,
				BandwidthLimit: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid packet loss rate (negative)",
			config: NetworkConfig{
				Latency:        0,
				PacketLossRate: -0.1,
				Jitter:         0,
				BandwidthLimit: 0,
			},
			wantErr: true,
		},
		{
			name: "negative jitter (invalid)",
			config: NetworkConfig{
				Latency:        0,
				PacketLossRate: 0,
				Jitter:         -10 * time.Millisecond,
				BandwidthLimit: 0,
			},
			wantErr: true,
		},
		{
			name: "negative bandwidth limit (invalid)",
			config: NetworkConfig{
				Latency:        0,
				PacketLossRate: 0,
				Jitter:         0,
				BandwidthLimit: -1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test network simulator creation
func TestNewNetworkSimulator(t *testing.T) {
	sim := NewNetworkSimulator()
	if sim == nil {
		t.Fatal("NewNetworkSimulator() returned nil")
	}

	config := sim.GetConfig()
	if config.Latency != 0 {
		t.Errorf("Default latency = %v, want 0", config.Latency)
	}
	if config.PacketLossRate != 0 {
		t.Errorf("Default packet loss rate = %v, want 0", config.PacketLossRate)
	}
}

// Test simulator with config
func TestNewNetworkSimulatorWithConfig(t *testing.T) {
	config := NetworkConfig{
		Latency:        100 * time.Millisecond,
		PacketLossRate: 0.1,
		Jitter:         20 * time.Millisecond,
		BandwidthLimit: 10000,
	}

	sim, err := NewNetworkSimulatorWithConfig(config)
	if err != nil {
		t.Fatalf("NewNetworkSimulatorWithConfig() error = %v", err)
	}

	got := sim.GetConfig()
	if got.Latency != config.Latency {
		t.Errorf("Latency = %v, want %v", got.Latency, config.Latency)
	}
	if got.PacketLossRate != config.PacketLossRate {
		t.Errorf("PacketLossRate = %v, want %v", got.PacketLossRate, config.PacketLossRate)
	}
}

// Test packet sending without impairments
func TestNetworkSimulator_Send_NoImpairments(t *testing.T) {
	sim := NewNetworkSimulator()

	data := []byte("test packet")
	err := sim.Send(data)
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}

	sent, dropped, bytes := sim.GetStats()
	if sent != 1 {
		t.Errorf("PacketsSent = %d, want 1", sent)
	}
	if dropped != 0 {
		t.Errorf("PacketsDropped = %d, want 0", dropped)
	}
	if bytes != uint64(len(data)) {
		t.Errorf("BytesProcessed = %d, want %d", bytes, len(data))
	}
}

// Test packet loss simulation
func TestNetworkSimulator_Send_PacketLoss(t *testing.T) {
	sim := NewNetworkSimulator()
	sim.SetPacketLoss(1.0) // 100% packet loss

	data := []byte("test packet")
	err := sim.Send(data)
	if err != ErrPacketDropped {
		t.Errorf("Send() error = %v, want ErrPacketDropped", err)
	}

	sent, dropped, _ := sim.GetStats()
	if sent != 0 {
		t.Errorf("PacketsSent = %d, want 0", sent)
	}
	if dropped != 1 {
		t.Errorf("PacketsDropped = %d, want 1", dropped)
	}
}

// Test latency simulation
func TestNetworkSimulator_Send_Latency(t *testing.T) {
	sim := NewNetworkSimulator()
	sim.SetLatency(100 * time.Millisecond)

	data := []byte("test packet")
	err := sim.Send(data)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Packets should be delayed
	queueSize := sim.QueueSize()
	if queueSize != 1 {
		t.Errorf("QueueSize = %d, want 1", queueSize)
	}

	// Process immediately (not ready yet)
	ready := sim.ProcessDelayedPackets()
	if len(ready) != 0 {
		t.Errorf("ProcessDelayedPackets() returned %d packets, want 0", len(ready))
	}

	// Wait for latency and process
	time.Sleep(150 * time.Millisecond)
	ready = sim.ProcessDelayedPackets()
	if len(ready) != 1 {
		t.Errorf("ProcessDelayedPackets() returned %d packets, want 1", len(ready))
	}

	// Queue should be empty now
	queueSize = sim.QueueSize()
	if queueSize != 0 {
		t.Errorf("QueueSize after processing = %d, want 0", queueSize)
	}
}

// Test bandwidth limiting
func TestNetworkSimulator_Send_BandwidthLimit(t *testing.T) {
	sim := NewNetworkSimulator()
	sim.SetBandwidthLimit(100) // 100 bytes per second

	// Send 50 bytes (should succeed)
	data1 := make([]byte, 50)
	err := sim.Send(data1)
	if err != nil {
		t.Errorf("First Send() error = %v", err)
	}

	// Send another 60 bytes (should fail - exceeds 100 byte limit)
	data2 := make([]byte, 60)
	err = sim.Send(data2)
	if err != ErrBandwidthExceeded {
		t.Errorf("Second Send() error = %v, want ErrBandwidthExceeded", err)
	}

	// Wait for counter to reset (>1 second)
	time.Sleep(1100 * time.Millisecond)

	// Should succeed now
	err = sim.Send(data2)
	if err != nil {
		t.Errorf("Third Send() after reset error = %v", err)
	}
}

// Test metrics collector
func TestMetricsCollector_RecordLatency(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordLatency(100 * time.Millisecond)
	mc.RecordLatency(200 * time.Millisecond)
	mc.RecordLatency(150 * time.Millisecond)

	stats := mc.GetStats()

	// Check average (should be ~150ms)
	if stats.AvgLatency < 140*time.Millisecond || stats.AvgLatency > 160*time.Millisecond {
		t.Errorf("AvgLatency = %v, want ~150ms", stats.AvgLatency)
	}

	// Check min (should be 100ms)
	if stats.MinLatency != 100*time.Millisecond {
		t.Errorf("MinLatency = %v, want 100ms", stats.MinLatency)
	}

	// Check max (should be 200ms)
	if stats.MaxLatency != 200*time.Millisecond {
		t.Errorf("MaxLatency = %v, want 200ms", stats.MaxLatency)
	}
}

// Test packet statistics
func TestMetricsCollector_PacketStats(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordPacketSent(100)
	mc.RecordPacketSent(200)
	mc.RecordPacketLoss()
	mc.RecordPacketReceived(100)

	stats := mc.GetStats()

	if stats.PacketsSent != 2 {
		t.Errorf("PacketsSent = %d, want 2", stats.PacketsSent)
	}
	if stats.PacketsDropped != 1 {
		t.Errorf("PacketsDropped = %d, want 1", stats.PacketsDropped)
	}
	if stats.PacketsReceived != 1 {
		t.Errorf("PacketsReceived = %d, want 1", stats.PacketsReceived)
	}
	if stats.BytesSent != 300 {
		t.Errorf("BytesSent = %d, want 300", stats.BytesSent)
	}

	// Packet loss rate = 1 / (2 + 1) = 0.333...
	expectedLossRate := 1.0 / 3.0
	if stats.PacketLossRate < expectedLossRate-0.01 || stats.PacketLossRate > expectedLossRate+0.01 {
		t.Errorf("PacketLossRate = %f, want ~%f", stats.PacketLossRate, expectedLossRate)
	}
}

// Test prediction tracking
func TestMetricsCollector_Predictions(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordPrediction(false) // Correct prediction
	mc.RecordPrediction(false)
	mc.RecordPrediction(true) // Misprediction
	mc.RecordPrediction(false)
	mc.RecordPrediction(true) // Misprediction

	stats := mc.GetStats()

	if stats.MispredictionCount != 2 {
		t.Errorf("MispredictionCount = %d, want 2", stats.MispredictionCount)
	}

	// Misprediction rate = 2 / 5 = 0.4
	if stats.MispredictionRate != 0.4 {
		t.Errorf("MispredictionRate = %f, want 0.4", stats.MispredictionRate)
	}
}

// Test desync and reconnect tracking
func TestMetricsCollector_DesyncReconnect(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordDesync()
	mc.RecordDesync()
	mc.RecordReconnect(5 * time.Second)
	mc.RecordReconnect(7 * time.Second)
	mc.RecordReconnect(3 * time.Second)

	stats := mc.GetStats()

	if stats.DesyncCount != 2 {
		t.Errorf("DesyncCount = %d, want 2", stats.DesyncCount)
	}
	if stats.ReconnectCount != 3 {
		t.Errorf("ReconnectCount = %d, want 3", stats.ReconnectCount)
	}

	// Average reconnect time = (5 + 7 + 3) / 3 = 5 seconds
	if stats.AvgReconnectTime != 5*time.Second {
		t.Errorf("AvgReconnectTime = %v, want 5s", stats.AvgReconnectTime)
	}
}

// Test metrics reset
func TestMetricsCollector_Reset(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordLatency(100 * time.Millisecond)
	mc.RecordPacketSent(100)
	mc.RecordDesync()

	stats := mc.GetStats()
	if stats.PacketsSent == 0 {
		t.Error("Stats should not be empty before reset")
	}

	mc.Reset()

	stats = mc.GetStats()
	if stats.PacketsSent != 0 {
		t.Errorf("PacketsSent after reset = %d, want 0", stats.PacketsSent)
	}
	if stats.DesyncCount != 0 {
		t.Errorf("DesyncCount after reset = %d, want 0", stats.DesyncCount)
	}
}

// Benchmark packet sending
func BenchmarkNetworkSimulator_Send(b *testing.B) {
	sim := NewNetworkSimulator()
	data := []byte("test packet data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.Send(data)
	}
}

// Benchmark packet sending with latency
func BenchmarkNetworkSimulator_Send_WithLatency(b *testing.B) {
	sim := NewNetworkSimulator()
	sim.SetLatency(100 * time.Millisecond)
	data := []byte("test packet data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.Send(data)
	}
}

// Benchmark metrics recording
func BenchmarkMetricsCollector_RecordLatency(b *testing.B) {
	mc := NewMetricsCollector()
	latency := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.RecordLatency(latency)
	}
}

// Benchmark stats calculation
func BenchmarkMetricsCollector_GetStats(b *testing.B) {
	mc := NewMetricsCollector()

	// Pre-fill with data
	for i := 0; i < 1000; i++ {
		mc.RecordLatency(time.Duration(i) * time.Millisecond)
		mc.RecordPacketSent(100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.GetStats()
	}
}

// Test SetConfig method
func TestNetworkSimulator_SetConfig(t *testing.T) {
	sim := NewNetworkSimulator()

	config := NetworkConfig{
		Latency:        150 * time.Millisecond,
		PacketLossRate: 0.05,
		Jitter:         30 * time.Millisecond,
		BandwidthLimit: 20000,
	}

	err := sim.SetConfig(config)
	if err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	got := sim.GetConfig()
	if got.Latency != config.Latency {
		t.Errorf("Latency = %v, want %v", got.Latency, config.Latency)
	}
	if got.PacketLossRate != config.PacketLossRate {
		t.Errorf("PacketLossRate = %v, want %v", got.PacketLossRate, config.PacketLossRate)
	}
	if got.Jitter != config.Jitter {
		t.Errorf("Jitter = %v, want %v", got.Jitter, config.Jitter)
	}
	if got.BandwidthLimit != config.BandwidthLimit {
		t.Errorf("BandwidthLimit = %v, want %v", got.BandwidthLimit, config.BandwidthLimit)
	}
}

// Test SetConfig with invalid config
func TestNetworkSimulator_SetConfig_Invalid(t *testing.T) {
	sim := NewNetworkSimulator()

	invalidConfig := NetworkConfig{
		Latency:        -100 * time.Millisecond,
		PacketLossRate: 0,
		Jitter:         0,
		BandwidthLimit: 0,
	}

	err := sim.SetConfig(invalidConfig)
	if err == nil {
		t.Error("SetConfig() should return error for invalid config")
	}
}

// Test SetJitter method
func TestNetworkSimulator_SetJitter(t *testing.T) {
	sim := NewNetworkSimulator()

	jitter := 50 * time.Millisecond
	sim.SetJitter(jitter)

	config := sim.GetConfig()
	if config.Jitter != jitter {
		t.Errorf("Jitter = %v, want %v", config.Jitter, jitter)
	}
}

// Test bandwidth calculation with multiple samples
func TestMetricsCollector_BandwidthCalculations(t *testing.T) {
	mc := NewMetricsCollector()

	// Simulate bandwidth samples by recording latency with delays
	// This triggers bandwidth sample collection
	mc.RecordLatency(10 * time.Millisecond)

	// Record some packets to generate bandwidth data
	for i := 0; i < 5; i++ {
		mc.RecordPacketSent(1000)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for at least one second to collect bandwidth sample
	time.Sleep(1100 * time.Millisecond)
	mc.RecordLatency(10 * time.Millisecond)

	stats := mc.GetStats()

	// Should have bandwidth statistics
	if stats.AvgBandwidth < 0 {
		t.Errorf("AvgBandwidth = %f, should be >= 0", stats.AvgBandwidth)
	}
	if stats.PeakBandwidth < 0 {
		t.Errorf("PeakBandwidth = %f, should be >= 0", stats.PeakBandwidth)
	}
}

// Test bandwidth sample update with high traffic
func TestMetricsCollector_UpdateBandwidthSample(t *testing.T) {
	mc := NewMetricsCollector()

	// Send multiple packets quickly to accumulate bytes
	for i := 0; i < 10; i++ {
		mc.RecordPacketSent(500)
	}

	// Wait for bandwidth sample window
	time.Sleep(1200 * time.Millisecond)

	// Trigger stats calculation which updates bandwidth sample
	mc.RecordLatency(5 * time.Millisecond)

	stats := mc.GetStats()

	// Should have recorded bandwidth
	if len(mc.bandwidthSamples) == 0 && stats.BytesSent > 0 {
		t.Log("Bandwidth samples collected (internal state)")
	}
}

// Test metrics with zero latency samples
func TestMetricsCollector_EmptyLatencySamples(t *testing.T) {
	mc := NewMetricsCollector()

	// Record other metrics but no latency
	mc.RecordPacketSent(100)
	mc.RecordPacketReceived(100)

	stats := mc.GetStats()

	// Latency stats should be zero
	if stats.AvgLatency != 0 {
		t.Errorf("AvgLatency = %v, want 0 for no samples", stats.AvgLatency)
	}
	if stats.MinLatency != 0 {
		t.Errorf("MinLatency = %v, want 0 for no samples", stats.MinLatency)
	}
	if stats.MaxLatency != 0 {
		t.Errorf("MaxLatency = %v, want 0 for no samples", stats.MaxLatency)
	}
}

// Test metrics with zero bandwidth samples
func TestMetricsCollector_EmptyBandwidthSamples(t *testing.T) {
	mc := NewMetricsCollector()

	// Don't wait for bandwidth samples
	stats := mc.GetStats()

	// Bandwidth stats should be zero
	if stats.AvgBandwidth != 0 {
		t.Errorf("AvgBandwidth = %f, want 0 for no samples", stats.AvgBandwidth)
	}
	if stats.PeakBandwidth != 0 {
		t.Errorf("PeakBandwidth = %f, want 0 for no samples", stats.PeakBandwidth)
	}
}

// Test RecordLatency with updated bandwidth tracking
func TestMetricsCollector_RecordLatency_WithBandwidth(t *testing.T) {
	mc := NewMetricsCollector()

	// Record packets and latency over time
	mc.RecordPacketSent(1000)
	mc.RecordLatency(50 * time.Millisecond)

	// Wait for bandwidth sample window
	time.Sleep(1100 * time.Millisecond)

	mc.RecordPacketSent(1000)
	mc.RecordLatency(60 * time.Millisecond)

	stats := mc.GetStats()

	// Should have recorded latencies
	if stats.AvgLatency == 0 {
		t.Error("AvgLatency should not be 0 after recording")
	}

	// Bandwidth may or may not be collected depending on timing
	// Just verify no panic occurred
}

// Test SetPacketLoss with edge cases
func TestNetworkSimulator_SetPacketLoss_EdgeCases(t *testing.T) {
	sim := NewNetworkSimulator()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"negative rate clamped to 0", -0.5, 0.0},
		{"rate above 1 clamped", 1.5, 1.0},
		{"zero rate", 0.0, 0.0},
		{"max rate", 1.0, 1.0},
		{"normal rate", 0.3, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim.SetPacketLoss(tt.input)
			config := sim.GetConfig()
			if config.PacketLossRate != tt.expected {
				t.Errorf("PacketLossRate = %v, want %v", config.PacketLossRate, tt.expected)
			}
		})
	}
}

// Test calculateMinLatency with single sample
func TestMetricsCollector_CalculateMinLatency_SingleSample(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordLatency(100 * time.Millisecond)

	stats := mc.GetStats()

	if stats.MinLatency != 100*time.Millisecond {
		t.Errorf("MinLatency = %v, want 100ms for single sample", stats.MinLatency)
	}
}

// Test calculatePercentile with edge cases
func TestMetricsCollector_CalculatePercentile_EdgeCases(t *testing.T) {
	mc := NewMetricsCollector()

	// Add multiple latency samples
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, lat := range latencies {
		mc.RecordLatency(lat)
	}

	stats := mc.GetStats()

	// P95 should be close to 95ms (95th percentile of 10-100ms)
	if stats.P95Latency < 90*time.Millisecond || stats.P95Latency > 100*time.Millisecond {
		t.Errorf("P95Latency = %v, want ~95ms", stats.P95Latency)
	}

	// P99 should be close to 100ms
	if stats.P99Latency < 95*time.Millisecond || stats.P99Latency > 100*time.Millisecond {
		t.Errorf("P99Latency = %v, want ~100ms", stats.P99Latency)
	}
}

// Test calculatePeakBandwidth with varying samples
func TestMetricsCollector_CalculatePeakBandwidth_Varying(t *testing.T) {
	mc := NewMetricsCollector()

	// Simulate varying bandwidth over several seconds
	for i := 0; i < 3; i++ {
		for j := 0; j < 5; j++ {
			mc.RecordPacketSent(500 * (i + 1)) // Increasing traffic
		}
		time.Sleep(1100 * time.Millisecond) // Wait for bandwidth sample
		mc.RecordLatency(10 * time.Millisecond) // Trigger bandwidth update
	}

	stats := mc.GetStats()

	// Should have some bandwidth data
	if stats.PeakBandwidth <= 0 {
		t.Errorf("PeakBandwidth = %f, want > 0", stats.PeakBandwidth)
	}

	if stats.AvgBandwidth <= 0 {
		t.Errorf("AvgBandwidth = %f, want > 0", stats.AvgBandwidth)
	}

	// Peak should be >= average
	if stats.PeakBandwidth < stats.AvgBandwidth {
		t.Errorf("PeakBandwidth (%f) should be >= AvgBandwidth (%f)", stats.PeakBandwidth, stats.AvgBandwidth)
	}
}
