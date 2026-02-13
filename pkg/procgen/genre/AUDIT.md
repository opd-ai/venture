# Audit: github.com/opd-ai/venture/pkg/procgen/genre
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The genre package provides centralized genre definitions and blending for procedural content generation. It has excellent test coverage (100%), comprehensive documentation, and clean API design. However, it contains one high-severity issue with non-deterministic random genre selection and two medium-severity error handling gaps that need addressing before production use.

## Issues Found
- [ ] **severity:high** Deterministic procgen — `GetRandomTheme()` uses `time.Now()` for non-deterministic selection, violating deterministic generation requirement (`predefined.go:125`)
- [ ] **severity:med** Error handling — `parseHexColor()` silently ignores `strconv.ParseInt` errors, returns (0,0,0) on malformed hex strings without logging (`blender_utils.go:121-123`)
- [ ] **severity:med** Error handling — `DefaultRegistry()` ignores genre registration errors; bugs in predefined genres would be silently swallowed (`registry.go:79`)
- [ ] **severity:low** Documentation — `GetRandomTheme()` godoc claims "non-deterministic selection" which conflicts with project determinism requirement (`predefined.go:121`)

## Test Coverage
100.0% (target: 65%) ✅

## Integration Status
Package is integrated with 3 consumers:
- `cmd/client/util.go` — Uses `GetTheme()` for genre selection from CLI flags
- `pkg/engine/genre_selection_menu.go` — Genre selection UI
- `pkg/rendering/palette/generator.go` — Color palette generation from genre themes

Package provides:
- Genre type definitions with validation
- Registry for genre management (DefaultRegistry with 5 predefined genres)
- GenreBlender for hybrid genre creation with 5 preset blends
- Helper function `GetTheme(genreID)` that handles "random" special value

**No system registration required** — Pure data/utility package, not an ECS system.

**Serialization support**: Not applicable — genres are configuration data, not entity components.

## Recommendations
1. **CRITICAL**: Replace `GetRandomTheme()` time-based selection with seed-based deterministic selection
   - Change signature to `GetRandomTheme(seed int64) *Genre`
   - Update callers to pass seed (from world seed or CLI seed flag)
   - Update `GetTheme("random")` to require seed parameter or derive from global world seed
   - Alternative: Remove "random" support entirely and require explicit genre selection
2. **HIGH**: Add error handling/logging to `parseHexColor()` for malformed hex strings
   - Log warning with `logrus.WithFields` when ParseInt fails
   - Consider returning error instead of silently defaulting to black (0,0,0)
3. **HIGH**: Fix `DefaultRegistry()` to panic or log error on registration failure
   - Change to `panic(err)` if predefined genre is invalid (fail-fast at init)
   - Or use `logrus.Fatal` with structured fields
   - Rationale: Registration failure indicates programmer error in predefined genres
4. **LOW**: Update `GetRandomTheme()` documentation to reflect deterministic requirement after fix
