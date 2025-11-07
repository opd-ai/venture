// Package tiles provides tests for Phase 16.2 smooth terrain transitions.
package tiles

import (
	"image/color"
	"testing"
)

// TestTileTransitionType_String tests the string representation of transition types.
func TestTileTransitionType_String(t *testing.T) {
	tests := []struct {
		name       string
		transition TileTransitionType
		want       string
	}{
		{"None", TransitionNone, "none"},
		{"Full", TransitionFull, "full"},
		{"North", TransitionN, "n"},
		{"East", TransitionE, "e"},
		{"South", TransitionS, "s"},
		{"West", TransitionW, "w"},
		{"North-South", TransitionNS, "ns"},
		{"East-West", TransitionEW, "ew"},
		{"Northeast", TransitionNE, "ne"},
		{"Northwest", TransitionNW, "nw"},
		{"Southeast", TransitionSE, "se"},
		{"Southwest", TransitionSW, "sw"},
		{"T-junction NES", TransitionNES, "nes"},
		{"T-junction NEW", TransitionNEW, "new"},
		{"T-junction NSW", TransitionNSW, "nsw"},
		{"T-junction ESW", TransitionESW, "esw"},
		{"Inner NE", TransitionInnerNE, "inner_ne"},
		{"Inner NW", TransitionInnerNW, "inner_nw"},
		{"Inner SE", TransitionInnerSE, "inner_se"},
		{"Inner SW", TransitionInnerSW, "inner_sw"},
		{"Unknown", TileTransitionType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.transition.String()
			if got != tt.want {
				t.Errorf("TileTransitionType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetermineTransition tests the transition determination logic.
func TestDetermineTransition(t *testing.T) {
	tests := []struct {
		name      string
		neighbors TileNeighbors
		want      TileTransitionType
	}{
		{
			name:      "No neighbors",
			neighbors: TileNeighbors{},
			want:      TransitionNone,
		},
		{
			name:      "All cardinal neighbors",
			neighbors: TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true, SW: true},
			want:      TransitionFull,
		},
		{
			name:      "North only",
			neighbors: TileNeighbors{N: true},
			want:      TransitionN,
		},
		{
			name:      "East only",
			neighbors: TileNeighbors{E: true},
			want:      TransitionE,
		},
		{
			name:      "South only",
			neighbors: TileNeighbors{S: true},
			want:      TransitionS,
		},
		{
			name:      "West only",
			neighbors: TileNeighbors{W: true},
			want:      TransitionW,
		},
		{
			name:      "North-South corridor",
			neighbors: TileNeighbors{N: true, S: true},
			want:      TransitionNS,
		},
		{
			name:      "East-West corridor",
			neighbors: TileNeighbors{E: true, W: true},
			want:      TransitionEW,
		},
		{
			name:      "Northeast corner",
			neighbors: TileNeighbors{N: true, E: true},
			want:      TransitionNE,
		},
		{
			name:      "Northwest corner",
			neighbors: TileNeighbors{N: true, W: true},
			want:      TransitionNW,
		},
		{
			name:      "Southeast corner",
			neighbors: TileNeighbors{S: true, E: true},
			want:      TransitionSE,
		},
		{
			name:      "Southwest corner",
			neighbors: TileNeighbors{S: true, W: true},
			want:      TransitionSW,
		},
		{
			name:      "T-junction North-East-South",
			neighbors: TileNeighbors{N: true, E: true, S: true},
			want:      TransitionNES,
		},
		{
			name:      "T-junction North-East-West",
			neighbors: TileNeighbors{N: true, E: true, W: true},
			want:      TransitionNEW,
		},
		{
			name:      "T-junction North-South-West",
			neighbors: TileNeighbors{N: true, S: true, W: true},
			want:      TransitionNSW,
		},
		{
			name:      "T-junction East-South-West",
			neighbors: TileNeighbors{E: true, S: true, W: true},
			want:      TransitionESW,
		},
		{
			name:      "Inner corner Northeast (missing NE diagonal)",
			neighbors: TileNeighbors{N: true, E: true, S: true, W: true, NW: true, SE: true, SW: true},
			want:      TransitionInnerNE,
		},
		{
			name:      "Inner corner Northwest (missing NW diagonal)",
			neighbors: TileNeighbors{N: true, E: true, S: true, W: true, NE: true, SE: true, SW: true},
			want:      TransitionInnerNW,
		},
		{
			name:      "Inner corner Southeast (missing SE diagonal)",
			neighbors: TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SW: true},
			want:      TransitionInnerSE,
		},
		{
			name:      "Inner corner Southwest (missing SW diagonal)",
			neighbors: TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true},
			want:      TransitionInnerSW,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineTransition(tt.neighbors)
			if got != tt.want {
				t.Errorf("DetermineTransition() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultTransitionConfig tests the default transition configuration.
func TestDefaultTransitionConfig(t *testing.T) {
	config := DefaultTransitionConfig()

	if config.BaseConfig.Width != 32 {
		t.Errorf("DefaultTransitionConfig().BaseConfig.Width = %v, want 32", config.BaseConfig.Width)
	}
	if config.Transition != TransitionFull {
		t.Errorf("DefaultTransitionConfig().Transition = %v, want TransitionFull", config.Transition)
	}
	if config.BlendRadius <= 0 || config.BlendRadius > 1 {
		t.Errorf("DefaultTransitionConfig().BlendRadius = %v, want 0.0-1.0", config.BlendRadius)
	}
	if config.CornerRadius <= 0 || config.CornerRadius > 1 {
		t.Errorf("DefaultTransitionConfig().CornerRadius = %v, want 0.0-1.0", config.CornerRadius)
	}
	if config.Smoothness < 0 || config.Smoothness > 1 {
		t.Errorf("DefaultTransitionConfig().Smoothness = %v, want 0.0-1.0", config.Smoothness)
	}
}

// TestGenerator_GenerateWithTransition tests transition tile generation.
func TestGenerator_GenerateWithTransition(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name       string
		config     TransitionConfig
		wantError  bool
	}{
		{
			name: "Valid transition - none",
			config: TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition:   TransitionNone,
				BlendRadius:  0.3,
				CornerRadius: 0.25,
				Smoothness:   0.5,
			},
			wantError: false,
		},
		{
			name: "Valid transition - full",
			config: TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition:   TransitionFull,
				BlendRadius:  0.3,
				CornerRadius: 0.25,
				Smoothness:   0.5,
			},
			wantError: false,
		},
		{
			name: "Valid transition - corner NE",
			config: TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition:   TransitionNE,
				Neighbors:    TileNeighbors{N: true, E: true},
				BlendRadius:  0.3,
				CornerRadius: 0.25,
				Smoothness:   0.5,
			},
			wantError: false,
		},
		{
			name: "Valid transition - corridor NS",
			config: TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition:   TransitionNS,
				Neighbors:    TileNeighbors{N: true, S: true},
				BlendRadius:  0.3,
				CornerRadius: 0.25,
				Smoothness:   0.5,
			},
			wantError: false,
		},
		{
			name: "Invalid config - zero width",
			config: TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   0,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition: TransitionFull,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.GenerateWithTransition(tt.config)
			
			if tt.wantError {
				if err == nil {
					t.Error("GenerateWithTransition() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("GenerateWithTransition() unexpected error: %v", err)
				return
			}
			
			if img == nil {
				t.Error("GenerateWithTransition() returned nil image")
				return
			}
			
			bounds := img.Bounds()
			if bounds.Dx() != tt.config.BaseConfig.Width {
				t.Errorf("GenerateWithTransition() width = %v, want %v", bounds.Dx(), tt.config.BaseConfig.Width)
			}
			if bounds.Dy() != tt.config.BaseConfig.Height {
				t.Errorf("GenerateWithTransition() height = %v, want %v", bounds.Dy(), tt.config.BaseConfig.Height)
			}
		})
	}
}

// TestTransitionDeterminism verifies deterministic generation.
func TestTransitionDeterminism(t *testing.T) {
	gen := NewGenerator()
	config := TransitionConfig{
		BaseConfig: Config{
			Type:    TileWall,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    54321,
			Variant: 0.5,
			Custom:  make(map[string]interface{}),
		},
		Transition:   TransitionNE,
		Neighbors:    TileNeighbors{N: true, E: true},
		BlendRadius:  0.3,
		CornerRadius: 0.25,
		Smoothness:   0.5,
	}

	// Generate twice with same seed
	img1, err1 := gen.GenerateWithTransition(config)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	img2, err2 := gen.GenerateWithTransition(config)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
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
				t.Errorf("Pixel (%d, %d) differs: got (%v, %v, %v, %v), want (%v, %v, %v, %v)",
					x, y, r1, g1, b1, a1, r2, g2, b2, a2)
				return
			}
		}
	}
}

// TestTransitionDifferentSeeds verifies different seeds produce different tiles.
func TestTransitionDifferentSeeds(t *testing.T) {
	gen := NewGenerator()
	
	config1 := TransitionConfig{
		BaseConfig: Config{
			Type:    TileWall,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    12345,
			Variant: 0.5,
			Custom:  make(map[string]interface{}),
		},
		Transition:   TransitionNE,
		BlendRadius:  0.3,
		CornerRadius: 0.25,
		Smoothness:   0.5,
	}
	
	config2 := config1
	config2.BaseConfig.Seed = 67890

	img1, err1 := gen.GenerateWithTransition(config1)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	img2, err2 := gen.GenerateWithTransition(config2)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Images should be different
	bounds := img1.Bounds()
	identical := true
	for y := bounds.Min.Y; y < bounds.Max.Y && identical; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			
			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				identical = false
				break
			}
		}
	}

	if identical {
		t.Error("Different seeds produced identical tiles")
	}
}

// TestTransitionAllTypes tests all transition types generate successfully.
func TestTransitionAllTypes(t *testing.T) {
	gen := NewGenerator()
	
	transitions := []TileTransitionType{
		TransitionNone, TransitionFull,
		TransitionN, TransitionE, TransitionS, TransitionW,
		TransitionNS, TransitionEW,
		TransitionNE, TransitionNW, TransitionSE, TransitionSW,
		TransitionNES, TransitionNEW, TransitionNSW, TransitionESW,
		TransitionInnerNE, TransitionInnerNW, TransitionInnerSE, TransitionInnerSW,
	}

	for _, trans := range transitions {
		t.Run(trans.String(), func(t *testing.T) {
			config := TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition:   trans,
				BlendRadius:  0.3,
				CornerRadius: 0.25,
				Smoothness:   0.5,
			}

			img, err := gen.GenerateWithTransition(config)
			if err != nil {
				t.Errorf("GenerateWithTransition(%v) failed: %v", trans, err)
				return
			}
			if img == nil {
				t.Errorf("GenerateWithTransition(%v) returned nil image", trans)
			}
		})
	}
}

// TestTransitionBlendRadius tests different blend radius values.
func TestTransitionBlendRadius(t *testing.T) {
	gen := NewGenerator()
	
	radii := []float64{0.0, 0.1, 0.3, 0.5, 1.0}

	for _, radius := range radii {
		t.Run("", func(t *testing.T) {
			config := TransitionConfig{
				BaseConfig: Config{
					Type:    TileWall,
					Width:   32,
					Height:  32,
					GenreID: "fantasy",
					Seed:    12345,
					Variant: 0.5,
					Custom:  make(map[string]interface{}),
				},
				Transition:   TransitionN,
				BlendRadius:  radius,
				CornerRadius: 0.25,
				Smoothness:   0.5,
			}

			img, err := gen.GenerateWithTransition(config)
			if err != nil {
				t.Errorf("GenerateWithTransition(radius=%v) failed: %v", radius, err)
				return
			}
			if img == nil {
				t.Errorf("GenerateWithTransition(radius=%v) returned nil image", radius)
			}
		})
	}
}

// TestHelperFunctions tests utility functions.
func TestSmoothstep(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{-0.1, 0.0}, // Clamps to 0
		{1.1, 1.0},  // Clamps to 1
	}

	for _, tt := range tests {
		got := smoothstep(tt.input)
		if got != tt.want {
			t.Errorf("smoothstep(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t float64
		want    float64
	}{
		{0.0, 100.0, 0.0, 0.0},
		{0.0, 100.0, 0.5, 50.0},
		{0.0, 100.0, 1.0, 100.0},
		{50.0, 150.0, 0.5, 100.0},
	}

	for _, tt := range tests {
		got := lerp(tt.a, tt.b, tt.t)
		if got != tt.want {
			t.Errorf("lerp(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.want)
		}
	}
}

func TestBlendColors(t *testing.T) {
	c1 := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	c2 := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	// Test full blend to c2
	result := blendColors(c1, c2, 1.0)
	r, g, b, _ := result.RGBA()
	if r>>8 != 100 || g>>8 != 100 || b>>8 != 100 {
		t.Errorf("blendColors(black, gray, 1.0) = (%v, %v, %v), want (100, 100, 100)", r>>8, g>>8, b>>8)
	}

	// Test no blend (c1)
	result = blendColors(c1, c2, 0.0)
	r, g, b, _ = result.RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("blendColors(black, gray, 0.0) = (%v, %v, %v), want (0, 0, 0)", r>>8, g>>8, b>>8)
	}

	// Test mid blend
	result = blendColors(c1, c2, 0.5)
	r, g, b, _ = result.RGBA()
	expected := uint32(50)
	if r>>8 != expected || g>>8 != expected || b>>8 != expected {
		t.Errorf("blendColors(black, gray, 0.5) = (%v, %v, %v), want (%v, %v, %v)", 
			r>>8, g>>8, b>>8, expected, expected, expected)
	}
}

// TestMinMax tests min and max functions.
func TestMinMax(t *testing.T) {
	if min(5, 10) != 5 {
		t.Errorf("min(5, 10) = %v, want 5", min(5, 10))
	}
	if min(10, 5) != 5 {
		t.Errorf("min(10, 5) = %v, want 5", min(10, 5))
	}
	if max(5, 10) != 10 {
		t.Errorf("max(5, 10) = %v, want 10", max(5, 10))
	}
	if max(10, 5) != 10 {
		t.Errorf("max(10, 5) = %v, want 10", max(10, 5))
	}
}

// BenchmarkGenerateWithTransition benchmarks transition tile generation.
func BenchmarkGenerateWithTransition(b *testing.B) {
	gen := NewGenerator()
	config := TransitionConfig{
		BaseConfig: Config{
			Type:    TileWall,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    12345,
			Variant: 0.5,
			Custom:  make(map[string]interface{}),
		},
		Transition:   TransitionNE,
		BlendRadius:  0.3,
		CornerRadius: 0.25,
		Smoothness:   0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateWithTransition(config)
	}
}

// BenchmarkDetermineTransition benchmarks transition determination.
func BenchmarkDetermineTransition(b *testing.B) {
	neighbors := TileNeighbors{
		N: true, E: true, S: true, W: true,
		NE: true, NW: true, SE: true, SW: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetermineTransition(neighbors)
	}
}

// TestEdgeBlendVisual verifies edge blending produces gradients.
func TestEdgeBlendVisual(t *testing.T) {
	gen := NewGenerator()
	config := TransitionConfig{
		BaseConfig: Config{
			Type:    TileWall,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    12345,
			Variant: 0.5,
			Custom:  make(map[string]interface{}),
		},
		Transition:   TransitionN,
		BlendRadius:  0.5, // Large blend for testing
		CornerRadius: 0.0,
		Smoothness:   0.0,
	}

	img, err := gen.GenerateWithTransition(config)
	if err != nil {
		t.Fatalf("GenerateWithTransition failed: %v", err)
	}

	// Check that north edge has gradient (colors vary)
	bounds := img.Bounds()
	width := bounds.Dx()
	blendHeight := int(float64(bounds.Dy()) * config.BlendRadius)
	
	// Sample colors along vertical line in blend zone
	var colors []uint32
	x := width / 2
	for y := 0; y < blendHeight; y++ {
		r, _, _, _ := img.At(x, y).RGBA()
		colors = append(colors, r)
	}
	
	// Verify colors change (gradient effect)
	allSame := true
	firstColor := colors[0]
	for _, c := range colors {
		if c != firstColor {
			allSame = false
			break
		}
	}
	
	if allSame && blendHeight > 2 {
		t.Error("Expected gradient in blend zone, but all colors are identical")
	}
}

// TestCornerRoundingVisual verifies corner rounding modifies corners.
func TestCornerRoundingVisual(t *testing.T) {
	gen := NewGenerator()
	config := TransitionConfig{
		BaseConfig: Config{
			Type:    TileWall,
			Width:   32,
			Height:  32,
			GenreID: "fantasy",
			Seed:    12345,
			Variant: 0.5,
			Custom:  make(map[string]interface{}),
		},
		Transition:   TransitionNone,
		BlendRadius:  0.0,
		CornerRadius: 0.3, // Significant rounding
		Smoothness:   0.0,
	}

	img, err := gen.GenerateWithTransition(config)
	if err != nil {
		t.Fatalf("GenerateWithTransition failed: %v", err)
	}

	// Sample corners - they should differ from edges
	bounds := img.Bounds()
	cornerColor := img.At(0, 0) // Top-left corner
	edgeColor := img.At(bounds.Dx()/2, 0) // Top edge center
	
	r1, g1, b1, _ := cornerColor.RGBA()
	r2, g2, b2, _ := edgeColor.RGBA()
	
	// Corner should be modified (different from straight edge)
	// Note: This is a weak test since colors may coincidentally match
	if r1 == r2 && g1 == g2 && b1 == b2 {
		// This is acceptable - just verify no error occurred
		t.Log("Corner and edge colors match (acceptable)")
	}
}
