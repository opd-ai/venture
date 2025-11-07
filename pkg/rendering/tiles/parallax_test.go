// Package tiles provides tests for Phase 16.3 parallax depth effects.
package tiles

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// TestTileLayerString tests the String() method for TileLayer.
func TestTileLayerString(t *testing.T) {
	tests := []struct {
		name     string
		layer    TileLayer
		expected string
	}{
		{"background", LayerBackground, "background"},
		{"base", LayerBase, "base"},
		{"foreground", LayerForeground, "foreground"},
		{"unknown", TileLayer(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.layer.String(); got != tt.expected {
				t.Errorf("TileLayer.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDefaultParallaxConfig tests the default parallax configuration.
func TestDefaultParallaxConfig(t *testing.T) {
	config := DefaultParallaxConfig()

	if config.Layer != LayerBase {
		t.Errorf("Default layer = %v, want %v", config.Layer, LayerBase)
	}
	if config.CameraX != 0.0 {
		t.Errorf("Default CameraX = %v, want 0.0", config.CameraX)
	}
	if config.CameraY != 0.0 {
		t.Errorf("Default CameraY = %v, want 0.0", config.CameraY)
	}
	if config.ParallaxDepth != 1.0 {
		t.Errorf("Default ParallaxDepth = %v, want 1.0", config.ParallaxDepth)
	}
	if config.AOIntensity != 0.5 {
		t.Errorf("Default AOIntensity = %v, want 0.5", config.AOIntensity)
	}
	if config.ShadowHeight != 0.3 {
		t.Errorf("Default ShadowHeight = %v, want 0.3", config.ShadowHeight)
	}
	if config.ShadowAngle != math.Pi/4 {
		t.Errorf("Default ShadowAngle = %v, want %v", config.ShadowAngle, math.Pi/4)
	}
}

// TestParallaxConfigValidate tests validation of parallax configurations.
func TestParallaxConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  ParallaxConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultParallaxConfig(),
			wantErr: false,
		},
		{
			name: "invalid parallax depth negative",
			config: ParallaxConfig{
				BaseConfig:    DefaultConfig(),
				ParallaxDepth: -0.1,
				AOIntensity:   0.5,
				ShadowHeight:  0.3,
			},
			wantErr: true,
		},
		{
			name: "invalid parallax depth too high",
			config: ParallaxConfig{
				BaseConfig:    DefaultConfig(),
				ParallaxDepth: 2.1,
				AOIntensity:   0.5,
				ShadowHeight:  0.3,
			},
			wantErr: true,
		},
		{
			name: "invalid AO intensity negative",
			config: ParallaxConfig{
				BaseConfig:    DefaultConfig(),
				ParallaxDepth: 1.0,
				AOIntensity:   -0.1,
				ShadowHeight:  0.3,
			},
			wantErr: true,
		},
		{
			name: "invalid AO intensity too high",
			config: ParallaxConfig{
				BaseConfig:    DefaultConfig(),
				ParallaxDepth: 1.0,
				AOIntensity:   1.1,
				ShadowHeight:  0.3,
			},
			wantErr: true,
		},
		{
			name: "invalid shadow height negative",
			config: ParallaxConfig{
				BaseConfig:    DefaultConfig(),
				ParallaxDepth: 1.0,
				AOIntensity:   0.5,
				ShadowHeight:  -0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid shadow height too high",
			config: ParallaxConfig{
				BaseConfig:    DefaultConfig(),
				ParallaxDepth: 1.0,
				AOIntensity:   0.5,
				ShadowHeight:  1.1,
			},
			wantErr: true,
		},
		{
			name: "invalid base config",
			config: ParallaxConfig{
				BaseConfig: Config{
					Width:   -32, // Invalid
					Height:  32,
					GenreID: "fantasy",
				},
				ParallaxDepth: 1.0,
				AOIntensity:   0.5,
				ShadowHeight:  0.3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParallaxConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParallaxOffset tests parallax offset calculation.
func TestParallaxOffset(t *testing.T) {
	tests := []struct {
		name            string
		config          ParallaxConfig
		expectedOffsetX float64
		expectedOffsetY float64
		tolerance       float64
	}{
		{
			name: "background layer at camera 10,10",
			config: ParallaxConfig{
				Layer:         LayerBackground,
				CameraX:       10.0,
				CameraY:       10.0,
				ParallaxDepth: 1.0,
			},
			expectedOffsetX: 3.0, // 10.0 * 1.0 * 0.3
			expectedOffsetY: 3.0, // 10.0 * 1.0 * 0.3
			tolerance:       0.01,
		},
		{
			name: "base layer at camera 10,10",
			config: ParallaxConfig{
				Layer:         LayerBase,
				CameraX:       10.0,
				CameraY:       10.0,
				ParallaxDepth: 1.0,
			},
			expectedOffsetX: 10.0, // 10.0 * 1.0 * 1.0
			expectedOffsetY: 10.0, // 10.0 * 1.0 * 1.0
			tolerance:       0.01,
		},
		{
			name: "foreground layer at camera 10,10",
			config: ParallaxConfig{
				Layer:         LayerForeground,
				CameraX:       10.0,
				CameraY:       10.0,
				ParallaxDepth: 1.0,
			},
			expectedOffsetX: 14.0, // 10.0 * 1.0 * 1.4
			expectedOffsetY: 14.0, // 10.0 * 1.0 * 1.4
			tolerance:       0.01,
		},
		{
			name: "base layer with custom parallax depth",
			config: ParallaxConfig{
				Layer:         LayerBase,
				CameraX:       5.0,
				CameraY:       8.0,
				ParallaxDepth: 0.5,
			},
			expectedOffsetX: 2.5, // 5.0 * 0.5 * 1.0
			expectedOffsetY: 4.0, // 8.0 * 0.5 * 1.0
			tolerance:       0.01,
		},
		{
			name: "zero camera position",
			config: ParallaxConfig{
				Layer:         LayerBase,
				CameraX:       0.0,
				CameraY:       0.0,
				ParallaxDepth: 1.0,
			},
			expectedOffsetX: 0.0,
			expectedOffsetY: 0.0,
			tolerance:       0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offsetX, offsetY := tt.config.ParallaxOffset()

			if math.Abs(offsetX-tt.expectedOffsetX) > tt.tolerance {
				t.Errorf("ParallaxOffset() offsetX = %v, want %v", offsetX, tt.expectedOffsetX)
			}
			if math.Abs(offsetY-tt.expectedOffsetY) > tt.tolerance {
				t.Errorf("ParallaxOffset() offsetY = %v, want %v", offsetY, tt.expectedOffsetY)
			}
		})
	}
}

// TestGenerateWithParallax tests parallax tile generation.
func TestGenerateWithParallax(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		config  ParallaxConfig
		wantErr bool
	}{
		{
			name: "background layer floor",
			config: ParallaxConfig{
				BaseConfig: Config{
					Type:    TileFloor,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
				},
				Layer:         LayerBackground,
				CameraX:       10.0,
				CameraY:       10.0,
				ParallaxDepth: 0.3,
				AOIntensity:   0.3,
				ShadowHeight:  0.1,
				ShadowAngle:   math.Pi / 4,
			},
			wantErr: false,
		},
		{
			name: "base layer wall with AO and shadows",
			config: ParallaxConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "scifi",
					Seed:    67890,
					Variant: 0.7,
				},
				Layer:         LayerBase,
				CameraX:       0.0,
				CameraY:       0.0,
				ParallaxDepth: 1.0,
				AOIntensity:   0.5,
				ShadowHeight:  0.3,
				ShadowAngle:   math.Pi / 4,
			},
			wantErr: false,
		},
		{
			name: "foreground layer door",
			config: ParallaxConfig{
				BaseConfig: Config{
					Type:    TileDoor,
					Width:   32,
					Height:  32,
					GenreID: "horror",
					Seed:    11111,
					Variant: 0.3,
				},
				Layer:         LayerForeground,
				CameraX:       5.0,
				CameraY:       -3.0,
				ParallaxDepth: 1.4,
				AOIntensity:   0.2,
				ShadowHeight:  0.5,
				ShadowAngle:   math.Pi / 3,
			},
			wantErr: false,
		},
		{
			name: "invalid parallax config",
			config: ParallaxConfig{
				BaseConfig: Config{
					Type:    TileFloor,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
				},
				ParallaxDepth: 3.0, // Invalid
				AOIntensity:   0.5,
				ShadowHeight:  0.3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.GenerateWithParallax(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWithParallax() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("GenerateWithParallax() returned nil image")
					return
				}

				// Verify image dimensions
				bounds := img.Bounds()
				if bounds.Dx() != tt.config.BaseConfig.Width {
					t.Errorf("Image width = %d, want %d", bounds.Dx(), tt.config.BaseConfig.Width)
				}
				if bounds.Dy() != tt.config.BaseConfig.Height {
					t.Errorf("Image height = %d, want %d", bounds.Dy(), tt.config.BaseConfig.Height)
				}

				// Verify image is not completely transparent
				hasContent := false
				for y := 0; y < bounds.Dy(); y++ {
					for x := 0; x < bounds.Dx(); x++ {
						c := img.RGBAAt(x, y)
						if c.A > 0 {
							hasContent = true
							break
						}
					}
					if hasContent {
						break
					}
				}
				if !hasContent {
					t.Error("Generated image has no visible content")
				}
			}
		})
	}
}

// TestGenerateLayeredTile tests generation of all three layers.
func TestGenerateLayeredTile(t *testing.T) {
	gen := NewGenerator()

	baseConfig := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	bg, base, fg, err := gen.GenerateLayeredTile(baseConfig, 10.0, 5.0)
	if err != nil {
		t.Fatalf("GenerateLayeredTile() error = %v", err)
	}

	if bg == nil {
		t.Error("Background layer is nil")
	}
	if base == nil {
		t.Error("Base layer is nil")
	}
	if fg == nil {
		t.Error("Foreground layer is nil")
	}

	// Verify all layers have correct dimensions
	if bg != nil && bg.Bounds().Dx() != 32 {
		t.Errorf("Background width = %d, want 32", bg.Bounds().Dx())
	}
	if base != nil && base.Bounds().Dx() != 32 {
		t.Errorf("Base width = %d, want 32", base.Bounds().Dx())
	}
	if fg != nil && fg.Bounds().Dx() != 32 {
		t.Errorf("Foreground width = %d, want 32", fg.Bounds().Dx())
	}
}

// TestCompositeLayers tests layer compositing.
func TestCompositeLayers(t *testing.T) {
	// Create simple test images
	bounds := image.Rect(0, 0, 4, 4)

	bg := image.NewRGBA(bounds)
	base := image.NewRGBA(bounds)
	fg := image.NewRGBA(bounds)

	// Fill with distinct colors
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			bg.SetRGBA(x, y, color.RGBA{255, 0, 0, 255})   // Red
			base.SetRGBA(x, y, color.RGBA{0, 255, 0, 255}) // Green
			fg.SetRGBA(x, y, color.RGBA{0, 0, 255, 255})   // Blue
		}
	}

	composite := CompositeLayers(bg, base, fg)

	if composite == nil {
		t.Fatal("CompositeLayers() returned nil")
	}

	// Verify dimensions
	if composite.Bounds() != bounds {
		t.Errorf("Composite bounds = %v, want %v", composite.Bounds(), bounds)
	}

	// Verify compositing (foreground should be on top)
	c := composite.RGBAAt(0, 0)
	if c.B != 255 { // Should be blue (foreground)
		t.Errorf("Composite pixel color incorrect, got RGB(%d,%d,%d)", c.R, c.G, c.B)
	}
}

// TestGenerateAOMap tests ambient occlusion map generation.
func TestGenerateAOMap(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:    TileWall,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	aoMap, err := gen.GenerateAOMap(config, 0.5)
	if err != nil {
		t.Fatalf("GenerateAOMap() error = %v", err)
	}

	if aoMap == nil {
		t.Fatal("GenerateAOMap() returned nil")
	}

	// Verify dimensions
	bounds := aoMap.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("AO map dimensions = %dx%d, want 32x32", bounds.Dx(), bounds.Dy())
	}

	// Verify AO map is grayscale (R == G == B)
	allGrayscale := true
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c := aoMap.RGBAAt(x, y)
			if c.R != c.G || c.G != c.B {
				allGrayscale = false
				break
			}
		}
		if !allGrayscale {
			break
		}
	}

	if !allGrayscale {
		t.Error("AO map is not grayscale")
	}
}

// TestParallaxDeterminism tests that parallax generation is deterministic.
func TestParallaxDeterminism(t *testing.T) {
	gen := NewGenerator()

	config := ParallaxConfig{
		BaseConfig: Config{
			Type:    TileFloor,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    12345, // Fixed seed
			Variant: 0.5,
		},
		Layer:         LayerBase,
		CameraX:       10.0,
		CameraY:       5.0,
		ParallaxDepth: 1.0,
		AOIntensity:   0.5,
		ShadowHeight:  0.3,
		ShadowAngle:   math.Pi / 4,
	}

	// Generate twice with same config
	img1, err1 := gen.GenerateWithParallax(config)
	img2, err2 := gen.GenerateWithParallax(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generation errors: %v, %v", err1, err2)
	}

	// Compare pixel by pixel
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)

			if c1.R != c2.R || c1.G != c2.G || c1.B != c2.B || c1.A != c2.A {
				t.Errorf("Pixel (%d,%d) differs: img1=%v, img2=%v", x, y, c1, c2)
				return
			}
		}
	}
}

// TestLayerVisualDifferences tests that different layers have distinct visual characteristics.
func TestLayerVisualDifferences(t *testing.T) {
	gen := NewGenerator()

	baseConfig := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	bg, base, fg, err := gen.GenerateLayeredTile(baseConfig, 0.0, 0.0)
	if err != nil {
		t.Fatalf("GenerateLayeredTile() error = %v", err)
	}

	// Calculate average brightness for each layer
	avgBrightness := func(img *image.RGBA) float64 {
		bounds := img.Bounds()
		total := 0.0
		count := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := img.RGBAAt(x, y)
				total += float64(c.R+c.G+c.B) / 3.0
				count++
			}
		}
		return total / float64(count)
	}

	bgBrightness := avgBrightness(bg)
	baseBrightness := avgBrightness(base)
	fgBrightness := avgBrightness(fg)

	// Background should be darker than base (70% darkening applied)
	if bgBrightness >= baseBrightness {
		t.Errorf("Background brightness (%f) should be less than base brightness (%f)", bgBrightness, baseBrightness)
	}

	// Verify all layers are different from each other
	if bgBrightness == baseBrightness || baseBrightness == fgBrightness || bgBrightness == fgBrightness {
		t.Error("All layers should have different brightness values")
	}

	// Background should be the darkest
	if bgBrightness > baseBrightness || bgBrightness > fgBrightness {
		t.Errorf("Background should be darkest: bg=%f, base=%f, fg=%f", bgBrightness, baseBrightness, fgBrightness)
	}
}

// BenchmarkGenerateWithParallax benchmarks parallax tile generation.
func BenchmarkGenerateWithParallax(b *testing.B) {
	gen := NewGenerator()

	config := ParallaxConfig{
		BaseConfig: Config{
			Type:    TileFloor,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    12345,
			Variant: 0.5,
		},
		Layer:         LayerBase,
		CameraX:       10.0,
		CameraY:       5.0,
		ParallaxDepth: 1.0,
		AOIntensity:   0.5,
		ShadowHeight:  0.3,
		ShadowAngle:   math.Pi / 4,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateWithParallax(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateLayeredTile benchmarks generation of all three layers.
func BenchmarkGenerateLayeredTile(b *testing.B) {
	gen := NewGenerator()

	config := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := gen.GenerateLayeredTile(config, 10.0, 5.0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCompositeLayers benchmarks layer compositing.
func BenchmarkCompositeLayers(b *testing.B) {
	bounds := image.Rect(0, 0, 32, 32)

	bg := image.NewRGBA(bounds)
	base := image.NewRGBA(bounds)
	fg := image.NewRGBA(bounds)

	// Fill with colors
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			bg.SetRGBA(x, y, color.RGBA{255, 0, 0, 255})   // Red
			base.SetRGBA(x, y, color.RGBA{0, 255, 0, 255}) // Green
			fg.SetRGBA(x, y, color.RGBA{0, 0, 255, 255})   // Blue
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompositeLayers(bg, base, fg)
	}
}

// BenchmarkGenerateAOMap benchmarks AO map generation.
func BenchmarkGenerateAOMap(b *testing.B) {
	gen := NewGenerator()

	config := Config{
		Type:    TileWall,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateAOMap(config, 0.5)
		if err != nil {
			b.Fatal(err)
		}
	}
}
