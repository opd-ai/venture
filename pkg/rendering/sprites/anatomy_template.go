//go:build !test
// +build !test

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
	case "flying", "bird", "dragon", "bat", "wyrm":
		return FlyingTemplate()
	default:
		// Default to humanoid for unknown types
		return HumanoidTemplate()
	}
}
