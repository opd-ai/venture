package engine

import "time"

// Helper methods for ServerFaction

// NewServerFaction creates a new server faction with the given parameters.
func NewServerFaction(serverID, factionName string, alignment Alignment) *ServerFaction {
	return &ServerFaction{
		ServerID:     serverID,
		FactionName:  factionName,
		Alignment:    alignment,
		AllyServers:  []string{},
		EnemyServers: []string{},
		Reputation:   make(map[string]float64),
	}
}

// IsAlly checks if the given server ID is an ally.
func (sf *ServerFaction) IsAlly(serverID string) bool {
	for _, ally := range sf.AllyServers {
		if ally == serverID {
			return true
		}
	}
	return false
}

// IsEnemy checks if the given server ID is an enemy.
func (sf *ServerFaction) IsEnemy(serverID string) bool {
	for _, enemy := range sf.EnemyServers {
		if enemy == serverID {
			return true
		}
	}
	return false
}

// AddAlly adds a server to the ally list if not already present.
func (sf *ServerFaction) AddAlly(serverID string) {
	if sf.IsAlly(serverID) || serverID == sf.ServerID {
		return
	}
	// Remove from enemies if present
	sf.RemoveEnemy(serverID)
	sf.AllyServers = append(sf.AllyServers, serverID)
}

// AddEnemy adds a server to the enemy list if not already present.
func (sf *ServerFaction) AddEnemy(serverID string) {
	if sf.IsEnemy(serverID) || serverID == sf.ServerID {
		return
	}
	// Remove from allies if present
	sf.RemoveAlly(serverID)
	sf.EnemyServers = append(sf.EnemyServers, serverID)
}

// RemoveAlly removes a server from the ally list.
func (sf *ServerFaction) RemoveAlly(serverID string) {
	for i, ally := range sf.AllyServers {
		if ally == serverID {
			sf.AllyServers = append(sf.AllyServers[:i], sf.AllyServers[i+1:]...)
			return
		}
	}
}

// RemoveEnemy removes a server from the enemy list.
func (sf *ServerFaction) RemoveEnemy(serverID string) {
	for i, enemy := range sf.EnemyServers {
		if enemy == serverID {
			sf.EnemyServers = append(sf.EnemyServers[:i], sf.EnemyServers[i+1:]...)
			return
		}
	}
}

// GetReputation returns the reputation score for a given player.
func (sf *ServerFaction) GetReputation(playerID string) float64 {
	if rep, exists := sf.Reputation[playerID]; exists {
		return rep
	}
	return 0.0 // Neutral reputation by default
}

// ModifyReputation adjusts a player's reputation, clamping to [-100, +100].
func (sf *ServerFaction) ModifyReputation(playerID string, delta float64) {
	current := sf.GetReputation(playerID)
	newRep := current + delta
	// Clamp to valid range
	if newRep < -100 {
		newRep = -100
	} else if newRep > 100 {
		newRep = 100
	}
	sf.Reputation[playerID] = newRep
}

// Helper methods for PoliticalEvent

// NewPoliticalEvent creates a new political event.
func NewPoliticalEvent(eventType string, partyServers []string, duration int64) *PoliticalEvent {
	return &PoliticalEvent{
		Type:         eventType,
		PartyServers: partyServers,
		StartTime:    time.Now().Unix(),
		Duration:     duration,
		Effects:      make(map[string]interface{}),
	}
}

// IsActive checks if the event is still active based on current time.
func (pe *PoliticalEvent) IsActive() bool {
	currentTime := time.Now().Unix()
	return currentTime < (pe.StartTime + pe.Duration)
}

// GetEffect retrieves an effect value by key.
func (pe *PoliticalEvent) GetEffect(key string) (interface{}, bool) {
	value, exists := pe.Effects[key]
	return value, exists
}

// SetEffect sets an effect value.
func (pe *PoliticalEvent) SetEffect(key string, value interface{}) {
	pe.Effects[key] = value
}

// Event type constants
const (
	EventTypeAlliance  = "alliance"
	EventTypeWar       = "war"
	EventTypeTreaty    = "treaty"
	EventTypeEmbargo   = "embargo"
	EventTypeTradePact = "trade_pact"
)
