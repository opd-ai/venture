package engine

import (
	"testing"
)

func TestNewProjectileSystem(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	if sys == nil {
		t.Fatal("NewProjectileSystem() returned nil")
	}
	if sys.world != w {
		t.Error("ProjectileSystem world not set correctly")
	}
	if sys.quadtree != nil {
		t.Error("ProjectileSystem quadtree should be nil initially")
	}
}

func TestProjectileSystem_SetQuadtree(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	bounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	qt := NewQuadtree(bounds, 4)
	sys.SetQuadtree(qt)

	if sys.quadtree != qt {
		t.Error("SetQuadtree() did not set quadtree correctly")
	}
}

func TestProjectileSystem_SpawnProjectile(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	projComp := NewProjectileComponent(25.0, 400.0, 5.0, "arrow", 1)
	entity := sys.SpawnProjectile(100.0, 200.0, 300.0, 0.0, projComp)

	if entity == nil {
		t.Fatal("SpawnProjectile() returned nil entity")
	}

	// Check position component
	posComp, ok := entity.GetComponent("position")
	if !ok {
		t.Fatal("Spawned projectile missing position component")
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		t.Fatal("Position component has wrong type")
	}
	if pos.X != 100.0 || pos.Y != 200.0 {
		t.Errorf("Position = (%v, %v), want (100, 200)", pos.X, pos.Y)
	}

	// Check velocity component
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		t.Fatal("Spawned projectile missing velocity component")
	}
	vel, ok := velComp.(*VelocityComponent)
	if !ok {
		t.Fatal("Velocity component has wrong type")
	}
	if vel.VX != 300.0 || vel.VY != 0.0 {
		t.Errorf("Velocity = (%v, %v), want (300, 0)", vel.VX, vel.VY)
	}

	// Check projectile component
	projComp2, ok := entity.GetComponent("projectile")
	if !ok {
		t.Fatal("Spawned projectile missing projectile component")
	}
	proj, ok := projComp2.(*ProjectileComponent)
	if !ok {
		t.Fatal("Projectile component has wrong type")
	}
	if proj.Damage != 25.0 {
		t.Errorf("Damage = %v, want 25", proj.Damage)
	}
}

func TestProjectileSystem_GetProjectileCount(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Initially no projectiles
	count := sys.GetProjectileCount()
	if count != 0 {
		t.Errorf("Initial count = %v, want 0", count)
	}

	// Spawn some projectiles
	projComp1 := NewProjectileComponent(25.0, 400.0, 5.0, "arrow", 1)
	sys.SpawnProjectile(100.0, 100.0, 100.0, 0.0, projComp1)

	projComp2 := NewProjectileComponent(30.0, 500.0, 3.0, "bullet", 2)
	sys.SpawnProjectile(200.0, 200.0, 200.0, 0.0, projComp2)

	count = sys.GetProjectileCount()
	if count != 2 {
		t.Errorf("Count after spawning = %v, want 2", count)
	}
}

func TestProjectileSystem_Update_Movement(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Spawn projectile moving right
	projComp := NewProjectileComponent(25.0, 400.0, 5.0, "arrow", 1)
	entity := sys.SpawnProjectile(100.0, 100.0, 400.0, 0.0, projComp)

	// Process pending entities
	w.Update(0.0)

	// Get all entities for update
	entities := w.GetEntities()

	// Update for 1 second
	sys.Update(entities, 1.0)

	// Check position moved
	posComp, ok := entity.GetComponent("position")
	if !ok {
		t.Fatal("Position component missing")
	}
	pos := posComp.(*PositionComponent)
	expectedX := 100.0 + 400.0*1.0 // old + velocity * time
	if pos.X != expectedX {
		t.Errorf("Position X = %v, want %v", pos.X, expectedX)
	}
	if pos.Y != 100.0 {
		t.Errorf("Position Y = %v, want 100", pos.Y)
	}

	// Check age increased
	projComp2, ok := entity.GetComponent("projectile")
	if !ok {
		t.Fatal("Projectile component missing")
	}
	proj := projComp2.(*ProjectileComponent)
	if proj.Age != 1.0 {
		t.Errorf("Age = %v, want 1.0", proj.Age)
	}
}

func TestProjectileSystem_Update_Expiration(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Spawn projectile with 2 second lifetime
	projComp := NewProjectileComponent(25.0, 400.0, 2.0, "arrow", 1)
	entity := sys.SpawnProjectile(100.0, 100.0, 100.0, 0.0, projComp)
	entityID := entity.ID

	// Process pending entities
	w.Update(0.0)

	// Get all entities for update
	entities := w.GetEntities()

	// Update for 1 second - should still exist
	sys.Update(entities, 1.0)
	w.Update(0.0)
	if _, ok := w.GetEntity(entityID); !ok {
		t.Error("Projectile expired early at 1 second")
	}

	// Update for another 1.5 seconds - should expire
	entities = w.GetEntities()
	sys.Update(entities, 1.5)
	w.Update(0.0)
	if _, ok := w.GetEntity(entityID); ok {
		t.Error("Projectile did not expire after lifetime")
	}
}

func TestProjectileSystem_EntityCollision(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Create target entity with health
	target := w.CreateEntity()
	target.AddComponent(&PositionComponent{X: 200.0, Y: 100.0})
	target.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	// Spawn projectile aimed at target
	projComp := NewProjectileComponent(25.0, 400.0, 5.0, "arrow", 1)
	entity := sys.SpawnProjectile(100.0, 100.0, 400.0, 0.0, projComp)
	entityID := entity.ID

	// Process pending entities
	w.Update(0.0)

	// Get all entities for update
	entities := w.GetEntities()

	// Update - projectile should hit target and despawn
	sys.Update(entities, 0.5) // Move projectile to ~300, should hit target at 200
	w.Update(0.0)

	// Check projectile despawned
	if _, ok := w.GetEntity(entityID); ok {
		t.Error("Projectile should have despawned after hit")
	}

	// Check target took damage
	healthComp, ok := target.GetComponent("health")
	if !ok {
		t.Fatal("Target health component missing")
	}
	health := healthComp.(*HealthComponent)
	expectedHealth := 100.0 - 25.0
	if health.Current != expectedHealth {
		t.Errorf("Target health = %v, want %v", health.Current, expectedHealth)
	}
}

func TestProjectileSystem_PierceCollision(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Create two target entities
	target1 := w.CreateEntity()
	target1.AddComponent(&PositionComponent{X: 120.0, Y: 100.0})
	target1.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	target2 := w.CreateEntity()
	target2.AddComponent(&PositionComponent{X: 160.0, Y: 100.0})
	target2.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	// Spawn piercing projectile
	projComp := NewPiercingProjectile(30.0, 200.0, 5.0, 1, "piercing_arrow", 1)
	entity := sys.SpawnProjectile(100.0, 100.0, 200.0, 0.0, projComp)
	entityID := entity.ID

	// Process pending entities
	w.Update(0.0)

	// Get all entities for update
	entities := w.GetEntities()

	// Update - should hit first target but continue
	sys.Update(entities, 0.15) // Move to ~130, hit target1
	w.Update(0.0)

	// Projectile should still exist (pierce = 1)
	if _, ok := w.GetEntity(entityID); !ok {
		t.Error("Piercing projectile should not despawn after first hit")
	}

	// First target should take damage
	healthComp1, ok := target1.GetComponent("health")
	if !ok {
		t.Fatal("Target1 health component missing")
	}
	health1 := healthComp1.(*HealthComponent)
	if health1.Current >= 100.0 {
		t.Error("First target should have taken damage")
	}

	// Update again - should hit second target and despawn
	entities = w.GetEntities()
	sys.Update(entities, 0.25) // Move to ~180, hit target2
	w.Update(0.0)

	// Projectile should now be despawned
	if _, ok := w.GetEntity(entityID); ok {
		t.Error("Piercing projectile should despawn after using all pierce charges")
	}

	// Second target should take damage
	healthComp2, ok := target2.GetComponent("health")
	if !ok {
		t.Fatal("Target2 health component missing")
	}
	health2 := healthComp2.(*HealthComponent)
	if health2.Current >= 100.0 {
		t.Error("Second target should have taken damage")
	}
}

func TestProjectileSystem_ExplosiveProjectile(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Create multiple targets within explosion radius
	target1 := w.CreateEntity()
	target1.AddComponent(&PositionComponent{X: 210.0, Y: 100.0}) // Close to explosion
	target1.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	target2 := w.CreateEntity()
	target2.AddComponent(&PositionComponent{X: 250.0, Y: 100.0}) // At edge of explosion
	target2.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	target3 := w.CreateEntity()
	target3.AddComponent(&PositionComponent{X: 400.0, Y: 100.0}) // Outside explosion
	target3.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	// Spawn explosive projectile
	projComp := NewExplosiveProjectile(50.0, 400.0, 5.0, 50.0, "grenade", 1)
	sys.SpawnProjectile(100.0, 100.0, 400.0, 0.0, projComp)

	// Process pending entities
	w.Update(0.0)

	// Get all entities for update
	entities := w.GetEntities()

	// Update - projectile hits first target and explodes
	sys.Update(entities, 0.3)
	w.Update(0.0)

	// Check damage to targets
	healthComp1, ok := target1.GetComponent("health")
	if !ok {
		t.Fatal("Target1 health component missing")
	}
	health1 := healthComp1.(*HealthComponent)
	if health1.Current >= 100.0 {
		t.Error("Target 1 should have taken explosion damage")
	}

	healthComp2, ok := target2.GetComponent("health")
	if !ok {
		t.Fatal("Target2 health component missing")
	}
	health2 := healthComp2.(*HealthComponent)
	if health2.Current >= 100.0 {
		t.Error("Target 2 should have taken explosion damage")
	}

	// Target 1 should take more damage than target 2 (closer to center)
	damage1 := 100.0 - health1.Current
	damage2 := 100.0 - health2.Current
	if damage1 <= damage2 {
		t.Error("Closer target should take more explosion damage")
	}

	// Target 3 should not take damage (outside radius)
	healthComp3, ok := target3.GetComponent("health")
	if !ok {
		t.Fatal("Target3 health component missing")
	}
	health3 := healthComp3.(*HealthComponent)
	if health3.Current != 100.0 {
		t.Error("Target 3 outside explosion radius should not take damage")
	}
}

func TestProjectileSystem_NilWorld(t *testing.T) {
	sys := NewProjectileSystem(nil)

	// Update should not crash with nil world
	var entities []*Entity
	sys.Update(entities, 1.0)

	// SpawnProjectile should return nil with nil world
	projComp := NewProjectileComponent(25.0, 400.0, 5.0, "arrow", 1)
	entity := sys.SpawnProjectile(100.0, 100.0, 100.0, 0.0, projComp)
	if entity != nil {
		t.Error("SpawnProjectile should return nil when world is nil")
	}

	// GetProjectileCount should return 0 with nil world
	count := sys.GetProjectileCount()
	if count != 0 {
		t.Errorf("GetProjectileCount() with nil world = %v, want 0", count)
	}
}
