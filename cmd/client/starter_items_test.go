//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// TestGenerateStarterWeapon tests starter weapon generation.
func TestGenerateStarterWeapon(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		genreID string
	}{
		{"fantasy seed 12345", 12345, "fantasy"},
		{"scifi seed 67890", 67890, "scifi"},
		{"horror seed 11111", 11111, "horror"},
		{"cyberpunk seed 22222", 22222, "cyberpunk"},
		{"postapoc seed 33333", 33333, "postapocalyptic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := &engine.InventoryComponent{
				Items:    make([]*item.Item, 0),
				MaxItems: 50,
			}
			itemGen := item.NewItemGenerator()
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)
			entry := logger.WithField("test", true)

			initialCount := len(inventory.Items)
			generateStarterWeapon(inventory, itemGen, tt.seed, tt.genreID, entry)

			// Should add exactly one weapon
			if len(inventory.Items) != initialCount+1 {
				t.Errorf("item count = %d, want %d", len(inventory.Items), initialCount+1)
			}

			// Verify weapon properties
			if len(inventory.Items) > 0 {
				weapon := inventory.Items[len(inventory.Items)-1]
				if weapon == nil {
					t.Fatal("weapon is nil")
				}
				// Starter weapons should have "Rusty" prefix
				if len(weapon.Name) < 5 || weapon.Name[:5] != "Rusty" {
					t.Errorf("weapon name = %q, should start with 'Rusty'", weapon.Name)
				}
				// Starter weapons have low value
				if weapon.Stats.Value != 5 {
					t.Errorf("weapon value = %d, want 5", weapon.Stats.Value)
				}
			}
		})
	}
}

// TestGenerateStarterWeaponDeterminism tests that same seed produces same weapon.
func TestGenerateStarterWeaponDeterminism(t *testing.T) {
	seed := int64(42424242)
	genreID := "fantasy"
	itemGen := item.NewItemGenerator()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	entry := logger.WithField("test", true)

	// Generate twice with same seed
	inv1 := &engine.InventoryComponent{Items: make([]*item.Item, 0), MaxItems: 50}
	inv2 := &engine.InventoryComponent{Items: make([]*item.Item, 0), MaxItems: 50}

	generateStarterWeapon(inv1, itemGen, seed, genreID, entry)
	generateStarterWeapon(inv2, itemGen, seed, genreID, entry)

	if len(inv1.Items) == 0 || len(inv2.Items) == 0 {
		t.Skip("no weapons generated")
	}

	// Same seed should produce same base weapon name (before "Rusty" modification)
	if inv1.Items[0].Name != inv2.Items[0].Name {
		t.Errorf("determinism failed: %q vs %q", inv1.Items[0].Name, inv2.Items[0].Name)
	}
}

// TestGenerateStarterPotions tests starter potion generation.
func TestGenerateStarterPotions(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		genreID string
	}{
		{"fantasy", 12345, "fantasy"},
		{"scifi", 67890, "scifi"},
		{"horror", 11111, "horror"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := &engine.InventoryComponent{
				Items:    make([]*item.Item, 0),
				MaxItems: 50,
			}
			itemGen := item.NewItemGenerator()
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)
			entry := logger.WithField("test", true)

			initialCount := len(inventory.Items)
			generateStarterPotions(inventory, itemGen, tt.seed, tt.genreID, entry)

			// Should add 2 potions (count: 2 in params)
			expectedCount := initialCount + 2
			if len(inventory.Items) != expectedCount {
				t.Errorf("item count = %d, want %d", len(inventory.Items), expectedCount)
			}

			// Verify potion properties
			for i := initialCount; i < len(inventory.Items); i++ {
				potion := inventory.Items[i]
				if potion == nil {
					continue
				}
				// All starter potions should be "Minor Health Potion"
				if potion.Name != "Minor Health Potion" {
					t.Errorf("potion name = %q, want 'Minor Health Potion'", potion.Name)
				}
				// Starter potions have low value
				if potion.Stats.Value != 10 {
					t.Errorf("potion value = %d, want 10", potion.Stats.Value)
				}
				// Starter potions have low weight
				if potion.Stats.Weight != 0.2 {
					t.Errorf("potion weight = %f, want 0.2", potion.Stats.Weight)
				}
			}
		})
	}
}

// TestGenerateStarterArmor tests starter armor generation.
func TestGenerateStarterArmor(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		genreID string
	}{
		{"fantasy", 12345, "fantasy"},
		{"scifi", 67890, "scifi"},
		{"cyberpunk", 22222, "cyberpunk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := &engine.InventoryComponent{
				Items:    make([]*item.Item, 0),
				MaxItems: 50,
			}
			itemGen := item.NewItemGenerator()
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)
			entry := logger.WithField("test", true)

			initialCount := len(inventory.Items)
			generateStarterArmor(inventory, itemGen, tt.seed, tt.genreID, entry)

			// Should add exactly one armor piece
			if len(inventory.Items) != initialCount+1 {
				t.Errorf("item count = %d, want %d", len(inventory.Items), initialCount+1)
			}

			// Verify armor properties
			if len(inventory.Items) > 0 {
				armor := inventory.Items[len(inventory.Items)-1]
				if armor == nil {
					t.Fatal("armor is nil")
				}
				// Starter armor should have "Worn" prefix
				if len(armor.Name) < 4 || armor.Name[:4] != "Worn" {
					t.Errorf("armor name = %q, should start with 'Worn'", armor.Name)
				}
				// Starter armor has low value
				if armor.Stats.Value != 8 {
					t.Errorf("armor value = %d, want 8", armor.Stats.Value)
				}
			}
		})
	}
}

// TestAddStarterItems tests the complete starter items function.
func TestAddStarterItems(t *testing.T) {
	tests := []struct {
		name          string
		seed          int64
		genreID       string
		minItemsAdded int // At least this many items should be added
	}{
		{"fantasy", 12345, "fantasy", 4}, // 1 weapon + 2 potions + 1 armor
		{"scifi", 67890, "scifi", 4},
		{"horror", 11111, "horror", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := &engine.InventoryComponent{
				Items:    make([]*item.Item, 0),
				MaxItems: 50,
			}
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			addStarterItems(inventory, tt.seed, tt.genreID, logger)

			if len(inventory.Items) < tt.minItemsAdded {
				t.Errorf("item count = %d, want at least %d", len(inventory.Items), tt.minItemsAdded)
			}
		})
	}
}

// TestAddStarterItemsDeterminism verifies same seed produces same items.
func TestAddStarterItemsDeterminism(t *testing.T) {
	seed := int64(98765)
	genreID := "fantasy"
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	inv1 := &engine.InventoryComponent{Items: make([]*item.Item, 0), MaxItems: 50}
	inv2 := &engine.InventoryComponent{Items: make([]*item.Item, 0), MaxItems: 50}

	addStarterItems(inv1, seed, genreID, logger)
	addStarterItems(inv2, seed, genreID, logger)

	if len(inv1.Items) != len(inv2.Items) {
		t.Fatalf("item counts differ: %d vs %d", len(inv1.Items), len(inv2.Items))
	}

	for i := range inv1.Items {
		if inv1.Items[i].Name != inv2.Items[i].Name {
			t.Errorf("item %d: names differ %q vs %q", i, inv1.Items[i].Name, inv2.Items[i].Name)
		}
	}
}

// TestAddTutorialQuest tests tutorial quest creation.
func TestAddTutorialQuest(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		genreID string
	}{
		{"fantasy", 12345, "fantasy"},
		{"scifi", 67890, "scifi"},
		{"horror", 11111, "horror"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &engine.QuestTrackerComponent{
				ActiveQuests:        make([]*engine.TrackedQuest, 0),
				CompletedQuests:     make([]*engine.TrackedQuest, 0),
				FailedQuests:        make([]*engine.TrackedQuest, 0),
				MaxActiveQuests:     20,
				StoryUnlockedQuests: make(map[string][]string),
			}
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			addTutorialQuest(tracker, tt.seed, tt.genreID, logger)

			// Should have exactly one active quest
			if len(tracker.ActiveQuests) != 1 {
				t.Errorf("active quests = %d, want 1", len(tracker.ActiveQuests))
			}

			// Verify quest ID matches expected format
			for _, tracked := range tracker.ActiveQuests {
				if tracked.Quest == nil {
					t.Error("tracked quest is nil")
					continue
				}
				if len(tracked.Quest.ID) < 9 || tracked.Quest.ID[:9] != "tutorial_" {
					t.Errorf("quest ID = %q, should start with 'tutorial_'", tracked.Quest.ID)
				}
			}
		})
	}
}

// TestAddTutorialQuestObjectives verifies quest objectives are correct.
func TestAddTutorialQuestObjectives(t *testing.T) {
	tracker := &engine.QuestTrackerComponent{
		ActiveQuests:        make([]*engine.TrackedQuest, 0),
		CompletedQuests:     make([]*engine.TrackedQuest, 0),
		FailedQuests:        make([]*engine.TrackedQuest, 0),
		MaxActiveQuests:     20,
		StoryUnlockedQuests: make(map[string][]string),
	}
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	addTutorialQuest(tracker, 12345, "fantasy", logger)

	// Get the quest
	if len(tracker.ActiveQuests) == 0 {
		t.Fatal("no active quests")
	}

	tracked := tracker.ActiveQuests[0]
	if tracked.Quest == nil {
		t.Fatal("quest is nil")
	}

	// Tutorial quest should have 3 objectives
	if len(tracked.Quest.Objectives) != 3 {
		t.Errorf("objective count = %d, want 3", len(tracked.Quest.Objectives))
	}

	// Verify objective targets
	expectedTargets := []string{"inventory", "questlog", "explore"}
	for i, obj := range tracked.Quest.Objectives {
		if i < len(expectedTargets) && obj.Target != expectedTargets[i] {
			t.Errorf("objective %d target = %q, want %q", i, obj.Target, expectedTargets[i])
		}
	}

	// Verify quest reward
	if tracked.Quest.Reward.XP != 50 {
		t.Errorf("reward XP = %d, want 50", tracked.Quest.Reward.XP)
	}
	if tracked.Quest.Reward.Gold != 25 {
		t.Errorf("reward Gold = %d, want 25", tracked.Quest.Reward.Gold)
	}
}

// TestAddTutorialQuestDeterminism verifies same seed produces same quest.
func TestAddTutorialQuestDeterminism(t *testing.T) {
	seed := int64(55555)
	genreID := "fantasy"
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	tracker1 := &engine.QuestTrackerComponent{
		ActiveQuests:        make([]*engine.TrackedQuest, 0),
		CompletedQuests:     make([]*engine.TrackedQuest, 0),
		FailedQuests:        make([]*engine.TrackedQuest, 0),
		MaxActiveQuests:     20,
		StoryUnlockedQuests: make(map[string][]string),
	}
	tracker2 := &engine.QuestTrackerComponent{
		ActiveQuests:        make([]*engine.TrackedQuest, 0),
		CompletedQuests:     make([]*engine.TrackedQuest, 0),
		FailedQuests:        make([]*engine.TrackedQuest, 0),
		MaxActiveQuests:     20,
		StoryUnlockedQuests: make(map[string][]string),
	}

	addTutorialQuest(tracker1, seed, genreID, logger)
	addTutorialQuest(tracker2, seed, genreID, logger)

	// Both should have same quest ID
	if len(tracker1.ActiveQuests) == 0 || len(tracker2.ActiveQuests) == 0 {
		t.Fatal("missing active quests")
	}

	id1 := tracker1.ActiveQuests[0].Quest.ID
	id2 := tracker2.ActiveQuests[0].Quest.ID

	if id1 != id2 {
		t.Errorf("quest IDs differ: %q vs %q", id1, id2)
	}
}

// BenchmarkGenerateStarterWeapon benchmarks weapon generation.
func BenchmarkGenerateStarterWeapon(b *testing.B) {
	itemGen := item.NewItemGenerator()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	entry := logger.WithField("bench", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inventory := &engine.InventoryComponent{Items: make([]*item.Item, 0), MaxItems: 50}
		generateStarterWeapon(inventory, itemGen, int64(i), "fantasy", entry)
	}
}

// BenchmarkAddStarterItems benchmarks complete starter item generation.
func BenchmarkAddStarterItems(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inventory := &engine.InventoryComponent{Items: make([]*item.Item, 0), MaxItems: 50}
		addStarterItems(inventory, int64(i), "fantasy", logger)
	}
}

// BenchmarkAddTutorialQuest benchmarks tutorial quest creation.
func BenchmarkAddTutorialQuest(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker := &engine.QuestTrackerComponent{
			ActiveQuests:        make([]*engine.TrackedQuest, 0),
			CompletedQuests:     make([]*engine.TrackedQuest, 0),
			FailedQuests:        make([]*engine.TrackedQuest, 0),
			MaxActiveQuests:     20,
			StoryUnlockedQuests: make(map[string][]string),
		}
		addTutorialQuest(tracker, int64(i), "fantasy", logger)
	}
}
