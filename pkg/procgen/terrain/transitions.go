// Package terrain provides biome transition utilities for composite terrain generation.
package terrain

import (
	"math/rand"
)

// BiomeType represents different biome styles for transition blending.
type BiomeType int

const (
	BiomeDungeon BiomeType = iota
	BiomeCave
	BiomeForest
	BiomeCity
	BiomeMaze
)

// TransitionStyle defines how tiles blend between two biomes.
type TransitionStyle struct {
	From        BiomeType
	To          BiomeType
	TileChoices []TileType // Possible tiles for transition zone
	Weights     []float64  // Probability weights for each tile
}

// GetBiomeType returns the biome type for a given generator name.
func GetBiomeType(generatorName string) BiomeType {
	switch generatorName {
	case "bsp":
		return BiomeDungeon
	case "cellular":
		return BiomeCave
	case "forest":
		return BiomeForest
	case "city":
		return BiomeCity
	case "maze":
		return BiomeMaze
	default:
		return BiomeDungeon
	}
}

// GetTransitionStyle returns the appropriate transition style between two biomes.
func GetTransitionStyle(from, to BiomeType) TransitionStyle {
	// Normalize: always process in consistent order (smaller ID first)
	if from > to {
		from, to = to, from
	}

	// Define transitions between biome pairs
	switch {
	case from == BiomeDungeon && to == BiomeCave:
		// Dungeon to Cave: broken walls, rubble
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileWall, TileCorridor},
			Weights:     []float64{0.5, 0.3, 0.2},
		}

	case from == BiomeDungeon && to == BiomeForest:
		// Dungeon to Forest: overgrown ruins
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileTree, TileStructure},
			Weights:     []float64{0.5, 0.3, 0.2},
		}

	case from == BiomeDungeon && to == BiomeCity:
		// Dungeon to City: urban ruins
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileStructure, TileWall},
			Weights:     []float64{0.5, 0.3, 0.2},
		}

	case from == BiomeDungeon && to == BiomeMaze:
		// Dungeon to Maze: crumbling passages
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileCorridor, TileFloor, TileWall},
			Weights:     []float64{0.4, 0.4, 0.2},
		}

	case from == BiomeCave && to == BiomeForest:
		// Cave to Forest: rocky clearings
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileTree, TileWall},
			Weights:     []float64{0.5, 0.3, 0.2},
		}

	case from == BiomeCave && to == BiomeCity:
		// Cave to City: underground city
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileStructure, TileWall},
			Weights:     []float64{0.5, 0.3, 0.2},
		}

	case from == BiomeCave && to == BiomeMaze:
		// Cave to Maze: narrow cavern passages
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileCorridor, TileFloor, TileWall},
			Weights:     []float64{0.4, 0.3, 0.3},
		}

	case from == BiomeForest && to == BiomeCity:
		// Forest to City: urban park
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileTree, TileStructure},
			Weights:     []float64{0.5, 0.25, 0.25},
		}

	case from == BiomeForest && to == BiomeMaze:
		// Forest to Maze: hedge maze
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileTree, TileCorridor},
			Weights:     []float64{0.4, 0.3, 0.3},
		}

	case from == BiomeCity && to == BiomeMaze:
		// City to Maze: alley network
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileCorridor, TileFloor, TileStructure},
			Weights:     []float64{0.4, 0.3, 0.3},
		}

	default:
		// Default: mix of floor and walls
		return TransitionStyle{
			From:        from,
			To:          to,
			TileChoices: []TileType{TileFloor, TileWall},
			Weights:     []float64{0.7, 0.3},
		}
	}
}

// ApplyTransitionZone applies transition tiles to the specified zone.
// Uses weighted random selection based on the transition style.
func ApplyTransitionZone(terrain *Terrain, zone []Point, style TransitionStyle, rng *rand.Rand) {
	if len(style.TileChoices) == 0 {
		return
	}

	// Apply transition tiles
	for _, p := range zone {
		// Skip if already stairs (preserve connectivity)
		currentTile := terrain.GetTile(p.X, p.Y)
		if currentTile == TileStairsUp || currentTile == TileStairsDown {
			continue
		}

		// Select tile based on weighted probabilities
		tile := weightedSelectTile(style.TileChoices, style.Weights, rng)
		terrain.SetTile(p.X, p.Y, tile)
	}
}

// weightedSelectTile selects a tile type based on weighted probabilities.
func weightedSelectTile(tiles []TileType, weights []float64, rng *rand.Rand) TileType {
	if len(tiles) == 0 {
		return TileFloor
	}

	if len(weights) != len(tiles) {
		// No weights, uniform selection
		return tiles[rng.Intn(len(tiles))]
	}

	// Calculate cumulative weights
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	// Select based on weighted probability
	r := rng.Float64() * totalWeight
	cumulative := 0.0

	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return tiles[i]
		}
	}

	// Fallback: return last tile
	return tiles[len(tiles)-1]
}

// BlendTransitionZones creates transition zones between all adjacent regions.
// Radius determines the width of the transition zone (typically 2-4 tiles).
func BlendTransitionZones(terrain *Terrain, diagram *VoronoiDiagram, biomeTypes map[int]BiomeType, radius int, rng *rand.Rand) {
	// Find all boundary tiles
	boundaries := diagram.FindBoundaryTiles()

	// Create transition zone by expanding boundaries
	transitionZone := ExpandBoundaryZone(boundaries, radius, terrain.Width, terrain.Height)

	// Group transition tiles by adjacent region pairs
	regionPairs := make(map[[2]int][]Point)

	for _, p := range transitionZone {
		currentRegion := diagram.GetRegion(p.X, p.Y)
		if currentRegion < 0 {
			continue
		}

		// Check neighbors for different regions
		neighbors := []Point{
			{X: p.X - 1, Y: p.Y},
			{X: p.X + 1, Y: p.Y},
			{X: p.X, Y: p.Y - 1},
			{X: p.X, Y: p.Y + 1},
		}

		for _, n := range neighbors {
			neighborRegion := diagram.GetRegion(n.X, n.Y)
			if neighborRegion >= 0 && neighborRegion != currentRegion {
				// Create ordered pair (smaller ID first)
				pair := [2]int{currentRegion, neighborRegion}
				if pair[0] > pair[1] {
					pair[0], pair[1] = pair[1], pair[0]
				}

				regionPairs[pair] = append(regionPairs[pair], p)
				break
			}
		}
	}

	// Apply transition styles for each region pair
	// Sort pairs for deterministic iteration
	sortedPairs := make([][2]int, 0, len(regionPairs))
	for pair := range regionPairs {
		sortedPairs = append(sortedPairs, pair)
	}
	
	// Sort by region IDs for consistent ordering
	for i := 0; i < len(sortedPairs); i++ {
		for j := i + 1; j < len(sortedPairs); j++ {
			pi, pj := sortedPairs[i], sortedPairs[j]
			if pi[0] > pj[0] || (pi[0] == pj[0] && pi[1] > pj[1]) {
				sortedPairs[i], sortedPairs[j] = sortedPairs[j], sortedPairs[i]
			}
		}
	}
	
	// Apply transitions in deterministic order
	for _, pair := range sortedPairs {
		tiles := regionPairs[pair]
		region1, region2 := pair[0], pair[1]

		biome1 := biomeTypes[region1]
		biome2 := biomeTypes[region2]

		style := GetTransitionStyle(biome1, biome2)
		ApplyTransitionZone(terrain, tiles, style, rng)
	}
}

// EnsureTransitionWalkability ensures transition zones don't create isolated regions.
// Carves walkable paths through transition zones if needed.
func EnsureTransitionWalkability(terrain *Terrain, transitionZone []Point, rng *rand.Rand) {
	// Every N tiles in transition zone, ensure at least one walkable path
	for i := 0; i < len(transitionZone); i += 5 {
		p := transitionZone[i]

		// Check if this point has a walkable neighbor
		hasWalkable := false
		neighbors := p.Neighbors()

		for _, n := range neighbors {
			if terrain.IsInBounds(n.X, n.Y) && terrain.IsWalkable(n.X, n.Y) {
				hasWalkable = true
				break
			}
		}

		// If isolated, make it walkable
		if !hasWalkable {
			terrain.SetTile(p.X, p.Y, TileFloor)
		}
	}
}
