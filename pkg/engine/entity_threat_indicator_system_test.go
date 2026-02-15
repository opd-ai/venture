package engine

import (
	"math"
	"testing"
)

// TestThreatIndicatorComponentType verifies the component type string.
func TestThreatIndicatorComponentType(t *testing.T) {
	c := NewThreatIndicatorComponent()
	if c.Type() != "threat_indicator" {
		t.Errorf("expected type 'threat_indicator', got %q", c.Type())
	}
}

// TestThreatIndicatorComponentDefaults verifies default values.
func TestThreatIndicatorComponentDefaults(t *testing.T) {
	c := NewThreatIndicatorComponent()
	if c.Enabled {
		t.Error("expected Enabled=false by default")
	}
	if c.RingRadius != 8.0 {
		t.Errorf("expected RingRadius=8.0, got %f", c.RingRadius)
	}
}

// TestComputeThreatTier verifies level-difference → tier mapping.
func TestComputeThreatTier(t *testing.T) {
	tests := []struct {
		name        string
		playerLevel int
		entityLevel int
		wantTier    int
	}{
		{"trivial_low", 10, 1, 0},
		{"trivial_exact", 10, 5, 0},
		{"easy_low", 10, 6, 1},
		{"easy_high", 10, 7, 1},
		{"fair_below", 10, 9, 2},
		{"fair_equal", 10, 10, 2},
		{"fair_above", 10, 12, 2},
		{"challenging_low", 10, 13, 3},
		{"challenging_high", 10, 14, 3},
		{"dangerous", 10, 15, 4},
		{"dangerous_high", 10, 20, 4},
		{"level_1_vs_1", 1, 1, 2},
		{"level_1_vs_6", 1, 6, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeThreatTier(tt.playerLevel, tt.entityLevel)
			if got != tt.wantTier {
				t.Errorf("computeThreatTier(%d, %d) = %d, want %d",
					tt.playerLevel, tt.entityLevel, got, tt.wantTier)
			}
		})
	}
}

// TestBuildThreatPalettes verifies all expected genres have 5 tiers.
func TestBuildThreatPalettes(t *testing.T) {
	palettes := buildThreatPalettes()
	expectedGenres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		p, ok := palettes[genre]
		if !ok {
			t.Errorf("missing palette for genre %q", genre)
			continue
		}
		for i, tier := range p.Tiers {
			if tier.Opacity < 0 || tier.Opacity > 1 {
				t.Errorf("genre %s tier %d: opacity %f out of [0,1]", genre, i, tier.Opacity)
			}
		}
	}
}

// TestNewEntityThreatIndicatorSystem verifies system creation with nil world.
func TestNewEntityThreatIndicatorSystem(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("expected updateInterval=0.5, got %f", sys.updateInterval)
	}
}

// TestSetGenre verifies genre setter.
func TestEntityThreatIndicatorSetGenre(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

// TestClampThreat verifies clamping.
func TestClampThreat(t *testing.T) {
	tests := []struct {
		in, want float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		got := clampThreat(tt.in)
		if got != tt.want {
			t.Errorf("clampThreat(%f) = %f, want %f", tt.in, got, tt.want)
		}
	}
}

// TestUpdatePulseAdvancesPhase verifies pulse animation.
func TestUpdatePulseAdvancesPhase(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 99)

	entity := NewEntity(1)
	ti := NewThreatIndicatorComponent()
	ti.Enabled = true
	ti.PulseSpeed = 1.0
	ti.BaseOpacity = 0.5
	ti.CurrentOpacity = 0.5
	entity.AddComponent(ti)

	entities := []*Entity{entity}
	sys.updatePulse(entities, 0.1)

	if ti.PulsePhase == 0 {
		t.Error("expected PulsePhase to advance after updatePulse")
	}
	if ti.CurrentOpacity == 0.5 {
		t.Error("expected CurrentOpacity to change after pulse")
	}
}

// TestUpdatePulseSkipsDisabled verifies disabled components are untouched.
func TestUpdatePulseSkipsDisabled(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 99)

	entity := NewEntity(2)
	ti := NewThreatIndicatorComponent()
	ti.Enabled = false
	ti.PulseSpeed = 1.0
	entity.AddComponent(ti)

	sys.updatePulse([]*Entity{entity}, 0.1)

	if ti.PulsePhase != 0 {
		t.Error("disabled component should not have pulse phase advanced")
	}
}

// TestUpdatePulseWrapsPhase verifies phase wrapping to prevent float growth.
func TestUpdatePulseWrapsPhase(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 99)

	entity := NewEntity(3)
	ti := NewThreatIndicatorComponent()
	ti.Enabled = true
	ti.PulseSpeed = 1.0
	ti.BaseOpacity = 0.5
	ti.PulsePhase = 6.0 // near 2π
	entity.AddComponent(ti)

	sys.updatePulse([]*Entity{entity}, 1.0)

	if ti.PulsePhase > 2*math.Pi+0.01 {
		t.Errorf("phase should wrap, got %f", ti.PulsePhase)
	}
}

// TestFindPlayerLevel returns player level from experience component.
func TestFindPlayerLevel(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)

	player := NewEntity(1)
	player.AddComponent(&StubInput{})
	xp := NewExperienceComponent()
	xp.Level = 7
	player.AddComponent(xp)

	got := sys.findPlayerLevel([]*Entity{player})
	if got != 7 {
		t.Errorf("expected player level 7, got %d", got)
	}
}

// TestFindPlayerLevelFallbackHealth returns health-based estimate.
func TestFindPlayerLevelFallbackHealth(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)

	player := NewEntity(1)
	player.AddComponent(&StubInput{})
	hc := &HealthComponent{Current: 300, Max: 300}
	player.AddComponent(hc)

	got := sys.findPlayerLevel([]*Entity{player})
	if got != 3 {
		t.Errorf("expected estimated level 3, got %d", got)
	}
}

// TestFindPlayerLevelNoPlayer returns -1 when no player exists.
func TestFindPlayerLevelNoPlayer(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)
	got := sys.findPlayerLevel([]*Entity{})
	if got != -1 {
		t.Errorf("expected -1 with no players, got %d", got)
	}
}

// TestEstimateEntityLevel verifies experience-based and fallback estimation.
func TestEstimateEntityLevel(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)

	// With experience
	e1 := NewEntity(1)
	xp := NewExperienceComponent()
	xp.Level = 5
	e1.AddComponent(xp)
	if got := sys.estimateEntityLevel(e1); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}

	// With health only
	e2 := NewEntity(2)
	e2.AddComponent(&HealthComponent{Current: 200, Max: 200})
	if got := sys.estimateEntityLevel(e2); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}

	// No components
	e3 := NewEntity(3)
	if got := sys.estimateEntityLevel(e3); got != 1 {
		t.Errorf("expected 1 (default), got %d", got)
	}
}

// TestUpdateAssignsThreatIndicator verifies full Update assigns components.
func TestUpdateAssignsThreatIndicator(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)
	sys.SetGenre("fantasy")

	player := NewEntity(1)
	player.AddComponent(&StubInput{})
	xp := NewExperienceComponent()
	xp.Level = 5
	player.AddComponent(xp)

	npc := NewEntity(2)
	npc.AddComponent(&AIComponent{DetectionRange: 100})
	npc.AddComponent(&StubSprite{})
	npcXP := NewExperienceComponent()
	npcXP.Level = 10
	npc.AddComponent(npcXP)

	entities := []*Entity{player, npc}

	// First update triggers assignment (timeSinceCheck starts at 0 < 0.5)
	sys.Update(entities, 1.0)

	comp, ok := npc.GetComponent("threat_indicator")
	if !ok {
		t.Fatal("expected threat_indicator component on NPC")
	}
	ti, ok := comp.(*ThreatIndicatorComponent)
	if !ok {
		t.Fatal("expected *ThreatIndicatorComponent")
	}
	if !ti.Enabled {
		t.Error("expected Enabled=true")
	}
	// Level diff = 10 - 5 = 5 → dangerous (tier 4)
	if ti.ThreatTier != 4 {
		t.Errorf("expected tier 4 (dangerous), got %d", ti.ThreatTier)
	}
}

// TestUpdateSkipsPlayerEntities verifies players don't get threat rings.
func TestUpdateSkipsPlayerEntities(t *testing.T) {
	sys := NewEntityThreatIndicatorSystem(nil, 42)

	player := NewEntity(1)
	player.AddComponent(&StubInput{})
	player.AddComponent(&AIComponent{})
	player.AddComponent(&StubSprite{})
	player.AddComponent(NewExperienceComponent())

	sys.Update([]*Entity{player}, 1.0)

	_, ok := player.GetComponent("threat_indicator")
	if ok {
		t.Error("player should not receive a threat indicator")
	}
}
