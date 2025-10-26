package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/sirupsen/logrus"
)

var (
	genre   = flag.String("genre", "fantasy", "Genre (fantasy, scifi)")
	count   = flag.Int("count", 3, "Number of skill trees to generate")
	depth   = flag.Int("depth", 5, "Depth level (affects scaling)")
	seed    = flag.Int64("seed", 0, "Random seed (0 for current time)")
	verbose = flag.Bool("verbose", false, "Show detailed output")
	output  = flag.String("output", "", "Output file (empty for stdout)")
)

func main() {
	flag.Parse()

	// Use current time as seed if not specified
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	// Initialize logger for test utility
	logger := logging.TestUtilityLogger("skilltest")
	testLogger := logger.WithFields(logrus.Fields{
		"genre": *genre,
		"count": *count,
		"depth": *depth,
		"seed":  *seed,
	})

	testLogger.Info("generating skill trees")

	// Create generator
	generator := skills.NewSkillTreeGenerator()

	// Set up generation parameters
	params := procgen.GenerationParams{
		Depth:      *depth,
		Difficulty: 0.5,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	// Generate skill trees
	genLogger := logging.GeneratorLogger(logger, "skill-tree", *seed, *genre)
	genLogger.Debug("starting skill tree generation")

	start := time.Now()
	result, err := generator.Generate(*seed, params)
	if err != nil {
		genLogger.WithError(err).Fatal("generation failed")
	}
	elapsed := time.Since(start)

	trees, ok := result.([]*skills.SkillTree)
	if !ok {
		genLogger.WithField("resultType", fmt.Sprintf("%T", result)).Fatal("unexpected result type")
	}

	// Validate
	if err := generator.Validate(result); err != nil {
		genLogger.WithError(err).Fatal("validation failed")
	}

	genLogger.WithFields(logrus.Fields{
		"treeCount": len(trees),
		"duration":  elapsed,
	}).Info("skill trees generated successfully")

	// Format output
	var out *os.File
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			testLogger.WithError(err).WithField("outputFile", *output).Fatal("failed to create output file")
		}
		defer f.Close()
		out = f
		testLogger.WithField("outputFile", *output).Info("writing skill trees to file")
	} else {
		out = os.Stdout
	}

	// Print trees
	printTrees(out, trees, *verbose)

	if *output != "" {
		log.Printf("Output written to %s", *output)
	}
}

func printTrees(out *os.File, trees []*skills.SkillTree, verbose bool) {
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(out, "                        SKILL TREE GENERATION                           \n")
	fmt.Fprintf(out, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(out, "\n")

	for i, tree := range trees {
		fmt.Fprintf(out, "┌─────────────────────────────────────────────────────────────────────┐\n")
		fmt.Fprintf(out, "│ Tree %d: %-60s │\n", i+1, tree.Name)
		fmt.Fprintf(out, "├─────────────────────────────────────────────────────────────────────┤\n")
		fmt.Fprintf(out, "│ Description: %-56s │\n", tree.Description)
		fmt.Fprintf(out, "│ Category:    %-56s │\n", tree.Category)
		fmt.Fprintf(out, "│ Genre:       %-56s │\n", tree.Genre)
		fmt.Fprintf(out, "│ Max Points:  %-56d │\n", tree.MaxPoints)
		fmt.Fprintf(out, "│ Total Skills: %-55d │\n", len(tree.Nodes))
		fmt.Fprintf(out, "│ Root Skills:  %-55d │\n", len(tree.RootNodes))
		fmt.Fprintf(out, "└─────────────────────────────────────────────────────────────────────┘\n")
		fmt.Fprintf(out, "\n")

		if verbose {
			printTreeDetailed(out, tree)
		} else {
			printTreeSummary(out, tree)
		}

		fmt.Fprintf(out, "\n")
	}

	// Statistics
	fmt.Fprintf(out, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(out, "                              STATISTICS                                \n")
	fmt.Fprintf(out, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(out, "\n")

	totalSkills := 0
	totalPassive := 0
	totalActive := 0
	totalUltimate := 0
	totalSynergy := 0

	for _, tree := range trees {
		totalSkills += len(tree.Nodes)
		for _, node := range tree.Nodes {
			switch node.Skill.Type {
			case skills.TypePassive:
				totalPassive++
			case skills.TypeActive:
				totalActive++
			case skills.TypeUltimate:
				totalUltimate++
			case skills.TypeSynergy:
				totalSynergy++
			}
		}
	}

	fmt.Fprintf(out, "Total Trees:      %d\n", len(trees))
	fmt.Fprintf(out, "Total Skills:     %d\n", totalSkills)
	fmt.Fprintf(out, "  - Passive:      %d (%.1f%%)\n", totalPassive, float64(totalPassive)/float64(totalSkills)*100)
	fmt.Fprintf(out, "  - Active:       %d (%.1f%%)\n", totalActive, float64(totalActive)/float64(totalSkills)*100)
	fmt.Fprintf(out, "  - Ultimate:     %d (%.1f%%)\n", totalUltimate, float64(totalUltimate)/float64(totalSkills)*100)
	fmt.Fprintf(out, "  - Synergy:      %d (%.1f%%)\n", totalSynergy, float64(totalSynergy)/float64(totalSkills)*100)
	fmt.Fprintf(out, "\n")
}

func printTreeSummary(out *os.File, tree *skills.SkillTree) {
	// Group by tier
	tierGroups := make(map[skills.Tier][]*skills.Skill)
	for _, node := range tree.Nodes {
		tierGroups[node.Skill.Tier] = append(tierGroups[node.Skill.Tier], node.Skill)
	}

	// Print tiers
	tiers := []skills.Tier{skills.TierBasic, skills.TierIntermediate, skills.TierAdvanced, skills.TierMaster}
	for _, tier := range tiers {
		skillsInTier := tierGroups[tier]
		if len(skillsInTier) == 0 {
			continue
		}

		fmt.Fprintf(out, "  %s Tier (%d skills):\n", tier, len(skillsInTier))
		for _, skill := range skillsInTier {
			typeIndicator := ""
			switch skill.Type {
			case skills.TypePassive:
				typeIndicator = "[P]"
			case skills.TypeActive:
				typeIndicator = "[A]"
			case skills.TypeUltimate:
				typeIndicator = "[U]"
			case skills.TypeSynergy:
				typeIndicator = "[S]"
			}

			fmt.Fprintf(out, "    %s %-40s (Max Lv: %d)\n", typeIndicator, skill.Name, skill.MaxLevel)
		}
		fmt.Fprintf(out, "\n")
	}
}

func printTreeDetailed(out *os.File, tree *skills.SkillTree) {
	// Group by tier
	tierGroups := make(map[skills.Tier][]*skills.SkillNode)
	for _, node := range tree.Nodes {
		tierGroups[node.Skill.Tier] = append(tierGroups[node.Skill.Tier], node)
	}

	// Print tiers
	tiers := []skills.Tier{skills.TierBasic, skills.TierIntermediate, skills.TierAdvanced, skills.TierMaster}
	for _, tier := range tiers {
		nodes := tierGroups[tier]
		if len(nodes) == 0 {
			continue
		}

		fmt.Fprintf(out, "  ┌───────────────────────────────────────────────────────────────────┐\n")
		fmt.Fprintf(out, "  │ %s Tier - %d Skills%46s │\n", tier, len(nodes), "")
		fmt.Fprintf(out, "  └───────────────────────────────────────────────────────────────────┘\n")
		fmt.Fprintf(out, "\n")

		for _, node := range nodes {
			skill := node.Skill

			typeStr := ""
			switch skill.Type {
			case skills.TypePassive:
				typeStr = "PASSIVE"
			case skills.TypeActive:
				typeStr = "ACTIVE"
			case skills.TypeUltimate:
				typeStr = "ULTIMATE"
			case skills.TypeSynergy:
				typeStr = "SYNERGY"
			}

			fmt.Fprintf(out, "    ┌─ %s [%s]\n", skill.Name, typeStr)
			fmt.Fprintf(out, "    │  %s\n", skill.Description)
			fmt.Fprintf(out, "    │\n")
			fmt.Fprintf(out, "    │  Requirements:\n")
			fmt.Fprintf(out, "    │    • Player Level: %d\n", skill.Requirements.PlayerLevel)
			fmt.Fprintf(out, "    │    • Skill Points: %d\n", skill.Requirements.SkillPoints)
			if len(skill.Requirements.PrerequisiteIDs) > 0 {
				fmt.Fprintf(out, "    │    • Prerequisites: %d skill(s)\n", len(skill.Requirements.PrerequisiteIDs))
			}
			fmt.Fprintf(out, "    │\n")
			fmt.Fprintf(out, "    │  Effects (per level, max %d):\n", skill.MaxLevel)
			for _, effect := range skill.Effects {
				fmt.Fprintf(out, "    │    • %s\n", effect.Description)
			}
			fmt.Fprintf(out, "    │\n")
			fmt.Fprintf(out, "    │  Tags: %v\n", skill.Tags)
			fmt.Fprintf(out, "    └──────────────────────────────────────────────────────────────\n")
			fmt.Fprintf(out, "\n")
		}
	}
}
