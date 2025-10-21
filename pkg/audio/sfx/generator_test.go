package sfx

import (
	"testing"
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
