// Phase 42 test tool demonstrates territory control, bounty board, and meta-game systems.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/sirupsen/logrus"
)

var (
	verbose  = flag.Bool("verbose", false, "Enable verbose logging")
	testType = flag.String("test", "all", "Test type: territory, bounty, ranking, metagame, all")
	duration = flag.Int("duration", 10, "Simulation duration in seconds")
	servers  = flag.Int("servers", 5, "Number of servers to simulate")
	seed     = flag.Int64("seed", 12345, "Random seed for deterministic generation")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	fmt.Println("=== Phase 42: Territory Control & Meta-Game Test ===")
	fmt.Printf("Test Type: %s\n", *testType)
	fmt.Printf("Duration: %d seconds\n", *duration)
	fmt.Printf("Servers: %d\n", *servers)
	fmt.Printf("Seed: %d\n\n", *seed)

	switch *testType {
	case "territory":
		testTerritorySystem()
	case "bounty":
		testBountySystem(logger)
	case "ranking":
		testRankingSystem()
	case "metagame":
		testMetaGameSystem()
	case "all":
		testTerritorySystem()
		fmt.Println()
		testBountySystem(logger)
		fmt.Println()
		testRankingSystem()
		fmt.Println()
		testMetaGameSystem()
	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		flag.Usage()
	}
}

func testTerritorySystem() {
	fmt.Println("--- Territory System Test ---")

	tm := world.NewTerritoryManager()

	for i := 0; i < *servers-1; i++ {
		serverA := fmt.Sprintf("server%d", i+1)
		serverB := fmt.Sprintf("server%d", i+2)
		zoneID := fmt.Sprintf("zone_%d_%d", i+1, i+2)

		zone := tm.CreateBorderZone(zoneID, serverA, serverB, 3)
		fmt.Printf("Created border zone: %s (%s vs %s)\n", zoneID, serverA, serverB)
		fmt.Printf("  Control points: %d\n", len(zone.ControlPoints))
		fmt.Printf("  Resource bonus: %.1f%%\n", zone.ResourceBonus*100)
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(*duration) * time.Second)
	tick := 0

	for time.Now().Before(endTime) {
		zones := tm.GetAllZones()
		for _, zone := range zones {
			for i, cp := range zone.ControlPoints {
				attackers := (tick + i) % 3
				defenders := (tick + i + 1) % 2

				var faction string
				if attackers > 0 {
					faction = zone.ServerA
				}

				tm.UpdateControlPoint(zone.ZoneID, i, attackers, defenders, faction)

				if tick%5 == 0 && i == 0 {
					fmt.Printf("[T+%ds] Zone %s CP%d: Progress=%.2f, Attackers=%d, Defenders=%d\n",
						tick, zone.ZoneID, i, cp.CaptureProgress, attackers, defenders)
				}
			}
		}

		tick++
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n--- Territory Summary ---")
	for i := 1; i <= *servers; i++ {
		faction := fmt.Sprintf("server%d", i)
		controlled := tm.GetControlledZones(faction)
		bonus := tm.GetResourceBonus(faction)
		fmt.Printf("Server %d: %d zones controlled, %.1f%% resource bonus\n",
			i, len(controlled), bonus*100)
	}

	contested := tm.GetContestedZones()
	fmt.Printf("Contested zones: %d\n", len(contested))
}

func testBountySystem(logger *logrus.Logger) {
	fmt.Println("--- Bounty System Test ---")

	bs := engine.NewBountySystem(nil, logger)

	bountyTypes := []engine.ObjectiveType{
		engine.ObjectiveKill,
		engine.ObjectiveDeliver,
		engine.ObjectiveEscort,
		engine.ObjectiveExplore,
		engine.ObjectiveCraft,
	}

	for i := 0; i < 10; i++ {
		issuer := fmt.Sprintf("server%d", (i%*servers)+1)
		target := fmt.Sprintf("server%d", ((i+1)%*servers)+1)
		objective := bountyTypes[i%len(bountyTypes)]
		reward := (i + 1) * 100
		difficulty := (i % 3) + 1

		bs.CreateBounty(issuer, target, objective, fmt.Sprintf("%s bounty %d", objective.String(), i+1),
			reward, difficulty, int64(*duration))

		fmt.Printf("Created bounty: %s -> %s, %s, Reward=%d, Difficulty=%d\n",
			issuer, target, objective.String(), reward, difficulty)
	}

	available := bs.GetAvailableBounties()
	fmt.Printf("\nAvailable bounties: %d\n", len(available))

	for i, bounty := range available {
		if i < 3 {
			playerID := fmt.Sprintf("player%d", i+1)
			err := bs.AcceptBounty(bounty.ID, playerID)
			if err != nil {
				fmt.Printf("Error accepting bounty %s: %v\n", bounty.ID, err)
			} else {
				fmt.Printf("Player %s accepted bounty: %s\n", playerID, bounty.Description)
			}
		}
	}

	time.Sleep(2 * time.Second)

	for i, bounty := range available {
		if i < 2 {
			playerID := fmt.Sprintf("player%d", i+1)
			err := bs.CompleteBounty(bounty.ID, playerID)
			if err != nil {
				fmt.Printf("Error completing bounty %s: %v\n", bounty.ID, err)
			} else {
				fmt.Printf("Player %s completed bounty: %s (Reward: %d)\n",
					playerID, bounty.Description, bounty.Reward)
			}
		}
	}

	fmt.Printf("\nCompletion rate: %.1f%%\n", bs.GetCompletionRate()*100)
	fmt.Printf("Active bounties: %d\n", bs.GetActiveBountyCount())

	for i := 1; i <= *servers; i++ {
		serverID := fmt.Sprintf("server%d", i)
		bounties := bs.GetBountiesByServer(serverID)
		fmt.Printf("Server %d bounties: %d\n", i, len(bounties))
	}
}

func testRankingSystem() {
	fmt.Println("--- Ranking System Test ---")

	rm := world.NewRankingManager()

	for i := 1; i <= *servers; i++ {
		serverID := fmt.Sprintf("server%d", i)
		rm.RegisterServer(serverID)

		population := 10 + (i * 5)
		economic := int64((i + 1) * 50000)
		military := i * 3
		diplomatic := (*servers - i) + 1

		rm.UpdatePopulation(serverID, population)
		rm.UpdateEconomicPower(serverID, economic)
		rm.UpdateMilitaryStrength(serverID, military)
		rm.UpdateDiplomaticInfluence(serverID, diplomatic)

		fmt.Printf("Registered %s: Pop=%d, Econ=%d, Mil=%d, Dip=%d\n",
			serverID, population, economic, military, diplomatic)
	}

	leaderboardTypes := []world.LeaderboardType{
		world.LeaderboardPopulation,
		world.LeaderboardEconomic,
		world.LeaderboardMilitary,
		world.LeaderboardDiplomatic,
		world.LeaderboardOverall,
	}

	for _, lbType := range leaderboardTypes {
		fmt.Printf("\n--- %s Leaderboard ---\n", lbType.String())
		leaders := rm.GetLeaderboard(lbType, 3)

		for i, rank := range leaders {
			fmt.Printf("%d. %s (Score: %.0f)\n", i+1, rank.ServerID, rank.TotalScore)
			fmt.Printf("   Pop: %d, Econ: %d, Mil: %d, Dip: %d\n",
				rank.Population, rank.EconomicPower, rank.MilitaryStrength, rank.DiplomaticInfluence)
		}

		if top, ok := rm.GetTopServer(lbType); ok {
			fmt.Printf("Top server: %s\n", top.ServerID)
		}
	}

	fmt.Printf("\nTotal registered servers: %d\n", rm.GetServerCount())
}

func testMetaGameSystem() {
	fmt.Println("--- Meta-Game Event System Test ---")

	em := world.NewEventManager(*seed)

	tournament := em.CreateTournament("Summer Championship", 7200, 4)
	fmt.Printf("Created tournament: %s\n", tournament.Name)
	fmt.Printf("  Duration: %d seconds\n", tournament.EndTime-tournament.StartTime)
	fmt.Printf("  Required servers: %d\n", tournament.RequiredServers)
	fmt.Printf("  Goals: %v\n", tournament.Goals)
	fmt.Printf("  Rewards: %v\n\n", tournament.Rewards)

	serverVsServer := em.CreateServerVsServer("Resource War", "server1", "server2", 3600)
	fmt.Printf("Created server vs server event: %s\n", serverVsServer.Name)
	fmt.Printf("  Participants: %v\n", serverVsServer.Participants)
	fmt.Printf("  Goals: %v\n\n", serverVsServer.Goals)

	worldThreat := em.CreateWorldThreat("Ancient Dragon Awakens", 10800, 5)
	fmt.Printf("Created world threat: %s\n", worldThreat.Name)
	fmt.Printf("  Boss health: %d\n", worldThreat.Goals["boss_damage"])
	fmt.Printf("  Required servers: %d\n", worldThreat.RequiredServers)
	fmt.Printf("  Rewards: %v\n\n", worldThreat.Rewards)

	seasonal := em.CreateSeasonalChallenge("Autumn", 86400)
	fmt.Printf("Created seasonal challenge: %s\n", seasonal.Name)
	fmt.Printf("  Goals: %v\n", seasonal.Goals)
	fmt.Printf("  Rewards: %v\n\n", seasonal.Rewards)

	crisis := em.CreateEconomicCrisis("Market Collapse", 14400)
	fmt.Printf("Created economic crisis: %s\n", crisis.Name)
	fmt.Printf("  Goals: %v\n", crisis.Goals)
	fmt.Printf("  Required servers: %d\n\n", crisis.RequiredServers)

	for i := 1; i <= 3; i++ {
		serverID := fmt.Sprintf("server%d", i)
		em.RegisterParticipant(tournament.ID, serverID)
		em.RegisterParticipant(worldThreat.ID, serverID)
		fmt.Printf("Registered %s for tournament and world threat\n", serverID)
	}

	fmt.Println("\nSimulating progress...")
	for i := 0; i < 5; i++ {
		for j := 1; j <= 3; j++ {
			serverID := fmt.Sprintf("server%d", j)
			em.UpdateProgress(tournament.ID, serverID, "wins", 2)
			em.UpdateProgress(worldThreat.ID, serverID, "damage_dealt", 10000)
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n--- Event Progress ---")
	fmt.Printf("Tournament: %v\n", tournament.Progress)
	fmt.Printf("World Threat: %v\n", worldThreat.Progress)

	completed, _ := em.CheckCompletion(tournament.ID)
	fmt.Printf("\nTournament completed: %v\n", completed)

	completed, _ = em.CheckCompletion(worldThreat.ID)
	fmt.Printf("World threat defeated: %v\n", completed)

	fmt.Printf("\nActive events: %d\n", em.GetActiveEventCount())
	fmt.Printf("Total events: %d\n", em.GetEventCount())

	activeEvents := em.GetActiveEvents()
	fmt.Println("\n--- Active Events Summary ---")
	for _, event := range activeEvents {
		fmt.Printf("%s (%s): %d participants\n", event.Name, event.Type.String(), len(event.Participants))
	}
}
