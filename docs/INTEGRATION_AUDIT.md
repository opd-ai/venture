# Integration Audit - December 2025

## Summary

- **Total pkg/ packages with code**: 97
- **Active (imported by client/server/mobile)**: 87 (89.7%)
- **Dormant/Test-Only**: 10 (10.3%)

**Note:** Integration dramatically improved in V19.0 which integrated all Priority 1 packages.

## Classification Criteria

| Completeness | Criteria |
|-------------|----------|
| **Complete** | Has doc.go, tests, exports, >200 LOC |
| **Partial** | Has exports but missing docs or tests |
| **Stub** | <100 LOC or only type definitions |

---

## Graphics Baseline (Always Active)

All graphics enhancements are unconditionally enabled:

- **Sprite Resolution**: 64x64 pixels (procedurally generated)
- **Tile Resolution**: 64x64 pixels
- **Particle System**: Unconditionally active for combat, magic, environmental effects
- **Dynamic Lighting**: Per-pixel lighting with shadow casting
- **Animation Cache**: Automatic caching of sprite sequences
- **Visual Effects**: Explosions, magic auras, weather particles always rendered
- **Sprite Cache**: LRU caching with 100% hit rate target

---

## Active Packages

These packages are currently imported and used by the game client, server, or mobile builds.

### Core Engine

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/engine` | client, server, mobile | 240,965 | Core ECS framework, all systems and components |
| `pkg/engine/physics/destruction` | client | 1,478 | Destructible environment physics |
| `pkg/engine/physics/fluids` | client, server | 1,907 | Fluid simulation (water, lava) |
| `pkg/engine/physics/vehicle` | client, server | 2,506 | Vehicle physics and movement |
| `pkg/engine/prestige` | client | 1,869 | Prestige/new game+ systems |
| `pkg/engine/qol` | client | 1,799 | Quality of life features |

### Combat & Classes

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/combat` | client, server | 507 | Combat mechanics and damage calculation |
| `pkg/class/advanced` | client | 3,198 | Advanced class progressions |
| `pkg/companion/learning` | client | 2,107 | Companion AI learning system |

### Networking

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/network` | client, server | 22,983 | Core networking, client-server communication |
| `pkg/network/chat` | client | 455 | In-game chat system |
| `pkg/network/federation` | client, server | 8,916 | Cross-server federation |
| `pkg/network/federation/guild` | client, server | 1,221 | Guild federation features |
| `pkg/network/federation/mobile` | client | 1,390 | Mobile-specific federation |
| `pkg/network/federation/webrtc` | client | 4,246 | WebRTC networking |
| `pkg/network/resilience` | server | 1,334 | Network error handling |
| `pkg/network/trade` | client | 1,429 | Player trading system |

### Procedural Generation

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/procgen` | client, server, mobile | 509 | Core procedural generation framework |
| `pkg/procgen/book` | client | 2,430 | Procedural book content |
| `pkg/procgen/building` | client, server | 1,932 | Building layout generation |
| `pkg/procgen/class` | client | 659 | Class template generation |
| `pkg/procgen/companion` | client, server | 492 | Companion generation |
| `pkg/procgen/environment` | client | 3,789 | Environmental asset generation |
| `pkg/procgen/faction` | client | 999 | Faction generation |
| `pkg/procgen/furniture` | client, server | 2,391 | Furniture generation |
| `pkg/procgen/genre` | client | 1,787 | Genre-based theming |
| `pkg/procgen/item` | client, server, mobile | 2,035 | Item generation |
| `pkg/procgen/magic` | client | 3,244 | Magic system generation |
| `pkg/procgen/minigame` | client | 1,094 | Minigame framework |
| `pkg/procgen/minigame/games` | client | 1,985 | Minigame implementations |
| `pkg/procgen/narrative` | client | 1,100 | Narrative content generation |
| `pkg/procgen/puzzle` | client | 2,055 | Puzzle generation |
| `pkg/procgen/quest` | client | 1,505 | Quest generation |
| `pkg/procgen/recipe` | client | 1,049 | Crafting recipe generation |
| `pkg/procgen/skills` | client | 1,983 | Skill tree generation |
| `pkg/procgen/station` | client | 859 | Crafting station generation |
| `pkg/procgen/story` | client | 4,739 | Story arc generation |
| `pkg/procgen/terrain` | client, server, mobile | 14,834 | Terrain/dungeon generation |
| `pkg/procgen/vehicle` | client, server | 1,774 | Vehicle generation |

### Rendering

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/rendering/animation` | client | 2,272 | Animation system |
| `pkg/rendering/cache` | client | 2,210 | Sprite/animation caching |
| `pkg/rendering/display` | client | 785 | Display management |
| `pkg/rendering/lighting` | client | 2,786 | Dynamic lighting |
| `pkg/rendering/palette` | client | 4,349 | Color palette generation |
| `pkg/rendering/parallel` | client | 1,129 | Parallel rendering |
| `pkg/rendering/particles` | client | 8,196 | Particle effects |
| `pkg/rendering/patterns` | client | 1,506 | Texture patterns |
| `pkg/rendering/pool` | client | 591 | Image pooling |
| `pkg/rendering/postprocess` | client | 3,034 | Post-processing effects |
| `pkg/rendering/quality` | client | 1,951 | Quality settings |
| `pkg/rendering/shapes` | client | 2,265 | Shape rendering |
| `pkg/rendering/sprites` | client, server | 15,479 | Sprite generation |
| `pkg/rendering/ui` | client | 11,957 | UI components |

### World & Social

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/world` | client | 4,776 | World state management |
| `pkg/world/housing` | client, server | 3,337 | Player housing |
| `pkg/world/raids` | client | 3,043 | Raid encounters |
| `pkg/world/territory` | client | 2,242 | Territory control |
| `pkg/social/persistence` | client, server | 3,747 | Social data persistence |

### Integration Packages

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/integration/companion_housing` | client, server | 1,137 | Companion-housing integration |
| `pkg/integration/guild_housing` | client, server | 1,358 | Guild-housing integration |
| `pkg/integration/housing_crafting` | client, server | 1,192 | Housing-crafting integration |
| `pkg/integration/narrative_world` | client | 2,449 | Narrative-world integration |
| `pkg/integration/political_warfare` | client | 1,553 | Political warfare integration |
| `pkg/integration/trade_routes` | client | 1,497 | Trade route integration |

### Infrastructure

| Package | Used By | LOC | Purpose |
|---------|---------|-----|---------|
| `pkg/logging` | client, server, mobile | 545 | Structured logging |
| `pkg/saveload` | client | 3,800 | Save/load system |
| `pkg/version` | client, server | 85 | Version information |
| `pkg/mobile` | client | 6,689 | Mobile platform support |
| `pkg/hostplay` | client | 2,497 | Host-and-play mode |
| `pkg/balance` | server | 1,232 | Game balance tuning |
| `pkg/migration` | server | 619 | Save migration |
| `pkg/modding` | server | 2,349 | Mod system |
| `pkg/security` | server | 1,530 | Security validation |
| `pkg/stability` | server | 599 | Server stability |
| `pkg/ux` | server | 1,571 | UX improvements |
| `pkg/narrative/branching` | client | 2,033 | Branching narratives |

---

## Previously Dormant Packages (Now Integrated)

The following packages were dormant prior to V19.0 but are now fully integrated:

### Priority 1 Packages ✅ (Integrated in V19.0)

All Priority 1 packages were integrated in Version 19.0 (December 16, 2025):

| Package | Status | Integrated In |
|---------|--------|---------------|
| `pkg/audio/synthesis` | ✅ Active | V19.0 |
| `pkg/procgen/entity` | ✅ Active | V19.0 |
| `pkg/procgen/dialog` | ✅ Active | V19.0 |
| `pkg/procgen/legendary` | ✅ Active | V19.0 |
| `pkg/world/economy` | ✅ Active | V19.0 |
| `pkg/integration/choice_consequences` | ✅ Active | V19.0 |
| `pkg/integration/guild_vehicle` | ✅ Active | V19.0 |
| `pkg/integration/world_events` | ✅ Active | V19.0 |

### Priority 2: Infrastructure/Testing Packages

#### `pkg/audit/features`
- **Completeness**: Complete (2,048 LOC, doc.go, 1 test)
- **Status**: Truly Dormant - feature audit framework
- **Dependencies**: None
- **Integration Type**: Development tooling
- **Integration Steps**:
  1. Add to CI/CD pipeline for automated feature auditing
  2. Import in build scripts for validation
- **Effort**: Small
- **Demo**: `examples/featureaudit/`

#### `pkg/procgen/audit`
- **Completeness**: Complete (1,878 LOC, doc.go, 3 tests)
- **Status**: Truly Dormant - procedural generation auditing
- **Dependencies**: None
- **Integration Type**: Development tooling
- **Integration Steps**:
  1. Add to test suite for determinism validation
  2. Run in CI/CD to verify generator consistency
- **Effort**: Small
- **Demo**: `examples/auditrunner/`

#### `pkg/visualtest`
- **Completeness**: Complete (5,176 LOC, doc.go, 7 tests)
- **Status**: Test-only usage
- **Dependencies**: Multiple rendering packages
- **Integration Type**: Testing infrastructure
- **Integration Steps**:
  1. Enable in CI/CD for visual regression testing
  2. Already used indirectly via test suites
- **Effort**: Small
- **Demo**: `examples/visualregressiontest/`

#### `pkg/visualtest/parity`
- **Completeness**: Complete (1,340 LOC, doc.go, 3 tests)
- **Status**: Test-only usage
- **Dependencies**: `pkg/visualtest`
- **Integration Type**: Testing infrastructure
- **Integration Steps**:
  1. Enable in CI/CD for cross-platform parity validation
- **Effort**: Small
- **Demo**: `examples/paritytest/`

### Priority 3: Stub/Empty Packages

These packages exist as directory structures but lack implementation:

| Package | Status | Action Required |
|---------|--------|-----------------|
| `pkg/audit` | Empty (0 Go files) | Remove or implement |
| `pkg/class` | Empty (0 Go files) | Remove (use `pkg/class/advanced`) |
| `pkg/companion` | Empty (0 Go files) | Remove (use `pkg/companion/learning`) |
| `pkg/engine/physics` | Stub (42 LOC, doc only) | Keep as namespace |
| `pkg/engine/saves` | Empty (0 Go files) | Remove (use `pkg/saveload`) |
| `pkg/integration` | Parent namespace (451 LOC) | Keep as coordinator |
| `pkg/narrative` | Empty (0 Go files) | Remove (use `pkg/narrative/branching`) |
| `pkg/social` | Minimal (626 LOC) | Evaluate for removal |

### Priority 4: Render/Display Packages (Partial Integration)

#### `pkg/rendering`
- **Completeness**: Minimal (445 LOC, doc.go, 1 test)
- **Status**: Indirectly Used - parent namespace
- **Dependencies**: Ebiten
- **Integration Type**: Keep as namespace package
- **Effort**: None needed

#### `pkg/rendering/tiles`
- **Completeness**: Complete (6,602 LOC, doc.go, 6 tests)
- **Status**: Indirectly Used - imported by `pkg/visualtest`
- **Dependencies**: `pkg/rendering/palette`
- **Integration Type**: Already integrated via terrain rendering
- **Integration Steps**:
  1. Verify tile generation used by terrain system
  2. No additional integration needed
- **Effort**: None (already working)

---

## Dependency Graph

Integration order must respect these dependencies:

```
Level 0 (No dependencies - integrate first):
├── pkg/audio/synthesis
├── pkg/procgen/dialog
├── pkg/world/economy
├── pkg/audit/features
├── pkg/procgen/audit

Level 1 (Depends on Level 0):
├── pkg/procgen/entity → pkg/procgen, pkg/procgen/item
├── pkg/procgen/legendary → pkg/procgen, pkg/world/raids

Level 2 (Depends on active packages):
├── pkg/integration/choice_consequences → pkg/narrative/branching
├── pkg/integration/guild_vehicle → pkg/engine/physics/vehicle
├── pkg/integration/world_events → pkg/world
```

---

## Component Integration Checklist

When integrating a new component/system, verify:

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

## Common Integration Failure Modes

### 1. Nil Component Panic
**Symptom**: `panic: interface conversion: engine.Component is nil`

**Fix**:
1. Add defensive nil check: `if ok && comp != nil`
2. Add component during entity creation, not lazily in Update()

### 2. Component Not Found After Adding
**Symptom**: `GetEntitiesWith("new_component")` returns empty

**Fix**:
1. Call `world.InvalidateQueryCache()` after adding components
2. Better: Add components during entity creation

### 3. Mobile/Desktop Feature Parity
**Symptom**: Feature works on desktop but crashes on mobile

**Fix**: Always integrate in both:
- `cmd/client/handlers.go` — Desktop client
- `cmd/mobile/mobile.go` — Mobile client

### 4. System Registration Order
**Symptom**: System A depends on data from System B but runs first

**Fix**: Register dependent systems AFTER their dependencies

---

## Statistics

| Metric | Value |
|--------|-------|
| Total Packages | 97 |
| Active Packages | 87 (89.7%) |
| Test/Infra Only | 10 (10.3%) |
| Total LOC | ~435,000 |
| Packages with doc.go | 89 (91.8%) |
| Packages with tests | 85 (87.6%) |
| Average test coverage | 82.4% |

---

## Registered Systems Count

| Location | Systems Registered |
|----------|-------------------|
| cmd/client/ | 120+ systems |
| cmd/server/ | 75+ systems |
| Total Unique | ~140 systems |

---

*Updated: December 16, 2025*
*Version: 20.0 (Post V19.0 Integration)*
