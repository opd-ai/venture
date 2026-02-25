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
	tests := []struct {
		name      string
		genre     string
		seed      int64
		context   string
		wantScale string
	}{
		{
			name:      "scifi genre maps to chromatic scale",
			genre:     "scifi",
			seed:      54321,
			context:   "combat",
			wantScale: "Chromatic",
		},
		{
			name:      "postapoc genre maps to pentatonic scale",
			genre:     "postapoc",
			seed:      54322,
			context:   "exploration",
			wantScale: "Pentatonic",
		},
		{
			name:      "fantasy genre maps to major scale",
			genre:     "fantasy",
			seed:      54323,
			context:   "exploration",
			wantScale: "Major",
		},
		{
			name:      "horror genre maps to minor scale",
			genre:     "horror",
			seed:      54324,
			context:   "combat",
			wantScale: "Minor",
		},
		{
			name:      "cyberpunk genre maps to blues scale",
			genre:     "cyberpunk",
			seed:      54325,
			context:   "exploration",
			wantScale: "Blues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify scale mapping
			scale := GetScaleForGenre(tt.genre)
			if scale.Name != tt.wantScale {
				t.Errorf("GetScaleForGenre(%s).Name = %s, want %s",
					tt.genre, scale.Name, tt.wantScale)
			}

			// Create composer and verify track generation succeeds
			composer := NewAdaptiveComposer(44100, tt.seed)
			composer.Initialize(tt.genre, 60)
			composer.SetContext(tt.context)

			track := composer.GenerateAdaptiveTrack(1.0)
			if track == nil {
				t.Errorf("GenerateAdaptiveTrack() returned nil for %s genre", tt.genre)
				return
			}
			if track.SampleRate != 44100 {
				t.Errorf("Track sample rate = %d, want 44100", track.SampleRate)
			}
		})
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
