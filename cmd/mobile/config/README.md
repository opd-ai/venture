# Mobile Seed Configuration

This document describes how to configure world seed and genre for mobile builds of Venture.

## Overview

Mobile builds support environment variable configuration for reproducible world generation, which is essential for:
- Testing and debugging specific world seeds
- Reproducing bug reports from mobile testers
- Comparing world generation between mobile and desktop platforms
- Automated testing with consistent world states

## Environment Variables

### VENTURE_SEED

Controls the world generation seed for deterministic procedural generation.

**Format:** Integer (int64)  
**Default:** Random time-based seed  

⚠️ **Determinism Exception:** The time-based fallback is an INTENTIONAL exception to 
the project's deterministic generation principle (Coding Guideline #2). This provides 
casual mobile players with a unique world each launch. For reproducible worlds 
(bug reports, testing, multiplayer coordination), always set VENTURE_SEED explicitly.

**Examples:**
```bash
# Set a specific seed for testing
export VENTURE_SEED=12345

# Use a negative seed
export VENTURE_SEED=-9876

# Maximum int64 value
export VENTURE_SEED=9223372036854775807
```

**Validation:**
- Must be a valid int64 integer
- Invalid values fall back to time-based seed with a warning

### VENTURE_GENRE

Controls the genre theme for procedural content generation.

**Format:** String (one of: fantasy, scifi, horror, cyberpunk, postapoc)  
**Default:** Random genre selection  
**Examples:**
```bash
# Fantasy theme
export VENTURE_GENRE=fantasy

# Science fiction theme
export VENTURE_GENRE=scifi

# Horror theme
export VENTURE_GENRE=horror

# Cyberpunk theme
export VENTURE_GENRE=cyberpunk

# Post-apocalyptic theme
export VENTURE_GENRE=postapoc
```

**Validation:**
- Must exactly match one of the allowed genres (case-sensitive)
- Invalid values fall back to random genre selection with a warning

## Usage Examples

### Android

For Android builds with ebitenmobile, set environment variables before running the app:

```bash
# Build with specific seed and genre
VENTURE_SEED=12345 VENTURE_GENRE=fantasy \
  ebitenmobile bind -target android -o mobile.aar ./cmd/mobile
```

### iOS

For iOS builds with ebitenmobile:

```bash
# Build with specific seed and genre
VENTURE_SEED=67890 VENTURE_GENRE=scifi \
  ebitenmobile bind -target ios -o Mobile.xcframework ./cmd/mobile
```

### Testing

For local testing during development:

```bash
# Test with specific configuration
VENTURE_SEED=12345 VENTURE_GENRE=fantasy go test ./cmd/mobile/...
```

## Determinism

When using `VENTURE_SEED`, the same seed will always generate:
- Identical terrain layouts
- Same enemy spawns and placements
- Identical item generation
- Same NPC merchants and dialog

This ensures reproducible gameplay for testing and debugging.

## Comparison with Desktop Client

Desktop client seed configuration via command-line flags:

```bash
# Desktop equivalent
./venture-client -seed 12345 -genre fantasy
```

Mobile environment variables provide equivalent functionality:

```bash
# Mobile equivalent
VENTURE_SEED=12345 VENTURE_GENRE=fantasy [run mobile build]
```

## Logging

The mobile client logs seed and genre sources on startup:

**Environment variable seed:**
```
INFO using seed from environment variable seed=12345 source=VENTURE_SEED
INFO using genre from environment variable genre=fantasy source=VENTURE_GENRE
```

**Time-based seed (default):**
```
INFO using time-based seed seed=1704067200000000000 source=time-based
INFO using random genre genre=scifi source=random
```

**Invalid configuration:**
```
WARN invalid VENTURE_SEED environment variable, using time-based seed seedStr=not-a-number
WARN invalid VENTURE_GENRE environment variable, using random genre genre=Unknown valid=[fantasy scifi horror cyberpunk postapoc]
```

## Implementation Details

Seed configuration is handled by the `cmd/mobile/config` package:

- `GetSeedFromEnv(logger)` - Retrieves seed from `VENTURE_SEED` or generates time-based seed
- `GetGenreFromEnv(genres, rng, logger)` - Retrieves genre from `VENTURE_GENRE` or selects random

Both functions support optional logger parameter for diagnostics.

## Testing

Test coverage: 73.9% (exceeds 40% minimum requirement)

Run tests:
```bash
go test -v ./cmd/mobile/config
```

Run benchmarks:
```bash
go test -bench=. ./cmd/mobile/config
```

## See Also

- [Mobile Build Documentation](../docs/BUILD_ANDROID.md)
- [Procedural Generation Overview](../docs/PROCGEN.md)
- [Project Overview](../README.md#mobile-support)
