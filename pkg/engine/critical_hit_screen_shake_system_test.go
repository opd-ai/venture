package engine

import (
	"testing"
)

func TestNewCriticalHitScreenShakeSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 12345)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", sys.seed)
	}
	if sys.baseIntensity != 4.0 {
		t.Errorf("expected default baseIntensity 4.0, got %f", sys.baseIntensity)
	}
	if sys.baseDuration != 0.15 {
		t.Errorf("expected default baseDuration 0.15, got %f", sys.baseDuration)
	}
}

func TestCriticalHitScreenShakeSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name          string
		genreID       string
		wantIntensity float64
		wantDuration  float64
		wantScale     float64
	}{
		{"horror", "horror", 6.0, 0.25, 0.03},
		{"cyberpunk", "cyberpunk", 5.0, 0.10, 0.025},
		{"fantasy", "fantasy", 4.0, 0.15, 0.02},
		{"scifi", "scifi", 3.5, 0.12, 0.02},
		{"postapoc", "postapoc", 5.5, 0.20, 0.025},
		{"default", "unknown", 4.0, 0.15, 0.02},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCriticalHitScreenShakeSystem(world, 42)
			sys.SetGenre(tt.genreID)

			if sys.baseIntensity != tt.wantIntensity {
				t.Errorf("baseIntensity = %f, want %f", sys.baseIntensity, tt.wantIntensity)
			}
			if sys.baseDuration != tt.wantDuration {
				t.Errorf("baseDuration = %f, want %f", sys.baseDuration, tt.wantDuration)
			}
			if sys.damageScale != tt.wantScale {
				t.Errorf("damageScale = %f, want %f", sys.damageScale, tt.wantScale)
			}
		})
	}
}

func TestCriticalHitScreenShakeSystem_OnCriticalHit(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)

	attacker := NewEntity(1)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 200})
	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 150, Y: 250})

	sys.OnCriticalHit(attacker, target, 50.0)

	if len(sys.pendingShakes) != 1 {
		t.Fatalf("expected 1 pending shake, got %d", len(sys.pendingShakes))
	}
	evt := sys.pendingShakes[0]
	if evt.damage != 50.0 {
		t.Errorf("damage = %f, want 50.0", evt.damage)
	}
	if evt.attackerX != 100 || evt.attackerY != 200 {
		t.Errorf("attacker pos = (%f,%f), want (100,200)", evt.attackerX, evt.attackerY)
	}
	if evt.targetX != 150 || evt.targetY != 250 {
		t.Errorf("target pos = (%f,%f), want (150,250)", evt.targetX, evt.targetY)
	}
}

func TestCriticalHitScreenShakeSystem_OnCriticalHit_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)

	attacker := NewEntity(1)
	target := NewEntity(2)
	sys.OnCriticalHit(attacker, target, 25.0)

	if len(sys.pendingShakes) != 1 {
		t.Fatalf("expected 1 pending shake, got %d", len(sys.pendingShakes))
	}
	if sys.pendingShakes[0].attackerX != 0 || sys.pendingShakes[0].targetX != 0 {
		t.Error("expected zero positions for entities without PositionComponent")
	}
}

func TestCriticalHitScreenShakeSystem_Update_NoCameraSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)

	attacker := NewEntity(1)
	target := NewEntity(2)
	sys.OnCriticalHit(attacker, target, 30.0)

	// Should not panic with nil camera system
	sys.Update(nil, 0.016)

	// Pending shakes should remain since camera system is nil
	if len(sys.pendingShakes) != 1 {
		t.Errorf("expected pending shakes preserved when no camera, got %d", len(sys.pendingShakes))
	}
}

func TestCriticalHitScreenShakeSystem_Update_NoPendingShakes(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)
	camera := NewCameraSystem(800, 600)
	sys.SetCameraSystem(camera)

	// No pending shakes - should be a no-op
	sys.Update(nil, 0.016)
}

func TestCriticalHitScreenShakeSystem_Update_ProcessesStrongest(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)
	sys.SetGenre("fantasy")

	camera := NewCameraSystem(800, 600)
	camEntity := NewEntity(1)
	camEntity.AddComponent(&PositionComponent{X: 150, Y: 250})
	camEntity.AddComponent(&CameraComponent{})
	camera.SetActiveCamera(camEntity)
	sys.SetCameraSystem(camera)

	attacker := NewEntity(2)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 200})
	target := NewEntity(3)
	target.AddComponent(&PositionComponent{X: 150, Y: 250})

	// Queue multiple crits - strongest should be processed
	sys.OnCriticalHit(attacker, target, 20.0)
	sys.OnCriticalHit(attacker, target, 80.0)
	sys.OnCriticalHit(attacker, target, 40.0)

	sys.Update(nil, 0.016)

	// All pending shakes should be consumed
	if len(sys.pendingShakes) != 0 {
		t.Errorf("expected 0 pending shakes after update, got %d", len(sys.pendingShakes))
	}
}

func TestCriticalHitScreenShakeSystem_Update_DistanceAttenuation(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)
	sys.SetGenre("fantasy")

	camera := NewCameraSystem(800, 600)
	camEntity := NewEntity(1)
	// Camera far away (>600px)
	camEntity.AddComponent(&PositionComponent{X: 1000, Y: 1000})
	camEntity.AddComponent(&CameraComponent{})
	camera.SetActiveCamera(camEntity)
	sys.SetCameraSystem(camera)

	attacker := NewEntity(2)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	target := NewEntity(3)
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.OnCriticalHit(attacker, target, 50.0)
	sys.Update(nil, 0.016)

	// Should have been skipped (distance > 600)
	if len(sys.pendingShakes) != 0 {
		t.Errorf("expected 0 pending shakes, got %d", len(sys.pendingShakes))
	}
}

func TestCriticalHitScreenShakeSystem_IntensityClamping(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)
	sys.SetGenre("fantasy") // baseIntensity = 4.0, damageScale = 0.02

	camera := NewCameraSystem(800, 600)
	camEntity := NewEntity(1)
	camEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	camEntity.AddComponent(&CameraComponent{})
	camera.SetActiveCamera(camEntity)
	sys.SetCameraSystem(camera)

	attacker := NewEntity(2)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	target := NewEntity(3)
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Very high damage that would exceed max
	sys.OnCriticalHit(attacker, target, 10000.0)
	sys.Update(nil, 0.016)

	// System should not panic and should consume the event
	if len(sys.pendingShakes) != 0 {
		t.Errorf("expected 0 pending shakes after update, got %d", len(sys.pendingShakes))
	}
}

func TestCriticalHitScreenShakeSystem_SetCameraSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)

	if sys.cameraSystem != nil {
		t.Error("expected nil camera system initially")
	}

	camera := NewCameraSystem(800, 600)
	sys.SetCameraSystem(camera)

	if sys.cameraSystem != camera {
		t.Error("camera system not set correctly")
	}
}

func TestCriticalHitScreenShakeSystem_NilWorld(t *testing.T) {
	sys := NewCriticalHitScreenShakeSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
	// Should not panic with nil logger
	sys.SetGenre("fantasy")
	sys.SetCameraSystem(nil)
}

func BenchmarkCriticalHitScreenShakeSystem_OnCriticalHit(b *testing.B) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)

	attacker := NewEntity(1)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 200})
	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 150, Y: 250})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnCriticalHit(attacker, target, 50.0)
		sys.pendingShakes = sys.pendingShakes[:0] // Reset between iterations
	}
}

func BenchmarkCriticalHitScreenShakeSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewCriticalHitScreenShakeSystem(world, 42)
	sys.SetGenre("fantasy")

	camera := NewCameraSystem(800, 600)
	camEntity := NewEntity(1)
	camEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	camEntity.AddComponent(&CameraComponent{})
	camera.SetActiveCamera(camEntity)
	sys.SetCameraSystem(camera)

	attacker := NewEntity(2)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	target := NewEntity(3)
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnCriticalHit(attacker, target, 50.0)
		sys.Update(nil, 0.016)
	}
}
