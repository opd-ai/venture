package engine

import (
	"encoding/json"
	"testing"
)

func TestNewVRControllerComponent(t *testing.T) {
	tests := []struct {
		name     string
		hand     string
		expected string
	}{
		{"left hand", ControllerLeft, ControllerLeft},
		{"right hand", ControllerRight, ControllerRight},
		{"invalid defaults to right", "invalid", ControllerRight},
		{"empty defaults to right", "", ControllerRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewVRControllerComponent(tt.hand)
			if c == nil {
				t.Fatal("NewVRControllerComponent returned nil")
			}
			if c.GetHand() != tt.expected {
				t.Errorf("Expected hand %s, got %s", tt.expected, c.GetHand())
			}
			if c.IsEnabled() {
				t.Error("Expected disabled by default")
			}
			if c.GetDeadZone() != DefaultDeadZone {
				t.Errorf("Expected dead zone %v, got %v", DefaultDeadZone, c.GetDeadZone())
			}
		})
	}
}

func TestVRControllerComponent_Type(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)
	if c.Type() != "vr_controller" {
		t.Errorf("Expected type 'vr_controller', got '%s'", c.Type())
	}
}

func TestVRControllerComponent_SetEnabled(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("Expected enabled")
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("Expected disabled")
	}
}

func TestVRControllerComponent_SetTrigger(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 0.5, 0.5},
		{"zero", 0.0, 0.0},
		{"max", 1.0, 1.0},
		{"negative clamp", -0.5, 0.0},
		{"above max clamp", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetTrigger(tt.input)
			if c.GetTrigger() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, c.GetTrigger())
			}
		})
	}
}

func TestVRControllerComponent_IsTriggerPressed(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetTrigger(0.3)
	if c.IsTriggerPressed() {
		t.Error("Trigger should not be pressed at 0.3")
	}

	c.SetTrigger(0.6)
	if !c.IsTriggerPressed() {
		t.Error("Trigger should be pressed at 0.6")
	}
}

func TestVRControllerComponent_IsTriggerJustPressed(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	// Initial state - not pressed
	c.SetTrigger(0.3)
	if c.IsTriggerJustPressed() {
		t.Error("Should not be just pressed initially")
	}

	// Press trigger
	c.SetTrigger(0.7)
	if !c.IsTriggerJustPressed() {
		t.Error("Should be just pressed when transitioning from low to high")
	}

	// Keep pressed
	c.SetTrigger(0.8)
	if c.IsTriggerJustPressed() {
		t.Error("Should not be just pressed when already pressed")
	}
}

func TestVRControllerComponent_SetGrip(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetGrip(0.7)
	if c.GetGrip() != 0.7 {
		t.Errorf("Expected 0.7, got %v", c.GetGrip())
	}

	if !c.IsGripPressed() {
		t.Error("Grip should be pressed at 0.7")
	}

	c.SetGrip(0.3)
	if c.IsGripPressed() {
		t.Error("Grip should not be pressed at 0.3")
	}
}

func TestVRControllerComponent_SetThumbstick(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetThumbstick(0.5, -0.3)
	x, y := c.GetThumbstickRaw()

	if x != 0.5 {
		t.Errorf("Expected X 0.5, got %v", x)
	}
	if y != -0.3 {
		t.Errorf("Expected Y -0.3, got %v", y)
	}
}

func TestVRControllerComponent_ThumbstickDeadZone(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)
	c.SetDeadZone(0.2)

	// Value within dead zone
	c.SetThumbstick(0.1, 0.1)
	x, y := c.GetThumbstick()

	if x != 0 {
		t.Errorf("Expected X 0 (dead zone), got %v", x)
	}
	if y != 0 {
		t.Errorf("Expected Y 0 (dead zone), got %v", y)
	}

	// Value outside dead zone
	c.SetThumbstick(0.5, -0.5)
	x, y = c.GetThumbstick()

	if x != 0.5 {
		t.Errorf("Expected X 0.5, got %v", x)
	}
	if y != -0.5 {
		t.Errorf("Expected Y -0.5, got %v", y)
	}
}

func TestVRControllerComponent_ThumbstickClamping(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetThumbstick(2.0, -2.0)
	x, y := c.GetThumbstickRaw()

	if x != 1.0 {
		t.Errorf("Expected X clamped to 1.0, got %v", x)
	}
	if y != -1.0 {
		t.Errorf("Expected Y clamped to -1.0, got %v", y)
	}
}

func TestVRControllerComponent_SetThumbstickPressed(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetThumbstickPressed(true)
	if !c.IsThumbstickPressed() {
		t.Error("Expected thumbstick pressed")
	}

	c.SetThumbstickPressed(false)
	if c.IsThumbstickPressed() {
		t.Error("Expected thumbstick not pressed")
	}
}

func TestVRControllerComponent_ButtonA(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetButtonA(true)
	if !c.IsButtonAPressed() {
		t.Error("Expected button A pressed")
	}
	if !c.IsButtonAJustPressed() {
		t.Error("Expected button A just pressed")
	}

	// Keep pressed
	c.SetButtonA(true)
	if c.IsButtonAJustPressed() {
		t.Error("Should not be just pressed when already pressed")
	}

	c.SetButtonA(false)
	if c.IsButtonAPressed() {
		t.Error("Expected button A not pressed")
	}
}

func TestVRControllerComponent_ButtonB(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetButtonB(true)
	if !c.IsButtonBPressed() {
		t.Error("Expected button B pressed")
	}
	if !c.IsButtonBJustPressed() {
		t.Error("Expected button B just pressed")
	}

	c.SetButtonB(false)
	if c.IsButtonBPressed() {
		t.Error("Expected button B not pressed")
	}
}

func TestVRControllerComponent_MenuButton(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetMenuButton(true)
	if !c.IsMenuButtonPressed() {
		t.Error("Expected menu button pressed")
	}
	if !c.IsMenuButtonJustPressed() {
		t.Error("Expected menu button just pressed")
	}

	c.SetMenuButton(false)
	if c.IsMenuButtonPressed() {
		t.Error("Expected menu button not pressed")
	}
}

func TestVRControllerComponent_SetHaptic(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetHaptic(0.8, 0.5)
	intensity, duration := c.GetHaptic()

	if intensity != 0.8 {
		t.Errorf("Expected intensity 0.8, got %v", intensity)
	}
	if duration != 0.5 {
		t.Errorf("Expected duration 0.5, got %v", duration)
	}
}

func TestVRControllerComponent_SetHapticClamping(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	// Test intensity clamping
	c.SetHaptic(1.5, 0.5)
	intensity, _ := c.GetHaptic()
	if intensity != 1.0 {
		t.Errorf("Expected intensity clamped to 1.0, got %v", intensity)
	}

	c.SetHaptic(-0.5, 0.5)
	intensity, _ = c.GetHaptic()
	if intensity != 0.0 {
		t.Errorf("Expected intensity clamped to 0.0, got %v", intensity)
	}

	// Test negative duration
	c.SetHaptic(0.5, -1.0)
	_, duration := c.GetHaptic()
	if duration != 0.0 {
		t.Errorf("Expected duration clamped to 0.0, got %v", duration)
	}
}

func TestVRControllerComponent_UpdateHaptic(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	c.SetHaptic(0.8, 0.5)
	c.UpdateHaptic(0.2)

	_, duration := c.GetHaptic()
	if duration != 0.3 {
		t.Errorf("Expected duration 0.3, got %v", duration)
	}

	// Complete haptic
	c.UpdateHaptic(0.5)
	intensity, duration := c.GetHaptic()
	if duration != 0.0 {
		t.Errorf("Expected duration 0.0, got %v", duration)
	}
	if intensity != 0.0 {
		t.Errorf("Expected intensity reset to 0.0, got %v", intensity)
	}
}

func TestVRControllerComponent_SetDeadZone(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 0.2, 0.2},
		{"zero", 0.0, 0.0},
		{"max", 0.5, 0.5},
		{"negative clamp", -0.1, 0.0},
		{"above max clamp", 0.6, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetDeadZone(tt.input)
			if c.GetDeadZone() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, c.GetDeadZone())
			}
		})
	}
}

func TestVRControllerComponent_Serialize(t *testing.T) {
	c := NewVRControllerComponent(ControllerLeft)
	c.SetEnabled(true)
	c.SetTrigger(0.7)
	c.SetGrip(0.5)
	c.SetThumbstick(0.3, -0.4)
	c.SetButtonA(true)

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if m["enabled"] != true {
		t.Error("Expected enabled true")
	}
	if m["hand"] != ControllerLeft {
		t.Errorf("Expected hand %s", ControllerLeft)
	}
}

func TestVRControllerComponent_Deserialize(t *testing.T) {
	original := NewVRControllerComponent(ControllerLeft)
	original.SetEnabled(true)
	original.SetTrigger(0.7)
	original.SetGrip(0.5)
	original.SetThumbstick(0.3, -0.4)
	original.SetButtonA(true)
	original.SetDeadZone(0.25)

	data, _ := original.Serialize()

	restored := NewVRControllerComponent(ControllerRight)
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.IsEnabled() != original.IsEnabled() {
		t.Error("Enabled mismatch")
	}
	if restored.GetHand() != original.GetHand() {
		t.Error("Hand mismatch")
	}
	if restored.GetTrigger() != original.GetTrigger() {
		t.Error("Trigger mismatch")
	}
	if restored.GetDeadZone() != original.GetDeadZone() {
		t.Error("DeadZone mismatch")
	}
}

func TestVRControllerComponent_ThreadSafety(t *testing.T) {
	c := NewVRControllerComponent(ControllerRight)

	done := make(chan bool, 6)

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetTrigger(float64(i) / 1000)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = c.GetTrigger()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetThumbstick(float64(i)/500-1, float64(i)/500-1)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.GetThumbstick()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetButtonA(i%2 == 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = c.IsButtonAJustPressed()
		}
		done <- true
	}()

	for i := 0; i < 6; i++ {
		<-done
	}
}

func BenchmarkVRControllerComponent_SetTrigger(b *testing.B) {
	c := NewVRControllerComponent(ControllerRight)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetTrigger(float64(i%100) / 100)
	}
}

func BenchmarkVRControllerComponent_GetThumbstick(b *testing.B) {
	c := NewVRControllerComponent(ControllerRight)
	c.SetThumbstick(0.5, -0.3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetThumbstick()
	}
}
