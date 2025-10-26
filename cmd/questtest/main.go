package main

import (
	"flag"
	"fmt"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/sirupsen/logrus"
)

var (
	seed       = flag.Int64("seed", 12345, "Random seed for generation")
	count      = flag.Int("count", 5, "Number of quests to generate")
	depth      = flag.Int("depth", 5, "Dungeon depth (affects difficulty and rewards)")
	difficulty = flag.Float64("difficulty", 0.5, "Difficulty multiplier (0.0-1.0)")
	genre      = flag.String("genre", "fantasy", "Genre (fantasy, scifi, horror, cyberpunk, postapoc)")
)

func main() {
	flag.Parse()

	// Initialize logger for test utility
	logger := logging.TestUtilityLogger("questtest")
	testLogger := logger.WithFields(logrus.Fields{
		"seed":       *seed,
		"genre":      *genre,
		"depth":      *depth,
		"difficulty": *difficulty,
		"count":      *count,
	})

	fmt.Println("=== Venture Quest Generator Test ===")
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Printf("Genre: %s\n", *genre)
	fmt.Printf("Depth: %d, Difficulty: %.2f\n", *depth, *difficulty)
	fmt.Printf("Generating %d quests...\n\n", *count)

	testLogger.Info("generating quests")

	// Create generator
	generator := quest.NewQuestGenerator()

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	// Generate quests
	genLogger := logging.GeneratorLogger(logger, "quest", *seed, *genre)
	genLogger.Debug("starting quest generation")

	result, err := generator.Generate(*seed, params)
	if err != nil {
		genLogger.WithError(err).Fatal("generation failed")
	}

	// Validate
	if err := generator.Validate(result); err != nil {
		genLogger.WithError(err).Fatal("validation failed")
	}

	quests := result.([]*quest.Quest)

	genLogger.WithField("questCount", len(quests)).Info("quests generated successfully")

	// Display quests
	for i, q := range quests {
		displayQuest(i+1, q)
	}

	// Summary statistics
	fmt.Println("\n=== Summary Statistics ===")
	displayStatistics(quests)
}

func displayQuest(num int, q *quest.Quest) {
	fmt.Printf("--- Quest %d: %s ---\n", num, q.Name)
	fmt.Printf("ID: %s\n", q.ID)
	fmt.Printf("Type: %s\n", q.Type)
	fmt.Printf("Difficulty: %s\n", q.Difficulty)
	fmt.Printf("Status: %s\n", q.Status)
	fmt.Printf("Required Level: %d\n", q.RequiredLevel)

	if q.GiverNPC != "" {
		fmt.Printf("Quest Giver: %s\n", q.GiverNPC)
	}

	if q.Location != "" {
		fmt.Printf("Location: %s\n", q.Location)
	}

	fmt.Printf("\nDescription:\n  %s\n", q.Description)

	fmt.Printf("\nObjectives:\n")
	for i, obj := range q.Objectives {
		fmt.Printf("  %d. %s\n", i+1, obj.Description)
		fmt.Printf("     Progress: %d/%d (%.1f%%)\n", obj.Current, obj.Required, obj.Progress()*100)
	}

	fmt.Printf("\nRewards:\n")
	fmt.Printf("  XP: %d\n", q.Reward.XP)
	fmt.Printf("  Gold: %d\n", q.Reward.Gold)

	if len(q.Reward.Items) > 0 {
		fmt.Printf("  Items: %d\n", len(q.Reward.Items))
		for _, item := range q.Reward.Items {
			fmt.Printf("    - %s\n", item)
		}
	}

	if q.Reward.SkillPoints > 0 {
		fmt.Printf("  Skill Points: %d\n", q.Reward.SkillPoints)
	}

	fmt.Printf("  Estimated Value: %d\n", q.GetRewardValue())

	if len(q.Tags) > 0 {
		fmt.Printf("\nTags: %v\n", q.Tags)
	}

	fmt.Printf("Seed: %d\n", q.Seed)
	fmt.Println()
}

func displayStatistics(quests []*quest.Quest) {
	// Count by type
	typeCounts := make(map[quest.QuestType]int)
	for _, q := range quests {
		typeCounts[q.Type]++
	}

	fmt.Println("Quest Types:")
	for qType, count := range typeCounts {
		fmt.Printf("  %s: %d\n", qType, count)
	}

	// Count by difficulty
	diffCounts := make(map[quest.Difficulty]int)
	for _, q := range quests {
		diffCounts[q.Difficulty]++
	}

	fmt.Println("\nDifficulty Distribution:")
	for diff, count := range diffCounts {
		fmt.Printf("  %s: %d\n", diff, count)
	}

	// Average rewards
	totalXP := 0
	totalGold := 0
	totalItems := 0
	totalSkillPoints := 0

	for _, q := range quests {
		totalXP += q.Reward.XP
		totalGold += q.Reward.Gold
		totalItems += len(q.Reward.Items)
		totalSkillPoints += q.Reward.SkillPoints
	}

	fmt.Println("\nAverage Rewards:")
	fmt.Printf("  XP: %d\n", totalXP/len(quests))
	fmt.Printf("  Gold: %d\n", totalGold/len(quests))
	fmt.Printf("  Items: %.1f\n", float64(totalItems)/float64(len(quests)))
	fmt.Printf("  Skill Points: %.1f\n", float64(totalSkillPoints)/float64(len(quests)))

	// Total value
	totalValue := 0
	for _, q := range quests {
		totalValue += q.GetRewardValue()
	}
	fmt.Printf("\nTotal Estimated Value: %d\n", totalValue)
	fmt.Printf("Average Value per Quest: %d\n", totalValue/len(quests))
}
