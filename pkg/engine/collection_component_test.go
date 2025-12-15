// Package engine provides collection tracking components for the ECS.
// This file contains tests for the collection components.
// Phase 97: Collection System (V18.0)

package engine

import (
	"testing"
)

func TestCollectionCategory(t *testing.T) {
	tests := []struct {
		name     string
		category CollectionCategory
		wantStr  string
	}{
		{"fish", CollectionCategoryFish, "fish"},
		{"resources", CollectionCategoryResources, "resources"},
		{"creatures", CollectionCategoryCreatures, "creatures"},
		{"artifacts", CollectionCategoryArtifacts, "artifacts"},
		{"lore", CollectionCategoryLore, "lore"},
		{"recipes", CollectionCategoryRecipes, "recipes"},
		{"cosmetics", CollectionCategoryCosmetics, "cosmetics"},
		{"achievements", CollectionCategoryAchievements, "achievements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.String(); got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestAllCollectionCategories(t *testing.T) {
	categories := AllCollectionCategories()
	if len(categories) != 8 {
		t.Errorf("AllCollectionCategories() returned %d categories, want 8", len(categories))
	}
}

func TestCollectionRarityPoints(t *testing.T) {
	tests := []struct {
		name   string
		rarity CollectionRarity
		want   int
	}{
		{"common", CollectionRarityCommon, 1},
		{"uncommon", CollectionRarityUncommon, 3},
		{"rare", CollectionRarityRare, 5},
		{"epic", CollectionRarityEpic, 10},
		{"legendary", CollectionRarityLegendary, 25},
		{"unknown", CollectionRarity("unknown"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rarity.Points(); got != tt.want {
				t.Errorf("Points() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultMilestones(t *testing.T) {
	milestones := DefaultMilestones()
	if len(milestones) != 4 {
		t.Errorf("DefaultMilestones() returned %d milestones, want 4", len(milestones))
	}

	expectedThresholds := []int{25, 50, 75, 100}
	for i, m := range milestones {
		if m.Threshold != expectedThresholds[i] {
			t.Errorf("Milestone %d threshold = %d, want %d", i, m.Threshold, expectedThresholds[i])
		}
	}
}

func TestNewCollectionComponent(t *testing.T) {
	c := NewCollectionComponent()

	if c == nil {
		t.Fatal("NewCollectionComponent() returned nil")
	}

	if c.Type() != "collection" {
		t.Errorf("Type() = %v, want collection", c.Type())
	}

	if c.TotalPoints != 0 {
		t.Errorf("TotalPoints = %d, want 0", c.TotalPoints)
	}

	if len(c.Discovered) != 0 {
		t.Errorf("Discovered has %d entries, want 0", len(c.Discovered))
	}

	// Check all categories are initialized
	for _, cat := range AllCollectionCategories() {
		if _, ok := c.CategoryCounts[cat]; !ok {
			t.Errorf("CategoryCounts missing category %s", cat)
		}
		if _, ok := c.TotalInCategory[cat]; !ok {
			t.Errorf("TotalInCategory missing category %s", cat)
		}
	}
}

func TestCollectionComponent_AddCollectible(t *testing.T) {
	c := NewCollectionComponent()

	// First discovery should return true
	isNew := c.AddCollectible("fish_001", "Golden Trout", CollectionCategoryFish, CollectionRarityRare, "A shimmering golden fish", 1000)
	if !isNew {
		t.Error("First AddCollectible should return true")
	}

	// Second discovery of same item should return false
	isNew = c.AddCollectible("fish_001", "Golden Trout", CollectionCategoryFish, CollectionRarityRare, "A shimmering golden fish", 2000)
	if isNew {
		t.Error("Second AddCollectible of same item should return false")
	}

	// Check count incremented
	entry := c.GetCollectible("fish_001")
	if entry == nil {
		t.Fatal("GetCollectible returned nil")
	}
	if entry.Count != 2 {
		t.Errorf("Count = %d, want 2", entry.Count)
	}

	// Check points
	if c.GetTotalPoints() != 5 { // Rare = 5 points
		t.Errorf("TotalPoints = %d, want 5", c.GetTotalPoints())
	}

	// Check category count
	discovered, _ := c.GetCategoryProgress(CollectionCategoryFish)
	if discovered != 1 {
		t.Errorf("Fish discovered = %d, want 1", discovered)
	}
}

func TestCollectionComponent_HasCollectible(t *testing.T) {
	c := NewCollectionComponent()

	if c.HasCollectible("nonexistent") {
		t.Error("HasCollectible should return false for nonexistent item")
	}

	c.AddCollectible("item_001", "Test Item", CollectionCategoryArtifacts, CollectionRarityCommon, "Test", 1000)

	if !c.HasCollectible("item_001") {
		t.Error("HasCollectible should return true for existing item")
	}
}

func TestCollectionComponent_GetCategoryProgress(t *testing.T) {
	c := NewCollectionComponent()

	// Add some fish
	c.AddCollectible("fish_001", "Trout", CollectionCategoryFish, CollectionRarityCommon, "", 1000)
	c.AddCollectible("fish_002", "Salmon", CollectionCategoryFish, CollectionRarityUncommon, "", 1001)
	c.AddCollectible("fish_003", "Bass", CollectionCategoryFish, CollectionRarityCommon, "", 1002)

	discovered, total := c.GetCategoryProgress(CollectionCategoryFish)
	if discovered != 3 {
		t.Errorf("Discovered = %d, want 3", discovered)
	}
	if total != 14 { // Default fish total
		t.Errorf("Total = %d, want 14", total)
	}
}

func TestCollectionComponent_GetCategoryCompletionPercent(t *testing.T) {
	c := NewCollectionComponent()
	c.SetCategoryTotal(CollectionCategoryFish, 10) // Set to 10 for easier testing

	c.AddCollectible("fish_001", "Fish 1", CollectionCategoryFish, CollectionRarityCommon, "", 1000)
	c.AddCollectible("fish_002", "Fish 2", CollectionCategoryFish, CollectionRarityCommon, "", 1001)
	c.AddCollectible("fish_003", "Fish 3", CollectionCategoryFish, CollectionRarityCommon, "", 1002)

	percent := c.GetCategoryCompletionPercent(CollectionCategoryFish)
	if percent != 30 {
		t.Errorf("Completion percent = %f, want 30", percent)
	}

	// Test empty category
	c.SetCategoryTotal(CollectionCategoryLore, 0)
	percent = c.GetCategoryCompletionPercent(CollectionCategoryLore)
	if percent != 0 {
		t.Errorf("Empty category percent = %f, want 0", percent)
	}
}

func TestCollectionComponent_GetOverallCompletionPercent(t *testing.T) {
	c := NewCollectionComponent()

	// Set all categories to 10 for easier testing
	for _, cat := range AllCollectionCategories() {
		c.SetCategoryTotal(cat, 10)
	}

	// Add 8 items (1 per category)
	c.AddCollectible("fish_001", "Fish", CollectionCategoryFish, CollectionRarityCommon, "", 1000)
	c.AddCollectible("res_001", "Resource", CollectionCategoryResources, CollectionRarityCommon, "", 1001)
	c.AddCollectible("cre_001", "Creature", CollectionCategoryCreatures, CollectionRarityCommon, "", 1002)
	c.AddCollectible("art_001", "Artifact", CollectionCategoryArtifacts, CollectionRarityCommon, "", 1003)
	c.AddCollectible("lore_001", "Lore", CollectionCategoryLore, CollectionRarityCommon, "", 1004)
	c.AddCollectible("rec_001", "Recipe", CollectionCategoryRecipes, CollectionRarityCommon, "", 1005)
	c.AddCollectible("cos_001", "Cosmetic", CollectionCategoryCosmetics, CollectionRarityCommon, "", 1006)
	c.AddCollectible("ach_001", "Achievement", CollectionCategoryAchievements, CollectionRarityCommon, "", 1007)

	// 8 items / 80 total = 10%
	percent := c.GetOverallCompletionPercent()
	if percent != 10 {
		t.Errorf("Overall completion = %f, want 10", percent)
	}
}

func TestCollectionComponent_GetDiscoveredCount(t *testing.T) {
	c := NewCollectionComponent()

	if c.GetDiscoveredCount() != 0 {
		t.Errorf("Initial count = %d, want 0", c.GetDiscoveredCount())
	}

	c.AddCollectible("item_001", "Item 1", CollectionCategoryArtifacts, CollectionRarityCommon, "", 1000)
	c.AddCollectible("item_002", "Item 2", CollectionCategoryArtifacts, CollectionRarityCommon, "", 1001)

	if c.GetDiscoveredCount() != 2 {
		t.Errorf("Count after adds = %d, want 2", c.GetDiscoveredCount())
	}
}

func TestCollectionComponent_GetDiscoveredInCategory(t *testing.T) {
	c := NewCollectionComponent()

	c.AddCollectible("fish_001", "Fish 1", CollectionCategoryFish, CollectionRarityCommon, "", 1000)
	c.AddCollectible("fish_002", "Fish 2", CollectionCategoryFish, CollectionRarityRare, "", 1001)
	c.AddCollectible("art_001", "Artifact 1", CollectionCategoryArtifacts, CollectionRarityEpic, "", 1002)

	fish := c.GetDiscoveredInCategory(CollectionCategoryFish)
	if len(fish) != 2 {
		t.Errorf("Fish count = %d, want 2", len(fish))
	}

	artifacts := c.GetDiscoveredInCategory(CollectionCategoryArtifacts)
	if len(artifacts) != 1 {
		t.Errorf("Artifacts count = %d, want 1", len(artifacts))
	}

	lore := c.GetDiscoveredInCategory(CollectionCategoryLore)
	if len(lore) != 0 {
		t.Errorf("Lore count = %d, want 0", len(lore))
	}
}

func TestCollectionComponent_CheckMilestone(t *testing.T) {
	c := NewCollectionComponent()
	c.SetCategoryTotal(CollectionCategoryFish, 4) // Small total for testing

	// Add 1 fish (25%)
	c.AddCollectible("fish_001", "Fish 1", CollectionCategoryFish, CollectionRarityCommon, "", 1000)

	milestone := c.CheckMilestone(CollectionCategoryFish, 25)
	if milestone == nil {
		t.Fatal("Expected milestone for 25% completion")
	}
	if milestone.Threshold != 25 {
		t.Errorf("Milestone threshold = %d, want 25", milestone.Threshold)
	}

	// Check again - should return nil (already unlocked)
	milestone = c.CheckMilestone(CollectionCategoryFish, 25)
	if milestone != nil {
		t.Error("Should not return milestone again for same threshold")
	}

	// Check 50% - should fail (only at 25%)
	milestone = c.CheckMilestone(CollectionCategoryFish, 50)
	if milestone != nil {
		t.Error("Should not return milestone for 50% when only at 25%")
	}

	// Add another fish (50%)
	c.AddCollectible("fish_002", "Fish 2", CollectionCategoryFish, CollectionRarityCommon, "", 1001)

	milestone = c.CheckMilestone(CollectionCategoryFish, 50)
	if milestone == nil {
		t.Fatal("Expected milestone for 50% completion")
	}
}

func TestCollectionComponent_ClaimReward(t *testing.T) {
	c := NewCollectionComponent()

	// First claim should succeed
	if !c.ClaimReward("reward_001") {
		t.Error("First ClaimReward should return true")
	}

	// Second claim should fail
	if c.ClaimReward("reward_001") {
		t.Error("Second ClaimReward should return false")
	}

	// Check claimed status
	if !c.HasClaimedReward("reward_001") {
		t.Error("HasClaimedReward should return true")
	}
	if c.HasClaimedReward("reward_002") {
		t.Error("HasClaimedReward should return false for unclaimed")
	}
}

func TestCollectionComponent_Favorites(t *testing.T) {
	c := NewCollectionComponent()

	// Add collectible first
	c.AddCollectible("item_001", "Item 1", CollectionCategoryArtifacts, CollectionRarityRare, "", 1000)

	// Can't add non-existent item as favorite
	if c.AddFavorite("nonexistent") {
		t.Error("AddFavorite should fail for non-existent item")
	}

	// Add as favorite
	if !c.AddFavorite("item_001") {
		t.Error("AddFavorite should succeed for existing item")
	}

	// Can't add same item twice
	if c.AddFavorite("item_001") {
		t.Error("AddFavorite should fail for already favorited item")
	}

	// Check IsFavorite
	if !c.IsFavorite("item_001") {
		t.Error("IsFavorite should return true")
	}
	if c.IsFavorite("item_002") {
		t.Error("IsFavorite should return false for non-favorite")
	}

	// Get favorites
	favs := c.GetFavorites()
	if len(favs) != 1 || favs[0] != "item_001" {
		t.Errorf("GetFavorites = %v, want [item_001]", favs)
	}

	// Remove favorite
	if !c.RemoveFavorite("item_001") {
		t.Error("RemoveFavorite should succeed")
	}

	// Can't remove again
	if c.RemoveFavorite("item_001") {
		t.Error("RemoveFavorite should fail for non-favorite")
	}

	if c.IsFavorite("item_001") {
		t.Error("IsFavorite should return false after removal")
	}
}

func TestCollectionComponent_Serialization(t *testing.T) {
	c := NewCollectionComponent()

	c.AddCollectible("fish_001", "Trout", CollectionCategoryFish, CollectionRarityRare, "A tasty fish", 1000)
	c.AddCollectible("art_001", "Ancient Coin", CollectionCategoryArtifacts, CollectionRarityLegendary, "Very old", 2000)
	c.AddFavorite("fish_001")
	c.ClaimReward("reward_001")

	// Serialize
	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize into new component
	c2 := NewCollectionComponent()
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify state
	if c2.GetDiscoveredCount() != 2 {
		t.Errorf("Deserialized count = %d, want 2", c2.GetDiscoveredCount())
	}
	if c2.GetTotalPoints() != 30 { // Rare(5) + Legendary(25)
		t.Errorf("Deserialized points = %d, want 30", c2.GetTotalPoints())
	}
	if !c2.IsFavorite("fish_001") {
		t.Error("Favorite not preserved after serialization")
	}
	if !c2.HasClaimedReward("reward_001") {
		t.Error("Claimed reward not preserved after serialization")
	}
}

func TestGetCategoryTotal(t *testing.T) {
	tests := []struct {
		category CollectionCategory
		want     int
	}{
		{CollectionCategoryFish, 14},
		{CollectionCategoryResources, 6},
		{CollectionCategoryCreatures, 50},
		{CollectionCategoryArtifacts, 25},
		{CollectionCategoryLore, 30},
		{CollectionCategoryRecipes, 40},
		{CollectionCategoryCosmetics, 20},
		{CollectionCategoryAchievements, 60},
		{CollectionCategory("unknown"), 10},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			if got := GetCategoryTotal(tt.category); got != tt.want {
				t.Errorf("GetCategoryTotal(%s) = %d, want %d", tt.category, got, tt.want)
			}
		})
	}
}

// CollectibleComponent tests

func TestNewCollectibleComponent(t *testing.T) {
	c := NewCollectibleComponent("art_ancient_sword", "Ancient Sword", CollectionCategoryArtifacts, CollectionRarityEpic, "A sword from ages past")

	if c == nil {
		t.Fatal("NewCollectibleComponent returned nil")
	}

	if c.Type() != "collectible" {
		t.Errorf("Type() = %v, want collectible", c.Type())
	}

	if c.GetCollectibleID() != "art_ancient_sword" {
		t.Errorf("GetCollectibleID() = %v, want art_ancient_sword", c.GetCollectibleID())
	}

	if c.GetCategory() != CollectionCategoryArtifacts {
		t.Errorf("GetCategory() = %v, want artifacts", c.GetCategory())
	}

	if c.GetRarity() != CollectionRarityEpic {
		t.Errorf("GetRarity() = %v, want epic", c.GetRarity())
	}

	if c.GetName() != "Ancient Sword" {
		t.Errorf("GetName() = %v, want Ancient Sword", c.GetName())
	}

	if c.GetDescription() != "A sword from ages past" {
		t.Errorf("GetDescription() = %v, want 'A sword from ages past'", c.GetDescription())
	}

	if c.IsAlreadyCollected() {
		t.Error("New collectible should not be collected")
	}
}

func TestCollectibleComponent_CanCollect(t *testing.T) {
	c := NewCollectibleComponent("item_001", "Test Item", CollectionCategoryArtifacts, CollectionRarityCommon, "")
	c.SetRequiredLevel(10)

	// Below required level
	if c.CanCollect(5) {
		t.Error("CanCollect should return false for level < required")
	}

	// At required level
	if !c.CanCollect(10) {
		t.Error("CanCollect should return true for level = required")
	}

	// Above required level
	if !c.CanCollect(15) {
		t.Error("CanCollect should return true for level > required")
	}

	// After collection
	c.Collect(1, 1000)
	if c.CanCollect(20) {
		t.Error("CanCollect should return false after collection")
	}
}

func TestCollectibleComponent_Collect(t *testing.T) {
	c := NewCollectibleComponent("item_001", "Test Item", CollectionCategoryArtifacts, CollectionRarityCommon, "")

	// First collection should succeed
	if !c.Collect(123, 1000) {
		t.Error("First Collect should return true")
	}

	if !c.IsAlreadyCollected() {
		t.Error("IsAlreadyCollected should return true after collection")
	}

	// Second collection should fail
	if c.Collect(456, 2000) {
		t.Error("Second Collect should return false")
	}
}

func TestCollectibleComponent_Hidden(t *testing.T) {
	c := NewCollectibleComponent("secret_001", "Secret Item", CollectionCategoryArtifacts, CollectionRarityLegendary, "")

	if c.IsHiddenCollectible() {
		t.Error("New collectible should not be hidden")
	}

	c.SetHidden(true, "Look under the old tree")

	if !c.IsHiddenCollectible() {
		t.Error("Should be hidden after SetHidden(true)")
	}

	if c.GetHint() != "Look under the old tree" {
		t.Errorf("GetHint() = %v, want 'Look under the old tree'", c.GetHint())
	}
}

func TestCollectibleComponent_RequiredLevel(t *testing.T) {
	c := NewCollectibleComponent("item_001", "Test Item", CollectionCategoryArtifacts, CollectionRarityCommon, "")

	if c.GetRequiredLevel() != 0 {
		t.Errorf("Initial required level = %d, want 0", c.GetRequiredLevel())
	}

	c.SetRequiredLevel(25)
	if c.GetRequiredLevel() != 25 {
		t.Errorf("Required level after set = %d, want 25", c.GetRequiredLevel())
	}
}

func TestCollectibleComponent_Serialization(t *testing.T) {
	c := NewCollectibleComponent("item_001", "Test Item", CollectionCategoryArtifacts, CollectionRarityEpic, "Test description")
	c.SetHidden(true, "A hidden hint")
	c.SetRequiredLevel(15)
	c.Collect(999, 12345)

	// Serialize
	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize
	c2 := NewCollectibleComponent("", "", "", "", "")
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if c2.GetCollectibleID() != "item_001" {
		t.Errorf("Deserialized ID = %v, want item_001", c2.GetCollectibleID())
	}
	if !c2.IsAlreadyCollected() {
		t.Error("Deserialized should be collected")
	}
	if !c2.IsHiddenCollectible() {
		t.Error("Deserialized should be hidden")
	}
	if c2.GetRequiredLevel() != 15 {
		t.Errorf("Deserialized required level = %d, want 15", c2.GetRequiredLevel())
	}
}

// Concurrent access tests

func TestCollectionComponent_ConcurrentAccess(t *testing.T) {
	c := NewCollectionComponent()

	done := make(chan bool)

	// Spawn multiple goroutines to add collectibles
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				id := string(rune('A'+idx)) + "_" + string(rune('0'+j%10))
				c.AddCollectible(id, "Test", CollectionCategoryArtifacts, CollectionRarityCommon, "", int64(j))
				c.HasCollectible(id)
				c.GetCategoryProgress(CollectionCategoryArtifacts)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
	count := c.GetDiscoveredCount()
	if count == 0 {
		t.Error("Expected some collectibles to be added")
	}
}

func TestCollectibleComponent_ConcurrentAccess(t *testing.T) {
	c := NewCollectibleComponent("item_001", "Test", CollectionCategoryArtifacts, CollectionRarityCommon, "")

	done := make(chan bool)

	// Read operations
	go func() {
		for i := 0; i < 100; i++ {
			c.GetCollectibleID()
			c.GetCategory()
			c.IsAlreadyCollected()
			c.CanCollect(10)
		}
		done <- true
	}()

	// Write operation (only once)
	go func() {
		c.Collect(123, 1000)
		done <- true
	}()

	<-done
	<-done
}
