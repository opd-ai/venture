// Package engine provides the collection system for the ECS.
// This file contains tests for the CollectionSystem.
// Phase 97: Collection System (V18.0)

package engine

import (
	"sync"
	"testing"
)

// testPlayerComponent is a stub component for marking entities as players in tests.
type testPlayerComponent struct{}

func (p *testPlayerComponent) Type() string { return "player" }

func TestNewCollectionSystem(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	if system == nil {
		t.Fatal("NewCollectionSystem returned nil")
	}

	if system.world != world {
		t.Error("World reference not set correctly")
	}

	milestones := system.GetMilestones()
	if len(milestones) != 4 {
		t.Errorf("Default milestones count = %d, want 4", len(milestones))
	}
}

func TestCollectionSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player entity with collection component
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(NewCollectionComponent())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(NewExperienceComponent()) // Has Level field

	// Create a collectible entity nearby
	collectible := world.CreateEntity()
	collectible.AddComponent(NewCollectibleComponent(
		"art_test_item",
		"Test Artifact",
		CollectionCategoryArtifacts,
		CollectionRarityRare,
		"A test collectible item",
	))
	collectible.AddComponent(&PositionComponent{X: 110, Y: 100}) // Within range

	// Flush entities to world
	world.FlushPendingEntities()

	// Run update
	entities := []*Entity{player, collectible}
	system.Update(entities, 0.016)

	// Check if collectible was picked up
	comp, ok := player.GetComponent("collection")
	if !ok {
		t.Fatal("Failed to get collection component")
	}
	collection := comp.(*CollectionComponent)
	if !collection.HasCollectible("art_test_item") {
		t.Error("Collectible should have been picked up")
	}

	// Check collectible is marked as collected
	collComp, ok := collectible.GetComponent("collectible")
	if !ok {
		t.Fatal("Failed to get collectible component")
	}
	collectable := collComp.(*CollectibleComponent)
	if !collectable.IsAlreadyCollected() {
		t.Error("Collectible component should be marked as collected")
	}
}

func TestCollectionSystem_UpdateOutOfRange(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player entity
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(NewCollectionComponent())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(NewExperienceComponent())

	// Create a collectible entity far away
	collectible := world.CreateEntity()
	collectible.AddComponent(NewCollectibleComponent(
		"art_far_item",
		"Far Artifact",
		CollectionCategoryArtifacts,
		CollectionRarityCommon,
		"",
	))
	collectible.AddComponent(&PositionComponent{X: 200, Y: 200}) // Out of range

	// Flush entities to world
	world.FlushPendingEntities()

	// Run update
	entities := []*Entity{player, collectible}
	system.Update(entities, 0.016)

	// Check if collectible was NOT picked up
	comp, _ := player.GetComponent("collection")
	collection := comp.(*CollectionComponent)
	if collection.HasCollectible("art_far_item") {
		t.Error("Collectible should NOT have been picked up (out of range)")
	}
}

func TestCollectionSystem_LevelRequirement(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a low-level player
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(NewCollectionComponent())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	exp := NewExperienceComponent()
	exp.Level = 5 // Low level
	player.AddComponent(exp)

	// Create a high-level collectible
	collectible := world.CreateEntity()
	collComp := NewCollectibleComponent(
		"art_high_level",
		"High Level Artifact",
		CollectionCategoryArtifacts,
		CollectionRarityEpic,
		"",
	)
	collComp.SetRequiredLevel(20)
	collectible.AddComponent(collComp)
	collectible.AddComponent(&PositionComponent{X: 105, Y: 100}) // In range

	// Flush entities to world
	world.FlushPendingEntities()

	// Run update
	entities := []*Entity{player, collectible}
	system.Update(entities, 0.016)

	// Check if collectible was NOT picked up (level too low)
	comp, _ := player.GetComponent("collection")
	collection := comp.(*CollectionComponent)
	if collection.HasCollectible("art_high_level") {
		t.Error("Collectible should NOT have been picked up (level requirement)")
	}
}

func TestCollectionSystem_FishingSync(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player with fishing and collection components
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	fishing := NewFishingComponent()
	fishing.TotalCaught["bass_001"] = 1
	fishing.TotalCaught["trout_002"] = 3
	player.AddComponent(fishing)
	player.AddComponent(NewCollectionComponent())

	// Flush entities to world
	world.FlushPendingEntities()

	// Run update
	entities := []*Entity{player}
	system.Update(entities, 0.016)

	// Check if fish were synced to collection
	comp, _ := player.GetComponent("collection")
	collection := comp.(*CollectionComponent)
	if !collection.HasCollectible("bass_001") {
		t.Error("bass_001 should have been synced to collection")
	}
	if !collection.HasCollectible("trout_002") {
		t.Error("trout_002 should have been synced to collection")
	}
}

func TestCollectionSystem_GatheringSync(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player with gathering and collection components
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	gathering := NewGatheringComponent()
	gathering.TotalHarvested[ResourceTypeOre] = 5
	gathering.TotalHarvested[ResourceTypeGem] = 2
	player.AddComponent(gathering)
	player.AddComponent(NewCollectionComponent())

	// Flush entities to world
	world.FlushPendingEntities()

	// Run update
	entities := []*Entity{player}
	system.Update(entities, 0.016)

	// Check if resources were synced to collection
	comp, _ := player.GetComponent("collection")
	collection := comp.(*CollectionComponent)
	if !collection.HasCollectible(string(ResourceTypeOre)) {
		t.Error("Ore should have been synced to collection")
	}
	if !collection.HasCollectible(string(ResourceTypeGem)) {
		t.Error("Gem should have been synced to collection")
	}

	// Check rarities
	oreEntry := collection.GetCollectible(string(ResourceTypeOre))
	if oreEntry.Rarity != CollectionRarityCommon {
		t.Errorf("Ore rarity = %v, want common", oreEntry.Rarity)
	}

	gemEntry := collection.GetCollectible(string(ResourceTypeGem))
	if gemEntry.Rarity != CollectionRarityRare {
		t.Errorf("Gem rarity = %v, want rare", gemEntry.Rarity)
	}
}

func TestCollectionSystem_DisableAutoSync(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Disable auto sync
	system.SetAutoRegisterFishing(false)
	system.SetAutoRegisterGathering(false)

	// Create a player with fishing/gathering
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	fishing := NewFishingComponent()
	fishing.TotalCaught["fish_001"] = 1
	player.AddComponent(fishing)
	gathering := NewGatheringComponent()
	gathering.TotalHarvested[ResourceTypeWood] = 3
	player.AddComponent(gathering)
	player.AddComponent(NewCollectionComponent())

	// Flush entities to world
	world.FlushPendingEntities()

	// Run update
	entities := []*Entity{player}
	system.Update(entities, 0.016)

	// Check that nothing was synced
	comp, _ := player.GetComponent("collection")
	collection := comp.(*CollectionComponent)
	if collection.HasCollectible("fish_001") {
		t.Error("Fishing sync should be disabled")
	}
	if collection.HasCollectible(string(ResourceTypeWood)) {
		t.Error("Gathering sync should be disabled")
	}
}

func TestCollectionSystem_RegisterCollectible(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(NewCollectionComponent())

	// Flush entities to world
	world.FlushPendingEntities()

	// Manually register a collectible
	isNew := system.RegisterCollectible(
		player.ID,
		"lore_001",
		"Ancient Scroll",
		CollectionCategoryLore,
		CollectionRarityUncommon,
		"An old scroll with mysterious text",
	)

	if !isNew {
		t.Error("RegisterCollectible should return true for new item")
	}

	// Verify it was added
	comp, _ := player.GetComponent("collection")
	collection := comp.(*CollectionComponent)
	if !collection.HasCollectible("lore_001") {
		t.Error("Manually registered collectible should exist")
	}

	// Register same item again
	isNew = system.RegisterCollectible(
		player.ID,
		"lore_001",
		"Ancient Scroll",
		CollectionCategoryLore,
		CollectionRarityUncommon,
		"",
	)

	if isNew {
		t.Error("RegisterCollectible should return false for existing item")
	}
}

func TestCollectionSystem_GetCollectionProgress(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	collection := NewCollectionComponent()
	player.AddComponent(collection)

	// Flush entities to world
	world.FlushPendingEntities()

	// Add some collectibles
	collection.AddCollectible("fish_001", "Fish", CollectionCategoryFish, CollectionRarityCommon, "", 1000)
	collection.AddCollectible("art_001", "Art", CollectionCategoryArtifacts, CollectionRarityRare, "", 1001)

	discovered, total, percent := system.GetCollectionProgress(player.ID)

	if discovered != 2 {
		t.Errorf("Discovered = %d, want 2", discovered)
	}

	if total <= 0 {
		t.Errorf("Total = %d, should be > 0", total)
	}

	if percent <= 0 {
		t.Errorf("Percent = %f, should be > 0", percent)
	}
}

func TestCollectionSystem_GetCategoryProgress(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	collection := NewCollectionComponent()
	collection.SetCategoryTotal(CollectionCategoryFish, 10)
	player.AddComponent(collection)

	// Flush entities to world
	world.FlushPendingEntities()

	// Add some fish
	collection.AddCollectible("fish_001", "Fish 1", CollectionCategoryFish, CollectionRarityCommon, "", 1000)
	collection.AddCollectible("fish_002", "Fish 2", CollectionCategoryFish, CollectionRarityCommon, "", 1001)

	discovered, total, percent := system.GetCategoryProgress(player.ID, CollectionCategoryFish)

	if discovered != 2 {
		t.Errorf("Fish discovered = %d, want 2", discovered)
	}

	if total != 10 {
		t.Errorf("Fish total = %d, want 10", total)
	}

	if percent != 20.0 {
		t.Errorf("Fish percent = %f, want 20.0", percent)
	}
}

func TestCollectionSystem_Callbacks(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	var discoveredItems []string
	var milestoneReached []string

	// Register callbacks
	system.RegisterDiscoveryCallback(func(entityID uint64, entry *CollectedEntry) {
		discoveredItems = append(discoveredItems, entry.ID)
	})

	system.RegisterRewardCallback(func(entityID uint64, milestone CollectionMilestone, category CollectionCategory) {
		milestoneReached = append(milestoneReached, milestone.Name)
	})

	// Create a player
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	collection := NewCollectionComponent()
	collection.SetCategoryTotal(CollectionCategoryArtifacts, 4) // Small total for milestone testing
	player.AddComponent(collection)

	// Flush entities to world
	world.FlushPendingEntities()

	// Register a collectible (should trigger discovery callback)
	system.RegisterCollectible(
		player.ID,
		"art_001",
		"Artifact 1",
		CollectionCategoryArtifacts,
		CollectionRarityCommon,
		"",
	)

	if len(discoveredItems) != 1 || discoveredItems[0] != "art_001" {
		t.Errorf("Discovery callback not triggered correctly: %v", discoveredItems)
	}

	// Milestone should be reached (25% with 1/4)
	if len(milestoneReached) != 1 {
		t.Errorf("Milestone callback triggered %d times, want 1", len(milestoneReached))
	}
}

func TestCollectionSystem_ExportImport(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Create a player with some collection data
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	collection := NewCollectionComponent()
	collection.AddCollectible("fish_001", "Fish", CollectionCategoryFish, CollectionRarityRare, "", 1000)
	collection.AddCollectible("art_001", "Artifact", CollectionCategoryArtifacts, CollectionRarityLegendary, "", 1001)
	collection.AddFavorite("fish_001")
	player.AddComponent(collection)

	// Flush entities to world
	world.FlushPendingEntities()

	// Export
	data, err := system.ExportCollectionData(player.ID)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Create new player and import
	player2 := world.CreateEntity()
	player2.AddComponent(&testPlayerComponent{})
	player2.AddComponent(NewCollectionComponent())

	// Flush entities to world
	world.FlushPendingEntities()

	err = system.ImportCollectionData(player2.ID, data)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify imported data
	comp, _ := player2.GetComponent("collection")
	imported := comp.(*CollectionComponent)
	if !imported.HasCollectible("fish_001") {
		t.Error("Imported collection missing fish_001")
	}
	if !imported.HasCollectible("art_001") {
		t.Error("Imported collection missing art_001")
	}
	if !imported.IsFavorite("fish_001") {
		t.Error("Imported collection missing favorite")
	}
}

func TestCollectionSystem_SetMilestones(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	customMilestones := []CollectionMilestone{
		{Threshold: 10, Name: "Beginner", RewardType: "title", RewardID: "t1", Points: 10},
		{Threshold: 50, Name: "Expert", RewardType: "cosmetic", RewardID: "c1", Points: 100},
	}

	system.SetMilestones(customMilestones)

	milestones := system.GetMilestones()
	if len(milestones) != 2 {
		t.Errorf("Milestones count = %d, want 2", len(milestones))
	}

	if milestones[0].Threshold != 10 || milestones[0].Name != "Beginner" {
		t.Errorf("First milestone = %v, want {10, Beginner, ...}", milestones[0])
	}
}

func TestCollectionSystem_GetFishRarity(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	tests := []struct {
		fishTypeID string
		want       CollectionRarity
	}{
		{"bass_001", CollectionRarityCommon},
		{"trout_002", CollectionRarityUncommon},
		{"salmon_003", CollectionRarityRare},
		{"fish", CollectionRarityCommon},
	}

	for _, tt := range tests {
		t.Run(tt.fishTypeID, func(t *testing.T) {
			got := system.getFishRarity(tt.fishTypeID)
			if got != tt.want {
				t.Errorf("getFishRarity(%s) = %v, want %v", tt.fishTypeID, got, tt.want)
			}
		})
	}
}

func TestCollectionSystem_GetResourceRarity(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	tests := []struct {
		resourceType ResourceType
		want         CollectionRarity
	}{
		{ResourceTypeOre, CollectionRarityCommon},
		{ResourceTypeWood, CollectionRarityCommon},
		{ResourceTypeFiber, CollectionRarityCommon},
		{ResourceTypeHerb, CollectionRarityUncommon},
		{ResourceTypeGem, CollectionRarityRare},
		{ResourceTypeEssence, CollectionRarityEpic},
	}

	for _, tt := range tests {
		t.Run(string(tt.resourceType), func(t *testing.T) {
			got := system.getResourceRarity(tt.resourceType)
			if got != tt.want {
				t.Errorf("getResourceRarity(%s) = %v, want %v", tt.resourceType, got, tt.want)
			}
		})
	}
}

func TestCollectionSystem_NilWorld(t *testing.T) {
	system := NewCollectionSystem(nil)

	// Should not panic with nil world
	discovered, total, percent := system.GetCollectionProgress(123)
	if discovered != 0 || total != 0 || percent != 0 {
		t.Error("Should return zeros for nil world")
	}

	isNew := system.RegisterCollectible(123, "test", "Test", CollectionCategoryArtifacts, CollectionRarityCommon, "")
	if isNew {
		t.Error("Should return false for nil world")
	}

	data, err := system.ExportCollectionData(123)
	if data != nil || err != nil {
		t.Error("Should return nil for nil world")
	}
}

func TestCollectionSystem_NonExistentEntity(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	// Try to get progress for non-existent entity
	discovered, total, percent := system.GetCollectionProgress(99999)
	if discovered != 0 || total != 0 || percent != 0 {
		t.Error("Should return zeros for non-existent entity")
	}

	// Try to register for non-existent entity
	isNew := system.RegisterCollectible(99999, "test", "Test", CollectionCategoryArtifacts, CollectionRarityCommon, "")
	if isNew {
		t.Error("Should return false for non-existent entity")
	}
}

func TestCollectionSystem_ConcurrentCallbacks(t *testing.T) {
	world := NewWorld()
	system := NewCollectionSystem(world)

	var mu sync.Mutex
	count := 0

	// Register callback
	system.RegisterDiscoveryCallback(func(entityID uint64, entry *CollectedEntry) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(NewCollectionComponent())

	// Flush entities to world
	world.FlushPendingEntities()

	// Concurrent registrations
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				id := string(rune('a'+idx)) + string(rune('0'+j))
				system.RegisterCollectible(
					player.ID,
					id,
					"Item",
					CollectionCategoryArtifacts,
					CollectionRarityCommon,
					"",
				)
			}
		}(i)
	}

	wg.Wait()

	mu.Lock()
	finalCount := count
	mu.Unlock()

	// All items should be unique, so all should trigger callbacks
	if finalCount != 100 {
		t.Errorf("Callback count = %d, want 100", finalCount)
	}
}
