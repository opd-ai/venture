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
	// ClassWarrior is a high HP, melee-focused class
	ClassWarrior CharacterClass = iota
	// ClassMage is a high mana, magic-focused class
	ClassMage
	// ClassRogue is a balanced, agility-focused class
	ClassRogue
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
	default:
		return ""
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
	if cd.Class < ClassWarrior || cd.Class > ClassRogue {
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

	// Mobile keyboard state (WASM/mobile platforms)
	keyboardShown bool // Tracks whether mobile keyboard is currently shown

	screenWidth  int
	screenHeight int

	// Panel layout cache (calculated in Draw, used in Update for hit detection)
	panelX      int
	panelY      int
	panelWidth  int
	panelHeight int
}

// NewCharacterCreation creates a new character creation system
func NewCharacterCreation(screenWidth, screenHeight int) *EbitenCharacterCreation {
	return &EbitenCharacterCreation{
		currentStep:   stepNameInput,
		selectedClass: ClassWarrior, // Default selection
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		inputBuffer:   make([]rune, 0),
		defaults: CharacterCreationDefaults{
			DefaultName:  "", // No default initially
			DefaultClass: ClassWarrior,
		},
	}
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

// GetDefaults returns the current default values
func (cc *EbitenCharacterCreation) GetDefaults() CharacterCreationDefaults {
	return cc.defaults
}

// Update handles input for character creation (keyboard/mouse navigation)
// Returns true when character creation is complete
func (cc *EbitenCharacterCreation) Update() bool {
	// Calculate panel dimensions first (needed for touch hit detection)
	// This must be done before processing input
	cc.updatePanelDimensions()

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
	// MOBILE/WASM: Show keyboard when entering name input step
	// The native mobile keyboard needs to be explicitly triggered on WASM builds
	// because the game runs in a canvas element which doesn't automatically focus
	if !cc.keyboardShown && mobile.IsWASM() {
		mobile.ShowKeyboard()
		cc.keyboardShown = true
	}

	// Handle text input
	cc.inputBuffer = ebiten.AppendInputChars(cc.inputBuffer[:0])
	for _, r := range cc.inputBuffer {
		// Only allow alphanumeric and spaces
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
			if len(cc.nameInput) < 20 {
				cc.nameInput += string(r)
			}
		}
	}

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(cc.nameInput) > 0 {
			cc.nameInput = cc.nameInput[:len(cc.nameInput)-1]
		}
	}

	// Handle enter to proceed
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if len(strings.TrimSpace(cc.nameInput)) > 0 {
			cc.characterData.Name = cc.nameInput
			cc.currentStep = stepClassSelection
			cc.errorMsg = ""

			// MOBILE/WASM: Hide keyboard when leaving name input
			if cc.keyboardShown && mobile.IsWASM() {
				mobile.HideKeyboard()
				cc.keyboardShown = false
			}
		} else {
			cc.errorMsg = "Name cannot be empty"
		}
	}

	// F2 to save current name as default
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		if len(strings.TrimSpace(cc.nameInput)) > 0 {
			cc.defaults.DefaultName = strings.TrimSpace(cc.nameInput)
			cc.errorMsg = "Default name saved!"
		}
	}
}

// updateClassSelection handles class selection with keyboard/mouse
func (cc *EbitenCharacterCreation) updateClassSelection() {
	// Arrow keys for selection
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

	// Number keys for direct selection
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		cc.selectedClass = ClassWarrior
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		cc.selectedClass = ClassMage
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		cc.selectedClass = ClassRogue
	}

	// Handle mouse and touch input (Touch support for WASM/mobile)
	if IsTouchOrMouseJustPressed() {
		mouseX, mouseY, _ := GetTouchOrMousePosition()

		// Use cached panel dimensions from Draw method
		// Class selection area starts at y+140 with 80px spacing
		startY := cc.panelY + 140
		classes := []CharacterClass{ClassWarrior, ClassMage, ClassRogue}

		// Check if touch/click is within any class option area
		for i, class := range classes {
			classY := startY + i*80
			// Each class box is from x+40 to x+w-40, and classY-5 to classY+65 (70px height)
			if mouseX >= cc.panelX+40 && mouseX <= cc.panelX+cc.panelWidth-40 &&
				mouseY >= classY-5 && mouseY <= classY+65 {
				// Clicked on this class - select it and proceed
				cc.selectedClass = class
				cc.characterData.Class = cc.selectedClass
				cc.currentStep = stepPortraitSelection
				return
			}
		}
	}

	// Update selection highlight on mouse/touch hover (Touch support for WASM/mobile)
	mouseX, mouseY, _ := GetTouchOrMousePosition()

	// Use cached panel dimensions from Draw method
	// Class selection area starts at y+140 with 80px spacing
	startY := cc.panelY + 140
	classes := []CharacterClass{ClassWarrior, ClassMage, ClassRogue}

	// Check if hovering over any class option
	for i, class := range classes {
		classY := startY + i*80
		if mouseX >= cc.panelX+40 && mouseX <= cc.panelX+cc.panelWidth-40 &&
			mouseY >= classY-5 && mouseY <= classY+65 {
			// Hovering over this class - highlight it
			cc.selectedClass = class
			break
		}
	}

	// Enter to proceed, Backspace to go back
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		cc.characterData.Class = cc.selectedClass
		cc.currentStep = stepPortraitSelection
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepNameInput
		// MOBILE/WASM FIX: Reset keyboard flag so updateNameInput will show it
		// when entering the name input step on next Update()
		cc.keyboardShown = false
	}

	// F2 to save current class as default
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

	// Handle mouse and touch input (Touch support for WASM/mobile)
	if IsTouchOrMouseJustPressed() {
		mouseX, mouseY, _ := GetTouchOrMousePosition()

		// Define button areas (matching drawPortraitSelection layout)
		helpY := cc.panelY + cc.panelHeight - 100
		buttonY := helpY - 10
		buttonX := cc.panelX + 50
		buttonW := cc.panelWidth - 100
		buttonH := 25

		// Browse button area
		browseButtonY := buttonY

		// Skip button area
		skipButtonY := browseButtonY + 35

		// Back button area
		backButtonY := skipButtonY + 35

		// Check browse button click
		if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
			mouseY >= browseButtonY && mouseY <= browseButtonY+buttonH {
			// Trigger file browser dialog
			go func() {
				filename, err := OpenPortraitDialog()
				if err != nil {
					cc.errorMsg = fmt.Sprintf("Dialog error: %v", err)
					return
				}
				if filename != "" {
					cc.portraitInput = filename
				}
			}()
			return
		}

		// Check skip button click
		if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
			mouseY >= skipButtonY && mouseY <= skipButtonY+buttonH {
			cc.characterData.PortraitPath = ""
			cc.characterData.Portrait = nil
			cc.currentStep = stepConfirmation
			cc.errorMsg = ""
			if cc.keyboardShown && mobile.IsWASM() {
				mobile.HideKeyboard()
				cc.keyboardShown = false
			}
			return
		}

		// Check back button click
		if mouseX >= buttonX && mouseX <= buttonX+buttonW &&
			mouseY >= backButtonY && mouseY <= backButtonY+buttonH {
			cc.currentStep = stepClassSelection
			if cc.keyboardShown && mobile.IsWASM() {
				mobile.HideKeyboard()
				cc.keyboardShown = false
			}
			return
		}
	}

	// SPACE or B key to open file browser dialog
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyB) {
		// Open file dialog (this will block until user selects or cancels)
		go func() {
			filename, err := OpenPortraitDialog()
			if err != nil {
				cc.errorMsg = fmt.Sprintf("Dialog error: %v", err)
				return
			}
			if filename != "" {
				cc.portraitInput = filename
			}
		}()
		return
	}

	// Manual text input for file path (fallback for advanced users)
	cc.inputBuffer = ebiten.AppendInputChars(cc.inputBuffer[:0])
	for _, r := range cc.inputBuffer {
		// Allow printable characters for file paths
		if r >= 32 && r <= 126 {
			cc.portraitInput += string(r)
		}
	}

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(cc.portraitInput) > 0 {
			cc.portraitInput = cc.portraitInput[:len(cc.portraitInput)-1]
		} else {
			// Empty backspace goes back to class selection
			cc.currentStep = stepClassSelection
			// MOBILE/WASM: Ensure keyboard is hidden when going back
			// (it should already be hidden since we don't show it on portrait step now)
			if cc.keyboardShown && mobile.IsWASM() {
				mobile.HideKeyboard()
				cc.keyboardShown = false
			}
			return
		}
	}

	// ESC to go back
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepClassSelection
		// MOBILE/WASM: Ensure keyboard is hidden when cancelling
		if cc.keyboardShown && mobile.IsWASM() {
			mobile.HideKeyboard()
			cc.keyboardShown = false
		}
		return
	}

	// Tab to skip portrait (optional)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		cc.characterData.PortraitPath = ""
		cc.characterData.Portrait = nil
		cc.currentStep = stepConfirmation
		cc.errorMsg = ""
		// MOBILE/WASM: Ensure keyboard is hidden when skipping
		if cc.keyboardShown && mobile.IsWASM() {
			mobile.HideKeyboard()
			cc.keyboardShown = false
		}
		return
	}

	// Enter to load portrait and proceed
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		portraitPath := strings.TrimSpace(cc.portraitInput)

		if portraitPath == "" {
			// Empty path is valid (no portrait)
			cc.characterData.PortraitPath = ""
			cc.characterData.Portrait = nil
			cc.currentStep = stepConfirmation
			cc.errorMsg = ""
			// MOBILE/WASM: Ensure keyboard is hidden when completing
			if cc.keyboardShown && mobile.IsWASM() {
				mobile.HideKeyboard()
				cc.keyboardShown = false
			}
			return
		}

		// Try to load the portrait
		portrait, err := LoadPortrait(portraitPath)
		if err != nil {
			cc.errorMsg = fmt.Sprintf("Failed to load portrait: %v", err)
			return
		}

		// Success - save portrait and proceed
		cc.characterData.PortraitPath = portraitPath
		cc.characterData.Portrait = portrait
		cc.currentStep = stepConfirmation
		cc.errorMsg = ""
		// MOBILE/WASM: Ensure keyboard is hidden when completing
		if cc.keyboardShown && mobile.IsWASM() {
			mobile.HideKeyboard()
			cc.keyboardShown = false
		}
	}

	// F2 to save current portrait path as default
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		if len(strings.TrimSpace(cc.portraitInput)) > 0 {
			cc.defaults.DefaultPortraitPath = strings.TrimSpace(cc.portraitInput)
			cc.errorMsg = "Default portrait path saved!"
		}
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
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
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

	// Draw error message if present
	if cc.errorMsg != "" {
		errorX := cc.panelX + cc.panelWidth/2 - len(cc.errorMsg)*3
		errorY := cc.panelY + cc.panelHeight - 30
		text.Draw(screen, cc.errorMsg, basicfont.Face7x13, errorX, errorY,
			color.RGBA{255, 100, 100, 255})
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

	// Show current default if set
	if cc.defaults.DefaultName != "" {
		defaultText := fmt.Sprintf("Current default: %s", cc.defaults.DefaultName)
		defaultX := x + w/2 - len(defaultText)*3
		text.Draw(screen, defaultText, basicfont.Face7x13, defaultX, y+200,
			color.RGBA{150, 150, 150, 255})
	}

	// Help text
	helpText1 := "Press ENTER to continue | F2 to save as default"
	helpX1 := x + w/2 - len(helpText1)*3
	text.Draw(screen, helpText1, basicfont.Face7x13, helpX1, y+h-60,
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
	helpText3 := "Press ENTER to continue | F2 to save as default"
	helpText4 := "BACKSPACE to go back"
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
	if cc.defaults.DefaultPortraitPath != "" {
		cc.portraitInput = cc.defaults.DefaultPortraitPath
		// Try to load the default portrait
		if portrait, err := LoadPortrait(cc.defaults.DefaultPortraitPath); err == nil {
			cc.characterData.PortraitPath = cc.defaults.DefaultPortraitPath
			cc.characterData.Portrait = portrait
		}
	} else {
		cc.portraitInput = ""
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
func ApplyClassStats(player *Entity, class CharacterClass) error {
	if player == nil {
		return fmt.Errorf("player entity is nil")
	}

	// Get components
	healthComp, hasHealth := player.GetComponent("health")
	if !hasHealth {
		return fmt.Errorf("player missing health component")
	}

	manaComp, hasMana := player.GetComponent("mana")
	if !hasMana {
		return fmt.Errorf("player missing mana component")
	}

	statsComp, hasStats := player.GetComponent("stats")
	if !hasStats {
		return fmt.Errorf("player missing stats component")
	}

	attackComp, hasAttack := player.GetComponent("attack")
	if !hasAttack {
		return fmt.Errorf("player missing attack component")
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return fmt.Errorf("health component has wrong type")
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return fmt.Errorf("mana component has wrong type")
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return fmt.Errorf("stats component has wrong type")
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return fmt.Errorf("attack component has wrong type")
	}

	// Apply class-specific stats
	switch class {
	case ClassWarrior:
		health.Max = 150
		health.Current = 150
		mana.Max = 50
		mana.Current = 50
		stats.Attack = 12
		stats.Defense = 8
		attack.Damage = 20
		// Warriors get bonus crit damage
		stats.CritChance = 0.05
		stats.CritDamage = 2.0

	case ClassMage:
		health.Max = 80
		health.Current = 80
		mana.Max = 150
		mana.Current = 150
		mana.Regen = 8.0 // Faster mana regen
		stats.Attack = 6
		stats.Defense = 3
		attack.Damage = 10
		// Mages get bonus spell power (reflected in mana)
		stats.CritChance = 0.10 // Higher spell crit
		stats.CritDamage = 1.8

	case ClassRogue:
		health.Max = 100
		health.Current = 100
		mana.Max = 80
		mana.Current = 80
		stats.Attack = 10
		stats.Defense = 5
		attack.Damage = 15
		attack.Cooldown = 0.3 // Faster attacks
		// Rogues get high crit and evasion
		stats.CritChance = 0.15
		stats.CritDamage = 2.5
		stats.Evasion = 0.15

	default:
		return fmt.Errorf("unknown character class: %v", class)
	}

	return nil
}
