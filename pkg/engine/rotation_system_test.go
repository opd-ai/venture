// Package engine provides tests for rotation system.
package engine

import (
	"math"
	"testing"
)

// TestNewRotationSystem tests system creation
func TestNewRotationSystem(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	if system == nil {
		t.Fatal("NewRotationSystem() returned nil")
	}
	if system.world != world {
		t.Error("System world reference not set correctly")
	}
}

// TestRotationSystem_Update tests basic rotation updates
func TestRotationSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	// Create entity with rotation component
	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	rotation.SetTargetAngle(math.Pi / 2)
	entity.AddComponent(rotation)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	// Update should rotate towards target
	system.Update(0.1) // 0.3 radians max rotation

	entities := world.GetEntities()
	if len(entities) == 0 {
		t.Fatal("Entity not found in world")
	}

	rotComp, ok := entities[0].GetComponent("rotation")
	if !ok {
		t.Fatal("Rotation component not found")
	}
	rot := rotComp.(*RotationComponent)

	// Should have rotated towards π/2 but not reached it yet
	if !floatEqual(rot.Angle, 0.3, 0.01) {
		t.Errorf("Angle = %v, want ~0.3", rot.Angle)
	}
}

// TestRotationSystem_UpdateWithAim tests rotation syncing with aim
func TestRotationSystem_UpdateWithAim(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	// Create entity with rotation, aim, and position components
	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	aim := NewAimComponent(0)
	position := &PositionComponent{X: 100, Y: 100}

	// Set aim target
	aim.SetAimTarget(200, 100) // Aim right (0 radians)

	entity.AddComponent(rotation)
	entity.AddComponent(aim)
	entity.AddComponent(position)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	// Update should sync rotation with aim
	system.Update(0.1)

	entities := world.GetEntities()
	if len(entities) == 0 {
		t.Fatal("Entity not found in world")
	}

	rotComp, ok := entities[0].GetComponent("rotation")
	if !ok {
		t.Fatal("Rotation component not found")
	}
	rot := rotComp.(*RotationComponent)

	// Rotation target should match aim angle (0 radians = right)
	if !floatEqual(rot.TargetAngle, 0, 0.01) {
		t.Errorf("TargetAngle = %v, want 0", rot.TargetAngle)
	}
}

// TestRotationSystem_UpdateMultipleEntities tests batch processing
func TestRotationSystem_UpdateMultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	// Create multiple entities with rotation
	for i := 1; i <= 3; i++ {
		entity := NewEntity(uint64(i))
		rotation := NewRotationComponent(0, 3.0)
		rotation.SetTargetAngle(float64(i) * math.Pi / 4)
		entity.AddComponent(rotation)
		world.AddEntity(entity)
	}
	world.Update(0.016) // Process pending additions

	// Update all entities
	system.Update(0.05)

	entities := world.GetEntities()
	if len(entities) != 3 {
		t.Fatalf("Expected 3 entities, got %d", len(entities))
	}

	// All entities should have updated rotation
	for _, entity := range entities {
		rotComp, ok := entity.GetComponent("rotation")
		if !ok {
			t.Errorf("Entity %d missing rotation component", entity.ID)
			continue
		}
		rot := rotComp.(*RotationComponent)

		// Each entity should have moved towards its target
		if rot.Angle == 0 {
			t.Errorf("Entity %d rotation not updated", entity.ID)
		}
	}
}

// TestRotationSystem_SyncRotationToAim tests immediate aim sync
func TestRotationSystem_SyncRotationToAim(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	aim := NewAimComponent(math.Pi / 2)

	entity.AddComponent(rotation)
	entity.AddComponent(aim)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	ok := system.SyncRotationToAim(1)
	if !ok {
		t.Fatal("SyncRotationToAim() failed")
	}

	entities := world.GetEntities()
	rotComp, _ := entities[0].GetComponent("rotation")
	rot := rotComp.(*RotationComponent)

	if !floatEqual(rot.Angle, math.Pi/2, 0.0001) {
		t.Errorf("Angle = %v, want %v", rot.Angle, math.Pi/2)
	}
}

// TestRotationSystem_SyncRotationToAimErrors tests error conditions
func TestRotationSystem_SyncRotationToAimErrors(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	// Entity doesn't exist
	ok := system.SyncRotationToAim(999)
	if ok {
		t.Error("Expected false for non-existent entity")
	}

	// Entity missing components
	entity := NewEntity(1)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	ok = system.SyncRotationToAim(1)
	if ok {
		t.Error("Expected false for entity missing components")
	}
}

// TestRotationSystem_SetEntityRotation tests direct rotation setting
func TestRotationSystem_SetEntityRotation(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	entity.AddComponent(rotation)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	ok := system.SetEntityRotation(1, math.Pi)
	if !ok {
		t.Fatal("SetEntityRotation() failed")
	}

	angle, ok := system.GetEntityRotation(1)
	if !ok {
		t.Fatal("GetEntityRotation() failed")
	}

	if !floatEqual(angle, math.Pi, 0.0001) {
		t.Errorf("Angle = %v, want %v", angle, math.Pi)
	}
}

// TestRotationSystem_GetEntityRotationErrors tests query error conditions
func TestRotationSystem_GetEntityRotationErrors(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	// Entity doesn't exist
	_, ok := system.GetEntityRotation(999)
	if ok {
		t.Error("Expected false for non-existent entity")
	}

	// Entity missing component
	entity := NewEntity(1)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	_, ok = system.GetEntityRotation(1)
	if ok {
		t.Error("Expected false for entity missing component")
	}
}

// TestRotationSystem_EnableSmoothRotation tests rotation mode switching
func TestRotationSystem_EnableSmoothRotation(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	entity.AddComponent(rotation)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	// Disable smooth rotation
	ok := system.EnableSmoothRotation(1, false)
	if !ok {
		t.Fatal("EnableSmoothRotation() failed")
	}

	entities := world.GetEntities()
	rotComp, _ := entities[0].GetComponent("rotation")
	rot := rotComp.(*RotationComponent)

	if rot.SmoothRotation {
		t.Error("SmoothRotation should be false")
	}

	// Enable smooth rotation
	ok = system.EnableSmoothRotation(1, true)
	if !ok {
		t.Fatal("EnableSmoothRotation() failed")
	}

	if !rot.SmoothRotation {
		t.Error("SmoothRotation should be true")
	}
}

// TestRotationSystem_SetRotationSpeed tests speed configuration
func TestRotationSystem_SetRotationSpeed(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	entity.AddComponent(rotation)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	ok := system.SetRotationSpeed(1, 5.0)
	if !ok {
		t.Fatal("SetRotationSpeed() failed")
	}

	entities := world.GetEntities()
	rotComp, _ := entities[0].GetComponent("rotation")
	rot := rotComp.(*RotationComponent)

	if rot.RotationSpeed != 5.0 {
		t.Errorf("RotationSpeed = %v, want 5.0", rot.RotationSpeed)
	}
}

// TestRotationSystem_UpdateWithoutAim tests rotation without aim sync
func TestRotationSystem_UpdateWithoutAim(t *testing.T) {
	world := NewWorld()
	system := NewRotationSystem(world)

	// Entity with only rotation (no aim component)
	entity := NewEntity(1)
	rotation := NewRotationComponent(0, 3.0)
	rotation.SetTargetAngle(math.Pi / 2)
	entity.AddComponent(rotation)
	world.AddEntity(entity)
	world.Update(0.016) // Process pending additions

	// Should still update rotation normally
	system.Update(0.1)

	entities := world.GetEntities()
	rotComp, _ := entities[0].GetComponent("rotation")
	rot := rotComp.(*RotationComponent)

	// Should have rotated towards target
	if rot.Angle == 0 {
		t.Error("Rotation should have updated without aim component")
	}
}
