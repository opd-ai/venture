// Package engine provides city evolution triggers for dynamic city state changes.
// EvolutionTriggers define events that affect city prosperity, infrastructure,
// population, and defense levels.
package engine

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// EvolutionTriggerType represents the category of city evolution trigger.
type EvolutionTriggerType string

const (
	// EvolutionTradeComplete occurs when a trade transaction completes in the city.
	EvolutionTradeComplete EvolutionTriggerType = "trade_complete"
	// EvolutionQuestComplete occurs when a quest is completed for the city.
	EvolutionQuestComplete EvolutionTriggerType = "quest_complete"
	// EvolutionRaidAttack occurs when the city is raided by enemies.
	EvolutionRaidAttack EvolutionTriggerType = "raid_attack"
	// EvolutionRaidDefended occurs when a raid is successfully repelled.
	EvolutionRaidDefended EvolutionTriggerType = "raid_defended"
	// EvolutionBuildingConstructed occurs when new infrastructure is built.
	EvolutionBuildingConstructed EvolutionTriggerType = "building_constructed"
	// EvolutionBuildingDestroyed occurs when infrastructure is destroyed.
	EvolutionBuildingDestroyed EvolutionTriggerType = "building_destroyed"
	// EvolutionPopulationArrived occurs when new inhabitants arrive.
	EvolutionPopulationArrived EvolutionTriggerType = "population_arrived"
	// EvolutionPopulationDeparted occurs when inhabitants leave the city.
	EvolutionPopulationDeparted EvolutionTriggerType = "population_departed"
	// EvolutionResourceDonation occurs when resources are donated to the city.
	EvolutionResourceDonation EvolutionTriggerType = "resource_donation"
	// EvolutionResourceShortage occurs when the city runs out of resources.
	EvolutionResourceShortage EvolutionTriggerType = "resource_shortage"
	// EvolutionGuardHired occurs when defensive forces are increased.
	EvolutionGuardHired EvolutionTriggerType = "guard_hired"
	// EvolutionGuardLost occurs when defensive forces are decreased.
	EvolutionGuardLost EvolutionTriggerType = "guard_lost"
)

// EvolutionTrigger represents an event that affects city evolution.
type EvolutionTrigger struct {
	// TriggerType identifies the category of trigger
	TriggerType EvolutionTriggerType `json:"trigger_type"`
	// Magnitude controls the strength of the effect (0.0-1.0)
	Magnitude float64 `json:"magnitude"`
	// SourceEntityID is the entity that caused this trigger
	SourceEntityID string `json:"source_entity_id"`
	// Timestamp when the trigger occurred
	Timestamp time.Time `json:"timestamp"`
	// Metadata stores additional trigger-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CityEvolutionTriggersComponent tracks pending evolution triggers for a city.
// Triggers are queued and processed by the CityEvolutionSystem.
type CityEvolutionTriggersComponent struct {
	// CityID links this component to its city
	CityID string `json:"city_id"`
	// PendingTriggers is the queue of triggers to process
	PendingTriggers []EvolutionTrigger `json:"pending_triggers"`
	// RecentTriggers is the history of processed triggers (last 50)
	RecentTriggers []EvolutionTrigger `json:"recent_triggers"`
	// MaxRecentTriggers is the maximum triggers to keep in history
	MaxRecentTriggers int `json:"max_recent_triggers"`
	// ProcessingEnabled controls whether triggers are processed
	ProcessingEnabled bool `json:"processing_enabled"`
}

// NewCityEvolutionTriggersComponent creates a new evolution triggers component.
func NewCityEvolutionTriggersComponent(cityID string) *CityEvolutionTriggersComponent {
	return &CityEvolutionTriggersComponent{
		CityID:            cityID,
		PendingTriggers:   make([]EvolutionTrigger, 0),
		RecentTriggers:    make([]EvolutionTrigger, 0),
		MaxRecentTriggers: 50,
		ProcessingEnabled: true,
	}
}

// Type returns the component type identifier.
func (c *CityEvolutionTriggersComponent) Type() string {
	return "city_evolution_triggers"
}

// QueueTrigger adds a new evolution trigger to the pending queue.
func (c *CityEvolutionTriggersComponent) QueueTrigger(trigger EvolutionTrigger) {
	if trigger.Timestamp.IsZero() {
		trigger.Timestamp = time.Now()
	}
	c.PendingTriggers = append(c.PendingTriggers, trigger)

	logrus.WithFields(logrus.Fields{
		"component_type": "city_evolution_triggers",
		"city_id":        c.CityID,
		"trigger_type":   trigger.TriggerType,
		"magnitude":      trigger.Magnitude,
	}).Debug("Queued evolution trigger")
}

// PopTrigger removes and returns the next pending trigger.
// Returns nil if no triggers are pending.
func (c *CityEvolutionTriggersComponent) PopTrigger() *EvolutionTrigger {
	if len(c.PendingTriggers) == 0 {
		return nil
	}
	trigger := c.PendingTriggers[0]
	c.PendingTriggers = c.PendingTriggers[1:]
	return &trigger
}

// HasPendingTriggers returns true if there are triggers waiting to be processed.
func (c *CityEvolutionTriggersComponent) HasPendingTriggers() bool {
	return len(c.PendingTriggers) > 0
}

// GetPendingCount returns the number of pending triggers.
func (c *CityEvolutionTriggersComponent) GetPendingCount() int {
	return len(c.PendingTriggers)
}

// RecordProcessed adds a trigger to the recent history.
func (c *CityEvolutionTriggersComponent) RecordProcessed(trigger EvolutionTrigger) {
	c.RecentTriggers = append(c.RecentTriggers, trigger)

	// Trim history if too long
	if len(c.RecentTriggers) > c.MaxRecentTriggers {
		c.RecentTriggers = c.RecentTriggers[len(c.RecentTriggers)-c.MaxRecentTriggers:]
	}
}

// GetRecentTriggersByType returns recent triggers of a specific type.
func (c *CityEvolutionTriggersComponent) GetRecentTriggersByType(triggerType EvolutionTriggerType) []EvolutionTrigger {
	result := make([]EvolutionTrigger, 0)
	for _, trigger := range c.RecentTriggers {
		if trigger.TriggerType == triggerType {
			result = append(result, trigger)
		}
	}
	return result
}

// ClearPending removes all pending triggers.
func (c *CityEvolutionTriggersComponent) ClearPending() {
	c.PendingTriggers = c.PendingTriggers[:0]
}

// Serialize encodes the component to bytes for persistence.
func (c *CityEvolutionTriggersComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "city_evolution_triggers",
		"city_id":        c.CityID,
		"pending_count":  len(c.PendingTriggers),
	}).Debug("Serializing city evolution triggers component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "city_evolution_triggers",
			"error":          err.Error(),
		}).Error("Failed to serialize city evolution triggers component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *CityEvolutionTriggersComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "city_evolution_triggers",
		"bytes":          len(data),
	}).Debug("Deserializing city evolution triggers component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "city_evolution_triggers",
			"error":          err.Error(),
		}).Error("Failed to deserialize city evolution triggers component")
		return err
	}
	return nil
}

// TriggerImpact defines how a trigger type affects city state.
type TriggerImpact struct {
	// ProsperityDelta is the change to prosperity (-1.0 to 1.0)
	ProsperityDelta float64
	// InfrastructureDelta is the change to infrastructure (-1.0 to 1.0)
	InfrastructureDelta float64
	// DefenseDelta is the change to defense (-1.0 to 1.0)
	DefenseDelta float64
	// PopulationDelta is the change to population (positive or negative)
	PopulationDelta int
	// ResourceDelta is the change to resource stockpile
	ResourceDelta float64
}

// GetTriggerImpact returns the standard impact for a trigger type.
// Impacts are scaled by the trigger magnitude.
func GetTriggerImpact(triggerType EvolutionTriggerType, magnitude float64) TriggerImpact {
	// Base impacts - will be multiplied by magnitude
	impacts := map[EvolutionTriggerType]TriggerImpact{
		EvolutionTradeComplete: {
			ProsperityDelta: 0.01,
			ResourceDelta:   10.0,
		},
		EvolutionQuestComplete: {
			ProsperityDelta: 0.02,
			DefenseDelta:    0.01,
		},
		EvolutionRaidAttack: {
			ProsperityDelta:     -0.05,
			InfrastructureDelta: -0.02,
			DefenseDelta:        -0.01,
			PopulationDelta:     -5,
			ResourceDelta:       -50.0,
		},
		EvolutionRaidDefended: {
			ProsperityDelta: 0.02,
			DefenseDelta:    0.02,
		},
		EvolutionBuildingConstructed: {
			ProsperityDelta:     0.03,
			InfrastructureDelta: 0.05,
			PopulationDelta:     10,
		},
		EvolutionBuildingDestroyed: {
			ProsperityDelta:     -0.03,
			InfrastructureDelta: -0.05,
			PopulationDelta:     -10,
		},
		EvolutionPopulationArrived: {
			ProsperityDelta: 0.01,
			PopulationDelta: 5,
		},
		EvolutionPopulationDeparted: {
			ProsperityDelta: -0.01,
			PopulationDelta: -5,
		},
		EvolutionResourceDonation: {
			ProsperityDelta: 0.02,
			ResourceDelta:   100.0,
		},
		EvolutionResourceShortage: {
			ProsperityDelta: -0.03,
			PopulationDelta: -3,
		},
		EvolutionGuardHired: {
			DefenseDelta:  0.03,
			ResourceDelta: -20.0,
		},
		EvolutionGuardLost: {
			DefenseDelta:    -0.03,
			ProsperityDelta: -0.01,
		},
	}

	impact, ok := impacts[triggerType]
	if !ok {
		return TriggerImpact{}
	}

	// Scale by magnitude
	return TriggerImpact{
		ProsperityDelta:     impact.ProsperityDelta * magnitude,
		InfrastructureDelta: impact.InfrastructureDelta * magnitude,
		DefenseDelta:        impact.DefenseDelta * magnitude,
		PopulationDelta:     int(float64(impact.PopulationDelta) * magnitude),
		ResourceDelta:       impact.ResourceDelta * magnitude,
	}
}
