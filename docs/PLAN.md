# Integration Plan - December 2025

## Principle

**All packages become unconditionally active.** No feature flags. No optional toggles. If a package exists, it should be integrated as always-on baseline functionality.

---

## Executive Summary

| Phase | Focus | Packages | Effort |
|-------|-------|----------|--------|
| Phase 1 | Indirectly Active Verification | 3 | Verify only |
| Phase 2 | Simple Imports | 4 | Small |
| Phase 3 | System Registration | 4 | Small |
| Phase 4 | Network Enhancement | 2 | Small-Medium |
| Phase 5 | Development Tools | 3 | Optional |
| Phase 6 | Future Features | 6 | Medium-Large |

**Total Dormant Packages**: 32  
**Already Active (indirect)**: 5  
**Requiring Integration**: 18  
**Test/Dev Tools (optional)**: 4  
**Future/Deferred**: 5

---

## Phase 1: Verify Indirectly Active Packages ✅

These packages are already integrated via `pkg/engine` imports. Verified working correctly.

### Timeline: Immediate (verification only)

- [x] **`pkg/procgen/dialog`**
  - Status: Imported by `pkg/engine/book_spawning.go`, `pkg/engine/npcdialog_system.go`
  - Verified: `go test ./pkg/procgen/dialog/...` passes

- [x] **`pkg/procgen/entity`**
  - Status: Imported by `pkg/engine/entity_spawning.go`, `pkg/engine/merchant_spawn.go`
  - Verified: `go test ./pkg/procgen/entity/...` passes

- [x] **`pkg/procgen/legendary`**
  - Status: Imported by `pkg/engine/legendary_quest_system.go`
  - Verified: `go test ./pkg/procgen/legendary/...` passes

- [x] **`pkg/engine/performance`**
  - Status: Imported by `pkg/engine/performance_monitoring_system.go`
  - Verified: `go test ./pkg/engine/performance/...` passes

- [x] **`pkg/audio/synthesis`**
  - Status: Used transitively via `pkg/audio/music`
  - Verified: `go test ./pkg/audio/synthesis/...` passes

**Completed: December 2025**

---

## Phase 2: Simple Import Integration ✅

Packages integrated with unconditional activation.

### 2.1 `pkg/rendering/tiles` ✅
- **Status**: Already integrated via `pkg/engine/terrain_render_system.go` and `pkg/engine/tile_cache.go`
- **Verified**: No additional import needed - transitively active

### 2.2 `pkg/world/economy` ✅
- **Status**: Already integrated via `pkg/engine/economy_system.go`
- **Verified**: No additional import needed - transitively active

### 2.3 `pkg/stability` ✅
- **Integrated**: Added import and `--stability-monitor` flag in `cmd/server/main.go`
- **Usage**: `./server --stability-monitor` to enable production validation
- **Verified**: `go test ./pkg/stability/...` passes

### 2.4 `pkg/security` ✅
- **Integrated**: Added import and `--security-audit` flag in `cmd/server/main.go`
- **Usage**: `./server --security-audit` to run security validation at startup
- **Verified**: `go test ./pkg/security/...` passes

**Completed: December 2025**

---

## Phase 3: System Registration (Week 1)

Packages with systems that need `AddSystem()` registration.

### 3.1 `pkg/integration/choice_consequences`
- **LOC**: 1,545 | **Tests**: 1 | **Completeness**: Complete
- **Blocker**: System not registered
- **Dependencies**: None (standalone)
- **Integration**:
  1. File: `cmd/client/handlers.go`
  2. Add import:
     ```go
     import "github.com/opd-ai/venture/pkg/integration/choice_consequences"
     ```
  3. Create and register system in `initializeV10Systems()`:
     ```go
     choiceSystem := choice_consequences.NewSystem()
     game.World.AddSystem(choiceSystem)
     ```
- **Effort**: Small

### 3.2 `pkg/integration/guild_vehicle`
- **LOC**: 1,433 | **Tests**: 2 | **Completeness**: Complete
- **Blocker**: System not registered
- **Dependencies**: `pkg/network/federation/guild`, `pkg/engine/physics/vehicle`
- **Integration**:
  1. File: `cmd/client/handlers.go`
  2. Add import:
     ```go
     import "github.com/opd-ai/venture/pkg/integration/guild_vehicle"
     ```
  3. Create and register system:
     ```go
     guildVehicleSys := guild_vehicle.NewSystem(federationGuildClient, vehicleSystem)
     game.World.AddSystem(guildVehicleSys)
     ```
- **Effort**: Small

### 3.3 `pkg/integration/territory_siege`
- **LOC**: 1,925 | **Tests**: 4 | **Completeness**: Complete
- **Blocker**: System not registered
- **Dependencies**: `pkg/engine`, `pkg/network/federation/guild`, `pkg/world`
- **Integration**:
  1. File: `cmd/client/handlers.go`
  2. Add import:
     ```go
     import "github.com/opd-ai/venture/pkg/integration/territory_siege"
     ```
  3. Register system:
     ```go
     siegeSys := territory_siege.NewSystem(world, guildClient)
     game.World.AddSystem(siegeSys)
     ```
- **Effort**: Small

### 3.4 `pkg/integration/world_events`
- **LOC**: 1,876 | **Tests**: 2 | **Completeness**: Complete
- **Blocker**: System not registered
- **Dependencies**: `pkg/procgen`
- **Integration**:
  1. File: `cmd/client/handlers.go`
  2. Add import:
     ```go
     import "github.com/opd-ai/venture/pkg/integration/world_events"
     ```
  3. Register system:
     ```go
     worldEventsSys := world_events.NewSystem(seed)
     game.World.AddSystem(worldEventsSys)
     ```
- **Effort**: Small

### Phase 3 Verification
```bash
go build ./cmd/client
go test ./pkg/integration/choice_consequences/... ./pkg/integration/guild_vehicle/... ./pkg/integration/territory_siege/... ./pkg/integration/world_events/...
# Run client and verify systems activate
./client -verbose 2>&1 | grep -i "world_events\|siege\|guild_vehicle\|choice"
```

---

## Phase 4: Network Enhancement (Week 2)

Network layer improvements requiring deeper integration.

### 4.1 `pkg/network/resilience`
- **LOC**: 1,334 | **Tests**: 1 | **Completeness**: Complete
- **Blocker**: Not imported
- **Integration**:
  1. File: `cmd/client/handlers.go` and `cmd/server/main.go`
  2. Add import:
     ```go
     import "github.com/opd-ai/venture/pkg/network/resilience"
     ```
  3. Wrap network connections with resilience layer:
     ```go
     resilientConn := resilience.WrapConnection(conn, resilience.DefaultConfig())
     ```
  4. Configure retry policies for high-latency scenarios
- **Effort**: Medium

### 4.2 `pkg/network/federation/webrtc`
- **LOC**: 4,246 | **Tests**: 5 | **Completeness**: Complete
- **Blocker**: Browser/WASM only
- **Integration**:
  1. File: `cmd/client/main_wasm.go` (create if needed)
  2. Add WASM-specific import:
     ```go
     //go:build js && wasm
     import "github.com/opd-ai/venture/pkg/network/federation/webrtc"
     ```
  3. Initialize WebRTC for browser-to-browser connections
  4. Fall back to standard federation when WebRTC unavailable
- **Effort**: Large (WASM-specific)

### Phase 4 Verification
```bash
go build ./cmd/client && go build ./cmd/server
go test ./pkg/network/resilience/...
# WASM build
GOOS=js GOARCH=wasm go build -o web/client.wasm ./cmd/client
```

---

## Phase 5: Development Tools (Optional)

These packages are development/testing tools, not gameplay features.

### 5.1 `pkg/audit/features`
- **LOC**: 2,048 | **Tests**: 1 | **Purpose**: Feature audit framework
- **Integration**: Use in CI/CD pipeline for feature validation
- **Command**: `go run ./examples/featureaudit`
- **Effort**: Optional

### 5.2 `pkg/visualtest`
- **LOC**: 5,176 | **Tests**: 7 | **Purpose**: Visual regression testing
- **Integration**: Add to CI/CD visual testing stage
- **Command**: `go run ./examples/visualregressiontest`
- **Effort**: Optional

### 5.3 `pkg/visualtest/parity`
- **LOC**: 1,340 | **Tests**: 3 | **Purpose**: Cross-platform parity
- **Integration**: Add to CI/CD cross-platform validation
- **Command**: `go run ./examples/paritytest`
- **Effort**: Optional

### Phase 5 Verification
```bash
go test ./pkg/audit/features/... ./pkg/visualtest/... ./pkg/visualtest/parity/...
```

---

## Phase 6: Future Features (Deferred)

These packages require significant work or have external dependencies.

### 6.1 `pkg/balance`
- **LOC**: 1,232 | **Status**: Complete but unused
- **Blocker**: No system consumers
- **Action**: Design integration points for combat/economy balancing
- **Effort**: Medium

### 6.2 `pkg/migration`
- **LOC**: 619 | **Status**: Complete
- **Blocker**: No trigger mechanism
- **Action**: Add migration hooks to saveload system
- **Effort**: Medium

### 6.3 `pkg/modding`
- **LOC**: 1,484 | **Status**: Complete
- **Blocker**: Security sandbox incomplete (per V10 audit)
- **Action**: Defer until mod sandbox implemented
- **Effort**: Large

### 6.4 `pkg/ux`
- **LOC**: 1,571 | **Status**: Complete
- **Blocker**: No UI integration
- **Action**: Connect to rendering/ui system
- **Effort**: Medium

### 6.5 `pkg/social`
- **LOC**: 574 | **Status**: Partial (missing doc.go)
- **Blocker**: Incomplete package
- **Action**: Complete package, add documentation
- **Effort**: Medium

### 6.6 Container Packages (No Action)
These are empty package roots:
- `pkg/audit` - Empty
- `pkg/class` - Empty (see `pkg/class/advanced`)
- `pkg/companion` - Empty (see `pkg/companion/learning`)
- `pkg/engine/saves` - Empty
- `pkg/narrative` - Empty (see `pkg/narrative/branching`)

---

## Integration Template

When integrating a new component, follow this template:

### Component Integration
```go
// In createPlayerEntity() or equivalent:

// === PHASE XX: ComponentName ===
// Add component for feature description
player.AddComponent(engine.NewFooComponent(requiredParams))
if verbose {
    clientLogger.WithFields(logrus.Fields{
        "component": "foo",
        "param1":    requiredParams.Field1,
    }).Debug("foo component added")
}
```

### System Integration
```go
// In initializeVXXSystems():

// Phase XX.X: Feature description
sys.fooSystem = foo.NewSystem(dependencies...)
game.World.AddSystem(sys.fooSystem)
if verbose {
    clientLogger.WithFields(logrus.Fields{
        "system": "foo",
    }).Debug("foo system registered")
}
```

### Wrapper Pattern (if needed)
```go
// In cmd/client/util.go:

type fooSystemWrapper struct {
    system *foo.System
}

func (w *fooSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
    w.system.Update(deltaTime)
}

// Usage:
game.World.AddSystem(&fooSystemWrapper{system: sys.fooSystem})
```

---

## Verification Checklist

After each phase:

- [ ] `go build ./cmd/client && go build ./cmd/server` — Builds succeed
- [ ] `go test ./pkg/...` — All tests pass
- [ ] Run client, verify features work
- [ ] Check logs for new system activation
- [ ] Verify no performance regression (60 FPS target)
- [ ] Verify memory usage (<500MB client)

### Full Verification Script
```bash
#!/bin/bash
set -e

echo "=== Building ==="
go build ./cmd/client
go build ./cmd/server

echo "=== Testing ==="
go test -cover ./pkg/... | tee coverage.txt

echo "=== Coverage Summary ==="
grep "coverage:" coverage.txt | awk '{sum+=$2; count++} END {print "Average:", sum/count "%"}'

echo "=== Build Successful ==="
```

---

## Anti-Patterns to Avoid

| ❌ Don't | ✅ Do |
|----------|-------|
| `if enableFeatureX { ... }` | Unconditional execution |
| `-enable-*` CLI flags | Hardcode as always-on |
| Lazy component initialization | Add components during entity creation |
| `comp, _ := GetComponent()` | `comp, ok := GetComponent(); if ok && comp != nil` |
| Adding to only one client | Update both desktop and mobile |

---

## Success Criteria

- [ ] All Priority 1-4 packages integrated
- [ ] All systems unconditionally registered
- [ ] No feature flags in codebase
- [ ] Test coverage maintained ≥65%
- [ ] Build passes on all platforms
- [ ] 60 FPS performance maintained

---

*Generated: December 2025*
*Venture v10.0 Integration Plan*
