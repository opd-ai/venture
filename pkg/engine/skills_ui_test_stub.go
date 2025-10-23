//go:build test
// +build test

// Package engine provides the skills UI types for testing.
package engine

import "github.com/opd-ai/venture/pkg/procgen/skills"

// NodeState represents the purchase/lock state of a skill node.
type NodeState int

const (
	NodeStateLocked NodeState = iota
	NodeStateUnlocked
	NodeStatePurchased
)

// Point is a 2D screen coordinate.
type Point struct {
	X, Y int
}

// SkillsUI stub for testing (full implementation in skills_ui.go).
type SkillsUI struct {
	visible       bool
	world         *World
	playerEntity  *Entity
	screenWidth   int
	screenHeight  int
	skillTreeComp *SkillTreeComponent
	selectedNode  *skills.SkillNode
	hoveredNode   *skills.SkillNode
	nodeSize      int
	nodeSpacing   int
	treeOffsetX   int
	treeOffsetY   int
	nodePositions map[string]Point
}

// NewSkillsUI creates a new skills UI system stub for testing.
func NewSkillsUI(world *World, screenWidth, screenHeight int) *SkillsUI {
	return &SkillsUI{
		visible:       false,
		world:         world,
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		nodeSize:      40,
		nodeSpacing:   100,
		treeOffsetX:   100,
		treeOffsetY:   100,
		nodePositions: make(map[string]Point),
	}
}

// SetPlayerEntity sets the player entity whose skill tree to display.
func (ui *SkillsUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Toggle shows or hides the skills UI.
func (ui *SkillsUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the skills UI is currently shown.
func (ui *SkillsUI) IsVisible() bool {
	return ui.visible
}

// Show displays the skills UI.
func (ui *SkillsUI) Show() {
	ui.visible = true
}

// Hide hides the skills UI.
func (ui *SkillsUI) Hide() {
	ui.visible = false
}

// Update processes input for the skills UI (stub).
func (ui *SkillsUI) Update(deltaTime float64) {
	// Stub for testing
}
