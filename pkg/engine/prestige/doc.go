// Package prestige provides post-max-level progression systems.
//
// The prestige system allows players to continue advancing after reaching the
// standard level cap (level 50). Players earn prestige levels infinitely with
// exponentially increasing XP requirements, gaining paragon points and unlocking
// special prestige abilities.
//
// # Prestige Levels
//
// Prestige levels start at 1 after reaching max level (50):
//   - Level 50 → Prestige 1 (first prestige level)
//   - Infinite progression (no cap on prestige levels)
//   - Exponential XP curve: XP required = baseXP * (2 ^ prestigeLevel)
//   - Visual effects at milestones (glowing auras at 10, 25, 50, 100)
//
// # Paragon Points
//
// Each prestige level grants 1 paragon point to spend on stat bonuses:
//   - +0.1% stat increase per point (multiplicative)
//   - Allocate points across: Health, Damage, Defense, Speed, Critical
//   - No point cap (unlimited allocation)
//   - Respec available for gold cost (1000g per point reallocated)
//
// # Prestige Abilities
//
// Unlock powerful class-specific abilities at prestige milestones:
//   - Prestige 10: First prestige ability (e.g., "Veteran's Resolve")
//   - Prestige 25: Second prestige ability (e.g., "Legendary Strike")
//   - Prestige 50: Third prestige ability (e.g., "Ascended Power")
//   - Prestige 100: Fourth prestige ability (e.g., "Transcendent Form")
//
// Each class has 4 unique prestige abilities (15 classes × 4 = 60 total).
//
// # Account-Wide Bonuses
//
// Prestige progress provides account-wide benefits:
//   - Each prestige 100 character: +5% XP gain for all characters
//   - Stacks multiplicatively (2 prestige 100 chars = +10.25% XP)
//   - Shared across all characters on account
//   - Persists across character deletion
//
// # Visual Effects
//
// Prestige levels grant visual indicators:
//   - Prestige 1-9: No visual effect
//   - Prestige 10-24: Subtle glow around character
//   - Prestige 25-49: Brighter glow with particle trail
//   - Prestige 50-99: Intense glow with aura
//   - Prestige 100+: Radiant aura with unique color per class
//
// # Usage Example
//
//	// Create prestige manager
//	mgr := prestige.NewManager()
//
//	// Add prestige XP for player
//	playerID := "player123"
//	xpGained := 1000
//	levelsGained := mgr.AddPrestigeXP(playerID, "Warrior", xpGained)
//	if levelsGained > 0 {
//		// Award paragon points
//		mgr.AddParagonPoints(playerID, levelsGained)
//	}
//
//	// Allocate paragon point
//	err := mgr.AllocateParagonPoint(playerID, prestige.StatHealth)
//
//	// Check for prestige ability unlock
//	level := mgr.GetPrestigeLevel(playerID)
//	if level == 10 {
//		ability := mgr.GetPrestigeAbility("Warrior", 10)
//		// Grant ability to player
//	}
//
//	// Calculate account-wide XP bonus
//	bonus := mgr.GetAccountXPBonus(accountID)
//	adjustedXP := baseXP * (1.0 + bonus)
//
// # Integration
//
// Prestige integrates with existing Venture systems:
//   - V2 Progression: Extends progression_system.go beyond level 50
//   - V4 Classes: Class-specific prestige abilities (4 per class)
//   - V8 Advanced Classes: Prestige abilities synergize with talents
//   - V8 Account System: Account-wide bonuses persist across characters
//
// # Performance
//
// Target metrics:
//   - XP addition: <1ms per player
//   - Point allocation: <5ms per point
//   - Ability unlock check: <0.1ms
//   - Account bonus calculation: <10ms per account
//
// # Testing
//
// The package includes comprehensive tests:
//   - XP curve validation (exponential growth)
//   - Paragon point stat calculations (+0.1% per point)
//   - Ability unlock at correct milestones (10, 25, 50, 100)
//   - Account bonus stacking (multiple prestige 100 chars)
//   - Respec cost calculation (1000g per point)
package prestige
