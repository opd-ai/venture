// Package engine provides genre selection menu for single-player game setup.
// This file implements the menu shown after selecting "New Game" from the
// single-player submenu, allowing players to choose their preferred game genre.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/procgen/genre"
)

// GenreSelectionMenu renders and handles input for the genre selection screen.
// Displays all available genres from the genre registry with their descriptions.
type GenreSelectionMenu struct {
	screenWidth  int
	screenHeight int
	selectedIdx  int
	genres       []*genre.Genre
	visible      bool

	// Callback for when a genre is selected
	onGenreSelect func(genreID string)
	onBack        func()
}

// NewGenreSelectionMenu creates a new genre selection menu using the default registry.
func NewGenreSelectionMenu(screenWidth, screenHeight int) *GenreSelectionMenu {
	registry := genre.DefaultRegistry()
	genres := registry.All()

	return &GenreSelectionMenu{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		selectedIdx:  0,
		genres:       genres,
		visible:      false,
	}
}

// SetGenreSelectCallback sets the callback for when a genre is selected.
func (m *GenreSelectionMenu) SetGenreSelectCallback(callback func(genreID string)) {
	m.onGenreSelect = callback
}

// SetBackCallback sets the callback for returning to the previous menu.
func (m *GenreSelectionMenu) SetBackCallback(callback func()) {
	m.onBack = callback
}

// Show makes the menu visible and resets selection to first option.
func (m *GenreSelectionMenu) Show() {
	m.visible = true
	m.selectedIdx = 0
}

// Hide makes the menu invisible.
func (m *GenreSelectionMenu) Hide() {
	m.visible = false
}

// IsVisible returns whether the menu is currently displayed.
func (m *GenreSelectionMenu) IsVisible() bool {
	return m.visible
}

// Update processes input for the genre selection menu.
// Returns true if a genre was selected.
func (m *GenreSelectionMenu) Update() bool {
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
			m.selectedIdx = len(m.genres) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedIdx++
		if m.selectedIdx >= len(m.genres) {
			m.selectedIdx = 0
		}
	}

	// Handle number key shortcuts (1-5 for genres)
	for i := 0; i < len(m.genres) && i < 9; i++ {
		key := ebiten.Key(int(ebiten.Key1) + i)
		if inpututil.IsKeyJustPressed(key) {
			m.selectedIdx = i
			return m.selectCurrentGenre()
		}
	}

	// Handle Enter/Space for selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return m.selectCurrentGenre()
	}

	// Handle mouse input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if idx := m.getGenreAtPosition(mx, my); idx >= 0 {
			m.selectedIdx = idx
			return m.selectCurrentGenre()
		}
	}

	// Update hover state for mouse
	mx, my := ebiten.CursorPosition()
	if idx := m.getGenreAtPosition(mx, my); idx >= 0 {
		m.selectedIdx = idx
	}

	return false
}

// selectCurrentGenre triggers the callback for the currently selected genre.
func (m *GenreSelectionMenu) selectCurrentGenre() bool {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.genres) {
		return false
	}

	selectedGenre := m.genres[m.selectedIdx]
	if m.onGenreSelect != nil {
		m.onGenreSelect(selectedGenre.ID)
	}
	return true
}

// getGenreAtPosition returns the genre index at the given screen coordinates,
// or -1 if no genre is at that position.
func (m *GenreSelectionMenu) getGenreAtPosition(x, y int) int {
	startY := m.screenHeight/2 - 100
	genreHeight := 50

	for i := range m.genres {
		genreY := startY + i*genreHeight
		if y >= genreY && y < genreY+40 {
			// Check X bounds (centered text, approximate width based on name length)
			genreName := m.genres[i].Name
			textWidth := len(genreName) * 7
			startX := m.screenWidth/2 - textWidth/2
			if x >= startX-20 && x <= startX+textWidth+20 {
				return i
			}
		}
	}
	return -1
}

// Draw renders the genre selection menu to the screen.
func (m *GenreSelectionMenu) Draw(screen *ebiten.Image) {
	if !m.visible {
		return
	}

	// Draw semi-transparent background
	bgColor := color.RGBA{0, 0, 0, 200}
	ebitenutil.DrawRect(screen, 0, 0, float64(m.screenWidth), float64(m.screenHeight), bgColor)

	// Draw title
	titleText := "Select Your Genre"
	titleX := float64(m.screenWidth/2 - len(titleText)*7/2)
	titleY := float64(m.screenHeight/2 - 150)
	ebitenutil.DebugPrintAt(screen, titleText, int(titleX), int(titleY))

	// Draw subtitle
	subtitleText := "Choose the world you want to explore"
	subtitleX := float64(m.screenWidth/2 - len(subtitleText)*7/2)
	subtitleY := titleY + 25
	ebitenutil.DebugPrintAt(screen, subtitleText, int(subtitleX), int(subtitleY))

	// Draw genre options
	startY := m.screenHeight/2 - 100
	genreHeight := 50

	for i, g := range m.genres {
		genreName := g.Name
		textWidth := len(genreName) * 7
		x := m.screenWidth/2 - textWidth/2
		y := startY + i*genreHeight

		// Draw selection indicator
		if i == m.selectedIdx {
			indicatorText := ">"
			indicatorX := x - 20
			ebitenutil.DebugPrintAt(screen, indicatorText, indicatorX, y)
		}

		// Draw genre name
		ebitenutil.DebugPrintAt(screen, genreName, x, y)

		// Draw genre description (smaller, below name)
		descText := g.Description
		if len(descText) > 80 {
			descText = descText[:77] + "..."
		}
		descWidth := len(descText) * 6
		descX := m.screenWidth/2 - descWidth/2
		descY := y + 15
		ebitenutil.DebugPrintAt(screen, descText, descX, descY)

		// Draw number shortcut hint
		if i < 9 {
			shortcutText := "[" + string(rune('1'+i)) + "]"
			shortcutX := x + textWidth + 10
			ebitenutil.DebugPrintAt(screen, shortcutText, shortcutX, y)
		}
	}

	// Draw controls hint
	hintText := "Arrow Keys/WASD: Navigate | Enter/Space: Select | ESC: Back | 1-5: Quick Select"
	hintX := m.screenWidth/2 - len(hintText)*7/2
	hintY := m.screenHeight - 40
	ebitenutil.DebugPrintAt(screen, hintText, hintX, hintY)
}

// GetSelectedGenre returns the currently selected genre.
func (m *GenreSelectionMenu) GetSelectedGenre() *genre.Genre {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.genres) {
		return nil
	}
	return m.genres[m.selectedIdx]
}

// GetGenreCount returns the number of available genres.
func (m *GenreSelectionMenu) GetGenreCount() int {
	return len(m.genres)
}
