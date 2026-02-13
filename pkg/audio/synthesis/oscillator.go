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
// If sampleRate is less than or equal to 0, it defaults to 44100 Hz.
func NewOscillator(sampleRate int, seed int64) *Oscillator {
	if sampleRate <= 0 {
		sampleRate = 44100
	}

	log.WithFields(logrus.Fields{
		"sample_rate": sampleRate,
		"seed":        seed,
	}).Debug("Creating oscillator")

	return &Oscillator{
		sampleRate: sampleRate,
		rng:        rand.New(rand.NewSource(seed)),
	}
}

// Generate creates an audio sample with the specified waveform, frequency, and duration.
func (o *Oscillator) Generate(waveform audio.WaveformType, frequency, duration float64) *audio.AudioSample {
	numSamples := int(float64(o.sampleRate) * duration)
	data := make([]float64, numSamples)

	switch waveform {
	case audio.WaveformSine:
		o.generateSine(data, frequency)
	case audio.WaveformSquare:
		o.generateSquare(data, frequency)
	case audio.WaveformSawtooth:
		o.generateSawtooth(data, frequency)
	case audio.WaveformTriangle:
		o.generateTriangle(data, frequency)
	case audio.WaveformNoise:
		o.generateNoise(data)
	}

	return &audio.AudioSample{
		SampleRate: o.sampleRate,
		Data:       data,
	}
}

// GenerateNote creates an audio sample for a musical note.
func (o *Oscillator) GenerateNote(note audio.Note, waveform audio.WaveformType) *audio.AudioSample {
	sample := o.Generate(waveform, note.Frequency, note.Duration)

	// Apply velocity (volume)
	for i := range sample.Data {
		sample.Data[i] *= note.Velocity
	}

	return sample
}

// generateSine creates a sine wave.
func (o *Oscillator) generateSine(data []float64, frequency float64) {
	for i := range data {
		t := float64(i) / float64(o.sampleRate)
		data[i] = math.Sin(2 * math.Pi * frequency * t)
	}
}

// generateSquare creates a square wave.
func (o *Oscillator) generateSquare(data []float64, frequency float64) {
	for i := range data {
		t := float64(i) / float64(o.sampleRate)
		sine := math.Sin(2 * math.Pi * frequency * t)
		if sine >= 0 {
			data[i] = 1.0
		} else {
			data[i] = -1.0
		}
	}
}

// generateSawtooth creates a sawtooth wave.
func (o *Oscillator) generateSawtooth(data []float64, frequency float64) {
	period := float64(o.sampleRate) / frequency
	for i := range data {
		t := float64(i)
		phase := math.Mod(t, period) / period
		data[i] = 2*phase - 1
	}
}

// generateTriangle creates a triangle wave.
func (o *Oscillator) generateTriangle(data []float64, frequency float64) {
	period := float64(o.sampleRate) / frequency
	for i := range data {
		t := float64(i)
		phase := math.Mod(t, period) / period
		if phase < 0.5 {
			data[i] = 4*phase - 1
		} else {
			data[i] = 3 - 4*phase
		}
	}
}

// generateNoise creates white noise.
func (o *Oscillator) generateNoise(data []float64) {
	for i := range data {
		data[i] = o.rng.Float64()*2 - 1
	}
}

// WaveformName returns a string representation of the waveform type.
// This is useful for debugging and logging waveform types in audio packages.
func WaveformName(waveform audio.WaveformType) string {
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
