# Tile Rendering System

The tiles package provides procedural tile image generation for terrain rendering in the Venture game.

## Overview

This package generates visual representations of terrain tiles using procedural techniques. It supports multiple tile types and can generate genre-specific visual styles, all deterministically based on seeds.

## Features

- **8 Tile Types**: Floor, Wall, Door, Corridor, Water, Lava, Trap, Stairs
- **Genre-Aware Styling**: Automatic color palette selection based on game genre
- **6 Visual Patterns**: Solid, Checkerboard, Dots, Lines, Brick, Grain
- **Deterministic Generation**: Same seed produces identical tiles
- **Configurable Variants**: Control visual variation with the variant parameter
- **High Test Coverage**: 92.6% test coverage

## Installation

```bash
go get github.com/opd-ai/venture/pkg/rendering/tiles
```

## Usage

### Basic Example

```go
import "github.com/opd-ai/venture/pkg/rendering/tiles"

// Create a generator
gen := tiles.NewGenerator()

// Configure tile generation
config := tiles.Config{
    Type:    tiles.TileFloor,
    Width:   32,
    Height:  32,
    GenreID: "fantasy",
    Seed:    12345,
    Variant: 0.5,
}

// Generate the tile
img, err := gen.Generate(config)
if err != nil {
    log.Fatal(err)
}

// Use the generated image...
```

### Generate Different Tile Types

```go
tileTypes := []tiles.TileType{
    tiles.TileFloor,
    tiles.TileWall,
    tiles.TileDoor,
    tiles.TileCorridor,
}

for _, tileType := range tileTypes {
    config := tiles.Config{
        Type:    tileType,
        Width:   32,
        Height:  32,
        GenreID: "scifi",
        Seed:    12345,
        Variant: 0.5,
    }
    
    img, _ := gen.Generate(config)
    // Use img...
}
```

### Using the CLI Tool

```bash
# Build the tool
go build -o tiletest ./cmd/tiletest

# Generate fantasy floor tiles
./tiletest -type floor -genre fantasy -count 10

# Generate sci-fi wall with verbose output
./tiletest -type wall -genre scifi -verbose

# Generate and save a tile
./tiletest -type door -output tile.png

# Generate with custom parameters
./tiletest -type trap -width 64 -height 64 -variant 0.8 -seed 99999
```

### CLI Options

- `-type`: Tile type (floor, wall, door, corridor, water, lava, trap, stairs)
- `-width`: Tile width in pixels (default: 32)
- `-height`: Tile height in pixels (default: 32)
- `-genre`: Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)
- `-seed`: Random seed for generation (default: 12345)
- `-variant`: Visual variant 0.0-1.0 (default: 0.5)
- `-count`: Number of tiles to generate (default: 10)
- `-output`: Output PNG file (saves first tile)
- `-verbose`: Show detailed output

## Tile Types

### TileFloor
Walkable floor tiles with subtle textures. Uses solid, checkerboard, or dot patterns.

### TileWall
Solid wall tiles with brick or stone patterns. Typically darker and more prominent.

### TileDoor
Door tiles with wood grain pattern and frame. Represents entryways between rooms.

### TileCorridor
Corridor tiles similar to floors but darker, representing connecting passages.

### TileWater
Water tiles with blue tones. Represents bodies of water.

### TileLava
Lava tiles with red/orange tones. Represents dangerous molten areas.

### TileTrap
Trap tiles that look like floors but with danger indicators in the center.

### TileStairs
Stair tiles with horizontal step lines indicating vertical movement.

## Configuration

### Config Structure

```go
type Config struct {
    Type    TileType              // Tile type to generate
    Width   int                   // Width in pixels
    Height  int                   // Height in pixels
    GenreID string                // Genre for styling
    Seed    int64                 // Seed for determinism
    Variant float64               // Visual variation (0.0-1.0)
    Custom  map[string]interface{} // Custom parameters
}
```

### Validation

All configurations are validated before generation:
- Width and height must be positive
- Variant must be between 0.0 and 1.0
- GenreID cannot be empty

## Visual Patterns

The package supports multiple visual patterns:

- **PatternSolid**: Uniform color with subtle random variation
- **PatternCheckerboard**: Alternating light/dark squares
- **PatternDots**: Dots on solid background
- **PatternLines**: Parallel horizontal lines
- **PatternBrick**: Brick-like mortar pattern (for walls)
- **PatternGrain**: Wood grain texture (for doors)

Pattern selection is influenced by tile type and variant parameter.

## Integration with Terrain System

The tiles package integrates seamlessly with the terrain generation system:

```go
import (
    "github.com/opd-ai/venture/pkg/procgen/terrain"
    "github.com/opd-ai/venture/pkg/rendering/tiles"
)

// Generate terrain
terrainGen := terrain.NewBSPGenerator()
terrainMap, _ := terrainGen.Generate(12345, params)

// Generate matching tiles
tileGen := tiles.NewGenerator()

for y := 0; y < terrainMap.Height; y++ {
    for x := 0; x < terrainMap.Width; x++ {
        terrainTile := terrainMap.GetTile(x, y)
        
        // Map terrain type to tile type
        var tileType tiles.TileType
        switch terrainTile {
        case terrain.TileFloor:
            tileType = tiles.TileFloor
        case terrain.TileWall:
            tileType = tiles.TileWall
        case terrain.TileDoor:
            tileType = tiles.TileDoor
        case terrain.TileCorridor:
            tileType = tiles.TileCorridor
        }
        
        config := tiles.Config{
            Type:    tileType,
            Width:   32,
            Height:  32,
            GenreID: "fantasy",
            Seed:    12345 + int64(y*terrainMap.Width+x),
            Variant: 0.5,
        }
        
        img, _ := tileGen.Generate(config)
        // Render img at position (x*32, y*32)
    }
}
```

## Performance

- **Generation Time**: ~50-200μs per 32x32 tile
- **Memory**: Minimal overhead, images use standard Go image buffers
- **Determinism**: Identical output for same seed across platforms

## Testing

```bash
# Run tests
go test -tags test ./pkg/rendering/tiles/...

# Run with coverage
go test -tags test -cover ./pkg/rendering/tiles/...

# Run benchmarks
go test -tags test -bench=. ./pkg/rendering/tiles/...
```

Test coverage: **92.6%**

## Architecture

The tile generator uses:
1. **Color Palette Generation**: Genre-based color schemes from the palette package
2. **Pattern Functions**: Specialized functions for each visual pattern
3. **Procedural Drawing**: Direct pixel manipulation for pattern generation
4. **Deterministic RNG**: Seeded random number generators for reproducibility

## Future Enhancements

- Animation frame generation for animated tiles
- Normal map generation for 3D lighting effects
- Texture overlay support (noise, scratches, wear)
- Multi-layer tile composition
- Shadow and highlight generation
- Edge detection and auto-tiling

## License

See project LICENSE file.
