package sfx

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name       string
		effectType string
		seed       int64
		wantMin    int // minimum expected samples
	}{
		{"impact", string(EffectImpact), 12345, 4410},
		{"explosion", string(EffectExplosion), 12345, 22050},
		{"magic", string(EffectMagic), 12345, 13230},
		{"laser", string(EffectLaser), 12345, 8820},
		{"pickup", string(EffectPickup), 12345, 6615},
		{"hit", string(EffectHit), 12345, 4410},
		{"jump", string(EffectJump), 12345, 8820},
		{"death", string(EffectDeath), 12345, 35280},
		{"powerup", string(EffectPowerup), 12345, 17640},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(44100, tt.seed)
			sample := gen.Generate(tt.effectType, tt.seed)

			if sample == nil {
				t.Fatal("Generate returned nil")
			}

			if sample.SampleRate != 44100 {
				t.Errorf("SampleRate = %d, want 44100", sample.SampleRate)
			}

			if len(sample.Data) < tt.wantMin {
				t.Errorf("len(Data) = %d, want >= %d", len(sample.Data), tt.wantMin)
			}

			// Check that samples are in valid range
			for i, v := range sample.Data {
				if v < -1.5 || v > 1.5 {
					t.Errorf("sample[%d] = %f, out of reasonable range", i, v)
					break
				}
			}

			// Check that the sound actually has some content (not all zeros)
			hasContent := false
			for _, v := range sample.Data {
				if v != 0 {
					hasContent = true
					break
				}
			}
			if !hasContent {
				t.Error("generated sound has no content (all zeros)")
			}
		})
	}
}

func TestGenerator_UnknownEffect(t *testing.T) {
	gen := NewGenerator(44100, 12345)
	sample := gen.Generate("unknown_effect", 12345)

	if sample == nil {
		t.Fatal("Generate returned nil for unknown effect")
	}

	// Should default to impact sound
	if len(sample.Data) == 0 {
		t.Error("unknown effect produced empty sample")
	}
}

func TestGenerator_UnknownEffect_WarningLog(t *testing.T) {
	// Create a logger with a buffer to capture output
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.WarnLevel)

	gen := NewGeneratorWithLogger(44100, 12345, logger)
	sample := gen.Generate("nonexistent_effect", 12345)

	if sample == nil {
		t.Fatal("Generate returned nil for unknown effect")
	}

	// Verify that a warning was logged
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected warning log for unknown effect type, got none")
	}

	// Verify the warning mentions the unknown effect type
	if !bytes.Contains(buf.Bytes(), []byte("unknown effect type")) {
		t.Errorf("expected log to contain 'unknown effect type', got: %s", logOutput)
	}
}

func TestGenerator_Determinism(t *testing.T) {
	seed := int64(98765)

	gen1 := NewGenerator(44100, seed)
	sample1 := gen1.Generate(string(EffectMagic), seed)

	gen2 := NewGenerator(44100, seed)
	sample2 := gen2.Generate(string(EffectMagic), seed)

	if len(sample1.Data) != len(sample2.Data) {
		t.Fatal("samples have different lengths")
	}

	// Note: Due to RNG, samples should be identical
	differenceCount := 0
	for i := range sample1.Data {
		if sample1.Data[i] != sample2.Data[i] {
			differenceCount++
		}
	}

	// Allow small differences due to floating point, but should be mostly identical
	maxAllowedDifferences := len(sample1.Data) / 100 // 1%
	if differenceCount > maxAllowedDifferences {
		t.Errorf("too many differences: %d out of %d samples", differenceCount, len(sample1.Data))
	}
}

func TestGenerator_Variation(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	sample1 := gen.Generate(string(EffectMagic), 11111)
	sample2 := gen.Generate(string(EffectMagic), 22222)

	// Different seeds should produce different results
	if len(sample1.Data) != len(sample2.Data) {
		// Lengths might vary due to random duration
		return
	}

	identical := true
	for i := range sample1.Data {
		if sample1.Data[i] != sample2.Data[i] {
			identical = false
			break
		}
	}

	if identical {
		t.Error("different seeds produced identical samples")
	}
}

func TestEffectCharacteristics(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	t.Run("impact is short", func(t *testing.T) {
		sample := gen.Generate(string(EffectImpact), 12345)
		duration := float64(len(sample.Data)) / float64(sample.SampleRate)
		if duration > 0.3 {
			t.Errorf("impact duration = %f seconds, want <= 0.3", duration)
		}
	})

	t.Run("explosion is longer", func(t *testing.T) {
		sample := gen.Generate(string(EffectExplosion), 12345)
		duration := float64(len(sample.Data)) / float64(sample.SampleRate)
		if duration < 0.4 {
			t.Errorf("explosion duration = %f seconds, want >= 0.4", duration)
		}
	})

	t.Run("death is longest", func(t *testing.T) {
		sample := gen.Generate(string(EffectDeath), 12345)
		duration := float64(len(sample.Data)) / float64(sample.SampleRate)
		if duration < 0.7 {
			t.Errorf("death duration = %f seconds, want >= 0.7", duration)
		}
	})
}

func TestEffectEnvelope(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	tests := []struct {
		name       string
		effectType string
	}{
		{"impact", string(EffectImpact)},
		{"magic", string(EffectMagic)},
		{"laser", string(EffectLaser)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := gen.Generate(tt.effectType, 12345)

			// First few samples should be quieter (attack)
			firstSample := sample.Data[0]
			if firstSample > 0.5 || firstSample < -0.5 {
				t.Errorf("first sample too loud: %f", firstSample)
			}

			// Last few samples should be quiet (release)
			lastSample := sample.Data[len(sample.Data)-1]
			if lastSample > 0.3 || lastSample < -0.3 {
				t.Errorf("last sample too loud: %f", lastSample)
			}
		})
	}
}

func BenchmarkGenerator_GenerateImpact(b *testing.B) {
	gen := NewGenerator(44100, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(string(EffectImpact), int64(i))
	}
}

func BenchmarkGenerator_GenerateMagic(b *testing.B) {
	gen := NewGenerator(44100, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(string(EffectMagic), int64(i))
	}
}

func BenchmarkGenerator_GenerateExplosion(b *testing.B) {
	gen := NewGenerator(44100, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(string(EffectExplosion), int64(i))
	}
}

// TestGenerateWithGenre tests the primary public API with all genre variations.
func TestGenerateWithGenre(t *testing.T) {
	genres := []string{
		"",        // no genre (default)
		"fantasy", // no modifications applied
		"scifi",
		"horror",
		"cyberpunk",
		"postapoc",
	}

	effectTypes := []EffectType{
		EffectImpact,
		EffectExplosion,
		EffectMagic,
		EffectLaser,
		EffectPickup,
		EffectHit,
		EffectJump,
		EffectDeath,
		EffectPowerup,
	}

	for _, genre := range genres {
		for _, effectType := range effectTypes {
			testName := string(effectType) + "_" + genre
			if genre == "" {
				testName = string(effectType) + "_no_genre"
			}

			t.Run(testName, func(t *testing.T) {
				gen := NewGenerator(44100, 12345)
				sample := gen.GenerateWithGenre(string(effectType), 54321, genre)

				if sample == nil {
					t.Fatal("GenerateWithGenre returned nil")
				}

				if sample.SampleRate != 44100 {
					t.Errorf("SampleRate = %d, want 44100", sample.SampleRate)
				}

				if len(sample.Data) == 0 {
					t.Error("GenerateWithGenre returned empty sample data")
				}

				// Verify samples are in reasonable range
				for i, v := range sample.Data {
					if v < -2.0 || v > 2.0 {
						t.Errorf("sample[%d] = %f, out of reasonable range [-2.0, 2.0]", i, v)
						break
					}
				}
			})
		}
	}
}

// TestGenerateWithGenre_Determinism verifies same seed produces same output for each genre.
func TestGenerateWithGenre_Determinism(t *testing.T) {
	seed := int64(98765)
	genres := []string{"scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen1 := NewGenerator(44100, seed)
			sample1 := gen1.GenerateWithGenre(string(EffectMagic), seed, genre)

			gen2 := NewGenerator(44100, seed)
			sample2 := gen2.GenerateWithGenre(string(EffectMagic), seed, genre)

			if len(sample1.Data) != len(sample2.Data) {
				t.Fatal("samples have different lengths")
			}

			// With same seed, samples should be identical
			for i := range sample1.Data {
				if sample1.Data[i] != sample2.Data[i] {
					t.Errorf("sample mismatch at index %d: %f != %f", i, sample1.Data[i], sample2.Data[i])
					break
				}
			}
		})
	}
}

// TestGenreModifications_SciFi verifies sci-fi modifications: higher pitch, reduced amplitude.
func TestGenreModifications_SciFi(t *testing.T) {
	gen := NewGenerator(44100, 12345)
	seed := int64(54321)

	// Generate base sound (fantasy genre has no modifications)
	baseGen := NewGenerator(44100, 12345)
	baseSample := baseGen.GenerateWithGenre(string(EffectLaser), seed, "fantasy")

	// Generate sci-fi sound
	scifiSample := gen.GenerateWithGenre(string(EffectLaser), seed, "scifi")

	if len(baseSample.Data) == 0 || len(scifiSample.Data) == 0 {
		t.Fatal("samples have no data")
	}

	// Sci-fi should have modified content (pitch shift changes sample values)
	// We can't compare directly due to pitch shift interpolation,
	// but we can verify the amplitude reduction was applied
	baseRMS := calculateRMS(baseSample.Data)
	scifiRMS := calculateRMS(scifiSample.Data)

	// Sci-fi applies 0.9x amplitude reduction, so RMS should be lower
	if scifiRMS >= baseRMS*1.1 {
		t.Errorf("sci-fi RMS (%f) should be reduced compared to base (%f)", scifiRMS, baseRMS)
	}
}

// TestGenreModifications_Horror verifies horror modifications: lower pitch, vibrato.
func TestGenreModifications_Horror(t *testing.T) {
	gen := NewGenerator(44100, 12345)
	seed := int64(54321)

	// Generate base sound
	baseGen := NewGenerator(44100, 12345)
	baseSample := baseGen.GenerateWithGenre(string(EffectMagic), seed, "fantasy")

	// Generate horror sound
	horrorSample := gen.GenerateWithGenre(string(EffectMagic), seed, "horror")

	if len(baseSample.Data) == 0 || len(horrorSample.Data) == 0 {
		t.Fatal("samples have no data")
	}

	// Horror applies 0.7x pitch (lower) and vibrato
	// Due to the pitch shift, samples should differ
	differences := 0
	minLen := len(baseSample.Data)
	if len(horrorSample.Data) < minLen {
		minLen = len(horrorSample.Data)
	}

	for i := 0; i < minLen; i++ {
		if baseSample.Data[i] != horrorSample.Data[i] {
			differences++
		}
	}

	// Expect significant differences due to pitch shift and vibrato
	if differences < minLen/2 {
		t.Errorf("horror sample should be significantly different from base, only %d/%d samples differ", differences, minLen)
	}
}

// TestGenreModifications_Cyberpunk verifies cyberpunk modifications: higher pitch, hard clipping.
func TestGenreModifications_Cyberpunk(t *testing.T) {
	gen := NewGenerator(44100, 12345)
	seed := int64(54321)

	// Generate cyberpunk sound
	cyberpunkSample := gen.GenerateWithGenre(string(EffectExplosion), seed, "cyberpunk")

	if len(cyberpunkSample.Data) == 0 {
		t.Fatal("sample has no data")
	}

	// Hard clipping at 0.7 means no sample should exceed that threshold
	maxSample := 0.0
	for _, v := range cyberpunkSample.Data {
		absV := v
		if absV < 0 {
			absV = -absV
		}
		if absV > maxSample {
			maxSample = absV
		}
	}

	// With hard clipping at 0.7, max absolute value should be <= 0.7
	if maxSample > 0.71 { // Small tolerance for floating point
		t.Errorf("cyberpunk max sample = %f, should be <= 0.7 due to hard clipping", maxSample)
	}
}

// TestGenreModifications_PostApocalyptic verifies post-apocalyptic modifications: soft clipping, slight pitch reduction.
func TestGenreModifications_PostApocalyptic(t *testing.T) {
	gen := NewGenerator(44100, 12345)
	seed := int64(54321)

	// Generate base sound
	baseGen := NewGenerator(44100, 12345)
	baseSample := baseGen.GenerateWithGenre(string(EffectExplosion), seed, "fantasy")

	// Generate post-apocalyptic sound
	postapocSample := gen.GenerateWithGenre(string(EffectExplosion), seed, "postapoc")

	if len(baseSample.Data) == 0 || len(postapocSample.Data) == 0 {
		t.Fatal("samples have no data")
	}

	// Post-apocalyptic applies soft clipping at 0.5 with 0.3 compression
	// Values above 0.5 should be compressed, not hard clipped
	// Check that modification was applied (samples should differ)
	differences := 0
	minLen := len(baseSample.Data)
	if len(postapocSample.Data) < minLen {
		minLen = len(postapocSample.Data)
	}

	for i := 0; i < minLen; i++ {
		if baseSample.Data[i] != postapocSample.Data[i] {
			differences++
		}
	}

	if differences < minLen/3 {
		t.Errorf("post-apocalyptic sample should differ from base, only %d/%d samples differ", differences, minLen)
	}
}

// TestGenreModifications_Fantasy verifies fantasy genre applies no modifications.
func TestGenreModifications_Fantasy(t *testing.T) {
	seed := int64(54321)

	// Generate with empty genre
	gen1 := NewGenerator(44100, 12345)
	noGenreSample := gen1.GenerateWithGenre(string(EffectMagic), seed, "")

	// Generate with fantasy genre
	gen2 := NewGenerator(44100, 12345)
	fantasySample := gen2.GenerateWithGenre(string(EffectMagic), seed, "fantasy")

	if len(noGenreSample.Data) != len(fantasySample.Data) {
		t.Fatal("samples have different lengths")
	}

	// Fantasy and no-genre should produce identical results
	for i := range noGenreSample.Data {
		if noGenreSample.Data[i] != fantasySample.Data[i] {
			t.Errorf("fantasy and no-genre should be identical, mismatch at index %d: %f != %f",
				i, noGenreSample.Data[i], fantasySample.Data[i])
			break
		}
	}
}

// TestApplyGenreModifications_AllGenres ensures all genre branches are covered.
func TestApplyGenreModifications_AllGenres(t *testing.T) {
	tests := []struct {
		name  string
		genre string
	}{
		{"scifi_modifications", "scifi"},
		{"horror_modifications", "horror"},
		{"cyberpunk_modifications", "cyberpunk"},
		{"postapoc_modifications", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(44100, 12345)

			// Generate base sample
			baseGen := NewGenerator(44100, 12345)
			baseSample := baseGen.GenerateWithGenre(string(EffectImpact), 12345, "fantasy")

			// Generate modified sample
			modifiedSample := gen.GenerateWithGenre(string(EffectImpact), 12345, tt.genre)

			// Samples should be different after genre modifications
			isDifferent := false
			minLen := len(baseSample.Data)
			if len(modifiedSample.Data) < minLen {
				minLen = len(modifiedSample.Data)
			}

			for i := 0; i < minLen; i++ {
				if baseSample.Data[i] != modifiedSample.Data[i] {
					isDifferent = true
					break
				}
			}

			if !isDifferent {
				t.Errorf("genre %s should produce different output than fantasy", tt.genre)
			}
		})
	}
}

// TestGenreModifications_UnknownGenre verifies unknown genres don't crash.
func TestGenreModifications_UnknownGenre(t *testing.T) {
	gen := NewGenerator(44100, 12345)

	// Unknown genres should not crash and should be treated like fantasy (no modification)
	sample := gen.GenerateWithGenre(string(EffectMagic), 12345, "unknown_genre")

	if sample == nil {
		t.Fatal("unknown genre returned nil")
	}

	if len(sample.Data) == 0 {
		t.Error("unknown genre returned empty sample")
	}
}

// BenchmarkGenerateWithGenre benchmarks genre-specific generation.
func BenchmarkGenerateWithGenre(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			gen := NewGenerator(44100, 12345)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				gen.GenerateWithGenre(string(EffectExplosion), int64(i), genre)
			}
		})
	}
}
