package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation"
)

func main() {
	// Parse command-line flags
	serverName := flag.String("name", "TestServer", "Server name for identity")
	listenAddr := flag.String("listen", ":8090", "UDP listen address for discovery")
	addPeer := flag.String("add-peer", "", "Manually add peer address (e.g., 192.168.1.100:8080)")
	mode := flag.String("mode", "listen", "Mode: listen, broadcast, or both")
	duration := flag.Int("duration", 60, "Duration to run (seconds), 0 for unlimited")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	// Create server identity
	fmt.Printf("Creating server identity: %s\n", *serverName)
	identity, err := federation.NewServerIdentity(*serverName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating identity: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server ID (fingerprint): %s\n", identity.GetFingerprint())
	fmt.Printf("Listen address: %s\n", *listenAddr)
	fmt.Println()

	// Create discovery system
	ds, err := federation.NewDiscoverySystem(identity, *listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating discovery system: %v\n", err)
		os.Exit(1)
	}

	// Set discovery callback
	ds.OnPeerDiscovered(func(peer *federation.DiscoveredPeer) {
		fmt.Printf("[DISCOVERED] %s (%s) at %s - Features: %v, Hops: %d\n",
			peer.ServerName,
			peer.ServerID[:16]+"...", // Truncate for display
			peer.Address,
			peer.Features,
			peer.Hops,
		)
	})

	// Manually add peer if specified
	if *addPeer != "" {
		fmt.Printf("Adding manual peer: %s\n", *addPeer)
		err := ds.AddManualPeer(
			"manual-peer-1",
			"Manual Peer",
			*addPeer,
			"6.0.0",
			[]string{"travel", "trade", "post"},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding manual peer: %v\n", err)
		} else {
			fmt.Println("Manual peer added successfully")
		}
		fmt.Println()
	}

	// Start discovery based on mode
	if *mode == "listen" || *mode == "both" {
		fmt.Println("Starting discovery system (listening for broadcasts)...")
		err = ds.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting discovery: %v\n", err)
			os.Exit(1)
		}
		defer ds.Stop()
		fmt.Println("Discovery system running")
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ticker for periodic status updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Timer for duration limit
	var durationTimer <-chan time.Time
	if *duration > 0 {
		durationTimer = time.After(time.Duration(*duration) * time.Second)
	} else {
		durationTimer = make(<-chan time.Time) // Never fires
	}

	fmt.Println("\n=== Discovery System Active ===")
	fmt.Println("Press Ctrl+C to stop\n")

	// Main loop
	for {
		select {
		case <-ticker.C:
			// Print current peer list
			peers := ds.GetPeers()
			fmt.Printf("[STATUS] %d peers discovered:\n", len(peers))
			if *verbose {
				for i, peer := range peers {
					fmt.Printf("  %d. %s (%s) at %s\n",
						i+1,
						peer.ServerName,
						peer.ServerID[:16]+"...",
						peer.Address,
					)
					fmt.Printf("     Version: %s, Features: %v, Hops: %d, LastSeen: %s ago\n",
						peer.Version,
						peer.Features,
						peer.Hops,
						time.Since(peer.LastSeen).Round(time.Second),
					)
				}
			}
			fmt.Println()

		case <-durationTimer:
			fmt.Println("\nDuration limit reached, shutting down...")
			return

		case sig := <-sigChan:
			fmt.Printf("\nReceived signal: %v, shutting down gracefully...\n", sig)

			// Print final statistics
			peers := ds.GetPeers()
			fmt.Printf("\n=== Final Statistics ===\n")
			fmt.Printf("Total peers discovered: %d\n", len(peers))
			if len(peers) > 0 {
				fmt.Println("\nPeer list:")
				for i, peer := range peers {
					fmt.Printf("  %d. %s (%s)\n", i+1, peer.ServerName, peer.ServerID[:16]+"...")
					fmt.Printf("     Address: %s, Version: %s\n", peer.Address, peer.Version)
					fmt.Printf("     Features: %v, Hops: %d\n", peer.Features, peer.Hops)
				}
			}

			return
		}
	}
}
