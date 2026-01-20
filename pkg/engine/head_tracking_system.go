// Package engine provides the head tracking system for VR support.

package engine

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// MockHeadset provides a test implementation of VRHeadsetAdapter.
type MockHeadset struct {
	mu        sync.RWMutex
	connected bool
	pitch     float64
	yaw       float64
	roll      float64
	posX      float64
	posY      float64
	posZ      float64
	ipd       float64
}

// NewMockHeadset creates a mock headset for testing.
func NewMockHeadset() *MockHeadset {
	return &MockHeadset{
		connected: true,
		ipd:       63.0,
	}
}

// IsConnected returns true if the mock headset is connected.
func (m *MockHeadset) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// SetConnected sets the connection status for testing.
func (m *MockHeadset) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

// GetHeadOrientation returns the mock orientation.
func (m *MockHeadset) GetHeadOrientation() (pitch, yaw, roll float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pitch, m.yaw, m.roll
}

// SetHeadOrientation sets the mock orientation for testing.
func (m *MockHeadset) SetHeadOrientation(pitch, yaw, roll float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pitch = pitch
	m.yaw = yaw
	m.roll = roll
}

// GetHeadPosition returns the mock position.
func (m *MockHeadset) GetHeadPosition() (x, y, z float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.posX, m.posY, m.posZ
}

// SetHeadPosition sets the mock position for testing.
func (m *MockHeadset) SetHeadPosition(x, y, z float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.posX = x
	m.posY = y
	m.posZ = z
}

// GetIPD returns the mock IPD.
func (m *MockHeadset) GetIPD() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ipd
}

// SetIPD sets the mock IPD for testing.
func (m *MockHeadset) SetIPD(ipd float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ipd = ipd
}

// HeadTrackingSystem manages VR head tracking by polling the headset adapter
// and updating entity components.
type HeadTrackingSystem struct {
	mu sync.RWMutex

	world   *World
	headset VRHeadsetAdapter

	// cameraUpdateCallback is called when head orientation changes
	cameraUpdateCallback func(pitch, yaw, roll float64)

	// positionUpdateCallback is called when head position changes
	positionUpdateCallback func(x, y, z float64)

	// enabled tracks if the system is active
	enabled bool

	// useMouseFallback enables mouse look when no headset is connected
	useMouseFallback bool

	// mouseX and mouseY track mouse position for fallback
	mouseX, mouseY float64

	// mouseSensitivity controls mouse look speed
	mouseSensitivity float64
}

// NewHeadTrackingSystem creates a new head tracking system.
func NewHeadTrackingSystem(world *World) *HeadTrackingSystem {
	log.WithFields(log.Fields{
		"system_name": "head_tracking",
	}).Debug("Creating head tracking system")

	return &HeadTrackingSystem{
		world:            world,
		enabled:          false,
		useMouseFallback: true,
		mouseSensitivity: 0.003,
	}
}

// SetHeadsetAdapter sets the VR headset adapter to poll.
func (s *HeadTrackingSystem) SetHeadsetAdapter(adapter VRHeadsetAdapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.headset = adapter
}

// GetHeadsetAdapter returns the current headset adapter.
func (s *HeadTrackingSystem) GetHeadsetAdapter() VRHeadsetAdapter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.headset
}

// Update processes head tracking for entities with HeadTrackingComponent.
func (s *HeadTrackingSystem) Update(entities []*Entity, deltaTime float64) {
	s.mu.RLock()
	if !s.enabled {
		s.mu.RUnlock()
		return
	}
	headset := s.headset
	useMouseFallback := s.useMouseFallback
	s.mu.RUnlock()

	for _, entity := range entities {
		comp, ok := entity.GetComponent("head_tracking")
		if !ok || comp == nil {
			continue
		}

		head, ok := comp.(*HeadTrackingComponent)
		if !ok {
			continue
		}

		if !head.IsEnabled() {
			continue
		}

		// Get tracking data from headset or fallback
		if headset != nil && headset.IsConnected() {
			s.updateFromHeadset(head, headset)
		} else if useMouseFallback {
			s.updateFromMouse(head, deltaTime)
		}

		// Notify callbacks
		s.notifyCallbacks(head)
	}
}

// updateFromHeadset updates the component from headset data.
func (s *HeadTrackingSystem) updateFromHeadset(head *HeadTrackingComponent, headset VRHeadsetAdapter) {
	pitch, yaw, roll := headset.GetHeadOrientation()
	head.SetOrientation(pitch, yaw, roll)

	x, y, z := headset.GetHeadPosition()
	head.SetPosition(x, y, z)
}

// updateFromMouse updates the component from mouse input (fallback mode).
func (s *HeadTrackingSystem) updateFromMouse(head *HeadTrackingComponent, deltaTime float64) {
	// Mouse fallback is handled externally via SetMousePosition
	// This method is a placeholder for integration
}

// notifyCallbacks calls registered callbacks with current head state.
func (s *HeadTrackingSystem) notifyCallbacks(head *HeadTrackingComponent) {
	pitch, yaw, roll := head.GetOrientation()
	x, y, z := head.GetPosition()

	s.mu.RLock()
	cameraCb := s.cameraUpdateCallback
	positionCb := s.positionUpdateCallback
	s.mu.RUnlock()

	if cameraCb != nil {
		cameraCb(pitch, yaw, roll)
	}
	if positionCb != nil {
		positionCb(x, y, z)
	}
}

// SetCameraUpdateCallback sets the callback for camera orientation changes.
func (s *HeadTrackingSystem) SetCameraUpdateCallback(cb func(pitch, yaw, roll float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cameraUpdateCallback = cb
}

// SetPositionUpdateCallback sets the callback for position changes.
func (s *HeadTrackingSystem) SetPositionUpdateCallback(cb func(x, y, z float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.positionUpdateCallback = cb
}

// SetEnabled enables or disables the head tracking system.
func (s *HeadTrackingSystem) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled

	log.WithFields(log.Fields{
		"system_name": "head_tracking",
		"enabled":     enabled,
	}).Info("Head tracking system toggled")
}

// IsEnabled returns whether the system is enabled.
func (s *HeadTrackingSystem) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetUseMouseFallback enables or disables mouse look fallback.
func (s *HeadTrackingSystem) SetUseMouseFallback(use bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.useMouseFallback = use
}

// IsUseMouseFallback returns whether mouse fallback is enabled.
func (s *HeadTrackingSystem) IsUseMouseFallback() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.useMouseFallback
}

// SetMouseSensitivity sets the mouse look sensitivity.
func (s *HeadTrackingSystem) SetMouseSensitivity(sensitivity float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sensitivity < 0.0001 {
		sensitivity = 0.0001
	}
	if sensitivity > 0.1 {
		sensitivity = 0.1
	}
	s.mouseSensitivity = sensitivity
}

// GetMouseSensitivity returns the mouse look sensitivity.
func (s *HeadTrackingSystem) GetMouseSensitivity() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mouseSensitivity
}

// SetMousePosition updates the mouse position for fallback mode.
// dx and dy are the mouse movement deltas since last frame.
func (s *HeadTrackingSystem) SetMousePosition(dx, dy float64) {
	s.mu.Lock()
	s.mouseX += dx * s.mouseSensitivity
	s.mouseY += dy * s.mouseSensitivity
	s.mu.Unlock()
}

// GetMouseOrientation returns the orientation from mouse movement.
// Returns (pitch, yaw) in radians.
func (s *HeadTrackingSystem) GetMouseOrientation() (pitch, yaw float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// mouseY controls pitch (inverted), mouseX controls yaw
	pitch = -s.mouseY
	yaw = s.mouseX

	// Clamp pitch
	if pitch > MaxPitch {
		pitch = MaxPitch
	}
	if pitch < MinPitch {
		pitch = MinPitch
	}

	// Normalize yaw
	yaw = normalizeAngle(yaw)

	return pitch, yaw
}

// HasHeadset returns true if a headset adapter is configured and connected.
func (s *HeadTrackingSystem) HasHeadset() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.headset != nil && s.headset.IsConnected()
}
