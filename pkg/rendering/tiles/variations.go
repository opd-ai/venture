// Package tiles provides tile variation generation.
// This file implements deterministic tile variation generation to increase
// visual variety while maintaining consistency within tile types.
package tiles

import (
	"fmt"
	"image"
)

// VariationSet represents a set of tile variations for a single tile type.
type VariationSet struct {
	Type       TileType
	Variations []*image.RGBA
	Count      int
}

// GenerateVariations creates multiple variations of a tile type.
// Each variation uses the same base configuration with a different seed offset.
func (g *Generator) GenerateVariations(config Config, count int) (*VariationSet, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if count <= 0 {
		return nil, fmt.Errorf("count must be positive, got %d", count)
	}

	variations := make([]*image.RGBA, count)

	for i := 0; i < count; i++ {
		// Use seed offset for deterministic variation
		variantConfig := config
		variantConfig.Seed = config.Seed + int64(i)*1000
		variantConfig.Variant = float64(i) / float64(count)

		img, err := g.Generate(variantConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate variation %d: %w", i, err)
		}

		variations[i] = img
	}

	return &VariationSet{
		Type:       config.Type,
		Variations: variations,
		Count:      count,
	}, nil
}

// GetVariation returns a specific variation from the set.
func (vs *VariationSet) GetVariation(index int) (*image.RGBA, error) {
	if index < 0 || index >= vs.Count {
		return nil, fmt.Errorf("variation index %d out of range [0, %d)", index, vs.Count)
	}
	return vs.Variations[index], nil
}

// GetVariationBySeed returns a variation deterministically based on a seed.
// This allows consistent tile selection for specific positions.
func (vs *VariationSet) GetVariationBySeed(seed int64) *image.RGBA {
	if vs.Count == 0 {
		return nil
	}
	index := int(seed % int64(vs.Count))
	if index < 0 {
		index = -index
	}
	return vs.Variations[index]
}

// TileSet contains variations for all tile types in a theme.
type TileSet struct {
	GenreID      string
	Seed         int64
	TileSize     int
	Variations   map[TileType]*VariationSet
	VariantCount int
}

// GenerateTileSet creates a complete set of tile variations for a genre.
func (g *Generator) GenerateTileSet(genreID string, seed int64, tileSize, variationCount int) (*TileSet, error) {
	if tileSize <= 0 {
		return nil, fmt.Errorf("tileSize must be positive, got %d", tileSize)
	}
	if variationCount <= 0 {
		return nil, fmt.Errorf("variationCount must be positive, got %d", variationCount)
	}

	tileTypes := []TileType{
		TileFloor,
		TileWall,
		TileDoor,
		TileCorridor,
		TileWater,
		TileLava,
		TileTrap,
		TileStairs,
	}

	variations := make(map[TileType]*VariationSet)

	for _, tileType := range tileTypes {
		config := Config{
			Type:    tileType,
			Width:   tileSize,
			Height:  tileSize,
			GenreID: genreID,
			Seed:    seed + int64(tileType)*10000,
			Variant: 0.5,
			Custom:  make(map[string]interface{}),
		}

		varSet, err := g.GenerateVariations(config, variationCount)
		if err != nil {
			return nil, fmt.Errorf("failed to generate variations for %s: %w", tileType, err)
		}

		variations[tileType] = varSet
	}

	return &TileSet{
		GenreID:      genreID,
		Seed:         seed,
		TileSize:     tileSize,
		Variations:   variations,
		VariantCount: variationCount,
	}, nil
}

// GetTile returns a tile image for a specific type and position.
// Uses position-based seed for deterministic variation selection.
func (ts *TileSet) GetTile(tileType TileType, x, y int) (*image.RGBA, error) {
	varSet, ok := ts.Variations[tileType]
	if !ok {
		return nil, fmt.Errorf("tile type %s not found in set", tileType)
	}

	// Use position to select variation deterministically
	positionSeed := ts.Seed + int64(x)*1000 + int64(y)
	return varSet.GetVariationBySeed(positionSeed), nil
}

// GetVariationSet returns all variations for a specific tile type.
func (ts *TileSet) GetVariationSet(tileType TileType) (*VariationSet, error) {
	varSet, ok := ts.Variations[tileType]
	if !ok {
		return nil, fmt.Errorf("tile type %s not found in set", tileType)
	}
	return varSet, nil
}

// ValidateTileSet checks if a tile set is complete and valid.
func ValidateTileSet(ts *TileSet) error {
	if ts == nil {
		return fmt.Errorf("tile set is nil")
	}

	if ts.GenreID == "" {
		return fmt.Errorf("genreID is empty")
	}

	if ts.TileSize <= 0 {
		return fmt.Errorf("tileSize must be positive, got %d", ts.TileSize)
	}

	if ts.VariantCount <= 0 {
		return fmt.Errorf("variantCount must be positive, got %d", ts.VariantCount)
	}

	requiredTypes := []TileType{
		TileFloor,
		TileWall,
		TileDoor,
		TileCorridor,
		TileWater,
		TileLava,
		TileTrap,
		TileStairs,
	}

	for _, tileType := range requiredTypes {
		varSet, ok := ts.Variations[tileType]
		if !ok {
			return fmt.Errorf("missing variations for tile type %s", tileType)
		}

		if varSet.Count != ts.VariantCount {
			return fmt.Errorf("tile type %s has %d variations, expected %d",
				tileType, varSet.Count, ts.VariantCount)
		}

		for i, img := range varSet.Variations {
			if img == nil {
				return fmt.Errorf("variation %d for tile type %s is nil", i, tileType)
			}

			bounds := img.Bounds()
			if bounds.Dx() != ts.TileSize || bounds.Dy() != ts.TileSize {
				return fmt.Errorf("variation %d for tile type %s has size %dx%d, expected %dx%d",
					i, tileType, bounds.Dx(), bounds.Dy(), ts.TileSize, ts.TileSize)
			}
		}
	}

	return nil
}
