Create a PLAN.md document that outlines a comprehensive, prioritized plan to optimize the Venture game engine's performance to eliminate visually apparent sluggishness and maintain the 60 FPS target. The plan should:

1. **Assessment Phase**: Identify current performance bottlenecks through profiling
   - CPU profiling of game loop, rendering pipeline, and generation systems
   - Memory profiling to detect allocation hotspots and potential leaks
   - Frame time analysis to identify lag spikes and frame drops
   - Network profiling for multiplayer bandwidth and latency impact

2. **Prioritized Optimization Tasks**: Organize improvements by impact and effort
   - Critical path optimizations (game loop, rendering, collision detection)
   - Hot path improvements (entity queries, component access patterns)
   - Memory allocation reduction (object pooling, buffer reuse)
   - Spatial partitioning optimizations (quadtree/grid efficiency)
   - Procedural generation caching strategies (maintaining determinism)
   - Network optimization (delta compression, culling, prediction accuracy)

3. **Implementation Strategy**: Define concrete steps with measurable targets
   - Specific code changes with file paths and function names
   - Before/after performance metrics (FPS, memory, bandwidth)
   - Testing procedures to validate improvements
   - Regression prevention measures

4. **Validation Criteria**: How to measure success
   - Consistent 60 FPS during typical gameplay scenarios
   - Frame time variance <16.67ms (60 FPS target)
   - Memory usage under 500MB for client
   - Network bandwidth under 100KB/s per player
   - Generation time under 2s for world areas

Include references to existing profiling tools (`go test -cpuprofile`, `-memprofile`), benchmark tests, and the Phase 8.4 (Performance Optimization) requirements from the project roadmap.