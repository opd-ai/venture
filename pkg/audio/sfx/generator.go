// Package sfx provides procedural sound effect generation.
// This file implements sound effect generators for game events like
// attacks, impacts, pickups, and environmental sounds.
package sfx

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/synthesis"
	"github.com/sirupsen/logrus"
)

// EffectType represents different types of sound effects.
type EffectType string

// Sound effect type constants.
const (
	EffectImpact    EffectType = "impact"
	EffectExplosion EffectType = "explosion"
	EffectMagic     EffectType = "magic"
	EffectLaser     EffectType = "laser"
	EffectPickup    EffectType = "pickup"
	EffectHit       EffectType = "hit"
	EffectJump      EffectType = "jump"
	EffectDeath     EffectType = "death"
	EffectPowerup   EffectType = "powerup"
)

// Generator creates procedural sound effects.
type Generator struct {
	sampleRate int
	osc        *synthesis.Oscillator
	rng        *rand.Rand
	logger     *logrus.Entry
}

// NewGenerator creates a new SFX generator.
func NewGenerator(sampleRate int, seed int64) *Generator {
	return NewGeneratorWithLogger(sampleRate, seed, nil)
}

// NewGeneratorWithLogger creates a new SFX generator with a logger.
func NewGeneratorWithLogger(sampleRate int, seed int64, logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator":  "sfx",
			"sampleRate": sampleRate,
		})
	}
	return &Generator{
		sampleRate: sampleRate,
		osc:        synthesis.NewOscillator(sampleRate, seed),
		rng:        rand.New(rand.NewSource(seed)),
		logger:     logEntry,
	}
}

// Generate creates a sound effect of the specified type.
// GAP-011 REPAIR: Added genre parameter for genre-specific sound variations.
func (g *Generator) Generate(effectType string, seed int64) *audio.AudioSample {
	return g.GenerateWithGenre(effectType, seed, "")
}

// GenerateWithGenre creates a sound effect with genre-specific characteristics.
// GAP-011 REPAIR: Genre affects frequency ranges, waveforms, and envelopes.
func (g *Generator) GenerateWithGenre(effectType string, seed int64, genre string) *audio.AudioSample {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"effectType": effectType,
			"seed":       seed,
			"genre":      genre,
		}).Debug("generating sound effect")
	}

	// Use provided seed for variation
	localRng := rand.New(rand.NewSource(seed))

	// Generate base sound effect
	var sample *audio.AudioSample
	switch EffectType(effectType) {
	case EffectImpact:
		sample = g.generateImpact(localRng)
	case EffectExplosion:
		sample = g.generateExplosion(localRng)
	case EffectMagic:
		sample = g.generateMagic(localRng)
	case EffectLaser:
		sample = g.generateLaser(localRng)
	case EffectPickup:
		sample = g.generatePickup(localRng)
	case EffectHit:
		sample = g.generateHit(localRng)
	case EffectJump:
		sample = g.generateJump(localRng)
	case EffectDeath:
		sample = g.generateDeath(localRng)
	case EffectPowerup:
		sample = g.generatePowerup(localRng)
	default:
		sample = g.generateImpact(localRng)
	}

	// Apply genre-specific modifications
	if genre != "" && genre != "fantasy" {
		g.applyGenreModifications(sample, genre)
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"effectType":  effectType,
			"sampleCount": len(sample.Data),
		}).Info("sound effect generated")
	}

	return sample
}

// applyGenreModifications modifies a sound sample based on genre.
// GAP-011 REPAIR: Different genres have different sonic characteristics.
func (g *Generator) applyGenreModifications(sample *audio.AudioSample, genre string) {
	switch genre {
	case "scifi":
		// Synthetic, clean, higher pitch
		g.applyPitchBend(sample.Data, 1.3, 1.3)
		// Reduce amplitude for cleaner sound
		for i := range sample.Data {
			sample.Data[i] *= 0.9
		}
	case "horror":
		// Dissonant, unsettling, lower pitch
		g.applyPitchBend(sample.Data, 0.7, 0.7)
		// Add vibrato for unsettling effect
		g.applyVibrato(sample.Data, 3.0, 0.2)
	case "cyberpunk":
		// Sharp, electronic, glitchy
		g.applyPitchBend(sample.Data, 1.4, 1.4)
		// Add hard clipping for digital effect
		for i := range sample.Data {
			if sample.Data[i] > 0.7 {
				sample.Data[i] = 0.7
			} else if sample.Data[i] < -0.7 {
				sample.Data[i] = -0.7
			}
		}
	case "postapoc":
		// Harsh, gritty, distorted
		g.applyPitchBend(sample.Data, 0.9, 0.9)
		// Add soft clipping for gritty effect
		for i := range sample.Data {
			if sample.Data[i] > 0.5 {
				sample.Data[i] = 0.5 + (sample.Data[i]-0.5)*0.3
			} else if sample.Data[i] < -0.5 {
				sample.Data[i] = -0.5 + (sample.Data[i]+0.5)*0.3
			}
		}
	}
}

// generateImpact creates a short, punchy impact sound.
func (g *Generator) generateImpact(rng *rand.Rand) *audio.AudioSample {
	duration := 0.1 + rng.Float64()*0.1
	frequency := 80.0 + rng.Float64()*40.0

	sample := g.osc.Generate(audio.WaveformNoise, frequency, duration)

	// Apply sharp attack, fast decay
	env := synthesis.Envelope{
		Attack:  0.001,
		Decay:   0.02,
		Sustain: 0.1,
		Release: 0.05,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Add pitch bend down
	g.applyPitchBend(sample.Data, 1.0, 0.5)

	return sample
}

// generateExplosion creates a big boom sound.
func (g *Generator) generateExplosion(rng *rand.Rand) *audio.AudioSample {
	duration := 0.5 + rng.Float64()*0.3

	sample := g.osc.Generate(audio.WaveformNoise, 0, duration)

	env := synthesis.Envelope{
		Attack:  0.001,
		Decay:   0.1,
		Sustain: 0.3,
		Release: 0.3,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Add low-frequency rumble
	rumble := g.osc.Generate(audio.WaveformSine, 40.0, duration)
	g.mix(sample.Data, rumble.Data, 0.5)

	return sample
}

// generateMagic creates a magical sparkle sound.
func (g *Generator) generateMagic(rng *rand.Rand) *audio.AudioSample {
	duration := 0.3 + rng.Float64()*0.2
	frequency := 800.0 + rng.Float64()*400.0

	sample := g.osc.Generate(audio.WaveformSine, frequency, duration)

	env := synthesis.Envelope{
		Attack:  0.01,
		Decay:   0.1,
		Sustain: 0.5,
		Release: 0.2,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Add shimmer with higher harmonics
	harmonic := g.osc.Generate(audio.WaveformSine, frequency*2.0, duration)
	g.mix(sample.Data, harmonic.Data, 0.3)

	// Apply vibrato
	g.applyVibrato(sample.Data, 5.0, 0.1)

	return sample
}

// generateLaser creates a sci-fi laser sound.
func (g *Generator) generateLaser(rng *rand.Rand) *audio.AudioSample {
	duration := 0.2 + rng.Float64()*0.1
	startFreq := 1200.0 + rng.Float64()*400.0

	sample := g.osc.Generate(audio.WaveformSquare, startFreq, duration)

	env := synthesis.Envelope{
		Attack:  0.001,
		Decay:   0.05,
		Sustain: 0.3,
		Release: 0.1,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Pitch sweep down
	g.applyPitchBend(sample.Data, 1.0, 0.3)

	return sample
}

// generatePickup creates an item pickup sound.
func (g *Generator) generatePickup(rng *rand.Rand) *audio.AudioSample {
	duration := 0.15
	frequency := 600.0 + rng.Float64()*200.0

	sample := g.osc.Generate(audio.WaveformTriangle, frequency, duration)

	env := synthesis.Envelope{
		Attack:  0.01,
		Decay:   0.03,
		Sustain: 0.5,
		Release: 0.1,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Add upward pitch sweep for positive feeling
	g.applyPitchBend(sample.Data, 1.0, 1.5)

	return sample
}

// generateHit creates a damage/hit sound.
func (g *Generator) generateHit(rng *rand.Rand) *audio.AudioSample {
	duration := 0.1
	frequency := 200.0 + rng.Float64()*100.0

	sample := g.osc.Generate(audio.WaveformSquare, frequency, duration)

	env := synthesis.Envelope{
		Attack:  0.001,
		Decay:   0.02,
		Sustain: 0.2,
		Release: 0.05,
	}
	env.Apply(sample.Data, sample.SampleRate)

	return sample
}

// generateJump creates a jump sound.
func (g *Generator) generateJump(rng *rand.Rand) *audio.AudioSample {
	duration := 0.2
	frequency := 300.0 + rng.Float64()*100.0

	sample := g.osc.Generate(audio.WaveformSquare, frequency, duration)

	env := synthesis.Envelope{
		Attack:  0.01,
		Decay:   0.05,
		Sustain: 0.3,
		Release: 0.1,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Upward pitch sweep
	g.applyPitchBend(sample.Data, 1.0, 1.3)

	return sample
}

// generateDeath creates a death/defeat sound.
func (g *Generator) generateDeath(rng *rand.Rand) *audio.AudioSample {
	duration := 0.8
	frequency := 440.0 + rng.Float64()*100.0

	sample := g.osc.Generate(audio.WaveformSawtooth, frequency, duration)

	env := synthesis.Envelope{
		Attack:  0.01,
		Decay:   0.2,
		Sustain: 0.3,
		Release: 0.4,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Downward pitch sweep for sad feeling
	g.applyPitchBend(sample.Data, 1.0, 0.5)

	return sample
}

// generatePowerup creates an energizing powerup sound.
func (g *Generator) generatePowerup(rng *rand.Rand) *audio.AudioSample {
	duration := 0.4
	frequency := 500.0 + rng.Float64()*200.0

	sample := g.osc.Generate(audio.WaveformSine, frequency, duration)

	env := synthesis.Envelope{
		Attack:  0.02,
		Decay:   0.1,
		Sustain: 0.6,
		Release: 0.2,
	}
	env.Apply(sample.Data, sample.SampleRate)

	// Add ascending arpeggio
	fifth := g.osc.Generate(audio.WaveformSine, frequency*1.5, duration*0.3)
	octave := g.osc.Generate(audio.WaveformSine, frequency*2.0, duration*0.3)

	// Mix in the harmonics at different times
	fifthStart := len(sample.Data) / 3
	octaveStart := 2 * len(sample.Data) / 3

	for i := range fifth.Data {
		if fifthStart+i < len(sample.Data) {
			sample.Data[fifthStart+i] += fifth.Data[i] * 0.5
		}
	}

	for i := range octave.Data {
		if octaveStart+i < len(sample.Data) {
			sample.Data[octaveStart+i] += octave.Data[i] * 0.3
		}
	}

	return sample
}

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
