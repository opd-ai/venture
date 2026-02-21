// Package engine provides the EconomyTerritoryIntegrationSystem which bridges
// the economy pricing engine with territory control mechanics. This creates
// emergent gameplay where market conditions affect territorial warfare costs.
//
// The system reads market price data and adjusts territory-related costs and
// bonuses dynamically. During economic booms, maintenance costs rise but resource
// bonuses also increase. During downturns, territories become cheaper to maintain
// but yield less reward. High market volatility increases capture costs, making
// war more expensive during uncertain economic times.
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// EconomyProvider provides market economy data for territory calculations.
// This interface allows the system to integrate with any economy implementation.
type EconomyProvider interface {
	// GetAveragePrice returns the average market price for an item type.
	// Returns 0 if no price data exists.
	GetAveragePrice(itemType string) float64

	// GetPriceVolatility returns price volatility (0-1) for an item type.
	// Higher values indicate more price fluctuation.
	GetPriceVolatility(itemType string) float64

	// GetTotalMarketVolume returns the total trading volume across all items.
	GetTotalMarketVolume() int

	// GetDemandScore returns a demand score (0-1) for an item type.
	// Higher values indicate higher demand relative to supply.
	GetDemandScore(itemType string) float64
}

// TerritoryDataProvider provides territory information for economic calculations.
type TerritoryDataProvider interface {
	// GetTerritoryResourceType returns the primary resource type for a territory.
	// Empty string means no specific resource association.
	GetTerritoryResourceType(territoryID string) string

	// GetTerritoryTradeVolume returns the trade volume flowing through a territory.
	GetTerritoryTradeVolume(territoryID string) int

	// ListTerritoryIDs returns all territory IDs in the system.
	ListTerritoryIDs() []string
}

// EconomyTerritoryIntegrationSystem updates territory economic modifiers based
// on market conditions. It runs periodically to recalculate multipliers.
type EconomyTerritoryIntegrationSystem struct {
	world            *World
	economyProvider  EconomyProvider
	territoryData    TerritoryDataProvider
	updateInterval   float64 // Seconds between economy recalculations
	lastUpdate       float64
	logger           *logrus.Entry
	baseResourceRate float64 // Base multiplier for resource calculations
}

// EconomyTerritoryConfig holds configuration for the integration system.
type EconomyTerritoryConfig struct {
	// UpdateInterval is seconds between economic recalculations. Default: 30.0
	UpdateInterval float64

	// BaseResourceRate is the baseline for resource multiplier calculations.
	// Default: 1.0 (neutral)
	BaseResourceRate float64
}

// DefaultEconomyTerritoryConfig returns sensible defaults.
func DefaultEconomyTerritoryConfig() EconomyTerritoryConfig {
	return EconomyTerritoryConfig{
		UpdateInterval:   30.0,
		BaseResourceRate: 1.0,
	}
}

// NewEconomyTerritoryIntegrationSystem creates a new system instance.
func NewEconomyTerritoryIntegrationSystem(world *World, config EconomyTerritoryConfig) *EconomyTerritoryIntegrationSystem {
	var logEntry *logrus.Entry
	if world != nil && world.GetLogger() != nil {
		logEntry = world.GetLogger().WithFields(logrus.Fields{
			"system_name": "economy_territory_integration",
		})
	}

	if config.UpdateInterval <= 0 {
		config.UpdateInterval = 30.0
	}
	if config.BaseResourceRate <= 0 {
		config.BaseResourceRate = 1.0
	}

	return &EconomyTerritoryIntegrationSystem{
		world:            world,
		updateInterval:   config.UpdateInterval,
		baseResourceRate: config.BaseResourceRate,
		lastUpdate:       0,
		logger:           logEntry,
	}
}

// SetEconomyProvider configures the economy data source.
func (s *EconomyTerritoryIntegrationSystem) SetEconomyProvider(provider EconomyProvider) {
	s.economyProvider = provider
	if s.logger != nil {
		s.logger.Debug("economy provider configured")
	}
}

// SetTerritoryDataProvider configures the territory data source.
func (s *EconomyTerritoryIntegrationSystem) SetTerritoryDataProvider(provider TerritoryDataProvider) {
	s.territoryData = provider
	if s.logger != nil {
		s.logger.Debug("territory data provider configured")
	}
}

// Update processes entities with EconomicInfluenceComponent and recalculates
// their economic modifiers based on current market conditions.
func (s *EconomyTerritoryIntegrationSystem) Update(entities []*Entity, deltaTime float64) {
	s.lastUpdate += deltaTime

	// Only recalculate at the configured interval
	if s.lastUpdate < s.updateInterval {
		// Still update dirty components immediately
		for _, entity := range entities {
			if comp := s.getEconomicInfluenceComponent(entity); comp != nil && comp.Dirty {
				s.recalculateComponent(comp)
			}
		}
		return
	}

	s.lastUpdate = 0

	// Full recalculation pass
	for _, entity := range entities {
		comp := s.getEconomicInfluenceComponent(entity)
		if comp == nil {
			continue
		}

		s.recalculateComponent(comp)
	}
}

// getEconomicInfluenceComponent safely retrieves the component from an entity.
func (s *EconomyTerritoryIntegrationSystem) getEconomicInfluenceComponent(entity *Entity) *EconomicInfluenceComponent {
	compRaw, ok := entity.GetComponent("economic_influence")
	if !ok {
		return nil
	}
	comp, ok := compRaw.(*EconomicInfluenceComponent)
	if !ok {
		return nil
	}
	return comp
}

// recalculateComponent updates all economic modifiers for a territory.
func (s *EconomyTerritoryIntegrationSystem) recalculateComponent(comp *EconomicInfluenceComponent) {
	if s.economyProvider == nil {
		// No economy data — use defaults
		comp.MaintenanceCostMultiplier = 1.0
		comp.CaptureCostMultiplier = 1.0
		comp.ResourceBonusMultiplier = 1.0
		comp.TradeRouteBonusMultiplier = 1.0
		comp.Dirty = false
		return
	}

	// Get the primary resource for this territory
	resourceType := ""
	if s.territoryData != nil {
		resourceType = s.territoryData.GetTerritoryResourceType(comp.TerritoryID)
	}

	// Calculate demand pressure from market data
	demandPressure := s.calculateDemandPressure(resourceType)
	comp.DemandPressure = demandPressure

	// Calculate market influence based on trade volume
	marketInfluence := s.calculateMarketInfluence(comp.TerritoryID)
	comp.MarketInfluenceScore = marketInfluence

	// Calculate multipliers based on economic conditions
	comp.MaintenanceCostMultiplier = s.calculateMaintenanceMultiplier(demandPressure, marketInfluence)
	comp.CaptureCostMultiplier = s.calculateCaptureMultiplier(resourceType)
	comp.ResourceBonusMultiplier = s.calculateResourceMultiplier(demandPressure, resourceType)
	comp.TradeRouteBonusMultiplier = s.calculateTradeRouteMultiplier(marketInfluence)

	comp.Dirty = false

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"territory_id":     comp.TerritoryID,
			"maintenance_mult": comp.MaintenanceCostMultiplier,
			"capture_mult":     comp.CaptureCostMultiplier,
			"resource_mult":    comp.ResourceBonusMultiplier,
			"trade_route_mult": comp.TradeRouteBonusMultiplier,
			"demand_pressure":  comp.DemandPressure,
			"market_influence": comp.MarketInfluenceScore,
		}).Debug("recalculated territory economic modifiers")
	}
}

// calculateDemandPressure computes aggregate demand pressure for a resource type.
// Returns a value in range [0, 1].
func (s *EconomyTerritoryIntegrationSystem) calculateDemandPressure(resourceType string) float64 {
	if s.economyProvider == nil {
		return 0.5 // Neutral
	}

	if resourceType == "" {
		// No specific resource — use overall market demand
		totalVolume := s.economyProvider.GetTotalMarketVolume()
		// Normalize volume to [0,1] using a logarithmic scale
		// Assume 10000 volume is "high" market activity
		normalized := math.Log10(float64(totalVolume+1)) / 4.0
		return clampFloat64(normalized, 0.0, 1.0)
	}

	// Use the specific resource's demand score
	return clampFloat64(s.economyProvider.GetDemandScore(resourceType), 0.0, 1.0)
}

// calculateMarketInfluence computes the market influence score for a territory.
// Returns a value in range [0, 100].
func (s *EconomyTerritoryIntegrationSystem) calculateMarketInfluence(territoryID string) float64 {
	if s.territoryData == nil {
		return 50.0 // Neutral
	}

	tradeVolume := s.territoryData.GetTerritoryTradeVolume(territoryID)
	// Normalize trade volume to [0, 100] using logarithmic scale
	// Assume 1000 volume is "high" for a single territory
	normalized := math.Log10(float64(tradeVolume+1)) / 3.0 * 100.0
	return clampFloat64(normalized, 0.0, 100.0)
}

// calculateMaintenanceMultiplier computes the maintenance cost multiplier.
// High demand and market influence increase maintenance costs.
// Range: [0.5, 2.0]
func (s *EconomyTerritoryIntegrationSystem) calculateMaintenanceMultiplier(demandPressure, marketInfluence float64) float64 {
	// Base: 1.0
	// Demand adds up to +0.5 (high demand = expensive upkeep)
	// Market influence adds up to +0.5 (high traffic = more infrastructure)
	base := s.baseResourceRate
	demandFactor := demandPressure * 0.5
	marketFactor := (marketInfluence / 100.0) * 0.5

	multiplier := base + demandFactor + marketFactor
	return clampFloat64(multiplier, 0.5, 2.0)
}

// calculateCaptureMultiplier computes the capture cost multiplier.
// High price volatility increases capture costs (unstable markets = risky war).
// Range: [0.75, 1.5]
func (s *EconomyTerritoryIntegrationSystem) calculateCaptureMultiplier(resourceType string) float64 {
	if s.economyProvider == nil {
		return 1.0
	}

	volatility := 0.3 // Default moderate volatility
	if resourceType != "" {
		volatility = s.economyProvider.GetPriceVolatility(resourceType)
	}

	// Base: 1.0
	// High volatility (>0.5) increases capture cost
	// Low volatility (<0.3) slightly reduces capture cost
	base := s.baseResourceRate
	volatilityFactor := (volatility - 0.3) * 1.0 // -0.3 to +0.7 range

	multiplier := base + volatilityFactor
	return clampFloat64(multiplier, 0.75, 1.5)
}

// calculateResourceMultiplier computes the resource bonus multiplier.
// High demand for the territory's resource increases the bonus.
// Range: [0.8, 1.5]
func (s *EconomyTerritoryIntegrationSystem) calculateResourceMultiplier(demandPressure float64, resourceType string) float64 {
	base := s.baseResourceRate

	// Demand directly increases resource value
	demandFactor := (demandPressure - 0.5) * 0.6 // -0.3 to +0.3 range

	// Additional bonus if specific resource has high average price
	priceFactor := 0.0
	if s.economyProvider != nil && resourceType != "" {
		avgPrice := s.economyProvider.GetAveragePrice(resourceType)
		// Assume 100 gold is "baseline" price
		if avgPrice > 0 {
			priceFactor = math.Log10(avgPrice/100.0+1) * 0.2
		}
	}

	multiplier := base + demandFactor + priceFactor
	return clampFloat64(multiplier, 0.8, 1.5)
}

// calculateTradeRouteMultiplier computes the trade route profit multiplier.
// High market influence means more profitable trade routes.
// Range: [0.7, 2.0]
func (s *EconomyTerritoryIntegrationSystem) calculateTradeRouteMultiplier(marketInfluence float64) float64 {
	base := s.baseResourceRate

	// Market influence strongly affects trade route profits
	influenceFactor := (marketInfluence / 100.0) * 1.0 // 0 to 1 range

	multiplier := base + influenceFactor
	return clampFloat64(multiplier, 0.7, 2.0)
}

// clampFloat64 restricts a value to the given range.
func clampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// CreateTerritoryEconomyEntity creates an entity with EconomicInfluenceComponent
// for a specific territory. Call this when a territory is created to enable
// economic integration.
func CreateTerritoryEconomyEntity(world *World, territoryID string) *Entity {
	entity := world.CreateEntity()
	entity.AddComponent(NewEconomicInfluenceComponent(territoryID))
	return entity
}

// GetEconomicModifiersForTerritory finds the economic modifiers for a territory.
// Searches all entities for the matching EconomicInfluenceComponent.
// Returns nil if not found.
func GetEconomicModifiersForTerritory(entities []*Entity, territoryID string) *EconomicInfluenceComponent {
	for _, entity := range entities {
		compRaw, ok := entity.GetComponent("economic_influence")
		if !ok {
			continue
		}
		comp, ok := compRaw.(*EconomicInfluenceComponent)
		if !ok {
			continue
		}
		if comp.TerritoryID == territoryID {
			return comp
		}
	}
	return nil
}
