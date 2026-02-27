package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestStubInput_TouchMethods tests the touch input methods.
func TestStubInput_TouchMethods(t *testing.T) {
	tests := []struct {
		name           string
		touchPositions map[ebiten.TouchID]struct{ X, Y int }
		wantTouchCount int
	}{
		{
			name:           "No touches",
			touchPositions: nil,
			wantTouchCount: 0,
		},
		{
			name:           "Single touch",
			touchPositions: map[ebiten.TouchID]struct{ X, Y int }{0: {100, 200}},
			wantTouchCount: 1,
		},
		{
			name: "Multiple touches",
			touchPositions: map[ebiten.TouchID]struct{ X, Y int }{
				0: {100, 100},
				1: {200, 200},
				2: {300, 300},
			},
			wantTouchCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &StubInput{
				TouchPositions: tt.touchPositions,
			}

			touchIDs := stub.GetTouchIDs()
			if len(touchIDs) != tt.wantTouchCount {
				t.Errorf("GetTouchIDs() returned %d touches, want %d", len(touchIDs), tt.wantTouchCount)
			}

			// Verify each touch position
			for id, expectedPos := range tt.touchPositions {
				x, y := stub.GetTouchPosition(id)
				if x != expectedPos.X || y != expectedPos.Y {
					t.Errorf("GetTouchPosition(%d) = (%d, %d), want (%d, %d)", id, x, y, expectedPos.X, expectedPos.Y)
				}
			}
		})
	}
}

// TestStubInput_TouchPosition_InvalidID tests behavior with invalid touch ID.
func TestStubInput_TouchPosition_InvalidID(t *testing.T) {
	stub := &StubInput{
		TouchPositions: map[ebiten.TouchID]struct{ X, Y int }{
			0: {100, 100},
		},
	}

	// Query non-existent ID
	x, y := stub.GetTouchPosition(999)
	if x != 0 || y != 0 {
		t.Errorf("GetTouchPosition(invalid) = (%d, %d), want (0, 0)", x, y)
	}
}

// TestStubInput_TouchPosition_NilMap tests behavior with nil TouchPositions map.
func TestStubInput_TouchPosition_NilMap(t *testing.T) {
	stub := &StubInput{
		TouchPositions: nil,
	}

	touchIDs := stub.GetTouchIDs()
	if touchIDs != nil {
		t.Errorf("GetTouchIDs() with nil map = %v, want nil", touchIDs)
	}

	x, y := stub.GetTouchPosition(0)
	if x != 0 || y != 0 {
		t.Errorf("GetTouchPosition(0) with nil map = (%d, %d), want (0, 0)", x, y)
	}
}

// TestStubInput_InterfaceCompliance verifies StubInput implements InputProvider.
func TestStubInput_InterfaceCompliance(t *testing.T) {
	var _ InputProvider = (*StubInput)(nil)
}
