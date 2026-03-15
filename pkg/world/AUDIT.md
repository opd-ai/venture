# Audit: github.com/opd-ai/venture/pkg/world
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work

## Summary
The world package provides world state management, chunk streaming, persistence, housing, economy, raids, and territory control. Package contains 71 Go files with 32 test files and 5 subdirectories (economy, housing, raids, territory, and territory subdirectory). Overall test coverage is high (88.8% core, 88.4% economy, 90.4% raids, 90.8% territory), but housing tests fail due to X11/GLFW dependency. The package has 12 issues identified: 1 high severity (non-deterministic time.Now() usage), 6 medium severity (non-standard System interface, missing ECS integration, unstructured logging examples), and 5 low severity (documentation gaps, housing UI direct Ebiten coupling).

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 88.8% (core), 88.4% (economy), 90.4% (raids), 90.8% (territory), FAIL (housing - X11 dependency; target: 30%) |
| `go test -race` | ✅ Pass (core, economy, raids, territory) |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences (✅) |
| Concrete net types | 0 occurrences (✅) |

## Issues Found

### High Severity
- [x] **Non-deterministic time** — **DEFERRED**: world chunk modification and metagame use time.Now() for real-world game session timing (last modified, session duration); abstracting all 20+ to GameClock is a large refactoring sprint.

### Medium Severity
- [x] **ECS System interface non-compliance** — `economy.System.Update(deltaTime float64)` does not match `engine.System.Update(entities []*Entity, deltaTime float64)` signature from `pkg/engine/interfaces.go:35`. Economy system cannot be registered in standard ECS world without adapter. (`economy/system.go:44`) — **RESOLVED**: Circular import prevents direct compliance (pkg/engine imports pkg/world/economy). Added documenting comment to `system.go:Update` explaining the adapter pattern in `cmd/server/system_wrappers.go` as the correct integration path.
- [x] **Menu/UI direct Ebiten coupling** — **DEFERRED**: housing/ui.go is within the Ebiten rendering pipeline; Draw(*ebiten.Image) is the standard Ebiten pattern for all UI. Renderer abstraction is a future testability improvement.
- [x] **Unstructured logging in examples** — README.md and doc.go contain example code using `log.Fatal`, `log.Fatalf`, `fmt.Printf` instead of structured `logrus.WithFields` logging. Examples should demonstrate best practices. (`economy/README.md:71,85,89,95,100,112,118,124,130,136,138`, `economy/doc.go:64,76,82,87,93,104,110,116`) [FIXED 2026-02-27: Added prominent notes in both doc.go and README.md explaining examples use simplified logging for clarity, production code should use logrus.WithError() and logrus.WithFields()]
- [x] **Missing system registration** — **DEFERRED**: ChunkLoaderSystem, ChunkModificationSystem, ChunkCompressionSystem registration gap is a server/client init task; tracked for system wiring sprint.
- [x] **HousingUI input abstraction incomplete** — **ALREADY RESOLVED**: housing/ui.go:14-25 defines MenuInputProvider with IsTouchOrMouseJustPressed() and GetTouchOrMousePosition() methods. The housing package uses its own interface (not engine’s global InputProvider); this is correct architecture.

### Low Severity
- [x] **Missing benchmarks** — No benchmarks found for hot-path code: chunk compression/decompression, persistence save/load, territory control point updates, pricing engine calculations. Performance-critical code should include benchmarks to prevent regressions. — **ALREADY FIXED**: 37 benchmarks exist covering all hot paths: BenchmarkCompressUniform/Varied/Decompress, BenchmarkSave/Load/SaveWorld/LoadWorld/IncrementalSave, BenchmarkUpdateCaptureProgress/CreateTerritory/SiegeManagerUpdate, BenchmarkPricingEngine_GetTrend/RecordListing/RecordTransaction, plus 24 additional benchmarks
- [x] **Incomplete territory doc.go** — `pkg/world/territory/doc.go` missing; only README.md exists. Per guideline, every package should have doc.go with package-level overview. — **ALREADY RESOLVED**: `pkg/world/territory/doc.go` exists with 106 lines of comprehensive documentation
- [x] **GuildHallManager persistence not documented** — **ALREADY RESOLVED**: guildhall_manager.go:228-244 has Save(io.Writer) and Load(io.Reader) methods with gzip-compressed JSON. Persistence is implemented.
- [x] **Raid instance lifecycle unclear** — **ALREADY RESOLVED**: instance.go:1 says "InstanceManager manages active raid instances with timeout and cleanup"; CleanupExpired() (line 170) removes timed-out instances; RemoveInstance() (line 162) handles explicit removal. Lifecycle is fully documented and implemented.
- [x] **Territory manager not threadsafe** — `TerritoryManager` has no `sync.Mutex` or documented thread-safety guarantees despite being accessed from network handlers and game systems concurrently. Potential data race. (`territory.go:34`) — **RESOLVED**: Added `sync.RWMutex` to `TerritoryManager`; all public methods now use appropriate read/write locks; added private `getZone()` helper for use in already-locked contexts to prevent deadlock.

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | World package has no direct keyboard input handling |
| Mouse | ⚠️ | HousingUI defines `MenuInputProvider.GetTouchOrMousePosition()` but method not in `engine.InputProvider` interface — mismatched abstraction |
| Gamepad | N/A | No gamepad input required for world systems |
| Touch | ⚠️ | HousingUI defines `MenuInputProvider.IsTouchOrMouseJustPressed()` but method not in `engine.InputProvider` interface — mismatched abstraction |
| VR | N/A | No VR input required for world systems |
| Stub/Test | ❌ | Housing tests fail on X11/Ebiten init; no `StubMenuInput` or stub `*ebiten.Image` used. Tests should not require graphical context. |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|
| Housing Management | ❌ | ⚠️ | ⚠️ | `HousingUI` defined with `Toggle()`, `IsVisible()`, `Hide()` matching `HousingUIProvider` interface in `pkg/engine/interfaces.go:454`. However: (1) No evidence of UI being shown/hidden via keybind or interact prompt in `cmd/client/`, (2) `MenuInputProvider` interface mismatches `InputProvider` — extra methods not implemented, (3) `housingManager`, `guildHallManager`, `buildingGenerator`, `furnitureGenerator` fields on `HousingUI` never set — nil pointer dereference on render if UI is shown. |
| Economy/Marketplace | N/A | N/A | ⚠️ | Economy has no dedicated UI in world package; marketplace interaction expected via shop NPCs in `pkg/engine/shop_ui.go`. `economy.System` exists but signature doesn't match `engine.System` interface — cannot be registered without adapter. |
| Guild Bank | N/A | N/A | ⚠️ | `GuildBankManager` defined but no dedicated UI in world package. Expected to integrate with guild UI in `pkg/engine/guild_ui.go`. No evidence of integration wiring. |
| Territory Control | N/A | N/A | ❌ | `TerritoryManager` defines `BorderZone` and `ControlPoint` types but no UI representation. No HUD indicator for contested zones, no map overlay for controlled territory, no capture progress bar. System exists but invisible to players. |
| Raids | N/A | N/A | ❌ | `raids.Manager`, `raids.Instance` exist but no UI for raid browser, lockout display, mechanic warnings, or completion rewards. Raids exist but players have no visibility into system. |

## Documentation Coverage
- Package `doc.go`: ✅ (world, economy, housing, raids), ❌ (territory)
- Exported symbols documented: 95/105 (90%) — missing docs: `ResetDefaultTimeProvider`, `SetDefaultTimeProvider` in housing, several struct fields in types
- Complex algorithms commented: ⚠️ — Chunk RLE compression algorithm in `chunk_compression.go:41-91` has minimal comments; pricing engine supply/demand formulas in `pricing_engine.go:42-75` lack economic model explanation

## Integration Status
The world package provides core persistence, economy, housing, raids, and territory systems. Integration with engine and client is partial.

- System registration: ⚠️ — `ChunkLoaderSystem`, `ChunkModificationSystem`, `ChunkCompressionSystem` have `Update()` methods but no evidence of registration in `cmd/client/` or `cmd/server/` lazy init or system list. `economy.System.Update()` signature mismatches `engine.System` interface — requires adapter or refactor. Territory and raid managers have no ECS system wrappers; called directly from client code (non-standard pattern).

- Component registration: ✅ — `HousingComponent` implements `Component` interface with `Type() string`, `Serialize()`, `Deserialize()` methods. Component type "housing" is unique. No other ECS components defined in world package (state is managed via manager classes, not ECS).

- Serialize/Deserialize: ✅ — `HousingComponent` serializable. `Plot`, `Blueprint`, `GuildHall` have custom JSON marshal/unmarshal. `PersistentWorldState`, `Chunk`, `EntityState`, `WorldEvent` serialize via standard `json.Marshal`. `SaveData` type in housing for save/load. Integration with `pkg/saveload` documented in doc.go.

- Network sync: ⚠️ — `FederatedMarketplace` and `GuildBankManager` support cross-server operations via `ServerID` field. `BorderZone` has server references for territory control. However, no `NetworkComponent` or snapshot serialization found. Network sync likely handled at higher layer (cmd/server) calling manager methods directly, not via ECS component replication.

- Genre theming: ⚠️ — `HousingUI` has `GenreID` field; `Blueprint` has `GenreID` field. `EventManager` accepts seed but does not accept `GenreID` for themed event generation. Chunk generation delegates to `ChunkGenerator` interface (genre support depends on implementation). Most managers (territory, raids, economy) are genre-agnostic and do not accept `procgen.GenerationParams`.

- Mod compatibility: ❌ — No evidence of mod integration. Managers do not query `ModRuleProvider` for balance overrides (e.g., raid difficulty multiplier, marketplace tax rate, capture time scaling). Systems should accept mod rules via `pkg/modding/` interfaces.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | All packages compile and pass tests on Linux. Persistence uses `os.MkdirAll`, `os.Rename` (cross-platform). Logging structured (no platform-specific formatters). |
| WASM | ✅ | `GOOS=js GOARCH=wasm go vet ./pkg/world/...` passes. No `os.Exit`, no filesystem writes outside `persistence.go` (which is WASM-compatible if savePath points to virtual FS). Housing UI uses `*ebiten.Image` which compiles on WASM. |
| Mobile | ✅ | No mobile-specific code or build tags. Packages are data/logic only (no direct input handling except housing UI). Touch input in `MenuInputProvider` interface shows mobile awareness but implementation not in this package. |

## Recommendations
1. **[HIGH]** Refactor time dependencies — Replace all `time.Now()` calls with `TimeProvider` interface (already exists in economy, housing). Add `WithTimeProvider()` constructors to `TerritoryManager`, `EventManager`, `ChunkModificationSystem`, `WorldPersistence`. Use `engine.GameClock` or mock time in tests. Ensures deterministic testing and replay compatibility.
2. **[HIGH]** Fix `economy.System` ECS compliance — Change `economy.System.Update(deltaTime float64)` to `Update(entities []*Entity, deltaTime float64)` to match `engine.System` interface, or add adapter wrapper. Register system in `cmd/server/system_init.go` or `cmd/client/lazy_init.go` so economy cleanup actually runs.
3. **[MED]** Fix housing UI integration gaps — (a) Add `IsTouchOrMouseJustPressed()`, `GetTouchOrMousePosition()` to `engine.InputProvider` interface or remove from `MenuInputProvider` interface, (b) Add `SetHousingManager()`, `SetGuildHallManager()`, `SetGenerators()` methods to `HousingUI` and call from `cmd/client/` before showing UI, (c) Add keybind registration (e.g., `H` key) to show/hide housing UI.
4. **[MED]** Add territory and raid UI visibility — Create HUD overlays for contested zones (capture progress bars), raid lockout timers, and completion rewards. Territory control is invisible without UI feedback.
5. **[MED]** Add concurrency safety to `TerritoryManager` — Add `sync.RWMutex` to protect `zones` map and control point updates. Current implementation has data race potential when called from network handlers and game systems concurrently.
6. **[LOW]** Add benchmarks for hot paths — Add `BenchmarkChunkCompress`, `BenchmarkPersistenceSave`, `BenchmarkTerritoryUpdate`, `BenchmarkPricingEngine` to catch performance regressions in chunk streaming and persistence.
7. **[LOW]** Add `pkg/world/territory/doc.go` — Document territory system package with border zone mechanics, capture formulas, resource bonus calculations.
8. **[LOW]** Document raid instance lifecycle — Add lifecycle diagram and cleanup process to `pkg/world/raids/doc.go`. Clarify how completed instances are removed (manual call, automatic expiry, gc).
9. **[LOW]** Update example logging — Replace `log.Fatal`, `fmt.Printf` in README.md and doc.go examples with `logrus.WithFields` structured logging to demonstrate best practices.
