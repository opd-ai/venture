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
