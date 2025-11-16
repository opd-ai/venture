package tiles

import (
	"image"
	"image/color"
	"testing"
)

func TestCornerType_String(t *testing.T) {
	tests := []struct {
		name     string
		corner   CornerType
		expected string
	}{
		{"none", CornerNone, "none"},
		{"L corner", CornerL, "L"},
		{"T junction", CornerT, "T"},
		{"cross junction", CornerCross, "cross"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.corner.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWallNeighbors_DetectCornerType(t *testing.T) {
	tests := []struct {
		name      string
		neighbors WallNeighbors
		expected  CornerType
	}{
		{
			name:      "no neighbors",
			neighbors: WallNeighbors{},
			expected:  CornerNone,
		},
		{
			name:      "single neighbor",
			neighbors: WallNeighbors{North: true},
			expected:  CornerNone,
		},
		{
			name:      "opposite neighbors (corridor)",
			neighbors: WallNeighbors{North: true, South: true},
			expected:  CornerNone,
		},
		{
			name:      "L corner NE",
			neighbors: WallNeighbors{North: true, East: true},
			expected:  CornerL,
		},
		{
			name:      "L corner SE",
			neighbors: WallNeighbors{South: true, East: true},
			expected:  CornerL,
		},
		{
			name:      "L corner SW",
			neighbors: WallNeighbors{South: true, West: true},
			expected:  CornerL,
		},
		{
			name:      "L corner NW",
			neighbors: WallNeighbors{North: true, West: true},
			expected:  CornerL,
		},
		{
			name:      "T junction (missing south)",
			neighbors: WallNeighbors{North: true, East: true, West: true},
			expected:  CornerT,
		},
		{
			name:      "T junction (missing north)",
			neighbors: WallNeighbors{South: true, East: true, West: true},
			expected:  CornerT,
		},
		{
			name:      "T junction (missing east)",
			neighbors: WallNeighbors{North: true, South: true, West: true},
			expected:  CornerT,
		},
		{
			name:      "T junction (missing west)",
			neighbors: WallNeighbors{North: true, South: true, East: true},
			expected:  CornerT,
		},
		{
			name:      "cross junction",
			neighbors: WallNeighbors{North: true, South: true, East: true, West: true},
			expected:  CornerCross,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.neighbors.DetectCornerType(); got != tt.expected {
				t.Errorf("DetectCornerType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultEnhancedWallConfig(t *testing.T) {
	config := DefaultEnhancedWallConfig()

	if config.EnableAntialiasing != true {
		t.Errorf("EnableAntialiasing = %v, want true", config.EnableAntialiasing)
	}
	if config.EnableShadows != true {
		t.Errorf("EnableShadows = %v, want true", config.EnableShadows)
	}
	if config.BlendRadius != 4 {
		t.Errorf("BlendRadius = %v, want 4", config.BlendRadius)
	}
}

func TestGenerateEnhancedWall_BasicGeneration(t *testing.T) {
	g := NewGenerator()

	config := DefaultEnhancedWallConfig()
	config.Config.Width = 32
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableAntialiasing = true

	img, err := g.GenerateEnhancedWall(config)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() error = %v", err)
	}

	if img == nil {
		t.Fatal("GenerateEnhancedWall() returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Image dimensions = %dx%d, want 32x32", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateEnhancedWall_WithoutAntialiasing(t *testing.T) {
	g := NewGenerator()

	config := DefaultEnhancedWallConfig()
	config.Config.Width = 32
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableAntialiasing = false

	img, err := g.GenerateEnhancedWall(config)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() error = %v", err)
	}

	if img == nil {
		t.Fatal("GenerateEnhancedWall() returned nil image")
	}
}

func TestGenerateEnhancedWall_WithCorners(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		name      string
		neighbors WallNeighbors
	}{
		{"L corner NE", WallNeighbors{North: true, East: true}},
		{"L corner SE", WallNeighbors{South: true, East: true}},
		{"T junction north", WallNeighbors{North: true, East: true, West: true}},
		{"cross junction", WallNeighbors{North: true, South: true, East: true, West: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultEnhancedWallConfig()
			config.Config.Width = 32
			config.Config.Height = 32
			config.Config.Seed = 12345
			config.Config.GenreID = "fantasy"
			config.Neighbors = tt.neighbors

			img, err := g.GenerateEnhancedWall(config)
			if err != nil {
				t.Fatalf("GenerateEnhancedWall() error = %v", err)
			}

			if img == nil {
				t.Fatal("GenerateEnhancedWall() returned nil image")
			}
		})
	}
}

func TestGenerateEnhancedWall_Deterministic(t *testing.T) {
	g := NewGenerator()
	seed := int64(42)

	config1 := DefaultEnhancedWallConfig()
	config1.Config.Width = 32
	config1.Config.Height = 32
	config1.Config.Seed = seed
	config1.Config.GenreID = "fantasy"
	config1.Neighbors = WallNeighbors{North: true, East: true}

	config2 := DefaultEnhancedWallConfig()
	config2.Config.Width = 32
	config2.Config.Height = 32
	config2.Config.Seed = seed
	config2.Config.GenreID = "fantasy"
	config2.Neighbors = WallNeighbors{North: true, East: true}

	img1, err := g.GenerateEnhancedWall(config1)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() first call error = %v", err)
	}

	img2, err := g.GenerateEnhancedWall(config2)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() second call error = %v", err)
	}

	// Compare images pixel by pixel
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)

			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Errorf("Pixel at (%d, %d) differs: got (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					x, y, r1, g1, b1, a1, r2, g2, b2, a2)
				return
			}
		}
	}
}

func TestGenerateEnhancedWall_InvalidConfig(t *testing.T) {
	g := NewGenerator()

	config := DefaultEnhancedWallConfig()
	config.Config.Width = -1 // Invalid
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"

	_, err := g.GenerateEnhancedWall(config)
	if err == nil {
		t.Error("GenerateEnhancedWall() expected error for invalid width, got nil")
	}
}

func TestGenerateEnhancedWall_HighResolution(t *testing.T) {
	g := NewGenerator()

	config := DefaultEnhancedWallConfig()
	config.Config.Width = 64
	config.Config.Height = 64
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableAntialiasing = true

	img, err := g.GenerateEnhancedWall(config)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() error = %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Image dimensions = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}
}

func TestDownsample2x2(t *testing.T) {
	g := NewGenerator()

	// Create a 4x4 test image with known colors
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	testColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}

	// Fill with test color
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src.Set(x, y, testColor)
		}
	}

	// Downsample to 2x2
	dst := g.downsample2x2(src, 2, 2)

	if dst == nil {
		t.Fatal("downsample2x2() returned nil")
	}

	bounds := dst.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Errorf("Downsampled dimensions = %dx%d, want 2x2", bounds.Dx(), bounds.Dy())
	}

	// Check that downsampled pixels match (averaging same color = same color)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			c := dst.At(x, y)
			r, g, b, a := c.RGBA()

			// Allow small rounding error
			if abs(int(r>>8)-100) > 1 || abs(int(g>>8)-150) > 1 || abs(int(b>>8)-200) > 1 {
				t.Errorf("Pixel at (%d,%d) = (%d,%d,%d,%d), want (~100,~150,~200,255)",
					x, y, r>>8, g>>8, b>>8, a>>8)
			}
		}
	}
}

func TestBlendCircularArea(t *testing.T) {
	g := NewGenerator()

	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	blendColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	// Fill with base color
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, baseColor)
		}
	}

	// Blend circular area at center
	g.blendCircularArea(img, 16, 16, 4, blendColor)

	// Check that center pixel is blended
	centerColor := img.At(16, 16)
	r, _, _, _ := centerColor.RGBA()

	// Center should be lighter than base (blended toward blendColor)
	if r>>8 <= 100 {
		t.Errorf("Center pixel not blended: R=%d, expected > 100", r>>8)
	}

	// Check that far corner is not affected
	cornerColor := img.At(0, 0)
	cr, _, _, _ := cornerColor.RGBA()
	if cr>>8 != 100 {
		t.Errorf("Corner pixel affected: R=%d, expected 100", cr>>8)
	}
}

func TestGenerateEnhancedWall_WithShadows(t *testing.T) {
	g := NewGenerator()

	config := DefaultEnhancedWallConfig()
	config.Config.Width = 32
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableShadows = true

	img, err := g.GenerateEnhancedWall(config)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() error = %v", err)
	}

	if img == nil {
		t.Fatal("GenerateEnhancedWall() returned nil image")
	}

	// Verify shadow gradient: bottom should be darker than top
	topColor := img.At(16, 2)
	bottomColor := img.At(16, 30)

	tr, tg, tb, _ := topColor.RGBA()
	br, bg, bb, _ := bottomColor.RGBA()

	topBrightness := int(tr>>8) + int(tg>>8) + int(tb>>8)
	bottomBrightness := int(br>>8) + int(bg>>8) + int(bb>>8)

	// Shadow should make bottom slightly darker (or at least not brighter)
	if bottomBrightness > topBrightness {
		t.Errorf("Shadow gradient reversed: bottom brightness %d > top brightness %d",
			bottomBrightness, topBrightness)
	}
}

func TestGenerateEnhancedWall_WithoutShadows(t *testing.T) {
	g := NewGenerator()

	config := DefaultEnhancedWallConfig()
	config.Config.Width = 32
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableShadows = false

	img, err := g.GenerateEnhancedWall(config)
	if err != nil {
		t.Fatalf("GenerateEnhancedWall() error = %v", err)
	}

	if img == nil {
		t.Fatal("GenerateEnhancedWall() returned nil image")
	}
}

func TestGenerateEnhancedWall_MultipleGenres(t *testing.T) {
	g := NewGenerator()

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := DefaultEnhancedWallConfig()
			config.Config.Width = 32
			config.Config.Height = 32
			config.Config.Seed = 12345
			config.Config.GenreID = genre

			img, err := g.GenerateEnhancedWall(config)
			if err != nil {
				t.Fatalf("GenerateEnhancedWall() error = %v for genre %s", err, genre)
			}

			if img == nil {
				t.Fatalf("GenerateEnhancedWall() returned nil image for genre %s", genre)
			}
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests

func BenchmarkGenerateEnhancedWall_NoAA(b *testing.B) {
	g := NewGenerator()
	config := DefaultEnhancedWallConfig()
	config.Config.Width = 32
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableAntialiasing = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateEnhancedWall(config)
	}
}

func BenchmarkGenerateEnhancedWall_WithAA(b *testing.B) {
	g := NewGenerator()
	config := DefaultEnhancedWallConfig()
	config.Config.Width = 32
	config.Config.Height = 32
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableAntialiasing = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateEnhancedWall(config)
	}
}

func BenchmarkGenerateEnhancedWall_LargeWithAA(b *testing.B) {
	g := NewGenerator()
	config := DefaultEnhancedWallConfig()
	config.Config.Width = 64
	config.Config.Height = 64
	config.Config.Seed = 12345
	config.Config.GenreID = "fantasy"
	config.EnableAntialiasing = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateEnhancedWall(config)
	}
}

func BenchmarkDownsample2x2(b *testing.B) {
	g := NewGenerator()
	src := image.NewRGBA(image.Rect(0, 0, 64, 64))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.downsample2x2(src, 32, 32)
	}
}

func BenchmarkBlendCircularArea(b *testing.B) {
	g := NewGenerator()
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	blendColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.blendCircularArea(img, 16, 16, 4, blendColor)
	}
}
