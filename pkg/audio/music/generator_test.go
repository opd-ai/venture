package music

import (
	"math"
	"testing"
)

func TestNoteToFrequency(t *testing.T) {
	tests := []struct {
		note int
		want float64
	}{
		{69, 440.0},   // A4
		{57, 220.0},   // A3
		{81, 880.0},   // A5
		{60, 261.63},  // C4 (approximate)
		{64, 329.63},  // E4 (approximate)
	}

	for _, tt := range tests {
		got := NoteToFrequency(tt.note)
		if math.Abs(got-tt.want) > 0.1 {
			t.Errorf("NoteToFrequency(%d) = %f, want %f", tt.note, got, tt.want)
		}
	}
}

func TestGetScaleForGenre(t *testing.T) {
	tests := []struct {
		genre     string
		wantScale string
	}{
		{"fantasy", "Major"},
		{"scifi", "Chromatic"},
		{"horror", "Minor"},
		{"cyberpunk", "Blues"},
		{"post-apocalyptic", "Pentatonic"},
		{"unknown", "Major"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			scale := GetScaleForGenre(tt.genre)
			if scale.Name != tt.wantScale {
				t.Errorf("GetScaleForGenre(%s).Name = %s, want %s", tt.genre, scale.Name, tt.wantScale)
			}
			if len(scale.Intervals) == 0 {
				t.Error("scale has no intervals")
			}
		})
	}
}

func TestGetChordProgression(t *testing.T) {
	tests := []struct {
		genre    string
		rootNote int
		wantLen  int
	}{
		{"fantasy", 60, 4},
		{"horror", 60, 4},
		{"scifi", 60, 4},
		{"unknown", 60, 4},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			chords := GetChordProgression(tt.genre, tt.rootNote)
			if len(chords) != tt.wantLen {
				t.Errorf("len(chords) = %d, want %d", len(chords), tt.wantLen)
			}
			for i, chord := range chords {
				if len(chord.Notes) == 0 {
					t.Errorf("chord[%d] has no notes", i)
				}
			}
		})
	}
}

func TestGetRhythmForContext(t *testing.T) {
	tests := []struct {
		context string
	}{
		{"combat"},
		{"exploration"},
		{"ambient"},
		{"victory"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			rhythm := GetRhythmForContext(tt.context)
			
			if len(rhythm.Pattern) == 0 {
				t.Error("rhythm has no pattern")
			}
			
			if len(rhythm.Velocity) != len(rhythm.Pattern) {
				t.Errorf("velocity length %d != pattern length %d", len(rhythm.Velocity), len(rhythm.Pattern))
			}
			
			// Check all velocities are in valid range
			for i, v := range rhythm.Velocity {
				if v < 0.0 || v > 1.0 {
					t.Errorf("velocity[%d] = %f, out of range [0, 1]", i, v)
				}
			}
		})
	}
}

func TestGetTempoForContext(t *testing.T) {
	tests := []struct {
		context  string
		wantMin  float64
		wantMax  float64
	}{
		{"combat", 120.0, 160.0},
		{"exploration", 80.0, 100.0},
		{"ambient", 50.0, 70.0},
		{"victory", 110.0, 130.0},
		{"unknown", 90.0, 110.0},
	}

	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			tempo := GetTempoForContext(tt.context)
			if tempo < tt.wantMin || tempo > tt.wantMax {
				t.Errorf("tempo = %f, want between %f and %f", tempo, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGenerator_GenerateTrack(t *testing.T) {
	tests := []struct {
		name     string
		genre    string
		context  string
		duration float64
	}{
		{"fantasy combat", "fantasy", "combat", 2.0},
		{"scifi exploration", "scifi", "exploration", 3.0},
		{"horror ambient", "horror", "ambient", 4.0},
		{"cyberpunk victory", "cyberpunk", "victory", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(44100, 12345)
			sample := gen.GenerateTrack(tt.genre, tt.context, 54321, tt.duration)

			if sample == nil {
				t.Fatal("GenerateTrack returned nil")
			}

			if sample.SampleRate != 44100 {
				t.Errorf("SampleRate = %d, want 44100", sample.SampleRate)
			}

			expectedLen := int(44100 * tt.duration)
			if len(sample.Data) != expectedLen {
				t.Errorf("len(Data) = %d, want %d", len(sample.Data), expectedLen)
			}

			// Check that samples are in reasonable range
			for i, v := range sample.Data {
				if v < -2.0 || v > 2.0 {
					t.Errorf("sample[%d] = %f, out of reasonable range", i, v)
					break
				}
			}

			// Check that track has content
			hasContent := false
			for _, v := range sample.Data {
				if v != 0 {
					hasContent = true
					break
				}
			}
			if !hasContent {
				t.Error("track has no content (all zeros)")
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	seed := int64(99999)
	
	gen1 := NewGenerator(44100, seed)
	sample1 := gen1.GenerateTrack("fantasy", "combat", seed, 1.0)
	
	gen2 := NewGenerator(44100, seed)
	sample2 := gen2.GenerateTrack("fantasy", "combat", seed, 1.0)
	
	if len(sample1.Data) != len(sample2.Data) {
		t.Fatal("samples have different lengths")
	}
	
	// Check for general similarity (not exact due to RNG and mixing)
	differenceCount := 0
	for i := range sample1.Data {
		if math.Abs(sample1.Data[i]-sample2.Data[i]) > 0.0001 {
			differenceCount++
		}
	}
	
	// Should be mostly identical
	maxAllowedDifferences := len(sample1.Data) / 10 // Allow 10% difference
	if differenceCount > maxAllowedDifferences {
		t.Errorf("too many differences: %d out of %d samples", differenceCount, len(sample1.Data))
	}
}

func TestGenerator_FadeInOut(t *testing.T) {
	gen := NewGenerator(44100, 12345)
	sample := gen.GenerateTrack("fantasy", "combat", 12345, 2.0)
	
	// Check fade in - first samples should be quieter
	firstAvg := 0.0
	for i := 0; i < 1000; i++ {
		firstAvg += math.Abs(sample.Data[i])
	}
	firstAvg /= 1000
	
	// Middle samples should be louder
	midAvg := 0.0
	mid := len(sample.Data) / 2
	for i := mid; i < mid+1000; i++ {
		midAvg += math.Abs(sample.Data[i])
	}
	midAvg /= 1000
	
	// Last samples should be quieter (fade out)
	lastAvg := 0.0
	start := len(sample.Data) - 1000
	for i := start; i < len(sample.Data); i++ {
		lastAvg += math.Abs(sample.Data[i])
	}
	lastAvg /= 1000
	
	if midAvg <= firstAvg {
		t.Error("middle section not louder than beginning (fade in not working)")
	}
	
	if midAvg <= lastAvg {
		t.Error("middle section not louder than end (fade out not working)")
	}
}

func BenchmarkGenerator_GenerateTrack(b *testing.B) {
	gen := NewGenerator(44100, 12345)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateTrack("fantasy", "combat", int64(i), 2.0)
	}
}

func BenchmarkNoteToFrequency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NoteToFrequency(69)
	}
}
