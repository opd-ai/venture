package skills

import "testing"

// TestIsSkillUnlocked tests the IsSkillUnlocked helper function.
func TestIsSkillUnlocked(t *testing.T) {
	tests := []struct {
		name          string
		skill         *Skill
		playerLevel   int
		skillPoints   int
		learnedSkills map[string]bool
		attributes    map[string]int
		expected      bool
	}{
		{
			name:          "nil skill returns false",
			skill:         nil,
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{},
			expected:      false,
		},
		{
			name: "all requirements met",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel:       5,
					SkillPoints:       1,
					PrerequisiteIDs:   []string{},
					AttributeMinimums: map[string]int{},
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{},
			expected:      true,
		},
		{
			name: "player level too low",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel: 15,
					SkillPoints: 1,
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{},
			expected:      false,
		},
		{
			name: "not enough skill points",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel: 5,
					SkillPoints: 10,
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{},
			expected:      false,
		},
		{
			name: "missing prerequisite",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel:     5,
					SkillPoints:     1,
					PrerequisiteIDs: []string{"prereq1"},
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{},
			expected:      false,
		},
		{
			name: "prerequisite met",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel:     5,
					SkillPoints:     1,
					PrerequisiteIDs: []string{"prereq1"},
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{"prereq1": true},
			attributes:    map[string]int{},
			expected:      true,
		},
		{
			name: "attribute too low",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel:       5,
					SkillPoints:       1,
					AttributeMinimums: map[string]int{"strength": 15},
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{"strength": 10},
			expected:      false,
		},
		{
			name: "attribute met",
			skill: &Skill{
				Requirements: Requirements{
					PlayerLevel:       5,
					SkillPoints:       1,
					AttributeMinimums: map[string]int{"strength": 15},
				},
			},
			playerLevel:   10,
			skillPoints:   5,
			learnedSkills: map[string]bool{},
			attributes:    map[string]int{"strength": 20},
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSkillUnlocked(tt.skill, tt.playerLevel, tt.skillPoints, tt.learnedSkills, tt.attributes)
			if got != tt.expected {
				t.Errorf("IsSkillUnlocked() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCanSkillLevelUp tests the CanSkillLevelUp helper function.
func TestCanSkillLevelUp(t *testing.T) {
	tests := []struct {
		name     string
		skill    *Skill
		expected bool
	}{
		{
			name:     "nil skill returns false",
			skill:    nil,
			expected: false,
		},
		{
			name:     "not learned (level 0) returns false",
			skill:    &Skill{Level: 0, MaxLevel: 5},
			expected: false,
		},
		{
			name:     "at max level returns false",
			skill:    &Skill{Level: 5, MaxLevel: 5},
			expected: false,
		},
		{
			name:     "can level up returns true",
			skill:    &Skill{Level: 3, MaxLevel: 5},
			expected: true,
		},
		{
			name:     "level 1 of 5 returns true",
			skill:    &Skill{Level: 1, MaxLevel: 5},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanSkillLevelUp(tt.skill)
			if got != tt.expected {
				t.Errorf("CanSkillLevelUp() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCalculateTreeTotalPoints tests the CalculateTreeTotalPoints helper function.
func TestCalculateTreeTotalPoints(t *testing.T) {
	tests := []struct {
		name     string
		tree     *SkillTree
		expected int
	}{
		{
			name:     "nil tree returns 0",
			tree:     nil,
			expected: 0,
		},
		{
			name:     "empty tree returns 0",
			tree:     &SkillTree{Nodes: []*SkillNode{}},
			expected: 0,
		},
		{
			name: "single skill with level 1",
			tree: &SkillTree{
				Nodes: []*SkillNode{
					{Skill: &Skill{Level: 1, Requirements: Requirements{SkillPoints: 2}}},
				},
			},
			expected: 2,
		},
		{
			name: "single skill with level 3",
			tree: &SkillTree{
				Nodes: []*SkillNode{
					{Skill: &Skill{Level: 3, Requirements: Requirements{SkillPoints: 2}}},
				},
			},
			expected: 6,
		},
		{
			name: "multiple skills",
			tree: &SkillTree{
				Nodes: []*SkillNode{
					{Skill: &Skill{Level: 2, Requirements: Requirements{SkillPoints: 1}}},
					{Skill: &Skill{Level: 0, Requirements: Requirements{SkillPoints: 3}}},
					{Skill: &Skill{Level: 3, Requirements: Requirements{SkillPoints: 2}}},
				},
			},
			expected: 8, // 2*1 + 0*3 + 3*2 = 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTreeTotalPoints(tt.tree)
			if got != tt.expected {
				t.Errorf("CalculateTreeTotalPoints() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFindSkillByID tests the FindSkillByID helper function.
func TestFindSkillByID(t *testing.T) {
	skill1 := &Skill{ID: "skill1", Name: "Skill One"}
	skill2 := &Skill{ID: "skill2", Name: "Skill Two"}
	tree := &SkillTree{
		Nodes: []*SkillNode{
			{Skill: skill1},
			{Skill: skill2},
		},
	}

	tests := []struct {
		name     string
		tree     *SkillTree
		id       string
		expected *Skill
	}{
		{
			name:     "nil tree returns nil",
			tree:     nil,
			id:       "skill1",
			expected: nil,
		},
		{
			name:     "empty tree returns nil",
			tree:     &SkillTree{Nodes: []*SkillNode{}},
			id:       "skill1",
			expected: nil,
		},
		{
			name:     "find skill1",
			tree:     tree,
			id:       "skill1",
			expected: skill1,
		},
		{
			name:     "find skill2",
			tree:     tree,
			id:       "skill2",
			expected: skill2,
		},
		{
			name:     "nonexistent skill returns nil",
			tree:     tree,
			id:       "nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSkillByID(tt.tree, tt.id)
			if got != tt.expected {
				t.Errorf("FindSkillByID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGetSkillsByTier tests the GetSkillsByTier helper function.
func TestGetSkillsByTier(t *testing.T) {
	basicSkill1 := &Skill{ID: "basic1", Tier: TierBasic}
	basicSkill2 := &Skill{ID: "basic2", Tier: TierBasic}
	advancedSkill := &Skill{ID: "advanced1", Tier: TierAdvanced}
	tree := &SkillTree{
		Nodes: []*SkillNode{
			{Skill: basicSkill1},
			{Skill: basicSkill2},
			{Skill: advancedSkill},
		},
	}

	tests := []struct {
		name          string
		tree          *SkillTree
		tier          Tier
		expectedCount int
	}{
		{
			name:          "nil tree returns nil",
			tree:          nil,
			tier:          TierBasic,
			expectedCount: -1, // special case for nil
		},
		{
			name:          "empty tree returns empty slice",
			tree:          &SkillTree{Nodes: []*SkillNode{}},
			tier:          TierBasic,
			expectedCount: 0,
		},
		{
			name:          "find basic tier skills",
			tree:          tree,
			tier:          TierBasic,
			expectedCount: 2,
		},
		{
			name:          "find advanced tier skills",
			tree:          tree,
			tier:          TierAdvanced,
			expectedCount: 1,
		},
		{
			name:          "find master tier skills (none)",
			tree:          tree,
			tier:          TierMaster,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSkillsByTier(tt.tree, tt.tier)
			if tt.expectedCount == -1 {
				if got != nil {
					t.Errorf("GetSkillsByTier() = %v, want nil", got)
				}
				return
			}
			if len(got) != tt.expectedCount {
				t.Errorf("GetSkillsByTier() returned %d skills, want %d", len(got), tt.expectedCount)
			}
		})
	}
}

// TestHelperFunctionsParity ensures helper functions match method behavior.
func TestHelperFunctionsParity(t *testing.T) {
	skill := &Skill{
		ID:       "test_skill",
		Level:    2,
		MaxLevel: 5,
		Requirements: Requirements{
			PlayerLevel:       5,
			SkillPoints:       2,
			PrerequisiteIDs:   []string{"prereq1"},
			AttributeMinimums: map[string]int{"strength": 10},
		},
	}

	tree := &SkillTree{
		Nodes: []*SkillNode{
			{Skill: skill},
			{Skill: &Skill{ID: "prereq1", Level: 1, Tier: TierBasic, Requirements: Requirements{SkillPoints: 1}}},
		},
	}

	playerLevel := 10
	skillPoints := 5
	learnedSkills := map[string]bool{"prereq1": true}
	attributes := map[string]int{"strength": 15}

	// Test IsUnlocked parity
	methodResult := skill.IsUnlocked(playerLevel, skillPoints, learnedSkills, attributes)
	helperResult := IsSkillUnlocked(skill, playerLevel, skillPoints, learnedSkills, attributes)
	if methodResult != helperResult {
		t.Errorf("IsUnlocked parity failed: method=%v, helper=%v", methodResult, helperResult)
	}

	// Test CanLevelUp parity
	methodLevelUp := skill.CanLevelUp()
	helperLevelUp := CanSkillLevelUp(skill)
	if methodLevelUp != helperLevelUp {
		t.Errorf("CanLevelUp parity failed: method=%v, helper=%v", methodLevelUp, helperLevelUp)
	}

	// Test TotalPoints parity
	methodPoints := tree.TotalPoints()
	helperPoints := CalculateTreeTotalPoints(tree)
	if methodPoints != helperPoints {
		t.Errorf("TotalPoints parity failed: method=%v, helper=%v", methodPoints, helperPoints)
	}

	// Test GetSkillByID parity
	methodSkill := tree.GetSkillByID("test_skill")
	helperSkill := FindSkillByID(tree, "test_skill")
	if methodSkill != helperSkill {
		t.Errorf("GetSkillByID parity failed: method=%v, helper=%v", methodSkill, helperSkill)
	}

	// Test GetTierSkills parity
	methodTier := tree.GetTierSkills(TierBasic)
	helperTier := GetSkillsByTier(tree, TierBasic)
	if len(methodTier) != len(helperTier) {
		t.Errorf("GetTierSkills parity failed: method len=%v, helper len=%v", len(methodTier), len(helperTier))
	}
}
