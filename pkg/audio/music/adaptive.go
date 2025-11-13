// Package music provides adaptive music composition.
// This file implements dynamic layer management for context-aware music.
package music

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/synthesis"
)

// MusicLayer represents a distinct audio layer in adaptive composition.
// Layers can be dynamically added or removed based on gameplay context.
type MusicLayer struct {
	// Name identifies the layer (e.g., "percussion", "melody", "harmony")
	Name string
	// Active indicates if the layer is currently playing
	Active bool
	// Volume controls the layer's volume (0.0-1.0)
	Volume float64
	// TargetVolume is the desired volume for smooth transitions
	TargetVolume float64
	// Data contains the audio samples for this layer
	Data []float64
	// Waveform is the wave type for this layer
	Waveform audio.WaveformType
	// Frequency is the base frequency for this layer
	Frequency float64
}

// AdaptiveComposer creates music that adapts to gameplay context.
type AdaptiveComposer struct {
	sampleRate int
	seed       int64
	osc        *synthesis.Oscillator
	rng        *rand.Rand

	// Layers contains all available music layers
	layers map[string]*MusicLayer

	// Current composition state
	currentGenre   string
	currentContext string
	tempo          float64
	rootNote       int
}

// AdaptiveMusicManager wraps AdaptiveComposer to implement audio.AdaptiveMusicSystem interface.
type AdaptiveMusicManager struct {
	composer *AdaptiveComposer
}

// NewAdaptiveMusicManager creates a new AdaptiveMusicManager that implements audio.AdaptiveMusicSystem.
func NewAdaptiveMusicManager(sampleRate int, seed int64) *AdaptiveMusicManager {
	return &AdaptiveMusicManager{
		composer: NewAdaptiveComposer(sampleRate, seed),
	}
}

// Initialize sets up the manager with genre and root note.
func (amm *AdaptiveMusicManager) Initialize(genre string, rootNote int) {
	amm.composer.Initialize(genre, rootNote)
}

// SetContext updates music based on a MusicContext struct (implements interface).
func (amm *AdaptiveMusicManager) SetContext(ctx audio.MusicContext) error {
	return amm.composer.SetContextFromStruct(ctx)
}

// UpdateIntensity adjusts the intensity layer volume (implements interface).
func (amm *AdaptiveMusicManager) UpdateIntensity(intensity float64) error {
	return amm.composer.UpdateIntensity(intensity)
}

// AddLayer activates a specific music layer (implements interface).
func (amm *AdaptiveMusicManager) AddLayer(layer audio.MusicLayer) error {
	return amm.composer.AddLayer(layer)
}

// RemoveLayer deactivates a specific music layer (implements interface).
func (amm *AdaptiveMusicManager) RemoveLayer(layer audio.MusicLayer) error {
	return amm.composer.RemoveLayer(layer)
}

// Update performs smooth transitions (implements interface).
func (amm *AdaptiveMusicManager) Update(deltaTime float64) {
	amm.composer.Update(deltaTime)
}

// GenerateTrack creates an audio sample (implements interface).
func (amm *AdaptiveMusicManager) GenerateTrack(duration float64) *audio.AudioSample {
	return amm.composer.GenerateTrack(duration)
}

// GetActiveLayerCount returns the number of active layers.
func (amm *AdaptiveMusicManager) GetActiveLayerCount() int {
	return amm.composer.GetActiveLayerCount()
}

// GetLayerVolume returns the volume of a specific layer.
func (amm *AdaptiveMusicManager) GetLayerVolume(layerName string) float64 {
	return amm.composer.GetLayerVolume(layerName)
}

// NewAdaptiveComposer creates a new adaptive music composer.
func NewAdaptiveComposer(sampleRate int, seed int64) *AdaptiveComposer {
	return &AdaptiveComposer{
		sampleRate: sampleRate,
		seed:       seed,
		osc:        synthesis.NewOscillator(sampleRate, seed),
		rng:        rand.New(rand.NewSource(seed)),
		layers:     make(map[string]*MusicLayer),
	}
}

// Initialize sets up the composition with genre and root note.
func (ac *AdaptiveComposer) Initialize(genre string, rootNote int) {
	ac.currentGenre = genre
	ac.rootNote = rootNote
	ac.tempo = 120.0 // Default tempo

	// Initialize base layers
	ac.layers["ambient"] = &MusicLayer{
		Name:         "ambient",
		Active:       true,
		Volume:       0.3,
		TargetVolume: 0.3,
		Waveform:     audio.WaveformSine,
	}

	ac.layers["melody"] = &MusicLayer{
		Name:         "melody",
		Active:       true,
		Volume:       0.4,
		TargetVolume: 0.4,
		Waveform:     audio.WaveformTriangle,
	}

	ac.layers["harmony"] = &MusicLayer{
		Name:         "harmony",
		Active:       false,
		Volume:       0.0,
		TargetVolume: 0.0,
		Waveform:     audio.WaveformSine,
	}

	ac.layers["percussion"] = &MusicLayer{
		Name:         "percussion",
		Active:       false,
		Volume:       0.0,
		TargetVolume: 0.0,
		Waveform:     audio.WaveformSquare,
	}

	ac.layers["intensity"] = &MusicLayer{
		Name:         "intensity",
		Active:       false,
		Volume:       0.0,
		TargetVolume: 0.0,
		Waveform:     audio.WaveformSawtooth,
	}
}

// SetContext adapts the music to a new gameplay context.
// Context examples: "exploration", "combat", "boss", "puzzle", "victory"
func (ac *AdaptiveComposer) SetContext(context string) {
	ac.currentContext = context

	// Adjust layer volumes and activation based on context
	switch context {
	case "exploration":
		ac.setLayerTarget("ambient", 0.4, true)
		ac.setLayerTarget("melody", 0.5, true)
		ac.setLayerTarget("harmony", 0.0, false)
		ac.setLayerTarget("percussion", 0.0, false)
		ac.setLayerTarget("intensity", 0.0, false)
		ac.tempo = 90.0

	case "combat":
		ac.setLayerTarget("ambient", 0.2, true)
		ac.setLayerTarget("melody", 0.4, true)
		ac.setLayerTarget("harmony", 0.3, true)
		ac.setLayerTarget("percussion", 0.5, true)
		ac.setLayerTarget("intensity", 0.3, true)
		ac.tempo = 140.0

	case "boss":
		ac.setLayerTarget("ambient", 0.1, true)
		ac.setLayerTarget("melody", 0.5, true)
		ac.setLayerTarget("harmony", 0.4, true)
		ac.setLayerTarget("percussion", 0.6, true)
		ac.setLayerTarget("intensity", 0.5, true)
		ac.tempo = 160.0

	case "puzzle":
		ac.setLayerTarget("ambient", 0.5, true)
		ac.setLayerTarget("melody", 0.3, true)
		ac.setLayerTarget("harmony", 0.0, false)
		ac.setLayerTarget("percussion", 0.0, false)
		ac.setLayerTarget("intensity", 0.0, false)
		ac.tempo = 80.0

	case "victory":
		ac.setLayerTarget("ambient", 0.3, true)
		ac.setLayerTarget("melody", 0.6, true)
		ac.setLayerTarget("harmony", 0.5, true)
		ac.setLayerTarget("percussion", 0.4, true)
		ac.setLayerTarget("intensity", 0.0, false)
		ac.tempo = 120.0

	default:
		// Default to exploration
		ac.SetContext("exploration")
	}
}

// setLayerTarget sets the target volume and active state for a layer.
func (ac *AdaptiveComposer) setLayerTarget(layerName string, volume float64, active bool) {
	if layer, exists := ac.layers[layerName]; exists {
		layer.TargetVolume = volume
		if volume > 0.0 {
			layer.Active = active
		}
	}
}

// UpdateLayers smoothly transitions layer volumes toward their targets.
// transitionSpeed controls how fast layers fade in/out (0.0-1.0).
func (ac *AdaptiveComposer) UpdateLayers(transitionSpeed float64) {
	for _, layer := range ac.layers {
		// Smooth volume transition
		diff := layer.TargetVolume - layer.Volume
		layer.Volume += diff * transitionSpeed

		// Deactivate layer if volume reaches zero
		if layer.Volume < 0.01 && layer.TargetVolume == 0.0 {
			layer.Active = false
			layer.Volume = 0.0
		}
	}
}

// GenerateAdaptiveTrack creates a music track with the current layer configuration.
func (ac *AdaptiveComposer) GenerateAdaptiveTrack(duration float64) *audio.AudioSample {
	numSamples := int(float64(ac.sampleRate) * duration)
	track := make([]float64, numSamples)

	// Generate each layer
	for _, layer := range ac.layers {
		if !layer.Active || layer.Volume <= 0.0 {
			continue
		}

		// Generate layer data based on layer type
		layerData := ac.generateLayer(layer, duration)

		// Mix layer into track with volume
		for i := 0; i < len(track) && i < len(layerData); i++ {
			track[i] += layerData[i] * layer.Volume
		}
	}

	// Normalize track to prevent clipping
	ac.normalizeTrack(track)

	// Apply master envelope for smooth start/end
	ac.applyMasterEnvelope(track, duration)

	return &audio.AudioSample{
		SampleRate: ac.sampleRate,
		Data:       track,
	}
}

// generateLayer creates audio data for a specific layer.
func (ac *AdaptiveComposer) generateLayer(layer *MusicLayer, duration float64) []float64 {
	numSamples := int(float64(ac.sampleRate) * duration)
	data := make([]float64, numSamples)

	scale := GetScaleForGenre(ac.currentGenre)
	beatDuration := 60.0 / ac.tempo // seconds per beat

	switch layer.Name {
	case "ambient":
		// Slow-moving pad sound
		freq := NoteToFrequency(ac.rootNote - 12) // One octave below root
		sample := ac.osc.Generate(layer.Waveform, freq, duration)
		copy(data, sample.Data)

	case "melody":
		// Melodic line
		ac.generateMelodyLayer(data, scale, beatDuration)

	case "harmony":
		// Harmonic support
		ac.generateHarmonyLayer(data, scale, beatDuration)

	case "percussion":
		// Rhythmic percussion
		ac.generatePercussionLayer(data, beatDuration)

	case "intensity":
		// High-frequency intensity layer
		ac.generateIntensityLayer(data, beatDuration)
	}

	return data
}

// generateMelodyLayer creates a melodic line.
func (ac *AdaptiveComposer) generateMelodyLayer(data []float64, scale Scale, beatDuration float64) {
	samplePos := 0
	noteDuration := beatDuration * 0.5 // Eighth notes

	for samplePos < len(data) {
		// Choose note from scale
		scaleIndex := ac.rng.Intn(len(scale.Intervals))
		note := ac.rootNote + scale.Intervals[scaleIndex] + 12 // One octave above root
		freq := NoteToFrequency(note)

		// Generate note
		noteLen := int(float64(ac.sampleRate) * noteDuration)
		noteSample := ac.osc.Generate(audio.WaveformTriangle, freq, noteDuration)

		// Apply ADSR envelope
		env := synthesis.Envelope{Attack: 0.01, Decay: 0.1, Sustain: 0.6, Release: 0.2}
		env.Apply(noteSample.Data, noteSample.SampleRate)

		// Mix into data
		for i := 0; i < noteLen && samplePos+i < len(data); i++ {
			data[samplePos+i] += noteSample.Data[i] * 0.5
		}

		samplePos += noteLen
	}
}

// generateHarmonyLayer creates harmonic support.
func (ac *AdaptiveComposer) generateHarmonyLayer(data []float64, scale Scale, beatDuration float64) {
	// Generate sustained chords
	chordDuration := beatDuration * 4.0 // Whole notes
	samplePos := 0

	for samplePos < len(data) {
		// Generate major triad from root
		chord := []int{0, 4, 7} // Major triad intervals
		chordLen := int(float64(ac.sampleRate) * chordDuration)

		for _, interval := range chord {
			note := ac.rootNote + interval
			freq := NoteToFrequency(note)
			noteSample := ac.osc.Generate(audio.WaveformSine, freq, chordDuration)

			// Mix chord note
			for i := 0; i < chordLen && samplePos+i < len(data); i++ {
				data[samplePos+i] += noteSample.Data[i] * 0.3
			}
		}

		samplePos += chordLen
	}
}

// generatePercussionLayer creates rhythmic percussion.
func (ac *AdaptiveComposer) generatePercussionLayer(data []float64, beatDuration float64) {
	beatSamples := int(float64(ac.sampleRate) * beatDuration)

	for i := 0; i < len(data); i += beatSamples {
		// Generate kick drum sound using low-frequency pulse
		kickDuration := 0.1
		kickSamples := int(float64(ac.sampleRate) * kickDuration)

		for j := 0; j < kickSamples && i+j < len(data); j++ {
			t := float64(j) / float64(ac.sampleRate)
			// Exponential decay envelope
			env := math.Exp(-t * 20.0)
			// Low sine wave for kick
			data[i+j] += math.Sin(2.0*math.Pi*60.0*t) * env * 0.4
		}
	}
}

// generateIntensityLayer creates a high-frequency intensity layer.
func (ac *AdaptiveComposer) generateIntensityLayer(data []float64, beatDuration float64) {
	// High-frequency sustained note for intensity
	freq := NoteToFrequency(ac.rootNote + 24) // Two octaves above root
	sample := ac.osc.Generate(audio.WaveformSawtooth, freq, float64(len(data))/float64(ac.sampleRate))

	for i := 0; i < len(data) && i < len(sample.Data); i++ {
		data[i] += sample.Data[i] * 0.3
	}
}

// normalizeTrack prevents clipping by scaling amplitude.
func (ac *AdaptiveComposer) normalizeTrack(track []float64) {
	maxAmp := 0.0
	for _, sample := range track {
		amp := math.Abs(sample)
		if amp > maxAmp {
			maxAmp = amp
		}
	}

	if maxAmp > 1.0 {
		scale := 0.95 / maxAmp
		for i := range track {
			track[i] *= scale
		}
	}
}

// applyMasterEnvelope applies fade in/out to the track.
func (ac *AdaptiveComposer) applyMasterEnvelope(track []float64, duration float64) {
	fadeInTime := 0.5  // 0.5 second fade in
	fadeOutTime := 1.0 // 1.0 second fade out

	fadeInSamples := int(float64(ac.sampleRate) * fadeInTime)
	fadeOutSamples := int(float64(ac.sampleRate) * fadeOutTime)

	// Fade in
	for i := 0; i < fadeInSamples && i < len(track); i++ {
		env := float64(i) / float64(fadeInSamples)
		track[i] *= env
	}

	// Fade out
	for i := 0; i < fadeOutSamples && i < len(track); i++ {
		idx := len(track) - 1 - i
		env := float64(i) / float64(fadeOutSamples)
		track[idx] *= env
	}
}

// GetActiveLayerCount returns the number of currently active layers.
func (ac *AdaptiveComposer) GetActiveLayerCount() int {
	count := 0
	for _, layer := range ac.layers {
		if layer.Active && layer.Volume > 0.01 {
			count++
		}
	}
	return count
}

// GetLayerVolume returns the current volume for a layer.
func (ac *AdaptiveComposer) GetLayerVolume(layerName string) float64 {
	if layer, exists := ac.layers[layerName]; exists {
		return layer.Volume
	}
	return 0.0
}

// SetContextFromStruct updates music based on a MusicContext struct.
// This implements the audio.AdaptiveMusicSystem interface.
func (ac *AdaptiveComposer) SetContextFromStruct(ctx audio.MusicContext) error {
	// Map MusicContext to string context and update danger-based intensity
	context := "exploration" // Default

	if ctx.BossNearby {
		context = "boss"
	} else if ctx.Combat {
		context = "combat"
	} else if ctx.Danger > 0.7 {
		context = "combat"
	} else if ctx.Danger > 0.3 {
		// Moderate danger - stay in exploration but increase intensity
		context = "exploration"
	}

	// Special contexts based on location
	if ctx.Location == "town" || ctx.Location == "settlement" {
		context = "exploration"
		ac.tempo = 80.0 // Slower tempo for towns
	} else if ctx.Location == "victory" {
		context = "victory"
	}

	ac.SetContext(context)

	// Adjust intensity based on danger level
	return ac.UpdateIntensity(ctx.Danger)
}

// UpdateIntensity adjusts the intensity layer volume based on a 0.0-1.0 scale.
// This implements the audio.AdaptiveMusicSystem interface.
func (ac *AdaptiveComposer) UpdateIntensity(intensity float64) error {
	// Clamp intensity to valid range
	if intensity < 0.0 {
		intensity = 0.0
	} else if intensity > 1.0 {
		intensity = 1.0
	}

	// Update intensity layer target volume
	if layer, exists := ac.layers["intensity"]; exists {
		layer.TargetVolume = intensity * 0.5 // Scale to max 0.5 volume
		if intensity > 0.0 {
			layer.Active = true
		}
	}

	return nil
}

// AddLayer activates a specific music layer.
// This implements the audio.AdaptiveMusicSystem interface.
func (ac *AdaptiveComposer) AddLayer(layer audio.MusicLayer) error {
	layerName := layer.String()

	if musicLayer, exists := ac.layers[layerName]; exists {
		// Set target volume based on layer type
		targetVolume := 0.4 // Default
		switch layer {
		case audio.MusicLayerBase:
			targetVolume = 0.3
		case audio.MusicLayerMelody:
			targetVolume = 0.5
		case audio.MusicLayerHarmony:
			targetVolume = 0.4
		case audio.MusicLayerPercussion:
			targetVolume = 0.5
		case audio.MusicLayerIntensity:
			targetVolume = 0.4
		}

		musicLayer.TargetVolume = targetVolume
		musicLayer.Active = true
		return nil
	}

	return nil
}

// RemoveLayer deactivates a specific music layer.
// This implements the audio.AdaptiveMusicSystem interface.
func (ac *AdaptiveComposer) RemoveLayer(layer audio.MusicLayer) error {
	layerName := layer.String()

	if musicLayer, exists := ac.layers[layerName]; exists {
		musicLayer.TargetVolume = 0.0
		return nil
	}

	return nil
}

// Update performs smooth layer transitions.
// This implements the audio.AdaptiveMusicSystem interface.
// deltaTime is the time elapsed since last update in seconds.
func (ac *AdaptiveComposer) Update(deltaTime float64) {
	// Calculate transition speed based on deltaTime
	// Target: complete transition in ~1 second
	transitionSpeed := deltaTime
	if transitionSpeed > 1.0 {
		transitionSpeed = 1.0
	}

	ac.UpdateLayers(transitionSpeed)
}

// GenerateTrack creates an audio sample with current settings.
// This implements the audio.AdaptiveMusicSystem interface.
func (ac *AdaptiveComposer) GenerateTrack(duration float64) *audio.AudioSample {
	return ac.GenerateAdaptiveTrack(duration)
}
