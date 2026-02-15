package engine

import (
	"math"
	"testing"
)

func TestMovementLeanComponentType(t *testing.T) {
	c := &MovementLeanComponent{}
	if got := c.Type(); got != "movement_lean" {
		t.Errorf("Type() = %q, want %q", got, "movement_lean")
	}
}

func TestNewMovementLeanSystem(t *testing.T) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)
	if sys == nil {
		t.Fatal("NewMovementLeanSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestMovementLeanSystemSetGenre(t *testing.T) {
	genres := []struct {
		name    string
		maxLean float64
	}{
		{"fantasy", 2.0},
		{"horror", 3.0},
		{"scifi", 1.5},
		{"cyberpunk", 2.5},
		{"postapoc", 3.5},
		{"unknown", 2.0},
	}
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)
	for _, tt := range genres {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.name)
			if sys.preset.MaxLean != tt.maxLean {
				t.Errorf("genre %q MaxLean = %v, want %v", tt.name, sys.preset.MaxLean, tt.maxLean)
			}
		})
	}
}

func TestMovementLeanSystemUpdate(t *testing.T) {
	tests := []struct {
		name       string
		velX, velY float64
		wantActive bool
		wantSign   int // -1, 0, or 1 for OffsetX sign
	}{
		{"stationary", 0, 0, false, 0},
		{"moving right", 100, 0, true, 1},
		{"moving left", -100, 0, true, -1},
		{"moving up only", 0, -100, true, 0},
		{"moving down only", 0, 100, true, 0},
		{"diagonal right-down", 80, 80, true, 1},
		{"diagonal left-up", -80, -80, true, -1},
		{"slow below threshold", 5, 5, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewMovementLeanSystem(world, 42)

			entity := NewEntity(0)
			entity.AddComponent(&VelocityComponent{VX: tt.velX, VY: tt.velY})
			entities := []*Entity{entity}

			// Run several frames to let lean converge
			for i := 0; i < 30; i++ {
				sys.Update(entities, 0.016)
			}

			lean := getLeanComponent(entity)
			if lean == nil {
				t.Fatal("entity missing MovementLeanComponent after Update")
			}

			if tt.wantActive && !lean.Active {
				t.Error("expected Active=true")
			}
			if !tt.wantActive && lean.Active {
				t.Error("expected Active=false")
			}

			switch tt.wantSign {
			case 1:
				if lean.OffsetX <= 0 {
					t.Errorf("expected positive OffsetX, got %v", lean.OffsetX)
				}
			case -1:
				if lean.OffsetX >= 0 {
					t.Errorf("expected negative OffsetX, got %v", lean.OffsetX)
				}
			case 0:
				if math.Abs(lean.OffsetX) > 0.1 {
					t.Errorf("expected near-zero OffsetX, got %v", lean.OffsetX)
				}
			}
		})
	}
}

func TestMovementLeanSystemDamping(t *testing.T) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(&VelocityComponent{VX: 150, VY: 0})
	entities := []*Entity{entity}

	// Build up lean
	for i := 0; i < 30; i++ {
		sys.Update(entities, 0.016)
	}

	lean := getLeanComponent(entity)
	if lean == nil || lean.OffsetX <= 0 {
		t.Fatal("lean not built up")
	}
	builtLean := lean.OffsetX

	// Stop movement
	vel := entity.GetVelocity()
	vel.VX = 0
	vel.VY = 0

	// Damp down
	for i := 0; i < 60; i++ {
		sys.Update(entities, 0.016)
	}

	if lean.OffsetX >= builtLean {
		t.Errorf("lean should decrease after stopping, got %v >= %v", lean.OffsetX, builtLean)
	}
	if lean.Active {
		t.Error("lean should be inactive after sufficient damping")
	}
}

func TestMovementLeanSystemZeroDeltaTime(t *testing.T) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	entities := []*Entity{entity}

	// Zero delta should be a no-op
	sys.Update(entities, 0)
	lean := getLeanComponent(entity)
	if lean != nil {
		t.Error("should not create component with zero delta")
	}
}

func TestMovementLeanSystemLargeDeltaClamped(t *testing.T) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	entities := []*Entity{entity}

	// Large delta should be clamped; should not overshoot MaxLean
	sys.Update(entities, 5.0)
	lean := getLeanComponent(entity)
	if lean == nil {
		t.Fatal("missing lean component")
	}
	if math.Abs(lean.OffsetX) > sys.preset.MaxLean+0.01 {
		t.Errorf("OffsetX %v exceeds MaxLean %v", lean.OffsetX, sys.preset.MaxLean)
	}
}

func TestMovementLeanSystemMaxLeanBound(t *testing.T) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)

	entity := NewEntity(0)
	// Very high speed to maximize lean
	entity.AddComponent(&VelocityComponent{VX: 1000, VY: 0})
	entities := []*Entity{entity}

	for i := 0; i < 100; i++ {
		sys.Update(entities, 0.016)
	}

	lean := getLeanComponent(entity)
	if lean == nil {
		t.Fatal("missing lean component")
	}
	if math.Abs(lean.OffsetX) > sys.preset.MaxLean+0.01 {
		t.Errorf("OffsetX %v exceeds MaxLean %v", lean.OffsetX, sys.preset.MaxLean)
	}
}

func TestMovementLeanSystemNoVelocitySkipped(t *testing.T) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)

	entity := NewEntity(0) // No velocity component
	entities := []*Entity{entity}

	sys.Update(entities, 0.016)
	// Should not crash and should not add lean to entities without velocity
	if _, ok := entity.GetComponent("movement_lean"); ok {
		t.Error("should not add lean to entity without velocity")
	}
}

func BenchmarkMovementLeanSystem(b *testing.B) {
	world := NewWorld()
	sys := NewMovementLeanSystem(world, 42)

	entities := make([]*Entity, 500)
	for i := range entities {
		e := NewEntity(0)
		e.AddComponent(&VelocityComponent{VX: float64(i%200 - 100), VY: float64(i%100 - 50)})
		e.AddComponent(&MovementLeanComponent{})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func getLeanComponent(entity *Entity) *MovementLeanComponent {
	comp, ok := entity.GetComponent("movement_lean")
	if !ok {
		return nil
	}
	lean, _ := comp.(*MovementLeanComponent)
	return lean
}
