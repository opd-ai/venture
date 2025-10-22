// Package genre provides a centralized genre definition system for the Venture game.
//
// The genre system manages all game genres (fantasy, sci-fi, horror, etc.) and provides
// a consistent interface for genre-based content generation across all procedural
// generation systems. It supports both predefined genres and cross-genre blending
// for hybrid content.
//
// # Features
//
//   - Predefined genre definitions with metadata
//   - Genre validation and lookup by ID
//   - Color palette associations
//   - Theme descriptions and naming conventions
//   - Cross-genre blending for hybrid genres
//   - Preset blended genre combinations
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
// # Genre Blending
//
// Create hybrid genres by blending two base genres:
//
//	blender := genre.NewGenreBlender(registry)
//	scifiHorror, err := blender.Blend("scifi", "horror", 0.5, 12345)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Blended genre: %s\n", scifiHorror.Name)
//
// Use preset blends for common combinations:
//
//	darkFantasy, err := blender.CreatePresetBlend("dark-fantasy", 12345)
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
// Preset blends include: sci-fi-horror, dark-fantasy, cyber-horror,
// post-apoc-scifi, and wasteland-fantasy.
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
