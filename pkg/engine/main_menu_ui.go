// Package engine provides main menu UI components.
// This file implements the main menu screen with navigation and rendering.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MainMenuOption represents a selectable option in the main menu.
type MainMenuOption int

const (
	// MainMenuOptionSinglePlayer leads to single player submenu.
	MainMenuOptionSinglePlayer MainMenuOption = iota
	// MainMenuOptionMultiPlayer leads to multiplayer submenu.
	MainMenuOptionMultiPlayer
	// MainMenuOptionSettings leads to settings menu (future).
	MainMenuOptionSettings
	// MainMenuOptionQuit exits the application.
	MainMenuOptionQuit
)

// String returns the display text for each menu option.
func (o MainMenuOption) String() string {
	switch o {
	case MainMenuOptionSinglePlayer:
		return "Single-Player"
	case MainMenuOptionMultiPlayer:
		return "Multi-Player"
	case MainMenuOptionSettings:
		return "Settings"
	case MainMenuOptionQuit:
		return "Quit"
	default:
		return "Unknown"
	}
}

// MainMenuUI renders and handles input for the main menu screen.
// Uses simple keyboard navigation (up/down arrows, enter to select).
type MainMenuUI struct {
	screenWidth  int
	screenHeight int
	selectedIdx  int
	options      []MainMenuOption

	// Callback for when an option is selected
	onSelect func(option MainMenuOption)
}

// NewMainMenuUI creates a new main menu UI.
func NewMainMenuUI(screenWidth, screenHeight int) *MainMenuUI {
	return &MainMenuUI{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		selectedIdx:  0,
		options: []MainMenuOption{
			MainMenuOptionSinglePlayer,
			MainMenuOptionMultiPlayer,
			MainMenuOptionSettings,
			MainMenuOptionQuit,
		},
	}
}

// SetSelectCallback sets the callback function called when an option is selected.
func (m *MainMenuUI) SetSelectCallback(callback func(option MainMenuOption)) {
	m.onSelect = callback
}

// Update processes input for the main menu.
// Returns true if an option was selected.
func (m *MainMenuUI) Update() bool {
	// Handle up/down arrow keys for navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedIdx--
		if m.selectedIdx < 0 {
			m.selectedIdx = len(m.options) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedIdx++
		if m.selectedIdx >= len(m.options) {
			m.selectedIdx = 0
		}
	}

	// Handle Enter key for selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if m.onSelect != nil {
			m.onSelect(m.options[m.selectedIdx])
		}
		return true
	}

	// Handle mouse input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		if clickedOption := m.getOptionAtPosition(mouseX, mouseY); clickedOption != -1 {
			m.selectedIdx = clickedOption
			if m.onSelect != nil {
				m.onSelect(m.options[m.selectedIdx])
			}
			return true
		}
	}

	// Update selection highlight on mouse hover
	mouseX, mouseY := ebiten.CursorPosition()
	if hoverOption := m.getOptionAtPosition(mouseX, mouseY); hoverOption != -1 {
		m.selectedIdx = hoverOption
	}

	return false
}

// Draw renders the main menu to the screen.
func (m *MainMenuUI) Draw(screen *ebiten.Image) {
	// Draw title
	titleText := "VENTURE"
	titleX := float64(m.screenWidth/2 - len(titleText)*8)
	titleY := float64(m.screenHeight / 4)
	ebitenutil.DebugPrintAt(screen, titleText, int(titleX), int(titleY))

	// Draw subtitle
	subtitleText := "Procedural Action RPG"
	subtitleX := float64(m.screenWidth/2 - len(subtitleText)*4)
	subtitleY := titleY + 30
	ebitenutil.DebugPrintAt(screen, subtitleText, int(subtitleX), int(subtitleY))

	// Draw menu options
	startY := m.screenHeight / 2
	spacing := 40

	for i, option := range m.options {
		optionText := option.String()
		x := m.screenWidth/2 - len(optionText)*4
		y := startY + i*spacing

		// Draw selection indicator
		if i == m.selectedIdx {
			// Draw selection box
			selectionColor := color.RGBA{100, 100, 200, 128}
			boxWidth := len(optionText)*8 + 20
			boxHeight := 25
			boxX := x - 10
			boxY := y - 5

			ebitenutil.DrawRect(screen, float64(boxX), float64(boxY), float64(boxWidth), float64(boxHeight), selectionColor)

			// Draw arrow indicator
			ebitenutil.DebugPrintAt(screen, ">", x-20, y)
		}

		// Draw option text
		ebitenutil.DebugPrintAt(screen, optionText, x, y)
	}

	// Draw controls hint
	controlsText := "Use Arrow Keys / WASD to navigate, Enter/Space to select"
	controlsX := m.screenWidth/2 - len(controlsText)*3
	controlsY := m.screenHeight - 50
	ebitenutil.DebugPrintAt(screen, controlsText, controlsX, controlsY)
}

// getOptionAtPosition returns the index of the menu option at the given screen position.
// Returns -1 if no option is at that position.
func (m *MainMenuUI) getOptionAtPosition(x, y int) int {
	startY := m.screenHeight / 2
	spacing := 40

	for i, option := range m.options {
		optionText := option.String()
		optionX := m.screenWidth/2 - len(optionText)*4
		optionY := startY + i*spacing

		// Check if mouse is within option bounds
		boxWidth := len(optionText)*8 + 20
		boxHeight := 25
		boxX := optionX - 10
		boxY := optionY - 5

		if x >= boxX && x <= boxX+boxWidth && y >= boxY && y <= boxY+boxHeight {
			return i
		}
	}

	return -1
}

// GetSelectedOption returns the currently selected option.
func (m *MainMenuUI) GetSelectedOption() MainMenuOption {
	return m.options[m.selectedIdx]
}

// Reset resets the menu to its initial state (first option selected).
func (m *MainMenuUI) Reset() {
	m.selectedIdx = 0
}
