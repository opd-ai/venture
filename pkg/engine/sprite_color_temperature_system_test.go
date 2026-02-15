package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewSpriteColorTemperatureSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewSpriteColorTemperatureSystem returned nil")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.world != world {
		t.Error("world not set")
	}
}

func TestSpriteColorTemperatureSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)
	sys.SetGenre("horror")
	if sys.genre != "horror" {
		t.Errorf("genre = %q, want %q", sys.genre, "horror")
	}
}

func TestSpriteColorTemperatureSystem_Update_SkipsNilSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)

	entity := world.CreateEntity()
	// No sprite component — should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestSpriteColorTemperatureSystem_Update_SkipsInvisible(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)

	entity := world.CreateEntity()
	sprite := &EbitenSprite{Visible: false, DepthProcessed: true}
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if sprite.ColorTempProcessed {
		t.Error("should not process invisible sprites")
	}
}

func TestSpriteColorTemperatureSystem_Update_SkipsNotDepthProcessed(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)

	entity := world.CreateEntity()
	sprite := &EbitenSprite{Visible: true, DepthProcessed: false}
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if sprite.ColorTempProcessed {
		t.Error("should not process sprites before depth enhancement")
	}
}

func TestSpriteColorTemperatureSystem_Update_SkipsFinalized(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)

	entity := world.CreateEntity()
	sprite := &EbitenSprite{Visible: true, DepthProcessed: true, Finalized: true}
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if sprite.ColorTempProcessed {
		t.Error("should not process already-finalized sprites")
	}
}

func TestSpriteColorTemperatureSystem_Update_SkipsAlreadyProcessed(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)

	entity := world.CreateEntity()
	sprite := &EbitenSprite{Visible: true, DepthProcessed: true, ColorTempProcessed: true}
	entity.AddComponent(sprite)

	// Should not re-process
	sys.Update([]*Entity{entity}, 0.016)
	// No crash = pass
}

func TestSpriteColorTemperatureSystem_GenreVariation(t *testing.T) {
	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc", ""}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpriteColorTemperatureSystem(world, 42)
			sys.SetGenre(genre)
			// Just verify no panics during update
			entity := world.CreateEntity()
			sprite := &EbitenSprite{Visible: true, DepthProcessed: true}
			entity.AddComponent(sprite)
			sys.Update([]*Entity{entity}, 0.016)
		})
	}
}

func TestSpriteColorTemperatureSystem_ProcessImageNil(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteColorTemperatureSystem(world, 1)

	result := sys.processImage(nil, defaultColorTempConfig(), 0)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func defaultColorTempConfig() sprites_ColorTemperatureConfig {
	return sprites_ColorTemperatureConfig{}
}

// sprites_ColorTemperatureConfig is a type alias used for testing.
type sprites_ColorTemperatureConfig = sprites.ColorTemperatureConfig
