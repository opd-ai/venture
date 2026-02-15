package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestClothingPatternTypeString(t *testing.T) {
	tests := []struct {
		name string
		p    ClothingPatternType
		want string
	}{
		{"none", PatternNone, "none"},
		{"hstripes", PatternHStripes, "horizontal_stripes"},
		{"vstripes", PatternVStripes, "vertical_stripes"},
		{"checker", PatternCheckerboard, "checkerboard"},
		{"dots", PatternDots, "dots"},
		{"border", PatternBorder, "border"},
		{"herring", PatternHerringbone, "herringbone"},
		{"diamond", PatternDiamondLattice, "diamond_lattice"},
		{"gradientv", PatternGradientV, "gradient_v"},
		{"unknown", ClothingPatternType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateClothingPatternSet_Deterministic(t *testing.T) {
	seed := int64(42)
	a := GenerateClothingPatternSet(seed)
	b := GenerateClothingPatternSet(seed)

	if a.TorsoPattern.Type != b.TorsoPattern.Type {
		t.Error("same seed should produce same torso pattern type")
	}
	if a.ArmPattern.Type != b.ArmPattern.Type {
		t.Error("same seed should produce same arm pattern type")
	}
	if a.LegPattern.Type != b.LegPattern.Type {
		t.Error("same seed should produce same leg pattern type")
	}
}

func TestGenerateClothingPatternSet_Variety(t *testing.T) {
	types := make(map[ClothingPatternType]int)
	for seed := int64(0); seed < 200; seed++ {
		set := GenerateClothingPatternSet(seed)
		types[set.TorsoPattern.Type]++
	}
	// With 200 seeds and 9 pattern types (including None), we expect at least 3 distinct types
	if len(types) < 3 {
		t.Errorf("expected variety in patterns, got only %d distinct types", len(types))
	}
}

func TestApplyClothingPattern_NoOpOnNone(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Fill with red
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 200
		img.Pix[i+1] = 50
		img.Pix[i+2] = 50
		img.Pix[i+3] = 255
	}
	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	ApplyClothingPattern(img, ClothingPattern{Type: PatternNone}, 1)

	for i, v := range img.Pix {
		if v != original[i] {
			t.Fatal("PatternNone should not modify pixels")
		}
	}
}

func TestApplyClothingPattern_ModifiesPixels(t *testing.T) {
	patterns := []ClothingPatternType{
		PatternHStripes, PatternVStripes, PatternCheckerboard,
		PatternDots, PatternBorder, PatternHerringbone,
		PatternDiamondLattice, PatternGradientV,
	}

	for _, ptype := range patterns {
		t.Run(ptype.String(), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 16, 16))
			// Fill opaque red
			for i := 0; i < len(img.Pix); i += 4 {
				img.Pix[i] = 200
				img.Pix[i+1] = 50
				img.Pix[i+2] = 50
				img.Pix[i+3] = 255
			}

			pattern := ClothingPattern{
				Type:         ptype,
				PatternColor: color.RGBA{R: 50, G: 50, B: 200, A: 255},
				Scale:        1.0,
				Intensity:    0.6,
			}

			ApplyClothingPattern(img, pattern, 42)

			// Check that at least some pixels changed
			changed := 0
			for i := 0; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 200 || img.Pix[i+1] != 50 || img.Pix[i+2] != 50 {
					changed++
				}
			}
			if changed == 0 {
				t.Errorf("pattern %s did not modify any pixels", ptype)
			}
		})
	}
}

func TestApplyClothingPattern_SkipsTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Leave all pixels transparent (alpha=0)
	pattern := ClothingPattern{
		Type:         PatternHStripes,
		PatternColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Scale:        1.0,
		Intensity:    1.0,
	}

	ApplyClothingPattern(img, pattern, 1)

	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 || img.Pix[i+1] != 0 || img.Pix[i+2] != 0 {
			t.Fatal("transparent pixels should not be modified")
		}
	}
}

func TestApplyClothingPattern_ZeroIntensity(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 100
		img.Pix[i+1] = 100
		img.Pix[i+2] = 100
		img.Pix[i+3] = 255
	}
	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	pattern := ClothingPattern{
		Type:         PatternVStripes,
		PatternColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Scale:        1.0,
		Intensity:    0.0,
	}

	ApplyClothingPattern(img, pattern, 1)

	for i, v := range img.Pix {
		if v != original[i] {
			t.Fatal("zero intensity should not modify pixels")
		}
	}
}

func TestClothingPatternSet_PatternForBodyRegion(t *testing.T) {
	set := ClothingPatternSet{
		TorsoPattern: ClothingPattern{Type: PatternHStripes},
		ArmPattern:   ClothingPattern{Type: PatternVStripes},
		LegPattern:   ClothingPattern{Type: PatternCheckerboard},
	}

	tests := []struct {
		name      string
		colorRole string
		zIndex    int
		wantType  ClothingPatternType
	}{
		{"torso", "primary", 10, PatternHStripes},
		{"legs", "primary", 5, PatternCheckerboard},
		{"arms", "accent1", 12, PatternVStripes},
		{"other", "secondary", 20, PatternNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := set.PatternForBodyRegion(tt.colorRole, tt.zIndex)
			if p.Type != tt.wantType {
				t.Errorf("got pattern type %v, want %v", p.Type, tt.wantType)
			}
		})
	}
}

func TestApplyClothingPattern_EmptyImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	pattern := ClothingPattern{
		Type:         PatternHStripes,
		PatternColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Scale:        1.0,
		Intensity:    0.5,
	}
	// Should not panic
	ApplyClothingPattern(img, pattern, 1)
}

func BenchmarkApplyClothingPattern(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 150
		img.Pix[i+1] = 80
		img.Pix[i+2] = 80
		img.Pix[i+3] = 255
	}
	pattern := ClothingPattern{
		Type:         PatternHerringbone,
		PatternColor: color.RGBA{R: 50, G: 50, B: 150, A: 255},
		Scale:        1.0,
		Intensity:    0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyClothingPattern(img, pattern, 42)
	}
}

func BenchmarkGenerateClothingPatternSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateClothingPatternSet(int64(i))
	}
}
