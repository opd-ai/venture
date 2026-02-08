package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/world/economy"
)

// TestEconomySystemApplyTradeImpact verifies engine.EconomySystem implements PriceUpdateHandler interface.
func TestEconomySystemApplyTradeImpact(t *testing.T) {
	world := NewWorld()
	economySystem := NewEconomySystem(world, "test-server")

	// Apply trade impact
	economySystem.ApplyTradeImpact("Timber", 0.9, 100)

	// Verify price trend was updated
	marketplace := economySystem.GetMarketplace()
	trend := marketplace.GetPriceTrend("Timber")

	if trend.ItemType != "Timber" {
		t.Errorf("Expected ItemType 'Timber', got '%s'", trend.ItemType)
	}

	if trend.TotalVolume != 100 {
		t.Errorf("Expected TotalVolume 100, got %d", trend.TotalVolume)
	}

	// Price should be around 90 (100 * 0.9)
	if trend.AveragePrice < 85 || trend.AveragePrice > 95 {
		t.Errorf("Expected AveragePrice around 90, got %d", trend.AveragePrice)
	}
}

// TestEconomySystemTradeImpactMultipleItems verifies independent item tracking.
func TestEconomySystemTradeImpactMultipleItems(t *testing.T) {
	world := NewWorld()
	economySystem := NewEconomySystem(world, "test-server")

	// Apply impacts to different items
	economySystem.ApplyTradeImpact("Timber", 0.9, 100)
	economySystem.ApplyTradeImpact("Ore", 1.1, 50)
	economySystem.ApplyTradeImpact("Grain", 0.95, 200)

	marketplace := economySystem.GetMarketplace()

	// Verify each item has independent trend
	timberTrend := marketplace.GetPriceTrend("Timber")
	oreTrend := marketplace.GetPriceTrend("Ore")
	grainTrend := marketplace.GetPriceTrend("Grain")

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

// TestEconomySystemTradeImpactWithListings verifies interaction with existing market listings.
func TestEconomySystemTradeImpactWithListings(t *testing.T) {
	world := NewWorld()
	economySystem := NewEconomySystem(world, "test-server")

	// Create initial listing
	listing := &economy.Listing{
		ItemType: "Spices",
		ItemName: "Exotic Spices",
		Price:    100,
		Quantity: 50,
		SellerID: "merchant-1",
	}

	err := economySystem.CreateListing(listing)
	if err != nil {
		t.Fatalf("Failed to create listing: %v", err)
	}

	// Get initial price trend
	marketplace := economySystem.GetMarketplace()
	initialTrend := marketplace.GetPriceTrend("Spices")
	initialAvg := initialTrend.AveragePrice

	// Apply trade impact (10% price reduction from increased supply)
	economySystem.ApplyTradeImpact("Spices", 0.9, 200)

	// Verify price decreased
	updatedTrend := marketplace.GetPriceTrend("Spices")
	if updatedTrend.AveragePrice >= initialAvg {
		t.Errorf("Expected price to decrease from %d, got %d", initialAvg, updatedTrend.AveragePrice)
	}

	// Verify volume increased
	expectedVolume := 250 // 50 from listing + 200 from trade
	if updatedTrend.TotalVolume != expectedVolume {
		t.Errorf("Expected TotalVolume %d, got %d", expectedVolume, updatedTrend.TotalVolume)
	}
}
