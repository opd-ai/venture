// Package sfx provides procedural sound effect generation.
// This file implements the core Generator type and public API methods.
//
// The Generator creates various game sound effects using waveform synthesis,
// including support for genre-specific variations.
package sfx

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/synthesis"
	"github.com/sirupsen/logrus"
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
		g.applyScifiModifications(sample)
	case "horror":
		g.applyHorrorModifications(sample)
	case "cyberpunk":
		g.applyCyberpunkModifications(sample)
	case "postapoc":
		g.applyPostApocalypticModifications(sample)
	}
}

// applyScifiModifications applies synthetic, clean, higher-pitch modifications for sci-fi genre.
func (g *Generator) applyScifiModifications(sample *audio.AudioSample) {
	// Synthetic, clean, higher pitch
	g.applyPitchBend(sample.Data, 1.3, 1.3)
	g.reduceAmplitude(sample.Data, 0.9)
}

// applyHorrorModifications applies dissonant, unsettling, lower-pitch modifications for horror genre.
func (g *Generator) applyHorrorModifications(sample *audio.AudioSample) {
	// Dissonant, unsettling, lower pitch
	g.applyPitchBend(sample.Data, 0.7, 0.7)
	// Add vibrato for unsettling effect
	g.applyVibrato(sample.Data, 3.0, 0.2)
}

// applyCyberpunkModifications applies sharp, electronic, glitchy modifications for cyberpunk genre.
func (g *Generator) applyCyberpunkModifications(sample *audio.AudioSample) {
	// Sharp, electronic, glitchy
	g.applyPitchBend(sample.Data, 1.4, 1.4)
	g.applyHardClipping(sample.Data, 0.7)
}

// applyPostApocalypticModifications applies harsh, gritty, distorted modifications for post-apocalyptic genre.
func (g *Generator) applyPostApocalypticModifications(sample *audio.AudioSample) {
	// Harsh, gritty, distorted
	g.applyPitchBend(sample.Data, 0.9, 0.9)
	g.applySoftClipping(sample.Data, 0.5, 0.3)
}

// reduceAmplitude multiplies all samples by the given factor for cleaner sound.
func (g *Generator) reduceAmplitude(data []float64, factor float64) {
	for i := range data {
		data[i] *= factor
	}
}

// applyHardClipping applies hard clipping at the specified threshold for digital effect.
func (g *Generator) applyHardClipping(data []float64, threshold float64) {
	for i := range data {
		if data[i] > threshold {
			data[i] = threshold
		} else if data[i] < -threshold {
			data[i] = -threshold
		}
	}
}

// applySoftClipping applies soft clipping at threshold with the given compression factor for gritty effect.
func (g *Generator) applySoftClipping(data []float64, threshold, compression float64) {
	for i := range data {
		if data[i] > threshold {
			data[i] = threshold + (data[i]-threshold)*compression
		} else if data[i] < -threshold {
			data[i] = -threshold + (data[i]+threshold)*compression
		}
	}
}
