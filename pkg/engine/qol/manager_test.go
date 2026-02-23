package qol

import (
	"fmt"
	"testing"
	"time"
)

// AutoLootManager Tests

func TestAutoLootManager_SetConfig(t *testing.T) {
	mgr := NewAutoLootManager()
	config := DefaultAutoLootConfig(1)
	config.Radius = 8.0

	mgr.SetConfig(config)

	got := mgr.GetConfig(1)
	if got.Radius != 8.0 {
		t.Errorf("Radius = %f, want 8.0", got.Radius)
	}
}

func TestAutoLootManager_SetRadius(t *testing.T) {
	mgr := NewAutoLootManager()

	tests := []struct {
		input float64
		want  float64
	}{
		{3.0, 5.0},   // Below min
		{7.0, 7.0},   // Valid
		{12.0, 10.0}, // Above max
	}

	for _, tt := range tests {
		mgr.SetRadius(1, tt.input)
		config := mgr.GetConfig(1)
		if config.Radius != tt.want {
			t.Errorf("SetRadius(%f): got %f, want %f", tt.input, config.Radius, tt.want)
		}
	}
}

func TestAutoLootManager_ShouldCollect(t *testing.T) {
	mgr := NewAutoLootManager()
	config := DefaultAutoLootConfig(1)
	config.MinRarity = 2
	config.IgnoreTypes = []string{"junk"}
	mgr.SetConfig(config)

	tests := []struct {
		name     string
		rarity   int
		itemType string
		want     bool
	}{
		{"below min rarity", 1, "weapon", false},
		{"valid rarity", 3, "weapon", true},
		{"ignored type", 3, "junk", false},
		{"valid item", 2, "armor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mgr.ShouldCollect(1, tt.rarity, tt.itemType); got != tt.want {
				t.Errorf("ShouldCollect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoLootManager_FilterTypes(t *testing.T) {
	mgr := NewAutoLootManager()
	config := DefaultAutoLootConfig(1)
	config.FilterTypes = []string{"weapon", "armor"}
	mgr.SetConfig(config)

	if mgr.ShouldCollect(1, 0, "potion") {
		t.Error("Should not collect potion when filtering for weapon/armor")
	}

	if !mgr.ShouldCollect(1, 0, "weapon") {
		t.Error("Should collect weapon")
	}
}

// CraftQueueManager Tests

func TestCraftQueueManager_AddRecipe(t *testing.T) {
	mgr := NewCraftQueueManager()

	err := mgr.AddRecipe(1, "iron_sword", 5)
	if err != nil {
		t.Errorf("AddRecipe failed: %v", err)
	}

	queue := mgr.GetQueue(1)
	if len(queue) != 1 {
		t.Errorf("Queue length = %d, want 1", len(queue))
	}
	if queue[0].RecipeID != "iron_sword" {
		t.Errorf("RecipeID = %s, want iron_sword", queue[0].RecipeID)
	}
	if queue[0].Quantity != 5 {
		t.Errorf("Quantity = %d, want 5", queue[0].Quantity)
	}
}

func TestCraftQueueManager_InvalidQuantity(t *testing.T) {
	mgr := NewCraftQueueManager()

	err := mgr.AddRecipe(1, "iron_sword", 0)
	if err == nil {
		t.Error("Expected error for zero quantity")
	}

	err = mgr.AddRecipe(1, "iron_sword", -1)
	if err == nil {
		t.Error("Expected error for negative quantity")
	}
}

func TestCraftQueueManager_QueueLimit(t *testing.T) {
	mgr := NewCraftQueueManager()

	for i := 0; i < 50; i++ {
		err := mgr.AddRecipe(1, fmt.Sprintf("recipe_%d", i), 1)
		if err != nil {
			t.Errorf("Failed to add recipe %d: %v", i, err)
		}
	}

	err := mgr.AddRecipe(1, "recipe_51", 1)
	if err == nil {
		t.Error("Expected error when queue is full")
	}
}

func TestCraftQueueManager_RemoveRecipe(t *testing.T) {
	mgr := NewCraftQueueManager()
	mgr.AddRecipe(1, "recipe_a", 1)
	mgr.AddRecipe(1, "recipe_b", 2)
	mgr.AddRecipe(1, "recipe_c", 3)

	err := mgr.RemoveRecipe(1, 1)
	if err != nil {
		t.Errorf("RemoveRecipe failed: %v", err)
	}

	queue := mgr.GetQueue(1)
	if len(queue) != 2 {
		t.Errorf("Queue length = %d, want 2", len(queue))
	}

	if queue[1].RecipeID != "recipe_c" || queue[1].Position != 1 {
		t.Error("Recipe positions not updated correctly")
	}
}

func TestCraftQueueManager_ClearQueue(t *testing.T) {
	mgr := NewCraftQueueManager()
	mgr.AddRecipe(1, "recipe_a", 1)
	mgr.AddRecipe(1, "recipe_b", 2)

	mgr.ClearQueue(1)

	queue := mgr.GetQueue(1)
	if len(queue) != 0 {
		t.Errorf("Queue length = %d, want 0", len(queue))
	}
}

// GuildInvitationManager Tests

func TestGuildInvitationManager_SendInvitation(t *testing.T) {
	mgr := NewGuildInvitationManager()

	inv := &GuildInvitation{
		InvitationID: "inv1",
		GuildID:      "guild1",
		GuildName:    "Test Guild",
		InviterID:    "player1",
		InviteeID:    "player2",
	}

	mgr.SendInvitation(inv)

	if inv.SentAt.IsZero() {
		t.Error("SentAt not set")
	}
	if inv.ExpiresAt.IsZero() {
		t.Error("ExpiresAt not set")
	}

	pending := mgr.GetPendingInvitations("player2")
	if len(pending) != 1 {
		t.Errorf("Pending invitations = %d, want 1", len(pending))
	}
}

func TestGuildInvitationManager_AcceptInvitation(t *testing.T) {
	mgr := NewGuildInvitationManager()

	inv := &GuildInvitation{
		InvitationID: "inv1",
		InviteeID:    "player2",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	mgr.SendInvitation(inv)

	err := mgr.AcceptInvitation("inv1")
	if err != nil {
		t.Errorf("AcceptInvitation failed: %v", err)
	}

	if !inv.Accepted {
		t.Error("Invitation not marked as accepted")
	}
	if inv.AcceptedAt.IsZero() {
		t.Error("AcceptedAt not set")
	}
}

func TestGuildInvitationManager_AcceptExpired(t *testing.T) {
	mgr := NewGuildInvitationManager()

	inv := &GuildInvitation{
		InvitationID: "inv1",
		InviteeID:    "player2",
		ExpiresAt:    time.Now().Add(-24 * time.Hour),
	}
	mgr.SendInvitation(inv)

	err := mgr.AcceptInvitation("inv1")
	if err == nil {
		t.Error("Expected error when accepting expired invitation")
	}
}

func TestGuildInvitationManager_CleanupExpired(t *testing.T) {
	mgr := NewGuildInvitationManager()

	inv1 := &GuildInvitation{
		InvitationID: "inv1",
		InviteeID:    "player2",
		ExpiresAt:    time.Now().Add(-24 * time.Hour),
	}
	inv2 := &GuildInvitation{
		InvitationID: "inv2",
		InviteeID:    "player2",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	mgr.SendInvitation(inv1)
	mgr.SendInvitation(inv2)

	removed := mgr.CleanupExpired()
	if removed != 1 {
		t.Errorf("Removed = %d, want 1", removed)
	}

	pending := mgr.GetPendingInvitations("player2")
	if len(pending) != 1 {
		t.Errorf("Pending invitations = %d, want 1", len(pending))
	}
}

// MountWhistleManager Tests

func TestMountWhistleManager_SummonMount(t *testing.T) {
	mgr := NewMountWhistleManager()

	summon := &MountSummon{
		PlayerID:   1,
		VehicleID:  100,
		CurrentPos: [2]float64{0, 0},
		TargetPos:  [2]float64{3, 4},
	}

	mgr.SummonMount(summon)

	if summon.Distance != 5.0 {
		t.Errorf("Distance = %f, want 5.0", summon.Distance)
	}
	if summon.EstimatedTime != 5.0 {
		t.Errorf("EstimatedTime = %f, want 5.0", summon.EstimatedTime)
	}
	if summon.Completed {
		t.Error("Summon should not be completed")
	}
}

func TestMountWhistleManager_GetActiveSummon(t *testing.T) {
	mgr := NewMountWhistleManager()

	summon := &MountSummon{
		PlayerID:   1,
		VehicleID:  100,
		CurrentPos: [2]float64{0, 0},
		TargetPos:  [2]float64{1, 0},
	}

	mgr.SummonMount(summon)

	active := mgr.GetActiveSummon(1)
	if active == nil {
		t.Error("Expected active summon")
	}
	if active.VehicleID != 100 {
		t.Errorf("VehicleID = %d, want 100", active.VehicleID)
	}
}

func TestMountWhistleManager_CompleteSummon(t *testing.T) {
	mgr := NewMountWhistleManager()

	summon := &MountSummon{
		PlayerID:   1,
		VehicleID:  100,
		CurrentPos: [2]float64{0, 0},
		TargetPos:  [2]float64{1, 0},
	}

	mgr.SummonMount(summon)
	mgr.CompleteSummon(1)

	active := mgr.GetActiveSummon(1)
	if !active.Completed {
		t.Error("Summon should be marked as completed")
	}
}

func TestMountWhistleManager_CancelSummon(t *testing.T) {
	mgr := NewMountWhistleManager()

	summon := &MountSummon{
		PlayerID:   1,
		VehicleID:  100,
		CurrentPos: [2]float64{0, 0},
		TargetPos:  [2]float64{1, 0},
	}

	mgr.SummonMount(summon)
	mgr.CancelSummon(1)

	active := mgr.GetActiveSummon(1)
	if active != nil {
		t.Error("Summon should be cancelled")
	}
}

// StorageSorter Tests

func TestStorageSorter_Presets(t *testing.T) {
	sorter := NewStorageSorter()

	presets := []string{"default", "rarity", "value"}
	for _, name := range presets {
		preset := sorter.GetPreset(name)
		if preset == nil {
			t.Errorf("Preset %s not found", name)
		}
	}
}

func TestStorageSorter_SortItems(t *testing.T) {
	sorter := NewStorageSorter()

	items := []*Item{
		{ID: "1", Name: "C", Type: "weapon", Rarity: 1, Value: 100, Quantity: 5},
		{ID: "2", Name: "A", Type: "armor", Rarity: 3, Value: 200, Quantity: 2},
		{ID: "3", Name: "B", Type: "weapon", Rarity: 2, Value: 150, Quantity: 10},
	}

	sorter.SortItems(items, SortByName)
	if items[0].Name != "A" {
		t.Error("Items not sorted by name correctly")
	}

	sorter.SortItems(items, SortByRarity)
	if items[0].Rarity != 3 {
		t.Error("Items not sorted by rarity correctly")
	}

	sorter.SortItems(items, SortByValue)
	if items[0].Value != 200 {
		t.Error("Items not sorted by value correctly")
	}

	sorter.SortItems(items, SortByQuantity)
	if items[0].Quantity != 10 {
		t.Error("Items not sorted by quantity correctly")
	}

	sorter.SortItems(items, SortByType)
	if items[0].Type != "armor" {
		t.Error("Items not sorted by type correctly")
	}
}

func TestStorageSorter_AddPreset(t *testing.T) {
	sorter := NewStorageSorter()

	preset := &StorageSortPreset{
		Name:              "custom",
		PrimaryCriteria:   SortByName,
		SecondaryCriteria: SortByValue,
		Descending:        false,
		GroupByType:       true,
	}

	sorter.AddPreset(preset)

	got := sorter.GetPreset("custom")
	if got == nil || got.Name != "custom" {
		t.Error("Custom preset not added correctly")
	}
}

// RecipeTracker Tests

func TestRecipeTracker_TrackRecipe(t *testing.T) {
	tracker := NewRecipeTracker()

	info := &RecipeTrackingInfo{
		RecipeID:      "iron_sword",
		RecipeName:    "Iron Sword",
		RequiredMats:  map[string]int{"iron": 3, "wood": 1},
		AvailableMats: map[string]int{"iron": 5, "wood": 2},
	}

	tracker.TrackRecipe(1, info)

	if !info.CanCraft {
		t.Error("Should be craftable with sufficient materials")
	}
	if info.MaxCraftable != 1 {
		t.Errorf("MaxCraftable = %d, want 1", info.MaxCraftable)
	}
}

func TestRecipeTracker_MissingMaterials(t *testing.T) {
	tracker := NewRecipeTracker()

	info := &RecipeTrackingInfo{
		RecipeID:      "iron_sword",
		RecipeName:    "Iron Sword",
		RequiredMats:  map[string]int{"iron": 3, "wood": 1},
		AvailableMats: map[string]int{"iron": 1, "wood": 2},
	}

	tracker.TrackRecipe(1, info)

	if info.CanCraft {
		t.Error("Should not be craftable with insufficient materials")
	}
	if info.MaxCraftable != 0 {
		t.Errorf("MaxCraftable = %d, want 0", info.MaxCraftable)
	}
	if info.MissingMats["iron"] != 2 {
		t.Errorf("Missing iron = %d, want 2", info.MissingMats["iron"])
	}
}

func TestRecipeTracker_UpdateMaterialAvailability(t *testing.T) {
	tracker := NewRecipeTracker()

	info := &RecipeTrackingInfo{
		RecipeID:      "iron_sword",
		RecipeName:    "Iron Sword",
		RequiredMats:  map[string]int{"iron": 3},
		AvailableMats: map[string]int{"iron": 1},
	}

	tracker.TrackRecipe(1, info)

	if info.CanCraft {
		t.Error("Should not be craftable initially")
	}

	tracker.UpdateMaterialAvailability(1, "iron_sword", map[string]int{"iron": 6})

	if !info.CanCraft {
		t.Error("Should be craftable after update")
	}
	if info.MaxCraftable != 2 {
		t.Errorf("MaxCraftable = %d, want 2", info.MaxCraftable)
	}
}

func TestRecipeTracker_UntrackRecipe(t *testing.T) {
	tracker := NewRecipeTracker()

	info := &RecipeTrackingInfo{
		RecipeID:      "iron_sword",
		RecipeName:    "Iron Sword",
		RequiredMats:  map[string]int{"iron": 3},
		AvailableMats: map[string]int{"iron": 5},
	}

	tracker.TrackRecipe(1, info)
	tracker.UntrackRecipe(1, "iron_sword")

	recipes := tracker.GetTrackedRecipes(1)
	if len(recipes) != 0 {
		t.Errorf("Tracked recipes = %d, want 0", len(recipes))
	}
}

func TestRecipeTracker_GetTrackedRecipes(t *testing.T) {
	tracker := NewRecipeTracker()

	t.Run("unknown player returns empty slice", func(t *testing.T) {
		recipes := tracker.GetTrackedRecipes(999)
		if recipes == nil {
			t.Error("Expected non-nil empty slice, got nil")
		}
		if len(recipes) != 0 {
			t.Errorf("Expected empty slice for unknown player, got %d items", len(recipes))
		}
	})

	t.Run("multiple recipes tracked", func(t *testing.T) {
		info1 := &RecipeTrackingInfo{
			RecipeID:      "iron_sword",
			RecipeName:    "Iron Sword",
			RequiredMats:  map[string]int{"iron": 3},
			AvailableMats: map[string]int{"iron": 5},
		}
		info2 := &RecipeTrackingInfo{
			RecipeID:      "steel_axe",
			RecipeName:    "Steel Axe",
			RequiredMats:  map[string]int{"steel": 2, "wood": 1},
			AvailableMats: map[string]int{"steel": 4, "wood": 3},
		}
		info3 := &RecipeTrackingInfo{
			RecipeID:      "leather_armor",
			RecipeName:    "Leather Armor",
			RequiredMats:  map[string]int{"leather": 5},
			AvailableMats: map[string]int{"leather": 3},
		}

		tracker.TrackRecipe(2, info1)
		tracker.TrackRecipe(2, info2)
		tracker.TrackRecipe(2, info3)

		recipes := tracker.GetTrackedRecipes(2)
		if len(recipes) != 3 {
			t.Errorf("Expected 3 tracked recipes, got %d", len(recipes))
		}

		recipeIDs := make(map[string]bool)
		for _, r := range recipes {
			recipeIDs[r.RecipeID] = true
		}
		if !recipeIDs["iron_sword"] || !recipeIDs["steel_axe"] || !recipeIDs["leather_armor"] {
			t.Error("Not all tracked recipes returned")
		}
	})
}

func TestStorageSorter_GetPreset(t *testing.T) {
	sorter := NewStorageSorter()

	t.Run("default presets exist", func(t *testing.T) {
		defaultPreset := sorter.GetPreset("default")
		if defaultPreset == nil {
			t.Error("Expected 'default' preset to exist")
		}
		if defaultPreset.Name != "Default" {
			t.Errorf("Default preset name = %q, want %q", defaultPreset.Name, "Default")
		}
		if defaultPreset.PrimaryCriteria != SortByType {
			t.Error("Default preset primary criteria should be SortByType")
		}

		rarityPreset := sorter.GetPreset("rarity")
		if rarityPreset == nil {
			t.Error("Expected 'rarity' preset to exist")
		}
		if !rarityPreset.Descending {
			t.Error("Rarity preset should have Descending=true")
		}

		valuePreset := sorter.GetPreset("value")
		if valuePreset == nil {
			t.Error("Expected 'value' preset to exist")
		}
	})

	t.Run("non-existent preset returns nil", func(t *testing.T) {
		preset := sorter.GetPreset("nonexistent")
		if preset != nil {
			t.Error("Expected nil for non-existent preset")
		}
	})

	t.Run("custom preset retrieval", func(t *testing.T) {
		customPreset := &StorageSortPreset{
			Name:              "my_preset",
			PrimaryCriteria:   SortByQuantity,
			SecondaryCriteria: SortByRarity,
			Descending:        true,
			GroupByType:       false,
		}
		sorter.AddPreset(customPreset)

		retrieved := sorter.GetPreset("my_preset")
		if retrieved == nil {
			t.Error("Expected custom preset to be retrievable")
		}
		if retrieved.PrimaryCriteria != SortByQuantity {
			t.Error("Custom preset primary criteria mismatch")
		}
		if !retrieved.Descending {
			t.Error("Custom preset Descending should be true")
		}
	})
}

// Benchmarks

func BenchmarkAutoLootManager_ShouldCollect(b *testing.B) {
	mgr := NewAutoLootManager()
	config := DefaultAutoLootConfig(1)
	mgr.SetConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.ShouldCollect(1, 2, "weapon")
	}
}

func BenchmarkCraftQueueManager_AddRecipe(b *testing.B) {
	mgr := NewCraftQueueManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use unique player IDs to avoid triggering "queue full" warnings
		// (each player has a 50-recipe limit)
		playerID := uint64(i/50 + 1)
		mgr.AddRecipe(playerID, fmt.Sprintf("recipe_%d", i%10), 1)
	}
}

func BenchmarkStorageSorter_SortItems(b *testing.B) {
	sorter := NewStorageSorter()
	items := make([]*Item, 100)
	for i := 0; i < 100; i++ {
		items[i] = &Item{
			ID:       fmt.Sprintf("item_%d", i),
			Name:     fmt.Sprintf("Item %d", i),
			Rarity:   i % 5,
			Value:    i * 10,
			Quantity: i,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sorter.SortItems(items, SortByRarity)
	}
}

func BenchmarkRecipeTracker_TrackRecipe(b *testing.B) {
	tracker := NewRecipeTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info := &RecipeTrackingInfo{
			RecipeID:      fmt.Sprintf("recipe_%d", i%10),
			RequiredMats:  map[string]int{"iron": 3, "wood": 1},
			AvailableMats: map[string]int{"iron": 5, "wood": 2},
		}
		tracker.TrackRecipe(uint64(i%5), info)
	}
}
