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
// Deprecated: Use vehicle.GetDamageMultiplier(collision) instead to maintain ECS purity.
// Used to reduce vehicle performance as it gets damaged.
func (c *CollisionResponseComponent) GetDamageMultiplier() float64 {
	return GetDamageMultiplier(c)
}

// IsDestroyed checks if structural integrity is depleted.
// Deprecated: Use vehicle.IsDestroyed(collision) instead to maintain ECS purity.
func (c *CollisionResponseComponent) IsDestroyed() bool {
	return IsDestroyed(c)
}

// GetIntegrity returns the current structural integrity [0.0, 1.0].
// Deprecated: Use direct field access (collision.StructuralIntegrity) instead to maintain ECS purity.
func (c *CollisionResponseComponent) GetIntegrity() float64 {
	return c.StructuralIntegrity
}

// Repair increases structural integrity.
// Deprecated: Use vehicle.RepairVehicle(collision, amount) instead to maintain ECS purity.
func (c *CollisionResponseComponent) Repair(amount float64) {
	RepairVehicle(c, amount)
}

// Reset resets collision tracking (used when respawning vehicle).
// Deprecated: Use vehicle.ResetCollisionResponse(collision) instead to maintain ECS purity.
func (c *CollisionResponseComponent) Reset() {
	ResetCollisionResponse(c)
}

// GetCollisionCount returns the number of collisions processed.
// Deprecated: Use direct field access (collision.CollisionCount) instead to maintain ECS purity.
func (c *CollisionResponseComponent) GetCollisionCount() int {
	return c.CollisionCount
}

// GetLastImpactForce returns the force of the most recent impact.
// Deprecated: Use direct field access (collision.LastImpactForce) instead to maintain ECS purity.
func (c *CollisionResponseComponent) GetLastImpactForce() float64 {
	return c.LastImpactForce
}

// GetLastImpactVelocity returns the velocity of the most recent impact.
// Deprecated: Use direct field access (collision.LastImpactVelocity) instead to maintain ECS purity.
func (c *CollisionResponseComponent) GetLastImpactVelocity() float64 {
	return c.LastImpactVelocity
}

// ShouldCauseDamage checks if the given impact speed exceeds the damage threshold.
// Deprecated: Use vehicle.ShouldCauseDamage(collision, impactSpeed) instead to maintain ECS purity.
func (c *CollisionResponseComponent) ShouldCauseDamage(impactSpeed float64) bool {
	return ShouldCauseDamage(c, impactSpeed)
}
