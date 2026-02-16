// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

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
