// Package engine provides the TerrainReflectionTintSystem which applies subtle
// color tints to entity sprites based on the terrain tile type they stand on.
// Water tiles add a cool blue tint, lava adds warm orange, trees add green,
// and structures add a grey stone tint. Genre presets modulate tint intensity
// so fantasy worlds have vivid reflections while horror worlds use muted,
// desaturated tones. Tints compose multiplicatively with existing weather,
// light, and status effect tints in the render pipeline.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainReflectionTintComponent stores terrain-driven tint multipliers for an entity.
// These multiply with other tint sources (weather, light, status) in the render pipeline.
type TerrainReflectionTintComponent struct {
	// RGB multipliers (1.0 = no change). Values < 1.0 darken, > 1.0 brighten.
	TintR float64
	TintG float64
	TintB float64

	// The tile type that produced the current tint (for cache invalidation)
	LastTileType int
}

// Type returns the component type identifier.
func (t *TerrainReflectionTintComponent) Type() string {
	return "terrain_reflection_tint"
}

// NewTerrainReflectionTintComponent creates a component with neutral (no-op) tint.
func NewTerrainReflectionTintComponent() *TerrainReflectionTintComponent {
	return &TerrainReflectionTintComponent{
		TintR:        1.0,
		TintG:        1.0,
		TintB:        1.0,
		LastTileType: -1,
	}
}

// terrainTintColor holds RGB multipliers for a single terrain type.
type terrainTintColor struct {
	R, G, B float64
}

// genreTerrainIntensity holds per-genre intensity scaling for terrain reflections.
type genreTerrainIntensity struct {
	Intensity  float64 // Overall tint strength multiplier (0.0–1.0)
	Saturation float64 // Color saturation multiplier (0.0–1.0)
}

// TerrainReflectionTintSystem reads entity positions, looks up the terrain tile
// beneath them, and writes genre-aware tint values to TerrainReflectionTintComponent.
type TerrainReflectionTintSystem struct {
	world   *World
	terr    *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// Tile size in pixels (default 32)
	tileSize int

	// Throttle: only update when entity moves to a new tile
	lastTileCache map[uint64]int // entityID -> last tile type as int

	// Pre-computed terrain color map
	tileColors map[terrain.TileType]terrainTintColor

	// Pre-computed genre intensity presets
	genrePresets map[string]genreTerrainIntensity
}

// NewTerrainReflectionTintSystem creates a new terrain reflection tint system.
func NewTerrainReflectionTintSystem(world *World, seed int64) *TerrainReflectionTintSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_reflection_tint")
	}

	sys := &TerrainReflectionTintSystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		genreID:       "fantasy",
		tileSize:      32,
		lastTileCache: make(map[uint64]int, 64),
		tileColors:    buildTerrainTintColors(),
		genrePresets:  buildGenreTerrainPresets(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainReflectionTintSystem created")
	}

	return sys
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainReflectionTintSystem) SetTerrain(terr *terrain.Terrain) {
	s.terr = terr
	s.lastTileCache = make(map[uint64]int, 64)
}

// SetGenre sets the genre for intensity and saturation modulation.
func (s *TerrainReflectionTintSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainReflectionTintSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// Update iterates entities with position and sprite components, looks up terrain,
// and writes tint values to TerrainReflectionTintComponent.
func (s *TerrainReflectionTintSystem) Update(entities []*Entity, _ float64) {
	if s.terr == nil {
		return
	}

	preset, ok := s.genrePresets[s.genreID]
	if !ok {
		preset = s.genrePresets["fantasy"]
	}

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		if !entity.HasComponent("position") || !entity.HasComponent("sprite") {
			continue
		}

		// Use the hot-path cached position accessor instead of the generic map lookup.
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		tileX := int(pos.X) / s.tileSize
		tileY := int(pos.Y) / s.tileSize
		tileType := s.terr.GetTile(tileX, tileY)
		tileInt := int(tileType)

		// Skip if entity is on the same tile type as last frame
		if last, cached := s.lastTileCache[entity.ID]; cached && last == tileInt {
			continue
		}
		s.lastTileCache[entity.ID] = tileInt

		// Get or create tint component
		tintComp := s.getOrCreateTintComponent(entity)

		// Look up base color for this tile type
		color, hasColor := s.tileColors[tileType]
		if !hasColor {
			// Neutral terrain (floor, corridor, door) — reset to no tint
			tintComp.TintR = 1.0
			tintComp.TintG = 1.0
			tintComp.TintB = 1.0
			tintComp.LastTileType = tileInt
			continue
		}

		// Apply genre intensity and saturation
		tintComp.TintR = applyTerrainTint(color.R, preset.Intensity, preset.Saturation)
		tintComp.TintG = applyTerrainTint(color.G, preset.Intensity, preset.Saturation)
		tintComp.TintB = applyTerrainTint(color.B, preset.Intensity, preset.Saturation)
		tintComp.LastTileType = tileInt
	}
}

// getOrCreateTintComponent retrieves or attaches a TerrainReflectionTintComponent.
func (s *TerrainReflectionTintSystem) getOrCreateTintComponent(entity *Entity) *TerrainReflectionTintComponent {
	if comp, ok := entity.GetComponent("terrain_reflection_tint"); ok {
		if tint, ok := comp.(*TerrainReflectionTintComponent); ok {
			return tint
		}
	}
	tint := NewTerrainReflectionTintComponent()
	entity.AddComponent(tint)
	return tint
}

// applyTerrainTint blends a base tint color toward 1.0 (neutral) using intensity and saturation.
// A baseColor of 0.9 with intensity 0.5 yields 0.95 (halfway between 0.9 and 1.0).
// Saturation further mutes the effect toward neutral.
func applyTerrainTint(baseColor, intensity, saturation float64) float64 {
	// Deviation from neutral
	deviation := baseColor - 1.0
	// Scale by intensity and saturation
	scaled := deviation * intensity * saturation
	return 1.0 + scaled
}

// buildTerrainTintColors returns the base tint RGB multipliers for terrain types
// that should tint entities standing on them.
func buildTerrainTintColors() map[terrain.TileType]terrainTintColor {
	return map[terrain.TileType]terrainTintColor{
		// Water adds cool blue tint
		terrain.TileWaterShallow: {R: 0.90, G: 0.93, B: 1.08},
		terrain.TileWaterDeep:    {R: 0.85, G: 0.88, B: 1.12},

		// Lava adds warm orange-red tint
		terrain.TileLavaFlow: {R: 1.15, G: 0.90, B: 0.80},

		// Trees add subtle green-brown tint from canopy filtering
		terrain.TileTree: {R: 0.92, G: 1.05, B: 0.90},

		// Bridges have a slight warm wood tint
		terrain.TileBridge: {R: 1.03, G: 0.98, B: 0.92},

		// Structures have a cool stone-grey tint
		terrain.TileStructure: {R: 0.95, G: 0.95, B: 0.97},

		// Platforms have a slight elevation-brightness boost
		terrain.TilePlatform: {R: 1.03, G: 1.03, B: 1.05},

		// Pits darken entities at the edge
		terrain.TilePit: {R: 0.88, G: 0.88, B: 0.90},

		// Ramps have a subtle transition tint
		terrain.TileRamp:     {R: 0.97, G: 0.97, B: 0.99},
		terrain.TileRampUp:   {R: 1.02, G: 1.02, B: 1.04},
		terrain.TileRampDown: {R: 0.94, G: 0.94, B: 0.96},

		// Trap doors have a faint danger tint
		terrain.TileTrapDoor: {R: 1.04, G: 0.95, B: 0.90},
	}
}

// buildGenreTerrainPresets returns per-genre intensity and saturation modifiers.
func buildGenreTerrainPresets() map[string]genreTerrainIntensity {
	return map[string]genreTerrainIntensity{
		// Fantasy: vivid, saturated terrain reflections
		"fantasy": {Intensity: 0.7, Saturation: 0.9},
		// Sci-fi: clean, moderate reflections
		"scifi": {Intensity: 0.5, Saturation: 0.7},
		// Horror: muted, desaturated—everything looks bleak
		"horror": {Intensity: 0.4, Saturation: 0.4},
		// Cyberpunk: high-contrast neon-ish reflections
		"cyberpunk": {Intensity: 0.6, Saturation: 0.8},
		// Post-apocalyptic: dusty, washed-out tints
		"postapoc": {Intensity: 0.45, Saturation: 0.5},
	}
}
