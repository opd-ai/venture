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

	// Welcome step display timer — prevents auto-advancing the welcome
	// overlay before the player has had time to read it.
	welcomeTimer float64

	// InputConsumed is set when the tutorial consumes an input event
	// (e.g., ENTER to dismiss welcome). Callers can check this to avoid
	// double-processing the same keypress.
	InputConsumed bool
}

// NewCharacterCreationTutorial creates a new character creation tutorial.
// The tutorial is enabled by default for first-time players.
func NewCharacterCreationTutorial() *CharacterCreationTutorial {
	return &CharacterCreationTutorial{
		Enabled:        true,
		ShowUI:         true,
		CurrentStepIdx: 0,
		Steps:          createCharacterCreationTutorialSteps(),
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
			ID:          "subclass_selection",
			Title:       "Choose Your Subclass",
			Description: "Select a hybrid subclass to specialize your base class, or choose None to keep the base class.",
			Hint:        "Use arrow keys or number keys to select, then press ENTER to confirm",
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

	// Reset per-frame flag so callers know whether tutorial consumed input.
	cct.InputConsumed = false

	cct.updateNotification(deltaTime)

	if cct.handleSkipInput() {
		return
	}

	// Handle the welcome overlay (tutorial step 0).
	// The welcome step has no corresponding creationStep; it must be
	// dismissed before the underlying character creation processes input.
	if cct.CurrentStepIdx == 0 {
		cct.welcomeTimer += deltaTime

		// Allow manual dismissal via ENTER, SPACE, or touch/click.
		manualDismiss := false
		if cct.welcomeTimer >= 0.5 {
			// After a short grace period, accept user input to dismiss.
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
				inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
				IsTouchOrMouseJustPressed() {
				manualDismiss = true
				cct.InputConsumed = true
			}
		}

		// Auto-advance after 3 seconds, or on manual dismissal.
		if cct.welcomeTimer >= 3.0 || manualDismiss {
			cct.advanceStep()
		}

		// While on the welcome step, consume ALL key/touch input so the
		// underlying character creation doesn't process it.
		cct.InputConsumed = true
		return
	}

	cct.synchronizeTutorialProgress(currentCreationStep)
}

// updateNotification decrements notification timer and clears message when expired.
func (cct *CharacterCreationTutorial) updateNotification(deltaTime float64) {
	if cct.NotificationTTL > 0 {
		cct.NotificationTTL -= deltaTime
		if cct.NotificationTTL <= 0 {
			cct.NotificationMsg = ""
		}
	}
}

// handleSkipInput processes F3 key press to skip tutorial. Returns true if skipped.
func (cct *CharacterCreationTutorial) handleSkipInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		cct.SkipTutorial()
		return true
	}
	return false
}

// handleWelcomeStep is no longer used — welcome step advancement is now
// handled inline in Update() via a timer, avoiding input conflicts with
// character creation (ENTER/SPACE/typed characters no longer consumed).
// Retained as a no-op for reference.
func (cct *CharacterCreationTutorial) handleWelcomeStep(currentCreationStep int) bool {
	return false
}

// synchronizeTutorialProgress advances tutorial steps to match character creation progress.
// Advances at most one step per frame to prevent skipping intermediate tutorial hints.
func (cct *CharacterCreationTutorial) synchronizeTutorialProgress(currentCreationStep int) {
	targetTutorialStep := currentCreationStep + 1

	// Advance at most ONE step per call so the player has at least one
	// frame to see each tutorial hint before it is marked complete.
	if targetTutorialStep > cct.CurrentStepIdx && cct.CurrentStepIdx < len(cct.Steps) {
		cct.Steps[cct.CurrentStepIdx].Completed = true
		cct.CurrentStepIdx++

		if cct.CurrentStepIdx < len(cct.Steps) {
			cct.NotificationMsg = fmt.Sprintf("✓ Step complete! Now: %s", cct.Steps[cct.CurrentStepIdx].Title)
			cct.NotificationTTL = 2.0
		}
	}

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
// Any remaining steps are marked as completed.
func (cct *CharacterCreationTutorial) CompleteTutorial() {
	for i := range cct.Steps {
		cct.Steps[i].Completed = true
	}
	cct.CurrentStepIdx = len(cct.Steps)
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

// SetEnabled enables or disables the character creation tutorial.
// This is called when the ShowTutorials setting is changed (Task 3.3 from PLAN.md).
func (cct *CharacterCreationTutorial) SetEnabled(enabled bool) {
	cct.Enabled = enabled
	cct.ShowUI = enabled
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
	skipText := "F3: Skip Tutorial"
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
	cct.welcomeTimer = 0
	cct.InputConsumed = false
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
// Phase 3.2: Extended to track onboarding state for seamless tutorial progression.
type TutorialCompletionComponent struct {
	CreationTutorialDone bool `json:"creation_tutorial_done"` // Whether character creation tutorial was completed
	CreationTutorialSkip bool `json:"creation_tutorial_skip"` // Whether character creation tutorial was skipped
	OnboardingState      int  `json:"onboarding_state"`       // Current onboarding phase (Phase 3.2)
	PlayerClass          int  `json:"player_class"`           // Selected class for class-aware tutorials (Phase 3.2)
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

// UpdateOnboardingState updates the onboarding state stored on the player entity.
// Phase 3.2: Enables save/load to preserve onboarding progress.
func UpdateOnboardingState(player *Entity, onboardingState OnboardingState, playerClass CharacterClass) {
	if player == nil {
		return
	}

	comp, ok := player.GetComponent("tutorial_completion")
	if ok {
		if tc, ok := comp.(*TutorialCompletionComponent); ok {
			tc.OnboardingState = int(onboardingState)
			tc.PlayerClass = int(playerClass)
			return
		}
	}

	// Add new component if it doesn't exist
	player.AddComponent(&TutorialCompletionComponent{
		OnboardingState: int(onboardingState),
		PlayerClass:     int(playerClass),
	})
}

// GetOnboardingStateFromPlayer retrieves the stored onboarding state from a player entity.
// Returns StateCharacterCreation and ClassWarrior if no state is stored.
func GetOnboardingStateFromPlayer(player *Entity) (OnboardingState, CharacterClass) {
	if player == nil {
		return StateCharacterCreation, ClassWarrior
	}

	comp, ok := player.GetComponent("tutorial_completion")
	if !ok {
		return StateCharacterCreation, ClassWarrior
	}

	tc, ok := comp.(*TutorialCompletionComponent)
	if !ok {
		return StateCharacterCreation, ClassWarrior
	}

	return OnboardingState(tc.OnboardingState), CharacterClass(tc.PlayerClass)
}
