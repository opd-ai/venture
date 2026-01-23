//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/sirupsen/logrus"
)

func TestNewV9ValidationService(t *testing.T) {
	tests := []struct {
		name       string
		withLogger bool
	}{
		{"with_logger", true},
		{"without_logger", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stationMgr := housingcrafting.NewStationManager()
			petHomeMgr := companionhousing.NewPetHomeManager()
			guildMgr := guildhousing.NewManager()

			var logger *logrus.Logger
			if tt.withLogger {
				logger = logrus.New()
				logger.SetLevel(logrus.DebugLevel)
			}

			service := NewV9ValidationService(stationMgr, petHomeMgr, guildMgr, logger)
			if service == nil {
				t.Fatal("NewV9ValidationService returned nil")
			}

			if service.stationManager != stationMgr {
				t.Error("stationManager not set correctly")
			}
			if service.petHomeManager != petHomeMgr {
				t.Error("petHomeManager not set correctly")
			}
			if service.guildHousingManager != guildMgr {
				t.Error("guildHousingManager not set correctly")
			}
		})
	}
}

func TestV9ValidationService_ValidateCraftingBonus(t *testing.T) {
	tests := []struct {
		name          string
		playerID      string
		recipeID      string
		claimedBonus  float64
		expectedBonus float64
		expectValid   bool
		nilManager    bool
	}{
		{
			name:          "valid_base_bonus",
			playerID:      "player1",
			recipeID:      "recipe1",
			claimedBonus:  1.0,
			expectedBonus: 1.0,
			expectValid:   true,
		},
		{
			name:          "claimed_exceeds_server",
			playerID:      "player1",
			recipeID:      "recipe1",
			claimedBonus:  2.0,
			expectedBonus: 1.0, // No stations registered, so base bonus
			expectValid:   false,
		},
		{
			name:          "nil_manager_base_valid",
			playerID:      "player1",
			recipeID:      "recipe1",
			claimedBonus:  1.0,
			expectedBonus: 1.0,
			expectValid:   true,
			nilManager:    true,
		},
		{
			name:          "nil_manager_exceeds_base",
			playerID:      "player1",
			recipeID:      "recipe1",
			claimedBonus:  1.5,
			expectedBonus: 1.0,
			expectValid:   false,
			nilManager:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stationMgr *housingcrafting.StationManager
			if !tt.nilManager {
				stationMgr = housingcrafting.NewStationManager()
			}

			service := NewV9ValidationService(stationMgr, nil, nil, nil)
			bonus, valid := service.ValidateCraftingBonus(tt.playerID, tt.recipeID, tt.claimedBonus)

			if bonus != tt.expectedBonus {
				t.Errorf("expected bonus %f, got %f", tt.expectedBonus, bonus)
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid=%t, got %t", tt.expectValid, valid)
			}
		})
	}
}

func TestV9ValidationService_ValidateSkillTrainingBonus(t *testing.T) {
	tests := []struct {
		name          string
		playerID      string
		skillName     string
		claimedBonus  float64
		expectedBonus float64
		expectValid   bool
	}{
		{
			name:          "valid_base_bonus",
			playerID:      "player1",
			skillName:     "smithing",
			claimedBonus:  1.0,
			expectedBonus: 1.0,
			expectValid:   true,
		},
		{
			name:          "claimed_exceeds_server",
			playerID:      "player1",
			skillName:     "smithing",
			claimedBonus:  1.5,
			expectedBonus: 1.0,
			expectValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stationMgr := housingcrafting.NewStationManager()
			service := NewV9ValidationService(stationMgr, nil, nil, nil)

			bonus, valid := service.ValidateSkillTrainingBonus(tt.playerID, tt.skillName, tt.claimedBonus)

			if bonus != tt.expectedBonus {
				t.Errorf("expected bonus %f, got %f", tt.expectedBonus, bonus)
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid=%t, got %t", tt.expectValid, valid)
			}
		})
	}
}

func TestV9ValidationService_ValidateLoyaltyBonus(t *testing.T) {
	tests := []struct {
		name          string
		companionID   uint64
		houseID       string
		claimedBonus  float64
		expectedBonus float64
		expectValid   bool
		nilManager    bool
	}{
		{
			name:          "no_bonus_unassigned_companion",
			companionID:   12345,
			houseID:       "house1",
			claimedBonus:  0.0,
			expectedBonus: 0.0,
			expectValid:   true,
		},
		{
			name:          "claimed_exceeds_zero",
			companionID:   12345,
			houseID:       "house1",
			claimedBonus:  0.5,
			expectedBonus: 0.0, // Companion not assigned
			expectValid:   false,
		},
		{
			name:          "nil_manager_zero_valid",
			companionID:   12345,
			houseID:       "house1",
			claimedBonus:  0.0,
			expectedBonus: 0.0,
			expectValid:   true,
			nilManager:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var petHomeMgr *companionhousing.PetHomeManager
			if !tt.nilManager {
				petHomeMgr = companionhousing.NewPetHomeManager()
			}

			service := NewV9ValidationService(nil, petHomeMgr, nil, nil)
			bonus, valid := service.ValidateLoyaltyBonus(tt.companionID, tt.houseID, tt.claimedBonus)

			if bonus != tt.expectedBonus {
				t.Errorf("expected bonus %f, got %f", tt.expectedBonus, bonus)
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid=%t, got %t", tt.expectValid, valid)
			}
		})
	}
}

func TestV9ValidationService_ValidateTrainingBonus(t *testing.T) {
	tests := []struct {
		name          string
		companionID   uint64
		houseID       string
		claimedBonus  float64
		expectedBonus float64
		expectValid   bool
	}{
		{
			name:          "valid_base_bonus",
			companionID:   12345,
			houseID:       "house1",
			claimedBonus:  1.0,
			expectedBonus: 1.0,
			expectValid:   true,
		},
		{
			name:          "claimed_exceeds_server",
			companionID:   12345,
			houseID:       "house1",
			claimedBonus:  1.5,
			expectedBonus: 1.0, // No training area
			expectValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			petHomeMgr := companionhousing.NewPetHomeManager()
			service := NewV9ValidationService(nil, petHomeMgr, nil, nil)

			bonus, valid := service.ValidateTrainingBonus(tt.companionID, tt.houseID, tt.claimedBonus)

			if bonus != tt.expectedBonus {
				t.Errorf("expected bonus %f, got %f", tt.expectedBonus, bonus)
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid=%t, got %t", tt.expectValid, valid)
			}
		})
	}
}

func TestV9ValidationService_ValidateGuildPermission(t *testing.T) {
	guildMgr := guildhousing.NewManager()
	// Create a guild house for testing
	house := guildMgr.CreateGuildHouse("guild1", "owner1", 1)

	tests := []struct {
		name                string
		houseID             string
		rank                guild.Rank
		requiredPermission  guildhousing.Permission
		expectHasPermission bool
		nilManager          bool
	}{
		{
			name:                "leader_has_admin_access",
			houseID:             house.HouseID,
			rank:                guild.RankLeader,
			requiredPermission:  guildhousing.PermissionAdmin,
			expectHasPermission: true, // Leader has admin access by default
		},
		{
			name:                "recruit_limited_access",
			houseID:             house.HouseID,
			rank:                guild.RankRecruit,
			requiredPermission:  guildhousing.PermissionAdmin,
			expectHasPermission: false,
		},
		{
			name:                "nonexistent_house",
			houseID:             "fake-house",
			rank:                guild.RankLeader,
			requiredPermission:  guildhousing.PermissionView,
			expectHasPermission: false,
		},
		{
			name:                "nil_manager",
			houseID:             house.HouseID,
			rank:                guild.RankLeader,
			requiredPermission:  guildhousing.PermissionView,
			expectHasPermission: false,
			nilManager:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mgr *guildhousing.Manager
			if !tt.nilManager {
				mgr = guildMgr
			}

			service := NewV9ValidationService(nil, nil, mgr, nil)
			hasPermission := service.ValidateGuildPermission(tt.houseID, tt.rank, tt.requiredPermission)

			if hasPermission != tt.expectHasPermission {
				t.Errorf("expected hasPermission=%t, got %t", tt.expectHasPermission, hasPermission)
			}
		})
	}
}

func TestV9ValidationService_GetGuildUpgradeBonus(t *testing.T) {
	guildMgr := guildhousing.NewManager()
	house := guildMgr.CreateGuildHouse("guild1", "owner1", 1)

	tests := []struct {
		name         string
		houseID      string
		expectBonus  float64
		expectExists bool
		nilManager   bool
	}{
		{
			name:         "existing_house_base_tier",
			houseID:      house.HouseID,
			expectBonus:  1.0, // Basic tier = 1.0x multiplier
			expectExists: true,
		},
		{
			name:         "nonexistent_house",
			houseID:      "fake-house",
			expectBonus:  1.0,
			expectExists: false,
		},
		{
			name:         "nil_manager",
			houseID:      house.HouseID,
			expectBonus:  1.0,
			expectExists: false,
			nilManager:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mgr *guildhousing.Manager
			if !tt.nilManager {
				mgr = guildMgr
			}

			service := NewV9ValidationService(nil, nil, mgr, nil)
			bonus, exists := service.GetGuildUpgradeBonus(tt.houseID)

			if bonus != tt.expectBonus {
				t.Errorf("expected bonus %f, got %f", tt.expectBonus, bonus)
			}
			if exists != tt.expectExists {
				t.Errorf("expected exists=%t, got %t", tt.expectExists, exists)
			}
		})
	}
}

func TestV9ValidationService_GetCompanionHome(t *testing.T) {
	tests := []struct {
		name        string
		companionID uint64
		expectHome  string
		nilManager  bool
	}{
		{
			name:        "unassigned_companion",
			companionID: 12345,
			expectHome:  "",
		},
		{
			name:        "nil_manager",
			companionID: 12345,
			expectHome:  "",
			nilManager:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var petHomeMgr *companionhousing.PetHomeManager
			if !tt.nilManager {
				petHomeMgr = companionhousing.NewPetHomeManager()
			}

			service := NewV9ValidationService(nil, petHomeMgr, nil, nil)
			home := service.GetCompanionHome(tt.companionID)

			if home != tt.expectHome {
				t.Errorf("expected home '%s', got '%s'", tt.expectHome, home)
			}
		})
	}
}

func TestV9ValidationService_ManagerGetters(t *testing.T) {
	stationMgr := housingcrafting.NewStationManager()
	petHomeMgr := companionhousing.NewPetHomeManager()
	guildMgr := guildhousing.NewManager()

	service := NewV9ValidationService(stationMgr, petHomeMgr, guildMgr, nil)

	if service.GetStationManager() != stationMgr {
		t.Error("GetStationManager returned wrong manager")
	}
	if service.GetPetHomeManager() != petHomeMgr {
		t.Error("GetPetHomeManager returned wrong manager")
	}
	if service.GetGuildHousingManager() != guildMgr {
		t.Error("GetGuildHousingManager returned wrong manager")
	}
}

func TestGetV9ValidationService_GlobalInstance(t *testing.T) {
	// Initially nil
	oldService := v9ValidationService
	v9ValidationService = nil

	if GetV9ValidationService() != nil {
		t.Error("expected nil before initialization")
	}

	// Set and verify
	service := NewV9ValidationService(nil, nil, nil, nil)
	v9ValidationService = service

	if GetV9ValidationService() != service {
		t.Error("GetV9ValidationService returned wrong instance")
	}

	// Restore
	v9ValidationService = oldService
}

func BenchmarkV9ValidationService_ValidateCraftingBonus(b *testing.B) {
	stationMgr := housingcrafting.NewStationManager()
	service := NewV9ValidationService(stationMgr, nil, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateCraftingBonus("player1", "recipe1", 1.0)
	}
}

func BenchmarkV9ValidationService_ValidateGuildPermission(b *testing.B) {
	guildMgr := guildhousing.NewManager()
	house := guildMgr.CreateGuildHouse("guild1", "owner1", 1)
	service := NewV9ValidationService(nil, nil, guildMgr, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateGuildPermission(house.HouseID, guild.RankLeader, guildhousing.PermissionAdmin)
	}
}
