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

// TestScaleIntervals verifies that scale intervals match expected music theory values.
// This test ensures scales are correctly defined according to music theory standards.
func TestScaleIntervals(t *testing.T) {
	tests := []struct {
		name              string
		scale             Scale
		expectedIntervals []int
		description       string
	}{
		{
			name:              "Major (Ionian mode)",
			scale:             ScaleMajor,
			expectedIntervals: []int{0, 2, 4, 5, 7, 9, 11},
			description:       "Whole-Whole-Half-Whole-Whole-Whole-Half pattern",
		},
		{
			name:              "Minor (Aeolian mode)",
			scale:             ScaleMinor,
			expectedIntervals: []int{0, 2, 3, 5, 7, 8, 10},
			description:       "Natural minor scale",
		},
		{
			name:              "Pentatonic",
			scale:             ScalePentatonic,
			expectedIntervals: []int{0, 2, 4, 7, 9},
			description:       "Five-note scale (major pentatonic)",
		},
		{
			name:              "Blues",
			scale:             ScaleBlues,
			expectedIntervals: []int{0, 3, 5, 6, 7, 10},
			description:       "Blues scale with flatted third and fifth",
		},
		{
			name:              "Chromatic",
			scale:             ScaleChromatic,
			expectedIntervals: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			description:       "All twelve semitones",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify scale name matches
			if tt.scale.Name == "" {
				t.Error("Scale name is empty")
			}

			// Verify interval count
			if len(tt.scale.Intervals) != len(tt.expectedIntervals) {
				t.Errorf("Scale %s has %d intervals, want %d",
					tt.scale.Name, len(tt.scale.Intervals), len(tt.expectedIntervals))
				return
			}

			// Verify each interval matches expected value
			for i, interval := range tt.scale.Intervals {
				if interval != tt.expectedIntervals[i] {
					t.Errorf("Scale %s interval[%d] = %d, want %d",
						tt.scale.Name, i, interval, tt.expectedIntervals[i])
				}
			}

			// Verify intervals are in ascending order
			for i := 1; i < len(tt.scale.Intervals); i++ {
				if tt.scale.Intervals[i] <= tt.scale.Intervals[i-1] {
					t.Errorf("Scale %s intervals not in ascending order at index %d: %d <= %d",
						tt.scale.Name, i, tt.scale.Intervals[i], tt.scale.Intervals[i-1])
				}
			}

			// Verify first interval is always 0 (root note)
			if len(tt.scale.Intervals) > 0 && tt.scale.Intervals[0] != 0 {
				t.Errorf("Scale %s first interval = %d, want 0 (root note)",
					tt.scale.Name, tt.scale.Intervals[0])
			}

			// Verify all intervals are within one octave (0-11)
			for i, interval := range tt.scale.Intervals {
				if interval < 0 || interval > 11 {
					t.Errorf("Scale %s interval[%d] = %d, out of range [0, 11]",
						tt.scale.Name, i, interval)
				}
			}
		})
	}
}

// TestGenreScaleIntervals verifies that genre-specific scales have correct intervals.
// This is a regression test to ensure GetScaleForGenre returns scales with valid intervals.
func TestGenreScaleIntervals(t *testing.T) {
	tests := []struct {
		genre             string
		expectedScale     string
		expectedIntervals []int
	}{
		{
			genre:             "fantasy",
			expectedScale:     "Major",
			expectedIntervals: []int{0, 2, 4, 5, 7, 9, 11},
		},
		{
			genre:             "scifi",
			expectedScale:     "Chromatic",
			expectedIntervals: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		{
			genre:             "horror",
			expectedScale:     "Minor",
			expectedIntervals: []int{0, 2, 3, 5, 7, 8, 10},
		},
		{
			genre:             "cyberpunk",
			expectedScale:     "Blues",
			expectedIntervals: []int{0, 3, 5, 6, 7, 10},
		},
		{
			genre:             "postapoc",
			expectedScale:     "Pentatonic",
			expectedIntervals: []int{0, 2, 4, 7, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			scale := GetScaleForGenre(tt.genre)

			// Verify scale name
			if scale.Name != tt.expectedScale {
				t.Errorf("GetScaleForGenre(%s).Name = %s, want %s",
					tt.genre, scale.Name, tt.expectedScale)
			}

			// Verify interval count
			if len(scale.Intervals) != len(tt.expectedIntervals) {
				t.Errorf("Genre %s scale has %d intervals, want %d",
					tt.genre, len(scale.Intervals), len(tt.expectedIntervals))
				return
			}

			// Verify each interval matches expected music theory value
			for i, interval := range scale.Intervals {
				if interval != tt.expectedIntervals[i] {
					t.Errorf("Genre %s scale interval[%d] = %d, want %d (music theory standard)",
						tt.genre, i, interval, tt.expectedIntervals[i])
				}
			}
		})
	}
}

// TestScaleIntervalSteps verifies that scales follow expected semitone step patterns.
func TestScaleIntervalSteps(t *testing.T) {
	tests := []struct {
		name          string
		scale         Scale
		expectedSteps []int // steps between consecutive notes
	}{
		{
			name:          "Major scale steps",
			scale:         ScaleMajor,
			expectedSteps: []int{2, 2, 1, 2, 2, 2}, // W-W-H-W-W-W (last H is to octave)
		},
		{
			name:          "Minor scale steps",
			scale:         ScaleMinor,
			expectedSteps: []int{2, 1, 2, 2, 1, 2}, // W-H-W-W-H-W (last W is to octave)
		},
		{
			name:          "Pentatonic scale steps",
			scale:         ScalePentatonic,
			expectedSteps: []int{2, 2, 3, 2}, // W-W-m3-W (last m3 is to octave)
		},
		{
			name:          "Blues scale steps",
			scale:         ScaleBlues,
			expectedSteps: []int{3, 2, 1, 1, 3}, // m3-W-H-H-m3 (last W is to octave)
		},
		{
			name:          "Chromatic scale steps",
			scale:         ScaleChromatic,
			expectedSteps: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, // all half steps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate steps between consecutive intervals
			steps := make([]int, 0, len(tt.scale.Intervals)-1)
			for i := 1; i < len(tt.scale.Intervals); i++ {
				step := tt.scale.Intervals[i] - tt.scale.Intervals[i-1]
				steps = append(steps, step)
			}

			// Verify step count
			if len(steps) != len(tt.expectedSteps) {
				t.Errorf("Scale %s has %d steps, want %d",
					tt.scale.Name, len(steps), len(tt.expectedSteps))
				return
			}

			// Verify each step matches expected pattern
			for i, step := range steps {
				if step != tt.expectedSteps[i] {
					t.Errorf("Scale %s step[%d] = %d semitones, want %d (music theory pattern)",
						tt.scale.Name, i, step, tt.expectedSteps[i])
				}
			}
		})
	}
}
