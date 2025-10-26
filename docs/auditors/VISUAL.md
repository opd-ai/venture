**Objective**: Design a plan(in PLAN.md) to enhance the procedural character avatar generation system with two critical improvements: (1) directional facing indicators and (2) migration from side-view to aerial-view perspective. Ensure that all implemented improvements are made available in the main game.

**Execution Mode**: Report generation - produce a detailed implementation plan for review, do not execute changes.

**Current Issues**:
1. No visual representation of character facing direction (N/S/E/W)
2. Side-view perspective incompatible with top-down gameplay camera

**Required Deliverables**:

1. **Architecture Analysis** (200 words max):
   - Review `pkg/rendering/sprites/` avatar generation system
   - Identify affected components and dependencies
   - Note integration points with movement/rendering systems

2. **Technical Design** (300 words max):
   - Directional sprite variants (4 or 8 directions)
   - Aerial-view template specifications (body shapes, proportions)
   - Sprite sheet layout or rotation strategy
   - Performance impact assessment

3. **Implementation Roadmap** (300 words max):
   - Ordered task list with file paths
   - Testing strategy (visual validation, determinism checks)
   - Backward compatibility considerations
   - Estimated effort per task

4. **Quality Improvements** (200 words max):
   - Consistency enhancements (anatomy proportions, color coherence)
   - Genre-specific aerial templates
   - Animation frame considerations

**Success Criteria**:
- Maintains seed-based determinism
- Supports 4 cardinal directions minimum
- Aerial perspective matches game camera
- No performance regression
- Passes existing sprite tests

**Constraints**:
- Must work with current ECS architecture
- Zero external assets (procedural only)
- Maintain <65ms generation time per sprite

Output as structured markdown with file references and code patterns where relevant.