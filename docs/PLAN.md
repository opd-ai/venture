# Integration Plan - December 2025

## Principle

**All packages become unconditionally active.** No feature flags. No optional toggles. If a package exists, it should be active.

---

## Executive Summary

| Phase | Focus | Packages | Effort | Timeline |
|-------|-------|----------|--------|----------|
| 1 | Verification | 4 | Small | 1-2 days |
| 2 | World Events | 2 | Medium | 2-3 days |
| 3 | Economy | 1 | Large | 3-5 days |
| **Total** | | **7** | | **6-10 days** |

The project is already **82% integrated**. Only 7 packages require active integration work.

---

## Phase 1: Verification (Small Effort)

**Goal:** Confirm packages are properly imported and systems are registered.

### 1.1 Verify `pkg/integration/guild_vehicle`

**Status**: System exists (`guildVehicleSystem`), verify package import.

```bash
# Check if package is imported
grep -rn "integration/guild_vehicle" cmd/client/

# Expected: Import in handlers.go
```

**If not imported:**
```go
// File: cmd/client/handlers.go
import (
    guildvehicle "github.com/opd-ai/venture/pkg/integration/guild_vehicle"
)

// In initializeV8Systems():
sys.guildVehicleFleetManager = guildvehicle.NewFleetManager()
```

**Verification:**
- [ ] Import statement present
- [ ] FleetManager created
- [ ] System uses fleet manager for formations

---

### 1.2 Verify `pkg/procgen/legendary`

**Status**: System exists (`legendaryQuestSystem`), verify package import.

```bash
# Check if package is imported
grep -rn "procgen/legendary" cmd/client/
```

**If not imported:**
```go
// File: cmd/client/handlers.go
import (
    "github.com/opd-ai/venture/pkg/procgen/legendary"
)

// In quest generation:
legendaryGen := legendary.NewGenerator()
```

**Verification:**
- [ ] Import statement present
- [ ] Generator used for legendary quest creation
- [ ] System registered unconditionally

---

### 1.3 Verify `pkg/procgen/dialog`

**Status**: Engine has `MarkovDialogProvider`, verify it uses this package.

```bash
# Check import chain
grep -rn "procgen/dialog" pkg/engine/ cmd/client/
```

**If not imported:**
```go
// File: pkg/engine/markov_dialog_provider.go
import (
    "github.com/opd-ai/venture/pkg/procgen/dialog"
)

// Use dialog.NewMarkovGenerator instead of inline implementation
```

**Verification:**
- [ ] MarkovDialogProvider uses procgen/dialog
- [ ] Personality system connected
- [ ] Genre-specific corpus loaded

---

### 1.4 Verify `pkg/audio/synthesis`

**Status**: Used transitively by `pkg/audio/music` and `pkg/audio/sfx`.

```bash
# Check if imported by audio packages
grep -rn "audio/synthesis" pkg/audio/
```

**Expected:** Imported by `pkg/audio/music/generator.go` and `pkg/audio/sfx/generator.go`.

**Verification:**
- [ ] Oscillator types used
- [ ] Envelope system connected
- [ ] Waveform generation working

---

## Phase 2: World Events Integration (Medium Effort)

**Goal:** Activate world event system and choice consequence tracking.

### 2.1 `pkg/integration/world_events`

**Purpose:** Dynamic world events from player actions (guild wars, economy, weather disasters).

**Files to modify:**
1. `cmd/client/handlers.go`

**Implementation:**

```go
// File: cmd/client/handlers.go

import (
    worldevents "github.com/opd-ai/venture/pkg/integration/world_events"
)

// Add to v8Systems struct:
type v8Systems struct {
    // ... existing fields ...
    worldEventsManager *worldevents.EventManager
    worldEventsSystem  *worldEventsSystemWrapper
}

// In initializeV8Systems():
sys.worldEventsManager = worldevents.NewEventManager(seed)
sys.worldEventsSystem = &worldEventsSystemWrapper{
    manager: sys.worldEventsManager,
    world:   game.World,
}

// System wrapper:
type worldEventsSystemWrapper struct {
    manager *worldevents.EventManager
    world   *engine.World
}

func (w *worldEventsSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
    // Check for triggered events every 5 seconds
    w.manager.ProcessTriggers(deltaTime)
    
    // Apply active event impacts
    for _, event := range w.manager.GetActiveEvents() {
        w.applyEventImpacts(event)
    }
}

// Register in initializeV8Systems():
game.World.AddSystem(sys.worldEventsSystem)
```

**Verification:**
- [ ] Import added
- [ ] EventManager created with world seed
- [ ] System registered unconditionally
- [ ] Events trigger from guild warfare
- [ ] Weather disasters connect to weather system

---

### 2.2 `pkg/integration/choice_consequences`

**Purpose:** Persistent choice tracking for branching narratives.

**Files to modify:**
1. `cmd/client/handlers.go`
2. Connect to `pkg/narrative/branching`

**Implementation:**

```go
// File: cmd/client/handlers.go

import (
    choicecons "github.com/opd-ai/venture/pkg/integration/choice_consequences"
)

// Add to v8Systems struct:
type v8Systems struct {
    // ... existing fields ...
    choiceTracker *choicecons.ChoiceTracker
}

// In initializeV8Systems():
sys.choiceTracker = choicecons.NewChoiceTracker()

// Wire to branching narrative system
if sys.branchingNarrativeSystem != nil {
    sys.branchingNarrativeSystem.SetChoiceTracker(sys.choiceTracker)
}

// Wire to dialog system for NPC attitude
if sys.npcDialogSystem != nil {
    sys.npcDialogSystem.SetChoiceTracker(sys.choiceTracker)
}
```

**Note:** This may require adding `SetChoiceTracker` methods to existing systems.

**Verification:**
- [ ] Import added
- [ ] ChoiceTracker created
- [ ] Connected to branching narrative
- [ ] NPCs remember player actions
- [ ] Choices persist across sessions

---

## Phase 3: Economy Integration (Large Effort)

**Goal:** Activate federated marketplace and guild banking.

### 3.1 `pkg/world/economy`

**Purpose:** Cross-server marketplace, guild banks, dynamic pricing.

**This is the largest integration task.** The package provides:
- Federated marketplace (10,000+ listings)
- Guild banks (5,000+ items per vault)
- Dynamic pricing engine
- Cross-server trade delivery

**Files to modify:**
1. `cmd/client/handlers.go`
2. `cmd/server/main.go` (server-side pricing)
3. Trade UI integration

**Implementation:**

```go
// File: cmd/client/handlers.go

import (
    "github.com/opd-ai/venture/pkg/world/economy"
)

// Add to v8Systems struct:
type v8Systems struct {
    // ... existing fields ...
    marketplace     *economy.FederatedMarketplace
    guildBankMgr    *economy.GuildBankManager
    pricingEngine   *economy.PricingEngine
}

// In initializeV8Systems():
// Create pricing engine (shared between marketplace and trade)
sys.pricingEngine = economy.NewPricingEngine()

// Create marketplace (requires federation and mail systems)
sys.marketplace = economy.NewFederatedMarketplace(
    sys.federationClient,
    sys.mailSystem,
    sys.pricingEngine,
)

// Create guild bank manager
sys.guildBankMgr = economy.NewGuildBankManager()

// Set default interest rate
sys.guildBankMgr.SetDefaultInterestRate(0.005) // 0.5% daily

// Wire to commerce system
if sys.commerceSystem != nil {
    sys.commerceSystem.SetMarketplace(sys.marketplace)
    sys.commerceSystem.SetGuildBank(sys.guildBankMgr)
}

// Wire to trade UI
if sys.tradeUI != nil {
    sys.tradeUI.SetMarketplace(sys.marketplace)
}
```

**Server-side (cmd/server/main.go):**

```go
import (
    "github.com/opd-ai/venture/pkg/world/economy"
)

// In server initialization:
pricingEngine := economy.NewPricingEngine()
marketplace := economy.NewFederatedMarketplace(federation, nil, pricingEngine)

// Start price update goroutine
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        pricingEngine.UpdatePrices(marketplace.GetTransactionHistory())
    }
}()
```

**Verification:**
- [ ] Import added to client and server
- [ ] Marketplace created with federation
- [ ] Guild banks accessible via guild UI
- [ ] Pricing updates every 5 minutes
- [ ] Cross-server purchases work
- [ ] Items delivered via mail system

---

## Phase 4: Entity Generation Consolidation (Optional)

**Goal:** Consolidate `pkg/procgen/entity` with engine entity spawning.

**Status:** The engine has `pkg/engine/entity_spawning.go` which may duplicate `pkg/procgen/entity`.

**Recommendation:** Audit both files and consolidate:
1. Keep `pkg/procgen/entity` as the generator
2. Have `pkg/engine/entity_spawning.go` import and use it

```go
// File: pkg/engine/entity_spawning.go

import (
    entitygen "github.com/opd-ai/venture/pkg/procgen/entity"
)

// Use entitygen.NewEntityGenerator() instead of inline generation
```

**This is optional and can be deferred.**

---

## Integration Checklist Template

For each package integration, verify:

- [ ] Package imported in `cmd/client/handlers.go`
- [ ] Package imported in `cmd/mobile/mobile.go` (if applicable)
- [ ] Package imported in `cmd/server/main.go` (if server-side)
- [ ] System/manager created during initialization
- [ ] System registered via `game.World.AddSystem()`
- [ ] No lazy component initialization in Update()
- [ ] Defensive nil checks: `if ok && comp != nil`
- [ ] Tests pass: `go test ./pkg/<package>/...`
- [ ] Build succeeds: `go build ./cmd/client && go build ./cmd/server`

---

## Anti-Patterns to Avoid

```go
// ❌ BAD: Feature flag
if enableWorldEvents {
    game.World.AddSystem(worldEventsSystem)
}

// ✅ GOOD: Unconditional
game.World.AddSystem(worldEventsSystem)
```

```go
// ❌ BAD: Lazy component init
func (s *System) Update(dt float64) {
    if !s.initialized {
        player.AddComponent(NewComponent()) // Cache staleness!
        s.initialized = true
    }
}

// ✅ GOOD: Init during creation
func createPlayer(...) {
    player.AddComponent(NewComponent())
}
```

```go
// ❌ BAD: Ignore boolean return
comp, _ := entity.GetComponent("foo")
foo := comp.(*FooComponent) // PANIC if nil

// ✅ GOOD: Check both
comp, ok := entity.GetComponent("foo")
if ok && comp != nil {
    foo := comp.(*FooComponent)
}
```

---

## Verification After Each Phase

```bash
# Build both binaries
go build ./cmd/client && go build ./cmd/server

# Run all tests
go test ./pkg/...

# Run client and verify:
# - New systems appear in debug output
# - No panics or nil pointer errors
# - Features function as expected

# For Phase 2, verify:
# - World events trigger from player actions
# - NPCs remember previous choices

# For Phase 3, verify:
# - Marketplace UI shows listings
# - Guild bank accessible from guild menu
# - Cross-server purchases work
```

---

## Timeline Summary

| Week | Focus | Deliverable |
|------|-------|-------------|
| Week 1 | Phase 1-2 | Verification complete, world events active |
| Week 2 | Phase 3 | Economy system fully integrated |
| Week 3 | Testing | Full regression, performance validation |

---

## Success Criteria

- [ ] All 7 integration packages active
- [ ] Zero feature flags in integration code
- [ ] All systems unconditionally registered
- [ ] Build passes: `go build ./cmd/...`
- [ ] Tests pass: `go test ./pkg/...`
- [ ] 60+ FPS maintained with all systems
- [ ] <500MB client memory
- [ ] No nil pointer panics
