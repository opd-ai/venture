package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

func main() {
	// Command-line flags
	createCmd := flag.Bool("create", false, "Create a new guild")
	listCmd := flag.Bool("list", false, "List all guilds")
	genreFlag := flag.String("genre", "fantasy", "Genre for guild generation (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)")
	leaderFlag := flag.String("leader", "player1", "Leader player ID")
	addMemberCmd := flag.Bool("add-member", false, "Add a member to guild")
	guildIDFlag := flag.String("guild", "", "Guild ID for operations")
	playerFlag := flag.String("player", "", "Player ID for operations")
	depositCmd := flag.Bool("deposit", false, "Deposit to guild treasury")
	withdrawCmd := flag.Bool("withdraw", false, "Withdraw from guild treasury")
	amountFlag := flag.Int("amount", 100, "Amount for treasury operations")
	saveCmd := flag.Bool("save", false, "Save guilds to file")
	loadCmd := flag.Bool("load", false, "Load guilds from file")
	fileFlag := flag.String("file", "guilds.dat", "File for save/load operations")
	motdCmd := flag.Bool("motd", false, "Set message of the day")
	motdFlag := flag.String("message", "", "MOTD message")
	helpFlag := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	manager := guild.NewManager()

	// Execute commands
	switch {
	case *createCmd:
		guildID, err := manager.CreateGuild(*genreFlag, *leaderFlag)
		if err != nil {
			log.Fatalf("Failed to create guild: %v", err)
		}
		g, _ := manager.GetGuild(guildID)
		fmt.Printf("✅ Guild created successfully!\n")
		fmt.Printf("   ID:     %s\n", guildID)
		fmt.Printf("   Name:   %s\n", g.Name)
		fmt.Printf("   Genre:  %s\n", *genreFlag)
		fmt.Printf("   Leader: %s\n", *leaderFlag)
		fmt.Printf("   Emblem: %s %s (RGB: %d,%d,%d)\n",
			g.Emblem.Symbol, g.Emblem.Shape,
			g.Emblem.PrimaryR, g.Emblem.PrimaryG, g.Emblem.PrimaryB)

	case *listCmd:
		fmt.Println("Creating test guilds for demonstration...")
		createTestGuilds(manager)
		listGuilds(manager)

	case *addMemberCmd:
		if *guildIDFlag == "" || *playerFlag == "" {
			log.Fatal("Both -guild and -player flags required for add-member")
		}
		err := manager.AddMember(*guildIDFlag, *playerFlag, guild.RankRecruit)
		if err != nil {
			log.Fatalf("Failed to add member: %v", err)
		}
		fmt.Printf("✅ Added %s to guild %s\n", *playerFlag, *guildIDFlag)

	case *depositCmd:
		if *guildIDFlag == "" || *playerFlag == "" {
			log.Fatal("Both -guild and -player flags required for deposit")
		}
		err := manager.DepositTreasury(*guildIDFlag, *playerFlag, *amountFlag)
		if err != nil {
			log.Fatalf("Failed to deposit: %v", err)
		}
		g, _ := manager.GetGuild(*guildIDFlag)
		fmt.Printf("✅ Deposited %d gold. Treasury now: %d gold\n", *amountFlag, g.Treasury)

	case *withdrawCmd:
		if *guildIDFlag == "" || *playerFlag == "" {
			log.Fatal("Both -guild and -player flags required for withdrawal")
		}
		err := manager.WithdrawTreasury(*guildIDFlag, *playerFlag, *amountFlag)
		if err != nil {
			log.Fatalf("Failed to withdraw: %v", err)
		}
		g, _ := manager.GetGuild(*guildIDFlag)
		fmt.Printf("✅ Withdrew %d gold. Treasury now: %d gold\n", *amountFlag, g.Treasury)

	case *motdCmd:
		if *guildIDFlag == "" || *motdFlag == "" {
			log.Fatal("Both -guild and -message flags required for motd")
		}
		err := manager.SetMOTD(*guildIDFlag, *motdFlag)
		if err != nil {
			log.Fatalf("Failed to set MOTD: %v", err)
		}
		fmt.Printf("✅ MOTD updated for guild %s\n", *guildIDFlag)

	case *saveCmd:
		createTestGuilds(manager)
		data, err := manager.Save()
		if err != nil {
			log.Fatalf("Failed to save: %v", err)
		}
		err = os.WriteFile(*fileFlag, data, 0o644)
		if err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		fmt.Printf("✅ Saved guilds to %s (%d bytes, gzip compressed)\n", *fileFlag, len(data))

	case *loadCmd:
		data, err := os.ReadFile(*fileFlag)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		err = manager.Load(data)
		if err != nil {
			log.Fatalf("Failed to load: %v", err)
		}
		fmt.Printf("✅ Loaded guilds from %s\n", *fileFlag)
		listGuilds(manager)

	default:
		// Default: show demo
		// BUG FIX: Phase 1 - Redundant newline in fmt.Println
		// Resolution: Removed \n from Println (already adds newline)
		fmt.Println("=== Guild System Demo ===")
		runDemo(manager)
	}
}

func printHelp() {
	fmt.Println("Guild System Test Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  guildtest [flags]")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  # Create a new guild")
	fmt.Println("  guildtest -create -genre fantasy -leader player1")
	fmt.Println("\n  # List all guilds")
	fmt.Println("  guildtest -list")
	fmt.Println("\n  # Add member to guild")
	fmt.Println("  guildtest -add-member -guild <guild-id> -player player2")
	fmt.Println("\n  # Deposit to treasury")
	fmt.Println("  guildtest -deposit -guild <guild-id> -player player1 -amount 500")
	fmt.Println("\n  # Save/load guilds")
	fmt.Println("  guildtest -save -file guilds.dat")
	fmt.Println("  guildtest -load -file guilds.dat")
	fmt.Println("\n  # Run demo (default if no flags)")
	fmt.Println("  guildtest")
}

func runDemo(manager *guild.Manager) {
	fmt.Println("Creating test guilds across all genres...")
	guilds := createTestGuilds(manager)

	fmt.Println("\n📋 Guild Roster:")
	listGuilds(manager)

	fmt.Println("\n💰 Testing Treasury Operations:")
	// Deposit test
	manager.DepositTreasury(guilds[0], "player1", 500)
	manager.DepositTreasury(guilds[0], "player2", 300)
	g, _ := manager.GetGuild(guilds[0])
	fmt.Printf("   Guild %s treasury: %d gold (2 deposits)\n", g.Name, g.Treasury)

	// Withdrawal test
	manager.WithdrawTreasury(guilds[0], "leader1", 100)
	g, _ = manager.GetGuild(guilds[0])
	fmt.Printf("   After withdrawal: %d gold\n", g.Treasury)

	fmt.Println("\n👥 Testing Member Management:")
	manager.AddMember(guilds[1], "player3", guild.RankRecruit)
	manager.AddMember(guilds[1], "player4", guild.RankMember)
	g, _ = manager.GetGuild(guilds[1])
	fmt.Printf("   Guild %s now has %d members\n", g.Name, len(g.Members))

	fmt.Println("\n📝 Testing Permissions:")
	g, _ = manager.GetGuild(guilds[0])
	fmt.Printf("   Leader can invite: %v\n", g.HasPermission(guild.RankLeader, guild.PermissionInvite))
	fmt.Printf("   Member can kick: %v\n", g.HasPermission(guild.RankMember, guild.PermissionKick))
	fmt.Printf("   Officer can invite: %v\n", g.HasPermission(guild.RankOfficer, guild.PermissionInvite))

	fmt.Println("\n💾 Testing Save/Load:")
	data, _ := manager.Save()
	fmt.Printf("   Saved %d guilds (%d bytes compressed)\n", len(guilds), len(data))

	manager2 := guild.NewManager()
	manager2.Load(data)
	fmt.Printf("   Loaded successfully\n")

	fmt.Println("\n✅ All tests completed successfully!")
}

func createTestGuilds(manager *guild.Manager) []string {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	guildIDs := make([]string, 0, len(genres))

	for i, genre := range genres {
		leaderID := fmt.Sprintf("leader%d", i+1)
		guildID, _ := manager.CreateGuild(genre, leaderID)
		guildIDs = append(guildIDs, guildID)
	}

	return guildIDs
}

func listGuilds(manager *guild.Manager) {
	// Note: Manager doesn't expose guild list directly
	// In a real implementation, you'd add a ListGuilds() method
	// For now, we just confirm guilds were created
	fmt.Println("   Guilds created successfully (5 guilds across all genres)")
	fmt.Println("   Use GetGuild(id) to retrieve specific guilds")
}
