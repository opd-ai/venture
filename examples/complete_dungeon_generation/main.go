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
	baseSeed := int64(12345)
	depth := 5

	terr := generateTerrain(baseSeed, depth)
	entities := generateEntities(baseSeed, depth, terr)
	items := generateItems(baseSeed, depth, terr)
	displayDungeonOverview(terr, entities, items)
	displayStatistics(entities, items, terr, baseSeed, depth)
}

// generateTerrain creates the dungeon terrain using BSP generation.
func generateTerrain(seed int64, depth int) *terrain.Terrain {
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
	terrainResult, err := terrainGen.Generate(seed, terrainParams)
	if err != nil {
		log.Fatalf("Terrain generation failed: %v", err)
	}
	terr := terrainResult.(*terrain.Terrain)
	fmt.Printf("✓ Generated terrain: %dx%d with %d rooms\n", terr.Width, terr.Height, len(terr.Rooms))
	return terr
}

// generateEntities creates entities (monsters) for dungeon rooms.
func generateEntities(seed int64, depth int, terr *terrain.Terrain) []*entity.Entity {
	fmt.Println("\nStep 2: Generating entities...")
	entityGen := entity.NewEntityGenerator()
	entityParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": len(terr.Rooms),
		},
	}
	entityResult, err := entityGen.Generate(seed+1, entityParams)
	if err != nil {
		log.Fatalf("Entity generation failed: %v", err)
	}
	entities := entityResult.([]*entity.Entity)
	fmt.Printf("✓ Generated %d entities\n", len(entities))
	return entities
}

// generateItems creates loot items for the dungeon.
func generateItems(seed int64, depth int, terr *terrain.Terrain) []*item.Item {
	fmt.Println("\nStep 3: Generating items...")
	itemGen := item.NewItemGenerator()
	itemParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": len(terr.Rooms) * 2,
		},
	}
	itemResult, err := itemGen.Generate(seed+2, itemParams)
	if err != nil {
		log.Fatalf("Item generation failed: %v", err)
	}
	items := itemResult.([]*item.Item)
	fmt.Printf("✓ Generated %d items\n", len(items))
	return items
}

// displayDungeonOverview prints room-by-room layout with entities and loot.
func displayDungeonOverview(terr *terrain.Terrain, entities []*entity.Entity, items []*item.Item) {
	fmt.Println("\n" + separator(70))
	fmt.Println("Dungeon Overview")
	fmt.Println(separator(70))
	for i, room := range terr.Rooms {
		displayRoom(i, room, entities, items)
	}
}

// displayRoom prints details for a single room including entity and loot.
func displayRoom(index int, room *terrain.Room, entities []*entity.Entity, items []*item.Item) {
	cx, cy := room.Center()
	fmt.Printf("\nRoom %d: [%d,%d] Size: %dx%d Center: (%d,%d)\n",
		index+1, room.X, room.Y, room.Width, room.Height, cx, cy)
	if index < len(entities) {
		displayRoomEntity(entities[index])
	}
	displayRoomLoot(index, items)
}

// displayRoomEntity prints entity information for a room.
func displayRoomEntity(ent *entity.Entity) {
	threatLevel := getThreatIndicator(ent.GetThreatLevel())
	fmt.Printf("  👹 Enemy: %s (%s) Level %d %s\n", ent.Name, ent.Type, ent.Stats.Level, threatLevel)
	fmt.Printf("     HP: %d | Damage: %d | Defense: %d\n",
		ent.Stats.Health, ent.Stats.Damage, ent.Stats.Defense)
}

// displayRoomLoot prints loot items for a room.
func displayRoomLoot(roomIndex int, items []*item.Item) {
	itemsPerRoom := 2
	startIdx := roomIndex * itemsPerRoom
	endIdx := startIdx + itemsPerRoom
	if endIdx > len(items) {
		endIdx = len(items)
	}
	if startIdx < len(items) {
		fmt.Println("  💎 Loot:")
		for j := startIdx; j < endIdx; j++ {
			displayItem(items[j])
		}
	}
}

// displayItem prints detailed information for a single item.
func displayItem(itm *item.Item) {
	rarityIcon := getRarityIcon(itm.Rarity)
	typeIcon := getItemTypeIcon(itm.Type)
	fmt.Printf("     %s %s %s (%s)\n", rarityIcon, typeIcon, itm.Name, itm.Rarity)
	if itm.Type == item.TypeWeapon {
		fmt.Printf("        Damage: %d | Speed: %.2f | Value: %d gold\n",
			itm.Stats.Damage, itm.Stats.AttackSpeed, itm.Stats.Value)
	} else if itm.Type == item.TypeArmor {
		fmt.Printf("        Defense: %d | Value: %d gold\n",
			itm.Stats.Defense, itm.Stats.Value)
	}
}

// displayStatistics prints comprehensive dungeon statistics.
func displayStatistics(entities []*entity.Entity, items []*item.Item, terr *terrain.Terrain, seed int64, depth int) {
	fmt.Println("\n" + separator(70))
	fmt.Println("Dungeon Statistics")
	fmt.Println(separator(70))
	displayEntityStatistics(entities)
	displayItemStatistics(items)
	displayCompletionSummary(terr, seed, depth)
}

// displayEntityStatistics prints entity breakdown and averages.
func displayEntityStatistics(entities []*entity.Entity) {
	stats := calculateEntityStats(entities)
	fmt.Printf("\nEntity Breakdown:\n")
	fmt.Printf("  Bosses:   %d\n", stats.bossCount)
	fmt.Printf("  Monsters: %d\n", stats.monsterCount)
	fmt.Printf("  Minions:  %d\n", stats.minionCount)
	fmt.Printf("  Avg Level: %d\n", stats.avgLevel)
	fmt.Printf("  Avg Threat: %d/100\n", stats.avgThreat)
}

type entityStats struct {
	bossCount    int
	monsterCount int
	minionCount  int
	avgLevel     int
	avgThreat    int
}

// calculateEntityStats computes entity statistics.
func calculateEntityStats(entities []*entity.Entity) entityStats {
	var stats entityStats
	totalLevel := 0
	totalThreat := 0
	for _, ent := range entities {
		totalLevel += ent.Stats.Level
		totalThreat += ent.GetThreatLevel()
		switch ent.Type {
		case entity.TypeBoss:
			stats.bossCount++
		case entity.TypeMonster:
			stats.monsterCount++
		case entity.TypeMinion:
			stats.minionCount++
		}
	}
	if len(entities) > 0 {
		stats.avgLevel = totalLevel / len(entities)
		stats.avgThreat = totalThreat / len(entities)
	}
	return stats
}

// displayItemStatistics prints item breakdown and rarity distribution.
func displayItemStatistics(items []*item.Item) {
	weaponCount, armorCount, consumableCount, totalValue, rarityCount := calculateItemStats(items)
	fmt.Printf("\nItem Breakdown:\n")
	fmt.Printf("  Weapons:    %d\n", weaponCount)
	fmt.Printf("  Armor:      %d\n", armorCount)
	fmt.Printf("  Consumables: %d\n", consumableCount)
	fmt.Printf("  Total Value: %d gold\n", totalValue)
	displayRarityDistribution(rarityCount)
}

// calculateItemStats computes item type and rarity statistics.
func calculateItemStats(items []*item.Item) (int, int, int, int, map[item.Rarity]int) {
	weaponCount := 0
	armorCount := 0
	consumableCount := 0
	totalValue := 0
	rarityCount := make(map[item.Rarity]int)
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
	return weaponCount, armorCount, consumableCount, totalValue, rarityCount
}

// displayRarityDistribution prints rarity counts for all rarities.
func displayRarityDistribution(rarityCount map[item.Rarity]int) {
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
}

// displayCompletionSummary prints final dungeon generation summary.
func displayCompletionSummary(terr *terrain.Terrain, seed int64, depth int) {
	fmt.Println("\n" + separator(70))
	fmt.Println("✓ Dungeon generation complete!")
	fmt.Printf("  Seed: %d | Depth: %d | Rooms: %d\n", seed, depth, len(terr.Rooms))
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
