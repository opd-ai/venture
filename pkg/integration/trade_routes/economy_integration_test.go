package trade_routes

import (
	"sync"
	"testing"
	"time"
)

// mockPriceUpdateHandler implements PriceUpdateHandler for testing.
type mockPriceUpdateHandler struct {
	mu      sync.Mutex
	updates []priceUpdate
}

type priceUpdate struct {
	itemType    string
	priceChange float64
	volume      int
}

func (m *mockPriceUpdateHandler) ApplyTradeImpact(itemType string, priceChange float64, volume int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updates = append(m.updates, priceUpdate{
		itemType:    itemType,
		priceChange: priceChange,
		volume:      volume,
	})
}

// getUpdates returns a snapshot of all updates (thread-safe).
func (m *mockPriceUpdateHandler) getUpdates() []priceUpdate {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]priceUpdate, len(m.updates))
	copy(result, m.updates)
	return result
}

// TestSetPriceUpdateHandler verifies handler injection.
func TestSetPriceUpdateHandler(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	handler := &mockPriceUpdateHandler{}

	rm.SetPriceUpdateHandler(handler)

	if rm.priceHandler == nil {
		t.Error("Expected price handler to be set")
	}
}

// TestCompleteRouteWithoutHandler verifies graceful handling when no handler is set.
func TestCompleteRouteWithoutHandler(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create and complete a route without setting handler
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	// Start the route
	err = rm.StartRoute(route.ID, 123)
	if err != nil {
		t.Fatalf("Failed to start route: %v", err)
	}

	// Mark route as completed
	route.Status = StatusCompleted
	route.Progress = 1.0

	// This should not panic even though handler is nil
	rm.completeRoute(route)

	// Verify route was cleaned up
	_, exists := rm.activeCaravans[123]
	if exists {
		t.Error("Expected caravan to be removed from active list")
	}
}

// TestCompleteRouteWithHandler verifies price updates are sent to handler.
func TestCompleteRouteWithHandler(t *testing.T) {
	rm := NewRouteManager("test-server", 54321)
	handler := &mockPriceUpdateHandler{}
	rm.SetPriceUpdateHandler(handler)

	// Create route with cargo
	route, err := rm.CreateRoute("region-a", "region-b", 5000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	// Start the route
	err = rm.StartRoute(route.ID, 456)
	if err != nil {
		t.Fatalf("Failed to start route: %v", err)
	}

	// Mark route as completed
	route.Status = StatusCompleted
	route.Progress = 1.0

	// Count cargo items with quantity > 0
	cargoCount := 0
	for _, item := range route.Cargo {
		if item.Quantity > 0 {
			cargoCount++
		}
	}

	// Complete the route
	rm.completeRoute(route)

	// Verify handler received updates for each cargo item
	if len(handler.updates) != cargoCount {
		t.Errorf("Expected %d price updates, got %d", cargoCount, len(handler.updates))
	}

	// Verify each update has valid parameters
	for i, update := range handler.updates {
		if update.itemType == "" {
			t.Errorf("Update %d: itemType is empty", i)
		}
		if update.priceChange <= 0 || update.priceChange > 1.0 {
			t.Errorf("Update %d: priceChange out of range (0-1.0): %.4f", i, update.priceChange)
		}
		if update.priceChange < 0.95 {
			t.Errorf("Update %d: priceChange below min (0.95): %.4f", i, update.priceChange)
		}
		if update.volume <= 0 {
			t.Errorf("Update %d: volume is not positive: %d", i, update.volume)
		}
	}
}

// TestCompleteRoutePriceImpactCalculation verifies price impact formula.
func TestCompleteRoutePriceImpactCalculation(t *testing.T) {
	tests := []struct {
		name              string
		quantity          int
		expectedMinImpact float64
		expectedMaxImpact float64
	}{
		{"Small quantity", 10, 0.9995, 1.0},  // 10/1000 * 0.05 = 0.0005 reduction
		{"Medium quantity", 100, 0.995, 1.0}, // 100/1000 * 0.05 = 0.005 reduction
		{"Large quantity", 500, 0.975, 1.0},  // 500/1000 * 0.05 = 0.025 reduction
		{"Max quantity", 1000, 0.95, 0.951},  // 1000/1000 * 0.05 = 0.05 reduction (capped)
		{"Over max", 2000, 0.95, 0.951},      // Should cap at 0.95
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewRouteManager("test-server", 99999)
			handler := &mockPriceUpdateHandler{}
			rm.SetPriceUpdateHandler(handler)

			// Create route
			route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
			rm.StartRoute(route.ID, 789)

			// Modify first cargo item quantity for testing
			route.Cargo[0].Quantity = tt.quantity

			// Complete route
			route.Status = StatusCompleted
			rm.completeRoute(route)

			// Check first update (for modified cargo item)
			if len(handler.updates) == 0 {
				t.Fatal("No price updates received")
			}

			impact := handler.updates[0].priceChange
			if impact < tt.expectedMinImpact || impact > tt.expectedMaxImpact {
				t.Errorf("Price impact %.4f not in expected range [%.4f, %.4f]",
					impact, tt.expectedMinImpact, tt.expectedMaxImpact)
			}
		})
	}
}

// TestCompleteRouteThreadSafety verifies concurrent route completions are safe.
func TestCompleteRouteThreadSafety(t *testing.T) {
	rm := NewRouteManager("test-server", 11111)
	handler := &mockPriceUpdateHandler{}
	rm.SetPriceUpdateHandler(handler)

	// Create multiple routes
	routes := make([]*TradeRoute, 10)
	for i := 0; i < 10; i++ {
		route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
		if err != nil {
			t.Fatalf("Failed to create route %d: %v", i, err)
		}
		err = rm.StartRoute(route.ID, uint64(1000+i))
		if err != nil {
			t.Fatalf("Failed to start route %d: %v", i, err)
		}
		route.Status = StatusCompleted
		routes[i] = route
	}

	// Complete all routes concurrently (hold lock for each call since
	// completeRoute is an internal method that expects caller to hold lock)
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(route *TradeRoute) {
			rm.mu.Lock()
			rm.completeRoute(route)
			rm.mu.Unlock()
			done <- true
		}(routes[i])
	}

	// Wait for all completions
	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for route completions")
		}
	}

	// Verify all updates were received (no data races)
	updates := handler.getUpdates()
	if len(updates) == 0 {
		t.Error("Expected price updates from concurrent completions")
	}
}

// TestCompleteRouteZeroQuantityCargo verifies items with zero quantity are skipped.
func TestCompleteRouteZeroQuantityCargo(t *testing.T) {
	rm := NewRouteManager("test-server", 22222)
	handler := &mockPriceUpdateHandler{}
	rm.SetPriceUpdateHandler(handler)

	// Create route
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	rm.StartRoute(route.ID, 999)

	// Set all cargo quantities to zero (simulating total loss)
	for i := range route.Cargo {
		route.Cargo[i].Quantity = 0
	}

	// Complete route
	route.Status = StatusCompleted
	rm.completeRoute(route)

	// Verify no price updates for zero-quantity cargo
	if len(handler.updates) != 0 {
		t.Errorf("Expected no price updates for zero-quantity cargo, got %d", len(handler.updates))
	}
}

// TestCompleteRoutePartialCargo verifies only delivered cargo affects prices.
func TestCompleteRoutePartialCargo(t *testing.T) {
	rm := NewRouteManager("test-server", 33333)
	handler := &mockPriceUpdateHandler{}
	rm.SetPriceUpdateHandler(handler)

	// Create route
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	rm.StartRoute(route.ID, 888)

	// Set half of cargo to zero (partial loss)
	deliveredItems := 0
	for i := range route.Cargo {
		if i%2 == 0 {
			route.Cargo[i].Quantity = 0 // Lost in bandit attack
		} else {
			deliveredItems++
		}
	}

	// Complete route
	route.Status = StatusCompleted
	rm.completeRoute(route)

	// Verify only delivered items generated price updates
	if len(handler.updates) != deliveredItems {
		t.Errorf("Expected %d price updates for delivered items, got %d",
			deliveredItems, len(handler.updates))
	}
}
