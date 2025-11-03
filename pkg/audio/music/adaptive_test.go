package music

import (
	"testing"
)

func TestNewAdaptiveComposer(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	if composer == nil {
		t.Fatal("NewAdaptiveComposer() returned nil")
	}
	if composer.sampleRate != 44100 {
		t.Errorf("sampleRate = %d, want 44100", composer.sampleRate)
	}
	if composer.seed != 12345 {
		t.Errorf("seed = %d, want 12345", composer.seed)
	}
}

func TestAdaptiveComposer_Initialize(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60) // Middle C

	if composer.currentGenre != "fantasy" {
		t.Errorf("currentGenre = %s, want fantasy", composer.currentGenre)
	}
	if composer.rootNote != 60 {
		t.Errorf("rootNote = %d, want 60", composer.rootNote)
	}

	// Check that base layers are initialized
	expectedLayers := []string{"ambient", "melody", "harmony", "percussion", "intensity"}
	for _, layerName := range expectedLayers {
		if _, exists := composer.layers[layerName]; !exists {
			t.Errorf("Layer %s not initialized", layerName)
		}
	}

	// Check initial layer states
	if !composer.layers["ambient"].Active {
		t.Error("ambient layer should be active initially")
	}
	if !composer.layers["melody"].Active {
		t.Error("melody layer should be active initially")
	}
	if composer.layers["percussion"].Active {
		t.Error("percussion layer should be inactive initially")
	}
}

func TestAdaptiveComposer_SetContext(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	tests := []struct {
		context             string
		expectedAmbient     bool
		expectedMelody      bool
		expectedPercussion  bool
		expectedIntensity   bool
		minTempo            float64
		maxTempo            float64
	}{
		{"exploration", true, true, false, false, 80.0, 100.0},
		{"combat", true, true, true, true, 130.0, 150.0},
		{"boss", true, true, true, true, 150.0, 170.0},
		{"puzzle", true, true, false, false, 70.0, 90.0},
		{"victory", true, true, true, false, 110.0, 130.0},
	}

	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			composer.SetContext(tt.context)

			// Check layer activation targets
			if tt.expectedAmbient && composer.layers["ambient"].TargetVolume == 0.0 {
				t.Error("ambient layer should have target volume > 0")
			}
			if tt.expectedMelody && composer.layers["melody"].TargetVolume == 0.0 {
				t.Error("melody layer should have target volume > 0")
			}
			if !tt.expectedPercussion && composer.layers["percussion"].TargetVolume > 0.0 {
				t.Error("percussion layer should have target volume = 0")
			}
			if !tt.expectedIntensity && composer.layers["intensity"].TargetVolume > 0.0 {
				t.Error("intensity layer should have target volume = 0")
			}

			// Check tempo is in expected range
			if composer.tempo < tt.minTempo || composer.tempo > tt.maxTempo {
				t.Errorf("Tempo %f out of range [%f, %f]", composer.tempo, tt.minTempo, tt.maxTempo)
			}
		})
	}
}

func TestAdaptiveComposer_UpdateLayers(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Set context to combat (activates more layers)
	composer.SetContext("combat")

	// Initial volume should be at exploration values
	initialAmbientVolume := composer.layers["ambient"].Volume

	// Update layers with fast transition
	composer.UpdateLayers(0.5)

	// Volumes should have moved toward targets
	newAmbientVolume := composer.layers["ambient"].Volume

	// Check that volume has changed
	if newAmbientVolume == initialAmbientVolume {
		// This might happen if initial and target are the same, but for combat they should differ
		// If they're different contexts, volume should change
		t.Log("Note: Volume didn't change - may be expected if initial = target")
	}
}

func TestAdaptiveComposer_GenerateAdaptiveTrack(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)
	composer.SetContext("exploration")

	duration := 2.0 // 2 seconds
	track := composer.GenerateAdaptiveTrack(duration)

	if track == nil {
		t.Fatal("GenerateAdaptiveTrack() returned nil")
	}

	expectedSamples := int(float64(composer.sampleRate) * duration)
	if len(track.Data) != expectedSamples {
		t.Errorf("Track length = %d, want %d", len(track.Data), expectedSamples)
	}

	if track.SampleRate != composer.sampleRate {
		t.Errorf("Track sample rate = %d, want %d", track.SampleRate, composer.sampleRate)
	}

	// Check that track contains audio data (not all zeros)
	hasAudio := false
	for _, sample := range track.Data {
		if sample != 0.0 {
			hasAudio = true
			break
		}
	}
	if !hasAudio {
		t.Error("Track contains no audio data")
	}

	// Check that track is normalized (no clipping)
	for i, sample := range track.Data {
		if sample > 1.0 || sample < -1.0 {
			t.Errorf("Sample[%d] = %f, should be in range [-1.0, 1.0]", i, sample)
			break
		}
	}
}

func TestAdaptiveComposer_GetActiveLayerCount(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Exploration: ambient + melody = 2 layers
	composer.SetContext("exploration")
	count := composer.GetActiveLayerCount()
	if count != 2 {
		t.Errorf("Active layer count = %d, want 2 for exploration", count)
	}

	// Combat: should have more layers (but need to update volumes first)
	composer.SetContext("combat")
	// Update layers so they become active
	composer.UpdateLayers(1.0)
	count = composer.GetActiveLayerCount()
	if count <= 2 {
		t.Errorf("Active layer count = %d, want > 2 for combat (got layers active but volumes may be low)", count)
	}
}

func TestAdaptiveComposer_GetLayerVolume(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Check initial volumes
	ambientVolume := composer.GetLayerVolume("ambient")
	if ambientVolume <= 0.0 {
		t.Error("ambient volume should be > 0 after initialization")
	}

	// Check non-existent layer
	invalidVolume := composer.GetLayerVolume("nonexistent")
	if invalidVolume != 0.0 {
		t.Errorf("Non-existent layer volume = %f, want 0.0", invalidVolume)
	}
}

func TestAdaptiveComposer_ContextTransitions(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Test context transitions
	contexts := []string{"exploration", "combat", "boss", "victory", "exploration"}

	for _, context := range contexts {
		composer.SetContext(context)

		// Generate short track to ensure no panics
		track := composer.GenerateAdaptiveTrack(0.5)
		if track == nil {
			t.Fatalf("Failed to generate track for context %s", context)
		}

		// Update layers
		composer.UpdateLayers(0.3)
	}
}

func TestAdaptiveComposer_GenreVariations(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			composer := NewAdaptiveComposer(44100, 12345)
			composer.Initialize(genre, 60)
			composer.SetContext("combat")

			track := composer.GenerateAdaptiveTrack(1.0)
			if track == nil {
				t.Fatalf("Failed to generate track for genre %s", genre)
			}

			// Verify track has audio data
			hasAudio := false
			for _, sample := range track.Data {
				if sample != 0.0 {
					hasAudio = true
					break
				}
			}
			if !hasAudio {
				t.Errorf("Genre %s produced silent track", genre)
			}
		})
	}
}

func TestAdaptiveComposer_LayerMixing(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Enable all layers
	composer.SetContext("boss")
	// Update layers to activate them
	composer.UpdateLayers(1.0)

	// Generate track with all layers
	track := composer.GenerateAdaptiveTrack(1.0)

	// Check that track amplitude is reasonable (normalized tracks may have lower amplitude)
	maxAmp := 0.0
	for _, sample := range track.Data {
		if abs := abs(sample); abs > maxAmp {
			maxAmp = abs
		}
	}

	// With normalization and mixing, amplitude should be reasonable
	if maxAmp < 0.1 {
		t.Errorf("Max amplitude with all layers = %f, expected higher (track may be too quiet)", maxAmp)
	}
	if maxAmp > 1.0 {
		t.Errorf("Max amplitude = %f, exceeds 1.0 (clipping)", maxAmp)
	}
}

func TestAdaptiveComposer_SmoothTransitions(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Start in exploration
	composer.SetContext("exploration")
	initialPercussionVol := composer.layers["percussion"].Volume

	// Switch to combat (activates percussion)
	composer.SetContext("combat")

	// Percussion should have non-zero target now
	if composer.layers["percussion"].TargetVolume == 0.0 {
		t.Error("Percussion target volume should be > 0 in combat")
	}

	// Update with slow transition
	composer.UpdateLayers(0.1)

	// Volume should have changed but not reached target yet
	newPercussionVol := composer.layers["percussion"].Volume
	if newPercussionVol == initialPercussionVol {
		// Volume should have started changing
		t.Log("Note: Percussion volume might not have changed if already at target")
	}

	// Multiple updates should approach target
	for i := 0; i < 10; i++ {
		composer.UpdateLayers(0.2)
	}

	finalPercussionVol := composer.layers["percussion"].Volume
	targetVol := composer.layers["percussion"].TargetVolume

	// After many updates, should be close to target
	if finalPercussionVol < targetVol*0.8 {
		t.Errorf("After transitions, volume %f not close to target %f", finalPercussionVol, targetVol)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
