// Package physics provides advanced physics simulation for vehicles and environmental interactions.
//
// This package implements sophisticated physics systems including:
//   - Vehicle suspension with spring-damper modeling
//   - Weight transfer during acceleration, braking, and turning
//   - Terrain deformation (tire tracks in mud, sand, snow)
//   - Realistic collision response with vehicle damage
//
// # Architecture
//
// The physics package is organized into specialized sub-systems:
//   - vehicle/: Vehicle-specific physics (suspension, weight transfer, terrain deformation, collision response)
//   - fluids/: Fluid simulation (buoyancy, flooding, swimming)
//   - destruction/: Environmental destruction (structural integrity, damage propagation, debris physics)
//
// # Integration
//
// Physics systems are initialized in cmd/client/handlers.go and use component-based data:
//
//	// Vehicle physics example
//	sys := vehicle.NewEnhancedVehicleSystem()
//	suspension := vehicle.NewSuspensionComponent(4) // 4-wheel suspension
//	weightTransfer := vehicle.NewWeightTransferComponent(mass, wheelbase, trackWidth)
//	state := sys.UpdateVehiclePhysics(suspension, weightTransfer, deform, collision, state, dt)
//
//	// Destruction physics example
//	destructionSys := destruction.NewSystem(destruction.DefaultConfig())
//	destructionSys.RegisterBuilding("building_01", 100, 200, 3, destruction.MaterialConcrete)
//	destructionSys.Update(deltaTime)
//
//	// Fluid simulation example
//	fluidSim := fluids.NewSimulator(fluids.DefaultSimulationConfig())
//	buoyancy := fluids.NewBuoyancyCalculator(gravity)
//	force := buoyancy.Calculate(entity, waterLevel)
//
// # Performance
//
// All physics calculations are designed for real-time simulation:
//   - Suspension update: <100µs per wheel (target)
//   - Weight transfer: <1ms per vehicle (target)
//   - Terrain deformation: <500µs per track update (target)
//   - Collision response: <200µs per impact (target)
//
// # Determinism
//
// Physics simulations maintain determinism for multiplayer synchronization:
//   - Fixed-step integration for consistent results
//   - Seed-based noise for terrain deformation
//   - Reproducible collision responses
//
// Phase 50.3: Enhanced Vehicle Physics (V8.0)
package physics
