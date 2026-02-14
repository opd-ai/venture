You are implementing ONE novel enhancement to a Go/Ebiten procedural multiplayer action-RPG called Venture. Act autonomously. Do not ask for approval.

STEP 1 — DISCOVER (spend ≤5 minutes here):
- Run `git log --oneline -20` to avoid duplicating recent work.
- Read pkg/engine/system_init.go to understand registered systems.
- Grep for TODO, FIXME, stub, placeholder in pkg/engine/ and pkg/procgen/.
- Pick ONE enhancement you have NOT seen in git history. Prefer:
  - Connecting two existing systems that don't yet interact
  - Adding depth to a system with minimal/placeholder logic
  - New visual feedback for existing gameplay mechanics
  - Genre-aware variation in procgen outputs
  - Improving visual detail (lighting falloff, shadow quality, sprite fidelity)
  - Enhancing visual realism (post-processing effects, color grading, ambient occlusion)
  - Animation improvements (smoother transitions, new states, distance-based LOD tuning)
  - Player character visuals (composite layering, anatomy detail, directional sprites, status overlays)
  - NPC and creature visuals (genre-aware body templates, size-based anatomy, silhouette quality, facial detail)
  - Equipment visuals (material rendering fidelity, damage-state degradation, enchantment glow/particles, rarity-based detail scaling)
- If multiple candidates exist, pick the one requiring fewest files changed.

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

Player Characters:
- Use composite layered rendering (see `pkg/rendering/sprites/composite.go`). Layer order: Shadow(0) → Legs(5) → Body(10) → Armor(15) → Head(20) → Weapon(25) → Accessory(30) → Effect(40).
- Anatomy templates (`pkg/rendering/sprites/anatomy_template.go`) define body part sizes for 32×32 top-down sprites: Head 6×4, Torso 8×13, Legs 6×15 (parts stack vertically, total=32px height). Respect PreferredPixelSize for sub-pixel precision.
- Support 4-direction (Up/Down/Left/Right) and 8-direction sprites. Maintain last facing direction in AnimationComponent.
- Status effect overlays (burning, frozen, poisoned, stunned, blessed, cursed) render at ZIndex 40 with color-coded intensity and particle counts.
- Player entities always animate at full rate (12 FPS) regardless of camera distance.

NPCs & Creatures:
- Use `pkg/procgen/entity/` templates for genre-aware generation. Entity types: Monster, Boss, NPC, Merchant. Sizes: Tiny, Small, Medium, Large, Huge.
- Anatomy templates must scale proportionally with entity size. Larger creatures need wider torsos and legs relative to head size.
- Silhouette quality (`pkg/rendering/sprites/silhouette.go`) must score ≥0.6 (Good). Measure Coverage, Compactness (4π×area/perimeter²), and EdgeClarity.
- NPC facial detail (eyes 2px, mouth 1–2px) uses ColorRole mapping (Primary/Secondary/Accent) from the genre palette.
- Apply genre-specific visual tags to influence sprite shape types (e.g., horror → Skull head shape, fantasy → Circle/Ellipse head shapes).

Equipment:
- Equipment overlays (`pkg/rendering/sprites/equipment.go`) render per-slot: Weapon, Armor, Accessory, Helmet, Boots, Gloves, Shield.
- Material types (Metal, Leather, Cloth, Wood, Crystal, Energy) have distinct visual properties — Sheen (0.1–1.0), Roughness (0.0–0.8), PatternType (crosshatch/grain/weave/rings/faceted/flow), Reflectivity (0.0–1.0). Metal: high sheen 0.9, low roughness 0.2. Cloth: low sheen 0.1, high roughness 0.8.
- Damage states degrade visuals progressively: Pristine (full opacity) → Worn (0.95 opacity, 0.1 darkening) → Damaged (0.85 opacity, 0.25 darkening, 0.4 crack density) → Broken (0.7 opacity, 0.4 darkening, 0.7 crack density + edge roughness).
- Enchantment glow is rarity-driven: Uncommon=Green, Rare=Blue, Epic=Purple, Legendary=Gold. Intensity scales 0.2–0.8, particle count 2–12, pulse speed 0.5–1.2.
- Rarity tiers control detail level: Common 0.3, Uncommon 0.4, Rare 0.6, Epic 0.8, Legendary 1.0. Higher detail = more shape complexity and material fidelity.
- Track equipment visuals via EquipmentVisualComponent with dirty flag for lazy regeneration. Visibility toggles per layer type.

Integration (mandatory — this is where past attempts fail):
- Register in `pkg/engine/system_init.go` → `InitializeGameSystems()`.
- Register in `cmd/client/handlers.go` → add to `systemsContainer` struct AND call `game.World.AddSystem()` in the appropriate init function or `registerNonCriticalSystems()`.
- If your system's Update signature differs from `(entities []*Entity, deltaTime float64)`, create a wrapper in `cmd/client/system_wrappers.go` matching existing patterns.
- The system MUST be on by default. No feature flags.
- Persistent component data must integrate with SerializeEntity/DeserializeEntity, or be explicitly transient.

Constraints:
- <500 lines total new/modified code.
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