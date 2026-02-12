//go:build !android && !ios
// +build !android,!ios

package main

import (
	"bytes"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	guildvehicle "github.com/opd-ai/venture/pkg/integration/guild_vehicle"
	"github.com/sirupsen/logrus"
)

// TestFleetManager_Integration verifies FleetManager is properly initialized and functional
func TestFleetManager_Integration(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.DebugLevel)
	seed := int64(12345)

	// Initialize V8 systems to get FleetManager
	guildManager, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	if guildManager == nil {
		t.Fatal("guildManager is nil")
	}

	if fleetManager == nil {
		t.Fatal("fleetManager is nil after V8 initialization")
	}

	// Test basic fleet operations
	guildID := "test_guild_warriors"
	fleetID := "primary_fleet"
	commanderID := "player_commander"

	// Create a fleet
	err := fleetManager.CreateFleet(guildID, fleetID, commanderID)
	if err != nil {
		t.Fatalf("CreateFleet failed: %v", err)
	}

	// Verify fleet exists
	fleet, err := fleetManager.GetFleet(guildID, fleetID)
	if err != nil {
		t.Fatalf("GetFleet failed: %v", err)
	}

	if fleet.FleetID != fleetID {
		t.Errorf("fleet.FleetID = %v, want %v", fleet.FleetID, fleetID)
	}

	if fleet.GuildID != guildID {
		t.Errorf("fleet.GuildID = %v, want %v", fleet.GuildID, guildID)
	}

	if fleet.CommanderID != commanderID {
		t.Errorf("fleet.CommanderID = %v, want %v", fleet.CommanderID, commanderID)
	}

	t.Log("FleetManager integration test passed: fleet created and retrieved successfully")
}

// TestFleetManager_VehicleOperations verifies vehicle management in fleets
func TestFleetManager_VehicleOperations(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "test_guild"
	fleetID := "siege_fleet"
	vehicleID := uint64(12345)

	// Add vehicle to fleet (auto-creates fleet if needed)
	err := fleetManager.AddVehicleWithType(
		guildID,
		vehicleID,
		fleetID,
		guildvehicle.SiegeBatteringRam,
		100, // maintenance cost
	)
	if err != nil {
		t.Fatalf("AddVehicleWithType failed: %v", err)
	}

	// Verify vehicle is in fleet
	fleet, err := fleetManager.GetFleet(guildID, fleetID)
	if err != nil {
		t.Fatalf("GetFleet failed: %v", err)
	}

	if len(fleet.Vehicles) != 1 {
		t.Errorf("fleet has %d vehicles, expected 1", len(fleet.Vehicles))
	}

	vehicle, exists := fleet.Vehicles[vehicleID]
	if !exists {
		t.Fatal("vehicle not found in fleet")
	}

	if vehicle.SiegeType != guildvehicle.SiegeBatteringRam {
		t.Errorf("vehicle.SiegeType = %v, want SiegeBatteringRam", vehicle.SiegeType)
	}

	// Test access control
	playerID := "test_player"
	err = fleetManager.GrantAccess(guildID, vehicleID, playerID)
	if err != nil {
		t.Fatalf("GrantAccess failed: %v", err)
	}

	hasAccess := fleetManager.CheckAccess(guildID, vehicleID, playerID)
	if !hasAccess {
		t.Error("player should have access after GrantAccess")
	}

	// Test access revocation
	err = fleetManager.RevokeAccess(guildID, vehicleID, playerID)
	if err != nil {
		t.Fatalf("RevokeAccess failed: %v", err)
	}

	hasAccess = fleetManager.CheckAccess(guildID, vehicleID, playerID)
	if hasAccess {
		t.Error("player should not have access after RevokeAccess")
	}

	t.Log("FleetManager vehicle operations test passed")
}

// TestFleetManager_FormationBonuses verifies formation bonus calculations
func TestFleetManager_FormationBonuses(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "test_guild"
	fleetID := "combat_fleet"
	commanderID := "commander"

	// Create fleet
	err := fleetManager.CreateFleet(guildID, fleetID, commanderID)
	if err != nil {
		t.Fatalf("CreateFleet failed: %v", err)
	}

	// Test formation bonuses
	formations := []struct {
		formation            guildvehicle.FormationType
		expectedDamageBonus  float64
		expectedDefenseBonus float64
	}{
		{guildvehicle.FormationLine, 1.05, 1.0},   // 5% damage
		{guildvehicle.FormationWedge, 1.07, 1.0},  // 7% damage
		{guildvehicle.FormationColumn, 1.0, 1.10}, // 10% defense
		{guildvehicle.FormationCircle, 1.0, 1.08}, // 8% defense
		{guildvehicle.FormationNone, 1.0, 1.0},    // no bonus
	}

	for _, tc := range formations {
		t.Run(tc.formation.String(), func(t *testing.T) {
			err := fleetManager.SetFormation(guildID, fleetID, tc.formation)
			if err != nil {
				t.Fatalf("SetFormation failed: %v", err)
			}

			bonuses := fleetManager.GetFleetBonuses(guildID, fleetID)

			if bonuses.DamageMultiplier != tc.expectedDamageBonus {
				t.Errorf("DamageMultiplier = %v, want %v", bonuses.DamageMultiplier, tc.expectedDamageBonus)
			}

			if bonuses.DefenseMultiplier != tc.expectedDefenseBonus {
				t.Errorf("DefenseMultiplier = %v, want %v", bonuses.DefenseMultiplier, tc.expectedDefenseBonus)
			}

			if bonuses.Formation != tc.formation {
				t.Errorf("Formation = %v, want %v", bonuses.Formation, tc.formation)
			}
		})
	}

	t.Log("FleetManager formation bonuses test passed")
}

// TestFleetManager_MaintenanceCost verifies maintenance cost calculations
func TestFleetManager_MaintenanceCost(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "test_guild"
	fleetID := "expensive_fleet"

	// Add multiple vehicles with different costs
	vehicles := []struct {
		id              uint64
		maintenanceCost int
	}{
		{100, 50},
		{101, 75},
		{102, 100},
	}

	expectedTotal := 0
	for _, v := range vehicles {
		err := fleetManager.AddVehicleWithType(guildID, v.id, fleetID, guildvehicle.SiegeNone, v.maintenanceCost)
		if err != nil {
			t.Fatalf("AddVehicleWithType failed for vehicle %d: %v", v.id, err)
		}
		expectedTotal += v.maintenanceCost
	}

	// Calculate total maintenance
	totalCost := fleetManager.CalculateMaintenanceCost(guildID, fleetID)

	if totalCost != expectedTotal {
		t.Errorf("CalculateMaintenanceCost = %d, want %d", totalCost, expectedTotal)
	}

	t.Log("FleetManager maintenance cost test passed")
}

// TestFleetManager_MultipleFleets verifies handling of multiple fleets per guild
func TestFleetManager_MultipleFleets(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "test_guild"
	fleetIDs := []string{"fleet_alpha", "fleet_beta", "fleet_gamma"}

	// Create multiple fleets
	for _, fleetID := range fleetIDs {
		err := fleetManager.CreateFleet(guildID, fleetID, "commander")
		if err != nil {
			t.Fatalf("CreateFleet failed for %s: %v", fleetID, err)
		}
	}

	// Get all fleets for guild
	allFleets := fleetManager.GetAllFleets(guildID)

	if len(allFleets) != len(fleetIDs) {
		t.Errorf("GetAllFleets returned %d fleets, expected %d", len(allFleets), len(fleetIDs))
	}

	// Verify each fleet exists
	fleetMap := make(map[string]bool)
	for _, fleet := range allFleets {
		fleetMap[fleet.FleetID] = true
	}

	for _, expectedID := range fleetIDs {
		if !fleetMap[expectedID] {
			t.Errorf("fleet %s not found in GetAllFleets", expectedID)
		}
	}

	t.Log("FleetManager multiple fleets test passed")
}

// TestFleetManager_ThreadSafety verifies concurrent access is safe
func TestFleetManager_ThreadSafety(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "concurrent_guild"
	fleetID := "concurrent_fleet"

	// Create initial fleet
	err := fleetManager.CreateFleet(guildID, fleetID, "commander")
	if err != nil {
		t.Fatalf("CreateFleet failed: %v", err)
	}

	// Simulate concurrent operations
	done := make(chan bool, 10)

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			_, _ = fleetManager.GetFleet(guildID, fleetID)
		}()
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		vehicleID := uint64(i + 1000)
		go func(id uint64) {
			defer func() { done <- true }()
			_ = fleetManager.AddVehicle(guildID, id, fleetID)
		}(vehicleID)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify fleet is still accessible
	fleet, err := fleetManager.GetFleet(guildID, fleetID)
	if err != nil {
		t.Fatalf("GetFleet failed after concurrent operations: %v", err)
	}

	if fleet == nil {
		t.Error("fleet is nil after concurrent operations")
	}

	t.Log("FleetManager thread safety test passed")
}

// BenchmarkFleetManager_AddVehicle measures vehicle addition performance
func BenchmarkFleetManager_AddVehicle(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "bench_guild"
	fleetID := "bench_fleet"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vehicleID := uint64(i)
		_ = fleetManager.AddVehicle(guildID, vehicleID, fleetID)
	}
}

// BenchmarkFleetManager_GetFleetBonuses measures bonus calculation performance
func BenchmarkFleetManager_GetFleetBonuses(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	seed := int64(12345)

	_, fleetManager := initializeV8SystemsServer(world, seed, "test-server", logger)

	guildID := "bench_guild"
	fleetID := "bench_fleet"

	// Create fleet with formation
	_ = fleetManager.CreateFleet(guildID, fleetID, "commander")
	_ = fleetManager.SetFormation(guildID, fleetID, guildvehicle.FormationWedge)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fleetManager.GetFleetBonuses(guildID, fleetID)
	}
}
