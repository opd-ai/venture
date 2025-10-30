// Package engine provides integration tests for projectile physics system.
// Phase 10.2: Projectile Physics System
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestProjectileSystemIntegration tests the full projectile workflow
func TestProjectileSystemIntegration(t *testing.T) {
	// Create world
	world := NewWorld()
	
	// Create projectile system
	ps := NewProjectileSystem(world)
	
	// Create attacker with ranged weapon
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	attacker.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	attacker.AddComponent(&RotationComponent{Angle: 0}) // Facing right
	attacker.AddComponent(&AimComponent{AimAngle: 0})
	
	// Create target
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 200, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	
	// Create combat system
	cs := NewCombatSystem(12345)
	cs.world = world
	cs.genreID = "fantasy"
	cs.projectileSystem = ps
	
	// Create ranged weapon (bow)
	weapon := &item.Item{
		Name:     "Test Bow",
		Type:     item.TypeWeapon,
		Rarity:   item.RarityCommon,
		WeaponType: item.WeaponBow,
		Stats: item.Stats{
			Attack:             10.0,
			IsProjectile:       true,
			ProjectileSpeed:    400.0,
			ProjectileLifetime: 3.0,
			ProjectileType:     "arrow",
			Pierce:             0,
			Bounce:             0,
			Explosive:          false,
		},
	}
	
	// Equip weapon
	equipment := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipment.Slots[SlotWeapon] = weapon
	attacker.AddComponent(equipment)
	
	// Add attack component
	attackComp := &AttackComponent{
		Damage:       10.0,
		Range:        500.0,
		AttackSpeed:  1.0,
		CurrentCooldown: 0.0,
	}
	attacker.AddComponent(attackComp)
	
	// Count entities before attack
	entitiesBefore := len(world.Entities)
	
	// Perform attack (should spawn projectile)
	success := cs.Attack(attacker, target)
	
	if !success {
		t.Error("expected attack to succeed and spawn projectile")
	}
	
	// Count entities after attack
	entitiesAfter := len(world.Entities)
	
	if entitiesAfter <= entitiesBefore {
		t.Errorf("expected new projectile entity, had %d entities, now have %d", entitiesBefore, entitiesAfter)
	}
	
	// Find projectile entity
	projectiles := world.GetEntitiesWith("projectile", "position", "velocity")
	if len(projectiles) != 1 {
		t.Errorf("expected 1 projectile, got %d", len(projectiles))
	}
	
	if len(projectiles) > 0 {
		proj := projectiles[0]
		
		// Check projectile has required components
		if !proj.HasComponent("projectile") {
			t.Error("projectile missing projectile component")
		}
		if !proj.HasComponent("position") {
			t.Error("projectile missing position component")
		}
		if !proj.HasComponent("velocity") {
			t.Error("projectile missing velocity component")
		}
		if !proj.HasComponent("rotation") {
			t.Error("projectile missing rotation component")
		}
		
		// Check projectile properties
		projComp := proj.GetComponent("projectile").(*ProjectileComponent)
		if projComp.Damage != 10.0 {
			t.Errorf("expected damage 10.0, got %f", projComp.Damage)
		}
		if projComp.Speed != 400.0 {
			t.Errorf("expected speed 400.0, got %f", projComp.Speed)
		}
		if projComp.LifeTime != 3.0 {
			t.Errorf("expected lifetime 3.0, got %f", projComp.LifeTime)
		}
		if projComp.ProjectileType != "arrow" {
			t.Errorf("expected type 'arrow', got '%s'", projComp.ProjectileType)
		}
		if projComp.OwnerID != attacker.ID {
			t.Errorf("expected ownerID %d, got %d", attacker.ID, projComp.OwnerID)
		}
		
		// Check velocity is set correctly (moving right)
		velComp := proj.GetComponent("velocity").(*VelocityComponent)
		if velComp.VX <= 0 {
			t.Errorf("expected positive VX (moving right), got %f", velComp.VX)
		}
	}
}

// TestProjectileSpawnWithPiercing tests projectile with pierce ability
func TestProjectileSpawnWithPiercing(t *testing.T) {
	world := NewWorld()
	ps := NewProjectileSystem(world)
	cs := NewCombatSystem(12345)
	cs.world = world
	cs.genreID = "fantasy"
	cs.projectileSystem = ps
	
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	attacker.AddComponent(&AimComponent{AimAngle: 0})
	attacker.AddComponent(&AttackComponent{Damage: 15.0, Range: 500.0, AttackSpeed: 1.0})
	
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 200, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	
	// Piercing weapon
	weapon := &item.Item{
		Name:     "Piercing Arrow",
		Type:     item.TypeWeapon,
		WeaponType: item.WeaponBow,
		Stats: item.Stats{
			Attack:             15.0,
			IsProjectile:       true,
			ProjectileSpeed:    400.0,
			ProjectileLifetime: 3.0,
			ProjectileType:     "arrow",
			Pierce:             2, // Can pierce 2 enemies
			Bounce:             0,
			Explosive:          false,
		},
	}
	
	equipment := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipment.Slots[SlotWeapon] = weapon
	attacker.AddComponent(equipment)
	
	cs.Attack(attacker, target)
	
	projectiles := world.GetEntitiesWith("projectile")
	if len(projectiles) != 1 {
		t.Fatalf("expected 1 projectile, got %d", len(projectiles))
	}
	
	projComp := projectiles[0].GetComponent("projectile").(*ProjectileComponent)
	if projComp.Pierce != 2 {
		t.Errorf("expected pierce=2, got %d", projComp.Pierce)
	}
	if !projComp.CanPierce() {
		t.Error("projectile should be able to pierce")
	}
}

// TestProjectileSpawnWithExplosive tests explosive projectile
func TestProjectileSpawnWithExplosive(t *testing.T) {
	world := NewWorld()
	ps := NewProjectileSystem(world)
	cs := NewCombatSystem(12345)
	cs.world = world
	cs.genreID = "fantasy"
	cs.projectileSystem = ps
	
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	attacker.AddComponent(&AimComponent{AimAngle: 0})
	attacker.AddComponent(&AttackComponent{Damage: 20.0, Range: 500.0, AttackSpeed: 1.0})
	
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 200, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	
	// Explosive weapon
	weapon := &item.Item{
		Name:     "Explosive Arrow",
		Type:     item.TypeWeapon,
		WeaponType: item.WeaponBow,
		Stats: item.Stats{
			Attack:             20.0,
			IsProjectile:       true,
			ProjectileSpeed:    400.0,
			ProjectileLifetime: 3.0,
			ProjectileType:     "fireball",
			Pierce:             0,
			Bounce:             0,
			Explosive:          true,
			ExplosionRadius:    50.0,
		},
	}
	
	equipment := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipment.Slots[SlotWeapon] = weapon
	attacker.AddComponent(equipment)
	
	cs.Attack(attacker, target)
	
	projectiles := world.GetEntitiesWith("projectile")
	if len(projectiles) != 1 {
		t.Fatalf("expected 1 projectile, got %d", len(projectiles))
	}
	
	projComp := projectiles[0].GetComponent("projectile").(*ProjectileComponent)
	if !projComp.Explosive {
		t.Error("expected explosive projectile")
	}
	if projComp.ExplosionRadius != 50.0 {
		t.Errorf("expected explosion radius 50.0, got %f", projComp.ExplosionRadius)
	}
}

// TestMeleeWeaponDoesNotSpawnProjectile tests that melee weapons don't spawn projectiles
func TestMeleeWeaponDoesNotSpawnProjectile(t *testing.T) {
	world := NewWorld()
	ps := NewProjectileSystem(world)
	cs := NewCombatSystem(12345)
	cs.world = world
	cs.genreID = "fantasy"
	cs.projectileSystem = ps
	
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	attacker.AddComponent(&AimComponent{AimAngle: 0})
	attacker.AddComponent(&AttackComponent{Damage: 10.0, Range: 50.0, AttackSpeed: 1.0})
	
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 120, Y: 100}) // Within melee range
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	
	// Melee weapon (NOT projectile)
	weapon := &item.Item{
		Name:     "Sword",
		Type:     item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Stats: item.Stats{
			Attack:       15.0,
			IsProjectile: false, // Not a projectile weapon
		},
	}
	
	equipment := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipment.Slots[SlotWeapon] = weapon
	attacker.AddComponent(equipment)
	
	entitiesBefore := len(world.Entities)
	cs.Attack(attacker, target)
	entitiesAfter := len(world.Entities)
	
	// No new entities should be created (no projectile spawned)
	if entitiesAfter != entitiesBefore {
		t.Errorf("melee weapon should not spawn projectile, had %d entities, now have %d", entitiesBefore, entitiesAfter)
	}
	
	// Target should take damage directly (melee hit)
	healthComp := target.GetComponent("health").(*HealthComponent)
	if healthComp.Current >= 100 {
		t.Error("target should have taken damage from melee attack")
	}
}

// TestProjectileSystemUpdate tests projectile movement and aging
func TestProjectileSystemUpdate(t *testing.T) {
	world := NewWorld()
	ps := NewProjectileSystem(world)
	
	// Create projectile entity
	proj := world.CreateEntity()
	proj.AddComponent(&PositionComponent{X: 100, Y: 100})
	proj.AddComponent(&VelocityComponent{VX: 100, VY: 0}) // Moving right at 100 px/s
	proj.AddComponent(NewProjectileComponent(10.0, 100.0, 1.0, "arrow", 999))
	
	// Update for 0.1 seconds
	deltaTime := 0.1
	entities := []*Entity{proj}
	ps.Update(entities, deltaTime)
	
	// Check projectile moved
	posComp := proj.GetComponent("position").(*PositionComponent)
	expectedX := 100.0 + 100.0*deltaTime // 110.0
	if posComp.X != expectedX {
		t.Errorf("expected X position %f, got %f", expectedX, posComp.X)
	}
	
	// Check projectile aged
	projComp := proj.GetComponent("projectile").(*ProjectileComponent)
	if projComp.Age != deltaTime {
		t.Errorf("expected age %f, got %f", deltaTime, projComp.Age)
	}
	
	// Update for remaining lifetime (projectile should despawn)
	ps.Update(entities, 1.0) // Total age now 1.1 seconds, lifetime is 1.0
	
	// Projectile should be expired (marked for removal)
	if !projComp.IsExpired() {
		t.Error("projectile should be expired after exceeding lifetime")
	}
}

// TestNoProjectileSpawnWithoutWeapon tests attack without equipped weapon
func TestNoProjectileSpawnWithoutWeapon(t *testing.T) {
	world := NewWorld()
	ps := NewProjectileSystem(world)
	cs := NewCombatSystem(12345)
	cs.world = world
	cs.genreID = "fantasy"
	cs.projectileSystem = ps
	
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	attacker.AddComponent(&AimComponent{AimAngle: 0})
	attacker.AddComponent(&AttackComponent{Damage: 5.0, Range: 50.0, AttackSpeed: 1.0})
	// No equipment component - unarmed
	
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 120, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	
	entitiesBefore := len(world.Entities)
	cs.Attack(attacker, target)
	entitiesAfter := len(world.Entities)
	
	// No projectile should be spawned (falls through to melee)
	if entitiesAfter != entitiesBefore {
		t.Error("unarmed attack should not spawn projectile")
	}
}
