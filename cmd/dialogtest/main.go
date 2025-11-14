// Command dialogtest demonstrates the procedural NPC dialog system using Markov chains.
// It allows interactive testing of dialog generation with different personalities and genres.
//
// Usage examples:
//
//	dialogtest -genre fantasy -personality helpful
//	dialogtest -genre scifi -personality scholarly -benchmark
//
// Available genres: fantasy, scifi, horror, cyberpunk, postapocalyptic
// Available personalities: helpful, merchant, hostile, mysterious, scholarly, warrior, timid, arrogant
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

var (
	seed        = flag.Int64("seed", time.Now().UnixNano(), "Random seed for generation")
	genre       = flag.String("genre", "fantasy", "Genre (fantasy, scifi, horror, cyberpunk, postapocalyptic)")
	personality = flag.String("personality", "helpful", "Personality type")
	interactive = flag.Bool("interactive", true, "Interactive mode")
	benchmark   = flag.Bool("benchmark", false, "Run performance benchmarks")
)

func main() {
	flag.Parse()

	fmt.Println("=== Venture NPC Dialog Test ===")
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Printf("Genre: %s\n", *genre)
	fmt.Printf("Personality: %s\n\n", *personality)

	// Create personality
	pers := getPersonality(*personality)
	if pers == nil {
		fmt.Printf("Unknown personality: %s\n", *personality)
		os.Exit(1)
	}

	// Create generator
	gen := dialog.NewMarkovGenerator(*seed, *genre, dialog.Order2)

	// Get corpus
	corpus := dialog.GetCorpus(*genre)
	if corpus == nil {
		fmt.Printf("Unknown genre: %s\n", *genre)
		os.Exit(1)
	}

	// Train generator
	fmt.Println("Training Markov generator...")
	gen.TrainFromCorpus(corpus.Sentences)
	fmt.Printf("Trained on %d sentences\n\n", len(corpus.Sentences))

	// Create parameters
	params := dialog.GenerateParams{
		MaxWords:    30,
		MinWords:    10,
		Temperature: 0.7,
	}
	pers.ApplyToGenerator(&params)

	if *benchmark {
		runBenchmarks(gen, pers, params)
	} else if *interactive {
		runInteractive(gen, pers, params)
	} else {
		runSingleShot(gen, pers, params)
	}
}

func getPersonality(name string) *dialog.Personality {
	switch strings.ToLower(name) {
	case "helpful":
		return dialog.NewPersonality(dialog.PersonalityHelpful)
	case "merchant":
		return dialog.NewPersonality(dialog.PersonalityMerchant)
	case "hostile":
		return dialog.NewPersonality(dialog.PersonalityHostile)
	case "mysterious":
		return dialog.NewPersonality(dialog.PersonalityMysterious)
	case "scholarly":
		return dialog.NewPersonality(dialog.PersonalityScholarly)
	case "warrior":
		return dialog.NewPersonality(dialog.PersonalityWarrior)
	case "timid":
		return dialog.NewPersonality(dialog.PersonalityTimid)
	case "arrogant":
		return dialog.NewPersonality(dialog.PersonalityArrogant)
	default:
		return nil
	}
}

func runSingleShot(gen *dialog.MarkovGenerator, pers *dialog.Personality, params dialog.GenerateParams) {
	fmt.Println("=== Single-Shot Generation ===")

	params.PlayerInput = "Hello, what can you tell me?"
	params.ConversationID = "conversation-1"

	response := gen.Generate(params)

	fmt.Printf("\nPlayer: %s\n", params.PlayerInput)
	fmt.Printf("NPC (%s): %s\n", pers.Type, response)
	fmt.Printf("\nWord count: %d\n", len(strings.Fields(response)))
}

func runInteractive(gen *dialog.MarkovGenerator, pers *dialog.Personality, params dialog.GenerateParams) {
	fmt.Println("=== Interactive Mode ===")
	fmt.Println("Enter player input (or 'quit' to exit, 'greeting' for greeting):")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	conversationID := fmt.Sprintf("conversation-%d", time.Now().Unix())
	exchangeCount := 0

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		playerInput := strings.TrimSpace(scanner.Text())

		if playerInput == "" {
			continue
		}

		if playerInput == "quit" || playerInput == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if playerInput == "greeting" {
			greeting := pers.GetGreeting(*genre)
			fmt.Printf("NPC (%s): %s\n", pers.Type, greeting)
			continue
		}

		exchangeCount++

		params.PlayerInput = playerInput
		params.ConversationID = conversationID

		response := gen.Generate(params)

		fmt.Printf("NPC (%s): %s\n", pers.Type, response)
		fmt.Printf("  [%d words, exchange #%d]\n", len(strings.Fields(response)), exchangeCount)
	}
}

func runBenchmarks(gen *dialog.MarkovGenerator, pers *dialog.Personality, params dialog.GenerateParams) {
	fmt.Println("=== Performance Benchmarks ===")
	fmt.Println()

	// Benchmark 1: Single generation
	params.PlayerInput = "test"
	params.ConversationID = "bench-1"

	start := time.Now()
	gen.Generate(params)
	single := time.Since(start)
	fmt.Printf("Single generation: %v\n", single)

	// Benchmark 2: 100 generations
	start = time.Now()
	for i := 0; i < 100; i++ {
		params.ConversationID = fmt.Sprintf("bench-%d", i)
		gen.Generate(params)
	}
	hundred := time.Since(start)
	fmt.Printf("100 generations: %v (avg: %v)\n", hundred, hundred/100)

	// Benchmark 3: Variation test
	fmt.Println("\n=== Variation Test ===")
	responses := make(map[string]bool)
	params.PlayerInput = "hello"
	for i := 0; i < 10; i++ {
		params.ConversationID = fmt.Sprintf("var-%d", i)
		resp := gen.Generate(params)
		responses[resp] = true
	}
	fmt.Printf("Unique responses: %d/10 (%.1f%%)\n", len(responses), float64(len(responses))*10.0)

	// Benchmark 4: Determinism test
	fmt.Println("\n=== Determinism Test (GenerateDeterministic) ===")
	first := gen.GenerateDeterministic(params)
	allSame := true
	for i := 0; i < 5; i++ {
		resp := gen.GenerateDeterministic(params)
		if resp != first {
			allSame = false
			break
		}
	}
	fmt.Printf("Deterministic consistency: %v\n", allSame)

	fmt.Printf("\n✓ Target: <50ms per response (achieved: %v)\n", single)
}
