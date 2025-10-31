package engine

import (
	"testing"
)

// TestNewProjectilePool verifies pool initialization.
func TestNewProjectilePool(t *testing.T) {
	pool := NewProjectilePool()
	if pool == nil {
		t.Fatal("NewProjectilePool returned nil")
	}
	if pool.pool == nil {
		t.Error("Pool's sync.Pool not initialized")
	}
}

// TestProjectilePool_Get verifies component acquisition from pool.
func TestProjectilePool_Get(t *testing.T) {
	pool := NewProjectilePool()
	proj := pool.Get()

	if proj == nil {
		t.Fatal("Get returned nil")
	}

	// Verify all fields zeroed
	if proj.Damage != 0.0 {
		t.Errorf("Expected Damage 0.0, got %.2f", proj.Damage)
	}
	if proj.Speed != 0.0 {
		t.Errorf("Expected Speed 0.0, got %.2f", proj.Speed)
	}
	if proj.LifeTime != 0.0 {
		t.Errorf("Expected LifeTime 0.0, got %.2f", proj.LifeTime)
	}
	if proj.Age != 0.0 {
		t.Errorf("Expected Age 0.0, got %.2f", proj.Age)
	}
	if proj.Pierce != 0 {
		t.Errorf("Expected Pierce 0, got %d", proj.Pierce)
	}
	if proj.Bounce != 0 {
		t.Errorf("Expected Bounce 0, got %d", proj.Bounce)
	}
	if proj.Explosive {
		t.Error("Expected Explosive false")
	}
	if proj.ExplosionRadius != 0.0 {
		t.Errorf("Expected ExplosionRadius 0.0, got %.2f", proj.ExplosionRadius)
	}
	if proj.OwnerID != 0 {
		t.Errorf("Expected OwnerID 0, got %d", proj.OwnerID)
	}
	if proj.ProjectileType != "" {
		t.Errorf("Expected empty ProjectileType, got '%s'", proj.ProjectileType)
	}
	if proj.HasHit {
		t.Error("Expected HasHit false")
	}
}

// TestProjectilePool_GetMultiple verifies multiple acquisitions.
func TestProjectilePool_GetMultiple(t *testing.T) {
	pool := NewProjectilePool()

	projs := make([]*ProjectileComponent, 10)
	for i := range projs {
		projs[i] = pool.Get()
		if projs[i] == nil {
			t.Fatalf("Get %d returned nil", i)
		}
	}

	// All should be distinct pointers (initially, before pooling kicks in)
	// But after Put/Get cycles, pointers may be reused (that's the point)
}

// TestProjectilePool_PutGet verifies put-get cycle reuses components.
func TestProjectilePool_PutGet(t *testing.T) {
	pool := NewProjectilePool()

	// Get component
	proj1 := pool.Get()
	proj1.Damage = 100.0
	proj1.ProjectileType = "test"

	// Return to pool
	pool.Put(proj1)

	// Get another component (may be same object)
	proj2 := pool.Get()

	// Should be zeroed even if same object
	if proj2.Damage != 0.0 {
		t.Errorf("Expected Damage 0.0 after put-get, got %.2f", proj2.Damage)
	}
	if proj2.ProjectileType != "" {
		t.Errorf("Expected empty ProjectileType after put-get, got '%s'", proj2.ProjectileType)
	}
}

// TestProjectilePool_PutNil verifies nil handling.
func TestProjectilePool_PutNil(t *testing.T) {
	pool := NewProjectilePool()
	pool.Put(nil) // Should not panic
}

// TestVelocityPool_Get verifies velocity component acquisition.
func TestVelocityPool_Get(t *testing.T) {
	pool := NewVelocityPool()
	vel := pool.Get()

	if vel == nil {
		t.Fatal("Get returned nil")
	}

	if vel.VX != 0.0 || vel.VY != 0.0 {
		t.Errorf("Expected velocity (0, 0), got (%.2f, %.2f)", vel.VX, vel.VY)
	}
}

// TestPositionPool_Get verifies position component acquisition.
func TestPositionPool_Get(t *testing.T) {
	pool := NewPositionPool()
	pos := pool.Get()

	if pos == nil {
		t.Fatal("Get returned nil")
	}

	if pos.X != 0.0 || pos.Y != 0.0 {
		t.Errorf("Expected position (0, 0), got (%.2f, %.2f)", pos.X, pos.Y)
	}
}

// TestProjectileEntityPool_AllocateComponents verifies complete allocation.
func TestProjectileEntityPool_AllocateComponents(t *testing.T) {
	pool := NewProjectileEntityPool()
	components := pool.AllocateComponents()

	if components.Projectile == nil {
		t.Error("Projectile component nil")
	}
	if components.Velocity == nil {
		t.Error("Velocity component nil")
	}
	if components.Position == nil {
		t.Error("Position component nil")
	}

	// Verify all components zeroed
	if components.Projectile.Damage != 0.0 {
		t.Error("Projectile not zeroed")
	}
	if components.Velocity.VX != 0.0 || components.Velocity.VY != 0.0 {
		t.Error("Velocity not zeroed")
	}
	if components.Position.X != 0.0 || components.Position.Y != 0.0 {
		t.Error("Position not zeroed")
	}
}

// TestProjectileEntityPool_DeallocateComponents verifies component return.
func TestProjectileEntityPool_DeallocateComponents(t *testing.T) {
	pool := NewProjectileEntityPool()

	// Allocate
	components := pool.AllocateComponents()

	// Modify
	components.Projectile.Damage = 50.0
	components.Velocity.VX = 100.0
	components.Position.X = 200.0

	// Deallocate
	pool.DeallocateComponents(components)

	// Allocate again - should get zeroed components
	components2 := pool.AllocateComponents()
	if components2.Projectile.Damage != 0.0 {
		t.Error("Projectile not reset after deallocation cycle")
	}
	if components2.Velocity.VX != 0.0 {
		t.Error("Velocity not reset after deallocation cycle")
	}
	if components2.Position.X != 0.0 {
		t.Error("Position not reset after deallocation cycle")
	}
}

// TestProjectileEntityPool_DeallocatePartial verifies partial deallocation.
func TestProjectileEntityPool_DeallocatePartial(t *testing.T) {
	pool := NewProjectileEntityPool()

	// Allocate with some nil components
	components := ProjectileComponents{
		Projectile: pool.projectilePool.Get(),
		Velocity:   nil,
		Position:   pool.positionPool.Get(),
	}

	// Should not panic
	pool.DeallocateComponents(components)
}

// TestProjectileEntityPool_Concurrent verifies thread safety.
func TestProjectileEntityPool_Concurrent(t *testing.T) {
	pool := NewProjectileEntityPool()

	// Spawn 100 goroutines allocating and deallocating
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				components := pool.AllocateComponents()
				components.Projectile.Damage = 25.0
				components.Velocity.VX = 300.0
				components.Position.X = 50.0
				pool.DeallocateComponents(components)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// BenchmarkProjectilePool_WithPooling benchmarks acquisition with pooling.
func BenchmarkProjectilePool_WithPooling(b *testing.B) {
	pool := NewProjectilePool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proj := pool.Get()
		proj.Damage = 25.0
		proj.Speed = 300.0
		pool.Put(proj)
	}
}

// BenchmarkProjectilePool_WithoutPooling benchmarks acquisition without pooling.
func BenchmarkProjectilePool_WithoutPooling(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proj := &ProjectileComponent{}
		proj.Damage = 25.0
		proj.Speed = 300.0
		_ = proj // Simulate use
	}
}

// BenchmarkProjectileEntityPool_AllocateWithPooling benchmarks full entity allocation with pooling.
func BenchmarkProjectileEntityPool_AllocateWithPooling(b *testing.B) {
	pool := NewProjectileEntityPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		components := pool.AllocateComponents()
		components.Projectile.Damage = 25.0
		components.Velocity.VX = 300.0
		components.Position.X = 50.0
		pool.DeallocateComponents(components)
	}
}

// BenchmarkProjectileEntityPool_AllocateWithoutPooling benchmarks full entity allocation without pooling.
func BenchmarkProjectileEntityPool_AllocateWithoutPooling(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proj := &ProjectileComponent{}
		vel := &VelocityComponent{}
		pos := &PositionComponent{}
		proj.Damage = 25.0
		vel.VX = 300.0
		pos.X = 50.0
		_, _, _ = proj, vel, pos // Simulate use
	}
}

// BenchmarkProjectilePool_Contention benchmarks pooling under contention.
func BenchmarkProjectilePool_Contention(b *testing.B) {
	pool := NewProjectilePool()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			proj := pool.Get()
			proj.Damage = 25.0
			pool.Put(proj)
		}
	})
}

// BenchmarkProjectilePool_BatchAllocate benchmarks batch allocation.
func BenchmarkProjectilePool_BatchAllocate(b *testing.B) {
	pool := NewProjectilePool()
	const batchSize = 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		projs := make([]*ProjectileComponent, batchSize)
		for j := 0; j < batchSize; j++ {
			projs[j] = pool.Get()
			projs[j].Damage = 25.0
		}
		for j := 0; j < batchSize; j++ {
			pool.Put(projs[j])
		}
	}
}

// BenchmarkProjectileEntityPool_HighThroughput simulates high projectile spawn/despawn rate.
func BenchmarkProjectileEntityPool_HighThroughput(b *testing.B) {
	pool := NewProjectileEntityPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate 20 projectiles active per frame
		components := make([]ProjectileComponents, 20)
		for j := range components {
			components[j] = pool.AllocateComponents()
			components[j].Projectile.Damage = 25.0
			components[j].Velocity.VX = 300.0
			components[j].Position.X = float64(j * 10)
		}

		// Despawn half
		for j := 0; j < 10; j++ {
			pool.DeallocateComponents(components[j])
		}

		// Despawn remaining half
		for j := 10; j < 20; j++ {
			pool.DeallocateComponents(components[j])
		}
	}
}
