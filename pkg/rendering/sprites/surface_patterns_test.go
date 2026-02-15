package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestSurfaceTextureTypeString(t *testing.T) {
	tests := []struct {
		tex  SurfaceTextureType
		want string
	}{
		{TexNone, "none"},
		{TexFur, "fur"},
		{TexScales, "scales"},
		{TexChitin, "chitin"},
		{TexMetal, "metal"},
		{TexBone, "bone"},
		{TexOoze, "ooze"},
		{TexFeathers, "feathers"},
		{TexBark, "bark"},
		{SurfaceTextureType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tex.String(); got != tt.want {
				t.Errorf("SurfaceTextureType(%d).String() = %q, want %q", tt.tex, got, tt.want)
			}
		})
	}
}

func TestTextureForCreatureForm(t *testing.T) {
	tests := []struct {
		form string
		want SurfaceTextureType
	}{
		{"quadruped", TexFur},
		{"serpentine", TexScales},
		{"arachnid", TexChitin},
		{"mechanical", TexMetal},
		{"undead", TexBone},
		{"blob", TexOoze},
		{"flying", TexFeathers},
		{"humanoid", TexNone},
		{"unknown_form", TexNone},
		{"", TexNone},
	}
	for _, tt := range tests {
		t.Run(tt.form, func(t *testing.T) {
			if got := TextureForCreatureForm(tt.form); got != tt.want {
				t.Errorf("TextureForCreatureForm(%q) = %v, want %v", tt.form, got, tt.want)
			}
		})
	}
}

func TestGenerateSurfaceTextureSet_Deterministic(t *testing.T) {
	seed := int64(42)
	form := "quadruped"
	genre := "fantasy"

	set1 := GenerateSurfaceTextureSet(seed, form, genre)
	set2 := GenerateSurfaceTextureSet(seed, form, genre)

	if set1.HeadTexture.Intensity != set2.HeadTexture.Intensity {
		t.Error("same seed should produce same head intensity")
	}
	if set1.TorsoTexture.Type != set2.TorsoTexture.Type {
		t.Error("same seed should produce same torso type")
	}
	if set1.LimbTexture.PrimaryColor != set2.LimbTexture.PrimaryColor {
		t.Error("same seed should produce same limb primary color")
	}
}

func TestGenerateSurfaceTextureSet_DifferentSeeds(t *testing.T) {
	form := "serpentine"
	genre := "fantasy"
	set1 := GenerateSurfaceTextureSet(100, form, genre)
	set2 := GenerateSurfaceTextureSet(200, form, genre)

	// Different seeds should produce different colors (probabilistic but highly likely)
	if set1.TorsoTexture.PrimaryColor == set2.TorsoTexture.PrimaryColor &&
		set1.HeadTexture.Intensity == set2.HeadTexture.Intensity {
		t.Error("different seeds should produce different texture sets (very unlikely to match)")
	}
}

func TestGenerateSurfaceTextureSet_Humanoid(t *testing.T) {
	set := GenerateSurfaceTextureSet(42, "humanoid", "fantasy")
	if set.HeadTexture.Type != TexNone {
		t.Errorf("humanoid should get TexNone, got %v", set.HeadTexture.Type)
	}
	if set.TorsoTexture.Type != TexNone {
		t.Errorf("humanoid torso should get TexNone, got %v", set.TorsoTexture.Type)
	}
}

func TestGenerateSurfaceTextureSet_GenreInfluence(t *testing.T) {
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			set := GenerateSurfaceTextureSet(42, "quadruped", genre)
			if set.TorsoTexture.Type != TexFur {
				t.Errorf("quadruped should always be TexFur regardless of genre, got %v", set.TorsoTexture.Type)
			}
			if set.TorsoTexture.Intensity <= 0 {
				t.Error("intensity should be positive")
			}
			if set.TorsoTexture.Intensity > 0.7 {
				t.Errorf("intensity should be clamped to 0.7, got %f", set.TorsoTexture.Intensity)
			}
		})
	}
}

func TestGenerateSurfaceTextureSet_AllForms(t *testing.T) {
	forms := []struct {
		form    string
		wantTex SurfaceTextureType
	}{
		{"quadruped", TexFur},
		{"serpentine", TexScales},
		{"arachnid", TexChitin},
		{"mechanical", TexMetal},
		{"undead", TexBone},
		{"blob", TexOoze},
		{"flying", TexFeathers},
	}
	for _, tt := range forms {
		t.Run(tt.form, func(t *testing.T) {
			set := GenerateSurfaceTextureSet(12345, tt.form, "fantasy")
			if set.TorsoTexture.Type != tt.wantTex {
				t.Errorf("form %q: expected %v, got %v", tt.form, tt.wantTex, set.TorsoTexture.Type)
			}
			if set.HeadTexture.Intensity >= set.TorsoTexture.Intensity {
				t.Error("head intensity should be lighter than torso")
			}
			if set.HeadTexture.Scale >= set.TorsoTexture.Scale {
				t.Error("head scale should be finer than torso")
			}
		})
	}
}

// makeTestBuf creates a test image buffer filled with a solid color.
func makeTestBuf(w, h int, fill color.RGBA) *image.RGBA {
	buf := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*buf.Stride + x*4)
			buf.Pix[idx+0] = fill.R
			buf.Pix[idx+1] = fill.G
			buf.Pix[idx+2] = fill.B
			buf.Pix[idx+3] = fill.A
		}
	}
	return buf
}

func TestApplySurfaceTexture_NoOp(t *testing.T) {
	buf := makeTestBuf(32, 32, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	original := make([]uint8, len(buf.Pix))
	copy(original, buf.Pix)

	// TexNone should not modify anything
	ApplySurfaceTexture(buf, buf.Bounds(), SurfaceTextureParams{Type: TexNone, Intensity: 0.5}, 42)
	for i := range buf.Pix {
		if buf.Pix[i] != original[i] {
			t.Fatal("TexNone should not modify buffer")
		}
	}

	// Zero intensity should not modify anything
	ApplySurfaceTexture(buf, buf.Bounds(), SurfaceTextureParams{Type: TexFur, Intensity: 0}, 42)
	for i := range buf.Pix {
		if buf.Pix[i] != original[i] {
			t.Fatal("zero intensity should not modify buffer")
		}
	}

	// Nil buffer should not panic
	ApplySurfaceTexture(nil, image.Rect(0, 0, 10, 10), SurfaceTextureParams{Type: TexFur, Intensity: 0.5}, 42)
}

func TestApplySurfaceTexture_AllTypes(t *testing.T) {
	types := []SurfaceTextureType{
		TexFur, TexScales, TexChitin, TexMetal,
		TexBone, TexOoze, TexFeathers, TexBark,
	}
	for _, texType := range types {
		t.Run(texType.String(), func(t *testing.T) {
			buf := makeTestBuf(32, 32, color.RGBA{R: 128, G: 128, B: 128, A: 255})
			original := make([]uint8, len(buf.Pix))
			copy(original, buf.Pix)

			params := SurfaceTextureParams{
				Type:           texType,
				Intensity:      0.4,
				Scale:          1.0,
				PrimaryColor:   color.RGBA{R: 100, G: 80, B: 60, A: 180},
				SecondaryColor: color.RGBA{R: 200, G: 200, B: 210, A: 150},
			}
			ApplySurfaceTexture(buf, buf.Bounds(), params, 42)

			// At least some pixels should have changed
			changed := 0
			for i := range buf.Pix {
				if buf.Pix[i] != original[i] {
					changed++
				}
			}
			if changed == 0 {
				t.Errorf("texture %v with intensity 0.4 should modify at least some pixels", texType)
			}
		})
	}
}

func TestApplySurfaceTexture_RespectsAlpha(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Fill half with opaque, half with transparent
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			idx := (y*buf.Stride + x*4)
			if x < 8 {
				buf.Pix[idx+0] = 128
				buf.Pix[idx+1] = 128
				buf.Pix[idx+2] = 128
				buf.Pix[idx+3] = 255
			} else {
				buf.Pix[idx+3] = 0 // transparent
			}
		}
	}

	originalRight := make([]uint8, 0)
	for y := 0; y < 16; y++ {
		for x := 8; x < 16; x++ {
			idx := (y*buf.Stride + x*4)
			originalRight = append(originalRight, buf.Pix[idx:idx+4]...)
		}
	}

	ApplySurfaceTexture(buf, buf.Bounds(), SurfaceTextureParams{
		Type:         TexFur,
		Intensity:    0.5,
		Scale:        1.0,
		PrimaryColor: color.RGBA{R: 200, G: 100, B: 50, A: 180},
	}, 42)

	// Transparent pixels should remain untouched
	i := 0
	for y := 0; y < 16; y++ {
		for x := 8; x < 16; x++ {
			idx := (y*buf.Stride + x*4)
			for c := 0; c < 4; c++ {
				if buf.Pix[idx+c] != originalRight[i+c] {
					t.Fatal("transparent region should not be modified")
				}
			}
			i += 4
		}
	}
}

func TestApplySurfaceTexture_SubRegion(t *testing.T) {
	buf := makeTestBuf(32, 32, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	original := make([]uint8, len(buf.Pix))
	copy(original, buf.Pix)

	// Apply to a small sub-region only
	ApplySurfaceTexture(buf, image.Rect(10, 10, 20, 20), SurfaceTextureParams{
		Type:         TexScales,
		Intensity:    0.5,
		Scale:        1.0,
		PrimaryColor: color.RGBA{R: 50, G: 150, B: 100, A: 160},
	}, 42)

	// Pixels outside the region should be unchanged
	for y := 0; y < 10; y++ {
		for x := 0; x < 32; x++ {
			idx := (y*buf.Stride + x*4)
			for c := 0; c < 4; c++ {
				if buf.Pix[idx+c] != original[idx+c] {
					t.Fatalf("pixel (%d,%d) outside region was modified", x, y)
				}
			}
		}
	}
}

func TestBlendPixel(t *testing.T) {
	buf := makeTestBuf(4, 4, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	blendPixel(buf, 1, 1, color.RGBA{R: 200, G: 200, B: 200, A: 255}, 0.5)

	idx := (1*buf.Stride + 1*4)
	r := buf.Pix[idx+0]
	// Should be roughly blended: 100*0.5 + 200*0.5 = 150
	if r < 140 || r > 160 {
		t.Errorf("expected blended R around 150, got %d", r)
	}
}

func TestBlendPixel_TransparentSkip(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Pixel (1,1) has alpha=0
	original := buf.Pix[(1*buf.Stride + 1*4):(1*buf.Stride + 1*4 + 4)]
	origCopy := make([]uint8, 4)
	copy(origCopy, original)

	blendPixel(buf, 1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 1.0)
	for i := 0; i < 4; i++ {
		if original[i] != origCopy[i] {
			t.Fatal("blendPixel should not modify transparent pixels")
		}
	}
}

func BenchmarkApplySurfaceTexture_Fur(b *testing.B) {
	buf := makeTestBuf(32, 32, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	params := SurfaceTextureParams{
		Type:           TexFur,
		Intensity:      0.4,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 140, G: 100, B: 60, A: 180},
		SecondaryColor: color.RGBA{R: 170, G: 130, B: 90, A: 120},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplySurfaceTexture(buf, buf.Bounds(), params, 42)
	}
}

func BenchmarkApplySurfaceTexture_Scales(b *testing.B) {
	buf := makeTestBuf(32, 32, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	params := SurfaceTextureParams{
		Type:           TexScales,
		Intensity:      0.4,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 40, G: 120, B: 80, A: 160},
		SecondaryColor: color.RGBA{R: 20, G: 100, B: 65, A: 200},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplySurfaceTexture(buf, buf.Bounds(), params, 42)
	}
}

func BenchmarkGenerateSurfaceTextureSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateSurfaceTextureSet(int64(i), "quadruped", "fantasy")
	}
}
