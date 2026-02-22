# Audit: github.com/opd-ai/venture/pkg/procgen/furniture
**Date**: 2026-02-22
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete

## Summary

The furniture package provides deterministic procedural generation of 30+ furniture types across 8 categories for player housing and guild halls. The package follows ECS data-driven patterns, uses seed-based deterministic generation throughout, and integrates well with the housing UI system. Test coverage exceeds the 65% target at 92.5%.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 92.5% (target: 65%) |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
None identified.

### Medium Severity
- [ ] **Doc coverage** — `doc.go` contains example code using `log.Fatal` and `fmt.Printf` which is acceptable for documentation examples but deviates from logrus guideline (`doc.go:48,52,64,70,176,180`)

### Low Severity
- [ ] **Missing Serialize/Deserialize** — `Furniture` struct lacks `Serialize()` and `Deserialize()` methods for save/load persistence integration (`types.go:168-199`)
- [ ] **No Component interface** — `Furniture` does not implement `Component` interface (no `Type() string` method), limiting direct ECS integration (`types.go:168-199`)
- [ ] **Missing mod compatibility** — No explicit mod loader hook for furniture data overrides (`generator.go`, `templates.go`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Package is a procgen generator, no direct input handling |
| Mouse | N/A | Package is a procgen generator, no direct input handling |
| Gamepad | N/A | Package is a procgen generator, no direct input handling |
| Touch | N/A | Package is a procgen generator, no direct input handling |
| VR | N/A | Package is a procgen generator, no direct input handling |
| Stub/Test | N/A | Generator tests don't require input stubs |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Housing UI | ✅ | ✅ | ✅ | Furniture generator wired via `pkg/world/housing/ui.go:34,206` |

The furniture generator is consumed by the Housing UI system (`pkg/world/housing/ui.go`), which provides the user-facing furniture selection and placement interface. The Housing UI is reachable via keybind in the client.

## Test Coverage
**Coverage**: 92.5% (target: 65%)
- Missing test areas: None significant - all major code paths covered
- Missing benchmarks: None - `BenchmarkGenerate`, `BenchmarkValidate`, `BenchmarkValidatePlacement`, `BenchmarkFindValidPlacement` all present
- Table-driven test compliance: ✅ Extensive use of table-driven tests throughout

## Documentation Coverage
- Package `doc.go`: ✅ Comprehensive 191-line package documentation with examples
- Exported symbols documented: 97/97 (100%) - All exported types, functions, and methods have godoc comments
- Complex algorithms commented: ✅ Rarity selection, material selection, dimension scaling, collision detection all documented

## Integration Status

### System registration
✅ — Generator is instantiated in `cmd/client/init_versions.go:449` via `furniture.NewGenerator()` and stored in `systemsContainer.furnitureGenerator`

### Component registration
❌ — `Furniture` struct does not implement `Component` interface. It is used as a data transfer object rather than an ECS component. Housing UI converts furniture data to ECS components for placement.

### Serialize/Deserialize
❌ — Not implemented. `Furniture` struct lacks persistence methods. Housing system likely handles persistence separately via `pkg/saveload/`.

### Network sync
N/A — Furniture generation is deterministic (same seed = same output), so clients can regenerate locally rather than syncing furniture data over network.

### Genre theming
✅ — Generator reads `GenreID` from `GenerationParams` and adapts:
- Material selection (`generator.go:354-420`: fantasy, scifi, cyberpunk, horror, postapoc branches)
- Color generation (`generator.go:426-538`: genre-specific palettes)
- Naming prefixes and suffixes (`generator.go:599-694`: genre-specific vocabulary)

### Mod compatibility
❌ — Templates are hardcoded in `templates.go`. No mod loader integration for custom furniture types or property overrides.

### Accessibility
N/A — Generator produces data; accessibility concerns handled by consuming UI systems.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code, pure Go package |
| WASM | ✅ | Passes WASM vet, no browser-incompatible imports |
| Mobile | ✅ | No platform-specific code |

## Recommendations
1. **[LOW]** Add `Serialize()`/`Deserialize()` methods to `Furniture` struct for direct save/load support. Currently housing system must handle persistence separately.
2. **[LOW]** Consider implementing `Component` interface (`Type() string`) on `Furniture` to enable direct ECS attachment when furniture is placed in the world.
3. **[LOW]** Add mod loader hook in `templates.go` to allow `pkg/modding/` to inject custom furniture templates at load time.
4. **[LOW]** Example code in `doc.go` uses `log.Fatal` and `fmt.Printf`; consider adding a note that production code should use logrus.
