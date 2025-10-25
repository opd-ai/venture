# Palette Generator

The palette package provides procedural color palette generation for genre-based theming in the Venture game. **Phase 4** enhanced the system with color harmony rules, mood variations, and rarity-based schemes.

## Features

- **Genre-Aware**: Generates palettes that match the visual style of each game genre (Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic)
- **Deterministic**: Same seed always produces the same palette
- **HSL Color Space**: Uses Hue-Saturation-Lightness for intuitive color generation
- **Comprehensive**: 16 named colors + configurable array (default 12+)
- **Color Harmony**: 6 harmony types based on color theory (Complementary, Analogous, Triadic, Tetradic, Split-Complementary, Monochromatic)
- **Mood Variations**: 7 emotional adjustments (Normal, Bright, Dark, Saturated, Muted, Vibrant, Pastel)
- **Rarity Tiers**: 5 intensity levels (Common, Uncommon, Rare, Epic, Legendary)
- **High Performance**: ~10-11μs per palette generation

## Usage

### Basic Usage

```go
import "github.com/opd-ai/venture/pkg/rendering/palette"

// Create a generator
gen := palette.NewGenerator()

// Generate a fantasy-themed palette (uses default options)
pal, err := gen.Generate("fantasy", 12345)
if err != nil {
    log.Fatal(err)
}

// Access named colors
fmt.Printf("Primary: %v\n", pal.Primary)
fmt.Printf("Secondary: %v\n", pal.Secondary)
fmt.Printf("Accent3: %v\n", pal.Accent3)
fmt.Printf("Highlight1: %v\n", pal.Highlight1)
```

### Advanced Usage with Options (Phase 4)

```go
// Configure harmony, mood, and rarity
opts := palette.GenerationOptions{
    Harmony:   palette.HarmonyTriadic,    // Triadic color relationships
    Mood:      palette.MoodVibrant,       // Vibrant emotional tone
    Rarity:    palette.RarityEpic,        // Epic-tier intensity
    MinColors: 16,                         // Generate 16+ colors
}

pal, err := gen.GenerateWithOptions("scifi", 54321, opts)
if err != nil {
    log.Fatal(err)
}

// Use the enhanced palette
fmt.Printf("Shadow1: %v\n", pal.Shadow1)
fmt.Printf("Warning: %v\n", pal.Warning)
fmt.Printf("Total colors: %d\n", len(pal.Colors))
```

### Default Options

```go
opts := palette.DefaultOptions()
// Harmony: Complementary
// Mood: Normal
// Rarity: Common
// MinColors: 12
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

### Types

#### Palette

```go
type Palette struct {
    // Core colors
    Primary    color.Color
    Secondary  color.Color
    Background color.Color
    Text       color.Color
    
    // Accent colors
    Accent1 color.Color
    Accent2 color.Color
    Accent3 color.Color  // NEW in Phase 4
    
    // Highlight colors (for emphasis)
    Highlight1 color.Color  // NEW in Phase 4
    Highlight2 color.Color  // NEW in Phase 4
    
    // Shadow colors (for depth)
    Shadow1 color.Color  // NEW in Phase 4
    Shadow2 color.Color  // NEW in Phase 4
    
    // Neutral color
    Neutral color.Color  // NEW in Phase 4
    
    // UI feedback colors
    Danger  color.Color
    Success color.Color
    Warning color.Color  // NEW in Phase 4
    Info    color.Color  // NEW in Phase 4
    
    // Additional colors (default 12+)
    Colors []color.Color
}
```

#### HarmonyType (NEW in Phase 4)

```go
type HarmonyType int

const (
    HarmonyComplementary      HarmonyType = iota  // Opposite hues (180°)
    HarmonyAnalogous                              // Adjacent hues (±30°)
    HarmonyTriadic                                // Three hues (120° apart)
    HarmonyTetradic                               // Four hues (90° apart)
    HarmonySplitComplementary                     // Base + two adjacent to complement
    HarmonyMonochromatic                          // Single hue variations
)
```

#### MoodType (NEW in Phase 4)

```go
type MoodType int

const (
    MoodNormal    MoodType = iota  // Standard values
    MoodBright                      // Cheerful, increased lightness
    MoodDark                        // Somber, decreased lightness
    MoodSaturated                   // Intense colors
    MoodMuted                       // Subdued colors
    MoodVibrant                     // Maximum saturation
    MoodPastel                      // High lightness, low saturation
)
```

#### Rarity (NEW in Phase 4)

```go
type Rarity int

const (
    RarityCommon    Rarity = iota  // Muted tones
    RarityUncommon                 // Slightly enhanced
    RarityRare                     // Vibrant colors
    RarityEpic                     // Intense with high contrast
    RarityLegendary                // Extraordinary colors
)
```

#### GenerationOptions (NEW in Phase 4)

```go
type GenerationOptions struct {
    Harmony   HarmonyType  // Color relationship type
    Mood      MoodType     // Emotional tone adjustment
    Rarity    Rarity       // Color intensity tier
    MinColors int          // Minimum colors to generate (default: 12)
}
```

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

Generates a color palette for the specified genre and seed using default options. Returns an error if the genre ID is not recognized.

#### GenerateWithOptions (NEW in Phase 4)

```go
func (g *Generator) GenerateWithOptions(genreID string, seed int64, opts GenerationOptions) (*Palette, error)
```

Generates a color palette with specific harmony, mood, and rarity settings. Allows full control over palette generation.

#### DefaultOptions (NEW in Phase 4)

```go
func DefaultOptions() GenerationOptions
```

Returns default generation options (Complementary harmony, Normal mood, Common rarity, 12 colors).

## Testing

```bash
# Run tests
go test -tags test ./pkg/rendering/palette/...

# Run tests with coverage
go test -tags test -cover ./pkg/rendering/palette/...

# Run benchmarks
go test -tags test -bench=. -benchmem ./pkg/rendering/palette/...
```

Current test coverage: **100%**

### Performance Benchmarks (Phase 4)

```
BenchmarkGenerateWithHarmony-16     101438    10571 ns/op    6000 B/op    33 allocs/op
BenchmarkGenerateWithMood-16        109117    10610 ns/op    5992 B/op    33 allocs/op
BenchmarkGenerateWithRarity-16      108553    10788 ns/op    5992 B/op    33 allocs/op
BenchmarkGenerate24Colors-16        100598    11544 ns/op    6248 B/op    45 allocs/op
```

Target: <5ms per palette (actual: ~10-11μs = **450x faster** than target)

## Implementation Details

The palette generator uses the HSL (Hue, Saturation, Lightness) color space for more intuitive color manipulation:

### Phase 4 Enhancements

1. **Color Harmony**: Generates harmonious hues based on color theory
   - Complementary: `hue2 = (hue1 + 180) mod 360`
   - Analogous: `[hue, hue+30, hue-30]`
   - Triadic: `[hue, hue+120, hue+240]`
   - Tetradic: `[hue, hue+90, hue+180, hue+270]`
   - Split-Complementary: `[hue, hue+150, hue+210]`

2. **Mood Adjustments**: HSL transformations
   - Bright: `L *= 1.3`, `S *= 1.1`
   - Dark: `L *= 0.6`, `S *= 0.8`
   - Saturated: `S *= 1.4`
   - Muted: `S *= 0.5`
   - Vibrant: `S *= 1.5`, `L *= 1.2`
   - Pastel: `L *= 1.5`, `S *= 0.4`

3. **Rarity Scaling**: Intensity multipliers
   - Common: `S *= 0.8`
   - Uncommon: `L *= 1.05`
   - Rare: `S *= 1.3`, `L *= 1.1`
   - Epic: `S *= 1.5`, `L *= 1.15`, `variation *= 1.3`
   - Legendary: `S *= 1.7`, `L *= 1.2`, `variation *= 1.5`

### Original Implementation (Phase 1-3)

1. **Genre Scheme**: Each genre has a base hue, saturation, and lightness with variation parameters
2. **Primary Color**: Generated from the base scheme
3. **Secondary Color**: Complementary hue (180° rotation) with variation
4. **Background**: Darker, desaturated version of the base
5. **Text**: High contrast with background (light or dark)
6. **Accents**: Triadic colors (120° and 240° rotations)
7. **Danger/Success**: Fixed red/green for UI consistency
8. **Additional Colors**: Distributed around the color wheel with random variation

The HSL to RGB conversion uses standard color space mathematics to ensure accurate color reproduction.

## Examples

See `examples/color_demo/` for a complete demonstration:
- All harmony types (6)
- All mood variations (7)
- All rarity tiers (5)
- Genre comparisons (5)
- Combined effects
- Performance characteristics

Run the demo:
```bash
go run examples/color_demo/main.go
```

## Version History

- **Phase 1-3**: Basic 8-color palette with genre support
- **Phase 4**: Enhanced to 12+ colors with harmony, mood, and rarity (October 2025)
- Current version: **1.3**
