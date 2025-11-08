package ui

import (
	"image"
	"testing"
)

// TestGenerateDecorativeFrame tests frame generation for all styles.
func TestGenerateDecorativeFrame(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		style   FrameStyle
		genreID string
		width   int
		height  int
		wantErr bool
	}{
		{"OrnateCorners fantasy", FrameOrnateCorners, "fantasy", 100, 100, false},
		{"TechAngular scifi", FrameTechAngular, "scifi", 100, 100, false},
		{"Weathered postapoc", FrameWeathered, "postapoc", 100, 100, false},
		{"Auto fantasy", FrameAuto, "fantasy", 100, 100, false},
		{"Auto cyberpunk", FrameAuto, "cyberpunk", 100, 100, false},
		{"Invalid size", FrameOrnateCorners, "fantasy", 0, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Type:    ElementFrame,
				Width:   tt.width,
				Height:  tt.height,
				GenreID: tt.genreID,
				Seed:    12345,
				State:   StateNormal,
			}

			img, err := gen.GenerateDecorativeFrame(config, tt.style)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDecorativeFrame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("GenerateDecorativeFrame() returned nil image")
					return
				}
				bounds := img.Bounds()
				if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
					t.Errorf("Image dimensions = %dx%d, want %dx%d",
						bounds.Dx(), bounds.Dy(), tt.width, tt.height)
				}
			}
		})
	}
}

// TestGenerateDecorativeFrame_States tests state-based visual effects.
func TestGenerateDecorativeFrame_States(t *testing.T) {
	gen := NewGenerator()

	states := []ElementState{StateNormal, StateHover, StatePressed, StateDisabled}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			config := Config{
				Type:    ElementFrame,
				Width:   100,
				Height:  100,
				GenreID: "fantasy",
				Seed:    12345,
				State:   state,
			}

			img, err := gen.GenerateDecorativeFrame(config, FrameOrnateCorners)
			if err != nil {
				t.Errorf("GenerateDecorativeFrame() error = %v", err)
				return
			}

			if img == nil {
				t.Error("GenerateDecorativeFrame() returned nil image")
			}
		})
	}
}

// TestGenerateSymbol tests symbol generation for all icon types.
func TestGenerateSymbol(t *testing.T) {
	gen := NewGenerator()

	symbols := []IconSymbol{
		IconSword, IconShield, IconPotion, IconCoin,
		IconHeart, IconStar, IconGear,
		IconCheckmark, IconX,
		IconArrowUp, IconArrowDown, IconArrowLeft, IconArrowRight,
	}

	for _, symbol := range symbols {
		t.Run(symbol.String(), func(t *testing.T) {
			config := Config{
				Type:    ElementIcon,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			}

			img, err := gen.GenerateSymbol(config, symbol)
			if err != nil {
				t.Errorf("GenerateSymbol(%v) error = %v", symbol, err)
				return
			}

			if img == nil {
				t.Errorf("GenerateSymbol(%v) returned nil image", symbol)
				return
			}

			bounds := img.Bounds()
			if bounds.Dx() != 32 || bounds.Dy() != 32 {
				t.Errorf("Image dimensions = %dx%d, want 32x32",
					bounds.Dx(), bounds.Dy())
			}

			// Verify at least some pixels are set (not all transparent)
			hasContent := false
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					if a > 0 && (r > 0 || g > 0 || b > 0) {
						hasContent = true
						break
					}
				}
				if hasContent {
					break
				}
			}

			if !hasContent {
				t.Errorf("GenerateSymbol(%v) generated empty image", symbol)
			}
		})
	}
}

// TestGenerateSymbol_Deterministic tests that symbols are deterministic.
func TestGenerateSymbol_Deterministic(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:    ElementIcon,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    99999,
	}

	// Generate twice with same seed
	img1, err1 := gen.GenerateSymbol(config, IconStar)
	img2, err2 := gen.GenerateSymbol(config, IconStar)

	if err1 != nil || err2 != nil {
		t.Fatalf("Errors generating symbols: %v, %v", err1, err2)
	}

	// Compare images pixel by pixel
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)

			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Errorf("Pixel at (%d,%d) differs: (%d,%d,%d,%d) vs (%d,%d,%d,%d)",
					x, y, r1, g1, b1, a1, r2, g2, b2, a2)
				return
			}
		}
	}
}

// TestGeneratePanelWithPattern tests panel generation with patterns.
func TestGeneratePanelWithPattern(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		genreID string
		width   int
		height  int
		wantErr bool
	}{
		{"fantasy panel", "fantasy", 200, 150, false},
		{"scifi panel", "scifi", 200, 150, false},
		{"horror panel", "horror", 200, 150, false},
		{"cyberpunk panel", "cyberpunk", 200, 150, false},
		{"postapoc panel", "postapoc", 200, 150, false},
		{"invalid size", "fantasy", 0, 150, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Type:    ElementPanel,
				Width:   tt.width,
				Height:  tt.height,
				GenreID: tt.genreID,
				Seed:    54321,
				State:   StateNormal,
			}

			img, err := gen.GeneratePanelWithPattern(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePanelWithPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("GeneratePanelWithPattern() returned nil image")
					return
				}

				bounds := img.Bounds()
				if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
					t.Errorf("Image dimensions = %dx%d, want %dx%d",
						bounds.Dx(), bounds.Dy(), tt.width, tt.height)
				}
			}
		})
	}
}

// TestFrameStyle_String tests FrameStyle string conversion.
func TestFrameStyle_String(t *testing.T) {
	tests := []struct {
		style FrameStyle
		want  string
	}{
		{FrameAuto, "auto"},
		{FrameOrnateCorners, "ornate-corners"},
		{FrameTechAngular, "tech-angular"},
		{FrameWeathered, "weathered"},
		{FrameCircuitPattern, "circuit-pattern"},
		{FrameTribalPattern, "tribal-pattern"},
		{FrameStyle(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.style.String()
			if got != tt.want {
				t.Errorf("FrameStyle.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIconSymbol_String tests IconSymbol string conversion.
func TestIconSymbol_String(t *testing.T) {
	tests := []struct {
		symbol IconSymbol
		want   string
	}{
		{IconSword, "sword"},
		{IconShield, "shield"},
		{IconPotion, "potion"},
		{IconCoin, "coin"},
		{IconHeart, "heart"},
		{IconStar, "star"},
		{IconGear, "gear"},
		{IconCheckmark, "checkmark"},
		{IconX, "x"},
		{IconArrowUp, "arrow-up"},
		{IconArrowDown, "arrow-down"},
		{IconArrowLeft, "arrow-left"},
		{IconArrowRight, "arrow-right"},
		{IconSymbol(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.symbol.String()
			if got != tt.want {
				t.Errorf("IconSymbol.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMinInt tests the minInt helper function.
func TestMinInt(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 7, 7, 7},
		{"negative", -5, 3, -5},
		{"both negative", -10, -5, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minInt(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// Benchmark tests

func BenchmarkGenerateDecorativeFrame(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementFrame,
		Width:   100,
		Height:  100,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateDecorativeFrame(config, FrameOrnateCorners)
	}
}

func BenchmarkGenerateSymbol(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementIcon,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateSymbol(config, IconStar)
	}
}

func BenchmarkGeneratePanelWithPattern(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementPanel,
		Width:   200,
		Height:  150,
		GenreID: "fantasy",
		Seed:    54321,
		State:   StateNormal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GeneratePanelWithPattern(config)
	}
}

// Helper function for test validation
func hasNonTransparentPixels(img *image.RGBA) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				return true
			}
		}
	}
	return false
}
