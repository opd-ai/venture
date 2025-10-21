// Package magic provides procedural generation of magic spells and abilities.
//
// The magic generation system creates diverse spells with varying types, elements,
// target patterns, and power levels. It supports both fantasy and sci-fi genres
// with appropriate theming.
//
// # Spell Types
//
// Spells are categorized into seven main types:
//   - Offensive: Damage-dealing spells (firebolt, lightning strike)
//   - Defensive: Protective spells (shields, barriers)
//   - Healing: Health restoration spells
//   - Buff: Stat-boosting spells for allies
//   - Debuff: Stat-reducing spells for enemies
//   - Utility: Non-combat spells (teleport, light)
//   - Summon: Entity summoning spells
//
// # Element System
//
// Spells have elemental affinities that affect damage types and effects:
//   - Fire: High damage, burning effects
//   - Ice: Moderate damage, slowing effects
//   - Lightning: Fast, chaining attacks
//   - Earth: High damage, stunning effects
//   - Wind: Speed and mobility
//   - Light: Holy damage, healing
//   - Dark: Shadow damage, debuffs
//   - Arcane: Pure magical energy
//   - None: Non-elemental magic
//
// # Targeting Patterns
//
// Spells can target different patterns:
//   - Self: Affects only the caster
//   - Single: Affects one target
//   - Area: Affects all targets in a radius
//   - Cone: Affects targets in a cone
//   - Line: Affects targets in a line
//   - All Allies: Affects all friendly targets
//   - All Enemies: Affects all hostile targets
//
// # Rarity System
//
// Spells have five rarity levels that affect their power:
//   - Common: Basic spells available early
//   - Uncommon: Improved versions with better stats
//   - Rare: Powerful spells with significant effects
//   - Epic: Very powerful spells with excellent stats
//   - Legendary: Unique, game-changing spells
//
// Higher rarity spells have:
//   - Increased damage/healing
//   - Lower cooldowns
//   - Faster cast times
//   - Better range and area effects
//   - Longer buff/debuff durations
//
// # Generation Parameters
//
// Spell generation is controlled by:
//   - Seed: Ensures deterministic generation
//   - Depth: Game progression level (affects power and rarity)
//   - Difficulty: Challenge multiplier (affects stat scaling)
//   - Genre: Fantasy or sci-fi theming
//   - Count: Number of spells to generate
//
// # Usage Example
//
//	gen := magic.NewSpellGenerator()
//	params := procgen.GenerationParams{
//		Difficulty: 0.5,
//		Depth:      10,
//		GenreID:    "fantasy",
//		Custom: map[string]interface{}{
//			"count": 20,
//		},
//	}
//	
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//		log.Fatal(err)
//	}
//	
//	spells := result.([]*magic.Spell)
//	for _, spell := range spells {
//		fmt.Printf("%s (%s): %s\n", 
//			spell.Name, spell.Rarity, spell.Description)
//	}
//
// # Stat Scaling
//
// Spell stats scale with multiple factors:
//   - Depth Scale: 1.0 + depth * 0.1
//   - Difficulty Scale: 0.8 + difficulty * 0.4
//   - Rarity Scale: 1.0 + rarity * 0.25
//
// This ensures spells remain balanced while becoming more powerful
// as the player progresses through the game.
//
// # Genre Differences
//
// Fantasy spells use traditional magical themes:
//   - Fire Bolt, Ice Storm, Lightning Strike
//   - Heal Touch, Mana Shield, Divine Blessing
//
// Sci-Fi spells use technological themes:
//   - Plasma Beam, Fusion Blast, Cryo Ray
//   - Nano Injection, Energy Barrier, Combat Stimulant
//
// # Determinism
//
// All generation is deterministic based on the seed value.
// The same seed and parameters will always produce identical spells,
// which is critical for multiplayer synchronization.
package magic
