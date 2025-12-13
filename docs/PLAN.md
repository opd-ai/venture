# Integration Plan - December 2025

**Generated:** December 13, 2025  
**Purpose:** Phased roadmap to integrate all dormant packages as always-on baseline features

---

## Core Principle

**All features are baseline features.**

Dormant packages become **unconditionally enabled** upon integration:
- ❌ No feature flags
- ❌ No toggles
- ❌ No optional activation
- ❌ No `if enableFeatureX { ... }` conditionals
- ✅ Unconditional `AddSystem()` calls
- ✅ Direct imports without guards
- ✅ Always-on initialization

**If a package exists, it should be active.**

---

## Integration Timeline

| Phase | Duration | Packages | Focus |
|-------|----------|----------|-------|
| **Phase 1** | Week 1 (5 days) | 6 packages | Critical gameplay features |
| **Phase 2** | Week 2 (5 days) | 5 packages | Content & progression |
| **Phase 3** | Week 3 (5 days) | 5 packages | Polish & optimization |
| **Phase 4** | Week 4 (5 days) | 3 packages | Optional & advanced features |
| **Ongoing** | As needed | 9 packages | Development tools (non-runtime) |

**Total Runtime Integrations:** 19 packages  
**Total Development Tools:** 9 packages

---

## Phase 1: Critical Gameplay Features (Week 1)

**Goal:** Integrate systems that enhance core gameplay mechanics

### Day 1: Balance & Stability

#### 1.1 `pkg/balance` - Game Balance System
**Priority:** Critical  
**Complexity:** Low  
**Estimated Time:** 2 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeCoreSystems() after world creation

import "github.com/opd-ai/venture/pkg/balance"

func initializeCoreSystems(...) *systemsContainer {
    // ... existing systems ...
    
    // INTEGRATION: Balance system for dynamic difficulty scaling
    balanceSystem := balance.NewBalanceSystem(world)
    sys.balanceSystem = balanceSystem
    
    return sys
}

// File: cmd/client/handlers.go  
// Location: In registerAllSystems() after progression system

func registerAllSystems(...) {
    // ... existing system registrations ...
    
    game.World.AddSystem(sys.balanceSystem)
    
    // ... rest of systems ...
}

// File: cmd/client/util.go
// Location: Add to systemsContainer struct

type systemsContainer struct {
    // ... existing fields ...
    balanceSystem *balance.BalanceSystem
}
```

**Server Integration:**
```go
// File: cmd/server/main.go
// Location: In createGameWorld() after progression system

func createGameWorld(logger *logrus.Logger) *engine.World {
    // ... existing systems ...
    
    // INTEGRATION: Server-authoritative balance
    balanceSystem := balance.NewBalanceSystem(world)
    world.AddSystem(balanceSystem)
    
    return world
}
```

**Verification:**
```bash
go build ./cmd/client
go build ./cmd/server
go test ./pkg/balance/...
./client -verbose
# Check logs for "balance system initialized"
```

---

#### 1.2 `pkg/stability` - Crash Prevention
**Priority:** Critical  
**Complexity:** Low  
**Estimated Time:** 1.5 hours

**Client Integration:**
```go
// File: cmd/client/main.go
// Location: At the very start of main()

import "github.com/opd-ai/venture/pkg/stability"

func main() {
    // INTEGRATION: Stability monitor MUST be first
    stabilityMon := stability.NewMonitor()
    defer stabilityMon.RecoverFromPanic()
    
    flag.Parse()
    // ... rest of main ...
}

// File: cmd/client/handlers.go
// Location: In initializeCoreSystems()

func initializeCoreSystems(...) *systemsContainer {
    sys := &systemsContainer{}
    
    // INTEGRATION: Stability system for runtime monitoring
    sys.stabilityMonitor = stability.NewMonitor()
    
    // ... rest of initialization ...
    return sys
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems() - register FIRST

func registerAllSystems(...) {
    // INTEGRATION: Stability monitor must be first system
    game.World.AddSystem(sys.stabilityMonitor)
    
    game.World.AddSystem(sys.performanceSystem)
    // ... rest of systems ...
}

// File: cmd/client/util.go
type systemsContainer struct {
    stabilityMonitor *stability.Monitor
    // ... existing fields ...
}
```

**Server Integration:**
```go
// File: cmd/server/main.go
// Location: At start of main() and in createGameWorld()

import "github.com/opd-ai/venture/pkg/stability"

func main() {
    // INTEGRATION: Server stability monitoring
    stabilityMon := stability.NewMonitor()
    defer stabilityMon.RecoverFromPanic()
    
    flag.Parse()
    // ... rest of main ...
}

func createGameWorld(logger *logrus.Logger) *engine.World {
    world := engine.NewWorldWithLogger(logger)
    
    // INTEGRATION: Stability system for server
    stabilitySystem := stability.NewMonitor()
    world.AddSystem(stabilitySystem)
    
    // ... rest of systems ...
    return world
}
```

**Verification:**
```bash
go build ./cmd/client && go build ./cmd/server
# Test panic recovery
./client &
kill -SEGV $!  # Should recover gracefully
```

---

### Day 2: Network Resilience

#### 2.1 `pkg/network/resilience` - Network Recovery
**Priority:** High  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeNetworkClient()

import "github.com/opd-ai/venture/pkg/network/resilience"

func initializeNetworkClient(...) *network.Client {
    if !*multiplayer && !*hostAndPlay {
        return nil
    }
    
    // ... determine server address ...
    
    // INTEGRATION: Wrap network client with resilience layer
    baseClient := network.NewClient(serverAddr, logger)
    
    resilienceConfig := resilience.Config{
        MaxRetries:     5,
        RetryDelay:     time.Second,
        BackoffFactor:  2.0,
        MaxRetryDelay:  time.Second * 30,
        EnableLogging:  *verbose,
    }
    
    resilientClient := resilience.NewResilientClient(baseClient, resilienceConfig)
    
    if *verbose {
        clientLogger.Info("network resilience enabled")
    }
    
    return resilientClient
}
```

**Server Integration:**
```go
// File: cmd/server/main.go
// Location: In initializeNetworkSystems()

import "github.com/opd-ai/venture/pkg/network/resilience"

func initializeNetworkSystems(logger *logrus.Logger) (*network.Server, ...) {
    // ... existing server creation ...
    
    // INTEGRATION: Wrap server with resilience layer
    resilienceConfig := resilience.Config{
        MaxRetries:     3,
        RetryDelay:     time.Millisecond * 500,
        BackoffFactor:  1.5,
        MaxRetryDelay:  time.Second * 10,
        EnableLogging:  *verbose,
    }
    
    resilientServer := resilience.NewResilientServer(server, resilienceConfig)
    
    return resilientServer, snapshotManager, lagCompensator
}
```

**Verification:**
```bash
go build ./cmd/client && go build ./cmd/server
./server &
./client --server localhost:8080
# Simulate network disruption
sudo iptables -A OUTPUT -p tcp --dport 8080 -j DROP
sleep 5
sudo iptables -D OUTPUT -p tcp --dport 8080 -j DROP
# Client should reconnect automatically
```

---

### Day 3: Companion Learning & Entity Generation

#### 3.1 `pkg/companion/learning` - Companion AI Learning
**Priority:** High  
**Complexity:** Low  
**Estimated Time:** 2 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeV4Systems() after companion systems

import "github.com/opd-ai/venture/pkg/companion/learning"

func initializeV4Systems(...) {
    // ... existing companion systems ...
    
    // INTEGRATION: Companion learning system
    learningSystem := learning.NewLearningSystem(world)
    sys.companionLearningSystem = learningSystem
    
    if *verbose {
        clientLogger.Info("companion learning system initialized")
    }
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems() after companion systems

func registerAllSystems(...) {
    // ... existing companion registrations ...
    game.World.AddSystem(sys.companionLearningSystem)
    // ...
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    companionLearningSystem *learning.LearningSystem
}
```

**Server Integration:**
```go
// File: cmd/server/v4_systems.go
// Location: After companion systems initialization

import "github.com/opd-ai/venture/pkg/companion/learning"

func initializeV4Systems(...) {
    // ... existing companion systems ...
    
    // INTEGRATION: Server-authoritative companion learning
    learningSystem := learning.NewLearningSystem(world)
    world.AddSystem(learningSystem)
    
    serverLogger.Info("companion learning system initialized")
}
```

**Verification:**
```bash
go test ./pkg/companion/learning/...
go build ./cmd/client && go build ./cmd/server
./client -verbose
# Summon companion, observe learning behavior in logs
```

---

#### 3.2 `pkg/procgen/entity` - Entity Generation
**Priority:** High  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeGenerators()

import entitygen "github.com/opd-ai/venture/pkg/procgen/entity"

func initializeGenerators(sys *systemsContainer) {
    // ... existing generators ...
    
    // INTEGRATION: Entity generator for NPCs/enemies
    sys.entityGenerator = entitygen.NewEntityGenerator()
    
    logger.Info("entity generator initialized")
}

// File: cmd/client/util.go
// Location: In spawnWorldEntities() - REPLACE manual entity creation

import entitygen "github.com/opd-ai/venture/pkg/procgen/entity"

func spawnWorldEntities(...) {
    // OLD CODE (manual entity creation):
    // for i := 0; i < enemyCount; i++ { ... }
    
    // NEW CODE (use entity generator):
    params := procgen.GenerationParams{
        Seed:       *seed,
        Difficulty: 0.5,
        Depth:      0,
        GenreID:    *genreID,
    }
    
    for i := 0; i < enemyCount; i++ {
        entityData, err := sys.entityGenerator.Generate(
            *seed+int64(i),
            params,
        )
        if err != nil {
            clientLogger.WithError(err).Warn("failed to generate entity")
            continue
        }
        
        entity := createEntityFromGenerated(game, entityData, generatedTerrain)
        // ... rest of setup ...
    }
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    entityGenerator *entitygen.EntityGenerator
}
```

**Server Integration:**
```go
// File: cmd/server/entity_spawning.go
// Location: Replace manual spawning with generator

import entitygen "github.com/opd-ai/venture/pkg/procgen/entity"

func spawnV4Entities(...) {
    entityGen := entitygen.NewEntityGenerator()
    
    params := procgen.GenerationParams{
        Seed:       *seed,
        Difficulty: 0.6, // Server uses higher difficulty
        Depth:      0,
        GenreID:    *genreID,
    }
    
    for i := 0; i < 100; i++ {
        entityData, err := entityGen.Generate(*seed+int64(i), params)
        if err != nil {
            logger.WithError(err).Warn("entity generation failed")
            continue
        }
        
        entity := createAuthoritativeEntity(world, entityData, generatedTerrain)
        // ... setup ...
    }
}
```

**Verification:**
```bash
go test ./pkg/procgen/entity/...
go build ./cmd/client && go build ./cmd/server
./client -seed 99999 -verbose
# Verify diverse enemy types in logs/gameplay
```

---

### Day 4: Dialog Generation

#### 4.1 `pkg/procgen/dialog` - Dynamic Dialog
**Priority:** High  
**Complexity:** Low  
**Estimated Time:** 2.5 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeGenerators()

import dialoggen "github.com/opd-ai/venture/pkg/procgen/dialog"

func initializeGenerators(sys *systemsContainer) {
    // ... existing generators ...
    
    // INTEGRATION: Dialog generator for NPC conversations
    sys.dialogGenerator = dialoggen.NewDialogGenerator()
    
    logger.Info("dialog generator initialized")
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    dialogGenerator *dialoggen.DialogGenerator
}

// File: cmd/client/handlers.go
// Location: Pass generator to dialog system

func initializeV4Systems(...) {
    // ... dialog system setup ...
    
    // INTEGRATION: Connect dialog generator to dialog system
    sys.dialogSystem.SetGenerator(sys.dialogGenerator)
}
```

**Server Integration:**
```go
// File: cmd/server/v4_systems.go
// Location: In dialog system initialization

import dialoggen "github.com/opd-ai/venture/pkg/procgen/dialog"

func initializeV4Systems(...) {
    dialogGen := dialoggen.NewDialogGenerator()
    
    // Pass to NPC dialog system
    npcDialogSystem := /* ... existing creation ... */
    // Inject generator into system for runtime dialog generation
    
    serverLogger.Info("dialog generator integrated")
}
```

**Verification:**
```bash
go test ./pkg/procgen/dialog/...
go build ./cmd/client
./client -verbose
# Interact with NPCs, verify unique dialog
```

---

### Day 5: Performance Monitoring

#### 5.1 `pkg/engine/performance` - Performance Monitor
**Priority:** Medium  
**Complexity:** Low  
**Estimated Time:** 2 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeCoreSystems() early

import "github.com/opd-ai/venture/pkg/engine/performance"

func initializeCoreSystems(...) *systemsContainer {
    sys := &systemsContainer{}
    
    // INTEGRATION: Performance monitoring system
    perfConfig := performance.Config{
        SampleInterval: time.Second,
        LogThreshold:   30 * time.Millisecond, // Log if frame > 30ms
        EnableLogging:  *verbose,
    }
    sys.performanceMonitor = performance.NewMonitor(perfConfig)
    
    // ... rest of initialization ...
    return sys
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems() - early registration

func registerAllSystems(...) {
    game.World.AddSystem(sys.stabilityMonitor)
    game.World.AddSystem(sys.performanceMonitor)  // Right after stability
    // ... rest of systems ...
}

// File: cmd/client/util.go
type systemsContainer struct {
    performanceMonitor *performance.Monitor
    // ... existing fields ...
}
```

**Server Integration:**
```go
// File: cmd/server/main.go
// Location: In createGameWorld() early

import "github.com/opd-ai/venture/pkg/engine/performance"

func createGameWorld(logger *logrus.Logger) *engine.World {
    world := engine.NewWorldWithLogger(logger)
    
    // INTEGRATION: Server performance monitoring
    perfConfig := performance.Config{
        SampleInterval: time.Second * 5,
        LogThreshold:   50 * time.Millisecond, // Server tick budget
        EnableLogging:  *verbose,
    }
    perfMonitor := performance.NewMonitor(perfConfig)
    world.AddSystem(perfMonitor)
    
    // ... rest of systems ...
    return world
}
```

**Verification:**
```bash
go build ./cmd/client && go build ./cmd/server
./client -verbose &
./server -verbose &
# Check logs for performance metrics
# Spawn 2000 entities, verify FPS metrics logged
```

---

## Phase 1 Summary

**Completed Integrations:** 6 packages
- ✅ `pkg/balance` - Game balance
- ✅ `pkg/stability` - Crash prevention
- ✅ `pkg/network/resilience` - Network recovery
- ✅ `pkg/companion/learning` - Companion AI
- ✅ `pkg/procgen/entity` - Entity generation
- ✅ `pkg/procgen/dialog` - Dialog generation

**Verification Checklist:**
- [ ] All packages build successfully
- [ ] All tests pass: `go test ./pkg/...`
- [ ] Client runs without errors: `./client -verbose`
- [ ] Server runs without errors: `./server -verbose`
- [ ] No panics during 10-minute playthrough
- [ ] FPS ≥60 with 2000 entities
- [ ] Network reconnection works
- [ ] Companions exhibit learning behavior
- [ ] NPCs have unique dialog

---

## Phase 2: Content & Progression (Week 2)

**Goal:** Integrate systems that add content depth and player progression

### Day 6: Choice Consequences & World Events

#### 6.1 `pkg/integration/choice_consequences` - Narrative Depth
**Priority:** High  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeV9Systems() or new narrative section

import "github.com/opd-ai/venture/pkg/integration/choice_consequences"

func initializeV9Systems(...) {
    // ... existing integrations ...
    
    // INTEGRATION: Choice consequence manager for narrative depth
    choiceConfig := choice_consequences.Config{
        TrackingDepth:     10,   // Remember last 10 choices
        ConsequenceDelay:  5,    // Consequences appear after 5 game events
        BranchingFactor:   3,    // Up to 3 consequence branches per choice
    }
    
    choiceManager := choice_consequences.NewManager(world, choiceConfig)
    sys.choiceConsequenceManager = choiceManager
    
    // Connect to narrative system
    sys.narrativeSystem.SetConsequenceManager(choiceManager)
    
    clientLogger.Info("choice consequence manager initialized")
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after narrative systems ...
    game.World.AddSystem(sys.choiceConsequenceManager)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    choiceConsequenceManager *choice_consequences.Manager
}
```

**Server Integration:**
```go
// File: cmd/server/v9_systems.go
// Location: Create if doesn't exist, or add to existing

import "github.com/opd-ai/venture/pkg/integration/choice_consequences"

func initializeV9Systems(world *engine.World, logger *logrus.Logger) {
    // INTEGRATION: Server-authoritative choice tracking
    choiceConfig := choice_consequences.Config{
        TrackingDepth:     10,
        ConsequenceDelay:  5,
        BranchingFactor:   3,
    }
    
    choiceManager := choice_consequences.NewManager(world, choiceConfig)
    world.AddSystem(choiceManager)
    
    logger.Info("choice consequence manager initialized")
}

// File: cmd/server/main.go
// Location: In createGameWorld() after V4 systems

func createGameWorld(logger *logrus.Logger) *engine.World {
    // ... existing systems ...
    
    // INTEGRATION: V9 systems
    initializeV9Systems(world, logger)
    
    return world
}
```

**Verification:**
```bash
go test ./pkg/integration/choice_consequences/...
go build ./cmd/client && go build ./cmd/server
./client -verbose
# Make dialog choices, verify consequences appear later
```

---

#### 6.2 `pkg/integration/world_events` - Dynamic World
**Priority:** High  
**Complexity:** Low  
**Estimated Time:** 2 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: After procgen systems

import "github.com/opd-ai/venture/pkg/integration/world_events"

func setupWorldTerrain(...) {
    // ... after terrain generation ...
    
    // INTEGRATION: World event manager for dynamic events
    eventConfig := world_events.Config{
        EventFrequency:   time.Minute * 5,  // Event every 5 minutes
        MaxActiveEvents:  3,                // Max 3 concurrent events
        UseGenreSeed:     true,
    }
    
    eventManager := world_events.NewManager(world, *seed, eventConfig)
    sys.worldEventManager = eventManager
    
    clientLogger.Info("world event manager initialized")
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after procgen systems ...
    game.World.AddSystem(sys.worldEventManager)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    worldEventManager *world_events.Manager
}
```

**Server Integration:**
```go
// File: cmd/server/main.go
// Location: After terrain generation

import "github.com/opd-ai/venture/pkg/integration/world_events"

func generateWorldTerrain(...) *terrain.GeneratedTerrain {
    // ... existing terrain generation ...
    
    // INTEGRATION: Server-authoritative world events
    eventConfig := world_events.Config{
        EventFrequency:   time.Minute * 3,  // Server spawns events faster
        MaxActiveEvents:  5,
        UseGenreSeed:     true,
    }
    
    eventManager := world_events.NewManager(world, *seed, eventConfig)
    world.AddSystem(eventManager)
    
    serverLogger.Info("world event manager initialized")
    
    return generatedTerrain
}
```

**Verification:**
```bash
go test ./pkg/integration/world_events/...
go build ./cmd/client && go build ./cmd/server
./client -verbose
# Wait 5 minutes, verify dynamic events spawn
```

---

### Day 7: Legendary Content

#### 7.1 `pkg/procgen/legendary` - Legendary Items/Quests
**Priority:** Medium  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeGenerators()

import legendarygen "github.com/opd-ai/venture/pkg/procgen/legendary"

func initializeGenerators(sys *systemsContainer) {
    // ... existing generators ...
    
    // INTEGRATION: Legendary content generator
    legendaryConfig := legendarygen.Config{
        DropRate:      0.01,   // 1% legendary drop rate
        MinPlayerLevel: 20,     // Legendaries at level 20+
        UseRaidBosses:  true,
    }
    
    sys.legendaryGenerator = legendarygen.NewGenerator(legendaryConfig)
    
    logger.Info("legendary generator initialized")
}

// File: cmd/client/handlers.go
// Location: Connect to legendary quest system (already active)

func initializeV4Systems(...) {
    // ... legendary quest system exists ...
    
    // INTEGRATION: Connect generator to quest system
    sys.legendaryQuestSystem.SetGenerator(sys.legendaryGenerator)
    
    clientLogger.Info("legendary generator connected to quest system")
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    legendaryGenerator *legendarygen.Generator
}
```

**Server Integration:**
```go
// File: cmd/server/v4_systems.go
// Location: After raid system initialization

import legendarygen "github.com/opd-ai/venture/pkg/procgen/legendary"

func initializeV4Systems(...) {
    // ... existing raid system ...
    
    // INTEGRATION: Legendary generator for raid boss loot
    legendaryConfig := legendarygen.Config{
        DropRate:      0.02,   // 2% on server (authoritative)
        MinPlayerLevel: 20,
        UseRaidBosses:  true,
    }
    
    legendaryGen := legendarygen.NewGenerator(legendaryConfig)
    
    // Connect to raid system for boss loot
    raidSystem.SetLegendaryGenerator(legendaryGen)
    
    serverLogger.Info("legendary generator integrated with raids")
}
```

**Verification:**
```bash
go test ./pkg/procgen/legendary/...
go build ./cmd/client && go build ./cmd/server
./client -verbose
# Defeat raid boss at level 20+, check for legendary loot
```

---

### Day 8: Territory Siege

#### 8.1 `pkg/integration/territory_siege` - Territory Warfare
**Priority:** High  
**Complexity:** Medium  
**Estimated Time:** 3.5 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeV9Systems()

import "github.com/opd-ai/venture/pkg/integration/territory_siege"

func initializeV9Systems(...) {
    // ... existing integrations ...
    
    // INTEGRATION: Territory siege manager for PvP warfare
    siegeConfig := territory_siege.Config{
        MinAttackers:      5,               // Need 5 players to start siege
        SiegeDuration:     time.Minute * 30, // 30-minute sieges
        DefenseBonus:      1.5,             // Defenders get 50% bonus
        RequireGuildWar:   true,            // Must be guild war
    }
    
    siegeManager := territory_siege.NewManager(
        world,
        sys.territorySystem,
        sys.guildSystem,
        siegeConfig,
    )
    sys.siegeManager = siegeManager
    
    clientLogger.Info("territory siege manager initialized")
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after territory and guild systems ...
    game.World.AddSystem(sys.siegeManager)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    siegeManager *territory_siege.Manager
}
```

**Server Integration:**
```go
// File: cmd/server/v9_systems.go
// Location: After territory and guild systems

import "github.com/opd-ai/venture/pkg/integration/territory_siege"

func initializeV9Systems(world *engine.World, logger *logrus.Logger) {
    // ... existing systems ...
    
    // INTEGRATION: Server-authoritative siege management
    siegeConfig := territory_siege.Config{
        MinAttackers:      5,
        SiegeDuration:     time.Minute * 30,
        DefenseBonus:      1.5,
        RequireGuildWar:   true,
    }
    
    siegeManager := territory_siege.NewManager(
        world,
        territorySystem,
        guildSystem,
        siegeConfig,
    )
    world.AddSystem(siegeManager)
    
    logger.Info("territory siege manager initialized")
}
```

**Verification:**
```bash
go test ./pkg/integration/territory_siege/...
go build ./cmd/client && go build ./cmd/server
./server &
# Start 2 clients in different guilds
# Initiate guild war, attack territory
```

---

### Day 9-10: Trade Routes

#### 9.1 `pkg/integration/trade_routes` - Economic Routes
**Priority:** Medium  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: After economy system

import "github.com/opd-ai/venture/pkg/integration/trade_routes"

func initializeProgressionSystems(...) {
    // ... after economy system ...
    
    // INTEGRATION: Trade route manager for dynamic economy
    routeConfig := trade_routes.Config{
        MaxRoutes:         10,              // 10 active trade routes
        RouteLifetime:     time.Hour,       // Routes last 1 hour
        CaravanSpeed:      50.0,            // pixels per second
        PirateChance:      0.1,             // 10% pirate encounter
    }
    
    tradeRouteManager := trade_routes.NewManager(
        world,
        sys.economySystem,
        routeConfig,
    )
    sys.tradeRouteManager = tradeRouteManager
    
    clientLogger.Info("trade route manager initialized")
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after economy system ...
    game.World.AddSystem(sys.tradeRouteManager)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    tradeRouteManager *trade_routes.Manager
}
```

**Server Integration:**
```go
// File: cmd/server/v9_systems.go
// Location: After economy initialization

import "github.com/opd-ai/venture/pkg/integration/trade_routes"

func initializeV9Systems(world *engine.World, logger *logrus.Logger) {
    // ... existing systems ...
    
    // INTEGRATION: Server-authoritative trade routes
    routeConfig := trade_routes.Config{
        MaxRoutes:         20,              // More routes on server
        RouteLifetime:     time.Hour,
        CaravanSpeed:      50.0,
        PirateChance:      0.15,            // Higher danger server-side
    }
    
    tradeRouteManager := trade_routes.NewManager(
        world,
        economySystem,
        routeConfig,
    )
    world.AddSystem(tradeRouteManager)
    
    logger.Info("trade route manager initialized")
}
```

**Verification:**
```bash
go test ./pkg/integration/trade_routes/...
go build ./cmd/client && go build ./cmd/server
./client -verbose
# Check map for trade routes, caravans moving
# Interact with caravan, verify trading
```

---

## Phase 2 Summary

**Completed Integrations:** 5 packages
- ✅ `pkg/integration/choice_consequences` - Narrative consequences
- ✅ `pkg/integration/world_events` - Dynamic events
- ✅ `pkg/procgen/legendary` - Legendary content
- ✅ `pkg/integration/territory_siege` - Territory warfare
- ✅ `pkg/integration/trade_routes` - Trade routes

**Verification Checklist:**
- [ ] All Phase 2 packages build
- [ ] Tests pass: `go test ./pkg/integration/... ./pkg/procgen/legendary/...`
- [ ] Choice consequences trigger after dialog
- [ ] World events spawn every 5 minutes
- [ ] Legendary items drop from raid bosses
- [ ] Territory sieges can be initiated
- [ ] Trade caravans move on map

---

## Phase 3: Polish & Optimization (Week 3)

**Goal:** Enhance existing features with polish and optimization

### Day 11: UX Improvements

#### 11.1 `pkg/ux` - User Experience System
**Priority:** Medium  
**Complexity:** Low  
**Estimated Time:** 2 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: After UI initialization

import "github.com/opd-ai/venture/pkg/ux"

func setupGameUI(...) {
    // ... existing UI setup ...
    
    // INTEGRATION: UX system for usability improvements
    uxConfig := ux.Config{
        EnableTooltips:      true,
        EnableTutorials:     true,
        EnableAccessibility: true,
        TooltipDelay:        time.Millisecond * 500,
    }
    
    uxSystem := ux.NewUXSystem(game.InputSystem, uxConfig)
    sys.uxSystem = uxSystem
    
    clientLogger.Info("UX system initialized")
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after UI systems ...
    game.World.AddSystem(sys.uxSystem)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    uxSystem *ux.UXSystem
}
```

**Verification:**
```bash
go test ./pkg/ux/...
go build ./cmd/client
./client -verbose
# Hover over UI elements, verify tooltips appear
```

---

### Day 12: Migration & Save Compatibility

#### 12.1 `pkg/migration` - Save File Migration
**Priority:** High  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: Before saveload operations

import "github.com/opd-ai/venture/pkg/migration"

func initializeCoreSystems(...) *systemsContainer {
    // ... early in initialization ...
    
    // INTEGRATION: Migration manager for save compatibility
    migrationConfig := migration.Config{
        CurrentVersion: "1.0.0",
        EnableLogging:  *verbose,
    }
    
    sys.migrationManager = migration.NewManager(migrationConfig)
    
    return sys
}

// File: pkg/saveload/manager.go
// Location: In Load() method - integrate with existing saveload

import "github.com/opd-ai/venture/pkg/migration"

func (m *Manager) Load(filename string) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }
    
    // INTEGRATION: Apply migrations before loading
    migrated, err := m.migrationManager.Migrate(data)
    if err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    
    // Load migrated data
    return m.deserialize(migrated)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    migrationManager *migration.Manager
}
```

**Verification:**
```bash
go test ./pkg/migration/...
go build ./cmd/client
./client -verbose
# Load old save file, verify migration succeeds
```

---

### Day 13: Rendering Optimization

#### 13.1 `pkg/rendering/tiles` - Tile-Based Rendering
**Priority:** Medium  
**Complexity:** Medium  
**Estimated Time:** 3.5 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In rendering initialization

import "github.com/opd-ai/venture/pkg/rendering/tiles"

func initializeCoreSystems(...) *systemsContainer {
    // ... after render system ...
    
    // INTEGRATION: Tile rendering system (parallel to sprites)
    tileConfig := tiles.Config{
        TileSize:         32,
        EnableCaching:    true,
        CacheSize:        1000,
        UseFastPath:      true,
    }
    
    tileRenderer := tiles.NewRenderer(game.RenderSystem, tileConfig)
    sys.tileRenderer = tileRenderer
    
    clientLogger.Info("tile renderer initialized")
    
    return sys
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after render system ...
    game.World.AddSystem(sys.tileRenderer)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    tileRenderer *tiles.Renderer
}
```

**Verification:**
```bash
go test ./pkg/rendering/tiles/...
go build ./cmd/client
./client -verbose
# Verify rendering performance with 2000 entities
# FPS should remain ≥60
```

---

### Day 14-15: Guild Vehicles

#### 14.1 `pkg/integration/guild_vehicle` - Guild Vehicle Integration
**Priority:** Low  
**Complexity:** Medium  
**Estimated Time:** 2.5 hours

**Client Integration:**
```go
// File: cmd/client/handlers.go
// Location: In initializeV9Systems()

import "github.com/opd-ai/venture/pkg/integration/guild_vehicle"

func initializeV9Systems(...) {
    // ... existing integrations ...
    
    // INTEGRATION: Guild vehicle manager
    guildVehicleConfig := guild_vehicle.Config{
        MaxGuildVehicles: 5,              // 5 vehicles per guild
        SharedControl:    true,           // Guild members share control
        RequireRank:      "Officer",      // Officers can spawn vehicles
    }
    
    guildVehicleManager := guild_vehicle.NewManager(
        world,
        sys.guildSystem,
        sys.vehicleSystem,
        guildVehicleConfig,
    )
    sys.guildVehicleManager = guildVehicleManager
    
    clientLogger.Info("guild vehicle manager initialized")
}

// File: cmd/client/handlers.go
// Location: In registerAllSystems()

func registerAllSystems(...) {
    // ... after guild and vehicle systems ...
    game.World.AddSystem(sys.guildVehicleManager)
}

// File: cmd/client/util.go
type systemsContainer struct {
    // ... existing fields ...
    guildVehicleManager *guild_vehicle.Manager
}
```

**Server Integration:**
```go
// File: cmd/server/v9_systems.go
// Location: After guild and vehicle systems

import "github.com/opd-ai/venture/pkg/integration/guild_vehicle"

func initializeV9Systems(world *engine.World, logger *logrus.Logger) {
    // ... existing systems ...
    
    // INTEGRATION: Server-authoritative guild vehicles
    guildVehicleConfig := guild_vehicle.Config{
        MaxGuildVehicles: 5,
        SharedControl:    true,
        RequireRank:      "Officer",
    }
    
    guildVehicleManager := guild_vehicle.NewManager(
        world,
        guildSystem,
        vehicleSystem,
        guildVehicleConfig,
    )
    world.AddSystem(guildVehicleManager)
    
    logger.Info("guild vehicle manager initialized")
}
```

**Verification:**
```bash
go test ./pkg/integration/guild_vehicle/...
go build ./cmd/client && go build ./cmd/server
./server &
# Join guild as officer, spawn guild vehicle
# Other guild members can control it
```

---

## Phase 3 Summary

**Completed Integrations:** 5 packages
- ✅ `pkg/ux` - User experience improvements
- ✅ `pkg/migration` - Save file compatibility
- ✅ `pkg/rendering/tiles` - Tile rendering
- ✅ `pkg/engine/performance` - Performance monitoring (from Phase 1)
- ✅ `pkg/integration/guild_vehicle` - Guild vehicles

**Verification Checklist:**
- [ ] Tooltips appear on UI hover
- [ ] Old save files load successfully
- [ ] Tile rendering works alongside sprites
- [ ] Performance metrics logged
- [ ] Guild vehicles spawn and control works

---

## Phase 4: Optional & Advanced Features (Week 4)

**Goal:** Integrate optional features and advanced systems

### Day 16-17: Security & Anti-Cheat

#### 16.1 `pkg/security` - Anti-Cheat System
**Priority:** Medium  
**Complexity:** Medium  
**Estimated Time:** 4 hours

**Server Integration (Server-Side Only):**
```go
// File: cmd/server/main.go
// Location: Early in createGameWorld()

import "github.com/opd-ai/venture/pkg/security"

func createGameWorld(logger *logrus.Logger) *engine.World {
    world := engine.NewWorldWithLogger(logger)
    
    // INTEGRATION: Security system for cheat detection
    securityConfig := security.Config{
        EnableValidation:     true,
        EnableRateLimiting:   true,
        MaxActionsPerSecond:  10,
        BanThreshold:         5,    // 5 violations = ban
        LogViolations:        true,
    }
    
    securitySys := security.NewSecuritySystem(world, securityConfig)
    world.AddSystem(securitySys)
    
    logger.Info("security system initialized")
    
    // ... rest of systems ...
    return world
}

// File: cmd/server/main.go
// Location: In player action handlers

func handlePlayerAction(player *engine.Entity, action Action) {
    // INTEGRATION: Validate action before processing
    if !securitySys.ValidateAction(player, action) {
        logger.WithFields(logrus.Fields{
            "playerID": player.ID,
            "action":   action.Type,
        }).Warn("suspicious action blocked")
        return
    }
    
    // Process valid action
    // ...
}
```

**Verification:**
```bash
go test ./pkg/security/...
go build ./cmd/server
./server -verbose &
# Attempt rapid actions from client, verify rate limiting
```

---

### Day 18-19: WebRTC P2P Networking

#### 18.1 `pkg/network/federation/webrtc` - WebRTC Transport
**Priority:** Low  
**Complexity:** High  
**Estimated Time:** 5 hours

**Client Integration (Optional, Flag-Gated):**
```go
// File: cmd/client/main.go
// Location: Add flag

var (
    // ... existing flags ...
    useWebRTC = flag.Bool("webrtc", false, "Use WebRTC for P2P networking (EXPERIMENTAL)")
)

// File: cmd/client/handlers.go
// Location: In initializeNetworkClient()

import "github.com/opd-ai/venture/pkg/network/federation/webrtc"

func initializeNetworkClient(...) *network.Client {
    if !*multiplayer && !*hostAndPlay {
        return nil
    }
    
    var transport network.Transport
    
    if *useWebRTC {
        // INTEGRATION: WebRTC transport for P2P
        webrtcConfig := webrtc.Config{
            StunServers: []string{
                "stun:stun.l.google.com:19302",
            },
            EnableLogging: *verbose,
        }
        
        transport = webrtc.NewTransport(webrtcConfig)
        clientLogger.Info("using WebRTC P2P transport")
    } else {
        // Default TCP/UDP transport
        transport = network.NewDefaultTransport()
    }
    
    return network.NewClientWithTransport(serverAddr, logger, transport)
}

// File: cmd/client/main.go (WASM-specific)
// Location: In WASM build, WebRTC is the ONLY option

if mobile.IsWASM() {
    if *multiplayer {
        // INTEGRATION: WASM requires WebRTC for P2P
        webrtcConfig := webrtc.Config{
            StunServers: []string{
                "stun:stun.l.google.com:19302",
            },
            EnableLogging: *verbose,
        }
        transport := webrtc.NewTransport(webrtcConfig)
        networkClient = network.NewClientWithTransport(serverAddr, logger, transport)
        
        clientLogger.Info("WASM build - using WebRTC for multiplayer")
    }
}
```

**Verification:**
```bash
go test ./pkg/network/federation/webrtc/...
go build ./cmd/client
./client -webrtc -server peer://abc123 -verbose
# Verify P2P connection established
```

---

### Day 20: Modding Support

#### 20.1 `pkg/modding` - Mod Loader (Optional)
**Priority:** Low  
**Complexity:** Medium  
**Estimated Time:** 3 hours

**Client Integration (Optional, Flag-Gated):**
```go
// File: cmd/client/main.go
// Location: Add flags

var (
    // ... existing flags ...
    enableMods = flag.Bool("mods", false, "Enable mod loading (EXPERIMENTAL)")
    modPath    = flag.String("mod-path", "./mods", "Path to mods directory")
)

// File: cmd/client/main.go
// Location: After game initialization, before system setup

import "github.com/opd-ai/venture/pkg/modding"

func main() {
    // ... flag parsing ...
    
    game := createGameInstance(logger, clientLogger)
    
    if *enableMods {
        // INTEGRATION: Mod loader (EXPERIMENTAL)
        modConfig := modding.Config{
            ModPath:         *modPath,
            EnableSandbox:   true,   // Sandbox mods for safety
            AllowScripting:  false,  // Disable scripting for now
            EnableLogging:   *verbose,
        }
        
        modLoader := modding.NewLoader(modConfig)
        if err := modLoader.LoadMods(); err != nil {
            clientLogger.WithError(err).Warn("failed to load mods")
        } else {
            game.World.AddSystem(modLoader)
            clientLogger.WithField("count", modLoader.Count()).Info("mods loaded")
        }
    }
    
    // ... rest of initialization ...
}
```

**Verification:**
```bash
go test ./pkg/modding/...
go build ./cmd/client
mkdir -p ./mods
# Create test mod
./client -mods -mod-path ./mods -verbose
```

---

## Phase 4 Summary

**Completed Integrations:** 3 packages
- ✅ `pkg/security` - Anti-cheat system (server-only)
- ✅ `pkg/network/federation/webrtc` - WebRTC P2P (optional)
- ✅ `pkg/modding` - Mod loader (optional)

**Verification Checklist:**
- [ ] Security system blocks suspicious actions
- [ ] WebRTC P2P connection works
- [ ] Mods load successfully

---

## Development Tools (Non-Runtime Integration)

These packages are **development and testing tools**, not runtime features. They don't require client/server integration but should be available for development workflows.

### Available Development Tools (9 packages)

1. **`pkg/audit/features`** - Feature detection and auditing
   - **Usage:** `go run ./examples/featureaudit`
   - **Purpose:** Verify all features are implemented and working

2. **`pkg/procgen/audit`** - Generator validation
   - **Usage:** Add `-audit-generators` flag to client
   - **Purpose:** Validate all procedural generators

3. **`pkg/visualtest`** - Visual regression testing
   - **Usage:** `go run ./examples/visualtest`
   - **Purpose:** CI/CD visual regression tests

4. **`pkg/visualtest/parity`** - Cross-platform parity
   - **Usage:** `go run ./examples/visualtest/parity`
   - **Purpose:** Verify visual parity across platforms

5. **`pkg/audio/synthesis`** - Low-level audio synthesis
   - **Usage:** Already integrated via `audio/music` and `audio/sfx`
   - **Purpose:** Base synthesis routines

6. **`pkg/rendering`** - Base rendering types
   - **Usage:** Already integrated via `rendering/*` subpackages
   - **Purpose:** Common rendering interfaces

7. **`pkg/integration`** - Base integration types
   - **Usage:** Already integrated via `integration/*` subpackages
   - **Purpose:** Common integration interfaces

8. **`pkg/social`** - Base social types
   - **Usage:** Already integrated via `social/persistence`
   - **Purpose:** Social graph types (needs doc.go)

9. **`pkg/engine/physics`** - Base physics types
   - **Usage:** Already integrated via `physics/*` subpackages
   - **Purpose:** Physics interfaces (stub, needs expansion)

### Integration into CI/CD

```yaml
# File: .github/workflows/test.yml
# Add development tool validation

- name: Run Feature Audit
  run: go run ./examples/featureaudit

- name: Run Visual Tests
  run: go run ./examples/visualtest

- name: Run Platform Parity
  run: go run ./examples/visualtest/parity
```

---

## Special Cases & Considerations

### Duplicate Implementation: `pkg/world/economy`

**Issue:** `pkg/world/economy` duplicates functionality in `pkg/engine/economy_system.go`

**Investigation Needed:**
1. Compare implementations: `diff pkg/world/economy/ pkg/engine/economy_system.go`
2. Determine canonical implementation
3. Choose migration path:
   - **Option A:** Migrate to `pkg/world/economy` (cleaner separation of concerns)
   - **Option B:** Keep `pkg/engine/economy_system.go`, deprecate `pkg/world/economy`
   - **Option C:** Merge implementations, keep best features

**Recommendation:** Defer to separate refactoring task. Current implementation works.

---

## Anti-Patterns to Avoid

### ❌ Don't Do This:

```go
// WRONG: Feature flag
if *enableBalance {
    balanceSystem := balance.NewBalanceSystem(world)
    game.World.AddSystem(balanceSystem)
}

// WRONG: Optional toggle
if config.Features.CompanionLearning {
    learningSystem := learning.NewLearningSystem(world)
}

// WRONG: Conditional import
var dialogGen *dialog.Generator
if *useDialogGen {
    dialogGen = dialog.NewGenerator()
}
```

### ✅ Do This Instead:

```go
// CORRECT: Unconditional integration
balanceSystem := balance.NewBalanceSystem(world)
game.World.AddSystem(balanceSystem)

// CORRECT: Always active
learningSystem := learning.NewLearningSystem(world)
game.World.AddSystem(learningSystem)

// CORRECT: Direct integration
dialogGen := dialog.NewGenerator()
sys.dialogGenerator = dialogGen
```

### Exception: Truly Optional Features

**Only these scenarios justify conditional integration:**

1. **Platform limitations** (e.g., WASM can't run embedded server)
2. **Experimental features** (e.g., WebRTC, modding - explicitly marked EXPERIMENTAL)
3. **Development tools** (e.g., visual tests only in CI/CD)

---

## Rollout Strategy

### Week-by-Week Deployment

**Week 1 (Phase 1):**
- Deploy to development environment
- Full testing with QA team
- Performance profiling
- If stable, deploy to staging

**Week 2 (Phase 2):**
- Deploy Phase 1 to production (if stable)
- Deploy Phase 2 to development
- Continue testing

**Week 3 (Phase 3):**
- Deploy Phase 2 to production
- Deploy Phase 3 to development
- Performance optimization

**Week 4 (Phase 4):**
- Deploy Phase 3 to production
- Experimental features to staging only
- Gather feedback

### Rollback Plan

Each phase is independent. If issues arise:

1. **Identify problematic package**
2. **Comment out its integration** (3 lines typically)
3. **Rebuild and redeploy**
4. **File issue for investigation**

Example rollback:
```go
// ROLLBACK: Balance system causing crashes
// balanceSystem := balance.NewBalanceSystem(world)
// sys.balanceSystem = balanceSystem
// game.World.AddSystem(sys.balanceSystem)
```

---

## Success Metrics

### Per-Phase Verification

After each phase, verify:

- [ ] **Build Success:** Both client and server build without errors
- [ ] **Test Coverage:** All package tests pass (`go test ./pkg/...`)
- [ ] **Runtime Stability:** No panics during 30-minute playthrough
- [ ] **Performance:** FPS ≥60 with 2000 entities, <500MB client RAM
- [ ] **Feature Visibility:** Integrated features appear in gameplay
- [ ] **Log Hygiene:** No unexpected errors/warnings in logs
- [ ] **Network Stability:** Multiplayer works reliably

### Overall Success Criteria

- [ ] All 19 runtime packages integrated
- [ ] All tests passing
- [ ] Performance targets met
- [ ] Zero known crashes
- [ ] User-visible feature improvements
- [ ] Clean separation of concerns
- [ ] No feature flags (except experimental)

---

## Conclusion

**Total Integration Effort:** 19 packages over 4 weeks

**Complexity Breakdown:**
- Low complexity: 10 packages (1-2 hours each)
- Medium complexity: 7 packages (2-3.5 hours each)
- High complexity: 2 packages (4-5 hours each)

**Estimated Total Hours:** 50-60 hours

**Key Principles:**
1. All features are baseline features
2. No feature flags, no toggles
3. Unconditional registration
4. Always-on initialization
5. If it exists, it's active

**Next Steps:**
1. Begin Phase 1 immediately
2. Create tracking issue for each package integration
3. Update this document with actual integration times
4. Document any unexpected issues

---

**Document Version:** 1.0  
**Last Updated:** December 13, 2025  
**Maintainer:** Integration Planning System
