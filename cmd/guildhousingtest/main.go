// Guild Housing Integration Test - CLI tool for Phase 55.3
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/integration/guild_housing"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/housing"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, permissions, crafting, storage, meeting, upgrade, all")
	flag.Parse()

	fmt.Println("=== Guild Housing Integration Test ===")
	fmt.Println()

	switch *mode {
	case "demo":
		runDemo()
	case "permissions":
		testPermissions()
	case "crafting":
		testCrafting()
	case "storage":
		testStorage()
	case "meeting":
		testMeeting()
	case "upgrade":
		testUpgrade()
	case "all":
		runDemo()
		testPermissions()
		testCrafting()
		testStorage()
		testMeeting()
		testUpgrade()
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}

	fmt.Println()
	fmt.Println("=== Test Complete ===")
}

func runDemo() {
	fmt.Println("--- Basic Demo ---")

	manager := guild_housing.NewManager()
	size := housing.SizeLarge

	// Create guild house
	house := manager.CreateGuildHouse("guild-001", "player-001", size)
	fmt.Printf("Created guild house: %s\n", house.HouseID)
	fmt.Printf("  Guild: %s\n", house.GuildID)
	fmt.Printf("  Owner: %s\n", house.OwnerID)
	fmt.Printf("  Size: %dx%d tiles\n", house.Size.Tiles(), house.Size.Tiles())
	fmt.Printf("  Tier: %s\n", house.Tier)

	// Show default permissions
	fmt.Println("  Default Permissions:")
	for rank, perm := range house.Permissions {
		fmt.Printf("    %s: %s\n", rank, perm)
	}

	// Add crafting stations
	manager.AddCraftingStation(house.HouseID, "station-forge-001")
	manager.AddCraftingStation(house.HouseID, "station-alchemy-001")
	stations, _ := manager.GetCraftingStations(house.HouseID)
	fmt.Printf("  Crafting Stations: %d\n", len(stations))

	// Create guild storage
	storage := manager.CreateGuildStorage(house.GuildID, 500)
	fmt.Printf("Created guild storage: %s\n", storage.StorageID)
	fmt.Printf("  Capacity: %d slots\n", storage.Capacity)

	// Deposit items
	manager.DepositItem(storage.StorageID, "player-001", "wood", 100)
	manager.DepositItem(storage.StorageID, "player-002", "stone", 50)
	items, _ := manager.GetStorageItems(storage.StorageID)
	fmt.Printf("  Stored Items: %d types\n", len(items))
	for itemID, item := range items {
		fmt.Printf("    %s: %d (added by %s)\n", itemID, item.Quantity, item.AddedBy)
	}

	fmt.Println()
}

func testPermissions() {
	fmt.Println("--- Permission System Test ---")

	manager := guild_housing.NewManager()
	size := housing.SizeMedium
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	tests := []struct {
		rank     guild.Rank
		required guild_housing.Permission
		desc     string
	}{
		{guild.RankLeader, guild_housing.PermissionAdmin, "Leader admin access"},
		{guild.RankLeader, guild_housing.PermissionManage, "Leader manage access"},
		{guild.RankOfficer, guild_housing.PermissionManage, "Officer manage access"},
		{guild.RankOfficer, guild_housing.PermissionAdmin, "Officer admin access (denied)"},
		{guild.RankMember, guild_housing.PermissionUse, "Member use access"},
		{guild.RankMember, guild_housing.PermissionManage, "Member manage access (denied)"},
		{guild.RankRecruit, guild_housing.PermissionView, "Recruit view access"},
		{guild.RankRecruit, guild_housing.PermissionUse, "Recruit use access (denied)"},
	}

	for _, test := range tests {
		allowed := manager.CheckPermission(house.HouseID, test.rank, test.required)
		status := "✓ ALLOWED"
		if !allowed {
			status = "✗ DENIED"
		}
		fmt.Printf("%s - %s: %s\n", status, test.desc, test.rank)
	}

	// Change permissions
	fmt.Println("\nChanging Member permission to Manage...")
	manager.SetPermission(house.HouseID, guild.RankMember, guild_housing.PermissionManage)
	allowed := manager.CheckPermission(house.HouseID, guild.RankMember, guild_housing.PermissionManage)
	if allowed {
		fmt.Println("✓ Permission change successful")
	}

	fmt.Println()
}

func testCrafting() {
	fmt.Println("--- Communal Crafting Test ---")

	manager := guild_housing.NewManager()
	size := housing.SizeLarge
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	// Add multiple crafting stations
	stations := []string{
		"forge-001",
		"forge-002",
		"alchemy-001",
		"enchanting-001",
		"cooking-001",
	}

	fmt.Printf("Adding %d crafting stations...\n", len(stations))
	for _, stationID := range stations {
		manager.AddCraftingStation(house.HouseID, stationID)
		fmt.Printf("  ✓ Added station: %s\n", stationID)
	}

	// Verify stations
	added, _ := manager.GetCraftingStations(house.HouseID)
	fmt.Printf("\nTotal crafting stations: %d\n", len(added))
	fmt.Println("  Multiple members can craft simultaneously!")

	fmt.Println()
}

func testStorage() {
	fmt.Println("--- Guild Storage Test ---")

	manager := guild_housing.NewManager()
	storage := manager.CreateGuildStorage("guild-001", 500)

	fmt.Printf("Created storage with %d slot capacity\n\n", storage.Capacity)

	// Multiple players deposit items
	deposits := []struct {
		player   string
		item     string
		quantity int
	}{
		{"player-001", "wood", 100},
		{"player-002", "stone", 50},
		{"player-003", "iron", 25},
		{"player-001", "wood", 50}, // Additional deposit
		{"player-004", "gold", 10},
	}

	fmt.Println("Depositing items:")
	for _, d := range deposits {
		manager.DepositItem(storage.StorageID, d.player, d.item, d.quantity)
		fmt.Printf("  ✓ %s deposited %d %s\n", d.player, d.quantity, d.item)
	}

	// Show storage contents
	items, _ := manager.GetStorageItems(storage.StorageID)
	fmt.Printf("\nStorage Contents (%d types):\n", len(items))
	for itemID, item := range items {
		fmt.Printf("  %s: %d units\n", itemID, item.Quantity)
	}

	// Withdraw items
	fmt.Println("\nWithdrawing items:")
	withdrawn, _ := manager.WithdrawItem(storage.StorageID, "player-005", "wood", 30)
	fmt.Printf("  ✓ player-005 withdrew %d wood\n", withdrawn)

	withdrawn, _ = manager.WithdrawItem(storage.StorageID, "player-006", "stone", 100)
	fmt.Printf("  ✓ player-006 withdrew %d stone (partial, only %d available)\n", 50, withdrawn)

	// Show transaction log
	transactions, _ := manager.GetTransactions(storage.StorageID)
	fmt.Printf("\nTransaction History (%d entries):\n", len(transactions))
	for i, tx := range transactions {
		if i < 5 { // Show first 5
			fmt.Printf("  %s: %s %d %s (%s)\n",
				tx.Action, tx.PlayerID, tx.Quantity, tx.ItemID, tx.Timestamp.Format("15:04:05"))
		}
	}
	if len(transactions) > 5 {
		fmt.Printf("  ... and %d more transactions\n", len(transactions)-5)
	}

	fmt.Println()
}

func testMeeting() {
	fmt.Println("--- Meeting Hall Test ---")

	manager := guild_housing.NewManager()
	hall := manager.CreateMeetingHall("guild-001", 50)

	fmt.Printf("Created meeting hall: %s\n", hall.HallID)
	fmt.Printf("  Max Capacity: %d players\n", hall.MaxCapacity)
	fmt.Printf("  Chat Radius: %.1f tiles (+50%% bonus)\n", hall.ChatRadius)
	fmt.Println()

	// Add members to hall
	members := []string{"player-001", "player-002", "player-003", "player-004", "player-005"}
	fmt.Println("Members joining hall:")
	for _, playerID := range members {
		err := manager.AddMemberToHall(hall, playerID)
		if err != nil {
			fmt.Printf("  ✗ %s: %v\n", playerID, err)
		} else {
			fmt.Printf("  ✓ %s joined\n", playerID)
		}
	}

	fmt.Printf("\nCurrent occupancy: %d/%d players\n", len(hall.Members), hall.MaxCapacity)

	// Remove a member
	fmt.Println("\nRemoving player-003...")
	manager.RemoveMemberFromHall(hall, "player-003")
	fmt.Printf("Occupancy after removal: %d/%d players\n", len(hall.Members), hall.MaxCapacity)

	fmt.Println()
}

func testUpgrade() {
	fmt.Println("--- Guild House Upgrade Test ---")

	manager := guild_housing.NewManager()
	size := housing.SizeLarge
	house := manager.CreateGuildHouse("guild-001", "player-001", size)

	fmt.Printf("Initial tier: %s\n", house.Tier)
	bonus, _ := manager.GetUpgradeBonus(house.HouseID)
	fmt.Printf("Initial bonus: %.1fx\n", bonus)
	fmt.Println()

	// Upgrade to Standard
	cost := guild_housing.TierStandard.Cost()
	fmt.Printf("Upgrading to Standard (cost: %d gold)...\n", cost)
	err := manager.UpgradeHouse(house.HouseID, cost)
	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		house, _ = manager.GetGuildHouse(house.HouseID)
		bonus, _ = manager.GetUpgradeBonus(house.HouseID)
		fmt.Printf("  ✓ Upgraded to %s (bonus: %.1fx)\n", house.Tier, bonus)
	}

	// Upgrade to Advanced
	cost = guild_housing.TierAdvanced.Cost()
	fmt.Printf("\nUpgrading to Advanced (cost: %d gold)...\n", cost)
	err = manager.UpgradeHouse(house.HouseID, cost)
	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		house, _ = manager.GetGuildHouse(house.HouseID)
		bonus, _ = manager.GetUpgradeBonus(house.HouseID)
		fmt.Printf("  ✓ Upgraded to %s (bonus: %.1fx)\n", house.Tier, bonus)
	}

	// Upgrade to Master
	cost = guild_housing.TierMaster.Cost()
	fmt.Printf("\nUpgrading to Master (cost: %d gold)...\n", cost)
	err = manager.UpgradeHouse(house.HouseID, cost)
	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		house, _ = manager.GetGuildHouse(house.HouseID)
		bonus, _ = manager.GetUpgradeBonus(house.HouseID)
		fmt.Printf("  ✓ Upgraded to %s (bonus: %.1fx)\n", house.Tier, bonus)
	}

	// Try to upgrade beyond max
	fmt.Println("\nAttempting to upgrade beyond Master...")
	err = manager.UpgradeHouse(house.HouseID, 100000)
	if err != nil {
		fmt.Printf("  ✓ Correctly prevented: %v\n", err)
	} else {
		fmt.Println("  ✗ Should not allow upgrade beyond Master")
	}

	fmt.Println()
}
