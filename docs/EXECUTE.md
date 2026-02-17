You are implementing ONE novel enhancement to a Go/Ebiten procedural multiplayer action-RPG called Venture. Act autonomously. Do not ask for approval.

KNOWN AVATAR PROBLEMS (read this first — high priority visual improvements):
The current avatars need significant improvement. Consider addressing one or more of these:
1. **WRONG PERSPECTIVE**: Sprites are rendered in profile/side view but the game uses a top-down camera. All entity sprites MUST be drawn as seen from above (aerial/overhead view). The head/shoulders should dominate the sprite; legs should be barely visible beneath the body. The default HumanoidTemplate() is wrong — it gives legs 48% of the sprite height, which is a profile view. Use and improve HumanoidAerialTemplate() proportions instead (head ~35%, torso ~50%, legs ~15%). Fix any sprite generation that draws entities as if viewed from the side.
2. **INSUFFICIENT DETAIL**: Sprites are visually barren — flat colors, no shading, no texture, no personality. At 32×32 every pixel matters. Add sub-pixel shading, color gradients, dithering, highlight/shadow on body parts, hair detail, clothing patterns, anything that makes sprites look crafted rather than placeholder.
3. **INSUFFICIENT VARIETY**: All avatars look nearly identical. Different NPCs, different players, different creature types should be immediately distinguishable at a glance. Vary body proportions, color palettes, head shapes, clothing silhouettes, and accessories. Seed-based generation should produce visually diverse output, not minor variations on one template.
4. **POOR NONHUMANOID REPRESENTATION**: Creatures, monsters, animals, and bosses use barely-modified humanoid templates. A spider should not look like a person. A dragon should not look like a person. Build and use dedicated nonhumanoid anatomy templates — quadrupeds, insects, serpents, amorphous blobs, winged creatures, multi-limbed horrors. Each creature type needs its own distinct body plan visible from above.

KNOWN SYSTEM PROBLEMS (read this first — high priority gameplay improvements):
Many core systems need depth and integration. Consider addressing one or more of these:
1. **SHALLOW PROGRESSION**: Many progression systems have placeholder logic or minimal depth. Skill trees, class progression, reputation systems, and achievements need meaningful choices, balanced rewards, and interconnected mechanics that create engaging long-term goals.
2. **DISCONNECTED SYSTEMS**: Systems exist in isolation without cross-system interactions. Economy should affect territory control, faction relationships should impact quests, weather should influence combat, housing should integrate with crafting. Build bridges between systems to create emergent gameplay.
3. **MINIMAL GENRE VARIATION**: Procedural generation often ignores genre context. Fantasy dungeons shouldn't look like sci-fi stations. Horror factions shouldn't behave like cyberpunk corporations. Each genre needs distinct procgen rules, AI behaviors, quest structures, and world-building patterns.
4. **PLACEHOLDER MECHANICS**: Core gameplay loops have stub implementations. AI behavior trees need more node types, combat needs tactical depth, crafting needs meaningful recipes, quests need better objective variety. Replace simple implementations with full-featured systems.

STEP 1 — DISCOVER (spend ≤5 minutes here):
- Run `git log --oneline -20` to avoid duplicating recent work.
- Read pkg/engine/system_init.go to understand registered systems.
- Grep for TODO, FIXME, stub, placeholder in pkg/engine/ and pkg/procgen/.
- Pick ONE enhancement you have NOT seen in git history. Roll a d20 to decide the category:
  - **System improvements (roll of 1–7 — address the KNOWN SYSTEM PROBLEMS above):**
    - **Progression depth** — skill tree branching, class synergies, reputation consequence systems, achievement chains, prestige mechanics
    - **System integration** — economy↔territory, faction↔quest, weather↔combat, housing↔crafting, companion↔skills, guild↔raids
    - **Genre variation** — genre-specific AI personalities, quest objective variety, loot table customization, dungeon layout algorithms, NPC behavior patterns
    - **Mechanic depth** — behavior tree node types, combat tactical options, crafting recipe complexity, quest chain branching, dialog response systems
    - **AI improvements** — squad tactics, companion learning, enemy adaptation, merchant pricing strategies, NPC schedules and routines
    - **World systems** — city evolution, economy simulation, faction warfare, territory sieges, world events, environmental destruction
    - **Social features** — guild progression, trade mechanics, mail system depth, chat channels, player housing interactions
  - **Avatar improvements (roll of 8–14 — address the KNOWN AVATAR PROBLEMS above):**
    - **Perspective fixes** — convert any profile/side-view sprites to proper top-down aerial view. This is the single most impactful fix.
    - **Nonhumanoid templates** — build dedicated top-down anatomy templates for creature types that are not humanoid (quadrupeds, insects, serpents, flying creatures, amorphous entities, multi-limbed creatures). Every creature type deserves its own body plan.
    - **Player character visuals** — composite layering, anatomy detail, directional sprites, proportions, body shapes, facial features, skin/hair color variety, idle poses, shading, clothing detail
    - **NPC variety** — genre-aware body templates, size-based anatomy, silhouette quality, visual personality, distinctive appearance per NPC, varied clothing and coloring
    - **Equipment visuals** — material rendering fidelity, damage-state degradation, enchantment glow/particles, rarity-based detail scaling, weapon silhouettes, armor shaping
    - **Sprite detail** — sub-pixel shading, color gradients, dithering, material textures, highlight/shadow, edge definition, anti-aliasing
    - **Animation improvements** — smoother transitions, new states, expressive movement, attack/cast/hurt animations, idle breathing/fidget
  - **Character Customization (roll of 15–20):**
    - **Appearance customization** — facial features (eyes, nose, mouth, ears), body proportions (height, build, muscle definition), skin tone palette expansion, hair styles and colors, age indicators (wrinkles, scars)
    - **Cosmetic layering** — clothing dye systems, pattern overlays, emblems and insignias, armor tinting, weapon skins, particle effect customization
    - **Character creation UI** — preview system with rotation, randomization options, preset templates by genre, saved appearance profiles, import/export character appearance
    - **Runtime customization** — barber shops, cosmetic vendors, appearance change items, transmogrification systems, wardrobe collections
    - **Cultural customization** — faction-specific appearance options, genre-themed cosmetics, cultural clothing and accessories, regional hair/skin variations
    - **Expression and emotes** — facial expression system, emote animations, idle pose variations, victory/defeat poses, social gesture animations
    - **Advanced features** — tattoos and body art, jewelry and piercings, makeup and face paint, battle scars and wounds, prosthetics and augmentations
- If multiple candidates exist within your category, pick the one that most improves the game experience. Within system work, integration and progression depth are highest-value. Within avatar work, perspective fixes and nonhumanoid templates are highest-value. Within character customization work, appearance customization and character creation UI are highest-value.

STEP 2 — IMPLEMENT (this is the bulk of the work):
Follow these rules strictly. Violations are build failures.

Architecture:
- Components = pure data + `Type() string`. Zero methods with logic.
- Systems = all logic. Signature: `Update(entities []*Entity, deltaTime float64)`.
- Procgen: `rand.New(rand.NewSource(seed))` only. Never global rand. Never time.Now().
- Logging: `logrus.WithFields(logrus.Fields{"system_name": "...", ...})`.
- No external assets. No new dependencies beyond go.mod.

Visual & Animation (for avatar improvements):
- Lighting must use radial gradients with proper falloff (linear, quadratic, inverse-square). No flat circles.
- Shadows use soft penumbra with distance-based falloff. Support genre-specific opacity presets.
- Post-processing effects (color grading, vignette, chromatic aberration) must be genre-aware.
- Animation playback: 12 FPS (0.083s per frame), 8 frames per state. Use distance-based LOD (full rate at ≤200px, half at ≤400px, minimal beyond).
- Sprite generation must be seeded and cached (LRU, max 100 entries). Pool image buffers by size bucket.
- All visual enhancements must maintain 60+ FPS. Profile before and after with `go test -bench`.

Player Characters (avatar improvements):
- **CRITICAL: All sprites must be TOP-DOWN / AERIAL VIEW.** The camera looks straight down. You see the top of the head, the shoulders, and barely any legs. If your sprite looks like a person standing facing you, it is WRONG. Use HumanoidAerialTemplate() proportions: head ~35%, torso/shoulders ~50%, legs ~15%.
- Use composite layered rendering (see `pkg/rendering/sprites/composite.go`). Layer order: Shadow(0) → Legs(5) → Body(10) → Armor(15) → Head(20) → Weapon(25) → Accessory(30) → Effect(40).
- Anatomy templates (`pkg/rendering/sprites/anatomy_template.go`) define body part sizes for 32×32 top-down sprites. Proportions may be reworked freely to improve visual quality — better proportions, more detailed features, and more expressive shapes are always welcome. **Delete all profile-view templates** (including HumanoidTemplate() and any other template that renders entities as seen from the side). Replace them with aerial-view equivalents. Do not leave incorrect profile-view code in the codebase.
- Support full 360-degree rotation for sprite facing. Maintain facing angle in AnimationComponent.
- Status effect overlays (burning, frozen, poisoned, stunned, blessed, cursed) render at ZIndex 40 with color-coded intensity and particle counts.
- Player entities always animate at full rate (12 FPS) regardless of camera distance.
- Focus on making characters look like recognizable people SEEN FROM ABOVE — visible head/hair, shoulder width indicating body type, equipment visible on the body, shadow underneath. Not blobs, not profile silhouettes.
- Every pixel matters at 32×32. Use shading, color gradients, and highlights to give depth. Hair color, skin tone, and clothing should all be visually distinct.

NPCs & Creatures (avatar improvements):
- **CRITICAL: All sprites must be TOP-DOWN / AERIAL VIEW.** Same as player characters — drawn as seen from directly above.
- Use `pkg/procgen/entity/` templates for genre-aware generation. Entity types: Monster, Boss, NPC, Merchant. Sizes: Tiny, Small, Medium, Large, Huge.
- **NONHUMANOID CREATURES NEED DEDICATED TEMPLATES.** Do not reuse humanoid body plans for creatures that are not humanoid. Build top-down anatomy templates for: quadrupeds (4 legs radiating from body center), insects (segmented body, 6+ legs), serpents (elongated sinuous body), winged creatures (wide wingspan from above), amorphous entities (irregular blobby shapes), multi-limbed horrors (radial or asymmetric limbs). Each type should be immediately recognizable from its silhouette alone.
- Anatomy templates must scale proportionally with entity size. Larger creatures need wider torsos and legs relative to head size.
- Silhouette quality (`pkg/rendering/sprites/silhouette.go`) should target at least "Good" (score ≥0.7; "Excellent" is >0.8). Measure Coverage, Compactness (4π×area/perimeter²), and EdgeClarity.
- NPCs should be visually distinct from each other — varied body shapes, hair, clothing, and facial features. No two NPCs should look the same. Seed-based generation must produce genuine variety, not trivial color swaps.
- Apply genre-specific visual tags to influence sprite shape types (e.g., horror → Skull head shape, fantasy → Circle/Ellipse head shapes).

Equipment (avatar improvements):
- Equipment overlays (`pkg/rendering/sprites/equipment.go`) render per-slot: Weapon, Armor, Accessory, Helmet, Boots, Gloves, Shield.
- Material types (Metal, Leather, Cloth, Wood, Crystal, Energy) should have visually distinct rendering. Use whatever visual properties best differentiate them — sheen, roughness, patterns, reflectivity, color shifts.
- Damage states degrade visuals progressively: Pristine → Worn → Damaged → Broken. Each state should be visually obvious at a glance.
- Enchantment glow is rarity-driven: Uncommon=Green, Rare=Blue, Epic=Purple, Legendary=Gold. Make enchantments visually exciting and clearly different from non-enchanted gear.
- Higher rarity = more visual complexity and material fidelity. Legendary items should look unmistakably special.
- Track equipment visuals via EquipmentVisualComponent with dirty flag for lazy regeneration. Visibility toggles per layer type.

Progression Systems (system improvements):
- Skill trees (`pkg/engine/skill_progression_system.go`) should offer meaningful branching choices. Each node should enable new playstyles or synergize with other skills. Avoid pure stat bonuses — prefer unlocking abilities, modifying existing abilities, or enabling cross-skill combos.
- Class progression (`pkg/engine/class_progression_system.go`) needs depth beyond level-up bonuses. Implement specializations, prestige classes, multiclass synergies. Each class should feel mechanically distinct with unique abilities and resource management.
- Reputation systems (`pkg/engine/reputation_system.go`) should have gameplay consequences. Faction standing affects quest availability, merchant prices, territory access, and NPC behavior. Build reputation curves that create meaningful long-term goals.
- Achievement systems (`pkg/engine/achievement.go`) should chain together and unlock content. Achievements should guide exploration, reward mastery, and grant permanent bonuses or cosmetic rewards.

System Integration (system improvements):
- Cross-system mechanics create emergent gameplay. Examples: economy price fluctuations affect territory control costs; faction wars generate dynamic quests; weather influences combat effectiveness; housing crafting stations boost recipe quality.
- Look for unused interfaces between systems. Economy system has price data — territory system should read it. Faction system tracks relationships — quest system should use them. Weather system affects world — combat system should respond.
- Add components to bridge systems (e.g., EconomicInfluenceComponent on territories, WeatherSensitivityComponent on combat entities). Systems communicate via components, not direct calls.
- Integration should feel natural, not forced. Start with small connections (weather → movement speed) before complex interactions (economy → territory → quests).

Genre Variation (system improvements):
- Genre context (`pkg/procgen/genre/`) should deeply influence generation. Each genre needs distinct rules for dungeons, factions, quests, loot, AI, and world-building.
- Use genre templates for procgen parameters. Fantasy → organic dungeon layouts, magic-focused loot, honor-based factions. Sci-fi → geometric structures, tech loot, corporate factions. Horror → claustrophobic spaces, survival resources, insanity mechanics. Cyberpunk → vertical architecture, augmentation loot, gang factions.
- AI personalities should vary by genre. Fantasy NPCs follow honor codes, sci-fi NPCs optimize efficiency, horror NPCs show fear/paranoia, cyberpunk NPCs pursue profit.
- Quest structures differ by genre. Fantasy → hero's journey, epic quests. Sci-fi → mission briefings, exploration. Horror → investigation, survival. Cyberpunk → heists, corporate espionage.

AI & Behavior (system improvements):
- Behavior trees (`pkg/engine/behavior_tree_system.go`, `pkg/engine/behavior_tree_nodes.go`) need more node types. Add: patrol routes, cover-seeking, flanking, retreating, calling for help, using items, environmental interaction.
- Squad tactics (`pkg/engine/squad_system.go`) should coordinate actions. Enemies in groups should focus-fire, protect wounded allies, use formations, and combine abilities.
- Companion AI (`pkg/engine/companion_ai_system.go`) should learn from player behavior. Track preferred tactics, adapt to player playstyle, suggest strategies.
- Merchant pricing (`pkg/world/economy/pricing_engine.go`) should respond to supply/demand, player reputation, and market trends. Prices should feel dynamic, not static.

World Systems (system improvements):
- City evolution (`pkg/engine/city_evolution_system.go`) should reflect player actions. Completed quests improve infrastructure, economic activity attracts merchants, faction control changes city appearance.
- Territory control (`pkg/world/territory/`) should involve meaningful strategic choices. Territory resources, defensive positions, trade routes, and faction influence all matter.
- World events (`pkg/engine/world_events_system.go`, `pkg/integration/world_events/`) should cascade into other systems. Natural disasters affect economy, faction wars change territory, festivals offer limited-time quests.
- Environmental destruction (`pkg/engine/physics/destruction/`) should persist and matter. Destroyed walls create new paths, flooded areas become swimming zones, burned forests change terrain types.

Integration (mandatory — this is where past attempts fail):
- Register in `pkg/engine/system_init.go` → `InitializeGameSystems()`.
- Register in `cmd/client/handlers.go` → add to `systemsContainer` struct AND call `game.World.AddSystem()` in the appropriate init function or `registerNonCriticalSystems()`.
- If your system's Update signature differs from `(entities []*Entity, deltaTime float64)`, create a wrapper in `cmd/client/system_wrappers.go` matching existing patterns.
- The system MUST be on by default. No feature flags.
- Persistent component data must integrate with SerializeEntity/DeserializeEntity, or be explicitly transient.

Constraints:
- Keep changes focused and targeted. Avatar improvements should focus on visual quality. System improvements should focus on gameplay depth.
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
