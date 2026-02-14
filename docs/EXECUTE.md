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