// Package engine provides tests for rotation functionality.
package engine

import (
	"math"
	"testing"
)

// TestRotationComponent_Type verifies component type identifier
func TestRotationComponent_Type(t *testing.T) {
	comp := NewRotationComponent(0, 0)
	if got := comp.Type(); got != "rotation" {
		t.Errorf("Type() = %q, want %q", got, "rotation")
	}
}

// TestNewRotationComponent tests component creation with defaults
func TestNewRotationComponent(t *testing.T) {
	tests := []struct {
		name          string
		initialAngle  float64
		rotationSpeed float64
		wantAngle     float64
		wantSpeed     float64
	}{
		{
			name:          "default speed",
			initialAngle:  0,
			rotationSpeed: 0,
			wantAngle:     0,
			wantSpeed:     3.0,
		},
		{
			name:          "custom speed",
			initialAngle:  math.Pi / 2,
			rotationSpeed: 5.0,
			wantAngle:     math.Pi / 2,
			wantSpeed:     5.0,
		},
		{
			name:          "angle normalization",
			initialAngle:  3 * math.Pi, // Should normalize to π
			rotationSpeed: 2.0,
			wantAngle:     math.Pi,
			wantSpeed:     2.0,
		},
		{
			name:          "negative angle normalization",
			initialAngle:  -math.Pi / 2,
			rotationSpeed: 4.0,
			wantAngle:     3 * math.Pi / 2,
			wantSpeed:     4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewRotationComponent(tt.initialAngle, tt.rotationSpeed)

			if !floatEqual(comp.Angle, tt.wantAngle, 0.0001) {
				t.Errorf("Angle = %v, want %v", comp.Angle, tt.wantAngle)
			}
			if !floatEqual(comp.TargetAngle, tt.wantAngle, 0.0001) {
				t.Errorf("TargetAngle = %v, want %v", comp.TargetAngle, tt.wantAngle)
			}
			if comp.RotationSpeed != tt.wantSpeed {
				t.Errorf("RotationSpeed = %v, want %v", comp.RotationSpeed, tt.wantSpeed)
			}
			if !comp.SmoothRotation {
				t.Error("SmoothRotation should default to true")
			}
			if comp.AngularVelocity != 0 {
				t.Errorf("AngularVelocity = %v, want 0", comp.AngularVelocity)
			}
		})
	}
}

// TestRotationComponent_SetTargetAngle tests target angle setting
func TestRotationComponent_SetTargetAngle(t *testing.T) {
	comp := NewRotationComponent(0, 3.0)

	tests := []struct {
		name       string
		targetAngle float64
		wantNormalized float64
	}{
		{"zero", 0, 0},
		{"quarter turn", math.Pi / 2, math.Pi / 2},
		{"half turn", math.Pi, math.Pi},
		{"three quarter", 3 * math.Pi / 2, 3 * math.Pi / 2},
		{"full turn", 2 * math.Pi, 0}, // Normalized to 0
		{"negative", -math.Pi / 4, 7 * math.Pi / 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.SetTargetAngle(tt.targetAngle)
			if !floatEqual(comp.TargetAngle, tt.wantNormalized, 0.0001) {
				t.Errorf("TargetAngle = %v, want %v", comp.TargetAngle, tt.wantNormalized)
			}
		})
	}
}

// TestRotationComponent_SetAngleImmediate tests instant rotation
func TestRotationComponent_SetAngleImmediate(t *testing.T) {
	comp := NewRotationComponent(0, 3.0)
	comp.AngularVelocity = 1.5

	comp.SetAngleImmediate(math.Pi)

	if !floatEqual(comp.Angle, math.Pi, 0.0001) {
		t.Errorf("Angle = %v, want %v", comp.Angle, math.Pi)
	}
	if !floatEqual(comp.TargetAngle, math.Pi, 0.0001) {
		t.Errorf("TargetAngle = %v, want %v", comp.TargetAngle, math.Pi)
	}
	if comp.AngularVelocity != 0 {
		t.Errorf("AngularVelocity = %v, want 0", comp.AngularVelocity)
	}
}

// TestRotationComponent_Update tests smooth rotation interpolation
func TestRotationComponent_Update(t *testing.T) {
	tests := []struct {
		name          string
		initialAngle  float64
		targetAngle   float64
		rotationSpeed float64
		deltaTime     float64
		wantAngle     float64
		wantComplete  bool
	}{
		{
			name:          "no rotation needed",
			initialAngle:  math.Pi,
			targetAngle:   math.Pi,
			rotationSpeed: 3.0,
			deltaTime:     0.016,
			wantAngle:     math.Pi,
			wantComplete:  true,
		},
		{
			name:          "partial rotation clockwise",
			initialAngle:  0,
			targetAngle:   math.Pi / 2,
			rotationSpeed: 3.0,
			deltaTime:     0.1, // 0.3 radians max rotation
			wantAngle:     0.3,
			wantComplete:  false,
		},
		{
			name:          "complete rotation in one step",
			initialAngle:  0,
			targetAngle:   0.2,
			rotationSpeed: 3.0,
			deltaTime:     0.1, // 0.3 radians max rotation (enough)
			wantAngle:     0.2,
			wantComplete:  true,
		},
		{
			name:          "rotation wraps around zero",
			initialAngle:  0.1,
			targetAngle:   2 * math.Pi - 0.1,
			rotationSpeed: 3.0,
			deltaTime:     0.1, // Should rotate counter-clockwise
			wantAngle:     2*math.Pi - 0.2, // Moved 0.3 radians counter-clockwise
			wantComplete:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewRotationComponent(tt.initialAngle, tt.rotationSpeed)
			comp.SetTargetAngle(tt.targetAngle)

			complete := comp.Update(tt.deltaTime)

			if complete != tt.wantComplete {
				t.Errorf("Update() complete = %v, want %v", complete, tt.wantComplete)
			}
			if !floatEqual(comp.Angle, tt.wantAngle, 0.1) {
				t.Errorf("Angle = %v, want ~%v", comp.Angle, tt.wantAngle)
			}
		})
	}
}

// TestRotationComponent_UpdateInstantMode tests non-smooth rotation
func TestRotationComponent_UpdateInstantMode(t *testing.T) {
	comp := NewRotationComponent(0, 3.0)
	comp.SmoothRotation = false
	comp.SetTargetAngle(math.Pi)

	complete := comp.Update(0.016)

	if !complete {
		t.Error("Update() should complete immediately in instant mode")
	}
	if !floatEqual(comp.Angle, math.Pi, 0.0001) {
		t.Errorf("Angle = %v, want %v", comp.Angle, math.Pi)
	}
	if comp.AngularVelocity != 0 {
		t.Errorf("AngularVelocity = %v, want 0", comp.AngularVelocity)
	}
}

// TestRotationComponent_GetDirectionVector tests direction vector calculation
func TestRotationComponent_GetDirectionVector(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
		wantX float64
		wantY float64
	}{
		{"right", 0, 1.0, 0.0},
		{"down", math.Pi / 2, 0.0, 1.0},
		{"left", math.Pi, -1.0, 0.0},
		{"up", 3 * math.Pi / 2, 0.0, -1.0},
		{"down-right", math.Pi / 4, 0.707, 0.707},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewRotationComponent(tt.angle, 3.0)
			x, y := comp.GetDirectionVector()

			if !floatEqual(x, tt.wantX, 0.01) {
				t.Errorf("x = %v, want %v", x, tt.wantX)
			}
			if !floatEqual(y, tt.wantY, 0.01) {
				t.Errorf("y = %v, want %v", y, tt.wantY)
			}
		})
	}
}

// TestRotationComponent_GetCardinalDirection tests cardinal direction mapping
func TestRotationComponent_GetCardinalDirection(t *testing.T) {
	tests := []struct {
		name          string
		angle         float64
		wantDirection int
	}{
		{"right", 0, 0},
		{"down-right", math.Pi / 4, 1},
		{"down", math.Pi / 2, 2},
		{"down-left", 3 * math.Pi / 4, 3},
		{"left", math.Pi, 4},
		{"up-left", 5 * math.Pi / 4, 5},
		{"up", 3 * math.Pi / 2, 6},
		{"up-right", 7 * math.Pi / 4, 7},
		{"near right", 0.1, 0}, // Within right sector
		{"near left", math.Pi - 0.1, 4}, // Within left sector
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewRotationComponent(tt.angle, 3.0)
			dir := comp.GetCardinalDirection()

			if dir != tt.wantDirection {
				t.Errorf("GetCardinalDirection() = %d, want %d", dir, tt.wantDirection)
			}
		})
	}
}

// TestNormalizeAngle tests angle normalization
func TestNormalizeAngle(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
		want  float64
	}{
		{"zero", 0, 0},
		{"pi", math.Pi, math.Pi},
		{"2pi", 2 * math.Pi, 0},
		{"3pi", 3 * math.Pi, math.Pi},
		{"negative pi/2", -math.Pi / 2, 3 * math.Pi / 2},
		{"negative pi", -math.Pi, math.Pi},
		{"large positive", 10 * math.Pi, 0},
		{"large negative", -10 * math.Pi, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAngle(tt.angle)
			if !floatEqual(got, tt.want, 0.0001) {
				t.Errorf("normalizeAngle(%v) = %v, want %v", tt.angle, got, tt.want)
			}
		})
	}
}

// TestShortestAngularDistance tests shortest rotation calculation
func TestShortestAngularDistance(t *testing.T) {
	tests := []struct {
		name   string
		angle1 float64
		angle2 float64
		want   float64
	}{
		{"no rotation", 0, 0, 0},
		{"quarter clockwise", 0, math.Pi / 2, math.Pi / 2},
		{"quarter counter-clockwise", math.Pi / 2, 0, -math.Pi / 2},
		{"half turn", 0, math.Pi, math.Pi},
		{"wrap around clockwise", 7 * math.Pi / 4, math.Pi / 4, math.Pi / 2},
		{"wrap around counter-clockwise", math.Pi / 4, 7 * math.Pi / 4, -math.Pi / 2},
		{"nearly full turn clockwise", 0, 2*math.Pi - 0.1, -0.1},
		{"nearly full turn counter-clockwise", 2*math.Pi - 0.1, 0, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortestAngularDistance(tt.angle1, tt.angle2)
			if !floatEqual(got, tt.want, 0.01) {
				t.Errorf("shortestAngularDistance(%v, %v) = %v, want %v",
					tt.angle1, tt.angle2, got, tt.want)
			}
		})
	}
}

// floatEqual compares two floats with epsilon tolerance
func floatEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}
