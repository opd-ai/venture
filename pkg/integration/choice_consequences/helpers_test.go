package choice_consequences

import "testing"

// TestAbs directly tests the abs helper function.
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		want float64
	}{
		{
			name: "positive value",
			x:    5.5,
			want: 5.5,
		},
		{
			name: "negative value",
			x:    -5.5,
			want: 5.5,
		},
		{
			name: "zero",
			x:    0.0,
			want: 0.0,
		},
		{
			name: "negative zero",
			x:    -0.0,
			want: 0.0,
		},
		{
			name: "small negative",
			x:    -0.001,
			want: 0.001,
		},
		{
			name: "large positive",
			x:    1000000.0,
			want: 1000000.0,
		},
		{
			name: "large negative",
			x:    -1000000.0,
			want: 1000000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := abs(tt.x)
			if got != tt.want {
				t.Errorf("abs(%f) = %f, want %f", tt.x, got, tt.want)
			}
		})
	}
}

// TestClamp directly tests the clamp helper function.
func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		min   float64
		max   float64
		want  float64
	}{
		{
			name:  "within range",
			value: 0.5,
			min:   0.0,
			max:   1.0,
			want:  0.5,
		},
		{
			name:  "below min",
			value: -0.5,
			min:   0.0,
			max:   1.0,
			want:  0.0,
		},
		{
			name:  "above max",
			value: 1.5,
			min:   0.0,
			max:   1.0,
			want:  1.0,
		},
		{
			name:  "at min boundary",
			value: 0.0,
			min:   0.0,
			max:   1.0,
			want:  0.0,
		},
		{
			name:  "at max boundary",
			value: 1.0,
			min:   0.0,
			max:   1.0,
			want:  1.0,
		},
		{
			name:  "negative range",
			value: 0.0,
			min:   -1.0,
			max:   -0.5,
			want:  -0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
