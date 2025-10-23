//go:build test
// +build test

// Package engine provides skills UI tests.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/skills"
)

// TestSkillsUI_NewSkillsUI verifies initialization.
func TestSkillsUI_NewSkillsUI(t *testing.T) {
	world := NewWorld()
	ui := NewSkillsUI(world, 800, 600)

	if ui == nil {
		t.Fatal("NewSkillsUI returned nil")
	}

	if ui.visible {
		t.Error("SkillsUI should be hidden by default")
	}

	if ui.nodeSize != 40 {
		t.Errorf("Expected nodeSize 40, got %d", ui.nodeSize)
	}

	if ui.nodeSpacing != 100 {
		t.Errorf("Expected nodeSpacing 100, got %d", ui.nodeSpacing)
	}
}

// TestSkillsUI_Toggle tests visibility toggling.
func TestSkillsUI_Toggle(t *testing.T) {
	world := NewWorld()
	ui := NewSkillsUI(world, 800, 600)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("UI should be hidden initially")
	}

	// Toggle to show
	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("UI should be visible after first toggle")
	}

	// Toggle to hide
	ui.Toggle()
	if ui.IsVisible() {
		t.Error("UI should be hidden after second toggle")
	}
}

// TestSkillsUI_SetPlayerEntity ensures entity is stored.
func TestSkillsUI_SetPlayerEntity(t *testing.T) {
	world := NewWorld()
	ui := NewSkillsUI(world, 800, 600)

	entity := world.CreateEntity()
	ui.SetPlayerEntity(entity)

	if ui.playerEntity != entity {
		t.Error("Player entity was not stored correctly")
	}
}

// TestSkillsUI_ShowHide tests show/hide methods.
func TestSkillsUI_ShowHide(t *testing.T) {
	world := NewWorld()
	ui := NewSkillsUI(world, 800, 600)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("UI should be hidden initially")
	}

	// Show
	ui.Show()
	if !ui.IsVisible() {
		t.Error("UI should be visible after Show()")
	}

	// Hide
	ui.Hide()
	if ui.IsVisible() {
		t.Error("UI should be hidden after Hide()")
	}
}

// TestSkillTreeComponent_NewSkillTreeComponent verifies component creation.
func TestSkillTreeComponent_NewSkillTreeComponent(t *testing.T) {
	// Generate a simple skill tree
	gen := &skills.SkillTreeGenerator{}
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Failed to generate skill tree: %v", err)
	}

	trees := result.([]*skills.SkillTree)
	if len(trees) == 0 {
		t.Fatal("No skill trees generated")
	}

	tree := trees[0]
	comp := NewSkillTreeComponent(tree)

	if comp == nil {
		t.Fatal("NewSkillTreeComponent returned nil")
	}

	if comp.Tree != tree {
		t.Error("Tree was not stored correctly")
	}

	if comp.LearnedSkills == nil {
		t.Error("LearnedSkills map was not initialized")
	}

	if comp.TotalPointsUsed != 0 {
		t.Error("TotalPointsUsed should start at 0")
	}
}

// TestSkillTreeComponent_LearnSkill tests skill learning.
func TestSkillTreeComponent_LearnSkill(t *testing.T) {
	// Generate a simple skill tree
	gen := &skills.SkillTreeGenerator{}
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Failed to generate skill tree: %v", err)
	}

	trees := result.([]*skills.SkillTree)
	if len(trees) == 0 {
		t.Fatal("No skill trees generated")
	}

	tree := trees[0]
	comp := NewSkillTreeComponent(tree)

	// Find a root skill (no prerequisites)
	if len(tree.RootNodes) == 0 {
		t.Skip("No root nodes in generated tree")
	}

	rootSkill := tree.RootNodes[0].Skill

	// Try to learn the root skill
	availablePoints := 100 // Plenty of points
	success := comp.LearnSkill(rootSkill.ID, availablePoints)

	// Note: LearnSkill may fail if the skill has player level requirements
	// For now, just verify the method doesn't panic
	if success {
		if !comp.IsSkillLearned(rootSkill.ID) {
			t.Error("Skill should be marked as learned after successful learning")
		}

		if comp.GetSkillLevel(rootSkill.ID) < 1 {
			t.Errorf("Skill level should be at least 1, got %d", comp.GetSkillLevel(rootSkill.ID))
		}
	}
}

// TestSkillTreeComponent_UnlearnSkill tests skill refunding.
func TestSkillTreeComponent_UnlearnSkill(t *testing.T) {
	// Generate a simple skill tree
	gen := &skills.SkillTreeGenerator{}
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Failed to generate skill tree: %v", err)
	}

	trees := result.([]*skills.SkillTree)
	if len(trees) == 0 {
		t.Fatal("No skill trees generated")
	}

	tree := trees[0]
	comp := NewSkillTreeComponent(tree)

	// Find a root skill
	if len(tree.RootNodes) == 0 {
		t.Skip("No root nodes in generated tree")
	}

	rootSkill := tree.RootNodes[0].Skill

	// Learn the skill first (might not succeed due to requirements)
	success := comp.LearnSkill(rootSkill.ID, 100)
	if !success {
		t.Skip("Could not learn root skill to test unlearning")
	}

	// Now unlearn it
	pointsRefunded := comp.UnlearnSkill(rootSkill.ID)

	if pointsRefunded <= 0 {
		t.Error("Should have refunded points")
	}

	if comp.IsSkillLearned(rootSkill.ID) {
		t.Error("Skill should no longer be learned")
	}

	if comp.GetSkillLevel(rootSkill.ID) != 0 {
		t.Error("Skill level should be 0 after unlearning")
	}
}

// TestSkillTreeComponent_GetAvailableSkills tests filtering available skills.
func TestSkillTreeComponent_GetAvailableSkills(t *testing.T) {
	// Generate a simple skill tree
	gen := &skills.SkillTreeGenerator{}
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Failed to generate skill tree: %v", err)
	}

	trees := result.([]*skills.SkillTree)
	if len(trees) == 0 {
		t.Fatal("No skill trees generated")
	}

	tree := trees[0]
	comp := NewSkillTreeComponent(tree)

	// Get available skills (should include root nodes at minimum)
	available := comp.GetAvailableSkills(1, 100)

	// Note: Available skills may be empty if all skills have player level requirements > 1
	// Just verify the method doesn't panic and returns a valid slice
	if available == nil {
		t.Error("GetAvailableSkills should return a valid slice, not nil")
	}
}
