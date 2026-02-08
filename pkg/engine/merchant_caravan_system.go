package engine

import (
	"time"
)

// MerchantCaravanSystem manages traveling merchant NPCs between servers
type MerchantCaravanSystem struct {
	world           *World
	hopDuration     float64 // Seconds per server hop (default: 300 = 5 minutes)
	routeCalculator func(origin, destination string) []string
	priceMarkupMin  float64 // Minimum markup (default: 1.1 = 10%)
	priceMarkupMax  float64 // Maximum markup (default: 1.5 = 50%)
	restDuration    float64 // Rest time at destination (default: 600 = 10 minutes)
}

// NewMerchantCaravanSystem creates a new merchant caravan system
func NewMerchantCaravanSystem(world *World) *MerchantCaravanSystem {
	return &MerchantCaravanSystem{
		world:           world,
		hopDuration:     300.0, // 5 minutes per hop
		priceMarkupMin:  1.1,   // 10% minimum markup
		priceMarkupMax:  1.5,   // 50% maximum markup
		restDuration:    600.0, // 10 minutes rest
		routeCalculator: defaultRouteCalculator,
	}
}

// SetRouteCalculator sets a custom route calculator function
func (s *MerchantCaravanSystem) SetRouteCalculator(calc func(origin, destination string) []string) {
	s.routeCalculator = calc
}

// SetHopDuration sets the travel time per server hop
func (s *MerchantCaravanSystem) SetHopDuration(seconds float64) {
	s.hopDuration = seconds
}

// Update processes merchant caravan movement
func (s *MerchantCaravanSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("merchantcaravan")

	for _, entity := range entities {
		comp := s.getCaravanComponent(entity)
		if comp == nil {
			continue
		}

		s.updateCaravan(entity, comp, deltaTime)
	}
}

// getCaravanComponent retrieves and validates the merchant caravan component.
func (s *MerchantCaravanSystem) getCaravanComponent(entity *Entity) *MerchantCaravanComponent {
	compInt, ok := entity.GetComponent("merchantcaravan")
	if !ok {
		return nil
	}
	return compInt.(*MerchantCaravanComponent)
}

// updateCaravan processes a single caravan's movement and rest state.
func (s *MerchantCaravanSystem) updateCaravan(entity *Entity, comp *MerchantCaravanComponent, deltaTime float64) {
	if s.isCaravanResting(comp) {
		return
	}

	s.processCaravanTravel(entity, comp, deltaTime)
}

// isCaravanResting checks if the caravan is still resting at destination.
func (s *MerchantCaravanSystem) isCaravanResting(comp *MerchantCaravanComponent) bool {
	if comp.NextDepartureTime <= 0 {
		return false
	}

	now := time.Now().Unix()
	if now < comp.NextDepartureTime {
		return true
	}

	// Time to depart - reset and continue journey
	comp.NextDepartureTime = 0
	return false
}

// processCaravanTravel advances the caravan along its route.
func (s *MerchantCaravanSystem) processCaravanTravel(entity *Entity, comp *MerchantCaravanComponent, deltaTime float64) {
	if !s.hasValidRoute(comp) {
		return
	}

	s.advanceTravelProgress(comp, deltaTime)
	s.checkServerHopCompletion(entity, comp)
}

// hasValidRoute verifies the caravan has a valid route to follow.
func (s *MerchantCaravanSystem) hasValidRoute(comp *MerchantCaravanComponent) bool {
	return len(comp.RouteServers) > 0 && comp.CurrentRouteIndex < len(comp.RouteServers)
}

// advanceTravelProgress increments the caravan's travel progress.
func (s *MerchantCaravanSystem) advanceTravelProgress(comp *MerchantCaravanComponent, deltaTime float64) {
	progressDelta := deltaTime / s.hopDuration
	comp.TravelProgress += progressDelta
}

// checkServerHopCompletion handles completion of a server hop.
func (s *MerchantCaravanSystem) checkServerHopCompletion(entity *Entity, comp *MerchantCaravanComponent) {
	if comp.TravelProgress < 1.0 {
		return
	}

	comp.TravelProgress = 0.0
	comp.CurrentRouteIndex++

	s.updateCurrentServer(comp)
	s.checkFinalDestination(entity, comp)
}

// updateCurrentServer sets the caravan's current server location.
func (s *MerchantCaravanSystem) updateCurrentServer(comp *MerchantCaravanComponent) {
	if comp.CurrentRouteIndex < len(comp.RouteServers) {
		comp.CurrentServer = comp.RouteServers[comp.CurrentRouteIndex]
	}
}

// checkFinalDestination handles arrival at the final destination.
func (s *MerchantCaravanSystem) checkFinalDestination(entity *Entity, comp *MerchantCaravanComponent) {
	if comp.CurrentRouteIndex >= len(comp.RouteServers)-1 {
		s.handleArrival(entity, comp)
	}
}

// handleArrival processes merchant arrival at destination
func (s *MerchantCaravanSystem) handleArrival(entity *Entity, comp *MerchantCaravanComponent) {
	// Set rest timer
	comp.NextDepartureTime = time.Now().Unix() + int64(s.restDuration)

	// Update marketplace with caravan items if merchant component exists
	// Note: Actual item instance conversion would require item lookup by ItemID
	// For now, merchant inventory is managed separately through commerce system
	// Caravan inventory tracks trade goods metadata (ItemID, quantities, pricing)
	// UI layer can query caravan inventory directly for display
}

// CreateCaravan spawns a new merchant caravan entity
func (s *MerchantCaravanSystem) CreateCaravan(origin, destination string, inventory []CaravanItem) *Entity {
	entity := s.world.CreateEntity()

	route := s.routeCalculator(origin, destination)

	caravan := &MerchantCaravanComponent{
		OriginServer:      origin,
		DestinationServer: destination,
		CurrentServer:     origin,
		RouteServers:      route,
		CurrentRouteIndex: 0,
		TravelProgress:    0.0,
		Inventory:         inventory,
		TravelSpeed:       1.0,
		NextDepartureTime: 0,
	}

	entity.AddComponent(caravan)

	// Add position component (merchants spawn at server entry point)
	position := &PositionComponent{X: 0, Y: 0}
	entity.AddComponent(position)

	return entity
}

// GetCaravansAtServer returns all caravans currently at a specific server
func (s *MerchantCaravanSystem) GetCaravansAtServer(serverID string) []*Entity {
	entities := s.world.GetEntitiesWith("merchantcaravan")
	result := make([]*Entity, 0)

	for _, entity := range entities {
		compInt, ok := entity.GetComponent("merchantcaravan")
		if !ok {
			continue
		}
		comp := compInt.(*MerchantCaravanComponent)
		if comp.CurrentServer == serverID && comp.TravelProgress < 0.1 {
			result = append(result, entity)
		}
	}

	return result
}

// EstimateArrivalTime calculates when a caravan will arrive at destination
func (s *MerchantCaravanSystem) EstimateArrivalTime(entity *Entity) int64 {
	compInt, ok := entity.GetComponent("merchantcaravan")
	if !ok {
		return 0
	}
	comp := compInt.(*MerchantCaravanComponent)

	// If already at destination
	if comp.CurrentRouteIndex >= len(comp.RouteServers)-1 {
		return time.Now().Unix()
	}

	// Current hop progress (remaining time on current hop)
	currentHopRemaining := (1.0 - comp.TravelProgress) * s.hopDuration

	// Calculate remaining hops AFTER current one
	remainingHops := len(comp.RouteServers) - comp.CurrentRouteIndex - 2
	if remainingHops < 0 {
		remainingHops = 0
	}

	// Total time
	totalSeconds := currentHopRemaining + float64(remainingHops)*s.hopDuration

	return time.Now().Unix() + int64(totalSeconds)
}

// CalculateSalePrice determines the price a merchant will charge for an item
func (s *MerchantCaravanSystem) CalculateSalePrice(purchasePrice float64, serverHops int) float64 {
	// Base markup from travel distance
	markup := s.priceMarkupMin + (s.priceMarkupMax-s.priceMarkupMin)*float64(serverHops)/10.0
	if markup > s.priceMarkupMax {
		markup = s.priceMarkupMax
	}

	return purchasePrice * markup
}

// defaultRouteCalculator provides a simple direct route
func defaultRouteCalculator(origin, destination string) []string {
	if origin == destination {
		return []string{origin}
	}
	return []string{origin, destination}
}
