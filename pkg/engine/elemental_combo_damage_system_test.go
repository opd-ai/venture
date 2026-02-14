package engine

import (
	"math/rand"
	"testing"
)

func TestNewElementalComboDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewElementalComboDamageSystem returned nil")
	}

	if sys.baseDamage != 15.0 {
		t.Errorf("Expected base damage 15.0, got %f", sys.baseDamage)
	}

	if sys.cooldownTime != 2.0 {
		t.Errorf("Expected cooldown time 2.0, got %f", sys.cooldownTime)
	}

	if sys.world != world {
		t.Error("World reference not set correctly")
	}
}

func TestElementalComboDamageSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	tests := []struct {
		name    string
		genre   string
		wantMul float64
	}{
		{"fantasy", "fantasy", 1.0},
		{"scifi", "scifi", 1.1},
		{"horror", "horror", 1.3},
		{"cyberpunk", "cyberpunk", 1.2},
		{"postapoc", "postapoc", 0.9},
		{"unknown", "unknown", 1.0}, // Defaults to 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("SetGenre() did not set genre, want %s, got %s", tt.genre, sys.genreID)
			}
			gotMul := sys.getGenreMultiplier()
			if gotMul != tt.wantMul {
				t.Errorf("Genre multiplier = %f, want %f", gotMul, tt.wantMul)
			}
		})
	}
}

func TestElementalComboDamageSystem_IsElementalEffect(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	tests := []struct {
		effectType string
		want       bool
	}{
		{"burning", true},
		{"frozen", true},
		{"shocked", true},
		{"poisoned", true},
		{"wet", true},
		{"chilled", true},
		{"stunned", false},
		{"blinded", false},
		{"slowed", false},
		{"hasted", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			got := sys.isElementalEffect(tt.effectType)
			if got != tt.want {
				t.Errorf("isElementalEffect(%q) = %v, want %v", tt.effectType, got, tt.want)
			}
		})
	}
}

func TestElementalComboDamageSystem_FindCombo(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	tests := []struct {
		name      string
		effect1   string
		effect2   string
		wantCombo bool
		comboName string
	}{
		{"fire+ice", "burning", "frozen", true, "steam_burst"},
		{"ice+fire", "frozen", "burning", true, "steam_burst"},
		{"fire+wet", "burning", "wet", true, "steam_burst"},
		{"fire+poison", "burning", "poisoned", true, "toxic_flames"},
		{"ice+shock", "frozen", "shocked", true, "shatter"},
		{"poison+wet", "poisoned", "wet", true, "toxic_pool"},
		{"fire+shock", "burning", "shocked", true, "plasma_burst"},
		{"ice+wet", "frozen", "wet", true, "deep_freeze"},
		{"no combo", "strength", "weakness", false, ""},
		{"same element", "burning", "burning", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combo := sys.findCombo(tt.effect1, tt.effect2)
			if tt.wantCombo {
				if combo == nil {
					t.Errorf("expected combo for %s+%s, got nil", tt.effect1, tt.effect2)
				} else if combo.comboName != tt.comboName {
					t.Errorf("combo name = %s, want %s", combo.comboName, tt.comboName)
				}
			} else {
				if combo != nil {
					t.Errorf("expected no combo for %s+%s, got %s", tt.effect1, tt.effect2, combo.comboName)
				}
			}
		})
	}
}

func TestElementalComboDamageSystem_ApplyDamage(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	world.AddEntity(entity)

	// Test applying damage directly via applyElementalComboDamage
	combo := sys.findCombo("burning", "frozen") // steam_burst, baseMult 1.5
	if combo == nil {
		t.Fatal("Failed to find combo for burning+frozen")
	}

	sys.applyElementalComboDamage(entity, combo)

	health := entity.GetHealth()
	if health.Current >= 100 {
		t.Errorf("Expected health to decrease, got %f", health.Current)
	}

	// Damage should be approximately baseDamage * combo.baseMult * genreMult (with variance)
	// steam_burst has baseMult 1.5, fantasy has genreMult 1.0
	// Expected damage: 15 * 1.5 * 1.0 * (0.85-1.15) = ~19.1 to ~25.9
	expectedMin := 15 * 1.5 * 1.0 * 0.85
	expectedMax := 15 * 1.5 * 1.0 * 1.15
	actualDamage := 100 - health.Current

	if actualDamage < expectedMin || actualDamage > expectedMax {
		t.Errorf("Damage %f outside expected range [%f, %f]", actualDamage, expectedMin, expectedMax)
	}
}

func TestElementalComboDamageSystem_Cooldown(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	// Test cooldown tracking
	if sys.isOnCooldown(1) {
		t.Error("Entity should not be on cooldown initially")
	}

	// Start cooldown
	sys.comboCooldowns[1] = 2.0

	if !sys.isOnCooldown(1) {
		t.Error("Entity should be on cooldown after adding")
	}

	// Update cooldowns
	sys.updateCooldowns(1.0) // 1 second passes
	if !sys.isOnCooldown(1) {
		t.Error("Entity should still be on cooldown after 1s")
	}

	if sys.comboCooldowns[1] != 1.0 {
		t.Errorf("Cooldown remaining = %f, want 1.0", sys.comboCooldowns[1])
	}

	// Update cooldowns to expire
	sys.updateCooldowns(1.5) // 1.5 more seconds passes

	if sys.isOnCooldown(1) {
		t.Error("Entity should not be on cooldown after expiration")
	}
}

func TestElementalComboDamageSystem_GenreMultipliers(t *testing.T) {
	world := NewWorld()
	seed := int64(54321)

	tests := []struct {
		genre        string
		expectedMult float64
	}{
		{"horror", 1.3}, // Horror has highest multiplier
		{"cyberpunk", 1.2},
		{"scifi", 1.1},
		{"fantasy", 1.0},
		{"postapoc", 0.9}, // Post-apoc has lowest
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewElementalComboDamageSystem(world, seed)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

			// Reset RNG for consistent variance
			sys.rng = rand.New(rand.NewSource(seed))

			// shatter combo has baseMult 2.0
			combo := sys.findCombo("frozen", "shocked")
			if combo == nil {
				t.Fatal("Failed to find shatter combo")
			}

			sys.applyElementalComboDamage(entity, combo)

			// Damage = 15 * 2.0 * genreMult * variance
			health := entity.GetHealth()
			damage := 100 - health.Current

			// Base damage without variance: 15 * 2.0 * multiplier = 30 * multiplier
			baseExpected := 30 * tt.expectedMult

			// With ±15% variance, allow ±20% tolerance for comparison
			minExpected := baseExpected * 0.80
			maxExpected := baseExpected * 1.20

			if damage < minExpected || damage > maxExpected {
				t.Errorf("Genre %s: damage %f outside expected range [%f, %f]", tt.genre, damage, minExpected, maxExpected)
			}
		})
	}
}

func TestElementalComboDamageSystem_DeadEntity(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100}) // Dead entity
	world.AddEntity(entity)

	// Should not apply damage to dead entity
	sys.Update([]*Entity{entity}, 0.016)

	health := entity.GetHealth()
	if health.Current != 0 {
		t.Errorf("Dead entity health changed: expected 0, got %f", health.Current)
	}
}

func TestElementalComboDamageSystem_SettersGetters(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	// Test base damage
	if sys.GetBaseDamage() != 15.0 {
		t.Errorf("GetBaseDamage() = %f, want 15.0", sys.GetBaseDamage())
	}

	sys.SetBaseDamage(25.0)
	if sys.GetBaseDamage() != 25.0 {
		t.Errorf("After SetBaseDamage(25.0), GetBaseDamage() = %f", sys.GetBaseDamage())
	}

	// Invalid base damage should not change
	sys.SetBaseDamage(-5.0)
	if sys.GetBaseDamage() != 25.0 {
		t.Errorf("Negative SetBaseDamage changed value: %f", sys.GetBaseDamage())
	}

	// Test cooldown time
	if sys.GetCooldownTime() != 2.0 {
		t.Errorf("GetCooldownTime() = %f, want 2.0", sys.GetCooldownTime())
	}

	sys.SetCooldownTime(3.0)
	if sys.GetCooldownTime() != 3.0 {
		t.Errorf("After SetCooldownTime(3.0), GetCooldownTime() = %f", sys.GetCooldownTime())
	}

	// Invalid cooldown should not change
	sys.SetCooldownTime(-1.0)
	if sys.GetCooldownTime() != 3.0 {
		t.Errorf("Negative SetCooldownTime changed value: %f", sys.GetCooldownTime())
	}
}

func TestElementalComboDamageSystem_AllCombinations(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	// Test all defined combos via findCombo (the working path)
	for _, combo := range comboDamageDefinitions {
		t.Run(combo.comboName, func(t *testing.T) {
			detected := sys.findCombo(combo.element1, combo.element2)
			if detected == nil {
				t.Errorf("Failed to find combo %s (%s + %s)", combo.comboName, combo.element1, combo.element2)
			} else if detected.comboName != combo.comboName {
				t.Errorf("Detected wrong combo: got %s, want %s", detected.comboName, combo.comboName)
			}

			// Also test reverse order
			detectedReverse := sys.findCombo(combo.element2, combo.element1)
			if detectedReverse == nil {
				t.Errorf("Failed to find combo %s in reverse order (%s + %s)", combo.comboName, combo.element2, combo.element1)
			}
		})
	}
}

func TestElementalComboDamageSystem_ComboDamageTypes(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	// Test that each combo has correct damage type and multiplier
	tests := []struct {
		effect1    string
		effect2    string
		damageType string
		baseMult   float64
	}{
		{"burning", "frozen", "physical", 1.5},
		{"burning", "wet", "fire", 1.3},
		{"burning", "poisoned", "fire", 1.8},
		{"frozen", "shocked", "physical", 2.0},
		{"poisoned", "wet", "poison", 1.4},
		{"burning", "shocked", "electric", 1.7},
		{"frozen", "wet", "ice", 1.6},
	}

	for _, tt := range tests {
		t.Run(tt.effect1+"+"+tt.effect2, func(t *testing.T) {
			combo := sys.findCombo(tt.effect1, tt.effect2)
			if combo == nil {
				t.Fatalf("Failed to find combo for %s+%s", tt.effect1, tt.effect2)
			}

			if combo.damageType != tt.damageType {
				t.Errorf("Damage type = %s, want %s", combo.damageType, tt.damageType)
			}

			if combo.baseMult != tt.baseMult {
				t.Errorf("Base mult = %f, want %f", combo.baseMult, tt.baseMult)
			}
		})
	}
}

func TestElementalComboDamageSystem_NilWorld(t *testing.T) {
	sys := NewElementalComboDamageSystem(nil, 12345)

	// Should not panic with nil world
	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Should safely return without crashing
	sys.Update([]*Entity{entity}, 0.016)
}

func TestElementalComboDamageSystem_NoHealthComponent(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	entity := NewEntity(1)
	// No health component
	world.AddEntity(entity)

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func BenchmarkElementalComboDamageSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create entities with health (no status effects since they can't coexist properly)
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entities[i] = entity
		world.AddEntity(entity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkElementalComboDamageSystem_FindCombo(b *testing.B) {
	world := NewWorld()
	sys := NewElementalComboDamageSystem(world, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.findCombo("burning", "frozen")
		sys.findCombo("frozen", "shocked")
		sys.findCombo("poisoned", "wet")
	}
}
