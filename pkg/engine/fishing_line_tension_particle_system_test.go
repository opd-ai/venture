package engine

import (
	"testing"
)

func TestNewFishingLineTensionParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewFishingLineTensionParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genreID = %s, want fantasy", sys.genreID)
	}
	if sys.emitInterval <= 0 {
		t.Error("emitInterval should be positive")
	}
}

func TestFishingLineTensionParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestFishingLineTensionParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genre)
			}
		})
	}
}

func TestFishingLineTensionParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)

	// Should not panic with nil particle system
	sys.Update([]*Entity{}, 0.016)
}

func TestFishingLineTensionParticleSystem_Update_NoWorld(t *testing.T) {
	sys := NewFishingLineTensionParticleSystem(nil, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Should not panic with nil world
	sys.Update([]*Entity{}, 0.016)
}

func TestFishingLineTensionParticleSystem_Update_IdleFishing(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity with fishing component in idle state
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	fishComp := NewFishingComponent()
	entity.AddComponent(fishComp)

	// Should not spawn particles for idle state
	sys.Update([]*Entity{entity}, 0.016)

	// No crash means success
}

func TestFishingLineTensionParticleSystem_Update_ReelingState(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity with fishing component in reeling state
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	fishComp := NewFishingComponent()
	fishComp.State = FishingStateReeling
	fishComp.TensionLevel = 0.5
	fishComp.ReelProgress = 0.3
	fishComp.HookedFishTypeID = "bass"
	fishComp.HookedFishWeight = 5.0
	entity.AddComponent(fishComp)

	// Accumulate enough time to trigger particle emission
	for i := 0; i < 15; i++ {
		sys.Update([]*Entity{entity}, 0.016)
	}

	// Verify reeling state is tracked
	if len(sys.previouslyReeling) != 1 {
		t.Errorf("previouslyReeling count = %d, want 1", len(sys.previouslyReeling))
	}
}

func TestFishingLineTensionParticleSystem_Update_ReelingTransition(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	fishComp := NewFishingComponent()
	entity.AddComponent(fishComp)

	// Start idle
	sys.Update([]*Entity{entity}, 0.016)

	// Transition to reeling (should trigger start burst)
	fishComp.State = FishingStateReeling
	fishComp.TensionLevel = 0.5
	fishComp.HookedFishTypeID = "bass"
	sys.Update([]*Entity{entity}, 0.016)

	if !sys.previouslyReeling[entity.ID] {
		t.Error("Entity should be tracked as reeling")
	}

	// Transition out of reeling
	fishComp.State = FishingStateCaught
	sys.Update([]*Entity{entity}, 0.016)

	if sys.previouslyReeling[entity.ID] {
		t.Error("Entity should not be tracked after stopping reeling")
	}
}

func TestFishingLineTensionParticleSystem_selectTensionParticles(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)

	tests := []struct {
		name        string
		genre       string
		tension     float64
		struggleDir int
		wantCount   int // Minimum expected count
	}{
		{"fantasy_low", "fantasy", 0.2, 0, 3},
		{"fantasy_high", "fantasy", 0.8, 0, 5},
		{"fantasy_critical", "fantasy", 0.95, 0, 6},
		{"scifi_high", "scifi", 0.75, 1, 5},
		{"horror_critical", "horror", 0.92, -1, 6},
		{"cyberpunk_low", "cyberpunk", 0.25, 0, 3},
		{"postapoc_high", "postapoc", 0.8, 1, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			_, count, gravity := sys.selectTensionParticles(tt.tension, tt.struggleDir)

			if count < tt.wantCount {
				t.Errorf("count = %d, want >= %d", count, tt.wantCount)
			}
			// Gravity should be non-zero
			if gravity == 0 && tt.genre != "horror" {
				// Horror at high tension has 0 gravity which is valid
				if tt.tension < sys.highTensionThreshold {
					t.Errorf("gravity should be non-zero for %s tension %.2f", tt.genre, tt.tension)
				}
			}
		})
	}
}

func TestFishingLineTensionParticleSystem_selectDuration(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)

	tests := []struct {
		tension float64
		wantMin float64
		wantMax float64
	}{
		{0.0, 0.29, 0.31},
		{0.5, 0.44, 0.46},
		{1.0, 0.59, 0.61},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			duration := sys.selectDuration(tt.tension)
			if duration < tt.wantMin || duration > tt.wantMax {
				t.Errorf("duration(%.1f) = %.2f, want %.2f-%.2f", tt.tension, duration, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFishingLineTensionParticleSystem_HighTensionWarning(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	fishComp := NewFishingComponent()
	fishComp.State = FishingStateReeling
	fishComp.TensionLevel = 0.95 // Critical
	fishComp.HookedFishTypeID = "shark"
	fishComp.HookedFishWeight = 15.0
	entity.AddComponent(fishComp)

	// Accumulate time to emit
	sys.timeSinceEmit = sys.emitInterval
	sys.Update([]*Entity{entity}, 0.016)

	// Should have spawned warning particles (no crash = success)
}

func TestFishingLineTensionParticleSystem_StruggleSplash(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	fishComp := NewFishingComponent()
	fishComp.State = FishingStateReeling
	fishComp.TensionLevel = 0.6
	fishComp.FishStruggleDirection = 1 // Fish fighting right
	fishComp.HookedFishTypeID = "trout"
	fishComp.HookedFishWeight = 3.0
	entity.AddComponent(fishComp)

	sys.timeSinceEmit = sys.emitInterval
	sys.Update([]*Entity{entity}, 0.016)

	// Should spawn struggle splash (no crash = success)
}

func TestFishingLineTensionParticleSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entities := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 50), Y: 100})
		fishComp := NewFishingComponent()
		fishComp.State = FishingStateReeling
		fishComp.TensionLevel = float64(i+1) * 0.3
		fishComp.HookedFishTypeID = "fish"
		entity.AddComponent(fishComp)
		entities[i] = entity
	}

	sys.timeSinceEmit = sys.emitInterval
	sys.Update(entities, 0.016)

	if len(sys.previouslyReeling) != 3 {
		t.Errorf("previouslyReeling count = %d, want 3", len(sys.previouslyReeling))
	}
}

func TestFishingLineTensionParticleSystem_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	fishComp := NewFishingComponent()
	fishComp.State = FishingStateReeling
	entity.AddComponent(fishComp)
	// No position component

	sys.Update([]*Entity{entity}, 0.016)

	// Should not track entity without position
	if len(sys.previouslyReeling) != 0 {
		t.Error("Should not track entity without position")
	}
}

func BenchmarkFishingLineTensionParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 50), Y: 100})
		fishComp := NewFishingComponent()
		fishComp.State = FishingStateReeling
		fishComp.TensionLevel = 0.5
		fishComp.HookedFishTypeID = "fish"
		entity.AddComponent(fishComp)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkFishingLineTensionParticleSystem_selectTensionParticles(b *testing.B) {
	world := NewWorld()
	sys := NewFishingLineTensionParticleSystem(world, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.selectTensionParticles(0.7, 1)
	}
}
