package engine

import (
	"testing"
)

func TestNewLowHealthVFXSystem(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewLowHealthVFXSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.criticalThreshold != 0.20 {
		t.Errorf("criticalThreshold = %f, want 0.20", sys.criticalThreshold)
	}
	if sys.warningThreshold != 0.35 {
		t.Errorf("warningThreshold = %f, want 0.35", sys.warningThreshold)
	}
}

func TestLowHealthVFXSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestLowHealthVFXSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", sys.genreID)
	}
}

func TestLowHealthVFXSystem_SetThresholds(t *testing.T) {
	tests := []struct {
		name           string
		warning        float64
		critical       float64
		wantWarning    float64
		wantCritical   float64
		expectWarning  bool
		expectCritical bool
	}{
		{
			name:           "valid thresholds",
			warning:        0.40,
			critical:       0.15,
			wantWarning:    0.40,
			wantCritical:   0.15,
			expectWarning:  true,
			expectCritical: true,
		},
		{
			name:           "critical >= warning ignored",
			warning:        0.30,
			critical:       0.50,
			wantWarning:    0.30,
			wantCritical:   0.20, // unchanged from default
			expectWarning:  true,
			expectCritical: false,
		},
		{
			name:           "invalid values ignored",
			warning:        -0.1,
			critical:       1.5,
			wantWarning:    0.35, // default
			wantCritical:   0.20, // default
			expectWarning:  false,
			expectCritical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewLowHealthVFXSystem(world, 12345)

			sys.SetThresholds(tt.warning, tt.critical)

			if sys.warningThreshold != tt.wantWarning {
				t.Errorf("warningThreshold = %f, want %f", sys.warningThreshold, tt.wantWarning)
			}
			if sys.criticalThreshold != tt.wantCritical {
				t.Errorf("criticalThreshold = %f, want %f", sys.criticalThreshold, tt.wantCritical)
			}
		})
	}
}

func TestLowHealthVFXSystem_isPlayerEntity(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)

	// Entity without input is not player
	nonPlayer := world.CreateEntity()
	nonPlayer.AddComponent(&HealthComponent{Current: 50, Max: 100})

	if sys.isPlayerEntity(nonPlayer) {
		t.Error("entity without input should not be player")
	}

	// Entity with input is player
	player := world.CreateEntity()
	player.AddComponent(&HealthComponent{Current: 50, Max: 100})
	player.AddComponent(NewStubInput())

	if !sys.isPlayerEntity(player) {
		t.Error("entity with input should be player")
	}
}

func TestLowHealthVFXSystem_getHealthRatio(t *testing.T) {
	tests := []struct {
		name      string
		current   float64
		max       float64
		wantRatio float64
	}{
		{"full health", 100, 100, 1.0},
		{"half health", 50, 100, 0.5},
		{"low health", 20, 100, 0.2},
		{"critical health", 10, 100, 0.1},
		{"no health component", 0, 0, -1}, // will be handled specially
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewLowHealthVFXSystem(world, 12345)

			entity := world.CreateEntity()
			if tt.max > 0 {
				entity.AddComponent(&HealthComponent{Current: tt.current, Max: tt.max})
			}

			ratio := sys.getHealthRatio(entity)
			if ratio != tt.wantRatio {
				t.Errorf("getHealthRatio() = %f, want %f", ratio, tt.wantRatio)
			}
		})
	}
}

func TestLowHealthVFXSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	// No particle system set

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 10, Max: 100})
	player.AddComponent(NewStubInput())

	// Should not panic
	sys.Update([]*Entity{player}, 1.0)
}

func TestLowHealthVFXSystem_Update_HealthyPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 80, Max: 100}) // 80% health
	player.AddComponent(NewStubInput())

	initialEntities := len(world.GetEntities())

	// Run update with enough time to trigger pulse
	sys.Update([]*Entity{player}, 1.0)

	// No particles should spawn for healthy player
	if len(world.GetEntities()) != initialEntities {
		t.Errorf("particles spawned for healthy player: %d new entities", len(world.GetEntities())-initialEntities)
	}
}

func TestLowHealthVFXSystem_Update_WarningHealth(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 30, Max: 100}) // 30% - warning level
	player.AddComponent(NewStubInput())

	world.FlushPendingEntities()
	initialEntities := len(world.GetEntities())

	// Run update with enough time to trigger pulse
	sys.Update([]*Entity{player}, 1.0)

	world.FlushPendingEntities()
	// Warning particles should spawn
	if len(world.GetEntities()) <= initialEntities {
		t.Error("warning particles should spawn for low health player")
	}
}

func TestLowHealthVFXSystem_Update_CriticalHealth(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("scifi")

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 15, Max: 100}) // 15% - critical level
	player.AddComponent(NewStubInput())

	world.FlushPendingEntities()
	initialEntities := len(world.GetEntities())

	// Run update with enough time to trigger pulse
	sys.Update([]*Entity{player}, 1.0)

	world.FlushPendingEntities()
	// Critical particles should spawn (2 effects: primary + secondary)
	newEntities := len(world.GetEntities()) - initialEntities
	if newEntities < 2 {
		t.Errorf("critical health should spawn at least 2 particle entities, got %d", newEntities)
	}
}

func TestLowHealthVFXSystem_Update_PulseTiming(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("horror")

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 10, Max: 100}) // critical
	player.AddComponent(NewStubInput())

	world.FlushPendingEntities()

	// First update with short delta - should not spawn yet
	sys.Update([]*Entity{player}, 0.1)
	world.FlushPendingEntities()
	firstCount := len(world.GetEntities())

	// Still not enough time
	sys.Update([]*Entity{player}, 0.1)
	world.FlushPendingEntities()
	if len(world.GetEntities()) != firstCount {
		t.Error("particles spawned before pulse interval")
	}

	// Now trigger with enough time
	sys.Update([]*Entity{player}, 0.7)
	world.FlushPendingEntities()
	if len(world.GetEntities()) == firstCount {
		t.Error("particles should spawn after pulse interval elapsed")
	}
}

func TestLowHealthVFXSystem_GenreParticleTypes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewLowHealthVFXSystem(world, 12345)
			sys.SetGenre(genre)

			warningType := sys.getWarningParticleType()
			criticalType := sys.getCriticalParticleType()

			// Just verify they return valid types (not panicking)
			if warningType.String() == "" {
				t.Errorf("warning particle type invalid for genre %s", genre)
			}
			if criticalType.String() == "" {
				t.Errorf("critical particle type invalid for genre %s", genre)
			}
		})
	}
}

func TestLowHealthVFXSystem_Update_NonPlayerIgnored(t *testing.T) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Entity without input (NPC)
	npc := world.CreateEntity()
	npc.AddComponent(&PositionComponent{X: 100, Y: 100})
	npc.AddComponent(&HealthComponent{Current: 5, Max: 100}) // very low health
	// No input component

	initialEntities := len(world.GetEntities())

	sys.Update([]*Entity{npc}, 1.0)

	// No particles should spawn for NPC
	if len(world.GetEntities()) != initialEntities {
		t.Error("particles should not spawn for non-player entities")
	}
}

func BenchmarkLowHealthVFXSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewLowHealthVFXSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 15, Max: 100})
	player.AddComponent(NewStubInput())

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceEmit = 0 // Reset to trigger each iteration
		sys.Update(entities, 1.0)
	}
}
