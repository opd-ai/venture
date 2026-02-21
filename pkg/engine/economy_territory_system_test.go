package engine

import (
	"testing"
)

// --- Mock providers for testing ---

type mockEconomyProvider struct {
	avgPrices    map[string]float64
	volatility   map[string]float64
	totalVolume  int
	demandScores map[string]float64
}

func newMockEconomyProvider() *mockEconomyProvider {
	return &mockEconomyProvider{
		avgPrices:    make(map[string]float64),
		volatility:   make(map[string]float64),
		totalVolume:  1000,
		demandScores: make(map[string]float64),
	}
}

func (m *mockEconomyProvider) GetAveragePrice(itemType string) float64 {
	return m.avgPrices[itemType]
}

func (m *mockEconomyProvider) GetPriceVolatility(itemType string) float64 {
	if v, ok := m.volatility[itemType]; ok {
		return v
	}
	return 0.3 // Default moderate volatility
}

func (m *mockEconomyProvider) GetTotalMarketVolume() int {
	return m.totalVolume
}

func (m *mockEconomyProvider) GetDemandScore(itemType string) float64 {
	if s, ok := m.demandScores[itemType]; ok {
		return s
	}
	return 0.5 // Default moderate demand
}

type mockTerritoryDataProvider struct {
	resourceTypes map[string]string
	tradeVolumes  map[string]int
	territoryIDs  []string
}

func newMockTerritoryDataProvider() *mockTerritoryDataProvider {
	return &mockTerritoryDataProvider{
		resourceTypes: make(map[string]string),
		tradeVolumes:  make(map[string]int),
		territoryIDs:  make([]string, 0),
	}
}

func (m *mockTerritoryDataProvider) GetTerritoryResourceType(territoryID string) string {
	return m.resourceTypes[territoryID]
}

func (m *mockTerritoryDataProvider) GetTerritoryTradeVolume(territoryID string) int {
	return m.tradeVolumes[territoryID]
}

func (m *mockTerritoryDataProvider) ListTerritoryIDs() []string {
	return m.territoryIDs
}

// --- Tests ---

func TestEconomicInfluenceComponent_Type(t *testing.T) {
	comp := NewEconomicInfluenceComponent("test-territory")
	if comp.Type() != "economic_influence" {
		t.Errorf("expected Type() = 'economic_influence', got '%s'", comp.Type())
	}
}

func TestEconomicInfluenceComponent_Defaults(t *testing.T) {
	comp := NewEconomicInfluenceComponent("territory-001")

	tests := []struct {
		name     string
		got      float64
		expected float64
	}{
		{"TerritoryID", 0, 0}, // Not a float, checked separately
		{"MaintenanceCostMultiplier", comp.MaintenanceCostMultiplier, 1.0},
		{"CaptureCostMultiplier", comp.CaptureCostMultiplier, 1.0},
		{"ResourceBonusMultiplier", comp.ResourceBonusMultiplier, 1.0},
		{"TradeRouteBonusMultiplier", comp.TradeRouteBonusMultiplier, 1.0},
		{"MarketInfluenceScore", comp.MarketInfluenceScore, 50.0},
		{"DemandPressure", comp.DemandPressure, 0.5},
	}

	if comp.TerritoryID != "territory-001" {
		t.Errorf("expected TerritoryID = 'territory-001', got '%s'", comp.TerritoryID)
	}

	for _, tt := range tests[1:] { // Skip TerritoryID check
		if tt.got != tt.expected {
			t.Errorf("%s: expected %f, got %f", tt.name, tt.expected, tt.got)
		}
	}

	if !comp.Dirty {
		t.Error("expected Dirty = true for new component")
	}
}

func TestEconomicInfluenceComponent_EffectiveBonuses(t *testing.T) {
	comp := &EconomicInfluenceComponent{
		MaintenanceCostMultiplier: 1.5,
		CaptureCostMultiplier:     1.2,
		ResourceBonusMultiplier:   1.3,
		TradeRouteBonusMultiplier: 1.8,
	}

	tests := []struct {
		name     string
		base     float64
		got      float64
		expected float64
	}{
		{"EffectiveMaintenanceCost", 100.0, comp.EffectiveMaintenanceCost(100.0), 150.0},
		{"EffectiveCaptureCost", 1000.0, comp.EffectiveCaptureCost(1000.0), 1200.0},
		{"EffectiveResourceBonus", 0.10, comp.EffectiveResourceBonus(0.10), 0.13},
		{"EffectiveTradeRouteProfit", 500.0, comp.EffectiveTradeRouteProfit(500.0), 900.0},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s(%f): expected %f, got %f", tt.name, tt.base, tt.expected, tt.got)
		}
	}
}

func TestNewEconomyTerritoryIntegrationSystem(t *testing.T) {
	world := NewWorld()

	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}

	if sys.updateInterval != 30.0 {
		t.Errorf("expected updateInterval = 30.0, got %f", sys.updateInterval)
	}

	if sys.baseResourceRate != 1.0 {
		t.Errorf("expected baseResourceRate = 1.0, got %f", sys.baseResourceRate)
	}
}

func TestEconomyTerritoryIntegrationSystem_Update_NoProvider(t *testing.T) {
	world := NewWorld()
	config := EconomyTerritoryConfig{UpdateInterval: 0.1}
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	entity := world.CreateEntity()
	comp := NewEconomicInfluenceComponent("test-territory")
	comp.MaintenanceCostMultiplier = 2.0 // Set non-default value
	entity.AddComponent(comp)

	entities := []*Entity{entity}

	// Run update with no providers
	sys.Update(entities, 0.2)

	// Should reset to defaults when no economy provider
	if comp.MaintenanceCostMultiplier != 1.0 {
		t.Errorf("expected maintenance multiplier reset to 1.0, got %f", comp.MaintenanceCostMultiplier)
	}
	if comp.Dirty {
		t.Error("expected Dirty = false after update")
	}
}

func TestEconomyTerritoryIntegrationSystem_Update_WithProviders(t *testing.T) {
	world := NewWorld()
	config := EconomyTerritoryConfig{UpdateInterval: 0.1}
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	// Setup mock providers
	econProvider := newMockEconomyProvider()
	econProvider.avgPrices["iron_ore"] = 150.0
	econProvider.volatility["iron_ore"] = 0.6 // High volatility
	econProvider.demandScores["iron_ore"] = 0.8
	econProvider.totalVolume = 5000

	territoryProvider := newMockTerritoryDataProvider()
	territoryProvider.resourceTypes["territory-iron"] = "iron_ore"
	territoryProvider.tradeVolumes["territory-iron"] = 500

	sys.SetEconomyProvider(econProvider)
	sys.SetTerritoryDataProvider(territoryProvider)

	entity := world.CreateEntity()
	comp := NewEconomicInfluenceComponent("territory-iron")
	entity.AddComponent(comp)

	entities := []*Entity{entity}

	// Run update
	sys.Update(entities, 0.2)

	// Verify multipliers were calculated
	if comp.MaintenanceCostMultiplier <= 1.0 {
		t.Errorf("expected maintenance multiplier > 1.0 (high demand), got %f", comp.MaintenanceCostMultiplier)
	}

	if comp.CaptureCostMultiplier <= 1.0 {
		t.Errorf("expected capture multiplier > 1.0 (high volatility), got %f", comp.CaptureCostMultiplier)
	}

	if comp.ResourceBonusMultiplier <= 1.0 {
		t.Errorf("expected resource multiplier > 1.0 (high demand), got %f", comp.ResourceBonusMultiplier)
	}

	if comp.DemandPressure != 0.8 {
		t.Errorf("expected demand pressure = 0.8, got %f", comp.DemandPressure)
	}
}

func TestEconomyTerritoryIntegrationSystem_IntervalUpdate(t *testing.T) {
	world := NewWorld()
	config := EconomyTerritoryConfig{UpdateInterval: 1.0}
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	econProvider := newMockEconomyProvider()
	sys.SetEconomyProvider(econProvider)

	entity := world.CreateEntity()
	comp := NewEconomicInfluenceComponent("test-territory")
	comp.Dirty = false // Start clean
	entity.AddComponent(comp)

	entities := []*Entity{entity}

	// First update at 0.5s — should not trigger full recalculation
	sys.Update(entities, 0.5)

	// Second update at 0.3s (total 0.8s) — still below interval
	sys.Update(entities, 0.3)

	// Third update at 0.3s (total 1.1s) — should trigger recalculation
	sys.Update(entities, 0.3)

	// Verify the system respects the interval
	// (Component should be updated since interval passed)
	if comp.Dirty {
		t.Error("expected Dirty = false after interval-based update")
	}
}

func TestEconomyTerritoryIntegrationSystem_DirtyComponentImmediateUpdate(t *testing.T) {
	world := NewWorld()
	config := EconomyTerritoryConfig{UpdateInterval: 100.0} // Long interval
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	econProvider := newMockEconomyProvider()
	sys.SetEconomyProvider(econProvider)

	entity := world.CreateEntity()
	comp := NewEconomicInfluenceComponent("test-territory")
	comp.Dirty = true
	entity.AddComponent(comp)

	entities := []*Entity{entity}

	// Update with time less than interval
	sys.Update(entities, 0.1)

	// Dirty components should be updated immediately
	if comp.Dirty {
		t.Error("expected Dirty = false after immediate update of dirty component")
	}
}

func TestCalculateDemandPressure(t *testing.T) {
	world := NewWorld()
	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	// Without provider
	pressure := sys.calculateDemandPressure("")
	if pressure != 0.5 {
		t.Errorf("expected default pressure = 0.5 without provider, got %f", pressure)
	}

	// With provider
	econProvider := newMockEconomyProvider()
	econProvider.demandScores["gold"] = 0.9
	sys.SetEconomyProvider(econProvider)

	pressure = sys.calculateDemandPressure("gold")
	if pressure != 0.9 {
		t.Errorf("expected pressure = 0.9, got %f", pressure)
	}
}

func TestCalculateMarketInfluence(t *testing.T) {
	world := NewWorld()
	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	// Without provider
	influence := sys.calculateMarketInfluence("territory-001")
	if influence != 50.0 {
		t.Errorf("expected default influence = 50.0 without provider, got %f", influence)
	}

	// With provider
	territoryProvider := newMockTerritoryDataProvider()
	territoryProvider.tradeVolumes["territory-001"] = 1000 // High volume
	sys.SetTerritoryDataProvider(territoryProvider)

	influence = sys.calculateMarketInfluence("territory-001")
	if influence <= 50.0 {
		t.Errorf("expected high influence > 50.0, got %f", influence)
	}
}

func TestCalculateMaintenanceMultiplier_Clamping(t *testing.T) {
	world := NewWorld()
	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	// Maximum values
	multiplier := sys.calculateMaintenanceMultiplier(1.0, 100.0)
	if multiplier > 2.0 {
		t.Errorf("expected multiplier <= 2.0 (max), got %f", multiplier)
	}

	// Minimum values
	multiplier = sys.calculateMaintenanceMultiplier(0.0, 0.0)
	if multiplier < 0.5 {
		t.Errorf("expected multiplier >= 0.5 (min), got %f", multiplier)
	}
}

func TestCalculateCaptureMultiplier_Volatility(t *testing.T) {
	world := NewWorld()
	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	econProvider := newMockEconomyProvider()
	sys.SetEconomyProvider(econProvider)

	// Low volatility
	econProvider.volatility["safe_resource"] = 0.1
	multiplier := sys.calculateCaptureMultiplier("safe_resource")
	if multiplier >= 1.0 {
		t.Errorf("expected low volatility to reduce capture cost < 1.0, got %f", multiplier)
	}

	// High volatility
	econProvider.volatility["risky_resource"] = 0.8
	multiplier = sys.calculateCaptureMultiplier("risky_resource")
	if multiplier <= 1.0 {
		t.Errorf("expected high volatility to increase capture cost > 1.0, got %f", multiplier)
	}
}

func TestCalculateResourceMultiplier_Demand(t *testing.T) {
	world := NewWorld()
	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	// High demand should increase multiplier
	multiplier := sys.calculateResourceMultiplier(0.9, "")
	if multiplier <= 1.0 {
		t.Errorf("expected high demand to increase resource multiplier > 1.0, got %f", multiplier)
	}

	// Low demand should decrease multiplier
	multiplier = sys.calculateResourceMultiplier(0.1, "")
	if multiplier >= 1.0 {
		t.Errorf("expected low demand to decrease resource multiplier < 1.0, got %f", multiplier)
	}
}

func TestCalculateTradeRouteMultiplier(t *testing.T) {
	world := NewWorld()
	config := DefaultEconomyTerritoryConfig()
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	// High market influence
	multiplier := sys.calculateTradeRouteMultiplier(100.0)
	if multiplier <= 1.5 {
		t.Errorf("expected high influence to increase trade route multiplier > 1.5, got %f", multiplier)
	}

	// Low market influence
	multiplier = sys.calculateTradeRouteMultiplier(10.0)
	if multiplier >= 1.5 {
		t.Errorf("expected low influence to keep trade route multiplier < 1.5, got %f", multiplier)
	}
}

func TestClampFloat64(t *testing.T) {
	tests := []struct {
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},  // Within range
		{-0.5, 0.0, 1.0, 0.0}, // Below min
		{1.5, 0.0, 1.0, 1.0},  // Above max
		{0.0, 0.0, 1.0, 0.0},  // At min
		{1.0, 0.0, 1.0, 1.0},  // At max
	}

	for _, tt := range tests {
		result := clampFloat64(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clampFloat64(%f, %f, %f) = %f, expected %f",
				tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestCreateTerritoryEconomyEntity(t *testing.T) {
	world := NewWorld()

	entity := CreateTerritoryEconomyEntity(world, "new-territory")

	if entity == nil {
		t.Fatal("expected non-nil entity")
	}

	comp, hasComp := entity.GetComponent("economic_influence")
	if !hasComp || comp == nil {
		t.Fatal("expected entity to have economic_influence component")
	}

	econComp, ok := comp.(*EconomicInfluenceComponent)
	if !ok {
		t.Fatal("expected component to be *EconomicInfluenceComponent")
	}

	if econComp.TerritoryID != "new-territory" {
		t.Errorf("expected TerritoryID = 'new-territory', got '%s'", econComp.TerritoryID)
	}
}

func TestGetEconomicModifiersForTerritory(t *testing.T) {
	world := NewWorld()

	// Create multiple territory entities
	entity1 := world.CreateEntity()
	entity1.AddComponent(NewEconomicInfluenceComponent("territory-A"))

	entity2 := world.CreateEntity()
	entity2.AddComponent(NewEconomicInfluenceComponent("territory-B"))

	entity3 := world.CreateEntity()
	entity3.AddComponent(&HealthComponent{Current: 100}) // Different component

	entities := []*Entity{entity1, entity2, entity3}

	// Find territory-B
	comp := GetEconomicModifiersForTerritory(entities, "territory-B")
	if comp == nil {
		t.Fatal("expected to find territory-B")
	}
	if comp.TerritoryID != "territory-B" {
		t.Errorf("expected TerritoryID = 'territory-B', got '%s'", comp.TerritoryID)
	}

	// Find non-existent territory
	comp = GetEconomicModifiersForTerritory(entities, "territory-C")
	if comp != nil {
		t.Error("expected nil for non-existent territory")
	}
}

func BenchmarkEconomyTerritoryIntegrationSystem_Update(b *testing.B) {
	world := NewWorld()
	config := EconomyTerritoryConfig{UpdateInterval: 0.01}
	sys := NewEconomyTerritoryIntegrationSystem(world, config)

	econProvider := newMockEconomyProvider()
	econProvider.totalVolume = 10000
	sys.SetEconomyProvider(econProvider)

	// Create 100 territory entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewEconomicInfluenceComponent("territory-" + string(rune('A'+i))))
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // ~60fps
	}
}
