// Package engine provides tests for genre-based UI color palettes.
// Gap #11 fix: Verify dynamic color palette application to UI components.
package engine

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// TestInventoryUI_GenrePaletteIntegration verifies inventory UI uses genre-appropriate colors.
func TestInventoryUI_GenrePaletteIntegration(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    int64
	}{
		{"fantasy palette", "fantasy", 12345},
		{"scifi palette", "scifi", 67890},
		{"cyberpunk palette", "cyberpunk", 11111},
		{"horror palette", "horror", 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test palette generation directly (avoids Ebiten initialization)
			paletteGen := palette.NewGenerator()
			p, err := paletteGen.Generate(tt.genreID, tt.seed)
			if err != nil {
				t.Fatalf("Failed to generate palette for genre %s: %v", tt.genreID, err)
			}

			// Verify palette has required colors
			if p.Primary == nil {
				t.Error("Expected Primary color in palette")
			}
			if p.Background == nil {
				t.Error("Expected Background color in palette")
			}
			if p.Shadow1 == nil {
				t.Error("Expected Shadow1 color in palette")
			}
		})
	}
}

// TestInventoryUI_PaletteColorRetrieval verifies color helper methods work correctly.
func TestInventoryUI_PaletteColorRetrieval(t *testing.T) {
	// Create a test palette manually
	testPalette := &palette.Palette{
		Primary:    testColor(100, 150, 200, 255),
		Secondary:  testColor(80, 80, 100, 255),
		Background: testColor(40, 40, 50, 255),
		Shadow1:    testColor(0, 0, 0, 180),
	}

	// Create UI struct with test palette (no Ebiten initialization)
	ui := &EbitenInventoryUI{
		uiPalette: testPalette,
	}

	// Test overlay color
	overlayColor := ui.getOverlayColor()
	if overlayColor == nil {
		t.Fatal("Expected overlay color, got nil")
	}

	// Test background color
	bgColor := ui.getBackgroundColor()
	if bgColor == nil {
		t.Fatal("Expected background color, got nil")
	}
}

// TestCharacterUI_PaletteColorRetrieval verifies character UI color helpers.
func TestCharacterUI_PaletteColorRetrieval(t *testing.T) {
	// Create a test palette manually
	testPalette := &palette.Palette{
		Primary:    testColor(100, 150, 200, 255),
		Secondary:  testColor(80, 80, 100, 255),
		Background: testColor(20, 20, 30, 255),
		Shadow1:    testColor(0, 0, 0, 180),
	}

	// Create UI struct with test palette (no Ebiten initialization)
	ui := &EbitenCharacterUI{
		uiPalette: testPalette,
	}

	// Test all color helpers
	tests := []struct {
		name      string
		colorFunc func() color.Color
	}{
		{"overlay color", ui.getOverlayColor},
		{"background color", ui.getBackgroundColor},
		{"primary color", ui.getPrimaryColor},
		{"secondary color", ui.getSecondaryColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.colorFunc()
			if c == nil {
				t.Fatalf("Expected %s, got nil", tt.name)
			}

			// Verify it's a valid color
			r, g, b, a := c.RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0 {
				t.Errorf("Expected non-zero %s", tt.name)
			}
		})
	}
}

// TestUI_GenrePaletteDeterminism verifies same seed produces same palette.
func TestUI_GenrePaletteDeterminism(t *testing.T) {
	seed := int64(12345)
	genreID := "fantasy"

	paletteGen := palette.NewGenerator()
	p1, err1 := paletteGen.Generate(genreID, seed)
	p2, err2 := paletteGen.Generate(genreID, seed)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to generate palettes: %v, %v", err1, err2)
	}

	// Compare primary colors (should be identical)
	r1, g1, b1, a1 := p1.Primary.RGBA()
	r2, g2, b2, a2 := p2.Primary.RGBA()

	if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
		t.Errorf("Expected identical palette colors with same seed, got different colors")
	}
}

// TestUI_GenrePaletteVariation verifies different genres produce different palettes.
func TestUI_GenrePaletteVariation(t *testing.T) {
	seed := int64(12345)

	paletteGen := palette.NewGenerator()
	pFantasy, err1 := paletteGen.Generate("fantasy", seed)
	pScifi, err2 := paletteGen.Generate("scifi", seed)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to generate palettes: %v, %v", err1, err2)
	}

	// Compare primary colors (should be different)
	r1, g1, b1, _ := pFantasy.Primary.RGBA()
	r2, g2, b2, _ := pScifi.Primary.RGBA()

	// At least one component should be different
	if r1 == r2 && g1 == g2 && b1 == b2 {
		t.Errorf("Expected different palette colors for different genres, got identical colors")
	}
}

// TestUI_FallbackPaletteOnNil verifies fallback when palette is nil.
func TestUI_FallbackPaletteOnNil(t *testing.T) {
	// Create UI with nil palette
	ui := &EbitenInventoryUI{
		uiPalette: nil,
	}

	// Should return fallback colors
	overlayColor := ui.getOverlayColor()
	if overlayColor == nil {
		t.Fatal("Expected fallback overlay color, got nil")
	}

	bgColor := ui.getBackgroundColor()
	if bgColor == nil {
		t.Fatal("Expected fallback background color, got nil")
	}
}

// TestUI_PaletteGeneration verifies palette generation for all supported genres.
func TestUI_PaletteGeneration(t *testing.T) {
	genres := []string{"fantasy", "scifi", "cyberpunk", "horror", "postapoc"}
	seed := int64(12345)

	paletteGen := palette.NewGenerator()

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			p, err := paletteGen.Generate(genreID, seed)
			if err != nil {
				t.Fatalf("Failed to generate palette for %s: %v", genreID, err)
			}

			// Verify all required colors are present
			if p.Primary == nil {
				t.Error("Primary color is nil")
			}
			if p.Secondary == nil {
				t.Error("Secondary color is nil")
			}
			if p.Background == nil {
				t.Error("Background color is nil")
			}
			if p.Shadow1 == nil {
				t.Error("Shadow1 color is nil")
			}
		})
	}
}

// Helper function to create test colors
func testColor(r, g, b, a uint8) color.Color {
	return color.RGBA{R: r, G: g, B: b, A: a}
}
