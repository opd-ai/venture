//go:build !test
// +build !test

// Package engine provides character stats UI rendering.
// This file implements CharacterUI which displays detailed character information
// including stats, equipment, and derived attributes.
package engine

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/combat"
	"golang.org/x/image/font/basicfont"
)

// Rectangle defines a UI panel bounds.
type Rectangle struct {
	X, Y, Width, Height int
}

// CharacterUI handles rendering and interaction for the character stats screen.
type CharacterUI struct {
	visible      bool
	world        *World
	playerEntity *Entity
	screenWidth  int
	screenHeight int

	// Layout sections
	statsPanel      Rectangle
	equipmentPanel  Rectangle
	attributesPanel Rectangle
}

// NewCharacterUI creates a new character UI system.
// Parameters:
//
//	world - ECS world instance for entity queries
//	screenWidth, screenHeight - Display dimensions for layout calculation
//
// Returns: Initialized CharacterUI ready for use
// Called by: Game.NewGame() during initialization
func NewCharacterUI(world *World, screenWidth, screenHeight int) *CharacterUI {
	ui := &CharacterUI{
		visible:      false,
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
	ui.calculateLayout()
	return ui
}

// SetPlayerEntity sets the player entity whose stats to display.
// Parameters:
//
//	entity - Player entity with StatsComponent, EquipmentComponent, etc.
//
// Called by: Game.SetPlayerEntity() after player creation
func (ui *CharacterUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
	if ui.visible {
		ui.calculateLayout()
	}
}

// Toggle shows or hides the character UI.
// Called by: InputSystem when C key is pressed
func (ui *CharacterUI) Toggle() {
	ui.visible = !ui.visible
	if ui.visible {
		ui.calculateLayout()
	}
}

// IsVisible returns whether the character UI is currently shown.
// Returns: true if visible, false otherwise
// Called by: Game.Update() to determine if input should be blocked
func (ui *CharacterUI) IsVisible() bool {
	return ui.visible
}

// Show displays the character UI.
func (ui *CharacterUI) Show() {
	ui.visible = true
	ui.calculateLayout()
}

// Hide hides the character UI.
func (ui *CharacterUI) Hide() {
	ui.visible = false
}

// Update processes input for the character UI.
// Parameters:
//
//	deltaTime - Time since last frame in seconds
//
// Called by: Game.Update() every frame
func (ui *CharacterUI) Update(deltaTime float64) {
	if !ui.visible {
		return
	}

	// Handle ESC key to close UI
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Hide()
	}

	// Handle C key to toggle UI
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		ui.Toggle()
	}
}

// Draw renders the character UI overlay.
// Parameters:
//
//	screen - Ebiten image to render to
//
// Called by: Game.Draw() every frame
func (ui *CharacterUI) Draw(screen *ebiten.Image) {
	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Fetch required components
	statsComp, hasStats := ui.playerEntity.GetComponent("stats")
	equipComp, hasEquip := ui.playerEntity.GetComponent("equipment")
	expComp, hasExp := ui.playerEntity.GetComponent("experience")
	invComp, hasInv := ui.playerEntity.GetComponent("inventory")

	if !hasStats {
		return // Need stats at minimum
	}

	stats := statsComp.(*StatsComponent)
	var equipment *EquipmentComponent
	if hasEquip {
		equipment = equipComp.(*EquipmentComponent)
	}

	// Draw semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 180}, false)

	// Draw main panel background (800x600 centered)
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

	// Panel background
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 30, 255}, false)

	// Panel border
	vector.StrokeRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)

	// Title bar
	titleText := "CHARACTER STATS"
	titleX := panelX + panelWidth/2 - len(titleText)*3
	titleY := panelY + 20
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY+13,
		color.RGBA{255, 255, 100, 255})

	// Level and Gold info
	if hasExp {
		exp := expComp.(*ExperienceComponent)
		levelText := fmt.Sprintf("Level %d", exp.Level)
		text.Draw(screen, levelText, basicfont.Face7x13, panelX+20, titleY+13,
			color.RGBA{100, 255, 100, 255})
	}

	if hasInv {
		inv := invComp.(*InventoryComponent)
		goldText := fmt.Sprintf("Gold: %d", inv.Gold)
		text.Draw(screen, goldText, basicfont.Face7x13, panelX+panelWidth-120, titleY+13,
			color.RGBA{255, 215, 0, 255})
	}

	// Draw three panels
	ui.drawStatsPanel(screen, stats, equipment)
	ui.drawEquipmentPanel(screen, equipment)
	ui.drawAttributesPanel(screen, stats)

	// Draw controls hint at bottom
	controlsText := "[ESC] or [C] to Close"
	controlsX := panelX + panelWidth/2 - len(controlsText)*3
	controlsY := panelY + panelHeight - 20
	text.Draw(screen, controlsText, basicfont.Face7x13, controlsX, controlsY,
		color.RGBA{180, 180, 180, 255})
}

// calculateLayout computes panel positions based on screen size.
// Called by: Draw() on first render or screen resize
func (ui *CharacterUI) calculateLayout() {
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

	// Layout: Left 30%, Center 40%, Right 30%
	leftWidth := int(float64(panelWidth) * 0.30)
	centerWidth := int(float64(panelWidth) * 0.40)
	rightWidth := panelWidth - leftWidth - centerWidth

	contentY := panelY + 60 // Below title
	contentHeight := panelHeight - 100

	ui.statsPanel = Rectangle{
		X:      panelX + 10,
		Y:      contentY,
		Width:  leftWidth - 20,
		Height: contentHeight,
	}

	ui.equipmentPanel = Rectangle{
		X:      panelX + leftWidth + 10,
		Y:      contentY,
		Width:  centerWidth - 20,
		Height: contentHeight,
	}

	ui.attributesPanel = Rectangle{
		X:      panelX + leftWidth + centerWidth + 10,
		Y:      contentY,
		Width:  rightWidth - 20,
		Height: contentHeight,
	}
}

// drawStatsPanel renders base stats and modifiers.
// Parameters:
//
//	screen - Target image
//	statsComp - StatsComponent with current values
//	equipComp - EquipmentComponent for bonus calculation
func (ui *CharacterUI) drawStatsPanel(screen *ebiten.Image, statsComp *StatsComponent, equipComp *EquipmentComponent) {
	panel := ui.statsPanel

	// Panel background
	vector.DrawFilledRect(screen, float32(panel.X), float32(panel.Y),
		float32(panel.Width), float32(panel.Height),
		color.RGBA{30, 30, 40, 200}, false)
	vector.StrokeRect(screen, float32(panel.X), float32(panel.Y),
		float32(panel.Width), float32(panel.Height), 1,
		color.RGBA{80, 80, 100, 255}, false)

	// Header
	text.Draw(screen, "Base Stats", basicfont.Face7x13,
		panel.X+10, panel.Y+20, color.RGBA{200, 200, 255, 255})

	y := panel.Y + 40
	lineHeight := 25

	// Get equipment bonuses
	var bonusAttack, bonusDefense float64
	if equipComp != nil {
		bonusAttack = float64(equipComp.CachedStats.Damage)
		bonusDefense = float64(equipComp.CachedStats.Defense)
	}

	// Attack
	baseStat := statsComp.Attack
	bonus := bonusAttack
	total := baseStat + bonus
	ui.drawStatLine(screen, panel.X+10, y, "ATK", baseStat, bonus, total)
	ui.drawStatBar(screen, panel.X+10, y+15, panel.Width-20, total, 200.0,
		color.RGBA{255, 150, 150, 255})
	y += lineHeight + 15

	// Defense
	baseStat = statsComp.Defense
	bonus = bonusDefense
	total = baseStat + bonus
	ui.drawStatLine(screen, panel.X+10, y, "DEF", baseStat, bonus, total)
	ui.drawStatBar(screen, panel.X+10, y+15, panel.Width-20, total, 150.0,
		color.RGBA{150, 150, 255, 255})
	y += lineHeight + 15

	// Magic Power
	baseStat = statsComp.MagicPower
	ui.drawStatLine(screen, panel.X+10, y, "MAG", baseStat, 0, baseStat)
	ui.drawStatBar(screen, panel.X+10, y+15, panel.Width-20, baseStat, 200.0,
		color.RGBA{150, 255, 150, 255})
	y += lineHeight + 15

	// Magic Defense
	baseStat = statsComp.MagicDefense
	ui.drawStatLine(screen, panel.X+10, y, "MDEF", baseStat, 0, baseStat)
	ui.drawStatBar(screen, panel.X+10, y+15, panel.Width-20, baseStat, 150.0,
		color.RGBA{200, 150, 255, 255})
}

// drawStatLine draws a stat with base and bonus values.
func (ui *CharacterUI) drawStatLine(screen *ebiten.Image, x, y int, label string, base, bonus, total float64) {
	// Label
	text.Draw(screen, label+":", basicfont.Face7x13, x, y,
		color.RGBA{200, 200, 200, 255})

	// Total value
	totalText := fmt.Sprintf("%.0f", total)
	text.Draw(screen, totalText, basicfont.Face7x13, x+60, y,
		color.RGBA{255, 255, 255, 255})

	// Bonus (if any)
	if bonus > 0 {
		bonusText := fmt.Sprintf("+%.0f", bonus)
		text.Draw(screen, bonusText, basicfont.Face7x13, x+100, y,
			color.RGBA{100, 255, 100, 255})
	}
}

// drawStatBar draws a visual bar for a stat value.
func (ui *CharacterUI) drawStatBar(screen *ebiten.Image, x, y, width int, value, max float64, col color.Color) {
	barHeight := 8

	// Background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(barHeight),
		color.RGBA{40, 40, 40, 255}, false)

	// Fill
	fillPct := value / max
	if fillPct > 1.0 {
		fillPct = 1.0
	}
	fillWidth := float32(width) * float32(fillPct)
	vector.DrawFilledRect(screen, float32(x), float32(y), fillWidth, float32(barHeight),
		col, false)

	// Border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(barHeight), 1,
		color.RGBA{100, 100, 100, 255}, false)
}

// drawEquipmentPanel renders equipped items with their stats.
// Parameters:
//
//	screen - Target image
//	equipComp - EquipmentComponent with equipped items
func (ui *CharacterUI) drawEquipmentPanel(screen *ebiten.Image, equipComp *EquipmentComponent) {
	panel := ui.equipmentPanel

	// Panel background
	vector.DrawFilledRect(screen, float32(panel.X), float32(panel.Y),
		float32(panel.Width), float32(panel.Height),
		color.RGBA{30, 30, 40, 200}, false)
	vector.StrokeRect(screen, float32(panel.X), float32(panel.Y),
		float32(panel.Width), float32(panel.Height), 1,
		color.RGBA{80, 80, 100, 255}, false)

	// Header
	text.Draw(screen, "Equipment", basicfont.Face7x13,
		panel.X+10, panel.Y+20, color.RGBA{200, 200, 255, 255})

	if equipComp == nil {
		text.Draw(screen, "No equipment", basicfont.Face7x13,
			panel.X+10, panel.Y+50, color.RGBA{150, 150, 150, 255})
		return
	}

	y := panel.Y + 40
	lineHeight := 45

	// Define slots to display
	slots := []struct {
		slot EquipmentSlot
		name string
	}{
		{SlotMainHand, "Main Hand"},
		{SlotOffHand, "Off Hand"},
		{SlotHead, "Head"},
		{SlotChest, "Chest"},
		{SlotLegs, "Legs"},
		{SlotBoots, "Boots"},
		{SlotGloves, "Gloves"},
		{SlotAccessory1, "Accessory 1"},
		{SlotAccessory2, "Accessory 2"},
	}

	for _, slotInfo := range slots {
		if y+lineHeight > panel.Y+panel.Height {
			break // Out of space
		}

		// Slot label
		text.Draw(screen, slotInfo.name+":", basicfont.Face7x13,
			panel.X+10, y, color.RGBA{180, 180, 180, 255})

		// Get equipped item
		item := equipComp.GetEquipped(slotInfo.slot)
		if item != nil {
			// Item icon (colored square)
			iconSize := 20
			iconX := panel.X + 10
			iconY := y + 5
			ui.drawItemIcon(screen, iconX, iconY, iconSize, item)

			// Item name
			itemName := item.Name
			if len(itemName) > 25 {
				itemName = itemName[:22] + "..."
			}
			text.Draw(screen, itemName, basicfont.Face7x13,
				iconX+iconSize+5, y+15, color.RGBA{255, 255, 255, 255})

			// Item stats
			statsText := fmt.Sprintf("DMG:%d DEF:%d", item.Stats.Damage, item.Stats.Defense)
			text.Draw(screen, statsText, basicfont.Face7x13,
				panel.X+10, y+30, color.RGBA{150, 255, 150, 255})
		} else {
			// Empty slot
			text.Draw(screen, "  (Empty)", basicfont.Face7x13,
				panel.X+10, y+15, color.RGBA{100, 100, 100, 255})
		}

		y += lineHeight
	}
}

// drawItemIcon renders a simple colored icon for an item.
func (ui *CharacterUI) drawItemIcon(screen *ebiten.Image, x, y, size int, item interface{}) {
	// Generate color based on item rarity (simplified)
	col := color.RGBA{180, 180, 180, 255} // Default gray

	// Draw icon background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(size), float32(size),
		col, false)

	// Draw border
	vector.StrokeRect(screen, float32(x), float32(y), float32(size), float32(size), 1,
		color.RGBA{255, 255, 255, 255}, false)
}

// drawAttributesPanel renders derived stats and resistances.
// Parameters:
//
//	screen - Target image
//	statsComp - StatsComponent for calculations
func (ui *CharacterUI) drawAttributesPanel(screen *ebiten.Image, statsComp *StatsComponent) {
	panel := ui.attributesPanel

	// Panel background
	vector.DrawFilledRect(screen, float32(panel.X), float32(panel.Y),
		float32(panel.Width), float32(panel.Height),
		color.RGBA{30, 30, 40, 200}, false)
	vector.StrokeRect(screen, float32(panel.X), float32(panel.Y),
		float32(panel.Width), float32(panel.Height), 1,
		color.RGBA{80, 80, 100, 255}, false)

	// Header
	text.Draw(screen, "Attributes", basicfont.Face7x13,
		panel.X+10, panel.Y+20, color.RGBA{200, 200, 255, 255})

	y := panel.Y + 40
	lineHeight := 20

	// Calculate derived stats
	derivedStats := ui.calculateDerivedStats(statsComp)

	// Display derived stats
	ui.drawAttributeLine(screen, panel.X+10, y, "Crit Chance",
		formatStatValue(derivedStats["crit_chance"], true),
		color.RGBA{255, 200, 100, 255})
	y += lineHeight

	ui.drawAttributeLine(screen, panel.X+10, y, "Crit Damage",
		formatStatValue(derivedStats["crit_damage"], false)+"x",
		color.RGBA{255, 150, 150, 255})
	y += lineHeight

	ui.drawAttributeLine(screen, panel.X+10, y, "Evasion",
		formatStatValue(derivedStats["evasion"], true),
		color.RGBA{150, 255, 150, 255})
	y += lineHeight * 2

	// Resistances section
	text.Draw(screen, "Resistances:", basicfont.Face7x13,
		panel.X+10, y, color.RGBA{200, 200, 200, 255})
	y += lineHeight

	// Display resistances
	damageTypes := []struct {
		typ  combat.DamageType
		name string
	}{
		{combat.DamageFire, "Fire"},
		{combat.DamageIce, "Ice"},
		{combat.DamageLightning, "Lightning"},
		{combat.DamagePoison, "Poison"},
	}

	for _, dmgType := range damageTypes {
		resist := statsComp.GetResistance(dmgType.typ)
		ui.drawAttributeLine(screen, panel.X+10, y, dmgType.name,
			formatStatValue(resist*100, true),
			ui.getResistanceColor(resist))
		y += lineHeight
	}
}

// drawAttributeLine draws a single attribute line.
func (ui *CharacterUI) drawAttributeLine(screen *ebiten.Image, x, y int, label, value string, col color.Color) {
	// Label
	text.Draw(screen, label+":", basicfont.Face7x13, x, y,
		color.RGBA{180, 180, 180, 255})

	// Value
	text.Draw(screen, value, basicfont.Face7x13, x+100, y, col)
}

// calculateDerivedStats computes crit chance, evasion, etc. from base stats.
// Parameters:
//
//	stats - Base stats component
//
// Returns: Map of derived stat names to values
func (ui *CharacterUI) calculateDerivedStats(stats *StatsComponent) map[string]float64 {
	derived := make(map[string]float64)

	// Critical chance: base + bonus from attack
	derived["crit_chance"] = stats.CritChance * 100 // Convert to percentage

	// Critical damage multiplier
	derived["crit_damage"] = stats.CritDamage

	// Evasion: base + bonus from speed (if speed stat exists, otherwise just base)
	derived["evasion"] = stats.Evasion * 100 // Convert to percentage

	return derived
}

// formatStatValue formats a stat value for display (e.g., "42" or "42.5%")
// Parameters:
//
//	value - Numeric stat value
//	isPercentage - Whether to format as percentage
//
// Returns: Formatted string
func formatStatValue(value float64, isPercentage bool) string {
	if isPercentage {
		return fmt.Sprintf("%.1f%%", value)
	}
	// Round to nearest integer for non-percentage values
	return fmt.Sprintf("%.0f", math.Round(value))
}

// getResistanceColor returns a color based on resistance value.
func (ui *CharacterUI) getResistanceColor(resist float64) color.Color {
	if resist >= 0.5 {
		return color.RGBA{100, 255, 100, 255} // High resistance (green)
	} else if resist >= 0.25 {
		return color.RGBA{255, 255, 100, 255} // Medium resistance (yellow)
	} else if resist > 0 {
		return color.RGBA{255, 200, 100, 255} // Low resistance (orange)
	}
	return color.RGBA{150, 150, 150, 255} // No resistance (gray)
}
