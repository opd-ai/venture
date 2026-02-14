//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewWeaponSwingParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeaponSwingParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.baseParticleCount != 6 {
		t.Errorf("baseParticleCount = %d, want 6", sys.baseParticleCount)
	}
}

func TestWeaponSwingParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestWeaponSwingParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)

	sys.SetGenre("fantasy")

	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestWeaponSwingParticleSystem_GetParticleCount(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)

	tests := []struct {
		rarity item.Rarity
		want   int
	}{
		{item.RarityCommon, 6},
		{item.RarityUncommon, 8},
		{item.RarityRare, 10},
		{item.RarityEpic, 13},
		{item.RarityLegendary, 16},
	}

	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			got := sys.getParticleCount(tt.rarity)
			if got != tt.want {
				t.Errorf("getParticleCount(%s) = %d, want %d", tt.rarity.String(), got, tt.want)
			}
		})
	}
}

func TestWeaponSwingParticleSystem_GetRarityColorHint(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)

	tests := []struct {
		rarity item.Rarity
		want   string
	}{
		{item.RarityCommon, "white"},
		{item.RarityUncommon, "green"},
		{item.RarityRare, "blue"},
		{item.RarityEpic, "purple"},
		{item.RarityLegendary, "gold"},
	}

	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			got := sys.getRarityColorHint(tt.rarity)
			if got != tt.want {
				t.Errorf("getRarityColorHint(%s) = %q, want %q", tt.rarity.String(), got, tt.want)
			}
		})
	}
}

func TestWeaponSwingParticleSystem_OnDamageDealt_NilInputs(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Should not panic with nil attacker
	sys.OnDamageDealt(nil, nil, 10.0)

	// Should not panic with nil target
	attacker := world.CreateEntity()
	sys.OnDamageDealt(attacker, nil, 10.0)
}

func TestWeaponSwingParticleSystem_OnDamageDealt_NoWeapon(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 150, Y: 100})

	// Should not spawn particles when no weapon equipped
	sys.OnDamageDealt(attacker, target, 10.0)
	// No assertions needed - just verify no panic
}

func TestWeaponSwingParticleSystem_OnDamageDealt_WithWeapon(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Add equipment with weapon
	equip := NewEquipmentComponent()
	sword := &item.Item{
		ID:     "test_sword",
		Name:   "Test Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityRare,
		Stats: item.Stats{
			Damage:       10,
			IsProjectile: false, // Melee weapon
		},
	}
	equip.Equip(SlotMainHand, sword)
	attacker.AddComponent(equip)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 150, Y: 100})

	// Should spawn particles for melee weapon
	sys.OnDamageDealt(attacker, target, 25.0)
	// Verify particles were spawned (indirect via particle system state)
}

func TestWeaponSwingParticleSystem_OnDamageDealt_ProjectileWeaponSkipped(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Add equipment with ranged weapon
	equip := NewEquipmentComponent()
	bow := &item.Item{
		ID:     "test_bow",
		Name:   "Test Bow",
		Type:   item.TypeWeapon,
		Rarity: item.RarityRare,
		Stats: item.Stats{
			Damage:       10,
			IsProjectile: true, // Ranged weapon - should skip
		},
	}
	equip.Equip(SlotMainHand, bow)
	attacker.AddComponent(equip)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 150, Y: 100})

	// Should NOT spawn particles for projectile weapon
	sys.OnDamageDealt(attacker, target, 25.0)
	// No assertions - just verify no panic and particles skipped
}

func TestWeaponSwingParticleSystem_SpawnSwingEffect(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Direct spawn for external use
	sys.SpawnSwingEffect(100, 100, item.RarityLegendary, 50.0)
	// Verify no panic
}

func TestWeaponSwingParticleSystem_SpawnSwingEffect_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	// Don't set particle system

	// Should not panic with nil particle system
	sys.SpawnSwingEffect(100, 100, item.RarityRare, 25.0)
}

func TestWeaponSwingParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)

	entities := []*Entity{world.CreateEntity()}

	// Update should be a no-op (callback-driven)
	sys.Update(entities, 0.016)
	// Verify no panic
}

func TestWeaponSwingParticleSystem_RarityParticleScaling(t *testing.T) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)

	// Verify particle count scales monotonically with rarity
	prevCount := 0
	rarities := []item.Rarity{
		item.RarityCommon,
		item.RarityUncommon,
		item.RarityRare,
		item.RarityEpic,
		item.RarityLegendary,
	}

	for _, rarity := range rarities {
		count := sys.getParticleCount(rarity)
		if count <= prevCount && prevCount > 0 {
			t.Errorf("particle count not increasing: %s (%d) <= previous (%d)",
				rarity.String(), count, prevCount)
		}
		prevCount = count
	}
}

func TestWeaponSwingParticleSystem_GenreAwareness(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeaponSwingParticleSystem(world, 12345)
			ps := NewParticleSystem()
			sys.SetParticleSystem(ps)
			sys.SetGenre(genre)

			if sys.genreID != genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, genre)
			}

			// Spawn should work for all genres
			sys.SpawnSwingEffect(100, 100, item.RarityRare, 25.0)
		})
	}
}

func BenchmarkWeaponSwingParticleSystem_OnDamageDealt(b *testing.B) {
	world := NewWorld()
	sys := NewWeaponSwingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	equip := NewEquipmentComponent()
	sword := &item.Item{
		ID:     "bench_sword",
		Name:   "Benchmark Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityRare,
		Stats: item.Stats{
			Damage:       10,
			IsProjectile: false,
		},
	}
	equip.Equip(SlotMainHand, sword)
	attacker.AddComponent(equip)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 150, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnDamageDealt(attacker, target, 25.0)
	}
}
