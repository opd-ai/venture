package fluids

import (
	"testing"
)

func TestNewSimulator(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)

	if sim == nil {
		t.Fatal("NewSimulator returned nil")
	}
	if sim.grid == nil {
		t.Error("Simulator grid is nil")
	}
	if sim.grid.Width != config.GridWidth {
		t.Errorf("Grid width = %v, want %v", sim.grid.Width, config.GridWidth)
	}
	if sim.grid.Height != config.GridHeight {
		t.Errorf("Grid height = %v, want %v", sim.grid.Height, config.GridHeight)
	}
	if sim.time != 0.0 {
		t.Errorf("Initial time = %v, want 0.0", sim.time)
	}
}

func TestNewGrid(t *testing.T) {
	grid := NewGrid(10, 20)

	if grid.Width != 10 {
		t.Errorf("Grid width = %v, want 10", grid.Width)
	}
	if grid.Height != 20 {
		t.Errorf("Grid height = %v, want 20", grid.Height)
	}
	if len(grid.Cells) != 20 {
		t.Errorf("Cells length = %v, want 20", len(grid.Cells))
	}
	if len(grid.Cells[0]) != 10 {
		t.Errorf("Cells[0] length = %v, want 10", len(grid.Cells[0]))
	}

	// Check all cells initialized to zero
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y][x]
			if cell.Amount != 0.0 {
				t.Errorf("Cell[%d][%d].Amount = %v, want 0.0", y, x, cell.Amount)
			}
			if cell.Pressure != 0.0 {
				t.Errorf("Cell[%d][%d].Pressure = %v, want 0.0", y, x, cell.Pressure)
			}
		}
	}
}

func TestAddFluid(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	err := sim.AddFluid(5, 5, 0.5, FluidWater)
	if err != nil {
		t.Errorf("AddFluid failed: %v", err)
	}

	amount, fluidType, err := sim.GetFluidAt(5, 5)
	if err != nil {
		t.Errorf("GetFluidAt failed: %v", err)
	}
	if amount != 0.5 {
		t.Errorf("Fluid amount = %v, want 0.5", amount)
	}
	if fluidType != FluidWater {
		t.Errorf("Fluid type = %v, want FluidWater", fluidType)
	}
}

func TestAddFluid_OutOfBounds(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	tests := []struct {
		name string
		x    int
		y    int
	}{
		{"Negative X", -1, 5},
		{"Negative Y", 5, -1},
		{"X Too Large", 10, 5},
		{"Y Too Large", 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sim.AddFluid(tt.x, tt.y, 0.5, FluidWater)
			if err == nil {
				t.Error("AddFluid should return error for out of bounds position")
			}
		})
	}
}

func TestAddFluid_Clamping(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	// Add more than 1.0 total
	sim.AddFluid(5, 5, 0.6, FluidWater)
	sim.AddFluid(5, 5, 0.7, FluidWater)

	amount, _, _ := sim.GetFluidAt(5, 5)
	if amount > 1.0 {
		t.Errorf("Fluid amount = %v, should be clamped to 1.0", amount)
	}
}

func TestRemoveFluid(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	sim.AddFluid(5, 5, 0.8, FluidWater)
	err := sim.RemoveFluid(5, 5, 0.3)
	if err != nil {
		t.Errorf("RemoveFluid failed: %v", err)
	}

	amount, _, _ := sim.GetFluidAt(5, 5)
	if amount != 0.5 {
		t.Errorf("Fluid amount = %v, want 0.5", amount)
	}
}

func TestRemoveFluid_CompleteRemoval(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	sim.AddFluid(5, 5, 0.5, FluidWater)
	sim.RemoveFluid(5, 5, 1.0)

	amount, _, _ := sim.GetFluidAt(5, 5)
	if amount != 0.0 {
		t.Errorf("Fluid amount = %v, want 0.0", amount)
	}

	// Verify cell is reset
	grid := sim.GetGrid()
	cell := grid.Cells[5][5]
	if cell.VelocityX != 0.0 || cell.VelocityY != 0.0 || cell.Pressure != 0.0 {
		t.Error("Cell should be fully reset after complete removal")
	}
}

func TestUpdate_TimeAdvancement(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)

	deltaTime := 0.016 // ~60 FPS
	sim.Update(deltaTime)

	if sim.time != deltaTime {
		t.Errorf("Time = %v, want %v", sim.time, deltaTime)
	}

	sim.Update(deltaTime)
	if sim.time != 2*deltaTime {
		t.Errorf("Time = %v, want %v", sim.time, 2*deltaTime)
	}
}

func TestUpdate_Gravity(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	sim.AddFluid(5, 5, 1.0, FluidWater)

	initialVelocityY := sim.grid.Cells[5][5].VelocityY

	sim.Update(0.1)

	finalVelocityY := sim.grid.Cells[5][5].VelocityY

	if finalVelocityY <= initialVelocityY {
		t.Error("Gravity should increase Y velocity")
	}
}

func TestUpdate_PressureCalculation(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	// Add fluid at different depths
	// y=2 is near top (depth = 10-2 = 8)
	// y=8 is near bottom (depth = 10-8 = 2)
	sim.AddFluid(5, 2, 1.0, FluidWater)
	sim.AddFluid(5, 8, 1.0, FluidWater)

	sim.Update(0.01)

	pressureNearTop := sim.grid.Cells[2][5].Pressure
	pressureNearBottom := sim.grid.Cells[8][5].Pressure

	// Fluid near top (y=2, depth=8) should have higher pressure than near bottom (y=8, depth=2)
	if pressureNearTop <= pressureNearBottom {
		t.Errorf("Fluid with more depth above should have higher pressure: top=%v, bottom=%v", pressureNearTop, pressureNearBottom)
	}
}

func TestReset(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	// Add fluid and run simulation
	sim.AddFluid(5, 5, 1.0, FluidWater)
	sim.Update(0.1)

	// Reset
	sim.Reset()

	if sim.time != 0.0 {
		t.Errorf("Time after reset = %v, want 0.0", sim.time)
	}

	// Check all cells are empty
	for y := 0; y < sim.grid.Height; y++ {
		for x := 0; x < sim.grid.Width; x++ {
			cell := sim.grid.Cells[y][x]
			if cell.Amount != 0.0 {
				t.Errorf("Cell[%d][%d].Amount = %v, want 0.0 after reset", y, x, cell.Amount)
			}
		}
	}
}

func TestGetGrid(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)

	grid := sim.GetGrid()

	if grid == nil {
		t.Fatal("GetGrid returned nil")
	}
	if grid.Width != config.GridWidth {
		t.Errorf("Grid width = %v, want %v", grid.Width, config.GridWidth)
	}
	if grid.Height != config.GridHeight {
		t.Errorf("Grid height = %v, want %v", grid.Height, config.GridHeight)
	}
}

func TestUpdate_ViscosityDamping(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)

	// Add water (low viscosity) and lava (high viscosity)
	sim.AddFluid(2, 5, 1.0, FluidWater)
	sim.AddFluid(7, 5, 1.0, FluidLava)

	// Set initial velocities
	sim.grid.Cells[5][2].VelocityX = 10.0
	sim.grid.Cells[5][7].VelocityX = 10.0

	sim.Update(0.1)

	waterVelocity := sim.grid.Cells[5][2].VelocityX
	lavaVelocity := sim.grid.Cells[5][7].VelocityX

	// Both should be damped, but lava more so
	if waterVelocity >= 10.0 {
		t.Error("Water velocity should be damped")
	}
	if lavaVelocity >= waterVelocity {
		t.Error("Lava should be damped more than water")
	}
}

func BenchmarkNewSimulator(b *testing.B) {
	config := DefaultSimulationConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSimulator(config)
	}
}

func BenchmarkUpdate(b *testing.B) {
	config := DefaultSimulationConfig()
	config.GridWidth = 100
	config.GridHeight = 100
	sim := NewSimulator(config)

	// Add some fluid
	for i := 0; i < 100; i++ {
		sim.AddFluid(i, 50, 0.5, FluidWater)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.Update(1.0 / 30.0)
	}
}

func BenchmarkAddFluid(b *testing.B) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % config.GridWidth
		y := (i / config.GridWidth) % config.GridHeight
		sim.AddFluid(x, y, 0.5, FluidWater)
	}
}

func BenchmarkGetFluidAt(b *testing.B) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)
	sim.AddFluid(50, 50, 1.0, FluidWater)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = sim.GetFluidAt(50, 50)
	}
}
