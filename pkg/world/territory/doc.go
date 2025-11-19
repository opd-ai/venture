// Package territory provides guild territory control and warfare mechanics.
//
// This package implements a comprehensive territory control system for guild warfare,
// including territory ownership, capture mechanics, defensive structures, and benefits.
//
// Key Features:
//   - Territory zones represented as 5×5 chunk areas on the world map
//   - Guild ownership assignment with capture mechanics
//   - Territory benefits: +10% resource spawn rate, +5% XP gain in controlled zones
//   - Guild war declaration system with formal conflict mechanics
//   - Defensive structures: Walls, towers, NPC guards to protect territories
//   - Capture progress tracking with attacker/defender mechanics
//   - Cross-server territory synchronization support
//
// Territory System:
//
// Territories are defined as 5×5 chunk zones (160×160 tiles at 32 tiles per chunk).
// Each territory can be:
//   - Neutral (unowned)
//   - Owned by a specific guild
//   - Contested (currently being captured)
//
// Capture Mechanics:
//
// Territories are captured through control points. Guild members must be present
// at control points to increase capture progress. The formula is:
//   - Progress increases when attackers > defenders
//   - Progress rate: 1.0 per 60 seconds base time
//   - Defenders slow capture: +30 seconds per defender
//   - Full capture at 100% progress
//
// Defensive Structures:
//
// Guilds can construct defensive structures in owned territories:
//   - Walls: Provide physical barriers, 1000 HP base
//   - Towers: Attack enemies, 100 damage/shot, 500 HP
//   - Guards: NPC defenders, level 30, 500 HP
//
// Territory Benefits:
//
// Controlling territories provides passive bonuses:
//   - +10% resource spawn rate in controlled zones
//   - +5% XP gain for guild members in controlled zones
//   - Strategic location control for guild warfare
//
// War Declaration:
//
// Guilds can formally declare war on other guilds:
//   - Declaration costs 1000 gold (prevents griefing)
//   - War lasts for 7 days or until manually ended
//   - During war, territories can be captured
//   - Peace declaration costs 500 gold from both sides
//
// Example Usage:
//
//	// Create territory manager
//	tm := territory.NewManager()
//
//	// Create a new territory zone
//	coords := territory.TerritoryCoords{ChunkX: 10, ChunkZ: 10}
//	terr, err := tm.CreateTerritory("territory-1", coords)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Assign territory to a guild
//	err = tm.AssignOwner("territory-1", "guild-123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Declare war between guilds
//	war, err := tm.DeclareWar("guild-123", "guild-456")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Update capture progress
//	err = tm.UpdateCaptureProgress("territory-1", 5, 2, "guild-456")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Build defensive structure
//	structure, err := tm.BuildDefensiveStructure("territory-1", territory.StructureTypeWall, 100.0, 100.0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get territory benefits
//	resourceBonus := tm.GetResourceBonus("guild-123")
//	xpBonus := tm.GetXPBonus("guild-123")
//
// Thread Safety:
//
// All operations are thread-safe using sync.RWMutex for concurrent access.
// The manager can be safely used from multiple goroutines.
//
// Performance:
//
// Target metrics for Phase 50.2:
//   - Territory load: <50ms for 100 territories
//   - Capture progress update: <10ms per tick
//   - Structure creation: <20ms per structure
//   - Benefits calculation: <5ms per guild
package territory
