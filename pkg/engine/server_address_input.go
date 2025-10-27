package engine

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
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
}

// Hide makes the server address input invisible.
func (s *ServerAddressInput) Hide() {
	s.isVisible = false
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

	// Update cursor blink (30 FPS blink rate = 15 frames per state)
	s.blinkTimer++
	if s.blinkTimer >= 15 {
		s.blinkTimer = 0
		s.showCursor = !s.showCursor
	}

	// Handle text input
	runes := ebiten.AppendInputChars(nil)
	for _, r := range runes {
		// Only accept printable ASCII characters
		if r >= 32 && r <= 126 && len(s.address) < s.maxLength {
			// Insert character at cursor position
			before := s.address[:s.cursorPos]
			after := s.address[s.cursorPos:]
			s.address = before + string(r) + after
			s.cursorPos++
			s.resetCursorBlink()
		}
	}

	// Backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if s.cursorPos > 0 {
			before := s.address[:s.cursorPos-1]
			after := s.address[s.cursorPos:]
			s.address = before + after
			s.cursorPos--
			s.resetCursorBlink()
		}
	}

	// Delete key
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		if s.cursorPos < len(s.address) {
			before := s.address[:s.cursorPos]
			after := s.address[s.cursorPos+1:]
			s.address = before + after
			s.resetCursorBlink()
		}
	}

	// Left arrow
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		if s.cursorPos > 0 {
			s.cursorPos--
			s.resetCursorBlink()
		}
	}

	// Right arrow
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		if s.cursorPos < len(s.address) {
			s.cursorPos++
			s.resetCursorBlink()
		}
	}

	// Home key
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		s.cursorPos = 0
		s.resetCursorBlink()
	}

	// End key
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		s.cursorPos = len(s.address)
		s.resetCursorBlink()
	}

	// Enter to connect
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		address := strings.TrimSpace(s.address)
		if address != "" && s.onConnect != nil {
			s.onConnect(address)
		}
	}

	// Escape to cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.onCancel != nil {
			s.onCancel()
		}
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
