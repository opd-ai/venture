// Package audit provides production-readiness validation for Phase 62: Procedural Generation Audit
//
// This package implements comprehensive testing for all procedural generators to ensure
// they meet production quality standards for determinism, quality thresholds, edge case handling,
// performance, and memory efficiency.
//
// # Phase 62.1: Generator Determinism Validation
//
// Validates that all generators produce deterministic output:
//   - Same seed → identical output (100% determinism in 1000 runs)
//   - Different seeds → varied output (>80% variation)
//   - Seed derivation → no collisions (<0.01% collision rate across 1M seeds)
//   - Platform consistency → Linux/macOS/Windows/WASM produce identical JSON
//   - Version stability → v10.0 output matches v9.0 for migration testing
//
// # Tested Generators (14 total)
//
//   - TerrainGenerator: Dungeon layouts with rooms and corridors
//   - EntityGenerator: Monsters, NPCs, bosses with stats and behaviors
//   - ItemGenerator: Weapons, armor, consumables with rarity tiers
//   - MagicGenerator: Spells and abilities with elemental types
//   - SkillGenerator: Skill trees with prerequisites and unlocks
//   - QuestGenerator: Quest objectives with rewards
//   - RecipeGenerator: Crafting recipes with requirements
//   - StationGenerator: Crafting stations with capabilities
//   - VehicleGenerator: Mounts and vehicles with physics stats
//   - CompanionGenerator: Pets and followers with AI behaviors
//   - BuildingGenerator: Procedural buildings with floor plans
//   - FurnitureGenerator: Furniture items with placement rules
//   - LegendaryGenerator: Legendary items with unique powers
//   - BookGenerator: In-game books with procedural content
//
// Note: EnvironmentGenerator (pkg/procgen/environment) uses a different API
// (Config-based via GenerateFromConfig(config Config) instead of the standard
// seed/params pattern Generate(seed int64, params GenerationParams)), making it
// incompatible with the generic generator audit framework used here.
// It is therefore excluded from this audit suite.
//
// # Usage Example
//
//	// Run all determinism tests
//	go test -v ./pkg/procgen/audit -run TestDeterminism
//
//	// Run acceptance test (1000 runs per generator, ~5-10 min)
//	go test -v ./pkg/procgen/audit -run TestDeterminism_AcceptanceCriteria
//
//	// Run with race detection
//	go test -race ./pkg/procgen/audit
//
// # Acceptance Criteria
//
//   - 100% determinism: Zero failures in 1000 runs per generator
//   - Seed collision rate: <0.01% across 1M generated seeds
//   - Cross-platform: Exact same JSON output on all platforms
//   - Version stability: v10.0 compatible with v9.0 seed output
//
// # Test Organization
//
// Tests are organized by Phase 62.1 requirements:
//   - TestDeterminism_SameSeedProducesIdenticalOutput: Requirement #1 (100% determinism)
//   - TestDeterminism_DifferentSeedsProduceVariedOutput: Requirement #2 (>80% variation)
//   - TestDeterminism_SeedDerivationNonCollision: Requirement #3 (<0.01% collision)
//   - TestDeterminism_PlatformConsistency: Requirement #4 (cross-platform)
//   - TestDeterminism_VersionStability: Requirement #5 (v9.0 → v10.0)
//   - TestDeterminism_AcceptanceCriteria_1000Runs: Full acceptance test (long-running)
//
// # Performance Characteristics
//
//   - Test suite runtime: ~30-60 seconds (100 runs per generator × 14 generators)
//   - Acceptance test runtime: ~5-10 minutes (1000 runs × 14 generators)
//   - Memory usage: <100MB peak (hash storage for variation tests)
//   - Concurrency: All tests run in parallel via t.Parallel()
//
// # Future Work (Phase 62.2+)
//
// Quality threshold validation, edge case generation, performance benchmarks,
// memory leak detection will be added in subsequent audit phases.
package audit
