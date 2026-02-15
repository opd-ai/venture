package engine

import (
	"math"
	"testing"
)

// getWoundComp is a helper to retrieve the wound component from an entity.
func getWoundComp(entity *Entity) Component {
	comp, _ := entity.GetComponent("battle_wound_overlay")
	return comp
}

func TestBattleWoundOverlayComponentType(t *testing.T) {
	c := &BattleWoundOverlayComponent{}
	if c.Type() != "battle_wound_overlay" {
		t.Errorf("expected type 'battle_wound_overlay', got %q", c.Type())
	}
}

func TestSeverityFromHealthRatio(t *testing.T) {
	tests := []struct {
		name     string
		ratio    float64
		expected WoundSeverity
	}{
		{"full health", 1.0, WoundNone},
		{"above full", 1.5, WoundNone},
		{"scratched high", 0.90, WoundScratched},
		{"scratched low", 0.76, WoundScratched},
		{"wounded high", 0.75, WoundWounded},
		{"wounded low", 0.51, WoundWounded},
		{"bloodied high", 0.50, WoundBloodied},
		{"bloodied low", 0.26, WoundBloodied},
		{"critical", 0.25, WoundCritical},
		{"near death", 0.05, WoundCritical},
		{"zero health", 0.0, WoundCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := severityFromHealthRatio(tt.ratio)
			if got != tt.expected {
				t.Errorf("severityFromHealthRatio(%f) = %d, want %d", tt.ratio, got, tt.expected)
			}
		})
	}
}

func TestBuildWoundPalettes(t *testing.T) {
	palettes := buildWoundPalettes()
	expectedGenres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		t.Run(genre, func(t *testing.T) {
			p, ok := palettes[genre]
			if !ok {
				t.Fatalf("missing palette for genre %q", genre)
			}
			if p.R < 0 || p.R > 1 || p.G < 0 || p.G > 1 || p.B < 0 || p.B > 1 {
				t.Errorf("wound color out of range: R=%f G=%f B=%f", p.R, p.G, p.B)
			}
			if p.PulseSpeed <= 0 {
				t.Errorf("pulse speed should be positive, got %f", p.PulseSpeed)
			}
		})
	}
}

func TestNewBattleWoundOverlaySystem(t *testing.T) {
	sys := NewBattleWoundOverlaySystem(nil, 42)
	if sys == nil {
		t.Fatal("system should not be nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre should be fantasy, got %q", sys.genreID)
	}
	if sys.rng == nil {
		t.Error("rng should not be nil")
	}
	if sys.prevHealth == nil {
		t.Error("prevHealth map should be initialized")
	}
}

func TestBattleWoundOverlaySystemSetGenre(t *testing.T) {
	sys := NewBattleWoundOverlaySystem(nil, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genre should be horror, got %q", sys.genreID)
	}
}

func TestBattleWoundOverlaySystemUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// First update: registers health, no wounds yet
	sys.Update(entities, 0.016)

	comp := getWoundComp(entity)
	if comp == nil {
		t.Fatal("entity should have battle_wound_overlay component after update")
	}
	wc := comp.(*BattleWoundOverlayComponent)
	if wc.Severity != WoundNone {
		t.Errorf("expected WoundNone at full health, got %d", wc.Severity)
	}

	// Simulate damage to 60% health
	health := entity.GetHealth()
	health.Current = 60
	sys.Update(entities, 0.016)

	wc = getWoundComp(entity).(*BattleWoundOverlayComponent)
	if wc.Severity != WoundWounded {
		t.Errorf("expected WoundWounded at 60%% health, got %d", wc.Severity)
	}
	if wc.MarkCount == 0 {
		t.Error("expected wound marks after taking damage")
	}
	if wc.Opacity <= 0 {
		t.Error("expected non-zero opacity when wounded")
	}
	if !wc.Dirty {
		t.Error("component should be dirty after severity change")
	}
}

func TestBattleWoundOverlaySystemCriticalPulse(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 99)
	sys.SetGenre("scifi")

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// Bootstrap tracking
	sys.Update(entities, 0.016)

	// Drop to critical health
	entity.GetHealth().Current = 10
	sys.Update(entities, 0.016)

	wc := getWoundComp(entity).(*BattleWoundOverlayComponent)
	if wc.Severity != WoundCritical {
		t.Fatalf("expected WoundCritical, got %d", wc.Severity)
	}

	initialPhase := wc.PulsePhase

	// Advance time to trigger pulse animation
	sys.Update(entities, 0.5)

	if wc.PulsePhase <= initialPhase {
		t.Error("pulse phase should advance over time at critical health")
	}
	if wc.Opacity <= 0.5 {
		t.Errorf("critical opacity should be high, got %f", wc.Opacity)
	}
}

func TestBattleWoundOverlaySystemHealing(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 55)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// Bootstrap, damage, then heal
	sys.Update(entities, 0.016)
	entity.GetHealth().Current = 30
	sys.Update(entities, 0.016)

	wc := getWoundComp(entity).(*BattleWoundOverlayComponent)
	if wc.Severity == WoundNone {
		t.Fatal("should be wounded after damage")
	}

	// Heal to full
	entity.GetHealth().Current = 100
	sys.Update(entities, 0.016)

	if wc.Severity != WoundNone {
		t.Errorf("should be WoundNone after full heal, got %d", wc.Severity)
	}
	if wc.MarkCount != 0 {
		t.Errorf("wound marks should be cleared after full heal, got %d", wc.MarkCount)
	}
	if wc.Opacity != 0 {
		t.Errorf("opacity should be 0 after heal, got %f", wc.Opacity)
	}
}

func TestBattleWoundOverlaySystemMultipleDamageEvents(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 777)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// Bootstrap
	sys.Update(entities, 0.016)

	// Apply multiple damage events
	for i := 0; i < 6; i++ {
		entity.GetHealth().Current -= 10
		sys.Update(entities, 0.016)
	}

	wc := getWoundComp(entity).(*BattleWoundOverlayComponent)
	if wc.MarkCount < 4 {
		t.Errorf("expected at least 4 wound marks after 6 damage events, got %d", wc.MarkCount)
	}
	if wc.MarkCount > 8 {
		t.Errorf("wound marks should cap at 8, got %d", wc.MarkCount)
	}
}

func TestBattleWoundOverlaySystemGenreColors(t *testing.T) {
	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewBattleWoundOverlaySystem(world, 42)
			sys.SetGenre(genre)

			entity := world.CreateEntity()
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

			entities := []*Entity{entity}

			sys.Update(entities, 0.016)
			entity.GetHealth().Current = 40
			sys.Update(entities, 0.016)

			wc := getWoundComp(entity).(*BattleWoundOverlayComponent)
			if wc.WoundR == 0 && wc.WoundG == 0 && wc.WoundB == 0 {
				t.Error("wound color should be set for genre " + genre)
			}
		})
	}
}

func TestBattleWoundOverlaySystemNilWorld(t *testing.T) {
	sys := NewBattleWoundOverlaySystem(nil, 1)
	// Should not panic with nil world
	sys.Update([]*Entity{}, 0.016)
}

func TestBattleWoundOverlaySystemEntityCleanup(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 1)
	sys.cleanupInterval = 0 // Force immediate cleanup

	e1 := world.CreateEntity()
	e1.AddComponent(&HealthComponent{Current: 50, Max: 100})
	e2 := world.CreateEntity()
	e2.AddComponent(&HealthComponent{Current: 80, Max: 100})

	sys.Update([]*Entity{e1, e2}, 0.016)
	if len(sys.prevHealth) != 2 {
		t.Fatalf("should track 2 entities, got %d", len(sys.prevHealth))
	}

	// Remove e2 from entity list
	sys.Update([]*Entity{e1}, 0.1)
	if len(sys.prevHealth) != 1 {
		t.Errorf("should have cleaned up to 1 tracked entity, got %d", len(sys.prevHealth))
	}
}

func TestBattleWoundOverlaySystemGetActivePaletteFallback(t *testing.T) {
	sys := NewBattleWoundOverlaySystem(nil, 1)
	sys.SetGenre("nonexistent")
	palette := sys.getActivePalette()
	// Should fall back to fantasy
	fantasyPalette := sys.palettes["fantasy"]
	if palette.R != fantasyPalette.R || palette.G != fantasyPalette.G {
		t.Error("should fall back to fantasy palette for unknown genre")
	}
}

func TestBattleWoundOverlayMarkBounds(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.016)
	entity.GetHealth().Current = 20
	sys.Update([]*Entity{entity}, 0.016)

	wc := getWoundComp(entity).(*BattleWoundOverlayComponent)
	for i := 0; i < wc.MarkCount; i++ {
		if wc.MarkOffsetsX[i] < 0.2 || wc.MarkOffsetsX[i] > 0.8 {
			t.Errorf("mark X offset out of bounds: %f", wc.MarkOffsetsX[i])
		}
		if wc.MarkOffsetsY[i] < 0.2 || wc.MarkOffsetsY[i] > 0.8 {
			t.Errorf("mark Y offset out of bounds: %f", wc.MarkOffsetsY[i])
		}
		if wc.MarkSizes[i] < 1.0 || wc.MarkSizes[i] > 4.0 {
			t.Errorf("mark size out of bounds: %f", wc.MarkSizes[i])
		}
	}
}

func BenchmarkBattleWoundOverlaySystem(b *testing.B) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 500)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&HealthComponent{Current: float64(50 + i%50), Max: 100})
		entities[i] = e
	}

	// Bootstrap tracking
	sys.Update(entities, 0.016)

	// Apply some damage
	for i := range entities {
		entities[i].GetHealth().Current -= 10
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func TestUpdateSeverityVisualsAllTiers(t *testing.T) {
	sys := NewBattleWoundOverlaySystem(nil, 1)
	palette := sys.palettes["fantasy"]

	tests := []struct {
		name        string
		severity    WoundSeverity
		healthRatio float64
		minOpacity  float64
		maxOpacity  float64
	}{
		{"scratched", WoundScratched, 0.80, 0.2, 0.3},
		{"wounded", WoundWounded, 0.60, 0.4, 0.5},
		{"bloodied", WoundBloodied, 0.35, 0.6, 0.7},
		{"critical", WoundCritical, 0.10, 0.7, 0.9},
		{"none", WoundNone, 1.0, 0.0, 0.01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &BattleWoundOverlayComponent{Severity: tt.severity}
			sys.updateSeverityVisuals(comp, palette, tt.healthRatio)
			if comp.Opacity < tt.minOpacity || comp.Opacity > tt.maxOpacity {
				t.Errorf("opacity %f not in range [%f, %f]", comp.Opacity, tt.minOpacity, tt.maxOpacity)
			}
		})
	}
}

// Verify pulse phase wraps correctly
func TestCriticalPulsePhaseWrap(t *testing.T) {
	world := NewWorld()
	sys := NewBattleWoundOverlaySystem(world, 1)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)
	entity.GetHealth().Current = 5
	sys.Update(entities, 0.016)

	// Run enough updates to wrap phase past 2*pi
	for i := 0; i < 100; i++ {
		sys.Update(entities, 0.1)
	}

	wc := getWoundComp(entity).(*BattleWoundOverlayComponent)
	if wc.PulsePhase >= 2*math.Pi {
		t.Errorf("pulse phase should wrap below 2*pi, got %f", wc.PulsePhase)
	}
}
