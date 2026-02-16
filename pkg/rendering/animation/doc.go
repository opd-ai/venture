// Package animation provides advanced animation fluidity system for Phase 46.
//
// This package implements 8-frame animations with 8-direction support,
// body part articulation, animation caching, and pre-computation capabilities.
//
// Key Features:
//   - 8-frame animation cycles (increased from 4 frames)
//   - 8-direction support (N, NE, E, SE, S, SW, W, NW)
//   - Body part articulation (arms ±3px, legs ±4px)
//   - Animation frame caching with LRU eviction
//   - Pre-computation of common animation sequences
//   - Sub-frame interpolation for smoother rendering
//
// Performance Targets:
//   - 8 frames/cycle at 60 FPS
//   - ≥85% cache hit rate
//   - <1ms per frame generation
//   - 60 FPS with 100 animated entities
//
// Usage:
//
//	// Create animation controller
//	controller := animation.NewController(spriteGenerator)
//
//	// Pre-compute common animations
//	controller.PrecomputeCommon(entitySeeds)
//
//	// Update animations each frame
//	controller.Update(entities, deltaTime)
//
// Integration with ECS:
//
// The animation system integrates with the engine via AnimationAdapter
// (pkg/engine/animation_adapter.go), which wraps animation.Controller as
// a System-level adapter. The adapter provides:
//   - On-demand frame generation with articulation and direction support
//   - LRU-cached animation frames via AnimationCache
//   - Feature toggle via SetEnabled/IsEnabled for optional activation
//
// The adapter pattern allows enhanced animation features without modifying
// the core AnimationSystem or requiring additional entity-level components.
package animation
