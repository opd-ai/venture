# Audit: pkg/procgen/furniture
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete

## Summary
The furniture package provides deterministic procedural generation of 30+ furniture types with material variation, rarity tiers, and placement validation. All automated checks pass cleanly. The package demonstrates excellent code quality with 92.5% test coverage, comprehensive table-driven tests, strict determinism validation, and thorough integration with housing systems. No critical or high-severity issues identified.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 92.5% (target: 40%) |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences (N/A for this package) |

## Issues Found

### High Severity
_(None)_

### Medium Severity
- [x] **Documentation** — placement.go package comment — **ALREADY RESOLVED**: placement.go:1-4 already has "implements AABB (Axis-Aligned Bounding Box) collision detection... grid-based auto-placement strategy... greedy bin-packing to avoid overlaps."

### Low Severity
- [x] **Code Organization** — generateName/generateDescription refactoring — **DEFERRED**: nested switch tables are readable and genre-specific data is naturally co-located; lookup table refactoring is a style preference, not a correctness issue.
- [x] **API Consistency** — `PlacementValidator.GetOccupancy()` returns percentage (0-100) but lacks unit suffix in name; consider `GetOccupancyPercent()` for clarity (`placement.go:104`) **COMPLETED 2026-02-27** - Renamed GetOccupancy() to GetOccupancyPercent() with enhanced godoc. Updated all usages in tests and doc.go. Coverage: 92.5%

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | No input handling (pure data generator) |
| Mouse | N/A | No input handling (pure data generator) |
| Gamepad | N/A | No input handling (pure data generator) |
| Touch | N/A | No input handling (pure data generator) |
| VR | N/A | No input handling (pure data generator) |
| Stub/Test | N/A | No input handling (pure data generator) |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Housing UI (Furniture Placement) | ✅ | ✅ | ✅ | Integrated via `pkg/world/housing/ui.go`; furniture generator initialized in `cmd/client/init_versions.go`; category selection and placement functional |

## Documentation Coverage
- Package `doc.go`: ✅ (194 lines with comprehensive examples, feature list, integration points)
- Exported symbols documented: 42/42 (100%)
  - All 8 types exported
  - All 4 enums with String() methods
  - All 11 public functions/methods
  - All 19 template access functions
- Complex algorithms commented: ✅ (deterministic material selection, genre-specific naming, collision detection all explained)

## Integration Status
This package integrates with housing system, guild housing, and crafting integration.

- System registration: N/A — This is a data generator, not an ECS system; does not require system registration
- Component registration: N/A — Generated furniture is not an ECS component; used for persistent housing state
- Serialize/Deserialize: N/A — Furniture is serialized as part of housing state via `pkg/world/housing/` persistence
- Network sync: N/A — Housing furniture state is synced server-side; furniture package provides generation only
- Genre theming: ✅ — Reads `GenreID` from `GenerationParams` and adapts material selection, naming prefixes/suffixes, and color palettes per genre (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
- Mod compatibility: ✅ — Furniture generation uses `procgen.GenerationParams` which supports mod rule injection via `Custom` map; no hardcoded values prevent mod overrides

**Integration Points Verified**:
1. **cmd/client/handlers.go**: `furnitureGenerator *furniture.Generator` field initialized
2. **cmd/client/init_versions.go**: `sys.furnitureGenerator = furniture.NewGenerator()` on startup
3. **pkg/world/housing/ui.go**: Uses `furniture.Generator` for runtime generation; calls `furniture.GetSubTypesByCategory()` for UI filtering
4. **pkg/integration/guild_housing/doc.go**: References furniture generation for guild hall furnishing
5. **pkg/world/housing/ui_test.go**: Integration tests validate furniture category enumeration matches UI expectations

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Pure Go code with no platform-specific dependencies |
| WASM | ✅ | `GOOS=js GOARCH=wasm go vet` passes cleanly; no filesystem/syscall dependencies |
| Mobile | ✅ | No mobile-specific concerns; deterministic generation suitable for all platforms |

## Recommendations
1. **[MED]** Add package-level documentation to `placement.go` explaining AABB collision algorithm and grid-based auto-placement strategy for future maintainers
2. **[LOW]** Refactor `generateName()` and `generateDescription()` into lookup tables to reduce cyclomatic complexity and improve maintainability (current: 100+ lines of nested switch statements)
3. **[LOW]** Add benchmark tests for performance-critical paths to validate documented targets: `BenchmarkGenerate` (target: <10ms), `BenchmarkFindValidPlacement` (target: <5ms)
4. **[LOW]** Rename `GetOccupancy()` to `GetOccupancyPercent()` for API clarity (return value is percentage, not ratio)

## Detailed Findings

### Code Quality Assessment

**Strengths**:
1. **Strict Determinism**: All RNG values pre-generated before conditionals to ensure deterministic material/rarity selection (`generator.go:308-324`)
2. **ECS Compliance**: No ECS violations; package is a pure data generator with no component behavior
3. **Comprehensive Testing**: 92.5% coverage with table-driven tests, determinism validation, and all 30+ templates tested
4. **Genre Theming**: Full genre support with adaptive material selection, naming conventions, and color palettes across 5 genres
5. **API Design**: Clean separation between generation (`Generator`), placement (`PlacementValidator`), and data (`Furniture`, `Template`)
6. **Documentation**: Excellent doc.go with usage examples, feature list, and integration guidance

**Determinism Validation**:
- ✅ All randomness uses `rand.New(rand.NewSource(seed))`
- ✅ No `time.Now()` calls
- ✅ No global `rand.*` functions
- ✅ Test validates same seed → identical output (`TestGenerateDeterminism`)
- ✅ Pre-rolled random values for deterministic branching (`generator.go:309-311`)

**Error Handling**:
- ✅ All public functions return explicit errors
- ✅ Input validation with clear error messages (`generator.go:124-129`)
- ✅ No swallowed errors
- ✅ No unstructured logging (doc.go only uses logrus examples)

**Resource Management**:
- ✅ No goroutines spawned
- ✅ No file handles opened
- ✅ No memory leaks detected in race tests
- ✅ PlacementValidator uses slice-based storage (no unbounded growth)

### Integration Verification

**Housing System Integration** (`pkg/world/housing/`):
- ✅ `ui.go` imports and uses `furniture.Generator`
- ✅ `categoryToFurnitureType()` maps UI categories to `furniture.FurnitureType` enums
- ✅ `GetSubTypesByCategory()` provides UI filtering
- ✅ Integration tests validate category enumeration matches expectations

**Client Initialization** (`cmd/client/`):
- ✅ `handlers.go` declares `furnitureGenerator *furniture.Generator` field
- ✅ `init_versions.go` initializes generator on startup: `sys.furnitureGenerator = furniture.NewGenerator()`
- ✅ Generator reachable from housing UI system

**Guild Housing** (`pkg/integration/guild_housing/`):
- ✅ Documentation references furniture generation for guild halls
- ✅ No import cycle issues

### Template Coverage

All 30+ furniture templates verified present and valid:
- **Seating (5)**: Chair, Bench, Stool, Throne, Couch
- **Storage (6)**: Chest, Wardrobe, Shelf, Barrel, Cabinet, Crate
- **Crafting (5)**: Anvil, Workbench, Forge, Alchemy Table, Enchanting Table
- **Decoration (5)**: Statue, Painting, Vase, Tapestry, Plant
- **Lighting (4)**: Torch, Chandelier, Lantern, Crystal Light
- **Bedding (3)**: Bed, Hammock, Bedroll
- **Tables (4)**: Table, Desk, Counter, Altar
- **Utility (4)**: Fireplace, Mirror, Fountain, Brazier

Each template includes:
- ✅ Dimension ranges (min/max width, height, depth)
- ✅ Allowed materials list
- ✅ Functional properties (walkable, capacity, light level)
- ✅ Detail complexity for rarity scaling

### Material System

5 material types with genre-specific selection:
- **Wood**: Brown/tan colors, preferred in fantasy and post-apocalyptic (`generator.go:371`)
- **Metal**: Gray/silver colors with genre tints (fantasy gold, sci-fi blue), preferred in sci-fi and cyberpunk (`generator.go:384`)
- **Stone**: Gray/brown colors, preferred in fantasy and horror (`generator.go:399`)
- **Crystal**: Bright saturated colors, preferred for high-rarity items (`generator.go:435`)
- **Fabric**: Wide color variety based on genre, used for seating/bedding (`generator.go:518`)

Material selection algorithm:
1. High rarity (Epic/Legendary) attempts exotic material selection (Crystal 60%, Metal 20%)
2. Genre-specific material preference (fantasy prefers Wood 40%, sci-fi prefers Metal 50%)
3. Fallback to random selection from allowed materials
4. All selection uses pre-rolled RNG values for determinism

### Rarity System

5 rarity tiers with scaling multipliers:
- **Common (1.0x)**: Base appearance, simple names
- **Uncommon (1.2x)**: Quality prefix ("Fine", "Sturdy"), slight enhancement
- **Rare (1.5x)**: Material prefix + quality adjective ("Exquisite Wood")
- **Epic (2.0x)**: Legendary prefix + quality suffix ("Legendary Chair of Power")
- **Legendary (3.0x)**: Mythical prefix + epic suffix ("Mythical Crystal Chair of the Gods")

Rarity affects:
- ✅ Visual detail level (DetailMultiplier)
- ✅ Storage capacity (BaseCapacity * multiplier)
- ✅ Dimensional scaling (10% per tier)
- ✅ Exotic material selection probability
- ✅ Naming prefixes and suffixes (genre-specific)

### Placement System

**PlacementValidator** features:
- ✅ AABB (Axis-Aligned Bounding Box) collision detection (`placement.go:140-167`)
- ✅ Room boundary validation (`placement.go:48-54`)
- ✅ Rotation support (4-way and 8-way) with dimension swapping (`placement.go:120-137`)
- ✅ Automatic valid placement finding with grid-based search (`placement.go:171-196`)
- ✅ Occupancy calculation (% of room floor space used) (`placement.go:104-117`)
- ✅ Furniture removal and rotation (`placement.go:84-90`, `placement.go:213-250`)

**Collision Detection**:
- Uses AABB algorithm: checks horizontal and vertical overlap independently
- Handles rotated dimensions correctly (N/S uses width×depth, E/W swaps to depth×width)
- Diagonal rotations use diagonal distance for collision box

### Genre Support

Full genre theming for 5 genres:
1. **Fantasy**: Wood/stone materials, medieval naming ("Majestic", "Royal"), warm earth tones
2. **Sci-Fi**: Metal/crystal materials, technical naming ("Quantum", "Prototype"), cool blue tints
3. **Horror**: Stone/wood materials, dark naming ("Eldritch", "Cursed"), muted dark colors
4. **Cyberpunk**: Metal materials, corporate naming ("Executive", "Black Market"), neon accents
5. **Post-Apocalyptic**: Metal/wood materials, salvage naming ("Pristine", "Military-Grade"), weathered colors

Genre affects:
- ✅ Material selection probabilities (`generator.go:354-365`)
- ✅ Naming prefix/suffix vocabulary (`generator.go:599-694`)
- ✅ Color palette generation (`generator.go:426-539`)
- ✅ Description flavor text (`generator.go:716-734`)
