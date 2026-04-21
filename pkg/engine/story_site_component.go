package engine

import "github.com/opd-ai/venture/pkg/procgen/story"

// ArchaeologicalSiteComponent marks an entity as an excavatable archaeological site.
// Spawned by spawnArchaeologicalSites; pending runtime consumer integration
// (excavation progression, discovery UI) before spawning is re-enabled.
type ArchaeologicalSiteComponent struct {
	// Site holds all generated data for the archaeological site.
	Site *story.ArchaeologicalSite
	// Discovered is true once the player has found the site.
	Discovered bool
}

// Type returns the component type identifier.
func (a *ArchaeologicalSiteComponent) Type() string { return "archaeo_site" }

// TimelineComponent attaches a world-historical timeline to an entity.
// Typically attached to a single dedicated entity at world initialisation;
// pending runtime consumer integration before spawning is re-enabled.
type TimelineComponent struct {
	// Timeline holds the generated world history.
	Timeline *story.Timeline
}

// Type returns the component type identifier.
func (t *TimelineComponent) Type() string { return "timeline" }

// CrossDungeonStoryComponent attaches a cross-dungeon narrative arc to an entity.
// The story spreads fragments across multiple dungeon levels;
// pending runtime consumer integration before spawning is re-enabled.
type CrossDungeonStoryComponent struct {
	// Story holds the generated cross-dungeon narrative.
	Story *story.CrossDungeonStory
	// Discovered is true once the player has found at least one fragment.
	Discovered bool
}

// Type returns the component type identifier.
func (c *CrossDungeonStoryComponent) Type() string { return "crossdungeon" }
