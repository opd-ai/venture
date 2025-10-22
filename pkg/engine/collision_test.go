package engine

import (
	"testing"
)

func TestCollisionSystemCreation(t *testing.T) {
	system := NewCollisionSystem(32.0)

	if system == nil {
		t.Fatal("NewCollisionSystem returned nil")
	}

	if system.CellSize != 32.0 {
		t.Errorf("CellSize = %f, want 32.0", system.CellSize)
	}
}

func TestCollisionSystemBasicCollision(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create two overlapping entities
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 5, Y: 5})
	e2.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})

	world.Update(0) // Process additions

	// Track collisions
	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	if collisionCount == 0 {
		t.Error("Expected collision to be detected")
	}
}

func TestCollisionSystemNoCollision(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create two non-overlapping entities
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 50, Y: 50})
	e2.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})

	world.Update(0)

	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	if collisionCount != 0 {
		t.Error("Did not expect collision to be detected")
	}
}

func TestCollisionSystemResolution(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create two overlapping solid entities
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})
	e1.AddComponent(&VelocityComponent{VX: 5, VY: 0})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 5, Y: 0})
	e2.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})
	e2.AddComponent(&VelocityComponent{VX: -5, VY: 0})

	world.Update(0)

	// Get initial positions
	pos1Before, _ := e1.GetComponent("position")
	pos2Before, _ := e2.GetComponent("position")
	x1Before := pos1Before.(*PositionComponent).X
	x2Before := pos2Before.(*PositionComponent).X

	system.Update(world.GetEntities(), 0.016)

	// Positions should be adjusted to separate entities
	pos1After, _ := e1.GetComponent("position")
	pos2After, _ := e2.GetComponent("position")
	x1After := pos1After.(*PositionComponent).X
	x2After := pos2After.(*PositionComponent).X

	// Entities should be pushed apart
	if x1After >= x1Before {
		t.Errorf("e1 should have been pushed left, x went from %f to %f", x1Before, x1After)
	}
	if x2After <= x2Before {
		t.Errorf("e2 should have been pushed right, x went from %f to %f", x2Before, x2After)
	}

	// Velocities should be zeroed
	vel1, _ := e1.GetComponent("velocity")
	vel2, _ := e2.GetComponent("velocity")

	if vel1.(*VelocityComponent).VX != 0 {
		t.Error("e1 velocity should be zeroed")
	}
	if vel2.(*VelocityComponent).VX != 0 {
		t.Error("e2 velocity should be zeroed")
	}
}

func TestCollisionSystemTrigger(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create entity and trigger
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})

	trigger := world.CreateEntity()
	trigger.AddComponent(&PositionComponent{X: 5, Y: 5})
	trigger.AddComponent(&ColliderComponent{Width: 10, Height: 10, IsTrigger: true})

	world.Update(0)

	// Get initial position
	pos1Before, _ := e1.GetComponent("position")
	x1Before := pos1Before.(*PositionComponent).X
	y1Before := pos1Before.(*PositionComponent).Y

	collisionDetected := false
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionDetected = true
	})

	system.Update(world.GetEntities(), 0.016)

	// Collision should be detected
	if !collisionDetected {
		t.Error("Trigger collision should be detected")
	}

	// But entities should not be separated (trigger doesn't block)
	pos1After, _ := e1.GetComponent("position")
	x1After := pos1After.(*PositionComponent).X
	y1After := pos1After.(*PositionComponent).Y

	if x1After != x1Before || y1After != y1Before {
		t.Error("Entity should not be moved by trigger")
	}
}

func TestCollisionSystemLayers(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create entities on different layers
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true, Layer: 1})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 5, Y: 5})
	e2.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true, Layer: 2})

	world.Update(0)

	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	// Collision should not be detected (different layers)
	if collisionCount != 0 {
		t.Error("Collision should not be detected between different layers")
	}

	// Test with layer 0 (collides with all)
	e3 := world.CreateEntity()
	e3.AddComponent(&PositionComponent{X: 5, Y: 5})
	e3.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true, Layer: 0})

	world.Update(0)

	system.Update(world.GetEntities(), 0.016)

	// Now should detect collisions (layer 0 collides with all)
	if collisionCount == 0 {
		t.Error("Layer 0 should collide with all layers")
	}
}

func TestCollisionSystemMultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create a grid of entities with some overlap
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: float64(x * 8), Y: float64(y * 8)})
			entity.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})
		}
	}

	world.Update(0)

	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	// Update should handle many entities efficiently
	system.Update(world.GetEntities(), 0.016)

	// Entities are overlapping (spaced 8 units apart with 10 unit width)
	if collisionCount == 0 {
		t.Error("Expected some collisions in entity grid")
	}
}

func TestCheckCollision(t *testing.T) {
	e1 := NewEntity(1)
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&ColliderComponent{Width: 10, Height: 10})

	e2 := NewEntity(2)
	e2.AddComponent(&PositionComponent{X: 5, Y: 5})
	e2.AddComponent(&ColliderComponent{Width: 10, Height: 10})

	// Should be colliding
	if !CheckCollision(e1, e2) {
		t.Error("Expected collision")
	}

	// Move e2 away
	SetPosition(e2, 50, 50)

	// Should not be colliding
	if CheckCollision(e1, e2) {
		t.Error("Did not expect collision")
	}
}

func TestCollisionSystemWithOffset(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create entity with centered collider (using offset)
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 10, Y: 10})
	e1.AddComponent(&ColliderComponent{
		Width:   10,
		Height:  10,
		OffsetX: -5, // Center the collider
		OffsetY: -5,
		Solid:   true,
	})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 10, Y: 10})
	e2.AddComponent(&ColliderComponent{Width: 10, Height: 10, Solid: true})

	world.Update(0)

	collisionDetected := false
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionDetected = true
	})

	system.Update(world.GetEntities(), 0.016)

	// Should detect collision with offset collider
	if !collisionDetected {
		t.Error("Expected collision with offset collider")
	}
}

func BenchmarkCollisionSystem(b *testing.B) {
	world := NewWorld()
	system := NewCollisionSystem(64.0)

	// Create 100 entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10 * 20),
			Y: float64(i / 10 * 20),
		})
		entity.AddComponent(&ColliderComponent{
			Width:  10,
			Height: 10,
			Solid:  true,
		})
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkMovementSystem(b *testing.B) {
	world := NewWorld()
	system := NewMovementSystem(0)

	// Create 100 moving entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: 0, Y: 0})
		entity.AddComponent(&VelocityComponent{VX: 10, VY: 5})
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
