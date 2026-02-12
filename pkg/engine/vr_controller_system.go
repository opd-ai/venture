// Package engine provides the VR controller system for VR support.

package engine

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// MockController provides a test implementation of VRControllerAdapter.
type MockController struct {
	mu sync.RWMutex

	connected  map[string]bool
	triggers   map[string]float64
	grips      map[string]float64
	thumbX     map[string]float64
	thumbY     map[string]float64
	thumbPress map[string]bool
	buttons    map[string]map[string]bool
	haptics    map[string]struct{ intensity, duration float64 }
}

// NewMockController creates a mock controller for testing.
func NewMockController() *MockController {
	return &MockController{
		connected:  map[string]bool{ControllerLeft: true, ControllerRight: true},
		triggers:   make(map[string]float64),
		grips:      make(map[string]float64),
		thumbX:     make(map[string]float64),
		thumbY:     make(map[string]float64),
		thumbPress: make(map[string]bool),
		buttons: map[string]map[string]bool{
			ControllerLeft:  make(map[string]bool),
			ControllerRight: make(map[string]bool),
		},
		haptics: make(map[string]struct{ intensity, duration float64 }),
	}
}

// IsConnected returns true if the mock controller is connected.
func (m *MockController) IsConnected(hand string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected[hand]
}

// SetConnected sets connection status for testing.
func (m *MockController) SetConnected(hand string, connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected[hand] = connected
}

// GetTrigger returns the mock trigger value.
func (m *MockController) GetTrigger(hand string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.triggers[hand]
}

// SetTrigger sets mock trigger value for testing.
func (m *MockController) SetTrigger(hand string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.triggers[hand] = value
}

// GetGrip returns the mock grip value.
func (m *MockController) GetGrip(hand string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.grips[hand]
}

// SetGrip sets mock grip value for testing.
func (m *MockController) SetGrip(hand string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.grips[hand] = value
}

// GetThumbstick returns the mock thumbstick values.
func (m *MockController) GetThumbstick(hand string) (x, y float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thumbX[hand], m.thumbY[hand]
}

// SetThumbstick sets mock thumbstick values for testing.
func (m *MockController) SetThumbstick(hand string, x, y float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thumbX[hand] = x
	m.thumbY[hand] = y
}

// IsThumbstickPressed returns mock thumbstick press state.
func (m *MockController) IsThumbstickPressed(hand string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thumbPress[hand]
}

// SetThumbstickPressed sets mock thumbstick press for testing.
func (m *MockController) SetThumbstickPressed(hand string, pressed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thumbPress[hand] = pressed
}

// GetButton returns mock button state.
func (m *MockController) GetButton(hand, button string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.buttons[hand][button]
}

// SetButton sets mock button state for testing.
func (m *MockController) SetButton(hand, button string, pressed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buttons[hand] == nil {
		m.buttons[hand] = make(map[string]bool)
	}
	m.buttons[hand][button] = pressed
}

// SetHaptic stores haptic request for verification.
func (m *MockController) SetHaptic(hand string, intensity, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.haptics[hand] = struct{ intensity, duration float64 }{intensity, duration}
}

// GetLastHaptic returns the last haptic request for testing.
func (m *MockController) GetLastHaptic(hand string) (intensity, duration float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.haptics[hand]
	return h.intensity, h.duration
}

// VRControllerSystem manages VR controller input by polling the adapter
// and updating entity components.
type VRControllerSystem struct {
	mu sync.RWMutex

	world      *World
	controller VRControllerAdapter

	// Action callbacks
	attackCallback   func(hand string)
	interactCallback func(hand string)
	menuCallback     func(hand string)
	movementCallback func(x, y float64)
	turnCallback     func(direction float64)

	// enabled tracks if the system is active
	enabled bool

	// Button mappings
	attackButton   string
	interactButton string
}

// Button name constants
const (
	ButtonA          = "A"
	ButtonB          = "B"
	ButtonMenu       = "Menu"
	ButtonTrigger    = "Trigger"
	ButtonGrip       = "Grip"
	ButtonThumbstick = "Thumbstick"
)

// NewVRControllerSystem creates a new VR controller system.
func NewVRControllerSystem(world *World) *VRControllerSystem {
	log.WithFields(log.Fields{
		"system_name": "vr_controller",
	}).Debug("Creating VR controller system")

	return &VRControllerSystem{
		world:          world,
		enabled:        true, // Enabled by default with graceful degradation when no VR controllers
		attackButton:   ButtonTrigger,
		interactButton: ButtonA,
	}
}

// SetControllerAdapter sets the VR controller adapter to poll.
func (s *VRControllerSystem) SetControllerAdapter(adapter VRControllerAdapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controller = adapter
}

// GetControllerAdapter returns the current controller adapter.
func (s *VRControllerSystem) GetControllerAdapter() VRControllerAdapter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.controller
}

// Update processes VR controller input for entities with VRControllerComponent.
func (s *VRControllerSystem) Update(entities []*Entity, deltaTime float64) {
	controller := s.getEnabledController()
	if controller == nil {
		return
	}

	for _, entity := range entities {
		s.updateVRControllerEntity(entity, controller, deltaTime)
	}
}

// getEnabledController returns the controller if enabled, nil otherwise.
func (s *VRControllerSystem) getEnabledController() VRControllerAdapter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.enabled {
		return nil
	}
	return s.controller
}

// updateVRControllerEntity updates a single VR controller entity.
func (s *VRControllerSystem) updateVRControllerEntity(entity *Entity, controller VRControllerAdapter, deltaTime float64) {
	ctrl, ok := s.getVRControllerComponent(entity)
	if !ok || !ctrl.IsEnabled() {
		return
	}

	if controller != nil && controller.IsConnected(ctrl.GetHand()) {
		s.updateFromAdapter(ctrl, controller)
	}

	ctrl.UpdateHaptic(deltaTime)
	s.processActions(ctrl)

	if controller != nil {
		s.sendHaptics(ctrl, controller)
	}
}

// getVRControllerComponent retrieves and validates the VR controller component.
func (s *VRControllerSystem) getVRControllerComponent(entity *Entity) (*VRControllerComponent, bool) {
	comp, ok := entity.GetComponent("vr_controller")
	if !ok || comp == nil {
		return nil, false
	}

	ctrl, ok := comp.(*VRControllerComponent)
	if !ok {
		return nil, false
	}

	return ctrl, true
}

// updateFromAdapter updates the component from controller adapter data.
func (s *VRControllerSystem) updateFromAdapter(ctrl *VRControllerComponent, adapter VRControllerAdapter) {
	hand := ctrl.GetHand()

	// Update analog inputs
	ctrl.SetTrigger(adapter.GetTrigger(hand))
	ctrl.SetGrip(adapter.GetGrip(hand))

	x, y := adapter.GetThumbstick(hand)
	ctrl.SetThumbstick(x, y)
	ctrl.SetThumbstickPressed(adapter.IsThumbstickPressed(hand))

	// Update buttons
	ctrl.SetButtonA(adapter.GetButton(hand, ButtonA))
	ctrl.SetButtonB(adapter.GetButton(hand, ButtonB))
	ctrl.SetMenuButton(adapter.GetButton(hand, ButtonMenu))
}

// processActions checks for input events and triggers callbacks.
func (s *VRControllerSystem) processActions(ctrl *VRControllerComponent) {
	callbacks := s.getCallbacks()
	hand := ctrl.GetHand()

	s.processAttackButton(ctrl, callbacks, hand)
	s.processInteractButton(ctrl, callbacks, hand)
	s.processMenuButton(ctrl, callbacks, hand)
	s.processMovement(ctrl, callbacks, hand)
	s.processTurn(ctrl, callbacks, hand)
}

// callbackSet holds all controller callbacks for action processing.
type callbackSet struct {
	attack      func(string)
	interact    func(string)
	menu        func(string)
	movement    func(float64, float64)
	turn        func(float64)
	attackBtn   string
	interactBtn string
}

// getCallbacks retrieves all callbacks and button mappings from the system.
func (s *VRControllerSystem) getCallbacks() callbackSet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return callbackSet{
		attack:      s.attackCallback,
		interact:    s.interactCallback,
		menu:        s.menuCallback,
		movement:    s.movementCallback,
		turn:        s.turnCallback,
		attackBtn:   s.attackButton,
		interactBtn: s.interactButton,
	}
}

// processAttackButton checks and triggers the attack button callback.
func (s *VRControllerSystem) processAttackButton(ctrl *VRControllerComponent, cb callbackSet, hand string) {
	if cb.attack == nil {
		return
	}

	if s.isButtonPressed(ctrl, cb.attackBtn) {
		cb.attack(hand)
	}
}

// processInteractButton checks and triggers the interact button callback.
func (s *VRControllerSystem) processInteractButton(ctrl *VRControllerComponent, cb callbackSet, hand string) {
	if cb.interact == nil {
		return
	}

	if s.isButtonPressed(ctrl, cb.interactBtn) {
		cb.interact(hand)
	}
}

// isButtonPressed checks if a specific button was just pressed.
func (s *VRControllerSystem) isButtonPressed(ctrl *VRControllerComponent, btn string) bool {
	switch btn {
	case ButtonTrigger:
		return ctrl.IsTriggerJustPressed()
	case ButtonA:
		return ctrl.IsButtonAJustPressed()
	case ButtonB:
		return ctrl.IsButtonBJustPressed()
	default:
		return false
	}
}

// processMenuButton checks and triggers the menu button callback.
func (s *VRControllerSystem) processMenuButton(ctrl *VRControllerComponent, cb callbackSet, hand string) {
	if cb.menu != nil && ctrl.IsMenuButtonJustPressed() {
		cb.menu(hand)
	}
}

// processMovement handles left thumbstick movement input.
func (s *VRControllerSystem) processMovement(ctrl *VRControllerComponent, cb callbackSet, hand string) {
	if cb.movement == nil || hand != ControllerLeft {
		return
	}

	x, y := ctrl.GetThumbstick()
	if x != 0 || y != 0 {
		cb.movement(x, y)
	}
}

// processTurn handles right thumbstick turning input.
func (s *VRControllerSystem) processTurn(ctrl *VRControllerComponent, cb callbackSet, hand string) {
	if cb.turn == nil || hand != ControllerRight {
		return
	}

	x, _ := ctrl.GetThumbstick()
	if x != 0 {
		cb.turn(x)
	}
}

// sendHaptics sends pending haptic feedback to the adapter.
func (s *VRControllerSystem) sendHaptics(ctrl *VRControllerComponent, adapter VRControllerAdapter) {
	intensity, duration := ctrl.GetHaptic()
	if intensity > 0 && duration > 0 {
		adapter.SetHaptic(ctrl.GetHand(), intensity, duration)
	}
}

// SetAttackCallback sets the callback for attack actions.
func (s *VRControllerSystem) SetAttackCallback(cb func(hand string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attackCallback = cb
}

// SetInteractCallback sets the callback for interact actions.
func (s *VRControllerSystem) SetInteractCallback(cb func(hand string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interactCallback = cb
}

// SetMenuCallback sets the callback for menu button.
func (s *VRControllerSystem) SetMenuCallback(cb func(hand string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.menuCallback = cb
}

// SetMovementCallback sets the callback for thumbstick movement.
func (s *VRControllerSystem) SetMovementCallback(cb func(x, y float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.movementCallback = cb
}

// SetTurnCallback sets the callback for turning (right thumbstick X).
func (s *VRControllerSystem) SetTurnCallback(cb func(direction float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.turnCallback = cb
}

// SetAttackButton sets which button triggers attack.
func (s *VRControllerSystem) SetAttackButton(button string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attackButton = button
}

// SetInteractButton sets which button triggers interact.
func (s *VRControllerSystem) SetInteractButton(button string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interactButton = button
}

// SetEnabled enables or disables the controller system.
func (s *VRControllerSystem) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled

	log.WithFields(log.Fields{
		"system_name": "vr_controller",
		"enabled":     enabled,
	}).Info("VR controller system toggled")
}

// IsEnabled returns whether the system is enabled.
func (s *VRControllerSystem) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// HasController returns true if a controller adapter is configured and connected.
func (s *VRControllerSystem) HasController(hand string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.controller != nil && s.controller.IsConnected(hand)
}

// TriggerHaptic triggers haptic feedback on a controller component.
func (s *VRControllerSystem) TriggerHaptic(ctrl *VRControllerComponent, intensity, duration float64) {
	ctrl.SetHaptic(intensity, duration)
}
