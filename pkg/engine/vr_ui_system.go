// Package engine provides the VR UI system for VR support.

package engine

import (
	"math"
	"sync"

	log "github.com/sirupsen/logrus"
)

// VRUISystem manages VR UI by updating panel positions and handling gaze interaction.
type VRUISystem struct {
	mu sync.RWMutex

	world *World

	// Callbacks
	gazeActivateCallback func(panelID string)
	menuOpenCallback     func()
	menuCloseCallback    func()

	// State
	enabled     bool
	menuOpen    bool
	lastGazeHit string

	// Head tracking reference for follow-head panels
	headPitch float64
	headYaw   float64
	headX     float64
	headY     float64
	headZ     float64

	// Controller positions for hand-locked panels
	leftHandX, leftHandY, leftHandZ    float64
	rightHandX, rightHandY, rightHandZ float64
}

// NewVRUISystem creates a new VR UI system.
func NewVRUISystem(world *World) *VRUISystem {
	log.WithFields(log.Fields{
		"system_name": "vr_ui",
	}).Debug("Creating VR UI system")

	return &VRUISystem{
		world:   world,
		enabled: false,
	}
}

// Update processes VR UI for entities with VRUIComponent.
func (s *VRUISystem) Update(entities []*Entity, deltaTime float64) {
	s.mu.RLock()
	if !s.enabled {
		s.mu.RUnlock()
		return
	}
	s.mu.RUnlock()

	for _, entity := range entities {
		comp, ok := entity.GetComponent("vr_ui")
		if !ok || comp == nil {
			continue
		}

		ui, ok := comp.(*VRUIComponent)
		if !ok {
			continue
		}

		if !ui.IsEnabled() {
			continue
		}

		// Update panel positions
		s.updatePanelPositions(ui)

		// Process gaze interaction
		s.processGaze(ui, deltaTime)

		// Update comfort vignette
		s.updateComfortVignette(ui, deltaTime)
	}
}

// updatePanelPositions updates positions of follow-head and hand-locked panels.
func (s *VRUISystem) updatePanelPositions(ui *VRUIComponent) {
	s.mu.RLock()
	headX, headY, headZ := s.headX, s.headY, s.headZ
	headYaw := s.headYaw
	leftX, leftY, leftZ := s.leftHandX, s.leftHandY, s.leftHandZ
	rightX, rightY, rightZ := s.rightHandX, s.rightHandY, s.rightHandZ
	s.mu.RUnlock()

	for _, panel := range ui.GetAllPanels() {
		if panel.FollowHead {
			// Calculate position in front of head using stored offsets
			// Use yaw to rotate the offset direction
			sinYaw := sinApprox(headYaw)
			cosYaw := cosApprox(headYaw)

			// Use stored offsets instead of current position to avoid drift
			offsetX := panel.FollowOffsetX*cosYaw - panel.FollowDistance*sinYaw
			offsetZ := panel.FollowOffsetX*sinYaw + panel.FollowDistance*cosYaw

			// Panel follows head at fixed relative offset
			panel.WorldX = headX + offsetX
			panel.WorldY = headY + panel.FollowOffsetY
			panel.WorldZ = headZ + offsetZ
		}

		if panel.LockedToHand == ControllerLeft {
			panel.WorldX = leftX
			panel.WorldY = leftY + 0.1 // Slightly above wrist
			panel.WorldZ = leftZ
		} else if panel.LockedToHand == ControllerRight {
			panel.WorldX = rightX
			panel.WorldY = rightY + 0.1
			panel.WorldZ = rightZ
		}
	}
}

// processGaze handles gaze-based UI interaction.
func (s *VRUISystem) processGaze(ui *VRUIComponent, deltaTime float64) {
	s.mu.RLock()
	gazeCallback := s.gazeActivateCallback
	headX, headY, headZ := s.headX, s.headY, s.headZ
	headPitch, headYaw := s.headPitch, s.headYaw
	s.mu.RUnlock()

	// Calculate gaze direction from head orientation
	cosPitch := cosApprox(headPitch)
	sinPitch := sinApprox(headPitch)
	cosYaw := cosApprox(headYaw)
	sinYaw := sinApprox(headYaw)

	// Gaze direction vector
	gazeDirX := cosPitch * sinYaw
	gazeDirY := sinPitch
	gazeDirZ := cosPitch * cosYaw

	// Find panel under gaze
	hitPanel := ui.CalculateGazeRayIntersection(headX, headY, headZ, gazeDirX, gazeDirY, gazeDirZ)

	s.mu.Lock()
	prevHit := s.lastGazeHit
	s.lastGazeHit = hitPanel
	s.mu.Unlock()

	// Update gaze target
	if hitPanel != prevHit {
		ui.SetGazeTarget(hitPanel)
	}

	// Check for gaze activation
	if hitPanel != "" {
		if ui.UpdateGazeHover(deltaTime) && gazeCallback != nil {
			gazeCallback(hitPanel)
		}
	}
}

// updateComfortVignette updates comfort vignette based on movement.
func (s *VRUISystem) updateComfortVignette(ui *VRUIComponent, deltaTime float64) {
	// Check if player is moving by examining velocity
	// This is a simplified check - in practice, would check actual movement input
	s.mu.RLock()
	isMoving := s.headX != 0 || s.headZ != 0 // Placeholder for actual movement detection
	s.mu.RUnlock()

	ui.UpdateComfortVignette(isMoving, deltaTime)
}

// SetHeadTracking updates the head position and orientation for UI calculations.
func (s *VRUISystem) SetHeadTracking(x, y, z, pitch, yaw float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.headX = x
	s.headY = y
	s.headZ = z
	s.headPitch = pitch
	s.headYaw = yaw
}

// SetHandTracking updates controller positions for hand-locked panels.
func (s *VRUISystem) SetHandTracking(leftX, leftY, leftZ, rightX, rightY, rightZ float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leftHandX = leftX
	s.leftHandY = leftY
	s.leftHandZ = leftZ
	s.rightHandX = rightX
	s.rightHandY = rightY
	s.rightHandZ = rightZ
}

// SetGazeActivateCallback sets the callback for gaze activation.
func (s *VRUISystem) SetGazeActivateCallback(cb func(panelID string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gazeActivateCallback = cb
}

// SetMenuOpenCallback sets the callback for menu opening.
func (s *VRUISystem) SetMenuOpenCallback(cb func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.menuOpenCallback = cb
}

// SetMenuCloseCallback sets the callback for menu closing.
func (s *VRUISystem) SetMenuCloseCallback(cb func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.menuCloseCallback = cb
}

// OpenMenu shows the VR menu panel.
func (s *VRUISystem) OpenMenu(ui *VRUIComponent) {
	ui.SetPanelVisible("menu", true)

	s.mu.Lock()
	s.menuOpen = true
	cb := s.menuOpenCallback
	s.mu.Unlock()

	if cb != nil {
		cb()
	}
}

// CloseMenu hides the VR menu panel.
func (s *VRUISystem) CloseMenu(ui *VRUIComponent) {
	ui.SetPanelVisible("menu", false)

	s.mu.Lock()
	s.menuOpen = false
	cb := s.menuCloseCallback
	s.mu.Unlock()

	if cb != nil {
		cb()
	}
}

// ToggleMenu toggles the VR menu visibility.
func (s *VRUISystem) ToggleMenu(ui *VRUIComponent) {
	s.mu.RLock()
	isOpen := s.menuOpen
	s.mu.RUnlock()

	if isOpen {
		s.CloseMenu(ui)
	} else {
		s.OpenMenu(ui)
	}
}

// IsMenuOpen returns whether the menu is open.
func (s *VRUISystem) IsMenuOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.menuOpen
}

// ToggleInventory toggles inventory panel visibility.
func (s *VRUISystem) ToggleInventory(ui *VRUIComponent) {
	panel := ui.GetPanel("inventory")
	if panel != nil {
		ui.SetPanelVisible("inventory", !panel.Visible)
	}
}

// ShowNotification shows a notification panel briefly.
func (s *VRUISystem) ShowNotification(ui *VRUIComponent, visible bool) {
	ui.SetPanelVisible("notifications", visible)
}

// SetEnabled enables or disables the VR UI system.
func (s *VRUISystem) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled

	log.WithFields(log.Fields{
		"system_name": "vr_ui",
		"enabled":     enabled,
	}).Info("VR UI system toggled")
}

// IsEnabled returns whether the system is enabled.
func (s *VRUISystem) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// sinApprox provides sine calculation using math package.
func sinApprox(angle float64) float64 {
	return math.Sin(angle)
}

// cosApprox provides cosine calculation using math package.
func cosApprox(angle float64) float64 {
	return math.Cos(angle)
}
