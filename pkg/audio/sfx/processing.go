package sfx

// Code relocated from: generator.go
//
// This file contains audio processing methods that modify sound samples
// through pitch manipulation, vibrato effects, and audio mixing.

import (
	"math"
)

// applyPitchBend applies a pitch bend effect to the sample.
func (g *Generator) applyPitchBend(data []float64, startRatio, endRatio float64) {
	// Create a copy to read from while we modify
	original := make([]float64, len(data))
	copy(original, data)

	for i := range data {
		progress := float64(i) / float64(len(data))
		ratio := startRatio + (endRatio-startRatio)*progress

		// Simple pitch shift by stretching/compressing
		sourceIdx := int(float64(i) / ratio)
		if sourceIdx >= 0 && sourceIdx < len(original) {
			data[i] = original[sourceIdx]
		} else {
			data[i] = 0
		}
	}
}

// applyVibrato applies vibrato effect to the sample.
func (g *Generator) applyVibrato(data []float64, rate, depth float64) {
	// Create a copy to read from while we modify
	original := make([]float64, len(data))
	copy(original, data)

	for i := range data {
		t := float64(i) / float64(g.sampleRate)
		offset := depth * math.Sin(2*math.Pi*rate*t)
		sourceIdx := i + int(offset*float64(g.sampleRate))

		if sourceIdx >= 0 && sourceIdx < len(original) {
			data[i] = original[sourceIdx]
		} else {
			data[i] = 0
		}
	}
}

// mix mixes two audio buffers together.
func (g *Generator) mix(dst, src []float64, srcVolume float64) {
	length := len(dst)
	if len(src) < length {
		length = len(src)
	}

	for i := 0; i < length; i++ {
		dst[i] += src[i] * srcVolume

		// Clamp to [-1, 1]
		if dst[i] > 1.0 {
			dst[i] = 1.0
		} else if dst[i] < -1.0 {
			dst[i] = -1.0
		}
	}
}
