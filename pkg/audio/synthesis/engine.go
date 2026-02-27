package synthesis

import (
	"sync"

	"github.com/opd-ai/venture/pkg/audio"
)

// Engine provides a unified audio synthesis API combining oscillators and envelopes.
// It is the primary interface for generating procedural audio waveforms.
type Engine struct {
	sampleRate int
	seed       int64
	osc        *Oscillator
	mu         sync.RWMutex
}

// Verify Engine implements audio.Synthesizer interface at compile time.
var _ audio.Synthesizer = (*Engine)(nil)

// NewEngine creates a new synthesis engine with the given seed.
func NewEngine(seed int64) *Engine {
	const defaultSampleRate = 44100
	return &Engine{
		sampleRate: defaultSampleRate,
		seed:       seed,
		osc:        NewOscillator(defaultSampleRate, seed),
	}
}

// NewEngineWithSampleRate creates a new synthesis engine with a custom sample rate.
// If sampleRate is less than or equal to 0, it defaults to 44100 Hz.
func NewEngineWithSampleRate(sampleRate int, seed int64) *Engine {
	if sampleRate <= 0 {
		sampleRate = 44100
	}
	return &Engine{
		sampleRate: sampleRate,
		seed:       seed,
		osc:        NewOscillator(sampleRate, seed),
	}
}

// GetSampleRate returns the engine's sample rate.
func (e *Engine) GetSampleRate() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sampleRate
}

// GetSeed returns the engine's seed for deterministic generation.
func (e *Engine) GetSeed() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.seed
}

// Generate generates a simple tone with the specified waveform, frequency, and duration.
// This method implements the audio.Synthesizer interface.
func (e *Engine) Generate(waveform audio.WaveformType, frequency, duration float64) *audio.AudioSample {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.osc.Generate(waveform, frequency, duration)
}

// GenerateToneWithEnvelope generates a tone shaped by an ADSR envelope.
func (e *Engine) GenerateToneWithEnvelope(waveform audio.WaveformType, frequency, duration float64, env Envelope) *audio.AudioSample {
	e.mu.Lock()
	defer e.mu.Unlock()

	sample := e.osc.Generate(waveform, frequency, duration)
	env.Apply(sample.Data, sample.SampleRate)
	return sample
}

// GenerateNote generates a musical note with velocity (volume) control.
func (e *Engine) GenerateNote(note audio.Note, waveform audio.WaveformType) *audio.AudioSample {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.osc.GenerateNote(note, waveform)
}

// GenerateNoteWithEnvelope generates a musical note shaped by an ADSR envelope.
func (e *Engine) GenerateNoteWithEnvelope(note audio.Note, waveform audio.WaveformType, env Envelope) *audio.AudioSample {
	e.mu.Lock()
	defer e.mu.Unlock()

	sample := e.osc.GenerateNote(note, waveform)
	env.Apply(sample.Data, sample.SampleRate)
	return sample
}

// GenerateChord generates multiple notes simultaneously, creating a chord.
func (e *Engine) GenerateChord(notes []audio.Note, waveform audio.WaveformType) *audio.AudioSample {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(notes) == 0 {
		return &audio.AudioSample{SampleRate: e.sampleRate, Data: []float64{}}
	}

	// Find longest duration to determine output length
	maxDuration := notes[0].Duration
	for _, n := range notes[1:] {
		if n.Duration > maxDuration {
			maxDuration = n.Duration
		}
	}

	numSamples := int(float64(e.sampleRate) * maxDuration)
	mixedData := make([]float64, numSamples)

	// Mix all notes together
	for _, note := range notes {
		noteSample := e.osc.GenerateNote(note, waveform)
		for i := 0; i < len(noteSample.Data) && i < numSamples; i++ {
			mixedData[i] += noteSample.Data[i]
		}
	}

	// Normalize to prevent clipping
	noteCount := float64(len(notes))
	for i := range mixedData {
		mixedData[i] /= noteCount
	}

	return &audio.AudioSample{
		SampleRate: e.sampleRate,
		Data:       mixedData,
	}
}

// GenerateChordWithEnvelope generates a chord shaped by an ADSR envelope.
// The envelope is applied to the generated chord in a thread-safe manner.
func (e *Engine) GenerateChordWithEnvelope(notes []audio.Note, waveform audio.WaveformType, env Envelope) *audio.AudioSample {
	sample := e.GenerateChord(notes, waveform)
	e.mu.Lock()
	defer e.mu.Unlock()
	env.Apply(sample.Data, sample.SampleRate)
	return sample
}

// MixSamples combines multiple audio samples into one, normalizing the output.
func (e *Engine) MixSamples(samples []*audio.AudioSample) *audio.AudioSample {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(samples) == 0 {
		return &audio.AudioSample{SampleRate: e.sampleRate, Data: []float64{}}
	}

	// Find longest sample
	maxLen := 0
	for _, s := range samples {
		if len(s.Data) > maxLen {
			maxLen = len(s.Data)
		}
	}

	mixedData := make([]float64, maxLen)
	for _, s := range samples {
		for i := 0; i < len(s.Data); i++ {
			mixedData[i] += s.Data[i]
		}
	}

	// Normalize
	sampleCount := float64(len(samples))
	for i := range mixedData {
		mixedData[i] /= sampleCount
	}

	return &audio.AudioSample{
		SampleRate: e.sampleRate,
		Data:       mixedData,
	}
}

// ApplyEnvelope applies an ADSR envelope to an existing audio sample.
// Note: The caller must ensure the sample is not being accessed by other
// goroutines during this call, as the sample data is modified in-place.
// For concurrent use, make a copy of the sample before calling this method.
func (e *Engine) ApplyEnvelope(sample *audio.AudioSample, env Envelope) {
	e.mu.Lock()
	defer e.mu.Unlock()
	env.Apply(sample.Data, sample.SampleRate)
}
