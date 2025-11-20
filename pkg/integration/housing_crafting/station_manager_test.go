package housing_crafting

import (
	"fmt"
	"sync"
	"testing"
)

// TestNewStationManager tests station manager creation
func TestNewStationManager(t *testing.T) {
	sm := NewStationManager()
	if sm == nil {
		t.Fatal("NewStationManager() returned nil")
	}
	if sm.stations == nil {
		t.Error("stations map is nil")
	}
	if sm.stationsByOwner == nil {
		t.Error("stationsByOwner map is nil")
	}
	if sm.stationsByHouse == nil {
		t.Error("stationsByHouse map is nil")
	}
	if sm.facilities == nil {
		t.Error("facilities map is nil")
	}
	if sm.facilitiesByOwner == nil {
		t.Error("facilitiesByOwner map is nil")
	}
}

// TestRegisterStation tests station registration
func TestRegisterStation(t *testing.T) {
	sm := NewStationManager()

	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeForge,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
	}

	// Test successful registration
	if err := sm.RegisterStation(station); err != nil {
		t.Errorf("RegisterStation() error = %v", err)
	}

	// Test duplicate ID
	if err := sm.RegisterStation(station); err == nil {
		t.Error("RegisterStation() should error on duplicate ID")
	}

	// Test nil station
	if err := sm.RegisterStation(nil); err == nil {
		t.Error("RegisterStation() should error on nil station")
	}

	// Test empty ID
	invalidStation := &CraftingStation{
		ID:      "",
		OwnerID: "player1",
		HouseID: "house1",
	}
	if err := sm.RegisterStation(invalidStation); err == nil {
		t.Error("RegisterStation() should error on empty ID")
	}

	// Test empty OwnerID
	invalidStation.ID = "station2"
	invalidStation.OwnerID = ""
	if err := sm.RegisterStation(invalidStation); err == nil {
		t.Error("RegisterStation() should error on empty OwnerID")
	}

	// Test empty HouseID
	invalidStation.OwnerID = "player1"
	invalidStation.HouseID = ""
	if err := sm.RegisterStation(invalidStation); err == nil {
		t.Error("RegisterStation() should error on empty HouseID")
	}
}

// TestUnregisterStation tests station removal
func TestUnregisterStation(t *testing.T) {
	sm := NewStationManager()

	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeForge,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
	}

	// Register then unregister
	sm.RegisterStation(station)
	if err := sm.UnregisterStation("station1"); err != nil {
		t.Errorf("UnregisterStation() error = %v", err)
	}

	// Test unregister non-existent
	if err := sm.UnregisterStation("station1"); err == nil {
		t.Error("UnregisterStation() should error on non-existent station")
	}

	// Verify station removed from all maps
	if _, err := sm.GetStation("station1"); err == nil {
		t.Error("Station still exists after unregister")
	}
	if len(sm.GetStationsByOwner("player1")) != 0 {
		t.Error("Station still in owner map after unregister")
	}
	if len(sm.GetStationsByHouse("house1")) != 0 {
		t.Error("Station still in house map after unregister")
	}
}

// TestGetStation tests station retrieval
func TestGetStation(t *testing.T) {
	sm := NewStationManager()

	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeForge,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
	}

	sm.RegisterStation(station)

	// Test successful retrieval
	retrieved, err := sm.GetStation("station1")
	if err != nil {
		t.Errorf("GetStation() error = %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetStation() returned nil")
	}
	if retrieved.ID != station.ID {
		t.Errorf("GetStation() ID = %v, want %v", retrieved.ID, station.ID)
	}

	// Test non-existent station
	if _, err := sm.GetStation("station2"); err == nil {
		t.Error("GetStation() should error on non-existent station")
	}
}

// TestGetStationsByOwner tests owner-based retrieval
func TestGetStationsByOwner(t *testing.T) {
	sm := NewStationManager()

	// Register multiple stations for same owner
	for i := 1; i <= 3; i++ {
		station := &CraftingStation{
			ID:      fmt.Sprintf("station%d", i),
			Type:    StationTypeForge,
			Quality: QualityMaster,
			OwnerID: "player1",
			HouseID: "house1",
		}
		sm.RegisterStation(station)
	}

	// Test retrieval
	stations := sm.GetStationsByOwner("player1")
	if len(stations) != 3 {
		t.Errorf("GetStationsByOwner() count = %v, want 3", len(stations))
	}

	// Test non-existent owner
	stations = sm.GetStationsByOwner("player2")
	if len(stations) != 0 {
		t.Errorf("GetStationsByOwner() for non-existent owner should return empty slice, got %v", len(stations))
	}
}

// TestGetStationsByHouse tests house-based retrieval
func TestGetStationsByHouse(t *testing.T) {
	sm := NewStationManager()

	// Register multiple stations for same house
	for i := 1; i <= 3; i++ {
		station := &CraftingStation{
			ID:      fmt.Sprintf("station%d", i),
			Type:    StationTypeForge,
			Quality: QualityMaster,
			OwnerID: "player1",
			HouseID: "house1",
		}
		sm.RegisterStation(station)
	}

	// Test retrieval
	stations := sm.GetStationsByHouse("house1")
	if len(stations) != 3 {
		t.Errorf("GetStationsByHouse() count = %v, want 3", len(stations))
	}

	// Test non-existent house
	stations = sm.GetStationsByHouse("house2")
	if len(stations) != 0 {
		t.Errorf("GetStationsByHouse() for non-existent house should return empty slice, got %v", len(stations))
	}
}

// TestGetCraftingBonus tests crafting bonus calculation
func TestGetCraftingBonus(t *testing.T) {
	sm := NewStationManager()

	// Register stations with different quality tiers
	stationBasic := &CraftingStation{
		ID:            "station1",
		Type:          StationTypeForge,
		Quality:       QualityBasic,
		OwnerID:       "player1",
		HouseID:       "house1",
		ActiveRecipes: []string{"sword_recipe"},
	}
	stationMaster := &CraftingStation{
		ID:            "station2",
		Type:          StationTypeForge,
		Quality:       QualityMaster,
		OwnerID:       "player1",
		HouseID:       "house1",
		ActiveRecipes: []string{"sword_recipe"},
	}

	sm.RegisterStation(stationBasic)
	sm.RegisterStation(stationMaster)

	// Test bonus calculation (should return highest quality)
	bonus := sm.GetCraftingBonus("player1", "sword_recipe")
	if bonus != 2.0 {
		t.Errorf("GetCraftingBonus() = %v, want 2.0", bonus)
	}

	// Test non-existent recipe
	bonus = sm.GetCraftingBonus("player1", "shield_recipe")
	if bonus != 1.0 {
		t.Errorf("GetCraftingBonus() for non-existent recipe = %v, want 1.0", bonus)
	}

	// Test non-existent player
	bonus = sm.GetCraftingBonus("player2", "sword_recipe")
	if bonus != 1.0 {
		t.Errorf("GetCraftingBonus() for non-existent player = %v, want 1.0", bonus)
	}
}

// TestUnlockRecipes tests recipe unlocking
func TestUnlockRecipes(t *testing.T) {
	sm := NewStationManager()

	tests := []struct {
		name          string
		stationType   StationType
		quality       QualityTier
		expectedCount int
	}{
		{"Basic", StationTypeForge, QualityBasic, 10},
		{"Standard", StationTypeForge, QualityStandard, 20},
		{"Advanced", StationTypeForge, QualityAdvanced, 30},
		{"Master", StationTypeForge, QualityMaster, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipes := sm.UnlockRecipes(tt.stationType, tt.quality)
			if len(recipes) != tt.expectedCount {
				t.Errorf("UnlockRecipes() count = %v, want %v", len(recipes), tt.expectedCount)
			}
		})
	}
}

// TestRegisterFacility tests facility registration
func TestRegisterFacility(t *testing.T) {
	sm := NewStationManager()

	facility := &SkillTrainingFacility{
		ID:           "facility1",
		OwnerID:      "player1",
		HouseID:      "house1",
		XPMultiplier: 1.5,
	}

	// Test successful registration
	if err := sm.RegisterFacility(facility); err != nil {
		t.Errorf("RegisterFacility() error = %v", err)
	}

	// Test duplicate ID
	if err := sm.RegisterFacility(facility); err == nil {
		t.Error("RegisterFacility() should error on duplicate ID")
	}

	// Test nil facility
	if err := sm.RegisterFacility(nil); err == nil {
		t.Error("RegisterFacility() should error on nil facility")
	}

	// Test empty ID
	invalidFacility := &SkillTrainingFacility{
		ID:      "",
		OwnerID: "player1",
	}
	if err := sm.RegisterFacility(invalidFacility); err == nil {
		t.Error("RegisterFacility() should error on empty ID")
	}

	// Test empty OwnerID
	invalidFacility.ID = "facility2"
	invalidFacility.OwnerID = ""
	if err := sm.RegisterFacility(invalidFacility); err == nil {
		t.Error("RegisterFacility() should error on empty OwnerID")
	}
}

// TestGetSkillTrainingBonus tests skill training bonus calculation
func TestGetSkillTrainingBonus(t *testing.T) {
	sm := NewStationManager()

	// Register facility with smithing training
	facility := &SkillTrainingFacility{
		ID:              "facility1",
		OwnerID:         "player1",
		HouseID:         "house1",
		TrainableSkills: []string{"smithing"},
		XPMultiplier:    1.5,
	}
	sm.RegisterFacility(facility)

	// Register station with alchemy bonus
	station := &CraftingStation{
		ID:      "station1",
		Type:    StationTypeAlchemy,
		Quality: QualityMaster,
		OwnerID: "player1",
		HouseID: "house1",
		SkillBonus: map[string]int{
			"alchemy": 50,
		},
	}
	sm.RegisterStation(station)

	// Test facility bonus
	bonus := sm.GetSkillTrainingBonus("player1", "smithing")
	if bonus != 1.5 {
		t.Errorf("GetSkillTrainingBonus(smithing) = %v, want 1.5", bonus)
	}

	// Test station bonus
	bonus = sm.GetSkillTrainingBonus("player1", "alchemy")
	if bonus != 1.5 {
		t.Errorf("GetSkillTrainingBonus(alchemy) = %v, want 1.5", bonus)
	}

	// Test no bonus
	bonus = sm.GetSkillTrainingBonus("player1", "cooking")
	if bonus != 1.0 {
		t.Errorf("GetSkillTrainingBonus(cooking) = %v, want 1.0", bonus)
	}

	// Test non-existent player
	bonus = sm.GetSkillTrainingBonus("player2", "smithing")
	if bonus != 1.0 {
		t.Errorf("GetSkillTrainingBonus() for non-existent player = %v, want 1.0", bonus)
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	sm := NewStationManager()

	var wg sync.WaitGroup
	concurrency := 10

	// Concurrent registrations
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			station := &CraftingStation{
				ID:      fmt.Sprintf("station%d", id),
				Type:    StationTypeForge,
				Quality: QualityMaster,
				OwnerID: fmt.Sprintf("player%d", id%3),
				HouseID: fmt.Sprintf("house%d", id%3),
			}
			sm.RegisterStation(station)
		}(i)
	}

	wg.Wait()

	// Verify all stations registered
	count := 0
	for i := 0; i < 3; i++ {
		count += len(sm.GetStationsByOwner(fmt.Sprintf("player%d", i)))
	}
	if count != concurrency {
		t.Errorf("Concurrent registrations: got %v stations, want %v", count, concurrency)
	}
}

// Benchmark tests

func BenchmarkGetCraftingBonus(b *testing.B) {
	sm := NewStationManager()

	// Register 10 stations
	for i := 0; i < 10; i++ {
		station := &CraftingStation{
			ID:            fmt.Sprintf("station%d", i),
			Type:          StationTypeForge,
			Quality:       QualityMaster,
			OwnerID:       "player1",
			HouseID:       "house1",
			ActiveRecipes: []string{"recipe1", "recipe2", "recipe3"},
		}
		sm.RegisterStation(station)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.GetCraftingBonus("player1", "recipe2")
	}
}

func BenchmarkUnlockRecipes(b *testing.B) {
	sm := NewStationManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.UnlockRecipes(StationTypeForge, QualityMaster)
	}
}

func BenchmarkGetSkillTrainingBonus(b *testing.B) {
	sm := NewStationManager()

	facility := &SkillTrainingFacility{
		ID:              "facility1",
		OwnerID:         "player1",
		TrainableSkills: []string{"smithing", "alchemy"},
		XPMultiplier:    1.5,
	}
	sm.RegisterFacility(facility)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.GetSkillTrainingBonus("player1", "smithing")
	}
}
