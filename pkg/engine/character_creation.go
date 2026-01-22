// Package engine provides character creation functionality for onboarding new players.
// This file implements the character creation UI and class selection system that
// integrates with the tutorial flow for a unified onboarding experience.
package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/recovery"
	"golang.org/x/image/draw"
	"golang.org/x/image/font/basicfont"
)

// GetDefaultPicturesDirectory returns the user's Pictures directory path
// Cross-platform: Works on Windows, macOS, Linux, and mobile
func GetDefaultPicturesDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	switch runtime.GOOS {
	case "windows":
		// Windows: %USERPROFILE%\Pictures
		return filepath.Join(homeDir, "Pictures")
	case "darwin":
		// macOS: ~/Pictures
		return filepath.Join(homeDir, "Pictures")
	case "linux":
		// Linux: ~/Pictures (XDG standard)
		return filepath.Join(homeDir, "Pictures")
	case "android", "ios":
		// Mobile: Use app's documents directory or similar
		// On mobile, file picking is handled differently via native pickers
		return homeDir
	default:
		return homeDir
	}
}

// CharacterClass represents a player archetype with specific stat distributions
type CharacterClass int

const (
	// Base Classes (6 original)
	// ClassWarrior is a high HP, melee-focused class
	ClassWarrior CharacterClass = iota
	// ClassMage is a high mana, magic-focused class
	ClassMage
	// ClassRogue is a balanced, agility-focused class
	ClassRogue
	// ClassRanger is a ranged combat class with pet bonding abilities (V4 Phase 25)
	ClassRanger
	// ClassCleric is a support class with healing and buffs (V4 Phase 25)
	ClassCleric
	// ClassNecromancer is a summoning class with life drain and debuffs (V4 Phase 25)
	ClassNecromancer

	// Hybrid Classes (15 combinations) - Phase 25.2 Extension
	// ClassBattlemage combines Warrior melee prowess with Mage spellcasting
	ClassBattlemage
	// ClassSpellblade combines Rogue agility with Mage magic
	ClassSpellblade
	// ClassPaladin combines Warrior strength with Cleric holy powers
	ClassPaladin
	// ClassMonk combines Rogue speed with Cleric spiritual discipline
	ClassMonk
	// ClassDeathKnight combines Warrior combat with Necromancer dark magic
	ClassDeathKnight
	// ClassWitchHunter combines Ranger precision with Cleric divine power
	ClassWitchHunter
	// ClassBeastlord combines Warrior might with Ranger beast mastery
	ClassBeastlord
	// ClassArcaneArcher combines Ranger marksmanship with Mage arcane arts
	ClassArcaneArcher
	// ClassShadowPriest combines Rogue shadows with Necromancer dark arts
	ClassShadowPriest
	// ClassDruid combines Ranger nature affinity with Mage elemental magic
	ClassDruid
	// ClassInquisitor combines Cleric faith with Rogue investigation
	ClassInquisitor
	// ClassBloodKnight combines Warrior combat with Necromancer blood magic
	ClassBloodKnight
	// ClassMystic combines Mage arcane knowledge with Cleric divine wisdom
	ClassMystic
	// ClassWarlock combines Mage magic with Necromancer dark pacts
	ClassWarlock
	// ClassNinja combines Rogue stealth with Ranger precision strikes
	ClassNinja
)

// String returns the human-readable class name
func (c CharacterClass) String() string {
	switch c {
	case ClassWarrior:
		return "Warrior"
	case ClassMage:
		return "Mage"
	case ClassRogue:
		return "Rogue"
	case ClassRanger:
		return "Ranger"
	case ClassCleric:
		return "Cleric"
	case ClassNecromancer:
		return "Necromancer"
	// Hybrid classes
	case ClassBattlemage:
		return "Battlemage"
	case ClassSpellblade:
		return "Spellblade"
	case ClassPaladin:
		return "Paladin"
	case ClassMonk:
		return "Monk"
	case ClassDeathKnight:
		return "Death Knight"
	case ClassWitchHunter:
		return "Witch Hunter"
	case ClassBeastlord:
		return "Beastlord"
	case ClassArcaneArcher:
		return "Arcane Archer"
	case ClassShadowPriest:
		return "Shadow Priest"
	case ClassDruid:
		return "Druid"
	case ClassInquisitor:
		return "Inquisitor"
	case ClassBloodKnight:
		return "Blood Knight"
	case ClassMystic:
		return "Mystic"
	case ClassWarlock:
		return "Warlock"
	case ClassNinja:
		return "Ninja"
	default:
		return "Unknown"
	}
}

// Description returns a short description of the class
func (c CharacterClass) Description() string {
	switch c {
	case ClassWarrior:
		return "Masters of melee combat with high HP and defense. Use WASD to move and SPACE to attack."
	case ClassMage:
		return "Wielders of arcane magic with powerful spells. Press 1-5 to cast spells. Low HP, high mana."
	case ClassRogue:
		return "Agile fighters with balanced stats and critical strikes. Quick attacks and evasion."
	case ClassRanger:
		return "Skilled archer and beast tamer. Excels at ranged combat and can bond with companions."
	case ClassCleric:
		return "Divine caster who heals allies and smites enemies. Balances support with holy combat."
	case ClassNecromancer:
		return "Dark mage who commands the undead. Summons minions and drains life force."
	// Hybrid classes
	case ClassBattlemage:
		return "Armored spellcaster combining martial prowess with destructive magic. High versatility."
	case ClassSpellblade:
		return "Agile warrior-mage weaving spells between swift strikes. Magic enhances combat."
	case ClassPaladin:
		return "Holy warrior blending heavy armor with divine healing. Protects allies with faith."
	case ClassMonk:
		return "Unarmed combatant using spiritual energy and incredible speed. Discipline over equipment."
	case ClassDeathKnight:
		return "Fallen warrior wielding dark necromantic powers. Life drain sustains in battle."
	case ClassWitchHunter:
		return "Divine marksman specializing in hunting supernatural threats. Faith guides arrows."
	case ClassBeastlord:
		return "Savage warrior commanding powerful beasts. Fights alongside animal companions."
	case ClassArcaneArcher:
		return "Ranger infusing arrows with arcane energy. Magic projectiles pierce defenses."
	case ClassShadowPriest:
		return "Stealthy cleric wielding shadow magic and forbidden knowledge. Darkness heals."
	case ClassDruid:
		return "Nature guardian shapeshifting between forms. Controls elements and beasts."
	case ClassInquisitor:
		return "Holy investigator rooting out corruption with divine judgment. Truth through faith."
	case ClassBloodKnight:
		return "Warrior sacrificing health for devastating blood magic attacks. Pain fuels power."
	case ClassMystic:
		return "Enlightened caster balancing arcane and divine magic. Wisdom guides spells."
	case ClassWarlock:
		return "Pact-bound mage wielding eldritch powers. Dark bargains grant forbidden magic."
	case ClassNinja:
		return "Master assassin combining stealth with precise strikes. Shadows are allies."
	default:
		return ""
	}
}

// LowerName returns the lowercase name of the class for matching with item restrictions.
// Phase 25.2: Used for class-specific equipment restrictions.
func (c CharacterClass) LowerName() string {
	switch c {
	case ClassWarrior:
		return "warrior"
	case ClassMage:
		return "mage"
	case ClassRogue:
		return "rogue"
	case ClassRanger:
		return "ranger"
	case ClassCleric:
		return "cleric"
	case ClassNecromancer:
		return "necromancer"
	// Hybrid classes
	case ClassBattlemage:
		return "battlemage"
	case ClassSpellblade:
		return "spellblade"
	case ClassPaladin:
		return "paladin"
	case ClassMonk:
		return "monk"
	case ClassDeathKnight:
		return "deathknight"
	case ClassWitchHunter:
		return "witchhunter"
	case ClassBeastlord:
		return "beastlord"
	case ClassArcaneArcher:
		return "arcanearcher"
	case ClassShadowPriest:
		return "shadowpriest"
	case ClassDruid:
		return "druid"
	case ClassInquisitor:
		return "inquisitor"
	case ClassBloodKnight:
		return "bloodknight"
	case ClassMystic:
		return "mystic"
	case ClassWarlock:
		return "warlock"
	case ClassNinja:
		return "ninja"
	default:
		return "unknown"
	}
}

// CharacterData holds the player's character creation choices
type CharacterData struct {
	Name         string
	Class        CharacterClass
	PortraitPath string        // Path to user's custom portrait image (optional)
	Portrait     *ebiten.Image // Loaded portrait image (optional, max 512x512)
}

// CharacterCreationDefaults holds custom default values for character creation
type CharacterCreationDefaults struct {
	DefaultName         string         // Default name to pre-fill
	DefaultClass        CharacterClass // Default class to pre-select
	DefaultPortraitPath string         // Default portrait path to pre-fill
}

// Validate checks if the character data is valid
func (cd *CharacterData) Validate() error {
	// Trim whitespace
	cd.Name = strings.TrimSpace(cd.Name)

	if cd.Name == "" {
		return fmt.Errorf("character name cannot be empty")
	}
	if len(cd.Name) > 20 {
		return fmt.Errorf("character name too long (max 20 characters)")
	}
	if cd.Class < ClassWarrior || cd.Class > ClassNecromancer {
		return fmt.Errorf("invalid character class")
	}
	return nil
}

// LoadPortrait loads a portrait image from the given path and downscales if needed
// Maximum size is 512x512, aspect ratio is preserved
func LoadPortrait(path string) (*ebiten.Image, error) {
	if path == "" {
		return nil, nil // No portrait is valid
	}

	// Validate extension first
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".png" {
		return nil, fmt.Errorf("portrait must be a .png file, got: %s", ext)
	}

	// Validate file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("portrait file not found: %s", path)
	}

	// Load image
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open portrait: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Downscale if needed (max 512x512, preserve aspect ratio)
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	const maxSize = 512
	if width > maxSize || height > maxSize {
		// Calculate scale factor to fit within maxSize x maxSize
		scale := float64(maxSize) / float64(max(width, height))
		newWidth := int(float64(width) * scale)
		newHeight := int(float64(height) * scale)

		// Create scaled image
		scaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.BiLinear.Scale(scaled, scaled.Bounds(), img, bounds, draw.Over, nil)
		img = scaled
	}

	// Convert to Ebiten image
	return ebiten.NewImageFromImage(img), nil
}

// OpenPortraitDialog opens a native file picker dialog for selecting a portrait image
// Returns the selected file path or empty string if cancelled
// Platform-specific implementations are provided via:
// - character_creation_desktop.go for desktop platforms (uses zenity)
// - character_creation_mobile.go for mobile/WASM platforms (returns error)

// creationStep represents the current step in character creation
type creationStep int

const (
	stepNameInput creationStep = iota
	stepClassSelection
	stepPortraitSelection
	stepConfirmation
)

// EbitenCharacterCreation handles the character creation UI and flow
type EbitenCharacterCreation struct {
	currentStep   creationStep
	characterData CharacterData
	nameInput     string
	selectedClass CharacterClass
	portraitInput string // Path input for portrait image
	confirmed     bool
	errorMsg      string

	// Custom defaults
	defaults CharacterCreationDefaults

	// Input state
	inputBuffer []rune

	// Step transition state - prevents same key press from being processed by multiple steps
	stepChangedThisFrame bool

	// Mobile keyboard state (WASM/mobile platforms)
	keyboardShown bool // Tracks whether mobile keyboard is currently shown

	screenWidth  int
	screenHeight int

	// Panel layout cache (calculated in Draw, used in Update for hit detection)
	panelX      int
	panelY      int
	panelWidth  int
	panelHeight int

	// WASM FIX: Lazy portrait loading
	// Store portrait path to load and defer actual image loading until Draw()
	// This prevents ebiten.NewImageFromImage() calls before graphics context is ready
	pendingPortraitPath   string
	portraitLoadAttempted bool

	// Touch support - WASM/mobile touch navigation
	touchHandler *mobile.TouchInputHandler
	nextButton   *mobile.TouchButton
	backButton   *mobile.TouchButton
	skipButton   *mobile.TouchButton // For skipping portrait selection

	// Preset name buttons for WASM/mobile fallback
	presetNameButtons []*mobile.TouchButton
}

// NewCharacterCreation creates a new character creation system
func NewCharacterCreation(screenWidth, screenHeight int) *EbitenCharacterCreation {
	cc := &EbitenCharacterCreation{
		currentStep:   stepNameInput,
		selectedClass: ClassWarrior, // Default selection
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		inputBuffer:   make([]rune, 0),
		defaults: CharacterCreationDefaults{
			DefaultName:  "", // No default initially
			DefaultClass: ClassWarrior,
		},
		touchHandler: mobile.NewTouchInputHandler(),
	}

	// Create touch buttons (positioned dynamically in updatePanelDimensions)
	// Next button (bottom-right) - advances to next step
	cc.nextButton = mobile.NewTouchButton(
		0, 0, // Position updated dynamically
		120, 44,
		"Next",
		func() { cc.handleNextButton() },
	)
	// Make Next button more visible with brighter colors
	cc.nextButton.BackgroundColor = color.RGBA{50, 100, 200, 255}
	cc.nextButton.PressedColor = color.RGBA{70, 140, 255, 255}
	cc.nextButton.BorderColor = color.RGBA{100, 150, 255, 255}

	// Back button (bottom-left) - returns to previous step
	cc.backButton = mobile.NewTouchButton(
		0, 0, // Position updated dynamically
		120, 44,
		"Back",
		func() { cc.handleBackButton() },
	)
	// Make Back button visible with distinct colors
	cc.backButton.BackgroundColor = color.RGBA{120, 60, 60, 255}
	cc.backButton.PressedColor = color.RGBA{180, 90, 90, 255}
	cc.backButton.BorderColor = color.RGBA{200, 100, 100, 255}

	// Skip button (bottom-center) - skips portrait selection
	cc.skipButton = mobile.NewTouchButton(
		0, 0, // Position updated dynamically
		120, 44,
		"Skip",
		func() { cc.handleSkipButton() },
	)
	// Make Skip button visible
	cc.skipButton.BackgroundColor = color.RGBA{100, 100, 50, 255}
	cc.skipButton.PressedColor = color.RGBA{150, 150, 70, 255}
	cc.skipButton.BorderColor = color.RGBA{180, 180, 100, 255}

	// Create preset name buttons (for WASM/mobile fallback)
	presetNames := []string{"Warrior", "Mage", "Rogue", "Ranger", "Auto"}
	cc.presetNameButtons = make([]*mobile.TouchButton, len(presetNames))
	for i, name := range presetNames {
		presetName := name // Capture for closure
		cc.presetNameButtons[i] = mobile.NewTouchButton(
			0, 0, // Position updated dynamically
			100, 36,
			presetName,
			func() { cc.handlePresetName(presetName) },
		)
	}

	return cc
}

// SetDefaults sets custom default values for character creation
func (cc *EbitenCharacterCreation) SetDefaults(defaults CharacterCreationDefaults) {
	cc.defaults = defaults
	// Apply defaults to current state
	if cc.currentStep == stepNameInput && cc.defaults.DefaultName != "" {
		cc.nameInput = cc.defaults.DefaultName
	}
	if cc.currentStep == stepClassSelection {
		cc.selectedClass = cc.defaults.DefaultClass
	}
}

// SetDefaultNameFromSeed sets the default character name based on world seed.
// Uses deterministic selection to ensure the same seed always produces the same name.
func (cc *EbitenCharacterCreation) SetDefaultNameFromSeed(seed int64) {
	defaultName := procgen.SelectDefaultName(seed)
	cc.defaults.DefaultName = defaultName
	// Apply to current state if we're in name input step
	if cc.currentStep == stepNameInput {
		cc.nameInput = defaultName
	}
}

// GetDefaults returns the current default values
func (cc *EbitenCharacterCreation) GetDefaults() CharacterCreationDefaults {
	return cc.defaults
}

// Update handles input for character creation (keyboard/mouse navigation)
// Returns true when character creation is complete
func (cc *EbitenCharacterCreation) Update() bool {
	// Reset step change flag at start of frame
	cc.stepChangedThisFrame = false

	// Calculate panel dimensions first (needed for touch hit detection)
	// This must be done before processing input
	cc.updatePanelDimensions()

	// Update touch handler and buttons
	if cc.touchHandler != nil {
		cc.touchHandler.Update()
	}

	// Update touch button positions based on panel layout
	cc.updateTouchButtonPositions()

	// Update touch buttons (conditionally based on current step)
	// Next button is always visible
	if cc.nextButton != nil {
		cc.nextButton.Update()
	}

	// Back button only on steps after nameInput
	if cc.backButton != nil && cc.currentStep != stepNameInput {
		cc.backButton.Update()
	}

	// Skip button only on portrait selection step
	if cc.skipButton != nil && cc.currentStep == stepPortraitSelection {
		cc.skipButton.Update()
	}

	// Update preset name buttons (only in name input step)
	if cc.currentStep == stepNameInput {
		for _, btn := range cc.presetNameButtons {
			if btn != nil {
				btn.Update()
			}
		}
	}

	switch cc.currentStep {
	case stepNameInput:
		cc.updateNameInput()
	case stepClassSelection:
		cc.updateClassSelection()
	case stepPortraitSelection:
		cc.updatePortraitSelection()
	case stepConfirmation:
		cc.updateConfirmation()
	}

	return cc.confirmed
}

// updateTouchButtonPositions positions touch buttons based on panel layout
func (cc *EbitenCharacterCreation) updateTouchButtonPositions() {
	// Next button (bottom-right of panel)
	if cc.nextButton != nil {
		nextX := cc.panelX + cc.panelWidth - 140
		nextY := cc.panelY + cc.panelHeight - 60
		cc.nextButton.SetPosition(
			float64(nextX),
			float64(nextY),
		)
	}

	// Back button (bottom-left of panel)
	if cc.backButton != nil {
		cc.backButton.SetPosition(
			float64(cc.panelX+20),
			float64(cc.panelY+cc.panelHeight-60),
		)
	}

	// Skip button (bottom-center of panel)
	if cc.skipButton != nil {
		cc.skipButton.SetPosition(
			float64(cc.panelX+cc.panelWidth/2-60),
			float64(cc.panelY+cc.panelHeight-60),
		)
	}

	// Position preset name buttons (only in name input step)
	// Arrange horizontally below the input field
	if cc.currentStep == stepNameInput {
		buttonSpacing := 10
		buttonWidth := 100
		totalWidth := len(cc.presetNameButtons)*buttonWidth + (len(cc.presetNameButtons)-1)*buttonSpacing
		startX := cc.panelX + cc.panelWidth/2 - totalWidth/2
		buttonY := cc.panelY + 200 // Below the input box

		for i, btn := range cc.presetNameButtons {
			if btn != nil {
				btn.SetPosition(
					float64(startX+i*(buttonWidth+buttonSpacing)),
					float64(buttonY),
				)
			}
		}
	}
}

// updatePanelDimensions calculates the panel layout
// Called from both Update (for hit detection) and Draw (for rendering)
func (cc *EbitenCharacterCreation) updatePanelDimensions() {
	cc.panelWidth = 600
	cc.panelHeight = 400
	cc.panelX = cc.screenWidth/2 - cc.panelWidth/2
	cc.panelY = cc.screenHeight/2 - cc.panelHeight/2
}

// updateNameInput handles name input with keyboard
func (cc *EbitenCharacterCreation) updateNameInput() {
	cc.showKeyboardIfNeeded()
	cc.processTextInput()
	cc.handleBackspaceKey()
	cc.handleEnterKey()
	cc.handleDefaultNameSave()
}

// showKeyboardIfNeeded shows mobile keyboard when entering name input
func (cc *EbitenCharacterCreation) showKeyboardIfNeeded() {
	if !cc.keyboardShown && mobile.IsWASM() {
		mobile.ShowKeyboard()
		cc.keyboardShown = true
	}
}

// processTextInput appends valid alphanumeric characters to name input
func (cc *EbitenCharacterCreation) processTextInput() {
	cc.inputBuffer = ebiten.AppendInputChars(cc.inputBuffer[:0])
	for _, r := range cc.inputBuffer {
		if isValidNameCharacter(r) && len(cc.nameInput) < 20 {
			cc.nameInput += string(r)
		}
	}
}

// isValidNameCharacter returns true if the rune is alphanumeric or space
func isValidNameCharacter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' '
}

// handleBackspaceKey removes last character from name input
func (cc *EbitenCharacterCreation) handleBackspaceKey() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(cc.nameInput) > 0 {
		cc.nameInput = cc.nameInput[:len(cc.nameInput)-1]
	}
}

// handleEnterKey validates name and proceeds to class selection
func (cc *EbitenCharacterCreation) handleEnterKey() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) || cc.stepChangedThisFrame {
		return
	}

	if len(strings.TrimSpace(cc.nameInput)) > 0 {
		cc.proceedToClassSelection()
	} else {
		cc.errorMsg = "Name cannot be empty"
	}
}

// proceedToClassSelection advances to class selection step
func (cc *EbitenCharacterCreation) proceedToClassSelection() {
	cc.characterData.Name = cc.nameInput
	cc.currentStep = stepClassSelection
	cc.stepChangedThisFrame = true
	cc.errorMsg = ""
	cc.hideKeyboardIfNeeded()
}

// hideKeyboardIfNeeded hides mobile keyboard when leaving name input
func (cc *EbitenCharacterCreation) hideKeyboardIfNeeded() {
	if cc.keyboardShown && mobile.IsWASM() {
		mobile.HideKeyboard()
		cc.keyboardShown = false
	}
}

// handleDefaultNameSave saves current name as default on F2
func (cc *EbitenCharacterCreation) handleDefaultNameSave() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) && len(strings.TrimSpace(cc.nameInput)) > 0 {
		cc.defaults.DefaultName = strings.TrimSpace(cc.nameInput)
		cc.errorMsg = "Default name saved!"
	}
}

// updateClassSelection handles class selection with keyboard/mouse
func (cc *EbitenCharacterCreation) updateClassSelection() {
	cc.handleArrowKeySelection()
	cc.handleNumberKeySelection()
	if cc.handleTouchOrMouseClick() {
		return
	}
	cc.handleTouchOrMouseHover()
	cc.handleConfirmationKeys()
	cc.handleBackKeys()
	cc.handleDefaultSave()
}

// handleArrowKeySelection processes arrow key navigation for class selection
func (cc *EbitenCharacterCreation) handleArrowKeySelection() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		cc.selectedClass--
		if cc.selectedClass < ClassWarrior {
			cc.selectedClass = ClassRogue
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		cc.selectedClass++
		if cc.selectedClass > ClassRogue {
			cc.selectedClass = ClassWarrior
		}
	}
}

// handleNumberKeySelection processes numeric key shortcuts for direct class selection
func (cc *EbitenCharacterCreation) handleNumberKeySelection() {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		cc.selectedClass = ClassWarrior
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		cc.selectedClass = ClassMage
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		cc.selectedClass = ClassRogue
	}
}

// handleTouchOrMouseClick processes touch and mouse click events for class selection
func (cc *EbitenCharacterCreation) handleTouchOrMouseClick() bool {
	if !IsTouchOrMouseJustPressed() {
		return false
	}
	mouseX, mouseY, _ := GetTouchOrMousePosition()
	startY := cc.panelY + 140
	classes := []CharacterClass{ClassWarrior, ClassMage, ClassRogue}

	for i, class := range classes {
		if cc.isClassBoxClicked(mouseX, mouseY, startY, i) {
			cc.selectedClass = class
			cc.characterData.Class = cc.selectedClass
			cc.currentStep = stepPortraitSelection
			return true
		}
	}
	return false
}

// isClassBoxClicked checks if coordinates are within a class option box
func (cc *EbitenCharacterCreation) isClassBoxClicked(mouseX, mouseY, startY, classIndex int) bool {
	classY := startY + classIndex*80
	return mouseX >= cc.panelX+40 && mouseX <= cc.panelX+cc.panelWidth-40 &&
		mouseY >= classY-5 && mouseY <= classY+65
}

// handleTouchOrMouseHover updates selection based on mouse/touch hover position
func (cc *EbitenCharacterCreation) handleTouchOrMouseHover() {
	mouseX, mouseY, _ := GetTouchOrMousePosition()
	startY := cc.panelY + 140
	classes := []CharacterClass{ClassWarrior, ClassMage, ClassRogue}

	for i, class := range classes {
		if cc.isClassBoxClicked(mouseX, mouseY, startY, i) {
			cc.selectedClass = class
			break
		}
	}
}

// handleConfirmationKeys processes Enter key to confirm selection
func (cc *EbitenCharacterCreation) handleConfirmationKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && !cc.stepChangedThisFrame {
		cc.characterData.Class = cc.selectedClass
		cc.currentStep = stepPortraitSelection
		cc.stepChangedThisFrame = true
	}
}

// handleBackKeys processes Backspace/Escape to return to previous step
func (cc *EbitenCharacterCreation) handleBackKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepNameInput
		cc.keyboardShown = false
	}
}

// handleDefaultSave processes F2 key to save current class as default
func (cc *EbitenCharacterCreation) handleDefaultSave() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		cc.defaults.DefaultClass = cc.selectedClass
		cc.errorMsg = fmt.Sprintf("Default class saved: %s", cc.selectedClass.String())
	}
}

// updatePortraitSelection handles portrait file selection via dialog
func (cc *EbitenCharacterCreation) updatePortraitSelection() {
	// MOBILE/WASM FIX: Don't show keyboard automatically on portrait step.
	// File path input is not practical on mobile devices. Users should press
	// Tab to skip portrait, or Enter with empty input to proceed.
	// If they do start typing, the keyboard will appear automatically via browser behavior.
	// This prevents unnecessary keyboard popup that blocks the UI.

	if cc.handlePortraitTouchInput() {
		return
	}

	if cc.handlePortraitKeyboardShortcuts() {
		return
	}

	cc.handlePortraitTextInput()

	if cc.handlePortraitBackspace() {
		return
	}

	if cc.handlePortraitNavigation() {
		return
	}

	if cc.handlePortraitConfirmation() {
		return
	}

	cc.handlePortraitDefaults()
}

// handlePortraitTouchInput processes mouse and touch input for portrait selection buttons.
func (cc *EbitenCharacterCreation) handlePortraitTouchInput() bool {
	if !IsTouchOrMouseJustPressed() {
		return false
	}

	mouseX, mouseY, _ := GetTouchOrMousePosition()
	helpY := cc.panelY + cc.panelHeight - 100
	buttonY := helpY - 10
	buttonX := cc.panelX + 50
	buttonW := cc.panelWidth - 100
	buttonH := 25

	browseButtonY := buttonY
	skipButtonY := browseButtonY + 35
	backButtonY := skipButtonY + 35

	if cc.checkPortraitBrowseButton(mouseX, mouseY, buttonX, browseButtonY, buttonW, buttonH) {
		return true
	}

	if cc.checkPortraitSkipButton(mouseX, mouseY, buttonX, skipButtonY, buttonW, buttonH) {
		return true
	}

	if cc.checkPortraitBackButton(mouseX, mouseY, buttonX, backButtonY, buttonW, buttonH) {
		return true
	}

	return false
}

// checkPortraitBrowseButton checks if browse button was clicked and triggers file dialog.
func (cc *EbitenCharacterCreation) checkPortraitBrowseButton(mouseX, mouseY, buttonX, buttonY, buttonW, buttonH int) bool {
	if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
		mouseY >= buttonY && mouseY <= buttonY+buttonH {
		go func() {
			defer recovery.RecoverPanicWithLogger("character_creation", "portrait dialog", nil)()
			filename, err := OpenPortraitDialog()
			if err != nil {
				cc.errorMsg = fmt.Sprintf("Dialog error: %v", err)
				return
			}
			if filename != "" {
				cc.portraitInput = filename
			}
		}()
		return true
	}
	return false
}

// checkPortraitSkipButton checks if skip button was clicked and advances to confirmation.
func (cc *EbitenCharacterCreation) checkPortraitSkipButton(mouseX, mouseY, buttonX, buttonY, buttonW, buttonH int) bool {
	if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
		mouseY >= buttonY && mouseY <= buttonY+buttonH {
		cc.skipPortrait()
		return true
	}
	return false
}

// checkPortraitBackButton checks if back button was clicked and returns to class selection.
func (cc *EbitenCharacterCreation) checkPortraitBackButton(mouseX, mouseY, buttonX, buttonY, buttonW, buttonH int) bool {
	if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
		mouseY >= buttonY && mouseY <= buttonY+buttonH {
		cc.returnToClassSelection()
		return true
	}
	return false
}

// handlePortraitKeyboardShortcuts processes keyboard shortcuts for opening file browser.
func (cc *EbitenCharacterCreation) handlePortraitKeyboardShortcuts() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyB) {
		go func() {
			defer recovery.RecoverPanicWithLogger("character_creation", "portrait keyboard shortcut", nil)()
			filename, err := OpenPortraitDialog()
			if err != nil {
				cc.errorMsg = fmt.Sprintf("Dialog error: %v", err)
				return
			}
			if filename != "" {
				cc.portraitInput = filename
			}
		}()
		return true
	}
	return false
}

// handlePortraitTextInput processes manual text input for file path entry.
func (cc *EbitenCharacterCreation) handlePortraitTextInput() {
	cc.inputBuffer = ebiten.AppendInputChars(cc.inputBuffer[:0])
	for _, r := range cc.inputBuffer {
		if r >= 32 && r <= 126 {
			cc.portraitInput += string(r)
		}
	}
}

// handlePortraitBackspace processes backspace key for text deletion or navigation.
func (cc *EbitenCharacterCreation) handlePortraitBackspace() bool {
	if !inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		return false
	}

	if len(cc.portraitInput) > 0 {
		cc.portraitInput = cc.portraitInput[:len(cc.portraitInput)-1]
		return false
	}

	cc.returnToClassSelection()
	return true
}

// handlePortraitNavigation processes ESC and Tab keys for navigation.
func (cc *EbitenCharacterCreation) handlePortraitNavigation() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.returnToClassSelection()
		return true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		cc.skipPortrait()
		return true
	}

	return false
}

// handlePortraitConfirmation processes Enter key for portrait confirmation and loading.
func (cc *EbitenCharacterCreation) handlePortraitConfirmation() bool {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) || cc.stepChangedThisFrame {
		return false
	}

	portraitPath := strings.TrimSpace(cc.portraitInput)

	if portraitPath == "" {
		cc.skipPortrait()
		return true
	}

	cc.characterData.PortraitPath = portraitPath
	cc.pendingPortraitPath = portraitPath
	cc.portraitLoadAttempted = false
	cc.characterData.Portrait = nil
	cc.currentStep = stepConfirmation
	cc.stepChangedThisFrame = true // Mark that we changed steps this frame
	cc.errorMsg = ""
	cc.hideKeyboardIfNeeded()
	return true
}

// handlePortraitDefaults processes F2 key for saving default portrait path.
func (cc *EbitenCharacterCreation) handlePortraitDefaults() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		if len(strings.TrimSpace(cc.portraitInput)) > 0 {
			cc.defaults.DefaultPortraitPath = strings.TrimSpace(cc.portraitInput)
			cc.errorMsg = "Default portrait path saved!"
		}
	}
}

// skipPortrait clears portrait data and advances to confirmation step.
func (cc *EbitenCharacterCreation) skipPortrait() {
	cc.characterData.PortraitPath = ""
	cc.characterData.Portrait = nil
	cc.currentStep = stepConfirmation
	cc.stepChangedThisFrame = true // Mark that we changed steps this frame
	cc.errorMsg = ""
	cc.hideKeyboardIfNeeded()
}

// returnToClassSelection navigates back to class selection step.
func (cc *EbitenCharacterCreation) returnToClassSelection() {
	cc.currentStep = stepClassSelection
	cc.hideKeyboardIfNeeded()
}

// handlePresetName handles preset name button presses
// Provides quick name options for WASM/mobile users if keyboard doesn't work
func (cc *EbitenCharacterCreation) handlePresetName(preset string) {
	if preset == "Auto" {
		// Auto-generate a random name based on current class
		cc.nameInput = cc.generateRandomName()
	} else {
		// Use preset name directly
		cc.nameInput = preset
	}
	cc.errorMsg = "Name set to: " + cc.nameInput
}

// generateRandomName generates a random character name based on selected class
// Provides fallback for WASM/mobile if keyboard isn't working
func (cc *EbitenCharacterCreation) generateRandomName() string {
	// Simple name generation - combine prefix and suffix
	prefixes := []string{"Brave", "Swift", "Dark", "Elder", "Noble", "Shadow", "Storm", "Iron", "Silver", "Golden"}
	suffixes := []string{"blade", "heart", "fist", "eye", "soul", "wind", "fire", "steel", "wing", "star"}

	// Use simple randomization based on current time-like value
	// For true randomness, would need proper seeding
	prefix := prefixes[len(cc.nameInput)%len(prefixes)]
	suffix := suffixes[cc.selectedClass%CharacterClass(len(suffixes))]

	return prefix + suffix
}

// handleNextButton processes the Next touch button press
// Advances to the next step in character creation
func (cc *EbitenCharacterCreation) handleNextButton() {
	switch cc.currentStep {
	case stepNameInput:
		// Validate name and proceed to class selection
		if len(strings.TrimSpace(cc.nameInput)) > 0 {
			cc.characterData.Name = cc.nameInput
			cc.currentStep = stepClassSelection
			cc.stepChangedThisFrame = true // Mark that we changed steps this frame
			cc.errorMsg = ""
			cc.hideKeyboardIfNeeded()
		} else {
			cc.errorMsg = "Name cannot be empty"
		}
	case stepClassSelection:
		// Proceed to portrait selection
		cc.characterData.Class = cc.selectedClass
		cc.currentStep = stepPortraitSelection
		cc.stepChangedThisFrame = true // Mark that we changed steps this frame
		cc.errorMsg = ""
	case stepPortraitSelection:
		// Proceed to confirmation (same as Enter key)
		if cc.portraitInput != "" {
			cc.characterData.PortraitPath = strings.TrimSpace(cc.portraitInput)
			// WASM FIX: Store path for lazy loading in Draw()
			cc.pendingPortraitPath = cc.characterData.PortraitPath
			cc.portraitLoadAttempted = false
		}
		cc.currentStep = stepConfirmation
		cc.stepChangedThisFrame = true // Mark that we changed steps this frame
		cc.errorMsg = ""
		cc.hideKeyboardIfNeeded()
	case stepConfirmation:
		// Confirm character creation
		if err := cc.characterData.Validate(); err != nil {
			cc.errorMsg = err.Error()
			cc.currentStep = stepNameInput
			cc.keyboardShown = false
		} else {
			cc.confirmed = true
		}
	}
}

// handleBackButton processes the Back touch button press
// Returns to the previous step in character creation
func (cc *EbitenCharacterCreation) handleBackButton() {
	switch cc.currentStep {
	case stepNameInput:
		// No back navigation from first step - do nothing
		return
	case stepClassSelection:
		cc.currentStep = stepNameInput
		cc.keyboardShown = false // Will trigger keyboard on re-entry
	case stepPortraitSelection:
		cc.currentStep = stepClassSelection
		cc.hideKeyboardIfNeeded()
	case stepConfirmation:
		cc.currentStep = stepPortraitSelection
		cc.keyboardShown = false
	}
}

// handleSkipButton processes the Skip touch button press
// Skips optional portrait selection step
func (cc *EbitenCharacterCreation) handleSkipButton() {
	if cc.currentStep == stepPortraitSelection {
		cc.skipPortrait()
	}
}

// updateConfirmation handles final confirmation
func (cc *EbitenCharacterCreation) updateConfirmation() {
	// Handle mouse and touch input (Touch support for WASM/mobile)
	if IsTouchOrMouseJustPressed() {
		mouseX, mouseY, _ := GetTouchOrMousePosition()

		// Define button areas (matching drawConfirmation layout)
		buttonX := cc.panelX + 50
		buttonW := cc.panelWidth - 100
		buttonH := 30

		// Confirm button area
		confirmButtonY := cc.panelY + cc.panelHeight - 85

		// Back button area
		backButtonY := confirmButtonY + 40

		// Check confirm button click
		if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
			mouseY >= confirmButtonY && mouseY <= confirmButtonY+buttonH {
			// Validate before confirming
			if err := cc.characterData.Validate(); err != nil {
				cc.errorMsg = err.Error()
				cc.currentStep = stepNameInput // Go back to fix
				// Reset keyboard shown flag so it will be shown again when entering name input
				cc.keyboardShown = false
			} else {
				cc.confirmed = true
			}
			return
		}

		// Check back button click
		if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
			mouseY >= backButtonY && mouseY <= backButtonY+buttonH {
			cc.currentStep = stepPortraitSelection
			// Reset keyboard shown flag so it will be shown again when entering portrait path
			cc.keyboardShown = false
			return
		}
	}

	// Enter/Space to confirm
	if (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) && !cc.stepChangedThisFrame {
		// Validate before confirming
		if err := cc.characterData.Validate(); err != nil {
			cc.errorMsg = err.Error()
			cc.currentStep = stepNameInput // Go back to fix
			// Reset keyboard shown flag so it will be shown again when entering name input
			cc.keyboardShown = false
		} else {
			cc.confirmed = true
		}
	}

	// Backspace to go back to portrait selection
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepPortraitSelection
		// Reset keyboard shown flag so it will be shown again when entering portrait path
		cc.keyboardShown = false
	}
}

// Draw renders the character creation UI
func (cc *EbitenCharacterCreation) Draw(screen *ebiten.Image) {
	// WASM FIX: Lazy load pending portrait if graphics context is now ready
	// This ensures ebiten.NewImageFromImage() is only called during Draw cycle
	if cc.pendingPortraitPath != "" && !cc.portraitLoadAttempted {
		cc.portraitLoadAttempted = true
		portrait, err := LoadPortrait(cc.pendingPortraitPath)
		if err != nil {
			// Portrait load failed - clear pending path and show error
			cc.pendingPortraitPath = ""
			cc.errorMsg = fmt.Sprintf("Failed to load portrait: %v", err)
			cc.characterData.Portrait = nil
		} else {
			// Portrait loaded successfully
			cc.characterData.Portrait = portrait
			cc.pendingPortraitPath = "" // Clear pending path
			cc.errorMsg = ""            // Clear any previous errors
		}
	}

	// Draw semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(cc.screenWidth), float32(cc.screenHeight),
		color.RGBA{0, 0, 0, 200}, false)

	// Ensure panel dimensions are calculated (Update already did this, but Draw might be called standalone)
	cc.updatePanelDimensions()

	// Draw panel background
	vector.DrawFilledRect(screen, float32(cc.panelX), float32(cc.panelY),
		float32(cc.panelWidth), float32(cc.panelHeight),
		color.RGBA{20, 20, 30, 255}, false)

	// Draw panel border
	vector.StrokeRect(screen, float32(cc.panelX), float32(cc.panelY),
		float32(cc.panelWidth), float32(cc.panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)

	// Draw content based on current step
	switch cc.currentStep {
	case stepNameInput:
		cc.drawNameInput(screen, cc.panelX, cc.panelY, cc.panelWidth, cc.panelHeight)
	case stepClassSelection:
		cc.drawClassSelection(screen, cc.panelX, cc.panelY, cc.panelWidth, cc.panelHeight)
	case stepPortraitSelection:
		cc.drawPortraitSelection(screen, cc.panelX, cc.panelY, cc.panelWidth, cc.panelHeight)
	case stepConfirmation:
		cc.drawConfirmation(screen, cc.panelX, cc.panelY, cc.panelWidth, cc.panelHeight)
	}

	// Draw touch buttons (WASM/mobile support)
	cc.drawTouchButtons(screen)

	// Draw error message if present
	if cc.errorMsg != "" {
		errorX := cc.panelX + cc.panelWidth/2 - len(cc.errorMsg)*3
		errorY := cc.panelY + cc.panelHeight - 30
		text.Draw(screen, cc.errorMsg, basicfont.Face7x13, errorX, errorY,
			color.RGBA{255, 100, 100, 255})
	}
}

// drawTouchButtons renders the appropriate touch buttons for current step
func (cc *EbitenCharacterCreation) drawTouchButtons(screen *ebiten.Image) {
	// Ensure Next button has correct label for current step
	// Set label RIGHT before drawing to prevent it from being cleared
	if cc.nextButton != nil {
		if cc.currentStep == stepConfirmation {
			cc.nextButton.Label = "Done"
		} else {
			cc.nextButton.Label = "Next"
		}
		// Immediately draw after setting label
		cc.nextButton.Draw(screen)
	}

	// Draw other buttons based on step
	switch cc.currentStep {
	case stepNameInput:
		// Draw preset name buttons (WASM/mobile fallback)
		for _, btn := range cc.presetNameButtons {
			if btn != nil {
				btn.Draw(screen)
			}
		}
	case stepClassSelection:
		if cc.backButton != nil {
			cc.backButton.Draw(screen)
		}
	case stepPortraitSelection:
		if cc.skipButton != nil {
			cc.skipButton.Draw(screen)
		}
		if cc.backButton != nil {
			cc.backButton.Draw(screen)
		}
	case stepConfirmation:
		if cc.backButton != nil {
			cc.backButton.Draw(screen)
		}
	}
}

// drawNameInput renders the name input screen
func (cc *EbitenCharacterCreation) drawNameInput(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 1 of 4: Choose Your Name"
	stepX := x + w/2 - len(stepText)*3
	text.Draw(screen, stepText, basicfont.Face7x13, stepX, y+70,
		color.RGBA{200, 200, 200, 255})

	// Instruction
	instruction := "Enter your character's name:"
	instrX := x + w/2 - len(instruction)*3
	text.Draw(screen, instruction, basicfont.Face7x13, instrX, y+120,
		color.RGBA{150, 150, 150, 255})

	// Name input box
	inputBoxY := y + 150
	inputBoxX := x + w/2 - 150
	vector.DrawFilledRect(screen, float32(inputBoxX), float32(inputBoxY), 300, 30,
		color.RGBA{40, 40, 50, 255}, false)
	vector.StrokeRect(screen, float32(inputBoxX), float32(inputBoxY), 300, 30, 1,
		color.RGBA{150, 150, 200, 255}, false)

	// Display current input with cursor
	displayText := cc.nameInput + "_"
	textX := inputBoxX + 10
	text.Draw(screen, displayText, basicfont.Face7x13, textX, inputBoxY+20,
		color.RGBA{255, 255, 255, 255})

	// Preset name buttons hint (for WASM/mobile)
	if mobile.IsWASM() {
		hintText := "Or tap a preset name below:"
		hintX := x + w/2 - len(hintText)*3
		text.Draw(screen, hintText, basicfont.Face7x13, hintX, y+185,
			color.RGBA{100, 200, 100, 255})
	}

	// Show current default if set
	if cc.defaults.DefaultName != "" {
		defaultText := fmt.Sprintf("Current default: %s", cc.defaults.DefaultName)
		defaultX := x + w/2 - len(defaultText)*3
		text.Draw(screen, defaultText, basicfont.Face7x13, defaultX, y+245,
			color.RGBA{150, 150, 150, 255})
	}

	// Help text
	helpText1 := "Press ENTER or click NEXT to continue"
	helpText2 := "F2 to save as default"
	helpX1 := x + w/2 - len(helpText1)*3
	helpX2 := x + w/2 - len(helpText2)*3
	text.Draw(screen, helpText1, basicfont.Face7x13, helpX1, y+h-75,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText2, basicfont.Face7x13, helpX2, y+h-60,
		color.RGBA{150, 200, 150, 255})
}

// drawClassSelection renders the class selection screen
func (cc *EbitenCharacterCreation) drawClassSelection(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 2 of 4: Choose Your Class"
	stepX := x + w/2 - len(stepText)*3
	text.Draw(screen, stepText, basicfont.Face7x13, stepX, y+70,
		color.RGBA{200, 200, 200, 255})

	// Display name
	nameText := fmt.Sprintf("Name: %s", cc.characterData.Name)
	nameX := x + 30
	text.Draw(screen, nameText, basicfont.Face7x13, nameX, y+100,
		color.RGBA{200, 200, 255, 255})

	// Class options
	classes := []CharacterClass{ClassWarrior, ClassMage, ClassRogue}
	startY := y + 140

	for i, class := range classes {
		classY := startY + i*80
		isSelected := class == cc.selectedClass

		// Selection indicator
		if isSelected {
			vector.DrawFilledRect(screen, float32(x+40), float32(classY-5), float32(w-80), 70,
				color.RGBA{50, 80, 120, 255}, false)
		}

		// Class name
		classColor := color.RGBA{200, 200, 200, 255}
		if isSelected {
			classColor = color.RGBA{255, 255, 100, 255}
		}

		className := fmt.Sprintf("%d. %s", i+1, class.String())
		text.Draw(screen, className, basicfont.Face7x13, x+50, classY+15, classColor)

		// Class description (wrapped)
		desc := class.Description()
		descLines := wrapText(desc, 60)
		for j, line := range descLines {
			text.Draw(screen, line, basicfont.Face7x13, x+70, classY+35+j*15,
				color.RGBA{180, 180, 180, 255})
		}
	}

	// Show current default if set (defaults to ClassWarrior as zero value)
	// Only show if explicitly set, which we track by checking if DefaultName is also set
	if cc.defaults.DefaultName != "" {
		defaultText := fmt.Sprintf("Current default: %s", cc.defaults.DefaultClass.String())
		defaultX := x + w/2 - len(defaultText)*3
		text.Draw(screen, defaultText, basicfont.Face7x13, defaultX, y+h-110,
			color.RGBA{150, 150, 150, 255})
	}

	// Help text
	helpText1 := "Use ARROW KEYS or 1-3 to select"
	helpText2 := "TAP/CLICK a class to select and continue"
	helpText3 := "Press ENTER or click NEXT to continue"
	helpText4 := "BACKSPACE or click BACK to go back | F2 to save default"
	helpX1 := x + w/2 - len(helpText1)*3
	helpX2 := x + w/2 - len(helpText2)*3
	helpX3 := x + w/2 - len(helpText3)*3
	helpX4 := x + w/2 - len(helpText4)*3
	text.Draw(screen, helpText1, basicfont.Face7x13, helpX1, y+h-105,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText2, basicfont.Face7x13, helpX2, y+h-85,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText3, basicfont.Face7x13, helpX3, y+h-65,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText4, basicfont.Face7x13, helpX4, y+h-45,
		color.RGBA{150, 200, 150, 255})
}

// drawPortraitSelection renders the portrait selection screen
func (cc *EbitenCharacterCreation) drawPortraitSelection(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 3 of 4: Choose Portrait (Optional)"
	stepX := x + w/2 - len(stepText)*3
	text.Draw(screen, stepText, basicfont.Face7x13, stepX, y+70,
		color.RGBA{200, 200, 200, 255})

	// Instructions
	instructionY := y + 110
	instructions := []string{
		"Press SPACE or B to browse for a .png file",
		"Or type path manually (max 512x512)",
		"Press TAB to skip (optional)",
	}
	for i, line := range instructions {
		lineX := x + w/2 - len(line)*3
		text.Draw(screen, line, basicfont.Face7x13, lineX, instructionY+i*20,
			color.RGBA{180, 180, 180, 255})
	}

	// Input box for file path
	inputBoxY := y + 170
	inputBoxX := x + 50
	inputBoxWidth := w - 100
	inputBoxHeight := 30
	vector.DrawFilledRect(screen, float32(inputBoxX), float32(inputBoxY),
		float32(inputBoxWidth), float32(inputBoxHeight),
		color.RGBA{40, 40, 50, 255}, false)
	vector.StrokeRect(screen, float32(inputBoxX), float32(inputBoxY),
		float32(inputBoxWidth), float32(inputBoxHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)

	// Display current input with cursor
	displayText := cc.portraitInput
	if len(displayText) > 60 {
		// Truncate display to fit
		displayText = "..." + displayText[len(displayText)-57:]
	}
	displayText += "_"
	textX := inputBoxX + 10
	text.Draw(screen, displayText, basicfont.Face7x13, textX, inputBoxY+20,
		color.RGBA{255, 255, 255, 255})

	// Show current default if set
	if cc.defaults.DefaultPortraitPath != "" {
		defaultText := fmt.Sprintf("Current default: %s", cc.defaults.DefaultPortraitPath)
		if len(defaultText) > 70 {
			defaultText = defaultText[:67] + "..."
		}
		defaultX := x + w/2 - len(defaultText)*3
		text.Draw(screen, defaultText, basicfont.Face7x13, defaultX, y+220,
			color.RGBA{150, 150, 150, 255})
	}

	// Portrait preview (if loaded)
	if cc.characterData.Portrait != nil {
		previewY := y + 250
		previewX := x + w/2 - cc.characterData.Portrait.Bounds().Dx()/2
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(previewX), float64(previewY))
		screen.DrawImage(cc.characterData.Portrait, opts)

		// Label
		labelText := "Preview:"
		labelX := x + w/2 - len(labelText)*3
		text.Draw(screen, labelText, basicfont.Face7x13, labelX, previewY-10,
			color.RGBA{200, 200, 100, 255})
	}

	// Help text
	helpY := y + h - 100

	// Draw clickable button areas for touch support
	buttonY := helpY - 10
	buttonX := x + 50
	buttonW := w - 100
	buttonH := 25

	// Browse button
	browseButtonY := buttonY
	vector.DrawFilledRect(screen, float32(buttonX), float32(browseButtonY),
		float32(buttonW), float32(buttonH),
		color.RGBA{50, 80, 120, 255}, false)
	vector.StrokeRect(screen, float32(buttonX), float32(browseButtonY),
		float32(buttonW), float32(buttonH), 2,
		color.RGBA{100, 150, 200, 255}, false)
	browseText := "Browse for Portrait (SPACE/B)"
	browseTextX := buttonX + buttonW/2 - len(browseText)*3
	text.Draw(screen, browseText, basicfont.Face7x13, browseTextX, browseButtonY+17,
		color.RGBA{255, 255, 255, 255})

	// Skip button
	skipButtonY := browseButtonY + 35
	vector.DrawFilledRect(screen, float32(buttonX), float32(skipButtonY),
		float32(buttonW), float32(buttonH),
		color.RGBA{50, 80, 120, 255}, false)
	vector.StrokeRect(screen, float32(buttonX), float32(skipButtonY),
		float32(buttonW), float32(buttonH), 2,
		color.RGBA{100, 150, 200, 255}, false)
	skipText := "Skip Portrait (TAB)"
	skipTextX := buttonX + buttonW/2 - len(skipText)*3
	text.Draw(screen, skipText, basicfont.Face7x13, skipTextX, skipButtonY+17,
		color.RGBA{255, 255, 255, 255})

	// Back button
	backButtonY := skipButtonY + 35
	vector.DrawFilledRect(screen, float32(buttonX), float32(backButtonY),
		float32(buttonW), float32(buttonH),
		color.RGBA{80, 50, 50, 255}, false)
	vector.StrokeRect(screen, float32(buttonX), float32(backButtonY),
		float32(buttonW), float32(buttonH), 2,
		color.RGBA{150, 100, 100, 255}, false)
	backText := "Back to Class Selection (BACKSPACE)"
	backTextX := buttonX + buttonW/2 - len(backText)*3
	text.Draw(screen, backText, basicfont.Face7x13, backTextX, backButtonY+17,
		color.RGBA{255, 255, 255, 255})
}

// drawConfirmation renders the confirmation screen
func (cc *EbitenCharacterCreation) drawConfirmation(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 4 of 4: Confirm Your Character"
	stepX := x + w/2 - len(stepText)*3
	text.Draw(screen, stepText, basicfont.Face7x13, stepX, y+70,
		color.RGBA{200, 200, 200, 255})

	// Character summary
	summaryY := y + 130

	nameText := fmt.Sprintf("Name: %s", cc.characterData.Name)
	text.Draw(screen, nameText, basicfont.Face7x13, x+w/2-len(nameText)*3, summaryY,
		color.RGBA{255, 255, 255, 255})

	classText := fmt.Sprintf("Class: %s", cc.characterData.Class.String())
	text.Draw(screen, classText, basicfont.Face7x13, x+w/2-len(classText)*3, summaryY+30,
		color.RGBA{255, 255, 255, 255})

	// Portrait preview (if set)
	if cc.characterData.Portrait != nil {
		portraitText := fmt.Sprintf("Portrait: Custom (%dx%d)",
			cc.characterData.Portrait.Bounds().Dx(),
			cc.characterData.Portrait.Bounds().Dy())
		text.Draw(screen, portraitText, basicfont.Face7x13, x+w/2-len(portraitText)*3, summaryY+60,
			color.RGBA{255, 255, 255, 255})

		// Show small preview
		previewSize := 64
		previewX := x + w/2 - previewSize/2
		previewY := summaryY + 80

		opts := &ebiten.DrawImageOptions{}
		// Scale down to 64x64 preview
		scaleX := float64(previewSize) / float64(cc.characterData.Portrait.Bounds().Dx())
		scaleY := float64(previewSize) / float64(cc.characterData.Portrait.Bounds().Dy())
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(float64(previewX), float64(previewY))
		screen.DrawImage(cc.characterData.Portrait, opts)
	} else {
		portraitText := "Portrait: None"
		text.Draw(screen, portraitText, basicfont.Face7x13, x+w/2-len(portraitText)*3, summaryY+60,
			color.RGBA{180, 180, 180, 255})
	}

	// Class stats preview
	statsY := summaryY + 180
	statsTitle := "Starting Stats:"
	text.Draw(screen, statsTitle, basicfont.Face7x13, x+w/2-len(statsTitle)*3, statsY,
		color.RGBA{200, 200, 100, 255})

	stats := cc.getClassStats()
	statY := statsY + 30
	for _, line := range stats {
		text.Draw(screen, line, basicfont.Face7x13, x+w/2-len(line)*3, statY,
			color.RGBA{180, 180, 180, 255})
		statY += 20
	}

	// Draw clickable buttons for touch support
	buttonX := x + 50
	buttonW := w - 100
	buttonH := 30

	// Confirm button (green/positive action)
	confirmButtonY := y + h - 85
	vector.DrawFilledRect(screen, float32(buttonX), float32(confirmButtonY),
		float32(buttonW), float32(buttonH),
		color.RGBA{50, 120, 50, 255}, false)
	vector.StrokeRect(screen, float32(buttonX), float32(confirmButtonY),
		float32(buttonW), float32(buttonH), 2,
		color.RGBA{100, 200, 100, 255}, false)
	confirmText := "BEGIN ADVENTURE (ENTER)"
	confirmTextX := buttonX + buttonW/2 - len(confirmText)*3
	text.Draw(screen, confirmText, basicfont.Face7x13, confirmTextX, confirmButtonY+20,
		color.RGBA{255, 255, 255, 255})

	// Back button
	backButtonY := confirmButtonY + 40
	vector.DrawFilledRect(screen, float32(buttonX), float32(backButtonY),
		float32(buttonW), float32(buttonH),
		color.RGBA{80, 50, 50, 255}, false)
	vector.StrokeRect(screen, float32(buttonX), float32(backButtonY),
		float32(buttonW), float32(buttonH), 2,
		color.RGBA{150, 100, 100, 255}, false)
	backText := "Go Back (BACKSPACE)"
	backTextX := buttonX + buttonW/2 - len(backText)*3
	text.Draw(screen, backText, basicfont.Face7x13, backTextX, backButtonY+20,
		color.RGBA{255, 255, 255, 255})
}

// getClassStats returns stat descriptions for the selected class
func (cc *EbitenCharacterCreation) getClassStats() []string {
	switch cc.characterData.Class {
	case ClassWarrior:
		return []string{
			"Health: 150 (High)",
			"Mana: 50 (Low)",
			"Attack: 12 (High)",
			"Defense: 8 (High)",
		}
	case ClassMage:
		return []string{
			"Health: 80 (Low)",
			"Mana: 150 (High)",
			"Attack: 6 (Low)",
			"Defense: 3 (Low)",
		}
	case ClassRogue:
		return []string{
			"Health: 100 (Medium)",
			"Mana: 80 (Medium)",
			"Attack: 10 (Medium)",
			"Defense: 5 (Medium)",
		}
	default:
		return []string{}
	}
}

// GetCharacterData returns the completed character data
func (cc *EbitenCharacterCreation) GetCharacterData() CharacterData {
	return cc.characterData
}

// IsComplete returns whether character creation is finished
func (cc *EbitenCharacterCreation) IsComplete() bool {
	return cc.confirmed
}

// Cleanup hides the mobile keyboard and performs any necessary cleanup.
// MOBILE/WASM FIX: This should be called when character creation completes
// and the game transitions to gameplay, ensuring the keyboard is dismissed.
func (cc *EbitenCharacterCreation) Cleanup() {
	// MOBILE/WASM: Hide keyboard when character creation is complete
	if cc.keyboardShown && mobile.IsWASM() {
		mobile.HideKeyboard()
	}
	// Always reset the flag regardless of platform for consistent state management
	cc.keyboardShown = false
}

// Reset resets the character creation to initial state
// Applies custom defaults if they are set
func (cc *EbitenCharacterCreation) Reset() {
	cc.currentStep = stepNameInput
	cc.characterData = CharacterData{}
	cc.confirmed = false
	cc.errorMsg = ""

	// MOBILE/WASM KEYBOARD FIX: Reset keyboard state flag to false.
	// The keyboard will be shown automatically on the first Update() call
	// when updateNameInput() detects keyboardShown=false and shows it.
	// This prevents premature keyboard display before the UI is ready.
	if mobile.IsWASM() {
		// Hide keyboard if it was shown from previous state
		if cc.keyboardShown {
			mobile.HideKeyboard()
		}
	}
	// Always reset flag regardless of platform for consistent state management
	cc.keyboardShown = false

	// Apply defaults to both input fields and character data
	if cc.defaults.DefaultName != "" {
		cc.nameInput = cc.defaults.DefaultName
		cc.characterData.Name = cc.defaults.DefaultName
	} else {
		cc.nameInput = ""
	}
	cc.selectedClass = cc.defaults.DefaultClass
	cc.characterData.Class = cc.defaults.DefaultClass

	// Apply portrait default
	// WASM FIX: Defer portrait loading until Draw() to ensure graphics context is ready
	if cc.defaults.DefaultPortraitPath != "" {
		cc.portraitInput = cc.defaults.DefaultPortraitPath
		cc.characterData.PortraitPath = cc.defaults.DefaultPortraitPath
		cc.pendingPortraitPath = cc.defaults.DefaultPortraitPath
		cc.portraitLoadAttempted = false
		cc.characterData.Portrait = nil // Will be loaded lazily in Draw()
	} else {
		cc.portraitInput = ""
		cc.pendingPortraitPath = ""
		cc.portraitLoadAttempted = false
		cc.characterData.Portrait = nil
	}
}

// SaveAsDefaults saves the current character data as defaults for future use
func wrapText(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	lines := []string{}
	currentLine := words[0]

	for i := 1; i < len(words); i++ {
		if len(currentLine)+1+len(words[i]) <= maxChars {
			currentLine += " " + words[i]
		} else {
			lines = append(lines, currentLine)
			currentLine = words[i]
		}
	}
	lines = append(lines, currentLine)

	return lines
}

// ApplyClassStats applies class-based stats to a player entity
// This should be called after entity creation but before gameplay starts
// classComponents holds the player's core stat components for class configuration
type classComponents struct {
	health *HealthComponent
	mana   *ManaComponent
	stats  *StatsComponent
	attack *AttackComponent
}

// extractPlayerComponents retrieves and validates all required components from a player entity
func extractPlayerComponents(player *Entity) (*classComponents, error) {
	if player == nil {
		return nil, fmt.Errorf("player entity is nil")
	}

	healthComp, hasHealth := player.GetComponent("health")
	if !hasHealth {
		return nil, fmt.Errorf("player missing health component")
	}

	manaComp, hasMana := player.GetComponent("mana")
	if !hasMana {
		return nil, fmt.Errorf("player missing mana component")
	}

	statsComp, hasStats := player.GetComponent("stats")
	if !hasStats {
		return nil, fmt.Errorf("player missing stats component")
	}

	attackComp, hasAttack := player.GetComponent("attack")
	if !hasAttack {
		return nil, fmt.Errorf("player missing attack component")
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return nil, fmt.Errorf("health component has wrong type")
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return nil, fmt.Errorf("mana component has wrong type")
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return nil, fmt.Errorf("stats component has wrong type")
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return nil, fmt.Errorf("attack component has wrong type")
	}

	return &classComponents{
		health: health,
		mana:   mana,
		stats:  stats,
		attack: attack,
	}, nil
}

// applyWarriorStats configures a player entity with warrior class statistics
func applyWarriorStats(comps *classComponents) {
	comps.health.Max = 150
	comps.health.Current = 150
	comps.mana.Max = 50
	comps.mana.Current = 50
	comps.stats.Attack = 12
	comps.stats.Defense = 8
	comps.attack.Damage = 20
	comps.stats.CritChance = 0.05
	comps.stats.CritDamage = 2.0
}

// applyMageStats configures a player entity with mage class statistics
func applyMageStats(comps *classComponents) {
	comps.health.Max = 80
	comps.health.Current = 80
	comps.mana.Max = 150
	comps.mana.Current = 150
	comps.mana.Regen = 8.0
	comps.stats.Attack = 6
	comps.stats.Defense = 3
	comps.attack.Damage = 10
	comps.stats.CritChance = 0.10
	comps.stats.CritDamage = 1.8
}

// applyRogueStats configures a player entity with rogue class statistics
func applyRogueStats(comps *classComponents) {
	comps.health.Max = 100
	comps.health.Current = 100
	comps.mana.Max = 80
	comps.mana.Current = 80
	comps.stats.Attack = 10
	comps.stats.Defense = 5
	comps.attack.Damage = 15
	comps.attack.Cooldown = 0.3
	comps.stats.CritChance = 0.15
	comps.stats.CritDamage = 2.5
	comps.stats.Evasion = 0.15
}

func ApplyClassStats(player *Entity, class CharacterClass) error {
	comps, err := extractPlayerComponents(player)
	if err != nil {
		return err
	}

	switch class {
	case ClassWarrior:
		applyWarriorStats(comps)
	case ClassMage:
		applyMageStats(comps)
	case ClassRogue:
		applyRogueStats(comps)
	default:
		return fmt.Errorf("unknown character class: %v", class)
	}

	return nil
}
