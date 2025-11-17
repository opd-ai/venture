package housing

import (
	"testing"
)

func TestSpatialGridInsertAndQuery(t *testing.T) {
	sg := NewSpatialGrid(64)

	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)

	sg.Insert(plot1)
	sg.Insert(plot2)

	// Query around plot1
	min := Vector2{X: 50, Y: 50}
	max := Vector2{X: 150, Y: 150}
	plots := sg.Query(min, max)

	if len(plots) != 1 {
		t.Errorf("Query() returned %v plots, want 1", len(plots))
	}
	if len(plots) > 0 && plots[0].ID != plot1.ID {
		t.Errorf("Query() returned plot %v, want %v", plots[0].ID, plot1.ID)
	}
}

func TestSpatialGridRemove(t *testing.T) {
	sg := NewSpatialGrid(64)

	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	sg.Insert(plot)

	// Query should find it
	min := Vector2{X: 50, Y: 50}
	max := Vector2{X: 150, Y: 150}
	plots := sg.Query(min, max)
	if len(plots) != 1 {
		t.Fatalf("Query() returned %v plots before remove, want 1", len(plots))
	}

	// Remove it
	sg.Remove(plot)

	// Query should not find it
	plots = sg.Query(min, max)
	if len(plots) != 0 {
		t.Errorf("Query() returned %v plots after remove, want 0", len(plots))
	}
}

func TestSpatialGridMultipleCells(t *testing.T) {
	sg := NewSpatialGrid(64)

	// Large plot that spans multiple cells
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeEstate)
	sg.Insert(plot)

	// Query different areas
	tests := []struct {
		name     string
		min      Vector2
		max      Vector2
		expected int
	}{
		{"center", Vector2{X: 90, Y: 90}, Vector2{X: 110, Y: 110}, 1},
		{"edge", Vector2{X: 115, Y: 115}, Vector2{X: 120, Y: 120}, 1},
		{"far away", Vector2{X: 200, Y: 200}, Vector2{X: 250, Y: 250}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plots := sg.Query(tt.min, tt.max)
			if len(plots) != tt.expected {
				t.Errorf("Query() returned %v plots, want %v", len(plots), tt.expected)
			}
		})
	}
}

func TestSpatialGridNegativeCoordinates(t *testing.T) {
	sg := NewSpatialGrid(64)

	plot := NewPlot("player1", Vector2{X: -100, Y: -100}, SizeMedium)
	sg.Insert(plot)

	min := Vector2{X: -150, Y: -150}
	max := Vector2{X: -50, Y: -50}
	plots := sg.Query(min, max)

	if len(plots) != 1 {
		t.Errorf("Query() with negative coords returned %v plots, want 1", len(plots))
	}
}

func TestSpatialGridDuplicatePrevention(t *testing.T) {
	sg := NewSpatialGrid(64)

	// Large plot that spans multiple cells
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeEstate)
	sg.Insert(plot)

	// Query area that spans multiple cells containing the same plot
	min := Vector2{X: 50, Y: 50}
	max := Vector2{X: 150, Y: 150}
	plots := sg.Query(min, max)

	// Should only return the plot once despite spanning multiple cells
	if len(plots) != 1 {
		t.Errorf("Query() returned %v plots (duplicates), want 1", len(plots))
	}
}

func BenchmarkSpatialGridInsert(b *testing.B) {
	sg := NewSpatialGrid(64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plot := NewPlot("player1", Vector2{X: float64(i), Y: float64(i)}, SizeMedium)
		sg.Insert(plot)
	}
}

func BenchmarkSpatialGridQuery(b *testing.B) {
	sg := NewSpatialGrid(64)

	// Setup: Insert 1000 plots
	for i := 0; i < 1000; i++ {
		plot := NewPlot("player1", Vector2{X: float64(i * 50), Y: float64(i * 50)}, SizeMedium)
		sg.Insert(plot)
	}

	min := Vector2{X: 0, Y: 0}
	max := Vector2{X: 1000, Y: 1000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Query(min, max)
	}
}

func BenchmarkSpatialGridRemove(b *testing.B) {
	sg := NewSpatialGrid(64)

	// Setup: Insert plots
	plots := make([]*Plot, b.N)
	for i := 0; i < b.N; i++ {
		plots[i] = NewPlot("player1", Vector2{X: float64(i * 50), Y: float64(i * 50)}, SizeMedium)
		sg.Insert(plots[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Remove(plots[i])
	}
}
