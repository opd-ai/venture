package genre

import (
	"math/rand"
	"strings"
	"testing"
)

func TestNewGenreBlender(t *testing.T) {
	tests := []struct {
		name     string
		registry *Registry
		wantNil  bool
	}{
		{
			name:     "with default registry",
			registry: DefaultRegistry(),
			wantNil:  false,
		},
		{
			name:     "with nil registry",
			registry: nil,
			wantNil:  false, // Should create default
		},
		{
			name:     "with custom registry",
			registry: NewRegistry(),
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blender := NewGenreBlender(tt.registry)
			if (blender == nil) != tt.wantNil {
				t.Errorf("NewGenreBlender() = %v, wantNil %v", blender, tt.wantNil)
			}
			if blender != nil && blender.registry == nil {
				t.Error("GenreBlender.registry should not be nil")
			}
		})
	}
}

func TestGenreBlender_Blend(t *testing.T) {
	blender := NewGenreBlender(DefaultRegistry())

	tests := []struct {
		name        string
		primaryID   string
		secondaryID string
		weight      float64
		seed        int64
		wantErr     bool
	}{
		{
			name:        "fantasy-scifi equal blend",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      0.5,
			seed:        12345,
			wantErr:     false,
		},
		{
			name:        "fantasy-scifi primary heavy",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      0.2,
			seed:        12345,
			wantErr:     false,
		},
		{
			name:        "fantasy-scifi secondary heavy",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      0.8,
			seed:        12345,
			wantErr:     false,
		},
		{
			name:        "horror-cyberpunk blend",
			primaryID:   "horror",
			secondaryID: "cyberpunk",
			weight:      0.5,
			seed:        54321,
			wantErr:     false,
		},
		{
			name:        "invalid primary genre",
			primaryID:   "nonexistent",
			secondaryID: "scifi",
			weight:      0.5,
			seed:        12345,
			wantErr:     true,
		},
		{
			name:        "invalid secondary genre",
			primaryID:   "fantasy",
			secondaryID: "nonexistent",
			weight:      0.5,
			seed:        12345,
			wantErr:     true,
		},
		{
			name:        "same genre blend",
			primaryID:   "fantasy",
			secondaryID: "fantasy",
			weight:      0.5,
			seed:        12345,
			wantErr:     true,
		},
		{
			name:        "weight too low",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      -0.1,
			seed:        12345,
			wantErr:     true,
		},
		{
			name:        "weight too high",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      1.1,
			seed:        12345,
			wantErr:     true,
		},
		{
			name:        "zero weight",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      0.0,
			seed:        12345,
			wantErr:     false,
		},
		{
			name:        "max weight",
			primaryID:   "fantasy",
			secondaryID: "scifi",
			weight:      1.0,
			seed:        12345,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blended, err := blender.Blend(tt.primaryID, tt.secondaryID, tt.weight, tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenreBlender.Blend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if blended == nil {
					t.Fatal("Expected blended genre, got nil")
				}
				if blended.Genre == nil {
					t.Fatal("Expected blended.Genre to not be nil")
				}
				if blended.PrimaryBase == nil {
					t.Error("Expected PrimaryBase to not be nil")
				}
				if blended.SecondaryBase == nil {
					t.Error("Expected SecondaryBase to not be nil")
				}
				if blended.BlendWeight != tt.weight {
					t.Errorf("Expected BlendWeight %f, got %f", tt.weight, blended.BlendWeight)
				}

				// Validate blended genre
				if err := blended.Genre.Validate(); err != nil {
					t.Errorf("Blended genre validation failed: %v", err)
				}

				// Check that ID is generated
				if blended.ID == "" {
					t.Error("Blended genre ID should not be empty")
				}

				// Check that themes are combined
				if len(blended.Themes) == 0 {
					t.Error("Blended genre should have themes")
				}
			}
		})
	}
}

func TestGenreBlender_BlendDeterminism(t *testing.T) {
	blender := NewGenreBlender(DefaultRegistry())

	// Generate same blend twice with same seed
	seed := int64(12345)
	blend1, err1 := blender.Blend("fantasy", "scifi", 0.5, seed)
	blend2, err2 := blender.Blend("fantasy", "scifi", 0.5, seed)

	if err1 != nil || err2 != nil {
		t.Fatalf("Blend failed: err1=%v, err2=%v", err1, err2)
	}

	// Check determinism
	if blend1.ID != blend2.ID {
		t.Errorf("IDs differ: %s vs %s", blend1.ID, blend2.ID)
	}
	if blend1.Name != blend2.Name {
		t.Errorf("Names differ: %s vs %s", blend1.Name, blend2.Name)
	}
	if blend1.Description != blend2.Description {
		t.Errorf("Descriptions differ: %s vs %s", blend1.Description, blend2.Description)
	}

	// Themes should be identical with same seed
	if len(blend1.Themes) != len(blend2.Themes) {
		t.Errorf("Theme counts differ: %d vs %d", len(blend1.Themes), len(blend2.Themes))
	} else {
		for i := range blend1.Themes {
			if blend1.Themes[i] != blend2.Themes[i] {
				t.Errorf("Theme %d differs: %s vs %s", i, blend1.Themes[i], blend2.Themes[i])
			}
		}
	}
}

func TestGenreBlender_BlendVariety(t *testing.T) {
	blender := NewGenreBlender(DefaultRegistry())

	// Generate same blend with different seeds
	blend1, _ := blender.Blend("fantasy", "scifi", 0.5, 12345)
	blend2, _ := blender.Blend("fantasy", "scifi", 0.5, 54321)

	// IDs should be same (deterministic based on genres and weight)
	if blend1.ID != blend2.ID {
		t.Errorf("IDs should be same for same blend parameters: %s vs %s", blend1.ID, blend2.ID)
	}

	// But themes might differ (random selection from available themes)
	// This is expected behavior - same blend can have different theme selections
}

func TestBlendedGenre_IsBlended(t *testing.T) {
	blender := NewGenreBlender(DefaultRegistry())

	blended, err := blender.Blend("fantasy", "scifi", 0.5, 12345)
	if err != nil {
		t.Fatalf("Blend failed: %v", err)
	}

	if !blended.IsBlended() {
		t.Error("IsBlended() should return true for blended genre")
	}
}

func TestBlendedGenre_GetBaseGenres(t *testing.T) {
	blender := NewGenreBlender(DefaultRegistry())

	blended, err := blender.Blend("fantasy", "scifi", 0.5, 12345)
	if err != nil {
		t.Fatalf("Blend failed: %v", err)
	}

	primary, secondary := blended.GetBaseGenres()
	if primary == nil || secondary == nil {
		t.Fatal("GetBaseGenres() returned nil")
	}

	if primary.ID != "fantasy" && primary.ID != "scifi" {
		t.Errorf("Unexpected primary ID: %s", primary.ID)
	}
	if secondary.ID != "fantasy" && secondary.ID != "scifi" {
		t.Errorf("Unexpected secondary ID: %s", secondary.ID)
	}
}

func TestGenerateBlendedID(t *testing.T) {
	registry := DefaultRegistry()
	fantasy, _ := registry.Get("fantasy")
	scifi, _ := registry.Get("scifi")

	tests := []struct {
		name      string
		primary   *Genre
		secondary *Genre
		weight    float64
		wantPart  string
	}{
		{
			name:      "fantasy-scifi 50%",
			primary:   fantasy,
			secondary: scifi,
			weight:    0.5,
			wantPart:  "fantasy-scifi-50",
		},
		{
			name:      "scifi-fantasy 50%",
			primary:   scifi,
			secondary: fantasy,
			weight:    0.5,
			wantPart:  "fantasy-scifi-50", // Should normalize order
		},
		{
			name:      "fantasy-scifi 25%",
			primary:   fantasy,
			secondary: scifi,
			weight:    0.25,
			wantPart:  "25",
		},
		{
			name:      "fantasy-scifi 75%",
			primary:   fantasy,
			secondary: scifi,
			weight:    0.75,
			wantPart:  "75",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := generateBlendedID(tt.primary, tt.secondary, tt.weight)
			if !strings.Contains(id, tt.wantPart) {
				t.Errorf("generateBlendedID() = %s, want to contain %s", id, tt.wantPart)
			}
		})
	}
}

func TestGenerateBlendedName(t *testing.T) {
	registry := DefaultRegistry()
	fantasy, _ := registry.Get("fantasy")
	scifi, _ := registry.Get("scifi")

	tests := []struct {
		name      string
		primary   *Genre
		secondary *Genre
		weight    float64
		wantPart  string
	}{
		{
			name:      "primary heavy",
			primary:   fantasy,
			secondary: scifi,
			weight:    0.2,
			wantPart:  "Fantasy",
		},
		{
			name:      "equal blend",
			primary:   fantasy,
			secondary: scifi,
			weight:    0.5,
			wantPart:  "/",
		},
		{
			name:      "secondary heavy",
			primary:   fantasy,
			secondary: scifi,
			weight:    0.8,
			wantPart:  "Sci-Fi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := generateBlendedName(tt.primary, tt.secondary, tt.weight)
			if !strings.Contains(name, tt.wantPart) {
				t.Errorf("generateBlendedName() = %s, want to contain %s", name, tt.wantPart)
			}
		})
	}
}

func TestBlendColor(t *testing.T) {
	tests := []struct {
		name   string
		color1 string
		color2 string
		weight float64
		want   string
	}{
		{
			name:   "equal blend of red and blue",
			color1: "#FF0000",
			color2: "#0000FF",
			weight: 0.5,
			want:   "#7F007F",
		},
		{
			name:   "all color1",
			color1: "#FF0000",
			color2: "#0000FF",
			weight: 0.0,
			want:   "#FF0000",
		},
		{
			name:   "all color2",
			color1: "#FF0000",
			color2: "#0000FF",
			weight: 1.0,
			want:   "#0000FF",
		},
		{
			name:   "25% blend",
			color1: "#FF0000",
			color2: "#0000FF",
			weight: 0.25,
			want:   "#BF003F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blendColor(tt.color1, tt.color2, tt.weight)
			if got != tt.want {
				t.Errorf("blendColor() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		wantR int
		wantG int
		wantB int
	}{
		{
			name:  "red",
			hex:   "#FF0000",
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "green",
			hex:   "#00FF00",
			wantR: 0,
			wantG: 255,
			wantB: 0,
		},
		{
			name:  "blue",
			hex:   "#0000FF",
			wantR: 0,
			wantG: 0,
			wantB: 255,
		},
		{
			name:  "white",
			hex:   "#FFFFFF",
			wantR: 255,
			wantG: 255,
			wantB: 255,
		},
		{
			name:  "black",
			hex:   "#000000",
			wantR: 0,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "without hash",
			hex:   "FF00FF",
			wantR: 255,
			wantG: 0,
			wantB: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := parseHexColor(tt.hex)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("parseHexColor() = (%d, %d, %d), want (%d, %d, %d)",
					r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestPresetBlends(t *testing.T) {
	presets := PresetBlends()

	if len(presets) == 0 {
		t.Fatal("PresetBlends() returned empty map")
	}

	// Check that some expected presets exist
	expectedPresets := []string{"sci-fi-horror", "dark-fantasy", "cyber-horror"}
	for _, name := range expectedPresets {
		if _, exists := presets[name]; !exists {
			t.Errorf("Expected preset '%s' not found", name)
		}
	}

	// Validate preset structure
	for name, preset := range presets {
		if preset.Primary == "" {
			t.Errorf("Preset '%s' has empty Primary", name)
		}
		if preset.Secondary == "" {
			t.Errorf("Preset '%s' has empty Secondary", name)
		}
		if preset.Weight < 0.0 || preset.Weight > 1.0 {
			t.Errorf("Preset '%s' has invalid weight: %f", name, preset.Weight)
		}
	}
}

func TestGenreBlender_CreatePresetBlend(t *testing.T) {
	blender := NewGenreBlender(DefaultRegistry())

	tests := []struct {
		name       string
		presetName string
		seed       int64
		wantErr    bool
	}{
		{
			name:       "sci-fi horror",
			presetName: "sci-fi-horror",
			seed:       12345,
			wantErr:    false,
		},
		{
			name:       "dark fantasy",
			presetName: "dark-fantasy",
			seed:       12345,
			wantErr:    false,
		},
		{
			name:       "invalid preset",
			presetName: "nonexistent",
			seed:       12345,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blended, err := blender.CreatePresetBlend(tt.presetName, tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePresetBlend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if blended == nil {
					t.Fatal("Expected blended genre, got nil")
				}
				if !blended.IsBlended() {
					t.Error("Preset blend should be a blended genre")
				}
			}
		})
	}
}

func TestBlendThemes(t *testing.T) {
	tests := []struct {
		name      string
		primary   []string
		secondary []string
		weight    float64
		minCount  int
		maxCount  int
	}{
		{
			name:      "equal blend",
			primary:   []string{"magic", "dragons", "knights"},
			secondary: []string{"lasers", "robots", "space"},
			weight:    0.5,
			minCount:  6,
			maxCount:  6,
		},
		{
			name:      "primary heavy",
			primary:   []string{"magic", "dragons", "knights"},
			secondary: []string{"lasers", "robots", "space"},
			weight:    0.2,
			minCount:  4, // At least one from each
			maxCount:  6,
		},
		{
			name:      "secondary heavy",
			primary:   []string{"magic", "dragons", "knights"},
			secondary: []string{"lasers", "robots", "space"},
			weight:    0.8,
			minCount:  4, // At least one from each
			maxCount:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use fixed seed for determinism
			rng := rand.New(rand.NewSource(12345))
			themes := blendThemes(tt.primary, tt.secondary, tt.weight, rng)

			if len(themes) < tt.minCount || len(themes) > tt.maxCount {
				t.Errorf("blendThemes() returned %d themes, want between %d and %d",
					len(themes), tt.minCount, tt.maxCount)
			}

			// Check that themes come from both sources
			hasFromPrimary := false
			hasFromSecondary := false
			for _, theme := range themes {
				for _, p := range tt.primary {
					if theme == p {
						hasFromPrimary = true
					}
				}
				for _, s := range tt.secondary {
					if theme == s {
						hasFromSecondary = true
					}
				}
			}

			// For most weights, we should have themes from both
			if tt.weight > 0.1 && tt.weight < 0.9 {
				if !hasFromPrimary {
					t.Error("Expected themes from primary genre")
				}
				if !hasFromSecondary {
					t.Error("Expected themes from secondary genre")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkBlend(b *testing.B) {
	blender := NewGenreBlender(DefaultRegistry())
	for i := 0; i < b.N; i++ {
		_, _ = blender.Blend("fantasy", "scifi", 0.5, int64(i))
	}
}

func BenchmarkCreatePresetBlend(b *testing.B) {
	blender := NewGenreBlender(DefaultRegistry())
	for i := 0; i < b.N; i++ {
		_, _ = blender.CreatePresetBlend("sci-fi-horror", int64(i))
	}
}
