package guild_housing

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/housing"
)

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
func (m *Manager) CreateGuildHouse(guildID, ownerID string, size housing.BuildingSize) *GuildHouse {
	m.mu.Lock()
	defer m.mu.Unlock()

	houseID := fmt.Sprintf("ghouse-%s-%d", guildID, time.Now().UnixNano())
	house := &GuildHouse{
		HouseID:     houseID,
		GuildID:     guildID,
		OwnerID:     ownerID,
		Size:        size,
		Permissions: DefaultPermissions(),
		Tier:        TierBasic,
		Stations:    []string{},
		Storage:     nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.houses[houseID] = house
	return house
}

// GetGuildHouse retrieves a guild house by ID.
func (m *Manager) GetGuildHouse(houseID string) (*GuildHouse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		return nil, fmt.Errorf("guild house not found: %s", houseID)
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
func (m *Manager) SetPermission(houseID string, rank guild.Rank, permission Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house, exists := m.houses[houseID]
	if !exists {
		return fmt.Errorf("guild house not found: %s", houseID)
	}

	house.Permissions[rank] = permission
	house.UpdatedAt = time.Now()
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
	m.mu.Lock()
	defer m.mu.Unlock()

	house, exists := m.houses[houseID]
	if !exists {
		return fmt.Errorf("guild house not found: %s", houseID)
	}

	house.Stations = append(house.Stations, stationID)
	house.UpdatedAt = time.Now()
	return nil
}

// GetCraftingStations returns all crafting stations in a guild house.
func (m *Manager) GetCraftingStations(houseID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		return nil, fmt.Errorf("guild house not found: %s", houseID)
	}

	return house.Stations, nil
}

// CreateGuildStorage creates storage for a guild house.
func (m *Manager) CreateGuildStorage(guildID string, capacity int) *GuildStorage {
	m.mu.Lock()
	defer m.mu.Unlock()

	storageID := fmt.Sprintf("gstorage-%s-%d", guildID, time.Now().UnixNano())
	storage := &GuildStorage{
		StorageID:    storageID,
		GuildID:      guildID,
		Capacity:     capacity,
		Items:        make(map[string]*StoredItem),
		Transactions: []*Transaction{},
		CreatedAt:    time.Now(),
	}

	m.storage[storageID] = storage
	return storage
}

// GetGuildStorage retrieves guild storage by ID.
func (m *Manager) GetGuildStorage(storageID string) (*GuildStorage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storage, exists := m.storage[storageID]
	if !exists {
		return nil, fmt.Errorf("guild storage not found: %s", storageID)
	}
	return storage, nil
}

// DepositItem deposits an item into guild storage.
// Capacity limits the number of unique item types (slots), not total quantity.
func (m *Manager) DepositItem(storageID, playerID, itemID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	storage, exists := m.storage[storageID]
	if !exists {
		return fmt.Errorf("guild storage not found: %s", storageID)
	}

	if existing, exists := storage.Items[itemID]; exists {
		// Adding to existing item doesn't consume a new slot
		existing.Quantity += quantity
	} else {
		// Adding a new item type requires an available slot
		if len(storage.Items) >= storage.Capacity {
			return fmt.Errorf("storage capacity reached: %d unique item types", storage.Capacity)
		}
		storage.Items[itemID] = &StoredItem{
			ItemID:   itemID,
			Quantity: quantity,
			AddedBy:  playerID,
			AddedAt:  time.Now(),
		}
	}

	transaction := &Transaction{
		TransactionID: fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		PlayerID:      playerID,
		ItemID:        itemID,
		Quantity:      quantity,
		Action:        TransactionDeposit,
		Timestamp:     time.Now(),
	}
	storage.Transactions = append(storage.Transactions, transaction)

	return nil
}

// WithdrawItem withdraws an item from guild storage.
func (m *Manager) WithdrawItem(storageID, playerID, itemID string, quantity int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	storage, exists := m.storage[storageID]
	if !exists {
		return 0, fmt.Errorf("guild storage not found: %s", storageID)
	}

	item, exists := storage.Items[itemID]
	if !exists {
		return 0, fmt.Errorf("item not found in storage: %s", itemID)
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
		TransactionID: fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		PlayerID:      playerID,
		ItemID:        itemID,
		Quantity:      withdrawn,
		Action:        TransactionWithdraw,
		Timestamp:     time.Now(),
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
		return nil, fmt.Errorf("guild storage not found: %s", storageID)
	}

	return storage.Items, nil
}

// GetTransactions returns transaction history for guild storage.
func (m *Manager) GetTransactions(storageID string) ([]*Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storage, exists := m.storage[storageID]
	if !exists {
		return nil, fmt.Errorf("guild storage not found: %s", storageID)
	}

	return storage.Transactions, nil
}

// UpgradeHouse upgrades a guild house to the next tier.
func (m *Manager) UpgradeHouse(houseID string, goldSpent int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house, exists := m.houses[houseID]
	if !exists {
		return fmt.Errorf("guild house not found: %s", houseID)
	}

	if house.Tier >= TierMaster {
		return fmt.Errorf("house already at maximum tier")
	}

	nextTier := house.Tier + 1
	requiredGold := nextTier.Cost()

	if goldSpent < requiredGold {
		return fmt.Errorf("insufficient gold: need %d, have %d", requiredGold, goldSpent)
	}

	house.Tier = nextTier
	house.UpdatedAt = time.Now()
	return nil
}

// GetUpgradeBonus returns the current upgrade bonus multiplier for a house.
func (m *Manager) GetUpgradeBonus(houseID string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	house, exists := m.houses[houseID]
	if !exists {
		return 1.0, fmt.Errorf("guild house not found: %s", houseID)
	}

	return house.Tier.BonusMultiplier(), nil
}

// CreateMeetingHall creates a meeting hall for a guild house.
func (m *Manager) CreateMeetingHall(guildID string, maxCapacity int) *MeetingHall {
	hall := &MeetingHall{
		HallID:      fmt.Sprintf("hall-%s-%d", guildID, time.Now().UnixNano()),
		GuildID:     guildID,
		ChatRadius:  150.0, // +50% from base 100.0
		MaxCapacity: maxCapacity,
		Members:     []string{},
		CreatedAt:   time.Now(),
	}
	return hall
}

// AddMemberToHall adds a member to the meeting hall.
func (m *Manager) AddMemberToHall(hall *MeetingHall, playerID string) error {
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
	for i, member := range hall.Members {
		if member == playerID {
			hall.Members = append(hall.Members[:i], hall.Members[i+1:]...)
			return
		}
	}
}

// Save serializes the manager state to JSON.
func (m *Manager) Save() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := map[string]interface{}{
		"houses":  m.houses,
		"storage": m.storage,
	}

	return json.Marshal(data)
}

// Load deserializes manager state from JSON.
func (m *Manager) Load(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	if houses, ok := state["houses"].(map[string]interface{}); ok {
		housesData, err := json.Marshal(houses)
		if err != nil {
			return fmt.Errorf("failed to marshal houses: %w", err)
		}
		if err := json.Unmarshal(housesData, &m.houses); err != nil {
			return fmt.Errorf("failed to unmarshal houses: %w", err)
		}
	}

	if storage, ok := state["storage"].(map[string]interface{}); ok {
		storageData, err := json.Marshal(storage)
		if err != nil {
			return fmt.Errorf("failed to marshal storage: %w", err)
		}
		if err := json.Unmarshal(storageData, &m.storage); err != nil {
			return fmt.Errorf("failed to unmarshal storage: %w", err)
		}
	}

	return nil
}
