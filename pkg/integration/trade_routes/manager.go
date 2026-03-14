// File: manager.go
// Purpose: RouteManager implementation for trade route management
//
// This file implements the core trade route management logic:
// - RouteManager: Central manager for all trade routes
// - Route lifecycle: Creation, activation, progress tracking, completion
// - Bandit encounter system: Spawning, combat resolution, outcomes
// - Escort missions: Player protection quests with rewards
// - Route optimization: Danger zones, profit margins, travel time
// - Caravan creation: Integration with vehicle generator
//
// All route updates run on a background goroutine (1-second tick rate).
package trade_routes

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
)

// RouteManager manages all active trade routes and caravan fleets.
//
// # Lifecycle and Cleanup
//
// RouteManager uses a goroutine-based update pattern where a background goroutine
// periodically updates routes every 10 seconds. The goroutine is started by Start()
// and terminated by closing stopChan via Stop(). This design provides graceful
// shutdown control and allows the manager to clean up resources properly.
//
// Start() is idempotent - calling it multiple times will only start one background
// goroutine. Stop() is also idempotent and safe to call multiple times.
//
// Callers MUST ensure Stop() is called when the manager is no longer needed,
// ideally using defer immediately after Start():
//
//	rm := NewRouteManager("server-1", 12345)
//	rm.Start()
//	defer rm.Stop() // Ensures cleanup even if panic occurs
//
//	// Use the route manager...
//	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
//
// Failure to call Stop() will result in a goroutine leak, as the background update
// loop will continue running indefinitely. The ticker will also not be properly
// cleaned up, leading to resource waste.
//
// Note: Once Stop() is called, the manager cannot be restarted. Create a new
// RouteManager instance if you need to start fresh.
type RouteManager struct {
	mu              sync.RWMutex
	routes          map[string]*TradeRoute       // RouteID -> Route
	encounters      map[string]*BanditEncounter  // EncounterID -> Encounter
	missions        map[string]*EscortMission    // MissionID -> Mission
	sponsorships    map[string]*GuildSponsorship // GuildID -> Sponsorship
	activeCaravans  map[uint64]*TradeRoute       // CaravanID -> Route
	serverID        string
	nextRouteID     int
	nextEncounterID int
	nextMissionID   int
	rng             *rand.Rand
	updateTicker    *time.Ticker
	stopChan        chan struct{}      // Closed by Stop() to signal goroutine termination
	startOnce       sync.Once          // Ensures Start() is idempotent
	stopOnce        sync.Once          // Ensures Stop() is idempotent
	priceHandler    PriceUpdateHandler // Optional: economy system for price updates
}

// NewRouteManager creates a new route manager instance.
func NewRouteManager(serverID string, seed int64) *RouteManager {
	return &RouteManager{
		routes:          make(map[string]*TradeRoute),
		encounters:      make(map[string]*BanditEncounter),
		missions:        make(map[string]*EscortMission),
		sponsorships:    make(map[string]*GuildSponsorship),
		activeCaravans:  make(map[uint64]*TradeRoute),
		serverID:        serverID,
		nextRouteID:     1,
		nextEncounterID: 1,
		nextMissionID:   1,
		rng:             rand.New(rand.NewSource(seed)),
		stopChan:        make(chan struct{}),
	}
}

// Start begins the route update loop (update every 10 seconds).
//
// This method creates a ticker and starts a background goroutine that periodically
// calls UpdateRoutes(). The goroutine will run until Stop() is called, which closes
// stopChan to signal termination.
//
// Start is idempotent - calling it multiple times will only start one background
// goroutine. This prevents goroutine leaks from accidental multiple Start() calls.
//
// Callers must ensure Stop() is called when the manager is no longer needed to
// prevent goroutine and ticker resource leaks. See RouteManager documentation for
// recommended usage patterns.
func (rm *RouteManager) Start() {
	// Use sync.Once to ensure only one goroutine is started
	rm.startOnce.Do(func() {
		rm.updateTicker = time.NewTicker(10 * time.Second)
		// Start background update loop.
		// This goroutine MUST be terminated by calling Stop(), which closes stopChan.
		// The select pattern ensures clean termination when stopChan is closed.
		go func() {
			for {
				select {
				case <-rm.updateTicker.C:
					rm.UpdateRoutes()
				case <-rm.stopChan:
					return
				}
			}
		}()
	})
}

// Stop halts the route update loop and cleans up resources.
//
// This method stops the ticker and closes stopChan to signal the background goroutine
// to exit. Stop is idempotent and safe to call multiple times (subsequent calls do nothing
// since stopChan is already closed).
//
// Callers should defer Stop() after Start() to ensure cleanup:
//
//	rm := NewRouteManager("server-1", 12345)
//	rm.Start()
//	defer rm.Stop() // Ensures cleanup even on panic
//
// Note: This method does not wait for the goroutine to exit (unlike some manager
// implementations with WaitGroups). The goroutine will exit quickly once stopChan
// is closed, as it's only performing periodic route updates.
//
// Once Stop() is called, the manager cannot be restarted. Create a new RouteManager
// instance if you need to start fresh.
func (rm *RouteManager) Stop() {
	// Use sync.Once to ensure cleanup happens exactly once
	rm.stopOnce.Do(func() {
		if rm.updateTicker != nil {
			rm.updateTicker.Stop()
		}
		close(rm.stopChan)
	})
}

// SetPriceUpdateHandler sets the economy system for price updates.
// This enables trade route completion to influence marketplace prices.
// handler: Implementation of PriceUpdateHandler (typically economy.System)
func (rm *RouteManager) SetPriceUpdateHandler(handler PriceUpdateHandler) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.priceHandler = handler
}

// CreateRoute creates a new trade route with optimized cargo and path.
func (rm *RouteManager) CreateRoute(startRegion, endRegion string, cargoValue float64) (*TradeRoute, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if startRegion == "" || endRegion == "" {
		log.WithFields(log.Fields{
			"startRegion": startRegion,
			"endRegion":   endRegion,
			"serverID":    rm.serverID,
		}).Warn("CreateRoute: start and end regions cannot be empty")
		return nil, fmt.Errorf("start and end regions cannot be empty")
	}

	if startRegion == endRegion {
		log.WithFields(log.Fields{
			"startRegion": startRegion,
			"endRegion":   endRegion,
			"serverID":    rm.serverID,
		}).Warn("CreateRoute: start and end regions must be different")
		return nil, fmt.Errorf("start and end regions must be different")
	}

	if cargoValue <= 0 {
		log.WithFields(log.Fields{
			"startRegion": startRegion,
			"endRegion":   endRegion,
			"cargoValue":  cargoValue,
			"serverID":    rm.serverID,
		}).Warn("CreateRoute: cargo value must be positive")
		return nil, fmt.Errorf("cargo value must be positive, got %.2f", cargoValue)
	}

	// Generate unique route ID
	routeID := fmt.Sprintf("%s-route-%d", rm.serverID, rm.nextRouteID)
	rm.nextRouteID++

	// Calculate route metrics
	serverHops := rm.calculateServerHops(startRegion, endRegion)
	travelTime := time.Duration(serverHops*30) * time.Minute
	dangerLevel := rm.calculateDangerLevel(startRegion, endRegion)
	profitMargin := rm.calculateProfitMargin(cargoValue, dangerLevel)

	// Create the route
	route := &TradeRoute{
		ID:               routeID,
		Name:             rm.generateRouteName(startRegion, endRegion),
		StartRegion:      startRegion,
		EndRegion:        endRegion,
		Status:           StatusPlanning,
		Cargo:            rm.generateCargo(cargoValue, profitMargin),
		ProfitMargin:     profitMargin,
		DangerLevel:      dangerLevel,
		Progress:         0.0,
		TravelTime:       travelTime,
		StartTime:        now(),
		EstimatedArrival: now().Add(travelTime),
		EscortPlayers:    []uint64{},
		BanditAttacks:    0,
		SuccessRate:      0.75, // Historical baseline
	}

	rm.routes[routeID] = route
	return route, nil
}

// StartRoute activates a trade route and spawns the caravan.
func (rm *RouteManager) StartRoute(routeID string, caravanID uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	route, exists := rm.routes[routeID]
	if !exists {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"serverID": rm.serverID,
		}).Warn("StartRoute: route not found")
		return fmt.Errorf("route not found: %s", routeID)
	}

	if route.Status != StatusPlanning {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"status":   route.Status,
			"serverID": rm.serverID,
		}).Warn("StartRoute: route must be in planning status")
		return fmt.Errorf("route must be in planning status, got %s", route.Status)
	}

	if caravanID == 0 {
		log.WithFields(log.Fields{
			"routeID":   routeID,
			"caravanID": caravanID,
			"serverID":  rm.serverID,
		}).Warn("StartRoute: caravan ID cannot be zero")
		return fmt.Errorf("caravan ID cannot be zero")
	}

	route.CaravanID = caravanID
	route.Status = StatusActive
	route.StartTime = now()
	route.EstimatedArrival = now().Add(route.TravelTime)

	rm.activeCaravans[caravanID] = route
	return nil
}

// UpdateRoutes processes all active routes and spawns encounters.
func (rm *RouteManager) UpdateRoutes() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	currentTime := now()

	for _, route := range rm.routes {
		if route.Status != StatusActive {
			continue
		}

		// Update progress
		elapsed := currentTime.Sub(route.StartTime)
		route.Progress = math.Min(1.0, float64(elapsed)/float64(route.TravelTime))

		// Check for completion
		if route.Progress >= 1.0 {
			route.Status = StatusCompleted
			rm.completeRoute(route)
			continue
		}

		// Spawn bandit encounters based on danger level
		if rm.shouldSpawnBandit(route, currentTime) {
			rm.spawnBanditEncounter(route, currentTime)
		}
	}

	// Update active encounters
	for _, encounter := range rm.encounters {
		if encounter.Outcome == OutcomePending {
			rm.updateEncounter(encounter, currentTime)
		}
	}
}

// AddEscort adds a player to the route's escort list.
func (rm *RouteManager) AddEscort(routeID string, playerID uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if playerID == 0 {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"serverID": rm.serverID,
		}).Warn("AddEscort: player ID cannot be zero")
		return fmt.Errorf("player ID cannot be zero")
	}

	route, exists := rm.routes[routeID]
	if !exists {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"serverID": rm.serverID,
		}).Warn("AddEscort: route not found")
		return fmt.Errorf("route not found: %s", routeID)
	}

	if route.Status != StatusActive {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"status":   route.Status,
			"serverID": rm.serverID,
		}).Warn("AddEscort: can only escort active routes")
		return fmt.Errorf("can only escort active routes, got status: %s", route.Status)
	}

	// Check if player is already escorting
	for _, escort := range route.EscortPlayers {
		if escort == playerID {
			log.WithFields(log.Fields{
				"routeID":  routeID,
				"playerID": playerID,
				"serverID": rm.serverID,
			}).Warn("AddEscort: player already escorting")
			return fmt.Errorf("player %d is already escorting this route", playerID)
		}
	}

	route.EscortPlayers = append(route.EscortPlayers, playerID)
	return nil
}

// RemoveEscort removes a player from the route's escort list.
func (rm *RouteManager) RemoveEscort(routeID string, playerID uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	route, exists := rm.routes[routeID]
	if !exists {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"serverID": rm.serverID,
		}).Warn("RemoveEscort: route not found")
		return fmt.Errorf("route not found: %s", routeID)
	}

	for i, escort := range route.EscortPlayers {
		if escort == playerID {
			route.EscortPlayers = append(route.EscortPlayers[:i], route.EscortPlayers[i+1:]...)
			return nil
		}
	}

	log.WithFields(log.Fields{
		"routeID":  routeID,
		"playerID": playerID,
		"serverID": rm.serverID,
	}).Warn("RemoveEscort: player not escorting route")
	return fmt.Errorf("player %d is not escorting route %s", playerID, routeID)
}

// GetRoute retrieves a route by ID.
func (rm *RouteManager) GetRoute(routeID string) (*TradeRoute, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	route, exists := rm.routes[routeID]
	if !exists {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"serverID": rm.serverID,
		}).Warn("GetRoute: route not found")
		return nil, fmt.Errorf("route not found: %s", routeID)
	}

	return route, nil
}

// GetActiveRoutes returns all currently active routes.
func (rm *RouteManager) GetActiveRoutes() []*TradeRoute {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	active := []*TradeRoute{}
	for _, route := range rm.routes {
		if route.Status == StatusActive {
			active = append(active, route)
		}
	}

	return active
}

// GetRouteByCaravan retrieves the route for a specific caravan.
func (rm *RouteManager) GetRouteByCaravan(caravanID uint64) (*TradeRoute, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	route, exists := rm.activeCaravans[caravanID]
	if !exists {
		log.WithFields(log.Fields{
			"caravanID": caravanID,
			"serverID":  rm.serverID,
		}).Warn("GetRouteByCaravan: no active route for caravan")
		return nil, fmt.Errorf("no active route for caravan %d", caravanID)
	}

	return route, nil
}

// OptimizeRoute calculates the most profitable path and cargo configuration.
func (rm *RouteManager) OptimizeRoute(route *TradeRoute) *RouteOptimization {
	if route == nil {
		return nil
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Calculate expected profit
	expectedProfit := 0.0
	for _, item := range route.Cargo {
		expectedProfit += item.Profit * float64(item.Quantity)
	}

	// Adjust for danger level
	riskAdjustedProfit := expectedProfit * (1.0 - route.DangerLevel*0.5)

	// Calculate danger zones
	dangerZones := rm.identifyDangerZones(route)

	return &RouteOptimization{
		Route:               route,
		ExpectedProfit:      expectedProfit,
		RiskAdjustedProfit:  riskAdjustedProfit,
		ServerHops:          rm.calculateServerHops(route.StartRegion, route.EndRegion),
		EstimatedTravelTime: route.TravelTime,
		DangerZones:         dangerZones,
		AlternateRoutes:     []*TradeRoute{}, // Future: implement alternate path finding
	}
}

// CreateEscortMission creates a player mission to protect a caravan.
func (rm *RouteManager) CreateEscortMission(routeID string, playerID uint64, baseReward float64) (*EscortMission, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if playerID == 0 {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"serverID": rm.serverID,
		}).Warn("CreateEscortMission: player ID cannot be zero")
		return nil, fmt.Errorf("player ID cannot be zero")
	}

	if baseReward <= 0 {
		log.WithFields(log.Fields{
			"routeID":    routeID,
			"playerID":   playerID,
			"baseReward": baseReward,
			"serverID":   rm.serverID,
		}).Warn("CreateEscortMission: base reward must be positive")
		return nil, fmt.Errorf("base reward must be positive, got %.2f", baseReward)
	}

	route, exists := rm.routes[routeID]
	if !exists {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"serverID": rm.serverID,
		}).Warn("CreateEscortMission: route not found")
		return nil, fmt.Errorf("route not found: %s", routeID)
	}

	if route.Status != StatusActive {
		log.WithFields(log.Fields{
			"routeID":  routeID,
			"playerID": playerID,
			"status":   route.Status,
			"serverID": rm.serverID,
		}).Warn("CreateEscortMission: can only create missions for active routes")
		return nil, fmt.Errorf("can only create missions for active routes")
	}

	missionID := fmt.Sprintf("escort-%d", rm.nextMissionID)
	rm.nextMissionID++

	// Calculate rewards based on danger level
	reward := baseReward * (1.0 + route.DangerLevel)
	bonusReward := reward * 0.5 // 50% bonus for defeating attacks

	mission := &EscortMission{
		ID:          missionID,
		RouteID:     routeID,
		PlayerID:    playerID,
		Reward:      reward,
		BonusReward: bonusReward,
		Status:      MissionAvailable,
		AcceptedAt:  now(),
	}

	rm.missions[missionID] = mission
	return mission, nil
}

// Helper functions

func (rm *RouteManager) calculateServerHops(start, end string) int {
	// Simplified hop calculation (future: implement graph-based pathfinding)
	if start == end {
		return 0
	}
	// Default: 1-3 hops based on hash
	hash := 0
	for i := 0; i < len(start) && i < len(end); i++ {
		hash += int(start[i]) + int(end[i])
	}
	return 1 + (hash % 3)
}

func (rm *RouteManager) calculateDangerLevel(start, end string) float64 {
	// Calculate based on region properties (future: integrate with faction politics)
	baseHops := rm.calculateServerHops(start, end)
	dangerPerHop := 0.2
	return math.Min(1.0, float64(baseHops)*dangerPerHop)
}

func (rm *RouteManager) calculateProfitMargin(cargoValue, dangerLevel float64) float64 {
	// Higher danger = higher profit (10-50%)
	baseProfitMargin := 0.10          // 10% base
	dangerBonus := dangerLevel * 0.40 // up to 40% danger bonus
	return math.Min(0.50, baseProfitMargin+dangerBonus)
}

func (rm *RouteManager) generateRouteName(start, end string) string {
	// Generate procedural route names
	prefixes := []string{"Golden", "Silk", "Iron", "Spice", "Crystal"}
	suffixes := []string{"Way", "Road", "Path", "Route", "Trail"}

	prefixIdx := rm.rng.Intn(len(prefixes))
	suffixIdx := rm.rng.Intn(len(suffixes))

	return fmt.Sprintf("%s %s (%s → %s)", prefixes[prefixIdx], suffixes[suffixIdx], start, end)
}

func (rm *RouteManager) generateCargo(totalValue, profitMargin float64) []CargoItem {
	// Generate 3-8 cargo items
	itemCount := 3 + rm.rng.Intn(6)
	cargo := make([]CargoItem, itemCount)

	valuePerItem := totalValue / float64(itemCount)

	for i := 0; i < itemCount; i++ {
		purchasePrice := valuePerItem * (0.8 + rm.rng.Float64()*0.4) // ±20% variance
		targetPrice := purchasePrice * (1.0 + profitMargin)
		profit := targetPrice - purchasePrice

		cargo[i] = CargoItem{
			ItemID:        fmt.Sprintf("cargo-item-%d", i),
			ItemName:      rm.generateCargoName(),
			Quantity:      10 + rm.rng.Intn(90), // 10-100 units
			PurchasePrice: purchasePrice,
			TargetPrice:   targetPrice,
			Profit:        profit,
		}
	}

	return cargo
}

func (rm *RouteManager) generateCargoName() string {
	commodities := []string{
		"Timber", "Ore", "Grain", "Spices", "Textiles",
		"Pottery", "Tools", "Weapons", "Armor", "Medicine",
	}
	return commodities[rm.rng.Intn(len(commodities))]
}

func (rm *RouteManager) shouldSpawnBandit(route *TradeRoute, now time.Time) bool {
	// Spawn rate: 0.1-1.0 attacks per hour based on danger
	spawnRate := route.DangerLevel * 1.0    // attacks per hour
	chancePer10Seconds := spawnRate / 360.0 // 360 * 10sec = 1 hour

	return rm.rng.Float64() < chancePer10Seconds
}

func (rm *RouteManager) spawnBanditEncounter(route *TradeRoute, now time.Time) {
	encounterID := fmt.Sprintf("bandit-%d", rm.nextEncounterID)
	rm.nextEncounterID++

	// Calculate bandit strength based on cargo value
	cargoValue := 0.0
	for _, item := range route.Cargo {
		cargoValue += item.PurchasePrice * float64(item.Quantity)
	}
	banditStrength := 0.3 + (route.DangerLevel * 0.4) // 0.3-0.7

	// Calculate defense strength (caravan + escorts)
	defenseStrength := 0.4 + float64(len(route.EscortPlayers))*0.15

	encounter := &BanditEncounter{
		ID:              encounterID,
		RouteID:         route.ID,
		LocationX:       route.Progress * 100,
		LocationY:       0,                  // Simplified position
		BanditCount:     3 + rm.rng.Intn(7), // 3-10 bandits
		BanditStrength:  banditStrength,
		DefenseStrength: defenseStrength,
		StartTime:       now,
		Duration:        time.Duration(30+rm.rng.Intn(90)) * time.Second, // 30-120 seconds
		Outcome:         OutcomePending,
		LostCargo:       []CargoItem{},
		PlayerRewards:   make(map[uint64]float64),
	}

	rm.encounters[encounterID] = encounter
	route.BanditAttacks++
	route.Status = StatusUnderAttack
}

func (rm *RouteManager) updateEncounter(encounter *BanditEncounter, now time.Time) {
	elapsed := now.Sub(encounter.StartTime)
	if elapsed < encounter.Duration {
		return // Combat still ongoing
	}

	// Resolve encounter
	if encounter.DefenseStrength > encounter.BanditStrength {
		encounter.Outcome = OutcomeDefended
		rm.resolveDefendedEncounter(encounter)
	} else {
		strengthRatio := encounter.DefenseStrength / encounter.BanditStrength
		if strengthRatio > 0.7 {
			encounter.Outcome = OutcomeCompromised
			rm.resolveCompromisedEncounter(encounter)
		} else {
			encounter.Outcome = OutcomeDestroyed
			rm.resolveDestroyedEncounter(encounter)
		}
	}

	// Update route status
	if route, exists := rm.routes[encounter.RouteID]; exists {
		if encounter.Outcome == OutcomeDestroyed {
			route.Status = StatusFailed
		} else {
			route.Status = StatusActive // Resume travel
		}
	}
}

func (rm *RouteManager) resolveDefendedEncounter(encounter *BanditEncounter) {
	route, exists := rm.routes[encounter.RouteID]
	if !exists {
		return
	}

	// Calculate player rewards (escort bonus)
	baseReward := 100.0
	for _, playerID := range route.EscortPlayers {
		encounter.PlayerRewards[playerID] = baseReward
	}
}

func (rm *RouteManager) resolveCompromisedEncounter(encounter *BanditEncounter) {
	route, exists := rm.routes[encounter.RouteID]
	if !exists {
		return
	}

	// Lose 20-50% of cargo
	lossPercentage := 0.2 + rm.rng.Float64()*0.3
	for i := range route.Cargo {
		lostQuantity := int(float64(route.Cargo[i].Quantity) * lossPercentage)
		if lostQuantity > 0 {
			encounter.LostCargo = append(encounter.LostCargo, CargoItem{
				ItemID:        route.Cargo[i].ItemID,
				ItemName:      route.Cargo[i].ItemName,
				Quantity:      lostQuantity,
				PurchasePrice: route.Cargo[i].PurchasePrice,
			})
			route.Cargo[i].Quantity -= lostQuantity
		}
	}
}

func (rm *RouteManager) resolveDestroyedEncounter(encounter *BanditEncounter) {
	route, exists := rm.routes[encounter.RouteID]
	if !exists {
		return
	}

	// Lose all cargo
	encounter.LostCargo = route.Cargo
	route.Cargo = []CargoItem{}
	route.Status = StatusFailed
}

func (rm *RouteManager) completeRoute(route *TradeRoute) {
	// Calculate final profit and apply price impacts in a single pass.
	totalProfit := 0.0
	for _, item := range route.Cargo {
		totalProfit += item.Profit * float64(item.Quantity)
		if rm.priceHandler != nil && item.Quantity > 0 {
			// Successful delivery increases supply, reduces price.
			// Impact: 0.95-0.99 (1-5% price reduction) based on quantity.
			priceImpact := 1.0 - (float64(item.Quantity) / 1000.0 * 0.05)
			if priceImpact < 0.95 {
				priceImpact = 0.95 // Max 5% price reduction per route
			}
			rm.priceHandler.ApplyTradeImpact(item.ItemName, priceImpact, item.Quantity)
		}
	}

	// Update success rate
	route.SuccessRate = route.SuccessRate*0.9 + 0.1 // EWMA with 10% weight

	// Clean up caravan mapping
	delete(rm.activeCaravans, route.CaravanID)
}

func (rm *RouteManager) identifyDangerZones(route *TradeRoute) []DangerZone {
	// Simplified: create danger zones based on route danger level
	zoneCount := 1 + int(route.DangerLevel*3) // 1-4 zones
	zones := make([]DangerZone, zoneCount)

	for i := 0; i < zoneCount; i++ {
		zoneDanger := route.DangerLevel * (0.5 + rm.rng.Float64()*0.5)
		zones[i] = DangerZone{
			RegionID:           fmt.Sprintf("zone-%d", i),
			DangerLevel:        zoneDanger,
			BanditSpawnRate:    zoneDanger * 0.5,    // 0-0.5 attacks per hour
			RecommendedEscorts: int(zoneDanger * 3), // 0-3 escorts
			Description:        rm.generateDangerDescription(zoneDanger),
		}
	}

	return zones
}

func (rm *RouteManager) generateDangerDescription(dangerLevel float64) string {
	if dangerLevel < 0.3 {
		return "Relatively safe passage with occasional patrols"
	} else if dangerLevel < 0.6 {
		return "Moderate risk with known bandit activity"
	} else {
		return "High danger zone - expect frequent attacks"
	}
}

// CreateCaravan generates a new caravan vehicle for a trade route.
func (rm *RouteManager) CreateCaravan(seed int64, genreID string) (*vehicle.Vehicle, error) {
	// Generate a cart-type vehicle for cargo transport
	gen := vehicle.NewVehicleGenerator()
	params := procgen.GenerationParams{
		GenreID:    genreID,
		Depth:      1,
		Difficulty: 0.5,
		Custom: map[string]interface{}{
			"count":       1,
			"vehicleType": vehicle.TypeCart,
			"rarity":      vehicle.RarityCommon,
		},
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.WithFields(log.Fields{
			"seed":     seed,
			"genreID":  genreID,
			"serverID": rm.serverID,
			"error":    err.Error(),
		}).Error("CreateCaravan: failed to generate caravan")
		return nil, fmt.Errorf("failed to generate caravan: %w", err)
	}

	vehicles, ok := result.([]*vehicle.Vehicle)
	if !ok || len(vehicles) == 0 {
		log.WithFields(log.Fields{
			"seed":     seed,
			"genreID":  genreID,
			"serverID": rm.serverID,
		}).Error("CreateCaravan: generator returned invalid type or empty list")
		return nil, fmt.Errorf("generator returned invalid type or empty list")
	}

	return vehicles[0], nil
}
