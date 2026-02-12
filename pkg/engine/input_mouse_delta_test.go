package engine

import (
	"testing"
)

// TestInputProvider_GetMouseDelta_API verifies that GetMouseDelta is available in InputProvider interface.
func TestInputProvider_GetMouseDelta_API(t *testing.T) {
	// Test with EbitenInput
	ebitenInput := &EbitenInput{
		MouseDeltaX: 10,
		MouseDeltaY: -5,
	}

	dx, dy := ebitenInput.GetMouseDelta()
	if dx != 10 || dy != -5 {
		t.Errorf("EbitenInput.GetMouseDelta() = (%d, %d), want (10, -5)", dx, dy)
	}

	// Test with StubInput (for test compatibility)
	stubInput := &StubInput{
		MouseDeltaX: 20,
		MouseDeltaY: 15,
	}

	dx, dy = stubInput.GetMouseDelta()
	if dx != 20 || dy != 15 {
		t.Errorf("StubInput.GetMouseDelta() = (%d, %d), want (20, 15)", dx, dy)
	}
}

// TestInputProvider_GetMouseDelta_InterfaceCompliance verifies interface compliance.
func TestInputProvider_GetMouseDelta_InterfaceCompliance(t *testing.T) {
	var _ InputProvider = (*EbitenInput)(nil)
	var _ InputProvider = (*StubInput)(nil)

	// Test that GetMouseDelta is callable through interface
	var provider InputProvider = &EbitenInput{MouseDeltaX: 5, MouseDeltaY: 10}
	dx, dy := provider.GetMouseDelta()
	if dx != 5 || dy != 10 {
		t.Errorf("Interface GetMouseDelta() = (%d, %d), want (5, 10)", dx, dy)
	}
}

// TestInputSystem_MouseDelta_Propagation verifies that InputSystem populates delta in input component.
func TestInputSystem_MouseDelta_Propagation(t *testing.T) {
	system := NewInputSystem()
	entity := NewEntity(1)
	input := &EbitenInput{}
	entity.AddComponent(input)

	// Set delta values in system
	system.mouseDeltaX = 15
	system.mouseDeltaY = -10

	// Process mouse state should populate delta
	system.processMouseState(input)

	if input.MouseDeltaX != 15 || input.MouseDeltaY != -10 {
		t.Errorf("Input component delta = (%d, %d), want (15, -10)", input.MouseDeltaX, input.MouseDeltaY)
	}
}

// TestInputSystem_GetMouseDelta_Method verifies the InputSystem's GetMouseDelta method.
func TestInputSystem_GetMouseDelta_Method(t *testing.T) {
	system := NewInputSystem()

	// Set delta values
	system.mouseDeltaX = 25
	system.mouseDeltaY = -30

	dx, dy := system.GetMouseDelta()
	if dx != 25 || dy != -30 {
		t.Errorf("InputSystem.GetMouseDelta() = (%d, %d), want (25, -30)", dx, dy)
	}
}

// TestMouseDelta_CalculationLogic verifies delta is zero when mouse doesn't move.
func TestMouseDelta_CalculationLogic(t *testing.T) {
	// Test delta calculation logic
	lastX, lastY := 100, 200
	currentX, currentY := 100, 200
	deltaX := currentX - lastX
	deltaY := currentY - lastY

	if deltaX != 0 || deltaY != 0 {
		t.Errorf("Delta with no movement = (%d, %d), want (0, 0)", deltaX, deltaY)
	}
}

// TestMouseDelta_NegativeMovement verifies negative delta values work correctly.
func TestMouseDelta_NegativeMovement(t *testing.T) {
	// Simulate mouse moving left and up (negative delta)
	lastX, lastY := 200, 300
	currentX, currentY := 180, 270

	deltaX := currentX - lastX
	deltaY := currentY - lastY

	if deltaX != -20 || deltaY != -30 {
		t.Errorf("Negative delta = (%d, %d), want (-20, -30)", deltaX, deltaY)
	}
}

// TestEbitenInput_MouseDeltaFields verifies MouseDeltaX/Y fields are accessible.
func TestEbitenInput_MouseDeltaFields(t *testing.T) {
	input := &EbitenInput{}

	// Set delta values
	input.MouseDeltaX = 42
	input.MouseDeltaY = -17

	// Read back
	if input.MouseDeltaX != 42 {
		t.Errorf("MouseDeltaX = %d, want 42", input.MouseDeltaX)
	}
	if input.MouseDeltaY != -17 {
		t.Errorf("MouseDeltaY = %d, want -17", input.MouseDeltaY)
	}
}

// TestStubInput_MouseDeltaFields verifies StubInput supports mouse delta (for testing).
func TestStubInput_MouseDeltaFields(t *testing.T) {
	input := NewStubInput()

	input.MouseDeltaX = 100
	input.MouseDeltaY = -50

	dx, dy := input.GetMouseDelta()
	if dx != 100 || dy != -50 {
		t.Errorf("StubInput delta = (%d, %d), want (100, -50)", dx, dy)
	}
}

// TestMouseDelta_UsageScenario_CameraControl demonstrates camera control usage pattern.
func TestMouseDelta_UsageScenario_CameraControl(t *testing.T) {
	// Simulate camera control using mouse delta
	input := &EbitenInput{
		MouseDeltaX: 10,  // Mouse moved right
		MouseDeltaY: -15, // Mouse moved up
	}

	// Camera sensitivity
	sensitivity := 0.5

	// Calculate camera rotation change
	dx, dy := input.GetMouseDelta()
	cameraYaw := float64(dx) * sensitivity   // Horizontal rotation
	cameraPitch := float64(dy) * sensitivity // Vertical rotation

	expectedYaw := 10.0 * 0.5    // = 5.0
	expectedPitch := -15.0 * 0.5 // = -7.5

	if cameraYaw != expectedYaw {
		t.Errorf("Camera yaw = %f, want %f", cameraYaw, expectedYaw)
	}
	if cameraPitch != expectedPitch {
		t.Errorf("Camera pitch = %f, want %f", cameraPitch, expectedPitch)
	}
}

// TestMouseDelta_UsageScenario_AimingAssist demonstrates aiming assist usage pattern.
func TestMouseDelta_UsageScenario_AimingAssist(t *testing.T) {
	// Simulate aiming assist that reduces large mouse movements
	input := &EbitenInput{
		MouseDeltaX: 100, // Large horizontal movement
		MouseDeltaY: 50,  // Medium vertical movement
	}

	dx, dy := input.GetMouseDelta()

	// Apply smoothing (reduce large movements)
	maxDelta := 50.0
	smoothedDX := float64(dx)
	smoothedDY := float64(dy)

	if smoothedDX > maxDelta {
		smoothedDX = maxDelta
	}
	if smoothedDY > maxDelta {
		smoothedDY = maxDelta
	}

	if smoothedDX != 50.0 {
		t.Errorf("Smoothed X = %f, want 50.0 (clamped)", smoothedDX)
	}
	if smoothedDY != 50.0 {
		t.Errorf("Smoothed Y = %f, want 50.0", smoothedDY)
	}
}

// TestMouseDelta_UsageScenario_DragAndDrop demonstrates drag-and-drop usage pattern.
func TestMouseDelta_UsageScenario_DragAndDrop(t *testing.T) {
	// Simulate dragging an item
	input := &EbitenInput{
		MouseX:       150,
		MouseY:       200,
		MouseDeltaX:  5,
		MouseDeltaY:  -3,
		MousePressed: true,
	}

	// Item being dragged
	itemX, itemY := 145.0, 203.0

	if input.IsMousePressed() {
		dx, dy := input.GetMouseDelta()
		itemX += float64(dx)
		itemY += float64(dy)
	}

	expectedX := 145.0 + 5.0  // = 150.0
	expectedY := 203.0 + -3.0 // = 200.0

	if itemX != expectedX {
		t.Errorf("Item X after drag = %f, want %f", itemX, expectedX)
	}
	if itemY != expectedY {
		t.Errorf("Item Y after drag = %f, want %f", itemY, expectedY)
	}
}

// TestMouseDelta_EdgeCases tests edge cases for mouse delta.
func TestMouseDelta_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		deltaX int
		deltaY int
		wantX  int
		wantY  int
	}{
		{"Zero delta", 0, 0, 0, 0},
		{"Positive X only", 100, 0, 100, 0},
		{"Positive Y only", 0, 50, 0, 50},
		{"Negative X only", -100, 0, -100, 0},
		{"Negative Y only", 0, -50, 0, -50},
		{"Large positive", 1000, 2000, 1000, 2000},
		{"Large negative", -1000, -2000, -1000, -2000},
		{"Mixed signs", 100, -100, 100, -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &EbitenInput{
				MouseDeltaX: tt.deltaX,
				MouseDeltaY: tt.deltaY,
			}

			dx, dy := input.GetMouseDelta()
			if dx != tt.wantX || dy != tt.wantY {
				t.Errorf("GetMouseDelta() = (%d, %d), want (%d, %d)", dx, dy, tt.wantX, tt.wantY)
			}
		})
	}
}
