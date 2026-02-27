// Package building provides procedural generation of building structures with floor plans.
//
// # Overview
//
// The building package generates procedural buildings with complete floor plans, including
// room layouts, door placement, window distribution, and roof types. All generation is
// deterministic and uses seed-based RNG for reproducibility.
//
// # Building Types
//
// Five building types are supported:
//   - House: Small residential buildings with 2-4 rooms
//   - Workshop: Work areas with main shop and storage (2-3 rooms)
//   - Storage: Large storage facilities (1-2 large rooms)
//   - Tower: Vertical structures with stacked floors (2-8 floors)
//   - Manor: Large estates with 4-8 rooms in grid layouts
//
// # Architectural Styles
//
// Each genre provides 5 distinct architectural styles:
//
// Fantasy: Medieval, Elven, Dwarven, WizardTower, Village
// Sci-Fi: Modular, Brutalist, Organic, Geometric, Crystalline
// Horror: Gothic, Decayed, Asylum, Mansion, Crypt
// Cyberpunk: Neon, Industrial, Corporate, Underground, Megastructure
// Post-Apocalyptic: Salvage, Bunker, Ruins, Fortified, Scrapyard
//
// # Usage Example
//
//	gen := building.NewGenerator()
//	params := procgen.GenerationParams{
//	    GenreID: "fantasy",
//	    Custom: map[string]interface{}{
//	        "buildingType": building.TypeManor,
//	    },
//	}
//
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//	    logrus.WithError(err).Fatal("building generation failed")
//	}
//
//	building := result.(*building.Building)
//	// Note: Production code should use logrus.WithFields instead of fmt.Printf
//	fmt.Printf("Generated %s with %d rooms\n", building.Type, len(building.Rooms))
//
// # Floor Plan Generation
//
// Floor plans are procedurally generated with the following constraints:
//   - All rooms must be accessible from the entrance (100% navigability)
//   - Rooms cannot overlap
//   - At least one entrance room required
//   - Maximum 8 rooms per building
//   - Dimensions between 4x4 and 64x64 tiles
//
// # Layout Algorithms
//
// Different building types use specialized layout algorithms:
//   - Houses: Horizontal subdivision
//   - Workshops: Dedicated workshop area with smaller entrance/storage
//   - Storage: Single large room or binary split
//   - Towers: Vertical stacking with stairs
//   - Manors: Grid-based multi-room layouts
//
// # Validation
//
// All generated buildings undergo validation:
//   - Dimension checks (4-64 tiles per axis)
//   - Room count limits (1-8 rooms)
//   - Entrance requirement verification
//   - Navigability testing (BFS connectivity check)
//   - Overlap detection for rooms
//
// # Performance
//
// Generation targets:
//   - <100ms per building (typical: 0.01-0.5ms)
//   - 100% navigable floor plans
//   - 85%+ genre recognition by players
//
// # Determinism
//
// Buildings generated with the same seed and parameters will be identical:
//
//	building1, _ := gen.Generate(42, params)
//	building2, _ := gen.Generate(42, params)
//	// building1 and building2 are identical
//
// # Integration
//
// Buildings are designed for integration with:
//   - Housing system (pkg/world/housing/)
//   - Guild halls (pkg/network/federation/guild/)
//   - Territory control (pkg/world/territory/)
//   - Furniture generator (pkg/procgen/furniture/)
package building
