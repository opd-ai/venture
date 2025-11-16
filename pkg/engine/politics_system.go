package engine

import (
	"fmt"
	"sync"
)

// PoliticsSystem manages server faction relationships and political events.
type PoliticsSystem struct {
	world         *World
	serverFaction *ServerFaction   // This server's faction
	activeEvents  []PoliticalEvent // Currently active political events
	eventHistory  []PoliticalEvent // Historical events
	mu            sync.RWMutex     // Protects concurrent access
}

// NewPoliticsSystem creates a new politics system.
func NewPoliticsSystem(world *World) *PoliticsSystem {
	return &PoliticsSystem{
		world:        world,
		activeEvents: []PoliticalEvent{},
		eventHistory: []PoliticalEvent{},
	}
}

// SetServerFaction sets this server's faction identity.
func (ps *PoliticsSystem) SetServerFaction(faction *ServerFaction) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.serverFaction = faction
}

// GetServerFaction returns this server's faction.
func (ps *PoliticsSystem) GetServerFaction() *ServerFaction {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.serverFaction
}

// CreateAlliance creates an alliance between servers.
func (ps *PoliticsSystem) CreateAlliance(serverID string, duration int64) (PoliticalEvent, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.serverFaction == nil {
		return PoliticalEvent{}, fmt.Errorf("server has no faction identity")
	}

	// Create alliance event
	event := NewPoliticalEvent(EventTypeAlliance, []string{ps.serverFaction.ServerID, serverID}, duration)
	event.SetEffect("trade_price_multiplier", 0.8) // 20% discount
	event.SetEffect("travel_cost_multiplier", 0.0) // Free travel

	// Update faction relationships
	ps.serverFaction.AddAlly(serverID)

	ps.activeEvents = append(ps.activeEvents, *event)
	return *event, nil
}

// DeclareWar declares war on another server.
func (ps *PoliticsSystem) DeclareWar(serverID string, duration int64) (PoliticalEvent, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.serverFaction == nil {
		return PoliticalEvent{}, fmt.Errorf("server has no faction identity")
	}

	// Create war event
	event := NewPoliticalEvent(EventTypeWar, []string{ps.serverFaction.ServerID, serverID}, duration)
	event.SetEffect("trade_price_multiplier", 1.5) // 50% markup
	event.SetEffect("contested_borders", true)     // Enable border PvP zones

	// Update faction relationships
	ps.serverFaction.AddEnemy(serverID)

	ps.activeEvents = append(ps.activeEvents, *event)
	return *event, nil
}

// SignTreaty signs a peace treaty with another server.
func (ps *PoliticsSystem) SignTreaty(serverID string, duration int64) (PoliticalEvent, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.serverFaction == nil {
		return PoliticalEvent{}, fmt.Errorf("server has no faction identity")
	}

	// Create treaty event
	event := NewPoliticalEvent(EventTypeTreaty, []string{ps.serverFaction.ServerID, serverID}, duration)
	event.SetEffect("trade_price_multiplier", 1.0) // Normal pricing
	event.SetEffect("contested_borders", false)    // Disable border PvP

	// Update faction relationships (remove from enemy list)
	ps.serverFaction.RemoveEnemy(serverID)

	ps.activeEvents = append(ps.activeEvents, *event)
	return *event, nil
}

// ImposeEmbargo imposes a trade embargo on another server.
func (ps *PoliticsSystem) ImposeEmbargo(serverID string, duration int64) (PoliticalEvent, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.serverFaction == nil {
		return PoliticalEvent{}, fmt.Errorf("server has no faction identity")
	}

	// Create embargo event
	event := NewPoliticalEvent(EventTypeEmbargo, []string{ps.serverFaction.ServerID, serverID}, duration)
	event.SetEffect("trade_blocked", true)     // Disable direct trade
	event.SetEffect("shipping_allowed", false) // No courier service

	ps.activeEvents = append(ps.activeEvents, *event)
	return *event, nil
}

// EstablishTradePact establishes a trade agreement with another server.
func (ps *PoliticsSystem) EstablishTradePact(serverID string, duration int64) (PoliticalEvent, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.serverFaction == nil {
		return PoliticalEvent{}, fmt.Errorf("server has no faction identity")
	}

	// Create trade pact event
	event := NewPoliticalEvent(EventTypeTradePact, []string{ps.serverFaction.ServerID, serverID}, duration)
	event.SetEffect("trade_price_multiplier", 0.9)   // 10% discount
	event.SetEffect("shipping_cost_multiplier", 0.8) // 20% shipping discount

	ps.activeEvents = append(ps.activeEvents, *event)
	return *event, nil
}

// Update processes political events and updates entity politics components.
func (ps *PoliticsSystem) Update(deltaTime float64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check for expired events
	remainingEvents := []PoliticalEvent{}
	for _, event := range ps.activeEvents {
		if event.IsActive() {
			remainingEvents = append(remainingEvents, event)
		} else {
			// Move to history
			ps.eventHistory = append(ps.eventHistory, event)
		}
	}
	ps.activeEvents = remainingEvents

	// Update PoliticsComponents on entities
	entities := ps.world.GetEntitiesWith("politics")
	for _, entity := range entities {
		comp, ok := entity.GetComponent("politics")
		if !ok {
			continue
		}
		politicsComp, ok := comp.(PoliticsComponent)
		if !ok {
			continue
		}

		// Update component with server faction and active events
		if ps.serverFaction != nil {
			politicsComp.Faction = ps.serverFaction
		}
		politicsComp.Events = ps.activeEvents

		// Store back
		entity.AddComponent(politicsComp)
	}
}

// GetActiveEvents returns all currently active political events.
func (ps *PoliticsSystem) GetActiveEvents() []PoliticalEvent {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Return copy to prevent external modification
	events := make([]PoliticalEvent, len(ps.activeEvents))
	copy(events, ps.activeEvents)
	return events
}

// GetEventHistory returns historical political events.
func (ps *PoliticsSystem) GetEventHistory() []PoliticalEvent {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Return copy to prevent external modification
	history := make([]PoliticalEvent, len(ps.eventHistory))
	copy(history, ps.eventHistory)
	return history
}

// GetTradeMultiplier returns the trade price multiplier for transactions with a given server.
func (ps *PoliticsSystem) GetTradeMultiplier(serverID string) float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	multiplier := 1.0 // Default: no modifier

	// Check active events for trade effects
	for _, event := range ps.activeEvents {
		if !event.IsActive() {
			continue
		}

		// Check if this event involves the target server
		involves := false
		for _, party := range event.PartyServers {
			if party == serverID {
				involves = true
				break
			}
		}

		if !involves {
			continue
		}

		// Apply trade multiplier from event effects
		if val, exists := event.GetEffect("trade_price_multiplier"); exists {
			if mult, ok := val.(float64); ok {
				multiplier *= mult
			}
		}
	}

	return multiplier
}

// IsTravelAllowed checks if travel to a given server is permitted.
func (ps *PoliticsSystem) IsTravelAllowed(serverID string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Travel is generally allowed in all cases
	// Wars and embargos don't block travel, just affect cost/safety
	return true
}

// IsTradeBlocked checks if trade with a given server is blocked.
func (ps *PoliticsSystem) IsTradeBlocked(serverID string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Check for embargo events
	for _, event := range ps.activeEvents {
		if !event.IsActive() {
			continue
		}

		// Check if this event involves the target server
		involves := false
		for _, party := range event.PartyServers {
			if party == serverID {
				involves = true
				break
			}
		}

		if !involves {
			continue
		}

		// Check if trade is blocked
		if event.Type == EventTypeEmbargo {
			if val, exists := event.GetEffect("trade_blocked"); exists {
				if blocked, ok := val.(bool); ok && blocked {
					return true
				}
			}
		}
	}

	return false // Trade allowed by default
}
