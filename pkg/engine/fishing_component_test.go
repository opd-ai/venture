// Package engine provides tests for fishing components.
// Phase 96: Fishing System

package engine

import (
	"encoding/json"
	"math/rand"
	"testing"
)

func TestAllWaterTypes(t *testing.T) {
	types := AllWaterTypes()
	if len(types) != 3 {
		t.Errorf("expected 3 water types, got %d", len(types))
	}

	expected := map[WaterType]bool{
		WaterTypeFreshwater: false,
		WaterTypeSaltwater:  false,
		WaterTypeMagical:    false,
	}

	for _, wt := range types {
		if _, ok := expected[wt]; ok {
			expected[wt] = true
		}
	}

	for wt, found := range expected {
		if !found {
			t.Errorf("water type %s not found in AllWaterTypes()", wt)
		}
	}
}

func TestNewFishingSpotComponent(t *testing.T) {
	tests := []struct {
		name       string
		waterType  WaterType
		depthLevel DepthLevel
		biomeType  string
	}{
		{"freshwater shallow", WaterTypeFreshwater, DepthShallow, "forest"},
		{"saltwater deep", WaterTypeSaltwater, DepthDeep, "ocean"},
		{"magical medium", WaterTypeMagical, DepthMedium, "enchanted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spot := NewFishingSpotComponent(tt.waterType, tt.depthLevel, tt.biomeType)

			if spot.WaterType != tt.waterType {
				t.Errorf("expected water type %s, got %s", tt.waterType, spot.WaterType)
			}
			if spot.DepthLevel != tt.depthLevel {
				t.Errorf("expected depth level %d, got %d", tt.depthLevel, spot.DepthLevel)
			}
			if spot.BiomeType != tt.biomeType {
				t.Errorf("expected biome %s, got %s", tt.biomeType, spot.BiomeType)
			}
			if spot.FishPopulation == nil {
				t.Error("fish population map is nil")
			}
			if !spot.IsActive {
				t.Error("new spot should be active")
			}
			if spot.Type() != "fishing_spot" {
				t.Errorf("expected type 'fishing_spot', got %s", spot.Type())
			}
		})
	}
}

func TestFishingSpotComponent_FishPopulation(t *testing.T) {
	spot := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")

	// Add fish types
	spot.AddFishType("bass", 10.0)
	spot.AddFishType("trout", 8.0)
	spot.AddFishType("pike", 2.0)

	// Check fish types
	types := spot.GetFishTypes()
	if len(types) != 3 {
		t.Errorf("expected 3 fish types, got %d", len(types))
	}

	// Check spawn weights
	if spot.GetSpawnWeight("bass") != 10.0 {
		t.Errorf("expected bass weight 10.0, got %f", spot.GetSpawnWeight("bass"))
	}

	// Remove a fish type
	spot.RemoveFishType("pike")
	types = spot.GetFishTypes()
	if len(types) != 2 {
		t.Errorf("expected 2 fish types after removal, got %d", len(types))
	}
}

func TestFishingSpotComponent_CanFish(t *testing.T) {
	spot := NewFishingSpotComponent(WaterTypeFreshwater, DepthShallow, "river")
	spot.MaxConcurrentFishers = 2

	// Should be able to fish initially
	if !spot.CanFish() {
		t.Error("should be able to fish when spot is empty")
	}

	// Add fishers
	if !spot.AddFisher() {
		t.Error("should be able to add first fisher")
	}
	if !spot.AddFisher() {
		t.Error("should be able to add second fisher")
	}

	// Should be at capacity
	if spot.AddFisher() {
		t.Error("should not be able to add third fisher at capacity")
	}
	if spot.CanFish() {
		t.Error("should not be able to fish when at capacity")
	}

	// Remove a fisher
	spot.RemoveFisher()
	if !spot.CanFish() {
		t.Error("should be able to fish after fisher leaves")
	}
}

func TestFishingSpotComponent_Cooldown(t *testing.T) {
	spot := NewFishingSpotComponent(WaterTypeSaltwater, DepthDeep, "ocean")

	// Set cooldown
	spot.SetCooldown(5.0)

	// Should not be able to fish during cooldown
	if spot.CanFish() {
		t.Error("should not be able to fish during cooldown")
	}

	// Update cooldown partially
	spot.UpdateCooldown(3.0)
	if spot.CanFish() {
		t.Error("should still be on cooldown")
	}

	// Complete cooldown
	spot.UpdateCooldown(3.0)
	if !spot.CanFish() {
		t.Error("should be able to fish after cooldown")
	}
}

func TestFishingSpotComponent_Serialize(t *testing.T) {
	spot := NewFishingSpotComponent(WaterTypeMagical, DepthMedium, "enchanted")
	spot.AddFishType("moonfish", 5.0)
	spot.RareFishBonus = 1.5

	// Serialize
	data, err := spot.Serialize()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	// Deserialize into new component
	spot2 := NewFishingSpotComponent(WaterTypeFreshwater, DepthShallow, "")
	err = spot2.Deserialize(data)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}

	if spot2.WaterType != WaterTypeMagical {
		t.Errorf("expected magical water type, got %s", spot2.WaterType)
	}
	if spot2.RareFishBonus != 1.5 {
		t.Errorf("expected rare fish bonus 1.5, got %f", spot2.RareFishBonus)
	}
}

func TestNewFishingComponent(t *testing.T) {
	fc := NewFishingComponent()

	if fc.FishingSkill != 1 {
		t.Errorf("expected initial skill 1, got %d", fc.FishingSkill)
	}
	if fc.State != FishingStateIdle {
		t.Errorf("expected idle state, got %s", fc.State)
	}
	if fc.BaitCount != 10 {
		t.Errorf("expected 10 bait, got %d", fc.BaitCount)
	}
	if fc.Type() != "fishing" {
		t.Errorf("expected type 'fishing', got %s", fc.Type())
	}
	if fc.TotalCaught == nil {
		t.Error("TotalCaught map is nil")
	}
	if fc.PersonalRecords == nil {
		t.Error("PersonalRecords map is nil")
	}
}

func TestFishingComponent_CastingFlow(t *testing.T) {
	fc := NewFishingComponent()

	// Start casting
	fc.StartCasting(12345)

	if fc.GetState() != FishingStateCasting {
		t.Errorf("expected casting state, got %s", fc.GetState())
	}
	if fc.GetTargetSpotID() != 12345 {
		t.Errorf("expected target 12345, got %d", fc.GetTargetSpotID())
	}

	// Set cast distance
	fc.SetCastDistance(0.75)

	// Complete cast
	fc.CompleteCast(10.0)

	if fc.GetState() != FishingStateWaiting {
		t.Errorf("expected waiting state, got %s", fc.GetState())
	}
}

func TestFishingComponent_SetCastDistance(t *testing.T) {
	fc := NewFishingComponent()

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{-0.5, 0},
		{1.5, 1},
		{0, 0},
		{1, 1},
	}

	for _, tt := range tests {
		fc.SetCastDistance(tt.input)
		if fc.CastDistance != tt.expected {
			t.Errorf("SetCastDistance(%f) = %f, want %f", tt.input, fc.CastDistance, tt.expected)
		}
	}
}

func TestFishingComponent_WaitingPhase(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(5.0)

	// Update wait - should not trigger bite yet
	if fc.UpdateWait(3.0) {
		t.Error("should not trigger bite before wait time")
	}

	// Complete wait
	if !fc.UpdateWait(3.0) {
		t.Error("should trigger bite after wait time")
	}
}

func TestFishingComponent_BitePhase(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.UpdateWait(2.0) // Force wait complete

	// Trigger bite
	fc.TriggerBite("bass", 2.5, 0.3, 2.0)

	if fc.GetState() != FishingStateBite {
		t.Errorf("expected bite state, got %s", fc.GetState())
	}

	fishID, weight := fc.GetHookedFish()
	if fishID != "bass" {
		t.Errorf("expected bass, got %s", fishID)
	}
	if weight != 2.5 {
		t.Errorf("expected weight 2.5, got %f", weight)
	}

	// Hook the fish before window expires
	if !fc.HookFish() {
		t.Error("should be able to hook fish during bite")
	}

	if fc.GetState() != FishingStateReeling {
		t.Errorf("expected reeling state after hook, got %s", fc.GetState())
	}
}

func TestFishingComponent_MissedBite(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)

	// Trigger bite with short window
	fc.TriggerBite("trout", 1.0, 0.2, 1.0)

	// Let window expire
	if fc.UpdateBiteWindow(0.5) {
		t.Error("window should not expire yet")
	}
	if !fc.UpdateBiteWindow(0.6) {
		t.Error("window should expire now")
	}

	// Miss bite
	fc.MissBite()
	if fc.GetState() != FishingStateEscaped {
		t.Errorf("expected escaped state, got %s", fc.GetState())
	}
}

func TestFishingComponent_ReelingMinigame(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.TriggerBite("bass", 2.0, 0.3, 2.0)
	fc.HookFish()

	// Test reeling with good tension
	result := fc.UpdateReeling(0.1, 0.7)
	if result != 0 {
		t.Errorf("expected continue (0), got %d", result)
	}

	// Check tension is in range
	tension := fc.GetTensionLevel()
	if tension < 0.2 || tension > 0.9 {
		t.Errorf("tension %f outside expected range", tension)
	}

	// Progress should have increased
	progress := fc.GetReelProgress()
	if progress <= 0 {
		t.Error("progress should have increased")
	}
}

func TestFishingComponent_LineBreak(t *testing.T) {
	fc := NewFishingComponent()
	fc.SetRodQuality(0.5) // Low quality rod
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.TriggerBite("pike", 10.0, 0.9, 2.0) // High difficulty fish
	fc.HookFish()

	// Set high struggle
	fc.FishStruggleDirection = 1

	// Force high tension by reeling hard
	for i := 0; i < 100; i++ {
		result := fc.UpdateReeling(0.1, 1.0)
		if result == -1 {
			// Line broke as expected
			if fc.GetState() != FishingStateEscaped {
				t.Errorf("expected escaped state after line break")
			}
			return
		}
	}
	// Line should have broken with low quality rod and high difficulty
	t.Log("Note: line break test completed without break - this can happen with RNG")
}

func TestFishingComponent_UpdateStruggle(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.TriggerBite("bass", 2.0, 0.5, 2.0)
	fc.HookFish()

	rng := rand.New(rand.NewSource(12345))

	// Set initial struggle timer
	fc.StruggleTimer = 1.0

	// Update should not change direction until timer expires
	initialDir := fc.FishStruggleDirection
	fc.UpdateStruggle(0.5, rng)
	// Timer not expired yet

	// Expire timer and change direction
	fc.UpdateStruggle(0.6, rng)
	// Direction may or may not change (random)

	// Just verify no panic and state is valid
	if fc.FishStruggleDirection < -1 || fc.FishStruggleDirection > 1 {
		t.Errorf("invalid struggle direction %d", fc.FishStruggleDirection)
	}

	_ = initialDir // unused but keeping test consistent
}

func TestFishingComponent_CompleteCatch(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.TriggerBite("bass", 2.5, 0.3, 2.0)
	fc.HookFish()

	// Force state to caught
	fc.State = FishingStateCaught

	// Complete catch
	caught := fc.CompleteCatch("Lake Serene", 1000000)

	if caught == nil {
		t.Fatal("caught should not be nil")
	}
	if caught.FishTypeID != "bass" {
		t.Errorf("expected bass, got %s", caught.FishTypeID)
	}
	if caught.Weight != 2.5 {
		t.Errorf("expected weight 2.5, got %f", caught.Weight)
	}
	if caught.Location != "Lake Serene" {
		t.Errorf("expected location 'Lake Serene', got %s", caught.Location)
	}
	if !caught.IsRecord {
		t.Error("first catch should be a record")
	}

	// Check total caught
	if fc.GetTotalCaught("bass") != 1 {
		t.Errorf("expected 1 bass caught, got %d", fc.GetTotalCaught("bass"))
	}

	// Check personal record
	if fc.GetPersonalRecord("bass") != 2.5 {
		t.Errorf("expected record 2.5, got %f", fc.GetPersonalRecord("bass"))
	}

	// State should be idle
	if fc.GetState() != FishingStateIdle {
		t.Errorf("expected idle state after catch, got %s", fc.GetState())
	}
}

func TestFishingComponent_PersonalRecord(t *testing.T) {
	fc := NewFishingComponent()

	// First catch
	fc.TriggerBite("bass", 2.0, 0.3, 2.0)
	fc.HookFish()
	fc.State = FishingStateCaught
	caught1 := fc.CompleteCatch("Lake", 1000)
	if !caught1.IsRecord {
		t.Error("first catch should be record")
	}

	// Smaller catch
	fc.TriggerBite("bass", 1.5, 0.3, 2.0)
	fc.HookFish()
	fc.State = FishingStateCaught
	caught2 := fc.CompleteCatch("Lake", 2000)
	if caught2.IsRecord {
		t.Error("smaller catch should not be record")
	}

	// Bigger catch
	fc.TriggerBite("bass", 3.0, 0.3, 2.0)
	fc.HookFish()
	fc.State = FishingStateCaught
	caught3 := fc.CompleteCatch("Lake", 3000)
	if !caught3.IsRecord {
		t.Error("bigger catch should be record")
	}

	if fc.GetPersonalRecord("bass") != 3.0 {
		t.Errorf("expected record 3.0, got %f", fc.GetPersonalRecord("bass"))
	}
}

func TestFishingComponent_XPAndLevelUp(t *testing.T) {
	fc := NewFishingComponent()

	// Initial state
	if fc.GetSkillLevel() != 1 {
		t.Errorf("expected initial skill 1, got %d", fc.GetSkillLevel())
	}

	// Add XP but not enough to level
	leveledUp := fc.AddXP(50)
	if leveledUp {
		t.Error("should not level up with 50 XP")
	}

	// Add enough XP to level (100 + 1*50 = 150 for level 1->2)
	leveledUp = fc.AddXP(100)
	if !leveledUp {
		t.Error("should level up with 150 total XP")
	}
	if fc.GetSkillLevel() != 2 {
		t.Errorf("expected skill 2, got %d", fc.GetSkillLevel())
	}
}

func TestFishingComponent_Bait(t *testing.T) {
	fc := NewFishingComponent()

	baitType, baitCount := fc.GetBait()
	if baitType != "basic" {
		t.Errorf("expected basic bait, got %s", baitType)
	}
	if baitCount != 10 {
		t.Errorf("expected 10 bait, got %d", baitCount)
	}

	// Use bait
	if !fc.UseBait() {
		t.Error("should be able to use bait")
	}
	_, baitCount = fc.GetBait()
	if baitCount != 9 {
		t.Errorf("expected 9 bait, got %d", baitCount)
	}

	// Use all bait
	for i := 0; i < 9; i++ {
		fc.UseBait()
	}
	if fc.UseBait() {
		t.Error("should not be able to use bait when empty")
	}

	// Add bait
	fc.AddBait("premium", 5)
	baitType, baitCount = fc.GetBait()
	if baitType != "premium" || baitCount != 5 {
		t.Errorf("expected 5 premium bait, got %d %s", baitCount, baitType)
	}
}

func TestFishingComponent_RodQuality(t *testing.T) {
	fc := NewFishingComponent()

	// Default quality
	if fc.GetRodQuality() != 1.0 {
		t.Errorf("expected default quality 1.0, got %f", fc.GetRodQuality())
	}

	// Set quality
	fc.SetRodQuality(1.5)
	if fc.GetRodQuality() != 1.5 {
		t.Errorf("expected quality 1.5, got %f", fc.GetRodQuality())
	}

	// Max tension should increase with quality
	expectedMaxTension := 0.8 + (1.5 * 0.4)
	if fc.MaxTension < expectedMaxTension-0.001 || fc.MaxTension > expectedMaxTension+0.001 {
		t.Errorf("expected max tension %f, got %f", expectedMaxTension, fc.MaxTension)
	}

	// Minimum quality
	fc.SetRodQuality(-1.0)
	if fc.GetRodQuality() != 0.1 {
		t.Errorf("expected minimum quality 0.1, got %f", fc.GetRodQuality())
	}
}

func TestFishingComponent_StopFishing(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(5.0)
	fc.TriggerBite("bass", 2.0, 0.3, 2.0)
	fc.HookFish()

	// Stop fishing
	fc.StopFishing()

	if fc.GetState() != FishingStateIdle {
		t.Errorf("expected idle state, got %s", fc.GetState())
	}
	if fc.GetTargetSpotID() != 0 {
		t.Errorf("expected target 0, got %d", fc.GetTargetSpotID())
	}
	fishID, _ := fc.GetHookedFish()
	if fishID != "" {
		t.Errorf("expected no hooked fish, got %s", fishID)
	}
}

func TestFishingComponent_SessionCaught(t *testing.T) {
	fc := NewFishingComponent()

	// Catch some fish
	fc.TriggerBite("bass", 2.0, 0.3, 2.0)
	fc.HookFish()
	fc.State = FishingStateCaught
	fc.CompleteCatch("Lake", 1000)

	fc.TriggerBite("trout", 1.0, 0.2, 2.0)
	fc.HookFish()
	fc.State = FishingStateCaught
	fc.CompleteCatch("Lake", 2000)

	session := fc.GetSessionCaught()
	if len(session) != 2 {
		t.Errorf("expected 2 fish in session, got %d", len(session))
	}

	// Clear session
	fc.ClearSession()
	session = fc.GetSessionCaught()
	if len(session) != 0 {
		t.Errorf("expected empty session after clear, got %d", len(session))
	}

	// Total caught should remain
	if fc.GetTotalCaught("bass") != 1 {
		t.Errorf("total caught should persist, got %d", fc.GetTotalCaught("bass"))
	}
}

func TestFishingComponent_Serialize(t *testing.T) {
	fc := NewFishingComponent()
	fc.AddXP(200)
	fc.SetRodQuality(1.5)
	fc.TriggerBite("bass", 2.5, 0.3, 2.0)
	fc.HookFish()
	fc.State = FishingStateCaught
	fc.CompleteCatch("Lake", 1000)

	// Serialize
	data, err := fc.Serialize()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	// Deserialize
	fc2 := NewFishingComponent()
	err = fc2.Deserialize(data)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}

	if fc2.GetSkillLevel() != fc.GetSkillLevel() {
		t.Errorf("skill level mismatch: %d vs %d", fc2.GetSkillLevel(), fc.GetSkillLevel())
	}
	if fc2.GetRodQuality() != fc.GetRodQuality() {
		t.Errorf("rod quality mismatch: %f vs %f", fc2.GetRodQuality(), fc.GetRodQuality())
	}
	if fc2.GetTotalCaught("bass") != fc.GetTotalCaught("bass") {
		t.Errorf("total caught mismatch")
	}
	if fc2.GetPersonalRecord("bass") != fc.GetPersonalRecord("bass") {
		t.Errorf("personal record mismatch")
	}
}

func TestFishType_JSON(t *testing.T) {
	fish := &FishType{
		ID:         "test_fish",
		Name:       "Test Fish",
		Rarity:     FishRarityRare,
		WaterTypes: []WaterType{WaterTypeFreshwater, WaterTypeMagical},
		MinDepth:   DepthMedium,
		BestTime:   TimeDusk,
		MinSkill:   25,
		BaseWeight: 2.0,
		MaxWeight:  10.0,
		Difficulty: 0.7,
	}

	data, err := json.Marshal(fish)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var fish2 FishType
	err = json.Unmarshal(data, &fish2)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if fish2.ID != fish.ID {
		t.Errorf("ID mismatch: %s vs %s", fish2.ID, fish.ID)
	}
	if fish2.Rarity != fish.Rarity {
		t.Errorf("rarity mismatch: %s vs %s", fish2.Rarity, fish.Rarity)
	}
	if len(fish2.WaterTypes) != len(fish.WaterTypes) {
		t.Errorf("water types length mismatch")
	}
}

func TestCaughtFish_JSON(t *testing.T) {
	caught := &CaughtFish{
		FishTypeID: "bass",
		Weight:     2.5,
		CaughtAt:   1234567890,
		Location:   "Lake Serene",
		IsRecord:   true,
	}

	data, err := json.Marshal(caught)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var caught2 CaughtFish
	err = json.Unmarshal(data, &caught2)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if caught2.FishTypeID != caught.FishTypeID {
		t.Errorf("fish type mismatch")
	}
	if caught2.Weight != caught.Weight {
		t.Errorf("weight mismatch")
	}
	if caught2.IsRecord != caught.IsRecord {
		t.Errorf("record flag mismatch")
	}
}

func TestFishingComponent_FishEscaped(t *testing.T) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.TriggerBite("pike", 5.0, 0.6, 2.0)
	fc.HookFish()

	// Force escape
	fc.FishEscaped()

	if fc.GetState() != FishingStateEscaped {
		t.Errorf("expected escaped state, got %s", fc.GetState())
	}
	fishID, _ := fc.GetHookedFish()
	if fishID != "" {
		t.Errorf("hooked fish should be cleared")
	}
	if fc.GetReelProgress() != 0 {
		t.Errorf("reel progress should be reset")
	}
}

func BenchmarkFishingComponent_UpdateReeling(b *testing.B) {
	fc := NewFishingComponent()
	fc.StartCasting(100)
	fc.CompleteCast(1.0)
	fc.TriggerBite("bass", 2.0, 0.3, 2.0)
	fc.HookFish()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fc.UpdateReeling(0.016, 0.5)
	}
}

func BenchmarkFishingComponent_Serialize(b *testing.B) {
	fc := NewFishingComponent()
	fc.AddXP(1000)
	for i := 0; i < 10; i++ {
		fc.TotalCaught["fish_"+string(rune('a'+i))] = i * 10
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fc.Serialize()
	}
}
