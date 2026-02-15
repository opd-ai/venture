// Package engine provides a creature visual classification component.
// CreatureVisualComponent stores the visual body form, size class, and
// descriptive tags that the animation/sprite pipeline uses to select the
// correct nonhumanoid aerial template for each entity.
package engine

import "strings"

// CreatureForm identifies the body plan used for sprite template selection.
type CreatureForm string

const (
	FormHumanoid   CreatureForm = "humanoid"
	FormQuadruped  CreatureForm = "quadruped"
	FormArachnid   CreatureForm = "arachnid"
	FormSerpentine CreatureForm = "serpentine"
	FormFlying     CreatureForm = "flying"
	FormBlob       CreatureForm = "blob"
	FormMechanical CreatureForm = "mechanical"
	FormUndead     CreatureForm = "undead"
)

// CreatureVisualComponent is a pure-data ECS component that carries visual
// classification data from entity spawning through to the sprite generator.
// It bridges the gap between procgen entity Tags/Names and the rendering
// pipeline's nonhumanoid template selector.
type CreatureVisualComponent struct {
	// Form is the creature body plan (humanoid, quadruped, arachnid, etc.).
	Form CreatureForm
	// SizeClass mirrors procgen EntitySize for sprite scaling.
	SizeClass string
	// VisualTags are genre/template hints (e.g. "undead", "robotic", "armored").
	VisualTags []string
}

// Type implements the Component interface.
func (c *CreatureVisualComponent) Type() string { return "creature_visual" }

// creatureFormKeywords maps keywords found in entity names and tags to forms.
var creatureFormKeywords = map[string]CreatureForm{
	// Quadrupeds
	"wolf": FormQuadruped, "bear": FormQuadruped, "hound": FormQuadruped,
	"dog": FormQuadruped, "horse": FormQuadruped, "boar": FormQuadruped,
	"lion": FormQuadruped, "tiger": FormQuadruped, "beast": FormQuadruped,
	"cat": FormQuadruped, "stag": FormQuadruped, "deer": FormQuadruped,
	"rat": FormQuadruped, "fox": FormQuadruped,
	// Arachnids / insects
	"spider": FormArachnid, "scorpion": FormArachnid, "insect": FormArachnid,
	"beetle": FormArachnid, "ant": FormArachnid, "centipede": FormArachnid,
	"crawler": FormArachnid, "tick": FormArachnid,
	// Serpentine
	"snake": FormSerpentine, "serpent": FormSerpentine, "worm": FormSerpentine,
	"tentacle": FormSerpentine, "wyrm": FormSerpentine, "naga": FormSerpentine,
	"eel": FormSerpentine, "leech": FormSerpentine,
	// Flying
	"dragon": FormFlying, "bat": FormFlying, "bird": FormFlying,
	"wyvern": FormFlying, "harpy": FormFlying, "gargoyle": FormFlying,
	"hawk": FormFlying, "eagle": FormFlying, "phoenix": FormFlying,
	// Blobs / amorphous
	"slime": FormBlob, "ooze": FormBlob, "blob": FormBlob,
	"amoeba": FormBlob, "jelly": FormBlob, "pudding": FormBlob,
	"thing": FormBlob, "abomination": FormBlob,
	// Mechanical
	"robot": FormMechanical, "golem": FormMechanical, "construct": FormMechanical,
	"android": FormMechanical, "mech": FormMechanical, "drone": FormMechanical,
	"bot": FormMechanical, "probe": FormMechanical, "turret": FormMechanical,
	"sentinel": FormMechanical,
	// Undead
	"skeleton": FormUndead, "zombie": FormUndead, "ghost": FormUndead,
	"lich": FormUndead, "wraith": FormUndead, "revenant": FormUndead,
	"corpse": FormUndead, "shade": FormUndead, "spectre": FormUndead,
	"ghoul": FormUndead, "shadow": FormUndead,
}

// ClassifyCreatureForm determines the creature body form from an entity name
// and a set of descriptive tags. It scans tags first (higher confidence), then
// name words. Returns FormHumanoid when no better classification is found.
func ClassifyCreatureForm(name string, tags []string) CreatureForm {
	// Tags take priority — they come from curated template definitions.
	for _, tag := range tags {
		lower := strings.ToLower(tag)
		if form, ok := creatureFormKeywords[lower]; ok {
			return form
		}
	}

	// Broad tag categories that map directly.
	for _, tag := range tags {
		lower := strings.ToLower(tag)
		switch {
		case lower == "undead" || lower == "horrifying":
			return FormUndead
		case lower == "robotic" || lower == "mechanical":
			return FormMechanical
		case lower == "wild" || lower == "feral" || lower == "mutant":
			return FormQuadruped
		case lower == "augmented":
			return FormHumanoid
		case lower == "human":
			return FormHumanoid
		}
	}

	// Scan individual words in the entity name.
	words := strings.Fields(strings.ToLower(name))
	for _, w := range words {
		if form, ok := creatureFormKeywords[w]; ok {
			return form
		}
	}

	return FormHumanoid
}
