// Package political_warfare integrates V6 Politics, V8 Guilds, and V6 Federation Market
// for guild-level political warfare mechanics including alliances, embargoes, war declarations,
// peace treaties, and diplomatic victories.
//
// Phase 56.3: Political Warfare Integration (ROADMAP_V9.md)
//
// Key Features:
//   - Political alliances affect siege reinforcements (60-80% call success)
//   - Trade embargoes block enemy guild market access (50-90% price increases)
//   - Diplomatic victories allow surrender without combat
//   - War declarations with 24-hour preparation periods
//   - Peace treaties with 7-14 day cooldown periods
//   - Reputation impacts from aggressive actions (-0.1 to -0.5 per action)
//
// Integration Dependencies:
//   - pkg/engine/politics_system.go (alliances, wars, treaties)
//   - pkg/network/federation/guild (guild relations)
//   - pkg/network/federation/market.go (embargoes)
//
// Example Usage:
//
//	// Create political warfare manager with default seed
//	manager := political_warfare.NewManager(world, guildManager)
//
//	// Or create with world seed for deterministic behavior
//	manager := political_warfare.NewManagerWithSeed(world, guildManager, worldSeed)
//
//	// Declare war with preparation period
//	war, err := manager.DeclareWar("guild1", "guild2", 24*time.Hour)
//
//	// Call allies for siege reinforcements
//	allies, err := manager.CallReinforcementAllies("guild1", "guild2")
//
//	// Impose trade embargo on enemy guild (0.5 = 50% to 0.9 = 90% price markup)
//	embargo, err := manager.ImposeEmbargo("guild1", "guild2", 0.9)
//
//	// Attempt diplomatic victory
//	success, err := manager.NegotiateDiplomaticVictory("guild1", "guild2", concessions)
//
//	// Sign peace treaty with cooldown
//	treaty, err := manager.SignPeaceTreaty("guild1", "guild2", 14*24*time.Hour)
//
//	// Apply reputation penalty for aggressive actions
//	manager.ApplyReputationPenalty("guild1", "attack", -0.3)
//
// Performance:
//   - War declaration: <1ms
//   - Alliance check: <100ns
//   - Embargo application: <1ms
//   - Reputation update: <1ms
//
// Thread Safety:
// All public methods are thread-safe using sync.RWMutex.
package political_warfare
