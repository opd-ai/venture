// Package ui provides procedural UI element generation for the Venture game.
//
// This package generates user interface elements using mathematical algorithms and
// genre-based visual themes. All UI generation is deterministic and follows the
// game's procedural design philosophy.
//
// # UI Element Types
//
// The package supports several types of UI elements:
//   - Button: Interactive buttons with genre-appropriate styling
//   - Panel: Container panels for UI sections
//   - HealthBar: Progress bars for health, mana, stamina
//   - Label: Text labels with backgrounds
//   - Icon: Small iconic UI elements
//   - Frame: Decorative frames and borders
//
// # Basic Usage
//
//	gen := ui.NewGenerator()
//	config := ui.Config{
//	    Type:    ui.ElementButton,
//	    Width:   100,
//	    Height:  30,
//	    GenreID: "fantasy",
//	    Seed:    12345,
//	    Text:    "Start Game",
//	}
//	element, err := gen.Generate(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Genre-Aware Styling
//
// UI elements automatically adapt to different game genres:
//   - Fantasy: Ornate borders, warm colors, medieval styling
//   - Sci-Fi: Clean lines, neon accents, futuristic look
//   - Horror: Dark tones, rough textures, ominous feel
//   - Cyberpunk: Glowing edges, high contrast, tech aesthetic
//   - Post-Apocalyptic: Worn textures, muted colors, gritty style
//
// # Performance
//
// UI generation is optimized for runtime creation with typical generation
// times under 1ms per element. Elements can be cached and reused for
// better performance in the game loop.
package ui
