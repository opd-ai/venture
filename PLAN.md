# Procedural Terrain Generation Expansion Plan

**Last Updated:** October 24, 2025  
**Target Completion:** Phase 8.5 (Polish & Optimization)

## Overview

Expand terrain generation from 2 algorithms (BSP dungeons, cellular caves) to 6+ algorithms with multi-level support, water features, and genre-specific variations.

**Goals:**
- Add 9+ new tile types (water, trees, stairs, trap doors, etc.)
- Implement 4 new generators (forest, city, maze, composite)
- Support multi-level dungeons with stair connectivity
- Maintain determinism, 80%+ test coverage, <2s generation for 200x200

---

## 1. Architecture Overview

### Current System
- **Generators:** `bsp.go` (dungeons), `cellular.go` (caves)
- **Tiles:** Wall, Floor, Door, Corridor (4 types)
- **Interface:** `procgen.Generator` with `Generate(seed, params)` and `Validate(result)`
- **Location:** `pkg/procgen/terrain/`

### Extensions Needed
```go
// types.go - Add to TileType enum
const (
    TileWall TileType = iota
    TileFloor
    TileDoor
    TileCorridor
    // NEW TILES:
    TileWaterShallow  // Walkable but slow
    TileWaterDeep     // Impassable
    TileTree          // Impassable natural obstacle
    TileStairsUp      // Level transition up
    TileStairsDown    // Level transition down
    TileTrapDoor      // Hidden/revealed trap door
    TileSecretDoor    // Hidden door
    TileBridge        // Walkable over water
    TileStructure     // Building/ruins
)

// Add multi-level support
type Terrain struct {
    Width, Height int
    Tiles         [][]TileType
    Rooms         []*Room
    Seed          int64
    Level         int           // NEW: Current dungeon level
    StairsUp      []Point       // NEW: Up stair positions
    StairsDown    []Point       // NEW: Down stair positions
}

type Point struct { X, Y int }
```

### Generator Selection Strategy
```go
// generator_selector.go (NEW FILE)
func SelectGenerator(genreID string, depth int, rng *rand.Rand) Generator {
    // Depth 1-3: BSP dungeons (structured)
    // Depth 4-6: Cellular caves (organic)
    // Depth 7-9: Maze (confusing)
    // Depth 10+: Composite (multi-biome)
    // Genre overrides: fantasy→forest, scifi→city, etc.
}
```

---

## 2. Implementation Phases

### Phase 1: Tile Types & Infrastructure (2-3 hours) ✅ COMPLETE
**Objective:** Add new tile types and update core functionality.

**Files to Modify:**
- `pkg/procgen/terrain/types.go:8-15` - Add TileType constants
- `pkg/procgen/terrain/types.go:17-35` - Update String() method
- `pkg/procgen/terrain/types.go:138-145` - Update IsWalkable() logic
- `pkg/procgen/terrain/types.go:110-120` - Add Level, StairsUp, StairsDown fields

**New Files:**
- `pkg/procgen/terrain/point.go` - Point struct and utilities

**Key Functions:**
```go
func (t TileType) IsWalkable() bool
func (t TileType) IsTransparent() bool  // For vision/rendering
func (t TileType) MovementCost() float64 // Shallow water = 2.0x
func (terr *Terrain) AddStairs(x, y int, up bool)
func (terr *Terrain) ValidateStairPlacement() error
```

**Tests:** `types_test.go` - tile properties, stair validation
**Validation:** All new tiles have correct properties, stairs placed in walkable areas

---

### Phase 2: Maze Generator (3-4 hours) ✅ COMPLETE
**Objective:** Implement recursive backtracking maze algorithm.

**New File:** `pkg/procgen/terrain/maze.go`

**Algorithm (Pseudo-code):**
```
1. Fill entire grid with walls
2. Pick random start position, mark as floor
3. Stack-based DFS:
   - Mark current cell as floor
   - Get unvisited neighbors (2 cells away)
   - If neighbors exist:
     - Choose random neighbor
     - Carve path (remove wall between)
     - Push neighbor to stack
   - Else: pop stack
4. Add rooms at dead ends (10% chance)
5. Place stairs in furthest corners
```

**Key Functions:**
```go
type MazeGenerator struct {
    roomChance    float64 // 0.1 = 10% of dead ends become rooms
    corridorWidth int     // 1 = single tile, 2 = double-wide
}

func (g *MazeGenerator) Generate(seed int64, params GenerationParams) (interface{}, error)
func (g *MazeGenerator) carvePassages(x, y int, terrain *Terrain, rng *rand.Rand)
func (g *MazeGenerator) findDeadEnds(terrain *Terrain) []Point
func (g *MazeGenerator) createRoomAtDeadEnd(x, y int, terrain *Terrain, rng *rand.Rand)
```

**Tests:** `maze_test.go`
- Determinism (same seed = same maze)
- Connectivity (all floors reachable)
- Dead end room creation
- Performance (<2s for 200x200)

---

### Phase 3: Forest Generator (4-5 hours) ✅ COMPLETE
**Objective:** Generate natural forest areas with clearings and paths.

**New File:** `pkg/procgen/terrain/forest.go`

**Algorithm:**
```
1. Fill with TileFloor (grassland)
2. Poisson disc sampling for tree placement
   - Min distance between trees: 3-5 tiles
   - Density based on params.Difficulty
3. Create clearings (rooms):
   - Random circular/elliptical clearings
   - Connect with organic paths (A* pathfinding)
4. Add water features:
   - Perlin noise for natural boundaries
   - Rivers: trace noise gradient
   - Lakes: flood-fill low-noise areas
   - Bridges: auto-place on paths crossing water
5. Place stairs in largest clearing
```

**Key Functions:**
```go
type ForestGenerator struct {
    treeDensity   float64 // 0.3 = 30% of tiles
    clearingCount int     // Number of open areas
}

func (g *ForestGenerator) generateTrees(terrain *Terrain, rng *rand.Rand)
func (g *ForestGenerator) poissonDiscSampling(width, height, minDist int, rng *rand.Rand) []Point
func (g *ForestGenerator) createClearings(terrain *Terrain, rng *rand.Rand) []*Room
func (g *ForestGenerator) createOrganicPath(start, end Point, terrain *Terrain, rng *rand.Rand)
func (g *ForestGenerator) addWaterFeatures(terrain *Terrain, rng *rand.Rand)
func (g *ForestGenerator) placeAutoBridges(terrain *Terrain)
```

**Tests:** `forest_test.go`
- Tree distribution (Poisson disc spacing)
- Clearing connectivity
- Water feature placement
- Bridge auto-placement over water

---

### Phase 4: City Generator (5-6 hours) ✅ COMPLETE
**Objective:** Generate urban environments with buildings and streets.

**New File:** `pkg/procgen/terrain/city.go`

**Algorithm:**
```
1. Grid subdivision:
   - Divide map into city blocks (8x8 to 16x16)
   - Leave 2-3 tile streets between blocks
2. Building placement:
   - 70% of blocks: place building (TileStructure)
   - 20% of blocks: open plaza (TileFloor)
   - 10% of blocks: park/water feature
3. Building interiors:
   - Small buildings: single room
   - Large buildings: BSP subdivide interior
4. Street network:
   - Grid pattern for main streets
   - Alleys between buildings (30% chance)
5. Place stairs in central plaza or large building
```

**Key Functions:**
```go
type CityGenerator struct {
    blockSize     int     // 8-16 tiles per block
    streetWidth   int     // 2-3 tiles
    buildingDensity float64 // 0.7 = 70% of blocks have buildings
}

func (g *CityGenerator) subdivideGrid(terrain *Terrain) []Rect
func (g *CityGenerator) placeBuildings(blocks []Rect, terrain *Terrain, rng *rand.Rand)
func (g *CityGenerator) createBuildingInterior(block Rect, terrain *Terrain, rng *rand.Rand)
func (g *CityGenerator) createStreetNetwork(blocks []Rect, terrain *Terrain)
```

**Tests:** `city_test.go`
- Grid subdivision correctness
- Street connectivity (all buildings accessible)
- Building placement density
- Interior room generation

---

### Phase 5: Water System (3-4 hours)
**Objective:** Add water generation utilities for all generators.

**New File:** `pkg/procgen/terrain/water.go`

**Utilities:**
```go
type WaterFeature struct {
    Type      WaterType  // Lake, River, Moat
    Tiles     []Point    // Coordinates of water tiles
    Bridges   []Point    // Auto-placed bridge locations
}

type WaterType int
const (
    WaterLake WaterType = iota
    WaterRiver
    WaterMoat
)

func GenerateLake(center Point, radius int, terrain *Terrain, rng *rand.Rand) *WaterFeature
func GenerateRiver(start, end Point, width int, terrain *Terrain, rng *rand.Rand) *WaterFeature
func GenerateMoat(room *Room, width int, terrain *Terrain) *WaterFeature
func PlaceBridges(feature *WaterFeature, terrain *Terrain, rng *rand.Rand)
func FloodFill(start Point, maxTiles int, terrain *Terrain) []Point
```

**Integration:**
- BSP: Moats around boss rooms
- Cellular: Underground lakes
- Forest: Lakes and rivers
- Maze: Water-filled dead ends

**Tests:** `water_test.go`
- Lake generation (circular shape)
- River generation (follows path)
- Bridge placement (over water crossings)
- Flood fill correctness

---

### Phase 6: Multi-Level Support (4-5 hours)
**Objective:** Generate connected multi-level dungeons.

**New File:** `pkg/procgen/terrain/multilevel.go`

**Functions:**
```go
type LevelGenerator struct {
    generators map[int]Generator // Depth -> Generator mapping
}

func NewLevelGenerator() *LevelGenerator
func (g *LevelGenerator) GenerateLevel(level int, seed int64, params GenerationParams) (*Terrain, error)
func (g *LevelGenerator) ConnectLevels(above, below *Terrain) error
func (g *LevelGenerator) ValidateLevelConnection(above, below *Terrain) error

// Stair placement strategies
func PlaceStairsRandom(terrain *Terrain, rng *rand.Rand)
func PlaceStairsInRoom(terrain *Terrain, roomType RoomType, rng *rand.Rand)
func PlaceStairsSymmetric(terrain *Terrain, rng *rand.Rand) // Corners/edges
```

**Validation Rules:**
- Every level (except first) has stairs up
- Every level (except last) has stairs down
- Stairs placed in walkable, accessible areas
- Stairs in level N+1 roughly align with stairs in level N

**Tests:** `multilevel_test.go`
- Multi-level generation (3-5 levels)
- Stair alignment validation
- Level connectivity (all levels reachable)
- Determinism across levels

---

### Phase 7: Composite Generator (5-6 hours)
**Objective:** Combine multiple biomes in a single level.

**New File:** `pkg/procgen/terrain/composite.go`

**Algorithm:**
```
1. Voronoi partitioning:
   - Place 2-4 biome seeds randomly
   - Assign each tile to nearest seed
2. Generate each region independently:
   - Dungeon region: Use BSP generator
   - Cave region: Use cellular generator
   - Forest region: Use forest generator
   - City region: Use city generator
3. Create transition zones:
   - 3-5 tile border between biomes
   - Blend tile types (forest→cave = rocky terrain)
4. Connect regions:
   - Ensure at least one path between all regions
   - Use appropriate transition tiles
5. Place stairs in central region or junction
```

**Key Functions:**
```go
type CompositeGenerator struct {
    biomeCount int // 2-4 biomes per map
}

type BiomeRegion struct {
    Generator Generator
    Seed      Point
    Tiles     []Point
}

func (g *CompositeGenerator) createVoronoiRegions(count int, terrain *Terrain, rng *rand.Rand) []BiomeRegion
func (g *CompositeGenerator) selectBiomeGenerator(genreID string, rng *rand.Rand) Generator
func (g *CompositeGenerator) generateBiomeRegion(region BiomeRegion, terrain *Terrain, rng *rand.Rand)
func (g *CompositeGenerator) createTransitionZones(regions []BiomeRegion, terrain *Terrain)
func (g *CompositeGenerator) connectRegions(regions []BiomeRegion, terrain *Terrain, rng *rand.Rand)
```

**Tests:** `composite_test.go`
- Voronoi partitioning (regions non-overlapping)
- Region connectivity (all reachable)
- Transition zone creation
- Multi-biome determinism

---

### Phase 8: Genre Integration (2-3 hours)
**Objective:** Map genres to terrain features and generators.

**File to Modify:** `pkg/procgen/terrain/genre_mapping.go` (NEW)

**Mappings:**
```go
var GenreTerrainPreferences = map[string]TerrainPreference{
    "fantasy": {
        Generators: []string{"bsp", "cellular", "forest"},
        TileThemes: map[TileType]string{
            TileWall:      "stone",
            TileFloor:     "cobblestone",
            TileTree:      "ancient_oak",
            TileStructure: "castle_ruins",
        },
        WaterChance: 0.3,  // 30% of maps have water
        TreeType:    "oak/pine",
    },
    "scifi": {
        Generators: []string{"city", "maze"},
        TileThemes: map[TileType]string{
            TileWall:      "metal_panel",
            TileFloor:     "deck_plating",
            TileStructure: "tech_building",
        },
        WaterChance: 0.0,  // No natural water
        TreeType:    "",   // No trees
    },
    "horror": {
        Generators: []string{"cellular", "maze"},
        WaterChance: 0.5,  // Lots of water (murky/bloody)
        TreeType:    "dead_tree/withered",
    },
    // ... postapoc, cyberpunk
}

func GetGeneratorForGenre(genreID string, depth int, rng *rand.Rand) Generator
func GetTileTheme(genreID string, tile TileType) string
```

**Tests:** `genre_mapping_test.go`
- Correct generator selection per genre
- Theme application
- Water/tree placement based on genre

---

### Phase 9: CLI Test Tool Enhancement (1-2 hours)
**Objective:** Update `terraintest` to support all new generators and features.

**File to Modify:** `cmd/terraintest/main.go`

**Add Flags:**
```go
-algorithm string   // "bsp", "cellular", "maze", "forest", "city", "composite"
-genre string       // "fantasy", "scifi", "horror", "cyberpunk", "postapoc"
-levels int         // Generate multi-level dungeon
-water bool         // Include water features
-visualize string   // "ascii", "color", "stats"
```

**Update Rendering:**
```go
func renderTerrain(terr *Terrain) string {
    // Use appropriate symbols for new tiles
    // W = shallow water, ~ = deep water
    // T = tree, ^ = stairs up, v = stairs down
    // = = bridge, # = structure
}

func renderTerrainColor(terr *Terrain) string {
    // ANSI color codes for different tiles
}

func renderStats(terr *Terrain) string {
    // Detailed statistics:
    // - Tile type distribution
    // - Room count and types
    // - Connectivity metrics
    // - Water coverage
    // - Stair locations
}
```

---

## 3. Technical Specifications

### Determinism Approach
```go
// Always use seeded RNG, never time.Now() or global rand
rng := rand.New(rand.NewSource(seed))

// For sub-generators in composite:
seedGen := procgen.NewSeedGenerator(baseSeed)
biomeSeed := seedGen.GetSeed("biome", biomeIndex)
```

### Performance Budget
| Generator | 100x100 | 200x200 | 500x500 |
|-----------|---------|---------|---------|
| BSP       | <50ms   | <200ms  | <1.5s   |
| Cellular  | <100ms  | <400ms  | <2.5s   |
| Maze      | <80ms   | <300ms  | <2.0s   |
| Forest    | <150ms  | <600ms  | <3.0s   |
| City      | <120ms  | <500ms  | <2.5s   |
| Composite | <300ms  | <1.2s   | <5.0s   |

### Walkability Validation
```go
func (g *Generator) Validate(result interface{}) error {
    terrain := result.(*Terrain)
    
    // 1. Minimum walkable percentage (30%)
    walkable := countWalkableTiles(terrain)
    if float64(walkable) / float64(terrain.Width * terrain.Height) < 0.3 {
        return fmt.Errorf("insufficient walkable area: %d tiles", walkable)
    }
    
    // 2. Connectivity check (flood fill from spawn)
    reachable := floodFillFromSpawn(terrain)
    if reachable < walkable * 0.95 {  // 95% of walkable tiles must be connected
        return fmt.Errorf("disconnected regions detected")
    }
    
    // 3. Stair validation
    if err := terrain.ValidateStairPlacement(); err != nil {
        return err
    }
    
    return nil
}
```

---

## 4. Testing Strategy

### Test Organization
```
terrain_test.go         # Core Terrain type tests (existing)
bsp_test.go            # BSP generator tests (existing)
cellular_test.go       # Cellular generator tests (existing)
maze_test.go           # NEW: Maze generator tests
forest_test.go         # NEW: Forest generator tests
city_test.go           # NEW: City generator tests
water_test.go          # NEW: Water system tests
multilevel_test.go     # NEW: Multi-level tests
composite_test.go      # NEW: Composite generator tests
genre_mapping_test.go  # NEW: Genre integration tests
validation_test.go     # NEW: Cross-generator validation
```

### Test Patterns
```go
// 1. Determinism test (apply to ALL generators)
func TestGeneratorDeterminism(t *testing.T) {
    gen := NewXXXGenerator()
    seed := int64(12345)
    params := procgen.GenerationParams{...}
    
    result1, _ := gen.Generate(seed, params)
    result2, _ := gen.Generate(seed, params)
    
    if !terrainsEqual(result1.(*Terrain), result2.(*Terrain)) {
        t.Error("Generation is not deterministic")
    }
}

// 2. Validation test
func TestGeneratorValidation(t *testing.T) {
    gen := NewXXXGenerator()
    result, _ := gen.Generate(12345, params)
    
    if err := gen.Validate(result); err != nil {
        t.Errorf("Validation failed: %v", err)
    }
}

// 3. Connectivity test
func TestGeneratorConnectivity(t *testing.T) {
    // Ensure all walkable tiles are reachable via flood fill
}

// 4. Performance benchmark
func BenchmarkGeneratorXXX(b *testing.B) {
    gen := NewXXXGenerator()
    params := procgen.GenerationParams{...}
    
    for i := 0; i < b.N; i++ {
        gen.Generate(int64(i), params)
    }
}
```

### Coverage Targets
- Types/utilities: 100%
- Each generator: 85%+
- Integration functions: 80%+
- Overall package: 90%+

---

## 5. Integration Points

### Rendering System
**File:** `pkg/engine/terrain_render_system.go`

```go
// Update tile rendering to handle new types
func (s *TerrainRenderSystem) getTileColor(tile TileType) color.RGBA {
    switch tile {
    case TileWaterShallow:
        return color.RGBA{100, 150, 255, 200}  // Light blue, transparent
    case TileWaterDeep:
        return color.RGBA{30, 60, 150, 255}     // Dark blue
    case TileTree:
        return color.RGBA{34, 139, 34, 255}     // Forest green
    // ... add cases for all new tiles
    }
}
```

### Entity Spawning
**File:** `pkg/world/world_generator.go`

```go
// Spawn entities based on tile types
func spawnEntitiesForTerrain(terrain *Terrain, world *engine.World) {
    for y := 0; y < terrain.Height; y++ {
        for x := 0; x < terrain.Width; x++ {
            switch terrain.GetTile(x, y) {
            case TileTree:
                // 5% chance to spawn tree enemy
            case TileWaterDeep:
                // 10% chance to spawn water enemy
            case TileStructure:
                // Spawn inside buildings
            }
        }
    }
}
```

### Movement/Collision System
**File:** `pkg/engine/movement_system.go`

```go
// Update movement costs
func (s *MovementSystem) getMovementCost(tile TileType) float64 {
    switch tile {
    case TileFloor, TileCorridor:
        return 1.0
    case TileWaterShallow:
        return 2.0  // Half speed
    case TileWaterDeep, TileWall, TileTree:
        return math.Inf(1)  // Impassable
    }
}
```

### Network Synchronization
**File:** `pkg/network/terrain_sync.go`

```go
// Only sync terrain once per level (it's deterministic)
type TerrainSyncMessage struct {
    Level int
    Seed  int64
    // Don't send full terrain, just regenerate on client
}
```

---

## 6. Implementation Timeline

| Phase | Estimated Time | Dependencies |
|-------|----------------|--------------|
| 1. Tile Types | 2-3 hours | None |
| 2. Maze Generator | 3-4 hours | Phase 1 |
| 3. Forest Generator | 4-5 hours | Phase 1 |
| 4. City Generator | 5-6 hours | Phase 1 |
| 5. Water System | 3-4 hours | Phase 1 |
| 6. Multi-Level | 4-5 hours | Phases 1-5 |
| 7. Composite | 5-6 hours | Phases 1-5 |
| 8. Genre Integration | 2-3 hours | Phases 1-7 |
| 9. CLI Tool | 1-2 hours | Phases 1-8 |
| **Total** | **29-38 hours** | Sequential |

**Suggested Schedule:**
- Week 1: Phases 1-3 (infrastructure + simple generators)
- Week 2: Phases 4-5 (complex generators + water)
- Week 3: Phases 6-7 (multi-level + composite)
- Week 4: Phases 8-9 (polish + testing)

---

## 7. Success Criteria

- ✅ All 9+ new tile types implemented and tested
- ✅ 4 new generators functional (maze, forest, city, composite)
- ✅ Multi-level dungeons with validated stair connectivity
- ✅ Water features integrated into all applicable generators
- ✅ Genre-specific terrain variations working
- ✅ 80%+ test coverage for all new code
- ✅ All generators pass determinism tests
- ✅ Performance targets met (<2s for 200x200)
- ✅ CLI tool supports all new features
- ✅ Integration with rendering/movement/collision systems complete
- ✅ Documentation updated (doc.go, README.md)

---

## 8. Future Enhancements (Post-Plan)

- **Dynamic terrain modification:** Dig, build walls, place structures
- **Environmental hazards:** Lava, poison gas, collapsing floors
- **Seasonal variations:** Snow, rain effects on terrain
- **Destructible terrain:** Breakable walls, crumbling bridges
- **Puzzle elements:** Pressure plates, moving platforms, teleporters
- **Biome transitions:** Gradual blending instead of hard borders
- **Procedural textures:** Per-tile texture variation
- **Lighting integration:** Light sources, shadows, darkness
- **Weather system:** Rain creates puddles, snow covers ground

---

## Notes

- **Maintain backward compatibility:** Existing BSP/Cellular generators unchanged
- **Zero external dependencies:** Use only Go stdlib + Ebiten
- **Document algorithms:** Add comments explaining each step
- **Profile before optimizing:** Use `go test -bench` to identify bottlenecks
- **Test edge cases:** Empty maps, single-tile maps, huge maps (1000x1000)
- **Genre-aware defaults:** Each generator has sensible defaults per genre
- **Validate early:** Check parameters before expensive generation
- **Fail gracefully:** Return errors, don't panic (except programmer errors)

