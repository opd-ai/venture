package vehicle

import (
	procvehicle "github.com/opd-ai/venture/pkg/procgen/vehicle"
)

// CreatePhysicsComponents creates all four vehicle physics components from a procedurally
// generated vehicle. This helper function bridges pkg/procgen/vehicle and pkg/engine/physics/vehicle.
//
// Component configuration is based on vehicle type:
//   - Mount: 4 wheels (legs), soft suspension, low damage threshold (organic)
//   - Cart: 2-4 wheels, medium suspension, medium damage threshold
//   - Boat: 0 wheels (water displacement instead), soft suspension, low damage threshold
//   - Glider: 0 wheels (aerodynamic lift), very soft suspension, very low damage threshold
//   - Mech: 2-4 wheels (feet), very stiff suspension, high damage threshold (armored)
//
// Suspension travel, spring stiffness, and damper strength are scaled by vehicle stats:
// handling affects stiffness, durability affects damage threshold.
//
// Returns: suspension, weightTransfer, collision, terrainDeformation components (or nil if not applicable).
func CreatePhysicsComponents(v *procvehicle.Vehicle) (
	*SuspensionComponent,
	*WeightTransferComponent,
	*CollisionResponseComponent,
	*TerrainDeformationComponent,
) {
	// Determine wheel count based on vehicle type
	wheelCount := 4 // default
	switch v.VehicleType {
	case procvehicle.TypeMount:
		wheelCount = 4 // four legs
	case procvehicle.TypeCart:
		wheelCount = 4 // four wheels (or two if light cart)
	case procvehicle.TypeBoat:
		wheelCount = 0 // no wheels (uses fluid simulation instead)
	case procvehicle.TypeGlider:
		wheelCount = 0 // no wheels (aerodynamic forces instead)
	case procvehicle.TypeMech:
		wheelCount = 2 // bipedal mech (or 4 for quadruped variant)
	}

	// Calculate mass from cargo weight and base vehicle mass
	// Base mass varies by type: mount ~500kg, cart ~200kg, boat ~800kg, glider ~50kg, mech ~1500kg
	baseMass := 500.0
	switch v.VehicleType {
	case procvehicle.TypeMount:
		baseMass = 500.0
	case procvehicle.TypeCart:
		baseMass = 200.0
	case procvehicle.TypeBoat:
		baseMass = 800.0
	case procvehicle.TypeGlider:
		baseMass = 50.0
	case procvehicle.TypeMech:
		baseMass = 1500.0
	}
	totalMass := baseMass + v.CargoWeight

	// Suspension parameters scaled by handling stat
	// Higher handling = stiffer suspension for better control
	handlingScale := v.Handling / 100.0 // normalize assuming Handling is in [0, 200] range
	springStiffness := 15000.0 * handlingScale
	damperStrength := 2000.0 * handlingScale
	suspensionTravel := 10.0 / handlingScale // softer = more travel

	// Create suspension component (skip if no wheels)
	var suspension *SuspensionComponent
	if wheelCount > 0 {
		suspension = NewSuspensionComponent(wheelCount)
		suspension.TotalMass = totalMass
		suspension.SuspensionTravel = suspensionTravel

		// Configure each wheel with spring-damper parameters
		for i := range suspension.Wheels {
			suspension.Wheels[i].SpringStiffness = springStiffness
			suspension.Wheels[i].DamperStrength = damperStrength
		}
	}

	// Create weight transfer component (always needed for acceleration/braking/turning dynamics)
	weightTransfer := &WeightTransferComponent{
		AccelerationX:         0.0,
		AccelerationY:         0.0,
		AngularAccel:          0.0,
		CenterOfMassHeight:    15.0, // pixels above axles
		Wheelbase:             32.0, // pixels front-to-rear
		TrackWidth:            16.0, // pixels left-to-right
		FrontLeftWeight:       0.25, // assume 25% per wheel at rest
		FrontRightWeight:      0.25,
		RearLeftWeight:        0.25,
		RearRightWeight:       0.25,
		StaticFrontWeight:     0.5, // 50% on front axle
		PrevVelocityX:         0.0,
		PrevVelocityY:         0.0,
		PrevAngularVel:        0.0,
		LastTransferMagnitude: 0.0,
	}

	// Create collision response component
	// Damage threshold scaled by MaxDurability
	damageThreshold := 5.0 // base threshold in m/s
	if v.MaxDurability > 0 {
		// Higher durability = higher damage threshold
		damageThreshold = 5.0 + (v.MaxDurability / 100.0 * 10.0) // range [5, 15] m/s
	}

	collision := &CollisionResponseComponent{
		DamageThreshold:     damageThreshold,
		MassForCalculation:  totalMass,
		StructuralIntegrity: 1.0, // full integrity
		Restitution:         0.3, // default bounce
		LastImpactVelocity:  0.0,
		LastImpactForce:     0.0,
		LastImpactAngle:     0.0,
		TotalImpactDamage:   0.0,
		CollisionCount:      0,
	}

	// Create terrain deformation component (skip for boats and gliders)
	var terrainDeformation *TerrainDeformationComponent
	if v.VehicleType != procvehicle.TypeBoat && v.VehicleType != procvehicle.TypeGlider {
		terrainDeformation = NewTerrainDeformationComponent(0) // seed will be set by calling code
		terrainDeformation.MaxTracks = 200
		terrainDeformation.MinTrackSpacing = 5.0

		// Heavy vehicles (mechs, carts with cargo) create deeper tracks
		if v.VehicleType == procvehicle.TypeMech || totalMass > 1000.0 {
			// Increase deformation depth for heavy vehicles
			terrainDeformation.DeformationDepth[TerrainFirm] = 0.4
			terrainDeformation.DeformationDepth[TerrainSoft] = 1.2
			terrainDeformation.DeformationDepth[TerrainSnow] = 1.5
		}
	}

	return suspension, weightTransfer, collision, terrainDeformation
}
