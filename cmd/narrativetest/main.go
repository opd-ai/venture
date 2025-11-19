package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/narrative/branching"
	"github.com/opd-ai/venture/pkg/procgen"
)

func main() {
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	genre := flag.String("genre", "fantasy", "Genre (fantasy, scifi, horror, cyberpunk, postapoc)")
	depth := flag.Int("depth", 5, "Story depth (1-10)")
	interactive := flag.Bool("interactive", false, "Interactive mode to play through the story")
	verbose := flag.Bool("verbose", false, "Verbose output with all nodes")

	flag.Parse()

	fmt.Println("=== Venture Narrative System Test ===")
	fmt.Printf("Seed: %d, Genre: %s, Depth: %d\n\n", *seed, *genre, *depth)

	// Generate story arc
	gen := branching.NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      *depth,
		GenreID:    *genre,
	}

	result, err := gen.Generate(*seed, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation failed: %v\n", err)
		os.Exit(1)
	}

	arc := result.(*branching.StoryArc)

	// Validate
	if err := gen.Validate(arc); err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated Story Arc: %s\n", arc.Title)
	fmt.Printf("Description: %s\n", arc.Description)
	fmt.Printf("Nodes: %d\n", len(arc.Nodes))
	fmt.Printf("Endings: %d\n\n", len(arc.Endings))

	// Count node types
	typeCounts := make(map[branching.NodeType]int)
	for _, node := range arc.Nodes {
		typeCounts[node.Type]++
	}

	fmt.Println("Node Type Distribution:")
	for nodeType, count := range typeCounts {
		fmt.Printf("  %s: %d\n", nodeType, count)
	}
	fmt.Println()

	// List endings
	fmt.Println("Available Endings:")
	for nodeID, endingType := range arc.Endings {
		node := arc.Nodes[nodeID]
		fmt.Printf("  [%s] %s - %s\n", endingType, node.Title, node.Description)
	}
	fmt.Println()

	if *verbose {
		printAllNodes(arc)
	}

	if *interactive {
		playInteractive(arc)
	} else {
		fmt.Println("Use -interactive to play through the story")
	}
}

func printAllNodes(arc *branching.StoryArc) {
	fmt.Println("=== All Nodes ===")

	// Start with the start node
	visited := make(map[string]bool)
	printNode(arc, arc.StartNodeID, 0, visited)
}

func printNode(arc *branching.StoryArc, nodeID string, depth int, visited map[string]bool) {
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	node := arc.Nodes[nodeID]
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	fmt.Printf("%s[%s] %s\n", indent, node.Type, node.Title)
	fmt.Printf("%s  %s\n", indent, node.Description)

	if node.Type == branching.NodeTypeChoice {
		fmt.Printf("%s  Choices:\n", indent)
		for _, choice := range node.Choices {
			fmt.Printf("%s    - %s", indent, choice.Text)
			if len(choice.AlignmentShift) > 0 {
				fmt.Printf(" (alignment shifts: ")
				first := true
				for axis, shift := range choice.AlignmentShift {
					if !first {
						fmt.Printf(", ")
					}
					fmt.Printf("%s: %+.2f", axis, shift)
					first = false
				}
				fmt.Printf(")")
			}
			fmt.Println()

			// Recursively print next nodes
			printNode(arc, choice.NextNodeID, depth+1, visited)
		}
	} else if node.NextNodeID != "" {
		printNode(arc, node.NextNodeID, depth+1, visited)
	}

	fmt.Println()
}

func playInteractive(arc *branching.StoryArc) {
	manager := branching.NewManager()
	manager.RegisterArc(arc)

	progress, err := manager.StartArc("player1", arc.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start arc: %v\n", err)
		return
	}

	fmt.Println("\n=== Interactive Story Mode ===")
	fmt.Println("You are about to experience a procedurally generated story.")
	fmt.Println("Your choices will affect the outcome.")
	fmt.Println()

	for !progress.Completed {
		node, err := manager.GetCurrentNode("player1", arc.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current node: %v\n", err)
			return
		}

		fmt.Printf("\n--- %s ---\n", node.Title)
		fmt.Println(node.Description)

		if node.Type == branching.NodeTypeChoice {
			fmt.Println("\nChoices:")
			for i, choice := range node.Choices {
				fmt.Printf("  %d. %s\n", i+1, choice.Text)
			}

			var choiceNum int
			fmt.Print("\nEnter choice number: ")
			fmt.Scanf("%d", &choiceNum)

			if choiceNum < 1 || choiceNum > len(node.Choices) {
				fmt.Println("Invalid choice")
				continue
			}

			choice := node.Choices[choiceNum-1]
			err = manager.MakeChoice("player1", arc.ID, choice.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making choice: %v\n", err)
				return
			}

			// Show alignment changes
			alignment, _ := manager.GetAlignment("player1", arc.ID)
			fmt.Printf("\nCurrent Alignment:\n")
			fmt.Printf("  Good/Evil: %+.2f\n", alignment[branching.AlignmentGoodEvil])
			fmt.Printf("  Law/Chaos: %+.2f\n", alignment[branching.AlignmentLawChaos])
			fmt.Printf("  Honor/Dishonor: %+.2f\n", alignment[branching.AlignmentHonorDishonor])

		} else {
			fmt.Print("\nPress Enter to continue...")
			fmt.Scanln()

			err = manager.AdvanceStory("player1", arc.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error advancing story: %v\n", err)
				return
			}
		}

		// Update progress
		progress, _ = manager.GetProgress("player1", arc.ID)
	}

	// Story completed
	fmt.Println("\n=== Story Complete ===")
	endingNode := arc.Nodes[progress.EndingReached]
	endingType := arc.Endings[progress.EndingReached]
	fmt.Printf("\nYou reached: [%s] %s\n", endingType, endingNode.Title)
	fmt.Println(endingNode.Description)

	alignment, _ := manager.GetAlignment("player1", arc.ID)
	fmt.Printf("\nFinal Alignment:\n")
	fmt.Printf("  Good/Evil: %+.2f\n", alignment[branching.AlignmentGoodEvil])
	fmt.Printf("  Law/Chaos: %+.2f\n", alignment[branching.AlignmentLawChaos])
	fmt.Printf("  Honor/Dishonor: %+.2f\n", alignment[branching.AlignmentHonorDishonor])

	fmt.Printf("\nNodes visited: %d/%d\n", len(progress.VisitedNodes), len(arc.Nodes))
	fmt.Printf("Choices made: %d\n", len(progress.ChoicesMade))
}
