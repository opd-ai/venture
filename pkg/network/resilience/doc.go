// Package resilience provides network resilience testing and simulation
// for validating multiplayer behavior under adverse network conditions.
//
// This package is part of Phase 64.1: Network Resilience Testing from
// the V10.0 Production Readiness Audit roadmap. It provides tools to
// simulate high latency, packet loss, jitter, bandwidth limits, and
// other network impairments to ensure the game remains playable and
// stable under realistic network conditions.
//
// # Overview
//
// The resilience package provides:
//   - Network simulator for simulating impairments (latency, packet loss, jitter)
//   - Metrics collector for tracking network performance
//   - Test scenarios for validating behavior under specific conditions
//   - Integration with existing client prediction and lag compensation systems
//
// # Network Simulator
//
// The NetworkSimulator applies configurable network impairments to
// outgoing packets, simulating real-world network conditions:
//
//	// For non-deterministic testing (time-based seed)
//	sim := resilience.NewNetworkSimulator()
//	sim.SetLatency(500 * time.Millisecond)
//	sim.SetPacketLoss(0.05) // 5% packet loss
//	sim.SetJitter(100 * time.Millisecond)
//	sim.SetBandwidthLimit(50 * 1024) // 50 KB/s
//
//	// Send packet through simulator
//	if err := sim.Send(packet); err != nil {
//	    // Packet was dropped or delayed
//	}
//
// For deterministic, reproducible testing use a fixed seed:
//
//	sim := resilience.NewNetworkSimulatorWithSeed(12345)
//	// Same seed = same packet drop/delay pattern
//
// # Metrics Collection
//
// The MetricsCollector tracks network performance over time:
//
//	collector := resilience.NewMetricsCollector()
//
//	// Record events
//	collector.RecordLatency(120 * time.Millisecond)
//	collector.RecordPacketLoss()
//	collector.RecordDesync()
//
//	// Get statistics
//	stats := collector.GetStats()
//	logrus.WithFields(logrus.Fields{
//	    "avg_latency": stats.AvgLatency,
//	    "packet_loss_rate": stats.PacketLossRate * 100,
//	    "desync_count": stats.DesyncCount,
//	}).Info("Network resilience metrics")
//
// # Test Scenarios
//
// Pre-defined test scenarios validate specific network conditions:
//
//	sim := resilience.NewNetworkSimulatorWithSeed(12345)
//	collector := resilience.NewMetricsCollector()
//	scenario := &resilience.HighLatencyScenario // 1000ms latency
//	result := resilience.RunScenario(context.Background(), scenario, sim, collector)
//
//	if result.Failed() {
//	    logrus.WithFields(logrus.Fields{
//	        "scenario": scenario.Name,
//	        "failure_reason": result.FailureReason,
//	    }).Error("Scenario failed")
//	}
//
// # Performance Targets
//
// The resilience testing validates these acceptance criteria:
//   - Playable at 200ms latency (smooth gameplay)
//   - Playable at 500ms latency (noticeable lag but functional)
//   - Playable at 1000ms latency (turn-based viable)
//   - Playable at 2000ms latency (degraded but stable)
//   - Gracefully degraded at 5000ms latency (minimal functionality)
//   - <10% misprediction rate at all latencies
//   - <1 desync per 1000 player-hours
//   - <100KB/s bandwidth per player
//   - Reconnection within 10 seconds
//
// # Integration
//
// The resilience package integrates with:
//   - pkg/network/prediction.go: Client-side prediction system
//   - pkg/network/lag_compensation.go: Server-side lag compensation
//   - pkg/network/client.go: Network client implementation
//   - pkg/engine: Entity synchronization and state management
//
// # Thread Safety
//
// All types in this package are thread-safe and can be used concurrently
// from multiple goroutines.
//
// # Determinism Exemption
//
// This package uses time.Now() for metrics timestamps, bandwidth tracking,
// and simulation timing (simulator.go, metrics.go, scenario.go). This is
// an intentional exemption from the project's strict deterministic procgen
// rule because:
//
//  1. This is testing infrastructure, not game content generation
//  2. Timestamps measure real-world elapsed time for performance metrics
//  3. Bandwidth limiting requires wall-clock time to enforce rate limits
//  4. Scenario execution needs real duration measurement for acceptance criteria
//
// For deterministic random behavior in packet drop/jitter simulation, use
// NewNetworkSimulatorWithSeed() which accepts a fixed seed for reproducible
// test scenarios.
package resilience
