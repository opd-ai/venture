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
	// PartEyes represents facial eyes (Phase 15.1: 2px detail)
	PartEyes
	// PartMouth represents facial mouth (Phase 15.1: 1-2px detail)
	PartMouth
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
	case PartEyes:
		return "eyes"
	case PartMouth:
		return "mouth"
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

// PixelDimensions specifies exact pixel dimensions for a body part.
// This enables enhanced detail control for Phase 15.1 sub-pixel rendering.
// Example: head 4×4, torso 4×6, legs 4×8 pixels for humanoid characters.
type PixelDimensions struct {
	// Width in pixels
	Width int
	// Height in pixels
	Height int
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
	// PreferredPixelSize optionally specifies exact pixel dimensions for enhanced detail.
	// When set, GetEffectiveWidth() and GetEffectiveHeight() use these exact dimensions
	// instead of calculating from RelativeWidth/RelativeHeight, enabling pixel-perfect control.
	// If nil, RelativeWidth/RelativeHeight are used for calculation.
	// Phase 15.1: Enables "head 4×4, torso 4×6, legs 4×8" specification.
	PreferredPixelSize *PixelDimensions
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

// EnhancedHumanoidTemplate returns a Phase 15.1 enhanced humanoid template
// with pixel-perfect dimensions for improved anatomical accuracy.
// Uses exact pixel specifications: head 4×4, torso 4×6, legs 4×8 pixels.
// Optimized for 28x28 pixel sprites with 40% more anatomical detail.
//
// This template demonstrates Phase 15.1 enhanced proportional scaling,
// providing clearer silhouettes and better player recognition at a glance.
func EnhancedHumanoidTemplate() AnatomicalTemplate {
	return AnatomicalTemplate{
		Name: "enhanced_humanoid",
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.286, // 8/28 for 8 pixel height on 28px sprite
				RelativeHeight: 0.286,
				PreferredPixelSize: &PixelDimensions{
					Width:  4,
					Height: 8, // Phase 15.1: 4×8 pixel legs
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:     5,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.214, // 6/28 for 6 pixel height on 28px sprite
				RelativeHeight: 0.214,
				PreferredPixelSize: &PixelDimensions{
					Width:  4,
					Height: 6, // Phase 15.1: 4×6 pixel torso
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeRectangle, shapes.ShapeEllipse},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.214, // Arms proportional to torso
				RelativeHeight: 0.179, // 5/28 for arm length
				PreferredPixelSize: &PixelDimensions{
					Width:  6, // Slightly wider for arm reach
					Height: 5,
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
				ZIndex:     8,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.143, // 4/28 for 4 pixel dimensions on 28px sprite
				RelativeHeight: 0.143,
				PreferredPixelSize: &PixelDimensions{
					Width:  4,
					Height: 4, // Phase 15.1: 4×4 pixel head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse, shapes.ShapeSkull},
				ZIndex:     15,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   0,
			},
		},
	}
}

// DetailedHumanoidTemplate returns a Phase 15.1 template with facial features.
// Includes eyes (2px) and mouth (1-2px) for close-up views and enhanced recognition.
// Builds on EnhancedHumanoidTemplate with additional facial detail.
//
// Use this template for:
// - Player characters that are frequently on-screen
// - NPCs with whom players interact closely
// - Character portraits and close-up views
// - Situations requiring clear emotional expression
func DetailedHumanoidTemplate() AnatomicalTemplate {
	// Start with the enhanced template
	base := EnhancedHumanoidTemplate()
	base.Name = "detailed_humanoid"

	// Add facial features with pixel-perfect dimensions
	// Eyes: 2 pixels wide, positioned on upper head
	base.BodyPartLayout[PartEyes] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.23,  // Slightly above head center
		RelativeWidth:  0.071, // 2/28
		RelativeHeight: 0.036, // 1/28 (height)
		PreferredPixelSize: &PixelDimensions{
			Width:  2,
			Height: 1, // Phase 15.1: 2×1 pixel eyes
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		ZIndex:     16, // Above head
		ColorRole:  "accent1",
		Opacity:    1.0,
		Rotation:   0,
	}

	// Mouth: 1-2 pixels, positioned on lower head
	base.BodyPartLayout[PartMouth] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.27,  // Below eyes
		RelativeWidth:  0.071, // 2/28
		RelativeHeight: 0.036, // 1/28
		PreferredPixelSize: &PixelDimensions{
			Width:  2,
			Height: 1, // Phase 15.1: 2×1 pixel mouth
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
		ZIndex:     16, // Above head, same as eyes
		ColorRole:  "accent2",
		Opacity:    1.0,
		Rotation:   0,
	}

	return base
}

// QuadrupedTemplate returns a template for four-legged creatures.
// Optimized for 32x32 pixels. Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.313, // 10/32 for 10 pixel width
				RelativeHeight: 0.125, // 4/32 for 4 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  10,
					Height: 4, // Phase 15.1: 10×4 pixel legs (horizontal orientation)
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:     5,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   90, // Horizontal orientation
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.313, // 10/32 for 10 pixel width
				RelativeHeight: 0.219, // 7/32 for 7 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  10,
					Height: 7, // Phase 15.1: 10×7 pixel torso (horizontal body)
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   90, // Horizontal body
			},
			PartHead: {
				RelativeX:      0.25,
				RelativeY:      0.35,
				RelativeWidth:  0.156, // 5/32 for 5 pixel width
				RelativeHeight: 0.188, // 6/32 for 6 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  5,
					Height: 6, // Phase 15.1: 5×6 pixel head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse, shapes.ShapeWedge},
				ZIndex:     15,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   270, // Face left
			},
		},
	}
}

// BlobTemplate returns a template for amorphous creatures.
// Optimized for 32x32 pixels (slimes, amoebas). Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.55,
				RelativeWidth:  0.500, // 16/32 for 16 pixel width
				RelativeHeight: 0.438, // 14/32 for 14 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  16,
					Height: 14, // Phase 15.1: 16×14 pixel blob core (large amorphous mass)
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeCircle, shapes.ShapeBean},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    0.9, // Slightly translucent
				Rotation:   0,
			},
		},
	}
}

// MechanicalTemplate returns a template for robots and constructs.
// Optimized for 32x32 pixels (robots, golems). Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.188, // 6/32 for 6 pixel width
				RelativeHeight: 0.188, // 6/32 for 6 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  6,
					Height: 6, // Phase 15.1: 6×6 pixel mechanical legs
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
				ZIndex:     5,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.250, // 8/32 for 8 pixel width
				RelativeHeight: 0.219, // 7/32 for 7 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  8,
					Height: 7, // Phase 15.1: 8×7 pixel mechanical chassis
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon, shapes.ShapeOctagon},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.281, // 9/32 for 9 pixel width
				RelativeHeight: 0.156, // 5/32 for 5 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  9,
					Height: 5, // Phase 15.1: 9×5 pixel mechanical arms
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
				ZIndex:     8,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.188, // 6/32 for 6 pixel width
				RelativeHeight: 0.156, // 5/32 for 5 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  6,
					Height: 5, // Phase 15.1: 6×5 pixel mechanical head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon, shapes.ShapeOctagon},
				ZIndex:     15,
				ColorRole:  "accent1",
				Opacity:    1.0,
				Rotation:   0,
			},
		},
	}
}

// FlyingTemplate returns a template for winged creatures.
// Optimized for 32x32 pixels (birds, dragons, flying enemies). Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			// Left wing (behind body)
			PartLegs: { // Reuse legs part for left wing
				RelativeX:      0.25,
				RelativeY:      0.50,
				RelativeWidth:  0.219, // 7/32 for 7 pixel width
				RelativeHeight: 0.188, // 6/32 for 6 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  7,
					Height: 6, // Phase 15.1: 7×6 pixel left wing
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle},
				ZIndex:     5,
				ColorRole:  "secondary",
				Opacity:    0.9,
				Rotation:   270, // Point left
			},
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.55,
				RelativeWidth:  0.219, // 7/32 for 7 pixel width
				RelativeHeight: 0.281, // 9/32 for 9 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  7,
					Height: 9, // Phase 15.1: 7×9 pixel flying body
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.30,
				RelativeWidth:  0.156, // 5/32 for 5 pixel width
				RelativeHeight: 0.156, // 5/32 for 5 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  5,
					Height: 5, // Phase 15.1: 5×5 pixel flying creature head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
				ZIndex:     12,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   0,
			},
			// Right wing (in front of body)
			PartArms: { // Reuse arms part for right wing
				RelativeX:      0.75,
				RelativeY:      0.50,
				RelativeWidth:  0.219, // 7/32 for 7 pixel width
				RelativeHeight: 0.188, // 6/32 for 6 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  7,
					Height: 6, // Phase 15.1: 7×6 pixel right wing
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle},
				ZIndex:     15,
				ColorRole:  "secondary",
				Opacity:    0.9,
				Rotation:   90, // Point right
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
// Optimized for 32x32 pixels (snakes, worms, tentacles). Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			// Use legs part for tail/lower body segment
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.80,
				RelativeWidth:  0.156, // 5/32 for 5 pixel width
				RelativeHeight: 0.219, // 7/32 for 7 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  5,
					Height: 7, // Phase 15.1: 5×7 pixel tail segment
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeEllipse, shapes.ShapeWave},
				ZIndex:     5,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			// Main body (elongated)
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.188, // 6/32 for 6 pixel width
				RelativeHeight: 0.406, // 13/32 for 13 pixel height (elongated)
				PreferredPixelSize: &PixelDimensions{
					Width:  6,
					Height: 13, // Phase 15.1: 6×13 pixel serpentine body
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeBean, shapes.ShapeWave},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			// Head at top
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.20,
				RelativeWidth:  0.156, // 5/32 for 5 pixel width
				RelativeHeight: 0.156, // 5/32 for 5 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  5,
					Height: 5, // Phase 15.1: 5×5 pixel serpent head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeEllipse, shapes.ShapeTriangle},
				ZIndex:     15,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   0,
			},
		},
	}
}

// ArachnidTemplate returns a template for spider-like creatures.
// Optimized for 32x32 pixels (spiders, insects with 6-8 legs). Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			// Legs (spread wide for multi-leg appearance)
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.70,
				RelativeWidth:  0.438, // 14/32 for 14 pixel width
				RelativeHeight: 0.188, // 6/32 for 6 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  14,
					Height: 6, // Phase 15.1: 14×6 pixel legs (wide spread)
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeLightning},
				ZIndex:     5,
				ColorRole:  "primary",
				Opacity:    0.95,
				Rotation:   0,
			},
			// Central body (small, oval)
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.45,
				RelativeWidth:  0.250, // 8/32 for 8 pixel width
				RelativeHeight: 0.281, // 9/32 for 9 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  8,
					Height: 9, // Phase 15.1: 8×9 pixel arachnid body
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCircle},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    1.0,
				Rotation:   0,
			},
			// Head/fangs (smaller, forward)
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.156, // 5/32 for 5 pixel width
				RelativeHeight: 0.125, // 4/32 for 4 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  5,
					Height: 4, // Phase 15.1: 5×4 pixel arachnid head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeWedge, shapes.ShapeEllipse},
				ZIndex:     12,
				ColorRole:  "secondary",
				Opacity:    1.0,
				Rotation:   0,
			},
			// Additional leg detail using arms slot
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.50,
				RelativeWidth:  0.469, // 15/32 for 15 pixel width
				RelativeHeight: 0.156, // 5/32 for 5 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  15,
					Height: 5, // Phase 15.1: 15×5 pixel upper legs
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeLightning, shapes.ShapeCapsule},
				ZIndex:     8,
				ColorRole:  "primary",
				Opacity:    0.9,
				Rotation:   15,
			},
		},
	}
}

// UndeadTemplate returns a template for undead creatures.
// Optimized for 32x32 pixels (skeletons, ghosts, zombies). Phase 15.1: Enhanced with pixel-perfect dimensions.
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
				// Shadow doesn't use pixel dimensions - scales with sprite
			},
			// Skeletal legs (thin)
			PartLegs: {
				RelativeX:      0.5,
				RelativeY:      0.75,
				RelativeWidth:  0.156, // 5/32 for 5 pixel width
				RelativeHeight: 0.219, // 7/32 for 7 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  5,
					Height: 7, // Phase 15.1: 5×7 pixel skeletal legs (thin)
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:     5,
				ColorRole:  "primary",
				Opacity:    0.85, // Slightly translucent
				Rotation:   0,
			},
			// Ribcage/torso (gaunt)
			PartTorso: {
				RelativeX:      0.5,
				RelativeY:      0.48,
				RelativeWidth:  0.219, // 7/32 for 7 pixel width
				RelativeHeight: 0.250, // 8/32 for 8 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  7,
					Height: 8, // Phase 15.1: 7×8 pixel skeletal torso
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeEllipse, shapes.ShapeOrganic},
				ZIndex:     10,
				ColorRole:  "primary",
				Opacity:    0.85,
				Rotation:   0,
			},
			// Bony arms (thin, angular)
			PartArms: {
				RelativeX:      0.5,
				RelativeY:      0.48,
				RelativeWidth:  0.281, // 9/32 for 9 pixel width
				RelativeHeight: 0.156, // 5/32 for 5 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  9,
					Height: 5, // Phase 15.1: 9×5 pixel skeletal arms
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
				ZIndex:     8,
				ColorRole:  "secondary",
				Opacity:    0.85,
				Rotation:   0,
			},
			// Skull head
			PartHead: {
				RelativeX:      0.5,
				RelativeY:      0.22,
				RelativeWidth:  0.188, // 6/32 for 6 pixel width
				RelativeHeight: 0.188, // 6/32 for 6 pixel height
				PreferredPixelSize: &PixelDimensions{
					Width:  6,
					Height: 6, // Phase 15.1: 6×6 pixel skull head
				},
				ShapeTypes: []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeCircle},
				ZIndex:     15,
				ColorRole:  "secondary",
				Opacity:    0.85,
				Rotation:   0,
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

// HumanoidAerialTemplate returns a humanoid template optimized for aerial/top-down perspective.
// Proportions are adjusted for top-down view: head 35%, torso 50%, legs 15% (compressed vertical).
// Creates visual asymmetry based on facing direction for clear directional indication.
//
// Key differences from side-view:
//   - Head more prominent (35% vs 30%) - more visible from above
//   - Torso compressed vertically but wider horizontally (50% vs 40%)
//   - Legs minimally visible (15% vs 30%) - mostly obscured from top-down view
//   - Shadow ellipse at base for depth perception
//   - Directional asymmetry in head position and arm visibility
func HumanoidAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "humanoid_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Shadow - ellipse at base for depth perception
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.90,
		RelativeWidth:  0.50,
		RelativeHeight: 0.15,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.35,
		Rotation:       0,
	}

	// Legs - minimal visibility from top-down, compressed
	template.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.80,
		RelativeWidth:  0.35,
		RelativeHeight: 0.15,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCapsule},
		ZIndex:         5,
		ColorRole:      "primary",
		Opacity:        0.8,
		Rotation:       0,
	}

	// Torso - wider horizontally, compressed vertically for aerial view
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.60,
		RelativeHeight: 0.50,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean, shapes.ShapeRectangle},
		ZIndex:         10,
		ColorRole:      "primary",
		Opacity:        1.0,
		Rotation:       0,
	}

	// Direction-specific head and arm positioning for visual asymmetry
	switch direction {
	case DirUp:
		// Facing away - head centered, arms symmetrical
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.20,
			RelativeWidth:  0.35,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.50,
			RelativeWidth:  0.70,
			RelativeHeight: 0.25,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         8, // Arms behind torso when facing up
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}

	case DirDown:
		// Facing toward viewer - head centered, arms asymmetric (forward reach)
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.20,
			RelativeWidth:  0.35,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.52,
			RelativeWidth:  0.65,
			RelativeHeight: 0.30,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         12, // Arms in front of torso when facing down
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}

	case DirLeft:
		// Facing left - head shifted left, left arm visible
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.42,
			RelativeY:      0.20,
			RelativeWidth:  0.35,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.35,
			RelativeY:      0.50,
			RelativeWidth:  0.35,
			RelativeHeight: 0.28,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         8,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       270,
		}

	case DirRight:
		// Facing right - head shifted right, right arm visible
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.58,
			RelativeY:      0.20,
			RelativeWidth:  0.35,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.65,
			RelativeY:      0.50,
			RelativeWidth:  0.35,
			RelativeHeight: 0.28,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:         8,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       90,
		}
	}

	return template
}

// EnhancedHumanoidAerialTemplate returns a Phase 15.1 enhanced aerial-view humanoid template
// with pixel-perfect dimensions for improved anatomical accuracy from top-down perspective.
// Uses exact pixel specifications optimized for aerial view: head 5×5, torso 7×6, legs 4×2 pixels.
// Optimized for 28x28 pixel sprites with 40% more anatomical detail.
//
// This template demonstrates Phase 15.1 enhanced proportional scaling for aerial views,
// providing clearer silhouettes and better entity recognition from above.
func EnhancedHumanoidAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "enhanced_humanoid_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Shadow - ellipse at base for depth perception (no pixel dimensions - scales with sprite)
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.90,
		RelativeWidth:  0.50,
		RelativeHeight: 0.15,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.35,
		Rotation:       0,
	}

	// Legs - minimal visibility from top-down, compressed
	template.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.80,
		RelativeWidth:  0.143, // 4/28 for 4 pixel width on 28px sprite
		RelativeHeight: 0.071, // 2/28 for 2 pixel height
		PreferredPixelSize: &PixelDimensions{
			Width:  4,
			Height: 2, // Phase 15.1: 4×2 pixel legs (minimal from aerial view)
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCapsule},
		ZIndex:     5,
		ColorRole:  "primary",
		Opacity:    0.8,
		Rotation:   0,
	}

	// Torso - wider horizontally, compressed vertically for aerial view
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.25,  // 7/28 for 7 pixel width
		RelativeHeight: 0.214, // 6/28 for 6 pixel height
		PreferredPixelSize: &PixelDimensions{
			Width:  7,
			Height: 6, // Phase 15.1: 7×6 pixel torso (wider for aerial view)
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean, shapes.ShapeRectangle},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    1.0,
		Rotation:   0,
	}

	// Direction-specific head and arm positioning for visual asymmetry
	switch direction {
	case DirUp:
		// Facing away - head centered, arms symmetrical
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.20,
			RelativeWidth:  0.179, // 5/28 for 5 pixel dimensions
			RelativeHeight: 0.179,
			PreferredPixelSize: &PixelDimensions{
				Width:  5,
				Height: 5, // Phase 15.1: 5×5 pixel head (larger for aerial visibility)
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:     15,
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.50,
			RelativeWidth:  0.214, // 6/28 for arm width
			RelativeHeight: 0.107, // 3/28 for arm height
			PreferredPixelSize: &PixelDimensions{
				Width:  6,
				Height: 3, // Phase 15.1: 6×3 pixel arms
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:     8, // Arms behind torso when facing up
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   0,
		}

	case DirDown:
		// Facing toward viewer - head centered, arms asymmetric (forward reach)
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.20,
			RelativeWidth:  0.179, // 5/28 for 5 pixel dimensions
			RelativeHeight: 0.179,
			PreferredPixelSize: &PixelDimensions{
				Width:  5,
				Height: 5, // Phase 15.1: 5×5 pixel head
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:     15,
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.52,
			RelativeWidth:  0.214, // 6/28 for arm width
			RelativeHeight: 0.107, // 3/28 for arm height
			PreferredPixelSize: &PixelDimensions{
				Width:  6,
				Height: 3, // Phase 15.1: 6×3 pixel arms
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:     12, // Arms in front of torso when facing down
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   0,
		}

	case DirLeft:
		// Facing left - head shifted left, left arm visible
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.42,
			RelativeY:      0.20,
			RelativeWidth:  0.179, // 5/28 for 5 pixel dimensions
			RelativeHeight: 0.179,
			PreferredPixelSize: &PixelDimensions{
				Width:  5,
				Height: 5, // Phase 15.1: 5×5 pixel head
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:     15,
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.35,
			RelativeY:      0.50,
			RelativeWidth:  0.107, // 3/28 for arm width when viewed from side
			RelativeHeight: 0.143, // 4/28 for arm height
			PreferredPixelSize: &PixelDimensions{
				Width:  3,
				Height: 4, // Phase 15.1: 3×4 pixel arms (narrower from side view)
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:     8,
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   270,
		}

	case DirRight:
		// Facing right - head shifted right, right arm visible
		template.BodyPartLayout[PartHead] = PartSpec{
			RelativeX:      0.58,
			RelativeY:      0.20,
			RelativeWidth:  0.179, // 5/28 for 5 pixel dimensions
			RelativeHeight: 0.179,
			PreferredPixelSize: &PixelDimensions{
				Width:  5,
				Height: 5, // Phase 15.1: 5×5 pixel head
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:     15,
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   0,
		}
		template.BodyPartLayout[PartArms] = PartSpec{
			RelativeX:      0.65,
			RelativeY:      0.50,
			RelativeWidth:  0.107, // 3/28 for arm width when viewed from side
			RelativeHeight: 0.143, // 4/28 for arm height
			PreferredPixelSize: &PixelDimensions{
				Width:  3,
				Height: 4, // Phase 15.1: 3×4 pixel arms (narrower from side view)
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
			ZIndex:     8,
			ColorRole:  "secondary",
			Opacity:    1.0,
			Rotation:   90,
		}
	}

	return template
}

// FantasyHumanoidAerial returns a fantasy-themed aerial humanoid template.
// Broader shoulders, visible helmet, cape shadow for medieval fantasy aesthetic.
// Phase 15.1: Uses enhanced base with pixel-perfect dimensions.
func FantasyHumanoidAerial(direction Direction) AnatomicalTemplate {
	base := EnhancedHumanoidAerialTemplate(direction)
	base.Name = "fantasy_aerial_" + string(direction)

	// Broader shoulders for armored knight appearance
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.RelativeWidth = 0.286  // 8/28 for 8 pixel width
	torsoSpec.RelativeHeight = 0.214 // Keep 6/28 for height
	torsoSpec.PreferredPixelSize = &PixelDimensions{
		Width:  8,
		Height: 6, // Phase 15.1: 8×6 pixel torso (broader for armor)
	}
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeBean, shapes.ShapeRectangle}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Add helmet shape on head (slightly larger for visibility)
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeCircle, shapes.ShapeOctagon}
	headSpec.RelativeWidth = 0.214  // 6/28 for 6 pixel dimensions
	headSpec.RelativeHeight = 0.214 // Helmet is slightly taller
	headSpec.PreferredPixelSize = &PixelDimensions{
		Width:  6,
		Height: 6, // Phase 15.1: 6×6 pixel helmeted head
	}
	base.BodyPartLayout[PartHead] = headSpec

	// Thicker arms for armored appearance (direction-dependent)
	armsSpec := base.BodyPartLayout[PartArms]
	// For up/down directions, arms are wider
	if direction == DirUp || direction == DirDown {
		armsSpec.RelativeHeight = 0.143 // 4/28 for thicker arms
		if armsSpec.PreferredPixelSize != nil {
			armsSpec.PreferredPixelSize = &PixelDimensions{
				Width:  armsSpec.PreferredPixelSize.Width,
				Height: 4, // Phase 15.1: Thicker arms for armor
			}
		}
	}
	base.BodyPartLayout[PartArms] = armsSpec

	return base
}

// SciFiHumanoidAerial returns a sci-fi themed aerial humanoid template.
// Angular shapes, glowing accents, jetpack indicator for futuristic aesthetic.
// Phase 15.1: Uses enhanced base with pixel-perfect dimensions.
func SciFiHumanoidAerial(direction Direction) AnatomicalTemplate {
	base := EnhancedHumanoidAerialTemplate(direction)
	base.Name = "scifi_aerial_" + string(direction)

	// Angular torso with tech aesthetic
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle}
	torsoSpec.RelativeWidth = 0.214 // 6/28 for 6 pixel width (sleeker than fantasy)
	torsoSpec.PreferredPixelSize = &PixelDimensions{
		Width:  6,
		Height: 6, // Phase 15.1: 6×6 pixel torso (angular tech body)
	}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Angular helmet head
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOctagon, shapes.ShapeHexagon}
	headSpec.RelativeWidth = 0.179  // 5/28 for 5 pixel dimensions
	headSpec.RelativeHeight = 0.179 // Symmetrical for helmet
	headSpec.PreferredPixelSize = &PixelDimensions{
		Width:  5,
		Height: 5, // Phase 15.1: 5×5 pixel angular helmet
	}
	base.BodyPartLayout[PartHead] = headSpec

	// Add jetpack indicator when facing up
	if direction == DirUp {
		base.BodyPartLayout[PartArmor] = PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.52,
			RelativeWidth:  0.143, // 4/28 for 4 pixel width
			RelativeHeight: 0.107, // 3/28 for 3 pixel height
			PreferredPixelSize: &PixelDimensions{
				Width:  4,
				Height: 3, // Phase 15.1: 4×3 pixel jetpack
			},
			ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon},
			ZIndex:     7, // Behind torso
			ColorRole:  "accent3",
			Opacity:    0.9,
			Rotation:   0,
		}
	}

	return base
}

// HorrorHumanoidAerial returns a horror-themed aerial humanoid template.
// Narrow head with skull shapes, irregular torso, reduced shadow for ghostly effect.
// Phase 15.1: Uses enhanced base with pixel-perfect dimensions.
func HorrorHumanoidAerial(direction Direction) AnatomicalTemplate {
	base := EnhancedHumanoidAerialTemplate(direction)
	base.Name = "horror_aerial_" + string(direction)

	// Narrow head for unsettling appearance
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeEllipse}
	// Keep 5×5 pixel dimensions from base
	base.BodyPartLayout[PartHead] = headSpec

	// Irregular torso shape (keep enhanced pixel dimensions)
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeBean}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Reduced shadow opacity for ghostly effect
	shadowSpec := base.BodyPartLayout[PartShadow]
	shadowSpec.Opacity = 0.2
	base.BodyPartLayout[PartShadow] = shadowSpec

	// Thin arms (keep base pixel dimensions but adjust shape)
	armsSpec := base.BodyPartLayout[PartArms]
	// Gaunt appearance is achieved through shape selection, not size change
	base.BodyPartLayout[PartArms] = armsSpec

	return base
}

// CyberpunkHumanoidAerial returns a cyberpunk-themed aerial humanoid template.
// Neon glow outlines, asymmetric tech implants, compact build.
// Phase 15.1: Uses enhanced base with pixel-perfect dimensions.
func CyberpunkHumanoidAerial(direction Direction) AnatomicalTemplate {
	base := EnhancedHumanoidAerialTemplate(direction)
	base.Name = "cyberpunk_aerial_" + string(direction)

	// Compact torso with angular shapes (keep enhanced pixel dimensions)
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeRectangle}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Angular head with tech aesthetic
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOctagon, shapes.ShapeHexagon}
	headSpec.ColorRole = "accent1" // Tech glow color
	base.BodyPartLayout[PartHead] = headSpec

	// Add neon glow outline as armor overlay
	base.BodyPartLayout[PartArmor] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.286, // 8/28 for 8 pixel width (28×28 sprite)
		RelativeHeight: 0.25,  // 7/28 for 7 pixel height (28×28 sprite)
		PreferredPixelSize: &PixelDimensions{
			Width:  8,
			Height: 7, // Phase 15.1: 8×7 pixel neon glow
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeHexagon},
		ZIndex:     9, // Just below torso
		ColorRole:  "accent1",
		Opacity:    0.3, // Subtle glow effect
		Rotation:   0,
	}

	return base
}

// PostApocHumanoidAerial returns a post-apocalyptic themed aerial humanoid template.
// Ragged edges, makeshift armor, irregular shapes for survival aesthetic.
// Phase 15.1: Uses enhanced base with pixel-perfect dimensions.
func PostApocHumanoidAerial(direction Direction) AnatomicalTemplate {
	base := EnhancedHumanoidAerialTemplate(direction)
	base.Name = "postapoc_aerial_" + string(direction)

	// Irregular torso with ragged appearance (keep enhanced pixel dimensions)
	torsoSpec := base.BodyPartLayout[PartTorso]
	torsoSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeBean, shapes.ShapeRectangle}
	base.BodyPartLayout[PartTorso] = torsoSpec

	// Covered head (masks, hoods, makeshift helmets)
	headSpec := base.BodyPartLayout[PartHead]
	headSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeOrganic, shapes.ShapeSkull}
	base.BodyPartLayout[PartHead] = headSpec

	// Rough, angular limbs (keep enhanced pixel dimensions)
	armsSpec := base.BodyPartLayout[PartArms]
	armsSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule}
	base.BodyPartLayout[PartArms] = armsSpec

	legsSpec := base.BodyPartLayout[PartLegs]
	legsSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule}
	base.BodyPartLayout[PartLegs] = legsSpec

	return base
}

// SelectAerialTemplate chooses an appropriate aerial template based on entity type, genre, and direction.
// For humanoid types, returns genre-specific aerial templates.
// For non-humanoid types, falls back to existing side-view templates.
func SelectAerialTemplate(entityType, genre string, direction Direction) AnatomicalTemplate {
	// Check if this is a humanoid type
	isHumanoid := false
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior":
		isHumanoid = true
	}

	if !isHumanoid {
		// Fall back to existing templates for non-humanoid entities
		return SelectTemplate(entityType)
	}

	// Apply genre-specific aerial styling for humanoids
	switch genre {
	case "fantasy":
		return FantasyHumanoidAerial(direction)
	case "scifi", "sci-fi":
		return SciFiHumanoidAerial(direction)
	case "horror":
		return HorrorHumanoidAerial(direction)
	case "cyberpunk":
		return CyberpunkHumanoidAerial(direction)
	case "postapoc", "post-apocalyptic":
		return PostApocHumanoidAerial(direction)
	default:
		// Use enhanced template for unknown genres (Phase 15.1)
		return EnhancedHumanoidAerialTemplate(direction)
	}
}

// ============================================================================
// Phase 15.1: Genre-Specific Anatomical Variations
// ============================================================================
// These functions apply genre-specific styling to non-humanoid creature templates.
// Each genre variation modifies shape types, proportions, and visual characteristics
// to match the genre's aesthetic while maintaining the base anatomical structure.

// ApplyFantasyVariation applies organic, natural styling to a creature template.
// Fantasy genre emphasizes softer shapes, organic forms, and natural proportions.
// Shape modifications: prefer organic, bean, ellipse shapes over geometric ones.
func ApplyFantasyVariation(template AnatomicalTemplate) AnatomicalTemplate {
	fantasy := template
	fantasy.Name = "fantasy_" + template.Name

	// Create new body part layout to avoid modifying original
	fantasy.BodyPartLayout = make(map[BodyPart]PartSpec)

	// Apply organic shape preferences to all body parts
	for part, spec := range template.BodyPartLayout {
		// Skip shadow - it remains ellipse
		if part == PartShadow {
			fantasy.BodyPartLayout[part] = spec
			continue
		}

		// Filter shapes to prefer organic variants
		organicShapes := []shapes.ShapeType{}
		for _, shapeType := range spec.ShapeTypes {
			switch shapeType {
			// Keep organic shapes
			case shapes.ShapeOrganic, shapes.ShapeBean, shapes.ShapeEllipse,
				shapes.ShapeCircle, shapes.ShapeCapsule, shapes.ShapeWave:
				organicShapes = append(organicShapes, shapeType)
			// Replace geometric shapes with organic equivalents
			case shapes.ShapeRectangle:
				organicShapes = append(organicShapes, shapes.ShapeCapsule)
			case shapes.ShapeHexagon, shapes.ShapeOctagon:
				organicShapes = append(organicShapes, shapes.ShapeEllipse)
			case shapes.ShapeTriangle:
				organicShapes = append(organicShapes, shapes.ShapeWedge)
			// Keep unique shapes that add character
			default:
				organicShapes = append(organicShapes, shapeType)
			}
		}

		if len(organicShapes) > 0 {
			spec.ShapeTypes = organicShapes
		}
		fantasy.BodyPartLayout[part] = spec
	}

	return fantasy
}

// ApplySciFiVariation applies geometric, angular styling to a creature template.
// Sci-fi genre emphasizes precise shapes, angular forms, and mechanical precision.
// Shape modifications: prefer hexagon, octagon, rectangle over organic shapes.
func ApplySciFiVariation(template AnatomicalTemplate) AnatomicalTemplate {
	scifi := template
	scifi.Name = "scifi_" + template.Name

	// Create new body part layout to avoid modifying original
	scifi.BodyPartLayout = make(map[BodyPart]PartSpec)

	// Apply geometric shape preferences to all body parts
	for part, spec := range template.BodyPartLayout {
		// Skip shadow - it remains ellipse
		if part == PartShadow {
			scifi.BodyPartLayout[part] = spec
			continue
		}

		// Filter shapes to prefer geometric variants
		geometricShapes := []shapes.ShapeType{}
		for _, shapeType := range spec.ShapeTypes {
			switch shapeType {
			// Keep geometric shapes
			case shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle,
				shapes.ShapeTriangle, shapes.ShapeGear, shapes.ShapeCrystal:
				geometricShapes = append(geometricShapes, shapeType)
			// Replace organic shapes with geometric equivalents
			case shapes.ShapeOrganic, shapes.ShapeBean:
				geometricShapes = append(geometricShapes, shapes.ShapeHexagon)
			case shapes.ShapeEllipse:
				geometricShapes = append(geometricShapes, shapes.ShapeOctagon)
			case shapes.ShapeCircle:
				geometricShapes = append(geometricShapes, shapes.ShapeHexagon)
			case shapes.ShapeCapsule:
				geometricShapes = append(geometricShapes, shapes.ShapeRectangle)
			case shapes.ShapeWave:
				geometricShapes = append(geometricShapes, shapes.ShapeLightning)
			// Keep unique shapes
			default:
				geometricShapes = append(geometricShapes, shapeType)
			}
		}

		if len(geometricShapes) > 0 {
			spec.ShapeTypes = geometricShapes
		}
		scifi.BodyPartLayout[part] = spec
	}

	return scifi
}

// ApplyHorrorVariation applies distorted, unsettling styling to a creature template.
// Horror genre emphasizes irregular shapes, distorted proportions, and unsettling forms.
// Modifies proportions: elongates some parts, shrinks others, reduces opacity for ghostly effect.
func ApplyHorrorVariation(template AnatomicalTemplate) AnatomicalTemplate {
	horror := template
	horror.Name = "horror_" + template.Name

	// Create new body part layout to avoid modifying original
	horror.BodyPartLayout = make(map[BodyPart]PartSpec)

	// Apply distortion to all body parts
	for part, spec := range template.BodyPartLayout {
		// Reduce shadow opacity for ghostly/otherworldly effect
		if part == PartShadow {
			spec.Opacity *= 0.6 // Make shadow fainter
			horror.BodyPartLayout[part] = spec
			continue
		}

		// Distort proportions based on body part
		switch part {
		case PartHead:
			// Elongate head for unsettling appearance
			spec.RelativeHeight *= 1.2
			spec.RelativeWidth *= 0.85
			// Prefer skull and organic shapes
			spec.ShapeTypes = []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeOrganic, shapes.ShapeEllipse}
		case PartTorso:
			// Irregular torso
			spec.ShapeTypes = []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeBean}
			spec.RelativeWidth *= 0.9
		case PartLegs, PartArms:
			// Elongate limbs for unnatural appearance
			spec.RelativeHeight *= 1.15
			spec.RelativeWidth *= 0.85
		}

		// Slightly reduce opacity for translucent/ethereal effect
		if spec.Opacity > 0.5 {
			spec.Opacity *= 0.95
		}

		horror.BodyPartLayout[part] = spec
	}

	return horror
}

// ApplyCyberpunkVariation applies augmented, tech-enhanced styling to a creature template.
// Cyberpunk genre emphasizes mechanical additions, tech implants, and neon accents.
// Adds armor/tech overlay parts and modifies colors to show augmentation.
func ApplyCyberpunkVariation(template AnatomicalTemplate) AnatomicalTemplate {
	cyberpunk := template
	cyberpunk.Name = "cyberpunk_" + template.Name

	// Create new body part layout to avoid modifying original
	cyberpunk.BodyPartLayout = make(map[BodyPart]PartSpec)

	// Apply angular shapes and tech aesthetic
	for part, spec := range template.BodyPartLayout {
		if part == PartShadow {
			cyberpunk.BodyPartLayout[part] = spec
			continue
		}

		// Prefer angular, tech-like shapes
		techShapes := []shapes.ShapeType{}
		for _, shapeType := range spec.ShapeTypes {
			switch shapeType {
			// Keep angular shapes
			case shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle:
				techShapes = append(techShapes, shapeType)
			// Replace organic with angular
			case shapes.ShapeOrganic, shapes.ShapeBean:
				techShapes = append(techShapes, shapes.ShapeHexagon)
			case shapes.ShapeCircle, shapes.ShapeEllipse:
				techShapes = append(techShapes, shapes.ShapeOctagon)
			default:
				techShapes = append(techShapes, shapeType)
			}
		}

		if len(techShapes) > 0 {
			spec.ShapeTypes = techShapes
		}

		// Change color roles to show tech augmentation
		if part == PartHead {
			spec.ColorRole = "accent1" // Neon glow for head
		}

		cyberpunk.BodyPartLayout[part] = spec
	}

	// Add tech overlay/armor if torso exists
	if torsoSpec, hasTorso := template.BodyPartLayout[PartTorso]; hasTorso {
		armorSpec := torsoSpec
		armorSpec.RelativeWidth *= 1.1 // Slightly larger
		armorSpec.RelativeHeight *= 1.1
		armorSpec.ZIndex = torsoSpec.ZIndex - 1 // Behind torso
		armorSpec.ColorRole = "accent1"
		armorSpec.Opacity = 0.4 // Translucent tech glow
		armorSpec.ShapeTypes = []shapes.ShapeType{shapes.ShapeHexagon, shapes.ShapeOctagon}
		cyberpunk.BodyPartLayout[PartArmor] = armorSpec
	}

	return cyberpunk
}

// ApplyPostApocVariation applies weathered, damaged styling to a creature template.
// Post-apocalyptic genre emphasizes rough edges, irregular forms, and worn appearance.
// Modifies shapes to appear makeshift and damaged.
func ApplyPostApocVariation(template AnatomicalTemplate) AnatomicalTemplate {
	postapoc := template
	postapoc.Name = "postapoc_" + template.Name

	// Create new body part layout to avoid modifying original
	postapoc.BodyPartLayout = make(map[BodyPart]PartSpec)

	// Apply rough, irregular styling
	for part, spec := range template.BodyPartLayout {
		if part == PartShadow {
			postapoc.BodyPartLayout[part] = spec
			continue
		}

		// Prefer rough, irregular shapes
		roughShapes := []shapes.ShapeType{}
		for _, shapeType := range spec.ShapeTypes {
			switch shapeType {
			// Keep rough shapes
			case shapes.ShapeOrganic, shapes.ShapeRectangle, shapes.ShapeCapsule:
				roughShapes = append(roughShapes, shapeType)
			// Replace smooth with rough equivalents
			case shapes.ShapeCircle:
				roughShapes = append(roughShapes, shapes.ShapeOrganic)
			case shapes.ShapeEllipse:
				roughShapes = append(roughShapes, shapes.ShapeBean)
			case shapes.ShapeHexagon, shapes.ShapeOctagon:
				roughShapes = append(roughShapes, shapes.ShapeRectangle)
			default:
				roughShapes = append(roughShapes, shapeType)
			}
		}

		if len(roughShapes) > 0 {
			spec.ShapeTypes = roughShapes
		}
		postapoc.BodyPartLayout[part] = spec
	}

	return postapoc
}

// SelectTemplateWithGenre chooses an appropriate template with genre-specific styling.
// This replaces SelectTemplate for genre-aware template selection.
// Applies genre variations to non-humanoid creature types.
// For humanoid types, use SelectHumanoidTemplate instead.
func SelectTemplateWithGenre(entityType, genre string) AnatomicalTemplate {
	// Get base template
	var baseTemplate AnatomicalTemplate
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior":
		// For humanoids, return default - caller should use SelectHumanoidTemplate with direction
		return HumanoidTemplate()
	case "quadruped", "wolf", "bear", "animal", "beast":
		baseTemplate = QuadrupedTemplate()
	case "blob", "slime", "amoeba", "ooze":
		baseTemplate = BlobTemplate()
	case "mechanical", "robot", "golem", "construct", "android":
		baseTemplate = MechanicalTemplate()
	case "flying", "bird", "dragon", "bat", "wyvern":
		baseTemplate = FlyingTemplate()
	case "serpentine", "snake", "worm", "tentacle", "wyrm":
		baseTemplate = SerpentineTemplate()
	case "arachnid", "spider", "insect", "beetle":
		baseTemplate = ArachnidTemplate()
	case "undead", "skeleton", "ghost", "zombie", "lich":
		baseTemplate = UndeadTemplate()
	default:
		// Default to humanoid for unknown types
		return HumanoidTemplate()
	}

	// Apply genre-specific variation
	switch genre {
	case "fantasy":
		return ApplyFantasyVariation(baseTemplate)
	case "scifi", "sci-fi":
		return ApplySciFiVariation(baseTemplate)
	case "horror":
		return ApplyHorrorVariation(baseTemplate)
	case "cyberpunk":
		return ApplyCyberpunkVariation(baseTemplate)
	case "postapoc", "post-apocalyptic":
		return ApplyPostApocVariation(baseTemplate)
	default:
		// No genre variation, return base template
		return baseTemplate
	}
}

// BossAerialTemplate creates a scaled boss variant of an aerial template.
// Applies uniform scaling to all body parts while preserving:
// - 35/50/15 proportion ratios
// - Directional asymmetry (head offset, arm positioning)
// - Color role assignments
// - Z-index layering
//
// The scale parameter is typically 2.5 for boss entities, making them
// 2.5× larger than normal entities while maintaining the same visual structure.
// Directional asymmetry remains intact (head offsets scale proportionally).
//
// Example usage:
//
//	base := FantasyHumanoidAerial(DirDown)
//	boss := BossAerialTemplate(base, 2.5)
func BossAerialTemplate(base AnatomicalTemplate, scale float64) AnatomicalTemplate {
	if scale <= 0 {
		scale = 1.0 // Safety: prevent invalid scaling
	}

	boss := AnatomicalTemplate{
		Name:           base.Name + "_boss",
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Scale all body parts proportionally
	for part, spec := range base.BodyPartLayout {
		scaledSpec := spec

		// Scale dimensions (width and height)
		scaledSpec.RelativeWidth *= scale
		scaledSpec.RelativeHeight *= scale

		// Scale position offsets from center (0.5, 0.5)
		// This maintains directional asymmetry:
		// - Left head (X=0.42) becomes further left
		// - Right head (X=0.58) becomes further right
		offsetX := spec.RelativeX - 0.5
		offsetY := spec.RelativeY - 0.5
		scaledSpec.RelativeX = 0.5 + (offsetX * scale)
		scaledSpec.RelativeY = 0.5 + (offsetY * scale)

		// Keep color roles, shapes, opacity, rotation, and Z-index unchanged
		// These define the visual character and should not scale

		boss.BodyPartLayout[part] = scaledSpec
	}

	return boss
}

// GetEffectiveWidth returns the effective width for a PartSpec in pixels.
// If PreferredPixelSize is set, uses that width. Otherwise, calculates from RelativeWidth.
// This enables Phase 15.1 enhanced proportional scaling with explicit pixel control.
func (p *PartSpec) GetEffectiveWidth(spriteWidth int) int {
	if p.PreferredPixelSize != nil {
		return p.PreferredPixelSize.Width
	}
	return int(float64(spriteWidth) * p.RelativeWidth)
}

// GetEffectiveHeight returns the effective height for a PartSpec in pixels.
// If PreferredPixelSize is set, uses that height. Otherwise, calculates from RelativeHeight.
// This enables Phase 15.1 enhanced proportional scaling with explicit pixel control.
func (p *PartSpec) GetEffectiveHeight(spriteHeight int) int {
	if p.PreferredPixelSize != nil {
		return p.PreferredPixelSize.Height
	}
	return int(float64(spriteHeight) * p.RelativeHeight)
}

// ToPixelDimensions converts relative dimensions to exact pixel dimensions.
// This is useful for creating templates with Phase 15.1 pixel-perfect specifications.
func (p *PartSpec) ToPixelDimensions(spriteWidth, spriteHeight int) PixelDimensions {
	return PixelDimensions{
		Width:  int(float64(spriteWidth) * p.RelativeWidth),
		Height: int(float64(spriteHeight) * p.RelativeHeight),
	}
}

// WithPixelDimensions creates a new PartSpec with the specified pixel dimensions set.
// This enables Phase 15.1 "head 4×4, torso 4×6, legs 4×8" style specifications.
// Returns a new PartSpec with PreferredPixelSize set, keeping all other fields unchanged.
// Width and height are clamped to minimum of 1 pixel to prevent rendering issues.
func (p PartSpec) WithPixelDimensions(width, height int) PartSpec {
	// Clamp to minimum 1 pixel to prevent rendering issues
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	p.PreferredPixelSize = &PixelDimensions{
		Width:  width,
		Height: height,
	}
	return p
}

// NewPartSpecFromPixels creates a PartSpec with explicit pixel dimensions.
// This is a convenience constructor for Phase 15.1 enhanced anatomical templates.
// The relative dimensions are calculated based on typical sprite sizes for reference,
// but PreferredPixelSize takes precedence during rendering.
// Width and height are clamped to minimum of 1 pixel to prevent rendering issues.
//
// Example for Phase 15.1 humanoid:
//
//	head := NewPartSpecFromPixels(4, 4, shapes.ShapeCircle, 15, "secondary")
//	torso := NewPartSpecFromPixels(4, 6, shapes.ShapeRectangle, 10, "primary")
//	legs := NewPartSpecFromPixels(4, 8, shapes.ShapeCapsule, 5, "primary")
func NewPartSpecFromPixels(width, height int, shapeType shapes.ShapeType, zIndex int, colorRole string) PartSpec {
	// Clamp to minimum 1 pixel to prevent rendering issues
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// Calculate relative dimensions assuming a typical 28x28 sprite size
	// These are fallbacks if PreferredPixelSize is ignored
	const typicalSize = 28.0
	relWidth := float64(width) / typicalSize
	relHeight := float64(height) / typicalSize

	return PartSpec{
		RelativeX:      0.5, // Centered by default
		RelativeY:      0.5, // Centered by default
		RelativeWidth:  relWidth,
		RelativeHeight: relHeight,
		PreferredPixelSize: &PixelDimensions{
			Width:  width,
			Height: height,
		},
		ShapeTypes: []shapes.ShapeType{shapeType},
		ZIndex:     zIndex,
		ColorRole:  colorRole,
		Opacity:    1.0,
		Rotation:   0,
	}
}
