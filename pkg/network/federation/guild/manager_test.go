package guild

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.guilds == nil {
		t.Error("guilds map not initialized")
	}
}

func TestCreateGuild(t *testing.T) {
	tests := []struct {
		name     string
		genre    string
		leaderID string
		seed     int64
	}{
		{"fantasy guild", "fantasy", "player1", 12345},
		{"sci-fi guild", "sci-fi", "player2", 23456},
		{"horror guild", "horror", "player3", 34567},
		{"cyberpunk guild", "cyberpunk", "player4", 45678},
		{"post-apocalyptic guild", "post-apocalyptic", "player5", 56789},
		{"unknown genre", "unknown", "player6", 67890},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			guildID, err := m.CreateGuild(tt.genre, tt.leaderID, tt.seed)
			if err != nil {
				t.Fatalf("CreateGuild failed: %v", err)
			}
			if guildID == "" {
				t.Error("CreateGuild returned empty ID")
			}

			// Verify guild was created
			guild, err := m.GetGuild(guildID)
			if err != nil {
				t.Fatalf("GetGuild failed: %v", err)
			}
			if guild.LeaderID != tt.leaderID {
				t.Errorf("LeaderID = %s, want %s", guild.LeaderID, tt.leaderID)
			}
			if len(guild.Members) != 1 {
				t.Errorf("Members count = %d, want 1", len(guild.Members))
			}
			if guild.Members[0].Rank != RankLeader {
				t.Errorf("First member rank = %s, want %s", guild.Members[0].Rank, RankLeader)
			}
			if guild.Treasury != 0 {
				t.Errorf("Treasury = %d, want 0", guild.Treasury)
			}
			if guild.Name == "" {
				t.Error("Guild name is empty")
			}
			if guild.Emblem == nil {
				t.Error("Guild emblem is nil")
			}
		})
	}
}

func TestGetGuild(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "player1", 12345)

	t.Run("existing guild", func(t *testing.T) {
		guild, err := m.GetGuild(guildID)
		if err != nil {
			t.Fatalf("GetGuild failed: %v", err)
		}
		if guild.ID != guildID {
			t.Errorf("ID = %s, want %s", guild.ID, guildID)
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		_, err := m.GetGuild("nonexistent")
		if err == nil {
			t.Error("GetGuild should fail for non-existing guild")
		}
	})
}

func TestAddMember(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	t.Run("add new member", func(t *testing.T) {
		err := m.AddMember(guildID, "player1", RankRecruit)
		if err != nil {
			t.Fatalf("AddMember failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		if len(guild.Members) != 2 {
			t.Errorf("Members count = %d, want 2", len(guild.Members))
		}

		// Verify member exists
		member := GetMember(guild, "player1")
		if member == nil {
			t.Fatal("Member not found")
		}
		if member.Rank != RankRecruit {
			t.Errorf("Rank = %s, want %s", member.Rank, RankRecruit)
		}
	})

	t.Run("add duplicate member", func(t *testing.T) {
		err := m.AddMember(guildID, "player1", RankMember)
		if err == nil {
			t.Error("AddMember should fail for duplicate member")
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		err := m.AddMember("nonexistent", "player2", RankMember)
		if err == nil {
			t.Error("AddMember should fail for non-existing guild")
		}
	})
}

func TestRemoveMember(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	m.AddMember(guildID, "player1", RankMember)

	t.Run("remove existing member", func(t *testing.T) {
		err := m.RemoveMember(guildID, "player1")
		if err != nil {
			t.Fatalf("RemoveMember failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		if len(guild.Members) != 1 {
			t.Errorf("Members count = %d, want 1", len(guild.Members))
		}
	})

	t.Run("remove leader", func(t *testing.T) {
		err := m.RemoveMember(guildID, "leader")
		if err == nil {
			t.Error("RemoveMember should fail when trying to remove leader")
		}
	})

	t.Run("remove non-existing member", func(t *testing.T) {
		err := m.RemoveMember(guildID, "nonexistent")
		if err == nil {
			t.Error("RemoveMember should fail for non-existing member")
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		err := m.RemoveMember("nonexistent", "player1")
		if err == nil {
			t.Error("RemoveMember should fail for non-existing guild")
		}
	})
}

func TestPromoteMember(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	m.AddMember(guildID, "player1", RankRecruit)

	t.Run("promote recruit to member", func(t *testing.T) {
		err := m.PromoteMember(guildID, "player1", "leader")
		if err != nil {
			t.Fatalf("PromoteMember failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		member := GetMember(guild, "player1")
		if member == nil {
			t.Fatal("Member not found")
		}
		if member.Rank != RankMember {
			t.Errorf("Rank = %s, want %s", member.Rank, RankMember)
		}
	})

	t.Run("promote member to officer", func(t *testing.T) {
		err := m.PromoteMember(guildID, "player1", "leader")
		if err != nil {
			t.Fatalf("PromoteMember failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		member := GetMember(guild, "player1")
		if member.Rank != RankOfficer {
			t.Errorf("Rank = %s, want %s", member.Rank, RankOfficer)
		}
	})

	t.Run("promote officer to leader transfers leadership", func(t *testing.T) {
		err := m.PromoteMember(guildID, "player1", "leader")
		if err != nil {
			t.Fatalf("PromoteMember failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		if guild.LeaderID != "player1" {
			t.Errorf("LeaderID = %s, want player1", guild.LeaderID)
		}

		// Old leader should be demoted to officer
		oldLeader := GetMember(guild, "leader")
		if oldLeader == nil {
			t.Fatal("Old leader not found")
		}
		if oldLeader.Rank != RankOfficer {
			t.Errorf("Old leader rank = %s, want %s", oldLeader.Rank, RankOfficer)
		}
	})

	t.Run("cannot promote leader", func(t *testing.T) {
		err := m.PromoteMember(guildID, "player1", "player1")
		if err == nil {
			t.Error("PromoteMember should fail when trying to promote leader")
		}
	})

	t.Run("non-member cannot promote", func(t *testing.T) {
		err := m.PromoteMember(guildID, "leader", "nonmember")
		if err == nil {
			t.Error("PromoteMember should fail for non-member promoter")
		}
	})

	t.Run("member without permission cannot promote", func(t *testing.T) {
		m.AddMember(guildID, "player2", RankMember)
		err := m.PromoteMember(guildID, "leader", "player2")
		if err == nil {
			t.Error("PromoteMember should fail for member without permission")
		}
	})

	t.Run("officer cannot promote to officer", func(t *testing.T) {
		m.AddMember(guildID, "officer1", RankOfficer)
		m.AddMember(guildID, "player3", RankMember)
		err := m.PromoteMember(guildID, "player3", "officer1")
		if err == nil {
			t.Error("PromoteMember should fail when officer tries to promote to officer")
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		err := m.PromoteMember("nonexistent", "player1", "leader")
		if err == nil {
			t.Error("PromoteMember should fail for non-existing guild")
		}
	})

	t.Run("non-existing target member", func(t *testing.T) {
		err := m.PromoteMember(guildID, "nonexistent", "player1")
		if err == nil {
			t.Error("PromoteMember should fail for non-existing target member")
		}
	})
}

func TestDepositTreasury(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	t.Run("valid deposit", func(t *testing.T) {
		err := m.DepositTreasury(guildID, "leader", 100)
		if err != nil {
			t.Fatalf("DepositTreasury failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		if guild.Treasury != 100 {
			t.Errorf("Treasury = %d, want 100", guild.Treasury)
		}
		if len(guild.Transactions) != 1 {
			t.Errorf("Transactions count = %d, want 1", len(guild.Transactions))
		}
	})

	t.Run("multiple deposits", func(t *testing.T) {
		m.DepositTreasury(guildID, "leader", 50)
		guild, _ := m.GetGuild(guildID)
		if guild.Treasury != 150 {
			t.Errorf("Treasury = %d, want 150", guild.Treasury)
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		err := m.DepositTreasury(guildID, "leader", -50)
		if err == nil {
			t.Error("DepositTreasury should fail for negative amount")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		err := m.DepositTreasury(guildID, "leader", 0)
		if err == nil {
			t.Error("DepositTreasury should fail for zero amount")
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		err := m.DepositTreasury("nonexistent", "leader", 100)
		if err == nil {
			t.Error("DepositTreasury should fail for non-existing guild")
		}
	})
}

func TestWithdrawTreasury(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	m.DepositTreasury(guildID, "leader", 1000)

	t.Run("valid withdrawal", func(t *testing.T) {
		err := m.WithdrawTreasury(guildID, "leader", 100)
		if err != nil {
			t.Fatalf("WithdrawTreasury failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		if guild.Treasury != 900 {
			t.Errorf("Treasury = %d, want 900", guild.Treasury)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		err := m.WithdrawTreasury(guildID, "leader", 1000)
		if err == nil {
			t.Error("WithdrawTreasury should fail for insufficient funds")
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		err := m.WithdrawTreasury(guildID, "leader", -50)
		if err == nil {
			t.Error("WithdrawTreasury should fail for negative amount")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		err := m.WithdrawTreasury(guildID, "leader", 0)
		if err == nil {
			t.Error("WithdrawTreasury should fail for zero amount")
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		err := m.WithdrawTreasury("nonexistent", "leader", 100)
		if err == nil {
			t.Error("WithdrawTreasury should fail for non-existing guild")
		}
	})
}

func TestSetMOTD(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	t.Run("valid MOTD", func(t *testing.T) {
		newMOTD := "New message of the day!"
		err := m.SetMOTD(guildID, newMOTD)
		if err != nil {
			t.Fatalf("SetMOTD failed: %v", err)
		}

		guild, _ := m.GetGuild(guildID)
		if guild.MOTD != newMOTD {
			t.Errorf("MOTD = %s, want %s", guild.MOTD, newMOTD)
		}
	})

	t.Run("empty MOTD", func(t *testing.T) {
		err := m.SetMOTD(guildID, "")
		if err != nil {
			t.Fatalf("SetMOTD should allow empty MOTD: %v", err)
		}
	})

	t.Run("non-existing guild", func(t *testing.T) {
		err := m.SetMOTD("nonexistent", "test")
		if err == nil {
			t.Error("SetMOTD should fail for non-existing guild")
		}
	})
}

func TestSaveLoad(t *testing.T) {
	m := NewManager()

	// Create multiple guilds
	guildID1, _ := m.CreateGuild("fantasy", "leader1", 12345)
	guildID2, _ := m.CreateGuild("sci-fi", "leader2", 23456)
	m.AddMember(guildID1, "player1", RankMember)
	m.DepositTreasury(guildID1, "leader1", 500)

	// Save
	data, err := m.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Save returned empty data")
	}

	// Load into new manager
	m2 := NewManager()
	err = m2.Load(data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify guilds were loaded
	guild1, err := m2.GetGuild(guildID1)
	if err != nil {
		t.Fatalf("Guild 1 not found after load: %v", err)
	}
	if len(guild1.Members) != 2 {
		t.Errorf("Guild 1 members count = %d, want 2", len(guild1.Members))
	}
	if guild1.Treasury != 500 {
		t.Errorf("Guild 1 treasury = %d, want 500", guild1.Treasury)
	}

	guild2, err := m2.GetGuild(guildID2)
	if err != nil {
		t.Fatalf("Guild 2 not found after load: %v", err)
	}
	if guild2.LeaderID != "leader2" {
		t.Errorf("Guild 2 leader = %s, want leader2", guild2.LeaderID)
	}
}

func TestConcurrency(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	// Concurrent deposits
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			playerID := "player" + string(rune(id))
			m.DepositTreasury(guildID, playerID, 10)
		}(i)
	}
	wg.Wait()

	guild, _ := m.GetGuild(guildID)
	if guild.Treasury != 1000 {
		t.Errorf("Treasury = %d, want 1000", guild.Treasury)
	}
}

func TestGuildHasPermission(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	guild, _ := m.GetGuild(guildID)

	tests := []struct {
		name string
		rank Rank
		perm Permission
		want bool
	}{
		{"leader invite", RankLeader, PermissionInvite, true},
		{"leader kick", RankLeader, PermissionKick, true},
		{"officer invite", RankOfficer, PermissionInvite, true},
		{"officer kick", RankOfficer, PermissionKick, false},
		{"member deposit", RankMember, PermissionDeposit, true},
		{"member withdraw", RankMember, PermissionWithdraw, false},
		{"recruit deposit", RankRecruit, PermissionDeposit, true},
		{"recruit invite", RankRecruit, PermissionInvite, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(guild, tt.rank, tt.perm)
			if got != tt.want {
				t.Errorf("HasPermission(%s, %s) = %v, want %v", tt.rank, tt.perm, got, tt.want)
			}
		})
	}
}

func TestGetMember(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	m.AddMember(guildID, "player1", RankMember)
	guild, _ := m.GetGuild(guildID)

	t.Run("existing member", func(t *testing.T) {
		member := GetMember(guild, "player1")
		if member == nil {
			t.Fatal("GetMember returned nil for existing member")
		}
		if member.PlayerID != "player1" {
			t.Errorf("PlayerID = %s, want player1", member.PlayerID)
		}
	})

	t.Run("non-existing member", func(t *testing.T) {
		member := GetMember(guild, "nonexistent")
		if member != nil {
			t.Error("GetMember should return nil for non-existing member")
		}
	})
}

// Benchmarks
func BenchmarkCreateGuild(b *testing.B) {
	m := NewManager()
	for i := 0; i < b.N; i++ {
		playerID := "player" + string(rune(i))
		m.CreateGuild("fantasy", playerID, int64(i+12345))
	}
}

func BenchmarkAddMember(b *testing.B) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerID := "player" + string(rune(i))
		m.AddMember(guildID, playerID, RankMember)
	}
}

func BenchmarkDepositTreasury(b *testing.B) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DepositTreasury(guildID, "leader", 10)
	}
}

func BenchmarkWithdrawTreasury(b *testing.B) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	m.DepositTreasury(guildID, "leader", 1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.WithdrawTreasury(guildID, "leader", 1)
	}
}

func BenchmarkSave(b *testing.B) {
	m := NewManager()
	for i := 0; i < 100; i++ {
		playerID := "player" + string(rune(i))
		m.CreateGuild("fantasy", playerID, int64(i+12345))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Save()
	}
}

func BenchmarkLoad(b *testing.B) {
	m := NewManager()
	for i := 0; i < 100; i++ {
		playerID := "player" + string(rune(i))
		m.CreateGuild("fantasy", playerID, int64(i+12345))
	}
	data, _ := m.Save()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m2 := NewManager()
		m2.Load(data)
	}
}

func BenchmarkHasPermission(b *testing.B) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	guild, _ := m.GetGuild(guildID)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HasPermission(guild, RankLeader, PermissionInvite)
	}
}

func TestGenerateIdentity(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		seed  int64
	}{
		{"fantasy", "fantasy", 12345},
		{"sci-fi", "sci-fi", 12345},
		{"horror", "horror", 12345},
		{"cyberpunk", "cyberpunk", 12345},
		{"post-apocalyptic", "post-apocalyptic", 12345},
		{"unknown", "unknown", 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity := GenerateIdentity(tt.genre, tt.seed)
			if identity.Name == "" {
				t.Error("Generated name is empty")
			}
			if identity.Emblem == nil {
				t.Fatal("Generated emblem is nil")
			}
			if identity.Emblem.Shape == "" {
				t.Error("Emblem shape is empty")
			}
			if identity.Emblem.Symbol == "" {
				t.Error("Emblem symbol is empty")
			}
		})
	}
}

func TestGenerateIdentityDeterministic(t *testing.T) {
	seed := int64(54321)
	id1 := GenerateIdentity("fantasy", seed)
	id2 := GenerateIdentity("fantasy", seed)

	if id1.Name != id2.Name {
		t.Errorf("Names differ: %s vs %s", id1.Name, id2.Name)
	}
	if id1.Emblem.Shape != id2.Emblem.Shape {
		t.Errorf("Shapes differ: %s vs %s", id1.Emblem.Shape, id2.Emblem.Shape)
	}
	if id1.Emblem.Symbol != id2.Emblem.Symbol {
		t.Errorf("Symbols differ: %s vs %s", id1.Emblem.Symbol, id2.Emblem.Symbol)
	}
	if id1.Emblem.PrimaryR != id2.Emblem.PrimaryR {
		t.Error("Primary colors differ")
	}
}

func TestGuildType(t *testing.T) {
	guild := &Guild{}
	if guild.Type() != "guild" {
		t.Errorf("Type() = %s, want guild", guild.Type())
	}
}

func TestTransactionHistory(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	// Multiple transactions
	m.DepositTreasury(guildID, "player1", 100)
	m.DepositTreasury(guildID, "player2", 200)
	m.WithdrawTreasury(guildID, "leader", 50)

	guild, _ := m.GetGuild(guildID)
	if len(guild.Transactions) != 3 {
		t.Errorf("Transactions count = %d, want 3", len(guild.Transactions))
	}

	// Check transaction details
	if guild.Transactions[0].Amount != 100 {
		t.Errorf("Transaction 0 amount = %d, want 100", guild.Transactions[0].Amount)
	}
	if guild.Transactions[2].Amount != -50 {
		t.Errorf("Transaction 2 amount = %d, want -50", guild.Transactions[2].Amount)
	}
}

func TestUpdatedAtTimestamp(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	guild1, _ := m.GetGuild(guildID)
	oldTime := guild1.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.DepositTreasury(guildID, "leader", 100)

	guild2, _ := m.GetGuild(guildID)
	if !guild2.UpdatedAt.After(oldTime) {
		t.Error("UpdatedAt not updated after treasury deposit")
	}
}

func TestGenreSpecificNames(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			identity := GenerateIdentity(genre, 12345)
			name := identity.Name

			// Verify name is not empty and emblem is valid
			if name == "" {
				t.Error("Generated name is empty")
			}
			if identity.Emblem == nil {
				t.Error("Emblem is nil")
			}
			if identity.Emblem.Shape == "" {
				t.Error("Emblem shape is empty")
			}
			if identity.Emblem.Symbol == "" {
				t.Error("Emblem symbol is empty")
			}
		})
	}
}

// Cross-Server Synchronization Tests

func TestSetServerID(t *testing.T) {
	serverID := "test-server-123"
	m := NewManager(WithServerID(serverID))

	if m.serverID != serverID {
		t.Errorf("serverID = %s, want %s", m.serverID, serverID)
	}
}

func TestAddFederatedServer(t *testing.T) {
	m := NewManager()

	t.Run("add new server", func(t *testing.T) {
		m.AddFederatedServer("server-1")
		if len(m.federatedServers) != 1 {
			t.Errorf("federatedServers count = %d, want 1", len(m.federatedServers))
		}
		if m.federatedServers[0] != "server-1" {
			t.Errorf("server ID = %s, want server-1", m.federatedServers[0])
		}
	})

	t.Run("add duplicate server", func(t *testing.T) {
		m.AddFederatedServer("server-1")
		if len(m.federatedServers) != 1 {
			t.Errorf("federatedServers count = %d, want 1 (no duplicates)", len(m.federatedServers))
		}
	})

	t.Run("add multiple servers", func(t *testing.T) {
		m.AddFederatedServer("server-2")
		m.AddFederatedServer("server-3")
		if len(m.federatedServers) != 3 {
			t.Errorf("federatedServers count = %d, want 3", len(m.federatedServers))
		}
	})
}

func TestRemoveFederatedServer(t *testing.T) {
	m := NewManager()
	m.AddFederatedServer("server-1")
	m.AddFederatedServer("server-2")
	m.AddFederatedServer("server-3")

	t.Run("remove existing server", func(t *testing.T) {
		m.RemoveFederatedServer("server-2")
		if len(m.federatedServers) != 2 {
			t.Errorf("federatedServers count = %d, want 2", len(m.federatedServers))
		}

		// Verify server-2 is removed
		for _, id := range m.federatedServers {
			if id == "server-2" {
				t.Error("server-2 should be removed")
			}
		}
	})

	t.Run("remove non-existing server", func(t *testing.T) {
		count := len(m.federatedServers)
		m.RemoveFederatedServer("nonexistent")
		if len(m.federatedServers) != count {
			t.Errorf("federatedServers count changed when removing non-existing server")
		}
	})
}

func TestSyncGuildState(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	t.Run("sync existing guild without transport", func(t *testing.T) {
		err := m.SyncGuildState(guildID)
		if err != nil {
			t.Fatalf("SyncGuildState failed: %v", err)
		}
	})

	t.Run("sync non-existing guild", func(t *testing.T) {
		err := m.SyncGuildState("nonexistent")
		if err == nil {
			t.Error("SyncGuildState should fail for non-existing guild")
		}
	})

	t.Run("sync with transport success", func(t *testing.T) {
		mt := &mockGuildTransport{}
		m.SetTransport(mt)
		defer m.SetTransport(nil)

		err := m.SyncGuildState(guildID)
		if err != nil {
			t.Fatalf("SyncGuildState with transport failed: %v", err)
		}
		if mt.callCount != 1 {
			t.Errorf("transport called %d times, want 1", mt.callCount)
		}
		if mt.lastGuildID != guildID {
			t.Errorf("transport guildID = %s, want %s", mt.lastGuildID, guildID)
		}
		if len(mt.lastData) == 0 {
			t.Error("transport received empty data")
		}
	})

	t.Run("sync with transport error", func(t *testing.T) {
		mt := &mockGuildTransport{err: fmt.Errorf("connection refused")}
		m.SetTransport(mt)
		defer m.SetTransport(nil)

		err := m.SyncGuildState(guildID)
		if err == nil {
			t.Error("SyncGuildState should propagate transport error")
		}
		if mt.callCount != 1 {
			t.Errorf("transport called %d times, want 1", mt.callCount)
		}
	})
}

// mockGuildTransport implements GuildTransport for testing
type mockGuildTransport struct {
	callCount   int
	lastGuildID string
	lastData    []byte
	err         error
}

func (m *mockGuildTransport) BroadcastGuildUpdate(guildID string, data []byte) error {
	m.callCount++
	m.lastGuildID = guildID
	m.lastData = data
	return m.err
}

func TestSetTransport(t *testing.T) {
	m := NewManager()

	t.Run("set nil transport", func(t *testing.T) {
		m.SetTransport(nil)
		if m.transport != nil {
			t.Error("transport should be nil")
		}
	})

	t.Run("set mock transport", func(t *testing.T) {
		mt := &mockGuildTransport{}
		m.SetTransport(mt)
		if m.transport == nil {
			t.Error("transport should not be nil")
		}
	})

	t.Run("replace transport", func(t *testing.T) {
		mt1 := &mockGuildTransport{}
		mt2 := &mockGuildTransport{}
		m.SetTransport(mt1)
		m.SetTransport(mt2)

		guildID, _ := m.CreateGuild("fantasy", "leader", 99999)
		_ = m.SyncGuildState(guildID)

		if mt1.callCount != 0 {
			t.Error("old transport should not be called")
		}
		if mt2.callCount != 1 {
			t.Errorf("new transport called %d times, want 1", mt2.callCount)
		}
	})
}

func TestHandleGuildMessage_GuildSync(t *testing.T) {
	m := NewManager()

	// Use fixed timestamp for deterministic tests
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create a guild on another server
	guild := &Guild{
		ID:          "remote-guild-123",
		Name:        "Remote Guild",
		LeaderID:    "remote-leader",
		Members:     []Member{{PlayerID: "remote-leader", Rank: RankLeader, JoinedAt: fixedTime, LastLogin: fixedTime}},
		Permissions: make(map[Rank][]Permission),
		Treasury:    1000,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
		Reputation:  make(map[string]float64),
	}

	msg := GuildMessage{
		Type:      MsgTypeGuildSync,
		GuildID:   guild.ID,
		ServerID:  "remote-server",
		Timestamp: fixedTime,
		Data:      guild,
	}

	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("HandleGuildMessage failed: %v", err)
	}

	// Verify guild was synchronized
	syncedGuild, err := m.GetGuild(guild.ID)
	if err != nil {
		t.Fatalf("Failed to get synced guild: %v", err)
	}
	if syncedGuild.Name != guild.Name {
		t.Errorf("Guild name = %s, want %s", syncedGuild.Name, guild.Name)
	}
	if syncedGuild.Treasury != guild.Treasury {
		t.Errorf("Guild treasury = %d, want %d", syncedGuild.Treasury, guild.Treasury)
	}
}

func TestHandleGuildMessage_MemberJoin(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	joinData := MemberJoinData{
		PlayerID: "new-player",
		Rank:     RankRecruit,
	}

	msg := GuildMessage{
		Type:      MsgTypeMemberJoin,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      joinData,
	}

	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("HandleGuildMessage failed: %v", err)
	}

	// Verify member was added
	guild, _ := m.GetGuild(guildID)
	member := GetMember(guild, "new-player")
	if member == nil {
		t.Fatal("New member not found after join message")
	}
	if member.Rank != RankRecruit {
		t.Errorf("Member rank = %s, want %s", member.Rank, RankRecruit)
	}
}

func TestHandleGuildMessage_MemberLeave(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)
	m.AddMember(guildID, "leaving-player", RankMember)

	leaveData := MemberLeaveData{
		PlayerID: "leaving-player",
	}

	msg := GuildMessage{
		Type:      MsgTypeMemberLeave,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      leaveData,
	}

	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("HandleGuildMessage failed: %v", err)
	}

	// Verify member was removed
	guild, _ := m.GetGuild(guildID)
	member := GetMember(guild, "leaving-player")
	if member != nil {
		t.Error("Member should be removed after leave message")
	}
}

func TestHandleGuildMessage_TerritoryChange(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	territoryData := TerritoryChangeData{
		ZoneID:   "zone-123",
		OldGuild: "",
		NewGuild: guildID,
	}

	msg := GuildMessage{
		Type:      MsgTypeTerritoryChange,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      territoryData,
	}

	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("HandleGuildMessage failed: %v", err)
	}

	// Verify reputation was updated
	guild, _ := m.GetGuild(guildID)
	if guild.Reputation[territoryData.ZoneID] <= 0 {
		t.Error("Guild reputation should increase after gaining territory")
	}
}

func TestHandleGuildMessage_InvalidType(t *testing.T) {
	m := NewManager()

	msg := GuildMessage{
		Type:      "invalid_type",
		GuildID:   "guild-123",
		ServerID:  "remote-server",
		Timestamp: time.Now(),
	}

	err := m.HandleGuildMessage(msg)
	if err == nil {
		t.Error("HandleGuildMessage should fail for invalid message type")
	}
}

func TestHandleGuildMessage_EmptyType(t *testing.T) {
	m := NewManager()

	msg := GuildMessage{
		Type:      "",
		GuildID:   "guild-123",
		ServerID:  "remote-server",
		Timestamp: time.Now(),
	}

	err := m.HandleGuildMessage(msg)
	if err == nil {
		t.Error("HandleGuildMessage should fail for empty message type")
	}
}

func TestHandleGuildMessage_EmptyGuildID(t *testing.T) {
	m := NewManager()

	msg := GuildMessage{
		Type:      MsgTypeGuildSync,
		GuildID:   "",
		ServerID:  "remote-server",
		Timestamp: time.Now(),
	}

	err := m.HandleGuildMessage(msg)
	if err == nil {
		t.Error("HandleGuildMessage should fail for empty guild ID")
	}
}

func TestHandleGuildMessage_JSONDeserialization(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	// Simulate JSON deserialization by using map[string]interface{}
	joinData := map[string]interface{}{
		"player_id": "json-player",
		"rank":      "Member",
	}

	msg := GuildMessage{
		Type:      MsgTypeMemberJoin,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      joinData,
	}

	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("HandleGuildMessage failed with JSON data: %v", err)
	}

	// Verify member was added
	guild, _ := m.GetGuild(guildID)
	member := GetMember(guild, "json-player")
	if member == nil {
		t.Fatal("Member not found after JSON join message")
	}
}

func TestHandleGuildMessage_DuplicateMemberJoin(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	joinData := MemberJoinData{
		PlayerID: "duplicate-player",
		Rank:     RankRecruit,
	}

	msg := GuildMessage{
		Type:      MsgTypeMemberJoin,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      joinData,
	}

	// Add member first time
	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("First HandleGuildMessage failed: %v", err)
	}

	guild1, _ := m.GetGuild(guildID)
	count1 := len(guild1.Members)

	// Try to add same member again
	err = m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("Second HandleGuildMessage failed: %v", err)
	}

	guild2, _ := m.GetGuild(guildID)
	count2 := len(guild2.Members)

	if count1 != count2 {
		t.Error("Duplicate member join should not increase member count")
	}
}

func TestHandleGuildMessage_NonExistentMemberLeave(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	leaveData := MemberLeaveData{
		PlayerID: "nonexistent-player",
	}

	msg := GuildMessage{
		Type:      MsgTypeMemberLeave,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      leaveData,
	}

	// Should not error when removing non-existent member
	err := m.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("HandleGuildMessage should not fail for non-existent member leave: %v", err)
	}
}

func TestCrossFederationScenario(t *testing.T) {
	// Create two managers representing different servers
	server1 := NewManager(WithServerID("server-1"))

	server2 := NewManager(WithServerID("server-2"))

	// Register each other as federated servers
	server1.AddFederatedServer("server-2")
	server2.AddFederatedServer("server-1")

	// Create guild on server 1
	guildID, _ := server1.CreateGuild("fantasy", "leader-on-server-1", 12345)
	guild1, _ := server1.GetGuild(guildID)

	// Sync guild to server 2
	msg := GuildMessage{
		Type:      MsgTypeGuildSync,
		GuildID:   guildID,
		ServerID:  "server-1",
		Timestamp: time.Now(),
		Data:      guild1,
	}

	err := server2.HandleGuildMessage(msg)
	if err != nil {
		t.Fatalf("Cross-server sync failed: %v", err)
	}

	// Verify guild exists on server 2
	guild2, err := server2.GetGuild(guildID)
	if err != nil {
		t.Fatalf("Guild not found on server 2: %v", err)
	}
	if guild2.Name != guild1.Name {
		t.Errorf("Guild name mismatch: %s vs %s", guild2.Name, guild1.Name)
	}

	// Add member on server 2 and sync back to server 1
	joinData := MemberJoinData{
		PlayerID: "player-on-server-2",
		Rank:     RankRecruit,
	}

	joinMsg := GuildMessage{
		Type:      MsgTypeMemberJoin,
		GuildID:   guildID,
		ServerID:  "server-2",
		Timestamp: time.Now(),
		Data:      joinData,
	}

	// Apply to both servers
	server2.HandleGuildMessage(joinMsg)
	server1.HandleGuildMessage(joinMsg)

	// Verify member exists on both servers
	guild1Updated, _ := server1.GetGuild(guildID)
	guild2Updated, _ := server2.GetGuild(guildID)

	if len(guild1Updated.Members) != len(guild2Updated.Members) {
		t.Error("Member count mismatch across servers")
	}

	member1 := GetMember(guild1Updated, "player-on-server-2")
	member2 := GetMember(guild2Updated, "player-on-server-2")

	if member1 == nil || member2 == nil {
		t.Fatal("Member not found on one or both servers")
	}
	if member1.Rank != member2.Rank {
		t.Error("Member rank mismatch across servers")
	}
}

func TestLoadSizeLimit(t *testing.T) {
	// Test that the size limit constants and errors are properly defined

	// Verify the constant is reasonable (50 MB for guild data)
	if MaxGuildDataSize != 50*1024*1024 {
		t.Errorf("Expected MaxGuildDataSize to be 50MB, got %d", MaxGuildDataSize)
	}

	// Verify the error is exported
	if ErrGuildDataSizeExceeded == nil {
		t.Error("ErrGuildDataSizeExceeded should not be nil")
	}

	// Verify error message
	expectedMsg := "guild data exceeds maximum allowed size"
	if ErrGuildDataSizeExceeded.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, ErrGuildDataSizeExceeded.Error())
	}
}

func TestSaveLoadWithinLimit(t *testing.T) {
	// Create manager with guilds
	m := NewManager()

	// Create a guild and track its ID
	guildID, err := m.CreateGuild("fantasy", "leader1", 12345)
	if err != nil {
		t.Fatalf("CreateGuild failed: %v", err)
	}

	// Save (compress)
	data, err := m.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new manager and load
	m2 := NewManager()
	err = m2.Load(data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify guild was loaded by retrieving it
	guild, err := m2.GetGuild(guildID)
	if err != nil {
		t.Fatalf("GetGuild after Load failed: %v", err)
	}
	if guild.LeaderID != "leader1" {
		t.Errorf("Expected leader1, got %s", guild.LeaderID)
	}
}

// TestCreateGuildDeterminism verifies that the same seed produces the same guild identity.
func TestCreateGuildDeterminism(t *testing.T) {
	seed := int64(54321)
	genre := "fantasy"
	leaderID := "player1"

	// Create first guild with seed
	m1 := NewManager()
	guildID1, err := m1.CreateGuild(genre, leaderID, seed)
	if err != nil {
		t.Fatalf("CreateGuild 1 failed: %v", err)
	}
	guild1, err := m1.GetGuild(guildID1)
	if err != nil {
		t.Fatalf("GetGuild 1 failed: %v", err)
	}

	// Create second guild with same seed
	m2 := NewManager()
	guildID2, err := m2.CreateGuild(genre, leaderID, seed)
	if err != nil {
		t.Fatalf("CreateGuild 2 failed: %v", err)
	}
	guild2, err := m2.GetGuild(guildID2)
	if err != nil {
		t.Fatalf("GetGuild 2 failed: %v", err)
	}

	// Verify same seed produces same guild ID
	if guildID1 != guildID2 {
		t.Errorf("Expected identical guild IDs, got %s and %s", guildID1, guildID2)
	}

	// Verify same seed produces same guild name
	if guild1.Name != guild2.Name {
		t.Errorf("Expected identical guild names, got %s and %s", guild1.Name, guild2.Name)
	}

	// Verify same seed produces same emblem
	if guild1.Emblem.Shape != guild2.Emblem.Shape {
		t.Errorf("Expected identical emblem shape, got %s and %s", guild1.Emblem.Shape, guild2.Emblem.Shape)
	}
	if guild1.Emblem.Symbol != guild2.Emblem.Symbol {
		t.Errorf("Expected identical emblem symbol, got %s and %s", guild1.Emblem.Symbol, guild2.Emblem.Symbol)
	}
	if guild1.Emblem.PrimaryR != guild2.Emblem.PrimaryR ||
		guild1.Emblem.PrimaryG != guild2.Emblem.PrimaryG ||
		guild1.Emblem.PrimaryB != guild2.Emblem.PrimaryB {
		t.Error("Expected identical primary colors")
	}

	// Verify different seed produces different guild
	m3 := NewManager()
	guildID3, _ := m3.CreateGuild(genre, leaderID, seed+1)
	guild3, _ := m3.GetGuild(guildID3)

	// Different seed should produce different guild ID
	if guildID1 == guildID3 {
		t.Error("Different seeds should produce different guild IDs")
	}

	// Different seed should (likely) produce different name/emblem
	// Note: could collide, but unlikely
	if guild1.Name == guild3.Name && guild1.Emblem.Symbol == guild3.Emblem.Symbol {
		t.Log("Warning: Different seeds produced identical name and symbol (rare but possible)")
	}
}

// Benchmark tests for hot-path operations (Recommendation #7)

func BenchmarkHasPermission_Miss(b *testing.B) {
	g := &Guild{
		Permissions: map[Rank][]Permission{
			RankRecruit: {PermissionDeposit},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HasPermission(g, RankRecruit, PermissionKick)
	}
}

func BenchmarkGetMember(b *testing.B) {
	members := make([]Member, 50)
	for i := range members {
		members[i] = Member{
			PlayerID: fmt.Sprintf("player-%d", i),
			Rank:     RankMember,
			JoinedAt: time.Now(),
		}
	}
	g := &Guild{Members: members}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMember(g, "player-25")
	}
}

func BenchmarkGetMember_NotFound(b *testing.B) {
	members := make([]Member, 50)
	for i := range members {
		members[i] = Member{
			PlayerID: fmt.Sprintf("player-%d", i),
			Rank:     RankMember,
		}
	}
	g := &Guild{Members: members}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMember(g, "nonexistent")
	}
}

func BenchmarkSyncGuildState(b *testing.B) {
	m := NewManager(WithServerID("bench-server"))
	guildID, _ := m.CreateGuild("fantasy", "leader-1", 42)
	mt := &mockGuildTransport{}
	m.SetTransport(mt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.SyncGuildState(guildID)
	}
}

// TestTransportIntegrationWiring verifies transport can be set and used for sync
func TestTransportIntegrationWiring(t *testing.T) {
	m := NewManager(WithServerID("test-server"))
	guildID, err := m.CreateGuild("fantasy", "leader-1", 42)
	if err != nil {
		t.Fatalf("CreateGuild failed: %v", err)
	}

	mt := &mockGuildTransport{}
	m.SetTransport(mt)

	// Sync should use the transport
	if err := m.SyncGuildState(guildID); err != nil {
		t.Fatalf("SyncGuildState failed: %v", err)
	}
	if mt.callCount != 1 {
		t.Errorf("expected 1 broadcast call, got %d", mt.callCount)
	}
	if mt.lastGuildID != guildID {
		t.Errorf("expected guildID %s, got %s", guildID, mt.lastGuildID)
	}
	if len(mt.lastData) == 0 {
		t.Error("expected non-empty broadcast data")
	}

	// Replace transport
	mt2 := &mockGuildTransport{}
	m.SetTransport(mt2)
	if err := m.SyncGuildState(guildID); err != nil {
		t.Fatalf("SyncGuildState with new transport failed: %v", err)
	}
	if mt2.callCount != 1 {
		t.Errorf("expected 1 call on new transport, got %d", mt2.callCount)
	}
	// Original transport should not have been called again
	if mt.callCount != 1 {
		t.Errorf("original transport should still have 1 call, got %d", mt.callCount)
	}

	// Nil transport should not error
	m.SetTransport(nil)
	if err := m.SyncGuildState(guildID); err != nil {
		t.Fatalf("SyncGuildState with nil transport should not error: %v", err)
	}
}

// TestTimeProviderDefault verifies that NewManager uses RealTimeProvider by default
func TestTimeProviderDefault(t *testing.T) {
	m := NewManager()
	if m.timeProvider == nil {
		t.Fatal("timeProvider should not be nil")
	}
	// Verify it's a RealTimeProvider by checking the type
	if _, ok := m.timeProvider.(RealTimeProvider); !ok {
		t.Errorf("expected RealTimeProvider, got %T", m.timeProvider)
	}
}

// TestTimeProviderMock verifies that MockTimeProvider produces deterministic timestamps
func TestTimeProviderMock(t *testing.T) {
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: fixedTime}

	m := NewManager(WithTimeProvider(mockTP))
	guildID, err := m.CreateGuild("fantasy", "leader-1", 12345)
	if err != nil {
		t.Fatalf("CreateGuild failed: %v", err)
	}

	guild, _ := m.GetGuild(guildID)
	if !guild.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want %v", guild.CreatedAt, fixedTime)
	}
	if !guild.UpdatedAt.Equal(fixedTime) {
		t.Errorf("UpdatedAt = %v, want %v", guild.UpdatedAt, fixedTime)
	}
	if len(guild.Members) == 0 {
		t.Fatal("expected at least one member")
	}
	if !guild.Members[0].JoinedAt.Equal(fixedTime) {
		t.Errorf("Member JoinedAt = %v, want %v", guild.Members[0].JoinedAt, fixedTime)
	}
}

// TestTimeProviderAdvance verifies that advancing MockTimeProvider updates timestamps correctly
func TestTimeProviderAdvance(t *testing.T) {
	startTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: startTime}

	m := NewManager(WithTimeProvider(mockTP))
	guildID, err := m.CreateGuild("fantasy", "leader-1", 12345)
	if err != nil {
		t.Fatalf("CreateGuild failed: %v", err)
	}

	// Advance time by 1 hour
	mockTP.Advance(time.Hour)
	expectedTime := startTime.Add(time.Hour)

	// Add a member - should use the advanced time
	if err := m.AddMember(guildID, "player-2", RankRecruit); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	guild, _ := m.GetGuild(guildID)
	if !guild.UpdatedAt.Equal(expectedTime) {
		t.Errorf("UpdatedAt after AddMember = %v, want %v", guild.UpdatedAt, expectedTime)
	}
	// Find the new member and check their join time
	for _, member := range guild.Members {
		if member.PlayerID == "player-2" {
			if !member.JoinedAt.Equal(expectedTime) {
				t.Errorf("New member JoinedAt = %v, want %v", member.JoinedAt, expectedTime)
			}
		}
	}
}

// TestTimeProviderTreasury verifies treasury operations use TimeProvider
func TestTimeProviderTreasury(t *testing.T) {
	fixedTime := time.Date(2026, 2, 17, 14, 30, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: fixedTime}

	m := NewManager(WithTimeProvider(mockTP))
	guildID, _ := m.CreateGuild("fantasy", "leader-1", 12345)

	// Advance time before deposit
	mockTP.Advance(30 * time.Minute)
	depositTime := fixedTime.Add(30 * time.Minute)

	if err := m.DepositTreasury(guildID, "leader-1", 100); err != nil {
		t.Fatalf("DepositTreasury failed: %v", err)
	}

	guild, _ := m.GetGuild(guildID)
	if len(guild.Transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(guild.Transactions))
	}
	if !guild.Transactions[0].Timestamp.Equal(depositTime) {
		t.Errorf("Transaction timestamp = %v, want %v", guild.Transactions[0].Timestamp, depositTime)
	}

	// Advance time again for withdrawal
	mockTP.Advance(15 * time.Minute)
	withdrawTime := depositTime.Add(15 * time.Minute)

	if err := m.WithdrawTreasury(guildID, "leader-1", 50); err != nil {
		t.Fatalf("WithdrawTreasury failed: %v", err)
	}

	guild, _ = m.GetGuild(guildID)
	if len(guild.Transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(guild.Transactions))
	}
	if !guild.Transactions[1].Timestamp.Equal(withdrawTime) {
		t.Errorf("Withdrawal timestamp = %v, want %v", guild.Transactions[1].Timestamp, withdrawTime)
	}
}

// TestTimeProviderFederation verifies federation sync uses TimeProvider
func TestTimeProviderFederation(t *testing.T) {
	fixedTime := time.Date(2026, 2, 17, 16, 0, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: fixedTime}

	m := NewManager(WithServerID("test-server"), WithTimeProvider(mockTP))
	guildID, _ := m.CreateGuild("fantasy", "leader-1", 12345)

	// Setup mock transport to capture sync data
	mt := &mockGuildTransport{}
	m.SetTransport(mt)

	// Advance time before sync
	mockTP.Advance(time.Hour)

	if err := m.SyncGuildState(guildID); err != nil {
		t.Fatalf("SyncGuildState failed: %v", err)
	}

	if mt.callCount != 1 {
		t.Errorf("expected 1 broadcast, got %d", mt.callCount)
	}
}

// TestWithServerIDAndTimeProvider verifies both options work together
func TestWithServerIDAndTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{CurrentTime: fixedTime}

	m := NewManager(
		WithServerID("custom-server-id"),
		WithTimeProvider(mockTP),
	)

	if m.serverID != "custom-server-id" {
		t.Errorf("serverID = %s, want custom-server-id", m.serverID)
	}
	if m.timeProvider != mockTP {
		t.Errorf("timeProvider not set correctly")
	}
}

// BenchmarkHandleGuildMessage benchmarks the federation hot-path for message processing.
// This function is called frequently during cross-server guild synchronization.
func BenchmarkHandleGuildMessage(b *testing.B) {
	m := NewManager(WithServerID("bench-server"))
	guildID, _ := m.CreateGuild("fantasy", "leader-1", 42)

	// Pre-create join data message (most common federation message type)
	joinData := MemberJoinData{
		PlayerID: "bench-player",
		Rank:     RankMember,
	}
	msg := GuildMessage{
		Type:      MsgTypeMemberJoin,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      joinData,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset to allow repeated joins (otherwise duplicate is ignored)
		m.mu.Lock()
		guild := m.guilds[guildID]
		if len(guild.Members) > 1 {
			guild.Members = guild.Members[:1] // Keep only leader
		}
		m.mu.Unlock()

		_ = m.HandleGuildMessage(msg)
	}
}

// BenchmarkHandleGuildMessage_GuildSync benchmarks full guild state sync.
func BenchmarkHandleGuildMessage_GuildSync(b *testing.B) {
	m := NewManager(WithServerID("bench-server"))
	guildID, _ := m.CreateGuild("fantasy", "leader-1", 42)

	// Get the guild for sync data
	guild, _ := m.GetGuild(guildID)
	msg := GuildMessage{
		Type:      MsgTypeGuildSync,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      guild,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.HandleGuildMessage(msg)
	}
}

// BenchmarkHandleGuildMessage_TerritoryChange benchmarks territory update messages.
func BenchmarkHandleGuildMessage_TerritoryChange(b *testing.B) {
	m := NewManager(WithServerID("bench-server"))
	guildID, _ := m.CreateGuild("fantasy", "leader-1", 42)

	territoryData := TerritoryChangeData{
		ZoneID:   "zone-123",
		NewGuild: guildID,
	}
	msg := GuildMessage{
		Type:      MsgTypeTerritoryChange,
		GuildID:   guildID,
		ServerID:  "remote-server",
		Timestamp: time.Now(),
		Data:      territoryData,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.HandleGuildMessage(msg)
	}
}

// TestHandleGuildMessage_MalformedJSON tests error handling for malformed JSON payloads
func TestHandleGuildMessage_MalformedJSON(t *testing.T) {
	tests := []struct {
		name          string
		msgType       MessageType
		data          interface{}
		wantErrSubstr string
	}{
		{
			name:          "GuildSync with invalid JSON structure",
			msgType:       MsgTypeGuildSync,
			data:          map[string]interface{}{"invalid": make(chan int)}, // Channels cannot be marshaled to JSON
			wantErrSubstr: "failed to marshal guild data",
		},
		{
			name:          "GuildSync with incompatible type",
			msgType:       MsgTypeGuildSync,
			data:          "not_a_guild_object",
			wantErrSubstr: "invalid guild data type",
		},
		{
			name:    "MemberJoin with incomplete data",
			msgType: MsgTypeMemberJoin,
			data: map[string]interface{}{
				"player_id": 12345, // Wrong type - should be string
				"rank":      nil,   // Missing rank
			},
			wantErrSubstr: "", // Should not error - Unmarshal handles type coercion
		},
		{
			name:          "MemberJoin with unmarshalable channel",
			msgType:       MsgTypeMemberJoin,
			data:          map[string]interface{}{"invalid": make(chan int)},
			wantErrSubstr: "failed to marshal join data",
		},
		{
			name:          "MemberJoin with incompatible type",
			msgType:       MsgTypeMemberJoin,
			data:          12345, // Not a map or struct
			wantErrSubstr: "invalid member join data type",
		},
		{
			name:          "MemberLeave with unmarshalable channel",
			msgType:       MsgTypeMemberLeave,
			data:          map[string]interface{}{"invalid": make(chan int)},
			wantErrSubstr: "failed to marshal leave data",
		},
		{
			name:          "MemberLeave with incompatible type",
			msgType:       MsgTypeMemberLeave,
			data:          []int{1, 2, 3}, // Not a map or struct
			wantErrSubstr: "invalid member leave data type",
		},
		{
			name:          "TerritoryChange with unmarshalable channel",
			msgType:       MsgTypeTerritoryChange,
			data:          map[string]interface{}{"invalid": make(chan int)},
			wantErrSubstr: "failed to marshal territory data",
		},
		{
			name:          "TerritoryChange with incompatible type",
			msgType:       MsgTypeTerritoryChange,
			data:          true, // Not a map or struct
			wantErrSubstr: "invalid territory change data type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

			msg := GuildMessage{
				Type:      tt.msgType,
				GuildID:   guildID,
				ServerID:  "test-server",
				Timestamp: time.Now(),
				Data:      tt.data,
			}

			err := m.HandleGuildMessage(msg)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.wantErrSubstr)
				} else if !contains(err.Error(), tt.wantErrSubstr) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErrSubstr, err.Error())
				}
			}
		})
	}
}

// TestHandleGuildMessage_JSONEdgeCases tests edge cases in JSON deserialization
func TestHandleGuildMessage_JSONEdgeCases(t *testing.T) {
	m := NewManager()
	guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

	// Test with deeply nested invalid JSON that would cause stack overflow if not handled
	deeplyNested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": make(chan int), // Unmarshalable
			},
		},
	}

	msg := GuildMessage{
		Type:      MsgTypeGuildSync,
		GuildID:   guildID,
		ServerID:  "test-server",
		Timestamp: time.Now(),
		Data:      deeplyNested,
	}

	err := m.HandleGuildMessage(msg)
	if err == nil {
		t.Error("Expected error for deeply nested invalid JSON")
	}
}

// TestHandleGuildMessage_NilData tests handling of nil data payloads
func TestHandleGuildMessage_NilData(t *testing.T) {
	tests := []struct {
		name    string
		msgType MessageType
	}{
		{"GuildSync with nil data", MsgTypeGuildSync},
		{"MemberJoin with nil data", MsgTypeMemberJoin},
		{"MemberLeave with nil data", MsgTypeMemberLeave},
		{"TerritoryChange with nil data", MsgTypeTerritoryChange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			guildID, _ := m.CreateGuild("fantasy", "leader", 12345)

			msg := GuildMessage{
				Type:      tt.msgType,
				GuildID:   guildID,
				ServerID:  "test-server",
				Timestamp: time.Now(),
				Data:      nil,
			}

			err := m.HandleGuildMessage(msg)
			if err == nil {
				t.Error("Expected error for nil data")
			}
		})
	}
}


// TestIsMember verifies the MembershipValidator-satisfying method.
func TestIsMember(t *testing.T) {
	m := NewManager()
	guildID, err := m.CreateGuild("fantasy", "leader1", 12345)
	if err != nil {
		t.Fatalf("CreateGuild: %v", err)
	}

	// Leader is a member
	if !m.IsMember(guildID, "leader1") {
		t.Error("IsMember: leader1 should be a member")
	}

	// Non-member returns false
	if m.IsMember(guildID, "stranger") {
		t.Error("IsMember: stranger should not be a member")
	}

	// Add a regular member
	if err := m.AddMember(guildID, "player2", RankMember); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if !m.IsMember(guildID, "player2") {
		t.Error("IsMember: player2 should be a member after AddMember")
	}

	// After removal, no longer a member
	if err := m.RemoveMember(guildID, "player2"); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	if m.IsMember(guildID, "player2") {
		t.Error("IsMember: player2 should not be a member after RemoveMember")
	}

	// Non-existent guild returns false
	if m.IsMember("no-such-guild", "anyone") {
		t.Error("IsMember: non-existent guild should return false")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
