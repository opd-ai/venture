package display

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager handles resolution changes and window management.
type Manager struct {
	config         *Config
	switchStarted  time.Time     // Time when the most recent resolution switch started (for performance tracking)
	switchDuration time.Duration // Duration of the most recent resolution switch operation (for diagnostics)
}

// NewManager creates a display manager.
func NewManager(cfg *Config) *Manager {
	return &Manager{
		config:         cfg,
		switchDuration: 0,
	}
}

// ApplyResolution applies current config to Ebiten window.
// Returns time taken for the switch operation.
// NOTE: Uses time.Now() for performance measurement (non-deterministic by design).
// This is acceptable as it's for observability, not game logic or procgen.
func (m *Manager) ApplyResolution() time.Duration {
	m.switchStarted = time.Now() // NON-DETERMINISTIC: performance measurement only

	ebiten.SetWindowSize(m.config.Width, m.config.Height)
	ebiten.SetFullscreen(m.config.Fullscreen)
	ebiten.SetVsyncEnabled(m.config.VSync)

	m.switchDuration = time.Since(m.switchStarted) // NON-DETERMINISTIC: performance measurement only
	return m.switchDuration
}

// SetResolution changes resolution with validation.
func (m *Manager) SetResolution(width, height int) error {
	if !IsValidResolution(width, height) {
		return ErrUnsupportedResolution
	}

	m.config.Width = width
	m.config.Height = height
	m.ApplyResolution()
	return nil
}

// SetFullscreen toggles fullscreen mode.
func (m *Manager) SetFullscreen(fullscreen bool) {
	m.config.Fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

// ToggleFullscreen switches between windowed and fullscreen.
func (m *Manager) ToggleFullscreen() {
	m.SetFullscreen(!m.config.Fullscreen)
}

// GetConfig returns current configuration (read-only copy).
func (m *Manager) GetConfig() Config {
	return *m.config
}

// GetLastSwitchDuration returns duration of last resolution switch.
func (m *Manager) GetLastSwitchDuration() time.Duration {
	return m.switchDuration
}

// SupportedResolutions returns list of all supported resolutions.
func (m *Manager) SupportedResolutions() []Resolution {
	return GetStandardResolutions()
}
