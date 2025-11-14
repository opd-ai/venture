package display

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager manages display resolution and configuration.
type Manager struct {
	config     Config
	mu         sync.RWMutex
	listeners  []ResolutionChangeListener
	lastWidth  int
	lastHeight int
}

// ResolutionChangeListener is called when resolution changes.
type ResolutionChangeListener func(oldWidth, oldHeight, newWidth, newHeight int)

// NewManager creates a new display manager with the given configuration.
func NewManager(config Config) *Manager {
	return &Manager{
		config:     config,
		listeners:  make([]ResolutionChangeListener, 0),
		lastWidth:  config.Width,
		lastHeight: config.Height,
	}
}

// Width returns the current display width.
func (m *Manager) Width() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Width
}

// Height returns the current display height.
func (m *Manager) Height() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Height
}

// Config returns a copy of the current configuration.
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// AspectRatio returns the current aspect ratio.
func (m *Manager) AspectRatio() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.AspectRatio()
}

// SetResolution changes the display resolution.
// Returns an error if the new resolution is invalid.
// Notifies all registered listeners of the change.
func (m *Manager) SetResolution(width, height int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	newConfig := m.config
	newConfig.Width = width
	newConfig.Height = height

	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid resolution: %w", err)
	}

	oldWidth := m.config.Width
	oldHeight := m.config.Height

	m.config = newConfig
	m.lastWidth = oldWidth
	m.lastHeight = oldHeight

	// Apply to Ebiten window
	ebiten.SetWindowSize(width, height)

	// Notify listeners
	for _, listener := range m.listeners {
		listener(oldWidth, oldHeight, width, height)
	}

	return nil
}

// SetFullscreen toggles fullscreen mode.
func (m *Manager) SetFullscreen(fullscreen bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

// IsFullscreen returns whether fullscreen mode is enabled.
func (m *Manager) IsFullscreen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Fullscreen
}

// SetScaleMode changes the scaling mode.
func (m *Manager) SetScaleMode(mode ScaleMode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.ScaleMode = mode
}

// ScaleMode returns the current scaling mode.
func (m *Manager) ScaleMode() ScaleMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.ScaleMode
}

// SetVSync enables or disables vertical synchronization.
func (m *Manager) SetVSync(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.VSync = enabled
	ebiten.SetVsyncEnabled(enabled)
}

// IsVSyncEnabled returns whether VSync is enabled.
func (m *Manager) IsVSyncEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.VSync
}

// AddResolutionChangeListener registers a callback for resolution changes.
func (m *Manager) AddResolutionChangeListener(listener ResolutionChangeListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// Apply applies the current configuration to the Ebiten window.
// This should be called during initialization.
func (m *Manager) Apply() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ebiten.SetWindowSize(m.config.Width, m.config.Height)
	ebiten.SetFullscreen(m.config.Fullscreen)
	ebiten.SetVsyncEnabled(m.config.VSync)
}

// GetCommonResolutions returns a list of common gaming resolutions.
func GetCommonResolutions() []struct{ Width, Height int } {
	return []struct{ Width, Height int }{
		{1280, 720},   // HD
		{1920, 1080},  // Full HD
		{2560, 1440},  // 2K
		{3840, 2160},  // 4K
	}
}

// CalculateScaledDimensions calculates the scaled dimensions based on scale mode.
// Returns the scaled width and height, and the offset (for letterboxing/pillarboxing).
func (m *Manager) CalculateScaledDimensions(contentWidth, contentHeight int) (scaledWidth, scaledHeight, offsetX, offsetY int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	windowWidth := m.config.Width
	windowHeight := m.config.Height

	switch m.config.ScaleMode {
	case ScaleModeFit:
		// Maintain aspect ratio, letterbox/pillarbox as needed
		contentAspect := float64(contentWidth) / float64(contentHeight)
		windowAspect := float64(windowWidth) / float64(windowHeight)

		if contentAspect > windowAspect {
			// Content is wider, scale by width
			scaledWidth = windowWidth
			scaledHeight = int(float64(windowWidth) / contentAspect)
			offsetX = 0
			offsetY = (windowHeight - scaledHeight) / 2
		} else {
			// Content is taller, scale by height
			scaledWidth = int(float64(windowHeight) * contentAspect)
			scaledHeight = windowHeight
			offsetX = (windowWidth - scaledWidth) / 2
			offsetY = 0
		}

	case ScaleModeFill, ScaleModeStretch:
		// Stretch to fill entire window
		scaledWidth = windowWidth
		scaledHeight = windowHeight
		offsetX = 0
		offsetY = 0
	}

	return
}

// GetScaleFactor returns the scale factor for the current resolution
// relative to the reference resolution (1920x1080).
func (m *Manager) GetScaleFactor() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Use height for scaling to maintain readability
	return float64(m.config.Height) / float64(DefaultHeight)
}
