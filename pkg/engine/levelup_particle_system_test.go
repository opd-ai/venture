package engine

import (
	"testing"
)

func TestNewLevelUpParticleSystem(t *testing.T) {
	tests := []struct {
		name      string
		world     *World
		seed      int64
		expectNil bool
	}{
		{"valid world", NewWorld(), 12345, false},
		{"nil world", nil, 12345, false},
		{"zero seed", NewWorld(), 0, false},
		{"negative seed", NewWorld(), -999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewLevelUpParticleSystem(tt.world, tt.seed)
			if sys == nil {
				t.Fatal("NewLevelUpParticleSystem returned nil")
			}
			if sys.seed != tt.seed {
				t.Errorf("seed = %d, want %d", sys.seed, tt.seed)
			}
			if sys.baseParticleCount <= 0 {
				t.Error("baseParticleCount should be positive")
			}
			if sys.spreadFactor <= 0 {
				t.Error("spreadFactor should be positive")
			}
			if sys.rng == nil {
				t.Error("rng should not be nil")
			}
		})
	}
}

func TestLevelUpParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewLevelUpParticleSystem(world, 12345)

	if sys.particleSystem != nil {
		t.Error("particleSystem should be nil initially")
	}

	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particleSystem was not set correctly")
	}
}

func TestLevelUpParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewLevelUpParticleSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestLevelUpParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewLevelUpParticleSystem(world, 12345)

	// Update should be no-op (callback-driven system)
	entities := []*Entity{world.CreateEntity()}
	sys.Update(entities, 0.016)
	// No panic = success for callback-driven system
}

func TestLevelUpParticleSystem_OnLevelUp(t *testing.T) {
	tests := []struct {
		name           string
		setupSystem    func(*LevelUpParticleSystem)
		entity         *Entity
		level          int
		expectParticle bool
	}{
		{
			name: "valid level-up with all systems",
			setupSystem: func(s *LevelUpParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			level:          2,
			expectParticle: true,
		},
		{
			name:           "no particle system",
			setupSystem:    func(s *LevelUpParticleSystem) {},
			level:          2,
			expectParticle: false,
		},
		{
			name: "milestone level (5)",
			setupSystem: func(s *LevelUpParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			level:          5,
			expectParticle: true,
		},
		{
			name: "milestone level (10)",
			setupSystem: func(s *LevelUpParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			level:          10,
			expectParticle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewLevelUpParticleSystem(world, 12345)
			tt.setupSystem(sys)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			initialEntityCount := len(world.GetEntities())
			sys.OnLevelUp(entity, tt.level)

			// Check if particles were spawned (new entity created)
			newEntityCount := len(world.GetEntities())
			particlesSpawned := newEntityCount > initialEntityCount

			if tt.expectParticle && !particlesSpawned {
				t.Error("expected particles to be spawned but none were")
			}
			if !tt.expectParticle && particlesSpawned {
				t.Error("did not expect particles but some were spawned")
			}
		})
	}
}

func TestLevelUpParticleSystem_OnLevelUp_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewLevelUpParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	// Should not panic with nil entity
	sys.OnLevelUp(nil, 5)
}

func TestLevelUpParticleSystem_OnLevelUp_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewLevelUpParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	entity := world.CreateEntity()
	// No position component added

	initialCount := len(world.GetEntities())
	sys.OnLevelUp(entity, 5)

	// No particles should be spawned without position
	if len(world.GetEntities()) != initialCount {
		t.Error("particles should not spawn without position component")
	}
}

func TestLevelUpParticleSystem_SpawnLevelUpEffect(t *testing.T) {
	tests := []struct {
		name           string
		x, y           float64
		level          int
		setupSystem    func(*LevelUpParticleSystem)
		expectParticle bool
	}{
		{
			name: "valid spawn",
			x:    200, y: 200,
			level: 3,
			setupSystem: func(s *LevelUpParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			expectParticle: true,
		},
		{
			name: "no particle system",
			x:    200, y: 200,
			level:          3,
			setupSystem:    func(s *LevelUpParticleSystem) {},
			expectParticle: false,
		},
		{
			name: "negative coordinates",
			x:    -100, y: -100,
			level: 3,
			setupSystem: func(s *LevelUpParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			expectParticle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewLevelUpParticleSystem(world, 12345)
			tt.setupSystem(sys)

			initialCount := len(world.GetEntities())
			sys.SpawnLevelUpEffect(tt.x, tt.y, tt.level)

			particlesSpawned := len(world.GetEntities()) > initialCount

			if tt.expectParticle && !particlesSpawned {
				t.Error("expected particles to be spawned")
			}
			if !tt.expectParticle && particlesSpawned {
				t.Error("did not expect particles but some were spawned")
			}
		})
	}
}

func TestLevelUpParticleSystem_ParticleScaling(t *testing.T) {
	tests := []struct {
		name  string
		level int
	}{
		{"level 1", 1},
		{"level 5 milestone", 5},
		{"level 10 milestone", 10},
		{"level 15 milestone", 15},
		{"level 20 milestone", 20},
		{"level 50 milestone", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewLevelUpParticleSystem(world, 12345)
			sys.SetParticleSystem(NewParticleSystem())

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			// Should not panic at any level
			sys.OnLevelUp(entity, tt.level)
		})
	}
}

func TestLevelUpParticleSystem_DeterministicSeed(t *testing.T) {
	world1 := NewWorld()
	world2 := NewWorld()

	// Same seed should produce consistent behavior
	sys1 := NewLevelUpParticleSystem(world1, 12345)
	sys2 := NewLevelUpParticleSystem(world2, 12345)

	sys1.SetParticleSystem(NewParticleSystem())
	sys2.SetParticleSystem(NewParticleSystem())

	entity1 := world1.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})

	entity2 := world2.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys1.OnLevelUp(entity1, 5)
	sys2.OnLevelUp(entity2, 5)

	// Both should have created same number of entities
	if len(world1.GetEntities()) != len(world2.GetEntities()) {
		t.Error("deterministic seed should produce same entity count")
	}
}

func TestLevelUpParticleSystem_GenreAware(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewLevelUpParticleSystem(world, 12345)
			sys.SetParticleSystem(NewParticleSystem())
			sys.SetGenre(genre)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			// Should not panic with any genre
			sys.OnLevelUp(entity, 5)

			if len(world.GetEntities()) <= 1 {
				t.Error("particles should be spawned for genre: " + genre)
			}
		})
	}
}

func BenchmarkLevelUpParticleSystem_OnLevelUp(b *testing.B) {
	world := NewWorld()
	sys := NewLevelUpParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnLevelUp(entity, i%20+1)
	}
}

func BenchmarkLevelUpParticleSystem_SpawnLevelUpEffect(b *testing.B) {
	world := NewWorld()
	sys := NewLevelUpParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.SpawnLevelUpEffect(float64(i%500), float64(i%500), i%20+1)
	}
}
