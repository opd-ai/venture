package guild

import (
	"fmt"
	"math/rand"
)

// Guild procedural identity generation.
//
// This file provides deterministic procedural generation of guild names and emblems
// based on genre and seed. Supports fantasy, sci-fi, horror, cyberpunk, and
// post-apocalyptic themes with appropriate naming conventions and visual elements.
//
// Code relocated from: manager.go

// GuildIdentity represents a procedurally generated guild identity
// Originally defined in: manager.go
type GuildIdentity struct {
	Name   string
	Emblem *Emblem
}

// GenerateIdentity generates a procedural guild name and emblem
// Originally defined in: manager.go
func GenerateIdentity(genre string, seed int64) GuildIdentity {
	rng := rand.New(rand.NewSource(seed))

	// Genre-specific name templates
	var prefixes, suffixes []string
	switch genre {
	case "fantasy":
		prefixes = []string{"The Ancient", "The Noble", "The Shadow", "The Crimson", "The Golden"}
		suffixes = []string{"Knights", "Guardians", "Brotherhood", "Order", "Legion"}
	case "sci-fi":
		prefixes = []string{"The Stellar", "The Nova", "The Quantum", "The Void", "The Nexus"}
		suffixes = []string{"Collective", "Syndicate", "Federation", "Alliance", "Corporation"}
	case "horror":
		prefixes = []string{"The Twisted", "The Cursed", "The Hollow", "The Forsaken", "The Damned"}
		suffixes = []string{"Cult", "Covenant", "Circle", "Cabal", "Congregation"}
	case "cyberpunk":
		prefixes = []string{"The Neon", "The Chrome", "The Digital", "The Neural", "The Cyber"}
		suffixes = []string{"Runners", "Hackers", "Collective", "Network", "Syndicate"}
	case "post-apocalyptic":
		prefixes = []string{"The Wasteland", "The Rad", "The Scrap", "The Dust", "The Rust"}
		suffixes = []string{"Survivors", "Scavengers", "Raiders", "Nomads", "Brotherhood"}
	default:
		prefixes = []string{"The", "Order of", "Guild of", "Company of", "League of"}
		suffixes = []string{"Adventurers", "Explorers", "Warriors", "Traders", "Builders"}
	}

	name := fmt.Sprintf("%s %s", prefixes[rng.Intn(len(prefixes))], suffixes[rng.Intn(len(suffixes))])

	// Generate emblem
	shapes := []string{"shield", "crest", "banner", "circle", "star"}
	symbols := []string{"sword", "dragon", "star", "flame", "skull", "crown", "hammer", "phoenix"}

	emblem := &Emblem{
		Shape:      shapes[rng.Intn(len(shapes))],
		PrimaryR:   uint8(rng.Intn(256)),
		PrimaryG:   uint8(rng.Intn(256)),
		PrimaryB:   uint8(rng.Intn(256)),
		SecondaryR: uint8(rng.Intn(256)),
		SecondaryG: uint8(rng.Intn(256)),
		SecondaryB: uint8(rng.Intn(256)),
		Symbol:     symbols[rng.Intn(len(symbols))],
	}

	return GuildIdentity{Name: name, Emblem: emblem}
}
