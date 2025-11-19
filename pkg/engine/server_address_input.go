package engine

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/opd-ai/venture/pkg/mobile"
	"golang.org/x/image/font/basicfont"
)

// ServerAddressInput represents a text input field for entering server addresses.
type ServerAddressInput struct {
	screenWidth  int
	screenHeight int
	address      string
	cursorPos    int
	isVisible    bool
	onConnect    func(string)
	onCancel     func()
	maxLength    int
	blinkTimer   int
	showCursor   bool

	// Mobile keyboard state (WASM/mobile platforms)
	keyboardShown bool // Tracks whether mobile keyboard is currently shown
}

// NewServerAddressInput creates a new server address input with the given screen dimensions.
func NewServerAddressInput(screenWidth, screenHeight int) *ServerAddressInput {
	return &ServerAddressInput{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		address:      "localhost:8080", // Default value
		cursorPos:    len("localhost:8080"),
		isVisible:    false,
		maxLength:    50,
		blinkTimer:   0,
		showCursor:   true,
	}
}

// Show makes the server address input visible and resets to default.
func (s *ServerAddressInput) Show() {
	s.isVisible = true
	s.address = "localhost:8080"
	s.cursorPos = len(s.address)
	s.blinkTimer = 0
	s.showCursor = true

	// MOBILE/WASM: Show keyboard when input becomes visible
	// The native mobile keyboard needs to be explicitly triggered on WASM builds
	if mobile.IsWASM() {
		mobile.ShowKeyboard()
		s.keyboardShown = true
	}
}

// Hide makes the server address input invisible.
func (s *ServerAddressInput) Hide() {
	s.isVisible = false

	// MOBILE/WASM: Hide keyboard when input is hidden
	if s.keyboardShown && mobile.IsWASM() {
		mobile.HideKeyboard()
		s.keyboardShown = false
	}
}

// IsVisible returns whether the server address input is currently visible.
func (s *ServerAddressInput) IsVisible() bool {
	return s.isVisible
}

// SetConnectCallback sets the callback function for when the user presses Enter.
func (s *ServerAddressInput) SetConnectCallback(callback func(string)) {
	s.onConnect = callback
}

// SetCancelCallback sets the callback function for when the user presses Escape.
func (s *ServerAddressInput) SetCancelCallback(callback func()) {
	s.onCancel = callback
}

// GetAddress returns the current address text.
func (s *ServerAddressInput) GetAddress() string {
	return s.address
}

// SetAddress sets the address text.
func (s *ServerAddressInput) SetAddress(address string) {
	if len(address) <= s.maxLength {
		s.address = address
		s.cursorPos = len(address)
	}
}

// Update handles input for the server address input field.
func (s *ServerAddressInput) Update() {
	if !s.isVisible {
		return
	}

	s.updateCursorBlink()
	s.handleTextInput()
	s.handleEditingKeys()
	s.handleNavigationKeys()
	s.handleActionKeys()
}

// updateCursorBlink updates the cursor blink animation state.
func (s *ServerAddressInput) updateCursorBlink() {
	s.blinkTimer++
	if s.blinkTimer >= 15 {
		s.blinkTimer = 0
		s.showCursor = !s.showCursor
	}
}

// handleTextInput processes character input and inserts printable ASCII characters.
func (s *ServerAddressInput) handleTextInput() {
	runes := ebiten.AppendInputChars(nil)
	for _, r := range runes {
		if r >= 32 && r <= 126 && len(s.address) < s.maxLength {
			s.insertCharacterAtCursor(r)
		}
	}
}

// insertCharacterAtCursor inserts a character at the current cursor position.
func (s *ServerAddressInput) insertCharacterAtCursor(r rune) {
	before := s.address[:s.cursorPos]
	after := s.address[s.cursorPos:]
	s.address = before + string(r) + after
	s.cursorPos++
	s.resetCursorBlink()
}

// handleEditingKeys processes backspace and delete key inputs.
func (s *ServerAddressInput) handleEditingKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		s.deleteCharacterBeforeCursor()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		s.deleteCharacterAtCursor()
	}
}

// deleteCharacterBeforeCursor removes the character before the cursor position.
func (s *ServerAddressInput) deleteCharacterBeforeCursor() {
	if s.cursorPos > 0 {
		before := s.address[:s.cursorPos-1]
		after := s.address[s.cursorPos:]
		s.address = before + after
		s.cursorPos--
		s.resetCursorBlink()
	}
}

// deleteCharacterAtCursor removes the character at the cursor position.
func (s *ServerAddressInput) deleteCharacterAtCursor() {
	if s.cursorPos < len(s.address) {
		before := s.address[:s.cursorPos]
		after := s.address[s.cursorPos+1:]
		s.address = before + after
		s.resetCursorBlink()
	}
}

// handleNavigationKeys processes arrow keys, Home, and End for cursor navigation.
func (s *ServerAddressInput) handleNavigationKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		s.moveCursorLeft()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		s.moveCursorRight()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		s.moveCursorToStart()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		s.moveCursorToEnd()
	}
}

// moveCursorLeft moves the cursor one position to the left.
func (s *ServerAddressInput) moveCursorLeft() {
	if s.cursorPos > 0 {
		s.cursorPos--
		s.resetCursorBlink()
	}
}

// moveCursorRight moves the cursor one position to the right.
func (s *ServerAddressInput) moveCursorRight() {
	if s.cursorPos < len(s.address) {
		s.cursorPos++
		s.resetCursorBlink()
	}
}

// moveCursorToStart moves the cursor to the beginning of the address.
func (s *ServerAddressInput) moveCursorToStart() {
	s.cursorPos = 0
	s.resetCursorBlink()
}

// moveCursorToEnd moves the cursor to the end of the address.
func (s *ServerAddressInput) moveCursorToEnd() {
	s.cursorPos = len(s.address)
	s.resetCursorBlink()
}

// handleActionKeys processes Enter and Escape keys for connect and cancel actions.
func (s *ServerAddressInput) handleActionKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		s.handleConnectAction()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.handleCancelAction()
	}
}

// handleConnectAction processes the connect action when Enter is pressed.
func (s *ServerAddressInput) handleConnectAction() {
	address := strings.TrimSpace(s.address)
	if address != "" && s.onConnect != nil {
		s.hideMobileKeyboard()
		s.onConnect(address)
	}
}

// handleCancelAction processes the cancel action when Escape is pressed.
func (s *ServerAddressInput) handleCancelAction() {
	s.hideMobileKeyboard()
	if s.onCancel != nil {
		s.onCancel()
	}
}

// hideMobileKeyboard hides the mobile keyboard on WASM platforms.
func (s *ServerAddressInput) hideMobileKeyboard() {
	if s.keyboardShown && mobile.IsWASM() {
		mobile.HideKeyboard()
		s.keyboardShown = false
	}
}

// resetCursorBlink resets the cursor blink timer to show the cursor immediately.
func (s *ServerAddressInput) resetCursorBlink() {
	s.blinkTimer = 0
	s.showCursor = true
}

// Draw renders the server address input to the screen.
func (s *ServerAddressInput) Draw(screen *ebiten.Image) {
	if !s.isVisible || screen == nil {
		return
	}

	// Draw title
	titleText := "Join Server"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleWidth := titleBounds.Dx()
	titleX := s.screenWidth/2 - titleWidth/2
	titleY := s.screenHeight/2 - 100
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY, color.White)

	// Draw instruction
	instructionText := "Enter server address:"
	instructionBounds := text.BoundString(basicfont.Face7x13, instructionText)
	instructionWidth := instructionBounds.Dx()
	instructionX := s.screenWidth/2 - instructionWidth/2
	instructionY := s.screenHeight/2 - 50
	text.Draw(screen, instructionText, basicfont.Face7x13, instructionX, instructionY, color.RGBA{200, 200, 200, 255})

	// Draw input box background
	inputBoxWidth := 400
	inputBoxHeight := 30
	inputBoxX := s.screenWidth/2 - inputBoxWidth/2
	inputBoxY := s.screenHeight/2 - 20

	// Draw border
	borderColor := color.RGBA{100, 100, 100, 255}
	for y := inputBoxY; y < inputBoxY+inputBoxHeight; y++ {
		screen.Set(inputBoxX, y, borderColor)
		screen.Set(inputBoxX+inputBoxWidth-1, y, borderColor)
	}
	for x := inputBoxX; x < inputBoxX+inputBoxWidth; x++ {
		screen.Set(x, inputBoxY, borderColor)
		screen.Set(x, inputBoxY+inputBoxHeight-1, borderColor)
	}

	// Draw input text
	textX := inputBoxX + 10
	textY := inputBoxY + 20
	textColor := color.White

	// Draw text before cursor
	beforeCursor := s.address[:s.cursorPos]
	text.Draw(screen, beforeCursor, basicfont.Face7x13, textX, textY, textColor)

	// Calculate cursor position
	beforeBounds := text.BoundString(basicfont.Face7x13, beforeCursor)
	cursorX := textX + beforeBounds.Dx()

	// Draw cursor (blinking vertical line)
	if s.showCursor {
		cursorColor := color.RGBA{255, 255, 100, 255}
		for y := textY - 12; y < textY+2; y++ {
			screen.Set(cursorX, y, cursorColor)
		}
	}

	// Draw text after cursor
	afterCursor := s.address[s.cursorPos:]
	afterX := cursorX + 1
	text.Draw(screen, afterCursor, basicfont.Face7x13, afterX, textY, textColor)

	// Draw controls hint at bottom
	hintText := "[Enter] Connect  [Esc] Cancel  [←→] Move Cursor  [Backspace] Delete"
	hintBounds := text.BoundString(basicfont.Face7x13, hintText)
	hintWidth := hintBounds.Dx()
	hintX := s.screenWidth/2 - hintWidth/2
	hintY := s.screenHeight - 30
	hintColor := color.RGBA{100, 100, 100, 255}
	text.Draw(screen, hintText, basicfont.Face7x13, hintX, hintY, hintColor)
}
