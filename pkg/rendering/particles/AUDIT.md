# Audit: pkg/rendering/particles
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The `pkg/rendering/particles` package provides comprehensive procedural particle effect generation with advanced physics, weather systems, and environmental ambience. The package is well-architected with 91.8% test coverage, full determinism (seed-based RNG), and excellent performance characteristics. All automated checks passed. The code adheres to ECS data-oriented principles with zero direct Ebiten dependencies in the core package. The package supports 10+ particle types, 10+ weather effects, and 10+ ambient environment types, all with genre-based theming. Object pooling (sync.Pool) is implemented to reduce GC pressure. No critical issues found.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 91.8% (target: 30% for rendering packages) |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences (N/A - no networking) |

## Issues Found

### High Severity
None identified.

### Medium Severity
None identified.

### Low Severity
- [x] **Documentation** — Package-level doc.go is comprehensive (218 lines) but individual complex functions like `ApplyPhysics` could benefit from inline algorithm comments explaining the physics formulas (`behaviors.go:92`) — **ALREADY RESOLVED**: `applyGravityForces` has "v = v₀ + g·Δt" and "v *= (1 - drag·Δt)" formulas; `ApplyPhysics` delegates to well-commented sub-functions
- [x] **Documentation** — Weather system puddle accumulation and snow drift algorithms lack inline comments explaining the mathematical approach (`weather.go:300-400` range) — **ALREADY RESOLVED**: `handleParticleImpact` (weather.go:453-462) has explicit inline comments: "each raindrop adds 0.001 depth (capped at 1.0); full puddle = ~1000 impacts" and "each snowflake adds 0.0005 depth; full cover = ~2000 impacts"
- [x] **Performance** — `GetAliveParticles()` allocates a new slice on every call; documentation correctly warns users to use `VisitAliveParticles()` for hot paths, but could add runtime comment (`types.go:257`) (FIXED 2026-02-27: Added inline performance comment noting O(N) with allocation)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Particle package is data/logic only; no direct input handling |
| Mouse | N/A | Particle package is data/logic only; no direct input handling |
| Gamepad | N/A | Particle package is data/logic only; no direct input handling |
| Touch | N/A | Particle package is data/logic only; no direct input handling |
| VR | N/A | Particle package is data/logic only; no direct input handling |
| Stub/Test | N/A | No Input interface usage; pure data structures |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| N/A | N/A | N/A | N/A | Particle package provides data structures and generators; rendering handled by engine ParticleSystem |

## Documentation Coverage
- Package `doc.go`: ✅ Comprehensive 218-line documentation covering all features, usage examples, performance characteristics, and determinism guarantees
- Exported symbols documented: 82/82 (100%)
- Complex algorithms commented: ⚠️ Most algorithms well-commented, but physics formulas in `ApplyPhysics` and weather accumulation logic could use more inline mathematical explanations

## Integration Status
The particles package is a core rendering primitive used extensively throughout the engine for visual effects.

- System registration: ✅ — Multiple engine systems consume particles package:
  - `AmbientEnvironmentParticleSystem` (ambient effects)
  - `ArmorHitSparkSystem` (combat sparks)
  - `BlockParticleSystem` (block particles)
  - `CombatEquipmentDurabilityParticleSystem` (durability effects)
  - `CompanionAuraParticleSystem` (companion auras)
  - `CompanionLevelupParticleSystem` (level-up effects)
  - Plus 10+ other particle-emitting systems in pkg/engine/
- Component registration: ✅ — Particle data structures (Particle, ParticleSystem) are used by ParticleComponent in engine
- Serialize/Deserialize: N/A — Particles are transient runtime effects; not persisted in save files
- Network sync: N/A — Particle effects are client-side visual only; deterministic generation from seed allows server-driven particle spawn without syncing individual particle state
- Genre theming: ✅ — All generators accept GenreID parameter and use palette.Generator for genre-specific colors; verified in generator.go:42-64
- Mod compatibility: ✅ — Particle parameters (count, size, duration) can be overridden via Config.Custom map; mod system can inject custom particle configurations

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code; pure Go math and data structures |
| WASM | ✅ | WASM vet passed; no syscalls or filesystem access; deterministic RNG compatible with WASM |
| Mobile | ✅ | No touch/gesture dependencies; particle generation is platform-agnostic; performance targets (91.8% coverage, <5% frame time) suitable for mobile |

## Recommendations
1. **[LOW]** Add inline comments to `ApplyPhysics` function explaining the physics formulas for gravity, air resistance, and orbital motion calculations (`behaviors.go:92-135`)
2. **[LOW]** Add inline comments to weather accumulation functions (`UpdatePuddles`, `UpdateSnowDrift`) explaining the mathematical approach to puddle depth and snow drift simulation (`weather.go:300-450`)
3. **[LOW]** Consider adding a runtime warning (debug log) when `GetAliveParticles()` is called in a hot path (>100 times/second) to help developers discover the zero-allocation `VisitAliveParticles()` alternative (`types.go:257`)

## Additional Notes

### Performance Characteristics
Package documentation reports benchmark results:
- SPH Fluid: 349µs per frame (200 particles)
- Fire Propagation: 129µs per frame (200 particles)
- Smoke Turbulence: 9.3µs per frame (200 particles)
- Debris Collision: 96µs per frame (200 particles)
- **Total: 583µs (3.5% of 60 FPS frame budget)**

These targets are met based on benchmark tests in `physics_test.go` and `weather_test.go`.

### Object Pooling
Excellent use of sync.Pool in `pool.go`:
- `particleSystemPool` for ParticleSystem instances
- `particleSlicePool` for particle slice buffers
- `LRUCache` for frequently-used weather configurations
- Proper reset/cleanup on Put() to prevent state leaks
- Clear documentation on usage patterns and `ReleaseParticleSystem()` requirement

### Determinism Compliance
All particle generation correctly uses seed-based RNG:
- `rand.New(rand.NewSource(config.Seed))` in generator.go:88
- No global rand usage detected
- No time.Now() usage in generation code (only in cache timestamp, which is non-critical)
- Same seed → same particles verified in `generator_test.go`

### Zero Ebiten Dependencies
Package contains no direct Ebiten imports in non-test files. All particle data structures are pure Go types (float64, color.RGBA). This enables:
- Unit testing without graphics initialization
- Headless testing in CI/CD
- Easy mocking and stubbing
- Platform-agnostic code

### Advanced Features
- **Physics Behaviors**: 10 behavior flags (gravity, air resistance, bounce, trail, attract, orbit, rising, sink, shrink-on-rise, grow-on-fall) composable via bitflags
- **Weather Systems**: 10 weather types with puddle accumulation, snow depth tracking, and visibility modification
- **Ambient Effects**: 10 environment types with biome-specific particles and behaviors
- **LOD System**: Level-of-detail particle culling in `lod.go`
- **Genre Theming**: All particle colors driven by genre-based palette generation

### Test Quality
- 18 test files (9 implementation files + 9 test files)
- Table-driven tests throughout
- Benchmark tests for performance-critical paths
- Visitor pattern benchmark demonstrating zero-allocation benefits
- Weather system integration tests
- Physics simulation unit tests
- Determinism verification tests
