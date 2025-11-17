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
