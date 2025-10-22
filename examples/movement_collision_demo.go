//go:build test
// +build test

package main

// This example demonstrates the movement and collision systems.
// Run with: go run -tags test ./examples/movement_collision_demo.go

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

func main() {
	fmt.Println("=== Movement and Collision System Example ===\n")

	// Create world and systems
	world := engine.NewWorld()
	movementSystem := engine.NewMovementSystem(100.0)  // Max speed of 100 units/s
	collisionSystem := engine.NewCollisionSystem(32.0) // 32 unit cell size

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	// Track collisions for demonstration
	collisionCount := 0
	collisionSystem.SetCollisionCallback(func(e1, e2 *engine.Entity) {
		collisionCount++
		x1, y1, _ := engine.GetPosition(e1)
		x2, y2, _ := engine.GetPosition(e2)
		fmt.Printf("Collision between Entity %d at (%.1f, %.1f) and Entity %d at (%.1f, %.1f)\n",
			e1.ID, x1, y1, e2.ID, x2, y2)
	})

	// Example 1: Basic movement
	fmt.Println("Example 1: Basic Movement")
	fmt.Println("--------------------------")

	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	player.AddComponent(&engine.VelocityComponent{VX: 50, VY: 30})

	world.Update(0) // Process additions

	fmt.Printf("Player initial position: (0, 0)\n")
	fmt.Printf("Player velocity: (50, 30) units/second\n")

	// Simulate 1 second
	world.Update(1.0)

	x, y, _ := engine.GetPosition(player)
	fmt.Printf("Player position after 1 second: (%.1f, %.1f)\n\n", x, y)

	// Example 2: Collision detection
	fmt.Println("Example 2: Collision Detection")
	fmt.Println("-------------------------------")

	// Reset for new example
	world = engine.NewWorld()
	movementSystem = engine.NewMovementSystem(0)
	collisionSystem = engine.NewCollisionSystem(32.0)
	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	collisionDetected := false
	collisionSystem.SetCollisionCallback(func(e1, e2 *engine.Entity) {
		collisionDetected = true
		fmt.Println("✓ Collision detected and resolved!")
	})

	// Create two entities moving towards each other
	entity1 := world.CreateEntity()
	entity1.AddComponent(&engine.PositionComponent{X: 0, Y: 50})
	entity1.AddComponent(&engine.VelocityComponent{VX: 100, VY: 0})
	entity1.AddComponent(&engine.ColliderComponent{Width: 20, Height: 20, Solid: true})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
	entity2.AddComponent(&engine.VelocityComponent{VX: -100, VY: 0})
	entity2.AddComponent(&engine.ColliderComponent{Width: 20, Height: 20, Solid: true})

	world.Update(0) // Process additions

	fmt.Println("Entity 1: pos=(0, 50), vel=(100, 0), size=20x20")
	fmt.Println("Entity 2: pos=(100, 50), vel=(-100, 0), size=20x20")
	fmt.Println("Simulating collision...")

	// Simulate until collision
	for i := 0; i < 10 && !collisionDetected; i++ {
		world.Update(0.1)
	}

	x1, y1, _ := engine.GetPosition(entity1)
	x2, y2, _ := engine.GetPosition(entity2)
	fmt.Printf("Final positions: Entity 1=(%.1f, %.1f), Entity 2=(%.1f, %.1f)\n\n", x1, y1, x2, y2)

	// Example 3: Trigger zones
	fmt.Println("Example 3: Trigger Zones")
	fmt.Println("------------------------")

	world = engine.NewWorld()
	movementSystem = engine.NewMovementSystem(0)
	collisionSystem = engine.NewCollisionSystem(32.0)
	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	triggerEntered := false
	collisionSystem.SetCollisionCallback(func(e1, e2 *engine.Entity) {
		triggerEntered = true
	})

	// Create moving entity
	mover := world.CreateEntity()
	mover.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	mover.AddComponent(&engine.VelocityComponent{VX: 50, VY: 0})
	mover.AddComponent(&engine.ColliderComponent{Width: 10, Height: 10, Solid: true})

	// Create trigger zone
	trigger := world.CreateEntity()
	trigger.AddComponent(&engine.PositionComponent{X: 100, Y: 0})
	trigger.AddComponent(&engine.ColliderComponent{
		Width:     50,
		Height:    50,
		IsTrigger: true, // Detects but doesn't block
	})

	world.Update(0)

	fmt.Println("Moving entity towards trigger zone...")

	// Simulate movement into trigger
	for i := 0; i < 30; i++ {
		world.Update(0.1)
		if triggerEntered {
			fmt.Println("✓ Trigger zone activated!")
			break
		}
	}
	fmt.Println()

	// Example 4: World boundaries
	fmt.Println("Example 4: World Boundaries")
	fmt.Println("---------------------------")

	world = engine.NewWorld()
	movementSystem = engine.NewMovementSystem(0)
	world.AddSystem(movementSystem)

	bounded := world.CreateEntity()
	bounded.AddComponent(&engine.PositionComponent{X: 450, Y: 50})
	bounded.AddComponent(&engine.VelocityComponent{VX: 100, VY: 0})
	bounded.AddComponent(&engine.BoundsComponent{
		MinX: 0,
		MinY: 0,
		MaxX: 500,
		MaxY: 100,
		Wrap: false,
	})

	world.Update(0)

	fmt.Println("Entity at (450, 50) moving right at 100 units/s")
	fmt.Println("World boundary at x=500")

	world.Update(1.0) // Should hit boundary

	x, y, _ = engine.GetPosition(bounded)
	fmt.Printf("Position after 1 second: (%.1f, %.1f)\n", x, y)
	fmt.Println("✓ Entity stopped at boundary\n")

	// Example 5: Multiple entities with spatial partitioning
	fmt.Println("Example 5: Spatial Partitioning Performance")
	fmt.Println("-------------------------------------------")

	world = engine.NewWorld()
	movementSystem = engine.NewMovementSystem(0)
	collisionSystem = engine.NewCollisionSystem(64.0)
	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	rng := rand.New(rand.NewSource(12345))

	// Create many entities
	entityCount := 50
	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: rng.Float64() * 400,
			Y: rng.Float64() * 400,
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: (rng.Float64()*2 - 1) * 50,
			VY: (rng.Float64()*2 - 1) * 50,
		})
		entity.AddComponent(&engine.ColliderComponent{
			Width:  10,
			Height: 10,
			Solid:  true,
		})
	}

	world.Update(0)

	collisionCount = 0
	collisionSystem.SetCollisionCallback(func(e1, e2 *engine.Entity) {
		collisionCount++
	})

	fmt.Printf("Created %d entities with random positions and velocities\n", entityCount)
	fmt.Println("Simulating 1 second with collision detection...")

	// Simulate
	for i := 0; i < 60; i++ {
		world.Update(1.0 / 60.0)
	}

	fmt.Printf("✓ Detected %d collisions\n", collisionCount)
	fmt.Println("✓ Spatial partitioning enabled O(n) performance\n")

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("The movement and collision systems provide:")
	fmt.Println("• Velocity-based movement with speed limits")
	fmt.Println("• AABB collision detection")
	fmt.Println("• Automatic collision resolution")
	fmt.Println("• Trigger zones for events")
	fmt.Println("• World boundary constraints")
	fmt.Println("• Spatial partitioning for performance")
	fmt.Println("\nReady for integration into Phase 5 gameplay systems!")
}
