// Package engine provides the EconomicInfluenceComponent for economy-territory
// integration. This pure-data component tracks economic modifiers that affect
// territory control costs and resource bonuses based on market conditions.
package engine

// EconomicInfluenceComponent stores economy-derived modifiers for territory entities.
// Pure data structure — all logic resides in EconomyTerritoryIntegrationSystem.
type EconomicInfluenceComponent struct {
	// TerritoryID links this component to a specific territory.
	TerritoryID string

	// MaintenanceCostMultiplier scales ongoing territory maintenance costs.
	// Values > 1.0 mean higher costs during economic inflation.
	// Values < 1.0 mean lower costs during economic deflation.
	// Range: 0.5 - 2.0 (clamped)
	MaintenanceCostMultiplier float64

	// CaptureCostMultiplier scales the gold/resource cost to capture a territory.
	// High market volatility increases capture costs.
	// Range: 0.75 - 1.5 (clamped)
	CaptureCostMultiplier float64

	// ResourceBonusMultiplier scales the resource gathering bonus from owning territory.
	// High demand for resources increases this bonus.
	// Range: 0.8 - 1.5 (clamped)
	ResourceBonusMultiplier float64

	// TradeRouteBonusMultiplier scales trade route profits through this territory.
	// Depends on nearby market prices and volume.
	// Range: 0.7 - 2.0 (clamped)
	TradeRouteBonusMultiplier float64

	// MarketInfluenceScore is a 0-100 score indicating how much the local market
	// affects this territory. High-traffic trade areas score higher.
	MarketInfluenceScore float64

	// DemandPressure tracks aggregate demand across resource types affecting
	// this territory. Higher values increase resource bonus but also maintenance.
	// Range: 0.0 - 1.0 (normalized)
	DemandPressure float64

	// LastEconomyUpdate is the in-game time of the last economic recalculation.
	LastEconomyUpdate float64

	// Dirty flags the component for recalculation on next system update.
	Dirty bool
}

// Type returns the component type identifier for ECS registration.
func (c *EconomicInfluenceComponent) Type() string {
	return "economic_influence"
}

// NewEconomicInfluenceComponent creates a component with neutral default values.
func NewEconomicInfluenceComponent(territoryID string) *EconomicInfluenceComponent {
	return &EconomicInfluenceComponent{
		TerritoryID:               territoryID,
		MaintenanceCostMultiplier: 1.0,
		CaptureCostMultiplier:     1.0,
		ResourceBonusMultiplier:   1.0,
		TradeRouteBonusMultiplier: 1.0,
		MarketInfluenceScore:      50.0, // Neutral starting score
		DemandPressure:            0.5,  // Moderate demand
		LastEconomyUpdate:         0.0,
		Dirty:                     true, // Calculate on first update
	}
}

// EffectiveResourceBonus returns the territory's adjusted resource bonus.
// baseBonus is the territory's base resource bonus (e.g., 0.10 for 10%).
func (c *EconomicInfluenceComponent) EffectiveResourceBonus(baseBonus float64) float64 {
	return baseBonus * c.ResourceBonusMultiplier
}

// EffectiveMaintenanceCost returns the adjusted maintenance cost.
// baseCost is the territory's base maintenance cost.
func (c *EconomicInfluenceComponent) EffectiveMaintenanceCost(baseCost float64) float64 {
	return baseCost * c.MaintenanceCostMultiplier
}

// EffectiveCaptureCost returns the adjusted capture cost.
// baseCost is the base cost to capture the territory.
func (c *EconomicInfluenceComponent) EffectiveCaptureCost(baseCost float64) float64 {
	return baseCost * c.CaptureCostMultiplier
}

// EffectiveTradeRouteProfit returns the adjusted trade route profit.
// baseProfit is the base profit from trade routes through this territory.
func (c *EconomicInfluenceComponent) EffectiveTradeRouteProfit(baseProfit float64) float64 {
	return baseProfit * c.TradeRouteBonusMultiplier
}
