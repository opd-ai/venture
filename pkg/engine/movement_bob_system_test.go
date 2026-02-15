package engine

import (
	"math"
	"testing"
)

func TestMovementBobComponent_Type(t *testing.T) {
	c := &MovementBobComponent{}
	if got := c.Type(); got != "movement_bob" {
		t.Errorf("Type() = %q, want %q", got, "movement_bob")
	}
}

func TestNewMovementBobSystem(t *testing.T) {
	world := NewWorld()
	sys := NewMovementBobSystem(world, 42)
	if sys == nil {
		t.Fatal("NewMovementBobSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestMovementBobSystem_SetGenre(t *testing.T) {
	tests := []struct {
		genre    string
		wantAmpl float64
		wantFreq float64
		wantDamp float64
	}{
		{"fantasy", 1.5, 1.2, 0.8},
		{"horror", 2.5, 0.7, 0.6},
		{"scifi", 0.8, 1.4, 0.9},
		{"cyberpunk", 1.8, 1.6, 0.85},
		{"postapoc", 2.2, 0.9, 0.7},
		{"unknown", 1.5, 1.2, 0.8}, // defaults to fantasy
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewMovementBobSystem(world, 42)
			sys.SetGenre(tt.genre)
			if sys.preset.Amplitude != tt.wantAmpl {
				t.Errorf("Amplitude = %v, want %v", sys.preset.Amplitude, tt.wantAmpl)
			}
			if sys.preset.Frequency != tt.wantFreq {
				t.Errorf("Frequency = %v, want %v", sys.preset.Frequency, tt.wantFreq)
			}
			if sys.preset.Damping != tt.wantDamp {
				t.Errorf("Damping = %v, want %v", sys.preset.Damping, tt.wantDamp)
			}
		})
	}
}

func TestMovementBobSystem_Update_MovingEntity(t *testing.T) {
	world := NewWorld()
	sys := NewMovementBobSystem(world, 99)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 80, VY: 60}) // speed = 100

	entities := []*Entity{entity}

	// Run several frames
	for i := 0; i < 10; i++ {
		sys.Update(entities, 1.0/60.0)
	}

	comp, ok := entity.GetComponent("movement_bob")
	if !ok {
		t.Fatal("expected movement_bob component to be created")
	}
	bob := comp.(*MovementBobComponent)
	if !bob.Active {
		t.Error("expected bob to be active while moving")
	}
	if bob.OffsetY == 0 {
		t.Error("expected non-zero OffsetY while moving")
	}
	if math.Abs(bob.OffsetY) > sys.preset.Amplitude+0.01 {
		t.Errorf("OffsetY %v exceeds amplitude %v", bob.OffsetY, sys.preset.Amplitude)
	}
}

func TestMovementBobSystem_Update_StationaryEntity(t *testing.T) {
	world := NewWorld()
	sys := NewMovementBobSystem(world, 99)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&MovementBobComponent{OffsetY: 2.0, Active: true, Phase: 1.0})

	entities := []*Entity{entity}

	// Run enough frames for damping to reach zero
	for i := 0; i < 120; i++ {
		sys.Update(entities, 1.0/60.0)
	}

	comp, _ := entity.GetComponent("movement_bob")
	bob := comp.(*MovementBobComponent)
	if bob.Active {
		t.Error("expected bob to be inactive when stationary")
	}
	if math.Abs(bob.OffsetY) > 0.05 {
		t.Errorf("OffsetY %v should have dampened to ~0", bob.OffsetY)
	}
}

func TestMovementBobSystem_Update_NoVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewMovementBobSystem(world, 99)

	entity := world.CreateEntity()
	// No velocity component — should be skipped
	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	if _, ok := entity.GetComponent("movement_bob"); ok {
		t.Error("should not create bob component for entity without velocity")
	}
}

func TestMovementBobSystem_PhaseBounds(t *testing.T) {
	world := NewWorld()
	sys := NewMovementBobSystem(world, 99)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 150, VY: 0})

	entities := []*Entity{entity}

	// Run many frames to cycle phase multiple times
	for i := 0; i < 600; i++ {
		sys.Update(entities, 1.0/60.0)
	}

	comp, _ := entity.GetComponent("movement_bob")
	bob := comp.(*MovementBobComponent)
	if bob.Phase < 0 || bob.Phase >= 2.0*math.Pi {
		t.Errorf("Phase %v out of [0, 2π) range", bob.Phase)
	}
}

func TestMovementBobSystem_GenreAmplitudeDifference(t *testing.T) {
	// Verify horror produces larger amplitude bob than sci-fi
	world := NewWorld()
	horrorSys := NewMovementBobSystem(world, 99)
	horrorSys.SetGenre("horror")

	scifiSys := NewMovementBobSystem(world, 99)
	scifiSys.SetGenre("scifi")

	eH := world.CreateEntity()
	eH.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	eS := world.CreateEntity()
	eS.AddComponent(&VelocityComponent{VX: 100, VY: 0})

	var maxH, maxS float64
	for i := 0; i < 120; i++ {
		horrorSys.Update([]*Entity{eH}, 1.0/60.0)
		scifiSys.Update([]*Entity{eS}, 1.0/60.0)
		bH, _ := eH.GetComponent("movement_bob")
		bS, _ := eS.GetComponent("movement_bob")
		aH := math.Abs(bH.(*MovementBobComponent).OffsetY)
		aS := math.Abs(bS.(*MovementBobComponent).OffsetY)
		if aH > maxH {
			maxH = aH
		}
		if aS > maxS {
			maxS = aS
		}
	}

	if maxH <= maxS {
		t.Errorf("horror max bob %v should exceed sci-fi max bob %v", maxH, maxS)
	}
}

func BenchmarkMovementBobSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewMovementBobSystem(world, 42)

	entities := make([]*Entity, 500)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&VelocityComponent{VX: float64(i % 200), VY: float64(i % 100)})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 1.0/60.0)
	}
}
