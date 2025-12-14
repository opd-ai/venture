# Integration Audit - December 2025

## Summary

| Metric | Count | Percentage |
|--------|-------|------------|
| **Total pkg/ directories** | 103 | 100% |
| **Active (imported by client/server/mobile)** | 71 | 68.9% |
| **Dormant (requires integration)** | 32 | 31.1% |

**Note**: Many "dormant" packages are indirectly used via pkg/engine imports or are support/testing packages. True feature gaps are smaller than the raw count suggests.

---

## Active Packages

Packages currently imported by `cmd/client/`, `cmd/server/`, or `cmd/mobile/`.

| Package | LOC | Tests | Doc | Used By | Purpose |
|---------|-----|-------|-----|---------|---------|
| `pkg/engine` | 175,112 | 241 | ✅ | All | Core ECS framework, all game systems |
| `pkg/combat` | 507 | 1 | ✅ | Client, Server | Combat mechanics |
| `pkg/network` | 22,983 | 30 | ✅ | Client, Server | Client-server networking |
| `pkg/network/federation` | 8,916 | 12 | ✅ | Client, Server | Cross-server federation |
| `pkg/network/federation/guild` | 1,221 | 1 | ✅ | Client, Server | Guild federation protocol |
| `pkg/network/federation/mobile` | 1,390 | 2 | ✅ | Client | Mobile-specific federation |
| `pkg/network/chat` | 455 | 1 | ✅ | Client | Chat system |
| `pkg/network/trade` | 1,429 | 1 | ✅ | Client | Trade protocol |
| `pkg/procgen` | 509 | 2 | ✅ | All | Procedural generation core |
| `pkg/procgen/terrain` | 14,834 | 17 | ✅ | All | Terrain generation |
| `pkg/procgen/item` | 2,035 | 3 | ✅ | Client, Server | Item generation |
| `pkg/procgen/quest` | 1,505 | 1 | ✅ | Client | Quest generation |
| `pkg/procgen/magic` | 3,244 | 2 | ✅ | Client | Magic spell generation |
| `pkg/procgen/skills` | 1,983 | 1 | ✅ | Client | Skill tree generation |
| `pkg/procgen/building` | 1,932 | 1 | ✅ | Client, Server | Building generation |
| `pkg/procgen/furniture` | 2,391 | 2 | ✅ | Client, Server | Furniture generation |
| `pkg/procgen/companion` | 492 | 1 | ✅ | Client, Server | Companion generation |
| `pkg/procgen/environment` | 3,789 | 3 | ✅ | Client | Environment generation |
| `pkg/procgen/faction` | 999 | 1 | ✅ | Client | Faction generation |
| `pkg/procgen/genre` | 1,787 | 2 | ✅ | Client | Genre theming |
| `pkg/procgen/class` | 659 | 1 | ✅ | Client | Class generation |
| `pkg/procgen/book` | 2,430 | 1 | ✅ | Client | Book content generation |
| `pkg/procgen/puzzle` | 2,055 | 2 | ✅ | Client | Puzzle generation |
| `pkg/procgen/minigame` | 1,094 | 2 | ✅ | Client | Minigame framework |
| `pkg/procgen/minigame/games` | 1,985 | 3 | ✅ | Client | Minigame implementations |
| `pkg/procgen/narrative` | 1,100 | 1 | ✅ | Client | Narrative generation |
| `pkg/procgen/recipe` | 1,049 | 1 | ✅ | Client | Crafting recipe generation |
| `pkg/procgen/station` | 859 | 1 | ✅ | Client | Crafting station generation |
| `pkg/procgen/story` | 4,739 | 5 | ✅ | Client | Story arc generation |
| `pkg/procgen/vehicle` | 1,774 | 1 | ✅ | Client, Server | Vehicle generation |
| `pkg/rendering/sprites` | 15,479 | 15 | ✅ | Client | Sprite generation |
| `pkg/rendering/animation` | 2,260 | 4 | ✅ | Client | Animation system |
| `pkg/rendering/cache` | 2,189 | 4 | ✅ | Client | Sprite/animation caching |
| `pkg/rendering/display` | 785 | 3 | ✅ | Client | Display management |
| `pkg/rendering/lighting` | 2,786 | 3 | ✅ | Client | Dynamic lighting |
| `pkg/rendering/palette` | 4,349 | 4 | ✅ | Client | Color palette system |
| `pkg/rendering/parallel` | 1,129 | 2 | ✅ | Client | Parallel rendering |
| `pkg/rendering/particles` | 8,124 | 8 | ✅ | Client | Particle effects |
| `pkg/rendering/patterns` | 1,506 | 2 | ✅ | Client | Procedural patterns |
| `pkg/rendering/pool` | 591 | 1 | ✅ | Client | Object pooling |
| `pkg/rendering/postprocess` | 3,034 | 4 | ✅ | Client | Post-processing effects |
| `pkg/rendering/quality` | 1,951 | 3 | ✅ | Client | Quality settings |
| `pkg/rendering/shapes` | 2,265 | 2 | ✅ | Client | Shape rendering |
| `pkg/rendering/ui` | 11,957 | 13 | ✅ | Client | UI rendering |
| `pkg/audio` | 1,125 | 2 | ✅ | Client | Audio manager |
| `pkg/audio/music` | 2,820 | 3 | ✅ | Client | Music synthesis |
| `pkg/audio/sfx` | 1,428 | 3 | ✅ | Client | Sound effects |
| `pkg/class/advanced` | 3,198 | 2 | ✅ | Client | Advanced class system |
| `pkg/companion/learning` | 2,107 | 2 | ✅ | Client | Companion AI learning |
| `pkg/engine/physics/destruction` | 1,459 | 1 | ✅ | Client | Destruction physics |
| `pkg/engine/physics/fluids` | 1,907 | 3 | ✅ | Client, Server | Fluid simulation |
| `pkg/engine/physics/vehicle` | 2,506 | 5 | ❌ | Client, Server | Vehicle physics |
| `pkg/engine/prestige` | 1,869 | 2 | ✅ | Client | Prestige system |
| `pkg/engine/qol` | 1,799 | 3 | ✅ | Client | Quality of life features |
| `pkg/narrative/branching` | 2,033 | 2 | ✅ | Client | Branching narrative |
| `pkg/integration/companion_housing` | 1,137 | 2 | ❌ | Client | Companion-housing integration |
| `pkg/integration/guild_housing` | 1,358 | 1 | ✅ | Client | Guild-housing integration |
| `pkg/integration/housing_crafting` | 1,192 | 2 | ✅ | Client | Housing-crafting integration |
| `pkg/integration/narrative_world` | 2,449 | 2 | ✅ | Client | Narrative-world integration |
| `pkg/integration/political_warfare` | 1,553 | 2 | ✅ | Client | Political warfare |
| `pkg/integration/trade_routes` | 1,497 | 1 | ✅ | Client | Trade route system |
| `pkg/world` | 4,776 | 8 | ✅ | Client | World state |
| `pkg/world/housing` | 3,337 | 6 | ✅ | Client, Server | Housing system |
| `pkg/world/raids` | 3,043 | 4 | ✅ | Client | Raid system |
| `pkg/world/territory` | 2,242 | 3 | ✅ | Client | Territory control |
| `pkg/social/persistence` | 3,747 | 5 | ✅ | Client, Server | Social data persistence |
| `pkg/saveload` | 3,065 | 3 | ✅ | Client | Save/load system |
| `pkg/hostplay` | 2,497 | 5 | ✅ | Client | Host-and-play mode |
| `pkg/mobile` | 6,689 | 9 | ✅ | Mobile | Mobile platform support |
| `pkg/logging` | 545 | 1 | ✅ | All | Structured logging |
| `pkg/version` | 85 | 1 | ✅ | All | Version info |

---

## Dormant Packages (To Integrate)

Packages not directly imported by client/server/mobile. Classified by integration priority and effort.

### Priority 1: Ready for Immediate Integration

These packages are complete and have dependencies already active.

| Package | LOC | Tests | Completeness | Blocker | Integration Effort |
|---------|-----|-------|--------------|---------|-------------------|
| `pkg/rendering/tiles` | 6,602 | 6 | Complete | Not registered | **Small** - Add import |
| `pkg/audio/synthesis` | 698 | 1 | Complete | Not used directly | **Small** - Already used via audio/music |
| `pkg/integration/choice_consequences` | 1,545 | 1 | Complete | System not registered | **Small** - AddSystem() |
| `pkg/integration/guild_vehicle` | 1,433 | 2 | Complete | System not registered | **Small** - AddSystem() |
| `pkg/integration/territory_siege` | 1,925 | 4 | Complete | System not registered | **Small** - AddSystem() |
| `pkg/integration/world_events` | 1,876 | 2 | Complete | System not registered | **Small** - AddSystem() |
| `pkg/world/economy` | 3,002 | 4 | Complete | Not imported in cmd/ | **Small** - Add import |
| `pkg/network/resilience` | 1,334 | 1 | Complete | Not imported | **Small** - Network layer enhancement |
| `pkg/procgen/dialog` | 2,573 | 3 | Complete | Indirectly used | **Already Active** - via pkg/engine |
| `pkg/procgen/entity` | 2,161 | 2 | Complete | Indirectly used | **Already Active** - via pkg/engine |
| `pkg/procgen/legendary` | 3,078 | 2 | Complete | Indirectly used | **Already Active** - via legendary_quest_system |

### Priority 2: Support/Framework Packages

These provide infrastructure but don't need direct integration.

| Package | LOC | Tests | Purpose | Status |
|---------|-----|-------|---------|--------|
| `pkg/engine/performance` | 1,488 | 1 | Performance monitoring | Indirectly used via engine |
| `pkg/engine/physics` | 42 | 0 | Physics package root | Container only |
| `pkg/integration` | 451 | 1 | Integration package root | Container only |
| `pkg/rendering` | 445 | 1 | Rendering package root | Container only |
| `pkg/procgen/audit` | 1,878 | 3 | Generator validation | Development tool |

### Priority 3: Testing/Development Tools

These are not meant for production integration.

| Package | LOC | Tests | Purpose |
|---------|-----|-------|---------|
| `pkg/audit` | 0 | 0 | Empty package root |
| `pkg/audit/features` | 2,048 | 1 | Feature audit framework |
| `pkg/visualtest` | 5,176 | 7 | Visual regression testing |
| `pkg/visualtest/parity` | 1,340 | 3 | Cross-platform parity testing |

### Priority 4: Low Priority / Future Features

These require more work or have missing dependencies.

| Package | LOC | Tests | Completeness | Blocker | Integration Effort |
|---------|-----|-------|--------------|---------|-------------------|
| `pkg/balance` | 1,232 | 1 | Complete | No consumers | **Medium** - Needs system design |
| `pkg/class` | 0 | 0 | Stub | Container only | N/A |
| `pkg/companion` | 0 | 0 | Stub | Container only | N/A |
| `pkg/engine/saves` | 0 | 0 | Stub | Empty | N/A |
| `pkg/migration` | 619 | 1 | Complete | No consumers | **Medium** - Needs trigger |
| `pkg/modding` | 1,484 | 1 | Complete | Security concerns | **Large** - Sandbox needed |
| `pkg/narrative` | 0 | 0 | Stub | Container only | N/A |
| `pkg/network/federation/webrtc` | 4,246 | 5 | Complete | Browser-only | **Large** - WASM integration |
| `pkg/security` | 1,503 | 1 | Complete | No consumers | **Medium** - Needs hooks |
| `pkg/social` | 574 | 1 | Partial | No doc.go | **Medium** - Needs completion |
| `pkg/stability` | 599 | 1 | Complete | No consumers | **Small** - Add monitoring |
| `pkg/ux` | 1,571 | 1 | Complete | No consumers | **Medium** - UI integration |

---

## Graphics Baseline (Always Active)

All graphics enhancements are unconditionally enabled as of Phase 2.1:

| Enhancement | Default Value | Location | Status |
|-------------|---------------|----------|--------|
| **Sprite Size** | Variable (28-64px) | `cmd/client/consts.go` | ✅ Active |
| **Tile Size** | 32x32 | `cmd/client/handlers.go` | ✅ Active |
| **Particle System** | Always enabled | `pkg/rendering/particles/` | ✅ Active |
| **Dynamic Lighting** | Always enabled | `pkg/rendering/lighting/` | ✅ Active |
| **Shadow Rendering** | Always enabled | `pkg/engine/lighting_system.go` | ✅ Active |
| **Animation Cache** | Always enabled | `pkg/rendering/cache/` | ✅ Active |
| **Sprite Cache** | Always enabled | `pkg/rendering/cache/` | ✅ Active |
| **Post-Processing** | Always enabled | `pkg/rendering/postprocess/` | ✅ Active |
| **Color Grading** | Always enabled | `cmd/client/handlers.go:1346` | ✅ Active |
| **Vignette Effect** | Always enabled | `cmd/client/handlers.go:1357` | ✅ Active |
| **Chromatic Aberration** | Always enabled | `cmd/client/handlers.go:1369` | ✅ Active |

**Note**: No CLI flags exist for sprite size, tile size, or visual effects. Values are hardcoded constants per Phase 2.1 requirements.

---

## Registered Systems (178 Total)

All systems are unconditionally registered via `game.World.AddSystem()` in `cmd/client/handlers.go`.

### Core Systems
- `performanceSystem` - Performance monitoring
- `inputSystem` - Input handling
- `CameraSystem` - Camera control
- `rotationSystem` - Entity rotation
- `movementSystem` - Entity movement
- `collisionSystem` - Collision detection

### Combat Systems
- `playerCombatSystem` - Player combat actions
- `playerItemUseSystem` - Item usage
- `playerSpellCasting` - Spell casting
- `combatSystem` - Combat resolution
- `statusEffectSystem` - Status effects
- `projectileSystem` - Projectile management

### AI Systems
- `aiSystem` - Basic AI
- `behaviorTreeSystem` - Advanced AI behaviors
- `squadSystem` - Squad coordination

### Progression Systems
- `progressionSystem` - XP/leveling
- `prestigeSystem` - Prestige mechanics
- `skillProgressionSystem` - Skill advancement
- `classProgressionSys` - Class progression
- `advancedClassSystem` - Advanced classes

### Social Systems
- `factionSystem` - Faction management
- `reputationSystem` - Reputation tracking
- `alignmentSystem` - Moral alignment
- `factionReactionSystem` - Faction reactions
- `guildSystem` - Guild management
- `chatSystem` - In-game chat
- `networkChatSystem` - Network chat
- `networkTradeSystem` - Trading
- `mailSystem` - Mail system
- `courierSystem` - NPC courier

### World Systems
- `economySystem` - Dynamic economy
- `raidSystem` - Raid events
- `territorySystem` - Territory control
- `weatherSystem` - Weather effects
- `firePropagationSystem` - Fire spread
- `destructibleSystem` - Destructible objects
- `destructionSystem` - Destruction physics
- `fluidSimulator` - Fluid physics

### Content Systems
- `legendaryQuestSystem` - Legendary quests
- `dialogSystem` - NPC dialog
- `craftingSystem` - Crafting
- `puzzleSystem` - Puzzles
- `miniGameSystem` - Minigames
- `minigameGamesSystem` - Minigame implementations
- `bookReadingSystem` - Book reading

### Vehicle & Companion Systems
- `vehicleSystem` - Vehicle control
- `companionSystem` - Companion AI
- `companionProgressionSys` - Companion progression
- `companionLoyaltySys` - Companion loyalty
- `companionInventorySys` - Companion inventory
- `companionLearningSys` - Companion learning
- `companionLearningSystem` - Learning integration

### Audio Systems
- `audioManagerSystem` - Audio management
- `adaptiveSoundtrackSystem` - Dynamic music
- `musicTriggerSystem` - Music triggers
- `positionalAudioSystem` - 3D audio
- `reverbSystem` - Reverb effects

### Visual Systems
- `animationSystemWrapper` - Animation
- `equipmentVisualSystem` - Equipment rendering
- `particleSystem` - Particles
- `visualFeedbackSystem` - Combat feedback
- `qualitySystem` - Quality settings
- `lightingSystem` - Dynamic lighting

### Integration Systems
- `choiceConsequencesSystem` - Choice tracking
- `guildVehicleSystem` - Guild vehicles
- `narrativeWorldSystem` - Narrative-world
- `politicalWarfareSystem` - Political warfare
- `siegeSystem` - Territory sieges
- `mobileFederationSystem` - Mobile federation

---

## Dependency Graph

Integration order respects these dependency chains:

```
pkg/engine (core)
├── pkg/combat
├── pkg/network
│   ├── pkg/network/federation
│   │   ├── pkg/network/federation/guild
│   │   └── pkg/network/federation/mobile
│   ├── pkg/network/chat
│   ├── pkg/network/trade
│   └── pkg/network/resilience (dormant)
├── pkg/procgen
│   ├── pkg/procgen/terrain
│   ├── pkg/procgen/entity (indirect)
│   ├── pkg/procgen/item
│   ├── pkg/procgen/quest
│   ├── pkg/procgen/magic
│   ├── pkg/procgen/skills
│   ├── pkg/procgen/dialog (indirect)
│   ├── pkg/procgen/legendary (indirect)
│   └── ... (all procgen subpackages)
├── pkg/rendering
│   ├── pkg/rendering/sprites
│   ├── pkg/rendering/animation
│   ├── pkg/rendering/particles
│   ├── pkg/rendering/lighting
│   ├── pkg/rendering/cache
│   ├── pkg/rendering/tiles (dormant)
│   └── ... (all rendering subpackages)
├── pkg/world
│   ├── pkg/world/housing
│   ├── pkg/world/territory
│   ├── pkg/world/raids
│   └── pkg/world/economy (dormant)
└── pkg/integration
    ├── pkg/integration/narrative_world
    │   └── depends on: companion/learning, engine, procgen/story
    ├── pkg/integration/political_warfare
    │   └── depends on: engine, network/federation/guild
    ├── pkg/integration/territory_siege
    │   └── depends on: engine, network/federation/guild, world
    └── ... (integration packages have complex cross-dependencies)
```

---

## Integration Checklist

### Component Initialization Rules

**CRITICAL**: All components must be added during entity creation, NEVER during `System.Update()`.

```go
// ✅ CORRECT: Add during entity creation
func createPlayerEntity(game *engine.EbitenGame, ...) *engine.Entity {
    player := game.World.CreateEntity()
    player.AddComponent(&engine.PositionComponent{X: x, Y: y})
    player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
    player.AddComponent(engine.NewAdaptiveSoundtrackComponent(genreID))
    return player
}

// ❌ WRONG: Lazy initialization in Update()
func (s *MySystem) Update(deltaTime float64) {
    if !hasComponent {
        entity.AddComponent(NewMyComponent()) // DANGER: Cache staleness!
    }
}
```

### Defensive Component Access

Always check both the boolean return AND nil:

```go
// ✅ CORRECT
comp, ok := entity.GetComponent("my_component")
if ok && comp != nil {
    myComp := comp.(*MyComponent)
    // Safe to use
}

// ❌ WRONG
comp, _ := entity.GetComponent("my_component")
myComp := comp.(*MyComponent) // PANIC if nil
```

### Files to Update for Integration

| Component Type | Files |
|---------------|-------|
| Player components | `cmd/client/handlers.go`, `cmd/mobile/mobile.go` |
| System registration | `cmd/client/handlers.go` (initializeV*Systems) |
| System wrappers | `cmd/client/util.go` |
| Server components | `cmd/server/main.go` |
| Procgen components | `pkg/engine/entity_spawning.go` |

---

## Verification Commands

```bash
# Check component is added during entity creation
grep -rn "AddComponent.*New<ComponentName>" cmd/client/ cmd/mobile/ cmd/server/

# Check system is unconditionally registered
grep -rn "AddSystem.*New<SystemName>" cmd/client/ cmd/server/

# Check defensive nil checks
grep -A3 "GetComponent.*<component_type>" pkg/engine/<system_file>.go

# Run tests for integrated system
go test -v ./pkg/engine/... -run <SystemName>

# Build verification
go build ./cmd/client && go build ./cmd/server

# Test suite
go test ./pkg/...
```

---

## Common Integration Failure Modes

### 1. Nil Component Panic
**Symptom**: `panic: interface conversion: engine.Component is nil`
**Fix**: Add defensive nil check and ensure component added during entity creation.

### 2. Component Not Found After Adding
**Symptom**: `GetEntitiesWith()` returns empty after `AddComponent()`
**Fix**: Add components during entity creation, not in Update().

### 3. Mobile/Desktop Feature Parity
**Symptom**: Feature works on desktop but crashes on mobile
**Fix**: Integrate in both `cmd/client/handlers.go` AND `cmd/mobile/mobile.go`.

### 4. System Registration Order
**Symptom**: System A depends on data from System B but runs first
**Fix**: Register dependent systems AFTER their dependencies.

### 5. Missing Wrapper for System Interface
**Symptom**: System signature mismatch
**Fix**: Create wrapper struct implementing `Update(entities []*Entity, deltaTime float64)`.

---

*Generated: December 2025*
*Venture v10.0 - 95% Complete*
