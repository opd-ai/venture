package qol

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	config := Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: true,
	}

	manager := NewManager(config)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.autoLoot == nil {
		t.Error("autoLoot manager not initialized")
	}

	if manager.craftQueue == nil {
		t.Error("craftQueue manager not initialized")
	}

	if manager.guildInvites == nil {
		t.Error("guildInvites manager not initialized")
	}

	if manager.mountWhistle == nil {
		t.Error("mountWhistle manager not initialized")
	}

	if manager.storageSorter == nil {
		t.Error("storageSorter not initialized")
	}

	if manager.recipeTracker == nil {
		t.Error("recipeTracker not initialized")
	}

	gotConfig := manager.GetConfig()
	if gotConfig.AutoLoot != config.AutoLoot {
		t.Errorf("AutoLoot = %v, want %v", gotConfig.AutoLoot, config.AutoLoot)
	}
	if gotConfig.AutoSort != config.AutoSort {
		t.Errorf("AutoSort = %v, want %v", gotConfig.AutoSort, config.AutoSort)
	}
	if gotConfig.QuickDeposit != config.QuickDeposit {
		t.Errorf("QuickDeposit = %v, want %v", gotConfig.QuickDeposit, config.QuickDeposit)
	}
}

func TestManagerAccessors(t *testing.T) {
	manager := NewManager(Config{})

	tests := []struct {
		name     string
		accessor func() interface{}
		wantNil  bool
	}{
		{"AutoLoot", func() interface{} { return manager.AutoLoot() }, false},
		{"CraftQueue", func() interface{} { return manager.CraftQueue() }, false},
		{"GuildInvites", func() interface{} { return manager.GuildInvites() }, false},
		{"MountWhistle", func() interface{} { return manager.MountWhistle() }, false},
		{"StorageSorter", func() interface{} { return manager.StorageSorter() }, false},
		{"RecipeTracker", func() interface{} { return manager.RecipeTracker() }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.accessor()
			if (result == nil) != tt.wantNil {
				t.Errorf("%s() = nil, want non-nil", tt.name)
			}
		})
	}
}

func TestManagerSetConfig(t *testing.T) {
	manager := NewManager(Config{
		AutoLoot:     false,
		AutoSort:     false,
		QuickDeposit: false,
	})

	newConfig := Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: true,
	}

	manager.SetConfig(newConfig)

	gotConfig := manager.GetConfig()
	if gotConfig.AutoLoot != newConfig.AutoLoot {
		t.Errorf("AutoLoot = %v, want %v", gotConfig.AutoLoot, newConfig.AutoLoot)
	}
	if gotConfig.AutoSort != newConfig.AutoSort {
		t.Errorf("AutoSort = %v, want %v", gotConfig.AutoSort, newConfig.AutoSort)
	}
	if gotConfig.QuickDeposit != newConfig.QuickDeposit {
		t.Errorf("QuickDeposit = %v, want %v", gotConfig.QuickDeposit, newConfig.QuickDeposit)
	}
}

func TestManagerConcurrentAccess(t *testing.T) {
	manager := NewManager(Config{AutoLoot: true})

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.AutoLoot()
			_ = manager.CraftQueue()
			_ = manager.GetConfig()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent writes
	done2 := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(val bool) {
			manager.SetConfig(Config{AutoLoot: val})
			done2 <- true
		}(i%2 == 0)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done2
	}

	// Verify manager is still functional
	config := manager.GetConfig()
	if manager.AutoLoot() == nil {
		t.Error("AutoLoot() returned nil after concurrent access")
	}
	// Config will be either true or false depending on last write
	t.Logf("Final config after concurrent writes: %+v", config)
}

func TestManagerIntegration(t *testing.T) {
	// Create manager with all features enabled
	manager := NewManager(Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: true,
	})

	// Test auto-loot functionality
	autoLoot := manager.AutoLoot()
	autoLoot.SetRadius(1, 8.0)
	config := autoLoot.GetConfig(1)
	if config.Radius != 8.0 {
		t.Errorf("AutoLoot radius = %v, want 8.0", config.Radius)
	}

	// Test craft queue functionality
	craftQueue := manager.CraftQueue()
	err := craftQueue.AddRecipe(1, "iron_sword", 5)
	if err != nil {
		t.Errorf("AddRecipe failed: %v", err)
	}
	queue := craftQueue.GetQueue(1)
	if len(queue) != 1 {
		t.Errorf("queue length = %d, want 1", len(queue))
	}

	// Test guild invitation functionality
	guildInvites := manager.GuildInvites()
	inv := &GuildInvitation{
		InvitationID: "test-inv-1",
		GuildID:      "guild-1",
		InviteeID:    "player-1",
	}
	guildInvites.SendInvitation(inv)
	pending := guildInvites.GetPendingInvitations("player-1")
	if len(pending) != 1 {
		t.Errorf("pending invitations = %d, want 1", len(pending))
	}

	// Test mount whistle functionality
	mountWhistle := manager.MountWhistle()
	summon := &MountSummon{
		PlayerID:    1,
		VehicleID:   100,
		VehicleType: "horse",
		CurrentPos:  [2]float64{0, 0},
		TargetPos:   [2]float64{10, 10},
	}
	mountWhistle.SummonMount(summon)
	active := mountWhistle.GetActiveSummon(1)
	if active == nil {
		t.Error("active summon is nil")
	}

	// Test storage sorter functionality
	sorter := manager.StorageSorter()
	preset := sorter.GetPreset("rarity")
	if preset == nil {
		t.Error("rarity preset is nil")
	}
	if preset.PrimaryCriteria != SortByRarity {
		t.Errorf("preset criteria = %v, want SortByRarity", preset.PrimaryCriteria)
	}

	// Test recipe tracker functionality
	tracker := manager.RecipeTracker()
	recipeInfo := &RecipeTrackingInfo{
		RecipeID:      "iron_sword",
		RecipeName:    "Iron Sword",
		RequiredMats:  map[string]int{"iron": 10, "wood": 2},
		AvailableMats: map[string]int{"iron": 5, "wood": 3},
	}
	tracker.TrackRecipe(1, recipeInfo)
	tracked := tracker.GetTrackedRecipes(1)
	if len(tracked) != 1 {
		t.Errorf("tracked recipes = %d, want 1", len(tracked))
	}
	if tracked[0].MissingMats["iron"] != 5 {
		t.Errorf("missing iron = %d, want 5", tracked[0].MissingMats["iron"])
	}
}
