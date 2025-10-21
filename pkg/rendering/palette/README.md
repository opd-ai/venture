# Palette Generator

The palette package provides procedural color palette generation for genre-based theming in the Venture game.

## Features

- **Genre-Aware**: Generates palettes that match the visual style of each game genre (Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic)
- **Deterministic**: Same seed always produces the same palette
- **HSL Color Space**: Uses Hue-Saturation-Lightness for intuitive color generation
- **Comprehensive**: Provides primary, secondary, background, text, accent, danger, and success colors
- **8 Additional Colors**: Extra palette colors for variety and detail

## Usage

```go
import "github.com/opd-ai/venture/pkg/rendering/palette"

// Create a generator
gen := palette.NewGenerator()

// Generate a fantasy-themed palette
pal, err := gen.Generate("fantasy", 12345)
if err != nil {
    log.Fatal(err)
}

// Use the palette colors
fmt.Printf("Primary: %v\n", pal.Primary)
fmt.Printf("Secondary: %v\n", pal.Secondary)
fmt.Printf("Background: %v\n", pal.Background)
```

## Color Schemes by Genre

### Fantasy
- **Base Hue**: 30° (warm earthy tones)
- **Saturation**: 0.6
- **Lightness**: 0.5
- **Style**: Browns, golds, royal blues

### Sci-Fi
- **Base Hue**: 210° (cool blues and cyans)
- **Saturation**: 0.7
- **Lightness**: 0.5
- **Style**: Techy blues, metallic colors

### Horror
- **Base Hue**: 0° (desaturated reds and grays)
- **Saturation**: 0.3
- **Lightness**: 0.3
- **Style**: Dark, muted, atmospheric

### Cyberpunk
- **Base Hue**: 300° (neon purples and magentas)
- **Saturation**: 0.9
- **Lightness**: 0.5
- **Style**: Vibrant neons, high contrast

### Post-Apocalyptic
- **Base Hue**: 45° (dusty browns and oranges)
- **Saturation**: 0.4
- **Lightness**: 0.4
- **Style**: Muted earth tones, weathered look

## API Reference

### Generator

```go
type Generator struct { ... }
```

Creates color palettes based on genre and seed.

#### NewGenerator

```go
func NewGenerator() *Generator
```

Creates a new palette generator with the default genre registry.

#### Generate

```go
func (g *Generator) Generate(genreID string, seed int64) (*Palette, error)
```

Generates a color palette for the specified genre and seed. Returns an error if the genre ID is not recognized.

### Palette

```go
type Palette struct {
    Primary    color.Color
    Secondary  color.Color
    Background color.Color
    Text       color.Color
    Accent1    color.Color
    Accent2    color.Color
    Danger     color.Color
    Success    color.Color
    Colors     []color.Color // 8 additional colors
}
```

Represents a cohesive color scheme for visual theming.

## Testing

```bash
# Run tests
go test -tags test ./pkg/rendering/palette/...

# Run tests with coverage
go test -tags test -cover ./pkg/rendering/palette/...
```

Current test coverage: **98.4%**

## Implementation Details

The palette generator uses the HSL (Hue, Saturation, Lightness) color space for more intuitive color manipulation:

1. **Genre Scheme**: Each genre has a base hue, saturation, and lightness with variation parameters
2. **Primary Color**: Generated from the base scheme
3. **Secondary Color**: Complementary hue (180° rotation) with variation
4. **Background**: Darker, desaturated version of the base
5. **Text**: High contrast with background (light or dark)
6. **Accents**: Triadic colors (120° and 240° rotations)
7. **Danger/Success**: Fixed red/green for UI consistency
8. **Additional Colors**: Distributed around the color wheel with random variation

The HSL to RGB conversion uses standard color space mathematics to ensure accurate color reproduction.
