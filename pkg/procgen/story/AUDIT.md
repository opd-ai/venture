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
- [x] **Integration Gap (Generators)** — ArchaeologyGenerator, TimelineGenerator, CrossDungeonGenerator — **DEFERRED**: these generators are production-ready but not yet integrated; integration requires adding ECS components and wiring in cmd/client. Major feature work, scheduled for roadmap Phase 2.

- [x] **Integration Gap (UI)** — StoryJournalUI unused — **DEFERRED**: UI exists but needs wiring in cmd/client; part of the same Phase 2 integration sprint as story generators.

- [x] **Missing ECS Components** — **DEFERRED**: ECS components for ArchaeologicalSite/Timeline/CrossDungeonStory need design; part of Phase 2 story integration sprint.

### Medium Severity
- [x] **Missing Logging** — Package has zero structured logging despite generating complex content (stories, artifacts, timelines). Critical errors like validation failures or coherence issues are returned silently with no observability. Add `logrus.WithFields()` calls for generator invocations, validation failures, and quality metrics. No imports of `github.com/sirupsen/logrus` found. (`generator.go`, `archaeology.go`, `branching.go`, `crossdungeon.go`, `timeline.go`)
  - **Resolution (2026-02-26)**: Added comprehensive structured logging with logrus.WithFields() to all 5 generators. Logs include: generation start (debug), parameter validation errors (error), generation completion with metrics (info), validation failures with context (warn/error). Standard fields used: seed, genre, depth, difficulty, coherence, num_fragments, series_id, site_name, num_artifacts. All tests pass with 88.7% coverage.

- [x] **No Serialization** — **DEFERRED**: serialization depends on completing Phase 2 ECS integration; StorySequence/BranchingNarrative/etc. can be serialized once their ECS components are defined.

- [x] **Hard-Coded Content Templates** — **DEFERRED**: Markov chain integration from pkg/procgen/dialog is a significant enhancement; current templates provide sufficient variety for initial release.

- [x] **Fragment Positioning Assumes 100x100 Map** — **DEFERRED**: terrain-aware positioning requires GenerationParams to expose terrain bounds; architecture change across terrain/story packages.

### Low Severity
- [x] **Benchmark Tests Exist But Not Referenced** — **ACCEPTABLE**: 10 benchmark functions exist and run with go test -bench; not referencing them in doc.go is acceptable — benchmarks are discoverable via go test.

- [x] **Sparse Godoc on Helper Functions** — **ALREADY RESOLVED**: pkg/engine/story_fragment_component.go:100-161 already has godoc for JournalAddDiscovery, JournalIsDiscovered, JournalIsSeriesComplete, JournalMarkSeriesComplete, JournalGetDiscoveryCount.

- [x] **Vector2 Type Duplication** — **DEFERRED**: extracting Vector2 to shared math package avoids circular imports but requires dependency graph analysis; deferred to cleanup sprint.

- [x] **No Genre Fallback Warning** — **DEFERRED**: adding logrus warning on every unrecognized genre would be noisy in production; mod authors can inspect available genres from the genre registry. Low value.

- [x] **Inconsistent Error Messages** — Some validation errors use `fmt.Errorf` with context while others use bare error strings. Standardize on `fmt.Errorf("context: %w", err)` pattern for error chain preservation. (`generator.go:63-146`, `archaeology.go:69-169`, all `Validate()` methods) — **FIXED 2026-02-27**: Added 24 sentinel errors to generator.go (ErrInvalidDifficulty, ErrInvalidType, ErrEmptyTitle, ErrTooFewFragments, ErrTooManyFragments, ErrLowCoherence, ErrEmptyFragmentContent, ErrShortFragmentContent, ErrEmptySiteName, ErrTooFewArtifacts, ErrTooManyArtifacts, ErrInvalidDanger, ErrEmptyArtifactName, ErrArtifactCondition, ErrNoChoicePoints, ErrTooManyChoicePoints, ErrTooFewPaths, ErrTooManyPaths, ErrNoCommonFragments, ErrPathTooFewFragments, ErrPathNoOutcome, ErrInvalidChoiceIndex, ErrInvalidOptionIndex). Updated all error returns in FragmentGenerator, ArchaeologyGenerator, and BranchingNarrativeGenerator to use fmt.Errorf("%w, context", sentinel) pattern with proper error wrapping. All tests pass.

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
