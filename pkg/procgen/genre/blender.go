// Package genre provides genre blending for procedural content.
// This file implements genre blending which combines multiple genre
// characteristics for cross-genre content generation.
// Code relocated: Utility functions moved to blender_utils.go
package genre

import (
	"fmt"
	"math/rand"
)

// BlendedGenre represents a genre created by blending two base genres.
// BlendedGenre represents a genre formed by blending two base genres.
type BlendedGenre struct {
	// Genre is the fully merged genre with combined properties.
	*Genre
	// PrimaryBase is the dominant base genre (contributes 1-BlendWeight of properties).
	PrimaryBase *Genre
	// SecondaryBase is the secondary base genre (contributes BlendWeight of properties).
	SecondaryBase *Genre
	// BlendWeight controls the mix ratio: 0.0 means all primary, 1.0 means all secondary.
	BlendWeight float64 // 0.0 (all primary) to 1.0 (all secondary)
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
