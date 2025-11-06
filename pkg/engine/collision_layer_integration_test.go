// Package engine provides tests for collision system integration with LayerComponent.
// Tests verify that entities on different terrain layers don't collide unless one can fly.
package engine

import (
	"testing"
)

// TestCollisionSystem_LayerComponent_DifferentLayers verifies entities on different layers don't collide.
func TestCollisionSystem_LayerComponent_DifferentLayers(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create entity on ground layer (Layer 0)
	groundEntity := world.CreateEntity()
	groundEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	groundEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	groundLayer := NewLayerComponent()
	groundLayer.CurrentLayer = 0
	groundEntity.AddComponent(&groundLayer)

	// Create entity on platform layer (Layer 2) at same position
	platformEntity := world.CreateEntity()
	platformEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	platformEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	platformLayer := NewLayerComponent()
	platformLayer.CurrentLayer = 2
	platformEntity.AddComponent(&platformLayer)

	world.Update(0)

	// Track collisions
	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	// Entities on different terrain layers should NOT collide
	if collisionCount != 0 {
		t.Errorf("Expected no collision between entities on different layers, but got %d collision(s)", collisionCount)
	}
}

// TestCollisionSystem_LayerComponent_SameLayer verifies entities on same layer do collide.
func TestCollisionSystem_LayerComponent_SameLayer(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create two entities on ground layer (Layer 0)
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	layer1 := NewLayerComponent()
	layer1.CurrentLayer = 0
	entity1.AddComponent(&layer1)

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 105, Y: 105})
	entity2.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	layer2 := NewLayerComponent()
	layer2.CurrentLayer = 0
	entity2.AddComponent(&layer2)

	world.Update(0)

	// Track collisions
	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	// Entities on same terrain layer should collide
	if collisionCount == 0 {
		t.Error("Expected collision between entities on same layer")
	}
}

// TestCollisionSystem_LayerComponent_FlyingEntity verifies flying entities collide with all layers.
func TestCollisionSystem_LayerComponent_FlyingEntity(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create ground entity
	groundEntity := world.CreateEntity()
	groundEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	groundEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	groundLayer := NewLayerComponent()
	groundLayer.CurrentLayer = 0
	groundEntity.AddComponent(&groundLayer)

	// Create flying entity at same position
	flyingEntity := world.CreateEntity()
	flyingEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	flyingEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	flyingLayer := NewFlyingLayerComponent()
	flyingLayer.CurrentLayer = 2 // On platform layer but can fly
	flyingEntity.AddComponent(&flyingLayer)

	world.Update(0)

	// Track collisions
	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	// Flying entities should collide with entities on other layers
	if collisionCount == 0 {
		t.Error("Expected collision between flying entity and ground entity")
	}
}

// TestCollisionSystem_WouldCollideWithEntity_DifferentLayers tests predictive collision with layers.
func TestCollisionSystem_WouldCollideWithEntity_DifferentLayers(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create ground entity at (100, 100)
	groundEntity := world.CreateEntity()
	groundEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	groundEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	groundLayer := NewLayerComponent()
	groundLayer.CurrentLayer = 0
	groundEntity.AddComponent(&groundLayer)

	// Create platform entity at same position
	platformEntity := world.CreateEntity()
	platformEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	platformEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	platformLayer := NewLayerComponent()
	platformLayer.CurrentLayer = 2
	platformEntity.AddComponent(&platformLayer)

	world.Update(0)

	// Test predictive collision
	wouldCollide := system.WouldCollideWithEntity(groundEntity, 100, 100, platformEntity)

	// Should NOT predict collision between different layers
	if wouldCollide {
		t.Error("WouldCollideWithEntity should return false for entities on different layers")
	}
}

// TestCollisionSystem_WouldCollideWithEntity_SameLayer tests predictive collision on same layer.
func TestCollisionSystem_WouldCollideWithEntity_SameLayer(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create two ground entities
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity1.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	layer1 := NewLayerComponent()
	layer1.CurrentLayer = 0
	entity1.AddComponent(&layer1)

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity2.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})
	layer2 := NewLayerComponent()
	layer2.CurrentLayer = 0
	entity2.AddComponent(&layer2)

	world.Update(0)

	// Test predictive collision - move entity1 to entity2's position
	wouldCollide := system.WouldCollideWithEntity(entity1, 100, 100, entity2)

	// Should predict collision on same layer
	if !wouldCollide {
		t.Error("WouldCollideWithEntity should return true for entities on same layer")
	}
}

// TestCollisionSystem_NoLayerComponent verifies backward compatibility.
func TestCollisionSystem_NoLayerComponent(t *testing.T) {
	world := NewWorld()
	system := NewCollisionSystem(32.0)

	// Create two entities WITHOUT LayerComponent (should use old behavior)
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 105, Y: 105})
	entity2.AddComponent(&ColliderComponent{Width: 16, Height: 16, Solid: true})

	world.Update(0)

	// Track collisions
	collisionCount := 0
	system.SetCollisionCallback(func(e1, e2 *Entity) {
		collisionCount++
	})

	system.Update(world.GetEntities(), 0.016)

	// Entities without LayerComponent should still collide normally
	if collisionCount == 0 {
		t.Error("Expected collision between entities without LayerComponent (backward compatibility)")
	}
}
