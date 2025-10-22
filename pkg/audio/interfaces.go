package audio

// WaveformType represents different basic waveform types.
type WaveformType int

const (
	WaveformSine WaveformType = iota
	WaveformSquare
	WaveformSawtooth
	WaveformTriangle
	WaveformNoise
)

// Note represents a musical note with frequency and duration.
type Note struct {
	// Frequency in Hz
	Frequency float64

	// Duration in seconds
	Duration float64

	// Velocity (volume) from 0.0 to 1.0
	Velocity float64
}

// AudioSample represents a generated audio buffer.
type AudioSample struct {
	// SampleRate in Hz (e.g., 44100)
	SampleRate int

	// Data contains the audio samples (-1.0 to 1.0)
	Data []float64
}

// Synthesizer generates audio waveforms.
type Synthesizer interface {
	// Generate creates an audio sample from parameters
	Generate(waveform WaveformType, frequency float64, duration float64) *AudioSample

	// GenerateNote creates an audio sample for a musical note
	GenerateNote(note Note, waveform WaveformType) *AudioSample
}

// MusicGenerator creates procedural music.
type MusicGenerator interface {
	// GenerateTrack creates a music track for the given context
	GenerateTrack(genre string, context string, seed int64, duration float64) *AudioSample
}

// SFXGenerator creates sound effects.
type SFXGenerator interface {
	// Generate creates a sound effect for the given action
	Generate(effectType string, seed int64) *AudioSample
}

// AudioMixer manages playback of multiple audio sources.
type AudioMixer interface {
	// PlaySample plays an audio sample
	PlaySample(sample *AudioSample, loop bool)

	// Stop stops playback
	Stop()

	// SetVolume sets the master volume (0.0 to 1.0)
	SetVolume(volume float64)
}
