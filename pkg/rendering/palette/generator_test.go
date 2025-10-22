package palette

import (
	"image/color"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.registry == nil {
		t.Error("Generator registry is nil")
	}
	if gen.seedGen == nil {
		t.Error("Generator seedGen is nil")
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    int64
		wantErr bool
	}{
		{
			name:    "fantasy genre",
			genreID: "fantasy",
			seed:    12345,
			wantErr: false,
		},
		{
			name:    "scifi genre",
			genreID: "scifi",
			seed:    54321,
			wantErr: false,
		},
		{
			name:    "horror genre",
			genreID: "horror",
			seed:    11111,
			wantErr: false,
		},
		{
			name:    "cyberpunk genre",
			genreID: "cyberpunk",
			seed:    22222,
			wantErr: false,
		},
		{
			name:    "postapoc genre",
			genreID: "postapoc",
			seed:    33333,
			wantErr: false,
		},
		{
			name:    "invalid genre",
			genreID: "invalid",
			seed:    12345,
			wantErr: true,
		},
	}

	gen := NewGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette, err := gen.Generate(tt.genreID, tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && palette == nil {
				t.Error("Generate() returned nil palette without error")
				return
			}
			if tt.wantErr {
				return // Expected error, test passed
			}

			// Validate palette structure
			if palette.Primary == nil {
				t.Error("Palette Primary color is nil")
			}
			if palette.Secondary == nil {
				t.Error("Palette Secondary color is nil")
			}
			if palette.Background == nil {
				t.Error("Palette Background color is nil")
			}
			if palette.Text == nil {
				t.Error("Palette Text color is nil")
			}
			if palette.Accent1 == nil {
				t.Error("Palette Accent1 color is nil")
			}
			if palette.Accent2 == nil {
				t.Error("Palette Accent2 color is nil")
			}
			if palette.Danger == nil {
				t.Error("Palette Danger color is nil")
			}
			if palette.Success == nil {
				t.Error("Palette Success color is nil")
			}
			if len(palette.Colors) != 8 {
				t.Errorf("Palette Colors length = %d, want 8", len(palette.Colors))
			}
		})
	}
}

func TestGenerateDeterminism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	// Generate palette twice with same seed
	palette1, err := gen.Generate("fantasy", seed)
	if err != nil {
		t.Fatalf("First Generate() error = %v", err)
	}

	palette2, err := gen.Generate("fantasy", seed)
	if err != nil {
		t.Fatalf("Second Generate() error = %v", err)
	}

	// Compare colors
	if !colorEqual(palette1.Primary, palette2.Primary) {
		t.Error("Primary colors differ for same seed")
	}
	if !colorEqual(palette1.Secondary, palette2.Secondary) {
		t.Error("Secondary colors differ for same seed")
	}
	if !colorEqual(palette1.Background, palette2.Background) {
		t.Error("Background colors differ for same seed")
	}
	if len(palette1.Colors) != len(palette2.Colors) {
		t.Errorf("Colors length differs: %d vs %d", len(palette1.Colors), len(palette2.Colors))
	}
	for i := range palette1.Colors {
		if !colorEqual(palette1.Colors[i], palette2.Colors[i]) {
			t.Errorf("Color[%d] differs for same seed", i)
		}
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	gen := NewGenerator()

	palette1, err := gen.Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("First Generate() error = %v", err)
	}

	palette2, err := gen.Generate("fantasy", 54321)
	if err != nil {
		t.Fatalf("Second Generate() error = %v", err)
	}

	// Colors should be different with different seeds
	// (allowing small chance they could be the same by coincidence)
	differentCount := 0
	if !colorEqual(palette1.Primary, palette2.Primary) {
		differentCount++
	}
	if !colorEqual(palette1.Secondary, palette2.Secondary) {
		differentCount++
	}
	if !colorEqual(palette1.Background, palette2.Background) {
		differentCount++
	}

	if differentCount == 0 {
		t.Error("No color differences found between different seeds (highly unlikely)")
	}
}

func TestHSLToColor(t *testing.T) {
	tests := []struct {
		name    string
		h, s, l float64
		want    color.RGBA
	}{
		{
			name: "red",
			h:    0, s: 1.0, l: 0.5,
			want: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name: "green",
			h:    120, s: 1.0, l: 0.5,
			want: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name: "blue",
			h:    240, s: 1.0, l: 0.5,
			want: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
		{
			name: "white",
			h:    0, s: 0, l: 1.0,
			want: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "black",
			h:    0, s: 0, l: 0,
			want: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name: "gray",
			h:    0, s: 0, l: 0.5,
			want: color.RGBA{R: 127, G: 127, B: 127, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hslToColor(tt.h, tt.s, tt.l)
			rgba := colorToRGBA(got)

			// Allow small rounding errors (±2)
			tolerance := uint8(2)
			if !withinTolerance(rgba.R, tt.want.R, tolerance) ||
				!withinTolerance(rgba.G, tt.want.G, tolerance) ||
				!withinTolerance(rgba.B, tt.want.B, tolerance) {
				t.Errorf("hslToColor(%v, %v, %v) = %v, want %v",
					tt.h, tt.s, tt.l, rgba, tt.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name                  string
		value, min, max, want float64
	}{
		{"below minimum", -1.0, 0.0, 1.0, 0.0},
		{"above maximum", 2.0, 0.0, 1.0, 1.0},
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"at minimum", 0.0, 0.0, 1.0, 0.0},
		{"at maximum", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%v, %v, %v) = %v, want %v",
					tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

// Helper functions

func colorEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func colorToRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func withinTolerance(a, b, tolerance uint8) bool {
	diff := int(a) - int(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= int(tolerance)
}
