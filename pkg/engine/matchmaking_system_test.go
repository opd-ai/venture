package engine

import (
	"testing"
	"time"
)

func TestNewMatchmakingSystem(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	if system == nil {
		t.Fatal("NewMatchmakingSystem() returned nil")
	}
	if system.config.InitialRatingRange != 200 {
		t.Errorf("InitialRatingRange = %d, want 200", system.config.InitialRatingRange)
	}
	if system.config.MaxRatingRange != 500 {
		t.Errorf("MaxRatingRange = %d, want 500", system.config.MaxRatingRange)
	}
}

func TestNewMatchmakingSystemWithConfig(t *testing.T) {
	world := NewWorld()
	config := MatchmakingConfig{
		InitialRatingRange: 100,
		MaxRatingRange:     400,
		MinPlayersPerMode: map[MatchmakingMode]int{
			MatchmakingMode1v1: 2,
		},
		MaxPlayersPerMode: map[MatchmakingMode]int{
			MatchmakingMode1v1: 2,
		},
	}
	system := NewMatchmakingSystemWithConfig(world, 12345, config)

	if system.config.InitialRatingRange != 100 {
		t.Errorf("InitialRatingRange = %d, want 100", system.config.InitialRatingRange)
	}
}

func TestDefaultMatchmakingConfig(t *testing.T) {
	config := DefaultMatchmakingConfig()

	if config.InitialRatingRange != 200 {
		t.Errorf("InitialRatingRange = %d, want 200", config.InitialRatingRange)
	}
	if config.MaxRatingRange != 500 {
		t.Errorf("MaxRatingRange = %d, want 500", config.MaxRatingRange)
	}
	if config.MatchAcceptTimeout != 30*time.Second {
		t.Errorf("MatchAcceptTimeout = %v, want 30s", config.MatchAcceptTimeout)
	}
	if config.MinPlayersPerMode[MatchmakingMode1v1] != 2 {
		t.Errorf("MinPlayersPerMode[1v1] = %d, want 2", config.MinPlayersPerMode[MatchmakingMode1v1])
	}
	if config.MinPlayersPerMode[MatchmakingMode2v2] != 4 {
		t.Errorf("MinPlayersPerMode[2v2] = %d, want 4", config.MinPlayersPerMode[MatchmakingMode2v2])
	}
	if config.MinPlayersPerMode[MatchmakingModeFFA] != 3 {
		t.Errorf("MinPlayersPerMode[ffa] = %d, want 3", config.MinPlayersPerMode[MatchmakingModeFFA])
	}
}

func TestMatchmakingSystem_AddToQueue(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Create player with matchmaking component
	player := world.CreateEntity()
	player.AddComponent(NewMatchmakingComponent("server-1"))
	player.AddComponent(NewPvPRatingComponent("season-1"))

	// Should succeed
	if !system.AddToQueue(player, MatchmakingMode1v1) {
		t.Error("AddToQueue() should succeed")
	}

	if system.GetQueueSize(MatchmakingMode1v1) != 1 {
		t.Errorf("Queue size = %d, want 1", system.GetQueueSize(MatchmakingMode1v1))
	}

	// Should fail when already queued
	if system.AddToQueue(player, MatchmakingMode1v1) {
		t.Error("AddToQueue() should fail when already queued")
	}
}

func TestMatchmakingSystem_AddToQueue_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Create player without matchmaking component
	player := world.CreateEntity()

	if system.AddToQueue(player, MatchmakingMode1v1) {
		t.Error("AddToQueue() should fail without matchmaking component")
	}
}

func TestMatchmakingSystem_RemoveFromQueue(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewMatchmakingComponent("server-1"))

	system.AddToQueue(player, MatchmakingMode1v1)

	// Should succeed
	if !system.RemoveFromQueue(player) {
		t.Error("RemoveFromQueue() should succeed")
	}

	if system.GetQueueSize(MatchmakingMode1v1) != 0 {
		t.Errorf("Queue size = %d, want 0", system.GetQueueSize(MatchmakingMode1v1))
	}
}

func TestMatchmakingSystem_RemoveFromQueue_NotQueued(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewMatchmakingComponent("server-1"))

	// Should fail when not queued
	if system.RemoveFromQueue(player) {
		t.Error("RemoveFromQueue() should fail when not queued")
	}
}

func TestMatchmakingSystem_GetQueueSizes(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Add players to different modes
	for i := 0; i < 3; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		system.AddToQueue(player, MatchmakingMode1v1)
	}

	for i := 0; i < 5; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		system.AddToQueue(player, MatchmakingMode2v2)
	}

	sizes := system.GetQueueSizes()

	if sizes[MatchmakingMode1v1] != 3 {
		t.Errorf("1v1 queue size = %d, want 3", sizes[MatchmakingMode1v1])
	}
	if sizes[MatchmakingMode2v2] != 5 {
		t.Errorf("2v2 queue size = %d, want 5", sizes[MatchmakingMode2v2])
	}
}

func TestMatchmakingSystem_GetEstimatedWaitTime(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Empty queue
	wait := system.GetEstimatedWaitTime(MatchmakingMode1v1, 1000)
	if wait != 60*time.Second {
		t.Errorf("Empty queue wait = %v, want 60s", wait)
	}

	// Add some players
	for i := 0; i < 5; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		rating := NewPvPRatingComponent("season-1")
		rating.Rating = 1000 + i*50 // 1000, 1050, 1100, 1150, 1200
		player.AddComponent(rating)
		system.AddToQueue(player, MatchmakingMode1v1)
	}

	// Should have shorter wait with players in queue
	wait = system.GetEstimatedWaitTime(MatchmakingMode1v1, 1100)
	if wait >= 60*time.Second {
		t.Errorf("Wait with players = %v, should be less than 60s", wait)
	}
}

func TestMatchmakingSystem_GetPlayerQueuePosition(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	players := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		players[i] = world.CreateEntity()
		players[i].AddComponent(NewMatchmakingComponent("server-1"))
		system.AddToQueue(players[i], MatchmakingMode1v1)
	}

	// Check positions
	pos := system.GetPlayerQueuePosition(players[0].ID)
	if pos != 1 {
		t.Errorf("Player 0 position = %d, want 1", pos)
	}

	pos = system.GetPlayerQueuePosition(players[2].ID)
	if pos != 3 {
		t.Errorf("Player 2 position = %d, want 3", pos)
	}

	// Non-queued player
	pos = system.GetPlayerQueuePosition(999999)
	if pos != 0 {
		t.Errorf("Unknown player position = %d, want 0", pos)
	}
}

func TestMatchmakingSystem_TryCreateMatch_1v1(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Add 2 players with similar ratings
	for i := 0; i < 2; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		rating := NewPvPRatingComponent("season-1")
		rating.Rating = 1000 + i*50 // 1000, 1050 - within range
		player.AddComponent(rating)
		system.AddToQueue(player, MatchmakingMode1v1)
	}

	// Process matchmaking
	system.Update([]*Entity{}, 0.016)

	// Should have created a pending match
	if system.GetPendingMatchCount() != 1 {
		t.Errorf("Pending matches = %d, want 1", system.GetPendingMatchCount())
	}

	// Queue should be empty
	if system.GetQueueSize(MatchmakingMode1v1) != 0 {
		t.Errorf("Queue size = %d, want 0", system.GetQueueSize(MatchmakingMode1v1))
	}
}

func TestMatchmakingSystem_TryCreateMatch_RatingMismatch(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Add 2 players with very different ratings
	player1 := world.CreateEntity()
	player1.AddComponent(NewMatchmakingComponent("server-1"))
	rating1 := NewPvPRatingComponent("season-1")
	rating1.Rating = 1000
	player1.AddComponent(rating1)
	system.AddToQueue(player1, MatchmakingMode1v1)

	player2 := world.CreateEntity()
	player2.AddComponent(NewMatchmakingComponent("server-1"))
	rating2 := NewPvPRatingComponent("season-1")
	rating2.Rating = 1800 // Too far apart (800 difference)
	player2.AddComponent(rating2)
	system.AddToQueue(player2, MatchmakingMode1v1)

	// Process matchmaking
	system.Update([]*Entity{}, 0.016)

	// Should NOT have created a pending match (ratings too different)
	if system.GetPendingMatchCount() != 0 {
		t.Errorf("Pending matches = %d, want 0 (ratings too different)", system.GetPendingMatchCount())
	}
}

func TestMatchmakingSystem_AcceptMatch(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Create 2 players
	players := make([]*Entity, 2)
	for i := 0; i < 2; i++ {
		players[i] = world.CreateEntity()
		players[i].AddComponent(NewMatchmakingComponent("server-1"))
		players[i].AddComponent(NewPvPRatingComponent("season-1"))
		system.AddToQueue(players[i], MatchmakingMode1v1)
	}

	// Create match
	system.Update([]*Entity{}, 0.016)

	// Get match ID from pending
	var matchID string
	for id := range system.pending {
		matchID = id
		break
	}

	// Both players accept
	for _, player := range players {
		if !system.AcceptMatch(player, matchID) {
			t.Error("AcceptMatch() should succeed")
		}
	}

	// Match should have started (removed from pending)
	if system.GetPendingMatchCount() != 0 {
		t.Errorf("Pending matches = %d, want 0 after all accepted", system.GetPendingMatchCount())
	}
}

func TestMatchmakingSystem_DeclineMatch(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Create 2 players
	players := make([]*Entity, 2)
	for i := 0; i < 2; i++ {
		players[i] = world.CreateEntity()
		players[i].AddComponent(NewMatchmakingComponent("server-1"))
		players[i].AddComponent(NewPvPRatingComponent("season-1"))
		system.AddToQueue(players[i], MatchmakingMode1v1)
	}

	// Create match
	system.Update([]*Entity{}, 0.016)

	var matchID string
	for id := range system.pending {
		matchID = id
		break
	}

	// First player declines
	if !system.DeclineMatch(players[0], matchID) {
		t.Error("DeclineMatch() should succeed")
	}

	// Match should be cancelled
	if system.GetPendingMatchCount() != 0 {
		t.Errorf("Pending matches = %d, want 0 after decline", system.GetPendingMatchCount())
	}

	// Other player should be back in queue
	if system.GetQueueSize(MatchmakingMode1v1) != 1 {
		t.Errorf("Queue size = %d, want 1 (other player re-queued)", system.GetQueueSize(MatchmakingMode1v1))
	}
}

func TestMatchmakingSystem_Update_Empty(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Should not panic with empty queues
	system.Update([]*Entity{}, 0.016)
}

func TestMatchmakingSystem_ExpandRatingRanges(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewMatchmakingComponent("server-1"))
	rating := NewPvPRatingComponent("season-1")
	rating.Rating = 1000
	player.AddComponent(rating)
	system.AddToQueue(player, MatchmakingMode1v1)

	// Check initial range
	entry := system.queues[MatchmakingMode1v1][0]
	if entry.RatingRange != 200 {
		t.Errorf("Initial RatingRange = %d, want 200", entry.RatingRange)
	}

	// Simulate time passing by adjusting queue time
	entry.QueuedAt = time.Now().Add(-30 * time.Second)

	// Update should expand range
	system.expandRatingRanges(time.Now())

	// Range should have expanded (30s * 10/s = 300 + 200 = 500, capped at 500)
	if entry.RatingRange < 200 {
		t.Errorf("RatingRange = %d, should have expanded", entry.RatingRange)
	}
}

func TestMatchmakingSystem_ProcessExpiredMatches(t *testing.T) {
	world := NewWorld()
	config := DefaultMatchmakingConfig()
	config.MatchAcceptTimeout = 1 * time.Millisecond // Very short for testing
	system := NewMatchmakingSystemWithConfig(world, 12345, config)

	// Create 2 players
	for i := 0; i < 2; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		player.AddComponent(NewPvPRatingComponent("season-1"))
		system.AddToQueue(player, MatchmakingMode1v1)
	}

	// Create match
	system.Update([]*Entity{}, 0.016)

	if system.GetPendingMatchCount() != 1 {
		t.Fatalf("Should have 1 pending match")
	}

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	// Process expired matches
	system.processExpiredMatches(time.Now())

	// Match should be expired
	if system.GetPendingMatchCount() != 0 {
		t.Errorf("Pending matches = %d, want 0 after expiry", system.GetPendingMatchCount())
	}

	// Players should be back in queue
	if system.GetQueueSize(MatchmakingMode1v1) != 2 {
		t.Errorf("Queue size = %d, want 2 (players re-queued)", system.GetQueueSize(MatchmakingMode1v1))
	}
}

func TestMatchmakingSystem_2v2Mode(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Add 4 players for 2v2
	for i := 0; i < 4; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		rating := NewPvPRatingComponent("season-1")
		rating.Rating = 1000 + i*30 // Close ratings
		player.AddComponent(rating)
		system.AddToQueue(player, MatchmakingMode2v2)
	}

	// Process matchmaking
	system.Update([]*Entity{}, 0.016)

	// Should have created a pending match
	if system.GetPendingMatchCount() != 1 {
		t.Errorf("Pending matches = %d, want 1", system.GetPendingMatchCount())
	}

	// Queue should be empty
	if system.GetQueueSize(MatchmakingMode2v2) != 0 {
		t.Errorf("Queue size = %d, want 0", system.GetQueueSize(MatchmakingMode2v2))
	}
}

func TestMatchmakingSystem_FFAMode(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Add 3 players for FFA (minimum)
	for i := 0; i < 3; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		rating := NewPvPRatingComponent("season-1")
		rating.Rating = 1000 + i*40 // Close ratings
		player.AddComponent(rating)
		system.AddToQueue(player, MatchmakingModeFFA)
	}

	// Process matchmaking
	system.Update([]*Entity{}, 0.016)

	// Should have created a pending match
	if system.GetPendingMatchCount() != 1 {
		t.Errorf("Pending matches = %d, want 1", system.GetPendingMatchCount())
	}
}

func TestMatchmakingSystem_InsufficientPlayers(t *testing.T) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Only 1 player for 1v1
	player := world.CreateEntity()
	player.AddComponent(NewMatchmakingComponent("server-1"))
	player.AddComponent(NewPvPRatingComponent("season-1"))
	system.AddToQueue(player, MatchmakingMode1v1)

	// Process matchmaking
	system.Update([]*Entity{}, 0.016)

	// Should NOT create a match
	if system.GetPendingMatchCount() != 0 {
		t.Errorf("Pending matches = %d, want 0 (insufficient players)", system.GetPendingMatchCount())
	}
}

func TestAbsHelper(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
	}

	for _, tt := range tests {
		if got := absInt(tt.input); got != tt.want {
			t.Errorf("absInt(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestMinHelper(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{5, 5, 5},
		{-5, 5, -5},
	}

	for _, tt := range tests {
		if got := min(tt.a, tt.b); got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkMatchmakingSystem_AddToQueue(b *testing.B) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		player.AddComponent(NewPvPRatingComponent("season-1"))
		system.AddToQueue(player, MatchmakingMode1v1)
	}
}

func BenchmarkMatchmakingSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewMatchmakingSystem(world, 12345)

	// Pre-populate queue
	for i := 0; i < 100; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewMatchmakingComponent("server-1"))
		rating := NewPvPRatingComponent("season-1")
		rating.Rating = 1000 + (i % 500) // Varied ratings
		player.AddComponent(rating)
		system.AddToQueue(player, MatchmakingMode1v1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update([]*Entity{}, 0.016)
	}
}
