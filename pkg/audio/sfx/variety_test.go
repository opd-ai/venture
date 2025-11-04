package sfx

import (
	"math"
	"testing"
)

func TestGenerator_GenerateVariant(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	// Generate base sound
	base := gen.Generate("impact", 54321)

	// Generate variant with pitch and volume variance
	variant := gen.GenerateVariant("impact", 54321, 2.0, 0.2)

	// Verify sample exists
	if variant == nil {
		t.Fatal("GenerateVariant returned nil")
	}

	// Verify same sample rate
	if variant.SampleRate != base.SampleRate {
		t.Errorf("SampleRate = %v, want %v", variant.SampleRate, base.SampleRate)
	}

	// Verify same length (pitch shift shouldn't change length significantly)
	if len(variant.Data) < len(base.Data)/2 || len(variant.Data) > len(base.Data)*2 {
		t.Errorf("Variant length %v differs too much from base %v", len(variant.Data), len(base.Data))
	}

	// Verify volume is different (with some tolerance)
	baseRMS := calculateRMS(base.Data)
	variantRMS := calculateRMS(variant.Data)

	if math.Abs(baseRMS-variantRMS) < 0.01 {
		t.Error("Expected volume variation, got nearly identical RMS")
	}
}

func TestGenerator_GenerateMultiVariant(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	numVariants := 5
	// Use higher variance to ensure significant differences
	variants := gen.GenerateMultiVariant("impact", 54321, numVariants, 3.0, 0.3)

	if len(variants) != numVariants {
		t.Errorf("Got %d variants, want %d", len(variants), numVariants)
	}

	// Verify all variants exist and are non-empty
	for i := 0; i < numVariants; i++ {
		if variants[i] == nil {
			t.Errorf("Variant %d is nil", i)
			continue
		}
		if len(variants[i].Data) == 0 {
			t.Errorf("Variant %d has empty data", i)
		}
	}

	// Verify that most variants show significant differences
	// (Some can be similar due to random variance, but not all)
	similarCount := 0
	for i := 0; i < numVariants; i++ {
		for j := i + 1; j < numVariants; j++ {
			if variants[i] == nil || variants[j] == nil {
				continue
			}

			rmsI := calculateRMS(variants[i].Data)
			rmsJ := calculateRMS(variants[j].Data)

			percentDiff := math.Abs(rmsI-rmsJ) / ((rmsI + rmsJ) / 2.0)
			if percentDiff < 0.05 { // Less than 5% difference
				similarCount++
			}
		}
	}

	// More than half should be different
	totalPairs := (numVariants * (numVariants - 1)) / 2
	if similarCount > totalPairs/2 {
		t.Errorf("Too many similar variants: %d out of %d pairs", similarCount, totalPairs)
	}
}

func TestGenerator_GenerateWithPitchShift(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	tests := []struct {
		name      string
		semitones float64
	}{
		{"one octave up", 12.0},
		{"one octave down", -12.0},
		{"perfect fifth", 7.0},
		{"minor third down", -3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := gen.GenerateWithPitchShift("impact", 54321, tt.semitones)

			if sample == nil {
				t.Fatal("GenerateWithPitchShift returned nil")
			}

			if len(sample.Data) == 0 {
				t.Error("Sample data is empty")
			}
		})
	}
}

func TestGenerator_GenerateWithVolume(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	tests := []struct {
		name   string
		volume float64
	}{
		{"silent", 0.0},
		{"half volume", 0.5},
		{"full volume", 1.0},
		{"amplified", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := gen.GenerateWithVolume("impact", 54321, tt.volume)

			if sample == nil {
				t.Fatal("GenerateWithVolume returned nil")
			}

			// Check that values are within valid range
			for i, v := range sample.Data {
				if v > 1.0 || v < -1.0 {
					t.Errorf("Sample[%d] = %v, out of range [-1.0, 1.0]", i, v)
				}
			}

			// Verify RMS is approximately scaled by volume
			rms := calculateRMS(sample.Data)
			if tt.volume == 0.0 {
				if rms > 0.001 {
					t.Errorf("Silent sample has RMS %v, expected ~0", rms)
				}
			} else {
				// For non-zero volume, RMS should be non-zero
				if rms < 0.001 {
					t.Errorf("Non-silent sample has RMS %v, expected >0", rms)
				}
			}
		})
	}
}

func TestGenerator_ApplyLowPassFilter(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	sample := gen.Generate("laser", 54321) // High-frequency sound

	// Calculate RMS before filtering
	rmsBefore := calculateRMS(sample.Data)

	// Apply low-pass filter (muffled)
	gen.ApplyLowPassFilter(sample, 0.3)

	// Calculate RMS after filtering
	rmsAfter := calculateRMS(sample.Data)

	// Low-pass filter should reduce high frequencies, lowering RMS
	if rmsAfter >= rmsBefore {
		t.Errorf("Low-pass filter didn't reduce RMS: before=%v, after=%v", rmsBefore, rmsAfter)
	}

	// Verify values are still in range
	for i, v := range sample.Data {
		if v > 1.0 || v < -1.0 {
			t.Errorf("Sample[%d] = %v, out of range [-1.0, 1.0]", i, v)
		}
	}
}

func TestGenerator_ApplyHighPassFilter(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	sample := gen.Generate("explosion", 54321) // Low-frequency sound

	// Calculate RMS before filtering
	rmsBefore := calculateRMS(sample.Data)

	// Apply high-pass filter (tinny)
	gen.ApplyHighPassFilter(sample, 0.3)

	// Calculate RMS after filtering
	rmsAfter := calculateRMS(sample.Data)

	// High-pass filter should reduce low frequencies, lowering RMS
	if rmsAfter >= rmsBefore {
		t.Errorf("High-pass filter didn't reduce RMS: before=%v, after=%v", rmsBefore, rmsAfter)
	}

	// Verify values are still in range
	for i, v := range sample.Data {
		if v > 1.0 || v < -1.0 {
			t.Errorf("Sample[%d] = %v, out of range [-1.0, 1.0]", i, v)
		}
	}
}

func TestPitchRatioFromSemitones(t *testing.T) {
	tests := []struct {
		name      string
		semitones float64
		wantRatio float64
		tolerance float64
	}{
		{"unison", 0.0, 1.0, 0.01},
		{"one octave up", 12.0, 2.0, 0.05},
		{"one octave down", -12.0, 0.5, 0.05},
		{"perfect fifth", 7.0, 1.5, 0.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pitchRatioFromSemitones(tt.semitones)

			if math.Abs(got-tt.wantRatio) > tt.tolerance {
				t.Errorf("pitchRatioFromSemitones(%v) = %v, want %v (±%v)",
					tt.semitones, got, tt.wantRatio, tt.tolerance)
			}
		})
	}
}

func TestPow2(t *testing.T) {
	tests := []struct {
		name      string
		x         float64
		want      float64
		tolerance float64
	}{
		{"2^0", 0.0, 1.0, 0.001},
		{"2^1", 1.0, 2.0, 0.001},
		{"2^-1", -1.0, 0.5, 0.001},
		{"2^2", 2.0, 4.0, 0.1},
		{"2^0.5", 0.5, 1.414, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pow2(tt.x)

			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("pow2(%v) = %v, want %v (±%v)",
					tt.x, got, tt.want, tt.tolerance)
			}
		})
	}
}

func TestGenerator_ApplyFilters_NoOp(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	sample := gen.Generate("impact", 54321)
	original := make([]float64, len(sample.Data))
	copy(original, sample.Data)

	// Apply filters with cutoffFactor = 1.0 (no filtering)
	gen.ApplyLowPassFilter(sample, 1.0)

	for i := range sample.Data {
		if sample.Data[i] != original[i] {
			t.Errorf("Low-pass filter with cutoffFactor=1.0 modified sample at index %d", i)
			break
		}
	}

	// Reset sample
	copy(sample.Data, original)

	gen.ApplyHighPassFilter(sample, 1.0)

	for i := range sample.Data {
		if sample.Data[i] != original[i] {
			t.Errorf("High-pass filter with cutoffFactor=1.0 modified sample at index %d", i)
			break
		}
	}
}

// Helper function to calculate RMS (Root Mean Square) of audio data.
func calculateRMS(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sumSquares := 0.0
	for _, v := range data {
		sumSquares += v * v
	}

	return math.Sqrt(sumSquares / float64(len(data)))
}
