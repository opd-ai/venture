// Package engine provides carriable object components for Phase 11.3.
// Environmental Destruction & Manipulation
//
// This file defines components for objects that can be picked up and thrown
// by the player. Carriable objects have weight (affects throw velocity) and
// can deal impact damage when thrown.
//
// Design Philosophy:
// - Objects attach to player when picked up (follow player position)
// - Throwing launches object in aim direction with arc physics
// - Impact damage based on object weight and velocity
// - Server-authoritative for multiplayer synchronization
package engine

// CarriableComponent marks an entity as something that can be picked up and thrown.
type CarriableComponent struct {
	// Weight affects throw velocity and impact damage (0.1 = light, 1.0 = heavy)
	Weight float64

	// IsCarried tracks if object is currently being carried
	IsCarried bool

	// CarriedBy is the entity ID of who is carrying this object (0 = not carried)
	CarriedBy uint64

	// ThrowVelocityMultiplier affects how far/fast object flies when thrown
	ThrowVelocityMultiplier float64

	// ImpactDamage is damage dealt when thrown object hits entity
	ImpactDamage float64

	// CanPickUp determines if player can interact with this object
	CanPickUp bool
}

// Type returns the component type identifier.
func (c *CarriableComponent) Type() string {
	return "carriable"
}

// NewCarriableComponent creates a carriable component.
// weight: 0.1 (light crate) to 1.0 (heavy barrel)
func NewCarriableComponent(weight float64) *CarriableComponent {
	if weight < 0.1 {
		weight = 0.1
	}
	if weight > 1.0 {
		weight = 1.0
	}

	// Lighter objects throw farther and faster
	throwVelMultiplier := 1.0 / weight

	// Heavier objects deal more impact damage
	impactDamage := 10.0 * weight

	return &CarriableComponent{
		Weight:                  weight,
		IsCarried:               false,
		CarriedBy:               0,
		ThrowVelocityMultiplier: throwVelMultiplier,
		ImpactDamage:            impactDamage,
		CanPickUp:               true,
	}
}

// Pickup marks object as carried by an entity.
func (c *CarriableComponent) Pickup(carrierID uint64) {
	c.IsCarried = true
	c.CarriedBy = carrierID
}

// Drop marks object as no longer carried.
func (c *CarriableComponent) Drop() {
	c.IsCarried = false
	c.CarriedBy = 0
}

// ContextActionComponent marks an entity as having context-sensitive interactions.
// When player is near, displays prompt: "Press F to [action]"
type ContextActionComponent struct {
	// ActionText is what displays in prompt: "Press F to Open", "Press F to Push"
	ActionText string

	// ActionType determines what happens when F is pressed
	ActionType ContextActionType

	// InteractionRange is how close player must be (pixels)
	InteractionRange float64

	// IsAvailable determines if action can be performed now
	IsAvailable bool

	// RequiresKey for doors that need a key item
	RequiresKey bool
	KeyItemID   string

	// RequiresLockPicking for doors/chests that need lock-picking mini-game (Phase 27.3)
	RequiresLockPicking bool
	// LockDifficulty is the difficulty of the lock-picking mini-game (0.0-1.0)
	LockDifficulty float64

	// CooldownTime prevents rapid re-activation (seconds)
	CooldownTime float64

	// CooldownElapsed tracks cooldown timer (seconds)
	CooldownElapsed float64
}

// ContextActionType represents different interaction types.
type ContextActionType int

const (
	// ActionOpen for doors and chests
	ActionOpen ContextActionType = iota
	// ActionClose for doors
	ActionClose
	// ActionPush for movable objects
	ActionPush
	// ActionPull for movable objects
	ActionPull
	// ActionActivate for levers, switches, buttons
	ActionActivate
	// ActionTalk for NPCs
	ActionTalk
	// ActionPickup for carriable objects
	ActionPickup
	// ActionRead for signs, books
	ActionRead
	// ActionPlayGame for mini-game stations (Phase 27.3)
	ActionPlayGame
	// ActionInvestigate for environmental investigation (Phase 30.2)
	ActionInvestigate
)

// String returns the string representation of an action type.
func (a ContextActionType) String() string {
	switch a {
	case ActionOpen:
		return "Open"
	case ActionClose:
		return "Close"
	case ActionPush:
		return "Push"
	case ActionPull:
		return "Pull"
	case ActionActivate:
		return "Activate"
	case ActionTalk:
		return "Talk"
	case ActionPickup:
		return "Pickup"
	case ActionRead:
		return "Read"
	case ActionPlayGame:
		return "Play"
	case ActionInvestigate:
		return "Investigate"
	default:
		return "Interact"
	}
}

// Type returns the component type identifier.
func (c *ContextActionComponent) Type() string {
	return "contextAction"
}

// NewContextActionComponent creates a context action component.
func NewContextActionComponent(actionType ContextActionType, actionText string) *ContextActionComponent {
	return &ContextActionComponent{
		ActionText:          actionText,
		ActionType:          actionType,
		InteractionRange:    48.0, // Default: 1.5 tiles
		IsAvailable:         true,
		RequiresKey:         false,
		RequiresLockPicking: false,
		LockDifficulty:      0.5, // Default medium difficulty
		CooldownTime:        0.5, // Default: 0.5 second cooldown
		CooldownElapsed:     0,
	}
}

// Update advances the cooldown timer.
func (c *ContextActionComponent) Update(deltaTime float64) {
	if c.CooldownElapsed > 0 {
		c.CooldownElapsed -= deltaTime
		if c.CooldownElapsed < 0 {
			c.CooldownElapsed = 0
		}
	}
}

// CanInteract returns true if action is available and off cooldown.
func (c *ContextActionComponent) CanInteract() bool {
	return c.IsAvailable && c.CooldownElapsed <= 0
}

// Activate triggers the action and starts cooldown.
func (c *ContextActionComponent) Activate() {
	c.CooldownElapsed = c.CooldownTime
}

// HazardComponent marks an entity as an environmental hazard (poison cloud, oil puddle).
type HazardComponent struct {
	// HazardType determines the effect
	HazardType HazardType

	// Duration is how long hazard persists (seconds, 0 = permanent)
	Duration float64

	// DamagePerSecond for damaging hazards
	DamagePerSecond float64

	// MovementMultiplier for movement-affecting hazards (1.0 = normal, 0.5 = slowed)
	MovementMultiplier float64

	// Radius is the affected area (pixels)
	Radius float64

	// IsLingering determines if hazard stays after source is destroyed
	IsLingering bool
}

// HazardType represents different environmental hazard types.
type HazardType int

const (
	// HazardPoison damages entities over time
	HazardPoison HazardType = iota
	// HazardOil slows movement and spreads fire
	HazardOil
	// HazardWater slows movement and extinguishes fire
	HazardWater
	// HazardSmoke obscures vision
	HazardSmoke
)

// String returns the string representation of a hazard type.
func (h HazardType) String() string {
	switch h {
	case HazardPoison:
		return "poison"
	case HazardOil:
		return "oil"
	case HazardWater:
		return "water"
	case HazardSmoke:
		return "smoke"
	default:
		return "unknown"
	}
}

// Type returns the component type identifier.
func (h *HazardComponent) Type() string {
	return "hazard"
}

// NewHazardComponent creates a hazard component.
func NewHazardComponent(hazardType HazardType, duration, radius float64) *HazardComponent {
	comp := &HazardComponent{
		HazardType:         hazardType,
		Duration:           duration,
		Radius:             radius,
		IsLingering:        true,
		MovementMultiplier: 1.0,
	}

	// Set type-specific properties
	switch hazardType {
	case HazardPoison:
		comp.DamagePerSecond = 5.0
	case HazardOil:
		comp.MovementMultiplier = 0.7 // 30% movement penalty
	case HazardWater:
		comp.MovementMultiplier = 0.8 // 20% movement penalty
	case HazardSmoke:
		// Smoke doesn't damage or slow, just obscures
	}

	return comp
}

// Update advances hazard duration timer.
func (h *HazardComponent) Update(deltaTime float64) {
	if h.Duration > 0 {
		h.Duration -= deltaTime
	}
}

// ShouldRemove returns true if hazard has expired.
func (h *HazardComponent) ShouldRemove() bool {
	return h.Duration < 0 && h.Duration != -1 // -1 = permanent
}

// IsDamaging returns true if hazard deals damage.
func (h *HazardComponent) IsDamaging() bool {
	return h.DamagePerSecond > 0
}

// AffectsMovement returns true if hazard slows movement.
func (h *HazardComponent) AffectsMovement() bool {
	return h.MovementMultiplier != 1.0
}
