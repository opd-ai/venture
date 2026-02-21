// genre_consistency_test.go tests that genre naming is consistent across all music
// subsystems (scale selection, waveform assignment, adaptive composition).
//
// These tests prevent regression of a bug where "scifi" and "sci-fi" were used
// inconsistently across different music components, causing fallback behavior
// instead of genre-specific customization.
package music

import (
	"testing"

	"github.com/opd-ai/venture/pkg/audio"
)

// TestGenreConsistency verifies that genre names are consistent across all music subsystems.
// This test was added to prevent regression of the genre naming inconsistency bug (AUDIT.md Priority 1).
func TestGenreConsistency(t *testing.T) {
	tests := []struct {
		genre         string
		wantScale     string
		wantWaveforms []audio.WaveformType
	}{
		{
			genre:         "scifi",
			wantScale:     "Chromatic",
			wantWaveforms: []audio.WaveformType{audio.WaveformSquare, audio.WaveformSawtooth},
		},
		{
			genre:         "postapoc",
			wantScale:     "Pentatonic",
			wantWaveforms: []audio.WaveformType{audio.WaveformSawtooth},
		},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			// Test scale mapping
			scale := GetScaleForGenre(tt.genre)
			if scale.Name != tt.wantScale {
				t.Errorf("GetScaleForGenre(%s).Name = %s, want %s",
					tt.genre, scale.Name, tt.wantScale)
			}

			// Test motif waveform mapping
			gen := NewMotifGenerator(44100, 12345)
			motif := gen.GenerateMotif("test_entity", tt.genre, MotifTypeCharacter)

			// Verify waveform is in expected set
			validWaveform := false
			for _, expectedWaveform := range tt.wantWaveforms {
				if motif.Waveform == expectedWaveform {
					validWaveform = true
					break
				}
			}
			if !validWaveform {
				t.Errorf("GenerateMotif(%s) waveform = %v, want one of %v",
					tt.genre, motif.Waveform, tt.wantWaveforms)
			}
		})
	}
}

// TestGenreNamingCompatibility ensures short-form genre names work across all music systems.
// This is a regression test for the bug where "scifi" worked in theory.go but "sci-fi" was
// checked in adaptive.go, causing fallback behavior.
func TestGenreNamingCompatibility(t *testing.T) {
	// Create composer with short-form genre
	composer := NewAdaptiveComposer(44100, 54321)
	composer.Initialize("scifi", 60)

	// Verify it gets chromatic scale (not default major scale)
	scale := GetScaleForGenre("scifi")
	if scale.Name != "Chromatic" {
		t.Errorf("scifi genre should map to Chromatic scale, got %s", scale.Name)
	}

	// Generate a track to ensure no runtime errors with scifi genre
	composer.SetContext("combat")
	track := composer.GenerateAdaptiveTrack(1.0)
	if track == nil {
		t.Error("GenerateAdaptiveTrack() returned nil for scifi genre")
	}
	if track.SampleRate != 44100 {
		t.Errorf("Track sample rate = %d, want 44100", track.SampleRate)
	}

	// Test postapoc genre as well
	composer2 := NewAdaptiveComposer(44100, 54322)
	composer2.Initialize("postapoc", 60)

	scale2 := GetScaleForGenre("postapoc")
	if scale2.Name != "Pentatonic" {
		t.Errorf("postapoc genre should map to Pentatonic scale, got %s", scale2.Name)
	}

	composer2.SetContext("exploration")
	track2 := composer2.GenerateAdaptiveTrack(1.0)
	if track2 == nil {
		t.Error("GenerateAdaptiveTrack() returned nil for postapoc genre")
	}
}

// BenchmarkGenreConsistency benchmarks the performance of genre lookups across systems.
func BenchmarkGenreConsistency(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetScaleForGenre("scifi")
		_ = GetScaleForGenre("postapoc")
	}
}
