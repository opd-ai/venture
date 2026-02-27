// Package faction provides procedural generation of factions for the Venture game world.
//
// # Overview
//
// The faction package generates complete faction systems for game worlds, including:
//   - Faction identities (names, types, descriptions)
//   - Inter-faction relationships (allies, enemies, neutral)
//   - Territory colors for visualization
//   - Member counts and organizational structure
//
// All generation is deterministic based on the world seed, ensuring multiplayer
// synchronization and reproducible worlds.
//
// # Faction Types
//
// Factions come in several types, each with genre-appropriate characteristics:
//   - Kingdom: Monarchies and realms (fantasy)
//   - Guild: Professional organizations and crafters
//   - Cult: Religious or ideological groups (horror, fantasy)
//   - Corporation: Megacorps and businesses (sci-fi, cyberpunk)
//   - Gang: Criminal organizations (horror, cyberpunk, post-apocalyptic)
//   - Rebels: Freedom fighters and resistance movements (sci-fi)
//   - Merchants: Trading companies and commerce groups
//
// # Genre Integration
//
// The generator creates genre-appropriate factions:
//
// Fantasy worlds feature kingdoms, guilds, and merchant leagues with medieval themes.
// Sci-fi worlds have megacorporations, rebel alliances, and trade consortiums.
// Horror worlds emphasize dark cults and violent gangs.
// Cyberpunk worlds balance corporations, street gangs, and underground resistance.
// Post-apocalyptic worlds feature wasteland gangs, survivor groups, and trade caravans.
//
// Usage Example
//
//	gen := faction.NewGenerator()
//	params := procgen.GenerationParams{
//	    Depth:      10,
//	    Difficulty: 0.5,
//	    GenreID:    "fantasy",
//	}
//
//	result, err := gen.Generate(worldSeed, params)
//	if err != nil {
//	    logger.WithError(err).Fatal("faction generation failed")
//	}
//
//	factions := result.([]*engine.Faction)
//	for _, faction := range factions {
//	    logger.WithFields(logrus.Fields{
//	        "faction": faction.Name,
//	        "type":    faction.Type,
//	        "members": faction.MemberCount,
//	    }).Info("generated faction")
//	}
//
// # Faction Relationships
//
// The generator creates relationships between all factions in the world:
//   - Enemy (-100 to -50): At war, attack on sight
//   - Unfriendly (-49 to 0): Distrustful, minimal cooperation
//   - Neutral (1 to 50): Trade only, no military alliance
//   - Allied (51 to 100): Mutual defense, shared objectives
//
// Relationships affect NPC behavior, quest availability, and world dynamics:
//
//	if faction1.IsEnemy(faction2.ID) {
//	    // Factions are at war
//	    // NPCs from these factions will attack each other
//	}
//
//	if faction1.IsAlly(faction2.ID) {
//	    // Factions are allied
//	    // Helping one helps the other
//	}
//
// # Integration with Reputation System
//
// Generated factions integrate with the FactionSystem for reputation tracking:
//
//	factionSystem := engine.NewFactionSystem(world, logger)
//
//	// Add generated factions to system
//	for _, faction := range factions {
//	    factionSystem.AddFaction(faction)
//	}
//
//	// Player reputation affects NPC behavior
//	if factionSystem.ShouldAttackPlayer(factionID) {
//	    // Set NPCs to hostile
//	}
//
// # Deterministic Generation
//
// Same seed with same parameters always produces identical factions:
//
//	factions1, _ := gen.Generate(12345, params)
//	factions2, _ := gen.Generate(12345, params)
//	// factions1 and factions2 are identical
//
// This ensures multiplayer clients see the same world and enables
// reproducible testing and debugging.
//
// # Performance
//
// Faction generation is fast and scales with world depth:
//   - Small worlds (depth 0-10): 3-4 factions, <1ms generation
//   - Medium worlds (depth 11-30): 4-6 factions, <2ms generation
//   - Large worlds (depth 31+): 6-7 factions, <3ms generation
//
// Faction count is capped at 7 to prevent overwhelming complexity.
package faction
