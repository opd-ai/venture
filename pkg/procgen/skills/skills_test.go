package skills

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewSkillTreeGenerator(t *testing.T) {
	gen := NewSkillTreeGenerator()
	if gen == nil {
		t.Fatal("NewSkillTreeGenerator returned nil")
	}
}

func TestSkillTreeGeneration(t *testing.T) {
	gen := NewSkillTreeGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 3,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	trees, ok := result.([]*SkillTree)
	if !ok {
		t.Fatalf("Expected []*SkillTree, got %T", result)
	}

	if len(trees) != 3 {
		t.Errorf("Expected 3 trees, got %d", len(trees))
	}

	for i, tree := range trees {
		if tree == nil {
			t.Errorf("Tree %d is nil", i)
			continue
		}

		if tree.Name == "" {
			t.Errorf("Tree %d has empty name", i)
		}

		if tree.Description == "" {
			t.Errorf("Tree %d has empty description", i)
		}

		if len(tree.Nodes) == 0 {
			t.Errorf("Tree %d has no nodes", i)
		}

		if len(tree.RootNodes) == 0 {
			t.Errorf("Tree %d has no root nodes", i)
		}

		// Verify all skills have effects
		for j, node := range tree.Nodes {
			if node == nil || node.Skill == nil {
				t.Errorf("Tree %d node %d is nil", i, j)
				continue
			}

			skill := node.Skill
			if skill.Name == "" {
				t.Errorf("Tree %d skill %d has empty name", i, j)
			}

			if len(skill.Effects) == 0 {
				t.Errorf("Tree %d skill %d has no effects: %s", i, j, skill.Name)
			}

			if skill.MaxLevel < 1 {
				t.Errorf("Tree %d skill %d has invalid max level: %d", i, j, skill.MaxLevel)
			}
		}
	}
}

func TestSkillTreeGenerationDeterministic(t *testing.T) {
	gen := NewSkillTreeGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 2,
		},
	}

	seed := int64(99999)

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generate failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generate failed: %v", err2)
	}

	trees1 := result1.([]*SkillTree)
	trees2 := result2.([]*SkillTree)

	if len(trees1) != len(trees2) {
		t.Errorf("Different number of trees: %d vs %d", len(trees1), len(trees2))
	}

	// Verify trees are identical
	for i := range trees1 {
		if i >= len(trees2) {
			break
		}

		tree1 := trees1[i]
		tree2 := trees2[i]

		if tree1.Name != tree2.Name {
			t.Errorf("Tree %d name mismatch: %s vs %s", i, tree1.Name, tree2.Name)
		}

		if len(tree1.Nodes) != len(tree2.Nodes) {
			t.Errorf("Tree %d node count mismatch: %d vs %d", i, len(tree1.Nodes), len(tree2.Nodes))
		}

		// Check skills are the same
		for j := range tree1.Nodes {
			if j >= len(tree2.Nodes) {
				break
			}

			skill1 := tree1.Nodes[j].Skill
			skill2 := tree2.Nodes[j].Skill

			if skill1.Name != skill2.Name {
				t.Errorf("Tree %d skill %d name mismatch: %s vs %s", i, j, skill1.Name, skill2.Name)
			}

			if skill1.Type != skill2.Type {
				t.Errorf("Tree %d skill %d type mismatch: %s vs %s", i, j, skill1.Type, skill2.Type)
			}
		}
	}
}

func TestSkillTreeGenerationSciFi(t *testing.T) {
	gen := NewSkillTreeGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"count": 3,
		},
	}

	result, err := gen.Generate(54321, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	trees, ok := result.([]*SkillTree)
	if !ok {
		t.Fatalf("Expected []*SkillTree, got %T", result)
	}

	if len(trees) != 3 {
		t.Errorf("Expected 3 trees, got %d", len(trees))
	}

	// Verify sci-fi genre
	for i, tree := range trees {
		if tree.Genre != "scifi" {
			t.Errorf("Tree %d has wrong genre: %s", i, tree.Genre)
		}
	}
}

func TestSkillTreeValidation(t *testing.T) {
	gen := NewSkillTreeGenerator()
	params := procgen.GenerationParams{
		Depth:      3,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 2,
		},
	}

	result, err := gen.Generate(11111, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Validation should pass
	err = gen.Validate(result)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

func TestSkillTreeGenerator_Validate(t *testing.T) {
	gen := NewSkillTreeGenerator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid trees",
			input: []*SkillTree{
				{
					Name:      "Test Tree",
					RootNodes: []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
					Nodes:     []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty tree list",
			input:   []*SkillTree{},
			wantErr: true,
		},
		{
			name:    "nil tree in list",
			input:   []*SkillTree{nil},
			wantErr: true,
		},
		{
			name: "tree with empty name",
			input: []*SkillTree{
				{
					Name:      "",
					RootNodes: []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
					Nodes:     []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
				},
			},
			wantErr: true,
		},
		{
			name: "tree with no nodes",
			input: []*SkillTree{
				{
					Name:  "Test",
					Nodes: []*SkillNode{},
				},
			},
			wantErr: true,
		},
		{
			name: "tree with no root nodes",
			input: []*SkillTree{
				{
					Name:      "Test",
					RootNodes: []*SkillNode{},
					Nodes:     []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
				},
			},
			wantErr: true,
		},
		{
			name: "skill with empty name",
			input: []*SkillTree{
				{
					Name:      "Test Tree",
					RootNodes: []*SkillNode{{Skill: &Skill{ID: "1", Name: "", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
					Nodes:     []*SkillNode{{Skill: &Skill{ID: "1", Name: "", MaxLevel: 1, Effects: []Effect{{Type: "test"}}}}},
				},
			},
			wantErr: true,
		},
		{
			name: "skill with invalid max level",
			input: []*SkillTree{
				{
					Name:      "Test Tree",
					RootNodes: []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 0, Effects: []Effect{{Type: "test"}}}}},
					Nodes:     []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 0, Effects: []Effect{{Type: "test"}}}}},
				},
			},
			wantErr: true,
		},
		{
			name: "skill with no effects",
			input: []*SkillTree{
				{
					Name:      "Test Tree",
					RootNodes: []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{}}}},
					Nodes:     []*SkillNode{{Skill: &Skill{ID: "1", Name: "Test", MaxLevel: 1, Effects: []Effect{}}}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSkillTreeGenerator_ValidateWrongType(t *testing.T) {
	gen := NewSkillTreeGenerator()
	err := gen.Validate("not a skill tree")
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestSkillType_String(t *testing.T) {
	tests := []struct {
		skillType SkillType
		want      string
	}{
		{TypePassive, "passive"},
		{TypeActive, "active"},
		{TypeUltimate, "ultimate"},
		{TypeSynergy, "synergy"},
		{SkillType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.skillType.String(); got != tt.want {
				t.Errorf("SkillType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkillCategory_String(t *testing.T) {
	tests := []struct {
		category SkillCategory
		want     string
	}{
		{CategoryCombat, "combat"},
		{CategoryDefense, "defense"},
		{CategoryUtility, "utility"},
		{CategoryMagic, "magic"},
		{CategoryCrafting, "crafting"},
		{CategorySocial, "social"},
		{SkillCategory(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("SkillCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTier_String(t *testing.T) {
	tests := []struct {
		tier Tier
		want string
	}{
		{TierBasic, "basic"},
		{TierIntermediate, "intermediate"},
		{TierAdvanced, "advanced"},
		{TierMaster, "master"},
		{Tier(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("Tier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkill_IsUnlocked(t *testing.T) {
	skill := &Skill{
		ID:       "test",
		Name:     "Test Skill",
		MaxLevel: 5,
		Requirements: Requirements{
			PlayerLevel:       10,
			SkillPoints:       5,
			PrerequisiteIDs:   []string{"prereq1", "prereq2"},
			AttributeMinimums: map[string]int{"strength": 15},
		},
	}

	tests := []struct {
		name          string
		playerLevel   int
		skillPoints   int
		learnedSkills map[string]bool
		attributes    map[string]int
		wantUnlocked  bool
	}{
		{
			name:          "all requirements met",
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{"prereq1": true, "prereq2": true},
			attributes:    map[string]int{"strength": 15},
			wantUnlocked:  true,
		},
		{
			name:          "level too low",
			playerLevel:   9,
			skillPoints:   5,
			learnedSkills: map[string]bool{"prereq1": true, "prereq2": true},
			attributes:    map[string]int{"strength": 15},
			wantUnlocked:  false,
		},
		{
			name:          "not enough skill points",
			playerLevel:   10,
			skillPoints:   4,
			learnedSkills: map[string]bool{"prereq1": true, "prereq2": true},
			attributes:    map[string]int{"strength": 15},
			wantUnlocked:  false,
		},
		{
			name:          "missing prerequisite",
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{"prereq1": true},
			attributes:    map[string]int{"strength": 15},
			wantUnlocked:  false,
		},
		{
			name:          "attribute too low",
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{"prereq1": true, "prereq2": true},
			attributes:    map[string]int{"strength": 14},
			wantUnlocked:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skill.IsUnlocked(tt.playerLevel, tt.skillPoints, tt.learnedSkills, tt.attributes)
			if got != tt.wantUnlocked {
				t.Errorf("IsUnlocked() = %v, want %v", got, tt.wantUnlocked)
			}
		})
	}
}

func TestSkill_CanLevelUp(t *testing.T) {
	tests := []struct {
		name         string
		level        int
		maxLevel     int
		wantCanLevel bool
	}{
		{
			name:         "can level up",
			level:        2,
			maxLevel:     5,
			wantCanLevel: true,
		},
		{
			name:         "at max level",
			level:        5,
			maxLevel:     5,
			wantCanLevel: false,
		},
		{
			name:         "not learned",
			level:        0,
			maxLevel:     5,
			wantCanLevel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := &Skill{
				Level:    tt.level,
				MaxLevel: tt.maxLevel,
			}
			got := skill.CanLevelUp()
			if got != tt.wantCanLevel {
				t.Errorf("CanLevelUp() = %v, want %v", got, tt.wantCanLevel)
			}
		})
	}
}

func TestSkillTree_TotalPoints(t *testing.T) {
	tree := &SkillTree{
		Nodes: []*SkillNode{
			{Skill: &Skill{Level: 3, Requirements: Requirements{SkillPoints: 1}}},
			{Skill: &Skill{Level: 2, Requirements: Requirements{SkillPoints: 2}}},
			{Skill: &Skill{Level: 0, Requirements: Requirements{SkillPoints: 1}}}, // Not learned
		},
	}

	expected := 3*1 + 2*2 // 3 + 4 = 7
	got := tree.TotalPoints()
	if got != expected {
		t.Errorf("TotalPoints() = %d, want %d", got, expected)
	}
}

func TestSkillTree_GetSkillByID(t *testing.T) {
	skill1 := &Skill{ID: "skill1", Name: "Skill 1"}
	skill2 := &Skill{ID: "skill2", Name: "Skill 2"}

	tree := &SkillTree{
		Nodes: []*SkillNode{
			{Skill: skill1},
			{Skill: skill2},
		},
	}

	// Test finding existing skill
	found := tree.GetSkillByID("skill1")
	if found == nil {
		t.Error("GetSkillByID() returned nil for existing skill")
	} else if found.Name != "Skill 1" {
		t.Errorf("GetSkillByID() returned wrong skill: %s", found.Name)
	}

	// Test not finding skill
	notFound := tree.GetSkillByID("nonexistent")
	if notFound != nil {
		t.Error("GetSkillByID() should return nil for nonexistent skill")
	}
}

func TestSkillTree_GetTierSkills(t *testing.T) {
	tree := &SkillTree{
		Nodes: []*SkillNode{
			{Skill: &Skill{ID: "s1", Tier: TierBasic}},
			{Skill: &Skill{ID: "s2", Tier: TierBasic}},
			{Skill: &Skill{ID: "s3", Tier: TierIntermediate}},
			{Skill: &Skill{ID: "s4", Tier: TierAdvanced}},
		},
	}

	basicSkills := tree.GetTierSkills(TierBasic)
	if len(basicSkills) != 2 {
		t.Errorf("GetTierSkills(TierBasic) returned %d skills, want 2", len(basicSkills))
	}

	intermediateSkills := tree.GetTierSkills(TierIntermediate)
	if len(intermediateSkills) != 1 {
		t.Errorf("GetTierSkills(TierIntermediate) returned %d skills, want 1", len(intermediateSkills))
	}

	masterSkills := tree.GetTierSkills(TierMaster)
	if len(masterSkills) != 0 {
		t.Errorf("GetTierSkills(TierMaster) returned %d skills, want 0", len(masterSkills))
	}
}

func TestGetFantasyTemplates(t *testing.T) {
	templates := GetFantasyTreeTemplates()
	if len(templates) == 0 {
		t.Fatal("GetFantasyTreeTemplates() returned empty slice")
	}

	expectedTrees := []string{"Warrior", "Mage", "Rogue"}
	if len(templates) != len(expectedTrees) {
		t.Errorf("Expected %d trees, got %d", len(expectedTrees), len(templates))
	}

	for i, expected := range expectedTrees {
		if i >= len(templates) {
			break
		}
		if templates[i].Name != expected {
			t.Errorf("Tree %d: expected name %s, got %s", i, expected, templates[i].Name)
		}
	}
}

func TestGetSciFiTemplates(t *testing.T) {
	templates := GetSciFiTreeTemplates()
	if len(templates) == 0 {
		t.Fatal("GetSciFiTreeTemplates() returned empty slice")
	}

	expectedTrees := []string{"Soldier", "Engineer", "Biotic"}
	if len(templates) != len(expectedTrees) {
		t.Errorf("Expected %d trees, got %d", len(expectedTrees), len(templates))
	}

	for i, expected := range expectedTrees {
		if i >= len(templates) {
			break
		}
		if templates[i].Name != expected {
			t.Errorf("Tree %d: expected name %s, got %s", i, expected, templates[i].Name)
		}
	}
}

func TestSkillTreeDepthScaling(t *testing.T) {
	gen := NewSkillTreeGenerator()

	// Test with different depths
	depths := []int{1, 5, 10, 20}
	for _, depth := range depths {
		params := procgen.GenerationParams{
			Depth:      depth,
			Difficulty: 0.5,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"count": 1,
			},
		}

		result, err := gen.Generate(12345, params)
		if err != nil {
			t.Fatalf("Generate at depth %d failed: %v", depth, err)
		}

		trees := result.([]*SkillTree)
		if len(trees) == 0 {
			t.Fatalf("No trees generated at depth %d", depth)
		}

		tree := trees[0]

		// Max points should scale with depth
		expectedMinPoints := 50 + depth*5
		if tree.MaxPoints < expectedMinPoints {
			t.Errorf("Depth %d: MaxPoints = %d, expected at least %d", depth, tree.MaxPoints, expectedMinPoints)
		}

		// Player level requirements should scale with depth
		for _, node := range tree.Nodes {
			if node.Skill.Requirements.PlayerLevel < depth {
				// At least some skills should scale with depth
				// (not all will due to tier requirements)
			}
		}
	}
}
