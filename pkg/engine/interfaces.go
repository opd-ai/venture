// Package engine provides the core ECS (Entity-Component-System) interfaces.
// This file contains the fundamental interfaces for the ECS architecture and
// dependency injection interfaces that enable testability by abstracting
// Ebiten-specific implementations.
//
// Interface Design Principles:
//   - Production implementations use Ebiten types (*ebiten.Image, input handling, etc.)
//   - Test implementations use simple stubs without external dependencies
//   - All interfaces support both production and test scenarios
//   - No build tags required - *_test.go suffix provides automatic test-only code
package engine

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/engine/ai/behavior"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/saveload"
)

// Component represents a data container attached to an Entity.
// Components should be pure data structures without behavior.
// Originally from: ecs.go
type Component interface {
	// Type returns a unique identifier for this component type
	Type() string
}

// System represents a behavior that operates on entities with specific components.
// Systems should be stateless where possible and operate on entity data.
// Originally from: ecs.go
type System interface {
	// Update is called every frame to update entities managed by this system
	Update(entities []*Entity, deltaTime float64)
}

// GameRunner manages the main game loop and state.
// This interface abstracts the game from Ebiten-specific implementation details.
//
// Implementations:
//   - EbitenGame: Production implementation that integrates with Ebiten game loop
//   - StubGame: Test implementation for unit testing without Ebiten dependencies
type GameRunner interface {
	// GetWorld returns the ECS world instance containing all entities and systems
	GetWorld() *World

	// GetScreenSize returns the current screen dimensions in pixels
	GetScreenSize() (width, height int)

	// IsPaused returns whether the game is currently paused
	IsPaused() bool

	// SetPaused sets the game pause state
	SetPaused(paused bool)

	// SetPlayerEntity sets the player entity reference for UI systems and game logic
	SetPlayerEntity(entity *Entity)

	// GetPlayerEntity returns the current player entity (can be nil)
	GetPlayerEntity() *Entity

	// Update is called every frame to update game state
	// Returns an error to comply with ebiten.Game interface requirements
	Update() error

	// SetInventorySystem connects the inventory system to the UI for player interaction
	SetInventorySystem(system *InventorySystem)

	// SetupInputCallbacks connects input system callbacks to UI systems
	// H-008 FIX: Returns error if callback registration fails
	SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) error
}

// ImageProvider provides access to image data without Ebiten-specific types.
// This allows image handling to be abstracted for testing.
//
// Implementations:
//   - EbitenImage: Wraps *ebiten.Image for production use
//   - StubImage: Simple test implementation with size and color data
type ImageProvider interface {
	// GetSize returns the image dimensions in pixels
	GetSize() (width, height int)

	// GetPixel returns the color at the given position
	// Returns color.Transparent if position is out of bounds
	GetPixel(x, y int) color.Color
}

// DrawOptions contains optional parameters for rendering operations.
// All fields are optional and have sensible defaults.
type DrawOptions struct {
	Rotation float64 // Rotation in radians
	ScaleX   float64 // Horizontal scale factor (1.0 = no scaling)
	ScaleY   float64 // Vertical scale factor (1.0 = no scaling)
	Alpha    float32 // Alpha transparency (0.0 = transparent, 1.0 = opaque)
	OffsetX  float64 // X offset from position
	OffsetY  float64 // Y offset from position
}

// Renderer handles drawing of visual elements to the screen.
// This interface abstracts rendering operations from Ebiten-specific implementation.
//
// Implementations:
//   - EbitenRenderer: Production renderer using *ebiten.Image operations
//   - StubRenderer: Test renderer that records draw calls for verification
type Renderer interface {
	// DrawImage draws an image at the specified position with optional transformations
	DrawImage(img ImageProvider, x, y float64, opts *DrawOptions)

	// DrawRect draws a filled rectangle at the specified position and size
	DrawRect(x, y, width, height float64, col color.Color)

	// DrawText draws text at the specified position
	// Position is in screen coordinates (top-left origin)
	DrawText(text string, x, y int, col color.Color)

	// Clear clears the entire screen with the specified color
	Clear(col color.Color)

	// GetBounds returns the rendering bounds (screen size) in pixels
	GetBounds() (width, height int)
}

// SpriteProvider provides sprite visual data without Ebiten dependencies.
// Sprites represent the visual appearance of entities in the game.
//
// Implementations:
//   - EbitenSprite: Production sprite with *ebiten.Image
//   - StubSprite: Test sprite with just properties (no actual image data)
type SpriteProvider interface {
	Component // Inherits Type() string

	// GetImage returns the sprite's image data
	// Returns nil if no image is set (will use solid color rendering)
	GetImage() ImageProvider

	// GetSize returns the sprite dimensions in game units (not pixels)
	GetSize() (width, height float64)

	// GetColor returns the sprite tint color
	// This color is multiplied with the image (or used as solid color if no image)
	GetColor() color.Color

	// GetRotation returns the sprite rotation in radians
	GetRotation() float64

	// GetLayer returns the rendering layer (higher values render on top)
	// Typical range: 0 (background) to 100 (foreground/UI)
	GetLayer() int

	// IsVisible returns whether the sprite should be rendered
	IsVisible() bool

	// SetVisible sets the sprite visibility state
	SetVisible(visible bool)

	// SetColor sets the sprite tint color
	SetColor(col color.Color)

	// SetRotation sets the sprite rotation in radians
	SetRotation(rotation float64)
}

// InputProvider provides access to player input state.
// This interface abstracts input handling from Ebiten's keyboard/mouse implementation.
//
// Implementations:
//   - EbitenInput: Production input reading from ebiten.IsKeyPressed, ebiten.CursorPosition
//   - StubInput: Test input with manually controllable state for deterministic testing
type InputProvider interface {
	Component // Inherits Type() string

	// GetMovement returns normalized movement input on both axes
	// Returns values in range [-1.0, 1.0] for each axis
	// (0, 0) means no movement input
	GetMovement() (x, y float64)

	// IsActionPressed returns whether the primary action button is currently held down
	// Typically mapped to SPACE or mouse click
	IsActionPressed() bool

	// IsActionJustPressed returns whether the action button was pressed THIS frame
	// Used for actions that should trigger once per press (attacks, confirmations)
	IsActionJustPressed() bool

	// IsAnyKeyPressed returns whether any key was pressed this frame
	// Used for "press any key to continue" interactions
	IsAnyKeyPressed() bool

	// IsUseItemPressed returns whether the use item button is currently held down
	IsUseItemPressed() bool

	// IsUseItemJustPressed returns whether use item was pressed THIS frame
	IsUseItemJustPressed() bool

	// IsSpellPressed returns whether a spell hotkey (1-5) is currently pressed
	// slot must be in range [1, 5]
	IsSpellPressed(slot int) bool

	// GetMousePosition returns the current mouse cursor position in screen coordinates
	GetMousePosition() (x, y int)

	// GetMouseDelta returns the mouse movement since the last frame
	// Gap #8 fix: Essential for first-person camera control and aiming
	// Returns the change in X and Y coordinates from the previous frame
	GetMouseDelta() (dx, dy int)

	// IsMousePressed returns whether the primary mouse button is pressed
	IsMousePressed() bool

	// SetMovement sets the movement input (primarily for testing)
	// Values should be in range [-1.0, 1.0]
	SetMovement(x, y float64)

	// SetActionPressed sets the action button state (primarily for testing)
	SetActionPressed(pressed bool)

	// Menu navigation methods for UI abstraction
	// These allow UI components to handle input without direct Ebiten dependencies

	// IsMenuUpJustPressed returns whether the menu up key was just pressed
	IsMenuUpJustPressed() bool

	// IsMenuDownJustPressed returns whether the menu down key was just pressed
	IsMenuDownJustPressed() bool

	// IsMenuConfirmJustPressed returns whether the menu confirm key was just pressed
	IsMenuConfirmJustPressed() bool

	// IsMenuBackJustPressed returns whether the menu back/cancel key was just pressed (ESC)
	IsMenuBackJustPressed() bool

	// IsMenuTabJustPressed returns whether the menu tab key was just pressed
	IsMenuTabJustPressed() bool

	// GetTouchIDs returns the list of current active touch IDs
	// Returns empty slice if no touches are active
	// Used for multi-touch input detection on mobile and touch-capable devices
	GetTouchIDs() []ebiten.TouchID

	// GetTouchPosition returns the screen position of the touch with the given ID
	// Returns (0, 0) if the touch ID is not found
	// x, y are in screen coordinates
	GetTouchPosition(id ebiten.TouchID) (x, y int)
}

// RenderingSystem handles visual rendering of entities.
// This interface abstracts the render system from Ebiten-specific drawing operations.
//
// Implementations:
//   - EbitenRenderSystem: Production renderer using Ebiten drawing primitives
//   - StubRenderSystem: Test renderer for verifying render logic without actual drawing
type RenderingSystem interface {
	System // Inherits Update(entities []*Entity, deltaTime float64)

	// Draw renders all visible entities to the screen
	// The screen parameter is interface{} to support both *ebiten.Image (production)
	// and Renderer interface (testing). Concrete implementations should type-assert.
	Draw(screen interface{}, entities []*Entity)

	// SetShowColliders enables or disables debug visualization of collision boxes
	SetShowColliders(show bool)

	// SetShowGrid enables or disables debug visualization of spatial grid
	SetShowGrid(show bool)
}

// UISystem handles user interface rendering and interaction.
// This interface provides a common contract for all UI systems (HUD, menus, dialogs, etc.)
//
// Implementations:
//   - Production: EbitenHUDSystem, EbitenMenuSystem, EbitenCharacterUI, etc.
//   - Test: StubHUDSystem, StubMenuSystem, StubCharacterUI, etc.
type UISystem interface {
	System // Inherits Update(entities []*Entity, deltaTime float64)

	// Draw renders the UI to the screen
	// The screen parameter is interface{} to support both *ebiten.Image (production)
	// and Renderer interface (testing)
	Draw(screen interface{})

	// IsActive returns whether the UI is currently visible and processing input
	IsActive() bool

	// SetActive sets the UI visibility and input processing state
	SetActive(active bool)

	// Note: HandleInput(input InputProvider) bool is optional and not part of the base
	// interface. Individual UI systems can implement it if they need direct input handling.
	// Return value indicates whether the UI consumed the input (blocking lower layers).
}

// Vehicle provides behavior interface for mounted vehicles and mounts.
// This interface abstracts vehicle physics and capabilities for different vehicle types
// (mounts, carts, boats, gliders, mechs).
//
// Phase 21.1: Vehicle Foundation
type Vehicle interface {
	// GetSpeed returns the current speed of the vehicle
	GetSpeed() float64

	// GetMaxSpeed returns the maximum speed the vehicle can achieve
	GetMaxSpeed() float64

	// GetAcceleration returns the acceleration rate
	GetAcceleration() float64

	// GetHandling returns the turn rate/handling ability
	GetHandling() float64

	// CanTraverse checks if the vehicle can move on a specific terrain type
	// Uses terrain.TileType from pkg/procgen/terrain
	CanTraverse(terrainType int) bool

	// GetFuelCost returns resource consumption per tile traveled
	GetFuelCost() float64
}

// VehicleController manages mounting/dismounting interactions with vehicles.
// This interface enables player and NPC entities to enter and exit vehicles.
//
// Phase 21.1: Vehicle Foundation
type VehicleController interface {
	// Mount attempts to mount a rider entity onto a vehicle entity
	// Returns error if mounting fails (vehicle full, incompatible, etc.)
	Mount(rider, vehicle *Entity) error

	// Dismount removes a rider entity from their current vehicle
	// Returns error if rider is not mounted
	Dismount(rider *Entity) error

	// IsMounted checks if an entity is currently riding a vehicle
	IsMounted(rider *Entity) bool

	// GetMountedVehicle returns the vehicle entity the rider is on
	// Returns nil if rider is not mounted
	GetMountedVehicle(rider *Entity) *Entity
}

// CharacterClassInfo provides metadata and mechanics for character classes.
// This interface abstracts class-specific information and stat calculations
// for the character class system.
//
// Phase 25.1: Character Classes Foundation
type CharacterClassInfo interface {
	// GetName returns the human-readable class name
	GetName() string

	// GetDescription returns a description of the class playstyle
	GetDescription() string

	// GetStartingStats returns base stats for a new character of this class
	// Returns values for HP, Mana, Attack, Defense, Magic Power, Speed
	GetStartingStats() map[string]float64

	// GetClassAbilities returns the list of ability IDs available to this class
	// Returns at least 8 base abilities per class
	GetClassAbilities() []string

	// GetStatGrowth returns the per-level stat scaling configuration
	// Defines how stats increase as the character levels up
	GetStatGrowth() *StatGrowth
}

// Expression represents a player emote or gesture with associated visual and audio effects.
// Expressions enable non-verbal communication between players in multiplayer mode.
//
// Phase 26.1: Expression Framework
type Expression interface {
	// GetAnimation returns the procedural animation sequence for this expression
	// The animation modifies sprite frames over time to create the gesture
	GetAnimation() AnimationSequence

	// GetSoundEffect returns the audio effect ID to play when expression triggers
	// Empty string means no sound effect
	GetSoundEffect() string

	// GetDuration returns how long the expression animation lasts in seconds
	// Returns math.Inf(1) for expressions that last until canceled (Sit, Sleep)
	GetDuration() float64
}

// AnimationSequence represents a procedural animation for expressions and other effects.
// The sequence defines keyframes and interpolation for sprite transformations.
//
// Phase 26.1: Expression Framework
type AnimationSequence interface {
	// GetFrameCount returns the number of animation frames
	GetFrameCount() int

	// GetFrameTime returns the duration of each frame in seconds
	GetFrameTime() float64

	// ShouldLoop returns whether the animation repeats
	ShouldLoop() bool
}

// MiniGame represents an embedded procedural mini-game that can be played within Venture.
// Mini-games are deterministically generated using a seed and provide rewards upon completion.
//
// Design: Mini-games follow ECS data-driven rendering pattern. They compute visual state
// via PrepareRender() and expose it through GetRenderOutput(). A separate ECS rendering
// system consumes this data to draw actual pixels. This separation enables:
//   - Clean separation between game logic and rendering
//   - Easy testing without graphics initialization
//   - Multiplayer sync of visual state
//   - Data-driven UI rendering
//
// Phase 27.1: Mini-Game Framework
type MiniGame interface {
	// Initialize sets up the mini-game with the given seed and difficulty
	// Returns an error if initialization fails (invalid parameters, etc.)
	Initialize(seed int64, difficulty float64) error

	// Update advances the mini-game state by deltaTime seconds
	// Returns an error if the update fails
	Update(deltaTime float64) error

	// PrepareRender computes the current visual state for the mini-game
	// This validates screen dimensions and populates internal render data
	// Returns an error if screen is invalid or game is not initialized
	//
	// NOTE: This does NOT draw pixels - it only prepares render data.
	// Use GetRenderOutput() to retrieve the computed visual state.
	PrepareRender(screenWidth, screenHeight int) error

	// GetRenderOutput returns the computed visual state from the last PrepareRender call
	// Returns nil if PrepareRender has not been called yet
	// The render system reads this data to draw actual pixels to the screen
	GetRenderOutput() MiniGameRenderOutput

	// IsComplete returns true when the mini-game has finished (won or lost)
	IsComplete() bool

	// GetReward returns the reward earned from completing the mini-game
	// Returns nil if the game is not complete or if the player lost
	GetReward() *Reward
}

// MiniGameRenderOutput provides visual state data for rendering a mini-game.
// This interface abstracts render output to avoid import cycles between engine and procgen packages.
// Implementations should provide title, status, dimensions, and visual elements.
//
// Phase 27.3: Mini-Game Rendering
type MiniGameRenderOutput interface {
	// GetTitle returns the display name of the mini-game
	GetTitle() string
	// GetStatus returns the current game state ("Playing", "Won", "Lost")
	GetStatus() string
	// GetDimensions returns the screen width and height in pixels
	GetDimensions() (width, height int)
	// GetElements returns the visual elements to be drawn (implementation-specific)
	// The rendering system casts this to the concrete type it expects
	GetElements() interface{}
}

// HousingUIProvider is the interface for housing UI panels.
// This interface avoids import cycles between engine and housing packages.
//
// INTEGRATION FIX [Category B]: V8.0 Housing UI Integration (Phase 49.1)
// Gap: InputSystem needs HousingUI reference for ESC key handling without importing housing package
// Fix: Define interface for housing UI visibility and hide operations
type HousingUIProvider interface {
	// IsVisible returns true if the housing UI is currently visible
	IsVisible() bool

	// Hide hides the housing UI
	Hide()

	// Toggle switches the housing UI visibility
	Toggle()
}

// PrestigeUIProvider provides a subset of prestige UI methods for input system integration.
// This interface avoids circular dependencies between pkg/engine and pkg/engine/prestige.
// StoryJournalProvider is the interface for the story journal UI.
// Stored on EbitenGame to avoid a circular import with pkg/rendering/ui
// (which imports pkg/engine). Implemented by storyJournalWrapper in cmd/client.
type StoryJournalProvider interface {
	// Toggle shows or hides the journal UI.
	Toggle(journal *StoryJournalComponent, world *World)
	// IsVisible returns true when the journal is currently visible.
	IsVisible() bool
	// Draw renders the journal onto the screen.
	Draw(screen interface{})
}

type PrestigeUIProvider interface {
	// IsVisible returns true if the prestige UI is currently visible
	IsVisible() bool

	// Hide hides the prestige UI
	Hide()

	// Show displays the prestige UI for a specific player
	Show(playerID, className string)
}

// BehaviorNode is a type alias for behavior.BehaviorNode.
// Defining it here as an alias (rather than a new named interface) means that
// []BehaviorNode and []behavior.BehaviorNode are the same type, so the wrapper
// constructors (NewSequenceNode, etc.) in behavior_tree_nodes.go can forward
// variadic slices to the behavior package without explicit conversion.
type BehaviorNode = behavior.BehaviorNode

// Synthesizer defines the interface for audio synthesis engines.
// This allows AudioManagerSystem to use different synthesis implementations.
// Originally from: audio_manager.go
type Synthesizer interface {
	GetSampleRate() int
	GetSeed() int64
}

// DialogProvider generates dialog content for NPCs.
// This interface enables extensibility for branching dialogs, quest dialogs,
// and dynamic content based on game state.
// Originally from: commerce_components.go
type DialogProvider interface {
	// GetDialog returns the current dialog text and available options.
	GetDialog() (text string, options []DialogOption)
}

// TransactionValidator validates commerce transactions.
// This interface enables server-authoritative validation in multiplayer
// and extensibility for reputation systems, barter, and trade quests.
// Originally from: commerce_components.go
type TransactionValidator interface {
	// CanBuyItem checks if player can purchase an item from merchant.
	// Returns true and empty string if valid, false and error message if invalid.
	CanBuyItem(playerGold, itemPrice int, playerInventoryFull bool) (bool, string)

	// CanSellItem checks if player can sell an item to merchant.
	// Returns true and empty string if valid, false and error message if invalid.
	CanSellItem(merchantGold, itemPrice int, merchantInventoryFull bool) (bool, string)
}

// QuestGeneratorInterface defines the interface for quest generation.
// This allows for dependency injection and testing.
// Originally from: discovery_system.go
type QuestGeneratorInterface interface {
	Generate(seed int64, params procgen.GenerationParams) (interface{}, error)
}

// ComponentSerializer defines the interface for components that can be serialized.
// Originally from: entity_persistence.go
type ComponentSerializer interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

// GameClock provides time tracking for game simulation.
// Supports both deterministic (simulation-based) and real-time clocks.
// Originally from: game_clock.go
type GameClock interface {
	Now() time.Time
	Advance(deltaTime float64)
	Reset(startTime time.Time)
}

// FederationBroadcaster defines the interface for broadcasting guild updates.
// Originally from: guild_system.go
type FederationBroadcaster interface {
	BroadcastGuildUpdate(guildID string, guildData []byte) error
}

// VRHeadsetAdapter abstracts VR headset hardware for head tracking.
//
// EXPERIMENTAL: VR support is currently experimental. Production code uses
// StubHeadsetAdapter which reports no hardware connected. Future releases
// will integrate OpenVR/OpenXR SDKs for real hardware support.
//
// Implementations:
//   - StubHeadsetAdapter: Production stub (no hardware SDK)
//   - MockHeadset: Test implementation with configurable values
//
// Originally from: head_tracking_system.go
type VRHeadsetAdapter interface {
	// IsConnected returns true if headset is available
	IsConnected() bool

	// GetHeadOrientation returns pitch, yaw, roll in radians
	GetHeadOrientation() (pitch, yaw, roll float64)

	// GetHeadPosition returns head position offset in meters
	GetHeadPosition() (x, y, z float64)

	// GetIPD returns interpupillary distance in millimeters
	GetIPD() float64
}

// FileWatcher defines the interface for watching mod file changes.
// This allows for mock implementations in testing.
// Originally from: hot_reload_system.go
type FileWatcher interface {
	// GetFileHash returns the hash of a mod's files.
	GetFileHash(modID string) (string, error)

	// GetModData returns the mod data for reloading.
	GetModData(modID string) ([]byte, error)

	// GetModVersion returns the version of a mod.
	GetModVersion(modID string) (string, error)
}

// StateMigrationHandler handles state migration during reload.
// Originally from: hot_reload_system.go
type StateMigrationHandler interface {
	// SaveState saves the current mod state before reload.
	SaveState(modID string) (map[string]any, map[string]any, error)

	// RestoreState restores mod state after reload.
	RestoreState(modID string, scripts, variables map[string]any) error
}

// NetworkClient is an interface for getting network stats.
// This allows the HUD to display network status without depending on concrete implementation.
// Originally from: hud_system.go
type NetworkClient interface {
	GetLatency() time.Duration
	IsConnected() bool
}

// ModRepository defines the interface for fetching mods from a repository.
// This allows for mock implementations in testing and different backends.
// Originally from: mod_browser_system.go
type ModRepository interface {
	// FetchMods retrieves available mods from the repository.
	FetchMods() ([]ModListing, error)

	// DownloadMod downloads a mod by ID. Returns mod data on success.
	DownloadMod(modID string, progressCallback func(downloaded, total int64)) ([]byte, error)

	// GetModDetails retrieves detailed information for a specific mod.
	GetModDetails(modID string) (*ModListing, error)
}

// ImagePoolProvider abstracts image pool operations for memory efficiency.
// Implementations provide pooled image allocation and recycling.
// Originally from: render_system.go
type ImagePoolProvider interface {
	GetImage(width, height int) *ebiten.Image
	PutImage(img *ebiten.Image)
}

// ParallelRendererProvider abstracts parallel rendering capabilities.
// Implementations distribute rendering tasks across worker goroutines.
// Originally from: render_system.go
type ParallelRendererProvider interface {
	Start()
	Stop()
	IsRunning() bool
}

// VRControllerAdapter abstracts VR controller hardware.
//
// EXPERIMENTAL: VR support is currently experimental. Production code uses
// StubControllerAdapter which reports no controllers connected. Future releases
// will integrate OpenVR/OpenXR SDKs for real hardware support.
//
// Implementations:
//   - StubControllerAdapter: Production stub (no hardware SDK)
//   - MockController: Test implementation with configurable values
//
// Originally from: vr_controller_system.go
type VRControllerAdapter interface {
	// IsConnected returns true if controller is available
	IsConnected(hand string) bool

	// GetTrigger returns trigger value 0.0-1.0
	GetTrigger(hand string) float64

	// GetGrip returns grip value 0.0-1.0
	GetGrip(hand string) float64

	// GetThumbstick returns thumbstick X,Y values
	GetThumbstick(hand string) (x, y float64)

	// IsThumbstickPressed returns thumbstick click state
	IsThumbstickPressed(hand string) bool

	// GetButton returns button pressed state
	GetButton(hand, button string) bool

	// SetHaptic triggers haptic feedback
	SetHaptic(hand string, intensity, duration float64)
}

// ContextualTutorialProvider manages context-sensitive tutorial popups.
// This interface allows the game to control tutorial visibility based on
// settings without creating circular imports with rendering/ui package.
//
// Implementations:
//   - TutorialManager (pkg/rendering/ui/tutorial.go): Production implementation
//
// Phase 3.3 (PLAN.md): ShowTutorials Setting Wiring
// Phase 4.5 (PLAN.md): Tutorial Save/Load
type ContextualTutorialProvider interface {
	// Enable enables tutorial popups
	Enable()

	// Disable disables tutorial popups
	Disable()

	// IsEnabled returns tutorial enabled state
	IsEnabled() bool

	// ExportState serializes the tutorial state for save/load.
	// Phase 4.5: Enables persistence of viewed context-sensitive tutorials.
	ExportState() saveload.ContextTutorialStateData

	// ImportState restores tutorial state from saved data.
	// Phase 4.5: Restores viewed context-sensitive tutorials after load.
	ImportState(data saveload.ContextTutorialStateData)
}

// ModRuleProvider provides access to mod rules for game systems.
// This interface enables systems to query mod-defined rule values without
// creating a circular import with pkg/modding.
//
// Implementations:
//   - modding.ProviderAdapter (pkg/modding/adapter.go): Production implementation
//
// Phase 6.3 (PLAN.md): Modding System Integration
type ModRuleProvider interface {
	// GetRule retrieves the current value of a rule. Returns false if not found.
	GetRule(ruleName string) (interface{}, bool)

	// GetRuleFloat64 retrieves a rule as float64 with a default value.
	GetRuleFloat64(ruleName string, defaultValue float64) float64

	// GetRuleBool retrieves a rule as bool with a default value.
	GetRuleBool(ruleName string, defaultValue bool) bool

	// TriggerEvent triggers a mod event that registered handlers can respond to.
	// eventType is the event kind (e.g., "entity_spawn", "item_pickup").
	// eventData contains event-specific information.
	TriggerEvent(eventType string, eventData map[string]interface{}) error
}
