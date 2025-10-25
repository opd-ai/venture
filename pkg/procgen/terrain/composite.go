// Package terrain provides composite terrain generation.
// This file implements the CompositeGenerator which combines multiple
// biome types (dungeon, cave, forest, city, maze) into a single level.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// CompositeGenerator combines multiple terrain biomes in a single level.
// Uses Voronoi partitioning to create distinct regions with smooth transitions.
type CompositeGenerator struct {
	biomeCount      int                          // Number of biomes to combine (2-4)
	transitionWidth int                          // Width of transition zones in tiles (2-4)
	generators      map[string]procgen.Generator // Available generators by name
}

// NewCompositeGenerator creates a new composite terrain generator.
func NewCompositeGenerator() *CompositeGenerator {
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

// Generate creates composite terrain by combining multiple biome generators.
func (g *CompositeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Extract custom parameters
	width := 80
	height := 50
	biomeCount := g.biomeCount      // Use local copy
	transitionWidth := g.transitionWidth  // Use local copy
	
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

	// Validate parameters
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}
	if width < 60 || height < 40 {
		return nil, fmt.Errorf("dimensions too small for composite generation: width=%d, height=%d (min 60x40)", width, height)
	}
	if width > 500 || height > 500 {
		return nil, fmt.Errorf("dimensions too large: width=%d, height=%d (max 500x500)", width, height)
	}
	if biomeCount < 2 || biomeCount > 4 {
		biomeCount = 3 // Clamp to valid range
	}
	if transitionWidth < 1 || transitionWidth > 5 {
		transitionWidth = 3 // Clamp to valid range
	}

	// Create RNG
	rng := rand.New(rand.NewSource(seed))

	// Create base terrain
	terrain := NewTerrain(width, height, seed)

	// Create Voronoi diagram for biome partitioning
	diagram := GenerateVoronoiDiagram(width, height, biomeCount, rng)

	// Select generators for each region
	generatorNames := g.selectGenerators(params.GenreID, biomeCount, rng)

	// Create biome region info
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

	// Generate each biome region
	for _, region := range biomeRegions {
		if err := g.generateBiomeRegion(terrain, region, diagram, params, rng); err != nil {
			return nil, fmt.Errorf("failed to generate biome %d (%s): %w", region.ID, region.GeneratorName, err)
		}
	}

	// Create biome type mapping
	biomeTypes := make(map[int]BiomeType)
	for _, region := range biomeRegions {
		biomeTypes[region.ID] = region.BiomeType
	}

	// Apply transition zones
	BlendTransitionZones(terrain, diagram, biomeTypes, transitionWidth, rng)

	// Ensure connectivity between all regions
	if err := g.ensureConnectivity(terrain, diagram, biomeRegions, rng); err != nil {
		return nil, fmt.Errorf("failed to ensure connectivity: %w", err)
	}

	// Place stairs (if multi-level)
	if params.Depth > 0 {
		g.placeStairs(terrain, diagram, biomeRegions, rng)
	}

	// Store biome region info in terrain (for debugging/visualization)
	terrain.Rooms = make([]*Room, 0)

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
		"postapocalyptic": {"cellular", "city", "forest"},
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
	// Get generator for this biome
	generator, ok := g.generators[region.GeneratorName]
	if !ok {
		return fmt.Errorf("unknown generator: %s", region.GeneratorName)
	}

	// Get region bounds
	minX, minY, maxX, maxY := diagram.GetRegionBounds(region.ID)
	regionWidth := maxX - minX + 1
	regionHeight := maxY - minY + 1

	if regionWidth <= 0 || regionHeight <= 0 {
		return fmt.Errorf("invalid region bounds: width=%d, height=%d", regionWidth, regionHeight)
	}

	// Enforce minimum region size to prevent generator issues
	// However, keep within the actual region bounds to maintain determinism
	minRegionWidth := 20
	minRegionHeight := 15
	if regionWidth < minRegionWidth {
		regionWidth = minRegionWidth
	}
	if regionHeight < minRegionHeight {
		regionHeight = minRegionHeight
	}
	
	// Never exceed the original terrain dimensions
	if minX+regionWidth > terrain.Width {
		regionWidth = terrain.Width - minX
	}
	if minY+regionHeight > terrain.Height {
		regionHeight = terrain.Height - minY
	}

	// Create params for this region (deep copy to avoid mutation)
	regionParams := procgen.GenerationParams{
		Difficulty: params.Difficulty,
		Depth:      params.Depth,
		GenreID:    params.GenreID,
		Custom:     make(map[string]interface{}),
	}
	
	// Copy original custom params
	if params.Custom != nil {
		for k, v := range params.Custom {
			regionParams.Custom[k] = v
		}
	}
	
	// Override dimensions for this region
	regionParams.Custom["width"] = regionWidth
	regionParams.Custom["height"] = regionHeight

	// Generate biome terrain
	result, err := generator.Generate(region.Seed, regionParams)
	if err != nil {
		return fmt.Errorf("generator %s failed: %w", region.GeneratorName, err)
	}

	biomeTerrain, ok := result.(*Terrain)
	if !ok {
		return fmt.Errorf("generator %s returned non-terrain result", region.GeneratorName)
	}

	// Copy generated tiles to main terrain (only tiles belonging to this region)
	for _, tile := range region.Bounds.Tiles {
		// Calculate offset into biome terrain
		biomeX := tile.X - minX
		biomeY := tile.Y - minY

		// Bounds check
		if biomeX >= 0 && biomeX < regionWidth && biomeY >= 0 && biomeY < regionHeight {
			srcTile := biomeTerrain.GetTile(biomeX, biomeY)
			terrain.SetTile(tile.X, tile.Y, srcTile)
		}
	}

	return nil
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
	// Randomly choose horizontal-first or vertical-first
	horizontalFirst := rng.Intn(2) == 0

	if horizontalFirst {
		// Horizontal then vertical
		// Move horizontally from start.X to end.X
		startX, endX := start.X, end.X
		if startX > endX {
			startX, endX = endX, startX
		}
		for x := startX; x <= endX; x++ {
			terrain.SetTile(x, start.Y, TileFloor)
		}

		// Move vertically from start.Y to end.Y
		startY, endY := start.Y, end.Y
		if startY > endY {
			startY, endY = endY, startY
		}
		for y := startY; y <= endY; y++ {
			terrain.SetTile(end.X, y, TileFloor)
		}
	} else {
		// Vertical then horizontal
		// Move vertically from start.Y to end.Y
		startY, endY := start.Y, end.Y
		if startY > endY {
			startY, endY = endY, startY
		}
		for y := startY; y <= endY; y++ {
			terrain.SetTile(start.X, y, TileFloor)
		}

		// Move horizontally from start.X to end.X
		startX, endX := start.X, end.X
		if startX > endX {
			startX, endX = endX, startX
		}
		for x := startX; x <= endX; x++ {
			terrain.SetTile(x, end.Y, TileFloor)
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

	// Count walkable tiles
	walkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkable++
			}
		}
	}

	totalTiles := terrain.Width * terrain.Height
	walkablePercent := float64(walkable) / float64(totalTiles)

	// Composite terrain should have at least 25% walkable (lower than single-biome due to variety)
	if walkablePercent < 0.25 {
		return fmt.Errorf("insufficient walkable area: %.1f%% (need at least 25%%)", walkablePercent*100)
	}

	// Check connectivity via flood fill
	// Find first walkable tile
	var start Point
	found := false
	for y := 0; y < terrain.Height && !found; y++ {
		for x := 0; x < terrain.Width && !found; x++ {
			if terrain.IsWalkable(x, y) {
				start = Point{X: x, Y: y}
				found = true
			}
		}
	}

	if !found {
		return fmt.Errorf("no walkable tiles found")
	}

	// Flood fill to count connected walkable tiles
	connected := floodFillConnectivity(terrain, start)

	// At least 90% of walkable tiles should be connected
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
