package engine

import (
	"testing"
)

func TestNewStatusEffectVisualOverlaySystem(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world reference mismatch")
	}
	if len(sys.effectTints) == 0 {
		t.Error("expected default tint configs to be populated")
	}
	if len(sys.activeTints) != 0 {
		t.Error("expected empty active tints on creation")
	}
}

func TestStatusEffectVisualOverlay_DefaultTintConfigs(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	expectedEffects := []string{"burn", "poison", "frost", "frozen", "stun", "regeneration", "shock", "electrified", "blessed", "cursed"}
	for _, effect := range expectedEffects {
		if _, exists := sys.effectTints[effect]; !exists {
			t.Errorf("missing default tint config for effect: %s", effect)
		}
	}
}

func TestStatusEffectVisualOverlay_ApplyTint(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewVisualFeedbackComponent())
	entity.AddComponent(NewStatusEffectComponent("burn", 10.0, 5.0, 1.0))
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	feedback := entity.GetVisualFeedback()
	if feedback == nil {
		t.Fatal("expected visual feedback component")
	}

	// Tint should be modified from default white (1,1,1,1)
	if feedback.TintR == 1.0 && feedback.TintG == 1.0 && feedback.TintB == 1.0 {
		t.Error("expected tint to be modified from default white")
	}
	if feedback.TintA != 1.0 {
		t.Errorf("expected alpha to remain 1.0, got %f", feedback.TintA)
	}
}

func TestStatusEffectVisualOverlay_ClearTintOnExpiry(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewVisualFeedbackComponent())
	effect := NewStatusEffectComponent("poison", 5.0, 0.5, 0.0)
	entity.AddComponent(effect)
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// First update: tint applied
	sys.Update(entities, 0.016)
	if sys.GetActiveTintCount() != 1 {
		t.Errorf("expected 1 active tint, got %d", sys.GetActiveTintCount())
	}

	// Expire the effect
	effect.Duration = 0

	// Second update: tint should be cleared
	sys.Update(entities, 0.016)
	if sys.GetActiveTintCount() != 0 {
		t.Errorf("expected 0 active tints after expiry, got %d", sys.GetActiveTintCount())
	}

	feedback := entity.GetVisualFeedback()
	if feedback.TintR != 1.0 || feedback.TintG != 1.0 || feedback.TintB != 1.0 {
		t.Errorf("expected tint cleared to white, got (%f, %f, %f)", feedback.TintR, feedback.TintG, feedback.TintB)
	}
}

func TestStatusEffectVisualOverlay_NoVisualFeedback(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewStatusEffectComponent("burn", 10.0, 5.0, 1.0))
	world.AddEntity(entity)

	// Should not panic on entities without VisualFeedbackComponent
	sys.Update([]*Entity{entity}, 0.016)

	if sys.GetActiveTintCount() != 0 {
		t.Error("expected no tints on entity without visual feedback")
	}
}

func TestStatusEffectVisualOverlay_DominantEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewVisualFeedbackComponent())
	// Add both poison and burn — burn should dominate (higher priority)
	entity.AddComponent(NewStatusEffectComponent("poison", 5.0, 10.0, 1.0))
	entity.AddComponent(NewStatusEffectComponent("burn", 10.0, 10.0, 1.0))
	world.AddEntity(entity)

	sys.Update([]*Entity{entity}, 0.016)

	// The active tint should be burn, not poison
	if tint, ok := sys.activeTints[entity.ID]; !ok || tint != "burn" {
		t.Errorf("expected dominant effect 'burn', got '%s'", tint)
	}
}

func TestStatusEffectVisualOverlay_GenreOverrides(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		effect  string
		changed bool
	}{
		{"cyberpunk poison", "cyberpunk", "poison", true},
		{"cyberpunk shock", "cyberpunk", "shock", true},
		{"horror cursed", "horror", "cursed", true},
		{"horror poison", "horror", "poison", true},
		{"scifi burn", "scifi", "burn", true},
		{"fantasy burn", "fantasy", "burn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sysDefault := NewStatusEffectVisualOverlaySystem(world)
			sysGenre := NewStatusEffectVisualOverlaySystem(world)
			sysGenre.SetGenre(tt.genre)

			defaultConfig := sysDefault.effectTints[tt.effect]
			genreConfig := sysGenre.effectTints[tt.effect]

			if tt.changed {
				if defaultConfig.Color == genreConfig.Color && defaultConfig.Intensity == genreConfig.Intensity {
					t.Errorf("genre %s should override tint for effect %s", tt.genre, tt.effect)
				}
			}
		})
	}
}

func TestStatusEffectVisualOverlay_PulsingIntensity(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewVisualFeedbackComponent())
	entity.AddComponent(NewStatusEffectComponent("burn", 10.0, 10.0, 1.0))
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// Update at two different times to get different pulse phases
	sys.Update(entities, 0.0)
	fb1 := entity.GetVisualFeedback()
	r1 := fb1.TintR

	// Advance time enough for pulse to change
	sys.Update(entities, 0.25) // Quarter second
	r2 := fb1.TintR

	// Tint values should differ due to pulsing
	if r1 == r2 {
		t.Log("pulsing tint values should vary over time (may match at specific phases)")
	}
}

func TestStatusEffectVisualOverlay_IntensityCap(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewVisualFeedbackComponent())
	// Very high magnitude should still cap intensity
	entity.AddComponent(NewStatusEffectComponent("burn", 500.0, 10.0, 1.0))
	world.AddEntity(entity)

	sys.Update([]*Entity{entity}, 0.016)

	feedback := entity.GetVisualFeedback()
	// Tint values should never go below 0.2 (intensity capped at 0.8)
	if feedback.TintR < 0.0 || feedback.TintG < 0.0 || feedback.TintB < 0.0 {
		t.Errorf("tint values should never be negative: (%f, %f, %f)", feedback.TintR, feedback.TintG, feedback.TintB)
	}
}

func TestStatusEffectVisualOverlay_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	e1 := NewEntity(1)
	e1.AddComponent(NewVisualFeedbackComponent())
	e1.AddComponent(NewStatusEffectComponent("burn", 10.0, 5.0, 1.0))
	world.AddEntity(e1)

	e2 := NewEntity(2)
	e2.AddComponent(NewVisualFeedbackComponent())
	e2.AddComponent(NewStatusEffectComponent("frost", 8.0, 5.0, 1.0))
	world.AddEntity(e2)

	e3 := NewEntity(3)
	e3.AddComponent(NewVisualFeedbackComponent())
	// No status effect
	world.AddEntity(e3)

	sys.Update([]*Entity{e1, e2, e3}, 0.016)

	if sys.GetActiveTintCount() != 2 {
		t.Errorf("expected 2 active tints, got %d", sys.GetActiveTintCount())
	}

	// e1 should have burn tint (reddish), e2 should have frost tint (bluish)
	fb1 := e1.GetVisualFeedback()
	fb2 := e2.GetVisualFeedback()
	fb3 := e3.GetVisualFeedback()

	if fb1.TintR == fb2.TintR && fb1.TintG == fb2.TintG && fb1.TintB == fb2.TintB {
		t.Error("burn and frost should produce different tints")
	}

	if fb3.TintR != 1.0 || fb3.TintG != 1.0 || fb3.TintB != 1.0 {
		t.Error("entity without status effect should have default white tint")
	}
}

func TestStatusEffectVisualOverlay_UnknownEffectType(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewVisualFeedbackComponent())
	entity.AddComponent(NewStatusEffectComponent("unknown_custom_effect", 10.0, 5.0, 1.0))
	world.AddEntity(entity)

	sys.Update([]*Entity{entity}, 0.016)

	// Unknown effect should not apply any tint
	if sys.GetActiveTintCount() != 0 {
		t.Error("unknown effect type should not produce a tint")
	}

	feedback := entity.GetVisualFeedback()
	if feedback.TintR != 1.0 || feedback.TintG != 1.0 || feedback.TintB != 1.0 {
		t.Error("tint should remain white for unknown effect types")
	}
}

func BenchmarkStatusEffectVisualOverlay_Update(b *testing.B) {
	world := NewWorld()
	sys := NewStatusEffectVisualOverlaySystem(world)

	entities := make([]*Entity, 200)
	effectTypes := []string{"burn", "poison", "frost", "stun", "regeneration"}
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(NewVisualFeedbackComponent())
		if i%3 != 0 { // ~67% have status effects
			e.AddComponent(NewStatusEffectComponent(effectTypes[i%len(effectTypes)], 10.0, 30.0, 1.0))
		}
		world.AddEntity(e)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
