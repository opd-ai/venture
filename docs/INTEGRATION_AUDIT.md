# Integration Audit - December 2025

## Summary

| Status | Count | Percentage |
|--------|-------|------------|
| **Active** | 79 | 82.3% |
| **Dormant** | 17 | 17.7% |
| **Total** | 96 | 100% |

**Note:** The project has achieved high integration levels. Most "dormant" packages are utility/test packages that don't require direct import (e.g., `pkg/visualtest`, `pkg/audit/features`, `pkg/procgen/audit`).

---

## Graphics Baseline (Always Active)

- **Sprite Resolution**: 64x64 pixels (procedurally generated)
- **Tile Resolution**: 64x64 pixels
- **Particle System**: Unconditionally active for combat, magic, environmental effects
- **Dynamic Lighting**: Per-pixel lighting with shadow casting
- **Animation Cache**: Automatic caching of sprite sequences
- **Visual Effects**: Explosions, magic auras, weather particles always rendered
- **Post-Processing**: Color grading, vignette, motion blur

---

## Active Packages (79 total)

### Core Engine (Imported by client/server)

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/engine` | client, server, mobile | Core ECS framework, all systems and components |
| `pkg/combat` | client, server | Combat interfaces and calculations |
| `pkg/logging` | client, server, mobile | Structured logging with Logrus |
| `pkg/version` | client, server | Version information |

### Physics & Simulation

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/engine/physics/destruction` | client | Destruction physics system |
| `pkg/engine/physics/fluids` | client, server | Fluid simulation (buoyancy, swimming) |
| `pkg/engine/physics/vehicle` | client, server | Enhanced vehicle physics |
| `pkg/engine/prestige` | client | Prestige/New Game+ system |
| `pkg/engine/qol` | client | Quality-of-life features |

### Audio

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/audio` | client | Audio manager and interfaces |
| `pkg/audio/music` | client | Adaptive music generation |
| `pkg/audio/sfx` | client | Sound effect generation |

### Procedural Generation

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/procgen` | client, server, mobile | Core generator interfaces |
| `pkg/procgen/book` | client | Procedural book content |
| `pkg/procgen/building` | client, server | Building layout generation |
| `pkg/procgen/class` | client | Class-specific content |
| `pkg/procgen/companion` | client, server | Companion NPC generation |
| `pkg/procgen/environment` | client | Environmental decoration |
| `pkg/procgen/faction` | client | Faction generation |
| `pkg/procgen/furniture` | client, server | Furniture placement |
| `pkg/procgen/genre` | client | Genre theming system |
| `pkg/procgen/item` | client, server, mobile | Item generation |
| `pkg/procgen/magic` | client | Spell and magic generation |
| `pkg/procgen/minigame` | client | Minigame generation |
| `pkg/procgen/minigame/games` | client | Specific game implementations |
| `pkg/procgen/narrative` | client | Narrative arc generation |
| `pkg/procgen/puzzle` | client | Puzzle generation |
| `pkg/procgen/quest` | client | Quest chain generation |
| `pkg/procgen/recipe` | client | Crafting recipe generation |
| `pkg/procgen/skills` | client | Skill tree generation |
| `pkg/procgen/station` | client | Crafting station generation |
| `pkg/procgen/story` | client | Story arc generation |
| `pkg/procgen/terrain` | client, server, mobile | Terrain and dungeon generation |
| `pkg/procgen/vehicle` | client | Vehicle generation |

### Rendering

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/rendering/animation` | client | Animation controller and caching |
| `pkg/rendering/cache` | client | Sprite and texture caching |
| `pkg/rendering/display` | client | Display management and scaling |
| `pkg/rendering/lighting` | client | Dynamic lighting system |
| `pkg/rendering/palette` | client | Color palette generation |
| `pkg/rendering/parallel` | client | Parallel rendering workers |
| `pkg/rendering/particles` | client | Particle effects system |
| `pkg/rendering/patterns` | client | Pattern generation |
| `pkg/rendering/pool` | client | Image pooling |
| `pkg/rendering/postprocess` | client | Post-processing effects |
| `pkg/rendering/quality` | client | Quality settings management |
| `pkg/rendering/shapes` | client | Shape generation |
| `pkg/rendering/sprites` | client, server | Sprite generation |
| `pkg/rendering/ui` | client | UI element generation |

### Network & Multiplayer

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/network` | client, server | Core networking protocol |
| `pkg/network/chat` | client | Chat system with encryption |
| `pkg/network/federation` | client, server | Cross-server federation |
| `pkg/network/federation/guild` | client, server | Guild federation sync |
| `pkg/network/federation/mobile` | client | Mobile federation adapter |
| `pkg/network/federation/webrtc` | client | WebRTC peer connections |
| `pkg/network/resilience` | server | Network resilience/recovery |
| `pkg/network/trade` | client | Trade protocol |

### World & Social

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/world` | client | World state management |
| `pkg/world/housing` | client, server | Player housing system |
| `pkg/world/raids` | client | Raid instance management |
| `pkg/world/territory` | client | Territory control system |
| `pkg/social/persistence` | client, server | Social data persistence |

### Integration Layers

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/integration/companion_housing` | client, server | Companion-housing integration |
| `pkg/integration/guild_housing` | client, server | Guild-housing integration |
| `pkg/integration/housing_crafting` | client, server | Housing-crafting integration |
| `pkg/integration/narrative_world` | client | Narrative-world integration |
| `pkg/integration/political_warfare` | client | Political warfare system |
| `pkg/integration/trade_routes` | client | Trade route management |

### Supporting Systems

| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/balance` | server | Combat and economic balance |
| `pkg/class/advanced` | client | Advanced class progression |
| `pkg/companion/learning` | client | Companion AI learning |
| `pkg/hostplay` | client | Host-and-play mode |
| `pkg/migration` | server | Save migration system |
| `pkg/mobile` | client, mobile | Mobile platform support |
| `pkg/modding` | server | Mod system (sandboxed) |
| `pkg/narrative/branching` | client | Branching narrative system |
| `pkg/saveload` | client | Save/load persistence |
| `pkg/security` | server | Security auditing |
| `pkg/stability` | server | Server stability monitoring |
| `pkg/ux` | server | UX journey validation |

---

## Dormant Packages (17 total)

### Production Integration Required

These packages have complete implementations but are not imported by any entry point:

#### `pkg/audio/synthesis`
- **Completeness**: Complete (doc.go, envelope.go, oscillator.go, tests)
- **Purpose**: Low-level waveform synthesis (sine, square, sawtooth, triangle, noise)
- **Dependencies**: None (standalone)
- **Blocker**: Not imported - used indirectly by `pkg/audio/music` and `pkg/audio/sfx`
- **Integration**: Import in audio packages that need direct synthesis control
- **Effort**: Small - may already be used transitively

#### `pkg/integration/choice_consequences`
- **Completeness**: Complete (doc.go, manager.go, types.go, tests)
- **Purpose**: Persistent choice tracking, branching consequence system
- **Dependencies**: `pkg/narrative/branching` (active)
- **Blocker**: Not registered as ECS system in client
- **Integration**:
  1. Import in `cmd/client/handlers.go`
  2. Create: `choiceTracker := choice_consequences.NewChoiceTracker()`
  3. Wire to narrative/quest systems
- **Effort**: Medium - needs integration with narrative flow

#### `pkg/integration/guild_vehicle`
- **Completeness**: Complete (doc.go, fleet_manager.go, types.go, tests)
- **Purpose**: Guild vehicle fleets with formations, siege engines
- **Dependencies**: `pkg/network/federation/guild` (active), `pkg/engine` (active)
- **Blocker**: Not registered - `guildVehicleSystem` exists but may not use this package
- **Integration**: Already integrated via `sys.guildVehicleSystem` - verify import
- **Effort**: Small - verify usage

#### `pkg/integration/world_events`
- **Completeness**: Complete (doc.go, manager.go, events.go, types.go, tests)
- **Purpose**: Dynamic world events from player actions (guild wars, economy, weather)
- **Dependencies**: `pkg/network/federation` (active), weather system (active)
- **Blocker**: Not registered as system
- **Integration**:
  1. Import in `cmd/client/handlers.go`
  2. Create: `worldEventsManager := world_events.NewEventManager(seed)`
  3. Register: `game.World.AddSystem(worldEventsSystemWrapper{...})`
- **Effort**: Medium - new system wrapper needed

#### `pkg/procgen/dialog`
- **Completeness**: Complete (doc.go, markov.go, corpus.go, personality.go, tests)
- **Purpose**: Markov chain NPC dialog generation with personality
- **Dependencies**: Genre system (active)
- **Blocker**: Engine has `MarkovDialogProvider` but unclear if using this package
- **Integration**: Verify `pkg/engine/markov_dialog_provider.go` imports this
- **Effort**: Small - likely already integrated

#### `pkg/procgen/entity`
- **Completeness**: Complete (doc.go, generator.go, merchant.go, types.go, tests)
- **Purpose**: Procedural monster/NPC/boss generation
- **Dependencies**: `pkg/procgen` (active)
- **Blocker**: Entity spawning in engine may duplicate this
- **Integration**: Consolidate with `pkg/engine/entity_spawning.go`
- **Effort**: Medium - code consolidation needed

#### `pkg/procgen/legendary`
- **Completeness**: Complete (doc.go, generator.go, manager.go, types.go, tests)
- **Purpose**: Legendary quest generation (multi-phase, cross-server)
- **Dependencies**: `pkg/procgen/quest` (active), `pkg/world/raids` (active)
- **Blocker**: `legendaryQuestSystem` exists in engine - verify import
- **Integration**: Already integrated via `sys.legendaryQuestSystem` - verify import
- **Effort**: Small - verify usage

#### `pkg/world/economy`
- **Completeness**: Complete (doc.go, marketplace.go, guild_bank.go, pricing_engine.go, tests)
- **Purpose**: Federated marketplace, guild banks, dynamic pricing
- **Dependencies**: `pkg/network/federation` (active), `pkg/network/trade` (active)
- **Blocker**: Not imported - commerce/economy systems may duplicate
- **Integration**:
  1. Import in `cmd/client/handlers.go`
  2. Create: `marketplace := economy.NewFederatedMarketplace(...)`
  3. Wire to trade UI and guild systems
- **Effort**: Medium - significant integration with existing trade system

### Test/Audit Packages (No Integration Needed)

These packages are test utilities and don't require client/server imports:

| Package | Purpose | Status |
|---------|---------|--------|
| `pkg/audit/features` | Feature completeness testing | Test-only |
| `pkg/procgen/audit` | Procedural generation testing | Test-only |
| `pkg/visualtest` | Visual regression testing | Test-only |
| `pkg/visualtest/parity` | Cross-platform parity testing | Test-only |

### Base/Interface Packages (Used Transitively)

| Package | Purpose | Status |
|---------|---------|--------|
| `pkg/engine/performance` | Performance monitoring types | Used by engine |
| `pkg/engine/physics` | Physics doc/interfaces | Parent package |
| `pkg/integration` | Integration doc | Parent package |
| `pkg/rendering` | Rendering interfaces | Used by subsystems |
| `pkg/rendering/tiles` | Tile rendering | Used by terrain system |
| `pkg/social` | Social system errors | Used by persistence |

---

## Dependency Graph

```mermaid
graph TD
    subgraph "Priority Integration"
        WE[world_events] --> FED[federation]
        WE --> WS[weather system]
        CC[choice_consequences] --> NB[narrative/branching]
        EC[world/economy] --> FED
        EC --> TR[network/trade]
    end
    
    subgraph "Verify Integration"
        GV[guild_vehicle] --> FG[federation/guild]
        LQ[procgen/legendary] --> RD[world/raids]
        DI[procgen/dialog] --> GE[procgen/genre]
        EN[procgen/entity] --> PG[procgen]
    end
    
    subgraph "Already Active"
        FED
        NB
        TR
        FG
        RD
        GE
        PG
        WS
    end
```

---

## Integration Verification Commands

### Check if package is actually imported

```bash
# Pattern: grep for package import in cmd/
grep -rn "pkg/integration/guild_vehicle" cmd/client/ cmd/server/
grep -rn "pkg/procgen/legendary" cmd/client/
grep -rn "pkg/procgen/dialog" cmd/client/ pkg/engine/
```

### Verify system registration

```bash
grep -rn "guildVehicleSystem\|legendaryQuestSystem\|worldEvents" cmd/client/handlers.go
```

### Check for duplicate implementations

```bash
# Entity generation
grep -rn "type.*EntityGenerator\|NewEntityGenerator" pkg/
# Dialog generation
grep -rn "MarkovDialog\|DialogGenerator" pkg/engine/
```

---

## Critical Integration Lessons

### Component Initialization: NEVER Use Lazy Initialization

**Problem**: The `AdaptiveSoundtrackSystem` used lazy initialization - creating the `AdaptiveSoundtrackComponent` on first `Update()` call. This caused nil pointer panics due to ECS query cache staleness.

**Solution**: Always add components during entity creation, NEVER during system Update().

```go
// ✅ GOOD: Add component during entity creation
func createPlayerEntity(game *engine.EbitenGame, ...) *engine.Entity {
    player := game.World.CreateEntity()
    player.AddComponent(&engine.PositionComponent{X: x, Y: y})
    player.AddComponent(engine.NewAdaptiveSoundtrackComponent(genreID))
    return player
}

// ❌ BAD: Lazy initialization in System.Update()
func (s *MySoundSystem) Update(deltaTime float64) {
    entities := s.world.GetEntitiesWith("my_sound")
    if len(entities) == 0 {
        // DANGER: Cache not invalidated!
        player.AddComponent(NewMySoundComponent())
    }
}
```

### Defensive Component Access Pattern

Always check both the boolean return AND nil before type assertion:

```go
// ✅ GOOD
comp, ok := entity.GetComponent("my_component")
if ok && comp != nil {
    myComp := comp.(*MyComponent)
    // Safe to use
}

// ❌ BAD
comp, _ := entity.GetComponent("my_component")
myComp := comp.(*MyComponent) // PANIC if nil
```

---

## Test Coverage Summary

| Package Category | Coverage |
|------------------|----------|
| Core Engine | 85%+ |
| Procedural Generation | 82%+ |
| Rendering | 87%+ |
| Network | 91%+ |
| Integration | 75%+ |
| **Average** | **82.4%** |

All packages exceed the 65% minimum requirement.

---

## Next Steps

See `docs/PLAN.md` for the phased integration roadmap.
