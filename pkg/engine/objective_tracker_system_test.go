package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// TestNewObjectiveTrackerSystem tests system creation
func TestNewObjectiveTrackerSystem(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	if sys == nil {
		t.Fatal("NewObjectiveTrackerSystem returned nil")
	}

	if sys.exploredTiles == nil {
		t.Error("exploredTiles map not initialized")
	}

	if sys.onQuestComplete != nil {
		t.Error("onQuestComplete callback should be nil initially")
	}
}

// TestSetQuestCompleteCallback tests callback setting
func TestSetQuestCompleteCallback(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	called := false
	callback := func(entity *Entity, qst *quest.Quest) {
		called = true
	}

	sys.SetQuestCompleteCallback(callback)

	if sys.onQuestComplete == nil {
		t.Fatal("Callback not set")
	}

	// Test callback works
	sys.onQuestComplete(nil, nil)
	if !called {
		t.Error("Callback was not called")
	}
}

// TestOnEnemyKilled tests kill objective tracking
func TestOnEnemyKilled(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	// Create player with quest tracker
	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Create a kill quest
	q := &quest.Quest{
		ID:   "kill-quest-1",
		Type: quest.TypeKill,
		Objectives: []quest.Objective{
			{Target: "enemy", Required: 5, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Create enemy
	enemy := NewEntity(2)

	// Kill an enemy
	sys.OnEnemyKilled(player, enemy)

	// Check progress
	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 1 {
		t.Errorf("Kill progress = %d, want 1", tracked.Quest.Objectives[0].Current)
	}

	// Kill multiple enemies
	for i := 0; i < 4; i++ {
		sys.OnEnemyKilled(player, enemy)
	}

	if tracked.Quest.Objectives[0].Current != 5 {
		t.Errorf("Kill progress = %d, want 5", tracked.Quest.Objectives[0].Current)
	}
}

// TestOnEnemyKilled_NoQuestTracker tests kill with non-quest entity
func TestOnEnemyKilled_NoQuestTracker(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	// Entity without quest tracker
	entity := NewEntity(1)
	enemy := NewEntity(2)

	// Should not panic
	sys.OnEnemyKilled(entity, enemy)
}

// TestOnItemCollected tests collect objective tracking
func TestOnItemCollected(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	// Create player with quest tracker
	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Create a collect quest
	q := &quest.Quest{
		ID:   "collect-quest-1",
		Type: quest.TypeCollect,
		Objectives: []quest.Objective{
			{Target: "potion", Required: 3, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Collect items
	sys.OnItemCollected(player, "potion")
	sys.OnItemCollected(player, "Healing Potion")
	sys.OnItemCollected(player, "POTION")

	// Check progress
	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 3 {
		t.Errorf("Collect progress = %d, want 3", tracked.Quest.Objectives[0].Current)
	}
}

// TestOnItemCollected_WrongItem tests collecting wrong items
func TestOnItemCollected_WrongItem(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	q := &quest.Quest{
		ID:   "collect-quest-1",
		Type: quest.TypeCollect,
		Objectives: []quest.Objective{
			{Target: "sword", Required: 1, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Collect wrong item
	sys.OnItemCollected(player, "potion")

	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 0 {
		t.Errorf("Progress should remain 0, got %d", tracked.Quest.Objectives[0].Current)
	}
}

// TestOnUIOpened tests UI interaction tracking
func TestOnUIOpened(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Create a tutorial quest with UI objectives
	q := &quest.Quest{
		ID:   "tutorial-ui",
		Type: quest.TypeKill, // Any type works
		Objectives: []quest.Objective{
			{Target: "inventory", Required: 1, Current: 0},
			{Target: "quest_log", Required: 1, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Open inventory
	sys.OnUIOpened(player, "inventory")

	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 1 {
		t.Errorf("Inventory objective progress = %d, want 1", tracked.Quest.Objectives[0].Current)
	}

	// Open quest log
	sys.OnUIOpened(player, "quest_log")

	if tracked.Quest.Objectives[1].Current != 1 {
		t.Errorf("Quest log objective progress = %d, want 1", tracked.Quest.Objectives[1].Current)
	}
}

// TestOnTileExplored tests exploration tracking
func TestOnTileExplored(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Create explore quest
	q := &quest.Quest{
		ID:   "explore-quest-1",
		Type: quest.TypeExplore,
		Objectives: []quest.Objective{
			{Target: "dungeon tiles", Required: 10, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Explore some tiles
	sys.OnTileExplored(player, 0, 0)
	sys.OnTileExplored(player, 1, 0)
	sys.OnTileExplored(player, 0, 1)
	sys.OnTileExplored(player, 1, 1)
	sys.OnTileExplored(player, 2, 2)

	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 5 {
		t.Errorf("Explore progress = %d, want 5", tracked.Quest.Objectives[0].Current)
	}

	// Re-explore same tile (should not increment)
	sys.OnTileExplored(player, 0, 0)

	if tracked.Quest.Objectives[0].Current != 5 {
		t.Errorf("Progress should remain 5 after re-exploring, got %d", tracked.Quest.Objectives[0].Current)
	}
}

// TestUpdateExplorationObjectives tests position-based exploration
func TestUpdateExplorationObjectives(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Add position component (32 pixels per tile)
	pos := &PositionComponent{X: 64, Y: 96} // Tile (2, 3)
	player.AddComponent(pos)

	// Create explore quest
	q := &quest.Quest{
		ID:   "explore-quest-1",
		Type: quest.TypeExplore,
		Objectives: []quest.Objective{
			{Target: "explore tiles", Required: 10, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Update should track current tile
	sys.updateExplorationObjectives(player)

	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 1 {
		t.Errorf("Explore progress = %d, want 1", tracked.Quest.Objectives[0].Current)
	}

	// Move to new position
	pos.X = 128
	pos.Y = 32
	sys.updateExplorationObjectives(player)

	if tracked.Quest.Objectives[0].Current != 2 {
		t.Errorf("Explore progress = %d, want 2 after moving", tracked.Quest.Objectives[0].Current)
	}
}

// TestCheckQuestCompletion tests quest completion detection
func TestCheckQuestCompletion(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Create quest with completion callback
	callbackCalled := false
	var completedQuest *quest.Quest
	sys.SetQuestCompleteCallback(func(entity *Entity, qst *quest.Quest) {
		callbackCalled = true
		completedQuest = qst
	})

	// Add quest with one objective
	q := &quest.Quest{
		ID:   "test-quest",
		Type: quest.TypeKill,
		Objectives: []quest.Objective{
			{Target: "enemy", Required: 1, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Complete objective
	tracked := tracker.ActiveQuests[0]
	tracked.Quest.Objectives[0].Current = 1

	// Check completion
	sys.checkQuestCompletion(player)

	if !callbackCalled {
		t.Error("Quest completion callback not called")
	}

	if completedQuest == nil {
		t.Fatal("Completed quest is nil")
	}

	if completedQuest.ID != "test-quest" {
		t.Errorf("Completed quest ID = %s, want test-quest", completedQuest.ID)
	}

	if tracked.Status != QuestStatusCompleted {
		t.Errorf("Quest status = %v, want %v", tracked.Status, QuestStatusCompleted)
	}
}

// TestCheckQuestCompletion_MultipleObjectives tests multi-objective completion
func TestCheckQuestCompletion_MultipleObjectives(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	player := NewEntity(1)
	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	callbackCalled := false
	sys.SetQuestCompleteCallback(func(entity *Entity, qst *quest.Quest) {
		callbackCalled = true
	})

	// Add quest with multiple objectives
	q := &quest.Quest{
		ID:   "multi-quest",
		Type: quest.TypeKill,
		Objectives: []quest.Objective{
			{Target: "enemy", Required: 3, Current: 0},
			{Target: "boss", Required: 1, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	tracked := tracker.ActiveQuests[0]

	// Complete first objective only
	tracked.Quest.Objectives[0].Current = 3
	sys.checkQuestCompletion(player)

	if callbackCalled {
		t.Error("Callback should not be called with incomplete objectives")
	}

	// Complete second objective
	tracked.Quest.Objectives[1].Current = 1
	sys.checkQuestCompletion(player)

	if !callbackCalled {
		t.Error("Callback should be called when all objectives complete")
	}
}

// TestMatchesTarget tests target matching logic
func TestMatchesTarget(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	tests := []struct {
		name    string
		target  string
		actual  string
		context string
		want    bool
	}{
		// Exact matches
		{"Exact match", "goblin", "goblin", "kill", true},
		{"Case insensitive", "Goblin", "GOBLIN", "kill", true},

		// Partial matches
		{"Target contains actual", "fire goblin", "goblin", "kill", true},
		{"Actual contains target", "goblin", "fire goblin", "kill", true},

		// Generic enemy matches
		{"Generic enemy", "enemy", "orc", "kill", true},
		{"Generic enemies", "enemies", "dragon", "kill", true},
		{"Generic monster", "monster", "zombie", "kill", true},

		// Generic item matches
		{"Generic item", "item", "sword", "collect", true},
		{"Generic items", "items", "potion", "collect", true},

		// UI matches
		{"Inventory UI", "inventory", "inventory", "ui", true},
		{"Quest log UI", "quest_log", "quest_log", "ui", true},
		{"Character UI", "character", "character", "ui", true},
		{"Skills UI", "skills", "skills", "ui", true},
		{"Map UI", "map", "map", "ui", true},

		// Non-matches
		{"Different items", "sword", "potion", "collect", false},
		{"Different enemies", "goblin", "orc", "kill", false},
		{"Wrong UI", "inventory", "quest_log", "ui", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.matchesTarget(tt.target, tt.actual, tt.context)
			if got != tt.want {
				t.Errorf("matchesTarget(%q, %q, %q) = %v, want %v",
					tt.target, tt.actual, tt.context, got, tt.want)
			}
		})
	}
}

// TestTileKeyFromCoords tests tile key generation
func TestTileKeyFromCoords(t *testing.T) {
	tests := []struct {
		x, y int
		want string
	}{
		{0, 0, "0,0"},
		{1, 2, "1,2"},
		{-5, 10, "-5,10"},
		{100, -50, "100,-50"},
	}

	for _, tt := range tests {
		got := tileKeyFromCoords(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("tileKeyFromCoords(%d, %d) = %q, want %q", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestAwardQuestRewards tests reward distribution
func TestAwardQuestRewards(t *testing.T) {
	player := NewEntity(1)

	// Add experience component
	exp := NewExperienceComponent()
	player.AddComponent(exp)

	// Add inventory component
	inv := NewInventoryComponent(10, 100.0)
	player.AddComponent(inv)

	// Create quest with rewards
	q := &quest.Quest{
		ID: "reward-quest",
		Reward: quest.Reward{
			XP:          100,
			Gold:        50,
			SkillPoints: 2,
		},
	}

	// Award rewards
	AwardQuestRewards(player, q)

	// Check XP
	if exp.CurrentXP != 100 {
		t.Errorf("XP = %d, want 100", exp.CurrentXP)
	}

	// Check gold
	if inv.Gold != 50 {
		t.Errorf("Gold = %d, want 50", inv.Gold)
	}

	// Check skill points
	if exp.SkillPoints != 2 {
		t.Errorf("SkillPoints = %d, want 2", exp.SkillPoints)
	}
}

// TestAwardQuestRewards_NoComponents tests rewards with missing components
func TestAwardQuestRewards_NoComponents(t *testing.T) {
	player := NewEntity(1)

	q := &quest.Quest{
		ID: "reward-quest",
		Reward: quest.Reward{
			XP:          100,
			Gold:        50,
			SkillPoints: 2,
		},
	}

	// Should not panic
	AwardQuestRewards(player, q)
}

// TestUpdate tests the main update loop
func TestUpdate(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	// Create player with position and quest tracker
	player := NewEntity(1)
	pos := &PositionComponent{X: 64, Y: 64} // Tile (2, 2)
	player.AddComponent(pos)

	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Add explore quest
	q := &quest.Quest{
		ID:   "explore-quest",
		Type: quest.TypeExplore,
		Objectives: []quest.Objective{
			{Target: "tiles", Required: 5, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Update system
	entities := []*Entity{player}
	sys.Update(entities, 0.016)

	// Should have explored initial tile
	tracked := tracker.ActiveQuests[0]
	if tracked.Quest.Objectives[0].Current != 1 {
		t.Errorf("Explore progress = %d, want 1", tracked.Quest.Objectives[0].Current)
	}

	// Move and update
	pos.X = 128
	pos.Y = 128
	sys.Update(entities, 0.016)

	if tracked.Quest.Objectives[0].Current != 2 {
		t.Errorf("Explore progress = %d, want 2 after move", tracked.Quest.Objectives[0].Current)
	}
}

// TestUpdate_QuestCompletion tests automatic completion in update loop
func TestUpdate_QuestCompletion(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	callbackCalled := false
	sys.SetQuestCompleteCallback(func(entity *Entity, qst *quest.Quest) {
		callbackCalled = true
	})

	player := NewEntity(1)
	pos := &PositionComponent{X: 0, Y: 0}
	player.AddComponent(pos)

	tracker := NewQuestTrackerComponent(10)
	player.AddComponent(tracker)

	// Add quest that's almost complete
	q := &quest.Quest{
		ID:   "almost-done",
		Type: quest.TypeExplore,
		Objectives: []quest.Objective{
			{Target: "tiles", Required: 1, Current: 0},
		},
	}
	tracker.AcceptQuest(q, 0)

	// Update should complete the quest
	entities := []*Entity{player}
	sys.Update(entities, 0.016)

	if !callbackCalled {
		t.Error("Quest completion callback should have been called during update")
	}
}

// TestObjectiveTrackerSystem_Integration tests full workflow
func TestObjectiveTrackerSystem_Integration(t *testing.T) {
	sys := NewObjectiveTrackerSystem()

	completedQuests := []string{}
	sys.SetQuestCompleteCallback(func(entity *Entity, qst *quest.Quest) {
		completedQuests = append(completedQuests, qst.ID)
		AwardQuestRewards(entity, qst)
	})

	// Create player with all components
	player := NewEntity(1)
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(NewQuestTrackerComponent(10))
	player.AddComponent(NewExperienceComponent())
	player.AddComponent(NewInventoryComponent(10, 100.0))

	trackerComp, ok := player.GetComponent("questtracker")
	if !ok {
		t.Fatal("Failed to get questtracker component")
	}
	tracker := trackerComp.(*QuestTrackerComponent)

	// Add multiple quests
	killQuest := &quest.Quest{
		ID:   "kill-5-enemies",
		Type: quest.TypeKill,
		Objectives: []quest.Objective{
			{Target: "enemy", Required: 5, Current: 0},
		},
		Reward: quest.Reward{XP: 100},
	}
	tracker.AcceptQuest(killQuest, 0)

	collectQuest := &quest.Quest{
		ID:   "collect-3-potions",
		Type: quest.TypeCollect,
		Objectives: []quest.Objective{
			{Target: "potion", Required: 3, Current: 0},
		},
		Reward: quest.Reward{Gold: 50},
	}
	tracker.AcceptQuest(collectQuest, 0)

	// Simulate gameplay
	enemy := NewEntity(2)

	// Kill some enemies
	sys.OnEnemyKilled(player, enemy)
	sys.OnEnemyKilled(player, enemy)
	sys.OnEnemyKilled(player, enemy)

	// Collect some items
	sys.OnItemCollected(player, "health potion")
	sys.OnItemCollected(player, "mana potion")

	// Update system
	entities := []*Entity{player}
	sys.Update(entities, 0.016)

	// Check progress
	if tracker.ActiveQuests[0].Quest.Objectives[0].Current != 3 {
		t.Error("Kill quest should have 3/5 progress")
	}
	if tracker.ActiveQuests[1].Quest.Objectives[0].Current != 2 {
		t.Error("Collect quest should have 2/3 progress")
	}

	// Complete collect quest
	sys.OnItemCollected(player, "potion")
	sys.Update(entities, 0.016)

	if len(completedQuests) != 1 {
		t.Errorf("Should have completed 1 quest, got %d", len(completedQuests))
	}
	if completedQuests[0] != "collect-3-potions" {
		t.Errorf("Completed quest = %s, want collect-3-potions", completedQuests[0])
	}

	// Complete kill quest
	sys.OnEnemyKilled(player, enemy)
	sys.OnEnemyKilled(player, enemy)
	sys.Update(entities, 0.016)

	if len(completedQuests) != 2 {
		t.Errorf("Should have completed 2 quests, got %d", len(completedQuests))
	}

	// Check rewards
	expComp, ok := player.GetComponent("experience")
	if !ok {
		t.Fatal("Failed to get experience component")
	}
	exp := expComp.(*ExperienceComponent)

	invComp, ok := player.GetComponent("inventory")
	if !ok {
		t.Fatal("Failed to get inventory component")
	}
	inv := invComp.(*InventoryComponent)

	if exp.CurrentXP != 100 {
		t.Errorf("XP = %d, want 100", exp.CurrentXP)
	}
	if inv.Gold != 50 {
		t.Errorf("Gold = %d, want 50", inv.Gold)
	}
}
