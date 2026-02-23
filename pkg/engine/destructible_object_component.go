// Package engine provides destructible object components for Phase 11.3.
// Environmental Destruction & Manipulation
//
// This file defines components for destructible objects (crates, barrels, furniture)
// that can be damaged, destroyed, and generate debris. Distinct from terrain destruction
// (DestructibleComponent), these are entity-based objects that exist in the world.
//
// Design Philosophy:
// - Objects are entities with health, not tiles
// - Destruction generates debris entities with physics
// - Objects can be explosive (barrels) or emit hazards (poison containers)
// - Server-authoritative for multiplayer synchronization
//
// ECS Compliance:
// - Components are pure data structures with only Type() method
// - All damage logic is handled by DestructibleObjectSystem
// - Use ApplyDamageToComponent() system method instead of component.TakeDamage()
package engine

// ObjectType represents the type of destructible object.
type ObjectType int

const (
	// ObjectCrate represents a wooden crate (common loot container)
	ObjectCrate ObjectType = iota
	// ObjectBarrel represents a barrel (can be explosive)
	ObjectBarrel
	// ObjectFurniture represents furniture (tables, chairs, beds)
	ObjectFurniture
	// ObjectWeakWall represents a cracked or damaged wall section
	ObjectWeakWall
	// ObjectPoisonContainer represents a container with poison/toxic contents
	ObjectPoisonContainer
	// ObjectExplosiveBarrel represents a barrel that explodes on destruction
	ObjectExplosiveBarrel
)

// String returns the string representation of an object type.
func (o ObjectType) String() string {
	switch o {
	case ObjectCrate:
		return "crate"
	case ObjectBarrel:
		return "barrel"
	case ObjectFurniture:
		return "furniture"
	case ObjectWeakWall:
		return "weak_wall"
	case ObjectPoisonContainer:
		return "poison_container"
	case ObjectExplosiveBarrel:
		return "explosive_barrel"
	default:
		return "unknown"
	}
}

// DestructibleObjectComponent marks an entity as a destructible object.
// Unlike DestructibleComponent (for tiles), this is for entity-based objects.
type DestructibleObjectComponent struct {
	// ObjectType determines behavior on destruction
	ObjectType ObjectType

	// Health is current durability (0 = destroyed)
	Health float64

	// MaxHealth is the starting durability
	MaxHealth float64

	// IsDestroyed tracks if this object has been destroyed
	IsDestroyed bool

	// ParticlesSpawned tracks if destruction particles have been emitted
	ParticlesSpawned bool

	// LastDamageTime tracks when the object was last damaged (game time in seconds).
	// This is set by DestructibleObjectSystem during damage processing.
	// Use game time (from World.Clock or delta accumulator) rather than wall-clock time
	// to ensure deterministic behavior in multiplayer.
	LastDamageTime float64

	// ExplosionRadius is the area damage radius for explosive objects (0 = not explosive)
	ExplosionRadius float64

	// ExplosionDamage is the damage dealt by explosion
	ExplosionDamage float64

	// PoisonDuration is how long poison cloud lingers after destruction (seconds)
	PoisonDuration float64

	// PoisonRadius is the area of effect for poison cloud (0 = not poison)
	PoisonRadius float64

	// DebrisCount is how many debris entities to spawn on destruction
	DebrisCount int

	// LootTable is optional reference to loot generation (for crates)
	LootTable string
}

// Type returns the component type identifier.
func (d *DestructibleObjectComponent) Type() string {
	return "destructibleObject"
}

// NewDestructibleObjectComponent creates a destructible object component.
func NewDestructibleObjectComponent(objType ObjectType) *DestructibleObjectComponent {
	comp := &DestructibleObjectComponent{
		ObjectType:     objType,
		IsDestroyed:    false,
		LastDamageTime: 0, // Will be set by system when damage is applied
		DebrisCount:    3, // Default: 3 debris pieces
	}

	// Set type-specific properties
	switch objType {
	case ObjectCrate:
		comp.Health = 20.0
		comp.MaxHealth = 20.0
		comp.LootTable = "common_items"
	case ObjectBarrel:
		comp.Health = 30.0
		comp.MaxHealth = 30.0
	case ObjectFurniture:
		comp.Health = 15.0
		comp.MaxHealth = 15.0
	case ObjectWeakWall:
		comp.Health = 40.0
		comp.MaxHealth = 40.0
		comp.DebrisCount = 5
	case ObjectPoisonContainer:
		comp.Health = 15.0
		comp.MaxHealth = 15.0
		comp.PoisonDuration = 10.0 // 10 seconds of poison cloud
		comp.PoisonRadius = 64.0   // 2 tiles radius
	case ObjectExplosiveBarrel:
		comp.Health = 25.0
		comp.MaxHealth = 25.0
		comp.ExplosionRadius = 96.0 // 3 tiles
		comp.ExplosionDamage = 50.0 // Heavy damage
		comp.DebrisCount = 8        // More debris from explosion
	}

	return comp
}

// TakeDamage applies damage to the object and returns true if destroyed.
//
// Deprecated: This method violates ECS principles by containing logic in a component.
// Use DestructibleObjectSystem.ApplyDamageToComponent() instead for new code.
// This method is preserved for backward compatibility but the LastDamageTime field
// will not be updated (it should be set by the system with game time).
func (d *DestructibleObjectComponent) TakeDamage(damage float64) bool {
	d.Health -= damage
	// NOTE: LastDamageTime is not updated here to encourage using the system method
	// which can properly track game time rather than wall-clock time.
	if d.Health <= 0 {
		d.Health = 0
		d.IsDestroyed = true
		return true
	}
	return false
}

// HealthPercent returns the health as a percentage (0.0-1.0).
func (d *DestructibleObjectComponent) HealthPercent() float64 {
	if d.MaxHealth <= 0 {
		return 0
	}
	return d.Health / d.MaxHealth
}

// IsExplosive returns true if this object explodes on destruction.
func (d *DestructibleObjectComponent) IsExplosive() bool {
	return d.ExplosionRadius > 0
}

// EmitsPoison returns true if this object emits poison on destruction.
func (d *DestructibleObjectComponent) EmitsPoison() bool {
	return d.PoisonDuration > 0
}

// DebrisComponent marks an entity as a debris piece from destroyed object.
// Debris has physics (velocity, rotation) and a lifetime before despawning.
type DebrisComponent struct {
	// SourceObjectType is what object this came from
	SourceObjectType ObjectType

	// Lifetime is how long debris exists before despawning (seconds)
	Lifetime float64

	// MaxLifetime is the initial lifetime value
	MaxLifetime float64

	// AngularVelocity is rotation speed (radians per second)
	AngularVelocity float64

	// IsStationary becomes true when velocity is very low
	IsStationary bool
}

// Type returns the component type identifier.
func (d *DebrisComponent) Type() string {
	return "debris"
}

// NewDebrisComponent creates a debris component.
func NewDebrisComponent(sourceType ObjectType, lifetime, angularVel float64) *DebrisComponent {
	if lifetime <= 0 {
		lifetime = 5.0 // Default: 5 seconds
	}
	return &DebrisComponent{
		SourceObjectType: sourceType,
		Lifetime:         lifetime,
		MaxLifetime:      lifetime,
		AngularVelocity:  angularVel,
		IsStationary:     false,
	}
}

// Update advances debris lifetime.
func (d *DebrisComponent) Update(deltaTime float64) {
	d.Lifetime -= deltaTime
}

// ShouldDespawn returns true if debris has expired.
func (d *DebrisComponent) ShouldDespawn() bool {
	return d.Lifetime < 0
}

// RemainingLifetimePercent returns lifetime as percentage (0.0-1.0).
func (d *DebrisComponent) RemainingLifetimePercent() float64 {
	if d.MaxLifetime <= 0 {
		return 0
	}
	percent := d.Lifetime / d.MaxLifetime
	if percent < 0 {
		return 0
	}
	if percent > 1.0 {
		return 1.0
	}
	return percent
}
