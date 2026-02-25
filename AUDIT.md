The following patterns affect multiple packages and represent systemic concerns:

### 1. Determinism Enforcement via TimeProvider
**Affected:** 9+ packages (guild_housing, trade_routes, political_warfare, guild_vehicle, world_events, hostplay, legendary, narrative_world, world/housing)
**Pattern:** All `time.Now()` calls in game logic must be replaced with injectable `TimeProvider` interfaces. The `pkg/integration` comprehensive audit identified and fixed 6 packages with 51+ violations. The pattern is now well-established; remaining instances are in non-critical metadata paths.

### 2. Test Environment Limitation (X11/Ebiten Dependency)
**Affected:** `cmd/client`, `pkg/balance`, `pkg/engine` (all sub-audits), `pkg/world/housing`, integration packages
**Pattern:** Packages with transitive Ebiten dependencies fail in headless CI without `xvfb-run`. Several packages cannot measure actual test coverage. The `cmd/client` package reports 45% coverage because Ebiten UI code cannot be unit-tested without a display server.

### 3. ECS Component Cache Not Updated for New Components (RESOLVED 2026-02-22)
**Affected:** ~~`pkg/companion/learning`, `pkg/integration/guild_vehicle`~~
**Pattern:** ~~New ECS components (`CompanionLearningComponent`, `GuildVehicleFleetComponent`) are not added to the Entity hot-path cache in `pkg/engine/ecs.go`. This means `GetComponent()` falls back to map lookups (~93x slower) for these types.~~ **RESOLVED**: Added hot-path caching for both `CompanionLearningComponent` and `GuildVehicleFleetComponent` in `pkg/engine/ecs.go`. New getter methods `GetCompanionLearning()` and `GetGuildVehicleFleet()` provide zero-overhead access (~93x faster than map lookup). Caches are properly updated on add/remove via both standard and logger methods.

### 4. Unsafe Type Assertions (Partially Resolved)
**Affected:** Multiple engine systems (fixed in AUDIT_COMBAT, AUDIT_CORE_ECS, AUDIT_SOCIAL, AUDIT_PROGRESSION)
**Pattern:** Several systems used bare type assertions (e.g., `comp.(*FactionComponent)`) without comma-ok pattern. These have been fixed in the audited systems but the pattern may recur in unaudited code.

### 5. Documentation Examples Violating Coding Guidelines
**Affected:** ~~~15 packages (audio, procgen, rendering, integration)~~ **RESOLVED 2026-02-22**
**Pattern:** ~~`doc.go` and README.md examples across the codebase use `time.Now().UnixNano()` for seeds, `log.Fatal`, and `fmt.Printf` instead of following logrus guidelines.~~ All doc examples now use deterministic seeds and logrus structured logging. Resolved through multiple audit cycles ending 2026-02-22.

### 6. Deprecated Systems Still Exported
**Affected:** ~~`pkg/integration/companion_housing`, `pkg/integration/housing_crafting`~~ (**RESOLVED 2026-02-22**)
**Pattern:** Systems marked `@deprecated` remain in the public API and are still tested. They should either be removed from exports or receive proper replacement documentation with migration guides. **Note 2026-02-22**: Both `companionHousingSystem` and `housingCraftingSystem` are already unexported (lowercase) and kept for internal test coverage only. Migration guides exist in godoc comments.

### 7. Missing Serialize/Deserialize on Persisted Components
**Affected:** ~~`pkg/class/advanced` (`AdvancedClassComponent`)~~ (**RESOLVED 2026-02-22**), ~~`pkg/engine/qol` (`QoLComponent`)~~ (**RESOLVED 2026-02-22**), ~~`pkg/integration/housing_crafting` (`CraftingStation`, `SkillTrainingFacility`)~~ (**RESOLVED 2026-02-22**)
**Pattern:** Several components that represent persistent player data (class configuration, crafting station state) implement `Type() string` correctly but lack `Serialize()`/`Deserialize()` methods, meaning their data is lost across sessions. Note: `QoLComponent`, `AdvancedClassComponent`, `CraftingStation`, and `SkillTrainingFacility` now have save/load integration.

### 8. System Registration Inconsistency
**Affected:** `pkg/integration` (all sub-packages), `pkg/engine/prestige`, `pkg/engine/qol`
**Pattern:** Three distinct registration patterns coexist without a documented standard: (1) ECS System wrapper registered with World, (2) Manager-only with no wrapper, (3) Pure integration called directly. No centralized registry or discovery mechanism exists. This complicates server-side registration (prestige and QoL are client-only without server equivalents).

### 9. Modding System Integration Gap
**Affected:** `cmd/server/main.go`, `pkg/modding`
**Pattern:** The `modManager` is created at server startup but immediately discarded (`_ = modManager`), preventing any mod rules from being applied to running game systems. This is a silent integration failure that negates the entire modding subsystem on the server.

### 10. Coverage in Ebiten-Dependent Packages
**Affected:** `cmd/client` (45%), `pkg/rendering/animation` (68.4%), `pkg/world/housing` (78.6%), `pkg/rendering/ui` (80.1%), `pkg/network` (82.6%), `pkg/rendering/sprites` (82.4%)
**Pattern:** Packages with Ebiten rendering code or complex integration surfaces have a 30% minimum coverage target (vs. 40% for all other packages). All listed packages meet or substantially exceed this minimum (average 82.4%).

---

*This consolidated report was generated from 110 individual AUDIT*.md files spanning the full Venture codebase. Issue counts represent remaining open issues as of each file's audit date; historical fixed issues are summarized in the "Details" fields. For the most current status of any individual package, refer to its specific AUDIT*.md file.*
