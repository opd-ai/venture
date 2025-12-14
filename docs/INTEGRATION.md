# Venture Integration Audit — Detailed Execution Plan

## Objective
Generate two documents:
1. `docs/INTEGRATION_AUDIT.md` — Comprehensive inventory of all `pkg/` packages with activation status
2. `docs/PLAN.md` — Phased roadmap to integrate dormant packages as always-on baseline features

## Execution Mode
**Report Generation** — Analyze codebase and produce documents. Do not modify source code.

## Core Principle
**All features are baseline features.** Dormant packages become unconditionally enabled upon integration. No feature flags, no toggles, no optional activation. If a package exists, it should be active.

---

## Critical Integration Lessons Learned

### Component Initialization: NEVER Use Lazy Initialization

**Problem Discovered**: The `AdaptiveSoundtrackSystem` used lazy initialization - creating the `AdaptiveSoundtrackComponent` on first `Update()` call. This caused nil pointer panics due to ECS query cache staleness.

**Root Cause**: 
1. `Entity.AddComponent()` does NOT invalidate the `World`'s query cache (Entity has no reference to World)
2. `GetEntitiesWith()` uses cached results that can become stale when components are added/removed
3. If cache returns an entity but the component value is nil, type assertions panic

**Solution**: **Always add components during entity creation, NEVER during system Update()**.

```go
// ❌ BAD: Lazy initialization in System.Update()
func (s *MySoundSystem) Update(deltaTime float64) {
    entities := s.world.GetEntitiesWith("my_sound")
    if len(entities) == 0 {
        // Fall back to finding player and adding component dynamically
        players := s.world.GetEntitiesWith("position", "health")
        player := players[0]
        player.AddComponent(NewMySoundComponent()) // DANGER: Cache not invalidated!
    }
}

// ✅ GOOD: Add component during entity creation
func createPlayerEntity(game *engine.EbitenGame, ...) *engine.Entity {
    player := game.World.CreateEntity()
    player.AddComponent(&engine.PositionComponent{X: x, Y: y})
    player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
    // ... other core components ...
    
    // Audio components - add during creation, not lazily
    player.AddComponent(engine.NewAdaptiveSoundtrackComponent(genreID))
    player.AddComponent(engine.NewMusicTriggerComponent())
    
    return player
}
```

### Defensive Component Access Pattern

**Always** check both the boolean return AND nil before type assertion:

```go
// ❌ BAD: Ignores boolean, panics on nil
comp, _ := entity.GetComponent("my_component")
myComp := comp.(*MyComponent) // PANIC if comp is nil

// ✅ GOOD: Check both ok AND nil
comp, ok := entity.GetComponent("my_component")
if ok && comp != nil {
    myComp := comp.(*MyComponent)
    // Safe to use myComp
}
```

### Component Integration Checklist

When integrating a new component type, verify:

- [ ] Component is added during entity creation (not lazily in Update)
- [ ] All `GetComponent()` calls check both `ok` AND `comp != nil`
- [ ] Component is added to ALL relevant entity creation paths:
  - [ ] `cmd/client/handlers.go` - desktop client player creation
  - [ ] `cmd/mobile/mobile.go` - mobile client player creation  
  - [ ] `cmd/server/main.go` - server-side entity creation (if applicable)
  - [ ] `pkg/engine/entity_spawning.go` - procedural entity spawning
- [ ] Corresponding system is unconditionally registered via `AddSystem()`
- [ ] Tests verify component exists after entity creation

---

## Phase 1: Establish Baseline (Active Systems)

### Step 1.1: Extract Client Imports
Identify all packages currently used by the game client:
```bash
grep -rh "github.com/opd-ai/venture/pkg/" cmd/client/ --include="*.go" | \
  grep -oP '"[^"]*"' | tr -d '"' | sort -u > /tmp/client_imports.txt
```

### Step 1.2: Extract Server Imports
Identify all packages currently used by the game server:
```bash
grep -rh "github.com/opd-ai/venture/pkg/" cmd/server/ --include="*.go" | \
  grep -oP '"[^"]*"' | tr -d '"' | sort -u > /tmp/server_imports.txt
```

### Step 1.3: Identify Registered ECS Systems
Find all systems added to the game world:
```bash
grep -rn "AddSystem\|RegisterSystem\|world\.Add" cmd/client/ cmd/server/ --include="*.go"
```

### Step 1.4: Find Existing Feature Flags (To Remove)
Identify flags gating functionality that should become always-on:
```bash
grep -rn 'enable\|disable\|toggle' cmd/client/ cmd/server/ --include="*.go" | grep -i flag
```
**Action**: These flags should be removed during integration; features become unconditional.

### Step 1.5: Audit Graphics Configuration
Identify current sprite/tile sizes and visual effect settings:
```bash
grep -rn "SpriteSize\|TileSize\|sprite.*width\|sprite.*height" pkg/rendering/ cmd/client/ --include="*.go"
grep -rn "particle\|lighting\|shadow\|effect" pkg/rendering/ --include="*.go"
```
**Action**: Ensure latest enhancements (larger sprites/tiles, improved effects) are hardcoded defaults, not configurable options.

---

## Phase 2: Inventory All Packages

### Step 2.1: Enumerate Package Directories
```bash
find pkg -type d -mindepth 1 -maxdepth 3 | sort > /tmp/all_packages.txt
```

### Step 2.2: For Each Package, Collect Metadata

| Check | Command | Meaning |
|-------|---------|---------|
| Has Go files | `ls pkg/<path>/*.go 2>/dev/null \| wc -l` | Package exists |
| Has doc.go | `test -f pkg/<path>/doc.go` | Documented |
| Has tests | `ls pkg/<path>/*_test.go 2>/dev/null \| wc -l` | Tested |
| Exports public API | `grep -l '^func [A-Z]' pkg/<path>/*.go` | Usable externally |
| Line count | `wc -l pkg/<path>/*.go 2>/dev/null \| tail -1` | Implementation size |

### Step 2.3: Classify Package Completeness
- **Complete**: Has doc.go, tests, exports, >200 LOC
- **Partial**: Has exports but missing docs or tests
- **Stub**: <100 LOC or only type definitions

---

## Phase 3: Cross-Reference (Find Dormant Packages)

### Step 3.1: Compute Set Difference
```bash
comm -23 /tmp/all_packages.txt <(cat /tmp/client_imports.txt /tmp/server_imports.txt | sort -u)
```

### Step 3.2: Check for Indirect Usage
```bash
grep -r "pkg/<dormant_package>" pkg/ --include="*.go" | grep -v "_test.go"
```
**Classify**: 
- "Truly Dormant" = not imported anywhere
- "Indirectly Used" = imported by other pkg/ but not client/server
- "Test Only" = only in *_test.go files

### Step 3.3: Check Example/Demo Usage
```bash
grep -r "pkg/<dormant_package>" examples/ cmd/*test/ --include="*.go"
```
**Note**: Demos prove functionality works — activation should be straightforward.

---

## Phase 4: Dependency Analysis

### Step 4.1: Build Dependency Graph
```bash
go list -f '{{.ImportPath}}: {{join .Imports ", "}}' ./pkg/<path>/...
```

### Step 4.2: Identify Activation Order
If pkg/A imports pkg/B, then B must be integrated before A.
Build topological order for integration phases.

### Step 4.3: Map to Integration Points
```bash
grep -rn "Component\|System" pkg/<path>/*.go | head -20
```
- Defines Components → Entity integration
- Defines Systems → World.AddSystem() unconditionally
- Defines Generators → Procgen pipeline integration

---

## Phase 5: Determine Integration Steps

### Step 5.1: Categorize Integration Type

| Integration Type | Detection | Integration Pattern |
|------------------|-----------|---------------------|
| **ECS System** | Has `Update(dt)` | Unconditional `AddSystem()` |
| **Procgen Generator** | Implements `Generator` | Add to generation pipeline |
| **Rendering Layer** | Imports `ebiten` | Add to render loop |
| **Network Protocol** | Imports `network/` | Register handlers at startup |
| **UI Component** | In `rendering/ui/` | Add to UI initialization |
| **Audio** | In `audio/` | Connect to AudioManager |
| **Save/Load** | Has serialization | Register with SaveLoad |

### Step 5.2: Write Concrete Integration Steps
For each dormant package, specify exact changes:
```
Example for `pkg/companion/learning/`:
1. Add import in cmd/client/main.go
2. Create LearningSystem: `learningSys := learning.NewSystem(companionSys)`
3. Register unconditionally: `world.AddSystem(learningSys)`
```
**No feature flags. No conditionals. Just add it.**

### Step 5.3: Estimate Effort
- **Small (S)**: Import + 1-2 line registration
- **Medium (M)**: New initialization code or UI wiring
- **Large (L)**: Multiple system integrations

### Step 5.4: ECS Query Cache Awareness

**Critical**: The ECS query cache is invalidated when:
- Entities are added/removed from World
- `World.Update()` processes pending entity additions

The cache is **NOT** invalidated when:
- Components are added/removed from existing entities via `Entity.AddComponent()`
- Component values are modified

**Implication**: Systems that add components during `Update()` may experience stale cache results on subsequent frames. Always add required components during entity creation.

### Step 5.5: Complete Integration Verification

After integrating a component/system, verify with this checklist:

```bash
# 1. Verify component is added during entity creation
grep -rn "AddComponent.*New<ComponentName>" cmd/client/ cmd/mobile/ cmd/server/

# 2. Verify system is unconditionally registered  
grep -rn "AddSystem.*New<SystemName>\|world\.AddSystem" cmd/client/ cmd/server/

# 3. Verify defensive nil checks in system Update()
grep -A3 "GetComponent.*<component_type>" pkg/engine/<system_file>.go

# 4. Run tests for the integrated system
go test -v ./pkg/engine/... -run <SystemName>
```

---

## Phase 6: Review Recent Changes

### Step 6.1: Check Roadmap Context
```bash
cat docs/ROADMAP*.md | grep -A5 -B5 "Phase\|TODO\|WIP\|Complete"
```

### Step 6.2: Git History
```bash
git log --since="2 months ago" --name-only --pretty=format: -- pkg/ | sort | uniq -c | sort -rn | head -30
```
Recently modified = likely being activated already.

---

## Phase 7: Graphics Enhancement Baseline

### Step 7.1: Identify Visual Enhancement Packages
```bash
grep -rn "particle\|lighting\|shadow\|animation\|cache" pkg/rendering/ --include="*.go" | grep "package\|type.*System"
```

### Step 7.2: Enforce Default Graphics Settings
All graphics enhancements must be unconditionally enabled with enhanced defaults:

| Enhancement | Default Value | Location | Action |
|-------------|---------------|----------|--------|
| **Sprite Size** | 64x64 (up from 32x32) | `pkg/rendering/sprites/` | Hardcode `DefaultSpriteSize = 64` |
| **Tile Size** | 64x64 (up from 32x32) | `pkg/rendering/` | Hardcode `DefaultTileSize = 64` |
| **Particle Effects** | Always enabled | `pkg/rendering/particles/` | Remove any enable flags |
| **Lighting System** | Always enabled | `pkg/rendering/lighting/` | Unconditional registration |
| **Shadow Rendering** | Always enabled | `pkg/rendering/lighting/` | Hardcode `EnableShadows = true` |
| **Animation Cache** | Always enabled | `pkg/rendering/cache/` | Unconditional initialization |
| **Sprite Cache** | Always enabled | `pkg/rendering/cache/` | Unconditional initialization |

### Step 7.3: Remove Graphics Flags
```bash
grep -rn "sprite.*size.*flag\|tile.*size.*flag\|enable.*particle\|enable.*lighting" cmd/client/ --include="*.go"
```
**Action**: Delete all CLI flags for graphics settings. Values are constants, not variables.

### Step 7.4: Verify Enhanced Rendering Active
```
Integration checklist:
- [ ] `pkg/rendering/particles` imported in cmd/client/main.go
- [ ] `pkg/rendering/lighting` system unconditionally added
- [ ] `pkg/rendering/cache` initialized at startup
- [ ] `pkg/rendering/animation` system registered
- [ ] All sprite generators use 64x64 default
- [ ] No `-sprite-size`, `-enable-particles`, or similar flags exist
```

### Step 7.5: Document Visual Baseline
Add to `docs/INTEGRATION_AUDIT.md`:
```markdown
## Graphics Baseline (Always Active)
- **Sprite Resolution**: 64x64 pixels (procedurally generated)
- **Tile Resolution**: 64x64 pixels
- **Particle System**: Unconditionally active for combat, magic, environmental effects
- **Dynamic Lighting**: Per-pixel lighting with shadow casting
- **Animation Cache**: Automatic caching of sprite sequences
- **Visual Effects**: Explosions, magic auras, weather particles always rendered
```

---

## Phase 8: Generate Output Documents

### Output 1: `docs/INTEGRATION_AUDIT.md`

```markdown
# Integration Audit - December 2025

## Summary
- **Total pkg/ directories**: X
- **Active**: Y  
- **Dormant (requires integration)**: Z

## Active Packages
| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/engine` | client, server | Core ECS |
| ... | ... | ... |

## Graphics Baseline (Always Active)
- **Sprite Resolution**: 64x64 pixels (procedurally generated)
- **Tile Resolution**: 64x64 pixels
- **Particle System**: Unconditionally active for combat, magic, environmental effects
- **Dynamic Lighting**: Per-pixel lighting with shadow casting
- **Animation Cache**: Automatic caching of sprite sequences
- **Visual Effects**: Explosions, magic auras, weather particles always rendered

## Dormant Packages (To Integrate)

### `pkg/companion/learning`
- **Completeness**: Complete
- **Blocker**: Not registered as system
- **Dependencies**: `pkg/companion` (active)
- **Integration**:
  1. Import in `cmd/client/main.go`
  2. `world.AddSystem(learning.NewSystem(companionSys))`
- **Effort**: Small

### `pkg/engine/physics/destruction`
...

## Dependency Graph
[Mermaid diagram showing integration order]
```

### Output 2: `docs/PLAN.md`

```markdown
# Integration Plan - December 2025

## Principle
All packages become unconditionally active. No feature flags. No optional toggles.

## Phase 1: Foundation (Week 1)
Activate core systems that others depend on.

- [ ] `pkg/engine/physics/destruction`
  - File: `cmd/client/main.go`
  - Add: `world.AddSystem(destruction.NewSystem())`
  - Verify: `go run ./cmd/destructiontest`

- [ ] `pkg/class/advanced`
  - File: `cmd/client/main.go`
  - Add: `world.AddSystem(advancedclass.NewSystem())`

- [ ] **Graphics Enhancement Baseline**
  - File: `pkg/rendering/sprites/generator.go`
  - Change: `const DefaultSpriteSize = 64` (remove flag)
  - File: `pkg/rendering/cache/cache.go`
  - Add: Unconditional initialization in `cmd/client/main.go`
  - File: `pkg/rendering/particles/system.go`
  - Add: `world.AddSystem(particles.NewSystem())`
  - File: `pkg/rendering/lighting/system.go`
  - Add: `world.AddSystem(lighting.NewSystem())`
  - Verify: Visual quality improvements visible at startup

## Phase 2: Social Systems (Week 2)
- [ ] `pkg/network/chat`
- [ ] `pkg/social/persistence`
...

## Phase 3: Content Generation (Week 3)
- [ ] `pkg/procgen/companion`
- [ ] `pkg/procgen/dialog`
...

## Verification
After each phase:
- [ ] `go build ./cmd/client && go build ./cmd/server`
- [ ] `go test ./pkg/...`
- [ ] Run client, verify features work
- [ ] Confirm 64x64 sprites rendering
- [ ] Verify particle effects active
- [ ] Check lighting system operational
```

---

## Success Criteria
- [ ] All `pkg/` subdirectories inventoried
- [ ] Every dormant package has exact file + code changes
- [ ] Integration order respects dependencies
- [ ] No feature flags in integration steps
- [ ] All features become always-on baseline
- [ ] Graphics enhancements (64x64 sprites, particles, lighting) unconditionally enabled
- [ ] No CLI flags for sprite size, tile size, or visual effects
- [ ] Excludes `cmd/*test/` (standalone tools, not features)

## Anti-Patterns to Avoid
- ❌ `if enableFeatureX { ... }` — No conditionals
- ❌ `-enable-*` flags — No toggles
- ❌ `-sprite-size`, `-tile-size` flags — Hardcode constants
- ❌ "Optional" integration — Everything is mandatory
- ❌ Lazy component initialization in `System.Update()` — Causes cache staleness
- ❌ Ignoring `GetComponent()` boolean return — Causes nil panics
- ❌ Adding components to only one client (desktop but not mobile)
- ✅ Unconditional `AddSystem()` calls
- ✅ Direct imports without guards
- ✅ Always-on initialization
- ✅ Hardcoded enhanced graphics defaults
- ✅ Components added during entity creation
- ✅ Defensive nil checks: `if ok && comp != nil`
- ✅ Consistent component addition across all clients

---

## Common Integration Failure Modes

### 1. Nil Component Panic
**Symptom**: `panic: interface conversion: engine.Component is nil, not *engine.FooComponent`

**Cause**: System calls `GetEntitiesWith("foo")`, gets entity from stale cache, component is nil.

**Fix**:
1. Add defensive nil check in system: `if ok && comp != nil`
2. Add component during entity creation, not lazily in Update()
3. Verify component added in ALL entity creation paths

### 2. Component Not Found After Adding
**Symptom**: `GetEntitiesWith("new_component")` returns empty even after `AddComponent()`

**Cause**: ECS query cache not invalidated after component addition to existing entity.

**Fix**:
1. Call `world.InvalidateQueryCache()` after adding components to existing entities
2. Better: Add components during entity creation before first `Update()`

### 3. Mobile/Desktop Feature Parity
**Symptom**: Feature works on desktop but crashes on mobile (or vice versa)

**Cause**: Component/system only integrated in one client entry point.

**Fix**: Always integrate in both:
- `cmd/client/handlers.go` — Desktop client
- `cmd/mobile/mobile.go` — Mobile client
- Verify with: `grep -l "NewFooComponent" cmd/client/ cmd/mobile/`

### 4. System Registration Order
**Symptom**: System A depends on data from System B but runs first

**Cause**: Systems registered in wrong order; Update() order matters.

**Fix**: Register dependent systems AFTER their dependencies:
```go
// ✅ Correct order
world.AddSystem(audioManagerSystem)           // Provides audio context
world.AddSystem(adaptiveSoundtrackSystem)     // Depends on audio context
world.AddSystem(musicTriggerSystem)           // Depends on soundtrack state
```

### 5. Missing Wrapper for System Interface
**Symptom**: System implements `Update(deltaTime float64)` but World expects `Update(entities []*Entity, deltaTime float64)`

**Cause**: System uses self-contained pattern but isn't wrapped.

**Fix**: Create wrapper struct:
```go
type fooSystemWrapper struct {
    system *FooSystem
}

func (w *fooSystemWrapper) Update(entities []*Entity, deltaTime float64) {
    w.system.Update(deltaTime) // Delegate to actual system
}

// Register wrapped version
world.AddSystem(&fooSystemWrapper{system: fooSystem})
```

---

## Entity Creation Integration Template

When integrating a new component, add it following this template pattern:

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

**Files to update for player components:**
1. `cmd/client/handlers.go` — `createPlayerEntity()` and/or `addPlayerComponents()`
2. `cmd/mobile/mobile.go` — Player creation section
3. `pkg/engine/entity_spawning.go` — If component applies to NPCs/enemies

**Files to update for system registration:**
1. `cmd/client/handlers.go` — `initializeV*Systems()` functions
2. `cmd/client/util.go` — System wrapper definitions (if needed)
3. `cmd/server/main.go` — Server-side systems (if applicable)