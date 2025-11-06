package engine

// ProjectileComponent represents a projectile entity with physics properties.
// Projectiles are spawned by ranged weapons and travel until hitting an obstacle,
// enemy, or expiring naturally.
type ProjectileComponent struct {
	// Damage is the base damage dealt on hit
	Damage float64

	// Speed is the movement speed in pixels per second
	Speed float64

	// LifeTime is the maximum duration before despawning (seconds)
	LifeTime float64

	// Age tracks how long the projectile has existed (seconds)
	Age float64

	// Pierce is the number of entities this projectile can pass through
	// 0 = normal (stops on first hit)
	// 1 = pierce 1 enemy
	// -1 = pierce all enemies (infinite)
	Pierce int

	// Bounce is the number of wall bounces remaining
	// 0 = despawn on wall hit
	// >0 = reflect off walls
	Bounce int

	// Explosive indicates if projectile explodes on impact
	Explosive bool

	// ExplosionRadius is the area damage radius in pixels (if Explosive)
	ExplosionRadius float64

	// OwnerID is the entity ID that fired this projectile
	// Used to prevent self-damage and track kills
	OwnerID uint64

	// ProjectileType describes the visual/logical type
	// Examples: "arrow", "bullet", "fireball", "ice_shard"
	ProjectileType string

	// HasHit tracks if projectile has hit anything (for pierce mechanics)
	HasHit bool
}

// Type returns the component type identifier.
func (p ProjectileComponent) Type() string {
	return "projectile"
}

// IsExpired checks if the projectile has exceeded its lifetime.
func (p *ProjectileComponent) IsExpired() bool {
	return p.Age >= p.LifeTime
}

// CanPierce checks if the projectile can pierce another entity.
func (p *ProjectileComponent) CanPierce() bool {
	return p.Pierce < 0 || p.Pierce > 0
}

// DecrementPierce reduces the pierce count after hitting an entity.
// Returns true if projectile should be destroyed (no pierce remaining).
func (p *ProjectileComponent) DecrementPierce() bool {
	if p.Pierce < 0 {
		// Infinite pierce
		return false
	}
	p.Pierce--
	return p.Pierce < 0
}

// CanBounce checks if the projectile can bounce off walls.
func (p *ProjectileComponent) CanBounce() bool {
	return p.Bounce > 0
}

// DecrementBounce reduces the bounce count after hitting a wall.
// Returns true if projectile should be destroyed (no bounces remaining).
func (p *ProjectileComponent) DecrementBounce() bool {
	p.Bounce--
	return p.Bounce < 0
}

// NewProjectileComponent creates a new projectile with standard settings.
func NewProjectileComponent(damage, speed, lifetime float64, projectileType string, ownerID uint64) *ProjectileComponent {
	return &ProjectileComponent{
		Damage:          damage,
		Speed:           speed,
		LifeTime:        lifetime,
		Age:             0.0,
		Pierce:          0,
		Bounce:          0,
		Explosive:       false,
		ExplosionRadius: 0.0,
		OwnerID:         ownerID,
		ProjectileType:  projectileType,
		HasHit:          false,
	}
}

// NewPiercingProjectile creates a projectile with pierce capability.
func NewPiercingProjectile(damage, speed, lifetime float64, pierce int, projectileType string, ownerID uint64) *ProjectileComponent {
	proj := NewProjectileComponent(damage, speed, lifetime, projectileType, ownerID)
	proj.Pierce = pierce
	return proj
}

// NewBouncingProjectile creates a projectile with bounce capability.
func NewBouncingProjectile(damage, speed, lifetime float64, bounce int, projectileType string, ownerID uint64) *ProjectileComponent {
	proj := NewProjectileComponent(damage, speed, lifetime, projectileType, ownerID)
	proj.Bounce = bounce
	return proj
}

// NewExplosiveProjectile creates a projectile that explodes on impact.
func NewExplosiveProjectile(damage, speed, lifetime, explosionRadius float64, projectileType string, ownerID uint64) *ProjectileComponent {
	proj := NewProjectileComponent(damage, speed, lifetime, projectileType, ownerID)
	proj.Explosive = true
	proj.ExplosionRadius = explosionRadius
	return proj
}
