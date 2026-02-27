# Codebase Audit Remediation Plan
**Generated**: 2026-02-27
**Scope**: All *AUDIT*.md files in repository
**Total Unresolved Findings**: 238

## Summary by Severity
| Severity | Count |
|----------|-------|
| CRITICAL | 2 |
| HIGH | 54 |
| MEDIUM | 114 |
| LOW | 68 |

## Findings

### CRITICAL

#### REM-097: Resource cleanup not verified
- **Source**: `./pkg/network/AUDIT.md`
- **Problem**: - [x] No explicit goroutine leak prevention tests. With 21 mutex/RWMutex usages and multiple goroutines , verify all goroutines have clean shutdown paths with timeout tests.
- **Fix**:
  1. Add `sync.RWMutex` field to manager struct
  2. Add `mu.Lock()` / `defer mu.Unlock()` to mutating methods
  3. Add `mu.RLock()` / `defer mu.RUnlock()` to read methods
  4. Document thread-safety guarantees in godoc
- **Verify**: `go test -race ./pkg/... -v`
- **Completed**: 2026-02-27 - Added 4 comprehensive goroutine leak tests verifying all server and client goroutines terminate on shutdown with timeout. Tests use runtime.NumGoroutine() for leak detection. All tests pass with race detector.

#### REM-159: Hard-Coded Content Templates
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(e.g., `generateBeginningFragment` at `generator.go:194-209`)`
- **Problem**: - [ ] Story content generation uses fixed template arrays with hard-coded strings . This limits narrative variety and makes content untranslatable. Consider data-driven templates or Markov chain integ...
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

### HIGH

#### REM-022: Test Execution
- **Source**: `./cmd/server/AUDIT.md`
- **Location**: `(`main_test.go:1`, all test files)`
- **Problem**: - [ ] Tests require X11/display but no Xvfb wrapper documented
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-023: WASM Build
- **Source**: `./cmd/server/AUDIT.md`
- **Problem**: - [ ] WASM vet fails due to `pkg/migration/validator.go:39` referencing `saveload.NewDefaultMigrator` which doesn't exist in WASM build context
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-034: <category>
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] <description>
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-035: <category>
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] <description>
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-043: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] Menu/UI integration table is populated for all menus relevant to the package
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-044: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] Platform-specific checks completed
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-045: Documentation
- **Source**: `./pkg/audit/features/AUDIT.md`
- **Location**: `(`constants.go:10-20`)`
- **Problem**: - [ ] Feature category constants  are exported but lack individual godoc comments . While the file has a package-level doc, each constant should document which feature domains it covers.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-048: No benchmark tests
- **Source**: `./pkg/balance/AUDIT.md`
- **Location**: `(`balance_test.go:1`)`
- **Problem**: - [ ] Package performs computationally intensive simulations  but lacks benchmarks to validate performance targets . Add `BenchmarkCombatValidator`, `BenchmarkEconomicValidator`, etc.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-052: Documentation
- **Source**: `./pkg/class/advanced/AUDIT.md`
- **Location**: `(`talents.go:713`)`
- **Problem**: - [ ] `initializeTalentTrees` method lacks godoc comment
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-053: Documentation
- **Source**: `./pkg/class/advanced/AUDIT.md`
- **Location**: `(`talents.go:4`)`
- **Problem**: - [ ] `buildSynergies` function lacks godoc comment
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-055: Documentation
- **Source**: `./pkg/combat/AUDIT.md`
- **Location**: `(`interfaces.go:6-8`)`
- **Problem**: - [ ] File-level comment in `interfaces.go` has redundant/stale package description lines 6-8 from historical refactoring
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-058: Time Dependency
- **Source**: `./pkg/hostplay/AUDIT.md`
- **Location**: `(`time_provider.go:19`)`
- **Problem**: - [ ] `time.Now()` used in production code . While properly abstracted via TimeProvider interface for testing, this creates non-deterministic behavior in production. Consider using an injected game cl...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-062: Documentation
- **Source**: `./pkg/integration/companion_housing/AUDIT.md`
- **Location**: `(`doc.go:18-34`)`
- **Problem**: - [ ] Component doc string in doc.go  references deprecated method names `IsInHouse()`, `HasBedding()`, `IsTraining()`, etc. that are now internal to `companionHousingSystem` . External code should us...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-066: Missing UI Integration
- **Source**: `./pkg/integration/guild_housing/AUDIT.md`
- **Location**: `(`pkg/engine/guild_ui.go:1-200`, `pkg/integration/guild_housing/guild_housing_manager.go:1-716`)`
- **Problem**: - [ ] Guild housing manager is initialized in server v9_systems.go and client handlers.go, but GuildUI  does NOT call any guild_housing manager methods. The UI only displays guild info, members, and t...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-067: Component Not Used
- **Source**: `./pkg/integration/guild_housing/AUDIT.md`
- **Location**: `(`types.go:57-66`)`
- **Problem**: - [ ] `GuildHousingComponent` is defined  but never attached to entities in engine code. Component exists for ECS integration but is not wired to any system.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-070: Integration Completeness
- **Source**: `./pkg/integration/guild_vehicle/AUDIT.md`
- **Location**: `(`fleet_manager.go:327-388`, `types.go:231-281`)`
- **Problem**: - [ ] FleetManager persistence  is implemented but no integration with `pkg/saveload/` persistence system. Fleets are saved to standalone gzip files instead of participating in unified save/load workf...
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-071: Integration Completeness
- **Source**: `./pkg/integration/guild_vehicle/AUDIT.md`
- **Location**: `(`time_provider.go:1-55`)`
- **Problem**: - [ ] TimeProvider abstraction is package-local; similar patterns exist in other packages . Consider extracting to shared utility package to avoid duplication and enable cross-package deterministic te...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-076: Network Sync
- **Source**: `./pkg/integration/narrative_world/AUDIT.md`
- **Problem**: - [ ] No evidence of network sync integration for cross-server companion state. StoryEventManager has Serialize/Deserialize but no network snapshot system integration.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-080: Test dependency
- **Source**: `./pkg/integration/political_warfare/AUDIT.md`
- **Location**: `(`manager_test.go:7`, `system_test.go:7`)`
- **Problem**: - [ ] Tests import `pkg/engine` which requires X11/Wayland/Ebiten, preventing CI execution. Pattern is consistent with other integration packages. Test-to-source ratio  indicates comprehensive testing...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-087: Documentation
- **Source**: `./pkg/integration/world_events/AUDIT.md`
- **Location**: `(`FactionResponse`, `EconomicEvent`, `WeatherDisaster`, `EventChain` in `types.go:96-133`)`
- **Problem**: - [ ] Package `doc.go` exists but some exported types lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-088: Documentation
- **Source**: `./pkg/integration/world_events/AUDIT.md`
- **Problem**: - [ ] Several exported functions in `events.go` have minimal or missing parameter documentation
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-095: Non-deterministic timing
- **Source**: `./pkg/network/AUDIT.md`
- **Location**: `(`bandwidth.go:41,54,67,101,132,158,169,177`, `chat.go:18,189,200,231,232,257,271,318,355`, `client.go:253,254,495`, `server.go`: multiple occurrences in ping/timeout logic)`
- **Problem**: - [ ] Extensive use of `time.Now()` in bandwidth monitoring, chat, and client/server state  makes networking non-deterministic and prevents time-controlled testing. Networking systems should accept in...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-098: Documentation
- **Source**: `./pkg/network/chat/AUDIT.md`
- **Location**: `(`system.go:132`)`
- **Problem**: - [ ] `generateMessageID()` function is unexported but lacks internal documentation explaining collision resistance properties and why 128 bits is sufficient
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-102: Documentation
- **Source**: `./pkg/network/federation/mobile/AUDIT.md`
- **Location**: `(`types.go:221-232`)`
- **Problem**: - [ ] Exported symbols `GetBytesAvailable` and `SetBytesAvailable` in types.go lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-103: Edge Case
- **Source**: `./pkg/network/federation/mobile/AUDIT.md`
- **Location**: `(`adapter.go:105`)`
- **Problem**: - [ ] `UpdateBatteryLevel` accepts values outside 0.0-1.0 range without validation; could lead to incorrect BatteryMode calculation
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-107: Stub Implementation Boundary
- **Source**: `./pkg/network/federation/webrtc/AUDIT.md`
- **Location**: `(`stun_test.go:221`, `stun_test.go:261`, `signaling_test.go:162-163`, `signaling_test.go:346`, `time_provider_test.go:10-12`)`
- **Problem**: - [ ] `time.Now()` usage in tests  uses real system clock instead of `MockTimeProvider`. This is acceptable for production tests but should be noted as non-deterministic test behavior.
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-114: Integration Gap
- **Source**: `./pkg/procgen/class/AUDIT.md`
- **Location**: `(`generator.go:1`, `cmd/client/handlers.go:52`, `pkg/engine/class_progression_component.go:189`)`
- **Problem**: - [ ] ClassGenerator is instantiated in `cmd/client/handlers.go:classGenerator` but **never called**. All class data is hardcoded in `pkg/engine/class_progression_component.go` with static `GetClassAb...
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-115: No Multiclass Support
- **Source**: `./pkg/procgen/class/AUDIT.md`
- **Location**: `(`generator.go:13-35`)`
- **Problem**: - [ ] ClassPreset does not expose hybrid class parent classes or stat blending ratios. `pkg/class/advanced/` exists for advanced multiclassing but has no integration with this generator.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-126: Documentation
- **Source**: `./pkg/procgen/furniture/AUDIT.md`
- **Location**: `(`placement.go:1`)`
- **Problem**: - [ ] `placement.go` lacks package-level documentation comment explaining AABB collision algorithm and grid-based auto-placement strategy
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-149: Test Execution
- **Source**: `./pkg/procgen/recipe/AUDIT.md`
- **Problem**: - [ ] Tests fail with "GLFW library is not initialized" due to engine import requiring X11. This is a common pattern in the codebase , but prevents automated CI testing in headless environments. Recom...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-152: Documentation
- **Source**: `./pkg/procgen/skills/AUDIT.md`
- **Problem**: - [ ] `GetFantasyTreeTemplates`, `GetSciFiTreeTemplates`, `GetHorrorTreeTemplates`, `GetCyberpunkTreeTemplates`, `GetPostApocalypticTreeTemplates` in `templates.go` lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-155: Integration Gap (Generators)
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`archaeology.go:60-398`, `timeline.go:66-622`, `crossdungeon.go:44-430`)`
- **Problem**: - [ ] `ArchaeologyGenerator`, `TimelineGenerator`, and `CrossDungeonGenerator` are fully implemented and tested but **never instantiated or called** in `cmd/client/`, `cmd/server/`, or `pkg/engine/`. ...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-156: Integration Gap (UI)
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`pkg/rendering/ui/story_journal.go:1-232`)`
- **Problem**: - [ ] `StoryJournalUI` exists in `pkg/rendering/ui/story_journal.go` with full implementation  including series list, fragment navigation, genre-based theming, and input handling, but is **never insta...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-157: Missing ECS Components
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`archaeology.go:31-58`, `timeline.go:47-65`, `crossdungeon.go:26-38`)`
- **Problem**: - [ ] No engine components exist for `ArchaeologicalSite`, `Timeline`, or `CrossDungeonStory` types, meaning even if generators were called, their output has no ECS representation. Compare to `StoryFr...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-176: Build tag coverage
- **Source**: `./pkg/rendering/lighting/AUDIT.md`
- **Location**: `(`gpu_bloom_headless.go:1-39`)`
- **Problem**: - [ ] `gpu_bloom.go` has `//go:build !headless` and `gpu_bloom_headless.go` has `//go:build headless`, but no automated test verifies the headless stub compiles and provides no-op behavior correctly
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-177: Deprecated field
- **Source**: `./pkg/rendering/lighting/AUDIT.md`
- **Location**: `(`types.go:117-122`)`
- **Problem**: - [ ] `LightingConfig.EnableShadows` is marked deprecated with clear documentation, but there's no linter annotation  to trigger static analysis warnings when used
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-183: Integration
- **Source**: `./pkg/rendering/patterns/AUDIT.md`
- **Location**: `(`cmd/client/handlers.go:260`)`
- **Problem**: - [ ] Pattern generator initialized but not actively called in game loop  — While the generator is correctly initialized, there's no evidence of active usage in terrain/tile generation systems. Verify...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-190: Determinism
- **Source**: `./pkg/rendering/quality/AUDIT.md`
- **Location**: `(`performance_monitor.go:48`, `performance_monitor.go:125`, `performance_monitor.go:138`)`
- **Problem**: - [ ] PerformanceMonitor uses `time.Now()` for `lastAdjustment` tracking . This is acceptable for UI performance monitoring but violates Coding Guideline #2 for deterministic systems. Consider using g...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-193: Documentation
- **Source**: `./pkg/rendering/shapes/AUDIT.md`
- **Location**: `(types.go:133)`
- **Problem**: - [ ] Shape.Type field in Shape struct  is redundant with Config.Type and unused throughout the codebase
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-194: Documentation
- **Source**: `./pkg/rendering/shapes/AUDIT.md`
- **Location**: `(types.go:132-145)`
- **Problem**: - [ ] Shape struct  appears to be legacy/unused; all code uses Config instead; consider deprecating or removing to reduce API confusion
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-197: Documentation
- **Source**: `./pkg/rendering/sprites/AUDIT.md`
- **Location**: `(`generator.go:39-42`, `generator.go:44-57`, `equipment.go:4-21`, `equipment.go:24-38`, `equipment.go:41-75`, `equipment.go:78-95`, `equipment.go:98-114`)`
- **Problem**: - [ ] Many exported functions lack godoc comments. Only 883 exported symbols found but manual inspection shows missing doc comments on several helper functions
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-201: Time.Now usage in tests
- **Source**: `./pkg/rendering/ui/AUDIT.md`
- **Problem**: - [ ] Multiple test files  use `time.Now()` for test data initialization. While acceptable in tests, consider using a deterministic time provider for more reliable test behavior across different execu...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-203: time.Now() usage
- **Source**: `./pkg/saveload/AUDIT.md`
- **Location**: `(`manager.go:97`, `recovery.go:242`, `types.go:686`, `memory_manager.go:59`, `storage_wasm.go:137`, `storage_wasm.go:475`)`
- **Problem**: - [ ] `time.Now()` is used in 6 locations for save timestamps . timestamps are metadata for user display, not used for deterministic game logic. Per coding guideline #2, deterministic generation appli...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-208: Integration
- **Source**: `./pkg/social/persistence/AUDIT.md`
- **Problem**: - [ ] Package not wired into server-side system initialization ` and `persistence.NewReputationManager()` but no actual initialization)
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-209: Integration
- **Source**: `./pkg/social/persistence/AUDIT.md`
- **Location**: `(`cmd/client/handlers.go:139-142`)`
- **Problem**: - [ ] TrustManager and ReputationManager instantiated in client handlers but no save/load integration with `pkg/saveload` or server persistence layer
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-214: Version Duplication
- **Source**: `./pkg/version/AUDIT.md`
- **Location**: `(`pkg/saveload/types.go:11`, `pkg/version/version.go:32`)`
- **Problem**: - [ ] SaveVersion is hardcoded as "1.0.0" in `pkg/saveload/types.go:11` instead of importing from `pkg/version`. This creates a second source of truth that can drift out of sync and cause save incompa...
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-215: Missing CLI Flag
- **Source**: `./pkg/version/AUDIT.md`
- **Location**: `(`pkg/version/version.go:71`, `cmd/client/main.go`, `cmd/server/main.go`)`
- **Problem**: - [ ] PrintVersion() uses fmt.Println instead of returning string for CLI integration. Client and server use --version flag but force os.Exit in main.go, preventing integration testing of version disp...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-222: Non-deterministic time
- **Source**: `./pkg/world/AUDIT.md`
- **Location**: `(`chunk_modification.go:102`, `territory.go:61,63,72,108,194`, `economy/types.go:20,135`, `persistence.go:102,218`, `metagame.go:85,86,109,110,136,137,160,161`)`
- **Problem**: - [ ] 20+ uses of `time.Now()` in production code . Violates deterministic generation guideline. Should accept time provider via dependency injection or use GameClock interface for testable, controlla...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-223: ECS System interface non-compliance
- **Source**: `./pkg/world/AUDIT.md`
- **Location**: `(`economy/system.go:44`)`
- **Problem**: - [ ] `economy.System.Update` does not match `engine.System.Update` signature from `pkg/engine/interfaces.go:35`. Economy system cannot be registered in standard ECS world without adapter.
- **Fix**:
  1. Update method signature to match interface: `Update(entities []*Entity, deltaTime float64)`
  2. Add entities parameter if missing
  3. Verify struct implements interface via type assertion in test
  4. Check compilation succeeds
- **Verify**: `go build ./pkg/...`

#### REM-224: Menu/UI direct Ebiten coupling
- **Source**: `./pkg/world/AUDIT.md`
- **Location**: `(`housing/ui.go:158`)`
- **Problem**: - [ ] `HousingUI.Draw` takes concrete `*ebiten.Image` instead of `Renderer` interface or `interface{}` with type assertion like other UISystem implementations. Makes testing without Ebiten runtime har...
- **Fix**:
  1. Replace direct `ebiten.Image` parameter with `Renderer` interface or `interface{}`
  2. Add type assertion `screen, ok := img.(*ebiten.Image)` inside method
  3. Update caller sites to pass screen as interface{}
  4. Verify tests can use StubRenderer without Ebiten runtime
- **Verify**: `go test ./pkg/... -v`

#### REM-232: Documentation
- **Source**: `./pkg/world/housing/AUDIT.md`
- **Location**: `(`doc.go:1-101`)`
- **Problem**: - [ ] Missing package-level overview of system architecture in doc.go beyond basic usage
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-236: Missing persistence
- **Source**: `./pkg/world/raids/AUDIT.md`
- **Problem**: - [ ] RaidInstance and PlayerLockout types do not implement `Serialize()`/`Deserialize()` methods for save/load support. Instances and lockouts lost on server restart. Add ComponentSerializer interfac...
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-237: Missing mod support
- **Source**: `./pkg/world/raids/AUDIT.md`
- **Problem**: - [ ] No integration with `pkg/modding` for adjusting raid difficulty multipliers, lockout periods, or boss mechanic parameters. Add ModRuleProvider integration to allow data-driven balance tuning.
- **Fix**:
  1. Import `pkg/modding` package
  2. Add `ModRuleProvider` field to manager struct
  3. Query mod rules for values: `modProvider.GetFloat('raid.difficulty_multiplier', default)`
  4. Register mod hooks in manager constructor
- **Verify**: `grep -r 'ModRuleProvider' ./pkg/world/raids/ | wc -l`

#### REM-241: No Persistence Layer
- **Source**: `./pkg/world/territory/AUDIT.md`
- **Location**: `(`manager.go:13-20`, `siege.go:73-100`, `types.go:56-67,92-103,106-114`)`
- **Problem**: - [ ] Manager, Siege, Territory, WarDeclaration, DefensiveStructure structs lack `Serialize()`/`Deserialize()` methods. Grep result: 0 matches for `func.*Serialize|func.*Deserialize` in `pkg/world/ter...
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

### MEDIUM

#### REM-001: uncategorized
- **Source**: `./AUDIT.md`
- **Problem**: - [ ] All 501 findings addressed
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-002: uncategorized
- **Source**: `./AUDIT.md`
- **Problem**: - [ ] All verification commands pass
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-003: uncategorized
- **Source**: `./AUDIT.md`
- **Problem**: - [ ] No `- [ ]` items remain in any *AUDIT*.md file
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-004: Missing Platform Detection
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `(`mobile.go:16-22`)`
- **Problem**: - [ ] No runtime platform detection  for platform-specific optimizations, haptic feedback, or safe area insets. Fixed screen dimensions  ignore device aspect ratios.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-005: No Mobile Federation
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `(`mobile.go:74-88`)`
- **Problem**: - [ ] `mobile_federation_system.go` exists in pkg/network/federation/mobile but is not initialized in cmd/mobile. Mobile multiplayer federation unavailable.
- **Fix**:
  1. Import `pkg/network/federation` package
  2. Add `BroadcastUpdate()` method calling `federation.SyncManager.Broadcast()`
  3. Implement `HandleFederatedRequest()` for incoming cross-server actions
  4. Add circuit breaker pattern for federation failures
- **Verify**: `grep -r 'federation' ./pkg/world/territory/ | wc -l`

#### REM-006: Time-Based Seed Fallback
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `(config/seed.go:36)`
- **Problem**: - [ ] `config.GetSeedFromEnv` uses `time.Now().UnixNano()` when env var unset . Violates deterministic generation guideline for non-reproducible worlds. **Mitigation**: Documented as intentional for m...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-007: No Orientation Handling
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `(`mobile.go:16-22`)`
- **Problem**: - [ ] Hard-coded portrait orientation . No landscape support, device rotation handling, or safe area insets for iOS notch/Android navigation bar.
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-013: No Virtual Controls
- **Source**: `./cmd/mobile/AUDIT.md`
- **Location**: `(`mobile.go:52-72`)`
- **Problem**: - [ ] Virtual on-screen controls  not initialized despite being available in `pkg/mobile`.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-014: Platform Detection Missing
- **Source**: `./cmd/mobile/AUDIT.md`
- **Location**: `(`mobile.go:1-410`)`
- **Problem**: - [ ] No runtime platform detection  for platform-specific optimizations or input mapping.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-015: No Mobile Federation
- **Source**: `./cmd/mobile/AUDIT.md`
- **Location**: `(`mobile.go:74-88`)`
- **Problem**: - [ ] Mobile federation system  exists but not initialized for mobile builds.
- **Fix**:
  1. Import `pkg/network/federation` package
  2. Add `BroadcastUpdate()` method calling `federation.SyncManager.Broadcast()`
  3. Implement `HandleFederatedRequest()` for incoming cross-server actions
  4. Add circuit breaker pattern for federation failures
- **Verify**: `grep -r 'federation' ./pkg/world/territory/ | wc -l`

#### REM-016: Time-Based Seed Fallback
- **Source**: `./cmd/mobile/AUDIT.md`
- **Location**: `(`config/seed.go:36`)`
- **Problem**: - [ ] Uses `time.Now()` for seed generation when env var not set. Documented as intentional for mobile UX, but violates determinism guideline. Consider showing seed in UI for reproducibility.
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-024: Time.Now Usage
- **Source**: `./cmd/server/AUDIT.md`
- **Location**: `(`main.go:298`, `main.go:751`)`
- **Problem**: - [ ] `time.Now()` used for server timing is legitimate but not deterministic for save/load replay; consider using GameClock abstraction
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-029: uncategorized
- **Source**: `./docs/DETERMINISM_AUDIT.md`
- **Problem**: - [ ] Accepts `seed int64` parameter
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-030: uncategorized
- **Source**: `./docs/DETERMINISM_AUDIT.md`
- **Problem**: - [ ] Creates local RNG from seed
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-031: uncategorized
- **Source**: `./docs/DETERMINISM_AUDIT.md`
- **Problem**: - [ ] Never uses global rand functions
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-032: uncategorized
- **Source**: `./docs/DETERMINISM_AUDIT.md`
- **Problem**: - [ ] Added to determinism test suite
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-033: uncategorized
- **Source**: `./docs/DETERMINISM_AUDIT.md`
- **Problem**: - [ ] Passes 1000-run acceptance test
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-036: <category>
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] <description>
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-046: Documentation
- **Source**: `./pkg/audit/features/AUDIT.md`
- **Location**: `(`meta_features.go:250`)`
- **Problem**: - [ ] `GetDefaultRegistry()` function lacks godoc comment explaining its purpose and usage . This is the primary public API entry point and should be documented.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-047: Code Organization
- **Source**: `./pkg/audit/features/AUDIT.md`
- **Location**: `(`feature_completeness.go:106-109`)`
- **Problem**: - [ ] `Register()` method silently ignores nil features without logging . Consider using structured logging with `logrus.WithFields.Warn` for better observability in test/audit runs.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-049: Progress logging frequency
- **Source**: `./pkg/balance/AUDIT.md`
- **Location**: `(`combat.go:131-136`, `economic.go:106-111`, etc.)`
- **Problem**: - [ ] Progress logs use fixed intervals . For long-running simulations , this creates 10 log entries that may be excessive. Consider dynamic adjustment based on total duration or make interval configu...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-054: Code Organization
- **Source**: `./pkg/class/advanced/AUDIT.md`
- **Problem**: - [ ] Large talent tree definitions  could benefit from splitting into per-class files for maintainability
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-056: Design/Dependency
- **Source**: `./pkg/config/AUDIT.md`
- **Location**: `(`validator.go:32`)`
- **Problem**: - [ ] Package depends on `pkg/procgen/dialog.GetAvailableGenres()` for genre validation. This creates a dependency from a utility package to a generation package. Consider:  extracting genre list to s...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-059: Context Timeout
- **Source**: `./pkg/hostplay/AUDIT.md`
- **Location**: `(`server_manager.go:624`)`
- **Problem**: - [ ] `Stop()` method uses hardcoded 5-second timeout . Consider making this configurable or documenting rationale for 5s choice
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-060: Code Organization
- **Source**: `./pkg/integration/choice_consequences/AUDIT.md`
- **Location**: `(`helpers.go:9,20`)`
- **Problem**: - [ ] `abs` and `clamp` helper functions in `helpers.go` could be replaced with `math.Abs` and a standard `math` library clamp once Go 1.21+ is adopted. Current implementation is correct but duplicate...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-064: Consistency
- **Source**: `./pkg/integration/companion_housing/AUDIT.md`
- **Location**: `(`companion_housing_system.go:72-83`)`
- **Problem**: - [ ] Method `UpdateFromManager` on `companionHousingSystem` uses `manager.GetLoyaltyBonus` pattern, but `PetHomeManager.GetLoyaltyBonus` already checks `companionHomes` map internally. Minor redundan...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-068: Missing Serialization
- **Source**: `./pkg/integration/guild_housing/AUDIT.md`
- **Location**: `(`pkg/engine/interfaces.go:515-519`)`
- **Problem**: - [ ] `GuildHousingComponent` does not implement `Serialize()/Deserialize()` methods required by `ComponentSerializer` interface . Manager has `Save()/Load()` but component persistence is not wired.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-072: Documentation
- **Source**: `./pkg/integration/guild_vehicle/AUDIT.md`
- **Location**: `(`doc.go:64-68`)`
- **Problem**: - [ ] `doc.go` states "PLANNED " integrations but does not specify timeline or tracking issue. Consider adding GitHub issue links or removing stale PLANNED section if integrations are deferred indefin...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-073: Documentation
- **Source**: `./pkg/integration/housing_crafting/AUDIT.md`
- **Problem**: - [ ] Missing package-level example code for V9ValidationService integration pattern
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-074: API Surface
- **Source**: `./pkg/integration/housing_crafting/AUDIT.md`
- **Location**: `(`housing_crafting_system.go:10`)`
- **Problem**: - [ ] `housingCraftingSystem` is unexported but only used in tests; consider removing if StationManager fully replaces it
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-077: Location Tracking
- **Source**: `./pkg/integration/narrative_world/AUDIT.md`
- **Location**: `(`manager.go:474`)`
- **Problem**: - [ ] RecordMemory hardcodes `Location: "unknown"` with comment "Could be enhanced with location tracking". Could integrate with PositionComponent for better memory context.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-078: Genre Propagation
- **Source**: `./pkg/integration/narrative_world/AUDIT.md`
- **Location**: `(`manager.go:402`)`
- **Problem**: - [ ] GeneratePersonalQuest hardcodes `GenreID: "fantasy"` instead of accepting/propagating actual genre from game state.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-081: Naming typo
- **Source**: `./pkg/integration/political_warfare/AUDIT.md`
- **Location**: `(`types.go:54`)`
- **Problem**: - [ ] `ResponingAllies` should be `RespondingAllies` in `AllianceCall` struct and JSON tags.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-082: Unexported helper exposure
- **Source**: `./pkg/integration/political_warfare/AUDIT.md`
- **Location**: `(`manager.go:106`, `manager.go:146`, `time_provider.go:45`)`
- **Problem**: - [ ] `now()` helper function uses package-level `defaultTimeProvider` which is testable but could be more explicit by accepting a time provider parameter in Manager constructor. Current pattern works...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-084: Documentation
- **Source**: `./pkg/integration/trade_routes/AUDIT.md`
- **Location**: `(`manager.go:72`)`
- **Problem**: - [ ] `RouteManager.priceHandler` field not documented in struct godoc. Field added later and doc not updated.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-085: Code organization
- **Source**: `./pkg/integration/trade_routes/AUDIT.md`
- **Location**: `(`manager_test.go:661,716,766,827,877,978,980`)`
- **Problem**: - [ ] `time.Now()` usage in test files acceptable but scattered across many tests. Consider centralizing test time mocking utilities.
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-092: Non-deterministic time usage
- **Source**: `./pkg/memprofile/AUDIT.md`
- **Location**: `(`profile.go:40,59,84`)`
- **Problem**: - [ ] Package uses `time.Now()` for real-time profiling, which is acceptable for this use case but violates general guideline. This is intentional for profiling real execution time.
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-093: fmt.Printf usage
- **Source**: `./pkg/memprofile/AUDIT.md`
- **Location**: `(`profile.go:195-225`)`
- **Problem**: - [ ] `PrintProfile()` uses `fmt.Printf` instead of structured logging. This is intentional for human-readable report output, but could optionally provide a structured JSON export method.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-094: Documentation Clarity
- **Source**: `./pkg/modding/AUDIT.md`
- **Location**: `(`doc.go:113-120`)`
- **Problem**: - [ ] `doc.go:113-120` states "This package uses time.Now() in the following non-procgen contexts" but the exception rationale could be clearer about why rate limiting is acceptable as non-determinist...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-096: No GameClock injection
- **Source**: `./pkg/network/AUDIT.md`
- **Problem**: - [ ] `TCPClient`, `TCPServer`, `ChatManager`, and `BandwidthMonitor` hardcode `time.Now()` instead of accepting injectable `GameClock` interface . This prevents deterministic testing and replay. Refa...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-099: Component Type Assertion
- **Source**: `./pkg/network/chat/AUDIT.md`
- **Location**: `(`system.go:115-121`)`
- **Problem**: - [ ] Uses direct type assertion without logging the actual type received when assertion fails, making debugging harder
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-100: Documentation
- **Source**: `./pkg/network/federation/guild/AUDIT.md`
- **Location**: `(`constants.go:1`, `treasury.go:1`, `persistence.go:1`)`
- **Problem**: - [ ] Package-level `doc.go` exists but individual files lack file-level comments explaining their purpose in the larger architecture
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-101: Time Dependency
- **Source**: `./pkg/network/federation/guild/AUDIT.md`
- **Problem**: - [ ] `time_provider.go:23` uses `time.Now()` in production RealTimeProvider, which is non-deterministic. This is acceptable for timestamps but documented here for awareness. The package correctly pro...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-104: Edge Case
- **Source**: `./pkg/network/federation/mobile/AUDIT.md`
- **Location**: `(`types.go:89`, `adapter.go:210`)`
- **Problem**: - [ ] `Config.MaxBandwidth` can be negative which breaks token bucket algorithm in `executeSyncWithBandwidthLimit`
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-109: API consistency
- **Source**: `./pkg/network/resilience/AUDIT.md`
- **Location**: `(`scenario.go:240-270`)`
- **Problem**: - [ ] `formatInt()` and `trimTrailingZeros()` in scenario.go  are custom string formatting implementations that could use standard library `strconv.FormatInt()` and `strconv.FormatFloat()` for maintai...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-110: Documentation
- **Source**: `./pkg/procgen/audit/AUDIT.md`
- **Location**: `(`doc.go:34`)`
- **Problem**: - [ ] Package doc.go mentions "EnvironmentGenerator" exclusion but doesn't explain why Config-based API is incompatible with seed/params audit pattern
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-111: Documentation
- **Source**: `./pkg/procgen/AUDIT.md`
- **Location**: `(`generator.go:47-53`)`
- **Problem**: - [ ] Missing inline algorithm explanation for polynomial rolling hash in `SeedGenerator.GetSeed`  - While the function has a godoc comment mentioning the algorithm, an inline comment explaining why b...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-112: Optimization
- **Source**: `./pkg/procgen/AUDIT.md`
- **Location**: `(`generator.go:89-99`)`
- **Problem**: - [ ] `ValidateDimensions` performs 6 comparisons which could be reduced to 4 by checking ranges together
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-113: Code Organization
- **Source**: `./pkg/procgen/building/AUDIT.md`
- **Location**: `(`generator.go:447-533`)`
- **Problem**: - [ ] `generator.go:447-533` guild hall layout generation methods  could be extracted into a separate file `guild_hall_layout.go` for better navigability given guild halls are a distinct feature with ...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-116: No Save/Load Integration
- **Source**: `./pkg/procgen/class/AUDIT.md`
- **Location**: `(`generator.go:13-35`)`
- **Problem**: - [ ] ClassPreset does not implement `ComponentSerializer` interface. Character class data persistence relies on hardcoded engine types, not generated presets.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-117: Documentation
- **Source**: `./pkg/procgen/companion/AUDIT.md`
- **Location**: `(`doc.go:1-96`)`
- **Problem**: - [ ] No explicit mention in doc.go that this package integrates with `CompanionAISystem`, `CompanionLearningSystem`, and housing systems, despite full integration existing.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-118: API Consistency
- **Source**: `./pkg/procgen/companion/AUDIT.md`
- **Location**: `(`generator.go:18`)`
- **Problem**: - [ ] `Companion` struct is not exported with godoc comment explaining it is the generation result type. Add doc comment: `// Companion represents a generated companion with stats, commands, and visua...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-119: Documentation
- **Source**: `./pkg/procgen/dialog/AUDIT.md`
- **Location**: `(`personality.go:205`)`
- **Problem**: - [ ] `GetGreeting()` method comment could clarify backward compatibility behavior vs. `GetGreetingWithSeed()` for better API discoverability
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-120: Code clarity
- **Source**: `./pkg/procgen/dialog/AUDIT.md`
- **Location**: `(`utils.go:136-147`)`
- **Problem**: - [ ] `calculateTemperatureWeights()` has complex temperature scaling logic that would benefit from inline comments explaining the mathematical transformation
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-121: Code organization
- **Source**: `./pkg/procgen/entity/AUDIT.md`
- **Location**: `(`queries.go:9-27`)`
- **Problem**: - [ ] `queries.go` uses nil safety checks in all functions but the pattern could be extracted to a helper
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-123: Documentation
- **Source**: `./pkg/procgen/environment/AUDIT.md`
- **Location**: `(`doc.go:1-45`)`
- **Problem**: - [ ] Package-level doc.go has triple package comment
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-124: Code organization
- **Source**: `./pkg/procgen/environment/AUDIT.md`
- **Location**: `(`generator.go:1-1296`)`
- **Problem**: - [ ] Generator.go is large  and could benefit from splitting drawing functions into a separate file
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-127: Code Organization
- **Source**: `./pkg/procgen/furniture/AUDIT.md`
- **Location**: `(`generator.go:573-737`)`
- **Problem**: - [ ] Large `generateName()` and `generateDescription()` functions  could be refactored into lookup tables for genre-specific prefixes/suffixes to improve maintainability
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-128: Documentation
- **Source**: `./pkg/procgen/genre/AUDIT.md`
- **Location**: `(`blender.go:13-18`)`
- **Problem**: - [ ] `BlendedGenre` struct fields lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-129: Documentation
- **Source**: `./pkg/procgen/genre/AUDIT.md`
- **Location**: `(`blender.go:21-23`)`
- **Problem**: - [ ] `GenreBlender` struct field lacks godoc comment
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-131: Code Organization
- **Source**: `./pkg/procgen/item/AUDIT.md`
- **Location**: `(`generator.go:156-159`)`
- **Problem**: - [ ] Accessory templates use armor templates as fallback ; should have dedicated accessory templates for completeness
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-133: API consistency
- **Source**: `./pkg/procgen/legendary/AUDIT.md`
- **Location**: `(`manager.go:609`)`
- **Problem**: - [ ] `QuestManager.GetStatistics()` uses `qm.timeProvider.Now()` but returns `time.Time` directly; consider wrapping in a struct that includes TimeProvider reference for deterministic testing
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-135: Documentation
- **Source**: `./pkg/procgen/magic/AUDIT.md`
- **Problem**: - [ ] Package-level godoc in `doc.go` lacks a "See Also" section linking to related packages  that consume this generator
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-136: API Consistency
- **Source**: `./pkg/procgen/magic/AUDIT.md`
- **Location**: `(`types.go:229`, `balance.go:79`)`
- **Problem**: - [ ] `Spell.GetPowerLevel()` uses a different power calculation algorithm than `BalanceConfig.calculatePower()`, potentially causing confusion. Consider consolidating or documenting the difference
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-137: Code Duplication
- **Source**: `./pkg/procgen/minigame/AUDIT.md`
- **Problem**: - [ ] Name generation functions in `generator.go`  could be refactored to use a shared helper with genre-specific lookup tables to reduce duplication
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-138: Documentation
- **Source**: `./pkg/procgen/minigame/games/AUDIT.md`
- **Location**: `(`dice.go:202`, `puzzle.go:231`)`
- **Problem**: - [ ] `dice.go`, `puzzle.go` lack detailed godoc for reward calculation formulas
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-139: API Consistency
- **Source**: `./pkg/procgen/minigame/games/AUDIT.md`
- **Location**: `(`memory.go:142`, `hacking.go:224`, etc.)`
- **Problem**: - [ ] `Render()` deprecated but still present for backward compatibility; consider removal in V5.0 to reduce API surface
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-140: Missing godoc
- **Source**: `./pkg/procgen/narrative/AUDIT.md`
- **Location**: `(`generator.go:12-39`)`
- **Problem**: - [ ] `StoryArc` struct fields lack individual field documentation comments
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-141: Missing godoc
- **Source**: `./pkg/procgen/narrative/AUDIT.md`
- **Location**: `(`generator.go:42-66`)`
- **Problem**: - [ ] `PlotPoint` struct fields lack individual field documentation comments
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-144: Documentation
- **Source**: `./pkg/procgen/puzzle/AUDIT.md`
- **Location**: `(`solver.go:43`)`
- **Problem**: - [ ] CSP.NewCSP accepts seed parameter but does not use it for any deterministic behavior; either remove unused parameter or document why it exists
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-145: Documentation
- **Source**: `./pkg/procgen/puzzle/AUDIT.md`
- **Location**: `(`generator.go:63`)`
- **Problem**: - [ ] PuzzleElement.State uses interface{} type without documented valid types; add godoc listing expected types per ElementType
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-147: API Consistency
- **Source**: `./pkg/procgen/quest/AUDIT.md`
- **Location**: `(`types.go:217-249`)`
- **Problem**: - [ ] Quest methods have both procedural function variants `) and receiver methods `), creating API duplication. Recommend deprecating one pattern in favor of the other for consistency
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-148: Code Organization
- **Source**: `./pkg/procgen/quest/AUDIT.md`
- **Location**: `(`types.go:281-635`)`
- **Problem**: - [ ] Large genre template functions  could be refactored into data-driven template tables loaded from constants or embedded JSON
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-150: Template Hardcoding
- **Source**: `./pkg/procgen/recipe/AUDIT.md`
- **Location**: `(`generator.go:374-765`)`
- **Problem**: - [ ] Template registration uses hardcoded arrays rather than loading from external data files or mod system. This makes it difficult for mods to add new recipe templates without code changes. However...
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-153: Code Style
- **Source**: `./pkg/procgen/skills/AUDIT.md`
- **Problem**: - [ ] `normalizeGenre` function at `generator.go:125` is unexported but could be extracted to a validation helper in `pkg/procgen` for reuse by other generators
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-154: Documentation
- **Source**: `./pkg/procgen/station/AUDIT.md`
- **Location**: `(`generator.go:20-34`)`
- **Problem**: - [ ] Missing `doc.go` package overview for `StationType` enum explaining mapping to recipe types
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-158: No Serialization
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`generator.go:42-50`, `archaeology.go:31-58`, `branching.go:27-37`, `timeline.go:47-65`)`
- **Problem**: - [ ] None of the story types  implement `Serialize()/Deserialize()` methods. This means discovered stories, active narratives, excavation progress, and timelines **cannot persist across save/load** d...
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-160: Fragment Positioning Assumes 100x100 Map
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`generator.go:267-276`)`
- **Problem**: - [ ] `generateLocation()` hard-codes `10.0 + progress*80.0` positioning logic assuming 100x100 dungeon space. This breaks layout on non-square or differently-sized terrains generated by `pkg/procgen/...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-164: No Genre Fallback Warning
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`generator.go:153-168`)`
- **Problem**: - [ ] Genre-specific functions like `getThemesForGenre()` default to generic themes for unknown genres, but this happens silently. Should log a warning when an unsupported genre is provided so mod aut...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-165: Documentation
- **Source**: `./pkg/procgen/terrain/AUDIT.md`
- **Problem**: - [ ] cache.go uses time.Now() for LRU tracking , but this is explicitly documented as non-deterministic cache management only and does not affect terrain generation determinism
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-166: Documentation
- **Source**: `./pkg/procgen/terrain/AUDIT.md`
- **Problem**: - [ ] Point.Equals() method at point.go:20 lacks godoc comment explaining the equality comparison logic
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-168: Documentation
- **Source**: `./pkg/procgen/vehicle/AUDIT.md`
- **Location**: `(`templates.go:8, :87, :168, :202, :237`)`
- **Problem**: - [ ] `GetFantasyTemplates()` and other genre template functions lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-169: Documentation
- **Source**: `./pkg/procgen/vehicle/AUDIT.md`
- **Problem**: - [ ] Helper functions in `generator_helpers.go` lack individual godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-171: Documentation
- **Source**: `./pkg/rendering/cache/AUDIT.md`
- **Location**: `(`memory_monitor.go:142`, `memory_monitor.go:156`)`
- **Problem**: - [ ] `time.Now()` usage could be documented as non-deterministic but acceptable for monitoring
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-172: Performance
- **Source**: `./pkg/rendering/cache/AUDIT.md`
- **Location**: `(`pregenerator.go:134`)`
- **Problem**: - [ ] `GenerateAsync` could benefit from context.Context for cancellation
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-175: Mobile Platform Support
- **Source**: `./pkg/rendering/display/AUDIT.md`
- **Location**: `(`config.go:38-49`)`
- **Problem**: - [ ] No mobile-specific DPI scaling or safe area inset handling. Package supports 4 standard 16:9 resolutions but lacks custom resolution support for non-standard aspect ratios . Consider adding `Get...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-178: Test race detector
- **Source**: `./pkg/rendering/lighting/AUDIT.md`
- **Problem**: - [ ] No race detector tests run due to X11 dependency; recommend adding integration tests using `StubInput`/`StubInput` patterns where possible to achieve partial race coverage
- **Fix**:
  1. Add `sync.RWMutex` field to manager struct
  2. Add `mu.Lock()` / `defer mu.Unlock()` to mutating methods
  3. Add `mu.RLock()` / `defer mu.RUnlock()` to read methods
  4. Document thread-safety guarantees in godoc
- **Verify**: `go test -race ./pkg/... -v`

#### REM-179: Code organization
- **Source**: `./pkg/rendering/palette/AUDIT.md`
- **Location**: `(`utils.go:8-36`)`
- **Problem**: - [ ] `utils.go` contains helper functions  that could be shared across packages. Consider moving to a shared `pkg/utils` or `pkg/math` package to reduce duplication.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-180: API consistency
- **Source**: `./pkg/rendering/palette/AUDIT.md`
- **Location**: `(gradient.go:217)`
- **Problem**: - [ ] `CreateGradientPalette` function  does not follow the Generator pattern used by the rest of the package. Consider adding a method ` GenerateFromGradient *Palette` for consistency.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-181: Documentation
- **Source**: `./pkg/rendering/particles/AUDIT.md`
- **Location**: `(`behaviors.go:92`)`
- **Problem**: - [ ] Package-level doc.go is comprehensive  but individual complex functions like `ApplyPhysics` could benefit from inline algorithm comments explaining the physics formulas
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-182: Documentation
- **Source**: `./pkg/rendering/particles/AUDIT.md`
- **Location**: `(`weather.go:300-400` range)`
- **Problem**: - [ ] Weather system puddle accumulation and snow drift algorithms lack inline comments explaining the mathematical approach
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-184: Documentation
- **Source**: `./pkg/rendering/patterns/AUDIT.md`
- **Location**: `(`generator.go:99`)`
- **Problem**: - [ ] `applyGenreVariations` modifies config but doesn't document mutation  — The function mutates the input config's `Scale` and `DetailLevel` fields. Should document this side effect or return a mod...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-185: Code Organization
- **Source**: `./pkg/rendering/patterns/AUDIT.md`
- **Location**: `(`generator.go:281-397`)`
- **Problem**: - [ ] Helper methods in generator.go could be grouped by concern  — The pixel/color manipulation helpers  are interspersed. Consider grouping them together with a comment block for easier navigation.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-187: Integration
- **Source**: `./pkg/rendering/pool/AUDIT.md`
- **Location**: `(`image_pool.go:36-37`)`
- **Problem**: - [ ] Package not explicitly initialized in `cmd/client/main.go` startup path; only created on-demand in `handlers.go:671` during lazy system init. Consider documenting that pool is created per-render...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-189: Documentation
- **Source**: `./pkg/rendering/postprocess/AUDIT.md`
- **Location**: `(`chromatic_aberration.go:91`)`
- **Problem**: - [ ] Package `doc.go` is comprehensive, but exported functions in `chromatic_aberration.go:91`  lack detailed parameter documentation describing angle range and expected visual output
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-191: Component Serialization
- **Source**: `./pkg/rendering/quality/AUDIT.md`
- **Location**: `(`quality_settings_component.go:8-33`)`
- **Problem**: - [ ] `QualitySettingsComponent` lacks `Serialize()/Deserialize()` methods. If quality overrides should persist across save/load, implement ComponentSerializer interface. Current design treats quality...
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-195: Documentation
- **Source**: `./pkg/rendering/shapes/AUDIT.md`
- **Location**: `(types.go:107-120)`
- **Problem**: - [ ] ShapeEllipse, ShapeCapsule, ShapeBean, ShapeWedge, ShapeShield, ShapeBlade, ShapeSkull type comments missing in String() switch cases , reducing discoverability
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-199: Documentation
- **Source**: `./pkg/rendering/tiles/AUDIT.md`
- **Location**: `(`doc.go:13,17,22,28,44,56`)`
- **Problem**: - [ ] Package-level comments reference "Phase 16.2", "Phase 16.3", "Phase 47" which lack context for new developers
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-202: Package scope
- **Source**: `./pkg/rendering/ui/AUDIT.md`
- **Problem**: - [ ] Package does not implement ECS `System` or `Component` interfaces. This is intentional , but docs should clarify that this package provides helpers used *by* systems, not systems themselves.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-204: Documentation gap
- **Source**: `./pkg/saveload/AUDIT.md`
- **Problem**: - [ ] `SaveManager` struct itself lacks a godoc comment . Defined at `manager.go:24` and `storage_wasm.go:46` .
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-205: Migrator interface duplication
- **Source**: `./pkg/saveload/AUDIT.md`
- **Problem**: - [ ] `Migrator` interface is defined twice: once in `migrator.go:9-20` for desktop, once in `storage_wasm.go:29-41` for WASM. While functionally identical , this duplication risks drift. Consider usi...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-210: Time.Now usage
- **Source**: `./pkg/social/persistence/AUDIT.md`
- **Location**: `(`types.go:36`)`
- **Problem**: - [ ] `RealTimeProvider.Now()` uses `time.Now()` directly, which is acceptable for server-side metadata timestamps per project guidelines, but should be documented as such
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-216: Unstructured Logging
- **Source**: `./pkg/version/AUDIT.md`
- **Location**: `(`pkg/version/version.go:71`)`
- **Problem**: - [ ] PrintVersion() uses `fmt.Println` instead of structured logging with logrus. This prevents version checks from being logged with correlation IDs or filtered by log level
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-218: time.Now usage in benchmarking/profiling
- **Source**: `./pkg/visualtest/AUDIT.md`
- **Location**: `(`benchmark.go:89`, `memory.go:41,60,79`, `memory_test.go:267,272,282,287,304,309,317,322,367-371,388-391`)`
- **Problem**: - [ ] Package uses `time.Now()` extensively in `benchmark.go:89, 393, 406` and `memory.go:41, 60, 79`. This is **acceptable** for benchmarking/profiling infrastructure where wall-clock time measuremen...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-219: documentation completeness
- **Source**: `./pkg/visualtest/AUDIT.md`
- **Location**: `(e.g., `calculateSimilarity` in `snapshot.go`, `extractDominantColors` in `genre.go:220-250`)`
- **Problem**: - [ ] While package-level `doc.go` is comprehensive , some internal helper functions lack inline comments explaining their algorithms .
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-220: Code Quality
- **Source**: `./pkg/visualtest/parity/AUDIT.md`
- **Location**: `(`validator.go:329-337`)`
- **Problem**: - [ ] `maxUint8` helper function could be made more generic or moved to a utility package for reuse
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-225: Missing system registration
- **Source**: `./pkg/world/AUDIT.md`
- **Problem**: - [ ] No evidence of `ChunkLoaderSystem`, `ChunkModificationSystem`, `ChunkCompressionSystem`, or `economy.System` being registered in `cmd/client/` or `cmd/server/` system initialization. Systems def...
- **Fix**:
  1. Open `cmd/client/main.go` or `cmd/server/main.go`
  2. Add system to world: `world.AddSystem(NewMySystem(...))`
  3. Verify system appears in `world.Systems()` list
  4. Check system Update() is called each frame
- **Verify**: `grep -r 'AddSystem.*MySystem' ./cmd/`

#### REM-226: HousingUI input abstraction incomplete
- **Source**: `./pkg/world/AUDIT.md`
- **Location**: `(`housing/ui.go:16-25`)`
- **Problem**: - [ ] `MenuInputProvider` interface defined but `IsTouchOrMouseJustPressed()` and `GetTouchOrMousePosition()` methods not present in `pkg/engine/interfaces.go:InputProvider`. Mismatched interface defi...
- **Fix**:
  1. Replace `ebiten.IsKeyPressed()` calls with `inputProvider.IsMenuUpJustPressed()`
  2. Replace `inpututil.IsKeyJustPressed()` with corresponding InputProvider method
  3. Accept `InputProvider` interface in constructor
  4. Update system registration to pass engine.InputProvider instance
- **Verify**: `grep -r 'ebiten.IsKey' ./pkg/ | wc -l  # Should be 0`

#### REM-231: Time Abstraction
- **Source**: `./pkg/world/economy/AUDIT.md`
- **Problem**: - [ ] `types.go:20,135` - `RealTimeProvider.Now()` and `Listing.IsExpired()` use `time.Now()` directly. While the TimeProvider abstraction pattern handles this correctly , production code paths still ...
- **Fix**:
  1. Add `TimeProvider` interface parameter to constructor
  2. Replace `time.Now()` calls with `timeProvider.Now()`
  3. Use `RealTimeProvider{}` in production, `MockTimeProvider{}` in tests
  4. Verify tests can control time progression
- **Verify**: `grep -r 'time.Now()' ./pkg/ --exclude='*_test.go' | wc -l`

#### REM-233: Documentation
- **Source**: `./pkg/world/housing/AUDIT.md`
- **Problem**: - [ ] `BuildingSize` type and constants lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-238: Missing genre blending
- **Source**: `./pkg/world/raids/AUDIT.md`
- **Problem**: - [ ] Boss name generation supports single genres but not genre blending from `pkg/procgen/genre`. Add genre weight parameter to `BossNameGenerator` for hybrid themes .
- **Fix**:
  1. Import `pkg/procgen/genre` package
  2. Add `GenreWeights map[string]float64` parameter to generator
  3. Use `genre.BlendGenres()` to combine genre properties
  4. Update name/loot/theme generation to sample from blended genres
- **Verify**: `go test -run TestGenreBlend ./pkg/... -v`

#### REM-242: Direct Ebiten Input Calls
- **Source**: `./pkg/world/territory/AUDIT.md`
- **Location**: `(`pkg/engine/territory_ui.go:92-113`, `pkg/engine/interfaces.go:171-236`)`
- **Problem**: - [ ] `TerritoryUI` violates Coding Guideline #2  by calling `ebiten.IsKeyPressed`, `ebiten.KeyUp`, `ebiten.KeyDown`, `inpututil.IsKeyJustPressed` directly instead of using the `InputProvider` interfa...
- **Fix**:
  1. Replace `ebiten.IsKeyPressed()` calls with `inputProvider.IsMenuUpJustPressed()`
  2. Replace `inpututil.IsKeyJustPressed()` with corresponding InputProvider method
  3. Accept `InputProvider` interface in constructor
  4. Update system registration to pass engine.InputProvider instance
- **Verify**: `grep -r 'ebiten.IsKey' ./pkg/ | wc -l  # Should be 0`

#### REM-243: No Cross-Server Federation
- **Source**: `./pkg/world/territory/AUDIT.md`
- **Location**: `(`doc.go:14`)`
- **Problem**: - [ ] Package lacks integration with `pkg/network/federation/` for cross-server territory control and guild wars. Documentation claims "cross-server territory synchronization support"  but search resu...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-244: Hard-Coded Capture Radius
- **Source**: `./pkg/world/territory/AUDIT.md`
- **Location**: `(`manager.go:34`, `types.go:117-131`)`
- **Problem**: - [ ] Manager initializes `captureRadius: 50.0` on line 34 without constructor parameter, configuration field, or genre-based adjustment. This fixed value cannot adapt to:  map scale ,  genre ,  diffi...
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

### LOW

#### REM-008: Package-Level Globals
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `(`mobile.go:24-32`)`
- **Problem**: - [ ] Uses package globals  which complicate testing and prevent multi-instance scenarios. Should use struct-based state.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-009: No Performance Targets
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Problem**: - [ ] No mobile-specific performance monitoring or targets. Desktop targets  may not apply to mobile. Typical target: 30+ FPS on mid-range devices, <300MB.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-010: No Exported Test Helpers
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Problem**: - [ ] config subpackage has strong tests but no exported helpers for other packages needing seed/genre mocking.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-011: Missing Mobile UX Documentation
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `(`doc.go:1-56`)`
- **Problem**: - [ ] doc.go has build instructions but lacks mobile UX guidance: touch-first design, virtual control placement, battery optimization, screen size adaptation.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-012: No Mod Support Documentation
- **Source**: `./cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Problem**: - [ ] Unclear if mobile builds support mod loading. No documentation of file system permissions , mod directory location, or platform-specific mod loading.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-017: No Exported Test Helpers
- **Source**: `./cmd/mobile/AUDIT.md`
- **Problem**: - [ ] Config package has strong internal tests but no exported test utilities for other packages needing seed/genre mocking.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-018: No Mobile-Specific Docs
- **Source**: `./cmd/mobile/AUDIT.md`
- **Location**: `(`doc.go:1-56`)`
- **Problem**: - [ ] Package doc.go mentions build instructions but lacks mobile UX considerations .
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-019: No Performance Targets
- **Source**: `./cmd/mobile/AUDIT.md`
- **Problem**: - [ ] No mobile-specific performance targets . Desktop targets  may not apply to mobile.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-020: No Orientation Handling
- **Source**: `./cmd/mobile/AUDIT.md`
- **Problem**: - [ ] No landscape/portrait detection or safe area insets handling .
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-021: Global Package Variables
- **Source**: `./cmd/mobile/AUDIT.md`
- **Location**: `(`mobile.go:24-32`)`
- **Problem**: - [ ] Uses package-level globals  which complicate testing and multi-instance scenarios.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-025: Documentation
- **Source**: `./cmd/server/AUDIT.md`
- **Problem**: - [ ] `entity_spawning.go` lacks package-level doc comment explaining server-side vs client-side spawning differences
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-026: Documentation
- **Source**: `./cmd/server/AUDIT.md`
- **Problem**: - [ ] `system_wrappers.go` lacks explanation of why wrappers are needed
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-027: Documentation
- **Source**: `./cmd/server/AUDIT.md`
- **Problem**: - [ ] `v9_validation.go` has good doc comments but could benefit from examples of validation scenarios
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-028: Code Organization
- **Source**: `./cmd/server/AUDIT.md`
- **Problem**: - [ ] `main.go` is 1,139 LOC; consider extracting initialization functions to `init_*.go` files following the pattern used in cmd/client
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-037: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] Sub-package `AUDIT.md` exists and follows the template exactly
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-038: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] Every issue in `AUDIT.md` has a `file.go:LINE` citation
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-039: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] Root `AUDIT.md` updated with checked entry including status, issue counts, and coverage
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-040: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] `go vet` passes on audited package
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-041: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] `go test` runs without failures
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-042: uncategorized
- **Source**: `./docs/META_AUDIT.md`
- **Problem**: - [ ] Input integration table is fully populated for all 6 input sources
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-050: Missing context cancellation check
- **Source**: `./pkg/balance/AUDIT.md`
- **Location**: `(`economic.go:248-260`)`
- **Problem**: - [ ] `validateGoldBalance()` in `economic.go` does not check `ctx.Done()` in the two simulation loops . All other validators properly check context in loops.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-051: Hardcoded stat budget comment
- **Source**: `./pkg/balance/AUDIT.md`
- **Location**: `(`combat.go:264-266`)`
- **Problem**: - [ ] Combat validator has hardcoded comment "Total stat budget ~165 per class" but no validation that sum equals this budget. Consider adding assertion or removing comment to avoid drift.
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-057: Enhancement
- **Source**: `./pkg/config/AUDIT.md`
- **Problem**: - [ ] Consider adding validation benchmarks to README.md "Related Packages" section to demonstrate performance characteristics.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-061: Memory Management
- **Source**: `./pkg/integration/choice_consequences/AUDIT.md`
- **Location**: `(`choice_tracker.go:545`)`
- **Problem**: - [ ] `CompanionReactions` slice in `PlayerState` hard-coded to keep last 20 reactions. Consider making this configurable via constructor option similar to `npcMemoryLimit` and `choiceLimit`.
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-065: Documentation
- **Source**: `./pkg/integration/companion_housing/AUDIT.md`
- **Location**: `(`types.go:1-23`)`
- **Problem**: - [ ] Package comment in `types.go`  duplicates information from `doc.go`. Consider making `types.go` comment more concise and pointing to `doc.go` for full overview.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-075: Documentation
- **Source**: `./pkg/integration/housing_crafting/AUDIT.md`
- **Location**: `(`types.go:22-43`)`
- **Problem**: - [ ] Type `QualityTier` and `StationType` enum String() methods don't document invalid case handling
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-079: Mod Compatibility
- **Source**: `./pkg/integration/narrative_world/AUDIT.md`
- **Problem**: - [ ] No integration with pkg/modding for customizable quest templates, conflict probabilities, or consequence rules.
- **Fix**:
  1. Import `pkg/modding` package
  2. Add `ModRuleProvider` field to manager struct
  3. Query mod rules for values: `modProvider.GetFloat('raid.difficulty_multiplier', default)`
  4. Register mod hooks in manager constructor
- **Verify**: `grep -r 'ModRuleProvider' ./pkg/world/raids/ | wc -l`

#### REM-086: Optimization
- **Source**: `./pkg/integration/trade_routes/AUDIT.md`
- **Location**: `(`manager.go:724-748`)`
- **Problem**: - [ ] `completeRoute()` iterates cargo twice: once to calculate profit, once to apply price impacts. Consider single-pass iteration for efficiency.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-090: Code Style
- **Source**: `./pkg/integration/world_events/AUDIT.md`
- **Location**: `(`types.go:181`)`
- **Problem**: - [ ] `EventManagerConfig.DefaultEventManagerConfig()` should be `NewDefaultEventManagerConfig()` to follow Go naming conventions for constructor functions
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-091: Documentation
- **Source**: `./pkg/integration/world_events/AUDIT.md`
- **Location**: `(`types.go:68-71`)`
- **Problem**: - [ ] `WorldEvent.CenterX` and `WorldEvent.CenterY` fields have inline comments but should be promoted to full godoc format for consistency
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-105: Documentation
- **Source**: `./pkg/network/federation/mobile/AUDIT.md`
- **Location**: `(`types.go:119`)`
- **Problem**: - [ ] `State.bytesAvailable` is unexported but used for public bandwidth limiting; consider adding comment explaining it's internal token bucket state
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-106: Code Style
- **Source**: `./pkg/network/federation/mobile/AUDIT.md`
- **Location**: `(`types.go:119`)`
- **Problem**: - [ ] `State.timeProvider` field comment could mention it defaults to real system time if nil
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-108: Channel Capacity Overflow
- **Source**: `./pkg/network/federation/webrtc/AUDIT.md`
- **Location**: `(`signaling.go:367-377`)`
- **Problem**: - [ ] `signaling.go:367,374` round-robin counter could theoretically overflow on very long-lived servers , though code at line 374-376 has overflow protection. Document the protection logic.
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-122: Documentation
- **Source**: `./pkg/procgen/entity/AUDIT.md`
- **Location**: `(`doc.go:1-40`)`
- **Problem**: - [ ] Package-level godoc in `doc.go` could mention integration with `pkg/engine/entity_spawning.go` and `merchant_spawn.go`
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-125: Test Assertion
- **Source**: `./pkg/procgen/faction/AUDIT.md`
- **Location**: `(`generator_test.go:446`)`
- **Problem**: - [ ] TestGenerator_SpecialRelationships uses t.Logf instead of assertion for corp-vs-rebel enemy check, noting randomness prevents guarantee. Consider using a fixed seed that produces the special rel...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-130: Documentation
- **Source**: `./pkg/procgen/genre/AUDIT.md`
- **Location**: `(`blender.go:95-104`)`
- **Problem**: - [ ] `PresetBlends()` return type could be extracted to named type for better godoc
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-132: Documentation
- **Source**: `./pkg/procgen/item/AUDIT.md`
- **Problem**: - [ ] Template file `templates.go` is 34KB and too large to view at once; consider splitting into genre-specific files for maintainability
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-134: Code organization
- **Source**: `./pkg/procgen/legendary/AUDIT.md`
- **Location**: `(`types.go:238-322`)`
- **Problem**: - [ ] Helper functions like `countKillProgress`, `countCollectionProgress`, etc. in `types.go:238-322` could be extracted to a separate progress calculation utility for better testability and reusabil...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-142: Missing godoc
- **Source**: `./pkg/procgen/narrative/AUDIT.md`
- **Location**: `(`generator.go:69-81`)`
- **Problem**: - [ ] `PlayerChoice` struct fields lack individual field documentation comments
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-143: Minor validation
- **Source**: `./pkg/procgen/narrative/AUDIT.md`
- **Location**: `(`generator.go:276`)`
- **Problem**: - [ ] `generateTitle()` returns "Untitled Story" for unknown genres, but this is never validated. Consider logging warning when falling back to default
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-146: Code Style
- **Source**: `./pkg/procgen/puzzle/AUDIT.md`
- **Location**: `(`generator.go:604-663`)`
- **Problem**: - [ ] Multiple helper functions  are only used by generateColorMatchingPuzzle; consider inlining or moving to closure for locality
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-151: Material Quantity Range
- **Source**: `./pkg/procgen/recipe/AUDIT.md`
- **Location**: `(`generator.go:189`)`
- **Problem**: - [ ] Material quantities are hardcoded to 1-3 per material `), with no template-level control. Consider making quantity range part of RecipeTemplate for more flexibility.
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-162: Sparse Godoc on Helper Functions
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`pkg/engine/story_fragment_component.go:97-161`)`
- **Problem**: - [ ] ECS-compliant helper functions like `JournalAddDiscovery`, `JournalIsDiscovered`, etc. lack individual godoc comments explaining parameters and return values. Only the deprecated methods have fu...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-163: Vector2 Type Duplication
- **Source**: `./pkg/procgen/story/AUDIT.md`
- **Location**: `(`types.go:18-26`)`
- **Problem**: - [ ] `Vector2` is defined locally in `types.go:18-26` but `pkg/engine/components.go` likely has an equivalent type for positions. This creates coupling and conversion overhead. Consider importing eng...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-167: Documentation
- **Source**: `./pkg/procgen/terrain/AUDIT.md`
- **Location**: `(types.go:310, 315)`
- **Problem**: - [ ] Room.Center() and Room.Overlaps() methods  lack godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-170: Documentation
- **Source**: `./pkg/procgen/vehicle/AUDIT.md`
- **Problem**: - [ ] Combat generation functions in `generator_combat.go` lack individual godoc comments
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-173: Documentation
- **Source**: `./pkg/rendering/cache/AUDIT.md`
- **Location**: `(`predictive_warmer.go:65-73`)`
- **Problem**: - [ ] `PredictiveWarmerConfig` validation could be more explicit about clamping to defaults
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-174: Code Quality
- **Source**: `./pkg/rendering/cache/AUDIT.md`
- **Location**: `(`predictive_warmer.go:162`)`
- **Problem**: - [ ] `predictNextLocked` comment could clarify RLock requirement
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-186: Performance
- **Source**: `./pkg/rendering/patterns/AUDIT.md`
- **Location**: `(`generator.go:465-468`)`
- **Problem**: - [ ] `cellularNoise` calculates hash twice per cell  — The hash calculation ` & 0x7FFFFFFF` is done once, then bit-shifted twice. Could be micro-optimized by pre-computing both offsets in one pass, t...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-192: Concurrency
- **Source**: `./pkg/rendering/quality/AUDIT.md`
- **Location**: `(`auto_adjuster.go:56-82`)`
- **Problem**: - [ ] `AutoAdjuster.Update()` holds write lock during entire update including callback invocation. If callback is slow, this could block updates. Consider releasing lock before callback or document th...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-196: Performance
- **Source**: `./pkg/rendering/shapes/AUDIT.md`
- **Location**: `(`generator.go:216, 371, 446`)`
- **Problem**: - [ ] Some shape algorithms use repeated math.Sqrt/math.Pow which could be cached for hot-path optimization
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-198: Documentation verbosity
- **Source**: `./pkg/rendering/sprites/AUDIT.md`
- **Problem**: - [ ] `doc.go` is 285 lines, very comprehensive but may benefit from splitting into separate markdown docs for maintainability
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-200: Integration
- **Source**: `./pkg/rendering/tiles/AUDIT.md`
- **Problem**: - [ ] Package is not imported by any engine, client, or server files - appears to be unused dead code or awaiting integration
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-206: WASM migration unsupported
- **Source**: `./pkg/saveload/AUDIT.md`
- **Problem**: - [ ] WASM version explicitly does not support save file migration . Documented in `storage_wasm.go:97-99` and `storage_wasm.go:50` comment. This is a known limitation but may surprise users upgrading...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-207: MemorySaveManager.SetMigrator is no-op
- **Source**: `./pkg/saveload/AUDIT.md`
- **Problem**: - [ ] `memory_manager.go:358` has `SetMigrator {}` as empty method . While in-memory storage doesn't persist across sessions, accepting but ignoring the parameter could confuse API users expecting mig...
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-211: Documentation
- **Source**: `./pkg/social/persistence/AUDIT.md`
- **Problem**: - [ ] Package doc.go is excellent but truncated in `go doc` output; consider adding cross-references to integration points
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-212: Resource management
- **Source**: `./pkg/social/persistence/AUDIT.md`
- **Location**: `(`trust_manager.go:241-259`)`
- **Problem**: - [ ] `TrustManager.StartAutomaticDecay()` creates a goroutine that runs indefinitely; ensure server shutdown calls `StopAutomaticDecay()` to prevent goroutine leak
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-213: API consistency
- **Source**: `./pkg/social/persistence/AUDIT.md`
- **Location**: `(`image_gallery.go:371-382`)`
- **Problem**: - [ ] `ImageGallery.GetThumbnails()` returns a new slice type `ImageThumbnail` not defined in types.go; consider moving to types.go for consistency
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-217: Missing Godoc
- **Source**: `./pkg/version/AUDIT.md`
- **Location**: `(`pkg/version/version.go:22-29`)`
- **Problem**: - [ ] Constants Major, Minor, Patch lack individual godoc comments explaining when to increment each
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-227: Incomplete territory doc.go
- **Source**: `./pkg/world/AUDIT.md`
- **Problem**: - [ ] `pkg/world/territory/doc.go` missing; only README.md exists. Per guideline, every package should have doc.go with package-level overview.
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-228: GuildHallManager persistence not documented
- **Source**: `./pkg/world/AUDIT.md`
- **Problem**: - [ ] `GuildHallManager` has no documented `Serialize()`/`Deserialize()` methods or integration with `pkg/saveload`. Guild halls may not persist across save/load cycles.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to struct
  2. Add `Deserialize([]byte) error` method to struct
  3. Implement `ComponentSerializer` interface from `pkg/saveload`
  4. Register with save/load manager in system initialization
- **Verify**: `go test -run TestSerialize ./pkg/... -v`

#### REM-229: Raid instance lifecycle unclear
- **Source**: `./pkg/world/AUDIT.md`
- **Problem**: - [ ] `raids.Instance` has `IsActive()` and `IsComplete()` methods but no documented cleanup/removal path. Memory leak potential if completed instances are never garbage collected.
- **Fix**:
  1. Add `Cleanup()` method to remove completed/expired instances
  2. Call cleanup in system Update() or dedicated timer goroutine
  3. Add TTL field to track instance lifetime
  4. Use `delete(map, key)` to remove entries explicitly
- **Verify**: `go test -run TestCleanup ./pkg/... -v`

#### REM-230: Territory manager not threadsafe
- **Source**: `./pkg/world/AUDIT.md`
- **Location**: `(`territory.go:34`)`
- **Problem**: - [ ] `TerritoryManager` has no `sync.Mutex` or documented thread-safety guarantees despite being accessed from network handlers and game systems concurrently. Potential data race.
- **Fix**:
  1. Add `sync.RWMutex` field to manager struct
  2. Add `mu.Lock()` / `defer mu.Unlock()` to mutating methods
  3. Add `mu.RLock()` / `defer mu.RUnlock()` to read methods
  4. Document thread-safety guarantees in godoc
- **Verify**: `go test -race ./pkg/... -v`

#### REM-234: Documentation
- **Source**: `./pkg/world/housing/AUDIT.md`
- **Problem**: - [ ] `Vector2` type lacks godoc comment
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

#### REM-235: API Consistency
- **Source**: `./pkg/world/housing/AUDIT.md`
- **Location**: `(`manager.go:177`)`
- **Problem**: - [ ] `CreateHouse` method accepts `interface{}` for buildingData instead of typed struct
- **Fix**:
  1. Review the finding description and locate affected code
  2. Apply minimal fix following project coding guidelines
  3. Add test coverage for the fixed behavior
  4. Verify no regressions in existing tests
- **Verify**: `go test ./pkg/... -v`

#### REM-240: Hardcoded genre fallback
- **Source**: `./pkg/world/raids/AUDIT.md`
- **Location**: `(`manager.go:40,89`)`
- **Problem**: - [ ] `Manager.GenerateRaid()` and `Manager.CreateInstance()` hardcode `GenreID: "fantasy"` instead of reading from world config . Pass genre as parameter or store in Manager struct.
- **Fix**:
  1. Add field to config struct or constructor parameter
  2. Remove hard-coded value and use field instead
  3. Update call sites to pass appropriate value
  4. Add genre-based or difficulty-based scaling if applicable
- **Verify**: `go test ./pkg/... -v`

#### REM-245: Defensive Copy Overhead
- **Source**: `./pkg/world/territory/AUDIT.md`
- **Location**: `(`manager.go:71-83,483-492`, `siege.go:381-393,427-444`)`
- **Problem**: - [ ] All 6 getter methods return deep copies via `copyTerritory()`  and `copySiege()` . While thread-safe, this creates allocation pressure: UI rendering 100 territories at 60 FPS = `100×60 = 6000` a...
- **Fix**:
  1. Add `sync.RWMutex` field to manager struct
  2. Add `mu.Lock()` / `defer mu.Unlock()` to mutating methods
  3. Add `mu.RLock()` / `defer mu.RUnlock()` to read methods
  4. Document thread-safety guarantees in godoc
- **Verify**: `go test -race ./pkg/... -v`

#### REM-246: Missing Godoc Examples
- **Source**: `./pkg/world/territory/AUDIT.md`
- **Location**: `(`doc.go:1-106`)`
- **Problem**: - [ ] Package `doc.go` has comprehensive prose documentation  but lacks runnable `Example*` test functions. `go doc` output shows no examples. Would benefit from: `ExampleManager_CreateTerritory()` , ...
- **Fix**:
  1. Create or update `doc.go` with package overview
  2. Add godoc comments to all exported types and functions
  3. Include usage examples in package documentation
  4. Run `go doc <package>` to verify output
- **Verify**: `go doc ./pkg/... | grep -c 'Package'`

## Completion Criteria
- [ ] All REM-### items implemented
- [ ] All verification commands pass
- [ ] No `- [ ]` items remain in any *AUDIT*.md (excluding test-coverage findings)
- [ ] Run full test suite: `go test ./... -v`
- [ ] Run race detector: `go test -race ./pkg/...`
- [ ] Build all targets: `make build`