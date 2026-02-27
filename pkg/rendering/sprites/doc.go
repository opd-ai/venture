// Package sprites provides procedural sprite generation for game entities.
// Sprites are created by combining shapes, colors, and procedural patterns
// to create unique visual representations without requiring asset files.
//
// # Sprite Generation Modes
//
// The package supports two perspective modes:
//
// 1. Side-View (Legacy): Vertical orientation for side-scrolling gameplay
// 2. Aerial-View (New): Top-down orientation for overhead camera gameplay
//
// # Basic Sprite Generation
//
// Generate a single sprite:
//
//	gen := sprites.NewGenerator()
//	result, err := gen.Generate(seed, procgen.GenerationParams{
//	    GenreID: "fantasy",
//	    Custom: map[string]interface{}{
//	        "width":  32,
//	        "height": 32,
//	        "type":   "monster",
//	    },
//	})
//	sprite := result.(*sprites.Sprite)
//
// # Directional Sprite Generation
//
// Generate 4-directional sprites for top-down gameplay:
//
//	gen := sprites.NewGenerator()
//	config := sprites.GenerationConfig{
//	    Width:      32,
//	    Height:     32,
//	    Seed:       12345,
//	    GenreID:    "fantasy",
//	    EntityType: "humanoid",
//	    UseAerial:  true,  // Enable aerial-view perspective
//	}
//
//	sprites, err := gen.GenerateDirectionalSprites(config)
//	if err != nil {
//	    // Production code should use: logrus.WithError(err).Error("failed to generate sprites")
//	    log.Fatal(err)
//	}
//
//	// sprites is map[Direction]*ebiten.Image
//	// Access by direction: sprites[DirUp], sprites[DirDown], etc.
//
// # Aerial-View Templates
//
// Aerial templates provide top-down character perspectives with consistent
// 35/50/15 proportions (head/torso/legs):
//
//	// Base template
//	template := sprites.HumanoidAerial()
//
//	// Genre-specific templates
//	fantasyTemplate := sprites.FantasyHumanoidAerial()
//	scifiTemplate := sprites.SciFiHumanoidAerial()
//	horrorTemplate := sprites.HorrorHumanoidAerial()
//	cyberpunkTemplate := sprites.CyberpunkHumanoidAerial()
//	postapocTemplate := sprites.PostApocalypticHumanoidAerial()
//
//	// Phase 15.1 enhanced template with pixel-perfect dimensions
//	enhancedTemplate := sprites.EnhancedHumanoidTemplate()
//
// # Enhanced Templates (Phase 15.1)
//
// Phase 15.1 introduces enhanced templates with pixel-perfect anatomical specifications:
//
//	// Use enhanced template for improved clarity and recognition
//	template := sprites.EnhancedHumanoidTemplate()
//	// Head: 4×4 pixels, Torso: 4×6 pixels, Legs: 4×8 pixels
//	// 40% more anatomical detail, better silhouette scores
//
//	// Use detailed template for facial features in close-up views
//	detailedTemplate := sprites.DetailedHumanoidTemplate()
//	// Includes all enhanced features PLUS:
//	// Eyes: 2×1 pixels, Mouth: 2×1 pixels
//	// Perfect for player characters and important NPCs
//
// # Phase 45: Enhanced 64x64 Sprite Templates
//
// Phase 45 introduces high-detail 64x64 sprite templates for 1920x1080 resolution.
// These templates use aerial/top-down proportions consistent with all other templates:
// head ~35%, torso ~50%, legs ~15% (the camera looks straight down).
// Target silhouette recognition: 0.85+ (up from 0.75).
//
//	// Automatically select appropriate template based on sprite size
//	template := sprites.SelectTemplate64("humanoid", "fantasy", 64, true)
//	// Returns Enhanced64HumanoidTemplate with detailed facial features
//
//	// Manual selection for specific needs
//	template := sprites.Enhanced64HumanoidTemplate()
//	// Head: dominant from above, Torso: wide shoulders, Legs: barely visible
//	// Arms: 12×10 pixels for wider reach and articulation
//
//	// Detailed variant with facial features and secondary details
//	detailedTemplate := sprites.Detailed64HumanoidTemplate()
//	// Includes all enhanced features PLUS:
//	// Eyes: 4×2 pixels, Mouth: 4×2 pixels
//	// Perfect for player characters at high resolution
//
//	// Other 64x64 creature templates
//	quadruped := sprites.Enhanced64QuadrupedTemplate()
//	// Head: 10×12 pixels, Torso: 20×14 pixels, Legs: 20×8 pixels, Tail: 8×16 pixels
//
//	blob := sprites.Enhanced64BlobTemplate()
//	// Torso: 32×28 pixels (large amorphous mass), Eyes: 6×4 pixels (nucleus)
//
//	mechanical := sprites.Enhanced64MechanicalTemplate()
//	// Head: 10×10 pixels (cubic sensor), Torso: 12×18 pixels (chassis)
//	// Arms: 14×10 pixels, Legs: 10×14 pixels
//
// Enhanced templates provide exact pixel dimensions that remain constant regardless
// of sprite size, ensuring consistent visual quality and improved player recognition.
// Detailed templates add facial features for emotional expression and close-up clarity.
//
// # Boss Scaling
//
// Scale any aerial template for boss entities while preserving proportions:
//
//	baseTemplate := sprites.FantasyHumanoidAerial()
//	bossTemplate := sprites.BossAerialTemplate(baseTemplate, 2.5)
//
//	config := sprites.GenerationConfig{
//	    Width:     64,
//	    Height:    64,
//	    Template:  &bossTemplate,
//	    UseAerial: true,
//	}
//	sprites, err := gen.GenerateDirectionalSprites(config)
//
// # Using with Movement System
//
// The movement system automatically updates entity facing direction based
// on velocity. The render system then displays the correct directional sprite:
//
//	// In your render loop:
//	anim, ok := entity.GetComponent("animation")
//	if ok {
//	    animation := anim.(*engine.AnimationComponent)
//	    currentSprite := directionalSprites[animation.Facing]
//	    screen.DrawImage(currentSprite, opts)
//	}
//
// No manual direction handling required - the integration is automatic!
//
// # Enhanced Proportional Scaling (Phase 15.1)
//
// The package supports pixel-perfect anatomical specifications for enhanced detail:
//
//	// Create body parts with exact pixel dimensions
//	head := sprites.NewPartSpecFromPixels(4, 4, shapes.ShapeCircle, 15, "secondary")
//	torso := sprites.NewPartSpecFromPixels(4, 6, shapes.ShapeRectangle, 10, "primary")
//	legs := sprites.NewPartSpecFromPixels(4, 8, shapes.ShapeCapsule, 5, "primary")
//
//	// Dimensions are exact regardless of sprite size
//	width := head.GetEffectiveWidth(28)   // Returns 4 pixels
//	height := head.GetEffectiveHeight(32) // Returns 4 pixels
//
//	// Upgrade existing templates with pixel dimensions
//	template := sprites.HumanoidTemplate()
//	headSpec := template.BodyPartLayout[sprites.PartHead]
//	enhancedHead := headSpec.WithPixelDimensions(4, 4)
//
// Pixel dimensions enable Phase 15.1 "head 4×4, torso 4×6, legs 4×8" specifications
// for improved anatomical accuracy and visual clarity. When PreferredPixelSize is set,
// it takes precedence over relative dimensions, enabling pixel-perfect control.
//
// Both relative and absolute sizing are supported - existing templates using relative
// dimensions continue to work unchanged.
//
// # Genre-Specific Anatomical Variations (Phase 15.1)
//
// The package provides genre-specific variations for all creature templates,
// automatically adapting visual style to match the genre aesthetic:
//
//	// Automatic genre-aware template selection
//	template := sprites.SelectTemplateWithGenre("quadruped", "fantasy")
//	// Returns fantasy_quadruped with organic shapes (bean, ellipse, capsule)
//
//	template = sprites.SelectTemplateWithGenre("mechanical", "scifi")
//	// Returns scifi_mechanical with geometric shapes (hexagon, octagon, rectangle)
//
//	template = sprites.SelectTemplateWithGenre("undead", "horror")
//	// Returns horror_undead with distorted proportions and reduced shadow
//
// # Genre Variation Functions
//
// Each genre applies specific visual transformations:
//
// Fantasy (Organic): Softer, natural shapes
//
//	fantasy := sprites.ApplyFantasyVariation(baseTemplate)
//	// Prefers: organic, bean, ellipse, circle, capsule, wave
//	// Replaces: rectangle→capsule, hexagon→ellipse, triangle→wedge
//
// Sci-Fi (Geometric): Angular, precise shapes
//
//	scifi := sprites.ApplySciFiVariation(baseTemplate)
//	// Prefers: hexagon, octagon, rectangle, triangle, gear, crystal
//	// Replaces: organic→hexagon, ellipse→octagon, circle→hexagon
//
// Horror (Distorted): Unsettling proportions
//
//	horror := sprites.ApplyHorrorVariation(baseTemplate)
//	// Elongates head (height ×1.2, width ×0.85)
//	// Elongates limbs (height ×1.15, width ×0.85)
//	// Reduces shadow opacity (×0.6) for ghostly effect
//	// Prefers: skull, organic shapes with irregular torso
//
// Cyberpunk (Augmented): Tech enhancements
//
//	cyberpunk := sprites.ApplyCyberpunkVariation(baseTemplate)
//	// Prefers: hexagon, octagon, rectangle (angular tech shapes)
//	// Adds: tech armor overlay with translucent glow (opacity 0.4)
//	// Modifies: head uses accent1 color role (neon glow)
//
// Post-Apocalyptic (Weathered): Rough, damaged appearance
//
//	postapoc := sprites.ApplyPostApocVariation(baseTemplate)
//	// Prefers: organic, rectangle, capsule (rough shapes)
//	// Replaces: circle→organic, ellipse→bean, hexagon→rectangle
//
// # Performance
//
// Genre variations are extremely fast, suitable for runtime application:
//
//	ApplyFantasyVariation:  671 ns/op (1,200 B, 10 allocs)
//	ApplySciFiVariation:    760 ns/op (1,248 B, 13 allocs)
//	ApplyHorrorVariation:   554 ns/op (1,128 B,  5 allocs)
//	SelectTemplateWithGenre: 1.2 µs/op (2,336 B, 16 allocs)
//
// # Determinism
//
// All genre variations are deterministic - the same base template
// always produces identical variations. This ensures:
//
// - Consistent visuals across clients in multiplayer
// - Reproducible results for testing and debugging
// - Cache-friendly behavior for sprite generation
//
// # Direction Enum
//
// Direction constants for 4-directional facing:
//
//	DirUp    = 0  // North, moving upward
//	DirDown  = 1  // South, moving downward
//	DirLeft  = 2  // West, moving left
//	DirRight = 3  // East, moving right
//
// # UseAerial Flag
//
// The UseAerial flag in GenerationConfig controls perspective mode:
//
//	config := sprites.GenerationConfig{
//	    UseAerial: true,   // Top-down aerial view (recommended for top-down gameplay)
//	    UseAerial: false,  // Side-view profile (legacy; all default templates now use aerial)
//	}
//
// When UseAerial is true:
// - Uses aerial-view anatomical templates
// - Maintains 35/50/15 proportions (head/torso/legs)
// - Optimized for top-down camera angles
// - Directional asymmetry for visual clarity
//
// When UseAerial is false:
// - Uses side-view templates
// - Traditional vertical proportions
// - Suitable for side-scrolling gameplay
//
// # Performance Characteristics
//
// Sprite generation is optimized for runtime efficiency:
//
// - 4-sprite generation: ~172 µs (0.172 ms)
// - Template creation: 455-662 ns/op
// - Memory per 4-sprite sheet: ~121 KB
// - Direction switching: <5 ns overhead
//
// Sprites are generated once per entity and cached. Direction switching
// uses simple map lookups with negligible performance impact.
package sprites
