package engine

import (
	"math"
	"testing"
)

// stubInputComponent is a test-only component that registers as "input".
type stubFactionOutlineInputComponent struct{}

func (s *stubFactionOutlineInputComponent) Type() string { return "input" }

func TestEntityFactionOutlineComponentType(t *testing.T) {
	c := &EntityFactionOutlineComponent{}
	if c.Type() != "entity_faction_outline" {
		t.Errorf("expected type 'entity_faction_outline', got %q", c.Type())
	}
}

func TestNewEntityFactionOutlineSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestEntityFactionOutlineSystemSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"horror"},
		{"scifi"},
		{"cyberpunk"},
		{"postapoc"},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("expected genre %q, got %q", tt.genre, sys.genreID)
			}
		})
	}
}

func TestEntityFactionOutlineUpdate_HostileTeam(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.SetGenre("fantasy")
	sys.timeSinceCheck = sys.checkInterval // Force allegiance check

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 2}) // enemy

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	comp, ok := entity.GetComponent("entity_faction_outline")
	if !ok {
		t.Fatal("expected entity_faction_outline component")
	}
	oc := comp.(*EntityFactionOutlineComponent)
	if !oc.Visible {
		t.Error("expected outline to be visible for hostile entity")
	}
	if oc.Allegiance != "hostile" {
		t.Errorf("expected 'hostile' allegiance, got %q", oc.Allegiance)
	}
	if oc.OutlineR < 0.5 {
		t.Errorf("expected high red channel for hostile, got %f", oc.OutlineR)
	}
	if oc.PulseSpeed <= 0 {
		t.Error("expected positive pulse speed for hostile")
	}
}

func TestEntityFactionOutlineUpdate_AllyTeam(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(2)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 1}) // player team ally

	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("entity_faction_outline")
	oc := comp.(*EntityFactionOutlineComponent)
	if !oc.Visible {
		t.Error("expected outline to be visible for ally")
	}
	if oc.Allegiance != "ally" {
		t.Errorf("expected 'ally' allegiance, got %q", oc.Allegiance)
	}
	if oc.OutlineG < 0.5 {
		t.Errorf("expected high green channel for ally, got %f", oc.OutlineG)
	}
}

func TestEntityFactionOutlineUpdate_NeutralTeam(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(3)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 0}) // neutral

	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("entity_faction_outline")
	oc := comp.(*EntityFactionOutlineComponent)
	if !oc.Visible {
		t.Error("expected outline to be visible for neutral")
	}
	if oc.Allegiance != "neutral" {
		t.Errorf("expected 'neutral' allegiance, got %q", oc.Allegiance)
	}
}

func TestEntityFactionOutlineUpdate_PlayerInputSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(4)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 1})
	entity.AddComponent(&stubFactionOutlineInputComponent{}) // player entity

	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("entity_faction_outline")
	if !ok {
		t.Fatal("expected component to be created")
	}
	oc := comp.(*EntityFactionOutlineComponent)
	if oc.Visible {
		t.Error("expected outline NOT visible for player's own entity")
	}
}

func TestEntityFactionOutlineUpdate_NoTeamSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(5)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	// No TeamComponent

	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("entity_faction_outline")
	if ok {
		t.Error("expected no outline component for entity without team")
	}
}

func TestEntityFactionOutlineUpdate_HostilePulse(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(6)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 2})

	// First update to create and set allegiance
	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("entity_faction_outline")
	oc := comp.(*EntityFactionOutlineComponent)
	initialPhase := oc.PulsePhase

	// Second update to animate pulse (no allegiance check needed)
	sys.timeSinceCheck = 0
	sys.Update([]*Entity{entity}, 0.5)

	if math.Abs(oc.PulsePhase-initialPhase) < 0.01 {
		t.Error("expected pulse phase to advance for hostile entity")
	}
}

func TestEntityFactionOutlineUpdate_FactionReputationOverride(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(7)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 0}) // neutral by team
	entity.AddComponent(&FactionComponent{
		FactionID:       "guild_a",
		Reputation:      75,
		IsPlayerFaction: true,
	}) // friendly by reputation

	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("entity_faction_outline")
	oc := comp.(*EntityFactionOutlineComponent)
	if oc.Allegiance != "ally" {
		t.Errorf("expected faction reputation to override to 'ally', got %q", oc.Allegiance)
	}
}

func TestEntityFactionOutlineUpdate_FactionHostileOverride(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = sys.checkInterval

	entity := NewEntity(8)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 0}) // neutral by team
	entity.AddComponent(&FactionComponent{
		FactionID:       "raiders",
		Reputation:      -80,
		IsPlayerFaction: true,
	}) // hostile by reputation

	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("entity_faction_outline")
	oc := comp.(*EntityFactionOutlineComponent)
	if oc.Allegiance != "hostile" {
		t.Errorf("expected faction reputation to override to 'hostile', got %q", oc.Allegiance)
	}
}

func TestEntityFactionOutlineGenrePresets(t *testing.T) {
	world := NewWorld()

	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewEntityFactionOutlineSystem(world, 42)
			sys.SetGenre(genre)
			sys.timeSinceCheck = sys.checkInterval

			entity := NewEntity(100)
			entity.AddComponent(&PositionComponent{X: 10, Y: 10})
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&TeamComponent{TeamID: 2}) // hostile

			sys.Update([]*Entity{entity}, 0.016)

			comp, _ := entity.GetComponent("entity_faction_outline")
			oc := comp.(*EntityFactionOutlineComponent)
			if !oc.Visible {
				t.Error("expected visible outline")
			}
			if oc.Intensity <= 0 {
				t.Error("expected positive intensity")
			}
			if oc.Thickness <= 0 {
				t.Error("expected positive thickness")
			}
			// Clean up for next test
			entity.RemoveComponent("entity_faction_outline")
		})
	}
}

func TestEntityFactionOutlineThrottle(t *testing.T) {
	world := NewWorld()
	sys := NewEntityFactionOutlineSystem(world, 42)
	sys.timeSinceCheck = 0 // Won't trigger check

	entity := NewEntity(9)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&TeamComponent{TeamID: 2})

	// First update with small deltaTime - won't trigger check
	sys.Update([]*Entity{entity}, 0.01)

	comp, ok := entity.GetComponent("entity_faction_outline")
	if !ok {
		t.Fatal("expected component created even without check")
	}
	oc := comp.(*EntityFactionOutlineComponent)
	// Should not be visible yet since allegiance check hasn't run
	if oc.Visible {
		t.Error("expected outline not visible before first allegiance check")
	}
}

func TestClampOutline(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		result := clampOutline(tt.input)
		if result != tt.expected {
			t.Errorf("clampOutline(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}
