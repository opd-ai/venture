package economy

import (
	"testing"
	"time"
)

// TestApplyTradeImpactNewItem verifies impact on previously unseen item types.
func TestApplyTradeImpactNewItem(t *testing.T) {
	pe := NewPricingEngine()

	// Apply trade impact for new item
	pe.ApplyTradeImpact("Timber", 0.9, 100)

	trend := pe.GetTrend("Timber")
	if trend.ItemType != "Timber" {
		t.Errorf("Expected ItemType 'Timber', got '%s'", trend.ItemType)
	}

	// New items start at base price 100, then apply multiplier
	// With 0.9 multiplier and 100 volume: (100*0 + 90*100) / 100 = 90
	expectedPrice := 90
	if trend.AveragePrice != expectedPrice {
		t.Errorf("Expected AveragePrice %d, got %d", expectedPrice, trend.AveragePrice)
	}

	if trend.TotalVolume != 100 {
		t.Errorf("Expected TotalVolume 100, got %d", trend.TotalVolume)
	}

	if trend.LastUpdated.IsZero() {
		t.Error("Expected LastUpdated to be set")
	}
}

// TestApplyTradeImpactExistingItem verifies impact on existing trends.
func TestApplyTradeImpactExistingItem(t *testing.T) {
	pe := NewPricingEngine()

	// Create initial trend via listing
	listing := &Listing{
		ItemType: "Ore",
		Price:    100,
		Quantity: 50,
	}
	pe.RecordListing(listing)

	initialTrend := pe.GetTrend("Ore")

	// Apply trade impact (10% price reduction from supply increase)
	pe.ApplyTradeImpact("Ore", 0.9, 200)

	trend := pe.GetTrend("Ore")

	// New average should be weighted: (100*50 + 90*200) / 250 = (5000 + 18000) / 250 = 92
	expectedPrice := 92
	if trend.AveragePrice != expectedPrice {
		t.Errorf("Expected AveragePrice %d, got %d", expectedPrice, trend.AveragePrice)
	}

	// Total volume should increase
	expectedVolume := 250 // 50 + 200
	if trend.TotalVolume != expectedVolume {
		t.Errorf("Expected TotalVolume %d, got %d", expectedVolume, trend.TotalVolume)
	}

	// Price should have decreased
	if trend.AveragePrice >= initialTrend.AveragePrice {
		t.Errorf("Expected price to decrease from %d, got %d", initialTrend.AveragePrice, trend.AveragePrice)
	}
}

// TestApplyTradeImpactPriceFloor verifies price cannot go below 1.
func TestApplyTradeImpactPriceFloor(t *testing.T) {
	pe := NewPricingEngine()

	// Set very low initial price
	listing := &Listing{
		ItemType: "Grain",
		Price:    2,
		Quantity: 10,
	}
	pe.RecordListing(listing)

	// Apply massive price reduction (90% decrease)
	pe.ApplyTradeImpact("Grain", 0.1, 100)

	trend := pe.GetTrend("Grain")

	// Price should never go below 1
	if trend.AveragePrice < 1 {
		t.Errorf("Price went below floor: %d", trend.AveragePrice)
	}
}

// TestApplyTradeImpactMinMaxUpdates verifies min/max price tracking.
func TestApplyTradeImpactMinMaxUpdates(t *testing.T) {
	pe := NewPricingEngine()

	// Create initial trend
	listing := &Listing{
		ItemType: "Spices",
		Price:    100,
		Quantity: 50,
	}
	pe.RecordListing(listing)

	// Apply impact that lowers price below current min
	pe.ApplyTradeImpact("Spices", 0.8, 100) // 20% reduction

	trend := pe.GetTrend("Spices")

	// New average: (100*50 + 80*100) / 150 = (5000 + 8000) / 150 = 86
	// Min should update to 80 (the new price calculated)
	if trend.MinPrice > trend.AveragePrice {
		t.Errorf("MinPrice %d should be <= AveragePrice %d", trend.MinPrice, trend.AveragePrice)
	}

	// Apply impact that raises price above current max
	pe.ApplyTradeImpact("Spices", 1.5, 100) // 50% increase

	trend = pe.GetTrend("Spices")

	if trend.MaxPrice < trend.AveragePrice {
		t.Errorf("MaxPrice %d should be >= AveragePrice %d", trend.MaxPrice, trend.AveragePrice)
	}
}

// TestApplyTradeImpactMultipleItems verifies independent tracking per item type.
func TestApplyTradeImpactMultipleItems(t *testing.T) {
	pe := NewPricingEngine()

	// Apply impacts to different items
	pe.ApplyTradeImpact("Timber", 0.9, 100)
	pe.ApplyTradeImpact("Ore", 1.1, 50)
	pe.ApplyTradeImpact("Grain", 0.95, 200)

	// Verify each item has independent trend
	timberTrend := pe.GetTrend("Timber")
	oreTrend := pe.GetTrend("Ore")
	grainTrend := pe.GetTrend("Grain")

	if timberTrend.ItemType != "Timber" {
		t.Error("Timber trend has wrong item type")
	}
	if oreTrend.ItemType != "Ore" {
		t.Error("Ore trend has wrong item type")
	}
	if grainTrend.ItemType != "Grain" {
		t.Error("Grain trend has wrong item type")
	}

	// Verify volumes are independent
	if timberTrend.TotalVolume != 100 {
		t.Errorf("Expected Timber volume 100, got %d", timberTrend.TotalVolume)
	}
	if oreTrend.TotalVolume != 50 {
		t.Errorf("Expected Ore volume 50, got %d", oreTrend.TotalVolume)
	}
	if grainTrend.TotalVolume != 200 {
		t.Errorf("Expected Grain volume 200, got %d", grainTrend.TotalVolume)
	}
}

// TestApplyTradeImpactCumulativeEffect verifies multiple impacts accumulate.
func TestApplyTradeImpactCumulativeEffect(t *testing.T) {
	pe := NewPricingEngine()

	// Apply first impact
	pe.ApplyTradeImpact("Textiles", 0.9, 100)
	firstTrend := pe.GetTrend("Textiles")
	firstVolume := firstTrend.TotalVolume

	// Apply second impact
	pe.ApplyTradeImpact("Textiles", 0.95, 50)
	secondTrend := pe.GetTrend("Textiles")

	// Volume should accumulate
	expectedVolume := firstVolume + 50
	if secondTrend.TotalVolume != expectedVolume {
		t.Errorf("Expected cumulative volume %d, got %d", expectedVolume, secondTrend.TotalVolume)
	}

	// Price should be weighted average of both impacts
	// First: (100*0 + 90*100) / 100 = 90
	// Second: newAvg = 90*0.95 = 85 (int), weighted = (90*100 + 85*50) / 150 = 88
	expectedPrice := 88
	if secondTrend.AveragePrice != expectedPrice {
		t.Errorf("Expected cumulative price %d, got %d", expectedPrice, secondTrend.AveragePrice)
	}
}

// TestApplyTradeImpactThreadSafety verifies concurrent impacts are safe.
func TestApplyTradeImpactThreadSafety(t *testing.T) {
	pe := NewPricingEngine()

	// Apply concurrent impacts
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			itemType := "Commodity"
			priceChange := 0.95 + float64(idx%10)*0.01 // 0.95-1.04
			volume := (idx % 20) + 10                  // 10-29
			pe.ApplyTradeImpact(itemType, priceChange, volume)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	timeout := time.After(5 * time.Second)
	for i := 0; i < 100; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent impacts")
		}
	}

	// Verify trend exists and has accumulated all impacts
	trend := pe.GetTrend("Commodity")
	if trend.TotalVolume == 0 {
		t.Error("Expected non-zero volume from concurrent impacts")
	}
	if trend.ItemType != "Commodity" {
		t.Error("Expected ItemType to be set correctly")
	}
}

// TestSystemApplyTradeImpact verifies System wrapper delegates to pricing engine.
func TestSystemApplyTradeImpact(t *testing.T) {
	// Create mock world
	world := &mockWorldForTesting{}
	system := NewSystem(world)

	// Apply trade impact via system
	system.ApplyTradeImpact("Weapons", 1.1, 75)

	// Verify pricing engine received the update
	trend := system.GetPriceTrend("Weapons")
	if trend.ItemType != "Weapons" {
		t.Errorf("Expected ItemType 'Weapons', got '%s'", trend.ItemType)
	}
	if trend.TotalVolume != 75 {
		t.Errorf("Expected TotalVolume 75, got %d", trend.TotalVolume)
	}
}

// TestSystemApplyTradeImpactThreadSafety verifies System-level concurrency safety.
func TestSystemApplyTradeImpactThreadSafety(t *testing.T) {
	world := &mockWorldForTesting{}
	system := NewSystem(world)

	// Apply concurrent impacts via system
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func(idx int) {
			system.ApplyTradeImpact("Armor", 0.98, 20)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	timeout := time.After(5 * time.Second)
	for i := 0; i < 50; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent system impacts")
		}
	}

	// Verify system state is consistent
	trend := system.GetPriceTrend("Armor")
	expectedVolume := 50 * 20 // 50 impacts * 20 volume each
	if trend.TotalVolume != expectedVolume {
		t.Errorf("Expected TotalVolume %d, got %d", expectedVolume, trend.TotalVolume)
	}
}

// mockWorldForTesting implements the World interface for testing.
type mockWorldForTesting struct{}

func (m *mockWorldForTesting) GetEntities() []Entity {
	return []Entity{}
}
