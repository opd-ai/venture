// Package sprites - anatomical template system for structured sprite generation.
// This file implements Phase 5.1 of the Visual Fidelity Enhancement Plan.
package sprites

import (
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// BodyPart represents a distinct anatomical region of a sprite.
type BodyPart int

const (
	// PartShadow represents a ground shadow
	PartShadow BodyPart = iota
	// PartLegs represents lower limbs
	PartLegs
	// PartTorso represents main body/chest
	PartTorso
	// PartArms represents upper limbs
	PartArms
	// PartHead represents head/face
	PartHead
	// PartWeapon represents equipped weapon
	PartWeapon
	// PartShield represents equipped shield
	PartShield
	// PartHelmet represents head armor
	PartHelmet
	// PartArmor represents body armor overlay
	PartArmor
	// PartTail represents tail (for certain creatures)
	PartTail
	// PartWings represents wings (left/right combined)
	PartWings
)

// Direction represents facing direction for directional sprites.
type Direction string

const (
	// DirUp represents facing upward
	DirUp Direction = "up"
	// DirDown represents facing downward
	DirDown Direction = "down"
	// DirLeft represents facing left
	DirLeft Direction = "left"
	// DirRight represents facing right
	DirRight Direction = "right"
)

// String returns the string representation of a body part.
func (b BodyPart) String() string {
	switch b {
	case PartShadow:
		return "shadow"
	case PartLegs:
		return "legs"
	case PartTorso:
		return "torso"
	case PartArms:
		return "arms"
	case PartHead:
		return "head"
	case PartWeapon:
		return "weapon"
	case PartShield:
		return "shield"
	case PartHelmet:
		return "helmet"
	case PartArmor:
		return "armor"
	case PartTail:
		return "tail"
	case PartWings:
		return "wings"
	default:
		return "unknown"
	}
}

// PartSpec defines the rendering specification for a body part.
type PartSpec struct {
	// RelativeX is the X position as a fraction of sprite width (0.0-1.0)
	RelativeX float64
	// RelativeY is the Y position as a fraction of sprite height (0.0-1.0)
	RelativeY float64
	// RelativeWidth is the width as a fraction of sprite width (0.0-1.0)
	RelativeWidth float64
	// RelativeHeight is the height as a fraction of sprite height (0.0-1.0)
	RelativeHeight float64
	// ShapeTypes are the allowed shapes for this part
	ShapeTypes []shapes.ShapeType
	// ZIndex determines draw order (lower drawn first)
	ZIndex int
	// ColorRole indicates which palette color to use ("primary", "secondary", "accent1", etc.)
	ColorRole string
	// Opacity is the alpha transparency (0.0-1.0)
	Opacity float64
	// Rotation is the rotation angle in degrees (0-360)
	Rotation float64
}

// AnatomicalTemplate defines the layout and structure of a sprite.
type AnatomicalTemplate struct {
	// Name identifies this template
	Name string
	// BodyPartLayout maps body parts to their specifications
	BodyPartLayout map[BodyPart]PartSpec
}

// GetSortedParts returns body parts sorted by Z-index for correct rendering order.
func (t *AnatomicalTemplate) GetSortedParts() []struct {
	Part BodyPart
	Spec PartSpec
} {
	// Create slice of parts with specs
	parts := make([]struct {
		Part BodyPart
		Spec PartSpec
	}, 0, len(t.BodyPartLayout))

	for part, spec := range t.BodyPartLayout {
		parts = append(parts, struct {
			Part BodyPart
			Spec PartSpec
		}{Part: part, Spec: spec})
	}

	// Sort by Z-index (bubble sort is fine for small slices)
	for i := 0; i < len(parts); i++ {
		for j := i + 1; j < len(parts); j++ {
			if parts[j].Spec.ZIndex < parts[i].Spec.ZIndex {
				parts[i], parts[j] = parts[j], parts[i]
			}
		}
	}

	return parts
}

// HumanoidTemplate returns the default humanoid anatomical template.
// This is optimized for 28x28 pixel sprites (player size).
// Proportions: Head 30%, Torso 40%, Legs 30% (top-down perspective).
func HumanoidTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "humanoid",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.93,
				RelativeWidth:  0.40,
				RelativeHeight: 0.12,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.3,
				Rotation:       0,
			},
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.35,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:         5,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.50,
				RelativeHeight: 0.45,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeRectangle, shapes.ShapeEllipse},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.65,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
				ZIndex:         8,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.35,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse, shapes.ShapeSkull},
				ZIndex:         15,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       0,
			},
		},
	}
}

// QuadrupedTemplate returns a template for four-legged creatures.
// Optimized for 32x32 pixels (standard enemy size).
func QuadrupedTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "quadruped",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.90,
				RelativeWidth:  0.60,
				RelativeHeight: 0.15,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.3,
				Rotation:       0,
			},
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.70,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:         5,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       90, // Horizontal orientation
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.70,
				RelativeHeight: 0.50,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       90, // Horizontal body
			},
			PartHead: {
				RelativeX:      0.25,
				RelativeY:      0.35,
				RelativeWidth:  0.30,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse, shapes.ShapeWedge},
				ZIndex:         15,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       270, // Face left
			},
		},
	}
}

// BlobTemplate returns a template for amorphous creatures.
// Optimized for 32x32 pixels (slimes, amoebas).
func BlobTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "blob",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.85,
				RelativeWidth:  0.70,
				RelativeHeight: 0.20,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.4,
				Rotation:       0,
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.55,
				RelativeWidth:  0.80,
				RelativeHeight: 0.70,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeCircle, shapes.ShapeBean},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        0.9, // Slightly translucent
				Rotation:       0,
			},
		},
	}
}

// MechanicalTemplate returns a template for robots and constructs.
// Optimized for 32x32 pixels (robots, golems).
func MechanicalTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "mechanical",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.93,
				RelativeWidth:  0.40,
				RelativeHeight: 0.12,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.3,
				Rotation:       0,
			},
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.35,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
				ZIndex:         5,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.55,
				RelativeHeight: 0.45,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon, shapes.ShapeOctagon},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.70,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
				ZIndex:         8,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.35,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon, shapes.ShapeOctagon},
				ZIndex:         15,
				ColorRole:      "accent1",
				Opacity:        1.0,
				Rotation:       0,
			},
		},
	}
}

// FlyingTemplate returns a template for winged creatures.
// Optimized for 32x32 pixels (birds, dragons, flying enemies).
func FlyingTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "flying",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.88,
				RelativeWidth:  0.35,
				RelativeHeight: 0.15,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.25, // Lighter shadow (flying)
				Rotation:       0,
			},
			// Left wing (behind body)
			PartLegs: { // Reuse legs part for left wing
				RelativeX:      0.25,
				RelativeY:      0.50,
				RelativeWidth:  0.45,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle},
				ZIndex:         5,
				ColorRole:      "secondary",
				Opacity:        0.9,
				Rotation:       270, // Point left
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.55,
				RelativeWidth:  0.40,
				RelativeHeight: 0.50,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.30,
				RelativeWidth:  0.30,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
				ZIndex:         12,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       0,
			},
			// Right wing (in front of body)
			PartArms: { // Reuse arms part for right wing
				RelativeX:      0.75,
				RelativeY:      0.50,
				RelativeWidth:  0.45,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle},
				ZIndex:         15,
				ColorRole:      "secondary",
				Opacity:        0.9,
				Rotation:       90, // Point right
			},
		},
	}
}

// SelectTemplate chooses an appropriate template based on entity type.
// entityType should be provided in Config.Custom["entityType"].
// Returns HumanoidTemplate as default fallback.
func SelectTemplate(entityType string) AnatomicalTemplate {
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior":
		return HumanoidTemplate()
	case "quadruped", "wolf", "bear", "animal", "beast":
		return QuadrupedTemplate()
	case "blob", "slime", "amoeba", "ooze":
		return BlobTemplate()
	case "mechanical", "robot", "golem", "construct", "android":
		return MechanicalTemplate()
	case "flying", "bird", "dragon", "bat", "wyvern":
		return FlyingTemplate()
	case "serpentine", "snake", "worm", "tentacle", "wyrm":
		return SerpentineTemplate()
	case "arachnid", "spider", "insect", "beetle":
		return ArachnidTemplate()
	case "undead", "skeleton", "ghost", "zombie", "lich":
		return UndeadTemplate()
	default:
		// Default to humanoid for unknown types
		return HumanoidTemplate()
	}
}

// SerpentineTemplate returns a template for snake-like creatures.
// Optimized for 32x32 pixels (snakes, worms, tentacles).
func SerpentineTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "serpentine",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.88,
				RelativeWidth:  0.70,
				RelativeHeight: 0.18,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.35,
				Rotation:       0,
			},
			// Use legs part for tail/lower body segment
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.80,
				RelativeWidth:  0.25,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeEllipse, shapes.ShapeWave},
				ZIndex:         5,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			// Main body (elongated)
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.35,
				RelativeHeight: 0.70,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeBean, shapes.ShapeWave},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			// Head at top
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.20,
				RelativeWidth:  0.30,
				RelativeHeight: 0.28,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeEllipse, shapes.ShapeTriangle},
				ZIndex:         15,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       0,
			},
		},
	}
}

// ArachnidTemplate returns a template for spider-like creatures.
// Optimized for 32x32 pixels (spiders, insects with 6-8 legs).
func ArachnidTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "arachnid",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.90,
				RelativeWidth:  0.75,
				RelativeHeight: 0.18,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.3,
				Rotation:       0,
			},
			// Legs (spread wide for multi-leg appearance)
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.70,
				RelativeWidth:  0.85,
				RelativeHeight: 0.35,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeLightning},
				ZIndex:         5,
				ColorRole:      "primary",
				Opacity:        0.95,
				Rotation:       0,
			},
			// Central body (small, oval)
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.45,
				RelativeWidth:  0.50,
				RelativeHeight: 0.55,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCircle},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        1.0,
				Rotation:       0,
			},
			// Head/fangs (smaller, forward)
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.35,
				RelativeHeight: 0.25,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeWedge, shapes.ShapeEllipse},
				ZIndex:         12,
				ColorRole:      "secondary",
				Opacity:        1.0,
				Rotation:       0,
			},
			// Additional leg detail using arms slot
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.90,
				RelativeHeight: 0.28,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeLightning, shapes.ShapeCapsule},
				ZIndex:         8,
				ColorRole:      "primary",
				Opacity:        0.9,
				Rotation:       15,
			},
		},
	}
}

// UndeadTemplate returns a template for undead creatures.
// Optimized for 32x32 pixels (skeletons, ghosts, zombies).
func UndeadTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "undead",
		BodyPartLayout: map[BodyPart]PartSpec{
			PartShadow: {
				RelativeX:      0.5,
				RelativeY:      0.92,
				RelativeWidth:  0.35,
				RelativeHeight: 0.12,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
				ZIndex:         0,
				ColorRole:      "shadow",
				Opacity:        0.2, // Fainter shadow for undead
				Rotation:       0,
			},
			// Skeletal legs (thin)
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.28,
				RelativeHeight: 0.38,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:         5,
				ColorRole:      "primary",
				Opacity:        0.85, // Slightly translucent
				Rotation:       0,
			},
			// Ribcage/torso (gaunt)
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.48,
				RelativeWidth:  0.42,
				RelativeHeight: 0.48,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeEllipse, shapes.ShapeOrganic},
				ZIndex:         10,
				ColorRole:      "primary",
				Opacity:        0.85,
				Rotation:       0,
			},
			// Bony arms (thin, angular)
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.48,
				RelativeWidth:  0.62,
				RelativeHeight: 0.32,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:         8,
				ColorRole:      "secondary",
				Opacity:        0.85,
				Rotation:       0,
			},
			// Skull head
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.22,
				RelativeWidth:  0.38,
				RelativeHeight: 0.38,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeCircle},
				ZIndex:         15,
				ColorRole:      "secondary",
				Opacity:        0.85,
				Rotation:       0,
			},
		},
	}
}

// BossTemplate returns a scaled-up version of any template for boss enemies.
// Scale should be 2.0-4.0 for bosses (2x to 4x larger than normal).
func BossTemplate(baseTemplate AnatomicalTemplate, scale float64) AnatomicalTemplate {
	boss := AnatomicalTemplate{
		Name:           "boss_" + baseTemplate.Name,
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Copy and scale all body parts
	for part, spec := range baseTemplate.BodyPartLayout {
		bossSpec := spec
		// Scale dimensions (but keep relative positions)
		bossSpec.RelativeWidth *= scale
		bossSpec.RelativeHeight *= scale

		// For bosses, enhance opacity and add slight size variations
		if part == PartTorso || part == PartHead {
			bossSpec.Opacity = 1.0 // Full opacity for prominent parts
		}

		boss.BodyPartLayout[part] = bossSpec
	}

	return boss
}

// ApplyBossEnhancements adds additional detail to boss sprites.
// This includes armor plates, spikes, or other prominent features.
func ApplyBossEnhancements(template AnatomicalTemplate) AnatomicalTemplate {
	enhanced := template
	enhanced.Name = "enhanced_" + template.Name

	// Add armor plating if torso exists
	if torsoSpec, hasTorso := enhanced.BodyPartLayout[PartTorso]; hasTorso {
		armorSpec := torsoSpec
		armorSpec.RelativeWidth *= 1.15 // Slightly larger than torso
		armorSpec.RelativeHeight *= 1.15
		armorSpec.ZIndex = torsoSpec.ZIndex - 1 // Behind torso
		armorSpec.ColorRole = "accent3"
		armorSpec.Opacity = 0.8
		armorSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle}
		enhanced.BodyPartLayout[PartArmor] = armorSpec
	}

	return enhanced
}

// HumanoidDirectionalTemplate returns a humanoid template with directional facing.
// Direction should be provided in Config.Custom["facing"] ("up", "down", "left", "right").
// This creates asymmetry to indicate facing direction.
func HumanoidDirectionalTemplate(direction Direction) AnatomicalTemplate {
	base := HumanoidTemplate()
	base.Name = "humanoid_" + string(direction)

	switch direction {
	case DirUp:
		// Facing away - head at top, arms spread slightly
		base.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.48,
			RelativeWidth:  0.70,
			RelativeHeight: 0.30,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         8,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}

	case DirDown:
		// Facing toward viewer - head at top, arms slightly forward
		base.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.52,
			RelativeWidth:  0.60,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         12, // Arms in front
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}

	case DirLeft:
		// Facing left - asymmetric arm positioning
		base.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.45,
			RelativeY:      0.50,
			RelativeWidth:  0.40,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         8,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       270,
		}
		// Shift head slightly left
		headSpec := base.BodyPartLayout[PartHead]
		headSpec.RelativeX = 0.45
		base.BodyPartLayout[PartHead] = headSpec

	case DirRight:
		// Facing right - asymmetric arm positioning
		base.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.55,
			RelativeY:      0.50,
			RelativeWidth:  0.40,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         8,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       90,
		}
		// Shift head slightly right
		headSpec := base.BodyPartLayout[PartHead]
		headSpec.RelativeX = 0.55
		base.BodyPartLayout[PartHead] = headSpec
	}

	return base
}

// HumanoidWithEquipment returns a humanoid template with weapon and shield positioning.
// hasWeapon and hasShield should be provided in Config.Custom.
func HumanoidWithEquipment(direction Direction, hasWeapon, hasShield bool) AnatomicalTemplate {
	base := HumanoidDirectionalTemplate(direction)
	base.Name = "humanoid_equipped_" + string(direction)

	if hasWeapon {
		var weaponSpec PartSpec
		switch direction {
		case DirUp:
			weaponSpec = PartSpec{
				RelativeX:      0.65,
				RelativeY:      0.50,
				RelativeWidth:  0.15,
				RelativeHeight: 0.40,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBlade, shapes.ShapeRectangle},
				ZIndex:         7,
				ColorRole:      "accent1",
				Opacity:        1.0,
				Rotation:       45,
			}
		case DirDown:
			weaponSpec = PartSpec{
				RelativeX:      0.70,
				RelativeY:      0.55,
				RelativeWidth:  0.15,
				RelativeHeight: 0.40,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBlade, shapes.ShapeRectangle},
				ZIndex:         13,
				ColorRole:      "accent1",
				Opacity:        1.0,
				Rotation:       135,
			}
		case DirLeft:
			weaponSpec = PartSpec{
				RelativeX:      0.30,
				RelativeY:      0.50,
				RelativeWidth:  0.40,
				RelativeHeight: 0.15,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBlade, shapes.ShapeRectangle},
				ZIndex:         7,
				ColorRole:      "accent1",
				Opacity:        1.0,
				Rotation:       270,
			}
		case DirRight:
			weaponSpec = PartSpec{
				RelativeX:      0.70,
				RelativeY:      0.50,
				RelativeWidth:  0.40,
				RelativeHeight: 0.15,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeBlade, shapes.ShapeRectangle},
				ZIndex:         7,
				ColorRole:      "accent1",
				Opacity:        1.0,
				Rotation:       90,
			}
		}
		base.BodyPartLayout[PartWeapon] = weaponSpec
	}

	if hasShield {
		var shieldSpec PartSpec
		switch direction {
		case DirUp:
			shieldSpec = PartSpec{
				RelativeX:      0.35,
				RelativeY:      0.50,
				RelativeWidth:  0.25,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeShield, shapes.ShapeCircle},
				ZIndex:         7,
				ColorRole:      "accent2",
				Opacity:        1.0,
				Rotation:       0,
			}
		case DirDown:
			shieldSpec = PartSpec{
				RelativeX:      0.30,
				RelativeY:      0.55,
				RelativeWidth:  0.25,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeShield, shapes.ShapeCircle},
				ZIndex:         13,
				ColorRole:      "accent2",
				Opacity:        1.0,
				Rotation:       0,
			}
		case DirLeft:
			shieldSpec = PartSpec{
				RelativeX:      0.25,
				RelativeY:      0.50,
				RelativeWidth:  0.20,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeShield, shapes.ShapeCircle},
				ZIndex:         13,
				ColorRole:      "accent2",
				Opacity:        1.0,
				Rotation:       270,
			}
		case DirRight:
			shieldSpec = PartSpec{
				RelativeX:      0.75,
				RelativeY:      0.50,
				RelativeWidth:  0.20,
				RelativeHeight: 0.30,
				ShapeTypes:     []shapes.ShapeType{shapes.ShapeShield, shapes.ShapeCircle},
				ZIndex:         7,
				ColorRole:      "accent2",
				Opacity:        1.0,
				Rotation:       90,
			}
		}
		base.BodyPartLayout[PartShield] = shieldSpec
	}

	return base
}

// FantasyHumanoidTemplate returns a humanoid with fantasy-specific proportions.
// Broader shoulders, medieval aesthetic.
func FantasyHumanoidTemplate(direction Direction) AnatomicalTemplate {
	base := HumanoidDirectionalTemplate(direction)
	base.Name = "fantasy_humanoid_" + string(direction)

	// Broader shoulders
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.RelativeWidth = 0.55
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeRectangle}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Thicker limbs
	armsSpec := base.BodyPartLayout[PartArms]
	armsSpec.RelativeWidth = 0.70
	armsSpec.RelativeHeight = 0.38
	base.BodyPartLayout[PartArms] = armsSpec

	legsSpec := base.BodyPartLayout[PartLegs]
	legsSpec.RelativeWidth = 0.40
	base.BodyPartLayout[PartLegs] = legsSpec

	return base
}

// SciFiHumanoidTemplate returns a humanoid with sci-fi aesthetic.
// Angular features, sleek profile, helmet shapes.
func SciFiHumanoidTemplate(direction Direction) AnatomicalTemplate {
	base := HumanoidDirectionalTemplate(direction)
	base.Name = "scifi_humanoid_" + string(direction)

	// Angular torso
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Sleeker limbs
	armsSpec := base.BodyPartLayout[PartArms]
	armsSpec.RelativeWidth = 0.60
	armsSpec.RelativeHeight = 0.32
	base.BodyPartLayout[PartArms] = armsSpec

	// Helmet-like head
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle}
	headSpec.RelativeWidth = 0.38
	headSpec.RelativeHeight = 0.38
	base.BodyPartLayout[PartHead] = headSpec

	return base
}

// HorrorHumanoidTemplate returns a humanoid with horror aesthetic.
// Distorted proportions, unnatural shapes.
func HorrorHumanoidTemplate(direction Direction) AnatomicalTemplate {
	base := HumanoidDirectionalTemplate(direction)
	base.Name = "horror_humanoid_" + string(direction)

	// Elongated head
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.RelativeHeight = 0.42
	headSpec.RelativeWidth = 0.30
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeEllipse}
	base.BodyPartLayout[PartHead] = headSpec

	// Thin, elongated limbs
	armsSpec := base.BodyPartLayout[PartArms]
	armsSpec.RelativeHeight = 0.45
	armsSpec.RelativeWidth = 0.55
	base.BodyPartLayout[PartArms] = armsSpec

	legsSpec := base.BodyPartLayout[PartLegs]
	legsSpec.RelativeHeight = 0.40
	legsSpec.RelativeWidth = 0.28
	base.BodyPartLayout[PartLegs] = legsSpec

	// Distorted torso
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.RelativeWidth = 0.45
	torsoSpec.RelativeHeight = 0.50
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeBean}
	base.BodyPartLayout[PartTorso] = torsoSpec

	return base
}

// CyberpunkHumanoidTemplate returns a humanoid with cyberpunk aesthetic.
// Compact build, angular limbs, tech implants.
func CyberpunkHumanoidTemplate(direction Direction) AnatomicalTemplate {
	base := HumanoidDirectionalTemplate(direction)
	base.Name = "cyberpunk_humanoid_" + string(direction)

	// Compact torso
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.RelativeWidth = 0.48
	torsoSpec.RelativeHeight = 0.42
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeRectangle}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Angular limbs
	armsSpec := base.BodyPartLayout[PartArms]
	armsSpec.RelativeWidth = 0.62
	armsSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule}
	base.BodyPartLayout[PartArms] = armsSpec

	// Helmeted head
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOctagon, shapes.ShapeHexagon}
	headSpec.RelativeWidth = 0.36
	headSpec.ColorRole = "accent1" // Tech glow
	base.BodyPartLayout[PartHead] = headSpec

	return base
}

// PostApocHumanoidTemplate returns a humanoid with post-apocalyptic aesthetic.
// Rough edges, tattered appearance, improvised equipment.
func PostApocHumanoidTemplate(direction Direction) AnatomicalTemplate {
	base := HumanoidDirectionalTemplate(direction)
	base.Name = "postapoc_humanoid_" + string(direction)

	// Irregular torso (tattered clothing)
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeBean, shapes.ShapeRectangle}
	torsoSpec.RelativeWidth = 0.52
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Rough limbs
	armsSpec := base.BodyPartLayout[PartArms]
	armsSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule}
	base.BodyPartLayout[PartArms] = armsSpec

	legsSpec := base.BodyPartLayout[PartLegs]
	legsSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule}
	base.BodyPartLayout[PartLegs] = legsSpec

	// Covered head (masks, hoods)
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeOrganic, shapes.ShapeSkull}
	base.BodyPartLayout[PartHead] = headSpec

	return base
}

// SelectHumanoidTemplate chooses an appropriate humanoid template based on genre and direction.
// Returns directional humanoid with genre-specific styling.
func SelectHumanoidTemplate(genre, entityType string, direction Direction) AnatomicalTemplate {
	// Check if this is a humanoid type
	isHumanoid := false
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior":
		isHumanoid = true
	}

	if !isHumanoid {
		return SelectTemplate(entityType)
	}

	// Apply genre-specific styling
	switch genre {
	case "fantasy":
		return FantasyHumanoidTemplate(direction)
	case "scifi", "sci-fi":
		return SciFiHumanoidTemplate(direction)
	case "horror":
		return HorrorHumanoidTemplate(direction)
	case "cyberpunk":
		return CyberpunkHumanoidTemplate(direction)
	case "postapoc", "post-apocalyptic":
		return PostApocHumanoidTemplate(direction)
	default:
		return HumanoidDirectionalTemplate(direction)
	}
}
