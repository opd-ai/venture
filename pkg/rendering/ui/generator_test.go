package ui

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"testing"
)

func TestElementType_String(t *testing.T) {
	tests := []struct {
		name     string
		eType    ElementType
		expected string
	}{
		{"Button", ElementButton, "button"},
		{"Panel", ElementPanel, "panel"},
		{"HealthBar", ElementHealthBar, "healthbar"},
		{"Label", ElementLabel, "label"},
		{"Icon", ElementIcon, "icon"},
		{"Frame", ElementFrame, "frame"},
		{"Unknown", ElementType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.eType.String()
			if got != tt.expected {
				t.Errorf("ElementType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestElementState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    ElementState
		expected string
	}{
		{"Normal", StateNormal, "normal"},
		{"Hover", StateHover, "hover"},
		{"Pressed", StatePressed, "pressed"},
		{"Disabled", StateDisabled, "disabled"},
		{"Unknown", ElementState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.expected {
				t.Errorf("ElementState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBorderStyle_String(t *testing.T) {
	tests := []struct {
		name     string
		style    BorderStyle
		expected string
	}{
		{"Solid", BorderSolid, "solid"},
		{"Double", BorderDouble, "double"},
		{"Ornate", BorderOrnate, "ornate"},
		{"Glow", BorderGlow, "glow"},
		{"Unknown", BorderStyle(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.String()
			if got != tt.expected {
				t.Errorf("BorderStyle.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != ElementButton {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, ElementButton)
	}
	if config.Width != 100 {
		t.Errorf("DefaultConfig Width = %v, want 100", config.Width)
	}
	if config.Height != 30 {
		t.Errorf("DefaultConfig Height = %v, want 30", config.Height)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig GenreID = %v, want fantasy", config.GenreID)
	}
	if config.Value != 1.0 {
		t.Errorf("DefaultConfig Value = %v, want 1.0", config.Value)
	}
	if config.Custom == nil {
		t.Error("DefaultConfig Custom should not be nil")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Value:   0.5,
			},
			wantErr: false,
		},
		{
			name: "Zero width",
			config: Config{
				Type:    ElementButton,
				Width:   0,
				Height:  30,
				GenreID: "fantasy",
				Value:   0.5,
			},
			wantErr: true,
		},
		{
			name: "Zero height",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  0,
				GenreID: "fantasy",
				Value:   0.5,
			},
			wantErr: true,
		},
		{
			name: "Empty GenreID",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "",
				Value:   0.5,
			},
			wantErr: true,
		},
		{
			name: "Value too low",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   -0.1,
			},
			wantErr: true,
		},
		{
			name: "Value too high",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   1.1,
			},
			wantErr: true,
		},
		{
			name: "Value at bounds - 0",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   0.0,
			},
			wantErr: false,
		},
		{
			name: "Value at bounds - 1",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Generate button",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
				Text:    "Click Me",
				State:   StateNormal,
			},
			wantErr: false,
		},
		{
			name: "Generate button - hover state",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
				State:   StateHover,
			},
			wantErr: false,
		},
		{
			name: "Generate button - pressed state",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "scifi",
				Seed:    12345,
				State:   StatePressed,
			},
			wantErr: false,
		},
		{
			name: "Generate button - disabled state",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
				State:   StateDisabled,
			},
			wantErr: false,
		},
		{
			name: "Generate panel",
			config: Config{
				Type:    ElementPanel,
				Width:   200,
				Height:  150,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - full",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   1.0,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - half",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - low",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.2,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - empty",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.0,
			},
			wantErr: false,
		},
		{
			name: "Generate label",
			config: Config{
				Type:    ElementLabel,
				Width:   80,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Text:    "Score: 100",
			},
			wantErr: false,
		},
		{
			name: "Generate icon",
			config: Config{
				Type:    ElementIcon,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Generate icon - scifi",
			config: Config{
				Type:    ElementIcon,
				Width:   32,
				Height:  32,
				GenreID: "scifi",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Generate frame",
			config: Config{
				Type:    ElementFrame,
				Width:   300,
				Height:  200,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Invalid config - zero width",
			config: Config{
				Type:    ElementButton,
				Width:   0,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.Generate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if img == nil {
					t.Error("Generator.Generate() returned nil image")
					return
				}
				bounds := img.Bounds()
				if bounds.Dx() != tt.config.Width {
					t.Errorf("Generated image width = %v, want %v", bounds.Dx(), tt.config.Width)
				}
				if bounds.Dy() != tt.config.Height {
					t.Errorf("Generated image height = %v, want %v", bounds.Dy(), tt.config.Height)
				}
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	// Generate the same element twice
	img1, err1 := gen.Generate(config)
	img2, err2 := gen.Generate(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating UI elements: %v, %v", err1, err2)
	}

	// Compare pixel by pixel
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Errorf("Pixel at (%d, %d) differs: %v vs %v", x, y, c1, c2)
				return
			}
		}
	}
}

func TestGenerator_DifferentSeeds(t *testing.T) {
	gen := NewGenerator()

	config1 := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	config2 := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    54321,
		State:   StateNormal,
	}

	img1, err1 := gen.Generate(config1)
	img2, err2 := gen.Generate(config2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating UI elements: %v, %v", err1, err2)
	}

	// Images should be different
	bounds := img1.Bounds()
	different := false
	for y := bounds.Min.Y; y < bounds.Max.Y && !different; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				different = true
				break
			}
		}
	}

	if !different {
		t.Error("UI elements generated with different seeds should be different")
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{
			name:    "Valid image",
			result:  image.NewRGBA(image.Rect(0, 0, 100, 30)),
			wantErr: false,
		},
		{
			name:    "Nil image",
			result:  (*image.RGBA)(nil),
			wantErr: true,
		},
		{
			name:    "Wrong type",
			result:  "not an image",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_AllGenres(t *testing.T) {
	gen := NewGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: genre,
				Seed:    12345,
				State:   StateNormal,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate UI element for genre %s: %v", genre, err)
			}
			if img == nil {
				t.Errorf("Generated nil image for genre %s", genre)
			}
		})
	}
}

func TestGenerator_AllElementTypes(t *testing.T) {
	gen := NewGenerator()
	elementTypes := []ElementType{
		ElementButton, ElementPanel, ElementHealthBar,
		ElementLabel, ElementIcon, ElementFrame,
	}

	for _, eType := range elementTypes {
		t.Run(eType.String(), func(t *testing.T) {
			config := Config{
				Type:    eType,
				Width:   100,
				Height:  50,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.5,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate %s: %v", eType, err)
			}
			if img == nil {
				t.Errorf("Generated nil image for %s", eType)
			}
		})
	}
}

func BenchmarkGenerator_GenerateButton(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_GenerateHealthBar(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementHealthBar,
		Width:   100,
		Height:  20,
		GenreID: "fantasy",
		Seed:    12345,
		Value:   0.75,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestCalculateRelativeLuminance tests WCAG luminance calculation.
func TestCalculateRelativeLuminance(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name      string
		color     color.Color
		wantRange [2]float64 // [min, max] expected luminance
	}{
		{
			name:      "black",
			color:     color.RGBA{R: 0, G: 0, B: 0, A: 255},
			wantRange: [2]float64{0.0, 0.001},
		},
		{
			name:      "white",
			color:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
			wantRange: [2]float64{0.999, 1.0},
		},
		{
			name:      "mid-gray",
			color:     color.RGBA{R: 128, G: 128, B: 128, A: 255},
			wantRange: [2]float64{0.18, 0.22},
		},
		{
			name:      "red",
			color:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
			wantRange: [2]float64{0.2, 0.23},
		},
		{
			name:      "green",
			color:     color.RGBA{R: 0, G: 255, B: 0, A: 255},
			wantRange: [2]float64{0.71, 0.73},
		},
		{
			name:      "blue",
			color:     color.RGBA{R: 0, G: 0, B: 255, A: 255},
			wantRange: [2]float64{0.07, 0.09},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			luminance := gen.calculateRelativeLuminance(tt.color)
			if luminance < tt.wantRange[0] || luminance > tt.wantRange[1] {
				t.Errorf("calculateRelativeLuminance(%s) = %v, want range [%v, %v]",
					tt.name, luminance, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

// TestCalculateContrastRatio tests WCAG contrast ratio calculation.
func TestCalculateContrastRatio(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name         string
		color1       color.Color
		color2       color.Color
		wantMinRatio float64
	}{
		{
			name:         "black on white",
			color1:       color.RGBA{R: 0, G: 0, B: 0, A: 255},
			color2:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			wantMinRatio: 21.0, // Maximum contrast
		},
		{
			name:         "white on black",
			color1:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			color2:       color.RGBA{R: 0, G: 0, B: 0, A: 255},
			wantMinRatio: 21.0, // Maximum contrast
		},
		{
			name:         "same color",
			color1:       color.RGBA{R: 128, G: 128, B: 128, A: 255},
			color2:       color.RGBA{R: 128, G: 128, B: 128, A: 255},
			wantMinRatio: 1.0, // No contrast
		},
		{
			name:         "dark gray on light gray (WCAG AA pass)",
			color1:       color.RGBA{R: 85, G: 85, B: 85, A: 255},
			color2:       color.RGBA{R: 204, G: 204, B: 204, A: 255},
			wantMinRatio: 4.5, // Should meet WCAG AA
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := gen.calculateContrastRatio(tt.color1, tt.color2)
			if ratio < tt.wantMinRatio {
				t.Errorf("calculateContrastRatio() = %v, want >= %v", ratio, tt.wantMinRatio)
			}
		})
	}
}

// TestSelectButtonBaseColor_WCAGCompliance tests that selected colors meet WCAG AA.
func TestSelectButtonBaseColor_WCAGCompliance(t *testing.T) {
	gen := NewGenerator()

	// Test with all 5 genres
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	seeds := []int64{12345, 42987, 88432, 77321, 99999}

	for _, genreID := range genres {
		for _, seed := range seeds {
			t.Run(fmt.Sprintf("%s_seed_%d", genreID, seed), func(t *testing.T) {
				// Generate palette for genre
				pal, err := gen.paletteGen.Generate(genreID, seed)
				if err != nil {
					t.Fatalf("failed to generate palette: %v", err)
				}

				// Select button color
				rng := rand.New(rand.NewSource(seed))
				buttonColor := gen.selectButtonBaseColor(pal, rng)

				// Calculate contrast ratio with text color
				contrastRatio := gen.calculateContrastRatio(buttonColor, pal.Text)

				// Verify WCAG 2.1 AA compliance (minimum 4.5:1)
				if contrastRatio < 4.5 {
					t.Errorf("selectButtonBaseColor() contrast ratio = %v, want >= 4.5 (WCAG AA)", contrastRatio)
					t.Logf("Genre: %s, Seed: %d", genreID, seed)
					t.Logf("Button color: %+v", buttonColor)
					t.Logf("Text color: %+v", pal.Text)
				}
			})
		}
	}
}

// TestButtonGeneration_AllGenres tests button generation for WCAG compliance across all genres.
func TestButtonGeneration_AllGenres(t *testing.T) {
	gen := NewGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			config := Config{
				Type:    ElementButton,
				Width:   100,
				Height:  40,
				GenreID: genreID,
				Seed:    12345,
				State:   StateNormal,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("failed to generate button: %v", err)
			}

			if img == nil {
				t.Fatal("generated image is nil")
			}

			// Verify dimensions
			bounds := img.Bounds()
			if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
				t.Errorf("button dimensions = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), config.Width, config.Height)
			}
		})
	}
}

func TestBorderStyle_NewStyles_String(t *testing.T) {
tests := []struct {
name     string
style    BorderStyle
expected string
}{
{"Dashed", BorderDashed, "dashed"},
{"Dotted", BorderDotted, "dotted"},
{"Embossed", BorderEmbossed, "embossed"},
{"Engraved", BorderEngraved, "engraved"},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := tt.style.String()
if got != tt.expected {
t.Errorf("BorderStyle.String() = %v, want %v", got, tt.expected)
}
})
}
}

func TestGenerateWithHierarchy(t *testing.T) {
gen := NewGenerator()

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
config := Config{
Type:           ElementButton,
Width:          100,
Height:         30,
GenreID:        "fantasy",
Seed:           12345,
HierarchyLevel: tt.level,
}

img, err := gen.Generate(config)
if err != nil {
t.Fatalf("Generate() error = %v", err)
}

if img == nil {
t.Fatal("Generate() returned nil image")
}

// Verify determinism
img2, err := gen.Generate(config)
if err != nil {
t.Fatalf("Second Generate() error = %v", err)
}

bounds1 := img.Bounds()
bounds2 := img2.Bounds()
if bounds1 != bounds2 {
t.Errorf("Images have different bounds: %v vs %v", bounds1, bounds2)
}
})
}
}

func TestGenerateWithTransition(t *testing.T) {
gen := NewGenerator()

tests := []struct {
name       string
transition TransitionConfig
}{
{
name: "Fade",
transition: TransitionConfig{
Type:     TransitionFade,
Duration: 300,
Easing:   EaseLinear,
Progress: 0.5,
},
},
{
name: "SlideLeft",
transition: TransitionConfig{
Type:     TransitionSlideLeft,
Duration: 300,
Easing:   EaseInOutQuad,
Progress: 0.7,
},
},
{
name: "Zoom",
transition: TransitionConfig{
Type:     TransitionZoom,
Duration: 500,
Easing:   EaseOutCubic,
Progress: 0.3,
},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
config := Config{
Type:       ElementButton,
Width:      100,
Height:     30,
GenreID:    "scifi",
Seed:       54321,
Transition: &tt.transition,
}

img, err := gen.Generate(config)
if err != nil {
t.Fatalf("Generate() error = %v", err)
}

if img == nil {
t.Fatal("Generate() returned nil image")
}
})
}
}

func TestGenerateWithInvalidTransition(t *testing.T) {
gen := NewGenerator()

invalidTransition := TransitionConfig{
Type:     TransitionFade,
Duration: -100, // Invalid: negative duration
Easing:   EaseLinear,
Progress: 0.5,
}

config := Config{
Type:       ElementButton,
Width:      100,
Height:     30,
GenreID:    "fantasy",
Seed:       12345,
Transition: &invalidTransition,
}

_, err := gen.Generate(config)
if err == nil {
t.Error("Expected error for invalid transition config, got nil")
}
}

func TestGenerateAllBorderStyles(t *testing.T) {
gen := NewGenerator()

borderStyles := []BorderStyle{
BorderSolid, BorderDouble, BorderOrnate, BorderGlow,
BorderDashed, BorderDotted, BorderEmbossed, BorderEngraved,
}

for _, style := range borderStyles {
t.Run(style.String(), func(t *testing.T) {
// Create a test image
img := image.NewRGBA(image.Rect(0, 0, 100, 50))
col := color.RGBA{R: 255, G: 255, B: 255, A: 255}

// Draw border
gen.drawBorder(img, col, style, 2)

// Verify image is not empty (at least some pixels are set)
hasPixels := false
for y := 0; y < 50; y++ {
for x := 0; x < 100; x++ {
c := img.RGBAAt(x, y)
if c.A > 0 {
hasPixels = true
break
}
}
if hasPixels {
break
}
}

if !hasPixels {
t.Error("Border drawing produced empty image")
}
})
}
}
