package guild_vehicle

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// FleetManager manages guild vehicle fleets with thread-safe operations
type FleetManager struct {
	fleets              map[string]*Fleet // GuildID -> Fleet map
	mu                  sync.RWMutex
	membershipValidator MembershipValidator // optional; validates guild membership on GrantAccess
	vehicleSyncer       VehicleSyncer       // optional; propagates fleet component to ECS entities
	structureDamager    StructureDamager    // optional; applies siege damage to territory structures
}

// NewFleetManager creates a new fleet manager instance
func NewFleetManager() *FleetManager {
	logrus.WithFields(logrus.Fields{
		"system_name": "fleet_manager",
	}).Debug("fleet manager created")
	return &FleetManager{
		fleets: make(map[string]*Fleet),
	}
}

// CreateFleet creates a new fleet for a guild
func (m *FleetManager) CreateFleet(guildID, fleetID, commanderID string) error {
	if guildID == "" || fleetID == "" {
		return fmt.Errorf("guildID and fleetID must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if fleet already exists
	key := m.getFleetKey(guildID, fleetID)
	if _, exists := m.fleets[key]; exists {
		return fmt.Errorf("fleet %s already exists for guild %s", fleetID, guildID)
	}

	// Create new fleet
	fleet := &Fleet{
		FleetID:     fleetID,
		GuildID:     guildID,
		Vehicles:    make(map[uint64]*GuildVehicle),
		Formation:   FormationNone,
		CommanderID: commanderID,
		CreatedAt:   now(),
		UpdatedAt:   now(),
	}

	m.fleets[key] = fleet
	return nil
}

// AddVehicle adds a vehicle to a guild fleet
func (m *FleetManager) AddVehicle(guildID string, vehicleID uint64, fleetID string) error {
	return m.AddVehicleWithType(guildID, vehicleID, fleetID, SiegeNone, 100)
}

// AddVehicleWithType adds a vehicle with specified siege type and maintenance cost
func (m *FleetManager) AddVehicleWithType(guildID string, vehicleID uint64, fleetID string, siegeType SiegeEngineType, maintenanceCost int) error {
	if guildID == "" || fleetID == "" {
		return fmt.Errorf("guildID and fleetID must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create fleet
	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		fleet = &Fleet{
			FleetID:     fleetID,
			GuildID:     guildID,
			Vehicles:    make(map[uint64]*GuildVehicle),
			Formation:   FormationNone,
			CommanderID: "", // No commander yet
			CreatedAt:   now(),
			UpdatedAt:   now(),
		}
		m.fleets[key] = fleet
	}

	// Check if vehicle already in fleet
	if _, exists := fleet.Vehicles[vehicleID]; exists {
		return fmt.Errorf("vehicle %d already in fleet %s", vehicleID, fleetID)
	}

	// Add vehicle
	vehicle := &GuildVehicle{
		VehicleID:       vehicleID,
		GuildID:         guildID,
		FleetID:         fleetID,
		SiegeType:       siegeType,
		SharedAccess:    make(map[string]bool),
		MaintenanceCost: maintenanceCost,
		AddedAt:         now(),
		LastMaintenance: now(),
	}

	fleet.Vehicles[vehicleID] = vehicle
	fleet.UpdatedAt = now()

	// Notify ECS syncer so the vehicle entity receives a GuildVehicleFleetComponent.
	syncer := m.vehicleSyncer
	if syncer != nil {
		comp := &GuildVehicleFleetComponent{
			GuildID:           guildID,
			FleetID:           fleetID,
			SiegeType:         siegeType,
			FormationPosition: len(fleet.Vehicles) - 1,
		}
		syncer.SyncVehicleFleetComponent(vehicleID, comp)
	}

	return nil
}

// RemoveVehicle removes a vehicle from a fleet
func (m *FleetManager) RemoveVehicle(guildID string, vehicleID uint64, fleetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		return fmt.Errorf("fleet %s not found for guild %s", fleetID, guildID)
	}

	if _, exists := fleet.Vehicles[vehicleID]; !exists {
		return fmt.Errorf("vehicle %d not found in fleet %s", vehicleID, fleetID)
	}

	delete(fleet.Vehicles, vehicleID)
	fleet.UpdatedAt = now()

	// Notify ECS syncer so the vehicle entity loses its GuildVehicleFleetComponent.
	syncer := m.vehicleSyncer
	if syncer != nil {
		syncer.ClearVehicleFleetComponent(vehicleID)
	}

	return nil
}

// GrantAccess grants a player access to a vehicle.
// If a MembershipValidator has been set via SetMembershipValidator, the player must
// be an active member of the guild; otherwise the request is rejected.
func (m *FleetManager) GrantAccess(guildID string, vehicleID uint64, playerID string) error {
	// Read validator under RLock to avoid a data race with SetMembershipValidator.
	m.mu.RLock()
	validator := m.membershipValidator
	m.mu.RUnlock()

	if validator != nil {
		if !validator.IsMember(guildID, playerID) {
			return fmt.Errorf("player %s is not a member of guild %s", playerID, guildID)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find vehicle across all fleets
	for _, fleet := range m.fleets {
		if fleet.GuildID != guildID {
			continue
		}

		if vehicle, exists := fleet.Vehicles[vehicleID]; exists {
			vehicle.SharedAccess[playerID] = true
			fleet.UpdatedAt = now()
			return nil
		}
	}

	return fmt.Errorf("vehicle %d not found in any fleet for guild %s", vehicleID, guildID)
}

// RevokeAccess revokes a player's access to a vehicle
func (m *FleetManager) RevokeAccess(guildID string, vehicleID uint64, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find vehicle across all fleets
	for _, fleet := range m.fleets {
		if fleet.GuildID != guildID {
			continue
		}

		if vehicle, exists := fleet.Vehicles[vehicleID]; exists {
			delete(vehicle.SharedAccess, playerID)
			fleet.UpdatedAt = now()
			return nil
		}
	}

	return fmt.Errorf("vehicle %d not found in any fleet for guild %s", vehicleID, guildID)
}

// CheckAccess checks if a player has access to a vehicle
func (m *FleetManager) CheckAccess(guildID string, vehicleID uint64, playerID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, fleet := range m.fleets {
		if fleet.GuildID != guildID {
			continue
		}

		if vehicle, exists := fleet.Vehicles[vehicleID]; exists {
			return vehicle.HasAccess(playerID)
		}
	}

	return false
}

// SetFormation sets the formation for a fleet
func (m *FleetManager) SetFormation(guildID, fleetID string, formation FormationType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		return fmt.Errorf("fleet %s not found for guild %s", fleetID, guildID)
	}

	fleet.Formation = formation
	fleet.UpdatedAt = now()

	return nil
}

// SetCommander sets the fleet commander
func (m *FleetManager) SetCommander(guildID, fleetID, commanderID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		return fmt.Errorf("fleet %s not found for guild %s", fleetID, guildID)
	}

	fleet.CommanderID = commanderID
	fleet.UpdatedAt = now()

	return nil
}

// GetFleetBonuses calculates formation bonuses for a fleet
func (m *FleetManager) GetFleetBonuses(guildID, fleetID string) FleetBonus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		return GetFormationBonus(FormationNone)
	}

	return GetFormationBonus(fleet.Formation)
}

// CalculateMaintenanceCost calculates total daily maintenance cost for a fleet
func (m *FleetManager) CalculateMaintenanceCost(guildID, fleetID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		return 0
	}

	return fleet.GetTotalMaintenanceCost()
}

// GetFleet retrieves a fleet by guild and fleet ID
func (m *FleetManager) GetFleet(guildID, fleetID string) (*Fleet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		return nil, fmt.Errorf("fleet %s not found for guild %s", fleetID, guildID)
	}

	// Return a copy to prevent external modification
	fleetCopy := *fleet
	fleetCopy.Vehicles = make(map[uint64]*GuildVehicle)
	for id, vehicle := range fleet.Vehicles {
		vehicleCopy := *vehicle
		vehicleCopy.SharedAccess = make(map[string]bool)
		for playerID, access := range vehicle.SharedAccess {
			vehicleCopy.SharedAccess[playerID] = access
		}
		fleetCopy.Vehicles[id] = &vehicleCopy
	}

	return &fleetCopy, nil
}

// GetAllFleets retrieves all fleets for a guild
func (m *FleetManager) GetAllFleets(guildID string) []*Fleet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var fleets []*Fleet
	for _, fleet := range m.fleets {
		if fleet.GuildID == guildID {
			// Make a copy
			fleetCopy := *fleet
			fleetCopy.Vehicles = make(map[uint64]*GuildVehicle)
			for id, vehicle := range fleet.Vehicles {
				vehicleCopy := *vehicle
				vehicleCopy.SharedAccess = make(map[string]bool)
				for playerID, access := range vehicle.SharedAccess {
					vehicleCopy.SharedAccess[playerID] = access
				}
				fleetCopy.Vehicles[id] = &vehicleCopy
			}
			fleets = append(fleets, &fleetCopy)
		}
	}

	return fleets
}

// GetVehicleFleetID returns the fleet ID for a vehicle
func (m *FleetManager) GetVehicleFleetID(guildID string, vehicleID uint64) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, fleet := range m.fleets {
		if fleet.GuildID != guildID {
			continue
		}

		if _, exists := fleet.Vehicles[vehicleID]; exists {
			return fleet.FleetID, nil
		}
	}

	return "", fmt.Errorf("vehicle %d not found in any fleet for guild %s", vehicleID, guildID)
}

// Save persists fleet data to a file with gzip compression
func (m *FleetManager) Save(filename string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	gzWriter := gzip.NewWriter(file)

	encoder := json.NewEncoder(gzWriter)
	if err := encoder.Encode(m.fleets); err != nil {
		gzWriter.Close()
		file.Close()
		return fmt.Errorf("failed to encode fleets: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		file.Close()
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// Load loads fleet data from a file with gzip decompression
func (m *FleetManager) Load(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	decoder := json.NewDecoder(gzReader)
	fleets := make(map[string]*Fleet)
	if err := decoder.Decode(&fleets); err != nil {
		if errors.Is(err, io.EOF) {
			// Empty file treated as empty fleet state
			m.fleets = fleets
			return nil
		}
		return fmt.Errorf("failed to decode fleets: %w", err)
	}

	m.fleets = fleets
	return nil
}

// getFleetKey generates a unique key for fleet lookup
func (m *FleetManager) getFleetKey(guildID, fleetID string) string {
	return guildID + ":" + fleetID
}
