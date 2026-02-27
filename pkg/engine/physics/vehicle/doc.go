// Package vehicle provides advanced vehicle physics including suspension dynamics,
// weight transfer, collision response, and terrain deformation for realistic
// vehicle simulation in the game engine.
//
// # Overview
//
// The vehicle package implements realistic vehicle physics through a component-based
// system that integrates with the game's ECS architecture. It provides accurate
// suspension simulation, weight transfer during acceleration/braking, collision damage,
// and terrain interaction including track marks and deformation.
//
// # Core Components
//
// The package provides four main ECS components for vehicle physics:
//
// ## SuspensionComponent
//
// Manages vehicle suspension using a spring-damper model with configurable parameters:
//
//	suspension := vehicle.NewSuspensionComponent(4) // 4-wheel vehicle
//	suspension.SpringStiffness = 50000.0  // N/m
//	suspension.DamperStrength = 3000.0    // N·s/m
//	suspension.RestLength = 0.3           // meters
//	suspension.MaxCompression = 0.15      // meters
//
// Each wheel has independent suspension with:
//   - Spring stiffness (controls bounce)
//   - Damper strength (controls oscillation)
//   - Rest length and maximum compression limits
//   - Ground contact detection
//
// ## WeightTransferComponent
//
// Simulates dynamic weight distribution during vehicle maneuvers:
//
//	weightTransfer := &vehicle.WeightTransferComponent{
//		FrontWeightRatio: 0.55,  // 55% weight on front axle at rest
//		WheelCount:       4,
//	}
//
// Weight transfer affects:
//   - Acceleration (weight shifts to rear wheels)
//   - Braking (weight shifts to front wheels)
//   - Cornering (weight shifts to outer wheels)
//   - Suspension compression and grip
//
// ## CollisionResponseComponent
//
// Handles realistic vehicle damage from collisions:
//
//	collision := vehicle.NewCollisionResponseComponent(1000.0)
//	result := collision.ProcessCollision(velocityX, velocityY, normalX, normalY)
//	if result.Damage > 0 {
//		// Apply damage to vehicle
//	}
//
// Features:
//   - Structural integrity tracking (0.0-1.0)
//   - Damage threshold based on impact speed
//   - Velocity reflection off surfaces
//   - Performance degradation as damage accumulates
//   - Impact force and velocity tracking
//
// ## TerrainDeformationComponent
//
// Creates realistic terrain effects from vehicle movement:
//
//	deformation := &vehicle.TerrainDeformationComponent{
//		TrackDepth:     0.05,  // meters
//		TrackWidth:     0.3,   // meters
//		MaxTrackMarks:  200,   // track history
//	}
//
// Terrain interaction:
//   - Track marks on soft terrain (mud, sand, snow)
//   - Deformation depth based on vehicle weight
//   - Track fade over time
//   - Performance impact on soft surfaces
//
// Use GetTerrainTypeFromTile() to map terrain tile types to deformation types:
//
//	tileType := world.GetTile(int(vehicleX/tileSize), int(vehicleY/tileSize))
//	deformation.TerrainType = vehicle.GetTerrainTypeFromTile(tileType)
//
// # Vehicle Physics System
//
// EnhancedVehicleSystem integrates all components for complete vehicle simulation.
// It implements the engine.System interface and can be registered with the ECS World
// for automatic per-frame updates:
//
//	system := vehicle.NewEnhancedVehicleSystem()
//	world.AddSystem(system)
//
// The system automatically processes all entities that have at least one vehicle
// physics component (suspension, weight_transfer, terrain_deformation, or
// collision_response), extracting position and velocity data to run physics
// calculations.
//
// For manual control, the system can also be used directly:
//
//	// Manual per-frame update
//	state := vehicle.VehicleState{
//		PositionX:  100.0,
//		PositionY:  50.0,
//		VelocityX:  10.0,
//		VelocityY:  0.0,
//		IsGrounded: true,
//	}
//
//	newState := system.UpdateVehiclePhysics(
//		suspension,
//		weightTransfer,
//		deformation,
//		collision,
//		state,
//		deltaTime,
//	)
//
// # Terrain Types
//
// The package supports different terrain characteristics:
//
//	const (
//		TerrainHard  vehicle.TerrainType = 0  // Stone, concrete - no deformation
//		TerrainFirm  vehicle.TerrainType = 1  // Packed dirt - minimal deformation
//		TerrainSoft  vehicle.TerrainType = 2  // Mud, sand - significant deformation
//		TerrainSnow  vehicle.TerrainType = 3  // Snow - deep tracks
//		TerrainWater vehicle.TerrainType = 4  // Water - no permanent tracks
//	)
//
// Each terrain type affects:
//   - Grip/traction coefficient
//   - Deformation depth
//   - Track mark visibility and persistence
//   - Vehicle performance and handling
//
// # Usage Example
//
// Complete vehicle setup in an ECS entity:
//
//	// Create vehicle entity
//	vehicle := world.CreateEntity()
//
//	// Add position and physics components
//	vehicle.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
//	vehicle.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
//
//	// Add vehicle-specific physics
//	suspension := vehicle.NewSuspensionComponent(4)
//	suspension.SpringStiffness = 50000.0
//	suspension.DamperStrength = 3000.0
//	vehicle.AddComponent(suspension)
//
//	vehicle.AddComponent(&vehicle.WeightTransferComponent{
//		FrontWeightRatio: 0.55,
//		WheelCount:       4,
//	})
//
//	vehicle.AddComponent(vehicle.NewCollisionResponseComponent(1000.0))
//
//	vehicle.AddComponent(&vehicle.TerrainDeformationComponent{
//		TrackDepth:    0.05,
//		TrackWidth:    0.3,
//		MaxTrackMarks: 200,
//	})
//
// # Performance Considerations
//
// The vehicle physics system is optimized for real-time simulation:
//
//   - Suspension calculations are O(n) where n is wheel count
//   - Track mark management uses fixed-size circular buffers
//   - Weight transfer uses cached coefficients
//   - Collision detection integrates with spatial partitioning
//
// Typical performance for 4-wheel vehicle:
//   - Suspension update: ~1-2 microseconds
//   - Weight transfer: <1 microsecond
//   - Collision response: ~1 microsecond
//   - Terrain deformation: ~2-3 microseconds
//
// # Testing
//
// The package includes comprehensive unit tests with >80% coverage:
//
//	func TestSuspensionCompression(t *testing.T) {
//		suspension := vehicle.NewSuspensionComponent(1)
//		suspension.Wheels[0].TerrainHeight = -0.1  // Wheel compressed
//		offset := suspension.Update(0.016)         // 60 FPS delta
//		// Verify spring physics
//	}
//
// Benchmarks are provided for performance validation:
//
//	BenchmarkSuspensionUpdate
//	BenchmarkWeightTransfer
//	BenchmarkCollisionResponse
//	BenchmarkTerrainDeformation
//
// # Integration
//
// This package integrates with other engine systems:
//
//   - engine.CollisionSystem: Collision detection triggers ProcessCollision()
//   - engine.RenderSystem: Visual representation of suspension compression
//   - procgen/terrain: Terrain type affects deformation behavior
//   - procgen/vehicle: Vehicle stats determine physics parameters
//
// # References
//
// For more information:
//   - Physics documentation: docs/ARCHITECTURE.md (Physics Systems section)
//   - ECS integration: docs/TECHNICAL_SPEC.md
//   - Vehicle generation: pkg/procgen/vehicle/doc.go
package vehicle
