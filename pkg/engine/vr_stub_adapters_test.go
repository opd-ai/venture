package engine

import (
	"testing"
)

func TestNewStubHeadsetAdapter(t *testing.T) {
	adapter := NewStubHeadsetAdapter()

	if adapter == nil {
		t.Fatal("NewStubHeadsetAdapter returned nil")
	}
}

func TestStubHeadsetAdapter_IsConnected(t *testing.T) {
	adapter := NewStubHeadsetAdapter()

	if adapter.IsConnected() {
		t.Error("Stub adapter should report not connected")
	}
}

func TestStubHeadsetAdapter_GetHeadOrientation(t *testing.T) {
	adapter := NewStubHeadsetAdapter()

	pitch, yaw, roll := adapter.GetHeadOrientation()

	if pitch != 0 || yaw != 0 || roll != 0 {
		t.Errorf("Expected zero orientation, got pitch=%v yaw=%v roll=%v", pitch, yaw, roll)
	}
}

func TestStubHeadsetAdapter_GetHeadPosition(t *testing.T) {
	adapter := NewStubHeadsetAdapter()

	x, y, z := adapter.GetHeadPosition()

	if x != 0 || y != 0 || z != 0 {
		t.Errorf("Expected zero position, got x=%v y=%v z=%v", x, y, z)
	}
}

func TestStubHeadsetAdapter_GetIPD(t *testing.T) {
	adapter := NewStubHeadsetAdapter()

	ipd := adapter.GetIPD()

	expectedIPD := 63.0
	if ipd != expectedIPD {
		t.Errorf("Expected IPD=%v, got %v", expectedIPD, ipd)
	}
}

func TestStubHeadsetAdapter_InterfaceCompliance(t *testing.T) {
	var _ VRHeadsetAdapter = (*StubHeadsetAdapter)(nil)
}

func TestNewStubControllerAdapter(t *testing.T) {
	adapter := NewStubControllerAdapter()

	if adapter == nil {
		t.Fatal("NewStubControllerAdapter returned nil")
	}
}

func TestStubControllerAdapter_IsConnected(t *testing.T) {
	adapter := NewStubControllerAdapter()

	tests := []struct {
		name string
		hand string
	}{
		{"left hand", ControllerLeft},
		{"right hand", ControllerRight},
		{"unknown hand", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if adapter.IsConnected(tt.hand) {
				t.Errorf("Stub adapter should report not connected for %s", tt.hand)
			}
		})
	}
}

func TestStubControllerAdapter_GetTrigger(t *testing.T) {
	adapter := NewStubControllerAdapter()

	tests := []string{ControllerLeft, ControllerRight}

	for _, hand := range tests {
		trigger := adapter.GetTrigger(hand)
		if trigger != 0.0 {
			t.Errorf("Expected trigger=0.0 for %s, got %v", hand, trigger)
		}
	}
}

func TestStubControllerAdapter_GetGrip(t *testing.T) {
	adapter := NewStubControllerAdapter()

	tests := []string{ControllerLeft, ControllerRight}

	for _, hand := range tests {
		grip := adapter.GetGrip(hand)
		if grip != 0.0 {
			t.Errorf("Expected grip=0.0 for %s, got %v", hand, grip)
		}
	}
}

func TestStubControllerAdapter_GetThumbstick(t *testing.T) {
	adapter := NewStubControllerAdapter()

	tests := []string{ControllerLeft, ControllerRight}

	for _, hand := range tests {
		x, y := adapter.GetThumbstick(hand)
		if x != 0 || y != 0 {
			t.Errorf("Expected thumbstick=(0,0) for %s, got (%v,%v)", hand, x, y)
		}
	}
}

func TestStubControllerAdapter_IsThumbstickPressed(t *testing.T) {
	adapter := NewStubControllerAdapter()

	tests := []string{ControllerLeft, ControllerRight}

	for _, hand := range tests {
		if adapter.IsThumbstickPressed(hand) {
			t.Errorf("Expected thumbstick not pressed for %s", hand)
		}
	}
}

func TestStubControllerAdapter_GetButton(t *testing.T) {
	adapter := NewStubControllerAdapter()

	buttons := []string{ButtonA, ButtonB, ButtonMenu, ButtonTrigger, ButtonThumbstick}
	hands := []string{ControllerLeft, ControllerRight}

	for _, hand := range hands {
		for _, button := range buttons {
			if adapter.GetButton(hand, button) {
				t.Errorf("Expected button %s not pressed for %s", button, hand)
			}
		}
	}
}

func TestStubControllerAdapter_SetHaptic(t *testing.T) {
	adapter := NewStubControllerAdapter()

	// Test that haptic trigger doesn't panic (it's a no-op)
	hands := []string{ControllerLeft, ControllerRight}

	for _, hand := range hands {
		adapter.SetHaptic(hand, 0.5, 0.1)
		adapter.SetHaptic(hand, 1.0, 0.5)
		adapter.SetHaptic(hand, 0.0, 0.0)
	}

	// If we get here without panic, test passes
}

func TestStubControllerAdapter_InterfaceCompliance(t *testing.T) {
	var _ VRControllerAdapter = (*StubControllerAdapter)(nil)
}

func TestStubAdapters_WithVRSystems(t *testing.T) {
	// Integration test: verify stub adapters work with VR systems

	world := &World{}

	// Test with HeadTrackingSystem
	headSystem := NewHeadTrackingSystem(world)
	stubHeadset := NewStubHeadsetAdapter()
	headSystem.SetHeadsetAdapter(stubHeadset)

	if headSystem.HasHeadset() {
		t.Error("HeadTrackingSystem should report no headset with stub adapter")
	}

	// Test with VRControllerSystem
	ctrlSystem := NewVRControllerSystem(world)
	stubController := NewStubControllerAdapter()
	ctrlSystem.SetControllerAdapter(stubController)

	if ctrlSystem.HasController(ControllerLeft) {
		t.Error("VRControllerSystem should report no left controller with stub adapter")
	}

	if ctrlSystem.HasController(ControllerRight) {
		t.Error("VRControllerSystem should report no right controller with stub adapter")
	}
}

func BenchmarkStubHeadsetAdapter_GetHeadOrientation(b *testing.B) {
	adapter := NewStubHeadsetAdapter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.GetHeadOrientation()
	}
}

func BenchmarkStubControllerAdapter_GetTrigger(b *testing.B) {
	adapter := NewStubControllerAdapter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.GetTrigger(ControllerRight)
	}
}
