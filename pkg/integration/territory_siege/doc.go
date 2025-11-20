// Package territory_siege implements siege mechanics for territory control.
//
// This package integrates the existing territory control system with guild warfare,
// providing 3-phase siege mechanics (preparation → assault → resolution), defensive
// structures, reinforcement systems, and loot distribution.
//
// # Integration with Existing Systems
//
// This package connects three V6-V8 systems into cohesive guild warfare:
//
//   - V6 Territory (pkg/world/territory.go): Control points and border zones
//   - V8 Guilds (pkg/network/federation/guild/): Ownership and treasury
//   - V6 Politics (pkg/engine/politics_system.go): Alliances and reinforcements
//
// # Siege Mechanics
//
// Sieges occur in three phases:
//
//  1. Preparation (1 hour): Attacker declares siege, defenders can call reinforcements
//  2. Assault (2 hours): Combat at control points and defensive structures
//  3. Resolution (immediate): Victory conditions checked, loot distributed
//
// # Defensive Structures
//
// Territories can have 5-15 procedurally generated defensive structures:
//
//   - Walls: Block movement, high HP (1000-5000)
//   - Towers: Provide defensive fire support, medium HP (500-1500)
//   - Gates: Entry points, destroyable, medium HP (800-2000)
//   - Barracks: Spawn defender NPCs, low HP (300-800)
//   - Keep: Guild hall defense, very high HP (10000-20000)
//
// # Victory Conditions
//
// Attackers win if they:
//   - Capture all control points (3+ points with 100% progress)
//   - Destroy the guild hall (keep HP reaches 0)
//
// Defenders win if they:
//   - Hold until assault phase expires (2 hours)
//   - Eliminate all attacking players
//
// # Loot Distribution
//
// Victors gain 10-30% of the defender's treasury based on:
//   - Percentage of control points captured
//   - Percentage of defensive structures destroyed
//   - Time taken to achieve victory (faster = higher reward)
//
// # Example Usage
//
//	manager := territory_siege.NewSiegeManager(world)
//
//	// Declare siege on a territory
//	siege, err := manager.DeclareSiege("attackerGuildID", "defenderGuildID", "zoneID")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Call reinforcements (defender allies)
//	err = manager.CallReinforcements(siege.SiegeID, "allyGuildID")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Update siege progress every frame
//	manager.Update(deltaTime)
//
//	// Check siege status
//	if siege.CurrentPhase == territory_siege.PhaseResolution {
//	    log.Printf("Siege complete! Victor: %s", siege.Victor)
//	}
//
// # Performance Targets
//
//   - Siege initialization: <10ms
//   - Structure HP update: <0.1ms per structure
//   - Control point capture: <1ms per point
//   - Loot distribution: <5ms
//
// # Test Coverage Target
//
// ≥65% coverage for all files in this package.
package territory_siege
