// Package synthesis provides audio waveform generation.
// This file implements oscillators for generating basic waveforms
// (sine, square, triangle, sawtooth) for audio synthesis.
package synthesis

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetReportCaller(true)
	log.SetLevel(logrus.InfoLevel)
}

// Oscillator generates basic waveforms for audio synthesis.
type Oscillator struct {
	sampleRate int
	rng        *rand.Rand
}

// NewOscillator creates a new oscillator with the given sample rate.
func NewOscillator(sampleRate int, seed int64) *Oscillator {
	log.WithFields(logrus.Fields{
		"sample_rate": sampleRate,
		"seed":        seed,
	}).Debug("Creating new oscillator")

	osc := &Oscillator{
		sampleRate: sampleRate,
		rng:        rand.New(rand.NewSource(seed)),
	}

	log.WithFields(logrus.Fields{
		"sample_rate": sampleRate,
		"seed":        seed,
	}).Debug("Oscillator created successfully")

	return osc
}

// Generate creates an audio sample with the specified waveform, frequency, and duration.
func (o *Oscillator) Generate(waveform audio.WaveformType, frequency, duration float64) *audio.AudioSample {
	log.WithFields(logrus.Fields{
		"waveform":    waveformName(waveform),
		"frequency":   frequency,
		"duration":    duration,
		"sample_rate": o.sampleRate,
	}).Debug("Starting waveform generation")

	numSamples := int(float64(o.sampleRate) * duration)
	data := make([]float64, numSamples)

	log.WithFields(logrus.Fields{
		"waveform":    waveformName(waveform),
		"num_samples": numSamples,
	}).Debug("Allocated sample buffer")

	switch waveform {
	case audio.WaveformSine:
		log.WithFields(logrus.Fields{
			"waveform":  waveformName(waveform),
			"frequency": frequency,
		}).Debug("Generating sine wave")
		o.generateSine(data, frequency)
	case audio.WaveformSquare:
		log.WithFields(logrus.Fields{
			"waveform":  waveformName(waveform),
			"frequency": frequency,
		}).Debug("Generating square wave")
		o.generateSquare(data, frequency)
	case audio.WaveformSawtooth:
		log.WithFields(logrus.Fields{
			"waveform":  waveformName(waveform),
			"frequency": frequency,
		}).Debug("Generating sawtooth wave")
		o.generateSawtooth(data, frequency)
	case audio.WaveformTriangle:
		log.WithFields(logrus.Fields{
			"waveform":  waveformName(waveform),
			"frequency": frequency,
		}).Debug("Generating triangle wave")
		o.generateTriangle(data, frequency)
	case audio.WaveformNoise:
		log.WithFields(logrus.Fields{
			"waveform": waveformName(waveform),
		}).Debug("Generating noise")
		o.generateNoise(data)
	}

	sample := &audio.AudioSample{
		SampleRate: o.sampleRate,
		Data:       data,
	}

	log.WithFields(logrus.Fields{
		"waveform":    waveformName(waveform),
		"frequency":   frequency,
		"duration":    duration,
		"num_samples": numSamples,
	}).Debug("Waveform generation complete")

	return sample
}

// GenerateNote creates an audio sample for a musical note.
func (o *Oscillator) GenerateNote(note audio.Note, waveform audio.WaveformType) *audio.AudioSample {
	log.WithFields(logrus.Fields{
		"waveform":  waveformName(waveform),
		"frequency": note.Frequency,
		"duration":  note.Duration,
		"velocity":  note.Velocity,
	}).Debug("Generating musical note")

	sample := o.Generate(waveform, note.Frequency, note.Duration)

	log.WithFields(logrus.Fields{
		"velocity":    note.Velocity,
		"num_samples": len(sample.Data),
	}).Debug("Applying velocity to note")

	// Apply velocity (volume)
	for i := range sample.Data {
		sample.Data[i] *= note.Velocity
	}

	log.WithFields(logrus.Fields{
		"waveform":  waveformName(waveform),
		"frequency": note.Frequency,
		"duration":  note.Duration,
		"velocity":  note.Velocity,
	}).Debug("Musical note generation complete")

	return sample
}

// generateSine creates a sine wave.
func (o *Oscillator) generateSine(data []float64, frequency float64) {
	log.WithFields(logrus.Fields{
		"waveform":    "sine",
		"frequency":   frequency,
		"num_samples": len(data),
	}).Debug("Entering sine wave generation")

	for i := range data {
		t := float64(i) / float64(o.sampleRate)
		data[i] = math.Sin(2 * math.Pi * frequency * t)
	}

	log.WithFields(logrus.Fields{
		"waveform":    "sine",
		"num_samples": len(data),
	}).Debug("Sine wave generation complete")
}

// generateSquare creates a square wave.
func (o *Oscillator) generateSquare(data []float64, frequency float64) {
	log.WithFields(logrus.Fields{
		"waveform":    "square",
		"frequency":   frequency,
		"num_samples": len(data),
	}).Debug("Entering square wave generation")

	for i := range data {
		t := float64(i) / float64(o.sampleRate)
		sine := math.Sin(2 * math.Pi * frequency * t)
		if sine >= 0 {
			data[i] = 1.0
		} else {
			data[i] = -1.0
		}
	}

	log.WithFields(logrus.Fields{
		"waveform":    "square",
		"num_samples": len(data),
	}).Debug("Square wave generation complete")
}

// generateSawtooth creates a sawtooth wave.
func (o *Oscillator) generateSawtooth(data []float64, frequency float64) {
	log.WithFields(logrus.Fields{
		"waveform":    "sawtooth",
		"frequency":   frequency,
		"num_samples": len(data),
	}).Debug("Entering sawtooth wave generation")

	period := float64(o.sampleRate) / frequency

	log.WithFields(logrus.Fields{
		"period": period,
	}).Debug("Calculated sawtooth period")

	for i := range data {
		t := float64(i)
		phase := math.Mod(t, period) / period
		data[i] = 2*phase - 1
	}

	log.WithFields(logrus.Fields{
		"waveform":    "sawtooth",
		"num_samples": len(data),
	}).Debug("Sawtooth wave generation complete")
}

// generateTriangle creates a triangle wave.
func (o *Oscillator) generateTriangle(data []float64, frequency float64) {
	log.WithFields(logrus.Fields{
		"waveform":    "triangle",
		"frequency":   frequency,
		"num_samples": len(data),
	}).Debug("Entering triangle wave generation")

	period := float64(o.sampleRate) / frequency

	log.WithFields(logrus.Fields{
		"period": period,
	}).Debug("Calculated triangle period")

	for i := range data {
		t := float64(i)
		phase := math.Mod(t, period) / period
		if phase < 0.5 {
			data[i] = 4*phase - 1
		} else {
			data[i] = 3 - 4*phase
		}
	}

	log.WithFields(logrus.Fields{
		"waveform":    "triangle",
		"num_samples": len(data),
	}).Debug("Triangle wave generation complete")
}

// generateNoise creates white noise.
func (o *Oscillator) generateNoise(data []float64) {
	log.WithFields(logrus.Fields{
		"waveform":    "noise",
		"num_samples": len(data),
	}).Debug("Entering noise generation")

	for i := range data {
		data[i] = o.rng.Float64()*2 - 1
	}

	log.WithFields(logrus.Fields{
		"waveform":    "noise",
		"num_samples": len(data),
	}).Debug("Noise generation complete")
}

// waveformName returns a string representation of the waveform type.
func waveformName(waveform audio.WaveformType) string {
	switch waveform {
	case audio.WaveformSine:
		return "sine"
	case audio.WaveformSquare:
		return "square"
	case audio.WaveformSawtooth:
		return "sawtooth"
	case audio.WaveformTriangle:
		return "triangle"
	case audio.WaveformNoise:
		return "noise"
	default:
		return "unknown"
	}
}
