package engine

import (
	"fmt"
	"sync"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
	"github.com/opd-ai/venture/pkg/audio/sfx"
)

// AudioManager manages game audio including music and sound effects.
// It wraps the procedural audio generation systems and provides a unified
// interface for playing genre-aware music and context-sensitive sound effects.
type AudioManager struct {
	musicGen       *music.Generator
	sfxGen         *sfx.Generator
	currentTrack   *audio.AudioSample
	currentGenre   string
	currentContext string
	musicVolume    float64
	sfxVolume      float64
	seed           int64
	mu             sync.RWMutex
	
	// Track whether audio is enabled
	musicEnabled bool
	sfxEnabled   bool
}

// NewAudioManager creates a new audio manager with the specified sample rate and seed.
// The seed ensures deterministic audio generation for the same game world.
func NewAudioManager(sampleRate int, seed int64) *AudioManager {
	return &AudioManager{
		musicGen:     music.NewGenerator(sampleRate, seed),
		sfxGen:       sfx.NewGenerator(sampleRate, seed),
		musicVolume:  1.0,
		sfxVolume:    1.0,
		seed:         seed,
		musicEnabled: true,
		sfxEnabled:   true,
	}
}

// SetMusicVolume sets the music volume (0.0 to 1.0).
func (am *AudioManager) SetMusicVolume(volume float64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	am.musicVolume = volume
	if volume == 0.0 {
		am.musicEnabled = false
	} else {
		am.musicEnabled = true
	}
}

// SetSFXVolume sets the sound effects volume (0.0 to 1.0).
func (am *AudioManager) SetSFXVolume(volume float64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	am.sfxVolume = volume
	if volume == 0.0 {
		am.sfxEnabled = false
	} else {
		am.sfxEnabled = true
	}
}

// GetMusicVolume returns the current music volume.
func (am *AudioManager) GetMusicVolume() float64 {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.musicVolume
}

// GetSFXVolume returns the current sound effects volume.
func (am *AudioManager) GetSFXVolume() float64 {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.sfxVolume
}

// PlayMusic generates and starts playing background music for the specified genre and context.
// Genre examples: "fantasy", "scifi", "horror", "cyberpunk", "postapoc"
// Context examples: "combat", "exploration", "boss", "victory", "defeat"
func (am *AudioManager) PlayMusic(genre, context string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if !am.musicEnabled {
		return nil // Music disabled, silently succeed
	}
	
	// Don't regenerate if already playing the same music
	if am.currentGenre == genre && am.currentContext == context && am.currentTrack != nil {
		return nil
	}
	
	// Generate a new music track (30 seconds duration)
	track := am.musicGen.GenerateTrack(genre, context, am.seed, 30.0)
	
	// Apply volume scaling to the generated track
	scaledTrack := am.applyVolumeToTrack(track, am.musicVolume)
	
	am.currentTrack = scaledTrack
	am.currentGenre = genre
	am.currentContext = context
	
	return nil
}

// PlaySFX generates and plays a sound effect of the specified type.
// Effect types: "impact", "explosion", "magic", "laser", "pickup", "hit", "jump", "death", "powerup"
func (am *AudioManager) PlaySFX(effectType string, effectSeed int64) error {
	am.mu.RLock()
	enabled := am.sfxEnabled
	volume := am.sfxVolume
	am.mu.RUnlock()
	
	if !enabled {
		return nil // SFX disabled, silently succeed
	}
	
	// Generate the sound effect
	sample := am.sfxGen.Generate(effectType, effectSeed)
	
	// Apply volume scaling
	_ = am.applyVolumeToTrack(sample, volume)
	
	// In a real implementation, we would play the sample through an audio system
	// For now, we just generate it (Phase 8 focus is on integration, not audio playback)
	
	return nil
}

// GetCurrentTrack returns the currently playing music track (if any).
func (am *AudioManager) GetCurrentTrack() *audio.AudioSample {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.currentTrack
}

// GetCurrentMusicInfo returns the genre and context of the currently playing music.
func (am *AudioManager) GetCurrentMusicInfo() (genre, context string) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.currentGenre, am.currentContext
}

// StopMusic stops the currently playing music.
func (am *AudioManager) StopMusic() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.currentTrack = nil
	am.currentGenre = ""
	am.currentContext = ""
}

// applyVolumeToTrack scales all samples in the track by the given volume multiplier.
func (am *AudioManager) applyVolumeToTrack(track *audio.AudioSample, volume float64) *audio.AudioSample {
	scaledData := make([]float64, len(track.Data))
	for i, sample := range track.Data {
		scaledData[i] = sample * volume
	}
	return &audio.AudioSample{
		SampleRate: track.SampleRate,
		Data:       scaledData,
	}
}

// AudioManagerSystem is an ECS system that updates audio state based on game context.
type AudioManagerSystem struct {
	audioManager *AudioManager
	lastGenre    string
	lastContext  string
	updateTimer  int
}

// NewAudioManagerSystem creates a new audio manager system.
func NewAudioManagerSystem(audioManager *AudioManager) *AudioManagerSystem {
	return &AudioManagerSystem{
		audioManager: audioManager,
		updateTimer:  0,
	}
}

// Update checks game state and updates audio context as needed.
// This runs every frame to detect context changes (e.g., entering/leaving combat).
func (ams *AudioManagerSystem) Update(entities []*Entity, deltaTime float64) {
	ams.updateTimer++
	
	// Check for context changes every 60 frames (1 second at 60 FPS)
	if ams.updateTimer < 60 {
		return
	}
	ams.updateTimer = 0
	
	// Determine current game context by checking for enemies
	context := "exploration"
	enemyCount := 0
	
	for _, entity := range entities {
		// Skip player entities
		if entity.HasComponent("input") {
			continue
		}
		
		// Count enemies (have health but not input = enemy)
		if entity.HasComponent("health") {
			enemyCount++
		}
	}
	
	// Switch to combat music if enemies present
	if enemyCount > 0 {
		context = "combat"
	}
	
	// Check for boss enemies (high attack power)
	for _, entity := range entities {
		if entity.HasComponent("stats") {
			statsComp, ok := entity.GetComponent("stats")
			if !ok {
				continue
			}
			stats := statsComp.(*StatsComponent)
			if stats.Attack > 20 {
				context = "boss"
				break
			}
		}
	}
	
	// Update music if context changed
	if context != ams.lastContext {
		// Get genre from world settings (default to fantasy)
		genre := "fantasy"
		// In a full implementation, we'd get this from world state
		
		err := ams.audioManager.PlayMusic(genre, context)
		if err != nil {
			// Log error but don't crash
			fmt.Printf("Warning: Failed to update music: %v\n", err)
		}
		
		ams.lastContext = context
		ams.lastGenre = genre
	}
}
