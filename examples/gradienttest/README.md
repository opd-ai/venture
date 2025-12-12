# Gradient Test Tool

CLI tool for testing and visualizing procedural gradient generation with mood-based color palettes.

Phase 19.2: Dynamic Palette System - Gradient Generation Test Tool

## Building

```bash
go build ./cmd/gradienttest
```

## Usage

```bash
./gradienttest [options]
```

### Options

- `-type` - Gradient type: `linear`, `radial`, `angular`, `diamond`, `spiral`, `conic` (default: `linear`)
- `-width` - Width of gradient image in pixels (default: `800`)
- `-height` - Height of gradient image in pixels (default: `600`)
- `-angle` - Angle for linear/angular/conic gradients in degrees 0-360 (default: `0`)
- `-centerx` - Center X position 0.0-1.0 for radial/angular gradients (default: `0.5`)
- `-centery` - Center Y position 0.0-1.0 for radial/angular gradients (default: `0.5`)
- `-radius` - Radius for radial gradients 0.0-1.0 (default: `0.5`)
- `-rotations` - Number of rotations for spiral gradients (default: `2.0`)
- `-reverse` - Reverse gradient direction (default: `false`)
- `-smoothness` - Smoothness factor 0.0-1.0 for easing (default: `0.5`)
- `-genre` - Genre for palette colors: `fantasy`, `scifi`, `horror`, `cyberpunk`, `postapoc` (default: `fantasy`)
- `-mood` - Mood for palette colors (see Mood Types below, default: `normal`)
- `-steps` - Number of color steps in gradient (default: `5`)
- `-output` - Output PNG file path (default: `gradient.png`)
- `-verbose` - Enable verbose logging (default: `false`)

## Mood Types

Phase 19.2 supports 24 mood types for emotional color adjustments:

**Original Moods (7):**
- `normal` - Standard color values
- `bright` - Increased lightness for cheerful tone
- `dark` - Decreased lightness for somber tone
- `saturated` - Increased saturation for intense colors
- `muted` - Decreased saturation for subdued colors
- `vibrant` - Maximum saturation and lightness
- `pastel` - High lightness with low saturation

**New Moods (17):**
- `tense` - Anxiety with desaturated, dark colors
- `calm` - Peace with soft, balanced colors
- `victorious` - Triumph with bright, saturated golds
- `melancholic` - Sadness with desaturated blues
- `energetic` - Excitement with bright, warm colors
- `mystical` - Wonder with purples and deep blues
- `ominous` - Dread with dark reds and blacks
- `serene` - Tranquility with soft blues and greens
- `aggressive` - Intensity with high saturation reds
- `playful` - Fun with varied bright colors
- `somber` - Gravity with dark, desaturated colors
- `ethereal` - Otherworldly with high lightness pastels
- `dangerous` - Threat with deep reds and oranges
- `peaceful` - Harmony with balanced greens
- `chaotic` - Disorder with high variation
- `regal` - Majesty with purples and golds
- `desolate` - Emptiness with grays and browns

## Examples

### Linear Gradient (Victorious Mood)
```bash
./gradienttest -type linear -mood victorious -angle 45 -output victory_gradient.png
```

### Radial Gradient (Calm Mood, Sci-Fi Genre)
```bash
./gradienttest -type radial -mood calm -genre scifi -output scifi_calm.png
```

### Angular Gradient (Mystical Mood)
```bash
./gradienttest -type angular -mood mystical -output mystical_angular.png
```

### Spiral Gradient (Chaotic Mood)
```bash
./gradienttest -type spiral -mood chaotic -rotations 5 -output chaotic_spiral.png
```

### Diamond Gradient (Peaceful Mood)
```bash
./gradienttest -type diamond -mood peaceful -genre fantasy -output peaceful_diamond.png
```

### Conic Gradient (Energetic Mood)
```bash
./gradienttest -type conic -mood energetic -angle 90 -output energetic_conic.png
```

## Output

The tool generates a PNG image at the specified output path and prints:
- Gradient generation parameters
- Color samples from 5 positions (corners and center)
- Total pixel count
- Performance statistics

## Performance

Typical generation times (256×256):
- Linear gradient: ~4.1ms
- Radial gradient: ~2.8ms
- Angular gradient: ~4.2ms
- Spiral gradient: ~3.5ms
- Diamond gradient: ~3.0ms
- Conic gradient: ~4.0ms

Palette generation: <1ms (12-15µs typical)

All generation is deterministic with seed-based algorithms.
