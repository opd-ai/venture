// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

import (
	"math"

	log "github.com/sirupsen/logrus"
)

// EnhancedVehicleSystem integrates suspension, weight transfer, terrain deformation, and collision response.
// All physics logic lives in this system; components are pure data structures.
// Implements the engine.System interface for automatic ECS integration.
type EnhancedVehicleSystem struct {
	// System configuration
	enabled bool
}

// Entity interface defines the minimal interface required from engine.Entity.
// This allows the system to work without importing the full engine package.
type Entity interface {
	HasComponent(componentType string) bool
	GetComponent(componentType string) interface{}
}

// Update implements the System interface, processing all entities with vehicle physics components.
// Entities require PositionComponent, VelocityComponent, and at least one vehicle physics component
// (SuspensionComponent, WeightTransferComponent, TerrainDeformationComponent, or CollisionResponseComponent).
func (evs *EnhancedVehicleSystem) Update(entities []Entity, deltaTime float64) {
	if !evs.enabled {
		return
	}

	for _, entity := range entities {
		// Only process entities that have at least one vehicle physics component
		if !entity.HasComponent("suspension") &&
			!entity.HasComponent("weight_transfer") &&
			!entity.HasComponent("terrain_deformation") &&
			!entity.HasComponent("collision_response") {
			continue
		}

		evs.updateEntity(entity, deltaTime)
	}
}

// updateEntity processes vehicle physics for a single entity.
func (evs *EnhancedVehicleSystem) updateEntity(entity Entity, deltaTime float64) {
	// Extract components (may be nil)
	var suspension *SuspensionComponent
	var weightTransfer *WeightTransferComponent
	var deformation *TerrainDeformationComponent
	var collision *CollisionResponseComponent

	if comp := entity.GetComponent("suspension"); comp != nil {
		if s, ok := comp.(*SuspensionComponent); ok {
			suspension = s
		}
	}

	if comp := entity.GetComponent("weight_transfer"); comp != nil {
		if w, ok := comp.(*WeightTransferComponent); ok {
			weightTransfer = w
		}
	}

	if comp := entity.GetComponent("terrain_deformation"); comp != nil {
		if t, ok := comp.(*TerrainDeformationComponent); ok {
			deformation = t
		}
	}

	if comp := entity.GetComponent("collision_response"); comp != nil {
		if c, ok := comp.(*CollisionResponseComponent); ok {
			collision = c
		}
	}

	// Extract required position and velocity data
	state := evs.extractVehicleState(entity)

	// Run physics update
	newState := evs.UpdateVehiclePhysics(suspension, weightTransfer, deformation, collision, state, deltaTime)

	// Write results back to entity components
	evs.applyVehicleState(entity, newState)
}

// extractVehicleState builds VehicleState from entity components.
func (evs *EnhancedVehicleSystem) extractVehicleState(entity Entity) VehicleState {
	state := VehicleState{
		TerrainHeight: make([]float64, 4),
		TerrainTypes:  make([]TerrainType, 4),
	}

	// Position
	if comp := entity.GetComponent("position"); comp != nil {
		type positionGetter interface {
			GetX() float64
			GetY() float64
		}
		if pos, ok := comp.(positionGetter); ok {
			state.PositionX = pos.GetX()
			state.PositionY = pos.GetY()
		}
	}

	// Velocity
	if comp := entity.GetComponent("velocity"); comp != nil {
		type velocityGetter interface {
			GetVelX() float64
			GetVelY() float64
		}
		if vel, ok := comp.(velocityGetter); ok {
			state.VelocityX = vel.GetVelX()
			state.VelocityY = vel.GetVelY()
			state.Speed = math.Sqrt(vel.GetVelX()*vel.GetVelX() + vel.GetVelY()*vel.GetVelY())
		}
	}

	// Rotation (if available)
	if comp := entity.GetComponent("rotation"); comp != nil {
		type rotationGetter interface {
			GetAngle() float64
		}
		if rot, ok := comp.(rotationGetter); ok {
			state.Rotation = rot.GetAngle()
		}
	}

	// Note: TerrainHeight and TerrainTypes would need to be provided by a terrain query system
	// For now, default to firm terrain at ground level
	for i := range state.TerrainTypes {
		state.TerrainTypes[i] = TerrainFirm
	}

	return state
}

// applyVehicleState writes updated state back to entity components.
func (evs *EnhancedVehicleSystem) applyVehicleState(entity Entity, state VehicleState) {
	// Update velocity if changed
	if comp := entity.GetComponent("velocity"); comp != nil {
		type velocitySetter interface {
			SetVelocity(vx, vy float64)
		}
		if vel, ok := comp.(velocitySetter); ok {
			vel.SetVelocity(state.VelocityX, state.VelocityY)
		}
	}
}

// NewEnhancedVehicleSystem creates a new enhanced vehicle physics system.
func NewEnhancedVehicleSystem() *EnhancedVehicleSystem {
	log.WithFields(log.Fields{
		"system_name": "enhanced_vehicle",
	}).Debug("Creating enhanced vehicle physics system")
	return &EnhancedVehicleSystem{
		enabled: true,
	}
}

// UpdateVehiclePhysics performs a complete physics update for a vehicle.
// This integrates suspension, weight transfer, and collision response.
func (evs *EnhancedVehicleSystem) UpdateVehiclePhysics(
	suspension *SuspensionComponent,
	weightTransfer *WeightTransferComponent,
	deformation *TerrainDeformationComponent,
	collision *CollisionResponseComponent,
	state VehicleState,
	deltaTime float64,
) VehicleState {
	if !evs.enabled {
		return state
	}

	evs.updateWeightTransfer(suspension, weightTransfer, state, deltaTime)
	state = evs.updateSuspension(suspension, weightTransfer, state, deltaTime)
	evs.updateTireTracks(suspension, deformation, state, deltaTime)
	state = evs.applyCollisionDamage(collision, state)

	return state
}

// updateWeightTransfer updates weight transfer and applies it to suspension.
func (evs *EnhancedVehicleSystem) updateWeightTransfer(
	suspension *SuspensionComponent,
	weightTransfer *WeightTransferComponent,
	state VehicleState,
	deltaTime float64,
) {
	if weightTransfer == nil {
		return
	}

	evs.UpdateWeightDistribution(weightTransfer, state.VelocityX, state.VelocityY, state.AngularVel, deltaTime)

	if suspension != nil {
		evs.applyWeightToSuspension(suspension, weightTransfer)
	}
}

// applyWeightToSuspension distributes mass to wheels based on weight transfer.
func (evs *EnhancedVehicleSystem) applyWeightToSuspension(suspension *SuspensionComponent, weightTransfer *WeightTransferComponent) {
	weights := GetWheelWeights(weightTransfer)
	totalMass := suspension.TotalMass

	for i := 0; i < len(suspension.Wheels) && i < 4; i++ {
		wheelMass := totalMass * weights[i]
		wheelLoad := wheelMass * 9.81
		SetWheelLoad(suspension, i, wheelLoad)
	}
}

// updateSuspension updates suspension physics and grounding state.
func (evs *EnhancedVehicleSystem) updateSuspension(
	suspension *SuspensionComponent,
	_ *WeightTransferComponent,
	state VehicleState,
	deltaTime float64,
) VehicleState {
	if suspension != nil && len(state.TerrainHeight) == len(suspension.Wheels) {
		_ = evs.UpdateSuspensionPhysics(suspension, deltaTime, state.TerrainHeight)
		groundedCount := GetGroundedWheelCount(suspension)
		state.IsGrounded = groundedCount >= 2
	}
	return state
}

// updateTireTracks adds tire tracks for moving vehicles on soft terrain.
func (evs *EnhancedVehicleSystem) updateTireTracks(
	suspension *SuspensionComponent,
	deformation *TerrainDeformationComponent,
	state VehicleState,
	deltaTime float64,
) {
	if deformation == nil || state.Speed <= 10.0 {
		return
	}

	if suspension != nil {
		for i := range suspension.Wheels {
			if IsWheelGrounded(suspension, i) {
				evs.addWheelTrack(suspension, deformation, state, i)
			}
		}
	}

	evs.UpdateTerrainTracks(deformation, deltaTime)
}

// addWheelTrack adds a single wheel track to terrain deformation.
func (evs *EnhancedVehicleSystem) addWheelTrack(
	suspension *SuspensionComponent,
	deformation *TerrainDeformationComponent,
	state VehicleState,
	wheelIndex int,
) {
	wheel := &suspension.Wheels[wheelIndex]

	wheelWorldX := state.PositionX + wheel.LocalX*math.Cos(state.Rotation) - wheel.LocalY*math.Sin(state.Rotation)
	wheelWorldY := state.PositionY + wheel.LocalX*math.Sin(state.Rotation) + wheel.LocalY*math.Cos(state.Rotation)

	terrainType := TerrainFirm
	if wheelIndex < len(state.TerrainTypes) {
		terrainType = state.TerrainTypes[wheelIndex]
	}

	wheelLoad := GetWheelLoad(suspension, wheelIndex)
	evs.AddTerrainTrack(deformation, wheelWorldX, wheelWorldY, state.Rotation, wheelLoad, terrainType)
}

// applyCollisionDamage applies damage multiplier to vehicle performance.
func (evs *EnhancedVehicleSystem) applyCollisionDamage(collision *CollisionResponseComponent, state VehicleState) VehicleState {
	if collision != nil && !IsDestroyed(collision) {
		damageMultiplier := GetDamageMultiplier(collision)

		if damageMultiplier < 1.0 {
			state.Speed *= damageMultiplier
			state.VelocityX *= damageMultiplier
			state.VelocityY *= damageMultiplier
		}
	}

	return state
}

// ProcessVehicleCollision handles a collision event for a vehicle.
// Returns updated velocity after collision response.
func (evs *EnhancedVehicleSystem) ProcessVehicleCollision(
	collision *CollisionResponseComponent,
	state VehicleState,
	normalX, normalY float64,
) (newVelX, newVelY, damage float64) {
	if collision == nil {
		return state.VelocityX, state.VelocityY, 0.0
	}

	result := evs.ProcessCollisionResponse(collision, state.VelocityX, state.VelocityY, normalX, normalY)

	return result.BounceVelocityX, result.BounceVelocityY, result.DamageDealt
}

// Enable enables the enhanced vehicle physics system.
func (evs *EnhancedVehicleSystem) Enable() {
	evs.enabled = true
}

// Disable disables the enhanced vehicle physics system.
func (evs *EnhancedVehicleSystem) Disable() {
	evs.enabled = false
}

// IsEnabled returns whether the system is enabled.
func (evs *EnhancedVehicleSystem) IsEnabled() bool {
	return evs.enabled
}

// --- Suspension physics logic (moved from SuspensionComponent) ---

// UpdateSuspensionPhysics simulates suspension physics for one frame.
// Returns the average vertical offset to apply to vehicle rendering.
func (evs *EnhancedVehicleSystem) UpdateSuspensionPhysics(s *SuspensionComponent, deltaTime float64, terrainHeights []float64) float64 {
	if len(terrainHeights) != len(s.Wheels) {
		return 0.0
	}

	s.LastUpdateTime += deltaTime

	totalVerticalOffset := 0.0
	groundedWheels := 0

	for i := range s.Wheels {
		wheel := &s.Wheels[i]
		_ = terrainHeights[i]

		compressionDistance := wheel.Compression * s.SuspensionTravel
		springForce := -wheel.SpringStiffness * compressionDistance
		damperForce := -wheel.DamperStrength * wheel.CompressionRate
		totalForce := springForce + damperForce

		wheelWeight := (s.TotalMass * 9.81) / float64(len(s.Wheels))
		netForce := wheelWeight + totalForce

		wheelMass := s.TotalMass / (10.0 * float64(len(s.Wheels)))
		compressionAccel := netForce / wheelMass

		wheel.CompressionRate += compressionAccel * deltaTime
		wheel.Compression += wheel.CompressionRate * deltaTime

		if wheel.Compression < 0.0 {
			wheel.Compression = 0.0
			wheel.CompressionRate = 0.0
			wheel.IsGrounded = false
		} else if wheel.Compression > 1.0 {
			wheel.Compression = 1.0
			wheel.CompressionRate = 0.0
			wheel.IsGrounded = true
		} else {
			wheel.IsGrounded = true
		}

		wheel.Load = math.Abs(springForce + damperForce)

		if wheel.IsGrounded {
			verticalOffset := (1.0 - wheel.Compression) * s.SuspensionTravel
			totalVerticalOffset += verticalOffset
			groundedWheels++
		}
	}

	if groundedWheels > 0 {
		return totalVerticalOffset / float64(groundedWheels)
	}
	return 0.0
}

// --- Collision response logic (moved from CollisionResponseComponent) ---

// ProcessCollisionResponse calculates damage and response from a collision.
// velocityX, velocityY: vehicle velocity before impact
// normalX, normalY: surface normal at collision point (unit vector)
// Returns: impact result with damage and velocity changes
func (evs *EnhancedVehicleSystem) ProcessCollisionResponse(c *CollisionResponseComponent, velocityX, velocityY, normalX, normalY float64) ImpactResult {
	impactSpeed := math.Sqrt(velocityX*velocityX + velocityY*velocityY)
	c.LastImpactVelocity = impactSpeed
	c.CollisionCount++

	if impactSpeed < c.DamageThreshold {
		reflectedVel := reflectVelocity(velocityX, velocityY, normalX, normalY)
		return ImpactResult{
			DamageDealt:       0.0,
			VelocityReduction: impactSpeed * (1.0 - c.Restitution),
			BounceVelocityX:   reflectedVel[0],
			BounceVelocityY:   reflectedVel[1],
			IntegrityLoss:     0.0,
		}
	}

	collisionTime := 0.1
	deltaV := impactSpeed
	force := (c.MassForCalculation * deltaV) / collisionTime
	c.LastImpactForce = force

	velocityMag := math.Sqrt(velocityX*velocityX + velocityY*velocityY)
	if velocityMag == 0 {
		velocityMag = 1.0
	}

	dotProduct := (velocityX*normalX + velocityY*normalY) / velocityMag
	dotProduct = math.Abs(dotProduct)
	c.LastImpactAngle = math.Acos(math.Max(-1.0, math.Min(1.0, dotProduct)))

	speedFactor := (impactSpeed - c.DamageThreshold) / 100.0
	angleFactor := dotProduct
	damageCoeff := 0.5

	damage := speedFactor * speedFactor * angleFactor * damageCoeff
	damage = math.Max(0.0, math.Min(damage, 100.0))

	c.TotalImpactDamage += damage

	integrityLoss := damage * 0.01
	c.StructuralIntegrity -= integrityLoss
	if c.StructuralIntegrity < 0.0 {
		c.StructuralIntegrity = 0.0
	}

	reflectedVel := reflectVelocity(velocityX, velocityY, normalX, normalY)

	effectiveRestitution := c.Restitution * c.StructuralIntegrity
	bounceVelX := reflectedVel[0] * effectiveRestitution
	bounceVelY := reflectedVel[1] * effectiveRestitution

	velocityReduction := impactSpeed - math.Sqrt(bounceVelX*bounceVelX+bounceVelY*bounceVelY)

	return ImpactResult{
		DamageDealt:       damage,
		VelocityReduction: velocityReduction,
		BounceVelocityX:   bounceVelX,
		BounceVelocityY:   bounceVelY,
		IntegrityLoss:     integrityLoss,
	}
}

// reflectVelocity calculates the reflected velocity vector given a surface normal.
// Formula: v' = v - 2(v·n)n
func reflectVelocity(vx, vy, nx, ny float64) [2]float64 {
	nMag := math.Sqrt(nx*nx + ny*ny)
	if nMag > 0 {
		nx /= nMag
		ny /= nMag
	}

	dotProduct := vx*nx + vy*ny

	reflectX := vx - 2.0*dotProduct*nx
	reflectY := vy - 2.0*dotProduct*ny

	return [2]float64{reflectX, reflectY}
}

// --- Terrain deformation logic (moved from TerrainDeformationComponent) ---

// AddTerrainTrack creates a new track mark at the specified position.
// wheelLoad: force on wheel (affects depth), vehicleAngle: direction of travel
func (evs *EnhancedVehicleSystem) AddTerrainTrack(t *TerrainDeformationComponent, x, y, vehicleAngle, wheelLoad float64, terrainType TerrainType) {
	dx := x - t.LastTrackX
	dy := y - t.LastTrackY
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < t.MinTrackSpacing {
		return
	}

	baseDepth := t.DeformationDepth[terrainType]
	fadeTime := t.FadeTime[terrainType]

	if baseDepth <= 0 || fadeTime <= 0 {
		return
	}

	loadFactor := math.Min(wheelLoad/5000.0, 2.0)
	depth := baseDepth * loadFactor
	if depth > 1.0 {
		depth = 1.0
	}

	depthNoise := t.rng.Float64()*0.1 - 0.05
	depth = math.Max(0.0, math.Min(1.0, depth+depthNoise))

	width := 10.0 + t.rng.Float64()*4.0 - 2.0

	track := TrackMark{
		X:           x,
		Y:           y,
		Angle:       vehicleAngle,
		Depth:       depth,
		Width:       width,
		Age:         0.0,
		TerrainType: terrainType,
		FadeTime:    fadeTime,
	}

	t.Tracks = append(t.Tracks, track)

	t.LastTrackX = x
	t.LastTrackY = y

	if len(t.Tracks) > t.MaxTracks {
		removeCount := t.MaxTracks / 10
		if removeCount < 1 {
			removeCount = 1
		}
		t.Tracks = t.Tracks[removeCount:]
	}
}

// UpdateTerrainTracks ages existing tracks and removes faded ones.
func (evs *EnhancedVehicleSystem) UpdateTerrainTracks(t *TerrainDeformationComponent, deltaTime float64) {
	for i := range t.Tracks {
		t.Tracks[i].Age += deltaTime
	}

	for i := len(t.Tracks) - 1; i >= 0; i-- {
		track := &t.Tracks[i]
		if track.Age >= track.FadeTime {
			t.Tracks = append(t.Tracks[:i], t.Tracks[i+1:]...)
		}
	}
}

// --- Weight transfer logic (moved from WeightTransferComponent) ---

// UpdateWeightDistribution calculates weight transfer based on current vehicle dynamics.
// velocityX, velocityY: current velocity (pixels/s)
// angularVel: current angular velocity (radians/s)
// deltaTime: time step (seconds)
func (evs *EnhancedVehicleSystem) UpdateWeightDistribution(w *WeightTransferComponent, velocityX, velocityY, angularVel, deltaTime float64) {
	if deltaTime <= 0 {
		return
	}

	w.AccelerationX = (velocityX - w.PrevVelocityX) / deltaTime
	w.AccelerationY = (velocityY - w.PrevVelocityY) / deltaTime
	w.AngularAccel = (angularVel - w.PrevAngularVel) / deltaTime

	w.PrevVelocityX = velocityX
	w.PrevVelocityY = velocityY
	w.PrevAngularVel = angularVel

	longitudinalTransfer := calculateLongitudinalTransfer(w)
	lateralTransfer := calculateLateralTransfer(w)

	applyWeightTransfers(w, longitudinalTransfer, lateralTransfer)

	w.LastTransferMagnitude = math.Sqrt(longitudinalTransfer*longitudinalTransfer + lateralTransfer*lateralTransfer)
}

// calculateLongitudinalTransfer computes front-rear weight shift.
func calculateLongitudinalTransfer(w *WeightTransferComponent) float64 {
	if w.Wheelbase <= 0 {
		return 0.0
	}

	accelMagnitude := math.Sqrt(w.AccelerationX*w.AccelerationX + w.AccelerationY*w.AccelerationY)

	isAccelerating := w.AccelerationX > 0 || w.AccelerationY > 0

	transfer := (accelMagnitude * w.CenterOfMassHeight) / w.Wheelbase

	if isAccelerating {
		transfer = -transfer
	}

	if transfer > 0.3 {
		transfer = 0.3
	} else if transfer < -0.3 {
		transfer = -0.3
	}

	return transfer
}

// calculateLateralTransfer computes left-right weight shift during turning.
func calculateLateralTransfer(w *WeightTransferComponent) float64 {
	if w.TrackWidth <= 0 {
		return 0.0
	}

	lateralAccel := math.Abs(w.AngularAccel) * w.TrackWidth

	transfer := (lateralAccel * w.CenterOfMassHeight) / w.TrackWidth

	if transfer > 0.25 {
		transfer = 0.25
	}

	return transfer
}

// applyWeightTransfers distributes weight to individual wheels based on transfers.
func applyWeightTransfers(w *WeightTransferComponent, longitudinal, lateral float64) {
	frontWeight := w.StaticFrontWeight
	rearWeight := 1.0 - w.StaticFrontWeight

	frontWeight += longitudinal
	rearWeight -= longitudinal

	if frontWeight < 0.1 {
		frontWeight = 0.1
	} else if frontWeight > 0.9 {
		frontWeight = 0.9
	}
	rearWeight = 1.0 - frontWeight

	leftRatio := 0.5
	rightRatio := 0.5

	if w.AngularAccel > 0 {
		rightRatio += lateral
		leftRatio -= lateral
	} else if w.AngularAccel < 0 {
		leftRatio += lateral
		rightRatio -= lateral
	}

	if leftRatio < 0.1 {
		leftRatio = 0.1
	} else if leftRatio > 0.9 {
		leftRatio = 0.9
	}
	rightRatio = 1.0 - leftRatio

	w.FrontLeftWeight = frontWeight * leftRatio
	w.FrontRightWeight = frontWeight * rightRatio
	w.RearLeftWeight = rearWeight * leftRatio
	w.RearRightWeight = rearWeight * rightRatio

	total := w.FrontLeftWeight + w.FrontRightWeight + w.RearLeftWeight + w.RearRightWeight
	if total > 0 {
		w.FrontLeftWeight /= total
		w.FrontRightWeight /= total
		w.RearLeftWeight /= total
		w.RearRightWeight /= total
	}
}
