//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/integration/trade_routes"
)

// TestTradeRouteManagerWrapper verifies the wrapper adapts RouteManager to ECS System interface.
func TestTradeRouteManagerWrapper(t *testing.T) {
	manager := trade_routes.NewRouteManager("test-server", 12345)
	wrapper := &tradeRouteManagerWrapper{system: manager}

	// Create mock entities (trade routes don't directly use entities parameter)
	entities := []*engine.Entity{}

	// Verify Update() doesn't panic
	wrapper.Update(entities, 0.016) // 16ms delta (60 FPS)

	// Create a route to verify manager state
	route, err := manager.CreateRoute("region-a", "region-b", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	// Start the route with a test caravan ID
	err = manager.StartRoute(route.ID, 12345)
	if err != nil {
		t.Fatalf("Failed to start route: %v", err)
	}

	// Verify route is active
	activeRoutes := manager.GetActiveRoutes()
	if len(activeRoutes) != 1 {
		t.Errorf("Expected 1 active route, got %d", len(activeRoutes))
	}

	// Call Update via wrapper to simulate ECS tick
	wrapper.Update(entities, 1.0) // 1 second delta

	// Verify route still exists
	retrievedRoute, err := manager.GetRoute(route.ID)
	if err != nil {
		t.Errorf("Failed to retrieve route after update: %v", err)
	}

	if retrievedRoute.Progress < 0.0 || retrievedRoute.Progress > 1.0 {
		t.Errorf("Invalid route progress: %f (should be 0.0-1.0)", retrievedRoute.Progress)
	}
}

// TestTradeRouteManagerWrapperDeterminism verifies route generation is deterministic.
func TestTradeRouteManagerWrapperDeterminism(t *testing.T) {
	seed := int64(54321)

	// Create two managers with same seed
	manager1 := trade_routes.NewRouteManager("test-server", seed)
	manager2 := trade_routes.NewRouteManager("test-server", seed)

	// Create identical routes
	route1, err1 := manager1.CreateRoute("region-x", "region-y", 2000.0)
	route2, err2 := manager2.CreateRoute("region-x", "region-y", 2000.0)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create routes: err1=%v, err2=%v", err1, err2)
	}

	// Verify cargo generation is deterministic (same seed = same cargo)
	if len(route1.Cargo) != len(route2.Cargo) {
		t.Errorf("Cargo count mismatch: %d vs %d", len(route1.Cargo), len(route2.Cargo))
	}

	// Verify route names are deterministic
	if route1.Name != route2.Name {
		t.Errorf("Route names differ: %s vs %s", route1.Name, route2.Name)
	}

	// Verify profit margins are identical
	if route1.ProfitMargin != route2.ProfitMargin {
		t.Errorf("Profit margins differ: %f vs %f", route1.ProfitMargin, route2.ProfitMargin)
	}
}

// TestTradeRouteManagerWrapperConcurrentUpdates verifies thread safety.
func TestTradeRouteManagerWrapperConcurrentUpdates(t *testing.T) {
	manager := trade_routes.NewRouteManager("test-server", 99999)
	wrapper := &tradeRouteManagerWrapper{system: manager}

	// Create multiple routes
	for i := 0; i < 5; i++ {
		route, err := manager.CreateRoute("region-a", "region-b", 1000.0)
		if err != nil {
			t.Fatalf("Failed to create route %d: %v", i, err)
		}
		err = manager.StartRoute(route.ID, uint64(10000+i))
		if err != nil {
			t.Fatalf("Failed to start route %d: %v", i, err)
		}
	}

	// Simulate concurrent ECS updates
	done := make(chan bool, 3)
	entities := []*engine.Entity{}

	for i := 0; i < 3; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				wrapper.Update(entities, 0.1)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify routes still exist and are valid
	activeRoutes := manager.GetActiveRoutes()
	if len(activeRoutes) != 5 {
		t.Errorf("Expected 5 active routes after concurrent updates, got %d", len(activeRoutes))
	}

	for _, route := range activeRoutes {
		if route.Progress < 0.0 || route.Progress > 1.0 {
			t.Errorf("Invalid route progress after concurrent updates: %f", route.Progress)
		}
	}
}

// TestTradeRouteManagerWrapperIntegration verifies full ECS integration.
func TestTradeRouteManagerWrapperIntegration(t *testing.T) {
	world := engine.NewWorld()
	manager := trade_routes.NewRouteManager("integration-server", 11111)
	wrapper := &tradeRouteManagerWrapper{system: manager}

	// Add wrapper as a system
	world.AddSystem(wrapper)

	// Create a route
	route, err := manager.CreateRoute("region-start", "region-end", 5000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	// Start the route
	err = manager.StartRoute(route.ID, 77777)
	if err != nil {
		t.Fatalf("Failed to start route: %v", err)
	}

	// Add player escort
	playerID := uint64(12345)
	err = manager.AddEscort(route.ID, playerID)
	if err != nil {
		t.Fatalf("Failed to add escort: %v", err)
	}

	// Simulate ECS updates (world.Update would call all systems)
	world.Update(0.016) // 60 FPS tick

	// Verify route was updated
	updatedRoute, err := manager.GetRoute(route.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated route: %v", err)
	}

	// Verify escort was added
	if len(updatedRoute.EscortPlayers) != 1 {
		t.Errorf("Expected 1 escort player, got %d", len(updatedRoute.EscortPlayers))
	}

	if updatedRoute.EscortPlayers[0] != playerID {
		t.Errorf("Expected escort player ID %d, got %d", playerID, updatedRoute.EscortPlayers[0])
	}

	// Verify route is still active
	if updatedRoute.Status != trade_routes.StatusActive {
		t.Errorf("Expected route status Active, got %s", updatedRoute.Status.String())
	}
}

// TestTradeRouteManagerWrapperEscortMissions verifies escort mission integration.
func TestTradeRouteManagerWrapperEscortMissions(t *testing.T) {
	manager := trade_routes.NewRouteManager("mission-server", 22222)
	wrapper := &tradeRouteManagerWrapper{system: manager}

	// Create and start a route
	route, err := manager.CreateRoute("town-a", "town-b", 3000.0)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	err = manager.StartRoute(route.ID, 88888)
	if err != nil {
		t.Fatalf("Failed to start route: %v", err)
	}

	// Create an escort mission
	playerID := uint64(99999)
	baseReward := 100.0
	mission, err := manager.CreateEscortMission(route.ID, playerID, baseReward)
	if err != nil {
		t.Fatalf("Failed to create escort mission: %v", err)
	}

	// Verify mission was created correctly
	if mission.RouteID != route.ID {
		t.Errorf("Mission route ID mismatch: expected %s, got %s", route.ID, mission.RouteID)
	}

	if mission.PlayerID != playerID {
		t.Errorf("Mission player ID mismatch: expected %d, got %d", playerID, mission.PlayerID)
	}

	// Verify reward scales with danger level
	expectedMinReward := baseReward * (1.0 + route.DangerLevel)
	if mission.Reward < expectedMinReward*0.99 || mission.Reward > expectedMinReward*1.01 {
		t.Errorf("Mission reward %f doesn't match expected %f (±1%%)", mission.Reward, expectedMinReward)
	}

	// Simulate ECS update
	entities := []*engine.Entity{}
	wrapper.Update(entities, 0.5)

	// Verify route still exists after update
	_, err = manager.GetRoute(route.ID)
	if err != nil {
		t.Errorf("Route disappeared after update: %v", err)
	}
}
