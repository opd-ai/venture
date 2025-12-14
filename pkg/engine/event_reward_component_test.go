package engine

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewEventRewardComponent(t *testing.T) {
	comp := NewEventRewardComponent()

	if comp == nil {
		t.Fatal("NewEventRewardComponent returned nil")
	}

	if comp.EarnedRewards == nil {
		t.Error("EarnedRewards should be initialized")
	}
	if comp.EventCurrency == nil {
		t.Error("EventCurrency should be initialized")
	}
	if comp.EarnedTitles == nil {
		t.Error("EarnedTitles should be initialized")
	}
	if comp.EarnedEffects == nil {
		t.Error("EarnedEffects should be initialized")
	}
	if comp.AchievementProgress == nil {
		t.Error("AchievementProgress should be initialized")
	}
	if comp.CompletedAchievements == nil {
		t.Error("CompletedAchievements should be initialized")
	}
}

func TestEventRewardComponentType(t *testing.T) {
	comp := NewEventRewardComponent()

	if comp.Type() != "event_reward" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "event_reward")
	}
}

func TestEventRewardComponent_Currency(t *testing.T) {
	tests := []struct {
		name      string
		eventID   string
		add       int
		wantTotal int
	}{
		{"add positive", "spring_festival", 100, 100},
		{"add zero", "spring_festival", 0, 0},
		{"add negative", "spring_festival", -10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewEventRewardComponent()
			comp.AddCurrency(tt.eventID, tt.add)

			got := comp.GetCurrency(tt.eventID)
			if got != tt.wantTotal {
				t.Errorf("GetCurrency() = %d, want %d", got, tt.wantTotal)
			}
		})
	}
}

func TestEventRewardComponent_CurrencyAccumulation(t *testing.T) {
	comp := NewEventRewardComponent()
	eventID := "summer_solstice"

	comp.AddCurrency(eventID, 50)
	comp.AddCurrency(eventID, 30)
	comp.AddCurrency(eventID, 20)

	if comp.GetCurrency(eventID) != 100 {
		t.Errorf("GetCurrency() = %d, want 100", comp.GetCurrency(eventID))
	}

	if comp.TotalCurrencyEarned != 100 {
		t.Errorf("TotalCurrencyEarned = %d, want 100", comp.TotalCurrencyEarned)
	}
}

func TestEventRewardComponent_SpendCurrency(t *testing.T) {
	tests := []struct {
		name         string
		initialFunds int
		spend        int
		wantSuccess  bool
		wantAfter    int
	}{
		{"spend valid", 100, 30, true, 70},
		{"spend exact", 100, 100, true, 0},
		{"spend insufficient", 50, 100, false, 50},
		{"spend zero", 100, 0, false, 100},
		{"spend negative", 100, -10, false, 100},
	}

	eventID := "autumn_harvest"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewEventRewardComponent()
			comp.AddCurrency(eventID, tt.initialFunds)

			got := comp.SpendCurrency(eventID, tt.spend)
			if got != tt.wantSuccess {
				t.Errorf("SpendCurrency() = %v, want %v", got, tt.wantSuccess)
			}
			if comp.GetCurrency(eventID) != tt.wantAfter {
				t.Errorf("Currency after = %d, want %d", comp.GetCurrency(eventID), tt.wantAfter)
			}
		})
	}
}

func TestEventRewardComponent_AddReward(t *testing.T) {
	comp := NewEventRewardComponent()

	reward := EventReward{
		ID:          "spring_item_1",
		EventID:     "spring_festival",
		Type:        EventRewardItem,
		Name:        "Spring Blossom Crown",
		Description: "A crown of spring blossoms",
		Value:       1,
		Rarity:      "rare",
	}

	comp.AddReward(reward)

	if len(comp.EarnedRewards) != 1 {
		t.Fatalf("Expected 1 reward, got %d", len(comp.EarnedRewards))
	}

	if comp.EarnedRewards[0].ID != reward.ID {
		t.Errorf("Reward ID = %q, want %q", comp.EarnedRewards[0].ID, reward.ID)
	}
}

func TestEventRewardComponent_ClaimReward(t *testing.T) {
	comp := NewEventRewardComponent()

	reward := EventReward{
		ID:      "test_reward",
		EventID: "test_event",
		Type:    EventRewardCurrency,
		Name:    "Test Currency",
		Value:   50,
		Claimed: false,
	}
	comp.AddReward(reward)

	// Claim the reward
	if !comp.ClaimReward("test_reward") {
		t.Error("ClaimReward should return true for valid unclaimed reward")
	}

	// Verify it's claimed
	if !comp.EarnedRewards[0].Claimed {
		t.Error("Reward should be marked as claimed")
	}

	// Try to claim again
	if comp.ClaimReward("test_reward") {
		t.Error("ClaimReward should return false for already claimed reward")
	}

	// Try to claim non-existent
	if comp.ClaimReward("nonexistent") {
		t.Error("ClaimReward should return false for non-existent reward")
	}
}

func TestEventRewardComponent_GetUnclaimedRewards(t *testing.T) {
	comp := NewEventRewardComponent()

	// Add multiple rewards
	for i := 0; i < 3; i++ {
		comp.AddReward(EventReward{
			ID:      "reward_" + string(rune('0'+i)),
			EventID: "test_event",
			Type:    EventRewardCurrency,
			Value:   10,
		})
	}

	// Claim one
	comp.ClaimReward("reward_1")

	unclaimed := comp.GetUnclaimedRewards()
	if len(unclaimed) != 2 {
		t.Errorf("GetUnclaimedRewards() returned %d, want 2", len(unclaimed))
	}
}

func TestEventRewardComponent_GetRewardsForEvent(t *testing.T) {
	comp := NewEventRewardComponent()

	// Add rewards for different events
	comp.AddReward(EventReward{ID: "r1", EventID: "event_a"})
	comp.AddReward(EventReward{ID: "r2", EventID: "event_a"})
	comp.AddReward(EventReward{ID: "r3", EventID: "event_b"})

	eventARewards := comp.GetRewardsForEvent("event_a")
	if len(eventARewards) != 2 {
		t.Errorf("GetRewardsForEvent(event_a) = %d, want 2", len(eventARewards))
	}

	eventBRewards := comp.GetRewardsForEvent("event_b")
	if len(eventBRewards) != 1 {
		t.Errorf("GetRewardsForEvent(event_b) = %d, want 1", len(eventBRewards))
	}
}

func TestEventRewardComponent_Titles(t *testing.T) {
	comp := NewEventRewardComponent()

	title := EventCosmeticTitle{
		ID:          "spring_herald",
		EventID:     "spring_festival",
		DisplayName: "Herald of Spring",
		Rarity:      "epic",
	}

	// Add title
	comp.AddTitle(title)
	if len(comp.EarnedTitles) != 1 {
		t.Fatalf("Expected 1 title, got %d", len(comp.EarnedTitles))
	}

	// Try to add duplicate
	comp.AddTitle(title)
	if len(comp.EarnedTitles) != 1 {
		t.Error("Should not add duplicate title")
	}

	// Set active title
	if !comp.SetActiveTitle("spring_herald") {
		t.Error("SetActiveTitle should return true for owned title")
	}
	if comp.ActiveTitle != "spring_herald" {
		t.Errorf("ActiveTitle = %q, want %q", comp.ActiveTitle, "spring_herald")
	}

	// Try to set non-owned title
	if comp.SetActiveTitle("unowned_title") {
		t.Error("SetActiveTitle should return false for unowned title")
	}

	// Get active title
	active := comp.GetActiveTitle()
	if active == nil {
		t.Fatal("GetActiveTitle returned nil")
	}
	if active.ID != "spring_herald" {
		t.Errorf("GetActiveTitle().ID = %q, want %q", active.ID, "spring_herald")
	}
}

func TestEventRewardComponent_Effects(t *testing.T) {
	comp := NewEventRewardComponent()

	effect := EventVisualEffect{
		ID:         "spring_aura",
		EventID:    "spring_festival",
		Name:       "Petal Aura",
		EffectType: "aura",
		ColorHex:   "#FF69B4",
		Intensity:  0.8,
	}

	// Add effect
	comp.AddEffect(effect)
	if len(comp.EarnedEffects) != 1 {
		t.Fatalf("Expected 1 effect, got %d", len(comp.EarnedEffects))
	}

	// Try to add duplicate
	comp.AddEffect(effect)
	if len(comp.EarnedEffects) != 1 {
		t.Error("Should not add duplicate effect")
	}

	// Set active effect
	if !comp.SetActiveEffect("spring_aura") {
		t.Error("SetActiveEffect should return true for owned effect")
	}
	if comp.ActiveEffect != "spring_aura" {
		t.Errorf("ActiveEffect = %q, want %q", comp.ActiveEffect, "spring_aura")
	}

	// Try to set non-owned effect
	if comp.SetActiveEffect("unowned_effect") {
		t.Error("SetActiveEffect should return false for unowned effect")
	}

	// Get active effect
	active := comp.GetActiveEffect()
	if active == nil {
		t.Fatal("GetActiveEffect returned nil")
	}
	if active.ID != "spring_aura" {
		t.Errorf("GetActiveEffect().ID = %q, want %q", active.ID, "spring_aura")
	}
}

func TestEventRewardComponent_AchievementProgress(t *testing.T) {
	comp := NewEventRewardComponent()

	// Update progress - not yet complete
	completed := comp.UpdateAchievementProgress("test_ach", 5, 10)
	if completed {
		t.Error("UpdateAchievementProgress should return false when not complete")
	}

	progress := comp.GetAchievementProgress("test_ach")
	if progress == nil {
		t.Fatal("GetAchievementProgress returned nil")
	}
	if progress.CurrentProgress != 5 {
		t.Errorf("CurrentProgress = %d, want 5", progress.CurrentProgress)
	}

	// Update to completion
	completed = comp.UpdateAchievementProgress("test_ach", 10, 10)
	if !completed {
		t.Error("UpdateAchievementProgress should return true when newly complete")
	}
	if !progress.Completed {
		t.Error("Achievement should be marked as completed")
	}

	// Update again when already complete
	completed = comp.UpdateAchievementProgress("test_ach", 15, 10)
	if completed {
		t.Error("UpdateAchievementProgress should return false when already complete")
	}
}

func TestEventRewardComponent_IncrementAchievementProgress(t *testing.T) {
	comp := NewEventRewardComponent()

	// Increment progress
	comp.IncrementAchievementProgress("inc_ach", 3, 10)
	comp.IncrementAchievementProgress("inc_ach", 4, 10)

	progress := comp.GetAchievementProgress("inc_ach")
	if progress.CurrentProgress != 7 {
		t.Errorf("CurrentProgress = %d, want 7", progress.CurrentProgress)
	}

	// Increment to completion
	completed := comp.IncrementAchievementProgress("inc_ach", 5, 10)
	if !completed {
		t.Error("Should return true when newly completed")
	}

	// Check capped at required
	if progress.CurrentProgress != 10 {
		t.Errorf("CurrentProgress = %d, want 10 (should cap at required)", progress.CurrentProgress)
	}
}

func TestEventRewardComponent_CompleteAchievement(t *testing.T) {
	comp := NewEventRewardComponent()

	// Complete achievement
	if !comp.CompleteAchievement("test_ach") {
		t.Error("CompleteAchievement should return true for new achievement")
	}

	// Verify completion
	if !comp.HasAchievement("test_ach") {
		t.Error("HasAchievement should return true for completed achievement")
	}

	// Try to complete again
	if comp.CompleteAchievement("test_ach") {
		t.Error("CompleteAchievement should return false for already completed")
	}

	// Check non-existent
	if comp.HasAchievement("nonexistent") {
		t.Error("HasAchievement should return false for non-existent")
	}
}

func TestEventRewardComponent_RecordStats(t *testing.T) {
	comp := NewEventRewardComponent()

	comp.RecordEventParticipation()
	comp.RecordEventParticipation()
	comp.RecordQuestCompletion()

	if comp.TotalEventsParticipated != 2 {
		t.Errorf("TotalEventsParticipated = %d, want 2", comp.TotalEventsParticipated)
	}
	if comp.TotalQuestsCompleted != 1 {
		t.Errorf("TotalQuestsCompleted = %d, want 1", comp.TotalQuestsCompleted)
	}
}

func TestEventRewardComponent_Serialization(t *testing.T) {
	comp := NewEventRewardComponent()

	// Add various data
	comp.AddCurrency("spring_festival", 150)
	comp.AddReward(EventReward{
		ID:      "test_reward",
		EventID: "spring_festival",
		Type:    EventRewardItem,
		Name:    "Test Item",
		Value:   1,
		Rarity:  "rare",
	})
	comp.AddTitle(EventCosmeticTitle{
		ID:          "test_title",
		EventID:     "spring_festival",
		DisplayName: "Test Title",
		Rarity:      "epic",
	})
	comp.SetActiveTitle("test_title")
	comp.CompleteAchievement("test_ach")
	comp.RecordEventParticipation()
	comp.RecordQuestCompletion()

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize into new component
	newComp := &EventRewardComponent{}
	err = newComp.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify data
	if newComp.GetCurrency("spring_festival") != 150 {
		t.Errorf("Currency not preserved: got %d", newComp.GetCurrency("spring_festival"))
	}
	if len(newComp.EarnedRewards) != 1 {
		t.Errorf("Rewards not preserved: got %d", len(newComp.EarnedRewards))
	}
	if len(newComp.EarnedTitles) != 1 {
		t.Errorf("Titles not preserved: got %d", len(newComp.EarnedTitles))
	}
	if newComp.ActiveTitle != "test_title" {
		t.Errorf("ActiveTitle not preserved: got %q", newComp.ActiveTitle)
	}
	if !newComp.HasAchievement("test_ach") {
		t.Error("Achievement not preserved")
	}
	if newComp.TotalEventsParticipated != 1 {
		t.Errorf("TotalEventsParticipated not preserved: got %d", newComp.TotalEventsParticipated)
	}
}

func TestEventRewardComponent_DeserializeEmpty(t *testing.T) {
	comp := &EventRewardComponent{}

	// Deserialize empty component
	err := comp.Deserialize([]byte("{}"))
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// EventCurrency should be initialized
	if comp.EventCurrency == nil {
		t.Error("EventCurrency should be initialized after deserialize")
	}

	// Should be able to use currency operations
	comp.AddCurrency("test", 10)
	if comp.GetCurrency("test") != 10 {
		t.Error("Currency operations should work after deserialize")
	}
}

func TestEventRewardComponent_DeserializeInvalid(t *testing.T) {
	comp := &EventRewardComponent{}

	err := comp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize should fail on invalid JSON")
	}
}

func TestGenerateEventRewards(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Spring Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		StartTime: time.Now(),
		EndTime:   time.Now().AddDate(0, 0, 7),
		Phase:     EventPhaseActive,
	}

	rewards := GenerateEventRewards(event, 12345)

	// Should generate 6 rewards
	if len(rewards) != 6 {
		t.Errorf("GenerateEventRewards returned %d rewards, want 6", len(rewards))
	}

	// Check reward types
	typeCount := map[EventRewardType]int{}
	for _, r := range rewards {
		typeCount[r.Type]++
		if r.EventID != event.Definition.ID {
			t.Errorf("Reward %s has wrong EventID: %q", r.ID, r.EventID)
		}
	}

	if typeCount[EventRewardCurrency] != 3 {
		t.Errorf("Expected 3 currency rewards, got %d", typeCount[EventRewardCurrency])
	}
	if typeCount[EventRewardItem] != 1 {
		t.Errorf("Expected 1 item reward, got %d", typeCount[EventRewardItem])
	}
	if typeCount[EventRewardTitle] != 1 {
		t.Errorf("Expected 1 title reward, got %d", typeCount[EventRewardTitle])
	}
	if typeCount[EventRewardEffect] != 1 {
		t.Errorf("Expected 1 effect reward, got %d", typeCount[EventRewardEffect])
	}
}

func TestGenerateEventRewards_Determinism(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "summer_solstice",
			Name:  "Summer Solstice",
			Theme: EventThemeSummer,
			Seed:  54321,
		},
	}

	rewards1 := GenerateEventRewards(event, 54321)
	rewards2 := GenerateEventRewards(event, 54321)

	if len(rewards1) != len(rewards2) {
		t.Fatalf("Different reward counts: %d vs %d", len(rewards1), len(rewards2))
	}

	for i := range rewards1 {
		if rewards1[i].ID != rewards2[i].ID {
			t.Errorf("Reward %d ID differs: %q vs %q", i, rewards1[i].ID, rewards2[i].ID)
		}
		if rewards1[i].Name != rewards2[i].Name {
			t.Errorf("Reward %d Name differs: %q vs %q", i, rewards1[i].Name, rewards2[i].Name)
		}
		if rewards1[i].Value != rewards2[i].Value {
			t.Errorf("Reward %d Value differs: %d vs %d", i, rewards1[i].Value, rewards2[i].Value)
		}
	}
}

func TestGenerateEventRewards_AllThemes(t *testing.T) {
	themes := []EventTheme{
		EventThemeSpring,
		EventThemeSummer,
		EventThemeAutumn,
		EventThemeWinter,
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			event := EventInstance{
				Definition: EventDefinition{
					ID:    string(theme) + "_event",
					Name:  string(theme) + " Event",
					Theme: theme,
					Seed:  int64(len(theme)),
				},
			}

			rewards := GenerateEventRewards(event, int64(len(theme)))

			if len(rewards) != 6 {
				t.Errorf("Expected 6 rewards, got %d", len(rewards))
			}

			// Verify each reward has content
			for _, r := range rewards {
				if r.Name == "" {
					t.Errorf("Reward %s has empty name", r.ID)
				}
			}
		})
	}
}

func TestGenerateEventAchievements(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "autumn_harvest",
			Name:  "Autumn Harvest",
			Theme: EventThemeAutumn,
			Seed:  99999,
		},
	}

	achievements := GenerateEventAchievements(event, 99999)

	// Should generate 5 achievements
	if len(achievements) != 5 {
		t.Errorf("GenerateEventAchievements returned %d, want 5", len(achievements))
	}

	// Check each achievement
	requirements := make(map[string]bool)
	for _, a := range achievements {
		if a.EventID != event.Definition.ID {
			t.Errorf("Achievement %s has wrong EventID: %q", a.ID, a.EventID)
		}
		if a.Name == "" {
			t.Errorf("Achievement %s has empty name", a.ID)
		}
		if a.Description == "" {
			t.Errorf("Achievement %s has empty description", a.ID)
		}
		if a.RequiredAmount <= 0 {
			t.Errorf("Achievement %s has invalid RequiredAmount: %d", a.ID, a.RequiredAmount)
		}
		if a.Reward.ID == "" {
			t.Errorf("Achievement %s has empty reward ID", a.ID)
		}
		requirements[a.Requirement] = true
	}

	// Verify we have diverse requirements
	expectedReqs := []string{"complete_quests", "earn_currency", "participate", "defeat_boss", "explore_location"}
	for _, req := range expectedReqs {
		if !requirements[req] {
			t.Errorf("Missing achievement with requirement %q", req)
		}
	}
}

func TestGenerateEventAchievements_Determinism(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "winter_celebration",
			Name:  "Winter Celebration",
			Theme: EventThemeWinter,
			Seed:  11111,
		},
	}

	ach1 := GenerateEventAchievements(event, 11111)
	ach2 := GenerateEventAchievements(event, 11111)

	if len(ach1) != len(ach2) {
		t.Fatalf("Different achievement counts: %d vs %d", len(ach1), len(ach2))
	}

	for i := range ach1 {
		if ach1[i].ID != ach2[i].ID {
			t.Errorf("Achievement %d ID differs", i)
		}
		if ach1[i].Reward.Value != ach2[i].Reward.Value {
			t.Errorf("Achievement %d reward value differs: %d vs %d", i, ach1[i].Reward.Value, ach2[i].Reward.Value)
		}
	}
}

func TestEventRewardTypes(t *testing.T) {
	types := []EventRewardType{
		EventRewardItem,
		EventRewardCurrency,
		EventRewardTitle,
		EventRewardEffect,
		EventRewardAchievement,
	}

	for _, rt := range types {
		if rt == "" {
			t.Error("EventRewardType should not be empty")
		}
	}
}

func TestEventRewardComponent_MultipleCurrencies(t *testing.T) {
	comp := NewEventRewardComponent()

	// Add currency for multiple events
	comp.AddCurrency("spring", 100)
	comp.AddCurrency("summer", 200)
	comp.AddCurrency("autumn", 300)

	if comp.GetCurrency("spring") != 100 {
		t.Errorf("Spring currency = %d, want 100", comp.GetCurrency("spring"))
	}
	if comp.GetCurrency("summer") != 200 {
		t.Errorf("Summer currency = %d, want 200", comp.GetCurrency("summer"))
	}
	if comp.GetCurrency("autumn") != 300 {
		t.Errorf("Autumn currency = %d, want 300", comp.GetCurrency("autumn"))
	}
	if comp.TotalCurrencyEarned != 600 {
		t.Errorf("TotalCurrencyEarned = %d, want 600", comp.TotalCurrencyEarned)
	}
}

func TestEventRewardComponent_JSONSerialization(t *testing.T) {
	comp := NewEventRewardComponent()
	comp.AddCurrency("test_event", 500)
	comp.TotalEventsParticipated = 3

	data, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded EventRewardComponent
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded.GetCurrency("test_event") != 500 {
		t.Errorf("Currency not preserved: got %d", decoded.GetCurrency("test_event"))
	}
	if decoded.TotalEventsParticipated != 3 {
		t.Errorf("TotalEventsParticipated not preserved: got %d", decoded.TotalEventsParticipated)
	}
}

func TestGetThemeHelperFunctions(t *testing.T) {
	themes := []EventTheme{
		EventThemeSpring,
		EventThemeSummer,
		EventThemeAutumn,
		EventThemeWinter,
		"unknown", // Test default case
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			currencies := getThemeCurrencyName(theme)
			if len(currencies) == 0 {
				t.Error("getThemeCurrencyName returned empty")
			}

			items := getThemeItemRewards(theme)
			if len(items) == 0 {
				t.Error("getThemeItemRewards returned empty")
			}

			titles := getThemeTitles(theme)
			if len(titles) == 0 {
				t.Error("getThemeTitles returned empty")
			}

			effects := getThemeEffects(theme)
			if len(effects) == 0 {
				t.Error("getThemeEffects returned empty")
			}
		})
	}
}
