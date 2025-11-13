// Package music provides procedural music composition with adaptive features.
//
// # Overview
//
// This package generates background music using music theory principles and
// genre-appropriate scales, rhythms, and instruments. It supports dynamic
// adaptation based on gameplay context through layered composition.
//
// # Key Features
//
//   - Adaptive music that responds to gameplay context (exploration, combat, boss battles, etc.)
//   - Layered composition with independent control over ambient, melody, harmony, percussion, and intensity layers
//   - Smooth crossfade transitions between contexts (<1 second)
//   - Genre-specific instrumentation and scales (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
//   - Deterministic generation from seed values for reproducibility
//
// # Basic Usage
//
//	// Create an adaptive music composer
//	composer := music.NewAdaptiveComposer(44100, seed)
//	composer.Initialize("fantasy", 60) // Fantasy genre, root note C4
//
//	// Set gameplay context
//	composer.SetContext("exploration")
//
//	// Generate a 10-second music track
//	track := composer.GenerateAdaptiveTrack(10.0)
//
//	// Change context dynamically
//	composer.SetContext("combat")
//	composer.Update(deltaTime) // Smooth transition
//
// # Layer System
//
// The adaptive music system uses five independent layers:
//
//   - Ambient: Low-frequency pad providing atmospheric foundation
//   - Melody: Primary melodic line using genre-appropriate scales
//   - Harmony: Harmonic support with chord progressions
//   - Percussion: Rhythmic elements (drums, beats)
//   - Intensity: High-frequency layer for tension and excitement
//
// Layers can be individually controlled:
//
//	// Add percussion layer
//	composer.AddLayer(audio.MusicLayerPercussion)
//	composer.Update(1.0) // Fade in over time
//
//	// Remove harmony layer
//	composer.RemoveLayer(audio.MusicLayerHarmony)
//	composer.Update(1.0) // Fade out over time
//
// # Context System
//
// Music adapts to gameplay situations:
//
//   - "exploration": Calm, ambient music (ambient + melody)
//   - "combat": Energetic music with percussion (all layers except intensity)
//   - "boss": Epic music with maximum intensity (all layers active)
//   - "puzzle": Contemplative music (ambient + light melody)
//   - "victory": Triumphant music (melody + harmony + percussion)
//
// # Intensity Scaling
//
// Fine-grained control over tension level:
//
//	// Low danger (0.0-0.3): Calm exploration
//	composer.UpdateIntensity(0.2)
//
//	// Medium danger (0.3-0.7): Heightened alertness
//	composer.UpdateIntensity(0.5)
//
//	// High danger (0.7-1.0): Maximum tension
//	composer.UpdateIntensity(0.9)
//
// # Interface Implementation
//
// AdaptiveComposer implements the audio.AdaptiveMusicSystem interface:
//
//	type AdaptiveMusicSystem interface {
//	    SetContext(context MusicContext) error
//	    UpdateIntensity(intensity float64) error
//	    AddLayer(layer MusicLayer) error
//	    RemoveLayer(layer MusicLayer) error
//	    Update(deltaTime float64)
//	    GenerateTrack(duration float64) *AudioSample
//	}
//
// # Performance
//
// Typical performance characteristics:
//
//   - Track generation: <50ms for 10 seconds of music
//   - Layer updates: <0.1ms per frame
//   - Memory usage: <5MB per composer instance
//   - Smooth transitions: Complete within 1 second
//
// # Testing
//
// Use the cmd/musictest tool for manual validation:
//
//	go run ./cmd/musictest -mode all -genre fantasy -seed 12345
//
// Test modes include:
//
//   - contexts: Test all music contexts
//   - layers: Test individual layer activation
//   - transitions: Test smooth context transitions
//   - intensity: Test intensity scaling
//   - all: Run all tests
//
// # Music Theory
//
// The package uses established music theory:
//
//   - Scales: Major, minor, pentatonic, chromatic, blues, whole tone
//   - Chord progressions: I-IV-V-I, ii-V-I, and genre-specific progressions
//   - Rhythms: Tempo ranges from 80 BPM (puzzle) to 160 BPM (boss)
//   - ADSR envelopes: Attack, decay, sustain, release for natural sound
//
// # Genre-Specific Features
//
// Each genre has distinct musical characteristics:
//
//   - Fantasy: Major/minor scales, orchestral feel, moderate tempo
//   - Sci-Fi: Whole tone scales, synthetic sounds, electronic percussion
//   - Horror: Chromatic scales, dissonant harmonies, unsettling rhythms
//   - Cyberpunk: Pentatonic scales, driving beats, neon energy
//   - Post-Apocalyptic: Blues scales, sparse instrumentation, slow tempo
package music
