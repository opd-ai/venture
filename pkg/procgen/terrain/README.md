# Terrain Generation

The terrain package provides procedural terrain and dungeon generation algorithms for the Venture game. All generators are deterministic based on seed values and produce varied, interesting layouts.

## Algorithms

### Binary Space Partitioning (BSP)

The BSP algorithm creates structured dungeon layouts by recursively dividing space into smaller regions and placing rooms within them. This produces organized, room-based dungeons similar to classic roguelikes.

**Features:**
- Rectangular rooms of varying sizes
- Corridor connections between rooms
- Good for structured dungeon layouts
- Predictable, game-friendly layouts

**Usage:**
```go
gen := terrain.NewBSPGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      1,
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "width":  80,
        "height": 50,
    },
}
result, err := gen.Generate(12345, params)
terrain := result.(*terrain.Terrain)
```

### Cellular Automata

The cellular automata algorithm creates organic, cave-like structures by starting with random noise and applying iterative rules. This produces natural-looking caverns and caves.

**Features:**
- Organic, natural-looking caves
- Varied open spaces
- Connected regions (post-processing ensures walkability)
- Good for cave systems and natural environments

**Usage:**
```go
gen := terrain.NewCellularGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      1,
    GenreID:    "caves",
    Custom: map[string]interface{}{
        "width":           80,
        "height":          50,
        "fillProbability": 0.40,
        "iterations":      5,
    },
}
result, err := gen.Generate(12345, params)
terrain := result.(*terrain.Terrain)
```

### Maze (Recursive Backtracking)

The maze algorithm creates complex, winding corridors using recursive backtracking. It produces confusing, labyrinthine layouts with optional rooms at dead ends.

**Features:**
- Complex, winding corridors
- Guaranteed connectivity (all areas reachable)
- Optional rooms at dead ends
- Configurable corridor width
- Stairs placed in opposite corners

**Usage:**
```go
gen := terrain.NewMazeGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      7, // Mazes are typically deeper levels
    GenreID:    "dungeon",
    Custom: map[string]interface{}{
        "width":         81,
        "height":        81,
        "roomChance":    0.1,  // 10% of dead ends become rooms
        "corridorWidth": 1,    // 1 = single tile, 2 = double-wide
    },
}
result, err := gen.Generate(12345, params)
terrain := result.(*terrain.Terrain)
```

**Parameters:**
- `roomChance` (float64): Probability (0.0-1.0) of creating a room at a dead end (default: 0.1)
- `corridorWidth` (int): Width of corridors, 1 for single-tile or 2 for double-wide (default: 1)

**Note:** The algorithm automatically adjusts even dimensions to odd values (required for the algorithm to work correctly).

### Forest (Poisson Disc Sampling)

The forest algorithm creates natural outdoor environments with trees, clearings, and water features using Poisson disc sampling for realistic tree distribution.

**Features:**
- Natural tree placement using Poisson disc sampling
- Elliptical clearings with organic shapes
- Lakes and rivers with shallow/deep water
- Automatic bridge placement over water
- Organic paths connecting clearings
- Stairs placed in clearings

**Usage:**
```go
gen := terrain.NewForestGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      1,
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "width":         80,
        "height":        50,
        "treeDensity":   0.3,  // 30% density for Poisson sampling
        "clearingCount": 3,    // Number of clearings to create
        "waterChance":   0.3,  // 30% probability of water features
    },
}
result, err := gen.Generate(12345, params)
terrain := result.(*terrain.Terrain)
```

**Parameters:**
- `treeDensity` (float64): Density parameter for Poisson disc sampling, controls tree spacing (default: 0.3)
- `clearingCount` (int): Number of open clearings to create (default: 3)
- `waterChance` (float64): Probability (0.0-1.0) of generating lakes or rivers (default: 0.3)

**Technical Details:**
- Uses Poisson disc sampling with grid-based spatial hashing for O(n) performance
- Minimum tree distance calculated from density: `minDist = 3.0 / sqrt(density)`
- Clearings are elliptical with configurable size (8-17 tiles in each dimension)
- Water features can be lakes (elliptical with noise) or rivers (winding paths)
- Bridges automatically placed where paths cross water

**Performance:**
- 40x30 forest: ~0.27ms
- 80x50 forest: ~0.83ms
- 150x100 forest: ~3.5ms

### City (Grid Subdivision)

The city algorithm creates urban environments with buildings, streets, and public spaces using grid-based block subdivision.

**Features:**
- Grid-based city blocks with configurable size
- Street network with grid pattern
- Three types of blocks: buildings, plazas, parks
- Building interiors with BSP subdivision
- Solid structures and accessible buildings with rooms
- Parks with trees and optional ponds

**Usage:**
```go
gen := terrain.NewCityGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      1,
    GenreID:    "scifi",
    Custom: map[string]interface{}{
        "width":           80,
        "height":          50,
        "blockSize":       12,   // Size of city blocks
        "streetWidth":     2,    // Width of streets
        "buildingDensity": 0.7,  // 70% of blocks have buildings
        "plazaDensity":    0.2,  // 20% of blocks are plazas
    },
}
result, err := gen.Generate(12345, params)
terrain := result.(*terrain.Terrain)
```

**Parameters:**
- `blockSize` (int): Size of city blocks in tiles, 4-30 (default: 12)
- `streetWidth` (int): Width of streets in tiles, 1-5 (default: 2)
- `buildingDensity` (float64): Percentage of blocks with buildings, 0.0-1.0 (default: 0.7)
- `plazaDensity` (float64): Percentage of blocks that are plazas, 0.0-1.0 (default: 0.2)

**Technical Details:**
- Grid subdivision creates regular city blocks separated by streets
- Buildings: 70% solid structures, 30% accessible with single room
- Large buildings (10x10+): BSP subdivided interiors with multiple rooms
- Plazas: Open public squares (tracked as rooms for stairs placement)
- Parks: Green spaces with trees (30% coverage) and optional ponds (20% chance)
- Streets provide full connectivity between all blocks

**Performance:**
- 40x30 city: ~0.02ms
- 80x50 city: ~0.06ms
- 200x200 city: ~2.9ms

## Water System

The water system provides utilities for generating lakes, rivers, moats, and other water features. All water generation is deterministic and integrates seamlessly with existing terrain generators.

**Features:**
- Natural-looking lakes with elliptical shapes
- Winding rivers between points
- Defensive moats around rooms
- Automatic bridge placement over water
- Flood fill utilities for water body creation
- Deep and shallow water zones

**Usage:**
```go
import "github.com/opd-ai/venture/pkg/procgen/terrain"

// Generate a lake
terrain := NewTerrain(80, 50)
rng := rand.New(rand.NewSource(12345))
lake := GenerateLake(terrain, 40, 25, 10, rng)
// Lake has shallow water at edges, deep water at center

// Generate a river
river := GenerateRiver(terrain, Point{X: 5, Y: 25}, Point{X: 75, Y: 25}, 3, rng)
// River meanders naturally between start and end points

// Generate a moat around a room
room := &Room{X: 30, Y: 20, Width: 10, Height: 8}
moat := GenerateMoat(terrain, room, 2, rng)
// Moat surrounds room at specified width

// Place bridges over water
path := []Point{{5, 20}, {10, 20}, {15, 20}} // Path crosses water
PlaceBridges(terrain, lake, path)
// Bridges automatically placed where path crosses water

// Flood fill for custom water bodies
start := Point{X: 40, Y: 25}
tiles := FloodFill(terrain, start, 100, rng)
FloodFillWater(terrain, tiles, 0.5, rng) // 50% deep, 50% shallow
```

### Water Types

The water system supports three main types of water features:

```go
type WaterType int

const (
    WaterTypeLake  WaterType = iota  // Natural lakes with elliptical shape
    WaterTypeRiver                    // Winding rivers between points
    WaterTypeMoat                     // Defensive moats around rooms
)
```

### Water Feature Structure

Each water feature tracks its tiles and any auto-placed bridges:

```go
type WaterFeature struct {
    Type    WaterType  // Type of water feature
    Tiles   []Point    // All water tiles in feature
    Bridges []Point    // Auto-placed bridge locations
}
```

### Water Generation Functions

**GenerateLake(terrain \*Terrain, centerX, centerY, radius int, rng \*rand.Rand) WaterFeature**
- Creates elliptical lake with radius variance (10-30% shape variation)
- Deep water (TileWaterDeep) at center (normalized distance ≤ 0.6)
- Shallow water (TileWaterShallow) at edges (normalized distance ≤ 1.0)
- Returns WaterFeature with all tiles

**GenerateRiver(terrain \*Terrain, start, end Point, width int, rng \*rand.Rand) WaterFeature**
- Creates winding river path between start and end points
- 20% perpendicular offset for natural meandering
- Width 1-5 tiles supported
- Deep water in center of wide rivers (width ≥ 3), shallow at edges
- Returns WaterFeature with all tiles

**GenerateMoat(terrain \*Terrain, room \*Room, width int, rng \*rand.Rand) WaterFeature**
- Surrounds room perimeter at specified width (1-3 tiles)
- Distance-based depth: deep water if dist ≤ width/2, else shallow
- Respects room boundaries (never overwrites room interior)
- Returns WaterFeature with all tiles

**PlaceBridges(terrain \*Terrain, waterFeature WaterFeature, path []Point)**
- Automatically places bridges where paths cross water
- Checks for water on both sides of path to detect crossings
- Converts TileCorridor and TileFloor to TileBridge
- Updates waterFeature.Bridges with bridge locations

**FloodFill(terrain \*Terrain, start Point, maxTiles int, rng \*rand.Rand) []Point**
- BFS traversal from starting point
- Only visits walkable tiles (TileFloor, TileCorridor, TileDoor)
- Stops when maxTiles reached or no more walkable neighbors
- Returns list of filled tiles

**FloodFillWater(terrain \*Terrain, tiles []Point, deepRatio float64, rng \*rand.Rand) WaterFeature**
- Converts flood-filled area into water body
- deepRatio controls proportion of deep vs shallow water (0.0-1.0)
- Randomly assigns depth to each tile based on ratio
- Returns WaterFeature with all tiles

### Integration with Terrain Generators

The water system integrates with all terrain generators:

**BSP Dungeons:**
- Moats around boss rooms (1-2 tile width based on room size ≥8x8)
- Creates defensive perimeters for challenging encounters

**Cellular Caves:**
- 1-3 underground lakes per map (30% chance)
- Radius 3-6 tiles, placed in open chambers
- Maintains 15+ Manhattan distance separation for variety

**Forest:**
- Natural lakes and winding rivers
- Automatic bridge placement where paths cross water
- Shallow edges for realism

**Maze:**
- Water hazards in dead ends (20% chance per dead end)
- 2-3 tile pools with deep water at end, shallow toward corridor
- Avoids room locations to preserve accessible areas

### Performance

Water generation is highly efficient:
- GenerateLake: ~0.019ms (40x30 terrain, radius 10)
- GenerateRiver: ~0.24ms (80x50 terrain, width 3)
- FloodFill: ~0.10ms (100 tile limit)
- Memory: 35-206 KB per generation operation

### Design Notes

**Depth System:**
Water uses a two-tier depth system for tactical gameplay:
- **Shallow Water** (TileWaterShallow): Walkable but slows movement (2.0x cost)
- **Deep Water** (TileWaterDeep): Impassable, blocks movement entirely

**Natural Appearance:**
- Lakes use elliptical shapes with radius variance (10-30%) for organic look
- Rivers meander with perpendicular offsets (20% of distance) to avoid straight lines
- Flood fill respects terrain connectivity for realistic water bodies

**Bridge Placement:**
- Automatic detection of water crossings in paths
- Only places bridges where water exists on opposite sides of path
- Preserves path connectivity while adding visual variety

## Multi-Level System

The multi-level system generates connected dungeons spanning multiple levels with automatic stair placement and connectivity validation. This enables traditional roguelike vertical exploration where players descend through increasingly challenging dungeon levels.

**Features:**
- Generate 1-20 connected dungeon levels
- Automatic stair placement with alignment
- Difficulty scaling with depth
- Mix different generators per level
- Connectivity validation ensures reachability
- Deterministic generation per level

**Usage:**
```go
import "github.com/opd-ai/venture/pkg/procgen/terrain"

// Create level generator
gen := terrain.NewLevelGenerator()

// Generate 5-level dungeon
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      1,
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "width":  40,
        "height": 30,
    },
}

levels, err := gen.GenerateMultiLevel(5, 12345, params)
// levels[0] = first level with stairs down
// levels[1-3] = middle levels with both stairs
// levels[4] = last level with stairs up

// Mix different generators per level
gen.SetGenerator(0, NewBSPGenerator())      // Upper dungeon
gen.SetGenerator(1, NewCellularGenerator()) // Cave layer
gen.SetGenerator(2, NewMazeGenerator())     // Deep labyrinth
levels, err = gen.GenerateMultiLevel(3, 54321, params)
```

### Level Generator

**LevelGenerator** manages multi-level dungeon creation:

```go
type LevelGenerator struct {
    generators map[int]procgen.Generator // Depth -> Generator mapping
}
```

**Key Methods:**
- `NewLevelGenerator()`: Creates new generator (defaults to BSP for all levels)
- `SetGenerator(depth, gen)`: Assign specific generator to a depth level
- `GetGenerator(depth)`: Retrieve generator for specific depth
- `GenerateMultiLevel(numLevels, seed, params)`: Generate connected levels
- `ConnectLevels(above, below, rng)`: Connect two adjacent levels with stairs
- `ValidateMultiLevelConnectivity(levels)`: Ensure all levels reachable

### Stair Placement Strategies

**PlaceStairsRandom(terrain, up, down, rng)**
- Places stairs at random walkable floor tiles
- Simple strategy suitable for any terrain type

**PlaceStairsInRoom(terrain, roomType, up, down, rng)**
- Places stairs in specific room type (Boss, Normal, etc.)
- Centers stairs in room
- Useful for placing stairs in special areas like boss rooms

**PlaceStairsSymmetric(terrain, up, down, rng)**
- Places stairs in opposite corners or edges
- Encourages full level exploration
- Creates visual balance

### Features

**Difficulty Scaling:**
- Difficulty increases with depth: `difficulty + depth * 0.1` (capped at 1.0)
- Each level becomes progressively more challenging
- Can be customized by setting different generators per depth

**Stair Alignment:**
- Stairs down in level N roughly align with stairs up in level N+1
- Searches within 10-tile radius for aligned placement
- Falls back to any walkable tile if alignment impossible
- Creates sense of vertical continuity

**Connectivity Validation:**
- First level has stairs down (if >1 level)
- Last level has stairs up (if >1 level)
- Middle levels have both stairs
- All stairs in walkable positions
- Automatic validation before returning levels

### CLI Tool Support

The `terraintest` CLI tool supports multi-level generation:

```bash
# Generate 3-level dungeon (show level 0 only)
./terraintest -algorithm multilevel -width 40 -height 30 -levels 3 -seed 12345

# Generate 5-level dungeon (show all levels)
./terraintest -algorithm multilevel -levels 5 -showAll -seed 99999

# Save to file
./terraintest -algorithm multilevel -levels 3 -showAll -output dungeon.txt
```

**Flags:**
- `-levels <n>`: Number of levels for multilevel generation (default: 1)
- `-showAll`: Show all levels instead of just first (multilevel only)

**Output Format:**
```
Multi-Level Dungeon: 3 levels
Size: 30x20 per level, Seed: 99999

=== LEVEL 0 ===
[ASCII map with 'v' showing stairs down]
Connections:
  Stairs Down: [{14 7}]
  Stairs Up (next level): [{22 4}]

=== LEVEL 1 ===
[ASCII map with '^' up and 'v' down stairs]
...
```

### Performance

Multi-level generation is highly efficient:
- 3 levels (40x30): ~0.051ms
- 5 levels (50x40): ~0.098ms
- 20 levels (max): ~400ms estimated
- Linear scaling: ~20μs per level

### Design Notes

**Depth-Specific Seeds:**
- Each level uses `baseSeed + level * 1000` for independent but deterministic generation
- Ensures reproducibility across multiplayer clients
- Allows individual level regeneration if needed

**Generator Mixing:**
- Different algorithms can be used for different depths
- Example: BSP upper levels, Cellular caves mid-game, Maze deep dungeons
- Creates varied gameplay experience as player descends

**Fallback Strategy:**
- Stair placement tries alignment first, falls back to any walkable tile
- Prevents generation failures due to impossible layouts
- Maintains high generation success rate

## Tile Types

The terrain system uses several tile types:

### Basic Tiles
- **TileWall**: Solid walls that block movement
- **TileFloor**: Walkable floor space (rooms)
- **TileCorridor**: Walkable passages connecting areas
- **TileDoor**: Doorways (walkable, but can be closed)

### Water Tiles
- **TileWaterShallow**: Walkable shallow water (movement cost: 2.0x, slows movement)
- **TileWaterDeep**: Impassable deep water

### Natural Obstacles
- **TileTree**: Tree or natural obstacle that blocks movement

### Level Transitions
- **TileStairsUp**: Stairs leading to an upper level
- **TileStairsDown**: Stairs leading to a lower level
- **TileTrapDoor**: Hidden or revealed trap door (movement cost: 1.5x)

### Special Tiles
- **TileSecretDoor**: Hidden door (blocks vision until discovered)
- **TileBridge**: Walkable bridge over water
- **TileStructure**: Building or ruins that blocks movement

### Tile Properties

Each tile type has three key properties:

```go
// Check if a tile type is walkable
isWalkable := tile.IsWalkableTile()

// Check if a tile type blocks vision
isTransparent := tile.IsTransparent()

// Get movement cost multiplier (1.0 = normal, 2.0 = half speed, -1 = impassable)
cost := tile.MovementCost()
```

## Terrain Structure

### Terrain Type

The `Terrain` type represents a generated map:

```go
type Terrain struct {
    Width      int           // Map width in tiles
    Height     int           // Map height in tiles
    Tiles      [][]TileType  // 2D grid of tiles
    Rooms      []*Room       // List of rooms (BSP only)
    Seed       int64         // Generation seed
    Level      int           // Dungeon level (0 = first level)
    StairsUp   []Point       // Positions of upward stairs
    StairsDown []Point       // Positions of downward stairs
}
```

**Methods:**
- `GetTile(x, y int) TileType` - Safely get a tile (returns wall for out-of-bounds)
- `SetTile(x, y int, tileType TileType)` - Safely set a tile
- `IsWalkable(x, y int) bool` - Check if a position is walkable
- `AddStairs(x, y int, up bool)` - Add stairs at the specified position
- `IsInBounds(x, y int) bool` - Check if coordinates are within terrain bounds
- `ValidateStairPlacement() error` - Validate that all stairs are placed correctly

### Point Type

The `Point` type represents a 2D coordinate:

```go
type Point struct {
    X, Y int
}
```

**Methods:**
- `Distance(other Point) float64` - Calculate Euclidean distance
- `ManhattanDistance(other Point) int` - Calculate Manhattan (taxicab) distance
- `Equals(other Point) bool` - Check if two points are equal
- `IsInBounds(width, height int) bool` - Check if point is within bounds
- `Neighbors() []Point` - Get the four orthogonal neighbors
- `AllNeighbors() []Point` - Get all eight neighbors (orthogonal and diagonal)

### Room Type

The `Room` type represents a rectangular room (BSP algorithm):

```go
type Room struct {
    X, Y          int // Top-left corner
    Width, Height int // Room dimensions
}
```

**Methods:**
- `Center() (int, int)` - Get the center coordinates
- `Overlaps(other *Room) bool` - Check if two rooms overlap

## Testing

Run the terrain generation tests:

```bash
go test ./pkg/procgen/terrain/...
```

Run with coverage:

```bash
go test -cover ./pkg/procgen/terrain/...
```

Run benchmarks:

```bash
go test -bench=. ./pkg/procgen/terrain/...
```

## CLI Tool

A command-line tool is provided for testing and visualizing terrain generation:

```bash
# Build the tool
go build -o terraintest ./cmd/terraintest

# Generate BSP dungeon
./terraintest -algorithm bsp -width 80 -height 50 -seed 12345

# Generate cellular caves
./terraintest -algorithm cellular -width 80 -height 50 -seed 54321

# Generate maze
./terraintest -algorithm maze -width 81 -height 81 -seed 99999

# Generate forest
./terraintest -algorithm forest -width 80 -height 50 -seed 12345

# Generate city
./terraintest -algorithm city -width 80 -height 50 -seed 67890

# Save to file
./terraintest -algorithm bsp -output dungeon.txt
```

### CLI Options

- `-algorithm` - Generation algorithm: "bsp", "cellular", "maze", "forest", or "city" (default: "bsp")
- `-width` - Map width in tiles (default: 80)
- `-height` - Map height in tiles (default: 50)
- `-seed` - Random seed for deterministic generation (default: 12345)
- `-output` - Output file path (default: console output)

### ASCII Visualization

The CLI tool renders terrain as ASCII art:

- `#` - Wall
- `.` - Floor
- `:` - Corridor
- `+` - Door
- `W` - Shallow Water
- `~` - Deep Water
- `T` - Tree
- `^` - Stairs Up
- `v` - Stairs Down
- `[` - Trap Door
- `?` - Secret Door
- `=` - Bridge
- `@` - Structure

## Performance

All algorithms are designed to be fast and efficient:

- **BSP**: O(n) where n is the number of nodes in the BSP tree
- **Cellular**: O(w × h × i) where w and h are dimensions and i is iterations
- **Maze**: O(w × h) recursive backtracking with stack-based traversal

Typical generation times:
- 80×50 BSP dungeon: < 1ms
- 80×50 Cellular caves: 2-5ms
- 81×81 Maze: 2-10ms

## Determinism

All generators produce identical output for the same seed and parameters. This is critical for:
- Multiplayer synchronization
- Reproducible worlds
- Testing and debugging
- Speedrunning and community sharing

## Future Enhancements

Potential additions to the terrain system:

- [x] **Point utilities for coordinates** (✓ Completed: Phase 1)
- [x] **Multi-level dungeons with stairs** (✓ Completed: Phase 1 - Infrastructure)
- [x] **Extended tile types** (water, trees, bridges, structures) (✓ Completed: Phase 1)
- [x] **Maze generator** (✓ Completed: Phase 2 - recursive backtracking)
- [x] **Forest generator** (✓ Completed: Phase 3 - Poisson disc sampling)
- [x] **City generator** (✓ Completed: Phase 4 - grid subdivision)
- [x] **Water system** (✓ Completed: Phase 5 - lakes, rivers, moats, bridges)
- [x] **Multi-level generator** (✓ Completed: Phase 6 - connects levels, 1-20 levels, stair placement)
- [ ] Room templates and prefabs
- [ ] Door placement algorithms
- [ ] Treasure room generation
- [ ] Themed room variants (treasure, boss, puzzle)
- [ ] **Composite generator** (Phase 7 - multi-biome maps)
- [ ] Drunkard's walk algorithm
- [ ] Voronoi diagram-based generation
- [ ] Integration with entity placement system

See `PLAN.md` in the project root for the complete terrain expansion roadmap.
