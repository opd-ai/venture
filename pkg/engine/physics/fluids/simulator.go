// Package fluids - simulator.go
// This file contains the Simulator implementation for grid-based fluid dynamics simulation.

package fluids

import (
	"fmt"
	"math"
)

// Simulator handles fluid dynamics simulation
type Simulator struct {
	config        SimulationConfig
	grid          *Grid
	time          float64
	bufferA       [][]Cell // First buffer for double-buffering
	bufferB       [][]Cell // Second buffer for double-buffering
	currentBuffer int      // 0 for bufferA, 1 for bufferB
}

// NewSimulator creates a new fluid simulator
func NewSimulator(config SimulationConfig) *Simulator {
	// Pre-allocate two buffers for double-buffering to avoid allocations during updates
	bufferA := allocateGrid(config.GridWidth, config.GridHeight)
	bufferB := allocateGrid(config.GridWidth, config.GridHeight)

	// Reuse bufferA as the initial grid to avoid a third allocation
	grid := &Grid{
		Width:  config.GridWidth,
		Height: config.GridHeight,
		Cells:  bufferA,
	}

	return &Simulator{
		config:        config,
		grid:          grid,
		time:          0.0,
		bufferA:       bufferA,
		bufferB:       bufferB,
		currentBuffer: 0,
	}
}

// allocateGrid creates a 2D grid of cells without initializing the Grid struct.
// Parameters follow (width, height) order to match NewGrid.
func allocateGrid(width, height int) [][]Cell {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
	}
	return cells
}

// NewGrid creates a new fluid grid
func NewGrid(width, height int) *Grid {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = Cell{
				Pressure:  0.0,
				VelocityX: 0.0,
				VelocityY: 0.0,
				Type:      FluidWater,
				Amount:    0.0,
			}
		}
	}
	return &Grid{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

// Update advances the fluid simulation by deltaTime
func (s *Simulator) Update(deltaTime float64) {
	s.time += deltaTime

	// Apply gravity
	s.applyGravity(deltaTime)

	// Calculate pressure
	s.calculatePressure()

	// Apply pressure forces
	s.applyPressure(deltaTime)

	// Apply viscosity
	s.applyViscosity(deltaTime)

	// Advect fluid
	s.advect(deltaTime)

	// Enforce boundaries
	s.enforceBoundaries()
}

// applyGravity applies gravitational force to fluid
func (s *Simulator) applyGravity(deltaTime float64) {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	for y := 0; y < s.grid.Height; y++ {
		for x := 0; x < s.grid.Width; x++ {
			cell := &s.grid.Cells[y][x]
			if cell.Amount > 0.0 {
				cell.VelocityY += s.config.Gravity * deltaTime
			}
		}
	}
}

// calculatePressure calculates pressure based on fluid amount
func (s *Simulator) calculatePressure() {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	for y := 0; y < s.grid.Height; y++ {
		for x := 0; x < s.grid.Width; x++ {
			cell := &s.grid.Cells[y][x]

			// Pressure is proportional to amount and depth
			depth := float64(s.grid.Height - y)
			cell.Pressure = cell.Amount * depth * s.config.PressureFactor
		}
	}
}

// applyPressure applies pressure forces to fluid
func (s *Simulator) applyPressure(deltaTime float64) {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	for y := 1; y < s.grid.Height-1; y++ {
		for x := 1; x < s.grid.Width-1; x++ {
			cell := &s.grid.Cells[y][x]
			if cell.Amount == 0.0 {
				continue
			}

			// Calculate pressure gradients
			pressureLeft := s.grid.Cells[y][x-1].Pressure
			pressureRight := s.grid.Cells[y][x+1].Pressure
			pressureUp := s.grid.Cells[y-1][x].Pressure
			pressureDown := s.grid.Cells[y+1][x].Pressure

			// Apply pressure forces
			cell.VelocityX += (pressureLeft - pressureRight) * deltaTime
			cell.VelocityY += (pressureUp - pressureDown) * deltaTime
		}
	}
}

// applyViscosity applies viscous damping to fluid
func (s *Simulator) applyViscosity(deltaTime float64) {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	for y := 0; y < s.grid.Height; y++ {
		for x := 0; x < s.grid.Width; x++ {
			cell := &s.grid.Cells[y][x]
			if cell.Amount == 0.0 {
				continue
			}

			props := GetFluidProperties(cell.Type)
			damping := math.Pow(s.config.ViscosityFactor, props.Viscosity)

			cell.VelocityX *= damping
			cell.VelocityY *= damping
		}
	}
}

// advect moves fluid according to velocity field
func (s *Simulator) advect(deltaTime float64) {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	// Use double-buffering to avoid allocation
	var targetBuffer [][]Cell
	if s.currentBuffer == 0 {
		targetBuffer = s.bufferB
	} else {
		targetBuffer = s.bufferA
	}

	// Copy current grid to target buffer
	// Note: Using copy() row-by-row ensures correctness during double-buffering.
	// While pointer swapping would be faster, it would be unsafe as it could
	// expose partially-updated state to concurrent readers of the grid.
	for y := 0; y < s.grid.Height; y++ {
		copy(targetBuffer[y], s.grid.Cells[y])
	}

	s.advectFluidCells(targetBuffer, deltaTime)

	// Safe to reassign Cells while holding grid.mu for internal simulator use.
	// NOTE: GetGrid returns a shallow view of the Grid; callers must not retain or
	// mutate grid.Cells without their own synchronization and must not rely on it
	// remaining stable across simulation steps.
	s.grid.Cells = targetBuffer

	// Swap buffers for next update
	s.currentBuffer = 1 - s.currentBuffer
}

// advectFluidCells processes all cells and distributes fluid based on velocity.
func (s *Simulator) advectFluidCells(newCells [][]Cell, deltaTime float64) {
	for y := 1; y < s.grid.Height-1; y++ {
		for x := 1; x < s.grid.Width-1; x++ {
			cell := &s.grid.Cells[y][x]
			if cell.Amount == 0.0 {
				continue
			}
			flowX := cell.VelocityX * deltaTime
			flowY := cell.VelocityY * deltaTime
			s.distributeFluidToNeighbors(newCells, x, y, cell.Amount, flowX, flowY)
		}
	}
}

// distributeFluidToNeighbors transfers fluid to adjacent cells based on flow.
func (s *Simulator) distributeFluidToNeighbors(newCells [][]Cell, x, y int, amount, flowX, flowY float64) {
	if math.Abs(flowX) <= 0.01 && math.Abs(flowY) <= 0.01 {
		return
	}

	s.distributeHorizontalFlow(newCells, x, y, amount, flowX)
	s.distributeVerticalFlow(newCells, x, y, amount, flowY)
}

// distributeHorizontalFlow handles left-right fluid transfer.
func (s *Simulator) distributeHorizontalFlow(newCells [][]Cell, x, y int, amount, flowX float64) {
	if flowX > 0 && x < s.grid.Width-1 {
		transfer := math.Min(amount*0.1, flowX)
		newCells[y][x].Amount -= transfer
		newCells[y][x+1].Amount += transfer
	} else if flowX < 0 && x > 0 {
		transfer := math.Min(amount*0.1, -flowX)
		newCells[y][x].Amount -= transfer
		newCells[y][x-1].Amount += transfer
	}
}

// distributeVerticalFlow handles up-down fluid transfer.
func (s *Simulator) distributeVerticalFlow(newCells [][]Cell, x, y int, amount, flowY float64) {
	if flowY > 0 && y < s.grid.Height-1 {
		transfer := math.Min(amount*0.1, flowY)
		newCells[y][x].Amount -= transfer
		newCells[y+1][x].Amount += transfer
	} else if flowY < 0 && y > 0 {
		transfer := math.Min(amount*0.1, -flowY)
		newCells[y][x].Amount -= transfer
		newCells[y-1][x].Amount += transfer
	}
}

// enforceBoundaries ensures fluid stays within grid bounds
func (s *Simulator) enforceBoundaries() {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	// Top and bottom boundaries
	for x := 0; x < s.grid.Width; x++ {
		s.grid.Cells[0][x].VelocityY = math.Max(0, s.grid.Cells[0][x].VelocityY)
		s.grid.Cells[s.grid.Height-1][x].VelocityY = math.Min(0, s.grid.Cells[s.grid.Height-1][x].VelocityY)
	}

	// Left and right boundaries
	for y := 0; y < s.grid.Height; y++ {
		s.grid.Cells[y][0].VelocityX = math.Max(0, s.grid.Cells[y][0].VelocityX)
		s.grid.Cells[y][s.grid.Width-1].VelocityX = math.Min(0, s.grid.Cells[y][s.grid.Width-1].VelocityX)
	}
}

// AddFluid adds fluid to the grid at a specific position
func (s *Simulator) AddFluid(x, y int, amount float64, fluidType FluidType) error {
	if x < 0 || x >= s.grid.Width || y < 0 || y >= s.grid.Height {
		return fmt.Errorf("position (%d, %d) out of bounds", x, y)
	}

	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	cell := &s.grid.Cells[y][x]
	cell.Amount = math.Min(1.0, cell.Amount+amount)
	cell.Type = fluidType

	return nil
}

// RemoveFluid removes fluid from the grid at a specific position
func (s *Simulator) RemoveFluid(x, y int, amount float64) error {
	if x < 0 || x >= s.grid.Width || y < 0 || y >= s.grid.Height {
		return fmt.Errorf("position (%d, %d) out of bounds", x, y)
	}

	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	cell := &s.grid.Cells[y][x]
	cell.Amount = math.Max(0.0, cell.Amount-amount)

	if cell.Amount == 0.0 {
		cell.VelocityX = 0.0
		cell.VelocityY = 0.0
		cell.Pressure = 0.0
	}

	return nil
}

// GetFluidAt returns the fluid amount at a specific position
func (s *Simulator) GetFluidAt(x, y int) (float64, FluidType, error) {
	if x < 0 || x >= s.grid.Width || y < 0 || y >= s.grid.Height {
		return 0.0, FluidWater, fmt.Errorf("position (%d, %d) out of bounds", x, y)
	}

	s.grid.mu.RLock()
	defer s.grid.mu.RUnlock()

	cell := s.grid.Cells[y][x]
	return cell.Amount, cell.Type, nil
}

// GetGrid returns a shallow copy of the simulation grid for read-only inspection.
//
// WARNING: The returned Grid shares its underlying Cells backing array with the
// simulator's internal grid. Callers must NOT:
//   - Retain references to the returned Grid across simulation steps (it may change)
//   - Mutate the Cells without their own external synchronization
//   - Assume the data remains stable between GetGrid() calls
//
// For safe iteration, callers should copy the data they need immediately or
// use GetFluidAt() for individual cell access which provides proper locking.
//
// This method acquires a read lock for the duration of the copy operation,
// ensuring consistency within a single call, but not across multiple calls.
func (s *Simulator) GetGrid() *Grid {
	s.grid.mu.RLock()
	defer s.grid.mu.RUnlock()

	// Return a shallow copy
	return &Grid{
		Width:  s.grid.Width,
		Height: s.grid.Height,
		Cells:  s.grid.Cells,
	}
}

// Reset clears all fluid from the simulation
func (s *Simulator) Reset() {
	s.grid.mu.Lock()
	defer s.grid.mu.Unlock()

	for y := 0; y < s.grid.Height; y++ {
		for x := 0; x < s.grid.Width; x++ {
			s.grid.Cells[y][x] = Cell{
				Pressure:  0.0,
				VelocityX: 0.0,
				VelocityY: 0.0,
				Type:      FluidWater,
				Amount:    0.0,
			}
		}
	}

	s.time = 0.0
}

// GetConfig returns a copy of the simulator's configuration.
//
// The returned SimulationConfig is a value copy, so modifications to it will
// not affect the simulator's internal configuration. This provides safe access
// to configuration values for debugging, logging, or validation purposes.
//
// Configuration includes grid dimensions, gravity, viscosity factor, pressure
// factor, and update rate settings. These values are set at simulator creation
// time and remain constant throughout the simulator's lifetime.
func (s *Simulator) GetConfig() SimulationConfig {
	return s.config
}
