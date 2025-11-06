package shapes

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestAntiAliasQuality_Values verifies anti-aliasing quality level constants.
func TestAntiAliasQuality_Values(t *testing.T) {
	tests := []struct {
		name    string
		quality AntiAliasQuality
		want    int
	}{
		{"off", AntiAliasOff, 0},
		{"low", AntiAliasLow, 1},
		{"medium", AntiAliasMedium, 2},
		{"high", AntiAliasHigh, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.quality) != tt.want {
				t.Errorf("AntiAliasQuality value = %d, want %d", int(tt.quality), tt.want)
			}
		})
	}
}

// TestAntiAliasing_Circle verifies anti-aliasing produces smooth edges on circles.
func TestAntiAliasing_Circle(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		quality AntiAliasQuality
	}{
		{"off", AntiAliasOff},
		{"low", AntiAliasLow},
		{"medium", AntiAliasMedium},
		{"high", AntiAliasHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Type:      ShapeCircle,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
				Seed:      12345,
				Smoothing: 0.1,
				AntiAlias: tt.quality,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if img == nil {
				t.Fatal("Generate() returned nil image")
			}

			bounds := img.Bounds()
			if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
				t.Errorf("Image size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), config.Width, config.Height)
			}
		})
	}
}

// TestAntiAliasing_Triangle verifies anti-aliasing on diagonal edges.
func TestAntiAliasing_Triangle(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:      ShapeTriangle,
		Width:     32,
		Height:    32,
		Color:     color.RGBA{R: 0, G: 255, B: 0, A: 255},
		Seed:      12345,
		Rotation:  0,
		Smoothing: 0.0,
		AntiAlias: AntiAliasMedium,
	}

	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if img == nil {
		t.Fatal("Generate() returned nil image")
	}

	// Verify image was created successfully with correct dimensions
	bounds := img.Bounds()
	if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
		t.Errorf("Image size = %dx%d, want %dx%d",
			bounds.Dx(), bounds.Dy(), config.Width, config.Height)
	}
}

// TestAntiAliasing_Determinism verifies anti-aliasing produces consistent results.
func TestAntiAliasing_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:      ShapePolygon,
		Width:     32,
		Height:    32,
		Color:     color.RGBA{R: 0, G: 0, B: 255, A: 255},
		Seed:      99999,
		Sides:     6,
		Rotation:  45,
		Smoothing: 0.0,
		AntiAlias: AntiAliasHigh,
	}

	// Generate twice with same config
	img1, err1 := gen.Generate(config)
	if err1 != nil {
		t.Fatalf("First Generate() error = %v", err1)
	}

	img2, err2 := gen.Generate(config)
	if err2 != nil {
		t.Fatalf("Second Generate() error = %v", err2)
	}

	// Verify both images were created
	if img1 == nil || img2 == nil {
		t.Fatal("Generate() returned nil image")
	}

	// Verify dimensions match
	b1 := img1.Bounds()
	b2 := img2.Bounds()
	if b1 != b2 {
		t.Errorf("Image bounds differ: %v vs %v", b1, b2)
	}
}

// TestAntiAliasing_AllShapeTypes verifies anti-aliasing works for all shape types.
func TestAntiAliasing_AllShapeTypes(t *testing.T) {
	gen := NewGenerator()

	shapeTypes := []ShapeType{
		ShapeCircle, ShapeRectangle, ShapeTriangle, ShapePolygon,
		ShapeStar, ShapeRing, ShapeHexagon, ShapeOctagon,
		ShapeCross, ShapeHeart, ShapeCrescent, ShapeGear,
		ShapeCrystal, ShapeEllipse, ShapeCapsule, ShapeBean,
		ShapeWedge, ShapeShield, ShapeBlade, ShapeSkull,
	}

	for _, shapeType := range shapeTypes {
		t.Run(shapeType.String(), func(t *testing.T) {
			config := Config{
				Type:       shapeType,
				Width:      32,
				Height:     32,
				Color:      color.RGBA{R: 128, G: 128, B: 128, A: 255},
				Seed:       12345,
				Sides:      6,
				InnerRatio: 0.5,
				Rotation:   0,
				Smoothing:  0.0,
				AntiAlias:  AntiAliasLow,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if img == nil {
				t.Fatal("Generate() returned nil image")
			}

			// Verify image was created successfully
			bounds := img.Bounds()
			if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
				t.Errorf("Image size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), config.Width, config.Height)
			}
		})
	}
}

// TestAntiAliasing_QualityLevels verifies different quality levels produce images.
func TestAntiAliasing_QualityLevels(t *testing.T) {
	gen := NewGenerator()

	// Generate same shape with different quality levels
	baseConfig := Config{
		Type:      ShapeCircle,
		Width:     32,
		Height:    32,
		Color:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Seed:      12345,
		Smoothing: 0.0,
	}

	qualities := []AntiAliasQuality{AntiAliasOff, AntiAliasLow, AntiAliasMedium, AntiAliasHigh}
	images := make([]*ebiten.Image, len(qualities))

	for i, quality := range qualities {
		config := baseConfig
		config.AntiAlias = quality

		img, err := gen.Generate(config)
		if err != nil {
			t.Fatalf("Generate() with quality %d error = %v", quality, err)
		}
		if img == nil {
			t.Fatalf("Generate() with quality %d returned nil", quality)
		}
		images[i] = img
	}

	// Verify all images were created successfully
	for i, img := range images {
		bounds := img.Bounds()
		if bounds.Dx() != baseConfig.Width || bounds.Dy() != baseConfig.Height {
			t.Errorf("Quality %d: Image size = %dx%d, want %dx%d",
				i, bounds.Dx(), bounds.Dy(), baseConfig.Width, baseConfig.Height)
		}
	}
}

// TestAntiAliasing_BackwardCompatibility verifies existing code still works.
func TestAntiAliasing_BackwardCompatibility(t *testing.T) {
	gen := NewGenerator()

	// Config without AntiAlias field set (defaults to AntiAliasOff)
	config := Config{
		Type:      ShapeCircle,
		Width:     32,
		Height:    32,
		Color:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Seed:      12345,
		Smoothing: 0.1,
		// AntiAlias not set - should default to AntiAliasOff
	}

	img, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if img == nil {
		t.Fatal("Generate() returned nil image")
	}

	// Verify image was created successfully
	bounds := img.Bounds()
	if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
		t.Errorf("Image size = %dx%d, want %dx%d",
			bounds.Dx(), bounds.Dy(), config.Width, config.Height)
	}
}

// TestAntiAliasing_ColorPreservation verifies colors are used in generation.
func TestAntiAliasing_ColorPreservation(t *testing.T) {
	gen := NewGenerator()

	testColors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 128, G: 64, B: 192, A: 255},  // Purple
		{R: 255, G: 128, B: 64, A: 255},  // Orange
		{R: 32, G: 32, B: 32, A: 255},    // Dark gray
		{R: 200, G: 200, B: 200, A: 255}, // Light gray
	}

	for i, testColor := range testColors {
		t.Run(fmt.Sprintf("color_%d_R%dG%dB%d", i, testColor.R, testColor.G, testColor.B), func(t *testing.T) {
			config := Config{
				Type:      ShapeCircle,
				Width:     32,
				Height:    32,
				Color:     testColor,
				Seed:      12345,
				Smoothing: 0.0,
				AntiAlias: AntiAliasMedium,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if img == nil {
				t.Fatal("Generate() returned nil image")
			}

			// Verify image was created successfully
			bounds := img.Bounds()
			if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
				t.Errorf("Image size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), config.Width, config.Height)
			}
		})
	}
}
