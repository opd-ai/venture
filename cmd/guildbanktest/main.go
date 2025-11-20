// guildbanktest is a CLI tool for testing and demonstrating the guild bank system.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/world/economy"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, vault, gold, items, interest, audit, all")
	flag.Parse()

	switch *mode {
	case "demo":
		demoMode()
	case "vault":
		vaultMode()
	case "gold":
		goldMode()
	case "items":
		itemsMode()
	case "interest":
		interestMode()
	case "audit":
		auditMode()
	case "all":
		demoMode()
		vaultMode()
		goldMode()
		itemsMode()
		interestMode()
		auditMode()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: demo, vault, gold, items, interest, audit, all")
	}
}

func demoMode() {
	fmt.Println("=== Guild Bank Demo ===")
	fmt.Println()

	manager := economy.NewGuildBankManager()

	// Create vaults for 3 guilds
	guilds := []struct {
		id           string
		name         string
		interestRate float64
	}{
		{"guild1", "Warriors of Light", 0.005},
		{"guild2", "Shadow Traders", 0.008},
		{"guild3", "Dragon Slayers", 0.01},
	}

	for _, g := range guilds {
		err := manager.CreateVault(g.id, g.interestRate)
		if err != nil {
			fmt.Printf("Error creating vault for %s: %v\n", g.name, err)
			continue
		}
		fmt.Printf("Created vault for %s (interest: %.1f%% daily)\n", g.name, g.interestRate*100)
	}

	fmt.Println()
	fmt.Println("Guild vaults created successfully!")
	fmt.Println()
}

func vaultMode() {
	fmt.Println("=== Vault Management ===")
	fmt.Println()

	manager := economy.NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Set withdrawal limits for different ranks
	ranks := []struct {
		id    string
		name  string
		limit int
	}{
		{"recruit", "Recruit", 500},
		{"member", "Member", 2000},
		{"officer", "Officer", 5000},
		{"leader", "Leader", 0}, // 0 = no limit
	}

	fmt.Println("Setting withdrawal limits by rank:")
	for _, r := range ranks {
		manager.SetWithdrawalLimit("guild1", r.id, r.limit)
		if r.limit == 0 {
			fmt.Printf("  %s: Unlimited\n", r.name)
		} else {
			fmt.Printf("  %s: %d gold/day\n", r.name, r.limit)
		}
	}

	vault, _ := manager.GetVault("guild1")
	fmt.Printf("\nVault created at: %s\n", vault.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Max item slots: 5000\n")
	fmt.Printf("Audit log retention: %d entries (~30 days)\n", vault.MaxAuditEntries)
	fmt.Println()
}

func goldMode() {
	fmt.Println("=== Gold Transactions ===")
	fmt.Println()

	manager := economy.NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.SetWithdrawalLimit("guild1", "member", 1000)

	// Deposit transactions
	deposits := []struct {
		member string
		amount int
	}{
		{"Alice", 5000},
		{"Bob", 3000},
		{"Charlie", 2000},
	}

	fmt.Println("Deposits:")
	totalDeposited := 0
	for _, d := range deposits {
		err := manager.DepositGold("guild1", d.member, d.member, d.amount)
		if err != nil {
			fmt.Printf("  %s: Error - %v\n", d.member, err)
		} else {
			fmt.Printf("  %s deposited %d gold\n", d.member, d.amount)
			totalDeposited += d.amount
		}
	}

	vault, _ := manager.GetVault("guild1")
	fmt.Printf("\nTotal vault balance: %d gold\n", vault.GoldBalance)

	// Withdrawal transactions
	fmt.Println("\nWithdrawals:")
	manager.WithdrawGold("guild1", "Alice", "Alice", "member", 500)
	fmt.Println("  Alice withdrew 500 gold (within daily limit)")

	manager.WithdrawGold("guild1", "Bob", "Bob", "member", 800)
	fmt.Println("  Bob withdrew 800 gold (within daily limit)")

	err := manager.WithdrawGold("guild1", "Bob", "Bob", "member", 500)
	if err != nil {
		fmt.Printf("  Bob tried to withdraw 500 gold: %v\n", err)
	}

	vault, _ = manager.GetVault("guild1")
	fmt.Printf("\nFinal vault balance: %d gold\n", vault.GoldBalance)
	fmt.Println()
}

func itemsMode() {
	fmt.Println("=== Item Storage ===")
	fmt.Println()

	manager := economy.NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Deposit items
	items := []struct {
		member   string
		id       string
		name     string
		itemType string
		quantity int
		stack    int
	}{
		{"Alice", "sword1", "Iron Sword", "weapon", 5, 10},
		{"Bob", "potion1", "Health Potion", "consumable", 50, 99},
		{"Charlie", "ore1", "Iron Ore", "material", 100, 999},
		{"Alice", "sword1", "Iron Sword", "weapon", 3, 10}, // Add more to existing
	}

	fmt.Println("Item deposits:")
	for _, item := range items {
		err := manager.DepositItem("guild1", item.member, item.member, item.id, item.name, item.itemType, item.quantity, item.stack)
		if err != nil {
			fmt.Printf("  %s: Error depositing %s - %v\n", item.member, item.name, err)
		} else {
			fmt.Printf("  %s deposited %d × %s\n", item.member, item.quantity, item.name)
		}
	}

	vault, _ := manager.GetVault("guild1")
	fmt.Printf("\nTotal unique items in vault: %d\n", len(vault.Items))

	fmt.Println("\nVault inventory:")
	for _, item := range vault.Items {
		fmt.Printf("  %d × %s (%s)\n", item.Quantity, item.ItemName, item.ItemType)
	}

	// Withdraw items
	fmt.Println("\nItem withdrawals:")
	manager.WithdrawItem("guild1", "Alice", "Alice", "sword1", 2)
	fmt.Println("  Alice withdrew 2 × Iron Sword")

	manager.WithdrawItem("guild1", "Bob", "Bob", "potion1", 20)
	fmt.Println("  Bob withdrew 20 × Health Potion")

	vault, _ = manager.GetVault("guild1")
	fmt.Println("\nUpdated inventory:")
	for _, item := range vault.Items {
		fmt.Printf("  %d × %s\n", item.Quantity, item.ItemName)
	}
	fmt.Println()
}

func interestMode() {
	fmt.Println("=== Interest Calculation ===")
	fmt.Println()

	manager := economy.NewGuildBankManager()

	// Create vaults with different interest rates
	vaults := []struct {
		id           string
		name         string
		interestRate float64
		balance      int
	}{
		{"guild1", "Low Interest Guild", 0.001, 10000},    // 0.1% daily
		{"guild2", "Medium Interest Guild", 0.005, 10000}, // 0.5% daily
		{"guild3", "High Interest Guild", 0.01, 10000},    // 1.0% daily
	}

	for _, v := range vaults {
		manager.CreateVault(v.id, v.interestRate)
		manager.DepositGold(v.id, "system", "System", v.balance)
		fmt.Printf("Created %s: %d gold at %.1f%% daily interest\n", v.name, v.balance, v.interestRate*100)
	}

	// Simulate 7 days of interest
	fmt.Println("\nSimulating 7 days of compound interest:")
	fmt.Println()

	for day := 1; day <= 7; day++ {
		fmt.Printf("Day %d:\n", day)
		for _, v := range vaults {
			// Set last interest time to 1 day ago
			manager.SetLastInterestTime(v.id, time.Now().Add(-25*time.Hour))

			// Calculate interest
			manager.CalculateInterest(v.id)
			vault, _ := manager.GetVault(v.id)
			fmt.Printf("  %s: %d gold\n", v.name, vault.GoldBalance)
		}
		fmt.Println()
	}

	// Show final balances and gains
	fmt.Println("Final balances after 7 days:")
	for _, v := range vaults {
		vault, _ := manager.GetVault(v.id)
		gain := vault.GoldBalance - v.balance
		gainPercent := (float64(gain) / float64(v.balance)) * 100
		fmt.Printf("  %s: %d gold (+%d, +%.2f%%)\n", v.name, vault.GoldBalance, gain, gainPercent)
	}
	fmt.Println()
}

func auditMode() {
	fmt.Println("=== Audit Log ===")
	fmt.Println()

	manager := economy.NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Generate various transactions
	fmt.Println("Generating transactions:")
	manager.DepositGold("guild1", "Alice", "Alice", 5000)
	fmt.Println("  Alice deposited 5000 gold")

	manager.WithdrawGold("guild1", "Bob", "Bob", "member", 1000)
	fmt.Println("  Bob withdrew 1000 gold")

	manager.DepositItem("guild1", "Charlie", "Charlie", "sword1", "Iron Sword", "weapon", 10, 10)
	fmt.Println("  Charlie deposited 10 × Iron Sword")

	manager.WithdrawItem("guild1", "Alice", "Alice", "sword1", 3)
	fmt.Println("  Alice withdrew 3 × Iron Sword")

	manager.DepositGold("guild1", "Bob", "Bob", 2000)
	fmt.Println("  Bob deposited 2000 gold")

	// Retrieve and display audit log
	entries, _ := manager.GetAuditLog("guild1", 10)

	fmt.Printf("\nAudit log (last %d entries):\n", len(entries))
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Printf("%-20s %-10s %-12s %-10s %-12s\n", "Time", "Member", "Action", "Amount", "Balance")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	for _, entry := range entries {
		timestamp := entry.Timestamp.Format("15:04:05")
		action := entry.ActionType.String()

		var details string
		if entry.ItemID != "" {
			details = fmt.Sprintf("%d × %s", entry.Quantity, entry.ItemName)
		} else {
			details = fmt.Sprintf("%d gold", entry.GoldAmount)
		}

		fmt.Printf("%-20s %-10s %-12s %-10s %-12d\n",
			timestamp, entry.MemberName, action, details, entry.BalanceAfter)
	}

	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Println()

	// Show audit log retention
	vault, _ := manager.GetVault("guild1")
	fmt.Printf("Audit log: %d/%d entries (30 day retention)\n", len(vault.AuditLog), vault.MaxAuditEntries)
	fmt.Println()
}
