// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

import (
	"math"
)

// CollisionResponseComponent handles realistic vehicle damage from impacts.
type CollisionResponseComponent struct {
	// Impact tracking
	LastImpactVelocity float64 // Speed at last collision (pixels/s)
	LastImpactForce    float64 // Force of last impact (approximated)
	LastImpactAngle    float64 // Angle of impact (radians)
	TotalImpactDamage  float64 // Cumulative damage from all impacts

	// Collision parameters
	DamageThreshold     float64 // Minimum velocity for damage (pixels/s)
	MassForCalculation  float64 // Vehicle mass for force calculation (kg)
	StructuralIntegrity float64 // 0.0 = destroyed, 1.0 = pristine

	// Bounce/restitution
	Restitution float64 // Bounciness [0.0, 1.0], typically 0.1-0.3 for vehicles

	// Performance tracking
	CollisionCount int
}

// Type returns the component type identifier.
func (c *CollisionResponseComponent) Type() string {
	return "collision_response"
}

// NewCollisionResponseComponent creates a collision response component.
func NewCollisionResponseComponent(mass float64) *CollisionResponseComponent {
	return &CollisionResponseComponent{
		DamageThreshold:     50.0, // 50 pixels/s minimum for damage
		MassForCalculation:  mass,
		StructuralIntegrity: 1.0, // Start pristine
		Restitution:         0.2, // Low bounce (20% energy retained)
		CollisionCount:      0,
	}
}

// ProcessCollision calculates damage and response from a collision.
// velocityX, velocityY: vehicle velocity before impact
// normalX, normalY: surface normal at collision point (unit vector)
// Returns: impact result with damage and velocity changes
func (c *CollisionResponseComponent) ProcessCollision(velocityX, velocityY, normalX, normalY float64) ImpactResult {
	// Calculate impact velocity magnitude
	impactSpeed := math.Sqrt(velocityX*velocityX + velocityY*velocityY)
	c.LastImpactVelocity = impactSpeed
	c.CollisionCount++

	// No damage below threshold
	if impactSpeed < c.DamageThreshold {
		// Just reflect velocity with restitution
		reflectedVel := c.reflectVelocity(velocityX, velocityY, normalX, normalY)
		return ImpactResult{
			DamageDealt:       0.0,
			VelocityReduction: impactSpeed * (1.0 - c.Restitution),
			BounceVelocityX:   reflectedVel[0],
			BounceVelocityY:   reflectedVel[1],
			IntegrityLoss:     0.0,
		}
	}

	// Calculate impact force (F = m * Δv)
	// Approximate collision time as 0.1 seconds
	collisionTime := 0.1
	deltaV := impactSpeed // Assuming full stop then bounce
	force := (c.MassForCalculation * deltaV) / collisionTime
	c.LastImpactForce = force

	// Calculate angle of impact (affects damage)
	// Head-on collision = max damage, glancing blow = less damage
	velocityMag := math.Sqrt(velocityX*velocityX + velocityY*velocityY)
	if velocityMag == 0 {
		velocityMag = 1.0 // Avoid division by zero
	}

	// Dot product gives cosine of angle between velocity and normal
	// cos(0°) = 1.0 (head-on), cos(90°) = 0.0 (glancing)
	dotProduct := (velocityX*normalX + velocityY*normalY) / velocityMag
	dotProduct = math.Abs(dotProduct) // Magnitude only
	c.LastImpactAngle = math.Acos(math.Max(-1.0, math.Min(1.0, dotProduct)))

	// Damage scales with impact speed and angle
	// Formula: damage = (speed - threshold)² * angleFactor * damageCoeff
	speedFactor := (impactSpeed - c.DamageThreshold) / 100.0 // Normalize
	angleFactor := dotProduct                                // More damage for head-on (closer to 1.0)
	damageCoeff := 0.5                                       // Base damage coefficient

	damage := speedFactor * speedFactor * angleFactor * damageCoeff
	damage = math.Max(0.0, math.Min(damage, 100.0)) // Clamp to [0, 100]

	c.TotalImpactDamage += damage

	// Structural integrity loss (permanent)
	integrityLoss := damage * 0.01 // 1% per point of damage
	c.StructuralIntegrity -= integrityLoss
	if c.StructuralIntegrity < 0.0 {
		c.StructuralIntegrity = 0.0
	}

	// Calculate bounce velocity (reflect and apply restitution)
	reflectedVel := c.reflectVelocity(velocityX, velocityY, normalX, normalY)

	// Apply restitution (scaled by structural integrity)
	// Damaged vehicles bounce less
	effectiveRestitution := c.Restitution * c.StructuralIntegrity
	bounceVelX := reflectedVel[0] * effectiveRestitution
	bounceVelY := reflectedVel[1] * effectiveRestitution

	// Velocity reduction is difference between original and bounce
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
func (c *CollisionResponseComponent) reflectVelocity(vx, vy, nx, ny float64) [2]float64 {
	// Normalize normal vector (should already be normalized, but safety check)
	nMag := math.Sqrt(nx*nx + ny*ny)
	if nMag > 0 {
		nx /= nMag
		ny /= nMag
	}

	// Dot product: v · n
	dotProduct := vx*nx + vy*ny

	// Reflection: v - 2(v·n)n
	reflectX := vx - 2.0*dotProduct*nx
	reflectY := vy - 2.0*dotProduct*ny

	return [2]float64{reflectX, reflectY}
}

// GetDamageMultiplier returns a multiplier based on structural integrity.
// Used to reduce vehicle performance as it gets damaged.
func (c *CollisionResponseComponent) GetDamageMultiplier() float64 {
	// At 100% integrity: 1.0x performance
	// At 50% integrity: 0.75x performance
	// At 0% integrity: 0.5x performance (minimum)
	return 0.5 + (c.StructuralIntegrity * 0.5)
}

// IsDestroyed checks if structural integrity is depleted.
func (c *CollisionResponseComponent) IsDestroyed() bool {
	return c.StructuralIntegrity <= 0.0
}

// GetIntegrity returns the current structural integrity [0.0, 1.0].
func (c *CollisionResponseComponent) GetIntegrity() float64 {
	return c.StructuralIntegrity
}

// Repair increases structural integrity.
func (c *CollisionResponseComponent) Repair(amount float64) {
	c.StructuralIntegrity += amount
	if c.StructuralIntegrity > 1.0 {
		c.StructuralIntegrity = 1.0
	}
}

// Reset resets collision tracking (used when respawning vehicle).
func (c *CollisionResponseComponent) Reset() {
	c.LastImpactVelocity = 0.0
	c.LastImpactForce = 0.0
	c.LastImpactAngle = 0.0
	c.TotalImpactDamage = 0.0
	c.StructuralIntegrity = 1.0
	c.CollisionCount = 0
}

// GetCollisionCount returns the number of collisions processed.
func (c *CollisionResponseComponent) GetCollisionCount() int {
	return c.CollisionCount
}

// GetLastImpactForce returns the force of the most recent impact.
func (c *CollisionResponseComponent) GetLastImpactForce() float64 {
	return c.LastImpactForce
}

// GetLastImpactVelocity returns the velocity of the most recent impact.
func (c *CollisionResponseComponent) GetLastImpactVelocity() float64 {
	return c.LastImpactVelocity
}

// ShouldCauseDamage checks if the given impact speed exceeds the damage threshold.
func (c *CollisionResponseComponent) ShouldCauseDamage(impactSpeed float64) bool {
	return impactSpeed >= c.DamageThreshold
}
