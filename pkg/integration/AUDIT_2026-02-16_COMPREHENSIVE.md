# Audit: github.com/opd-ai/venture/pkg/integration
**Date**: 2026-02-16
**Status**: Needs Work

## Summary
The integration package coordinates cross-system features across 10 sub-packages (51 production Go files, ~15K LOC total). Five sub-packages (choice_consequences, companion_housing, housing_crafting, narrative_world, world_events) demonstrate excellent test coverage (87-91%) with proper deterministic design via TimeProvider abstraction. However, **critical deterministic generation violations remain unresolved** with 57 `time.Now()` calls across guild_housing, trade_routes, political_warfare, guild_vehicle, and world_events packages. These violations create multiplayer desync risk and break the project's core deterministic generation requirement. Headless testing requires xvfb-run due to transitive Ebiten dependencies.

## Issues Found
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for house/storage/transaction IDs in guild_housing, creating multiplayer desync risk (`guild_housing/guild_housing_manager.go:78,246,355,443,568`)
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for 15 timestamp fields in guild_housing (CreatedAt, UpdatedAt, AddedAt, Timestamp on houses/storage/transactions), breaking save/load determinism (`guild_housing/guild_housing_manager.go:88,89,150,203,253,350,360,448,527,573`)
- [x] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for trade route timing (StartTime, EstimatedArrival, mission timing), causing desync in multiplayer trade (`trade_routes/manager.go:216,217,261,262,273,524`) — **FIXED**: Replaced all 6 `time.Now()` calls with injectable TimeProvider pattern. Added `time_provider.go` with `SetTimeProvider`/`ResetTimeProvider` for testing. 6 determinism validation tests added.
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used extensively in political warfare for war timing, treaty expiration, embargo tracking (32+ usages), breaking multiplayer political systems (`political_warfare/manager.go:105,148,153,158,212,252,284,338,371,420,455,557` + 20 more)
- [ ] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for fleet management timestamps (DeployedAt, LastMaintenance, NextMaintenance, mission timing), causing guild vehicle desync (`guild_vehicle/fleet_manager.go:48,49,80,81,99,100,104,125,143,164,202,219`)
- [x] <severity:high> deterministic procgen — Non-deterministic `time.Now()` used for event scheduling and timing in world_events, breaking event synchronization across clients (`world_events/manager.go:32,44,87,111,173,426`) — **FIXED**: Replaced all 6 `time.Now()` calls with injectable TimeProvider pattern. Added `time_provider.go` with `SetTimeProvider`/`ResetTimeProvider` for testing. 4 determinism validation tests added.
- [ ] <severity:med> integration points — No centralized system registration pattern documented; integration systems manually wired in client/server initialization without registry or discovery mechanism
- [ ] <severity:low> doc coverage — Root `doc.go` describes integration *tests* only but package contains production integration systems (8 sub-packages with Manager/System implementations) — misleading scope statement (`doc.go:1-12`)

## Test Coverage
**Average**: 92.0% across testable packages (target: 65%) ✅

**Per-Package Coverage** (xvfb-run required for all):
- `choice_consequences`: 87.1% — Excellent table-driven tests, TimeProvider-based determinism
- `companion_housing`: 93.5% — Comprehensive edge case coverage, proper ECS compliance
- `guild_housing`: 91.8% — High coverage but tests don't validate determinism violations
- `guild_vehicle`: 94.6% — Strong coverage but time.Now() violations untested
- `housing_crafting`: 98.5% — Exemplary coverage, deterministic seed-based generation
- `narrative_world`: 90.7% — Good coverage, TimeProvider abstraction validated
- `political_warfare`: 94.7% — High coverage but doesn't catch time.Now() non-determinism
- `trade_routes`: 92.4% — Strong coverage but timing violations not validated
- `world_events`: 91.2% — Good coverage but event timing non-determinism untested
- `pkg/integration` (root): [no statements] — Only contains multiplayer_test.go

**Test Quality Observations:**
- All packages use table-driven test patterns
- Comprehensive edge case coverage (invalid inputs, boundary conditions)
- Concurrent access tests present in guild_housing, guild_vehicle
- **Critical Gap**: No determinism validation tests for time-dependent packages
- **Critical Gap**: Tests don't fail on time.Now() usage, allowing violations to persist

## Integration Status
**Cross-Domain Integration** — Connects 5 major domains (engine, world, network, procgen, narrative)

### Engine Integration (`pkg/engine/`)
- `choice_consequences` ✅ — Uses ComponentHaver interface for choice tracking
- `narrative_world` ✅ — Registered as System in engine, uses CompanionComponent
- `political_warfare` ⚠️ — Registered as System but has time.Now() violations
- All integration systems follow ECS architecture with separate Manager (data) and System (logic)

### World Integration (`pkg/world/`)
- `companion_housing` ✅ — Integrates housing blueprints, deterministic via TimeProvider
- `guild_housing` ❌ — Integrates housing + guild federation but non-deterministic IDs/timestamps
- `housing_crafting` ✅ — Integrates housing furniture, fully deterministic seed-based

### Network Integration (`pkg/network/`)
- `guild_housing` ❌ — Integrates federation/guild for shared spaces, non-deterministic
- `guild_vehicle` ❌ — Integrates federation/guild for fleet management, non-deterministic timestamps
- `political_warfare` ❌ — Integrates guild federation for wars/treaties, extensive time.Now() usage

### Procgen Integration (`pkg/procgen/`)
- `housing_crafting` ✅ — Uses recipe generator, deterministic
- `trade_routes` ✅ — Uses vehicle generator with deterministic TimeProvider for timing
- `narrative_world` ✅ — Uses quest generator, deterministic with TimeProvider

### System Registration Patterns
**Inconsistent** — Three different patterns found:
1. **ECS System Pattern** (narrative_world, political_warfare, world_events) — Manager + System wrapper registered in client/server
2. **Manager-Only Pattern** (guild_housing, trade_routes, guild_vehicle) — No ECS System wrapper, used as utility in client handlers
3. **Pure Integration Pattern** (choice_consequences, companion_housing, housing_crafting) — No registration, called directly from systems

**No Centralized Registry** — All registration happens manually in `cmd/client/handlers.go` and `cmd/server/main.go` without discovery mechanism or documented pattern.

## Recommendations
1. **CRITICAL**: Replace remaining 45 `time.Now()` calls with injectable TimeProvider pattern across guild_housing (15 calls), political_warfare (32 calls), guild_vehicle (12 calls). Use existing TimeProvider implementations from choice_consequences/narrative_world/world_events/trade_routes as reference. This is required for multiplayer determinism and save/load stability. (world_events and trade_routes already fixed.)

2. **CRITICAL**: Add determinism validation tests for all time-dependent packages. Create test cases that:
   - Generate IDs/timestamps with fixed TimeProvider
   - Verify same inputs produce same outputs
   - Detect time.Now() usage via test failures
   - Validate multiplayer synchronization scenarios

3. **HIGH**: Document system registration pattern in `pkg/integration/INTEGRATION.md` covering:
   - When to use ECS System wrapper vs Manager-only
   - How to register in client/server initialization
   - TimeProvider injection requirements for determinism
   - Testing requirements for integration systems

4. **MEDIUM**: Standardize on System wrapper pattern for all integration systems with manager state (guild_housing, trade_routes, guild_vehicle should follow narrative_world/political_warfare pattern).

5. **LOW**: Update root `doc.go` to clarify package contains both production integration *systems* and integration *tests*, not just tests. Current description is misleading for 8/10 sub-packages that are production code.

6. **LOW**: Consider xvfb-run wrapper script (`scripts/test-integration.sh`) to simplify CI testing of packages with transitive Ebiten dependencies.

## Deterministic Generation Compliance ❌
**Non-Compliant** — 45 `time.Now()` violations across 3 packages create multiplayer desync risk.

### Violations by Package:
- `guild_housing/guild_housing_manager.go`: **15 violations** (lines 78,88,89,150,203,246,253,350,355,360,443,448,527,568,573)
- `political_warfare/manager.go`: **32 violations** (partial list: 105,148,153,158,212,252,284,338,371,420,455,557 + 20 more)
- `guild_vehicle/fleet_manager.go`: **12 violations** (lines 48,49,80,81,99,100,104,125,143,164,202,219)

### Recently Fixed Packages:
- ✅ `trade_routes/manager.go`: **0 violations** — Fixed 2026-02-16 (6 time.Now() → TimeProvider)
- ✅ `world_events/manager.go`: **0 violations** — Fixed 2026-02-16 (6 time.Now() → TimeProvider)

### Compliant Packages (Examples to Follow):
- ✅ `choice_consequences` — Uses injectable TimeProvider for all time operations
- ✅ `narrative_world` — Uses injectable TimeProvider for all time operations
- ✅ `companion_housing` — Uses injectable TimeProvider for all time operations
- ✅ `housing_crafting` — No time dependency, purely seed-based generation
- ✅ `world_events` — Uses injectable TimeProvider for all time operations (fixed 2026-02-16)
- ✅ `trade_routes` — Uses injectable TimeProvider for all time operations (fixed 2026-02-16)

### Required Fix Pattern:
```go
// ❌ BEFORE (non-deterministic)
houseID := fmt.Sprintf("ghouse-%s-%d", guildID, time.Now().UnixNano())
house.CreatedAt = time.Now()

// ✅ AFTER (deterministic)
type Manager struct {
    timeProvider TimeProvider // Injectable interface
}
houseID := fmt.Sprintf("ghouse-%s-%d", guildID, m.timeProvider.Now().UnixNano())
house.CreatedAt = m.timeProvider.Now()
```

### Impact Assessment:
- **Multiplayer Desync Risk**: Different clients will generate different IDs/timestamps for guild houses, trade routes, wars, and events, causing synchronization failures
- **Save/Load Instability**: Timestamps change on every load, breaking save file determinism
- **Testing Fragility**: Time-dependent tests cannot be reliably reproduced
- **Federation Protocol Violation**: Cross-server operations will fail due to ID mismatches

## Network Interface Compliance ✅
**Not Applicable** — Package does not define network transports directly.

Integration systems use network interfaces from `pkg/network/federation/guild` which already follows interface-based design (net.Addr, net.Conn). No violations in integration layer.

## ECS Compliance ✅
**Compliant** — Proper separation of components and systems across all integration packages.

### Component Examples (Pure Data):
- `choice_consequences/component.go:ConsequenceComponent` — Pure data with only `Type()` method ✅
- `narrative_world/component.go:NarrativeWorldStateComponent` — Pure data struct ✅
- All integration components follow ECS purity requirement

### System Examples (Logic):
- `narrative_world/system.go:NarrativeWorldSystem.Update()` — All logic in systems ✅
- `political_warfare/system.go:PoliticalWarfareSystem.Update()` — Proper ECS pattern ✅
- `world_events/system.go:WorldEventsSystem.Update()` — Follows architecture ✅

### Manager Pattern (Data Management):
- Guild housing Manager handles data, no behavior on components ✅
- Trade routes Manager owns business logic separate from data ✅
- All packages maintain strict component/system separation ✅

## Error Handling ✅
**Good** — Comprehensive error returns with structured logging.

### Strengths:
- ✅ All exported functions return `error` types where failure is possible
- ✅ Error wrapping with context: `fmt.Errorf("failed to X: %w", err)` pattern used
- ✅ Structured logging present in all packages (logrus.WithFields)
- ✅ No swallowed errors detected (grep verification passed)
- ✅ Nil checks before dereferencing pointers
- ✅ Validation errors returned with descriptive messages

### Logging Quality:
- All packages use `logrus.WithFields` for contextual logging
- Field names consistent: `guildID`, `playerID`, `houseID`, `routeID`, `warID`, `eventType`
- Log levels appropriate: Debug for trace, Info for operations, Warn for edge cases, Error for failures
- No println/log.Print usage detected

### Example Error Handling:
```go
// guild_housing/guild_housing_manager.go:76-95
func (m *Manager) CreateGuildHouse(...) error {
    if guildID == "" || ownerID == "" {
        return fmt.Errorf("guild and owner IDs required")
    }
    // Detailed validation and error context
}
```

## Documentation Coverage ⚠️
**Good** — Most packages have comprehensive documentation, minor gaps exist.

### Package-Level Documentation:
- ✅ All 10 sub-packages have `doc.go` files with detailed usage examples
- ✅ File headers explain code organization and relocation history
- ✅ Overview sections describe integration purpose and features
- ⚠️ Root `doc.go` misleading (describes tests but contains production systems)

### API Documentation:
- ✅ All exported types have godoc comments
- ✅ All exported functions have godoc comments
- ✅ Complex algorithms explained (e.g., political warfare balance calculations)
- ✅ Integration points documented (which systems connect to what)

### Example Documentation Quality:
- `guild_housing/doc.go`: 60+ lines covering features, permissions, usage examples
- `trade_routes/doc.go`: Detailed route generation, profit calculations, mission system
- `political_warfare/doc.go`: War mechanics, treaty system, embargo effects

### Missing Documentation:
- No `pkg/integration/INTEGRATION.md` guide for system registration patterns
- No migration guide for adopting TimeProvider in existing systems
- No architecture diagram showing cross-domain integration flow

## Code Quality ✅
**Excellent** — Clean architecture, well-organized, follows Go idioms.

### Architecture Strengths:
- **Separation of Concerns**: Manager (data) / System (update logic) / Component (state) cleanly separated
- **Consistent Patterns**: All 10 packages follow similar structure (doc.go, types.go, manager.go, system.go, component.go)
- **Dependency Injection**: TimeProvider abstraction enables testability in 5/10 packages
- **Thread Safety**: Mutex usage present where needed (guild_housing concurrent access, political_warfare race protection)

### File Organization (Typical Package):
```
pkg/integration/<feature>/
├── doc.go                 - Package documentation with examples
├── types.go               - Type definitions and constants
├── manager.go             - Business logic and data management
├── system.go              - ECS System wrapper (Update loop)
├── component.go           - ECS Component definition (pure data)
├── <feature>_test.go      - Comprehensive table-driven tests
```

### Code Style:
- ✅ Consistent naming (camelCase private, PascalCase public)
- ✅ Table-driven tests throughout
- ✅ Short variable names in tight loops, descriptive names for complex operations
- ✅ Comments explain "why" not "what"
- ✅ No magic numbers — constants defined for all thresholds
- ✅ Error messages descriptive and actionable

### Performance Considerations:
- Efficient data structures (maps for O(1) lookups)
- Mutex usage minimized to critical sections
- No unnecessary allocations in hot paths
- Lazy initialization where appropriate

## Sub-Package Status Summary

| Package | Coverage | Determinism | Issues | Status |
|---------|----------|-------------|--------|--------|
| choice_consequences | 87.1% | ✅ TimeProvider | 0 | Complete |
| companion_housing | 93.5% | ✅ TimeProvider | 0 | Complete |
| housing_crafting | 98.5% | ✅ Seed-based | 0 | Complete |
| narrative_world | 90.7% | ✅ TimeProvider | 0 | Complete |
| world_events | 91.2% | ✅ TimeProvider | 0 | Complete |
| guild_housing | 91.8% | ❌ time.Now() | 15 | Needs Work |
| trade_routes | 92.4% | ✅ TimeProvider | 0 | Complete |
| political_warfare | 94.7% | ❌ time.Now() | 32 | Needs Work |
| guild_vehicle | 94.6% | ❌ time.Now() | 12 | Needs Work |

**Overall**: 6/9 packages production-ready, 3/9 require TimeProvider migration for multiplayer determinism.
