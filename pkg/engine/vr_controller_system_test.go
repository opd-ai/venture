package engine

import (
	"testing"
)

func TestNewMockController(t *testing.T) {
	mock := NewMockController()
	if mock == nil {
		t.Fatal("NewMockController returned nil")
	}

	// Both controllers should be connected by default
	if !mock.IsConnected(ControllerLeft) {
		t.Error("Left controller should be connected by default")
	}
	if !mock.IsConnected(ControllerRight) {
		t.Error("Right controller should be connected by default")
	}
}

func TestMockController_SetConnected(t *testing.T) {
	mock := NewMockController()

	mock.SetConnected(ControllerLeft, false)
	if mock.IsConnected(ControllerLeft) {
		t.Error("Left controller should be disconnected")
	}

	mock.SetConnected(ControllerLeft, true)
	if !mock.IsConnected(ControllerLeft) {
		t.Error("Left controller should be connected")
	}
}

func TestMockController_Trigger(t *testing.T) {
	mock := NewMockController()

	mock.SetTrigger(ControllerRight, 0.7)
	if mock.GetTrigger(ControllerRight) != 0.7 {
		t.Errorf("Expected trigger 0.7, got %v", mock.GetTrigger(ControllerRight))
	}
}

func TestMockController_Grip(t *testing.T) {
	mock := NewMockController()

	mock.SetGrip(ControllerLeft, 0.5)
	if mock.GetGrip(ControllerLeft) != 0.5 {
		t.Errorf("Expected grip 0.5, got %v", mock.GetGrip(ControllerLeft))
	}
}

func TestMockController_Thumbstick(t *testing.T) {
	mock := NewMockController()

	mock.SetThumbstick(ControllerRight, 0.5, -0.3)
	x, y := mock.GetThumbstick(ControllerRight)

	if x != 0.5 {
		t.Errorf("Expected X 0.5, got %v", x)
	}
	if y != -0.3 {
		t.Errorf("Expected Y -0.3, got %v", y)
	}
}

func TestMockController_ThumbstickPressed(t *testing.T) {
	mock := NewMockController()

	mock.SetThumbstickPressed(ControllerRight, true)
	if !mock.IsThumbstickPressed(ControllerRight) {
		t.Error("Expected thumbstick pressed")
	}

	mock.SetThumbstickPressed(ControllerRight, false)
	if mock.IsThumbstickPressed(ControllerRight) {
		t.Error("Expected thumbstick not pressed")
	}
}

func TestMockController_Button(t *testing.T) {
	mock := NewMockController()

	mock.SetButton(ControllerRight, ButtonA, true)
	if !mock.GetButton(ControllerRight, ButtonA) {
		t.Error("Expected button A pressed")
	}

	mock.SetButton(ControllerRight, ButtonA, false)
	if mock.GetButton(ControllerRight, ButtonA) {
		t.Error("Expected button A not pressed")
	}
}

func TestMockController_Haptic(t *testing.T) {
	mock := NewMockController()

	mock.SetHaptic(ControllerRight, 0.8, 0.5)
	intensity, duration := mock.GetLastHaptic(ControllerRight)

	if intensity != 0.8 {
		t.Errorf("Expected intensity 0.8, got %v", intensity)
	}
	if duration != 0.5 {
		t.Errorf("Expected duration 0.5, got %v", duration)
	}
}

func TestNewVRControllerSystem(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)

	if sys == nil {
		t.Fatal("NewVRControllerSystem returned nil")
	}

	if !sys.IsEnabled() {
		t.Error("Expected enabled by default")
	}
}

func TestVRControllerSystem_SetEnabled(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)

	sys.SetEnabled(true)
	if !sys.IsEnabled() {
		t.Error("Expected enabled")
	}

	sys.SetEnabled(false)
	if sys.IsEnabled() {
		t.Error("Expected disabled")
	}
}

func TestVRControllerSystem_SetControllerAdapter(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)

	mock := NewMockController()
	sys.SetControllerAdapter(mock)

	if sys.GetControllerAdapter() != mock {
		t.Error("Controller adapter not set correctly")
	}
}

func TestVRControllerSystem_HasController(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)

	// No adapter
	if sys.HasController(ControllerRight) {
		t.Error("Expected no controller")
	}

	// Connected adapter
	mock := NewMockController()
	sys.SetControllerAdapter(mock)
	if !sys.HasController(ControllerRight) {
		t.Error("Expected right controller connected")
	}

	// Disconnected
	mock.SetConnected(ControllerRight, false)
	if sys.HasController(ControllerRight) {
		t.Error("Expected right controller disconnected")
	}
}

func TestVRControllerSystem_Update_FromAdapter(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	mock.SetTrigger(ControllerRight, 0.7)
	mock.SetGrip(ControllerRight, 0.5)
	mock.SetThumbstick(ControllerRight, 0.3, -0.4)
	mock.SetButton(ControllerRight, ButtonA, true)
	sys.SetControllerAdapter(mock)

	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	entity.AddComponent(ctrl)

	sys.Update([]*Entity{entity}, 0.016)

	// Verify values were updated
	if ctrl.GetTrigger() != 0.7 {
		t.Errorf("Expected trigger 0.7, got %v", ctrl.GetTrigger())
	}
	if ctrl.GetGrip() != 0.5 {
		t.Errorf("Expected grip 0.5, got %v", ctrl.GetGrip())
	}

	x, y := ctrl.GetThumbstickRaw()
	if x != 0.3 {
		t.Errorf("Expected thumbstick X 0.3, got %v", x)
	}
	if y != -0.4 {
		t.Errorf("Expected thumbstick Y -0.4, got %v", y)
	}

	if !ctrl.IsButtonAPressed() {
		t.Error("Expected button A pressed")
	}
}

func TestVRControllerSystem_Update_Disabled(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(false) // Explicitly disable for this test

	mock := NewMockController()
	mock.SetTrigger(ControllerRight, 0.7)
	sys.SetControllerAdapter(mock)

	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	entity.AddComponent(ctrl)

	sys.Update([]*Entity{entity}, 0.016)

	// Trigger should not be updated
	if ctrl.GetTrigger() != 0 {
		t.Error("Trigger should not be updated when system disabled")
	}
}

func TestVRControllerSystem_Callbacks(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	sys.SetControllerAdapter(mock)

	// Track callbacks
	attackCalled := false
	var attackHand string
	sys.SetAttackCallback(func(hand string) {
		attackCalled = true
		attackHand = hand
	})

	// Create entity with trigger pressed
	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	entity.AddComponent(ctrl)

	// First update - trigger not pressed
	sys.Update([]*Entity{entity}, 0.016)

	// Second update - trigger pressed (edge detection)
	mock.SetTrigger(ControllerRight, 0.8)
	sys.Update([]*Entity{entity}, 0.016)

	if !attackCalled {
		t.Error("Attack callback should be called")
	}
	if attackHand != ControllerRight {
		t.Errorf("Expected hand %s, got %s", ControllerRight, attackHand)
	}
}

func TestVRControllerSystem_MovementCallback(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	sys.SetControllerAdapter(mock)

	moveCalled := false
	var moveX, moveY float64
	sys.SetMovementCallback(func(x, y float64) {
		moveCalled = true
		moveX = x
		moveY = y
	})

	// Create LEFT controller entity for movement
	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerLeft)
	ctrl.SetEnabled(true)
	ctrl.SetDeadZone(0.1) // Low dead zone
	entity.AddComponent(ctrl)

	mock.SetThumbstick(ControllerLeft, 0.5, -0.3)
	sys.Update([]*Entity{entity}, 0.016)

	if !moveCalled {
		t.Error("Movement callback should be called")
	}
	if moveX != 0.5 {
		t.Errorf("Expected moveX 0.5, got %v", moveX)
	}
	if moveY != -0.3 {
		t.Errorf("Expected moveY -0.3, got %v", moveY)
	}
}

func TestVRControllerSystem_TurnCallback(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	sys.SetControllerAdapter(mock)

	turnCalled := false
	var turnDir float64
	sys.SetTurnCallback(func(direction float64) {
		turnCalled = true
		turnDir = direction
	})

	// Create RIGHT controller entity for turning
	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	ctrl.SetDeadZone(0.1)
	entity.AddComponent(ctrl)

	mock.SetThumbstick(ControllerRight, 0.7, 0)
	sys.Update([]*Entity{entity}, 0.016)

	if !turnCalled {
		t.Error("Turn callback should be called")
	}
	if turnDir != 0.7 {
		t.Errorf("Expected turnDir 0.7, got %v", turnDir)
	}
}

func TestVRControllerSystem_MenuCallback(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	sys.SetControllerAdapter(mock)

	menuCalled := false
	sys.SetMenuCallback(func(hand string) {
		menuCalled = true
	})

	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	entity.AddComponent(ctrl)

	// First update without menu button
	sys.Update([]*Entity{entity}, 0.016)

	// Second update with menu button pressed
	mock.SetButton(ControllerRight, ButtonMenu, true)
	sys.Update([]*Entity{entity}, 0.016)

	if !menuCalled {
		t.Error("Menu callback should be called")
	}
}

func TestVRControllerSystem_SetButtonMappings(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)

	sys.SetAttackButton(ButtonA)
	sys.SetInteractButton(ButtonB)

	// Verify internal state changed (tested via Update behavior)
	// This is a smoke test for the setters
}

func TestVRControllerSystem_HapticFeedback(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	sys.SetControllerAdapter(mock)

	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	entity.AddComponent(ctrl)

	// Trigger haptic
	sys.TriggerHaptic(ctrl, 0.8, 0.5)
	sys.Update([]*Entity{entity}, 0.016)

	// Check haptic was sent to adapter
	intensity, duration := mock.GetLastHaptic(ControllerRight)
	if intensity != 0.8 {
		t.Errorf("Expected haptic intensity 0.8, got %v", intensity)
	}
	if duration < 0.4 { // Some duration consumed
		t.Errorf("Expected haptic duration ~0.5, got %v", duration)
	}
}

func TestVRControllerSystem_ThreadSafety(t *testing.T) {
	world := &World{}
	sys := NewVRControllerSystem(world)

	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 1000; i++ {
			sys.SetEnabled(i%2 == 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sys.IsEnabled()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			sys.SetAttackCallback(func(hand string) {})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sys.HasController(ControllerRight)
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func BenchmarkVRControllerSystem_Update(b *testing.B) {
	world := &World{}
	sys := NewVRControllerSystem(world)
	sys.SetEnabled(true)

	mock := NewMockController()
	mock.SetTrigger(ControllerRight, 0.7)
	mock.SetThumbstick(ControllerRight, 0.5, -0.3)
	sys.SetControllerAdapter(mock)

	entity := NewEntity(1)
	ctrl := NewVRControllerComponent(ControllerRight)
	ctrl.SetEnabled(true)
	entity.AddComponent(ctrl)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
