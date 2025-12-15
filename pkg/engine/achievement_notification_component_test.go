// Package engine provides tests for the achievement notification component.
//
// Phase 85: Achievement Notifications & Rewards (V15.0)
package engine

import (
	"encoding/json"
	"testing"
)

func TestAchievementNotificationComponentType(t *testing.T) {
	comp := NewAchievementNotificationComponent()
	if comp.Type() != "achievement_notification" {
		t.Errorf("Expected type 'achievement_notification', got '%s'", comp.Type())
	}
}

func TestNewAchievementNotificationComponent(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	if comp.PendingNotifications == nil {
		t.Error("PendingNotifications should be initialized")
	}
	if comp.DisplayedNotifications == nil {
		t.Error("DisplayedNotifications should be initialized")
	}
	if comp.TotalAchievementPoints != 0 {
		t.Error("TotalAchievementPoints should be 0")
	}
	if !comp.PlaySoundOnUnlock {
		t.Error("PlaySoundOnUnlock should be true by default")
	}
	if comp.MaxHistorySize != 100 {
		t.Errorf("MaxHistorySize should be 100, got %d", comp.MaxHistorySize)
	}
}

func TestQueueNotification(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	notification := AchievementNotification{
		AchievementID:   "combat_first_blood",
		AchievementName: "First Blood",
		Category:        AchievementCategoryCombat,
		Tier:            AchievementTierBronze,
		Points:          10,
		Timestamp:       1000,
	}

	comp.QueueNotification(notification)

	if comp.GetPendingCount() != 1 {
		t.Errorf("Expected 1 pending notification, got %d", comp.GetPendingCount())
	}
	if comp.GetTotalPoints() != 10 {
		t.Errorf("Expected 10 total points, got %d", comp.GetTotalPoints())
	}

	// Queue another
	notification2 := AchievementNotification{
		AchievementID:   "combat_first_blood",
		AchievementName: "First Blood",
		Tier:            AchievementTierSilver,
		Points:          25,
		Timestamp:       2000,
	}
	comp.QueueNotification(notification2)

	if comp.GetPendingCount() != 2 {
		t.Errorf("Expected 2 pending notifications, got %d", comp.GetPendingCount())
	}
	if comp.GetTotalPoints() != 35 {
		t.Errorf("Expected 35 total points, got %d", comp.GetTotalPoints())
	}
}

func TestPopNotification(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	// Pop from empty queue
	result := comp.PopNotification()
	if result != nil {
		t.Error("Expected nil from empty queue")
	}

	// Add notifications
	n1 := AchievementNotification{
		AchievementID: "first",
		Points:        10,
	}
	n2 := AchievementNotification{
		AchievementID: "second",
		Points:        20,
	}
	comp.QueueNotification(n1)
	comp.QueueNotification(n2)

	// Pop first
	popped := comp.PopNotification()
	if popped == nil {
		t.Fatal("Expected notification, got nil")
	}
	if popped.AchievementID != "first" {
		t.Errorf("Expected 'first', got '%s'", popped.AchievementID)
	}
	if !popped.Displayed {
		t.Error("Popped notification should be marked as displayed")
	}

	// Check queue state
	if comp.GetPendingCount() != 1 {
		t.Errorf("Expected 1 pending, got %d", comp.GetPendingCount())
	}
	if comp.GetHistoryCount() != 1 {
		t.Errorf("Expected 1 in history, got %d", comp.GetHistoryCount())
	}

	// Pop second
	popped = comp.PopNotification()
	if popped.AchievementID != "second" {
		t.Errorf("Expected 'second', got '%s'", popped.AchievementID)
	}

	if comp.GetPendingCount() != 0 {
		t.Errorf("Expected 0 pending, got %d", comp.GetPendingCount())
	}
	if comp.GetHistoryCount() != 2 {
		t.Errorf("Expected 2 in history, got %d", comp.GetHistoryCount())
	}
}

func TestPeekNotification(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	// Peek empty queue
	result := comp.PeekNotification()
	if result != nil {
		t.Error("Expected nil from empty queue")
	}

	// Add notification
	n := AchievementNotification{
		AchievementID: "test",
		Points:        15,
	}
	comp.QueueNotification(n)

	// Peek should return it without removing
	peeked := comp.PeekNotification()
	if peeked == nil {
		t.Fatal("Expected notification, got nil")
	}
	if peeked.AchievementID != "test" {
		t.Errorf("Expected 'test', got '%s'", peeked.AchievementID)
	}

	// Queue should still have the notification
	if comp.GetPendingCount() != 1 {
		t.Errorf("Expected 1 pending after peek, got %d", comp.GetPendingCount())
	}
}

func TestMaxHistorySize(t *testing.T) {
	comp := NewAchievementNotificationComponent()
	comp.MaxHistorySize = 5

	// Add and pop 10 notifications
	for i := 0; i < 10; i++ {
		comp.QueueNotification(AchievementNotification{
			AchievementID: string(rune('a' + i)),
			Points:        1,
		})
	}
	for i := 0; i < 10; i++ {
		comp.PopNotification()
	}

	// History should be capped at 5
	if comp.GetHistoryCount() != 5 {
		t.Errorf("Expected history capped at 5, got %d", comp.GetHistoryCount())
	}

	// Should have the last 5 (f, g, h, i, j)
	history := comp.GetRecentHistory(10)
	if len(history) != 5 {
		t.Fatalf("Expected 5 history items, got %d", len(history))
	}
}

func TestGetRecentHistory(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	// Empty history
	result := comp.GetRecentHistory(5)
	if result != nil {
		t.Error("Expected nil from empty history")
	}

	// Add some notifications
	for i := 0; i < 5; i++ {
		comp.QueueNotification(AchievementNotification{
			AchievementID: string(rune('a' + i)),
		})
		comp.PopNotification()
	}

	// Get last 3
	recent := comp.GetRecentHistory(3)
	if len(recent) != 3 {
		t.Fatalf("Expected 3 recent, got %d", len(recent))
	}
	if recent[0].AchievementID != "c" {
		t.Errorf("Expected 'c' first, got '%s'", recent[0].AchievementID)
	}
	if recent[2].AchievementID != "e" {
		t.Errorf("Expected 'e' last, got '%s'", recent[2].AchievementID)
	}
}

func TestGetHistoryByCategory(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	// Add mixed categories
	comp.QueueNotification(AchievementNotification{
		AchievementID: "combat1",
		Category:      AchievementCategoryCombat,
	})
	comp.QueueNotification(AchievementNotification{
		AchievementID: "quest1",
		Category:      AchievementCategoryQuest,
	})
	comp.QueueNotification(AchievementNotification{
		AchievementID: "combat2",
		Category:      AchievementCategoryCombat,
	})

	// Pop all
	for comp.GetPendingCount() > 0 {
		comp.PopNotification()
	}

	// Filter by category
	combat := comp.GetHistoryByCategory(AchievementCategoryCombat)
	if len(combat) != 2 {
		t.Errorf("Expected 2 combat achievements, got %d", len(combat))
	}

	quest := comp.GetHistoryByCategory(AchievementCategoryQuest)
	if len(quest) != 1 {
		t.Errorf("Expected 1 quest achievement, got %d", len(quest))
	}

	pvp := comp.GetHistoryByCategory(AchievementCategoryPvP)
	if len(pvp) != 0 {
		t.Errorf("Expected 0 pvp achievements, got %d", len(pvp))
	}
}

func TestSetPlaySound(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	if !comp.ShouldPlaySound() {
		t.Error("ShouldPlaySound should be true by default")
	}

	comp.SetPlaySound(false)
	if comp.ShouldPlaySound() {
		t.Error("ShouldPlaySound should be false after SetPlaySound(false)")
	}

	comp.SetPlaySound(true)
	if !comp.ShouldPlaySound() {
		t.Error("ShouldPlaySound should be true after SetPlaySound(true)")
	}
}

func TestClearPending(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	// Add notifications
	for i := 0; i < 5; i++ {
		comp.QueueNotification(AchievementNotification{
			AchievementID: string(rune('a' + i)),
		})
	}

	if comp.GetPendingCount() != 5 {
		t.Fatalf("Expected 5 pending, got %d", comp.GetPendingCount())
	}

	comp.ClearPending()

	if comp.GetPendingCount() != 0 {
		t.Errorf("Expected 0 pending after clear, got %d", comp.GetPendingCount())
	}
}

func TestAchievementNotificationSerializeDeserialize(t *testing.T) {
	comp := NewAchievementNotificationComponent()
	comp.SetPlaySound(false)
	comp.MaxHistorySize = 50

	// Add notifications
	comp.QueueNotification(AchievementNotification{
		AchievementID:   "test1",
		AchievementName: "Test Achievement",
		Category:        AchievementCategoryCrafting,
		Tier:            AchievementTierGold,
		Points:          50,
		Timestamp:       12345,
		Rewards: []AchievementReward{
			{Type: AchievementRewardXP, Name: "XP Bonus", Value: 100},
		},
	})
	comp.PopNotification()

	comp.QueueNotification(AchievementNotification{
		AchievementID: "test2",
		Points:        25,
	})

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize into new component
	comp2 := NewAchievementNotificationComponent()
	err = comp2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify
	if comp2.GetTotalPoints() != comp.GetTotalPoints() {
		t.Errorf("TotalPoints mismatch: %d vs %d", comp2.GetTotalPoints(), comp.GetTotalPoints())
	}
	if comp2.GetPendingCount() != 1 {
		t.Errorf("Expected 1 pending, got %d", comp2.GetPendingCount())
	}
	if comp2.GetHistoryCount() != 1 {
		t.Errorf("Expected 1 in history, got %d", comp2.GetHistoryCount())
	}
	if comp2.ShouldPlaySound() {
		t.Error("PlaySoundOnUnlock should be false")
	}
	if comp2.MaxHistorySize != 50 {
		t.Errorf("MaxHistorySize should be 50, got %d", comp2.MaxHistorySize)
	}

	// Check history item
	history := comp2.GetRecentHistory(1)
	if len(history) != 1 {
		t.Fatal("Expected 1 history item")
	}
	if history[0].AchievementName != "Test Achievement" {
		t.Errorf("Expected 'Test Achievement', got '%s'", history[0].AchievementName)
	}
	if len(history[0].Rewards) != 1 {
		t.Errorf("Expected 1 reward, got %d", len(history[0].Rewards))
	}
}

func TestAchievementNotificationDeserializeEmpty(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	// Deserialize empty data
	err := comp.Deserialize([]byte(`{}`))
	if err != nil {
		t.Fatalf("Deserialize empty failed: %v", err)
	}

	if comp.PendingNotifications == nil {
		t.Error("PendingNotifications should be initialized after empty deserialize")
	}
	if comp.DisplayedNotifications == nil {
		t.Error("DisplayedNotifications should be initialized after empty deserialize")
	}
	if comp.MaxHistorySize != 100 {
		t.Errorf("MaxHistorySize should default to 100, got %d", comp.MaxHistorySize)
	}
}

func TestAchievementNotificationDeserializeInvalid(t *testing.T) {
	comp := NewAchievementNotificationComponent()

	err := comp.Deserialize([]byte(`not valid json`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestAchievementRewardTypes(t *testing.T) {
	rewards := []struct {
		rewardType AchievementRewardType
		expected   string
	}{
		{AchievementRewardXP, "xp"},
		{AchievementRewardItem, "item"},
		{AchievementRewardTitle, "title"},
		{AchievementRewardCurrency, "currency"},
	}

	for _, tc := range rewards {
		if string(tc.rewardType) != tc.expected {
			t.Errorf("Expected '%s', got '%s'", tc.expected, string(tc.rewardType))
		}
	}
}

func TestAchievementNotificationConcurrency(t *testing.T) {
	comp := NewAchievementNotificationComponent()
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				comp.QueueNotification(AchievementNotification{
					AchievementID: string(rune('a' + id)),
					Points:        1,
				})
			}
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = comp.GetPendingCount()
				_ = comp.GetTotalPoints()
				_ = comp.PeekNotification()
			}
			done <- true
		}()
	}

	// Concurrent poppers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				comp.PopNotification()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify no panics and data integrity
	total := comp.GetPendingCount() + comp.GetHistoryCount()
	if total > 1000 {
		t.Errorf("Total notifications should not exceed 1000, got %d", total)
	}
}

func TestAchievementRewardSerialization(t *testing.T) {
	reward := AchievementReward{
		Type:     AchievementRewardItem,
		Name:     "Test Item",
		Value:    5,
		ItemSeed: 12345,
	}

	data, err := json.Marshal(reward)
	if err != nil {
		t.Fatalf("Failed to marshal reward: %v", err)
	}

	var restored AchievementReward
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal reward: %v", err)
	}

	if restored.Type != AchievementRewardItem {
		t.Errorf("Type mismatch: %v", restored.Type)
	}
	if restored.Name != "Test Item" {
		t.Errorf("Name mismatch: %s", restored.Name)
	}
	if restored.Value != 5 {
		t.Errorf("Value mismatch: %d", restored.Value)
	}
	if restored.ItemSeed != 12345 {
		t.Errorf("ItemSeed mismatch: %d", restored.ItemSeed)
	}
}
