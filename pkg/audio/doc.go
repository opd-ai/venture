// Package audio provides procedural audio synthesis for all sounds in the game.
// All audio is generated at runtime using waveform synthesis and music theory.
//
// The audio system supports procedural music composition, sound effects generation,
// and adaptive audio based on game context and genre.
//
// # Architecture
//
// The package provides a unified Manager that coordinates music and SFX systems:
//   - Manager: Unified audio manager with volume controls and dependency injection
//   - AdaptiveMusicSystem: Interface for context-aware music composition
//   - SFXGenerator: Interface for sound effect generation
//
// Sub-packages:
//   - music: Adaptive music composition with genre awareness
//   - sfx: Sound effect generation with variety management
//   - synthesis: Low-level waveform synthesis and envelopes
//
// # Usage
//
//	// Create unified manager
//	manager := audio.NewManager(44100, seed)
//
//	// Set music system
//	musicSystem := music.NewAdaptiveMusicManager(44100, seed)
//	manager.SetMusicManager(musicSystem)
//
//	// Set SFX system
//	sfxSystem := sfx.NewVarietyManager(44100, seed)
//	manager.SetSFXManager(sfxSystem)
//
//	// Generate audio
//	musicSample := manager.GenerateMusicTrack(2.0)
//	sfxSample := manager.GenerateSFX("impact", 12345)
//
// Phase 1.1 (PLAN.md): Full audio integration complete - December 2025
package audio
