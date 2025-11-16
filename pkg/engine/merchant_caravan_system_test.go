package engine

import (
	"testing"
	"time"
)

func TestNewMerchantCaravanSystem(t *testing.T) {
	world := NewWorld()
	sys := NewMerchantCaravanSystem(world)

	if sys == nil {
		t.Fatal("NewMerchantCaravanSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.hopDuration != 300.0 {
		t.Errorf("hopDuration = %f, want 300.0", sys.hopDuration)
	}
	if sys.priceMarkupMin != 1.1 {
		t.Errorf("priceMarkupMin = %f, want 1.1", sys.priceMarkupMin)
	}
	if sys.priceMarkupMax != 1.5 {
		t.Errorf("priceMarkupMax = %f, want 1.5", sys.priceMarkupMax)
	}
}

func TestCreateCaravan(t *testing.T) {
	world := NewWorld()
	world.Update(0) // Commit entity IDs
	sys := NewMerchantCaravanSystem(world)

	inventory := []CaravanItem{
		{ItemID: "sword1", Quantity: 5, PurchasePrice: 100, SalePrice: 120, OriginServer: "server1"},
	}

	entity := sys.CreateCaravan("server1", "server2", inventory)
	if entity == nil {
		t.Fatal("CreateCaravan returned nil")
	}

	compInt, ok := entity.GetComponent("merchantcaravan")
	if !ok {
		t.Fatal("caravan component not added")
	}

	caravan := compInt.(*MerchantCaravanComponent)
	if caravan.OriginServer != "server1" {
		t.Errorf("OriginServer = %s, want server1", caravan.OriginServer)
	}
	if caravan.DestinationServer != "server2" {
		t.Errorf("DestinationServer = %s, want server2", caravan.DestinationServer)
	}
	if len(caravan.Inventory) != 1 {
		t.Errorf("Inventory length = %d, want 1", len(caravan.Inventory))
	}
	if caravan.CurrentServer != "server1" {
		t.Errorf("CurrentServer = %s, want server1", caravan.CurrentServer)
	}
}

func TestCaravanUpdate_TravelProgress(t *testing.T) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)
	sys.SetHopDuration(10.0) // 10 seconds for faster testing

	entity := sys.CreateCaravan("server1", "server2", nil)
	world.Update(0) // Commit

	compInt, _ := entity.GetComponent("merchantcaravan")
	comp := compInt.(*MerchantCaravanComponent)

	// Update with 5 seconds (half of hop duration)
	sys.Update(5.0)
	if comp.TravelProgress < 0.49 || comp.TravelProgress > 0.51 {
		t.Errorf("TravelProgress = %f, want ~0.5", comp.TravelProgress)
	}

	// Update with another 5 seconds (should complete hop)
	sys.Update(5.0)
	if comp.CurrentRouteIndex != 1 {
		t.Errorf("CurrentRouteIndex = %d, want 1", comp.CurrentRouteIndex)
	}
	if comp.CurrentServer != "server2" {
		t.Errorf("CurrentServer = %s, want server2", comp.CurrentServer)
	}
}

func TestCaravanUpdate_Arrival(t *testing.T) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)
	sys.SetHopDuration(1.0) // 1 second for fast testing

	entity := sys.CreateCaravan("server1", "server2", nil)
	world.Update(0)

	compInt, _ := entity.GetComponent("merchantcaravan")
	comp := compInt.(*MerchantCaravanComponent)

	// Complete the journey
	sys.Update(1.5)

	// Should be resting at destination
	if comp.NextDepartureTime == 0 {
		t.Error("NextDepartureTime should be set after arrival")
	}
	if comp.CurrentServer != "server2" {
		t.Errorf("CurrentServer = %s, want server2 (destination)", comp.CurrentServer)
	}
}

func TestCaravanUpdate_RestPeriod(t *testing.T) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)
	sys.SetHopDuration(1.0)

	entity := sys.CreateCaravan("server1", "server2", nil)
	world.Update(0)

	compInt, _ := entity.GetComponent("merchantcaravan")
	comp := compInt.(*MerchantCaravanComponent)

	// Arrive at destination
	sys.Update(1.5)

	restTime := comp.NextDepartureTime

	// Try to update during rest - should not progress
	sys.Update(1.0)

	if comp.NextDepartureTime != restTime {
		t.Error("NextDepartureTime changed during rest period")
	}
}

func TestGetCaravansAtServer(t *testing.T) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)

	// Create caravans
	sys.CreateCaravan("server1", "server2", nil)
	sys.CreateCaravan("server1", "server3", nil)
	sys.CreateCaravan("server2", "server1", nil)
	world.Update(0)

	// All should start at their origin servers
	caravansAt1 := sys.GetCaravansAtServer("server1")
	if len(caravansAt1) != 2 {
		t.Errorf("Caravans at server1 = %d, want 2", len(caravansAt1))
	}

	caravansAt2 := sys.GetCaravansAtServer("server2")
	if len(caravansAt2) != 1 {
		t.Errorf("Caravans at server2 = %d, want 1", len(caravansAt2))
	}
}

func TestEstimateArrivalTime(t *testing.T) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)
	sys.SetHopDuration(100.0) // 100 seconds per hop

	entity := sys.CreateCaravan("server1", "server2", nil)
	world.Update(0)

	// Should arrive in approximately 100 seconds
	eta := sys.EstimateArrivalTime(entity)
	expectedETA := time.Now().Unix() + 100

	if eta < expectedETA-5 || eta > expectedETA+5 {
		t.Errorf("ETA = %d, want ~%d (within 5s)", eta, expectedETA)
	}
}

func TestEstimateArrivalTime_InProgress(t *testing.T) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)
	sys.SetHopDuration(100.0)

	entity := sys.CreateCaravan("server1", "server2", nil)
	world.Update(0)

	// Travel halfway
	sys.Update(50.0)

	// Should arrive in approximately 50 seconds
	eta := sys.EstimateArrivalTime(entity)
	expectedETA := time.Now().Unix() + 50

	if eta < expectedETA-5 || eta > expectedETA+5 {
		t.Errorf("ETA = %d, want ~%d (within 5s)", eta, expectedETA)
	}
}

func TestCalculateSalePrice(t *testing.T) {
	world := NewWorld()
	sys := NewMerchantCaravanSystem(world)

	tests := []struct {
		name          string
		purchasePrice float64
		serverHops    int
		wantMin       float64
		wantMax       float64
	}{
		{"zero hops", 100.0, 0, 109.9, 110.1},
		{"one hop", 100.0, 1, 113.0, 115.0},
		{"five hops", 100.0, 5, 129.0, 131.0},
		{"ten hops", 100.0, 10, 149.0, 151.0},
		{"many hops", 100.0, 20, 149.0, 151.0}, // Capped at max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := sys.CalculateSalePrice(tt.purchasePrice, tt.serverHops)
			if price < tt.wantMin || price > tt.wantMax {
				t.Errorf("CalculateSalePrice = %f, want between %f and %f", price, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSetRouteCalculator(t *testing.T) {
	world := NewWorld()
	sys := NewMerchantCaravanSystem(world)

	called := false
	customCalc := func(origin, destination string) []string {
		called = true
		return []string{origin, "intermediate", destination}
	}

	sys.SetRouteCalculator(customCalc)
	sys.CreateCaravan("server1", "server2", nil)

	if !called {
		t.Error("Custom route calculator not called")
	}
}

func TestDefaultRouteCalculator(t *testing.T) {
	route := defaultRouteCalculator("server1", "server2")
	if len(route) != 2 {
		t.Errorf("Route length = %d, want 2", len(route))
	}
	if route[0] != "server1" || route[1] != "server2" {
		t.Errorf("Route = %v, want [server1, server2]", route)
	}
}

func TestDefaultRouteCalculator_SameServer(t *testing.T) {
	route := defaultRouteCalculator("server1", "server1")
	if len(route) != 1 {
		t.Errorf("Route length = %d, want 1 for same server", len(route))
	}
	if route[0] != "server1" {
		t.Errorf("Route = %v, want [server1]", route)
	}
}

func BenchmarkCaravanUpdate(b *testing.B) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)

	// Create 100 caravans
	for i := 0; i < 100; i++ {
		sys.CreateCaravan("server1", "server2", nil)
	}
	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(0.016) // 60 FPS
	}
}

func BenchmarkCreateCaravan(b *testing.B) {
	world := NewWorld()
	world.Update(0)
	sys := NewMerchantCaravanSystem(world)

	inventory := []CaravanItem{
		{ItemID: "item1", Quantity: 10, PurchasePrice: 100, SalePrice: 120},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CreateCaravan("server1", "server2", inventory)
	}
}
