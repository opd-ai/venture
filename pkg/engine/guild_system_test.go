package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

// PlayerComponent is a test helper component to identify player entities
type PlayerComponent struct{}

func (p *PlayerComponent) Type() string { return "player" }

func TestGuildComponent(t *testing.T) {
	t.Run("Type", func(t *testing.T) {
		comp := &GuildComponent{
			GuildID:  "guild-123",
			Rank:     "Member",
			JoinedAt: time.Now().Unix(),
		}
		if comp.Type() != "guild" {
			t.Errorf("expected type 'guild', got '%s'", comp.Type())
		}
	})

	t.Run("EmptyGuild", func(t *testing.T) {
		comp := &GuildComponent{}
		if comp.GuildID != "" {
			t.Errorf("expected empty guild ID, got '%s'", comp.GuildID)
		}
	})
}

func TestGuildSystemCreate(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	t.Run("CreateGuild", func(t *testing.T) {
		err := system.CreateGuild(player, "fantasy")
		if err != nil {
			t.Fatalf("failed to create guild: %v", err)
		}

		guildComp, ok := player.GetComponent("guild")
		if !ok || guildComp == nil {
			t.Fatal("guild component not added to player")
		}

		gc := guildComp.(*GuildComponent)
		if gc.GuildID == "" {
			t.Error("guild ID is empty")
		}
		if gc.Rank != string(guild.RankLeader) {
			t.Errorf("expected rank Leader, got %s", gc.Rank)
		}
	})

	t.Run("CreateGuildNonPlayer", func(t *testing.T) {
		nonPlayer := world.CreateEntity()
		err := system.CreateGuild(nonPlayer, "fantasy")
		if err == nil {
			t.Error("expected error creating guild for non-player")
		}
	})

	t.Run("CreateGuildAlreadyInGuild", func(t *testing.T) {
		player2 := world.CreateEntity()
		player2.AddComponent(&PlayerComponent{})

		// Create first guild
		err := system.CreateGuild(player2, "sci-fi")
		if err != nil {
			t.Fatalf("failed to create first guild: %v", err)
		}

		// Try to create second guild
		err = system.CreateGuild(player2, "horror")
		if err == nil {
			t.Error("expected error when creating second guild")
		}
	})
}

func TestGuildSystemJoin(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create leader and guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	leaderGuildComp, ok := leader.GetComponent("guild")
	if !ok {
		t.Fatalf("leader guild component not found")
	}
	guildID := leaderGuildComp.(*GuildComponent).GuildID

	t.Run("JoinGuild", func(t *testing.T) {
		member := world.CreateEntity()
		member.AddComponent(&PlayerComponent{})

		err := system.JoinGuild(member, guildID)
		if err != nil {
			t.Fatalf("failed to join guild: %v", err)
		}

		guildComp, ok := member.GetComponent("guild")
		if !ok || guildComp == nil {
			t.Fatal("guild component not added")
		}

		gc := guildComp.(*GuildComponent)
		if gc.GuildID != guildID {
			t.Errorf("expected guild ID %s, got %s", guildID, gc.GuildID)
		}
		if gc.Rank != string(guild.RankRecruit) {
			t.Errorf("expected rank Recruit, got %s", gc.Rank)
		}
	})

	t.Run("JoinGuildNonPlayer", func(t *testing.T) {
		nonPlayer := world.CreateEntity()
		err := system.JoinGuild(nonPlayer, guildID)
		if err == nil {
			t.Error("expected error joining guild as non-player")
		}
	})

	t.Run("JoinNonexistentGuild", func(t *testing.T) {
		member := world.CreateEntity()
		member.AddComponent(&PlayerComponent{})

		err := system.JoinGuild(member, "guild-nonexistent")
		if err == nil {
			t.Error("expected error joining nonexistent guild")
		}
	})

	t.Run("JoinGuildAlreadyInGuild", func(t *testing.T) {
		member := world.CreateEntity()
		member.AddComponent(&PlayerComponent{})

		// Join first time
		err := system.JoinGuild(member, guildID)
		if err != nil {
			t.Fatalf("failed to join guild: %v", err)
		}

		// Try to join again
		err = system.JoinGuild(member, guildID)
		if err == nil {
			t.Error("expected error joining guild when already a member")
		}
	})
}

func TestGuildSystemLeave(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	leaderGuildComp, ok := leader.GetComponent("guild")
	if !ok {
		t.Fatalf("leader guild component not found")
	}
	guildID := leaderGuildComp.(*GuildComponent).GuildID

	// Add member
	member := world.CreateEntity()
	member.AddComponent(&PlayerComponent{})
	err = system.JoinGuild(member, guildID)
	if err != nil {
		t.Fatalf("failed to join guild: %v", err)
	}

	t.Run("LeaveGuild", func(t *testing.T) {
		err := system.LeaveGuild(member)
		if err != nil {
			t.Fatalf("failed to leave guild: %v", err)
		}

		guildComp, ok := member.GetComponent("guild")
		if !ok {
			t.Fatal("guild component not found")
		}
		gc := guildComp.(*GuildComponent)
		if gc.GuildID != "" {
			t.Errorf("guild ID not cleared after leaving")
		}
		if gc.Rank != "" {
			t.Errorf("rank not cleared after leaving")
		}
	})

	t.Run("LeaveGuildNotInGuild", func(t *testing.T) {
		nonMember := world.CreateEntity()
		nonMember.AddComponent(&PlayerComponent{})

		err := system.LeaveGuild(nonMember)
		if err == nil {
			t.Error("expected error leaving guild when not a member")
		}
	})
}

func TestGuildSystemTreasury(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	leaderGuildComp, ok := leader.GetComponent("guild")
	if !ok {
		t.Fatalf("leader guild component not found")
	}
	guildID := leaderGuildComp.(*GuildComponent).GuildID

	t.Run("DepositTreasury", func(t *testing.T) {
		err := system.DepositTreasury(leader, 1000)
		if err != nil {
			t.Fatalf("failed to deposit: %v", err)
		}

		guildInfo, err := system.GetGuildInfo(guildID)
		if err != nil {
			t.Fatalf("failed to get guild info: %v", err)
		}

		if guildInfo.Treasury != 1000 {
			t.Errorf("expected treasury 1000, got %d", guildInfo.Treasury)
		}
	})

	t.Run("WithdrawTreasury", func(t *testing.T) {
		err := system.WithdrawTreasury(leader, 500)
		if err != nil {
			t.Fatalf("failed to withdraw: %v", err)
		}

		guildInfo, err := system.GetGuildInfo(guildID)
		if err != nil {
			t.Fatalf("failed to get guild info: %v", err)
		}

		if guildInfo.Treasury != 500 {
			t.Errorf("expected treasury 500, got %d", guildInfo.Treasury)
		}
	})

	t.Run("DepositNotInGuild", func(t *testing.T) {
		nonMember := world.CreateEntity()
		nonMember.AddComponent(&PlayerComponent{})

		err := system.DepositTreasury(nonMember, 100)
		if err == nil {
			t.Error("expected error depositing when not in guild")
		}
	})

	t.Run("WithdrawNotInGuild", func(t *testing.T) {
		nonMember := world.CreateEntity()
		nonMember.AddComponent(&PlayerComponent{})

		err := system.WithdrawTreasury(nonMember, 100)
		if err == nil {
			t.Error("expected error withdrawing when not in guild")
		}
	})
}

func TestGuildSystemPromote(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	leaderGuildComp, ok := leader.GetComponent("guild")
	if !ok {
		t.Fatalf("leader guild component not found")
	}
	guildID := leaderGuildComp.(*GuildComponent).GuildID

	// Add member
	member := world.CreateEntity()
	member.AddComponent(&PlayerComponent{})
	err = system.JoinGuild(member, guildID)
	if err != nil {
		t.Fatalf("failed to join guild: %v", err)
	}

	t.Run("PromoteMember", func(t *testing.T) {
		err := system.PromoteMember(leader, fmt.Sprintf("%d", member.ID))
		if err != nil {
			t.Fatalf("failed to promote member: %v", err)
		}

		// Check updated rank in component
		memberGuildComp, ok := member.GetComponent("guild")
		if !ok {
			t.Fatal("member guild component not found")
		}
		gc := memberGuildComp.(*GuildComponent)
		if gc.Rank == string(guild.RankRecruit) {
			t.Error("member rank not updated after promotion")
		}
	})

	t.Run("PromoteNonMember", func(t *testing.T) {
		nonMember := world.CreateEntity()
		nonMember.AddComponent(&PlayerComponent{})

		err := system.PromoteMember(leader, "player-nonexistent")
		if err == nil {
			t.Error("expected error promoting non-member")
		}
	})
}

func TestGuildSystemPersistence(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	t.Run("SaveLoadGuildData", func(t *testing.T) {
		// Save guild data
		data, err := system.SaveGuildData()
		if err != nil {
			t.Fatalf("failed to save guild data: %v", err)
		}

		if len(data) == 0 {
			t.Error("saved data is empty")
		}

		// Create new system and load data
		world2 := NewWorld()
		manager2 := guild.NewManager()
		system2 := NewGuildSystem(world2, manager2)

		err = system2.LoadGuildData(data)
		if err != nil {
			t.Fatalf("failed to load guild data: %v", err)
		}

		// Verify guild was loaded
		leaderGuildComp, ok := leader.GetComponent("guild")
		if !ok {
			t.Fatal("leader guild component not found")
		}
		gc := leaderGuildComp.(*GuildComponent)
		guildInfo, err := system2.GetGuildInfo(gc.GuildID)
		if err != nil {
			t.Fatalf("failed to get guild info after load: %v", err)
		}

		if guildInfo.LeaderID != fmt.Sprintf("%d", leader.ID) {
			t.Errorf("expected leader ID %d, got %s", leader.ID, guildInfo.LeaderID)
		}
	})
}

func TestGuildSystemSync(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)
	system.syncInterval = time.Millisecond * 100 // Fast sync for testing

	// Create guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	t.Run("AutomaticSync", func(t *testing.T) {
		// Trigger updates
		system.Update(nil, 0.1)

		// Wait for sync interval
		time.Sleep(time.Millisecond * 150)

		// Trigger sync
		system.Update(nil, 0.1)

		// Verify pending updates cleared (sync occurred)
		if len(system.pendingUpdates) > 0 {
			t.Errorf("expected pending updates to be cleared after sync, got %d", len(system.pendingUpdates))
		}
	})
}

func TestGuildSystemGetInfo(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create guild
	leader := world.CreateEntity()
	leader.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(leader, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	leaderGuildComp, ok := leader.GetComponent("guild")
	if !ok {
		t.Fatalf("leader guild component not found")
	}
	guildID := leaderGuildComp.(*GuildComponent).GuildID

	t.Run("GetGuildInfo", func(t *testing.T) {
		info, err := system.GetGuildInfo(guildID)
		if err != nil {
			t.Fatalf("failed to get guild info: %v", err)
		}

		if info.ID != guildID {
			t.Errorf("expected guild ID %s, got %s", guildID, info.ID)
		}
		if info.LeaderID != fmt.Sprintf("%d", leader.ID) {
			t.Errorf("expected leader ID %d, got %s", leader.ID, info.LeaderID)
		}
		if len(info.Members) != 1 {
			t.Errorf("expected 1 member, got %d", len(info.Members))
		}
	})

	t.Run("GetGuildInfoNonexistent", func(t *testing.T) {
		_, err := system.GetGuildInfo("guild-nonexistent")
		if err == nil {
			t.Error("expected error getting nonexistent guild info")
		}
	})
}

// MockFederationBroadcaster is a test implementation of FederationBroadcaster
type MockFederationBroadcaster struct {
	broadcasts []struct {
		guildID   string
		guildData []byte
	}
	shouldFail bool
}

func (m *MockFederationBroadcaster) BroadcastGuildUpdate(guildID string, guildData []byte) error {
	if m.shouldFail {
		return fmt.Errorf("mock broadcast failure")
	}
	m.broadcasts = append(m.broadcasts, struct {
		guildID   string
		guildData []byte
	}{guildID: guildID, guildData: guildData})
	return nil
}

func TestGuildSystemFederationSync(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create mock federation broadcaster
	mockFederation := &MockFederationBroadcaster{}
	system.SetFederation(mockFederation)

	// Create player and guild
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	err := system.CreateGuild(player, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	guildComp, _ := player.GetComponent("guild")
	gc := guildComp.(*GuildComponent)
	guildID := gc.GuildID

	t.Run("SyncTriggeredByCreate", func(t *testing.T) {
		// Force sync by calling Update after sufficient time
		system.lastSync = time.Now().Add(-6 * time.Second)
		system.Update(nil, 0.1)

		if len(mockFederation.broadcasts) == 0 {
			t.Error("expected guild update to be broadcast after create")
		}
		if mockFederation.broadcasts[0].guildID != guildID {
			t.Errorf("expected broadcast for guild %s, got %s",
				guildID, mockFederation.broadcasts[0].guildID)
		}
	})

	t.Run("SyncTriggeredByMemberAdd", func(t *testing.T) {
		// Reset broadcasts
		mockFederation.broadcasts = nil

		// Add a member
		newPlayer := world.CreateEntity()
		newPlayer.AddComponent(&PlayerComponent{})
		err := system.JoinGuild(newPlayer, guildID)
		if err != nil {
			t.Fatalf("failed to join guild: %v", err)
		}

		// Force sync
		system.lastSync = time.Now().Add(-6 * time.Second)
		system.Update(nil, 0.1)

		if len(mockFederation.broadcasts) == 0 {
			t.Error("expected guild update to be broadcast after member add")
		}
	})

	t.Run("SyncWithoutFederation", func(t *testing.T) {
		// Create system without federation
		systemNoFed := NewGuildSystem(world, guild.NewManager())
		player2 := world.CreateEntity()
		player2.AddComponent(&PlayerComponent{})

		err := systemNoFed.CreateGuild(player2, "sci-fi")
		if err != nil {
			t.Fatalf("failed to create guild without federation: %v", err)
		}

		// Should not panic
		systemNoFed.lastSync = time.Now().Add(-6 * time.Second)
		systemNoFed.Update(nil, 0.1)
	})

	t.Run("SyncFailureDoesNotBreakSystem", func(t *testing.T) {
		// Reset and set to fail
		mockFederation.broadcasts = nil
		mockFederation.shouldFail = true

		// Add another member
		player3 := world.CreateEntity()
		player3.AddComponent(&PlayerComponent{})
		err := system.JoinGuild(player3, guildID)
		if err != nil {
			t.Fatalf("failed to join guild: %v", err)
		}

		// Force sync - should log warning but not fail
		system.lastSync = time.Now().Add(-6 * time.Second)
		system.Update(nil, 0.1)

		// Guild should still be valid
		info, err := system.GetGuildInfo(guildID)
		if err != nil {
			t.Fatalf("guild should still be valid after sync failure: %v", err)
		}
		if len(info.Members) != 3 {
			t.Errorf("expected 3 members, got %d", len(info.Members))
		}
	})
}

func TestGuildSystemApplyUpdate(t *testing.T) {
	world := NewWorld()
	manager := guild.NewManager()
	system := NewGuildSystem(world, manager)

	// Create initial guild on "server A"
	player1 := world.CreateEntity()
	player1.AddComponent(&PlayerComponent{})
	err := system.CreateGuild(player1, "fantasy")
	if err != nil {
		t.Fatalf("failed to create guild: %v", err)
	}

	guildComp, _ := player1.GetComponent("guild")
	gc := guildComp.(*GuildComponent)
	guildID := gc.GuildID

	// Get initial guild data
	initialData, err := system.SaveGuildData()
	if err != nil {
		t.Fatalf("failed to save guild data: %v", err)
	}

	t.Run("ApplyUpdate", func(t *testing.T) {
		// Create "server B" that receives the update
		world2 := NewWorld()
		manager2 := guild.NewManager()
		system2 := NewGuildSystem(world2, manager2)

		// Apply update from server A to server B
		err := system2.ApplyGuildUpdate(guildID, initialData)
		if err != nil {
			t.Fatalf("failed to apply guild update: %v", err)
		}

		// Verify guild exists on server B
		info, err := system2.GetGuildInfo(guildID)
		if err != nil {
			t.Fatalf("guild should exist after update: %v", err)
		}
		if info.ID != guildID {
			t.Errorf("expected guild ID %s, got %s", guildID, info.ID)
		}
	})

	t.Run("ApplyUpdateWithLocalEntity", func(t *testing.T) {
		// Add member to original guild through JoinGuild (properly registers with manager)
		player2 := world.CreateEntity()
		player2.AddComponent(&PlayerComponent{})

		err := system.JoinGuild(player2, guildID)
		if err != nil {
			t.Fatalf("failed to add member to guild: %v", err)
		}

		// Promote member on "server A"
		err = system.PromoteMember(player1, fmt.Sprintf("%d", player2.ID))
		if err != nil {
			t.Fatalf("failed to promote member: %v", err)
		}

		// Save updated guild state
		updatedData, err := system.SaveGuildData()
		if err != nil {
			t.Fatalf("failed to save updated guild data: %v", err)
		}

		// Apply update back to same world (simulating receiving own broadcast)
		err = system.ApplyGuildUpdate(guildID, updatedData)
		if err != nil {
			t.Fatalf("failed to apply guild update: %v", err)
		}

		// Verify local entity rank updated
		guildComp2, _ := player2.GetComponent("guild")
		gc2 := guildComp2.(*GuildComponent)
		if gc2.Rank != string(guild.RankMember) {
			t.Errorf("expected rank %s, got %s", guild.RankMember, gc2.Rank)
		}
	})

	t.Run("ApplyInvalidData", func(t *testing.T) {
		world3 := NewWorld()
		manager3 := guild.NewManager()
		system3 := NewGuildSystem(world3, manager3)

		err := system3.ApplyGuildUpdate(guildID, []byte("invalid data"))
		if err == nil {
			t.Error("expected error applying invalid data")
		}
	})
}
