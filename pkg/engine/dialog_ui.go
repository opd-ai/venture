//go:build !headless
// +build !headless

package engine

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/procgen/dialog"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// DialogUIState represents the current state of the dialog UI.
type DialogUIState int

const (
	DialogUIIdle         DialogUIState = iota // No dialog active
	DialogUIGreeting                          // Showing NPC greeting
	DialogUIConversation                      // Active conversation
	DialogUIOptions                           // Showing response options
	DialogUIEnding                            // Dialog ending
)

// DialogUI manages the interactive dialog interface for NPC conversations.
// It displays NPC dialog, player response options, and conversation history.
type DialogUI struct {
	world          *World
	player         *Entity
	dialogSystem   *DialogSystem
	npcDialogSys   *NPCDialogSystem
	screenWidth    int
	screenHeight   int
	visible        bool
	state          DialogUIState
	currentNPCID   uint64
	currentNPCName string
	npcText        string
	playerOptions  []string
	selectedOption int
	history        []string // Recent conversation history
	maxHistory     int
}

// NewDialogUI creates a new dialog UI.
func NewDialogUI(world *World, dialogSystem *DialogSystem, npcDialogSys *NPCDialogSystem, screenWidth, screenHeight int) *DialogUI {
	return &DialogUI{
		world:        world,
		dialogSystem: dialogSystem,
		npcDialogSys: npcDialogSys,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		visible:      false,
		state:        DialogUIIdle,
		maxHistory:   5,
		history:      make([]string, 0, 5),
	}
}

// SetPlayerEntity sets the player entity for the dialog UI.
func (ui *DialogUI) SetPlayerEntity(player *Entity) {
	ui.player = player
}

// IsVisible returns whether the dialog UI is visible.
func (ui *DialogUI) IsVisible() bool {
	return ui.visible
}

// Show displays the dialog UI and starts a conversation with an NPC.
func (ui *DialogUI) Show(npcID uint64) error {
	if ui.player == nil {
		return fmt.Errorf("player entity not set")
	}

	// Get NPC entity
	npc, ok := ui.world.GetEntity(npcID)
	if !ok {
		return fmt.Errorf("NPC entity %d not found", npcID)
	}

	// Get NPC name - use simple approach
	ui.currentNPCName = fmt.Sprintf("NPC #%d", npcID)

	// Start dialog
	success, err := ui.dialogSystem.StartDialog(ui.player.ID, npcID)
	if !success || err != nil {
		return fmt.Errorf("failed to start dialog: %w", err)
	}

	ui.currentNPCID = npcID
	ui.visible = true
	ui.state = DialogUIGreeting
	ui.history = make([]string, 0, ui.maxHistory)

	// Generate greeting
	if err := ui.generateGreeting(npc); err != nil {
		return fmt.Errorf("failed to generate greeting: %w", err)
	}

	ui.setupInitialOptions()

	return nil
}

// Hide closes the dialog UI.
func (ui *DialogUI) Hide() {
	if ui.visible && ui.state != DialogUIIdle {
		ui.dialogSystem.EndDialog()
	}
	ui.visible = false
	ui.state = DialogUIIdle
	ui.currentNPCID = 0
	ui.npcText = ""
	ui.playerOptions = nil
	ui.selectedOption = 0
	ui.history = make([]string, 0, ui.maxHistory)
}

// Update processes input for the dialog UI.
func (ui *DialogUI) Update() error {
	if !ui.visible {
		return nil
	}

	// Handle touch/mouse input for option selection
	if IsTouchOrMouseJustPressed() {
		mouseX, mouseY, _ := GetTouchOrMousePosition()
		// Check if touch/click is within options area
		optionY := ui.screenHeight - 250 + 120
		panelX := ui.screenWidth/2 - 300
		for i := range ui.playerOptions {
			optY := optionY + i*25
			// Check if within option bounds (approximate hit area)
			if mouseX >= panelX && mouseX <= panelX+580 &&
				mouseY >= optY-12 && mouseY <= optY+12 {
				ui.selectedOption = i
				return ui.handleOptionSelect()
			}
		}
	}

	// Keyboard navigation - Platform parity fix: Use edge-triggered detection
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if len(ui.playerOptions) > 0 {
			ui.selectedOption = (ui.selectedOption + 1) % len(ui.playerOptions)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if len(ui.playerOptions) > 0 {
			ui.selectedOption = (ui.selectedOption - 1 + len(ui.playerOptions)) % len(ui.playerOptions)
		}
	}

	// Select option - Platform parity fix: Use edge-triggered detection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return ui.handleOptionSelect()
	}

	// Close dialog - Platform parity fix: Use edge-triggered detection
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Hide()
		return nil
	}

	return nil
}

// Draw renders the dialog UI.
func (ui *DialogUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Dialog panel dimensions
	panelX := float32(ui.screenWidth/2 - 300)
	panelY := float32(ui.screenHeight - 250)
	panelWidth := float32(600)
	panelHeight := float32(220)

	// Draw semi-transparent background overlay
	vector.DrawFilledRect(screen, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 128}, false)

	// Draw dialog panel background
	vector.DrawFilledRect(screen, panelX, panelY, panelWidth, panelHeight,
		color.RGBA{40, 40, 50, 240}, false)

	// Draw panel border
	vector.StrokeRect(screen, panelX, panelY, panelWidth, panelHeight, 2,
		color.RGBA{100, 100, 150, 255}, false)

	// Get font
	fnt := basicfont.Face7x13

	// Draw NPC name
	nameText := ui.currentNPCName
	nameColor := color.RGBA{255, 220, 100, 255}
	text.Draw(screen, nameText, fnt, int(panelX+10), int(panelY+20), nameColor)

	// Draw NPC text (wrapped)
	ui.drawWrappedText(screen, ui.npcText, fnt, int(panelX+10), int(panelY+45), int(panelWidth-20), color.White)

	// Draw player options
	optionY := int(panelY + 120)
	for i, option := range ui.playerOptions {
		optionColor := color.RGBA{200, 200, 200, 255}
		prefix := "  "
		if i == ui.selectedOption {
			optionColor = color.RGBA{255, 255, 100, 255}
			prefix = "> "
		}
		optionText := prefix + option
		text.Draw(screen, optionText, fnt, int(panelX+15), optionY+i*25, optionColor)
	}

	// Draw close hint
	hintText := "[ESC] Close"
	hintColor := color.RGBA{150, 150, 150, 255}
	text.Draw(screen, hintText, fnt, int(panelX+panelWidth-120), int(panelY+20), hintColor)

	// Draw conversation history (if any)
	if len(ui.history) > 0 {
		historyY := int(panelY - 150)
		for i, line := range ui.history {
			historyColor := color.RGBA{180, 180, 180, 200}
			text.Draw(screen, line, fnt, int(panelX+10), historyY+i*20, historyColor)
		}
	}
}

// generateGreeting generates the NPC's initial greeting.
func (ui *DialogUI) generateGreeting(npc *Entity) error {
	// Get NPC dialog component
	dialogComp, ok := npc.GetComponent("npcdialog")
	if !ok {
		// Fallback greeting
		ui.npcText = "Hello, traveler."
		return nil
	}

	npcDialog, ok := dialogComp.(*NPCDialogComponent)
	if !ok {
		ui.npcText = "Greetings."
		return nil
	}

	// Use NPC personality or default
	personality := npcDialog.NPCPersonality
	if personality == nil {
		personality = dialog.NewPersonality(dialog.PersonalityHelpful)
	}

	// Generate greeting using personality
	// Use fantasy as default genre for greeting
	greeting := personality.GetGreeting("fantasy")
	ui.npcText = greeting

	return nil
}

// setupInitialOptions sets up the initial player response options.
func (ui *DialogUI) setupInitialOptions() {
	ui.playerOptions = []string{
		"Tell me about yourself",
		"What do you know about this place?",
		"Goodbye",
	}
	ui.selectedOption = 0
	ui.state = DialogUIOptions
}

// handleOptionSelect processes the player's option selection.
func (ui *DialogUI) handleOptionSelect() error {
	if len(ui.playerOptions) == 0 {
		return nil
	}

	selectedText := ui.playerOptions[ui.selectedOption]

	// Add to history
	ui.addToHistory(fmt.Sprintf("You: %s", selectedText))

	// Check for goodbye
	if strings.Contains(strings.ToLower(selectedText), "goodbye") || strings.Contains(strings.ToLower(selectedText), "farewell") {
		ui.npcText = "Farewell, traveler. Safe travels."
		ui.state = DialogUIEnding
		ui.playerOptions = []string{"[Close]"}
		ui.selectedOption = 0
		return nil
	}

	// Generate NPC response
	npc, ok := ui.world.GetEntity(ui.currentNPCID)
	if !ok {
		return fmt.Errorf("NPC entity no longer exists")
	}

	response, err := ui.npcDialogSys.GenerateResponse(npc, selectedText)
	if err != nil {
		// Fallback response
		response = "I see..."
	}

	ui.npcText = response
	ui.addToHistory(fmt.Sprintf("%s: %s", ui.currentNPCName, response))

	// Setup next options
	if ui.state == DialogUIEnding {
		ui.Hide()
	} else {
		ui.setupConversationOptions()
	}

	return nil
}

// setupConversationOptions sets up options for ongoing conversation.
func (ui *DialogUI) setupConversationOptions() {
	ui.playerOptions = []string{
		"Tell me more",
		"What else?",
		"That's interesting",
		"Goodbye",
	}
	ui.selectedOption = 0
	ui.state = DialogUIConversation
}

// addToHistory adds a line to the conversation history.
func (ui *DialogUI) addToHistory(line string) {
	ui.history = append(ui.history, line)
	if len(ui.history) > ui.maxHistory {
		ui.history = ui.history[len(ui.history)-ui.maxHistory:]
	}
}

// drawWrappedText draws text with word wrapping.
func (ui *DialogUI) drawWrappedText(screen *ebiten.Image, txt string, fnt font.Face, x, y, maxWidth int, col color.Color) {
	words := strings.Fields(txt)
	line := ""
	lineY := y

	for _, word := range words {
		testLine := line + word + " "
		bounds := text.BoundString(fnt, testLine)
		if bounds.Dx() > maxWidth && line != "" {
			text.Draw(screen, line, fnt, x, lineY, col)
			line = word + " "
			lineY += 20
		} else {
			line = testLine
		}
	}
	if line != "" {
		text.Draw(screen, line, fnt, x, lineY, col)
	}
}
