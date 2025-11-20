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
	config := parseCommandLineFlags()
	identity := createServerIdentity(config.serverName)
	ds := createDiscoverySystem(identity, config.listenAddr)

	setupPeerDiscoveryCallback(ds)
	addManualPeerIfSpecified(ds, config.addPeer)
	startDiscoverySystem(ds, config.mode)

	runDiscoveryLoop(ds, config.duration, config.verbose)
}

type discoveryConfig struct {
	serverName string
	listenAddr string
	addPeer    string
	mode       string
	duration   int
	verbose    bool
}

func parseCommandLineFlags() discoveryConfig {
	serverName := flag.String("name", "TestServer", "Server name for identity")
	listenAddr := flag.String("listen", ":8090", "UDP listen address for discovery")
	addPeer := flag.String("add-peer", "", "Manually add peer address (e.g., 192.168.1.100:8080)")
	mode := flag.String("mode", "listen", "Mode: listen, broadcast, or both")
	duration := flag.Int("duration", 60, "Duration to run (seconds), 0 for unlimited")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	return discoveryConfig{
		serverName: *serverName,
		listenAddr: *listenAddr,
		addPeer:    *addPeer,
		mode:       *mode,
		duration:   *duration,
		verbose:    *verbose,
	}
}

func createServerIdentity(serverName string) *federation.ServerIdentity {
	fmt.Printf("Creating server identity: %s\n", serverName)
	identity, err := federation.NewServerIdentity(serverName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating identity: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Server ID (fingerprint): %s\n", identity.GetFingerprint())
	return identity
}

func createDiscoverySystem(identity *federation.ServerIdentity, listenAddr string) *federation.DiscoverySystem {
	fmt.Printf("Listen address: %s\n\n", listenAddr)
	ds, err := federation.NewDiscoverySystem(identity, listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating discovery system: %v\n", err)
		os.Exit(1)
	}
	return ds
}

func setupPeerDiscoveryCallback(ds *federation.DiscoverySystem) {
	ds.OnPeerDiscovered(func(peer *federation.DiscoveredPeer) {
		fmt.Printf("[DISCOVERED] %s (%s) at %s - Features: %v, Hops: %d\n",
			peer.ServerName,
			peer.ServerID[:16]+"...",
			peer.Address,
			peer.Features,
			peer.Hops,
		)
	})
}

func addManualPeerIfSpecified(ds *federation.DiscoverySystem, addPeer string) {
	if addPeer == "" {
		return
	}

	fmt.Printf("Adding manual peer: %s\n", addPeer)
	err := ds.AddManualPeer("manual-peer-1", "Manual Peer", addPeer, "6.0.0", []string{"travel", "trade", "post"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding manual peer: %v\n", err)
	} else {
		fmt.Println("Manual peer added successfully")
	}
	fmt.Println()
}

func startDiscoverySystem(ds *federation.DiscoverySystem, mode string) {
	if mode != "listen" && mode != "both" {
		return
	}

	fmt.Println("Starting discovery system (listening for broadcasts)...")
	err := ds.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting discovery: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Discovery system running")
}

func runDiscoveryLoop(ds *federation.DiscoverySystem, duration int, verbose bool) {
	defer ds.Stop()

	sigChan := setupSignalHandling()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	durationTimer := createDurationTimer(duration)

	fmt.Println("\n=== Discovery System Active ===")
	fmt.Println("Press Ctrl+C to stop")

	for {
		select {
		case <-ticker.C:
			printPeerStatus(ds, verbose)
		case <-durationTimer:
			fmt.Println("\nDuration limit reached, shutting down...")
			return
		case sig := <-sigChan:
			handleGracefulShutdown(ds, sig)
			return
		}
	}
}

func setupSignalHandling() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	return sigChan
}

func createDurationTimer(duration int) <-chan time.Time {
	if duration > 0 {
		return time.After(time.Duration(duration) * time.Second)
	}
	return make(<-chan time.Time)
}

func printPeerStatus(ds *federation.DiscoverySystem, verbose bool) {
	peers := ds.GetPeers()
	fmt.Printf("[STATUS] %d peers discovered:\n", len(peers))
	if verbose {
		for i, peer := range peers {
			fmt.Printf("  %d. %s (%s) at %s\n", i+1, peer.ServerName, peer.ServerID[:16]+"...", peer.Address)
			fmt.Printf("     Version: %s, Features: %v, Hops: %d, LastSeen: %s ago\n",
				peer.Version, peer.Features, peer.Hops, time.Since(peer.LastSeen).Round(time.Second))
		}
	}
	fmt.Println()
}

func handleGracefulShutdown(ds *federation.DiscoverySystem, sig os.Signal) {
	fmt.Printf("\nReceived signal: %v, shutting down gracefully...\n", sig)
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
}
