// Package engine provides the city state component for dynamic city evolution.
// CityStateComponent enables cities to evolve based on player actions and world events,
// transitioning between struggling, stable, and thriving states.
package engine

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// CityState represents the current prosperity tier of a city.
type CityState string

const (
	// CityStateStruggling indicates a city in decline with poor services.
	CityStateStruggling CityState = "struggling"
	// CityStateStable indicates a balanced city with adequate services.
	CityStateStable CityState = "stable"
	// CityStateThriving indicates a prosperous city with excellent services.
	CityStateThriving CityState = "thriving"
)

// CityStateComponent tracks a city's prosperity, population, and infrastructure.
// State transitions occur based on prosperity thresholds:
// - struggling: prosperity < 0.3
// - stable: 0.3 <= prosperity < 0.7
// - thriving: prosperity >= 0.7
type CityStateComponent struct {
	// CityID is the unique identifier for this city
	CityID string `json:"city_id"`
	// CityName is the display name of the city
	CityName string `json:"city_name"`
	// Prosperity is the overall economic health (0.0-1.0)
	Prosperity float64 `json:"prosperity"`
	// Population is the current number of inhabitants
	Population int `json:"population"`
	// MaxPopulation is the population capacity based on infrastructure
	MaxPopulation int `json:"max_population"`
	// Infrastructure represents building quality and services (0.0-1.0)
	Infrastructure float64 `json:"infrastructure"`
	// Defense represents military/guard strength (0.0-1.0)
	Defense float64 `json:"defense"`
	// State is the current city state tier
	State CityState `json:"state"`
	// TradeVolume tracks recent economic activity (gold per in-game day)
	TradeVolume float64 `json:"trade_volume"`
	// ResourceStockpile tracks stored resources for city operations
	ResourceStockpile float64 `json:"resource_stockpile"`
	// Seed is the deterministic seed for this city's generation
	Seed int64 `json:"seed"`
}

// NewCityStateComponent creates a new city state component with default values.
func NewCityStateComponent(cityID, cityName string, seed int64) *CityStateComponent {
	return &CityStateComponent{
		CityID:            cityID,
		CityName:          cityName,
		Prosperity:        0.5, // Start stable
		Population:        100,
		MaxPopulation:     200,
		Infrastructure:    0.5,
		Defense:           0.5,
		State:             CityStateStable,
		TradeVolume:       0.0,
		ResourceStockpile: 100.0,
		Seed:              seed,
	}
}

// Type returns the component type identifier.
func (c *CityStateComponent) Type() string {
	return "city_state"
}

// UpdateState recalculates the city state based on prosperity.
// Returns true if state changed.
func (c *CityStateComponent) UpdateState() bool {
	oldState := c.State

	if c.Prosperity < 0.3 {
		c.State = CityStateStruggling
	} else if c.Prosperity >= 0.7 {
		c.State = CityStateThriving
	} else {
		c.State = CityStateStable
	}

	return c.State != oldState
}

// GetProsperityTier returns a human-readable prosperity description.
func (c *CityStateComponent) GetProsperityTier() string {
	switch c.State {
	case CityStateStruggling:
		return "Struggling"
	case CityStateThriving:
		return "Thriving"
	default:
		return "Stable"
	}
}

// GetPopulationRatio returns population as a fraction of max capacity.
func (c *CityStateComponent) GetPopulationRatio() float64 {
	if c.MaxPopulation <= 0 {
		return 0.0
	}
	return float64(c.Population) / float64(c.MaxPopulation)
}

// CanGrowPopulation returns true if the city can support more inhabitants.
func (c *CityStateComponent) CanGrowPopulation() bool {
	return c.Population < c.MaxPopulation && c.Prosperity >= 0.3
}

// IsOvercrowded returns true if population exceeds comfortable capacity.
func (c *CityStateComponent) IsOvercrowded() bool {
	return c.GetPopulationRatio() > 0.9
}

// Serialize encodes the component to bytes for persistence.
func (c *CityStateComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "city_state",
		"city_id":        c.CityID,
	}).Debug("Serializing city state component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "city_state",
			"error":          err.Error(),
		}).Error("Failed to serialize city state component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *CityStateComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "city_state",
		"bytes":          len(data),
	}).Debug("Deserializing city state component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "city_state",
			"error":          err.Error(),
		}).Error("Failed to deserialize city state component")
		return err
	}
	return nil
}
