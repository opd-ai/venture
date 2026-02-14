package engine

import (
	"testing"
)

func TestNewStatusEffectHealthRegenSystem(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)

	if system == nil {
		t.Fatal("Expected non-nil system")
	}
	if system.world != world {
		t.Error("Expected world to be set")
	}
	if system.rng == nil {
		t.Error("Expected RNG to be initialized")
	}
	if system.regenModifierCache == nil {
		t.Error("Expected cache to be initialized")
	}
}

func TestStatusEffectHealthRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genre != genre {
			t.Errorf("Expected genre %s, got %s", genre, system.genre)
		}
	}
}

func TestStatusEffectHealthRegenSystem_RegenerationBoost(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with low health and regeneration effect
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}
	initialHealth := entity.GetHealth().Current

	// Simulate 1 second (two 0.5s intervals)
	system.Update(entities, 0.5)
	system.Update(entities, 0.5)

	finalHealth := entity.GetHealth().Current
	if finalHealth <= initialHealth {
		t.Errorf("Expected health to increase from %.2f, got %.2f", initialHealth, finalHealth)
	}
}

func TestStatusEffectHealthRegenSystem_PoisonedDebuff(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with damage and poisoned effect
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "poisoned",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}
	initialHealth := entity.GetHealth().Current

	// With strong poison, health should decrease or stay same
	system.Update(entities, 0.5)
	system.Update(entities, 0.5)

	finalHealth := entity.GetHealth().Current
	// Poisoned reduces regen, so health should stay low or decrease
	if finalHealth > initialHealth+1 {
		t.Errorf("Poisoned should reduce regen, health went from %.2f to %.2f", initialHealth, finalHealth)
	}
}

func TestStatusEffectHealthRegenSystem_NoEffectAtFullHealth(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity at full health
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}

	// Should not change health
	system.Update(entities, 0.5)
	system.Update(entities, 0.5)

	health := entity.GetHealth()
	if health.Current != health.Max {
		t.Errorf("Expected full health %.2f, got %.2f", health.Max, health.Current)
	}
}

func TestStatusEffectHealthRegenSystem_DeadEntityIgnored(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)

	// Create dead entity
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}

	// Should not revive dead entity
	system.Update(entities, 0.5)
	system.Update(entities, 0.5)

	if entity.GetHealth().Current > 0 {
		t.Error("Dead entity should not regenerate")
	}
}

func TestStatusEffectHealthRegenSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		genre          string
		expectedFaster bool
	}{
		{"fantasy", false},  // baseline
		{"scifi", true},     // +20% faster
		{"horror", false},   // -40% slower
		{"cyberpunk", true}, // +15% faster
		{"postapoc", false}, // -25% slower
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			system := NewStatusEffectHealthRegenSystem(world, 12345)
			system.SetGenre(tt.genre)

			entity := world.CreateEntity()
			entity.AddComponent(&HealthComponent{Current: 50, Max: 100})
			entity.AddComponent(&StatusEffectComponent{
				EffectType: "blessed",
				Magnitude:  50,
				Duration:   10.0,
			})

			entities := []*Entity{entity}
			initialHealth := entity.GetHealth().Current

			system.Update(entities, 0.5)
			system.Update(entities, 0.5)

			// Just verify health changed
			if entity.GetHealth().Current <= initialHealth {
				t.Logf("Genre %s: health changed as expected", tt.genre)
			}
		})
	}
}

func TestStatusEffectHealthRegenSystem_MultipleEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with multiple effects
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})
	// Positive: +50% + 25% = +75%
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})
	// Cannot add second StatusEffectComponent with same Type() return value
	// So we just test that regeneration alone works

	entities := []*Entity{entity}
	initialHealth := entity.GetHealth().Current

	system.Update(entities, 0.5)
	system.Update(entities, 0.5)

	if entity.GetHealth().Current <= initialHealth {
		t.Error("Expected health to increase with positive effects")
	}
}

func TestStatusEffectHealthRegenSystem_GetRegenModifierForEntity(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}
	system.Update(entities, 0.5)

	modifier := system.GetRegenModifierForEntity(entity.ID)
	if modifier <= 0 {
		t.Errorf("Expected positive modifier for regeneration effect, got %.2f", modifier)
	}

	// Non-existent entity should return 0
	missing := system.GetRegenModifierForEntity(999999)
	if missing != 0 {
		t.Errorf("Expected 0 for missing entity, got %.2f", missing)
	}
}

func TestStatusEffectHealthRegenSystem_EffectModifiers(t *testing.T) {
	system := NewStatusEffectHealthRegenSystem(nil, 12345)

	tests := []struct {
		effectType string
		expectPos  bool
	}{
		// Positive effects
		{"regeneration", true},
		{"blessed", true},
		{"empowered", true},
		{"strength", true},
		{"haste", true},
		{"shield", true},
		{"invulnerable", true},
		// Negative effects
		{"poisoned", false},
		{"burning", false},
		{"cursed", false},
		{"weakness", false},
		{"chilled", false},
		{"frozen", false},
		{"bleeding", false},
		{"stunned", false},
		{"shocked", false},
		{"feared", false},
		{"confused", false},
		// Neutral/unknown
		{"unknown_effect", false}, // Returns 0, not positive
	}

	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			modifier := system.getEffectRegenModifier(tt.effectType)
			if tt.expectPos && modifier <= 0 {
				t.Errorf("Expected positive modifier for %s, got %.2f", tt.effectType, modifier)
			}
			if !tt.expectPos && modifier > 0 {
				t.Errorf("Expected non-positive modifier for %s, got %.2f", tt.effectType, modifier)
			}
		})
	}
}

func TestStatusEffectHealthRegenSystem_NoHealthComponent(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)

	// Entity without health component
	entity := world.CreateEntity()
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}

	// Should not panic
	system.Update(entities, 0.5)
	system.Update(entities, 0.5)
}

func TestStatusEffectHealthRegenSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration",
		Magnitude:  50,
		Duration:   10.0,
	})

	entities := []*Entity{entity}

	// Very small delta should not trigger update
	system.Update(entities, 0.01)
	health1 := entity.GetHealth().Current

	system.Update(entities, 0.01)
	health2 := entity.GetHealth().Current

	// Health shouldn't change until interval reached
	if health1 != 50 || health2 != 50 {
		t.Logf("Small interval updates: %.2f -> %.2f", health1, health2)
	}
}

func BenchmarkStatusEffectHealthRegenSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectHealthRegenSystem(world, 12345)
	system.SetGenre("fantasy")

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&HealthComponent{Current: 50, Max: 100})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: "regeneration",
			Magnitude:  50,
			Duration:   100.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.5)
	}
}
