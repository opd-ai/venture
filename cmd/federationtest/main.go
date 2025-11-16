package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/network/federation"
)

func main() {
	mode := flag.String("mode", "list", "Test mode: list, identity, handshake, verify, negotiate")
	serverName := flag.String("name", "TestServer", "Server name for identity generation")
	version := flag.String("version", "6.0.0", "Protocol version")
	features := flag.String("features", "travel,trade,post", "Comma-separated features")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	switch *mode {
	case "list":
		listModes()
	case "identity":
		testIdentity(*serverName, *verbose)
	case "handshake":
		testHandshake(*serverName, *version, *features, *verbose)
	case "verify":
		testVerify(*serverName, *version, *features, *verbose)
	case "negotiate":
		testNegotiate(*features, *verbose)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		listModes()
		os.Exit(1)
	}
}

func listModes() {
	fmt.Println("Available test modes:")
	fmt.Println("  list       - List all test modes")
	fmt.Println("  identity   - Generate and display server identity")
	fmt.Println("  handshake  - Create and display handshake")
	fmt.Println("  verify     - Create handshake and verify it")
	fmt.Println("  negotiate  - Test feature negotiation")
	fmt.Println()
	fmt.Println("Usage: federationtest -mode <mode> [options]")
	fmt.Println()
	flag.PrintDefaults()
}

func testIdentity(serverName string, verbose bool) {
	fmt.Printf("=== Server Identity Generation ===\n\n")

	identity, err := federation.NewServerIdentity(serverName)
	if err != nil {
		log.Fatalf("Failed to create identity: %v", err)
	}

	fmt.Printf("Server Name: %s\n", identity.ServerName)
	fmt.Printf("Server ID (Fingerprint): %s\n", identity.GetFingerprint())
	fmt.Printf("Created: %s\n", identity.Created.Format("2006-01-02 15:04:05"))

	if verbose {
		fmt.Printf("\nPublic Key (hex): %s\n", hex.EncodeToString(identity.PublicKey))
		fmt.Printf("Public Key Size: %d bytes\n", len(identity.PublicKey))
		fmt.Printf("Private Key Size: %d bytes\n", len(identity.PrivateKey))
	}

	fmt.Println("\n✓ Identity generation successful")
}

func testHandshake(serverName, version, featuresStr string, verbose bool) {
	fmt.Printf("=== Handshake Creation ===\n\n")

	identity, err := federation.NewServerIdentity(serverName)
	if err != nil {
		log.Fatalf("Failed to create identity: %v", err)
	}

	features := parseFeatures(featuresStr)

	handshake, err := identity.CreateHandshake(version, features, federation.TrustVerified)
	if err != nil {
		log.Fatalf("Failed to create handshake: %v", err)
	}

	fmt.Printf("Server ID: %s\n", handshake.ServerID)
	fmt.Printf("Server Name: %s\n", handshake.ServerName)
	fmt.Printf("Version: %s\n", handshake.Version)
	fmt.Printf("Trust Level: %s\n", handshake.TrustLevel)
	fmt.Printf("Features: %s\n", strings.Join(handshake.Features, ", "))
	fmt.Printf("Timestamp: %d\n", handshake.Timestamp)
	fmt.Printf("Nonce Size: %d bytes\n", len(handshake.Nonce))
	fmt.Printf("Signature Size: %d bytes\n", len(handshake.Signature))

	if verbose {
		fmt.Printf("\nNonce (hex): %s\n", hex.EncodeToString(handshake.Nonce))
		fmt.Printf("Signature (hex): %s\n", hex.EncodeToString(handshake.Signature))
		fmt.Printf("Public Key (hex): %s\n", hex.EncodeToString(handshake.PublicKey))
	}

	fmt.Println("\n✓ Handshake creation successful")
}

func testVerify(serverName, version, featuresStr string, verbose bool) {
	fmt.Printf("=== Handshake Verification ===\n\n")

	// Create identity and handshake
	identity, err := federation.NewServerIdentity(serverName)
	if err != nil {
		log.Fatalf("Failed to create identity: %v", err)
	}

	features := parseFeatures(featuresStr)

	handshake, err := identity.CreateHandshake(version, features, federation.TrustTrusted)
	if err != nil {
		log.Fatalf("Failed to create handshake: %v", err)
	}

	fmt.Println("Created handshake:")
	fmt.Printf("  Server: %s\n", handshake.ServerName)
	fmt.Printf("  Version: %s\n", handshake.Version)
	fmt.Printf("  Trust: %s\n", handshake.TrustLevel)

	// Verify the handshake
	fmt.Println("\nVerifying handshake...")

	err = federation.VerifyHandshake(handshake)
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	fmt.Println("✓ Signature verification passed")
	fmt.Println("✓ Fingerprint matches public key")
	fmt.Println("✓ Timestamp within acceptable range")
	fmt.Println("✓ All fields valid")

	// Test with HandshakeManager for replay detection
	fmt.Println("\nTesting replay prevention...")
	manager := federation.NewHandshakeManager(identity)

	err = manager.ProcessHandshake(handshake)
	if err != nil {
		log.Fatalf("First processing failed: %v", err)
	}
	fmt.Println("✓ First handshake processing successful")

	err = manager.ProcessHandshake(handshake)
	if err == nil {
		log.Fatalf("Replay attack not detected!")
	}
	fmt.Printf("✓ Replay attack detected: %v\n", err)

	fmt.Println("\n✓ Verification test successful")
}

func testNegotiate(featuresStr string, verbose bool) {
	fmt.Printf("=== Feature Negotiation ===\n\n")

	ourFeatures := parseFeatures(featuresStr)
	theirFeatures := []string{"travel", "post", "bounty"}

	fmt.Printf("Our Features: %s\n", strings.Join(ourFeatures, ", "))
	fmt.Printf("Their Features: %s\n", strings.Join(theirFeatures, ", "))

	common := federation.NegotiateFeatures(ourFeatures, theirFeatures)

	fmt.Printf("\nNegotiated Features: %s\n", strings.Join(common, ", "))
	fmt.Printf("Common Count: %d\n", len(common))

	// Test version compatibility
	fmt.Println("\nVersion Compatibility:")
	testVersionPair := []struct {
		ours   string
		theirs string
	}{
		{"6.0.0", "6.0.0"},
		{"6.0.0", "6.1.0"},
		{"6.0.0", "5.0.0"},
		{"6.0.0", "7.0.0"},
	}

	for _, test := range testVersionPair {
		compatible := federation.IsCompatibleVersion(test.ours, test.theirs)
		result := "✗ Incompatible"
		if compatible {
			result = "✓ Compatible"
		}
		fmt.Printf("  %s vs %s: %s\n", test.ours, test.theirs, result)
	}

	fmt.Println("\n✓ Negotiation test successful")
}

func parseFeatures(featuresStr string) []string {
	if featuresStr == "" {
		return nil
	}
	parts := strings.Split(featuresStr, ",")
	var features []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			features = append(features, trimmed)
		}
	}
	return features
}
