// Package branching implements complex procedural storytelling with branching narratives.
//
// This package provides a complete system for generating and managing story arcs with
// multiple possible paths, player choices, moral alignment tracking, faction reputation,
// and consequence systems. All content is procedurally generated based on genre and seed.
//
// # Key Features
//
// - Procedural story arc generation with 10-20 nodes
// - Multiple ending types (Heroic, Tragic, Neutral, Mystery, Triumph, Betrayal)
// - Player choice tracking with moral alignment shifts
// - Faction reputation system
// - Consequence system affecting future quests and NPCs
// - Branching quest chains based on player decisions
//
// # Architecture
//
// The package consists of several key components:
//
// - Generator: Implements procgen.Generator for creating story arcs
// - Manager: Handles player progress and choice processing
// - StoryArc: Complete narrative structure with nodes and endings
// - StoryNode: Individual points in the narrative (Start, Choice, Event, Consequence, Ending)
// - PlayerProgress: Tracks player's journey through a story arc
// - Consequence: Represents effects of player choices on the game world
//
// # Usage Example
//
//	// Generate a story arc
//	gen := branching.NewGenerator()
//	params := procgen.GenerationParams{
//		Difficulty: 0.5,
//		Depth:      5,
//		GenreID:    "fantasy",
//	}
//	result, _ := gen.Generate(12345, params)
//	arc := result.(*branching.StoryArc)
//
//	// Start the arc for a player
//	manager := branching.NewManager()
//	manager.RegisterArc(arc)
//	progress, _ := manager.StartArc("player1", arc.ID)
//
//	// Get current node
//	node, _ := manager.GetCurrentNode("player1", arc.ID)
//	// Use node.Title and node.Description in your UI
//
//	// Make a choice
//	if node.Type == branching.NodeTypeChoice {
//		choice := node.Choices[0]
//		manager.MakeChoice("player1", arc.ID, choice.ID)
//	}
//
//	// Check alignment
//	alignment, _ := manager.GetAlignment("player1", arc.ID)
//	// Use alignment values for narrative branching
//
// # Story Arc Structure
//
// Story arcs consist of multiple layers of nodes:
//
//	Start Node → Choice/Event Nodes → Choice/Event Nodes → Ending Nodes
//
// Each choice affects:
// - Moral alignment (Good/Evil, Law/Chaos, Honor/Dishonor)
// - Faction reputation (genre-specific factions)
// - Story variables (custom data for quest tracking)
// - Available paths (some nodes have requirements)
//
// # Alignment System
//
// Three alignment axes tracked independently:
//
// - AlignmentGoodEvil: Moral righteousness vs. selfishness (-1.0 to 1.0)
// - AlignmentLawChaos: Order vs. freedom (-1.0 to 1.0)
// - AlignmentHonorDishonor: Integrity vs. deceit (-1.0 to 1.0)
//
// Each choice can shift alignment by -0.2 to +0.2 on each axis.
//
// # Faction System
//
// Genre-specific factions:
//
// - Fantasy: The Order, The Dark Guild, Merchant's League, Arcane Circle
// - Sci-Fi: Federation, Rebel Alliance, Corporate Syndicate, AI Coalition
// - Horror: The Survivors, The Cult, The Resistance, The Lost
// - Cyberpunk: MegaCorp, The Underground, Net Runners, Street Gangs
// - Post-Apocalyptic: Vault Dwellers, Raiders, Traders, Mutants
//
// Reputation ranges from -1.0 (hostile) to 1.0 (allied).
//
// # Performance
//
// Target metrics:
// - Story generation: <500ms per arc (10-20 nodes)
// - Choice processing: <10ms
// - Narrative memory: <5MB per player (1000 choices)
//
// All generation is deterministic (same seed = identical story arc).
//
// # Determinism Exception
//
// This package uses time.Now() for progress tracking timestamps:
//   - StartTime: when a player begins a story arc (manager.go)
//   - LastUpdate: when a player makes a choice or advances (manager.go)
//
// These are non-procgen metadata used for analytics and debugging only.
// They do not affect story generation, choice outcomes, or any gameplay logic.
// Story arc generation and choice processing remain fully deterministic.
//
// # ECS Integration
//
// For integration with the engine's ECS:
//
//	component := &branching.NarrativeComponent{
//		ActiveArcs: []string{arcID},
//		Progress:   make(map[string]*branching.PlayerProgress),
//	}
//	entity.AddComponent(component)
//
// The component tracks active story arcs and player progress for each entity.
package branching
