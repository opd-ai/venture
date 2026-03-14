package choice_consequences

import "math"

// helpers.go contains shared utility functions used throughout the choice_consequences package.
// These are internal helpers for mathematical operations and value manipulation.
//
// Code relocated from: manager.go and types.go

// clamp restricts a value to a min/max range.
func clamp(value, minimum, maximum float64) float64 {
	return max(minimum, min(maximum, value))
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	return math.Abs(x)
}
