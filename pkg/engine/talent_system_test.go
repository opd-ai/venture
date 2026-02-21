// Package engine provides tests for the talent system.
package engine

import (
	"testing"
)

// TestTalentComponentType tests the component type identifier.
func TestTalentComponentType(t *testing.T) {
	comp := NewTalentComponent()
	if comp.Type() != "talent" {
		t.Errorf("TalentComponent.Type() = %q, want %q", comp.Type(), "talent")
	}
}

// TestNewTalentComponent tests talent component creation.
func TestNewTalentComponent(t *testing.T) {
	comp := NewTalentComponent()
	if comp.UnspentPoints != 0 {
		t.Errorf("NewTalentComponent().UnspentPoints = %d, want 0", comp.UnspentPoints)
	}
	if comp.TotalPointsEarned != 0 {
		t.Errorf("NewTalentComponent().TotalPointsEarned = %d, want 0", comp.TotalPointsEarned)
	}
	if len(comp.Allocations) != 0 {
		t.Errorf("NewTalentComponent().Allocations should be empty")
	}
	if !comp.Dirty {
		t.Error("NewTalentComponent().Dirty should be true initially")
	}
}

// TestAddTalentPoints tests adding talent points.
func TestAddTalentPoints(t *testing.T) {
	tests := []struct {
		name        string
		pointsToAdd int
		wantUnspent int
		wantTotal   int
	}{
		{"add positive points", 5, 5, 5},
		{"add zero points", 0, 0, 0},
		{"add negative points", -5, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewTalentComponent()
			comp.AddTalentPoints(tt.pointsToAdd)
			if comp.UnspentPoints != tt.wantUnspent {
				t.Errorf("UnspentPoints = %d, want %d", comp.UnspentPoints, tt.wantUnspent)
			}
			if comp.TotalPointsEarned != tt.wantTotal {
				t.Errorf("TotalPointsEarned = %d, want %d", comp.TotalPointsEarned, tt.wantTotal)
			}
		})
	}
}

// TestCanAllocate tests talent allocation validation.
func TestCanAllocate(t *testing.T) {
	tier1Talent := &TalentDefinition{
		ID:            "test_tier1",
		Name:          "Test Tier 1",
		Category:      TalentCategoryOffense,
		MaxRanks:      5,
		RequiredLevel: 1,
	}

	tier2Talent := &TalentDefinition{
		ID:                           "test_tier2",
		Name:                         "Test Tier 2",
		Category:                     TalentCategoryOffense,
		MaxRanks:                     3,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
	}

	tests := []struct {
		name           string
		setup          func(*TalentComponent)
		talent         *TalentDefinition
		characterLevel int
		want           bool
	}{
		{
			name:           "no points available",
			setup:          func(c *TalentComponent) {},
			talent:         tier1Talent,
			characterLevel: 1,
			want:           false,
		},
		{
			name: "level too low",
			setup: func(c *TalentComponent) {
				c.UnspentPoints = 5
			},
			talent:         tier2Talent,
			characterLevel: 1,
			want:           false,
		},
		{
			name: "can allocate basic talent",
			setup: func(c *TalentComponent) {
				c.UnspentPoints = 5
			},
			talent:         tier1Talent,
			characterLevel: 1,
			want:           true,
		},
		{
			name: "max ranks reached",
			setup: func(c *TalentComponent) {
				c.UnspentPoints = 5
				c.Allocations[tier1Talent.ID] = tier1Talent.MaxRanks
			},
			talent:         tier1Talent,
			characterLevel: 10,
			want:           false,
		},
		{
			name: "category prereq not met",
			setup: func(c *TalentComponent) {
				c.UnspentPoints = 5
				c.PointsInCategory[TalentCategoryOffense] = 3
			},
			talent:         tier2Talent,
			characterLevel: 10,
			want:           false,
		},
		{
			name: "category prereq met",
			setup: func(c *TalentComponent) {
				c.UnspentPoints = 5
				c.PointsInCategory[TalentCategoryOffense] = 5
			},
			talent:         tier2Talent,
			characterLevel: 10,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewTalentComponent()
			tt.setup(comp)
			got := comp.CanAllocate(tt.talent, tt.characterLevel)
			if got != tt.want {
				t.Errorf("CanAllocate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAllocatePoint tests talent point allocation.
func TestAllocatePoint(t *testing.T) {
	talent := &TalentDefinition{
		ID:            "test_allocate",
		Name:          "Test Allocate",
		Category:      TalentCategoryDefense,
		MaxRanks:      3,
		RequiredLevel: 1,
	}

	t.Run("successful allocation", func(t *testing.T) {
		comp := NewTalentComponent()
		comp.AddTalentPoints(5)
		comp.Dirty = false

		success := comp.AllocatePoint(talent, 1)
		if !success {
			t.Error("AllocatePoint() should succeed")
		}
		if comp.UnspentPoints != 4 {
			t.Errorf("UnspentPoints = %d, want 4", comp.UnspentPoints)
		}
		if comp.Allocations[talent.ID] != 1 {
			t.Errorf("Allocations[%q] = %d, want 1", talent.ID, comp.Allocations[talent.ID])
		}
		if comp.PointsInCategory[talent.Category] != 1 {
			t.Errorf("PointsInCategory = %d, want 1", comp.PointsInCategory[talent.Category])
		}
		if !comp.Dirty {
			t.Error("Dirty should be true after allocation")
		}
	})

	t.Run("failed allocation - no points", func(t *testing.T) {
		comp := NewTalentComponent()
		success := comp.AllocatePoint(talent, 1)
		if success {
			t.Error("AllocatePoint() should fail with no points")
		}
	})
}

// TestDeallocatePoint tests talent point deallocation.
func TestDeallocatePoint(t *testing.T) {
	talent := &TalentDefinition{
		ID:            "test_deallocate",
		Name:          "Test Deallocate",
		Category:      TalentCategoryUtility,
		MaxRanks:      3,
		RequiredLevel: 1,
	}

	t.Run("successful deallocation", func(t *testing.T) {
		comp := NewTalentComponent()
		comp.Allocations[talent.ID] = 2
		comp.PointsInCategory[talent.Category] = 2
		comp.Dirty = false

		success := comp.DeallocatePoint(talent)
		if !success {
			t.Error("DeallocatePoint() should succeed")
		}
		if comp.Allocations[talent.ID] != 1 {
			t.Errorf("Allocations = %d, want 1", comp.Allocations[talent.ID])
		}
		if comp.UnspentPoints != 1 {
			t.Errorf("UnspentPoints = %d, want 1", comp.UnspentPoints)
		}
		if !comp.Dirty {
			t.Error("Dirty should be true after deallocation")
		}
	})

	t.Run("failed deallocation - no ranks", func(t *testing.T) {
		comp := NewTalentComponent()
		success := comp.DeallocatePoint(talent)
		if success {
			t.Error("DeallocatePoint() should fail with no ranks")
		}
	})
}

// TestResetAll tests resetting all talent allocations.
func TestResetAll(t *testing.T) {
	comp := NewTalentComponent()
	comp.Allocations["talent1"] = 3
	comp.Allocations["talent2"] = 2
	comp.PointsInCategory[TalentCategoryOffense] = 3
	comp.PointsInCategory[TalentCategoryDefense] = 2
	comp.Dirty = false

	comp.ResetAll()

	if comp.UnspentPoints != 5 {
		t.Errorf("UnspentPoints = %d, want 5", comp.UnspentPoints)
	}
	if len(comp.Allocations) != 0 {
		t.Error("Allocations should be empty after reset")
	}
	if len(comp.PointsInCategory) != 0 {
		t.Error("PointsInCategory should be empty after reset")
	}
	if !comp.Dirty {
		t.Error("Dirty should be true after reset")
	}
}

// TestTalentComponentSerialization tests serialization and deserialization.
func TestTalentComponentSerialization(t *testing.T) {
	original := NewTalentComponent()
	original.AddTalentPoints(10)
	original.Allocations["test_talent"] = 3
	original.PointsInCategory[TalentCategoryOffense] = 3
	original.UnspentPoints = 7

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	restored := &TalentComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if restored.UnspentPoints != original.UnspentPoints {
		t.Errorf("UnspentPoints = %d, want %d", restored.UnspentPoints, original.UnspentPoints)
	}
	if restored.TotalPointsEarned != original.TotalPointsEarned {
		t.Errorf("TotalPointsEarned = %d, want %d", restored.TotalPointsEarned, original.TotalPointsEarned)
	}
	if restored.Allocations["test_talent"] != original.Allocations["test_talent"] {
		t.Errorf("Allocations mismatch")
	}
}

// TestTalentCategoryString tests category string representation.
func TestTalentCategoryString(t *testing.T) {
	tests := []struct {
		category TalentCategory
		want     string
	}{
		{TalentCategoryOffense, "Offense"},
		{TalentCategoryDefense, "Defense"},
		{TalentCategoryUtility, "Utility"},
		{TalentCategoryMastery, "Mastery"},
		{TalentCategory(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("TalentCategory.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetTalentDefinition tests talent registry lookup.
func TestGetTalentDefinition(t *testing.T) {
	// Test a known talent from definitions
	def := GetTalentDefinition("offense_raw_power")
	if def == nil {
		t.Error("GetTalentDefinition() should return offense_raw_power")
	}
	if def != nil && def.Name != "Raw Power" {
		t.Errorf("Talent name = %q, want %q", def.Name, "Raw Power")
	}

	// Test unknown talent
	unknown := GetTalentDefinition("nonexistent_talent")
	if unknown != nil {
		t.Error("GetTalentDefinition() should return nil for unknown talent")
	}
}

// TestGetTalentsByCategory tests category grouping.
func TestGetTalentsByCategory(t *testing.T) {
	categories := []TalentCategory{
		TalentCategoryOffense,
		TalentCategoryDefense,
		TalentCategoryUtility,
		TalentCategoryMastery,
	}

	for _, cat := range categories {
		talents := GetTalentsByCategory(cat)
		if len(talents) == 0 {
			t.Errorf("GetTalentsByCategory(%v) returned empty, expected talents", cat)
		}
		for _, talent := range talents {
			if talent.Category != cat {
				t.Errorf("Talent %q has category %v, expected %v", talent.ID, talent.Category, cat)
			}
		}
	}
}

// TestGetAllTalentDefinitions tests getting all talents.
func TestGetAllTalentDefinitions(t *testing.T) {
	all := GetAllTalentDefinitions()
	if len(all) < 20 {
		t.Errorf("GetAllTalentDefinitions() returned %d talents, expected at least 20", len(all))
	}

	// Verify each talent has required fields
	for _, def := range all {
		if def.ID == "" {
			t.Error("Talent has empty ID")
		}
		if def.Name == "" {
			t.Errorf("Talent %q has empty Name", def.ID)
		}
		if def.MaxRanks < 1 {
			t.Errorf("Talent %q has invalid MaxRanks: %d", def.ID, def.MaxRanks)
		}
	}
}

// TestTalentCombatBonusComponentType tests the component type.
func TestTalentCombatBonusComponentType(t *testing.T) {
	comp := &TalentCombatBonusComponent{}
	if comp.Type() != "talent_combat_bonus" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "talent_combat_bonus")
	}
}

// TestNewTalentSystem tests system creation.
func TestNewTalentSystem(t *testing.T) {
	world := NewWorld()
	system := NewTalentSystem(world)

	if system == nil {
		t.Fatal("NewTalentSystem() returned nil")
	}
	if system.world != world {
		t.Error("System world reference mismatch")
	}
	if system.updateInterval != 30 {
		t.Errorf("updateInterval = %d, want 30", system.updateInterval)
	}
}

// TestTalentSystemUpdate tests the update cycle.
func TestTalentSystemUpdate(t *testing.T) {
	world := NewWorld()
	system := NewTalentSystem(world)

	// Create entity with talent component
	entity := world.CreateEntity()
	talentComp := NewTalentComponent()
	talentComp.AddTalentPoints(5)
	talentComp.Allocations["offense_raw_power"] = 3
	entity.AddComponent(talentComp)

	// Add health component to receive bonuses
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	// Add stats component
	stats := &StatsComponent{Attack: 10, Defense: 5}
	entity.AddComponent(stats)

	// Run update cycles until processing occurs
	entities := []*Entity{entity}
	for i := 0; i < 35; i++ {
		system.Update(entities, 0.016)
	}

	// Verify dirty flag cleared
	if talentComp.Dirty {
		t.Error("Talent component should not be dirty after processing")
	}

	// Verify bonuses were calculated
	if talentComp.CachedBonuses.DamagePercent <= 0 {
		t.Error("Cached damage bonus should be positive")
	}
}

// TestTalentSystemAllocateTalentPoint tests the system allocation method.
func TestTalentSystemAllocateTalentPoint(t *testing.T) {
	world := NewWorld()
	system := NewTalentSystem(world)

	entity := world.CreateEntity()
	talentComp := NewTalentComponent()
	talentComp.AddTalentPoints(5)
	entity.AddComponent(talentComp)

	// Add experience for level check
	expComp := NewExperienceComponent()
	expComp.Level = 10
	entity.AddComponent(expComp)

	success := system.AllocateTalentPoint(entity, "offense_raw_power")
	if !success {
		t.Error("AllocateTalentPoint() should succeed")
	}

	if talentComp.Allocations["offense_raw_power"] != 1 {
		t.Errorf("Talent allocation = %d, want 1", talentComp.Allocations["offense_raw_power"])
	}
}

// TestTalentSystemResetTalents tests the reset method.
func TestTalentSystemResetTalents(t *testing.T) {
	world := NewWorld()
	system := NewTalentSystem(world)

	entity := world.CreateEntity()
	talentComp := NewTalentComponent()
	talentComp.AddTalentPoints(10)
	talentComp.Allocations["offense_raw_power"] = 3
	talentComp.PointsInCategory[TalentCategoryOffense] = 3
	talentComp.UnspentPoints = 7
	entity.AddComponent(talentComp)

	system.ResetTalents(entity)

	if talentComp.UnspentPoints != 10 {
		t.Errorf("UnspentPoints = %d, want 10", talentComp.UnspentPoints)
	}
	if len(talentComp.Allocations) != 0 {
		t.Error("Allocations should be empty after reset")
	}
}

// TestTalentSystemCallbacks tests callback invocation.
func TestTalentSystemCallbacks(t *testing.T) {
	world := NewWorld()
	system := NewTalentSystem(world)

	allocatedCalled := false
	system.SetOnTalentAllocated(func(entity *Entity, talentID string, newRank int) {
		allocatedCalled = true
		if talentID != "offense_raw_power" {
			t.Errorf("Callback talentID = %q, want %q", talentID, "offense_raw_power")
		}
	})

	entity := world.CreateEntity()
	talentComp := NewTalentComponent()
	talentComp.AddTalentPoints(5)
	entity.AddComponent(talentComp)

	expComp := NewExperienceComponent()
	expComp.Level = 5
	entity.AddComponent(expComp)

	system.AllocateTalentPoint(entity, "offense_raw_power")

	if !allocatedCalled {
		t.Error("Allocation callback should have been called")
	}
}

// TestTalentBonusCalculation tests bonus calculation accuracy.
func TestTalentBonusCalculation(t *testing.T) {
	world := NewWorld()
	system := NewTalentSystem(world)

	entity := world.CreateEntity()
	talentComp := NewTalentComponent()
	// Allocate 5 ranks of offense_raw_power (+2% damage per rank)
	talentComp.Allocations["offense_raw_power"] = 5
	talentComp.Dirty = true
	entity.AddComponent(talentComp)

	bonuses := system.calculateTotalBonuses(talentComp)

	// 5 ranks * 0.02 = 0.10 (10% damage increase)
	expected := 0.10
	if bonuses.DamagePercent != expected {
		t.Errorf("DamagePercent = %f, want %f", bonuses.DamagePercent, expected)
	}
}

// BenchmarkTalentSystemUpdate benchmarks the system update.
func BenchmarkTalentSystemUpdate(b *testing.B) {
	world := NewWorld()
	system := NewTalentSystem(world)

	// Create 100 entities with talents
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		talentComp := NewTalentComponent()
		talentComp.Allocations["offense_raw_power"] = 3
		talentComp.Allocations["defense_fortitude"] = 2
		talentComp.Dirty = true
		entity.AddComponent(talentComp)
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatsComponent{Attack: 10, Defense: 5})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Force dirty flag for each benchmark iteration
		for _, e := range entities {
			comp, _ := e.GetComponent("talent")
			if talent, ok := comp.(*TalentComponent); ok {
				talent.Dirty = true
			}
		}
		system.frameCounter = 0 // Force processing
		system.Update(entities, 0.016)
	}
}
