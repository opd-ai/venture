# Package sfx

Procedural sound effect generation for the Venture game engine.

## Overview

The `sfx` package provides runtime generation of game sound effects using waveform synthesis and audio processing. It creates various sound effects (impacts, explosions, magic, etc.) without requiring external audio files, enabling infinite variety within defined parameters.

## Features

- **9 Effect Types**: impact, explosion, magic, laser, pickup, hit, jump, death, powerup
- **Genre Variations**: Genre-specific sound modifications (fantasy, scifi, horror, cyberpunk, postapoc)
- **Sound Variety**: Automatic pitch and volume variations to avoid repetitive audio
- **Audio Processing**: Pitch shifting, vibrato, filtering (low-pass, high-pass)
- **Caching**: VarietyManager caches variants for performance (8.875 ns/op average)
- **Thread-Safe**: Concurrent access supported via sync.RWMutex
- **Deterministic**: Same seed always produces same output

## File Structure

```
pkg/audio/sfx/
├── doc.go              - Package documentation and usage examples
├── types.go            - Type definitions and constants
├── generator.go        - Main Generator struct and core API
├── effects.go          - Specific effect generation methods
├── processing.go       - Audio processing (pitch, vibrato, mixing)
├── variety.go          - Variation and filtering methods
├── variety_manager.go  - Caching manager for sound variants
└── helpers.go          - Utility functions (pitch calculations)
```

## Usage

### Basic Sound Generation

```go
import "github.com/opd-ai/venture/pkg/audio/sfx"

// Create generator
gen := sfx.NewGenerator(44100, seed)

// Generate impact sound
sample := gen.Generate("impact", seed)

// Generate with genre
sample = gen.GenerateWithGenre("explosion", seed, "scifi")
```

### Sound Variety (Recommended)

```go
// Create variety manager for natural-sounding repeated sounds
vm := sfx.NewVarietyManager(44100, seed)
vm.SetVariantsPerEffect(5)      // Cache 5 variants per effect
vm.SetPitchVariance(2.0)         // ±2 semitones
vm.SetVolumeVariance(0.2)        // ±20% volume

// Generate with automatic variety
sample := vm.Generate("impact", seed)
```

### Advanced Variations

```go
// Generate with specific pitch shift
sample := gen.GenerateWithPitchShift("magic", seed, 7.0) // +7 semitones

// Generate with specific volume
sample = gen.GenerateWithVolume("explosion", seed, 0.5) // 50% volume

// Apply filters
gen.ApplyLowPassFilter(sample, 0.3)  // Muffled sound
gen.ApplyHighPassFilter(sample, 0.3) // Tinny sound
```

## Effect Types

| Type | Description | Duration | Characteristics |
|------|-------------|----------|-----------------|
| `impact` | Short punch/thud | 0.1-0.2s | Noise, sharp attack, pitch bend down |
| `explosion` | Large boom | 0.5-0.8s | Noise + low rumble, longer envelope |
| `magic` | Magical sparkle | 0.3-0.5s | Sine waves, harmonics, vibrato |
| `laser` | Sci-fi beam | 0.2-0.3s | Square wave, pitch sweep down |
| `pickup` | Item collect | 0.15s | Triangle wave, pitch sweep up |
| `hit` | Damage/attack | 0.1s | Square wave, short duration |
| `jump` | Character jump | 0.2s | Square wave, pitch sweep up |
| `death` | Defeat sound | 0.8s | Sawtooth, pitch sweep down |
| `powerup` | Power-up | 0.4s | Sine, ascending arpeggio |

## Genre Modifications

Different genres apply characteristic sonic changes:

- **Fantasy** (default): No modifications
- **Sci-fi**: Higher pitch (+30%), cleaner/synthetic sound
- **Horror**: Lower pitch (-30%), vibrato for unsettling effect
- **Cyberpunk**: Sharp pitch (+40%), hard clipping for digital effect
- **Post-apocalyptic**: Slightly lower pitch (-10%), soft clipping for gritty effect

## Architecture

### Generator
Core type that creates sound effects. Holds oscillator and RNG state.

**Key Methods:**
- `Generate(effectType, seed)` - Generate sound effect
- `GenerateWithGenre(effectType, seed, genre)` - Generate with genre modifications
- `GenerateVariant(effectType, seed, pitchVar, volVar)` - Generate with variations
- `GenerateWithPitchShift(effectType, seed, semitones)` - Generate with specific pitch
- `GenerateWithVolume(effectType, seed, volume)` - Generate with specific volume
- `ApplyLowPassFilter(sample, cutoff)` - Apply low-pass filter
- `ApplyHighPassFilter(sample, cutoff)` - Apply high-pass filter

### VarietyManager
Manages cached variants of sounds for performance and variety.

**Key Methods:**
- `Generate(effectType, seed)` - Get cached variant or generate new ones
- `SetVariantsPerEffect(count)` - Configure cache size (default: 5)
- `SetPitchVariance(variance)` - Configure pitch range (default: 2.0 semitones)
- `SetVolumeVariance(variance)` - Configure volume range (default: 0.2 = ±20%)
- `ClearCache()` - Clear all cached variants
- `GetCacheSize()` - Get total cached variant count

## Performance

- **Cold generation**: Variable based on effect complexity (0.1-0.8s duration)
- **Cached retrieval**: 8.875 ns/op average (VarietyManager)
- **Memory**: ~5-10KB per cached variant
- **Recommended cache size**: 5 variants per effect type

## Testing

Package has 89.9% test coverage with 23 tests covering:
- All 9 effect types
- Genre variations
- Pitch shifting and volume modification
- Filter applications
- VarietyManager caching and determinism
- Edge cases (unknown effects, zero values, extreme parameters)

Run tests:
```bash
go test ./pkg/audio/sfx/...
go test -cover ./pkg/audio/sfx/...
go test -bench=. ./pkg/audio/sfx/...
```

## Dependencies

- `github.com/opd-ai/venture/pkg/audio` - AudioSample type
- `github.com/opd-ai/venture/pkg/audio/synthesis` - Oscillator, Envelope
- `github.com/sirupsen/logrus` - Structured logging
- Standard library: `math`, `math/rand`, `sync`

## Thread Safety

- **Generator**: Not thread-safe. Create separate instances per goroutine.
- **VarietyManager**: Thread-safe with sync.RWMutex for concurrent access.

## Design Decisions

### Reorganization
Code was reorganized from 3 files (generator.go, variety.go, variety_manager.go) into 8 files for better navigability:
- **types.go** - Centralized type definitions
- **effects.go** - Separated effect generation methods
- **processing.go** - Separated audio processing methods
- **helpers.go** - Extracted utility functions

### Private Methods
Most effect generation and processing methods are private (lowercase). This encapsulates implementation details and provides a clean public API through the Generator type.

### Caching Strategy
VarietyManager caches generated variants rather than generating on-demand to:
1. Improve performance for repeated sounds
2. Ensure consistent variety within a seed
3. Reduce CPU load during gameplay

## Future Enhancements

See AUDIT.md for optional enhancement ideas, including:
- Additional effect types (footsteps, ambient, UI sounds)
- More genre-specific modifications
- Advanced filters (bandpass, notch)
- Spatial effects (reverb, delay)
- Preset configuration system

## License

Part of the Venture game engine. See repository LICENSE for details.
