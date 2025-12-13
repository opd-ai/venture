package audio

import (
	"testing"
)

func TestManager_NewManager(t *testing.T) {
	m := NewManager(44100, 12345)
	if m == nil {
		t.Fatal("expected non-nil Manager")
	}
	if m.sampleRate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", m.sampleRate)
	}
	if m.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", m.seed)
	}
	if m.masterVolume != 1.0 {
		t.Errorf("expected masterVolume 1.0, got %f", m.masterVolume)
	}
}

func TestManager_VolumeControls(t *testing.T) {
	m := NewManager(44100, 12345)

	tests := []struct {
		name         string
		setMaster    float64
		setMusic     float64
		setSFX       float64
		expectMaster float64
		expectMusic  float64
		expectSFX    float64
		musicEnabled bool
		sfxEnabled   bool
	}{
		{"normal", 0.8, 0.6, 0.7, 0.8, 0.6, 0.7, true, true},
		{"zero music", 1.0, 0.0, 1.0, 1.0, 0.0, 1.0, false, true},
		{"zero sfx", 1.0, 1.0, 0.0, 1.0, 1.0, 0.0, true, false},
		{"clamp high", 1.5, 2.0, 1.2, 1.0, 1.0, 1.0, true, true},
		{"clamp low", -0.5, -1.0, -0.2, 0.0, 0.0, 0.0, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.SetMasterVolume(tt.setMaster)
			m.SetMusicVolume(tt.setMusic)
			m.SetSFXVolume(tt.setSFX)

			if got := m.GetMasterVolume(); got != tt.expectMaster {
				t.Errorf("masterVolume: got %f, want %f", got, tt.expectMaster)
			}
			if got := m.GetMusicVolume(); got != tt.expectMusic {
				t.Errorf("musicVolume: got %f, want %f", got, tt.expectMusic)
			}
			if got := m.GetSFXVolume(); got != tt.expectSFX {
				t.Errorf("sfxVolume: got %f, want %f", got, tt.expectSFX)
			}
			if m.musicEnabled != tt.musicEnabled {
				t.Errorf("musicEnabled: got %v, want %v", m.musicEnabled, tt.musicEnabled)
			}
			if m.sfxEnabled != tt.sfxEnabled {
				t.Errorf("sfxEnabled: got %v, want %v", m.sfxEnabled, tt.sfxEnabled)
			}
		})
	}
}

// MockAdaptiveMusicSystem for testing.
type MockAdaptiveMusicSystem struct {
	setContextCalled      bool
	updateIntensityCalled bool
	addLayerCalled        bool
	removeLayerCalled     bool
	updateCalled          bool
	generateTrackCalled   bool
}

func (m *MockAdaptiveMusicSystem) SetContext(context MusicContext) error {
	m.setContextCalled = true
	return nil
}

func (m *MockAdaptiveMusicSystem) UpdateIntensity(intensity float64) error {
	m.updateIntensityCalled = true
	return nil
}

func (m *MockAdaptiveMusicSystem) AddLayer(layer MusicLayer) error {
	m.addLayerCalled = true
	return nil
}

func (m *MockAdaptiveMusicSystem) RemoveLayer(layer MusicLayer) error {
	m.removeLayerCalled = true
	return nil
}

func (m *MockAdaptiveMusicSystem) Update(deltaTime float64) {
	m.updateCalled = true
}

func (m *MockAdaptiveMusicSystem) GenerateTrack(duration float64) *AudioSample {
	m.generateTrackCalled = true
	return &AudioSample{
		SampleRate: 44100,
		Data:       []float64{0.1, 0.2, 0.3},
	}
}

// MockSFXGenerator for testing.
type MockSFXGenerator struct {
	generateCalled bool
}

func (m *MockSFXGenerator) Generate(effectType string, seed int64) *AudioSample {
	m.generateCalled = true
	return &AudioSample{
		SampleRate: 44100,
		Data:       []float64{0.5, 0.6, 0.7},
	}
}

func TestManager_SetManagers(t *testing.T) {
	m := NewManager(44100, 12345)
	mockMusic := &MockAdaptiveMusicSystem{}
	mockSFX := &MockSFXGenerator{}

	m.SetMusicManager(mockMusic)
	m.SetSFXManager(mockSFX)

	if m.GetMusicManager() != mockMusic {
		t.Error("music manager not set correctly")
	}
	if m.GetSFXManager() != mockSFX {
		t.Error("SFX manager not set correctly")
	}
}

func TestManager_MusicOperations(t *testing.T) {
	m := NewManager(44100, 12345)
	mockMusic := &MockAdaptiveMusicSystem{}
	m.SetMusicManager(mockMusic)

	// Test SetMusicContext
	err := m.SetMusicContext(MusicContext{Location: "dungeon"})
	if err != nil {
		t.Errorf("SetMusicContext failed: %v", err)
	}
	if !mockMusic.setContextCalled {
		t.Error("SetContext not called on music manager")
	}

	// Test UpdateMusicIntensity
	err = m.UpdateMusicIntensity(0.8)
	if err != nil {
		t.Errorf("UpdateMusicIntensity failed: %v", err)
	}
	if !mockMusic.updateIntensityCalled {
		t.Error("UpdateIntensity not called on music manager")
	}

	// Test AddMusicLayer
	err = m.AddMusicLayer(MusicLayerPercussion)
	if err != nil {
		t.Errorf("AddMusicLayer failed: %v", err)
	}
	if !mockMusic.addLayerCalled {
		t.Error("AddLayer not called on music manager")
	}

	// Test RemoveMusicLayer
	err = m.RemoveMusicLayer(MusicLayerPercussion)
	if err != nil {
		t.Errorf("RemoveMusicLayer failed: %v", err)
	}
	if !mockMusic.removeLayerCalled {
		t.Error("RemoveLayer not called on music manager")
	}

	// Test Update
	m.Update(0.016)
	if !mockMusic.updateCalled {
		t.Error("Update not called on music manager")
	}

	// Test GenerateMusicTrack
	sample := m.GenerateMusicTrack(1.0)
	if sample == nil {
		t.Error("expected non-nil sample")
	}
	if !mockMusic.generateTrackCalled {
		t.Error("GenerateTrack not called on music manager")
	}
}

func TestManager_SFXOperations(t *testing.T) {
	m := NewManager(44100, 12345)
	mockSFX := &MockSFXGenerator{}
	m.SetSFXManager(mockSFX)

	sample := m.GenerateSFX("impact", 999)
	if sample == nil {
		t.Error("expected non-nil sample")
	}
	if !mockSFX.generateCalled {
		t.Error("Generate not called on SFX manager")
	}
}

func TestManager_DisabledAudio(t *testing.T) {
	m := NewManager(44100, 12345)
	mockMusic := &MockAdaptiveMusicSystem{}
	mockSFX := &MockSFXGenerator{}
	m.SetMusicManager(mockMusic)
	m.SetSFXManager(mockSFX)

	// Disable music
	m.SetMusicVolume(0.0)
	m.SetMusicContext(MusicContext{})
	if mockMusic.setContextCalled {
		t.Error("SetContext called despite music being disabled")
	}

	// Disable SFX
	m.SetSFXVolume(0.0)
	m.GenerateSFX("impact", 999)
	if mockSFX.generateCalled {
		t.Error("Generate called despite SFX being disabled")
	}
}

func TestManager_VolumeApplication(t *testing.T) {
	m := NewManager(44100, 12345)
	mockMusic := &MockAdaptiveMusicSystem{}
	m.SetMusicManager(mockMusic)

	m.SetMasterVolume(0.5)
	m.SetMusicVolume(0.8)

	sample := m.GenerateMusicTrack(1.0)
	if sample == nil {
		t.Fatal("expected non-nil sample")
	}

	// Expected volume multiplier: 0.5 * 0.8 = 0.4
	// Original data: [0.1, 0.2, 0.3]
	// Expected: [0.04, 0.08, 0.12]
	expectedData := []float64{0.04, 0.08, 0.12}

	if len(sample.Data) != len(expectedData) {
		t.Fatalf("data length mismatch: got %d, want %d", len(sample.Data), len(expectedData))
	}

	const epsilon = 0.0001
	for i, expected := range expectedData {
		diff := sample.Data[i] - expected
		if diff < 0 {
			diff = -diff
		}
		if diff > epsilon {
			t.Errorf("data[%d]: got %f, want %f (diff %f > epsilon %f)", i, sample.Data[i], expected, diff, epsilon)
		}
	}
}
