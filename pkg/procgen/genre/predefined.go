// Package genre provides predefined genre definitions.
// This file contains constructors for all standard game genres.
// Code relocated from: types.go
package genre

import "math/rand"

// PredefinedGenres returns a slice of all predefined genre definitions.
func PredefinedGenres() []*Genre {
	return []*Genre{
		FantasyGenre(),
		SciFiGenre(),
		HorrorGenre(),
		CyberpunkGenre(),
		PostApocalypticGenre(),
	}
}

// FantasyGenre returns the Fantasy genre definition.
func FantasyGenre() *Genre {
	return &Genre{
		ID:             "fantasy",
		Name:           "Fantasy",
		Description:    "Traditional medieval fantasy with magic, dragons, and ancient mysteries",
		Themes:         []string{"medieval", "magic", "dragons", "knights", "wizards", "dungeons"},
		PrimaryColor:   "#8B4513", // Saddle Brown
		SecondaryColor: "#DAA520", // Goldenrod
		AccentColor:    "#4169E1", // Royal Blue
		EntityPrefix:   "Ancient",
		ItemPrefix:     "Enchanted",
		LocationPrefix: "The",
	}
}

// SciFiGenre returns the Science Fiction genre definition.
func SciFiGenre() *Genre {
	return &Genre{
		ID:             "scifi",
		Name:           "Sci-Fi",
		Description:    "Science fiction with advanced technology, space exploration, and alien encounters",
		Themes:         []string{"technology", "space", "aliens", "robots", "lasers", "future"},
		PrimaryColor:   "#00CED1", // Dark Turquoise
		SecondaryColor: "#7B68EE", // Medium Slate Blue
		AccentColor:    "#00FF00", // Lime
		EntityPrefix:   "Prototype",
		ItemPrefix:     "Advanced",
		LocationPrefix: "Station",
	}
}

// HorrorGenre returns the Horror genre definition.
func HorrorGenre() *Genre {
	return &Genre{
		ID:             "horror",
		Name:           "Horror",
		Description:    "Dark, atmospheric horror with supernatural threats and psychological terror",
		Themes:         []string{"dark", "supernatural", "undead", "cursed", "twisted", "nightmare"},
		PrimaryColor:   "#8B0000", // Dark Red
		SecondaryColor: "#2F4F4F", // Dark Slate Gray
		AccentColor:    "#9370DB", // Medium Purple
		EntityPrefix:   "Cursed",
		ItemPrefix:     "Twisted",
		LocationPrefix: "The Haunted",
	}
}

// CyberpunkGenre returns the Cyberpunk genre definition.
func CyberpunkGenre() *Genre {
	return &Genre{
		ID:             "cyberpunk",
		Name:           "Cyberpunk",
		Description:    "High-tech dystopian future with cybernetic enhancements and corporate dominance",
		Themes:         []string{"cybernetic", "neon", "corporate", "hacker", "augmented", "dystopian"},
		PrimaryColor:   "#FF1493", // Deep Pink
		SecondaryColor: "#00FFFF", // Cyan
		AccentColor:    "#FFD700", // Gold
		EntityPrefix:   "Augmented",
		ItemPrefix:     "Cyber",
		LocationPrefix: "Neo",
	}
}

// PostApocalypticGenre returns the Post-Apocalyptic genre definition.
func PostApocalypticGenre() *Genre {
	return &Genre{
		ID:             "postapoc",
		Name:           "Post-Apocalyptic",
		Description:    "Wasteland survival in a world devastated by catastrophe",
		Themes:         []string{"wasteland", "survival", "scavenged", "mutated", "ruined", "barren"},
		PrimaryColor:   "#CD853F", // Peru
		SecondaryColor: "#696969", // Dim Gray
		AccentColor:    "#FF6347", // Tomato
		EntityPrefix:   "Mutated",
		ItemPrefix:     "Salvaged",
		LocationPrefix: "Ruins of",
	}
}

// GetTheme retrieves a genre theme by ID.
// Returns the genre for the given ID, or defaults to Fantasy if not found.
// Special value "random" triggers random genre selection using the provided seed.
func GetTheme(genreID string) *Genre {
	// Handle special "random" value - use 0 seed for backwards compatibility
	// Callers should use GetThemeWithSeed for deterministic random selection
	if genreID == "random" {
		return GetRandomTheme(0)
	}

	registry := DefaultRegistry()
	genre, err := registry.Get(genreID)
	if err != nil {
		// Default to fantasy if genre not found
		return FantasyGenre()
	}
	return genre
}

// GetThemeWithSeed retrieves a genre theme by ID with explicit seed for random selection.
// Returns the genre for the given ID, or defaults to Fantasy if not found.
// When genreID is "random", uses the seed for deterministic selection.
func GetThemeWithSeed(genreID string, seed int64) *Genre {
	if genreID == "random" {
		return GetRandomTheme(seed)
	}

	registry := DefaultRegistry()
	genre, err := registry.Get(genreID)
	if err != nil {
		return FantasyGenre()
	}
	return genre
}

// GetRandomTheme returns a random genre from predefined genres using deterministic seed.
// Same seed always returns the same genre, ensuring reproducible procedural generation.
func GetRandomTheme(seed int64) *Genre {
	genres := PredefinedGenres()
	rng := rand.New(rand.NewSource(seed))
	idx := rng.Intn(len(genres))
	return genres[idx]
}
