// +build test

package main

// This example demonstrates the networking system.
// Run with: go run -tags test ./examples/network_demo.go

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
	
	"github.com/opd-ai/venture/pkg/network"
)

func main() {
	fmt.Println("=== Venture Networking Demo ===\n")
	
	// Demonstrate protocol serialization
	demonstrateProtocol()
	
	// Demonstrate client/server setup
	demonstrateClientServer()
}

func demonstrateProtocol() {
	fmt.Println("1. Binary Protocol Serialization")
	fmt.Println("----------------------------------")
	
	protocol := network.NewBinaryProtocol()
	
	// Create a state update with position data
	x, y := 123.45, 678.90
	posData := make([]byte, 16)
	binary.LittleEndian.PutUint64(posData[0:8], math.Float64bits(x))
	binary.LittleEndian.PutUint64(posData[8:16], math.Float64bits(y))
	
	update := &network.StateUpdate{
		Timestamp:      uint64(time.Now().UnixNano()),
		EntityID:       42,
		Priority:       200,
		SequenceNumber: 100,
		Components: []network.ComponentData{
			{Type: "position", Data: posData},
			{Type: "health", Data: []byte{100, 150}}, // current, max
		},
	}
	
	fmt.Printf("Original StateUpdate:\n")
	fmt.Printf("  Entity ID: %d\n", update.EntityID)
	fmt.Printf("  Sequence: %d\n", update.SequenceNumber)
	fmt.Printf("  Priority: %d\n", update.Priority)
	fmt.Printf("  Components: %d\n", len(update.Components))
	
	// Encode
	encoded, err := protocol.EncodeStateUpdate(update)
	if err != nil {
		fmt.Printf("❌ Encode failed: %v\n", err)
		return
	}
	fmt.Printf("✓ Encoded to %d bytes\n", len(encoded))
	
	// Decode
	decoded, err := protocol.DecodeStateUpdate(encoded)
	if err != nil {
		fmt.Printf("❌ Decode failed: %v\n", err)
		return
	}
	
	// Verify
	if decoded.EntityID != update.EntityID {
		fmt.Printf("❌ Entity ID mismatch\n")
		return
	}
	if decoded.SequenceNumber != update.SequenceNumber {
		fmt.Printf("❌ Sequence mismatch\n")
		return
	}
	
	decodedX := math.Float64frombits(binary.LittleEndian.Uint64(decoded.Components[0].Data[0:8]))
	decodedY := math.Float64frombits(binary.LittleEndian.Uint64(decoded.Components[0].Data[8:16]))
	
	fmt.Printf("✓ Decoded successfully\n")
	fmt.Printf("  Position: (%.2f, %.2f)\n", decodedX, decodedY)
	fmt.Printf("  Health: %d/%d\n", decoded.Components[1].Data[0], decoded.Components[1].Data[1])
	
	// Input command
	fmt.Println("\nInput Command:")
	cmd := &network.InputCommand{
		PlayerID:       123,
		Timestamp:      uint64(time.Now().UnixNano()),
		SequenceNumber: 50,
		InputType:      "move",
		Data:           []byte{1, 0}, // dx=1, dy=0
	}
	
	cmdEncoded, err := protocol.EncodeInputCommand(cmd)
	if err != nil {
		fmt.Printf("❌ Encode failed: %v\n", err)
		return
	}
	fmt.Printf("✓ Encoded input to %d bytes\n", len(cmdEncoded))
	
	cmdDecoded, err := protocol.DecodeInputCommand(cmdEncoded)
	if err != nil {
		fmt.Printf("❌ Decode failed: %v\n", err)
		return
	}
	fmt.Printf("✓ Decoded: Player %d, Type '%s', Data [%d, %d]\n",
		cmdDecoded.PlayerID, cmdDecoded.InputType, cmdDecoded.Data[0], cmdDecoded.Data[1])
	
	fmt.Println()
}

func demonstrateClientServer() {
	fmt.Println("2. Client/Server Configuration")
	fmt.Println("-------------------------------")
	
	// Server configuration
	serverConfig := network.DefaultServerConfig()
	fmt.Printf("Server Config:\n")
	fmt.Printf("  Address: %s\n", serverConfig.Address)
	fmt.Printf("  Max Players: %d\n", serverConfig.MaxPlayers)
	fmt.Printf("  Update Rate: %d Hz\n", serverConfig.UpdateRate)
	fmt.Printf("  Buffer Size: %d messages/player\n", serverConfig.BufferSize)
	
	// Create server
	server := network.NewServer(serverConfig)
	fmt.Printf("✓ Server created (not started)\n")
	fmt.Printf("  Running: %v\n", server.IsRunning())
	fmt.Printf("  Players: %d\n", server.GetPlayerCount())
	
	// Client configuration
	fmt.Println("\nClient Config:")
	clientConfig := network.DefaultClientConfig()
	fmt.Printf("  Server Address: %s\n", clientConfig.ServerAddress)
	fmt.Printf("  Connection Timeout: %v\n", clientConfig.ConnectionTimeout)
	fmt.Printf("  Ping Interval: %v\n", clientConfig.PingInterval)
	fmt.Printf("  Max Latency: %v\n", clientConfig.MaxLatency)
	
	// Create client
	client := network.NewClient(clientConfig)
	fmt.Printf("✓ Client created (not connected)\n")
	fmt.Printf("  Connected: %v\n", client.IsConnected())
	fmt.Printf("  Latency: %v\n", client.GetLatency())
	
	// Demonstrate broadcasting (without actual network)
	fmt.Println("\n3. State Broadcasting Simulation")
	fmt.Println("---------------------------------")
	
	// Simulate multiple entities updating
	for i := 1; i <= 3; i++ {
		update := &network.StateUpdate{
			Timestamp:  uint64(time.Now().UnixNano()),
			EntityID:   uint64(i),
			Priority:   128,
			Components: []network.ComponentData{
				{Type: "position", Data: []byte{byte(i * 10), byte(i * 20)}},
			},
		}
		
		// Server would broadcast this
		server.BroadcastStateUpdate(update)
		fmt.Printf("✓ Broadcast state for Entity %d (Seq: %d)\n", update.EntityID, update.SequenceNumber)
	}
	
	fmt.Println("\n4. Performance Summary")
	fmt.Println("----------------------")
	fmt.Println("Typical Performance:")
	fmt.Println("  StateUpdate encode: ~450 ns/op (2.2M ops/sec)")
	fmt.Println("  StateUpdate decode: ~590 ns/op (1.7M ops/sec)")
	fmt.Println("  InputCommand encode: ~210 ns/op (4.7M ops/sec)")
	fmt.Println("  InputCommand decode: ~280 ns/op (3.6M ops/sec)")
	fmt.Println("\nBandwidth Estimates (20 Hz update rate):")
	fmt.Println("  StateUpdate (3 components): ~75 bytes")
	fmt.Println("  32 entities × 20 updates/sec: ~48 KB/s downstream")
	fmt.Println("  InputCommand: ~35 bytes")
	fmt.Println("  20 inputs/sec: ~0.7 KB/s upstream")
	fmt.Println("  Total per player: ~49 KB/s (within 100 KB/s target ✓)")
	
	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nNote: This demo shows configuration and serialization.")
	fmt.Println("For actual client-server connection, start a dedicated server.")
	fmt.Println("\nNext Steps:")
	fmt.Println("  1. Implement client-side prediction")
	fmt.Println("  2. Add state interpolation")
	fmt.Println("  3. Create full multiplayer game demo")
}
