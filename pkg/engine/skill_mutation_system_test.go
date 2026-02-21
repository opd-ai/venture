// Package engine provides tests for the skill mutation system.
package engine

import (
	"testing"
)

// TestMutationType_String tests mutation type string conversion.
func TestMutationType_String(t *testing.T) {
	tests := []struct {
		name     string
		mutType  MutationType
		expected string
	}{
		{"Damage", MutationDamage, "Damage"},
		{"Cooldown", MutationCooldown, "Cooldown"},
		{"ManaCost", MutationManaCost, "Mana Cost"},
		{"Range", MutationRange, "Range"},
		{"Area", MutationArea, "Area"},
		{"Duration", MutationDuration, "Duration"},
		{"Elemental", MutationElemental, "Elemental"},
		{"Chain", MutationChain, "Chain"},
		{"Lifesteal", MutationLifesteal, "Lifesteal"},
		{"Pierce", MutationPierce, "Pierce"},
		{"Split", MutationSplit, "Split"},
		{"Echo", MutationEcho, "Echo"},
		{"Unknown", MutationType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mutType.String(); got != tt.expected {
				t.Errorf("MutationType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestMutationRarity_String tests mutation rarity string conversion.
func TestMutationRarity_String(t *testing.T) {
	tests := []struct {
		name     string
		rarity   MutationRarity
		expected string
	}{
		{"Common", MutationRarityCommon, "Common"},
		{"Uncommon", MutationRarityUncommon, "Uncommon"},
		{"Rare", MutationRarityRare, "Rare"},
		{"Epic", MutationRarityEpic, "Epic"},
		{"Legendary", MutationRarityLegendary, "Legendary"},
		{"Unknown", MutationRarity(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rarity.String(); got != tt.expected {
				t.Errorf("MutationRarity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSkillMutation_Copy tests mutation deep copy.
func TestSkillMutation_Copy(t *testing.T) {
	original := &SkillMutation{
		ID:             "test_123",
		Name:           "Test Mutation",
		Description:    "A test mutation",
		Type:           MutationDamage,
		Rarity:         MutationRarityRare,
		PrimaryValue:   25.0,
		SecondaryValue: 10.0,
		TradeoffType:   MutationManaCost,
		TradeoffValue:  15.0,
		Seed:           12345,
		RequiredLevel:  10,
		Incompatible:   []string{"other_mutation"},
	}

	copied := original.Copy()

	// Verify values match
	if copied.ID != original.ID {
		t.Errorf("Copy ID mismatch: got %v, want %v", copied.ID, original.ID)
	}
	if copied.PrimaryValue != original.PrimaryValue {
		t.Errorf("Copy PrimaryValue mismatch: got %v, want %v", copied.PrimaryValue, original.PrimaryValue)
	}

	// Verify deep copy of slice
	if len(copied.Incompatible) != len(original.Incompatible) {
		t.Errorf("Copy Incompatible length mismatch")
	}

	// Modify copy shouldn't affect original
	copied.Incompatible = append(copied.Incompatible, "new_item")
	if len(original.Incompatible) != 1 {
		t.Error("Copy did not create independent slice")
	}
}

// TestSkillMutation_GetPowerScore tests power score calculation.
func TestSkillMutation_GetPowerScore(t *testing.T) {
	tests := []struct {
		name     string
		mutation *SkillMutation
		minPower int
		maxPower int
	}{
		{
			name: "Basic damage mutation",
			mutation: &SkillMutation{
				Type:          MutationDamage,
				Rarity:        MutationRarityCommon,
				PrimaryValue:  20,
				TradeoffValue: 10,
			},
			minPower: 10,
			maxPower: 25,
		},
		{
			name: "Cooldown reduction (negative primary)",
			mutation: &SkillMutation{
				Type:          MutationCooldown,
				Rarity:        MutationRarityRare,
				PrimaryValue:  -25, // Negative for cooldown = good
				TradeoffValue: 10,
			},
			minPower: 20,
			maxPower: 50,
		},
		{
			name: "Legendary mutation",
			mutation: &SkillMutation{
				Type:          MutationDamage,
				Rarity:        MutationRarityLegendary,
				PrimaryValue:  50,
				TradeoffValue: 10,
			},
			minPower: 100,
			maxPower: 100, // Capped at 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			power := tt.mutation.GetPowerScore()
			if power < tt.minPower || power > tt.maxPower {
				t.Errorf("GetPowerScore() = %v, want between %v and %v", power, tt.minPower, tt.maxPower)
			}
		})
	}
}

// TestMutatedSkill_Operations tests mutated skill operations.
func TestMutatedSkill_Operations(t *testing.T) {
	ms := &MutatedSkill{
		SkillID:      "fireball",
		SpellSlot:    0,
		Mutations:    make([]*SkillMutation, 0),
		MaxMutations: 3,
		Locked:       false,
	}

	// Test CanAddMutation
	if !ms.CanAddMutation() {
		t.Error("CanAddMutation should return true for empty skill")
	}

	// Test GetMutationCount
	if ms.GetMutationCount() != 0 {
		t.Error("GetMutationCount should return 0 for empty skill")
	}

	// Add a mutation
	mutation := &SkillMutation{
		ID:           "damage_boost",
		Type:         MutationDamage,
		PrimaryValue: 20,
	}
	ms.Mutations = append(ms.Mutations, mutation)

	// Test HasMutation
	if !ms.HasMutation("damage_boost") {
		t.Error("HasMutation should return true for added mutation")
	}
	if ms.HasMutation("nonexistent") {
		t.Error("HasMutation should return false for missing mutation")
	}

	// Test GetMutationCount after add
	if ms.GetMutationCount() != 1 {
		t.Errorf("GetMutationCount = %d, want 1", ms.GetMutationCount())
	}

	// Test locked state
	ms.Locked = true
	if ms.CanAddMutation() {
		t.Error("CanAddMutation should return false when locked")
	}
}

// TestMutatedSkill_Compatibility tests mutation compatibility checking.
func TestMutatedSkill_Compatibility(t *testing.T) {
	ms := &MutatedSkill{
		SkillID:      "fireball",
		Mutations:    make([]*SkillMutation, 0),
		MaxMutations: 3,
	}

	chainMutation := &SkillMutation{
		ID:           "chain_lightning",
		Type:         MutationChain,
		Incompatible: []string{"split_bolt"},
	}
	ms.Mutations = append(ms.Mutations, chainMutation)

	// Compatible mutation
	compatible := &SkillMutation{
		ID:   "damage_boost",
		Type: MutationDamage,
	}
	if !ms.IsMutationCompatible(compatible) {
		t.Error("Damage mutation should be compatible with chain")
	}

	// Incompatible mutation (explicitly listed)
	incompatible := &SkillMutation{
		ID:   "split_bolt",
		Type: MutationSplit,
	}
	if ms.IsMutationCompatible(incompatible) {
		t.Error("Split mutation should be incompatible with chain (listed in chain's incompatible)")
	}
}

// TestMutatedSkill_Modifiers tests modifier calculations.
func TestMutatedSkill_Modifiers(t *testing.T) {
	ms := &MutatedSkill{
		SkillID:      "fireball",
		Mutations:    make([]*SkillMutation, 0),
		MaxMutations: 3,
	}

	// Add damage mutation +20%
	ms.Mutations = append(ms.Mutations, &SkillMutation{
		Type:         MutationDamage,
		PrimaryValue: 20,
	})

	// Add another damage mutation +15% with mana cost tradeoff
	ms.Mutations = append(ms.Mutations, &SkillMutation{
		Type:          MutationDamage,
		PrimaryValue:  15,
		TradeoffType:  MutationManaCost,
		TradeoffValue: 10,
	})

	// Test GetTotalModifier for damage
	damageMod := ms.GetTotalModifier(MutationDamage)
	if damageMod != 35 {
		t.Errorf("GetTotalModifier(Damage) = %v, want 35", damageMod)
	}

	// Test GetTotalModifier for mana cost (should have tradeoff penalty)
	manaMod := ms.GetTotalModifier(MutationManaCost)
	if manaMod != -10 { // Only tradeoff penalty
		t.Errorf("GetTotalModifier(ManaCost) = %v, want -10", manaMod)
	}

	// Test effective multipliers
	damageMultiplier := ms.GetEffectiveDamageMultiplier()
	if damageMultiplier != 1.35 {
		t.Errorf("GetEffectiveDamageMultiplier = %v, want 1.35", damageMultiplier)
	}
}

// TestMutatedSkill_SpecialEffects tests special effect getters.
func TestMutatedSkill_SpecialEffects(t *testing.T) {
	ms := &MutatedSkill{
		SkillID:   "fireball",
		Mutations: make([]*SkillMutation, 0),
	}

	// Add chain mutation
	ms.Mutations = append(ms.Mutations, &SkillMutation{
		Type:         MutationChain,
		PrimaryValue: 2,
	})

	// Add lifesteal mutation
	ms.Mutations = append(ms.Mutations, &SkillMutation{
		Type:         MutationLifesteal,
		PrimaryValue: 15,
	})

	// Add echo mutation
	ms.Mutations = append(ms.Mutations, &SkillMutation{
		Type:           MutationEcho,
		SecondaryValue: 25,
	})

	// Test GetChainTargets
	if chains := ms.GetChainTargets(); chains != 2 {
		t.Errorf("GetChainTargets = %d, want 2", chains)
	}

	// Test GetLifestealPercent
	if lifesteal := ms.GetLifestealPercent(); lifesteal != 15 {
		t.Errorf("GetLifestealPercent = %v, want 15", lifesteal)
	}

	// Test GetEchoChance
	if echo := ms.GetEchoChance(); echo != 25 {
		t.Errorf("GetEchoChance = %v, want 25", echo)
	}
}

// TestSkillMutationComponent_Creation tests component creation.
func TestSkillMutationComponent_Creation(t *testing.T) {
	comp := NewSkillMutationComponent()

	if comp == nil {
		t.Fatal("NewSkillMutationComponent returned nil")
	}
	if comp.Type() != "skill_mutation" {
		t.Errorf("Type() = %v, want skill_mutation", comp.Type())
	}
	if comp.MutatedSkills == nil {
		t.Error("MutatedSkills map should be initialized")
	}
	if comp.AvailableMutations == nil {
		t.Error("AvailableMutations slice should be initialized")
	}
	if comp.MaxMutationsPerSkill != 3 {
		t.Errorf("MaxMutationsPerSkill = %d, want 3", comp.MaxMutationsPerSkill)
	}
	if comp.MutationSlots != 20 {
		t.Errorf("MutationSlots = %d, want 20", comp.MutationSlots)
	}
}

// TestSkillMutationComponent_GetMutatedSkill tests mutated skill retrieval.
func TestSkillMutationComponent_GetMutatedSkill(t *testing.T) {
	comp := NewSkillMutationComponent()

	// Get creates if doesn't exist
	ms := comp.GetMutatedSkill("fireball")
	if ms == nil {
		t.Fatal("GetMutatedSkill should create skill if not exists")
	}
	if ms.SkillID != "fireball" {
		t.Errorf("SkillID = %v, want fireball", ms.SkillID)
	}
	if ms.MaxMutations != comp.MaxMutationsPerSkill {
		t.Error("MaxMutations should match component default")
	}

	// Get again returns same instance
	ms2 := comp.GetMutatedSkill("fireball")
	if ms != ms2 {
		t.Error("GetMutatedSkill should return same instance")
	}
}

// TestSkillMutationComponent_AddMutation tests mutation application.
func TestSkillMutationComponent_AddMutation(t *testing.T) {
	comp := NewSkillMutationComponent()

	mutation := &SkillMutation{
		ID:           "test_mutation",
		Type:         MutationDamage,
		PrimaryValue: 20,
	}

	// Add mutation
	if !comp.AddMutation("fireball", mutation) {
		t.Error("AddMutation should succeed")
	}
	if comp.TotalMutationsApplied != 1 {
		t.Errorf("TotalMutationsApplied = %d, want 1", comp.TotalMutationsApplied)
	}
	if !comp.Dirty {
		t.Error("Dirty flag should be set after AddMutation")
	}

	// Adding same mutation should fail
	if comp.AddMutation("fireball", mutation) {
		t.Error("AddMutation should fail for duplicate mutation")
	}
}

// TestSkillMutationComponent_RemoveMutation tests mutation removal.
func TestSkillMutationComponent_RemoveMutation(t *testing.T) {
	comp := NewSkillMutationComponent()

	mutation := &SkillMutation{
		ID:           "test_mutation",
		Type:         MutationDamage,
		PrimaryValue: 20,
	}
	comp.AddMutation("fireball", mutation)
	comp.Dirty = false

	// Remove mutation
	if !comp.RemoveMutation("fireball", "test_mutation") {
		t.Error("RemoveMutation should succeed")
	}
	if !comp.Dirty {
		t.Error("Dirty flag should be set after RemoveMutation")
	}

	// Remove from nonexistent skill
	if comp.RemoveMutation("nonexistent", "test_mutation") {
		t.Error("RemoveMutation should fail for nonexistent skill")
	}
}

// TestSkillMutationComponent_Inventory tests inventory operations.
func TestSkillMutationComponent_Inventory(t *testing.T) {
	comp := NewSkillMutationComponent()
	comp.MutationSlots = 3

	mutation1 := &SkillMutation{ID: "mut1", Name: "Mutation 1"}
	mutation2 := &SkillMutation{ID: "mut2", Name: "Mutation 2"}
	mutation3 := &SkillMutation{ID: "mut3", Name: "Mutation 3"}
	mutation4 := &SkillMutation{ID: "mut4", Name: "Mutation 4"}

	// Add to inventory
	if !comp.AddToInventory(mutation1) {
		t.Error("AddToInventory should succeed")
	}
	if !comp.AddToInventory(mutation2) {
		t.Error("AddToInventory should succeed")
	}
	if !comp.AddToInventory(mutation3) {
		t.Error("AddToInventory should succeed")
	}
	if comp.AddToInventory(mutation4) {
		t.Error("AddToInventory should fail when full")
	}

	// Test GetInventoryCount
	if comp.GetInventoryCount() != 3 {
		t.Errorf("GetInventoryCount = %d, want 3", comp.GetInventoryCount())
	}

	// Test GetMutationByID
	idx, found := comp.GetMutationByID("mut2")
	if found == nil || idx != 1 {
		t.Error("GetMutationByID should find mutation")
	}

	// Test RemoveFromInventory
	removed := comp.RemoveFromInventory(1)
	if removed == nil || removed.ID != "mut2" {
		t.Error("RemoveFromInventory should return removed mutation")
	}
	if comp.GetInventoryCount() != 2 {
		t.Errorf("GetInventoryCount after remove = %d, want 2", comp.GetInventoryCount())
	}
}

// TestSkillMutationComponent_Serialization tests persistence.
func TestSkillMutationComponent_Serialization(t *testing.T) {
	comp := NewSkillMutationComponent()
	comp.AddMutation("fireball", &SkillMutation{
		ID:           "test_mut",
		Name:         "Test Mutation",
		Type:         MutationDamage,
		PrimaryValue: 25,
	})
	comp.AddToInventory(&SkillMutation{
		ID:           "inv_mut",
		Name:         "Inventory Mutation",
		Type:         MutationCooldown,
		PrimaryValue: -15,
	})

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize into new component
	newComp := NewSkillMutationComponent()
	if err := newComp.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify data
	if len(newComp.MutatedSkills) != 1 {
		t.Error("MutatedSkills not restored")
	}
	if len(newComp.AvailableMutations) != 1 {
		t.Error("AvailableMutations not restored")
	}
	if !newComp.Dirty {
		t.Error("Dirty flag should be set after Deserialize")
	}
}

// TestSkillMutationSystem_Creation tests system creation.
func TestSkillMutationSystem_Creation(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	if sys == nil {
		t.Fatal("NewSkillMutationSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set")
	}
}

// TestSkillMutationSystem_GenerateMutation tests mutation generation.
func TestSkillMutationSystem_GenerateMutation(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	tests := []struct {
		name        string
		seed        int64
		rarity      MutationRarity
		playerLevel int
	}{
		{"Common mutation", 12345, MutationRarityCommon, 1},
		{"Uncommon mutation", 23456, MutationRarityUncommon, 10},
		{"Rare mutation", 34567, MutationRarityRare, 20},
		{"Epic mutation", 45678, MutationRarityEpic, 35},
		{"Legendary mutation", 56789, MutationRarityLegendary, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutation := sys.GenerateMutation(tt.seed, tt.rarity, tt.playerLevel)

			if mutation == nil {
				t.Fatal("GenerateMutation returned nil")
			}
			if mutation.ID == "" {
				t.Error("Mutation ID should not be empty")
			}
			if mutation.Name == "" {
				t.Error("Mutation Name should not be empty")
			}
			if mutation.Rarity != tt.rarity {
				t.Errorf("Rarity = %v, want %v", mutation.Rarity, tt.rarity)
			}
			if mutation.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", mutation.Seed, tt.seed)
			}

			// Verify deterministic generation
			mutation2 := sys.GenerateMutation(tt.seed, tt.rarity, tt.playerLevel)
			if mutation.ID != mutation2.ID {
				t.Error("Same seed should produce same mutation ID")
			}
			if mutation.Name != mutation2.Name {
				t.Error("Same seed should produce same mutation name")
			}
		})
	}
}

// TestSkillMutationSystem_Update tests system update processing.
func TestSkillMutationSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	entity := NewEntity(1)
	mutComp := NewSkillMutationComponent()
	mutComp.Dirty = true
	entity.AddComponent(mutComp)

	entities := []*Entity{entity}

	// Update should process dirty component
	sys.Update(entities, 0.016)

	// Dirty flag should be cleared
	if mutComp.Dirty {
		t.Error("Update should clear dirty flag")
	}
}

// TestSkillMutationSystem_ApplyMutationFromInventory tests mutation application.
func TestSkillMutationSystem_ApplyMutationFromInventory(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	entity := NewEntity(1)
	mutComp := NewSkillMutationComponent()
	mutComp.AddToInventory(&SkillMutation{
		ID:            "test_mut",
		Name:          "Test Mutation",
		Type:          MutationDamage,
		RequiredLevel: 1,
	})
	entity.AddComponent(mutComp)

	// Apply mutation
	err := sys.ApplyMutationFromInventory(entity, "fireball", 0)
	if err != nil {
		t.Errorf("ApplyMutationFromInventory failed: %v", err)
	}

	// Verify mutation applied
	ms := mutComp.GetMutatedSkill("fireball")
	if len(ms.Mutations) != 1 {
		t.Errorf("Expected 1 mutation on skill, got %d", len(ms.Mutations))
	}

	// Verify removed from inventory
	if len(mutComp.AvailableMutations) != 0 {
		t.Error("Mutation should be removed from inventory")
	}
}

// TestSkillMutationSystem_RemoveMutationFromSkill tests mutation removal.
func TestSkillMutationSystem_RemoveMutationFromSkill(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	entity := NewEntity(1)
	mutComp := NewSkillMutationComponent()
	mutComp.AddMutation("fireball", &SkillMutation{
		ID:   "test_mut",
		Name: "Test Mutation",
	})
	entity.AddComponent(mutComp)

	// Remove mutation
	err := sys.RemoveMutationFromSkill(entity, "fireball", "test_mut")
	if err != nil {
		t.Errorf("RemoveMutationFromSkill failed: %v", err)
	}

	// Verify mutation removed from skill
	ms := mutComp.GetMutatedSkill("fireball")
	if len(ms.Mutations) != 0 {
		t.Error("Mutation should be removed from skill")
	}

	// Verify returned to inventory
	if len(mutComp.AvailableMutations) != 1 {
		t.Error("Mutation should be returned to inventory")
	}
}

// TestSkillMutationSystem_GrantRandomMutation tests random mutation granting.
func TestSkillMutationSystem_GrantRandomMutation(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	entity := NewEntity(1)
	mutComp := NewSkillMutationComponent()
	entity.AddComponent(mutComp)

	// Grant mutation
	mutation, err := sys.GrantRandomMutation(entity, 12345, MutationRarityRare)
	if err != nil {
		t.Fatalf("GrantRandomMutation failed: %v", err)
	}

	if mutation == nil {
		t.Fatal("Granted mutation should not be nil")
	}
	if mutation.Rarity != MutationRarityRare {
		t.Errorf("Rarity = %v, want Rare", mutation.Rarity)
	}

	// Verify in inventory
	if len(mutComp.AvailableMutations) != 1 {
		t.Error("Mutation should be in inventory")
	}
}

// TestSkillMutationSystem_ModifiedSpellStats tests spell stat modification.
func TestSkillMutationSystem_ModifiedSpellStats(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	entity := NewEntity(1)
	mutComp := NewSkillMutationComponent()
	mutComp.AddMutation("spell_slot_0", &SkillMutation{
		Type:         MutationDamage,
		PrimaryValue: 50, // +50% damage
	})
	mutComp.AddMutation("spell_slot_0", &SkillMutation{
		Type:         MutationCooldown,
		PrimaryValue: -20, // -20% cooldown
	})
	entity.AddComponent(mutComp)

	// Test damage modification
	modifiedDamage := sys.GetModifiedSpellDamage(entity, 0, 100)
	if modifiedDamage != 150 {
		t.Errorf("GetModifiedSpellDamage = %d, want 150", modifiedDamage)
	}

	// Test cooldown modification
	modifiedCooldown := sys.GetModifiedSpellCooldown(entity, 0, 10.0)
	if modifiedCooldown != 8.0 {
		t.Errorf("GetModifiedSpellCooldown = %v, want 8.0", modifiedCooldown)
	}

	// Test entity without component
	entityNoComp := NewEntity(1)
	damage := sys.GetModifiedSpellDamage(entityNoComp, 0, 100)
	if damage != 100 {
		t.Errorf("Entity without component should return base damage, got %d", damage)
	}
}

// TestSkillMutationSystem_ErrorCases tests error handling.
func TestSkillMutationSystem_ErrorCases(t *testing.T) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	// Entity without component
	entityNoComp := NewEntity(1)
	err := sys.ApplyMutationFromInventory(entityNoComp, "fireball", 0)
	if err == nil {
		t.Error("Should error for entity without component")
	}

	// Invalid mutation index
	entity := NewEntity(1)
	mutComp := NewSkillMutationComponent()
	entity.AddComponent(mutComp)

	err = sys.ApplyMutationFromInventory(entity, "fireball", 99)
	if err == nil {
		t.Error("Should error for invalid mutation index")
	}

	// Remove from nonexistent skill
	err = sys.RemoveMutationFromSkill(entity, "nonexistent", "mut_id")
	if err == nil {
		t.Error("Should error for nonexistent skill")
	}
}

// BenchmarkSkillMutationSystem_Update benchmarks system update.
func BenchmarkSkillMutationSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(1)
		mutComp := NewSkillMutationComponent()
		mutComp.AddMutation("skill1", &SkillMutation{
			Type:         MutationDamage,
			PrimaryValue: 20,
		})
		mutComp.AddMutation("skill1", &SkillMutation{
			Type:         MutationCooldown,
			PrimaryValue: -10,
		})
		mutComp.Dirty = true
		e.AddComponent(mutComp)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset dirty flags
		for _, e := range entities {
			if comp, ok := e.GetComponent("skill_mutation"); ok && comp != nil {
				comp.(*SkillMutationComponent).Dirty = true
			}
		}
		sys.Update(entities, 0.016)
	}
}

// BenchmarkSkillMutationSystem_GenerateMutation benchmarks mutation generation.
func BenchmarkSkillMutationSystem_GenerateMutation(b *testing.B) {
	world := NewWorld()
	sys := NewSkillMutationSystem(world)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GenerateMutation(int64(i), MutationRarityRare, 20)
	}
}
