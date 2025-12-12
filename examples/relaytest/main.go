// relaytest demonstrates the P2P relay network and NAT traversal system.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/webrtc"
)

func main() {
	mode := flag.String("mode", "traversal", "Mode: traversal, relay, stun")
	stunServers := flag.String("stun", "stun:stun.l.google.com:19302", "Comma-separated STUN servers")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	switch *mode {
	case "traversal":
		testNATTraversal(*stunServers, *verbose)
	case "relay":
		testRelayManager(*verbose)
	case "stun":
		testSTUNClient(*stunServers, *verbose)
	default:
		log.Fatalf("Unknown mode: %s (use: traversal, relay, stun)", *mode)
	}
}

func testNATTraversal(stunServers string, verbose bool) {
	fmt.Println("=== NAT Traversal Test ===")
	fmt.Println()

	// Create relay manager with multiple relays
	rm := webrtc.NewRelayManager(webrtc.StrategyLowestLatency)
	defer rm.Close()

	// Add sample TURN relays
	relays := []struct {
		id       string
		url      string
		region   string
		latency  time.Duration
		capacity int
	}{
		{"relay1", "turn:relay1.example.com:3478", "us-east", 30 * time.Millisecond, 100},
		{"relay2", "turn:relay2.example.com:3478", "us-west", 50 * time.Millisecond, 100},
		{"relay3", "turn:relay3.example.com:3478", "eu-central", 80 * time.Millisecond, 100},
	}

	for _, r := range relays {
		node := webrtc.NewRelayNode(r.id, r.url, "user", "pass", r.region, r.capacity)
		node.UpdateLatency(r.latency)
		node.UpdateBandwidth(5 * 1024 * 1024) // 5 MB/s
		rm.AddRelay(node)
		if verbose {
			fmt.Printf("Added relay: %s (%s, latency: %v)\n", r.id, r.region, r.latency)
		}
	}

	// Create NAT traversal coordinator
	stunList := []string{stunServers}
	nt := webrtc.NewNATTraversal(stunList, rm)

	fmt.Println("\nAttempting NAT traversal...")
	fmt.Println()

	// Attempt multiple connections
	attempts := 5
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		result, err := nt.EstablishConnection(ctx)
		cancel()

		if err != nil {
			fmt.Printf("Attempt %d: FAILED - %v\n", i+1, err)
			if verbose && result != nil {
				fmt.Printf("  Setup time: %v\n", result.SetupTime)
			}
		} else {
			fmt.Printf("Attempt %d: SUCCESS via %s\n", i+1, result.Method)
			if verbose {
				fmt.Printf("  Setup time: %v\n", result.SetupTime)
				if result.Method == webrtc.MethodSTUN {
					fmt.Printf("  Public IP: %s:%d\n", result.PublicIP, result.PublicPort)
					fmt.Printf("  NAT Type: %v\n", result.NATType)
				}
				if result.Method == webrtc.MethodTURN && result.RelayNode != nil {
					fmt.Printf("  Relay: %s (%s)\n", result.RelayNode.ID, result.RelayNode.Region)
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Display statistics
	fmt.Println("\n=== NAT Traversal Statistics ===")
	stats := nt.GetStats()
	fmt.Printf("Total attempts: %d\n", stats.TotalAttempts)
	fmt.Printf("Direct success: %d\n", stats.DirectSuccess)
	fmt.Printf("STUN success: %d\n", stats.STUNSuccess)
	fmt.Printf("TURN success: %d\n", stats.TURNSuccess)
	fmt.Printf("Failures: %d\n", stats.Failures)
	fmt.Printf("Success rate: %.1f%%\n", stats.SuccessRate*100)
	fmt.Printf("TURN fallback rate: %.1f%%\n", stats.TURNFallbackRate*100)
	fmt.Printf("Average setup time: %v\n", stats.AverageSetupTime)

	// Display relay statistics
	fmt.Println("\n=== Relay Statistics ===")
	relayStats := rm.GetRelayStats()
	for _, s := range relayStats {
		fmt.Printf("\nRelay: %s (%s)\n", s.ID, s.Region)
		fmt.Printf("  Active connections: %d / max %d (%.1f%% utilization)\n",
			s.ActiveConnections, 100, s.Utilization*100)
		fmt.Printf("  Total connections: %d\n", s.TotalConnections)
		fmt.Printf("  Bytes relayed: %d\n", s.BytesRelayed)
		fmt.Printf("  Latency: %v\n", s.Latency)
		fmt.Printf("  Bandwidth: %d MB/s\n", s.Bandwidth/(1024*1024))
		fmt.Printf("  Healthy: %v\n", s.Healthy)
	}
}

func testRelayManager(verbose bool) {
	fmt.Println("=== Relay Manager Test ===")
	fmt.Println()

	// Test all selection strategies
	strategies := []webrtc.SelectionStrategy{
		webrtc.StrategyLowestLatency,
		webrtc.StrategyHighestBandwidth,
		webrtc.StrategyLowestUtilization,
		webrtc.StrategyRoundRobin,
	}

	for _, strategy := range strategies {
		fmt.Printf("\nTesting strategy: %s\n", strategy)

		rm := webrtc.NewRelayManager(strategy)

		// Add relays with different characteristics
		relay1 := webrtc.NewRelayNode("fast", "turn:fast.example.com:3478", "user", "pass", "local", 10)
		relay1.UpdateLatency(10 * time.Millisecond)
		relay1.UpdateBandwidth(10 * 1024 * 1024)
		for i := 0; i < 2; i++ {
			relay1.Acquire()
		}

		relay2 := webrtc.NewRelayNode("slow", "turn:slow.example.com:3478", "user", "pass", "remote", 10)
		relay2.UpdateLatency(100 * time.Millisecond)
		relay2.UpdateBandwidth(1 * 1024 * 1024)
		for i := 0; i < 8; i++ {
			relay2.Acquire()
		}

		relay3 := webrtc.NewRelayNode("medium", "turn:medium.example.com:3478", "user", "pass", "regional", 10)
		relay3.UpdateLatency(50 * time.Millisecond)
		relay3.UpdateBandwidth(5 * 1024 * 1024)
		for i := 0; i < 5; i++ {
			relay3.Acquire()
		}

		rm.AddRelay(relay1)
		rm.AddRelay(relay2)
		rm.AddRelay(relay3)

		// Select relay
		selected, err := rm.SelectRelay()
		if err != nil {
			fmt.Printf("  Selection failed: %v\n", err)
		} else {
			stats := selected.GetStats()
			fmt.Printf("  Selected: %s (latency: %v, bandwidth: %d MB/s, util: %.1f%%)\n",
				selected.ID, stats.Latency, stats.Bandwidth/(1024*1024), stats.Utilization*100)
		}

		rm.Close()
	}
}

func testSTUNClient(stunServers string, verbose bool) {
	fmt.Println("=== STUN Client Test ===")
	fmt.Println()

	client := webrtc.NewSTUNClient([]string{stunServers})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get public address
	fmt.Println("Querying STUN server for public address...")
	resp, err := client.GetPublicAddress(ctx)
	if err != nil {
		log.Fatalf("STUN query failed: %v", err)
	}

	fmt.Printf("\nPublic IP: %s\n", resp.PublicIP)
	fmt.Printf("Public Port: %d\n", resp.PublicPort)
	fmt.Printf("STUN Server: %s\n", resp.Server)
	fmt.Printf("RTT: %v\n", resp.RTT)
	fmt.Printf("NAT Type: %v\n", resp.NATType)

	// Test caching
	fmt.Println("\nTesting cache (second request)...")
	start := time.Now()
	resp2, err := client.GetPublicAddress(ctx)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Cached query failed: %v", err)
	}

	fmt.Printf("Response time: %v (cached: %v)\n", elapsed, elapsed < 1*time.Millisecond)
	fmt.Printf("Same result: %v\n", resp.PublicIP.Equal(resp2.PublicIP) && resp.PublicPort == resp2.PublicPort)

	// Display statistics
	fmt.Println("\n=== STUN Statistics ===")
	stats := client.GetStats()
	fmt.Printf("Total requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful requests: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed requests: %d\n", stats.FailedRequests)
	fmt.Printf("Success rate: %.1f%%\n", stats.SuccessRate*100)
	fmt.Printf("Cache valid: %v\n", stats.CacheValid)
}
