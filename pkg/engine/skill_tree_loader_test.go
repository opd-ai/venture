package engine

import (
	"testing"
)

// TestLoadPlayerSkillTree verifies skill tree generation and loading.
func TestLoadPlayerSkillTree(t *testing.T) {
	player := NewEntity(1)
	seed := int64(12345)
	genreID := "fantasy"
	depth := 0

	err := LoadPlayerSkillTree(player, seed, genreID, depth)
	if err != nil {
		t.Fatalf("LoadPlayerSkillTree failed: %v", err)
	}

	// Verify skill tree component was added
	if !player.HasComponent("skill_tree") {
		t.Fatal("Skill tree component not added to player")
	}

	comp, ok := player.GetComponent("skill_tree")
	if !ok {
		t.Fatal("GetComponent failed for skill_tree")
	}

	treeComp, ok := comp.(*SkillTreeComponent)
	if !ok {
		t.Fatal("Component is not SkillTreeComponent")
	}

	// Verify tree was populated
	if treeComp.Tree == nil {
		t.Fatal("Skill tree is nil")
	}

	if len(treeComp.Tree.Nodes) == 0 {
		t.Fatal("Skill tree has no nodes")
	}

	if len(treeComp.Tree.RootNodes) == 0 {
		t.Fatal("Skill tree has no root nodes")
	}

	t.Logf("Loaded skill tree '%s' with %d skills", treeComp.Tree.Name, len(treeComp.Tree.Nodes))
}

// TestLoadPlayerSkillTree_Deterministic verifies same seed produces same tree.
func TestLoadPlayerSkillTree_Deterministic(t *testing.T) {
	seed := int64(67890)
	genreID := "scifi"
	depth := 5

	// Generate first tree
	player1 := NewEntity(1)
	err1 := LoadPlayerSkillTree(player1, seed, genreID, depth)
	if err1 != nil {
		t.Fatalf("First LoadPlayerSkillTree failed: %v", err1)
	}

	comp1, _ := player1.GetComponent("skill_tree")
	tree1 := comp1.(*SkillTreeComponent).Tree

	// Generate second tree with same seed
	player2 := NewEntity(2)
	err2 := LoadPlayerSkillTree(player2, seed, genreID, depth)
	if err2 != nil {
		t.Fatalf("Second LoadPlayerSkillTree failed: %v", err2)
	}

	comp2, _ := player2.GetComponent("skill_tree")
	tree2 := comp2.(*SkillTreeComponent).Tree

	// Verify trees are identical
	if tree1.Name != tree2.Name {
		t.Errorf("Tree names differ: %s vs %s", tree1.Name, tree2.Name)
	}

	if len(tree1.Nodes) != len(tree2.Nodes) {
		t.Errorf("Tree node counts differ: %d vs %d", len(tree1.Nodes), len(tree2.Nodes))
	}

	// Verify first skill in each tree matches
	if len(tree1.Nodes) > 0 && len(tree2.Nodes) > 0 {
		skill1 := tree1.Nodes[0].Skill
		skill2 := tree2.Nodes[0].Skill
		if skill1.Name != skill2.Name {
			t.Errorf("First skill names differ: %s vs %s", skill1.Name, skill2.Name)
		}
	}
}

// TestLoadPlayerSkillTree_GenreVariation verifies different genres produce different trees.
func TestLoadPlayerSkillTree_GenreVariation(t *testing.T) {
	seed := int64(11111)
	depth := 0

	// Generate fantasy tree
	playerFantasy := NewEntity(1)
	err := LoadPlayerSkillTree(playerFantasy, seed, "fantasy", depth)
	if err != nil {
		t.Fatalf("Fantasy LoadPlayerSkillTree failed: %v", err)
	}

	comp, _ := playerFantasy.GetComponent("skill_tree")
	fantasyTree := comp.(*SkillTreeComponent).Tree

	// Generate scifi tree
	playerScifi := NewEntity(2)
	err = LoadPlayerSkillTree(playerScifi, seed, "scifi", depth)
	if err != nil {
		t.Fatalf("Scifi LoadPlayerSkillTree failed: %v", err)
	}

	comp, _ = playerScifi.GetComponent("skill_tree")
	scifiTree := comp.(*SkillTreeComponent).Tree

	// Verify trees differ (at least in name or genre)
	if fantasyTree.Name == scifiTree.Name && fantasyTree.Genre == scifiTree.Genre {
		t.Error("Fantasy and scifi trees are identical (expected variation)")
	}

	t.Logf("Fantasy tree: %s (genre: %s)", fantasyTree.Name, fantasyTree.Genre)
	t.Logf("Scifi tree: %s (genre: %s)", scifiTree.Name, scifiTree.Genre)
}

// TestGetPlayerSkillPoints verifies skill point calculation.
func TestGetPlayerSkillPoints(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 0},   // Level 1: 0 points
		{2, 1},   // Level 2: 1 point
		{5, 4},   // Level 5: 4 points
		{10, 11}, // Level 10: 9 base + 2 bonus = 11
		{11, 12}, // Level 11: 10 base + 2 bonus = 12
		{20, 23}, // Level 20: 19 base + 4 bonus = 23
		{30, 35}, // Level 30: 29 base + 6 bonus = 35
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.level)), func(t *testing.T) {
			points := GetPlayerSkillPoints(tt.level)
			if points != tt.expected {
				t.Errorf("GetPlayerSkillPoints(%d) = %d, expected %d", tt.level, points, tt.expected)
			}
		})
	}
}

// TestGetUnspentSkillPoints verifies unspent points calculation.
func TestGetUnspentSkillPoints(t *testing.T) {
	// Create player with level 5 (4 total points)
	player := NewEntity(1)
	player.AddComponent(&ExperienceComponent{
		Level:     5,
		CurrentXP: 500,
	})

	// Load skill tree
	err := LoadPlayerSkillTree(player, 12345, "fantasy", 0)
	if err != nil {
		t.Fatalf("LoadPlayerSkillTree failed: %v", err)
	}

	// Initially, all points should be unspent
	unspent := GetUnspentSkillPoints(player)
	expected := 4 // Level 5 = 4 points
	if unspent != expected {
		t.Errorf("Initial unspent points = %d, expected %d", unspent, expected)
	}

	// Spend 2 points by manually setting TotalPointsUsed
	comp, _ := player.GetComponent("skill_tree")
	treeComp := comp.(*SkillTreeComponent)
	treeComp.TotalPointsUsed = 2

	// Check unspent points again
	unspent = GetUnspentSkillPoints(player)
	expected = 2 // 4 total - 2 used = 2 unspent
	if unspent != expected {
		t.Errorf("After spending 2 points, unspent = %d, expected %d", unspent, expected)
	}
}

// TestGetUnspentSkillPoints_NoExperience verifies behavior without experience component.
func TestGetUnspentSkillPoints_NoExperience(t *testing.T) {
	player := NewEntity(1)

	// Load skill tree (no experience component)
	err := LoadPlayerSkillTree(player, 12345, "fantasy", 0)
	if err != nil {
		t.Fatalf("LoadPlayerSkillTree failed: %v", err)
	}

	// Should default to level 1 (0 points)
	unspent := GetUnspentSkillPoints(player)
	if unspent != 0 {
		t.Errorf("Unspent points without experience = %d, expected 0", unspent)
	}
}

// TestLoadPlayerSkillTree_UpdateExisting verifies updating existing tree.
func TestLoadPlayerSkillTree_UpdateExisting(t *testing.T) {
	player := NewEntity(1)

	// Load initial tree
	err := LoadPlayerSkillTree(player, 12345, "fantasy", 0)
	if err != nil {
		t.Fatalf("Initial LoadPlayerSkillTree failed: %v", err)
	}

	comp, _ := player.GetComponent("skill_tree")
	tree1 := comp.(*SkillTreeComponent).Tree
	tree1Name := tree1.Name

	// Load new tree (different seed)
	err = LoadPlayerSkillTree(player, 67890, "scifi", 5)
	if err != nil {
		t.Fatalf("Second LoadPlayerSkillTree failed: %v", err)
	}

	comp, _ = player.GetComponent("skill_tree")
	tree2 := comp.(*SkillTreeComponent).Tree
	tree2Name := tree2.Name

	// Verify tree was updated (should be different)
	if tree1Name == tree2Name && tree1.Genre == tree2.Genre {
		t.Error("Tree was not updated (names and genres match)")
	}

	t.Logf("Updated tree from '%s' to '%s'", tree1Name, tree2Name)
}

// TestSkillTreeComponent_Integration verifies full integration workflow.
func TestSkillTreeComponent_Integration(t *testing.T) {
	// Create player with level 5
	player := NewEntity(1)
	player.AddComponent(&ExperienceComponent{
		Level:     5,
		CurrentXP: 500,
	})

	// Load skill tree
	err := LoadPlayerSkillTree(player, 12345, "fantasy", 0)
	if err != nil {
		t.Fatalf("LoadPlayerSkillTree failed: %v", err)
	}

	// Get skill tree component
	comp, ok := player.GetComponent("skill_tree")
	if !ok {
		t.Fatal("Skill tree component not found")
	}
	treeComp := comp.(*SkillTreeComponent)

	// Get unspent points
	unspent := GetUnspentSkillPoints(player)
	t.Logf("Player level %d has %d unspent skill points", 5, unspent)

	// Try to learn a root skill (should have prerequisites met)
	if len(treeComp.Tree.RootNodes) == 0 {
		t.Fatal("No root nodes to test learning")
	}

	rootSkill := treeComp.Tree.RootNodes[0].Skill
	success := treeComp.LearnSkill(rootSkill.ID, unspent)
	if !success {
		t.Errorf("Failed to learn root skill '%s'", rootSkill.Name)
	}

	// Verify skill was learned
	if !treeComp.IsSkillLearned(rootSkill.ID) {
		t.Error("Skill not marked as learned")
	}

	if treeComp.GetSkillLevel(rootSkill.ID) != 1 {
		t.Error("Skill level not incremented")
	}

	// Verify points were spent
	newUnspent := GetUnspentSkillPoints(player)
	if newUnspent >= unspent {
		t.Error("Skill points were not spent")
	}

	t.Logf("Successfully learned skill '%s', remaining points: %d", rootSkill.Name, newUnspent)
}
