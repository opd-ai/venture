# System Registration Patterns in Venture

This document describes the three standardized patterns for integrating game logic into the Venture ECS architecture. Understanding when to use each pattern is critical for maintainability and performance.

## Pattern Overview

| Pattern | Use Case | Registration | Example Packages |
|---------|----------|--------------|------------------|
| **ECS System Wrapper** | Logic that runs every frame | `world.AddSystem(wrapper)` | `prestige`, `qol`, `political_warfare` |
| **Manager-Only** | Event-driven logic called on demand | Stored in World or passed to systems | `guild_housing`, `choice_consequences`, `guild_vehicle` |
| **Hybrid System** | Combination of per-frame updates + API calls | Both registration and direct access | `narrative_world`, `trade_routes` |

## Pattern 1: ECS System Wrapper

**When to use:**
- Logic must run every frame (or every N frames)
- Processes entities with specific components
- Needs to react to game state changes in real-time
- Examples: AI updates, physics simulation, status effect ticking

**Structure:**
```go
// In pkg/engine/mysystem/system.go
type System struct {
    manager *Manager
    logger  *logrus.Entry
}

func NewSystem() *System {
    return &System{
        manager: NewManager(),
        logger:  logrus.WithField("system", "mysystem"),
    }
}

// Standard ECS Update signature
func (s *System) Update(entities []*Entity, deltaTime float64) {
    for _, entity := range entities {
        if !entity.HasComponent("mycomponent") {
            continue
        }
        // Process entity
    }
}

// Optional: Provide Manager access for direct API calls
func (s *System) GetManager() *Manager {
    return s.manager
}
```

**Registration (server/client):**
```go
// cmd/server/main.go or cmd/client/main.go
mySystem := mysystem.NewSystem()
world.AddSystem(mySystem)
```

**Registration with Wrapper (if signature doesn't match):**
```go
// cmd/server/system_wrappers.go
type mySystemWrapper struct {
    system *mysystem.System
}

func (w *mySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
    // Adapt if needed (e.g., convert entity types, ignore entities parameter)
    w.system.Update(deltaTime)
}

// cmd/server/main.go
mySystem := mysystem.NewSystem()
world.AddSystem(&mySystemWrapper{system: mySystem})
```

**Examples:**
- `pkg/engine/prestige` - Checks for prestige ability unlocks every frame
- `pkg/engine/qol` - Processes auto-loot, craft queues per frame
- `pkg/integration/political_warfare` - Updates war/treaty timers
- `pkg/integration/narrative_world` - Progresses story beats

## Pattern 2: Manager-Only

**When to use:**
- Logic is event-driven (triggered by player actions)
- No per-frame processing needed
- Pure data management (CRUD operations)
- Examples: Housing management, guild storage, mailboxes

**Structure:**
```go
// In pkg/integration/mymanager/manager.go
type Manager struct {
    data map[string]*MyData
    mu   sync.RWMutex
}

func NewManager() *Manager {
    return &Manager{
        data: make(map[string]*MyData),
    }
}

// API methods called on-demand
func (m *Manager) CreateResource(id string, params Params) (*MyData, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    // Create resource
}

func (m *Manager) GetResource(id string) (*MyData, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    // Retrieve resource
}
```

**Registration (server/client):**
```go
// Store in World for global access
myManager := mymanager.NewManager()
// Pass to systems that need it
housingSystem := housing.NewSystem(myManager)
world.AddSystem(housingSystem)
```

**Or store manager directly:**
```go
// Add GetManager() method to World
world.guildHousingManager = guild_housing.NewManager()

// Systems access via World reference
func (s *SomeSystem) Update(entities []*Entity, deltaTime float64) {
    manager := s.world.GetGuildHousingManager()
    manager.CreateGuildHouse(...)
}
```

**Examples:**
- `pkg/integration/guild_housing` - Manages guild house permissions, storage
- `pkg/integration/choice_consequences` - Tracks player choices
- `pkg/integration/guild_vehicle` - Manages guild fleet ownership
- `pkg/world/housing` - Housing plot allocation and persistence

## Pattern 3: Hybrid System

**When to use:**
- Needs both per-frame updates AND on-demand API calls
- Updates background state + responds to events
- Examples: Trade route updates, world events

**Structure:**
```go
// In pkg/integration/mysystem/system.go
type System struct {
    world   *engine.World
    manager *Manager
    logger  *logrus.Entry
}

func NewSystem(world *engine.World) *System {
    return &System{
        world:   world,
        manager: NewManager(),
        logger:  logrus.WithField("system", "mysystem"),
    }
}

// Standard ECS Update for per-frame logic
func (s *System) Update(entities []*engine.Entity, deltaTime float64) {
    // Update background state
    s.manager.UpdateTimers(deltaTime)
}

// Public API for event-driven calls
func (s *System) TriggerEvent(eventID string) error {
    return s.manager.ProcessEvent(eventID)
}

// Provide Manager access
func (s *System) GetManager() *Manager {
    return s.manager
}
```

**Registration:**
```go
mySystem := mysystem.NewSystem(world)
world.AddSystem(mySystem)

// Store reference for API access
world.mySystem = mySystem

// Later, trigger events
world.mySystem.TriggerEvent("player-entered-zone")
```

**Examples:**
- `pkg/integration/trade_routes` - Updates routes per frame + manual route creation
- `pkg/integration/narrative_world` - Progresses narrative + manual story triggers
- `pkg/integration/world_events` - Spawns events per frame + manual event dispatch

## Client vs. Server Registration

### Client-Only Systems
UI, rendering, input handling, audio, visual effects.

**Examples:**
- `pkg/rendering/*` - All rendering systems
- `pkg/audio/*` - Audio synthesis and playback
- Menu, HUD, inventory UI systems

**Registration location:** `cmd/client/main.go` only

### Server-Only Systems
Authoritative game state, anti-cheat validation, server-side economy.

**Examples:**
- Network snapshot generation
- Authoritative combat resolution
- Economy price calculations

**Registration location:** `cmd/server/main.go` only

### Shared Systems (Client + Server)
Core gameplay logic that must stay synchronized.

**Examples:**
- Movement, collision, physics
- Combat, status effects
- Quests, progression
- Prestige, QoL

**Registration location:** Both `cmd/client/main.go` and `cmd/server/main.go`

**Critical:** Shared systems MUST use deterministic logic (seed-based RNG, no `time.Now()` in gameplay).

## Decision Tree

```
Does the logic need to run every frame?
├─ YES → ECS System Wrapper (Pattern 1)
│   └─ Does Update() signature match engine.System?
│       ├─ YES → Register directly with world.AddSystem()
│       └─ NO → Create wrapper in system_wrappers.go
│
└─ NO → Is it event-driven or API-based?
    ├─ ONLY events/API → Manager-Only (Pattern 2)
    │   └─ Store in World or pass to dependent systems
    │
    └─ BOTH events + background updates → Hybrid System (Pattern 3)
        └─ Register System + expose Manager via GetManager()
```

## Performance Guidelines

### ECS System Wrapper
- **Hot path:** Optimized for 60+ FPS with 2000+ entities
- **Entity filtering:** Use `HasComponent()` early to skip irrelevant entities
- **Component access:** Prefer hot-path cache methods (`GetHealth()`, `GetPosition()`)
- **Avoid:** String concatenation, allocations in tight loops

### Manager-Only
- **Thread safety:** All public methods MUST use mutex locks
- **Lock granularity:** Use `RWMutex` for read-heavy workloads
- **Avoid:** Holding locks during I/O or expensive operations

### Hybrid System
- **Balance:** Keep Update() lightweight, offload heavy work to API methods
- **Defer:** Use goroutines for async operations (with proper error handling)

## Testing Patterns

### ECS System Wrapper
```go
func TestSystemUpdate(t *testing.T) {
    world := engine.NewWorld()
    system := NewSystem()
    
    // Create test entity with required components
    entity := world.CreateEntity()
    entity.AddComponent(&MyComponent{Value: 10})
    
    // Update and verify
    system.Update([]*engine.Entity{entity}, 0.016)
    
    comp := entity.GetComponent("mycomponent").(*MyComponent)
    if comp.Value != 11 {
        t.Errorf("expected 11, got %d", comp.Value)
    }
}
```

### Manager-Only
```go
func TestManagerCRUD(t *testing.T) {
    manager := NewManager()
    
    // Create
    data, err := manager.CreateResource("test-id", Params{})
    if err != nil {
        t.Fatalf("create failed: %v", err)
    }
    
    // Read
    retrieved, err := manager.GetResource("test-id")
    if err != nil {
        t.Fatalf("get failed: %v", err)
    }
    
    if retrieved.ID != data.ID {
        t.Errorf("ID mismatch")
    }
}
```

## Migration Guide

### Converting Manager-Only to ECS System Wrapper

If you realize a Manager-Only pattern needs per-frame updates:

1. Keep existing Manager code unchanged
2. Create a new `System` struct that wraps the Manager
3. Add `Update([]*Entity, deltaTime)` method to System
4. Provide `GetManager()` for API access
5. Update registration from direct Manager to System

**Example:**
```go
// Before (Manager-Only)
manager := mymanager.NewManager()
world.StoreManager("mymanager", manager)

// After (ECS System Wrapper)
system := mymanager.NewSystem() // Wraps Manager internally
world.AddSystem(system)
// API access still available
system.GetManager().DoSomething()
```

### Converting ECS System Wrapper to Manager-Only

If a system no longer needs per-frame updates:

1. Remove `Update()` method
2. Remove System wrapper
3. Keep Manager and its API methods
4. Update registration to store Manager directly

**Warning:** Only do this if you're certain no per-frame logic is needed. Review all call sites.

## Common Pitfalls

1. **Mixing patterns inconsistently** - Pick one pattern per package and stick to it
2. **Heavy Update() logic** - Offload expensive operations to background goroutines
3. **Missing thread safety** - All Manager methods MUST be thread-safe
4. **Forgetting server registration** - Shared systems must be registered on both client and server
5. **Type assertion panics** - Always use comma-ok pattern for component type assertions

## References

- `cmd/server/system_wrappers.go` - All server-side wrapper implementations
- `cmd/client/system_wrappers.go` - All client-side wrapper implementations
- `pkg/engine/interfaces.go` - System and Component interfaces
- `pkg/engine/ecs.go` - World and Entity implementation
- `AUDIT.md` - Historical context on system registration decisions
