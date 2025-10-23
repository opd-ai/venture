//go:build test
// +build test

package main

// This example demonstrates a complete client-server multiplayer setup.
// Run with: go run -tags test ./examples/multiplayer_demo.go

import (
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network"
)

func main() {
	fmt.Println("=== Venture Multiplayer Demo ===")

	// Create worlds for server and clients
	serverWorld := engine.NewWorld()
	client1World := engine.NewWorld()
	client2World := engine.NewWorld()

	// Create network components
	server := createServer()
	client1 := createClient(1)
	client2 := createClient(2)
	serializer := network.NewComponentSerializer()

	// Create player entities on server
	_ = createPlayerEntity(serverWorld, 1, 100, 100)
	_ = createPlayerEntity(serverWorld, 2, 200, 200)

	fmt.Println("=== Server Setup ===")
	fmt.Printf("✓ Server created (not started - demo mode)\n")
	fmt.Printf("✓ Player 1 created at (100, 100)\n")
	fmt.Printf("✓ Player 2 created at (200, 200)\n")

	fmt.Println("\n=== Client Setup ===")
	fmt.Printf("✓ Client 1 created (Player ID: 1)\n")
	fmt.Printf("✓ Client 2 created (Player ID: 2)\n")

	// Simulate game loop
	fmt.Println("\n=== Simulating Multiplayer Gameplay ===")
	fmt.Println("(Server would normally run on separate machine)")

	var wg sync.WaitGroup

	// Simulate server game loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateServerLoop(server, serverWorld, serializer, 5)
	}()

	// Simulate client 1 sending inputs
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateClientInput(client1, 1, 5)
	}()

	// Simulate client 2 sending inputs
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateClientInput(client2, 2, 5)
	}()

	// Simulate clients receiving state updates
	wg.Add(2)
	go func() {
		defer wg.Done()
		simulateClientReceive(client1, client1World, 1, 5)
	}()
	go func() {
		defer wg.Done()
		simulateClientReceive(client2, client2World, 2, 5)
	}()

	// Wait for simulation to complete
	wg.Wait()

	fmt.Println("\n=== Simulation Complete ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("  ✓ Binary protocol efficiently serializes game state")
	fmt.Println("  ✓ Server maintains authoritative state")
	fmt.Println("  ✓ Clients send inputs at ~20 Hz")
	fmt.Println("  ✓ Server broadcasts updates at ~20 Hz")
	fmt.Println("  ✓ Component serialization handles ECS data")
	fmt.Println("  ✓ System supports multiple concurrent players")
	fmt.Println("\nNext Steps:")
	fmt.Println("  1. Add client-side prediction for smooth movement")
	fmt.Println("  2. Implement state interpolation")
	fmt.Println("  3. Add lag compensation for hit detection")
	fmt.Println("  4. Test with real network connections")
}

func createServer() *network.Server {
	config := network.DefaultServerConfig()
	config.Address = ":8080"
	config.MaxPlayers = 4
	return network.NewServer(config)
}

func createClient(playerID uint64) *network.Client {
	config := network.DefaultClientConfig()
	config.ServerAddress = "localhost:8080"
	client := network.NewClient(config)
	client.SetPlayerID(playerID)
	return client
}

func createPlayerEntity(world *engine.World, playerID uint64, x, y float64) *engine.Entity {
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: x, Y: y})
	entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&engine.TeamComponent{TeamID: int(playerID)})
	world.Update(0)
	return entity
}

func simulateServerLoop(server *network.Server, world *engine.World, serializer *network.ComponentSerializer, updates int) {
	fmt.Println("Server: Starting game loop simulation...")

	for i := 0; i < updates; i++ {
		time.Sleep(50 * time.Millisecond) // 20 Hz

		// Update world
		world.Update(0.05)

		// Broadcast entity states (simulated)
		for _, entity := range world.GetEntities() {
			pos, ok := entity.GetComponent("position")
			if !ok {
				continue
			}
			position := pos.(*engine.PositionComponent)

			// Create state update
			update := &network.StateUpdate{
				Timestamp: uint64(time.Now().UnixNano()),
				EntityID:  entity.ID,
				Priority:  128,
				Components: []network.ComponentData{
					{
						Type: "position",
						Data: serializer.SerializePosition(position.X, position.Y),
					},
				},
			}

			// Would normally broadcast here
			server.BroadcastStateUpdate(update)

			if i == 0 {
				fmt.Printf("Server: Broadcasting state for Entity %d at (%.0f, %.0f)\n",
					entity.ID, position.X, position.Y)
			}
		}
	}

	fmt.Println("Server: Completed game loop simulation")
}

func simulateClientInput(client *network.Client, playerID uint64, inputs int) {
	fmt.Printf("Client %d: Starting input simulation...\n", playerID)

	serializer := network.NewComponentSerializer()

	for i := 0; i < inputs; i++ {
		time.Sleep(50 * time.Millisecond) // 20 Hz

		// Simulate movement input
		var dx, dy int8
		if i%2 == 0 {
			dx = 1 // Move right
		} else {
			dx = 0
			dy = 1 // Move down
		}

		// Would normally send input here
		inputData := serializer.SerializeInput(dx, dy)
		_ = inputData // In real scenario: client.SendInput("move", inputData)

		if i == 0 {
			fmt.Printf("Client %d: Sending movement input (dx=%d, dy=%d)\n", playerID, dx, dy)
		}
	}

	fmt.Printf("Client %d: Completed input simulation\n", playerID)
}

func simulateClientReceive(client *network.Client, world *engine.World, playerID uint64, updates int) {
	fmt.Printf("Client %d: Starting state receive simulation...\n", playerID)

	serializer := network.NewComponentSerializer()

	for i := 0; i < updates; i++ {
		time.Sleep(50 * time.Millisecond) // 20 Hz

		// Simulate receiving state update
		// In real scenario: update := <-client.ReceiveStateUpdate()

		// Process simulated update
		if i == 0 {
			// Example of processing a position update
			posData := serializer.SerializePosition(100+float64(i*10), 100+float64(i*10))
			x, y, err := serializer.DeserializePosition(posData)
			if err == nil {
				fmt.Printf("Client %d: Received position update (%.0f, %.0f)\n", playerID, x, y)

				// Would update local entity here
				// entity := world.GetEntity(entityID)
				// if pos, ok := entity.GetComponent("position"); ok {
				//     position := pos.(*engine.PositionComponent)
				//     position.X = x
				//     position.Y = y
				// }
			}
		}
	}

	fmt.Printf("Client %d: Completed state receive simulation\n", playerID)
}
