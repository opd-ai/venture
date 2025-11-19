package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
)

func main() {
	var (
		companionID  = flag.String("companion", "companion_001", "Companion ID")
		learningRate = flag.Float64("learning-rate", 1.0, "Learning rate multiplier")
		seed         = flag.Int64("seed", 12345, "Random seed for behavior adaptation")
		demo         = flag.String("demo", "full", "Demo mode: full, skills, personality, memory")
	)
	flag.Parse()

	fmt.Println("=== Venture Companion Learning System Demo ===")
	fmt.Printf("Companion: %s (Learning Rate: %.2f)\n\n", *companionID, *learningRate)

	manager := learning.NewManager()
	comp := manager.AddCompanion(*companionID, *learningRate)

	switch *demo {
	case "skills":
		demoSkills(comp)
	case "personality":
		demoPersonality(comp)
	case "memory":
		demoMemory(comp)
	case "full":
		demoFull(comp, *seed)
	default:
		fmt.Printf("Unknown demo mode: %s\n", *demo)
		fmt.Println("Available modes: full, skills, personality, memory")
	}
}

func demoSkills(comp *learning.CompanionLearningComponent) {
	fmt.Println("--- Skill Progression Demo ---")
	fmt.Println("\nInitial Skills: (none learned)")

	fmt.Println("\n[Simulating 10 aggressive combat actions...]")
	for i := 0; i < 10; i++ {
		learning.ProcessCombatAction(comp, true, i%3 == 0)
	}

	fmt.Println("\nSkills after combat:")
	printActiveSkills(comp)

	fmt.Printf("\nAvailable Skill Points: %d\n", comp.SkillTree.AvailablePoints)
	fmt.Printf("Total XP Earned: %.1f\n", comp.SkillTree.TotalXP)
}

func demoPersonality(comp *learning.CompanionLearningComponent) {
	fmt.Println("--- Personality Evolution Demo ---")
	fmt.Println("\nInitial Personality: All traits at 50%")

	fmt.Println("\n[Simulating aggressive combat...]")
	for i := 0; i < 20; i++ {
		learning.ProcessCombatAction(comp, true, true)
	}

	printPersonality(comp)
	fmt.Printf("\nDominant Trait: %s\n", comp.Personality.GetDominantTrait().String())
}

func demoMemory(comp *learning.CompanionLearningComponent) {
	fmt.Println("--- Event Memory Demo ---")
	fmt.Println("\n[Generating diverse events...]")

	for i := 0; i < 5; i++ {
		learning.ProcessCombatAction(comp, i%2 == 0, i%3 == 0)
		time.Sleep(time.Millisecond)
	}
	for i := 0; i < 3; i++ {
		learning.ProcessSocialInteraction(comp, fmt.Sprintf("player_%d", i), i%2 == 0)
		time.Sleep(time.Millisecond)
	}

	fmt.Println("\nMemory Summary:")
	fmt.Println(learning.GetMemorySummary(comp))
}

func demoFull(comp *learning.CompanionLearningComponent, seed int64) {
	fmt.Println("--- Full Companion Learning Demo ---")
	fmt.Println("\n[Simulating gameplay session...]")

	for i := 0; i < 10; i++ {
		learning.ProcessCombatAction(comp, i%2 == 0, i%4 == 0)
	}
	for i := 0; i < 8; i++ {
		learning.ProcessSocialInteraction(comp, "player_main", i%3 != 0)
	}
	for i := 0; i < 6; i++ {
		learning.ProcessExploration(comp, i%2 == 0)
	}

	learning.AdaptBehaviorToCombatStyle(comp, seed)

	fmt.Println("\n=== Final State ===")
	printActiveSkills(comp)
	fmt.Println()
	printPersonality(comp)
	fmt.Printf("\nLearning Progress: %.1f%%\n", learning.CalculateLearningProgress(comp)*100)
	fmt.Printf("Dominant Personality: %s\n", comp.Personality.GetDominantTrait().String())
}

func printActiveSkills(comp *learning.CompanionLearningComponent) {
	for _, skill := range comp.SkillTree.Skills {
		if skill.Level > 0 || skill.Experience > 0 {
			fmt.Printf("  %s: Level %d (%.1f XP)\n", skill.Name, skill.Level, skill.Experience)
		}
	}
}

func printPersonality(comp *learning.CompanionLearningComponent) {
	fmt.Println("Personality Traits:")
	traits := []learning.PersonalityTrait{
		learning.TraitBrave, learning.TraitCautious,
		learning.TraitOutgoing, learning.TraitShy,
		learning.TraitAggressive, learning.TraitPacifist,
	}
	for _, trait := range traits {
		value := comp.Personality.Traits[trait]
		fmt.Printf("  %-12s %.1f%%\n", trait.String(), value*100)
	}
}
