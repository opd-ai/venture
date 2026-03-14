// BossNameGenerator generates procedural raid and boss names based on genre.
// This file creates thematically appropriate names for raids and bosses using
// genre-specific prefixes, suffixes, and title templates.
package raids

import (
	"fmt"
	"math/rand"
)

// BossNameGenerator generates procedural boss and raid names.
type BossNameGenerator struct{}

// NewBossNameGenerator creates a new boss name generator.
func NewBossNameGenerator() *BossNameGenerator {
	return &BossNameGenerator{}
}

// GenerateRaidName creates a raid dungeon name based on genre and tier.
func (b *BossNameGenerator) GenerateRaidName(rng *rand.Rand, genreID string, tier RaidTier) string {
	prefixes := b.getPrefixesByGenre(genreID)
	suffixes := b.getSuffixesByGenre(genreID)

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	tierName := tier.String()

	return fmt.Sprintf("%s %s (%s)", prefix, suffix, tierName)
}

// GenerateBlendedRaidName creates a raid name that blends two genres.
// primaryID and secondaryID are genre identifiers; blendWeight (0.0–1.0) controls
// how much influence the secondary genre has. At blendWeight=0.0 the result is
// equivalent to GenerateRaidName(primaryID); at 1.0 it is equivalent to
// GenerateRaidName(secondaryID). Intermediate values draw prefixes from the
// primary genre and suffixes from the secondary genre (or vice-versa depending
// on rng), creating hybrid names that reflect the blended theme.
func (b *BossNameGenerator) GenerateBlendedRaidName(rng *rand.Rand, primaryID, secondaryID string, blendWeight float64, tier RaidTier) string {
	// Clamp blend weight to valid range.
	if blendWeight < 0 {
		blendWeight = 0
	}
	if blendWeight > 1 {
		blendWeight = 1
	}

	// Choose prefix pool: primary unless rng exceeds (1 - blendWeight).
	var prefix string
	prefPrimary := b.getPrefixesByGenre(primaryID)
	prefSecondary := b.getPrefixesByGenre(secondaryID)
	if rng.Float64() < blendWeight {
		prefix = prefSecondary[rng.Intn(len(prefSecondary))]
	} else {
		prefix = prefPrimary[rng.Intn(len(prefPrimary))]
	}

	// Choose suffix pool: secondary unless rng falls below (1 - blendWeight).
	var suffix string
	sufPrimary := b.getSuffixesByGenre(primaryID)
	sufSecondary := b.getSuffixesByGenre(secondaryID)
	if rng.Float64() < blendWeight {
		suffix = sufSecondary[rng.Intn(len(sufSecondary))]
	} else {
		suffix = sufPrimary[rng.Intn(len(sufPrimary))]
	}

	return fmt.Sprintf("%s %s (%s)", prefix, suffix, tier.String())
}

// getPrefixesByGenre returns genre-specific name prefixes.
func (b *BossNameGenerator) getPrefixesByGenre(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{
			"Ancient", "Cursed", "Forgotten", "Eternal", "Dark",
			"Shadow", "Crimson", "Obsidian", "Crystal", "Infernal",
		}
	case "scifi":
		return []string{
			"Stellar", "Quantum", "Cyber", "Neural", "Plasma",
			"Nexus", "Void", "Particle", "Fusion", "Orbital",
		}
	case "horror":
		return []string{
			"Haunted", "Twisted", "Corrupted", "Nightmare", "Cursed",
			"Rotting", "Bleeding", "Screaming", "Forsaken", "Damned",
		}
	case "cyberpunk":
		return []string{
			"Corporate", "Neon", "Chrome", "Digital", "Syndicate",
			"Underground", "Encrypted", "Rogue", "Black Market", "Grid",
		}
	case "postapoc":
		return []string{
			"Wasteland", "Irradiated", "Ruined", "Scorched", "Mutant",
			"Survivor", "Raider", "Bunker", "Fallout", "Desolate",
		}
	default:
		return []string{
			"Ancient", "Forgotten", "Lost", "Hidden", "Secret",
		}
	}
}

// getSuffixesByGenre returns genre-specific name suffixes.
func (b *BossNameGenerator) getSuffixesByGenre(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{
			"Citadel", "Sanctum", "Crypt", "Temple", "Fortress",
			"Ruins", "Catacombs", "Stronghold", "Throne", "Vault",
		}
	case "scifi":
		return []string{
			"Station", "Laboratory", "Reactor", "Array", "Complex",
			"Facility", "Nexus", "Chamber", "Core", "Platform",
		}
	case "horror":
		return []string{
			"Asylum", "Manor", "Morgue", "Pit", "Abyss",
			"Sepulcher", "Charnel House", "Depths", "Lair", "Den",
		}
	case "cyberpunk":
		return []string{
			"Tower", "Datacenter", "Hub", "Spire", "Network",
			"Vault", "Exchange", "Terminal", "Node", "Mainframe",
		}
	case "postapoc":
		return []string{
			"Bunker", "Shelter", "Outpost", "Stronghold", "Compound",
			"Ruins", "Crater", "Wasteland", "Zone", "Silo",
		}
	default:
		return []string{
			"Dungeon", "Ruins", "Fortress", "Temple", "Vault",
		}
	}
}

// GenerateBossName creates a boss name.
func (b *BossNameGenerator) GenerateBossName(rng *rand.Rand, genreID string, index int) string {
	titles := b.getTitlesByGenre(genreID)
	names := b.getNamesByGenre(genreID)

	title := titles[rng.Intn(len(titles))]
	name := names[rng.Intn(len(names))]

	return fmt.Sprintf("%s, the %s", name, title)
}

// getTitlesByGenre returns genre-specific boss titles.
func (b *BossNameGenerator) getTitlesByGenre(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{
			"Ancient One", "Destroyer", "Eternal Guardian", "Dark Lord",
			"Shadow King", "Flame Warden", "Ice Queen", "Void Monarch",
		}
	case "scifi":
		return []string{
			"AI Overlord", "Quantum Entity", "Neural Matrix", "Cybernetic Titan",
			"Plasma Lord", "Void Commander", "Core Sentinel", "System Admin",
		}
	case "horror":
		return []string{
			"Devourer", "Tortured Soul", "Nightmare Incarnate", "Flesh Sculptor",
			"Blood Hunter", "Corpse Lord", "Undying Horror", "Screamer",
		}
	case "cyberpunk":
		return []string{
			"Corporate Enforcer", "Netrunner Prime", "Chrome Tyrant", "Data Thief",
			"Street King", "Syndicate Boss", "Grid Master", "Black ICE",
		}
	case "postapoc":
		return []string{
			"Wasteland Warlord", "Mutant King", "Raider Chief", "Survivor Prime",
			"Bunker Lord", "Scavenger Boss", "Radiation Master", "Vault Guardian",
		}
	default:
		return []string{
			"Ancient One", "Guardian", "Lord", "Destroyer",
		}
	}
}

// getNamesByGenre returns genre-specific boss base names.
func (b *BossNameGenerator) getNamesByGenre(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{
			"Malakar", "Thalor", "Zephyra", "Drakthar", "Norvath",
			"Sylvara", "Grimthorn", "Asheron", "Valdis", "Eryndor",
		}
	case "scifi":
		return []string{
			"Unit-7X9", "Nexus-Prime", "Cipher-Omega", "CoreMind-Alpha",
			"Synth-Delta", "Grid-Zeta", "Quantum-Phi", "Neural-Sigma",
		}
	case "horror":
		return []string{
			"The Flenser", "The Weeping", "The Gnawer", "The Rotting",
			"The Wailing", "The Hunger", "The Darkness", "The Screamer",
		}
	case "cyberpunk":
		return []string{
			"Chrome-Jack", "Neon-Viper", "Data-Reaper", "Grid-Phantom",
			"Synth-Killer", "Net-Wraith", "Code-Breaker", "Ice-Runner",
		}
	case "postapoc":
		return []string{
			"Rad-King", "Scorch", "Rust", "Fallout", "Ash",
			"Wastrel", "Decay", "Ruin", "Blight", "Cinder",
		}
	default:
		return []string{
			"Unknown", "Nameless", "Forgotten", "Ancient",
		}
	}
}
