package engine

import (
	"math"
	"testing"
)

func TestNewMockHeadset(t *testing.T) {
	mock := NewMockHeadset()

	if mock == nil {
		t.Fatal("NewMockHeadset returned nil")
	}

	if !mock.IsConnected() {
		t.Error("Expected connected by default")
	}

	if mock.GetIPD() != 63.0 {
		t.Errorf("Expected IPD 63.0, got %v", mock.GetIPD())
	}
}

func TestMockHeadset_SetConnected(t *testing.T) {
	mock := NewMockHeadset()

	mock.SetConnected(false)
	if mock.IsConnected() {
		t.Error("Expected disconnected")
	}

	mock.SetConnected(true)
	if !mock.IsConnected() {
		t.Error("Expected connected")
	}
}

func TestMockHeadset_SetHeadOrientation(t *testing.T) {
	mock := NewMockHeadset()

	mock.SetHeadOrientation(0.5, 1.0, 0.2)
	p, y, r := mock.GetHeadOrientation()

	if p != 0.5 {
		t.Errorf("Expected pitch 0.5, got %v", p)
	}
	if y != 1.0 {
		t.Errorf("Expected yaw 1.0, got %v", y)
	}
	if r != 0.2 {
		t.Errorf("Expected roll 0.2, got %v", r)
	}
}

func TestMockHeadset_SetHeadPosition(t *testing.T) {
	mock := NewMockHeadset()

	mock.SetHeadPosition(1.0, 2.0, 3.0)
	x, y, z := mock.GetHeadPosition()

	if x != 1.0 {
		t.Errorf("Expected x 1.0, got %v", x)
	}
	if y != 2.0 {
		t.Errorf("Expected y 2.0, got %v", y)
	}
	if z != 3.0 {
		t.Errorf("Expected z 3.0, got %v", z)
	}
}

func TestMockHeadset_SetIPD(t *testing.T) {
	mock := NewMockHeadset()

	mock.SetIPD(65.0)
	if mock.GetIPD() != 65.0 {
		t.Errorf("Expected IPD 65.0, got %v", mock.GetIPD())
	}
}

func TestNewHeadTrackingSystem(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

	if sys == nil {
		t.Fatal("NewHeadTrackingSystem returned nil")
	}

	if !sys.IsEnabled() {
		t.Error("Expected enabled by default")
	}

	if !sys.IsUseMouseFallback() {
		t.Error("Expected mouse fallback enabled by default")
	}
}

func TestHeadTrackingSystem_SetEnabled(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

	sys.SetEnabled(true)
	if !sys.IsEnabled() {
		t.Error("Expected enabled")
	}

	sys.SetEnabled(false)
	if sys.IsEnabled() {
		t.Error("Expected disabled")
	}
}

func TestHeadTrackingSystem_SetHeadsetAdapter(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

	mock := NewMockHeadset()
	sys.SetHeadsetAdapter(mock)

	if sys.GetHeadsetAdapter() != mock {
		t.Error("Headset adapter not set correctly")
	}
}

func TestHeadTrackingSystem_HasHeadset(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

	// No headset
	if sys.HasHeadset() {
		t.Error("Expected no headset")
	}

	// Connected headset
	mock := NewMockHeadset()
	sys.SetHeadsetAdapter(mock)
	if !sys.HasHeadset() {
		t.Error("Expected headset connected")
	}

	// Disconnected headset
	mock.SetConnected(false)
	if sys.HasHeadset() {
		t.Error("Expected headset disconnected")
	}
}

func TestHeadTrackingSystem_Update_FromHeadset(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)
	sys.SetEnabled(true)

	mock := NewMockHeadset()
	mock.SetHeadOrientation(0.5, 1.0, 0.2)
	mock.SetHeadPosition(1, 2, 3)
	sys.SetHeadsetAdapter(mock)

	// Create entity with head tracking component
	entity := NewEntity(1)
	head := NewHeadTrackingComponent()
	head.SetEnabled(true)
	head.SmoothingFactor = 0
	head.PredictionMs = 0
	entity.AddComponent(head)

	sys.Update([]*Entity{entity}, 0.016)

	// Verify orientation was updated
	p, y, r := head.GetOrientation()
	if math.Abs(p-0.5) > 0.1 {
		t.Errorf("Expected pitch ~0.5, got %v", p)
	}
	if math.Abs(y-1.0) > 0.1 {
		t.Errorf("Expected yaw ~1.0, got %v", y)
	}
	if math.Abs(r-0.2) > 0.1 {
		t.Errorf("Expected roll ~0.2, got %v", r)
	}

	// Verify position was updated
	x, py, z := head.GetPosition()
	if math.Abs(x-1.0) > 0.01 {
		t.Errorf("Expected x ~1.0, got %v", x)
	}
	if math.Abs(py-2.0) > 0.01 {
		t.Errorf("Expected y ~2.0, got %v", py)
	}
	if math.Abs(z-3.0) > 0.01 {
		t.Errorf("Expected z ~3.0, got %v", z)
	}
}

func TestHeadTrackingSystem_Update_Disabled(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)
	sys.SetEnabled(false) // Explicitly disable for this test

	mock := NewMockHeadset()
	mock.SetHeadOrientation(0.5, 1.0, 0.2)
	sys.SetHeadsetAdapter(mock)

	entity := NewEntity(1)
	head := NewHeadTrackingComponent()
	head.SetEnabled(true)
	entity.AddComponent(head)

	sys.Update([]*Entity{entity}, 0.016)

	// Orientation should not be updated
	p, _, _ := head.GetOrientation()
	if p != 0 {
		t.Error("Orientation should not be updated when system disabled")
	}
}

func TestHeadTrackingSystem_Update_DisabledComponent(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)
	sys.SetEnabled(true)

	mock := NewMockHeadset()
	mock.SetHeadOrientation(0.5, 1.0, 0.2)
	sys.SetHeadsetAdapter(mock)

	entity := NewEntity(1)
	head := NewHeadTrackingComponent()
	head.SetEnabled(false) // Component disabled
	entity.AddComponent(head)

	sys.Update([]*Entity{entity}, 0.016)

	// Orientation should not be updated
	p, _, _ := head.GetOrientation()
	if p != 0 {
		t.Error("Orientation should not be updated when component disabled")
	}
}

func TestHeadTrackingSystem_Callbacks(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)
	sys.SetEnabled(true)

	mock := NewMockHeadset()
	mock.SetHeadOrientation(0.5, 1.0, 0.2)
	mock.SetHeadPosition(1, 2, 3)
	sys.SetHeadsetAdapter(mock)

	cameraCalled := false
	positionCalled := false
	var camPitch float64
	var posX float64

	sys.SetCameraUpdateCallback(func(pitch, yaw, roll float64) {
		cameraCalled = true
		camPitch = pitch
		_ = yaw  // Unused but captured
		_ = roll // Unused but captured
	})

	sys.SetPositionUpdateCallback(func(x, y, z float64) {
		positionCalled = true
		posX = x
		_ = y // Unused but captured
		_ = z // Unused but captured
	})

	entity := NewEntity(1)
	head := NewHeadTrackingComponent()
	head.SetEnabled(true)
	head.SmoothingFactor = 0
	head.PredictionMs = 0
	entity.AddComponent(head)

	sys.Update([]*Entity{entity}, 0.016)

	if !cameraCalled {
		t.Error("Camera callback not called")
	}
	if !positionCalled {
		t.Error("Position callback not called")
	}

	if math.Abs(camPitch-0.5) > 0.1 {
		t.Errorf("Camera pitch mismatch: %v", camPitch)
	}
	if math.Abs(posX-1.0) > 0.01 {
		t.Errorf("Position X mismatch: %v", posX)
	}
}

func TestHeadTrackingSystem_SetUseMouseFallback(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

	sys.SetUseMouseFallback(false)
	if sys.IsUseMouseFallback() {
		t.Error("Expected mouse fallback disabled")
	}

	sys.SetUseMouseFallback(true)
	if !sys.IsUseMouseFallback() {
		t.Error("Expected mouse fallback enabled")
	}
}

func TestHeadTrackingSystem_SetMouseSensitivity(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal", 0.005, 0.005},
		{"too low clamp", 0.00001, 0.0001},
		{"too high clamp", 0.5, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetMouseSensitivity(tt.input)
			if sys.GetMouseSensitivity() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, sys.GetMouseSensitivity())
			}
		})
	}
}

func TestHeadTrackingSystem_SetMousePosition(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)
	sys.SetMouseSensitivity(0.01) // 1% of movement

	// Move mouse
	sys.SetMousePosition(100, 50) // dx=100, dy=50

	pitch, yaw := sys.GetMouseOrientation()

	// yaw should be 100 * 0.01 = 1.0
	if math.Abs(yaw-1.0) > 0.01 {
		t.Errorf("Expected yaw ~1.0, got %v", yaw)
	}

	// pitch should be -50 * 0.01 = -0.5 (inverted)
	if math.Abs(pitch-(-0.5)) > 0.01 {
		t.Errorf("Expected pitch ~-0.5, got %v", pitch)
	}
}

func TestHeadTrackingSystem_ThreadSafety(t *testing.T) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)

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
			sys.SetMousePosition(1, 1)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			sys.GetMouseOrientation()
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func BenchmarkHeadTrackingSystem_Update(b *testing.B) {
	world := &World{}
	sys := NewHeadTrackingSystem(world)
	sys.SetEnabled(true)

	mock := NewMockHeadset()
	mock.SetHeadOrientation(0.5, 1.0, 0.2)
	sys.SetHeadsetAdapter(mock)

	entity := NewEntity(1)
	head := NewHeadTrackingComponent()
	head.SetEnabled(true)
	entity.AddComponent(head)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
