// Package trade_routes implements automated AI merchant caravan systems for cross-server trading.
//
// Phase 57.3: Automated Trade Routes
//
// # Overview
//
// This package provides a complete automated trading system where AI-controlled merchant caravans
// transport goods between regions/servers, creating dynamic economic opportunities and player
// engagement through escort missions and bandit encounters.
//
// # Core Features
//
//   - AI Merchant Caravans: NPC-controlled vehicle fleets automatically trade between regions
//   - Route Optimization: Calculates profitable trade paths considering danger and distance
//   - Risk/Reward Mechanics: Dangerous routes offer higher profits (10-50% margins)
//   - Player Escort Missions: Protect caravans for gold rewards and bonus for defeating attacks
//   - Procedural Bandit Attacks: Dynamic encounters threaten shipments (0.1-1.0 per hour)
//   - Guild Sponsorship: Fund caravans to manipulate regional prices
//
// # Usage Example
//
//	// Create a route manager
//	rm := trade_routes.NewRouteManager("server-01", 12345)
//	rm.Start() // Begin background processing
//	defer rm.Stop()
//
//	// Create a new trade route
//	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Generate a caravan vehicle
//	caravan, err := rm.CreateCaravan(54321, "fantasy")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start the route
//	err = rm.StartRoute(route.ID, caravan.ID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Add player escorts
//	err = rm.AddEscort(route.ID, playerEntityID)
//
//	// Create escort mission for players
//	mission, err := rm.CreateEscortMission(route.ID, playerEntityID, 50.0)
//
//	// Optimize route for profitability
//	optimization := rm.OptimizeRoute(route)
//	fmt.Printf("Expected profit: %.2f gold\n", optimization.ExpectedProfit)
//	fmt.Printf("Danger zones: %d\n", len(optimization.DangerZones))
//
// # Route Lifecycle
//
//  1. Planning: Route created with cargo and pricing calculated
//  2. Active: Caravan travels, progress updates every 10 seconds
//  3. Under Attack: Bandit encounter spawned based on danger level
//  4. Completed: Route finished, profits distributed
//  5. Failed: Caravan destroyed or cargo lost
//  6. Cancelled: Manually cancelled by system/player
//
// # Encounter Resolution
//
// Bandit encounters are resolved based on relative strength:
//
//   - Defense > Bandits: Defended (100% cargo retained, player rewards)
//   - Defense 70-100% of Bandits: Compromised (20-50% cargo loss)
//   - Defense <70% of Bandits: Destroyed (100% cargo loss, route failed)
//
// Defense strength = 0.4 (caravan base) + 0.15 per escort player
//
// # Performance Characteristics
//
//   - Route creation: <1ms per route
//   - Route update: <1ms for 50 active routes
//   - Optimization: <0.5ms per route
//   - Memory: ~2KB per active route
//   - Network bandwidth: Integrated with existing federation protocol
//
// # Integration Points
//
//   - V4 Vehicles: pkg/procgen/vehicle (caravan vehicle generation)
//   - V6 Federation Market: pkg/network/federation/market.go (dynamic pricing)
//   - V4 AI: pkg/engine/ai_system.go (merchant pathfinding)
//   - V6 Politics: Danger levels affected by faction relationships
//   - V8 Guilds: Guild sponsorship for price manipulation
//
// # Design Constraints
//
//   - Deterministic cargo generation (same seed = same cargo)
//   - Non-deterministic encounter spawning (based on real-time probabilities)
//   - Server-authoritative (prevents client-side manipulation)
//   - Thread-safe (all methods use RWMutex)
//
// # Success Metrics (from ROADMAP_V9.md)
//
//   - Active routes: 10-50 per server
//   - Caravan speed: 1 region per 30 minutes real-time
//   - Success rate: 70-90% (higher with player escorts)
//   - Profit margin: 10-50% on successful deliveries
//   - Test coverage: ≥65%
package trade_routes
