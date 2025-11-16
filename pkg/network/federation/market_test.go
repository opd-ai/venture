package federation

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestNewFederatedMarket(t *testing.T) {
	market := NewFederatedMarket()
	if market == nil {
		t.Fatal("NewFederatedMarket returned nil")
	}
	if market.itemPrices == nil {
		t.Error("itemPrices map not initialized")
	}
	if market.supply == nil {
		t.Error("supply map not initialized")
	}
	if market.demand == nil {
		t.Error("demand map not initialized")
	}
}

func TestRegisterItem(t *testing.T) {
	market := NewFederatedMarket()

	market.RegisterItem("item1", "server1", 100.0)

	history, err := market.GetPriceHistory("item1")
	if err != nil {
		t.Fatalf("GetPriceHistory failed: %v", err)
	}

	if history.ItemID != "item1" {
		t.Errorf("ItemID = %s, want item1", history.ItemID)
	}
	if history.ServerID != "server1" {
		t.Errorf("ServerID = %s, want server1", history.ServerID)
	}
	if history.BasePrice != 100.0 {
		t.Errorf("BasePrice = %f, want 100.0", history.BasePrice)
	}
	if history.CurrentPrice != 100.0 {
		t.Errorf("CurrentPrice = %f, want 100.0", history.CurrentPrice)
	}
}

func TestRegisterItemIdempotent(t *testing.T) {
	market := NewFederatedMarket()

	market.RegisterItem("item1", "server1", 100.0)
	market.RegisterItem("item1", "server2", 200.0) // Should not overwrite

	history, _ := market.GetPriceHistory("item1")
	if history.BasePrice != 100.0 {
		t.Errorf("BasePrice changed on re-register: got %f, want 100.0", history.BasePrice)
	}
}

func TestUpdateSupply(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	market.UpdateSupply("item1", 50)
	if got := market.GetSupply("item1"); got != 50 {
		t.Errorf("Supply = %d, want 50", got)
	}

	market.UpdateSupply("item1", 30)
	if got := market.GetSupply("item1"); got != 80 {
		t.Errorf("Supply = %d, want 80", got)
	}

	market.UpdateSupply("item1", -20)
	if got := market.GetSupply("item1"); got != 60 {
		t.Errorf("Supply = %d, want 60", got)
	}
}

func TestUpdateSupplyNegativeClamping(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	market.UpdateSupply("item1", -100)
	if got := market.GetSupply("item1"); got != 0 {
		t.Errorf("Supply = %d, want 0 (should clamp negative)", got)
	}
}

func TestUpdateDemand(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	market.UpdateDemand("item1", 20)
	if got := market.GetDemand("item1"); got != 20 {
		t.Errorf("Demand = %d, want 20", got)
	}

	market.UpdateDemand("item1", 15)
	if got := market.GetDemand("item1"); got != 35 {
		t.Errorf("Demand = %d, want 35", got)
	}
}

func TestUpdateDemandNegativeClamping(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	market.UpdateDemand("item1", -50)
	if got := market.GetDemand("item1"); got != 0 {
		t.Errorf("Demand = %d, want 0 (should clamp negative)", got)
	}
}

func TestCalculatePrice_NoSupplyNoDemand(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	price := market.CalculatePrice("item1", 1.0)
	if price != 100.0 {
		t.Errorf("Price = %f, want 100.0 (no supply/demand)", price)
	}
}

func TestCalculatePrice_NoSupplyHighDemand(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateDemand("item1", 50)

	price := market.CalculatePrice("item1", 1.0)
	if price != 300.0 {
		t.Errorf("Price = %f, want 300.0 (3x base for zero supply)", price)
	}
}

func TestCalculatePrice_BalancedSupplyDemand(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 50)
	market.UpdateDemand("item1", 50)

	price := market.CalculatePrice("item1", 1.0)
	if price != 100.0 {
		t.Errorf("Price = %f, want 100.0 (1:1 ratio)", price)
	}
}

func TestCalculatePrice_HighDemandLowSupply(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 10)
	market.UpdateDemand("item1", 50)

	price := market.CalculatePrice("item1", 1.0)
	// Ratio = 50/10 = 5.0, clamped at 5.0 max
	expected := 100.0 * 5.0
	if price != expected {
		t.Errorf("Price = %f, want %f (high demand)", price, expected)
	}
}

func TestCalculatePrice_LowDemandHighSupply(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 100)
	market.UpdateDemand("item1", 10)

	price := market.CalculatePrice("item1", 1.0)
	// Ratio = 10/100 = 0.1, clamped at 0.2 min
	expected := 100.0 * 0.2
	if price != expected {
		t.Errorf("Price = %f, want %f (low demand)", price, expected)
	}
}

func TestCalculatePrice_ServerMultiplier(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 50)
	market.UpdateDemand("item1", 50)

	tests := []struct {
		name       string
		multiplier float64
		want       float64
	}{
		{"ally", 0.8, 80.0},
		{"neutral", 1.0, 100.0},
		{"enemy", 1.5, 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := market.CalculatePrice("item1", tt.multiplier)
			if math.Abs(price-tt.want) > 0.01 {
				t.Errorf("Price = %f, want %f", price, tt.want)
			}
		})
	}
}

func TestGetPrice(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	price := market.GetPrice("item1", 1.0)
	if price != 100.0 {
		t.Errorf("GetPrice = %f, want 100.0", price)
	}

	// GetPrice returns current price, not calculated
	market.UpdateSupply("item1", 10)
	market.UpdateDemand("item1", 50)
	price = market.GetPrice("item1", 1.0)
	// Should still be 100.0 until UpdatePrices() is called
	if price != 100.0 {
		t.Errorf("GetPrice = %f, want 100.0 (before update)", price)
	}
}

func TestUpdatePrices(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 10)
	market.UpdateDemand("item1", 50)

	market.UpdatePrices()

	history, _ := market.GetPriceHistory("item1")
	expected := 100.0 * 5.0 // 5:1 demand/supply ratio
	if history.CurrentPrice != expected {
		t.Errorf("CurrentPrice = %f, want %f", history.CurrentPrice, expected)
	}

	if len(history.History) != 1 {
		t.Errorf("History length = %d, want 1", len(history.History))
	}
}

func TestUpdatePrices_HistoryTrimming(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	// Add 300 price points (more than 288 max)
	for i := 0; i < 300; i++ {
		market.UpdatePrices()
	}

	history, _ := market.GetPriceHistory("item1")
	if len(history.History) > 288 {
		t.Errorf("History length = %d, want ≤288 (trimmed)", len(history.History))
	}
}

func TestGetPriceHistory_NotFound(t *testing.T) {
	market := NewFederatedMarket()

	_, err := market.GetPriceHistory("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent item")
	}
}

func TestCalculateShippingCost(t *testing.T) {
	tests := []struct {
		name      string
		basePrice float64
		hops      int
		want      float64
	}{
		{"zero hops", 100.0, 0, 0.0},
		{"negative hops", 100.0, -1, 0.0},
		{"one hop", 100.0, 1, 10.0},
		{"three hops", 100.0, 3, 30.0},
		{"five hops", 50.0, 5, 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateShippingCost(tt.basePrice, tt.hops)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("CalculateShippingCost = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.RegisterItem("item2", "server1", 200.0)
	market.UpdateSupply("item1", 50)
	market.UpdateSupply("item2", 30)
	market.UpdateDemand("item1", 20)
	market.UpdateDemand("item2", 40)

	stats := market.GetStats()

	if stats.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", stats.TotalItems)
	}
	if stats.TotalSupply != 80 {
		t.Errorf("TotalSupply = %d, want 80", stats.TotalSupply)
	}
	if stats.TotalDemand != 60 {
		t.Errorf("TotalDemand = %d, want 60", stats.TotalDemand)
	}
}

func TestStartStop(t *testing.T) {
	market := NewFederatedMarket()

	market.Start()
	time.Sleep(10 * time.Millisecond)
	market.Stop()

	// Should not panic or hang
}

func TestMarketConcurrentAccess(t *testing.T) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			market.UpdateSupply("item1", 1)
			market.UpdateDemand("item1", 1)
			market.GetPrice("item1", 1.0)
			market.CalculatePrice("item1", 1.0)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not race
}

func BenchmarkCalculatePrice(b *testing.B) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 50)
	market.UpdateDemand("item1", 75)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		market.CalculatePrice("item1", 1.0)
	}
}

func BenchmarkUpdatePrices(b *testing.B) {
	market := NewFederatedMarket()
	for i := 0; i < 100; i++ {
		market.RegisterItem(fmt.Sprintf("item%d", i), "server1", float64(100+i))
		market.UpdateSupply(fmt.Sprintf("item%d", i), 50)
		market.UpdateDemand(fmt.Sprintf("item%d", i), 75)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		market.UpdatePrices()
	}
}

func BenchmarkConcurrentRead(b *testing.B) {
	market := NewFederatedMarket()
	market.RegisterItem("item1", "server1", 100.0)
	market.UpdateSupply("item1", 50)
	market.UpdateDemand("item1", 75)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			market.GetPrice("item1", 1.0)
		}
	})
}
