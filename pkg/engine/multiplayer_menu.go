package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// MultiplayerMenuOption represents an option in the multiplayer menu.
type MultiplayerMenuOption int

const (
	// MultiplayerMenuOptionJoin allows player to input server address and connect.
	MultiplayerMenuOptionJoin MultiplayerMenuOption = iota
	// MultiplayerMenuOptionHost starts a local server and auto-connects.
	MultiplayerMenuOptionHost
	// MultiplayerMenuOptionBack returns to the main menu.
	MultiplayerMenuOptionBack
)

// String returns the display name for the multiplayer menu option.
func (m MultiplayerMenuOption) String() string {
	switch m {
	case MultiplayerMenuOptionJoin:
		return "Join Server"
	case MultiplayerMenuOptionHost:
		return "Host Game"
	case MultiplayerMenuOptionBack:
		return "Back"
	default:
		return "Unknown"
	}
}

// MultiplayerMenu represents the multiplayer submenu UI.
type MultiplayerMenu struct {
	screenWidth    int
	screenHeight   int
	selectedIndex  int
	options        []MultiplayerMenuOption
	isVisible      bool
	onJoin         func()
	onHost         func()
	onBack         func()
	mouseX         int
	mouseY         int
	lastPressedKey ebiten.Key
}

// NewMultiplayerMenu creates a new multiplayer menu with the given screen dimensions.
func NewMultiplayerMenu(screenWidth, screenHeight int) *MultiplayerMenu {
	return &MultiplayerMenu{
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		selectedIndex: 0,
		options: []MultiplayerMenuOption{
			MultiplayerMenuOptionJoin,
			MultiplayerMenuOptionHost,
			MultiplayerMenuOptionBack,
		},
		isVisible: false,
	}
}

// Show makes the multiplayer menu visible.
func (m *MultiplayerMenu) Show() {
	m.isVisible = true
	m.selectedIndex = 0
}

// Hide makes the multiplayer menu invisible.
func (m *MultiplayerMenu) Hide() {
	m.isVisible = false
}

// IsVisible returns whether the multiplayer menu is currently visible.
func (m *MultiplayerMenu) IsVisible() bool {
	return m.isVisible
}

// SetJoinCallback sets the callback function for the Join option.
func (m *MultiplayerMenu) SetJoinCallback(callback func()) {
	m.onJoin = callback
}

// SetHostCallback sets the callback function for the Host option.
func (m *MultiplayerMenu) SetHostCallback(callback func()) {
	m.onHost = callback
}

// SetBackCallback sets the callback function for the Back option.
func (m *MultiplayerMenu) SetBackCallback(callback func()) {
	m.onBack = callback
}

// Update handles input for the multiplayer menu.
func (m *MultiplayerMenu) Update() {
	if !m.isVisible {
		return
	}

	// Track mouse position
	m.mouseX, m.mouseY = ebiten.CursorPosition()

	// Keyboard navigation - Arrow keys
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedIndex--
		if m.selectedIndex < 0 {
			m.selectedIndex = len(m.options) - 1 // Wrap to bottom
		}
		m.lastPressedKey = ebiten.KeyUp
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedIndex++
		if m.selectedIndex >= len(m.options) {
			m.selectedIndex = 0 // Wrap to top
		}
		m.lastPressedKey = ebiten.KeyDown
	}

	// Number shortcuts (1-3)
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		m.selectedIndex = 0
		m.selectCurrentOption()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		m.selectedIndex = 1
		m.selectCurrentOption()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		m.selectedIndex = 2
		m.selectCurrentOption()
	}

	// Enter or Space to select
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.selectCurrentOption()
	}

	// ESC to go back (dual-exit pattern)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if m.onBack != nil {
			m.onBack()
		}
	}

	// Mouse hover detection
	if mouseIndex := m.getOptionAtPosition(m.mouseX, m.mouseY); mouseIndex >= 0 {
		m.selectedIndex = mouseIndex
	}

	// Mouse click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if clickIndex := m.getOptionAtPosition(m.mouseX, m.mouseY); clickIndex >= 0 {
			m.selectedIndex = clickIndex
			m.selectCurrentOption()
		}
	}
}

// selectCurrentOption triggers the callback for the currently selected option.
func (m *MultiplayerMenu) selectCurrentOption() {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.options) {
		return
	}

	option := m.options[m.selectedIndex]
	switch option {
	case MultiplayerMenuOptionJoin:
		if m.onJoin != nil {
			m.onJoin()
		}
	case MultiplayerMenuOptionHost:
		if m.onHost != nil {
			m.onHost()
		}
	case MultiplayerMenuOptionBack:
		if m.onBack != nil {
			m.onBack()
		}
	}
}

// getOptionAtPosition returns the index of the option at the given screen position,
// or -1 if no option is at that position.
func (m *MultiplayerMenu) getOptionAtPosition(x, y int) int {
	// Calculate menu layout (same as Draw)
	titleText := "Multiplayer"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleHeight := titleBounds.Dy()

	startY := m.screenHeight/2 - 50
	optionHeight := 50

	// Check each option's bounding box
	for i := range m.options {
		optionY := startY + i*optionHeight
		optionText := m.options[i].String()
		
		// Add number shortcut to text
		displayText := ""
		if i < 9 {
			displayText = string(rune('1'+i)) + ". " + optionText
		} else {
			displayText = optionText
		}
		
		bounds := text.BoundString(basicfont.Face7x13, displayText)
		textWidth := bounds.Dx()
		textHeight := bounds.Dy()

		// Calculate text position (centered)
		textX := m.screenWidth/2 - textWidth/2
		textY := optionY + textHeight/2

		// Create hit box around text (with some padding)
		padding := 10
		hitBoxLeft := textX - padding
		hitBoxRight := textX + textWidth + padding
		hitBoxTop := textY - textHeight - padding
		hitBoxBottom := textY + padding

		if x >= hitBoxLeft && x <= hitBoxRight && y >= hitBoxTop && y <= hitBoxBottom {
			return i
		}
	}

	// Account for title space
	_ = titleHeight

	return -1
}

// Draw renders the multiplayer menu to the screen.
func (m *MultiplayerMenu) Draw(screen *ebiten.Image) {
	if !m.isVisible || screen == nil {
		return
	}

	// Draw title
	titleText := "Multiplayer"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleWidth := titleBounds.Dx()
	titleX := m.screenWidth/2 - titleWidth/2
	titleY := m.screenHeight/2 - 150
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY, color.White)

	// Draw options
	startY := m.screenHeight/2 - 50
	optionHeight := 50

	for i, option := range m.options {
		optionY := startY + i*optionHeight

		// Add number shortcut to display text
		displayText := ""
		if i < 9 {
			displayText = string(rune('1'+i)) + ". " + option.String()
		} else {
			displayText = option.String()
		}

		bounds := text.BoundString(basicfont.Face7x13, displayText)
		textWidth := bounds.Dx()
		textHeight := bounds.Dy()
		textX := m.screenWidth/2 - textWidth/2
		textY := optionY + textHeight/2

		// Highlight selected option
		optionColor := color.RGBA{180, 180, 180, 255}
		if i == m.selectedIndex {
			optionColor = color.RGBA{255, 255, 100, 255} // Yellow highlight
		}

		text.Draw(screen, displayText, basicfont.Face7x13, textX, textY, optionColor)
	}

	// Draw controls hint at bottom
	hintText := "[1-3] Quick Select  [↑↓/WS] Navigate  [Enter] Select  [Esc] Back"
	hintBounds := text.BoundString(basicfont.Face7x13, hintText)
	hintWidth := hintBounds.Dx()
	hintX := m.screenWidth/2 - hintWidth/2
	hintY := m.screenHeight - 30
	hintColor := color.RGBA{100, 100, 100, 255}
	text.Draw(screen, hintText, basicfont.Face7x13, hintX, hintY, hintColor)
}

// GetSelectedOption returns the currently selected option.
func (m *MultiplayerMenu) GetSelectedOption() MultiplayerMenuOption {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.options) {
		return MultiplayerMenuOptionBack
	}
	return m.options[m.selectedIndex]
}
