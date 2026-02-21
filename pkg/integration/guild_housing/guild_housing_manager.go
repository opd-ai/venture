package guild_housing

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/housing"
	"github.com/sirupsen/logrus"
)

// logger is the package-level logger for structured logging.
var logger = logrus.StandardLogger()

// validateID validates a string ID parameter.
// Returns an error if the ID is empty or exceeds 256 characters.
func validateID(fieldName, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	if len(value) > 256 {
		return fmt.Errorf("%s exceeds maximum length of 256 characters", fieldName)
	}
	return nil
}

// guild_housing_manager.go implements the Manager for guild housing operations.
// This includes creating guild houses, managing permissions, storage, upgrades,
// and meeting halls with thread-safe concurrent access.

// Manager manages guild housing operations.
type Manager struct {
	houses  map[string]*GuildHouse
	storage map[string]*GuildStorage
	mu      sync.RWMutex
}

// NewManager creates a new guild housing manager.
func NewManager() *Manager {
	return &Manager{
		houses:  make(map[string]*GuildHouse),
		storage: make(map[string]*GuildStorage),
	}
}

// CreateGuildHouse creates a new guild-owned house.
// Returns an error if guildID, ownerID are empty, or if size is invalid.
func (m *Manager) CreateGuildHouse(guildID, ownerID string, size housing.BuildingSize) (*GuildHouse, error) {
	if err := validateID("guildID", guildID); err != nil {
		logger.WithFields(logrus.Fields{
			"guildID": guildID,
			"ownerID": ownerID,
		}).Error("CreateGuildHouse validation failed: ", err)
		return nil, err
	}
	if err := validateID("ownerID", ownerID); err != nil {
		logger.WithFields(logrus.Fields{
			"guildID": guildID,
			"ownerID": ownerID,
		}).Error("CreateGuildHouse validation failed: ", err)
		return nil, err
	}
	if size <= 0 {
		err := fmt.Errorf("building size must be positive, got %d", size)
		logger.WithFields(logrus.Fields{
			"guildID": guildID,
			"ownerID": ownerID,
			"size":    size,
		}).Error("CreateGuildHouse validation failed: ", err)
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	houseID := fmt.Sprintf("ghouse-%s-%d", guildID, now().UnixNano())
	house := &GuildHouse{
		HouseID:     houseID,
		GuildID:     guildID,
		OwnerID:     ownerID,
		Size:        size,
		Permissions: DefaultPermissions(),
		Tier:        TierBasic,
		Stations:    []string{},
		Storage:     nil,
		CreatedAt:   now(),
		UpdatedAt:   now(),
	}

	m.houses[houseID] = house
	return house, nil
}

// GetGuildHouse retrieves a guild house by ID.
func (m *Manager) GetGuildHouse(houseID string) (*GuildHouse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		err := fmt.Errorf("guild house not found: %s", houseID)
		logger.WithFields(logrus.Fields{
			"houseID": houseID,
		}).Warn(err.Error())
		return nil, err
	}
	return house, nil
}

// GetGuildHouses returns all houses for a guild.
func (m *Manager) GetGuildHouses(guildID string) []*GuildHouse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var houses []*GuildHouse
	for _, house := range m.houses {
		if house.GuildID == guildID {
			houses = append(houses, house)
		}
	}
	return houses
}

// SetPermission sets rank-based permission for a guild house.
// The permission value must be in the valid range (PermissionNone through PermissionAdmin, 0-4).
func (m *Manager) SetPermission(houseID string, rank guild.Rank, permission Permission) error {
	if !permission.Valid() {
		err := fmt.Errorf("invalid permission value: %d (must be 0-%d)", permission, PermissionAdmin)
		logger.WithFields(logrus.Fields{
			"houseID":    houseID,
			"rank":       rank,
			"permission": permission,
		}).Error("SetPermission validation failed: ", err)
		return err
	}
	if err := validateID("houseID", houseID); err != nil {
		logger.WithFields(logrus.Fields{
			"houseID":    houseID,
			"rank":       rank,
			"permission": permission,
		}).Error("SetPermission validation failed: ", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	house, exists := m.houses[houseID]
	if !exists {
		err := fmt.Errorf("guild house not found: %s", houseID)
		logger.WithFields(logrus.Fields{
			"houseID": houseID,
		}).Warn(err.Error())
		return err
	}

	house.Permissions[rank] = permission
	house.UpdatedAt = now()
	return nil
}

// CheckPermission verifies if a player with given rank has the required permission.
func (m *Manager) CheckPermission(houseID string, rank guild.Rank, required Permission) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		return false
	}

	playerPerm, exists := house.Permissions[rank]
	if !exists {
		return false
	}

	return playerPerm >= required
}

// AddCraftingStation adds a crafting station to a guild house.
func (m *Manager) AddCraftingStation(houseID, stationID string) error {
	if err := validateID("houseID", houseID); err != nil {
		logger.WithFields(logrus.Fields{
			"houseID":   houseID,
			"stationID": stationID,
		}).Error("AddCraftingStation validation failed: ", err)
		return err
	}
	if err := validateID("stationID", stationID); err != nil {
		logger.WithFields(logrus.Fields{
			"houseID":   houseID,
			"stationID": stationID,
		}).Error("AddCraftingStation validation failed: ", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	house, exists := m.houses[houseID]
	if !exists {
		err := fmt.Errorf("guild house not found: %s", houseID)
		logger.WithFields(logrus.Fields{
			"houseID":   houseID,
			"stationID": stationID,
		}).Warn(err.Error())
		return err
	}

	house.Stations = append(house.Stations, stationID)
	house.UpdatedAt = now()
	return nil
}

// GetCraftingStations returns all crafting stations in a guild house.
func (m *Manager) GetCraftingStations(houseID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		err := fmt.Errorf("guild house not found: %s", houseID)
		logger.WithFields(logrus.Fields{
			"houseID": houseID,
		}).Warn(err.Error())
		return nil, err
	}

	return house.Stations, nil
}

// CreateGuildStorage creates storage for a guild house.
// Returns an error if guildID is empty or capacity is not positive.
func (m *Manager) CreateGuildStorage(guildID string, capacity int) (*GuildStorage, error) {
	if err := validateID("guildID", guildID); err != nil {
		logger.WithFields(logrus.Fields{
			"guildID":  guildID,
			"capacity": capacity,
		}).Error("CreateGuildStorage validation failed: ", err)
		return nil, err
	}
	if capacity <= 0 {
		err := fmt.Errorf("capacity must be positive, got %d", capacity)
		logger.WithFields(logrus.Fields{
			"guildID":  guildID,
			"capacity": capacity,
		}).Error("CreateGuildStorage validation failed: ", err)
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	storageID := fmt.Sprintf("gstorage-%s-%d", guildID, now().UnixNano())
	storage := &GuildStorage{
		StorageID:    storageID,
		GuildID:      guildID,
		Capacity:     capacity,
		Items:        make(map[string]*StoredItem),
		Transactions: []*Transaction{},
		CreatedAt:    now(),
	}

	m.storage[storageID] = storage
	return storage, nil
}

// GetGuildStorage retrieves guild storage by ID.
func (m *Manager) GetGuildStorage(storageID string) (*GuildStorage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storage, exists := m.storage[storageID]
	if !exists {
		err := fmt.Errorf("guild storage not found: %s", storageID)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
		}).Warn(err.Error())
		return nil, err
	}
	return storage, nil
}

// DepositItem deposits an item into guild storage.
// Capacity limits the number of unique item types (slots), not total quantity.
func (m *Manager) DepositItem(storageID, playerID, itemID string, quantity int) error {
	if err := validateID("storageID", storageID); err != nil {
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("DepositItem validation failed: ", err)
		return err
	}
	if err := validateID("playerID", playerID); err != nil {
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("DepositItem validation failed: ", err)
		return err
	}
	if err := validateID("itemID", itemID); err != nil {
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("DepositItem validation failed: ", err)
		return err
	}
	if quantity <= 0 {
		err := fmt.Errorf("quantity must be positive, got %d", quantity)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("DepositItem validation failed: ", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	storage, exists := m.storage[storageID]
	if !exists {
		err := fmt.Errorf("guild storage not found: %s", storageID)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
		}).Warn(err.Error())
		return err
	}

	if existing, exists := storage.Items[itemID]; exists {
		// Adding to existing item doesn't consume a new slot
		existing.Quantity += quantity
	} else {
		// Adding a new item type requires an available slot
		if len(storage.Items) >= storage.Capacity {
			err := fmt.Errorf("storage capacity reached: %d unique item types", storage.Capacity)
			logger.WithFields(logrus.Fields{
				"storageID": storageID,
				"playerID":  playerID,
				"itemID":    itemID,
				"capacity":  storage.Capacity,
			}).Warn(err.Error())
			return err
		}
		storage.Items[itemID] = &StoredItem{
			ItemID:   itemID,
			Quantity: quantity,
			AddedBy:  playerID,
			AddedAt:  now(),
		}
	}

	transaction := &Transaction{
		TransactionID: fmt.Sprintf("tx-%d", now().UnixNano()),
		PlayerID:      playerID,
		ItemID:        itemID,
		Quantity:      quantity,
		Action:        TransactionDeposit,
		Timestamp:     now(),
	}
	storage.Transactions = append(storage.Transactions, transaction)

	return nil
}

// WithdrawItem withdraws an item from guild storage.
func (m *Manager) WithdrawItem(storageID, playerID, itemID string, quantity int) (int, error) {
	if err := validateID("storageID", storageID); err != nil {
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("WithdrawItem validation failed: ", err)
		return 0, err
	}
	if err := validateID("playerID", playerID); err != nil {
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("WithdrawItem validation failed: ", err)
		return 0, err
	}
	if err := validateID("itemID", itemID); err != nil {
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("WithdrawItem validation failed: ", err)
		return 0, err
	}
	if quantity <= 0 {
		err := fmt.Errorf("quantity must be positive, got %d", quantity)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
			"quantity":  quantity,
		}).Error("WithdrawItem validation failed: ", err)
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	storage, exists := m.storage[storageID]
	if !exists {
		err := fmt.Errorf("guild storage not found: %s", storageID)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
		}).Warn(err.Error())
		return 0, err
	}

	item, exists := storage.Items[itemID]
	if !exists {
		err := fmt.Errorf("item not found in storage: %s", itemID)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
			"playerID":  playerID,
			"itemID":    itemID,
		}).Warn(err.Error())
		return 0, err
	}

	withdrawn := quantity
	if item.Quantity < quantity {
		withdrawn = item.Quantity
	}

	item.Quantity -= withdrawn
	if item.Quantity == 0 {
		delete(storage.Items, itemID)
	}

	transaction := &Transaction{
		TransactionID: fmt.Sprintf("tx-%d", now().UnixNano()),
		PlayerID:      playerID,
		ItemID:        itemID,
		Quantity:      withdrawn,
		Action:        TransactionWithdraw,
		Timestamp:     now(),
	}
	storage.Transactions = append(storage.Transactions, transaction)

	return withdrawn, nil
}

// GetStorageItems returns all items in guild storage.
func (m *Manager) GetStorageItems(storageID string) (map[string]*StoredItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storage, exists := m.storage[storageID]
	if !exists {
		err := fmt.Errorf("guild storage not found: %s", storageID)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
		}).Warn(err.Error())
		return nil, err
	}

	return storage.Items, nil
}

// GetTransactions returns transaction history for guild storage.
func (m *Manager) GetTransactions(storageID string) ([]*Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storage, exists := m.storage[storageID]
	if !exists {
		err := fmt.Errorf("guild storage not found: %s", storageID)
		logger.WithFields(logrus.Fields{
			"storageID": storageID,
		}).Warn(err.Error())
		return nil, err
	}

	return storage.Transactions, nil
}

// UpgradeHouse upgrades a guild house to the next tier.
func (m *Manager) UpgradeHouse(houseID string, goldSpent int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house, exists := m.houses[houseID]
	if !exists {
		err := fmt.Errorf("guild house not found: %s", houseID)
		logger.WithFields(logrus.Fields{
			"houseID":   houseID,
			"goldSpent": goldSpent,
		}).Warn(err.Error())
		return err
	}

	if house.Tier >= TierMaster {
		err := fmt.Errorf("house already at maximum tier")
		logger.WithFields(logrus.Fields{
			"houseID": houseID,
			"tier":    house.Tier,
		}).Warn(err.Error())
		return err
	}

	nextTier := house.Tier + 1
	requiredGold := nextTier.Cost()

	if goldSpent < requiredGold {
		err := fmt.Errorf("insufficient gold: need %d, have %d", requiredGold, goldSpent)
		logger.WithFields(logrus.Fields{
			"houseID":      houseID,
			"requiredGold": requiredGold,
			"goldSpent":    goldSpent,
		}).Warn(err.Error())
		return err
	}

	house.Tier = nextTier
	house.UpdatedAt = now()
	return nil
}

// GetUpgradeBonus returns the current upgrade bonus multiplier for a house.
func (m *Manager) GetUpgradeBonus(houseID string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		err := fmt.Errorf("guild house not found: %s", houseID)
		logger.WithFields(logrus.Fields{
			"houseID": houseID,
		}).Warn(err.Error())
		return 1.0, err
	}

	return house.Tier.BonusMultiplier(), nil
}

// CreateMeetingHall creates a meeting hall for a guild house.
// Returns an error if guildID is empty or maxCapacity is not positive.
func (m *Manager) CreateMeetingHall(guildID string, maxCapacity int) (*MeetingHall, error) {
	if err := validateID("guildID", guildID); err != nil {
		logger.WithFields(logrus.Fields{
			"guildID":     guildID,
			"maxCapacity": maxCapacity,
		}).Error("CreateMeetingHall validation failed: ", err)
		return nil, err
	}
	if maxCapacity <= 0 {
		err := fmt.Errorf("maxCapacity must be positive, got %d", maxCapacity)
		logger.WithFields(logrus.Fields{
			"guildID":     guildID,
			"maxCapacity": maxCapacity,
		}).Error("CreateMeetingHall validation failed: ", err)
		return nil, err
	}

	hall := &MeetingHall{
		HallID:      fmt.Sprintf("hall-%s-%d", guildID, now().UnixNano()),
		GuildID:     guildID,
		ChatRadius:  150.0, // +50% from base 100.0
		MaxCapacity: maxCapacity,
		Members:     []string{},
		CreatedAt:   now(),
	}
	return hall, nil
}

// AddMemberToHall adds a member to the meeting hall.
func (m *Manager) AddMemberToHall(hall *MeetingHall, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(hall.Members) >= hall.MaxCapacity {
		return fmt.Errorf("meeting hall at capacity: %d", hall.MaxCapacity)
	}

	for _, member := range hall.Members {
		if member == playerID {
			return fmt.Errorf("player already in hall: %s", playerID)
		}
	}

	hall.Members = append(hall.Members, playerID)
	return nil
}

// RemoveMemberFromHall removes a member from the meeting hall.
func (m *Manager) RemoveMemberFromHall(hall *MeetingHall, playerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, member := range hall.Members {
		if member == playerID {
			hall.Members = append(hall.Members[:i], hall.Members[i+1:]...)
			return
		}
	}
}

// managerState is a typed helper struct for direct JSON serialization/deserialization
// of Manager state. Using a typed struct instead of map[string]interface{} provides
// type safety and avoids the double marshal/unmarshal overhead. Unknown fields from
// newer versions are silently ignored by json.Unmarshal, preserving forward compatibility.
type managerState struct {
	Houses  map[string]*GuildHouse   `json:"houses"`
	Storage map[string]*GuildStorage `json:"storage"`
}

// Save serializes the manager state to JSON.
func (m *Manager) Save() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := managerState{
		Houses:  m.houses,
		Storage: m.storage,
	}

	return json.Marshal(state)
}

// Load deserializes manager state from JSON.
// Unknown fields from newer save formats are silently ignored, providing
// forward compatibility without requiring migration logic.
func (m *Manager) Load(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var state managerState
	if err := json.Unmarshal(data, &state); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Load failed to unmarshal state")
		return err
	}

	if state.Houses != nil {
		m.houses = state.Houses
	}
	if state.Storage != nil {
		m.storage = state.Storage
	}

	return nil
}
