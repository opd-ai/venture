// Package engine provides single-player submenu UI for game mode selection.
// This file implements the submenu shown when the player selects "Single-Player"
// from the main menu, offering options for New Game, Load Game, and Back.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// SinglePlayerMenuOption represents a selectable option in the single-player submenu.
type SinglePlayerMenuOption int

const (
	// SinglePlayerMenuOptionNewGame starts a new single-player game.
	SinglePlayerMenuOptionNewGame SinglePlayerMenuOption = iota
	// SinglePlayerMenuOptionLoadGame loads a saved game (Phase 8.3).
	SinglePlayerMenuOptionLoadGame
	// SinglePlayerMenuOptionBack returns to the main menu.
	SinglePlayerMenuOptionBack
)

// String returns the display text for each submenu option.
func (o SinglePlayerMenuOption) String() string {
	switch o {
	case SinglePlayerMenuOptionNewGame:
		return "New Game"
	case SinglePlayerMenuOptionLoadGame:
		return "Load Game"
	case SinglePlayerMenuOptionBack:
		return "Back"
	default:
		return "Unknown"
	}
}

// SinglePlayerMenu renders and handles input for the single-player submenu.
// Follows the same patterns as MainMenuUI for consistency.
type SinglePlayerMenu struct {
	screenWidth  int
	screenHeight int
	selectedIdx  int
	options      []SinglePlayerMenuOption
	visible      bool

	// Callbacks for actions
	onNewGame  func()
	onLoadGame func()
	onBack     func()
}

// NewSinglePlayerMenu creates a new single-player submenu.
func NewSinglePlayerMenu(screenWidth, screenHeight int) *SinglePlayerMenu {
	return &SinglePlayerMenu{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		selectedIdx:  0,
		options: []SinglePlayerMenuOption{
			SinglePlayerMenuOptionNewGame,
			SinglePlayerMenuOptionLoadGame,
			SinglePlayerMenuOptionBack,
		},
		visible: false,
	}
}

// SetNewGameCallback sets the callback for starting a new game.
func (m *SinglePlayerMenu) SetNewGameCallback(callback func()) {
	m.onNewGame = callback
}

// SetLoadGameCallback sets the callback for loading a saved game.
func (m *SinglePlayerMenu) SetLoadGameCallback(callback func()) {
	m.onLoadGame = callback
}

// SetBackCallback sets the callback for returning to main menu.
func (m *SinglePlayerMenu) SetBackCallback(callback func()) {
	m.onBack = callback
}

// Show makes the submenu visible and resets selection to first option.
func (m *SinglePlayerMenu) Show() {
	m.visible = true
	m.selectedIdx = 0
}

// Hide makes the submenu invisible.
func (m *SinglePlayerMenu) Hide() {
	m.visible = false
}

// IsVisible returns whether the submenu is currently displayed.
func (m *SinglePlayerMenu) IsVisible() bool {
	return m.visible
}

// Update processes input for the single-player submenu.
// Returns true if an option was selected.
func (m *SinglePlayerMenu) Update() bool {
	if !m.visible {
		return false
	}

	// Handle ESC key for back navigation (dual-exit pattern)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if m.onBack != nil {
			m.onBack()
		}
		return true
	}

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

	// Handle number key shortcuts (1-3)
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		m.selectedIdx = 0
		return m.selectCurrentOption()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		m.selectedIdx = 1
		return m.selectCurrentOption()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		m.selectedIdx = 2
		return m.selectCurrentOption()
	}

	// Handle Enter/Space for selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return m.selectCurrentOption()
	}

	// Handle mouse input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if idx := m.getOptionAtPosition(mx, my); idx >= 0 {
			m.selectedIdx = idx
			return m.selectCurrentOption()
		}
	}

	// Update hover state for mouse
	mx, my := ebiten.CursorPosition()
	if idx := m.getOptionAtPosition(mx, my); idx >= 0 {
		m.selectedIdx = idx
	}

	return false
}

// selectCurrentOption triggers the callback for the currently selected option.
func (m *SinglePlayerMenu) selectCurrentOption() bool {
	option := m.options[m.selectedIdx]
	switch option {
	case SinglePlayerMenuOptionNewGame:
		if m.onNewGame != nil {
			m.onNewGame()
		}
		return true
	case SinglePlayerMenuOptionLoadGame:
		// Load Game is disabled until Phase 8.3 (Save/Load System)
		// No callback triggered for disabled option
		return false
	case SinglePlayerMenuOptionBack:
		if m.onBack != nil {
			m.onBack()
		}
		return true
	}
	return false
}

// getOptionAtPosition returns the option index at the given screen coordinates,
// or -1 if no option is at that position.
func (m *SinglePlayerMenu) getOptionAtPosition(x, y int) int {
	startY := m.screenHeight/2 - 60
	optionHeight := 40

	for i := range m.options {
		optionY := startY + i*optionHeight
		if y >= optionY && y < optionY+30 {
			// Check X bounds (centered text, approximate width)
			textWidth := len(m.options[i].String()) * 7
			startX := m.screenWidth/2 - textWidth/2
			if x >= startX-20 && x <= startX+textWidth+20 {
				return i
			}
		}
	}
	return -1
}

// Draw renders the single-player submenu to the screen.
func (m *SinglePlayerMenu) Draw(screen *ebiten.Image) {
	if !m.visible {
		return
	}

	// Draw semi-transparent background
	bgColor := color.RGBA{0, 0, 0, 200}
	ebitenutil.DrawRect(screen, 0, 0, float64(m.screenWidth), float64(m.screenHeight), bgColor)

	// Draw title
	titleText := "Single-Player"
	titleX := float64(m.screenWidth/2 - len(titleText)*7/2)
	titleY := float64(m.screenHeight/2 - 120)
	ebitenutil.DebugPrintAt(screen, titleText, int(titleX), int(titleY))

	// Draw subtitle
	subtitleText := "Select a game mode"
	subtitleX := float64(m.screenWidth/2 - len(subtitleText)*7/2)
	subtitleY := titleY + 25
	ebitenutil.DebugPrintAt(screen, subtitleText, int(subtitleX), int(subtitleY))

	// Draw options
	startY := m.screenHeight/2 - 60
	optionHeight := 40

	for i, option := range m.options {
		optionText := option.String()
		textWidth := len(optionText) * 7
		x := m.screenWidth/2 - textWidth/2
		y := startY + i*optionHeight

		// Check if option is disabled
		isDisabled := option == SinglePlayerMenuOptionLoadGame

		// Append "Coming Soon" text for disabled options
		if isDisabled {
			optionText = optionText + " (Coming Soon)"
		}

		// Draw selection indicator
		if i == m.selectedIdx && !isDisabled {
			indicatorText := ">"
			indicatorX := x - 20
			ebitenutil.DebugPrintAt(screen, indicatorText, indicatorX, y)
		}

		// Draw option text with color (approximation using DebugPrintAt)
		// Note: DebugPrintAt doesn't support colors, so we use position to indicate selection
		ebitenutil.DebugPrintAt(screen, optionText, x, y)

		// Draw hint text for disabled option
		if isDisabled && i == m.selectedIdx {
			hintText := "Save/Load system coming in Phase 8.3"
			hintX := m.screenWidth/2 - len(hintText)*7/2
			hintY := startY + len(m.options)*optionHeight + 20
			ebitenutil.DebugPrintAt(screen, hintText, hintX, hintY)
		}
	}

	// Draw controls hint
	hintText := "Arrow Keys/WASD: Navigate | Enter/Space: Select | ESC: Back | 1-3: Quick Select"
	hintX := m.screenWidth/2 - len(hintText)*7/2
	hintY := m.screenHeight - 40
	ebitenutil.DebugPrintAt(screen, hintText, hintX, hintY)
}
