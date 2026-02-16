package guild_vehicle

import (
	"compress/gzip"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNewFleetManager(t *testing.T) {
	manager := NewFleetManager()
	if manager == nil {
		t.Fatal("NewFleetManager() returned nil")
	}
	if manager.fleets == nil {
		t.Error("fleets map not initialized")
	}
}

func TestFleetManager_CreateFleet(t *testing.T) {
	manager := NewFleetManager()

	err := manager.CreateFleet("guild1", "fleet1", "player1")
	if err != nil {
		t.Fatalf("CreateFleet() error = %v", err)
	}

	// Verify fleet exists
	fleet, err := manager.GetFleet("guild1", "fleet1")
	if err != nil {
		t.Fatalf("GetFleet() error = %v", err)
	}
	if fleet.FleetID != "fleet1" {
		t.Errorf("fleet.FleetID = %v, want fleet1", fleet.FleetID)
	}
	if fleet.GuildID != "guild1" {
		t.Errorf("fleet.GuildID = %v, want guild1", fleet.GuildID)
	}
	if fleet.CommanderID != "player1" {
		t.Errorf("fleet.CommanderID = %v, want player1", fleet.CommanderID)
	}

	// Test duplicate fleet
	err = manager.CreateFleet("guild1", "fleet1", "player2")
	if err == nil {
		t.Error("expected error when creating duplicate fleet")
	}

	// Test empty guildID
	err = manager.CreateFleet("", "fleet2", "player1")
	if err == nil {
		t.Error("expected error when creating fleet with empty guildID")
	}

	// Test empty fleetID
	err = manager.CreateFleet("guild2", "", "player1")
	if err == nil {
		t.Error("expected error when creating fleet with empty fleetID")
	}
}

func TestFleetManager_AddVehicle(t *testing.T) {
	manager := NewFleetManager()

	// Add to non-existent fleet (should auto-create)
	err := manager.AddVehicle("guild1", 1, "fleet1")
	if err != nil {
		t.Fatalf("AddVehicle() error = %v", err)
	}

	// Verify vehicle added
	fleet, err := manager.GetFleet("guild1", "fleet1")
	if err != nil {
		t.Fatalf("GetFleet() error = %v", err)
	}
	if len(fleet.Vehicles) != 1 {
		t.Errorf("fleet.Vehicles count = %v, want 1", len(fleet.Vehicles))
	}
	if _, exists := fleet.Vehicles[1]; !exists {
		t.Error("vehicle 1 not found in fleet")
	}

	// Test adding duplicate vehicle
	err = manager.AddVehicle("guild1", 1, "fleet1")
	if err == nil {
		t.Error("expected error when adding duplicate vehicle")
	}
}

func TestFleetManager_AddVehicleWithType(t *testing.T) {
	manager := NewFleetManager()

	err := manager.AddVehicleWithType("guild1", 1, "fleet1", SiegeCatapult, 500)
	if err != nil {
		t.Fatalf("AddVehicleWithType() error = %v", err)
	}

	fleet, err := manager.GetFleet("guild1", "fleet1")
	if err != nil {
		t.Fatalf("GetFleet() error = %v", err)
	}

	vehicle := fleet.Vehicles[1]
	if vehicle.SiegeType != SiegeCatapult {
		t.Errorf("vehicle.SiegeType = %v, want SiegeCatapult", vehicle.SiegeType)
	}
	if vehicle.MaintenanceCost != 500 {
		t.Errorf("vehicle.MaintenanceCost = %v, want 500", vehicle.MaintenanceCost)
	}

	// Test empty IDs validation
	err = manager.AddVehicleWithType("", 2, "fleet1", SiegeNone, 100)
	if err == nil {
		t.Error("expected error when adding vehicle with empty guildID")
	}
	err = manager.AddVehicleWithType("guild1", 2, "", SiegeNone, 100)
	if err == nil {
		t.Error("expected error when adding vehicle with empty fleetID")
	}
}

func TestFleetManager_RemoveVehicle(t *testing.T) {
	manager := NewFleetManager()
	manager.AddVehicle("guild1", 1, "fleet1")

	err := manager.RemoveVehicle("guild1", 1, "fleet1")
	if err != nil {
		t.Fatalf("RemoveVehicle() error = %v", err)
	}

	fleet, _ := manager.GetFleet("guild1", "fleet1")
	if len(fleet.Vehicles) != 0 {
		t.Errorf("fleet.Vehicles count = %v, want 0", len(fleet.Vehicles))
	}

	// Test removing non-existent vehicle
	err = manager.RemoveVehicle("guild1", 99, "fleet1")
	if err == nil {
		t.Error("expected error when removing non-existent vehicle")
	}

	// Test removing from non-existent fleet
	err = manager.RemoveVehicle("guild1", 1, "fleet99")
	if err == nil {
		t.Error("expected error when removing from non-existent fleet")
	}
}

func TestFleetManager_GrantAccess(t *testing.T) {
	manager := NewFleetManager()
	manager.AddVehicle("guild1", 1, "fleet1")

	err := manager.GrantAccess("guild1", 1, "player1")
	if err != nil {
		t.Fatalf("GrantAccess() error = %v", err)
	}

	// Verify access
	if !manager.CheckAccess("guild1", 1, "player1") {
		t.Error("expected player1 to have access")
	}

	// Test granting access to non-existent vehicle
	err = manager.GrantAccess("guild1", 99, "player1")
	if err == nil {
		t.Error("expected error when granting access to non-existent vehicle")
	}
}

func TestFleetManager_RevokeAccess(t *testing.T) {
	manager := NewFleetManager()
	manager.AddVehicle("guild1", 1, "fleet1")
	manager.GrantAccess("guild1", 1, "player1")

	err := manager.RevokeAccess("guild1", 1, "player1")
	if err != nil {
		t.Fatalf("RevokeAccess() error = %v", err)
	}

	// Verify access revoked
	if manager.CheckAccess("guild1", 1, "player1") {
		t.Error("expected player1 to not have access")
	}

	// Test revoking access from non-existent vehicle
	err = manager.RevokeAccess("guild1", 99, "player1")
	if err == nil {
		t.Error("expected error when revoking access from non-existent vehicle")
	}
}

func TestFleetManager_CheckAccess(t *testing.T) {
	manager := NewFleetManager()
	manager.AddVehicle("guild1", 1, "fleet1")

	// Test no access
	if manager.CheckAccess("guild1", 1, "player1") {
		t.Error("expected no access for player1")
	}

	// Grant and test access
	manager.GrantAccess("guild1", 1, "player1")
	if !manager.CheckAccess("guild1", 1, "player1") {
		t.Error("expected access for player1")
	}

	// Test different player
	if manager.CheckAccess("guild1", 1, "player2") {
		t.Error("expected no access for player2")
	}

	// Test non-existent vehicle
	if manager.CheckAccess("guild1", 99, "player1") {
		t.Error("expected no access for non-existent vehicle")
	}
}

func TestFleetManager_SetFormation(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")

	err := manager.SetFormation("guild1", "fleet1", FormationWedge)
	if err != nil {
		t.Fatalf("SetFormation() error = %v", err)
	}

	fleet, _ := manager.GetFleet("guild1", "fleet1")
	if fleet.Formation != FormationWedge {
		t.Errorf("fleet.Formation = %v, want FormationWedge", fleet.Formation)
	}

	// Test setting formation on non-existent fleet
	err = manager.SetFormation("guild1", "fleet99", FormationLine)
	if err == nil {
		t.Error("expected error when setting formation on non-existent fleet")
	}
}

func TestFleetManager_SetCommander(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")

	err := manager.SetCommander("guild1", "fleet1", "player2")
	if err != nil {
		t.Fatalf("SetCommander() error = %v", err)
	}

	fleet, _ := manager.GetFleet("guild1", "fleet1")
	if fleet.CommanderID != "player2" {
		t.Errorf("fleet.CommanderID = %v, want player2", fleet.CommanderID)
	}

	// Test setting commander on non-existent fleet
	err = manager.SetCommander("guild1", "fleet99", "player3")
	if err == nil {
		t.Error("expected error when setting commander on non-existent fleet")
	}
}

func TestFleetManager_GetFleetBonuses(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")

	// Test with no formation
	bonus := manager.GetFleetBonuses("guild1", "fleet1")
	if bonus.DamageMultiplier != 1.0 {
		t.Errorf("bonus.DamageMultiplier = %v, want 1.0", bonus.DamageMultiplier)
	}

	// Set formation and test
	manager.SetFormation("guild1", "fleet1", FormationWedge)
	bonus = manager.GetFleetBonuses("guild1", "fleet1")
	if bonus.DamageMultiplier != 1.07 {
		t.Errorf("bonus.DamageMultiplier = %v, want 1.07", bonus.DamageMultiplier)
	}

	// Test non-existent fleet
	bonus = manager.GetFleetBonuses("guild1", "fleet99")
	if bonus.Formation != FormationNone {
		t.Errorf("bonus.Formation = %v, want FormationNone", bonus.Formation)
	}
}

func TestFleetManager_CalculateMaintenanceCost(t *testing.T) {
	manager := NewFleetManager()
	manager.AddVehicleWithType("guild1", 1, "fleet1", SiegeNone, 100)
	manager.AddVehicleWithType("guild1", 2, "fleet1", SiegeCatapult, 500)
	manager.AddVehicleWithType("guild1", 3, "fleet1", SiegeBatteringRam, 300)

	cost := manager.CalculateMaintenanceCost("guild1", "fleet1")
	if cost != 900 {
		t.Errorf("CalculateMaintenanceCost() = %v, want 900", cost)
	}

	// Test non-existent fleet
	cost = manager.CalculateMaintenanceCost("guild1", "fleet99")
	if cost != 0 {
		t.Errorf("CalculateMaintenanceCost() for non-existent fleet = %v, want 0", cost)
	}
}

func TestFleetManager_GetFleet(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	manager.AddVehicleWithType("guild1", 1, "fleet1", SiegeNone, 100)

	fleet, err := manager.GetFleet("guild1", "fleet1")
	if err != nil {
		t.Fatalf("GetFleet() error = %v", err)
	}

	// Verify it's a copy (modifications don't affect manager)
	fleet.Formation = FormationWedge
	originalFleet, _ := manager.GetFleet("guild1", "fleet1")
	if originalFleet.Formation == FormationWedge {
		t.Error("GetFleet() should return a copy, not original")
	}

	// Test non-existent fleet
	_, err = manager.GetFleet("guild1", "fleet99")
	if err == nil {
		t.Error("expected error when getting non-existent fleet")
	}
}

func TestFleetManager_GetAllFleets(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	manager.CreateFleet("guild1", "fleet2", "player2")
	manager.CreateFleet("guild2", "fleet1", "player3")

	fleets := manager.GetAllFleets("guild1")
	if len(fleets) != 2 {
		t.Errorf("GetAllFleets(guild1) count = %v, want 2", len(fleets))
	}

	fleets = manager.GetAllFleets("guild2")
	if len(fleets) != 1 {
		t.Errorf("GetAllFleets(guild2) count = %v, want 1", len(fleets))
	}

	fleets = manager.GetAllFleets("guild99")
	if len(fleets) != 0 {
		t.Errorf("GetAllFleets(guild99) count = %v, want 0", len(fleets))
	}
}

func TestFleetManager_GetAllFleets_EmptyManager(t *testing.T) {
	// Test edge case: empty fleet manager
	manager := NewFleetManager()

	fleets := manager.GetAllFleets("any_guild")
	// Go convention: nil slice is equivalent to empty slice for range/len operations
	if len(fleets) != 0 {
		t.Errorf("GetAllFleets() on empty manager = %v fleets, want 0", len(fleets))
	}
}

func TestFleetManager_GetAllFleets_MultipleGuildsWithVehicles(t *testing.T) {
	// Test edge case: multiple guilds with vehicles and shared access
	manager := NewFleetManager()

	// Create fleets for multiple guilds with vehicles
	manager.CreateFleet("guildA", "fleet1", "commanderA")
	manager.CreateFleet("guildA", "fleet2", "commanderA2")
	manager.CreateFleet("guildB", "fleet1", "commanderB")
	manager.CreateFleet("guildC", "fleet1", "commanderC")

	// Add vehicles to each fleet
	manager.AddVehicleWithType("guildA", 1, "fleet1", SiegeCatapult, 500)
	manager.AddVehicleWithType("guildA", 2, "fleet1", SiegeBatteringRam, 300)
	manager.AddVehicleWithType("guildA", 3, "fleet2", SiegeNone, 100)
	manager.AddVehicleWithType("guildB", 4, "fleet1", SiegeTower, 400)
	manager.AddVehicleWithType("guildC", 5, "fleet1", SiegeNone, 100)

	// Grant access to some vehicles
	manager.GrantAccess("guildA", 1, "player1")
	manager.GrantAccess("guildA", 1, "player2")
	manager.GrantAccess("guildB", 4, "player3")

	// Set formations
	manager.SetFormation("guildA", "fleet1", FormationWedge)
	manager.SetFormation("guildB", "fleet1", FormationCircle)

	// Test getting all fleets for each guild
	tests := []struct {
		guildID       string
		expectedCount int
	}{
		{"guildA", 2},
		{"guildB", 1},
		{"guildC", 1},
		{"guildD", 0}, // non-existent guild
	}

	for _, tc := range tests {
		t.Run(tc.guildID, func(t *testing.T) {
			fleets := manager.GetAllFleets(tc.guildID)
			if len(fleets) != tc.expectedCount {
				t.Errorf("GetAllFleets(%s) = %d fleets, want %d", tc.guildID, len(fleets), tc.expectedCount)
			}

			// Verify returned fleets are copies (modification doesn't affect manager)
			for _, fleet := range fleets {
				originalFormation := fleet.Formation
				fleet.Formation = FormationNone // Attempt to modify returned copy

				// Get fleet again and verify it wasn't modified
				if tc.expectedCount > 0 {
					refetch := manager.GetAllFleets(tc.guildID)
					for _, f := range refetch {
						if f.FleetID == fleet.FleetID && f.Formation == FormationNone && originalFormation != FormationNone {
							t.Errorf("GetAllFleets() should return copies, not originals")
						}
					}
				}
			}
		})
	}

	// Verify vehicle data is correctly copied
	guildAFleets := manager.GetAllFleets("guildA")
	totalVehicles := 0
	for _, fleet := range guildAFleets {
		totalVehicles += len(fleet.Vehicles)
	}
	if totalVehicles != 3 {
		t.Errorf("guildA total vehicles = %d, want 3", totalVehicles)
	}
}

func TestFleetManager_GetVehicleFleetID(t *testing.T) {
	manager := NewFleetManager()
	manager.AddVehicle("guild1", 1, "fleet1")
	manager.AddVehicle("guild1", 2, "fleet2")

	fleetID, err := manager.GetVehicleFleetID("guild1", 1)
	if err != nil {
		t.Fatalf("GetVehicleFleetID() error = %v", err)
	}
	if fleetID != "fleet1" {
		t.Errorf("GetVehicleFleetID() = %v, want fleet1", fleetID)
	}

	// Test non-existent vehicle
	_, err = manager.GetVehicleFleetID("guild1", 99)
	if err == nil {
		t.Error("expected error when getting fleet ID for non-existent vehicle")
	}
}

func TestFleetManager_SaveLoad(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	manager.AddVehicleWithType("guild1", 1, "fleet1", SiegeCatapult, 500)
	manager.AddVehicleWithType("guild1", 2, "fleet1", SiegeBatteringRam, 300)
	manager.SetFormation("guild1", "fleet1", FormationWedge)

	// Save
	filename := "test_fleet_save.json.gz"
	defer os.Remove(filename)

	err := manager.Save(filename)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load into new manager
	manager2 := NewFleetManager()
	err = manager2.Load(filename)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded data
	fleet, err := manager2.GetFleet("guild1", "fleet1")
	if err != nil {
		t.Fatalf("GetFleet() after load error = %v", err)
	}
	if fleet.CommanderID != "player1" {
		t.Errorf("loaded fleet.CommanderID = %v, want player1", fleet.CommanderID)
	}
	if fleet.Formation != FormationWedge {
		t.Errorf("loaded fleet.Formation = %v, want FormationWedge", fleet.Formation)
	}
	if len(fleet.Vehicles) != 2 {
		t.Errorf("loaded fleet vehicle count = %v, want 2", len(fleet.Vehicles))
	}

	// Test loading non-existent file
	err = manager2.Load("nonexistent.json.gz")
	if err == nil {
		t.Error("expected error when loading non-existent file")
	}
}

func TestFleetManager_Save_InvalidPath(t *testing.T) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")

	// Test saving to invalid path (directory doesn't exist)
	err := manager.Save("/nonexistent/path/fleets.json.gz")
	if err == nil {
		t.Error("expected error when saving to invalid path")
	}
}

func TestFleetManager_Load_InvalidGzip(t *testing.T) {
	// Create a non-gzip file
	filename := "test_invalid_gzip.json.gz"
	defer os.Remove(filename)

	// Write non-gzip content
	err := os.WriteFile(filename, []byte("not gzip content"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	manager := NewFleetManager()
	err = manager.Load(filename)
	if err == nil {
		t.Error("expected error when loading invalid gzip file")
	}
}

func TestFleetManager_Load_InvalidJSON(t *testing.T) {
	// Create a valid gzip file with invalid JSON
	filename := "test_invalid_json.json.gz"
	defer os.Remove(filename)

	// Create gzip file with invalid JSON
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	gzWriter := gzip.NewWriter(file)
	gzWriter.Write([]byte("{invalid json content"))
	gzWriter.Close()
	file.Close()

	manager := NewFleetManager()
	err = manager.Load(filename)
	if err == nil {
		t.Error("expected error when loading invalid JSON from gzip file")
	}
}

func TestFleetManager_ConcurrentAccess(t *testing.T) {
	manager := NewFleetManager()
	var wg sync.WaitGroup

	// Concurrent fleet creation
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			manager.CreateFleet("guild1", string(rune('A'+idx)), "player1")
		}(i)
	}
	wg.Wait()

	// Verify all created
	fleets := manager.GetAllFleets("guild1")
	if len(fleets) != 10 {
		t.Errorf("concurrent creation resulted in %v fleets, want 10", len(fleets))
	}

	// Concurrent vehicle additions
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			manager.AddVehicle("guild1", uint64(idx), "A")
		}(i)
	}
	wg.Wait()

	// Verify all added
	fleet, _ := manager.GetFleet("guild1", "A")
	if len(fleet.Vehicles) != 20 {
		t.Errorf("concurrent addition resulted in %v vehicles, want 20", len(fleet.Vehicles))
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.GetFleet("guild1", "A")
			manager.CheckAccess("guild1", 5, "player1")
			manager.GetFleetBonuses("guild1", "A")
		}()
	}
	wg.Wait()
}

// Benchmarks

func BenchmarkFleetManager_CreateFleet(b *testing.B) {
	manager := NewFleetManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.CreateFleet("guild1", string(rune('A'+i%26)), "player1")
	}
}

func BenchmarkFleetManager_AddVehicle(b *testing.B) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.AddVehicle("guild1", uint64(i), "fleet1")
	}
}

func BenchmarkFleetManager_CheckAccess(b *testing.B) {
	manager := NewFleetManager()
	manager.AddVehicle("guild1", 1, "fleet1")
	manager.GrantAccess("guild1", 1, "player1")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.CheckAccess("guild1", 1, "player1")
	}
}

func BenchmarkFleetManager_GetFleetBonuses(b *testing.B) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	manager.SetFormation("guild1", "fleet1", FormationWedge)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.GetFleetBonuses("guild1", "fleet1")
	}
}

func BenchmarkFleetManager_CalculateMaintenanceCost(b *testing.B) {
	manager := NewFleetManager()
	for i := uint64(0); i < 20; i++ {
		manager.AddVehicleWithType("guild1", i, "fleet1", SiegeNone, 100)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.CalculateMaintenanceCost("guild1", "fleet1")
	}
}

func BenchmarkFleetManager_GetFleet(b *testing.B) {
	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	for i := uint64(0); i < 20; i++ {
		manager.AddVehicle("guild1", i, "fleet1")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.GetFleet("guild1", "fleet1")
	}
}

// TimeProvider determinism validation tests

func TestTimeProvider_RealTimeProvider(t *testing.T) {
	rtp := RealTimeProvider{}
	before := time.Now()
	got := rtp.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("RealTimeProvider.Now() = %v, want between %v and %v", got, before, after)
	}
}

func TestTimeProvider_FixedTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	ftp := FixedTimeProvider{FixedTime: fixedTime}

	for i := 0; i < 10; i++ {
		got := ftp.Now()
		if !got.Equal(fixedTime) {
			t.Errorf("FixedTimeProvider.Now() call %d = %v, want %v", i, got, fixedTime)
		}
	}
}

func TestTimeProvider_SetAndReset(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	got := now()
	if !got.Equal(fixedTime) {
		t.Errorf("now() with FixedTimeProvider = %v, want %v", got, fixedTime)
	}

	ResetTimeProvider()
	got = now()
	if got.Equal(fixedTime) {
		t.Error("now() after ResetTimeProvider should not return fixed time")
	}
}

func TestFleetDeterminism_CreateFleet(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	// Create two managers with same operations - timestamps must match
	m1 := NewFleetManager()
	m2 := NewFleetManager()

	m1.CreateFleet("guild1", "fleet1", "cmd1")
	m2.CreateFleet("guild1", "fleet1", "cmd1")

	f1, _ := m1.GetFleet("guild1", "fleet1")
	f2, _ := m2.GetFleet("guild1", "fleet1")

	if !f1.CreatedAt.Equal(f2.CreatedAt) {
		t.Errorf("CreatedAt mismatch: %v vs %v", f1.CreatedAt, f2.CreatedAt)
	}
	if !f1.UpdatedAt.Equal(f2.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: %v vs %v", f1.UpdatedAt, f2.UpdatedAt)
	}
	if !f1.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want fixed time %v", f1.CreatedAt, fixedTime)
	}
}

func TestFleetDeterminism_AddVehicle(t *testing.T) {
	fixedTime := time.Date(2026, 6, 1, 8, 30, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	m1 := NewFleetManager()
	m2 := NewFleetManager()

	m1.AddVehicleWithType("guild1", 1, "fleet1", SiegeCatapult, 500)
	m2.AddVehicleWithType("guild1", 1, "fleet1", SiegeCatapult, 500)

	f1, _ := m1.GetFleet("guild1", "fleet1")
	f2, _ := m2.GetFleet("guild1", "fleet1")

	v1 := f1.Vehicles[1]
	v2 := f2.Vehicles[1]

	if !v1.AddedAt.Equal(v2.AddedAt) {
		t.Errorf("AddedAt mismatch: %v vs %v", v1.AddedAt, v2.AddedAt)
	}
	if !v1.LastMaintenance.Equal(v2.LastMaintenance) {
		t.Errorf("LastMaintenance mismatch: %v vs %v", v1.LastMaintenance, v2.LastMaintenance)
	}
	if !v1.AddedAt.Equal(fixedTime) {
		t.Errorf("AddedAt = %v, want fixed time %v", v1.AddedAt, fixedTime)
	}
}

func TestFleetDeterminism_UpdatedAtConsistency(t *testing.T) {
	fixedTime := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	m1 := NewFleetManager()
	m2 := NewFleetManager()

	// Perform identical sequences of operations
	ops := func(m *FleetManager) {
		m.CreateFleet("g1", "f1", "p1")
		m.AddVehicle("g1", 1, "f1")
		m.AddVehicle("g1", 2, "f1")
		m.GrantAccess("g1", 1, "p2")
		m.SetFormation("g1", "f1", FormationWedge)
		m.SetCommander("g1", "f1", "p2")
		m.RevokeAccess("g1", 1, "p2")
		m.RemoveVehicle("g1", 2, "f1")
	}

	ops(m1)
	ops(m2)

	f1, _ := m1.GetFleet("g1", "f1")
	f2, _ := m2.GetFleet("g1", "f1")

	if !f1.UpdatedAt.Equal(f2.UpdatedAt) {
		t.Errorf("UpdatedAt after identical ops: %v vs %v", f1.UpdatedAt, f2.UpdatedAt)
	}
	if !f1.UpdatedAt.Equal(fixedTime) {
		t.Errorf("UpdatedAt = %v, want fixed time %v", f1.UpdatedAt, fixedTime)
	}
}

func TestFleetDeterminism_SaveLoadWithFixedTime(t *testing.T) {
	fixedTime := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	manager := NewFleetManager()
	manager.CreateFleet("guild1", "fleet1", "player1")
	manager.AddVehicleWithType("guild1", 1, "fleet1", SiegeBatteringRam, 300)

	filename := "test_determinism_save.json.gz"
	defer os.Remove(filename)

	if err := manager.Save(filename); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded := NewFleetManager()
	if err := loaded.Load(filename); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	fleet, _ := loaded.GetFleet("guild1", "fleet1")
	if !fleet.CreatedAt.Equal(fixedTime) {
		t.Errorf("Loaded CreatedAt = %v, want %v", fleet.CreatedAt, fixedTime)
	}

	vehicle := fleet.Vehicles[1]
	if !vehicle.AddedAt.Equal(fixedTime) {
		t.Errorf("Loaded AddedAt = %v, want %v", vehicle.AddedAt, fixedTime)
	}
	if !vehicle.LastMaintenance.Equal(fixedTime) {
		t.Errorf("Loaded LastMaintenance = %v, want %v", vehicle.LastMaintenance, fixedTime)
	}
}
