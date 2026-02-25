# Audit: pkg/procgen/class
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The class package provides procedural character class generation with 21 presets (6 base + 15 hybrid classes). Code quality is excellent with proper deterministic generation, comprehensive tests, and no anti-patterns. However, **the package is not integrated into the actual character creation flow** — all classes are hardcoded in `pkg/engine/class_progression_component.go` using static switch statements instead of calling this generator. The generator is instantiated in `cmd/client/handlers.go` but never used.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | Unmeasurable (requires X11; 87.4% test-to-source ratio: 412 test lines / 471 source lines) |
| `go test -race` | Unmeasurable (requires X11; race detector requires display) |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **Integration Gap** — ClassGenerator is instantiated in `cmd/client/handlers.go:classGenerator` but **never called**. All class data is hardcoded in `pkg/engine/class_progression_component.go` with static `GetClassAbilities()` and `GetAvailableSpecializations()` functions using switch statements. Character creation (`pkg/engine/character_creation.go`) uses the hardcoded data, not the generator. (`generator.go:1`, `cmd/client/handlers.go:52`, `pkg/engine/class_progression_component.go:189`)
- [ ] **Genre Theming Missing** — `Generate()` accepts `params.GenreID` but **does not use it** to adapt class names, descriptions, or abilities. All 21 classes have fantasy-themed hardcoded names/descriptions regardless of genre (sci-fi, horror, cyberpunk). (`generator.go:350-394`)

### Medium Severity
- [ ] **No Multiclass Support** — ClassPreset does not expose hybrid class parent classes or stat blending ratios. `pkg/class/advanced/` exists for advanced multiclassing but has no integration with this generator. (`generator.go:13-35`)
- [ ] **No Save/Load Integration** — ClassPreset does not implement `ComponentSerializer` interface. Character class data persistence relies on hardcoded engine types, not generated presets. (`generator.go:13-35`)

### Low Severity
- [ ] **No Benchmark for Validate()** — Only `BenchmarkClassGenerator_Generate()` exists; validation performance is untested. (`generator_test.go:365-377`)
- [ ] **Custom Logger Not Tested in Generate** — `TestNewClassGeneratorWithLogger` only tests constructor, not whether custom logger is actually used during Generate/Validate. (`generator_test.go:379-412`)
- [ ] **GetAllPresets() Relies on Implicit Enum Order** — Loop `for i := 0; i < len(g.presets); i++` assumes CharacterClass enum is contiguous starting at 0. If enum gaps exist (e.g., deprecated class removed), this silently skips them. (`generator.go:458-466`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Generator is data-only; no input handling |
| Mouse | N/A | Generator is data-only; no input handling |
| Gamepad | N/A | Generator is data-only; no input handling |
| Touch | N/A | Generator is data-only; no input handling |
| VR | N/A | Generator is data-only; no input handling |
| Stub/Test | N/A | Generator is data-only; no input handling |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Character Creation | ✅ | ✅ | ❌ | Character creation UI exists in `pkg/engine/character_creation.go` and is reachable from main menu. UI displays hardcoded classes from `GetClassAbilities()` switch statements, **not** from ClassGenerator. Generator is instantiated but unused. |

## Test Coverage
**Coverage**: Unmeasurable (requires X11; 30% target accepted; 87.4% test-to-source ratio)
- Missing test areas: None (comprehensive table-driven tests for all 21 classes)
- Missing benchmarks: `Validate()` performance, `GetAllPresets()` performance
- Table-driven test compliance: ✅

## Documentation Coverage
- Package `doc.go`: ✅ (concise summary)
- Exported symbols documented: 6/6 (100%)
- Complex algorithms commented: ✅ (preset initialization, random selection)

## Integration Status
**The class generator is a data provider that is never consumed by the game systems.**

- System registration: ❌ — ClassGenerator is not a System (no `Update()`); it is a standalone generator
- Component registration: ❌ — ClassPreset is not a Component (no `Type()` method); it is a plain struct
- Serialize/Deserialize: ❌ — ClassPreset does not implement serialization; character classes are saved as `engine.CharacterClass` enum values, not generated presets
- Network sync: ❌ — Character classes sync as enum values, not as generated ClassPresets
- Genre theming: ❌ — `params.GenreID` accepted but **not used** to theme class names/descriptions/abilities
- Mod compatibility: ❌ — No mod hook for overriding class presets; all data is compiled-in

**Current Architecture**:
1. `cmd/client/handlers.go` instantiates `classGenerator` field (line 52)
2. Generator is never called in character creation flow
3. `pkg/engine/character_creation.go` uses hardcoded `CharacterClass` enum
4. `pkg/engine/class_progression_component.go` provides abilities via `GetClassAbilities()` switch statement
5. ClassPreset data exists but is unreachable from gameplay

**Expected Architecture**:
1. Character creation should call `classGenerator.GetAllPresets()` to list classes
2. Player selection should call `classGenerator.Generate(seed, params)` with chosen class type
3. ClassPreset should be converted to `ClassProgressionComponent` on entity
4. Genre parameter should influence class theming (names, descriptions, abilities)

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code; pure data generator |
| WASM | ✅ | WASM vet passes; no syscalls or filesystem access |
| Mobile | ✅ | No platform-specific code; pure data generator |

## Recommendations
1. **[HIGH]** Wire ClassGenerator into character creation flow: `pkg/engine/character_creation.go` should call `classGenerator.GetAllPresets()` to display class list, then `Generate()` on selection. Replace hardcoded `GetClassAbilities()` switch with generator output.
2. **[HIGH]** Implement genre theming: Use `params.GenreID` to adapt class names (e.g., "Warrior" → "Cyberknight" for sci-fi, "Survivor" for horror) and descriptions. Load genre-specific ability name mappings.
3. **[MED]** Add multiclass metadata to ClassPreset: For hybrid classes, expose `ParentClasses []CharacterClass` and `StatBlendRatio float64` fields. Integrate with `pkg/class/advanced/` for advanced multiclassing.
4. **[MED]** Implement serialization: Add `Serialize()`/`Deserialize()` to ClassPreset or create a conversion function to `ClassProgressionComponent` for save/load.
5. **[LOW]** Add validation benchmark: `BenchmarkClassGenerator_Validate(b *testing.B)` to ensure validation performance is acceptable.
6. **[LOW]** Test custom logger in Generate: Extend `TestNewClassGeneratorWithLogger` to verify logger fields are used during error logging.
7. **[LOW]** Add enum safety to GetAllPresets: Check for missing enum values and log warning or return error if preset map is incomplete.
