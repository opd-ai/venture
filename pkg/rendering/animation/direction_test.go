package animation

import (
	"math"
	"testing"
)

func TestDirection8String(t *testing.T) {
	tests := []struct {
		dir  Direction8
		want string
	}{
		{Dir8North, "north"},
		{Dir8NorthEast, "northeast"},
		{Dir8East, "east"},
		{Dir8SouthEast, "southeast"},
		{Dir8South, "south"},
		{Dir8SouthWest, "southwest"},
		{Dir8West, "west"},
		{Dir8NorthWest, "northwest"},
		{Direction8(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.dir.String(); got != tt.want {
			t.Errorf("Direction8.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestDirection8Angle(t *testing.T) {
	tests := []struct {
		dir       Direction8
		wantAngle float64
	}{
		{Dir8East, 0},
		{Dir8NorthEast, math.Pi / 4},
		{Dir8North, math.Pi / 2},
		{Dir8NorthWest, 3 * math.Pi / 4},
		{Dir8West, math.Pi},
		{Dir8SouthWest, 5 * math.Pi / 4},
		{Dir8South, 3 * math.Pi / 2},
		{Dir8SouthEast, 7 * math.Pi / 4},
	}

	for _, tt := range tests {
		got := tt.dir.Angle()
		if math.Abs(got-tt.wantAngle) > 0.001 {
			t.Errorf("Direction8.Angle() = %v, want %v", got, tt.wantAngle)
		}
	}
}

func TestDirection8IsDiagonal(t *testing.T) {
	tests := []struct {
		dir  Direction8
		want bool
	}{
		{Dir8North, false},
		{Dir8NorthEast, true},
		{Dir8East, false},
		{Dir8SouthEast, true},
		{Dir8South, false},
		{Dir8SouthWest, true},
		{Dir8West, false},
		{Dir8NorthWest, true},
	}

	for _, tt := range tests {
		if got := tt.dir.IsDiagonal(); got != tt.want {
			t.Errorf("Direction8.IsDiagonal() for %v = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestFromVelocity(t *testing.T) {
	tests := []struct {
		name string
		vx   float64
		vy   float64
		want Direction8
	}{
		{"zero velocity", 0, 0, Dir8South},
		{"north", 0, -1, Dir8North},
		{"northeast", 1, -1, Dir8NorthEast},
		{"east", 1, 0, Dir8East},
		{"southeast", 1, 1, Dir8SouthEast},
		{"south", 0, 1, Dir8South},
		{"southwest", -1, 1, Dir8SouthWest},
		{"west", -1, 0, Dir8West},
		{"northwest", -1, -1, Dir8NorthWest},
		{"small velocity", 0.005, 0.005, Dir8South},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromVelocity(tt.vx, tt.vy)
			if got != tt.want {
				t.Errorf("FromVelocity(%v, %v) = %v, want %v", tt.vx, tt.vy, got, tt.want)
			}
		})
	}
}

func TestDirection8To4Direction(t *testing.T) {
	tests := []struct {
		dir  Direction8
		want string
	}{
		{Dir8North, "up"},
		{Dir8NorthEast, "up"},
		{Dir8East, "right"},
		{Dir8SouthEast, "down"},
		{Dir8South, "down"},
		{Dir8SouthWest, "down"},
		{Dir8West, "left"},
		{Dir8NorthWest, "up"},
	}

	for _, tt := range tests {
		if got := tt.dir.To4Direction(); got != tt.want {
			t.Errorf("Direction8.To4Direction() for %v = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestDirection8Opposite(t *testing.T) {
	tests := []struct {
		dir  Direction8
		want Direction8
	}{
		{Dir8North, Dir8South},
		{Dir8NorthEast, Dir8SouthWest},
		{Dir8East, Dir8West},
		{Dir8SouthEast, Dir8NorthWest},
	}

	for _, tt := range tests {
		if got := tt.dir.Opposite(); got != tt.want {
			t.Errorf("Direction8.Opposite() for %v = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestDirection8Rotate(t *testing.T) {
	tests := []struct {
		name      string
		dir       Direction8
		want      Direction8
		clockwise bool
	}{
		{"north CW", Dir8North, Dir8NorthEast, true},
		{"northeast CW", Dir8NorthEast, Dir8East, true},
		{"north CCW", Dir8North, Dir8NorthWest, false},
		{"northeast CCW", Dir8NorthEast, Dir8North, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.clockwise {
				if got := tt.dir.RotateClockwise(); got != tt.want {
					t.Errorf("RotateClockwise() = %v, want %v", got, tt.want)
				}
			} else {
				if got := tt.dir.RotateCounterClockwise(); got != tt.want {
					t.Errorf("RotateCounterClockwise() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFromVelocityDeterminism(t *testing.T) {
	// Test that same velocity always produces same direction
	vx, vy := 1.5, -0.8
	dir1 := FromVelocity(vx, vy)
	dir2 := FromVelocity(vx, vy)
	if dir1 != dir2 {
		t.Errorf("FromVelocity not deterministic: got %v and %v", dir1, dir2)
	}
}

func BenchmarkFromVelocity(b *testing.B) {
	vx, vy := 1.0, -1.0
	for i := 0; i < b.N; i++ {
		FromVelocity(vx, vy)
	}
}

func BenchmarkDirection8Angle(b *testing.B) {
	dir := Dir8NorthEast
	for i := 0; i < b.N; i++ {
		dir.Angle()
	}
}
