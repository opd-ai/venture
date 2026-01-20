// Package engine provides the world persistence system for managing world state
// across save/load cycles. This system coordinates saving city states, NPC states,
// and world events, and handles time progression when the player returns.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WorldPersistenceSystem manages world state persistence and restoration.
// It coordinates with WorldMemoryComponent to save and load all living world data,
// and handles time progression to advance the world while the player was away.
type WorldPersistenceSystem struct {
	logger *logrus.Entry
}

// NewWorldPersistenceSystem creates a new world persistence system.
func NewWorldPersistenceSystem() *WorldPersistenceSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "world_persistence",
	})
	logger.Debug("Creating world persistence system")

	return &WorldPersistenceSystem{
		logger: logger,
	}
}

// Update processes entities with world memory components.
// This system primarily operates on save/load events, not per-frame.
func (s *WorldPersistenceSystem) Update(entities []*Entity, deltaTime float64) {
	// Per-frame updates are minimal - most work happens in Save/Load methods
}

// SaveWorldState saves all living world state to the WorldMemoryComponent.
// This should be called before game save.
func (s *WorldPersistenceSystem) SaveWorldState(
	worldMemory *WorldMemoryComponent,
	cities []*Entity,
	npcs []*Entity,
	currentGameTime float64,
) error {
	if worldMemory == nil {
		return fmt.Errorf("world memory component is nil")
	}

	s.logger.WithFields(logrus.Fields{
		"city_count": len(cities),
		"npc_count":  len(npcs),
		"game_time":  currentGameTime,
	}).Info("Saving world state")

	// Update last save time
	worldMemory.LastSaveTime = currentGameTime

	// Save city states
	for _, cityEntity := range cities {
		if cityState := getCityStateComponent(cityEntity); cityState != nil {
			worldMemory.SaveCityState(cityState)
		}
	}

	// Save NPC states
	for _, npcEntity := range npcs {
		s.saveNPCState(worldMemory, npcEntity)
	}

	s.logger.WithFields(logrus.Fields{
		"saved_cities": worldMemory.GetCityCount(),
		"saved_npcs":   worldMemory.GetNPCCount(),
	}).Info("World state saved")

	return nil
}

// LoadWorldState restores living world state from the WorldMemoryComponent.
// This should be called after game load.
func (s *WorldPersistenceSystem) LoadWorldState(
	worldMemory *WorldMemoryComponent,
	cities []*Entity,
	npcs []*Entity,
	currentGameTime float64,
) error {
	if worldMemory == nil {
		return fmt.Errorf("world memory component is nil")
	}

	s.logger.WithFields(logrus.Fields{
		"city_count": len(cities),
		"npc_count":  len(npcs),
		"game_time":  currentGameTime,
	}).Info("Loading world state")

	// Calculate time progression if enabled
	elapsedTime := currentGameTime - worldMemory.LastSaveTime
	if elapsedTime < 0 {
		elapsedTime = 0 // Handle edge case of time going backwards
	}

	timeAdvancement := worldMemory.CalculateTimeProgression(currentGameTime, elapsedTime)
	if timeAdvancement > 0 {
		s.logger.WithFields(logrus.Fields{
			"elapsed_time":     elapsedTime,
			"time_advancement": timeAdvancement,
		}).Info("Applying time progression")
	}

	// Restore city states
	for _, cityEntity := range cities {
		s.loadCityState(worldMemory, cityEntity, timeAdvancement)
	}

	// Restore NPC states
	for _, npcEntity := range npcs {
		s.loadNPCState(worldMemory, npcEntity)
	}

	s.logger.Info("World state loaded")

	return nil
}

// saveNPCState saves a single NPC's state to world memory.
func (s *WorldPersistenceSystem) saveNPCState(worldMemory *WorldMemoryComponent, npcEntity *Entity) {
	schedule := getScheduleComponent(npcEntity)
	if schedule == nil {
		return
	}

	position := getPositionComponent(npcEntity)
	var x, y float64
	if position != nil {
		x = position.X
		y = position.Y
	}

	name := getEntityName(npcEntity)
	entityID := fmt.Sprintf("%d", npcEntity.ID)

	worldMemory.SaveNPCState(entityID, name, x, y, schedule)
}

// loadCityState restores a city's state from world memory and applies time progression.
func (s *WorldPersistenceSystem) loadCityState(
	worldMemory *WorldMemoryComponent,
	cityEntity *Entity,
	timeAdvancement float64,
) {
	cityState := getCityStateComponent(cityEntity)
	if cityState == nil {
		return
	}

	savedState := worldMemory.LoadCityState(cityState.CityID)
	if savedState == nil {
		return
	}

	// Restore saved values
	cityState.Prosperity = savedState.Prosperity
	cityState.Population = savedState.Population
	cityState.MaxPopulation = savedState.MaxPopulation
	cityState.Infrastructure = savedState.Infrastructure
	cityState.Defense = savedState.Defense
	cityState.State = savedState.State
	cityState.TradeVolume = savedState.TradeVolume
	cityState.ResourceStockpile = savedState.ResourceStockpile

	// Apply time progression effects
	if timeAdvancement > 0 {
		s.applyTimeProgressionToCity(worldMemory, cityState, timeAdvancement)
	}
}

// loadNPCState restores an NPC's state from world memory.
func (s *WorldPersistenceSystem) loadNPCState(
	worldMemory *WorldMemoryComponent,
	npcEntity *Entity,
) {
	schedule := getScheduleComponent(npcEntity)
	if schedule == nil {
		return
	}

	entityID := fmt.Sprintf("%d", npcEntity.ID)
	savedState := worldMemory.LoadNPCState(entityID)
	if savedState == nil {
		return
	}

	// Restore schedule state
	schedule.CurrentActivityIdx = savedState.CurrentActivityIdx
	schedule.IsMoving = savedState.IsMoving

	// Restore position if available
	position := getPositionComponent(npcEntity)
	if position != nil {
		position.X = savedState.X
		position.Y = savedState.Y
	}
}

// applyTimeProgressionToCity simulates city evolution during offline time.
func (s *WorldPersistenceSystem) applyTimeProgressionToCity(
	worldMemory *WorldMemoryComponent,
	cityState *CityStateComponent,
	timeAdvancement float64,
) {
	// Use city seed for deterministic progression
	rng := rand.New(rand.NewSource(cityState.Seed + int64(timeAdvancement*1000)))

	// Calculate days elapsed (assuming 1 game-hour = 1 real-second for this calculation)
	daysElapsed := timeAdvancement / 24.0

	// Natural evolution based on state
	evolutionRate := 0.01 * daysElapsed // 1% change per day base rate

	if cityState.State == CityStateThriving {
		// Thriving cities grow slowly
		cityState.Prosperity += evolutionRate * (0.5 + rng.Float64()*0.5)
		if cityState.Population < cityState.MaxPopulation {
			growth := int(daysElapsed * (2 + rng.Float64()*3))
			cityState.Population = min(cityState.Population+growth, cityState.MaxPopulation)
		}
	} else if cityState.State == CityStateStruggling {
		// Struggling cities decline
		cityState.Prosperity -= evolutionRate * (0.5 + rng.Float64()*0.5)
		departure := int(daysElapsed * (1 + rng.Float64()*2))
		cityState.Population = max(cityState.Population-departure, 10) // Minimum 10 pop
	}

	// Clamp prosperity
	if cityState.Prosperity > 1.0 {
		cityState.Prosperity = 1.0
	} else if cityState.Prosperity < 0.0 {
		cityState.Prosperity = 0.0
	}

	// Update state based on new prosperity
	stateChanged := updateCityStateFromProsperity(cityState)

	// Record significant changes as events
	if stateChanged {
		event := WorldEventRecord{
			EventID:        uuid.New().String(),
			EventType:      "city_evolution",
			Description:    fmt.Sprintf("%s evolved to %s", cityState.CityName, cityState.State),
			GameTime:       worldMemory.LastSaveTime + timeAdvancement,
			AffectedCityID: cityState.CityID,
			Magnitude:      0.7,
			Details: map[string]interface{}{
				"new_state":    string(cityState.State),
				"prosperity":   cityState.Prosperity,
				"time_elapsed": timeAdvancement,
				"from_offline": true,
			},
		}
		worldMemory.RecordEvent(event)

		s.logger.WithFields(logrus.Fields{
			"city_id":    cityState.CityID,
			"city_name":  cityState.CityName,
			"new_state":  cityState.State,
			"prosperity": cityState.Prosperity,
		}).Info("City evolved during offline time")
	}
}

// RecordWorldEvent adds a significant event to world history.
func (s *WorldPersistenceSystem) RecordWorldEvent(
	worldMemory *WorldMemoryComponent,
	eventType string,
	description string,
	gameTime float64,
	cityID string,
	playerID string,
	magnitude float64,
	details map[string]interface{},
) {
	if worldMemory == nil {
		return
	}

	event := WorldEventRecord{
		EventID:          uuid.New().String(),
		EventType:        eventType,
		Description:      description,
		GameTime:         gameTime,
		AffectedCityID:   cityID,
		AffectedPlayerID: playerID,
		Magnitude:        magnitude,
		Details:          details,
	}

	worldMemory.RecordEvent(event)

	s.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": eventType,
		"city_id":    cityID,
		"magnitude":  magnitude,
	}).Debug("Recorded world event")
}

// UpdatePlayerReputation modifies a player's reputation in a city and records the change.
func (s *WorldPersistenceSystem) UpdatePlayerReputation(
	worldMemory *WorldMemoryComponent,
	playerID string,
	cityID string,
	delta float64,
	reason string,
	gameTime float64,
) {
	if worldMemory == nil {
		return
	}

	oldRep := worldMemory.GetPlayerCityReputation(playerID, cityID)
	worldMemory.AdjustPlayerCityReputation(playerID, cityID, delta)
	newRep := worldMemory.GetPlayerCityReputation(playerID, cityID)

	// Record significant reputation changes
	if absFloat64(delta) >= 10.0 {
		event := WorldEventRecord{
			EventID:          uuid.New().String(),
			EventType:        "reputation_change",
			Description:      reason,
			GameTime:         gameTime,
			AffectedCityID:   cityID,
			AffectedPlayerID: playerID,
			Magnitude:        absFloat64(delta) / 100.0,
			Details: map[string]interface{}{
				"old_reputation": oldRep,
				"new_reputation": newRep,
				"delta":          delta,
			},
		}
		worldMemory.RecordEvent(event)
	}

	s.logger.WithFields(logrus.Fields{
		"player_id":      playerID,
		"city_id":        cityID,
		"delta":          delta,
		"new_reputation": newRep,
		"reason":         reason,
	}).Debug("Updated player reputation")
}

// Helper functions

func getCityStateComponent(entity *Entity) *CityStateComponent {
	if entity == nil {
		return nil
	}
	comp, ok := entity.GetComponent("city_state")
	if !ok || comp == nil {
		return nil
	}
	if cityState, ok := comp.(*CityStateComponent); ok {
		return cityState
	}
	return nil
}

func getScheduleComponent(entity *Entity) *ScheduleComponent {
	if entity == nil {
		return nil
	}
	comp, ok := entity.GetComponent("schedule")
	if !ok || comp == nil {
		return nil
	}
	if schedule, ok := comp.(*ScheduleComponent); ok {
		return schedule
	}
	return nil
}

func getPositionComponent(entity *Entity) *PositionComponent {
	if entity == nil {
		return nil
	}
	comp, ok := entity.GetComponent("position")
	if !ok || comp == nil {
		return nil
	}
	if pos, ok := comp.(*PositionComponent); ok {
		return pos
	}
	return nil
}

func getEntityName(entity *Entity) string {
	if entity == nil {
		return ""
	}
	return fmt.Sprintf("Entity-%d", entity.ID)
}

func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
