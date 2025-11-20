package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/integration/world_events"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, faction, economic, weather, propagate, chain, all")
	seed := flag.Int64("seed", 12345, "Random seed")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	fmt.Printf("World Events Test Tool - Phase 58.2\n")
	fmt.Printf("Mode: %s | Seed: %d\n\n", *mode, *seed)

	switch *mode {
	case "demo":
		runDemoMode(*seed, *verbose)
	case "faction":
		runFactionMode(*seed, *verbose)
	case "economic":
		runEconomicMode(*seed, *verbose)
	case "weather":
		runWeatherMode(*seed, *verbose)
	case "propagate":
		runPropagateMode(*seed, *verbose)
	case "chain":
		runChainMode(*seed, *verbose)
	case "all":
		runAllModes(*seed, *verbose)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func runDemoMode(seed int64, verbose bool) {
	fmt.Println("=== Demo Mode: Basic Event Generation ===\n")

	manager := world_events.NewEventManager(seed)

	triggers := []struct {
		trigger world_events.TriggerType
		params  world_events.TriggerParams
	}{
		{
			trigger: world_events.TriggerGuildWar,
			params: world_events.TriggerParams{
				TriggerType: world_events.TriggerGuildWar,
				Severity:    world_events.SeverityMajor,
				Location:    "northern_border",
				ServerID:    "server_1",
				GuildID:     "guild_crimson",
			},
		},
		{
			trigger: world_events.TriggerTradeVolume,
			params: world_events.TriggerParams{
				TriggerType: world_events.TriggerTradeVolume,
				Severity:    world_events.SeverityModerate,
				Location:    "market_square",
				ServerID:    "server_1",
				ItemType:    "iron_ore",
			},
		},
		{
			trigger: world_events.TriggerWeatherChange,
			params: world_events.TriggerParams{
				TriggerType: world_events.TriggerWeatherChange,
				Severity:    world_events.SeverityCritical,
				Location:    "coastal_region",
				ServerID:    "server_1",
			},
		},
	}

	for i, t := range triggers {
		fmt.Printf("Event %d: %s\n", i+1, t.trigger)
		event, err := manager.GenerateEvent(t.trigger, t.params)
		if err != nil {
			log.Fatalf("Failed to generate event: %v", err)
		}

		displayEvent(event, verbose)
		fmt.Println()
	}

	stats := manager.GetStats()
	displayStats(stats)
}

func runFactionMode(seed int64, verbose bool) {
	fmt.Println("=== Faction Response Mode ===\n")

	factions := []struct {
		id       string
		action   string
		severity world_events.Severity
	}{
		{"faction_merchants", "guild_war", world_events.SeverityMajor},
		{"faction_guards", "market_crash", world_events.SeverityModerate},
		{"faction_nobles", "invasion", world_events.SeverityCritical},
	}

	for i, f := range factions {
		fmt.Printf("Faction %d: %s responds to %s\n", i+1, f.id, f.action)
		response := world_events.GenerateFactionResponse(seed+int64(i), f.id, f.action, f.severity)

		fmt.Printf("  Response Type: %s\n", response.ResponseType)
		fmt.Printf("  Reputation Change: %+.2f\n", response.ReputationChange)
		fmt.Printf("  Hostility Change: %+.2f\n", response.HostilityChange)
		fmt.Printf("  Trade Bonus: %+.2f\n", response.TradeBonus)
		fmt.Printf("  Message: %s\n", response.Message)
		fmt.Println()
	}
}

func runEconomicMode(seed int64, verbose bool) {
	fmt.Println("=== Economic Events Mode ===\n")

	items := []struct {
		eventID  string
		itemType string
		severity world_events.Severity
	}{
		{"econ_1", "iron_ore", world_events.SeverityMinor},
		{"econ_2", "gold", world_events.SeverityMajor},
		{"econ_3", "food", world_events.SeverityCritical},
	}

	for i, item := range items {
		fmt.Printf("Economic Event %d: %s\n", i+1, item.itemType)
		event := world_events.GenerateEconomicEvent(seed+int64(i), item.eventID, item.itemType, item.severity)

		fmt.Printf("  Event ID: %s\n", event.EventID)
		fmt.Printf("  Item Type: %s\n", event.ItemType)
		fmt.Printf("  Price Modifier: %+.1f%%\n", event.PriceModifier*100)
		fmt.Printf("  Supply Change: %+d units\n", event.SupplyChange)
		fmt.Printf("  Duration: %v\n", event.Duration)
		fmt.Printf("  Affected Zones: %d\n", len(event.AffectedZones))
		if verbose {
			for _, zone := range event.AffectedZones {
				fmt.Printf("    - %s\n", zone)
			}
		}
		fmt.Println()
	}
}

func runWeatherMode(seed int64, verbose bool) {
	fmt.Println("=== Weather Disasters Mode ===\n")

	disasters := []struct {
		centerX  float64
		centerY  float64
		severity world_events.Severity
	}{
		{100.0, 100.0, world_events.SeverityMinor},
		{200.0, 200.0, world_events.SeverityMajor},
		{300.0, 300.0, world_events.SeverityCritical},
	}

	for i, d := range disasters {
		fmt.Printf("Weather Disaster %d: Severity %d\n", i+1, d.severity)
		disaster := world_events.GenerateWeatherDisaster(seed+int64(i), d.centerX, d.centerY, d.severity)

		fmt.Printf("  Type: %s\n", disaster.DisasterType)
		fmt.Printf("  Intensity: %.2f\n", disaster.Intensity)
		fmt.Printf("  Radius: %.1f tiles\n", disaster.Radius)
		fmt.Printf("  Center: (%.1f, %.1f)\n", disaster.CenterX, disaster.CenterY)
		fmt.Printf("  Duration: %v\n", disaster.Duration)
		fmt.Printf("  Damage: %.1f/sec\n", disaster.Damage)
		fmt.Println()
	}
}

func runPropagateMode(seed int64, verbose bool) {
	fmt.Println("=== Cross-Server Propagation Mode ===\n")

	manager := world_events.NewEventManager(seed)
	params := world_events.TriggerParams{
		TriggerType: world_events.TriggerGuildWar,
		Severity:    world_events.SeverityMajor,
		Location:    "guild_hall",
		ServerID:    "server_1",
		GuildID:     "guild_alpha",
	}

	event, err := manager.GenerateEvent(world_events.TriggerGuildWar, params)
	if err != nil {
		log.Fatalf("Failed to generate event: %v", err)
	}

	fmt.Println("Original Event:")
	displayEvent(event, verbose)
	fmt.Println()

	targetServers := []string{"server_2", "server_3", "server_4"}
	delay := 30 * time.Second

	propagated := world_events.PropagateEventCrossServer(event, targetServers, delay)

	fmt.Printf("Propagated to %d servers:\n\n", len(propagated))
	for i, prop := range propagated {
		fmt.Printf("Propagated Event %d (Server %s):\n", i+1, prop.ServerID)
		displayEvent(prop, verbose)
		fmt.Println()
	}
}

func runChainMode(seed int64, verbose bool) {
	fmt.Println("=== Event Chain Mode ===\n")

	config := world_events.EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     1.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      1 * time.Second,
		ResponseTimeMax:      2 * time.Second,
	}
	manager := world_events.NewEventManagerWithConfig(seed, config)

	params := world_events.TriggerParams{
		TriggerType: world_events.TriggerGuildWar,
		Severity:    world_events.SeverityMajor,
		Location:    "fortress",
		ServerID:    "server_1",
		GuildID:     "guild_beta",
		PlayerID:    "player_1",
	}

	event, err := manager.GenerateEvent(world_events.TriggerGuildWar, params)
	if err != nil {
		log.Fatalf("Failed to generate event: %v", err)
	}

	fmt.Println("Initial Event:")
	displayEvent(event, verbose)
	fmt.Println()

	chain := manager.GetEventChain(event.ID)
	fmt.Printf("Event Chain (%d events):\n\n", len(chain))

	for i, eventID := range chain {
		chainEvent, ok := manager.GetEvent(eventID)
		if !ok {
			fmt.Printf("  %d. %s (not found)\n", i+1, eventID)
			continue
		}
		fmt.Printf("Chain Event %d:\n", i+1)
		displayEvent(chainEvent, verbose)
		fmt.Println()
	}
}

func runAllModes(seed int64, verbose bool) {
	modes := []struct {
		name string
		fn   func(int64, bool)
	}{
		{"Demo", runDemoMode},
		{"Faction", runFactionMode},
		{"Economic", runEconomicMode},
		{"Weather", runWeatherMode},
		{"Propagate", runPropagateMode},
		{"Chain", runChainMode},
	}

	for _, m := range modes {
		fmt.Println("\n" + strings.Repeat("═", 60))
		m.fn(seed, verbose)
	}
	fmt.Println(strings.Repeat("═", 60))
}

func displayEvent(event *world_events.WorldEvent, verbose bool) {
	fmt.Printf("  ID: %s\n", event.ID)
	fmt.Printf("  Type: %s | Trigger: %s | Severity: %d\n", event.Type, event.Trigger, event.Severity)
	fmt.Printf("  Title: %s\n", event.Title)
	fmt.Printf("  Description: %s\n", event.Description)
	fmt.Printf("  Location: %s | Server: %s\n", event.Location, event.ServerID)
	fmt.Printf("  Start: %s | Duration: %v\n", event.StartTime.Format("15:04:05"), event.Duration)
	fmt.Printf("  Impacts: %d | Chain Events: %d | Permanent: %v\n", len(event.Impacts), len(event.ChainEvents), event.Permanent)

	if verbose && len(event.Impacts) > 0 {
		fmt.Println("  Detailed Impacts:")
		for i, impact := range event.Impacts {
			fmt.Printf("    %d. Type: %s | Target: %s | Modifier: %+.2f | Duration: %v\n",
				i+1, impact.Type, impact.Target, impact.Modifier, impact.Duration)
		}
	}
}

func displayStats(stats map[string]interface{}) {
	fmt.Println("=== Manager Statistics ===")
	fmt.Printf("Active Events: %d\n", stats["active_events"])
	fmt.Printf("Event Chains: %d\n", stats["event_chains"])
	fmt.Printf("Total Generated: %d\n", stats["total_generated"])

	if typeCounts, ok := stats["type_counts"].(map[world_events.EventType]int); ok {
		fmt.Println("Type Breakdown:")
		for eventType, count := range typeCounts {
			fmt.Printf("  %s: %d\n", eventType, count)
		}
	}
}
