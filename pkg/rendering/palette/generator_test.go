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
			if palette.Accent3 == nil {
				t.Error("Palette Accent3 color is nil")
			}
			if palette.Highlight1 == nil {
				t.Error("Palette Highlight1 color is nil")
			}
			if palette.Highlight2 == nil {
				t.Error("Palette Highlight2 color is nil")
			}
			if palette.Shadow1 == nil {
				t.Error("Palette Shadow1 color is nil")
			}
			if palette.Shadow2 == nil {
				t.Error("Palette Shadow2 color is nil")
			}
			if palette.Neutral == nil {
				t.Error("Palette Neutral color is nil")
			}
			if palette.Danger == nil {
				t.Error("Palette Danger color is nil")
			}
			if palette.Success == nil {
				t.Error("Palette Success color is nil")
			}
			if palette.Warning == nil {
				t.Error("Palette Warning color is nil")
			}
			if palette.Info == nil {
				t.Error("Palette Info color is nil")
			}
			if len(palette.Colors) < 12 {
				t.Errorf("Palette Colors length = %d, want >= 12", len(palette.Colors))
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

// Phase 4 Tests: Harmony, Mood, and Rarity

func TestHarmonyType_String(t *testing.T) {
	tests := []struct {
		harmony HarmonyType
		want    string
	}{
		{HarmonyComplementary, "Complementary"},
		{HarmonyAnalogous, "Analogous"},
		{HarmonyTriadic, "Triadic"},
		{HarmonyTetradic, "Tetradic"},
		{HarmonySplitComplementary, "SplitComplementary"},
		{HarmonyMonochromatic, "Monochromatic"},
		{HarmonyType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.harmony.String()
			if got != tt.want {
				t.Errorf("HarmonyType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMoodType_String(t *testing.T) {
	tests := []struct {
		mood MoodType
		want string
	}{
		{MoodNormal, "Normal"},
		{MoodBright, "Bright"},
		{MoodDark, "Dark"},
		{MoodSaturated, "Saturated"},
		{MoodMuted, "Muted"},
		{MoodVibrant, "Vibrant"},
		{MoodPastel, "Pastel"},
		{MoodType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mood.String()
			if got != tt.want {
				t.Errorf("MoodType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRarity_String(t *testing.T) {
	tests := []struct {
		rarity Rarity
		want   string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityEpic, "Epic"},
		{RarityLegendary, "Legendary"},
		{Rarity(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.rarity.String()
			if got != tt.want {
				t.Errorf("Rarity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateWithOptions_Harmony(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)
	
	tests := []struct {
		name    string
		harmony HarmonyType
	}{
		{"complementary", HarmonyComplementary},
		{"analogous", HarmonyAnalogous},
		{"triadic", HarmonyTriadic},
		{"tetradic", HarmonyTetradic},
		{"split complementary", HarmonySplitComplementary},
		{"monochromatic", HarmonyMonochromatic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.Harmony = tt.harmony

			palette, err := gen.GenerateWithOptions("fantasy", seed, opts)
			if err != nil {
				t.Fatalf("GenerateWithOptions() error = %v", err)
			}
			if palette == nil {
				t.Fatal("GenerateWithOptions() returned nil palette")
			}

			// Validate palette has all required colors
			if palette.Primary == nil {
				t.Error("Primary color is nil")
			}
			if palette.Secondary == nil {
				t.Error("Secondary color is nil")
			}
			if len(palette.Colors) < 12 {
				t.Errorf("Colors length = %d, want >= 12", len(palette.Colors))
			}
		})
	}
}

func TestGenerateWithOptions_Mood(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	tests := []struct {
		name string
		mood MoodType
	}{
		{"normal", MoodNormal},
		{"bright", MoodBright},
		{"dark", MoodDark},
		{"saturated", MoodSaturated},
		{"muted", MoodMuted},
		{"vibrant", MoodVibrant},
		{"pastel", MoodPastel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.Mood = tt.mood

			palette, err := gen.GenerateWithOptions("fantasy", seed, opts)
			if err != nil {
				t.Fatalf("GenerateWithOptions() error = %v", err)
			}
			if palette == nil {
				t.Fatal("GenerateWithOptions() returned nil palette")
			}

			// All moods should produce valid palettes
			if len(palette.Colors) < 12 {
				t.Errorf("Colors length = %d, want >= 12", len(palette.Colors))
			}
		})
	}
}

func TestGenerateWithOptions_Rarity(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	tests := []struct {
		name   string
		rarity Rarity
	}{
		{"common", RarityCommon},
		{"uncommon", RarityUncommon},
		{"rare", RarityRare},
		{"epic", RarityEpic},
		{"legendary", RarityLegendary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.Rarity = tt.rarity

			palette, err := gen.GenerateWithOptions("fantasy", seed, opts)
			if err != nil {
				t.Fatalf("GenerateWithOptions() error = %v", err)
			}
			if palette == nil {
				t.Fatal("GenerateWithOptions() returned nil palette")
			}

			// All rarities should produce valid palettes
			if len(palette.Colors) < 12 {
				t.Errorf("Colors length = %d, want >= 12", len(palette.Colors))
			}
		})
	}
}

func TestGenerateWithOptions_MinColors(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	tests := []struct {
		name      string
		minColors int
	}{
		{"12 colors", 12},
		{"16 colors", 16},
		{"20 colors", 20},
		{"24 colors", 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.MinColors = tt.minColors

			palette, err := gen.GenerateWithOptions("fantasy", seed, opts)
			if err != nil {
				t.Fatalf("GenerateWithOptions() error = %v", err)
			}
			if palette == nil {
				t.Fatal("GenerateWithOptions() returned nil palette")
			}

			if len(palette.Colors) < tt.minColors {
				t.Errorf("Colors length = %d, want >= %d", len(palette.Colors), tt.minColors)
			}
		})
	}
}

func TestGenerateWithOptions_Determinism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	opts := GenerationOptions{
		Harmony:   HarmonyTriadic,
		Mood:      MoodVibrant,
		Rarity:    RarityEpic,
		MinColors: 16,
	}

	// Generate twice with same parameters
	palette1, err := gen.GenerateWithOptions("scifi", seed, opts)
	if err != nil {
		t.Fatalf("First GenerateWithOptions() error = %v", err)
	}

	palette2, err := gen.GenerateWithOptions("scifi", seed, opts)
	if err != nil {
		t.Fatalf("Second GenerateWithOptions() error = %v", err)
	}

	// Verify determinism
	if !colorEqual(palette1.Primary, palette2.Primary) {
		t.Error("Primary colors differ for same seed and options")
	}
	if !colorEqual(palette1.Secondary, palette2.Secondary) {
		t.Error("Secondary colors differ for same seed and options")
	}
	if len(palette1.Colors) != len(palette2.Colors) {
		t.Errorf("Colors length differs: %d vs %d", len(palette1.Colors), len(palette2.Colors))
	}
	for i := range palette1.Colors {
		if !colorEqual(palette1.Colors[i], palette2.Colors[i]) {
			t.Errorf("Color[%d] differs for same seed and options", i)
		}
	}
}

func TestGetHarmonyHues(t *testing.T) {
	gen := NewGenerator()
	baseHue := 30.0

	tests := []struct {
		name        string
		harmony     HarmonyType
		wantHueCount int
	}{
		{"complementary", HarmonyComplementary, 2},
		{"analogous", HarmonyAnalogous, 3},
		{"triadic", HarmonyTriadic, 3},
		{"tetradic", HarmonyTetradic, 4},
		{"split complementary", HarmonySplitComplementary, 3},
		{"monochromatic", HarmonyMonochromatic, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hues := gen.getHarmonyHues(baseHue, tt.harmony)
			if len(hues) != tt.wantHueCount {
				t.Errorf("getHarmonyHues() returned %d hues, want %d", len(hues), tt.wantHueCount)
			}

			// Verify base hue is included
			if hues[0] != baseHue {
				t.Errorf("First hue = %v, want %v (base hue)", hues[0], baseHue)
			}

			// Verify all hues are in valid range [0, 360)
			for i, hue := range hues {
				if hue < 0 || hue >= 360 {
					t.Errorf("Hue[%d] = %v, want [0, 360)", i, hue)
				}
			}
		})
	}
}

func TestApplyMood(t *testing.T) {
	gen := NewGenerator()
	baseScheme := ColorScheme{
		BaseHue:             100,
		Saturation:          0.6,
		Lightness:           0.5,
		HueVariation:        30,
		SaturationVariation: 0.2,
		LightnessVariation:  0.2,
	}

	tests := []struct {
		name           string
		mood           MoodType
		checkLightness func(float64) bool
		checkSaturation func(float64) bool
	}{
		{
			name: "bright increases lightness",
			mood: MoodBright,
			checkLightness: func(l float64) bool { return l > baseScheme.Lightness },
			checkSaturation: func(s float64) bool { return true }, // Don't check
		},
		{
			name: "dark decreases lightness",
			mood: MoodDark,
			checkLightness: func(l float64) bool { return l < baseScheme.Lightness },
			checkSaturation: func(s float64) bool { return true },
		},
		{
			name: "saturated increases saturation",
			mood: MoodSaturated,
			checkLightness: func(l float64) bool { return true },
			checkSaturation: func(s float64) bool { return s > baseScheme.Saturation },
		},
		{
			name: "muted decreases saturation",
			mood: MoodMuted,
			checkLightness: func(l float64) bool { return true },
			checkSaturation: func(s float64) bool { return s < baseScheme.Saturation },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := gen.applyMood(baseScheme, tt.mood)

			if !tt.checkLightness(adjusted.Lightness) {
				t.Errorf("Lightness adjustment failed: got %v, base was %v", adjusted.Lightness, baseScheme.Lightness)
			}
			if !tt.checkSaturation(adjusted.Saturation) {
				t.Errorf("Saturation adjustment failed: got %v, base was %v", adjusted.Saturation, baseScheme.Saturation)
			}

			// Verify values stay in valid range
			if adjusted.Saturation < 0 || adjusted.Saturation > 1 {
				t.Errorf("Saturation out of range: %v", adjusted.Saturation)
			}
			if adjusted.Lightness < 0 || adjusted.Lightness > 1 {
				t.Errorf("Lightness out of range: %v", adjusted.Lightness)
			}
		})
	}
}

func TestApplyRarity(t *testing.T) {
	gen := NewGenerator()
	baseScheme := ColorScheme{
		BaseHue:             200,
		Saturation:          0.5,
		Lightness:           0.5,
		HueVariation:        20,
		SaturationVariation: 0.15,
		LightnessVariation:  0.15,
	}

	tests := []struct {
		name   string
		rarity Rarity
		checkIntensity func(ColorScheme) bool
	}{
		{
			name:   "common is muted",
			rarity: RarityCommon,
			checkIntensity: func(s ColorScheme) bool {
				return s.Saturation <= baseScheme.Saturation
			},
		},
		{
			name:   "rare is vibrant",
			rarity: RarityRare,
			checkIntensity: func(s ColorScheme) bool {
				return s.Saturation > baseScheme.Saturation
			},
		},
		{
			name:   "epic is more intense",
			rarity: RarityEpic,
			checkIntensity: func(s ColorScheme) bool {
				return s.Saturation > baseScheme.Saturation*1.2
			},
		},
		{
			name:   "legendary is most intense",
			rarity: RarityLegendary,
			checkIntensity: func(s ColorScheme) bool {
				return s.Saturation > baseScheme.Saturation*1.3 &&
					   s.HueVariation > baseScheme.HueVariation
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := gen.applyRarity(baseScheme, tt.rarity)

			if !tt.checkIntensity(adjusted) {
				t.Errorf("Rarity intensity check failed for %s", tt.name)
			}

			// Verify values stay in valid range
			if adjusted.Saturation < 0 || adjusted.Saturation > 1 {
				t.Errorf("Saturation out of range: %v", adjusted.Saturation)
			}
			if adjusted.Lightness < 0 || adjusted.Lightness > 1 {
				t.Errorf("Lightness out of range: %v", adjusted.Lightness)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Harmony != HarmonyComplementary {
		t.Errorf("Default Harmony = %v, want %v", opts.Harmony, HarmonyComplementary)
	}
	if opts.Mood != MoodNormal {
		t.Errorf("Default Mood = %v, want %v", opts.Mood, MoodNormal)
	}
	if opts.Rarity != RarityCommon {
		t.Errorf("Default Rarity = %v, want %v", opts.Rarity, RarityCommon)
	}
	if opts.MinColors != 12 {
		t.Errorf("Default MinColors = %d, want 12", opts.MinColors)
	}
}

// Benchmarks for Phase 4 features

func BenchmarkGenerateWithHarmony(b *testing.B) {
	gen := NewGenerator()
	opts := GenerationOptions{
		Harmony:   HarmonyTriadic,
		Mood:      MoodNormal,
		Rarity:    RarityCommon,
		MinColors: 12,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateWithOptions("fantasy", 12345, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateWithMood(b *testing.B) {
	gen := NewGenerator()
	opts := GenerationOptions{
		Harmony:   HarmonyComplementary,
		Mood:      MoodVibrant,
		Rarity:    RarityCommon,
		MinColors: 12,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateWithOptions("scifi", 54321, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateWithRarity(b *testing.B) {
	gen := NewGenerator()
	opts := GenerationOptions{
		Harmony:   HarmonyComplementary,
		Mood:      MoodNormal,
		Rarity:    RarityLegendary,
		MinColors: 12,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateWithOptions("cyberpunk", 11111, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerate24Colors(b *testing.B) {
	gen := NewGenerator()
	opts := GenerationOptions{
		Harmony:   HarmonyTetradic,
		Mood:      MoodNormal,
		Rarity:    RarityCommon,
		MinColors: 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateWithOptions("horror", 22222, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
