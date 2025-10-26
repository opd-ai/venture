package main

import (
	"flag"
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

var (
	seed    = flag.Int64("seed", 12345, "Generation seed")
	count   = flag.Int("count", 10, "Number of items to generate")
	depth   = flag.Int("depth", 5, "Dungeon depth for item generation")
	verbose = flag.Bool("verbose", false, "Show detailed output")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("inventorytest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"seed":  *seed,
		"count": *count,
		"depth": *depth,
	}).Info("Inventory Test Tool started")

	fmt.Println("=== Venture Inventory & Equipment System Demo ===")
	fmt.Printf("Seed: %d, Items: %d, Depth: %d\n\n", *seed, *count, *depth)

	// Create world and systems
	world := engine.NewWorld()
	invSystem := engine.NewInventorySystem(world)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	player.AddComponent(engine.NewEquipmentComponent())
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.StatsComponent{})
	player.AddComponent(&engine.AttackComponent{})
	world.Update(0.0)

	logger.WithField("playerID", player.ID).Info("player entity created")

	fmt.Println("Created player entity with inventory and equipment")
	fmt.Printf("- Inventory Capacity: 20 items, 100.0 kg\n")
	fmt.Printf("- Equipment Slots: 10 (weapons, armor, accessories)\n\n")

	// Generate items
	itemGen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      *depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	result, err := itemGen.Generate(*seed, params)
	if err != nil {
		logger.WithError(err).Fatal("item generation failed")
	}

	items := result.([]*item.Item)
	logger.WithField("itemCount", len(items)).Info("items generated")

	fmt.Printf("Generated %d items:\n", len(items))
	for i, itm := range items {
		fmt.Printf("%2d. %-30s %-12s Dmg:%-3d Def:%-3d Weight:%.1fkg Value:%d\n",
			i+1, itm.Name, itm.Type.String(), itm.Stats.Damage, itm.Stats.Defense,
			itm.Stats.Weight, itm.Stats.Value)
	}
	fmt.Println()

	// Add items to inventory
	fmt.Println("Adding items to player inventory...")
	addedCount := 0
	for _, itm := range items {
		success, err := invSystem.AddItemToInventory(player.ID, itm)
		if err != nil {
			log.Printf("Error adding item: %v", err)
			continue
		}
		if success {
			addedCount++
		} else {
			fmt.Printf("  Inventory full! Could not add: %s\n", itm.Name)
			break
		}
	}
	fmt.Printf("Successfully added %d/%d items to inventory\n\n", addedCount, len(items))

	// Get inventory component
	comp, _ := player.GetComponent("inventory")
	inv := comp.(*engine.InventoryComponent)

	// Show inventory status
	fmt.Println("Inventory Status:")
	fmt.Printf("- Items: %d/%d\n", inv.GetItemCount(), inv.MaxItems)
	fmt.Printf("- Weight: %.1f/%.1f kg\n", inv.GetCurrentWeight(), inv.MaxWeight)
	totalValue, _ := invSystem.GetInventoryValue(player.ID)
	fmt.Printf("- Total Value: %d gold\n\n", totalValue)

	// Equip some items
	fmt.Println("Attempting to equip items...")
	equippedCount := 0
	for i := 0; i < inv.GetItemCount(); i++ {
		itm := inv.Items[i]
		if itm.IsEquippable() {
			err := invSystem.EquipItem(player.ID, i)
			if err != nil {
				continue
			}
			equippedCount++
			fmt.Printf("  Equipped: %s (%s)\n", itm.Name, itm.Type.String())
			i--
			if equippedCount >= 5 {
				break
			}
		}
	}
	fmt.Printf("Equipped %d items\n\n", equippedCount)

	// Show equipment status
	comp2, _ := player.GetComponent("equipment")
	equip := comp2.(*engine.EquipmentComponent)

	fmt.Println("Equipment Status:")
	for slot := engine.SlotMainHand; slot <= engine.SlotAccessory3; slot++ {
		equipped := equip.GetEquipped(slot)
		if equipped != nil {
			fmt.Printf("  %-15s: %s (Dmg:%d Def:%d)\n",
				slot.String(), equipped.Name,
				equipped.Stats.Damage, equipped.Stats.Defense)
		}
	}
	fmt.Println()

	// Show calculated stats
	equipStats := equip.GetStats()
	fmt.Println("Equipment Bonuses:")
	fmt.Printf("- Total Damage: %d\n", equipStats.Damage)
	fmt.Printf("- Total Defense: %d\n", equipStats.Defense)
	fmt.Printf("- Attack Speed: %.2fx\n", equipStats.AttackSpeed)
	fmt.Printf("- Total Weight: %.1f kg\n", equipStats.Weight)
	fmt.Printf("- Total Value: %d gold\n\n", equipStats.Value)

	// Demonstrate consumable usage
	fmt.Println("Looking for consumables to use...")
	for i, itm := range inv.Items {
		if itm.IsConsumable() {
			fmt.Printf("Using consumable: %s\n", itm.Name)
			err := invSystem.UseConsumable(player.ID, i)
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
			} else {
				fmt.Printf("  Successfully used %s\n", itm.Name)
				comp3, _ := player.GetComponent("health")
				health := comp3.(*engine.HealthComponent)
				fmt.Printf("  Player Health: %.0f/%.0f\n", health.Current, health.Max)
			}
			break
		}
	}
	fmt.Println()

	// Sort inventory by value
	fmt.Println("Sorting inventory by value (descending)...")
	invSystem.SortInventoryByValue(player.ID)
	fmt.Println("Top 5 most valuable items in inventory:")
	for i := 0; i < 5 && i < len(inv.Items); i++ {
		itm := inv.Items[i]
		fmt.Printf("%2d. %-30s Value:%d gold\n", i+1, itm.Name, itm.Stats.Value)
	}
	fmt.Println()

	// Final inventory status
	fmt.Println("Final Inventory Summary:")
	fmt.Printf("- Items Carried: %d/%d\n", inv.GetItemCount(), inv.MaxItems)
	fmt.Printf("- Weight Carried: %.1f/%.1f kg\n", inv.GetCurrentWeight(), inv.MaxWeight)
	totalValue, _ = invSystem.GetInventoryValue(player.ID)
	fmt.Printf("- Total Wealth: %d gold (inventory + equipment)\n", totalValue)

	fmt.Println("\n=== Demo Complete ===")
}
