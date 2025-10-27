// Package engine provides settings menu UI components.
// This file implements the settings menu with interactive controls for adjusting game settings.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// SettingsOption represents a configurable setting in the settings menu.
type SettingsOption int

const (
	SettingsOptionMasterVolume SettingsOption = iota
	SettingsOptionMusicVolume
	SettingsOptionSFXVolume
	SettingsOptionGraphicsQuality
	SettingsOptionVSync
	SettingsOptionShowFPS
	SettingsOptionFullscreen
	SettingsOptionBack
)

// String returns the display text for each setting option.
func (o SettingsOption) String() string {
	switch o {
	case SettingsOptionMasterVolume:
		return "Master Volume"
	case SettingsOptionMusicVolume:
		return "Music Volume"
	case SettingsOptionSFXVolume:
		return "SFX Volume"
	case SettingsOptionGraphicsQuality:
		return "Graphics Quality"
	case SettingsOptionVSync:
		return "VSync"
	case SettingsOptionShowFPS:
		return "Show FPS"
	case SettingsOptionFullscreen:
		return "Fullscreen"
	case SettingsOptionBack:
		return "Back"
	default:
		return "Unknown"
	}
}

// SettingsUI renders and handles input for the settings menu.
// Provides interactive controls for adjusting settings with immediate feedback.
type SettingsUI struct {
	screenWidth  int
	screenHeight int
	selectedIdx  int
	options      []SettingsOption

	// Settings manager for persistence
	settingsManager *SettingsManager

	// Current settings (modified in realtime)
	currentSettings GameSettings

	// Callback for when back is selected
	onBack func()

	// Visibility flag
	visible bool
}

// NewSettingsUI creates a new settings UI with the provided settings manager.
func NewSettingsUI(screenWidth, screenHeight int, settingsManager *SettingsManager) *SettingsUI {
	return &SettingsUI{
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		selectedIdx:     0,
		settingsManager: settingsManager,
		currentSettings: settingsManager.GetSettings(),
		options: []SettingsOption{
			SettingsOptionMasterVolume,
			SettingsOptionMusicVolume,
			SettingsOptionSFXVolume,
			SettingsOptionGraphicsQuality,
			SettingsOptionVSync,
			SettingsOptionShowFPS,
			SettingsOptionFullscreen,
			SettingsOptionBack,
		},
		visible: false,
	}
}

// SetBackCallback sets the callback function called when "Back" is selected.
func (s *SettingsUI) SetBackCallback(callback func()) {
	s.onBack = callback
}

// Show displays the settings menu and loads current settings.
func (s *SettingsUI) Show() {
	s.visible = true
	s.currentSettings = s.settingsManager.GetSettings()
	s.selectedIdx = 0
}

// Hide hides the settings menu and saves any changes.
func (s *SettingsUI) Hide() {
	s.visible = false
	// Save settings on hide
	s.settingsManager.UpdateSettings(s.currentSettings)
}

// IsVisible returns whether the settings menu is currently visible.
func (s *SettingsUI) IsVisible() bool {
	return s.visible
}

// Update processes input for the settings menu.
// Returns true if a significant action occurred (e.g., back selected).
func (s *SettingsUI) Update() bool {
	if !s.visible {
		return false
	}

	// Handle up/down navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.selectedIdx--
		if s.selectedIdx < 0 {
			s.selectedIdx = len(s.options) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.selectedIdx++
		if s.selectedIdx >= len(s.options) {
			s.selectedIdx = 0
		}
	}

	// Handle value adjustments (left/right for selected option)
	selectedOption := s.options[s.selectedIdx]
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		s.decreaseValue(selectedOption)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		s.increaseValue(selectedOption)
	}

	// Handle Enter/Space to toggle booleans or go back
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.activateOption(selectedOption)
	}

	// Handle ESC key to go back
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.Hide()
		if s.onBack != nil {
			s.onBack()
		}
		return true
	}

	return false
}

// decreaseValue decreases the value of the selected setting.
func (s *SettingsUI) decreaseValue(option SettingsOption) {
	switch option {
	case SettingsOptionMasterVolume:
		s.currentSettings.MasterVolume -= 0.1
		if s.currentSettings.MasterVolume < 0.0 {
			s.currentSettings.MasterVolume = 0.0
		}
	case SettingsOptionMusicVolume:
		s.currentSettings.MusicVolume -= 0.1
		if s.currentSettings.MusicVolume < 0.0 {
			s.currentSettings.MusicVolume = 0.0
		}
	case SettingsOptionSFXVolume:
		s.currentSettings.SFXVolume -= 0.1
		if s.currentSettings.SFXVolume < 0.0 {
			s.currentSettings.SFXVolume = 0.0
		}
	case SettingsOptionGraphicsQuality:
		// Cycle: high -> medium -> low -> high
		switch s.currentSettings.GraphicsQuality {
		case "high":
			s.currentSettings.GraphicsQuality = "medium"
		case "medium":
			s.currentSettings.GraphicsQuality = "low"
		case "low":
			s.currentSettings.GraphicsQuality = "high"
		}
	case SettingsOptionVSync:
		s.currentSettings.VSync = !s.currentSettings.VSync
	case SettingsOptionShowFPS:
		s.currentSettings.ShowFPS = !s.currentSettings.ShowFPS
	case SettingsOptionFullscreen:
		s.currentSettings.Fullscreen = !s.currentSettings.Fullscreen
	}
}

// increaseValue increases the value of the selected setting.
func (s *SettingsUI) increaseValue(option SettingsOption) {
	switch option {
	case SettingsOptionMasterVolume:
		s.currentSettings.MasterVolume += 0.1
		if s.currentSettings.MasterVolume > 1.0 {
			s.currentSettings.MasterVolume = 1.0
		}
	case SettingsOptionMusicVolume:
		s.currentSettings.MusicVolume += 0.1
		if s.currentSettings.MusicVolume > 1.0 {
			s.currentSettings.MusicVolume = 1.0
		}
	case SettingsOptionSFXVolume:
		s.currentSettings.SFXVolume += 0.1
		if s.currentSettings.SFXVolume > 1.0 {
			s.currentSettings.SFXVolume = 1.0
		}
	case SettingsOptionGraphicsQuality:
		// Cycle: low -> medium -> high -> low
		switch s.currentSettings.GraphicsQuality {
		case "low":
			s.currentSettings.GraphicsQuality = "medium"
		case "medium":
			s.currentSettings.GraphicsQuality = "high"
		case "high":
			s.currentSettings.GraphicsQuality = "low"
		}
	case SettingsOptionVSync:
		s.currentSettings.VSync = !s.currentSettings.VSync
	case SettingsOptionShowFPS:
		s.currentSettings.ShowFPS = !s.currentSettings.ShowFPS
	case SettingsOptionFullscreen:
		s.currentSettings.Fullscreen = !s.currentSettings.Fullscreen
	}
}

// activateOption activates the selected option (toggle or navigate).
func (s *SettingsUI) activateOption(option SettingsOption) {
	switch option {
	case SettingsOptionBack:
		s.Hide()
		if s.onBack != nil {
			s.onBack()
		}
	case SettingsOptionVSync:
		s.currentSettings.VSync = !s.currentSettings.VSync
	case SettingsOptionShowFPS:
		s.currentSettings.ShowFPS = !s.currentSettings.ShowFPS
	case SettingsOptionFullscreen:
		s.currentSettings.Fullscreen = !s.currentSettings.Fullscreen
	}
}

// Draw renders the settings menu to the screen.
func (s *SettingsUI) Draw(screen *ebiten.Image) {
	if !s.visible || screen == nil {
		return
	}

	// Draw semi-transparent background
	bgColor := color.RGBA{0, 0, 0, 200}
	ebitenutil.DrawRect(screen, 0, 0, float64(s.screenWidth), float64(s.screenHeight), bgColor)

	// Draw title
	titleX := float64(s.screenWidth/2 - 100)
	titleY := 50.0
	ebitenutil.DebugPrintAt(screen, "=== SETTINGS ===", int(titleX), int(titleY))

	// Draw options
	startY := 120
	lineHeight := 40

	for i, option := range s.options {
		y := startY + i*lineHeight

		// Highlight selected option
		if i == s.selectedIdx {
			// Draw selection indicator
			ebitenutil.DebugPrintAt(screen, ">", s.screenWidth/2-150, y)
		}

		// Draw option label
		labelX := s.screenWidth/2 - 120
		ebitenutil.DebugPrintAt(screen, option.String(), labelX, y)

		// Draw option value
		valueX := s.screenWidth/2 + 50
		valueStr := s.getValueString(option)
		if i == s.selectedIdx {
			// Draw adjustment hint for selected option
			if s.isAdjustable(option) {
				valueStr = fmt.Sprintf("< %s >", valueStr)
			}
		}
		ebitenutil.DebugPrintAt(screen, valueStr, valueX, y)
	}

	// Draw controls hint at bottom
	hintY := s.screenHeight - 80
	ebitenutil.DebugPrintAt(screen, "Controls: Up/Down - Navigate", s.screenWidth/2-120, hintY)
	ebitenutil.DebugPrintAt(screen, "Left/Right - Adjust | Enter/Space - Toggle", s.screenWidth/2-180, hintY+20)
	ebitenutil.DebugPrintAt(screen, "ESC - Back and Save", s.screenWidth/2-80, hintY+40)
}

// getValueString returns the current value of a setting as a string.
func (s *SettingsUI) getValueString(option SettingsOption) string {
	switch option {
	case SettingsOptionMasterVolume:
		return fmt.Sprintf("%.0f%%", s.currentSettings.MasterVolume*100)
	case SettingsOptionMusicVolume:
		return fmt.Sprintf("%.0f%%", s.currentSettings.MusicVolume*100)
	case SettingsOptionSFXVolume:
		return fmt.Sprintf("%.0f%%", s.currentSettings.SFXVolume*100)
	case SettingsOptionGraphicsQuality:
		return s.currentSettings.GraphicsQuality
	case SettingsOptionVSync:
		if s.currentSettings.VSync {
			return "ON"
		}
		return "OFF"
	case SettingsOptionShowFPS:
		if s.currentSettings.ShowFPS {
			return "ON"
		}
		return "OFF"
	case SettingsOptionFullscreen:
		if s.currentSettings.Fullscreen {
			return "ON"
		}
		return "OFF"
	case SettingsOptionBack:
		return ""
	default:
		return ""
	}
}

// isAdjustable returns whether the option can be adjusted with left/right keys.
func (s *SettingsUI) isAdjustable(option SettingsOption) bool {
	return option != SettingsOptionBack
}

// GetCurrentSettings returns the current settings being edited.
// Note: These may not be saved yet until Hide() or manual save.
func (s *SettingsUI) GetCurrentSettings() GameSettings {
	return s.currentSettings
}
