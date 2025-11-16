package engine

import (
	"fmt"
	"time"
)

// CourierSystem manages courier NPCs that deliver mail between servers
type CourierSystem struct {
	world       *World
	mailSystem  *MailSystem
	serverGraph map[string][]string // Server connectivity graph for pathfinding
	travelSpeed float64             // Base travel speed in tiles/second (default: 2.0)
}

// NewCourierSystem creates a new courier system
func NewCourierSystem(world *World, mailSystem *MailSystem) *CourierSystem {
	return &CourierSystem{
		world:       world,
		mailSystem:  mailSystem,
		serverGraph: make(map[string][]string),
		travelSpeed: 2.0,
	}
}

// SetServerGraph sets the server connectivity graph for pathfinding
func (s *CourierSystem) SetServerGraph(graph map[string][]string) {
	s.serverGraph = graph
}

// SetTravelSpeed sets the courier travel speed in tiles per second
func (s *CourierSystem) SetTravelSpeed(speed float64) {
	s.travelSpeed = speed
}

// Update processes all courier NPCs
func (s *CourierSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("courier")
	for _, entity := range entities {
		s.updateCourier(entity, deltaTime)
	}
}

// updateCourier updates a single courier NPC
func (s *CourierSystem) updateCourier(entity *Entity, deltaTime float64) {
	comp, ok := entity.GetComponent("courier")
	if !ok {
		return
	}
	courier := comp.(*CourierComponent)

	// Skip idle couriers
	if !courier.IsCarryingMail() {
		return
	}

	// Check if mail system has this courier tracked
	courierPos := s.mailSystem.GetCourierPosition(courier.CurrentMessageID)
	if courierPos == nil {
		// Mail was delivered or failed, clear courier assignment
		courier.CompleteDelivery()
		return
	}

	// Update progress based on travel time
	// Progress is time-based in MailSystem, we just sync the route progress
	// when the courier reaches a new server
	if courierPos.CurrentServer != courier.GetCurrentServer() {
		// Courier should advance to next server
		if !courier.AdvanceRoute() {
			// Reached destination, complete delivery
			courier.CompleteDelivery()
		}
	}

	// Visual movement if courier has position component
	if posComp, ok := entity.GetComponent("position"); ok {
		s.updateCourierMovement(entity, posComp.(*PositionComponent), courier, deltaTime)
	}
}

// updateCourierMovement updates courier position for visual representation
func (s *CourierSystem) updateCourierMovement(entity *Entity, pos *PositionComponent, courier *CourierComponent, deltaTime float64) {
	// This is a simplified visual update - actual pathfinding would be more complex
	// For now, couriers are invisible/teleport between servers as per roadmap
	// This allows for future expansion with visible courier movement

	// Move toward portal or destination point (placeholder logic)
	// Real implementation would query portal positions and move toward them
}

// AssignDeliveryToCourier assigns a mail delivery to an available courier
func (s *CourierSystem) AssignDeliveryToCourier(courierID uint64, messageID, fromServer, toServer string) error {
	entity, exists := s.world.GetEntity(courierID)
	if !exists {
		return fmt.Errorf("courier entity %d not found", courierID)
	}

	comp, ok := entity.GetComponent("courier")
	if !ok {
		return fmt.Errorf("entity %d has no courier component", courierID)
	}
	courier := comp.(*CourierComponent)

	if courier.IsCarryingMail() {
		return fmt.Errorf("courier %d is already carrying mail", courierID)
	}

	// Calculate route using pathfinding
	route := s.findRoute(fromServer, toServer)
	if len(route) == 0 {
		return fmt.Errorf("no route found from %s to %s", fromServer, toServer)
	}

	// Assign delivery to courier
	courier.AssignDelivery(messageID, route)

	return nil
}

// findRoute finds the shortest path between two servers using BFS
func (s *CourierSystem) findRoute(fromServer, toServer string) []string {
	if fromServer == toServer {
		return []string{fromServer}
	}

	// BFS to find shortest path
	queue := []string{fromServer}
	visited := make(map[string]bool)
	parent := make(map[string]string)
	visited[fromServer] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == toServer {
			// Reconstruct path
			path := []string{toServer}
			for parent[toServer] != "" {
				toServer = parent[toServer]
				path = append([]string{toServer}, path...)
			}
			return path
		}

		// Explore neighbors
		for _, neighbor := range s.serverGraph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	// No route found, return direct route (will fail if not connected)
	return []string{fromServer, toServer}
}

// FindAvailableCourier finds an idle courier NPC on the specified server
func (s *CourierSystem) FindAvailableCourier(serverID string) uint64 {
	entities := s.world.GetEntitiesWith("courier")
	for _, entity := range entities {
		comp, ok := entity.GetComponent("courier")
		if !ok {
			continue
		}
		courier := comp.(*CourierComponent)

		// Check if courier is idle and on the right server
		if !courier.IsCarryingMail() {
			// Check server location (simplified - would need server component)
			// For now, assume any idle courier is available
			return entity.ID
		}
	}
	return 0 // No available courier
}

// GetCourierStatus returns information about a courier's current delivery
func (s *CourierSystem) GetCourierStatus(courierID uint64) (messageID, currentServer string, progress, totalHops int, err error) {
	entity, exists := s.world.GetEntity(courierID)
	if !exists {
		return "", "", 0, 0, fmt.Errorf("courier entity %d not found", courierID)
	}

	comp, ok := entity.GetComponent("courier")
	if !ok {
		return "", "", 0, 0, fmt.Errorf("entity %d has no courier component", courierID)
	}
	courier := comp.(*CourierComponent)

	if !courier.IsCarryingMail() {
		return "", "", 0, 0, nil // Idle courier
	}

	return courier.CurrentMessageID, courier.GetCurrentServer(), courier.RouteProgress, len(courier.CurrentRoute), nil
}

// EstimateDeliveryTime estimates the time remaining for a courier delivery in seconds
func (s *CourierSystem) EstimateDeliveryTime(courierID uint64) (float64, error) {
	entity, exists := s.world.GetEntity(courierID)
	if !exists {
		return 0, fmt.Errorf("courier entity %d not found", courierID)
	}

	comp, ok := entity.GetComponent("courier")
	if !ok {
		return 0, fmt.Errorf("entity %d has no courier component", courierID)
	}
	courier := comp.(*CourierComponent)

	if !courier.IsCarryingMail() {
		return 0, nil
	}

	// Calculate remaining hops
	remainingHops := len(courier.CurrentRoute) - courier.RouteProgress - 1
	if remainingHops < 0 {
		remainingHops = 0
	}

	// Use mail system's delivery time per hop
	deliveryTimePerHop := 300.0 // 5 minutes default
	if s.mailSystem != nil {
		// Would query actual delivery time from mail system
		// For now use default
	}

	return float64(remainingHops) * deliveryTimePerHop, nil
}

// SpawnCourierNPC spawns a courier NPC at the specified location
func (s *CourierSystem) SpawnCourierNPC(x, y float64, name string) uint64 {
	entity := s.world.CreateEntity()

	// Add position component
	pos := &PositionComponent{X: x, Y: y}
	entity.AddComponent(pos)

	// Add courier component
	courier := &CourierComponent{
		TravelSpeed: s.travelSpeed,
	}
	entity.AddComponent(courier)

	// Add AI component for basic behavior
	ai := &AIComponent{
		State: AIStateIdle,
	}
	entity.AddComponent(ai)

	// Note: Sprite/visual components would be added here in production
	// For now, focusing on backend courier logic

	return entity.ID
}

// SpawnPostOffice spawns a post office building with a clerk NPC
func (s *CourierSystem) SpawnPostOffice(x, y float64, clerkName string) (buildingID, clerkID uint64) {
	// Spawn post office building
	building := s.world.CreateEntity()
	building.AddComponent(&PositionComponent{X: x, Y: y})
	building.AddComponent(&PostOfficeComponent{
		ClerkName:   clerkName,
		ServiceFee:  10,
		MaxDistance: 100,
	})
	// Note: Visual components would be added here in production

	// Spawn clerk NPC
	clerk := s.world.CreateEntity()
	clerk.AddComponent(&PositionComponent{X: x + 2, Y: y + 2}) // Offset from building
	clerk.AddComponent(&PostOfficeClerkComponent{
		PostOfficeID:     building.ID,
		GreetingDialogue: fmt.Sprintf("Greetings! I am %s, how may I help you?", clerkName),
		ServiceFee:       5,
	})
	clerk.AddComponent(&AIComponent{
		State: AIStateIdle,
	})
	// Note: Visual and interaction components would be added here in production

	return building.ID, clerk.ID
}

// NotifyDeliveryComplete is called when the mail system completes a delivery
func (s *CourierSystem) NotifyDeliveryComplete(messageID string) {
	// Find courier carrying this message and clear their assignment
	entities := s.world.GetEntitiesWith([]string{"courier"})
	for _, entity := range entities {
		comp, ok := entity.GetComponent("courier")
		if !ok {
			continue
		}
		courier := comp.(*CourierComponent)

		if courier.CurrentMessageID == messageID {
			courier.CompleteDelivery()
			break
		}
	}
}

// Helper to format courier arrival time
func formatArrivalTime(seconds float64) string {
	if seconds <= 0 {
		return "Arrived"
	}

	duration := time.Duration(seconds) * time.Second
	if duration < time.Minute {
		return fmt.Sprintf("%.0fs", seconds)
	}
	if duration < time.Hour {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	return fmt.Sprintf("%.1fh", duration.Hours())
}
