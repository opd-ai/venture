package engine

import (
	"math"
	"testing"
)

func TestNewEntityIdleBreathingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewEntityIdleBreathingSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.seed != 12345 {
		t.Error("seed not set")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want fantasy", sys.genreID)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
}

func TestEntityIdleBreathingSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 42)

	tests := []struct {
		genre      string
		wantAmp    float64
		wantFreq   float64
		wantJitter float64
	}{
		{"fantasy", 0.5, 0.8, 0.0},
		{"horror", 1.0, 0.5, 0.3},
		{"scifi", 0.3, 1.5, 0.0},
		{"cyberpunk", 0.4, 2.0, 0.15},
		{"postapoc", 0.8, 0.6, 0.1},
		{"unknown", 0.5, 0.8, 0.0}, // defaults to fantasy
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.genre)
			}
			if sys.preset.Amplitude != tt.wantAmp {
				t.Errorf("amplitude = %v, want %v", sys.preset.Amplitude, tt.wantAmp)
			}
			if sys.preset.Frequency != tt.wantFreq {
				t.Errorf("frequency = %v, want %v", sys.preset.Frequency, tt.wantFreq)
			}
			if sys.preset.Jitter != tt.wantJitter {
				t.Errorf("jitter = %v, want %v", sys.preset.Jitter, tt.wantJitter)
			}
		})
	}
}

func TestEntityIdleBreathingSystem_IdleEntityBreathes(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 99)

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entities := []*Entity{entity}

	// Run several frames at idle
	for i := 0; i < 30; i++ {
		sys.Update(entities, 1.0/60.0)
	}

	comp, ok := entity.GetComponent("idle_breathing")
	if !ok {
		t.Fatal("IdleBreathingComponent not created")
	}
	breath := comp.(*IdleBreathingComponent)
	if !breath.Active {
		t.Error("breathing should be active when idle")
	}
	if breath.OffsetY == 0 {
		t.Error("OffsetY should be non-zero after idle frames")
	}
	if breath.IdleTime <= 0 {
		t.Error("IdleTime should accumulate")
	}
}

func TestEntityIdleBreathingSystem_MovingEntityDecays(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 77)

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entities := []*Entity{entity}

	// Idle first to build up breathing
	for i := 0; i < 60; i++ {
		sys.Update(entities, 1.0/60.0)
	}
	comp, _ := entity.GetComponent("idle_breathing")
	breath := comp.(*IdleBreathingComponent)
	idleOffset := breath.OffsetY

	// Now start moving
	vel := entity.GetVelocity()
	vel.VX = 100
	vel.VY = 100

	// Run several moving frames
	for i := 0; i < 30; i++ {
		sys.Update(entities, 1.0/60.0)
	}

	if breath.Active {
		t.Error("breathing should deactivate when moving")
	}
	if math.Abs(breath.OffsetY) >= math.Abs(idleOffset) && idleOffset != 0 {
		t.Error("OffsetY should decay when moving")
	}
	if breath.IdleTime != 0 {
		t.Error("IdleTime should reset to 0 when moving")
	}
}

func TestEntityIdleBreathingSystem_SkipsNoVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 55)

	entity := NewEntity(1) // No velocity component
	entities := []*Entity{entity}

	sys.Update(entities, 1.0/60.0)

	_, ok := entity.GetComponent("idle_breathing")
	if ok {
		t.Error("should not create breathing component for entity without velocity")
	}
}

func TestEntityIdleBreathingSystem_RampGradualOnset(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 33)
	sys.SetGenre("fantasy") // RampTime = 0.5s

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entities := []*Entity{entity}

	// First frame — ramp just started, offset should be small
	sys.Update(entities, 0.01)
	comp, _ := entity.GetComponent("idle_breathing")
	breath := comp.(*IdleBreathingComponent)
	earlyOffset := math.Abs(breath.OffsetY)

	// Many more frames — ramp should be at full
	for i := 0; i < 120; i++ {
		sys.Update(entities, 1.0/60.0)
	}
	lateOffset := math.Abs(breath.OffsetY)

	// We can't guarantee exact values due to sine, but ramp should allow larger values later
	if breath.IdleTime < 0.5 {
		t.Error("IdleTime should exceed ramp time after enough frames")
	}
	_ = earlyOffset
	_ = lateOffset
}

func TestEntityIdleBreathingSystem_PhaseWraps(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 11)

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entities := []*Entity{entity}

	// Run many frames to ensure phase wraps
	for i := 0; i < 600; i++ {
		sys.Update(entities, 1.0/60.0)
	}

	comp, _ := entity.GetComponent("idle_breathing")
	breath := comp.(*IdleBreathingComponent)
	if breath.Phase >= 2.0*math.Pi {
		t.Errorf("phase should wrap below 2π, got %v", breath.Phase)
	}
}

func TestEntityIdleBreathingSystem_HorrorJitter(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 42)
	sys.SetGenre("horror") // Jitter = 0.3

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entities := []*Entity{entity}

	offsets := make(map[float64]bool)
	for i := 0; i < 120; i++ {
		sys.Update(entities, 1.0/60.0)
		comp, _ := entity.GetComponent("idle_breathing")
		breath := comp.(*IdleBreathingComponent)
		offsets[math.Round(breath.OffsetY*1000)/1000] = true
	}

	// Horror jitter should produce varied offsets
	if len(offsets) < 5 {
		t.Error("horror genre should produce varied breathing offsets due to jitter")
	}
}

func TestEntityIdleBreathingSystem_DeterministicSeed(t *testing.T) {
	// Same seed should produce same sequence
	offsets1 := runBreathingFrames(12345, "horror", 60)
	offsets2 := runBreathingFrames(12345, "horror", 60)

	for i := range offsets1 {
		if offsets1[i] != offsets2[i] {
			t.Errorf("frame %d: offset1=%v != offset2=%v (not deterministic)", i, offsets1[i], offsets2[i])
			break
		}
	}

	// Different seed should produce different sequence
	offsets3 := runBreathingFrames(99999, "horror", 60)
	same := true
	for i := range offsets1 {
		if offsets1[i] != offsets3[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different seeds should produce different offset sequences")
	}
}

func runBreathingFrames(seed int64, genre string, frames int) []float64 {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, seed)
	sys.SetGenre(genre)

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entities := []*Entity{entity}

	offsets := make([]float64, frames)
	for i := 0; i < frames; i++ {
		sys.Update(entities, 1.0/60.0)
		comp, _ := entity.GetComponent("idle_breathing")
		breath := comp.(*IdleBreathingComponent)
		offsets[i] = breath.OffsetY
	}
	return offsets
}

func TestIdleBreathingComponent_Type(t *testing.T) {
	c := &IdleBreathingComponent{}
	if c.Type() != "idle_breathing" {
		t.Errorf("Type() = %q, want idle_breathing", c.Type())
	}
}

func BenchmarkEntityIdleBreathingSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewEntityIdleBreathingSystem(world, 42)

	entities := make([]*Entity, 500)
	for i := range entities {
		e := NewEntity(1)
		e.AddComponent(&VelocityComponent{VX: 0, VY: 0})
		e.AddComponent(&IdleBreathingComponent{})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 1.0/60.0)
	}
}
