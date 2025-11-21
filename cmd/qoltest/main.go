package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine/qol"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, autoloot, crafting, guild, mount, storage, recipe, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("=== Quality of Life Features Test ===")

	switch *mode {
	case "demo":
		runDemo(*verbose)
	case "autoloot":
		testAutoLoot(*verbose)
	case "crafting":
		testCrafting(*verbose)
	case "guild":
		testGuildInvitations(*verbose)
	case "mount":
		testMountWhistle(*verbose)
	case "storage":
		testStorageSorting(*verbose)
	case "recipe":
		testRecipeTracking(*verbose)
	case "all":
		runDemo(*verbose)
		testAutoLoot(*verbose)
		testCrafting(*verbose)
		testGuildInvitations(*verbose)
		testMountWhistle(*verbose)
		testStorageSorting(*verbose)
		testRecipeTracking(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
	}
}

func runDemo(verbose bool) {
	fmt.Println("--- QoL Features Demo ---")
	fmt.Println("This demo showcases all 6 quality of life features:")
	fmt.Println("1. Auto-Loot: Companions collect nearby items automatically")
	fmt.Println("2. Smart Crafting: Queue multiple recipes for batch processing")
	fmt.Println("3. Guild Invitations: Offline member acceptance system")
	fmt.Println("4. Mount Whistle: Summon nearby vehicles instantly")
	fmt.Println("5. Storage Sorting: Auto-organize inventory by type/rarity")
	fmt.Println("6. Recipe Tracking: Show missing materials for crafts")
	fmt.Println()
}

func testAutoLoot(verbose bool) {
	fmt.Println("--- Auto-Loot System Test ---")

	mgr := qol.NewAutoLootManager()

	config := qol.DefaultAutoLootConfig(1)
	config.Radius = 8.0
	config.MinRarity = 1
	config.IgnoreTypes = []string{"junk", "trash"}

	mgr.SetConfig(config)

	fmt.Printf("Companion #1 auto-loot configured:\n")
	fmt.Printf("  Radius: %.1f tiles\n", config.Radius)
	fmt.Printf("  Min Rarity: %d (Uncommon+)\n", config.MinRarity)
	fmt.Printf("  Ignored Types: %v\n", config.IgnoreTypes)
	fmt.Printf("  Max Per Cycle: %d items\n", config.MaxPerCycle)
	fmt.Println()

	testItems := []struct {
		name     string
		rarity   int
		itemType string
	}{
		{"Iron Sword", 0, "weapon"},
		{"Rare Helm", 2, "armor"},
		{"Epic Staff", 3, "weapon"},
		{"Junk Item", 4, "junk"},
	}

	fmt.Println("Testing item collection:")
	for _, item := range testItems {
		shouldCollect := mgr.ShouldCollect(1, item.rarity, item.itemType)
		status := "SKIP"
		if shouldCollect {
			status = "COLLECT"
		}
		fmt.Printf("  [%s] %s (Rarity: %d, Type: %s)\n", status, item.name, item.rarity, item.itemType)
	}
	fmt.Println()
}

func testCrafting(verbose bool) {
	fmt.Println("--- Smart Crafting Queue Test ---")

	mgr := qol.NewCraftQueueManager()

	recipes := []struct {
		id  string
		qty int
	}{
		{"iron_sword", 5},
		{"health_potion", 10},
		{"leather_armor", 3},
	}

	fmt.Println("Adding recipes to queue:")
	for _, recipe := range recipes {
		err := mgr.AddRecipe(1, recipe.id, recipe.qty)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else {
			fmt.Printf("  [ADDED] %s x%d\n", recipe.id, recipe.qty)
		}
	}
	fmt.Println()

	queue := mgr.GetQueue(1)
	fmt.Printf("Current queue (%d recipes):\n", len(queue))
	for i, entry := range queue {
		fmt.Printf("  %d. %s x%d (added %s ago)\n",
			i+1, entry.RecipeID, entry.Quantity,
			time.Since(entry.AddedAt).Round(time.Second))
	}
	fmt.Println()

	fmt.Println("Removing recipe at position 1...")
	err := mgr.RemoveRecipe(1, 1)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Println("Recipe removed successfully")
	}

	queue = mgr.GetQueue(1)
	fmt.Printf("Updated queue (%d recipes):\n", len(queue))
	for i, entry := range queue {
		fmt.Printf("  %d. %s x%d\n", i+1, entry.RecipeID, entry.Quantity)
	}
	fmt.Println()
}

func testGuildInvitations(verbose bool) {
	fmt.Println("--- Guild Invitations Test ---")

	mgr := qol.NewGuildInvitationManager()

	inv1 := &qol.GuildInvitation{
		InvitationID: "inv_001",
		GuildID:      "guild_001",
		GuildName:    "Elite Warriors",
		InviterID:    "player_001",
		InviterName:  "Alice",
		InviteeID:    "player_002",
		InviteeName:  "Bob",
		Message:      "Join our guild and conquer dungeons together!",
	}

	inv2 := &qol.GuildInvitation{
		InvitationID: "inv_002",
		GuildID:      "guild_002",
		GuildName:    "Merchant Alliance",
		InviterID:    "player_003",
		InviterName:  "Charlie",
		InviteeID:    "player_002",
		InviteeName:  "Bob",
		Message:      "We need skilled traders!",
		ExpiresAt:    time.Now().Add(-24 * time.Hour), // Expired
	}

	mgr.SendInvitation(inv1)
	mgr.SendInvitation(inv2)

	fmt.Println("Sent invitations:")
	fmt.Printf("  1. %s -> %s (%s)\n", inv1.GuildName, inv1.InviteeName, inv1.Message)
	fmt.Printf("     Days until expiry: %.1f\n", inv1.DaysUntilExpiry())
	fmt.Printf("  2. %s -> %s (EXPIRED)\n", inv2.GuildName, inv2.InviteeName)
	fmt.Println()

	pending := mgr.GetPendingInvitations("player_002")
	fmt.Printf("Bob's pending invitations: %d\n", len(pending))
	for _, inv := range pending {
		fmt.Printf("  - %s (%.1f days remaining)\n", inv.GuildName, inv.DaysUntilExpiry())
	}
	fmt.Println()

	fmt.Println("Bob accepts invitation from Elite Warriors...")
	err := mgr.AcceptInvitation("inv_001")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Invitation accepted! Welcome to %s\n", inv1.GuildName)
	}
	fmt.Println()

	fmt.Println("Cleaning up expired invitations...")
	removed := mgr.CleanupExpired()
	fmt.Printf("Removed %d expired invitations\n", removed)
	fmt.Println()
}

func testMountWhistle(verbose bool) {
	fmt.Println("--- Mount Whistle Test ---")

	mgr := qol.NewMountWhistleManager()

	summons := []struct {
		name       string
		vehicleID  uint64
		currentPos [2]float64
		targetPos  [2]float64
	}{
		{"Horse", 101, [2]float64{0, 0}, [2]float64{3, 4}},
		{"Dragon", 102, [2]float64{10, 10}, [2]float64{12, 10}},
		{"Airship", 103, [2]float64{0, 0}, [2]float64{10, 10}},
	}

	fmt.Println("Summoning vehicles:")
	for _, s := range summons {
		summon := &qol.MountSummon{
			PlayerID:    1,
			VehicleID:   s.vehicleID,
			VehicleType: s.name,
			CurrentPos:  s.currentPos,
			TargetPos:   s.targetPos,
		}

		mgr.SummonMount(summon)

		fmt.Printf("  %s (ID: %d)\n", s.name, s.vehicleID)
		fmt.Printf("    Distance: %.2f tiles\n", summon.Distance)
		fmt.Printf("    ETA: %.1f seconds\n", summon.EstimatedTime)

		if verbose {
			fmt.Printf("    From: (%.1f, %.1f) -> To: (%.1f, %.1f)\n",
				s.currentPos[0], s.currentPos[1],
				s.targetPos[0], s.targetPos[1])
		}
	}
	fmt.Println()

	fmt.Println("Completing Horse summon...")
	mgr.CompleteSummon(1)

	active := mgr.GetActiveSummon(1)
	if active != nil && active.Completed {
		fmt.Printf("%s has arrived!\n", active.VehicleType)
	}
	fmt.Println()
}

func testStorageSorting(verbose bool) {
	fmt.Println("--- Storage Sorting Test ---")

	sorter := qol.NewStorageSorter()

	items := []*qol.Item{
		{ID: "1", Name: "Iron Sword", Type: "weapon", Rarity: 1, Value: 100, Quantity: 1},
		{ID: "2", Name: "Legendary Staff", Type: "weapon", Rarity: 4, Value: 1000, Quantity: 1},
		{ID: "3", Name: "Health Potion", Type: "consumable", Rarity: 0, Value: 50, Quantity: 10},
		{ID: "4", Name: "Epic Helm", Type: "armor", Rarity: 3, Value: 500, Quantity: 1},
		{ID: "5", Name: "Rare Boots", Type: "armor", Rarity: 2, Value: 200, Quantity: 1},
	}

	fmt.Printf("Original inventory (%d items):\n", len(items))
	for i, item := range items {
		fmt.Printf("  %d. %s (Rarity: %d, Value: %d, Qty: %d)\n",
			i+1, item.Name, item.Rarity, item.Value, item.Quantity)
	}
	fmt.Println()

	sortCriteria := []qol.SortCriteria{
		qol.SortByRarity,
		qol.SortByValue,
		qol.SortByType,
	}

	for _, criteria := range sortCriteria {
		itemsCopy := make([]*qol.Item, len(items))
		copy(itemsCopy, items)

		sorter.SortItems(itemsCopy, criteria)

		fmt.Printf("Sorted by %s:\n", criteria.String())
		for i, item := range itemsCopy {
			fmt.Printf("  %d. %s\n", i+1, item.Name)
		}
		fmt.Println()
	}

	fmt.Println("Available presets:")
	for _, name := range []string{"default", "rarity", "value"} {
		preset := sorter.GetPreset(name)
		if preset != nil {
			fmt.Printf("  - %s: Primary=%s, Secondary=%s\n",
				preset.Name, preset.PrimaryCriteria.String(), preset.SecondaryCriteria.String())
		}
	}
	fmt.Println()
}

func testRecipeTracking(verbose bool) {
	fmt.Println("--- Recipe Tracking Test ---")

	tracker := qol.NewRecipeTracker()

	recipes := []struct {
		id        string
		name      string
		required  map[string]int
		available map[string]int
	}{
		{
			id:        "iron_sword",
			name:      "Iron Sword",
			required:  map[string]int{"iron": 3, "wood": 1},
			available: map[string]int{"iron": 10, "wood": 5},
		},
		{
			id:        "health_potion",
			name:      "Health Potion",
			required:  map[string]int{"herb": 2, "water": 1},
			available: map[string]int{"herb": 1, "water": 10},
		},
		{
			id:        "leather_armor",
			name:      "Leather Armor",
			required:  map[string]int{"leather": 5, "thread": 3},
			available: map[string]int{"leather": 8, "thread": 6},
		},
	}

	fmt.Println("Tracking recipes:")
	for _, r := range recipes {
		info := &qol.RecipeTrackingInfo{
			RecipeID:      r.id,
			RecipeName:    r.name,
			RequiredMats:  r.required,
			AvailableMats: r.available,
		}

		tracker.TrackRecipe(1, info)

		status := "CAN CRAFT"
		if !info.CanCraft {
			status = "MISSING MATERIALS"
		}

		fmt.Printf("  %s [%s]\n", r.name, status)

		if verbose {
			fmt.Printf("    Required: %v\n", info.RequiredMats)
			fmt.Printf("    Available: %v\n", info.AvailableMats)
			if len(info.MissingMats) > 0 {
				fmt.Printf("    Missing: %v\n", info.MissingMats)
			}
		}

		if info.CanCraft {
			fmt.Printf("    Can craft: %d\n", info.MaxCraftable)
		} else {
			fmt.Printf("    Missing:")
			for mat, qty := range info.MissingMats {
				fmt.Printf(" %s x%d", mat, qty)
			}
			fmt.Println()
		}
	}
	fmt.Println()

	fmt.Println("Updating materials for Health Potion...")
	tracker.UpdateMaterialAvailability(1, "health_potion", map[string]int{"herb": 4, "water": 10})

	tracked := tracker.GetTrackedRecipes(1)
	for _, info := range tracked {
		if info.RecipeID == "health_potion" {
			fmt.Printf("  %s is now %s\n", info.RecipeName,
				map[bool]string{true: "craftable", false: "not craftable"}[info.CanCraft])
			if info.CanCraft {
				fmt.Printf("  Can craft: %d\n", info.MaxCraftable)
			}
		}
	}
	fmt.Println()
}
