package housing_crafting

import (
	"fmt"
	"sync"
)

// StationManager manages all crafting stations and skill training facilities
type StationManager struct {
	mu                sync.RWMutex
	stations          map[string]*CraftingStation         // Station ID → station
	stationsByOwner   map[string][]*CraftingStation       // Owner ID → stations
	stationsByHouse   map[string][]*CraftingStation       // House ID → stations
	facilities        map[string]*SkillTrainingFacility   // Facility ID → facility
	facilitiesByOwner map[string][]*SkillTrainingFacility // Owner ID → facilities
}

// NewStationManager creates a new station manager
func NewStationManager() *StationManager {
	return &StationManager{
		stations:          make(map[string]*CraftingStation),
		stationsByOwner:   make(map[string][]*CraftingStation),
		stationsByHouse:   make(map[string][]*CraftingStation),
		facilities:        make(map[string]*SkillTrainingFacility),
		facilitiesByOwner: make(map[string][]*SkillTrainingFacility),
	}
}

// RegisterStation registers a new crafting station
func (sm *StationManager) RegisterStation(station *CraftingStation) error {
	if station == nil {
		return fmt.Errorf("station cannot be nil")
	}
	if station.ID == "" {
		return fmt.Errorf("station ID cannot be empty")
	}
	if station.OwnerID == "" {
		return fmt.Errorf("station OwnerID cannot be empty")
	}
	if station.HouseID == "" {
		return fmt.Errorf("station HouseID cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check for duplicate ID
	if _, exists := sm.stations[station.ID]; exists {
		return fmt.Errorf("station with ID %s already exists", station.ID)
	}

	// Register station
	sm.stations[station.ID] = station
	sm.stationsByOwner[station.OwnerID] = append(sm.stationsByOwner[station.OwnerID], station)
	sm.stationsByHouse[station.HouseID] = append(sm.stationsByHouse[station.HouseID], station)

	return nil
}

// UnregisterStation removes a crafting station
func (sm *StationManager) UnregisterStation(stationID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	station, exists := sm.stations[stationID]
	if !exists {
		return fmt.Errorf("station with ID %s not found", stationID)
	}

	// Remove from main map
	delete(sm.stations, stationID)

	// Remove from owner map
	ownerStations := sm.stationsByOwner[station.OwnerID]
	sm.stationsByOwner[station.OwnerID] = sm.removeStationFromSliceValue(ownerStations, stationID)

	// Remove from house map
	houseStations := sm.stationsByHouse[station.HouseID]
	sm.stationsByHouse[station.HouseID] = sm.removeStationFromSliceValue(houseStations, stationID)

	return nil
}

// removeStationFromSliceValue removes a station from a slice by ID and returns the new slice
func (sm *StationManager) removeStationFromSliceValue(stations []*CraftingStation, stationID string) []*CraftingStation {
	for i, s := range stations {
		if s.ID == stationID {
			return append(stations[:i], stations[i+1:]...)
		}
	}
	return stations
}

// GetStation retrieves a station by ID
func (sm *StationManager) GetStation(stationID string) (*CraftingStation, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	station, exists := sm.stations[stationID]
	if !exists {
		return nil, fmt.Errorf("station with ID %s not found", stationID)
	}

	return station, nil
}

// GetStationsByOwner retrieves all stations owned by a player
func (sm *StationManager) GetStationsByOwner(ownerID string) []*CraftingStation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.stationsByOwner[ownerID]
}

// GetStationsByHouse retrieves all stations in a house
func (sm *StationManager) GetStationsByHouse(houseID string) []*CraftingStation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.stationsByHouse[houseID]
}

// GetCraftingBonus calculates the crafting bonus for a player crafting a specific recipe
func (sm *StationManager) GetCraftingBonus(playerID, recipeID string) float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Get player's stations
	stations := sm.stationsByOwner[playerID]
	if len(stations) == 0 {
		return 1.0 // No bonus
	}

	// Find the best station that has this recipe
	var bestMultiplier float64 = 1.0
	for _, station := range stations {
		if station.HasRecipe(recipeID) {
			multiplier := station.Quality.Multiplier()
			if multiplier > bestMultiplier {
				bestMultiplier = multiplier
			}
		}
	}

	return bestMultiplier
}

// UnlockRecipes returns all recipes available for a station type at a given quality
func (sm *StationManager) UnlockRecipes(stationType StationType, quality QualityTier) []string {
	// Base recipes available at all quality levels
	baseRecipes := sm.getBaseRecipes(stationType)

	// Advanced recipes unlock at higher quality tiers
	var recipes []string
	recipes = append(recipes, baseRecipes...)

	if quality >= QualityStandard {
		recipes = append(recipes, sm.getStandardRecipes(stationType)...)
	}
	if quality >= QualityAdvanced {
		recipes = append(recipes, sm.getAdvancedRecipes(stationType)...)
	}
	if quality >= QualityMaster {
		recipes = append(recipes, sm.getMasterRecipes(stationType)...)
	}

	return recipes
}

// RegisterFacility registers a skill training facility
func (sm *StationManager) RegisterFacility(facility *SkillTrainingFacility) error {
	if facility == nil {
		return fmt.Errorf("facility cannot be nil")
	}
	if facility.ID == "" {
		return fmt.Errorf("facility ID cannot be empty")
	}
	if facility.OwnerID == "" {
		return fmt.Errorf("facility OwnerID cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check for duplicate ID
	if _, exists := sm.facilities[facility.ID]; exists {
		return fmt.Errorf("facility with ID %s already exists", facility.ID)
	}

	// Register facility
	sm.facilities[facility.ID] = facility
	sm.facilitiesByOwner[facility.OwnerID] = append(sm.facilitiesByOwner[facility.OwnerID], facility)

	return nil
}

// GetSkillTrainingBonus calculates the XP bonus for skill training in owned facilities
func (sm *StationManager) GetSkillTrainingBonus(playerID, skillName string) float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check skill training facilities
	facilities := sm.facilitiesByOwner[playerID]
	for _, facility := range facilities {
		if facility.CanTrainSkill(skillName) {
			return facility.XPMultiplier
		}
	}

	// Check crafting stations for skill bonuses
	stations := sm.stationsByOwner[playerID]
	for _, station := range stations {
		if bonus := station.GetSkillTrainingBonus(skillName); bonus > 0 {
			// Convert percentage bonus to multiplier (e.g., 50% → 1.5)
			return 1.0 + (float64(bonus) / 100.0)
		}
	}

	return 1.0 // No bonus
}
