// Package prestige provides prestige system UI components.
// This file implements the prestige menu for paragon point allocation and progress visualization.
package prestige

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/mobile"
)

// PrestigeMenuOption represents a selectable option in the prestige menu.
type PrestigeMenuOption int

const (
	// PrestigeOptionHealth allocates points to health stat
	PrestigeOptionHealth PrestigeMenuOption = iota
	// PrestigeOptionDamage allocates points to damage stat
	PrestigeOptionDamage
	// PrestigeOptionDefense allocates points to defense stat
	PrestigeOptionDefense
	// PrestigeOptionSpeed allocates points to speed stat
	PrestigeOptionSpeed
	// PrestigeOptionCritical allocates points to critical stat
	PrestigeOptionCritical
	// PrestigeOptionRespec resets all point allocations
	PrestigeOptionRespec
	// PrestigeOptionBack closes the menu
	PrestigeOptionBack
)

// String returns the display text for each menu option.
func (o PrestigeMenuOption) String() string {
	switch o {
	case PrestigeOptionHealth:
		return "Health"
	case PrestigeOptionDamage:
		return "Damage"
	case PrestigeOptionDefense:
		return "Defense"
	case PrestigeOptionSpeed:
		return "Speed"
	case PrestigeOptionCritical:
		return "Critical"
	case PrestigeOptionRespec:
		return "Respec All"
	case PrestigeOptionBack:
		return "Back"
	default:
		return "Unknown"
	}
}

// toParagonStat converts a menu option to its corresponding ParagonStat.
// Returns -1 for non-stat options.
func (o PrestigeMenuOption) toParagonStat() ParagonStat {
	switch o {
	case PrestigeOptionHealth:
		return StatHealth
	case PrestigeOptionDamage:
		return StatDamage
	case PrestigeOptionDefense:
		return StatDefense
	case PrestigeOptionSpeed:
		return StatSpeed
	case PrestigeOptionCritical:
		return StatCritical
	default:
		return ParagonStat(-1)
	}
}

// PrestigeUI renders and handles input for the prestige menu.
// Provides paragon point allocation and prestige progress visualization.
type PrestigeUI struct {
	screenWidth  int
	screenHeight int
	selectedIdx  int
	options      []PrestigeMenuOption

	// Manager for prestige data
	manager *Manager

	// Player being displayed
	playerID  string
	className string

	// Callbacks
	onBack          func()
	onRespecConfirm func(cost int) bool // Returns true if player can afford respec

	// Visibility flag
	visible bool

	// Input provider for testing - uses standard engine.InputProvider interface
	inputProvider engine.InputProvider

	// Touch support
	touchHandler *mobile.TouchInputHandler
	closeButton  *mobile.TouchButton
	allocButtons []*mobile.TouchButton // Allocate point buttons
	respecButton *mobile.TouchButton
	backButton   *mobile.TouchButton
}

// NewPrestigeUI creates a new prestige UI with the provided manager.
func NewPrestigeUI(screenWidth, screenHeight int, manager *Manager) *PrestigeUI {
	p := &PrestigeUI{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		selectedIdx:  0,
		manager:      manager,
		options: []PrestigeMenuOption{
			PrestigeOptionHealth,
			PrestigeOptionDamage,
			PrestigeOptionDefense,
			PrestigeOptionSpeed,
			PrestigeOptionCritical,
			PrestigeOptionRespec,
			PrestigeOptionBack,
		},
		visible:      false,
		touchHandler: mobile.NewTouchInputHandler(),
	}

	// Create close button (top-right)
	p.closeButton = mobile.NewTouchButton(
		float64(screenWidth-64),
		10,
		44, 44,
		"✕",
		func() {
			p.Hide()
			if p.onBack != nil {
				p.onBack()
			}
		},
	)

	// Create allocate buttons for stat options
	p.allocButtons = make([]*mobile.TouchButton, 5) // 5 stats
	for i := 0; i < 5; i++ {
		idx := i
		p.allocButtons[i] = mobile.NewTouchButton(
			0, 0, // Position set dynamically
			44, 44,
			"+",
			func() { p.allocatePoint(PrestigeMenuOption(idx)) },
		)
	}

	// Create respec button
	p.respecButton = mobile.NewTouchButton(
		0, 0,
		100, 44,
		"Respec",
		func() { p.activateOption(PrestigeOptionRespec) },
	)

	// Create back button
	p.backButton = mobile.NewTouchButton(
		0, 0,
		80, 44,
		"Back",
		func() { p.activateOption(PrestigeOptionBack) },
	)

	return p
}

// SetInputProvider sets the input provider for testing.
// Accepts the standard engine.InputProvider interface from pkg/engine/interfaces.go.
func (p *PrestigeUI) SetInputProvider(provider engine.InputProvider) {
	p.inputProvider = provider
}

// SetBackCallback sets the callback function called when "Back" is selected.
func (p *PrestigeUI) SetBackCallback(callback func()) {
	p.onBack = callback
}

// SetRespecCallback sets the callback for respec cost confirmation.
// The callback receives the gold cost and returns true if the player can afford it.
func (p *PrestigeUI) SetRespecCallback(callback func(cost int) bool) {
	p.onRespecConfirm = callback
}

// Show displays the prestige menu for a specific player.
func (p *PrestigeUI) Show(playerID, className string) {
	p.visible = true
	p.playerID = playerID
	p.className = className
	p.selectedIdx = 0
}

// Hide hides the prestige menu.
func (p *PrestigeUI) Hide() {
	p.visible = false
}

// IsVisible returns whether the prestige menu is currently visible.
func (p *PrestigeUI) IsVisible() bool {
	return p.visible
}

// Update processes input for the prestige menu.
// Returns true if a significant action occurred (e.g., back selected).
func (p *PrestigeUI) Update() bool {
	if !p.visible {
		return false
	}

	p.updateButtons()
	p.handleNavigation()
	p.handleActivation()
	return p.handleEscapeKey()
}

// updateButtons updates all interactive buttons.
func (p *PrestigeUI) updateButtons() {
	if p.touchHandler != nil {
		p.touchHandler.Update()
	}
	if p.closeButton != nil {
		p.closeButton.Update()
	}
	for i := range p.allocButtons {
		if p.allocButtons[i] != nil {
			p.allocButtons[i].Update()
		}
	}
	if p.respecButton != nil {
		p.respecButton.Update()
	}
	if p.backButton != nil {
		p.backButton.Update()
	}
}

// handleNavigation processes up/down keyboard navigation.
func (p *PrestigeUI) handleNavigation() {
	if p.inputProvider != nil && p.inputProvider.IsMenuUpJustPressed() {
		p.selectedIdx--
		if p.selectedIdx < 0 {
			p.selectedIdx = len(p.options) - 1
		}
	}
	if p.inputProvider != nil && p.inputProvider.IsMenuDownJustPressed() {
		p.selectedIdx++
		if p.selectedIdx >= len(p.options) {
			p.selectedIdx = 0
		}
	}
}

// handleActivation processes Enter/Space key to activate options.
func (p *PrestigeUI) handleActivation() {
	if p.inputProvider != nil && p.inputProvider.IsMenuConfirmJustPressed() {
		selectedOption := p.options[p.selectedIdx]
		p.activateOption(selectedOption)
	}
}

// handleEscapeKey processes the ESC key for closing the menu.
func (p *PrestigeUI) handleEscapeKey() bool {
	if p.inputProvider != nil && p.inputProvider.IsMenuBackJustPressed() {
		p.Hide()
		if p.onBack != nil {
			p.onBack()
		}
		return true
	}
	return false
}

// allocatePoint allocates a paragon point to the specified stat.
func (p *PrestigeUI) allocatePoint(option PrestigeMenuOption) {
	stat := option.toParagonStat()
	if stat < 0 {
		return
	}

	// Attempt allocation via manager
	_ = p.manager.AllocateParagonPoint(p.playerID, stat)
}

// activateOption processes the selected menu option.
func (p *PrestigeUI) activateOption(option PrestigeMenuOption) {
	switch option {
	case PrestigeOptionHealth, PrestigeOptionDamage, PrestigeOptionDefense,
		PrestigeOptionSpeed, PrestigeOptionCritical:
		p.allocatePoint(option)
	case PrestigeOptionRespec:
		p.handleRespec()
	case PrestigeOptionBack:
		p.Hide()
		if p.onBack != nil {
			p.onBack()
		}
	}
}

// handleRespec handles the respec action.
func (p *PrestigeUI) handleRespec() {
	totalAllocated := p.manager.GetTotalAllocatedPoints(p.playerID)
	if totalAllocated == 0 {
		return
	}

	cost := totalAllocated * RespecCostPerPoint

	// Check if player can afford it
	if p.onRespecConfirm != nil && !p.onRespecConfirm(cost) {
		return
	}

	// Perform respec
	_, _ = p.manager.RespecParagonPoints(p.playerID)
}

// Draw renders the prestige menu to the screen.
func (p *PrestigeUI) Draw(screen *ebiten.Image) {
	if !p.visible || screen == nil {
		return
	}

	// Draw semi-transparent background
	bgColor := color.RGBA{0, 0, 0, 200}
	ebitenutil.DrawRect(screen, 0, 0, float64(p.screenWidth), float64(p.screenHeight), bgColor)

	// Draw title
	titleX := float64(p.screenWidth/2 - 100)
	titleY := 30.0
	ebitenutil.DebugPrintAt(screen, "=== PRESTIGE ===", int(titleX), int(titleY))

	// Draw close button
	if p.closeButton != nil {
		p.closeButton.Draw(screen)
	}

	// Draw prestige info section
	p.drawPrestigeInfo(screen)

	// Draw paragon allocation section
	p.drawParagonSection(screen)

	// Draw unlocked abilities section
	p.drawAbilitiesSection(screen)

	// Draw controls hint
	p.drawControlsHint(screen)
}

// drawPrestigeInfo draws the prestige level and XP progress.
func (p *PrestigeUI) drawPrestigeInfo(screen *ebiten.Image) {
	startY := 60
	leftX := 50

	// Get prestige data
	prestige := p.manager.GetPlayerPrestige(p.playerID)
	if prestige == nil {
		ebitenutil.DebugPrintAt(screen, "No prestige data found", leftX, startY)
		return
	}

	currentXP, requiredXP := p.manager.GetXPProgress(p.playerID)
	visualTier := p.manager.GetVisualTier(prestige.PrestigeLevel)

	// Draw prestige level
	levelStr := fmt.Sprintf("Prestige Level: %d", prestige.PrestigeLevel)
	ebitenutil.DebugPrintAt(screen, levelStr, leftX, startY)

	// Draw XP progress bar
	xpY := startY + 25
	progressWidth := 200.0
	progressHeight := 16.0
	progress := 0.0
	if requiredXP > 0 {
		progress = float64(currentXP) / float64(requiredXP)
		if progress > 1.0 {
			progress = 1.0
		}
	}

	// Background
	ebitenutil.DrawRect(screen, float64(leftX), float64(xpY), progressWidth, progressHeight, color.RGBA{50, 50, 50, 255})
	// Progress fill
	if progress > 0 {
		ebitenutil.DrawRect(screen, float64(leftX), float64(xpY), progressWidth*progress, progressHeight, color.RGBA{100, 180, 100, 255})
	}
	// XP text
	xpStr := fmt.Sprintf("XP: %d / %d", currentXP, requiredXP)
	ebitenutil.DebugPrintAt(screen, xpStr, leftX+int(progressWidth)+10, xpY)

	// Draw visual tier
	tierY := xpY + 25
	tierStr := fmt.Sprintf("Visual Tier: %s", visualTier.String())
	ebitenutil.DebugPrintAt(screen, tierStr, leftX, tierY)

	// Draw available paragon points
	pointsY := tierY + 25
	pointsStr := fmt.Sprintf("Paragon Points: %d", prestige.ParagonPoints)
	ebitenutil.DebugPrintAt(screen, pointsStr, leftX, pointsY)
}

// drawParagonSection draws the paragon point allocation options.
func (p *PrestigeUI) drawParagonSection(screen *ebiten.Image) {
	startY := 180
	leftX := p.screenWidth/2 - 150
	lineHeight := 35

	ebitenutil.DebugPrintAt(screen, "--- Paragon Stats ---", leftX, startY-25)

	prestige := p.manager.GetPlayerPrestige(p.playerID)

	for i, option := range p.options {
		y := startY + i*lineHeight

		// Highlight selected option
		if i == p.selectedIdx {
			ebitenutil.DebugPrintAt(screen, ">", leftX-20, y)
		}

		// Draw option label
		ebitenutil.DebugPrintAt(screen, option.String(), leftX, y)

		// Draw value/status for stat options
		stat := option.toParagonStat()
		if stat >= 0 && prestige != nil {
			allocatedPts := prestige.ParagonAllocations[stat]
			bonus := float64(allocatedPts) * ParagonPointBonus * 100
			valueStr := fmt.Sprintf("%d pts (+%.1f%%)", allocatedPts, bonus)
			ebitenutil.DebugPrintAt(screen, valueStr, leftX+150, y)

			// Draw allocate button
			if i < len(p.allocButtons) && p.allocButtons[i] != nil {
				p.allocButtons[i].SetPosition(float64(leftX+300), float64(y-8))
				p.allocButtons[i].Draw(screen)
			}
		} else if option == PrestigeOptionRespec {
			totalAllocated := 0
			if prestige != nil {
				for _, pts := range prestige.ParagonAllocations {
					totalAllocated += pts
				}
			}
			cost := totalAllocated * RespecCostPerPoint
			costStr := fmt.Sprintf("Cost: %d gold", cost)
			ebitenutil.DebugPrintAt(screen, costStr, leftX+150, y)

			if p.respecButton != nil {
				p.respecButton.SetPosition(float64(leftX+300), float64(y-8))
				p.respecButton.Draw(screen)
			}
		} else if option == PrestigeOptionBack {
			if p.backButton != nil {
				p.backButton.SetPosition(float64(leftX+150), float64(y-8))
				p.backButton.Draw(screen)
			}
		}
	}
}

// drawAbilitiesSection draws unlocked prestige abilities.
func (p *PrestigeUI) drawAbilitiesSection(screen *ebiten.Image) {
	startY := p.screenHeight - 180
	leftX := 50

	ebitenutil.DebugPrintAt(screen, "--- Unlocked Abilities ---", leftX, startY)

	prestige := p.manager.GetPlayerPrestige(p.playerID)
	if prestige == nil || len(prestige.UnlockedAbilities) == 0 {
		ebitenutil.DebugPrintAt(screen, "No abilities unlocked yet", leftX, startY+20)
		ebitenutil.DebugPrintAt(screen, "(Unlock at prestige 10, 25, 50, 100)", leftX, startY+40)
		return
	}

	y := startY + 20
	for _, level := range prestige.UnlockedAbilities {
		ability := p.manager.GetPrestigeAbility(prestige.ClassName, level)
		if ability != nil {
			abilityStr := fmt.Sprintf("[Lv%d] %s", level, ability.Name)
			ebitenutil.DebugPrintAt(screen, abilityStr, leftX, y)
			y += 20
		}
	}
}

// drawControlsHint draws the controls help text.
func (p *PrestigeUI) drawControlsHint(screen *ebiten.Image) {
	hintY := p.screenHeight - 60
	centerX := p.screenWidth/2 - 150
	ebitenutil.DebugPrintAt(screen, "Controls: Up/Down - Navigate | Enter/Space - Allocate/Select | ESC - Back", centerX, hintY)
}
