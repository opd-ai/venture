// Package engine provides fluid physics integration tests.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
)

func TestNewFluidPhysicsSystem(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	if system == nil {
		t.Fatal("Expected non-nil system")
	}
	if system.world != world {
		t.Error("Expected world to be set")
	}
	if system.simulator == nil {
		t.Error("Expected simulator to be set")
	}
}

func TestFluidPhysicsSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	// Create entity with buoyancy component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&fluids.BuoyancyComponent{
		Mass:   100.0,
		Volume: 0.2,
	})

	// Add fluid at entity position
	system.AddFluid(50, 50, 0.8, fluids.FluidWater)

	// Update should not panic
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Check that buoyancy was calculated
	buoyancyComp, _ := entity.GetComponent("buoyancy")
	buoyancy := buoyancyComp.(*fluids.BuoyancyComponent)
	if buoyancy.Submerged == 0 {
		t.Error("Expected entity to be submerged")
	}
}

func TestFluidPhysicsSystem_Swimming(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	// Create entity with swimming component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 0})
	entity.AddComponent(&fluids.SwimmingComponent{
		Stamina:        100.0,
		MaxStamina:     100.0,
		StaminaDrain:   10.0,
		StaminaRegen:   5.0,
		SwimSpeed:      0.5,
		DrowningDamage: 10.0,
	})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Add water at entity position
	system.AddFluid(50, 50, 0.8, fluids.FluidWater)

	// Update should trigger swimming
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Check that swimming state was updated
	swimmingComp, _ := entity.GetComponent("swimming")
	swimming := swimmingComp.(*fluids.SwimmingComponent)
	if !swimming.IsSwimming {
		t.Error("Expected entity to be swimming")
	}
}

func TestFluidPhysicsSystem_Drowning(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	// Create entity with exhausted swimming component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&fluids.SwimmingComponent{
		IsSwimming:     true,
		Stamina:        0, // Exhausted
		MaxStamina:     100.0,
		StaminaDrain:   10.0,
		SwimSpeed:      0.5,
		DrowningDamage: 10.0,
	})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Add water at entity position
	system.AddFluid(50, 50, 0.8, fluids.FluidWater)

	// Update should trigger drowning damage
	entities := []*Entity{entity}
	system.Update(entities, 1.0) // 1 second

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	// Should have taken some drowning damage
	if health.Current >= 100 {
		t.Error("Expected health to decrease from drowning")
	}
}

func TestFluidPhysicsSystem_Flooding(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	// Create entity with flooding component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&fluids.FloodingComponent{
		AreaID:        "test-area",
		FloodLevel:    0.0,
		FloodRate:     0.1,
		MaxFloodLevel: 1.0,
		Sources: []fluids.FloodSource{
			{X: 50, Y: 50, FlowRate: 0.1},
		},
	})

	// Update should increase flood level
	entities := []*Entity{entity}
	floodComp1, _ := entity.GetComponent("flooding")
	initialLevel := floodComp1.(*fluids.FloodingComponent).FloodLevel
	system.Update(entities, 1.0) // 1 second

	floodComp2, _ := entity.GetComponent("flooding")
	flooding := floodComp2.(*fluids.FloodingComponent)
	if flooding.FloodLevel <= initialLevel {
		t.Error("Expected flood level to increase")
	}
}

func TestFluidPhysicsSystem_AddRemoveFluid(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	// Add fluid
	system.AddFluid(50, 50, 0.5, fluids.FluidWater)

	// Check fluid was added
	amount, fluidType := system.getFluidAtPosition(50, 50)
	if amount == 0 {
		t.Error("Expected fluid to be present")
	}
	if fluidType != fluids.FluidWater {
		t.Error("Expected water fluid type")
	}

	// Remove fluid
	system.RemoveFluid(50, 50, 0.5)

	// Check fluid was removed
	amount, _ = system.getFluidAtPosition(50, 50)
	if amount > 0.01 { // Allow small epsilon
		t.Errorf("Expected fluid to be removed, got %f", amount)
	}
}

func TestFluidPhysicsSystem_GetSimulator(t *testing.T) {
	world := NewWorld()
	system := NewFluidPhysicsSystemWithDefaults(world)

	simulator := system.GetSimulator()
	if simulator == nil {
		t.Error("Expected non-nil simulator")
	}
}
