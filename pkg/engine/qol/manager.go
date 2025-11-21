package qol

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// AutoLootManager manages companion auto-loot behavior
type AutoLootManager struct {
	configs map[uint64]*AutoLootConfig
	mu      sync.RWMutex
}

// NewAutoLootManager creates a new auto-loot manager
func NewAutoLootManager() *AutoLootManager {
	return &AutoLootManager{
		configs: make(map[uint64]*AutoLootConfig),
	}
}

// SetConfig sets auto-loot configuration for a companion
func (m *AutoLootManager) SetConfig(config *AutoLootConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[config.CompanionID] = config
}

// GetConfig retrieves auto-loot configuration for a companion
func (m *AutoLootManager) GetConfig(companionID uint64) *AutoLootConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if config, exists := m.configs[companionID]; exists {
		return config
	}
	return DefaultAutoLootConfig(companionID)
}

// SetRadius sets the loot collection radius for a companion
func (m *AutoLootManager) SetRadius(companionID uint64, radius float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if config, exists := m.configs[companionID]; exists {
		config.Radius = clampRadius(radius)
	} else {
		config := DefaultAutoLootConfig(companionID)
		config.Radius = clampRadius(radius)
		m.configs[companionID] = config
	}
}

// ShouldCollect checks if an item should be auto-collected
func (m *AutoLootManager) ShouldCollect(companionID uint64, itemRarity int, itemType string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, exists := m.configs[companionID]
	if !exists || !config.Enabled {
		return false
	}

	if itemRarity < config.MinRarity {
		return false
	}

	for _, ignore := range config.IgnoreTypes {
		if itemType == ignore {
			return false
		}
	}

	if len(config.FilterTypes) > 0 {
		found := false
		for _, filter := range config.FilterTypes {
			if itemType == filter {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// clampRadius clamps the auto-loot radius to valid range (5-10 tiles)
func clampRadius(radius float64) float64 {
	if radius < 5.0 {
		return 5.0
	}
	if radius > 10.0 {
		return 10.0
	}
	return radius
}

// CraftQueueManager manages crafting recipe queue
type CraftQueueManager struct {
	queue map[uint64][]*CraftQueueEntry
	mu    sync.RWMutex
}

// NewCraftQueueManager creates a new craft queue manager
func NewCraftQueueManager() *CraftQueueManager {
	return &CraftQueueManager{
		queue: make(map[uint64][]*CraftQueueEntry),
	}
}

// AddRecipe adds a recipe to the player's craft queue
func (m *CraftQueueManager) AddRecipe(playerID uint64, recipeID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	if _, exists := m.queue[playerID]; !exists {
		m.queue[playerID] = make([]*CraftQueueEntry, 0)
	}

	if len(m.queue[playerID]) >= 50 {
		return fmt.Errorf("craft queue full (max 50 recipes)")
	}

	entry := &CraftQueueEntry{
		RecipeID:       recipeID,
		Quantity:       quantity,
		MaterialsReady: false,
		Position:       len(m.queue[playerID]),
		AddedAt:        time.Now(),
	}

	m.queue[playerID] = append(m.queue[playerID], entry)
	return nil
}

// RemoveRecipe removes a recipe from the queue
func (m *CraftQueueManager) RemoveRecipe(playerID uint64, position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queue[playerID]
	if !exists || position < 0 || position >= len(queue) {
		return fmt.Errorf("invalid queue position")
	}

	m.queue[playerID] = append(queue[:position], queue[position+1:]...)

	for i := range m.queue[playerID] {
		m.queue[playerID][i].Position = i
	}

	return nil
}

// GetQueue retrieves the player's craft queue
func (m *CraftQueueManager) GetQueue(playerID uint64) []*CraftQueueEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queue, exists := m.queue[playerID]
	if !exists {
		return make([]*CraftQueueEntry, 0)
	}

	result := make([]*CraftQueueEntry, len(queue))
	copy(result, queue)
	return result
}

// ClearQueue clears the player's craft queue
func (m *CraftQueueManager) ClearQueue(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.queue, playerID)
}

// GuildInvitationManager manages offline guild invitations
type GuildInvitationManager struct {
	invitations map[string]*GuildInvitation // invitationID -> invitation
	byInvitee   map[string][]string         // inviteeID -> invitationIDs
	mu          sync.RWMutex
}

// NewGuildInvitationManager creates a new guild invitation manager
func NewGuildInvitationManager() *GuildInvitationManager {
	return &GuildInvitationManager{
		invitations: make(map[string]*GuildInvitation),
		byInvitee:   make(map[string][]string),
	}
}

// SendInvitation creates a new guild invitation
func (m *GuildInvitationManager) SendInvitation(inv *GuildInvitation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inv.ExpiresAt.IsZero() {
		inv.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	if inv.SentAt.IsZero() {
		inv.SentAt = time.Now()
	}

	m.invitations[inv.InvitationID] = inv

	if _, exists := m.byInvitee[inv.InviteeID]; !exists {
		m.byInvitee[inv.InviteeID] = make([]string, 0)
	}
	m.byInvitee[inv.InviteeID] = append(m.byInvitee[inv.InviteeID], inv.InvitationID)
}

// GetPendingInvitations retrieves pending invitations for a player
func (m *GuildInvitationManager) GetPendingInvitations(playerID string) []*GuildInvitation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	invIDs, exists := m.byInvitee[playerID]
	if !exists {
		return make([]*GuildInvitation, 0)
	}

	result := make([]*GuildInvitation, 0)
	for _, invID := range invIDs {
		inv, exists := m.invitations[invID]
		if exists && !inv.Accepted && !inv.IsExpired() {
			result = append(result, inv)
		}
	}

	return result
}

// AcceptInvitation accepts a guild invitation
func (m *GuildInvitationManager) AcceptInvitation(invitationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inv, exists := m.invitations[invitationID]
	if !exists {
		return fmt.Errorf("invitation not found")
	}

	if inv.Accepted {
		return fmt.Errorf("invitation already accepted")
	}

	if inv.IsExpired() {
		return fmt.Errorf("invitation expired")
	}

	inv.Accepted = true
	inv.AcceptedAt = time.Now()
	return nil
}

// CleanupExpired removes expired invitations
func (m *GuildInvitationManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for invID, inv := range m.invitations {
		if inv.IsExpired() {
			delete(m.invitations, invID)
			removed++
		}
	}

	for inviteeID, invIDs := range m.byInvitee {
		newIDs := make([]string, 0)
		for _, invID := range invIDs {
			if _, exists := m.invitations[invID]; exists {
				newIDs = append(newIDs, invID)
			}
		}
		if len(newIDs) == 0 {
			delete(m.byInvitee, inviteeID)
		} else {
			m.byInvitee[inviteeID] = newIDs
		}
	}

	return removed
}

// MountWhistleManager manages vehicle summoning
type MountWhistleManager struct {
	summons map[uint64]*MountSummon
	mu      sync.RWMutex
}

// NewMountWhistleManager creates a new mount whistle manager
func NewMountWhistleManager() *MountWhistleManager {
	return &MountWhistleManager{
		summons: make(map[uint64]*MountSummon),
	}
}

// SummonMount summons a vehicle to the player
func (m *MountWhistleManager) SummonMount(summon *MountSummon) {
	m.mu.Lock()
	defer m.mu.Unlock()

	summon.Distance = math.Sqrt(
		math.Pow(summon.TargetPos[0]-summon.CurrentPos[0], 2) +
			math.Pow(summon.TargetPos[1]-summon.CurrentPos[1], 2),
	)
	summon.EstimatedTime = EstimateArrivalTime(summon.Distance)
	summon.RequestTime = time.Now()
	summon.Completed = false

	m.summons[summon.PlayerID] = summon
}

// GetActiveSummon retrieves the active summon for a player
func (m *MountWhistleManager) GetActiveSummon(playerID uint64) *MountSummon {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.summons[playerID]
}

// CompleteSummon marks a summon as completed
func (m *MountWhistleManager) CompleteSummon(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if summon, exists := m.summons[playerID]; exists {
		summon.Completed = true
	}
}

// CancelSummon cancels an active summon
func (m *MountWhistleManager) CancelSummon(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.summons, playerID)
}

// StorageSorter handles inventory and storage sorting
type StorageSorter struct {
	presets map[string]*StorageSortPreset
	mu      sync.RWMutex
}

// NewStorageSorter creates a new storage sorter
func NewStorageSorter() *StorageSorter {
	s := &StorageSorter{
		presets: make(map[string]*StorageSortPreset),
	}

	s.presets["default"] = &StorageSortPreset{
		Name:              "Default",
		PrimaryCriteria:   SortByType,
		SecondaryCriteria: SortByRarity,
		Descending:        false,
		GroupByType:       true,
	}

	s.presets["rarity"] = &StorageSortPreset{
		Name:              "Rarity",
		PrimaryCriteria:   SortByRarity,
		SecondaryCriteria: SortByName,
		Descending:        true,
		GroupByType:       false,
	}

	s.presets["value"] = &StorageSortPreset{
		Name:              "Value",
		PrimaryCriteria:   SortByValue,
		SecondaryCriteria: SortByQuantity,
		Descending:        true,
		GroupByType:       false,
	}

	return s
}

// AddPreset adds a custom sort preset
func (s *StorageSorter) AddPreset(preset *StorageSortPreset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presets[preset.Name] = preset
}

// GetPreset retrieves a sort preset by name
func (s *StorageSorter) GetPreset(name string) *StorageSortPreset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.presets[name]
}

// Item represents an item for sorting purposes
type Item struct {
	ID       string
	Name     string
	Type     string
	Rarity   int
	Value    int
	Quantity int
}

// SortItems sorts a slice of items using the specified criteria
func (s *StorageSorter) SortItems(items []*Item, criteria SortCriteria) {
	sort.SliceStable(items, func(i, j int) bool {
		switch criteria {
		case SortByType:
			return items[i].Type < items[j].Type
		case SortByRarity:
			return items[i].Rarity > items[j].Rarity
		case SortByName:
			return items[i].Name < items[j].Name
		case SortByValue:
			return items[i].Value > items[j].Value
		case SortByQuantity:
			return items[i].Quantity > items[j].Quantity
		default:
			return false
		}
	})
}

// RecipeTracker tracks recipe availability and missing materials
type RecipeTracker struct {
	tracking map[uint64]map[string]*RecipeTrackingInfo // playerID -> recipeID -> tracking info
	mu       sync.RWMutex
}

// NewRecipeTracker creates a new recipe tracker
func NewRecipeTracker() *RecipeTracker {
	return &RecipeTracker{
		tracking: make(map[uint64]map[string]*RecipeTrackingInfo),
	}
}

// TrackRecipe adds a recipe to tracking for a player
func (r *RecipeTracker) TrackRecipe(playerID uint64, info *RecipeTrackingInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tracking[playerID]; !exists {
		r.tracking[playerID] = make(map[string]*RecipeTrackingInfo)
	}

	info.MissingMats = make(map[string]int)
	info.CanCraft = true
	info.MaxCraftable = int(^uint(0) >> 1) // Max int

	for matID, required := range info.RequiredMats {
		available := info.AvailableMats[matID]
		if available < required {
			info.MissingMats[matID] = required - available
			info.CanCraft = false
			info.MaxCraftable = 0
		} else {
			maxFromThisMat := available / required
			if maxFromThisMat < info.MaxCraftable {
				info.MaxCraftable = maxFromThisMat
			}
		}
	}

	if !info.CanCraft {
		info.MaxCraftable = 0
	}

	r.tracking[playerID][info.RecipeID] = info
}

// GetTrackedRecipes retrieves all tracked recipes for a player
func (r *RecipeTracker) GetTrackedRecipes(playerID uint64) []*RecipeTrackingInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	recipes, exists := r.tracking[playerID]
	if !exists {
		return make([]*RecipeTrackingInfo, 0)
	}

	result := make([]*RecipeTrackingInfo, 0, len(recipes))
	for _, info := range recipes {
		result = append(result, info)
	}

	return result
}

// UntrackRecipe removes a recipe from tracking
func (r *RecipeTracker) UntrackRecipe(playerID uint64, recipeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if recipes, exists := r.tracking[playerID]; exists {
		delete(recipes, recipeID)
	}
}

// UpdateMaterialAvailability updates available materials and recalculates craftability
func (r *RecipeTracker) UpdateMaterialAvailability(playerID uint64, recipeID string, availableMats map[string]int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	recipes, exists := r.tracking[playerID]
	if !exists {
		return
	}

	info, exists := recipes[recipeID]
	if !exists {
		return
	}

	info.AvailableMats = availableMats
	info.MissingMats = make(map[string]int)
	info.CanCraft = true
	info.MaxCraftable = int(^uint(0) >> 1)

	for matID, required := range info.RequiredMats {
		available := availableMats[matID]
		if available < required {
			info.MissingMats[matID] = required - available
			info.CanCraft = false
			info.MaxCraftable = 0
		} else {
			maxFromThisMat := available / required
			if maxFromThisMat < info.MaxCraftable {
				info.MaxCraftable = maxFromThisMat
			}
		}
	}

	if !info.CanCraft {
		info.MaxCraftable = 0
	}
}
