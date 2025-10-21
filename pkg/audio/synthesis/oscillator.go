package synthesis

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
)

// Oscillator generates basic waveforms for audio synthesis.
type Oscillator struct {
	sampleRate int
	rng        *rand.Rand
}

// NewOscillator creates a new oscillator with the given sample rate.
func NewOscillator(sampleRate int, seed int64) *Oscillator {
	return &Oscillator{
		sampleRate: sampleRate,
		rng:        rand.New(rand.NewSource(seed)),
	}
}

// Generate creates an audio sample with the specified waveform, frequency, and duration.
func (o *Oscillator) Generate(waveform audio.WaveformType, frequency float64, duration float64) *audio.AudioSample {
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
