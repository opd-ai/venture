// Package world provides meta-game event systems for cross-server gameplay.
package world

import (
	"fmt"
	"math/rand"
	"time"
)

// EventType defines categories of meta-game events.
type EventType int

const (
	EventTournament EventType = iota
	EventServerVsServer
	EventWorldThreat
	EventSeasonalChallenge
	EventEconomicCrisis
)

// String returns human-readable event type.
func (e EventType) String() string {
	switch e {
	case EventTournament:
		return "Tournament"
	case EventServerVsServer:
		return "Server vs Server"
	case EventWorldThreat:
		return "World Threat"
	case EventSeasonalChallenge:
		return "Seasonal Challenge"
	case EventEconomicCrisis:
		return "Economic Crisis"
	default:
		return "Unknown"
	}
}

// MetaGameEvent represents a cross-server event.
type MetaGameEvent struct {
	ID              string
	Type            EventType
	Name            string
	Description     string
	StartTime       int64
	EndTime         int64
	Participants    []string
	Goals           map[string]int
	Progress        map[string]int
	Rewards         map[string]int
	Active          bool
	RequiredServers int
}

// EventManager manages meta-game events across servers.
type EventManager struct {
	events     map[string]*MetaGameEvent
	activeOnly bool
	seed       int64
	rng        *rand.Rand
	nextID     int
}

// NewEventManager creates a new event manager.
func NewEventManager(seed int64) *EventManager {
	return &EventManager{
		events:     make(map[string]*MetaGameEvent),
		activeOnly: true,
		seed:       seed,
		rng:        rand.New(rand.NewSource(seed)),
		nextID:     1,
	}
}

// CreateTournament creates a cross-server PvP tournament event.
func (em *EventManager) CreateTournament(name string, duration int64, requiredServers int) *MetaGameEvent {
	id := fmt.Sprintf("tournament_%d", em.nextID)
	em.nextID++

	event := &MetaGameEvent{
		ID:              id,
		Type:            EventTournament,
		Name:            name,
		Description:     "Compete in cross-server PvP battles for glory and rewards",
		StartTime:       time.Now().Unix(),
		EndTime:         time.Now().Unix() + duration,
		Participants:    make([]string, 0),
		Goals:           map[string]int{"wins": 10},
		Progress:        make(map[string]int),
		Rewards:         map[string]int{"gold": 1000, "renown": 100},
		Active:          true,
		RequiredServers: requiredServers,
	}

	em.events[id] = event
	return event
}

// CreateServerVsServer creates a server vs server competition event.
func (em *EventManager) CreateServerVsServer(name, serverA, serverB string, duration int64) *MetaGameEvent {
	id := fmt.Sprintf("svs_%d", em.nextID)
	em.nextID++

	event := &MetaGameEvent{
		ID:              id,
		Type:            EventServerVsServer,
		Name:            name,
		Description:     fmt.Sprintf("%s vs %s - collective resource gathering competition", serverA, serverB),
		StartTime:       time.Now().Unix(),
		EndTime:         time.Now().Unix() + duration,
		Participants:    []string{serverA, serverB},
		Goals:           map[string]int{"resources": 10000},
		Progress:        map[string]int{serverA: 0, serverB: 0},
		Rewards:         map[string]int{"bonus_multiplier": 150},
		Active:          true,
		RequiredServers: 2,
	}

	em.events[id] = event
	return event
}

// CreateWorldThreat creates a cooperative threat event requiring multiple servers.
func (em *EventManager) CreateWorldThreat(name string, duration int64, requiredServers int) *MetaGameEvent {
	id := fmt.Sprintf("threat_%d", em.nextID)
	em.nextID++

	threatLevel := em.rng.Intn(5) + 3
	bossHealth := threatLevel * 50000

	event := &MetaGameEvent{
		ID:              id,
		Type:            EventWorldThreat,
		Name:            name,
		Description:     "Ancient evil awakens - all servers must cooperate to defeat it",
		StartTime:       time.Now().Unix(),
		EndTime:         time.Now().Unix() + duration,
		Participants:    make([]string, 0),
		Goals:           map[string]int{"boss_damage": bossHealth},
		Progress:        map[string]int{"damage_dealt": 0},
		Rewards:         map[string]int{"legendary_items": threatLevel, "experience": bossHealth / 100},
		Active:          true,
		RequiredServers: requiredServers,
	}

	em.events[id] = event
	return event
}

// CreateSeasonalChallenge creates a seasonal event with time-limited objectives.
func (em *EventManager) CreateSeasonalChallenge(season string, duration int64) *MetaGameEvent {
	id := fmt.Sprintf("seasonal_%d", em.nextID)
	em.nextID++

	event := &MetaGameEvent{
		ID:              id,
		Type:            EventSeasonalChallenge,
		Name:            fmt.Sprintf("%s Challenge", season),
		Description:     fmt.Sprintf("Complete seasonal objectives during the %s season", season),
		StartTime:       time.Now().Unix(),
		EndTime:         time.Now().Unix() + duration,
		Participants:    make([]string, 0),
		Goals:           map[string]int{"quests_completed": 50, "bosses_defeated": 5},
		Progress:        make(map[string]int),
		Rewards:         map[string]int{"seasonal_currency": 500, "cosmetic_unlock": 1},
		Active:          true,
		RequiredServers: 1,
	}

	em.events[id] = event
	return event
}

// CreateEconomicCrisis creates an economic challenge event.
func (em *EventManager) CreateEconomicCrisis(name string, duration int64) *MetaGameEvent {
	id := fmt.Sprintf("crisis_%d", em.nextID)
	em.nextID++

	event := &MetaGameEvent{
		ID:              id,
		Type:            EventEconomicCrisis,
		Name:            name,
		Description:     "Economic instability affects all servers - stabilize markets",
		StartTime:       time.Now().Unix(),
		EndTime:         time.Now().Unix() + duration,
		Participants:    make([]string, 0),
		Goals:           map[string]int{"trade_volume": 100000},
		Progress:        map[string]int{"current_volume": 0},
		Rewards:         map[string]int{"market_stability_bonus": 120},
		Active:          true,
		RequiredServers: 3,
	}

	em.events[id] = event
	return event
}

// RegisterParticipant adds a server to an event.
func (em *EventManager) RegisterParticipant(eventID, serverID string) error {
	event, exists := em.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	if !event.Active {
		return fmt.Errorf("event is not active: %s", eventID)
	}

	for _, p := range event.Participants {
		if p == serverID {
			return fmt.Errorf("server already registered: %s", serverID)
		}
	}

	event.Participants = append(event.Participants, serverID)
	event.Progress[serverID] = 0
	return nil
}

// UpdateProgress updates progress for a participant in an event.
func (em *EventManager) UpdateProgress(eventID, serverID, goalKey string, amount int) error {
	event, exists := em.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	if !event.Active {
		return fmt.Errorf("event is not active: %s", eventID)
	}

	found := false
	for _, p := range event.Participants {
		if p == serverID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("server not registered for event: %s", serverID)
	}

	progressKey := serverID
	if event.Type == EventWorldThreat {
		progressKey = goalKey
	}

	event.Progress[progressKey] += amount
	return nil
}

// CheckCompletion checks if an event's goals have been met.
func (em *EventManager) CheckCompletion(eventID string) (bool, error) {
	event, exists := em.events[eventID]
	if !exists {
		return false, fmt.Errorf("event not found: %s", eventID)
	}

	now := time.Now().Unix()
	if now > event.EndTime {
		event.Active = false
		return false, nil
	}

	switch event.Type {
	case EventWorldThreat:
		damageDealt := event.Progress["damage_dealt"]
		requiredDamage := event.Goals["boss_damage"]
		if damageDealt >= requiredDamage {
			event.Active = false
			return true, nil
		}
	case EventServerVsServer:
		for server, progress := range event.Progress {
			if progress >= event.Goals["resources"] {
				event.Active = false
				return true, nil
			}
			_ = server
		}
	default:
		for _, progress := range event.Progress {
			allGoalsMet := true
			for goalKey, goalValue := range event.Goals {
				if progress < goalValue {
					allGoalsMet = false
					break
				}
				_ = goalKey
			}
			if allGoalsMet {
				event.Active = false
				return true, nil
			}
		}
	}

	return false, nil
}

// GetActiveEvents returns all currently active events.
func (em *EventManager) GetActiveEvents() []*MetaGameEvent {
	active := make([]*MetaGameEvent, 0)
	now := time.Now().Unix()

	for _, event := range em.events {
		if event.Active && now <= event.EndTime {
			active = append(active, event)
		}
	}

	return active
}

// GetEventsByType returns events of a specific type.
func (em *EventManager) GetEventsByType(eventType EventType) []*MetaGameEvent {
	events := make([]*MetaGameEvent, 0)

	for _, event := range em.events {
		if event.Type == eventType {
			events = append(events, event)
		}
	}

	return events
}

// GetEvent retrieves an event by ID.
func (em *EventManager) GetEvent(eventID string) (*MetaGameEvent, error) {
	event, exists := em.events[eventID]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}
	return event, nil
}

// Update processes event expiration.
func (em *EventManager) Update(deltaTime float64) {
	now := time.Now().Unix()

	for _, event := range em.events {
		if event.Active && now > event.EndTime {
			event.Active = false
		}
	}
}

// GetEventCount returns the number of events.
func (em *EventManager) GetEventCount() int {
	return len(em.events)
}

// GetActiveEventCount returns the number of active events.
func (em *EventManager) GetActiveEventCount() int {
	count := 0
	for _, event := range em.events {
		if event.Active {
			count++
		}
	}
	return count
}

// CancelEvent cancels an active event.
func (em *EventManager) CancelEvent(eventID string) error {
	event, exists := em.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	event.Active = false
	return nil
}
