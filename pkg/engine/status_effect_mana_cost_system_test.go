package engine

import (
	"math/rand"
	"testing"
)

func TestNewStatusEffectManaCostSystem(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectManaCostSystem(world, 12345)
	if system == nil {
		t.Fatal("NewStatusEffectManaCostSystem returned nil")
	}
	if system.genre != "fantasy" {
		t.Errorf("Default genre = %q, want 'fantasy'", system.genre)
	}
}

func TestStatusEffectManaCostSystem_EffectModifiers(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectManaCostSystem(world, 12345)

	tests := []struct {
		effectType     string
		wantMultiplier float64
	}{
		{"haste", 0.85},
		{"focused", 0.80},
		{"blessed", 0.90},
		{"chilled", 1.15},
		{"frozen", 1.30},
		{"cursed", 1.20},
		{"stunned", 1.25},
		{"fortify", 1.0},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		effect := &StatusEffectComponent{EffectType: tt.effectType, Duration: 5.0}
		got := system.effectToCostModifier(effect)
		if got != tt.wantMultiplier {
			t.Errorf("effectToCostModifier(%q) = %v, want %v", tt.effectType, got, tt.wantMultiplier)
		}
	}
}

func TestStatusEffectManaCostSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectManaCostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatusEffectComponent{EffectType: "haste", Duration: 5.0})

	system.Update([]*Entity{entity}, 0.016)
	multiplier := system.GetCostMultiplier(entity.ID)
	if multiplier >= 1.0 {
		t.Errorf("Expected multiplier < 1.0 for haste buff, got %v", multiplier)
	}
}

func TestStatusEffectManaCostSystem_GenreScaling(t *testing.T) {
	tests := []struct {
		genre, effectType string
		wantLower         bool
	}{
		{"fantasy", "haste", true},
		{"scifi", "haste", true},
		{"horror", "cursed", false},
		{"cyberpunk", "focused", true},
	}

	for _, tt := range tests {
		world := NewWorld()
		system := NewStatusEffectManaCostSystem(world, 12345)
		system.SetGenre(tt.genre)

		entity := world.CreateEntity()
		entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatusEffectComponent{EffectType: tt.effectType, Duration: 5.0})
		system.Update([]*Entity{entity}, 0.016)

		mult := system.GetCostMultiplier(entity.ID)
		if tt.wantLower && mult >= 1.0 {
			t.Errorf("Genre %q with %q: expected < 1.0, got %v", tt.genre, tt.effectType, mult)
		}
		if !tt.wantLower && mult <= 1.0 {
			t.Errorf("Genre %q with %q: expected > 1.0, got %v", tt.genre, tt.effectType, mult)
		}
	}
}

func TestStatusEffectManaCostSystem_GetEffectiveManaCost(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectManaCostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
	entity.AddComponent(&StatusEffectComponent{EffectType: "haste", Duration: 5.0})

	system.Update([]*Entity{entity}, 0.016)
	effectiveCost := system.GetEffectiveManaCost(entity.ID, 100)
	if effectiveCost >= 100 {
		t.Errorf("Effective cost %d should be less than 100 with haste", effectiveCost)
	}
	if minCost := system.GetEffectiveManaCost(entity.ID, 1); minCost < 1 {
		t.Errorf("Minimum cost should be 1, got %d", minCost)
	}
}

func TestStatusEffectManaCostSystem_HasCostEffect(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectManaCostSystem(world, 12345)

	e1 := world.CreateEntity()
	e1.AddComponent(&StatusEffectComponent{EffectType: "haste", Duration: 5.0})
	if !system.HasCostEffect(e1) {
		t.Error("HasCostEffect should return true for haste")
	}

	e2 := world.CreateEntity()
	e2.AddComponent(&StatusEffectComponent{EffectType: "fortify", Duration: 5.0})
	if system.HasCostEffect(e2) {
		t.Error("HasCostEffect should return false for fortify")
	}
}

func BenchmarkStatusEffectManaCostSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectManaCostSystem(world, 12345)
	effectTypes := []string{"haste", "poisoned", "blessed", "cursed"}
	rng := rand.New(rand.NewSource(12345))

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: effectTypes[rng.Intn(len(effectTypes))], Duration: 5.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
