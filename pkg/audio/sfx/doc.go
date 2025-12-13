// Package sfx provides procedural sound effect generation.
// It creates various game sound effects (impacts, explosions, magic, etc.)
// using waveform synthesis and audio processing techniques.
//
// # Sound Variety
//
// Phase 1.1 (PLAN.md): Added VarietyManager for automatic sound variation.
// The VarietyManager caches multiple variants of each sound type with
// pitch and volume variations to avoid repetitive audio.
//
// # Usage
//
//	// Create variety manager for natural-sounding repeated sounds
//	vm := sfx.NewVarietyManager(44100, seed)
//	vm.SetVariantsPerEffect(5)
//	vm.SetPitchVariance(2.0)    // ±2 semitones
//	vm.SetVolumeVariance(0.2)   // ±20% volume
//
//	// Generate sound with automatic variety
//	sample := vm.Generate("impact", seed)
//
// Performance: VarietyManager.Generate() averages 8.875 ns/op with caching.
package sfx
