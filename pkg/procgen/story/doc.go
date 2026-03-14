// Package story provides procedural generation of environmental story fragments
// and advanced narrative systems including branching narratives, cross-dungeon
// story arcs, historical timelines, and genre-specific archaeology.
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
// # Advanced Narrative Systems
//
// ## Branching Narratives
//
// BranchingNarrativeGenerator creates stories with player choice points:
//
//   - 1-3 choice points per narrative
//
//   - 2-8 unique story paths based on decisions
//
//   - Binary choices at each decision point
//
//   - MakeChoice() processes player decisions
//
//   - GetActivePath() retrieves current narrative state
//
//     bnGen := story.NewBranchingNarrativeGenerator()
//     result, _ := bnGen.Generate(seed, params)
//     narrative := result.(*story.BranchingNarrative)
//     narrative.MakeChoice(0, 1) // Choose option 1 at first choice point
//     activePath := narrative.GetActivePath()
//
// ## Cross-Dungeon Stories
//
// CrossDungeonGenerator creates narratives spanning multiple dungeon levels:
//
//   - Stories span 2-5 dungeon levels
//
//   - Prerequisite system ensures proper progression
//
//   - Fragments unlock as player progresses through dungeons
//
//   - IsFragmentAccessible() checks if content is unlocked
//
//   - GetFragmentsForLevel() retrieves level-specific content
//
//     cdGen := story.NewCrossDungeonGenerator()
//     result, _ := cdGen.Generate(seed, params)
//     cdStory := result.(*story.CrossDungeonStory)
//     fragments := cdStory.GetFragmentsForLevel(3)
//     if cdStory.IsFragmentAccessible(5, []int{1, 2, 3}) {
//     // Fragment 5 is accessible
//     }
//
// ## Historical Timelines
//
// TimelineGenerator creates world history for lore consistency:
//
//   - 2-5 historical eras per world
//
//   - 10+ events spanning world history
//
//   - 8 event types: Foundation, War, Discovery, Catastrophe, Renaissance, Collapse, Contact, Ritual
//
//   - GetEventsInPeriod() queries events by time range
//
//   - GetEventsByType() queries events by category
//
//   - GetCurrentEra() identifies present-day era
//
//     tlGen := story.NewTimelineGenerator()
//     result, _ := tlGen.Generate(seed, params)
//     timeline := result.(*story.Timeline)
//     recentEvents := timeline.GetEventsInPeriod(800, 1000)
//     wars := timeline.GetEventsByType(story.EventTypeWar)
//
// ## Genre-Specific Archaeology
//
// ArchaeologyGenerator creates discoverable archaeological sites:
//
//   - 2-6 artifacts per site
//
//   - 6 artifact types: Magical, Technology, Ritual, Data, Pre-War, Relic
//
//   - Genre-specific content (fantasy: ancient magic, sci-fi: alien ruins, horror: dark rituals)
//
//   - Excavate() progress mechanics (0.0-1.0 completion)
//
//   - Curse/trap generation for danger (0.0-1.0 rating)
//
//     archGen := story.NewArchaeologyGenerator()
//     result, _ := archGen.Generate(seed, params)
//     site := result.(*story.ArchaeologicalSite)
//     site.Excavate(0.25) // 25% excavation progress
//     if site.IsFullyExcavated() {
//     // All artifacts recovered
//     }
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
// produces identical story content and fragment placements across all systems:
// fragments, branching narratives, cross-dungeon stories, timelines, and archaeology.
//
// # Performance
//
// - Generation time: <20ms per story (actual: <0.03ms for all systems)
// - Memory: <50KB per fragment (actual: <20KB for all systems)
// - Coherence: Validated via grammar rules and consistency checks
//
// # Benchmarks
//
// Run story benchmarks with:
//
//	go test -bench=. -benchmem ./pkg/procgen/story/...
//
// # Genre Support
//
// All systems support five core genres with themed content:
//   - Fantasy: Ancient magic, dragons, medieval lore, mystical artifacts
//   - Sci-Fi: Alien civilizations, advanced technology, space exploration, AI
//   - Horror: Dark rituals, cursed objects, supernatural events, eldritch horrors
//   - Cyberpunk: Corporate conspiracies, neural implants, digital archives, hackers
//   - Post-Apocalyptic: Pre-war relics, survival struggles, wasteland ruins, mutations
package story
