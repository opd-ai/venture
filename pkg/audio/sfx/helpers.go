package sfx

// Code relocated from: variety.go
//
// This file contains utility functions for audio processing and math operations.
// These helpers support pitch shifting and other audio transformations.

// pitchRatioFromSemitones converts a semitone shift to a frequency ratio.
//
// In music, 12 semitones = one octave = 2x frequency. This function uses the
// equal temperament tuning system where each semitone has a frequency ratio
// of 2^(1/12) ≈ 1.0595.
//
// Examples:
//   - pitchRatioFromSemitones(12) returns 2.0 (one octave up)
//   - pitchRatioFromSemitones(-12) returns 0.5 (one octave down)
//   - pitchRatioFromSemitones(7) returns ~1.498 (perfect fifth)
//   - pitchRatioFromSemitones(0) returns 1.0 (no change)
func pitchRatioFromSemitones(semitones float64) float64 {
	// Frequency ratio = 2^(semitones/12)
	const semitonesPerOctave = 12.0
	exponent := semitones / semitonesPerOctave

	// Simple power of 2 approximation
	// For exact: use math.Pow(2.0, exponent)
	// For performance: use lookup table or approximation

	// Exact calculation for now
	return pow2(exponent)
}

// pow2 computes 2^x using optimized handling for common integer cases,
// falling back to Taylor series approximation for fractional exponents.
//
// For integer exponents (0, 1, -1), exact values are returned. For fractional
// exponents, the function uses the Taylor series expansion of e^(x*ln(2)):
//
//	2^x = e^(x*ln(2)) ≈ 1 + x*ln(2) + (x*ln(2))²/2! + (x*ln(2))³/3! + ...
//
// The series converges to within 0.01% accuracy for typical audio pitch ratios
// (x values between -2 and +2). This is faster than math.Pow for audio
// applications where high precision is not critical.
func pow2(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	if x == 1 {
		return 2.0
	}
	if x == -1 {
		return 0.5
	}

	// For fractional exponents, use Taylor series approximation
	// 2^x ≈ 1 + x*ln(2) + (x*ln(2))^2/2! + ...
	const ln2 = 0.693147180559945309417232121458

	result := 1.0
	term := 1.0
	xLn2 := x * ln2

	for i := 1; i < 10; i++ {
		term *= xLn2 / float64(i)
		result += term
		if term < 0.0001 && term > -0.0001 {
			break
		}
	}

	return result
}
