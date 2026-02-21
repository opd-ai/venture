package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/skills"
)

// createTestSkillTree creates a skill tree for testing.
func createTestSkillTree() *skills.SkillTree {
	tree := &skills.SkillTree{
		ID:        "test_tree",
		Name:      "Test Tree",
		MaxPoints: 100,
		Nodes:     make([]*skills.SkillNode, 0),
	}

	// Create test skills with prerequisites
	skill1 := &skills.Skill{
		ID:       "skill_basic",
		Name:     "Basic Strike",
		Type:     skills.TypeActive,
		Category: skills.CategoryCombat,
		Tier:     skills.TierBasic,
		MaxLevel: 5,
		Requirements: skills.Requirements{
			PlayerLevel: 1,
			SkillPoints: 1,
		},
		Effects: []skills.Effect{
			{Type: "damage", Value: 5, IsPercent: false},
		},
	}

	skill2 := &skills.Skill{
		ID:       "skill_advanced",
		Name:     "Power Strike",
		Type:     skills.TypeActive,
		Category: skills.CategoryCombat,
		Tier:     skills.TierIntermediate,
		MaxLevel: 3,
		Requirements: skills.Requirements{
			PlayerLevel:     5,
			SkillPoints:     2,
			PrerequisiteIDs: []string{"skill_basic"},
		},
		Effects: []skills.Effect{
			{Type: "damage", Value: 15, IsPercent: false},
		},
	}

	skill3 := &skills.Skill{
		ID:       "skill_passive",
		Name:     "Toughness",
		Type:     skills.TypePassive,
		Category: skills.CategoryDefense,
		Tier:     skills.TierBasic,
		MaxLevel: 5,
		Requirements: skills.Requirements{
			PlayerLevel: 1,
			SkillPoints: 1,
		},
		Effects: []skills.Effect{
			{Type: "defense", Value: 3, IsPercent: true},
		},
	}

	tree.Nodes = []*skills.SkillNode{
		{Skill: skill1, Position: skills.Position{X: 0, Y: 0}},
		{Skill: skill2, Position: skills.Position{X: 0, Y: 1}},
		{Skill: skill3, Position: skills.Position{X: 1, Y: 0}},
	}
	tree.RootNodes = tree.Nodes[:1]

	return tree
}

// TestSkillLoadoutComponent_Type tests component type identifier.
func TestSkillLoadoutComponent_Type(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	if comp.Type() != "skill_loadout" {
		t.Errorf("Expected type 'skill_loadout', got '%s'", comp.Type())
	}
}

// TestNewSkillLoadoutComponent tests component initialization.
func TestNewSkillLoadoutComponent(t *testing.T) {
	comp := NewSkillLoadoutComponent()

	if comp == nil {
		t.Fatal("NewSkillLoadoutComponent returned nil")
	}
	if comp.Loadouts == nil {
		t.Error("Loadouts slice should not be nil")
	}
	if comp.ActiveIndex != -1 {
		t.Errorf("ActiveIndex should be -1, got %d", comp.ActiveIndex)
	}
	if comp.MaxLoadouts != 10 {
		t.Errorf("MaxLoadouts should be 10, got %d", comp.MaxLoadouts)
	}
	if comp.UnlockedSlots != 2 {
		t.Errorf("UnlockedSlots should be 2, got %d", comp.UnlockedSlots)
	}
}

// TestSkillLoadoutComponent_SaveLoadout tests saving loadouts.
func TestSkillLoadoutComponent_SaveLoadout(t *testing.T) {
	tests := []struct {
		name        string
		loadoutName string
		description string
		skillLevels map[string]int
		treeID      string
		time        float64
		wantIndex   int
		wantSuccess bool
	}{
		{
			name:        "save first loadout",
			loadoutName: "Combat Build",
			description: "Pure damage focus",
			skillLevels: map[string]int{"skill_basic": 3, "skill_passive": 2},
			treeID:      "test_tree",
			time:        100.0,
			wantIndex:   0,
			wantSuccess: true,
		},
		{
			name:        "save second loadout",
			loadoutName: "Tank Build",
			description: "Max defense",
			skillLevels: map[string]int{"skill_passive": 5},
			treeID:      "test_tree",
			time:        200.0,
			wantIndex:   1,
			wantSuccess: true,
		},
	}

	comp := NewSkillLoadoutComponent()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := comp.SaveLoadout(tt.loadoutName, tt.description, tt.skillLevels, tt.treeID, tt.time)

			if tt.wantSuccess && index < 0 {
				t.Error("Expected successful save, got failure")
				return
			}

			if !tt.wantSuccess && index >= 0 {
				t.Error("Expected failed save, got success")
				return
			}

			if index != tt.wantIndex {
				t.Errorf("Expected index %d, got %d", tt.wantIndex, index)
			}

			if tt.wantSuccess {
				loadout := comp.GetLoadout(index)
				if loadout == nil {
					t.Fatal("GetLoadout returned nil")
				}
				if loadout.Name != tt.loadoutName {
					t.Errorf("Name mismatch: got %s, want %s", loadout.Name, tt.loadoutName)
				}
				if loadout.Description != tt.description {
					t.Errorf("Description mismatch: got %s, want %s", loadout.Description, tt.description)
				}
			}
		})
	}
}

// TestSkillLoadoutComponent_SaveLoadout_SlotLimit tests slot limit enforcement.
func TestSkillLoadoutComponent_SaveLoadout_SlotLimit(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.UnlockedSlots = 2

	// Fill all slots
	comp.SaveLoadout("Build 1", "", map[string]int{"s1": 1}, "tree", 1.0)
	comp.SaveLoadout("Build 2", "", map[string]int{"s2": 1}, "tree", 2.0)

	// Third should fail
	index := comp.SaveLoadout("Build 3", "", map[string]int{"s3": 1}, "tree", 3.0)
	if index != -1 {
		t.Error("Expected save to fail when slots are full")
	}
}

// TestSkillLoadoutComponent_DeleteLoadout tests loadout deletion.
func TestSkillLoadoutComponent_DeleteLoadout(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Build 1", "", map[string]int{"s1": 1}, "tree", 1.0)
	comp.SaveLoadout("Build 2", "", map[string]int{"s2": 1}, "tree", 2.0)
	comp.ActiveIndex = 1
	comp.AssignQuickSlot(0, 0)
	comp.AssignQuickSlot(1, 1)

	// Delete first loadout
	if !comp.DeleteLoadout(0) {
		t.Error("DeleteLoadout should succeed")
	}

	if comp.GetLoadoutCount() != 1 {
		t.Errorf("Expected 1 loadout remaining, got %d", comp.GetLoadoutCount())
	}

	// Quick slot 0 should be cleared (was pointing to deleted loadout)
	if comp.QuickSlots[0] != -1 {
		t.Errorf("Quick slot 0 should be -1, got %d", comp.QuickSlots[0])
	}

	// Quick slot 1 should be updated (index shifted down)
	if comp.QuickSlots[1] != 0 {
		t.Errorf("Quick slot 1 should be 0, got %d", comp.QuickSlots[1])
	}

	// Active index should be updated
	if comp.ActiveIndex != 0 {
		t.Errorf("ActiveIndex should be 0, got %d", comp.ActiveIndex)
	}
}

// TestSkillLoadoutComponent_Cooldown tests swap cooldown mechanics.
func TestSkillLoadoutComponent_Cooldown(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SwapCooldown = 30.0
	comp.LastSwapTime = 100.0

	tests := []struct {
		time     float64
		canSwap  bool
		expected float64 // expected remaining cooldown
	}{
		{time: 100.0, canSwap: false, expected: 30.0},
		{time: 115.0, canSwap: false, expected: 15.0},
		{time: 129.9, canSwap: false, expected: 0.1},
		{time: 130.0, canSwap: true, expected: 0.0},
		{time: 200.0, canSwap: true, expected: 0.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			canSwap := comp.CanSwapLoadout(tt.time)
			if canSwap != tt.canSwap {
				t.Errorf("At time %.1f: CanSwapLoadout = %v, want %v", tt.time, canSwap, tt.canSwap)
			}

			remaining := comp.GetSwapCooldownRemaining(tt.time)
			// Allow small floating point tolerance
			if remaining < tt.expected-0.01 || remaining > tt.expected+0.01 {
				t.Errorf("At time %.1f: remaining = %.1f, want %.1f", tt.time, remaining, tt.expected)
			}
		})
	}
}

// TestSkillLoadoutComponent_QuickSlots tests quick slot assignment.
func TestSkillLoadoutComponent_QuickSlots(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Build 1", "", map[string]int{"s1": 1}, "tree", 1.0)
	comp.SaveLoadout("Build 2", "", map[string]int{"s2": 1}, "tree", 2.0)

	// Assign to quick slot
	if !comp.AssignQuickSlot(0, 0) {
		t.Error("AssignQuickSlot should succeed")
	}

	// Get loadout from quick slot
	loadout := comp.GetQuickSlotLoadout(0)
	if loadout == nil {
		t.Error("GetQuickSlotLoadout should return loadout")
	}
	if loadout.Name != "Build 1" {
		t.Errorf("Expected 'Build 1', got '%s'", loadout.Name)
	}

	// Clear quick slot
	if !comp.AssignQuickSlot(0, -1) {
		t.Error("Clearing quick slot should succeed")
	}
	if comp.QuickSlots[0] != -1 {
		t.Error("Quick slot should be cleared")
	}
}

// TestSkillLoadoutComponent_Serialization tests JSON serialization.
func TestSkillLoadoutComponent_Serialization(t *testing.T) {
	original := NewSkillLoadoutComponent()
	original.SaveLoadout("Combat", "Damage build", map[string]int{"skill1": 3, "skill2": 2}, "tree1", 100.0)
	original.SaveLoadout("Tank", "Defense build", map[string]int{"skill3": 5}, "tree1", 200.0)
	original.ActiveIndex = 0
	original.QuickSlots[0] = 0
	original.QuickSlots[1] = 1

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	// Deserialize
	restored := &SkillLoadoutComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	// Verify
	if restored.GetLoadoutCount() != 2 {
		t.Errorf("Expected 2 loadouts, got %d", restored.GetLoadoutCount())
	}
	if restored.ActiveIndex != 0 {
		t.Errorf("ActiveIndex mismatch: got %d, want 0", restored.ActiveIndex)
	}
	if restored.QuickSlots[0] != 0 || restored.QuickSlots[1] != 1 {
		t.Error("QuickSlots not restored correctly")
	}

	loadout := restored.GetLoadout(0)
	if loadout == nil || loadout.Name != "Combat" {
		t.Error("First loadout not restored correctly")
	}
}

// TestSkillLoadout_Copy tests deep copy of loadout.
func TestSkillLoadout_Copy(t *testing.T) {
	original := &SkillLoadout{
		Name:        "Original",
		Description: "Test",
		SkillLevels: map[string]int{"s1": 3, "s2": 2},
		TreeID:      "tree1",
		CreatedAt:   100.0,
		LastUsedAt:  200.0,
	}

	copy := original.Copy()

	// Verify independent
	copy.Name = "Copy"
	copy.SkillLevels["s1"] = 5
	copy.SkillLevels["s3"] = 1

	if original.Name != "Original" {
		t.Error("Original name was modified")
	}
	if original.SkillLevels["s1"] != 3 {
		t.Error("Original skill levels were modified")
	}
	if _, exists := original.SkillLevels["s3"]; exists {
		t.Error("Original gained new skill")
	}
}

// TestSkillLoadoutSystem_Creation tests system creation.
func TestSkillLoadoutSystem_Creation(t *testing.T) {
	world := NewWorld()
	system := NewSkillLoadoutSystem(world)

	if system == nil {
		t.Fatal("NewSkillLoadoutSystem returned nil")
	}
	if system.world != world {
		t.Error("World reference not set")
	}
}

// TestSkillLoadoutSystem_Update_NoEntities tests update with no entities.
func TestSkillLoadoutSystem_Update_NoEntities(t *testing.T) {
	world := NewWorld()
	system := NewSkillLoadoutSystem(world)

	// Should not panic
	system.Update([]*Entity{}, 0.016)
}

// TestSkillLoadoutSystem_SaveAndSwap tests saving and swapping loadouts.
func TestSkillLoadoutSystem_SaveAndSwap(t *testing.T) {
	world := NewWorld()
	system := NewSkillLoadoutSystem(world)

	// Create entity with skill tree
	entity := NewEntity(1001)
	tree := createTestSkillTree()
	skillTreeComp := NewSkillTreeComponent(tree)

	// Learn some skills
	skillTreeComp.LearnSkill("skill_basic", 10)
	skillTreeComp.LearnSkill("skill_basic", 9)
	skillTreeComp.LearnSkill("skill_passive", 8)

	// Add experience component for skill points
	expComp := &ExperienceComponent{Level: 10}
	entity.AddComponent(expComp)
	entity.AddComponent(skillTreeComp)

	// Save current as loadout
	index := system.SaveCurrentAsLoadout(entity, "Build A", "First build")
	if index != 0 {
		t.Errorf("Expected index 0, got %d", index)
	}

	// Verify loadout was saved
	loadoutComp, _ := entity.GetComponent("skill_loadout")
	loadout := loadoutComp.(*SkillLoadoutComponent)

	if loadout.GetLoadoutCount() != 1 {
		t.Errorf("Expected 1 loadout, got %d", loadout.GetLoadoutCount())
	}

	savedLoadout := loadout.GetLoadout(0)
	if savedLoadout.Name != "Build A" {
		t.Errorf("Expected name 'Build A', got '%s'", savedLoadout.Name)
	}
}

// TestSkillLoadoutComponent_CalculateSkillDifference tests difference calculation.
func TestSkillLoadoutComponent_CalculateSkillDifference(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Target", "", map[string]int{"s1": 3, "s2": 2, "s4": 1}, "tree", 1.0)

	current := map[string]int{"s1": 1, "s3": 2}

	toAdd, toRemove := comp.CalculateSkillDifference(0, current)

	// s1: need +2 levels
	if toAdd["s1"] != 2 {
		t.Errorf("s1 toAdd should be 2, got %d", toAdd["s1"])
	}
	// s2: need +2 levels (new skill)
	if toAdd["s2"] != 2 {
		t.Errorf("s2 toAdd should be 2, got %d", toAdd["s2"])
	}
	// s4: need +1 level (new skill)
	if toAdd["s4"] != 1 {
		t.Errorf("s4 toAdd should be 1, got %d", toAdd["s4"])
	}
	// s3: need to remove 2 levels
	if toRemove["s3"] != 2 {
		t.Errorf("s3 toRemove should be 2, got %d", toRemove["s3"])
	}
}

// TestSkillLoadoutComponent_ValidateCompatibility tests tree compatibility.
func TestSkillLoadoutComponent_ValidateCompatibility(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Build", "", map[string]int{"s1": 1}, "tree_a", 1.0)

	// Same tree - should pass
	if err := comp.ValidateLoadoutCompatibility(0, "tree_a"); err != nil {
		t.Errorf("Same tree should be compatible: %v", err)
	}

	// Different tree - should fail
	if err := comp.ValidateLoadoutCompatibility(0, "tree_b"); err == nil {
		t.Error("Different tree should be incompatible")
	}

	// Empty tree ID in loadout - should pass
	comp.Loadouts[0].TreeID = ""
	if err := comp.ValidateLoadoutCompatibility(0, "any_tree"); err != nil {
		t.Errorf("Empty tree ID should be compatible with any: %v", err)
	}
}

// TestSkillLoadoutComponent_UnlockSlot tests slot unlocking.
func TestSkillLoadoutComponent_UnlockSlot(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	initialSlots := comp.UnlockedSlots

	if !comp.UnlockSlot() {
		t.Error("UnlockSlot should succeed")
	}

	if comp.UnlockedSlots != initialSlots+1 {
		t.Errorf("Expected %d slots, got %d", initialSlots+1, comp.UnlockedSlots)
	}

	// Fill to max
	comp.UnlockedSlots = comp.MaxLoadouts
	if comp.UnlockSlot() {
		t.Error("UnlockSlot should fail at max capacity")
	}
}

// TestEnsureLoadoutComponent tests component creation helper.
func TestEnsureLoadoutComponent(t *testing.T) {
	entity := NewEntity(1002)

	// First call - creates component
	comp1 := EnsureLoadoutComponent(entity)
	if comp1 == nil {
		t.Fatal("EnsureLoadoutComponent returned nil")
	}

	// Second call - returns existing
	comp2 := EnsureLoadoutComponent(entity)
	if comp1 != comp2 {
		t.Error("EnsureLoadoutComponent should return same component")
	}

	// Nil entity
	if EnsureLoadoutComponent(nil) != nil {
		t.Error("Should return nil for nil entity")
	}
}

// TestSkillLoadoutComponent_UpdateLoadout tests updating existing loadouts.
func TestSkillLoadoutComponent_UpdateLoadout(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Build 1", "Desc", map[string]int{"s1": 1}, "tree", 1.0)

	// Update loadout
	newSkills := map[string]int{"s1": 3, "s2": 2}
	if !comp.UpdateLoadout(0, newSkills, 2.0) {
		t.Error("UpdateLoadout should succeed")
	}

	loadout := comp.GetLoadout(0)
	if loadout.SkillLevels["s1"] != 3 {
		t.Errorf("Expected s1=3, got %d", loadout.SkillLevels["s1"])
	}
	if loadout.SkillLevels["s2"] != 2 {
		t.Errorf("Expected s2=2, got %d", loadout.SkillLevels["s2"])
	}

	// Update invalid index
	if comp.UpdateLoadout(99, newSkills, 3.0) {
		t.Error("UpdateLoadout should fail for invalid index")
	}
}

// TestSkillLoadoutComponent_RenameLoadout tests renaming loadouts.
func TestSkillLoadoutComponent_RenameLoadout(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Old Name", "Old Desc", map[string]int{}, "tree", 1.0)

	if !comp.RenameLoadout(0, "New Name", "New Desc") {
		t.Error("RenameLoadout should succeed")
	}

	loadout := comp.GetLoadout(0)
	if loadout.Name != "New Name" {
		t.Errorf("Expected 'New Name', got '%s'", loadout.Name)
	}
	if loadout.Description != "New Desc" {
		t.Errorf("Expected 'New Desc', got '%s'", loadout.Description)
	}

	// Rename invalid index
	if comp.RenameLoadout(99, "X", "Y") {
		t.Error("RenameLoadout should fail for invalid index")
	}
}

// TestSkillLoadoutComponent_GetAvailableSlots tests available slot counting.
func TestSkillLoadoutComponent_GetAvailableSlots(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.UnlockedSlots = 5

	if comp.GetAvailableSlots() != 5 {
		t.Errorf("Expected 5 available, got %d", comp.GetAvailableSlots())
	}

	comp.SaveLoadout("B1", "", map[string]int{}, "tree", 1.0)
	comp.SaveLoadout("B2", "", map[string]int{}, "tree", 2.0)

	if comp.GetAvailableSlots() != 3 {
		t.Errorf("Expected 3 available, got %d", comp.GetAvailableSlots())
	}
}

// TestSkillLoadoutComponent_GetLoadoutNames tests name list retrieval.
func TestSkillLoadoutComponent_GetLoadoutNames(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Alpha", "", map[string]int{}, "tree", 1.0)
	comp.SaveLoadout("Beta", "", map[string]int{}, "tree", 2.0)

	names := comp.GetLoadoutNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
	if names[0] != "Alpha" || names[1] != "Beta" {
		t.Errorf("Names mismatch: got %v", names)
	}
}

// TestSkillLoadoutComponent_IsActiveLoadout tests active loadout checking.
func TestSkillLoadoutComponent_IsActiveLoadout(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.ActiveIndex = 1

	if comp.IsActiveLoadout(0) {
		t.Error("Index 0 should not be active")
	}
	if !comp.IsActiveLoadout(1) {
		t.Error("Index 1 should be active")
	}
}

// TestSkillLoadoutComponent_ClearDirtyFlag tests dirty flag clearing.
func TestSkillLoadoutComponent_ClearDirtyFlag(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.IsDirty = true

	comp.ClearDirtyFlag()

	if comp.IsDirty {
		t.Error("IsDirty should be false after clearing")
	}
}

// TestSkillLoadout_TotalPointsUsed tests point calculation.
func TestSkillLoadout_TotalPointsUsed(t *testing.T) {
	loadout := &SkillLoadout{
		SkillLevels: map[string]int{"s1": 3, "s2": 2, "s3": 5},
	}

	if loadout.TotalPointsUsed() != 10 {
		t.Errorf("Expected 10, got %d", loadout.TotalPointsUsed())
	}

	// Empty loadout
	emptyLoadout := &SkillLoadout{SkillLevels: map[string]int{}}
	if emptyLoadout.TotalPointsUsed() != 0 {
		t.Error("Empty loadout should have 0 points")
	}
}

// TestSkillLoadoutComponent_GetLoadoutByName tests name-based lookup.
func TestSkillLoadoutComponent_GetLoadoutByName(t *testing.T) {
	comp := NewSkillLoadoutComponent()
	comp.SaveLoadout("Alpha", "First", map[string]int{"s1": 1}, "tree", 1.0)
	comp.SaveLoadout("Beta", "Second", map[string]int{"s2": 2}, "tree", 2.0)

	idx, loadout := comp.GetLoadoutByName("Beta")
	if idx != 1 {
		t.Errorf("Expected index 1, got %d", idx)
	}
	if loadout == nil || loadout.Description != "Second" {
		t.Error("Wrong loadout returned")
	}

	// Not found
	idx, loadout = comp.GetLoadoutByName("Gamma")
	if idx != -1 || loadout != nil {
		t.Error("Should return -1, nil for not found")
	}
}
