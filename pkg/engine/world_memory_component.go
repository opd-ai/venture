// Package engine provides world memory components for persistent world state.
// WorldMemoryComponent tracks city states, NPC states, and world events
// across game sessions, enabling a living world that evolves over time.
package engine

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// WorldMemoryComponent persists world state across save/load cycles.
// It tracks city states, NPC schedules, world events, and player reputation
// to create a living world that remembers and evolves.
type WorldMemoryComponent struct {
	// WorldSeed is the deterministic seed for this world
	WorldSeed int64 `json:"world_seed"`
	// LastSaveTime is the in-game time when world was last saved
	LastSaveTime float64 `json:"last_save_time"`
	// CityStates stores serialized city state data by city ID
	CityStates map[string]CityStateData `json:"city_states"`
	// NPCStates stores serialized NPC state data by entity ID
	NPCStates map[string]NPCStateData `json:"npc_states"`
	// EventHistory stores recent significant world events
	EventHistory []WorldEventRecord `json:"event_history"`
	// MaxEventHistory is the maximum events to keep in history
	MaxEventHistory int `json:"max_event_history"`
	// PlayerReputations stores per-city reputation by player ID
	PlayerReputations map[string]map[string]float64 `json:"player_reputations"`
	// TimeProgressionEnabled allows world to advance while player is away
	TimeProgressionEnabled bool `json:"time_progression_enabled"`
	// TimeProgressionRate is the multiplier for offline time advancement
	TimeProgressionRate float64 `json:"time_progression_rate"`
}

// CityStateData represents serialized city state for persistence.
type CityStateData struct {
	CityID            string  `json:"city_id"`
	CityName          string  `json:"city_name"`
	Prosperity        float64 `json:"prosperity"`
	Population        int     `json:"population"`
	MaxPopulation     int     `json:"max_population"`
	Infrastructure    float64 `json:"infrastructure"`
	Defense           float64 `json:"defense"`
	State             string  `json:"state"`
	TradeVolume       float64 `json:"trade_volume"`
	ResourceStockpile float64 `json:"resource_stockpile"`
	Seed              int64   `json:"seed"`
}

// NPCStateData represents serialized NPC state for persistence.
type NPCStateData struct {
	EntityID           string              `json:"entity_id"`
	Name               string              `json:"name"`
	X                  float64             `json:"x"`
	Y                  float64             `json:"y"`
	HomeX              float64             `json:"home_x"`
	HomeY              float64             `json:"home_y"`
	CurrentActivityIdx int                 `json:"current_activity_idx"`
	Schedule           []ScheduledActivity `json:"schedule,omitempty"`
	IsMoving           bool                `json:"is_moving"`
	LastUpdateHour     int                 `json:"last_update_hour"`
}

// WorldEventRecord represents a significant world event for history tracking.
type WorldEventRecord struct {
	// EventID is a unique identifier for this event
	EventID string `json:"event_id"`
	// EventType categorizes the event (city_evolved, raid, trade, etc.)
	EventType string `json:"event_type"`
	// Description is a human-readable event description
	Description string `json:"description"`
	// GameTime is the in-game time when the event occurred
	GameTime float64 `json:"game_time"`
	// AffectedCityID is the city involved (if applicable)
	AffectedCityID string `json:"affected_city_id,omitempty"`
	// AffectedPlayerID is the player involved (if applicable)
	AffectedPlayerID string `json:"affected_player_id,omitempty"`
	// Magnitude represents the significance of the event (0.0-1.0)
	Magnitude float64 `json:"magnitude"`
	// Details stores event-specific data
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewWorldMemoryComponent creates a new world memory component.
func NewWorldMemoryComponent(worldSeed int64) *WorldMemoryComponent {
	return &WorldMemoryComponent{
		WorldSeed:              worldSeed,
		LastSaveTime:           0.0,
		CityStates:             make(map[string]CityStateData),
		NPCStates:              make(map[string]NPCStateData),
		EventHistory:           make([]WorldEventRecord, 0),
		MaxEventHistory:        100,
		PlayerReputations:      make(map[string]map[string]float64),
		TimeProgressionEnabled: false,
		TimeProgressionRate:    0.1, // 10% of real time passes while offline
	}
}

// Type returns the component type identifier.
func (w *WorldMemoryComponent) Type() string {
	return "world_memory"
}

// SaveCityState stores a city's current state for persistence.
func (w *WorldMemoryComponent) SaveCityState(city *CityStateComponent) {
	if city == nil {
		return
	}
	w.CityStates[city.CityID] = CityStateData{
		CityID:            city.CityID,
		CityName:          city.CityName,
		Prosperity:        city.Prosperity,
		Population:        city.Population,
		MaxPopulation:     city.MaxPopulation,
		Infrastructure:    city.Infrastructure,
		Defense:           city.Defense,
		State:             string(city.State),
		TradeVolume:       city.TradeVolume,
		ResourceStockpile: city.ResourceStockpile,
		Seed:              city.Seed,
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"city_id":        city.CityID,
		"prosperity":     city.Prosperity,
		"state":          city.State,
	}).Debug("Saved city state")
}

// LoadCityState restores a city's state from persistence.
// Returns nil if no saved state exists.
func (w *WorldMemoryComponent) LoadCityState(cityID string) *CityStateComponent {
	data, ok := w.CityStates[cityID]
	if !ok {
		return nil
	}

	city := &CityStateComponent{
		CityID:            data.CityID,
		CityName:          data.CityName,
		Prosperity:        data.Prosperity,
		Population:        data.Population,
		MaxPopulation:     data.MaxPopulation,
		Infrastructure:    data.Infrastructure,
		Defense:           data.Defense,
		State:             CityState(data.State),
		TradeVolume:       data.TradeVolume,
		ResourceStockpile: data.ResourceStockpile,
		Seed:              data.Seed,
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"city_id":        cityID,
		"prosperity":     city.Prosperity,
	}).Debug("Loaded city state")

	return city
}

// SaveNPCState stores an NPC's current state for persistence.
func (w *WorldMemoryComponent) SaveNPCState(entityID, name string, x, y float64, schedule *ScheduleComponent) {
	if schedule == nil {
		return
	}

	w.NPCStates[entityID] = NPCStateData{
		EntityID:           entityID,
		Name:               name,
		X:                  x,
		Y:                  y,
		HomeX:              schedule.HomeX,
		HomeY:              schedule.HomeY,
		CurrentActivityIdx: schedule.CurrentActivityIdx,
		Schedule:           schedule.Activities,
		IsMoving:           schedule.IsMoving,
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"entity_id":      entityID,
		"name":           name,
	}).Debug("Saved NPC state")
}

// LoadNPCState restores an NPC's state from persistence.
// Returns nil if no saved state exists.
func (w *WorldMemoryComponent) LoadNPCState(entityID string) *NPCStateData {
	data, ok := w.NPCStates[entityID]
	if !ok {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"entity_id":      entityID,
		"name":           data.Name,
	}).Debug("Loaded NPC state")

	return &data
}

// RecordEvent adds a world event to history.
func (w *WorldMemoryComponent) RecordEvent(event WorldEventRecord) {
	w.EventHistory = append(w.EventHistory, event)

	// Trim history if too long
	if len(w.EventHistory) > w.MaxEventHistory {
		w.EventHistory = w.EventHistory[len(w.EventHistory)-w.MaxEventHistory:]
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"event_id":       event.EventID,
		"event_type":     event.EventType,
		"magnitude":      event.Magnitude,
	}).Debug("Recorded world event")
}

// GetRecentEvents returns the N most recent world events.
func (w *WorldMemoryComponent) GetRecentEvents(count int) []WorldEventRecord {
	if count <= 0 {
		return []WorldEventRecord{}
	}

	total := len(w.EventHistory)
	if total == 0 {
		return []WorldEventRecord{}
	}

	if count > total {
		count = total
	}

	return w.EventHistory[total-count:]
}

// GetEventsByCity returns world events affecting a specific city.
func (w *WorldMemoryComponent) GetEventsByCity(cityID string) []WorldEventRecord {
	result := make([]WorldEventRecord, 0)
	for _, event := range w.EventHistory {
		if event.AffectedCityID == cityID {
			result = append(result, event)
		}
	}
	return result
}

// GetEventsByType returns world events of a specific type.
func (w *WorldMemoryComponent) GetEventsByType(eventType string) []WorldEventRecord {
	result := make([]WorldEventRecord, 0)
	for _, event := range w.EventHistory {
		if event.EventType == eventType {
			result = append(result, event)
		}
	}
	return result
}

// SetPlayerCityReputation sets a player's reputation in a specific city.
func (w *WorldMemoryComponent) SetPlayerCityReputation(playerID, cityID string, reputation float64) {
	// Clamp reputation to [-100, 100]
	if reputation > 100.0 {
		reputation = 100.0
	} else if reputation < -100.0 {
		reputation = -100.0
	}

	if w.PlayerReputations[playerID] == nil {
		w.PlayerReputations[playerID] = make(map[string]float64)
	}
	w.PlayerReputations[playerID][cityID] = reputation

	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"player_id":      playerID,
		"city_id":        cityID,
		"reputation":     reputation,
	}).Debug("Set player city reputation")
}

// GetPlayerCityReputation returns a player's reputation in a specific city.
// Returns 0 (neutral) if no reputation is recorded.
func (w *WorldMemoryComponent) GetPlayerCityReputation(playerID, cityID string) float64 {
	if playerReps, ok := w.PlayerReputations[playerID]; ok {
		if rep, ok := playerReps[cityID]; ok {
			return rep
		}
	}
	return 0.0
}

// AdjustPlayerCityReputation modifies a player's reputation in a city.
func (w *WorldMemoryComponent) AdjustPlayerCityReputation(playerID, cityID string, delta float64) {
	current := w.GetPlayerCityReputation(playerID, cityID)
	w.SetPlayerCityReputation(playerID, cityID, current+delta)
}

// GetPlayerAllCityReputations returns all city reputations for a player.
func (w *WorldMemoryComponent) GetPlayerAllCityReputations(playerID string) map[string]float64 {
	if playerReps, ok := w.PlayerReputations[playerID]; ok {
		// Return a copy to prevent modification
		result := make(map[string]float64, len(playerReps))
		for k, v := range playerReps {
			result[k] = v
		}
		return result
	}
	return make(map[string]float64)
}

// GetReputationTier returns a string describing the reputation level.
func (w *WorldMemoryComponent) GetReputationTier(reputation float64) string {
	if reputation >= 75.0 {
		return "Revered"
	} else if reputation >= 50.0 {
		return "Honored"
	} else if reputation >= 25.0 {
		return "Friendly"
	} else if reputation > -25.0 {
		return "Neutral"
	} else if reputation > -50.0 {
		return "Unfriendly"
	} else if reputation > -75.0 {
		return "Hostile"
	}
	return "Hated"
}

// CalculateTimeProgression calculates how much the world should advance
// based on elapsed real time since last save.
func (w *WorldMemoryComponent) CalculateTimeProgression(currentGameTime, elapsedRealSeconds float64) float64 {
	if !w.TimeProgressionEnabled {
		return 0.0
	}
	return elapsedRealSeconds * w.TimeProgressionRate
}

// GetCityCount returns the number of saved city states.
func (w *WorldMemoryComponent) GetCityCount() int {
	return len(w.CityStates)
}

// GetNPCCount returns the number of saved NPC states.
func (w *WorldMemoryComponent) GetNPCCount() int {
	return len(w.NPCStates)
}

// GetEventCount returns the number of recorded world events.
func (w *WorldMemoryComponent) GetEventCount() int {
	return len(w.EventHistory)
}

// ClearHistory removes all event history.
func (w *WorldMemoryComponent) ClearHistory() {
	w.EventHistory = w.EventHistory[:0]
}

// Serialize encodes the component to bytes for persistence.
func (w *WorldMemoryComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"city_count":     len(w.CityStates),
		"npc_count":      len(w.NPCStates),
		"event_count":    len(w.EventHistory),
	}).Debug("Serializing world memory component")

	data, err := json.Marshal(w)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "world_memory",
			"error":          err.Error(),
		}).Error("Failed to serialize world memory component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (w *WorldMemoryComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "world_memory",
		"bytes":          len(data),
	}).Debug("Deserializing world memory component")

	if err := json.Unmarshal(data, w); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "world_memory",
			"error":          err.Error(),
		}).Error("Failed to deserialize world memory component")
		return err
	}

	// Initialize nil maps after deserialization
	if w.CityStates == nil {
		w.CityStates = make(map[string]CityStateData)
	}
	if w.NPCStates == nil {
		w.NPCStates = make(map[string]NPCStateData)
	}
	if w.PlayerReputations == nil {
		w.PlayerReputations = make(map[string]map[string]float64)
	}

	return nil
}
