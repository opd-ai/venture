package housing

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()

	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	if !m.IsEnabled() {
		t.Error("Manager should be enabled by default")
	}

	if m.PlotCount() != 0 {
		t.Errorf("PlotCount() = %v, want 0", m.PlotCount())
	}
}

func TestManagerSetEnabled(t *testing.T) {
	m := NewManager()

	m.SetEnabled(false)
	if m.IsEnabled() {
		t.Error("IsEnabled() = true, want false")
	}

	m.SetEnabled(true)
	if !m.IsEnabled() {
		t.Error("IsEnabled() = false, want true")
	}
}

func TestPlacePlot(t *testing.T) {
	m := NewManager()

	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)

	err := m.PlacePlot(plot)
	if err != nil {
		t.Fatalf("PlacePlot() error = %v, want nil", err)
	}

	if m.PlotCount() != 1 {
		t.Errorf("PlotCount() = %v, want 1", m.PlotCount())
	}

	// Verify plot can be retrieved
	retrieved, ok := m.GetPlot(plot.ID)
	if !ok {
		t.Error("GetPlot() returned false, want true")
	}
	if retrieved.ID != plot.ID {
		t.Errorf("Retrieved plot ID = %v, want %v", retrieved.ID, plot.ID)
	}
}

func TestPlacePlotDisabled(t *testing.T) {
	m := NewManager()
	m.SetEnabled(false)

	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)

	err := m.PlacePlot(plot)
	if err == nil {
		t.Error("PlacePlot() with disabled manager should return error")
	}
}

func TestPlacePlotNil(t *testing.T) {
	m := NewManager()

	err := m.PlacePlot(nil)
	if err == nil {
		t.Error("PlacePlot(nil) should return error")
	}
}

func TestPlacePlotNoOwner(t *testing.T) {
	m := NewManager()

	plot := NewPlot("", Vector2{X: 100, Y: 100}, SizeMedium)

	err := m.PlacePlot(plot)
	if err == nil {
		t.Error("PlacePlot() with no owner should return error")
	}
}

func TestPlacePlotOverlap(t *testing.T) {
	m := NewManager()

	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	err := m.PlacePlot(plot1)
	if err != nil {
		t.Fatalf("PlacePlot() error = %v, want nil", err)
	}

	// Try to place overlapping plot
	plot2 := NewPlot("player2", Vector2{X: 105, Y: 105}, SizeMedium)
	err = m.PlacePlot(plot2)
	if err == nil {
		t.Error("PlacePlot() with overlapping plot should return error")
	}
}

func TestRemovePlot(t *testing.T) {
	m := NewManager()

	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	m.PlacePlot(plot)

	err := m.RemovePlot(plot.ID)
	if err != nil {
		t.Fatalf("RemovePlot() error = %v, want nil", err)
	}

	if m.PlotCount() != 0 {
		t.Errorf("PlotCount() = %v, want 0", m.PlotCount())
	}

	// Verify plot cannot be retrieved
	_, ok := m.GetPlot(plot.ID)
	if ok {
		t.Error("GetPlot() returned true after removal, want false")
	}
}

func TestRemovePlotNotFound(t *testing.T) {
	m := NewManager()

	err := m.RemovePlot("nonexistent")
	if err == nil {
		t.Error("RemovePlot() with nonexistent ID should return error")
	}
}

func TestGetPlayerPlots(t *testing.T) {
	m := NewManager()

	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)
	plot3 := NewPlot("player2", Vector2{X: 300, Y: 300}, SizeLarge)

	m.PlacePlot(plot1)
	m.PlacePlot(plot2)
	m.PlacePlot(plot3)

	player1Plots := m.GetPlayerPlots("player1")
	if len(player1Plots) != 2 {
		t.Errorf("GetPlayerPlots(player1) returned %v plots, want 2", len(player1Plots))
	}

	player2Plots := m.GetPlayerPlots("player2")
	if len(player2Plots) != 1 {
		t.Errorf("GetPlayerPlots(player2) returned %v plots, want 1", len(player2Plots))
	}
}

func TestGetPlotsInArea(t *testing.T) {
	m := NewManager()

	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)
	plot3 := NewPlot("player2", Vector2{X: 300, Y: 300}, SizeLarge)

	m.PlacePlot(plot1)
	m.PlacePlot(plot2)
	m.PlacePlot(plot3)

	// Query area around plot1
	min := Vector2{X: 50, Y: 50}
	max := Vector2{X: 150, Y: 150}
	plots := m.GetPlotsInArea(min, max)

	if len(plots) != 1 {
		t.Errorf("GetPlotsInArea() returned %v plots, want 1", len(plots))
	}
	if len(plots) > 0 && plots[0].ID != plot1.ID {
		t.Errorf("GetPlotsInArea() returned plot %v, want %v", plots[0].ID, plot1.ID)
	}
}

func TestGetAllPlots(t *testing.T) {
	m := NewManager()

	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)

	m.PlacePlot(plot1)
	m.PlacePlot(plot2)

	plots := m.GetAllPlots()
	if len(plots) != 2 {
		t.Errorf("GetAllPlots() returned %v plots, want 2", len(plots))
	}
}

func TestClear(t *testing.T) {
	m := NewManager()

	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)

	m.PlacePlot(plot1)
	m.PlacePlot(plot2)

	m.Clear()

	if m.PlotCount() != 0 {
		t.Errorf("PlotCount() = %v after Clear(), want 0", m.PlotCount())
	}

	plots := m.GetAllPlots()
	if len(plots) != 0 {
		t.Errorf("GetAllPlots() returned %v plots after Clear(), want 0", len(plots))
	}
}

func BenchmarkPlacePlot(b *testing.B) {
	m := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plot := NewPlot("player1", Vector2{X: float64(i * 50), Y: float64(i * 50)}, SizeMedium)
		m.PlacePlot(plot)
	}
}

func BenchmarkGetPlotsInArea(b *testing.B) {
	m := NewManager()

	// Setup: Add 1000 plots
	for i := 0; i < 1000; i++ {
		plot := NewPlot("player1", Vector2{X: float64(i * 50), Y: float64(i * 50)}, SizeMedium)
		m.PlacePlot(plot)
	}

	min := Vector2{X: 0, Y: 0}
	max := Vector2{X: 1000, Y: 1000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetPlotsInArea(min, max)
	}
}

// TestCreateHouse tests the CreateHouse method with seed-based positioning
func TestCreateHouse(t *testing.T) {
	m := NewManager()
	
	// Test basic house creation
	houseID, err := m.CreateHouse("player1", nil, 12345)
	if err != nil {
		t.Fatalf("CreateHouse() error = %v, want nil", err)
	}
	
	if houseID == "" {
		t.Error("CreateHouse() returned empty house ID")
	}
	
	// Verify house was added to manager
	if m.PlotCount() != 1 {
		t.Errorf("PlotCount() = %v, want 1", m.PlotCount())
	}
	
	// Verify GetHouse returns the created house
	house := m.GetHouse(houseID)
	if house == nil {
		t.Fatal("GetHouse() returned nil")
	}
	if house.OwnerID != "player1" {
		t.Errorf("House.OwnerID = %v, want player1", house.OwnerID)
	}
}

// TestCreateHouseMultiple tests creating multiple houses with different seeds
func TestCreateHouseMultiple(t *testing.T) {
	m := NewManager()
	
	// Create 5 houses for the same player with different seeds
	positions := make(map[string]Vector2)
	for i := 0; i < 5; i++ {
		houseID, err := m.CreateHouse("player1", nil, int64(i*1000))
		if err != nil {
			t.Fatalf("CreateHouse(%d) error = %v, want nil", i, err)
		}
		
		house := m.GetHouse(houseID)
		if house == nil {
			t.Fatalf("GetHouse(%s) returned nil", houseID)
		}
		
		positions[houseID] = house.Plot.Position
	}
	
	// Verify all houses have different positions (no overlaps at 0,0)
	if len(positions) != 5 {
		t.Errorf("Expected 5 unique positions, got %d", len(positions))
	}
	
	// Verify at least one house is not at (0,0)
	foundNonZero := false
	for _, pos := range positions {
		if pos.X != 0 || pos.Y != 0 {
			foundNonZero = true
			break
		}
	}
	
	if !foundNonZero {
		t.Error("All houses positioned at (0,0), spatial distribution not working")
	}
}

// TestCreateHouseDeterministic tests that same seed produces same position
func TestCreateHouseDeterministic(t *testing.T) {
	const seed = int64(42)
	
	// Create house in first manager
	m1 := NewManager()
	houseID1, err := m1.CreateHouse("player1", nil, seed)
	if err != nil {
		t.Fatalf("CreateHouse() error = %v, want nil", err)
	}
	house1 := m1.GetHouse(houseID1)
	
	// Create house in second manager with same seed
	m2 := NewManager()
	houseID2, err := m2.CreateHouse("player1", nil, seed)
	if err != nil {
		t.Fatalf("CreateHouse() error = %v, want nil", err)
	}
	house2 := m2.GetHouse(houseID2)
	
	// Positions should be identical
	if house1.Plot.Position.X != house2.Plot.Position.X || 
	   house1.Plot.Position.Y != house2.Plot.Position.Y {
		t.Errorf("Position mismatch: (%v,%v) != (%v,%v)", 
			house1.Plot.Position.X, house1.Plot.Position.Y,
			house2.Plot.Position.X, house2.Plot.Position.Y)
	}
}

// TestCreateHouseDisabled tests that CreateHouse fails when manager is disabled
func TestCreateHouseDisabled(t *testing.T) {
	m := NewManager()
	m.SetEnabled(false)
	
	_, err := m.CreateHouse("player1", nil, 12345)
	if err == nil {
		t.Error("CreateHouse() with disabled manager should return error")
	}
}

