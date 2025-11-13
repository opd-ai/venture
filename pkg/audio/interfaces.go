// Package audio provides core audio synthesis interfaces.
// This file defines audio interfaces and waveform types used by
// the audio generation and synthesis subsystem.
package audio

// WaveformType represents different basic waveform types.
type WaveformType int

// Waveform type constants.
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
	Generate(waveform WaveformType, frequency, duration float64) *AudioSample

	// GenerateNote creates an audio sample for a musical note
	GenerateNote(note Note, waveform WaveformType) *AudioSample
}

// MusicGenerator creates procedural music.
type MusicGenerator interface {
	// GenerateTrack creates a music track for the given context
	GenerateTrack(genre, context string, seed int64, duration float64) *AudioSample
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

// MusicContext represents the gameplay context for adaptive music.
type MusicContext struct {
	// Location describes where the player is (e.g., "dungeon", "town", "wilderness")
	Location string

	// Combat indicates if combat is active
	Combat bool

	// BossNearby indicates if a boss enemy is nearby
	BossNearby bool

	// TimeOfDay represents the time (e.g., "dawn", "day", "dusk", "night")
	TimeOfDay string

	// Danger level from 0.0 (safe) to 1.0 (deadly)
	Danger float64
}

// MusicLayer represents a layer in adaptive music composition.
type MusicLayer int

const (
	// MusicLayerBase is the foundational ambient layer
	MusicLayerBase MusicLayer = iota
	// MusicLayerHarmony adds harmonic support
	MusicLayerHarmony
	// MusicLayerPercussion adds rhythmic elements
	MusicLayerPercussion
	// MusicLayerMelody adds the primary melodic line
	MusicLayerMelody
	// MusicLayerIntensity adds high-energy intensity
	MusicLayerIntensity
)

// String returns the string representation of a MusicLayer.
func (m MusicLayer) String() string {
	switch m {
	case MusicLayerBase:
		return "base"
	case MusicLayerHarmony:
		return "harmony"
	case MusicLayerPercussion:
		return "percussion"
	case MusicLayerMelody:
		return "melody"
	case MusicLayerIntensity:
		return "intensity"
	default:
		return "unknown"
	}
}

// AdaptiveMusicSystem manages context-aware music composition.
type AdaptiveMusicSystem interface {
	// SetContext updates the music based on gameplay context
	SetContext(context MusicContext) error

	// UpdateIntensity adjusts the intensity level (0.0-1.0)
	UpdateIntensity(intensity float64) error

	// AddLayer activates a specific music layer
	AddLayer(layer MusicLayer) error

	// RemoveLayer deactivates a specific music layer
	RemoveLayer(layer MusicLayer) error

	// Update performs smooth transitions between states
	Update(deltaTime float64)

	// GenerateTrack creates an audio sample with current settings
	GenerateTrack(duration float64) *AudioSample
}
