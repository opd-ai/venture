// Package destruction implements building structural integrity simulation,
// damage propagation, and environmental physics for destructible buildings.
//
// # Overview
//
// The destruction package provides systems for realistic building damage,
// structural collapse, and physics-based debris generation. Buildings are
// tracked with structural integrity points distributed across load-bearing
// supports (walls, columns, beams, foundations). Damage propagates over time
// from damaged areas, and buildings collapse when structural integrity falls
// below configurable thresholds.
//
// # Key Features
//
// - Structural integrity tracking with support points (walls, columns, beams, foundations)
// - Damage propagation from impact areas with configurable spread rate
// - Realistic collapse mechanics based on load-bearing structure health
// - Physics-based debris generation with material properties
// - Falling object simulation with gravity, bouncing, friction, and rotation
// - Material-specific properties (wood, stone, metal, glass, concrete, brick)
// - Performance-optimized fixed timestep updates (30 Hz default, configurable)
//
// # Architecture
//
// The package consists of:
//   - System: Main manager coordinating integrity checks, damage propagation, and physics
//   - StructuralIntegrity: Per-building health tracking with support points and damage areas
//   - DebrisParticle: Lightweight particles for visual effects
//   - FallingObject: Full physics simulation for larger objects
//   - Material system: Physical properties (density, bounciness, friction, durability)
//
// # Usage Example
//
//	// Create destruction system
//	config := destruction.DefaultConfig()
//	config.MaxDebrisParticles = 1000
//	config.CollapseThreshold = 0.2 // Collapse at 20% health
//	sys := destruction.NewSystem(config)
//
//	// Register a building
//	if err := sys.RegisterBuilding("house1", 16, 16, 2, destruction.MaterialWood); err != nil {
//	    logrus.WithFields(logrus.Fields{
//	        "building_id": "house1",
//	        "error":       err,
//	    }).Fatal("failed to register building")
//	}
//
//	// Apply damage (explosion at x=8, y=8, radius=5.0, amount=0.6)
//	if err := sys.ApplyDamage("house1", 8, 8, 0, 0.6, 5.0); err != nil {
//	    logrus.WithFields(logrus.Fields{
//	        "building_id": "house1",
//	        "error":       err,
//	    }).Warn("failed to apply damage")
//	}
//
//	// Update system each frame
//	for {
//	    deltaTime := 0.016 // 60 FPS
//	    sys.Update(deltaTime)
//
//	    // Check integrity
//	    integrity, _ := sys.GetIntegrity("house1")
//	    if integrity.State == destruction.IntegrityCollapsed {
//	        logrus.WithFields(logrus.Fields{
//	            "building_id": "house1",
//	            "state":       "collapsed",
//	        }).Info("building collapsed")
//	    }
//
//	    // Render debris
//	    for _, debris := range sys.GetDebris() {
//	        renderDebris(debris.X, debris.Y, debris.Size, debris.Angle)
//	    }
//
//	    // Render falling objects
//	    for _, obj := range sys.GetFallingObjects() {
//	        if !obj.IsGrounded() {
//	            renderFallingObject(obj.X, obj.Y, obj.Z, obj.Angle)
//	        }
//	    }
//	}
//
// # Structural Integrity
//
// Buildings are tracked with structural health distributed across support points.
// Each support has:
//   - Position (X, Y, Floor)
//   - Type (Wall, Column, Beam, Foundation)
//   - Health (0.0-1.0)
//   - Load capacity and current load
//
// Damage is applied to supports within a radius of impact, with falloff based on distance.
// Weakened supports reduce overall building health and increase collapse risk.
//
// # Damage Propagation
//
// Damage spreads over time from damaged areas at a configurable rate (default 10% per second).
// Each damaged area has:
//   - Position and radius
//   - Severity (0.0-1.0)
//   - Propagation age (for time-based spreading)
//
// Damage propagates to nearby supports outside the initial damage radius,
// simulating structural stress and weakness spreading through the building.
//
// # Collapse Mechanics
//
// Buildings collapse when:
//   - Overall health falls below collapse threshold (default 15%)
//   - Critical percentage of load-bearing supports are damaged (>30% at <30% health)
//
// Collapse triggers:
//   - State change to IntegrityCollapsed
//   - Debris generation (3 particles per support point)
//   - Optional falling object spawning for large structural elements
//
// # Material Properties
//
// Each material has distinct physical properties:
//
//   - Wood: Light, medium bounce, flammable, medium durability
//   - Stone: Heavy, low bounce, non-flammable, very durable
//   - Metal: Heavy, medium bounce, non-flammable, very durable
//   - Glass: Light, low bounce, non-flammable, fragile
//   - Concrete: Very heavy, very low bounce, non-flammable, durable
//   - Brick: Medium-heavy, low bounce, non-flammable, durable
//
// Properties affect:
//   - Debris/object mass and gravity response
//   - Bounce height and frequency
//   - Sliding friction
//   - Damage resistance
//
// # Physics Simulation
//
// The system simulates realistic physics for debris and falling objects:
//
//   - Gravity: Constant downward acceleration (default 980 pixels/sec²)
//   - Air resistance: Velocity damping (default 5%)
//   - Bouncing: Material-dependent restitution coefficient
//   - Friction: Ground and air friction for sliding and tumbling
//   - Rotation: Angular velocity for tumbling motion
//
// Falling objects settle on the ground after a maximum number of bounces
// (default 3) when vertical velocity falls below a threshold.
//
// # Performance Optimization
//
// The system uses several optimizations:
//
//   - Fixed timestep updates (30 Hz default, independent of game FPS)
//   - Time accumulator for smooth interpolation
//   - Active particle culling (lifetime-based removal)
//   - Configurable limits (max debris, max falling objects)
//   - Efficient distance calculations for damage checks
//   - Early-exit for settled objects (IsGrounded check)
//
// Target performance:
//   - <10ms per building integrity check
//   - <50ms for large structure damage propagation
//   - Supports 100+ buildings, 500 debris particles, 100 falling objects at 60 FPS
//
// # Integration with ECS
//
// The destruction system can be integrated with an Entity-Component-System:
//
//	type DestructibleComponent struct {
//	    BuildingID   string
//	    Registered   bool
//	}
//
//	type DestructionSystem struct {
//	    destruction *destruction.System
//	}
//
//	func (s *DestructionSystem) Update(deltaTime float64) {
//	    s.destruction.Update(deltaTime)
//
//	    // Sync entity positions with destruction state
//	    for _, entity := range world.GetEntitiesWith("destructible") {
//	        comp := entity.GetComponent("destructible").(*DestructibleComponent)
//	        integrity, _ := s.destruction.GetIntegrity(comp.BuildingID)
//
//	        if integrity.State == destruction.IntegrityCollapsed {
//	            entity.Destroy()
//	        }
//	    }
//	}
//
// # Configuration
//
// Default configuration provides balanced settings for typical gameplay:
//
//   - Integrity checks: Enabled
//   - Damage propagation rate: 10% per second
//   - Collapse threshold: 15% health
//   - Max debris particles: 500
//   - Max falling objects: 100
//   - Debris lifetime: 10 seconds
//   - Gravity: 980 pixels/sec² (Earth-like)
//   - Update frequency: 30 Hz
//
// These can be tuned per-game or per-building type for different effects
// (e.g., fragile buildings with lower collapse threshold, slow propagation
// for dramatic sequences, high debris count for spectacular effects).
//
// # Testing
//
// The package includes comprehensive tests covering:
//   - String methods for all enum types
//   - Material property validation
//   - Structural integrity tracking
//   - Damage application and propagation
//   - Collapse mechanics
//   - Physics simulation (gravity, bouncing, friction)
//   - Debris lifecycle
//   - Performance benchmarks
//
// Run tests with:
//
//	go test -v ./pkg/engine/physics/destruction/
//	go test -bench=. ./pkg/engine/physics/destruction/
//
// # Future Enhancements
//
// Potential improvements for future phases:
//   - Fire propagation for flammable materials
//   - Water damage and flooding
//   - Chain reaction collapses (building A falls onto building B)
//   - Structural reinforcement (player can strengthen buildings)
//   - Repair mechanics (restore integrity over time)
//   - Visual damage states (cracks, holes, missing walls)
//   - Sound effects for creaking, collapse, impacts
package destruction
