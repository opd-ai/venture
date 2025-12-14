package engine

import (
	"math"
	"testing"
	"time"
)

// getPvPRating is a helper to get PvPRatingComponent from an entity in tests.
func getPvPRating(e *Entity) *PvPRatingComponent {
	comp, ok := e.GetComponent("pvp_rating")
	if !ok {
		return nil
	}
	rating, ok := comp.(*PvPRatingComponent)
	if !ok {
		return nil
	}
	return rating
}

func TestNewPvPRatingSystem(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	if system == nil {
		t.Fatal("NewPvPRatingSystem() returned nil")
	}
	if system.baseKFactor != 32.0 {
		t.Errorf("baseKFactor = %v, want 32.0", system.baseKFactor)
	}
	if system.decayDays != 14 {
		t.Errorf("decayDays = %d, want 14", system.decayDays)
	}
}

func TestCalculateELO(t *testing.T) {
	tests := []struct {
		name         string
		winnerRating int
		loserRating  int
		kFactor      float64
		wantWinner   int
		wantLoser    int
	}{
		{
			name:         "equal ratings",
			winnerRating: 1000,
			loserRating:  1000,
			kFactor:      32.0,
			wantWinner:   1016,
			wantLoser:    984,
		},
		{
			name:         "upset - lower beats higher",
			winnerRating: 1000,
			loserRating:  1200,
			kFactor:      32.0,
			wantWinner:   1024,
			wantLoser:    1176,
		},
		{
			name:         "expected - higher beats lower",
			winnerRating: 1200,
			loserRating:  1000,
			kFactor:      32.0,
			wantWinner:   1208,
			wantLoser:    992,
		},
		{
			name:         "large gap",
			winnerRating: 1000,
			loserRating:  1400,
			kFactor:      32.0,
			wantWinner:   1029,
			wantLoser:    1371,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWinner, gotLoser := CalculateELO(tt.winnerRating, tt.loserRating, tt.kFactor)

			// Allow for rounding differences of ±1
			if math.Abs(float64(gotWinner-tt.wantWinner)) > 1 {
				t.Errorf("winner rating = %d, want %d (±1)", gotWinner, tt.wantWinner)
			}
			if math.Abs(float64(gotLoser-tt.wantLoser)) > 1 {
				t.Errorf("loser rating = %d, want %d (±1)", gotLoser, tt.wantLoser)
			}
		})
	}
}

func TestCalculateELO_Determinism(t *testing.T) {
	// Run 100 times and verify same result
	for i := 0; i < 100; i++ {
		w, l := CalculateELO(1234, 1567, 32.0)
		w2, l2 := CalculateELO(1234, 1567, 32.0)
		if w != w2 || l != l2 {
			t.Errorf("Non-deterministic: got (%d,%d) and (%d,%d)", w, l, w2, l2)
		}
	}
}

func TestPvPRatingSystem_RecordMatchResult(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	winner := world.CreateEntity()
	loser := world.CreateEntity()
	winner.AddComponent(NewPvPRatingComponent("season-1"))
	loser.AddComponent(NewPvPRatingComponent("season-1"))

	matchTime := time.Now()
	newWinner, newLoser, err := system.RecordMatchResult(winner, loser, matchTime)
	if err != nil {
		t.Fatalf("RecordMatchResult() error = %v", err)
	}

	winnerRating := getPvPRating(winner)
	loserRating := getPvPRating(loser)

	// Starting at 1000 each, with K=64 for new players
	// Expected winner = 1000 + 64*0.5 = 1032
	// Expected loser = 1000 - 64*0.5 = 968
	if newWinner != winnerRating.Rating {
		t.Errorf("returned winner rating %d != stored rating %d", newWinner, winnerRating.Rating)
	}
	if newLoser != loserRating.Rating {
		t.Errorf("returned loser rating %d != stored rating %d", newLoser, loserRating.Rating)
	}

	if winnerRating.Wins != 1 {
		t.Errorf("winner wins = %d, want 1", winnerRating.Wins)
	}
	if loserRating.Losses != 1 {
		t.Errorf("loser losses = %d, want 1", loserRating.Losses)
	}
	if winnerRating.MatchStreak != 1 {
		t.Errorf("winner streak = %d, want 1", winnerRating.MatchStreak)
	}
	if loserRating.MatchStreak != -1 {
		t.Errorf("loser streak = %d, want -1", loserRating.MatchStreak)
	}
}

func TestPvPRatingSystem_RecordMatchResult_MissingComponent(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	winner := world.CreateEntity()
	loser := world.CreateEntity()
	// No PvP rating components

	_, _, err := system.RecordMatchResult(winner, loser, time.Now())
	if err != ErrMissingComponent {
		t.Errorf("expected ErrMissingComponent, got %v", err)
	}
}

func TestPvPRatingSystem_RecordMatchResult_Streak(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	winner := world.CreateEntity()
	loser := world.CreateEntity()
	winner.AddComponent(NewPvPRatingComponent("season-1"))
	loser.AddComponent(NewPvPRatingComponent("season-1"))

	// Record 3 wins
	for i := 0; i < 3; i++ {
		system.RecordMatchResult(winner, loser, time.Now())
	}

	winnerRating := getPvPRating(winner)
	if winnerRating.MatchStreak != 3 {
		t.Errorf("winner streak = %d, want 3", winnerRating.MatchStreak)
	}

	loserRating := getPvPRating(loser)
	if loserRating.MatchStreak != -3 {
		t.Errorf("loser streak = %d, want -3", loserRating.MatchStreak)
	}
}

func TestPvPRatingSystem_RecordMatchResult_PeakRating(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	winner := world.CreateEntity()
	loser := world.CreateEntity()
	winner.AddComponent(NewPvPRatingComponent("season-1"))
	loser.AddComponent(NewPvPRatingComponent("season-1"))

	system.RecordMatchResult(winner, loser, time.Now())

	winnerRating := getPvPRating(winner)
	if winnerRating.PeakRating < 1000 {
		t.Errorf("peak rating = %d, should be >= 1000", winnerRating.PeakRating)
	}
	if winnerRating.Rating != winnerRating.PeakRating {
		t.Errorf("rating %d != peak %d after only wins", winnerRating.Rating, winnerRating.PeakRating)
	}
}

func TestPvPRatingSystem_ResetSeasonRatings(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	player := world.CreateEntity()
	rating := NewPvPRatingComponent("season-1")
	rating.Rating = 1500
	rating.PeakRating = 1600
	rating.Wins = 50
	rating.Losses = 30
	player.AddComponent(rating)

	system.ResetSeasonRatings([]*Entity{player}, "season-2")

	updatedRating := getPvPRating(player)

	// Soft reset: (peak + 1000) / 2 = (1600 + 1000) / 2 = 1300
	expectedRating := 1300
	if updatedRating.Rating != expectedRating {
		t.Errorf("rating = %d, want %d", updatedRating.Rating, expectedRating)
	}
	if updatedRating.PeakRating != expectedRating {
		t.Errorf("peak = %d, want %d", updatedRating.PeakRating, expectedRating)
	}
	if updatedRating.Wins != 0 {
		t.Errorf("wins = %d, want 0", updatedRating.Wins)
	}
	if updatedRating.Losses != 0 {
		t.Errorf("losses = %d, want 0", updatedRating.Losses)
	}
	if updatedRating.SeasonID != "season-2" {
		t.Errorf("season = %q, want %q", updatedRating.SeasonID, "season-2")
	}
}

func TestPvPRatingSystem_GetPlayerRank(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	player := world.CreateEntity()
	rating := NewPvPRatingComponent("season-1")
	rating.RankTier = RankGold
	rating.RankDivision = 2
	player.AddComponent(rating)

	rank := system.GetPlayerRank(player)
	if rank != "Gold II" {
		t.Errorf("rank = %q, want %q", rank, "Gold II")
	}

	// Test unranked player
	unrankedPlayer := NewEntity()
	rank = system.GetPlayerRank(unrankedPlayer)
	if rank != "Unranked" {
		t.Errorf("unranked = %q, want %q", rank, "Unranked")
	}
}

func TestPvPRatingSystem_GetLeaderboard(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	// Create players with different ratings
	players := make([]*Entity, 5)
	ratings := []int{1500, 1200, 1800, 1000, 1400}
	expectedOrder := []int{2, 0, 4, 1, 3} // indices sorted by rating descending

	for i := 0; i < 5; i++ {
		players[i] = world.CreateEntity()
		rating := NewPvPRatingComponent("season-1")
		rating.Rating = ratings[i]
		rating.Wins = 10 // Complete placement
		players[i].AddComponent(rating)
	}

	leaderboard := system.GetLeaderboard(players, 3)

	if len(leaderboard) != 3 {
		t.Errorf("leaderboard length = %d, want 3", len(leaderboard))
	}

	for i, entity := range leaderboard {
		if entity != players[expectedOrder[i]] {
			t.Errorf("position %d: got player with rating %d, want rating %d",
				i,
				getPvPRating(entity).Rating,
				ratings[expectedOrder[i]])
		}
	}
}

func TestPvPRatingSystem_GetLeaderboard_ExcludesPlacement(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	player1 := world.CreateEntity()
	rating1 := NewPvPRatingComponent("season-1")
	rating1.Rating = 2000
	rating1.Wins = 5 // Not complete placement
	player1.AddComponent(rating1)

	player2 := world.CreateEntity()
	rating2 := NewPvPRatingComponent("season-1")
	rating2.Rating = 1500
	rating2.Wins = 10 // Complete placement
	player2.AddComponent(rating2)

	leaderboard := system.GetLeaderboard([]*Entity{player1, player2}, 10)

	if len(leaderboard) != 1 {
		t.Errorf("leaderboard length = %d, want 1", len(leaderboard))
	}
	if leaderboard[0] != player2 {
		t.Error("expected only player2 in leaderboard")
	}
}

func TestPvPRatingSystem_calculateKFactor(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	tests := []struct {
		name    string
		matches int
		rating  int
		want    float64
	}{
		{"new player", 5, 1000, 64.0},      // 2x base
		{"early player", 20, 1000, 48.0},   // 1.5x base
		{"regular player", 50, 1000, 32.0}, // base
		{"high rated", 50, 2100, 24.0},     // 0.75x base
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating := NewPvPRatingComponent("test")
			rating.Wins = tt.matches
			rating.Rating = tt.rating

			got := system.calculateKFactor(rating)
			if got != tt.want {
				t.Errorf("calculateKFactor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPvPRatingSystem_RatingDecay(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	player := world.CreateEntity()
	rating := NewPvPRatingComponent("season-1")
	rating.Rating = 1600
	rating.RankTier = RankDiamond
	rating.LastMatch = time.Now().Add(-21 * 24 * time.Hour) // 21 days ago
	player.AddComponent(rating)

	// Force decay check
	system.processRatingDecay([]*Entity{player}, time.Now())

	updatedRating := player.GetComponent("pvp_rating").(*PvPRatingComponent)
	if updatedRating.Rating >= 1600 {
		t.Errorf("rating should have decayed from 1600, got %d", updatedRating.Rating)
	}
}

func TestPvPRatingSystem_RatingDecay_NoDecayForSilver(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	player := world.CreateEntity()
	rating := NewPvPRatingComponent("season-1")
	rating.Rating = 1050
	rating.RankTier = RankSilver
	rating.LastMatch = time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	player.AddComponent(rating)

	system.processRatingDecay([]*Entity{player}, time.Now())

	updatedRating := player.GetComponent("pvp_rating").(*PvPRatingComponent)
	if updatedRating.Rating != 1050 {
		t.Errorf("silver player should not decay, got %d", updatedRating.Rating)
	}
}

func TestPvPRatingSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	// Update should not panic with empty entities
	system.Update([]*Entity{}, 0.016)

	player := world.CreateEntity()
	player.AddComponent(NewPvPRatingComponent("season-1"))
	system.Update([]*Entity{player}, 0.016)
}

func TestPvPRatingSystem_updateRankFromRating(t *testing.T) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	tests := []struct {
		rating       int
		wantTier     RankTier
		wantDivision int
	}{
		{500, RankBronze, 2},
		{1000, RankSilver, 3},
		{1150, RankSilver, 1},
		{1200, RankGold, 3},
		{1350, RankGold, 1},
		{1400, RankPlatinum, 3},
		{1600, RankDiamond, 3},
		{1800, RankMaster, 3},
		{2000, RankLegend, 3},
		{2200, RankLegend, 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.wantTier), func(t *testing.T) {
			comp := NewPvPRatingComponent("test")
			comp.Rating = tt.rating
			system.updateRankFromRating(comp)

			if comp.RankTier != tt.wantTier {
				t.Errorf("tier = %q, want %q", comp.RankTier, tt.wantTier)
			}
		})
	}
}

func BenchmarkCalculateELO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateELO(1234, 1567, 32.0)
	}
}

func BenchmarkRecordMatchResult(b *testing.B) {
	world := NewWorld()
	system := NewPvPRatingSystem(world)

	winner := world.CreateEntity()
	loser := world.CreateEntity()
	winner.AddComponent(NewPvPRatingComponent("season-1"))
	loser.AddComponent(NewPvPRatingComponent("season-1"))

	matchTime := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.RecordMatchResult(winner, loser, matchTime)
	}
}
