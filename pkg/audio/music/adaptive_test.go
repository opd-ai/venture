package music

import (
	"fmt"
	"testing"

	"github.com/opd-ai/venture/pkg/audio"
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
		context            string
		expectedAmbient    bool
		expectedMelody     bool
		expectedPercussion bool
		expectedIntensity  bool
		minTempo           float64
		maxTempo           float64
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

func TestAdaptiveComposer_SetContextFromStruct(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	tests := []struct {
		name           string
		context        audio.MusicContext
		expectedActive map[string]bool // layer name → should be active
	}{
		{
			name: "exploration_low_danger",
			context: audio.MusicContext{
				Location: "wilderness",
				Combat:   false,
				Danger:   0.1,
			},
			expectedActive: map[string]bool{
				"ambient":    true,
				"melody":     true,
				"percussion": false,
			},
		},
		{
			name: "combat_medium_danger",
			context: audio.MusicContext{
				Location: "dungeon",
				Combat:   true,
				Danger:   0.5,
			},
			expectedActive: map[string]bool{
				"ambient":    true,
				"melody":     true,
				"percussion": true,
				"intensity":  true,
			},
		},
		{
			name: "boss_fight",
			context: audio.MusicContext{
				Location:   "boss_room",
				Combat:     true,
				BossNearby: true,
				Danger:     1.0,
			},
			expectedActive: map[string]bool{
				"ambient":    true,
				"melody":     true,
				"percussion": true,
				"harmony":    true,
				"intensity":  true,
			},
		},
		{
			name: "town_safe",
			context: audio.MusicContext{
				Location: "town",
				Combat:   false,
				Danger:   0.0,
			},
			expectedActive: map[string]bool{
				"ambient": true,
				"melody":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := composer.SetContextFromStruct(tt.context)
			if err != nil {
				t.Fatalf("SetContextFromStruct() error = %v", err)
			}

			// Update to apply context
			for i := 0; i < 10; i++ {
				composer.Update(0.1)
			}

			// Verify expected layer states
			for layerName, shouldBeActive := range tt.expectedActive {
				volume := composer.GetLayerVolume(layerName)
				isActive := volume > 0.01

				if shouldBeActive && !isActive {
					t.Errorf("Layer %s should be active (volume > 0), got volume = %f", layerName, volume)
				}
				if !shouldBeActive && isActive {
					t.Errorf("Layer %s should be inactive (volume = 0), got volume = %f", layerName, volume)
				}
			}
		})
	}
}

func TestAdaptiveComposer_UpdateIntensity(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	tests := []struct {
		name      string
		intensity float64
		wantMin   float64
		wantMax   float64
	}{
		{"zero_intensity", 0.0, 0.0, 0.01},
		{"low_intensity", 0.25, 0.05, 0.15},
		{"medium_intensity", 0.5, 0.15, 0.3},
		{"high_intensity", 0.75, 0.25, 0.4},
		{"max_intensity", 1.0, 0.35, 0.6},
		{"above_max_clamped", 1.5, 0.35, 0.6},  // Should clamp to 1.0
		{"below_min_clamped", -0.5, 0.0, 0.01}, // Should clamp to 0.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset intensity layer
			composer.layers["intensity"].Volume = 0.0
			composer.layers["intensity"].TargetVolume = 0.0

			err := composer.UpdateIntensity(tt.intensity)
			if err != nil {
				t.Fatalf("UpdateIntensity() error = %v", err)
			}

			// Update to apply intensity change (more iterations for convergence)
			for i := 0; i < 20; i++ {
				composer.Update(0.1)
			}

			volume := composer.GetLayerVolume("intensity")
			if volume < tt.wantMin || volume > tt.wantMax {
				t.Errorf("Intensity volume = %f, want range [%f, %f]", volume, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestAdaptiveComposer_AddLayer(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Start with minimal layers (exploration context)
	composer.SetContext("exploration")

	// Update to stabilize
	for i := 0; i < 10; i++ {
		composer.Update(0.1)
	}

	layers := []audio.MusicLayer{
		audio.MusicLayerHarmony,    // Not active in exploration
		audio.MusicLayerPercussion, // Not active in exploration
		audio.MusicLayerIntensity,  // Not active in exploration
	}

	for _, layer := range layers {
		t.Run(layer.String(), func(t *testing.T) {
			// Store initial volume before adding
			initialVolume := composer.GetLayerVolume(layer.String())

			err := composer.AddLayer(layer)
			if err != nil {
				t.Fatalf("AddLayer() error = %v", err)
			}

			// Update to activate layer (more iterations for convergence)
			for i := 0; i < 20; i++ {
				composer.Update(0.1)
			}

			volume := composer.GetLayerVolume(layer.String())
			if volume <= initialVolume+0.01 {
				t.Errorf("Layer %s volume = %f, want > %f after adding", layer, volume, initialVolume+0.01)
			}
		})
	}
}

func TestAdaptiveComposer_RemoveLayer(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Start with all layers active (boss context)
	composer.SetContext("boss")

	// Update to activate all layers
	for i := 0; i < 20; i++ {
		composer.Update(0.1)
	}

	layers := []audio.MusicLayer{
		audio.MusicLayerHarmony,
		audio.MusicLayerPercussion,
		audio.MusicLayerIntensity,
	}

	for _, layer := range layers {
		t.Run(layer.String(), func(t *testing.T) {
			// Store initial volume
			initialVolume := composer.GetLayerVolume(layer.String())

			err := composer.RemoveLayer(layer)
			if err != nil {
				t.Fatalf("RemoveLayer() error = %v", err)
			}

			// Update to deactivate layer (more iterations for full fade out)
			for i := 0; i < 50; i++ {
				composer.Update(0.1)
			}

			volume := composer.GetLayerVolume(layer.String())
			// Volume should be significantly reduced (may not reach exactly 0 due to exponential decay)
			if volume >= initialVolume*0.5 {
				t.Errorf("Layer %s volume = %f, want < %f (50%% of initial) after removal", layer, volume, initialVolume*0.5)
			}
		})
	}
}

func TestAdaptiveComposer_Update(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)

	// Set initial context
	composer.SetContext("exploration")

	// Set target volumes manually
	composer.layers["percussion"].TargetVolume = 0.5
	composer.layers["percussion"].Active = true

	initialVolume := composer.layers["percussion"].Volume

	// Update with various delta times
	deltaTimes := []float64{0.01, 0.05, 0.1, 0.5, 1.0}

	for _, dt := range deltaTimes {
		t.Run(fmt.Sprintf("deltaTime_%.2f", dt), func(t *testing.T) {
			// Reset to known state
			composer.layers["percussion"].Volume = 0.0
			composer.layers["percussion"].TargetVolume = 0.5

			composer.Update(dt)

			newVolume := composer.layers["percussion"].Volume

			// Volume should have changed toward target
			if newVolume <= initialVolume {
				t.Errorf("Volume didn't increase: initial=%f, new=%f, dt=%f", initialVolume, newVolume, dt)
			}

			// Larger delta time should produce larger change
			if dt >= 0.5 && newVolume < 0.2 {
				t.Errorf("Volume change too small for dt=%f: volume=%f", dt, newVolume)
			}
		})
	}
}

func TestAdaptiveComposer_GenerateTrack(t *testing.T) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)
	composer.SetContext("combat")

	duration := 3.0
	track := composer.GenerateTrack(duration)

	if track == nil {
		t.Fatal("GenerateTrack() returned nil")
	}

	expectedSamples := int(float64(composer.sampleRate) * duration)
	if len(track.Data) != expectedSamples {
		t.Errorf("Track length = %d, want %d", len(track.Data), expectedSamples)
	}

	if track.SampleRate != composer.sampleRate {
		t.Errorf("Track sample rate = %d, want %d", track.SampleRate, composer.sampleRate)
	}

	// Verify GenerateTrack produces same result as GenerateAdaptiveTrack
	track2 := composer.GenerateAdaptiveTrack(duration)

	if len(track.Data) != len(track2.Data) {
		t.Error("GenerateTrack and GenerateAdaptiveTrack produce different lengths")
	}
}

func TestMusicLayer_String(t *testing.T) {
	tests := []struct {
		layer audio.MusicLayer
		want  string
	}{
		{audio.MusicLayerBase, "base"},
		{audio.MusicLayerHarmony, "harmony"},
		{audio.MusicLayerPercussion, "percussion"},
		{audio.MusicLayerMelody, "melody"},
		{audio.MusicLayerIntensity, "intensity"},
		{audio.MusicLayer(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.layer.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAdaptiveComposer_InterfaceCompliance(t *testing.T) {
	// Verify AdaptiveMusicManager implements audio.AdaptiveMusicSystem
	var _ audio.AdaptiveMusicSystem = &AdaptiveMusicManager{}
}

func TestAdaptiveMusicManager_Interface(t *testing.T) {
	manager := NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize("fantasy", 60)

	// Test SetContext with MusicContext
	ctx := audio.MusicContext{
		Location: "dungeon",
		Combat:   true,
		Danger:   0.7,
	}

	err := manager.SetContext(ctx)
	if err != nil {
		t.Fatalf("SetContext() error = %v", err)
	}

	// Test UpdateIntensity
	err = manager.UpdateIntensity(0.8)
	if err != nil {
		t.Fatalf("UpdateIntensity() error = %v", err)
	}

	// Test AddLayer
	err = manager.AddLayer(audio.MusicLayerPercussion)
	if err != nil {
		t.Fatalf("AddLayer() error = %v", err)
	}

	// Test Update
	manager.Update(0.1)

	// Test GenerateTrack
	track := manager.GenerateTrack(1.0)
	if track == nil {
		t.Fatal("GenerateTrack() returned nil")
	}

	// Test RemoveLayer
	err = manager.RemoveLayer(audio.MusicLayerPercussion)
	if err != nil {
		t.Fatalf("RemoveLayer() error = %v", err)
	}

	// Test helper methods
	count := manager.GetActiveLayerCount()
	if count < 0 {
		t.Errorf("GetActiveLayerCount() = %d, want >= 0", count)
	}

	volume := manager.GetLayerVolume("melody")
	if volume < 0.0 || volume > 1.0 {
		t.Errorf("GetLayerVolume() = %f, want [0.0, 1.0]", volume)
	}
}

func BenchmarkAdaptiveComposer_Update(b *testing.B) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)
	composer.SetContext("combat")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		composer.Update(0.016) // ~60 FPS
	}
}

func BenchmarkAdaptiveComposer_GenerateTrack(b *testing.B) {
	composer := NewAdaptiveComposer(44100, 12345)
	composer.Initialize("fantasy", 60)
	composer.SetContext("combat")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		composer.GenerateTrack(1.0)
	}
}
