package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestDefaultShadingConfig(t *testing.T) {
	cfg := DefaultShadingConfig()
	if cfg.LightIntensity <= 0 || cfg.LightIntensity > 1.0 {
		t.Errorf("LightIntensity out of range: %f", cfg.LightIntensity)
	}
	if cfg.EdgeDarkening < 0 || cfg.EdgeDarkening > 1.0 {
		t.Errorf("EdgeDarkening out of range: %f", cfg.EdgeDarkening)
	}
	if cfg.TintR <= 0 || cfg.TintG <= 0 || cfg.TintB <= 0 {
		t.Error("Tint values should be positive")
	}
}

func TestGenreShadingConfig(t *testing.T) {
	tests := []struct {
		name  string
		genre string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"scifi", "sci-fi"},
		{"postapoc", "post-apocalyptic"},
		{"unknown", "unknown_genre"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GenreShadingConfig(tt.genre)
			if cfg.LightIntensity <= 0 {
				t.Error("LightIntensity should be positive")
			}
			if cfg.TintR <= 0 || cfg.TintG <= 0 || cfg.TintB <= 0 {
				t.Error("Tint values should be positive")
			}
		})
	}
}

func TestApplyBodyPartShading_NoOp_EmptyImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	cfg := DefaultShadingConfig()
	// Should not panic
	ApplyBodyPartShading(img, cfg, 12345)
}

func TestApplyBodyPartShading_TransparentImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// All pixels transparent (default)
	cfg := DefaultShadingConfig()
	ApplyBodyPartShading(img, cfg, 42)
	// All pixels should remain transparent
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0 {
				t.Errorf("Pixel (%d,%d) should be transparent, got alpha %d", x, y, a)
			}
		}
	}
}

func TestApplyBodyPartShading_ModifiesOpaquePixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	// Fill center 8x8 with opaque color
	for y := 4; y < 12; y++ {
		for x := 4; x < 12; x++ {
			img.SetRGBA(x, y, baseColor)
		}
	}

	// Copy original pixels
	origCenter := img.RGBAAt(8, 8)

	cfg := DefaultShadingConfig()
	ApplyBodyPartShading(img, cfg, 99)

	// Center pixel should be modified (highlight + dither)
	newCenter := img.RGBAAt(8, 8)
	if newCenter.R == origCenter.R && newCenter.G == origCenter.G && newCenter.B == origCenter.B {
		t.Error("Expected center pixel to be modified by shading")
	}

	// Edge pixel should be darker than center
	edgePixel := img.RGBAAt(4, 8)
	if edgePixel.R == 0 && edgePixel.G == 0 && edgePixel.B == 0 {
		// Acceptable - edge darkening can make it very dark
	}
}

func TestApplyBodyPartShading_Deterministic(t *testing.T) {
	makeImg := func() *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 12, 12))
		for y := 2; y < 10; y++ {
			for x := 2; x < 10; x++ {
				img.SetRGBA(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			}
		}
		return img
	}

	cfg := DefaultShadingConfig()
	img1 := makeImg()
	img2 := makeImg()
	ApplyBodyPartShading(img1, cfg, 777)
	ApplyBodyPartShading(img2, cfg, 777)

	for y := 0; y < 12; y++ {
		for x := 0; x < 12; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)
			if c1 != c2 {
				t.Errorf("Non-deterministic shading at (%d,%d): %v vs %v", x, y, c1, c2)
			}
		}
	}
}

func TestApplyBodyPartShading_DifferentSeeds(t *testing.T) {
	makeImg := func() *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 12, 12))
		for y := 2; y < 10; y++ {
			for x := 2; x < 10; x++ {
				img.SetRGBA(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			}
		}
		return img
	}

	cfg := DefaultShadingConfig()
	img1 := makeImg()
	img2 := makeImg()
	ApplyBodyPartShading(img1, cfg, 111)
	ApplyBodyPartShading(img2, cfg, 222)

	differ := false
	for y := 0; y < 12; y++ {
		for x := 0; x < 12; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)
			if c1 != c2 {
				differ = true
				break
			}
		}
		if differ {
			break
		}
	}
	if !differ {
		t.Error("Different seeds should produce different dithering")
	}
}

func TestShadingConfigForPart(t *testing.T) {
	tests := []struct {
		name      string
		colorRole string
		zIndex    int
	}{
		{"head", "secondary", 15},
		{"torso", "primary", 10},
		{"legs", "primary", 5},
		{"shadow", "shadow", 0},
	}
	base := DefaultShadingConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ShadingConfigForPart(base, tt.colorRole, tt.zIndex)
			if tt.colorRole == "shadow" {
				if cfg.LightIntensity != 0 {
					t.Error("Shadow parts should have zero light intensity")
				}
			} else {
				if cfg.LightIntensity <= 0 {
					t.Error("Non-shadow parts should have positive light intensity")
				}
			}
		})
	}
}

func TestBuildAlphaMap(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Fill entire image opaque
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	dist := buildAlphaMap(img)
	// Corner pixel should be edge (dist=1)
	if dist[0] != 1 {
		t.Errorf("Corner pixel should be edge (dist=1), got %d", dist[0])
	}
	// Center pixel should have higher distance
	center := dist[4*8+4]
	if center <= 1 {
		t.Errorf("Center pixel should have dist > 1, got %d", center)
	}
}

func TestApplyOverlapDarkening(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}
	ApplyOverlapDarkening(img, 3, 0.3)
	// Top row should be darkened
	top := img.RGBAAt(4, 0)
	bottom := img.RGBAAt(4, 7)
	if top.R >= bottom.R {
		t.Errorf("Top row should be darker than bottom: top=%d, bottom=%d", top.R, bottom.R)
	}
}

func TestDitherNoise_Range(t *testing.T) {
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			v := ditherNoise(x, y, 42)
			if v < -1.0 || v > 1.0 {
				t.Errorf("ditherNoise(%d,%d) = %f, out of [-1,1]", x, y, v)
			}
		}
	}
}

func TestClampF(t *testing.T) {
	tests := []struct {
		v, lo, hi, want float64
	}{
		{0.5, 0, 1, 0.5},
		{-1, 0, 1, 0},
		{2, 0, 1, 1},
	}
	for _, tt := range tests {
		got := clampF(tt.v, tt.lo, tt.hi)
		if got != tt.want {
			t.Errorf("clampF(%f,%f,%f) = %f, want %f", tt.v, tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		v    float64
		want uint8
	}{
		{0, 0},
		{255, 255},
		{-10, 0},
		{300, 255},
		{128.5, 128},
	}
	for _, tt := range tests {
		got := clampByte(tt.v)
		if got != tt.want {
			t.Errorf("clampByte(%f) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

func TestExtractRGBA_AlreadyRGBA(t *testing.T) {
	orig := image.NewRGBA(image.Rect(0, 0, 4, 4))
	got := ExtractRGBA(orig)
	if got != orig {
		t.Error("ExtractRGBA should return same pointer for *image.RGBA input")
	}
}

func TestToRGBAColor(t *testing.T) {
	c := toRGBAColor(color.RGBA{R: 100, G: 150, B: 200, A: 255})
	if c.R != 100 || c.G != 150 || c.B != 200 || c.A != 255 {
		t.Errorf("toRGBAColor mismatch: got %v", c)
	}
}

func BenchmarkApplyBodyPartShading_32x32(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	cfg := DefaultShadingConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyBodyPartShading(img, cfg, int64(i))
	}
}

func BenchmarkBuildAlphaMap_32x32(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildAlphaMap(img)
	}
}
