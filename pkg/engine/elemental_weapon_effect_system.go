// Package engine provides the ElementalWeaponEffectSystem which applies
// elemental visual effects to equipped weapons based on enchantment type.
// This system bridges weapon enchantment data (fire, ice, lightning, etc.)
// with the visual effect rendering pipeline.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ElementalEnchantmentType represents the element type of a weapon enchantment.
type ElementalEnchantmentType int

const (
	// ElementalNone indicates no elemental enchantment
	ElementalNone ElementalEnchantmentType = iota
	// ElementalFire adds flames and heat effects
	ElementalFire
	// ElementalIce adds frost and cold effects
	ElementalIce
	// ElementalLightning adds electric arcs and sparks
	ElementalLightning
	// ElementalPoison adds toxic dripping effects
	ElementalPoison
	// ElementalHoly adds golden divine radiance
	ElementalHoly
	// ElementalShadow adds dark void wisps
	ElementalShadow
)

// String returns the string name of the element type.
func (e ElementalEnchantmentType) String() string {
	switch e {
	case ElementalFire:
		return "fire"
	case ElementalIce:
		return "ice"
	case ElementalLightning:
		return "lightning"
	case ElementalPoison:
		return "poison"
	case ElementalHoly:
		return "holy"
	case ElementalShadow:
		return "shadow"
	default:
		return "none"
	}
}

// ParseElementalType converts a string to ElementalEnchantmentType.
func ParseElementalType(s string) ElementalEnchantmentType {
	switch s {
	case "fire", "flame", "burning", "inferno":
		return ElementalFire
	case "ice", "frost", "frozen", "cold", "glacial":
		return ElementalIce
	case "lightning", "electric", "shock", "thunder", "storm":
		return ElementalLightning
	case "poison", "toxic", "venom", "acid", "corrosive":
		return ElementalPoison
	case "holy", "sacred", "divine", "radiant", "light", "blessed":
		return ElementalHoly
	case "shadow", "dark", "void", "unholy", "necrotic", "cursed":
		return ElementalShadow
	default:
		return ElementalNone
	}
}

// ElementalWeaponComponent stores elemental enchantment data for a weapon.
// This component is attached to entities that have elemental weapon enchantments.
type ElementalWeaponComponent struct {
	// ElementType is the primary elemental type of the weapon
	ElementType ElementalEnchantmentType

	// Intensity controls the visual effect strength (0.0-1.0)
	Intensity float64

	// AnimationPhase is the current phase of the effect animation (0.0-1.0, cycling)
	AnimationPhase float64

	// ParticleCount controls how many particles are rendered
	ParticleCount int

	// Seed for deterministic visual generation
	Seed int64

	// IsDirty indicates the visual needs regeneration
	IsDirty bool
}

// Type returns the component type identifier.
func (c *ElementalWeaponComponent) Type() string {
	return "elemental_weapon"
}

// NewElementalWeaponComponent creates a new elemental weapon component.
func NewElementalWeaponComponent(element ElementalEnchantmentType, intensity float64, seed int64) *ElementalWeaponComponent {
	// Scale particle count based on intensity
	particleCount := 4 + int(intensity*6)

	return &ElementalWeaponComponent{
		ElementType:    element,
		Intensity:      intensity,
		AnimationPhase: 0.0,
		ParticleCount:  particleCount,
		Seed:           seed,
		IsDirty:        true,
	}
}

// ElementalWeaponEffectSystem updates elemental weapon visual effects.
// It advances animation phases and marks components dirty when visual refresh is needed.
type ElementalWeaponEffectSystem struct {
	world               *World
	logger              *logrus.Entry
	animationSpeed      float64 // Animation cycles per second
	lastAnimationUpdate float64 // Accumulated time for animation
}

// NewElementalWeaponEffectSystem creates a new elemental weapon effect system.
func NewElementalWeaponEffectSystem(world *World) *ElementalWeaponEffectSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "elemental_weapon_effect",
	})
	logger.Debug("Creating elemental weapon effect system")

	return &ElementalWeaponEffectSystem{
		world:               world,
		logger:              logger,
		animationSpeed:      0.5, // Complete one animation cycle every 2 seconds
		lastAnimationUpdate: 0,
	}
}

// Update processes all entities with elemental weapon components.
// It advances animation phases based on delta time.
func (s *ElementalWeaponEffectSystem) Update(entities []*Entity, deltaTime float64) {
	if deltaTime <= 0 {
		return
	}

	// Calculate phase increment for this frame
	phaseIncrement := s.animationSpeed * deltaTime

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		// Check for elemental weapon component
		comp, ok := entity.GetComponent("elemental_weapon")
		if !ok || comp == nil {
			continue
		}

		ewc, ok := comp.(*ElementalWeaponComponent)
		if !ok || ewc.ElementType == ElementalNone {
			continue
		}

		// Advance animation phase
		ewc.AnimationPhase += phaseIncrement
		if ewc.AnimationPhase >= 1.0 {
			ewc.AnimationPhase = math.Mod(ewc.AnimationPhase, 1.0)
		}

		// Mark as dirty for visual refresh (at reduced rate for performance)
		s.lastAnimationUpdate += deltaTime
		if s.lastAnimationUpdate >= 0.1 { // Refresh visuals ~10 times per second
			ewc.IsDirty = true
			s.lastAnimationUpdate = 0
		}
	}
}

// GetElementalParamsForEntity extracts elemental effect parameters from an entity.
// Returns nil if the entity has no elemental weapon component.
func GetElementalParamsForEntity(entity *Entity) *ElementalWeaponComponent {
	if entity == nil {
		return nil
	}

	comp, ok := entity.GetComponent("elemental_weapon")
	if !ok || comp == nil {
		return nil
	}

	ewc, ok := comp.(*ElementalWeaponComponent)
	if !ok {
		return nil
	}

	return ewc
}

// CreateElementalWeaponFromItem creates an ElementalWeaponComponent from item enchantment data.
// enchantmentType is the enchantment name (e.g., "fire", "frost", "lightning").
// rarity affects the intensity (common=0.4, uncommon=0.5, rare=0.6, epic=0.75, legendary=0.9).
// seed provides deterministic visual generation.
func CreateElementalWeaponFromItem(enchantmentType, rarity string, seed int64) *ElementalWeaponComponent {
	element := ParseElementalType(enchantmentType)
	if element == ElementalNone {
		return nil
	}

	// Intensity scales with rarity
	var intensity float64
	switch rarity {
	case "common":
		intensity = 0.4
	case "uncommon":
		intensity = 0.5
	case "rare":
		intensity = 0.65
	case "epic":
		intensity = 0.8
	case "legendary":
		intensity = 0.95
	default:
		intensity = 0.5
	}

	return NewElementalWeaponComponent(element, intensity, seed)
}

// ApplyElementalEnchantmentToEntity adds or updates an elemental weapon component on an entity.
// If the entity already has an elemental weapon component, it updates the values.
func ApplyElementalEnchantmentToEntity(entity *Entity, enchantmentType, rarity string, seed int64) bool {
	if entity == nil {
		return false
	}

	comp := CreateElementalWeaponFromItem(enchantmentType, rarity, seed)
	if comp == nil {
		return false
	}

	// Check for existing component and update or add
	existing, ok := entity.GetComponent("elemental_weapon")
	if ok && existing != nil {
		if ewc, ok := existing.(*ElementalWeaponComponent); ok {
			ewc.ElementType = comp.ElementType
			ewc.Intensity = comp.Intensity
			ewc.ParticleCount = comp.ParticleCount
			ewc.Seed = seed
			ewc.IsDirty = true
			return true
		}
	}

	// Add new component
	entity.AddComponent(comp)
	return true
}

// GenerateElementalSeed creates a deterministic seed for elemental effects from item properties.
func GenerateElementalSeed(itemID string, baseSeed int64) int64 {
	// Simple hash combination
	rng := rand.New(rand.NewSource(baseSeed))
	hash := int64(0)
	for _, c := range itemID {
		hash = hash*31 + int64(c)
	}
	return hash ^ rng.Int63()
}

// ElementalEffectColors returns the primary color components for an element type.
// Returns R, G, B values (0-255) for the element's signature color.
func ElementalEffectColors(element ElementalEnchantmentType) (r, g, b uint8) {
	switch element {
	case ElementalFire:
		return 255, 140, 30 // Orange
	case ElementalIce:
		return 150, 220, 255 // Light blue
	case ElementalLightning:
		return 220, 220, 255 // Electric blue-white
	case ElementalPoison:
		return 100, 200, 60 // Green
	case ElementalHoly:
		return 255, 230, 150 // Golden
	case ElementalShadow:
		return 80, 40, 120 // Purple
	default:
		return 200, 200, 200 // Gray
	}
}
