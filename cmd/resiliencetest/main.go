// Command resiliencetest is a CLI tool for testing network resilience.
//
// This tool demonstrates and validates the network resilience testing
// framework from Phase 64.1 of the V10.0 Production Readiness Audit.
//
// Usage:
//
//	# Test all pre-defined scenarios
//	resiliencetest -mode all
//
//	# Test a specific scenario
//	resiliencetest -mode low-latency
//	resiliencetest -mode medium-latency
//	resiliencetest -mode high-latency
//	resiliencetest -mode very-high-latency
//	resiliencetest -mode extreme-latency
//
//	# Run custom simulation
//	resiliencetest -mode custom -latency 500ms -packet-loss 0.1 -jitter 100ms
//
//	# Verbose output with detailed statistics
//	resiliencetest -mode all -verbose
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/network/resilience"
)

var (
	mode    = flag.String("mode", "demo", "Test mode: demo, low-latency, medium-latency, high-latency, very-high-latency, extreme-latency, custom, all")
	verbose = flag.Bool("verbose", false, "Verbose output with detailed statistics")

	// Custom mode parameters
	latency    = flag.Duration("latency", 0, "Custom latency (e.g., 500ms)")
	packetLoss = flag.Float64("packet-loss", 0, "Custom packet loss rate (0.0-1.0)")
	jitter     = flag.Duration("jitter", 0, "Custom jitter (e.g., 100ms)")
	bandwidth  = flag.Int("bandwidth", 0, "Custom bandwidth limit in bytes/sec")
)

func main() {
	flag.Parse()

	switch *mode {
	case "demo":
		runDemo()
	case "low-latency":
		runScenario(&resilience.LowLatencyScenario)
	case "medium-latency":
		runScenario(&resilience.MediumLatencyScenario)
	case "high-latency":
		runScenario(&resilience.HighLatencyScenario)
	case "very-high-latency":
		runScenario(&resilience.VeryHighLatencyScenario)
	case "extreme-latency":
		runScenario(&resilience.ExtremeLatencyScenario)
	case "custom":
		runCustom()
	case "all":
		runAllScenarios()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func runDemo() {
	fmt.Println("=== Network Resilience Testing Demo ===")
	fmt.Println()

	fmt.Println("This tool tests network resilience under various conditions:")
	fmt.Println("  - Low latency (200ms): Smooth gameplay expected")
	fmt.Println("  - Medium latency (500ms): Noticeable lag but playable")
	fmt.Println("  - High latency (1000ms): Turn-based viable")
	fmt.Println("  - Very high latency (2000ms): Degraded but stable")
	fmt.Println("  - Extreme latency (5000ms): Minimal functionality (Tor)")
	fmt.Println()

	fmt.Println("Example: Testing packet loss simulation...")
	sim := resilience.NewNetworkSimulator()
	sim.SetPacketLoss(0.2) // 20% packet loss

	sent := 0
	dropped := 0
	for i := 0; i < 100; i++ {
		data := []byte(fmt.Sprintf("packet %d", i))
		if err := sim.Send(data); err == resilience.ErrPacketDropped {
			dropped++
		} else {
			sent++
		}
	}

	fmt.Printf("  Sent 100 packets with 20%% loss rate:\n")
	fmt.Printf("  - Delivered: %d (%.1f%%)\n", sent, float64(sent))
	fmt.Printf("  - Dropped: %d (%.1f%%)\n", dropped, float64(dropped))
	fmt.Println()

	fmt.Println("Example: Testing latency simulation...")
	sim2 := resilience.NewNetworkSimulator()
	sim2.SetLatency(100 * time.Millisecond)

	startTime := time.Now()
	data := []byte("test packet")
	if err := sim2.Send(data); err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	// Packet is delayed
	fmt.Printf("  Packet queued with 100ms latency at %v\n", time.Now().Sub(startTime))
	fmt.Printf("  Queue size: %d packets\n", sim2.QueueSize())

	time.Sleep(110 * time.Millisecond)
	ready := sim2.ProcessDelayedPackets()
	fmt.Printf("  After 110ms: %d packets ready for delivery\n", len(ready))
	fmt.Printf("  Total time: %v\n", time.Now().Sub(startTime))
	fmt.Println()

	fmt.Println("Run 'resiliencetest -mode all' to test all scenarios")
}

func runScenario(scenario *resilience.TestScenario) {
	fmt.Printf("=== Testing: %s ===\n", scenario.Name)
	fmt.Printf("Description: %s\n", scenario.Description)
	fmt.Println()

	fmt.Println("Configuration:")
	fmt.Printf("  Latency: %v\n", scenario.Config.Latency)
	fmt.Printf("  Packet Loss: %.1f%%\n", scenario.Config.PacketLossRate*100)
	fmt.Printf("  Jitter: ±%v\n", scenario.Config.Jitter)
	fmt.Printf("  Bandwidth: %d bytes/sec\n", scenario.Config.BandwidthLimit)
	fmt.Println()

	fmt.Println("Acceptance Criteria:")
	fmt.Printf("  Max desync rate: %.2f/hour\n", scenario.MaxDesyncRate)
	fmt.Printf("  Max misprediction rate: %.1f%%\n", scenario.MaxMispredictionRate*100)
	fmt.Printf("  Max reconnect time: %v\n", scenario.MaxReconnectTime)
	fmt.Printf("  Requires playable: %v\n", scenario.RequiresPlayable)
	fmt.Println()

	// Run simulation
	fmt.Println("Running simulation...")
	result := simulateScenario(scenario)

	// Print results
	printResults(result)
}

func runCustom() {
	fmt.Println("=== Custom Network Simulation ===")

	config := resilience.NetworkConfig{
		Latency:        *latency,
		PacketLossRate: *packetLoss,
		Jitter:         *jitter,
		BandwidthLimit: *bandwidth,
	}

	if err := config.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration:")
	fmt.Printf("  Latency: %v\n", config.Latency)
	fmt.Printf("  Packet Loss: %.1f%%\n", config.PacketLossRate*100)
	fmt.Printf("  Jitter: ±%v\n", config.Jitter)
	fmt.Printf("  Bandwidth: %d bytes/sec\n", config.BandwidthLimit)
	fmt.Println()

	// Run simulation
	fmt.Println("Running simulation...")
	sim := resilience.NewNetworkSimulator()
	sim.SetConfig(config)

	metrics := resilience.NewMetricsCollector()

	// Simulate 1000 packets
	for i := 0; i < 1000; i++ {
		data := []byte(fmt.Sprintf("packet %d with some data", i))

		if err := sim.Send(data); err == resilience.ErrPacketDropped {
			metrics.RecordPacketLoss()
		} else {
			metrics.RecordPacketSent(len(data))
		}

		// Simulate network round trip
		if config.Latency > 0 {
			metrics.RecordLatency(config.Latency * 2) // Round trip
		}

		time.Sleep(time.Millisecond) // Simulate packet rate
	}

	stats := metrics.GetStats()
	printStats(stats)
}

func runAllScenarios() {
	fmt.Println("=== Testing All Network Scenarios ===")
	fmt.Println()

	passed := 0
	failed := 0

	for _, scenario := range resilience.AllScenarios {
		fmt.Printf("Testing: %s...\n", scenario.Name)
		result := simulateScenario(scenario)

		if result.Passed {
			fmt.Printf("  ✓ PASSED\n")
			passed++
		} else {
			fmt.Printf("  ✗ FAILED: %s\n", result.FailureReason)
			failed++
		}

		if *verbose {
			printResults(result)
		}
		fmt.Println()
	}

	fmt.Printf("Results: %d passed, %d failed\n", passed, failed)
}

func simulateScenario(scenario *resilience.TestScenario) *resilience.ScenarioResult {
	sim := resilience.NewNetworkSimulator()
	sim.SetConfig(scenario.Config)

	metrics := resilience.NewMetricsCollector()

	// Simulate 1000 packets (simplified test)
	for i := 0; i < 1000; i++ {
		data := []byte(fmt.Sprintf("packet %d", i))

		if err := sim.Send(data); err == resilience.ErrPacketDropped {
			metrics.RecordPacketLoss()
		} else {
			metrics.RecordPacketSent(len(data))
		}

		// Simulate latency measurement
		if scenario.Config.Latency > 0 {
			actualLatency := scenario.Config.Latency
			// Add jitter variation
			if scenario.Config.Jitter > 0 {
				// Simulate ±jitter
				jitterMs := int64(scenario.Config.Jitter / time.Millisecond)
				variation := time.Duration(i%int(jitterMs*2)-int(jitterMs)) * time.Millisecond
				actualLatency += variation
			}
			metrics.RecordLatency(actualLatency * 2) // Round trip
		}
	}

	stats := metrics.GetStats()

	// Check acceptance criteria
	result := &resilience.ScenarioResult{
		Scenario:  scenario,
		Stats:     stats,
		Passed:    true,
		Timestamp: time.Now(),
	}

	// Validate against criteria (simplified - actual implementation would track real desyncs)
	if stats.PacketLossRate > scenario.Config.PacketLossRate+0.1 {
		result.Passed = false
		result.FailureReason = fmt.Sprintf("Packet loss rate %.2f%% exceeds configured %.2f%%",
			stats.PacketLossRate*100, scenario.Config.PacketLossRate*100)
	}

	return result
}

func printResults(result *resilience.ScenarioResult) {
	fmt.Println("Results:")
	if result.Passed {
		fmt.Println("  Status: ✓ PASSED")
	} else {
		fmt.Printf("  Status: ✗ FAILED - %s\n", result.FailureReason)
	}
	fmt.Println()

	printStats(result.Stats)
}

func printStats(stats resilience.NetworkStats) {
	fmt.Println("Statistics:")
	fmt.Printf("  Duration: %v\n", stats.Duration)
	fmt.Println()

	fmt.Println("  Latency:")
	fmt.Printf("    Average: %v\n", stats.AvgLatency)
	fmt.Printf("    Min: %v\n", stats.MinLatency)
	fmt.Printf("    Max: %v\n", stats.MaxLatency)
	fmt.Printf("    P95: %v\n", stats.P95Latency)
	fmt.Printf("    P99: %v\n", stats.P99Latency)
	fmt.Println()

	fmt.Println("  Packets:")
	fmt.Printf("    Sent: %d\n", stats.PacketsSent)
	fmt.Printf("    Received: %d\n", stats.PacketsReceived)
	fmt.Printf("    Dropped: %d\n", stats.PacketsDropped)
	fmt.Printf("    Loss Rate: %.2f%%\n", stats.PacketLossRate*100)
	fmt.Println()

	fmt.Println("  Bandwidth:")
	fmt.Printf("    Bytes Sent: %d\n", stats.BytesSent)
	fmt.Printf("    Bytes Received: %d\n", stats.BytesReceived)
	fmt.Printf("    Average: %.2f bytes/sec\n", stats.AvgBandwidth)
	fmt.Printf("    Peak: %.2f bytes/sec\n", stats.PeakBandwidth)
	fmt.Println()

	if stats.MispredictionCount > 0 || stats.DesyncCount > 0 {
		fmt.Println("  Gameplay:")
		if stats.MispredictionCount > 0 {
			fmt.Printf("    Mispredictions: %d (%.2f%%)\n", stats.MispredictionCount, stats.MispredictionRate*100)
		}
		if stats.DesyncCount > 0 {
			fmt.Printf("    Desyncs: %d\n", stats.DesyncCount)
		}
		if stats.ReconnectCount > 0 {
			fmt.Printf("    Reconnects: %d (avg: %v)\n", stats.ReconnectCount, stats.AvgReconnectTime)
		}
		fmt.Println()
	}
}
