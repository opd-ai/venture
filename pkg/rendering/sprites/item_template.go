// Package sprites - item template system for recognizable equipment and consumables.
// This file implements Phase 5.4 of the Visual Fidelity Enhancement Plan.
package sprites

import (
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// ItemCategory represents the broad category of an item.
type ItemCategory string

const (
	// CategoryWeapon represents offensive equipment
	CategoryWeapon ItemCategory = "weapon"
	// CategoryArmor represents defensive equipment
	CategoryArmor ItemCategory = "armor"
	// CategoryConsumable represents usable items
	CategoryConsumable ItemCategory = "consumable"
	// CategoryAccessory represents worn items that provide passive bonuses
	CategoryAccessory ItemCategory = "accessory"
	// CategoryQuest represents key items for quests
	CategoryQuest ItemCategory = "quest"
)

// ItemType represents the specific type within a category.
type ItemType string

const (
	// Weapons
	ItemSword  ItemType = "sword"
	ItemAxe    ItemType = "axe"
	ItemBow    ItemType = "bow"
	ItemStaff  ItemType = "staff"
	ItemGun    ItemType = "gun"
	ItemDagger ItemType = "dagger"
	ItemHammer ItemType = "hammer"
	ItemSpear  ItemType = "spear"

	// Armor
	ItemHelmet ItemType = "helmet"
	ItemChest  ItemType = "chest"
	ItemShield ItemType = "shield"
	ItemBoots  ItemType = "boots"
	ItemGloves ItemType = "gloves"

	// Consumables
	ItemPotion ItemType = "potion"
	ItemFood   ItemType = "food"
	ItemScroll ItemType = "scroll"
	ItemElixir ItemType = "elixir"

	// Accessories
	ItemRing    ItemType = "ring"
	ItemAmulet  ItemType = "amulet"
	ItemTrinket ItemType = "trinket"

	// Quest Items
	ItemKey      ItemType = "key"
	ItemArtifact ItemType = "artifact"
	ItemRelic    ItemType = "relic"
)

// ItemRarity represents item rarity tiers.
type ItemRarity int

const (
	// RarityCommon represents common items
	RarityCommon ItemRarity = iota
	// RarityUncommon represents uncommon items
	RarityUncommon
	// RarityRare represents rare items
	RarityRare
	// RarityEpic represents epic items
	RarityEpic
	// RarityLegendary represents legendary items
	RarityLegendary
)

// String returns the string representation of an item rarity.
func (r ItemRarity) String() string {
	switch r {
	case RarityCommon:
		return "common"
	case RarityUncommon:
		return "uncommon"
	case RarityRare:
		return "rare"
	case RarityEpic:
		return "epic"
	case RarityLegendary:
		return "legendary"
	default:
		return "unknown"
	}
}

// ItemTemplate defines the visual structure of an item.
type ItemTemplate struct {
	// Name identifies this template
	Name string
	// Category is the broad item category
	Category ItemCategory
	// Type is the specific item type
	Type ItemType
	// Parts defines the shapes that make up this item
	Parts []ItemPartSpec
}

// ItemPartSpec defines a single visual component of an item.
type ItemPartSpec struct {
	// RelativeX is the X position as a fraction (0.0-1.0)
	RelativeX float64
	// RelativeY is the Y position as a fraction (0.0-1.0)
	RelativeY float64
	// RelativeWidth is the width as a fraction (0.0-1.0)
	RelativeWidth float64
	// RelativeHeight is the height as a fraction (0.0-1.0)
	RelativeHeight float64
	// ShapeTypes are the allowed shapes for this part
	ShapeTypes []shapes.ShapeType
	// ZIndex determines draw order
	ZIndex int
	// ColorRole indicates which color to use
	ColorRole string
	// Opacity is the alpha transparency (0.0-1.0)
	Opacity float64
	// Rotation is the rotation angle in degrees
	Rotation float64
}

// GetRarityColorRole returns the appropriate color role for a given rarity.
func GetRarityColorRole(rarity ItemRarity) string {
	switch rarity {
	case RarityCommon:
		return "primary"
	case RarityUncommon:
		return "secondary"
	case RarityRare:
		return "accent1"
	case RarityEpic:
		return "accent2"
	case RarityLegendary:
		return "accent3"
	default:
		return "primary"
	}
}

// SwordTemplate returns a template for sword weapons.
func SwordTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Blade
		{
			RelativeX:      0.5,
			RelativeY:      0.35,
			RelativeWidth:  0.25,
			RelativeHeight: 0.70,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeBlade, shapes.ShapeRectangle},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
		// Hilt/Crossguard
		{
			RelativeX:      0.5,
			RelativeY:      0.70,
			RelativeWidth:  0.50,
			RelativeHeight: 0.15,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCross},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
		// Pommel
		{
			RelativeX:      0.5,
			RelativeY:      0.85,
			RelativeWidth:  0.25,
			RelativeHeight: 0.20,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeHexagon},
			ZIndex:         15,
			ColorRole:      "accent1",
			Opacity:        1.0,
			Rotation:       0,
		},
	}

	// Add glow for higher rarities
	if rarity >= RarityRare {
		parts = append(parts, ItemPartSpec{
			RelativeX:      0.5,
			RelativeY:      0.35,
			RelativeWidth:  0.35,
			RelativeHeight: 0.80,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeBlade},
			ZIndex:         5,
			ColorRole:      GetRarityColorRole(rarity),
			Opacity:        0.3,
			Rotation:       0,
		})
	}

	return ItemTemplate{
		Name:     "sword_" + rarity.String(),
		Category: CategoryWeapon,
		Type:     ItemSword,
		Parts:    parts,
	}
}

// AxeTemplate returns a template for axe weapons.
func AxeTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Handle
		{
			RelativeX:      0.5,
			RelativeY:      0.60,
			RelativeWidth:  0.20,
			RelativeHeight: 0.75,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
			ZIndex:         10,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
		// Axe Head (wedge)
		{
			RelativeX:      0.5,
			RelativeY:      0.25,
			RelativeWidth:  0.60,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle},
			ZIndex:         15,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       90,
		},
	}

	// Add accent detail for higher rarities
	if rarity >= RarityUncommon {
		parts = append(parts, ItemPartSpec{
			RelativeX:      0.5,
			RelativeY:      0.25,
			RelativeWidth:  0.50,
			RelativeHeight: 0.25,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeTriangle},
			ZIndex:         20,
			ColorRole:      "accent1",
			Opacity:        0.8,
			Rotation:       90,
		})
	}

	return ItemTemplate{
		Name:     "axe_" + rarity.String(),
		Category: CategoryWeapon,
		Type:     ItemAxe,
		Parts:    parts,
	}
}

// BowTemplate returns a template for bow weapons.
func BowTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Bow body (curved)
		{
			RelativeX:      0.5,
			RelativeY:      0.5,
			RelativeWidth:  0.40,
			RelativeHeight: 0.80,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCrescent, shapes.ShapeBean},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       90,
		},
		// Bowstring (thin line)
		{
			RelativeX:      0.5,
			RelativeY:      0.5,
			RelativeWidth:  0.10,
			RelativeHeight: 0.75,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle},
			ZIndex:         5,
			ColorRole:      "secondary",
			Opacity:        0.8,
			Rotation:       0,
		},
	}

	return ItemTemplate{
		Name:     "bow_" + rarity.String(),
		Category: CategoryWeapon,
		Type:     ItemBow,
		Parts:    parts,
	}
}

// StaffTemplate returns a template for staff weapons.
func StaffTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Staff shaft
		{
			RelativeX:      0.5,
			RelativeY:      0.60,
			RelativeWidth:  0.18,
			RelativeHeight: 0.85,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
			ZIndex:         10,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
		// Orb at top
		{
			RelativeX:      0.5,
			RelativeY:      0.18,
			RelativeWidth:  0.35,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeStar, shapes.ShapeCrystal},
			ZIndex:         15,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
	}

	// Add glow for magical effect
	if rarity >= RarityRare {
		parts = append(parts, ItemPartSpec{
			RelativeX:      0.5,
			RelativeY:      0.18,
			RelativeWidth:  0.45,
			RelativeHeight: 0.45,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle},
			ZIndex:         12,
			ColorRole:      GetRarityColorRole(rarity),
			Opacity:        0.4,
			Rotation:       0,
		})
	}

	return ItemTemplate{
		Name:     "staff_" + rarity.String(),
		Category: CategoryWeapon,
		Type:     ItemStaff,
		Parts:    parts,
	}
}

// GunTemplate returns a template for gun weapons (sci-fi).
func GunTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Gun body
		{
			RelativeX:      0.45,
			RelativeY:      0.5,
			RelativeWidth:  0.55,
			RelativeHeight: 0.40,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
		// Barrel
		{
			RelativeX:      0.75,
			RelativeY:      0.5,
			RelativeWidth:  0.35,
			RelativeHeight: 0.20,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
		// Grip/Handle
		{
			RelativeX:      0.30,
			RelativeY:      0.65,
			RelativeWidth:  0.22,
			RelativeHeight: 0.30,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle},
			ZIndex:         10,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       15,
		},
	}

	return ItemTemplate{
		Name:     "gun_" + rarity.String(),
		Category: CategoryWeapon,
		Type:     ItemGun,
		Parts:    parts,
	}
}

// HelmetTemplate returns a template for helmet armor.
func HelmetTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Helm dome
		{
			RelativeX:      0.5,
			RelativeY:      0.45,
			RelativeWidth:  0.70,
			RelativeHeight: 0.70,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeCircle, shapes.ShapeEllipse},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
		// Visor/faceplate
		{
			RelativeX:      0.5,
			RelativeY:      0.55,
			RelativeWidth:  0.50,
			RelativeHeight: 0.30,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeEllipse},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        0.7,
			Rotation:       0,
		},
	}

	return ItemTemplate{
		Name:     "helmet_" + rarity.String(),
		Category: CategoryArmor,
		Type:     ItemHelmet,
		Parts:    parts,
	}
}

// PotionTemplate returns a template for potion consumables.
func PotionTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Bottle body
		{
			RelativeX:      0.5,
			RelativeY:      0.55,
			RelativeWidth:  0.50,
			RelativeHeight: 0.70,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
			ZIndex:         10,
			ColorRole:      "secondary",
			Opacity:        0.8,
			Rotation:       0,
		},
		// Liquid inside
		{
			RelativeX:      0.5,
			RelativeY:      0.60,
			RelativeWidth:  0.40,
			RelativeHeight: 0.50,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeWave},
			ZIndex:         15,
			ColorRole:      baseColor,
			Opacity:        0.9,
			Rotation:       0,
		},
		// Cork/stopper
		{
			RelativeX:      0.5,
			RelativeY:      0.22,
			RelativeWidth:  0.30,
			RelativeHeight: 0.18,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCircle},
			ZIndex:         20,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
	}

	return ItemTemplate{
		Name:     "potion_" + rarity.String(),
		Category: CategoryConsumable,
		Type:     ItemPotion,
		Parts:    parts,
	}
}

// ScrollTemplate returns a template for scroll consumables.
func ScrollTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Scroll body (rolled)
		{
			RelativeX:      0.5,
			RelativeY:      0.5,
			RelativeWidth:  0.70,
			RelativeHeight: 0.80,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
		// Endcap left
		{
			RelativeX:      0.5,
			RelativeY:      0.15,
			RelativeWidth:  0.60,
			RelativeHeight: 0.15,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeRectangle},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
		// Endcap right
		{
			RelativeX:      0.5,
			RelativeY:      0.85,
			RelativeWidth:  0.60,
			RelativeHeight: 0.15,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeRectangle},
			ZIndex:         15,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
	}

	return ItemTemplate{
		Name:     "scroll_" + rarity.String(),
		Category: CategoryConsumable,
		Type:     ItemScroll,
		Parts:    parts,
	}
}

// RingTemplate returns a template for ring accessories.
func RingTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Ring band
		{
			RelativeX:      0.5,
			RelativeY:      0.55,
			RelativeWidth:  0.75,
			RelativeHeight: 0.75,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRing, shapes.ShapeCircle},
			ZIndex:         10,
			ColorRole:      "secondary",
			Opacity:        1.0,
			Rotation:       0,
		},
		// Gem/stone
		{
			RelativeX:      0.5,
			RelativeY:      0.35,
			RelativeWidth:  0.35,
			RelativeHeight: 0.35,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeCrystal, shapes.ShapeStar, shapes.ShapeCircle},
			ZIndex:         15,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
	}

	// Add glow for higher rarities
	if rarity >= RarityEpic {
		parts = append(parts, ItemPartSpec{
			RelativeX:      0.5,
			RelativeY:      0.35,
			RelativeWidth:  0.50,
			RelativeHeight: 0.50,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeStar},
			ZIndex:         12,
			ColorRole:      GetRarityColorRole(rarity),
			Opacity:        0.5,
			Rotation:       0,
		})
	}

	return ItemTemplate{
		Name:     "ring_" + rarity.String(),
		Category: CategoryAccessory,
		Type:     ItemRing,
		Parts:    parts,
	}
}

// KeyTemplate returns a template for quest key items.
func KeyTemplate(rarity ItemRarity) ItemTemplate {
	baseColor := GetRarityColorRole(rarity)

	parts := []ItemPartSpec{
		// Key shaft
		{
			RelativeX:      0.5,
			RelativeY:      0.60,
			RelativeWidth:  0.20,
			RelativeHeight: 0.65,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
		// Key head (ring)
		{
			RelativeX:      0.5,
			RelativeY:      0.25,
			RelativeWidth:  0.40,
			RelativeHeight: 0.40,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRing, shapes.ShapeCircle},
			ZIndex:         10,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
		// Key teeth
		{
			RelativeX:      0.5,
			RelativeY:      0.88,
			RelativeWidth:  0.35,
			RelativeHeight: 0.18,
			ShapeTypes:     []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeLightning},
			ZIndex:         15,
			ColorRole:      baseColor,
			Opacity:        1.0,
			Rotation:       0,
		},
	}

	return ItemTemplate{
		Name:     "key_" + rarity.String(),
		Category: CategoryQuest,
		Type:     ItemKey,
		Parts:    parts,
	}
}

// SelectItemTemplate returns the appropriate item template based on item type and rarity.
func SelectItemTemplate(itemType ItemType, rarity ItemRarity) ItemTemplate {
	switch itemType {
	// Weapons
	case ItemSword:
		return SwordTemplate(rarity)
	case ItemAxe:
		return AxeTemplate(rarity)
	case ItemBow:
		return BowTemplate(rarity)
	case ItemStaff:
		return StaffTemplate(rarity)
	case ItemGun:
		return GunTemplate(rarity)

	// Armor
	case ItemHelmet:
		return HelmetTemplate(rarity)

	// Consumables
	case ItemPotion:
		return PotionTemplate(rarity)
	case ItemScroll:
		return ScrollTemplate(rarity)

	// Accessories
	case ItemRing:
		return RingTemplate(rarity)

	// Quest Items
	case ItemKey:
		return KeyTemplate(rarity)

	default:
		// Default to simple sword
		return SwordTemplate(rarity)
	}
}
