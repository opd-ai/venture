// Package engine provides character creation functionality for onboarding new players.
// This file implements the character creation UI and class selection system that
// integrates with the tutorial flow for a unified onboarding experience.
package engine

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

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
	Name  string
	Class CharacterClass
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

// creationStep represents the current step in character creation
type creationStep int

const (
	stepNameInput creationStep = iota
	stepClassSelection
	stepConfirmation
)

// EbitenCharacterCreation handles the character creation UI and flow
type EbitenCharacterCreation struct {
	currentStep   creationStep
	characterData CharacterData
	nameInput     string
	selectedClass CharacterClass
	confirmed     bool
	errorMsg      string

	// Input state
	inputBuffer []rune

	screenWidth  int
	screenHeight int
}

// NewCharacterCreation creates a new character creation system
func NewCharacterCreation(screenWidth, screenHeight int) *EbitenCharacterCreation {
	return &EbitenCharacterCreation{
		currentStep:   stepNameInput,
		selectedClass: ClassWarrior, // Default selection
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		inputBuffer:   make([]rune, 0),
	}
}

// Update handles input for character creation (keyboard/mouse navigation)
// Returns true when character creation is complete
func (cc *EbitenCharacterCreation) Update() bool {
	switch cc.currentStep {
	case stepNameInput:
		cc.updateNameInput()
	case stepClassSelection:
		cc.updateClassSelection()
	case stepConfirmation:
		cc.updateConfirmation()
	}

	return cc.confirmed
}

// updateNameInput handles name input with keyboard
func (cc *EbitenCharacterCreation) updateNameInput() {
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
		} else {
			cc.errorMsg = "Name cannot be empty"
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

	// Enter to proceed, Backspace to go back
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		cc.characterData.Class = cc.selectedClass
		cc.currentStep = stepConfirmation
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepNameInput
	}
}

// updateConfirmation handles final confirmation
func (cc *EbitenCharacterCreation) updateConfirmation() {
	// Enter/Space to confirm
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		// Validate before confirming
		if err := cc.characterData.Validate(); err != nil {
			cc.errorMsg = err.Error()
			cc.currentStep = stepNameInput // Go back to fix
		} else {
			cc.confirmed = true
		}
	}

	// Backspace to go back
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.currentStep = stepClassSelection
	}
}

// Draw renders the character creation UI
func (cc *EbitenCharacterCreation) Draw(screen *ebiten.Image) {
	// Draw semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(cc.screenWidth), float32(cc.screenHeight),
		color.RGBA{0, 0, 0, 200}, false)

	// Calculate panel dimensions
	panelWidth := 600
	panelHeight := 400
	panelX := cc.screenWidth/2 - panelWidth/2
	panelY := cc.screenHeight/2 - panelHeight/2

	// Draw panel background
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 30, 255}, false)

	// Draw panel border
	vector.StrokeRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)

	// Draw content based on current step
	switch cc.currentStep {
	case stepNameInput:
		cc.drawNameInput(screen, panelX, panelY, panelWidth, panelHeight)
	case stepClassSelection:
		cc.drawClassSelection(screen, panelX, panelY, panelWidth, panelHeight)
	case stepConfirmation:
		cc.drawConfirmation(screen, panelX, panelY, panelWidth, panelHeight)
	}

	// Draw error message if present
	if cc.errorMsg != "" {
		errorX := panelX + panelWidth/2 - len(cc.errorMsg)*3
		errorY := panelY + panelHeight - 30
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
	stepText := "Step 1 of 3: Choose Your Name"
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

	// Help text
	helpText := "Press ENTER to continue"
	helpX := x + w/2 - len(helpText)*3
	text.Draw(screen, helpText, basicfont.Face7x13, helpX, y+h-60,
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
	stepText := "Step 2 of 3: Choose Your Class"
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

	// Help text
	helpText1 := "Use ARROW KEYS or 1-3 to select"
	helpText2 := "Press ENTER to continue | BACKSPACE to go back"
	helpX1 := x + w/2 - len(helpText1)*3
	helpX2 := x + w/2 - len(helpText2)*3
	text.Draw(screen, helpText1, basicfont.Face7x13, helpX1, y+h-75,
		color.RGBA{150, 200, 150, 255})
	text.Draw(screen, helpText2, basicfont.Face7x13, helpX2, y+h-55,
		color.RGBA{150, 200, 150, 255})
}

// drawConfirmation renders the confirmation screen
func (cc *EbitenCharacterCreation) drawConfirmation(screen *ebiten.Image, x, y, w, h int) {
	// Title
	title := "CHARACTER CREATION"
	titleX := x + w/2 - len(title)*3
	text.Draw(screen, title, basicfont.Face7x13, titleX, y+40,
		color.RGBA{255, 255, 100, 255})

	// Step indicator
	stepText := "Step 3 of 3: Confirm Your Character"
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

	// Class stats preview
	statsY := summaryY + 80
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

	// Help text
	helpText1 := "Press ENTER to begin your adventure"
	helpText2 := "Press BACKSPACE to change class"
	helpX1 := x + w/2 - len(helpText1)*3
	helpX2 := x + w/2 - len(helpText2)*3
	text.Draw(screen, helpText1, basicfont.Face7x13, helpX1, y+h-75,
		color.RGBA{100, 255, 100, 255})
	text.Draw(screen, helpText2, basicfont.Face7x13, helpX2, y+h-55,
		color.RGBA{150, 200, 150, 255})
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

// Reset resets the character creation to initial state
func (cc *EbitenCharacterCreation) Reset() {
	cc.currentStep = stepNameInput
	cc.characterData = CharacterData{}
	cc.nameInput = ""
	cc.selectedClass = ClassWarrior
	cc.confirmed = false
	cc.errorMsg = ""
}

// wrapText splits text into lines of approximately maxChars length
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

	health := healthComp.(*HealthComponent)
	mana := manaComp.(*ManaComponent)
	stats := statsComp.(*StatsComponent)
	attack := attackComp.(*AttackComponent)

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
