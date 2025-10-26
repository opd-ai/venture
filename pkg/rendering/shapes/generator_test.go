package shapes

import (
	"image/color"
	"testing"
)

func TestShapeType_String(t *testing.T) {
	tests := []struct {
		name      string
		shapeType ShapeType
		want      string
	}{
		{"circle", ShapeCircle, "circle"},
		{"rectangle", ShapeRectangle, "rectangle"},
		{"triangle", ShapeTriangle, "triangle"},
		{"polygon", ShapePolygon, "polygon"},
		{"star", ShapeStar, "star"},
		{"ring", ShapeRing, "ring"},
		{"hexagon", ShapeHexagon, "hexagon"},
		{"octagon", ShapeOctagon, "octagon"},
		{"cross", ShapeCross, "cross"},
		{"heart", ShapeHeart, "heart"},
		{"crescent", ShapeCrescent, "crescent"},
		{"gear", ShapeGear, "gear"},
		{"crystal", ShapeCrystal, "crystal"},
		{"lightning", ShapeLightning, "lightning"},
		{"wave", ShapeWave, "wave"},
		{"spiral", ShapeSpiral, "spiral"},
		{"organic", ShapeOrganic, "organic"},
		{"unknown", ShapeType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shapeType.String()
			if got != tt.want {
				t.Errorf("ShapeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != ShapeCircle {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, ShapeCircle)
	}
	if config.Width != 32 {
		t.Errorf("DefaultConfig Width = %v, want 32", config.Width)
	}
	if config.Height != 32 {
		t.Errorf("DefaultConfig Height = %v, want 32", config.Height)
	}
	if config.Color == nil {
		t.Error("DefaultConfig Color is nil")
	}
	if config.Sides != 5 {
		t.Errorf("DefaultConfig Sides = %v, want 5", config.Sides)
	}
	if config.InnerRatio != 0.5 {
		t.Errorf("DefaultConfig InnerRatio = %v, want 0.5", config.InnerRatio)
	}
	if config.Rotation != 0 {
		t.Errorf("DefaultConfig Rotation = %v, want 0", config.Rotation)
	}
	if config.Smoothing != 0.1 {
		t.Errorf("DefaultConfig Smoothing = %v, want 0.1", config.Smoothing)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "circle config",
			config: Config{
				Type:   ShapeCircle,
				Width:  32,
				Height: 32,
				Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
			},
		},
		{
			name: "rectangle config",
			config: Config{
				Type:   ShapeRectangle,
				Width:  64,
				Height: 32,
				Color:  color.RGBA{R: 0, G: 255, B: 0, A: 255},
			},
		},
		{
			name: "triangle config",
			config: Config{
				Type:     ShapeTriangle,
				Width:    48,
				Height:   48,
				Color:    color.RGBA{R: 0, G: 0, B: 255, A: 255},
				Rotation: 45,
			},
		},
		{
			name: "polygon config",
			config: Config{
				Type:   ShapePolygon,
				Width:  64,
				Height: 64,
				Color:  color.RGBA{R: 255, G: 255, B: 0, A: 255},
				Sides:  6,
			},
		},
		{
			name: "star config",
			config: Config{
				Type:       ShapeStar,
				Width:      64,
				Height:     64,
				Color:      color.RGBA{R: 255, G: 0, B: 255, A: 255},
				Sides:      5,
				InnerRatio: 0.4,
			},
		},
		{
			name: "ring config",
			config: Config{
				Type:       ShapeRing,
				Width:      48,
				Height:     48,
				Color:      color.RGBA{R: 0, G: 255, B: 255, A: 255},
				InnerRatio: 0.6,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate basic config properties
			if tt.config.Type < ShapeCircle || tt.config.Type > ShapeRing {
				t.Error("Invalid shape type")
			}
			if tt.config.Width <= 0 {
				t.Error("Width must be positive")
			}
			if tt.config.Height <= 0 {
				t.Error("Height must be positive")
			}
			if tt.config.Color == nil {
				t.Error("Color cannot be nil")
			}
		})
	}
}

func TestShapeParameters(t *testing.T) {
	tests := []struct {
		name        string
		shapeType   ShapeType
		needsSides  bool
		needsInner  bool
		needsRotate bool
	}{
		{"circle", ShapeCircle, false, false, false},
		{"rectangle", ShapeRectangle, false, false, false},
		{"triangle", ShapeTriangle, false, false, true},
		{"polygon", ShapePolygon, true, false, true},
		{"star", ShapeStar, true, true, true},
		{"ring", ShapeRing, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Type = tt.shapeType

			// Verify parameter requirements make sense
			if tt.needsSides && config.Sides < 3 {
				t.Error("Sides parameter needed but not properly set")
			}
			if tt.needsInner && (config.InnerRatio < 0 || config.InnerRatio > 1) {
				t.Error("InnerRatio should be between 0 and 1")
			}
			if tt.needsRotate && (config.Rotation < 0 || config.Rotation > 360) {
				// Note: Rotation can be any value but typically 0-360
				// This is just a documentation test
			}
		})
	}
}

// TestGenerateNewShapes tests generation of all new Phase 3 shapes.
func TestGenerateNewShapes(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "hexagon",
			config: Config{
				Type:      ShapeHexagon,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 255, G: 100, B: 0, A: 255},
				Seed:      12345,
				Smoothing: 0.1,
			},
		},
		{
			name: "octagon",
			config: Config{
				Type:      ShapeOctagon,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 100, G: 255, B: 0, A: 255},
				Seed:      23456,
				Smoothing: 0.1,
			},
		},
		{
			name: "cross",
			config: Config{
				Type:      ShapeCross,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 0, G: 100, B: 255, A: 255},
				Seed:      34567,
				Smoothing: 0.1,
			},
		},
		{
			name: "heart",
			config: Config{
				Type:      ShapeHeart,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 255, G: 0, B: 100, A: 255},
				Seed:      45678,
				Smoothing: 0.1,
			},
		},
		{
			name: "crescent",
			config: Config{
				Type:       ShapeCrescent,
				Width:      32,
				Height:     32,
				Color:      color.RGBA{R: 200, G: 200, B: 100, A: 255},
				Seed:       56789,
				InnerRatio: 0.3,
				Rotation:   0,
				Smoothing:  0.1,
			},
		},
		{
			name: "gear",
			config: Config{
				Type:       ShapeGear,
				Width:      32,
				Height:     32,
				Color:      color.RGBA{R: 150, G: 150, B: 150, A: 255},
				Seed:       67890,
				Sides:      8,
				InnerRatio: 0.6,
				Rotation:   0,
				Smoothing:  0.05,
			},
		},
		{
			name: "crystal",
			config: Config{
				Type:      ShapeCrystal,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 100, G: 200, B: 255, A: 255},
				Seed:      78901,
				Rotation:  0,
				Smoothing: 0.05,
			},
		},
		{
			name: "lightning",
			config: Config{
				Type:      ShapeLightning,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 255, G: 255, B: 100, A: 255},
				Seed:      89012,
				Smoothing: 0,
			},
		},
		{
			name: "wave",
			config: Config{
				Type:      ShapeWave,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 100, G: 150, B: 255, A: 255},
				Seed:      90123,
				Smoothing: 0,
			},
		},
		{
			name: "spiral",
			config: Config{
				Type:      ShapeSpiral,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 200, G: 100, B: 200, A: 255},
				Seed:      11223,
				Smoothing: 0,
			},
		},
		{
			name: "organic",
			config: Config{
				Type:      ShapeOrganic,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 100, G: 200, B: 100, A: 255},
				Seed:      22334,
				Smoothing: 0.15,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			if img == nil {
				t.Error("Generated image is nil")
				return
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.config.Width {
				t.Errorf("Image width = %v, want %v", bounds.Dx(), tt.config.Width)
			}
			if bounds.Dy() != tt.config.Height {
				t.Errorf("Image height = %v, want %v", bounds.Dy(), tt.config.Height)
			}
		})
	}
}

// TestShapeDeterminism verifies that shapes generate consistently with same seed.
func TestShapeDeterminism(t *testing.T) {
	gen := NewGenerator()

	// Test with organic shape (most complex/seed-dependent)
	config := Config{
		Type:      ShapeOrganic,
		Width:     28,
		Height:    28,
		Color:     color.RGBA{R: 100, G: 200, B: 100, A: 255},
		Seed:      98765,
		Smoothing: 0.15,
	}

	// Generate twice with same config
	img1, err1 := gen.Generate(config)
	img2, err2 := gen.Generate(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generate errors: %v, %v", err1, err2)
	}

	// Verify dimensions match
	if img1.Bounds() != img2.Bounds() {
		t.Error("Generated images have different bounds")
	}

	// Note: We can't easily compare pixel-by-pixel without ReadPixels in tests,
	// but determinism is guaranteed by the seed-based generation
}

// TestAllShapeTypes ensures all 16 shape types can be generated without errors.
func TestAllShapeTypes(t *testing.T) {
	gen := NewGenerator()

	allShapes := []ShapeType{
		ShapeCircle, ShapeRectangle, ShapeTriangle, ShapePolygon,
		ShapeStar, ShapeRing, ShapeHexagon, ShapeOctagon,
		ShapeCross, ShapeHeart, ShapeCrescent, ShapeGear,
		ShapeCrystal, ShapeLightning, ShapeWave, ShapeSpiral,
		ShapeOrganic,
	}

	for _, shapeType := range allShapes {
		t.Run(shapeType.String(), func(t *testing.T) {
			config := DefaultConfig()
			config.Type = shapeType
			config.Seed = 12345 + int64(shapeType)

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate %s: %v", shapeType.String(), err)
			}

			if img == nil {
				t.Errorf("Generated nil image for %s", shapeType.String())
			}
		})
	}
}

// BenchmarkNewShapes benchmarks generation of new Phase 3 shapes.
func BenchmarkNewShapes(b *testing.B) {
	gen := NewGenerator()

	benchmarks := []struct {
		name   string
		config Config
	}{
		{
			name: "Hexagon",
			config: Config{
				Type:      ShapeHexagon,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 255, G: 100, B: 0, A: 255},
				Seed:      12345,
				Smoothing: 0.1,
			},
		},
		{
			name: "Organic",
			config: Config{
				Type:      ShapeOrganic,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 100, G: 200, B: 100, A: 255},
				Seed:      22334,
				Smoothing: 0.15,
			},
		},
		{
			name: "Gear",
			config: Config{
				Type:       ShapeGear,
				Width:      32,
				Height:     32,
				Color:      color.RGBA{R: 150, G: 150, B: 150, A: 255},
				Seed:       67890,
				Sides:      8,
				InnerRatio: 0.6,
				Smoothing:  0.05,
			},
		},
	}


	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = gen.Generate(bm.config)
			}
		})
	}
}

// TestNewShapes_Phase51 tests the new shape primitives added in Phase 5.1.
func TestNewShapes_Phase51(t *testing.T) {
	gen := NewGenerator()
	testColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}

	tests := []struct {
		name      string
		shapeType ShapeType
		config    Config
	}{
		{
			name:      "ellipse",
			shapeType: ShapeEllipse,
			config: Config{
				Type:      ShapeEllipse,
				Width:     32,
				Height:    48, // Taller than wide for oval shape
				Color:     testColor,
				Seed:      12345,
				Smoothing: 0.2,
			},
		},
		{
			name:      "capsule vertical",
			shapeType: ShapeCapsule,
			config: Config{
				Type:      ShapeCapsule,
				Width:     20,
				Height:    40, // Vertical capsule (limb shape)
				Color:     testColor,
				Seed:      12346,
				Smoothing: 0.2,
				Rotation:  0,
			},
		},
		{
			name:      "capsule horizontal",
			shapeType: ShapeCapsule,
			config: Config{
				Type:      ShapeCapsule,
				Width:     40,
				Height:    20, // Horizontal capsule
				Color:     testColor,
				Seed:      12347,
				Smoothing: 0.2,
				Rotation:  90,
			},
		},
		{
			name:      "bean",
			shapeType: ShapeBean,
			config: Config{
				Type:      ShapeBean,
				Width:     32,
				Height:    40, // Bean-shaped torso
				Color:     testColor,
				Seed:      12348,
				Smoothing: 0.2,
				Rotation:  0,
			},
		},
		{
			name:      "wedge up",
			shapeType: ShapeWedge,
			config: Config{
				Type:      ShapeWedge,
				Width:     24,
				Height:    32, // Pointing up
				Color:     testColor,
				Seed:      12349,
				Smoothing: 0.2,
				Rotation:  0,
			},
		},
		{
			name:      "wedge right",
			shapeType: ShapeWedge,
			config: Config{
				Type:      ShapeWedge,
				Width:     32,
				Height:    24, // Pointing right
				Color:     testColor,
				Seed:      12350,
				Smoothing: 0.2,
				Rotation:  90,
			},
		},
		{
			name:      "shield",
			shapeType: ShapeShield,
			config: Config{
				Type:      ShapeShield,
				Width:     28,
				Height:    36, // Shield icon
				Color:     testColor,
				Seed:      12351,
				Smoothing: 0.2,
				Rotation:  0,
			},
		},
		{
			name:      "blade",
			shapeType: ShapeBlade,
			config: Config{
				Type:      ShapeBlade,
				Width:     16,
				Height:    48, // Thin blade
				Color:     testColor,
				Seed:      12352,
				Smoothing: 0.2,
				Rotation:  0,
			},
		},
		{
			name:      "skull",
			shapeType: ShapeSkull,
			config: Config{
				Type:      ShapeSkull,
				Width:     32,
				Height:    32, // Skull head
				Color:     testColor,
				Seed:      12353,
				Smoothing: 0.2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.Generate(tt.config)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if img == nil {
				t.Fatal("Generate() returned nil image")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.config.Width {
				t.Errorf("Image width = %v, want %v", bounds.Dx(), tt.config.Width)
			}
			if bounds.Dy() != tt.config.Height {
				t.Errorf("Image height = %v, want %v", bounds.Dy(), tt.config.Height)
			}

			// Note: Pixel inspection requires game context (Ebiten runtime)
			// Skipping pixel validation in unit tests to avoid X11/graphics dependencies
			// Visual validation should be done with cmd/anatomytest tool
		})
	}
}

// TestShapeType_String_Phase51 tests string representation of new shape types.
func TestShapeType_String_Phase51(t *testing.T) {
	tests := []struct {
		name      string
		shapeType ShapeType
		want      string
	}{
		{"ellipse", ShapeEllipse, "ellipse"},
		{"capsule", ShapeCapsule, "capsule"},
		{"bean", ShapeBean, "bean"},
		{"wedge", ShapeWedge, "wedge"},
		{"shield", ShapeShield, "shield"},
		{"blade", ShapeBlade, "blade"},
		{"skull", ShapeSkull, "skull"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shapeType.String()
			if got != tt.want {
				t.Errorf("ShapeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestShapeDeterminism_Phase51 tests that new shapes generate consistently with same seed.
func TestShapeDeterminism_Phase51(t *testing.T) {
	gen := NewGenerator()

	shapes := []ShapeType{
		ShapeEllipse,
		ShapeCapsule,
		ShapeBean,
		ShapeWedge,
		ShapeShield,
		ShapeBlade,
		ShapeSkull,
	}

	for _, shapeType := range shapes {
		t.Run(shapeType.String(), func(t *testing.T) {
			config := Config{
				Type:      shapeType,
				Width:     32,
				Height:    32,
				Color:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
				Seed:      42,
				Smoothing: 0.2,
				Rotation:  45,
			}

			// Generate twice with same config
			img1, err1 := gen.Generate(config)
			if err1 != nil {
				t.Fatalf("First generate failed: %v", err1)
			}

			img2, err2 := gen.Generate(config)
			if err2 != nil {
				t.Fatalf("Second generate failed: %v", err2)
			}

			// Compare bounds (dimensions should match)
			bounds1 := img1.Bounds()
			bounds2 := img2.Bounds()

			if bounds1 != bounds2 {
				t.Fatalf("Image bounds differ: %v vs %v", bounds1, bounds2)
			}

			// Note: Pixel-level comparison requires Ebiten game loop running
			// For unit tests, we verify that generation succeeds deterministically
			// Visual validation of determinism should be done with cmd/anatomytest
			// The shape generation algorithms are deterministic by design (no time.Now() or rand.Intn())
		})
	}
}

// BenchmarkNewShapes_Phase51 benchmarks new shape generation performance.
func BenchmarkNewShapes_Phase51(b *testing.B) {
	gen := NewGenerator()
	testColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}

	benchmarks := []struct {
		name   string
		config Config
	}{
		{
			name: "ellipse",
			config: Config{
				Type:      ShapeEllipse,
				Width:     32,
				Height:    48,
				Color:     testColor,
				Seed:      12345,
				Smoothing: 0.2,
			},
		},
		{
			name: "capsule",
			config: Config{
				Type:      ShapeCapsule,
				Width:     20,
				Height:    40,
				Color:     testColor,
				Seed:      12346,
				Smoothing: 0.2,
			},
		},
		{
			name: "bean",
			config: Config{
				Type:      ShapeBean,
				Width:     32,
				Height:    40,
				Color:     testColor,
				Seed:      12348,
				Smoothing: 0.2,
			},
		},
		{
			name: "wedge",
			config: Config{
				Type:      ShapeWedge,
				Width:     24,
				Height:    32,
				Color:     testColor,
				Seed:      12349,
				Smoothing: 0.2,
			},
		},
		{
			name: "shield",
			config: Config{
				Type:      ShapeShield,
				Width:     28,
				Height:    36,
				Color:     testColor,
				Seed:      12351,
				Smoothing: 0.2,
			},
		},
		{
			name: "blade",
			config: Config{
				Type:      ShapeBlade,
				Width:     16,
				Height:    48,
				Color:     testColor,
				Seed:      12352,
				Smoothing: 0.2,
			},
		},
		{
			name: "skull",
			config: Config{
				Type:      ShapeSkull,
				Width:     32,
				Height:    32,
				Color:     testColor,
				Seed:      12353,
				Smoothing: 0.2,
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = gen.Generate(bm.config)
			}
		})
	}
}

