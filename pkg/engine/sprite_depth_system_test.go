package engine

import (
	"testing"
)

func TestNewSpriteDepthEnhanceSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 42 {
		t.Errorf("expected seed 42, got %d", sys.seed)
	}
	if sys.world != world {
		t.Error("world reference mismatch")
	}
}

func TestSpriteDepthEnhanceSystem_SkipsNilSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	entity := NewEntity(1)
	// No sprite component — should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestSpriteDepthEnhanceSystem_SkipsInvisibleSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	entity := NewEntity(1)
	sprite := &EbitenSprite{Visible: false}
	entity.AddComponent(sprite)
	sys.Update([]*Entity{entity}, 0.016)
	if sprite.DepthProcessed {
		t.Error("invisible sprite should not be depth processed")
	}
}

func TestSpriteDepthEnhanceSystem_SkipsAlreadyProcessed(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	entity := NewEntity(1)
	sprite := &EbitenSprite{Visible: true, DepthProcessed: true}
	entity.AddComponent(sprite)
	sys.Update([]*Entity{entity}, 0.016)
	// No crash, flag still set
	if !sprite.DepthProcessed {
		t.Error("flag should remain true")
	}
}

func TestSpriteDepthEnhanceSystem_SkipsFinalized(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	entity := NewEntity(1)
	sprite := &EbitenSprite{Visible: true, Finalized: true}
	entity.AddComponent(sprite)
	sys.Update([]*Entity{entity}, 0.016)
	if sprite.DepthProcessed {
		t.Error("finalized sprite should be skipped")
	}
}

func TestSpriteDepthEnhanceSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genre != "horror" {
		t.Errorf("expected horror genre, got %s", sys.genre)
	}
}

func TestSpriteDepthEnhanceSystem_BuildConfig(t *testing.T) {
	genres := []string{"", "horror", "cyberpunk", "scifi", "sci-fi", "postapoc", "post-apocalyptic", "fantasy"}
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)

	for _, genre := range genres {
		t.Run("genre_"+genre, func(t *testing.T) {
			sys.SetGenre(genre)
			cfg := sys.buildConfig(42)
			if cfg.SpecularPower <= 0 {
				t.Error("SpecularPower should be positive")
			}
			if cfg.DiffuseStrength <= 0 {
				t.Error("DiffuseStrength should be positive")
			}
		})
	}
}

func TestSpriteDepthEnhanceSystem_GetProcessedCount(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)
	if sys.GetProcessedCount() != 0 {
		t.Error("expected 0 processed count initially")
	}
}

func TestSpriteDepthEnhanceSystem_NilWorld(t *testing.T) {
	sys := NewSpriteDepthEnhanceSystem(nil, 42)
	if sys.GetProcessedCount() != 0 {
		t.Error("expected 0 for nil world")
	}
}

func TestSpriteDepthEnhanceSystem_EntityWithCreatureVisual(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteDepthEnhanceSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{Visible: true}
	entity.AddComponent(sprite)
	entity.AddComponent(&CreatureVisualComponent{
		Form:      FormQuadruped,
		SizeClass: "medium",
	})

	// Should not panic even with nil Image
	sys.Update([]*Entity{entity}, 0.016)
}
