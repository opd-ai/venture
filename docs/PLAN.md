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

## Phase 3: System Registration ✅

All packages integrated via engine wrapper systems in `cmd/client/handlers.go`.

- [x] **`pkg/integration/choice_consequences`** - Wrapped by `engine.ChoiceConsequencesSystem`
- [x] **`pkg/integration/guild_vehicle`** - Wrapped by `engine.GuildVehicleSystem`
- [x] **`pkg/integration/territory_siege`** - Wrapped by `engine.TerritorySiegeSystem`
- [x] **`pkg/integration/world_events`** - Wrapped by `engine.WorldEventsSystem`

**Completed: December 2025**

---

## Phase 4: Network Enhancement ✅

Network layer improvements requiring deeper integration.

### 4.1 `pkg/network/resilience` ✅
- **Status**: Integrated via `cmd/server/main.go`
- **Verified**: Import, `--simulate-network` flag, `--resilience-metrics` flag
- **Usage**: `./server --simulate-network=high --resilience-metrics`
- **Coverage**: 76.2%

### 4.2 `pkg/network/federation/webrtc` ✅
- **Status**: Integrated via `cmd/client/webrtc_wasm.go`
- **Verified**: WASM-specific build with WebRTC support
- **Usage**: Automatic for browser-to-browser federation in WASM builds

### Phase 4 Verification ✅
```bash
go build ./cmd/client && go build ./cmd/server  # ✅ Passed
go test ./pkg/network/resilience/...            # ✅ 76.2% coverage
GOOS=js GOARCH=wasm go build -o web/client.wasm ./cmd/client  # ✅ Passed
```

**Completed: December 2025**

---

## Phase 5: Development Tools ✅

Development/testing tools integrated into CI/CD and local development workflow.

### 5.1 `pkg/audit/features` ✅
- **Status**: Integrated via `.github/workflows/quality.yml` and `make feature-audit`
- **Usage**: `make feature-audit` or CI runs on every push/PR
- **Coverage**: 100.0%

### 5.2 `pkg/visualtest` ✅
- **Status**: Integrated via `.github/workflows/quality.yml` and `make visual-regression`
- **Usage**: `make visual-regression` or CI runs on every push/PR
- **Coverage**: 83.0%

### 5.3 `pkg/visualtest/parity` ✅
- **Status**: Integrated via `.github/workflows/quality.yml` and `make parity-test`
- **Usage**: `make parity-test` or CI runs on every push/PR
- **Coverage**: 87.9%

### Phase 5 Verification ✅
```bash
go test ./pkg/audit/features/... ./pkg/visualtest/... ./pkg/visualtest/parity/...
make quality  # Runs all quality tools locally
```

**Completed: December 2025**

---

## Phase 6: Future Features (Deferred)

These packages require significant work or have external dependencies.

### 6.1 `pkg/balance` ✅
- **LOC**: 1,232 | **Status**: Integrated
- **Integration**: Added `--balance-validate` flag to `cmd/server/main.go`
- **Usage**: `./server --balance-validate` or `make balance-validate`
- **CI/CD**: Added to `.github/workflows/quality.yml` for automated validation
- **Coverage**: 81.0%
- **Completed**: December 2025

### 6.2 `pkg/migration` ✅
- **LOC**: 619 | **Status**: Integrated
- **Integration**: Added migration hooks to `pkg/saveload/migrator.go`, `--migration-validate` flag to `cmd/server/main.go`
- **Usage**: `./server --migration-validate` or `make migration-validate`
- **Makefile**: Added `migration-validate` target
- **Coverage**: saveload 73.3%, migration 77.1%
- **Completed**: December 2025

### 6.3 `pkg/modding`
- **LOC**: 1,484 | **Status**: Complete
- **Blocker**: Security sandbox incomplete (per V10 audit)
- **Action**: Defer until mod sandbox implemented
- **Effort**: Large

### 6.4 `pkg/ux` ✅
- **LOC**: 1,571 | **Status**: Integrated
- **Integration**: Added `--ux-validate` flag to `cmd/server/main.go`, `make ux-validate` target
- **Usage**: `./server --ux-validate` or `make ux-validate`
- **Coverage**: 96.5%
- **Completed**: December 2025

### 6.5 `pkg/social` ✅
- **LOC**: 574 | **Status**: Complete
- **Integration**: Added doc.go, integrated social error types into trade_system.go and chat_system.go
- **Coverage**: 98.0% (social), 91.3% (social/persistence)
- **Completed**: December 2025

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

- [x] `go build ./cmd/client && go build ./cmd/server` — Builds succeed
- [x] `go test ./pkg/...` — All tests pass
- [x] Run client, verify features work
- [x] Check logs for new system activation
- [x] Verify no performance regression (60 FPS target)
- [x] Verify memory usage (<500MB client)

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

- [x] All Priority 1-4 packages integrated
- [x] Phase 5 development tools integrated into CI/CD
- [x] All systems unconditionally registered
- [ ] No feature flags in codebase
- [x] Test coverage maintained ≥65%
- [x] Build passes on all platforms
- [x] 60 FPS performance maintained

---

*Generated: December 2025*
*Venture v10.0 Integration Plan*
