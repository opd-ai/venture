package saveload

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// ItemToData converts an item.Item to ItemData for serialization.
func ItemToData(itm *item.Item) ItemData {
	if itm == nil {
		return ItemData{}
	}

	data := ItemData{
		Name:          itm.Name,
		Type:          itm.Type.String(),
		Rarity:        itm.Rarity.String(),
		Seed:          itm.Seed,
		Tags:          itm.Tags,
		Description:   itm.Description,
		Damage:        itm.Stats.Damage,
		Defense:       itm.Stats.Defense,
		AttackSpeed:   itm.Stats.AttackSpeed,
		Value:         itm.Stats.Value,
		Weight:        itm.Stats.Weight,
		RequiredLevel: itm.Stats.RequiredLevel,
		DurabilityMax: itm.Stats.DurabilityMax,
		Durability:    itm.Stats.Durability,
	}

	// Set type-specific fields
	switch itm.Type {
	case item.TypeWeapon:
		data.WeaponType = itm.WeaponType.String()
	case item.TypeArmor:
		data.ArmorType = itm.ArmorType.String()
	case item.TypeConsumable:
		data.ConsumableType = itm.ConsumableType.String()
	}

	return data
}

// DataToItem converts ItemData back to item.Item.
func DataToItem(data ItemData) *item.Item {
	itm := &item.Item{
		Name:        data.Name,
		Seed:        data.Seed,
		Tags:        data.Tags,
		Description: data.Description,
		Stats: item.Stats{
			Damage:        data.Damage,
			Defense:       data.Defense,
			AttackSpeed:   data.AttackSpeed,
			Value:         data.Value,
			Weight:        data.Weight,
			RequiredLevel: data.RequiredLevel,
			DurabilityMax: data.DurabilityMax,
			Durability:    data.Durability,
		},
	}

	// Parse type
	itm.Type = parseItemType(data.Type)
	itm.Rarity = parseItemRarity(data.Rarity)

	// Parse type-specific fields
	switch itm.Type {
	case item.TypeWeapon:
		itm.WeaponType = parseWeaponType(data.WeaponType)
	case item.TypeArmor:
		itm.ArmorType = parseArmorType(data.ArmorType)
	case item.TypeConsumable:
		itm.ConsumableType = parseConsumableType(data.ConsumableType)
	}

	return itm
}

// SpellToData converts a magic.Spell to SpellData for serialization.
func SpellToData(spell *magic.Spell) SpellData {
	if spell == nil {
		return SpellData{}
	}

	return SpellData{
		Name:        spell.Name,
		Type:        spell.Type.String(),
		Element:     spell.Element.String(),
		Target:      spell.Target.String(),
		Rarity:      spell.Rarity.String(),
		Seed:        spell.Seed,
		Tags:        spell.Tags,
		Description: spell.Description,
		Damage:      spell.Stats.Damage,
		Healing:     spell.Stats.Healing,
		ManaCost:    spell.Stats.ManaCost,
		Cooldown:    spell.Stats.Cooldown,
		CastTime:    spell.Stats.CastTime,
		Range:       spell.Stats.Range,
		AreaSize:    spell.Stats.AreaSize,
		Duration:    spell.Stats.Duration,
	}
}

// DataToSpell converts SpellData back to magic.Spell.
func DataToSpell(data SpellData) *magic.Spell {
	return &magic.Spell{
		Name:        data.Name,
		Type:        parseSpellType(data.Type),
		Element:     parseElementType(data.Element),
		Target:      parseTargetType(data.Target),
		Rarity:      parseMagicRarity(data.Rarity),
		Seed:        data.Seed,
		Tags:        data.Tags,
		Description: data.Description,
		Stats: magic.Stats{
			Damage:   data.Damage,
			Healing:  data.Healing,
			ManaCost: data.ManaCost,
			Cooldown: data.Cooldown,
			CastTime: data.CastTime,
			Range:    data.Range,
			AreaSize: data.AreaSize,
			Duration: data.Duration,
		},
	}
}

// Helper functions to parse string enums back to typed constants

func parseItemType(s string) item.ItemType {
	switch s {
	case "weapon":
		return item.TypeWeapon
	case "armor":
		return item.TypeArmor
	case "consumable":
		return item.TypeConsumable
	case "accessory":
		return item.TypeAccessory
	default:
		return item.TypeWeapon
	}
}

func parseItemRarity(s string) item.Rarity {
	switch s {
	case "common":
		return item.RarityCommon
	case "uncommon":
		return item.RarityUncommon
	case "rare":
		return item.RarityRare
	case "epic":
		return item.RarityEpic
	case "legendary":
		return item.RarityLegendary
	default:
		return item.RarityCommon
	}
}

func parseWeaponType(s string) item.WeaponType {
	switch s {
	case "sword":
		return item.WeaponSword
	case "axe":
		return item.WeaponAxe
	case "bow":
		return item.WeaponBow
	case "staff":
		return item.WeaponStaff
	case "dagger":
		return item.WeaponDagger
	case "spear":
		return item.WeaponSpear
	default:
		return item.WeaponSword
	}
}

func parseArmorType(s string) item.ArmorType {
	switch s {
	case "helmet":
		return item.ArmorHelmet
	case "chest":
		return item.ArmorChest
	case "legs":
		return item.ArmorLegs
	case "boots":
		return item.ArmorBoots
	case "gloves":
		return item.ArmorGloves
	case "shield":
		return item.ArmorShield
	default:
		return item.ArmorChest
	}
}

func parseConsumableType(s string) item.ConsumableType {
	switch s {
	case "potion":
		return item.ConsumablePotion
	case "scroll":
		return item.ConsumableScroll
	case "food":
		return item.ConsumableFood
	case "bomb":
		return item.ConsumableBomb
	default:
		return item.ConsumablePotion
	}
}

func parseSpellType(s string) magic.SpellType {
	switch s {
	case "offensive":
		return magic.TypeOffensive
	case "defensive":
		return magic.TypeDefensive
	case "utility":
		return magic.TypeUtility
	case "healing":
		return magic.TypeHealing
	case "buff":
		return magic.TypeBuff
	case "debuff":
		return magic.TypeDebuff
	case "summon":
		return magic.TypeSummon
	default:
		return magic.TypeOffensive
	}
}

func parseElementType(s string) magic.ElementType {
	switch s {
	case "none":
		return magic.ElementNone
	case "fire":
		return magic.ElementFire
	case "ice":
		return magic.ElementIce
	case "lightning":
		return magic.ElementLightning
	case "earth":
		return magic.ElementEarth
	case "wind":
		return magic.ElementWind
	case "light":
		return magic.ElementLight
	case "dark":
		return magic.ElementDark
	case "arcane":
		return magic.ElementArcane
	default:
		return magic.ElementNone
	}
}

func parseTargetType(s string) magic.TargetType {
	switch s {
	case "self":
		return magic.TargetSelf
	case "single":
		return magic.TargetSingle
	case "area":
		return magic.TargetArea
	case "cone":
		return magic.TargetCone
	case "line":
		return magic.TargetLine
	case "all_allies":
		return magic.TargetAllAllies
	case "all_enemies":
		return magic.TargetAllEnemies
	default:
		return magic.TargetSingle
	}
}

func parseMagicRarity(s string) magic.Rarity {
	switch s {
	case "common":
		return magic.RarityCommon
	case "uncommon":
		return magic.RarityUncommon
	case "rare":
		return magic.RarityRare
	case "epic":
		return magic.RarityEpic
	case "legendary":
		return magic.RarityLegendary
	default:
		return magic.RarityCommon
	}
}
