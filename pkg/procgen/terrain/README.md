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

## Tile Types

The terrain system uses several tile types:

- **TileWall**: Solid walls that block movement
- **TileFloor**: Walkable floor space (rooms)
- **TileCorridor**: Walkable passages connecting areas
- **TileDoor**: Doorways (walkable, but can be closed)

## Terrain Structure

### Terrain Type

The `Terrain` type represents a generated map:

```go
type Terrain struct {
    Width  int           // Map width in tiles
    Height int           // Map height in tiles
    Tiles  [][]TileType  // 2D grid of tiles
    Rooms  []*Room       // List of rooms (BSP only)
    Seed   int64         // Generation seed
}
```

**Methods:**
- `GetTile(x, y int) TileType` - Safely get a tile (returns wall for out-of-bounds)
- `SetTile(x, y int, tileType TileType)` - Safely set a tile
- `IsWalkable(x, y int) bool` - Check if a position is walkable

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

# Save to file
./terraintest -algorithm bsp -output dungeon.txt
```

### CLI Options

- `-algorithm` - Generation algorithm: "bsp" or "cellular" (default: "bsp")
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

## Performance

Both algorithms are designed to be fast and efficient:

- **BSP**: O(n) where n is the number of nodes in the BSP tree
- **Cellular**: O(w × h × i) where w and h are dimensions and i is iterations

Typical generation times:
- 80×50 BSP dungeon: < 1ms
- 80×50 Cellular caves: 2-5ms

## Determinism

All generators produce identical output for the same seed and parameters. This is critical for:
- Multiplayer synchronization
- Reproducible worlds
- Testing and debugging
- Speedrunning and community sharing

## Future Enhancements

Potential additions to the terrain system:

- [ ] Room templates and prefabs
- [ ] Door placement algorithms
- [ ] Treasure room generation
- [ ] Multi-level dungeons with stairs
- [ ] Themed room variants (treasure, boss, puzzle)
- [ ] Drunkard's walk algorithm
- [ ] Voronoi diagram-based generation
- [ ] Integration with entity placement system
