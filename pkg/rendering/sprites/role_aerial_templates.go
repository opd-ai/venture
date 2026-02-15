// Package sprites provides role-based aerial anatomy templates for humanoid NPCs.
// Each role (mage, warrior, knight, rogue, merchant, ranger, priest) has a
// dedicated top-down body plan with distinct proportions and silhouettes so that
// players can identify NPC roles at a glance even at 32×32 pixel resolution.
package sprites

import (
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// VisualRole is a string alias representing a humanoid NPC's visual archetype.
type VisualRole string

const (
	RoleMage     VisualRole = "mage"
	RoleWarrior  VisualRole = "warrior"
	RoleKnight   VisualRole = "knight"
	RoleRogue    VisualRole = "rogue"
	RoleMerchant VisualRole = "merchant"
	RoleRanger   VisualRole = "ranger"
	RolePriest   VisualRole = "priest"
)

// MapEntityTypeToRole maps an entity type string (from procgen or class system)
// to a VisualRole for template selection. Returns empty string if not a
// recognized role-specific type.
func MapEntityTypeToRole(entityType string) VisualRole {
	switch entityType {
	case "mage", "elementalist", "necromancer", "enchanter":
		return RoleMage
	case "warrior", "berserker":
		return RoleWarrior
	case "knight", "paladin":
		return RoleKnight
	case "rogue", "assassin", "ninja":
		return RoleRogue
	case "merchant":
		return RoleMerchant
	case "ranger":
		return RoleRanger
	case "cleric", "druid", "bard", "priest":
		return RolePriest
	}
	return ""
}

// SelectRoleAerialTemplate returns a role-specific aerial anatomy template.
// Falls back to EnhancedHumanoidAerialTemplate if role is unknown.
func SelectRoleAerialTemplate(role VisualRole, direction Direction) AnatomicalTemplate {
	switch role {
	case RoleMage:
		return MageAerialTemplate(direction)
	case RoleWarrior:
		return WarriorAerialTemplate(direction)
	case RoleKnight:
		return KnightAerialTemplate(direction)
	case RoleRogue:
		return RogueAerialTemplate(direction)
	case RoleMerchant:
		return MerchantAerialTemplate(direction)
	case RoleRanger:
		return RangerAerialTemplate(direction)
	case RolePriest:
		return PriestAerialTemplate(direction)
	default:
		return EnhancedHumanoidAerialTemplate(direction)
	}
}

// ---------------------------------------------------------------------------
// Mage — wide pointed hat dominates top-down view, narrow body under robes
// ---------------------------------------------------------------------------

func MageAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "mage_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.44, RelativeHeight: 0.13,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.35,
	}

	// Robes visible as a wide lower shape beneath the torso
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.78,
		RelativeWidth: 0.28, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 9, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCapsule},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.85,
	}

	// Narrow torso — the body is hidden under layered robes
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.52,
		RelativeWidth: 0.19, RelativeHeight: 0.22,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 7},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	// Large hat brim — the most distinctive feature from above
	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.22,
		RelativeWidth: 0.31, RelativeHeight: 0.22,
		PreferredPixelSize: &PixelDimensions{Width: 10, Height: 7},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeHexagon},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	// Thin arms holding staff — asymmetric placement hints at held item
	armSpec := PartSpec{
		RelativeWidth: 0.13, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 4, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
		ZIndex: 8, ColorRole: "secondary", Opacity: 1.0,
	}

	switch direction {
	case DirUp:
		headSpec.RelativeY = 0.22
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.52
		armSpec.RelativeWidth = 0.22
	case DirDown:
		headSpec.RelativeY = 0.22
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.52
		armSpec.RelativeWidth = 0.22
	case DirLeft:
		headSpec.RelativeX = 0.45
		armSpec.RelativeX, armSpec.RelativeY = 0.30, 0.48
		armSpec.RelativeWidth = 0.13
	case DirRight:
		headSpec.RelativeX = 0.55
		armSpec.RelativeX, armSpec.RelativeY = 0.70, 0.48
		armSpec.RelativeWidth = 0.13
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	return t
}

// ---------------------------------------------------------------------------
// Warrior — broad shoulders, wide muscular build, sturdy stance
// ---------------------------------------------------------------------------

func WarriorAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "warrior_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.56, RelativeHeight: 0.14,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.35,
	}

	// Sturdy legs, wider stance visible from above
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.80,
		RelativeWidth: 0.19, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.9,
	}

	// Wide, muscular torso
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.50,
		RelativeWidth: 0.31, RelativeHeight: 0.22,
		PreferredPixelSize: &PixelDimensions{Width: 10, Height: 7},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule, shapes.ShapeEllipse},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.22,
		RelativeWidth: 0.19, RelativeHeight: 0.16,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 5},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	// Broad arms — wider than normal, muscular
	armSpec := PartSpec{
		RelativeWidth: 0.28, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 9, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
		ZIndex: 9, ColorRole: "secondary", Opacity: 1.0,
	}

	switch direction {
	case DirUp:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
	case DirDown:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
	case DirLeft:
		headSpec.RelativeX = 0.45
		armSpec.RelativeX, armSpec.RelativeY = 0.32, 0.48
		armSpec.RelativeWidth = 0.16
	case DirRight:
		headSpec.RelativeX = 0.55
		armSpec.RelativeX, armSpec.RelativeY = 0.68, 0.48
		armSpec.RelativeWidth = 0.16
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	return t
}

// ---------------------------------------------------------------------------
// Knight — heavily armored, very broad with pauldrons, large helm
// ---------------------------------------------------------------------------

func KnightAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "knight_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.59, RelativeHeight: 0.16,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.4,
	}

	// Heavy armored legs
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.80,
		RelativeWidth: 0.19, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.9,
	}

	// Massive armored torso with pauldrons
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.48,
		RelativeWidth: 0.34, RelativeHeight: 0.25,
		PreferredPixelSize: &PixelDimensions{Width: 11, Height: 8},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeOctagon},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	// Large helmet — distinctive from above
	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.20,
		RelativeWidth: 0.22, RelativeHeight: 0.19,
		PreferredPixelSize: &PixelDimensions{Width: 7, Height: 6},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeOctagon},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	// Armored arms, very wide — pauldron-like extension
	armSpec := PartSpec{
		RelativeWidth: 0.34, RelativeHeight: 0.13,
		PreferredPixelSize: &PixelDimensions{Width: 11, Height: 4},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
		ZIndex: 11, ColorRole: "accent1", Opacity: 1.0,
	}

	switch direction {
	case DirUp:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.48
	case DirDown:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.48
	case DirLeft:
		headSpec.RelativeX = 0.44
		armSpec.RelativeX, armSpec.RelativeY = 0.30, 0.46
		armSpec.RelativeWidth = 0.19
	case DirRight:
		headSpec.RelativeX = 0.56
		armSpec.RelativeX, armSpec.RelativeY = 0.70, 0.46
		armSpec.RelativeWidth = 0.19
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	return t
}

// ---------------------------------------------------------------------------
// Rogue — lean, compact, hooded silhouette, narrow build
// ---------------------------------------------------------------------------

func RogueAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "rogue_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.38, RelativeHeight: 0.11,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.30,
	}

	// Nimble legs, narrow
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.80,
		RelativeWidth: 0.13, RelativeHeight: 0.06,
		PreferredPixelSize: &PixelDimensions{Width: 4, Height: 2},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.8,
	}

	// Compact lean torso
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.50,
		RelativeWidth: 0.19, RelativeHeight: 0.19,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 6},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	// Hooded head — slightly pointed/triangular from above
	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.22,
		RelativeWidth: 0.19, RelativeHeight: 0.19,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 6},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle, shapes.ShapeCircle},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	// Thin arms, close to body
	armSpec := PartSpec{
		RelativeWidth: 0.16, RelativeHeight: 0.06,
		PreferredPixelSize: &PixelDimensions{Width: 5, Height: 2},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
		ZIndex: 8, ColorRole: "secondary", Opacity: 1.0,
	}

	switch direction {
	case DirUp:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
	case DirDown:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
	case DirLeft:
		headSpec.RelativeX = 0.46
		armSpec.RelativeX, armSpec.RelativeY = 0.34, 0.48
	case DirRight:
		headSpec.RelativeX = 0.54
		armSpec.RelativeX, armSpec.RelativeY = 0.66, 0.48
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	return t
}

// ---------------------------------------------------------------------------
// Merchant — rotund body, wider torso, pack/satchel on back
// ---------------------------------------------------------------------------

func MerchantAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "merchant_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.53, RelativeHeight: 0.16,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.35,
	}

	// Short legs beneath rotund body
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.82,
		RelativeWidth: 0.13, RelativeHeight: 0.06,
		PreferredPixelSize: &PixelDimensions{Width: 4, Height: 2},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.8,
	}

	// Wide, round torso — distinctive rotund build
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.50,
		RelativeWidth: 0.31, RelativeHeight: 0.28,
		PreferredPixelSize: &PixelDimensions{Width: 10, Height: 9},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	// Normal head, maybe with hat
	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.20,
		RelativeWidth: 0.19, RelativeHeight: 0.16,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 5},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	// Arms out — holding or gesturing
	armSpec := PartSpec{
		RelativeWidth: 0.22, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 7, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeEllipse},
		ZIndex: 8, ColorRole: "secondary", Opacity: 1.0,
	}

	// Backpack — small rectangle/ellipse offset behind the body
	packSpec := PartSpec{
		RelativeWidth: 0.13, RelativeHeight: 0.13,
		PreferredPixelSize: &PixelDimensions{Width: 4, Height: 4},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
		ZIndex: 12, ColorRole: "accent1", Opacity: 0.9,
	}

	switch direction {
	case DirUp:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
		packSpec.RelativeX, packSpec.RelativeY = 0.5, 0.62
	case DirDown:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
		packSpec.RelativeX, packSpec.RelativeY = 0.5, 0.36
	case DirLeft:
		headSpec.RelativeX = 0.44
		armSpec.RelativeX, armSpec.RelativeY = 0.30, 0.48
		armSpec.RelativeWidth = 0.13
		packSpec.RelativeX, packSpec.RelativeY = 0.64, 0.48
	case DirRight:
		headSpec.RelativeX = 0.56
		armSpec.RelativeX, armSpec.RelativeY = 0.70, 0.48
		armSpec.RelativeWidth = 0.13
		packSpec.RelativeX, packSpec.RelativeY = 0.36, 0.48
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	t.BodyPartLayout[PartArmor] = packSpec // reuse armor slot for the backpack visual
	return t
}

// ---------------------------------------------------------------------------
// Ranger — athletic build, asymmetric quiver on back, hooded
// ---------------------------------------------------------------------------

func RangerAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "ranger_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.47, RelativeHeight: 0.13,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.33,
	}

	// Active-stance legs
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.80,
		RelativeWidth: 0.16, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 5, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeEllipse},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.85,
	}

	// Athletic torso
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.50,
		RelativeWidth: 0.22, RelativeHeight: 0.22,
		PreferredPixelSize: &PixelDimensions{Width: 7, Height: 7},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCapsule},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	// Hooded head
	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.22,
		RelativeWidth: 0.19, RelativeHeight: 0.16,
		PreferredPixelSize: &PixelDimensions{Width: 6, Height: 5},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeWedge},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	armSpec := PartSpec{
		RelativeWidth: 0.22, RelativeHeight: 0.09,
		PreferredPixelSize: &PixelDimensions{Width: 7, Height: 3},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule},
		ZIndex: 8, ColorRole: "secondary", Opacity: 1.0,
	}

	// Quiver — narrow rectangle offset to one side
	quiverSpec := PartSpec{
		RelativeWidth: 0.06, RelativeHeight: 0.16,
		PreferredPixelSize: &PixelDimensions{Width: 2, Height: 5},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
		ZIndex: 12, ColorRole: "accent1", Opacity: 0.85,
	}

	switch direction {
	case DirUp:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
		quiverSpec.RelativeX, quiverSpec.RelativeY = 0.66, 0.50
	case DirDown:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
		quiverSpec.RelativeX, quiverSpec.RelativeY = 0.34, 0.38
	case DirLeft:
		headSpec.RelativeX = 0.46
		armSpec.RelativeX, armSpec.RelativeY = 0.32, 0.48
		armSpec.RelativeWidth = 0.13
		quiverSpec.RelativeX, quiverSpec.RelativeY = 0.62, 0.44
	case DirRight:
		headSpec.RelativeX = 0.54
		armSpec.RelativeX, armSpec.RelativeY = 0.68, 0.48
		armSpec.RelativeWidth = 0.13
		quiverSpec.RelativeX, quiverSpec.RelativeY = 0.38, 0.44
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	t.BodyPartLayout[PartArmor] = quiverSpec // reuse armor slot for quiver
	return t
}

// ---------------------------------------------------------------------------
// Priest/Support — tall, flowing vestments, wider robes, halo-like head detail
// ---------------------------------------------------------------------------

func PriestAerialTemplate(direction Direction) AnatomicalTemplate {
	t := AnatomicalTemplate{
		Name:           "priest_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	t.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.88,
		RelativeWidth: 0.47, RelativeHeight: 0.13,
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 0, ColorRole: "shadow", Opacity: 0.35,
	}

	// Robes cover legs almost entirely
	t.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.82,
		RelativeWidth: 0.09, RelativeHeight: 0.06,
		PreferredPixelSize: &PixelDimensions{Width: 3, Height: 2},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex: 5, ColorRole: "primary", Opacity: 0.7,
	}

	// Tall flowing torso/vestments
	t.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX: 0.5, RelativeY: 0.50,
		RelativeWidth: 0.22, RelativeHeight: 0.28,
		PreferredPixelSize: &PixelDimensions{Width: 7, Height: 9},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean, shapes.ShapeCapsule},
		ZIndex: 10, ColorRole: "primary", Opacity: 1.0,
	}

	// Head with halo — slightly larger circle to suggest aureole
	headSpec := PartSpec{
		RelativeX: 0.5, RelativeY: 0.18,
		RelativeWidth: 0.22, RelativeHeight: 0.19,
		PreferredPixelSize: &PixelDimensions{Width: 7, Height: 6},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle},
		ZIndex: 15, ColorRole: "secondary", Opacity: 1.0,
	}

	// Wide sleeves
	armSpec := PartSpec{
		RelativeWidth: 0.25, RelativeHeight: 0.13,
		PreferredPixelSize: &PixelDimensions{Width: 8, Height: 4},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCapsule},
		ZIndex: 9, ColorRole: "accent1", Opacity: 0.95,
	}

	switch direction {
	case DirUp:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
	case DirDown:
		armSpec.RelativeX, armSpec.RelativeY = 0.5, 0.50
	case DirLeft:
		headSpec.RelativeX = 0.46
		armSpec.RelativeX, armSpec.RelativeY = 0.30, 0.48
		armSpec.RelativeWidth = 0.16
	case DirRight:
		headSpec.RelativeX = 0.56
		armSpec.RelativeX, armSpec.RelativeY = 0.70, 0.48
		armSpec.RelativeWidth = 0.16
	}

	t.BodyPartLayout[PartHead] = headSpec
	t.BodyPartLayout[PartArms] = armSpec
	return t
}
