# Post-Processing System Guide

**Phase 5.3 - Post-Processing Effects Integration**

The post-processing system provides cinematic visual effects that can be applied to the game scene for enhanced visual quality. Effects are opt-in (disabled by default) to maintain baseline performance.

## Features

### Available Effects

1. **Color Grading**: Adjust saturation, contrast, brightness, temperature (warm/cool), and tint (green/magenta)
2. **Vignette**: Edge darkening for cinematic feel with configurable intensity and softness
3. **Chromatic Aberration**: Color channel separation for analog camera aesthetic
4. **Motion Blur**: Velocity-based directional blur (requires velocity maps, not yet exposed)
5. **Depth Blur**: Depth-of-field effect (requires depth maps, not yet exposed)

### Genre Presets

Pre-configured visual styles for different game genres:

- **Fantasy**: Warm colors (+15% temperature), high saturation (1.2), soft vignette (0.4)
- **Sci-Fi**: Cool colors (-20% temperature), high contrast (1.25), chromatic aberration (0.15)
- **Horror**: Desaturated (0.6), dark (-15% brightness), strong vignette (0.7)
- **Cyberpunk**: High saturation (1.4), harsh contrast (1.4), chromatic aberration (0.3)
- **Post-Apocalyptic**: Dusty tint, low saturation (0.7), harsh vignette (0.6)
- **Neutral**: Balanced settings for general use
- **Cinematic**: Film-like look with moderate vignette and color grading

## Command-Line Usage

### Enable with Preset

Apply a genre-specific visual style:

```bash
# Fantasy (warm, saturated)
./client --enable-postprocessing --postprocess-preset fantasy

# Sci-Fi (cool, high contrast)
./client --enable-postprocessing --postprocess-preset scifi

# Horror (desaturated, dark)
./client --enable-postprocessing --postprocess-preset horror

# Cyberpunk (neon, harsh)
./client --enable-postprocessing --postprocess-preset cyberpunk

# Cinematic (film-like)
./client --enable-postprocessing --postprocess-preset cinematic
```

### Custom Configuration

Enable specific effects with custom parameters:

```bash
# Color grading only
./client --enable-postprocessing \
  --postprocess-color-grading \
  --postprocess-saturation 1.3 \
  --postprocess-contrast 1.2 \
  --postprocess-brightness 0.1

# Vignette only
./client --enable-postprocessing \
  --postprocess-vignette \
  --postprocess-vignette-intensity 0.6 \
  --postprocess-vignette-softness 0.4

# Chromatic aberration only
./client --enable-postprocessing \
  --postprocess-chromatic \
  --postprocess-chromatic-intensity 0.5

# Combine multiple effects
./client --enable-postprocessing \
  --postprocess-color-grading \
  --postprocess-vignette \
  --postprocess-saturation 1.2 \
  --postprocess-contrast 1.1 \
  --postprocess-vignette-intensity 0.5
```

## Available Flags

### Master Control
- `--enable-postprocessing` (default: false) - Enable/disable all post-processing

### Preset
- `--postprocess-preset` (default: "") - Apply genre preset (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic, neutral, cinematic)

### Effect Toggles
- `--postprocess-color-grading` (default: false) - Enable color grading
- `--postprocess-vignette` (default: false) - Enable vignette
- `--postprocess-chromatic` (default: false) - Enable chromatic aberration

### Color Grading Parameters
- `--postprocess-saturation` (default: 1.0) - Saturation multiplier (0.0-2.0)
  - 0.0 = grayscale
  - 1.0 = normal
  - 2.0 = highly saturated
- `--postprocess-contrast` (default: 1.0) - Contrast multiplier (0.0-2.0)
  - 0.5 = low contrast
  - 1.0 = normal
  - 1.5 = high contrast
- `--postprocess-brightness` (default: 0.0) - Brightness adjustment (-1.0 to 1.0)
  - -0.5 = darker
  - 0.0 = normal
  - 0.5 = brighter

### Vignette Parameters
- `--postprocess-vignette-intensity` (default: 0.5) - Darkness intensity (0.0-1.0)
  - 0.0 = no vignette
  - 0.5 = moderate darkening
  - 1.0 = strong darkening
- `--postprocess-vignette-softness` (default: 0.3) - Edge softness (0.0-1.0)
  - 0.0 = hard edge
  - 0.5 = soft edge
  - 1.0 = very soft edge

### Chromatic Aberration Parameters
- `--postprocess-chromatic-intensity` (default: 0.5) - Aberration strength (0.0-1.0)
  - 0.0 = no aberration
  - 0.3 = subtle effect
  - 0.7 = strong effect

## Performance

### Frame Time Impact (800x600 resolution)

When all effects are enabled:
- Color Grading: ~2-5ms
- Vignette: ~1-3ms
- Chromatic Aberration: ~3-8ms
- **Total**: ~10-30ms overhead per frame

At 60 FPS target (16.67ms/frame):
- Post-processing adds 60-180% overhead when fully enabled
- Recommended for 120+ FPS capable systems or lower resolutions
- Disabled by default for optimal baseline performance

### Optimization Tips

1. **Use Presets**: Presets are optimized for visual quality/performance balance
2. **Selective Effects**: Enable only needed effects (e.g., vignette only)
3. **Lower Resolution**: Post-processing cost scales with pixel count
4. **Disable When Needed**: Turn off during intense gameplay (combat, raids)

## Integration Details

### Rendering Pipeline

Post-processing is applied in the rendering pipeline:

1. Terrain rendering
2. Entity rendering
3. Lighting effects (if enabled)
4. **Post-processing** ← Applied here
5. UI overlays

### Technical Implementation

- **Package**: `pkg/rendering/postprocess` (core implementation)
- **Adapter**: `pkg/engine/PostProcessorAdapter` (ECS integration)
- **Test Coverage**: 84.4% (postprocess), 56.3% (engine)
- **Image Conversion**: Ebiten → RGBA → Processing → Ebiten
- **Buffer Reuse**: Scene buffer reused for efficiency

## Examples

### Match Visual Style to Genre

Automatically apply post-processing that matches your game genre:

```bash
# Fantasy game with matching visual style
./client --genre fantasy --enable-postprocessing --postprocess-preset fantasy

# Sci-fi game with cool, high-contrast look
./client --genre scifi --enable-postprocessing --postprocess-preset scifi

# Horror game with dark, desaturated atmosphere
./client --genre horror --enable-postprocessing --postprocess-preset horror
```

### Cinematic Screenshots

Enable post-processing for high-quality screenshots:

```bash
# Cinematic mode with all effects
./client --enable-postprocessing --postprocess-preset cinematic

# Custom "photo mode" settings
./client --enable-postprocessing \
  --postprocess-color-grading \
  --postprocess-vignette \
  --postprocess-saturation 1.3 \
  --postprocess-contrast 1.2 \
  --postprocess-vignette-intensity 0.4 \
  --postprocess-vignette-softness 0.6
```

### Vintage/Retro Look

Create a vintage aesthetic with desaturation and vignette:

```bash
./client --enable-postprocessing \
  --postprocess-color-grading \
  --postprocess-vignette \
  --postprocess-saturation 0.8 \
  --postprocess-contrast 1.1 \
  --postprocess-brightness -0.05 \
  --postprocess-vignette-intensity 0.5
```

## Future Enhancements

Planned improvements for post-processing system:

1. **In-Game Toggle**: Runtime enable/disable via UI (e.g., P key)
2. **Velocity Maps**: Expose motion blur with automatic velocity calculation
3. **Depth Maps**: Expose depth-of-field blur with Z-buffer generation
4. **More Presets**: Add game-mode-specific presets (combat, exploration, dialogue)
5. **Quality Levels**: Low/Medium/High quality presets for performance tuning
6. **Bloom Integration**: Combine with lighting system bloom for unified glow effects

## See Also

- [Lighting System](LIGHTING_SYSTEM.md) - Dynamic lighting with bloom and ambient occlusion
- [Tile Rendering](docs/PLAN.md#52-tile-rendering) - Advanced tile rendering features
- [Performance Guide](PERFORMANCE.md) - Performance optimization techniques
- Phase 5.3 in [PLAN.md](PLAN.md) - Implementation details and testing

## Support

Post-processing is a **Phase 5.3** feature. For issues or questions:
- Check test coverage: `go test ./pkg/rendering/postprocess/... -cover`
- Review logs with `--verbose` flag
- Disable if experiencing performance issues: `--enable-postprocessing=false`
