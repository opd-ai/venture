package engine

import (
	"testing"
)

// BenchmarkProjectileSystem_CheckEntityCollision benchmarks the collision detection
// which was optimized to use typed getters for ~10x faster component access.
func BenchmarkProjectileSystem_CheckEntityCollision(b *testing.B) {
	world := NewWorld()
	sys := NewProjectileSystem(world)

	// Create 100 target entities with position and health (using AddComponent to populate cache)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 50), Y: float64(i * 50)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}
	world.FlushPendingEntities()

	// Create a projectile at position (0, 0)
	projEntity := world.CreateEntity()
	projEntity.AddComponent(&PositionComponent{X: 0, Y: 0})
	projEntity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	projComp := &ProjectileComponent{
		Damage:   10,
		Speed:    100,
		LifeTime: 5.0,
		OwnerID:  999, // Non-existent owner
	}
	projEntity.AddComponent(projComp)
	world.FlushPendingEntities()

	posComp := projEntity.GetPosition()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sys.checkEntityCollision(projEntity, posComp, projComp)
	}
}

// BenchmarkProjectileSystem_Update benchmarks the full update cycle with multiple projectiles.
func BenchmarkProjectileSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewProjectileSystem(world)

	// Create 50 target entities with position and health
	for i := 0; i < 50; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: float64(i * 100)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	// Create 10 projectile entities
	for i := 0; i < 10; i++ {
		projEntity := world.CreateEntity()
		projEntity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		projEntity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
		projEntity.AddComponent(&ProjectileComponent{
			Damage:   10,
			Speed:    100,
			LifeTime: 5.0,
			OwnerID:  999,
		})
	}
	world.FlushPendingEntities()

	entities := world.GetEntities()
	deltaTime := 0.016 // 60 FPS

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sys.Update(entities, deltaTime)
	}
}
