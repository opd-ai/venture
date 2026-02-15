package engine

import (
	"testing"
)

func TestDamageFlashTintSystem_Creation(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestDamageFlashTintSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		wantR   uint8
	}{
		{"fantasy", "fantasy", 255},
		{"horror", "horror", 180},
		{"cyberpunk", "cyberpunk", 0},
		{"scifi", "scifi", 255},
		{"postapoc", "postapoc", 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewDamageFlashTintSystem(nil)
			sys.SetGenre(tt.genreID)
			if sys.preset.Color.R != tt.wantR {
				t.Errorf("genre %q: expected color R=%d, got %d", tt.genreID, tt.wantR, sys.preset.Color.R)
			}
		})
	}
}

func TestDamageFlashTintSystem_TriggerOnDamage(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(1)
	health := &HealthComponent{Current: 100, Max: 100}
	feedback := NewVisualFeedbackComponent()
	entity.AddComponent(health)
	entity.AddComponent(feedback)

	entities := []*Entity{entity}

	// First update: record baseline health
	sys.Update(entities, 0.016)
	if feedback.IsFlashing() {
		t.Error("should not flash on first frame (no previous health)")
	}

	// Take damage
	health.Current = 70

	// Second update: should trigger flash
	sys.Update(entities, 0.016)
	if !feedback.IsFlashing() {
		t.Error("expected flash after taking damage")
	}
	if feedback.FlashIntensity <= 0 {
		t.Error("expected positive flash intensity")
	}
}

func TestDamageFlashTintSystem_NoFlashOnHeal(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(2)
	health := &HealthComponent{Current: 50, Max: 100}
	feedback := NewVisualFeedbackComponent()
	entity.AddComponent(health)
	entity.AddComponent(feedback)

	entities := []*Entity{entity}

	// Record baseline
	sys.Update(entities, 0.016)

	// Heal
	health.Current = 80
	sys.Update(entities, 0.016)

	if feedback.IsFlashing() {
		t.Error("should not flash on healing")
	}
}

func TestDamageFlashTintSystem_IntensityScaling(t *testing.T) {
	tests := []struct {
		name       string
		maxHP      float64
		startHP    float64
		damage     float64
		wantMinInt float64
		wantMaxInt float64
	}{
		{"small_hit", 100, 100, 5, 0.15, 0.35},
		{"medium_hit", 100, 100, 30, 0.30, 0.60},
		{"big_hit", 100, 100, 80, 0.60, 0.90},
		{"one_shot", 100, 100, 100, 0.80, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewDamageFlashTintSystem(nil)

			entity := NewEntity(10)
			health := &HealthComponent{Current: tt.startHP, Max: tt.maxHP}
			feedback := NewVisualFeedbackComponent()
			entity.AddComponent(health)
			entity.AddComponent(feedback)

			entities := []*Entity{entity}
			sys.Update(entities, 0.016)

			health.Current = tt.startHP - tt.damage
			sys.Update(entities, 0.016)

			if feedback.FlashIntensity < tt.wantMinInt || feedback.FlashIntensity > tt.wantMaxInt {
				t.Errorf("intensity %.3f not in [%.2f, %.2f]",
					feedback.FlashIntensity, tt.wantMinInt, tt.wantMaxInt)
			}
		})
	}
}

func TestDamageFlashTintSystem_NoHealthComponent(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(3)
	feedback := NewVisualFeedbackComponent()
	entity.AddComponent(feedback)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016) // Should not panic
	if feedback.IsFlashing() {
		t.Error("should not flash without health component")
	}
}

func TestDamageFlashTintSystem_NoFeedbackComponent(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(4)
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)
	health.Current = 50
	sys.Update(entities, 0.016) // Should not panic
}

func TestDamageFlashTintSystem_GenreFlashColors(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		wantR   float64 // TintR should be < 1.0 (tinted away from white)
	}{
		{"horror_red", "horror", 1.0},        // Red channel stays high
		{"cyberpunk_cyan", "cyberpunk", 0.5}, // Red channel reduced (cyan = no red)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewDamageFlashTintSystem(nil)
			sys.SetGenre(tt.genreID)

			entity := NewEntity(20)
			health := &HealthComponent{Current: 100, Max: 100}
			feedback := NewVisualFeedbackComponent()
			entity.AddComponent(health)
			entity.AddComponent(feedback)

			entities := []*Entity{entity}
			sys.Update(entities, 0.016)

			health.Current = 50
			sys.Update(entities, 0.016)

			if tt.genreID == "cyberpunk" && feedback.TintR >= 1.0 {
				t.Error("cyberpunk flash should reduce red channel (cyan tint)")
			}
			if tt.genreID == "horror" && feedback.TintR < 0.5 {
				t.Error("horror flash should keep red channel high")
			}
		})
	}
}

func TestDamageFlashTintSystem_CleanupStaleEntries(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(5)
	health := &HealthComponent{Current: 100, Max: 100}
	feedback := NewVisualFeedbackComponent()
	entity.AddComponent(health)
	entity.AddComponent(feedback)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	if len(sys.prevHealth) != 1 {
		t.Fatalf("expected 1 tracked entity, got %d", len(sys.prevHealth))
	}

	// Remove entity and trigger cleanup
	sys.cleanupTimer = sys.cleanupInterval
	sys.Update([]*Entity{}, 0.016)

	if len(sys.prevHealth) != 0 {
		t.Errorf("expected 0 tracked entities after cleanup, got %d", len(sys.prevHealth))
	}
}

func TestDamageFlashTintSystem_ZeroDamage(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(6)
	health := &HealthComponent{Current: 100, Max: 100}
	feedback := NewVisualFeedbackComponent()
	entity.AddComponent(health)
	entity.AddComponent(feedback)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// No health change
	sys.Update(entities, 0.016)
	if feedback.IsFlashing() {
		t.Error("should not flash when health unchanged")
	}
}

func TestDamageFlashTintSystem_ZeroMaxHealth(t *testing.T) {
	sys := NewDamageFlashTintSystem(nil)

	entity := NewEntity(7)
	health := &HealthComponent{Current: 0, Max: 0}
	feedback := NewVisualFeedbackComponent()
	entity.AddComponent(health)
	entity.AddComponent(feedback)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Should not panic or flash with zero max health
	health.Current = -10
	sys.Update(entities, 0.016)
	// Division by zero guarded by Max > 0 check
}

func BenchmarkDamageFlashTintSystem(b *testing.B) {
	sys := NewDamageFlashTintSystem(nil)

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 100))
		e.AddComponent(&HealthComponent{Current: 100, Max: 100})
		e.AddComponent(NewVisualFeedbackComponent())
		entities[i] = e
	}

	// Prime
	sys.Update(entities, 0.016)

	// Apply damage to half the entities
	for i := 0; i < 100; i++ {
		h := entities[i].GetHealth()
		h.Current = 70
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
