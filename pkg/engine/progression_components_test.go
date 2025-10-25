package engine

import (
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/skills"
)

// TestExperienceComponent_Type tests component type
func TestExperienceComponent_Type(t *testing.T) {
	exp := NewExperienceComponent()
	if exp.Type() != "experience" {
		t.Errorf("Type() = %v, want 'experience'", exp.Type())
	}
}

// TestNewExperienceComponent tests component initialization
func TestNewExperienceComponent(t *testing.T) {
	exp := NewExperienceComponent()

	if exp == nil {
		t.Fatal("NewExperienceComponent returned nil")
	}

	if exp.Level != 1 {
		t.Errorf("Level = %d, want 1", exp.Level)
	}
	if exp.CurrentXP != 0 {
		t.Errorf("CurrentXP = %d, want 0", exp.CurrentXP)
	}
	if exp.RequiredXP != 100 {
		t.Errorf("RequiredXP = %d, want 100", exp.RequiredXP)
	}
	if exp.TotalXP != 0 {
		t.Errorf("TotalXP = %d, want 0", exp.TotalXP)
	}
	if exp.SkillPoints != 0 {
		t.Errorf("SkillPoints = %d, want 0", exp.SkillPoints)
	}
}

// TestExperienceComponent_AddXP tests XP addition
func TestExperienceComponent_AddXP(t *testing.T) {
	tests := []struct {
		name        string
		initialXP   int
		requiredXP  int
		addXP       int
		wantLevelUp bool
		wantCurrent int
		wantTotal   int
	}{
		{"Add 50 XP", 0, 100, 50, false, 50, 50},
		{"Add 100 XP (level up)", 0, 100, 100, true, 100, 100},
		{"Add 150 XP (level up)", 0, 100, 150, true, 150, 150},
		{"Zero XP", 50, 100, 0, false, 50, 50},
		{"Negative XP", 50, 100, -10, false, 50, 50},
		{"Multiple additions", 75, 100, 30, true, 105, 105},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := NewExperienceComponent()
			exp.CurrentXP = tt.initialXP
			exp.TotalXP = tt.initialXP
			exp.RequiredXP = tt.requiredXP

			leveledUp := exp.AddXP(tt.addXP)

			if leveledUp != tt.wantLevelUp {
				t.Errorf("AddXP() levelUp = %v, want %v", leveledUp, tt.wantLevelUp)
			}
			if exp.CurrentXP != tt.wantCurrent {
				t.Errorf("CurrentXP = %d, want %d", exp.CurrentXP, tt.wantCurrent)
			}
			if exp.TotalXP != tt.wantTotal {
				t.Errorf("TotalXP = %d, want %d", exp.TotalXP, tt.wantTotal)
			}
		})
	}
}

// TestExperienceComponent_CanLevelUp tests level up check
func TestExperienceComponent_CanLevelUp(t *testing.T) {
	tests := []struct {
		name       string
		currentXP  int
		requiredXP int
		want       bool
	}{
		{"Not enough XP", 50, 100, false},
		{"Exactly enough XP", 100, 100, true},
		{"More than enough XP", 150, 100, true},
		{"Zero required", 50, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := NewExperienceComponent()
			exp.CurrentXP = tt.currentXP
			exp.RequiredXP = tt.requiredXP

			got := exp.CanLevelUp()
			if got != tt.want {
				t.Errorf("CanLevelUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExperienceComponent_ProgressToNextLevel tests progress calculation
func TestExperienceComponent_ProgressToNextLevel(t *testing.T) {
	tests := []struct {
		name       string
		currentXP  int
		requiredXP int
		wantMin    float64
		wantMax    float64
	}{
		{"0% progress", 0, 100, 0.0, 0.01},
		{"50% progress", 50, 100, 0.49, 0.51},
		{"100% progress", 100, 100, 1.0, 1.0},
		{"Over 100%", 150, 100, 1.0, 1.0},
		{"Zero required", 50, 0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := NewExperienceComponent()
			exp.CurrentXP = tt.currentXP
			exp.RequiredXP = tt.requiredXP

			got := exp.ProgressToNextLevel()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("ProgressToNextLevel() = %v, want between %v and %v",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestExperienceComponent_String tests string representation
func TestExperienceComponent_String(t *testing.T) {
	exp := NewExperienceComponent()
	exp.Level = 5
	exp.CurrentXP = 75
	exp.RequiredXP = 150
	exp.SkillPoints = 3

	str := exp.String()

	if !strings.Contains(str, "Level 5") {
		t.Errorf("String should contain 'Level 5', got: %s", str)
	}
	if !strings.Contains(str, "75/150") {
		t.Errorf("String should contain '75/150', got: %s", str)
	}
	if !strings.Contains(str, "3 skill points") {
		t.Errorf("String should contain '3 skill points', got: %s", str)
	}
}

// TestLevelScalingComponent_Type tests component type
func TestLevelScalingComponent_Type(t *testing.T) {
	scaling := NewLevelScalingComponent()
	if scaling.Type() != "level_scaling" {
		t.Errorf("Type() = %v, want 'level_scaling'", scaling.Type())
	}
}

// TestNewLevelScalingComponent tests component initialization
func TestNewLevelScalingComponent(t *testing.T) {
	scaling := NewLevelScalingComponent()

	if scaling == nil {
		t.Fatal("NewLevelScalingComponent returned nil")
	}

	// Verify default values
	if scaling.HealthPerLevel != 10.0 {
		t.Errorf("HealthPerLevel = %v, want 10.0", scaling.HealthPerLevel)
	}
	if scaling.AttackPerLevel != 2.0 {
		t.Errorf("AttackPerLevel = %v, want 2.0", scaling.AttackPerLevel)
	}
	if scaling.DefensePerLevel != 1.5 {
		t.Errorf("DefensePerLevel = %v, want 1.5", scaling.DefensePerLevel)
	}
	if scaling.BaseHealth != 100.0 {
		t.Errorf("BaseHealth = %v, want 100.0", scaling.BaseHealth)
	}
}

// TestLevelScalingComponent_CalculateStatForLevel tests stat calculation
func TestLevelScalingComponent_CalculateStatForLevel(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		name     string
		base     float64
		perLevel float64
		level    int
		want     float64
	}{
		{"Level 1", 100.0, 10.0, 1, 100.0},
		{"Level 2", 100.0, 10.0, 2, 110.0},
		{"Level 5", 100.0, 10.0, 5, 140.0},
		{"Level 10", 100.0, 10.0, 10, 190.0},
		{"Level 0 (clamped)", 100.0, 10.0, 0, 100.0},
		{"Negative level", 100.0, 10.0, -5, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scaling.CalculateStatForLevel(tt.base, tt.perLevel, tt.level)
			if got != tt.want {
				t.Errorf("CalculateStatForLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLevelScalingComponent_CalculateHealthForLevel tests health calculation
func TestLevelScalingComponent_CalculateHealthForLevel(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		level int
		want  float64
	}{
		{1, 100.0},
		{2, 110.0},
		{5, 140.0},
		{10, 190.0},
	}

	for _, tt := range tests {
		got := scaling.CalculateHealthForLevel(tt.level)
		if got != tt.want {
			t.Errorf("CalculateHealthForLevel(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

// TestLevelScalingComponent_CalculateAttackForLevel tests attack calculation
func TestLevelScalingComponent_CalculateAttackForLevel(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		level int
		want  float64
	}{
		{1, 10.0},
		{2, 12.0},
		{5, 18.0},
		{10, 28.0},
	}

	for _, tt := range tests {
		got := scaling.CalculateAttackForLevel(tt.level)
		if got != tt.want {
			t.Errorf("CalculateAttackForLevel(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

// TestLevelScalingComponent_CalculateDefenseForLevel tests defense calculation
func TestLevelScalingComponent_CalculateDefenseForLevel(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		level int
		want  float64
	}{
		{1, 5.0},
		{2, 6.5},
		{5, 11.0},
		{10, 18.5},
	}

	for _, tt := range tests {
		got := scaling.CalculateDefenseForLevel(tt.level)
		if got != tt.want {
			t.Errorf("CalculateDefenseForLevel(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

// TestLevelScalingComponent_CalculateMagicPowerForLevel tests magic power calculation
func TestLevelScalingComponent_CalculateMagicPowerForLevel(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		level int
		want  float64
	}{
		{1, 10.0},
		{2, 12.0},
		{5, 18.0},
	}

	for _, tt := range tests {
		got := scaling.CalculateMagicPowerForLevel(tt.level)
		if got != tt.want {
			t.Errorf("CalculateMagicPowerForLevel(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

// TestLevelScalingComponent_CalculateMagicDefenseForLevel tests magic defense calculation
func TestLevelScalingComponent_CalculateMagicDefenseForLevel(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		level int
		want  float64
	}{
		{1, 5.0},
		{2, 6.5},
		{5, 11.0},
	}

	for _, tt := range tests {
		got := scaling.CalculateMagicDefenseForLevel(tt.level)
		if got != tt.want {
			t.Errorf("CalculateMagicDefenseForLevel(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

// TestLevelScalingComponent_String tests string representation
func TestLevelScalingComponent_String(t *testing.T) {
	scaling := NewLevelScalingComponent()
	str := scaling.String()

	if !strings.Contains(str, "HP+10") {
		t.Errorf("String should contain 'HP+10', got: %s", str)
	}
	if !strings.Contains(str, "ATK+2") {
		t.Errorf("String should contain 'ATK+2', got: %s", str)
	}
}

// TestSkillTreeComponent_Type tests component type
func TestSkillTreeComponent_Type(t *testing.T) {
	tree := &skills.SkillTree{
		ID:    "test-tree",
		Nodes: []*skills.SkillNode{},
	}
	comp := NewSkillTreeComponent(tree)

	if comp.Type() != "skill_tree" {
		t.Errorf("Type() = %v, want 'skill_tree'", comp.Type())
	}
}

// TestNewSkillTreeComponent tests component initialization
func TestNewSkillTreeComponent(t *testing.T) {
	tree := &skills.SkillTree{
		ID:    "warrior-tree",
		Nodes: []*skills.SkillNode{},
	}

	comp := NewSkillTreeComponent(tree)

	if comp == nil {
		t.Fatal("NewSkillTreeComponent returned nil")
	}
	if comp.Tree != tree {
		t.Error("Tree not set correctly")
	}
	if comp.LearnedSkills == nil {
		t.Error("LearnedSkills map not initialized")
	}
	if comp.SkillLevels == nil {
		t.Error("SkillLevels map not initialized")
	}
	if comp.Attributes == nil {
		t.Error("Attributes map not initialized")
	}
	if comp.TotalPointsUsed != 0 {
		t.Errorf("TotalPointsUsed = %d, want 0", comp.TotalPointsUsed)
	}
}

// TestSkillTreeComponent_GetSkillLevel tests skill level retrieval
func TestSkillTreeComponent_GetSkillLevel(t *testing.T) {
	tree := &skills.SkillTree{
		ID:    "test-tree",
		Nodes: []*skills.SkillNode{},
	}
	comp := NewSkillTreeComponent(tree)

	// Unlearned skill should return 0
	if level := comp.GetSkillLevel("unknown-skill"); level != 0 {
		t.Errorf("GetSkillLevel('unknown-skill') = %d, want 0", level)
	}

	// Set a skill level
	comp.SkillLevels["fireball"] = 3

	if level := comp.GetSkillLevel("fireball"); level != 3 {
		t.Errorf("GetSkillLevel('fireball') = %d, want 3", level)
	}
}

// TestSkillTreeComponent_IsSkillLearned tests skill learned check
func TestSkillTreeComponent_IsSkillLearned(t *testing.T) {
	tree := &skills.SkillTree{
		ID:    "test-tree",
		Nodes: []*skills.SkillNode{},
	}
	comp := NewSkillTreeComponent(tree)

	// Unlearned skill
	if comp.IsSkillLearned("unknown-skill") {
		t.Error("IsSkillLearned('unknown-skill') should be false")
	}

	// Mark skill as learned
	comp.LearnedSkills["fireball"] = true

	if !comp.IsSkillLearned("fireball") {
		t.Error("IsSkillLearned('fireball') should be true")
	}
}

// TestSkillTreeComponent_LearnSkill tests skill learning
func TestSkillTreeComponent_LearnSkill(t *testing.T) {
	// Create a skill tree with one skill
	skill := &skills.Skill{
		ID:       "basic-attack",
		Name:     "Basic Attack",
		MaxLevel: 3,
		Requirements: skills.Requirements{
			SkillPoints: 1,
		},
	}

	tree := &skills.SkillTree{
		ID: "warrior-tree",
		Nodes: []*skills.SkillNode{
			{Skill: skill},
		},
	}

	comp := NewSkillTreeComponent(tree)

	// Learn skill with enough points
	availablePoints := 5
	if !comp.LearnSkill("basic-attack", availablePoints) {
		t.Error("LearnSkill should succeed with enough points")
	}

	// Verify skill was learned
	if !comp.IsSkillLearned("basic-attack") {
		t.Error("Skill should be marked as learned")
	}
	if comp.GetSkillLevel("basic-attack") != 1 {
		t.Errorf("Skill level = %d, want 1", comp.GetSkillLevel("basic-attack"))
	}
	if comp.TotalPointsUsed != 1 {
		t.Errorf("TotalPointsUsed = %d, want 1", comp.TotalPointsUsed)
	}

	// Level up the skill
	if !comp.LearnSkill("basic-attack", availablePoints) {
		t.Error("LearnSkill should succeed for level 2")
	}
	if comp.GetSkillLevel("basic-attack") != 2 {
		t.Errorf("Skill level = %d, want 2", comp.GetSkillLevel("basic-attack"))
	}

	// Level up again
	if !comp.LearnSkill("basic-attack", availablePoints) {
		t.Error("LearnSkill should succeed for level 3")
	}
	if comp.GetSkillLevel("basic-attack") != 3 {
		t.Errorf("Skill level = %d, want 3", comp.GetSkillLevel("basic-attack"))
	}

	// Try to exceed max level
	if comp.LearnSkill("basic-attack", availablePoints) {
		t.Error("LearnSkill should fail at max level")
	}
	if comp.GetSkillLevel("basic-attack") != 3 {
		t.Errorf("Skill level should remain 3, got %d", comp.GetSkillLevel("basic-attack"))
	}
}

// TestSkillTreeComponent_LearnSkill_InsufficientPoints tests learning with insufficient points
func TestSkillTreeComponent_LearnSkill_InsufficientPoints(t *testing.T) {
	skill := &skills.Skill{
		ID:       "expensive-skill",
		Name:     "Expensive Skill",
		MaxLevel: 1,
		Requirements: skills.Requirements{
			SkillPoints: 5,
		},
	}

	tree := &skills.SkillTree{
		ID: "test-tree",
		Nodes: []*skills.SkillNode{
			{Skill: skill},
		},
	}

	comp := NewSkillTreeComponent(tree)

	// Try to learn with insufficient points
	if comp.LearnSkill("expensive-skill", 3) {
		t.Error("LearnSkill should fail with insufficient points")
	}

	// Verify skill was not learned
	if comp.IsSkillLearned("expensive-skill") {
		t.Error("Skill should not be learned")
	}
	if comp.TotalPointsUsed != 0 {
		t.Errorf("TotalPointsUsed = %d, want 0", comp.TotalPointsUsed)
	}
}

// TestSkillTreeComponent_LearnSkill_NonexistentSkill tests learning nonexistent skill
func TestSkillTreeComponent_LearnSkill_NonexistentSkill(t *testing.T) {
	tree := &skills.SkillTree{
		ID:    "empty-tree",
		Nodes: []*skills.SkillNode{},
	}

	comp := NewSkillTreeComponent(tree)

	if comp.LearnSkill("nonexistent", 10) {
		t.Error("LearnSkill should fail for nonexistent skill")
	}
}

// TestSkillTreeComponent_UnlearnSkill tests skill unlearning
func TestSkillTreeComponent_UnlearnSkill(t *testing.T) {
	skill := &skills.Skill{
		ID:       "basic-attack",
		Name:     "Basic Attack",
		MaxLevel: 3,
		Requirements: skills.Requirements{
			SkillPoints: 1,
		},
	}

	tree := &skills.SkillTree{
		ID: "warrior-tree",
		Nodes: []*skills.SkillNode{
			{Skill: skill},
		},
	}

	comp := NewSkillTreeComponent(tree)

	// Learn skill first
	comp.LearnSkill("basic-attack", 5)
	comp.LearnSkill("basic-attack", 5)

	// Verify learned
	if comp.GetSkillLevel("basic-attack") != 2 {
		t.Fatalf("Skill level should be 2, got %d", comp.GetSkillLevel("basic-attack"))
	}
	if comp.TotalPointsUsed != 2 {
		t.Fatalf("TotalPointsUsed should be 2, got %d", comp.TotalPointsUsed)
	}

	// Unlearn skill
	refunded := comp.UnlearnSkill("basic-attack")

	if refunded != 2 {
		t.Errorf("Refunded points = %d, want 2", refunded)
	}
	if comp.IsSkillLearned("basic-attack") {
		t.Error("Skill should not be learned after unlearning")
	}
	if comp.GetSkillLevel("basic-attack") != 0 {
		t.Errorf("Skill level = %d, want 0", comp.GetSkillLevel("basic-attack"))
	}
	if comp.TotalPointsUsed != 0 {
		t.Errorf("TotalPointsUsed = %d, want 0", comp.TotalPointsUsed)
	}
}

// TestSkillTreeComponent_UnlearnSkill_Nonexistent tests unlearning nonexistent skill
func TestSkillTreeComponent_UnlearnSkill_Nonexistent(t *testing.T) {
	tree := &skills.SkillTree{
		ID:    "test-tree",
		Nodes: []*skills.SkillNode{},
	}

	comp := NewSkillTreeComponent(tree)

	refunded := comp.UnlearnSkill("nonexistent")
	if refunded != 0 {
		t.Errorf("Refunded points = %d, want 0", refunded)
	}
}

// TestExperienceComponent_Integration tests full XP and leveling workflow
func TestExperienceComponent_Integration(t *testing.T) {
	exp := NewExperienceComponent()
	scaling := NewLevelScalingComponent()

	// Gain XP gradually
	exp.AddXP(50)
	if exp.CanLevelUp() {
		t.Error("Should not be able to level up yet")
	}

	// Gain enough XP to level up
	exp.AddXP(60)
	if !exp.CanLevelUp() {
		t.Error("Should be able to level up")
	}

	// Level up manually (would be done by progression system)
	exp.Level++
	exp.CurrentXP -= exp.RequiredXP
	exp.RequiredXP = 150 // Next level requires more
	exp.SkillPoints++

	// Verify new level
	if exp.Level != 2 {
		t.Errorf("Level = %d, want 2", exp.Level)
	}

	// Calculate stats for new level
	health := scaling.CalculateHealthForLevel(exp.Level)
	if health != 110.0 {
		t.Errorf("Health at level 2 = %v, want 110.0", health)
	}
}

// TestSkillTreeComponent_FullWorkflow tests full skill tree workflow
func TestSkillTreeComponent_FullWorkflow(t *testing.T) {
	// Create a small skill tree
	basicSkill := &skills.Skill{
		ID:       "basic",
		Name:     "Basic Skill",
		MaxLevel: 1,
		Requirements: skills.Requirements{
			SkillPoints: 1,
		},
	}

	advancedSkill := &skills.Skill{
		ID:       "advanced",
		Name:     "Advanced Skill",
		MaxLevel: 1,
		Requirements: skills.Requirements{
			SkillPoints:     2,
			PrerequisiteIDs: []string{"basic"},
		},
	}

	tree := &skills.SkillTree{
		ID: "test-tree",
		Nodes: []*skills.SkillNode{
			{Skill: basicSkill},
			{Skill: advancedSkill},
		},
	}

	comp := NewSkillTreeComponent(tree)

	// Learn basic skill
	if !comp.LearnSkill("basic", 5) {
		t.Error("Should be able to learn basic skill")
	}

	// Try to learn advanced skill (should work since basic is learned)
	if !comp.LearnSkill("advanced", 5) {
		t.Error("Should be able to learn advanced skill after prerequisite")
	}

	// Try to unlearn basic (should fail due to dependency)
	refunded := comp.UnlearnSkill("basic")
	if refunded != 0 {
		t.Error("Should not be able to unlearn skill with dependencies")
	}

	// Unlearn advanced first
	refunded = comp.UnlearnSkill("advanced")
	if refunded != 2 {
		t.Errorf("Should refund 2 points, got %d", refunded)
	}

	// Now unlearn basic (should work)
	refunded = comp.UnlearnSkill("basic")
	if refunded != 1 {
		t.Errorf("Should refund 1 point, got %d", refunded)
	}

	// Verify all unlearned
	if comp.TotalPointsUsed != 0 {
		t.Errorf("TotalPointsUsed = %d, want 0", comp.TotalPointsUsed)
	}
}
