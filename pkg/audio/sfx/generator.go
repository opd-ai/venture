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
