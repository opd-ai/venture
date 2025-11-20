package economy

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNewGuildBankManager(t *testing.T) {
	manager := NewGuildBankManager()
	if manager == nil {
		t.Fatal("expected manager, got nil")
	}
	if manager.vaults == nil {
		t.Error("expected vaults map to be initialized")
	}
}

func TestCreateVault(t *testing.T) {
	tests := []struct {
		name         string
		guildID      string
		interestRate float64
		wantErr      bool
	}{
		{"valid vault", "guild1", 0.005, false},
		{"min interest rate", "guild2", 0.001, false},
		{"max interest rate", "guild3", 0.01, false},
		{"interest too low", "guild4", 0.0005, true},
		{"interest too high", "guild5", 0.02, true},
		{"duplicate guild", "guild1", 0.005, true},
	}

	manager := NewGuildBankManager()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CreateVault(tt.guildID, tt.interestRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVault() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				vault, err := manager.GetVault(tt.guildID)
				if err != nil {
					t.Errorf("GetVault() error = %v", err)
				}
				if vault.GuildID != tt.guildID {
					t.Errorf("vault.GuildID = %v, want %v", vault.GuildID, tt.guildID)
				}
				if vault.InterestRate != tt.interestRate {
					t.Errorf("vault.InterestRate = %v, want %v", vault.InterestRate, tt.interestRate)
				}
				if vault.GoldBalance != 0 {
					t.Errorf("vault.GoldBalance = %v, want 0", vault.GoldBalance)
				}
			}
		})
	}
}

func TestGetVault(t *testing.T) {
	manager := NewGuildBankManager()

	// Test non-existent vault
	_, err := manager.GetVault("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent vault")
	}

	// Create and retrieve vault
	if err := manager.CreateVault("guild1", 0.005); err != nil {
		t.Fatalf("CreateVault() error = %v", err)
	}

	vault, err := manager.GetVault("guild1")
	if err != nil {
		t.Fatalf("GetVault() error = %v", err)
	}

	// Verify it's a copy by modifying returned vault
	vault.GoldBalance = 1000
	vault2, _ := manager.GetVault("guild1")
	if vault2.GoldBalance != 0 {
		t.Error("GetVault() should return a copy, not original")
	}
}

func TestDepositGold(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	tests := []struct {
		name       string
		guildID    string
		memberID   string
		memberName string
		amount     int
		wantErr    bool
	}{
		{"valid deposit", "guild1", "member1", "Alice", 1000, false},
		{"another deposit", "guild1", "member2", "Bob", 500, false},
		{"zero amount", "guild1", "member1", "Alice", 0, true},
		{"negative amount", "guild1", "member1", "Alice", -100, true},
		{"invalid guild", "guild999", "member1", "Alice", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.DepositGold(tt.guildID, tt.memberID, tt.memberName, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("DepositGold() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify total balance
	vault, _ := manager.GetVault("guild1")
	expectedBalance := 1500 // 1000 + 500
	if vault.GoldBalance != expectedBalance {
		t.Errorf("vault.GoldBalance = %v, want %v", vault.GoldBalance, expectedBalance)
	}

	// Verify audit log
	if len(vault.AuditLog) != 2 {
		t.Errorf("len(audit log) = %v, want 2", len(vault.AuditLog))
	}
}

func TestWithdrawGold(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositGold("guild1", "member1", "Alice", 1000)

	tests := []struct {
		name       string
		guildID    string
		memberID   string
		memberName string
		rankID     string
		amount     int
		wantErr    bool
	}{
		{"valid withdrawal", "guild1", "member1", "Alice", "member", 100, false},
		{"another withdrawal", "guild1", "member2", "Bob", "member", 200, false},
		{"zero amount", "guild1", "member1", "Alice", "member", 0, true},
		{"negative amount", "guild1", "member1", "Alice", "member", -50, true},
		{"insufficient funds", "guild1", "member1", "Alice", "member", 10000, true},
		{"invalid guild", "guild999", "member1", "Alice", "member", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.WithdrawGold(tt.guildID, tt.memberID, tt.memberName, tt.rankID, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawGold() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify balance after withdrawals
	vault, _ := manager.GetVault("guild1")
	expectedBalance := 700 // 1000 - 100 - 200
	if vault.GoldBalance != expectedBalance {
		t.Errorf("vault.GoldBalance = %v, want %v", vault.GoldBalance, expectedBalance)
	}
}

func TestWithdrawalLimits(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositGold("guild1", "member1", "Alice", 10000)
	manager.SetWithdrawalLimit("guild1", "recruit", 500)

	// First withdrawal within limit
	err := manager.WithdrawGold("guild1", "member1", "Alice", "recruit", 300)
	if err != nil {
		t.Errorf("first withdrawal error = %v", err)
	}

	// Second withdrawal within daily limit
	err = manager.WithdrawGold("guild1", "member1", "Alice", "recruit", 100)
	if err != nil {
		t.Errorf("second withdrawal error = %v", err)
	}

	// Third withdrawal exceeds daily limit
	err = manager.WithdrawGold("guild1", "member1", "Alice", "recruit", 200)
	if err == nil {
		t.Error("expected error for exceeding daily limit")
	}

	// Verify balance (300 + 100 withdrawn)
	vault, _ := manager.GetVault("guild1")
	expectedBalance := 9600
	if vault.GoldBalance != expectedBalance {
		t.Errorf("vault.GoldBalance = %v, want %v", vault.GoldBalance, expectedBalance)
	}
}

func TestDepositItem(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	tests := []struct {
		name         string
		guildID      string
		memberID     string
		memberName   string
		itemID       string
		itemName     string
		itemType     string
		quantity     int
		maxStackSize int
		wantErr      bool
	}{
		{"valid deposit", "guild1", "member1", "Alice", "sword1", "Iron Sword", "weapon", 5, 10, false},
		{"add to existing", "guild1", "member2", "Bob", "sword1", "Iron Sword", "weapon", 3, 10, false},
		{"new item", "guild1", "member1", "Alice", "potion1", "Health Potion", "consumable", 20, 99, false},
		{"zero quantity", "guild1", "member1", "Alice", "item1", "Test", "misc", 0, 10, true},
		{"negative quantity", "guild1", "member1", "Alice", "item1", "Test", "misc", -5, 10, true},
		{"invalid guild", "guild999", "member1", "Alice", "item1", "Test", "misc", 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.DepositItem(tt.guildID, tt.memberID, tt.memberName, tt.itemID, tt.itemName, tt.itemType, tt.quantity, tt.maxStackSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("DepositItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify item quantities
	vault, _ := manager.GetVault("guild1")
	if vault.Items["sword1"].Quantity != 8 { // 5 + 3
		t.Errorf("sword1 quantity = %v, want 8", vault.Items["sword1"].Quantity)
	}
	if vault.Items["potion1"].Quantity != 20 {
		t.Errorf("potion1 quantity = %v, want 20", vault.Items["potion1"].Quantity)
	}
}

func TestWithdrawItem(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositItem("guild1", "member1", "Alice", "sword1", "Iron Sword", "weapon", 10, 10)

	tests := []struct {
		name       string
		guildID    string
		memberID   string
		memberName string
		itemID     string
		quantity   int
		wantErr    bool
	}{
		{"valid withdrawal", "guild1", "member1", "Alice", "sword1", 3, false},
		{"another withdrawal", "guild1", "member2", "Bob", "sword1", 2, false},
		{"zero quantity", "guild1", "member1", "Alice", "sword1", 0, true},
		{"negative quantity", "guild1", "member1", "Alice", "sword1", -1, true},
		{"insufficient quantity", "guild1", "member1", "Alice", "sword1", 100, true},
		{"non-existent item", "guild1", "member1", "Alice", "item999", 1, true},
		{"invalid guild", "guild999", "member1", "Alice", "sword1", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.WithdrawItem(tt.guildID, tt.memberID, tt.memberName, tt.itemID, tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify remaining quantity
	vault, _ := manager.GetVault("guild1")
	expectedQuantity := 5 // 10 - 3 - 2
	if vault.Items["sword1"].Quantity != expectedQuantity {
		t.Errorf("sword1 quantity = %v, want %v", vault.Items["sword1"].Quantity, expectedQuantity)
	}
}

func TestWithdrawItemRemoval(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositItem("guild1", "member1", "Alice", "sword1", "Iron Sword", "weapon", 5, 10)

	// Withdraw all items
	err := manager.WithdrawItem("guild1", "member1", "Alice", "sword1", 5)
	if err != nil {
		t.Fatalf("WithdrawItem() error = %v", err)
	}

	// Verify item is removed from vault
	vault, _ := manager.GetVault("guild1")
	if _, exists := vault.Items["sword1"]; exists {
		t.Error("expected item to be removed when quantity reaches zero")
	}
}

func TestCalculateInterest(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.01) // 1% daily
	manager.DepositGold("guild1", "member1", "Alice", 10000)

	// Set last interest time to 2 days ago
	manager.SetLastInterestTime("guild1", time.Now().Add(-48*time.Hour))

	// Calculate interest
	err := manager.CalculateInterest("guild1")
	if err != nil {
		t.Fatalf("CalculateInterest() error = %v", err)
	}

	// Verify balance increased (compound interest for 2 days)
	vault, _ := manager.GetVault("guild1")
	// Day 1: 10000 * 1.01 = 10100
	// Day 2: 10100 * 1.01 = 10201
	expectedBalance := 10201
	if vault.GoldBalance != expectedBalance {
		t.Errorf("vault.GoldBalance = %v, want %v", vault.GoldBalance, expectedBalance)
	}

	// Verify audit log has interest entries
	foundInterest := false
	for _, entry := range vault.AuditLog {
		if entry.ActionType == AuditInterest {
			foundInterest = true
			break
		}
	}
	if !foundInterest {
		t.Error("expected interest entries in audit log")
	}
}

func TestCalculateInterestNoOp(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.01)
	manager.DepositGold("guild1", "member1", "Alice", 1000)

	// Calculate interest immediately (should be no-op)
	err := manager.CalculateInterest("guild1")
	if err != nil {
		t.Fatalf("CalculateInterest() error = %v", err)
	}

	// Balance should remain unchanged
	vault, _ := manager.GetVault("guild1")
	if vault.GoldBalance != 1000 {
		t.Errorf("vault.GoldBalance = %v, want 1000", vault.GoldBalance)
	}
}

func TestSetWithdrawalLimit(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	tests := []struct {
		name    string
		guildID string
		rankID  string
		limit   int
		wantErr bool
	}{
		{"valid limit", "guild1", "recruit", 500, false},
		{"zero limit", "guild1", "member", 0, false},
		{"high limit", "guild1", "officer", 10000, false},
		{"negative limit", "guild1", "leader", -100, true},
		{"invalid guild", "guild999", "recruit", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SetWithdrawalLimit(tt.guildID, tt.rankID, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetWithdrawalLimit() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				vault, _ := manager.GetVault(tt.guildID)
				if vault.WithdrawalLimits[tt.rankID] != tt.limit {
					t.Errorf("limit = %v, want %v", vault.WithdrawalLimits[tt.rankID], tt.limit)
				}
			}
		})
	}
}

func TestGetAuditLog(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Generate multiple transactions
	for i := 0; i < 10; i++ {
		manager.DepositGold("guild1", "member1", "Alice", 100)
	}

	tests := []struct {
		name      string
		guildID   string
		limit     int
		wantCount int
		wantErr   bool
	}{
		{"get all", "guild1", 100, 10, false},
		{"get last 5", "guild1", 5, 5, false},
		{"get more than available", "guild1", 20, 10, false},
		{"get with zero limit", "guild1", 0, 10, false},
		{"invalid guild", "guild999", 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := manager.GetAuditLog(tt.guildID, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuditLog() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if len(entries) != tt.wantCount {
					t.Errorf("len(entries) = %v, want %v", len(entries), tt.wantCount)
				}
			}
		})
	}
}

func TestAuditLogEviction(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Set low max audit entries for testing
	manager.SetMaxAuditEntries("guild1", 10)

	// Generate more transactions than max
	for i := 0; i < 20; i++ {
		manager.DepositGold("guild1", "member1", "Alice", 100)
	}

	// Verify audit log was evicted
	vault, _ := manager.GetVault("guild1")
	if len(vault.AuditLog) > 10 {
		t.Errorf("len(audit log) = %v, want <= 10", len(vault.AuditLog))
	}
}

func TestUpdateSyncTime(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Get initial sync time
	vault1, _ := manager.GetVault("guild1")
	initialTime := vault1.LastSyncTime

	time.Sleep(10 * time.Millisecond)

	// Update sync time
	err := manager.UpdateSyncTime("guild1")
	if err != nil {
		t.Fatalf("UpdateSyncTime() error = %v", err)
	}

	// Verify sync time was updated
	vault2, _ := manager.GetVault("guild1")
	if !vault2.LastSyncTime.After(initialTime) {
		t.Error("expected sync time to be updated")
	}

	// Test invalid guild
	err = manager.UpdateSyncTime("guild999")
	if err == nil {
		t.Error("expected error for invalid guild")
	}
}

func TestSaveLoad(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositGold("guild1", "member1", "Alice", 1000)
	manager.DepositItem("guild1", "member1", "Alice", "sword1", "Iron Sword", "weapon", 5, 10)
	manager.SetWithdrawalLimit("guild1", "recruit", 500)

	// Save to file
	filename := "test_guild_bank.json.gz"
	defer os.Remove(filename)

	err := manager.Save(filename)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load into new manager
	manager2 := NewGuildBankManager()
	err = manager2.Load(filename)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded data
	vault, err := manager2.GetVault("guild1")
	if err != nil {
		t.Fatalf("GetVault() error = %v", err)
	}

	if vault.GoldBalance != 1000 {
		t.Errorf("GoldBalance = %v, want 1000", vault.GoldBalance)
	}

	if vault.Items["sword1"].Quantity != 5 {
		t.Errorf("sword1 quantity = %v, want 5", vault.Items["sword1"].Quantity)
	}

	if vault.WithdrawalLimits["recruit"] != 500 {
		t.Errorf("recruit limit = %v, want 500", vault.WithdrawalLimits["recruit"])
	}

	if len(vault.AuditLog) == 0 {
		t.Error("expected audit log to be loaded")
	}
}

func TestVaultCapacity(t *testing.T) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Fill vault to capacity (5000 items)
	for i := 0; i < 5000; i++ {
		itemID := fmt.Sprintf("item%d", i)
		err := manager.DepositItem("guild1", "member1", "Alice", itemID, "Test Item", "misc", 1, 10)
		if err != nil {
			t.Fatalf("failed to deposit item %d: %v", i, err)
		}
	}

	// Try to add one more unique item (should fail)
	err := manager.DepositItem("guild1", "member1", "Alice", "item5000", "Overflow Item", "misc", 1, 10)
	if err == nil {
		t.Error("expected error when vault capacity exceeded")
	}

	// Adding to existing item should still work
	err = manager.DepositItem("guild1", "member1", "Alice", "item0", "Test Item", "misc", 5, 10)
	if err != nil {
		t.Errorf("expected to be able to add to existing item: %v", err)
	}
}

// Benchmarks

func BenchmarkCreateVault(b *testing.B) {
	manager := NewGuildBankManager()
	for i := 0; i < b.N; i++ {
		guildID := fmt.Sprintf("guild%d", i)
		manager.CreateVault(guildID, 0.005)
	}
}

func BenchmarkDepositGold(b *testing.B) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.DepositGold("guild1", "member1", "Alice", 100)
	}
}

func BenchmarkWithdrawGold(b *testing.B) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositGold("guild1", "member1", "Alice", 1000000) // Large initial balance

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.WithdrawGold("guild1", "member1", "Alice", "member", 10)
	}
}

func BenchmarkDepositItem(b *testing.B) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.DepositItem("guild1", "member1", "Alice", "sword1", "Iron Sword", "weapon", 1, 10)
	}
}

func BenchmarkWithdrawItem(b *testing.B) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)
	manager.DepositItem("guild1", "member1", "Alice", "sword1", "Iron Sword", "weapon", 100000, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.WithdrawItem("guild1", "member1", "Alice", "sword1", 1)
	}
}

func BenchmarkCalculateInterest(b *testing.B) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.01)
	manager.DepositGold("guild1", "member1", "Alice", 10000)

	// Set last interest time to 1 day ago for each run
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		manager.SetLastInterestTime("guild1", time.Now().Add(-25*time.Hour))
		b.StartTimer()

		manager.CalculateInterest("guild1")
	}
}

func BenchmarkGetAuditLog(b *testing.B) {
	manager := NewGuildBankManager()
	manager.CreateVault("guild1", 0.005)

	// Generate audit log entries
	for i := 0; i < 100; i++ {
		manager.DepositGold("guild1", "member1", "Alice", 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetAuditLog("guild1", 50)
	}
}

func BenchmarkSave(b *testing.B) {
	manager := NewGuildBankManager()
	for i := 0; i < 10; i++ {
		guildID := fmt.Sprintf("guild%d", i)
		manager.CreateVault(guildID, 0.005)
		manager.DepositGold(guildID, "member1", "Alice", 10000)
		for j := 0; j < 10; j++ {
			itemID := fmt.Sprintf("item%d", j)
			manager.DepositItem(guildID, "member1", "Alice", itemID, "Test Item", "misc", 10, 99)
		}
	}

	filename := "bench_guild_bank.json.gz"
	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Save(filename)
	}
}

func BenchmarkLoad(b *testing.B) {
	manager := NewGuildBankManager()
	for i := 0; i < 10; i++ {
		guildID := fmt.Sprintf("guild%d", i)
		manager.CreateVault(guildID, 0.005)
		manager.DepositGold(guildID, "member1", "Alice", 10000)
	}

	filename := "bench_guild_bank.json.gz"
	defer os.Remove(filename)
	manager.Save(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager2 := NewGuildBankManager()
		manager2.Load(filename)
	}
}
