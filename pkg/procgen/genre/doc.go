// Package genre provides a centralized genre definition system for the Venture game.
//
// The genre system manages all game genres (fantasy, sci-fi, horror, etc.) and provides
// a consistent interface for genre-based content generation across all procedural
// generation systems.
//
// # Features
//
//   - Predefined genre definitions with metadata
//   - Genre validation and lookup by ID
//   - Color palette associations
//   - Theme descriptions and naming conventions
//   - Easy extension for new genres
//
// # Usage
//
//	// Get the default genre registry
//	registry := genre.DefaultRegistry()
//
//	// Look up a genre
//	fantasy, err := registry.Get("fantasy")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Validate a genre ID
//	if !registry.Has("fantasy") {
//	    log.Fatal("Fantasy genre not found")
//	}
//
//	// List all available genres
//	genres := registry.All()
//	for _, g := range genres {
//	    fmt.Printf("%s: %s\n", g.ID, g.Name)
//	}
//
// # Supported Genres
//
// The system comes with the following predefined genres:
//
//   - Fantasy: Traditional medieval fantasy with magic and monsters
//   - Sci-Fi: Science fiction with technology and aliens
//   - Horror: Dark, atmospheric horror themes
//   - Cyberpunk: High-tech dystopian future
//   - Post-Apocalyptic: Wasteland survival themes
//
// # Adding New Genres
//
// To add a new genre, create a Genre struct and register it:
//
//	steampunk := &genre.Genre{
//	    ID:          "steampunk",
//	    Name:        "Steampunk",
//	    Description: "Victorian-era technology and aesthetics",
//	    Themes:      []string{"industrial", "clockwork", "steam"},
//	}
//	registry.Register(steampunk)
package genre
