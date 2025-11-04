package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// Phase 14.2: Tests for viewport culling and distance-based LOD

// TestAnimationSystem_ViewportCulling tests viewport culling optimization.
func TestAnimationSystem_ViewportCulling(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Create a world
	world := NewWorld()

	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)

	// Create player entity with camera
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 400, Y: 300})
	cameraComp := NewCameraComponent()
	cameraComp.X = 400
	cameraComp.Y = 300
	cameraComp.Zoom = 1.0
	player.AddComponent(cameraComp)
	cameraSystem.SetActiveCamera(player)

	// Configure animation system with camera
	sys.SetCameraSystem(cameraSystem)
	sys.EnableViewportCulling(true)

	// Create entities at different positions
	tests := []struct {
		name          string
		x, y          float64
		shouldAnimate bool // Whether entity should be animated (in viewport)
	}{
		{"in viewport", 400, 300, true},
		{"near edge", 800, 300, true},   // Within margin
		{"far left", -500, 300, false},  // Outside viewport
		{"far right", 1500, 300, false}, // Outside viewport
		{"far up", 400, -500, false},    // Outside viewport
		{"far down", 400, 1200, false},  // Outside viewport
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create entity
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: tt.x, Y: tt.y})
			entity.AddComponent(&EbitenSprite{Width: 28, Height: 28, Visible: true, Image: ebiten.NewImage(28, 28)})

			animComp := NewAnimationComponent(12345 + int64(entity.ID))
			animComp.CurrentState = AnimationStateIdle
			animComp.Playing = true
			animComp.Frames = []*ebiten.Image{ebiten.NewImage(28, 28)}
			animComp.FrameIndex = 0
			animComp.TimeAccumulator = 0
			entity.AddComponent(animComp)

			// Update animation system
			err := sys.Update([]*Entity{player, entity}, 0.1)
			if err != nil {
				t.Fatalf("Update failed: %v", err)
			}

			// Check statistics
			stats := sys.GetStats()
			if tt.shouldAnimate {
				if stats.CulledByViewport >= 1 {
					t.Errorf("Expected entity to be animated, but was culled")
				}
			} else {
				if stats.CulledByViewport < 1 {
					t.Errorf("Expected entity to be culled, but was animated")
				}
			}
		})
	}
}

// TestAnimationSystem_DistanceLOD tests distance-based level-of-detail.
func TestAnimationSystem_DistanceLOD(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Create a world
	world := NewWorld()

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 400, Y: 300})
	player.AddComponent(&InputComponent{})

	// Configure animation system with player
	sys.SetPlayerEntity(player)
	sys.EnableDistanceLOD(true)
	sys.SetDistanceThresholds(200.0, 400.0)

	// Create entities at different distances
	tests := []struct {
		name         string
		x, y         float64
		expectedTier string // "full", "half", or "static"
	}{
		{"close", 450, 300, "full"},       // 50px away
		{"mid-close", 550, 300, "full"},   // 150px away
		{"mid-far", 650, 300, "half"},     // 250px away
		{"far", 900, 300, "static"},       // 500px away
		{"very far", 1200, 800, "static"}, // ~850px away
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create entity
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: tt.x, Y: tt.y})
			entity.AddComponent(&EbitenSprite{Width: 28, Height: 28, Visible: true, Image: ebiten.NewImage(28, 28)})

			animComp := NewAnimationComponent(12345 + int64(entity.ID))
			animComp.CurrentState = AnimationStateWalk
			animComp.Playing = true
			animComp.Frames = []*ebiten.Image{
				ebiten.NewImage(28, 28),
				ebiten.NewImage(28, 28),
				ebiten.NewImage(28, 28),
				ebiten.NewImage(28, 28),
			}
			animComp.FrameIndex = 0
			animComp.TimeAccumulator = 0
			animComp.FrameTime = 0.1
			entity.AddComponent(animComp)

			// Update animation system multiple times to advance frames
			for i := 0; i < 5; i++ {
				err := sys.Update([]*Entity{player, entity}, 0.1)
				if err != nil {
					t.Fatalf("Update failed: %v", err)
				}
			}

			// Check statistics
			stats := sys.GetStats()
			switch tt.expectedTier {
			case "full":
				if stats.FullRateEntities < 1 {
					t.Errorf("Expected entity at full rate, got stats: %+v", stats)
				}
			case "half":
				if stats.HalfRateEntities < 1 {
					t.Errorf("Expected entity at half rate, got stats: %+v", stats)
				}
			case "static":
				if stats.StaticEntities < 1 {
					t.Errorf("Expected entity to be static, got stats: %+v", stats)
				}
			}
		})
	}
}

// TestAnimationSystem_Configuration tests system configuration methods.
func TestAnimationSystem_Configuration(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Test default values
	if !sys.enableViewportCull {
		t.Error("Expected viewport culling enabled by default")
	}
	if !sys.enableDistanceLOD {
		t.Error("Expected distance LOD enabled by default")
	}
	if sys.distanceCloseThresh != 200.0 {
		t.Errorf("Expected close threshold 200, got %f", sys.distanceCloseThresh)
	}
	if sys.distanceMidThresh != 400.0 {
		t.Errorf("Expected mid threshold 400, got %f", sys.distanceMidThresh)
	}

	// Test configuration methods
	sys.EnableViewportCulling(false)
	if sys.enableViewportCull {
		t.Error("Expected viewport culling disabled after EnableViewportCulling(false)")
	}

	sys.EnableDistanceLOD(false)
	if sys.enableDistanceLOD {
		t.Error("Expected distance LOD disabled after EnableDistanceLOD(false)")
	}

	sys.SetDistanceThresholds(300.0, 600.0)
	if sys.distanceCloseThresh != 300.0 || sys.distanceMidThresh != 600.0 {
		t.Errorf("Expected thresholds (300, 600), got (%f, %f)",
			sys.distanceCloseThresh, sys.distanceMidThresh)
	}

	// Test camera system configuration
	cameraSystem := NewCameraSystem(800, 600)
	sys.SetCameraSystem(cameraSystem)
	if sys.cameraSystem != cameraSystem {
		t.Error("Expected camera system to be set")
	}

	// Test player entity configuration
	world := NewWorld()
	player := world.CreateEntity()
	sys.SetPlayerEntity(player)
	if sys.playerEntity != player {
		t.Error("Expected player entity to be set")
	}
}

// TestAnimationSystem_Statistics tests performance statistics.
func TestAnimationSystem_Statistics(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 400, Y: 300})
	player.AddComponent(&InputComponent{})

	// Create multiple entities at different positions
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		// Place entities at different distances
		distance := float64(i * 100)
		entity.AddComponent(&PositionComponent{X: 400 + distance, Y: 300})
		entity.AddComponent(&EbitenSprite{Width: 28, Height: 28, Visible: true, Image: ebiten.NewImage(28, 28)})

		animComp := NewAnimationComponent(12345 + int64(i))
		animComp.Frames = []*ebiten.Image{ebiten.NewImage(28, 28)}
		entity.AddComponent(animComp)
	}

	// Configure system
	sys.SetPlayerEntity(player)
	sys.EnableDistanceLOD(true)

	// Update
	entities := world.GetEntities()
	err := sys.Update(entities, 0.1)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Check statistics
	stats := sys.GetStats()
	if stats.TotalEntities != len(entities) {
		t.Errorf("Expected %d total entities, got %d", len(entities), stats.TotalEntities)
	}

	if stats.AnimatedEntities < 1 {
		t.Error("Expected at least 1 animated entity")
	}

	if stats.FullRateEntities < 1 && stats.HalfRateEntities < 1 && stats.StaticEntities < 1 {
		t.Error("Expected at least one entity in some distance tier")
	}

	// Verify stats add up (excluding culled entities)
	nonCulled := stats.AnimatedEntities - stats.CulledByViewport
	categorized := stats.FullRateEntities + stats.HalfRateEntities + stats.StaticEntities
	if nonCulled > 0 && categorized == 0 {
		t.Error("Expected entities to be categorized by distance")
	}
}
