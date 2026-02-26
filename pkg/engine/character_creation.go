// Package engine provides character creation functionality for onboarding new players.
// This file implements the character creation UI and class selection system that
// integrates with the tutorial flow for a unified onboarding experience.
package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/config"
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

// CharacterClass represents a player archetype with specific stat distributions.
// This is a type alias to config.CharacterClass to avoid circular dependencies.
type CharacterClass = config.CharacterClass

// Class constants are re-exported from config package for convenience
const (
	ClassWarrior      = config.ClassWarrior
	ClassMage         = config.ClassMage
	ClassRogue        = config.ClassRogue
	ClassRanger       = config.ClassRanger
	ClassCleric       = config.ClassCleric
	ClassNecromancer  = config.ClassNecromancer
	ClassBattlemage   = config.ClassBattlemage
	ClassSpellblade   = config.ClassSpellblade
	ClassPaladin      = config.ClassPaladin
	ClassMonk         = config.ClassMonk
	ClassDeathKnight  = config.ClassDeathKnight
	ClassWitchHunter  = config.ClassWitchHunter
	ClassBeastlord    = config.ClassBeastlord
	ClassArcaneArcher = config.ClassArcaneArcher
	ClassShadowPriest = config.ClassShadowPriest
	ClassDruid        = config.ClassDruid
	ClassInquisitor   = config.ClassInquisitor
	ClassBloodKnight  = config.ClassBloodKnight
	ClassMystic       = config.ClassMystic
	ClassWarlock      = config.ClassWarlock
	ClassNinja        = config.ClassNinja
)

// CharacterData holds the player's character creation choices
type CharacterData struct {
	Name            string
	Class           CharacterClass
	PortraitPath    string            // Path to user's custom portrait image (optional)
	Portrait        *ebiten.Image     // Loaded portrait image (optional, max 512x512)
	StartingLoadout *EquipmentLoadout // Selected starting equipment loadout
}

// EquipmentLoadout represents a starting equipment set for a character class.
// Each class has 3 distinct loadout options generated deterministically from the world seed.
type EquipmentLoadout struct {
	Name        string // Display name (e.g., "Heavy Armor", "Balanced", "Berserker")
	Description string // Brief description of the loadout style
	MainHand    string // Main weapon name
	OffHand     string // Off-hand item name (shield, second weapon, etc.)
	Armor       string // Primary armor piece name
	Accessory   string // Starting accessory name
	// Stats modifiers for this loadout (applied on top of class base stats)
	BonusHP      int
	BonusAttack  int
	BonusDefense int
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
	if cd.Class < ClassWarrior || cd.Class > ClassNinja {
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
	stepSubclassSelection
	stepPortraitSelection
	stepConfirmation
)

// noSubclassSelected is a sentinel value indicating no subclass has been chosen
const noSubclassSelected CharacterClass = -1

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

	// Deterministic name generation
	worldSeed      int64 // World seed for deterministic name generation
	nameGenCounter int   // Counter to allow multiple unique name generations per seed

	// Class pagination - allows viewing hybrid classes via PageUp/PageDown or Tab
	classPage int // 0 = base classes (1-6), 1+ = hybrid class pages

	// selectedSubclass is the chosen hybrid class (-1 = none, use base class)
	selectedSubclass CharacterClass
}

// NewCharacterCreation creates a new character creation system
func NewCharacterCreation(screenWidth, screenHeight int) *EbitenCharacterCreation {
	cc := &EbitenCharacterCreation{
		currentStep:      stepNameInput,
		selectedClass:    ClassWarrior, // Default selection
		selectedSubclass: noSubclassSelected,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
		inputBuffer:      make([]rune, 0),
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

	// Preset name buttons are initialized lazily when the world seed is
	// set via SetDefaultNameFromSeed, since we need the seed to generate
	// deterministic random names. Start with a placeholder set.
	cc.presetNameButtons = nil

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
// Also generates preset name buttons with seed-derived random names.
func (cc *EbitenCharacterCreation) SetDefaultNameFromSeed(seed int64) {
	cc.worldSeed = seed // Store seed for generateRandomName
	defaultName := procgen.SelectDefaultName(seed)
	cc.defaults.DefaultName = defaultName
	// Apply to current state if we're in name input step
	if cc.currentStep == stepNameInput {
		cc.nameInput = defaultName
	}

	// Generate preset name buttons with deterministic random names.
	// Uses the world seed offset by index to produce varied but reproducible names.
	cc.generatePresetNameButtons(seed)
}

// generatePresetNameButtons creates touch buttons with seed-derived random names
// plus an "Auto" button that generates a fresh random name each press.
func (cc *EbitenCharacterCreation) generatePresetNameButtons(seed int64) {
	const presetCount = 5
	rng := rand.New(rand.NewSource(seed + 9999)) // Offset to differ from default name

	// Pick presetCount-1 unique names, plus an "Auto" button
	used := make(map[int]bool)
	presetNames := make([]string, 0, presetCount-1)
	for len(presetNames) < presetCount-1 {
		idx := rng.Intn(len(procgen.DefaultNames))
		if used[idx] {
			continue
		}
		used[idx] = true
		presetNames = append(presetNames, procgen.DefaultNames[idx])
	}

	cc.presetNameButtons = make([]*mobile.TouchButton, presetCount)
	for i, name := range presetNames {
		capturedName := name
		cc.presetNameButtons[i] = mobile.NewTouchButton(
			0, 0,
			100, 36,
			capturedName,
			func() { cc.handlePresetName(capturedName) },
		)
	}
	// Last button is "Random" to generate a fresh name
	cc.presetNameButtons[presetCount-1] = mobile.NewTouchButton(
		0, 0,
		100, 36,
		"Random",
		func() { cc.handlePresetName("Auto") },
	)
}

// GetDefaults returns the current default values
func (cc *EbitenCharacterCreation) GetDefaults() CharacterCreationDefaults {
	return cc.defaults
}

// Update handles input for character creation (keyboard/mouse navigation)
// Returns true when character creation is complete
func (cc *EbitenCharacterCreation) Update() bool {
	cc.stepChangedThisFrame = false
	cc.updatePanelDimensions()

	cc.updateTouchControls()
	cc.processCurrentStep()

	return cc.confirmed
}

// updateTouchControls updates all touch-based UI controls.
func (cc *EbitenCharacterCreation) updateTouchControls() {
	if cc.touchHandler != nil {
		cc.touchHandler.Update()
	}

	cc.updateTouchButtonPositions()

	if cc.nextButton != nil {
		cc.nextButton.Update()
	}

	if cc.backButton != nil && cc.currentStep != stepNameInput {
		cc.backButton.Update()
	}

	if cc.skipButton != nil && cc.currentStep == stepPortraitSelection {
		cc.skipButton.Update()
	}

	if cc.currentStep == stepNameInput {
		for _, btn := range cc.presetNameButtons {
			if btn != nil {
				btn.Update()
			}
		}
	}
}

// processCurrentStep processes the current character creation step.
func (cc *EbitenCharacterCreation) processCurrentStep() {
	switch cc.currentStep {
	case stepNameInput:
		cc.updateNameInput()
	case stepClassSelection:
		cc.updateClassSelection()
	case stepSubclassSelection:
		cc.updateSubclassSelection()
	case stepPortraitSelection:
		cc.updatePortraitSelection()
	case stepConfirmation:
		cc.updateConfirmation()
	}
}

// updateTouchButtonPositions positions touch buttons based on panel layout.
// Touch buttons are placed at the very bottom of the panel, below all drawn
// content, so they never overlap with in-panel buttons or help text.
func (cc *EbitenCharacterCreation) updateTouchButtonPositions() {
	// All navigation buttons sit in a row at the panel's bottom edge.
	buttonRowY := cc.panelY + cc.panelHeight - 54 // 44px button + 10px bottom padding

	// Next button (bottom-right of panel)
	if cc.nextButton != nil {
		nextX := cc.panelX + cc.panelWidth - 140
		cc.nextButton.SetPosition(
			float64(nextX),
			float64(buttonRowY),
		)
	}

	// Back button (bottom-left of panel)
	if cc.backButton != nil {
		cc.backButton.SetPosition(
			float64(cc.panelX+20),
			float64(buttonRowY),
		)
	}

	// Skip button (bottom-center of panel)
	if cc.skipButton != nil {
		cc.skipButton.SetPosition(
			float64(cc.panelX+cc.panelWidth/2-60),
			float64(buttonRowY),
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
	cc.panelWidth = 660
	cc.panelHeight = 600

	// Clamp to screen size with margin so the panel is always fully visible
	maxH := cc.screenHeight - 40 // 20px margin top+bottom
	if cc.panelHeight > maxH {
		cc.panelHeight = maxH
	}
	maxW := cc.screenWidth - 40
	if cc.panelWidth > maxW {
		cc.panelWidth = maxW
	}

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

// baseClasses contains the 6 selectable base classes in character creation
var baseClasses = []CharacterClass{
	ClassWarrior, ClassMage, ClassRogue, ClassRanger, ClassCleric, ClassNecromancer,
}

// hybridClasses contains the 15 hybrid classes split into pages of 6
var hybridClassPages = [][]CharacterClass{
	// Page 1 (index 1): First 6 hybrids
	{ClassBattlemage, ClassSpellblade, ClassPaladin, ClassMonk, ClassDeathKnight, ClassWitchHunter},
	// Page 2 (index 2): Next 6 hybrids
	{ClassBeastlord, ClassArcaneArcher, ClassShadowPriest, ClassDruid, ClassInquisitor, ClassBloodKnight},
	// Page 3 (index 3): Final 3 hybrids (Mystic, Warlock, Ninja)
	{ClassMystic, ClassWarlock, ClassNinja},
}

// totalClassPages returns the total number of class pages (1 base + N hybrid pages)
func totalClassPages() int {
	return 1 + len(hybridClassPages)
}

// getClassesForPage returns the classes to display on the given page
func getClassesForPage(page int) []CharacterClass {
	if page == 0 {
		return baseClasses
	}
	if page > 0 && page <= len(hybridClassPages) {
		return hybridClassPages[page-1]
	}
	return nil
}

// getPageTitle returns a descriptive title for the given class page
func getPageTitle(page int) string {
	if page == 0 {
		return "Base Classes"
	}
	return "Advanced Classes"
}

// getSubclassesForBaseClass returns the hybrid class options available for the given base class.
// Each hybrid class appears as a subclass for each of its parent base classes.
func getSubclassesForBaseClass(base CharacterClass) []CharacterClass {
	switch base {
	case ClassWarrior:
		return []CharacterClass{ClassBattlemage, ClassPaladin, ClassDeathKnight, ClassBeastlord, ClassBloodKnight}
	case ClassMage:
		return []CharacterClass{ClassBattlemage, ClassSpellblade, ClassArcaneArcher, ClassDruid, ClassMystic, ClassWarlock}
	case ClassRogue:
		return []CharacterClass{ClassSpellblade, ClassMonk, ClassShadowPriest, ClassInquisitor, ClassNinja}
	case ClassRanger:
		return []CharacterClass{ClassWitchHunter, ClassBeastlord, ClassArcaneArcher, ClassDruid, ClassNinja}
	case ClassCleric:
		return []CharacterClass{ClassPaladin, ClassMonk, ClassWitchHunter, ClassShadowPriest, ClassInquisitor, ClassMystic}
	case ClassNecromancer:
		return []CharacterClass{ClassDeathKnight, ClassBloodKnight, ClassShadowPriest, ClassWarlock}
	default:
		return nil
	}
}

// handleArrowKeySelection processes arrow key navigation for class selection
func (cc *EbitenCharacterCreation) handleArrowKeySelection() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		idx := cc.selectedClassIndex()
		if idx <= 0 {
			cc.selectedClass = baseClasses[len(baseClasses)-1]
		} else {
			cc.selectedClass = baseClasses[idx-1]
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		idx := cc.selectedClassIndex()
		if idx < 0 || idx >= len(baseClasses)-1 {
			cc.selectedClass = baseClasses[0]
		} else {
			cc.selectedClass = baseClasses[idx+1]
		}
	}
}

// selectedClassIndex returns the index of selectedClass in baseClasses, or -1 if not found
func (cc *EbitenCharacterCreation) selectedClassIndex() int {
	for i, class := range baseClasses {
		if class == cc.selectedClass {
			return i
		}
	}
	return -1
}

// handleNumberKeySelection processes numeric key shortcuts for direct class selection
// Numbers 1-6 select the corresponding base class
func (cc *EbitenCharacterCreation) handleNumberKeySelection() {
	keys := []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6}
	for i, key := range keys {
		if inpututil.IsKeyJustPressed(key) && i < len(baseClasses) {
			cc.selectedClass = baseClasses[i]
			break
		}
	}
}

// handleTouchOrMouseClick processes touch and mouse click events for class selection.
// A click selects the class (highlights it) but does NOT advance to the next step.
// The player must press ENTER or click Next to confirm their choice.
func (cc *EbitenCharacterCreation) handleTouchOrMouseClick() bool {
	if !IsTouchOrMouseJustPressed() || cc.stepChangedThisFrame {
		return false
	}
	mouseX, mouseY, _ := GetTouchOrMousePosition()
	startY := cc.panelY + 160

	for i, class := range baseClasses {
		if cc.isClassBoxClicked(mouseX, mouseY, startY, i) {
			cc.selectedClass = class
			return false // Select only — don't advance step
		}
	}
	return false
}

// isClassBoxClicked checks if coordinates are within a class option box
func (cc *EbitenCharacterCreation) isClassBoxClicked(mouseX, mouseY, startY, classIndex int) bool {
	classY := startY + classIndex*55
	return mouseX >= cc.panelX+40 && mouseX <= cc.panelX+cc.panelWidth-40 &&
		mouseY >= classY-5 && mouseY <= classY+45
}

// handleTouchOrMouseHover is intentionally a no-op.
// Mouse hover no longer changes the selected class because it was causing
// unintentional class changes when the cursor rested over a different class box.
// Selection is done only via arrow keys, number keys, or explicit clicks.
func (cc *EbitenCharacterCreation) handleTouchOrMouseHover() {
	// No-op: hover selection removed to prevent accidental class changes.
}

// handleConfirmationKeys processes Enter key to confirm selection
func (cc *EbitenCharacterCreation) handleConfirmationKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && !cc.stepChangedThisFrame {
		cc.characterData.Class = cc.selectedClass
		cc.selectedSubclass = noSubclassSelected
		cc.currentStep = stepSubclassSelection
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

// updateSubclassSelection handles the subclass (hybrid class) selection step.
// Shows hybrid classes available for the chosen base class and allows "None" selection.
func (cc *EbitenCharacterCreation) updateSubclassSelection() {
	subclasses := getSubclassesForBaseClass(cc.selectedClass)

	// Build full option list: None + subclasses
	optionCount := len(subclasses) + 1 // +1 for "None"

	// Find current selection index (0 = None, 1+ = subclass index)
	currentIdx := 0
	if cc.selectedSubclass != noSubclassSelected {
		for i, sc := range subclasses {
			if sc == cc.selectedSubclass {
				currentIdx = i + 1
				break
			}
		}
	}

	// Arrow key navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		currentIdx--
		if currentIdx < 0 {
			currentIdx = optionCount - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		currentIdx++
		if currentIdx >= optionCount {
			currentIdx = 0
		}
	}

	// Number key selection
	keys := []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7}
	for i, key := range keys {
		if inpututil.IsKeyJustPressed(key) && i < optionCount {
			currentIdx = i
			break
		}
	}

	// Update selectedSubclass from currentIdx
	if currentIdx == 0 {
		cc.selectedSubclass = noSubclassSelected
	} else if currentIdx-1 < len(subclasses) {
		cc.selectedSubclass = subclasses[currentIdx-1]
	}

	// Mouse/touch click selection — select only, don't auto-advance.
	// The player must press ENTER or click Next to confirm.
	if IsTouchOrMouseJustPressed() && !cc.stepChangedThisFrame {
		mouseX, mouseY, _ := GetTouchOrMousePosition()
		startY := cc.panelY + 130
		for i := 0; i < optionCount; i++ {
			optionY := startY + i*55
			if mouseX >= cc.panelX+40 && mouseX <= cc.panelX+cc.panelWidth-40 &&
				mouseY >= optionY-5 && mouseY <= optionY+45 {
				if i == 0 {
					cc.selectedSubclass = noSubclassSelected
				} else if i-1 < len(subclasses) {
					cc.selectedSubclass = subclasses[i-1]
				}
				return
			}
		}
	}

	// Enter to proceed
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && !cc.stepChangedThisFrame {
		cc.handleNextButton()
	}

	// Backspace/Escape to go back
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepClassSelection
	}
}

// generateClassLoadouts creates 3 deterministic equipment loadouts for a class.
// Uses seed-based RNG for reproducibility: same seed always produces same loadouts.
func generateClassLoadouts(class CharacterClass, seed int64) []EquipmentLoadout {
	// Combine seed with class for unique loadouts per class
	rng := rand.New(rand.NewSource(seed + int64(class)*1000))

	// Class-specific loadout templates
	switch class {
	case ClassWarrior:
		return []EquipmentLoadout{
			{
				Name: "Heavy Armor", Description: "Maximum defense, slower attacks",
				MainHand: "Greatsword", OffHand: "None (2H)", Armor: "Plate Mail", Accessory: "Iron Ring",
				BonusHP: 20, BonusAttack: -1, BonusDefense: 4,
			},
			{
				Name: "Balanced", Description: "Versatile sword and shield combo",
				MainHand: "Longsword", OffHand: "Tower Shield", Armor: "Chain Mail", Accessory: "Leather Gloves",
				BonusHP: 10, BonusAttack: 1, BonusDefense: 2,
			},
			{
				Name: "Berserker", Description: "High damage, lower defense",
				MainHand: "Battle Axe", OffHand: "Throwing Axes", Armor: "Hide Armor", Accessory: "Wolf Fang",
				BonusHP: 5, BonusAttack: 3, BonusDefense: 0,
			},
		}
	case ClassMage:
		return []EquipmentLoadout{
			{
				Name: "Elementalist", Description: "Balanced elemental magic",
				MainHand: "Oak Staff", OffHand: "Spell Focus", Armor: "Silk Robes", Accessory: "Crystal Orb",
				BonusHP: 0, BonusAttack: 2, BonusDefense: 1,
			},
			{
				Name: "Battle Mage", Description: "Combat-focused spellcaster",
				MainHand: "War Staff", OffHand: "Buckler", Armor: "Leather Robes", Accessory: "Combat Amulet",
				BonusHP: 15, BonusAttack: 1, BonusDefense: 2,
			},
			{
				Name: "Scholar", Description: "Maximum mana regeneration",
				MainHand: "Elder Wand", OffHand: "Tome of Wisdom", Armor: "Scholar's Robes", Accessory: "Mana Ring",
				BonusHP: -5, BonusAttack: 3, BonusDefense: 0,
			},
		}
	case ClassRogue:
		return []EquipmentLoadout{
			{
				Name: "Assassin", Description: "High critical hit chance",
				MainHand: "Shadow Dagger", OffHand: "Throwing Knives", Armor: "Shadow Leather", Accessory: "Poison Vial",
				BonusHP: 0, BonusAttack: 3, BonusDefense: 0,
			},
			{
				Name: "Duelist", Description: "Quick strikes with parrying",
				MainHand: "Rapier", OffHand: "Parrying Dagger", Armor: "Studded Leather", Accessory: "Swift Boots",
				BonusHP: 5, BonusAttack: 2, BonusDefense: 1,
			},
			{
				Name: "Brigand", Description: "Versatile combat rogue",
				MainHand: "Short Sword", OffHand: "Buckler", Armor: "Traveler's Garb", Accessory: "Lockpicks",
				BonusHP: 10, BonusAttack: 1, BonusDefense: 2,
			},
		}
	case ClassRanger:
		return []EquipmentLoadout{
			{
				Name: "Sharpshooter", Description: "Maximum ranged damage",
				MainHand: "Longbow", OffHand: "Quiver", Armor: "Camouflage Cloak", Accessory: "Eagle Eye",
				BonusHP: 0, BonusAttack: 3, BonusDefense: 0,
			},
			{
				Name: "Beastmaster", Description: "Enhanced pet bonding",
				MainHand: "Hunting Bow", OffHand: "Beast Whistle", Armor: "Fur Cloak", Accessory: "Pet Collar",
				BonusHP: 10, BonusAttack: 1, BonusDefense: 1,
			},
			{
				Name: "Survivalist", Description: "Balanced wilderness warrior",
				MainHand: "Composite Bow", OffHand: "Hunting Knife", Armor: "Scout Armor", Accessory: "Trap Kit",
				BonusHP: 5, BonusAttack: 2, BonusDefense: 1,
			},
		}
	case ClassCleric:
		return []EquipmentLoadout{
			{
				Name: "Battle Priest", Description: "Melee-focused healer",
				MainHand: "War Mace", OffHand: "Holy Shield", Armor: "Blessed Plate", Accessory: "Holy Symbol",
				BonusHP: 15, BonusAttack: 1, BonusDefense: 2,
			},
			{
				Name: "Divine Healer", Description: "Maximum healing power",
				MainHand: "Healing Staff", OffHand: "Prayer Book", Armor: "Priest Robes", Accessory: "Ankh",
				BonusHP: 5, BonusAttack: 0, BonusDefense: 1,
			},
			{
				Name: "Crusader", Description: "Offensive holy warrior",
				MainHand: "Morning Star", OffHand: "Blessed Banner", Armor: "Chain Mail", Accessory: "Sun Pendant",
				BonusHP: 10, BonusAttack: 2, BonusDefense: 1,
			},
		}
	case ClassNecromancer:
		return []EquipmentLoadout{
			{
				Name: "Summoner", Description: "Enhanced undead minions",
				MainHand: "Bone Staff", OffHand: "Skull Talisman", Armor: "Death Shroud", Accessory: "Soul Gem",
				BonusHP: 5, BonusAttack: 1, BonusDefense: 1,
			},
			{
				Name: "Blood Mage", Description: "Life drain focused",
				MainHand: "Blood Dagger", OffHand: "Vampire Cloak", Armor: "Crimson Robes", Accessory: "Blood Ruby",
				BonusHP: 10, BonusAttack: 2, BonusDefense: 0,
			},
			{
				Name: "Plague Bearer", Description: "Disease and debuffs",
				MainHand: "Plague Staff", OffHand: "Poison Flask", Armor: "Tattered Robes", Accessory: "Rat Skull",
				BonusHP: 0, BonusAttack: 3, BonusDefense: 0,
			},
		}
	default:
		// Hybrid and other classes get generic loadouts with randomized bonuses
		prefixes := []string{"Light", "Balanced", "Heavy"}
		return generateGenericLoadouts(class, rng, prefixes)
	}
}

// generateGenericLoadouts creates loadouts for hybrid classes using randomized stats.
func generateGenericLoadouts(class CharacterClass, rng *rand.Rand, prefixes []string) []EquipmentLoadout {
	loadouts := make([]EquipmentLoadout, 3)
	className := class.String()

	for i := 0; i < 3; i++ {
		loadouts[i] = EquipmentLoadout{
			Name:         fmt.Sprintf("%s %s", prefixes[i], className),
			Description:  fmt.Sprintf("A %s loadout suited for %s", strings.ToLower(prefixes[i]), className),
			MainHand:     fmt.Sprintf("%s Weapon", className),
			OffHand:      fmt.Sprintf("%s Off-Hand", className),
			Armor:        fmt.Sprintf("%s Armor", prefixes[i]),
			Accessory:    fmt.Sprintf("%s Trinket", className),
			BonusHP:      rng.Intn(15) - 5,
			BonusAttack:  rng.Intn(4),
			BonusDefense: rng.Intn(3),
		}
	}
	return loadouts
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
	if !IsTouchOrMouseJustPressed() || cc.stepChangedThisFrame {
		return false
	}

	mouseX, mouseY, _ := GetTouchOrMousePosition()
	helpY := cc.panelY + cc.panelHeight - 160
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

// returnToSubclassSelection navigates back to subclass selection step from portrait.
func (cc *EbitenCharacterCreation) returnToClassSelection() {
	cc.currentStep = stepSubclassSelection
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

// generateRandomName generates a random character name deterministically.
// Uses the stored world seed combined with a counter for reproducible but varied names.
func (cc *EbitenCharacterCreation) generateRandomName() string {
	prefixes := []string{"Brave", "Swift", "Dark", "Elder", "Noble", "Shadow", "Storm", "Iron", "Silver", "Golden"}
	suffixes := []string{"blade", "heart", "fist", "eye", "soul", "wind", "fire", "steel", "wing", "star"}

	// Combine world seed with counter for deterministic but varied generation
	// Each call produces a different name, but same seed+counter always produces same name
	combinedSeed := cc.worldSeed + int64(cc.nameGenCounter)*1000 + int64(cc.selectedClass)*100
	cc.nameGenCounter++

	rng := rand.New(rand.NewSource(combinedSeed))
	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

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
		// Proceed to subclass selection
		cc.characterData.Class = cc.selectedClass
		cc.selectedSubclass = noSubclassSelected // Reset subclass selection
		cc.currentStep = stepSubclassSelection
		cc.stepChangedThisFrame = true
		cc.errorMsg = ""
	case stepSubclassSelection:
		// Set final class (subclass if chosen, else base class) and proceed to portrait
		if cc.selectedSubclass == noSubclassSelected {
			cc.characterData.Class = cc.selectedClass
		} else {
			cc.characterData.Class = cc.selectedSubclass
		}
		cc.currentStep = stepPortraitSelection
		cc.stepChangedThisFrame = true
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
	case stepSubclassSelection:
		cc.currentStep = stepClassSelection
		cc.hideKeyboardIfNeeded()
	case stepPortraitSelection:
		cc.currentStep = stepSubclassSelection
		cc.hideKeyboardIfNeeded()
	case stepConfirmation:
		cc.currentStep = stepPortraitSelection
		cc.keyboardShown = false
	}
	// Mark step changed so the new step's click handlers don't consume
	// the same mouse/touch event that triggered this back navigation.
	cc.stepChangedThisFrame = true
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
	if IsTouchOrMouseJustPressed() && !cc.stepChangedThisFrame {
		mouseX, mouseY, _ := GetTouchOrMousePosition()

		// Define button areas (matching drawConfirmation layout)
		buttonX := cc.panelX + 50
		buttonW := cc.panelWidth - 100
		buttonH := 30

		// Confirm button area (matches drawConfirmation layout)
		confirmButtonY := cc.panelY + cc.panelHeight - 130

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
	case stepSubclassSelection:
		cc.drawSubclassSelection(screen, cc.panelX, cc.panelY, cc.panelWidth, cc.panelHeight)
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
	cc.drawNextButton(screen)
	cc.drawStepSpecificButtons(screen)
}

// drawNextButton renders the Next/Done button with appropriate label.
func (cc *EbitenCharacterCreation) drawNextButton(screen *ebiten.Image) {
	if cc.nextButton == nil {
		return
	}
	if cc.currentStep == stepConfirmation {
		cc.nextButton.Label = "Done"
	} else {
		cc.nextButton.Label = "Next"
	}
	cc.nextButton.Draw(screen)
}

// drawStepSpecificButtons renders buttons specific to the current creation step.
func (cc *EbitenCharacterCreation) drawStepSpecificButtons(screen *ebiten.Image) {
	switch cc.currentStep {
	case stepNameInput:
		cc.drawPresetNameButtons(screen)
	case stepClassSelection:
		cc.drawBackButton(screen)
	case stepSubclassSelection:
		cc.drawBackButton(screen)
	case stepPortraitSelection:
		cc.drawSkipAndBackButtons(screen)
	case stepConfirmation:
		cc.drawBackButton(screen)
	}
}

// drawPresetNameButtons renders preset name selection buttons for mobile.
func (cc *EbitenCharacterCreation) drawPresetNameButtons(screen *ebiten.Image) {
	for _, btn := range cc.presetNameButtons {
		if btn != nil {
			btn.Draw(screen)
		}
	}
}

// drawBackButton renders the back navigation button.
func (cc *EbitenCharacterCreation) drawBackButton(screen *ebiten.Image) {
	if cc.backButton != nil {
		cc.backButton.Draw(screen)
	}
}

// drawSkipAndBackButtons renders both skip and back buttons.
func (cc *EbitenCharacterCreation) drawSkipAndBackButtons(screen *ebiten.Image) {
	if cc.skipButton != nil {
		cc.skipButton.Draw(screen)
	}
	cc.drawBackButton(screen)
}

// drawNameInput renders the name input screen
func (cc *EbitenCharacterCreation) drawNameInput(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 1 of 5: Choose Your Name"
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
	stepText := "Step 2 of 5: Choose Your Base Class"
	stepX := x + w/2 - len(stepText)*3
	text.Draw(screen, stepText, basicfont.Face7x13, stepX, y+70,
		color.RGBA{200, 200, 200, 255})

	// Display name
	nameText := fmt.Sprintf("Name: %s", cc.characterData.Name)
	nameX := x + 30
	text.Draw(screen, nameText, basicfont.Face7x13, nameX, y+100,
		color.RGBA{200, 200, 255, 255})

	// Page indicator and category title
	pageTitle := getPageTitle(cc.classPage)
	pageInfo := fmt.Sprintf("%s (Page %d/%d)", pageTitle, cc.classPage+1, totalClassPages())
	pageInfoX := x + w/2 - len(pageInfo)*3
	text.Draw(screen, pageInfo, basicfont.Face7x13, pageInfoX, y+120,
		color.RGBA{100, 200, 255, 255})

	// Get classes for current page
	currentClasses := getClassesForPage(cc.classPage)
	startY := y + 140

	for i, class := range currentClasses {
		classY := startY + i*55
		isSelected := class == cc.selectedClass

		// Selection indicator
		if isSelected {
			vector.DrawFilledRect(screen, float32(x+40), float32(classY-5), float32(w-80), 50,
				color.RGBA{50, 80, 120, 255}, false)
		}

		// Class name
		classColor := color.RGBA{200, 200, 200, 255}
		if isSelected {
			classColor = color.RGBA{255, 255, 100, 255}
		}

		className := fmt.Sprintf("%d. %s", i+1, class.String())
		text.Draw(screen, className, basicfont.Face7x13, x+50, classY+15, classColor)

		// Class description (compact - single line)
		desc := class.Description()
		descLines := wrapText(desc, 55)
		if len(descLines) > 0 {
			text.Draw(screen, descLines[0], basicfont.Face7x13, x+70, classY+32,
				color.RGBA{180, 180, 180, 255})
		}
	}

	// Show current default if set (defaults to ClassWarrior as zero value)
	// Only show if explicitly set, which we track by checking if DefaultName is also set
	if cc.defaults.DefaultName != "" {
		defaultText := fmt.Sprintf("Current default: %s", cc.defaults.DefaultClass.String())
		defaultX := x + w/2 - len(defaultText)*3
		text.Draw(screen, defaultText, basicfont.Face7x13, defaultX, y+h-130,
			color.RGBA{150, 150, 150, 255})
	}

	// Help text with page navigation hint
	helpText1 := "Use ARROW KEYS or 1-6 to select"
	helpText2 := "TAB/PageUp/PageDown to switch class pages"
	helpText3 := "TAP/CLICK a class to select and continue"
	helpText4 := "Press ENTER or click NEXT to continue"
	helpText5 := "BACKSPACE or click BACK to go back | F2 to save default"
	helpX1 := x + w/2 - len(helpText1)*3
	helpX2 := x + w/2 - len(helpText2)*3
	helpX3 := x + w/2 - len(helpText3)*3
	helpX4 := x + w/2 - len(helpText4)*3
	helpX5 := x + w/2 - len(helpText5)*3
	text.Draw(screen, helpText1, basicfont.Face7x13, helpX1, y+h-105,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText2, basicfont.Face7x13, helpX2, y+h-90,
		color.RGBA{100, 180, 255, 255}) // Blue to highlight page nav
	text.Draw(screen, helpText3, basicfont.Face7x13, helpX3, y+h-75,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText4, basicfont.Face7x13, helpX4, y+h-60,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText5, basicfont.Face7x13, helpX5, y+h-45,
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
	stepText := "Step 4 of 5: Choose Portrait (Optional)"
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

	// Help text — position drawn buttons well above the touch button row
	// Touch buttons occupy the bottom 54px of the panel.
	helpY := y + h - 160

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

// drawEquipmentSelection renders the equipment loadout selection screen.
// drawSubclassSelection renders the subclass selection screen.
// Shows hybrid class options for the chosen base class, plus a "None" option.
func (cc *EbitenCharacterCreation) drawSubclassSelection(screen *ebiten.Image, x, y, w, h int) {
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	stepText := "Step 3 of 5: Choose Your Subclass (Optional)"
	stepX := x + w/2 - len(stepText)*3
	text.Draw(screen, stepText, basicfont.Face7x13, stepX, y+60,
		color.RGBA{200, 200, 200, 255})

	baseText := fmt.Sprintf("Base Class: %s", cc.selectedClass.String())
	baseX := x + w/2 - len(baseText)*3
	text.Draw(screen, baseText, basicfont.Face7x13, baseX, y+80,
		color.RGBA{150, 200, 255, 255})

	subclasses := getSubclassesForBaseClass(cc.selectedClass)
	startY := y + 130

	// Draw "None" option first
	noneSelected := cc.selectedSubclass == noSubclassSelected
	noneColor := color.RGBA{150, 150, 150, 255}
	if noneSelected {
		noneColor = color.RGBA{255, 255, 100, 255}
		vector.DrawFilledRect(screen, float32(x+35), float32(startY-5),
			float32(w-70), 45, color.RGBA{40, 60, 40, 255}, false)
	}
	noneText := fmt.Sprintf("1. None (Stay as %s)", cc.selectedClass.String())
	text.Draw(screen, noneText, basicfont.Face7x13, x+45, startY+15, noneColor)

	// Draw subclass options
	for i, sc := range subclasses {
		optionY := startY + (i+1)*55
		isSelected := cc.selectedSubclass == sc

		bgColor := color.RGBA{20, 20, 30, 255}
		nameColor := color.RGBA{200, 200, 200, 255}
		if isSelected {
			bgColor = color.RGBA{40, 60, 80, 255}
			nameColor = color.RGBA{255, 255, 100, 255}
		}

		vector.DrawFilledRect(screen, float32(x+35), float32(optionY-5),
			float32(w-70), 45, bgColor, false)
		vector.StrokeRect(screen, float32(x+35), float32(optionY-5),
			float32(w-70), 45, 1, color.RGBA{80, 80, 120, 255}, false)

		labelText := fmt.Sprintf("%d. %s", i+2, sc.String())
		text.Draw(screen, labelText, basicfont.Face7x13, x+45, optionY+10, nameColor)

		descText := sc.Description()
		if len(descText) > 60 {
			descText = descText[:57] + "..."
		}
		text.Draw(screen, descText, basicfont.Face7x13, x+45, optionY+25,
			color.RGBA{150, 150, 170, 255})
	}

	// Controls hint
	hintY := y + h - 60
	hintText := "Arrow Keys/WASD: Navigate | Enter: Confirm | ESC/Backspace: Back | 1-7: Quick Select"
	hintX := x + w/2 - len(hintText)*3
	text.Draw(screen, hintText, basicfont.Face7x13, hintX, hintY,
		color.RGBA{150, 150, 150, 255})
}

// drawConfirmation renders the confirmation screen
func (cc *EbitenCharacterCreation) drawConfirmation(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 5 of 5: Confirm Your Character"
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

	// Equipment loadout preview
	if cc.characterData.StartingLoadout != nil {
		loadoutText := fmt.Sprintf("Loadout: %s", cc.characterData.StartingLoadout.Name)
		text.Draw(screen, loadoutText, basicfont.Face7x13, x+w/2-len(loadoutText)*3, summaryY+60,
			color.RGBA{180, 200, 255, 255})
	} else {
		loadoutText := "Loadout: Default"
		text.Draw(screen, loadoutText, basicfont.Face7x13, x+w/2-len(loadoutText)*3, summaryY+60,
			color.RGBA{180, 180, 180, 255})
	}

	// Portrait preview (if set)
	if cc.characterData.Portrait != nil {
		portraitText := fmt.Sprintf("Portrait: Custom (%dx%d)",
			cc.characterData.Portrait.Bounds().Dx(),
			cc.characterData.Portrait.Bounds().Dy())
		text.Draw(screen, portraitText, basicfont.Face7x13, x+w/2-len(portraitText)*3, summaryY+90,
			color.RGBA{255, 255, 255, 255})

		// Show small preview
		previewSize := 48
		previewX := x + w/2 - previewSize/2
		previewY := summaryY + 110

		opts := &ebiten.DrawImageOptions{}
		// Scale down to 48x48 preview
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
		text.Draw(screen, portraitText, basicfont.Face7x13, x+w/2-len(portraitText)*3, summaryY+90,
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

	// Draw clickable buttons for touch support — above the touch button row
	// Touch buttons occupy the bottom 54px of the panel.
	buttonX := x + 50
	buttonW := w - 100
	buttonH := 30

	// Confirm button (green/positive action)
	confirmButtonY := y + h - 130
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
	case ClassRanger:
		return []string{
			"Health: 110 (Medium)",
			"Mana: 70 (Low)",
			"Attack: 11 (Medium)",
			"Defense: 5 (Medium)",
			"Crit: 12% (High)",
		}
	case ClassCleric:
		return []string{
			"Health: 120 (Medium)",
			"Mana: 120 (High)",
			"Attack: 7 (Low)",
			"Defense: 6 (Medium)",
			"Mana Regen: 6.0 (High)",
		}
	case ClassNecromancer:
		return []string{
			"Health: 90 (Low)",
			"Mana: 140 (High)",
			"Attack: 8 (Low)",
			"Defense: 4 (Low)",
			"Mana Regen: 7.0 (High)",
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

	// Reset subclass selection state
	cc.selectedSubclass = noSubclassSelected

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

	rawComps, err := getRawComponents(player)
	if err != nil {
		return nil, err
	}

	return assertComponentTypes(rawComps)
}

// getRawComponents retrieves required components from player entity.
func getRawComponents(player *Entity) (map[string]interface{}, error) {
	required := []string{"health", "mana", "stats", "attack"}
	rawComps := make(map[string]interface{})

	for _, name := range required {
		comp, has := player.GetComponent(name)
		if !has {
			return nil, fmt.Errorf("player missing %s component", name)
		}
		rawComps[name] = comp
	}

	return rawComps, nil
}

// assertComponentTypes type-asserts components to their concrete types.
func assertComponentTypes(rawComps map[string]interface{}) (*classComponents, error) {
	health, ok := rawComps["health"].(*HealthComponent)
	if !ok {
		return nil, fmt.Errorf("health component has wrong type")
	}
	mana, ok := rawComps["mana"].(*ManaComponent)
	if !ok {
		return nil, fmt.Errorf("mana component has wrong type")
	}
	stats, ok := rawComps["stats"].(*StatsComponent)
	if !ok {
		return nil, fmt.Errorf("stats component has wrong type")
	}
	attack, ok := rawComps["attack"].(*AttackComponent)
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

// applyRangerStats configures a player entity with ranger class statistics
func applyRangerStats(comps *classComponents) {
	comps.health.Max = 110
	comps.health.Current = 110
	comps.mana.Max = 70
	comps.mana.Current = 70
	comps.stats.Attack = 11
	comps.stats.Defense = 5
	comps.attack.Damage = 16
	comps.attack.Cooldown = 0.4
	comps.stats.CritChance = 0.12
	comps.stats.CritDamage = 2.2
	comps.stats.Evasion = 0.10
}

// applyClericStats configures a player entity with cleric class statistics
func applyClericStats(comps *classComponents) {
	comps.health.Max = 120
	comps.health.Current = 120
	comps.mana.Max = 120
	comps.mana.Current = 120
	comps.mana.Regen = 6.0
	comps.stats.Attack = 7
	comps.stats.Defense = 6
	comps.attack.Damage = 12
	comps.stats.CritChance = 0.08
	comps.stats.CritDamage = 1.6
}

// applyNecromancerStats configures a player entity with necromancer class statistics
func applyNecromancerStats(comps *classComponents) {
	comps.health.Max = 90
	comps.health.Current = 90
	comps.mana.Max = 140
	comps.mana.Current = 140
	comps.mana.Regen = 7.0
	comps.stats.Attack = 8
	comps.stats.Defense = 4
	comps.attack.Damage = 14
	comps.stats.CritChance = 0.10
	comps.stats.CritDamage = 1.9
}

// applyBattlemageStats configures a player entity with battlemage class statistics
func applyBattlemageStats(comps *classComponents) {
	comps.health.Max = 115
	comps.health.Current = 115
	comps.mana.Max = 100
	comps.mana.Current = 100
	comps.mana.Regen = 5.0
	comps.stats.Attack = 10
	comps.stats.Defense = 6
	comps.attack.Damage = 16
	comps.stats.CritChance = 0.08
	comps.stats.CritDamage = 1.9
}

// applySpellbladeStats configures a player entity with spellblade class statistics
func applySpellbladeStats(comps *classComponents) {
	comps.health.Max = 90
	comps.health.Current = 90
	comps.mana.Max = 110
	comps.mana.Current = 110
	comps.mana.Regen = 6.0
	comps.stats.Attack = 9
	comps.stats.Defense = 4
	comps.attack.Damage = 14
	comps.attack.Cooldown = 0.35
	comps.stats.CritChance = 0.12
	comps.stats.CritDamage = 2.1
	comps.stats.Evasion = 0.10
}

// applyPaladinStats configures a player entity with paladin class statistics
func applyPaladinStats(comps *classComponents) {
	comps.health.Max = 140
	comps.health.Current = 140
	comps.mana.Max = 80
	comps.mana.Current = 80
	comps.mana.Regen = 4.0
	comps.stats.Attack = 10
	comps.stats.Defense = 9
	comps.attack.Damage = 17
	comps.stats.CritChance = 0.06
	comps.stats.CritDamage = 1.8
}

// applyMonkStats configures a player entity with monk class statistics
func applyMonkStats(comps *classComponents) {
	comps.health.Max = 100
	comps.health.Current = 100
	comps.mana.Max = 90
	comps.mana.Current = 90
	comps.mana.Regen = 5.0
	comps.stats.Attack = 9
	comps.stats.Defense = 5
	comps.attack.Damage = 13
	comps.attack.Cooldown = 0.25
	comps.stats.CritChance = 0.14
	comps.stats.CritDamage = 2.3
	comps.stats.Evasion = 0.18
}

// applyDeathKnightStats configures a player entity with death knight class statistics
func applyDeathKnightStats(comps *classComponents) {
	comps.health.Max = 130
	comps.health.Current = 130
	comps.mana.Max = 90
	comps.mana.Current = 90
	comps.mana.Regen = 4.0
	comps.stats.Attack = 11
	comps.stats.Defense = 7
	comps.attack.Damage = 19
	comps.stats.CritChance = 0.07
	comps.stats.CritDamage = 2.0
}

// applyWitchHunterStats configures a player entity with witch hunter class statistics
func applyWitchHunterStats(comps *classComponents) {
	comps.health.Max = 115
	comps.health.Current = 115
	comps.mana.Max = 90
	comps.mana.Current = 90
	comps.mana.Regen = 5.0
	comps.stats.Attack = 10
	comps.stats.Defense = 5
	comps.attack.Damage = 15
	comps.attack.Cooldown = 0.4
	comps.stats.CritChance = 0.11
	comps.stats.CritDamage = 2.1
}

// applyBeastlordStats configures a player entity with beastlord class statistics
func applyBeastlordStats(comps *classComponents) {
	comps.health.Max = 135
	comps.health.Current = 135
	comps.mana.Max = 60
	comps.mana.Current = 60
	comps.stats.Attack = 11
	comps.stats.Defense = 7
	comps.attack.Damage = 18
	comps.stats.CritChance = 0.08
	comps.stats.CritDamage = 2.0
	comps.stats.Evasion = 0.05
}

// applyArcaneArcherStats configures a player entity with arcane archer class statistics
func applyArcaneArcherStats(comps *classComponents) {
	comps.health.Max = 95
	comps.health.Current = 95
	comps.mana.Max = 110
	comps.mana.Current = 110
	comps.mana.Regen = 6.0
	comps.stats.Attack = 10
	comps.stats.Defense = 4
	comps.attack.Damage = 15
	comps.attack.Cooldown = 0.4
	comps.stats.CritChance = 0.12
	comps.stats.CritDamage = 2.1
}

// applyShadowPriestStats configures a player entity with shadow priest class statistics
func applyShadowPriestStats(comps *classComponents) {
	comps.health.Max = 85
	comps.health.Current = 85
	comps.mana.Max = 130
	comps.mana.Current = 130
	comps.mana.Regen = 7.0
	comps.stats.Attack = 8
	comps.stats.Defense = 4
	comps.attack.Damage = 13
	comps.stats.CritChance = 0.13
	comps.stats.CritDamage = 2.2
	comps.stats.Evasion = 0.08
}

// applyDruidStats configures a player entity with druid class statistics
func applyDruidStats(comps *classComponents) {
	comps.health.Max = 105
	comps.health.Current = 105
	comps.mana.Max = 115
	comps.mana.Current = 115
	comps.mana.Regen = 6.0
	comps.stats.Attack = 9
	comps.stats.Defense = 5
	comps.attack.Damage = 14
	comps.stats.CritChance = 0.10
	comps.stats.CritDamage = 1.9
	comps.stats.Evasion = 0.05
}

// applyInquisitorStats configures a player entity with inquisitor class statistics
func applyInquisitorStats(comps *classComponents) {
	comps.health.Max = 110
	comps.health.Current = 110
	comps.mana.Max = 100
	comps.mana.Current = 100
	comps.mana.Regen = 5.0
	comps.stats.Attack = 9
	comps.stats.Defense = 6
	comps.attack.Damage = 14
	comps.attack.Cooldown = 0.35
	comps.stats.CritChance = 0.11
	comps.stats.CritDamage = 2.0
	comps.stats.Evasion = 0.08
}

// applyBloodKnightStats configures a player entity with blood knight class statistics
func applyBloodKnightStats(comps *classComponents) {
	comps.health.Max = 125
	comps.health.Current = 125
	comps.mana.Max = 85
	comps.mana.Current = 85
	comps.mana.Regen = 4.0
	comps.stats.Attack = 12
	comps.stats.Defense = 6
	comps.attack.Damage = 21
	comps.stats.CritChance = 0.09
	comps.stats.CritDamage = 2.1
}

// applyMysticStats configures a player entity with mystic class statistics
func applyMysticStats(comps *classComponents) {
	comps.health.Max = 95
	comps.health.Current = 95
	comps.mana.Max = 135
	comps.mana.Current = 135
	comps.mana.Regen = 8.0
	comps.stats.Attack = 7
	comps.stats.Defense = 5
	comps.attack.Damage = 11
	comps.stats.CritChance = 0.10
	comps.stats.CritDamage = 1.8
}

// applyWarlockStats configures a player entity with warlock class statistics
func applyWarlockStats(comps *classComponents) {
	comps.health.Max = 85
	comps.health.Current = 85
	comps.mana.Max = 145
	comps.mana.Current = 145
	comps.mana.Regen = 7.0
	comps.stats.Attack = 9
	comps.stats.Defense = 3
	comps.attack.Damage = 16
	comps.stats.CritChance = 0.11
	comps.stats.CritDamage = 2.0
}

// applyNinjaStats configures a player entity with ninja class statistics
func applyNinjaStats(comps *classComponents) {
	comps.health.Max = 90
	comps.health.Current = 90
	comps.mana.Max = 75
	comps.mana.Current = 75
	comps.stats.Attack = 11
	comps.stats.Defense = 4
	comps.attack.Damage = 17
	comps.attack.Cooldown = 0.25
	comps.stats.CritChance = 0.18
	comps.stats.CritDamage = 2.8
	comps.stats.Evasion = 0.20
}

// ApplyClassStats configures a player entity with stats appropriate for the given class
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
	case ClassRanger:
		applyRangerStats(comps)
	case ClassCleric:
		applyClericStats(comps)
	case ClassNecromancer:
		applyNecromancerStats(comps)
	case ClassBattlemage:
		applyBattlemageStats(comps)
	case ClassSpellblade:
		applySpellbladeStats(comps)
	case ClassPaladin:
		applyPaladinStats(comps)
	case ClassMonk:
		applyMonkStats(comps)
	case ClassDeathKnight:
		applyDeathKnightStats(comps)
	case ClassWitchHunter:
		applyWitchHunterStats(comps)
	case ClassBeastlord:
		applyBeastlordStats(comps)
	case ClassArcaneArcher:
		applyArcaneArcherStats(comps)
	case ClassShadowPriest:
		applyShadowPriestStats(comps)
	case ClassDruid:
		applyDruidStats(comps)
	case ClassInquisitor:
		applyInquisitorStats(comps)
	case ClassBloodKnight:
		applyBloodKnightStats(comps)
	case ClassMystic:
		applyMysticStats(comps)
	case ClassWarlock:
		applyWarlockStats(comps)
	case ClassNinja:
		applyNinjaStats(comps)
	default:
		return fmt.Errorf("unknown character class: %v", class)
	}

	return nil
}
