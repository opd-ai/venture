# Audit: pkg/companion/learning
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete

## Summary
The companion learning subsystem implements AI skill progression, personality evolution, and behavioral memory for companion NPCs. It provides well-architected deterministic learning mechanics with 92.5% test coverage, clean ECS integration via wrapper system, proper time abstraction for reproducibility, and comprehensive JSON serialization. The package demonstrates strong engineering practices with structured logging, thread-safe operations, and fail-soft error handling. One medium-severity issue identified: use of `time.Now()` in tests should use `MockTimeProvider` for full determinism.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 92.5% (target: 40%) |
| `go test -race` | ✅ Pass |
| WASM vet | N/A (no WASM-specific code) |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 (all uses seed-based via rand.New) |
| Concrete net types | 0 |

## Issues Found

### High Severity
*None identified*

### Medium Severity
- [x] **Deterministic Testing** — Test files use `time.Now()` directly instead of `MockTimeProvider` in 15 locations (`system_test.go:31,53,82,84,204`, `manager_test.go:356,377,398,413-415,483,523,584`). While production code correctly uses TimeProvider abstraction, tests should also use MockTimeProvider for full determinism and reproducibility. This violates the codebase's determinism principle for test isolation. **FIXED 2026-02-27**: Replaced all 15 time.Now() calls with MockTimeProvider usage. System tests now create MockTimeProvider and pass it to NewManagerWithOptions. All manual CompanionLearningSystem struct initializations now include timeProvider and lastUpdate fields. Coverage improved from 92.5% to 92.7%.

### Low Severity
- [x] **Documentation** — `time_provider.go:23` contains only implementation (`time.Now()`). While acceptable for RealTimeProvider, adding a comment explaining this is intentional production behavior vs. test mocks would improve clarity for maintainers. **FIXED 2026-02-27**: Added clarifying comment explaining time.Now() is intentional for companion learning metadata (real elapsed time), not procedural generation. References MockTimeProvider for deterministic testing.
- [x] **Constants** — Several "magic numbers" exist in `manager.go` without named constants: `0.01` (trait adjustment), `0.02` (social interaction), `0.05` (combat style adaptation). These are defined in `constants.go` but not consistently used. Suggest refactoring to use constants for all personality deltas. **FIXED 2026-02-27**: Replaced all magic numbers with named constants: TraitSmallDelta (0.01), TraitMediumDelta (0.02), TraitLargeDelta (0.05), ExplorationCuriosityDelta (0.015), and ExplorationPracticalDelta (-0.005). All personality trait adjustments now use these constants for better maintainability and consistency.
- [x] **Error Messages** — Functions `AddExperience`, `CanLearnSkill`, `LearnSkill` return errors with simple strings. Consider using wrapped errors with `fmt.Errorf(...%w...)` for better error chain preservation, matching error handling guideline in custom instructions. - **FIXED 2026-02-27**: Added sentinel errors (ErrSkillNotFound, ErrInsufficientSkillPoints, ErrPrerequisiteNotFound, ErrPrerequisiteNotMet) to types.go and updated all error returns to use fmt.Errorf with %w wrapping

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Package has no direct input responsibilities |
| Mouse | N/A | Package has no direct input responsibilities |
| Gamepad | N/A | Package has no direct input responsibilities |
| Touch | N/A | Package has no direct input responsibilities |
| VR | N/A | Package has no direct input responsibilities |
| Stub/Test | N/A | Package has no direct input responsibilities |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Companion Skills UI | ⚠️ | ⚠️ | ✅ | No dedicated UI system found in pkg/rendering/ui/ or pkg/engine/ for companion skill tree display. Backend system fully functional, but UI integration not verified. Skill data accessible via `GetCompanionSkillBonus`, `GetSkillsByType`, etc. |
| Companion Personality UI | ⚠️ | ⚠️ | ✅ | No dedicated UI system found for personality trait display. Backend fully functional with `GetDominantTrait`, `GeneratePersonalityDescription`. May be integrated into existing companion info panels. |
| Companion Memory UI | ⚠️ | ⚠️ | ✅ | No dedicated UI for viewing companion memories. Backend provides `GetRecentEvents`, `GetEventsByType`, `GetMemorySummary`. May be accessible through dialog or companion info. |

## Documentation Coverage
- Package `doc.go`: ✅ Comprehensive 124-line package documentation with usage examples, performance characteristics, determinism guarantees
- Exported symbols documented: 62/62 (100%)
- Complex algorithms commented: ✅ Skill tree initialization, trait balancing, LRU eviction, combat style adaptation all well-commented

## Integration Status
**System registration**: ✅ — `pkg/engine/companion_learning_system.go` provides ECS wrapper. Registered in `cmd/client/init_versions.go:176` unconditionally. Wired into `cmd/client/handlers.go:566` as `companionLearningSystem`. System added to world in `cmd/client/handlers.go:2053`.

**Component registration**: ✅ — `CompanionLearningComponent` implements `Component` interface with `Type() string` returning "companion_learning". Added to entities via `entity.AddComponent(learningComp)` in `pkg/engine/companion_learning_system.go:117`.

**Serialize/Deserialize**: ✅ — Full JSON serialization implemented in `types.go:266-445`. Supports skill tree, personality, memory, and LastSkillUse persistence. Uses Unix timestamps for cross-platform determinism. Backward-compatible with default cost fallback at line 385.

**Network sync**: ⚠️ — Component is serializable but not verified in network snapshot system. No references found in `pkg/network/` to "companion_learning" component. Server-side integration not confirmed (no hits in `cmd/server/`). Likely client-only currently, which may cause multiplayer desyncs if companions travel between players.

**Genre theming**: N/A — Learning mechanics are genre-agnostic. Skill names/descriptions could be themed but are not currently parameterized by genre.

**Mod compatibility**: ⚠️ — No mod integration points found. Skill trees, XP rates, personality trait weights are hardcoded. `pkg/modding/` could expose these as tunable rules. No mod hooks for custom skills or personality traits.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code. Pure data structures and business logic. |
| WASM | ✅ | No filesystem access, no syscall/js. TimeProvider abstraction supports WASM. JSON serialization works in WASM. |
| Mobile | ✅ | No platform-specific dependencies. Touch input irrelevant (no UI layer). |

## Recommendations
1. **[HIGH]** **Network Synchronization** — Verify CompanionLearningComponent is included in network snapshot system for multiplayer. Add component to server-side entity spawning in `cmd/server/entity_spawning.go`. Add sync tests to `cmd/server/snapshot_conversion_test.go`. Without this, companion skill/personality state will desync between clients in multiplayer games.
2. **[MED]** **Test Determinism** — Refactor all test `time.Now()` calls to use `MockTimeProvider`. Add a test utility function `newTestCompanion(seed int64) *CompanionLearningComponent` that creates companions with fully deterministic time. This ensures tests are reproducible across CI runs and developer machines.
3. **[MED]** **UI Integration Verification** — Audit pkg/rendering/ui/ and pkg/engine/*_ui.go for companion skill tree, personality, and memory UIs. If missing, create design document for companion UI integration. If present but not found in audit, update this document with file references.
4. **[LOW]** **Performance Benchmarks** — Add benchmarks for `AddExperience`, `AdjustTrait`, `AddEvent`, and `Update` (100 companions). Validate doc.go performance claims (<10µs, <5µs, <2µs, <50ms). Use `testing.B` with `b.ReportAllocs()` to track memory pressure.
5. **[LOW]** **Mod Support** — Expose skill tree definitions, XP curves, personality trait weights, and decay rates via `pkg/modding/` rule system. Add mod events for skill learned, personality shift, and memorable event recorded. Enable data-driven companion tuning without code changes.
