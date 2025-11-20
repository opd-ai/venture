// Command traderoutetest demonstrates automated trade route functionality.
//
// Usage:
//
//	traderoutetest                       # Run default test scenario
//	traderoutetest -mode single          # Single route test
//	traderoutetest -mode multi           # Multiple routes test
//	traderoutetest -mode escort          # Escort mission test
//	traderoutetest -mode encounter       # Bandit encounter test
//	traderoutetest -mode optimization    # Route optimization test
//	traderoutetest -verbose              # Enable detailed logging
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/integration/trade_routes"
)

var (
	mode    = flag.String("mode", "single", "Test mode: single, multi, escort, encounter, optimization")
	verbose = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	fmt.Println("=== Trade Route System Test ===")
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Println()

	switch *mode {
	case "single":
		testSingleRoute()
	case "multi":
		testMultipleRoutes()
	case "escort":
		testEscortMission()
	case "encounter":
		testBanditEncounter()
	case "optimization":
		testRouteOptimization()
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func testSingleRoute() {
	rm := trade_routes.NewRouteManager("test-server", 12345)
	rm.Start()
	defer rm.Stop()

	fmt.Println("Creating trade route...")
	route, err := rm.CreateRoute("region-alpha", "region-beta", 1000.0)
	if err != nil {
		log.Fatal(err)
	}

	printRoute(route)

	fmt.Println("\nGenerating caravan vehicle...")
	caravan, err := rm.CreateCaravan(54321, "fantasy")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Caravan: %s\n", caravan.Name)

	fmt.Println("\nStarting route...")
	err = rm.StartRoute(route.ID, 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nRoute started successfully!")
	fmt.Println("Simulating travel (5 updates over 10 seconds)...")

	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		rm.UpdateRoutes()
		updatedRoute, _ := rm.GetRoute(route.ID)
		fmt.Printf("Progress: %.1f%% | Status: %s\n", updatedRoute.Progress*100, updatedRoute.Status)
	}
}

func testMultipleRoutes() {
	rm := trade_routes.NewRouteManager("test-server", 12345)
	rm.Start()
	defer rm.Stop()

	fmt.Println("Creating 5 trade routes...")
	routes := []*trade_routes.TradeRoute{}

	regions := [][]string{
		{"region-alpha", "region-beta"},
		{"region-gamma", "region-delta"},
		{"region-epsilon", "region-zeta"},
		{"region-alpha", "region-gamma"},
		{"region-delta", "region-epsilon"},
	}

	for i, pair := range regions {
		cargoValue := 1000.0 + float64(i)*500.0
		route, err := rm.CreateRoute(pair[0], pair[1], cargoValue)
		if err != nil {
			log.Fatal(err)
		}
		routes = append(routes, route)

		err = rm.StartRoute(route.ID, uint64(i+1))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Route %d: %s | Cargo Value: %.0f | Profit Margin: %.1f%% | Danger: %.1f%%\n",
			i+1, route.Name, cargoValue, route.ProfitMargin*100, route.DangerLevel*100)
	}

	fmt.Println("\nActive routes:")
	activeRoutes := rm.GetActiveRoutes()
	fmt.Printf("Total active: %d\n", len(activeRoutes))

	fmt.Println("\nSimulating simultaneous travel...")
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		rm.UpdateRoutes()
		activeRoutes = rm.GetActiveRoutes()
		fmt.Printf("Update %d: %d routes active\n", i+1, len(activeRoutes))
	}
}

func testEscortMission() {
	rm := trade_routes.NewRouteManager("test-server", 12345)
	rm.Start()
	defer rm.Stop()

	fmt.Println("Creating high-danger trade route...")
	route, err := rm.CreateRoute("region-alpha", "region-omega", 5000.0)
	if err != nil {
		log.Fatal(err)
	}

	printRoute(route)

	err = rm.StartRoute(route.ID, 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nCreating escort missions for 3 players...")
	playerIDs := []uint64{100, 200, 300}
	baseReward := 50.0

	for _, playerID := range playerIDs {
		mission, err := rm.CreateEscortMission(route.ID, playerID, baseReward)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Mission for Player %d:\n", playerID)
		fmt.Printf("  Reward: %.2f gold\n", mission.Reward)
		fmt.Printf("  Bonus: %.2f gold (for defeating attacks)\n", mission.BonusReward)
		fmt.Printf("  Status: %s\n", mission.Status)

		// Add player as escort
		err = rm.AddEscort(route.ID, playerID)
		if err != nil {
			log.Fatal(err)
		}
	}

	updatedRoute, _ := rm.GetRoute(route.ID)
	fmt.Printf("\nRoute now has %d escorts\n", len(updatedRoute.EscortPlayers))
}

func testBanditEncounter() {
	rm := trade_routes.NewRouteManager("test-server", 12345)
	rm.Start()
	defer rm.Stop()

	fmt.Println("Creating high-danger route to trigger encounters...")
	// Create route with high danger (long distance = higher danger)
	route, err := rm.CreateRoute("region-alpha", "region-omega", 5000.0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Route: %s\n", route.Name)
	fmt.Printf("Danger Level: %.1f%%\n", route.DangerLevel*100)

	err = rm.StartRoute(route.ID, 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSimulating travel to trigger bandit encounters...")
	fmt.Println("(Bandit spawn rate: 0.1-1.0 per hour based on danger)")

	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		rm.UpdateRoutes()
		updatedRoute, _ := rm.GetRoute(route.ID)

		fmt.Printf("Update %d: Progress %.1f%% | Attacks: %d | Status: %s\n",
			i+1, updatedRoute.Progress*100, updatedRoute.BanditAttacks, updatedRoute.Status)

		if updatedRoute.Status == trade_routes.StatusUnderAttack {
			fmt.Println("  ⚔️  BANDIT ATTACK IN PROGRESS!")
		}
	}
}

func testRouteOptimization() {
	rm := trade_routes.NewRouteManager("test-server", 12345)

	fmt.Println("Creating routes with varying danger levels...")
	routes := []*trade_routes.TradeRoute{}

	cargoValues := []float64{1000.0, 2500.0, 5000.0}
	regionPairs := [][]string{
		{"region-alpha", "region-beta"},  // Low danger (close)
		{"region-alpha", "region-delta"}, // Medium danger
		{"region-alpha", "region-omega"}, // High danger (far)
	}

	for i, pair := range regionPairs {
		route, err := rm.CreateRoute(pair[0], pair[1], cargoValues[i])
		if err != nil {
			log.Fatal(err)
		}
		routes = append(routes, route)
	}

	fmt.Println("\nOptimizing routes...")
	for i, route := range routes {
		optimization := rm.OptimizeRoute(route)

		fmt.Printf("\n--- Route %d: %s ---\n", i+1, route.Name)
		fmt.Printf("Expected Profit: %.2f gold\n", optimization.ExpectedProfit)
		fmt.Printf("Risk-Adjusted Profit: %.2f gold (%.1f%% of expected)\n",
			optimization.RiskAdjustedProfit,
			(optimization.RiskAdjustedProfit/optimization.ExpectedProfit)*100)
		fmt.Printf("Server Hops: %d\n", optimization.ServerHops)
		fmt.Printf("Estimated Travel Time: %s\n", optimization.EstimatedTravelTime)
		fmt.Printf("Danger Zones: %d\n", len(optimization.DangerZones))

		if len(optimization.DangerZones) > 0 {
			fmt.Println("\nDanger Zones:")
			for j, zone := range optimization.DangerZones {
				fmt.Printf("  %d. %s\n", j+1, zone.RegionID)
				fmt.Printf("     Danger: %.1f%% | Spawn Rate: %.2f/hr | Escorts: %d\n",
					zone.DangerLevel*100, zone.BanditSpawnRate, zone.RecommendedEscorts)
				fmt.Printf("     %s\n", zone.Description)
			}
		}
	}
}

func printRoute(route *trade_routes.TradeRoute) {
	fmt.Printf("\n--- Route Details ---\n")
	fmt.Printf("ID: %s\n", route.ID)
	fmt.Printf("Name: %s\n", route.Name)
	fmt.Printf("Start: %s → End: %s\n", route.StartRegion, route.EndRegion)
	fmt.Printf("Status: %s\n", route.Status)
	fmt.Printf("Profit Margin: %.1f%%\n", route.ProfitMargin*100)
	fmt.Printf("Danger Level: %.1f%%\n", route.DangerLevel*100)
	fmt.Printf("Travel Time: %s\n", route.TravelTime)
	fmt.Printf("Success Rate: %.1f%%\n", route.SuccessRate*100)
	fmt.Printf("Cargo Items: %d\n", len(route.Cargo))

	if *verbose {
		fmt.Println("\nCargo Manifest:")
		totalValue := 0.0
		for i, item := range route.Cargo {
			fmt.Printf("  %d. %s (×%d)\n", i+1, item.ItemName, item.Quantity)
			fmt.Printf("     Purchase: %.2f | Target: %.2f | Profit: %.2f\n",
				item.PurchasePrice, item.TargetPrice, item.Profit)
			totalValue += item.PurchasePrice
		}
		fmt.Printf("\nTotal Cargo Value: %.2f gold\n", totalValue)
	}
}
