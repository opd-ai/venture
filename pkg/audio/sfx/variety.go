package sfx

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
)

// GenerateVariant creates a sound effect variant with pitch and volume randomization.
// Phase 14.4: Audio System Enhancement - Sound Variety
// This adds natural variation to repeated sounds (footsteps, attacks, impacts).
func (g *Generator) GenerateVariant(effectType string, seed int64, pitchVariance, volumeVariance float64) *audio.AudioSample {
	// Generate base sound
	sample := g.Generate(effectType, seed)

	// Create variant RNG
	rng := rand.New(rand.NewSource(seed + 1000)) // Offset seed for variance

	// Apply pitch variation (-pitchVariance to +pitchVariance semitones)
	pitchShift := 1.0
	if pitchVariance > 0 {
		semitones := (rng.Float64()*2 - 1) * pitchVariance // Random in [-variance, +variance]
		pitchShift = pitchRatioFromSemitones(semitones)
	}

	// Apply volume variation (0.0 to volumeVariance reduction)
	volumeMult := 1.0
	if volumeVariance > 0 {
		volumeMult = 1.0 - rng.Float64()*volumeVariance
	}

	// Apply pitch shift
	if pitchShift != 1.0 {
		g.applyPitchBend(sample.Data, pitchShift, pitchShift)
	}

	// Apply volume adjustment
	if volumeMult != 1.0 {
		for i := range sample.Data {
			sample.Data[i] *= volumeMult
		}
	}

	return sample
}

// GenerateMultiVariant creates multiple variants of a sound effect for variety.
// Phase 14.4: Audio System Enhancement - Sound Variety
// Returns an array of variants that can be randomly selected during playback.
func (g *Generator) GenerateMultiVariant(effectType string, baseSeed int64, numVariants int, pitchVariance, volumeVariance float64) []*audio.AudioSample {
	variants := make([]*audio.AudioSample, numVariants)

	for i := 0; i < numVariants; i++ {
		// Use different seed for each variant
		variantSeed := baseSeed + int64(i*1000)
		variants[i] = g.GenerateVariant(effectType, variantSeed, pitchVariance, volumeVariance)
	}

	return variants
}

// GenerateWithPitchShift generates a sound effect with a specific pitch shift.
// Phase 14.4: Audio System Enhancement - Sound Variety
// pitchShift in semitones: -12 = one octave down, +12 = one octave up.
func (g *Generator) GenerateWithPitchShift(effectType string, seed int64, semitones float64) *audio.AudioSample {
	sample := g.Generate(effectType, seed)

	pitchRatio := pitchRatioFromSemitones(semitones)
	g.applyPitchBend(sample.Data, pitchRatio, pitchRatio)

	return sample
}

// GenerateWithVolume generates a sound effect with a specific volume multiplier.
// Phase 14.4: Audio System Enhancement - Sound Variety
// volume: 0.0 = silent, 1.0 = original, >1.0 = amplified (may clip).
func (g *Generator) GenerateWithVolume(effectType string, seed int64, volume float64) *audio.AudioSample {
	sample := g.Generate(effectType, seed)

	for i := range sample.Data {
		sample.Data[i] *= volume

		// Clamp to prevent clipping
		if sample.Data[i] > 1.0 {
			sample.Data[i] = 1.0
		} else if sample.Data[i] < -1.0 {
			sample.Data[i] = -1.0
		}
	}

	return sample
}

// ApplyLowPassFilter applies a simple low-pass filter for muffled sounds.
// Phase 14.4: Audio System Enhancement - Sound Variety
// cutoffFactor: 0.0 = maximum filtering (very muffled), 1.0 = no filtering.
func (g *Generator) ApplyLowPassFilter(sample *audio.AudioSample, cutoffFactor float64) {
	if cutoffFactor >= 1.0 {
		return // No filtering needed
	}

	// Simple one-pole low-pass filter
	alpha := cutoffFactor
	previous := 0.0

	for i := range sample.Data {
		filtered := alpha*sample.Data[i] + (1-alpha)*previous
		previous = filtered
		sample.Data[i] = filtered
	}
}

// ApplyHighPassFilter applies a simple high-pass filter for tinny sounds.
// Phase 14.4: Audio System Enhancement - Sound Variety
// cutoffFactor: 0.0 = maximum filtering (very tinny), 1.0 = no filtering.
func (g *Generator) ApplyHighPassFilter(sample *audio.AudioSample, cutoffFactor float64) {
	if cutoffFactor >= 1.0 {
		return // No filtering needed
	}

	// Simple one-pole high-pass filter
	alpha := 1.0 - cutoffFactor
	previous := 0.0

	for i := range sample.Data {
		highPass := alpha * (sample.Data[i] - previous)
		previous = sample.Data[i]

		// Clamp to prevent clipping
		if highPass > 1.0 {
			highPass = 1.0
		} else if highPass < -1.0 {
			highPass = -1.0
		}

		sample.Data[i] = highPass
	}
}

