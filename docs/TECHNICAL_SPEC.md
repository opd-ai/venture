# Technical Specification

**Version:** 3.0  
**Last Updated:** November 2025

Technical architecture and implementation details for Venture.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Technology Stack](#technology-stack)
3. [Core Systems](#core-systems)
4. [Data Structures](#data-structures)
5. [Performance](#performance)

---

## Architecture Overview

### Design Principles

**Entity-Component-System (ECS):** Separation of data (components) and logic (systems)  
**Deterministic Generation:** Seed-based procedural content (reproducible, testable, multiplayer-safe)  
**Minimal Dependencies:** Go stdlib + Ebiten + logrus/zenity  
**Cross-Platform:** Single codebase for desktop/web/mobile

### Package Structure

```
pkg/
├── engine/          # ECS framework, core systems
├── procgen/         # Procedural generation (terrain, entities, items, etc.)
├── rendering/       # Visual generation (sprites, tiles, particles, UI)
│   ├── sprites/     # Enhanced sprite generation (V3.0: anatomical templates, anti-aliasing)
│   ├── tiles/       # Advanced tile rendering (V3.0: texture patterns, transitions)
│   ├── lighting/    # Sophisticated lighting (V3.0: soft shadows, bloom, colored lights)
│   ├── particles/   # Rich particle systems (V3.0: weather, fluid simulation)
│   ├── ui/          # UI rendering (V3.0: dynamic colors, visual hierarchy)
│   ├── palette/     # Color palette generation
│   ├── patterns/    # Texture pattern generation (V3.0)
│   ├── cache/       # Sprite caching (95.9% hit rate)
│   ├── pool/        # Object pooling
│   ├── postprocess/ # Post-processing effects (V3.0: parallax, time-of-day)
│   ├── quality/     # Rendering quality settings (V3.0)
│   └── shapes/      # Procedural shape generation
├── audio/           # Sound synthesis (waveforms, music, SFX)
├── network/         # Multiplayer (client-server, prediction, lag compensation)
├── combat/          # Combat mechanics
├── world/           # World state management
├── saveload/        # Persistence
├── logging/         # Structured logging
├── version/         # Version management (V3.0)
├── mobile/          # Mobile-specific code (touch input)
└── visualtest/      # Visual testing utilities
```

---

## Technology Stack

**Language:** Go 1.24.5+  
**Game Engine:** Ebiten 2.9.3 (2D, cross-platform)  
**Logging:** logrus 1.9.3 (structured JSON)  
**Dialogs:** zenity 0.10.14 (native dialogs)  
**Image:** golang.org/x/image 0.32.0 (extended image processing)

**Build:** Single binary, cross-compilation for Linux/macOS/Windows/WASM/iOS/Android

---

## Core Systems

### ECS Framework

**World:** Central container for entities and systems  
**Entity:** ID + component collection  
**Component:** Pure data structures implementing `Type() string`  
**System:** Logic processors implementing `Update(entities, deltaTime)`

**Query Optimization:** Quadtree spatial partitioning for efficient entity queries

### Procedural Generation

**Generators:** Implement `Generate(seed, params) (result, error)` and `Validate(result) error`  
**Seed Derivation:** `SeedGenerator` creates deterministic sub-seeds  
**Templates:** Genre-specific templates for entities/items/magic  
**Rarity System:** Common (1.0x), Uncommon (1.2x), Rare (1.5x), Epic (2.0x), Legendary (3.0x)

### Rendering Pipeline

**V3.0 Enhanced Pipeline:**

1. **Sprite Generation** (V3.0 Enhanced)
   - Anatomical templates with pixel-perfect dimensions (head 4×4, torso 4×6, legs 4×8)
   - Facial features for close-up views (eyes 2×1, mouth 2×1)
   - Anti-aliasing with 4 quality levels (Off, Low, Medium, High)
   - Sub-pixel rendering using 2x2 to 8x8 super-sampling
   - Genre variations (organic, geometric, distorted, augmented, weathered)
   - 40% more anatomical detail than V2.0

2. **Tile Rendering** (V3.0 Enhanced)
   - Procedural texture patterns: stone, wood, metal, organic
   - 50+ unique patterns per genre via seed-based generation
   - Smooth transitions with automated edge detection
   - Multi-layer depth effects for visual richness
   - Detail layers for texture complexity
   - Normal mapping simulation for depth perception

3. **Lighting System** (V3.0 Enhanced)
   - Soft shadows with gradient edges
   - Colored lighting matching light sources
   - Bloom effects for magical/technological lights
   - Advanced ambient occlusion for depth
   - Dynamic torches with flickering animation
   - Genre-specific lighting presets
   - <5% frame time overhead

4. **Weather & Particles** (V3.0 New)
   - Comprehensive weather: rain, snow, fog, dust, ash
   - Genre-specific weather (neon rain, radiation zones, smog)
   - Fluid simulation for realistic particle behavior
   - Intensity levels: light, medium, heavy, extreme
   - Environmental interactions (wind, gravity effects)

5. **Post-Processing** (V3.0 New)
   - Parallax backgrounds with multi-layer depth
   - Time-of-day system with dynamic lighting shifts
   - Screen-space enhancements
   - Genre-specific visual filters

6. **Rendering Optimizations** (V2.0, maintained in V3.0)
   - Viewport culling (1,635x speedup)
   - Batch rendering by layer/texture (1,667x speedup)
   - Sprite cache lookup (95.9% hit rate, 37x speedup)
   - Object pooling (2x speedup)
   - Draw to screen

**Color Palettes:** Genre-specific color schemes  
**Sprites:** Enhanced silhouette-based generation with caching  
**Particles:** Pooled particle system with weather support

### Audio System

**Synthesis:** Waveform generation (sine, square, triangle, sawtooth)  
**Music:** Procedural composition with genre themes  
**SFX:** Event-triggered sound generation

### Networking

**Architecture:** Authoritative server, client-side prediction  
**Lag Compensation:** Snapshot history, server rewind for hit detection  
**Interpolation:** Buffer 100-200ms, interpolate between snapshots  
**Latency Support:** 200-5000ms (including Tor)

**Protocol:** Binary message format over TCP  
**Sync:** Delta compression, spatial culling, priority system

---

## Data Structures

### Component Examples

```go
type PositionComponent struct {
    X, Y float64
}

type VelocityComponent struct {
    VX, VY float64
}

type HealthComponent struct {
    Current, Max float64
}

type StatsComponent struct {
    Attack, Defense, MagicPower, MagicDefense float64
    CritChance, CritDamage, Evasion float64
    Resistances map[DamageType]float64
}
```

### Terrain Structure

```go
type Terrain struct {
    Width, Height int
    Tiles         [][]TileType
    Rooms         []*Room
    Seed          int64
}

type Room struct {
    X, Y, Width, Height int
    Type                RoomType  // Start, Boss, Treasure, etc.
    Connections         []*Room
}
```

### Entity Structure

```go
type Entity struct {
    Type      EntityType  // Monster, NPC, Boss
    Stats     Stats       // HP, Attack, Defense, etc.
    Behavior  Behavior    // AI behavior pattern
    Inventory []Item
}
```

---

## Performance

### Targets vs Achieved (V3.0)

| Metric | Target | V2.0 Achieved | V3.0 Achieved |
|--------|--------|---------------|---------------|
| Frame Rate | 60 FPS | 106 FPS ✅ | 106 FPS ✅ |
| Frame Time | <16.67ms | 0.02ms ✅ | 0.02ms ✅ |
| Memory | <500 MB | 73 MB ✅ | 73 MB ✅ |
| Generation | <2s | <2s ✅ | <2s ✅ |
| Bandwidth | <100 KB/s | <100 KB/s ✅ | <100 KB/s ✅ |
| Sprite Gen | <5ms | 2-3ms ✅ | 3-5ms ✅ |

**V3.0 Visual Quality Improvements:**
- **Sprite Detail:** +40% anatomical detail
- **Silhouette Scores:** 0.75 average (up from 0.65)
- **Texture Variety:** 50+ patterns per genre (new)
- **Animation Smoothness:** 12 FPS close range (new)
- **Lighting Overhead:** <5% frame time increase
- **Cache Hit Rate:** 95.9% maintained

### Optimization Stack

1. **Viewport Culling:** Only process/render visible entities
2. **Batch Rendering:** Group draw calls
3. **Sprite Caching:** Reuse generated sprites (95.9% hit rate)
4. **Object Pooling:** Reuse allocated objects
5. **Spatial Partitioning:** Quadtree for collision detection

**Combined Improvement:** 1,625x performance gain

### Memory Management

**GC Tuning:** `GOGC` environment variable (default 100)  
**Allocation Reduction:** Preallocate slices, reuse buffers  
**Pooling:** Particles, temporary objects

---

## Determinism

**Critical for:**
- Multiplayer synchronization (same seed = same world)
- Testing (reproducible bugs)
- Debugging (replay scenarios)

**Implementation:**
- Seed-based RNG: `rand.New(rand.NewSource(seed))`
- No `time.Now()` in generation
- No global `math/rand`
- Deterministic algorithms (BSP, cellular automata)

**Validation:** Generate twice with same seed, compare byte-for-byte

---

## Testing Strategy

**Coverage:** 82.4% average (≥65% target)  
**Methodology:** Table-driven tests, stub implementations  
**Benchmarks:** Performance regression detection  
**Race Detection:** `-race` flag for concurrency issues

**Excluded:** Ebiten initialization code (requires display)

---

## Build System

**Makefile Targets:**
- `make build` - Client + server
- `make build-all` - All platforms
- `make build-wasm` - WebAssembly
- `make android-apk` / `make ios-ipa` - Mobile
- `make test` - Run tests
- `make profile-cpu` / `make profile-mem` - Profiling

**CI/CD:** GitHub Actions for automated testing/building/deployment

---

## Additional Resources

- [Architecture](ARCHITECTURE.md) - High-level design
- [Development](DEVELOPMENT.md) - Developer workflow
- [Performance](PERFORMANCE.md) - Optimization guide
- [API Reference](API_REFERENCE.md) - API documentation

---

**Version:** 3.0  
**Last Updated:** November 2025  
**Maintained By:** Venture Development Team
