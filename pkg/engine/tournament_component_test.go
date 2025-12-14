package engine

import (
	"testing"
	"time"
)

func TestTournamentComponent_Type(t *testing.T) {
	c := NewTournamentComponent()
	if c.Type() != "tournament" {
		t.Errorf("Type() = %q, want %q", c.Type(), "tournament")
	}
}

func TestNewTournamentComponent(t *testing.T) {
	c := NewTournamentComponent()

	if c.CurrentTournamentID != "" {
		t.Error("CurrentTournamentID should be empty")
	}
	if c.MaxHistorySize != 50 {
		t.Errorf("MaxHistorySize = %d, want 50", c.MaxHistorySize)
	}
	if c.TotalTournamentsEntered != 0 {
		t.Error("TotalTournamentsEntered should be 0")
	}
}

func TestTournamentComponent_EnterTournament(t *testing.T) {
	c := NewTournamentComponent()

	// Should succeed when not in tournament
	if !c.EnterTournament("tourney-1", 5) {
		t.Error("EnterTournament() should succeed when not in tournament")
	}

	if c.CurrentTournamentID != "tourney-1" {
		t.Errorf("CurrentTournamentID = %q, want %q", c.CurrentTournamentID, "tourney-1")
	}
	if c.Seed != 5 {
		t.Errorf("Seed = %d, want 5", c.Seed)
	}
	if c.TotalTournamentsEntered != 1 {
		t.Errorf("TotalTournamentsEntered = %d, want 1", c.TotalTournamentsEntered)
	}

	// Should fail when already in tournament
	if c.EnterTournament("tourney-2", 3) {
		t.Error("EnterTournament() should fail when already in tournament")
	}
}

func TestTournamentComponent_LeaveTournament(t *testing.T) {
	c := NewTournamentComponent()

	// Should fail when not in tournament
	if c.LeaveTournament() {
		t.Error("LeaveTournament() should fail when not in tournament")
	}

	// Enter first
	c.EnterTournament("tourney-1", 5)

	// Should succeed
	if !c.LeaveTournament() {
		t.Error("LeaveTournament() should succeed")
	}

	if c.CurrentTournamentID != "" {
		t.Error("CurrentTournamentID should be empty after leaving")
	}
	if c.Seed != 0 {
		t.Error("Seed should be 0 after leaving")
	}
}

func TestTournamentComponent_RecordWin(t *testing.T) {
	c := NewTournamentComponent()
	c.EnterTournament("tourney-1", 1)

	c.RecordWin()

	if c.MatchesWon != 1 {
		t.Errorf("MatchesWon = %d, want 1", c.MatchesWon)
	}
	if c.TotalTournamentMatches != 1 {
		t.Errorf("TotalTournamentMatches = %d, want 1", c.TotalTournamentMatches)
	}
}

func TestTournamentComponent_RecordLoss_SingleElim(t *testing.T) {
	c := NewTournamentComponent()
	c.EnterTournament("tourney-1", 1)

	c.RecordLoss(false) // Single elimination

	if c.LossesInTournament != 1 {
		t.Errorf("LossesInTournament = %d, want 1", c.LossesInTournament)
	}
	if !c.Eliminated {
		t.Error("Should be eliminated after one loss in single elim")
	}
}

func TestTournamentComponent_RecordLoss_DoubleElim(t *testing.T) {
	c := NewTournamentComponent()
	c.EnterTournament("tourney-1", 1)

	c.RecordLoss(true) // Double elimination

	if c.LossesInTournament != 1 {
		t.Errorf("LossesInTournament = %d, want 1", c.LossesInTournament)
	}
	if c.Eliminated {
		t.Error("Should NOT be eliminated after one loss in double elim")
	}

	c.RecordLoss(true) // Second loss

	if !c.Eliminated {
		t.Error("Should be eliminated after two losses in double elim")
	}
}

func TestTournamentComponent_CompleteTournament(t *testing.T) {
	c := NewTournamentComponent()
	c.MaxHistorySize = 3

	c.EnterTournament("tourney-1", 1)
	c.RecordWin()
	c.RecordWin()

	result := TournamentResult{
		TournamentID:      "tourney-1",
		TournamentName:    "Test Tournament",
		Placement:         1,
		TotalParticipants: 8,
		MatchesWon:        2,
		MatchesLost:       0,
		CompletedAt:       time.Now(),
	}

	c.CompleteTournament(result)

	if c.CurrentTournamentID != "" {
		t.Error("CurrentTournamentID should be empty after completion")
	}
	if c.TotalTournamentWins != 1 {
		t.Errorf("TotalTournamentWins = %d, want 1", c.TotalTournamentWins)
	}
	if len(c.TournamentHistory) != 1 {
		t.Errorf("TournamentHistory length = %d, want 1", len(c.TournamentHistory))
	}

	// Test history limit
	for i := 0; i < 5; i++ {
		c.EnterTournament("tourney-"+string(rune('2'+i)), 1)
		result.TournamentID = "tourney-" + string(rune('2'+i))
		result.Placement = 2 // Not a win
		c.CompleteTournament(result)
	}

	if len(c.TournamentHistory) != c.MaxHistorySize {
		t.Errorf("TournamentHistory length = %d, want %d", len(c.TournamentHistory), c.MaxHistorySize)
	}
}

func TestTournamentComponent_IsInTournament(t *testing.T) {
	c := NewTournamentComponent()

	if c.IsInTournament() {
		t.Error("IsInTournament() should be false initially")
	}

	c.EnterTournament("tourney-1", 1)

	if !c.IsInTournament() {
		t.Error("IsInTournament() should be true after entering")
	}
}

func TestTournamentComponent_IsEliminated(t *testing.T) {
	c := NewTournamentComponent()

	if c.IsEliminated() {
		t.Error("IsEliminated() should be false initially")
	}

	c.Eliminated = true

	if !c.IsEliminated() {
		t.Error("IsEliminated() should be true when eliminated")
	}
}

func TestTournamentComponent_Spectating(t *testing.T) {
	c := NewTournamentComponent()

	// Should succeed when not in tournament
	if !c.StartSpectating("tourney-1") {
		t.Error("StartSpectating() should succeed")
	}

	if !c.IsSpectating {
		t.Error("IsSpectating should be true")
	}
	if c.SpectatingTournamentID != "tourney-1" {
		t.Errorf("SpectatingTournamentID = %q, want %q", c.SpectatingTournamentID, "tourney-1")
	}

	c.StopSpectating()

	if c.IsSpectating {
		t.Error("IsSpectating should be false after stopping")
	}

	// Should fail when in tournament
	c.EnterTournament("tourney-2", 1)
	if c.StartSpectating("tourney-3") {
		t.Error("StartSpectating() should fail when in tournament")
	}
}

func TestTournamentComponent_GetRecentTournaments(t *testing.T) {
	c := NewTournamentComponent()

	// Empty history
	if c.GetRecentTournaments(5) != nil {
		t.Error("GetRecentTournaments() should return nil for empty history")
	}
	if c.GetRecentTournaments(0) != nil {
		t.Error("GetRecentTournaments(0) should return nil")
	}
	if c.GetRecentTournaments(-1) != nil {
		t.Error("GetRecentTournaments(-1) should return nil")
	}

	// Add some history
	for i := 0; i < 10; i++ {
		c.TournamentHistory = append(c.TournamentHistory, TournamentResult{
			TournamentID: "t-" + string(rune('0'+i)),
			Placement:    i + 1,
		})
	}

	recent := c.GetRecentTournaments(3)
	if len(recent) != 3 {
		t.Errorf("GetRecentTournaments(3) length = %d, want 3", len(recent))
	}
	if recent[0].TournamentID != "t-7" {
		t.Errorf("First recent = %q, want %q", recent[0].TournamentID, "t-7")
	}

	// More than available
	recent = c.GetRecentTournaments(20)
	if len(recent) != 10 {
		t.Errorf("GetRecentTournaments(20) length = %d, want 10", len(recent))
	}
}

func TestTournamentComponent_GetWinRate(t *testing.T) {
	c := NewTournamentComponent()

	// No tournaments
	if c.GetWinRate() != 0.0 {
		t.Error("GetWinRate() should be 0 with no tournaments")
	}

	// Some wins
	c.TotalTournamentsEntered = 10
	c.TotalTournamentWins = 4

	rate := c.GetWinRate()
	if rate != 40.0 {
		t.Errorf("GetWinRate() = %f, want 40.0", rate)
	}
}

func TestTournamentComponent_GetAveragePlacement(t *testing.T) {
	c := NewTournamentComponent()

	// No history
	if c.GetAveragePlacement() != 0.0 {
		t.Error("GetAveragePlacement() should be 0 with no history")
	}

	// Add placements
	c.TournamentHistory = []TournamentResult{
		{Placement: 1},
		{Placement: 2},
		{Placement: 3},
	}

	avg := c.GetAveragePlacement()
	if avg != 2.0 {
		t.Errorf("GetAveragePlacement() = %f, want 2.0", avg)
	}
}

func TestTournamentComponent_Serialize(t *testing.T) {
	c := NewTournamentComponent()
	c.TotalTournamentsEntered = 25
	c.TotalTournamentWins = 10
	c.TournamentHistory = append(c.TournamentHistory, TournamentResult{
		TournamentID: "t-1",
		Placement:    1,
	})

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	c2 := &TournamentComponent{}
	if err := c2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if c2.TotalTournamentsEntered != 25 {
		t.Errorf("TotalTournamentsEntered = %d, want 25", c2.TotalTournamentsEntered)
	}
	if c2.TotalTournamentWins != 10 {
		t.Errorf("TotalTournamentWins = %d, want 10", c2.TotalTournamentWins)
	}
	if len(c2.TournamentHistory) != 1 {
		t.Errorf("TournamentHistory length = %d, want 1", len(c2.TournamentHistory))
	}
}

func TestTournamentComponent_Deserialize_Invalid(t *testing.T) {
	c := &TournamentComponent{}
	err := c.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func TestGenerateSingleElimBracket(t *testing.T) {
	tests := []struct {
		name         string
		participants int
		wantMatches  int
	}{
		{"2 players", 2, 1},
		{"4 players", 4, 3},
		{"8 players", 8, 7},
		{"16 players", 16, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			participants := make([]uint64, tt.participants)
			for i := range participants {
				participants[i] = uint64(i + 1)
			}

			bracket := GenerateSingleElimBracket(12345, participants)

			if len(bracket) != tt.wantMatches {
				t.Errorf("bracket length = %d, want %d", len(bracket), tt.wantMatches)
			}
		})
	}
}

func TestGenerateSingleElimBracket_Deterministic(t *testing.T) {
	participants := []uint64{1, 2, 3, 4, 5, 6, 7, 8}
	seed := int64(54321)

	bracket1 := GenerateSingleElimBracket(seed, participants)
	bracket2 := GenerateSingleElimBracket(seed, participants)

	if len(bracket1) != len(bracket2) {
		t.Fatal("Brackets should have same length")
	}

	for i := range bracket1 {
		if bracket1[i].MatchID != bracket2[i].MatchID {
			t.Error("Brackets should be identical for same seed")
		}
		if bracket1[i].Player1ID != bracket2[i].Player1ID {
			t.Error("Player assignments should be identical for same seed")
		}
	}
}

func TestGenerateSingleElimBracket_EdgeCases(t *testing.T) {
	// Less than 2 players
	bracket := GenerateSingleElimBracket(12345, []uint64{1})
	if bracket != nil {
		t.Error("Should return nil for less than 2 players")
	}

	bracket = GenerateSingleElimBracket(12345, []uint64{})
	if bracket != nil {
		t.Error("Should return nil for empty participants")
	}
}

func TestGenerateSingleElimBracket_WithByes(t *testing.T) {
	// 5 players (needs byes to make 8)
	participants := []uint64{1, 2, 3, 4, 5}
	bracket := GenerateSingleElimBracket(12345, participants)

	// Should have 7 matches (for bracket of 8)
	if len(bracket) != 7 {
		t.Errorf("bracket length = %d, want 7", len(bracket))
	}

	// Count first round matches with byes (where one player is 0)
	byeMatches := 0
	for _, match := range bracket {
		if match.Round == 1 && match.Completed {
			// A bye is when one player is 0 and the match auto-completed
			if match.Player1ID == 0 || match.Player2ID == 0 {
				byeMatches++
			}
		}
	}

	// With 5 players and 8 slots, there are 3 empty slots (byes)
	// But bye matches are only when exactly one player is 0
	// In first round of 4 matches: 5 players fill 5 slots, 3 slots empty
	// Expected bye matches: 3 (each empty slot creates one bye)
	// However, shuffling may pair empty slots together, reducing bye matches
	// Just check that there are some byes
	if byeMatches < 1 {
		t.Error("Expected at least 1 bye match")
	}
}

func TestGenerateDoubleElimBracket(t *testing.T) {
	participants := []uint64{1, 2, 3, 4, 5, 6, 7, 8}

	winners, losers := GenerateDoubleElimBracket(12345, participants)

	if winners == nil {
		t.Fatal("Winners bracket should not be nil")
	}
	if losers == nil {
		t.Fatal("Losers bracket should not be nil")
	}
	if len(winners) != 7 { // Standard 8-player single elim
		t.Errorf("winners length = %d, want 7", len(winners))
	}
	if len(losers) == 0 {
		t.Error("Losers bracket should have matches")
	}
}

func TestGenerateDoubleElimBracket_EdgeCase(t *testing.T) {
	// Single player
	winners, losers := GenerateDoubleElimBracket(12345, []uint64{1})
	if winners != nil || losers != nil {
		t.Error("Should return nil for less than 2 players")
	}
}

func TestCalculateTotalRounds(t *testing.T) {
	tests := []struct {
		participants int
		wantRounds   int
	}{
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{8, 3},
		{16, 4},
		{32, 5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			rounds := CalculateTotalRounds(tt.participants)
			if rounds != tt.wantRounds {
				t.Errorf("CalculateTotalRounds(%d) = %d, want %d", tt.participants, rounds, tt.wantRounds)
			}
		})
	}
}

func TestGetDefaultTournamentDefinitions(t *testing.T) {
	defs := GetDefaultTournamentDefinitions(12345)

	if len(defs) != 3 {
		t.Errorf("definitions length = %d, want 3", len(defs))
	}

	// Verify daily tournament
	daily := defs[0]
	if daily.ID != "daily_1v1" {
		t.Errorf("daily ID = %q, want %q", daily.ID, "daily_1v1")
	}
	if daily.Format != TournamentFormatSingleElim {
		t.Errorf("daily format = %q, want %q", daily.Format, TournamentFormatSingleElim)
	}
	if daily.Frequency != TournamentFrequencyDaily {
		t.Errorf("daily frequency = %q, want %q", daily.Frequency, TournamentFrequencyDaily)
	}

	// Verify weekly tournament
	weekly := defs[1]
	if weekly.Format != TournamentFormatDoubleElim {
		t.Errorf("weekly format = %q, want %q", weekly.Format, TournamentFormatDoubleElim)
	}
	if weekly.RatingRequirement != 1000 {
		t.Errorf("weekly rating req = %d, want 1000", weekly.RatingRequirement)
	}
}

func TestTournamentFormat_Constants(t *testing.T) {
	formats := []TournamentFormat{
		TournamentFormatSingleElim,
		TournamentFormatDoubleElim,
		TournamentFormatRoundRobin,
	}

	for _, f := range formats {
		if f == "" {
			t.Error("Format should not be empty")
		}
	}
}

func TestTournamentPhase_Constants(t *testing.T) {
	phases := []TournamentPhase{
		TournamentPhaseRegistration,
		TournamentPhaseSeeding,
		TournamentPhaseInProgress,
		TournamentPhaseComplete,
		TournamentPhaseCancelled,
	}

	for _, p := range phases {
		if p == "" {
			t.Error("Phase should not be empty")
		}
	}
}

func TestTournamentFrequency_Constants(t *testing.T) {
	frequencies := []TournamentFrequency{
		TournamentFrequencyDaily,
		TournamentFrequencyWeekly,
		TournamentFrequencyMonthly,
		TournamentFrequencySpecial,
	}

	for _, f := range frequencies {
		if f == "" {
			t.Error("Frequency should not be empty")
		}
	}
}

func BenchmarkGenerateSingleElimBracket(b *testing.B) {
	participants := make([]uint64, 64)
	for i := range participants {
		participants[i] = uint64(i + 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSingleElimBracket(int64(i), participants)
	}
}

func BenchmarkTournamentComponent_RecordWin(b *testing.B) {
	c := NewTournamentComponent()
	c.EnterTournament("tourney-1", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.RecordWin()
	}
}
