package engine

import (
	"math"
	"testing"
)

func TestShieldBubbleOverlayComponentType(t *testing.T) {
	comp := &ShieldBubbleOverlayComponent{}
	if comp.Type() != "shield_bubble_overlay" {
		t.Errorf("expected shield_bubble_overlay, got %s", comp.Type())
	}
}

func TestNewShieldBubbleOverlaySystem(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre fantasy, got %s", sys.genreID)
	}
}

func TestShieldBubbleOverlaySystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		wantR   float64
	}{
		{"fantasy", "fantasy", 0.9},
		{"horror", "horror", 0.8},
		{"scifi", "scifi", 0.2},
		{"cyberpunk", "cyberpunk", 0.1},
		{"postapoc", "postapoc", 0.75},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewShieldBubbleOverlaySystem(world, 42)
			sys.SetGenre(tt.genreID)
			if sys.preset.R != tt.wantR {
				t.Errorf("genre %s: expected R=%.2f, got %.2f", tt.genreID, tt.wantR, sys.preset.R)
			}
		})
	}
}

func TestShieldBubbleOverlaySystem_UpdateActivatesOnShield(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Width: 32, Height: 32, Visible: true})
	entity.AddComponent(&ShieldComponent{Amount: 50, MaxAmount: 100, Duration: 5, MaxDuration: 10})

	// Force immediate check by exceeding interval
	sys.timeSinceCheck = sys.checkInterval + 0.01
	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("shield_bubble_overlay")
	if !ok {
		t.Fatal("expected shield_bubble_overlay component to be created")
	}
	bc := comp.(*ShieldBubbleOverlayComponent)
	if !bc.Visible {
		t.Error("expected bubble to be visible")
	}
	if bc.Opacity <= 0 {
		t.Error("expected positive opacity")
	}
	if bc.Radius <= 0 {
		t.Error("expected positive radius")
	}
}

func TestShieldBubbleOverlaySystem_SkipsWithoutShield(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Width: 32, Height: 32, Visible: true})

	sys.timeSinceCheck = sys.checkInterval + 0.01
	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("shield_bubble_overlay")
	if ok {
		t.Error("entity without shield should not get overlay component")
	}
}

func TestShieldBubbleOverlaySystem_SkipsWithoutPosition(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&StubSprite{Width: 32, Height: 32, Visible: true})
	entity.AddComponent(&ShieldComponent{Amount: 50, MaxAmount: 100, Duration: 5, MaxDuration: 10})

	sys.timeSinceCheck = sys.checkInterval + 0.01
	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("shield_bubble_overlay")
	if ok {
		t.Error("entity without position should not get overlay component")
	}
}

func TestShieldBubbleOverlaySystem_OpacityScalesWithShieldFraction(t *testing.T) {
	tests := []struct {
		name      string
		amount    float64
		maxAmount float64
		wantHigh  bool
	}{
		{"full shield", 100, 100, true},
		{"half shield", 50, 100, false},
		{"low shield", 10, 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewShieldBubbleOverlaySystem(world, 42)
			bc := &ShieldBubbleOverlayComponent{}
			shield := &ShieldComponent{Amount: tt.amount, MaxAmount: tt.maxAmount, Duration: 5, MaxDuration: 10}
			sys.updateBubbleFromShield(bc, shield)

			expectedOpacity := sys.preset.BaseAlpha * (tt.amount / tt.maxAmount)
			if math.Abs(bc.Opacity-expectedOpacity) > 0.01 {
				t.Errorf("expected opacity ~%.3f, got %.3f", expectedOpacity, bc.Opacity)
			}
		})
	}
}

func TestShieldBubbleOverlaySystem_FlickerAtLowShield(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)
	bc := &ShieldBubbleOverlayComponent{}

	// Shield at 10% — should trigger flicker
	shield := &ShieldComponent{Amount: 10, MaxAmount: 100, Duration: 5, MaxDuration: 10}
	sys.updateBubbleFromShield(bc, shield)

	if bc.FlickerIntensity <= 0 {
		t.Error("expected flicker at low shield")
	}

	// Shield at 50% — no flicker
	shield2 := &ShieldComponent{Amount: 50, MaxAmount: 100, Duration: 5, MaxDuration: 10}
	sys.updateBubbleFromShield(bc, shield2)
	if bc.FlickerIntensity != 0 {
		t.Error("expected no flicker at half shield")
	}
}

func TestShieldBubbleOverlaySystem_FadeOutOnExpiry(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Width: 32, Height: 32, Visible: true})

	// Manually add overlay as if shield was previously active
	bc := &ShieldBubbleOverlayComponent{
		Visible: true,
		Opacity: 0.4,
		FadeMax: 0.5,
	}
	entity.AddComponent(bc)

	// No ShieldComponent — system should fade out
	sys.timeSinceCheck = sys.checkInterval + 0.01
	sys.Update([]*Entity{entity}, 0.016)

	if bc.FadeTimer <= 0 {
		t.Error("expected fade timer to be started")
	}

	// Run enough updates to complete fade
	for i := 0; i < 100; i++ {
		sys.Update([]*Entity{entity}, 0.016)
	}

	if bc.Visible {
		t.Error("expected bubble to be invisible after fade completes")
	}
	if bc.Opacity != 0 {
		t.Errorf("expected opacity 0 after fade, got %.3f", bc.Opacity)
	}
}

func TestShieldBubbleOverlaySystem_PulseAdvances(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	bc := &ShieldBubbleOverlayComponent{
		PulseSpeed: 2.5,
		PulsePhase: 0,
		Opacity:    0.4,
		Visible:    true,
	}

	initialPhase := bc.PulsePhase
	sys.animateBubble(bc, 0.1)

	if bc.PulsePhase <= initialPhase {
		t.Error("expected pulse phase to advance")
	}
}

func TestShieldBubbleOverlaySystem_NilEntitySafe(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	// Should not panic
	sys.Update([]*Entity{nil}, 0.016)
}

func TestShieldBubbleOverlaySystem_InactiveShieldNoActivation(t *testing.T) {
	world := NewWorld()
	sys := NewShieldBubbleOverlaySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Width: 32, Height: 32, Visible: true})
	// Shield with 0 duration = inactive
	entity.AddComponent(&ShieldComponent{Amount: 50, MaxAmount: 100, Duration: 0, MaxDuration: 10})

	sys.timeSinceCheck = sys.checkInterval + 0.01
	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("shield_bubble_overlay")
	if ok {
		t.Error("inactive shield should not create overlay")
	}
}

func TestClampShieldBubble(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"negative", -0.5, 0.0},
		{"zero", 0.0, 0.0},
		{"mid", 0.5, 0.5},
		{"one", 1.0, 1.0},
		{"over", 1.5, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampShieldBubble(tt.in)
			if got != tt.want {
				t.Errorf("clampShieldBubble(%f) = %f, want %f", tt.in, got, tt.want)
			}
		})
	}
}
