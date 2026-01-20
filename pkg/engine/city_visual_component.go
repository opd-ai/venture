// Package engine provides the city visual component for state-based city rendering.
// CityVisualComponent stores visual parameters that change based on city state,
// allowing the rendering system to display cities differently based on prosperity.
package engine

import (
	"encoding/json"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CityVisualStyle defines the visual appearance category for a city.
type CityVisualStyle string

const (
	// VisualStyleRundown represents dilapidated buildings, broken roads.
	VisualStyleRundown CityVisualStyle = "rundown"
	// VisualStyleModest represents simple but maintained buildings.
	VisualStyleModest CityVisualStyle = "modest"
	// VisualStyleProsperous represents grand architecture, decorations.
	VisualStyleProsperous CityVisualStyle = "prosperous"
)

// CityVisualComponent stores visual parameters for city rendering.
// These values change based on the city's prosperity state.
type CityVisualComponent struct {
	// CityID links to the city state component
	CityID string `json:"city_id"`
	// VisualStyle is the current visual appearance tier
	VisualStyle CityVisualStyle `json:"visual_style"`
	// BuildingCondition affects building sprite selection (0.0-1.0)
	BuildingCondition float64 `json:"building_condition"`
	// RoadCondition affects road/path sprite selection (0.0-1.0)
	RoadCondition float64 `json:"road_condition"`
	// DecorationDensity controls how many decorative elements appear (0.0-1.0)
	DecorationDensity float64 `json:"decoration_density"`
	// LightingLevel affects ambient lighting brightness (0.0-1.0)
	LightingLevel float64 `json:"lighting_level"`
	// PopulationActivity controls NPC spawn density multiplier (0.5-2.0)
	PopulationActivity float64 `json:"population_activity"`
	// MarketActivity affects merchant spawn and stall decorations (0.0-1.0)
	MarketActivity float64 `json:"market_activity"`
	// GuardPresence affects guard patrol visibility (0.0-1.0)
	GuardPresence float64 `json:"guard_presence"`
	// DebrisDensity controls rubble/trash sprites (0.0-1.0, inverse of prosperity)
	DebrisDensity float64 `json:"debris_density"`
	// BannerCount number of city banners/flags to display
	BannerCount int `json:"banner_count"`
	// PrimaryColor is the city's main color (HSV hue 0-360)
	PrimaryColor int `json:"primary_color"`
	// SecondaryColor is the city's accent color (HSV hue 0-360)
	SecondaryColor int `json:"secondary_color"`
}

// NewCityVisualComponent creates a visual component with default modest values.
func NewCityVisualComponent(cityID string) *CityVisualComponent {
	return &CityVisualComponent{
		CityID:             cityID,
		VisualStyle:        VisualStyleModest,
		BuildingCondition:  0.5,
		RoadCondition:      0.5,
		DecorationDensity:  0.3,
		LightingLevel:      0.6,
		PopulationActivity: 1.0,
		MarketActivity:     0.5,
		GuardPresence:      0.5,
		DebrisDensity:      0.2,
		BannerCount:        2,
		PrimaryColor:       200, // Blue
		SecondaryColor:     40,  // Gold
	}
}

// Type returns the component type identifier.
func (c *CityVisualComponent) Type() string {
	return "city_visual"
}

// UpdateFromCityState updates visual parameters based on city state.
func (c *CityVisualComponent) UpdateFromCityState(cityState *CityStateComponent) {
	if cityState == nil {
		return
	}

	prosperity := cityState.Prosperity
	infrastructure := cityState.Infrastructure
	defense := cityState.Defense
	populationRatio := getCityPopulationRatio(cityState)

	// Determine visual style from state
	switch cityState.State {
	case CityStateStruggling:
		c.VisualStyle = VisualStyleRundown
	case CityStateThriving:
		c.VisualStyle = VisualStyleProsperous
	default:
		c.VisualStyle = VisualStyleModest
	}

	// Building condition follows infrastructure
	c.BuildingCondition = clampFloat(infrastructure, 0.0, 1.0)

	// Road condition is slightly lower than infrastructure
	c.RoadCondition = clampFloat(infrastructure*0.9, 0.0, 1.0)

	// Decorations scale with prosperity
	c.DecorationDensity = clampFloat(prosperity*0.8, 0.0, 1.0)

	// Lighting improves with prosperity (struggling cities are dim)
	c.LightingLevel = clampFloat(0.4+prosperity*0.5, 0.0, 1.0)

	// Population activity based on population ratio and prosperity
	c.PopulationActivity = clampFloat(0.5+populationRatio+prosperity*0.5, 0.5, 2.0)

	// Market activity follows prosperity
	c.MarketActivity = clampFloat(prosperity, 0.0, 1.0)

	// Guard presence follows defense stat
	c.GuardPresence = clampFloat(defense, 0.0, 1.0)

	// Debris is inverse of prosperity
	c.DebrisDensity = clampFloat(1.0-prosperity, 0.0, 1.0)

	// Banner count based on prosperity tier
	switch cityState.State {
	case CityStateStruggling:
		c.BannerCount = 0
	case CityStateStable:
		c.BannerCount = 2
	case CityStateThriving:
		c.BannerCount = 5
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "city_visual",
		"city_id":        c.CityID,
		"visual_style":   c.VisualStyle,
		"prosperity":     prosperity,
	}).Debug("Updated city visuals from state")
}

// GetBuildingSpriteVariant returns the sprite variant to use for buildings.
// 0 = damaged, 1 = normal, 2 = fancy
func (c *CityVisualComponent) GetBuildingSpriteVariant() int {
	if c.BuildingCondition < 0.3 {
		return 0 // damaged
	} else if c.BuildingCondition >= 0.7 {
		return 2 // fancy
	}
	return 1 // normal
}

// ShouldSpawnDecoration returns true if a decoration should be placed.
// Uses random sampling to match decoration density.
func (c *CityVisualComponent) ShouldSpawnDecoration(rng *rand.Rand) bool {
	return rng.Float64() < c.DecorationDensity
}

// ShouldSpawnDebris returns true if debris should be placed.
func (c *CityVisualComponent) ShouldSpawnDebris(rng *rand.Rand) bool {
	return rng.Float64() < c.DebrisDensity
}

// GetNPCSpawnMultiplier returns the multiplier for NPC spawn rates.
func (c *CityVisualComponent) GetNPCSpawnMultiplier() float64 {
	return c.PopulationActivity
}

// Serialize encodes the component to bytes for persistence.
func (c *CityVisualComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "city_visual",
		"city_id":        c.CityID,
	}).Debug("Serializing city visual component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "city_visual",
			"error":          err.Error(),
		}).Error("Failed to serialize city visual component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *CityVisualComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "city_visual",
		"bytes":          len(data),
	}).Debug("Deserializing city visual component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "city_visual",
			"error":          err.Error(),
		}).Error("Failed to deserialize city visual component")
		return err
	}
	return nil
}

// GenerateCityVisualFromSeed creates a visual component with deterministic colors.
func GenerateCityVisualFromSeed(cityID string, seed int64) *CityVisualComponent {
	rng := rand.New(rand.NewSource(seed))

	visual := NewCityVisualComponent(cityID)
	visual.PrimaryColor = rng.Intn(360)
	visual.SecondaryColor = (visual.PrimaryColor + 30 + rng.Intn(120)) % 360

	logrus.WithFields(logrus.Fields{
		"component_type":  "city_visual",
		"city_id":         cityID,
		"seed":            seed,
		"primary_color":   visual.PrimaryColor,
		"secondary_color": visual.SecondaryColor,
	}).Debug("Generated city visual from seed")

	return visual
}
