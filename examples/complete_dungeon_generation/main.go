package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// This example demonstrates how to integrate terrain, entity, and item generation
// to create a complete dungeon level with enemies and loot.

func main() {
	fmt.Println("=== Venture - Complete Dungeon Generation Example ===")

	// Use a fixed seed for reproducible results
	baseSeed := int64(12345)
	depth := 5 // Dungeon level

	// Step 1: Generate terrain
	fmt.Println("Step 1: Generating terrain...")
	terrainGen := terrain.NewBSPGenerator()
	terrainParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		Custom: map[string]interface{}{
			"width":  60,
			"height": 40,
		},
	}

	terrainResult, err := terrainGen.Generate(baseSeed, terrainParams)
	if err != nil {
		log.Fatalf("Terrain generation failed: %v", err)
	}

	terr := terrainResult.(*terrain.Terrain)
	fmt.Printf("✓ Generated terrain: %dx%d with %d rooms\n", terr.Width, terr.Height, len(terr.Rooms))

	// Step 2: Generate entities (monsters for rooms)
	fmt.Println("\nStep 2: Generating entities...")
	entityGen := entity.NewEntityGenerator()
	entityParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": len(terr.Rooms), // One entity per room
		},
	}

	entityResult, err := entityGen.Generate(baseSeed+1, entityParams)
	if err != nil {
		log.Fatalf("Entity generation failed: %v", err)
	}

	entities := entityResult.([]*entity.Entity)
	fmt.Printf("✓ Generated %d entities\n", len(entities))

	// Step 3: Generate items (loot for the dungeon)
	fmt.Println("\nStep 3: Generating items...")
	itemGen := item.NewItemGenerator()
	itemParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": len(terr.Rooms) * 2, // 2 items per room on average
		},
	}

	itemResult, err := itemGen.Generate(baseSeed+2, itemParams)
	if err != nil {
		log.Fatalf("Item generation failed: %v", err)
	}

	items := itemResult.([]*item.Item)
	fmt.Printf("✓ Generated %d items\n", len(items))

	// Step 4: Display dungeon layout
	fmt.Println("\n" + separator(70))
	fmt.Println("Dungeon Overview")
	fmt.Println(separator(70))

	// Show room assignments
	for i, room := range terr.Rooms {
		cx, cy := room.Center()
		fmt.Printf("\nRoom %d: [%d,%d] Size: %dx%d Center: (%d,%d)\n",
			i+1, room.X, room.Y, room.Width, room.Height, cx, cy)

		// Assign entity to room
		if i < len(entities) {
			ent := entities[i]
			threatLevel := getThreatIndicator(ent.GetThreatLevel())
			fmt.Printf("  👹 Enemy: %s (%s) Level %d %s\n",
				ent.Name, ent.Type, ent.Stats.Level, threatLevel)
			fmt.Printf("     HP: %d | Damage: %d | Defense: %d\n",
				ent.Stats.Health, ent.Stats.Damage, ent.Stats.Defense)
		}

		// Assign items to room (2 items per room)
		itemsPerRoom := 2
		startIdx := i * itemsPerRoom
		endIdx := startIdx + itemsPerRoom
		if endIdx > len(items) {
			endIdx = len(items)
		}

		if startIdx < len(items) {
			fmt.Println("  💎 Loot:")
			for j := startIdx; j < endIdx; j++ {
				itm := items[j]
				rarityIcon := getRarityIcon(itm.Rarity)
				typeIcon := getItemTypeIcon(itm.Type)
				fmt.Printf("     %s %s %s (%s)\n",
					rarityIcon, typeIcon, itm.Name, itm.Rarity)
				if itm.Type == item.TypeWeapon {
					fmt.Printf("        Damage: %d | Speed: %.2f | Value: %d gold\n",
						itm.Stats.Damage, itm.Stats.AttackSpeed, itm.Stats.Value)
				} else if itm.Type == item.TypeArmor {
					fmt.Printf("        Defense: %d | Value: %d gold\n",
						itm.Stats.Defense, itm.Stats.Value)
				}
			}
		}
	}

	// Step 5: Display statistics
	fmt.Println("\n" + separator(70))
	fmt.Println("Dungeon Statistics")
	fmt.Println(separator(70))

	// Entity statistics
	bossCount := 0
	monsterCount := 0
	minionCount := 0
	avgLevel := 0
	totalThreat := 0

	for _, ent := range entities {
		avgLevel += ent.Stats.Level
		totalThreat += ent.GetThreatLevel()
		switch ent.Type {
		case entity.TypeBoss:
			bossCount++
		case entity.TypeMonster:
			monsterCount++
		case entity.TypeMinion:
			minionCount++
		}
	}

	if len(entities) > 0 {
		avgLevel /= len(entities)
		avgThreat := totalThreat / len(entities)

		fmt.Printf("\nEntity Breakdown:\n")
		fmt.Printf("  Bosses:   %d\n", bossCount)
		fmt.Printf("  Monsters: %d\n", monsterCount)
		fmt.Printf("  Minions:  %d\n", minionCount)
		fmt.Printf("  Avg Level: %d\n", avgLevel)
		fmt.Printf("  Avg Threat: %d/100\n", avgThreat)
	}

	// Item statistics
	weaponCount := 0
	armorCount := 0
	consumableCount := 0
	rarityCount := make(map[item.Rarity]int)
	totalValue := 0

	for _, itm := range items {
		totalValue += itm.Stats.Value
		rarityCount[itm.Rarity]++
		switch itm.Type {
		case item.TypeWeapon:
			weaponCount++
		case item.TypeArmor:
			armorCount++
		case item.TypeConsumable:
			consumableCount++
		}
	}

	fmt.Printf("\nItem Breakdown:\n")
	fmt.Printf("  Weapons:    %d\n", weaponCount)
	fmt.Printf("  Armor:      %d\n", armorCount)
	fmt.Printf("  Consumables: %d\n", consumableCount)
	fmt.Printf("  Total Value: %d gold\n", totalValue)

	fmt.Printf("\nRarity Distribution:\n")
	for _, rarity := range []item.Rarity{
		item.RarityCommon,
		item.RarityUncommon,
		item.RarityRare,
		item.RarityEpic,
		item.RarityLegendary,
	} {
		count := rarityCount[rarity]
		if count > 0 {
			fmt.Printf("  %s %-12s: %d\n", getRarityIcon(rarity), rarity, count)
		}
	}

	fmt.Println("\n" + separator(70))
	fmt.Println("✓ Dungeon generation complete!")
	fmt.Printf("  Seed: %d | Depth: %d | Rooms: %d\n", baseSeed, depth, len(terr.Rooms))
	fmt.Println(separator(70))
}

func separator(width int) string {
	result := ""
	for i := 0; i < width; i++ {
		result += "="
	}
	return result
}

func getThreatIndicator(threat int) string {
	switch {
	case threat < 20:
		return "⚪ Low"
	case threat < 40:
		return "🟢 Medium"
	case threat < 60:
		return "🟡 High"
	case threat < 80:
		return "🟠 Dangerous"
	default:
		return "🔴 Deadly"
	}
}

func getRarityIcon(r item.Rarity) string {
	switch r {
	case item.RarityCommon:
		return "⚪"
	case item.RarityUncommon:
		return "🟢"
	case item.RarityRare:
		return "🔵"
	case item.RarityEpic:
		return "🟣"
	case item.RarityLegendary:
		return "🟠"
	default:
		return "  "
	}
}

func getItemTypeIcon(t item.ItemType) string {
	switch t {
	case item.TypeWeapon:
		return "⚔️"
	case item.TypeArmor:
		return "🛡️"
	case item.TypeConsumable:
		return "🧪"
	case item.TypeAccessory:
		return "💍"
	default:
		return "📦"
	}
}
