package engine

// CourierComponent represents an NPC courier that delivers mail
type CourierComponent struct {
	CurrentMessageID string   // Message being carried (empty if no delivery active)
	CurrentRoute     []string // Server path from origin to destination
	RouteProgress    int      // Current position in route (index into CurrentRoute)
	TravelSpeed      float64  // Tiles per second when visible (default: 2.0)
}

// Type returns the component type
func (c CourierComponent) Type() string {
	return "courier"
}

// IsCarryingMail returns true if the courier is currently delivering a message
func (c *CourierComponent) IsCarryingMail() bool {
	return c.CurrentMessageID != ""
}

// GetCurrentServer returns the server the courier is currently on
func (c *CourierComponent) GetCurrentServer() string {
	if c.RouteProgress < 0 || c.RouteProgress >= len(c.CurrentRoute) {
		return ""
	}
	return c.CurrentRoute[c.RouteProgress]
}

// GetNextServer returns the next server in the route, or empty if at destination
func (c *CourierComponent) GetNextServer() string {
	nextIdx := c.RouteProgress + 1
	if nextIdx >= len(c.CurrentRoute) {
		return ""
	}
	return c.CurrentRoute[nextIdx]
}

// HasReachedDestination returns true if the courier has completed its route
func (c *CourierComponent) HasReachedDestination() bool {
	return c.RouteProgress >= len(c.CurrentRoute)-1
}

// AssignDelivery assigns a mail delivery to the courier
func (c *CourierComponent) AssignDelivery(messageID string, route []string) {
	c.CurrentMessageID = messageID
	c.CurrentRoute = make([]string, len(route))
	copy(c.CurrentRoute, route)
	c.RouteProgress = 0
}

// AdvanceRoute moves the courier to the next server in the route
func (c *CourierComponent) AdvanceRoute() bool {
	if c.RouteProgress >= len(c.CurrentRoute)-1 {
		return false // Already at destination
	}
	c.RouteProgress++
	return true
}

// CompleteDelivery clears the current delivery assignment
func (c *CourierComponent) CompleteDelivery() {
	c.CurrentMessageID = ""
	c.CurrentRoute = nil
	c.RouteProgress = 0
}

// PostOfficeClerkComponent marks an NPC as a post office clerk
type PostOfficeClerkComponent struct {
	PostOfficeID     uint64 // Entity ID of the post office building this clerk works at
	GreetingDialogue string // Procedurally generated greeting
	ServiceFee       int    // Additional fee charged by this clerk
}

// Type returns the component type
func (p PostOfficeClerkComponent) Type() string {
	return "postoffice_clerk"
}
