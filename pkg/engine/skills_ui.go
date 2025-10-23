//go:build !test
// +build !test

// Package engine provides skills tree UI rendering.
// This file implements SkillsUI which displays skill tree progression,
// allows skill point spending, and visualizes skill node relationships.
package engine

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"golang.org/x/image/font/basicfont"
)

// NodeState represents the purchase/lock state of a skill node.
type NodeState int

const (
	// NodeStateLocked means prerequisites are not met
	NodeStateLocked NodeState = iota
	// NodeStateUnlocked means available for purchase
	NodeStateUnlocked
	// NodeStatePurchased means already purchased
	NodeStatePurchased
)

// Point is a 2D screen coordinate.
type Point struct {
	X, Y int
}

// SkillsUI handles rendering and interaction for the skill tree screen.
type SkillsUI struct {
	visible      bool
	world        *World
	playerEntity *Entity
	screenWidth  int
	screenHeight int

	// Skill tree data
	skillTreeComp *SkillTreeComponent
	selectedNode  *skills.SkillNode
	hoveredNode   *skills.SkillNode

	// Layout
	nodeSize      int              // Diameter of skill node circle
	nodeSpacing   int              // Space between nodes
	treeOffsetX   int              // X offset for centering
	treeOffsetY   int              // Y offset for header
	nodePositions map[string]Point // Cache of node screen positions
}

// NewSkillsUI creates a new skills UI system.
// Parameters:
//
//	world - ECS world instance
//	screenWidth, screenHeight - Display dimensions
//
// Returns: Initialized SkillsUI
// Called by: Game.NewGame() during initialization
func NewSkillsUI(world *World, screenWidth, screenHeight int) *SkillsUI {
	return &SkillsUI{
		visible:       false,
		world:         world,
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		nodeSize:      40,  // 40px diameter circles
		nodeSpacing:   100, // 100px between node centers
		treeOffsetX:   100, // Left margin
		treeOffsetY:   100, // Top margin (below header)
		nodePositions: make(map[string]Point),
	}
}

// SetPlayerEntity sets the player entity whose skill tree to display.
// Parameters:
//
//	entity - Player entity with SkillTreeComponent
//
// Called by: Game.SetPlayerEntity()
func (ui *SkillsUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
	if ui.visible {
		ui.loadSkillTree()
		ui.calculateNodeLayout()
	}
}

// Toggle shows or hides the skills UI.
// Called by: InputSystem when K key is pressed
func (ui *SkillsUI) Toggle() {
	ui.visible = !ui.visible
	if ui.visible {
		ui.loadSkillTree()
		ui.calculateNodeLayout()
	}
}

// IsVisible returns whether the skills UI is currently shown.
// Returns: true if visible, false otherwise
// Called by: Game.Update() to block input
func (ui *SkillsUI) IsVisible() bool {
	return ui.visible
}

// Show displays the skills UI.
func (ui *SkillsUI) Show() {
	ui.visible = true
	ui.loadSkillTree()
	ui.calculateNodeLayout()
}

// Hide hides the skills UI.
func (ui *SkillsUI) Hide() {
	ui.visible = false
	ui.selectedNode = nil
	ui.hoveredNode = nil
}

// loadSkillTree loads the skill tree component from the player entity.
func (ui *SkillsUI) loadSkillTree() {
	if ui.playerEntity == nil {
		return
	}

	if comp, ok := ui.playerEntity.GetComponent("skill_tree"); ok {
		ui.skillTreeComp = comp.(*SkillTreeComponent)
	}
}

// Update processes input for the skills UI.
// Parameters:
//
//	deltaTime - Time since last frame
//
// Called by: Game.Update() every frame
func (ui *SkillsUI) Update(deltaTime float64) {
	if !ui.visible {
		return
	}

	// Handle ESC key to close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Hide()
		return
	}

	// Handle K key to toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		ui.Toggle()
		return
	}

	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()

	// Find hovered node
	ui.hoveredNode = ui.findNodeAtPosition(mouseX, mouseY)

	// Handle left click to purchase skill
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && ui.hoveredNode != nil {
		ui.attemptPurchaseSkill(ui.hoveredNode.Skill.ID)
	}

	// Handle right click to refund skill (if implemented)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && ui.hoveredNode != nil {
		ui.attemptRefundSkill(ui.hoveredNode.Skill.ID)
	}
}

// Draw renders the skills UI overlay.
// Parameters:
//
//	screen - Ebiten image
//
// Called by: Game.Draw() every frame
func (ui *SkillsUI) Draw(screen *ebiten.Image) {
	if !ui.visible || ui.playerEntity == nil || ui.skillTreeComp == nil {
		return
	}

	// Draw semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 180}, false)

	// Draw main panel background
	panelWidth := 800
	panelHeight := 600
	if ui.screenWidth < 800 {
		panelWidth = ui.screenWidth - 40
	}
	if ui.screenHeight < 600 {
		panelHeight = ui.screenHeight - 40
	}

	panelX := (ui.screenWidth - panelWidth) / 2
	panelY := (ui.screenHeight - panelHeight) / 2

	vector.DrawFilledRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 30, 255}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)

	// Title bar
	titleText := "SKILL TREE"
	if ui.skillTreeComp.Tree != nil {
		titleText = ui.skillTreeComp.Tree.Name + " - SKILL TREE"
	}
	titleX := panelX + panelWidth/2 - len(titleText)*3
	titleY := panelY + 20
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY+13,
		color.RGBA{255, 255, 100, 255})

	// Display available skill points
	availablePoints := ui.getAvailableSkillPoints()
	pointsText := fmt.Sprintf("Skill Points: %d", availablePoints)
	text.Draw(screen, pointsText, basicfont.Face7x13, panelX+panelWidth-150, titleY+13,
		color.RGBA{100, 255, 100, 255})

	// Ensure node layout is calculated
	if len(ui.nodePositions) == 0 {
		ui.calculateNodeLayout()
	}

	// Draw skill tree
	ui.drawSkillTree(screen, panelX, panelY, panelWidth, panelHeight)

	// Draw tooltip for hovered node
	if ui.hoveredNode != nil {
		mouseX, mouseY := ebiten.CursorPosition()
		ui.drawSkillTooltip(screen, ui.hoveredNode, mouseX, mouseY)
	}

	// Draw controls hint
	controlsText := "[Left Click] Purchase | [Right Click] Refund | [ESC] or [K] Close"
	controlsX := panelX + 10
	controlsY := panelY + panelHeight - 20
	text.Draw(screen, controlsText, basicfont.Face7x13, controlsX, controlsY,
		color.RGBA{180, 180, 180, 255})
}

// calculateNodeLayout computes screen positions for all skill nodes.
// Uses tree structure to arrange nodes in tiers (rows).
func (ui *SkillsUI) calculateNodeLayout() {
	if ui.skillTreeComp == nil || ui.skillTreeComp.Tree == nil {
		return
	}

	// Clear existing positions
	ui.nodePositions = make(map[string]Point)

	// Group nodes by tier
	tiers := make(map[skills.Tier][]*skills.SkillNode)
	for _, node := range ui.skillTreeComp.Tree.Nodes {
		tier := node.Skill.Tier
		tiers[tier] = append(tiers[tier], node)
	}

	// Layout each tier
	tierY := ui.treeOffsetY
	tierIndex := 0

	for tier := skills.TierBasic; tier <= skills.TierMaster; tier++ {
		nodesInTier := tiers[tier]
		if len(nodesInTier) == 0 {
			continue
		}

		// Calculate horizontal spacing for this tier
		startX := ui.treeOffsetX

		for i, node := range nodesInTier {
			x := startX + (i * ui.nodeSpacing)
			y := tierY + (tierIndex * ui.nodeSpacing)

			ui.nodePositions[node.Skill.ID] = Point{X: x, Y: y}
		}

		tierIndex++
	}
}

// drawSkillTree renders all skill nodes and connections.
func (ui *SkillsUI) drawSkillTree(screen *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	if ui.skillTreeComp.Tree == nil {
		return
	}

	// First pass: draw connections
	for _, node := range ui.skillTreeComp.Tree.Nodes {
		ui.drawNodeConnections(screen, node, ui.nodePositions)
	}

	// Second pass: draw nodes
	availablePoints := ui.getAvailableSkillPoints()
	playerLevel := ui.getPlayerLevel()

	for _, node := range ui.skillTreeComp.Tree.Nodes {
		pos, exists := ui.nodePositions[node.Skill.ID]
		if !exists {
			continue
		}

		state := ui.getNodeState(node, playerLevel, availablePoints)
		ui.drawSkillNode(screen, node, pos.X, pos.Y, state)
	}
}

// drawSkillNode renders a single skill node.
// Parameters:
//
//	screen - Target image
//	node - Skill node data
//	x, y - Center position
//	state - Locked/Unlocked/Purchased state
func (ui *SkillsUI) drawSkillNode(screen *ebiten.Image, node *skills.SkillNode, x, y int, state NodeState) {
	radius := float32(ui.nodeSize / 2)

	// Determine node color based on state
	var nodeColor color.Color
	switch state {
	case NodeStateLocked:
		nodeColor = color.RGBA{100, 100, 100, 255} // Gray
	case NodeStateUnlocked:
		nodeColor = color.RGBA{100, 150, 255, 255} // Blue
	case NodeStatePurchased:
		nodeColor = color.RGBA{100, 255, 100, 255} // Green
	}

	// Draw node circle
	vector.DrawFilledCircle(screen, float32(x), float32(y), radius, nodeColor, false)

	// Draw border
	borderColor := color.RGBA{255, 255, 255, 255}
	if ui.hoveredNode != nil && ui.hoveredNode.Skill.ID == node.Skill.ID {
		borderColor = color.RGBA{255, 255, 100, 255} // Yellow highlight
	}
	vector.StrokeCircle(screen, float32(x), float32(y), radius, 2, borderColor, false)

	// Draw skill level indicator (if purchased)
	if state == NodeStatePurchased {
		level := ui.skillTreeComp.GetSkillLevel(node.Skill.ID)
		maxLevel := node.Skill.MaxLevel
		levelText := fmt.Sprintf("%d/%d", level, maxLevel)
		textX := x - len(levelText)*3
		textY := y + 5
		text.Draw(screen, levelText, basicfont.Face7x13, textX, textY,
			color.RGBA{255, 255, 255, 255})
	} else {
		// Draw first letter of skill name
		if len(node.Skill.Name) > 0 {
			letter := string(node.Skill.Name[0])
			textX := x - 4
			textY := y + 5
			text.Draw(screen, letter, basicfont.Face7x13, textX, textY,
				color.RGBA{255, 255, 255, 255})
		}
	}
}

// drawNodeConnections renders lines between prerequisite nodes.
// Parameters:
//
//	screen - Target image
//	node - Current node
//	nodePositions - Map of node ID to screen position
func (ui *SkillsUI) drawNodeConnections(screen *ebiten.Image, node *skills.SkillNode, nodePositions map[string]Point) {
	currentPos, exists := nodePositions[node.Skill.ID]
	if !exists {
		return
	}

	// Draw lines to prerequisites
	for _, prereqID := range node.Skill.Requirements.PrerequisiteIDs {
		prereqPos, prereqExists := nodePositions[prereqID]
		if !prereqExists {
			continue
		}

		// Determine line color based on learning state
		lineColor := color.RGBA{100, 100, 100, 255} // Gray for locked
		if ui.skillTreeComp.IsSkillLearned(prereqID) {
			lineColor = color.RGBA{100, 255, 100, 200} // Green for learned
		}

		// Draw line from prerequisite to current node
		vector.StrokeLine(screen,
			float32(prereqPos.X), float32(prereqPos.Y),
			float32(currentPos.X), float32(currentPos.Y),
			2, lineColor, false)
	}
}

// drawSkillTooltip renders detailed skill information on hover.
// Parameters:
//
//	screen - Target image
//	node - Hovered skill node
//	mouseX, mouseY - Mouse position for tooltip placement
func (ui *SkillsUI) drawSkillTooltip(screen *ebiten.Image, node *skills.SkillNode, mouseX, mouseY int) {
	skill := node.Skill

	// Calculate tooltip size
	tooltipWidth := 250
	tooltipHeight := 150

	// Position tooltip near mouse (avoid screen edges)
	tooltipX := mouseX + 20
	tooltipY := mouseY + 20

	if tooltipX+tooltipWidth > ui.screenWidth {
		tooltipX = mouseX - tooltipWidth - 20
	}
	if tooltipY+tooltipHeight > ui.screenHeight {
		tooltipY = ui.screenHeight - tooltipHeight - 10
	}

	// Draw tooltip background
	vector.DrawFilledRect(screen, float32(tooltipX), float32(tooltipY),
		float32(tooltipWidth), float32(tooltipHeight),
		color.RGBA{10, 10, 20, 240}, false)
	vector.StrokeRect(screen, float32(tooltipX), float32(tooltipY),
		float32(tooltipWidth), float32(tooltipHeight), 2,
		color.RGBA{200, 200, 255, 255}, false)

	// Draw skill name (header)
	y := tooltipY + 15
	text.Draw(screen, skill.Name, basicfont.Face7x13, tooltipX+10, y,
		color.RGBA{255, 255, 100, 255})
	y += 20

	// Draw skill type and category
	typeText := fmt.Sprintf("Type: %s | Cat: %s", skill.Type.String(), skill.Category.String())
	text.Draw(screen, typeText, basicfont.Face7x13, tooltipX+10, y,
		color.RGBA{180, 180, 255, 255})
	y += 20

	// Draw description (wrapped)
	desc := skill.Description
	if len(desc) > 35 {
		desc = desc[:32] + "..."
	}
	text.Draw(screen, desc, basicfont.Face7x13, tooltipX+10, y,
		color.RGBA{200, 200, 200, 255})
	y += 20

	// Draw cost
	costText := fmt.Sprintf("Cost: %d skill points", skill.Requirements.SkillPoints)
	text.Draw(screen, costText, basicfont.Face7x13, tooltipX+10, y,
		color.RGBA{255, 200, 100, 255})
	y += 20

	// Draw prerequisites
	if len(skill.Requirements.PrerequisiteIDs) > 0 {
		prereqText := fmt.Sprintf("Requires: %d skills", len(skill.Requirements.PrerequisiteIDs))
		text.Draw(screen, prereqText, basicfont.Face7x13, tooltipX+10, y,
			color.RGBA{255, 150, 150, 255})
		y += 20
	}

	// Draw action hint
	state := ui.getNodeState(node, ui.getPlayerLevel(), ui.getAvailableSkillPoints())
	if state == NodeStateUnlocked {
		hintText := "Click to purchase"
		text.Draw(screen, hintText, basicfont.Face7x13, tooltipX+10, y,
			color.RGBA{100, 255, 100, 255})
	} else if state == NodeStatePurchased {
		hintText := "Right-click to refund"
		text.Draw(screen, hintText, basicfont.Face7x13, tooltipX+10, y,
			color.RGBA{255, 200, 100, 255})
	} else {
		hintText := "Locked"
		text.Draw(screen, hintText, basicfont.Face7x13, tooltipX+10, y,
			color.RGBA{255, 100, 100, 255})
	}
}

// attemptPurchaseSkill attempts to purchase the selected skill node.
// Parameters:
//
//	skillID - ID of skill to purchase
func (ui *SkillsUI) attemptPurchaseSkill(skillID string) {
	if ui.skillTreeComp == nil || ui.playerEntity == nil {
		return
	}

	// Get available skill points
	availablePoints := ui.getAvailableSkillPoints()

	// Attempt to learn the skill
	if ui.skillTreeComp.LearnSkill(skillID, availablePoints) {
		// Deduct skill points from experience component
		if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
			exp := expComp.(*ExperienceComponent)
			skill := ui.skillTreeComp.Tree.GetSkillByID(skillID)
			if skill != nil {
				exp.SkillPoints -= skill.Requirements.SkillPoints
			}
		}
	}
}

// attemptRefundSkill refunds a purchased skill.
// Parameters:
//
//	skillID - ID of skill to refund
func (ui *SkillsUI) attemptRefundSkill(skillID string) {
	if ui.skillTreeComp == nil || ui.playerEntity == nil {
		return
	}

	// Attempt to unlearn the skill
	pointsRefunded := ui.skillTreeComp.UnlearnSkill(skillID)

	if pointsRefunded > 0 {
		// Refund skill points to experience component
		if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
			exp := expComp.(*ExperienceComponent)
			exp.SkillPoints += pointsRefunded
		}
	}
}

// getNodeState determines if a node is locked/unlocked/purchased.
// Parameters:
//
//	node - Skill node to check
//	playerLevel - Player's current level
//	availablePoints - Available skill points
//
// Returns: NodeState enum value
func (ui *SkillsUI) getNodeState(node *skills.SkillNode, playerLevel, availablePoints int) NodeState {
	if ui.skillTreeComp.IsSkillLearned(node.Skill.ID) {
		return NodeStatePurchased
	}

	// Check if prerequisites are met
	for _, prereqID := range node.Skill.Requirements.PrerequisiteIDs {
		if !ui.skillTreeComp.IsSkillLearned(prereqID) {
			return NodeStateLocked
		}
	}

	// Check other requirements
	if !node.Skill.IsUnlocked(playerLevel, availablePoints, ui.skillTreeComp.LearnedSkills, ui.skillTreeComp.Attributes) {
		return NodeStateLocked
	}

	return NodeStateUnlocked
}

// findNodeAtPosition returns the node at the given screen position.
// Parameters:
//
//	x, y - Screen coordinates
//
// Returns: Node at position or nil
func (ui *SkillsUI) findNodeAtPosition(x, y int) *skills.SkillNode {
	if ui.skillTreeComp == nil || ui.skillTreeComp.Tree == nil {
		return nil
	}

	radius := ui.nodeSize / 2

	for _, node := range ui.skillTreeComp.Tree.Nodes {
		pos, exists := ui.nodePositions[node.Skill.ID]
		if !exists {
			continue
		}

		// Calculate distance from cursor to node center
		dx := float64(x - pos.X)
		dy := float64(y - pos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance <= float64(radius) {
			return node
		}
	}

	return nil
}

// getAvailableSkillPoints returns the number of unspent skill points.
func (ui *SkillsUI) getAvailableSkillPoints() int {
	if ui.playerEntity == nil {
		return 0
	}

	if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
		return expComp.(*ExperienceComponent).SkillPoints
	}

	return 0
}

// getPlayerLevel returns the player's current level.
func (ui *SkillsUI) getPlayerLevel() int {
	if ui.playerEntity == nil {
		return 1
	}

	if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
		return expComp.(*ExperienceComponent).Level
	}

	return 1
}
