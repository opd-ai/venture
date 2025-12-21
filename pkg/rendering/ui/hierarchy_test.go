package ui

import (
	"image"
	"image/color"
	"testing"
)

func TestHierarchyLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    HierarchyLevel
		expected string
	}{
		{"Primary", HierarchyPrimary, "primary"},
		{"Secondary", HierarchySecondary, "secondary"},
		{"Tertiary", HierarchyTertiary, "tertiary"},
		{"Quaternary", HierarchyQuaternary, "quaternary"},
		{"Unknown", HierarchyLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("HierarchyLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetHierarchyStyle(t *testing.T) {
	tests := []struct {
		name  string
		level HierarchyLevel
	}{
		{"Primary", HierarchyPrimary},
		{"Secondary", HierarchySecondary},
		{"Tertiary", HierarchyTertiary},
		{"Quaternary", HierarchyQuaternary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetHierarchyStyle(tt.level)

			// Verify all fields are set
			if style.Scale <= 0 {
				t.Errorf("Scale should be positive, got %f", style.Scale)
			}
			if style.Opacity < 0 || style.Opacity > 1.0 {
				t.Errorf("Opacity should be in [0, 1], got %f", style.Opacity)
			}
			if style.BorderThickness < 0 {
				t.Errorf("BorderThickness should be non-negative, got %d", style.BorderThickness)
			}
			if style.BackgroundOpacity < 0 || style.BackgroundOpacity > 1.0 {
				t.Errorf("BackgroundOpacity should be in [0, 1], got %f", style.BackgroundOpacity)
			}
			if style.FontScale <= 0 {
				t.Errorf("FontScale should be positive, got %f", style.FontScale)
			}
		})
	}
}

// TestGetHierarchyStyle_Phase45BorderThickness verifies Phase 45 scaled border thicknesses.
// Border thicknesses are scaled 2× for 64×64 UI elements.
func TestGetHierarchyStyle_Phase45BorderThickness(t *testing.T) {
	tests := []struct {
		name              string
		level             HierarchyLevel
		expectedThickness int
	}{
		{"Primary", HierarchyPrimary, 6},       // Phase 45: 3 → 6
		{"Secondary", HierarchySecondary, 4},   // Phase 45: 2 → 4
		{"Tertiary", HierarchyTertiary, 2},     // Phase 45: 1 → 2
		{"Quaternary", HierarchyQuaternary, 2}, // Phase 45: 1 → 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetHierarchyStyle(tt.level)
			if style.BorderThickness != tt.expectedThickness {
				t.Errorf("BorderThickness = %d, want %d (Phase 45 2× scaling)",
					style.BorderThickness, tt.expectedThickness)
			}
		})
	}
}

func TestGetHierarchyStyle_Ordering(t *testing.T) {
	// Verify that primary has larger scale than secondary, etc.
	primary := GetHierarchyStyle(HierarchyPrimary)
	secondary := GetHierarchyStyle(HierarchySecondary)
	tertiary := GetHierarchyStyle(HierarchyTertiary)
	quaternary := GetHierarchyStyle(HierarchyQuaternary)

	if primary.Scale <= secondary.Scale {
		t.Errorf("Primary scale (%f) should be > secondary scale (%f)", primary.Scale, secondary.Scale)
	}
	if secondary.Scale <= tertiary.Scale {
		t.Errorf("Secondary scale (%f) should be > tertiary scale (%f)", secondary.Scale, tertiary.Scale)
	}
	if tertiary.Scale <= quaternary.Scale {
		t.Errorf("Tertiary scale (%f) should be > quaternary scale (%f)", tertiary.Scale, quaternary.Scale)
	}
}

func TestApplyHierarchyOpacity(t *testing.T) {
	baseColor := color.RGBA{R: 255, G: 128, B: 64, A: 255}

	tests := []struct {
		name  string
		level HierarchyLevel
	}{
		{"Primary", HierarchyPrimary},
		{"Secondary", HierarchySecondary},
		{"Tertiary", HierarchyTertiary},
		{"Quaternary", HierarchyQuaternary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyHierarchyOpacity(baseColor, tt.level)
			r, g, b, a := result.RGBA()

			// RGB values should be preserved
			if uint8(r>>8) != baseColor.R {
				t.Errorf("R channel changed: got %d, want %d", uint8(r>>8), baseColor.R)
			}
			if uint8(g>>8) != baseColor.G {
				t.Errorf("G channel changed: got %d, want %d", uint8(g>>8), baseColor.G)
			}
			if uint8(b>>8) != baseColor.B {
				t.Errorf("B channel changed: got %d, want %d", uint8(b>>8), baseColor.B)
			}

			// Alpha should be modulated by hierarchy level
			style := GetHierarchyStyle(tt.level)
			expectedAlpha := uint8(float64(baseColor.A) * style.Opacity)
			if uint8(a>>8) != expectedAlpha {
				t.Errorf("Alpha modulation incorrect: got %d, want %d", uint8(a>>8), expectedAlpha)
			}
		})
	}
}

func TestSeparatorStyle_String(t *testing.T) {
	tests := []struct {
		name     string
		style    SeparatorStyle
		expected string
	}{
		{"Line", SeparatorLine, "line"},
		{"Dashed", SeparatorDashed, "dashed"},
		{"Dotted", SeparatorDotted, "dotted"},
		{"Gradient", SeparatorGradient, "gradient"},
		{"Ornamental", SeparatorOrnamental, "ornamental"},
		{"Unknown", SeparatorStyle(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.String()
			if got != tt.expected {
				t.Errorf("SeparatorStyle.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateSeparator(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		width   int
		height  int
		style   SeparatorStyle
		genreID string
		seed    int64
	}{
		{"Line", 100, 10, SeparatorLine, "fantasy", 12345},
		{"Dashed", 100, 10, SeparatorDashed, "scifi", 12346},
		{"Dotted", 100, 10, SeparatorDotted, "horror", 12347},
		{"Gradient", 100, 10, SeparatorGradient, "cyberpunk", 12348},
		{"Ornamental", 100, 10, SeparatorOrnamental, "postapoc", 12349},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := gen.GenerateSeparator(tt.width, tt.height, tt.style, tt.genreID, tt.seed)

			if img == nil {
				t.Fatal("GenerateSeparator returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.width {
				t.Errorf("Width = %d, want %d", bounds.Dx(), tt.width)
			}
			if bounds.Dy() != tt.height {
				t.Errorf("Height = %d, want %d", bounds.Dy(), tt.height)
			}

			// Verify determinism: same seed should produce same result
			img2 := gen.GenerateSeparator(tt.width, tt.height, tt.style, tt.genreID, tt.seed)
			if !imagesEqual(img, img2) {
				t.Error("GenerateSeparator is not deterministic")
			}
		})
	}
}

func TestGenerateSeparator_Determinism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(99999)

	img1 := gen.GenerateSeparator(100, 10, SeparatorOrnamental, "fantasy", seed)
	img2 := gen.GenerateSeparator(100, 10, SeparatorOrnamental, "fantasy", seed)

	if !imagesEqual(img1, img2) {
		t.Error("GenerateSeparator with same seed produced different results")
	}
}

func TestGenerateGroupContainer(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name             string
		config           GroupConfig
		expectBorder     bool
		expectBackground bool
	}{
		{
			name: "WithBorderAndBackground",
			config: GroupConfig{
				Width:          200,
				Height:         100,
				Padding:        10,
				Level:          HierarchyPrimary,
				GenreID:        "fantasy",
				Seed:           12345,
				ShowBorder:     true,
				ShowBackground: true,
			},
			expectBorder:     true,
			expectBackground: true,
		},
		{
			name: "BorderOnly",
			config: GroupConfig{
				Width:          150,
				Height:         80,
				Padding:        5,
				Level:          HierarchySecondary,
				GenreID:        "scifi",
				Seed:           54321,
				ShowBorder:     true,
				ShowBackground: false,
			},
			expectBorder:     true,
			expectBackground: false,
		},
		{
			name: "BackgroundOnly",
			config: GroupConfig{
				Width:          120,
				Height:         60,
				Padding:        8,
				Level:          HierarchyTertiary,
				GenreID:        "horror",
				Seed:           11111,
				ShowBorder:     false,
				ShowBackground: true,
			},
			expectBorder:     false,
			expectBackground: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := gen.GenerateGroupContainer(tt.config)

			if img == nil {
				t.Fatal("GenerateGroupContainer returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.config.Width {
				t.Errorf("Width = %d, want %d", bounds.Dx(), tt.config.Width)
			}
			if bounds.Dy() != tt.config.Height {
				t.Errorf("Height = %d, want %d", bounds.Dy(), tt.config.Height)
			}

			// Verify determinism
			img2 := gen.GenerateGroupContainer(tt.config)
			if !imagesEqual(img, img2) {
				t.Error("GenerateGroupContainer is not deterministic")
			}
		})
	}
}

func TestGenerateGroupContainer_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := GroupConfig{
		Width:          100,
		Height:         50,
		Padding:        5,
		Level:          HierarchySecondary,
		GenreID:        "fantasy",
		Seed:           77777,
		ShowBorder:     true,
		ShowBackground: true,
	}

	img1 := gen.GenerateGroupContainer(config)
	img2 := gen.GenerateGroupContainer(config)

	if !imagesEqual(img1, img2) {
		t.Error("GenerateGroupContainer with same config produced different results")
	}
}

// Helper function to compare two images for equality
func imagesEqual(img1, img2 *image.RGBA) bool {
	if img1 == nil || img2 == nil {
		return img1 == img2
	}

	b1 := img1.Bounds()
	b2 := img2.Bounds()

	if b1 != b2 {
		return false
	}

	for y := b1.Min.Y; y < b1.Max.Y; y++ {
		for x := b1.Min.X; x < b1.Max.X; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)
			if c1 != c2 {
				return false
			}
		}
	}

	return true
}
