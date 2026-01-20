package engine

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
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
	log.WithFields(log.Fields{
		"system_name":  "courier",
		"travel_speed": 2.0,
	}).Debug("Creating new courier system")

	system := &CourierSystem{
		world:       world,
		mailSystem:  mailSystem,
		serverGraph: make(map[string][]string),
		travelSpeed: 2.0,
	}

	log.WithFields(log.Fields{
		"system_name": "courier",
	}).Debug("Courier system initialized successfully")
	return system
}

// SetServerGraph sets the server connectivity graph for pathfinding
func (s *CourierSystem) SetServerGraph(graph map[string][]string) {
	log.WithFields(log.Fields{
		"system_name":  "courier",
		"server_count": len(graph),
	}).Debug("Setting server connectivity graph")

	s.serverGraph = graph

	log.WithFields(log.Fields{
		"system_name":  "courier",
		"server_count": len(graph),
	}).Info("Server connectivity graph updated")
}

// SetTravelSpeed sets the courier travel speed in tiles per second
func (s *CourierSystem) SetTravelSpeed(speed float64) {
	log.WithFields(log.Fields{
		"system_name":     "courier",
		"old_speed":       s.travelSpeed,
		"new_speed":       speed,
		"speed_tiles_sec": speed,
	}).Debug("Updating courier travel speed")

	s.travelSpeed = speed

	log.WithFields(log.Fields{
		"system_name": "courier",
		"speed":       speed,
	}).Info("Courier travel speed updated")
}

// Update processes all courier NPCs
func (s *CourierSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("courier")

	log.WithFields(log.Fields{
		"system_name":   "courier",
		"courier_count": len(entities),
		"delta_time_ms": deltaTime * 1000,
	}).Debug("Processing courier entities")

	for _, entity := range entities {
		s.updateCourier(entity, deltaTime)
	}
}

// updateCourier updates a single courier NPC
func (s *CourierSystem) updateCourier(entity *Entity, deltaTime float64) {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   entity.ID,
	}).Debug("Updating courier entity")

	comp, ok := entity.GetComponent("courier")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "courier",
			"entity_id":      entity.ID,
			"component_type": "courier",
		}).Warn("Entity missing courier component")
		return
	}
	courier := comp.(*CourierComponent)

	// Skip idle couriers
	if !courier.IsCarryingMail() {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   entity.ID,
		}).Debug("Courier is idle, skipping update")
		return
	}

	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   entity.ID,
		"message_id":  courier.CurrentMessageID,
	}).Debug("Processing active courier delivery")

	// Check if mail system has this courier tracked
	courierPos := s.mailSystem.GetCourierPosition(courier.CurrentMessageID)
	if courierPos == nil {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   entity.ID,
			"message_id":  courier.CurrentMessageID,
		}).Info("Mail delivery completed or failed, clearing courier assignment")

		// Mail was delivered or failed, clear courier assignment
		courier.CompleteDelivery()
		return
	}

	// Update progress based on travel time
	// Progress is time-based in MailSystem, we just sync the route progress
	// when the courier reaches a new server
	if courierPos.CurrentServer != courier.GetCurrentServer() {
		log.WithFields(log.Fields{
			"system_name":    "courier",
			"entity_id":      entity.ID,
			"message_id":     courier.CurrentMessageID,
			"current_server": courier.GetCurrentServer(),
			"new_server":     courierPos.CurrentServer,
		}).Info("Courier advancing to next server")

		// Courier should advance to next server
		if !courier.AdvanceRoute() {
			log.WithFields(log.Fields{
				"system_name": "courier",
				"entity_id":   entity.ID,
				"message_id":  courier.CurrentMessageID,
			}).Info("Courier reached destination, completing delivery")

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
	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   entity.ID,
		"x":           pos.X,
		"y":           pos.Y,
		"delta_time":  deltaTime,
	}).Debug("Updating courier visual movement")

	// This is a simplified visual update - actual pathfinding would be more complex
	// For now, couriers are invisible/teleport between servers as per roadmap
	// This allows for future expansion with visible courier movement

	// Move toward portal or destination point (placeholder logic)
	// Real implementation would query portal positions and move toward them
}

// AssignDeliveryToCourier assigns a mail delivery to an available courier
func (s *CourierSystem) AssignDeliveryToCourier(courierID uint64, messageID, fromServer, toServer string) error {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   courierID,
		"message_id":  messageID,
		"from_server": fromServer,
		"to_server":   toServer,
	}).Debug("Assigning delivery to courier")

	entity, exists := s.world.GetEntity(courierID)
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   courierID,
		}).Error("Courier entity not found")
		return fmt.Errorf("courier entity %d not found", courierID)
	}

	comp, ok := entity.GetComponent("courier")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "courier",
			"entity_id":      courierID,
			"component_type": "courier",
		}).Error("Entity missing courier component")
		return fmt.Errorf("entity %d has no courier component", courierID)
	}
	courier := comp.(*CourierComponent)

	if courier.IsCarryingMail() {
		log.WithFields(log.Fields{
			"system_name":        "courier",
			"entity_id":          courierID,
			"current_message_id": courier.CurrentMessageID,
		}).Warn("Courier already carrying mail, assignment failed")
		return fmt.Errorf("courier %d is already carrying mail", courierID)
	}

	// Calculate route using pathfinding
	log.WithFields(log.Fields{
		"system_name": "courier",
		"from_server": fromServer,
		"to_server":   toServer,
	}).Debug("Calculating delivery route")

	route := s.findRoute(fromServer, toServer)
	if len(route) == 0 {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"from_server": fromServer,
			"to_server":   toServer,
		}).Error("No route found between servers")
		return fmt.Errorf("no route found from %s to %s", fromServer, toServer)
	}

	log.WithFields(log.Fields{
		"system_name": "courier",
		"route_hops":  len(route),
		"route_path":  route,
	}).Debug("Route calculated successfully")

	// Assign delivery to courier
	courier.AssignDelivery(messageID, route)

	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   courierID,
		"message_id":  messageID,
		"route_hops":  len(route),
	}).Info("Delivery assigned to courier successfully")

	return nil
}

// findRoute finds the shortest path between two servers using BFS
func (s *CourierSystem) findRoute(fromServer, toServer string) []string {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"operation":   "pathfinding",
		"from_server": fromServer,
		"to_server":   toServer,
	}).Debug("Starting BFS pathfinding")

	if fromServer == toServer {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"server":      fromServer,
		}).Debug("Source and destination are same server")
		return []string{fromServer}
	}

	// BFS to find shortest path
	queue := []string{fromServer}
	visited := make(map[string]bool)
	parent := make(map[string]string)
	visited[fromServer] = true

	nodesExplored := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		nodesExplored++

		log.WithFields(log.Fields{
			"system_name":    "courier",
			"current_server": current,
			"queue_size":     len(queue),
			"nodes_explored": nodesExplored,
		}).Debug("Exploring server node")

		if current == toServer {
			// Reconstruct path
			path := []string{toServer}
			for parent[toServer] != "" {
				toServer = parent[toServer]
				path = append([]string{toServer}, path...)
			}

			log.WithFields(log.Fields{
				"system_name":    "courier",
				"route_hops":     len(path),
				"nodes_explored": nodesExplored,
				"route_path":     path,
			}).Info("Route found successfully")

			return path
		}

		// Explore neighbors
		neighbors := s.serverGraph[current]
		log.WithFields(log.Fields{
			"system_name":    "courier",
			"current_server": current,
			"neighbor_count": len(neighbors),
		}).Debug("Exploring neighbors")

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"from_server":    fromServer,
		"to_server":      toServer,
		"nodes_explored": nodesExplored,
	}).Warn("No route found, returning direct path")

	// No route found, return direct route (will fail if not connected)
	return []string{fromServer, toServer}
}

// FindAvailableCourier finds an idle courier NPC on the specified server
func (s *CourierSystem) FindAvailableCourier(serverID string) uint64 {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"server_id":   serverID,
	}).Debug("Searching for available courier")

	entities := s.world.GetEntitiesWith("courier")

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"total_couriers": len(entities),
	}).Debug("Processing courier entities")

	idleCount := 0
	for _, entity := range entities {
		comp, ok := entity.GetComponent("courier")
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "courier",
				"entity_id":      entity.ID,
				"component_type": "courier",
			}).Warn("Entity missing courier component")
			continue
		}
		courier := comp.(*CourierComponent)

		// Check if courier is idle and on the right server
		if !courier.IsCarryingMail() {
			idleCount++
			log.WithFields(log.Fields{
				"system_name": "courier",
				"entity_id":   entity.ID,
			}).Info("Found available idle courier")

			// Check server location (simplified - would need server component)
			// For now, assume any idle courier is available
			return entity.ID
		}
	}

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"server_id":      serverID,
		"idle_couriers":  idleCount,
		"total_couriers": len(entities),
	}).Warn("No available courier found")

	return 0 // No available courier
}

// GetCourierStatus returns information about a courier's current delivery
func (s *CourierSystem) GetCourierStatus(courierID uint64) (messageID, currentServer string, progress, totalHops int, err error) {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   courierID,
	}).Debug("Querying courier status")

	entity, exists := s.world.GetEntity(courierID)
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   courierID,
		}).Error("Courier entity not found")
		return "", "", 0, 0, fmt.Errorf("courier entity %d not found", courierID)
	}

	comp, ok := entity.GetComponent("courier")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "courier",
			"entity_id":      courierID,
			"component_type": "courier",
		}).Error("Entity missing courier component")
		return "", "", 0, 0, fmt.Errorf("entity %d has no courier component", courierID)
	}
	courier := comp.(*CourierComponent)

	if !courier.IsCarryingMail() {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   courierID,
		}).Debug("Courier is idle")
		return "", "", 0, 0, nil // Idle courier
	}

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"entity_id":      courierID,
		"message_id":     courier.CurrentMessageID,
		"current_server": courier.GetCurrentServer(),
		"progress":       courier.RouteProgress,
		"total_hops":     len(courier.CurrentRoute),
	}).Debug("Retrieved courier status")

	return courier.CurrentMessageID, courier.GetCurrentServer(), courier.RouteProgress, len(courier.CurrentRoute), nil
}

// EstimateDeliveryTime estimates the time remaining for a courier delivery in seconds
func (s *CourierSystem) EstimateDeliveryTime(courierID uint64) (float64, error) {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   courierID,
	}).Debug("Estimating delivery time")

	entity, exists := s.world.GetEntity(courierID)
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   courierID,
		}).Error("Courier entity not found")
		return 0, fmt.Errorf("courier entity %d not found", courierID)
	}

	comp, ok := entity.GetComponent("courier")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "courier",
			"entity_id":      courierID,
			"component_type": "courier",
		}).Error("Entity missing courier component")
		return 0, fmt.Errorf("entity %d has no courier component", courierID)
	}
	courier := comp.(*CourierComponent)

	if !courier.IsCarryingMail() {
		log.WithFields(log.Fields{
			"system_name": "courier",
			"entity_id":   courierID,
		}).Debug("Courier is idle, no delivery time estimate")
		return 0, nil
	}

	// Calculate remaining hops
	remainingHops := len(courier.CurrentRoute) - courier.RouteProgress - 1
	if remainingHops < 0 {
		remainingHops = 0
	}

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"entity_id":      courierID,
		"total_hops":     len(courier.CurrentRoute),
		"route_progress": courier.RouteProgress,
		"remaining_hops": remainingHops,
	}).Debug("Calculated remaining hops")

	// Use mail system's delivery time per hop
	deliveryTimePerHop := 300.0 // 5 minutes default
	if s.mailSystem != nil {
		// Would query actual delivery time from mail system
		// For now use default
	}

	estimatedTime := float64(remainingHops) * deliveryTimePerHop

	log.WithFields(log.Fields{
		"system_name":        "courier",
		"entity_id":          courierID,
		"remaining_hops":     remainingHops,
		"time_per_hop_sec":   deliveryTimePerHop,
		"estimated_time_sec": estimatedTime,
		"estimated_time_min": estimatedTime / 60.0,
	}).Info("Delivery time estimated")

	return estimatedTime, nil
}

// SpawnCourierNPC spawns a courier NPC at the specified location
func (s *CourierSystem) SpawnCourierNPC(x, y float64, name string) uint64 {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"operation":   "spawn",
		"x":           x,
		"y":           y,
		"name":        name,
	}).Debug("Spawning courier NPC")

	entity := s.world.CreateEntity()

	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   entity.ID,
	}).Debug("Entity created for courier")

	// Add position component
	pos := &PositionComponent{X: x, Y: y}
	entity.AddComponent(pos)

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"entity_id":      entity.ID,
		"component_type": "position",
	}).Debug("Position component added")

	// Add courier component
	courier := &CourierComponent{
		TravelSpeed: s.travelSpeed,
	}
	entity.AddComponent(courier)

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"entity_id":      entity.ID,
		"component_type": "courier",
		"travel_speed":   s.travelSpeed,
	}).Debug("Courier component added")

	// Add AI component for basic behavior
	ai := &AIComponent{
		State: AIStateIdle,
	}
	entity.AddComponent(ai)

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"entity_id":      entity.ID,
		"component_type": "ai",
		"ai_state":       "idle",
	}).Debug("AI component added")

	// Note: Sprite/visual components would be added here in production
	// For now, focusing on backend courier logic

	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   entity.ID,
		"name":        name,
		"x":           x,
		"y":           y,
	}).Info("Courier NPC spawned successfully")

	return entity.ID
}

// SpawnPostOffice spawns a post office building with a clerk NPC
func (s *CourierSystem) SpawnPostOffice(x, y float64, clerkName string) (buildingID, clerkID uint64) {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"operation":   "spawn",
		"x":           x,
		"y":           y,
		"clerk_name":  clerkName,
	}).Debug("Spawning post office")

	// Spawn post office building
	building := s.world.CreateEntity()

	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   building.ID,
		"entity_type": "building",
	}).Debug("Post office building entity created")

	building.AddComponent(&PositionComponent{X: x, Y: y})
	building.AddComponent(&PostOfficeComponent{
		ClerkName:   clerkName,
		ServiceFee:  10,
		MaxDistance: 100,
	})

	log.WithFields(log.Fields{
		"system_name":  "courier",
		"entity_id":    building.ID,
		"service_fee":  10,
		"max_distance": 100,
	}).Debug("Post office components added")

	// Note: Visual components would be added here in production

	// Spawn clerk NPC
	clerk := s.world.CreateEntity()

	log.WithFields(log.Fields{
		"system_name": "courier",
		"entity_id":   clerk.ID,
		"entity_type": "clerk",
	}).Debug("Post office clerk entity created")

	clerk.AddComponent(&PositionComponent{X: x + 2, Y: y + 2}) // Offset from building
	clerk.AddComponent(&PostOfficeClerkComponent{
		PostOfficeID:     building.ID,
		GreetingDialogue: fmt.Sprintf("Greetings! I am %s, how may I help you?", clerkName),
		ServiceFee:       5,
	})
	clerk.AddComponent(&AIComponent{
		State: AIStateIdle,
	})

	log.WithFields(log.Fields{
		"system_name":    "courier",
		"entity_id":      clerk.ID,
		"post_office_id": building.ID,
		"service_fee":    5,
	}).Debug("Clerk components added")

	// Note: Visual and interaction components would be added here in production

	log.WithFields(log.Fields{
		"system_name": "courier",
		"building_id": building.ID,
		"clerk_id":    clerk.ID,
		"clerk_name":  clerkName,
		"x":           x,
		"y":           y,
	}).Info("Post office spawned successfully")

	return building.ID, clerk.ID
}

// NotifyDeliveryComplete is called when the mail system completes a delivery
func (s *CourierSystem) NotifyDeliveryComplete(messageID string) {
	log.WithFields(log.Fields{
		"system_name": "courier",
		"message_id":  messageID,
	}).Debug("Processing delivery completion notification")

	// Find courier carrying this message and clear their assignment
	entities := s.world.GetEntitiesWith("courier")

	log.WithFields(log.Fields{
		"system_name":   "courier",
		"message_id":    messageID,
		"courier_count": len(entities),
	}).Debug("Searching for courier with message")

	for _, entity := range entities {
		comp, ok := entity.GetComponent("courier")
		if !ok {
			continue
		}
		courier := comp.(*CourierComponent)

		if courier.CurrentMessageID == messageID {
			log.WithFields(log.Fields{
				"system_name": "courier",
				"entity_id":   entity.ID,
				"message_id":  messageID,
			}).Info("Found courier, completing delivery")

			courier.CompleteDelivery()

			log.WithFields(log.Fields{
				"system_name": "courier",
				"entity_id":   entity.ID,
				"message_id":  messageID,
			}).Info("Courier assignment cleared")

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
