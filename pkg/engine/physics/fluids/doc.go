/*
Package fluids implements fluid dynamics simulation for water, lava, and other liquids.

This package provides a grid-based fluid simulation system using a simplified Navier-Stokes
approximation. It supports multiple fluid types (water, lava, oil, acid, poison) with
distinct physical properties, buoyancy calculations for entities and vehicles, swimming
mechanics with stamina management, and flooding systems for enclosed spaces.

# Overview

The fluid simulation operates on a discrete grid where each cell contains:
  - Pressure: Calculated from fluid amount and depth
  - Velocity: X and Y components for flow direction
  - Amount: Fluid quantity (0.0 = empty, 1.0 = full)
  - Type: Fluid type (water, lava, etc.)

The simulation runs at a configurable update rate (default 30 FPS) separate from the
main game loop to maintain performance while providing smooth fluid behavior.

# Fluid Types

Five fluid types are supported, each with unique properties:

	Water: Low viscosity, neutral buoyancy, no damage
	Lava: High viscosity, high density, 50 damage/sec
	Oil: Medium viscosity, low density, flammable
	Acid: Low viscosity, corrosive, 25 damage/sec
	Poison: Low viscosity, toxic, 15 damage/sec

# Buoyancy System

Entities and vehicles interact with fluids through buoyancy calculations:

	F_buoyant = ρ_fluid * V_submerged * g

Where:
  - ρ_fluid is fluid density (kg/m³)
  - V_submerged is submerged volume (m³)
  - g is gravity (9.81 m/s²)

Entities float when buoyant force exceeds weight (F_buoyant >= m * g).

# Swimming Mechanics

Swimming is managed through stamina-based mechanics:
  - Stamina drains while swimming (default: 10 points/sec)
  - Treading water drains half stamina
  - Stamina regenerates on land (default: 20 points/sec)
  - Drowning occurs when stamina reaches zero
  - Drowning damage: configurable damage/sec (default: 10)

Speed is reduced while swimming based on stamina:

	speed_multiplier = swim_speed * (0.5 + 0.5 * stamina_ratio)

# Flooding System

Enclosed areas can flood through water entry points (FloodSources):
  - Each source contributes flow_rate units/sec
  - Flood level rises until max_flood_level reached
  - Multiple sources combine additively
  - Integration with main fluid simulation

# Performance

The fluid simulation is optimized for large grids:
  - Target: <5ms per update for 100×100 grid at 30 FPS
  - Buoyancy calculations: <100µs per entity
  - Thread-safe grid access with RWMutex
  - Separate update rate from main game loop
  - Zero-allocation updates via double-buffering (recent optimization)

# Memory Management

The simulator uses a double-buffering strategy to eliminate per-update allocations:
  - Two grid buffers (bufferA and bufferB) are pre-allocated at initialization
  - During each Update(), the simulator swaps between buffers
  - This achieves zero allocations per update (down from 412KB/update)
  - Initial allocation cost is higher but amortized over runtime
  - Benchmark: 0 B/op, 0 allocs/op per update (100% allocation reduction)

# Usage Example

Basic fluid simulation setup:

	config := fluids.DefaultSimulationConfig()
	config.GridWidth = 100
	config.GridHeight = 100
	simulator := fluids.NewSimulator(config)

	// Add water at position (50, 10)
	simulator.AddFluid(50, 10, 1.0, fluids.FluidWater)

	// Update simulation at 30 FPS
	ticker := time.NewTicker(time.Second / 30)
	for range ticker.C {
		simulator.Update(1.0 / 30.0)
	}

Buoyancy calculation for an entity:

	// Create buoyancy component (wooden boat)
	buoyancy := &fluids.BuoyancyComponent{
		Mass:   500.0,  // 500 kg
		Volume: 2.0,    // 2 m³
	}

	// Calculate buoyancy in water
	calc := fluids.NewBuoyancyCalculator(9.81)
	calc.UpdateDensity(buoyancy)  // Density = 250 kg/m³ (new method)
	calc.CalculateBuoyancy(buoyancy, 1.0, fluids.FluidWater)

	if buoyancy.Buoyant {
		// Boat floats (water density 1000 kg/m³ > boat density 250 kg/m³)
		logrus.Info("Floating!")
	}

	// Alternative: Use package-level function for convenience
	buoyancy2 := &fluids.BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}
	fluids.UpdateDensity(buoyancy2)  // Package-level function

Swimming mechanics:

	// Create swimming component
	swimming := &fluids.SwimmingComponent{
		Stamina:        100.0,
		MaxStamina:     100.0,
		StaminaDrain:   10.0,  // 10 points/sec
		StaminaRegen:   20.0,  // 20 points/sec on land
		SwimSpeed:      0.5,   // 50% normal speed
		DrowningDamage: 10.0,
	}

	// Update swimming state
	manager := fluids.NewSwimmingManager(9.81)
	manager.UpdateSwimming(swimming, buoyancy, fluidAmount, deltaTime)

	// Get current swim speed multiplier
	speedMult := manager.GetSwimSpeedMultiplier(swimming)

	// Check for drowning damage
	damage := manager.GetDrowningDamage(swimming)

Flooding system:

	// Create flooding component for a room
	flooding := &fluids.FloodingComponent{
		AreaID:        "room_1",
		FloodLevel:    0.0,
		FloodRate:     0.1,  // 0.1 units/sec
		MaxFloodLevel: 1.0,
	}

	// Add water entry points
	floodMgr := fluids.NewFloodingManager(simulator)
	floodMgr.AddFloodSource(flooding, 10, 5, 0.05)  // 0.05 units/sec
	floodMgr.AddFloodSource(flooding, 20, 5, 0.05)  // 0.05 units/sec

	// Update flooding
	floodMgr.UpdateFlooding(flooding, deltaTime)

	// Check flood status
	if floodMgr.IsFullyFlooded(flooding) {
		logrus.Info("Room completely flooded!")
	}

# Safe Grid Inspection

GetGrid() returns a shallow copy of the simulation grid for read-only inspection.
To safely access grid data, copy the cells you need immediately:

	grid := simulator.GetGrid()

	// Safe: Copy cells for later use
	cellsCopy := make([]fluids.Cell, len(grid.Cells))
	copy(cellsCopy, grid.Cells)

	// Now iterate over the copy safely
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := cellsCopy[y*grid.Width+x]
			if cell.Amount > 0.5 {
				logrus.WithFields(logrus.Fields{
					"x": x, "y": y,
					"fluid_type": cell.Type,
					"amount": cell.Amount,
				}).Debug("High fluid level detected")
			}
		}
	}

For single cell queries, use GetFluidAt() which provides proper locking:

	// Safe: Single cell access with locking
	amount, fluidType := simulator.GetFluidAt(50, 10)
	if amount > 0 {
		logrus.WithFields(logrus.Fields{
			"x": 50, "y": 10,
			"amount": amount,
			"fluid_type": fluidType,
		}).Info("Fluid present at location")
	}

# Integration with ECS

The package defines three component types for ECS integration:

	BuoyancyComponent: Attach to entities that interact with fluids
	SwimmingComponent: Attach to entities that can swim
	FloodingComponent: Attach to areas that can flood

Component Registration (optional convenience):

	import "github.com/opd-ai/venture/pkg/engine/physics/fluids"

	// Optional: Call once during initialization for documentation clarity
	fluids.RegisterComponentFactories()

	// Components are then added to entities normally:
	entity.AddComponent(&fluids.BuoyancyComponent{
		Mass:   100.0,
		Volume: 0.2,
	})
	entity.AddComponent(&fluids.SwimmingComponent{
		Stamina:    100.0,
		MaxStamina: 100.0,
	})

Note: Component registration is currently automatic via AddComponent().
The RegisterComponentFactories() function is provided for future-proofing
and as a documentation aid to clearly identify all fluid components.

Use the Simulator, BuoyancyCalculator, SwimmingManager, and FloodingManager
in corresponding systems to update entity state based on fluid interactions.

# Determinism

While fluid simulation uses floating-point calculations that may have minor
platform-dependent variations, the system maintains consistency through:
  - Fixed update rate (independent of frame rate)
  - Convergence thresholds for iterative solvers
  - Clamped values to prevent divergence
  - Reproducible initial conditions

For deterministic gameplay, use buoyancy and swimming components with fixed
parameters rather than relying on exact fluid flow patterns.
*/
package fluids
