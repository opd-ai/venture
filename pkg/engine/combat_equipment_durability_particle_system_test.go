package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewCombatEquipmentDurabilityParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewCombatEquipmentDurabilityParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.lastStateCache == nil {
		t.Error("lastStateCache not initialized")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)

	sys.SetGenre("sci-fi")

	if sys.genreID != "sci-fi" {
		t.Errorf("genreID = %s, want sci-fi", sys.genreID)
	}
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_NilTarget(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	// Should not panic
	sys.OnDamageTaken(nil, nil, 100)
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_ZeroDamage(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)

	entity := world.CreateEntity()
	equipComp := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	sys.OnDamageTaken(nil, entity, 0)

	// Durability should be unchanged
	if equipComp.Slots[SlotChest].Stats.Durability != 100 {
		t.Errorf("durability = %d, want 100", equipComp.Slots[SlotChest].Stats.Durability)
	}
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_NoEquipment(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()

	// Should not panic
	sys.OnDamageTaken(nil, entity, 100)
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_ReducesDurability(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	equipComp := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	attacker := world.CreateEntity()

	// Apply 100 damage (should lose 2 durability base)
	sys.OnDamageTaken(attacker, entity, 100)

	if equipComp.Slots[SlotChest].Stats.Durability >= 100 {
		t.Errorf("durability should have decreased from 100, got %d", equipComp.Slots[SlotChest].Stats.Durability)
	}
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_MultipleArmorPieces(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	equipComp := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{Durability: 100, DurabilityMax: 100},
	}
	equipComp.Slots[SlotHelmet] = &item.Item{
		Stats: item.ItemStats{Durability: 100, DurabilityMax: 100},
	}
	equipComp.Slots[SlotBoots] = &item.Item{
		Stats: item.ItemStats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	attacker := world.CreateEntity()

	// Apply 300 damage across 3 armor pieces
	sys.OnDamageTaken(attacker, entity, 300)

	// Each piece should have lost some durability
	if equipComp.Slots[SlotChest].Stats.Durability == 100 {
		t.Error("chest durability should have decreased")
	}
	if equipComp.Slots[SlotHelmet].Stats.Durability == 100 {
		t.Error("helmet durability should have decreased")
	}
	if equipComp.Slots[SlotBoots].Stats.Durability == 100 {
		t.Error("boots durability should have decreased")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_StateTransition(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	equipComp := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	// Start at 51% durability (just above Worn -> Damaged threshold)
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{Durability: 51, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	attacker := world.CreateEntity()

	// Initialize cache with first damage
	sys.OnDamageTaken(attacker, entity, 10)

	// Now the item should have transitioned toward Damaged state
	newDur := equipComp.Slots[SlotChest].Stats.Durability
	if newDur >= 51 {
		t.Errorf("durability should have decreased from 51, got %d", newDur)
	}

	// Check cache was updated
	if _, ok := sys.lastStateCache[entity.ID]; !ok {
		t.Error("entity state cache not created")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_OnDamageTaken_DurabilityClamped(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	equipComp := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{Durability: 1, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	attacker := world.CreateEntity()

	// Apply massive damage
	sys.OnDamageTaken(attacker, entity, 10000)

	// Durability should be clamped to 0, not negative
	if equipComp.Slots[SlotChest].Stats.Durability < 0 {
		t.Errorf("durability = %d, should be >= 0", equipComp.Slots[SlotChest].Stats.Durability)
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetImpactColors(t *testing.T) {
	tests := []struct {
		name          string
		genre         string
		wantPrimary   string
		wantSecondary string
	}{
		{"fantasy", "fantasy", "silver", "white"},
		{"sci-fi", "sci-fi", "cyan", "blue"},
		{"horror", "horror", "gray", "darkgray"},
		{"cyberpunk", "cyberpunk", "orange", "yellow"},
		{"postapoc", "post-apocalyptic", "rust", "brown"},
		{"default", "unknown", "silver", "white"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)
			sys.SetGenre(tt.genre)

			primary, secondary := sys.getImpactColors()
			if primary != tt.wantPrimary {
				t.Errorf("primary = %s, want %s", primary, tt.wantPrimary)
			}
			if secondary != tt.wantSecondary {
				t.Errorf("secondary = %s, want %s", secondary, tt.wantSecondary)
			}
		})
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetDegradationColors(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	tests := []struct {
		state         sprites.DamageState
		wantPrimary   string
		wantSecondary string
	}{
		{sprites.DamageStateWorn, "yellow", "orange"},
		{sprites.DamageStateDamaged, "orange", "red"},
		{sprites.DamageStateBroken, "darkred", "brown"},
		{sprites.DamageStatePristine, "gray", "white"},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			primary, secondary := sys.getDegradationColors(tt.state)
			if primary != tt.wantPrimary {
				t.Errorf("primary = %s, want %s", primary, tt.wantPrimary)
			}
			if secondary != tt.wantSecondary {
				t.Errorf("secondary = %s, want %s", secondary, tt.wantSecondary)
			}
		})
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetParticleType(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	tests := []struct {
		state sprites.DamageState
	}{
		{sprites.DamageStateWorn},
		{sprites.DamageStateDamaged},
		{sprites.DamageStateBroken},
		{sprites.DamageStatePristine},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			pt := sys.getParticleType(tt.state)
			// Just verify it returns a valid type
			if pt < 0 {
				t.Errorf("invalid particle type for state %v", tt.state)
			}
		})
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetParticleCount(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	tests := []struct {
		state   sprites.DamageState
		wantMin int
		wantMax int
	}{
		{sprites.DamageStateWorn, 8, 15},
		{sprites.DamageStateDamaged, 15, 25},
		{sprites.DamageStateBroken, 25, 35},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			count := sys.getParticleCount(tt.state)
			if count < tt.wantMin || count > tt.wantMax {
				t.Errorf("count = %d, want in range [%d, %d]", count, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetDuration(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	// Test duration increases with damage severity
	wornDur := sys.getDuration(sprites.DamageStateWorn)
	damagedDur := sys.getDuration(sprites.DamageStateDamaged)
	brokenDur := sys.getDuration(sprites.DamageStateBroken)

	if wornDur >= damagedDur {
		t.Error("worn duration should be less than damaged")
	}
	if damagedDur >= brokenDur {
		t.Error("damaged duration should be less than broken")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetGravity(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	// Worn should float upward (negative gravity)
	wornGrav := sys.getGravity(sprites.DamageStateWorn)
	if wornGrav >= 0 {
		t.Error("worn gravity should be negative (upward float)")
	}

	// Damaged and Broken should fall (positive gravity)
	damagedGrav := sys.getGravity(sprites.DamageStateDamaged)
	if damagedGrav <= 0 {
		t.Error("damaged gravity should be positive (falling)")
	}

	brokenGrav := sys.getGravity(sprites.DamageStateBroken)
	if brokenGrav <= damagedGrav {
		t.Error("broken gravity should be greater than damaged")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetSizes(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	states := []sprites.DamageState{
		sprites.DamageStateWorn,
		sprites.DamageStateDamaged,
		sprites.DamageStateBroken,
	}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			minSize := sys.getMinSize(state)
			maxSize := sys.getMaxSize(state)

			if minSize <= 0 {
				t.Error("min size should be positive")
			}
			if maxSize <= minSize {
				t.Error("max size should be greater than min size")
			}
		})
	}
}

func TestCombatEquipmentDurabilityParticleSystem_GetSpread(t *testing.T) {
	sys := NewCombatEquipmentDurabilityParticleSystem(nil, 12345)

	// Spread should increase with damage severity
	wornSpread := sys.getSpread(sprites.DamageStateWorn)
	damagedSpread := sys.getSpread(sprites.DamageStateDamaged)
	brokenSpread := sys.getSpread(sprites.DamageStateBroken)

	if wornSpread >= damagedSpread {
		t.Error("worn spread should be less than damaged")
	}
	if damagedSpread >= brokenSpread {
		t.Error("damaged spread should be less than broken")
	}
}

func TestCombatEquipmentDurabilityParticleSystem_Update(t *testing.T) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)

	// Update should not panic (callback-driven system)
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.016)
}

func BenchmarkCombatEquipmentDurabilityParticleSystem_OnDamageTaken(b *testing.B) {
	world := NewWorld(nil)
	sys := NewCombatEquipmentDurabilityParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	equipComp := &EquipmentComponent{Slots: make(map[EquipmentSlot]*item.Item)}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{Durability: 10000, DurabilityMax: 10000},
	}
	entity.AddComponent(equipComp)

	attacker := world.CreateEntity()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnDamageTaken(attacker, entity, 50.0)
	}
}
