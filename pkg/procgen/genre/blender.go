// Package genre provides genre blending for procedural content.
// This file implements genre blending which combines multiple genre
// characteristics for cross-genre content generation.
package genre

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// BlendedGenre represents a genre created by blending two base genres.
type BlendedGenre struct {
	*Genre
	PrimaryBase   *Genre
	SecondaryBase *Genre
	BlendWeight   float64 // 0.0 (all primary) to 1.0 (all secondary)
}

// GenreBlender creates blended genres from two base genres.
type GenreBlender struct {
	registry *Registry
}

// NewGenreBlender creates a new genre blender with the default registry.
func NewGenreBlender(registry *Registry) *GenreBlender {
	if registry == nil {
		registry = DefaultRegistry()
	}
	return &GenreBlender{
		registry: registry,
	}
}

// Blend creates a new genre by blending two existing genres.
// weight determines the blend ratio: 0.0 = all primary, 0.5 = equal, 1.0 = all secondary
// seed is used for deterministic selection of themes and other properties.
func (gb *GenreBlender) Blend(primaryID, secondaryID string, weight float64, seed int64) (*BlendedGenre, error) {
	// Validate weight
	if weight < 0.0 || weight > 1.0 {
		return nil, fmt.Errorf("blend weight must be between 0.0 and 1.0, got %f", weight)
	}

	// Get base genres
	primary, err := gb.registry.Get(primaryID)
	if err != nil {
		return nil, fmt.Errorf("primary genre: %w", err)
	}

	secondary, err := gb.registry.Get(secondaryID)
	if err != nil {
		return nil, fmt.Errorf("secondary genre: %w", err)
	}

	// Don't blend a genre with itself
	if primaryID == secondaryID {
		return nil, fmt.Errorf("cannot blend genre with itself")
	}

	rng := rand.New(rand.NewSource(seed))

	// Create blended genre
	blended := &Genre{
		ID:             generateBlendedID(primary, secondary, weight),
		Name:           generateBlendedName(primary, secondary, weight),
		Description:    generateBlendedDescription(primary, secondary, weight),
		Themes:         blendThemes(primary.Themes, secondary.Themes, weight, rng),
		PrimaryColor:   blendColor(primary.PrimaryColor, secondary.PrimaryColor, weight),
		SecondaryColor: blendColor(primary.SecondaryColor, secondary.SecondaryColor, weight),
		AccentColor:    blendColor(primary.AccentColor, secondary.AccentColor, weight),
		EntityPrefix:   selectPrefix(primary.EntityPrefix, secondary.EntityPrefix, weight, rng),
		ItemPrefix:     selectPrefix(primary.ItemPrefix, secondary.ItemPrefix, weight, rng),
		LocationPrefix: selectPrefix(primary.LocationPrefix, secondary.LocationPrefix, weight, rng),
	}

	return &BlendedGenre{
		Genre:         blended,
		PrimaryBase:   primary,
		SecondaryBase: secondary,
		BlendWeight:   weight,
	}, nil
}

// generateBlendedID creates a unique ID for the blended genre.
func generateBlendedID(primary, secondary *Genre, weight float64) string {
	// Order genres alphabetically for consistency
	if primary.ID > secondary.ID {
		primary, secondary = secondary, primary
		weight = 1.0 - weight
	}

	// Create a descriptive ID
	weightPercent := int(weight * 100)
	return fmt.Sprintf("%s-%s-%d", primary.ID, secondary.ID, weightPercent)
}

// generateBlendedName creates a human-readable name for the blended genre.
func generateBlendedName(primary, secondary *Genre, weight float64) string {
	// Determine which genre to put first based on weight
	if weight < 0.5 {
		return fmt.Sprintf("%s-%s", primary.Name, secondary.Name)
	} else if weight > 0.5 {
		return fmt.Sprintf("%s-%s", secondary.Name, primary.Name)
	}
	// Equal blend
	return fmt.Sprintf("%s/%s", primary.Name, secondary.Name)
}

// generateBlendedDescription creates a description for the blended genre.
func generateBlendedDescription(primary, secondary *Genre, weight float64) string {
	if weight < 0.33 {
		return fmt.Sprintf("%s with elements of %s", primary.Description, strings.ToLower(secondary.Name))
	} else if weight > 0.67 {
		return fmt.Sprintf("%s with elements of %s", secondary.Description, strings.ToLower(primary.Name))
	}
	return fmt.Sprintf("A blend of %s and %s themes", strings.ToLower(primary.Name), strings.ToLower(secondary.Name))
}

// blendThemes combines themes from both genres based on weight.
func blendThemes(primary, secondary []string, weight float64, rng *rand.Rand) []string {
	// Calculate how many themes to take from each
	totalThemes := 6 // Target number of themes for blended genre
	primaryCount := int(float64(totalThemes) * (1.0 - weight))
	secondaryCount := totalThemes - primaryCount

	// Ensure at least one theme from each
	if primaryCount == 0 && len(primary) > 0 {
		primaryCount = 1
		secondaryCount--
	}
	if secondaryCount == 0 && len(secondary) > 0 {
		secondaryCount = 1
		primaryCount--
	}

	result := make([]string, 0, totalThemes)

	// Select themes from primary
	if primaryCount > 0 && len(primary) > 0 {
		selected := selectRandomThemes(primary, primaryCount, rng)
		result = append(result, selected...)
	}

	// Select themes from secondary
	if secondaryCount > 0 && len(secondary) > 0 {
		selected := selectRandomThemes(secondary, secondaryCount, rng)
		result = append(result, selected...)
	}

	return result
}

// selectRandomThemes randomly selects n themes from the list without duplicates.
func selectRandomThemes(themes []string, n int, rng *rand.Rand) []string {
	if n >= len(themes) {
		// Return all themes if n is larger than available
		return append([]string{}, themes...)
	}

	// Fisher-Yates shuffle and take first n
	shuffled := append([]string{}, themes...)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:n]
}

// blendColor blends two hex colors based on weight.
func blendColor(color1, color2 string, weight float64) string {
	// Parse colors
	r1, g1, b1 := parseHexColor(color1)
	r2, g2, b2 := parseHexColor(color2)

	// Blend
	r := int(float64(r1)*(1.0-weight) + float64(r2)*weight)
	g := int(float64(g1)*(1.0-weight) + float64(g2)*weight)
	b := int(float64(b1)*(1.0-weight) + float64(b2)*weight)

	// Format as hex
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// parseHexColor parses a hex color string into RGB components.
func parseHexColor(hex string) (r, g, b int) {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Parse hex values
	if len(hex) == 6 {
		r64, _ := strconv.ParseInt(hex[0:2], 16, 0)
		g64, _ := strconv.ParseInt(hex[2:4], 16, 0)
		b64, _ := strconv.ParseInt(hex[4:6], 16, 0)
		r, g, b = int(r64), int(g64), int(b64)
	}

	return r, g, b
}

// selectPrefix selects a prefix based on weight and randomness.
func selectPrefix(prefix1, prefix2 string, weight float64, rng *rand.Rand) string {
	// Use probabilistic selection based on weight
	if rng.Float64() < (1.0 - weight) {
		return prefix1
	}
	return prefix2
}

// IsBlended returns true if the genre is a blended genre.
func (bg *BlendedGenre) IsBlended() bool {
	return bg.PrimaryBase != nil && bg.SecondaryBase != nil
}

// GetBaseGenres returns the base genres used to create this blended genre.
func (bg *BlendedGenre) GetBaseGenres() (*Genre, *Genre) {
	return bg.PrimaryBase, bg.SecondaryBase
}

// PresetBlends returns common preset blended genres.
func PresetBlends() map[string]struct {
	Primary   string
	Secondary string
	Weight    float64
} {
	return map[string]struct {
		Primary   string
		Secondary string
		Weight    float64
	}{
		"sci-fi-horror": {
			Primary:   "scifi",
			Secondary: "horror",
			Weight:    0.5,
		},
		"dark-fantasy": {
			Primary:   "fantasy",
			Secondary: "horror",
			Weight:    0.3,
		},
		"post-apoc-scifi": {
			Primary:   "postapoc",
			Secondary: "scifi",
			Weight:    0.5,
		},
		"cyber-horror": {
			Primary:   "cyberpunk",
			Secondary: "horror",
			Weight:    0.4,
		},
		"wasteland-fantasy": {
			Primary:   "postapoc",
			Secondary: "fantasy",
			Weight:    0.6,
		},
	}
}

// CreatePresetBlend creates a blended genre from a preset.
func (gb *GenreBlender) CreatePresetBlend(presetName string, seed int64) (*BlendedGenre, error) {
	presets := PresetBlends()
	preset, exists := presets[presetName]
	if !exists {
		return nil, fmt.Errorf("preset blend '%s' not found", presetName)
	}

	return gb.Blend(preset.Primary, preset.Secondary, preset.Weight, seed)
}
