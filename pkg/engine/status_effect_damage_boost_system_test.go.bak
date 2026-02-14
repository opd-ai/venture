package engine

import (
	"testing"
)

func TestNewStatusEffectDamageBoostSystem(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	if system == nil {
		t.Fatal("NewStatusEffectDamageBoostSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set")
	}
	if system.rng == nil {
		t.Error("rng not initialized")
	}
	if system.attackCache == nil {
		t.Error("attackCache not initialized")
	}
	if system.magicCache == nil {
		t.Error("magicCache not initialized")
	}
}

func TestStatusEffectDamageBoostSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		system.SetGenre(genre)
		if system.genre != genre {
			t.Errorf("SetGenre(%s) failed, got %s", genre, system.genre)
		}
	}
}

func TestStatusEffectDamageBoostSystem_EnragedBoost(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 50.0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "enraged",
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	stats := entity.GetStats()
	if stats == nil {
		t.Fatal("stats not found")
	}

	// Enraged: +20% physical, +5% magic
	expectedAttack := 100.0 * 1.20
	expectedMagic := 50.0 * 1.05

	if stats.Attack != expectedAttack {
		t.Errorf("Attack = %f, want %f", stats.Attack, expectedAttack)
	}
	if stats.MagicPower != expectedMagic {
		t.Errorf("MagicPower = %f, want %f", stats.MagicPower, expectedMagic)
	}
}

func TestStatusEffectDamageBoostSystem_WeaknessDebuff(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 50.0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "weakness",
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	stats := entity.GetStats()
	if stats == nil {
		t.Fatal("stats not found")
	}

	// Weakness: -20% physical, -10% magic
	expectedAttack := 100.0 * 0.80
	expectedMagic := 50.0 * 0.90

	if stats.Attack != expectedAttack {
		t.Errorf("Attack = %f, want %f", stats.Attack, expectedAttack)
	}
	if stats.MagicPower != expectedMagic {
		t.Errorf("MagicPower = %f, want %f", stats.MagicPower, expectedMagic)
	}
}

func TestStatusEffectDamageBoostSystem_StackingEffects(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 100.0})
	// Add multiple effects
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "enraged", // +20% attack, +5% magic
		Duration:   5.0,
	})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "strength", // +15% attack
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	stats := entity.GetStats()
	if stats == nil {
		t.Fatal("stats not found")
	}

	// Combined: +35% attack, +5% magic
	expectedAttack := 100.0 * 1.35
	expectedMagic := 100.0 * 1.05

	if stats.Attack != expectedAttack {
		t.Errorf("Attack = %f, want %f", stats.Attack, expectedAttack)
	}
	if stats.MagicPower != expectedMagic {
		t.Errorf("MagicPower = %f, want %f", stats.MagicPower, expectedMagic)
	}
}

func TestStatusEffectDamageBoostSystem_ExpiredEffectIgnored(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 50.0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "enraged",
		Duration:   0.0, // Expired
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	stats := entity.GetStats()
	if stats == nil {
		t.Fatal("stats not found")
	}

	// Expired effect should be ignored
	if stats.Attack != 100.0 {
		t.Errorf("Attack = %f, want 100.0 (no modification)", stats.Attack)
	}
	if stats.MagicPower != 50.0 {
		t.Errorf("MagicPower = %f, want 50.0 (no modification)", stats.MagicPower)
	}
}

func TestStatusEffectDamageBoostSystem_GenreScaling(t *testing.T) {
	tests := []struct {
		name          string
		genre         string
		effectType    string
		baseAttack    float64
		expectedScale float64 // Expected modifier multiplier for buffs
	}{
		{"fantasy_enraged", "fantasy", "enraged", 100.0, 1.0},
		{"scifi_enraged", "scifi", "enraged", 100.0, 1.1},          // +10% buff boost
		{"cyberpunk_enraged", "cyberpunk", "enraged", 100.0, 1.15}, // +15% buff boost
		{"horror_weakness", "horror", "weakness", 100.0, 1.25},     // +25% debuff boost
		{"postapoc_enraged", "postapoc", "enraged", 100.0, 1.1},    // +10% buff boost
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewStatusEffectDamageBoostSystem(world, 12345)
			system.SetGenre(tt.genre)

			entity := NewEntity()
			entity.AddComponent(&StatsComponent{Attack: tt.baseAttack, MagicPower: 50.0})
			entity.AddComponent(&StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   5.0,
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			system.Update(entities, 0.016)

			stats := entity.GetStats()
			if stats == nil {
				t.Fatal("stats not found")
			}

			// Verify attack was modified (detailed check depends on genre)
			if stats.Attack == tt.baseAttack {
				t.Errorf("Attack unchanged, expected modification for %s in %s genre", tt.effectType, tt.genre)
			}
		})
	}
}

func TestStatusEffectDamageBoostSystem_HasDamageEffect(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	// Entity with damage effect
	entity1 := NewEntity()
	entity1.AddComponent(&StatusEffectComponent{
		EffectType: "enraged",
		Duration:   5.0,
	})

	// Entity without damage effect
	entity2 := NewEntity()
	entity2.AddComponent(&StatusEffectComponent{
		EffectType: "regeneration", // No damage effect
		Duration:   5.0,
	})

	// Entity with no effects
	entity3 := NewEntity()

	if !system.HasDamageEffect(entity1) {
		t.Error("HasDamageEffect should return true for enraged")
	}
	if system.HasDamageEffect(entity2) {
		t.Error("HasDamageEffect should return false for regeneration")
	}
	if system.HasDamageEffect(entity3) {
		t.Error("HasDamageEffect should return false for entity with no effects")
	}
	if system.HasDamageEffect(nil) {
		t.Error("HasDamageEffect should return false for nil entity")
	}
}

func TestStatusEffectDamageBoostSystem_GetDamageModifiers(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 50.0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "blessed", // +10% both
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	attackMod, magicMod := system.GetDamageModifiers(entity.ID)

	if attackMod != 0.10 {
		t.Errorf("attackMod = %f, want 0.10", attackMod)
	}
	if magicMod != 0.10 {
		t.Errorf("magicMod = %f, want 0.10", magicMod)
	}
}

func TestStatusEffectDamageBoostSystem_GetEffectiveDamage(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 50.0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "enraged",
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	attack, magic := system.GetEffectiveDamage(entity)

	expectedAttack := 100.0 * 1.20
	expectedMagic := 50.0 * 1.05

	if attack != expectedAttack {
		t.Errorf("GetEffectiveDamage attack = %f, want %f", attack, expectedAttack)
	}
	if magic != expectedMagic {
		t.Errorf("GetEffectiveDamage magic = %f, want %f", magic, expectedMagic)
	}
}

func TestStatusEffectDamageBoostSystem_FrozenDebuff(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 100.0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen", // -25% both
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	stats := entity.GetStats()
	if stats == nil {
		t.Fatal("stats not found")
	}

	expectedAttack := 100.0 * 0.75
	expectedMagic := 100.0 * 0.75

	if stats.Attack != expectedAttack {
		t.Errorf("Attack = %f, want %f", stats.Attack, expectedAttack)
	}
	if stats.MagicPower != expectedMagic {
		t.Errorf("MagicPower = %f, want %f", stats.MagicPower, expectedMagic)
	}
}

func TestStatusEffectDamageBoostSystem_NoStatsEntity(t *testing.T) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	// Entity without StatsComponent
	entity := NewEntity()
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "enraged",
		Duration:   5.0,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	// Should not panic
	system.Update(entities, 0.016)

	// Verify no crash and no cache entry
	attackMod, magicMod := system.GetDamageModifiers(entity.ID)
	if attackMod != 0.0 || magicMod != 0.0 {
		t.Error("modifiers should be 0 for entity without stats")
	}
}

func TestStatusEffectDamageBoostSystem_AllEffectTypes(t *testing.T) {
	tests := []struct {
		effectType    string
		expectedAtk   float64 // Expected attack modifier
		expectedMagic float64 // Expected magic modifier
	}{
		{"enraged", 0.20, 0.05},
		{"empowered", 0.05, 0.15},
		{"berserk", 0.30, 0.10},
		{"blessed", 0.10, 0.10},
		{"cursed", -0.15, -0.15},
		{"weakness", -0.20, -0.10},
		{"strength", 0.15, 0.0},
		{"poisoned", -0.05, -0.05},
		{"burning", 0.05, 0.0},
		{"chilled", -0.10, -0.10},
		{"frozen", -0.25, -0.25},
		{"haste", 0.10, 0.10},
		{"fortify", 0.0, 0.0},
		{"vulnerability", 0.0, 0.0},
		{"regeneration", 0.0, 0.0},
		{"unknown_effect", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			world := NewWorld()
			system := NewStatusEffectDamageBoostSystem(world, 12345)

			entity := NewEntity()
			entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 100.0})
			entity.AddComponent(&StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   5.0,
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			system.Update(entities, 0.016)

			attackMod, magicMod := system.GetDamageModifiers(entity.ID)

			if attackMod != tt.expectedAtk {
				t.Errorf("attackMod = %f, want %f for %s", attackMod, tt.expectedAtk, tt.effectType)
			}
			if magicMod != tt.expectedMagic {
				t.Errorf("magicMod = %f, want %f for %s", magicMod, tt.expectedMagic, tt.effectType)
			}
		})
	}
}

func BenchmarkStatusEffectDamageBoostSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewStatusEffectDamageBoostSystem(world, 12345)

	// Create 100 entities with status effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity()
		entity.AddComponent(&StatsComponent{Attack: 100.0, MagicPower: 50.0})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: "enraged",
			Duration:   10.0,
		})
		world.AddEntity(entity)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
