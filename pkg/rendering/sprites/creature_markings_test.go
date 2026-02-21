package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestCreatureMarkingType_String(t *testing.T) {
	tests := []struct {
		marking CreatureMarkingType
		want    string
	}{
		{MarkingNone, "none"},
		{MarkingSpots, "spots"},
		{MarkingStripes, "stripes"},
		{MarkingPatches, "patches"},
		{MarkingRings, "rings"},
		{MarkingScales, "scales"},
		{MarkingGradient, "gradient"},
		{MarkingDappled, "dappled"},
		{MarkingBanded, "banded"},
		{MarkingMottled, "mottled"},
		{MarkingTiger, "tiger"},
		{MarkingLeopard, "leopard"},
		{MarkingZebra, "zebra"},
		{CreatureMarkingType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.marking.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateCreatureMarkings_Deterministic(t *testing.T) {
	forms := []string{"quadruped", "serpentine", "arachnid", "flying", "blob", "mechanical", "undead", "insect", "multi_limbed"}

	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			seed := int64(12345)
			m1 := GenerateCreatureMarkings(seed, form)
			m2 := GenerateCreatureMarkings(seed, form)

			if m1.Type != m2.Type {
				t.Errorf("Non-deterministic Type: %v vs %v", m1.Type, m2.Type)
			}
			if m1.SecondaryType != m2.SecondaryType {
				t.Errorf("Non-deterministic SecondaryType: %v vs %v", m1.SecondaryType, m2.SecondaryType)
			}
			if m1.PrimaryColor != m2.PrimaryColor {
				t.Errorf("Non-deterministic PrimaryColor: %v vs %v", m1.PrimaryColor, m2.PrimaryColor)
			}
			if m1.Density != m2.Density {
				t.Errorf("Non-deterministic Density: %v vs %v", m1.Density, m2.Density)
			}
			if m1.Scale != m2.Scale {
				t.Errorf("Non-deterministic Scale: %v vs %v", m1.Scale, m2.Scale)
			}
		})
	}
}

func TestGenerateCreatureMarkings_DifferentSeeds(t *testing.T) {
	form := "quadruped"
	seen := make(map[CreatureMarkingType]int)

	// Generate many markings to verify variety
	for seed := int64(0); seed < 100; seed++ {
		m := GenerateCreatureMarkings(seed, form)
		seen[m.Type]++
	}

	// Should see multiple different marking types
	if len(seen) < 3 {
		t.Errorf("Expected variety in markings, only got %d types: %v", len(seen), seen)
	}
}

func TestGenerateCreatureMarkings_FormSpecific(t *testing.T) {
	// Different forms should have different marking probabilities
	serpent := GenerateCreatureMarkings(42, "serpentine")
	quadruped := GenerateCreatureMarkings(42, "quadruped")

	// Same seed but different forms should produce different results
	// (due to different weight tables)
	// Note: with the same seed, different RNG paths may still yield different results
	// This test verifies the function doesn't crash and produces valid output
	if serpent.Type >= markingCount {
		t.Errorf("Serpent marking type out of range: %v", serpent.Type)
	}
	if quadruped.Type >= markingCount {
		t.Errorf("Quadruped marking type out of range: %v", quadruped.Type)
	}
}

func TestGenerateCreatureMarkings_ValidRanges(t *testing.T) {
	forms := []string{"quadruped", "serpentine", "arachnid", "flying", "blob"}

	for _, form := range forms {
		for seed := int64(0); seed < 50; seed++ {
			m := GenerateCreatureMarkings(seed, form)

			if m.Type >= markingCount {
				t.Errorf("Type out of range for seed %d form %s: %v", seed, form, m.Type)
			}
			if m.SecondaryType >= markingCount {
				t.Errorf("SecondaryType out of range: %v", m.SecondaryType)
			}
			if m.Density < 0 || m.Density > 1 {
				t.Errorf("Density out of range: %v", m.Density)
			}
			if m.Scale < 0.5 || m.Scale > 2.0 {
				t.Errorf("Scale out of range: %v", m.Scale)
			}
			if m.Intensity < 0 || m.Intensity > 1 {
				t.Errorf("Intensity out of range: %v", m.Intensity)
			}
		}
	}
}

func TestRenderCreatureMarkings_NilBuffer(t *testing.T) {
	// Should not panic with nil buffer
	RenderCreatureMarkings(nil, CreatureMarkingParams{
		Width:    32,
		Height:   32,
		Markings: CreatureMarkings{Type: MarkingSpots},
	})
}

func TestRenderCreatureMarkings_ZeroSize(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 0, 0))
	// Should not panic with zero-size params
	RenderCreatureMarkings(buf, CreatureMarkingParams{
		Width:    0,
		Height:   0,
		Markings: CreatureMarkings{Type: MarkingSpots},
	})
}

func TestRenderCreatureMarkings_NoMarkings(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	// Fill with a base color
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			buf.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	originalPixels := make([]byte, len(buf.Pix))
	copy(originalPixels, buf.Pix)

	// MarkingNone should not modify buffer
	RenderCreatureMarkings(buf, CreatureMarkingParams{
		Width:    32,
		Height:   32,
		Markings: CreatureMarkings{Type: MarkingNone},
	})

	for i := range buf.Pix {
		if buf.Pix[i] != originalPixels[i] {
			t.Error("MarkingNone modified the buffer")
			break
		}
	}
}

func TestRenderCreatureMarkings_AllTypes(t *testing.T) {
	markingTypes := []CreatureMarkingType{
		MarkingSpots, MarkingStripes, MarkingPatches, MarkingRings,
		MarkingScales, MarkingGradient, MarkingDappled, MarkingBanded,
		MarkingMottled, MarkingTiger, MarkingLeopard, MarkingZebra,
	}

	for _, mt := range markingTypes {
		t.Run(mt.String(), func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
			// Fill with base color (need opaque pixels for markings to apply)
			for y := 8; y < 24; y++ {
				for x := 8; x < 24; x++ {
					buf.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
				}
			}

			params := CreatureMarkingParams{
				Width:     32,
				Height:    32,
				Form:      "quadruped",
				Direction: "down",
				Seed:      12345,
				Markings: CreatureMarkings{
					Type:         mt,
					PrimaryColor: color.RGBA{R: 200, G: 150, B: 100, A: 200},
					Density:      0.5,
					Scale:        1.0,
					Intensity:    0.7,
					Rotation:     15,
				},
			}

			// Should not panic
			RenderCreatureMarkings(buf, params)

			// Verify some pixels were modified (marking applied)
			modified := false
			for y := 8; y < 24; y++ {
				for x := 8; x < 24; x++ {
					c := buf.RGBAAt(x, y)
					if c.R != 100 || c.G != 100 || c.B != 100 {
						modified = true
						break
					}
				}
				if modified {
					break
				}
			}

			if !modified {
				t.Logf("Warning: %s marking did not visibly modify buffer", mt.String())
			}
		})
	}
}

func TestSelectMarkingForForm_Distribution(t *testing.T) {
	forms := []string{"quadruped", "serpentine", "arachnid", "blob", "mechanical"}

	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			counts := make(map[CreatureMarkingType]int)
			for seed := int64(0); seed < 500; seed++ {
				m := GenerateCreatureMarkings(seed, form)
				counts[m.Type]++
			}

			// Each form should produce a reasonable variety
			nonNoneCount := 0
			for mt, c := range counts {
				if mt != MarkingNone && c > 0 {
					nonNoneCount++
				}
			}

			if nonNoneCount < 2 {
				t.Errorf("Form %s produced too few marking varieties: %v", form, counts)
			}
		})
	}
}

func TestGenerateMarkingColors_ValidColors(t *testing.T) {
	forms := []string{"quadruped", "serpentine", "flying", "blob", "undead"}

	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			m := GenerateCreatureMarkings(42, form)

			// Verify colors are valid (non-zero alpha)
			if m.PrimaryColor.A == 0 {
				t.Error("Primary color has zero alpha")
			}
			if m.SecondaryColor.A == 0 {
				t.Error("Secondary color has zero alpha")
			}
		})
	}
}

func TestRenderCreatureMarkings_SecondaryMarkings(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	// Fill with base color
	for y := 5; y < 27; y++ {
		for x := 5; x < 27; x++ {
			buf.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	params := CreatureMarkingParams{
		Width:     32,
		Height:    32,
		Form:      "quadruped",
		Direction: "down",
		Seed:      99999,
		Markings: CreatureMarkings{
			Type:           MarkingSpots,
			SecondaryType:  MarkingStripes,
			PrimaryColor:   color.RGBA{R: 200, G: 150, B: 100, A: 200},
			SecondaryColor: color.RGBA{R: 80, G: 60, B: 40, A: 180},
			Density:        0.6,
			Scale:          1.0,
			Intensity:      0.8,
		},
	}

	// Should not panic with secondary markings
	RenderCreatureMarkings(buf, params)
}

func BenchmarkGenerateCreatureMarkings(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateCreatureMarkings(int64(i), "quadruped")
	}
}

func BenchmarkRenderCreatureMarkings_Spots(b *testing.B) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			buf.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	params := CreatureMarkingParams{
		Width:     32,
		Height:    32,
		Form:      "quadruped",
		Direction: "down",
		Seed:      12345,
		Markings: CreatureMarkings{
			Type:         MarkingSpots,
			PrimaryColor: color.RGBA{R: 200, G: 150, B: 100, A: 200},
			Density:      0.5,
			Scale:        1.0,
			Intensity:    0.7,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderCreatureMarkings(buf, params)
	}
}

func BenchmarkRenderCreatureMarkings_Stripes(b *testing.B) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			buf.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	params := CreatureMarkingParams{
		Width:     32,
		Height:    32,
		Form:      "serpentine",
		Direction: "up",
		Seed:      12345,
		Markings: CreatureMarkings{
			Type:         MarkingStripes,
			PrimaryColor: color.RGBA{R: 80, G: 120, B: 60, A: 200},
			Density:      0.5,
			Scale:        1.0,
			Intensity:    0.7,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderCreatureMarkings(buf, params)
	}
}
