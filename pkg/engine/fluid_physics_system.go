// Package engine provides fluid physics integration with the ECS.
// This file implements the FluidPhysicsSystem that applies buoyancy, swimming,
// and flooding mechanics to entities with the appropriate components.
package engine

import (
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	log "github.com/sirupsen/logrus"
)

// fluidDebugEnabled caches whether debug-level fluid logging is enabled.
var fluidDebugEnabled bool

// FluidPhysicsSystem updates entities with fluid-related components.
// It handles buoyancy calculations, swimming mechanics, and drowning damage.
type FluidPhysicsSystem struct {
	world           *World
	buoyancyCalc    *fluids.BuoyancyCalculator
	swimmingManager *fluids.SwimmingManager
	floodingManager *fluids.FloodingManager
	simulator       *fluids.Simulator
	gravity         float64
	lastDeltaTime   float64
}

// NewFluidPhysicsSystem creates a new fluid physics system.
func NewFluidPhysicsSystem(world *World, config fluids.SimulationConfig) *FluidPhysicsSystem {
	fluidDebugEnabled = log.GetLevel() >= log.DebugLevel

	if fluidDebugEnabled {
		log.WithFields(log.Fields{
			"system_name": "fluid_physics",
			"grid_width":  config.GridWidth,
			"grid_height": config.GridHeight,
			"gravity":     config.Gravity,
		}).Debug("Creating fluid physics system")
	}

	simulator := fluids.NewSimulator(config)
	return &FluidPhysicsSystem{
		world:           world,
		buoyancyCalc:    fluids.NewBuoyancyCalculator(config.Gravity),
		swimmingManager: fluids.NewSwimmingManager(config.Gravity),
		floodingManager: fluids.NewFloodingManager(simulator),
		simulator:       simulator,
		gravity:         config.Gravity,
	}
}

// NewFluidPhysicsSystemWithDefaults creates a fluid physics system with default config.
func NewFluidPhysicsSystemWithDefaults(world *World) *FluidPhysicsSystem {
	return NewFluidPhysicsSystem(world, fluids.DefaultSimulationConfig())
}

// Update processes all entities with fluid components.
func (s *FluidPhysicsSystem) Update(entities []*Entity, deltaTime float64) {
	s.lastDeltaTime = deltaTime

	// Update fluid simulation
	s.simulator.Update(deltaTime)

	for _, entity := range entities {
		s.updateBuoyancy(entity, deltaTime)
		s.updateSwimming(entity, deltaTime)
		s.updateFlooding(entity, deltaTime)
	}
}

// updateBuoyancy handles entities with buoyancy components.
func (s *FluidPhysicsSystem) updateBuoyancy(entity *Entity, deltaTime float64) {
	buoyancyComp, ok := entity.GetComponent("buoyancy")
	if !ok {
		return
	}

	buoyancy, ok := buoyancyComp.(*fluids.BuoyancyComponent)
	if !ok {
		return
	}

	// Get entity position for fluid lookup
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Get fluid amount at entity position
	fluidAmount, fluidType := s.getFluidAtPosition(pos.X, pos.Y)

	// Calculate buoyancy
	s.buoyancyCalc.CalculateBuoyancy(buoyancy, fluidAmount, fluidType)

	// Apply buoyant force to velocity if entity has velocity component
	if buoyancy.Submerged > 0 {
		velComp, ok := entity.GetComponent("velocity")
		if ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				netForce := s.buoyancyCalc.GetNetForce(buoyancy)
				// Apply vertical force (positive = up)
				// F = ma, a = F/m, v += a*dt
				if buoyancy.Mass > 0 {
					vel.VY += (netForce / buoyancy.Mass) * deltaTime
				}
			}
		}
	}
}

// updateSwimming handles entities with swimming components.
func (s *FluidPhysicsSystem) updateSwimming(entity *Entity, deltaTime float64) {
	swimmingComp, ok := entity.GetComponent("swimming")
	if !ok {
		return
	}

	swimming, ok := swimmingComp.(*fluids.SwimmingComponent)
	if !ok {
		return
	}

	// Get buoyancy for integration
	var buoyancy *fluids.BuoyancyComponent
	buoyancyComp, ok := entity.GetComponent("buoyancy")
	if ok {
		buoyancy, _ = buoyancyComp.(*fluids.BuoyancyComponent)
	}

	// Get fluid amount at entity position
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	fluidAmount, _ := s.getFluidAtPosition(pos.X, pos.Y)

	// Update swimming state
	s.swimmingManager.UpdateSwimming(swimming, buoyancy, fluidAmount, deltaTime)

	// Apply drowning damage
	if swimming.Drowning {
		healthComp, ok := entity.GetComponent("health")
		if ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				damage := s.swimmingManager.GetDrowningDamage(swimming) * deltaTime
				health.Current -= damage
				if health.Current < 0 {
					health.Current = 0
				}

				if fluidDebugEnabled {
					log.WithFields(log.Fields{
						"system_name": "fluid_physics",
						"entity_id":   entity.ID,
						"damage":      damage,
						"health":      health.Current,
					}).Debug("Entity drowning")
				}
			}
		}
	}

	// Apply swim speed modifier to velocity
	speedMult := s.swimmingManager.GetSwimSpeedMultiplier(swimming)
	if speedMult < 1.0 {
		velComp, ok := entity.GetComponent("velocity")
		if ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				vel.VX *= speedMult
				vel.VY *= speedMult
			}
		}
	}
}

// updateFlooding handles entities with flooding components (area markers).
func (s *FluidPhysicsSystem) updateFlooding(entity *Entity, deltaTime float64) {
	floodingComp, ok := entity.GetComponent("flooding")
	if !ok {
		return
	}

	flooding, ok := floodingComp.(*fluids.FloodingComponent)
	if !ok {
		return
	}

	// Update flooding state
	s.floodingManager.UpdateFlooding(flooding, deltaTime)

	if fluidDebugEnabled && flooding.FloodLevel > 0 {
		log.WithFields(log.Fields{
			"system_name":      "fluid_physics",
			"area_id":          flooding.AreaID,
			"flood_level":      flooding.FloodLevel,
			"flood_percentage": s.floodingManager.GetFloodPercentage(flooding),
		}).Debug("Flooding update")
	}
}

// getFluidAtPosition returns the fluid amount and type at the given world position.
func (s *FluidPhysicsSystem) getFluidAtPosition(x, y float64) (float64, fluids.FluidType) {
	// Convert world coordinates to grid coordinates
	gridX := int(x / s.simulator.GetConfig().CellSize)
	gridY := int(y / s.simulator.GetConfig().CellSize)

	// Get cell data
	amount, fluidType, err := s.simulator.GetFluidAt(gridX, gridY)
	if err != nil {
		return 0.0, fluids.FluidWater
	}

	return amount, fluidType
}

// GetSimulator returns the fluid simulator for external access.
func (s *FluidPhysicsSystem) GetSimulator() *fluids.Simulator {
	return s.simulator
}

// AddFluid adds fluid to the simulation at the given position.
func (s *FluidPhysicsSystem) AddFluid(worldX, worldY, amount float64, fluidType fluids.FluidType) {
	cellSize := s.simulator.GetConfig().CellSize
	gridX := int(worldX / cellSize)
	gridY := int(worldY / cellSize)
	err := s.simulator.AddFluid(gridX, gridY, amount, fluidType)
	if err != nil && fluidDebugEnabled {
		log.WithFields(log.Fields{
			"system_name": "fluid_physics",
			"world_x":     worldX,
			"world_y":     worldY,
			"error":       err.Error(),
		}).Debug("Failed to add fluid")
		return
	}

	if fluidDebugEnabled {
		log.WithFields(log.Fields{
			"system_name": "fluid_physics",
			"world_x":     worldX,
			"world_y":     worldY,
			"grid_x":      gridX,
			"grid_y":      gridY,
			"amount":      amount,
			"fluid_type":  fluidType.String(),
		}).Debug("Added fluid")
	}
}

// RemoveFluid removes fluid from the simulation at the given position.
func (s *FluidPhysicsSystem) RemoveFluid(worldX, worldY, amount float64) {
	cellSize := s.simulator.GetConfig().CellSize
	gridX := int(worldX / cellSize)
	gridY := int(worldY / cellSize)
	_ = s.simulator.RemoveFluid(gridX, gridY, amount)
}
