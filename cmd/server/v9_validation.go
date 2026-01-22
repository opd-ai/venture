//go:build !android && !ios
// +build !android,!ios

// INTEGRATION FIX [Category A]: V9.0 Server-Authoritative Validation Service
// Gap: V9 managers initialized but not actively used for validation
// Fix: Created V9ValidationService to provide server-side validation functions
// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
// Impact: Enables server-authoritative validation of crafting bonuses, companion loyalty, and guild permissions

package main

import (
	"sync"

	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/sirupsen/logrus"
)

// V9ValidationService provides server-authoritative validation for V9.0 integration systems.
// It wraps the three V9 managers and provides validation functions that prevent client-side exploits.
//
// Thread-safe for concurrent access from multiple goroutines.
//
// Usage by systems:
//   - CraftingSystem: Call ValidateCraftingBonus() before applying crafting bonuses
//   - CompanionLoyaltySystem: Call ValidateLoyaltyBonus() before applying loyalty gains
//   - GuildSystem: Call ValidateGuildPermission() before allowing guild resource access
type V9ValidationService struct {
	mu                  sync.RWMutex
	stationManager      *housingcrafting.StationManager
	petHomeManager      *companionhousing.PetHomeManager
	guildHousingManager *guildhousing.Manager
	logger              *logrus.Entry
}

// NewV9ValidationService creates a new V9 validation service with the given managers.
func NewV9ValidationService(
	stationManager *housingcrafting.StationManager,
	petHomeManager *companionhousing.PetHomeManager,
	guildHousingManager *guildhousing.Manager,
	logger *logrus.Logger,
) *V9ValidationService {
	var entry *logrus.Entry
	if logger != nil {
		entry = logger.WithField("component", "v9_validation")
	}
	return &V9ValidationService{
		stationManager:      stationManager,
		petHomeManager:      petHomeManager,
		guildHousingManager: guildHousingManager,
		logger:              entry,
	}
}

// ValidateCraftingBonus validates a client's claimed crafting bonus against server state.
// Returns the server-authoritative bonus multiplier. If the client claims a higher bonus
// than they're entitled to, the server value is returned instead.
//
// Parameters:
//   - playerID: The player attempting to craft
//   - recipeID: The recipe being crafted
//   - claimedBonus: The bonus multiplier claimed by the client
//
// Returns:
//   - validatedBonus: The server-authoritative bonus (may be lower than claimed)
//   - isValid: True if the claimed bonus matches server state
func (v *V9ValidationService) ValidateCraftingBonus(playerID, recipeID string, claimedBonus float64) (validatedBonus float64, isValid bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.stationManager == nil {
		// No station manager configured, allow base bonus only
		return 1.0, claimedBonus <= 1.0
	}

	serverBonus := v.stationManager.GetCraftingBonus(playerID, recipeID)

	if claimedBonus > serverBonus {
		if v.logger != nil {
			v.logger.WithFields(logrus.Fields{
				"playerID":    playerID,
				"recipeID":    recipeID,
				"claimedBonus": claimedBonus,
				"serverBonus": serverBonus,
				"action":      "crafting_bonus_rejected",
			}).Warn("Client claimed crafting bonus exceeds server-validated bonus")
		}
		return serverBonus, false
	}

	return claimedBonus, true
}

// ValidateSkillTrainingBonus validates a client's claimed skill training XP bonus.
// Returns the server-authoritative bonus multiplier.
//
// Parameters:
//   - playerID: The player training a skill
//   - skillName: The skill being trained
//   - claimedBonus: The XP multiplier claimed by the client
//
// Returns:
//   - validatedBonus: The server-authoritative bonus
//   - isValid: True if the claimed bonus matches server state
func (v *V9ValidationService) ValidateSkillTrainingBonus(playerID, skillName string, claimedBonus float64) (validatedBonus float64, isValid bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.stationManager == nil {
		return 1.0, claimedBonus <= 1.0
	}

	serverBonus := v.stationManager.GetSkillTrainingBonus(playerID, skillName)

	if claimedBonus > serverBonus {
		if v.logger != nil {
			v.logger.WithFields(logrus.Fields{
				"playerID":    playerID,
				"skillName":   skillName,
				"claimedBonus": claimedBonus,
				"serverBonus": serverBonus,
				"action":      "skill_training_bonus_rejected",
			}).Warn("Client claimed skill training bonus exceeds server-validated bonus")
		}
		return serverBonus, false
	}

	return claimedBonus, true
}

// ValidateLoyaltyBonus validates a client's claimed companion loyalty bonus.
// Returns the server-authoritative bonus for a companion in a specific house.
//
// Parameters:
//   - companionID: The companion's entity ID
//   - houseID: The house where the companion resides
//   - claimedBonus: The loyalty bonus claimed by the client
//
// Returns:
//   - validatedBonus: The server-authoritative bonus
//   - isValid: True if the claimed bonus matches server state
func (v *V9ValidationService) ValidateLoyaltyBonus(companionID uint64, houseID string, claimedBonus float64) (validatedBonus float64, isValid bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.petHomeManager == nil {
		return 0.0, claimedBonus <= 0.0
	}

	serverBonus := v.petHomeManager.GetLoyaltyBonus(companionID, houseID)

	if claimedBonus > serverBonus {
		if v.logger != nil {
			v.logger.WithFields(logrus.Fields{
				"companionID": companionID,
				"houseID":     houseID,
				"claimedBonus": claimedBonus,
				"serverBonus": serverBonus,
				"action":      "loyalty_bonus_rejected",
			}).Warn("Client claimed loyalty bonus exceeds server-validated bonus")
		}
		return serverBonus, false
	}

	return claimedBonus, true
}

// ValidateTrainingBonus validates a client's claimed companion training XP bonus.
//
// Parameters:
//   - companionID: The companion's entity ID
//   - houseID: The house where training is occurring
//   - claimedBonus: The XP multiplier claimed by the client
//
// Returns:
//   - validatedBonus: The server-authoritative bonus
//   - isValid: True if the claimed bonus matches server state
func (v *V9ValidationService) ValidateTrainingBonus(companionID uint64, houseID string, claimedBonus float64) (validatedBonus float64, isValid bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.petHomeManager == nil {
		return 1.0, claimedBonus <= 1.0
	}

	serverBonus := v.petHomeManager.GetTrainingBonus(companionID, houseID)

	if claimedBonus > serverBonus {
		if v.logger != nil {
			v.logger.WithFields(logrus.Fields{
				"companionID": companionID,
				"houseID":     houseID,
				"claimedBonus": claimedBonus,
				"serverBonus": serverBonus,
				"action":      "training_bonus_rejected",
			}).Warn("Client claimed training bonus exceeds server-validated bonus")
		}
		return serverBonus, false
	}

	return claimedBonus, true
}

// ValidateGuildPermission checks if a player with a given rank has permission to access guild resources.
// Server-authoritative check that prevents permission exploits.
//
// Parameters:
//   - houseID: The guild house being accessed
//   - rank: The player's guild rank
//   - requiredPermission: The permission level required for the action
//
// Returns:
//   - hasPermission: True if the player has the required permission
func (v *V9ValidationService) ValidateGuildPermission(houseID string, rank guild.Rank, requiredPermission guildhousing.Permission) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.guildHousingManager == nil {
		return false
	}

	hasPermission := v.guildHousingManager.CheckPermission(houseID, rank, requiredPermission)

	if !hasPermission && v.logger != nil {
		v.logger.WithFields(logrus.Fields{
			"houseID":            houseID,
			"rank":               rank,
			"requiredPermission": requiredPermission,
			"action":             "guild_permission_denied",
		}).Debug("Guild permission check failed")
	}

	return hasPermission
}

// GetGuildUpgradeBonus returns the server-authoritative upgrade bonus for a guild house.
//
// Parameters:
//   - houseID: The guild house ID
//
// Returns:
//   - bonus: The upgrade bonus multiplier (1.0 if no upgrades)
//   - exists: True if the guild house exists
func (v *V9ValidationService) GetGuildUpgradeBonus(houseID string) (bonus float64, exists bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.guildHousingManager == nil {
		return 1.0, false
	}

	serverBonus, err := v.guildHousingManager.GetUpgradeBonus(houseID)
	if err != nil {
		return 1.0, false
	}

	return serverBonus, true
}

// GetCompanionHome returns the house ID where a companion is assigned (server-authoritative).
//
// Parameters:
//   - companionID: The companion's entity ID
//
// Returns:
//   - houseID: The house where the companion lives (empty if no home assigned)
func (v *V9ValidationService) GetCompanionHome(companionID uint64) string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.petHomeManager == nil {
		return ""
	}

	return v.petHomeManager.GetCompanionHome(companionID)
}

// GetStationManager returns the underlying station manager for direct access.
// Prefer using validation methods over direct manager access.
func (v *V9ValidationService) GetStationManager() *housingcrafting.StationManager {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.stationManager
}

// GetPetHomeManager returns the underlying pet home manager for direct access.
// Prefer using validation methods over direct manager access.
func (v *V9ValidationService) GetPetHomeManager() *companionhousing.PetHomeManager {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.petHomeManager
}

// GetGuildHousingManager returns the underlying guild housing manager for direct access.
// Prefer using validation methods over direct manager access.
func (v *V9ValidationService) GetGuildHousingManager() *guildhousing.Manager {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.guildHousingManager
}

// Global V9 validation service instance (set during server initialization)
var v9ValidationService *V9ValidationService

// GetV9ValidationService returns the global V9 validation service.
// Returns nil if the server has not been initialized.
func GetV9ValidationService() *V9ValidationService {
	return v9ValidationService
}
