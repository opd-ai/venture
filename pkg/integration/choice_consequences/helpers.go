package choice_consequences

// helpers.go contains shared utility functions used throughout the choice_consequences package.
// These are internal helpers for mathematical operations and value manipulation.
//
// Code relocated from: manager.go and types.go

// clamp restricts a value to a min/max range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
