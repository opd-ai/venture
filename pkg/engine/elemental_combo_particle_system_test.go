package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewElementalComboParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	sys := NewElementalComboParticleSystem(world, seed)
	if sys == nil {
		t.Fatal("NewElementalComboParticleSystem returned nil")
	}

	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != seed {
		t.Errorf("seed = %d, want %d", sys.seed, seed)
	}
	if sys.comboCooldowns == nil {
		t.Error("comboCooldowns map not initialized")
	}
	if sys.cooldownTime <= 0 {
		t.Error("cooldownTime should be positive")
	}
	if sys.baseParticleCount <= 0 {
		t.Error("baseParticleCount should be positive")
	}
}

func TestElementalComboParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestElementalComboParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("genre = %s, want %s", sys.genreID, genre)
		}
	}
}

func TestElementalComboParticleSystem_IsElementalEffect(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	tests := []struct {
		effect   string
		expected bool
	}{
		{"burning", true},
		{"frozen", true},
		{"shocked", true},
		{"poisoned", true},
		{"wet", true},
		{"chilled", true},
		{"strength", false},
		{"weakness", false},
		{"regeneration", false},
		{"random_effect", false},
	}

	for _, tt := range tests {
		t.Run(tt.effect, func(t *testing.T) {
			result := sys.isElementalEffect(tt.effect)
			if result != tt.expected {
				t.Errorf("isElementalEffect(%s) = %v, want %v", tt.effect, result, tt.expected)
			}
		})
	}
}

func TestElementalComboParticleSystem_FindCombo(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

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
		{"single element", "burning", "burning", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combo := sys.findCombo(tt.effect1, tt.effect2)
			if tt.wantCombo {
				if combo == nil {
					t.Errorf("expected combo for %s+%s, got nil", tt.effect1, tt.effect2)
				} else if combo.ComboName != tt.comboName {
					t.Errorf("combo name = %s, want %s", combo.ComboName, tt.comboName)
				}
			} else {
				if combo != nil {
					t.Errorf("expected no combo for %s+%s, got %s", tt.effect1, tt.effect2, combo.ComboName)
				}
			}
		})
	}
}

func TestElementalComboParticleSystem_DetectElementalCombo(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	tests := []struct {
		name      string
		effects   []string
		wantCombo bool
	}{
		{"no effects", []string{}, false},
		{"single effect", []string{"burning"}, false},
		{"fire+ice combo", []string{"burning", "frozen"}, true},
		{"three effects", []string{"burning", "frozen", "poisoned"}, true},
		{"non-elemental", []string{"strength", "regeneration"}, false},
		{"mixed effects", []string{"burning", "strength"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(uint64(100))
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})

			// Add status effects using set component for multiple concurrent effects
			if len(tt.effects) > 0 {
				effectSet := &StatusEffectSetComponent{}
				for _, effectType := range tt.effects {
					effectSet.AddEffect(&StatusEffectComponent{
						EffectType: effectType,
						Duration:   5.0,
						Magnitude:  10.0,
					})
				}
				entity.AddComponent(effectSet)
			}

			combo := sys.detectElementalCombo(entity)
			if tt.wantCombo && combo == nil {
				t.Error("expected combo, got nil")
			}
			if !tt.wantCombo && combo != nil {
				t.Errorf("expected no combo, got %s", combo.ComboName)
			}
		})
	}
}

func TestElementalComboParticleSystem_Cooldowns(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)
	sys.cooldownTime = 1.0

	entityID := uint64(42)

	// Initially not on cooldown
	if sys.isOnCooldown(entityID) {
		t.Error("entity should not start on cooldown")
	}

	// Add cooldown
	sys.comboCooldowns[entityID] = 1.0
	if !sys.isOnCooldown(entityID) {
		t.Error("entity should be on cooldown after adding")
	}

	// Partial update
	sys.updateCooldowns(0.5)
	if !sys.isOnCooldown(entityID) {
		t.Error("entity should still be on cooldown after partial update")
	}

	// Full update to expire
	sys.updateCooldowns(0.6)
	if sys.isOnCooldown(entityID) {
		t.Error("entity should not be on cooldown after expiry")
	}
}

func TestElementalComboParticleSystem_GetComboParticleTypes(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	combos := []struct {
		comboName string
		primary   string
		secondary string
	}{
		{"steam_burst", "burning", "frozen"},
		{"toxic_flames", "burning", "poisoned"},
		{"shatter", "frozen", "shocked"},
		{"toxic_pool", "poisoned", "wet"},
		{"plasma_burst", "burning", "shocked"},
		{"deep_freeze", "frozen", "wet"},
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, combo := range combos {
		for _, genre := range genres {
			t.Run(combo.comboName+"_"+genre, func(t *testing.T) {
				sys.SetGenre(genre)
				ec := &ElementalCombo{
					Primary:   combo.primary,
					Secondary: combo.secondary,
					ComboName: combo.comboName,
				}
				primary, secondary := sys.getComboParticleTypes(ec)

				// Verify we get valid particle types
				if primary.String() == "unknown" {
					t.Errorf("invalid primary particle type for %s/%s", combo.comboName, genre)
				}
				if secondary.String() == "unknown" {
					t.Errorf("invalid secondary particle type for %s/%s", combo.comboName, genre)
				}
			})
		}
	}
}

func TestElementalComboParticleSystem_GetComboGravity(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	tests := []struct {
		comboName string
		wantSign  int // -1 for negative (rising), 0 for zero, 1 for positive (falling)
	}{
		{"steam_burst", -1},
		{"toxic_flames", -1},
		{"shatter", 1},
		{"toxic_pool", 1},
		{"plasma_burst", -1},
		{"deep_freeze", 1},
		{"unknown_combo", 0},
	}

	for _, tt := range tests {
		t.Run(tt.comboName, func(t *testing.T) {
			combo := &ElementalCombo{ComboName: tt.comboName}
			gravity := sys.getComboGravity(combo)

			switch tt.wantSign {
			case -1:
				if gravity >= 0 {
					t.Errorf("expected negative gravity for %s, got %f", tt.comboName, gravity)
				}
			case 1:
				if gravity <= 0 {
					t.Errorf("expected positive gravity for %s, got %f", tt.comboName, gravity)
				}
			case 0:
				if gravity != 0 {
					t.Errorf("expected zero gravity for %s, got %f", tt.comboName, gravity)
				}
			}
		})
	}
}

func TestElementalComboParticleSystem_Update_WithCombo(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Create entity with fire+ice combo using set component
	entity := NewEntity(uint64(1))
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	effectSet := &StatusEffectSetComponent{}
	effectSet.AddEffect(&StatusEffectComponent{
		EffectType: "burning",
		Duration:   5.0,
		Magnitude:  10.0,
	})
	effectSet.AddEffect(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   5.0,
		Magnitude:  10.0,
	})
	entity.AddComponent(effectSet)

	entities := []*Entity{entity}

	// First update should spawn particles and start cooldown
	sys.Update(entities, 0.016)

	if !sys.isOnCooldown(entity.ID) {
		t.Error("entity should be on cooldown after combo trigger")
	}

	// Second update during cooldown should not trigger again
	sys.Update(entities, 0.016)

	// Verify entity is still on cooldown (cooldown lasts 1 second)
	if !sys.isOnCooldown(entity.ID) {
		t.Error("entity should still be on cooldown after short update")
	}
}

func TestElementalComboParticleSystem_Update_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)
	// Don't set particle system

	entity := NewEntity(uint64(2))
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatusEffectComponent{EffectType: "burning", Duration: 5.0})
	entity.AddComponent(&StatusEffectComponent{EffectType: "frozen", Duration: 5.0})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestElementalComboParticleSystem_Update_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Create entity without position
	entity := NewEntity(uint64(3))
	entity.AddComponent(&StatusEffectComponent{EffectType: "burning", Duration: 5.0})
	entity.AddComponent(&StatusEffectComponent{EffectType: "frozen", Duration: 5.0})

	// Should not panic, just skip entity
	sys.Update([]*Entity{entity}, 0.016)

	// Should not be on cooldown since no particles spawned
	if sys.isOnCooldown(entity.ID) {
		t.Error("entity without position should not trigger cooldown")
	}
}

func TestElementalComboParticleSystem_Determinism(t *testing.T) {
	seed := int64(99999)

	// Create two identical systems
	world1 := NewWorld()
	sys1 := NewElementalComboParticleSystem(world1, seed)

	world2 := NewWorld()
	sys2 := NewElementalComboParticleSystem(world2, seed)

	// Should have same initial state
	if sys1.seed != sys2.seed {
		t.Error("seeds should match")
	}
	if sys1.cooldownTime != sys2.cooldownTime {
		t.Error("cooldown times should match")
	}
	if sys1.baseParticleCount != sys2.baseParticleCount {
		t.Error("particle counts should match")
	}
}

func TestElementalCombo_Struct(t *testing.T) {
	combo := &ElementalCombo{
		Primary:   "burning",
		Secondary: "frozen",
		ComboName: "steam_burst",
	}

	if combo.Primary != "burning" {
		t.Errorf("Primary = %s, want burning", combo.Primary)
	}
	if combo.Secondary != "frozen" {
		t.Errorf("Secondary = %s, want frozen", combo.Secondary)
	}
	if combo.ComboName != "steam_burst" {
		t.Errorf("ComboName = %s, want steam_burst", combo.ComboName)
	}
}

func TestElementalCombos_AllDefinitionsValid(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	// Verify all defined combos have elemental effects
	for _, def := range elementalCombos {
		if !sys.isElementalEffect(def.element1) {
			t.Errorf("element1 %s is not recognized as elemental", def.element1)
		}
		if !sys.isElementalEffect(def.element2) {
			t.Errorf("element2 %s is not recognized as elemental", def.element2)
		}
		if def.comboName == "" {
			t.Errorf("combo between %s and %s has empty name", def.element1, def.element2)
		}
	}
}

func TestElementalComboParticleSystem_AllGenreParticleTypes(t *testing.T) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	comboNames := []string{"steam_burst", "toxic_flames", "shatter", "toxic_pool", "plasma_burst", "deep_freeze"}

	for _, genre := range genres {
		sys.SetGenre(genre)

		for _, comboName := range comboNames {
			t.Run(genre+"_"+comboName, func(t *testing.T) {
				combo := &ElementalCombo{ComboName: comboName}
				primary, secondary := sys.getComboParticleTypes(combo)

				// Verify both types are valid
				validTypes := map[particles.ParticleType]bool{
					particles.ParticleSpark:      true,
					particles.ParticleSmoke:      true,
					particles.ParticleMagic:      true,
					particles.ParticleFlame:      true,
					particles.ParticleBlood:      true,
					particles.ParticleDust:       true,
					particles.ParticleEmber:      true,
					particles.ParticleSparkle:    true,
					particles.ParticleSmokePlume: true,
					particles.ParticleDebris:     true,
				}

				if !validTypes[primary] {
					t.Errorf("invalid primary particle type %v for %s/%s", primary, genre, comboName)
				}
				if !validTypes[secondary] {
					t.Errorf("invalid secondary particle type %v for %s/%s", secondary, genre, comboName)
				}
			})
		}
	}
}

// BenchmarkElementalComboDetection benchmarks combo detection performance.
func BenchmarkElementalComboDetection(b *testing.B) {
	world := NewWorld()
	sys := NewElementalComboParticleSystem(world, 12345)

	entity := NewEntity(uint64(100))
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatusEffectComponent{EffectType: "burning", Duration: 5.0})
	entity.AddComponent(&StatusEffectComponent{EffectType: "frozen", Duration: 5.0})
	entity.AddComponent(&StatusEffectComponent{EffectType: "poisoned", Duration: 5.0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.detectElementalCombo(entity)
	}
}
