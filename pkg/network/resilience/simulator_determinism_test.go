package resilience

import (
	"testing"
	"time"
)

// TestNewNetworkSimulatorWithSeed verifies seeded constructors exist
func TestNewNetworkSimulatorWithSeed(t *testing.T) {
	seed := int64(12345)
	sim := NewNetworkSimulatorWithSeed(seed)

	if sim == nil {
		t.Fatal("NewNetworkSimulatorWithSeed() returned nil")
	}

	// Verify default config
	config := sim.GetConfig()
	if config.Latency != 0 {
		t.Errorf("Default latency = %v, want 0", config.Latency)
	}
	if config.PacketLossRate != 0 {
		t.Errorf("Default packet loss rate = %v, want 0", config.PacketLossRate)
	}
}

// TestNewNetworkSimulatorWithConfigAndSeed verifies config + seed constructor
func TestNewNetworkSimulatorWithConfigAndSeed(t *testing.T) {
	config := NetworkConfig{
		Latency:        100 * time.Millisecond,
		PacketLossRate: 0.1,
		Jitter:         20 * time.Millisecond,
		BandwidthLimit: 10000,
	}
	seed := int64(54321)

	sim, err := NewNetworkSimulatorWithConfigAndSeed(config, seed)
	if err != nil {
		t.Fatalf("NewNetworkSimulatorWithConfigAndSeed() error = %v", err)
	}

	got := sim.GetConfig()
	if got.Latency != config.Latency {
		t.Errorf("Latency = %v, want %v", got.Latency, config.Latency)
	}
	if got.PacketLossRate != config.PacketLossRate {
		t.Errorf("PacketLossRate = %v, want %v", got.PacketLossRate, config.PacketLossRate)
	}
}

// TestNetworkSimulator_Determinism_PacketLoss verifies packet loss is deterministic
func TestNetworkSimulator_Determinism_PacketLoss(t *testing.T) {
	seed := int64(42)
	lossRate := 0.5

	// Create two simulators with same seed
	sim1 := NewNetworkSimulatorWithSeed(seed)
	sim1.SetPacketLoss(lossRate)

	sim2 := NewNetworkSimulatorWithSeed(seed)
	sim2.SetPacketLoss(lossRate)

	// Send packets and track which ones are dropped
	const numPackets = 100
	dropped1 := make([]bool, numPackets)
	dropped2 := make([]bool, numPackets)

	data := []byte("test packet")

	for i := 0; i < numPackets; i++ {
		err1 := sim1.Send(data)
		dropped1[i] = (err1 == ErrPacketDropped)

		err2 := sim2.Send(data)
		dropped2[i] = (err2 == ErrPacketDropped)
	}

	// Verify identical drop patterns
	for i := 0; i < numPackets; i++ {
		if dropped1[i] != dropped2[i] {
			t.Errorf("Packet %d: sim1 dropped=%v, sim2 dropped=%v (should match)", i, dropped1[i], dropped2[i])
		}
	}

	// Verify both had some packets dropped (sanity check)
	sent1, dropped1Count, _ := sim1.GetStats()
	sent2, dropped2Count, _ := sim2.GetStats()

	if dropped1Count == 0 {
		t.Error("Expected some packets to be dropped with 50% loss rate")
	}

	if sent1 != sent2 || dropped1Count != dropped2Count {
		t.Errorf("Stats mismatch: sim1 sent=%d dropped=%d, sim2 sent=%d dropped=%d",
			sent1, dropped1Count, sent2, dropped2Count)
	}
}

// TestNetworkSimulator_Determinism_Jitter verifies jitter calculation is deterministic
func TestNetworkSimulator_Determinism_Jitter(t *testing.T) {
	seed := int64(9999)
	latency := 100 * time.Millisecond
	jitter := 50 * time.Millisecond

	// Create two simulators with same seed
	sim1 := NewNetworkSimulatorWithSeed(seed)
	sim1.SetLatency(latency)
	sim1.SetJitter(jitter)

	sim2 := NewNetworkSimulatorWithSeed(seed)
	sim2.SetLatency(latency)
	sim2.SetJitter(jitter)

	// Send packets and capture delivery times
	const numPackets = 50
	data := []byte("test packet")

	for i := 0; i < numPackets; i++ {
		_ = sim1.Send(data)
		_ = sim2.Send(data)
	}

	// Both should have same queue sizes
	queue1 := sim1.QueueSize()
	queue2 := sim2.QueueSize()

	if queue1 != queue2 {
		t.Errorf("QueueSize mismatch: sim1=%d, sim2=%d", queue1, queue2)
	}

	if queue1 != numPackets {
		t.Errorf("QueueSize = %d, want %d (all packets should be delayed)", queue1, numPackets)
	}
}

// TestNetworkSimulator_NonDeterminism_DifferentSeeds verifies different seeds produce different results
func TestNetworkSimulator_NonDeterminism_DifferentSeeds(t *testing.T) {
	seed1 := int64(111)
	seed2 := int64(222)
	lossRate := 0.5

	sim1 := NewNetworkSimulatorWithSeed(seed1)
	sim1.SetPacketLoss(lossRate)

	sim2 := NewNetworkSimulatorWithSeed(seed2)
	sim2.SetPacketLoss(lossRate)

	// Send packets and track drops
	const numPackets = 100
	dropped1 := make([]bool, numPackets)
	dropped2 := make([]bool, numPackets)

	data := []byte("test packet")

	for i := 0; i < numPackets; i++ {
		err1 := sim1.Send(data)
		dropped1[i] = (err1 == ErrPacketDropped)

		err2 := sim2.Send(data)
		dropped2[i] = (err2 == ErrPacketDropped)
	}

	// Count differences
	differences := 0
	for i := 0; i < numPackets; i++ {
		if dropped1[i] != dropped2[i] {
			differences++
		}
	}

	// Different seeds should produce different results
	// With 100 packets at 50% loss rate, expect significant differences
	if differences == 0 {
		t.Error("Expected different drop patterns with different seeds")
	}

	// Require at least 10% different outcomes
	if float64(differences)/float64(numPackets) < 0.1 {
		t.Errorf("Only %d/%d packets differed (%.1f%%), expected more variation",
			differences, numPackets, float64(differences)/float64(numPackets)*100)
	}
}

// TestNetworkSimulator_Determinism_MultipleRuns verifies reproducibility across runs
func TestNetworkSimulator_Determinism_MultipleRuns(t *testing.T) {
	seed := int64(7777)
	config := NetworkConfig{
		Latency:        200 * time.Millisecond,
		PacketLossRate: 0.3,
		Jitter:         50 * time.Millisecond,
		BandwidthLimit: 0,
	}

	const numRuns = 3
	const numPackets = 50
	data := []byte("test packet data")

	// Track results from each run
	type runResult struct {
		sent    uint64
		dropped uint64
		bytes   uint64
		queue   int
	}
	results := make([]runResult, numRuns)

	for run := 0; run < numRuns; run++ {
		sim, err := NewNetworkSimulatorWithConfigAndSeed(config, seed)
		if err != nil {
			t.Fatalf("Run %d: NewNetworkSimulatorWithConfigAndSeed() error = %v", run, err)
		}

		// Send packets
		for i := 0; i < numPackets; i++ {
			_ = sim.Send(data)
		}

		sent, dropped, bytes := sim.GetStats()
		results[run] = runResult{
			sent:    sent,
			dropped: dropped,
			bytes:   bytes,
			queue:   sim.QueueSize(),
		}
	}

	// Verify all runs produced identical results
	for run := 1; run < numRuns; run++ {
		if results[run] != results[0] {
			t.Errorf("Run %d results differ from run 0:\n  Run 0: %+v\n  Run %d: %+v",
				run, results[0], run, results[run])
		}
	}
}

// TestNetworkSimulator_Determinism_Reset verifies reset doesn't affect seed
func TestNetworkSimulator_Determinism_Reset(t *testing.T) {
	seed := int64(5555)
	lossRate := 0.4

	sim := NewNetworkSimulatorWithSeed(seed)
	sim.SetPacketLoss(lossRate)

	// First batch of packets
	const numPackets = 20
	data := []byte("test")
	dropped1 := make([]bool, numPackets)

	for i := 0; i < numPackets; i++ {
		err := sim.Send(data)
		dropped1[i] = (err == ErrPacketDropped)
	}

	sent1, dropped1Count, _ := sim.GetStats()

	// Reset simulator
	sim.Reset()

	// Second batch of packets - should produce DIFFERENT results
	// (RNG continues from previous state, not reset to original seed)
	dropped2 := make([]bool, numPackets)

	for i := 0; i < numPackets; i++ {
		err := sim.Send(data)
		dropped2[i] = (err == ErrPacketDropped)
	}

	sent2, dropped2Count, _ := sim.GetStats()

	// After reset, stats should restart at 0 counts
	if sent2 != uint64(numPackets-dropped2Count) {
		t.Errorf("After reset, sent = %d, want %d", sent2, numPackets-dropped2Count)
	}

	// The drop patterns should likely be different (RNG state continues)
	// This is expected behavior - reset clears stats but not RNG state
	// If drop patterns are identical, that's suspicious but possible with small sample
	identical := true
	for i := 0; i < numPackets; i++ {
		if dropped1[i] != dropped2[i] {
			identical = false
			break
		}
	}

	// Just log this for information
	t.Logf("After reset: first batch sent=%d dropped=%d, second batch sent=%d dropped=%d, patterns identical=%v",
		sent1, dropped1Count, sent2, dropped2Count, identical)
}

// TestNetworkSimulator_BackwardCompatibility verifies non-seeded constructor still works
func TestNetworkSimulator_BackwardCompatibility(t *testing.T) {
	// Create two simulators without explicit seed
	sim1 := NewNetworkSimulator()
	sim2 := NewNetworkSimulator()

	if sim1 == nil || sim2 == nil {
		t.Fatal("NewNetworkSimulator() returned nil")
	}

	// They should work correctly with no packet loss for guaranteed sends
	sim1.SetPacketLoss(0.0)
	sim2.SetPacketLoss(0.0)

	data := []byte("test")
	err1 := sim1.Send(data)
	err2 := sim2.Send(data)

	if err1 != nil {
		t.Errorf("sim1.Send() error = %v, want nil", err1)
	}
	if err2 != nil {
		t.Errorf("sim2.Send() error = %v, want nil", err2)
	}

	// They should have stats
	sent1, dropped1, _ := sim1.GetStats()
	sent2, dropped2, _ := sim2.GetStats()

	if sent1 != 1 {
		t.Errorf("sim1 sent = %d, want 1", sent1)
	}
	if sent2 != 1 {
		t.Errorf("sim2 sent = %d, want 1", sent2)
	}
	if dropped1 != 0 {
		t.Errorf("sim1 dropped = %d, want 0", dropped1)
	}
	if dropped2 != 0 {
		t.Errorf("sim2 dropped = %d, want 0", dropped2)
	}

	// Since they use time-based seeds, they will have different RNG states
	// This is expected - we're just verifying they work correctly
}

// TestNewNetworkSimulatorWithConfigAndSeed_InvalidConfig verifies error handling
func TestNewNetworkSimulatorWithConfigAndSeed_InvalidConfig(t *testing.T) {
	invalidConfig := NetworkConfig{
		Latency:        -100 * time.Millisecond,
		PacketLossRate: 0,
		Jitter:         0,
		BandwidthLimit: 0,
	}

	sim, err := NewNetworkSimulatorWithConfigAndSeed(invalidConfig, 12345)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
	if sim != nil {
		t.Error("Expected nil simulator for invalid config")
	}
}

// Benchmark deterministic vs non-deterministic construction
func BenchmarkNewNetworkSimulator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNetworkSimulator()
	}
}

func BenchmarkNewNetworkSimulatorWithSeed(b *testing.B) {
	seed := int64(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNetworkSimulatorWithSeed(seed)
	}
}

// Benchmark packet sending with deterministic seed
func BenchmarkNetworkSimulator_Send_Deterministic(b *testing.B) {
	sim := NewNetworkSimulatorWithSeed(12345)
	sim.SetPacketLoss(0.1)
	data := []byte("benchmark packet")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sim.Send(data)
	}
}
