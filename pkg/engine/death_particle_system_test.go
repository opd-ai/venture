package engine

import (
	"testing"
)

func TestNewDeathParticleSystem(t *testing.T) {
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
			sys := NewDeathParticleSystem(tt.world, tt.seed)
			if sys == nil {
				t.Fatal("NewDeathParticleSystem returned nil")
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

func TestDeathParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewDeathParticleSystem(world, 12345)

	if sys.particleSystem != nil {
		t.Error("particleSystem should be nil initially")
	}

	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particleSystem was not set correctly")
	}
}

func TestDeathParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDeathParticleSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestDeathParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewDeathParticleSystem(world, 12345)

	// Update should be no-op (callback-driven system)
	entities := []*Entity{world.CreateEntity()}
	sys.Update(entities, 0.016)
	// No panic = success for callback-driven system
}

func TestDeathParticleSystem_OnDeath(t *testing.T) {
	tests := []struct {
		name           string
		setupSystem    func(*DeathParticleSystem)
		entity         *Entity
		expectParticle bool
	}{
		{
			name: "valid death with all systems",
			setupSystem: func(s *DeathParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			expectParticle: true,
		},
		{
			name:           "no particle system",
			setupSystem:    func(s *DeathParticleSystem) {},
			expectParticle: false,
		},
		{
			name: "with genre fantasy",
			setupSystem: func(s *DeathParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
				s.SetGenre("fantasy")
			},
			expectParticle: true,
		},
		{
			name: "with genre horror",
			setupSystem: func(s *DeathParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
				s.SetGenre("horror")
			},
			expectParticle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDeathParticleSystem(world, 12345)
			tt.setupSystem(sys)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			initialEntityCount := len(world.GetEntities())
			sys.OnDeath(entity)

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

func TestDeathParticleSystem_OnDeath_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewDeathParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	// Should not panic with nil entity
	sys.OnDeath(nil)
}

func TestDeathParticleSystem_OnDeath_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewDeathParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	entity := world.CreateEntity()
	// No position component added

	initialCount := len(world.GetEntities())
	sys.OnDeath(entity)

	// No particles should be spawned without position
	if len(world.GetEntities()) != initialCount {
		t.Error("particles should not spawn without position component")
	}
}

func TestDeathParticleSystem_SpawnDeathEffect(t *testing.T) {
	tests := []struct {
		name           string
		x, y           float64
		setupSystem    func(*DeathParticleSystem)
		expectParticle bool
	}{
		{
			name: "valid spawn",
			x:    200, y: 200,
			setupSystem: func(s *DeathParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			expectParticle: true,
		},
		{
			name: "no particle system",
			x:    200, y: 200,
			setupSystem:    func(s *DeathParticleSystem) {},
			expectParticle: false,
		},
		{
			name: "negative coordinates",
			x:    -100, y: -100,
			setupSystem: func(s *DeathParticleSystem) {
				s.SetParticleSystem(NewParticleSystem())
			},
			expectParticle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDeathParticleSystem(world, 12345)
			tt.setupSystem(sys)

			initialCount := len(world.GetEntities())
			sys.SpawnDeathEffect(tt.x, tt.y)

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

func TestDeathParticleSystem_GenreAware(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewDeathParticleSystem(world, 12345)
			sys.SetParticleSystem(NewParticleSystem())
			sys.SetGenre(genre)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			// Should not panic with any genre
			sys.OnDeath(entity)

			if len(world.GetEntities()) <= 1 {
				t.Error("particles should be spawned for genre: " + genre)
			}
		})
	}
}

func TestDeathParticleSystem_getPrimaryParticleType(t *testing.T) {
	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "smoke"},
		{"scifi", "spark"},
		{"horror", "blood"},
		{"cyberpunk", "ember"},
		{"postapoc", "debris"},
		{"unknown", "smoke_plume"},
		{"", "smoke_plume"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewDeathParticleSystem(world, 12345)
			sys.SetGenre(tt.genre)

			pType := sys.getPrimaryParticleType()
			if pType.String() != tt.expected {
				t.Errorf("getPrimaryParticleType() = %s, want %s", pType.String(), tt.expected)
			}
		})
	}
}

func TestDeathParticleSystem_getSecondaryParticleType(t *testing.T) {
	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "dust"},
		{"scifi", "smoke"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "dust"},
		{"unknown", "dust"},
		{"", "dust"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewDeathParticleSystem(world, 12345)
			sys.SetGenre(tt.genre)

			pType := sys.getSecondaryParticleType()
			if pType.String() != tt.expected {
				t.Errorf("getSecondaryParticleType() = %s, want %s", pType.String(), tt.expected)
			}
		})
	}
}

func TestDeathParticleSystem_DeterministicSeed(t *testing.T) {
	world1 := NewWorld()
	world2 := NewWorld()

	// Same seed should produce consistent behavior
	sys1 := NewDeathParticleSystem(world1, 12345)
	sys2 := NewDeathParticleSystem(world2, 12345)

	sys1.SetParticleSystem(NewParticleSystem())
	sys2.SetParticleSystem(NewParticleSystem())

	entity1 := world1.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})

	entity2 := world2.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys1.OnDeath(entity1)
	sys2.OnDeath(entity2)

	// Both should have created same number of entities
	if len(world1.GetEntities()) != len(world2.GetEntities()) {
		t.Error("deterministic seed should produce same entity count")
	}
}
