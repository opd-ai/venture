// Package tiles provides procedural tile image generation for terrain rendering.
//
// The tiles package generates visual representations of terrain tiles using
// procedural techniques. It supports multiple tile types (floor, wall, door,
// corridor) and can generate genre-specific visual styles.
//
// Features:
//   - Deterministic tile generation using seeds
//   - Genre-aware styling using color palettes
//   - Pattern variations for visual diversity
//   - Integration with terrain generation
//   - Configurable tile sizes
//   - Phase 16.2: Smooth terrain transitions with auto-tiling (Marching Squares)
//   - Edge blending between different tile types
//   - Corner rounding for organic feel
//   - Edge smoothing for visual polish
//   - Phase 16.3: Parallax depth effects for 3D perception
//   - Multi-layer rendering (background, base, foreground)
//   - Parallax scrolling based on camera position
//   - Ambient occlusion for corners and edges
//   - Height-based shadow casting
//   - Phase 47: Enhanced wall rendering for 1920x1080 resolution
//   - 2x2 super-sampling anti-aliasing for smooth edges
//   - Seamless corner blending (L/T/Cross junctions)
//   - Wall/floor boundary blending (50/50 color mix)
//   - Directional shadow gradients for depth
//
// Phase 16.2 Transition System:
//
// The transition system implements Marching Squares algorithm for seamless
// tile connections. It analyzes 8-directional neighbors and generates
// 47 unique tile variants including:
//   - Single edge connections (4 types)
//   - Adjacent corners (4 types)
//   - Opposite corridors (2 types)
//   - T-junctions (4 types)
//   - Inner corners (4 types)
//   - Full connections and isolated tiles
//
// The system provides gradient blending at tile boundaries, corner rounding
// for walls, and edge smoothing for organic appearance. All transitions are
// deterministic and maintain performance targets (<3% frame time increase).
//
// Phase 16.3 Parallax Depth System:
//
// The parallax system creates depth perception through multi-layer rendering.
// Three layers are supported:
//   - Background: Furthest layer, moves slower (parallax depth 0.2-0.4)
//   - Base: Main tile content, moves with camera (parallax depth 1.0)
//   - Foreground: Closest layer, moves faster (parallax depth 1.2-1.5)
//
// Each layer can have ambient occlusion applied to darken corners and edges,
// and height-based shadows for enhanced 3D effect. Parallax offset is calculated
// based on camera position and layer-specific depth multipliers.
//
// Phase 47 Enhanced Wall Rendering:
//
// The enhanced wall system provides high-quality wall rendering optimized for
// 1920x1080 resolution. Key features include:
//
//   - Anti-aliasing: 2x2 super-sampling downsampling eliminates jagged edges
//   - Corner blending: Automatic detection and blending for L, T, and Cross junctions
//   - Boundary blending: 50/50 color mixing between wall and floor tiles
//   - Shadow gradients: Directional lighting creates depth perception
//   - Configurable blend radius: Default 4px for smooth corners
//
// Corner detection analyzes neighboring wall tiles to determine junction types:
//   - L corners: Two adjacent walls meeting at 90 degrees
//   - T junctions: Three walls meeting (one stem, two branches)
//   - Cross junctions: Four walls meeting (intersection)
//
// All effects are deterministic and maintain performance targets:
//   - <0.5ms per tile generation time
//   - No jagged pixels at 1920x1080
//   - Seamless corner transitions
//
// Performance targets:
//   - <5% frame time increase over base rendering
//   - All effects maintain deterministic generation
//   - Ambient occlusion: <1ms overhead
//   - Shadow generation: <1ms overhead
//   - Layer compositing: <0.5ms for 32x32 tile
//   - Enhanced wall rendering: <0.5ms for 32x32, <1.5ms for 64x64
//
// Example Usage:
//
//	gen := tiles.NewGenerator()
//	config := tiles.Config{
//	    Type:    tiles.TileFloor,
//	    Width:   32,
//	    Height:  32,
//	    GenreID: "fantasy",
//	    Seed:    12345,
//	}
//	tileImg, err := gen.Generate(config)
//	if err != nil {
//	    // Production code should use: logrus.WithError(err).Error("failed to generate tile")
//	    log.Fatal(err)
//	}
//
// Example with Transitions:
//
//	neighbors := tiles.TileNeighbors{N: true, E: true, S: true}
//	transitionType := tiles.DetermineTransition(neighbors)
//
//	transConfig := tiles.TransitionConfig{
//	    BaseConfig:   config,
//	    Transition:   transitionType,
//	    Neighbors:    neighbors,
//	    BlendRadius:  0.3,
//	    CornerRadius: 0.25,
//	    Smoothness:   0.5,
//	}
//	transitionTile, err := gen.GenerateWithTransition(transConfig)
//	if err != nil {
//	    // Production code should use: logrus.WithError(err).Error("failed to generate transition")
//	    log.Fatal(err)
//	}
//
// Example with Parallax Depth:
//
//	parallaxConfig := tiles.ParallaxConfig{
//	    BaseConfig:    config,
//	    Layer:         tiles.LayerBase,
//	    CameraX:       10.0,
//	    CameraY:       5.0,
//	    ParallaxDepth: 1.0,
//	    AOIntensity:   0.5,
//	    ShadowHeight:  0.3,
//	    ShadowAngle:   math.Pi / 4,
//	}
//	parallaxTile, err := gen.GenerateWithParallax(parallaxConfig)
//	if err != nil {
//	    // Production code should use: logrus.WithError(err).Error("failed to generate parallax tile")
//	    log.Fatal(err)
//	}
//
// Example with All Three Layers:
//
//	bg, base, fg, err := gen.GenerateLayeredTile(config, cameraX, cameraY)
//	if err != nil {
//	    // Production code should use: logrus.WithError(err).Error("failed to generate layered tile")
//	    log.Fatal(err)
//	}
//	// Render layers separately with parallax offsets, or composite:
//	composite := tiles.CompositeLayers(bg, base, fg)
//
// Example with Enhanced Wall Rendering (Phase 47):
//
//	wallConfig := tiles.DefaultEnhancedWallConfig()
//	wallConfig.Config.Width = 64
//	wallConfig.Config.Height = 64
//	wallConfig.Config.Seed = 12345
//	wallConfig.Config.GenreID = "fantasy"
//	wallConfig.Neighbors = tiles.WallNeighbors{North: true, East: true}
//	wallConfig.EnableAntialiasing = true
//	wallConfig.EnableShadows = true
//	wallConfig.BlendRadius = 4
//
//	wallTile, err := gen.GenerateEnhancedWall(wallConfig)
//	if err != nil {
//	    // Production code should use: logrus.WithError(err).Error("failed to generate wall")
//	    log.Fatal(err)
//	}
package tiles
