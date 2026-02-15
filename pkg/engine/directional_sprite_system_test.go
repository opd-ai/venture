package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// canUseGPU returns true if ebiten GPU pixel operations are available.
// Even with a DISPLAY, ebiten requires RunGame() to be called before
// ReadPixels works. Since tests don't run the game loop, GPU-dependent
// sprite generation tests must be skipped.
func canUseGPU() bool {
	// ebiten.ReadPixels panics unless RunGame has been called,
	// which never happens in unit tests. The FinalizeEntitySprite
	// call in the sprite generator reads pixels, so full generation
	// tests are GPU-only.
	return false
}

// stubDirectionalInputComponent is a minimal input component stub for tests.
type stubDirectionalInputComponent struct{}

func (s *stubDirectionalInputComponent) Type() string { return "input" }

// TestNewDirectionalSpriteSystem verifies system creation.
func TestNewDirectionalSpriteSystem(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.spriteGenerator == nil {
		t.Error("expected non-nil sprite generator")
	}
	if sys.GetProcessedCount() != 0 {
		t.Errorf("expected 0 processed, got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_SetGenre verifies genre changes clear cache.
func TestDirectionalSpriteSystem_SetGenre(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	// Simulate a processed entity
	sys.processed[1] = 12345

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("expected genre cyberpunk, got %s", sys.genreID)
	}
	if sys.GetProcessedCount() != 0 {
		t.Errorf("expected cache cleared after genre change, got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_SetGenre_SameGenre verifies same genre doesn't clear.
func TestDirectionalSpriteSystem_SetGenre_SameGenre(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	sys.genreID = "fantasy"
	sys.processed[1] = 12345

	sys.SetGenre("fantasy")
	if sys.GetProcessedCount() != 1 {
		t.Errorf("expected cache preserved for same genre, got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_SkipsNoSprite verifies entities without sprites are skipped.
func TestDirectionalSpriteSystem_SkipsNoSprite(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	entity := NewEntity(1)
	entity.AddComponent(&AnimationComponent{Seed: 42})

	sys.Update([]*Entity{entity}, 0.016)
	if sys.GetProcessedCount() != 0 {
		t.Errorf("expected 0 processed (no sprite), got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_SkipsNoAnimation verifies entities without animation are skipped.
func TestDirectionalSpriteSystem_SkipsNoAnimation(t *testing.T) {
	if !canUseGPU() {
		t.Skip("skipping: no display available for ebiten.NewImage")
	}
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	entity := NewEntity(1)
	sprite := NewSpriteComponent(32, 32, nil)
	sprite.Image = ebiten.NewImage(32, 32)
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if sys.GetProcessedCount() != 0 {
		t.Errorf("expected 0 processed (no animation), got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_SkipsNilImage verifies entities with nil sprite image are skipped.
func TestDirectionalSpriteSystem_SkipsNilImage(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	entity := NewEntity(1)
	sprite := NewSpriteComponent(32, 32, nil)
	entity.AddComponent(sprite)
	entity.AddComponent(&AnimationComponent{Seed: 42})

	sys.Update([]*Entity{entity}, 0.016)
	if sys.GetProcessedCount() != 0 {
		t.Errorf("expected 0 processed (nil image), got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_IsAlreadyProcessed verifies cache-hit logic.
func TestDirectionalSpriteSystem_IsAlreadyProcessed(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	anim := &AnimationComponent{Seed: 12345}
	entity := NewEntity(1)

	// Not yet processed
	if sys.isAlreadyProcessed(entity, anim) {
		t.Error("expected false for unprocessed entity")
	}

	// Mark as processed
	sys.processed[entity.ID] = 12345
	if !sys.isAlreadyProcessed(entity, anim) {
		t.Error("expected true for processed entity with same seed")
	}

	// Change seed
	anim.Seed = 99999
	if sys.isAlreadyProcessed(entity, anim) {
		t.Error("expected false after seed change")
	}
}

// TestDirectionalSpriteSystem_ClearCache verifies cache clearing.
func TestDirectionalSpriteSystem_ClearCache(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	sys.processed[1] = 100
	sys.processed[2] = 200

	sys.ClearCache()
	if sys.GetProcessedCount() != 0 {
		t.Errorf("expected 0 after ClearCache, got %d", sys.GetProcessedCount())
	}
}

// TestDirectionalSpriteSystem_DetermineEntityType tests entity type classification.
func TestDirectionalSpriteSystem_DetermineEntityType(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	tests := []struct {
		name     string
		setup    func(*Entity)
		expected string
	}{
		{
			name:     "default humanoid",
			setup:    func(e *Entity) {},
			expected: "humanoid",
		},
		{
			name: "creature visual quadruped",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormQuadruped})
			},
			expected: "quadruped",
		},
		{
			name: "creature visual arachnid",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormArachnid})
			},
			expected: "arachnid",
		},
		{
			name: "creature visual serpentine",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormSerpentine})
			},
			expected: "serpentine",
		},
		{
			name: "creature visual flying",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormFlying})
			},
			expected: "flying",
		},
		{
			name: "creature visual blob",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormBlob})
			},
			expected: "blob",
		},
		{
			name: "creature visual mechanical",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormMechanical})
			},
			expected: "mechanical",
		},
		{
			name: "creature visual undead",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormUndead})
			},
			expected: "undead",
		},
		{
			name: "creature visual humanoid falls through",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormHumanoid})
			},
			expected: "humanoid",
		},
		{
			name: "boss from high damage",
			setup: func(e *Entity) {
				e.AddComponent(&AttackComponent{Damage: 75})
			},
			expected: "boss",
		},
		{
			name: "quadruped from high hp",
			setup: func(e *Entity) {
				e.AddComponent(&HealthComponent{Max: 600})
			},
			expected: "quadruped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			tt.setup(entity)
			got := sys.determineEntityType(entity)
			if got != tt.expected {
				t.Errorf("determineEntityType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDirectionalSpriteSystem_GetGenreID tests genre resolution.
func TestDirectionalSpriteSystem_GetGenreID(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)

	tests := []struct {
		name     string
		sysGenre string
		setup    func(*Entity)
		expected string
	}{
		{
			name:     "fallback to fantasy",
			sysGenre: "",
			setup:    func(e *Entity) {},
			expected: "fantasy",
		},
		{
			name:     "system genre",
			sysGenre: "horror",
			setup:    func(e *Entity) {},
			expected: "horror",
		},
		{
			name:     "cyberpunk genre",
			sysGenre: "cyberpunk",
			setup:    func(e *Entity) {},
			expected: "cyberpunk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.genreID = tt.sysGenre
			entity := NewEntity(1)
			tt.setup(entity)
			got := sys.getGenreID(entity)
			if got != tt.expected {
				t.Errorf("getGenreID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDirectionalSpriteSystem_BuildSpriteConfig tests config construction.
func TestDirectionalSpriteSystem_BuildSpriteConfig(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	sys.genreID = "fantasy"

	tests := []struct {
		name           string
		setup          func(*Entity)
		checkConfig    func(t *testing.T, cfg sprites.Config)
	}{
		{
			name:  "basic humanoid config",
			setup: func(e *Entity) {},
			checkConfig: func(t *testing.T, cfg sprites.Config) {
				if cfg.Width != 32 || cfg.Height != 32 {
					t.Errorf("expected 32x32, got %dx%d", cfg.Width, cfg.Height)
				}
				if cfg.Custom["useAerial"] != true {
					t.Error("expected useAerial=true")
				}
				if cfg.Custom["entityType"] != "humanoid" {
					t.Errorf("expected entityType=humanoid, got %v", cfg.Custom["entityType"])
				}
				if cfg.GenreID != "fantasy" {
					t.Errorf("expected genre fantasy, got %s", cfg.GenreID)
				}
			},
		},
		{
			name: "player with equipment",
			setup: func(e *Entity) {
				e.AddComponent(&stubDirectionalInputComponent{})
				e.AddComponent(&EquipmentComponent{})
			},
			checkConfig: func(t *testing.T, cfg sprites.Config) {
				if cfg.Custom["entityType"] != "humanoid" {
					t.Error("expected player entityType=humanoid")
				}
				if cfg.Custom["hasWeapon"] != true {
					t.Error("expected hasWeapon=true for equipped player")
				}
			},
		},
		{
			name: "nonhumanoid creature",
			setup: func(e *Entity) {
				e.AddComponent(&CreatureVisualComponent{Form: FormArachnid})
			},
			checkConfig: func(t *testing.T, cfg sprites.Config) {
				if cfg.Custom["entityType"] != "arachnid" {
					t.Errorf("expected entityType=arachnid, got %v", cfg.Custom["entityType"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			sprite := NewSpriteComponent(32, 32, nil)
			anim := &AnimationComponent{Seed: 42}
			entity.AddComponent(sprite)
			entity.AddComponent(anim)
			tt.setup(entity)

			cfg := sys.buildSpriteConfig(entity, sprite, anim)
			tt.checkConfig(t, cfg)
		})
	}
}

// TestDirectionalSpriteSystem_GeneratesDirectionalSprites verifies full sprite generation.
func TestDirectionalSpriteSystem_GeneratesDirectionalSprites(t *testing.T) {
	if !canUseGPU() {
		t.Skip("skipping: no display available for sprite generation")
	}
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	sys.genreID = "fantasy"

	entity := NewEntity(1)
	sprite := NewSpriteComponent(32, 32, nil)
	sprite.Image = ebiten.NewImage(32, 32)
	entity.AddComponent(sprite)
	entity.AddComponent(&AnimationComponent{Seed: 12345})

	sys.Update([]*Entity{entity}, 0.016)

	if sys.GetProcessedCount() != 1 {
		t.Errorf("expected 1 processed, got %d", sys.GetProcessedCount())
	}
	if len(sprite.DirectionalImages) != 4 {
		t.Errorf("expected 4 directional images, got %d", len(sprite.DirectionalImages))
	}
	if sprite.Finalized {
		t.Error("expected Finalized to be false after directional generation")
	}
}

// TestDirectionalSpriteSystem_FrameBudget verifies per-frame generation limit.
func TestDirectionalSpriteSystem_FrameBudget(t *testing.T) {
	if !canUseGPU() {
		t.Skip("skipping: no display available for sprite generation")
	}
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	sys.genreID = "fantasy"

	entities := make([]*Entity, maxDirectionalGenerationsPerFrame+4)
	for i := range entities {
		entity := NewEntity(uint64(i + 1))
		sprite := NewSpriteComponent(32, 32, nil)
		sprite.Image = ebiten.NewImage(32, 32)
		entity.AddComponent(sprite)
		entity.AddComponent(&AnimationComponent{Seed: int64(i * 1000)})
		entities[i] = entity
	}

	sys.Update(entities, 0.016)
	if sys.GetProcessedCount() > maxDirectionalGenerationsPerFrame {
		t.Errorf("expected at most %d processed in one frame, got %d",
			maxDirectionalGenerationsPerFrame, sys.GetProcessedCount())
	}

	sys.Update(entities, 0.016)
	if sys.GetProcessedCount() != len(entities) {
		t.Errorf("expected all %d processed after two frames, got %d",
			len(entities), sys.GetProcessedCount())
	}
}

// BenchmarkDirectionalSpriteSystem_CacheHit benchmarks the cache-hit path (no generation).
func BenchmarkDirectionalSpriteSystem_CacheHit(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewDirectionalSpriteSystem(gen)
	sys.genreID = "fantasy"

	entity := NewEntity(1)
	sprite := NewSpriteComponent(32, 32, nil)
	anim := &AnimationComponent{Seed: 12345}
	entity.AddComponent(sprite)
	entity.AddComponent(anim)

	// Mark as already processed (no GPU needed)
	sys.processed[entity.ID] = anim.Seed

	entities := []*Entity{entity}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
