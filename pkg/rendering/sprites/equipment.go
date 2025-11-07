// Package sprites provides equipment visual helper functions.
package sprites

// GetMaterialTypeFromWeaponType determines material type based on weapon classification.
func GetMaterialTypeFromWeaponType(weaponType, genreID string) MaterialType {
	// Default materials based on weapon type
	switch weaponType {
	case "sword", "axe", "spear":
		return MaterialMetal
	case "bow", "crossbow":
		return MaterialWood
	case "staff", "wand":
		return MaterialWood
	case "gun":
		return MaterialMetal
	case "dagger":
		return MaterialMetal
	default:
		return MaterialMetal
	}
}

// GetMaterialTypeFromArmorType determines material type based on armor classification.
func GetMaterialTypeFromArmorType(armorType, genreID string) MaterialType {
	// Default materials based on armor type
	switch armorType {
	case "helmet", "chest", "boots", "gloves":
		// Heavy armor defaults to metal
		return MaterialMetal
	case "legs":
		// Legs can be leather or metal
		return MaterialLeather
	case "shield":
		return MaterialMetal
	default:
		return MaterialMetal
	}
}

// GetMaterialTypeFromTags determines material type from item tags and genre.
func GetMaterialTypeFromTags(tags []string, genreID string) MaterialType {
	// Check for explicit material tags
	for _, tag := range tags {
		switch tag {
		case "metal", "metallic", "steel", "iron", "silver", "gold":
			return MaterialMetal
		case "leather", "hide":
			return MaterialLeather
		case "cloth", "fabric", "robe":
			return MaterialCloth
		case "wood", "wooden":
			return MaterialWood
		case "crystal", "crystalline", "glass":
			return MaterialCrystal
		case "energy", "magical", "arcane":
			return MaterialEnergy
		}
	}

	// Genre-specific defaults for when no explicit tag is found
	switch genreID {
	case "fantasy":
		return MaterialMetal
	case "sci-fi":
		return MaterialEnergy
	case "horror":
		return MaterialWood
	case "cyberpunk":
		return MaterialMetal
	case "post-apocalyptic":
		return MaterialLeather
	default:
		return MaterialMetal
	}
}

// GetDamageStateFromDurability calculates damage state from durability percentage.
func GetDamageStateFromDurability(current, max int) DamageState {
	if max == 0 {
		return DamageStatePristine
	}

	percentage := float64(current) / float64(max)

	switch {
	case percentage >= 1.0:
		return DamageStatePristine
	case percentage >= 0.5:
		return DamageStateWorn
	case percentage >= 0.25:
		return DamageStateDamaged
	default:
		return DamageStateBroken
	}
}

// GetEnchantmentFromRarity creates enchantment glow based on item rarity.
func GetEnchantmentFromRarity(rarityStr string) EnchantmentGlow {
	enchant := EnchantmentGlow{
		Active:        false,
		Color:         "white",
		Intensity:     0.0,
		PulseSpeed:    0.0,
		ParticleCount: 0,
	}

	switch rarityStr {
	case "uncommon":
		enchant.Active = true
		enchant.Color = "green"
		enchant.Intensity = 0.2
		enchant.PulseSpeed = 0.5
		enchant.ParticleCount = 2
	case "rare":
		enchant.Active = true
		enchant.Color = "blue"
		enchant.Intensity = 0.4
		enchant.PulseSpeed = 0.7
		enchant.ParticleCount = 4
	case "epic":
		enchant.Active = true
		enchant.Color = "purple"
		enchant.Intensity = 0.6
		enchant.PulseSpeed = 1.0
		enchant.ParticleCount = 8
	case "legendary":
		enchant.Active = true
		enchant.Color = "gold"
		enchant.Intensity = 0.8
		enchant.PulseSpeed = 1.2
		enchant.ParticleCount = 12
	}

	return enchant
}

// GetDetailLevelFromRarity determines visual complexity based on rarity.
func GetDetailLevelFromRarity(rarityStr string) float64 {
	switch rarityStr {
	case "common":
		return 0.3
	case "uncommon":
		return 0.4
	case "rare":
		return 0.6
	case "epic":
		return 0.8
	case "legendary":
		return 1.0
	default:
		return 0.5
	}
}

// MaterialVisualProperties contains rendering properties for a material type.
type MaterialVisualProperties struct {
	// BaseColor is the primary color tone (can be overridden by item palette)
	BaseColor string

	// Sheen is the metallic/glossy appearance (0.0-1.0)
	Sheen float64

	// Roughness for texture (0.0=smooth, 1.0=rough)
	Roughness float64

	// PatternType for procedural texture ("none", "grain", "scales", "weave", "dots")
	PatternType string

	// PatternScale affects pattern size (0.5-2.0)
	PatternScale float64

	// Reflectivity for highlights (0.0-1.0)
	Reflectivity float64
}

// GetMaterialVisualProperties returns rendering properties for a material type.
func GetMaterialVisualProperties(material MaterialType) MaterialVisualProperties {
	switch material {
	case MaterialMetal:
		return MaterialVisualProperties{
			BaseColor:    "silver",
			Sheen:        0.9,
			Roughness:    0.2,
			PatternType:  "grain",
			PatternScale: 0.8,
			Reflectivity: 0.8,
		}
	case MaterialLeather:
		return MaterialVisualProperties{
			BaseColor:    "brown",
			Sheen:        0.3,
			Roughness:    0.6,
			PatternType:  "grain",
			PatternScale: 1.2,
			Reflectivity: 0.2,
		}
	case MaterialCloth:
		return MaterialVisualProperties{
			BaseColor:    "white",
			Sheen:        0.1,
			Roughness:    0.8,
			PatternType:  "weave",
			PatternScale: 1.5,
			Reflectivity: 0.1,
		}
	case MaterialWood:
		return MaterialVisualProperties{
			BaseColor:    "brown",
			Sheen:        0.4,
			Roughness:    0.5,
			PatternType:  "grain",
			PatternScale: 1.0,
			Reflectivity: 0.3,
		}
	case MaterialCrystal:
		return MaterialVisualProperties{
			BaseColor:    "cyan",
			Sheen:        1.0,
			Roughness:    0.1,
			PatternType:  "dots",
			PatternScale: 0.6,
			Reflectivity: 1.0,
		}
	case MaterialEnergy:
		return MaterialVisualProperties{
			BaseColor:    "blue",
			Sheen:        0.9,
			Roughness:    0.0,
			PatternType:  "none",
			PatternScale: 1.0,
			Reflectivity: 0.9,
		}
	default:
		return MaterialVisualProperties{
			BaseColor:    "gray",
			Sheen:        0.5,
			Roughness:    0.5,
			PatternType:  "none",
			PatternScale: 1.0,
			Reflectivity: 0.5,
		}
	}
}

// DamageVisualEffects contains rendering modifications for damage states.
type DamageVisualEffects struct {
	// OpacityMultiplier reduces opacity for broken items
	OpacityMultiplier float64

	// ColorDarken reduces brightness (0.0-1.0, 0=no change, 1=black)
	ColorDarken float64

	// CrackDensity for visual damage overlay (0.0-1.0)
	CrackDensity float64

	// EdgeRoughness increases edge jaggedness (0.0-1.0)
	EdgeRoughness float64

	// Dirtiness adds grime/dirt overlay (0.0-1.0)
	Dirtiness float64
}

// GetDamageVisualEffects returns rendering modifications for a damage state.
func GetDamageVisualEffects(state DamageState) DamageVisualEffects {
	switch state {
	case DamageStatePristine:
		return DamageVisualEffects{
			OpacityMultiplier: 1.0,
			ColorDarken:       0.0,
			CrackDensity:      0.0,
			EdgeRoughness:     0.0,
			Dirtiness:         0.0,
		}
	case DamageStateWorn:
		return DamageVisualEffects{
			OpacityMultiplier: 0.95,
			ColorDarken:       0.1,
			CrackDensity:      0.1,
			EdgeRoughness:     0.1,
			Dirtiness:         0.2,
		}
	case DamageStateDamaged:
		return DamageVisualEffects{
			OpacityMultiplier: 0.85,
			ColorDarken:       0.25,
			CrackDensity:      0.4,
			EdgeRoughness:     0.3,
			Dirtiness:         0.4,
		}
	case DamageStateBroken:
		return DamageVisualEffects{
			OpacityMultiplier: 0.7,
			ColorDarken:       0.4,
			CrackDensity:      0.7,
			EdgeRoughness:     0.5,
			Dirtiness:         0.6,
		}
	default:
		return DamageVisualEffects{
			OpacityMultiplier: 1.0,
			ColorDarken:       0.0,
			CrackDensity:      0.0,
			EdgeRoughness:     0.0,
			Dirtiness:         0.0,
		}
	}
}
