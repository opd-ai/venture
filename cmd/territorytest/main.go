// Territory test demonstrates guild territory control and warfare mechanics.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/world/territory"
)

func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	fmt.Println("=== Territory Control & Guild Warfare Demo ===")
	fmt.Println()

	tm := territory.NewManager()
	fmt.Println("✓ Created territory manager")

	createTerritories(tm)
	assignTerritoryOwnership(tm)
	displayGuildBenefits(tm)
	wall, tower := buildDefensiveStructures(tm)
	declareGuildWar(tm)
	simulateTerritoryCapture(tm)
	damageDefensiveStructures(tm, wall.ID, tower.ID)
	displayWarSummary(tm)
	displayFinalSummary(tm, *verbose)

	fmt.Println("\n✓ Territory control demo complete!")
}

// createTerritories creates three test territories at different chunk coordinates.
func createTerritories(tm *territory.Manager) {
	fmt.Println("\n--- Creating Territories ---")
	coords1 := territory.TerritoryCoords{ChunkX: 10, ChunkZ: 10}
	coords2 := territory.TerritoryCoords{ChunkX: 20, ChunkZ: 20}
	coords3 := territory.TerritoryCoords{ChunkX: 30, ChunkZ: 30}

	terr1, err := tm.CreateTerritory("territory-north", coords1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created territory: %s at chunk (%d, %d) - Status: %s\n",
		terr1.ID, terr1.Coords.ChunkX, terr1.Coords.ChunkZ, terr1.Status)

	terr2, err := tm.CreateTerritory("territory-south", coords2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created territory: %s at chunk (%d, %d) - Status: %s\n",
		terr2.ID, terr2.Coords.ChunkX, terr2.Coords.ChunkZ, terr2.Status)

	terr3, err := tm.CreateTerritory("territory-west", coords3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created territory: %s at chunk (%d, %d) - Status: %s\n",
		terr3.ID, terr3.Coords.ChunkX, terr3.Coords.ChunkZ, terr3.Status)
}

// assignTerritoryOwnership assigns territories to guilds and displays ownership counts.
func assignTerritoryOwnership(tm *territory.Manager) {
	fmt.Println("\n--- Assigning Territory Ownership ---")
	err := tm.AssignOwner("territory-north", "guild-dragons")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Assigned territory-north to guild-dragons")

	err = tm.AssignOwner("territory-south", "guild-knights")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Assigned territory-south to guild-knights")

	fmt.Println("\n--- Guild Territory Ownership ---")
	dragonsTerritory := tm.GetGuildTerritories("guild-dragons")
	fmt.Printf("Guild Dragons controls %d territory(ies)\n", len(dragonsTerritory))

	knightsTerritory := tm.GetGuildTerritories("guild-knights")
	fmt.Printf("Guild Knights controls %d territory(ies)\n", len(knightsTerritory))
}

// displayGuildBenefits shows resource and XP bonuses for each guild.
func displayGuildBenefits(tm *territory.Manager) {
	fmt.Println("\n--- Territory Benefits ---")
	dragonResourceBonus := tm.GetResourceBonus("guild-dragons")
	dragonXPBonus := tm.GetXPBonus("guild-dragons")
	fmt.Printf("Guild Dragons bonuses: +%.1f%% resource spawn, +%.1f%% XP gain\n",
		dragonResourceBonus*100, dragonXPBonus*100)

	knightsResourceBonus := tm.GetResourceBonus("guild-knights")
	knightsXPBonus := tm.GetXPBonus("guild-knights")
	fmt.Printf("Guild Knights bonuses: +%.1f%% resource spawn, +%.1f%% XP gain\n",
		knightsResourceBonus*100, knightsXPBonus*100)
}

// buildDefensiveStructures constructs walls, towers, and guards in territories and returns wall and tower structures.
func buildDefensiveStructures(tm *territory.Manager) (*territory.DefensiveStructure, *territory.DefensiveStructure) {
	fmt.Println("\n--- Building Defensive Structures ---")
	wall, err := tm.BuildDefensiveStructure("territory-north", territory.StructureTypeWall, 100.0, 100.0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Built %s at (%.1f, %.1f) - HP: %.0f/%.0f\n",
		wall.Type, wall.X, wall.Y, wall.HP, wall.MaxHP)

	tower, err := tm.BuildDefensiveStructure("territory-north", territory.StructureTypeTower, 200.0, 200.0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Built %s at (%.1f, %.1f) - HP: %.0f/%.0f, Damage: %.0f\n",
		tower.Type, tower.X, tower.Y, tower.HP, tower.MaxHP, tower.Damage)

	guard, err := tm.BuildDefensiveStructure("territory-south", territory.StructureTypeGuard, 300.0, 300.0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Built %s at (%.1f, %.1f) - HP: %.0f/%.0f, Level: %d\n",
		guard.Type, guard.X, guard.Y, guard.HP, guard.MaxHP, guard.Level)

	return wall, tower
}

// declareGuildWar initiates war between two guilds and displays war information.
func declareGuildWar(tm *territory.Manager) {
	fmt.Println("\n--- Declaring War ---")
	war, err := tm.DeclareWar("guild-knights", "guild-dragons")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ War declared: %s vs %s\n", war.AttackerGuild, war.DefenderGuild)
	fmt.Printf("  Declared at: %s\n", war.DeclaredAt.Format(time.RFC3339))
	fmt.Printf("  Ends at: %s\n", war.EndsAt.Format(time.RFC3339))
	fmt.Printf("  Cost: %d gold\n", war.Cost)

	isAtWar := tm.IsAtWar("guild-knights", "guild-dragons")
	fmt.Printf("  Are guilds at war? %v\n", isAtWar)
}

// simulateTerritoryCapture runs multiple capture attempts and shows contested territories.
func simulateTerritoryCapture(tm *territory.Manager) {
	fmt.Println("\n--- Simulating Territory Capture ---")
	terr1, _ := tm.GetTerritory("territory-north")
	terr1.LastUpdate = time.Now().Add(-10 * time.Second)

	err := tm.UpdateCaptureProgress("territory-north", 5, 2, "guild-knights")
	if err != nil {
		log.Fatal(err)
	}
	terr1, _ = tm.GetTerritory("territory-north")
	fmt.Printf("Capture attempt 1: Progress %.2f%%, Status: %s\n",
		terr1.CaptureProgress*100, terr1.Status)

	terr1.LastUpdate = time.Now().Add(-20 * time.Second)
	err = tm.UpdateCaptureProgress("territory-north", 8, 1, "guild-knights")
	if err != nil {
		log.Fatal(err)
	}
	terr1, _ = tm.GetTerritory("territory-north")
	fmt.Printf("Capture attempt 2: Progress %.2f%%, Status: %s\n",
		terr1.CaptureProgress*100, terr1.Status)

	terr1.LastUpdate = time.Now().Add(-30 * time.Second)
	err = tm.UpdateCaptureProgress("territory-north", 10, 0, "guild-knights")
	if err != nil {
		log.Fatal(err)
	}
	terr1, _ = tm.GetTerritory("territory-north")
	fmt.Printf("Capture attempt 3: Progress %.2f%%, Status: %s\n",
		terr1.CaptureProgress*100, terr1.Status)

	if terr1.CaptureProgress >= 1.0 {
		fmt.Printf("✓ Territory captured! New owner: %s\n", terr1.OwnerGuildID)
	}

	fmt.Println("\n--- Contested Territories ---")
	contested := tm.GetContestedTerritories()
	fmt.Printf("Currently contested territories: %d\n", len(contested))
	for _, t := range contested {
		fmt.Printf("  - %s: %.1f%% captured by %s\n",
			t.ID, t.CaptureProgress*100, t.CapturingGuild)
	}
}

// damageDefensiveStructures applies damage to wall and tower structures.
func damageDefensiveStructures(tm *territory.Manager, wallID, towerID string) {
	fmt.Println("\n--- Damaging Defensive Structures ---")
	err := tm.DamageStructure("territory-north", wallID, 300.0)
	if err != nil {
		log.Fatal(err)
	}
	terr1, _ := tm.GetTerritory("territory-north")
	for _, s := range terr1.Structures {
		if s.ID == wallID {
			fmt.Printf("Wall damaged: %.0f/%.0f HP remaining\n", s.HP, s.MaxHP)
		}
	}

	err = tm.DamageStructure("territory-north", towerID, 600.0)
	if err != nil {
		log.Fatal(err)
	}
	terr1, _ = tm.GetTerritory("territory-north")
	towerDestroyed := true
	for _, s := range terr1.Structures {
		if s.ID == towerID {
			towerDestroyed = false
		}
	}
	if towerDestroyed {
		fmt.Println("✓ Tower destroyed!")
	}
}

// displayWarSummary shows active wars with their details.
func displayWarSummary(tm *territory.Manager) {
	fmt.Println("\n--- Active Wars ---")
	activeWars := tm.GetActiveWars()
	fmt.Printf("Active wars: %d\n", len(activeWars))
	for _, w := range activeWars {
		fmt.Printf("  - %s vs %s (declared %s ago)\n",
			w.AttackerGuild, w.DefenderGuild, time.Since(w.DeclaredAt).Round(time.Second))
	}
}

// displayFinalSummary shows territory ownership and optionally verbose statistics.
func displayFinalSummary(tm *territory.Manager, verbose bool) {
	fmt.Println("\n--- Final Territory Summary ---")
	allTerritories := tm.GetAllTerritories()
	for _, t := range allTerritories {
		fmt.Printf("%s:\n", t.ID)
		fmt.Printf("  Owner: %s\n", t.OwnerGuildID)
		fmt.Printf("  Status: %s\n", t.Status)
		fmt.Printf("  Structures: %d\n", len(t.Structures))
		if t.CapturingGuild != "" {
			fmt.Printf("  Being captured by: %s (%.1f%%)\n", t.CapturingGuild, t.CaptureProgress*100)
		}
	}

	if verbose {
		displayVerboseStats(tm, allTerritories)
	}
}

// displayVerboseStats shows detailed statistics about territories and guilds.
func displayVerboseStats(tm *territory.Manager, allTerritories []*territory.Territory) {
	fmt.Println("\n--- Verbose Stats ---")
	fmt.Printf("Total territories: %d\n", len(allTerritories))

	activeWars := tm.GetActiveWars()
	fmt.Printf("Total active wars: %d\n", len(activeWars))

	contested := tm.GetContestedTerritories()
	fmt.Printf("Total contested territories: %d\n", len(contested))

	knightsUpdatedTerritories := tm.GetGuildTerritories("guild-knights")
	fmt.Printf("Guild Knights now controls: %d territory(ies)\n", len(knightsUpdatedTerritories))

	dragonsUpdatedTerritories := tm.GetGuildTerritories("guild-dragons")
	fmt.Printf("Guild Dragons now controls: %d territory(ies)\n", len(dragonsUpdatedTerritories))
}
