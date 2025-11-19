package fluids

import (
	"fmt"
	"math"
)

// Simulator handles fluid dynamics simulation
type Simulator struct {
	config SimulationConfig
	grid   *Grid
	time   float64
}

// NewSimulator creates a new fluid simulator
func NewSimulator(config SimulationConfig) *Simulator {
	grid := NewGrid(config.GridWidth, config.GridHeight)
	return &Simulator{
		config: config,
		grid:   grid,
		time:   0.0,
	}
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

	// Create temporary grid for new values
	newCells := make([][]Cell, s.grid.Height)
	for y := 0; y < s.grid.Height; y++ {
		newCells[y] = make([]Cell, s.grid.Width)
		copy(newCells[y], s.grid.Cells[y])
	}

	// Advect fluid
	for y := 1; y < s.grid.Height-1; y++ {
		for x := 1; x < s.grid.Width-1; x++ {
			cell := &s.grid.Cells[y][x]
			if cell.Amount == 0.0 {
				continue
			}

			// Calculate flow based on velocity
			flowX := cell.VelocityX * deltaTime
			flowY := cell.VelocityY * deltaTime

			// Distribute fluid to neighbors
			if math.Abs(flowX) > 0.01 || math.Abs(flowY) > 0.01 {
				if flowX > 0 && x < s.grid.Width-1 {
					transfer := math.Min(cell.Amount*0.1, flowX)
					newCells[y][x].Amount -= transfer
					newCells[y][x+1].Amount += transfer
				} else if flowX < 0 && x > 0 {
					transfer := math.Min(cell.Amount*0.1, -flowX)
					newCells[y][x].Amount -= transfer
					newCells[y][x-1].Amount += transfer
				}

				if flowY > 0 && y < s.grid.Height-1 {
					transfer := math.Min(cell.Amount*0.1, flowY)
					newCells[y][x].Amount -= transfer
					newCells[y+1][x].Amount += transfer
				} else if flowY < 0 && y > 0 {
					transfer := math.Min(cell.Amount*0.1, -flowY)
					newCells[y][x].Amount -= transfer
					newCells[y-1][x].Amount += transfer
				}
			}
		}
	}

	// Copy back to grid
	s.grid.Cells = newCells
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

// GetGrid returns a read-only copy of the simulation grid
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
