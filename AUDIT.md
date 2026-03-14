# Codebase Audit Remediation Plan

**Generated**: 2026-02-27
**Scope**: All `*AUDIT*.md` files in repository (75 files with unchecked items across 109 total audit files)
**Total Unresolved Findings**: 168
**Exclusions Applied**: Test coverage percentage findings, missing test findings, test infrastructure findings, coverage tooling findings

## Summary by Severity

| Severity | Count |
|----------|-------|
| Critical | 8     |
| High     | 32    |
| Medium   | 79    |
| Low      | 49    |

---

## Findings

### CRITICAL

#### REM-001: Territory manager lacks thread safety
- **Source**: `pkg/world/AUDIT.md`
- **Location**: `pkg/world/territory/manager.go:34`
- **Problem**: `TerritoryManager` has no `sync.Mutex` or documented thread-safety guarantees despite being accessed from network handlers and game systems concurrently, creating potential data races.
- **Fix**:
  1. Add a `sync.RWMutex` field to `TerritoryManager` struct in `pkg/world/territory/manager.go`
  2. Wrap all read methods (`GetTerritory`, `GetAllTerritories`, `GetSiege`, etc.) with `mu.RLock()`/`mu.RUnlock()`
  3. Wrap all write methods (`CreateTerritory`, `UpdateCaptureProgress`, `DeclareWar`, etc.) with `mu.Lock()`/`mu.Unlock()`
  4. Follow the same pattern used in `pkg/network/federation/guild/` which already uses `sync.RWMutex`
- **Verify**: `go test -race ./pkg/world/territory/...`

#### REM-002: Economy system ECS interface non-compliance
- **Source**: `pkg/world/AUDIT.md`
- **Location**: `pkg/world/economy/system.go:44`
- **Problem**: `economy.System.Update(deltaTime float64)` does not match `engine.System.Update(entities []*Entity, deltaTime float64)` signature from `pkg/engine/interfaces.go:35`, preventing standard ECS registration.
- **Fix**:
  1. Update `economy.System.Update` signature in `pkg/world/economy/system.go` to `Update(entities []*engine.Entity, deltaTime float64)`
  2. Filter entities for economy-related components at the start of the method
  3. Alternatively, create an adapter wrapper in `cmd/server/` that bridges the signature mismatch (similar to `system_wrappers.go` pattern)
- **Verify**: `go vet ./pkg/world/economy/...`

#### REM-003: Missing world system registrations in game loop
- **Source**: `pkg/world/AUDIT.md`
- **Location**: `cmd/client/`, `cmd/server/` (system initialization files)
- **Problem**: No evidence of `ChunkLoaderSystem`, `ChunkModificationSystem`, `ChunkCompressionSystem`, or `economy.System` being registered in client or server system initialization. Systems are defined but not wired into the game loop.
- **Fix**:
  1. In `cmd/server/` system initialization (e.g., `v9_systems.go`), add registration calls for each missing system
  2. For `ChunkLoaderSystem`: `world.AddSystem(NewChunkLoaderSystem(...))`
  3. For `economy.System`: Create adapter wrapper in `cmd/server/system_wrappers.go` if signature mismatch exists (see REM-002), then register
  4. Verify each system's constructor parameters match available server state
- **Verify**: `grep -rn "ChunkLoader\|ChunkModification\|ChunkCompression\|economy.System\|economy.New" cmd/server/ cmd/client/`

#### REM-004: Territory UI uses direct Ebiten input calls instead of InputProvider
- **Source**: `pkg/world/territory/AUDIT.md`
- **Location**: `pkg/engine/territory_ui.go:92-113`
- **Problem**: `TerritoryUI` calls `ebiten.IsKeyPressed`, `inpututil.IsKeyJustPressed` directly instead of using the `InputProvider` interface, breaking testability, input rebinding, gamepad support, and mobile touch input.
- **Fix**:
  1. Add an `InputProvider` field to the `TerritoryUI` struct
  2. Accept `InputProvider` as a constructor parameter
  3. Replace `inpututil.IsKeyJustPressed(ebiten.KeyUp)` at line 94 with `input.IsMenuUpJustPressed()`
  4. Replace `ebiten.KeyDown` at line 100 with `input.IsMenuDownJustPressed()`
  5. Replace `ebiten.KeyW` at line 105 with `input.IsMenuConfirmJustPressed()`
  6. Replace `ebiten.KeyB` at line 110 with `input.IsMenuBackJustPressed()`
- **Verify**: `go vet ./pkg/engine/... && grep -n "ebiten.IsKeyPressed\|inpututil.IsKeyJustPressed" pkg/engine/territory_ui.go`

#### REM-005: Territory persistence completely missing
- **Source**: `pkg/world/territory/AUDIT.md`
- **Location**: `pkg/world/territory/manager.go:13-20`, `siege.go:73-100`, `types.go:56-67,92-103,106-114`
- **Problem**: Manager, Siege, Territory, WarDeclaration, DefensiveStructure structs lack `Serialize()`/`Deserialize()` methods. All territory state (ownership, capture progress, wars, structures, sieges) is lost on server restart.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` and `Deserialize(data []byte) error` methods to `Territory` struct in `types.go`
  2. Add same methods to `Siege`, `WarDeclaration`, and `DefensiveStructure` structs
  3. Use `encoding/json` for serialization (consistent with other packages)
  4. Add `Save(path string) error` and `Load(path string) error` methods to `Manager`
  5. Wire into `pkg/saveload/manager.go` save/load lifecycle
- **Verify**: `grep -n "Serialize\|Deserialize" pkg/world/territory/*.go`

#### REM-006: Raid instance persistence missing
- **Source**: `pkg/world/raids/AUDIT.md`
- **Location**: `pkg/world/raids/` (instance.go, types)
- **Problem**: `RaidInstance` and `PlayerLockout` types do not implement `Serialize()`/`Deserialize()` for save/load support. Instances and lockouts are lost on server restart.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to `RaidInstance` using `encoding/json`
  2. Add `Deserialize(data []byte) error` method to `RaidInstance`
  3. Add same methods to `PlayerLockout`
  4. Add `Save(path string) error` and `Load(path string) error` to the raid `Manager`
  5. Wire into `pkg/saveload/manager.go` persistence lifecycle
- **Verify**: `grep -n "Serialize\|Deserialize" pkg/world/raids/*.go`

#### REM-007: Story types lack serialization despite documentation claiming persistence
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/generator.go:42-50`, `archaeology.go:31-58`, `branching.go:27-37`, `timeline.go:47-65`
- **Problem**: None of the story types (`StorySequence`, `BranchingNarrative`, `ArchaeologicalSite`, `Timeline`, `CrossDungeonStory`) implement `Serialize()`/`Deserialize()`. Discovered stories, active narratives, and excavation progress cannot persist across save/load.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` and `Deserialize(data []byte) error` to each type using `encoding/json`
  2. For `StorySequence` in `generator.go`, add JSON struct tags to all fields
  3. For `ArchaeologicalSite` in `archaeology.go`, serialize excavated layers state
  4. For `BranchingNarrative` in `branching.go`, serialize current path and choices made
  5. For `Timeline` in `timeline.go`, serialize events and periods
  6. For `CrossDungeonStory` in `crossdungeon.go`, serialize fragment accessibility state
- **Verify**: `grep -n "func.*Serialize\|func.*Deserialize" pkg/procgen/story/*.go`

#### REM-008: Guild housing component missing serialization
- **Source**: `pkg/integration/guild_housing/AUDIT.md`
- **Location**: `pkg/integration/guild_housing/types.go:57-66`
- **Problem**: `GuildHousingComponent` does not implement `Serialize()/Deserialize()` required by `ComponentSerializer` interface. Manager has `Save()/Load()` but component persistence is not wired.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` method to `GuildHousingComponent` in `types.go` using `encoding/json`
  2. Add `Deserialize(data []byte) error` method
  3. Register the component with the save/load system's component serializer registry
- **Verify**: `grep -n "Serialize\|Deserialize" pkg/integration/guild_housing/types.go`

---

### HIGH

#### REM-009: 20+ time.Now() calls in world package violate determinism guideline
- **Source**: `pkg/world/AUDIT.md`
- **Location**: `chunk_modification.go:102`, `territory.go:61,63,72,108,194`, `persistence.go:102,218`, `metagame.go:85,86,109,110,136,137,160,161`
- **Problem**: Production code uses `time.Now()` directly in ~20 locations instead of injectable time provider, violating deterministic generation guideline and preventing time-controlled testing.
- **Fix**:
  1. Define a `TimeProvider` interface in `pkg/world/types.go`: `type TimeProvider interface { Now() time.Time }`
  2. Add a `timeProvider TimeProvider` field to each struct that calls `time.Now()`
  3. Accept `TimeProvider` in constructors; default to `RealTimeProvider{}`
  4. Replace all `time.Now()` calls with `s.timeProvider.Now()`
  5. Follow the pattern already used in `pkg/integration/world_events/` and `pkg/hostplay/`
- **Verify**: `grep -rn "time.Now()" pkg/world/*.go pkg/world/territory/*.go pkg/world/economy/*.go | grep -v _test.go | grep -v doc.go`

#### REM-010: Network package hardcodes time.Now() instead of GameClock injection
- **Source**: `pkg/network/AUDIT.md`
- **Location**: `bandwidth.go:41,54,67,101,132,158,169,177`, `chat.go:18,189,200,231,232,257,271,318,355`, `client.go:253,254,495`, `server.go`
- **Problem**: `TCPClient`, `TCPServer`, `ChatManager`, and `BandwidthMonitor` hardcode `time.Now()` (~20+ occurrences) instead of accepting injectable `GameClock` interface, preventing deterministic testing and replay.
- **Fix**:
  1. Add `gameClock engine.GameClock` field to `TCPClient`, `TCPServer`, `ChatManager`, `BandwidthMonitor`
  2. Accept `GameClock` as optional constructor parameter (nil = default `RealGameClock`)
  3. Replace `time.Now()` with `s.gameClock.Now()` in each occurrence
  4. Use `GameClock` interface from `pkg/engine/interfaces.go`
- **Verify**: `grep -rn "time.Now()" pkg/network/*.go | grep -v _test.go | wc -l`

#### REM-011: HousingUI takes concrete *ebiten.Image instead of interface
- **Source**: `pkg/world/AUDIT.md`
- **Location**: `pkg/world/housing/ui.go:158`
- **Problem**: `HousingUI.Draw(screen *ebiten.Image)` takes concrete type instead of `Renderer` interface, making testing without Ebiten runtime impossible.
- **Fix**:
  1. Define or use an existing `Renderer` interface that wraps `*ebiten.Image` drawing methods
  2. Change `Draw` signature to accept the interface type
  3. Alternatively, if other UI systems use `*ebiten.Image` directly, document this as an accepted pattern and mark finding as wontfix
- **Verify**: `grep -n "func.*Draw.*ebiten.Image" pkg/world/housing/ui.go`

#### REM-012: HousingUI InputProvider interface mismatch
- **Source**: `pkg/world/AUDIT.md`
- **Location**: `pkg/world/housing/ui.go:16-25`
- **Problem**: `MenuInputProvider` interface defines `IsTouchOrMouseJustPressed()` and `GetTouchOrMousePosition()` not present in `pkg/engine/interfaces.go:InputProvider`, creating an integration gap.
- **Fix**:
  1. Add `IsTouchOrMouseJustPressed() bool` and `GetTouchOrMousePosition() (int, int)` to `InputProvider` interface in `pkg/engine/interfaces.go`
  2. Implement these methods in the concrete `InputProvider` implementation
  3. Alternatively, have `HousingUI` accept `engine.InputProvider` and compose the touch/mouse check from existing methods
- **Verify**: `grep -n "IsTouchOrMouseJustPressed\|GetTouchOrMousePosition" pkg/engine/interfaces.go`

#### REM-013: Guild housing UI not integrated
- **Source**: `pkg/integration/guild_housing/AUDIT.md`
- **Location**: `pkg/engine/guild_ui.go:1-200`, `pkg/integration/guild_housing/guild_housing_manager.go:1-716`
- **Problem**: GuildUI displays guild info, members, and treasury but has no housing tab, permissions UI, or storage UI. Guild housing manager methods are never called from the UI.
- **Fix**:
  1. Add a "Housing" tab to `GuildUI` in `pkg/engine/guild_ui.go`
  2. Wire tab to call `guild_housing_manager.GetGuildHouse()`, `GetPermissions()`, `GetStorage()`
  3. Add UI rendering for housing status, permission toggles, and storage grid
  4. Follow existing tab pattern used for members/treasury tabs
- **Verify**: `grep -n "housing\|Housing" pkg/engine/guild_ui.go`

#### REM-014: GuildHousingComponent defined but never attached to entities
- **Source**: `pkg/integration/guild_housing/AUDIT.md`
- **Location**: `pkg/integration/guild_housing/types.go:57-66`
- **Problem**: `GuildHousingComponent` exists for ECS integration but is never attached to entities by any engine code.
- **Fix**:
  1. In the guild housing initialization code (e.g., `cmd/server/v9_systems.go`), attach `GuildHousingComponent` to guild entities when guild houses are created
  2. Ensure `AddComponent` is called with the component when a guild acquires a house
  3. Update the guild housing system to query entities with this component
- **Verify**: `grep -rn "GuildHousingComponent" cmd/ pkg/engine/`

#### REM-015: Story generators never instantiated (60% dead code)
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/archaeology.go:60-398`, `timeline.go:66-622`, `crossdungeon.go:44-430`
- **Problem**: `ArchaeologyGenerator`, `TimelineGenerator`, and `CrossDungeonGenerator` are fully implemented but never called in `cmd/client/`, `cmd/server/`, or `pkg/engine/`. ~60% of the package is dead code from the player's perspective.
- **Fix**:
  1. In `cmd/client/handlers.go` or `cmd/server/` initialization, instantiate each generator
  2. Wire `ArchaeologyGenerator` to terrain generation for placing archaeology sites
  3. Wire `TimelineGenerator` to world creation for generating world history
  4. Wire `CrossDungeonGenerator` to dungeon generation for cross-dungeon narratives
  5. Add ECS components for each type (see REM-016)
- **Verify**: `grep -rn "ArchaeologyGenerator\|TimelineGenerator\|CrossDungeonGenerator" cmd/ pkg/engine/`

#### REM-016: Missing ECS components for story types
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/archaeology.go:31-58`, `timeline.go:47-65`, `crossdungeon.go:26-38`
- **Problem**: No engine components exist for `ArchaeologicalSite`, `Timeline`, or `CrossDungeonStory`, meaning generators' output has no ECS representation.
- **Fix**:
  1. Create `ArchaeologicalSiteComponent` in `pkg/engine/` with `Type() string` returning `"archaeological_site"`
  2. Create `TimelineComponent` with `Type() string` returning `"timeline"`
  3. Create `CrossDungeonStoryComponent` with `Type() string` returning `"cross_dungeon_story"`
  4. Each component should be a pure data struct wrapping the corresponding procgen type
  5. Follow pattern of existing `StoryFragmentComponent` in `pkg/engine/story_fragment_component.go`
- **Verify**: `grep -rn "ArchaeologicalSiteComponent\|TimelineComponent\|CrossDungeonStoryComponent" pkg/engine/`

#### REM-017: StoryJournalUI never instantiated in client
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/rendering/ui/story_journal.go:1-232`
- **Problem**: `StoryJournalUI` is fully implemented (232 LOC) with series list, fragment navigation, and genre theming, but is never instantiated in `cmd/client`. Players can discover fragments but have no way to review them.
- **Fix**:
  1. In `cmd/client/handlers.go` or UI initialization, instantiate `StoryJournalUI`
  2. Add a hotkey or menu option to open the story journal (e.g., "J" key)
  3. Wire the input handler to toggle journal visibility
  4. Connect journal to the existing `DiscoverySystem` fragment data
- **Verify**: `grep -rn "StoryJournalUI" cmd/client/`

#### REM-018: Class generator instantiated but never called
- **Source**: `pkg/procgen/class/AUDIT.md`
- **Location**: `pkg/procgen/class/generator.go:1`, `cmd/client/handlers.go:52`, `pkg/engine/class_progression_component.go:189`
- **Problem**: `ClassGenerator` is instantiated in `cmd/client/handlers.go` but never invoked. All class data is hardcoded in `pkg/engine/class_progression_component.go` with static switch statements instead of using the generator.
- **Fix**:
  1. In character creation flow (`pkg/engine/character_creation.go`), call `classGenerator.Generate(seed, params)` to produce class data
  2. Replace hardcoded `GetClassAbilities()` and `GetAvailableSpecializations()` calls with generator output
  3. Or if hardcoded data is intentional, remove the unused generator instantiation from `handlers.go` to eliminate dead code
- **Verify**: `grep -rn "classGenerator" cmd/client/handlers.go`

#### REM-019: Social persistence not wired into server initialization
- **Source**: `pkg/social/persistence/AUDIT.md`
- **Location**: `cmd/server/v8_systems.go` (TODO comments)
- **Problem**: `TrustManager` and `ReputationManager` are mentioned in TODO comments but never initialized in server-side system initialization. No save/load integration exists.
- **Fix**:
  1. In `cmd/server/v8_systems.go`, replace TODO comments with actual initialization: `trustMgr := persistence.NewTrustManager(...)` and `repMgr := persistence.NewReputationManager(...)`
  2. Wire managers into server state
  3. Add save/load integration with `pkg/saveload`
  4. Ensure `StopAutomaticDecay()` is called on server shutdown (see REM-028)
- **Verify**: `grep -n "TrustManager\|ReputationManager\|persistence.New" cmd/server/v8_systems.go`

#### REM-020: Missing mobile virtual controls
- **Source**: `cmd/mobile/AUDIT.md`
- **Location**: `cmd/mobile/mobile.go:52-72`
- **Problem**: Virtual on-screen controls (dual joystick, action buttons) not initialized despite being available in `pkg/mobile`.
- **Fix**:
  1. In `mobile.go` system initialization (around lines 52-72), import and initialize `pkg/mobile` virtual controls
  2. Call the virtual controls setup function after Ebiten game initialization
  3. Wire touch input events to the virtual control layer
- **Verify**: `grep -rn "pkg/mobile\|virtual.*control\|VirtualControl" cmd/mobile/mobile.go`

#### REM-021: Missing mobile platform detection
- **Source**: `cmd/mobile/AUDIT.md`, `cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `cmd/mobile/mobile.go:16-22`
- **Problem**: No runtime platform detection (iOS vs Android) for platform-specific optimizations, haptic feedback, safe area insets, or device aspect ratios. Hard-coded 720x1280 dimensions.
- **Fix**:
  1. Add runtime platform detection using `runtime.GOOS` and build tags
  2. Query actual device screen dimensions instead of hardcoding 720x1280
  3. Add safe area inset detection for iOS notch and Android navigation bar
  4. Create platform-specific optimization paths based on detected platform
- **Verify**: `grep -n "720\|1280\|runtime.GOOS" cmd/mobile/mobile.go`

#### REM-022: Mobile federation system not initialized
- **Source**: `cmd/mobile/AUDIT.md`, `cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `cmd/mobile/mobile.go:74-88`
- **Problem**: `mobile_federation_system.go` exists in `pkg/network/federation/mobile` but is not initialized in `cmd/mobile`, so mobile multiplayer federation is unavailable.
- **Fix**:
  1. In `mobile.go` system initialization, import `pkg/network/federation/mobile`
  2. Initialize the mobile federation adapter with appropriate mobile-specific config (bandwidth limits, battery awareness)
  3. Wire into the network client initialization path
- **Verify**: `grep -rn "federation.*mobile\|mobile.*federation" cmd/mobile/mobile.go`

#### REM-023: Guild fleet persistence not integrated with saveload
- **Source**: `pkg/integration/guild_vehicle/AUDIT.md`
- **Location**: `pkg/integration/guild_vehicle/fleet_manager.go:327-388`, `types.go:231-281`
- **Problem**: FleetManager saves to standalone gzip files instead of participating in unified save/load workflow via `pkg/saveload`.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` and `Deserialize(data []byte) error` to `GuildVehicleFleetComponent`
  2. Register component with `pkg/saveload` serializer registry
  3. Optionally keep standalone save as backup, but primary persistence should flow through saveload
- **Verify**: `grep -n "Serialize\|Deserialize" pkg/integration/guild_vehicle/types.go`

#### REM-024: Raids hardcode fantasy genre instead of reading world config
- **Source**: `pkg/world/raids/AUDIT.md`
- **Location**: `pkg/world/raids/manager.go:40,89`
- **Problem**: `Manager.GenerateRaid()` and `Manager.CreateInstance()` hardcode `GenreID: "fantasy"` instead of reading from world configuration.
- **Fix**:
  1. Add `genreID string` field to `Manager` struct
  2. Accept `genreID` as a constructor parameter in `NewManager()`
  3. Replace hardcoded `"fantasy"` with `m.genreID` in `GenerateRaid()` and `CreateInstance()`
- **Verify**: `grep -n '"fantasy"' pkg/world/raids/manager.go`

#### REM-025: Territory lacks cross-server federation integration
- **Source**: `pkg/world/territory/AUDIT.md`
- **Location**: `pkg/world/territory/manager.go:1-499`, `doc.go:14`
- **Problem**: Documentation claims "cross-server territory synchronization support" but no federation integration exists. Cross-server guilds cannot declare wars or participate in sieges across server boundaries.
- **Fix**:
  1. Add federation event broadcasting methods to `Manager`: `BroadcastTerritoryUpdate(territoryID, ownerGuildID string)`
  2. Add `BroadcastWarDeclaration(attackerGuild, defenderGuild string)` method
  3. Add `HandleIncomingSiegeJoin(playerID, siegeID, sourceServerID string)` handler
  4. Integrate with `pkg/network/federation/sync.go` `FederationSyncManager`
  5. Add circuit breaker for federation failures (fallback to local-only sieges)
  6. Or update `doc.go` to remove the cross-server claim if deferred
- **Verify**: `grep -rn "federation" pkg/world/territory/`

#### REM-026: Narrative world integration hardcodes fantasy genre
- **Source**: `pkg/integration/narrative_world/AUDIT.md`
- **Location**: `pkg/integration/narrative_world/manager.go:402`
- **Problem**: `GeneratePersonalQuest` hardcodes `GenreID: "fantasy"` instead of accepting/propagating actual genre from game state.
- **Fix**:
  1. Add `genreID string` field to the manager struct
  2. Accept `genreID` in constructor or as parameter to `GeneratePersonalQuest`
  3. Replace hardcoded `"fantasy"` with the dynamic genre value
- **Verify**: `grep -n '"fantasy"' pkg/integration/narrative_world/manager.go`

#### REM-027: Missing context cancellation check in balance validator
- **Source**: `pkg/balance/AUDIT.md`
- **Location**: `pkg/balance/economic.go:248-260`
- **Problem**: `validateGoldBalance()` does not check `ctx.Done()` in two simulation loops. All other validators properly check context in their loops.
- **Fix**:
  1. In `economic.go`, inside the loops at lines 248-260, add:
     ```go
     select {
     case <-ctx.Done():
         return ctx.Err()
     default:
     }
     ```
  2. Follow the same pattern used in other validator loops in the same file
- **Verify**: `grep -n "ctx.Done" pkg/balance/economic.go`

#### REM-028: TrustManager goroutine leak on server shutdown
- **Source**: `pkg/social/persistence/AUDIT.md`
- **Location**: `pkg/social/persistence/trust_manager.go:241-259`
- **Problem**: `TrustManager.StartAutomaticDecay()` creates a goroutine that runs indefinitely. If server shutdown doesn't call `StopAutomaticDecay()`, goroutine leaks.
- **Fix**:
  1. Ensure server shutdown code in `cmd/server/main.go` calls `trustMgr.StopAutomaticDecay()`
  2. Add `StopAutomaticDecay()` to server cleanup/shutdown sequence alongside other resource cleanup
  3. If TrustManager is not yet initialized in server (see REM-019), this becomes part of that integration
- **Verify**: `grep -rn "StopAutomaticDecay" cmd/server/`

#### REM-029: Mobile battery level validation missing
- **Source**: `pkg/network/federation/mobile/AUDIT.md`
- **Location**: `pkg/network/federation/mobile/adapter.go:105`
- **Problem**: `UpdateBatteryLevel` accepts values outside 0.0-1.0 range without validation, which could lead to incorrect BatteryMode calculation.
- **Fix**:
  1. At the start of `UpdateBatteryLevel`, add clamping:
     ```go
     if level < 0.0 { level = 0.0 }
     if level > 1.0 { level = 1.0 }
     ```
- **Verify**: `grep -A5 "func.*UpdateBatteryLevel" pkg/network/federation/mobile/adapter.go`

#### REM-030: Mobile federation Config.MaxBandwidth can be negative
- **Source**: `pkg/network/federation/mobile/AUDIT.md`
- **Location**: `pkg/network/federation/mobile/types.go:89`, `adapter.go:210`
- **Problem**: `Config.MaxBandwidth` can be negative which breaks token bucket algorithm.
- **Fix**:
  1. In the `Config` validation or constructor, add: `if c.MaxBandwidth < 0 { c.MaxBandwidth = 0 }`
  2. Or add validation in `NewAdapter()` that rejects negative bandwidth values with an error
- **Verify**: `grep -n "MaxBandwidth" pkg/network/federation/mobile/types.go pkg/network/federation/mobile/adapter.go`

#### REM-031: WASM build fails due to missing saveload.NewDefaultMigrator
- **Source**: `cmd/server/AUDIT.md`
- **Location**: `pkg/migration/validator.go:39`
- **Problem**: WASM vet fails because `pkg/migration/validator.go:39` references `saveload.NewDefaultMigrator` which doesn't exist in WASM build context.
- **Fix**:
  1. Add build tags to `pkg/migration/validator.go`: `//go:build !js` at the top
  2. Create a stub `validator_wasm.go` with `//go:build js` that provides no-op or error-returning migration validation
  3. Follow the pattern used in `pkg/saveload/` which has separate `storage_wasm.go`
- **Verify**: `GOOS=js GOARCH=wasm go vet ./pkg/migration/...`

#### REM-032: Rendering patterns generator initialized but not actively used
- **Source**: `pkg/rendering/patterns/AUDIT.md`
- **Location**: `cmd/client/handlers.go:260`
- **Problem**: Pattern generator is initialized but there's no evidence of active usage in terrain/tile generation systems. May be dead code awaiting integration.
- **Fix**:
  1. Verify whether tile texture generation in `pkg/rendering/tiles/` uses this generator
  2. If unused, either wire into tile generation pipeline or remove initialization to eliminate dead code
  3. If already used indirectly, document the usage chain
- **Verify**: `grep -rn "patternGenerator\|PatternGenerator\|rendering/patterns" cmd/client/ pkg/rendering/tiles/ pkg/engine/`

#### REM-033: Rendering tiles package appears to be unused dead code
- **Source**: `pkg/rendering/tiles/AUDIT.md`
- **Location**: `pkg/rendering/tiles/`
- **Problem**: Package is not imported by any engine, client, or server files — appears to be unused dead code or awaiting integration.
- **Fix**:
  1. Search for imports: `grep -rn "rendering/tiles" cmd/ pkg/engine/`
  2. If truly unused, either integrate into the rendering pipeline or mark as planned/experimental with a doc comment
  3. If integration is planned, add a TODO with tracking issue reference
- **Verify**: `grep -rn "rendering/tiles" cmd/ pkg/ --include="*.go" | grep -v _test.go | grep -v AUDIT`

#### REM-034: Server main.go too large (1139 LOC)
- **Source**: `cmd/server/AUDIT.md`
- **Location**: `cmd/server/main.go`
- **Problem**: `main.go` is 1,139 LOC. Should extract initialization functions to `init_*.go` files following the pattern used in `cmd/client`.
- **Fix**:
  1. Extract flag parsing into `cmd/server/flags.go`
  2. Extract system initialization into `cmd/server/init_systems.go`
  3. Extract network setup into `cmd/server/init_network.go`
  4. Keep `main()` as a thin orchestrator calling initialization functions
  5. Follow the pattern used in `cmd/client/`
- **Verify**: `wc -l cmd/server/main.go`

#### REM-035: Rendering pool package not explicitly initialized in client
- **Source**: `pkg/rendering/pool/AUDIT.md`
- **Location**: `pkg/rendering/pool/image_pool.go:36-37`
- **Problem**: Package not initialized in `cmd/client/main.go` startup; only created on-demand in `handlers.go:671`. Lazy initialization may cause unexpected allocation at runtime.
- **Fix**:
  1. Document in `image_pool.go` that the pool is intentionally created per-render-system instance rather than as global singleton
  2. Or move initialization to client startup if a single shared pool is preferred
- **Verify**: `grep -rn "image_pool\|ImagePool\|NewImagePool" cmd/client/`

#### REM-036: Quality settings component lacks serialization
- **Source**: `pkg/rendering/quality/AUDIT.md`
- **Location**: `pkg/rendering/quality/quality_settings_component.go:8-33`
- **Problem**: `QualitySettingsComponent` lacks `Serialize()/Deserialize()` methods. Quality overrides won't persist across save/load.
- **Fix**:
  1. If quality settings should persist, add `Serialize() ([]byte, error)` and `Deserialize(data []byte) error` methods
  2. If runtime-only is intentional, add a comment: `// QualitySettingsComponent is runtime-only and intentionally does not persist across save/load.`
- **Verify**: `grep -n "Serialize\|Deserialize" pkg/rendering/quality/quality_settings_component.go`

#### REM-037: AutoAdjuster callback blocks update lock
- **Source**: `pkg/rendering/quality/AUDIT.md`
- **Location**: `pkg/rendering/quality/auto_adjuster.go:56-82`
- **Problem**: `AutoAdjuster.Update()` holds write lock during entire update including callback invocation. Slow callbacks block all updates.
- **Fix**:
  1. Collect callback reference under lock, release lock, then invoke callback outside the lock
  2. Or document that callbacks must complete quickly (< 1ms)
  3. Example pattern:
     ```go
     a.mu.Lock()
     // ... compute adjustments ...
     cb := a.callback
     a.mu.Unlock()
     if cb != nil { cb(newSettings) }
     ```
- **Verify**: `grep -n "Lock\|Unlock\|callback" pkg/rendering/quality/auto_adjuster.go`

#### REM-038: Version duplication between version and saveload packages
- **Source**: `pkg/version/AUDIT.md`
- **Location**: `pkg/saveload/types.go:11`, `pkg/version/version.go:32`
- **Problem**: `SaveVersion` is hardcoded as `"1.0.0"` in `pkg/saveload/types.go` instead of importing from `pkg/version`, creating a second source of truth that can drift.
- **Fix**:
  1. In `pkg/saveload/types.go`, replace the hardcoded version:
     ```go
     import "github.com/opd-ai/venture/pkg/version"
     var SaveVersion = version.Version
     ```
  2. Remove the hardcoded `"1.0.0"` constant
- **Verify**: `grep -n "SaveVersion\|1.0.0" pkg/saveload/types.go`

#### REM-039: Fragment positioning assumes 100x100 map
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/generator.go:267-276`
- **Problem**: `generateLocation()` hardcodes `10.0 + progress*80.0` positioning logic assuming a 100x100 dungeon, breaking layout on differently-sized terrains.
- **Fix**:
  1. Accept terrain bounds as parameters (e.g., `mapWidth`, `mapHeight` floats)
  2. Replace `10.0 + progress*80.0` with `margin + progress*(mapWidth - 2*margin)` where margin is proportional
  3. Pass actual terrain dimensions from `GenerationParams` or terrain metadata
- **Verify**: `grep -n "10.0.*progress\|80.0" pkg/procgen/story/generator.go`

#### REM-040: Migrator interface defined twice (desktop and WASM)
- **Source**: `pkg/saveload/AUDIT.md`
- **Location**: `pkg/saveload/migrator.go:9-20`, `pkg/saveload/storage_wasm.go:29-41`
- **Problem**: `Migrator` interface is defined twice — once for desktop, once for WASM. This duplication risks drift.
- **Fix**:
  1. Create `pkg/saveload/migrator_interface.go` without build tags containing the shared `Migrator` interface
  2. Remove the interface definition from `migrator.go` and `storage_wasm.go`
  3. Both files import/use the shared definition
- **Verify**: `grep -rn "type Migrator interface" pkg/saveload/`

---

### MEDIUM

#### REM-041: Mobile time-based seed fallback violates determinism
- **Source**: `cmd/mobile/AUDIT.md`, `cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `cmd/mobile/config/seed.go:36`
- **Problem**: `config.GetSeedFromEnv` uses `time.Now().UnixNano()` when env var is unset. Documented as intentional for mobile UX, but the seed should be shown in-game UI for reproducibility.
- **Fix**:
  1. When a random seed is generated, store it in a visible UI element or log it at Info level
  2. Add seed display to the mobile game HUD or settings screen
  3. Add structured log: `logger.WithField("seed", seed).Info("generated random seed")`
- **Verify**: `grep -n "time.Now\|UnixNano" cmd/mobile/config/seed.go`

#### REM-042: Mobile package-level globals complicate testing
- **Source**: `cmd/mobile/AUDIT.md`, `cmd/mobile/AUDIT_2026-02-26_COMPREHENSIVE.md`
- **Location**: `cmd/mobile/mobile.go:24-32`
- **Problem**: Uses package globals (`gameInstance`, `logger`, `systemsInitResult`, `playerEntity`, `worldSeed`, `genreID`) which complicate testing and prevent multi-instance scenarios.
- **Fix**:
  1. Create a `MobileApp` struct containing all current package-level state
  2. Move `gameInstance`, `logger`, `systemsInitResult`, `playerEntity`, `worldSeed`, `genreID` as fields
  3. Update exported functions to be methods on `MobileApp`
  4. Keep a single package-level `var app *MobileApp` for the Ebiten mobile entry point
- **Verify**: `grep -n "^var " cmd/mobile/mobile.go`

#### REM-043: Server tests require X11 but lack Xvfb documentation
- **Source**: `cmd/server/AUDIT.md`
- **Location**: `cmd/server/main_test.go:1`
- **Problem**: Tests require X11/display but no Xvfb wrapper is documented for headless CI.
- **Fix**:
  1. Add a comment to `main_test.go` documenting: `// Tests require X11. Run with: xvfb-run go test ./cmd/server/...`
  2. Add `xvfb-run` wrapper to CI test commands in `.github/workflows/`
  3. Or add `t.Skip("requires display")` when `os.Getenv("DISPLAY")` is empty
- **Verify**: `grep -n "Xvfb\|DISPLAY\|xvfb" cmd/server/main_test.go .github/workflows/*.yml`

#### REM-044: Server time.Now() usage for timing lacks GameClock abstraction
- **Source**: `cmd/server/AUDIT.md`
- **Location**: `cmd/server/main.go:298`, `cmd/server/main.go:751`
- **Problem**: `time.Now()` used for server timing is not deterministic for save/load replay.
- **Fix**:
  1. Import or define a `GameClock` interface in `cmd/server/`
  2. Replace `time.Now()` at lines 298 and 751 with `clock.Now()`
  3. Accept `GameClock` in the server initialization, defaulting to system time
- **Verify**: `grep -n "time.Now()" cmd/server/main.go`

#### REM-045: Missing server documentation for entity_spawning.go
- **Source**: `cmd/server/AUDIT.md`
- **Location**: `cmd/server/entity_spawning.go`
- **Problem**: File lacks package-level doc comment explaining server-side vs client-side spawning differences.
- **Fix**:
  1. Add a file-level comment at the top of `entity_spawning.go`:
     ```go
     // entity_spawning.go handles server-authoritative entity spawning.
     // Unlike client-side spawning which is prediction-based, server spawning
     // is authoritative and broadcasts entity creation to all connected clients.
     ```
- **Verify**: `head -5 cmd/server/entity_spawning.go`

#### REM-046: Missing server documentation for system_wrappers.go
- **Source**: `cmd/server/AUDIT.md`
- **Location**: `cmd/server/system_wrappers.go`
- **Problem**: File lacks explanation of why wrappers are needed (signature mismatch pattern).
- **Fix**:
  1. Add a file-level comment:
     ```go
     // system_wrappers.go provides adapter wrappers for systems whose Update()
     // signatures don't match the standard engine.System interface. These wrappers
     // bridge the gap between domain-specific system signatures and ECS registration.
     ```
- **Verify**: `head -5 cmd/server/system_wrappers.go`

#### REM-047: Config package has coupling to procgen/dialog for genre validation
- **Source**: `pkg/config/AUDIT.md`
- **Location**: `pkg/config/validator.go:32`
- **Problem**: `pkg/config` depends on `pkg/procgen/dialog.GetAvailableGenres()` for genre validation, creating an upward dependency from utility to generation package.
- **Fix**:
  1. Extract available genre list to a shared constants package (e.g., `pkg/procgen/genre/constants.go`)
  2. Have both `pkg/config` and `pkg/procgen/dialog` import from the shared location
  3. Or accept the coupling with a doc comment explaining it ensures genre consistency
- **Verify**: `grep -n "procgen/dialog\|GetAvailableGenres" pkg/config/validator.go`

#### REM-048: Register() silently ignores nil features
- **Source**: `pkg/audit/features/AUDIT.md`
- **Location**: `pkg/audit/features/feature_completeness.go:106-109`
- **Problem**: `Register()` method silently ignores nil features without logging.
- **Fix**:
  1. Add structured warning log:
     ```go
     if feature == nil {
         logrus.WithFields(logrus.Fields{
             "operation": "register_feature",
         }).Warn("attempted to register nil feature")
         return
     }
     ```
- **Verify**: `grep -A3 "nil" pkg/audit/features/feature_completeness.go | head -10`

#### REM-049: Missing godoc on feature category constants
- **Source**: `pkg/audit/features/AUDIT.md`
- **Location**: `pkg/audit/features/constants.go:10-20`
- **Problem**: Feature category constants (`CategoryCore`, etc.) are exported but lack individual godoc comments.
- **Fix**:
  1. Add godoc comment above each constant explaining which feature domains it covers
  2. Example: `// CategoryCore covers fundamental engine infrastructure features.`
- **Verify**: `grep -B1 "Category" pkg/audit/features/constants.go | head -20`

#### REM-050: Missing godoc on GetDefaultRegistry()
- **Source**: `pkg/audit/features/AUDIT.md`
- **Location**: `pkg/audit/features/meta_features.go:250`
- **Problem**: Primary public API entry point `GetDefaultRegistry()` lacks godoc comment.
- **Fix**:
  1. Add godoc: `// GetDefaultRegistry returns a pre-populated FeatureRegistry containing all known feature definitions for the Venture project. This is the primary entry point for feature completeness auditing.`
- **Verify**: `grep -B2 "func GetDefaultRegistry" pkg/audit/features/meta_features.go`

#### REM-051: Progress logging frequency may be excessive for long simulations
- **Source**: `pkg/balance/AUDIT.md`
- **Location**: `pkg/balance/combat.go:131-136`, `economic.go:106-111`
- **Problem**: Progress logs use fixed intervals (every 10%, every 25 recipes) which creates many log entries for 10K-battle simulations.
- **Fix**:
  1. Make progress logging interval configurable via validator config
  2. Or reduce to log at 25%, 50%, 75%, 100% milestones
  3. Use `logrus.Debug` level for intermediate progress, `logrus.Info` for completion only
- **Verify**: `grep -n "progress\|Progress\|%" pkg/balance/combat.go pkg/balance/economic.go`

#### REM-052: Hardcoded stat budget comment lacks validation
- **Source**: `pkg/balance/AUDIT.md`
- **Location**: `pkg/balance/combat.go:264-266`
- **Problem**: Comment says "Total stat budget ~165 per class" but no assertion validates this.
- **Fix**:
  1. Add a runtime assertion in the validator: `budget := class.Attack + class.Defense + class.MaxHP; if budget != expectedBudget { log.Warn(...) }`
  2. Or remove the comment if the budget is approximate and not enforced
- **Verify**: `grep -n "165\|stat budget" pkg/balance/combat.go`

#### REM-053: Choice consequences CompanionReactions hard-coded limit
- **Source**: `pkg/integration/choice_consequences/AUDIT.md`
- **Location**: `pkg/integration/choice_consequences/choice_tracker.go:545`
- **Problem**: `CompanionReactions` slice hard-coded to keep last 20 reactions. Should be configurable like `npcMemoryLimit` and `choiceLimit`.
- **Fix**:
  1. Add `companionReactionLimit int` field to the tracker config/constructor
  2. Replace hardcoded `20` with `t.companionReactionLimit`
  3. Set default to 20 in constructor
- **Verify**: `grep -n "20\|CompanionReactions" pkg/integration/choice_consequences/choice_tracker.go`

#### REM-054: Companion housing doc.go references deprecated method names
- **Source**: `pkg/integration/companion_housing/AUDIT.md`
- **Location**: `pkg/integration/companion_housing/doc.go:18-34`
- **Problem**: Doc string references deprecated methods `IsInHouse()`, `HasBedding()`, `IsTraining()` that are now internal. External code should use `PetHomeManager`.
- **Fix**:
  1. Update `doc.go` to reference `PetHomeManager` methods instead of deprecated internal names
  2. Replace example code snippets showing old API with new API
- **Verify**: `grep -n "IsInHouse\|HasBedding\|IsTraining" pkg/integration/companion_housing/doc.go`

#### REM-055: Companion housing has redundant unexported system
- **Source**: `pkg/integration/companion_housing/AUDIT.md`
- **Location**: `pkg/integration/companion_housing/companion_housing_system.go:27-37`
- **Problem**: `companionHousingSystem` is unexported and exists "for internal test coverage only" while `PetHomeManager` supersedes functionality.
- **Fix**:
  1. If `PetHomeManager` fully replaces the system, remove `companionHousingSystem`
  2. Move any test-only functionality to test files using `_test.go` suffix
  3. Or document why the unexported system must remain
- **Verify**: `grep -rn "companionHousingSystem" pkg/integration/companion_housing/`

#### REM-056: Housing crafting system unexported but test-only
- **Source**: `pkg/integration/housing_crafting/AUDIT.md`
- **Location**: `pkg/integration/housing_crafting/housing_crafting_system.go:10`
- **Problem**: `housingCraftingSystem` is unexported but only used in tests; consider removing if `StationManager` fully replaces it.
- **Fix**:
  1. Move system to a `_test.go` file if only used in tests
  2. Or remove entirely if `StationManager` supersedes all functionality
- **Verify**: `grep -rn "housingCraftingSystem" pkg/integration/housing_crafting/`

#### REM-057: Narrative world location tracking hardcodes "unknown"
- **Source**: `pkg/integration/narrative_world/AUDIT.md`
- **Location**: `pkg/integration/narrative_world/manager.go:474`
- **Problem**: `RecordMemory` hardcodes `Location: "unknown"`. Could integrate with `PositionComponent` for better context.
- **Fix**:
  1. Accept a `location string` parameter in `RecordMemory`
  2. At the call site, extract location from entity's `PositionComponent` and pass it
  3. Fallback to `"unknown"` if position is unavailable
- **Verify**: `grep -n "unknown\|Location" pkg/integration/narrative_world/manager.go`

#### REM-058: Political warfare naming typo: ResponingAllies
- **Source**: `pkg/integration/political_warfare/AUDIT.md`
- **Location**: `pkg/integration/political_warfare/types.go:54`
- **Problem**: `ResponingAllies` should be `RespondingAllies` in `AllianceCall` struct and JSON tags.
- **Fix**:
  1. Rename field in `types.go:54`: `ResponingAllies` → `RespondingAllies`
  2. Keep JSON tag as `"responding_allies"` (already correct)
  3. Update all references in `manager.go:252,287` and `manager_test.go:192,447`
- **Verify**: `grep -rn "ResponingAllies" pkg/integration/political_warfare/`

#### REM-059: Political warfare tests require X11/Ebiten
- **Source**: `pkg/integration/political_warfare/AUDIT.md`
- **Location**: `pkg/integration/political_warfare/manager_test.go:7`, `system_test.go:7`
- **Problem**: Tests import `pkg/engine` which requires X11/Wayland/Ebiten, preventing headless CI execution.
- **Fix**:
  1. Add build tag `//go:build !headless` to test files, or
  2. Add `t.Skip("requires display")` when `os.Getenv("DISPLAY")` is empty
  3. Or create interface abstractions to decouple from engine types
- **Verify**: `DISPLAY= go test ./pkg/integration/political_warfare/... 2>&1 | head -5`

#### REM-060: Trade routes tests cannot run in headless CI
- **Source**: `pkg/integration/trade_routes/AUDIT.md`
- **Location**: `pkg/integration/trade_routes/manager_test.go`, `economy_integration_test.go`
- **Problem**: Ebiten dependency via `pkg/procgen/vehicle` prevents headless CI testing.
- **Fix**:
  1. Define a `VehicleGenerator` interface in the trade_routes package
  2. Replace direct import of `pkg/procgen/vehicle` with interface
  3. Provide a stub implementation for tests
- **Verify**: `go test ./pkg/integration/trade_routes/... 2>&1 | head -5`

#### REM-061: Trade route completeRoute iterates cargo twice
- **Source**: `pkg/integration/trade_routes/AUDIT.md`
- **Location**: `pkg/integration/trade_routes/manager.go:724-748`
- **Problem**: `completeRoute()` iterates cargo twice (profit calculation, then price impacts). Could use single-pass.
- **Fix**:
  1. Combine both loops into a single iteration that calculates profit and applies price impacts in one pass
  2. Accumulate totals in local variables during the single pass
- **Verify**: `sed -n '724,748p' pkg/integration/trade_routes/manager.go`

#### REM-062: World events exported types lack godoc
- **Source**: `pkg/integration/world_events/AUDIT.md`
- **Location**: `pkg/integration/world_events/types.go:96-133`
- **Problem**: `FactionResponse`, `EconomicEvent`, `WeatherDisaster`, `EventChain` lack godoc comments.
- **Fix**:
  1. Add godoc comment above each exported type in `types.go`
  2. Example: `// FactionResponse represents a faction's reaction to a world event, including attitude change and action taken.`
- **Verify**: `grep -B1 "type FactionResponse\|type EconomicEvent\|type WeatherDisaster\|type EventChain" pkg/integration/world_events/types.go`

#### REM-063: World events constructor naming convention
- **Source**: `pkg/integration/world_events/AUDIT.md`
- **Location**: `pkg/integration/world_events/types.go:181`
- **Problem**: `EventManagerConfig.DefaultEventManagerConfig()` should follow Go `New*` naming convention.
- **Fix**:
  1. Rename `DefaultEventManagerConfig()` to `NewDefaultEventManagerConfig()`
  2. Update all call sites
- **Verify**: `grep -rn "DefaultEventManagerConfig" pkg/integration/world_events/`

#### REM-064: Network chat type assertion lacks error detail
- **Source**: `pkg/network/chat/AUDIT.md`
- **Location**: `pkg/network/chat/system.go:115-121`
- **Problem**: Uses direct type assertion without logging the actual type received when assertion fails.
- **Fix**:
  1. Replace bare type assertion with checked assertion:
     ```go
     comp, ok := entity.GetComponent("chat").(*ChatComponent)
     if !ok {
         logger.WithFields(logrus.Fields{
             "actual_type": fmt.Sprintf("%T", entity.GetComponent("chat")),
         }).Warn("unexpected component type for chat")
         continue
     }
     ```
- **Verify**: `grep -n "type assertion\|\.(\*" pkg/network/chat/system.go`

#### REM-065: Network resilience custom string formatting
- **Source**: `pkg/network/resilience/AUDIT.md`
- **Location**: `pkg/network/resilience/scenario.go:240-270`
- **Problem**: Custom `formatInt()` and `trimTrailingZeros()` implementations should use `strconv.FormatInt()` and `strconv.FormatFloat()`.
- **Fix**:
  1. Replace `formatInt()` with `strconv.FormatInt(n, 10)`
  2. Replace `trimTrailingZeros()` usage with `strconv.FormatFloat(f, 'f', -1, 64)`
  3. Remove the custom helper functions
- **Verify**: `grep -n "formatInt\|trimTrailingZeros" pkg/network/resilience/scenario.go`

#### REM-066: Procgen audit doc.go lacks EnvironmentGenerator exclusion explanation
- **Source**: `pkg/procgen/audit/AUDIT.md`
- **Location**: `pkg/procgen/audit/doc.go:34`
- **Problem**: Doc mentions "EnvironmentGenerator" exclusion but doesn't explain why Config-based API is incompatible with seed/params audit pattern.
- **Fix**:
  1. Add explanation to `doc.go`: `// EnvironmentGenerator uses a Config-based API (GenerateEnvironment(config Config)) rather than the standard (seed int64, params GenerationParams) pattern, making it incompatible with the generic generator audit framework.`
- **Verify**: `grep -n "EnvironmentGenerator" pkg/procgen/audit/doc.go`

#### REM-067: Procgen generator.go missing inline hash algorithm docs
- **Source**: `pkg/procgen/AUDIT.md`
- **Location**: `pkg/procgen/generator.go:47-53`
- **Problem**: Polynomial rolling hash in `SeedGenerator.GetSeed` lacks inline comments explaining why base 31 was chosen and collision characteristics.
- **Fix**:
  1. Add inline comments:
     ```go
     // Uses polynomial rolling hash with base 31 (a common prime for string hashing,
     // chosen for good distribution across ASCII character codes).
     // Collision rate: ~1/2^63 for distinct inputs under 100 chars.
     ```
- **Verify**: `sed -n '47,53p' pkg/procgen/generator.go`

#### REM-068: Building generator guild hall methods should be extracted
- **Source**: `pkg/procgen/building/AUDIT.md`
- **Location**: `pkg/procgen/building/generator.go:447-533`
- **Problem**: 5 guild hall layout generation methods are mixed with general building generation. Should be in separate file.
- **Fix**:
  1. Create `pkg/procgen/building/guild_hall_layout.go`
  2. Move `calculateGuildHallLayout`, `generateFloorRooms`, `determineGuildRoomType`, `addFloorDoors`, and related methods to the new file
  3. Keep same package declaration
- **Verify**: `grep -n "GuildHall\|guildHall\|guild_hall" pkg/procgen/building/generator.go`

#### REM-069: Class generator has no multiclass support
- **Source**: `pkg/procgen/class/AUDIT.md`
- **Location**: `pkg/procgen/class/generator.go:13-35`
- **Problem**: `ClassPreset` does not expose hybrid class parent classes or stat blending ratios. `pkg/class/advanced/` exists for advanced multiclassing but has no integration with this generator.
- **Fix**:
  1. Add `ParentClasses []string` and `BlendRatios []float64` fields to `ClassPreset`
  2. Add hybrid generation method that combines two class presets
  3. Or wire `pkg/class/advanced/` multiclass types into the generator output
- **Verify**: `grep -n "Multiclass\|multiclass\|hybrid\|Hybrid" pkg/procgen/class/generator.go`

#### REM-070: Class generator presets lack save/load integration
- **Source**: `pkg/procgen/class/AUDIT.md`
- **Location**: `pkg/procgen/class/generator.go:13-35`
- **Problem**: `ClassPreset` does not implement `ComponentSerializer` interface. Character class data persistence relies on hardcoded engine types.
- **Fix**:
  1. Add `Serialize() ([]byte, error)` and `Deserialize(data []byte) error` to `ClassPreset`
  2. Add JSON struct tags to all `ClassPreset` fields
- **Verify**: `grep -n "Serialize\|Deserialize" pkg/procgen/class/generator.go`

#### REM-071: Companion type lacks godoc
- **Source**: `pkg/procgen/companion/AUDIT.md`
- **Location**: `pkg/procgen/companion/generator.go:18`
- **Problem**: `Companion` struct is exported but lacks godoc comment.
- **Fix**:
  1. Add: `// Companion represents a generated companion with stats, commands, abilities, and visual description derived from genre-specific templates.`
- **Verify**: `grep -B1 "type Companion struct" pkg/procgen/companion/generator.go`

#### REM-072: Dialog temperature weights lack inline comments
- **Source**: `pkg/procgen/dialog/AUDIT.md`
- **Location**: `pkg/procgen/dialog/utils.go:136-147`
- **Problem**: `calculateTemperatureWeights()` has complex temperature scaling logic without inline comments.
- **Fix**:
  1. Add inline comments explaining the mathematical transformation at each step
  2. Document the expected input range and output distribution
- **Verify**: `sed -n '136,147p' pkg/procgen/dialog/utils.go`

#### REM-073: Entity queries nil safety could use helper
- **Source**: `pkg/procgen/entity/AUDIT.md`
- **Location**: `pkg/procgen/entity/queries.go:9-27`
- **Problem**: Nil safety checks repeated in all functions. Pattern could be extracted to a helper.
- **Fix**:
  1. Create a `checkNil(entity interface{}, operation string) error` helper function
  2. Replace repeated nil checks with helper calls
  3. Or accept the pattern if it's clearer for readability (add doc comment explaining choice)
- **Verify**: `grep -n "nil" pkg/procgen/entity/queries.go | head -10`

#### REM-074: Environment generator.go is large (1296 LOC)
- **Source**: `pkg/procgen/environment/AUDIT.md`
- **Location**: `pkg/procgen/environment/generator.go:1-1296`
- **Problem**: File could benefit from splitting drawing functions into a separate file.
- **Fix**:
  1. Extract drawing/rendering helper functions to `pkg/procgen/environment/drawing.go`
  2. Keep generation logic in `generator.go`
- **Verify**: `wc -l pkg/procgen/environment/generator.go`

#### REM-075: Environment doc.go has triple package comment
- **Source**: `pkg/procgen/environment/AUDIT.md`
- **Location**: `pkg/procgen/environment/doc.go:1-45`
- **Problem**: Triple package comment declaration in doc.go.
- **Fix**:
  1. Consolidate into a single `// Package environment ...` block at lines 1-45
  2. Remove duplicate `package environment` declarations
- **Verify**: `grep -c "^// Package\|^package" pkg/procgen/environment/doc.go`

#### REM-076: Faction test uses t.Logf instead of assertion
- **Source**: `pkg/procgen/faction/AUDIT.md`
- **Location**: `pkg/procgen/faction/generator_test.go:446`
- **Problem**: `TestGenerator_SpecialRelationships` uses `t.Logf` instead of assertion for corp-vs-rebel enemy check due to randomness.
- **Fix**:
  1. Use a fixed seed known to produce the special relationship
  2. Replace `t.Logf` with `t.Errorf` or `require.True` using the deterministic seed
- **Verify**: `sed -n '440,450p' pkg/procgen/faction/generator_test.go`

#### REM-077: Furniture placement.go lacks documentation
- **Source**: `pkg/procgen/furniture/AUDIT.md`
- **Location**: `pkg/procgen/furniture/placement.go:1`
- **Problem**: Missing package-level documentation explaining AABB collision algorithm and grid-based auto-placement strategy.
- **Fix**:
  1. Add file-level comment: `// placement.go implements AABB (Axis-Aligned Bounding Box) collision detection for furniture placement within rooms. Uses a grid-based auto-placement strategy that divides available floor space into cells and assigns furniture positions using greedy bin-packing.`
- **Verify**: `head -5 pkg/procgen/furniture/placement.go`

#### REM-078: Furniture name/description generation functions are large
- **Source**: `pkg/procgen/furniture/AUDIT.md`
- **Location**: `pkg/procgen/furniture/generator.go:573-737`
- **Problem**: `generateName()` and `generateDescription()` (100+ lines with nested switch statements) should be refactored into lookup tables.
- **Fix**:
  1. Create genre-specific name/description lookup maps: `var genreNames = map[string][]string{...}`
  2. Replace switch statements with map lookups
  3. Reduces nesting and makes adding new genres easier
- **Verify**: `sed -n '573,580p' pkg/procgen/furniture/generator.go`

#### REM-079: Genre BlendedGenre struct fields lack godoc
- **Source**: `pkg/procgen/genre/AUDIT.md`
- **Location**: `pkg/procgen/genre/blender.go:13-18`
- **Problem**: `BlendedGenre` struct fields lack godoc comments.
- **Fix**:
  1. Add godoc comments to each field in the `BlendedGenre` struct
  2. Document field semantics (e.g., weight range, genre ID format)
- **Verify**: `sed -n '13,18p' pkg/procgen/genre/blender.go`

#### REM-080: Item generator uses armor templates as accessory fallback
- **Source**: `pkg/procgen/item/AUDIT.md`
- **Location**: `pkg/procgen/item/generator.go:156-159`
- **Problem**: Accessory templates fallback to armor templates instead of having dedicated accessory templates.
- **Fix**:
  1. Create dedicated accessory templates in `templates.go` (rings, amulets, trinkets, etc.)
  2. Update `generator.go:156-159` to use accessory-specific templates
- **Verify**: `grep -n "accessory\|Accessory" pkg/procgen/item/generator.go`

#### REM-081: Item templates.go too large (34KB)
- **Source**: `pkg/procgen/item/AUDIT.md`
- **Location**: `pkg/procgen/item/templates.go`
- **Problem**: Single 34KB template file is too large. Should split into genre-specific files.
- **Fix**:
  1. Create `templates_fantasy.go`, `templates_scifi.go`, `templates_horror.go`, `templates_cyberpunk.go`, `templates_postapoc.go`
  2. Move genre-specific template arrays to respective files
  3. Keep shared/common templates in `templates.go`
- **Verify**: `wc -l pkg/procgen/item/templates.go`

#### REM-082: Legendary manager GetStatistics returns time.Time directly
- **Source**: `pkg/procgen/legendary/AUDIT.md`
- **Location**: `pkg/procgen/legendary/manager.go:609`
- **Problem**: `QuestManager.GetStatistics()` uses `qm.timeProvider.Now()` but returns `time.Time` directly instead of wrapping in a struct.
- **Fix**:
  1. This is minor — document that the returned `time.Time` comes from the injected `TimeProvider`
  2. Or wrap return in a `Statistics` struct that also includes the time source for clarity
- **Verify**: `grep -n "GetStatistics" pkg/procgen/legendary/manager.go`

#### REM-083: Magic package doc.go lacks cross-references
- **Source**: `pkg/procgen/magic/AUDIT.md`
- **Location**: `pkg/procgen/magic/doc.go`
- **Problem**: Missing "See Also" section linking to related packages that consume this generator.
- **Fix**:
  1. Add to `doc.go`:
     ```go
     // See Also:
     //   - pkg/engine/spell_casting.go - Spell execution system
     //   - pkg/engine/spell_effect_system.go - Spell effect resolution
     //   - pkg/engine/spell_combination_system.go - Spell combination mechanics
     ```
- **Verify**: `grep -n "See Also" pkg/procgen/magic/doc.go`

#### REM-084: Magic power calculation inconsistency
- **Source**: `pkg/procgen/magic/AUDIT.md`
- **Location**: `pkg/procgen/magic/types.go:229`, `balance.go:79`
- **Problem**: `Spell.GetPowerLevel()` uses different algorithm than `BalanceConfig.calculatePower()`, potentially causing confusion.
- **Fix**:
  1. Document the difference: `GetPowerLevel()` is for display/UI, `calculatePower()` is for balance validation
  2. Or consolidate to a single algorithm used by both
- **Verify**: `grep -n "GetPowerLevel\|calculatePower" pkg/procgen/magic/types.go pkg/procgen/magic/balance.go`

#### REM-085: Minigame name generation has code duplication
- **Source**: `pkg/procgen/minigame/AUDIT.md`
- **Location**: `pkg/procgen/minigame/generator.go:325-383`
- **Problem**: Name generation functions could use shared helper with genre-specific lookup tables.
- **Fix**:
  1. Create a shared `generateGenreName(rng *rand.Rand, genre string, nameMap map[string][]string) string` helper
  2. Replace duplicated genre switch statements with map lookups
- **Verify**: `sed -n '325,383p' pkg/procgen/minigame/generator.go`

#### REM-086: Minigame games Render() deprecated but still present
- **Source**: `pkg/procgen/minigame/games/AUDIT.md`
- **Location**: `pkg/procgen/minigame/games/memory.go:142`, `hacking.go:224`
- **Problem**: Deprecated `Render()` method still present for backward compatibility. Consider removal in V5.0.
- **Fix**:
  1. Add `// Deprecated: Use RenderToImage instead. Will be removed in V5.0.` godoc comment
  2. Track removal in a future version milestone
- **Verify**: `grep -n "func.*Render()" pkg/procgen/minigame/games/*.go`

#### REM-087: Narrative StoryArc and PlotPoint fields lack godoc
- **Source**: `pkg/procgen/narrative/AUDIT.md`
- **Location**: `pkg/procgen/narrative/generator.go:12-81`
- **Problem**: `StoryArc`, `PlotPoint`, and `PlayerChoice` struct fields lack individual documentation.
- **Fix**:
  1. Add godoc comments to each field in all three structs
- **Verify**: `sed -n '12,81p' pkg/procgen/narrative/generator.go`

#### REM-088: Narrative generateTitle returns "Untitled Story" without logging
- **Source**: `pkg/procgen/narrative/AUDIT.md`
- **Location**: `pkg/procgen/narrative/generator.go:276`
- **Problem**: Returns "Untitled Story" for unknown genres silently.
- **Fix**:
  1. Add warning log: `logrus.WithField("genre", genre).Warn("unknown genre, falling back to default title")`
- **Verify**: `grep -n "Untitled Story" pkg/procgen/narrative/generator.go`

#### REM-089: Puzzle CSP.NewCSP accepts unused seed parameter
- **Source**: `pkg/procgen/puzzle/AUDIT.md`
- **Location**: `pkg/procgen/puzzle/solver.go:43`
- **Problem**: `NewCSP` accepts seed parameter but doesn't use it for any deterministic behavior.
- **Fix**:
  1. Either use the seed to initialize an RNG for heuristic tie-breaking in solver
  2. Or remove the unused seed parameter and update call sites
- **Verify**: `grep -n "func NewCSP" pkg/procgen/puzzle/solver.go`

#### REM-090: Puzzle PuzzleElement.State uses interface{} without documented types
- **Source**: `pkg/procgen/puzzle/AUDIT.md`
- **Location**: `pkg/procgen/puzzle/generator.go:63`
- **Problem**: `PuzzleElement.State` uses `interface{}` without documenting valid types per `ElementType`.
- **Fix**:
  1. Add godoc: `// State holds element-specific state. Valid types: bool (pressure_plate, lever), string (door), map[string]interface{} (pushable_block), int (rotating_tile).`
- **Verify**: `grep -n "State.*interface" pkg/procgen/puzzle/generator.go`

#### REM-091: Quest methods have procedural and receiver API duplication
- **Source**: `pkg/procgen/quest/AUDIT.md`
- **Location**: `pkg/procgen/quest/types.go:217-249`
- **Problem**: Both `QuestIsComplete(q *Quest)` and `q.IsComplete()` exist, creating API duplication.
- **Fix**:
  1. Deprecate the procedural function variants with godoc: `// Deprecated: Use q.IsComplete() instead.`
  2. Or remove procedural variants if no external callers exist
- **Verify**: `grep -rn "QuestIsComplete" pkg/`

#### REM-092: Recipe tests fail due to GLFW/X11 dependency
- **Source**: `pkg/procgen/recipe/AUDIT.md`
- **Location**: `pkg/procgen/recipe/generator_test.go`
- **Problem**: Tests import engine which requires X11, preventing headless CI.
- **Fix**:
  1. Add `t.Skip("requires display")` at the top of TestMain when DISPLAY is unset
  2. Or create a stub `engine.Recipe` type in a test helper file
- **Verify**: `DISPLAY= go test ./pkg/procgen/recipe/... 2>&1 | head -3`

#### REM-093: Recipe material quantities hardcoded to 1-3
- **Source**: `pkg/procgen/recipe/AUDIT.md`
- **Location**: `pkg/procgen/recipe/generator.go:189`
- **Problem**: Material quantities hardcoded to `1 + rng.Intn(3)` with no template-level control.
- **Fix**:
  1. Add `MinQuantity` and `MaxQuantity` fields to `RecipeTemplate`
  2. Replace `1 + rng.Intn(3)` with `template.MinQuantity + rng.Intn(template.MaxQuantity - template.MinQuantity + 1)`
  3. Default to `MinQuantity: 1, MaxQuantity: 3` in existing templates
- **Verify**: `grep -n "Intn(3)" pkg/procgen/recipe/generator.go`

#### REM-094: Skills template functions lack godoc
- **Source**: `pkg/procgen/skills/AUDIT.md`
- **Location**: `pkg/procgen/skills/templates.go`
- **Problem**: `GetFantasyTreeTemplates`, `GetSciFiTreeTemplates`, etc. lack godoc comments.
- **Fix**:
  1. Add godoc to each function: `// GetFantasyTreeTemplates returns skill tree templates themed for fantasy genre worlds, including warrior, mage, ranger, and healer trees.`
- **Verify**: `grep -n "func Get.*Templates" pkg/procgen/skills/templates.go`

#### REM-095: Station generator missing doc.go for StationType enum
- **Source**: `pkg/procgen/station/AUDIT.md`
- **Location**: `pkg/procgen/station/generator.go:20-34`
- **Problem**: Missing package overview explaining `StationType` enum mapping to recipe types.
- **Fix**:
  1. Create or update `doc.go` with: `// Package station generates crafting station configurations. StationType enum maps to recipe categories: Forge→weapons/armor, Alchemy→potions, Enchanting→enchantments, etc.`
- **Verify**: `head -10 pkg/procgen/station/doc.go 2>/dev/null || echo "no doc.go"`

#### REM-096: Story hard-coded content templates limit variety
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/generator.go:170-247`, `archaeology.go:209-262`, `timeline.go:241-320`
- **Problem**: Story content uses fixed template arrays with hard-coded strings, limiting narrative variety and preventing translation.
- **Fix**:
  1. Extract template strings to data-driven tables (e.g., `var beginningTemplates = map[string][]string{...}`)
  2. Consider integration with `pkg/procgen/dialog/` Markov chain system for more variety
  3. Add genre as a key in template maps for genre-specific narratives
- **Verify**: `grep -c "\".*fragment\|\".*story\|\".*tale" pkg/procgen/story/generator.go`

#### REM-097: Story Vector2 type duplicates engine position type
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/types.go:18-26`
- **Problem**: `Vector2` defined locally while `pkg/engine/components.go` has an equivalent.
- **Fix**:
  1. Replace local `Vector2` with import of engine's position type
  2. Or extract to a shared `pkg/math` package if engine import creates unwanted dependency
- **Verify**: `grep -n "type Vector2" pkg/procgen/story/types.go pkg/engine/components.go`

#### REM-098: Story no genre fallback warning
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/generator.go:153-168`
- **Problem**: Genre-specific functions default to generic themes silently for unknown genres.
- **Fix**:
  1. Add: `logrus.WithField("genre", genre).Warn("unknown genre, using default themes")`
- **Verify**: `grep -n "default:" pkg/procgen/story/generator.go`

#### REM-099: Terrain Point.Equals lacks godoc
- **Source**: `pkg/procgen/terrain/AUDIT.md`
- **Location**: `pkg/procgen/terrain/point.go:20`
- **Problem**: `Point.Equals()` method lacks godoc comment.
- **Fix**:
  1. Add: `// Equals returns true if both points have the same X and Y coordinates.`
- **Verify**: `grep -B1 "func.*Equals" pkg/procgen/terrain/point.go`

#### REM-100: Terrain Room.Center and Room.Overlaps lack godoc
- **Source**: `pkg/procgen/terrain/AUDIT.md`
- **Location**: `pkg/procgen/terrain/types.go:310,315`
- **Problem**: Methods lack godoc comments.
- **Fix**:
  1. Add: `// Center returns the center point of the room.` and `// Overlaps returns true if this room's bounding box intersects with other's.`
- **Verify**: `grep -B1 "func.*Center\|func.*Overlaps" pkg/procgen/terrain/types.go`

#### REM-101: Vehicle template functions lack godoc
- **Source**: `pkg/procgen/vehicle/AUDIT.md`
- **Location**: `pkg/procgen/vehicle/templates.go:8,87,168,202,237`
- **Problem**: `GetFantasyTemplates()` and other genre template functions lack godoc.
- **Fix**:
  1. Add godoc to each: `// GetFantasyTemplates returns vehicle templates themed for fantasy settings (wagons, chariots, flying carpets, etc.).`
- **Verify**: `grep -n "func Get.*Templates" pkg/procgen/vehicle/templates.go`

#### REM-102: Rendering cache GenerateAsync lacks context.Context for cancellation
- **Source**: `pkg/rendering/cache/AUDIT.md`
- **Location**: `pkg/rendering/cache/pregenerator.go:134`
- **Problem**: `GenerateAsync` could benefit from `context.Context` for cancellation support.
- **Fix**:
  1. Add `ctx context.Context` as first parameter to `GenerateAsync`
  2. Check `ctx.Done()` in the generation loop
  3. Update call sites to pass appropriate context
- **Verify**: `grep -n "func.*GenerateAsync" pkg/rendering/cache/pregenerator.go`

#### REM-103: Rendering display lacks mobile DPI/safe area support
- **Source**: `pkg/rendering/display/AUDIT.md`
- **Location**: `pkg/rendering/display/config.go:38-49`
- **Problem**: No mobile-specific DPI scaling or safe area inset handling. Only 4 standard 16:9 resolutions supported.
- **Fix**:
  1. Add `GetNearestValidResolution(width, height int) Resolution` helper for non-standard aspect ratios
  2. Add DPI scaling factor field to config
  3. Add safe area inset fields (top, bottom, left, right) for mobile notch/nav bar
- **Verify**: `grep -n "Resolution\|DPI\|SafeArea" pkg/rendering/display/config.go`

#### REM-104: Lighting gpu_bloom_headless.go untested
- **Source**: `pkg/rendering/lighting/AUDIT.md`
- **Location**: `pkg/rendering/lighting/gpu_bloom_headless.go:1-39`
- **Problem**: No automated test verifies the headless stub compiles and provides no-op behavior.
- **Fix**:
  1. Add a test file `gpu_bloom_headless_test.go` with `//go:build headless` tag
  2. Test that all functions return expected zero/nil values
  3. Add headless build to CI: `go build -tags headless ./pkg/rendering/lighting/...`
- **Verify**: `go build -tags headless ./pkg/rendering/lighting/...`

#### REM-105: Lighting EnableShadows deprecated without godoc convention
- **Source**: `pkg/rendering/lighting/AUDIT.md`
- **Location**: `pkg/rendering/lighting/types.go:117-122`
- **Problem**: `LightingConfig.EnableShadows` marked deprecated but lacks `// Deprecated:` godoc convention for static analysis.
- **Fix**:
  1. Change comment to: `// Deprecated: EnableShadows is no longer used. Shadow configuration is now per-light. This field will be removed in a future version.`
- **Verify**: `grep -n "EnableShadows\|Deprecated" pkg/rendering/lighting/types.go`

#### REM-106: Rendering palette utils.go helpers should be shared
- **Source**: `pkg/rendering/palette/AUDIT.md`
- **Location**: `pkg/rendering/palette/utils.go:8-36`
- **Problem**: `clamp`, `max`, `min` helpers could be shared across packages.
- **Fix**:
  1. For Go 1.21+, replace `max`/`min` with built-in `max()`/`min()`
  2. For `clamp`, use `max(minVal, min(maxVal, val))` with built-ins
  3. Or keep package-local if minimal and well-tested
- **Verify**: `grep -rn "func clamp\|func max\|func min" pkg/rendering/palette/utils.go`

#### REM-107: Rendering patterns applyGenreVariations mutates config
- **Source**: `pkg/rendering/patterns/AUDIT.md`
- **Location**: `pkg/rendering/patterns/generator.go:99`
- **Problem**: Function mutates input config's `Scale` and `DetailLevel` fields without documenting side effect.
- **Fix**:
  1. Add godoc: `// applyGenreVariations modifies config.Scale and config.DetailLevel in-place based on the specified genre.`
  2. Or return a modified copy instead of mutating
- **Verify**: `grep -B2 "func.*applyGenreVariations" pkg/rendering/patterns/generator.go`

#### REM-108: Rendering patterns cellularNoise double hash calculation
- **Source**: `pkg/rendering/patterns/AUDIT.md`
- **Location**: `pkg/rendering/patterns/generator.go:465-468`
- **Problem**: Hash calculation done once then bit-shifted twice; could micro-optimize.
- **Fix**:
  1. Pre-compute both offsets from the single hash value in one pass
  2. Note: current performance is already 1-2ms per 32x32 texture, so this is low priority
- **Verify**: `sed -n '465,468p' pkg/rendering/patterns/generator.go`

#### REM-109: Rendering quality PerformanceMonitor uses time.Now()
- **Source**: `pkg/rendering/quality/AUDIT.md`
- **Location**: `pkg/rendering/quality/performance_monitor.go:48,125,138`
- **Problem**: Uses `time.Now()` for performance tracking instead of `GameClock` interface.
- **Fix**:
  1. Add `GameClock` field to `PerformanceMonitor`
  2. Accept in constructor, default to system time
  3. Replace `time.Now()` with `p.clock.Now()`
- **Verify**: `grep -n "time.Now" pkg/rendering/quality/performance_monitor.go`

#### REM-110: Rendering shapes Shape.Type field is redundant
- **Source**: `pkg/rendering/shapes/AUDIT.md`
- **Location**: `pkg/rendering/shapes/types.go:132-133`
- **Problem**: `Shape.Type` field is redundant with `Config.Type` and unused throughout codebase.
- **Fix**:
  1. Remove `Type` field from `Shape` struct
  2. Or deprecate: `// Deprecated: Use Config.Type instead. Will be removed in next major version.`
- **Verify**: `grep -rn "\.Type" pkg/rendering/shapes/ | grep -v Config | grep -v _test.go | grep -v AUDIT`

#### REM-111: Rendering shapes missing String() switch cases
- **Source**: `pkg/rendering/shapes/AUDIT.md`
- **Location**: `pkg/rendering/shapes/types.go:107-120`
- **Problem**: `ShapeEllipse`, `ShapeCapsule`, `ShapeBean`, `ShapeWedge`, `ShapeShield`, `ShapeBlade`, `ShapeSkull` missing from `String()` switch cases.
- **Fix**:
  1. Add missing cases to the `String()` method switch statement
  2. Example: `case ShapeEllipse: return "ellipse"`
- **Verify**: `grep -n "case Shape" pkg/rendering/shapes/types.go`

#### REM-112: Rendering shapes repeated math.Sqrt/Pow in hot path
- **Source**: `pkg/rendering/shapes/AUDIT.md`
- **Location**: `pkg/rendering/shapes/generator.go:216,371,446`
- **Problem**: `inCircle`, `inEllipse`, `inBean` use repeated `math.Sqrt`/`math.Pow` which could be cached.
- **Fix**:
  1. Pre-compute squared distances and compare against squared thresholds to avoid sqrt
  2. Cache pow results when called in loops with same base
- **Verify**: `grep -n "math.Sqrt\|math.Pow" pkg/rendering/shapes/generator.go`

#### REM-113: Rendering sprites exported functions lack godoc
- **Source**: `pkg/rendering/sprites/AUDIT.md`
- **Location**: `pkg/rendering/sprites/generator.go:39-42,44-57`, `equipment.go:4-95,98-114`
- **Problem**: Multiple exported helper functions lack godoc comments.
- **Fix**:
  1. Add godoc to each exported function in `generator.go` and `equipment.go`
  2. Cover parameters, return values, and usage context
- **Verify**: `grep -n "^func [A-Z]" pkg/rendering/sprites/generator.go pkg/rendering/sprites/equipment.go`

#### REM-114: Rendering tiles doc.go references unknown phase numbers
- **Source**: `pkg/rendering/tiles/AUDIT.md`
- **Location**: `pkg/rendering/tiles/doc.go:13,17,22,28,44,56`
- **Problem**: References "Phase 16.2", "Phase 16.3", "Phase 47" which lack context for new developers.
- **Fix**:
  1. Replace phase references with descriptive names (e.g., "Phase 16.2" → "tile transition system")
  2. Or add a glossary comment mapping phase numbers to feature names
- **Verify**: `grep -n "Phase" pkg/rendering/tiles/doc.go`

#### REM-115: Saveload SaveManager lacks type-level godoc
- **Source**: `pkg/saveload/AUDIT.md`
- **Location**: `pkg/saveload/manager.go:24`
- **Problem**: `SaveManager` struct lacks a type-level godoc summary.
- **Fix**:
  1. Add: `// SaveManager handles game state persistence including save, load, auto-save, and migration. Platform-specific implementations exist for desktop (file-based) and WASM (localStorage-based).`
- **Verify**: `grep -B2 "type SaveManager struct" pkg/saveload/manager.go`

#### REM-116: WASM save migration unsupported
- **Source**: `pkg/saveload/AUDIT.md`
- **Location**: `pkg/saveload/storage_wasm.go:97-99`
- **Problem**: WASM explicitly does not support save file migration. Users upgrading game versions on WASM lose saves.
- **Fix**:
  1. Add prominent documentation in WASM storage: `// NOTE: WASM storage does not support migration. Save files from incompatible versions will be rejected.`
  2. Document in user-facing docs (FAQ or WASM guide)
  3. Consider adding basic version-compatible migration in future
- **Verify**: `grep -n "migration\|migrate\|incompatible" pkg/saveload/storage_wasm.go`

#### REM-117: MemorySaveManager.SetMigrator is silent no-op
- **Source**: `pkg/saveload/AUDIT.md`
- **Location**: `pkg/saveload/memory_manager.go:358`
- **Problem**: `SetMigrator(_ Migrator) {}` silently discards the migrator, which may confuse API users.
- **Fix**:
  1. Add warning log: `logrus.Warn("MemorySaveManager.SetMigrator: in-memory storage does not support migration, migrator ignored")`
- **Verify**: `grep -n "SetMigrator" pkg/saveload/memory_manager.go`

#### REM-118: Version PrintVersion uses fmt.Println instead of structured logging
- **Source**: `pkg/version/AUDIT.md`
- **Location**: `pkg/version/version.go:71`
- **Problem**: `PrintVersion()` uses `fmt.Println` instead of logrus, preventing log level filtering and correlation.
- **Fix**:
  1. Replace `fmt.Println` with `logrus.WithFields(logrus.Fields{"version": Version, "go_version": GoVersion}).Info("version info")`
  2. Or keep `fmt.Println` for CLI output and add a separate `LogVersion()` for structured logging
- **Verify**: `grep -n "fmt.Println\|fmt.Printf" pkg/version/version.go`

#### REM-119: Version constants lack godoc
- **Source**: `pkg/version/AUDIT.md`
- **Location**: `pkg/version/version.go:22-29`
- **Problem**: Constants `Major`, `Minor`, `Patch` lack individual godoc explaining when to increment.
- **Fix**:
  1. Add godoc:
     ```go
     // Major is incremented for breaking API changes.
     // Minor is incremented for backward-compatible feature additions.
     // Patch is incremented for backward-compatible bug fixes.
     ```
- **Verify**: `grep -B1 "Major\|Minor\|Patch" pkg/version/version.go`

---

### LOW

#### REM-120: Determinism audit checklist items (template)
- **Source**: `docs/DETERMINISM_AUDIT.md`
- **Location**: `docs/DETERMINISM_AUDIT.md:227-231`
- **Problem**: Generic checklist template items for determinism validation remain unchecked (accepts seed parameter, creates local RNG, never uses global rand, added to test suite, passes 1000-run test).
- **Fix**:
  1. These are template items for future audits — if the template is complete, mark them as checked
  2. If audits are ongoing, track each generator's compliance against these criteria
- **Verify**: `grep -c '\- \[ \]' docs/DETERMINISM_AUDIT.md`

#### REM-121: Meta audit template placeholders
- **Source**: `docs/META_AUDIT.md`
- **Location**: `docs/META_AUDIT.md:238-244,327-335`
- **Problem**: Template placeholder items and per-package checklist items remain unchecked.
- **Fix**:
  1. Template items (lines 238-244) are intentionally unchecked as they're copy-paste templates
  2. Per-package checklist items (327-335) should be checked as each package audit is completed
  3. Run through each checklist item verifying completion
- **Verify**: `grep -c '\- \[ \]' docs/META_AUDIT.md`

#### REM-122: Choice consequences helpers could use stdlib
- **Source**: `pkg/integration/choice_consequences/AUDIT.md`
- **Location**: `pkg/integration/choice_consequences/helpers.go:9,20`
- **Problem**: Custom `abs` and `clamp` helper functions duplicate standard library functionality.
- **Fix**:
  1. Replace `abs(x float64)` with `math.Abs(x)`
  2. Replace `clamp(val, min, max float64)` with `max(min, min(max, val))` using Go 1.21+ built-ins
- **Verify**: `grep -n "func abs\|func clamp" pkg/integration/choice_consequences/helpers.go`

#### REM-123: Companion housing types.go duplicate doc
- **Source**: `pkg/integration/companion_housing/AUDIT.md`
- **Location**: `pkg/integration/companion_housing/types.go:1-23`
- **Problem**: Package comment in types.go duplicates doc.go information.
- **Fix**:
  1. Reduce types.go comment to: `// types.go defines data types for the companion housing integration. See doc.go for package overview.`
- **Verify**: `head -5 pkg/integration/companion_housing/types.go`

#### REM-124: Guild vehicle TimeProvider duplication across packages
- **Source**: `pkg/integration/guild_vehicle/AUDIT.md`
- **Location**: `pkg/integration/guild_vehicle/time_provider.go:1-55`
- **Problem**: TimeProvider pattern duplicated across multiple packages.
- **Fix**:
  1. Extract to a shared `pkg/time/provider.go` package
  2. Or accept duplication as intentional to avoid cross-package coupling
  3. Document decision in `time_provider.go`: `// TimeProvider is package-local by design to avoid coupling. The same pattern exists in cmd/client/ and cmd/server/.`
- **Verify**: `grep -rn "type TimeProvider interface" pkg/ cmd/ | grep -v _test.go | grep -v AUDIT`

#### REM-125: Guild vehicle doc.go has stale PLANNED section
- **Source**: `pkg/integration/guild_vehicle/AUDIT.md`
- **Location**: `pkg/integration/guild_vehicle/doc.go:64-68`
- **Problem**: States "PLANNED (not yet implemented)" integrations without timeline or tracking issues.
- **Fix**:
  1. Either add GitHub issue links for planned features
  2. Or remove the PLANNED section if integrations are deferred indefinitely
- **Verify**: `grep -n "PLANNED" pkg/integration/guild_vehicle/doc.go`

#### REM-126: Housing crafting missing V9ValidationService example
- **Source**: `pkg/integration/housing_crafting/AUDIT.md`
- **Location**: `pkg/integration/housing_crafting/doc.go`
- **Problem**: Missing example code for V9ValidationService integration pattern.
- **Fix**:
  1. Add example in doc.go showing how V9ValidationService wraps StationManager
- **Verify**: `grep -n "V9Validation\|example" pkg/integration/housing_crafting/doc.go`

#### REM-127: Housing crafting types enum String() doesn't document invalid cases
- **Source**: `pkg/integration/housing_crafting/AUDIT.md`
- **Location**: `pkg/integration/housing_crafting/types.go:22-43`
- **Problem**: `QualityTier` and `StationType` `String()` don't document invalid case handling.
- **Fix**:
  1. Add godoc: `// String returns the human-readable name. Returns "unknown" for invalid values.`
- **Verify**: `grep -n "func.*String()" pkg/integration/housing_crafting/types.go`

#### REM-128: Narrative world no mod integration
- **Source**: `pkg/integration/narrative_world/AUDIT.md`
- **Location**: `pkg/integration/narrative_world/manager.go`
- **Problem**: No integration with `pkg/modding` for customizable quest templates, conflict probabilities, or consequence rules.
- **Fix**:
  1. Add `ModRuleProvider` field to manager
  2. Check mod overrides before using hardcoded values for probabilities and templates
  3. Or document this as a future enhancement in doc.go
- **Verify**: `grep -n "modding\|ModRule" pkg/integration/narrative_world/manager.go`

#### REM-129: Narrative world no network sync for companion state
- **Source**: `pkg/integration/narrative_world/AUDIT.md`
- **Location**: `pkg/integration/narrative_world/manager.go`, `serialization.go`
- **Problem**: No network snapshot system integration for cross-server companion state despite having Serialize/Deserialize.
- **Fix**:
  1. Wire `StoryEventManager` serialization into network snapshot system
  2. Or document as a future enhancement if cross-server narrative continuity is not yet required
- **Verify**: `grep -n "snapshot\|Snapshot\|network" pkg/integration/narrative_world/serialization.go`

#### REM-130: Trade routes RouteManager.priceHandler undocumented
- **Source**: `pkg/integration/trade_routes/AUDIT.md`
- **Location**: `pkg/integration/trade_routes/manager.go:72`
- **Problem**: `priceHandler` field added later without updating struct godoc.
- **Fix**:
  1. Add field comment: `// priceHandler applies buy/sell price impacts when routes complete.`
- **Verify**: `grep -n "priceHandler" pkg/integration/trade_routes/manager.go`

#### REM-131: World events tests use time.Now instead of TimeProvider
- **Source**: `pkg/integration/world_events/AUDIT.md`
- **Location**: `pkg/integration/world_events/events_test.go:206,299,492,655,674`, `manager_test.go` (multiple)
- **Problem**: Tests use `time.Now()` directly despite package having `TimeProvider` abstraction.
- **Fix**:
  1. Replace `time.Now()` in tests with mock time provider
  2. Use `mockTimeProvider := &MockTimeProvider{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}`
- **Verify**: `grep -n "time.Now" pkg/integration/world_events/*_test.go | wc -l`

#### REM-132: World events CenterX/CenterY fields use inline comments
- **Source**: `pkg/integration/world_events/AUDIT.md`
- **Location**: `pkg/integration/world_events/types.go:68-71`
- **Problem**: Fields have inline comments instead of full godoc format.
- **Fix**:
  1. Move inline comments to godoc-style comments above each field
- **Verify**: `sed -n '68,71p' pkg/integration/world_events/types.go`

#### REM-133: Memprofile fmt.Printf usage
- **Source**: `pkg/memprofile/AUDIT.md`
- **Location**: `pkg/memprofile/profile.go:195-225`
- **Problem**: `PrintProfile()` uses `fmt.Printf` instead of structured logging. Intentional for human-readable output.
- **Fix**:
  1. Add optional `ExportJSON() ([]byte, error)` method for structured export
  2. Keep `PrintProfile()` for CLI usage with godoc noting it uses stdout directly
- **Verify**: `grep -n "fmt.Printf\|fmt.Println" pkg/memprofile/profile.go`

#### REM-134: Modding doc.go time.Now rationale could be clearer
- **Source**: `pkg/modding/AUDIT.md`
- **Location**: `pkg/modding/doc.go:113-120`
- **Problem**: Exception rationale for non-deterministic time.Now usage could be more explicit.
- **Fix**:
  1. Update doc.go text: `// This package uses time.Now() for server-side operational behavior (rate limiting, mod load timestamps) that is NOT replicated to clients and does NOT affect deterministic game state.`
- **Verify**: `sed -n '113,120p' pkg/modding/doc.go`

#### REM-135: Network chat generateMessageID lacks documentation
- **Source**: `pkg/network/chat/AUDIT.md`
- **Location**: `pkg/network/chat/system.go:132`
- **Problem**: Unexported function lacks documentation explaining collision resistance.
- **Fix**:
  1. Add comment: `// generateMessageID creates a 128-bit random identifier encoded as hex. Collision probability is ~1/2^64 with birthday paradox for practical message volumes (<10^9).`
- **Verify**: `grep -B1 "func generateMessageID" pkg/network/chat/system.go`

#### REM-136: Federation guild files lack file-level comments
- **Source**: `pkg/network/federation/guild/AUDIT.md`
- **Location**: `pkg/network/federation/guild/constants.go:1`, `treasury.go:1`, `persistence.go:1`
- **Problem**: Individual files lack comments explaining their purpose.
- **Fix**:
  1. Add file-level comments: `// constants.go defines constants for cross-server guild federation.`, `// treasury.go implements federated guild treasury operations.`, `// persistence.go handles guild federation state persistence.`
- **Verify**: `head -3 pkg/network/federation/guild/constants.go pkg/network/federation/guild/treasury.go pkg/network/federation/guild/persistence.go`

#### REM-137: Federation mobile exported symbols lack godoc
- **Source**: `pkg/network/federation/mobile/AUDIT.md`
- **Location**: `pkg/network/federation/mobile/types.go:221-232`
- **Problem**: `GetBytesAvailable` and `SetBytesAvailable` lack godoc comments.
- **Fix**:
  1. Add godoc: `// GetBytesAvailable returns the current token bucket bandwidth allowance in bytes.`
  2. Add: `// SetBytesAvailable updates the token bucket bandwidth allowance.`
- **Verify**: `grep -B1 "func.*GetBytesAvailable\|func.*SetBytesAvailable" pkg/network/federation/mobile/types.go`

#### REM-138: Federation mobile State.bytesAvailable undocumented
- **Source**: `pkg/network/federation/mobile/AUDIT.md`
- **Location**: `pkg/network/federation/mobile/types.go:119`
- **Problem**: Internal token bucket state field lacks documentation.
- **Fix**:
  1. Add comment: `// bytesAvailable is the internal token bucket state tracking remaining bandwidth allowance. Accessed via GetBytesAvailable/SetBytesAvailable.`
- **Verify**: `grep -n "bytesAvailable" pkg/network/federation/mobile/types.go`

#### REM-139: Federation webrtc signaling overflow protection undocumented
- **Source**: `pkg/network/federation/webrtc/AUDIT.md`
- **Location**: `pkg/network/federation/webrtc/signaling.go:367-377`
- **Problem**: Round-robin counter overflow protection at line 374-376 exists but is not documented.
- **Fix**:
  1. Add inline comment: `// Counter overflow protection: reset to 0 when approaching INT_MAX to prevent wraparound on long-lived servers.`
- **Verify**: `sed -n '367,377p' pkg/network/federation/webrtc/signaling.go`

#### REM-140: Procgen companion doc.go missing integration references
- **Source**: `pkg/procgen/companion/AUDIT.md`
- **Location**: `pkg/procgen/companion/doc.go:1-96`
- **Problem**: No mention that package integrates with `CompanionAISystem`, `CompanionLearningSystem`, and housing systems.
- **Fix**:
  1. Add "Integration" section to doc.go listing: `CompanionAISystem`, `CompanionLearningSystem`, `pkg/integration/companion_housing/`
- **Verify**: `grep -n "Integration\|CompanionAI\|CompanionLearning" pkg/procgen/companion/doc.go`

#### REM-141: Dialog GetGreeting backward compatibility unclear
- **Source**: `pkg/procgen/dialog/AUDIT.md`
- **Location**: `pkg/procgen/dialog/personality.go:205`
- **Problem**: `GetGreeting()` method comment could clarify backward compatibility vs. `GetGreetingWithSeed()`.
- **Fix**:
  1. Add godoc: `// GetGreeting returns a greeting using the personality's internal RNG state. For deterministic greetings, use GetGreetingWithSeed instead. Maintained for backward compatibility.`
- **Verify**: `grep -B2 "func.*GetGreeting()" pkg/procgen/dialog/personality.go`

#### REM-142: Entity doc.go missing integration references
- **Source**: `pkg/procgen/entity/AUDIT.md`
- **Location**: `pkg/procgen/entity/doc.go:1-40`
- **Problem**: Missing mention of integration with `pkg/engine/entity_spawning.go` and `merchant_spawn.go`.
- **Fix**:
  1. Add integration section to doc.go
- **Verify**: `grep -n "entity_spawning\|merchant_spawn" pkg/procgen/entity/doc.go`

#### REM-143: Puzzle helper functions only used by one caller
- **Source**: `pkg/procgen/puzzle/AUDIT.md`
- **Location**: `pkg/procgen/puzzle/generator.go:604-663`
- **Problem**: Multiple helpers only used by `generateColorMatchingPuzzle`. Could be inlined or moved to closures.
- **Fix**:
  1. Convert to closures inside `generateColorMatchingPuzzle` for locality
  2. Or keep as-is with a comment: `// Helper functions for generateColorMatchingPuzzle`
- **Verify**: `grep -n "calculateElementCount\|createColoredElements\|selectTargetColors\|buildColorMatchSolution" pkg/procgen/puzzle/generator.go`

#### REM-144: Quest genre template functions too large
- **Source**: `pkg/procgen/quest/AUDIT.md`
- **Location**: `pkg/procgen/quest/types.go:281-635`
- **Problem**: Large genre template functions (200+ lines) could be data-driven tables.
- **Fix**:
  1. Extract template strings to `var questTemplates = map[string][]QuestTemplate{...}`
  2. Replace function bodies with map lookups
- **Verify**: `wc -l pkg/procgen/quest/types.go`

#### REM-145: Recipe template registration hardcoded
- **Source**: `pkg/procgen/recipe/AUDIT.md`
- **Location**: `pkg/procgen/recipe/generator.go:374-765`
- **Problem**: Templates hardcoded rather than loading from mod system.
- **Fix**:
  1. Add `ModRuleProvider` integration for custom templates
  2. Or document in doc.go that template extension is via code contributions (intentional design)
- **Verify**: `grep -n "mod\|Mod" pkg/procgen/recipe/generator.go`

#### REM-146: Skills normalizeGenre could be shared
- **Source**: `pkg/procgen/skills/AUDIT.md`
- **Location**: `pkg/procgen/skills/generator.go:125`
- **Problem**: `normalizeGenre` is unexported but could be shared across generators.
- **Fix**:
  1. Extract to `pkg/procgen/genre/normalize.go` as `NormalizeGenre`
  2. Or keep package-local if genre normalization differs per generator
- **Verify**: `grep -rn "normalizeGenre\|NormalizeGenre" pkg/procgen/`

#### REM-147: Story benchmarks exist but undocumented
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/procgen/story/*_test.go`
- **Problem**: 10 benchmark functions exist but are not documented in doc.go or CI validation.
- **Fix**:
  1. Add benchmarks section to doc.go: `// Benchmarks: Run with go test -bench=. -benchmem ./pkg/procgen/story/...`
  2. Add to CI benchmark workflow if not already included
- **Verify**: `grep -n "Benchmark" pkg/procgen/story/*_test.go | wc -l`

#### REM-148: Story helper functions in engine lack godoc
- **Source**: `pkg/procgen/story/AUDIT.md`
- **Location**: `pkg/engine/story_fragment_component.go:97-161`
- **Problem**: ECS-compliant helper functions like `JournalAddDiscovery`, `JournalIsDiscovered` lack individual godoc.
- **Fix**:
  1. Add godoc to each function explaining parameters and return values
- **Verify**: `grep -n "func Journal" pkg/engine/story_fragment_component.go`

#### REM-149: Rendering cache time.Now documentation
- **Source**: `pkg/rendering/cache/AUDIT.md`
- **Location**: `pkg/rendering/cache/memory_monitor.go:142,156`
- **Problem**: `time.Now()` usage could be documented as acceptable for monitoring.
- **Fix**:
  1. Add inline comment: `// time.Now() is acceptable here for non-deterministic performance monitoring.`
- **Verify**: `grep -n "time.Now" pkg/rendering/cache/memory_monitor.go`

#### REM-150: Rendering cache PredictiveWarmerConfig validation unclear
- **Source**: `pkg/rendering/cache/AUDIT.md`
- **Location**: `pkg/rendering/cache/predictive_warmer.go:65-73`
- **Problem**: Config validation could be more explicit about clamping to defaults.
- **Fix**:
  1. Add comments explaining each clamp: `// Default to X if not specified or out of range.`
- **Verify**: `sed -n '65,73p' pkg/rendering/cache/predictive_warmer.go`

#### REM-151: Rendering cache predictNextLocked RLock requirement
- **Source**: `pkg/rendering/cache/AUDIT.md`
- **Location**: `pkg/rendering/cache/predictive_warmer.go:162`
- **Problem**: Comment could clarify RLock requirement.
- **Fix**:
  1. Add: `// predictNextLocked must be called with w.mu held for reading (RLock).`
- **Verify**: `grep -B2 "func.*predictNextLocked" pkg/rendering/cache/predictive_warmer.go`

#### REM-152: Rendering particles ApplyPhysics lacks inline algorithm comments
- **Source**: `pkg/rendering/particles/AUDIT.md`
- **Location**: `pkg/rendering/particles/behaviors.go:92`
- **Problem**: Complex physics functions lack inline algorithm comments.
- **Fix**:
  1. Add inline comments explaining physics formulas (gravity, drag, wind effects)
- **Verify**: `grep -n "func.*ApplyPhysics" pkg/rendering/particles/behaviors.go`

#### REM-153: Rendering particles weather algorithms undocumented
- **Source**: `pkg/rendering/particles/AUDIT.md`
- **Location**: `pkg/rendering/particles/weather.go:300-400`
- **Problem**: Puddle accumulation and snow drift algorithms lack inline comments.
- **Fix**:
  1. Add inline comments explaining mathematical approach for puddle/snow simulation
- **Verify**: `sed -n '300,310p' pkg/rendering/particles/weather.go`

#### REM-154: Rendering postprocess ApplyPrismaticAberration lacks param docs
- **Source**: `pkg/rendering/postprocess/AUDIT.md`
- **Location**: `pkg/rendering/postprocess/chromatic_aberration.go:91`
- **Problem**: `ApplyPrismaticAberration` lacks detailed parameter documentation.
- **Fix**:
  1. Add godoc: `// ApplyPrismaticAberration applies angle-based chromatic separation. angle is in radians [0, 2π]. intensity controls displacement magnitude [0.0, 1.0].`
- **Verify**: `grep -B2 "func.*ApplyPrismaticAberration" pkg/rendering/postprocess/chromatic_aberration.go`

#### REM-155: Rendering pool tests lack headless skip logic
- **Source**: `pkg/rendering/pool/AUDIT.md`
- **Location**: `pkg/rendering/pool/image_pool_test.go:1`
- **Problem**: Tests require X11/Wayland but lack build tags or skip logic for headless environments.
- **Fix**:
  1. Add `//go:build !headless` tag to test file, or
  2. Add skip logic: `if os.Getenv("DISPLAY") == "" { t.Skip("requires display") }`
- **Verify**: `head -5 pkg/rendering/pool/image_pool_test.go`

#### REM-156: Rendering UI time.Now in tests
- **Source**: `pkg/rendering/ui/AUDIT.md`
- **Location**: `pkg/rendering/ui/chat_test.go`, `notifications_test.go`, `trade_test.go`
- **Problem**: Tests use `time.Now()` for data initialization instead of deterministic time.
- **Fix**:
  1. Replace with fixed time: `testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)`
- **Verify**: `grep -n "time.Now" pkg/rendering/ui/*_test.go`

#### REM-157: Rendering UI package scope documentation
- **Source**: `pkg/rendering/ui/AUDIT.md`
- **Location**: `pkg/rendering/ui/` (general)
- **Problem**: Package does not implement ECS interfaces; docs should clarify it provides helpers used by systems.
- **Fix**:
  1. Add to doc.go: `// Package ui provides rendering utilities used by ECS systems. It does not implement System or Component interfaces directly.`
- **Verify**: `head -5 pkg/rendering/ui/doc.go`

#### REM-158: Social persistence ImageGallery type not in types.go
- **Source**: `pkg/social/persistence/AUDIT.md`
- **Location**: `pkg/social/persistence/image_gallery.go:371-382`
- **Problem**: `ImageThumbnail` type returned by `GetThumbnails()` not defined in `types.go`.
- **Fix**:
  1. Move `ImageThumbnail` definition to `types.go` for consistency
- **Verify**: `grep -n "type ImageThumbnail" pkg/social/persistence/*.go`

#### REM-159: Social persistence doc.go truncated in go doc
- **Source**: `pkg/social/persistence/AUDIT.md`
- **Location**: `pkg/social/persistence/doc.go`
- **Problem**: Excellent doc.go but truncated in `go doc` output. Needs cross-references to integration points.
- **Fix**:
  1. Add "See Also" section at top of doc.go with cross-references
- **Verify**: `go doc ./pkg/social/persistence/ 2>&1 | tail -5`

#### REM-160: Visualtest helper functions lack inline comments
- **Source**: `pkg/visualtest/AUDIT.md`
- **Location**: `pkg/visualtest/snapshot.go`, `genre.go:220-250`
- **Problem**: `calculateSimilarity` and `extractDominantColors` lack inline algorithm comments.
- **Fix**:
  1. Add inline comments explaining the similarity comparison algorithm and color extraction approach
- **Verify**: `grep -n "func.*calculateSimilarity\|func.*extractDominantColors" pkg/visualtest/snapshot.go pkg/visualtest/genre.go`

#### REM-161: Visualtest parity maxUint8 helper could be shared
- **Source**: `pkg/visualtest/parity/AUDIT.md`
- **Location**: `pkg/visualtest/parity/validator.go:329-337`
- **Problem**: `maxUint8` helper function could be made more generic or moved to utility package.
- **Fix**:
  1. Replace with Go 1.21+ built-in `max()` function
  2. Or keep package-local with comment: `// maxUint8 is package-local to avoid external dependencies in test infrastructure.`
- **Verify**: `grep -n "func maxUint8" pkg/visualtest/parity/validator.go`

#### REM-162: World housing doc.go lacks architecture overview
- **Source**: `pkg/world/housing/AUDIT.md`
- **Location**: `pkg/world/housing/doc.go:1-101`
- **Problem**: Missing system architecture overview beyond basic usage.
- **Fix**:
  1. Add architecture section to doc.go: describe HousingManager, BlueprintManager, SpatialManager, and their relationships
- **Verify**: `grep -n "Architecture\|architecture" pkg/world/housing/doc.go`

#### REM-163: World housing BuildingSize type lacks godoc
- **Source**: `pkg/world/housing/AUDIT.md`
- **Location**: `pkg/world/housing/types.go:~52-62`
- **Problem**: `BuildingSize` type and constants lack godoc comments.
- **Fix**:
  1. Add godoc: `// BuildingSize represents the footprint category of a housing structure.`
  2. Document each constant (Small, Medium, Large, etc.)
- **Verify**: `grep -n "BuildingSize" pkg/world/housing/types.go`

#### REM-164: World housing Vector2 lacks godoc
- **Source**: `pkg/world/housing/AUDIT.md`
- **Location**: `pkg/world/housing/types.go:~64-68`
- **Problem**: `Vector2` type lacks godoc comment.
- **Fix**:
  1. Add: `// Vector2 represents a 2D position or direction in the housing coordinate system.`
- **Verify**: `grep -B1 "type Vector2" pkg/world/housing/types.go`

#### REM-165: World housing CreateHouse accepts interface{}
- **Source**: `pkg/world/housing/AUDIT.md`
- **Location**: `pkg/world/housing/manager.go:177`
- **Problem**: `CreateHouse` accepts `interface{}` for buildingData instead of typed struct.
- **Fix**:
  1. Define a `BuildingData` struct type with the expected fields
  2. Replace `interface{}` parameter with `BuildingData`
  3. Or add a type assertion with proper error handling if flexibility is needed
- **Verify**: `grep -n "func.*CreateHouse" pkg/world/housing/manager.go`

#### REM-166: World raids missing genre blending support
- **Source**: `pkg/world/raids/AUDIT.md`
- **Location**: `pkg/world/raids/` (BossNameGenerator)
- **Problem**: Boss name generation supports single genres but not blended genres from `pkg/procgen/genre`.
- **Fix**:
  1. Add genre weight parameter to `BossNameGenerator`
  2. Support hybrid themes (e.g., 70% fantasy + 30% horror)
- **Verify**: `grep -n "genre\|Genre" pkg/world/raids/*.go | grep -i "blend\|weight"`

#### REM-167: World territory hard-coded capture radius
- **Source**: `pkg/world/territory/AUDIT.md`
- **Location**: `pkg/world/territory/manager.go:34`
- **Problem**: `captureRadius: 50.0` is hardcoded without constructor parameter or genre-based adjustment.
- **Fix**:
  1. Add `captureRadius float64` parameter to `NewManager()`
  2. Accept from config or `GenerationParams`
  3. Default to 50.0 if not specified
- **Verify**: `grep -n "captureRadius.*50" pkg/world/territory/manager.go`

#### REM-168: World territory defensive copy overhead
- **Source**: `pkg/world/territory/AUDIT.md`
- **Location**: `pkg/world/territory/manager.go:71-83,483-492`, `siege.go:381-393,427-444`
- **Problem**: All getter methods return deep copies, creating ~6000 allocations/second for 100 territories at 60 FPS, causing GC pressure.
- **Fix**:
  1. Define read-only view interfaces: `TerritoryView` (getters only, no setters)
  2. Return views instead of copies for read-only access
  3. Use copy-on-write: copy only when mutation is needed
  4. Or add a per-frame snapshot cache: `GetTerritoriesSnapshot()` called once per frame
- **Verify**: `grep -n "copyTerritory\|copySiege" pkg/world/territory/manager.go pkg/world/territory/siege.go`

---

## Completion Criteria

- [x] All REM-### items implemented
- [x] All verification commands pass
- [x] No `- [ ]` items remain in any `*AUDIT*.md`
