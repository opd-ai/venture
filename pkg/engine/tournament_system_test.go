package engine

import (
	"testing"
	"time"
)

func TestNewTournamentSystem(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTournamentSystem() returned nil")
	}
	if len(system.definitions) != 3 {
		t.Errorf("definitions length = %d, want 3", len(system.definitions))
	}
}

func TestTournamentSystem_CreateTournament(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test_tourney",
		Name:            "Test Tournament",
		Format:          TournamentFormatSingleElim,
		MinParticipants: 4,
		MaxParticipants: 16,
	}

	startTime := time.Now().Add(1 * time.Hour)
	instance := system.CreateTournament(def, startTime)

	if instance == nil {
		t.Fatal("CreateTournament() returned nil")
	}
	if instance.InstanceID == "" {
		t.Error("InstanceID should not be empty")
	}
	if instance.Phase != TournamentPhaseRegistration {
		t.Errorf("Phase = %q, want %q", instance.Phase, TournamentPhaseRegistration)
	}
	if len(system.scheduledTournaments) != 1 {
		t.Errorf("scheduled count = %d, want 1", len(system.scheduledTournaments))
	}
}

func TestTournamentSystem_RegisterPlayer(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	// Create tournament
	def := TournamentDefinition{
		ID:              "test",
		Name:            "Test",
		Format:          TournamentFormatSingleElim,
		MinParticipants: 2,
		MaxParticipants: 8,
	}
	startTime := time.Now().Add(1 * time.Hour)
	instance := system.CreateTournament(def, startTime)

	// Create player with tournament component
	player := world.CreateEntity()
	player.AddComponent(NewTournamentComponent())
	player.AddComponent(NewPvPRatingComponent("season-1"))

	// Register
	if !system.RegisterPlayer(player, instance.InstanceID) {
		t.Error("RegisterPlayer() should succeed")
	}

	if len(instance.Participants) != 1 {
		t.Errorf("Participants = %d, want 1", len(instance.Participants))
	}

	// Can't register twice
	if system.RegisterPlayer(player, instance.InstanceID) {
		t.Error("RegisterPlayer() should fail for duplicate")
	}
}

func TestTournamentSystem_RegisterPlayer_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		MinParticipants: 2,
		MaxParticipants: 8,
	}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	// Player without tournament component
	player := world.CreateEntity()

	if system.RegisterPlayer(player, instance.InstanceID) {
		t.Error("RegisterPlayer() should fail without component")
	}
}

func TestTournamentSystem_RegisterPlayer_RatingRequirement(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:                "ranked",
		MinParticipants:   2,
		MaxParticipants:   8,
		RatingRequirement: 1500,
	}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	// Player with low rating
	player := world.CreateEntity()
	player.AddComponent(NewTournamentComponent())
	pvp := NewPvPRatingComponent("season-1")
	pvp.Rating = 1200
	player.AddComponent(pvp)

	if system.RegisterPlayer(player, instance.InstanceID) {
		t.Error("RegisterPlayer() should fail for low rating")
	}

	// Player with high enough rating
	player2 := world.CreateEntity()
	player2.AddComponent(NewTournamentComponent())
	pvp2 := NewPvPRatingComponent("season-1")
	pvp2.Rating = 1600
	player2.AddComponent(pvp2)

	if !system.RegisterPlayer(player2, instance.InstanceID) {
		t.Error("RegisterPlayer() should succeed for high rating")
	}
}

func TestTournamentSystem_RegisterPlayer_Full(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "small",
		MinParticipants: 2,
		MaxParticipants: 2, // Very small
	}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	// Fill up
	for i := 0; i < 2; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewTournamentComponent())
		player.AddComponent(NewPvPRatingComponent("season-1"))
		system.RegisterPlayer(player, instance.InstanceID)
	}

	// Third player should fail
	player3 := world.CreateEntity()
	player3.AddComponent(NewTournamentComponent())
	player3.AddComponent(NewPvPRatingComponent("season-1"))

	if system.RegisterPlayer(player3, instance.InstanceID) {
		t.Error("RegisterPlayer() should fail when full")
	}
}

func TestTournamentSystem_RegisterPlayer_NotFound(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewTournamentComponent())

	if system.RegisterPlayer(player, "nonexistent") {
		t.Error("RegisterPlayer() should fail for nonexistent tournament")
	}
}

func TestTournamentSystem_UnregisterPlayer(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		MinParticipants: 2,
		MaxParticipants: 8,
	}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	player := world.CreateEntity()
	player.AddComponent(NewTournamentComponent())
	player.AddComponent(NewPvPRatingComponent("season-1"))
	system.RegisterPlayer(player, instance.InstanceID)

	// Unregister
	if !system.UnregisterPlayer(player, instance.InstanceID) {
		t.Error("UnregisterPlayer() should succeed")
	}

	if len(instance.Participants) != 0 {
		t.Errorf("Participants = %d, want 0", len(instance.Participants))
	}

	// Check component state
	tcComp, _ := player.GetComponent("tournament")
	tc := tcComp.(*TournamentComponent)
	if tc.IsInTournament() {
		t.Error("Player should not be in tournament after unregistering")
	}
}

func TestTournamentSystem_GetTournament(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{ID: "test"}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	// Find scheduled
	found := system.GetTournament(instance.InstanceID)
	if found == nil {
		t.Error("GetTournament() should find scheduled tournament")
	}

	// Not found
	if system.GetTournament("nonexistent") != nil {
		t.Error("GetTournament() should return nil for nonexistent")
	}
}

func TestTournamentSystem_GetTournamentCount(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	active, scheduled := system.GetTournamentCount()
	if active != 0 || scheduled != 0 {
		t.Error("Initial counts should be 0")
	}

	def := TournamentDefinition{ID: "test"}
	system.CreateTournament(def, time.Now().Add(1*time.Hour))

	active, scheduled = system.GetTournamentCount()
	if active != 0 {
		t.Errorf("active = %d, want 0", active)
	}
	if scheduled != 1 {
		t.Errorf("scheduled = %d, want 1", scheduled)
	}
}

func TestTournamentSystem_CancelTournament(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{ID: "test", MinParticipants: 2}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	// Register a player
	player := world.CreateEntity()
	player.AddComponent(NewTournamentComponent())
	player.AddComponent(NewPvPRatingComponent("season-1"))
	system.RegisterPlayer(player, instance.InstanceID)
	world.FlushPendingEntities()

	// Cancel
	if !system.CancelTournament(instance.InstanceID) {
		t.Error("CancelTournament() should succeed")
	}

	if instance.Phase != TournamentPhaseCancelled {
		t.Errorf("Phase = %q, want %q", instance.Phase, TournamentPhaseCancelled)
	}

	_, scheduled := system.GetTournamentCount()
	if scheduled != 0 {
		t.Errorf("scheduled = %d, want 0 after cancel", scheduled)
	}

	// Player should be released
	tc, _ := player.GetComponent("tournament")
	tournComp := tc.(*TournamentComponent)
	if tournComp.IsInTournament() {
		t.Error("Player should not be in tournament after cancel")
	}
}

func TestTournamentSystem_CancelTournament_NotFound(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	if system.CancelTournament("nonexistent") {
		t.Error("CancelTournament() should fail for nonexistent")
	}
}

func TestTournamentSystem_AddTournamentDefinition(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	initial := len(system.definitions)

	def := TournamentDefinition{
		ID:   "custom",
		Name: "Custom Tournament",
	}
	system.AddTournamentDefinition(def)

	if len(system.definitions) != initial+1 {
		t.Errorf("definitions = %d, want %d", len(system.definitions), initial+1)
	}
}

func TestTournamentSystem_Update_Empty(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	// Should not panic
	system.Update([]*Entity{}, 0.016)
}

func TestTournamentSystem_StartTournament(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		Name:            "Test",
		Format:          TournamentFormatSingleElim,
		MinParticipants: 2,
		MaxParticipants: 8,
		Seed:            12345,
	}
	startTime := time.Now().Add(-1 * time.Second) // In the past
	instance := system.CreateTournament(def, startTime)

	// Register enough players
	for i := 0; i < 4; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewTournamentComponent())
		pvp := NewPvPRatingComponent("season-1")
		pvp.Rating = 1000 + i*100 // Different ratings for seeding
		player.AddComponent(pvp)
		system.RegisterPlayer(player, instance.InstanceID)
	}
	world.FlushPendingEntities()

	// Process should start the tournament
	system.Update([]*Entity{}, 0.016)

	active, scheduled := system.GetTournamentCount()
	if active != 1 {
		t.Errorf("active = %d, want 1", active)
	}
	if scheduled != 0 {
		t.Errorf("scheduled = %d, want 0", scheduled)
	}

	// Verify tournament state
	tournament := system.GetTournament(instance.InstanceID)
	if tournament.Phase != TournamentPhaseInProgress {
		t.Errorf("Phase = %q, want %q", tournament.Phase, TournamentPhaseInProgress)
	}
	if len(tournament.Bracket) == 0 {
		t.Error("Bracket should be generated")
	}
}

func TestTournamentSystem_CancelInsufficientParticipants(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		MinParticipants: 4,
		MaxParticipants: 8,
	}
	startTime := time.Now().Add(-1 * time.Second) // In the past
	instance := system.CreateTournament(def, startTime)

	// Only register 2 players (need 4)
	for i := 0; i < 2; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewTournamentComponent())
		player.AddComponent(NewPvPRatingComponent("season-1"))
		system.RegisterPlayer(player, instance.InstanceID)
	}
	world.FlushPendingEntities()

	// Process should cancel
	system.Update([]*Entity{}, 0.016)

	if instance.Phase != TournamentPhaseCancelled {
		t.Errorf("Phase = %q, want %q", instance.Phase, TournamentPhaseCancelled)
	}
}

func TestTournamentSystem_RecordMatchResult(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		Format:          TournamentFormatSingleElim,
		MinParticipants: 2,
		MaxParticipants: 4,
		Seed:            12345,
	}
	startTime := time.Now().Add(-1 * time.Second)
	instance := system.CreateTournament(def, startTime)

	// Register 2 players
	players := make([]*Entity, 2)
	for i := 0; i < 2; i++ {
		players[i] = world.CreateEntity()
		players[i].AddComponent(NewTournamentComponent())
		players[i].AddComponent(NewPvPRatingComponent("season-1"))
		system.RegisterPlayer(players[i], instance.InstanceID)
	}
	world.FlushPendingEntities()

	// Start
	system.Update([]*Entity{}, 0.016)

	// Get match
	tournament := system.GetTournament(instance.InstanceID)
	if len(tournament.Bracket) == 0 {
		t.Fatal("Bracket should exist")
	}

	match := tournament.Bracket[0]
	winnerID := match.Player1ID
	loserID := match.Player2ID
	if winnerID == 0 {
		winnerID = match.Player2ID
		loserID = match.Player1ID
	}

	// Record result
	if !system.RecordMatchResult(instance.InstanceID, match.MatchID, winnerID, loserID) {
		t.Error("RecordMatchResult() should succeed")
	}

	// Verify match updated
	if !tournament.Bracket[0].Completed {
		t.Error("Match should be completed")
	}
	if tournament.Bracket[0].WinnerID != winnerID {
		t.Errorf("WinnerID = %d, want %d", tournament.Bracket[0].WinnerID, winnerID)
	}
}

func TestTournamentSystem_RecordMatchResult_NotFound(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	if system.RecordMatchResult("nonexistent", "match-1", 1, 2) {
		t.Error("RecordMatchResult() should fail for nonexistent tournament")
	}
}

func TestTournamentSystem_GetPlayerMatches(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		Format:          TournamentFormatSingleElim,
		MinParticipants: 2,
		MaxParticipants: 4,
		Seed:            12345,
	}
	startTime := time.Now().Add(-1 * time.Second)
	instance := system.CreateTournament(def, startTime)

	// Register players
	players := make([]*Entity, 2)
	for i := 0; i < 2; i++ {
		players[i] = world.CreateEntity()
		players[i].AddComponent(NewTournamentComponent())
		players[i].AddComponent(NewPvPRatingComponent("season-1"))
		system.RegisterPlayer(players[i], instance.InstanceID)
	}
	world.FlushPendingEntities()

	// Start
	system.Update([]*Entity{}, 0.016)

	// Get matches for first player
	matches := system.GetPlayerMatches(players[0].ID)
	if len(matches) == 0 {
		t.Error("Should have pending matches")
	}
}

func TestTournamentSystem_GetActiveTournaments(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	// Initially empty
	active := system.GetActiveTournaments()
	if len(active) != 0 {
		t.Errorf("active length = %d, want 0", len(active))
	}
}

func TestTournamentSystem_GetScheduledTournaments(t *testing.T) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{ID: "test"}
	system.CreateTournament(def, time.Now().Add(1*time.Hour))

	scheduled := system.GetScheduledTournaments()
	if len(scheduled) != 1 {
		t.Errorf("scheduled length = %d, want 1", len(scheduled))
	}
}

func BenchmarkTournamentSystem_RegisterPlayer(b *testing.B) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	def := TournamentDefinition{
		ID:              "test",
		MinParticipants: 2,
		MaxParticipants: 10000,
	}
	instance := system.CreateTournament(def, time.Now().Add(1*time.Hour))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewTournamentComponent())
		player.AddComponent(NewPvPRatingComponent("season-1"))
		system.RegisterPlayer(player, instance.InstanceID)
	}
}

func BenchmarkTournamentSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTournamentSystem(world, 12345)

	// Create some tournaments
	for i := 0; i < 5; i++ {
		def := TournamentDefinition{
			ID:              "test-" + string(rune('0'+i)),
			MinParticipants: 2,
			MaxParticipants: 16,
		}
		system.CreateTournament(def, time.Now().Add(1*time.Hour))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update([]*Entity{}, 0.016)
	}
}
