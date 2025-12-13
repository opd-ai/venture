# Terrain Generation

Procedural terrain and dungeon generation algorithms for the Venture game. All generators are deterministic based on seed values and produce varied, interesting layouts suitable for classic roguelike exploration.

## Algorithms

| Algorithm | Description | Performance |
|-----------|-------------|-------------|
| BSP | Structured dungeons with rooms and corridors | <1ms (80×50) |
| Cellular | Organic cave-like structures | 2-5ms (80×50) |
| Maze | Winding corridors with recursive backtracking | 2-10ms (81×81) |
| Forest | Natural outdoor with Poisson disc sampling | ~0.8ms (80×50) |
| City | Urban grid with buildings and streets | ~0.06ms (80×50) |

## Quick Start

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

## Tile Types

| Category | Tiles | Description |
|----------|-------|-------------|
| Basic | Wall, Floor, Corridor, Door | Core walkable/blocking tiles |
| Water | WaterShallow (2x cost), WaterDeep (impassable) | Water features |
| Natural | Tree | Blocking natural obstacles |
| Transitions | StairsUp, StairsDown, TrapDoor | Level connections |
| Special | SecretDoor, Bridge, Structure | Special features |

## Water System

- **Lakes**: Elliptical with shallow edges, deep center
- **Rivers**: Winding paths between points
- **Moats**: Defensive perimeters around rooms
- **Bridges**: Auto-placed where paths cross water

## Multi-Level System

```go
gen := terrain.NewLevelGenerator()
levels, err := gen.GenerateMultiLevel(5, 12345, params)
// levels[0] = first level with stairs down
// levels[4] = last level with stairs up
```

Features:
- 1-20 connected levels with automatic stair placement
- Difficulty scaling with depth
- Mix different generators per level
- Stair alignment for vertical continuity

## Configuration

| Algorithm | Key Parameters |
|-----------|----------------|
| BSP | minRoomSize, maxRoomSize, splitDepth |
| Cellular | fillProbability (0.40), iterations (5) |
| Maze | roomChance (0.1), corridorWidth (1-2) |
| Forest | treeDensity (0.3), clearingCount (3), waterChance (0.3) |
| City | blockSize (12), streetWidth (2), buildingDensity (0.7) |

## Terrain Structure

```go
type Terrain struct {
    Width, Height  int
    Tiles          [][]TileType
    Rooms          []*Room
    Seed           int64
    Level          int
    StairsUp       []Point
    StairsDown     []Point
}
```

Methods: `GetTile()`, `SetTile()`, `IsWalkable()`, `AddStairs()`, `IsInBounds()`

## CLI Tool

```bash
go build -o terraintest ./cmd/terraintest

./terraintest -algorithm bsp -width 80 -height 50 -seed 12345
./terraintest -algorithm multilevel -levels 5 -showAll
```

Options: `-algorithm`, `-width`, `-height`, `-seed`, `-output`, `-levels`, `-showAll`

## Testing

```bash
go test ./pkg/procgen/terrain/...
go test -cover ./pkg/procgen/terrain/...
go test -bench=. ./pkg/procgen/terrain/...
```

## Determinism

All generators produce identical output for the same seed and parameters, ensuring:
- Multiplayer synchronization
- Reproducible worlds
- Testing and debugging
