// Package engine provides tutorial integration for the character creation flow.
// This file implements CharacterCreationTutorial which guides first-time players
// through each character creation step with contextual hints and tooltips.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// basicFontCharWidth is the character width for basicfont.Face7x13 (7 pixels per character).
const basicFontCharWidth = 7

// CharacterCreationTutorialStep represents a guided step in the character creation tutorial.
type CharacterCreationTutorialStep struct {
	ID          string
	Title       string
	Description string
	Hint        string
	Completed   bool
}

// CharacterCreationTutorial manages tutorial overlays during character creation.
// It wraps the existing character creation system to provide guidance for
// first-time players while remaining invisible to returning players.
type CharacterCreationTutorial struct {
	Enabled        bool
	CurrentStepIdx int
	Steps          []CharacterCreationTutorialStep
	ShowUI         bool

	// Tutorial completion tracking
	Completed bool

	// Skip option
	Skipped bool

	// Notification state
	NotificationMsg string
	NotificationTTL float64

	// Screen dimensions for layout
	screenWidth  int
	screenHeight int
}

// NewCharacterCreationTutorial creates a new character creation tutorial.
// The tutorial is enabled by default for first-time players.
func NewCharacterCreationTutorial(screenWidth, screenHeight int) *CharacterCreationTutorial {
	return &CharacterCreationTutorial{
		Enabled:        true,
		ShowUI:         true,
		CurrentStepIdx: 0,
		Steps:          createCharacterCreationTutorialSteps(),
		screenWidth:    screenWidth,
		screenHeight:   screenHeight,
	}
}

// createCharacterCreationTutorialSteps generates the tutorial steps for character creation.
func createCharacterCreationTutorialSteps() []CharacterCreationTutorialStep {
	return []CharacterCreationTutorialStep{
		{
			ID:          "welcome_creation",
			Title:       "Welcome to Character Creation!",
			Description: "Let's create your hero. We'll walk you through choosing a name, class, and portrait.",
			Hint:        "Press ENTER or click Next to begin",
			Completed:   false,
		},
		{
			ID:          "name_input",
			Title:       "Choose Your Name",
			Description: "Type your character's name. This is how other players and NPCs will know you.",
			Hint:        "Type a name and press ENTER, or tap a preset name button",
			Completed:   false,
		},
		{
			ID:          "class_selection",
			Title:       "Choose Your Class",
			Description: "Each class has unique stats and abilities. Warrior excels in melee, Mage in magic, and Rogue in agility.",
			Hint:        "Use arrow keys or click a class, then press ENTER to confirm",
			Completed:   false,
		},
		{
			ID:          "portrait_selection",
			Title:       "Choose a Portrait (Optional)",
			Description: "You can upload a custom portrait image or skip this step. Portraits are purely cosmetic.",
			Hint:        "Press TAB to skip, or browse for a .png file",
			Completed:   false,
		},
		{
			ID:          "confirmation",
			Title:       "Confirm Your Character",
			Description: "Review your choices. You can go back to change anything, or confirm to start your adventure!",
			Hint:        "Press ENTER to begin your adventure, or BACKSPACE to go back",
			Completed:   false,
		},
	}
}

// Update processes the character creation tutorial each frame.
// It should be called alongside the character creation system's Update.
// The creationStep parameter indicates the current step in character creation.
func (cct *CharacterCreationTutorial) Update(currentCreationStep int, deltaTime float64) {
	if !cct.Enabled || cct.Completed {
		return
	}

	// Update notification TTL
	if cct.NotificationTTL > 0 {
		cct.NotificationTTL -= deltaTime
		if cct.NotificationTTL <= 0 {
			cct.NotificationMsg = ""
		}
	}

	// Handle skip input (F1 key)
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		cct.SkipTutorial()
		return
	}

	// Synchronize tutorial step with character creation step.
	// Map creation steps to tutorial steps:
	// stepNameInput (0) -> tutorial step 1 (name_input)
	// stepClassSelection (1) -> tutorial step 2 (class_selection)
	// stepPortraitSelection (2) -> tutorial step 3 (portrait_selection)
	// stepConfirmation (3) -> tutorial step 4 (confirmation)
	targetTutorialStep := currentCreationStep + 1 // +1 because tutorial step 0 is welcome

	// Auto-advance the welcome step when the user starts interacting
	if cct.CurrentStepIdx == 0 && currentCreationStep == 0 {
		// Welcome step - advance on any key press
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			cct.advanceStep()
		}
		return
	}

	// Advance tutorial steps to match creation progress
	if targetTutorialStep > cct.CurrentStepIdx && cct.CurrentStepIdx < len(cct.Steps) {
		// Mark intermediate steps as completed
		for cct.CurrentStepIdx < targetTutorialStep && cct.CurrentStepIdx < len(cct.Steps) {
			cct.Steps[cct.CurrentStepIdx].Completed = true
			cct.CurrentStepIdx++
		}

		if cct.CurrentStepIdx < len(cct.Steps) {
			cct.NotificationMsg = fmt.Sprintf("✓ Step complete! Now: %s", cct.Steps[cct.CurrentStepIdx].Title)
			cct.NotificationTTL = 2.0
		}
	}

	// Check if tutorial is complete (all steps done)
	if cct.CurrentStepIdx >= len(cct.Steps) {
		cct.CompleteTutorial()
	}
}

// advanceStep moves to the next tutorial step.
func (cct *CharacterCreationTutorial) advanceStep() {
	if cct.CurrentStepIdx < len(cct.Steps) {
		cct.Steps[cct.CurrentStepIdx].Completed = true
		cct.CurrentStepIdx++

		if cct.CurrentStepIdx < len(cct.Steps) {
			cct.NotificationMsg = fmt.Sprintf("Next: %s", cct.Steps[cct.CurrentStepIdx].Title)
			cct.NotificationTTL = 2.0
		} else {
			cct.CompleteTutorial()
		}
	}
}

// CompleteTutorial marks the tutorial as finished.
func (cct *CharacterCreationTutorial) CompleteTutorial() {
	cct.Completed = true
	cct.Enabled = false
	cct.NotificationMsg = "Character creation tutorial complete! Enjoy your adventure!"
	cct.NotificationTTL = 3.0
}

// SkipTutorial allows returning players to skip the tutorial entirely.
func (cct *CharacterCreationTutorial) SkipTutorial() {
	cct.Skipped = true
	cct.Completed = true
	cct.Enabled = false
	cct.ShowUI = false
	cct.NotificationMsg = "Tutorial skipped"
	cct.NotificationTTL = 2.0
}

// Draw renders the tutorial overlay on top of the character creation UI.
func (cct *CharacterCreationTutorial) Draw(screen *ebiten.Image) {
	if !cct.Enabled || !cct.ShowUI || cct.Completed {
		// Still draw notification if active
		if cct.NotificationTTL > 0 && cct.NotificationMsg != "" {
			cct.drawNotification(screen)
		}
		return
	}

	step := cct.GetCurrentStep()
	if step == nil {
		return
	}

	cct.drawTutorialOverlay(screen, step)

	if cct.NotificationTTL > 0 && cct.NotificationMsg != "" {
		cct.drawNotification(screen)
	}
}

// drawTutorialOverlay renders the tutorial hint panel.
func (cct *CharacterCreationTutorial) drawTutorialOverlay(screen *ebiten.Image, step *CharacterCreationTutorialStep) {
	screenWidth := screen.Bounds().Dx()

	// Draw a thin tutorial banner at the top of the screen
	bannerHeight := 70
	bannerY := 5

	// Semi-transparent background
	vector.DrawFilledRect(screen,
		float32(10), float32(bannerY),
		float32(screenWidth-20), float32(bannerHeight),
		color.RGBA{20, 40, 80, 220}, false)

	// Border
	vector.StrokeRect(screen,
		float32(10), float32(bannerY),
		float32(screenWidth-20), float32(bannerHeight),
		2, color.RGBA{100, 180, 255, 255}, false)

	// Progress indicator
	progress := cct.GetProgress()
	progressWidth := int(float64(screenWidth-40) * progress)
	vector.DrawFilledRect(screen,
		float32(20), float32(bannerY+5),
		float32(progressWidth), 3,
		color.RGBA{100, 200, 255, 255}, false)

	// Title
	titleText := fmt.Sprintf("[Tutorial] %s (%d/%d)", step.Title, cct.CurrentStepIdx+1, len(cct.Steps))
	text.Draw(screen, titleText, basicfont.Face7x13, 20, bannerY+25,
		color.RGBA{255, 255, 200, 255})

	// Description
	text.Draw(screen, step.Description, basicfont.Face7x13, 20, bannerY+43,
		color.RGBA{200, 200, 200, 255})

	// Hint
	text.Draw(screen, step.Hint, basicfont.Face7x13, 20, bannerY+60,
		color.RGBA{100, 255, 100, 255})

	// Skip hint (right-aligned)
	skipText := "F1: Skip Tutorial"
	skipX := screenWidth - 30 - len(skipText)*basicFontCharWidth
	text.Draw(screen, skipText, basicfont.Face7x13, skipX, bannerY+25,
		color.RGBA{180, 180, 180, 200})
}

// drawNotification renders a temporary notification message.
func (cct *CharacterCreationTutorial) drawNotification(screen *ebiten.Image) {
	if cct.NotificationMsg == "" {
		return
	}

	screenWidth := screen.Bounds().Dx()

	notifWidth := 500
	notifHeight := 40
	notifX := (screenWidth - notifWidth) / 2
	notifY := 80

	// Fade effect
	alpha := uint8(255)
	if cct.NotificationTTL < 0.5 {
		alpha = uint8(cct.NotificationTTL * 510)
	}

	vector.DrawFilledRect(screen,
		float32(notifX), float32(notifY),
		float32(notifWidth), float32(notifHeight),
		color.RGBA{50, 120, 50, alpha}, false)

	vector.StrokeRect(screen,
		float32(notifX), float32(notifY),
		float32(notifWidth), float32(notifHeight),
		1, color.RGBA{100, 200, 100, alpha}, false)

	textColor := color.RGBA{255, 255, 255, alpha}
	textX := notifX + (notifWidth-len(cct.NotificationMsg)*basicFontCharWidth)/2
	text.Draw(screen, cct.NotificationMsg, basicfont.Face7x13, textX, notifY+25, textColor)
}

// GetCurrentStep returns the current tutorial step, or nil if complete.
func (cct *CharacterCreationTutorial) GetCurrentStep() *CharacterCreationTutorialStep {
	if !cct.Enabled || cct.CurrentStepIdx >= len(cct.Steps) {
		return nil
	}
	return &cct.Steps[cct.CurrentStepIdx]
}

// GetProgress returns the tutorial progress as a value from 0.0 to 1.0.
func (cct *CharacterCreationTutorial) GetProgress() float64 {
	if len(cct.Steps) == 0 {
		return 1.0
	}
	return float64(cct.CurrentStepIdx) / float64(len(cct.Steps))
}

// IsComplete returns whether the tutorial has been completed or skipped.
func (cct *CharacterCreationTutorial) IsComplete() bool {
	return cct.Completed
}

// IsActive returns true if the tutorial is currently showing guidance.
func (cct *CharacterCreationTutorial) IsActive() bool {
	return cct.Enabled && cct.ShowUI && !cct.Completed
}

// Reset resets the tutorial to its initial state.
func (cct *CharacterCreationTutorial) Reset() {
	cct.Enabled = true
	cct.ShowUI = true
	cct.CurrentStepIdx = 0
	cct.Completed = false
	cct.Skipped = false
	cct.NotificationMsg = ""
	cct.NotificationTTL = 0
	for i := range cct.Steps {
		cct.Steps[i].Completed = false
	}
}

// ExportState exports the tutorial state for persistence.
// Returns completion status, skip status, and per-step completion map.
func (cct *CharacterCreationTutorial) ExportState() (completed, skipped bool, completedSteps map[string]bool) {
	completedSteps = make(map[string]bool)
	for _, step := range cct.Steps {
		if step.Completed {
			completedSteps[step.ID] = true
		}
	}
	return cct.Completed, cct.Skipped, completedSteps
}

// ImportState restores the tutorial state from saved data.
// If the tutorial was previously completed or skipped, it remains disabled.
func (cct *CharacterCreationTutorial) ImportState(completed, skipped bool, completedSteps map[string]bool) {
	cct.Completed = completed
	cct.Skipped = skipped

	if completed || skipped {
		cct.Enabled = false
		cct.ShowUI = false
	}

	for i := range cct.Steps {
		if completedSteps != nil {
			if done, ok := completedSteps[cct.Steps[i].ID]; ok {
				cct.Steps[i].Completed = done
			}
		}
	}

	// Count completed steps to set index
	idx := 0
	for _, step := range cct.Steps {
		if step.Completed {
			idx++
		} else {
			break
		}
	}
	cct.CurrentStepIdx = idx
}

// GetStepByID returns the tutorial step with the given ID, or nil if not found.
func (cct *CharacterCreationTutorial) GetStepByID(stepID string) *CharacterCreationTutorialStep {
	for i := range cct.Steps {
		if cct.Steps[i].ID == stepID {
			return &cct.Steps[i]
		}
	}
	return nil
}

// IsStepCompleted returns true if the step with the given ID has been completed.
func (cct *CharacterCreationTutorial) IsStepCompleted(stepID string) bool {
	for _, step := range cct.Steps {
		if step.ID == stepID {
			return step.Completed
		}
	}
	return false
}

// TutorialCompletionComponent tracks whether the character creation tutorial
// has been completed for a player entity. This component is persisted with
// the save system to ensure returning players skip the tutorial.
type TutorialCompletionComponent struct {
	CreationTutorialDone bool // Whether character creation tutorial was completed
	CreationTutorialSkip bool // Whether character creation tutorial was skipped
}

// Type returns the component type identifier for ECS registration.
func (tc *TutorialCompletionComponent) Type() string { return "tutorial_completion" }

// ShouldShowCreationTutorial checks whether the character creation tutorial
// should be displayed. Returns true for first-time players (no completion
// component) and false for returning players.
func ShouldShowCreationTutorial(player *Entity) bool {
	if player == nil {
		return true // No player entity yet means first time
	}
	comp, ok := player.GetComponent("tutorial_completion")
	if !ok {
		return true // No completion component means first time
	}
	tc, ok := comp.(*TutorialCompletionComponent)
	if !ok {
		return true // Invalid component means treat as first time
	}
	return !tc.CreationTutorialDone && !tc.CreationTutorialSkip
}

// MarkCreationTutorialComplete adds or updates the tutorial completion component
// on a player entity to indicate the character creation tutorial is done.
func MarkCreationTutorialComplete(player *Entity, skipped bool) {
	if player == nil {
		return
	}

	comp, ok := player.GetComponent("tutorial_completion")
	if ok {
		if tc, ok := comp.(*TutorialCompletionComponent); ok {
			tc.CreationTutorialDone = true
			tc.CreationTutorialSkip = skipped
			return
		}
	}

	// Add new component
	player.AddComponent(&TutorialCompletionComponent{
		CreationTutorialDone: true,
		CreationTutorialSkip: skipped,
	})
}
