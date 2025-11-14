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

// MelodyPattern defines melodic movement patterns.
type MelodyPattern int

const (
	PatternAscending MelodyPattern = iota // Upward motion
	PatternDescending                     // Downward motion
	PatternArpeggio                       // Chord arpeggiation
	PatternWave                           // Wave-like contour
	PatternRepeat                         // Repeated motif
)

// generateMelodyLayer creates a melodic line with patterns and variation.
func (ac *AdaptiveComposer) generateMelodyLayer(data []float64, scale Scale, beatDuration float64) {
	samplePos := 0
	noteDuration := beatDuration * 0.5 // Eighth notes
	
	// Choose melodic pattern based on context and genre
	pattern := ac.chooseMelodyPattern()
	motifLength := 4 // 4-note motifs
	scalePos := ac.rng.Intn(len(scale.Intervals))
	direction := 1

	for samplePos < len(data) {
		// Generate motif (4-note phrase)
		for i := 0; i < motifLength && samplePos < len(data); i++ {
			// Apply pattern to choose next note
			switch pattern {
			case PatternAscending:
				scalePos = (scalePos + 1) % len(scale.Intervals)
			case PatternDescending:
				scalePos = (scalePos - 1 + len(scale.Intervals)) % len(scale.Intervals)
			case PatternArpeggio:
				scalePos = (scalePos + 2) % len(scale.Intervals) // Skip notes (thirds)
			case PatternWave:
				scalePos = scalePos + direction
				// Handle wrap-around for both positive and negative
				if scalePos < 0 {
					scalePos = len(scale.Intervals) - 1
				} else if scalePos >= len(scale.Intervals) {
					scalePos = 0
				}
				if i%2 == 1 {
					direction = -direction // Reverse direction every other note
				}
			case PatternRepeat:
				// Keep same position for repetition
			}

			note := ac.rootNote + scale.Intervals[scalePos] + 12 // One octave above root
			
			// Add occasional octave jumps for interest
			if ac.rng.Float64() < 0.15 {
				note += 12
			}
			
			freq := NoteToFrequency(note)

			// Vary note duration for rhythm
			noteDur := noteDuration
			if i%4 == 3 && ac.rng.Float64() < 0.3 { // Occasional longer notes
				noteDur *= 1.5
			}

			// Generate note
			noteLen := int(float64(ac.sampleRate) * noteDur)
			noteSample := ac.osc.Generate(audio.WaveformTriangle, freq, noteDur)

			// Apply ADSR envelope with variation
			attack := 0.01
			if ac.currentContext == "combat" || ac.currentContext == "boss" {
				attack = 0.005 // Sharper attack in combat
			}
			env := synthesis.Envelope{Attack: attack, Decay: 0.1, Sustain: 0.6, Release: 0.2}
			env.Apply(noteSample.Data, noteSample.SampleRate)

			// Mix into data with dynamic volume
			volume := 0.5
			if i == 0 { // Accent first note of motif
				volume = 0.6
			}
			
			for j := 0; j < noteLen && samplePos+j < len(data); j++ {
				data[samplePos+j] += noteSample.Data[j] * volume
			}

			samplePos += noteLen
		}

		// Occasional rests between motifs for phrasing
		if ac.rng.Float64() < 0.2 {
			restLen := int(float64(ac.sampleRate) * noteDuration)
			samplePos += restLen
		}

		// Vary pattern occasionally
		if ac.rng.Float64() < 0.25 {
			pattern = ac.chooseMelodyPattern()
		}
	}
}

// chooseMelodyPattern selects a melodic pattern based on context and genre.
func (ac *AdaptiveComposer) chooseMelodyPattern() MelodyPattern {
	// Context influences pattern choice
	if ac.currentContext == "combat" || ac.currentContext == "boss" {
		// More aggressive patterns in combat
		patterns := []MelodyPattern{PatternAscending, PatternArpeggio, PatternWave}
		return patterns[ac.rng.Intn(len(patterns))]
	} else if ac.currentContext == "exploration" {
		// Calmer patterns for exploration
		patterns := []MelodyPattern{PatternWave, PatternRepeat, PatternDescending}
		return patterns[ac.rng.Intn(len(patterns))]
	}
	
	// Default: any pattern
	return MelodyPattern(ac.rng.Intn(5))
}

// ChordProgression defines common chord progressions.
type ChordProgression [][]int

var (
	// Common progressions (intervals from root)
	ProgressionPopular    = ChordProgression{{0, 4, 7}, {5, 9, 12}, {7, 11, 14}, {0, 4, 7}}         // I-IV-V-I
	ProgressionJazz       = ChordProgression{{2, 5, 9}, {7, 10, 14}, {0, 4, 7}}                      // ii-V-I
	ProgressionMinor      = ChordProgression{{0, 3, 7}, {5, 8, 12}, {7, 10, 14}, {0, 3, 7}}         // i-iv-V-i
	ProgressionTense      = ChordProgression{{0, 4, 7}, {1, 5, 8}, {7, 11, 14}, {0, 4, 7}}          // I-bII-V-I (tension)
	ProgressionDescending = ChordProgression{{0, 4, 7}, {10, 14, 17}, {7, 11, 14}, {5, 9, 12}}      // I-bVII-V-IV
)

// generateHarmonyLayer creates harmonic support with chord progressions.
func (ac *AdaptiveComposer) generateHarmonyLayer(data []float64, scale Scale, beatDuration float64) {
	// Choose progression based on context and genre
	progression := ac.chooseChordProgression()
	chordDuration := beatDuration * 4.0 // Whole notes per chord
	samplePos := 0
	chordIndex := 0

	for samplePos < len(data) {
		// Get current chord from progression
		chord := progression[chordIndex%len(progression)]
		chordLen := int(float64(ac.sampleRate) * chordDuration)

		// Generate each note in the chord
		for _, interval := range chord {
			note := ac.rootNote + interval
			freq := NoteToFrequency(note)
			
			// Use sine for smooth harmony
			noteSample := ac.osc.Generate(audio.WaveformSine, freq, chordDuration)

			// Apply gentle envelope for chord
			env := synthesis.Envelope{Attack: 0.05, Decay: 0.2, Sustain: 0.7, Release: 0.5}
			env.Apply(noteSample.Data, noteSample.SampleRate)

			// Mix chord note with reduced volume for subtlety
			volume := 0.25
			if interval == chord[0] { // Root note slightly louder
				volume = 0.3
			}

			for i := 0; i < chordLen && samplePos+i < len(data); i++ {
				data[samplePos+i] += noteSample.Data[i] * volume
			}
		}

		samplePos += chordLen
		chordIndex++
	}
}

// chooseChordProgression selects a progression based on context and genre.
func (ac *AdaptiveComposer) chooseChordProgression() ChordProgression {
	// Context influences progression choice
	switch ac.currentContext {
	case "combat", "boss":
		// Tense progressions for combat
		if ac.rng.Float64() < 0.6 {
			return ProgressionTense
		}
		return ProgressionMinor
	case "exploration":
		// Pleasant progressions for exploration
		return ProgressionPopular
	case "puzzle":
		// Jazz-influenced for puzzle thinking
		return ProgressionJazz
	case "victory":
		// Triumphant ascending progression
		return ProgressionDescending
	}

	// Genre influences progression choice
	switch ac.currentGenre {
	case "fantasy":
		return ProgressionPopular
	case "horror":
		return ProgressionMinor
	case "sci-fi", "cyberpunk":
		return ProgressionJazz
	default:
		return ProgressionPopular
	}
}

// DrumPattern defines when different drum sounds play in a measure.
type DrumPattern struct {
	Kick   []float64 // Beat positions for kick drum (0.0-1.0 in measure)
	Snare  []float64 // Beat positions for snare
	HiHat  []float64 // Beat positions for hi-hat
}

var (
	// Genre-specific drum patterns
	PatternRock      = DrumPattern{Kick: []float64{0.0, 0.5}, Snare: []float64{0.25, 0.75}, HiHat: []float64{0.0, 0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.875}}
	PatternElectronic = DrumPattern{Kick: []float64{0.0, 0.25, 0.5, 0.75}, Snare: []float64{0.5}, HiHat: []float64{0.125, 0.375, 0.625, 0.875}}
	PatternOrchestral = DrumPattern{Kick: []float64{0.0}, Snare: []float64{0.5}, HiHat: []float64{0.25, 0.75}}
	PatternIndustrial = DrumPattern{Kick: []float64{0.0, 0.33, 0.66}, Snare: []float64{0.25, 0.75}, HiHat: []float64{0.16, 0.5, 0.83}}
	PatternMinimal    = DrumPattern{Kick: []float64{0.0, 0.5}, Snare: []float64{}, HiHat: []float64{0.25, 0.75}}
)

// generatePercussionLayer creates rhythmic percussion with genre-specific patterns.
func (ac *AdaptiveComposer) generatePercussionLayer(data []float64, beatDuration float64) {
	// Choose pattern based on genre
	pattern := ac.chooseDrumPattern()
	measureDuration := beatDuration * 4.0 // 4 beats per measure
	measureSamples := int(float64(ac.sampleRate) * measureDuration)
	
	measurePos := 0
	for measurePos < len(data) {
		// Generate kick drum hits
		for _, pos := range pattern.Kick {
			hitSample := int(measurePos) + int(pos*float64(measureSamples))
			if hitSample < len(data) {
				ac.generateKickDrum(data, hitSample)
			}
		}
		
		// Generate snare drum hits
		for _, pos := range pattern.Snare {
			hitSample := int(measurePos) + int(pos*float64(measureSamples))
			if hitSample < len(data) {
				ac.generateSnareDrum(data, hitSample)
			}
		}
		
		// Generate hi-hat hits
		for _, pos := range pattern.HiHat {
			hitSample := int(measurePos) + int(pos*float64(measureSamples))
			if hitSample < len(data) {
				ac.generateHiHat(data, hitSample)
			}
		}
		
		measurePos += measureSamples
	}
}

// chooseDrumPattern selects a drum pattern based on genre.
func (ac *AdaptiveComposer) chooseDrumPattern() DrumPattern {
	switch ac.currentGenre {
	case "fantasy":
		return PatternOrchestral
	case "sci-fi", "cyberpunk":
		return PatternElectronic
	case "horror":
		return PatternIndustrial
	case "post-apocalyptic":
		return PatternMinimal
	default:
		return PatternRock
	}
}

// generateKickDrum creates a kick drum sound.
func (ac *AdaptiveComposer) generateKickDrum(data []float64, startPos int) {
	duration := 0.15
	samples := int(float64(ac.sampleRate) * duration)
	
	for i := 0; i < samples && startPos+i < len(data); i++ {
		t := float64(i) / float64(ac.sampleRate)
		// Exponential pitch sweep from 150Hz to 50Hz
		freq := 150.0 * math.Exp(-t*15.0) + 50.0
		// Exponential decay envelope
		env := math.Exp(-t * 25.0)
		// Generate sine wave with envelope
		data[startPos+i] += math.Sin(2.0*math.Pi*freq*t) * env * 0.5
	}
}

// generateSnareDrum creates a snare drum sound.
func (ac *AdaptiveComposer) generateSnareDrum(data []float64, startPos int) {
	duration := 0.1
	samples := int(float64(ac.sampleRate) * duration)
	
	for i := 0; i < samples && startPos+i < len(data); i++ {
		t := float64(i) / float64(ac.sampleRate)
		// Exponential decay
		env := math.Exp(-t * 30.0)
		// Mix of tone (200Hz) and noise for snare character
		tone := math.Sin(2.0 * math.Pi * 200.0 * t) * 0.3
		noise := (ac.rng.Float64()*2.0 - 1.0) * 0.7 // White noise
		data[startPos+i] += (tone + noise) * env * 0.3
	}
}

// generateHiHat creates a hi-hat sound.
func (ac *AdaptiveComposer) generateHiHat(data []float64, startPos int) {
	duration := 0.05
	samples := int(float64(ac.sampleRate) * duration)
	
	for i := 0; i < samples && startPos+i < len(data); i++ {
		t := float64(i) / float64(ac.sampleRate)
		// Very fast exponential decay
		env := math.Exp(-t * 50.0)
		// High-frequency noise for metallic sound
		noise := (ac.rng.Float64()*2.0 - 1.0)
		data[startPos+i] += noise * env * 0.15
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
