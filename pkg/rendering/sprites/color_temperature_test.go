package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestGenreColorTemperatureConfig_Deterministic(t *testing.T) {
	genres := []string{"", "fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		c1 := GenreColorTemperatureConfig(genre, 42)
		c2 := GenreColorTemperatureConfig(genre, 42)
		if c1 != c2 {
			t.Errorf("GenreColorTemperatureConfig(%q, 42) not deterministic", genre)
		}
	}
}

func TestGenreColorTemperatureConfig_Genres(t *testing.T) {
	tests := []struct {
		genre       string
		wantWarm    bool // WarmShift > 0
		wantCool    bool // CoolShift > 0
		wantSpecInt bool // SpecularIntensity > 0
	}{
		{"fantasy", true, true, true},
		{"horror", true, true, true},
		{"scifi", true, true, true},
		{"cyberpunk", true, true, true},
		{"postapoc", true, true, true},
		{"", true, true, true},
		{"unknown_genre", true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			cfg := GenreColorTemperatureConfig(tt.genre, 123)
			if (cfg.WarmShift > 0) != tt.wantWarm {
				t.Errorf("WarmShift = %f, want positive = %v", cfg.WarmShift, tt.wantWarm)
			}
			if (cfg.CoolShift > 0) != tt.wantCool {
				t.Errorf("CoolShift = %f, want positive = %v", cfg.CoolShift, tt.wantCool)
			}
			if (cfg.SpecularIntensity > 0) != tt.wantSpecInt {
				t.Errorf("SpecularIntensity = %f, want positive = %v", cfg.SpecularIntensity, tt.wantSpecInt)
			}
		})
	}
}

func TestApplyColorTemperature_NilImage(t *testing.T) {
	cfg := GenreColorTemperatureConfig("fantasy", 1)
	if n := ApplyColorTemperature(nil, cfg); n != 0 {
		t.Errorf("expected 0 modified for nil image, got %d", n)
	}
}

func TestApplyColorTemperature_EmptyImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	cfg := GenreColorTemperatureConfig("fantasy", 1)
	if n := ApplyColorTemperature(img, cfg); n != 0 {
		t.Errorf("expected 0 modified for empty image, got %d", n)
	}
}

func TestApplyColorTemperature_FullyTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	cfg := GenreColorTemperatureConfig("fantasy", 1)
	if n := ApplyColorTemperature(img, cfg); n != 0 {
		t.Errorf("expected 0 modified for transparent image, got %d", n)
	}
}

func TestApplyColorTemperature_ModifiesOpaquePixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	base := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, base)
		}
	}

	cfg := GenreColorTemperatureConfig("fantasy", 42)
	n := ApplyColorTemperature(img, cfg)
	if n == 0 {
		t.Fatal("expected some pixels to be modified")
	}

	// Verify at least some pixels changed from the base color
	changed := 0
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			c := img.RGBAAt(x, y)
			if c.R != base.R || c.G != base.G || c.B != base.B {
				changed++
			}
		}
	}
	if changed == 0 {
		t.Error("no pixels were color-shifted; expected warm/cool temperature changes")
	}
}

func TestApplyColorTemperature_Deterministic(t *testing.T) {
	makeImage := func() *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		for y := 6; y < 26; y++ {
			for x := 6; x < 26; x++ {
				img.SetRGBA(x, y, color.RGBA{R: 120, G: 100, B: 90, A: 255})
			}
		}
		return img
	}

	cfg := GenreColorTemperatureConfig("scifi", 999)
	img1 := makeImage()
	img2 := makeImage()
	ApplyColorTemperature(img1, cfg)
	ApplyColorTemperature(img2, cfg)

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)
			if c1 != c2 {
				t.Fatalf("non-deterministic at (%d,%d): %v vs %v", x, y, c1, c2)
			}
		}
	}
}

func TestApplyColorTemperature_WarmCoolSplit(t *testing.T) {
	// Place a circle of opaque pixels. Light from top-left (angle ~5.5 rad).
	// Top-left pixels should be warmer (higher R relative to B),
	// bottom-right should be cooler (higher B relative to R).
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	base := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			img.SetRGBA(x, y, base)
		}
	}

	cfg := GenreColorTemperatureConfig("fantasy", 77)
	ApplyColorTemperature(img, cfg)

	// Sample warm side (top-left quadrant center)
	warmPixel := img.RGBAAt(8, 8)
	// Sample cool side (bottom-right quadrant center)
	coolPixel := img.RGBAAt(24, 24)

	warmBalance := int(warmPixel.R) - int(warmPixel.B)
	coolBalance := int(coolPixel.R) - int(coolPixel.B)

	if warmBalance <= coolBalance {
		t.Errorf("warm side R-B (%d) should be greater than cool side R-B (%d)",
			warmBalance, coolBalance)
	}
}

func TestApplyColorTemperature_SpecularHighlight(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	base := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			img.SetRGBA(x, y, base)
		}
	}

	cfg := GenreColorTemperatureConfig("cyberpunk", 55)
	cfg.SpecularIntensity = 0.8 // boost for test visibility
	ApplyColorTemperature(img, cfg)

	// Find the brightest pixel — it should be significantly brighter than base
	var maxBrightness int
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			c := img.RGBAAt(x, y)
			brightness := int(c.R) + int(c.G) + int(c.B)
			if brightness > maxBrightness {
				maxBrightness = brightness
			}
		}
	}
	baseBrightness := int(base.R) + int(base.G) + int(base.B)
	if maxBrightness <= baseBrightness+30 {
		t.Errorf("specular highlight too weak: max brightness %d, base %d",
			maxBrightness, baseBrightness)
	}
}

func TestApplyColorTemperature_GenresProduceDifferentResults(t *testing.T) {
	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	results := make(map[string]color.RGBA)

	for _, genre := range genres {
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))
		for y := 2; y < 14; y++ {
			for x := 2; x < 14; x++ {
				img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
			}
		}
		cfg := GenreColorTemperatureConfig(genre, 42)
		ApplyColorTemperature(img, cfg)
		results[genre] = img.RGBAAt(4, 4)
	}

	// At least some genres should produce different pixel values
	unique := make(map[color.RGBA]bool)
	for _, c := range results {
		unique[c] = true
	}
	if len(unique) < 2 {
		t.Errorf("expected different genres to produce different results, got %d unique values", len(unique))
	}
}

func BenchmarkApplyColorTemperature_32x32(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	cfg := GenreColorTemperatureConfig("fantasy", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyColorTemperature(img, cfg)
	}
}

func BenchmarkApplyColorTemperature_64x64(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 8; y < 56; y++ {
		for x := 8; x < 56; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	cfg := GenreColorTemperatureConfig("scifi", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyColorTemperature(img, cfg)
	}
}
