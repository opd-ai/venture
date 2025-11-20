// Package advanced provides multi-classing, prestige classes, and talent tree systems
// for deep character customization in the Venture action-RPG.
//
// # Overview
//
// The advanced class system extends basic character progression with three major features:
//
//  1. Multi-Classing: Players can combine a primary and secondary class for hybrid builds
//  2. Prestige Classes: Advanced specializations unlocked at level 20+ with specific requirements
//  3. Talent Trees: Deep customization with 30 talents per class across 3 categories
//
// # Basic Usage
//
//	manager := advanced.NewManager()
//
//	// Set primary class
//	manager.SetPrimaryClass("player123", advanced.ClassWarrior)
//	manager.SetLevel("player123", 10)
//
//	// Add secondary class for multi-classing
//	manager.SetSecondaryClass("player123", advanced.ClassMage)
//
//	// Allocate talent points
//	manager.AllocateTalent("player123", "warrior_power_strike")
//
//	// Calculate total stats
//	stats, _ := manager.CalculateTotalStats("player123")
//
// # Classes
//
// The system includes 15 base classes across 4 categories:
//
//   - Warrior: Warrior, Berserker, Paladin, Knight
//   - Rogue: Rogue, Assassin, Ranger, Ninja
//   - Mage: Mage, Elementalist, Necromancer, Enchanter
//   - Support: Cleric, Bard, Druid
//
// # Prestige Classes
//
// 20 prestige classes available at level 20+ with specific requirements:
//
//   - Warrior: Blade Master, Champion, Dragon Knight, Dreadnought
//   - Rogue: Shadow Dancer, Deadeye, Duelist, Phantom
//   - Mage: Archmage, Soul Reaper, Elemental Lord, Time Mage
//   - Support: High Priest, Maestro, Archdruid, Oracle
//   - Hybrid: Battlemage, Spellblade, Voidwalker, Runesmith
//
// Example prestige class unlock:
//
//	manager.SetLevel("player123", 20)
//	manager.SetPrestigeClass("player123", advanced.PrestigeBladeMaster)
//
// # Talent Trees
//
// Each base class has 30 talents organized in 3 categories:
//
//   - Offensive: 10 talents for damage and combat effectiveness
//   - Defensive: 10 talents for survivability and damage mitigation
//   - Utility: 10 talents for non-combat abilities and party support
//
// Talents have ranks (1-5), prerequisites, and provide stat bonuses.
//
// Example talent allocation:
//
//	// Allocate 5 points in Power Strike
//	for i := 0; i < 5; i++ {
//	    manager.AllocateTalent("player123", "warrior_power_strike")
//	}
//
//	// Unlock Weapon Mastery (requires Power Strike)
//	manager.AllocateTalent("player123", "warrior_weapon_mastery")
//
// # Multi-Classing Synergies
//
// Compatible class combinations provide synergy bonuses:
//
//   - Warrior + Mage = Spellsword: +5 STR, +5 INT, +20 Mana
//   - Rogue + Mage = Trickster: +5 DEX, +5 INT, +5% Crit
//   - Paladin + Cleric = Divine Protector: +50 HP, +8 WIS, +5 DEF
//
// Secondary class stats are scaled to 50% effectiveness.
//
// # Talent Respec
//
// Players can reset talents for a gold cost:
//
//	cost := manager.GetRespecCost("player123")
//	manager.RespecTalents("player123", goldAmount)
//
// Respec cost increases with each use: Base 1000g + 500g per respec, max 10000g.
//
// # ECS Integration
//
// The AdvancedClassComponent integrates with the Venture ECS:
//
//	component := advanced.AdvancedClassComponent{
//	    PrimaryClass:   advanced.ClassWarrior,
//	    SecondaryClass: advanced.ClassMage,
//	    PrestigeClass:  advanced.PrestigeBattlemage,
//	    Level:          25,
//	    TalentPoints: advanced.TalentAllocation{
//	        Talents:      make(map[advanced.TalentID]int),
//	        PointsSpent:  0,
//	        PointsTotal:  25,
//	    },
//	}
//
// # Performance
//
// All operations are designed for low latency:
//
//   - Class assignment: <1µs
//   - Talent allocation: <10ms with validation
//   - Stat calculation: <5ms combining all sources
//   - Respec operation: <100ms
//
// The Manager uses read-write locks for thread-safe concurrent access.
package advanced
