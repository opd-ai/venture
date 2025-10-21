// Package item provides procedural item generation for the Venture RPG.
//
// The item package generates weapons, armor, consumables, and accessories with
// procedurally generated names, stats, and properties. All generation is
// deterministic based on seed values, ensuring consistency across game sessions.
//
// # Item Types
//
// The package supports four main item types:
//   - Weapons: Offensive equipment with damage and attack speed
//   - Armor: Defensive equipment with defense values
//   - Consumables: Single-use items like potions and scrolls
//   - Accessories: Stat-boosting equipment
//
// # Generation
//
// Items are generated using the ItemGenerator, which implements the
// procgen.Generator interface. Generation is controlled by parameters:
//   - Seed: Determines all random values
//   - Depth: Affects item level and stat scaling
//   - Difficulty: Modifies stat values
//   - GenreID: Selects templates (fantasy, scifi)
//
// # Rarity System
//
// Items have five rarity levels that affect stats and value:
//   - Common: Base stats
//   - Uncommon: 20% stat boost
//   - Rare: 50% stat boost
//   - Epic: 100% stat boost
//   - Legendary: 200% stat boost
//
// Higher depths increase the chance of rare items.
//
// # Example Usage
//
//	generator := item.NewItemGenerator()
//	params := procgen.GenerationParams{
//		Depth: 5,
//		Difficulty: 0.5,
//		GenreID: "fantasy",
//		Custom: map[string]interface{}{
//			"count": 20,
//			"type": "weapon",
//		},
//	}
//	result, err := generator.Generate(12345, params)
//	if err != nil {
//		log.Fatal(err)
//	}
//	items := result.([]*item.Item)
//	for _, item := range items {
//		fmt.Printf("%s (%s): Damage=%d, Value=%d\n",
//			item.Name, item.Rarity, item.Stats.Damage, item.Stats.Value)
//	}
package item
