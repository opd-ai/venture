// Package resilience types defines all data structures for network resilience
// testing including NetworkConfig, NetworkStats, TestScenario, ScenarioResult,
// Packet, and pre-defined test scenarios for various latency conditions.
package resilience

import (
	"fmt"
	"time"
)

// NetworkConfig defines network impairment parameters for simulation.
type NetworkConfig struct {
	// Latency is the simulated one-way latency (e.g., 200ms)
	Latency time.Duration

	// PacketLossRate is the probability of dropping a packet (0.0-1.0)
	PacketLossRate float64

	// Jitter is the variance in latency (±jitter)
	Jitter time.Duration

	// BandwidthLimit is the maximum bytes per second (0 = unlimited)
	BandwidthLimit int
}

// Validate checks if the network config is valid.
func (nc *NetworkConfig) Validate() error {
	if nc.Latency < 0 {
		return fmt.Errorf("latency cannot be negative: %v", nc.Latency)
	}
	if nc.PacketLossRate < 0 || nc.PacketLossRate > 1 {
		return fmt.Errorf("packet loss rate must be 0-1: %f", nc.PacketLossRate)
	}
	if nc.Jitter < 0 {
		return fmt.Errorf("jitter cannot be negative: %v", nc.Jitter)
	}
	if nc.BandwidthLimit < 0 {
		return fmt.Errorf("bandwidth limit cannot be negative: %d", nc.BandwidthLimit)
	}
	return nil
}

// NetworkStats represents collected network performance statistics.
type NetworkStats struct {
	// Latency measurements
	AvgLatency time.Duration
	MinLatency time.Duration
	MaxLatency time.Duration
	P95Latency time.Duration // 95th percentile
	P99Latency time.Duration // 99th percentile

	// Packet statistics
	PacketsSent     uint64
	PacketsReceived uint64
	PacketsDropped  uint64
	PacketLossRate  float64

	// Bandwidth statistics
	BytesSent     uint64
	BytesReceived uint64
	AvgBandwidth  float64 // bytes per second
	PeakBandwidth float64

	// Gameplay statistics
	MispredictionCount uint64
	MispredictionRate  float64
	DesyncCount        uint64
	ReconnectCount     uint64
	AvgReconnectTime   time.Duration

	// Time period
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// TestScenario defines a network resilience test scenario.
type TestScenario struct {
	Name        string
	Description string
	Config      NetworkConfig
	Duration    time.Duration

	// Acceptance criteria
	MaxDesyncRate        float64 // desyncs per hour
	MaxMispredictionRate float64 // 0.0-1.0
	MaxReconnectTime     time.Duration
	RequiresPlayable     bool // must maintain >20 FPS
}

// ScenarioResult contains the results of running a test scenario.
type ScenarioResult struct {
	Scenario      *TestScenario
	Stats         NetworkStats
	Passed        bool
	FailureReason string
	Timestamp     time.Time
}

// Failed returns true if the scenario failed acceptance criteria.
func (sr *ScenarioResult) Failed() bool {
	return !sr.Passed
}

// Packet represents a network packet for simulation.
type Packet struct {
	Data      []byte
	Timestamp time.Time
	Size      int
	Dropped   bool
}

// Pre-defined test scenarios based on Phase 64.1 requirements.
// Each scenario represents a different network condition tier with
// specific acceptance criteria for validation testing.
//
// Scenarios are ordered from best to worst network conditions:
//   - LowLatencyScenario: Optimal conditions, smooth gameplay
//   - MediumLatencyScenario: Typical home network, playable with minor lag
//   - HighLatencyScenario: Degraded conditions, turn-based viable
//   - VeryHighLatencyScenario: Poor conditions, degraded but stable
//   - ExtremeLatencyScenario: Tor/satellite, minimal functionality
var (
	// LowLatencyScenario represents optimal network conditions (200ms latency).
	// Use this scenario to validate smooth, responsive gameplay with minimal lag.
	// Acceptance criteria are strict: <0.1 desyncs/hour, <5% mispredictions.
	// Expected use case: Local/regional servers, fiber/cable connections.
	LowLatencyScenario = TestScenario{
		Name:        "Low Latency",
		Description: "200ms latency, 1% packet loss - should be smooth",
		Config: NetworkConfig{
			Latency:        200 * time.Millisecond,
			PacketLossRate: 0.01,
			Jitter:         20 * time.Millisecond,
			BandwidthLimit: 100 * 1024, // 100 KB/s
		},
		Duration:             5 * time.Minute,
		MaxDesyncRate:        0.1,  // 0.1 desyncs/hour
		MaxMispredictionRate: 0.05, // 5%
		MaxReconnectTime:     5 * time.Second,
		RequiresPlayable:     true,
	}

	// MediumLatencyScenario represents typical home network conditions (500ms latency).
	// Use this scenario to validate playable but laggy gameplay.
	// Acceptance criteria are moderate: <0.5 desyncs/hour, <10% mispredictions.
	// Expected use case: Mobile networks, international servers, VPN connections.
	MediumLatencyScenario = TestScenario{
		Name:        "Medium Latency",
		Description: "500ms latency, 5% packet loss - noticeable but playable",
		Config: NetworkConfig{
			Latency:        500 * time.Millisecond,
			PacketLossRate: 0.05,
			Jitter:         100 * time.Millisecond,
			BandwidthLimit: 50 * 1024, // 50 KB/s
		},
		Duration:             5 * time.Minute,
		MaxDesyncRate:        0.5,
		MaxMispredictionRate: 0.10,
		MaxReconnectTime:     10 * time.Second,
		RequiresPlayable:     true,
	}

	// HighLatencyScenario represents degraded network conditions (1000ms latency).
	// Use this scenario to validate turn-based gameplay viability.
	// Acceptance criteria are relaxed: <1.0 desyncs/hour, <15% mispredictions.
	// Expected use case: Satellite internet, congested networks, cross-continent servers.
	HighLatencyScenario = TestScenario{
		Name:        "High Latency",
		Description: "1000ms latency, 10% packet loss - turn-based viable",
		Config: NetworkConfig{
			Latency:        1000 * time.Millisecond,
			PacketLossRate: 0.10,
			Jitter:         200 * time.Millisecond,
			BandwidthLimit: 20 * 1024, // 20 KB/s
		},
		Duration:             5 * time.Minute,
		MaxDesyncRate:        1.0,
		MaxMispredictionRate: 0.15,
		MaxReconnectTime:     15 * time.Second,
		RequiresPlayable:     false,
	}

	// VeryHighLatencyScenario represents poor network conditions (2000ms latency).
	// Use this scenario to validate degraded but stable operation.
	// Acceptance criteria are permissive: <2.0 desyncs/hour, <20% mispredictions.
	// Expected use case: Remote areas, multi-hop VPN, severely congested networks.
	VeryHighLatencyScenario = TestScenario{
		Name:        "Very High Latency",
		Description: "2000ms latency, 20% packet loss - degraded but stable",
		Config: NetworkConfig{
			Latency:        2000 * time.Millisecond,
			PacketLossRate: 0.20,
			Jitter:         500 * time.Millisecond,
			BandwidthLimit: 10 * 1024, // 10 KB/s
		},
		Duration:             5 * time.Minute,
		MaxDesyncRate:        2.0,
		MaxMispredictionRate: 0.20,
		MaxReconnectTime:     20 * time.Second,
		RequiresPlayable:     false,
	}

	// ExtremeLatencyScenario represents worst-case network conditions (5000ms latency).
	// Use this scenario to validate graceful degradation under extreme conditions.
	// Acceptance criteria are very permissive: <5.0 desyncs/hour, <30% mispredictions.
	// Expected use case: Tor/onion routing, extremely remote locations, testing limits.
	ExtremeLatencyScenario = TestScenario{
		Name:        "Extreme Latency (Tor)",
		Description: "5000ms latency, 20% packet loss - minimal functionality",
		Config: NetworkConfig{
			Latency:        5000 * time.Millisecond,
			PacketLossRate: 0.20,
			Jitter:         1000 * time.Millisecond,
			BandwidthLimit: 10 * 1024, // 10 KB/s
		},
		Duration:             5 * time.Minute,
		MaxDesyncRate:        5.0,
		MaxMispredictionRate: 0.30,
		MaxReconnectTime:     30 * time.Second,
		RequiresPlayable:     false,
	}

	// All scenarios for batch testing
	AllScenarios = []*TestScenario{
		&LowLatencyScenario,
		&MediumLatencyScenario,
		&HighLatencyScenario,
		&VeryHighLatencyScenario,
		&ExtremeLatencyScenario,
	}
)
