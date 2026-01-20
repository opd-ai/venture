package story

// Fragment Type Constants
// Originally from: generator.go

const (
	FragmentNote FragmentType = iota
	FragmentCarving
	FragmentCorpse
	FragmentRelic
	FragmentGraffiti
	FragmentBlood
)

// Artifact Type Constants
// Originally from: archaeology.go

const (
	ArtifactMagical ArtifactType = iota // Fantasy: magical items, enchantments
	ArtifactTech                        // Sci-Fi: alien/ancient technology
	ArtifactRitual                      // Horror: cursed objects, ritual items
	ArtifactData                        // Cyberpunk: data crystals, memory chips
	ArtifactPreWar                      // Post-Apocalyptic: pre-fall artifacts
	ArtifactRelic                       // Generic: historical relics
)

// Event Type Constants
// Originally from: timeline.go

const (
	EventFoundation  EventType = iota // Founding of civilizations/factions
	EventWar                          // Conflicts and battles
	EventDiscovery                    // Scientific/magical discoveries
	EventCatastrophe                  // Disasters and calamities
	EventRenaissance                  // Cultural/technological advances
	EventCollapse                     // Fall of civilizations
	EventContact                      // First contact with new species/factions
	EventRitual                       // Major magical/religious ceremonies
)
