# Parallax Depth Effects Test Tool

CLI tool for testing and visualizing Phase 16.3 parallax depth tile generation.

## Building

```bash
go build -o parallaxtest ./cmd/parallaxtest/
```

## Usage

```bash
./parallaxtest [flags]
```

### Flags

- `-output` - Output directory for generated images (default: `./parallax_output`)
- `-type` - Tile type: floor, wall, door, corridor, water, lava, trap, stairs (default: `floor`)
- `-genre` - Genre ID: fantasy, scifi, horror, cyberpunk, postapocalyptic (default: `fantasy`)
- `-seed` - Random seed for generation (default: `12345`)
- `-width` - Tile width in pixels (default: `32`)
- `-height` - Tile height in pixels (default: `32`)
- `-camera-x` - Camera X position for parallax (default: `10.0`)
- `-camera-y` - Camera Y position for parallax (default: `5.0`)
- `-ao` - Ambient occlusion intensity 0.0-1.0 (default: `0.5`)
- `-shadow` - Shadow height 0.0-1.0 (default: `0.3`)

### Examples

Generate fantasy wall tiles with default settings:
```bash
./parallaxtest -type wall -genre fantasy
```

Generate sci-fi floor tiles at 64x64 with high AO:
```bash
./parallaxtest -type floor -genre scifi -width 64 -height 64 -ao 0.8
```

Generate horror tiles with custom camera position:
```bash
./parallaxtest -type corridor -genre horror -camera-x 20 -camera-y 15
```

## Output

The tool generates 11 PNG images demonstrating different aspects of the parallax depth system:

1. `01_base_tile.png` - Original tile without any effects
2. `02_background_layer.png` - Background layer (darkened, slow parallax)
3. `03_base_layer_with_effects.png` - Base layer with AO and shadows
4. `04_foreground_layer.png` - Foreground layer (brightened, fast parallax)
5. `05_layered_background.png` - Background from layered generation
6. `06_layered_base.png` - Base from layered generation
7. `07_layered_foreground.png` - Foreground from layered generation
8. `08_composite.png` - All layers composited together
9. `09_ao_map.png` - Ambient occlusion map (grayscale)
10. `10_no_effects.png` - Comparison: no effects applied
11. `11_with_effects.png` - Comparison: full effects applied

## Technical Details

### Layer System

- **Background Layer**: Furthest layer, moves slower (parallax depth 0.2-0.4)
  - 30% darkening applied
  - Slight desaturation for atmospheric perspective
  - Reduced AO and shadow intensity

- **Base Layer**: Main tile content, moves with camera (parallax depth 1.0)
  - Full ambient occlusion applied
  - Height-based shadows with configurable angle
  - Standard color rendering

- **Foreground Layer**: Closest layer, moves faster (parallax depth 1.2-1.5)
  - 10% brightening for emphasis
  - Enhanced shadow height for depth
  - Reduced AO intensity

### Parallax Calculation

Offset is calculated as: `cameraPosition * parallaxDepth * layerMultiplier`

- Background: multiplier = 0.3 (moves 30% of camera movement)
- Base: multiplier = 1.0 (moves 100% of camera movement)
- Foreground: multiplier = 1.4 (moves 140% of camera movement)

### Ambient Occlusion

- Uses 3x3 neighborhood edge detection
- Compares brightness of neighboring pixels
- Darkens areas where geometry transitions occur
- Configurable intensity (0.0 = none, 1.0 = maximum)

### Height-Based Shadows

- Shadows cast based on tile type and height parameter
- Wall tiles cast stronger shadows than floor tiles
- Shadow angle and length configurable
- Edge detection determines shadow origin points

## Performance

- Single layer generation: ~161µs per 32x32 tile
- All three layers: ~305µs per 32x32 tile
- Layer compositing: ~17µs per 32x32 tile
- Total overhead: <1% frame time increase vs base rendering

## Testing

Generate tiles with different parameters to verify:
- Deterministic generation (same seed = identical output)
- Layer visual differences (brightness variations)
- AO map quality (corners and edges darkened)
- Shadow placement and intensity
- Parallax offset calculations
