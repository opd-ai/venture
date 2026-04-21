// Package terrain provides composite terrain generation.
// This file implements the CompositeGenerator which combines multiple
// biome types (dungeon, cave, forest, city, maze) into a single level.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// CompositeGenerator combines multiple terrain biomes in a single level.
// Uses Voronoi partitioning to create distinct regions with smooth transitions.
type CompositeGenerator struct {
	biomeCount      int                          // Number of biomes to combine (2-4)
	transitionWidth int                          // Width of transition zones in tiles (2-4)
	generators      map[string]procgen.Generator // Available generators by name
	logger          *logrus.Entry
}

// NewCompositeGenerator creates a new composite terrain generator.
func NewCompositeGenerator() *CompositeGenerator {
	return NewCompositeGeneratorWithLogger(nil)
}

// NewCompositeGeneratorWithLogger creates a new composite terrain generator with a logger.
func NewCompositeGeneratorWithLogger(logger *logrus.Logger) *CompositeGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "composite")
	}
	return &CompositeGenerator{
		biomeCount:      3, // Default: 3 biomes
		transitionWidth: 3, // Default: 3-tile transitions
		generators: map[string]procgen.Generator{
			"bsp":      NewBSPGenerator(),
			"cellular": NewCellularGenerator(),
			"maze":     NewMazeGenerator(),
			"forest":   NewForestGenerator(),
			"city":     NewCityGenerator(),
		},
		logger: logEntry,
	}
}

// BiomeRegionInfo contains information about a generated biome region.
type BiomeRegionInfo struct {
	ID            int            // Region identifier
	GeneratorName string         // Generator used ("bsp", "cellular", etc.)
	BiomeType     BiomeType      // Biome type for transitions
	Seed          int64          // Seed used for this region
	Bounds        *VoronoiRegion // Region tiles and bounds
}

// extractParameters extracts generation parameters from custom params.
func (g *CompositeGenerator) extractParameters(params procgen.GenerationParams) (int, int, int, int) {
	width := 80
	height := 50
	biomeCount := g.biomeCount
	transitionWidth := g.transitionWidth

	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok {
			width = w
		}
		if h, ok := params.Custom["height"].(int); ok {
			height = h
		}
		if bc, ok := params.Custom["biomeCount"].(int); ok {
			biomeCount = bc
		}
		if tw, ok := params.Custom["transitionWidth"].(int); ok {
			transitionWidth = tw
		}
	}

	return width, height, biomeCount, transitionWidth
}

// validateAndClampParameters validates dimensions and clamps biome/transition values.
func (g *CompositeGenerator) validateAndClampParameters(width, height, biomeCount, transitionWidth int) (int, int, int, int, error) {
	if width <= 0 || height <= 0 {
		return 0, 0, 0, 0, fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}
	if width < 60 || height < 40 {
		return 0, 0, 0, 0, fmt.Errorf("dimensions too small for composite generation: width=%d, height=%d (min 60x40)", width, height)
	}
	if width > 500 || height > 500 {
		return 0, 0, 0, 0, fmt.Errorf("dimensions too large: width=%d, height=%d (max 500x500)", width, height)
	}

	if biomeCount < 2 || biomeCount > 4 {
		biomeCount = 3
	}
	if transitionWidth < 1 || transitionWidth > 5 {
		transitionWidth = 3
	}

	return width, height, biomeCount, transitionWidth, nil
}

// createBiomeRegions creates biome region information structures.
func (g *CompositeGenerator) createBiomeRegions(seed int64, biomeCount int, generatorNames []string, diagram *VoronoiDiagram) []*BiomeRegionInfo {
	biomeRegions := make([]*BiomeRegionInfo, biomeCount)
	seedGen := procgen.NewSeedGenerator(seed)

	for i := 0; i < biomeCount; i++ {
		biomeRegions[i] = &BiomeRegionInfo{
			ID:            i,
			GeneratorName: generatorNames[i],
			BiomeType:     GetBiomeType(generatorNames[i]),
			Seed:          seedGen.GetSeed("biome", i),
			Bounds:        diagram.Regions[i],
		}
	}
	return biomeRegions
}

// applyPostProcessing applies transitions, connectivity, and stairs to the terrain.
func (g *CompositeGenerator) applyPostProcessing(terrain *Terrain, diagram *VoronoiDiagram, biomeRegions []*BiomeRegionInfo, transitionWidth, depth int, rng *rand.Rand) error {
	biomeTypes := make(map[int]BiomeType)
	for _, region := range biomeRegions {
		biomeTypes[region.ID] = region.BiomeType
	}

	BlendTransitionZones(terrain, diagram, biomeTypes, transitionWidth, rng)

	if err := g.ensureConnectivity(terrain, diagram, biomeRegions, rng); err != nil {
		return fmt.Errorf("failed to ensure connectivity: %w", err)
	}

	if depth > 0 {
		g.placeStairs(terrain, diagram, biomeRegions, rng)
	}

	return nil
}

// Generate creates composite terrain by combining multiple biome generators.
func (g *CompositeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting composite terrain generation")
	}

	// Extract parameters and validate/clamp them in sequence
	rawWidth, rawHeight, rawBiomeCount, rawTransitionWidth := g.extractParameters(params)
	width, height, biomeCount, transitionWidth, err := g.validateAndClampParameters(rawWidth, rawHeight, rawBiomeCount, rawTransitionWidth)
	if err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(seed))
	terrain := NewTerrain(width, height, seed)
	diagram := GenerateVoronoiDiagram(width, height, biomeCount, rng)
	generatorNames := g.selectGenerators(params.GenreID, biomeCount, rng)
	biomeRegions := g.createBiomeRegions(seed, biomeCount, generatorNames, diagram)

	for _, region := range biomeRegions {
		if err := g.generateBiomeRegion(terrain, region, diagram, params, rng); err != nil {
			return nil, fmt.Errorf("failed to generate biome %d (%s): %w", region.ID, region.GeneratorName, err)
		}
	}

	if err := g.applyPostProcessing(terrain, diagram, biomeRegions, transitionWidth, params.Depth, rng); err != nil {
		return nil, err
	}

	terrain.Rooms = make([]*Room, 0)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":        terrain.Width,
			"height":       terrain.Height,
			"biomeCount":   len(biomeRegions),
			"biomeRegions": len(biomeRegions),
		}).Info("composite terrain generation complete")
	}

	return terrain, nil
}

// GenerateCached generates terrain with caching support.
// Uses the default terrain cache to store and retrieve terrains by seed/params.
// Returns cached terrain if available (near-instant), otherwise generates and caches.
func (g *CompositeGenerator) GenerateCached(seed int64, params procgen.GenerationParams) (*Terrain, error) {
	// Check cache first
	if cached := GetCached(seed, params); cached != nil {
		if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
			g.logger.WithFields(logrus.Fields{
				"seed":    seed,
				"genreID": params.GenreID,
				"cached":  true,
			}).Debug("returning cached terrain")
		}
		return cached, nil
	}

	// Generate new terrain
	result, err := g.Generate(seed, params)
	if err != nil {
		return nil, err
	}

	terrain := result.(*Terrain)

	// Store in cache
	PutCached(seed, params, terrain)

	return terrain, nil
}

// selectGenerators chooses which generators to use based on genre and count.
func (g *CompositeGenerator) selectGenerators(genreID string, count int, rng *rand.Rand) []string {
	// Define genre preferences for generator selection
	genrePrefs := map[string][]string{
		"fantasy":         {"bsp", "cellular", "forest"},
		"scifi":           {"city", "maze", "bsp"},
		"horror":          {"cellular", "maze", "bsp"},
		"cyberpunk":       {"city", "maze", "bsp"},
		"postapoc": {"cellular", "city", "forest"},
	}

	// Get preferred generators for this genre
	prefs, ok := genrePrefs[genreID]
	if !ok {
		prefs = []string{"bsp", "cellular", "maze", "forest", "city"}
	}

	// Select generators (avoid duplicates)
	selected := make([]string, 0, count)
	used := make(map[string]bool)

	// First, use genre preferences
	for _, gen := range prefs {
		if len(selected) >= count {
			break
		}
		if !used[gen] {
			selected = append(selected, gen)
			used[gen] = true
		}
	}

	// If we need more, add remaining generators randomly
	allGens := []string{"bsp", "cellular", "maze", "forest", "city"}
	for len(selected) < count {
		gen := allGens[rng.Intn(len(allGens))]
		if !used[gen] {
			selected = append(selected, gen)
			used[gen] = true
		}
	}

	// Shuffle for variety
	rng.Shuffle(len(selected), func(i, j int) {
		selected[i], selected[j] = selected[j], selected[i]
	})

	return selected
}

// generateBiomeRegion generates terrain for a specific biome region.
func (g *CompositeGenerator) generateBiomeRegion(terrain *Terrain, region *BiomeRegionInfo, diagram *VoronoiDiagram, params procgen.GenerationParams, rng *rand.Rand) error {
	generator, ok := g.generators[region.GeneratorName]
	if !ok {
		return fmt.Errorf("unknown generator: %s", region.GeneratorName)
	}

	minX, minY, maxX, maxY := diagram.GetRegionBounds(region.ID)
	regionWidth, regionHeight, err := g.calculateRegionDimensions(minX, minY, maxX, maxY, terrain.Width, terrain.Height)
	if err != nil {
		return err
	}

	regionParams := g.buildRegionParams(params, regionWidth, regionHeight)
	biomeTerrain, err := g.generateBiomeTerrain(generator, region, regionParams)
	if err != nil {
		return err
	}

	g.copyBiomeTilesToTerrain(terrain, biomeTerrain, region, minX, minY, regionWidth, regionHeight)
	return nil
}

// calculateRegionDimensions computes and validates region dimensions with minimum sizes.
func (g *CompositeGenerator) calculateRegionDimensions(minX, minY, maxX, maxY, terrainWidth, terrainHeight int) (width, height int, err error) {
	width = maxX - minX + 1
	height = maxY - minY + 1

	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid region bounds: width=%d, height=%d", width, height)
	}

	width = g.enforceMinimumDimension(width, 20)
	height = g.enforceMinimumDimension(height, 15)

	if minX+width > terrainWidth {
		width = terrainWidth - minX
	}
	if minY+height > terrainHeight {
		height = terrainHeight - minY
	}

	return width, height, nil
}

// enforceMinimumDimension ensures dimension meets minimum threshold.
func (g *CompositeGenerator) enforceMinimumDimension(dimension, minimum int) int {
	if dimension < minimum {
		return minimum
	}
	return dimension
}

// buildRegionParams creates generation parameters for the biome region.
func (g *CompositeGenerator) buildRegionParams(params procgen.GenerationParams, width, height int) procgen.GenerationParams {
	regionParams := procgen.GenerationParams{
		Difficulty: params.Difficulty,
		Depth:      params.Depth,
		GenreID:    params.GenreID,
		Custom:     make(map[string]interface{}),
	}

	if params.Custom != nil {
		for k, v := range params.Custom {
			regionParams.Custom[k] = v
		}
	}

	regionParams.Custom["width"] = width
	regionParams.Custom["height"] = height

	return regionParams
}

// generateBiomeTerrain generates terrain for the biome using the appropriate generator.
func (g *CompositeGenerator) generateBiomeTerrain(generator procgen.Generator, region *BiomeRegionInfo, params procgen.GenerationParams) (*Terrain, error) {
	result, err := generator.Generate(region.Seed, params)
	if err != nil {
		return nil, fmt.Errorf("generator %s failed: %w", region.GeneratorName, err)
	}

	biomeTerrain, ok := result.(*Terrain)
	if !ok {
		return nil, fmt.Errorf("generator %s returned non-terrain result", region.GeneratorName)
	}

	return biomeTerrain, nil
}

// copyBiomeTilesToTerrain copies generated biome tiles to the main terrain.
func (g *CompositeGenerator) copyBiomeTilesToTerrain(terrain, biomeTerrain *Terrain, region *BiomeRegionInfo, minX, minY, regionWidth, regionHeight int) {
	for _, tile := range region.Bounds.Tiles {
		biomeX := tile.X - minX
		biomeY := tile.Y - minY

		if biomeX >= 0 && biomeX < regionWidth && biomeY >= 0 && biomeY < regionHeight {
			srcTile := biomeTerrain.GetTile(biomeX, biomeY)
			terrain.SetTile(tile.X, tile.Y, srcTile)
		}
	}
}

// ensureConnectivity ensures all biome regions are connected.
// Uses flood fill to detect disconnected regions and carves corridors if needed.
func (g *CompositeGenerator) ensureConnectivity(terrain *Terrain, diagram *VoronoiDiagram, regions []*BiomeRegionInfo, rng *rand.Rand) error {
	// Find one walkable tile in each region as a connection point
	regionConnections := make([]Point, len(regions))

	for i, region := range regions {
		// Find first walkable tile in this region
		found := false
		for _, tile := range region.Bounds.Tiles {
			if terrain.IsWalkable(tile.X, tile.Y) {
				regionConnections[i] = tile
				found = true
				break
			}
		}

		if !found {
			// No walkable tiles, make one
			if len(region.Bounds.Tiles) > 0 {
				centerTile := region.Bounds.Tiles[len(region.Bounds.Tiles)/2]
				terrain.SetTile(centerTile.X, centerTile.Y, TileFloor)
				regionConnections[i] = centerTile
			}
		}
	}

	// Connect each region to the next one
	for i := 0; i < len(regionConnections)-1; i++ {
		start := regionConnections[i]
		end := regionConnections[i+1]

		g.carveConnectingCorridor(terrain, start, end, rng)
	}

	// Also connect first and last for better connectivity
	if len(regionConnections) > 2 {
		start := regionConnections[0]
		end := regionConnections[len(regionConnections)-1]
		g.carveConnectingCorridor(terrain, start, end, rng)
	}

	return nil
}

// carveConnectingCorridor creates a walkable corridor between two points.
// Uses L-shaped corridors (horizontal then vertical, or vice versa).
func (g *CompositeGenerator) carveConnectingCorridor(terrain *Terrain, start, end Point, rng *rand.Rand) {
	horizontalFirst := rng.Intn(2) == 0

	if horizontalFirst {
		carveHorizontalThenVertical(terrain, start, end)
	} else {
		carveVerticalThenHorizontal(terrain, start, end)
	}
}

// carveHorizontalThenVertical carves a corridor horizontally first, then vertically.
func carveHorizontalThenVertical(terrain *Terrain, start, end Point) {
	carveHorizontalSegment(terrain, start.X, end.X, start.Y)
	carveVerticalSegment(terrain, start.Y, end.Y, end.X)
}

// carveVerticalThenHorizontal carves a corridor vertically first, then horizontally.
func carveVerticalThenHorizontal(terrain *Terrain, start, end Point) {
	carveVerticalSegment(terrain, start.Y, end.Y, start.X)
	carveHorizontalSegment(terrain, start.X, end.X, end.Y)
}

// carveHorizontalSegment carves a horizontal corridor segment.
// Carves 3 tiles wide (center ± 1) to accommodate 64×64 player sprites.
func carveHorizontalSegment(terrain *Terrain, startX, endX, y int) {
	if startX > endX {
		startX, endX = endX, startX
	}
	for x := startX; x <= endX; x++ {
		for dy := -corridorHalfWidth; dy <= corridorHalfWidth; dy++ {
			if terrain.IsInBounds(x, y+dy) {
				terrain.SetTile(x, y+dy, TileFloor)
			}
		}
	}
}

// carveVerticalSegment carves a vertical corridor segment.
// Carves 3 tiles wide (center ± 1) to accommodate 64×64 player sprites.
func carveVerticalSegment(terrain *Terrain, startY, endY, x int) {
	if startY > endY {
		startY, endY = endY, startY
	}
	for y := startY; y <= endY; y++ {
		for dx := -corridorHalfWidth; dx <= corridorHalfWidth; dx++ {
			if terrain.IsInBounds(x+dx, y) {
				terrain.SetTile(x+dx, y, TileFloor)
			}
		}
	}
}

// placeStairs places stairs in the terrain for multi-level dungeons.
func (g *CompositeGenerator) placeStairs(terrain *Terrain, diagram *VoronoiDiagram, regions []*BiomeRegionInfo, rng *rand.Rand) {
	// Place stairs down in first region
	if len(regions) > 0 {
		region := regions[0]
		for _, tile := range region.Bounds.Tiles {
			if terrain.IsWalkable(tile.X, tile.Y) {
				terrain.AddStairs(tile.X, tile.Y, false) // down
				break
			}
		}
	}

	// Place stairs up in last region
	if len(regions) > 1 {
		region := regions[len(regions)-1]
		for i := len(region.Bounds.Tiles) - 1; i >= 0; i-- {
			tile := region.Bounds.Tiles[i]
			if terrain.IsWalkable(tile.X, tile.Y) {
				terrain.AddStairs(tile.X, tile.Y, true) // up
				break
			}
		}
	}
}

// Validate checks if the generated composite terrain meets quality standards.
func (g *CompositeGenerator) Validate(result interface{}) error {
	terrain, ok := result.(*Terrain)
	if !ok {
		return fmt.Errorf("result is not a Terrain")
	}

	walkable := g.countWalkableTiles(terrain)
	if err := g.validateWalkablePercentage(terrain, walkable); err != nil {
		return err
	}

	start, found := g.findFirstWalkableTile(terrain)
	if !found {
		return fmt.Errorf("no walkable tiles found")
	}

	connected := floodFillConnectivity(terrain, start)
	return g.validateConnectivity(walkable, connected)
}

// countWalkableTiles counts the number of walkable tiles in terrain.
func (g *CompositeGenerator) countWalkableTiles(terrain *Terrain) int {
	walkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkable++
			}
		}
	}
	return walkable
}

// validateWalkablePercentage checks if terrain has sufficient walkable area.
func (g *CompositeGenerator) validateWalkablePercentage(terrain *Terrain, walkable int) error {
	totalTiles := terrain.Width * terrain.Height
	walkablePercent := float64(walkable) / float64(totalTiles)
	if walkablePercent < 0.25 {
		return fmt.Errorf("insufficient walkable area: %.1f%% (need at least 25%%)", walkablePercent*100)
	}
	return nil
}

// findFirstWalkableTile finds the first walkable tile in terrain for connectivity check.
func (g *CompositeGenerator) findFirstWalkableTile(terrain *Terrain) (Point, bool) {
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				return Point{X: x, Y: y}, true
			}
		}
	}
	return Point{}, false
}

// validateConnectivity checks if sufficient walkable tiles are connected.
func (g *CompositeGenerator) validateConnectivity(walkable, connected int) error {
	connectedPercent := float64(connected) / float64(walkable)
	if connectedPercent < 0.90 {
		return fmt.Errorf("disconnected regions: only %.1f%% of walkable tiles are connected", connectedPercent*100)
	}
	return nil
}

// floodFillConnectivity counts how many walkable tiles are reachable from start.
func floodFillConnectivity(terrain *Terrain, start Point) int {
	visited := make(map[Point]bool)
	queue := []Point{start}
	count := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}

		if !terrain.IsWalkable(current.X, current.Y) {
			continue
		}

		visited[current] = true
		count++

		// Add neighbors
		neighbors := current.Neighbors()
		for _, n := range neighbors {
			if terrain.IsInBounds(n.X, n.Y) && !visited[n] {
				queue = append(queue, n)
			}
		}
	}

	return count
}
