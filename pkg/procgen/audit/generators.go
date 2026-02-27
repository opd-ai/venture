package audit

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/book"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/legendary"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
)

// GeneratorEntry represents a generator with its name and base parameters.
// Centralizes generator list to prevent desync between determinism and edge case tests.
type GeneratorEntry struct {
	Name      string
	Generator procgen.Generator
	Params    procgen.GenerationParams
}

// GetAllGenerators returns the complete list of generators for audit testing.
// This is the single source of truth for which generators are tested.
// Add new generators here to ensure they're tested in all audit scenarios.
func GetAllGenerators() []GeneratorEntry {
	baseParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	bookParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type": engine.BookTypeLore,
		},
	}

	return []GeneratorEntry{
		{"Entity", entity.NewEntityGenerator(), baseParams},
		{"Item", item.NewItemGenerator(), baseParams},
		{"Magic", magic.NewSpellGenerator(), baseParams},
		{"Skills", skills.NewSkillTreeGenerator(), baseParams},
		{"Quest", quest.NewQuestGenerator(), baseParams},
		{"Recipe", recipe.NewRecipeGenerator(), baseParams},
		{"Station", station.NewStationGenerator(), baseParams},
		{"Terrain", terrain.NewBSPGenerator(), baseParams},
		{"Vehicle", vehicle.NewVehicleGenerator(), baseParams},
		{"Companion", companion.NewGenerator(), baseParams},
		{"Building", building.NewGenerator(), baseParams},
		{"Furniture", furniture.NewGenerator(), baseParams},
		{"Legendary", legendary.NewLegendaryQuestGenerator(), baseParams},
		{"Book", book.NewGenerator(), bookParams},
	}
}
