// Package engine provides tests for VehicleCombatSystem.
// Phase 21.2: Advanced Vehicle Features - Vehicle Weapon Projectile Spawning
package engine

import (
	"math"
	"testing"
)

func TestVehicleCombatSystem_NewVehicleCombatSystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)

	if system == nil {
		t.Fatal("NewVehicleCombatSystem returned nil")
	}
	if system.world != world {
		t.Error("world reference not set correctly")
	}
}

func TestVehicleCombatSystem_NewVehicleCombatSystem_NilWorld(t *testing.T) {
	system := NewVehicleCombatSystem(nil)
	if system == nil {
		t.Fatal("NewVehicleCombatSystem returned nil for nil world")
	}
}

func TestVehicleCombatSystem_SetProjectileSystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)
	projSystem := NewProjectileSystem(world)

	system.SetProjectileSystem(projSystem)

	if system.projectileSystem != projSystem {
		t.Error("SetProjectileSystem did not set projectile system correctly")
	}
}

func TestVehicleCombatSystem_SetCombatSystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)
	combatSystem := &CombatSystem{}

	system.SetCombatSystem(combatSystem)

	if system.combatSystem != combatSystem {
		t.Error("SetCombatSystem did not set combat system correctly")
	}
}

func TestVehicleCombatSystem_SpawnWeaponProjectile_WithProjectileSystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)
	projSystem := NewProjectileSystem(world)
	system.SetProjectileSystem(projSystem)

	// Create vehicle entity
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 100, Y: 100}
	vehicle.AddComponent(pos)

	// Create combat component with weapon settings
	combat := NewVehicleCombatComponent()
	combat.WeaponMounted = true
	combat.WeaponDamage = 25.0
	combat.WeaponRange = 300.0
	combat.WeaponProjectileSpeed = 400.0
	combat.WeaponType = "Cannon"
	vehicle.AddComponent(combat)

	// Initial projectile count
	initialCount := projSystem.GetProjectileCount()

	// Spawn projectile at 0 degrees (right)
	angle := 0.0
	system.spawnWeaponProjectile(vehicle, pos, angle, combat)

	// Process entity creation
	world.Update(0)

	// Check projectile was created
	newCount := projSystem.GetProjectileCount()
	if newCount != initialCount+1 {
		t.Errorf("Expected %d projectiles, got %d", initialCount+1, newCount)
	}
}

func TestVehicleCombatSystem_SpawnWeaponProjectile_NoProjectileSystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)
	// Do not set projectile system

	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 100, Y: 100}
	vehicle.AddComponent(pos)

	combat := NewVehicleCombatComponent()
	combat.WeaponMounted = true
	vehicle.AddComponent(combat)

	// This should not panic, just log and return
	system.spawnWeaponProjectile(vehicle, pos, 0.0, combat)
}

func TestVehicleCombatSystem_SpawnWeaponProjectile_VelocityCalculation(t *testing.T) {
	tests := []struct {
		name           string
		angle          float64
		expectedVxSign float64 // 1 for positive, -1 for negative
		expectedVySign float64 // 1 for positive, -1 for negative
	}{
		{"right (0 degrees)", 0.0, 1.0, 0.0},
		{"down (90 degrees)", math.Pi / 2, 0.0, 1.0},
		{"left (180 degrees)", math.Pi, -1.0, 0.0},
		{"up (-90 degrees)", -math.Pi / 2, 0.0, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewVehicleCombatSystem(world)
			projSystem := NewProjectileSystem(world)
			system.SetProjectileSystem(projSystem)

			vehicle := world.CreateEntity()
			pos := &PositionComponent{X: 100, Y: 100}
			vehicle.AddComponent(pos)

			combat := NewVehicleCombatComponent()
			combat.WeaponMounted = true
			combat.WeaponProjectileSpeed = 300.0
			vehicle.AddComponent(combat)

			system.spawnWeaponProjectile(vehicle, pos, tt.angle, combat)

			// Process entity creation
			world.Update(0)

			// Find the projectile entity
			projectiles := world.GetEntitiesWith("projectile", "velocity")
			if len(projectiles) == 0 {
				t.Fatal("No projectile entity created")
			}

			proj := projectiles[0]
			vel := proj.GetVelocity()
			if vel == nil {
				t.Fatal("Projectile has no velocity component")
			}

			// Check velocity direction
			if tt.expectedVxSign > 0 && vel.VX <= 0 {
				t.Errorf("Expected positive VX, got %f", vel.VX)
			}
			if tt.expectedVxSign < 0 && vel.VX >= 0 {
				t.Errorf("Expected negative VX, got %f", vel.VX)
			}
			if tt.expectedVySign > 0 && vel.VY <= 0 {
				t.Errorf("Expected positive VY, got %f", vel.VY)
			}
			if tt.expectedVySign < 0 && vel.VY >= 0 {
				t.Errorf("Expected negative VY, got %f", vel.VY)
			}
		})
	}
}

func TestVehicleCombatSystem_WeaponTypeToProjectileType(t *testing.T) {
	system := &VehicleCombatSystem{}

	tests := []struct {
		weaponType       string
		expectedProjType string
	}{
		{"Cannon", "bullet"},
		{"MachineGun", "bullet"},
		{"Laser", "lightning_bolt"},
		{"Magic", "magic_missile"},
		{"Ballista", "arrow"},
		{"Unknown", "bullet"},
		{"", "bullet"},
	}

	for _, tt := range tests {
		t.Run(tt.weaponType, func(t *testing.T) {
			result := system.weaponTypeToProjectileType(tt.weaponType)
			if result != tt.expectedProjType {
				t.Errorf("weaponTypeToProjectileType(%q) = %q, want %q",
					tt.weaponType, result, tt.expectedProjType)
			}
		})
	}
}

func TestVehicleCombatSystem_SpawnWeaponProjectile_ProjectileProperties(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)
	projSystem := NewProjectileSystem(world)
	system.SetProjectileSystem(projSystem)

	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 150, Y: 200}
	vehicle.AddComponent(pos)

	combat := NewVehicleCombatComponent()
	combat.WeaponMounted = true
	combat.WeaponDamage = 50.0
	combat.WeaponRange = 400.0
	combat.WeaponProjectileSpeed = 500.0
	combat.WeaponType = "Magic"
	vehicle.AddComponent(combat)

	system.spawnWeaponProjectile(vehicle, pos, 0.0, combat)

	// Process entity creation
	world.Update(0)

	// Find the projectile entity
	projectiles := world.GetEntitiesWith("projectile")
	if len(projectiles) == 0 {
		t.Fatal("No projectile entity created")
	}

	proj := projectiles[0]
	projComp, ok := proj.GetComponent("projectile")
	if !ok {
		t.Fatal("Projectile entity has no projectile component")
	}

	pc := projComp.(*ProjectileComponent)

	// Check damage
	if pc.Damage != 50.0 {
		t.Errorf("Expected damage 50.0, got %f", pc.Damage)
	}

	// Check projectile type
	if pc.ProjectileType != "magic_missile" {
		t.Errorf("Expected projectile type 'magic_missile', got %q", pc.ProjectileType)
	}

	// Check owner ID
	if pc.OwnerID != vehicle.ID {
		t.Errorf("Expected owner ID %d, got %d", vehicle.ID, pc.OwnerID)
	}

	// Check lifetime is calculated from range and speed
	expectedLifetime := combat.WeaponRange / combat.WeaponProjectileSpeed
	if math.Abs(pc.LifeTime-expectedLifetime) > 0.001 {
		t.Errorf("Expected lifetime %f, got %f", expectedLifetime, pc.LifeTime)
	}
}

func TestVehicleCombatComponent_WeaponProjectileSpeed(t *testing.T) {
	comp := NewVehicleCombatComponent()

	// Check default value is set
	if comp.WeaponProjectileSpeed != 300.0 {
		t.Errorf("Expected default WeaponProjectileSpeed 300.0, got %f", comp.WeaponProjectileSpeed)
	}

	// Test setting custom value
	comp.WeaponProjectileSpeed = 500.0
	if comp.WeaponProjectileSpeed != 500.0 {
		t.Errorf("Expected WeaponProjectileSpeed 500.0, got %f", comp.WeaponProjectileSpeed)
	}
}

func TestVehicleCombatSystem_SpawnWeaponProjectile_DefaultSpeed(t *testing.T) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)
	projSystem := NewProjectileSystem(world)
	system.SetProjectileSystem(projSystem)

	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 100, Y: 100}
	vehicle.AddComponent(pos)

	combat := NewVehicleCombatComponent()
	combat.WeaponMounted = true
	combat.WeaponProjectileSpeed = 0 // Zero speed should use default
	combat.WeaponRange = 300.0
	vehicle.AddComponent(combat)

	// Should not panic and should use default speed
	system.spawnWeaponProjectile(vehicle, pos, 0.0, combat)

	// Process entity creation
	world.Update(0)

	// Check projectile was created
	projectiles := world.GetEntitiesWith("projectile", "velocity")
	if len(projectiles) == 0 {
		t.Fatal("No projectile entity created")
	}

	vel := projectiles[0].GetVelocity()
	if vel == nil {
		t.Fatal("Projectile has no velocity")
	}

	// Default speed is 300.0, so VX at angle 0 should be 300.0
	if math.Abs(vel.VX-300.0) > 0.001 {
		t.Errorf("Expected VX ~300.0 (default speed), got %f", vel.VX)
	}
}
