package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/integration/narrative_world"
)

func main() {
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	mode := flag.String("mode", "quest", "Test mode: quest, memory, conflict, cross-story, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Printf("=== Companion-Driven Narrative System Test ===\n")
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Printf("Mode: %s\n\n", *mode)

	manager := narrative_world.NewStoryEventManager(*seed)

	switch *mode {
	case "quest":
		testPersonalQuests(manager, *seed, *verbose)
	case "memory":
		testMemorySystem(manager, *verbose)
	case "conflict":
		testConflictDetection(manager, *verbose)
	case "cross-story":
		testCrossCompanionStory(manager, *seed, *verbose)
	case "all":
		testPersonalQuests(manager, *seed, *verbose)
		fmt.Println()
		testMemorySystem(manager, *verbose)
		fmt.Println()
		testConflictDetection(manager, *verbose)
		fmt.Println()
		testCrossCompanionStory(manager, *seed, *verbose)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func testPersonalQuests(manager *narrative_world.StoryEventManager, seed int64, verbose bool) {
	fmt.Println("--- Personal Quest Generation Test ---")

	companions := []struct {
		id       uint64
		compType engine.CompanionType
		loyalty  float64
		level    int
	}{
		{1, engine.CompanionTypePet, 0.75, 5},
		{2, engine.CompanionTypeHireling, 0.85, 10},
		{3, engine.CompanionTypeSummon, 0.9, 8},
		{4, engine.CompanionTypeUndead, 0.8, 7},
		{5, engine.CompanionTypeRobot, 0.95, 12},
	}

	for i, comp := range companions {
		companion := &engine.CompanionComponent{
			CompanionType: comp.compType,
			Loyalty:       comp.loyalty,
			Level:         comp.level,
		}

		questSeed := seed + int64(i*1000)
		quest, err := manager.GeneratePersonalQuest(comp.id, companion, questSeed)
		if err != nil {
			log.Printf("Failed to generate quest for companion %d: %v", comp.id, err)
			continue
		}

		fmt.Printf("\nCompanion %d (%v, Loyalty %.2f):\n", comp.id, comp.compType, comp.loyalty)
		fmt.Printf("  Quest: %s\n", quest.Title)
		fmt.Printf("  Unlock Loyalty: %.2f\n", quest.UnlockLoyalty)
		fmt.Printf("  Objectives (%d):\n", len(quest.Objectives))

		if verbose {
			for j, obj := range quest.Objectives {
				fmt.Printf("    %d. [%s] %s (0/%d)\n", j+1, obj.Type, obj.Description, obj.Required)
			}

			fmt.Printf("  Consequences:\n")
			for _, cons := range quest.Consequences {
				permanent := ""
				if cons.Permanent {
					permanent = " [PERMANENT]"
				}
				fmt.Printf("    - %s (Severity: %.2f)%s\n", cons.Description, cons.Severity, permanent)
			}

			if quest.StoryBranches != nil {
				fmt.Printf("  Story Branches: %d paths\n", len(quest.StoryBranches.Paths))
			}
		}
	}

	fmt.Printf("\n✅ Generated %d personal quests\n", len(companions))
}

func testMemorySystem(manager *narrative_world.StoryEventManager, verbose bool) {
	fmt.Println("--- Memory System Test ---")

	companionID := uint64(100)

	events := []struct {
		eventType   narrative_world.EventType
		description string
	}{
		{narrative_world.EventTypeCombat, "Defeated dragon in cave"},
		{narrative_world.EventTypeTreasure, "Found ancient artifact"},
		{narrative_world.EventTypeBonding, "Saved owner from trap"},
		{narrative_world.EventTypeSacrifice, "Took arrow meant for owner"},
		{narrative_world.EventTypeDanger, "Survived ambush together"},
		{narrative_world.EventTypeDiscovery, "Uncovered hidden passage"},
		{narrative_world.EventTypeConflict, "Argued about strategy"},
		{narrative_world.EventTypeBetray, "Witnessed NPC betrayal"},
	}

	for _, event := range events {
		manager.RecordMemory(companionID, event.eventType, event.description)
		if verbose {
			fmt.Printf("  Recorded: [%s] %s\n", event.eventType, event.description)
		}
	}

	memCount := manager.GetMemoryCount(companionID)
	totalCount := manager.GetTotalEventsRecorded(companionID)

	fmt.Printf("\nCompanion %d Memory:\n", companionID)
	fmt.Printf("  Stored Events: %d\n", memCount)
	fmt.Printf("  Total Recorded: %d\n", totalCount)

	// Test dialogue context
	context := manager.GetDialogueContext(companionID, 5)
	fmt.Printf("  Recent Events: %d\n", len(context.RecentEvents))
	fmt.Printf("  Important Events: %d\n", len(context.ImportantEvents))

	if verbose && len(context.ImportantEvents) > 0 {
		fmt.Println("\n  High Importance Memories:")
		for _, event := range context.ImportantEvents {
			fmt.Printf("    - [%s] %s (Importance: %.2f)\n",
				event.Type, event.Description, event.Importance)
		}
	}

	fmt.Println("\n✅ Memory system operational")
}

func testConflictDetection(manager *narrative_world.StoryEventManager, verbose bool) {
	fmt.Println("--- Conflict Detection Test ---")

	manager.SetConflictChance(0.5) // 50% for testing

	// Create companions with opposing personalities
	companions := []struct {
		id          uint64
		compType    engine.CompanionType
		personality *learning.PersonalityEvolution
	}{
		{
			id:       1,
			compType: engine.CompanionTypePet,
			personality: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitAggressive: 0.9,
					learning.TraitBrave:      0.8,
				},
			},
		},
		{
			id:       2,
			compType: engine.CompanionTypeHireling,
			personality: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitPacifist: 0.9,
					learning.TraitCautious: 0.8,
				},
			},
		},
		{
			id:       3,
			compType: engine.CompanionTypeSummon,
			personality: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitOutgoing: 0.9,
					learning.TraitCurious:  0.8,
				},
			},
		},
		{
			id:       4,
			compType: engine.CompanionTypeUndead,
			personality: &learning.PersonalityEvolution{
				Traits: map[learning.PersonalityTrait]float64{
					learning.TraitShy:         0.9,
					learning.TraitIndependent: 0.7,
				},
			},
		},
	}

	conflictCount := 0

	// Check all pairs
	for i := 0; i < len(companions); i++ {
		for j := i + 1; j < len(companions); j++ {
			comp1 := &engine.CompanionComponent{CompanionType: companions[i].compType}
			comp2 := &engine.CompanionComponent{CompanionType: companions[j].compType}

			conflict, exists := manager.CheckConflict(
				comp1, comp2,
				companions[i].id, companions[j].id,
				companions[i].personality, companions[j].personality,
			)

			if exists {
				conflictCount++
				fmt.Printf("\nConflict detected between Companion %d and %d:\n",
					companions[i].id, companions[j].id)
				fmt.Printf("  Type: %s\n", conflict.ConflictType)
				fmt.Printf("  Severity: %.2f\n", conflict.Severity)
				fmt.Printf("  Description: %s\n", conflict.Description)

				if verbose {
					fmt.Printf("  Active: %v\n", conflict.Active)
				}
			}
		}
	}

	activeConflicts := manager.GetActiveConflicts()
	fmt.Printf("\n✅ Detected %d conflicts (%d active)\n", conflictCount, len(activeConflicts))

	// Test conflict escalation
	if len(activeConflicts) > 0 {
		fmt.Println("\nSimulating 3 days of conflict escalation...")
		for i := range activeConflicts {
			initialSeverity := activeConflicts[i].Severity
			manager.UpdateConflict(i, 72*time.Hour)
			fmt.Printf("  Conflict %d: Severity %.2f → %.2f\n",
				i, initialSeverity, manager.GetActiveConflicts()[i].Severity)
		}
	}
}

func testCrossCompanionStory(manager *narrative_world.StoryEventManager, seed int64, verbose bool) {
	fmt.Println("--- Cross-Companion Story Test ---")

	// Add some shared memories
	companionIDs := []uint64{10, 11, 12}
	sharedEvent := "Defended village from raiders"

	for _, id := range companionIDs {
		manager.RecordMemory(id, narrative_world.EventTypeBonding, sharedEvent)
		manager.RecordMemory(id, narrative_world.EventTypeCombat, "Fought side by side")
	}

	if verbose {
		fmt.Println("Recorded shared memories for companions...")
	}

	story, err := manager.GenerateCrossCompanionStory(companionIDs, seed)
	if err != nil {
		log.Fatalf("Failed to generate cross-companion story: %v", err)
	}

	fmt.Printf("\nCross-Companion Story:\n")
	fmt.Printf("  Title: %s\n", story.Title)
	fmt.Printf("  Description: %s\n", story.Description)
	fmt.Printf("  Participants: %v\n", story.Participants)
	fmt.Printf("  Shared Events: %d\n", len(story.Events))
	fmt.Printf("  Outcome: %s\n", story.Outcome)
	fmt.Printf("  Active: %v\n", story.Active)

	if verbose && story.Narrative != nil {
		fmt.Printf("\n  Story Structure:\n")
		fmt.Printf("    Theme: %s\n", story.Narrative.Theme)
		fmt.Printf("    Choice Points: %d\n", len(story.Narrative.ChoicePoints))
		fmt.Printf("    Possible Paths: %d\n", len(story.Narrative.Paths))
		fmt.Printf("    Coherence: %.2f\n", story.Narrative.Coherence)
	}

	activeStories := manager.GetActiveCrossStories()
	fmt.Printf("\n✅ Generated cross-companion story (%d active)\n", len(activeStories))
}
