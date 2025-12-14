package engine

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPvPRatingComponent_Type(t *testing.T) {
	c := NewPvPRatingComponent("season-1")
	if c.Type() != "pvp_rating" {
		t.Errorf("Type() = %q, want %q", c.Type(), "pvp_rating")
	}
}

func TestNewPvPRatingComponent(t *testing.T) {
	seasonID := "season-2025-winter"
	c := NewPvPRatingComponent(seasonID)

	if c.Rating != 1000 {
		t.Errorf("Rating = %d, want 1000", c.Rating)
	}
	if c.PeakRating != 1000 {
		t.Errorf("PeakRating = %d, want 1000", c.PeakRating)
	}
	if c.RankTier != RankSilver {
		t.Errorf("RankTier = %q, want %q", c.RankTier, RankSilver)
	}
	if c.RankDivision != 3 {
		t.Errorf("RankDivision = %d, want 3", c.RankDivision)
	}
	if c.Wins != 0 {
		t.Errorf("Wins = %d, want 0", c.Wins)
	}
	if c.Losses != 0 {
		t.Errorf("Losses = %d, want 0", c.Losses)
	}
	if c.SeasonID != seasonID {
		t.Errorf("SeasonID = %q, want %q", c.SeasonID, seasonID)
	}
}

func TestPvPRatingComponent_GetWinRate(t *testing.T) {
	tests := []struct {
		name   string
		wins   int
		losses int
		want   float64
	}{
		{"no matches", 0, 0, 0.0},
		{"all wins", 10, 0, 100.0},
		{"all losses", 0, 10, 0.0},
		{"50-50", 5, 5, 50.0},
		{"75% wins", 3, 1, 75.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewPvPRatingComponent("test")
			c.Wins = tt.wins
			c.Losses = tt.losses
			got := c.GetWinRate()
			if got != tt.want {
				t.Errorf("GetWinRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPvPRatingComponent_GetTotalMatches(t *testing.T) {
	c := NewPvPRatingComponent("test")
	c.Wins = 7
	c.Losses = 3
	if got := c.GetTotalMatches(); got != 10 {
		t.Errorf("GetTotalMatches() = %d, want 10", got)
	}
}

func TestPvPRatingComponent_GetRankDisplay(t *testing.T) {
	tests := []struct {
		tier     RankTier
		division int
		want     string
	}{
		{RankBronze, 3, "Bronze III"},
		{RankBronze, 2, "Bronze II"},
		{RankBronze, 1, "Bronze I"},
		{RankSilver, 3, "Silver III"},
		{RankGold, 2, "Gold II"},
		{RankPlatinum, 1, "Platinum I"},
		{RankDiamond, 3, "Diamond III"},
		{RankMaster, 2, "Master II"},
		{RankLegend, 1, "Legend I"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			c := NewPvPRatingComponent("test")
			c.RankTier = tt.tier
			c.RankDivision = tt.division
			if got := c.GetRankDisplay(); got != tt.want {
				t.Errorf("GetRankDisplay() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTierFromRating(t *testing.T) {
	tests := []struct {
		rating int
		want   RankTier
	}{
		{0, RankBronze},
		{500, RankBronze},
		{999, RankBronze},
		{1000, RankSilver},
		{1100, RankSilver},
		{1199, RankSilver},
		{1200, RankGold},
		{1300, RankGold},
		{1400, RankPlatinum},
		{1500, RankPlatinum},
		{1600, RankDiamond},
		{1700, RankDiamond},
		{1800, RankMaster},
		{1900, RankMaster},
		{2000, RankLegend},
		{2500, RankLegend},
		{3000, RankLegend},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			if got := GetTierFromRating(tt.rating); got != tt.want {
				t.Errorf("GetTierFromRating(%d) = %q, want %q", tt.rating, got, tt.want)
			}
		})
	}
}

func TestGetDivisionFromRating(t *testing.T) {
	tests := []struct {
		name   string
		rating int
		want   int
	}{
		{"silver bottom", 1000, 3},
		{"silver mid", 1067, 2},
		{"silver top", 1134, 1},
		{"gold bottom", 1200, 3},
		{"gold mid", 1267, 2},
		{"gold top", 1334, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDivisionFromRating(tt.rating); got != tt.want {
				t.Errorf("GetDivisionFromRating(%d) = %d, want %d", tt.rating, got, tt.want)
			}
		})
	}
}

func TestPvPRatingComponent_Serialize(t *testing.T) {
	c := NewPvPRatingComponent("season-1")
	c.Rating = 1500
	c.PeakRating = 1600
	c.RankTier = RankPlatinum
	c.RankDivision = 2
	c.Wins = 50
	c.Losses = 30
	c.LastMatch = time.Date(2025, 12, 14, 12, 0, 0, 0, time.UTC)
	c.MatchStreak = 3

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	c2 := &PvPRatingComponent{}
	if err := c2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if c2.Rating != c.Rating {
		t.Errorf("Rating = %d, want %d", c2.Rating, c.Rating)
	}
	if c2.PeakRating != c.PeakRating {
		t.Errorf("PeakRating = %d, want %d", c2.PeakRating, c.PeakRating)
	}
	if c2.RankTier != c.RankTier {
		t.Errorf("RankTier = %q, want %q", c2.RankTier, c.RankTier)
	}
	if c2.RankDivision != c.RankDivision {
		t.Errorf("RankDivision = %d, want %d", c2.RankDivision, c.RankDivision)
	}
	if c2.Wins != c.Wins {
		t.Errorf("Wins = %d, want %d", c2.Wins, c.Wins)
	}
	if c2.Losses != c.Losses {
		t.Errorf("Losses = %d, want %d", c2.Losses, c.Losses)
	}
	if c2.SeasonID != c.SeasonID {
		t.Errorf("SeasonID = %q, want %q", c2.SeasonID, c.SeasonID)
	}
	if c2.MatchStreak != c.MatchStreak {
		t.Errorf("MatchStreak = %d, want %d", c2.MatchStreak, c.MatchStreak)
	}
}

func TestPvPRatingComponent_Deserialize_Invalid(t *testing.T) {
	c := &PvPRatingComponent{}
	err := c.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func TestPvPRatingComponent_IsPlacementComplete(t *testing.T) {
	tests := []struct {
		name   string
		wins   int
		losses int
		want   bool
	}{
		{"no matches", 0, 0, false},
		{"5 matches", 3, 2, false},
		{"9 matches", 5, 4, false},
		{"10 matches", 5, 5, true},
		{"15 matches", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewPvPRatingComponent("test")
			c.Wins = tt.wins
			c.Losses = tt.losses
			if got := c.IsPlacementComplete(); got != tt.want {
				t.Errorf("IsPlacementComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPvPRatingComponent_GetPlacementProgress(t *testing.T) {
	tests := []struct {
		name          string
		wins          int
		losses        int
		wantCompleted int
		wantTotal     int
	}{
		{"no matches", 0, 0, 0, 10},
		{"5 matches", 3, 2, 5, 10},
		{"10 matches", 5, 5, 10, 10},
		{"20 matches", 15, 5, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewPvPRatingComponent("test")
			c.Wins = tt.wins
			c.Losses = tt.losses
			completed, total := c.GetPlacementProgress()
			if completed != tt.wantCompleted {
				t.Errorf("completed = %d, want %d", completed, tt.wantCompleted)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestTierToName(t *testing.T) {
	tests := []struct {
		tier RankTier
		want string
	}{
		{RankBronze, "Bronze"},
		{RankSilver, "Silver"},
		{RankGold, "Gold"},
		{RankPlatinum, "Platinum"},
		{RankDiamond, "Diamond"},
		{RankMaster, "Master"},
		{RankLegend, "Legend"},
		{RankTier("unknown"), "Unranked"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tierToName(tt.tier); got != tt.want {
				t.Errorf("tierToName(%q) = %q, want %q", tt.tier, got, tt.want)
			}
		})
	}
}

func TestDivisionToRoman(t *testing.T) {
	tests := []struct {
		division int
		want     string
	}{
		{1, "I"},
		{2, "II"},
		{3, "III"},
		{0, "III"},
		{4, "III"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := divisionToRoman(tt.division); got != tt.want {
				t.Errorf("divisionToRoman(%d) = %q, want %q", tt.division, got, tt.want)
			}
		})
	}
}

func TestGetTierIndex(t *testing.T) {
	tests := []struct {
		tier RankTier
		want int
	}{
		{RankBronze, 0},
		{RankSilver, 1},
		{RankGold, 2},
		{RankPlatinum, 3},
		{RankDiamond, 4},
		{RankMaster, 5},
		{RankLegend, 6},
		{RankTier("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			if got := getTierIndex(tt.tier); got != tt.want {
				t.Errorf("getTierIndex(%q) = %d, want %d", tt.tier, got, tt.want)
			}
		})
	}
}

func TestPvPRatingComponent_JSON(t *testing.T) {
	c := NewPvPRatingComponent("test-season")
	c.Rating = 1350
	c.Wins = 25
	c.Losses = 15

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if parsed["rating"].(float64) != 1350 {
		t.Errorf("rating = %v, want 1350", parsed["rating"])
	}
	if parsed["wins"].(float64) != 25 {
		t.Errorf("wins = %v, want 25", parsed["wins"])
	}
}
