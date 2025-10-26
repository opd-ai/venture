Create a comprehensive PLAN.md document that outlines a systematic approach to enhance the visual fidelity and recognizability of procedurally-generated sprites and graphical elements in the Venture game engine. The objective is to improve the representational quality of generated content so that characters, items, and entities are more visually identifiable and resemble their intended forms.

**Primary Goals:**

1. **Character Anatomy Enhancement**: Develop algorithms to generate more recognizable character sprites with distinguishable anatomical features including head, torso, arms, and legs. Use top-down perspective conventions to create approximate but clear body part representations.

2. **Strict Constraint Compliance**: All player character sprites must remain exactly 28x28 pixels. Entity sprites should adhere to their existing dimensional constraints. Solutions must work within current memory and performance budgets (60 FPS, <500MB client memory).

3. **Procedural Realism Balance**: Acknowledge that perfect photorealistic representation is unattainable with pure procedural generation. Focus on achieving "good enough" visual clarity—sprites should be immediately recognizable as their entity type at a glance.

**Required Analysis:**

- Survey current sprite generation pipeline (`pkg/rendering/sprites/`) and identify specific visual ambiguity issues
- Evaluate existing shape primitives (`pkg/rendering/shapes/`) and their limitations for anatomical representation
- Review genre-specific color palette constraints that may affect visibility
- Assess performance impact of more complex generation algorithms

**Implementation Strategy:**

The plan should propose incremental improvements organized into phases:

- **Template Enhancement**: Refine entity templates with more detailed body part specifications
- **Shape Primitive Expansion**: Add new procedural primitives for common anatomical features (heads, limbs, torsos)
- **Layered Composition**: Implement sprite layer system for overlaying body parts, equipment, and effects
- **Silhouette Definition**: Ensure sprites have clear, readable silhouettes distinguishable from backgrounds
- **Genre-Appropriate Styling**: Maintain genre aesthetic consistency (fantasy vs. sci-fi vs. horror) while improving clarity

**Deliverables:**

- Prioritized list of visual improvements with effort estimates
- Technical approach for each enhancement (algorithm modifications, new components)
- Testing criteria to validate improved recognizability
- Performance benchmarks to ensure optimizations don't degrade frame rate
- Example mockups or pseudocode demonstrating key techniques

The plan should be actionable, technically feasible within the existing ECS architecture, maintain deterministic seed-based generation, and preserve the project's zero-external-assets philosophy.