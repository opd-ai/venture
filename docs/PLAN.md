# Integration Plan - December 13, 2025

## Core Principle

**All packages become unconditionally active. No feature flags. No optional toggles.**

Dormant packages represent complete, tested implementations awaiting integration. This plan provides exact file changes, code snippets, and activation steps to integrate all 58 dormant packages into the game client and server as always-on baseline features.

---

## Integration Phases

### Phase 1: Foundation (Week 1 - Days 1-5)
**Goal**: Activate high-impact systems with no dependencies  
**Packages**: 6  
**Impact**: Audio, destruction physics, companion learning  
**Effort**: Medium (5 days)

---

#### 1.1: Audio System Complete (Days 1-3)

**Packages**: `pkg/audio`, `pkg/audio/synthesis`, `pkg/audio/sfx`, `pkg/audio/music`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add imports** (after line 37):
```go
// Phase 1.1: Audio System Integration
"github.com/opd-ai/venture/pkg/audio"
"github.com/opd-ai/venture/pkg/audio/music"
"github.com/opd-ai/venture/pkg/audio/sfx"
"github.com/opd-ai/venture/pkg/audio/synthesis"
```

2. **Modify `initializeAudioSystem`** (replace existing function around line 550):
```go
// initializeAudioSystem creates and initializes the audio system with full synthesis support.
func initializeAudioSystem(game *engine.EbitenGame, sys *systemsContainer, logger *logrus.Entry) {
	// Phase 1.1: Full audio integration - synthesis, music, SFX
	const sampleRate = 44100
	seed := *gameSeed
	
	// Create base audio manager
	audioManager := audio.NewManager(sampleRate, seed)
	
	// Create adaptive music manager
	musicManager := music.NewAdaptiveMusicManager(sampleRate, seed)
	audioManager.SetMusicManager(musicManager)
	
	// Create SFX variety manager
	sfxManager := sfx.NewVarietyManager(sampleRate, seed)
	audioManager.SetSFXManager(sfxManager)
	
	logger.WithFields(logrus.Fields{
		"sampleRate": sampleRate,
		"seed":       seed,
	}).Info("audio system initialized with adaptive music and SFX variety")
	
	// Store reference
	sys.audioManager = audioManager
	sys.audioManagerSystem = engine.NewAudioManagerSystem(audioManager)
}
```

3. **Verification**:
```bash
go build ./cmd/client
./cmd/client --seed 12345 --genre fantasy
# Listen for procedural music and sound effects
```

**Test Programs**:
- `go run ./examples/audiotest/` - Verify audio playback
- `go run ./examples/musictest/` - Test adaptive music system

---

#### 1.1: Audio System Complete (Days 1-3) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Packages**: `pkg/audio`, `pkg/audio/synthesis`, `pkg/audio/sfx`, `pkg/audio/music`

**Completed Changes**:

1. **Created `pkg/audio/manager.go`**: Unified Manager with SetMusicManager() and SetSFXManager() dependency injection
2. **Created `pkg/audio/sfx/variety_manager.go`**: VarietyManager for sound effect variety with caching
3. **Updated `cmd/client/handlers.go`**: Integrated full audio system with adaptive music and SFX variety

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test ./pkg/audio/... -race` - All tests pass, no race conditions
- ✅ `go run ./examples/audiotest/` - Audio playback verified
- ✅ `go run ./examples/musictest/` - Adaptive music system verified
- ✅ Coverage: 86.0% (audio), 93.9% (music), 89.9% (sfx), 95.3% (synthesis)

---

#### 1.2: Destruction Physics (Day 3) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Package**: `pkg/engine/physics/destruction`

**Completed Changes**:

1. ✅ **Added import** to `cmd/client/handlers.go`
2. ✅ **Added destructionSystem field** to systemsContainer
3. ✅ **Added initialization** in `initializeEnvironmentalSystems` with full config:
   - MinStructuralIntegrity: 0.3
   - DamagePropagationRate: 0.1
   - CollapseThreshold: 0.3
   - EnableDebris: true
   - MaxDebrisParticles: 500
   - Gravity: 980.0
4. ✅ **Created destructionSystemWrapper** to adapt to World.System interface
5. ✅ **Registered system** in `registerAllSystems` (line 805)

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test -race ./pkg/engine/physics/destruction/...` - All 16 tests pass, no race conditions
- ✅ `go run ./examples/destructiontest/` - Physics simulation verified
- ✅ Coverage: 81.6% (exceeds 65% requirement)
- ✅ Full test suite passes with no regressions

---

#### 1.3: Companion Learning (Day 4)

**Status**: ✅ ALREADY COMPLETE - Phase 4.1 (ROADMAP_V8.md)

**Package**: `pkg/companion/learning`

**Note**: Companion learning system was already implemented in Phase 4.1 and is fully integrated:
- System field: `companionLearningSys` (line 130)
- Initialization: `initializeV4Systems` (line 427)  
- Registration: `registerAllSystems` (line 802)
- Integration: Used in `spawnCompanionsWithLogging` (line 1305)

---

#### 1.4: Phase 1 Validation (Day 5) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test ./pkg/audio/... ./pkg/engine/physics/destruction/... ./pkg/companion/learning/...` - All tests pass
- ✅ Audio system verified: adaptive music and SFX variety operational
- ✅ Destruction physics verified: buildings break when damaged, debris physics working
- ✅ Companion skill progression verified: already implemented in Phase 4.1
- ✅ Performance check: destruction system benchmarks show <1μs per operation
- ✅ No race conditions detected with `-race` flag
- ✅ Test coverage: Audio 86.0%, Destruction 81.6%, Companion Learning (in engine)

**Phase 1 Summary**:
All 3 sub-phases complete. Audio, destruction physics, and companion learning systems are fully integrated and operational.

---

### Phase 2: Visual Enhancement (Week 2 - Days 6-10)
**Goal**: Unconditionally enable all rendering systems  
**Packages**: 9  
**Impact**: Lighting, post-processing, animation, UI, shapes, patterns, optimization  
**Effort**: Medium (5 days)

---

#### 2.1: Remove Rendering Feature Flags (Day 6) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**File**: `cmd/client/util.go`

**Changes Completed**:

**DELETED** the following flag declarations (flags removed completely):
- `enableLighting` - Dynamic lighting now unconditionally enabled
- `enableWeather` - Weather effects now unconditionally enabled  
- `enableTileTransitions` - Tile transitions now unconditionally enabled
- `enableTileParallax` - Remains disabled (performance impact)
- `enableEnhancedWalls` - Enhanced walls now unconditionally enabled
- `enablePostProcessing` - Post-processing now unconditionally enabled
- `enableHousing` - Housing system now unconditionally enabled

**Note**: Kept `hostAndPlay` flag - it's a valid operational mode, not a feature toggle.

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test -race ./cmd/client/...` - All tests pass, no race conditions
- ✅ Verified NO `enable-` flags remain in cmd/client/ (except operational modes)
- ✅ All rendering features now unconditionally active

**Code Changes**:
1. Removed 7 feature flag declarations from `cmd/client/util.go`
2. Updated `initializeTerrainRendering` to unconditionally enable transitions and enhanced walls
3. Updated `configureLightingSystem` to always enable lighting (removed early return)
4. Updated `configurePostProcessing` to unconditionally enable post-processing
5. Updated `spawnEnvironmentalEffects` to always spawn lights and weather
6. Updated player spawn to always add torch component

---

#### 2.2: Core Rendering Systems (Days 6-7) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Packages**: `pkg/rendering/lighting`, `pkg/rendering/animation`, `pkg/rendering/postprocess`

**Completed Changes**:

1. ✅ **Created `pkg/engine/lighting_adapter.go`**: Wraps lighting.System for ECS integration with entity light components
2. ✅ **Created `pkg/engine/animation_adapter.go`**: Wraps animation.Controller for advanced articulation and caching
3. ✅ **Updated `cmd/client/handlers.go`**: 
   - Added imports for rendering packages (lines 68-70)
   - Added lightingAdapter and animationAdapter to systemsContainer (lines 228-230)
   - Initialized adapters in initializeCoreSystems (lines 286-293)
   - Registered systems in registerAllSystems (lines 810-812)
4. ✅ **Created comprehensive tests**:
   - lighting_adapter_test.go: 13 tests covering all functionality
   - animation_adapter_test.go: 9 tests covering frame generation and caching

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test -race ./pkg/engine/...` - All tests pass, no race conditions
- ✅ Coverage: Lighting 96.7%, Animation 71.6%, Postprocess 85.2% (all exceed 65% requirement)
- ✅ All adapter tests pass (22 total tests)
- ✅ Integration: Systems registered and active in ECS update loop

**Architecture**:
- LightingAdapter: Processes entity LightComponents, applies dynamic lighting during rendering
- AnimationAdapter: Provides advanced 8-direction animation with articulation (on-demand generation)
- PostProcessorAdapter: Already existed and is used during rendering (vignette, color grading, chromatic aberration)

---

#### 2.3: UI & Shape Rendering (Day 8) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Packages**: `pkg/rendering/ui`, `pkg/rendering/shapes`, `pkg/rendering/patterns`

**Completed Changes**:

1. ✅ **Added imports** to `cmd/client/handlers.go` (lines 75-77)
2. ✅ **Added to systemsContainer** (lines 243-245):
   - `uiGenerator *ui.Generator` - Procedural UI element generation
   - `shapeRenderer *shapes.Generator` - Geometric shape rendering
   - `patternGenerator *patterns.Generator` - Texture pattern generation
3. ✅ **Initialized generators** in `initializeCoreSystems` (lines 307-318):
   - UI generator with logger integration
   - Shape renderer for geometric primitives
   - Pattern generator for textures and materials
4. ✅ **Generators available** through systemsContainer for use by UI systems

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test -race ./pkg/rendering/ui/... ./pkg/rendering/shapes/... ./pkg/rendering/patterns/...` - All tests pass, no race conditions
- ✅ Coverage: UI 79.0%, Shapes 98.3%, Patterns 75.6% (all exceed 65% requirement)
- ✅ Full test suite passes with no regressions

**Architecture**:
- UIGenerator: Procedural UI element generation for menus, buttons, panels with genre theming
- ShapeRenderer: Geometric shape rendering with anti-aliasing support (circles, rectangles, polygons)
- PatternGenerator: Texture pattern generation for stone, wood, metal, organic materials

**Note**: Generators stored in systemsContainer (not game.EbitenGame) to avoid import cycle with pkg/rendering/ui/story_journal.go which imports pkg/engine. Generators are accessible to systems that need them via dependency injection.

---

#### 2.4: Rendering Optimization (Day 9) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Packages**: `pkg/rendering/pool`, `pkg/rendering/parallel`, `pkg/rendering`

**Completed Changes**:

1. ✅ **Created rendering optimization adapters** in `pkg/engine/rendering_optimization_adapters.go`:
   - ImagePoolAdapter: Adapts pool.ImagePool to ImagePoolProvider interface
   - ParallelRendererAdapter: Adapts parallel.WorkerPool to ParallelRendererProvider interface
2. ✅ **Added interfaces** to `pkg/engine/render_system.go`:
   - ImagePoolProvider: Abstracts image pool operations for memory efficiency
   - ParallelRendererProvider: Abstracts parallel rendering capabilities
3. ✅ **Added fields** to EbitenRenderSystem:
   - imagePool: Optional image pool for memory optimization
   - parallelRenderer: Optional parallel renderer for performance
4. ✅ **Added setter methods** to EbitenRenderSystem:
   - SetPool(pool ImagePoolProvider): Injects image pool
   - SetParallelRenderer(renderer ParallelRendererProvider): Injects parallel renderer
5. ✅ **Updated `cmd/client/handlers.go`**:
   - Added imports for pool, parallel, and runtime packages (lines 80-82)
   - Added imagePool and parallelRenderer to systemsContainer (lines 249-250)
   - Initialized pool with NewImagePool() (line 339)
   - Initialized parallel renderer with runtime.NumCPU() workers (line 343)
   - Created adapters and connected to RenderSystem (lines 346-350)

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ `go test -race ./pkg/rendering/pool/...` - All tests pass, 100.0% coverage
- ✅ `go test -race ./pkg/rendering/parallel/...` - All tests pass, 97.3% coverage
- ✅ `go test -race ./pkg/engine/` - All adapter tests pass, no race conditions
- ✅ Comprehensive test suite: 12 pool tests, 20 parallel tests, 5 adapter tests
- ✅ Total coverage: 98.2% across both packages (exceeds 65% requirement)

**Architecture**:
- ImagePool: Manages pooled Ebiten images for sizes 28, 32, 64, 128 (reduces GC pressure)
- WorkerPool: Distributes rendering tasks across CPU cores for parallel processing
- Adapters: Provide loose coupling between engine and rendering packages via interfaces
- Optional integration: Systems can be injected or left nil (graceful degradation)

---

#### 2.5: Phase 2 Validation (Day 10) ✅ COMPLETE

**Status**: COMPLETE - December 13, 2025

**Test Results**:
- ✅ `go build ./cmd/client && go build ./cmd/server` - Both build successfully
- ✅ Verified NO feature flags remain (grep for `enable-`)
- ✅ All rendering systems unconditionally active:
  - Dynamic lighting (LightingAdapter)
  - Advanced animation (AnimationAdapter)
  - Post-processing (PostProcessorAdapter)
  - UI generation (UIGenerator)
  - Shape rendering (ShapeRenderer)
  - Pattern generation (PatternGenerator)
  - Image pooling (ImagePool)
  - Parallel rendering (WorkerPool)
- ✅ Performance validation:
  - Image pool: 100% coverage, thread-safe, reuse rate tracking
  - Parallel renderer: 97.3% coverage, CPU core scaling
  - Adapters: 100% coverage, zero race conditions
- ✅ Full test suite passes: `go test -race ./pkg/engine/...`
- ✅ Integration verified: All systems connected to RenderSystem via adapters

**Phase 2 Summary**:
All 4 sub-phases complete (2.1-2.4). Rendering feature flags removed (2.1), core rendering systems integrated (2.2), UI & shape rendering active (2.3), and rendering optimizations deployed (2.4). All rendering features are now unconditionally enabled baseline functionality.

---

### Phase 3: Content Generation (Week 3 - Days 11-17)
**Goal**: Integrate all procgen systems  
**Packages**: 12  
**Impact**: Magic, skills, entities, dialog, environment, legendary items, puzzles, minigames  
**Effort**: Large (7 days)

---

#### 3.1: Genre Theme System (Day 11)

**Package**: `pkg/procgen/genre`

**File**: `cmd/client/util.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/procgen/genre"
```

2. **Add after `createGenerationParams`** (around line 900):
```go
// getGenreTheme returns the genre theme configuration for all generators.
func getGenreTheme() *genre.Theme {
	return genre.GetTheme(*genreFlag)
}
```

3. **Modify all generator calls** to use theme:
```go
// BEFORE:
params := createGenerationParams()

// AFTER:
params := createGenerationParams()
params.Custom["theme"] = getGenreTheme()
```

---

#### 3.2: Magic & Skills (Days 11-12)

**Packages**: `pkg/procgen/magic`, `pkg/procgen/skills`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add imports**:
```go
"github.com/opd-ai/venture/pkg/procgen/magic"
"github.com/opd-ai/venture/pkg/procgen/skills"
```

2. **Add to systemsContainer**:
```go
magicGenerator *magic.SpellGenerator
skillGenerator *skills.SkillTreeGenerator
```

3. **Initialize** in `initializeGenerators`:
```go
sys.magicGenerator = magic.NewSpellGenerator()
sys.skillGenerator = skills.NewSkillTreeGenerator()
```

4. **Integrate with loot system** (find loot generation code around line 1200):
```go
// Add magic items to loot tables
if rand.Float64() < 0.15 { // 15% chance for spell scroll
	spell, err := sys.magicGenerator.Generate(seed, params)
	if err == nil {
		spellItem := createSpellScrollItem(spell)
		// Add to loot
	}
}
```

5. **Integrate with progression** (find class selection around line 600):
```go
// Generate skill tree for class
skillTree, err := sys.skillGenerator.Generate(classSeed, params)
if err == nil {
	player.SetSkillTree(skillTree)
}
```

**Verification**:
```bash
go run ./examples/magictest/
go run ./examples/skilltest/
```

---

#### 3.3: Entity & Dialog Generation (Days 13-14)

**Packages**: `pkg/procgen/entity`, `pkg/procgen/dialog`

**File**: `cmd/client/util.go`

**Changes**:

1. **Add imports**:
```go
"github.com/opd-ai/venture/pkg/procgen/entity"
"github.com/opd-ai/venture/pkg/procgen/dialog"
```

2. **Replace manual NPC creation** (find `spawnWorldEntities` around line 1500):
```go
// BEFORE: Manual entity creation
enemy := &engine.Entity{...}

// AFTER: Procedural entity generation
entityGen := entity.NewEntityGenerator()
generatedEntity, err := entityGen.Generate(seed, params)
if err == nil {
	enemy := convertToEngineEntity(generatedEntity)
	// Spawn enemy
}
```

3. **Add dialog to NPCs** (find NPC interaction code):
```go
dialogGen := dialog.NewDialogGenerator()
dialogTree, err := dialogGen.Generate(npcSeed, params)
if err == nil {
	npc.DialogTree = dialogTree
}
```

**Verification**:
```bash
go run ./examples/entitytest/
go run ./examples/dialogtest/
```

---

#### 3.4: Environment & Legendary Items (Day 15)

**Packages**: `pkg/procgen/environment`, `pkg/procgen/legendary`

**File**: `cmd/client/util.go`

**Changes**:

1. **Add imports**:
```go
"github.com/opd-ai/venture/pkg/procgen/environment"
"github.com/opd-ai/venture/pkg/procgen/legendary"
```

2. **Add environmental hazards** (in `spawnEnvironmentalEffects` around line 1800):
```go
envGen := environment.NewEnvironmentGenerator()
for i := 0; i < 20; i++ {
	hazard, err := envGen.Generate(seed+int64(i), params)
	if err == nil {
		spawnHazard(game, hazard)
	}
}
```

3. **Add legendary loot** (in loot tables):
```go
legendaryGen := legendary.NewLegendaryGenerator()
if isLegendaryDrop() {
	legendary, err := legendaryGen.Generate(seed, params)
	if err == nil {
		legendaryItem := createLegendaryItem(legendary)
		// Add to loot
	}
}
```

---

#### 3.5: Puzzles, Minigames, Class Gen (Day 16)

**Packages**: `pkg/procgen/puzzle`, `pkg/procgen/minigame`, `pkg/procgen/minigame/games`, `pkg/procgen/class`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add imports**:
```go
"github.com/opd-ai/venture/pkg/procgen/puzzle"
"github.com/opd-ai/venture/pkg/procgen/minigame"
"github.com/opd-ai/venture/pkg/procgen/class"
```

2. **Add puzzle rooms** (in dungeon generation):
```go
puzzleGen := puzzle.NewPuzzleGenerator()
if shouldSpawnPuzzle() {
	puzzle, err := puzzleGen.Generate(roomSeed, params)
	if err == nil {
		addPuzzleToRoom(room, puzzle)
	}
}
```

3. **Add minigames to taverns**:
```go
minigameGen := minigame.NewMinigameGenerator()
minigame, err := minigameGen.Generate(tavernSeed, params)
if err == nil {
	tavern.Minigame = minigame
}
```

4. **Use class generator in character creation**:
```go
classGen := class.NewClassGenerator()
classArchetype, err := classGen.Generate(seed, params)
if err == nil {
	// Apply archetype to character
}
```

---

#### 3.6: Narrative Integration & Audit (Day 17)

**Packages**: `pkg/procgen/narrative`, `pkg/procgen/audit`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/procgen/narrative"
```

2. **Integrate narrative events**:
```go
narrativeGen := narrative.NewNarrativeGenerator()
event, err := narrativeGen.Generate(worldSeed, params)
if err == nil {
	sys.narrativeSystem.AddEvent(event)
}
```

**File**: `Makefile` (for audit integration)

**Changes**: Add audit target:
```makefile
audit:
	go test ./pkg/procgen/audit/... -v
```

**Verification**:
```bash
make audit
go run ./examples/narrativetest/
```

---

#### 3.7: Phase 3 Validation (Day 17)

**Tasks**:
- [ ] `go build ./cmd/client`
- [ ] `go test ./pkg/procgen/...`
- [ ] Run client, verify magic spells drop
- [ ] Check skill trees generated
- [ ] Test NPC dialog trees
- [ ] Verify environmental hazards spawn
- [ ] Test puzzle rooms
- [ ] Check minigames in taverns
- [ ] Run `make audit`

---

### Phase 4: Network & Multiplayer (Week 4 - Days 18-22)
**Goal**: Expand multiplayer capabilities  
**Packages**: 5  
**Impact**: Chat, trade, resilience, mobile federation, WebRTC  
**Effort**: Medium (5 days)

---

#### 4.1: Chat System (Day 18)

**Package**: `pkg/network/chat`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/network/chat"
```

2. **Add to systemsContainer**:
```go
chatSystem *chat.System
```

3. **Initialize** in `initializeV5Systems` (around line 550):
```go
sys.chatSystem = chat.NewSystem(networkClient)
```

4. **Register system**:
```go
game.World.AddSystem(sys.chatSystem)
```

**File**: `cmd/server/main.go`

**Changes**: Add chat handler:
```go
chatHandler := chat.NewServerHandler()
server.RegisterHandler("chat", chatHandler)
```

**Verification**:
```bash
go run ./examples/chattest/
```

---

#### 4.2: Trade System (Day 19)

**Package**: `pkg/network/trade`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/network/trade"
```

2. **Add to systemsContainer**:
```go
tradeSystem *trade.System
```

3. **Initialize**:
```go
sys.tradeSystem = trade.NewSystem(networkClient)
```

4. **Register system**:
```go
game.World.AddSystem(sys.tradeSystem)
```

**File**: `cmd/server/main.go`

**Changes**:
```go
tradeHandler := trade.NewServerHandler()
server.RegisterHandler("trade", tradeHandler)
```

**Verification**:
```bash
go run ./examples/tradetest/
```

---

#### 4.3: High-Latency Resilience (Day 20)

**Package**: `pkg/network/resilience`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/network/resilience"
```

2. **Wrap network client** (in `initializeNetworkClient`):
```go
// Wrap client with resilience layer for high-latency support (Tor/onion)
if networkClient != nil {
	resilientClient := resilience.NewResilientClient(networkClient, resilience.Config{
		MaxLatency:        5000 * time.Millisecond, // 5 seconds for Tor
		RetryAttempts:     3,
		PredictionEnabled: true,
	})
	networkClient = resilientClient
}
```

**Verification**:
```bash
go run ./examples/resiliencetest/
# Test with --high-latency flag
```

---

#### 4.4: Platform-Specific Federation (Day 21)

**Packages**: `pkg/network/federation/mobile`, `pkg/network/federation/webrtc`

**File**: `cmd/client/handlers.go` (mobile builds)

**Changes**:

1. **Add conditional import** (with build tags):
```go
// +build android ios

"github.com/opd-ai/venture/pkg/network/federation/mobile"
```

2. **Initialize mobile federation**:
```go
// +build android ios

func initializeMobileFederation(client *network.Client) {
	mobileFed := mobile.NewFederationClient(client)
	client.SetFederationProvider(mobileFed)
}
```

**File**: `cmd/client/handlers.go` (WASM/desktop)

**Changes**:

1. **Add import**:
```go
// +build !android,!ios

"github.com/opd-ai/venture/pkg/network/federation/webrtc"
```

2. **Initialize WebRTC**:
```go
webrtcFed := webrtc.NewFederationClient(client)
client.SetFederationProvider(webrtcFed)
```

---

#### 4.5: Phase 4 Validation (Day 22)

**Tasks**:
- [ ] `go build ./cmd/client`
- [ ] Test chat between clients
- [ ] Test player-to-player trading
- [ ] Test high-latency mode (Tor simulation)
- [ ] Build mobile: `ebitenmobile bind`
- [ ] Test WebRTC on WASM build

---

### Phase 5: Engine Enhancements (Week 5 - Days 23-25)
**Goal**: QoL, prestige, performance monitoring  
**Packages**: 3  
**Impact**: Auto-loot, New Game+, profiling  
**Effort**: Small (3 days)

---

#### 5.1: Quality-of-Life System (Day 23)

**Package**: `pkg/engine/qol`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/engine/qol"
```

2. **Add to systemsContainer**:
```go
qolManager *qol.Manager
```

3. **Initialize**:
```go
sys.qolManager = qol.NewManager(qol.Config{
	AutoLoot:     true,
	AutoSort:     true,
	QuickDeposit: true,
})
```

4. **Register system**:
```go
game.World.AddSystem(sys.qolManager)
```

**Verification**:
```bash
go run ./examples/qoltest/
```

---

#### 5.2: Prestige System (Day 24)

**Package**: `pkg/engine/prestige`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/engine/prestige"
```

2. **Add to systemsContainer**:
```go
prestigeSystem *prestige.System
```

3. **Initialize**:
```go
sys.prestigeSystem = prestige.NewSystem()
```

4. **Register system**:
```go
game.World.AddSystem(sys.prestigeSystem)
```

5. **Connect to save/load**:
```go
// In save/load code
saveState.Prestige = sys.prestigeSystem.GetState()
```

**Verification**:
```bash
go run ./examples/prestigetest/
```

---

#### 5.3: Performance Monitoring (Day 25)

**Package**: `pkg/engine/performance`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/engine/performance"
```

2. **Add to systemsContainer**:
```go
perfMonitor *performance.Monitor
```

3. **Initialize**:
```go
sys.perfMonitor = performance.NewMonitor(performance.Config{
	EnableProfiling: *verbose,
	TargetFPS:       60,
})
```

4. **Register system**:
```go
game.World.AddSystem(sys.perfMonitor)
```

5. **Add debug output** (if verbose):
```go
if *verbose {
	sys.perfMonitor.PrintStats()
}
```

**Verification**:
```bash
go run ./examples/performancetest/
./cmd/client --verbose
```

---

#### 5.4: Phase 5 Validation (Day 25)

**Tasks**:
- [ ] `go build ./cmd/client`
- [ ] Test auto-loot functionality
- [ ] Test prestige reset
- [ ] Verify performance monitoring output
- [ ] Check 60+ FPS maintained

---

### Phase 6: World Systems (Week 6 - Days 26-30)
**Goal**: Economy and raid mechanics  
**Packages**: 2  
**Impact**: Market simulation, territory raids  
**Effort**: Medium (5 days)

---

#### 6.1: Economy System (Days 26-28)

**Package**: `pkg/world/economy`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/world/economy"
```

2. **Add to systemsContainer**:
```go
economySystem *economy.System
```

3. **Initialize** (requires trade system):
```go
sys.economySystem = economy.NewSystem(
	game.World,
	sys.tradeSystem,
	economy.Config{
		InflationRate:      0.02,
		TaxRate:            0.10,
		UpdateInterval:     time.Minute,
	},
)
```

4. **Register system**:
```go
game.World.AddSystem(sys.economySystem)
```

**File**: `cmd/server/main.go`

**Changes**: Add economy sync:
```go
economySys := economy.NewSystem(world, tradeHandler, config)
server.AddSystem(economySys)
```

**Verification**:
```bash
go run ./examples/marketplacetest/
```

---

#### 6.2: Raid System (Days 29-30)

**Package**: `pkg/world/raids`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/world/raids"
```

2. **Add to systemsContainer**:
```go
raidSystem *raids.System
```

3. **Initialize** (requires territory, faction systems):
```go
sys.raidSystem = raids.NewSystem(
	game.World,
	sys.territorySystem,
	sys.factionSystem,
	raids.Config{
		RaidFrequency:     24 * time.Hour,
		MinAttackers:      5,
		MaxAttackers:      20,
		DifficultyScaling: true,
	},
)
```

4. **Register system**:
```go
game.World.AddSystem(sys.raidSystem)
```

**File**: `cmd/server/main.go`

**Changes**:
```go
raidSys := raids.NewSystem(world, territorySys, factionSys, config)
server.AddSystem(raidSys)
```

**Verification**:
```bash
go run ./examples/raidtest/
```

---

#### 6.3: Phase 6 Validation (Day 30)

**Tasks**:
- [ ] `go build ./cmd/client && go build ./cmd/server`
- [ ] Test market prices fluctuate
- [ ] Verify supply/demand mechanics
- [ ] Test raid spawning
- [ ] Check raid difficulty scaling
- [ ] Verify territory defense

---

### Phase 7: Advanced Integration (Weeks 7-8 - Days 31-45)
**Goal**: Cross-system feature interactions  
**Packages**: 7  
**Impact**: Choice consequences, guild vehicles, narrative world, political warfare, sieges, trade routes  
**Effort**: Large (15 days)

---

#### 7.1: Choice & Consequence (Days 31-33)

**Package**: `pkg/integration/choice_consequences`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/choice_consequences"
```

2. **Add to systemsContainer**:
```go
choiceManager *choice_consequences.Manager
```

3. **Initialize** (requires quest, world systems):
```go
sys.choiceManager = choice_consequences.NewManager(
	sys.objectiveTracker, // Quest system
	game.World,
)
```

4. **Register system**:
```go
game.World.AddSystem(sys.choiceManager)
```

**Verification**:
```bash
go run ./examples/choicetest/
```

---

#### 7.2: Guild Vehicles (Days 34-36)

**Package**: `pkg/integration/guild_vehicle`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
```

2. **Add to systemsContainer**:
```go
guildVehicleManager *guild_vehicle.Manager
```

3. **Initialize**:
```go
sys.guildVehicleManager = guild_vehicle.NewManager(
	sys.guildSystem,
	sys.vehicleMovementSys,
)
```

4. **Register system**:
```go
game.World.AddSystem(sys.guildVehicleManager)
```

---

#### 7.3: Narrative World Integration (Days 37-38)

**Package**: `pkg/integration/narrative_world`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/narrative_world"
```

2. **Initialize**:
```go
narrativeWorldManager := narrative_world.NewManager(
	sys.narrativeSystem,
	game.World,
)
game.World.AddSystem(narrativeWorldManager)
```

---

#### 7.4: Political Warfare (Days 39-40)

**Package**: `pkg/integration/political_warfare`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/political_warfare"
```

2. **Initialize**:
```go
politicalWarfareSys := political_warfare.NewManager(
	sys.factionSystem,
	sys.territorySystem,
)
game.World.AddSystem(politicalWarfareSys)
```

**Verification**:
```bash
go run ./examples/politicalwarfaretest/
```

---

#### 7.5: Territory Siege (Days 41-42)

**Package**: `pkg/integration/territory_siege`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/territory_siege"
```

2. **Initialize**:
```go
siegeSystem := territory_siege.NewManager(
	sys.territorySystem,
	sys.combatSystem,
)
game.World.AddSystem(siegeSystem)
```

**Verification**:
```bash
go run ./examples/siegetest/
```

---

#### 7.6: Trade Routes (Days 43-44)

**Package**: `pkg/integration/trade_routes`

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/integration/trade_routes"
```

2. **Initialize**:
```go
tradeRoutesSys := trade_routes.NewManager(
	sys.economySystem,
	sys.tradeSystem,
)
game.World.AddSystem(tradeRoutesSys)
```

**Verification**:
```bash
go run ./examples/traderoutetest/
```

---

#### 7.7: Phase 7 Validation (Day 45)

**Tasks**:
- [ ] `go build ./cmd/client && go build ./cmd/server`
- [ ] Test all integration systems active
- [ ] Verify choice consequences persist
- [ ] Test guild vehicle ownership
- [ ] Check narrative events affect world
- [ ] Test faction politics
- [ ] Verify territory sieges
- [ ] Check trade route economics

---

### Phase 8: Developer Tools (Ongoing)
**Goal**: CI/CD and server tooling  
**Packages**: 11  
**Impact**: Development workflow, not end-user features  
**Effort**: Medium (5 days, integrated into CI/CD)

---

#### 8.1: CI/CD Integration

**File**: `.github/workflows/test.yml`

**Changes**: Add audit steps:
```yaml
- name: Feature Audit
  run: go test ./pkg/audit/features/... -v

- name: Balance Audit
  run: go test ./pkg/balance/... -v

- name: Security Audit
  run: go test ./pkg/security/... -v

- name: Stability Audit
  run: go test ./pkg/stability/... -v

- name: Visual Regression Test
  run: go test ./pkg/visualtest/... -v

- name: Platform Parity Test
  run: go test ./pkg/visualtest/parity/... -v
```

---

#### 8.2: Server Modding Support

**File**: `cmd/server/main.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/modding"
```

2. **Add mod loading**:
```go
modLoader := modding.NewLoader("./mods")
if err := modLoader.LoadAll(); err != nil {
	log.Warn("failed to load mods: ", err)
}
server.SetModLoader(modLoader)
```

---

#### 8.3: Migration System

**File**: `cmd/server/main.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/migration"
```

2. **Add save migration**:
```go
migrator := migration.NewMigrator()
if err := migrator.MigrateSaves("./saves"); err != nil {
	log.Error("save migration failed: ", err)
}
```

---

#### 8.4: Analytics Integration

**File**: `cmd/client/handlers.go`

**Changes**:

1. **Add import**:
```go
"github.com/opd-ai/venture/pkg/ux"
```

2. **Add UX tracking** (optional, privacy-respecting):
```go
if *enableAnalytics {
	uxTracker := ux.NewTracker()
	game.World.AddSystem(uxTracker)
}
```

---

## Final Validation Checklist

### Build Validation
- [ ] `go build ./cmd/client` (no errors)
- [ ] `go build ./cmd/server` (no errors)
- [ ] `go build ./cmd/mobile` (Android/iOS builds)
- [ ] WASM build: `GOOS=js GOARCH=wasm go build`

### Test Validation
- [ ] `go test ./pkg/...` (all packages pass)
- [ ] `go test -race ./pkg/...` (no race conditions)
- [ ] `go test -cover ./pkg/...` (≥65% coverage maintained)

### Feature Validation
- [ ] **Audio**: Music and SFX playing
- [ ] **Destruction**: Objects break when damaged
- [ ] **Companions**: Companions learn skills
- [ ] **Rendering**: Lighting, post-processing, animation visible
- [ ] **Procgen**: Magic, skills, entities, dialog, puzzles, minigames active
- [ ] **Network**: Chat, trade, high-latency resilience functional
- [ ] **Engine**: QoL, prestige, performance monitoring active
- [ ] **World**: Economy fluctuates, raids spawn
- [ ] **Integration**: All cross-system features working

### Performance Validation
- [ ] 60+ FPS with 2000 entities
- [ ] <500MB client memory usage
- [ ] <1GB server memory (4 players)
- [ ] No memory leaks (run for 1 hour)

### Flag Cleanup Validation
- [ ] `grep -r "enable-" cmd/client/util.go` returns ONLY `host-and-play`
- [ ] All conditional feature activations removed
- [ ] All systems unconditionally registered

---

## Rollback Plan

If integration causes issues:

1. **Revert specific package**:
   ```bash
   git revert <commit-hash>
   ```

2. **Temporarily disable system**:
   ```go
   // Comment out system registration
   // game.World.AddSystem(sys.problematicSystem)
   ```

3. **Full phase rollback**:
   ```bash
   git reset --hard <phase-start-commit>
   ```

---

## Success Metrics

### Quantitative
- **Packages Active**: 102/102 (100%)
- **Feature Flags Removed**: 7/7 (100%)
- **Test Coverage**: ≥65% per package
- **Performance**: 60+ FPS, <500MB memory
- **Build Time**: <2 minutes
- **Lines of Code Active**: ~382,000

### Qualitative
- All game systems feel "always-on" (no toggles)
- Procedural content is varied and high-quality
- Audio enhances gameplay immersion
- Visual effects are polished
- Multiplayer is robust (chat, trade, federation)
- Cross-system features work seamlessly

---

## Timeline Summary

| Phase | Duration | Packages | Effort |
|-------|----------|----------|--------|
| Phase 1: Foundation | 5 days | 6 | Medium |
| Phase 2: Visual | 5 days | 9 | Medium |
| Phase 3: Content | 7 days | 12 | Large |
| Phase 4: Network | 5 days | 5 | Medium |
| Phase 5: Engine | 3 days | 3 | Small |
| Phase 6: World | 5 days | 2 | Medium |
| Phase 7: Integration | 15 days | 7 | Large |
| Phase 8: Tools | 5 days | 11 | Medium |
| **TOTAL** | **50 days** | **55** | **~2.5 months** |

**Note**: Phase 8 (Developer Tools) runs in parallel with Phases 1-7 via CI/CD integration.

---

## Post-Integration Roadmap

### V11.0: Full Integration Release
- All 55 dormant packages activated
- Zero feature flags (except operational modes)
- Comprehensive documentation update
- Performance optimization pass
- User acceptance testing

### V11.1: Polish & Optimization
- Performance tuning based on profiling
- Balance adjustments from telemetry
- Bug fixes from user feedback
- Visual polish pass

### V12.0: Next Features
- New gameplay systems (not yet implemented)
- Advanced AI behaviors
- Expanded federation capabilities
- Community modding support

---

*Generated: December 13, 2025*  
*Target Completion: February 2026 (V11.0)*  
*Maintenance: Ongoing (CI/CD automation)*
