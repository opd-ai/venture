// Package story provides procedural generation of environmental story fragments.
//
// Environmental storytelling creates discoverable narratives through fragments
// scattered in the game world: notes, carvings, corpses, relics, graffiti, and
// blood trails. Fragments form coherent story arcs when discovered in sequence.
//
// # Fragment Types
//
// The system generates 6 types of story fragments:
//   - Notes: Written notes, journals, letters
//   - Carvings: Wall inscriptions, ancient texts
//   - Corpses: Bodies with clues (items, wounds)
//   - Relics: Ancient artifacts with histories
//   - Graffiti: Recent markings, warnings
//   - Blood Trails: Physical evidence of events
//
// # Story Generation
//
// Each dungeon receives a procedural story with:
//   - Beginning: Setup and introduction
//   - Middle: Development and complications
//   - End: Resolution or cliffhanger
//   - 5-15 fragments placed throughout
//
// # Usage Example
//
//	gen := story.NewFragmentGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      5,
//	    GenreID:    "fantasy",
//	}
//	result, err := gen.Generate(12345, params)
//	fragments := result.(*story.StorySequence)
//
// # Determinism
//
// All generation is deterministic based on dungeon seed. Same seed + parameters
// produces identical story content and fragment placements.
//
// # Performance
//
// - Generation time: <20ms per story
// - Memory: <50KB per fragment
// - Coherence: Validated via grammar rules
package story
