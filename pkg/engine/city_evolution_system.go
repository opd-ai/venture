// Package engine provides the city evolution system for dynamic city state changes.
// CityEvolutionSystem processes evolution triggers and updates city states over time,
// causing cities to grow, decline, or stabilize based on player actions and world events.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CityEvolutionSystem processes city evolution triggers and updates city states.
// It manages prosperity, population, infrastructure, and defense levels
// based on player actions and world events.
type CityEvolutionSystem struct {
	world           *World
	clock           GameClock
	updateInterval  float64
	timeAccumulator float64
	logger          *logrus.Entry
}

// NewCityEvolutionSystem creates a new city evolution system.
func NewCityEvolutionSystem(world *World, clock GameClock) *CityEvolutionSystem {
	logger := logrus.WithField("system_name", "city_evolution")
	logger.Debug("Creating city evolution system")

	return &CityEvolutionSystem{
		world:           world,
		clock:           clock,
		updateInterval:  1.0, // Process triggers every second
		timeAccumulator: 0.0,
		logger:          logger,
	}
}

// Update processes all city entities and applies evolution changes.
func (s *CityEvolutionSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeAccumulator += deltaTime
	if s.timeAccumulator < s.updateInterval {
		return
	}
	s.timeAccumulator = 0.0

	for _, entity := range entities {
		if !entity.HasComponent("city_state") {
			continue
		}

		s.processCity(entity, deltaTime)
	}
}

// processCity handles evolution for a single city entity.
func (s *CityEvolutionSystem) processCity(entity *Entity, deltaTime float64) {
	cityStateComp, ok := entity.GetComponent("city_state")
	if !ok || cityStateComp == nil {
		return
	}
	cityState := cityStateComp.(*CityStateComponent)

	// Process pending triggers if triggers component exists
	if triggersComp, ok := entity.GetComponent("city_evolution_triggers"); ok && triggersComp != nil {
		triggers := triggersComp.(*CityEvolutionTriggersComponent)
		if triggers.ProcessingEnabled {
			s.processTriggers(cityState, triggers)
		}
	}

	// Apply natural decay/growth
	s.applyNaturalEvolution(cityState, deltaTime)

	// Update state based on new prosperity
	if cityState.UpdateState() {
		s.logger.WithFields(logrus.Fields{
			"city_id":   cityState.CityID,
			"city_name": cityState.CityName,
			"new_state": cityState.State,
		}).Info("City state changed")
	}
}

// processTriggers applies all pending evolution triggers to a city.
func (s *CityEvolutionSystem) processTriggers(cityState *CityStateComponent, triggers *CityEvolutionTriggersComponent) {
	processedCount := 0

	for triggers.HasPendingTriggers() {
		trigger := triggers.PopTrigger()
		if trigger == nil {
			break
		}

		impact := GetTriggerImpact(trigger.TriggerType, trigger.Magnitude)
		s.applyImpact(cityState, impact)
		triggers.RecordProcessed(*trigger)
		processedCount++

		s.logger.WithFields(logrus.Fields{
			"city_id":      cityState.CityID,
			"trigger_type": trigger.TriggerType,
			"magnitude":    trigger.Magnitude,
			"prosperity":   cityState.Prosperity,
		}).Debug("Applied evolution trigger")
	}

	if processedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"city_id": cityState.CityID,
			"count":   processedCount,
		}).Debug("Processed evolution triggers")
	}
}

// applyImpact applies a trigger impact to the city state.
func (s *CityEvolutionSystem) applyImpact(cityState *CityStateComponent, impact TriggerImpact) {
	// Apply deltas with clamping
	cityState.Prosperity = clampFloat(cityState.Prosperity+impact.ProsperityDelta, 0.0, 1.0)
	cityState.Infrastructure = clampFloat(cityState.Infrastructure+impact.InfrastructureDelta, 0.0, 1.0)
	cityState.Defense = clampFloat(cityState.Defense+impact.DefenseDelta, 0.0, 1.0)

	// Apply population change
	cityState.Population += impact.PopulationDelta
	if cityState.Population < 0 {
		cityState.Population = 0
	}
	if cityState.Population > cityState.MaxPopulation {
		cityState.Population = cityState.MaxPopulation
	}

	// Apply resource change
	cityState.ResourceStockpile += impact.ResourceDelta
	if cityState.ResourceStockpile < 0 {
		cityState.ResourceStockpile = 0
	}
}

// applyNaturalEvolution applies slow natural changes based on city state.
func (s *CityEvolutionSystem) applyNaturalEvolution(cityState *CityStateComponent, deltaTime float64) {
	// Natural prosperity decay/growth rate per in-game day
	// Thriving cities grow slowly, struggling cities decline slowly
	const prosperityDecayRate = 0.001 // Per update cycle

	switch cityState.State {
	case CityStateThriving:
		// Thriving cities attract population and maintain prosperity
		if cityState.CanGrowPopulation() {
			cityState.Population++
		}
		// Slight prosperity maintenance cost
		cityState.ResourceStockpile -= 1.0

	case CityStateStruggling:
		// Struggling cities lose population and decline
		if cityState.Population > 10 {
			cityState.Population--
		}
		// Prosperity continues to decline
		cityState.Prosperity = clampFloat(cityState.Prosperity-prosperityDecayRate, 0.0, 1.0)

	case CityStateStable:
		// Stable cities slowly trend toward average
		if cityState.Prosperity < 0.5 {
			cityState.Prosperity = clampFloat(cityState.Prosperity+prosperityDecayRate*0.5, 0.0, 1.0)
		} else if cityState.Prosperity > 0.5 {
			cityState.Prosperity = clampFloat(cityState.Prosperity-prosperityDecayRate*0.5, 0.0, 1.0)
		}
	}

	// Infrastructure decay if resources are depleted
	if cityState.ResourceStockpile <= 0 {
		cityState.Infrastructure = clampFloat(cityState.Infrastructure-0.001, 0.0, 1.0)
	}

	// Update max population based on infrastructure
	cityState.MaxPopulation = 100 + int(cityState.Infrastructure*400)
}

// clampFloat restricts a value to the given range.
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// TriggerTradeEvent creates a trade completion trigger for a city.
func (s *CityEvolutionSystem) TriggerTradeEvent(entity *Entity, tradeValue float64, sourceEntityID string) {
	s.queueTriggerForEntity(entity, EvolutionTrigger{
		TriggerType:    EvolutionTradeComplete,
		Magnitude:      math.Min(tradeValue/1000.0, 1.0), // Normalize trade value
		SourceEntityID: sourceEntityID,
	})
}

// EvolutionQuestComplete creates a quest completion trigger for a city.
func (s *CityEvolutionSystem) EvolutionQuestComplete(entity *Entity, questDifficulty float64, sourceEntityID string) {
	s.queueTriggerForEntity(entity, EvolutionTrigger{
		TriggerType:    EvolutionQuestComplete,
		Magnitude:      questDifficulty,
		SourceEntityID: sourceEntityID,
	})
}

// TriggerRaid creates a raid event trigger for a city.
func (s *CityEvolutionSystem) TriggerRaid(entity *Entity, raidStrength float64, defended bool, sourceEntityID string) {
	triggerType := EvolutionRaidAttack
	if defended {
		triggerType = EvolutionRaidDefended
	}
	s.queueTriggerForEntity(entity, EvolutionTrigger{
		TriggerType:    triggerType,
		Magnitude:      raidStrength,
		SourceEntityID: sourceEntityID,
	})
}

// TriggerBuildingChange creates a building construction/destruction trigger.
func (s *CityEvolutionSystem) TriggerBuildingChange(entity *Entity, constructed bool, buildingValue float64, sourceEntityID string) {
	triggerType := EvolutionBuildingDestroyed
	if constructed {
		triggerType = EvolutionBuildingConstructed
	}
	s.queueTriggerForEntity(entity, EvolutionTrigger{
		TriggerType:    triggerType,
		Magnitude:      buildingValue,
		SourceEntityID: sourceEntityID,
	})
}

// EvolutionResourceDonation creates a resource donation trigger.
func (s *CityEvolutionSystem) EvolutionResourceDonation(entity *Entity, amount float64, sourceEntityID string) {
	s.queueTriggerForEntity(entity, EvolutionTrigger{
		TriggerType:    EvolutionResourceDonation,
		Magnitude:      math.Min(amount/500.0, 1.0),
		SourceEntityID: sourceEntityID,
	})
}

// queueTriggerForEntity adds a trigger to an entity's evolution triggers component.
func (s *CityEvolutionSystem) queueTriggerForEntity(entity *Entity, trigger EvolutionTrigger) {
	triggersComp, ok := entity.GetComponent("city_evolution_triggers")
	if !ok || triggersComp == nil {
		// Create triggers component if it doesn't exist
		triggers := NewCityEvolutionTriggersComponent("")
		if cityStateComp, ok := entity.GetComponent("city_state"); ok && cityStateComp != nil {
			triggers.CityID = cityStateComp.(*CityStateComponent).CityID
		}
		entity.AddComponent(triggers)
		triggersComp = triggers
	}
	triggersComp.(*CityEvolutionTriggersComponent).QueueTrigger(trigger)
}

// GenerateCity creates a new city entity with city state and evolution triggers.
// Uses deterministic generation based on seed.
func GenerateCity(world *World, cityID, cityName string, seed int64, x, y float64) *Entity {
	rng := rand.New(rand.NewSource(seed))

	entity := world.CreateEntity()

	// Add position component
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Create city state with seed-based initial values
	cityState := NewCityStateComponent(cityID, cityName, seed)
	cityState.Prosperity = 0.3 + rng.Float64()*0.4     // 0.3-0.7
	cityState.Infrastructure = 0.3 + rng.Float64()*0.4 // 0.3-0.7
	cityState.Defense = 0.2 + rng.Float64()*0.3        // 0.2-0.5
	cityState.Population = 50 + rng.Intn(100)          // 50-150
	cityState.MaxPopulation = 150 + rng.Intn(100)      // 150-250
	cityState.ResourceStockpile = 50.0 + rng.Float64()*100.0
	cityState.UpdateState()

	entity.AddComponent(cityState)
	entity.AddComponent(NewCityEvolutionTriggersComponent(cityID))

	logrus.WithFields(logrus.Fields{
		"system_name":   "city_evolution",
		"city_id":       cityID,
		"city_name":     cityName,
		"seed":          seed,
		"prosperity":    cityState.Prosperity,
		"initial_state": cityState.State,
	}).Info("Generated city")

	return entity
}

// GetCityByID finds a city entity by its city ID.
func (s *CityEvolutionSystem) GetCityByID(cityID string) *Entity {
	for _, entity := range s.world.GetEntitiesWith("city_state") {
		cityStateComp, ok := entity.GetComponent("city_state")
		if !ok || cityStateComp == nil {
			continue
		}
		cityState := cityStateComp.(*CityStateComponent)
		if cityState.CityID == cityID {
			return entity
		}
	}
	return nil
}

// GetCitiesInState returns all city entities in a specific state.
func (s *CityEvolutionSystem) GetCitiesInState(state CityState) []*Entity {
	result := make([]*Entity, 0)
	for _, entity := range s.world.GetEntitiesWith("city_state") {
		cityStateComp, ok := entity.GetComponent("city_state")
		if !ok || cityStateComp == nil {
			continue
		}
		cityState := cityStateComp.(*CityStateComponent)
		if cityState.State == state {
			result = append(result, entity)
		}
	}
	return result
}
