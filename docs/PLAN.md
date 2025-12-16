# Integration Plan - December 2025

## Principle

**All packages become unconditionally active. No feature flags. No optional toggles.**

If a package exists in `pkg/`, it should be integrated and always-on. This plan provides the exact changes needed to activate all dormant packages.

---

## Critical Integration Rules

### Component Initialization: NEVER Use Lazy Initialization

**Always add components during entity creation, NEVER during system Update().**

```go
// ❌ BAD: Lazy initialization in System.Update()
func (s *MySoundSystem) Update(deltaTime float64) {
    entities := s.world.GetEntitiesWith("my_sound")
    if len(entities) == 0 {
        player := s.world.GetEntitiesWith("position", "health")[0]
        player.AddComponent(NewMySoundComponent()) // DANGER: Cache not invalidated!
    }
}

// ✅ GOOD: Add component during entity creation
func createPlayerEntity(game *engine.EbitenGame, ...) *engine.Entity {
    player := game.World.CreateEntity()
    player.AddComponent(&engine.PositionComponent{X: x, Y: y})
    player.AddComponent(engine.NewMySoundComponent()) // Safe: before first Update()
    return player
}
```

### Defensive Component Access

```go
// ✅ GOOD: Check both ok AND nil
comp, ok := entity.GetComponent("my_component")
if ok && comp != nil {
    myComp := comp.(*MyComponent)
    // Safe to use
}
```

---

## Phase 1: Foundation (Week 1) ✅ COMPLETE

Activate core systems that others depend on.

**Status:** All 5 subsections complete. Systems integrated and verified.

### 1.1 Audio Synthesis Engine ✅

**Package**: `pkg/audio/synthesis`  
**Effort**: Small  
**Demo**: `examples/audiotest/`

Created `synthesis.Engine` type providing unified synthesis API (oscillators + envelopes).
Added `Synthesizer` interface and `SetSynthesizer()` method to `AudioManagerSystem`.
Coverage: 96.4%

### 1.2 World Economy System ✅

**Package**: `pkg/world/economy`  
**Effort**: Small  
**Demo**: `examples/marketplacetest/`

Created `economy.System` ECS wrapper for marketplace functionality.
Integrated into both client (via `engine.EconomySystem`) and server.
Coverage: 87.3%

### 1.3 Choice Consequences System ✅

**Package**: `pkg/integration/choice_consequences`  
**Effort**: Small  
**Demo**: `examples/choicetest/`

Already integrated via `engine.ChoiceConsequencesSystem` wrapper.

### 1.4 Guild Vehicle Fleet System ✅

**Package**: `pkg/integration/guild_vehicle`  
**Effort**: Small  
**Demo**: `examples/fleettest/`

Already integrated via `engine.GuildVehicleSystem` wrapper.

### 1.5 World Events System ✅

**Package**: `pkg/integration/world_events`  
**Effort**: Small  
**Demo**: `examples/worldeventstest/`

Already integrated via `engine.WorldEventsSystem` wrapper.

---

## Phase 2: Content Generation (Week 2)

Activate procedural generation enhancements.

### 2.1 Entity Generator Integration

**Package**: `pkg/procgen/entity`  
**Effort**: Medium  
**Demo**: `examples/entitytest/`

**Files to modify**:

```go
// cmd/client/handlers.go - Add import
import "github.com/opd-ai/venture/pkg/procgen/entity"

// cmd/client/handlers.go - In entity spawning initialization
entityGenerator := entity.NewGenerator(seed)

// pkg/engine/entity_spawning.go - Replace manual enemy creation
func (s *EntitySpawningSystem) SpawnEnemy(seed int64, params procgen.GenerationParams) *Entity {
    entityGen := entity.NewGenerator(seed)
    result, err := entityGen.Generate(params)
    if err != nil {
        log.WithError(err).Error("entity generation failed")
        return nil
    }
    return s.createFromTemplate(result.(*entity.EntityTemplate))
}
```

**Verification**:
```bash
go run ./examples/entitytest/
go test ./pkg/procgen/entity/...
```

### 2.2 Dialog Generator Integration

**Package**: `pkg/procgen/dialog`  
**Effort**: Medium  
**Demo**: `examples/dialogtest/`

**Files to modify**:

```go
// cmd/client/handlers.go - Add import
import "github.com/opd-ai/venture/pkg/procgen/dialog"

// cmd/client/handlers.go - In dialog system initialization
dialogGenerator := dialog.NewGenerator(seed)
sys.npcDialogSystem.SetGenerator(dialogGenerator)
```

**Verification**:
```bash
go run ./examples/dialogtest/
go run ./examples/dialog_demo/
```

### 2.3 Legendary Item Generator

**Package**: `pkg/procgen/legendary`  
**Effort**: Small  
**Demo**: `examples/legendarytest/`

**Files to modify**:

```go
// cmd/client/handlers.go - Add import
import "github.com/opd-ai/venture/pkg/procgen/legendary"

// cmd/client/handlers.go - In item generation initialization
legendaryGenerator := legendary.NewGenerator(seed)
sys.legendaryQuestSystem.SetGenerator(legendaryGenerator)
```

**Verification**:
```bash
go run ./examples/legendarytest/
```

---

## Phase 3: Testing Infrastructure (Week 3)

Enable audit and testing tools in CI/CD.

### 3.1 Feature Audit Framework

**Package**: `pkg/audit/features`  
**Effort**: Small  
**Demo**: `examples/featureaudit/`

**Files to modify**:

```yaml
# .github/workflows/ci.yml - Add audit step
- name: Run Feature Audit
  run: go run ./examples/featureaudit/
```

**Verification**:
```bash
go run ./examples/featureaudit/
```

### 3.2 Procgen Audit Framework

**Package**: `pkg/procgen/audit`  
**Effort**: Small  
**Demo**: `examples/auditrunner/`

**Files to modify**:

```yaml
# .github/workflows/ci.yml - Add determinism check
- name: Verify Generator Determinism
  run: go run ./examples/auditrunner/
```

**Verification**:
```bash
go run ./examples/auditrunner/
```

### 3.3 Visual Regression Testing

**Package**: `pkg/visualtest`, `pkg/visualtest/parity`  
**Effort**: Small  
**Demo**: `examples/visualregressiontest/`, `examples/paritytest/`

**Files to modify**:

```yaml
# .github/workflows/ci.yml - Add visual tests
- name: Visual Regression Tests
  run: go run ./examples/visualregressiontest/

- name: Cross-Platform Parity
  run: go run ./examples/paritytest/
```

**Verification**:
```bash
go run ./examples/visualregressiontest/
go run ./examples/paritytest/
```

---

## Phase 4: Cleanup (Week 4)

Remove empty/stub packages and consolidate.

### 4.1 Remove Empty Directories

These packages have no implementation and should be removed:

```bash
# Remove empty package directories
rm -rf pkg/audit/            # Empty, use pkg/audit/features
rm -rf pkg/class/            # Empty, use pkg/class/advanced  
rm -rf pkg/companion/        # Empty, use pkg/companion/learning
rm -rf pkg/engine/saves/     # Empty, use pkg/saveload
rm -rf pkg/narrative/        # Empty, use pkg/narrative/branching
```

### 4.2 Evaluate Minimal Packages

Review these packages for consolidation:

| Package | LOC | Action |
|---------|-----|--------|
| `pkg/social` | 626 | Merge into `pkg/social/persistence` |
| `pkg/engine/physics` | 42 | Keep as namespace (has doc.go) |
| `pkg/integration` | 451 | Keep as coordinator package |
| `pkg/rendering` | 445 | Keep as namespace package |

---

## Verification Checklist

After each phase:

- [ ] `go build ./cmd/client && go build ./cmd/server`
- [ ] `go test ./pkg/...`
- [ ] Run client, verify features work
- [ ] Check no new nil panics
- [ ] Verify memory usage still <500MB
- [ ] Confirm 60+ FPS maintained

---

## Integration Template

When adding a new system, follow this pattern:

### System Registration

```go
// 1. Add import at top of file
import "github.com/opd-ai/venture/pkg/path/to/package"

// 2. Create system instance (in initialization function)
mySystem := package.NewSystem(requiredDeps...)

// 3. Register unconditionally (no feature flags!)
game.World.AddSystem(mySystem)

// OR if wrapper needed:
game.World.AddSystem(&mySystemWrapper{system: mySystem})
```

### Component Addition

```go
// In entity creation function (NOT in system Update)
player.AddComponent(package.NewMyComponent(params))
```

### Wrapper Pattern (if needed)

```go
// When system signature doesn't match World.AddSystem expectations
type mySystemWrapper struct {
    system *package.MySystem
}

func (w *mySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
    w.system.Update(deltaTime) // Delegate to actual system
}
```

---

## Files Modified Summary

| Phase | Files Modified | Systems Added |
|-------|----------------|---------------|
| Phase 1 | `cmd/client/handlers.go`, `cmd/server/main.go` | 5 |
| Phase 2 | `cmd/client/handlers.go`, `pkg/engine/entity_spawning.go` | 3 |
| Phase 3 | `.github/workflows/ci.yml` | 0 (CI only) |
| Phase 4 | Directory cleanup | 0 |
| **Total** | ~5 files | 8 systems |

---

## Anti-Patterns to Avoid

- ❌ `if enableFeatureX { ... }` — No conditionals
- ❌ `-enable-*` flags — No toggles  
- ❌ `-sprite-size`, `-tile-size` flags — Hardcode constants
- ❌ "Optional" integration — Everything is mandatory
- ❌ Lazy component initialization in `System.Update()`
- ❌ Ignoring `GetComponent()` boolean return
- ❌ Adding components to only one client (desktop but not mobile)

---

## Success Criteria

- [ ] All 8 dormant packages integrated
- [ ] Zero new nil panics introduced
- [ ] All tests pass: `go test ./pkg/...`
- [ ] Build succeeds: `go build ./cmd/...`
- [ ] Performance maintained: 60+ FPS, <500MB memory
- [ ] No feature flags added
- [ ] CI/CD includes audit tools

---

*Generated: December 2025*
*Estimated Effort: 4 weeks*
*Risk Level: Low (all packages have demos/tests)*
