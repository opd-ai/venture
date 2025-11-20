// Package narrative_world implements companion-driven story event management.
//
// This package provides Phase 58.1 (V9.0) companion-driven narrative systems
// that create personal quests, track memory-based dialogue, manage companion
// conflicts, enable cross-companion stories, and support permanent consequences.
//
// # Key Features
//
// 1. Companion Personal Quests: Unlock at loyalty 0.7+, 3-5 quests per companion type
// 2. Memory-Based Dialogue: Companions reference 50-100 significant past events
// 3. Companion Conflicts: Personality clashes create 10-20% interaction conflicts
// 4. Cross-Companion Stories: Multiple companions interact with shared narratives
// 5. Permanent Consequences: Companion death/departure possible, no reload reversal
//
// # Architecture
//
// The package integrates:
//   - V8 Companion Learning (pkg/companion/learning): Personality and memory tracking
//   - V4 Companions (pkg/engine/companion_component): Loyalty and behavior
//   - V8 Branching Narratives (pkg/procgen/story/branching): Story structure
//
// # Usage Example
//
//	manager := narrative_world.NewStoryEventManager(12345)
//
//	// Generate companion personal quest at loyalty 0.7+
//	quest, err := manager.GeneratePersonalQuest(companionID, companion, 12345)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Record significant event for memory-based dialogue
//	manager.RecordMemory(companionID, narrative_world.EventTypeCombat, "Defeated dragon together")
//
//	// Check for personality conflicts between companions
//	conflict, exists := manager.CheckConflict(companion1, companion2)
//	if exists {
//	    fmt.Printf("Conflict: %s\n", conflict.Description)
//	}
//
//	// Generate cross-companion story
//	story, err := manager.GenerateCrossCompanionStory([]uint64{comp1ID, comp2ID}, 12345)
//
// # Performance
//
// - Quest generation: <5ms per quest
// - Memory lookup: <1ms for 100 events
// - Conflict detection: <100ns per check
// - Story generation: <10ms per cross-companion story
//
// # Test Coverage
//
// Target: ≥65% (matches V9 requirement)
package narrative_world
