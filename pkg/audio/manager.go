package audio

import (
	"sync"
)

// Manager is a unified audio manager that coordinates music, SFX, and voice systems.
// It provides a single interface for managing all game audio with support
// for adaptive music, varied sound effects, and voice chat.
type Manager struct {
	sampleRate int
	seed       int64
	mu         sync.RWMutex

	// Sub-managers (set via dependency injection)
	musicManager   AdaptiveMusicSystem
	sfxManager     SFXGenerator
	voiceCodec     VoiceCodec
	voiceProcessor *VoiceProcessor

	// Volume controls
	masterVolume float64
	musicVolume  float64
	sfxVolume    float64
	voiceVolume  float64

	// Enabled states
	musicEnabled bool
	sfxEnabled   bool
	voiceEnabled bool
}

// NewManager creates a new unified audio manager.
// sampleRate must be positive (typically 44100 or 48000).
func NewManager(sampleRate int, seed int64) *Manager {
	if sampleRate <= 0 {
		sampleRate = 44100 // Default to standard CD quality
	}
	return &Manager{
		sampleRate:   sampleRate,
		seed:         seed,
		masterVolume: 1.0,
		musicVolume:  1.0,
		sfxVolume:    1.0,
		voiceVolume:  1.0,
		musicEnabled: true,
		sfxEnabled:   true,
		voiceEnabled: true,
	}
}

// SetMusicManager sets the adaptive music manager implementation.
func (m *Manager) SetMusicManager(manager AdaptiveMusicSystem) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.musicManager = manager
}

// SetSFXManager sets the sound effects generator implementation.
func (m *Manager) SetSFXManager(manager SFXGenerator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sfxManager = manager
}

// GetMusicManager returns the current music manager.
func (m *Manager) GetMusicManager() AdaptiveMusicSystem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.musicManager
}

// GetSFXManager returns the current SFX manager.
func (m *Manager) GetSFXManager() SFXGenerator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sfxManager
}

// SetMasterVolume sets the master volume (0.0-1.0).
func (m *Manager) SetMasterVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.masterVolume = clampVolume(volume)
}

// SetMusicVolume sets the music volume (0.0-1.0).
// Note: Setting volume to 0.0 implicitly disables music playback.
// This is intentional to prevent silent music generation overhead.
// Use GetMusicVolume() to check current volume.
func (m *Manager) SetMusicVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.musicVolume = clampVolume(volume)
	// Intentionally disable subsystem when volume is 0.0 to avoid
	// CPU overhead of generating silent audio samples
	m.musicEnabled = volume > 0.0
}

// SetSFXVolume sets the sound effects volume (0.0-1.0).
// Note: Setting volume to 0.0 implicitly disables SFX playback.
// This is intentional to prevent silent sound generation overhead.
// Use GetSFXVolume() to check current volume.
func (m *Manager) SetSFXVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sfxVolume = clampVolume(volume)
	// Intentionally disable subsystem when volume is 0.0 to avoid
	// CPU overhead of generating silent audio samples
	m.sfxEnabled = volume > 0.0
}

// GetMasterVolume returns the master volume.
func (m *Manager) GetMasterVolume() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.masterVolume
}

// GetMusicVolume returns the music volume.
func (m *Manager) GetMusicVolume() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.musicVolume
}

// GetSFXVolume returns the sound effects volume.
func (m *Manager) GetSFXVolume() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sfxVolume
}

// SetMusicContext updates music based on gameplay context.
func (m *Manager) SetMusicContext(context MusicContext) error {
	m.mu.RLock()
	manager := m.musicManager
	enabled := m.musicEnabled
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return nil
	}

	return manager.SetContext(context)
}

// UpdateMusicIntensity adjusts music intensity (0.0-1.0).
func (m *Manager) UpdateMusicIntensity(intensity float64) error {
	m.mu.RLock()
	manager := m.musicManager
	enabled := m.musicEnabled
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return nil
	}

	return manager.UpdateIntensity(intensity)
}

// AddMusicLayer activates a music layer.
func (m *Manager) AddMusicLayer(layer MusicLayer) error {
	m.mu.RLock()
	manager := m.musicManager
	enabled := m.musicEnabled
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return nil
	}

	return manager.AddLayer(layer)
}

// RemoveMusicLayer deactivates a music layer.
func (m *Manager) RemoveMusicLayer(layer MusicLayer) error {
	m.mu.RLock()
	manager := m.musicManager
	enabled := m.musicEnabled
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return nil
	}

	return manager.RemoveLayer(layer)
}

// Update performs smooth transitions for music.
func (m *Manager) Update(deltaTime float64) {
	m.mu.RLock()
	manager := m.musicManager
	enabled := m.musicEnabled
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return
	}

	manager.Update(deltaTime)
}

// GenerateMusicTrack creates a music track with current settings.
func (m *Manager) GenerateMusicTrack(duration float64) *AudioSample {
	m.mu.RLock()
	manager := m.musicManager
	enabled := m.musicEnabled
	musicVol := m.musicVolume
	masterVol := m.masterVolume
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return nil
	}

	sample := manager.GenerateTrack(duration)
	if sample != nil {
		applyVolume(sample, musicVol*masterVol)
	}
	return sample
}

// GenerateSFX creates a sound effect.
func (m *Manager) GenerateSFX(effectType string, seed int64) *AudioSample {
	m.mu.RLock()
	manager := m.sfxManager
	enabled := m.sfxEnabled
	sfxVol := m.sfxVolume
	masterVol := m.masterVolume
	m.mu.RUnlock()

	if !enabled || manager == nil {
		return nil
	}

	sample := manager.Generate(effectType, seed)
	if sample != nil {
		applyVolume(sample, sfxVol*masterVol)
	}
	return sample
}

// GetSampleRate returns the sample rate.
func (m *Manager) GetSampleRate() int {
	return m.sampleRate
}

// GetSeed returns the seed.
func (m *Manager) GetSeed() int64 {
	return m.seed
}

// SetVoiceCodec sets the voice codec implementation.
func (m *Manager) SetVoiceCodec(codec VoiceCodec) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.voiceCodec = codec
}

// GetVoiceCodec returns the current voice codec.
func (m *Manager) GetVoiceCodec() VoiceCodec {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.voiceCodec
}

// SetVoiceProcessor sets the voice processor implementation.
func (m *Manager) SetVoiceProcessor(processor *VoiceProcessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.voiceProcessor = processor
}

// GetVoiceProcessor returns the current voice processor.
func (m *Manager) GetVoiceProcessor() *VoiceProcessor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.voiceProcessor
}

// SetVoiceVolume sets the voice chat volume (0.0-1.0).
func (m *Manager) SetVoiceVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.voiceVolume = clampVolume(volume)
	m.voiceEnabled = volume > 0.0
}

// GetVoiceVolume returns the voice chat volume.
func (m *Manager) GetVoiceVolume() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.voiceVolume
}

// IsVoiceEnabled returns whether voice chat is enabled.
func (m *Manager) IsVoiceEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.voiceEnabled && m.voiceCodec != nil && m.voiceProcessor != nil
}

// InitializeVoice sets up voice codec with the specified quality.
// The transport parameter may be nil for testing or when voice chat
// is not yet integrated with the network layer. Production deployments
// should provide a concrete VoiceTransport implementation that integrates
// with pkg/network/chat for proximity/guild/party voice channels.
//
// Server-side voice broadcast is handled by TCPServer.routeVoiceCommand in
// pkg/network/server.go, which fans out VoicePackets (serialised as
// StateUpdate.ComponentData with Type="_voice") to all connected peers.
// Client-side demultiplexing is performed by TCPVoiceTransport.HandleReceivedPacket
// in pkg/network/voice_transport.go. See cmd/client/handlers.go (initializeVoiceTransport)
// for the wiring call site.
func (m *Manager) InitializeVoice(quality VoiceQuality, transport VoiceTransport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	codec := NewSimpleVoiceCodec(m.sampleRate, quality)
	processor := NewVoiceProcessor(codec, transport)

	m.voiceCodec = codec
	m.voiceProcessor = processor
	m.voiceEnabled = true

	return nil
}

// clampVolume ensures volume is in valid range [0.0, 1.0].
func clampVolume(volume float64) float64 {
	if volume < 0.0 {
		return 0.0
	}
	if volume > 1.0 {
		return 1.0
	}
	return volume
}

// applyVolume scales audio sample data by volume multiplier.
func applyVolume(sample *AudioSample, volume float64) {
	if sample == nil || volume == 1.0 {
		return
	}

	for i := range sample.Data {
		sample.Data[i] *= volume

		// Clamp to prevent clipping
		if sample.Data[i] > 1.0 {
			sample.Data[i] = 1.0
		} else if sample.Data[i] < -1.0 {
			sample.Data[i] = -1.0
		}
	}
}
