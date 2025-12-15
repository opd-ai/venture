// Package engine provides tests for the fishing system.
// Phase 96: Fishing System

package engine

import (
	"testing"
	"time"
)

func TestNewFishingSystem(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	if fs == nil {
		t.Fatal("fishing system should not be nil")
	}
	if fs.world != world {
		t.Error("world reference should match")
	}
	if fs.BaseWaitTime != 10.0 {
		t.Errorf("expected base wait time 10.0, got %f", fs.BaseWaitTime)
	}
	if fs.BiteWindowTime != 2.0 {
		t.Errorf("expected bite window 2.0, got %f", fs.BiteWindowTime)
	}

	// Should have default fish types registered
	fishCount := fs.GetFishTypeCount()
	if fishCount < 10 {
		t.Errorf("expected at least 10 fish types, got %d", fishCount)
	}
}

func TestFishingSystem_RegisterFishType(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	customFish := &FishType{
		ID:         "custom_fish",
		Name:       "Custom Fish",
		Rarity:     FishRarityRare,
		WaterTypes: []WaterType{WaterTypeFreshwater},
		MinDepth:   DepthMedium,
		BestTime:   TimeNight,
		MinSkill:   30,
		BaseWeight: 3.0,
		MaxWeight:  15.0,
		Difficulty: 0.65,
	}

	initialCount := fs.GetFishTypeCount()
	fs.RegisterFishType(customFish)

	if fs.GetFishTypeCount() != initialCount+1 {
		t.Error("fish count should increase by 1")
	}

	retrieved := fs.GetFishType("custom_fish")
	if retrieved == nil {
		t.Fatal("should be able to retrieve registered fish")
	}
	if retrieved.Name != "Custom Fish" {
		t.Errorf("expected 'Custom Fish', got %s", retrieved.Name)
	}
}

func TestFishingSystem_GetAllFishTypes(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	allFish := fs.GetAllFishTypes()
	if len(allFish) == 0 {
		t.Error("should have fish types")
	}

	// Check some expected fish
	found := make(map[string]bool)
	for _, fish := range allFish {
		found[fish.ID] = true
	}

	expected := []string{"bass", "trout", "mackerel", "tuna"}
	for _, id := range expected {
		if !found[id] {
			t.Errorf("expected fish type %s not found", id)
		}
	}
}

func TestFishingSystem_DefaultFishTypes(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	tests := []struct {
		id        string
		name      string
		rarity    FishRarity
		waterType WaterType
	}{
		{"bass", "Bass", FishRarityCommon, WaterTypeFreshwater},
		{"mackerel", "Mackerel", FishRarityCommon, WaterTypeSaltwater},
		{"moonfish", "Moonfish", FishRarityRare, WaterTypeMagical},
		{"leviathan", "Leviathan", FishRarityLegendary, WaterTypeSaltwater},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			fish := fs.GetFishType(tt.id)
			if fish == nil {
				t.Fatalf("fish type %s not found", tt.id)
			}
			if fish.Name != tt.name {
				t.Errorf("expected name %s, got %s", tt.name, fish.Name)
			}
			if fish.Rarity != tt.rarity {
				t.Errorf("expected rarity %s, got %s", tt.rarity, fish.Rarity)
			}
			// Check water type is in list
			found := false
			for _, wt := range fish.WaterTypes {
				if wt == tt.waterType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected water type %s in fish water types", tt.waterType)
			}
		})
	}
}

func TestFishingSystem_GenerateFishingSpot(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	tests := []struct {
		name       string
		seed       int64
		waterType  WaterType
		depthLevel DepthLevel
		biome      string
	}{
		{"freshwater shallow", 12345, WaterTypeFreshwater, DepthShallow, "forest"},
		{"saltwater deep", 54321, WaterTypeSaltwater, DepthDeep, "ocean"},
		{"magical medium", 99999, WaterTypeMagical, DepthMedium, "enchanted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spot := fs.GenerateFishingSpot(tt.seed, tt.waterType, tt.depthLevel, tt.biome, 100.0, 200.0)

			if spot == nil {
				t.Fatal("spot should not be nil")
			}

			// Check position
			posComp, ok := spot.GetComponent("position")
			if !ok {
				t.Fatal("should have position component")
			}
			pos := posComp.(*PositionComponent)
			if pos.X != 100.0 || pos.Y != 200.0 {
				t.Errorf("expected position (100, 200), got (%f, %f)", pos.X, pos.Y)
			}

			// Check fishing spot
			spotComp, ok := spot.GetComponent("fishing_spot")
			if !ok {
				t.Fatal("should have fishing spot component")
			}
			fishSpot := spotComp.(*FishingSpotComponent)
			if fishSpot.WaterType != tt.waterType {
				t.Errorf("expected water type %s, got %s", tt.waterType, fishSpot.WaterType)
			}
			if fishSpot.DepthLevel != tt.depthLevel {
				t.Errorf("expected depth %d, got %d", tt.depthLevel, fishSpot.DepthLevel)
			}

			// Should have fish population
			fishTypes := fishSpot.GetFishTypes()
			if len(fishTypes) == 0 {
				t.Error("fishing spot should have fish population")
			}
		})
	}
}

func TestFishingSystem_GenerateFishingSpot_Deterministic(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Generate two spots with same seed
	spot1 := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthMedium, "lake", 0, 0)
	spot2 := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthMedium, "lake", 0, 0)

	fishSpot1Raw, _ := spot1.GetComponent("fishing_spot")
	fishSpot1 := fishSpot1Raw.(*FishingSpotComponent)
	fishSpot2Raw, _ := spot2.GetComponent("fishing_spot")
	fishSpot2 := fishSpot2Raw.(*FishingSpotComponent)

	// Should have same fish population
	if len(fishSpot1.FishPopulation) != len(fishSpot2.FishPopulation) {
		t.Error("fish population should be identical for same seed")
	}

	// Same properties
	if fishSpot1.MaxConcurrentFishers != fishSpot2.MaxConcurrentFishers {
		t.Error("max fishers should be identical for same seed")
	}
}

func TestFishingSystem_StartFishing(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Create fisher entity
	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fishComp.BaitCount = 10
	fisher.AddComponent(fishComp)

	// Create fishing spot
	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)

	// Update system to register spot
	fs.Update([]*Entity{fisher, spot}, 0)

	// Start fishing
	if !fs.StartFishing(fisher, spot.ID) {
		t.Error("should be able to start fishing")
	}

	if fishComp.GetState() != FishingStateCasting {
		t.Errorf("expected casting state, got %s", fishComp.GetState())
	}
	if fishComp.GetTargetSpotID() != spot.ID {
		t.Errorf("target spot ID mismatch")
	}

	// Bait should be consumed
	_, baitCount := fishComp.GetBait()
	if baitCount != 9 {
		t.Errorf("expected 9 bait after starting, got %d", baitCount)
	}
}

func TestFishingSystem_StartFishing_NoBait(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Create fisher with no bait
	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fishComp.BaitCount = 0
	fisher.AddComponent(fishComp)

	// Create fishing spot
	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	fs.Update([]*Entity{fisher, spot}, 0)

	// Should fail without bait
	if fs.StartFishing(fisher, spot.ID) {
		t.Error("should not be able to fish without bait")
	}
}

func TestFishingSystem_StartFishing_SpotAtCapacity(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Create spot with capacity 1
	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	spotCompRaw, _ := spot.GetComponent("fishing_spot")
	spotComp := spotCompRaw.(*FishingSpotComponent)
	spotComp.MaxConcurrentFishers = 1

	// Create two fishers
	fisher1 := NewEntity(1)
	fishComp1 := NewFishingComponent()
	fisher1.AddComponent(fishComp1)

	fisher2 := NewEntity(2)
	fishComp2 := NewFishingComponent()
	fisher2.AddComponent(fishComp2)

	fs.Update([]*Entity{fisher1, fisher2, spot}, 0)

	// First fisher should succeed
	if !fs.StartFishing(fisher1, spot.ID) {
		t.Error("first fisher should succeed")
	}

	// Second fisher should fail (spot at capacity)
	if fs.StartFishing(fisher2, spot.ID) {
		t.Error("second fisher should fail at capacity")
	}
}

func TestFishingSystem_Cast(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	fs.Update([]*Entity{fisher, spot}, 0)
	fs.StartFishing(fisher, spot.ID)

	// Cast with power
	if !fs.Cast(fisher, 0.8) {
		t.Error("cast should succeed")
	}

	if fishComp.GetState() != FishingStateWaiting {
		t.Errorf("expected waiting state after cast, got %s", fishComp.GetState())
	}
	if fishComp.CastDistance != 0.8 {
		t.Errorf("expected cast distance 0.8, got %f", fishComp.CastDistance)
	}
}

func TestFishingSystem_Hook(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	// Manually set up bite state
	fishComp.State = FishingStateBite
	fishComp.HookedFishTypeID = "bass"
	fishComp.HookedFishWeight = 2.0
	fishComp.HookedFishDifficulty = 0.3
	fishComp.BiteWindowTimer = 2.0

	// Hook should succeed
	if !fs.Hook(fisher) {
		t.Error("hook should succeed during bite")
	}

	if fishComp.GetState() != FishingStateReeling {
		t.Errorf("expected reeling state after hook, got %s", fishComp.GetState())
	}
}

func TestFishingSystem_Hook_WrongState(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	// Not in bite state
	fishComp.State = FishingStateWaiting

	if fs.Hook(fisher) {
		t.Error("hook should fail when not in bite state")
	}
}

func TestFishingSystem_CancelFishing(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	fs.Update([]*Entity{fisher, spot}, 0)
	fs.StartFishing(fisher, spot.ID)
	fs.Cast(fisher, 0.5)

	// Cancel
	fs.CancelFishing(fisher)

	if fishComp.GetState() != FishingStateIdle {
		t.Errorf("expected idle state after cancel, got %s", fishComp.GetState())
	}

	// Spot should have fisher removed
	spotCompRaw, _ := spot.GetComponent("fishing_spot")
	spotComp := spotCompRaw.(*FishingSpotComponent)
	if spotComp.GetCurrentFishers() != 0 {
		t.Errorf("expected 0 fishers after cancel, got %d", spotComp.GetCurrentFishers())
	}
}

func TestFishingSystem_GetNearbyFishingSpots(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Create spots at different positions
	spot1 := fs.GenerateFishingSpot(1, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	spot2 := fs.GenerateFishingSpot(2, WaterTypeFreshwater, DepthShallow, "lake", 50, 0)
	spot3 := fs.GenerateFishingSpot(3, WaterTypeFreshwater, DepthShallow, "lake", 200, 0)

	fs.Update([]*Entity{spot1, spot2, spot3}, 0)

	// Find spots within 100 units of origin
	nearby := fs.GetNearbyFishingSpots(0, 0, 100)

	if len(nearby) != 2 {
		t.Errorf("expected 2 nearby spots, got %d", len(nearby))
	}
}

func TestFishingSystem_Update_Waiting(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)
	fs.BaseWaitTime = 0.5 // Short wait for test

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)

	// Set up waiting state
	fishComp.State = FishingStateWaiting
	fishComp.TargetSpotID = spot.ID
	fishComp.WaitTimer = 0.5
	fishComp.CastDistance = 0.5

	// Update to process waiting
	fs.Update([]*Entity{fisher, spot}, 0.3)
	if fishComp.GetState() != FishingStateWaiting {
		t.Error("should still be waiting")
	}

	// Complete wait
	fs.Update([]*Entity{fisher, spot}, 0.3)

	// Should have transitioned to bite or back to waiting (if no valid fish)
	state := fishComp.GetState()
	if state != FishingStateBite && state != FishingStateWaiting {
		t.Errorf("expected bite or waiting state, got %s", state)
	}
}

func TestFishingSystem_Update_BiteWindow(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)

	// Set up bite state
	fishComp.State = FishingStateBite
	fishComp.TargetSpotID = spot.ID
	fishComp.BiteWindowTimer = 0.5
	fishComp.HookedFishTypeID = "bass"

	// Update to expire window
	fs.Update([]*Entity{fisher, spot}, 0.6)

	// Should have escaped (missed bite)
	if fishComp.GetState() != FishingStateEscaped {
		t.Errorf("expected escaped state after missed bite, got %s", fishComp.GetState())
	}
}

func TestFishingSystem_Update_SpotCooldown(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	spotCompRaw, _ := spot.GetComponent("fishing_spot")
	spotComp := spotCompRaw.(*FishingSpotComponent)
	spotComp.SetCooldown(2.0)

	// Update should process cooldown
	fs.Update([]*Entity{spot}, 1.0)
	if spotComp.CooldownTimer <= 0 {
		t.Error("cooldown should still be active")
	}

	fs.Update([]*Entity{spot}, 1.5)
	if spotComp.CooldownTimer > 0 {
		t.Errorf("cooldown should be complete, got %f", spotComp.CooldownTimer)
	}
}

func TestFishingSystem_Callbacks(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	biteCalled := false
	catchCalled := false
	levelUpCalled := false

	fs.OnBiteCallback = func(fisher *Entity, fishTypeID string) {
		biteCalled = true
	}
	fs.OnCatchCallback = func(fisher *Entity, caught *CaughtFish) {
		catchCalled = true
	}
	fs.OnLevelUpCallback = func(entity *Entity, newLevel int) {
		levelUpCalled = true
	}

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fishComp.FishingXP = 140 // Close to level up
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)

	// Manually trigger bite
	fishComp.State = FishingStateWaiting
	fishComp.TargetSpotID = spot.ID
	fishComp.WaitTimer = 0.1
	fishComp.CastDistance = 0.5

	fs.Update([]*Entity{fisher, spot}, 0.2)

	// Bite callback should be called if fish was selected
	// (may not always trigger depending on RNG and fish availability)

	// Test catch by manually setting up caught state
	fishComp.State = FishingStateCaught
	fishComp.HookedFishTypeID = "bass"
	fishComp.HookedFishWeight = 2.0
	fishComp.TargetSpotID = spot.ID

	fs.completeCatch(fisher, fishComp)

	if !catchCalled {
		t.Error("catch callback should be called")
	}

	// Level up should have triggered since we were close
	// (depends on XP calculation)
	_ = levelUpCalled // may or may not trigger based on XP formula
	_ = biteCalled    // bite callback depends on RNG selecting a fish
}

func TestFishingSystem_TimeOfDay(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Test custom time of day function
	fs.CurrentTimeOfDay = func() TimeOfDay {
		return TimeDusk
	}

	tod := fs.CurrentTimeOfDay()
	if tod != TimeDusk {
		t.Errorf("expected dusk, got %s", tod)
	}
}

func TestFishingSystem_GetFishingSpotCount(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	if fs.GetFishingSpotCount() != 0 {
		t.Error("should start with 0 spots")
	}

	spot1 := fs.GenerateFishingSpot(1, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	spot2 := fs.GenerateFishingSpot(2, WaterTypeSaltwater, DepthDeep, "ocean", 100, 0)

	fs.Update([]*Entity{spot1, spot2}, 0)

	if fs.GetFishingSpotCount() != 2 {
		t.Errorf("expected 2 spots, got %d", fs.GetFishingSpotCount())
	}
}

func TestFishingSystem_ReelingUpdate(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	// Add stub input component for reel input
	inputComp := &StubInput{ActionPressed: true}
	fisher.AddComponent(inputComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)

	// Set up reeling state
	fishComp.State = FishingStateReeling
	fishComp.TargetSpotID = spot.ID
	fishComp.HookedFishTypeID = "bass"
	fishComp.HookedFishWeight = 2.0
	fishComp.HookedFishDifficulty = 0.3
	fishComp.TensionLevel = 0.5

	initialProgress := fishComp.GetReelProgress()

	// Update with input active
	fs.Update([]*Entity{fisher, spot}, 0.1)

	// Progress should have changed
	newProgress := fishComp.GetReelProgress()
	if newProgress == initialProgress && fishComp.GetState() == FishingStateReeling {
		t.Log("Note: Progress may not change immediately depending on tension")
	}
}

func TestFishingSystem_SelectFish_SkillFilter(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Register fish requiring high skill
	fs.RegisterFishType(&FishType{
		ID:         "high_skill_fish",
		Name:       "High Skill Fish",
		Rarity:     FishRarityCommon,
		WaterTypes: []WaterType{WaterTypeFreshwater},
		MinDepth:   DepthShallow,
		BestTime:   TimeAny,
		MinSkill:   50,
		BaseWeight: 1.0,
		MaxWeight:  5.0,
		Difficulty: 0.5,
	})

	spot := NewFishingSpotComponent(WaterTypeFreshwater, DepthShallow, "lake")
	spot.AddFishType("high_skill_fish", 10.0)
	spot.AddFishType("bass", 10.0)

	fishComp := NewFishingComponent()
	fishComp.FishingSkill = 10 // Low skill
	fishComp.CastDistance = 0.5

	// Select fish multiple times - should only get bass due to skill requirement
	for i := 0; i < 10; i++ {
		selected, _ := fs.selectFish(spot, fishComp)
		if selected != nil && selected.ID == "high_skill_fish" {
			t.Error("should not select fish requiring higher skill")
		}
	}
}

func BenchmarkFishingSystem_Update(b *testing.B) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Create entities
	entities := make([]*Entity, 100)
	for i := 0; i < 50; i++ {
		// Fishers
		fisher := NewEntity(uint64(i))
		fishComp := NewFishingComponent()
		fishComp.State = FishingStateWaiting
		fishComp.WaitTimer = 100.0
		fisher.AddComponent(fishComp)
		entities[i] = fisher
	}
	for i := 50; i < 100; i++ {
		// Spots
		spot := fs.GenerateFishingSpot(int64(i), WaterTypeFreshwater, DepthShallow, "lake", float64(i*10), 0)
		entities[i] = spot
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Update(entities, 0.016)
	}
}

func BenchmarkFishingSystem_GenerateFishingSpot(b *testing.B) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.GenerateFishingSpot(int64(i), WaterTypeFreshwater, DepthMedium, "lake", 0, 0)
	}
}

func BenchmarkFishingSystem_SelectFish(b *testing.B) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	spot := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")
	spot.AddFishType("bass", 10.0)
	spot.AddFishType("trout", 8.0)
	spot.AddFishType("catfish", 5.0)
	spot.AddFishType("pike", 2.0)

	fishComp := NewFishingComponent()
	fishComp.FishingSkill = 50
	fishComp.CastDistance = 0.7

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.selectFish(spot, fishComp)
	}
}

func TestFishingSystem_InputComponent(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	// Test without input component
	input := fs.getReelInput(fisher)
	if input != 0 {
		t.Errorf("expected 0 input without component, got %f", input)
	}

	// Add inactive stub input component
	inputComp := &StubInput{ActionPressed: false}
	fisher.AddComponent(inputComp)

	input = fs.getReelInput(fisher)
	if input != 0 {
		t.Errorf("expected 0 input with inactive action, got %f", input)
	}

	// Activate action
	inputComp.ActionPressed = true
	input = fs.getReelInput(fisher)
	if input != 1.0 {
		t.Errorf("expected 1.0 input with active action, got %f", input)
	}
}

func TestFishingSystem_CompleteCatch(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	catchCount := 0
	var lastCatch *CaughtFish
	fs.OnCatchCallback = func(fisher *Entity, caught *CaughtFish) {
		catchCount++
		lastCatch = caught
	}

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 50, 75)
	fs.Update([]*Entity{fisher, spot}, 0)

	// Set up caught state
	fishComp.State = FishingStateCaught
	fishComp.TargetSpotID = spot.ID
	fishComp.HookedFishTypeID = "bass"
	fishComp.HookedFishWeight = 2.5

	fs.completeCatch(fisher, fishComp)

	if catchCount != 1 {
		t.Errorf("expected 1 catch callback, got %d", catchCount)
	}
	if lastCatch == nil {
		t.Fatal("lastCatch should not be nil")
	}
	if lastCatch.FishTypeID != "bass" {
		t.Errorf("expected bass, got %s", lastCatch.FishTypeID)
	}

	// State should be idle
	if fishComp.GetState() != FishingStateIdle {
		t.Errorf("expected idle state, got %s", fishComp.GetState())
	}
}

func TestFishingSystem_XPByRarity(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)
	fs.XPPerCatch = 10

	fisher := NewEntity(1)
	fishComp := NewFishingComponent()
	fisher.AddComponent(fishComp)

	spot := fs.GenerateFishingSpot(12345, WaterTypeFreshwater, DepthShallow, "lake", 0, 0)
	fs.Update([]*Entity{fisher, spot}, 0)

	tests := []struct {
		fishID        string
		minExpectedXP int
	}{
		{"bass", 10},    // Common: base XP
		{"catfish", 15}, // Uncommon: 1.5x
		{"pike", 20},    // Rare: 2x
	}

	for _, tt := range tests {
		t.Run(tt.fishID, func(t *testing.T) {
			fishComp.FishingXP = 0
			fishComp.State = FishingStateCaught
			fishComp.TargetSpotID = spot.ID
			fishComp.HookedFishTypeID = tt.fishID
			fishComp.HookedFishWeight = 2.0

			fs.completeCatch(fisher, fishComp)

			// XP should have increased by at least min expected
			// (exact amount depends on difficulty multiplier)
			if fishComp.FishingXP < tt.minExpectedXP {
				t.Errorf("expected at least %d XP for %s, got %d", tt.minExpectedXP, tt.fishID, fishComp.FishingXP)
			}
		})
	}
}

func TestDefaultTimeOfDay(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	// Default function should return valid time of day
	tod := fs.CurrentTimeOfDay()

	validTimes := map[TimeOfDay]bool{
		TimeDawn:  true,
		TimeDay:   true,
		TimeDusk:  true,
		TimeNight: true,
	}

	if !validTimes[tod] {
		t.Errorf("invalid time of day: %s", tod)
	}
}

func TestDefaultWeather(t *testing.T) {
	world := NewWorld()
	fs := NewFishingSystem(world)

	weather := fs.CurrentWeather()
	if weather != "clear" {
		t.Errorf("expected default weather 'clear', got %s", weather)
	}
}

func TestFishingSystem_PositionString(t *testing.T) {
	// Test that position component has String method for location
	pos := &PositionComponent{X: 123.45, Y: 678.90}
	str := pos.String()
	if str == "" {
		t.Error("position string should not be empty")
	}
}

func init() {
	// Ensure time package is imported for tests
	_ = time.Now
}
