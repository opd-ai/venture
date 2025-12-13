# Integration Plan - December 13, 2025

## Core Principle

**All packages become unconditionally active. No feature flags. No optional toggles.**

Dormant packages represent complete, tested implementations awaiting integration. This plan provides exact file changes, code snippets, and activation steps to integrate all 22 runtime-relevant dormant packages into the game client and server as always-on baseline features.

---

## Exclusions

The following dormant packages are **excluded** from integration as they are development tools, not runtime features:
- `pkg/audit/features` - Feature audit tool
- `pkg/procgen/audit` - Procgen audit tool  
- `pkg/visualtest` - Visual testing framework
- `pkg/visualtest/parity` - Visual parity testing
- `pkg/migration` - Database migration utility

**Total runtime dormant packages to integrate**: 22 (excluding 14 stubs/parent directories and 4 dev tools)

---

## Phase 1: Core Systems (Week 1 - Days 1-3)

**Goal**: Activate foundational systems that enhance gameplay without external dependencies  
**Packages**: 6  
**Effort**: Small-Medium  
**Verification**: Build + manual testing

---

### 1.1: Companion Learning System

**Package**: `pkg/companion/learning`  
**LOC**: 2,107  
**Type**: ECS System  
**Completeness**: Complete (doc, tests, 2,107 LOC)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import** (after line 40):
```go
"github.com/opd-ai/venture/pkg/companion/learning"
```

2. **Add to system container** (around line 110):
```go
type systemsContainer struct {
    // ... existing fields ...
    companionLearningSystem *learning.System
}
```

3. **Initialize system** (around line 600, after companion system initialization):
```go
// Phase 1.1: Companion learning system
sys.companionLearningSystem = learning.NewSystem()
logger.WithField("system_name", "companion_learning").Debug("Created companion learning system")
```

4. **Register system** (around line 935):
```go
game.World.AddSystem(sys.companionLearningSystem) // Phase 1.1: Companion AI learning
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Spawn companion, observe learning behavior over time
```

**Effort**: Small (10 minutes)

---

### 1.2: Performance Monitoring System

**Package**: `pkg/engine/performance`  
**LOC**: 1,488  
**Type**: ECS System  
**Completeness**: Complete (doc, tests, 1,488 LOC)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/engine/performance"
```

2. **Add to system container**:
```go
performanceSystem *performance.MonitoringSystem
```

3. **Initialize system** (early, around line 500):
```go
// Phase 1.2: Performance monitoring
sys.performanceSystem = performance.NewMonitoringSystem()
logger.WithField("system_name", "performance").Debug("Created performance monitoring system")
```

4. **Register system** (early, around line 850):
```go
game.World.AddSystem(sys.performanceSystem) // Phase 1.2: Performance monitoring
```

5. **Optional**: Add metrics display to debug UI (in `pkg/engine/debug_ui.go`):
```go
func (d *DebugUI) renderPerformanceMetrics(screen *ebiten.Image, perfSystem *performance.MonitoringSystem) {
    metrics := perfSystem.GetMetrics()
    // Display FPS, memory, entity count, etc.
}
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Check debug overlay for performance metrics
```

**Effort**: Medium (30 minutes with UI integration)

---

### 1.3: Prestige System

**Package**: `pkg/engine/prestige`  
**LOC**: 1,299  
**Type**: System  
**Completeness**: Complete (doc, tests, 1,299 LOC)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/engine/prestige"
```

2. **Add to system container**:
```go
prestigeSystem *prestige.System
```

3. **Initialize system** (after progression system, around line 700):
```go
// Phase 1.3: Prestige system
sys.prestigeSystem = prestige.NewSystem()
logger.WithField("system_name", "prestige").Debug("Created prestige system")
```

4. **Register system** (after progression system, around line 881):
```go
game.World.AddSystem(sys.prestigeSystem) // Phase 1.3: Prestige levels and resets
```

5. **Add UI display** (in character sheet or progression UI):
```go
// Display prestige level and bonuses
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Play to high level, trigger prestige reset
```

**Effort**: Medium (45 minutes with UI)

---

### 1.4: Balance Calculator Integration

**Package**: `pkg/balance`  
**LOC**: 1,232  
**Type**: Utility Library  
**Completeness**: Complete (doc, tests, 1,232 LOC)

**File**: Multiple systems that calculate difficulty/scaling

**Changes**:

1. **Import in combat system** (`pkg/engine/combat_system.go`):
```go
"github.com/opd-ai/venture/pkg/balance"
```

2. **Use in damage calculation**:
```go
func (cs *CombatSystem) calculateDamage(attacker, defender *Entity) float64 {
    baseDamage := attacker.GetComponent("attack").(*AttackComponent).Power
    scaling := balance.CalculateScaling(attacker.Level, cs.difficulty)
    return baseDamage * scaling
}
```

3. **Import in spawn systems** (`pkg/engine/spawn_system.go`):
```go
difficulty := balance.CalculateEnemyDifficulty(playerLevel, depth, genreID)
```

**Verification**:
```bash
go test ./pkg/balance/...
go build ./cmd/client
# Verify difficulty scaling feels appropriate
```

**Effort**: Medium (1 hour - multiple integration points)

---

### 1.5: Stability Monitoring

**Package**: `pkg/stability`  
**LOC**: 599  
**Type**: Monitoring/Crash Reporting  
**Completeness**: Complete (doc, tests, 599 LOC)

**Files**: `cmd/client/main.go`, `cmd/server/main.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/stability"
```

2. **Enable monitoring** (in `main()`, early):
```go
func main() {
    // Phase 1.5: Stability monitoring
    stability.EnableMonitoring()
    defer stability.ReportCrash()
    
    // ... rest of main
}
```

3. **Add crash log directory** (optional):
```go
stability.SetCrashDir("./crash_logs")
```

**Verification**:
```bash
go build ./cmd/client ./cmd/server
# Intentionally crash the client
# Verify crash report generated
```

**Effort**: Small (15 minutes)

---

### 1.6: User Experience Enhancements

**Package**: `pkg/ux`  
**LOC**: 1,571  
**Type**: UX Systems  
**Completeness**: Complete (doc, tests, 1,571 LOC)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/ux"
```

2. **Add to system container**:
```go
uxManager     *ux.Manager
tooltipSystem *ux.TooltipSystem
```

3. **Initialize** (around line 650):
```go
// Phase 1.6: UX enhancements
sys.uxManager = ux.NewManager()
sys.tooltipSystem = sys.uxManager.TooltipSystem()
logger.WithField("system_name", "ux").Debug("Created UX enhancement systems")
```

4. **Register systems** (around line 920):
```go
game.World.AddSystem(sys.tooltipSystem) // Phase 1.6: Tooltips
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Hover over items, verify tooltips appear
```

**Effort**: Small (20 minutes)

---

## Phase 2: World & Economy (Week 1 - Days 4-5)

**Goal**: Activate world simulation systems  
**Packages**: 2  
**Effort**: Small

---

### 2.1: Economy System

**Package**: `pkg/world/economy`  
**LOC**: 3,002  
**Type**: ECS System  
**Completeness**: Complete (doc, tests, 8 exports)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/world/economy"
```

2. **Add to system container**:
```go
economySystem *economy.System
```

3. **Initialize** (around line 720):
```go
// Phase 2.1: Economy system
sys.economySystem = economy.NewSystem()
logger.WithField("system_name", "economy").Debug("Created economy system")
```

4. **Register** (after commerce system, around line 907):
```go
game.World.AddSystem(sys.economySystem) // Phase 2.1: Dynamic economy
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Trade with NPCs, observe price fluctuations
```

**Effort**: Small (10 minutes)

---

### 2.2: Raid System

**Package**: `pkg/world/raids`  
**LOC**: 3,043  
**Type**: ECS System + Generator  
**Completeness**: Complete (doc, tests, 11 exports)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/world/raids"
```

2. **Add to system container**:
```go
raidSystem *raids.System
```

3. **Initialize** (around line 730):
```go
// Phase 2.2: Raid system
sys.raidSystem = raids.NewSystem(game.World)
logger.WithField("system_name", "raids").Debug("Created raid system")
```

4. **Register** (after world events, around line 929):
```go
game.World.AddSystem(sys.raidSystem) // Phase 2.2: Dynamic raids
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Wait for raid events to trigger
```

**Effort**: Small (15 minutes)

---

## Phase 3: Procedural Generation (Week 2 - Days 1-3)

**Goal**: Activate dormant generators  
**Packages**: 4  
**Effort**: Medium-Large

---

### 3.1: Dialog Generator

**Package**: `pkg/procgen/dialog`  
**LOC**: 2,573  
**Type**: Generator  
**Completeness**: Complete (doc, tests, 6 exports)

**File**: `pkg/engine/dialog_system.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/procgen/dialog"
```

2. **Add generator to DialogSystem**:
```go
type DialogSystem struct {
    world       *World
    dialogGen   *dialog.Generator // Phase 3.1
    // ... existing fields
}
```

3. **Initialize in NewDialogSystem**:
```go
func NewDialogSystem(world *World, seed int64) *DialogSystem {
    return &DialogSystem{
        world:     world,
        dialogGen: dialog.NewGenerator(), // Phase 3.1
        // ...
    }
}
```

4. **Use generator in NPC dialog creation**:
```go
func (ds *DialogSystem) createNPCDialog(npc *Entity, seed int64) []DialogOption {
    params := dialog.GenerationParams{
        NPCType:    npc.Type,
        Personality: npc.Personality,
        QuestGiver: npc.HasQuest,
    }
    return ds.dialogGen.Generate(seed, params)
}
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Talk to NPCs, verify varied dialog
```

**Effort**: Medium (1 hour)

---

### 3.2: Entity Generator

**Package**: `pkg/procgen/entity`  
**LOC**: 2,161  
**Type**: Generator  
**Completeness**: Complete (doc, tests, 5 exports)

**File**: Multiple spawn systems

**Changes**:

1. **Add import** to spawn systems:
```go
"github.com/opd-ai/venture/pkg/procgen/entity"
```

2. **Create generator instance**:
```go
entityGen := entity.NewGenerator()
```

3. **Replace manual entity creation**:
```go
func (ss *SpawnSystem) spawnEnemy(x, y float64, seed int64) *Entity {
    params := entity.GenerationParams{
        EntityType: "enemy",
        Level:      ss.calculateLevel(),
        GenreID:    ss.genreID,
    }
    return ss.entityGen.Generate(seed, params)
}
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Verify enemy variety and stats
```

**Effort**: Large (2-3 hours - affects multiple systems)

---

### 3.3: Legendary Item System

**Package**: `pkg/procgen/legendary`  
**LOC**: 3,078  
**Type**: ECS System + Generator  
**Completeness**: Complete (doc, tests, 5 exports)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/procgen/legendary"
```

2. **Add to system container**:
```go
legendarySystem *legendary.System
```

3. **Initialize** (around line 750):
```go
// Phase 3.3: Legendary item system
sys.legendarySystem = legendary.NewSystem(game.World, *gameSeed)
logger.WithField("system_name", "legendary").Debug("Created legendary item system")
```

4. **Register** (after item system, around line 903):
```go
game.World.AddSystem(sys.legendarySystem) // Phase 3.3: Legendary items
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Play until legendary item drops
```

**Effort**: Small (15 minutes)

---

### 3.4: Minigame Implementations

**Package**: `pkg/procgen/minigame/games`  
**LOC**: 1,792  
**Type**: ECS System  
**Completeness**: Complete (doc, tests, 9 exports)

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/procgen/minigame/games"
```

2. **Add to system container**:
```go
minigameGamesSystem *games.System
```

3. **Initialize** (after minigame framework):
```go
// Phase 3.4: Minigame implementations
sys.minigameGamesSystem = games.NewSystem(game.World)
logger.WithField("system_name", "minigame_games").Debug("Created minigame games system")
```

4. **Register**:
```go
game.World.AddSystem(sys.minigameGamesSystem) // Phase 3.4: Minigame implementations
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Trigger minigames (lockpicking, hacking, etc.)
```

**Effort**: Small (10 minutes)

---

## Phase 4: Integration Modules (Week 2 - Days 4-5)

**Goal**: Activate cross-system integrations  
**Packages**: 5  
**Effort**: Small (each is self-contained)

---

### 4.1: Choice & Consequences

**Package**: `pkg/integration/choice_consequences`  
**LOC**: 1,545  
**Type**: Integration System

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/choice_consequences"
```

2. **Initialize and register**:
```go
choiceSystem := choice_consequences.NewSystem(game.World)
game.World.AddSystem(choiceSystem) // Phase 4.1: Choice consequences
```

**Effort**: Small (5 minutes)

---

### 4.2: Guild Vehicle Integration

**Package**: `pkg/integration/guild_vehicle`

**Changes**:
```go
"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
guildVehicleSystem := guild_vehicle.NewSystem(game.World)
game.World.AddSystem(guildVehicleSystem) // Phase 4.2: Guild vehicles
```

**Effort**: Small (5 minutes)

---

### 4.3: Narrative-World Integration

**Package**: `pkg/integration/narrative_world`

**Changes**:
```go
"github.com/opd-ai/venture/pkg/integration/narrative_world"
narrativeWorldSystem := narrative_world.NewSystem(game.World)
game.World.AddSystem(narrativeWorldSystem) // Phase 4.3: Narrative affects world
```

**Effort**: Small (5 minutes)

---

### 4.4: Political Warfare Integration

**Package**: `pkg/integration/political_warfare`

**Changes**:
```go
"github.com/opd-ai/venture/pkg/integration/political_warfare"
politicalWarfareSystem := political_warfare.NewSystem(game.World)
game.World.AddSystem(politicalWarfareSystem) // Phase 4.4: Political warfare
```

**Effort**: Small (5 minutes)

---

### 4.5: Territory Siege Integration

**Package**: `pkg/integration/territory_siege`

**Changes**:
```go
"github.com/opd-ai/venture/pkg/integration/territory_siege"
territorySiegeSystem := territory_siege.NewSystem(game.World)
game.World.AddSystem(territorySiegeSystem) // Phase 4.5: Territory sieges
```

**Effort**: Small (5 minutes)

---

## Phase 5: Network Extensions (Week 3 - Days 1-2)

**Goal**: Activate advanced networking features  
**Packages**: 3  
**Effort**: Small-Large

---

### 5.1: Mobile Federation

**Package**: `pkg/network/federation/mobile`  
**LOC**: 1,349  
**Type**: ECS System

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/network/federation/mobile"
```

2. **Initialize** (after federation system):
```go
// Phase 5.1: Mobile federation support
mobileFedSystem := mobile.NewSystem(sys.federationManager)
logger.WithField("system_name", "mobile_federation").Debug("Created mobile federation system")
```

3. **Register**:
```go
game.World.AddSystem(mobileFedSystem) // Phase 5.1: Mobile federation
```

**Verification**:
```bash
go build ./cmd/client
# Test on mobile device with federation
```

**Effort**: Small (10 minutes)

---

### 5.2: Network Resilience

**Package**: `pkg/network/resilience`  
**LOC**: 1,334  
**Type**: Network Middleware

**File**: `cmd/server/main.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/network/resilience"
```

2. **Wrap server** (after server creation):
```go
// Phase 5.2: Network resilience (retry, circuit breaker)
resilientServer := resilience.WrapServer(server, resilience.DefaultConfig())
```

3. **Use resilient server**:
```go
if err := resilientServer.Start(); err != nil {
    log.Fatal(err)
}
```

**Verification**:
```bash
go build ./cmd/server
./cmd/server/server
# Test with network interruptions
```

**Effort**: Small (15 minutes)

---

### 5.3: Security Middleware

**Package**: `pkg/security`  
**LOC**: 1,503  
**Type**: Network Middleware

**File**: `cmd/server/main.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/security"
```

2. **Add security manager**:
```go
// Phase 5.3: Security validation
securityMgr := security.NewManager()
server.AddMiddleware(securityMgr.ValidatePacket)
server.AddMiddleware(securityMgr.RateLimitCheck)
```

**Verification**:
```bash
go build ./cmd/server
# Test with malformed packets
```

**Effort**: Medium (30 minutes)

---

## Phase 6: Advanced Features (Week 3 - Days 3-5)

**Goal**: Activate remaining specialized systems  
**Packages**: 2  
**Effort**: Large

---

### 6.1: Tile Rendering System

**Package**: `pkg/rendering/tiles`  
**LOC**: 6,602  
**Type**: Generator

**File**: Terrain rendering systems

**Changes**:

1. **Add import** to terrain render system:
```go
"github.com/opd-ai/venture/pkg/rendering/tiles"
```

2. **Create tile generator**:
```go
type TerrainRenderSystem struct {
    // ... existing fields
    tileGen *tiles.Generator // Phase 6.1
}
```

3. **Use in terrain rendering**:
```go
func (trs *TerrainRenderSystem) renderTerrain(screen *ebiten.Image, terrain *Terrain) {
    tileSet := trs.tileGen.GenerateTileSet(terrain.Seed, terrain.Type)
    // Use tileSet for rendering
}
```

**Verification**:
```bash
go build ./cmd/client
./cmd/client/client -seed 12345
# Verify terrain looks better with tiles
```

**Effort**: Large (2-3 hours - render system refactor)

---

### 6.2: WebRTC Federation

**Package**: `pkg/network/federation/webrtc`  
**LOC**: 4,246  
**Type**: Network Transport

**File**: `cmd/client/main.go`, `cmd/server/main.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/network/federation/webrtc"
```

2. **Add flag**:
```go
var useWebRTC = flag.Bool("webrtc", false, "Use WebRTC for federation")
```

3. **Conditional initialization**:
```go
if *useWebRTC {
    transport := webrtc.NewTransport()
    federationManager.SetTransport(transport)
}
```

**Verification**:
```bash
go build ./cmd/client ./cmd/server
./cmd/server/server -webrtc
./cmd/client/client -webrtc -connect localhost:8080
```

**Effort**: Large (4-6 hours - network architecture changes)

---

## Verification Strategy

### Per-Phase Verification
After each phase:

1. **Build**: `go build ./cmd/client ./cmd/server`
2. **Test**: `go test ./pkg/...`
3. **Run**: Launch client and manually verify new features
4. **Check logs**: `grep "Phase [0-9]" client.log | grep -i error`

### Integration Test
After all phases:

```bash
# Full build
make clean && make

# Run all tests
go test ./pkg/... -cover

# Run client for 10 minutes, verify:
# - All systems appear in logs
# - No crashes
# - Features work as expected

# Check system registration count
grep "AddSystem" cmd/client/handlers.go | wc -l
# Should be: 52 (current) + 22 (new) = 74 systems
```

---

## Success Criteria

- [ ] All 22 runtime dormant packages integrated
- [ ] No feature flags introduced
- [ ] All systems registered unconditionally
- [ ] `go build ./cmd/client ./cmd/server` succeeds
- [ ] `go test ./pkg/...` passes
- [ ] Client runs without crashes for 30+ minutes
- [ ] All new systems appear in debug logs
- [ ] Manual testing confirms feature functionality

---

## Rollback Strategy

If integration causes issues:

1. **Git revert**: Each phase should be a separate commit
2. **Disable system**: Comment out `AddSystem()` call temporarily
3. **Fix forward**: Fix bugs rather than reverting if possible

**Anti-Pattern**: Do NOT add feature flags to "disable" systems. Fix bugs instead.

---

## Timeline Summary

| Phase | Duration | Packages | Effort |
|-------|----------|----------|--------|
| Phase 1: Core Systems | 3 days | 6 | Medium |
| Phase 2: World & Economy | 2 days | 2 | Small |
| Phase 3: Procedural Gen | 3 days | 4 | Large |
| Phase 4: Integration Modules | 2 days | 5 | Small |
| Phase 5: Network Extensions | 2 days | 3 | Medium |
| Phase 6: Advanced Features | 3 days | 2 | Large |
| **Total** | **15 days** | **22** | **Mixed** |

---

## Post-Integration

After all phases complete:

1. **Update documentation**: Reflect all active features
2. **Remove old feature flags**: From `cmd/client/util.go`
3. **Update README**: List all baseline features
4. **Release notes**: Document newly-activated features
5. **Performance testing**: Ensure 60 FPS maintained
6. **User testing**: Gather feedback on new features

---

## Notes

- Each system is independent - phases can be reordered if dependencies allow
- Integration modules (Phase 4) require their dependent systems to be active
- Network extensions (Phase 5) require server restarts for testing
- Advanced features (Phase 6) require more extensive testing

**Remember**: No feature flags. All integrations are unconditional. If it exists, it's active.
