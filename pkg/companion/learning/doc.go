/*
Package learning implements companion AI skill progression, personality evolution, and behavioral memory.

The companion learning system enables AI companions to learn from player interactions, develop
skills through experience, evolve their personality traits over time, and remember significant
events. This creates emergent companion behavior that adapts to the player's playstyle.

# Core Features

1. Skill Progression: 24 learnable skills across 8 categories (Combat, Defense, Utility, Social, Healing, Magic, Crafting, Stealth)
2. Personality Evolution: 10 personality traits that shift based on companion actions and player interactions
3. Event Memory: LRU-based memory system storing up to 1000 memorable events per companion
4. Behavioral Adaptation: Companions learn and mirror player combat styles and decision patterns

# Skill System

Skills are organized in a tree structure with prerequisites. Companions gain experience (XP) through
actions and level up skills from 0 to 10. Each level provides a 10% bonus to related actions.
Skill points are earned through leveling and spent to unlock new skills.

Example skill chains:
  - Combat: Basic Attack → Power Strike → Combat Mastery
  - Defense: Block → Iron Skin → Defensive Stance
  - Social: Persuasion → Charm → Leadership

# Personality System

Personality traits are represented as values from 0.0 to 1.0 and evolve based on companion behavior:
  - Cautious ↔ Brave
  - Shy ↔ Outgoing
  - Aggressive ↔ Pacifist
  - Loyal ↔ Independent
  - Curious ↔ Practical

Opposing traits are balanced: increasing Brave decreases Cautious proportionally.

# Memory System

Companions remember up to 1000 significant events with importance scores (0.0-1.0). Events are
categorized by type (Combat, Dialog, Trade, Quest, etc.) and can be filtered or summarized.
Older events are evicted using LRU when the limit is reached.

# Usage Example

	// Create a companion learning system
	system := learning.NewCompanionLearningSystem(time.Second)
	manager := system.GetManager()

	// Add a companion with 1.2x learning rate
	comp := manager.AddCompanion("companion_001", 1.2)

	// Process a combat action
	learning.ProcessCombatAction(comp, true, true) // aggressive, successful

	// Process social interaction
	learning.ProcessSocialInteraction(comp, "player_123", true) // positive

	// Check skill bonus
	bonus := learning.GetSkillBonus(comp, "Basic Attack")

	// Get dominant personality trait
	dominant := comp.Personality.GetDominantTrait()

	// Retrieve recent memories
	recent := comp.Memory.GetRecentEvents(5)

# Persistence / Serialization

CompanionLearningComponent supports JSON serialization for save/load workflows:

	// Save companion state
	data, err := comp.Serialize()
	if err != nil {
		log.Printf("Failed to serialize companion: %v", err)
	}
	// data is a []byte that can be written to a save file

	// Load companion state
	restoredComp := &learning.CompanionLearningComponent{}
	err = restoredComp.Deserialize(data)
	if err != nil {
		log.Printf("Failed to restore companion: %v", err)
	}
	// restoredComp now has skill tree, personality, memory, and LastSkillUse restored

# ECS Integration

CompanionLearningComponent implements the component interface and can be attached to companion
entities. The CompanionLearningSystem runs periodic updates for skill decay and personality
normalization.

	// In game loop
	system.Update(deltaTime)

# Performance

All operations are designed for minimal overhead:
  - AddExperience: <10µs per call
  - AdjustTrait: <5µs per call
  - AddEvent: <2µs per call (LRU eviction amortized)
  - Memory storage: <1MB per companion (1000 events)
  - System update: <50ms for 100 companions

# Determinism

Skill progression and personality evolution are deterministic based on player actions. Memory
events are recorded with timestamps but retrieval is deterministic. Use seed-based RNG for
behavioral adaptation functions like AdaptBehaviorToCombatStyle().

# Testing

Package includes comprehensive tests for:
  - Skill progression and leveling
  - Personality trait adjustment and balancing
  - Memory LRU eviction
  - Behavioral adaptation
  - ECS system integration
*/
package learning
