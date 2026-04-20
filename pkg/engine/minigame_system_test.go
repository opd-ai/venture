package engine

import (
	"fmt"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewMiniGameSystem(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	if sys == nil {
		t.Fatal("NewMiniGameSystem returned nil")
	}
	if sys.world != world {
		t.Error("MiniGameSystem.world not set correctly")
	}
	if sys.rngSource == nil {
		t.Error("MiniGameSystem.rngSource should be initialized")
	}
}

func TestMiniGameSystem_SetSeed(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Set a specific seed
	sys.SetSeed(12345)

	// Generate rewards and verify they're deterministic
	reward1 := sys.generateReward(MiniGameCard, 0.5)
	sys.SetSeed(12345) // Reset to same seed
	reward2 := sys.generateReward(MiniGameCard, 0.5)

	if reward1.Gold != reward2.Gold {
		t.Errorf("Rewards not deterministic: gold %d != %d", reward1.Gold, reward2.Gold)
	}
	if reward1.XP != reward2.XP {
		t.Errorf("Rewards not deterministic: xp %f != %f", reward1.XP, reward2.XP)
	}
}

func TestMiniGameSystem_StartGame(t *testing.T) {
	tests := []struct {
		name       string
		gameType   MiniGameType
		difficulty float64
		wantErr    bool
	}{
		{"valid card game", MiniGameCard, 0.5, false},
		{"valid dice game", MiniGameDice, 0.3, false},
		{"valid puzzle", MiniGamePuzzle, 0.8, false},
		{"zero difficulty", MiniGameMemory, 0.0, false},
		{"max difficulty", MiniGameLockPicking, 1.0, false},
		{"negative difficulty (clamped)", MiniGameHacking, -0.5, false},
		{"high difficulty (clamped)", MiniGameRitual, 2.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewMiniGameSystem(world)
			entity := world.CreateEntity()
			world.Update(0) // Process entity addition

			err := sys.StartGame(entity.ID, tt.gameType, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("StartGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify component was added
				comp, ok := entity.GetComponent("minigame")
				if !ok {
					t.Fatal("minigame component not added to entity")
				}

				gameComp := comp.(*MiniGameComponent)
				if gameComp.GameType != tt.gameType {
					t.Errorf("GameType = %v, want %v", gameComp.GameType, tt.gameType)
				}
				if !gameComp.Active {
					t.Error("Game should be active after start")
				}

				// Check difficulty clamping
				expectedDiff := tt.difficulty
				if expectedDiff < 0.0 {
					expectedDiff = 0.0
				}
				if expectedDiff > 1.0 {
					expectedDiff = 1.0
				}
				if gameComp.Difficulty != expectedDiff {
					t.Errorf("Difficulty = %v, want %v", gameComp.Difficulty, expectedDiff)
				}

				if gameComp.TimeElapsed != 0 {
					t.Errorf("TimeElapsed = %v, want 0", gameComp.TimeElapsed)
				}
				if gameComp.Reward == nil {
					t.Error("Reward should be generated")
				}
			}
		})
	}
}

func TestMiniGameSystem_StartGame_InvalidEntity(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	err := sys.StartGame(999, MiniGameCard, 0.5)
	if err == nil {
		t.Error("StartGame with invalid entity should return error")
	}
}

func TestMiniGameSystem_Update_TimeElapsed(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	err := sys.StartGame(entity.ID, MiniGameCard, 0.5)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Update with deltaTime
	sys.Update(nil, 1.0)

	comp, _ := entity.GetComponent("minigame")
	gameComp := comp.(*MiniGameComponent)

	if gameComp.TimeElapsed != 1.0 {
		t.Errorf("TimeElapsed = %v, want 1.0", gameComp.TimeElapsed)
	}

	// Update again
	sys.Update(nil, 0.5)
	if gameComp.TimeElapsed != 1.5 {
		t.Errorf("TimeElapsed = %v, want 1.5", gameComp.TimeElapsed)
	}
}

func TestMiniGameSystem_Update_Timeout(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Start a lock-picking game (2 minute timeout)
	err := sys.StartGame(entity.ID, MiniGameLockPicking, 0.5)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Add inventory for reward tracking
	entity.AddComponent(&InventoryComponent{Gold: 0})

	// Update to just before timeout
	sys.Update(nil, 119.0)

	comp, _ := entity.GetComponent("minigame")
	gameComp := comp.(*MiniGameComponent)
	if !gameComp.Active {
		t.Error("Game should still be active before timeout")
	}

	// Update past timeout
	sys.Update(nil, 2.0) // Total: 121 seconds > 120 second limit

	if gameComp.Active {
		t.Error("Game should be inactive after timeout")
	}

	// Verify no reward was given (timeout counts as failure)
	invComp, _ := entity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 0 {
		t.Errorf("Gold should be 0 after timeout, got %d", inv.Gold)
	}
}

func TestMiniGameSystem_Update_WithGameInstance(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Start game
	err := sys.StartGame(entity.ID, MiniGamePuzzle, 0.5)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Create a stub game instance
	stub := &StubMiniGame{
		isComplete: false,
		reward:     &Reward{Gold: 100, XP: 50.0},
	}

	// Set the game instance
	err = sys.SetGameInstance(entity.ID, stub)
	if err != nil {
		t.Fatalf("SetGameInstance failed: %v", err)
	}

	// Add inventory and experience components
	entity.AddComponent(&InventoryComponent{Gold: 0})
	entity.AddComponent(&ExperienceComponent{CurrentXP: 0})

	// Update should call the game instance
	sys.Update(nil, 1.0)

	comp, _ := entity.GetComponent("minigame")
	gameComp := comp.(*MiniGameComponent)
	if !gameComp.Active {
		t.Error("Game should still be active when not complete")
	}

	// Mark game as complete
	stub.isComplete = true

	// Update again
	sys.Update(nil, 1.0)

	// Game should be ended and reward awarded
	if gameComp.Active {
		t.Error("Game should be inactive after completion")
	}

	invComp, _ := entity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 100 {
		t.Errorf("Gold = %d, want 100", inv.Gold)
	}

	expComp, _ := entity.GetComponent("experience")
	exp := expComp.(*ExperienceComponent)
	if exp.CurrentXP != 50 {
		t.Errorf("CurrentXP = %d, want 50", exp.CurrentXP)
	}
}

func TestMiniGameSystem_Update_WithGameInstanceError(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	err := sys.StartGame(entity.ID, MiniGamePuzzle, 0.5)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Create a stub that returns an error
	stub := &StubMiniGame{
		updateErr: fmt.Errorf("game error"),
	}

	err = sys.SetGameInstance(entity.ID, stub)
	if err != nil {
		t.Fatalf("SetGameInstance failed: %v", err)
	}

	// Update should handle the error and end the game
	sys.Update(nil, 1.0)

	comp, _ := entity.GetComponent("minigame")
	gameComp := comp.(*MiniGameComponent)
	if gameComp.Active {
		t.Error("Game should be ended after error")
	}
}

func TestMiniGameSystem_EndGame(t *testing.T) {
	tests := []struct {
		name        string
		success     bool
		wantGold    int
		wantXP      int
		initialGold int
		initialXP   int
	}{
		{"success awards reward", true, 100, 50, 0, 0},
		{"failure no reward", false, 0, 0, 0, 0},
		{"success adds to existing", true, 150, 75, 50, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewMiniGameSystem(world)
			entity := world.CreateEntity()
			world.Update(0) // Process entity addition

			// Add components
			entity.AddComponent(&InventoryComponent{Gold: tt.initialGold})
			entity.AddComponent(&ExperienceComponent{CurrentXP: tt.initialXP})

			// Start game
			err := sys.StartGame(entity.ID, MiniGameCard, 0.5)
			if err != nil {
				t.Fatalf("StartGame failed: %v", err)
			}

			// Set a known reward
			comp, _ := entity.GetComponent("minigame")
			gameComp := comp.(*MiniGameComponent)
			gameComp.Reward = &Reward{Gold: 100, XP: 50.0}

			// End the game
			err = sys.EndGame(entity.ID, tt.success)
			if err != nil {
				t.Fatalf("EndGame failed: %v", err)
			}

			// Verify game is inactive
			if gameComp.Active {
				t.Error("Game should be inactive after end")
			}

			// Verify rewards
			invComp, _ := entity.GetComponent("inventory")
			inv := invComp.(*InventoryComponent)
			if inv.Gold != tt.wantGold {
				t.Errorf("Gold = %d, want %d", inv.Gold, tt.wantGold)
			}

			expComp, _ := entity.GetComponent("experience")
			exp := expComp.(*ExperienceComponent)
			if exp.CurrentXP != tt.wantXP {
				t.Errorf("CurrentXP = %d, want %d", exp.CurrentXP, tt.wantXP)
			}
		})
	}
}

func TestMiniGameSystem_EndGame_InvalidEntity(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	err := sys.EndGame(999, true)
	if err == nil {
		t.Error("EndGame with invalid entity should return error")
	}
}

func TestMiniGameSystem_EndGame_NoComponent(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	err := sys.EndGame(entity.ID, true)
	if err == nil {
		t.Error("EndGame without minigame component should return error")
	}
}

func TestMiniGameSystem_IsGameActive(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Not active initially
	if sys.IsGameActive(entity.ID) {
		t.Error("Game should not be active initially")
	}

	// Start game
	sys.StartGame(entity.ID, MiniGameCard, 0.5)
	if !sys.IsGameActive(entity.ID) {
		t.Error("Game should be active after start")
	}

	// End game
	sys.EndGame(entity.ID, false)
	if sys.IsGameActive(entity.ID) {
		t.Error("Game should not be active after end")
	}

	// Invalid entity
	if sys.IsGameActive(999) {
		t.Error("Invalid entity should not be active")
	}
}

func TestMiniGameSystem_GetGameComponent(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// No component initially
	comp := sys.GetGameComponent(entity.ID)
	if comp != nil {
		t.Error("GetGameComponent should return nil when no component exists")
	}

	// Add component
	sys.StartGame(entity.ID, MiniGameCard, 0.5)
	comp = sys.GetGameComponent(entity.ID)
	if comp == nil {
		t.Fatal("GetGameComponent should return component after start")
	}
	if comp.GameType != MiniGameCard {
		t.Errorf("GameType = %v, want %v", comp.GameType, MiniGameCard)
	}

	// Invalid entity
	comp = sys.GetGameComponent(999)
	if comp != nil {
		t.Error("GetGameComponent should return nil for invalid entity")
	}
}

func TestMiniGameSystem_SetGameInstance(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Start game
	sys.StartGame(entity.ID, MiniGamePuzzle, 0.5)

	// Create stub instance
	stub := &StubMiniGame{}

	// Set instance
	err := sys.SetGameInstance(entity.ID, stub)
	if err != nil {
		t.Errorf("SetGameInstance failed: %v", err)
	}

	// Verify instance was set
	comp := sys.GetGameComponent(entity.ID)
	if comp.GameInstance != stub {
		t.Error("GameInstance not set correctly")
	}
}

func TestMiniGameSystem_SetGameInstance_Errors(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	stub := &StubMiniGame{}

	// Invalid entity
	err := sys.SetGameInstance(999, stub)
	if err == nil {
		t.Error("SetGameInstance should fail for invalid entity")
	}

	// Entity without component
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	err = sys.SetGameInstance(entity.ID, stub)
	if err == nil {
		t.Error("SetGameInstance should fail when entity has no component")
	}
}

func TestMiniGameSystem_GetTimeLimit(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	tests := []struct {
		gameType MiniGameType
		want     float64
	}{
		{MiniGameCard, 600.0},
		{MiniGameDice, 300.0},
		{MiniGamePuzzle, 420.0},
		{MiniGameMemory, 240.0},
		{MiniGameLockPicking, 120.0},
		{MiniGameHacking, 180.0},
		{MiniGameRitual, 300.0},
		{MiniGameType(999), 300.0}, // Unknown defaults to 300
	}

	for _, tt := range tests {
		t.Run(tt.gameType.String(), func(t *testing.T) {
			got := sys.getTimeLimit(tt.gameType)
			if got != tt.want {
				t.Errorf("getTimeLimit(%v) = %v, want %v", tt.gameType, got, tt.want)
			}
		})
	}
}

func TestMiniGameSystem_GenerateReward_Deterministic(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Set seed for determinism
	sys.SetSeed(12345)

	// Generate reward
	reward1 := sys.generateReward(MiniGameCard, 0.5)

	// Reset seed and generate again
	sys.SetSeed(12345)
	reward2 := sys.generateReward(MiniGameCard, 0.5)

	// Should be identical
	if reward1.Gold != reward2.Gold {
		t.Errorf("Rewards not deterministic: gold %d != %d", reward1.Gold, reward2.Gold)
	}
	if reward1.XP != reward2.XP {
		t.Errorf("Rewards not deterministic: xp %f != %f", reward1.XP, reward2.XP)
	}
}

func TestMiniGameSystem_GenerateReward_DifficultyScaling(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetSeed(12345)

	// Generate rewards at different difficulties
	rewardEasy := sys.generateReward(MiniGameCard, 0.2)
	sys.SetSeed(12345)
	rewardHard := sys.generateReward(MiniGameCard, 0.8)

	// Higher difficulty should give more rewards
	if rewardHard.Gold <= rewardEasy.Gold {
		t.Errorf("Hard difficulty gold (%d) should be > easy (%d)", rewardHard.Gold, rewardEasy.Gold)
	}
	if rewardHard.XP <= rewardEasy.XP {
		t.Errorf("Hard difficulty XP (%.1f) should be > easy (%.1f)", rewardHard.XP, rewardEasy.XP)
	}
}

func TestMiniGameSystem_GenerateReward_GameTypeMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Card games should have higher rewards than lock-picking
	sys.SetSeed(12345)
	rewardCard := sys.generateReward(MiniGameCard, 0.5)

	sys.SetSeed(12345)
	rewardLock := sys.generateReward(MiniGameLockPicking, 0.5)

	if rewardCard.Gold <= rewardLock.Gold {
		t.Errorf("Card game gold (%d) should be > lock-picking (%d)", rewardCard.Gold, rewardLock.Gold)
	}
}

func TestMiniGameSystem_Update_InactiveGames(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Start and immediately end game
	sys.StartGame(entity.ID, MiniGameCard, 0.5)
	sys.EndGame(entity.ID, false)

	comp, _ := entity.GetComponent("minigame")
	gameComp := comp.(*MiniGameComponent)
	initialTime := gameComp.TimeElapsed

	// Update should not change inactive games
	sys.Update(nil, 1.0)

	if gameComp.TimeElapsed != initialTime {
		t.Error("Inactive game time should not update")
	}
}

func TestMiniGameSystem_Update_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Create multiple entities with mini-games
	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()
	entity3 := world.CreateEntity()
	world.Update(0) // Process entity additions

	sys.StartGame(entity1.ID, MiniGameCard, 0.5)
	sys.StartGame(entity2.ID, MiniGamePuzzle, 0.7)
	sys.StartGame(entity3.ID, MiniGameDice, 0.3)

	// Update all
	sys.Update(nil, 1.0)

	// All should have time updated
	comp1, ok := entity1.GetComponent("minigame")
	if !ok || comp1 == nil {
		t.Fatal("Entity 1 missing minigame component")
	}
	if comp1.(*MiniGameComponent).TimeElapsed != 1.0 {
		t.Error("Entity 1 time not updated")
	}

	comp2, ok := entity2.GetComponent("minigame")
	if !ok || comp2 == nil {
		t.Fatal("Entity 2 missing minigame component")
	}
	if comp2.(*MiniGameComponent).TimeElapsed != 1.0 {
		t.Error("Entity 2 time not updated")
	}

	comp3, ok := entity3.GetComponent("minigame")
	if !ok || comp3 == nil {
		t.Fatal("Entity 3 missing minigame component")
	}
	if comp3.(*MiniGameComponent).TimeElapsed != 1.0 {
		t.Error("Entity 3 time not updated")
	}
}

func TestMiniGameSystem_AwardReward_WithItems(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	playerEntity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Add inventory component to player with capacity for items
	playerEntity.AddComponent(NewInventoryComponent(10, 100.0))

	// Create item entities with ItemComponent
	testItem1 := &item.Item{
		ID:   "test-item-1",
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Damage: 10,
			Value:  100,
			Weight: 5.0,
		},
	}
	itemEntity1 := world.CreateEntity()
	itemEntity1.AddComponent(&ItemComponent{Item: testItem1})
	world.Update(0)

	testItem2 := &item.Item{
		ID:   "test-item-2",
		Name: "Test Potion",
		Type: item.TypeConsumable,
		Stats: item.Stats{
			Value:  50,
			Weight: 1.0,
		},
	}
	itemEntity2 := world.CreateEntity()
	itemEntity2.AddComponent(&ItemComponent{Item: testItem2})
	world.Update(0)

	// Create reward with item entity IDs
	reward := &Reward{
		Gold:  100,
		XP:    50.0,
		Items: []uint64{itemEntity1.ID, itemEntity2.ID},
	}

	// Award the reward
	sys.awardReward(playerEntity.ID, reward)

	// Verify gold and XP were awarded
	invComp, _ := playerEntity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 100 {
		t.Errorf("Gold = %d, want 100", inv.Gold)
	}

	// Verify items were added to inventory
	if len(inv.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(inv.Items))
	}

	// Check items are correct
	foundSword := false
	foundPotion := false
	for _, itm := range inv.Items {
		if itm.Name == "Test Sword" {
			foundSword = true
		}
		if itm.Name == "Test Potion" {
			foundPotion = true
		}
	}
	if !foundSword {
		t.Error("Test Sword not found in inventory")
	}
	if !foundPotion {
		t.Error("Test Potion not found in inventory")
	}
}

func TestMiniGameSystem_AwardReward_WithInvalidItemEntity(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	playerEntity := world.CreateEntity()
	world.Update(0)

	// Add inventory component
	playerEntity.AddComponent(NewInventoryComponent(10, 100.0))

	// Create reward with non-existent entity ID
	reward := &Reward{
		Gold:  100,
		XP:    50.0,
		Items: []uint64{99999}, // Non-existent entity
	}

	// Should not panic, just skip invalid items
	sys.awardReward(playerEntity.ID, reward)

	// Verify gold was still awarded
	invComp, _ := playerEntity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 100 {
		t.Errorf("Gold = %d, want 100", inv.Gold)
	}

	// No items should be added
	if len(inv.Items) != 0 {
		t.Errorf("Items count = %d, want 0", len(inv.Items))
	}
}

func TestMiniGameSystem_AwardReward_EntityWithoutItemComponent(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	playerEntity := world.CreateEntity()
	world.Update(0)

	// Add inventory component
	playerEntity.AddComponent(NewInventoryComponent(10, 100.0))

	// Create an entity without ItemComponent
	nonItemEntity := world.CreateEntity()
	world.Update(0)

	// Create reward with entity that has no item component
	reward := &Reward{
		Gold:  100,
		XP:    50.0,
		Items: []uint64{nonItemEntity.ID},
	}

	// Should not panic, just skip entities without item component
	sys.awardReward(playerEntity.ID, reward)

	// Verify gold was still awarded
	invComp, _ := playerEntity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 100 {
		t.Errorf("Gold = %d, want 100", inv.Gold)
	}

	// No items should be added
	if len(inv.Items) != 0 {
		t.Errorf("Items count = %d, want 0", len(inv.Items))
	}
}

func TestMiniGameSystem_AwardReward_InventoryFull(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	playerEntity := world.CreateEntity()
	world.Update(0)

	// Add inventory component with very limited capacity (1 item max, 10.0 weight max)
	playerEntity.AddComponent(NewInventoryComponent(1, 10.0))

	// Create two item entities
	testItem1 := &item.Item{
		ID:   "test-item-1",
		Name: "First Item",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Value:  100,
			Weight: 5.0,
		},
	}
	itemEntity1 := world.CreateEntity()
	itemEntity1.AddComponent(&ItemComponent{Item: testItem1})
	world.Update(0)

	testItem2 := &item.Item{
		ID:   "test-item-2",
		Name: "Second Item",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Value:  100,
			Weight: 5.0,
		},
	}
	itemEntity2 := world.CreateEntity()
	itemEntity2.AddComponent(&ItemComponent{Item: testItem2})
	world.Update(0)

	// Create reward with both items - but inventory can only hold 1
	reward := &Reward{
		Gold:  100,
		XP:    50.0,
		Items: []uint64{itemEntity1.ID, itemEntity2.ID},
	}

	// Award the reward
	sys.awardReward(playerEntity.ID, reward)

	// Verify gold was still awarded
	invComp, _ := playerEntity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 100 {
		t.Errorf("Gold = %d, want 100", inv.Gold)
	}

	// Only 1 item should be added (inventory full by count)
	if len(inv.Items) != 1 {
		t.Errorf("Items count = %d, want 1 (inventory should be full)", len(inv.Items))
	}

	// First item should be in inventory
	if inv.Items[0].Name != "First Item" {
		t.Errorf("Expected 'First Item' in inventory, got '%s'", inv.Items[0].Name)
	}
}

func TestMiniGameSystem_AwardReward_InventoryWeightExceeded(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	playerEntity := world.CreateEntity()
	world.Update(0)

	// Add inventory component with limited weight capacity (10 items max, but only 6.0 weight max)
	playerEntity.AddComponent(NewInventoryComponent(10, 6.0))

	// Create two items - first fits, second exceeds weight limit
	testItem1 := &item.Item{
		ID:   "test-item-1",
		Name: "Light Item",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Value:  100,
			Weight: 5.0,
		},
	}
	itemEntity1 := world.CreateEntity()
	itemEntity1.AddComponent(&ItemComponent{Item: testItem1})
	world.Update(0)

	testItem2 := &item.Item{
		ID:   "test-item-2",
		Name: "Heavy Item",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Value:  100,
			Weight: 5.0, // Would push total to 10.0, exceeding 6.0 limit
		},
	}
	itemEntity2 := world.CreateEntity()
	itemEntity2.AddComponent(&ItemComponent{Item: testItem2})
	world.Update(0)

	// Create reward with both items
	reward := &Reward{
		Gold:  100,
		XP:    50.0,
		Items: []uint64{itemEntity1.ID, itemEntity2.ID},
	}

	// Award the reward
	sys.awardReward(playerEntity.ID, reward)

	// Verify gold was still awarded
	invComp, _ := playerEntity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if inv.Gold != 100 {
		t.Errorf("Gold = %d, want 100", inv.Gold)
	}

	// Only 1 item should be added (second exceeds weight limit)
	if len(inv.Items) != 1 {
		t.Errorf("Items count = %d, want 1 (weight limit should prevent second item)", len(inv.Items))
	}

	// First item should be in inventory
	if inv.Items[0].Name != "Light Item" {
		t.Errorf("Expected 'Light Item' in inventory, got '%s'", inv.Items[0].Name)
	}
}

// Benchmark tests
func BenchmarkMiniGameSystem_StartGame(b *testing.B) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	entities := make([]*Entity, b.N)
	for i := 0; i < b.N; i++ {
		entities[i] = world.CreateEntity()
	}
	world.Update(0) // Process all entity additions

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.StartGame(entities[i].ID, MiniGameCard, 0.5)
	}
}

func BenchmarkMiniGameSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Create 100 active mini-games
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		world.Update(0) // Process entity addition
		sys.StartGame(entity.ID, MiniGameCard, 0.5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(nil, 0.016) // ~60 FPS
	}
}

func BenchmarkMiniGameSystem_GenerateReward(b *testing.B) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetSeed(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.generateReward(MiniGameCard, 0.5)
	}
}

// TestMiniGameSystem_SetItemGenerator verifies item generator setter.
func TestMiniGameSystem_SetItemGenerator(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Create item generator
	itemGen := item.NewItemGenerator()

	// Set the generator
	sys.SetItemGenerator(itemGen)

	// Verify it's set
	if sys.itemGenerator != itemGen {
		t.Error("SetItemGenerator did not set the generator correctly")
	}
}

// TestMiniGameSystem_SetItemGenerator_Nil verifies nil generator handling.
func TestMiniGameSystem_SetItemGenerator_Nil(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Set nil generator (should be allowed for graceful degradation)
	sys.SetItemGenerator(nil)

	// Verify it's nil
	if sys.itemGenerator != nil {
		t.Error("SetItemGenerator should allow nil")
	}
}

// TestMiniGameSystem_SetGenreID verifies genre ID setter.
func TestMiniGameSystem_SetGenreID(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)

	// Set genre ID
	sys.SetGenreID("fantasy")

	// Verify it's set
	if sys.genreID != "fantasy" {
		t.Errorf("SetGenreID = %q, want %q", sys.genreID, "fantasy")
	}

	// Test different genre
	sys.SetGenreID("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("SetGenreID = %q, want %q", sys.genreID, "scifi")
	}
}

// TestMiniGameSystem_GenerateReward_WithItemGenerator verifies item generation.
func TestMiniGameSystem_GenerateReward_WithItemGenerator(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetSeed(12345)
	sys.SetItemGenerator(item.NewItemGenerator())
	sys.SetGenreID("fantasy")

	tests := []struct {
		name          string
		difficulty    float64
		expectedItems int
		minItems      int
		maxItems      int
	}{
		{
			name:          "low difficulty",
			difficulty:    0.3,
			expectedItems: 1,
			minItems:      1,
			maxItems:      1,
		},
		{
			name:          "medium difficulty",
			difficulty:    0.8,
			expectedItems: 2,
			minItems:      2,
			maxItems:      2,
		},
		{
			name:          "high difficulty",
			difficulty:    0.95,
			expectedItems: 3,
			minItems:      3,
			maxItems:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := sys.generateReward(MiniGameCard, tt.difficulty)

			if reward == nil {
				t.Fatal("generateReward returned nil")
			}

			// Process entity additions
			world.Update(0)

			itemCount := len(reward.Items)
			if itemCount < tt.minItems || itemCount > tt.maxItems {
				t.Errorf("Item count = %d, want between %d and %d",
					itemCount, tt.minItems, tt.maxItems)
			}

			// Verify each item entity exists and has ItemComponent
			for _, itemID := range reward.Items {
				entity, exists := world.GetEntity(itemID)
				if !exists {
					t.Errorf("Item entity %d does not exist", itemID)
					continue
				}

				itemCompRaw, hasItem := entity.GetComponent("item")
				if !hasItem {
					t.Errorf("Item entity %d missing ItemComponent", itemID)
					continue
				}

				itemComp, ok := itemCompRaw.(*ItemComponent)
				if !ok || itemComp.Item == nil {
					t.Errorf("Item entity %d has invalid ItemComponent", itemID)
				}
			}
		})
	}
}

// TestMiniGameSystem_GenerateReward_WithoutItemGenerator verifies graceful degradation.
func TestMiniGameSystem_GenerateReward_WithoutItemGenerator(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetSeed(12345)
	// Don't set item generator

	reward := sys.generateReward(MiniGameCard, 0.8)

	if reward == nil {
		t.Fatal("generateReward returned nil")
	}

	// Should still get gold and XP
	if reward.Gold == 0 {
		t.Error("Gold should be non-zero even without item generator")
	}
	if reward.XP == 0 {
		t.Error("XP should be non-zero even without item generator")
	}

	// Should have no items
	if len(reward.Items) != 0 {
		t.Errorf("Items = %d, want 0 when no item generator set", len(reward.Items))
	}
}

// TestMiniGameSystem_GenerateReward_ItemsDeterministic verifies items are deterministic.
func TestMiniGameSystem_GenerateReward_ItemsDeterministic(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetItemGenerator(item.NewItemGenerator())
	sys.SetGenreID("fantasy")

	// Generate with seed 1
	sys.SetSeed(12345)
	reward1 := sys.generateReward(MiniGameCard, 0.8)
	world.Update(0) // Process entity additions

	// Generate with same seed again
	sys.SetSeed(12345)
	reward2 := sys.generateReward(MiniGameCard, 0.8)
	world.Update(0) // Process entity additions

	// Should have same number of items
	if len(reward1.Items) != len(reward2.Items) {
		t.Errorf("Item counts differ: %d != %d", len(reward1.Items), len(reward2.Items))
	}

	// Items should have same properties (though different entity IDs)
	for i := 0; i < len(reward1.Items) && i < len(reward2.Items); i++ {
		entity1, _ := world.GetEntity(reward1.Items[i])
		entity2, _ := world.GetEntity(reward2.Items[i])

		item1Comp, _ := entity1.GetComponent("item")
		item2Comp, _ := entity2.GetComponent("item")

		item1 := item1Comp.(*ItemComponent).Item
		item2 := item2Comp.(*ItemComponent).Item

		// Check items have same type and name
		if item1.Type != item2.Type {
			t.Errorf("Item %d type differs: %v != %v", i, item1.Type, item2.Type)
		}
		if item1.Name != item2.Name {
			t.Errorf("Item %d name differs: %s != %s", i, item1.Name, item2.Name)
		}
	}
}

// TestMiniGameSystem_GenerateReward_ItemsCreatedAsEntities verifies items are proper entities.
func TestMiniGameSystem_GenerateReward_ItemsCreatedAsEntities(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetSeed(12345)
	sys.SetItemGenerator(item.NewItemGenerator())
	sys.SetGenreID("fantasy")

	reward := sys.generateReward(MiniGameCard, 0.9) // High difficulty = 3 items
	world.Update(0)                                 // Process entity additions

	if len(reward.Items) == 0 {
		t.Fatal("Expected items to be generated")
	}

	// Verify each item is a proper entity with ItemComponent
	for i, itemID := range reward.Items {
		entity, exists := world.GetEntity(itemID)
		if !exists {
			t.Errorf("Item %d: Entity %d not found in world", i, itemID)
			continue
		}

		itemCompRaw, hasItem := entity.GetComponent("item")
		if !hasItem {
			t.Errorf("Item %d: Entity %d missing ItemComponent", i, itemID)
			continue
		}

		itemComp, ok := itemCompRaw.(*ItemComponent)
		if !ok {
			t.Errorf("Item %d: Entity %d has invalid ItemComponent type", i, itemID)
			continue
		}

		if itemComp.Item == nil {
			t.Errorf("Item %d: Entity %d has nil Item", i, itemID)
			continue
		}

		// Verify item has basic properties
		if itemComp.Item.Name == "" {
			t.Errorf("Item %d: Item has empty name", i)
		}
	}
}

// TestMiniGameSystem_EndGame_WithGeneratedItems verifies full integration.
func TestMiniGameSystem_EndGame_WithGeneratedItems(t *testing.T) {
	world := NewWorld()
	sys := NewMiniGameSystem(world)
	sys.SetSeed(12345)
	sys.SetItemGenerator(item.NewItemGenerator())
	sys.SetGenreID("fantasy")

	// Create player entity with inventory
	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(NewInventoryComponent(20, 200.0))
	playerEntity.AddComponent(&ExperienceComponent{Level: 1, CurrentXP: 0})
	world.Update(0)

	// Start a game
	err := sys.StartGame(playerEntity.ID, MiniGameCard, 0.9)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Process entity additions for reward items
	world.Update(0)

	// Get the game component to verify reward was generated
	gameComp := sys.GetGameComponent(playerEntity.ID)
	if gameComp == nil {
		t.Fatal("GetGameComponent returned nil")
	}

	initialItemCount := len(gameComp.Reward.Items)
	if initialItemCount == 0 {
		t.Error("Expected reward to have items")
	}

	// End the game successfully
	err = sys.EndGame(playerEntity.ID, true)
	if err != nil {
		t.Fatalf("EndGame failed: %v", err)
	}

	// Verify items were added to inventory
	invCompRaw, _ := playerEntity.GetComponent("inventory")
	inv := invCompRaw.(*InventoryComponent)

	if len(inv.Items) == 0 {
		t.Error("Expected items to be added to inventory")
	}

	// Verify items in inventory match generated items count
	if len(inv.Items) != initialItemCount {
		t.Errorf("Inventory has %d items, expected %d", len(inv.Items), initialItemCount)
	}

	// Verify XP and gold were also awarded
	expCompRaw, _ := playerEntity.GetComponent("experience")
	exp := expCompRaw.(*ExperienceComponent)
	if exp.CurrentXP == 0 {
		t.Error("Expected XP to be awarded")
	}

	if inv.Gold == 0 {
		t.Error("Expected gold to be awarded")
	}
}
