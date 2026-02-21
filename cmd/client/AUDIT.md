# Audit Report: cmd/client

**Date**: 2026-02-16
**Auditor**: Automated
**Coverage**: 32.6% (limited by Ebiten display server dependency; core logic in pkg/ averages 82.4%)

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| High     | 1     | 1     | 0         |
| Medium   | 4     | 4     | 0         |
| Low      | 2     | 2     | 0         |

## Issues

### HIGH — Fixed

#### H1: Nil pointer dereference in lazy initialization (handlers.go)
- **Location**: `initializeEnvironmentalSystems()` and `registerNonCriticalSystems()`
- **Description**: Lazy-init functions accessed `sys.progressionSystem`, `sys.combatSystem`, `sys.itemPickupSystem`, and `sys.statusEffectSystem` without nil guards. These systems are initialized in `initializeProgressionSystems`/`initializeCombatSystems` before `scheduleLazyInit` in normal flow, but tests calling `initializeCoreSystems` + `scheduleLazyInit` directly would crash.
- **Fix**: Added nil guards around 7 callback registrations:
  - `sys.combatSystem.SetCriticalHitCallback` → guarded
  - `sys.progressionSystem.AddLevelUpCallback` → guarded
  - `sys.itemPickupSystem.SetPickupCallback` → guarded
  - `sys.combatSystem.SetDamageResistedCallback` → guarded
  - `sys.combatSystem.SetShieldAbsorbCallback` → guarded
  - `sys.combatSystem.SetDamageCallback` → guarded
  - `sys.statusEffectSystem.SetTickCallback` → guarded
  - `sys.combatSystem.AddDamageCallback` → guarded (registerNonCriticalSystems)
  - `sys.combatSystem.AddCriticalHitCallback` → guarded (registerNonCriticalSystems)
- **Impact**: Tests now pass (previously panic with SIGSEGV)

### MEDIUM — Fixed

#### M1: Unchecked type assertions on Generator results (util.go)
- **Location**: Lines 928, 962, 993, 1270
- **Description**: `itemGen.Generate()` returns `interface{}`, and results were asserted to `[]*item.Item` or `*item.Item` without the `ok` check pattern.
- **Fix**: Added safe type assertion pattern (`value, ok := result.(Type); if !ok { return }`) at all 4 locations.

#### M2: Unsafe component type assertions (util.go, handlers.go)
- **Location**: Multiple locations after `GetComponent()` calls
- **Description**: Pattern `comp.(*engine.SomeComponent)` used without `ok` check after `GetComponent()` returns true.
- **Fix**: Added safe assertion pattern with `ok` check at all component type assertion sites in both util.go (~20 locations) and handlers.go (4 locations). Returns early or continues loop on assertion failure.

#### M3: Mutex unlock without defer in lazy init (handlers.go:747-749)
- **Location**: `scheduleLazyInit()` completion handler
- **Description**: `sys.lazyInitMutex.Lock()` / `sys.lazyInitCompleted = true` / `sys.lazyInitMutex.Unlock()` pattern without `defer`.
- **Fix**: Changed to `defer sys.lazyInitMutex.Unlock()` for consistency with `isLazyInitCompleted()`.

#### M4: WeatherXPBonusSystem receives nil progressionSystem (handlers.go:1186)
- **Location**: `initializeEnvironmentalSystems()` line 1186
- **Description**: `sys.weatherXPBonusSystem.SetProgressionSystem(sys.progressionSystem)` stores a nil reference when progressionSystem hasn't been initialized.
- **Fix**: Added nil guard: `if sys.progressionSystem != nil { ... }`.

### LOW — Fixed

#### L1: doc.go references Go 1.24 and Ebiten 2.9 (doc.go)
- **Description**: Version references may become stale. Consider referencing go.mod versions.
- **Status**: ✅ **FIXED 2026-02-21** — Updated to Go 1.24.5+ and Ebiten 2.9.3

#### L2: Large function count in handlers.go
- **Description**: handlers.go was ~4800 lines with 80+ functions. Consider splitting into domain-specific files.
- **Status**: ✅ **FIXED 2026-02-21** — Extracted performance monitoring functions (8 functions, ~100 lines) to `init_monitoring.go`. Extracted spawn/entity generation functions (16 functions, ~260 lines) to `init_spawning.go`. handlers.go reduced from 82 to 66 functions (3617 lines).

## Architecture Notes

- Client follows layered architecture: Input → Systems → Rendering → Audio → Network
- Lazy initialization pattern defers ~80% of systems until after first frame
- `systemsContainer` acts as dependency injection container for 200+ systems
- System wrappers in `system_wrappers.go` adapt diverse Update signatures to ECS interface
- Mobile stub (`main_mobile.go`) correctly prevents compilation on Android/iOS
- WebRTC WASM support (`webrtc_wasm.go`) properly build-tagged for browser federation

## Test Coverage

- 32.6% statement coverage (32 tests passing)
- Coverage limited by Ebiten display server requirement (requires Xvfb in CI)
- Core game logic tested in pkg/ packages (82.4% average)
- Tests cover: lazy initialization, performance monitoring, sprite warming, high-latency config, parallel init, integration, snapshot conversion, system init, minigame systems
