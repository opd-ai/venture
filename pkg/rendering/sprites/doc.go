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
// Enhanced templates provide exact pixel dimensions that remain constant regardless
// of sprite size, ensuring consistent visual quality and improved player recognition.
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
// Full backward compatibility maintained - existing templates using relative
// dimensions continue to work unchanged.
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
//	    UseAerial: true,   // Top-down aerial view (recommended)
//	    UseAerial: false,  // Side-view (legacy, default)
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
