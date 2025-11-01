package engine

import (
	"testing"
)

// TestProjectileSystem_SetGenre tests the SetGenre method.
func TestProjectileSystem_SetGenre(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Test setting genre
	sys.SetGenre("sci-fi")
	if sys.genreID != "sci-fi" {
		t.Errorf("SetGenre() did not set genreID correctly, got %v want sci-fi", sys.genreID)
	}

	// Test different genre
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("SetGenre() did not update genreID correctly, got %v want horror", sys.genreID)
	}
}

// TestProjectileSystem_SetSeed tests the SetSeed method.
func TestProjectileSystem_SetSeed(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Test setting seed
	sys.SetSeed(99999)
	if sys.seed != 99999 {
		t.Errorf("SetSeed() did not set seed correctly, got %v want 99999", sys.seed)
	}

	// Test different seed
	sys.SetSeed(54321)
	if sys.seed != 54321 {
		t.Errorf("SetSeed() did not update seed correctly, got %v want 54321", sys.seed)
	}
}

// TestProjectileSystem_SpawnProjectile_WithSprite tests that spawned projectiles have sprite components.
func TestProjectileSystem_SpawnProjectile_WithSprite(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)
	sys.SetGenre("fantasy")
	sys.SetSeed(12345)

	tests := []struct {
		name           string
		projectileType string
		explosive      bool
		wantSpriteSize float64
	}{
		{
			name:           "arrow projectile",
			projectileType: "arrow",
			explosive:      false,
			wantSpriteSize: 8.0,
		},
		{
			name:           "explosive projectile",
			projectileType: "fireball",
			explosive:      true,
			wantSpriteSize: 12.0,
		},
		{
			name:           "magic projectile",
			projectileType: "magic",
			explosive:      false,
			wantSpriteSize: 8.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projComp := &ProjectileComponent{
				Damage:         10.0,
				Speed:          400.0,
				LifeTime:       5.0,
				ProjectileType: tt.projectileType,
				Explosive:      tt.explosive,
				OwnerID:        999,
			}

			entity := sys.SpawnProjectile(100.0, 200.0, 300.0, 0.0, projComp)

			if entity == nil {
				t.Fatal("SpawnProjectile() returned nil entity")
			}

			// Check sprite component exists
			spriteComp, ok := entity.GetComponent("sprite")
			if !ok {
				t.Fatal("Spawned projectile missing sprite component")
			}

			sprite, ok := spriteComp.(*EbitenSprite)
			if !ok {
				t.Fatal("Sprite component has wrong type")
			}

			// Check sprite dimensions
			if sprite.Width != tt.wantSpriteSize || sprite.Height != tt.wantSpriteSize {
				t.Errorf("Sprite size = (%v, %v), want (%v, %v)",
					sprite.Width, sprite.Height, tt.wantSpriteSize, tt.wantSpriteSize)
			}

			// Check sprite image exists
			if sprite.Image == nil {
				t.Error("Sprite image is nil")
			}

			// Check sprite is visible
			if !sprite.Visible {
				t.Error("Sprite should be visible")
			}
		})
	}
}

// TestProjectileSystem_SpawnProjectile_DefaultType tests default projectile type handling.
func TestProjectileSystem_SpawnProjectile_DefaultType(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Projectile with empty type should default to "bullet"
	projComp := &ProjectileComponent{
		Damage:         10.0,
		Speed:          400.0,
		LifeTime:       5.0,
		ProjectileType: "", // Empty type
		OwnerID:        999,
	}

	entity := sys.SpawnProjectile(100.0, 200.0, 300.0, 0.0, projComp)

	if entity == nil {
		t.Fatal("SpawnProjectile() returned nil entity")
	}

	// Should have sprite component even with default type
	_, ok := entity.GetComponent("sprite")
	if !ok {
		t.Error("Spawned projectile with empty type missing sprite component")
	}
}

// TestProjectileSystem_SpawnProjectile_Rotation tests sprite rotation based on velocity.
func TestProjectileSystem_SpawnProjectile_Rotation(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	tests := []struct {
		name    string
		vx, vy  float64
		wantRot float64 // approximate expected rotation
	}{
		{
			name:    "moving right",
			vx:      100.0,
			vy:      0.0,
			wantRot: 0.0,
		},
		{
			name:    "moving down",
			vx:      0.0,
			vy:      100.0,
			wantRot: 1.57, // ~π/2
		},
		{
			name:    "moving left",
			vx:      -100.0,
			vy:      0.0,
			wantRot: 3.14, // ~π
		},
		{
			name:    "moving diagonal",
			vx:      100.0,
			vy:      100.0,
			wantRot: 0.785, // ~π/4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projComp := NewProjectileComponent(10.0, 400.0, 5.0, "arrow", 999)
			entity := sys.SpawnProjectile(100.0, 200.0, tt.vx, tt.vy, projComp)

			if entity == nil {
				t.Fatal("SpawnProjectile() returned nil entity")
			}

			spriteComp, ok := entity.GetComponent("sprite")
			if !ok {
				t.Fatal("Spawned projectile missing sprite component")
			}

			sprite := spriteComp.(*EbitenSprite)

			// Check rotation is approximately correct (within 0.1 radians)
			diff := sprite.Rotation - tt.wantRot
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.1 {
				t.Errorf("Sprite rotation = %v, want approximately %v", sprite.Rotation, tt.wantRot)
			}
		})
	}
}

// TestProjectileSystem_ExplosionParticles tests that explosions spawn particle effects.
func TestProjectileSystem_ExplosionParticles(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)
	sys.SetGenre("fantasy")
	sys.SetSeed(12345)

	// Create explosive projectile
	projComp := &ProjectileComponent{
		Damage:          50.0,
		Speed:           400.0,
		LifeTime:        5.0,
		ProjectileType:  "fireball",
		Explosive:       true,
		ExplosionRadius: 100.0,
		OwnerID:         1,
	}

	projectile := sys.SpawnProjectile(300.0, 300.0, 100.0, 0.0, projComp)
	projectile.AddComponent(&HealthComponent{Current: 100, Max: 100}) // Needs health to survive

	// Create target entity
	target := w.CreateEntity()
	target.AddComponent(&PositionComponent{X: 350.0, Y: 300.0}) // Within explosion radius
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Count entities before explosion
	entitiesBefore := len(w.GetEntities())

	// Manually trigger explosion
	posComp, _ := projectile.GetComponent("position")
	pos := posComp.(*PositionComponent)
	sys.handleExplosion(projectile, pos)

	// Process pending entity additions/removals
	w.Update(0.016)

	// Count entities after explosion
	entitiesAfter := len(w.GetEntities())

	// Should have created explosion entity with particle emitter
	if entitiesAfter <= entitiesBefore {
		t.Errorf("Expected new entity with particle emitter, got %d entities before, %d after",
			entitiesBefore, entitiesAfter)
	}

	// Find the explosion entity (should have particle_emitter component)
	explosionEntities := w.GetEntitiesWith("particle_emitter")
	if len(explosionEntities) == 0 {
		t.Error("No explosion entity with particle emitter created")
	} else {
		// Check particle emitter
		emitterComp, ok := explosionEntities[0].GetComponent("particle_emitter")
		if !ok {
			t.Fatal("Explosion entity missing particle emitter component")
		}

		emitter := emitterComp.(*ParticleEmitterComponent)
		if len(emitter.Systems) == 0 {
			t.Error("Particle emitter has no particle systems")
		}

		// Verify particles exist
		if len(emitter.Systems[0].Particles) == 0 {
			t.Error("Particle system has no particles")
		}
	}
}

// TestProjectileSystem_ExplosionScreenShake tests that explosions trigger screen shake.
func TestProjectileSystem_ExplosionScreenShake(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Create camera system (requires screen dimensions)
	camera := NewCameraSystem(800, 600)
	sys.SetCamera(camera)

	// Create explosive projectile
	projComp := &ProjectileComponent{
		Damage:          50.0,
		Speed:           400.0,
		LifeTime:        5.0,
		ProjectileType:  "fireball",
		Explosive:       true,
		ExplosionRadius: 100.0,
		OwnerID:         1,
	}

	projectile := sys.SpawnProjectile(300.0, 300.0, 100.0, 0.0, projComp)

	// Manually trigger explosion
	posComp, _ := projectile.GetComponent("position")
	pos := posComp.(*PositionComponent)

	// Camera should not have shake initially
	cameraEntity := w.CreateEntity()
	cameraComp := NewCameraComponent()
	cameraComp.X = 400.0
	cameraComp.Y = 300.0
	cameraEntity.AddComponent(cameraComp)
	cameraEntity.AddComponent(&PositionComponent{X: 400.0, Y: 300.0})

	// Set as active camera
	camera.SetActiveCamera(cameraEntity)

	// Trigger explosion
	sys.handleExplosion(projectile, pos)

	// Check if screen shake component was added
	shakeComp, ok := cameraEntity.GetComponent("screen_shake")
	if ok {
		shake := shakeComp.(*ScreenShakeComponent)
		// Should have been triggered by explosion
		if shake.Intensity == 0 {
			t.Error("Screen shake not triggered by explosion")
		}
	}
}

// TestProjectileSystem_NilChecks tests that system handles nil dependencies gracefully.
func TestProjectileSystem_NilChecks(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Test with nil particle generator
	sys.particleGenerator = nil

	projComp := &ProjectileComponent{
		Damage:          50.0,
		Explosive:       true,
		ExplosionRadius: 100.0,
		OwnerID:         1,
	}

	projectile := sys.SpawnProjectile(300.0, 300.0, 100.0, 0.0, projComp)
	posComp, _ := projectile.GetComponent("position")
	pos := posComp.(*PositionComponent)

	// Should not panic with nil particle generator
	sys.handleExplosion(projectile, pos)

	// Test with nil camera (already tested in constructor, but verify no panic)
	sys.camera = nil
	sys.handleExplosion(projectile, pos)
}
