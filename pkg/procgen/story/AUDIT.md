# Audit: github.com/opd-ai/venture/pkg/procgen/story
**Date**: 2026-02-26
**Auditor**: GitHub Copilot (META_AUDIT v2 — Re-audit #2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
`pkg/procgen/story` generates environmental story fragments and advanced narrative systems (branching narratives, cross-dungeon stories, historical timelines, archaeology). The package is well-implemented with deterministic generation, clean ECS integration via `StoryFragmentComponent`, and strong test coverage (88.7%). However, **three of four major generators (archaeology, timeline, cross-dungeon) are not integrated** into the main game loop despite being fully implemented, tested, and documented. Additionally, **StoryJournalUI exists but is never instantiated or rendered**, leaving discovered fragments invisible to players. This represents significant feature gaps where ~60% of implemented code (4809 LOC) exists but is unreachable from gameplay.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 88.7% (target: 40%) |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **Integration Gap (Generators)** — `ArchaeologyGenerator`, `TimelineGenerator`, and `CrossDungeonGenerator` are fully implemented and tested but **never instantiated or called** in `cmd/client/`, `cmd/server/`, or `pkg/engine/`. Only `FragmentGenerator` and `BranchingNarrativeGenerator` are integrated. This means ~60% of the package's functionality (archaeology sites, historical timelines, cross-dungeon narratives) is dead code from the player's perspective despite being production-ready. (`archaeology.go:60-398`, `timeline.go:66-622`, `crossdungeon.go:44-430`)

- [ ] **Integration Gap (UI)** — `StoryJournalUI` exists in `pkg/rendering/ui/story_journal.go` with full implementation (232 LOC) including series list, fragment navigation, genre-based theming, and input handling, but is **never instantiated in cmd/client**. Players can discover fragments via `DiscoverySystem` but have no way to review them. Search results: 0 references to `StoryJournalUI` in `cmd/client/*.go`. (`pkg/rendering/ui/story_journal.go:1-232`)

- [ ] **Missing ECS Components** — No engine components exist for `ArchaeologicalSite`, `Timeline`, or `CrossDungeonStory` types, meaning even if generators were called, their output has no ECS representation. Compare to `StoryFragmentComponent` (`pkg/engine/story_fragment_component.go`) which properly integrates `FragmentGenerator`. (`archaeology.go:31-58`, `timeline.go:47-65`, `crossdungeon.go:26-38`)

### Medium Severity
- [x] **Missing Logging** — Package has zero structured logging despite generating complex content (stories, artifacts, timelines). Critical errors like validation failures or coherence issues are returned silently with no observability. Add `logrus.WithFields()` calls for generator invocations, validation failures, and quality metrics. No imports of `github.com/sirupsen/logrus` found. (`generator.go`, `archaeology.go`, `branching.go`, `crossdungeon.go`, `timeline.go`)
  - **Resolution (2026-02-26)**: Added comprehensive structured logging with logrus.WithFields() to all 5 generators. Logs include: generation start (debug), parameter validation errors (error), generation completion with metrics (info), validation failures with context (warn/error). Standard fields used: seed, genre, depth, difficulty, coherence, num_fragments, series_id, site_name, num_artifacts. All tests pass with 88.7% coverage.

- [ ] **No Serialization** — None of the story types (`StorySequence`, `BranchingNarrative`, `ArchaeologicalSite`, `Timeline`, `CrossDungeonStory`) implement `Serialize()/Deserialize()` methods. This means discovered stories, active narratives, excavation progress, and timelines **cannot persist across save/load** despite the package documentation claiming persistence. Grep results: 0 matches for `func.*Serialize\|func.*Deserialize` in `pkg/procgen/story/*.go`. (`generator.go:42-50`, `archaeology.go:31-58`, `branching.go:27-37`, `timeline.go:47-65`)

- [ ] **Hard-Coded Content Templates** — Story content generation uses fixed template arrays with hard-coded strings (e.g., `generateBeginningFragment` at `generator.go:194-209`). This limits narrative variety and makes content untranslatable. Consider data-driven templates or Markov chain integration from `pkg/procgen/dialog/`. (`generator.go:170-247`, `archaeology.go:209-262`, `timeline.go:241-320`)

- [ ] **Fragment Positioning Assumes 100x100 Map** — `generateLocation()` hard-codes `10.0 + progress*80.0` positioning logic assuming 100x100 dungeon space. This breaks layout on non-square or differently-sized terrains generated by `pkg/procgen/terrain/`. Should query actual terrain bounds from `GenerationParams` or terrain metadata. (`generator.go:267-276`)

### Low Severity
- [ ] **Benchmark Tests Exist But Not Referenced** — Package HAS comprehensive benchmarks (10 benchmark functions: `BenchmarkArchaeologyGenerate`, `BenchmarkExcavate`, `BenchmarkBranchingNarrativeGenerate`, `BenchmarkMakeChoice`, `BenchmarkCrossDungeonGenerate`, `BenchmarkIsFragmentAccessible`, `BenchmarkGenerate`, `BenchmarkValidate`, `BenchmarkTimelineGenerate`, `BenchmarkGetEventsInPeriod`) but previous audit incorrectly claimed they were missing. Issue downgraded from missing to documentation: benchmarks exist but are not documented in `doc.go` or CI/CD validation. (`archaeology_test.go:437-465`, `branching_test.go:362-390`, `crossdungeon_test.go:443-471`, `generator_test.go:402-433`, `timeline_test.go:437-465`)

- [ ] **Sparse Godoc on Helper Functions** — ECS-compliant helper functions like `JournalAddDiscovery`, `JournalIsDiscovered`, etc. lack individual godoc comments explaining parameters and return values. Only the deprecated methods have full documentation. (`pkg/engine/story_fragment_component.go:97-161`)

- [ ] **Vector2 Type Duplication** — `Vector2` is defined locally in `types.go:18-26` but `pkg/engine/components.go` likely has an equivalent type for positions. This creates coupling and conversion overhead. Consider importing engine's position type or extracting Vector2 to a shared math package. (`types.go:18-26`)

- [ ] **No Genre Fallback Warning** — Genre-specific functions like `getThemesForGenre()` default to generic themes for unknown genres, but this happens silently. Should log a warning when an unsupported genre is provided so mod authors know their custom genres need theme definitions. (`generator.go:153-168`)

- [ ] **Inconsistent Error Messages** — Some validation errors use `fmt.Errorf` with context while others use bare error strings. Standardize on `fmt.Errorf("context: %w", err)` pattern for error chain preservation. (`generator.go:63-146`, `archaeology.go:69-169`, all `Validate()` methods)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Package is data generation only, no input handling |
| Mouse | N/A | Package is data generation only, no input handling |
| Gamepad | N/A | Package is data generation only, no input handling |
| Touch | N/A | Package is data generation only, no input handling |
| VR | N/A | Package is data generation only, no input handling |
| Stub/Test | ✅ | All tests use deterministic `rand.New(rand.NewSource(seed))` |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Story Journal UI | ❌ | ✅ | ⚠️ | `StoryJournalUI` exists (`pkg/rendering/ui/story_journal.go`) with full navigation, genre theming, and input handling, but **never instantiated in cmd/client**. `DiscoverySystem` tracks `StoryJournalComponent` but players have no way to view discovered fragments. |
| Archaeology UI | ❌ | N/A | ❌ | No UI exists for excavation sites despite `ArchaeologicalSite.Excavation` progress field. Generator and data structures exist but are unused. |
| Timeline Viewer | ❌ | N/A | ❌ | `Timeline` with historical events and eras has no in-game viewer UI. Would integrate with lore/codex system if one existed. |

## Test Coverage
**Coverage**: 88.7% (target: 40%)
- Missing test areas: None significant; all generators, validators, and helper functions well-covered
- Missing benchmarks: ❌ CORRECTION — Package HAS 10 comprehensive benchmarks covering all generators and key operations (archaeology excavation, branching choice, cross-dungeon accessibility, fragment generation, narrative validation, timeline generation, event queries). Previous audit error corrected.
- Table-driven test compliance: ✅ — All `*_test.go` files use table-driven patterns

## Documentation Coverage
- Package `doc.go`: ✅ — Comprehensive 147-line package doc with usage examples
- Exported symbols documented: 47/47 (100%)
- Complex algorithms commented: ✅ — Coherence calculation, prerequisite setup, and fragment distribution logic all have inline comments

## Integration Status
Package generates story content consumed by `DiscoverySystem` via `StoryFragmentComponent`. Only 2 of 5 generators are reachable from gameplay.

- System registration: ⚠️ — `DiscoverySystem` registered (`cmd/client/init_versions.go:148`), `BranchingNarrativeSystem` registered (`cmd/client/handlers.go:2036`), but no systems exist for archaeology, timeline, or cross-dungeon stories
- Component registration: ⚠️ — `StoryFragmentComponent` and `StoryJournalComponent` registered in engine, `BranchingNarrativeComponent` exists, but missing components for archaeology sites, timelines, and cross-dungeon narratives
- Serialize/Deserialize: ❌ — None of the story types implement persistence despite `StoryJournalComponent.LastDiscoveryTime` suggesting state tracking
- Network sync: N/A — Story generation is client-side and deterministic (same seed → same content on all clients)
- Genre theming: ✅ — All generators read `params.GenreID` and adapt content for fantasy/scifi/horror/cyberpunk/postapocalyptic
- Mod compatibility: ⚠️ — No integration with `pkg/modding/` for custom story templates or fragment types, but deterministic generation ensures consistent results

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | Pure Go with no platform-specific imports |
| WASM | ✅ | WASM vet passes; no `syscall/js` or file I/O |
| Mobile | ✅ | No mobile-specific concerns; generator is stateless |

## Recommendations
1. **[HIGH]** Wire up `StoryJournalUI` in `cmd/client/handlers.go` with keybind (suggest `L` for Lore) and register in game state machine. Add to HUD with "Journal" button. This is the highest-impact fix as it makes existing discovered fragments visible to players.
2. **[HIGH]** Implement ECS components and spawning logic for `ArchaeologicalSite`, `Timeline`, and `CrossDungeonStory` in `pkg/engine/` (following `StoryFragmentComponent` pattern). Add client-side calls to spawn these in `cmd/client/util.go` alongside `spawnStoryFragments()`.
3. **[HIGH]** Add `Serialize()/Deserialize()` methods to all story types and register them with `pkg/saveload/`. Excavation progress, discovered timelines, and cross-dungeon story state must persist.
4. **[MED]** Add structured logging with `logrus.WithFields(logrus.Fields{"seed": seed, "genre": params.GenreID, "generator": "story"})` to all `Generate()` methods. Log validation failures and coherence metrics at WARN level.
5. **[MED]** Extract hard-coded story templates into JSON files under `mods/` directory, allowing mod authors to define custom narratives. Integrate with `pkg/modding/` for override support.
6. **[MED]** Query actual terrain bounds from `pkg/procgen/terrain/` instead of assuming 100x100 in `generateLocation()`. Pass terrain metadata via `GenerationParams.Custom` map with keys `"terrain_width"` and `"terrain_height"`. Fallback to 100x100 if not provided for backward compatibility.
7. **[LOW]** Document existing benchmarks in package `doc.go` (lines 134-147) and add CI/CD performance regression check with thresholds (e.g., <20ms per generate call, <50ms for cross-dungeon stories).
8. **[LOW]** Replace local `Vector2` with engine's `PositionComponent` (X, Y fields) or extract to shared `pkg/math/` package to avoid duplication and conversion overhead.
9. **[LOW]** Add godoc comments to all ECS-compliant helper functions in `story_fragment_component.go` (lines 97-161).
10. **[LOW]** Log warning when unknown genre is provided to theme/title selection functions so custom genres can be detected (e.g., `log.WithFields(log.Fields{"genre": genreID}).Warn("unknown genre in story generation, using fallback themes")`).
