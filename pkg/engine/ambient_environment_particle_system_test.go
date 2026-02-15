package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewAmbientEnvironmentParticleSystem(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		world *World
	}{
		{"with world", 12345, NewWorld()},
		{"nil world", 99999, nil},
		{"zero seed", 0, NewWorld()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewAmbientEnvironmentParticleSystem(tt.world, tt.seed)
			if sys == nil {
				t.Fatal("expected non-nil system")
			}
			if sys.seed != tt.seed {
				t.Errorf("seed = %d, want %d", sys.seed, tt.seed)
			}
			if sys.spawnInterval <= 0 {
				t.Error("expected positive spawn interval")
			}
			if sys.spawnRadius <= 0 {
				t.Error("expected positive spawn radius")
			}
		})
	}
}

func TestAmbientEnvironmentParticleSystem_SetTileSize(t *testing.T) {
	sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"valid size", 64, 64},
		{"zero ignored", 0, 64},
		{"negative ignored", -1, 64},
		{"small valid", 16, 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetTileSize(tt.size)
			if sys.tileSize != tt.expected {
				t.Errorf("tileSize = %d, want %d", sys.tileSize, tt.expected)
			}
		})
	}
}

func TestAmbientEnvironmentParticleSystem_SetGenre(t *testing.T) {
	sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("genreID = %q, want %q", sys.genreID, g)
		}
	}
}

func TestAmbientEnvironmentParticleSystem_UpdateNilGuards(t *testing.T) {
	sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
	entities := []*Entity{NewEntity(1)}

	// Should not panic with nil particle system / terrain
	sys.Update(entities, 0.016)

	sys.SetParticleSystem(&ParticleSystem{})
	sys.Update(entities, 0.016) // still nil terrain

	sys.SetTerrain(terrain.NewTerrain(10, 10, 42))
	// Now all set but entity has no position - should not panic
	sys.Update(entities, 0.016)
}

// TestAmbientEnvironmentParticleSystem_DirectConfigs tests the individual config
// generators which are deterministic (no probability gating).
func TestAmbientEnvironmentParticleSystem_DirectConfigs(t *testing.T) {
	tests := []struct {
		name     string
		genre    string
		getFn    func(*AmbientEnvironmentParticleSystem) *particles.Config
		wantType particles.ParticleType
		wantZL   particles.ZLayer
	}{
		{"lava ember", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getLavaEmberConfig(100) }, particles.ParticleEmber, particles.ZLayerSky},
		{"deep water mist", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getDeepWaterMistConfig(100) }, particles.ParticleSmokePlume, particles.ZLayerGround},
		{"shallow water mist", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getMistConfig(100) }, particles.ParticleSmoke, particles.ZLayerGround},
		{"shallow water horror", "horror", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getMistConfig(100) }, particles.ParticleSmokePlume, particles.ZLayerGround},
		{"forest fantasy", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getForestAmbientConfig(100) }, particles.ParticleSparkle, particles.ZLayerAbove},
		{"forest horror", "horror", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getForestAmbientConfig(100) }, particles.ParticleSmoke, particles.ZLayerAbove},
		{"forest scifi", "scifi", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getForestAmbientConfig(100) }, particles.ParticleSpark, particles.ZLayerAbove},
		{"forest cyberpunk", "cyberpunk", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getForestAmbientConfig(100) }, particles.ParticleSpark, particles.ZLayerAbove},
		{"bridge mist", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getBridgeMistConfig(100) }, particles.ParticleSmoke, particles.ZLayerGround},
		{"structure dust fantasy", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getStructureDustConfig(100) }, particles.ParticleDust, particles.ZLayerGround},
		{"structure dust cyberpunk", "cyberpunk", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getStructureDustConfig(100) }, particles.ParticleSpark, particles.ZLayerGround},
		{"dust mote fantasy", "fantasy", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getDustMoteConfig(100) }, particles.ParticleDust, particles.ZLayerAbove},
		{"dust mote scifi", "scifi", func(s *AmbientEnvironmentParticleSystem) *particles.Config { return s.getDustMoteConfig(100) }, particles.ParticleSpark, particles.ZLayerAbove},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
			sys.SetGenre(tt.genre)
			config := tt.getFn(sys)
			if config == nil {
				t.Fatal("expected non-nil config")
			}
			if config.Type != tt.wantType {
				t.Errorf("particle type = %v, want %v", config.Type, tt.wantType)
			}
			if config.ZLayer != tt.wantZL {
				t.Errorf("ZLayer = %v, want %v", config.ZLayer, tt.wantZL)
			}
			if config.GenreID != tt.genre {
				t.Errorf("genre = %q, want %q", config.GenreID, tt.genre)
			}
			if config.Duration <= 0 {
				t.Error("expected positive duration")
			}
			if config.Count <= 0 {
				t.Error("expected positive count")
			}
		})
	}
}

func TestAmbientEnvironmentParticleSystem_GetAmbientConfigRouting(t *testing.T) {
	// Test routing: non-probabilistic tile types always return config
	tests := []struct {
		name     string
		tileType terrain.TileType
		wantNil  bool
	}{
		{"lava always spawns", terrain.TileLavaFlow, false},
		{"deep water always spawns", terrain.TileWaterDeep, false},
		{"shallow water always spawns", terrain.TileWaterShallow, false},
		{"tree always spawns", terrain.TileTree, false},
		{"wall never spawns", terrain.TileWall, true},
		{"door never spawns", terrain.TileDoor, true},
		{"stairs never spawn", terrain.TileStairsUp, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
			sys.SetGenre("fantasy")
			config := sys.getAmbientConfig(tt.tileType, 100, 100)
			if tt.wantNil && config != nil {
				t.Errorf("expected nil config for tile %v", tt.tileType)
			}
			if !tt.wantNil && config == nil {
				t.Errorf("expected non-nil config for tile %v", tt.tileType)
			}
		})
	}
}

func TestAmbientEnvironmentParticleSystem_CooldownTracking(t *testing.T) {
	world := NewWorld()
	sys := NewAmbientEnvironmentParticleSystem(world, 42)
	sys.SetTerrain(terrain.NewTerrain(10, 10, 42))
	sys.SetParticleSystem(&ParticleSystem{})

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entities := []*Entity{entity}

	// First update should set lastSpawn
	sys.Update(entities, 0.016)
	if _, ok := sys.lastSpawn[entity.ID]; !ok {
		t.Error("expected entity to be tracked in lastSpawn")
	}

	// Immediate second update should be skipped (cooldown)
	initialElapsed := sys.elapsed
	sys.Update(entities, 0.016)
	if sys.elapsed <= initialElapsed {
		t.Error("elapsed should increase")
	}
}

func TestAmbientEnvironmentParticleSystem_DustMoteGenreVariation(t *testing.T) {
	tests := []struct {
		genre    string
		wantType particles.ParticleType
	}{
		{"fantasy", particles.ParticleDust},
		{"horror", particles.ParticleDust},
		{"scifi", particles.ParticleSpark},
		{"cyberpunk", particles.ParticleSpark},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
			sys.SetGenre(tt.genre)
			config := sys.getDustMoteConfig(100)
			if config.Type != tt.wantType {
				t.Errorf("dust mote type = %v, want %v for genre %s", config.Type, tt.wantType, tt.genre)
			}
		})
	}
}

func TestAmbientEnvironmentParticleSystem_LavaConfig(t *testing.T) {
	sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
	sys.SetGenre("fantasy")
	config := sys.getLavaEmberConfig(100)
	if config.Type != particles.ParticleEmber {
		t.Errorf("expected ParticleEmber, got %v", config.Type)
	}
	if config.Gravity >= 0 {
		t.Error("lava embers should rise (negative gravity)")
	}
	if config.ZLayer != particles.ZLayerSky {
		t.Error("lava embers should use ZLayerSky")
	}
}

func TestAmbientEnvironmentParticleSystem_BridgeMist(t *testing.T) {
	sys := NewAmbientEnvironmentParticleSystem(NewWorld(), 42)
	config := sys.getBridgeMistConfig(100)
	if config.Type != particles.ParticleSmoke {
		t.Errorf("expected ParticleSmoke, got %v", config.Type)
	}
	if config.Gravity >= 0 {
		t.Error("bridge mist should drift upward")
	}
}

func TestAmbientEnvironmentParticleSystem_Deterministic(t *testing.T) {
	// Same seed should produce same rng sequence
	sys1 := NewAmbientEnvironmentParticleSystem(NewWorld(), 777)
	sys2 := NewAmbientEnvironmentParticleSystem(NewWorld(), 777)
	sys1.SetGenre("fantasy")
	sys2.SetGenre("fantasy")

	// Reset both to same seed state
	sys1.rng = rand.New(rand.NewSource(777))
	sys2.rng = rand.New(rand.NewSource(777))

	for i := 0; i < 10; i++ {
		v1 := sys1.rng.Float64()
		v2 := sys2.rng.Float64()
		if v1 != v2 {
			t.Fatalf("rng diverged at iteration %d: %f != %f", i, v1, v2)
		}
	}
}
