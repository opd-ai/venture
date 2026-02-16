package engine

import (
	"testing"
)

// TestGetWeatherSpriteTint_Cached verifies the cached getter returns the added component.
func TestGetWeatherSpriteTint_Cached(t *testing.T) {
	e := NewEntity(1)
	if e.GetWeatherSpriteTint() != nil {
		t.Error("expected nil before adding component")
	}

	wt := &WeatherSpriteTintComponent{TintR: 0.8, TintG: 0.9, TintB: 1.0}
	e.AddComponent(wt)
	got := e.GetWeatherSpriteTint()
	if got != wt {
		t.Error("cached getter should return added component")
	}
	if got.TintR != 0.8 || got.TintG != 0.9 || got.TintB != 1.0 {
		t.Errorf("tint values mismatch: got R=%f G=%f B=%f", got.TintR, got.TintG, got.TintB)
	}
}

// TestGetWeatherSpriteTint_Remove verifies cache is cleared on removal.
func TestGetWeatherSpriteTint_Remove(t *testing.T) {
	e := NewEntity(2)
	wt := &WeatherSpriteTintComponent{TintR: 1.0, TintG: 1.0, TintB: 1.0}
	e.AddComponent(wt)
	if e.GetWeatherSpriteTint() == nil {
		t.Fatal("expected non-nil after add")
	}
	e.RemoveComponent("weather_sprite_tint")
	if e.GetWeatherSpriteTint() != nil {
		t.Error("expected nil after removal")
	}
}

// TestGetCreatureGenreTint_Cached verifies the cached getter returns the added component.
func TestGetCreatureGenreTint_Cached(t *testing.T) {
	e := NewEntity(3)
	if e.GetCreatureGenreTint() != nil {
		t.Error("expected nil before adding component")
	}

	ct := &CreatureGenreTintComponent{TintR: 0.5, TintG: 0.6, TintB: 0.7}
	e.AddComponent(ct)
	got := e.GetCreatureGenreTint()
	if got != ct {
		t.Error("cached getter should return added component")
	}
	if got.TintR != 0.5 || got.TintG != 0.6 || got.TintB != 0.7 {
		t.Errorf("tint values mismatch: got R=%f G=%f B=%f", got.TintR, got.TintG, got.TintB)
	}
}

// TestGetCreatureGenreTint_Remove verifies cache is cleared on removal.
func TestGetCreatureGenreTint_Remove(t *testing.T) {
	e := NewEntity(4)
	ct := &CreatureGenreTintComponent{TintR: 1.0, TintG: 1.0, TintB: 1.0}
	e.AddComponent(ct)
	if e.GetCreatureGenreTint() == nil {
		t.Fatal("expected non-nil after add")
	}
	e.RemoveComponent("creature_genre_tint")
	if e.GetCreatureGenreTint() != nil {
		t.Error("expected nil after removal")
	}
}

// TestExtractVisualFeedback_TintComposition verifies tint composition uses cached getters.
func TestExtractVisualFeedback_TintComposition(t *testing.T) {
	r := NewRenderSystem(nil)
	e := NewEntity(5)

	// No tint components - should return defaults
	_, tintR, tintG, tintB, tintA := r.extractVisualFeedback(e)
	if tintR != 1.0 || tintG != 1.0 || tintB != 1.0 || tintA != 1.0 {
		t.Errorf("expected default tints (1,1,1,1), got (%f,%f,%f,%f)", tintR, tintG, tintB, tintA)
	}

	// Add weather tint
	e.AddComponent(&WeatherSpriteTintComponent{TintR: 0.5, TintG: 0.6, TintB: 0.7})
	_, tintR, tintG, tintB, _ = r.extractVisualFeedback(e)
	if tintR != 0.5 || tintG != 0.6 || tintB != 0.7 {
		t.Errorf("weather tint not applied: got (%f,%f,%f)", tintR, tintG, tintB)
	}

	// Add creature genre tint - should compose multiplicatively
	e.AddComponent(&CreatureGenreTintComponent{TintR: 0.8, TintG: 0.5, TintB: 1.0})
	_, tintR, tintG, tintB, _ = r.extractVisualFeedback(e)
	expectedR, expectedG, expectedB := 0.5*0.8, 0.6*0.5, 0.7*1.0
	if tintR != expectedR || tintG != expectedG || tintB != expectedB {
		t.Errorf("composed tint mismatch: expected (%f,%f,%f), got (%f,%f,%f)",
			expectedR, expectedG, expectedB, tintR, tintG, tintB)
	}
}

// TestFilterParticleEntities verifies the particle entity filtering.
func TestFilterParticleEntities(t *testing.T) {
	r := NewRenderSystem(nil)

	// Create mix of entities with and without particle emitters
	entities := make([]*Entity, 5)
	for i := range entities {
		entities[i] = NewEntity(uint64(i + 1))
		entities[i].AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
	}

	// Add particle emitters to entities 1 and 3
	entities[0].AddComponent(&ParticleEmitterComponent{})
	entities[2].AddComponent(&ParticleEmitterComponent{})

	filtered := r.filterParticleEntities(entities)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 particle entities, got %d", len(filtered))
	}
	if filtered[0].ID != 1 || filtered[1].ID != 3 {
		t.Errorf("expected entity IDs [1,3], got [%d,%d]", filtered[0].ID, filtered[1].ID)
	}
}

// TestFilterParticleEntities_Empty verifies empty entity list handling.
func TestFilterParticleEntities_Empty(t *testing.T) {
	r := NewRenderSystem(nil)
	filtered := r.filterParticleEntities(nil)
	if len(filtered) != 0 {
		t.Errorf("expected 0 particle entities for nil input, got %d", len(filtered))
	}
}

// TestFilterParticleEntities_NoEmitters verifies no emitters case.
func TestFilterParticleEntities_NoEmitters(t *testing.T) {
	r := NewRenderSystem(nil)
	entities := []*Entity{NewEntity(1), NewEntity(2)}
	filtered := r.filterParticleEntities(entities)
	if len(filtered) != 0 {
		t.Errorf("expected 0 particle entities, got %d", len(filtered))
	}
}

// TestFilterParticleEntities_BufferReuse verifies buffer reuse across calls.
func TestFilterParticleEntities_BufferReuse(t *testing.T) {
	r := NewRenderSystem(nil)

	e1 := NewEntity(1)
	e1.AddComponent(&ParticleEmitterComponent{})
	e2 := NewEntity(2)
	e2.AddComponent(&ParticleEmitterComponent{})

	// First call
	r.filterParticleEntities([]*Entity{e1, e2})

	// Second call with different entities - buffer should be reused (not grow unbounded)
	e3 := NewEntity(3)
	e3.AddComponent(&ParticleEmitterComponent{})
	filtered := r.filterParticleEntities([]*Entity{e3})
	if len(filtered) != 1 {
		t.Errorf("expected 1 particle entity on second call, got %d", len(filtered))
	}
}
