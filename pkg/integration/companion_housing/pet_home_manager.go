package companion_housing

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PetHomeManager manages companion interactions with housing furniture.
// Thread-safe for concurrent access.
type PetHomeManager struct {
	mu             sync.RWMutex
	bedding        map[string]*CompanionBedding // furnitureID → bedding
	trainingAreas  map[string]*TrainingArea     // furnitureID → training area
	storageChests  map[string]*StorageChest     // furnitureID → chest
	companionHomes map[uint64]string            // companionID → houseID
	houseBedding   map[string][]string          // houseID → furnitureIDs of bedding
	houseTraining  map[string][]string          // houseID → furnitureIDs of training
	houseStorage   map[string][]string          // houseID → furnitureIDs of storage
	logger         *logrus.Entry                // optional logger for structured logging
}

// NewPetHomeManager creates a new pet home manager.
func NewPetHomeManager() *PetHomeManager {
	return &PetHomeManager{
		bedding:        make(map[string]*CompanionBedding),
		trainingAreas:  make(map[string]*TrainingArea),
		storageChests:  make(map[string]*StorageChest),
		companionHomes: make(map[uint64]string),
		houseBedding:   make(map[string][]string),
		houseTraining:  make(map[string][]string),
		houseStorage:   make(map[string][]string),
	}
}

// NewPetHomeManagerWithLogger creates a new pet home manager with injectable logger.
// The logger is used for structured logging of operations and warnings.
func NewPetHomeManagerWithLogger(logger *logrus.Entry) *PetHomeManager {
	m := NewPetHomeManager()
	m.logger = logger
	return m
}

// AddBedding registers companion bedding furniture in a house.
func (m *PetHomeManager) AddBedding(houseID, furnitureID string, quality BeddingQuality) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bedding := &CompanionBedding{
		FurnitureID:  furnitureID,
		HouseID:      houseID,
		Quality:      quality,
		LastRestTime: time.Time{}, // Zero time initially
	}
	m.bedding[furnitureID] = bedding
	m.houseBedding[houseID] = append(m.houseBedding[houseID], furnitureID)
}

// RemoveBedding unregisters companion bedding furniture.
func (m *PetHomeManager) RemoveBedding(furnitureID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if bedding, ok := m.bedding[furnitureID]; ok {
		// Remove companion assignment
		if bedding.CompanionID != 0 {
			delete(m.companionHomes, bedding.CompanionID)
		}
		// Remove from house list
		m.houseBedding[bedding.HouseID] = m.removeFromSlice(m.houseBedding[bedding.HouseID], furnitureID)
		delete(m.bedding, furnitureID)
	}
}

// AssignCompanionToBed assigns a companion to specific bedding.
// Returns error if furniture doesn't exist or bed already occupied.
func (m *PetHomeManager) AssignCompanionToBed(companionID uint64, furnitureID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bedding, ok := m.bedding[furnitureID]
	if !ok {
		m.logWarn("bedding furniture not found", logrus.Fields{
			"companionID": companionID,
			"furnitureID": furnitureID,
		})
		return fmt.Errorf("bedding furniture %s not found", furnitureID)
	}
	if bedding.CompanionID != 0 && bedding.CompanionID != companionID {
		m.logWarn("bedding already occupied", logrus.Fields{
			"companionID":         companionID,
			"furnitureID":         furnitureID,
			"existingCompanionID": bedding.CompanionID,
		})
		return fmt.Errorf("bedding %s already occupied by companion %d", furnitureID, bedding.CompanionID)
	}

	bedding.CompanionID = companionID
	m.companionHomes[companionID] = bedding.HouseID
	return nil
}

// UnassignCompanionBed removes companion assignment from bedding.
func (m *PetHomeManager) UnassignCompanionBed(companionID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, bedding := range m.bedding {
		if bedding.CompanionID == companionID {
			bedding.CompanionID = 0
			delete(m.companionHomes, companionID)
			return
		}
	}
}

// GetLoyaltyBonus calculates daily loyalty bonus for a companion in a house.
// Returns 0.0 if companion has no assigned bedding.
func (m *PetHomeManager) GetLoyaltyBonus(companionID uint64, houseID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if companion has a home in this house
	if assignedHouse, ok := m.companionHomes[companionID]; !ok || assignedHouse != houseID {
		return 0.0
	}

	// Find companion's bedding and calculate bonus
	for _, bedding := range m.bedding {
		if bedding.CompanionID == companionID && bedding.HouseID == houseID {
			return bedding.LoyaltyBonus()
		}
	}

	return 0.0
}

// RecordRest updates the last rest time for a companion.
// The now parameter must be provided for deterministic gameplay.
// Returns error if companion not assigned to bedding.
func (m *PetHomeManager) RecordRest(companionID uint64, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, bedding := range m.bedding {
		if bedding.CompanionID == companionID {
			bedding.LastRestTime = now
			return nil
		}
	}
	m.logWarn("companion has no assigned bedding", logrus.Fields{
		"companionID": companionID,
	})
	return fmt.Errorf("companion %d has no assigned bedding", companionID)
}

// AddTrainingArea registers a training area in a house.
func (m *PetHomeManager) AddTrainingArea(houseID, furnitureID string, areaType TrainingAreaType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	area := &TrainingArea{
		FurnitureID:    furnitureID,
		HouseID:        houseID,
		Type:           areaType,
		ActiveSessions: make(map[uint64]time.Time),
	}
	m.trainingAreas[furnitureID] = area
	m.houseTraining[houseID] = append(m.houseTraining[houseID], furnitureID)
}

// RemoveTrainingArea unregisters a training area.
func (m *PetHomeManager) RemoveTrainingArea(furnitureID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if area, ok := m.trainingAreas[furnitureID]; ok {
		m.houseTraining[area.HouseID] = m.removeFromSlice(m.houseTraining[area.HouseID], furnitureID)
		delete(m.trainingAreas, furnitureID)
	}
}

// StartTrainingSession begins a training session for a companion.
// The now parameter must be provided for deterministic gameplay.
// Returns error if training area doesn't exist.
func (m *PetHomeManager) StartTrainingSession(companionID uint64, furnitureID string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	area, ok := m.trainingAreas[furnitureID]
	if !ok {
		m.logWarn("training area not found", logrus.Fields{
			"companionID": companionID,
			"furnitureID": furnitureID,
		})
		return fmt.Errorf("training area %s not found", furnitureID)
	}

	area.ActiveSessions[companionID] = now
	return nil
}

// EndTrainingSession ends a training session for a companion.
func (m *PetHomeManager) EndTrainingSession(companionID uint64, furnitureID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if area, ok := m.trainingAreas[furnitureID]; ok {
		delete(area.ActiveSessions, companionID)
	}
}

// GetTrainingBonus calculates XP multiplier for a companion training in a house.
// Returns 1.0 (no bonus) if no training area available.
func (m *PetHomeManager) GetTrainingBonus(companionID uint64, houseID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if companion is in active training session
	for furnitureID, area := range m.trainingAreas {
		if area.HouseID == houseID {
			if _, active := area.ActiveSessions[companionID]; active {
				return m.trainingAreas[furnitureID].XPBonus()
			}
		}
	}

	return 1.0 // No bonus
}

// AddStorageChest registers a storage chest in a house.
func (m *PetHomeManager) AddStorageChest(houseID, furnitureID string, capacity int, sharedWithPets bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	chest := &StorageChest{
		FurnitureID:    furnitureID,
		HouseID:        houseID,
		Capacity:       capacity,
		SharedWithPets: sharedWithPets,
		Items:          []string{},
	}
	m.storageChests[furnitureID] = chest
	m.houseStorage[houseID] = append(m.houseStorage[houseID], furnitureID)
}

// RemoveStorageChest unregisters a storage chest.
func (m *PetHomeManager) RemoveStorageChest(furnitureID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if chest, ok := m.storageChests[furnitureID]; ok {
		m.houseStorage[chest.HouseID] = m.removeFromSlice(m.houseStorage[chest.HouseID], furnitureID)
		delete(m.storageChests, furnitureID)
	}
}

// GetStorageChest retrieves a storage chest by furniture ID.
// Returns nil if not found.
func (m *PetHomeManager) GetStorageChest(furnitureID string) *StorageChest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.storageChests[furnitureID]
}

// GetSharedStorageCapacity returns total shared storage slots in a house.
func (m *PetHomeManager) GetSharedStorageCapacity(houseID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	if chestIDs, ok := m.houseStorage[houseID]; ok {
		for _, chestID := range chestIDs {
			if chest, ok := m.storageChests[chestID]; ok && chest.SharedWithPets {
				total += chest.Capacity
			}
		}
	}
	return total
}

// GetCompanionHome returns the house ID where a companion is assigned.
// Returns empty string if no home assigned.
func (m *PetHomeManager) GetCompanionHome(companionID uint64) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.companionHomes[companionID]
}

// GetHouseBeddingCount returns the number of bedding furniture in a house.
func (m *PetHomeManager) GetHouseBeddingCount(houseID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.houseBedding[houseID])
}

// GetHouseTrainingCount returns the number of training areas in a house.
func (m *PetHomeManager) GetHouseTrainingCount(houseID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.houseTraining[houseID])
}

// removeFromSlice removes the first occurrence of item from slice.
// Returns a new slice without the item, or the original slice if not found.
// Caller must assign the return value to update the slice reference.
func (m *PetHomeManager) removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// logWarn logs a warning message using the injectable logger if available,
// otherwise falls back to global logrus.
func (m *PetHomeManager) logWarn(msg string, fields logrus.Fields) {
	if m.logger != nil {
		m.logger.WithFields(fields).Warn(msg)
	} else {
		logrus.WithFields(fields).Warn(msg)
	}
}
