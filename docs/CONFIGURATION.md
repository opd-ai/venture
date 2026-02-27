# Runtime Configuration Reference

This document provides a comprehensive reference for all runtime configuration options in Venture, including command-line flags and environment variables.

## Table of Contents

- [Server Flags](#server-flags)
- [Client Flags](#client-flags)
- [Environment Variables](#environment-variables)
- [Terrain Generation Types](#terrain-generation-types)
- [Genre Types](#genre-types)
- [Post-Processing Presets](#post-processing-presets)

## Server Flags

The following flags are available when running the dedicated server (`cmd/server/main.go`):

| Flag | Type | Default | Valid Range | Description |
|------|------|---------|-------------|-------------|
| `--port` | string | `"8080"` | Valid port number | Server port |
| `--max-players` | int | `8` | 1-100 | Maximum number of players |
| `--seed` | int64 | `12345` | Any int64 | World generation seed |
| `--genre` | string | `"fantasy"` | `fantasy`, `scifi`, `horror`, `cyberpunk`, `postapoc` | Genre ID for world generation |
| `--terrain-type` | string | `"bsp"` | `bsp`, `cellular`, `city`, `forest`, `composite`, `grammar`, `maze` | Terrain generator type (see [Terrain Generation Types](#terrain-generation-types)) |
| `--tick-rate` | int | `30` | 10-120 | Server update rate (updates per second) |
| `--verbose` | bool | `true` | true/false | Enable verbose logging |
| `--aerial-sprites` | bool | `true` | true/false | Enable aerial-view perspective sprites for top-down gameplay |
| `--high-latency` | bool | `false` | true/false | Use high-latency configuration optimized for Tor/onion services (200-5000ms latency) |
| `--security-audit` | bool | `true` | true/false | Run security audit at startup and log results |
| `--stability-monitor` | bool | `true` | true/false | Enable stability monitoring for production validation |
| `--simulate-network` | string | `""` | `low`, `medium`, `high`, `very-high`, `extreme` | Simulate network conditions for testing |
| `--resilience-metrics` | bool | `true` | true/false | Enable network resilience metrics collection |
| `--balance-validate` | bool | `false` | true/false | Run combat and economic balance validation at startup |
| `--migration-validate` | bool | `false` | true/false | Run save file migration validation at startup |
| `--ux-validate` | bool | `false` | true/false | Run user experience journey validation at startup |
| `--enable-mods` | bool | `true` | true/false | Enable mod system with sandbox security |
| `--mods-dir` | string | `"mods"` | Valid directory path | Directory to load mods from |
| `--metrics-port` | string | `"9090"` | Valid port number | Port for Prometheus metrics HTTP endpoint |
| `--enable-metrics` | bool | `true` | true/false | Enable Prometheus metrics export at /metrics endpoint |
| `--server-name` | string | `SERVER_NAME` env or `"venture-server"` | Any non-empty string | Server name for federation identity |
| `--version` | bool | `false` | true/false | Print version information and exit |

## Client Flags

The following flags are available when running the game client (`cmd/client/util.go`):

| Flag | Type | Default | Valid Range | Description |
|------|------|---------|-------------|-------------|
| `--width` | int | `1920` | 1280, 1920, 2560, 3840 | Screen width in pixels |
| `--height` | int | `1080` | 720, 1080, 1440, 2160 | Screen height in pixels |
| `--fullscreen` | bool | `false` | true/false | Start in fullscreen mode |
| `--seed` | int64 | Random | Any int64 | World generation seed |
| `--genre` | string | `"random"` | `fantasy`, `scifi`, `horror`, `cyberpunk`, `postapoc`, `random` | Genre ID |
| `--weather` | string | `""` | `rain`, `snow`, `fog`, `dust`, `ash`, `neonrain`, `smog`, `radiation` | Weather type (empty for genre-appropriate random) |
| `--weather-intensity` | string | `"heavy"` | `light`, `medium`, `heavy`, `extreme` | Weather intensity |
| `--postprocess-preset` | string | `"cinematic"` | `fantasy`, `sci-fi`, `horror`, `cyberpunk`, `post-apocalyptic`, `neutral`, `cinematic` | Post-processing preset |
| `--postprocess-color-grading` | bool | `true` | true/false | Enable color grading effect |
| `--postprocess-vignette` | bool | `true` | true/false | Enable vignette effect |
| `--postprocess-chromatic` | bool | `true` | true/false | Enable chromatic aberration effect |
| `--postprocess-saturation` | float64 | `1.1` | 0.0-2.0 | Color grading saturation |
| `--postprocess-contrast` | float64 | `1.05` | 0.0-2.0 | Color grading contrast |
| `--postprocess-brightness` | float64 | `0.02` | -1.0 to 1.0 | Color grading brightness |
| `--postprocess-vignette-intensity` | float64 | `0.6` | 0.0-1.0 | Vignette intensity |
| `--postprocess-vignette-softness` | float64 | `0.4` | 0.0-1.0 | Vignette softness |
| `--postprocess-chromatic-intensity` | float64 | `0.3` | 0.0-1.0 | Chromatic aberration intensity |
| `--palette-harmony` | string | `"triadic"` | `complementary`, `analogous`, `triadic`, `tetradic`, `split-complementary`, `monochromatic` | Color harmony type |
| `--palette-mood` | string | `"vibrant"` | See full list below | Palette mood |
| `--palette-rarity` | string | `"epic"` | `common`, `uncommon`, `rare`, `epic`, `legendary` | Palette rarity/intensity |
| `--verbose` | bool | `true` | true/false | Enable verbose logging |
| `--profile` | bool | `true` | true/false | Enable performance profiling with frame time tracking |
| `--multiplayer` | bool | `false` | true/false | Connect to remote multiplayer server |
| `--server` | string | `"localhost:8080"` | host:port | Server address for multiplayer |
| `--high-latency` | bool | `false` | true/false | Use high-latency configuration optimized for Tor/onion services |
| `--host-and-play` | bool | `false` | true/false | Explicitly enable host-and-play mode (default when --multiplayer not specified) |
| `--host-lan` | bool | `false` | true/false | Bind server to 0.0.0.0 for LAN access instead of 127.0.0.1 |
| `--port` | int | `8080` | Valid port number | Server port for --host-and-play mode |
| `--max-players` | int | `4` | 1-100 | Maximum players for --host-and-play mode |
| `--tick-rate` | int | `30` | 10-120 | Server tick rate for --host-and-play mode |
| `--no-tutorial` | bool | `false` | true/false | Disable tutorial for experienced players |
| `--vr` | bool | `false` | true/false | Enable VR mode (requires VR headset, auto-detects hardware) |
| `--force-vr` | bool | `false` | true/false | Force VR mode even without detected hardware (for testing) |
| `--version` | bool | `false` | true/false | Print version information and exit |

### Palette Mood Values

Complete list of `--palette-mood` options:
- `normal`, `bright`, `dark`, `saturated`, `muted`, `vibrant`, `pastel`
- `tense`, `calm`, `victorious`, `melancholic`, `energetic`, `mystical`, `ominous`, `serene`
- `aggressive`, `playful`, `somber`, `ethereal`, `dangerous`, `peaceful`, `chaotic`, `regal`, `desolate`

## Environment Variables

Environment variables take precedence over default flag values.

| Variable | Values | Description | Overrides Flag |
|----------|--------|-------------|----------------|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error`, `fatal` | Logging verbosity (unknown values default to `info`). Takes precedence over `--verbose` flag. When `LOG_LEVEL` is not set, `--verbose=true` sets level to `debug`, `--verbose=false` sets level to `info`. | `--verbose` |
| `LOG_FORMAT` | `json`, `text` | Log output format | N/A |
| `SERVER_NAME` | Any non-empty string | Server name for federation identity. Used to identify the server in cross-server guild federation and generates a unique ed25519 keypair for secure server authentication. | `--server-name` |

## Terrain Generation Types

The `--terrain-type` flag controls which procedural algorithm generates the world layout:

| Type | Description | Best For |
|------|-------------|----------|
| `bsp` | Binary Space Partitioning — Creates dungeon-style layouts with rectangular rooms connected by corridors | Traditional dungeon crawling, fantasy genres |
| `cellular` | Cellular Automata — Generates organic cave systems with natural-looking formations | Cave exploration, underground environments |
| `city` | City Generator — Creates urban layouts with roads, buildings, and districts | Urban exploration, cyberpunk/modern settings |
| `forest` | L-System Trees — Generates forest layouts with procedural tree placement | Natural outdoor environments, wilderness exploration |
| `composite` | Multi-Biome Composite — Combines multiple terrain types using Voronoi regions | Large varied worlds with diverse environments |
| `grammar` | Graph Grammar — Uses L-systems to create dungeons with narrative flow and structured room connections | Story-driven dungeons with meaningful layouts and progression |
| `maze` | Recursive Backtracking Maze — Creates complex winding corridors with optional rooms at dead ends | Puzzle-focused dungeons, labyrinth exploration |

### Usage Examples

```bash
# Generate a story-driven dungeon with graph grammar
./venture-server -terrain-type grammar -genre fantasy -seed 42

# Generate an organic cave system
./venture-server -terrain-type cellular -genre horror -seed 99

# Generate a cyberpunk city
./venture-server -terrain-type city -genre cyberpunk -seed 2077
```

## Genre Types

Available genre values for `--genre` flag:

| Genre | Description |
|-------|-------------|
| `fantasy` | Traditional fantasy setting with magic and medieval themes |
| `scifi` | Science fiction setting with advanced technology |
| `horror` | Horror-themed setting with dark atmosphere |
| `cyberpunk` | Cyberpunk setting with neon aesthetics and technology |
| `postapoc` | Post-apocalyptic setting with survival themes |
| `random` | Randomly select one of the above genres (client only) |

## Post-Processing Presets

The `--postprocess-preset` flag selects a pre-configured set of visual effects:

| Preset | Description |
|--------|-------------|
| `fantasy` | Warm colors, soft lighting, slight saturation boost |
| `sci-fi` | Cool colors, high contrast, chromatic aberration |
| `horror` | Desaturated, dark vignette, heavy grain |
| `cyberpunk` | Neon colors, heavy chromatic aberration, high saturation |
| `post-apocalyptic` | Desaturated browns/greys, dust effects, vignette |
| `neutral` | Minimal processing, close to raw rendering |
| `cinematic` | Balanced effects for visual polish (default) |

## Configuration Best Practices

### Production Server

```bash
# Recommended production server configuration
./venture-server \
  --port 8080 \
  --max-players 32 \
  --seed $(date +%s) \
  --genre fantasy \
  --terrain-type composite \
  --tick-rate 30 \
  --security-audit true \
  --stability-monitor true \
  --enable-metrics true \
  --metrics-port 9090
```

### High-Latency Server (Tor/Onion Services)

```bash
# Configuration for Tor hidden service deployment
./venture-server \
  --port 8080 \
  --high-latency true \
  --tick-rate 10 \
  --enable-metrics false
```

### Local Development

```bash
# Client with host-and-play for local testing
./client \
  --width 1920 \
  --height 1080 \
  --genre fantasy \
  --seed 12345 \
  --host-and-play true \
  --max-players 2 \
  --verbose true
```

### Performance Testing

```bash
# Server with network simulation for testing
./venture-server \
  --simulate-network high \
  --resilience-metrics true \
  --balance-validate true
```

## Compatibility Notes

- Screen resolutions (`--width`, `--height`): Use standard 16:9 ratios for best results (1920x1080, 2560x1440, 3840x2160)
- Tick rate (`--tick-rate`): Higher values increase CPU usage; 30 is recommended for normal latency, 10 for high-latency
- Post-processing effects: Disable for better performance on low-end hardware
- VR mode (`--vr`): Requires compatible VR headset; performance impact is significant
