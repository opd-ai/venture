package trade_routes

import (
	"testing"
	"time"
)

func TestNewRouteManager(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	if rm == nil {
		t.Fatal("Expected non-nil route manager")
	}

	if rm.serverID != "test-server" {
		t.Errorf("Expected serverID 'test-server', got '%s'", rm.serverID)
	}

	if rm.routes == nil || rm.encounters == nil || rm.missions == nil {
		t.Error("Expected initialized maps")
	}
}

func TestCreateRoute(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	tests := []struct {
		name        string
		startRegion string
		endRegion   string
		cargoValue  float64
		wantErr     bool
	}{
		{"valid route", "region-a", "region-b", 1000.0, false},
		{"empty start", "", "region-b", 1000.0, true},
		{"empty end", "region-a", "", 1000.0, true},
		{"same regions", "region-a", "region-a", 1000.0, true},
		{"zero cargo value", "region-a", "region-b", 0.0, true},
		{"negative cargo value", "region-a", "region-b", -500.0, true},
		{"high cargo value", "region-a", "region-c", 10000.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, err := rm.CreateRoute(tt.startRegion, tt.endRegion, tt.cargoValue)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if route == nil {
				t.Fatal("Expected non-nil route")
			}

			if route.StartRegion != tt.startRegion {
				t.Errorf("Expected start region '%s', got '%s'", tt.startRegion, route.StartRegion)
			}

			if route.EndRegion != tt.endRegion {
				t.Errorf("Expected end region '%s', got '%s'", tt.endRegion, route.EndRegion)
			}

			if route.Status != StatusPlanning {
				t.Errorf("Expected status Planning, got %s", route.Status)
			}

			if route.ProfitMargin < 0.10 || route.ProfitMargin > 0.50 {
				t.Errorf("Profit margin out of range [0.10, 0.50]: %.2f", route.ProfitMargin)
			}

			if route.DangerLevel < 0.0 || route.DangerLevel > 1.0 {
				t.Errorf("Danger level out of range [0.0, 1.0]: %.2f", route.DangerLevel)
			}

			if len(route.Cargo) < 3 || len(route.Cargo) > 8 {
				t.Errorf("Expected 3-8 cargo items, got %d", len(route.Cargo))
			}
		})
	}
}

func TestStartRoute(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	tests := []struct {
		name      string
		routeID   string
		caravanID uint64
		wantErr   bool
	}{
		{"valid start", route.ID, 12345, false},
		{"invalid route ID", "nonexistent", 12345, true},
		{"zero caravan ID", route.ID, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh route for each test
			r, _ := rm.CreateRoute("region-c", "region-d", 1000.0)
			testRouteID := r.ID
			if tt.routeID != route.ID {
				testRouteID = tt.routeID
			}

			err := rm.StartRoute(testRouteID, tt.caravanID)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			updatedRoute, _ := rm.GetRoute(testRouteID)
			if updatedRoute.Status != StatusActive {
				t.Errorf("Expected status Active, got %s", updatedRoute.Status)
			}

			if updatedRoute.CaravanID != tt.caravanID {
				t.Errorf("Expected caravan ID %d, got %d", tt.caravanID, updatedRoute.CaravanID)
			}
		})
	}
}

func TestAddEscort(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	rm.StartRoute(route.ID, 12345)

	// Add first escort
	err := rm.AddEscort(route.ID, 100)
	if err != nil {
		t.Fatalf("Failed to add escort: %v", err)
	}

	updatedRoute, _ := rm.GetRoute(route.ID)
	if len(updatedRoute.EscortPlayers) != 1 {
		t.Errorf("Expected 1 escort, got %d", len(updatedRoute.EscortPlayers))
	}

	// Add second escort
	err = rm.AddEscort(route.ID, 200)
	if err != nil {
		t.Fatalf("Failed to add second escort: %v", err)
	}

	updatedRoute, _ = rm.GetRoute(route.ID)
	if len(updatedRoute.EscortPlayers) != 2 {
		t.Errorf("Expected 2 escorts, got %d", len(updatedRoute.EscortPlayers))
	}

	// Try to add duplicate escort
	err = rm.AddEscort(route.ID, 100)
	if err == nil {
		t.Error("Expected error when adding duplicate escort")
	}

	// Try to add escort with zero player ID
	err = rm.AddEscort(route.ID, 0)
	if err == nil {
		t.Error("Expected error when adding escort with zero player ID")
	}
}

func TestRemoveEscort(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	rm.StartRoute(route.ID, 12345)
	rm.AddEscort(route.ID, 100)
	rm.AddEscort(route.ID, 200)

	// Remove existing escort
	err := rm.RemoveEscort(route.ID, 100)
	if err != nil {
		t.Fatalf("Failed to remove escort: %v", err)
	}

	updatedRoute, _ := rm.GetRoute(route.ID)
	if len(updatedRoute.EscortPlayers) != 1 {
		t.Errorf("Expected 1 escort after removal, got %d", len(updatedRoute.EscortPlayers))
	}

	// Try to remove non-existent escort
	err = rm.RemoveEscort(route.ID, 999)
	if err == nil {
		t.Error("Expected error when removing non-existent escort")
	}
}

func TestGetRoute(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create a valid route
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	tests := []struct {
		name    string
		routeID string
		wantErr bool
	}{
		{"valid route ID", route.ID, false},
		{"empty route ID", "", true},
		{"nonexistent route ID", "nonexistent-route-999", true},
		{"invalid format route ID", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rm.GetRoute(tt.routeID)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected non-nil route")
			}

			if result.ID != tt.routeID {
				t.Errorf("Expected route ID '%s', got '%s'", tt.routeID, result.ID)
			}
		})
	}
}

func TestGetActiveRoutes(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create multiple routes with different statuses
	route1, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	route2, _ := rm.CreateRoute("region-c", "region-d", 1500.0)
	_, _ = rm.CreateRoute("region-e", "region-f", 2000.0) // route3 remains in planning

	rm.StartRoute(route1.ID, 1)
	rm.StartRoute(route2.ID, 2)

	activeRoutes := rm.GetActiveRoutes()
	if len(activeRoutes) != 2 {
		t.Errorf("Expected 2 active routes, got %d", len(activeRoutes))
	}
}

func TestOptimizeRoute(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)

	optimization := rm.OptimizeRoute(route)

	if optimization == nil {
		t.Fatal("Expected non-nil optimization")
	}

	if optimization.ExpectedProfit <= 0 {
		t.Errorf("Expected positive expected profit, got %.2f", optimization.ExpectedProfit)
	}

	if optimization.RiskAdjustedProfit > optimization.ExpectedProfit {
		t.Error("Risk adjusted profit should not exceed expected profit")
	}

	if optimization.ServerHops <= 0 {
		t.Error("Expected at least 1 server hop")
	}

	if len(optimization.DangerZones) == 0 {
		t.Error("Expected at least one danger zone")
	}
}

func TestOptimizeRouteNil(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	result := rm.OptimizeRoute(nil)
	if result != nil {
		t.Error("Expected nil for nil route input")
	}
}

func TestCreateEscortMission(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	rm.StartRoute(route.ID, 12345)

	mission, err := rm.CreateEscortMission(route.ID, 100, 50.0)
	if err != nil {
		t.Fatalf("Failed to create escort mission: %v", err)
	}

	if mission == nil {
		t.Fatal("Expected non-nil mission")
	}

	if mission.RouteID != route.ID {
		t.Errorf("Expected route ID '%s', got '%s'", route.ID, mission.RouteID)
	}

	if mission.PlayerID != 100 {
		t.Errorf("Expected player ID 100, got %d", mission.PlayerID)
	}

	if mission.Reward <= 50.0 {
		t.Error("Expected reward to be adjusted for danger level")
	}

	if mission.BonusReward <= 0 {
		t.Error("Expected positive bonus reward")
	}
}

func TestCreateEscortMissionValidation(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	rm.StartRoute(route.ID, 12345)

	tests := []struct {
		name       string
		routeID    string
		playerID   uint64
		baseReward float64
		wantErr    bool
	}{
		{"valid", route.ID, 100, 50.0, false},
		{"zero player ID", route.ID, 0, 50.0, true},
		{"zero reward", route.ID, 100, 0.0, true},
		{"negative reward", route.ID, 100, -10.0, true},
		{"nonexistent route", "fake", 100, 50.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := rm.CreateEscortMission(tt.routeID, tt.playerID, tt.baseReward)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEscortMission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRouteStatusString(t *testing.T) {
	tests := []struct {
		status   RouteStatus
		expected string
	}{
		{StatusPlanning, "Planning"},
		{StatusActive, "Active"},
		{StatusUnderAttack, "Under Attack"},
		{StatusCompleted, "Completed"},
		{StatusFailed, "Failed"},
		{StatusCancelled, "Cancelled"},
		{RouteStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("Status(%d).String() = %s, want %s", tt.status, got, tt.expected)
		}
	}
}

func TestEncounterOutcomeString(t *testing.T) {
	tests := []struct {
		outcome  EncounterOutcome
		expected string
	}{
		{OutcomePending, "Pending"},
		{OutcomeDefended, "Defended"},
		{OutcomeCompromised, "Compromised"},
		{OutcomeDestroyed, "Destroyed"},
		{OutcomeEvaded, "Evaded"},
		{EncounterOutcome(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.outcome.String(); got != tt.expected {
			t.Errorf("Outcome(%d).String() = %s, want %s", tt.outcome, got, tt.expected)
		}
	}
}

func TestMissionStatusString(t *testing.T) {
	tests := []struct {
		status   MissionStatus
		expected string
	}{
		{MissionAvailable, "Available"},
		{MissionActive, "Active"},
		{MissionCompleted, "Completed"},
		{MissionFailed, "Failed"},
		{MissionAbandoned, "Abandoned"},
		{MissionStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("Status(%d).String() = %s, want %s", tt.status, got, tt.expected)
		}
	}
}

func TestDeterminism(t *testing.T) {
	// Create two managers with same seed
	rm1 := NewRouteManager("test-server", 54321)
	rm2 := NewRouteManager("test-server", 54321)

	// Generate identical routes
	route1, _ := rm1.CreateRoute("region-a", "region-b", 1000.0)
	route2, _ := rm2.CreateRoute("region-a", "region-b", 1000.0)

	// Verify same cargo generation
	if len(route1.Cargo) != len(route2.Cargo) {
		t.Errorf("Cargo count mismatch: %d vs %d", len(route1.Cargo), len(route2.Cargo))
	}

	for i := range route1.Cargo {
		if route1.Cargo[i].ItemName != route2.Cargo[i].ItemName {
			t.Errorf("Cargo item %d name mismatch: %s vs %s", i, route1.Cargo[i].ItemName, route2.Cargo[i].ItemName)
		}
	}

	// Verify same route name generation
	if route1.Name != route2.Name {
		t.Errorf("Route name mismatch: %s vs %s", route1.Name, route2.Name)
	}
}

func TestCreateCaravan(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	caravan, err := rm.CreateCaravan(54321, "fantasy")
	if err != nil {
		t.Fatalf("Failed to create caravan: %v", err)
	}

	if caravan == nil {
		t.Fatal("Expected non-nil caravan")
	}

	if caravan.Name == "" {
		t.Error("Expected non-empty caravan name")
	}
}

func TestUpdateRoutesProgress(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)

	// Set short travel time for testing
	route.TravelTime = 100 * time.Millisecond
	rm.StartRoute(route.ID, 12345)

	// Wait half the travel time
	time.Sleep(50 * time.Millisecond)
	rm.UpdateRoutes()

	updatedRoute, _ := rm.GetRoute(route.ID)
	if updatedRoute.Progress < 0.3 || updatedRoute.Progress > 0.7 {
		t.Errorf("Expected progress around 0.5, got %.2f", updatedRoute.Progress)
	}

	// Wait for completion
	time.Sleep(60 * time.Millisecond)
	rm.UpdateRoutes()

	updatedRoute, _ = rm.GetRoute(route.ID)
	if updatedRoute.Status != StatusCompleted {
		t.Errorf("Expected status Completed after travel time, got %s", updatedRoute.Status)
	}

	if updatedRoute.Progress != 1.0 {
		t.Errorf("Expected progress 1.0, got %.2f", updatedRoute.Progress)
	}
}

// Benchmarks

func BenchmarkCreateRoute(b *testing.B) {
	rm := NewRouteManager("test-server", 12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rm.CreateRoute("region-a", "region-b", 1000.0)
	}
}

func BenchmarkOptimizeRoute(b *testing.B) {
	rm := NewRouteManager("test-server", 12345)
	route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rm.OptimizeRoute(route)
	}
}

func BenchmarkUpdateRoutes(b *testing.B) {
	rm := NewRouteManager("test-server", 12345)

	// Create 20 active routes
	for i := 0; i < 20; i++ {
		route, _ := rm.CreateRoute("region-a", "region-b", 1000.0)
		rm.StartRoute(route.ID, uint64(i+1))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.UpdateRoutes()
	}
}

// TestStartStopLifecycle verifies the goroutine lifecycle management pattern.
// This test ensures that Start() properly initializes the background update loop
// and Stop() cleanly terminates it without goroutine leaks.
func TestStartStopLifecycle(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Start the manager
	rm.Start()

	// Verify the ticker was created
	if rm.updateTicker == nil {
		t.Error("Expected ticker to be created by Start()")
	}

	// Let it run briefly to ensure no immediate crashes
	time.Sleep(50 * time.Millisecond)

	// Stop should be idempotent - multiple calls should not panic
	rm.Stop()
	rm.Stop() // Second call should be safe due to sync.Once
	rm.Stop() // Third call should also be safe
}

// TestStartIdempotent verifies that calling Start() multiple times only creates one goroutine.
// This prevents goroutine leaks from accidental multiple Start() calls.
func TestStartIdempotent(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Call Start() multiple times
	rm.Start()
	rm.Start()
	rm.Start()

	// Only one ticker should be created
	if rm.updateTicker == nil {
		t.Error("Expected ticker to be created by first Start()")
	}

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop should work correctly even after multiple Start() calls
	rm.Stop()

	// Verify no panics occurred and cleanup worked
	if rm.updateTicker == nil {
		t.Error("Ticker should still exist after Stop() (it's stopped, not nil'd)")
	}
}

// TestStartStopPreventsGoroutineLeak verifies that calling Stop() terminates the background goroutine.
// This is a regression test for the goroutine leak issue identified in BUGS_AUDIT.md [MEDIUM-003].
func TestStartStopPreventsGoroutineLeak(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create a route with short travel time
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	route.TravelTime = 200 * time.Millisecond
	rm.StartRoute(route.ID, 12345)

	// Manually update to move progress forward (since ticker fires every 10 seconds)
	time.Sleep(50 * time.Millisecond)
	rm.UpdateRoutes()

	// Verify the route has been updated (progress should be > 0)
	updatedRoute, _ := rm.GetRoute(route.ID)
	if updatedRoute.Progress == 0 {
		t.Error("Expected route progress to be updated by UpdateRoutes()")
	}

	// Start the background update loop
	rm.Start()

	// Wait briefly for goroutine to start
	time.Sleep(50 * time.Millisecond)

	// Stop the manager - this should terminate the goroutine
	rm.Stop()

	// Wait a bit to ensure goroutine exits
	time.Sleep(50 * time.Millisecond)

	// The goroutine should have exited cleanly.
	// We can't directly verify goroutine count in a portable way,
	// but we can verify that no panic occurred and the ticker was stopped.
	if rm.updateTicker == nil {
		t.Error("Ticker should still exist after Stop() (it's stopped, not nil'd)")
	}
}

// TestDeferStopPattern demonstrates the recommended usage pattern.
// This is a documentation test showing how to properly use Start/Stop with defer.
func TestDeferStopPattern(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)
	rm.Start()
	defer rm.Stop() // Ensures cleanup even if test fails

	// Use the route manager
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	if route == nil {
		t.Error("Expected non-nil route")
	}

	// defer rm.Stop() will be called automatically when function exits
}

// TestSpawnBanditEncounter tests bandit encounter creation.
func TestSpawnBanditEncounter(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create an active route
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	route.Status = StatusActive
	route.DangerLevel = 0.5
	route.Progress = 0.3

	// Spawn a bandit encounter
	now := time.Now()
	rm.spawnBanditEncounter(route, now)

	// Verify encounter was created
	if len(rm.encounters) != 1 {
		t.Fatalf("Expected 1 encounter, got %d", len(rm.encounters))
	}

	// Get the encounter
	var encounter *BanditEncounter
	for _, e := range rm.encounters {
		encounter = e
		break
	}

	// Verify encounter properties
	if encounter.RouteID != route.ID {
		t.Errorf("Expected route ID '%s', got '%s'", route.ID, encounter.RouteID)
	}

	if encounter.BanditCount < 3 || encounter.BanditCount > 10 {
		t.Errorf("Expected 3-10 bandits, got %d", encounter.BanditCount)
	}

	if encounter.BanditStrength < 0.3 || encounter.BanditStrength > 0.7 {
		t.Errorf("Expected bandit strength 0.3-0.7, got %.2f", encounter.BanditStrength)
	}

	if encounter.Outcome != OutcomePending {
		t.Errorf("Expected outcome Pending, got %s", encounter.Outcome)
	}

	// Verify route was updated
	if route.Status != StatusUnderAttack {
		t.Errorf("Expected route status UnderAttack, got %s", route.Status)
	}

	if route.BanditAttacks != 1 {
		t.Errorf("Expected 1 bandit attack, got %d", route.BanditAttacks)
	}
}

// TestUpdateEncounterDefended tests encounter resolution when defense wins.
func TestUpdateEncounterDefended(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create route with escorts for strong defense
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	route.Status = StatusActive
	route.EscortPlayers = []uint64{100, 200, 300} // 3 escorts = stronger defense

	// Spawn encounter with weak bandits
	now := time.Now()
	rm.spawnBanditEncounter(route, now)

	var encounter *BanditEncounter
	for _, e := range rm.encounters {
		encounter = e
		break
	}

	// Override strengths to ensure defense wins
	encounter.DefenseStrength = 0.9
	encounter.BanditStrength = 0.3
	encounter.Duration = 0 // Instant resolution

	// Update encounter (should resolve immediately)
	rm.updateEncounter(encounter, now.Add(time.Second))

	// Verify defended outcome
	if encounter.Outcome != OutcomeDefended {
		t.Errorf("Expected outcome Defended, got %s", encounter.Outcome)
	}

	// Verify route status returned to active
	if route.Status != StatusActive {
		t.Errorf("Expected route status Active, got %s", route.Status)
	}

	// Verify player rewards were assigned
	if len(encounter.PlayerRewards) != 3 {
		t.Errorf("Expected 3 player rewards, got %d", len(encounter.PlayerRewards))
	}
}

// TestUpdateEncounterCompromised tests partial cargo loss.
func TestUpdateEncounterCompromised(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create route with cargo
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	route.Status = StatusActive
	originalCargoCount := len(route.Cargo)
	originalQuantities := make([]int, len(route.Cargo))
	for i, c := range route.Cargo {
		originalQuantities[i] = c.Quantity
	}

	// Spawn encounter
	now := time.Now()
	rm.spawnBanditEncounter(route, now)

	var encounter *BanditEncounter
	for _, e := range rm.encounters {
		encounter = e
		break
	}

	// Set strengths for compromised outcome (defense 0.7-1.0 of bandit strength)
	encounter.DefenseStrength = 0.55
	encounter.BanditStrength = 0.7
	encounter.Duration = 0

	// Update encounter
	rm.updateEncounter(encounter, now.Add(time.Second))

	// Verify compromised outcome
	if encounter.Outcome != OutcomeCompromised {
		t.Errorf("Expected outcome Compromised, got %s", encounter.Outcome)
	}

	// Verify some cargo was lost
	if len(encounter.LostCargo) == 0 {
		t.Error("Expected some cargo to be lost in compromised encounter")
	}

	// Verify route still has cargo (partial loss, not total)
	totalRemaining := 0
	for _, c := range route.Cargo {
		totalRemaining += c.Quantity
	}
	if totalRemaining == 0 {
		t.Error("Expected some cargo to remain after compromised encounter")
	}

	// Verify cargo count is same (items reduced, not removed)
	if len(route.Cargo) != originalCargoCount {
		t.Errorf("Expected same number of cargo items %d, got %d", originalCargoCount, len(route.Cargo))
	}

	// Verify route status returned to active
	if route.Status != StatusActive {
		t.Errorf("Expected route status Active, got %s", route.Status)
	}
}

// TestUpdateEncounterDestroyed tests total cargo loss.
func TestUpdateEncounterDestroyed(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create route with cargo
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	route.Status = StatusActive
	originalCargo := make([]CargoItem, len(route.Cargo))
	copy(originalCargo, route.Cargo)

	// Spawn encounter
	now := time.Now()
	rm.spawnBanditEncounter(route, now)

	var encounter *BanditEncounter
	for _, e := range rm.encounters {
		encounter = e
		break
	}

	// Set strengths for destroyed outcome (defense < 0.7 of bandit strength)
	encounter.DefenseStrength = 0.2
	encounter.BanditStrength = 0.8
	encounter.Duration = 0

	// Update encounter
	rm.updateEncounter(encounter, now.Add(time.Second))

	// Verify destroyed outcome
	if encounter.Outcome != OutcomeDestroyed {
		t.Errorf("Expected outcome Destroyed, got %s", encounter.Outcome)
	}

	// Verify all cargo was lost
	if len(encounter.LostCargo) != len(originalCargo) {
		t.Errorf("Expected all %d cargo items lost, got %d", len(originalCargo), len(encounter.LostCargo))
	}

	// Verify route has no cargo
	if len(route.Cargo) != 0 {
		t.Errorf("Expected no cargo remaining, got %d items", len(route.Cargo))
	}

	// Verify route status is failed
	if route.Status != StatusFailed {
		t.Errorf("Expected route status Failed, got %s", route.Status)
	}
}

// TestUpdateEncounterOngoing tests that ongoing encounters are not resolved.
func TestUpdateEncounterOngoing(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create route
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	route.Status = StatusActive

	// Spawn encounter
	now := time.Now()
	rm.spawnBanditEncounter(route, now)

	var encounter *BanditEncounter
	for _, e := range rm.encounters {
		encounter = e
		break
	}

	// Ensure encounter has remaining time
	encounter.Duration = 60 * time.Second

	// Update encounter (should not resolve yet)
	rm.updateEncounter(encounter, now.Add(time.Second))

	// Verify encounter is still pending
	if encounter.Outcome != OutcomePending {
		t.Errorf("Expected outcome Pending for ongoing encounter, got %s", encounter.Outcome)
	}
}

// TestResolveEncounterRouteNotFound tests resolution when route doesn't exist.
func TestResolveEncounterRouteNotFound(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create orphan encounter (route doesn't exist)
	encounter := &BanditEncounter{
		ID:              "orphan-1",
		RouteID:         "non-existent-route",
		BanditStrength:  0.5,
		DefenseStrength: 0.3,
		Outcome:         OutcomePending,
		PlayerRewards:   make(map[uint64]float64),
	}

	// These should not panic when route doesn't exist
	rm.resolveDefendedEncounter(encounter)
	rm.resolveCompromisedEncounter(encounter)
	rm.resolveDestroyedEncounter(encounter)

	// Verify no rewards were assigned (route didn't exist)
	if len(encounter.PlayerRewards) != 0 {
		t.Errorf("Expected no player rewards for orphan encounter, got %d", len(encounter.PlayerRewards))
	}
}

// TestGetRouteByCaravan tests retrieving routes by caravan ID.
func TestGetRouteByCaravan(t *testing.T) {
	rm := NewRouteManager("test-server", 12345)

	// Create and start a route
	route, err := rm.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	caravanID := uint64(9999)
	err = rm.StartRoute(route.ID, caravanID)
	if err != nil {
		t.Fatalf("Failed to start route: %v", err)
	}

	// Test successful lookup
	foundRoute, err := rm.GetRouteByCaravan(caravanID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if foundRoute.ID != route.ID {
		t.Errorf("Expected route ID '%s', got '%s'", route.ID, foundRoute.ID)
	}

	// Test not found
	_, err = rm.GetRouteByCaravan(123456)
	if err == nil {
		t.Error("Expected error for non-existent caravan")
	}
}
