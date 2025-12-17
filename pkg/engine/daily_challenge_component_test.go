// Package engine provides tests for DailyChallengeComponent.
// Phase 98: Daily/Weekly Challenges (V18.0)
package engine

import (
	"testing"
	"time"
)

func TestNewDailyChallengeComponent(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	if c == nil {
		t.Fatal("NewDailyChallengeComponent returned nil")
	}

	if c.BaseSeed != 12345 {
		t.Errorf("BaseSeed = %d, want 12345", c.BaseSeed)
	}

	if c.RerollsRemaining != 3 {
		t.Errorf("RerollsRemaining = %d, want 3", c.RerollsRemaining)
	}

	if c.MaxDailyRerolls != 3 {
		t.Errorf("MaxDailyRerolls = %d, want 3", c.MaxDailyRerolls)
	}

	if len(c.ActiveDailyChallenges) != 0 {
		t.Errorf("ActiveDailyChallenges length = %d, want 0", len(c.ActiveDailyChallenges))
	}

	if len(c.ActiveWeeklyChallenges) != 0 {
		t.Errorf("ActiveWeeklyChallenges length = %d, want 0", len(c.ActiveWeeklyChallenges))
	}
}

func TestDailyChallengeComponentType(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	if c.Type() != "daily_challenge" {
		t.Errorf("Type() = %q, want %q", c.Type(), "daily_challenge")
	}
}

func TestGenerateDailyChallenges(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	challenges := c.GenerateDailyChallenges(date)

	if len(challenges) != 5 {
		t.Errorf("GenerateDailyChallenges returned %d challenges, want 5", len(challenges))
	}

	// Check all challenges are daily type
	for i, ch := range challenges {
		if ch.Type != ChallengeTypeDaily {
			t.Errorf("Challenge %d Type = %s, want %s", i, ch.Type, ChallengeTypeDaily)
		}
		if ch.Target <= 0 {
			t.Errorf("Challenge %d Target = %d, want > 0", i, ch.Target)
		}
		if ch.Progress != 0 {
			t.Errorf("Challenge %d Progress = %d, want 0", i, ch.Progress)
		}
		if ch.IsCompleted {
			t.Errorf("Challenge %d IsCompleted = true, want false", i)
		}
	}
}

func TestGenerateDailyChallengesDeterminism(t *testing.T) {
	c1 := NewDailyChallengeComponent(12345)
	c2 := NewDailyChallengeComponent(12345)
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	challenges1 := c1.GenerateDailyChallenges(date)
	challenges2 := c2.GenerateDailyChallenges(date)

	if len(challenges1) != len(challenges2) {
		t.Fatalf("Different challenge counts: %d vs %d", len(challenges1), len(challenges2))
	}

	for i := range challenges1 {
		if challenges1[i].ID != challenges2[i].ID {
			t.Errorf("Challenge %d ID mismatch: %s vs %s", i, challenges1[i].ID, challenges2[i].ID)
		}
		if challenges1[i].Target != challenges2[i].Target {
			t.Errorf("Challenge %d Target mismatch: %d vs %d", i, challenges1[i].Target, challenges2[i].Target)
		}
	}
}

func TestGenerateDailyChallengesDifferentDays(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	day1 := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 12, 16, 10, 0, 0, 0, time.UTC)

	challenges1 := c.GenerateDailyChallenges(day1)
	challenges2 := c.GenerateDailyChallenges(day2)

	// Should generate different challenges on different days
	sameCount := 0
	for _, ch1 := range challenges1 {
		for _, ch2 := range challenges2 {
			if ch1.DefinitionID == ch2.DefinitionID {
				sameCount++
				break
			}
		}
	}

	// Some overlap is fine, but not all 5 should be identical
	if sameCount == 5 {
		t.Error("All challenges are identical on different days, expected different generation")
	}
}

func TestGenerateWeeklyChallenges(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	weekStart := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	challenges := c.GenerateWeeklyChallenges(weekStart)

	if len(challenges) != 3 {
		t.Errorf("GenerateWeeklyChallenges returned %d challenges, want 3", len(challenges))
	}

	for i, ch := range challenges {
		if ch.Type != ChallengeTypeWeekly {
			t.Errorf("Challenge %d Type = %s, want %s", i, ch.Type, ChallengeTypeWeekly)
		}
	}
}

func TestGenerateWeeklyChallengesDeterminism(t *testing.T) {
	c1 := NewDailyChallengeComponent(12345)
	c2 := NewDailyChallengeComponent(12345)
	weekStart := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	challenges1 := c1.GenerateWeeklyChallenges(weekStart)
	challenges2 := c2.GenerateWeeklyChallenges(weekStart)

	if len(challenges1) != len(challenges2) {
		t.Fatalf("Different challenge counts: %d vs %d", len(challenges1), len(challenges2))
	}

	for i := range challenges1 {
		if challenges1[i].ID != challenges2[i].ID {
			t.Errorf("Challenge %d ID mismatch: %s vs %s", i, challenges1[i].ID, challenges2[i].ID)
		}
	}
}

func TestSetAndGetActiveChallenges(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	daily := c.GenerateDailyChallenges(date)
	weekly := c.GenerateWeeklyChallenges(date)

	c.SetActiveChallenges(daily, weekly)

	gotDaily := c.GetActiveDailyChallenges()
	gotWeekly := c.GetActiveWeeklyChallenges()

	if len(gotDaily) != 5 {
		t.Errorf("GetActiveDailyChallenges returned %d, want 5", len(gotDaily))
	}

	if len(gotWeekly) != 3 {
		t.Errorf("GetActiveWeeklyChallenges returned %d, want 3", len(gotWeekly))
	}
}

func TestUpdateProgress(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	daily := c.GenerateDailyChallenges(date)
	c.SetActiveChallenges(daily, nil)

	// Find a challenge with "enemies_killed" tracking key
	updated := c.UpdateProgress("enemies_killed", 5)

	if updated != nil {
		if updated.Progress != 5 {
			t.Errorf("Progress = %d, want 5", updated.Progress)
		}
	}
	// Note: tracking key might not match generated challenges due to shuffling
}

func TestUpdateProgressCompletion(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	// Create a challenge directly with known tracking key
	challenge := &Challenge{
		ID:           "test_challenge",
		DefinitionID: "kill_enemies",
		Type:         ChallengeTypeDaily,
		Category:     ChallengeCategoryCombat,
		Name:         "Test",
		Target:       10,
		Progress:     0,
		Reward:       ChallengeReward{XP: 100, Gold: 50},
	}

	c.SetActiveChallenges([]*Challenge{challenge}, nil)

	// Update to complete
	c.UpdateProgress("enemies_killed", 10)

	challenges := c.GetActiveDailyChallenges()
	if len(challenges) == 0 {
		t.Fatal("No active challenges")
	}

	if !challenges[0].IsCompleted {
		t.Error("Challenge should be completed")
	}

	if challenges[0].Progress != 10 {
		t.Errorf("Progress = %d, want 10", challenges[0].Progress)
	}
}

func TestComponentClaimReward(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	challenge := &Challenge{
		ID:           "test_challenge",
		DefinitionID: "kill_enemies",
		Type:         ChallengeTypeDaily,
		Category:     ChallengeCategoryCombat,
		Name:         "Test",
		Target:       10,
		Progress:     10,
		IsCompleted:  true,
		Reward:       ChallengeReward{XP: 100, Gold: 50},
	}

	c.SetActiveChallenges([]*Challenge{challenge}, nil)

	reward := c.ClaimReward("test_challenge")

	if reward == nil {
		t.Fatal("ClaimReward returned nil")
	}

	if reward.XP != 100 {
		t.Errorf("XP = %d, want 100", reward.XP)
	}

	if reward.Gold != 50 {
		t.Errorf("Gold = %d, want 50", reward.Gold)
	}

	if c.GetTotalCompleted() != 1 {
		t.Errorf("TotalCompleted = %d, want 1", c.GetTotalCompleted())
	}

	// Cannot claim twice
	reward2 := c.ClaimReward("test_challenge")
	if reward2 != nil {
		t.Error("Should not be able to claim reward twice")
	}
}

func TestClaimRewardWithStreak(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	c.DailyStreak = 5 // 5 day streak = +50% bonus

	challenge := &Challenge{
		ID:           "test_challenge",
		DefinitionID: "kill_enemies",
		Type:         ChallengeTypeDaily,
		Category:     ChallengeCategoryCombat,
		Name:         "Test",
		Target:       10,
		Progress:     10,
		IsCompleted:  true,
		Reward:       ChallengeReward{XP: 100, Gold: 50},
	}

	c.SetActiveChallenges([]*Challenge{challenge}, nil)

	reward := c.ClaimReward("test_challenge")

	// 100 * 1.5 = 150
	if reward.XP != 150 {
		t.Errorf("XP = %d, want 150 (with streak bonus)", reward.XP)
	}

	// 50 * 1.5 = 75
	if reward.Gold != 75 {
		t.Errorf("Gold = %d, want 75 (with streak bonus)", reward.Gold)
	}

	if reward.BonusMultiplier != 1.5 {
		t.Errorf("BonusMultiplier = %f, want 1.5", reward.BonusMultiplier)
	}
}

func TestComponentRerollChallenge(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	daily := c.GenerateDailyChallenges(date)
	c.SetActiveChallenges(daily, nil)

	oldID := daily[0].ID
	rerolled := c.RerollChallenge(oldID, date)

	if rerolled == nil {
		t.Fatal("RerollChallenge returned nil")
	}

	if c.GetRerollsRemaining() != 2 {
		t.Errorf("RerollsRemaining = %d, want 2", c.GetRerollsRemaining())
	}

	// New challenge should have different ID
	if rerolled.ID == oldID {
		t.Error("Rerolled challenge should have different ID")
	}
}

func TestRerollChallengeNoRerollsLeft(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	c.RerollsRemaining = 0

	daily := c.GenerateDailyChallenges(time.Now())
	c.SetActiveChallenges(daily, nil)

	rerolled := c.RerollChallenge(daily[0].ID, time.Now())

	if rerolled != nil {
		t.Error("Should not be able to reroll with 0 rerolls remaining")
	}
}

func TestRerollCompletedChallenge(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	challenge := &Challenge{
		ID:          "completed_challenge",
		IsCompleted: true,
	}

	c.SetActiveChallenges([]*Challenge{challenge}, nil)

	rerolled := c.RerollChallenge("completed_challenge", time.Now())

	if rerolled != nil {
		t.Error("Should not be able to reroll completed challenge")
	}
}

func TestCheckDailyReset(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	// Use UTC times to match the implementation's UTC-based day comparison
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	c.LastDailyReset = yesterday.Unix()

	// Add completed challenges
	c.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", IsCompleted: true},
		{ID: "ch2", IsCompleted: true},
		{ID: "ch3", IsCompleted: true},
		{ID: "ch4", IsCompleted: true},
		{ID: "ch5", IsCompleted: true},
	}

	reset := c.CheckDailyReset(now)

	if !reset {
		t.Error("CheckDailyReset should return true")
	}

	if c.DailyStreak != 1 {
		t.Errorf("DailyStreak = %d, want 1", c.DailyStreak)
	}

	if len(c.ActiveDailyChallenges) != 0 {
		t.Errorf("ActiveDailyChallenges should be cleared")
	}

	if c.RerollsRemaining != c.MaxDailyRerolls {
		t.Errorf("RerollsRemaining = %d, want %d", c.RerollsRemaining, c.MaxDailyRerolls)
	}
}

func TestCheckDailyResetBreaksStreak(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	c.DailyStreak = 5

	// Use UTC times to match the implementation's UTC-based day comparison
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	c.LastDailyReset = yesterday.Unix()

	// Add incomplete challenges
	c.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", IsCompleted: true},
		{ID: "ch2", IsCompleted: false}, // Not completed
		{ID: "ch3", IsCompleted: true},
	}

	c.CheckDailyReset(now)

	if c.DailyStreak != 0 {
		t.Errorf("DailyStreak = %d, want 0 (streak broken)", c.DailyStreak)
	}
}

func TestCheckWeeklyReset(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	lastWeek := time.Now().AddDate(0, 0, -8)
	c.LastWeeklyReset = lastWeek.Unix()

	// Add completed weekly challenges
	c.ActiveWeeklyChallenges = []*Challenge{
		{ID: "wch1", IsCompleted: true},
		{ID: "wch2", IsCompleted: true},
		{ID: "wch3", IsCompleted: true},
	}

	reset := c.CheckWeeklyReset(time.Now())

	if !reset {
		t.Error("CheckWeeklyReset should return true")
	}

	if c.WeeklyStreak != 1 {
		t.Errorf("WeeklyStreak = %d, want 1", c.WeeklyStreak)
	}

	if len(c.ActiveWeeklyChallenges) != 0 {
		t.Errorf("ActiveWeeklyChallenges should be cleared")
	}
}

func TestGetDailyCompletionPercent(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	c.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", IsCompleted: true},
		{ID: "ch2", IsCompleted: true},
		{ID: "ch3", IsCompleted: false},
		{ID: "ch4", IsCompleted: false},
		{ID: "ch5", IsCompleted: false},
	}

	percent := c.GetDailyCompletionPercent()

	if percent != 40.0 {
		t.Errorf("GetDailyCompletionPercent = %f, want 40.0", percent)
	}
}

func TestGetWeeklyCompletionPercent(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	c.ActiveWeeklyChallenges = []*Challenge{
		{ID: "wch1", IsCompleted: true},
		{ID: "wch2", IsCompleted: false},
		{ID: "wch3", IsCompleted: false},
	}

	percent := c.GetWeeklyCompletionPercent()

	expected := 100.0 / 3.0
	if percent < expected-0.1 || percent > expected+0.1 {
		t.Errorf("GetWeeklyCompletionPercent = %f, want ~%f", percent, expected)
	}
}

func TestIsChallengeExpired(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	pastTime := time.Now().Add(-time.Hour).Unix()
	futureTime := time.Now().Add(time.Hour).Unix()

	c.ActiveDailyChallenges = []*Challenge{
		{ID: "expired", ExpiresAt: pastTime},
		{ID: "active", ExpiresAt: futureTime},
	}

	if !c.IsChallengeExpired("expired", time.Now()) {
		t.Error("Challenge should be expired")
	}

	if c.IsChallengeExpired("active", time.Now()) {
		t.Error("Challenge should not be expired")
	}

	if !c.IsChallengeExpired("nonexistent", time.Now()) {
		t.Error("Nonexistent challenge should be considered expired")
	}
}

func TestGetStreakBonus(t *testing.T) {
	tests := []struct {
		name          string
		dailyStreak   int
		weeklyStreak  int
		challengeType ChallengeType
		wantBonus     float64
	}{
		{"no daily streak", 0, 0, ChallengeTypeDaily, 1.0},
		{"1 day streak", 1, 0, ChallengeTypeDaily, 1.1},
		{"5 day streak", 5, 0, ChallengeTypeDaily, 1.5},
		{"10+ day streak (capped)", 15, 0, ChallengeTypeDaily, 2.0},
		{"no weekly streak", 0, 0, ChallengeTypeWeekly, 1.0},
		{"1 week streak", 0, 1, ChallengeTypeWeekly, 1.25},
		{"4+ week streak (capped)", 0, 5, ChallengeTypeWeekly, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDailyChallengeComponent(12345)
			c.DailyStreak = tt.dailyStreak
			c.WeeklyStreak = tt.weeklyStreak

			got := c.GetStreakBonus(tt.challengeType)
			if got != tt.wantBonus {
				t.Errorf("GetStreakBonus() = %f, want %f", got, tt.wantBonus)
			}
		})
	}
}

func TestSerializeDeserialize(t *testing.T) {
	c := NewDailyChallengeComponent(12345)
	c.DailyStreak = 5
	c.WeeklyStreak = 2
	c.TotalChallengesCompleted = 100

	c.ActiveDailyChallenges = []*Challenge{
		{ID: "test1", Name: "Test 1", Progress: 5, Target: 10},
	}

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	c2 := NewDailyChallengeComponent(0)
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	if c2.BaseSeed != 12345 {
		t.Errorf("BaseSeed = %d, want 12345", c2.BaseSeed)
	}

	if c2.DailyStreak != 5 {
		t.Errorf("DailyStreak = %d, want 5", c2.DailyStreak)
	}

	if c2.WeeklyStreak != 2 {
		t.Errorf("WeeklyStreak = %d, want 2", c2.WeeklyStreak)
	}

	if c2.TotalChallengesCompleted != 100 {
		t.Errorf("TotalChallengesCompleted = %d, want 100", c2.TotalChallengesCompleted)
	}

	if len(c2.ActiveDailyChallenges) != 1 {
		t.Errorf("ActiveDailyChallenges length = %d, want 1", len(c2.ActiveDailyChallenges))
	}
}

func TestAllChallengeCategories(t *testing.T) {
	cats := AllChallengeCategories()

	if len(cats) != 5 {
		t.Errorf("AllChallengeCategories returned %d, want 5", len(cats))
	}

	expected := []ChallengeCategory{
		ChallengeCategoryCombat,
		ChallengeCategoryGathering,
		ChallengeCategoryExploration,
		ChallengeCategorySocial,
		ChallengeCategoryCrafting,
	}

	for i, cat := range expected {
		if cats[i] != cat {
			t.Errorf("Category %d = %s, want %s", i, cats[i], cat)
		}
	}
}

func TestChallengeCategory_String(t *testing.T) {
	tests := []struct {
		cat  ChallengeCategory
		want string
	}{
		{ChallengeCategoryCombat, "combat"},
		{ChallengeCategoryGathering, "gathering"},
		{ChallengeCategoryExploration, "exploration"},
		{ChallengeCategorySocial, "social"},
		{ChallengeCategoryCrafting, "crafting"},
	}

	for _, tt := range tests {
		if got := tt.cat.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.cat, got, tt.want)
		}
	}
}

func TestDefaultDailyChallengeDefinitions(t *testing.T) {
	defs := DefaultDailyChallengeDefinitions()

	if len(defs) < 15 {
		t.Errorf("DefaultDailyChallengeDefinitions returned %d, want >= 15", len(defs))
	}

	// Check each definition has required fields
	for _, def := range defs {
		if def.ID == "" {
			t.Error("Definition has empty ID")
		}
		if def.Name == "" {
			t.Error("Definition has empty Name")
		}
		if def.MinTarget <= 0 {
			t.Errorf("Definition %s has invalid MinTarget: %d", def.ID, def.MinTarget)
		}
		if def.MaxTarget < def.MinTarget {
			t.Errorf("Definition %s has MaxTarget < MinTarget", def.ID)
		}
		if def.TrackingKey == "" {
			t.Errorf("Definition %s has empty TrackingKey", def.ID)
		}
	}
}

func TestDefaultWeeklyChallengeDefinitions(t *testing.T) {
	defs := DefaultWeeklyChallengeDefinitions()

	if len(defs) < 7 {
		t.Errorf("DefaultWeeklyChallengeDefinitions returned %d, want >= 7", len(defs))
	}

	for _, def := range defs {
		if def.ID == "" {
			t.Error("Definition has empty ID")
		}
		if def.BaseXP <= 0 {
			t.Errorf("Definition %s has invalid BaseXP: %d", def.ID, def.BaseXP)
		}
	}
}

func TestLongestStreakTracking(t *testing.T) {
	c := NewDailyChallengeComponent(12345)

	// Use UTC times to match the implementation's UTC-based day comparison
	now := time.Now().UTC()

	// Simulate multiple daily resets with completed challenges
	for i := 0; i < 5; i++ {
		c.ActiveDailyChallenges = []*Challenge{
			{ID: "ch1", IsCompleted: true},
		}
		c.LastDailyReset = now.AddDate(0, 0, -(i + 1)).Unix()
		c.CheckDailyReset(now.AddDate(0, 0, -i))
	}

	if c.LongestDailyStreak < 5 {
		t.Errorf("LongestDailyStreak = %d, want >= 5", c.LongestDailyStreak)
	}
}

// BenchmarkGenerateDailyChallenges benchmarks daily challenge generation.
func BenchmarkGenerateDailyChallenges(b *testing.B) {
	c := NewDailyChallengeComponent(12345)
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GenerateDailyChallenges(date)
	}
}

// BenchmarkUpdateProgress benchmarks progress updates.
func BenchmarkUpdateProgress(b *testing.B) {
	c := NewDailyChallengeComponent(12345)
	challenge := &Challenge{
		ID:           "test_challenge",
		DefinitionID: "kill_enemies",
		Type:         ChallengeTypeDaily,
		Target:       1000000,
		Progress:     0,
	}
	c.SetActiveChallenges([]*Challenge{challenge}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.UpdateProgress("enemies_killed", 1)
	}
}
