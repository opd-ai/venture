You are implementing ONE novel enhancement to a Go/Ebiten procedural multiplayer action-RPG called Venture. Act autonomously. Do not ask for approval.

KNOWN AVATAR PROBLEMS (read this first — these are the project's highest priority):
The current avatars are deeply broken. Every improvement you make should address one or more of these:
1. **WRONG PERSPECTIVE**: Sprites are rendered in profile/side view but the game uses a top-down camera. All entity sprites MUST be drawn as seen from above (aerial/overhead view). The head/shoulders should dominate the sprite; legs should be barely visible beneath the body. The default HumanoidTemplate() is wrong — it gives legs 48% of the sprite height, which is a profile view. Use and improve HumanoidAerialTemplate() proportions instead (head ~35%, torso ~50%, legs ~15%). Fix any sprite generation that draws entities as if viewed from the side.
2. **INSUFFICIENT DETAIL**: Sprites are visually barren — flat colors, no shading, no texture, no personality. At 32×32 every pixel matters. Add sub-pixel shading, color gradients, dithering, highlight/shadow on body parts, hair detail, clothing patterns, anything that makes sprites look crafted rather than placeholder.
3. **INSUFFICIENT VARIETY**: All avatars look nearly identical. Different NPCs, different players, different creature types should be immediately distinguishable at a glance. Vary body proportions, color palettes, head shapes, clothing silhouettes, and accessories. Seed-based generation should produce visually diverse output, not minor variations on one template.
4. **POOR NONHUMANOID REPRESENTATION**: Creatures, monsters, animals, and bosses use barely-modified humanoid templates. A spider should not look like a person. A dragon should not look like a person. Build and use dedicated nonhumanoid anatomy templates — quadrupeds, insects, serpents, amorphous blobs, winged creatures, multi-limbed horrors. Each creature type needs its own distinct body plan visible from above.

STEP 1 — DISCOVER (spend ≤5 minutes here):
- Run `git log --oneline -20` to avoid duplicating recent work.
- Read pkg/engine/system_init.go to understand registered systems.
- Grep for TODO, FIXME, stub, placeholder in pkg/engine/ and pkg/procgen/.
- Pick ONE enhancement you have NOT seen in git history. Roll a dice (1/20) to decide the category:
  - **Avatar overhaul (pick 19/20 of the time — address the KNOWN AVATAR PROBLEMS above):**
    - **Perspective fixes** — convert any profile/side-view sprites to proper top-down aerial view. This is the single most impactful fix.
    - **Nonhumanoid templates** — build dedicated top-down anatomy templates for creature types that are not humanoid (quadrupeds, insects, serpents, flying creatures, amorphous entities, multi-limbed creatures). Every creature type deserves its own body plan.
    - **Player character visuals** — composite layering, anatomy detail, directional sprites, proportions, body shapes, facial features, skin/hair color variety, idle poses, shading, clothing detail
    - **NPC variety** — genre-aware body templates, size-based anatomy, silhouette quality, visual personality, distinctive appearance per NPC, varied clothing and coloring
    - **Equipment visuals** — material rendering fidelity, damage-state degradation, enchantment glow/particles, rarity-based detail scaling, weapon silhouettes, armor shaping
    - **Sprite detail** — sub-pixel shading, color gradients, dithering, material textures, highlight/shadow, edge definition, anti-aliasing
    - **Animation improvements** — smoother transitions, new states, expressive movement, attack/cast/hurt animations, idle breathing/fidget
  - **Other systems (pick 1/20 of the time, avoid visual changes):**
    - Connecting two existing systems that don't yet interact
    - Adding depth to a system with minimal/placeholder logic
    - New gameplay mechanics, customization
    - Genre-aware variation in procgen outputs
- If multiple candidates exist, pick the one that most improves avatar quality. Perspective fixes and nonhumanoid templates take priority over everything else.

STEP 2 — IMPLEMENT (this is the bulk of the work):
Follow these rules strictly. Violations are build failures.

Architecture:
- Components = pure data + `Type() string`. Zero methods with logic.
- Systems = all logic. Signature: `Update(entities []*Entity, deltaTime float64)`.
- Procgen: `rand.New(rand.NewSource(seed))` only. Never global rand. Never time.Now().
- Logging: `logrus.WithFields(logrus.Fields{"system_name": "...", ...})`.
- No external assets. No new dependencies beyond go.mod.

Visual & Animation:
- Lighting must use radial gradients with proper falloff (linear, quadratic, inverse-square). No flat circles.
- Shadows use soft penumbra with distance-based falloff. Support genre-specific opacity presets.
- Post-processing effects (color grading, vignette, chromatic aberration) must be genre-aware.
- Animation playback: 12 FPS (0.083s per frame), 8 frames per state. Use distance-based LOD (full rate at ≤200px, half at ≤400px, minimal beyond).
- Sprite generation must be seeded and cached (LRU, max 100 entries). Pool image buffers by size bucket.
- All visual enhancements must maintain 60+ FPS. Profile before and after with `go test -bench`.

Player Characters (improve aggressively — current quality is unacceptable):
- **CRITICAL: All sprites must be TOP-DOWN / AERIAL VIEW.** The camera looks straight down. You see the top of the head, the shoulders, and barely any legs. If your sprite looks like a person standing facing you, it is WRONG. Use HumanoidAerialTemplate() proportions: head ~35%, torso/shoulders ~50%, legs ~15%.
- Use composite layered rendering (see `pkg/rendering/sprites/composite.go`). Layer order: Shadow(0) → Legs(5) → Body(10) → Armor(15) → Head(20) → Weapon(25) → Accessory(30) → Effect(40).
- Anatomy templates (`pkg/rendering/sprites/anatomy_template.go`) define body part sizes for 32×32 top-down sprites. Proportions may be reworked freely to improve visual quality — better proportions, more detailed features, and more expressive shapes are always welcome. The default HumanoidTemplate() proportions are WRONG for top-down; fix or bypass them.
- Support 4-direction (Up/Down/Left/Right) and 8-direction sprites. Maintain last facing direction in AnimationComponent.
- Status effect overlays (burning, frozen, poisoned, stunned, blessed, cursed) render at ZIndex 40 with color-coded intensity and particle counts.
- Player entities always animate at full rate (12 FPS) regardless of camera distance.
- Focus on making characters look like recognizable people SEEN FROM ABOVE — visible head/hair, shoulder width indicating body type, equipment visible on the body, shadow underneath. Not blobs, not profile silhouettes.
- Every pixel matters at 32×32. Use shading, color gradients, and highlights to give depth. Hair color, skin tone, and clothing should all be visually distinct.

NPCs & Creatures (improve aggressively — current quality is unacceptable):
- **CRITICAL: All sprites must be TOP-DOWN / AERIAL VIEW.** Same as player characters — drawn as seen from directly above.
- Use `pkg/procgen/entity/` templates for genre-aware generation. Entity types: Monster, Boss, NPC, Merchant. Sizes: Tiny, Small, Medium, Large, Huge.
- **NONHUMANOID CREATURES NEED DEDICATED TEMPLATES.** Do not reuse humanoid body plans for creatures that are not humanoid. Build top-down anatomy templates for: quadrupeds (4 legs radiating from body center), insects (segmented body, 6+ legs), serpents (elongated sinuous body), winged creatures (wide wingspan from above), amorphous entities (irregular blobby shapes), multi-limbed horrors (radial or asymmetric limbs). Each type should be immediately recognizable from its silhouette alone.
- Anatomy templates must scale proportionally with entity size. Larger creatures need wider torsos and legs relative to head size.
- Silhouette quality (`pkg/rendering/sprites/silhouette.go`) should target ≥0.7 (Good-to-Excellent). Measure Coverage, Compactness (4π×area/perimeter²), and EdgeClarity.
- NPCs should be visually distinct from each other — varied body shapes, hair, clothing, and facial features. No two NPCs should look the same. Seed-based generation must produce genuine variety, not trivial color swaps.
- Apply genre-specific visual tags to influence sprite shape types (e.g., horror → Skull head shape, fantasy → Circle/Ellipse head shapes).

Equipment (improve aggressively — current quality is unacceptable):
- Equipment overlays (`pkg/rendering/sprites/equipment.go`) render per-slot: Weapon, Armor, Accessory, Helmet, Boots, Gloves, Shield.
- Material types (Metal, Leather, Cloth, Wood, Crystal, Energy) should have visually distinct rendering. Use whatever visual properties best differentiate them — sheen, roughness, patterns, reflectivity, color shifts.
- Damage states degrade visuals progressively: Pristine → Worn → Damaged → Broken. Each state should be visually obvious at a glance.
- Enchantment glow is rarity-driven: Uncommon=Green, Rare=Blue, Epic=Purple, Legendary=Gold. Make enchantments visually exciting and clearly different from non-enchanted gear.
- Higher rarity = more visual complexity and material fidelity. Legendary items should look unmistakably special.
- Track equipment visuals via EquipmentVisualComponent with dirty flag for lazy regeneration. Visibility toggles per layer type.

Integration (mandatory — this is where past attempts fail):
- Register in `pkg/engine/system_init.go` → `InitializeGameSystems()`.
- Register in `cmd/client/handlers.go` → add to `systemsContainer` struct AND call `game.World.AddSystem()` in the appropriate init function or `registerNonCriticalSystems()`.
- If your system's Update signature differs from `(entities []*Entity, deltaTime float64)`, create a wrapper in `cmd/client/system_wrappers.go` matching existing patterns.
- The system MUST be on by default. No feature flags.
- Persistent component data must integrate with SerializeEntity/DeserializeEntity, or be explicitly transient.

Constraints:
- <800 lines total new/modified code (more is acceptable for avatar quality improvements if needed).
- `go build ./...` and `go vet ./...` must pass.
- Write table-driven tests. Target ≥65% coverage on new code.
- No breaking changes to saves, network protocol, or configs.
- Maintain 60+ FPS. Cache/pool on hot paths.

STEP 3 — VERIFY:
Run `go build ./...` and `go vet ./...`. Fix any errors before reporting.

STEP 4 — REPORT (keep concise):
1. **Enhancement**: What and why, 2-3 sentences.
2. **Files**: List with one-line change summary each.
3. **Integration**: Where the system is registered (exact file + function).
4. **Verification**: How to observe the improvement in-game.

STOP when the report is written and builds pass. Do not refactor unrelated code. Do not write documentation files. Do not suggest follow-up work.
