# Integration Plan - December 2025

## Principle

**All packages are baseline features.** This plan documents the current integration status and identifies any remaining work. Per audit findings, the codebase is already 92% integrated—most "dormant" packages are either indirectly used or testing infrastructure.

---

## Executive Summary

| Status | Count | Action Required |
|--------|-------|-----------------|
| **Fully Integrated** | 79 packages | None - actively imported |
| **Indirectly Active** | 15 packages | None - used by other packages |
| **Test Infrastructure** | 5 packages | None - intentionally separate |
| **Stub/Empty** | 3 packages | None - placeholder directories |
| **Requires Integration** | 0 packages | ✅ **No action needed** |

**Conclusion**: The integration audit found no truly dormant packages requiring integration. All substantial code is already active through direct imports or indirect usage chains.

---

## Current Integration Status

### ✅ Phase 1: Foundation (COMPLETE)

All core systems are fully integrated:

| Component | Status | Integration Point |
|-----------|--------|-------------------|
| `pkg/engine` | ✅ Active | `cmd/client/handlers.go`, `cmd/server/main.go` |
| `pkg/procgen` | ✅ Active | All procgen subpackages |
| `pkg/logging` | ✅ Active | Universal logging |
| `pkg/version` | ✅ Active | Version display |

### ✅ Phase 2: Procedural Generation (COMPLETE)

All generators integrated:

| Generator | Status | Used By |
|-----------|--------|---------|
| `pkg/procgen/terrain` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/entity` | ✅ Active | `engine/entity_spawning.go`, `world/raids` |
| `pkg/procgen/item` | ✅ Active | Direct import in cmd/client, cmd/server |
| `pkg/procgen/magic` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/skills` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/quest` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/dialog` | ✅ Active | `engine/npcdialog_system.go`, `engine/book_spawning.go` |
| `pkg/procgen/companion` | ✅ Active | Direct import in cmd/client, cmd/server |
| `pkg/procgen/building` | ✅ Active | Direct import in cmd/client, cmd/server |
| `pkg/procgen/furniture` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/environment` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/story` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/narrative` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/vehicle` | ✅ Active | Direct import in cmd/client, cmd/server |
| `pkg/procgen/legendary` | ✅ Active | `engine/legendary_quest_system.go` |
| `pkg/procgen/book` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/puzzle` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/minigame` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/minigame/games` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/class` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/genre` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/faction` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/recipe` | ✅ Active | Direct import in cmd/client |
| `pkg/procgen/station` | ✅ Active | Direct import in cmd/client |

### ✅ Phase 3: Rendering & Visual (COMPLETE)

All rendering systems integrated with 64x64 sprite baseline:

| Package | Status | System Registration |
|---------|--------|---------------------|
| `pkg/rendering/sprites` | ✅ Active | Sprite generation pipeline |
| `pkg/rendering/tiles` | ✅ Active | `engine/terrain_render_system.go` |
| `pkg/rendering/animation` | ✅ Active | `animationSystemWrapper` |
| `pkg/rendering/cache` | ✅ Active | Sprite/tile caching |
| `pkg/rendering/particles` | ✅ Active | `particleSystem` |
| `pkg/rendering/lighting` | ✅ Active | `lightingAdapter`, `shadowSystem` |
| `pkg/rendering/postprocess` | ✅ Active | Direct import |
| `pkg/rendering/ui` | ✅ Active | UI component generation |
| `pkg/rendering/palette` | ✅ Active | Color palette system |
| `pkg/rendering/shapes` | ✅ Active | Shape primitives |
| `pkg/rendering/patterns` | ✅ Active | Texture patterns |
| `pkg/rendering/quality` | ✅ Active | `qualitySystem` |
| `pkg/rendering/parallel` | ✅ Active | Parallel rendering |
| `pkg/rendering/pool` | ✅ Active | Object pooling |
| `pkg/rendering/display` | ✅ Active | Display configuration |

**Graphics Baseline (Hardcoded)**:
- Sprite resolution: 64x64 pixels
- Tile resolution: 64x64 pixels
- Particle effects: Always enabled
- Dynamic lighting: Always enabled with shadows
- Animation caching: Automatic

### ✅ Phase 4: Audio (COMPLETE)

All audio systems integrated:

| Package | Status | System Registration |
|---------|--------|---------------------|
| `pkg/audio` | ✅ Active | `audioManagerSystem` |
| `pkg/audio/music` | ✅ Active | `adaptiveSoundtrackSystem`, `musicTriggerSystem` |
| `pkg/audio/sfx` | ✅ Active | `positionalAudioSystem`, `reverbSystem` |
| `pkg/audio/synthesis` | ✅ Active | Used by music/sfx packages |

### ✅ Phase 5: Network & Multiplayer (COMPLETE)

All network systems integrated:

| Package | Status | System Registration |
|---------|--------|---------------------|
| `pkg/network` | ✅ Active | Core networking |
| `pkg/network/federation` | ✅ Active | Federation system |
| `pkg/network/federation/guild` | ✅ Active | Guild federation |
| `pkg/network/federation/mobile` | ✅ Active | `mobileFederationSystem` |
| `pkg/network/federation/webrtc` | ✅ Active | WebRTC transport |
| `pkg/network/chat` | ✅ Active | `networkChatSystem` |
| `pkg/network/trade` | ✅ Active | `networkTradeSystem` |
| `pkg/network/resilience` | ✅ Active | Server resilience |

### ✅ Phase 6: World & Social (COMPLETE)

All world and social systems integrated:

| Package | Status | System Registration |
|---------|--------|---------------------|
| `pkg/world` | ✅ Active | World state management |
| `pkg/world/housing` | ✅ Active | Housing system |
| `pkg/world/economy` | ✅ Active | `economySystem` |
| `pkg/world/raids` | ✅ Active | `raidSystem` |
| `pkg/world/territory` | ✅ Active | `territorySystem` |
| `pkg/social` | ✅ Active | Social base (via persistence) |
| `pkg/social/persistence` | ✅ Active | Social data persistence |

### ✅ Phase 7: Integration Packages (COMPLETE)

All cross-system integration packages active:

| Package | Status | System Registration |
|---------|--------|---------------------|
| `pkg/integration` | ✅ Active | Integration framework |
| `pkg/integration/choice_consequences` | ✅ Active | `choiceConsequencesSystem` |
| `pkg/integration/companion_housing` | ✅ Active | Companion housing |
| `pkg/integration/guild_housing` | ✅ Active | Guild housing |
| `pkg/integration/guild_vehicle` | ✅ Active | `guildVehicleSystem` |
| `pkg/integration/housing_crafting` | ✅ Active | Housing crafting |
| `pkg/integration/narrative_world` | ✅ Active | `narrativeWorldSystem` |
| `pkg/integration/political_warfare` | ✅ Active | `politicalWarfareSystem` |
| `pkg/integration/trade_routes` | ✅ Active | Trade routes |
| `pkg/integration/world_events` | ✅ Active | `worldEventsSystem` |

### ✅ Phase 8: Advanced Features (COMPLETE)

All advanced game systems integrated:

| Package | Status | System Registration |
|---------|--------|---------------------|
| `pkg/combat` | ✅ Active | Combat calculations |
| `pkg/class/advanced` | ✅ Active | `advancedClassSystem` |
| `pkg/companion/learning` | ✅ Active | `companionLearningSys`, `companionLearningSystem` |
| `pkg/narrative/branching` | ✅ Active | `branchingNarrativeSystem` |
| `pkg/engine/prestige` | ✅ Active | `prestigeSystem` |
| `pkg/engine/qol` | ✅ Active | `qolSystem` |
| `pkg/engine/physics/destruction` | ✅ Active | `destructionSystem` |
| `pkg/engine/physics/fluids` | ✅ Active | `fluidSimulator` |
| `pkg/engine/physics/vehicle` | ✅ Active | `vehicleMovementSys`, `vehicleDurabilitySys` |

### ✅ Phase 9: Platform & Infrastructure (COMPLETE)

All platform support integrated:

| Package | Status | Used By |
|---------|--------|---------|
| `pkg/mobile` | ✅ Active | `cmd/mobile/` |
| `pkg/hostplay` | ✅ Active | Host-and-play mode |
| `pkg/saveload` | ✅ Active | Game persistence |
| `pkg/modding` | ✅ Active | Mod system |
| `pkg/security` | ✅ Active | Security features |
| `pkg/balance` | ✅ Active | Game balance |
| `pkg/migration` | ✅ Active | Data migration |
| `pkg/stability` | ✅ Active | Stability monitoring |
| `pkg/ux` | ✅ Active | UX improvements |

---

## Testing Infrastructure (Not Runtime)

These packages are intentionally NOT integrated into runtime—they are testing and audit tools:

| Package | Purpose | Usage |
|---------|---------|-------|
| `pkg/visualtest` | Visual regression testing | `examples/visualregressiontest/` |
| `pkg/visualtest/parity` | Cross-platform parity | `examples/paritytest/` |
| `pkg/audit/features` | Feature completeness validation | `examples/featureaudit/` |
| `pkg/procgen/audit` | Generator determinism testing | `examples/auditrunner/` |
| `pkg/engine/performance` | Performance profiling | `examples/performancetest/` |

**These are run via example programs during development and CI/CD, not at game runtime.**

---

## Stub Packages (Intentionally Empty)

These are parent directories or placeholder packages:

| Package | Status | Notes |
|---------|--------|-------|
| `pkg/audit` | Stub | Parent for `audit/features` |
| `pkg/class` | Stub | Types for `class/advanced` |
| `pkg/companion` | Stub | Types for `companion/learning` |
| `pkg/narrative` | Stub | Types for `narrative/branching` |
| `pkg/engine/physics` | Stub | Parent for physics subpackages |
| `pkg/engine/saves` | Stub | Planned future feature |

**No action required—these are architectural placeholders.**

---

## Verification Checklist

After any future integration work, verify:

### Build Verification
```bash
# Build all targets
go build ./cmd/client && go build ./cmd/server && go build ./cmd/mobile

# Run tests
go test ./pkg/... -race -cover

# Check for unused imports
go vet ./...
```

### Runtime Verification
```bash
# Verify systems registered
grep -c "AddSystem" cmd/client/handlers.go
# Expected: 109+

# Verify graphics baseline
grep -c "SelectTemplate64\|64x64" pkg/rendering/sprites/*.go
# Expected: Multiple matches

# Verify no feature flags for graphics
grep -E "sprite.*flag|tile.*flag|enable.*particle" cmd/client/*.go
# Expected: No matches
```

### Coverage Verification
```bash
# Overall coverage
go test ./pkg/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Expected: 65%+ (current: 82.4%)
```

---

## Component Integration Template

For future component additions, follow this pattern:

### 1. Add Component During Entity Creation

```go
// In cmd/client/handlers.go createPlayerEntity() or addPlayerComponents()
player.AddComponent(engine.NewMyComponent(params))
if verbose {
    clientLogger.WithFields(logrus.Fields{
        "component": "mycomponent",
    }).Debug("my component added")
}
```

### 2. Register System Unconditionally

```go
// In cmd/client/handlers.go initializeV*Systems()
sys.mySystem = engine.NewMySystem(game.World)
game.World.AddSystem(&mySystemWrapper{system: sys.mySystem})
```

### 3. Update All Entry Points

Files to update for player components:
- `cmd/client/handlers.go` — Desktop client
- `cmd/mobile/mobile.go` — Mobile client
- `cmd/server/main.go` — Server (if applicable)
- `pkg/engine/entity_spawning.go` — NPCs/enemies (if applicable)

### 4. Verify Integration

```bash
# Check component is added during creation
grep -rn "AddComponent.*NewMyComponent" cmd/client/ cmd/mobile/

# Check system is registered
grep -rn "AddSystem.*mySystem" cmd/client/

# Run tests
go test -v ./pkg/engine/... -run MySystem
```

---

## Anti-Patterns to Avoid

| ❌ Don't | ✅ Do |
|---------|------|
| `if enableFeatureX { ... }` | Unconditional `AddSystem()` |
| `-enable-*` flags | Direct imports without guards |
| `-sprite-size` flags | Hardcode `const SpriteSize = 64` |
| Lazy init in `Update()` | Add components during creation |
| `comp, _ := GetComponent()` | `comp, ok := GetComponent(); if ok && comp != nil` |
| Desktop-only integration | Update desktop, mobile, AND server |

---

## Common Failure Modes

### 1. Nil Component Panic
```
panic: interface conversion: engine.Component is nil, not *engine.FooComponent
```
**Fix**: Add defensive nil check and ensure component added during entity creation.

### 2. Stale ECS Query Cache
```
GetEntitiesWith("new_component") returns empty after AddComponent()
```
**Fix**: Add components during entity creation, not in system `Update()`.

### 3. Mobile/Desktop Parity
```
Feature works on desktop but crashes on mobile
```
**Fix**: Update both `cmd/client/handlers.go` AND `cmd/mobile/mobile.go`.

---

## Conclusion

The Venture codebase is fully integrated. All 102 packages are either:
- **Actively imported** (79 packages)
- **Indirectly used** (15 packages)
- **Testing infrastructure** (5 packages, intentionally separate)
- **Stub directories** (3 packages, architectural placeholders)

**No integration work is required.** This plan serves as documentation of the current state and a template for future additions.

---

## Document History

| Date | Author | Changes |
|------|--------|---------|
| December 2025 | Integration Audit | Initial comprehensive audit |
