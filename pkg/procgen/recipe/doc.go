// Package recipe provides procedural recipe generation for crafting systems.
//
// The recipe package implements deterministic, seed-based generation of crafting
// recipes for potions, enchantments, and magic items. Recipes are genre-themed
// and scale with skill requirements, depth, and difficulty parameters.
//
// Key Features:
//   - Deterministic generation: same seed + params = same recipes
//   - Genre-specific templates for fantasy, sci-fi, horror, cyberpunk, post-apocalyptic
//   - Three recipe types: potions (consumables), enchanting (item upgrades), magic items
//   - Rarity-based difficulty: common to legendary recipes with scaling requirements
//   - Balanced material costs and success chances
//
// Usage Example:
//
//	gen := recipe.NewRecipeGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      5,
//	    GenreID:    "fantasy",
//	    Custom: map[string]interface{}{
//	        "count": 10,
//	        "type":  "potion",
//	    },
//	}
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	recipes := result.([]*engine.Recipe)
//	for _, recipe := range recipes {
//	    fmt.Printf("Recipe: %s (Skill %d, Success %.0f%%)\n",
//	        recipe.Name, recipe.SkillRequired, recipe.BaseSuccessChance*100)
//	}
//
// Recipe Types:
//
// Potions: Quick crafts (3-8 seconds) with moderate success rates (70-85%).
// Produce consumables that restore health, mana, or provide buffs. Require
// herbs, fluids, and other alchemical materials.
//
// Enchanting: Moderate crafts (8-15 seconds) with good success rates (65-80%).
// Enhance existing weapons/armor with stat bonuses. Require enchantment scrolls,
// magical inks, and rare dusts.
//
// Magic Items: Long crafts (10-18 seconds) with lower success rates (60-70%).
// Create new weapons, wands, rings, and amulets with magical properties.
// Require base materials, magic cores, and crafting components.
//
// Template System:
//
// Each genre has dedicated recipe templates defining material pools, naming
// conventions, and stat ranges. Templates ensure thematic consistency while
// maintaining recipe variety through randomization within constraints.
//
// Rarity Progression:
//
// Recipe rarity affects skill requirements, material costs, success chances,
// and output quality. Distribution:
//   - Common: 50% (skill 0-5, 75-85% success)
//   - Uncommon: 30% (skill 3-7, 65-75% success)
//   - Rare: 15% (skill 5-10, 60-70% success)
//   - Epic: 4% (skill 8-15, 55-65% success)
//   - Legendary: 1% (skill 12+, 50-60% success)
//
// Depth and difficulty parameters shift these distributions, making higher-tier
// recipes more common in deeper dungeons and harder game modes.
package recipe
