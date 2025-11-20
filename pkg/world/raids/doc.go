// Package raids provides procedural raid dungeon generation and instance management.
//
// Raid dungeons are endgame content designed for groups of 5-10 players, featuring
// multi-boss encounters with procedurally generated mechanics. Each raid tier offers
// increasing difficulty and better loot rewards.
//
// # Raid Tiers
//
// Five difficulty tiers are available:
//
//   - Normal: Entry-level raids for learning mechanics
//   - Heroic: Increased difficulty with better rewards
//   - Mythic: High skill requirement, epic loot
//   - Legendary: Extreme challenge, legendary items
//   - Nightmare: Ultimate difficulty, unique titles
//
// # Instance System
//
// Each raid group receives a separate dungeon instance to prevent griefing.
// Instances persist for the group's session and are cleaned up after completion
// or timeout (4 hours).
//
// # Lockout System
//
// Players can complete each raid tier once per week. Lockouts reset every
// 7 days and are tracked per-player, per-tier.
//
// # Boss Mechanics
//
// Each raid contains 3-5 bosses with procedurally generated abilities:
//
//   - Phase transitions at 75%, 50%, 25% health
//   - Unique mechanics per boss (summons, ground effects, debuffs)
//   - Scaling difficulty based on group size
//   - Deterministic generation for fairness
//
// # Integration
//
// Raids integrate with multiple systems:
//
//   - V2 Terrain: Dungeon layout generation
//   - V2 Entities: Boss stat scaling
//   - V8 Guilds: Group coordination and lockout tracking
//   - V9 Economy: Epic/Legendary loot distribution
//
// # Usage Example
//
//	gen := raids.NewRaidGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.8,
//	    Depth:      15,
//	    GenreID:    "fantasy",
//	    Custom: map[string]interface{}{
//	        "group_size": 8,
//	    },
//	}
//
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	raid := result.(*raids.RaidDungeon)
//	fmt.Printf("Raid: %s (%d bosses)\n", raid.Name, len(raid.Bosses))
//
// # Performance
//
// Generation targets:
//   - Generation time: <5s per raid dungeon
//   - Memory usage: <50MB per instance
//   - Boss count: 3-5 per raid
//   - Room count: 10-20 per raid
package raids
