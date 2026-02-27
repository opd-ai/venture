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
//	    // Production code should use: logrus.WithError(err).Error("failed to generate UI element")
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
// # Visual Hierarchy (Phase 19.1)
//
// The package supports four hierarchy levels for organizing UI elements by importance:
//   - Primary: Most important elements (titles, main actions) - largest, most prominent
//   - Secondary: Important content (section headers, key info) - standard emphasis
//   - Tertiary: Supporting content (details, descriptions) - reduced emphasis
//   - Quaternary: Minimal emphasis (footnotes, hints) - subtle presentation
//
// Example:
//
//	config := ui.Config{
//	    Type:           ui.ElementButton,
//	    HierarchyLevel: ui.HierarchyPrimary,
//	    Width:          200,
//	    Height:         60,
//	    GenreID:        "fantasy",
//	}
//	button, err := gen.Generate(config)
//
// # UI Transitions (Phase 19.1)
//
// Animated transitions provide smooth visual effects:
//   - Fade: Alpha-based fade in/out
//   - Slide: Directional slide animations (left, right, up, down)
//   - Zoom: Scale-based zoom effects
//
// Supports multiple easing functions:
//   - Linear: Constant speed
//   - EaseInQuad/OutQuad/InOutQuad: Quadratic acceleration/deceleration
//   - EaseInCubic/OutCubic/InOutCubic: Cubic acceleration/deceleration
//
// Example:
//
//	transition := ui.TransitionConfig{
//	    Type:     ui.TransitionFade,
//	    Duration: 300, // milliseconds
//	    Easing:   ui.EaseInOutQuad,
//	    Progress: 0.5, // 0.0 to 1.0
//	}
//	config := ui.Config{
//	    Type:       ui.ElementPanel,
//	    Transition: &transition,
//	    // ... other config
//	}
//
// # Visual Separators (Phase 19.1)
//
// Generate visual separators for dividing UI sections:
//   - Line: Simple horizontal line
//   - Dashed: Dashed line pattern
//   - Dotted: Dotted line pattern
//   - Gradient: Gradient fade effect
//   - Ornamental: Decorative separator with patterns
//
// Example:
//
//	separator := gen.GenerateSeparator(400, 10, ui.SeparatorGradient, "fantasy", seed)
//
// # Group Containers (Phase 19.1)
//
// Create visual grouping containers with configurable borders and backgrounds:
//
//	groupConfig := ui.GroupConfig{
//	    Width:          300,
//	    Height:         200,
//	    Level:          ui.HierarchySecondary,
//	    GenreID:        "fantasy",
//	    ShowBorder:     true,
//	    ShowBackground: true,
//	}
//	container := gen.GenerateGroupContainer(groupConfig)
//
// # Enhanced Border Styles (Phase 19.1)
//
// Eight border styles available:
//   - Solid: Simple solid border
//   - Double: Double-line border
//   - Ornate: Decorative border with corner embellishments
//   - Glow: Gradient glow effect
//   - Dashed: Dashed pattern
//   - Dotted: Dotted pattern
//   - Embossed: 3D raised effect
//   - Engraved: 3D recessed effect
//
// Border styles are automatically selected based on genre, or can be
// specified manually through config.Custom["borderStyle"].
//
// # Performance
//
// UI generation is optimized for runtime creation with typical generation
// times under 1ms per element. Elements can be cached and reused for
// better performance in the game loop.
//
// Phase 19.1 enhancements maintain performance targets:
//   - Generation time: <1ms per element
//   - Transition overhead: <0.1ms per frame
//   - Memory usage: Minimal (sub-KB per element)
//   - Test coverage: 92.6%
package ui
