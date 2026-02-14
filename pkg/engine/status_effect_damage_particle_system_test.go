package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewStatusEffectDamageParticleSystem(t *testing.T) {
	tests := []struct {
		name        string
		useWorld    bool
		seed        int64
		wantNil     bool
		wantGenre   string
		wantBurning int
		wantPoison  int
		wantRegen   int
	}{
		{
			name:        "creates system with valid world",
			useWorld:    true,
			seed:        12345,
			wantNil:     false,
			wantGenre:   "fantasy",
			wantBurning: 8,
			wantPoison:  6,
			wantRegen:   10,
		},
		{
			name:        "creates system with nil world",
			useWorld:    false,
			seed:        54321,
			wantNil:     false,
			wantGenre:   "fantasy",
			wantBurning: 8,
			wantPoison:  6,
			wantRegen:   10,
		},
		{
			name:        "different seed produces same defaults",
			useWorld:    true,
			seed:        99999,
			wantNil:     false,
			wantGenre:   "fantasy",
			wantBurning: 8,
			wantPoison:  6,
			wantRegen:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var world *World
			if tt.useWorld {
				world = NewWorldWithLogger(logrus.New())
			}

			system := NewStatusEffectDamageParticleSystem(world, tt.seed)

			if (system == nil) != tt.wantNil {
				t.Errorf("NewStatusEffectDamageParticleSystem() nil = %v, want %v", system == nil, tt.wantNil)
				return
			}

			if system.genreID != tt.wantGenre {
				t.Errorf("genreID = %v, want %v", system.genreID, tt.wantGenre)
			}
			if system.burningParticleCount != tt.wantBurning {
				t.Errorf("burningParticleCount = %v, want %v", system.burningParticleCount, tt.wantBurning)
			}
			if system.poisonParticleCount != tt.wantPoison {
				t.Errorf("poisonParticleCount = %v, want %v", system.poisonParticleCount, tt.wantPoison)
			}
			if system.regenParticleCount != tt.wantRegen {
				t.Errorf("regenParticleCount = %v, want %v", system.regenParticleCount, tt.wantRegen)
			}
			if system.seed != tt.seed {
				t.Errorf("seed = %v, want %v", system.seed, tt.seed)
			}
		})
	}
}

func TestStatusEffectDamageParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		wantGenre string
	}{
		{"fantasy genre", "fantasy", "fantasy"},
		{"scifi genre", "scifi", "scifi"},
		{"horror genre", "horror", "horror"},
		{"cyberpunk genre", "cyberpunk", "cyberpunk"},
		{"postapoc genre", "postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorldWithLogger(logrus.New())
			system := NewStatusEffectDamageParticleSystem(world, 12345)

			system.SetGenre(tt.genre)

			if system.genreID != tt.wantGenre {
				t.Errorf("genreID = %v, want %v", system.genreID, tt.wantGenre)
			}
		})
	}
}

func TestStatusEffectDamageParticleSystem_Update(t *testing.T) {
	world := NewWorldWithLogger(logrus.New())
	system := NewStatusEffectDamageParticleSystem(world, 12345)

	// Create test entities
	entities := []*Entity{
		world.CreateEntity(),
		world.CreateEntity(),
	}

	// Update should not panic even without particle system
	system.Update(entities, 0.016)

	// Update with empty entity list
	system.Update([]*Entity{}, 0.016)

	// Update with nil entities
	system.Update(nil, 0.016)
}

func TestStatusEffectDamageParticleSystem_OnStatusEffectTick_NilChecks(t *testing.T) {
	world := NewWorldWithLogger(logrus.New())

	t.Run("nil particle system", func(t *testing.T) {
		system := NewStatusEffectDamageParticleSystem(world, 12345)
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: 100, Y: 100})

		// Should not panic with nil particle system
		system.OnStatusEffectTick(entity, "burning", 5.0)
	})

	t.Run("nil entity", func(t *testing.T) {
		system := NewStatusEffectDamageParticleSystem(world, 12345)
		// Skip SetParticleSystem to avoid Ebiten initialization

		// Should not panic with nil entity
		system.OnStatusEffectTick(nil, "burning", 5.0)
	})

	t.Run("nil world", func(t *testing.T) {
		system := NewStatusEffectDamageParticleSystem(nil, 12345)
		// Skip SetParticleSystem to avoid Ebiten initialization

		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: 100, Y: 100})

		// Should not panic with nil world
		system.OnStatusEffectTick(entity, "burning", 5.0)
	})

	t.Run("entity without position", func(t *testing.T) {
		system := NewStatusEffectDamageParticleSystem(world, 12345)
		entity := world.CreateEntity()
		// No position component

		// Should not panic without position
		system.OnStatusEffectTick(entity, "burning", 5.0)
	})
}

func TestStatusEffectDamageParticleSystem_CallbackSignature(t *testing.T) {
	// Test that the callback can be used as StatusEffectTickCallback
	world := NewWorldWithLogger(logrus.New())
	system := NewStatusEffectDamageParticleSystem(world, 12345)

	// Verify callback can be assigned to StatusEffectTickCallback type
	var callback StatusEffectTickCallback = system.OnStatusEffectTick

	if callback == nil {
		t.Error("callback should not be nil")
	}

	// Create entity and test callback doesn't panic
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 200, Y: 200})

	// These should all complete without panic (particle system is nil, so no actual spawning)
	callback(entity, "burning", 5.0)
	callback(entity, "poisoned", 3.0)
	callback(entity, "regeneration", 8.0)
	callback(entity, "unknown", 10.0)
}

func TestStatusEffectDamageParticleSystem_EffectTypes(t *testing.T) {
	tests := []struct {
		name       string
		effectType string
		magnitude  float64
	}{
		{"burning low", "burning", 5.0},
		{"burning high", "burning", 15.0},
		{"burning extreme", "burning", 100.0},
		{"poisoned low", "poisoned", 3.0},
		{"poisoned high", "poisoned", 20.0},
		{"poisoned extreme", "poisoned", 100.0},
		{"regeneration low", "regeneration", 5.0},
		{"regeneration high", "regeneration", 25.0},
		{"regeneration extreme", "regeneration", 100.0},
		{"unknown type", "unknown", 10.0},
		{"empty type", "", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorldWithLogger(logrus.New())
			system := NewStatusEffectDamageParticleSystem(world, 12345)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			// Should not panic for any effect type (particle system is nil)
			system.OnStatusEffectTick(entity, tt.effectType, tt.magnitude)
		})
	}
}

func TestStatusEffectDamageParticleSystem_GenreSettings(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorldWithLogger(logrus.New())
			system := NewStatusEffectDamageParticleSystem(world, 12345)
			system.SetGenre(genre)

			if system.genreID != genre {
				t.Errorf("genre = %v, want %v", system.genreID, genre)
			}

			// Test all effect types work with each genre
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			system.OnStatusEffectTick(entity, "burning", 10.0)
			system.OnStatusEffectTick(entity, "poisoned", 10.0)
			system.OnStatusEffectTick(entity, "regeneration", 10.0)
		})
	}
}

func TestStatusEffectDamageParticleSystem_SpreadFactor(t *testing.T) {
	world := NewWorldWithLogger(logrus.New())
	system := NewStatusEffectDamageParticleSystem(world, 12345)

	// Default spread factor should be 40.0
	if system.spreadFactor != 40.0 {
		t.Errorf("spreadFactor = %v, want 40.0", system.spreadFactor)
	}
}

func TestStatusEffectDamageParticleSystem_Determinism(t *testing.T) {
	// Verify that same seed produces same RNG state
	system1 := NewStatusEffectDamageParticleSystem(nil, 12345)
	system2 := NewStatusEffectDamageParticleSystem(nil, 12345)

	// Generate some random values from each
	val1 := system1.rng.Float64()
	val2 := system2.rng.Float64()

	if val1 != val2 {
		t.Errorf("Same seed should produce same RNG values: got %v and %v", val1, val2)
	}
}
