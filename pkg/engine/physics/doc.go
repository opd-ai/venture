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
//   - vehicle/: Vehicle-specific physics (suspension, weight transfer)
//   - terrain/: Terrain interaction and deformation
//   - collision/: Advanced collision response
//
// # Integration
//
// Physics components integrate with the ECS (Entity-Component-System) pattern:
//
//	// Create a vehicle with physics
//	entity := world.CreateEntity()
//	entity.AddComponent(engine.NewVehicleComponent(engine.VehicleCart))
//	entity.AddComponent(physics.NewSuspensionComponent(4)) // 4-wheel suspension
//	entity.AddComponent(physics.NewWeightTransferComponent())
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
