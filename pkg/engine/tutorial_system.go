// Package engine provides tutorial and guidance for new players.
// This file implements EbitenTutorialSystem which displays step-by-step tutorials
// and hints to help players learn the game mechanics.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// TutorialStep represents a single step in the tutorial sequence
type TutorialStep struct {
	ID          string
	Title       string
	Description string
	Objective   string
	Completed   bool
	Condition   func(*World) bool // Function that returns true when step is complete
}

// EbitenTutorialSystem manages the in-game tutorial progression (Ebiten implementation).
// Implements UISystem interface.
type EbitenTutorialSystem struct {
	Enabled         bool
	CurrentStepIdx  int
	Steps           []TutorialStep
	ShowUI          bool
	NotificationMsg string
	NotificationTTL float64 // Time-to-live for notification (seconds)
}

// NewTutorialSystem creates a new tutorial system with default steps.
func NewTutorialSystem() *EbitenTutorialSystem {
	return &EbitenTutorialSystem{
		Enabled:        true,
		ShowUI:         true,
		Steps:          createDefaultTutorialSteps(),
		CurrentStepIdx: 0,
	}
}

// createDefaultTutorialSteps generates the default tutorial sequence
func createDefaultTutorialSteps() []TutorialStep {
	return []TutorialStep{
		{
			ID:          "welcome",
			Title:       "Welcome to Venture!",
			Description: "Welcome to the world of procedural adventure. Every dungeon, enemy, and item is unique!",
			Objective:   "Press any key to continue", // GAP-005 REPAIR: Changed from "Press SPACE"
			Completed:   false,
			Condition: func(world *World) bool {
				// GAP-001/GAP-005 REPAIR: Check for any key press using frame-persistent flag
				for _, entity := range world.GetEntities() {
					if entity.HasComponent("input") {
						comp, ok := entity.GetComponent("input")
						if !ok {
							continue
						}
						// Use InputProvider interface instead of concrete type
						input, ok := comp.(InputProvider)
						if !ok {
							continue
						}
						// Check for any key press using interface method
						return input.IsAnyKeyPressed()
					}
				}
				return false
			},
		},
		{
			ID:          "movement",
			Title:       "Movement",
			Description: "Use WASD keys to move your character around the dungeon.",
			Objective:   "Move at least 50 units in any direction",
			Completed:   false,
			Condition: func(world *World) bool {
				// Check if player has moved sufficiently
				for _, entity := range world.GetEntities() {
					if entity.HasComponent("input") && entity.HasComponent("position") {
						comp, ok := entity.GetComponent("position")
						if !ok {
							continue
						}
						pos := comp.(*PositionComponent)
						// Simple distance check from origin (400, 300 typical spawn)
						distFromStart := (pos.X-400)*(pos.X-400) + (pos.Y-300)*(pos.Y-300)
						return distFromStart > 2500 // ~50 units
					}
				}
				return false
			},
		},
		{
			ID:          "combat",
			Title:       "Combat Basics",
			Description: "Press SPACE near an enemy to attack. Enemies appear as red sprites.",
			Objective:   "Defeat your first enemy",
			Completed:   false,
			Condition: func(world *World) bool {
				// Check if player has the "attack" component and has used it
				for _, entity := range world.GetEntities() {
					if entity.HasComponent("input") && entity.HasComponent("attack") {
						comp, ok := entity.GetComponent("attack")
						if !ok {
							continue
						}
						attack := comp.(*AttackComponent)
						// Check if attack cooldown is active (means they attacked)
						return attack.CooldownTimer > 0 || attack.CooldownTimer < attack.Cooldown
					}
				}
				return false
			},
		},
		{
			ID:          "health",
			Title:       "Health Management",
			Description: "Watch your health bar in the top-left corner. Don't let it reach zero!",
			Objective:   "Survive combat and maintain health above 50%",
			Completed:   false,
			Condition: func(world *World) bool {
				// Check player health after taking damage
				for _, entity := range world.GetEntities() {
					if entity.HasComponent("input") && entity.HasComponent("health") {
						comp, ok := entity.GetComponent("health")
						if !ok {
							continue
						}
						health := comp.(*HealthComponent)
						// Complete if health is damaged but still above 50%
						return health.Current < health.Max && health.Current > health.Max/2
					}
				}
				return false
			},
		},
		{
			ID:          "inventory",
			Title:       "Inventory System",
			Description: "Press I to open your inventory. Collect items dropped by enemies.",
			Objective:   "Pick up an item and open inventory",
			Completed:   false,
			Condition: func(world *World) bool {
				// Check if player has items in inventory
				for _, entity := range world.GetEntities() {
					if entity.HasComponent("input") && entity.HasComponent("inventory") {
						comp, ok := entity.GetComponent("inventory")
						if !ok {
							continue
						}
						inv := comp.(*InventoryComponent)
						return len(inv.Items) > 0
					}
				}
				return false
			},
		},
		{
			ID:          "skills",
			Title:       "Character Progression",
			Description: "Defeat enemies to gain XP. Level up to become stronger and unlock new abilities!",
			Objective:   "Reach level 2",
			Completed:   false,
			Condition: func(world *World) bool {
				// Check if player has leveled up
				for _, entity := range world.GetEntities() {
					if entity.HasComponent("input") && entity.HasComponent("experience") {
						comp, ok := entity.GetComponent("experience")
						if !ok {
							continue
						}
						exp := comp.(*ExperienceComponent)
						return exp.Level >= 2
					}
				}
				return false
			},
		},
		{
			ID:          "exploration",
			Title:       "Dungeon Exploration",
			Description: "Explore the dungeon to find treasure, secrets, and the stairs to deeper levels.",
			Objective:   "Continue your adventure! Tutorial complete.",
			Completed:   false,
			Condition: func(world *World) bool {
				// Tutorial complete after player has basic understanding
				return true // Always marked complete once reached
			},
		},
	}
}

// Update processes the tutorial system each frame
func (ts *EbitenTutorialSystem) Update(entities []*Entity, deltaTime float64) {
	if !ts.Enabled || ts.CurrentStepIdx >= len(ts.Steps) {
		return
	}

	// Create temporary world for condition checking
	world := &World{entities: make(map[uint64]*Entity), entityListDirty: true}
	for _, entity := range entities {
		world.entities[entity.ID] = entity
	}

	// Update notification TTL
	if ts.NotificationTTL > 0 {
		ts.NotificationTTL -= deltaTime
		if ts.NotificationTTL <= 0 {
			ts.NotificationMsg = ""
		}
	}

	// Check current step completion
	currentStep := &ts.Steps[ts.CurrentStepIdx]
	if !currentStep.Completed && currentStep.Condition(world) {
		currentStep.Completed = true
		ts.CurrentStepIdx++

		// Show notification for completing step
		if ts.CurrentStepIdx < len(ts.Steps) {
			nextStep := &ts.Steps[ts.CurrentStepIdx]
			ts.NotificationMsg = fmt.Sprintf("✓ %s Complete! Next: %s", currentStep.Title, nextStep.Title)
			ts.NotificationTTL = 3.0 // Show for 3 seconds
		} else {
			ts.NotificationMsg = "Tutorial Complete! You're ready to adventure!"
			ts.NotificationTTL = 5.0
			ts.Enabled = false // Disable tutorial after completion
		}
	}
}

// GetCurrentStep returns the current tutorial step, or nil if complete
func (ts *EbitenTutorialSystem) GetCurrentStep() *TutorialStep {
	if !ts.Enabled || ts.CurrentStepIdx >= len(ts.Steps) {
		return nil
	}
	return &ts.Steps[ts.CurrentStepIdx]
}

// GetProgress returns the tutorial progress (0.0 to 1.0)
func (ts *EbitenTutorialSystem) GetProgress() float64 {
	if len(ts.Steps) == 0 {
		return 1.0
	}
	return float64(ts.CurrentStepIdx) / float64(len(ts.Steps))
}

// Skip skips the current tutorial step
func (ts *EbitenTutorialSystem) Skip() {
	if ts.Enabled && ts.CurrentStepIdx < len(ts.Steps) {
		ts.Steps[ts.CurrentStepIdx].Completed = true
		ts.CurrentStepIdx++
		if ts.CurrentStepIdx >= len(ts.Steps) {
			ts.Enabled = false
		}
	}
}

// SkipAll disables the tutorial entirely
func (ts *EbitenTutorialSystem) SkipAll() {
	ts.Enabled = false
	ts.ShowUI = false
}

// Reset resets the tutorial to the beginning
func (ts *EbitenTutorialSystem) Reset() {
	ts.Enabled = true
	ts.ShowUI = true
	ts.CurrentStepIdx = 0
	ts.NotificationMsg = ""
	ts.NotificationTTL = 0
	for i := range ts.Steps {
		ts.Steps[i].Completed = false
	}
}

// GAP-006 REPAIR: Public API for querying tutorial state

// IsStepCompleted returns true if the step with given ID has been completed
func (ts *EbitenTutorialSystem) IsStepCompleted(stepID string) bool {
	for _, step := range ts.Steps {
		if step.ID == stepID {
			return step.Completed
		}
	}
	return false
}

// GetStepByID returns the tutorial step with the given ID, or nil if not found
func (ts *EbitenTutorialSystem) GetStepByID(stepID string) *TutorialStep {
	for i := range ts.Steps {
		if ts.Steps[i].ID == stepID {
			return &ts.Steps[i]
		}
	}
	return nil
}

// IsActive returns true if the tutorial system is currently enabled and showing UI
func (ts *EbitenTutorialSystem) IsActive() bool {
	return ts.Enabled && ts.ShowUI
}

// GetCurrentStepID returns the ID of the current step, or empty string if complete
func (ts *EbitenTutorialSystem) GetCurrentStepID() string {
	step := ts.GetCurrentStep()
	if step == nil {
		return ""
	}
	return step.ID
}

// GetAllSteps returns all tutorial steps (read-only access for UI integration)
func (ts *EbitenTutorialSystem) GetAllSteps() []TutorialStep {
	// Return copy to prevent external modification
	steps := make([]TutorialStep, len(ts.Steps))
	copy(steps, ts.Steps)
	return steps
}

// GAP-003 REPAIR: Tutorial state serialization for save/load

// ExportState exports the current tutorial state for saving
// Returns map of step IDs to completion status, current index, and enabled flags
func (ts *EbitenTutorialSystem) ExportState() (enabled, showUI bool, currentStepIdx int, completedSteps map[string]bool) {
	completedSteps = make(map[string]bool)
	for _, step := range ts.Steps {
		if step.Completed {
			completedSteps[step.ID] = true
		}
	}
	return ts.Enabled, ts.ShowUI, ts.CurrentStepIdx, completedSteps
}

// ImportState restores tutorial state from saved data
// Applies saved completion status to matching step IDs
func (ts *EbitenTutorialSystem) ImportState(enabled, showUI bool, currentStepIdx int, completedSteps map[string]bool) {
	ts.Enabled = enabled
	ts.ShowUI = showUI
	ts.CurrentStepIdx = currentStepIdx

	// Apply completion status from save data
	for i := range ts.Steps {
		stepID := ts.Steps[i].ID
		if completed, ok := completedSteps[stepID]; ok {
			ts.Steps[i].Completed = completed
		}
	}

	// Validate currentStepIdx (in case tutorial steps changed between save/load)
	// Clamp negative values to 0
	if ts.CurrentStepIdx < 0 {
		ts.CurrentStepIdx = 0
	}
	// Clamp values beyond step count to last step
	if ts.CurrentStepIdx >= len(ts.Steps) {
		ts.CurrentStepIdx = len(ts.Steps) - 1
		if ts.CurrentStepIdx < 0 {
			ts.CurrentStepIdx = 0
		}
	}
}

// ShowNotification displays a notification message for the specified duration.
// This can be used to show feedback for game actions like saving/loading.
func (ts *EbitenTutorialSystem) ShowNotification(msg string, duration float64) {
	ts.NotificationMsg = msg
	ts.NotificationTTL = duration
}

// Draw renders the tutorial UI overlay (implements UISystem interface).
// The screen parameter should be *ebiten.Image in production.
func (ts *EbitenTutorialSystem) Draw(screen interface{}) {
	// Type assert to *ebiten.Image
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok {
		return // Invalid screen type
	}

	if !ts.Enabled || !ts.ShowUI {
		return
	}

	step := ts.GetCurrentStep()
	if step == nil {
		// Show notification if available
		if ts.NotificationTTL > 0 {
			ts.drawNotification(ebitenScreen)
		}
		return
	}

	// Draw tutorial panel with responsive positioning to avoid HUD overlap
	screenWidth := ebitenScreen.Bounds().Dx()
	screenHeight := ebitenScreen.Bounds().Dy()

	// Scale panel size based on screen dimensions
	panelWidth := 400
	if screenWidth < 800 {
		// Minimum 300px, or leave 100px margin
		if screenWidth-100 > 300 {
			panelWidth = screenWidth - 100
		} else {
			panelWidth = 300
		}
	}
	panelHeight := 150

	// HUD element margins (approximate positions)
	const hudMarginTop = 120   // Health bar + stats panel height
	const hudMarginBottom = 60 // XP bar height
	const hudMarginRight = 220 // Stats panel width + margin

	// Position panel to avoid HUD elements
	var panelX, panelY int
	if screenWidth >= 800 && screenHeight >= 600 {
		// Standard position: bottom-right, above XP bar
		panelX = screenWidth - panelWidth - 20
		panelY = screenHeight - panelHeight - hudMarginBottom
	} else if screenHeight >= 400 {
		// Small screen: center-bottom, above XP bar
		panelX = (screenWidth - panelWidth) / 2
		panelY = screenHeight - panelHeight - hudMarginBottom
	} else {
		// Tiny screen: center-center overlay (last resort)
		panelX = (screenWidth - panelWidth) / 2
		panelY = (screenHeight - panelHeight) / 2
	}

	// Semi-transparent background
	vector.DrawFilledRect(ebitenScreen,
		float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{0, 0, 0, 200}, false)

	// Border
	vector.StrokeRect(ebitenScreen,
		float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		2, color.RGBA{100, 200, 255, 255}, false)

	// Progress bar
	progressWidth := int(float64(panelWidth-20) * ts.GetProgress())
	vector.DrawFilledRect(ebitenScreen,
		float32(panelX+10), float32(panelY+10),
		float32(progressWidth), 4,
		color.RGBA{100, 200, 255, 255}, false)

	// Title
	titleColor := color.RGBA{255, 255, 100, 255}
	text.Draw(ebitenScreen, fmt.Sprintf("Tutorial (%d/%d)", ts.CurrentStepIdx+1, len(ts.Steps)),
		basicfont.Face7x13, panelX+10, panelY+35, titleColor)

	text.Draw(ebitenScreen, step.Title, basicfont.Face7x13, panelX+10, panelY+55, color.White)

	// Description
	descColor := color.RGBA{200, 200, 200, 255}
	ts.drawWrappedText(ebitenScreen, step.Description, panelX+10, panelY+75, panelWidth-20, descColor)

	// Objective
	objColor := color.RGBA{100, 255, 100, 255}
	text.Draw(ebitenScreen, "Objective: "+step.Objective, basicfont.Face7x13, panelX+10, panelY+120, objColor)

	// Skip hint
	hintColor := color.RGBA{150, 150, 150, 255}
	text.Draw(ebitenScreen, "Press ESC to skip current step", basicfont.Face7x13, panelX+10, panelY+140, hintColor)

	// Show notification if available
	if ts.NotificationTTL > 0 {
		ts.drawNotification(ebitenScreen)
	}
}

// drawNotification renders a temporary notification message
func (ts *EbitenTutorialSystem) drawNotification(screen *ebiten.Image) {
	if ts.NotificationMsg == "" {
		return
	}

	screenWidth := screen.Bounds().Dx()

	notifWidth := 500
	notifHeight := 50
	notifX := (screenWidth - notifWidth) / 2
	notifY := 100

	// Fade effect based on TTL
	alpha := uint8(255)
	if ts.NotificationTTL < 0.5 {
		alpha = uint8(ts.NotificationTTL * 510) // Fade out in last 0.5s
	}

	// Background
	vector.DrawFilledRect(screen,
		float32(notifX), float32(notifY),
		float32(notifWidth), float32(notifHeight),
		color.RGBA{50, 150, 50, alpha}, false)

	// Border
	vector.StrokeRect(screen,
		float32(notifX), float32(notifY),
		float32(notifWidth), float32(notifHeight),
		2, color.RGBA{100, 255, 100, alpha}, false)

	// Text
	textColor := color.RGBA{255, 255, 255, alpha}
	// Center text (approximate)
	textX := notifX + (notifWidth-len(ts.NotificationMsg)*7)/2
	text.Draw(screen, ts.NotificationMsg, basicfont.Face7x13, textX, notifY+30, textColor)
}

// drawWrappedText draws text with word wrapping
func (ts *EbitenTutorialSystem) drawWrappedText(screen *ebiten.Image, str string, x, y, maxWidth int, clr color.Color) {
	charWidth := 7 // basicfont.Face7x13 character width
	maxChars := maxWidth / charWidth

	words := splitWords(str)
	currentLine := ""
	lineY := y

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > maxChars && currentLine != "" {
			// Draw current line and start new one
			text.Draw(screen, currentLine, basicfont.Face7x13, x, lineY, clr)
			currentLine = word
			lineY += 15 // Line height
		} else {
			currentLine = testLine
		}
	}

	// Draw remaining text
	if currentLine != "" {
		text.Draw(screen, currentLine, basicfont.Face7x13, x, lineY, clr)
	}
}

// splitWords splits a string into words
func splitWords(str string) []string {
	var words []string
	currentWord := ""

	for _, ch := range str {
		if ch == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(ch)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// SetActive implements UISystem interface.
func (ts *EbitenTutorialSystem) SetActive(active bool) {
	ts.ShowUI = active
}

// Compile-time interface check
var _ UISystem = (*EbitenTutorialSystem)(nil)
