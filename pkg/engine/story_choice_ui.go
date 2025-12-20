// Package engine provides UI for branching narrative story choices.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/narrative/branching"
	"golang.org/x/image/font/basicfont"
)

// StoryChoiceUI displays branching narrative choices to the player.
type StoryChoiceUI struct {
	world           *World
	playerEntity    *Entity
	narrativeSystem *BranchingNarrativeSystem
	visible         bool
	selectedChoice  int
	screenWidth     int
	screenHeight    int
	currentNode     *branching.StoryNode
	pendingChoices  []branching.Choice
	lastCheckTime   float64
	backgroundColor color.Color
	textColor       color.Color
	selectedColor   color.Color
	titleColor      color.Color
}

// NewStoryChoiceUI creates a new story choice UI.
func NewStoryChoiceUI(world *World, narrativeSystem *BranchingNarrativeSystem, screenWidth, screenHeight int) *StoryChoiceUI {
	return &StoryChoiceUI{
		world:           world,
		narrativeSystem: narrativeSystem,
		visible:         false,
		selectedChoice:  0,
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		backgroundColor: color.RGBA{20, 20, 30, 230},
		textColor:       color.RGBA{200, 200, 200, 255},
		selectedColor:   color.RGBA{255, 215, 0, 255},
		titleColor:      color.RGBA{255, 255, 255, 255},
	}
}

// SetPlayerEntity sets the player entity for the UI.
func (ui *StoryChoiceUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Update checks for pending choices and updates UI state.
func (ui *StoryChoiceUI) Update(deltaTime float64) {
	if ui.playerEntity == nil || ui.narrativeSystem == nil {
		return
	}

	ui.lastCheckTime += deltaTime
	// Check for new choices every 0.5 seconds
	if ui.lastCheckTime < 0.5 {
		return
	}
	ui.lastCheckTime = 0

	// Get current node
	comp, ok := ui.playerEntity.GetComponent("branching_narrative")
	if !ok {
		ui.visible = false
		return
	}

	narComp, ok := comp.(*BranchingNarrativeComponent)
	if !ok {
		ui.visible = false
		return
	}

	// Check for pending choices
	if len(narComp.PendingChoices) > 0 {
		node, err := ui.narrativeSystem.GetCurrentNode(ui.playerEntity)
		if err == nil && node != nil {
			ui.currentNode = node
			ui.pendingChoices = narComp.PendingChoices
			ui.visible = true
			ui.selectedChoice = 0 // Reset selection
		}
	} else {
		ui.visible = false
	}
}

// Draw renders the story choice UI.
func (ui *StoryChoiceUI) Draw(screen *ebiten.Image) {
	if !ui.visible || ui.currentNode == nil || len(ui.pendingChoices) == 0 {
		return
	}

	// Calculate UI dimensions
	panelWidth := float32(ui.screenWidth) * 0.8
	panelHeight := float32(ui.screenHeight) * 0.6
	panelX := (float32(ui.screenWidth) - panelWidth) / 2
	panelY := (float32(ui.screenHeight) - panelHeight) / 2

	// Draw background panel
	vector.DrawFilledRect(screen, panelX, panelY, panelWidth, panelHeight, ui.backgroundColor, false)

	// Draw border
	borderColor := color.RGBA{100, 100, 120, 255}
	vector.StrokeRect(screen, panelX, panelY, panelWidth, panelHeight, 2, borderColor, false)

	// Draw title
	titleY := int(panelY) + 20
	titleX := int(panelX) + 20
	text.Draw(screen, ui.currentNode.Title, basicfont.Face7x13, titleX, titleY, ui.titleColor)

	// Draw description
	descY := titleY + 30
	text.Draw(screen, ui.currentNode.Description, basicfont.Face7x13, titleX, descY, ui.textColor)

	// Draw choices
	choiceStartY := descY + 40
	for i, choice := range ui.pendingChoices {
		choiceY := choiceStartY + i*30
		choiceText := choice.Text

		// Highlight selected choice
		textCol := ui.textColor
		if i == ui.selectedChoice {
			choiceText = "> " + choiceText
			textCol = ui.selectedColor
		}

		text.Draw(screen, choiceText, basicfont.Face7x13, titleX, choiceY, textCol)
	}

	// Draw controls help
	helpY := int(panelY+panelHeight) - 30
	helpText := "UP/DOWN: Select | ENTER: Confirm | ESC: Cancel"
	text.Draw(screen, helpText, basicfont.Face7x13, titleX, helpY, ui.textColor)
}

// HandleInput processes keyboard and touch/mouse input for choice selection.
// Returns true if input was handled.
// Platform parity fix: Uses edge-triggered detection and touch support.
func (ui *StoryChoiceUI) HandleInput() bool {
	if !ui.visible || len(ui.pendingChoices) == 0 {
		return false
	}

	// Handle touch/mouse input for choice selection
	if IsTouchOrMouseJustPressed() {
		mouseX, mouseY, _ := GetTouchOrMousePosition()
		// Calculate choice area position to match Draw() layout
		panelWidth := float64(ui.screenWidth) * 0.8
		panelXFloat := (float64(ui.screenWidth) - panelWidth) / 2
		titleX := int(panelXFloat) + 20

		panelYFloat := float64(ui.screenHeight) * 0.2
		titleY := int(panelYFloat) + 20
		choiceStartY := titleY + 30 + 40

		for i := range ui.pendingChoices {
			choiceY := choiceStartY + i*30
			// Check if within choice bounds (approximate hit area)
			if mouseX >= titleX && mouseX <= titleX+int(panelWidth)-40 &&
				mouseY >= choiceY-12 && mouseY <= choiceY+12 {
				ui.selectedChoice = i
				ui.confirmCurrentChoice()
				return true
			}
		}
	}

	// Navigate choices - Platform parity fix: Use edge-triggered detection
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		ui.selectedChoice--
		if ui.selectedChoice < 0 {
			ui.selectedChoice = len(ui.pendingChoices) - 1
		}
		return true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		ui.selectedChoice++
		if ui.selectedChoice >= len(ui.pendingChoices) {
			ui.selectedChoice = 0
		}
		return true
	}

	// Confirm choice - Platform parity fix: Use edge-triggered detection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		ui.confirmCurrentChoice()
		return true
	}

	// Cancel (hide UI without making choice - player can come back later)
	// Platform parity fix: Use edge-triggered detection
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.visible = false
		return true
	}

	return false
}

// confirmCurrentChoice confirms the currently selected choice.
// It applies the selected choice via the narrative system and hides the UI if successful.
func (ui *StoryChoiceUI) confirmCurrentChoice() {
	if ui.selectedChoice >= 0 && ui.selectedChoice < len(ui.pendingChoices) {
		choice := ui.pendingChoices[ui.selectedChoice]
		if err := ui.narrativeSystem.MakeChoice(ui.playerEntity, choice.ID); err == nil {
			ui.visible = false
			ui.currentNode = nil
			ui.pendingChoices = nil
		}
	}
}

// IsVisible returns whether the UI is currently visible.
func (ui *StoryChoiceUI) IsVisible() bool {
	return ui.visible
}

// Show forces the UI to be visible (used for testing or manual triggers).
func (ui *StoryChoiceUI) Show() {
	if ui.playerEntity != nil {
		ui.Update(1.0) // Force a check for choices
	}
}

// Hide forces the UI to be hidden.
func (ui *StoryChoiceUI) Hide() {
	ui.visible = false
}
