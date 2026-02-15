# Audit: github.com/opd-ai/venture/pkg/procgen/furniture
**Date**: 2026-02-15
**Status**: Complete

## Summary
The furniture package implements procedural generation of 30+ furniture types across 8 categories with deterministic seed-based algorithms, placement validation, and genre-specific theming. Overall code quality is high with 89.5% test coverage (exceeding 65% target), proper ECS-compliant architecture (pure data types, no behavioral methods), comprehensive documentation, and excellent determinism enforcement. Minor issues identified relate to future enhancement comments and missing input parameter validation.

## Issues Found
- [ ] <severity:low> **Stub/incomplete code** — `chooseRandomSubType` has comment "Could be weighted by depth/difficulty in future" indicating incomplete depth-based weighting logic (`generator.go:196`)
- [ ] <severity:low> **Error handling** — `Generate` method does not validate `params.Difficulty` range (should be 0.0-1.0) or `params.Depth` (should be non-negative) before use (`generator.go:119`)
- [ ] <severity:low> **Error handling** — `Generate` method does not validate `params.GenreID` is non-empty before using in genre-specific logic (`generator.go:119`)
- [ ] <severity:low> **Doc coverage** — Unexported helper functions lack godoc comments: `generateDimensions`, `calculateCollisionBox`, `calculateCapacity`, `calculateLightIntensity`, `buildFurniture` (`generator.go:35-117`)

## Test Coverage
89.5% (target: 65%) ✅

**Test files**: 3 files, 1,329 LOC
- `generator_test.go` — 13 test functions, 2 benchmarks
- `placement_test.go` — 11 test functions, 2 benchmarks  
- `coverage_improvement_test.go` — 11 test functions

**Test quality**:
- ✅ Table-driven tests used throughout
- ✅ Determinism verified with same-seed tests (`TestGenerateDeterminism`)
- ✅ Benchmarks for performance-critical paths
- ✅ All 30+ furniture templates tested
- ✅ Edge cases covered (invalid placements, boundary conditions, rotation)

## Integration Status
**Fully integrated** with the following packages:

1. **Housing System** (`pkg/world/housing/`):
   - `HousingUI` uses `furniture.Generator` for furniture creation
   - Integration tests verify furniture placement and crafting
   - UI provides furniture placement menu (`menuState = "furniture"`)

2. **Procedural Generation Audit** (`pkg/procgen/audit/`):
   - Determinism tests (`determinism_test.go`)
   - Quality validation tests (`quality_test.go`)
   - Edge case testing (`edgecase_test.go`)

3. **Standard Generator Interface** (`pkg/procgen/`):
   - ✅ Implements `Generate(seed int64, params GenerationParams) (interface{}, error)`
   - ✅ Implements `Validate(result interface{}) error`
   - ✅ Uses `GenerationParams` with Difficulty, Depth, GenreID, Custom fields

**No ECS components found** — This is expected and correct. The `Furniture` type is a pure data structure used for procedural generation, not an ECS component. Housing placement would use separate components in the engine layer.

**No serialization methods** — `Furniture` and `PlacedFurniture` lack `Serialize()`/`Deserialize()` methods. This may be needed if furniture state needs persistence in save files. Current integration suggests furniture is regenerated on demand rather than persisted.

## Recommendations
1. **[LOW] Implement depth-based furniture weighting** — Complete the `chooseRandomSubType` logic to prefer common furniture at low depth and decorative/rare types at high depth as documented (`generator.go:186-197`)
2. **[LOW] Add input parameter validation** — Validate `params.Difficulty` (0.0-1.0), `params.Depth` (≥0), and `params.GenreID` (non-empty) at the start of `Generate()` for defensive programming
3. **[LOW] Document helper functions** — Add godoc comments to unexported helper functions for code maintainability
4. **[OPTIONAL] Consider serialization** — Evaluate if `Furniture` needs `Serialize()`/`Deserialize()` methods for save file persistence; current regeneration approach may be sufficient
