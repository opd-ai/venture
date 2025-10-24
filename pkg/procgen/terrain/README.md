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

# Save to file
./terraintest -algorithm bsp -output dungeon.txt
```

### CLI Options

- `-algorithm` - Generation algorithm: "bsp", "cellular", or "maze" (default: "bsp")
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
- [ ] Room templates and prefabs
- [ ] Door placement algorithms
- [ ] Treasure room generation
- [ ] **Multi-level generator** (Phase 6 - connects levels)
- [ ] Themed room variants (treasure, boss, puzzle)
- [ ] **Forest generator** (Phase 3 - natural environments)
- [ ] **City generator** (Phase 4 - urban environments)
- [ ] **Water system** (Phase 5 - lakes, rivers, moats)
- [ ] **Composite generator** (Phase 7 - multi-biome maps)
- [ ] Drunkard's walk algorithm
- [ ] Voronoi diagram-based generation
- [ ] Integration with entity placement system

See `PLAN.md` in the project root for the complete terrain expansion roadmap.
