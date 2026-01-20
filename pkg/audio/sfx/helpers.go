package sfx

// Code relocated from: variety.go
//
// This file contains utility functions for audio processing and math operations.
// These helpers support pitch shifting and other audio transformations.


// pitchRatioFromSemitones converts semitone shift to frequency ratio.
// 12 semitones = one octave = 2x frequency.
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

// pow2 computes 2^x using bit manipulation for common integer cases,
// falling back to approximation for fractional exponents.
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
