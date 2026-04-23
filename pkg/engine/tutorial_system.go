// Package engine provides tutorial and guidance for new players.
// This file implements EbitenTutorialSystem which displays step-by-step tutorials
// and hints to help players learn the game mechanics.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/mobile"
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

	// Touch support
	touchHandler *mobile.TouchInputHandler
	nextButton   *mobile.TouchButton
	skipButton   *mobile.TouchButton
	screenWidth  int
	screenHeight int

	// Class-aware tutorial content (Phase 3.4)
	PlayerClass CharacterClass

	// OnCompleteCallback is called when all tutorial steps are completed or the tutorial is skipped.
	// Used by OnboardingManager to transition to the next onboarding phase.
	OnCompleteCallback func()

	// enableGraceFrames counts down after the tutorial is first enabled,
	// preventing the welcome step from auto-completing due to held keys
	// (e.g., ENTER still held from confirming character creation).
	enableGraceFrames int
}

// NewTutorialSystem creates a new tutorial system with default steps.
func NewTutorialSystem() *EbitenTutorialSystem {
	return NewTutorialSystemWithSize(800, 600) // Default screen size
}

// NewTutorialSystemWithSize creates a new tutorial system with specified screen dimensions.
func NewTutorialSystemWithSize(screenWidth, screenHeight int) *EbitenTutorialSystem {
	ts := &EbitenTutorialSystem{
		Enabled:        true,
		ShowUI:         true,
		Steps:          createDefaultTutorialSteps(),
		CurrentStepIdx: 0,
		touchHandler:   mobile.NewTouchInputHandler(),
		screenWidth:    screenWidth,
		screenHeight:   screenHeight,
	}

	// Create next button (bottom-right)
	ts.nextButton = mobile.NewTouchButton(
		float64(screenWidth-164),
		float64(screenHeight-64),
		120, 44,
		"Next",
		func() { ts.advanceToNextStep() },
	)

	// Create skip button (bottom-left)
	ts.skipButton = mobile.NewTouchButton(
		44,
		float64(screenHeight-64),
		120, 44,
		"Skip Tutorial",
		func() { ts.DisableTutorial() },
	)

	return ts
}

// SetPlayerClass sets the player's class for class-aware tutorial content.
// This allows the tutorial to display class-specific hints and descriptions.
func (ts *EbitenTutorialSystem) SetPlayerClass(class CharacterClass) {
	ts.PlayerClass = class
}

// GetPlayerClass returns the currently set player class.
func (ts *EbitenTutorialSystem) GetPlayerClass() CharacterClass {
	return ts.PlayerClass
}

// classStepOverrides defines class-specific description overrides for tutorial steps.
// Key format is "stepID:classConstant". If no override exists, uses default text.
var classStepOverrides = map[string]map[CharacterClass]string{
	"combat": {
		ClassWarrior:      "Press SPACE near an enemy to strike with your sword. Warriors excel at close combat!",
		ClassMage:         "Press 1-5 to cast spells at enemies from a safe distance. Mana regenerates over time.",
		ClassRogue:        "Press SPACE near an enemy to attack. Get behind foes for critical backstab damage!",
		ClassRanger:       "Click on enemies to fire arrows. Keep your distance for maximum effectiveness!",
		ClassCleric:       "Press SPACE to strike enemies. As a Cleric, you can also heal yourself with spells!",
		ClassNecromancer:  "Press 1-5 to cast dark magic. Summon minions to fight alongside you!",
		ClassBattlemage:   "Combine SPACE attacks with spells (1-5). You have both martial and magical prowess!",
		ClassSpellblade:   "Use SPACE for quick strikes, then 1-5 for spell combos. Agility meets arcane!",
		ClassPaladin:      "Press SPACE for holy strikes. Your attacks can smite evil and heal allies!",
		ClassMonk:         "Press SPACE for rapid martial arts strikes. Channel spiritual energy for special moves!",
		ClassDeathKnight:  "Press SPACE for devastating dark strikes. Your attacks drain life from enemies!",
		ClassWitchHunter:  "Click to fire blessed bolts. Your shots deal extra damage to magical foes!",
		ClassBeastlord:    "Press SPACE while your beast companion attacks. Command your ally with number keys!",
		ClassArcaneArcher: "Click to fire arcane arrows that explode on impact. Combine shots with spell effects!",
		ClassShadowPriest: "Use stealth and dark magic together. Strike from shadows, then curse your foes!",
		ClassDruid:        "Shape-shift forms with number keys. Each form offers unique combat abilities!",
		ClassInquisitor:   "Press SPACE for precise strikes. Your holy power reveals hidden enemies!",
		ClassBloodKnight:  "Press SPACE for blood-powered attacks. Damage heals you through blood magic!",
		ClassMystic:       "Press 1-5 to cast combined arcane and divine spells. Balance offense and support!",
		ClassWarlock:      "Press 1-5 to unleash pact magic. Your dark patron grants powerful abilities!",
		ClassNinja:        "Press SPACE for swift strikes. Use stealth (Shift) and thrown weapons for combos!",
	},
	"skills": {
		ClassWarrior:      "Level up to unlock new combat stances and powerful weapon techniques!",
		ClassMage:         "Level up to learn new spells and increase your mana pool!",
		ClassRogue:        "Level up to improve stealth abilities and unlock deadly finishing moves!",
		ClassRanger:       "Level up to gain new arrow types and improve your companion bond!",
		ClassCleric:       "Level up to unlock more powerful healing spells and holy buffs!",
		ClassNecromancer:  "Level up to summon stronger undead and learn devastating curses!",
		ClassBattlemage:   "Level up to unlock spell-sword combos and battle enchantments!",
		ClassSpellblade:   "Level up to chain magic and melee into devastating combo attacks!",
		ClassPaladin:      "Level up to strengthen your auras and unlock divine smite powers!",
		ClassMonk:         "Level up to master new ki techniques and unlock powerful stances!",
		ClassDeathKnight:  "Level up to strengthen your dark powers and raise more powerful minions!",
		ClassWitchHunter:  "Level up to gain blessed ammunition and track magical enemies better!",
		ClassBeastlord:    "Level up to evolve your beast companion and unlock pack tactics!",
		ClassArcaneArcher: "Level up to imbue arrows with new elemental effects!",
		ClassShadowPriest: "Level up to deepen shadow magic and strengthen curse effects!",
		ClassDruid:        "Level up to unlock new animal forms and nature spells!",
		ClassInquisitor:   "Level up to strengthen holy interrogation and gain truth-seeing powers!",
		ClassBloodKnight:  "Level up to increase blood magic potency and life drain effects!",
		ClassMystic:       "Level up to unlock deeper arcane-divine spell fusions!",
		ClassWarlock:      "Level up to strengthen your pact bond and unlock new dark gifts!",
		ClassNinja:        "Level up to learn new ninjutsu techniques and shadow arts!",
	},
	"inventory": {
		ClassWarrior:     "Press I to open your inventory. Warriors benefit most from armor and weapons!",
		ClassMage:        "Press I to open your inventory. Look for staffs, robes, and mana-boosting items!",
		ClassRogue:       "Press I to open your inventory. Daggers and light armor improve your stealth!",
		ClassRanger:      "Press I to open your inventory. Bows, arrows, and pet gear are your priorities!",
		ClassCleric:      "Press I to open your inventory. Maces and holy symbols boost your healing!",
		ClassNecromancer: "Press I to open your inventory. Dark tomes and cursed items empower you!",
	},
}

// getClassAwareDescription returns the class-specific description for a step,
// or the default description if no override exists for the current class.
func (ts *EbitenTutorialSystem) getClassAwareDescription(step *TutorialStep) string {
	if step == nil {
		return ""
	}

	// Check if there are overrides for this step
	stepOverrides, hasStepOverrides := classStepOverrides[step.ID]
	if !hasStepOverrides {
		return step.Description
	}

	// Check if there's an override for this specific class
	if classDesc, hasClassOverride := stepOverrides[ts.PlayerClass]; hasClassOverride {
		return classDesc
	}

	// Fall back to default description
	return step.Description
}

// createDefaultTutorialSteps generates the default tutorial sequence
func createDefaultTutorialSteps() []TutorialStep {
	return []TutorialStep{
		{
			ID:          "welcome",
			Title:       "Welcome to Venture!",
			Description: "Welcome to the world of procedural adventure. Every dungeon, enemy, and item is unique!",
			Objective:   "Press any key",
			Completed:   false,
			Condition:   checkWelcomeCondition,
		},
		{
			ID:          "movement",
			Title:       "Movement",
			Description: "Use WASD keys to move your character around the dungeon.",
			Objective:   "Move at least 50 units in any direction",
			Completed:   false,
			Condition:   checkMovementCondition,
		},
		{
			ID:          "combat",
			Title:       "Combat Basics",
			Description: "Press SPACE near an enemy to attack. Enemies appear as red sprites.",
			Objective:   "Defeat your first enemy",
			Completed:   false,
			Condition:   checkCombatCondition,
		},
		{
			ID:          "health",
			Title:       "Health Management",
			Description: "Watch your health bar in the top-left corner. Don't let it reach zero!",
			Objective:   "Survive combat and maintain health above 50%",
			Completed:   false,
			Condition:   checkHealthCondition,
		},
		{
			ID:          "inventory",
			Title:       "Inventory System",
			Description: "Press I to open your inventory. Collect items dropped by enemies.",
			Objective:   "Pick up an item and open inventory",
			Completed:   false,
			Condition:   checkInventoryCondition,
		},
		{
			ID:          "skills",
			Title:       "Character Progression",
			Description: "Defeat enemies to gain XP. Level up to become stronger and unlock new abilities!",
			Objective:   "Reach level 2",
			Completed:   false,
			Condition:   checkSkillsCondition,
		},
		{
			ID:          "exploration",
			Title:       "Dungeon Exploration",
			Description: "Explore the dungeon to find treasure, secrets, and the stairs to deeper levels.",
			Objective:   "Visit 3 different areas to complete the tutorial",
			Completed:   false,
			Condition:   checkExplorationCondition,
		},
	}
}

// findPlayerEntity returns the first entity with an input component (the player).
func findPlayerEntity(world *World) *Entity {
	for _, entity := range world.GetEntities() {
		if entity.HasComponent("input") {
			return entity
		}
	}
	return nil
}

// checkWelcomeCondition verifies any key has been pressed.
func checkWelcomeCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil {
		return false
	}
	comp, ok := player.GetComponent("input")
	if !ok {
		return false
	}
	inputProvider, ok := comp.(InputProvider)
	if !ok {
		return false
	}
	return inputProvider.IsAnyKeyPressed()
}

// checkMovementCondition verifies the player has moved at least 50 units total.
// Uses the player_statistics component to track distance traveled during the session,
// which works correctly regardless of screen resolution or spawn position.
func checkMovementCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil || !player.HasComponent("player_statistics") {
		return false
	}
	comp, ok := player.GetComponent("player_statistics")
	if !ok {
		return false
	}
	stats, ok := comp.(*PlayerStatisticsComponent)
	if !ok {
		return false
	}
	// Check if player has moved at least 50 units total this session
	return stats.GetSessionStat("explore_distance_traveled") >= 50
}

// checkCombatCondition verifies the player has attacked at least once.
func checkCombatCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil || !player.HasComponent("attack") {
		return false
	}
	comp, ok := player.GetComponent("attack")
	if !ok {
		return false
	}
	attack, ok := comp.(*AttackComponent)
	if !ok {
		return false
	}
	return attack.CooldownTimer > 0 && attack.CooldownTimer < attack.Cooldown
}

// checkHealthCondition verifies the player has taken damage but remains above 50% health.
func checkHealthCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil || !player.HasComponent("health") {
		return false
	}
	comp, ok := player.GetComponent("health")
	if !ok {
		return false
	}
	health, ok := comp.(*HealthComponent)
	if !ok {
		return false
	}
	return health.Current < health.Max && health.Current > health.Max/2
}

// checkInventoryCondition verifies the player has collected at least one item.
func checkInventoryCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil || !player.HasComponent("inventory") {
		return false
	}
	comp, ok := player.GetComponent("inventory")
	if !ok {
		return false
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return false
	}
	return len(inv.Items) > 0
}

// checkSkillsCondition verifies the player has reached level 2.
func checkSkillsCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil || !player.HasComponent("experience") {
		return false
	}
	comp, ok := player.GetComponent("experience")
	if !ok {
		return false
	}
	exp, ok := comp.(*ExperienceComponent)
	if !ok {
		return false
	}
	return exp.Level >= 2
}

// checkExplorationCondition verifies the player has visited at least 3 areas.
func checkExplorationCondition(world *World) bool {
	player := findPlayerEntity(world)
	if player == nil || !player.HasComponent("player_statistics") {
		return false
	}
	comp, ok := player.GetComponent("player_statistics")
	if !ok {
		return false
	}
	stats, ok := comp.(*PlayerStatisticsComponent)
	if !ok {
		return false
	}
	return stats.GetSessionStat("explore_areas_visited") >= 3
}

// Update processes the tutorial system each frame
func (ts *EbitenTutorialSystem) Update(entities []*Entity, deltaTime float64) {
	if !ts.Enabled || ts.CurrentStepIdx >= len(ts.Steps) {
		return
	}

	// Grace period: skip condition checks for a few frames after the
	// tutorial is first enabled so that held keys from the previous
	// game phase (e.g., ENTER from character creation) don't
	// auto-complete the welcome step.
	if ts.enableGraceFrames > 0 {
		ts.enableGraceFrames--
		ts.updateNotificationTTL(deltaTime)
		return
	}

	ts.updateInputHandlers()

	if ts.handleEscapeKey() {
		return
	}

	if ts.handleNextStepKey() {
		return
	}

	world := ts.createTemporaryWorld(entities)
	ts.updateNotificationTTL(deltaTime)
	ts.checkStepCompletion(world)
}

// updateInputHandlers updates all input handlers for the tutorial.
func (ts *EbitenTutorialSystem) updateInputHandlers() {
	if ts.touchHandler != nil {
		ts.touchHandler.Update()
	}
	if ts.nextButton != nil {
		ts.nextButton.Update()
	}
	if ts.skipButton != nil {
		ts.skipButton.Update()
	}
}

// handleEscapeKey checks for ESC key press to hide tutorial UI.
func (ts *EbitenTutorialSystem) handleEscapeKey() bool {
	if ts.ShowUI && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ts.HideTutorialUI()
		return true
	}
	return false
}

// handleNextStepKey checks for N key press to advance to the next tutorial step.
func (ts *EbitenTutorialSystem) handleNextStepKey() bool {
	if ts.ShowUI && inpututil.IsKeyJustPressed(ebiten.KeyN) {
		ts.AdvanceStep()
		return true
	}
	return false
}

// AdvanceStep manually advances the tutorial to the next step.
// This allows users to skip a step using the N key.
func (ts *EbitenTutorialSystem) AdvanceStep() {
	if ts.CurrentStepIdx < len(ts.Steps) {
		ts.Steps[ts.CurrentStepIdx].Completed = true
		ts.CurrentStepIdx++
		if ts.CurrentStepIdx < len(ts.Steps) {
			ts.NotificationMsg = "Step skipped - " + ts.Steps[ts.CurrentStepIdx].Title
			ts.NotificationTTL = 2.0
		} else {
			ts.Enabled = false
			ts.ShowUI = false
			ts.NotificationMsg = "Tutorial completed!"
			ts.NotificationTTL = 3.0

			// Notify OnboardingManager that in-game tutorial is complete
			if ts.OnCompleteCallback != nil {
				ts.OnCompleteCallback()
			}
		}
	}
}

// createTemporaryWorld creates a temporary world for condition checking.
func (ts *EbitenTutorialSystem) createTemporaryWorld(entities []*Entity) *World {
	world := &World{entities: make(map[uint64]*Entity)}
	world.entityListDirty.Store(true)
	for _, entity := range entities {
		world.entities[entity.ID] = entity
	}
	return world
}

// updateNotificationTTL updates and clears expired notification messages.
func (ts *EbitenTutorialSystem) updateNotificationTTL(deltaTime float64) {
	if ts.NotificationTTL > 0 {
		ts.NotificationTTL -= deltaTime
		if ts.NotificationTTL <= 0 {
			ts.NotificationMsg = ""
		}
	}
}

// checkStepCompletion checks if the current step is completed and advances tutorial.
func (ts *EbitenTutorialSystem) checkStepCompletion(world *World) {
	if ts.CurrentStepIdx >= len(ts.Steps) {
		return
	}
	currentStep := &ts.Steps[ts.CurrentStepIdx]
	if !currentStep.Completed && currentStep.Condition(world) {
		currentStep.Completed = true
		ts.CurrentStepIdx++

		if ts.CurrentStepIdx < len(ts.Steps) {
			ts.showNextStepNotification(currentStep)
		} else {
			ts.showCompletionNotification()
		}
	}
}

// showNextStepNotification displays notification for advancing to next step.
func (ts *EbitenTutorialSystem) showNextStepNotification(completedStep *TutorialStep) {
	nextStep := &ts.Steps[ts.CurrentStepIdx]
	ts.NotificationMsg = fmt.Sprintf("✓ %s Complete! Next: %s", completedStep.Title, nextStep.Title)
	ts.NotificationTTL = 3.0
}

// showCompletionNotification displays tutorial completion notification.
func (ts *EbitenTutorialSystem) showCompletionNotification() {
	ts.NotificationMsg = "Tutorial Complete! You're ready to adventure!"
	ts.NotificationTTL = 5.0
	ts.Enabled = false

	// Notify OnboardingManager that in-game tutorial is complete
	if ts.OnCompleteCallback != nil {
		ts.OnCompleteCallback()
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

// IsStepCompleted returns true if the step with given ID has been completed.
// Returns false if the step ID doesn't exist.
// Thread-safe for read-only access.
func (ts *EbitenTutorialSystem) IsStepCompleted(stepID string) bool {
	for _, step := range ts.Steps {
		if step.ID == stepID {
			return step.Completed
		}
	}
	return false
}

// GetStepByID returns the tutorial step with the given ID, or nil if not found.
// Used for querying specific step details and progress.
// Returns a pointer to the actual step (not a copy), so modifications will affect state.
func (ts *EbitenTutorialSystem) GetStepByID(stepID string) *TutorialStep {
	for i := range ts.Steps {
		if ts.Steps[i].ID == stepID {
			return &ts.Steps[i]
		}
	}
	return nil
}

// IsActive returns true if the tutorial system is currently enabled and showing UI.
// When false, tutorial UI is hidden and tutorial progression is paused.
// Use this to determine whether tutorial overlay should be rendered.
func (ts *EbitenTutorialSystem) IsActive() bool {
	return ts.Enabled && ts.ShowUI
}

// GetCurrentStepID returns the ID of the current step, or empty string if tutorial is complete.
// Convenience method for logging, save/load, and UI integration.
func (ts *EbitenTutorialSystem) GetCurrentStepID() string {
	step := ts.GetCurrentStep()
	if step == nil {
		return ""
	}
	return step.ID
}

// GetAllSteps returns all tutorial steps (read-only copy for UI integration).
// Returns a copy of the steps slice to prevent external modification.
// Use this for displaying tutorial progress, completion status, or step list UI.
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

	// Check if all steps are completed - if so, keep tutorial disabled
	allComplete := len(ts.Steps) > 0
	for _, step := range ts.Steps {
		if !step.Completed {
			allComplete = false
			break
		}
	}
	if allComplete {
		// Tutorial was completed - don't re-enable or clamp the index
		ts.Enabled = false
		ts.ShowUI = false
		ts.CurrentStepIdx = len(ts.Steps)
		return
	}

	// Validate currentStepIdx (in case tutorial steps changed between save/load)
	// Clamp negative values to 0
	if ts.CurrentStepIdx < 0 {
		ts.CurrentStepIdx = 0
	}
	// Clamp values beyond step count to last step (only for incomplete tutorials)
	if ts.CurrentStepIdx >= len(ts.Steps) {
		ts.CurrentStepIdx = len(ts.Steps) - 1
		if ts.CurrentStepIdx < 0 {
			ts.CurrentStepIdx = 0
		}
	}
}

// advanceToNextStep manually advances to the next tutorial step.
// Called by the Next button.
func (ts *EbitenTutorialSystem) advanceToNextStep() {
	if ts.Enabled && ts.CurrentStepIdx < len(ts.Steps) {
		// Mark current step as complete and move to next
		ts.Steps[ts.CurrentStepIdx].Completed = true
		ts.CurrentStepIdx++

		if ts.CurrentStepIdx >= len(ts.Steps) {
			ts.NotificationMsg = "Tutorial Complete!"
			ts.NotificationTTL = 3.0
			ts.Enabled = false

			// Notify OnboardingManager that in-game tutorial is complete
			if ts.OnCompleteCallback != nil {
				ts.OnCompleteCallback()
			}
		}
	}
}

// DisableTutorial completely disables the tutorial system.
// Called by the Skip Tutorial button.
func (ts *EbitenTutorialSystem) DisableTutorial() {
	ts.Enabled = false
	ts.ShowUI = false
	ts.NotificationMsg = "Tutorial skipped"
	ts.NotificationTTL = 2.0

	// Notify OnboardingManager that in-game tutorial was skipped
	if ts.OnCompleteCallback != nil {
		ts.OnCompleteCallback()
	}
}

// HideTutorialUI hides the tutorial overlay without disabling progression.
// BUG FIX: Phase 2 - Allows ESC key to hide tutorial UI while keeping tutorial active
// Players can continue completing tutorial objectives without the overlay visible
func (ts *EbitenTutorialSystem) HideTutorialUI() {
	ts.ShowUI = false
	ts.NotificationMsg = "Tutorial minimized (press H to see controls)"
	ts.NotificationTTL = 3.0
}

// ShowNotification displays a notification message for the specified duration.
// This can be used to show feedback for game actions like saving/loading.
func (ts *EbitenTutorialSystem) ShowNotification(msg string, duration float64) {
	ts.NotificationMsg = msg
	ts.NotificationTTL = duration
}

// Resize updates the tutorial system's screen dimensions and repositions touch buttons.
// This should be called when the screen size changes to ensure buttons remain correctly positioned.
func (ts *EbitenTutorialSystem) Resize(screenWidth, screenHeight int) {
	if ts.screenWidth == screenWidth && ts.screenHeight == screenHeight {
		return // No change needed
	}

	ts.screenWidth = screenWidth
	ts.screenHeight = screenHeight

	// Reposition buttons based on new screen dimensions
	if ts.nextButton != nil {
		ts.nextButton.SetPosition(float64(screenWidth-164), float64(screenHeight-64))
	}
	if ts.skipButton != nil {
		ts.skipButton.SetPosition(44, float64(screenHeight-64))
	}
}

// Draw renders the tutorial UI overlay (implements UISystem interface).
// The screen parameter should be *ebiten.Image in production.
func (ts *EbitenTutorialSystem) Draw(screen interface{}) {
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok || !ts.shouldDrawTutorialUI() {
		return
	}

	// Update button positions to match actual screen size
	ts.Resize(ebitenScreen.Bounds().Dx(), ebitenScreen.Bounds().Dy())

	step := ts.GetCurrentStep()
	if step == nil {
		if ts.NotificationTTL > 0 {
			ts.drawNotification(ebitenScreen)
		}
		return
	}

	ts.drawTutorialPanel(ebitenScreen, step)

	if ts.NotificationTTL > 0 {
		ts.drawNotification(ebitenScreen)
	}
}

// shouldDrawTutorialUI returns true if the tutorial UI should be drawn.
func (ts *EbitenTutorialSystem) shouldDrawTutorialUI() bool {
	return ts.Enabled && ts.ShowUI
}

// drawTutorialPanel renders the complete tutorial panel.
func (ts *EbitenTutorialSystem) drawTutorialPanel(screen *ebiten.Image, step *TutorialStep) {
	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	panelWidth, panelX, panelY := ts.calculatePanelLayout(screenWidth, screenHeight)

	ts.drawPanelBackground(screen, panelX, panelY, panelWidth, 150)
	ts.drawPanelContent(screen, step, panelX, panelY, panelWidth)
	ts.drawTutorialButtons(screen)
}

// calculatePanelLayout calculates panel dimensions and position.
func (ts *EbitenTutorialSystem) calculatePanelLayout(screenWidth, screenHeight int) (panelWidth, panelX, panelY int) {
	panelWidth = 400
	if screenWidth < 800 {
		if screenWidth-100 > 300 {
			panelWidth = screenWidth - 100
		} else {
			panelWidth = 300
		}
	}
	panelHeight := 150

	const hudMarginBottom = 60

	if screenWidth >= 800 && screenHeight >= 600 {
		panelX = screenWidth - panelWidth - 20
		panelY = screenHeight - panelHeight - hudMarginBottom
	} else if screenHeight >= 400 {
		panelX = (screenWidth - panelWidth) / 2
		panelY = screenHeight - panelHeight - hudMarginBottom
	} else {
		panelX = (screenWidth - panelWidth) / 2
		panelY = (screenHeight - panelHeight) / 2
	}

	return panelWidth, panelX, panelY
}

// drawPanelBackground renders the panel background, border, and progress bar.
func (ts *EbitenTutorialSystem) drawPanelBackground(screen *ebiten.Image, x, y, width, height int) {
	vector.DrawFilledRect(screen,
		float32(x), float32(y),
		float32(width), float32(height),
		color.RGBA{0, 0, 0, 200}, false)

	vector.StrokeRect(screen,
		float32(x), float32(y),
		float32(width), float32(height),
		2, color.RGBA{100, 200, 255, 255}, false)

	progressWidth := int(float64(width-20) * ts.GetProgress())
	vector.DrawFilledRect(screen,
		float32(x+10), float32(y+10),
		float32(progressWidth), 4,
		color.RGBA{100, 200, 255, 255}, false)
}

// drawPanelContent renders the tutorial text content.
func (ts *EbitenTutorialSystem) drawPanelContent(screen *ebiten.Image, step *TutorialStep, x, y, width int) {
	titleColor := color.RGBA{255, 255, 100, 255}
	text.Draw(screen, fmt.Sprintf("Tutorial (%d/%d)", ts.CurrentStepIdx+1, len(ts.Steps)),
		basicfont.Face7x13, x+10, y+35, titleColor)

	text.Draw(screen, step.Title, basicfont.Face7x13, x+10, y+55, color.White)

	// Use class-aware description for personalized hints
	descColor := color.RGBA{200, 200, 200, 255}
	description := ts.getClassAwareDescription(step)
	ts.drawWrappedText(screen, description, x+10, y+75, width-20, descColor)

	objColor := color.RGBA{100, 255, 100, 255}
	text.Draw(screen, "Objective: "+step.Objective, basicfont.Face7x13, x+10, y+120, objColor)

	hintColor := color.RGBA{150, 150, 150, 255}
	text.Draw(screen, "ESC: minimize | N: next step", basicfont.Face7x13, x+10, y+140, hintColor)
}

// drawTutorialButtons renders the touch buttons.
func (ts *EbitenTutorialSystem) drawTutorialButtons(screen *ebiten.Image) {
	if ts.nextButton != nil {
		ts.nextButton.Draw(screen)
	}
	if ts.skipButton != nil {
		ts.skipButton.Draw(screen)
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
// Controls tutorial UI visibility. When set to false, hides tutorial overlay
// but does not disable tutorial progression (use SkipAll() to fully disable).
func (ts *EbitenTutorialSystem) SetActive(active bool) {
	ts.ShowUI = active
}

// Compile-time interface check
var _ UISystem = (*EbitenTutorialSystem)(nil)
