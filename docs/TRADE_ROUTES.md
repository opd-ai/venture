# Trade Routes System - Activation and Configuration Guide

**Last Updated:** 2026-02-07  
**Package:** `pkg/integration/trade_routes`  
**Status:** Fully Implemented - Configuration Required

---

## Overview

The Trade Routes system enables automated AI-controlled merchant caravans that transport goods between regions/servers, creating dynamic economic opportunities through:

- **Automated Trading**: NPC merchant caravans with procedurally generated cargo
- **Route Optimization**: Profitable trade path calculations with risk/reward balance
- **Player Escort Missions**: Protect caravans for gold rewards and combat bonuses
- **Bandit Encounters**: Procedural attacks threaten shipments (spawn rate: 0.1-1.0 per hour)
- **Guild Sponsorship**: Fund caravans to manipulate regional market prices
- **Cross-Server Integration**: Works with federation market system for dynamic pricing

**Performance Targets:**
- Route creation: <1ms per route
- Route updates: <1ms for 50 active routes
- Memory footprint: ~2KB per active route
- Recommended active routes: 10-50 per server

---

## Quick Start

### Basic Server Activation

Trade routes are **not enabled by default** and require programmatic initialization. Currently, there are no CLI flags for automatic activation. Server operators must integrate the `RouteManager` into their server code.

**Minimal Integration Example:**

```go
package main

import (
    "github.com/opd-ai/venture/pkg/integration/trade_routes"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create route manager
    serverID := "my-server-01"
    seed := int64(12345) // For deterministic cargo generation
    rm := trade_routes.NewRouteManager(serverID, seed)
    
    // Start background processing (updates every 10 seconds)
    rm.Start()
    defer rm.Stop() // Always call Stop() to prevent goroutine leaks
    
    logger.Info("Trade routes system activated")
    
    // Your server main loop here...
}
```

### Creating and Starting a Route

```go
// 1. Create a trade route between regions
route, err := rm.CreateRoute("region-alpha", "region-beta", 1000.0)
if err != nil {
    log.Fatalf("Failed to create route: %v", err)
}

// 2. Generate a caravan vehicle (requires vehicle generator)
caravan, err := rm.CreateCaravan(54321, "fantasy")
if err != nil {
    log.Fatalf("Failed to generate caravan: %v", err)
}

// 3. Start the route (activates background travel simulation)
err = rm.StartRoute(route.ID, caravan.ID)
if err != nil {
    log.Fatalf("Failed to start route: %v", err)
}

log.WithFields(logrus.Fields{
    "route_id":      route.ID,
    "route_name":    route.Name,
    "cargo_items":   len(route.Cargo),
    "profit_margin": route.ProfitMargin,
    "danger_level":  route.DangerLevel,
    "travel_time":   route.TravelTime,
}).Info("Trade route started")
```

---

## Server Integration Patterns

### Pattern 1: V9 Systems Integration (Recommended)

Add trade route initialization to `cmd/server/v9_systems.go`:

```go
// In v9_systems.go
import (
    "github.com/opd-ai/venture/pkg/integration/trade_routes"
)

func initializeV9SystemsServer(logger *logrus.Logger) (
    *housingcrafting.StationManager,
    *companionhousing.PetHomeManager,
    *guildhousing.Manager,
    *trade_routes.RouteManager, // Add this
) {
    // ... existing managers ...
    
    // Phase 57.3: Automated Trade Routes
    // Server manages NPC merchant caravans and player escort missions
    tradeRouteManager := trade_routes.NewRouteManager("server-01", 12345)
    tradeRouteManager.Start()
    
    logger.WithFields(logrus.Fields{
        "system": "trade_routes",
        "status": "initialized",
    }).Info("Trade routes manager activated")
    
    return stationManager, petHomeManager, guildHousingManager, tradeRouteManager
}
```

**Important:** When adding to `v9_systems.go`, update the return signature and ensure `Stop()` is called during server shutdown.

### Pattern 2: Optional CLI Flag Activation

Add a CLI flag to `cmd/server/main.go`:

```go
var (
    // ... existing flags ...
    
    enableTradeRoutes = flag.Bool("enable-trade-routes", false, 
        "Enable automated trade route system with AI merchant caravans")
    tradeRouteCount   = flag.Int("trade-route-count", 20, 
        "Number of active trade routes to maintain (10-50 recommended)")
)

func main() {
    flag.Parse()
    
    // ... existing initialization ...
    
    var routeManager *trade_routes.RouteManager
    if *enableTradeRoutes {
        routeManager = initializeTradeRoutes(logger, *seed)
        defer routeManager.Stop()
    }
}

func initializeTradeRoutes(logger *logrus.Logger, seed int64) *trade_routes.RouteManager {
    rm := trade_routes.NewRouteManager("server-01", seed)
    rm.Start()
    
    logger.WithFields(logrus.Fields{
        "system":         "trade_routes",
        "status":         "enabled",
        "target_routes":  *tradeRouteCount,
    }).Info("Trade routes activated via CLI flag")
    
    return rm
}
```

**Server Start Command:**
```bash
./server -enable-trade-routes -trade-route-count 30
```

### Pattern 3: World Event System Integration

For automatic route spawning based on server conditions:

```go
// In your server main loop or world event system
func (s *Server) SpawnPeriodicTradeRoutes(rm *trade_routes.RouteManager, targetRoutes int) {
    ticker := time.NewTicker(5 * time.Minute) // Spawn check every 5 minutes
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            activeRoutes := rm.GetActiveRoutes()
            if len(activeRoutes) < targetRoutes {
                // Spawn new routes to maintain target count
                toSpawn := targetRoutes - len(activeRoutes)
                for i := 0; i < toSpawn; i++ {
                    s.spawnTradeRoute(rm)
                }
            }
        case <-s.shutdownChan:
            return
        }
    }
}

func (s *Server) spawnTradeRoute(rm *trade_routes.RouteManager) {
    // Random region selection (replace with your region logic)
    regions := []string{"region-a", "region-b", "region-c", "region-d"}
    start := regions[rand.Intn(len(regions))]
    end := regions[rand.Intn(len(regions))]
    if start == end {
        return // Skip same-region routes
    }
    
    cargoValue := 500.0 + rand.Float64()*2000.0 // 500-2500 gold
    
    route, err := rm.CreateRoute(start, end, cargoValue)
    if err != nil {
        s.logger.WithError(err).Warn("Failed to create trade route")
        return
    }
    
    caravan, err := rm.CreateCaravan(s.nextCaravanID(), s.genreID)
    if err != nil {
        s.logger.WithError(err).Warn("Failed to create caravan")
        return
    }
    
    if err := rm.StartRoute(route.ID, caravan.ID); err != nil {
        s.logger.WithError(err).Warn("Failed to start trade route")
        return
    }
    
    s.logger.WithFields(logrus.Fields{
        "route_id":   route.ID,
        "start":      start,
        "end":        end,
        "cargo_value": cargoValue,
    }).Info("Spawned periodic trade route")
}
```

---

## Configuration Parameters

### RouteManager Constructor

```go
rm := trade_routes.NewRouteManager(serverID string, seed int64)
```

**Parameters:**
- `serverID`: Unique server identifier (e.g., `"server-01"`, `"us-west-1"`)
- `seed`: Deterministic seed for cargo generation (same seed = same cargo items)

### Route Creation Parameters

```go
route, err := rm.CreateRoute(startRegion, endRegion string, cargoValue float64)
```

**Parameters:**
- `startRegion`: Origin region/server ID
- `endRegion`: Destination region/server ID (must differ from start)
- `cargoValue`: Total cargo value in gold (must be positive)

**Auto-Calculated Metrics:**
- `ProfitMargin`: 10-50% based on danger level
- `DangerLevel`: 0.0-1.0 based on server hop count
- `TravelTime`: 30 minutes per region hop
- `ServerHops`: 1-3 hops based on region hash (simplified pathfinding)

### Caravan Generation Parameters

```go
caravan, err := rm.CreateCaravan(seed int64, genreID string)
```

**Parameters:**
- `seed`: Deterministic seed for vehicle generation
- `genreID`: Genre theme (`"fantasy"`, `"sci-fi"`, `"horror"`, `"cyberpunk"`, `"post-apocalyptic"`)

**Generated Vehicle:**
- Type: `vehicle.TypeCart` (cargo transport)
- Rarity: `vehicle.RarityCommon`
- Count: 1 vehicle per call

---

## Player Escort System

### Creating Escort Missions

```go
mission, err := rm.CreateEscortMission(routeID string, playerID uint64, baseReward float64)
```

**Parameters:**
- `routeID`: Active route to protect
- `playerID`: Entity ID of accepting player
- `baseReward`: Base gold payment (modified by danger level)

**Reward Calculation:**
```go
actualReward = baseReward * (1.0 + dangerLevel)       // Travel completion
bonusReward = actualReward * 0.5                       // Defeating bandit attacks
```

**Example:**
```go
baseReward := 50.0 // 50 gold base
mission, err := rm.CreateEscortMission(route.ID, playerEntityID, baseReward)

// If route has dangerLevel = 0.6 (high risk):
// - Completion reward: 50 * (1.0 + 0.6) = 80 gold
// - Bonus (if attacks defeated): 80 * 0.5 = 40 gold
// - Total potential: 120 gold
```

### Adding Player Escorts

```go
err := rm.AddEscort(routeID string, playerID uint64)
```

**Effects:**
- Increases route defense strength by 0.15 per player
- Player receives combat notifications for bandit encounters
- Mission rewards paid on route completion
- Players can escort multiple routes simultaneously

### Removing Player Escorts

```go
err := rm.RemoveEscort(routeID string, playerID uint64)
```

Use when player disconnects, abandons mission, or completes escort.

---

## Route Monitoring and Optimization

### Querying Route Status

```go
// Get specific route
route, err := rm.GetRoute(routeID string)

// Get all active routes
activeRoutes := rm.GetActiveRoutes() // []*TradeRoute

// Get route by caravan vehicle ID
route, err := rm.GetRouteByCaravan(caravanID uint64)
```

**Route Status Values:**
- `StatusPlanning`: Route created, not yet started
- `StatusActive`: Caravan traveling, progress updating
- `StatusUnderAttack`: Bandit encounter in progress
- `StatusCompleted`: Successful delivery
- `StatusFailed`: Caravan destroyed or critical cargo loss
- `StatusCancelled`: Manually terminated

### Route Optimization Analysis

```go
optimization := rm.OptimizeRoute(route *TradeRoute)
```

**Returns `RouteOptimization` with:**
```go
type RouteOptimization struct {
    ExpectedProfit      float64       // Total cargo profit
    RiskAdjustedProfit  float64       // Profit adjusted for danger
    ServerHops          int           // Cross-server transitions
    EstimatedTravelTime time.Duration // Real-time duration
    DangerZones         []DangerZone  // High-risk segments
    AlternateRoutes     []*TradeRoute // Safer/slower options (future)
}
```

**Example Usage:**
```go
opt := rm.OptimizeRoute(route)

if opt.RiskAdjustedProfit < opt.ExpectedProfit*0.5 {
    log.Warn("High-risk route: consider adding escorts")
}

for _, zone := range opt.DangerZones {
    log.WithFields(logrus.Fields{
        "region":             zone.RegionID,
        "danger":             zone.DangerLevel,
        "bandit_spawn_rate":  zone.BanditSpawnRate,
        "recommended_escorts": zone.RecommendedEscorts,
    }).Info("Danger zone identified")
}
```

---

## Bandit Encounter Mechanics

### Spawn Calculation

Bandits spawn probabilistically based on danger level:

```go
spawnRate := dangerLevel * 1.0    // attacks per hour (0.0-1.0)
chancePer10Sec := spawnRate / 360.0 // 360 * 10sec = 1 hour
```

**Examples:**
- `dangerLevel = 0.2` → 0.2 attacks/hour → ~1 attack per 5 hours
- `dangerLevel = 0.6` → 0.6 attacks/hour → ~1 attack per 100 minutes
- `dangerLevel = 1.0` → 1.0 attacks/hour → ~1 attack per hour

### Combat Resolution

Encounters resolve after `duration` (30-120 seconds) based on strength comparison:

```go
defenseStrength = 0.4 + (escortCount * 0.15)  // Caravan base + escorts
banditStrength = 0.3 + (dangerLevel * 0.4)    // 0.3-0.7 range
```

**Outcomes:**
1. **Defended** (`defense > bandits`):
   - 100% cargo retained
   - Escort players receive base reward (100 gold)
   - Route continues normally

2. **Compromised** (`defense 70-100% of bandits`):
   - 20-50% cargo loss (random items stolen)
   - Route continues with reduced cargo
   - Reduced profit at destination

3. **Destroyed** (`defense <70% of bandits`):
   - 100% cargo loss
   - Route status → `StatusFailed`
   - Caravan vehicle destroyed
   - Escort missions fail

### Monitoring Encounters

```go
// Encounters are stored in RouteManager
for encounterID, encounter := range rm.encounters {
    log.WithFields(logrus.Fields{
        "encounter_id":     encounterID,
        "route_id":         encounter.RouteID,
        "bandit_count":     encounter.BanditCount,
        "bandit_strength":  encounter.BanditStrength,
        "defense_strength": encounter.DefenseStrength,
        "outcome":          encounter.Outcome.String(),
        "lost_cargo":       len(encounter.LostCargo),
    }).Info("Bandit encounter")
}
```

---

## Guild Sponsorship System

Guild sponsorship allows guilds to fund trade routes for regional price manipulation (integrated with federation market system).

### Creating Guild Sponsorships

```go
type GuildSponsorship struct {
    GuildID           string
    RouteID           string
    FundingAmount     float64
    TargetPriceChange float64  // -0.5 to +0.5 (±50% price change)
    ActiveRoutes      int
    TotalProfit       float64
    StartDate         time.Time
    EndDate           time.Time
}
```

**Usage Pattern (Future Implementation):**
```go
// Note: Guild sponsorship is defined but not fully wired to market system yet
sponsorship := &trade_routes.GuildSponsorship{
    GuildID:           "guild-123",
    RouteID:           route.ID,
    FundingAmount:     5000.0,
    TargetPriceChange: -0.2, // Decrease regional prices by 20%
    ActiveRoutes:      1,
    StartDate:         time.Now(),
    EndDate:           time.Now().Add(7 * 24 * time.Hour), // 1 week
}

// Integration with pkg/network/federation/market.go required
// for actual price manipulation
```

---

## Performance and Resource Management

### Memory Management

**Per-Route Memory Footprint:**
- Route struct: ~500 bytes
- Cargo items (3-8): ~400 bytes
- Encounter data: ~300 bytes
- Metadata/maps: ~800 bytes
- **Total:** ~2KB per active route

**50 Active Routes:** ~100KB memory

### Background Processing

The `RouteManager.Start()` method creates a background goroutine that updates all routes every 10 seconds:

```go
// Update cycle (every 10 seconds):
// 1. Update route progress (elapsed time / travel time)
// 2. Check for route completion (progress >= 1.0)
// 3. Spawn bandit encounters (probabilistic)
// 4. Resolve ongoing encounters (after duration)
// 5. Update route status (active/completed/failed)
```

**Performance Characteristics:**
- Update 50 routes: <1ms total
- Encounter spawning: <0.1ms per check
- Combat resolution: <0.2ms per encounter

### Graceful Shutdown

**Always call `Stop()` to prevent goroutine leaks:**

```go
rm := trade_routes.NewRouteManager("server-01", 12345)
rm.Start()
defer rm.Stop() // Required! Prevents goroutine + ticker leak

// Server runs...
```

**Shutdown Behavior:**
- Stops ticker (no new updates)
- Closes `stopChan` (signals goroutine termination)
- Goroutine exits cleanly within 10 seconds (max wait time)
- Active routes remain in memory (persist if needed)

**Note:** Once `Stop()` is called, the manager cannot be restarted. Create a new instance if needed.

---

## Integration with Other Systems

### V4 Vehicles (`pkg/procgen/vehicle`)

Trade routes use the vehicle generator to create cart-type caravans:

```go
caravan, err := rm.CreateCaravan(seed, genreID)
// Internally calls:
// - vehicle.NewVehicleGenerator()
// - gen.Generate(seed, params) with TypeCart + RarityCommon
```

**Genre-Specific Variants:**
- **Fantasy:** Horse-drawn wagons, merchant carts
- **Sci-Fi:** Hover transports, cargo drones
- **Cyberpunk:** Corporate delivery trucks, armored convoys
- **Post-Apocalyptic:** Scrap metal vehicles, fortified trucks

### V6 Federation Market (`pkg/network/federation/market.go`)

Trade routes integrate with dynamic pricing for cross-server economies:

```go
// Route cargo uses current market prices (future integration)
// - PurchasePrice: Origin region market price
// - TargetPrice: Destination region market price
// - Profit: (TargetPrice - PurchasePrice) * Quantity
```

**Pending Integration:**
- Guild sponsorships affecting regional supply/demand
- Route completion updating market prices
- Cross-server price arbitrage detection

### V4 AI System (`pkg/engine/ai_system.go`)

AI-controlled merchant NPCs use pathfinding for route navigation:

```go
// Future: Merchant AI integration for dynamic route selection
// - AI evaluates profitable routes
// - Pathfinding avoids high-danger regions
// - Adaptive cargo selection based on market trends
```

### V8 Guild System (`pkg/engine/guild_system.go`)

Guild features integrate with trade routes:

```go
// Implemented:
// - Guild bank funding for sponsorships
// - Guild-wide escort mission boards
// - Shared profits from sponsored routes

// Pending:
// - Guild fleet management (multiple sponsored routes)
// - Cross-guild trade agreements
// - Guild reputation from successful routes
```

---

## Example: Complete Server Integration

```go
package main

import (
    "flag"
    "log"
    "math/rand"
    "time"
    
    "github.com/opd-ai/venture/pkg/integration/trade_routes"
    "github.com/sirupsen/logrus"
)

var (
    enableTradeRoutes = flag.Bool("enable-trade-routes", false, "Enable automated trade routes")
    targetRoutes      = flag.Int("trade-route-count", 20, "Target active routes")
    seed              = flag.Int64("seed", 12345, "World seed")
    genreID           = flag.String("genre", "fantasy", "Genre ID")
)

func main() {
    flag.Parse()
    
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    
    if !*enableTradeRoutes {
        logger.Info("Trade routes disabled (use -enable-trade-routes)")
        return
    }
    
    // Initialize route manager
    rm := trade_routes.NewRouteManager("server-01", *seed)
    rm.Start()
    defer rm.Stop()
    
    logger.WithFields(logrus.Fields{
        "system":        "trade_routes",
        "target_routes": *targetRoutes,
        "genre":         *genreID,
    }).Info("Trade routes system activated")
    
    // Spawn initial routes
    regions := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
    for i := 0; i < *targetRoutes; i++ {
        spawnRoute(rm, regions, *genreID, logger)
    }
    
    // Monitor and maintain route count
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        active := rm.GetActiveRoutes()
        logger.WithField("active_routes", len(active)).Info("Route status check")
        
        if len(active) < *targetRoutes {
            toSpawn := *targetRoutes - len(active)
            for i := 0; i < toSpawn; i++ {
                spawnRoute(rm, regions, *genreID, logger)
            }
        }
    }
}

func spawnRoute(rm *trade_routes.RouteManager, regions []string, genreID string, logger *logrus.Logger) {
    start := regions[rand.Intn(len(regions))]
    end := regions[rand.Intn(len(regions))]
    if start == end {
        return
    }
    
    cargoValue := 500.0 + rand.Float64()*2000.0
    
    route, err := rm.CreateRoute(start, end, cargoValue)
    if err != nil {
        logger.WithError(err).Error("Failed to create route")
        return
    }
    
    caravan, err := rm.CreateCaravan(rand.Int63(), genreID)
    if err != nil {
        logger.WithError(err).Error("Failed to create caravan")
        return
    }
    
    if err := rm.StartRoute(route.ID, caravan.ID); err != nil {
        logger.WithError(err).Error("Failed to start route")
        return
    }
    
    logger.WithFields(logrus.Fields{
        "route_id":      route.ID,
        "route_name":    route.Name,
        "start":         start,
        "end":           end,
        "cargo_value":   cargoValue,
        "profit_margin": route.ProfitMargin,
        "danger_level":  route.DangerLevel,
        "travel_time":   route.TravelTime,
    }).Info("Trade route spawned")
}
```

**Run Command:**
```bash
go run server.go -enable-trade-routes -trade-route-count 30 -seed 98765 -genre fantasy
```

---

## Testing and Validation

### Unit Tests

Run the trade routes test suite:

```bash
go test -v ./pkg/integration/trade_routes/
```

**Coverage:** ≥65% (as of 2026-02-07)

### Integration Testing

```go
func TestRouteLifecycle(t *testing.T) {
    rm := trade_routes.NewRouteManager("test-server", 12345)
    rm.Start()
    defer rm.Stop()
    
    // Create route
    route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
    if err != nil {
        t.Fatalf("CreateRoute failed: %v", err)
    }
    
    // Generate caravan
    caravan, err := rm.CreateCaravan(54321, "fantasy")
    if err != nil {
        t.Fatalf("CreateCaravan failed: %v", err)
    }
    
    // Start route
    err = rm.StartRoute(route.ID, caravan.ID)
    if err != nil {
        t.Fatalf("StartRoute failed: %v", err)
    }
    
    // Verify active status
    active := rm.GetActiveRoutes()
    if len(active) != 1 {
        t.Errorf("Expected 1 active route, got %d", len(active))
    }
    
    // Simulate progress (wait for update tick)
    time.Sleep(11 * time.Second)
    
    retrievedRoute, _ := rm.GetRoute(route.ID)
    if retrievedRoute.Progress == 0.0 {
        t.Error("Route progress not updated")
    }
}
```

### Performance Benchmarks

```bash
go test -bench=. -benchmem ./pkg/integration/trade_routes/
```

**Expected Results:**
- Route creation: <1ms/op, <2KB alloc
- Route updates (50 routes): <1ms/op, <5KB alloc
- Optimization: <0.5ms/op, <1KB alloc

---

## Troubleshooting

### Issue: Goroutine Leak After Server Shutdown

**Symptom:** Server process doesn't exit cleanly  
**Cause:** `RouteManager.Stop()` not called  
**Fix:**
```go
rm := trade_routes.NewRouteManager("server-01", 12345)
rm.Start()
defer rm.Stop() // Add this line
```

### Issue: No Bandit Encounters Spawning

**Symptom:** Routes complete without attacks  
**Cause:** Low danger level or insufficient time  
**Debug:**
```go
route, _ := rm.GetRoute(routeID)
log.WithFields(logrus.Fields{
    "danger_level":    route.DangerLevel,
    "expected_rate":   route.DangerLevel * 1.0, // attacks per hour
    "bandit_attacks":  route.BanditAttacks,
}).Info("Encounter debug")
```

**Fix:** Create longer routes or increase region hop count (affects danger calculation)

### Issue: Routes Fail Immediately After Start

**Symptom:** Route status changes to `StatusFailed` without progress  
**Cause:** Caravan ID collision or invalid caravan entity  
**Debug:**
```go
route, _ := rm.GetRoute(routeID)
log.WithFields(logrus.Fields{
    "caravan_id": route.CaravanID,
    "status":     route.Status.String(),
    "progress":   route.Progress,
}).Error("Route failure")
```

**Fix:** Ensure unique caravan IDs (don't reuse entity IDs)

### Issue: High Memory Usage with Many Routes

**Symptom:** Server memory grows beyond expected levels  
**Cause:** Completed routes not cleaned up  
**Debug:**
```go
log.WithFields(logrus.Fields{
    "total_routes":     len(rm.routes),
    "active_routes":    len(rm.GetActiveRoutes()),
    "encounters":       len(rm.encounters),
}).Info("Route memory stats")
```

**Fix:** Periodically clean up completed routes:
```go
// Cleanup completed routes (run periodically)
for id, route := range rm.routes {
    if route.Status == trade_routes.StatusCompleted || 
       route.Status == trade_routes.StatusFailed {
        // Archive route data if needed, then delete
        delete(rm.routes, id)
    }
}
```

---

## Roadmap and Future Enhancements

### Planned Features

1. **Alternate Route Pathfinding** (Phase 58.0)
   - Graph-based pathfinding for multiple route options
   - Safety vs. speed tradeoffs
   - Dynamic rerouting on bandit encounters

2. **Market Price Integration** (Phase 59.0)
   - Real-time market price updates from federation
   - Guild sponsorships affecting supply/demand
   - Cross-server price arbitrage detection

3. **Merchant AI Personalities** (Phase 60.0)
   - Risk-averse vs. risk-seeking merchants
   - Reputation system for reliable caravans
   - Adaptive cargo selection based on market trends

4. **Player-Owned Caravans** (Phase 61.0)
   - Players can purchase and customize caravans
   - Hire NPC guards instead of player escorts
   - Multiple simultaneous routes per player

### Known Limitations

- **Route Pathfinding:** Currently uses simplified hash-based hop calculation (1-3 hops). Full graph-based pathfinding planned for Phase 58.0.
- **Market Integration:** Cargo prices are calculated at route creation time. Real-time market updates not yet implemented.
- **Guild Sponsorships:** Defined but not fully wired to federation market system. Price manipulation mechanics planned for Phase 59.0.
- **Alternate Routes:** `AlternateRoutes` field in `RouteOptimization` currently returns empty slice (future implementation).

---

## API Reference Summary

### RouteManager Methods

| Method | Parameters | Returns | Description |
|--------|-----------|---------|-------------|
| `NewRouteManager` | `serverID string, seed int64` | `*RouteManager` | Create manager instance |
| `Start` | - | - | Begin background processing |
| `Stop` | - | - | Halt processing, cleanup |
| `CreateRoute` | `start, end string, value float64` | `*TradeRoute, error` | Create new route |
| `StartRoute` | `routeID string, caravanID uint64` | `error` | Activate route |
| `GetRoute` | `routeID string` | `*TradeRoute, error` | Retrieve route by ID |
| `GetActiveRoutes` | - | `[]*TradeRoute` | List active routes |
| `GetRouteByCaravan` | `caravanID uint64` | `*TradeRoute, error` | Find route by caravan |
| `OptimizeRoute` | `route *TradeRoute` | `*RouteOptimization` | Calculate metrics |
| `AddEscort` | `routeID string, playerID uint64` | `error` | Add player escort |
| `RemoveEscort` | `routeID string, playerID uint64` | `error` | Remove escort |
| `CreateEscortMission` | `routeID string, playerID uint64, reward float64` | `*EscortMission, error` | Create mission |
| `CreateCaravan` | `seed int64, genreID string` | `*vehicle.Vehicle, error` | Generate caravan |
| `UpdateRoutes` | - | - | Manual update trigger |

### Key Data Structures

```go
type TradeRoute struct {
    ID               string
    Name             string
    StartRegion      string
    EndRegion        string
    CaravanID        uint64
    Status           RouteStatus
    Cargo            []CargoItem
    ProfitMargin     float64
    DangerLevel      float64
    Progress         float64
    TravelTime       time.Duration
    StartTime        time.Time
    EstimatedArrival time.Time
    EscortPlayers    []uint64
    BanditAttacks    int
    SuccessRate      float64
}

type CargoItem struct {
    ItemID        string
    ItemName      string
    Quantity      int
    PurchasePrice float64
    TargetPrice   float64
    Profit        float64
}

type BanditEncounter struct {
    ID              string
    RouteID         string
    BanditCount     int
    BanditStrength  float64
    DefenseStrength float64
    StartTime       time.Time
    Duration        time.Duration
    Outcome         EncounterOutcome
    LostCargo       []CargoItem
    PlayerRewards   map[uint64]float64
}
```

---

## Additional Resources

- **Package Documentation:** `pkg/integration/trade_routes/doc.go`
- **Test Suite:** `pkg/integration/trade_routes/manager_test.go`
- **Integration Examples:** `pkg/integration/trade_routes/AUDIT.md`
- **Vehicle System:** `docs/VEHICLES.md` (if exists)
- **Federation Market:** `docs/FEDERATION_MARKET.md` (if exists)
- **Guild System:** `docs/SOCIAL_SYSTEMS.md`

---

## Support and Contribution

For questions, issues, or feature requests related to the trade routes system:

1. Check this documentation for configuration guidance
2. Review the test suite for usage examples
3. Examine `pkg/integration/trade_routes/doc.go` for package-level details
4. Submit issues with reproduction steps and debug logs

**Example Debug Logging:**
```go
logger.SetLevel(logrus.DebugLevel)

rm := trade_routes.NewRouteManager("server-01", 12345)
rm.Start()
defer rm.Stop()

// Enable verbose route logging
for {
    routes := rm.GetActiveRoutes()
    for _, route := range routes {
        logger.WithFields(logrus.Fields{
            "route_id":      route.ID,
            "status":        route.Status.String(),
            "progress":      route.Progress,
            "escorts":       len(route.EscortPlayers),
            "bandit_attacks": route.BanditAttacks,
        }).Debug("Route status")
    }
    time.Sleep(10 * time.Second)
}
```

---

**Document Version:** 1.0  
**Last Reviewed:** 2026-02-07  
**Maintainer:** Venture Development Team
