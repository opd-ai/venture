// Package main provides a minimal WebAssembly test application to verify
// virtual keyboard functionality in mobile browsers.
//
// This test app creates a simple text input field to isolate keyboard behavior
// and provide diagnostic logging for debugging keyboard issues.
package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/mobile"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Game implements the minimal keyboard test game.
type Game struct {
	inputText     string
	inputBuffer   []rune
	keyboardShown bool
	tapCount      int
	eventLog      []string
	maxLogEntries int
}

// NewGame creates a new keyboard test game.
func NewGame() *Game {
	return &Game{
		inputText:     "",
		maxLogEntries: 10,
		eventLog:      make([]string, 0),
	}
}

// Update updates the game logic.
func (g *Game) Update() error {
	// Show keyboard on first tap
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.tapCount++
		g.addLog(fmt.Sprintf("Tap detected (count: %d)", g.tapCount))

		if !g.keyboardShown && mobile.IsWASM() {
			g.addLog("Calling mobile.ShowKeyboard()")
			mobile.ShowKeyboard()
			g.keyboardShown = true
			g.addLog("Keyboard shown flag set to true")
		}
	}

	// Capture input characters
	g.inputBuffer = ebiten.AppendInputChars(g.inputBuffer[:0])
	if len(g.inputBuffer) > 0 {
		for _, r := range g.inputBuffer {
			g.inputText += string(r)
			g.addLog(fmt.Sprintf("Character received: '%c' (code: %d)", r, r))
		}
	}

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(g.inputText) > 0 {
			g.inputText = g.inputText[:len(g.inputText)-1]
			g.addLog("Backspace pressed")
		}
	}

	// Handle Enter (hide keyboard)
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.addLog("Enter pressed - hiding keyboard")
		mobile.HideKeyboard()
		g.keyboardShown = false
	}

	// Handle Escape (clear and hide)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.addLog("Escape pressed - clearing and hiding keyboard")
		g.inputText = ""
		mobile.HideKeyboard()
		g.keyboardShown = false
	}

	return nil
}

// Draw renders the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw title
	title := "Virtual Keyboard Test"
	text.Draw(screen, title, basicfont.Face7x13, 20, 30, color.White)

	// Draw platform info
	platform := fmt.Sprintf("Platform: WASM=%v, IsKeyboardSupported=%v",
		mobile.IsWASM(), mobile.IsKeyboardSupported())
	text.Draw(screen, platform, basicfont.Face7x13, 20, 60, color.RGBA{150, 150, 150, 255})

	// Draw instructions
	instructions := []string{
		"Instructions:",
		"1. TAP anywhere to show keyboard",
		"2. TYPE on your device keyboard",
		"3. Press ENTER to hide keyboard",
		"4. Press ESC to clear text",
	}
	y := 90
	for _, line := range instructions {
		text.Draw(screen, line, basicfont.Face7x13, 20, y, color.RGBA{100, 200, 100, 255})
		y += 20
	}

	// Draw input box
	boxX := 20
	boxY := 220
	boxWidth := screenWidth - 40
	boxHeight := 50

	// Background
	vector.DrawFilledRect(screen, float32(boxX), float32(boxY),
		float32(boxWidth), float32(boxHeight),
		color.RGBA{40, 40, 50, 255}, false)

	// Border
	borderColor := color.RGBA{100, 100, 120, 255}
	if g.keyboardShown {
		borderColor = color.RGBA{100, 200, 100, 255}
	}
	vector.StrokeRect(screen, float32(boxX), float32(boxY),
		float32(boxWidth), float32(boxHeight),
		2, borderColor, false)

	// Input text
	displayText := g.inputText
	if len(displayText) == 0 {
		displayText = "[Tap to enter text]"
		text.Draw(screen, displayText, basicfont.Face7x13,
			boxX+10, boxY+30, color.RGBA{100, 100, 100, 255})
	} else {
		text.Draw(screen, displayText, basicfont.Face7x13,
			boxX+10, boxY+30, color.White)
	}

	// Draw status
	statusY := boxY + boxHeight + 30
	status := fmt.Sprintf("Keyboard Status: %v | Tap Count: %d | Input Length: %d",
		map[bool]string{true: "SHOWN", false: "HIDDEN"}[g.keyboardShown],
		g.tapCount, len(g.inputText))
	text.Draw(screen, status, basicfont.Face7x13, 20, statusY,
		color.RGBA{150, 150, 200, 255})

	// Draw event log
	logY := statusY + 40
	text.Draw(screen, "Event Log:", basicfont.Face7x13, 20, logY,
		color.RGBA{200, 200, 100, 255})
	logY += 20

	// Show last N events
	startIdx := 0
	if len(g.eventLog) > g.maxLogEntries {
		startIdx = len(g.eventLog) - g.maxLogEntries
	}
	for i := startIdx; i < len(g.eventLog); i++ {
		text.Draw(screen, g.eventLog[i], basicfont.Face7x13, 30, logY,
			color.RGBA{180, 180, 180, 255})
		logY += 15
	}
}

// Layout returns the game screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// addLog adds an entry to the event log.
func (g *Game) addLog(message string) {
	g.eventLog = append(g.eventLog, message)
	// Also log to console for browser debugging
	log.Println("KeyboardTest:", message)
}

func main() {
	log.Println("Starting Keyboard Test Application")
	log.Printf("Platform: WASM=%v, IsKeyboardSupported=%v",
		mobile.IsWASM(), mobile.IsKeyboardSupported())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Virtual Keyboard Test")

	game := NewGame()
	game.addLog("Application started")
	game.addLog(fmt.Sprintf("Platform: WASM=%v", mobile.IsWASM()))

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
