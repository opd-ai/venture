package music

import (
	"testing"

	"github.com/opd-ai/venture/pkg/audio"
)

func TestMotifType_String(t *testing.T) {
	tests := []struct {
		name      string
		motifType MotifType
		want      string
	}{
		{"character", MotifTypeCharacter, "character"},
		{"faction", MotifTypeFaction, "faction"},
		{"location", MotifTypeLocation, "location"},
		{"unknown", MotifType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.motifType.String(); got != tt.want {
				t.Errorf("MotifType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMotifGenerator(t *testing.T) {
	gen := NewMotifGenerator(44100, 12345)
	if gen == nil {
		t.Fatal("NewMotifGenerator() returned nil")
	}
	if gen.sampleRate != 44100 {
		t.Errorf("sampleRate = %d, want 44100", gen.sampleRate)
	}
	if gen.seed != 12345 {
		t.Errorf("seed = %d, want 12345", gen.seed)
	}
}

func TestMotifGenerator_GenerateMotif(t *testing.T) {
	gen := NewMotifGenerator(44100, 12345)

	tests := []struct {
		name      string
		entityID  string
		genre     string
		motifType MotifType
		wantNotes int // expected note range
	}{
		{"fantasy character", "hero_001", "fantasy", MotifTypeCharacter, 4},
		{"scifi faction", "corp_mega", "scifi", MotifTypeFaction, 4},
		{"horror location", "mansion_13", "horror", MotifTypeLocation, 4},
		{"cyberpunk character", "netrunner", "cyberpunk", MotifTypeCharacter, 4},
		{"postapoc faction", "raiders", "post-apocalyptic", MotifTypeFaction, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			motif := gen.GenerateMotif(tt.entityID, tt.genre, tt.motifType)

			if motif == nil {
				t.Fatal("GenerateMotif() returned nil")
			}

			if motif.ID != tt.entityID {
				t.Errorf("ID = %s, want %s", motif.ID, tt.entityID)
			}

			if motif.Type != tt.motifType {
				t.Errorf("Type = %v, want %v", motif.Type, tt.motifType)
			}

			if len(motif.Notes) < tt.wantNotes || len(motif.Notes) > 8 {
				t.Errorf("Notes count = %d, want 4-8 notes", len(motif.Notes))
			}

			if len(motif.Rhythm.Pattern) != len(motif.Notes) {
				t.Errorf("Rhythm pattern length = %d, want %d", len(motif.Rhythm.Pattern), len(motif.Notes))
			}

			if len(motif.Rhythm.Velocity) != len(motif.Notes) {
				t.Errorf("Rhythm velocity length = %d, want %d", len(motif.Rhythm.Velocity), len(motif.Notes))
			}

			if motif.Tempo <= 0 {
				t.Errorf("Tempo = %f, want > 0", motif.Tempo)
			}

			// Check that first and last notes are the same (motif identity)
			if len(motif.Notes) > 2 && motif.Notes[0] != motif.Notes[len(motif.Notes)-1] {
				t.Errorf("First note (%d) should equal last note (%d) for motif identity",
					motif.Notes[0], motif.Notes[len(motif.Notes)-1])
			}
		})
	}
}

func TestMotifGenerator_Determinism(t *testing.T) {
	// Same seed should produce identical motifs
	gen1 := NewMotifGenerator(44100, 12345)
	gen2 := NewMotifGenerator(44100, 12345)

	motif1 := gen1.GenerateMotif("hero", "fantasy", MotifTypeCharacter)
	motif2 := gen2.GenerateMotif("hero", "fantasy", MotifTypeCharacter)

	if len(motif1.Notes) != len(motif2.Notes) {
		t.Errorf("Note counts differ: %d vs %d", len(motif1.Notes), len(motif2.Notes))
	}

	for i := range motif1.Notes {
		if motif1.Notes[i] != motif2.Notes[i] {
			t.Errorf("Note %d differs: %d vs %d", i, motif1.Notes[i], motif2.Notes[i])
		}
	}

	if motif1.Tempo != motif2.Tempo {
		t.Errorf("Tempos differ: %f vs %f", motif1.Tempo, motif2.Tempo)
	}
}

func TestMotifGenerator_UniqueMotifs(t *testing.T) {
	// Different entities should produce different motifs
	gen := NewMotifGenerator(44100, 12345)

	motif1 := gen.GenerateMotif("entity_1", "fantasy", MotifTypeCharacter)
	motif2 := gen.GenerateMotif("entity_2", "fantasy", MotifTypeCharacter)

	// Check that motifs are different
	same := true
	if len(motif1.Notes) != len(motif2.Notes) {
		same = false
	} else {
		for i := range motif1.Notes {
			if motif1.Notes[i] != motif2.Notes[i] {
				same = false
				break
			}
		}
	}

	if same {
		t.Error("Different entities produced identical motifs")
	}
}

func TestMotifGenerator_GenreWaveforms(t *testing.T) {
	gen := NewMotifGenerator(44100, 12345)

	tests := []struct {
		genre         string
		expectedTypes []audio.WaveformType
	}{
		{"fantasy", []audio.WaveformType{audio.WaveformSine, audio.WaveformTriangle}},
		{"scifi", []audio.WaveformType{audio.WaveformSquare, audio.WaveformSawtooth}},
		{"horror", []audio.WaveformType{audio.WaveformSawtooth, audio.WaveformSquare}},
		{"cyberpunk", []audio.WaveformType{audio.WaveformSquare}},
		{"post-apocalyptic", []audio.WaveformType{audio.WaveformSawtooth}},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			motif := gen.GenerateMotif("entity", tt.genre, MotifTypeCharacter)

			validWaveform := false
			for _, expectedType := range tt.expectedTypes {
				if motif.Waveform == expectedType {
					validWaveform = true
					break
				}
			}

			if !validWaveform {
				t.Errorf("Genre %s produced unexpected waveform %v, expected one of %v",
					tt.genre, motif.Waveform, tt.expectedTypes)
			}
		})
	}
}

func TestMotifGenerator_MotifTypeTempos(t *testing.T) {
	gen := NewMotifGenerator(44100, 12345)

	tests := []struct {
		motifType MotifType
		minTempo  float64
		maxTempo  float64
	}{
		{MotifTypeCharacter, 100.0, 140.0},
		{MotifTypeFaction, 100.0, 120.0},
		{MotifTypeLocation, 80.0, 110.0},
	}

	for _, tt := range tests {
		t.Run(tt.motifType.String(), func(t *testing.T) {
			// Generate multiple motifs to test tempo range
			for i := 0; i < 5; i++ {
				entityID := "entity_" + string(rune(i+'0'))
				motif := gen.GenerateMotif(entityID, "fantasy", tt.motifType)

				if motif.Tempo < tt.minTempo || motif.Tempo > tt.maxTempo {
					t.Errorf("Tempo %f out of range [%f, %f] for %s",
						motif.Tempo, tt.minTempo, tt.maxTempo, tt.motifType.String())
				}
			}
		})
	}
}

func TestMotifGenerator_RhythmVelocities(t *testing.T) {
	gen := NewMotifGenerator(44100, 12345)
	motif := gen.GenerateMotif("entity", "fantasy", MotifTypeCharacter)

	// Check velocity ranges
	for i, vel := range motif.Rhythm.Velocity {
		if vel < 0.0 || vel > 1.0 {
			t.Errorf("Velocity[%d] = %f, want 0.0-1.0", i, vel)
		}
	}

	// First and last notes should have higher velocity (emphasis)
	firstVel := motif.Rhythm.Velocity[0]
	lastVel := motif.Rhythm.Velocity[len(motif.Rhythm.Velocity)-1]

	if firstVel < 0.8 {
		t.Errorf("First note velocity = %f, want >= 0.8 for emphasis", firstVel)
	}
	if lastVel < 0.8 {
		t.Errorf("Last note velocity = %f, want >= 0.8 for emphasis", lastVel)
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"simple", "test"},
		{"complex", "hero_001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashString(tt.input)
			hash2 := hashString(tt.input)

			// Same string should produce same hash
			if hash1 != hash2 {
				t.Errorf("hashString produced different hashes for same input")
			}

			// Different strings should produce different hashes (not guaranteed but highly likely)
			if tt.input != "" {
				differentHash := hashString(tt.input + "x")
				if hash1 == differentHash {
					t.Logf("Note: hash collision detected (rare but possible)")
				}
			}
		})
	}
}
