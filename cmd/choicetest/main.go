package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/integration/choice_consequences"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, alignment, npc, quest, class, companion, save, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("=== Choice Consequence System Test ===")
	fmt.Printf("Mode: %s\n\n", *mode)

	switch *mode {
	case "demo":
		runDemo(*verbose)
	case "alignment":
		testAlignment(*verbose)
	case "npc":
		testNPCRelationships(*verbose)
	case "quest":
		testQuestBranches(*verbose)
	case "class":
		testClassQuests(*verbose)
	case "companion":
		testCompanionReactions(*verbose)
	case "save":
		testSaveLoad(*verbose)
	case "all":
		runDemo(*verbose)
		testAlignment(*verbose)
		testNPCRelationships(*verbose)
		testQuestBranches(*verbose)
		testClassQuests(*verbose)
		testCompanionReactions(*verbose)
		testSaveLoad(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
	}
}

func runDemo(verbose bool) {
	fmt.Println("--- Demo: Complete Story Path ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "hero_demo"

	// Story: Village under attack
	fmt.Println("\n1. Village Attack Scenario")

	choice1 := &choice_consequences.PlayerChoice{
		ChoiceID:    "defend_village",
		StoryNodeID: "village_attack",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &choice_consequences.AlignmentShift{
			GoodEvil:      0.3,
			LawChaos:      0.2,
			HonorDishonor: 0.3,
		},
		NPCsAffected: []string{"village_elder", "village_guard", "villagers"},
		Irreversible: false,
	}

	tracker.RecordChoice(playerID, choice1)
	fmt.Printf("  Choice: Defend the village from bandits\n")

	if verbose {
		alignment := tracker.GetAlignment(playerID)
		fmt.Printf("  Alignment: Good=%0.2f, Law=%0.2f, Honor=%0.2f\n",
			alignment.GoodEvil, alignment.LawChaos, alignment.HonorDishonor)
	}

	// Choice 2: Spare or execute bandit leader
	choice2 := &choice_consequences.PlayerChoice{
		ChoiceID:    "spare_bandit_leader",
		StoryNodeID: "bandit_confrontation",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &choice_consequences.AlignmentShift{
			GoodEvil:      0.4,  // Merciful (good)
			LawChaos:      -0.3, // Breaking law (chaotic)
			HonorDishonor: 0.1,
		},
		NPCsAffected: []string{"village_elder", "bandit_leader"},
		Irreversible: true,
		Consequences: []string{"lock_quest_execute_bandits", "unlock_quest_redemption"},
	}

	tracker.RecordChoice(playerID, choice2)
	fmt.Printf("\n2. Bandit Leader Captured\n")
	fmt.Printf("  Choice: Spare the bandit leader (mercy over justice)\n")

	// Check what content is available
	canExecute := tracker.IsContentAvailable(playerID, "execute_bandits")
	canRedeem := tracker.IsContentAvailable(playerID, "quest_redemption")

	fmt.Printf("  Execute quest available: %v\n", canExecute)
	fmt.Printf("  Redemption quest available: %v\n", canRedeem)

	// NPC reactions
	elderAttitude := tracker.GetNPCAttitude(playerID, "village_elder")
	banditAttitude := tracker.GetNPCAttitude(playerID, "bandit_leader")

	fmt.Printf("\n3. NPC Relationships:\n")
	fmt.Printf("  Village Elder: %s (attitude: %0.2f)\n",
		getAttitudeDescription(elderAttitude), elderAttitude)
	fmt.Printf("  Bandit Leader: %s (attitude: %0.2f)\n",
		getAttitudeDescription(banditAttitude), banditAttitude)

	// Final alignment
	alignment := tracker.GetAlignment(playerID)
	fmt.Printf("\n4. Final Alignment:\n")
	fmt.Printf("  Good/Evil: %s (%0.2f)\n",
		getAlignmentDescription(alignment.GoodEvil), alignment.GoodEvil)
	fmt.Printf("  Law/Chaos: %s (%0.2f)\n",
		getAlignmentDescription(alignment.LawChaos), alignment.LawChaos)
	fmt.Printf("  Honor/Dishonor: %s (%0.2f)\n",
		getAlignmentDescription(alignment.HonorDishonor), alignment.HonorDishonor)

	fmt.Printf("\nTotal choices made: %d\n", tracker.GetChoiceCount(playerID))
	fmt.Printf("NPC relationships: %d\n", tracker.GetNPCRelationshipCount(playerID))
}

func testAlignment(verbose bool) {
	fmt.Println("\n--- Alignment System Test ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "alignment_test"

	choices := []struct {
		name  string
		shift *choice_consequences.AlignmentShift
	}{
		{"Save innocent", &choice_consequences.AlignmentShift{GoodEvil: 0.5}},
		{"Follow the law", &choice_consequences.AlignmentShift{LawChaos: 0.3}},
		{"Act honorably", &choice_consequences.AlignmentShift{HonorDishonor: 0.4}},
		{"Commit crime", &choice_consequences.AlignmentShift{LawChaos: -0.5}},
		{"Selfish act", &choice_consequences.AlignmentShift{GoodEvil: -0.2}},
	}

	for i, c := range choices {
		choice := &choice_consequences.PlayerChoice{
			ChoiceID:       fmt.Sprintf("choice_%d", i),
			StoryNodeID:    "test_node",
			Timestamp:      time.Now().Unix(),
			MoralAlignment: c.shift,
		}
		tracker.RecordChoice(playerID, choice)

		if verbose {
			alignment := tracker.GetAlignment(playerID)
			fmt.Printf("  %s: Good=%0.2f, Law=%0.2f, Honor=%0.2f\n",
				c.name, alignment.GoodEvil, alignment.LawChaos, alignment.HonorDishonor)
		}
	}

	alignment := tracker.GetAlignment(playerID)
	fmt.Printf("\nFinal Alignment:\n")
	fmt.Printf("  Good/Evil: %0.2f (%s)\n",
		alignment.GoodEvil, getAlignmentDescription(alignment.GoodEvil))
	fmt.Printf("  Law/Chaos: %0.2f (%s)\n",
		alignment.LawChaos, getAlignmentDescription(alignment.LawChaos))
	fmt.Printf("  Honor/Dishonor: %0.2f (%s)\n",
		alignment.HonorDishonor, getAlignmentDescription(alignment.HonorDishonor))
}

func testNPCRelationships(verbose bool) {
	fmt.Println("\n--- NPC Relationship Test ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "npc_test"

	npcs := []struct {
		id    string
		name  string
		shift *choice_consequences.AlignmentShift
	}{
		{"merchant", "Honest Merchant", &choice_consequences.AlignmentShift{GoodEvil: 0.4, HonorDishonor: 0.3}},
		{"guard", "Town Guard", &choice_consequences.AlignmentShift{LawChaos: 0.5}},
		{"thief", "Shadow Thief", &choice_consequences.AlignmentShift{GoodEvil: -0.3, LawChaos: -0.4}},
	}

	fmt.Println("\nRecording interactions:")
	for i, npc := range npcs {
		choice := &choice_consequences.PlayerChoice{
			ChoiceID:       fmt.Sprintf("interact_%s", npc.id),
			StoryNodeID:    "interaction",
			Timestamp:      time.Now().Unix(),
			MoralAlignment: npc.shift,
			NPCsAffected:   []string{npc.id},
		}
		tracker.RecordChoice(playerID, choice)

		attitude := tracker.GetNPCAttitude(playerID, npc.id)
		fmt.Printf("  %d. %s: %s (attitude: %0.2f)\n",
			i+1, npc.name, getAttitudeDescription(attitude), attitude)
	}

	fmt.Printf("\nTotal NPC relationships tracked: %d\n",
		tracker.GetNPCRelationshipCount(playerID))
}

func testQuestBranches(verbose bool) {
	fmt.Println("\n--- Quest Branch System Test ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "quest_test"

	// Register branching quest
	branch := &choice_consequences.QuestBranch{
		QuestID:       "civil_war",
		BranchID:      "support_rebels",
		Prerequisites: []string{"meet_rebels", "distrust_king"},
	}
	tracker.RegisterQuestBranch(branch)

	fmt.Println("Quest: Civil War")
	fmt.Println("Branch: Support Rebels")
	fmt.Printf("Prerequisites: %v\n\n", branch.Prerequisites)

	// Check availability before prerequisites
	available := tracker.IsQuestBranchAvailable(playerID, "civil_war", "support_rebels")
	fmt.Printf("1. Initially available: %v\n", available)

	// Complete first prerequisite
	tracker.RecordChoice(playerID, &choice_consequences.PlayerChoice{
		ChoiceID:    "meet_rebels",
		StoryNodeID: "forest_camp",
		Timestamp:   time.Now().Unix(),
	})

	available = tracker.IsQuestBranchAvailable(playerID, "civil_war", "support_rebels")
	fmt.Printf("2. After meeting rebels: %v\n", available)

	// Complete second prerequisite
	tracker.RecordChoice(playerID, &choice_consequences.PlayerChoice{
		ChoiceID:    "distrust_king",
		StoryNodeID: "throne_room",
		Timestamp:   time.Now().Unix(),
	})

	available = tracker.IsQuestBranchAvailable(playerID, "civil_war", "support_rebels")
	fmt.Printf("3. After distrusting king: %v\n", available)
}

func testClassQuests(verbose bool) {
	fmt.Println("\n--- Class-Specific Quest Test ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "class_test"

	// Register paladin quest
	quest := &choice_consequences.ClassSpecificQuest{
		QuestID:       "holy_trial",
		RequiredClass: "Paladin",
		MinLevel:      20,
		AlignmentReq: &choice_consequences.AlignmentRequirement{
			MinGoodEvil:      0.5,
			MaxGoodEvil:      1.0,
			MinLawChaos:      0.3,
			MaxLawChaos:      1.0,
			MinHonorDishonor: 0.4,
			MaxHonorDishonor: 1.0,
		},
	}
	tracker.RegisterClassQuest(quest)

	fmt.Println("Quest: Holy Trial")
	fmt.Println("Class: Paladin")
	fmt.Println("Min Level: 20")
	fmt.Println("Alignment: Good (0.5+), Lawful (0.3+), Honorable (0.4+)\n")

	// Test different scenarios
	scenarios := []struct {
		name  string
		class string
		level int
		shift *choice_consequences.AlignmentShift
	}{
		{"Wrong class", "Warrior", 25, &choice_consequences.AlignmentShift{GoodEvil: 0.8}},
		{"Too low level", "Paladin", 15, &choice_consequences.AlignmentShift{GoodEvil: 0.8}},
		{"Wrong alignment", "Paladin", 25, &choice_consequences.AlignmentShift{GoodEvil: -0.5}},
		{"Perfect match", "Paladin", 25, &choice_consequences.AlignmentShift{GoodEvil: 0.8, LawChaos: 0.5, HonorDishonor: 0.6}},
	}

	for i, scenario := range scenarios {
		testID := fmt.Sprintf("%s_%d", playerID, i)

		if scenario.shift != nil {
			choice := &choice_consequences.PlayerChoice{
				ChoiceID:       "setup",
				StoryNodeID:    "test",
				Timestamp:      time.Now().Unix(),
				MoralAlignment: scenario.shift,
			}
			tracker.RecordChoice(testID, choice)
		}

		available := tracker.IsClassQuestAvailable(testID, "holy_trial", scenario.class, scenario.level)
		fmt.Printf("%d. %s: %v\n", i+1, scenario.name, available)
	}
}

func testCompanionReactions(verbose bool) {
	fmt.Println("\n--- Companion Reaction Test ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "companion_test"

	companions := []struct {
		id       string
		name     string
		choiceID string
		delta    float64
		approval bool
		comment  string
	}{
		{"wolf", "Wolf Companion", "hunt_deer", 0.15, true, "Wolf enjoys the hunt"},
		{"cleric", "Holy Cleric", "spare_enemy", 0.20, true, "Cleric approves of mercy"},
		{"assassin", "Dark Assassin", "spare_enemy", -0.10, false, "Assassin thinks you're weak"},
		{"knight", "Loyal Knight", "defend_innocent", 0.25, true, "Knight respects your honor"},
	}

	fmt.Println("Recording companion reactions:")
	for i, comp := range companions {
		reaction := &choice_consequences.CompanionReaction{
			CompanionID:  comp.id,
			ChoiceID:     comp.choiceID,
			LoyaltyDelta: comp.delta,
			Approval:     comp.approval,
			Comment:      comp.comment,
		}

		tracker.RecordCompanionReaction(playerID, reaction)

		approvalText := "approves"
		if !comp.approval {
			approvalText = "disapproves"
		}

		fmt.Printf("  %d. %s %s: %+0.2f loyalty\n",
			i+1, comp.name, approvalText, comp.delta)

		if verbose {
			fmt.Printf("     \"%s\"\n", comp.comment)
		}
	}

	// Get reactions for specific companions
	fmt.Println("\nWolf companion reactions:")
	wolfReactions := tracker.GetCompanionReactions(playerID, "wolf")
	for _, r := range wolfReactions {
		fmt.Printf("  - %s (%+0.2f loyalty)\n", r.Comment, r.LoyaltyDelta)
	}
}

func testSaveLoad(verbose bool) {
	fmt.Println("\n--- Save/Load Test ---")

	tracker := choice_consequences.NewChoiceTracker()
	playerID := "save_test"

	// Create some data
	for i := 0; i < 5; i++ {
		choice := &choice_consequences.PlayerChoice{
			ChoiceID:    fmt.Sprintf("choice_%d", i),
			StoryNodeID: "test_node",
			Timestamp:   time.Now().Unix(),
			MoralAlignment: &choice_consequences.AlignmentShift{
				GoodEvil: 0.1 * float64(i),
			},
			NPCsAffected: []string{fmt.Sprintf("npc_%d", i)},
		}
		tracker.RecordChoice(playerID, choice)
	}

	originalChoices := tracker.GetChoiceCount(playerID)
	originalNPCs := tracker.GetNPCRelationshipCount(playerID)
	originalAlignment := tracker.GetAlignment(playerID)

	fmt.Printf("Original data:\n")
	fmt.Printf("  Choices: %d\n", originalChoices)
	fmt.Printf("  NPC relationships: %d\n", originalNPCs)
	fmt.Printf("  Alignment: Good=%0.2f\n", originalAlignment.GoodEvil)

	// Save
	filename := "test_choices.json.gz"
	err := tracker.Save(filename)
	if err != nil {
		log.Fatalf("Save failed: %v", err)
	}
	fmt.Printf("\nSaved to: %s\n", filename)

	// Load into new tracker
	newTracker := choice_consequences.NewChoiceTracker()
	err = newTracker.Load(filename)
	if err != nil {
		log.Fatalf("Load failed: %v", err)
	}

	loadedChoices := newTracker.GetChoiceCount(playerID)
	loadedNPCs := newTracker.GetNPCRelationshipCount(playerID)
	loadedAlignment := newTracker.GetAlignment(playerID)

	fmt.Printf("\nLoaded data:\n")
	fmt.Printf("  Choices: %d\n", loadedChoices)
	fmt.Printf("  NPC relationships: %d\n", loadedNPCs)
	fmt.Printf("  Alignment: Good=%0.2f\n", loadedAlignment.GoodEvil)

	success := originalChoices == loadedChoices &&
		originalNPCs == loadedNPCs &&
		originalAlignment.GoodEvil == loadedAlignment.GoodEvil

	if success {
		fmt.Println("\n✓ Save/Load successful!")
	} else {
		fmt.Println("\n✗ Save/Load verification failed")
	}
}

func getAttitudeDescription(attitude float64) string {
	if attitude >= 0.8 {
		return "Beloved"
	} else if attitude >= 0.5 {
		return "Friendly"
	} else if attitude >= 0.2 {
		return "Positive"
	} else if attitude >= -0.2 {
		return "Neutral"
	} else if attitude >= -0.5 {
		return "Unfriendly"
	} else if attitude >= -0.8 {
		return "Hostile"
	}
	return "Hated"
}

func getAlignmentDescription(value float64) string {
	if value >= 0.7 {
		return "Very High"
	} else if value >= 0.4 {
		return "High"
	} else if value >= 0.1 {
		return "Slightly Positive"
	} else if value >= -0.1 {
		return "Neutral"
	} else if value >= -0.4 {
		return "Slightly Negative"
	} else if value >= -0.7 {
		return "Low"
	}
	return "Very Low"
}
