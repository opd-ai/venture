// Package engine provides player input handling.
// This file implements InputSystem which processes keyboard, mouse, and touch input
// for player-controlled entities and game controls.
package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/mobile"
)

// GameState represents the current state of the game for input filtering.
// Priority 2.3: State-based input filtering to prevent conflicts.
type GameState int

const (
	// StateExploring is the default gameplay state (movement, combat, exploration)
	StateExploring GameState = iota
	// StateCombat is active combat state (may have different input rules)
	StateCombat
	// StateMenu is when pause/settings menu is open (blocks gameplay input)
	StateMenu
	// StateDialogue is when talking to NPCs (blocks movement, allows choices)
	StateDialogue
	// StateInventory is when inventory UI is open (blocks movement)
	StateInventory
	// StateCharacterScreen is when character sheet is open
	StateCharacterScreen
	// StateSkillTree is when skill tree UI is open
	StateSkillTree
	// StateQuestLog is when quest log UI is open
	StateQuestLog
	// StateMap is when map UI is open
	StateMap
)

// String returns human-readable state name for debugging.
func (s GameState) String() string {
	switch s {
	case StateExploring:
		return "Exploring"
	case StateCombat:
		return "Combat"
	case StateMenu:
		return "Menu"
	case StateDialogue:
		return "Dialogue"
	case StateInventory:
		return "Inventory"
	case StateCharacterScreen:
		return "Character"
	case StateSkillTree:
		return "Skills"
	case StateQuestLog:
		return "Quests"
	case StateMap:
		return "Map"
	default:
		return "Unknown"
	}
}

// AllowsMovement returns true if player can move in this state.
func (s GameState) AllowsMovement() bool {
	switch s {
	case StateExploring, StateCombat:
		return true
	default:
		return false
	}
}

// AllowsCombat returns true if player can attack/cast spells in this state.
func (s GameState) AllowsCombat() bool {
	switch s {
	case StateExploring, StateCombat:
		return true
	default:
		return false
	}
}

// AllowsUIToggle returns true if UI can be opened/closed in this state.
func (s GameState) AllowsUIToggle() bool {
	// UI can be toggled in any state except dialogue (must finish dialogue first)
	return s != StateDialogue
}

// IsUIState returns true if this is a UI-focused state (not gameplay).
func (s GameState) IsUIState() bool {
	switch s {
	case StateMenu, StateInventory, StateCharacterScreen, StateSkillTree, StateQuestLog, StateMap:
		return true
	default:
		return false
	}
}

// AllowsInteraction returns true if player can interact with objects in this state.
// INPUT CONFLICT FIX: Added to prevent F key interactions when UIs are open.
func (s GameState) AllowsInteraction() bool {
	switch s {
	case StateExploring, StateCombat:
		return true
	default:
		return false
	}
}

// EbitenInput stores the current input state for an entity (Ebiten implementation).
// This is typically only used for player-controlled entities.
// Implements InputProvider interface.
//
// GAP-001/GAP-002 REPAIR: Separated immediate-consumption flags from persistent detection flags.
// ActionPressed/UseItemPressed are for immediate consumption (combat, actions).
// ActionJustPressed/UseItemJustPressed persist for the entire frame for UI/tutorial detection.
type EbitenInput struct {
	// Movement input (-1.0 to 1.0 for each axis)
	MoveX, MoveY float64

	// Action buttons - Immediate consumption (reset after first system uses them)
	ActionPressed   bool
	SecondaryAction bool
	UseItemPressed  bool

	// Action buttons - Frame-persistent detection (available to all systems this frame)
	// GAP-001 REPAIR: These flags persist for the full frame for tutorial/UI detection
	ActionJustPressed  bool // Set when action key first pressed this frame
	UseItemJustPressed bool // Set when use item key first pressed this frame
	AnyKeyPressed      bool // GAP-005 REPAIR: Set when any key pressed this frame

	// GAP-002 REPAIR: Spell casting input flags (keys 1-5)
	Spell1Pressed bool
	Spell2Pressed bool
	Spell3Pressed bool
	Spell4Pressed bool
	Spell5Pressed bool

	// Mouse state
	MouseX, MouseY int
	MousePressed   bool

	// Mouse delta (movement since last frame)
	// Gap #8 fix: Expose mouse delta for camera control and aiming
	MouseDeltaX, MouseDeltaY int

	// Menu navigation input (for UI abstraction)
	MenuUpJustPressed      bool
	MenuDownJustPressed    bool
	MenuConfirmJustPressed bool
	MenuBackJustPressed    bool
	MenuTabJustPressed     bool
}

// Type returns the component type identifier (implements Component).
func (i *EbitenInput) Type() string {
	return "input"
}

// GetMovement implements InputProvider interface.
func (i *EbitenInput) GetMovement() (x, y float64) {
	return i.MoveX, i.MoveY
}

// IsActionPressed implements InputProvider interface.
func (i *EbitenInput) IsActionPressed() bool {
	return i.ActionPressed
}

// IsActionJustPressed implements InputProvider interface.
func (i *EbitenInput) IsActionJustPressed() bool {
	return i.ActionJustPressed
}

// IsAnyKeyPressed implements InputProvider interface.
func (i *EbitenInput) IsAnyKeyPressed() bool {
	return i.AnyKeyPressed
}

// IsUseItemPressed implements InputProvider interface.
func (i *EbitenInput) IsUseItemPressed() bool {
	return i.UseItemPressed
}

// IsUseItemJustPressed implements InputProvider interface.
func (i *EbitenInput) IsUseItemJustPressed() bool {
	return i.UseItemJustPressed
}

// IsSpellPressed implements InputProvider interface.
func (i *EbitenInput) IsSpellPressed(slot int) bool {
	switch slot {
	case 1:
		return i.Spell1Pressed
	case 2:
		return i.Spell2Pressed
	case 3:
		return i.Spell3Pressed
	case 4:
		return i.Spell4Pressed
	case 5:
		return i.Spell5Pressed
	default:
		return false
	}
}

// GetMousePosition implements InputProvider interface.
func (i *EbitenInput) GetMousePosition() (x, y int) {
	return i.MouseX, i.MouseY
}

// IsMousePressed implements InputProvider interface.
func (i *EbitenInput) IsMousePressed() bool {
	return i.MousePressed
}

// SetMovement implements InputProvider interface.
func (i *EbitenInput) SetMovement(x, y float64) {
	i.MoveX, i.MoveY = x, y
}

// SetActionPressed implements InputProvider interface.
func (i *EbitenInput) SetActionPressed(pressed bool) {
	i.ActionPressed = pressed
}

// GetMouseDelta implements InputProvider interface.
// Gap #8 fix: Returns the mouse movement since the last frame.
// Useful for camera controls, aiming, and mouse-based interactions.
func (i *EbitenInput) GetMouseDelta() (dx, dy int) {
	return i.MouseDeltaX, i.MouseDeltaY
}

// IsMenuUpJustPressed implements InputProvider interface.
func (i *EbitenInput) IsMenuUpJustPressed() bool {
	return i.MenuUpJustPressed
}

// IsMenuDownJustPressed implements InputProvider interface.
func (i *EbitenInput) IsMenuDownJustPressed() bool {
	return i.MenuDownJustPressed
}

// IsMenuConfirmJustPressed implements InputProvider interface.
func (i *EbitenInput) IsMenuConfirmJustPressed() bool {
	return i.MenuConfirmJustPressed
}

// IsMenuBackJustPressed implements InputProvider interface.
func (i *EbitenInput) IsMenuBackJustPressed() bool {
	return i.MenuBackJustPressed
}

// IsMenuTabJustPressed implements InputProvider interface.
func (i *EbitenInput) IsMenuTabJustPressed() bool {
	return i.MenuTabJustPressed
}

// GetTouchIDs implements InputProvider interface.
// Returns the list of current active touch IDs.
func (i *EbitenInput) GetTouchIDs() []ebiten.TouchID {
	return ebiten.TouchIDs()
}

// GetTouchPosition implements InputProvider interface.
// Returns the screen position of the touch with the given ID.
func (i *EbitenInput) GetTouchPosition(id ebiten.TouchID) (x, y int) {
	return ebiten.TouchPosition(id)
}

// Compile-time interface check
var _ InputProvider = (*EbitenInput)(nil)

// InputSystem processes keyboard, mouse, and touch input and updates input components.
type InputSystem struct {
	// Movement speed multiplier
	MoveSpeed float64

	// Key bindings - Movement
	KeyUp    ebiten.Key
	KeyDown  ebiten.Key
	KeyLeft  ebiten.Key
	KeyRight ebiten.Key

	// Key bindings - Actions
	KeyAction   ebiten.Key
	KeyUseItem  ebiten.Key
	KeyInteract ebiten.Key // F key for interacting with NPCs/merchants

	// GAP-002 REPAIR: Spell casting key bindings (keys 1-5)
	KeySpell1 ebiten.Key
	KeySpell2 ebiten.Key
	KeySpell3 ebiten.Key
	KeySpell4 ebiten.Key
	KeySpell5 ebiten.Key

	// Key bindings - UI
	KeyInventory ebiten.Key // I key for inventory
	KeyCharacter ebiten.Key // C key for character screen
	KeySkills    ebiten.Key // K key for skills screen
	KeyQuests    ebiten.Key // J key for quest log
	KeyMap       ebiten.Key // M key for map
	KeyCrafting  ebiten.Key // R key for crafting
	KeyMailbox   ebiten.Key // L key for mailbox
	KeyTrade     ebiten.Key // T key for trading (Phase 3.3)
	KeyClasses   ebiten.Key // X key for advanced classes (Phase 4.2) - Changed from A to avoid WASD conflict
	KeyTerritory ebiten.Key // Y key for territory control (Phase 4.3)
	KeyGuild     ebiten.Key // O key for guild management (Phase 3.2) - Changed from U to avoid Achievements conflict
	KeyPrestige  ebiten.Key // P key for prestige and paragon points (Phase 1.3)

	// INTEGRATION FIX [Category B]: V8.0 UI key bindings
	// Gap: Housing and Gallery UIs created but no key binding fields
	// Fix: Added key binding fields for H (housing) and G (gallery)
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4
	KeyHousing ebiten.Key // H key for housing management (Phase 49.1)
	KeyGallery ebiten.Key // G key for image gallery (Phase 49.4)

	// Key bindings - System
	KeyHelp         ebiten.Key // ESC key for help menu
	KeyQuickSave    ebiten.Key // F5 key for quick save
	KeyQuickLoad    ebiten.Key // F9 key for quick load
	KeyFullscreen   ebiten.Key // F11 key for fullscreen toggle
	KeyCycleTargets ebiten.Key // Tab key for cycling targets

	// References to game systems for special key handling
	helpSystem       *EbitenHelpSystem
	tutorialSystem   *EbitenTutorialSystem
	cameraSystem     *CameraSystem     // Phase 10.1: For screen-to-world coordinate conversion
	expressionSystem *ExpressionSystem // Phase 26.1: For expression/emote handling
	mailboxUI        *MailboxUI        // BUG FIX: Reference for ESC key handling

	// BUG FIX: Phase 3 - Menu Trap - UI references for ESC key dual-exit pattern
	// Resolution: ESC now closes open UI panels before opening pause menu
	inventoryUI     *EbitenInventoryUI
	characterUI     *EbitenCharacterUI
	skillsUI        *EbitenSkillsUI
	questUI         *EbitenQuestUI
	mapUI           *EbitenMapUI
	craftingUI      *CraftingUI
	shopUI          *ShopUI
	tradeUI         *TradeUI           // Phase 3.3 (PLAN.md): Trade UI for T key toggle
	advancedClassUI *AdvancedClassUI   // Phase 4.2 (PLAN.md): Advanced class UI for A key toggle
	territoryUI     *TerritoryUI       // Phase 4.3 (PLAN.md): Territory UI for Y key toggle
	dialogUI        *DialogUI          // Phase 6.2 (PLAN.md): Dialog UI for D key toggle
	prestigeUI      PrestigeUIProvider // Phase 1.3 (PLAN.md): Prestige UI for P key toggle
	housingUI       HousingUIProvider  // INTEGRATION FIX [Category B]: V8.0 Housing UI ESC key handling (Phase 49.1)

	// Mobile input support
	touchHandler    *mobile.TouchInputHandler
	virtualControls *mobile.VirtualControlsLayout
	mobileEnabled   bool
	useTouchInput   bool // Auto-detected or manually set

	// Platform parity fix: Gamepad input support for Desktop and WASM
	gamepadHandler  *GamepadInputHandler
	useGamepadInput bool // Auto-detected when gamepad is connected

	// Mouse state tracking for delta calculation (BUG-010 fix)
	lastMouseX, lastMouseY   int
	mouseDeltaX, mouseDeltaY int

	// Callbacks for UI and save/load operations
	onQuickSave        func() error
	onQuickLoad        func() error
	onFullscreenToggle func() error
	onInventoryOpen    func()
	onCharacterOpen    func()
	onSkillsOpen       func()
	onQuestsOpen       func()
	onMapOpen          func()
	onCraftingOpen     func() // Callback for crafting UI toggle
	onMailboxOpen      func() // Callback for mailbox UI toggle (Phase 40.3)
	onTradeOpen        func() // Callback for trade UI toggle (Phase 3.3)
	onClassesOpen      func() // Callback for advanced class UI toggle (Phase 4.2)
	onTerritoryOpen    func() // Callback for territory UI toggle (Phase 4.3)
	onGuildOpen        func() // Callback for guild UI toggle (Phase 3.2)
	onDialogOpen       func() // Callback for dialog UI toggle (Phase 6.2)
	onPrestigeOpen     func() // Callback for prestige UI toggle (Phase 1.3)
	onCycleTargets     func()
	onMenuToggle       func() // Callback for ESC menu toggle
	onInteract         func() // Callback for F key NPC/merchant interaction

	// INTEGRATION FIX [Category B]: V8.0 UI callbacks
	// Gap: Housing and Gallery UIs created but no callback fields for key bindings
	// Fix: Added callback fields for H (housing) and G (gallery) key handling
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4
	onHousingOpen func() // Callback for housing UI toggle (H key) - Phase 49.1
	onGalleryOpen func() // Callback for gallery UI toggle (G key) - Phase 49.4

	// Priority 2.3: Game state for input filtering
	currentState GameState

	// Priority 2.1: Key binding registry for centralized binding management
	keyBindings *KeyBindingRegistry

	// Reusable buffer for pressed keys to reduce per-frame allocations
	pressedKeysBuffer []ebiten.Key
}

// NewInputSystem creates a new input system with default key bindings.
// Platform parity fix: Now initializes gamepad handler for Desktop/WASM support.
// WASM fix: Pre-initializes virtual controls to eliminate first-touch delay.
func NewInputSystem() *InputSystem {
	system := &InputSystem{
		MoveSpeed: 100.0, // pixels per second

		// Movement keys
		KeyUp:    ebiten.KeyW,
		KeyDown:  ebiten.KeyS,
		KeyLeft:  ebiten.KeyA,
		KeyRight: ebiten.KeyD,

		// Action keys
		KeyAction:   ebiten.KeySpace,
		KeyUseItem:  ebiten.KeyE,
		KeyInteract: ebiten.KeyF, // F key for interaction with NPCs/merchants

		// GAP-002 REPAIR: Spell casting keys (1-5)
		KeySpell1: ebiten.Key1,
		KeySpell2: ebiten.Key2,
		KeySpell3: ebiten.Key3,
		KeySpell4: ebiten.Key4,
		KeySpell5: ebiten.Key5,

		// UI keys
		KeyInventory: ebiten.KeyI,
		KeyCharacter: ebiten.KeyC,
		KeySkills:    ebiten.KeyK,
		KeyQuests:    ebiten.KeyJ,
		KeyMap:       ebiten.KeyM,
		KeyCrafting:  ebiten.KeyR,
		KeyMailbox:   ebiten.KeyL, // Phase 40.3: Mailbox UI
		KeyTrade:     ebiten.KeyT, // Phase 3.3: Trade UI
		KeyClasses:   ebiten.KeyX, // Phase 4.2: Advanced Classes UI - Changed from A to avoid WASD conflict
		KeyTerritory: ebiten.KeyY, // Phase 4.3: Territory UI
		KeyGuild:     ebiten.KeyO, // Phase 3.2: Guild UI - Changed from U to avoid Achievements conflict
		KeyPrestige:  ebiten.KeyP, // Phase 1.3: Prestige UI
		KeyHousing:   ebiten.KeyH, // Phase 49.1: Housing UI (V8.0)
		KeyGallery:   ebiten.KeyG, // Phase 49.4: Gallery UI (V8.0)

		// System keys
		KeyHelp:         ebiten.KeyEscape,
		KeyQuickSave:    ebiten.KeyF5,
		KeyQuickLoad:    ebiten.KeyF9,
		KeyFullscreen:   ebiten.KeyF11,
		KeyCycleTargets: ebiten.KeyTab,

		// Mobile input
		touchHandler:  mobile.NewTouchInputHandler(),
		mobileEnabled: mobile.IsMobilePlatform(),
		useTouchInput: mobile.IsTouchCapable(), // Enable touch input for mobile AND WASM

		// Platform parity fix: Gamepad input support for Desktop and WASM
		gamepadHandler:  NewGamepadInputHandler(),
		useGamepadInput: false, // Will be auto-enabled when gamepad is detected

		// Priority 2.3: Initialize game state to exploring
		currentState: StateExploring,

		// Priority 2.1: Initialize key binding registry
		keyBindings: NewKeyBindingRegistry(),

		// Pre-allocate buffer for pressed keys (typical case: 1-4 keys pressed)
		pressedKeysBuffer: make([]ebiten.Key, 0, 10),
	}

	// WASM FIX (Gap #3): Pre-initialize virtual controls to eliminate first-touch delay
	// On WASM, touch-capable devices need controls ready immediately to avoid 1-frame lag
	if mobile.IsWASM() && mobile.IsTouchCapable() {
		// Pre-initialize with default screen size (will resize on first Update)
		system.virtualControls = mobile.NewVirtualControlsLayout(800, 600)
		system.virtualControls.SetVisible(false) // Hidden until first touch
	}

	return system
}

// InitializeVirtualControls sets up virtual controls for mobile platforms and WASM/browser.
// Should be called after screen size is known.
func (s *InputSystem) InitializeVirtualControls(screenWidth, screenHeight int) {
	// Initialize virtual controls for any touch-capable platform (mobile or WASM)
	if s.useTouchInput {
		s.virtualControls = mobile.NewVirtualControlsLayout(screenWidth, screenHeight)
	}
}

// SetMobileEnabled manually enables or disables mobile input support.
func (s *InputSystem) SetMobileEnabled(enabled bool) {
	s.mobileEnabled = enabled
	s.useTouchInput = enabled
}

// IsMobileEnabled returns true if mobile input support is active.
func (s *InputSystem) IsMobileEnabled() bool {
	return s.mobileEnabled
}

// GetGameState returns the current game state for input filtering.
// Priority 2.3: State management for input filtering
func (s *InputSystem) GetGameState() GameState {
	return s.currentState
}

// SetGameState changes the current game state and adjusts input filtering accordingly.
// Priority 2.3: Allows game to switch between states (e.g., opening inventory, entering combat)
func (s *InputSystem) SetGameState(state GameState) {
	s.currentState = state
}

// GetKeyBindings returns the key binding registry for UI label generation.
// Priority 2.1: Allows UI systems to query keybindings for accurate labels
func (s *InputSystem) GetKeyBindings() *KeyBindingRegistry {
	return s.keyBindings
}

// IsGameplayInputAllowed returns true if gameplay input (movement, combat, interaction) is allowed.
// Returns false if any UI is open or a modal dialog is active according to the current GameState.
// Note: Uses AllowsMovement() as a proxy since movement and gameplay share the same allowed states.
func (s *InputSystem) IsGameplayInputAllowed() bool {
	return s.currentState.AllowsMovement()
}

// IsInteractionAllowed returns true if interaction input (F key) should be processed.
// INPUT CONFLICT FIX: Prevents interaction when UIs are open according to the current GameState.
func (s *InputSystem) IsInteractionAllowed() bool {
	return s.currentState.AllowsInteraction()
}

// Update processes input for all entities with input components.
// Platform parity fix: Now processes gamepad input alongside keyboard/mouse/touch.
func (s *InputSystem) Update(entities []*Entity, deltaTime float64) {
	s.ensureTouchInputInitialized()
	s.updateGamepadState()
	s.updateMousePosition()
	s.updateTouchInput()
	s.handleEscapeKey()
	s.handleQuickSaveLoad()
	s.handleUIShortcuts()
	s.handleGamepadUIShortcuts()
	s.handleExpressionHotkeys(entities)
	s.handleInteractionKeys()

	if s.handleHelpTopics() {
		return
	}

	s.processEntityInputs(entities, deltaTime)
}

// ensureTouchInputInitialized validates and initializes touch input controls if needed.
func (s *InputSystem) ensureTouchInputInitialized() {
	if s.useTouchInput && s.virtualControls == nil {
		s.InitializeVirtualControls(800, 600)
	}
}

// updateGamepadState updates gamepad connection and auto-enables gamepad input.
// Platform parity fix: Enables seamless controller support on Desktop and WASM.
func (s *InputSystem) updateGamepadState() {
	if s.gamepadHandler != nil {
		s.gamepadHandler.Update()
		// Auto-enable gamepad input when controller is connected
		s.useGamepadInput = s.gamepadHandler.HasGamepad()
	}
}

// updateMousePosition tracks mouse position for delta calculation.
func (s *InputSystem) updateMousePosition() {
	currentMouseX, currentMouseY := ebiten.CursorPosition()
	s.mouseDeltaX = currentMouseX - s.lastMouseX
	s.mouseDeltaY = currentMouseY - s.lastMouseY
	s.lastMouseX = currentMouseX
	s.lastMouseY = currentMouseY
}

// updateTouchInput processes touch handler and virtual controls updates.
// Platform parity fix: Now handles all extended virtual control buttons.
func (s *InputSystem) updateTouchInput() {
	if !s.useTouchInput || s.touchHandler == nil {
		return
	}

	s.touchHandler.Update()

	if s.virtualControls != nil {
		s.virtualControls.Update()
		s.handleVirtualControlCallbacks()
	}
}

// handleVirtualControlCallbacks processes all virtual control button presses.
func (s *InputSystem) handleVirtualControlCallbacks() {
	s.handleCoreControlCallbacks()
	s.handleExtendedControlCallbacks()
	s.handleUIShortcutCallbacks()
}

// handleCoreControlCallbacks processes menu button presses.
func (s *InputSystem) handleCoreControlCallbacks() {
	if s.virtualControls.IsMenuPressed() && s.onMenuToggle != nil {
		s.onMenuToggle()
	}
}

// handleExtendedControlCallbacks processes inventory, target, and interact buttons.
func (s *InputSystem) handleExtendedControlCallbacks() {
	if s.virtualControls.IsInventoryPressed() && s.onInventoryOpen != nil {
		s.onInventoryOpen()
	}
	if s.virtualControls.IsTargetPressed() && s.onCycleTargets != nil {
		s.onCycleTargets()
	}
	if s.virtualControls.IsInteractPressed() && s.onInteract != nil {
		s.onInteract()
	}
}

// handleUIShortcutCallbacks processes character, skills, quest log, and map buttons.
func (s *InputSystem) handleUIShortcutCallbacks() {
	if s.virtualControls.IsCharacterPressed() && s.onCharacterOpen != nil {
		s.onCharacterOpen()
	}
	if s.virtualControls.IsSkillsPressed() && s.onSkillsOpen != nil {
		s.onSkillsOpen()
	}
	if s.virtualControls.IsQuestLogPressed() && s.onQuestsOpen != nil {
		s.onQuestsOpen()
	}
	if s.virtualControls.IsMapPressed() && s.onMapOpen != nil {
		s.onMapOpen()
	}
}

// handleEscapeKey processes ESC key with priority: tutorial > help > any open UI > pause menu.
// BUG FIX: Phase 3 - Menu Trap - Complete dual-exit pattern for all UI panels
// Resolution: ESC now closes ALL open UI panels (inventory, character, skills, quests, map, crafting, shop, mailbox) before opening pause menu
func (s *InputSystem) handleEscapeKey() {
	if !inpututil.IsKeyJustPressed(s.KeyHelp) {
		return
	}

	if s.handleHighPriorityEscapeActions() {
		return
	}

	if s.handleGameUIEscapeActions() {
		return
	}

	if s.handleShopUIEscapeActions() {
		return
	}

	s.togglePauseMenu()
}

// handleHighPriorityEscapeActions processes tutorial and help menu escape actions.
func (s *InputSystem) handleHighPriorityEscapeActions() bool {
	if s.tutorialSystem != nil && s.tutorialSystem.Enabled && s.tutorialSystem.ShowUI {
		// BUG FIX: ESC should minimize the tutorial overlay, not skip the
		// current step.  Skip() was marking the current step complete which
		// caused unintended step advancement.  HideTutorialUI() hides the
		// panel but keeps progression tracking active so the player can
		// still complete objectives organically.
		s.tutorialSystem.HideTutorialUI()
		return true
	}

	if s.helpSystem != nil && s.helpSystem.Visible {
		s.helpSystem.Toggle()
		return true
	}

	return false
}

// handleGameUIEscapeActions closes open game UI panels (inventory, character, skills, quest, map).
func (s *InputSystem) handleGameUIEscapeActions() bool {
	if s.inventoryUI != nil && s.inventoryUI.IsVisible() {
		s.inventoryUI.Hide()
		return true
	}

	if s.characterUI != nil && s.characterUI.IsVisible() {
		s.characterUI.Hide()
		return true
	}

	if s.skillsUI != nil && s.skillsUI.IsVisible() {
		s.skillsUI.Hide()
		return true
	}

	if s.questUI != nil && s.questUI.IsVisible() {
		s.questUI.Hide()
		return true
	}

	if s.mapUI != nil && s.mapUI.fullScreen {
		s.mapUI.HideFullScreen()
		return true
	}

	return false
}

// handleShopUIEscapeActions closes open shop-related UI panels (crafting, shop, mailbox, trade, classes).
func (s *InputSystem) handleShopUIEscapeActions() bool {
	if s.craftingUI != nil && s.craftingUI.IsVisible() {
		s.craftingUI.Close()
		return true
	}

	if s.shopUI != nil && s.shopUI.IsVisible() {
		s.shopUI.Close()
		return true
	}

	// Phase 3.3 (PLAN.md): Close trade UI on ESC
	if s.tradeUI != nil && s.tradeUI.IsVisible() {
		s.tradeUI.Close()
		return true
	}

	// Phase 4.2 (PLAN.md): Close advanced class UI on ESC
	if s.advancedClassUI != nil && s.advancedClassUI.IsVisible() {
		s.advancedClassUI.SetVisible(false)
		return true
	}

	// Phase 4.3 (PLAN.md): Close territory UI on ESC
	if s.territoryUI != nil && s.territoryUI.IsVisible() {
		s.territoryUI.Hide()
		return true
	}

	// Phase 6.2 (PLAN.md): Close dialog UI on ESC
	if s.dialogUI != nil && s.dialogUI.IsVisible() {
		s.dialogUI.Hide()
		return true
	}

	// Phase 1.3 (PLAN.md): Close prestige UI on ESC
	if s.prestigeUI != nil && s.prestigeUI.IsVisible() {
		s.prestigeUI.Hide()
		return true
	}

	if s.mailboxUI != nil && s.mailboxUI.IsOpen() {
		s.mailboxUI.Close()
		return true
	}

	// INTEGRATION FIX [Category B]: V8.0 Housing UI ESC key handling (Phase 49.1)
	// Gap: ESC key didn't close HousingUI
	// Fix: Added ESC key handling for HousingUI to close panel
	if s.housingUI != nil && s.housingUI.IsVisible() {
		s.housingUI.Hide()
		return true
	}

	return false
}

// togglePauseMenu activates the pause menu if no UI is open.
func (s *InputSystem) togglePauseMenu() {
	if s.onMenuToggle != nil {
		s.onMenuToggle()
	}
}

// handleQuickSaveLoad processes F5 quick save, F9 quick load, and F11 fullscreen toggle.
func (s *InputSystem) handleQuickSaveLoad() {
	if inpututil.IsKeyJustPressed(s.KeyQuickSave) && s.onQuickSave != nil {
		s.executeQuickSave()
	}

	if inpututil.IsKeyJustPressed(s.KeyQuickLoad) && s.onQuickLoad != nil {
		s.executeQuickLoad()
	}

	if inpututil.IsKeyJustPressed(s.KeyFullscreen) && s.onFullscreenToggle != nil {
		s.executeFullscreenToggle()
	}
}

// executeQuickSave performs quick save operation with notification feedback.
func (s *InputSystem) executeQuickSave() {
	if err := s.onQuickSave(); err != nil {
		s.showNotification("Save Failed: " + err.Error())
	} else {
		s.showNotification("Game Saved!")
	}
}

// executeQuickLoad performs quick load operation with notification feedback.
func (s *InputSystem) executeQuickLoad() {
	if err := s.onQuickLoad(); err != nil {
		s.showNotification("Load Failed: " + err.Error())
	} else {
		s.showNotification("Game Loaded!")
	}
}

// executeFullscreenToggle performs fullscreen toggle with notification feedback.
func (s *InputSystem) executeFullscreenToggle() {
	if err := s.onFullscreenToggle(); err != nil {
		s.showNotification("Fullscreen Toggle Failed: " + err.Error())
	}
}

// showNotification displays a notification message through the tutorial system.
func (s *InputSystem) showNotification(message string) {
	if s.tutorialSystem != nil {
		duration := 2.0
		if len(message) > 20 {
			duration = 3.0
		}
		s.tutorialSystem.ShowNotification(message, duration)
	}
}

// handleUIShortcuts processes keyboard shortcuts for opening UI panels.
// INPUT CONFLICT FIX: Added state guard to prevent UI shortcuts during dialogue
func (s *InputSystem) handleUIShortcuts() {
	// Only allow UI shortcuts when UI toggle is permitted by current state
	if !s.currentState.AllowsUIToggle() {
		return
	}
	s.handleCoreUIShortcuts()
	s.handleCommerceUIShortcuts()
	s.handlePhaseUIShortcuts()
}

// handleCoreUIShortcuts processes shortcuts for primary UI screens.
func (s *InputSystem) handleCoreUIShortcuts() {
	if inpututil.IsKeyJustPressed(s.KeyInventory) && s.onInventoryOpen != nil {
		s.onInventoryOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyCharacter) && s.onCharacterOpen != nil {
		s.onCharacterOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeySkills) && s.onSkillsOpen != nil {
		s.onSkillsOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyQuests) && s.onQuestsOpen != nil {
		s.onQuestsOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyMap) && s.onMapOpen != nil {
		s.onMapOpen()
	}
}

// handleCommerceUIShortcuts processes shortcuts for commerce-related UIs.
func (s *InputSystem) handleCommerceUIShortcuts() {
	if inpututil.IsKeyJustPressed(s.KeyCrafting) && s.onCraftingOpen != nil {
		s.onCraftingOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyMailbox) && s.onMailboxOpen != nil {
		s.onMailboxOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyTrade) && s.onTradeOpen != nil {
		s.onTradeOpen()
	}
}

// handlePhaseUIShortcuts processes shortcuts for phase-specific UI screens.
func (s *InputSystem) handlePhaseUIShortcuts() {
	if inpututil.IsKeyJustPressed(s.KeyClasses) && s.onClassesOpen != nil {
		s.onClassesOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyTerritory) && s.onTerritoryOpen != nil {
		s.onTerritoryOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyGuild) && s.onGuildOpen != nil {
		s.onGuildOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyPrestige) && s.onPrestigeOpen != nil {
		s.onPrestigeOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyHousing) && s.onHousingOpen != nil {
		s.onHousingOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyGallery) && s.onGalleryOpen != nil {
		s.onGalleryOpen()
	}
}

// handleGamepadUIShortcuts processes gamepad button presses for UI actions.
// Platform parity fix: Enables full UI navigation via gamepad controller.
func (s *InputSystem) handleGamepadUIShortcuts() {
	if !s.useGamepadInput || s.gamepadHandler == nil {
		return
	}

	// Menu button (Start)
	if s.gamepadHandler.IsMenuJustPressed() && s.onMenuToggle != nil {
		s.onMenuToggle()
	}

	// Inventory button (RB)
	if s.gamepadHandler.IsInventoryJustPressed() && s.onInventoryOpen != nil {
		s.onInventoryOpen()
	}

	// Map button (Back/Select)
	if s.gamepadHandler.IsMapJustPressed() && s.onMapOpen != nil {
		s.onMapOpen()
	}

	// Cycle targets (LB)
	if s.gamepadHandler.IsCycleTargetJustPressed() && s.onCycleTargets != nil {
		s.onCycleTargets()
	}

	// Interact (Y button)
	if s.gamepadHandler.IsInteractJustPressed() && s.onInteract != nil {
		s.onInteract()
	}

	// Platform parity fix: UI shortcut buttons for complete gamepad accessibility
	// Character (L3 - left stick click)
	if s.gamepadHandler.IsCharacterJustPressed() && s.onCharacterOpen != nil {
		s.onCharacterOpen()
	}

	// Skills (R3 - right stick click)
	if s.gamepadHandler.IsSkillsJustPressed() && s.onSkillsOpen != nil {
		s.onSkillsOpen()
	}
}

// handleExpressionHotkeys processes Shift+number key combinations for player expressions.
func (s *InputSystem) handleExpressionHotkeys(entities []*Entity) {
	if s.expressionSystem == nil || !ebiten.IsKeyPressed(ebiten.KeyShift) {
		return
	}

	playerEntity := s.findPlayerEntity(entities)
	if playerEntity == nil {
		return
	}

	s.processExpressionKeys(playerEntity)
}

// findPlayerEntity locates the player-controlled entity from the entity list.
func (s *InputSystem) findPlayerEntity(entities []*Entity) *Entity {
	for _, entity := range entities {
		if inputComp, ok := entity.GetComponent("input"); ok && inputComp != nil {
			return entity
		}
	}
	return nil
}

// processExpressionKeys checks expression key bindings and triggers expressions.
func (s *InputSystem) processExpressionKeys(playerEntity *Entity) {
	expressionKeys := []struct {
		key     ebiten.Key
		expType ExpressionType
	}{
		{ebiten.Key1, ExpressionWave},
		{ebiten.Key2, ExpressionCheer},
		{ebiten.Key3, ExpressionDance},
		{ebiten.Key4, ExpressionLaugh},
		{ebiten.Key5, ExpressionCry},
		{ebiten.Key6, ExpressionSit},
		{ebiten.Key7, ExpressionPoint},
		{ebiten.Key8, ExpressionSalute},
		{ebiten.Key9, ExpressionShrug},
		{ebiten.Key0, ExpressionThumbsUp},
		{ebiten.KeyMinus, ExpressionFacepalm},
		{ebiten.KeyEqual, ExpressionSleep},
	}

	for _, binding := range expressionKeys {
		if inpututil.IsKeyJustPressed(binding.key) {
			s.expressionSystem.TriggerExpression(playerEntity.ID, binding.expType)
			break
		}
	}
}

// handleInteractionKeys processes NPC interaction and target cycling keys.
func (s *InputSystem) handleInteractionKeys() {
	if inpututil.IsKeyJustPressed(s.KeyInteract) && s.onInteract != nil {
		s.onInteract()
	}

	if inpututil.IsKeyJustPressed(s.KeyCycleTargets) && s.onCycleTargets != nil {
		s.onCycleTargets()
	}
}

// handleHelpTopics processes help topic switching and returns true if handled.
func (s *InputSystem) handleHelpTopics() bool {
	if s.helpSystem == nil || !s.helpSystem.Visible {
		return false
	}

	topicKeys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3,
		ebiten.Key4, ebiten.Key5, ebiten.Key6,
	}
	topicIDs := []string{
		"controls", "combat", "inventory",
		"progression", "world", "multiplayer",
	}

	for i, key := range topicKeys {
		if inpututil.IsKeyJustPressed(key) {
			s.helpSystem.ShowTopic(topicIDs[i])
			return true
		}
	}

	return false
}

// processEntityInputs iterates through entities and processes their input components.
func (s *InputSystem) processEntityInputs(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		inputComp, ok := entity.GetComponent("input")
		if !ok {
			continue
		}

		input, ok := inputComp.(*EbitenInput)
		if !ok {
			continue
		}
		s.processInput(entity, input, deltaTime)
	}
}

// processInput handles input processing for a single entity.
func (s *InputSystem) processInput(entity *Entity, input *EbitenInput, deltaTime float64) {
	s.resetInputFlags(input)
	s.detectInputMethod()
	s.updateInputFromSource(input)
	s.updateEntityAim(entity, input)
	s.applyInputToVelocity(entity, input)
}

// resetInputFlags clears all input flags at the start of each frame.
func (s *InputSystem) resetInputFlags(input *EbitenInput) {
	input.MoveX = 0
	input.MoveY = 0
	input.ActionPressed = false
	input.UseItemPressed = false
	input.Spell1Pressed = false
	input.Spell2Pressed = false
	input.Spell3Pressed = false
	input.Spell4Pressed = false
	input.Spell5Pressed = false
	input.ActionJustPressed = false
	input.UseItemJustPressed = false
	input.AnyKeyPressed = false
	// Menu navigation flags
	input.MenuUpJustPressed = false
	input.MenuDownJustPressed = false
	input.MenuConfirmJustPressed = false
	input.MenuBackJustPressed = false
	input.MenuTabJustPressed = false
}

// detectInputMethod auto-detects whether to use touch or keyboard/mouse input.
// WASM fix: Shows pre-initialized virtual controls on first touch.
func (s *InputSystem) detectInputMethod() {
	if len(ebiten.TouchIDs()) > 0 {
		s.useTouchInput = true

		// WASM FIX (Gap #3): Show pre-initialized controls on first touch
		if s.virtualControls != nil && !s.virtualControls.IsVisible() {
			s.virtualControls.SetVisible(true)
		}

		// Fallback: Initialize if controls don't exist (native mobile platforms)
		if s.virtualControls == nil && mobile.IsTouchCapable() {
			screenW, screenH := ebiten.WindowSize()
			s.InitializeVirtualControls(screenW, screenH)
		}
	} else if !s.mobileEnabled && len(ebiten.TouchIDs()) == 0 {
		s.useTouchInput = false
	}
}

// updateInputFromSource processes input from touch, gamepad, or keyboard/mouse.
// Platform parity fix: Unified input processing from all available sources.
func (s *InputSystem) updateInputFromSource(input *EbitenInput) {
	// Priority: Touch > Gamepad > Keyboard/Mouse
	if s.useTouchInput && s.virtualControls != nil {
		s.processTouchInput(input)
	} else if s.useGamepadInput && s.gamepadHandler != nil {
		s.processGamepadInput(input)
	} else {
		s.processKeyboardInput(input)
	}

	// Platform parity fix: Allow gamepad to supplement keyboard on Desktop
	// This enables simultaneous keyboard + gamepad usage
	if !s.useTouchInput && s.useGamepadInput && s.gamepadHandler != nil {
		s.mergeGamepadInput(input)
	}
}

// processTouchInput handles touch input from virtual controls.
// Platform parity fix: Now processes all extended virtual control buttons.
// Note: VirtualButton.IsPressed() returns true ONLY on the frame when touch is released,
// so ActionJustPressed is correctly set for a single frame per button press.
func (s *InputSystem) processTouchInput(input *EbitenInput) {
	moveX, moveY := s.virtualControls.GetMovementInput()
	input.MoveX = moveX
	input.MoveY = moveY

	if s.virtualControls.IsActionPressed() {
		input.ActionPressed = true
		input.ActionJustPressed = true
		input.AnyKeyPressed = true
	}
	if s.virtualControls.IsSecondaryPressed() {
		input.UseItemPressed = true
		input.UseItemJustPressed = true
		input.AnyKeyPressed = true
	}

	// Platform parity fix: Process spell buttons from virtual controls
	for i := 1; i <= 5; i++ {
		if s.virtualControls.IsSpellPressed(i) {
			s.setSpellPressed(input, i)
			input.AnyKeyPressed = true
		}
	}

	s.processTouchMousePosition(input)
}

// processGamepadInput handles gamepad input for movement and actions.
// Platform parity fix: Full gamepad support for Desktop and WASM.
func (s *InputSystem) processGamepadInput(input *EbitenInput) {
	// Always process menu navigation regardless of movement state
	// Housing UI Low-Priority Fix: D-pad menu navigation for all menu contexts
	s.processGamepadMenuNavigation(input)

	if !s.currentState.AllowsMovement() {
		// Still process non-movement input
		s.processGamepadActionKeys(input)
		return
	}

	// Movement from left stick
	moveX, moveY := s.gamepadHandler.GetMovementInput()
	input.MoveX = moveX
	input.MoveY = moveY

	// Process action buttons
	s.processGamepadActionKeys(input)

	// Process aim from right stick for mouse position
	aimX, aimY := s.gamepadHandler.GetAimInput()
	if aimX != 0 || aimY != 0 {
		// Convert aim direction to screen position relative to center
		// AimSensitivity controls how far from center the cursor moves
		const aimSensitivity = 200.0 // pixels at max stick deflection
		screenW, screenH := ebiten.WindowSize()
		input.MouseX = screenW/2 + int(aimX*aimSensitivity)
		input.MouseY = screenH/2 + int(aimY*aimSensitivity)
	} else {
		// When gamepad aim is neutral, keep cursor in sync with the actual mouse
		// position for consistent behavior across input modes.
		s.processMouseState(input)
	}
}

// processGamepadActionKeys handles gamepad button presses for combat actions.
// Platform parity fix: Maps gamepad buttons to action inputs.
func (s *InputSystem) processGamepadActionKeys(input *EbitenInput) {
	if !s.currentState.AllowsCombat() {
		return
	}

	if s.gamepadHandler.IsAttackJustPressed() {
		input.ActionPressed = true
		input.ActionJustPressed = true
		input.AnyKeyPressed = true
	}

	if s.gamepadHandler.IsUseItemJustPressed() {
		input.UseItemPressed = true
		input.UseItemJustPressed = true
		input.AnyKeyPressed = true
	}

	// D-pad spell casting (slots 1-4)
	for slot := 1; slot <= 4; slot++ {
		if s.gamepadHandler.IsSpellJustPressed(slot) {
			s.setSpellPressed(input, slot)
			input.AnyKeyPressed = true
		}
	}

	// Right trigger for spell 5 (alternative)
	if s.gamepadHandler.IsRightTriggerPressed() {
		input.Spell5Pressed = true
		input.AnyKeyPressed = true
	}
}

// mergeGamepadInput merges gamepad input with existing keyboard input.
// Platform parity fix: Allows simultaneous keyboard + gamepad on Desktop.
func (s *InputSystem) mergeGamepadInput(input *EbitenInput) {
	// Always merge menu navigation regardless of combat state
	// Housing UI Low-Priority Fix: D-pad menu navigation when using keyboard + gamepad
	s.processGamepadMenuNavigation(input)

	// Only merge if no keyboard movement detected
	if input.MoveX == 0 && input.MoveY == 0 {
		moveX, moveY := s.gamepadHandler.GetMovementInput()
		input.MoveX = moveX
		input.MoveY = moveY
		if moveX != 0 || moveY != 0 {
			input.AnyKeyPressed = true
		}
	}

	// Only merge combat inputs if combat is allowed (consistent with processGamepadActionKeys)
	if !s.currentState.AllowsCombat() {
		return
	}

	// Merge action buttons (OR logic)
	if s.gamepadHandler.IsAttackJustPressed() {
		input.ActionPressed = true
		input.ActionJustPressed = true
		input.AnyKeyPressed = true
	}
	if s.gamepadHandler.IsUseItemJustPressed() {
		input.UseItemPressed = true
		input.UseItemJustPressed = true
		input.AnyKeyPressed = true
	}

	// Merge spell casting inputs (D-pad slots 1-4)
	for slot := 1; slot <= 4; slot++ {
		if s.gamepadHandler.IsSpellJustPressed(slot) {
			s.setSpellPressed(input, slot)
			input.AnyKeyPressed = true
		}
	}

	// Merge right trigger as alternative spell 5
	if s.gamepadHandler.IsRightTriggerPressed() {
		input.Spell5Pressed = true
		input.AnyKeyPressed = true
	}
}

// setSpellPressed sets the spell pressed flag for the given slot.
func (s *InputSystem) setSpellPressed(input *EbitenInput, slot int) {
	switch slot {
	case 1:
		input.Spell1Pressed = true
	case 2:
		input.Spell2Pressed = true
	case 3:
		input.Spell3Pressed = true
	case 4:
		input.Spell4Pressed = true
	case 5:
		input.Spell5Pressed = true
	}
}

// processTouchMousePosition maps touch position outside controls to mouse position.
func (s *InputSystem) processTouchMousePosition(input *EbitenInput) {
	if s.touchHandler == nil {
		return
	}

	touches := s.touchHandler.GetActiveTouches()
	screenW, _ := ebiten.WindowSize()

	for _, touch := range touches {
		if touch.X > 200 && touch.X < screenW-200 {
			input.MouseX = touch.X
			input.MouseY = touch.Y
			input.MousePressed = true
			break
		}
	}
}

// processKeyboardInput handles keyboard and mouse input.
func (s *InputSystem) processKeyboardInput(input *EbitenInput) {
	s.processMovementKeys(input)
	s.processActionKeys(input)
	s.processMenuNavigationKeys(input)
	s.detectAnyKeyPress(input)
	s.processMouseState(input)
}

// processMovementKeys handles WASD movement with diagonal normalization.
func (s *InputSystem) processMovementKeys(input *EbitenInput) {
	if !s.currentState.AllowsMovement() {
		return
	}

	if ebiten.IsKeyPressed(s.KeyUp) {
		input.MoveY = -1.0
	}
	if ebiten.IsKeyPressed(s.KeyDown) {
		input.MoveY = 1.0
	}
	if ebiten.IsKeyPressed(s.KeyLeft) {
		input.MoveX = -1.0
	}
	if ebiten.IsKeyPressed(s.KeyRight) {
		input.MoveX = 1.0
	}

	if input.MoveX != 0 && input.MoveY != 0 {
		input.MoveX *= 0.707
		input.MoveY *= 0.707
	}
}

// processActionKeys handles combat action and spell casting keys.
func (s *InputSystem) processActionKeys(input *EbitenInput) {
	if !s.currentState.AllowsCombat() {
		return
	}

	if inpututil.IsKeyJustPressed(s.KeyAction) {
		input.ActionPressed = true
		input.ActionJustPressed = true
		input.AnyKeyPressed = true
	}
	if inpututil.IsKeyJustPressed(s.KeyUseItem) {
		input.UseItemPressed = true
		input.UseItemJustPressed = true
		input.AnyKeyPressed = true
	}

	s.processSpellKeys(input)
}

// processSpellKeys handles spell casting keys 1-5.
func (s *InputSystem) processSpellKeys(input *EbitenInput) {
	spellKeys := []struct {
		key     ebiten.Key
		pressed *bool
	}{
		{s.KeySpell1, &input.Spell1Pressed},
		{s.KeySpell2, &input.Spell2Pressed},
		{s.KeySpell3, &input.Spell3Pressed},
		{s.KeySpell4, &input.Spell4Pressed},
		{s.KeySpell5, &input.Spell5Pressed},
	}

	for _, spell := range spellKeys {
		if inpututil.IsKeyJustPressed(spell.key) {
			*spell.pressed = true
			input.AnyKeyPressed = true
		}
	}
}

// processMenuNavigationKeys handles menu navigation keys for UI abstraction.
// This allows UI components to use InputProvider instead of direct Ebiten calls.
func (s *InputSystem) processMenuNavigationKeys(input *EbitenInput) {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		input.MenuUpJustPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		input.MenuDownJustPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		input.MenuConfirmJustPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		input.MenuBackJustPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		input.MenuTabJustPressed = true
	}
}

// processGamepadMenuNavigation handles gamepad D-pad for menu navigation.
// Housing UI Low-Priority Fix: Wires gamepad D-pad to InputProvider menu navigation methods.
// D-pad Up/Down navigate menus, A confirms, B goes back, LB/RB switch tabs.
func (s *InputSystem) processGamepadMenuNavigation(input *EbitenInput) {
	if s.gamepadHandler == nil || !s.gamepadHandler.HasGamepad() {
		return
	}

	// D-pad for menu navigation
	if s.gamepadHandler.IsDPadUpJustPressed() {
		input.MenuUpJustPressed = true
		input.AnyKeyPressed = true
	}
	if s.gamepadHandler.IsDPadDownJustPressed() {
		input.MenuDownJustPressed = true
		input.AnyKeyPressed = true
	}

	// A button for confirm (same as attack button)
	if s.gamepadHandler.IsConfirmJustPressed() {
		input.MenuConfirmJustPressed = true
		input.AnyKeyPressed = true
	}

	// B button for back/cancel
	if s.gamepadHandler.IsCancelJustPressed() {
		input.MenuBackJustPressed = true
		input.AnyKeyPressed = true
	}

	// Shoulder buttons (LB/RB) for tab switching
	if s.gamepadHandler.IsButtonJustPressed(ebiten.StandardGamepadButtonFrontTopLeft) {
		input.MenuTabJustPressed = true
		input.AnyKeyPressed = true
	}
	if s.gamepadHandler.IsButtonJustPressed(ebiten.StandardGamepadButtonFrontTopRight) {
		input.MenuTabJustPressed = true
		input.AnyKeyPressed = true
	}
}

// detectAnyKeyPress sets AnyKeyPressed flag if any key is pressed.
// Uses reusable buffer to reduce per-frame allocations.
func (s *InputSystem) detectAnyKeyPress(input *EbitenInput) {
	if !input.AnyKeyPressed {
		// Reuse buffer to reduce allocations in hot path
		s.pressedKeysBuffer = inpututil.AppendPressedKeys(s.pressedKeysBuffer[:0])
		if len(s.pressedKeysBuffer) > 0 {
			input.AnyKeyPressed = true
		}
	}
}

// processMouseState updates mouse position and button state.
func (s *InputSystem) processMouseState(input *EbitenInput) {
	input.MouseX, input.MouseY = ebiten.CursorPosition()
	input.MousePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Gap #8 fix: Populate mouse delta from tracked values
	input.MouseDeltaX = s.mouseDeltaX
	input.MouseDeltaY = s.mouseDeltaY
}

// updateEntityAim updates the aim component with mouse position in world coordinates.
func (s *InputSystem) updateEntityAim(entity *Entity, input *EbitenInput) {
	if !entity.HasComponent("aim") || s.cameraSystem == nil {
		return
	}

	aimComp, ok := entity.GetComponent("aim")
	if !ok {
		return
	}

	aim, ok := aimComp.(*AimComponent)
	if !ok {
		return
	}

	worldX, worldY := s.cameraSystem.ScreenToWorld(float64(input.MouseX), float64(input.MouseY))
	aim.SetAimTarget(worldX, worldY)
}

// applyInputToVelocity converts movement input to velocity and updates animation.
func (s *InputSystem) applyInputToVelocity(entity *Entity, input *EbitenInput) {
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		return
	}

	velocity, ok := velComp.(*VelocityComponent)
	if !ok {
		return
	}

	velocity.VX = input.MoveX * s.MoveSpeed
	velocity.VY = input.MoveY * s.MoveSpeed

	s.updateMovementAnimation(entity, velocity)
}

// updateMovementAnimation switches between idle and walk animations based on movement.
func (s *InputSystem) updateMovementAnimation(entity *Entity, velocity *VelocityComponent) {
	animComp, ok := entity.GetComponent("animation")
	if !ok {
		return
	}

	anim, ok := animComp.(*AnimationComponent)
	if !ok {
		return
	}

	isMoving := velocity.VX != 0 || velocity.VY != 0

	if isMoving && anim.CurrentState == AnimationStateIdle {
		anim.SetState(AnimationStateWalk)
	} else if !isMoving && anim.CurrentState == AnimationStateWalk {
		anim.SetState(AnimationStateIdle)
	}
}

// SetKeyBindings allows customizing key bindings.
func (s *InputSystem) SetKeyBindings(up, down, left, right, action, useItem ebiten.Key) {
	s.KeyUp = up
	s.KeyDown = down
	s.KeyLeft = left
	s.KeyRight = right
	s.KeyAction = action
	s.KeyUseItem = useItem
}

// SetHelpSystem connects the help system for ESC key toggling.
func (s *InputSystem) SetHelpSystem(helpSystem *EbitenHelpSystem) {
	s.helpSystem = helpSystem
}

// SetTutorialSystem connects the tutorial system for ESC key handling.
func (s *InputSystem) SetTutorialSystem(tutorialSystem *EbitenTutorialSystem) {
	s.tutorialSystem = tutorialSystem
}

// SetCameraSystem connects the camera system for screen-to-world coordinate conversion.
// Phase 10.1: Required for mouse aim to convert cursor position to world coordinates.
func (s *InputSystem) SetCameraSystem(cameraSystem *CameraSystem) {
	s.cameraSystem = cameraSystem
}

// SetExpressionSystem connects the expression system for expression/emote handling.
// Phase 26.1: Required for hotkey-triggered expressions (Shift+1 through Shift+=).
func (s *InputSystem) SetExpressionSystem(expressionSystem *ExpressionSystem) {
	s.expressionSystem = expressionSystem
}

// SetMailboxUI connects the mailbox UI for ESC key handling.
// BUG FIX: Menu Trap - Allows ESC key to close mailbox UI (dual-exit pattern).
// Resolution: Mailbox UI can now be closed with ESC key in addition to L key toggle.
func (s *InputSystem) SetMailboxUI(mailboxUI *MailboxUI) {
	s.mailboxUI = mailboxUI
}

// BUG FIX: Phase 3 - Menu Trap - UI reference setters for ESC key handling
// Resolution: Added setters so game can connect UI panels to InputSystem for dual-exit pattern

// SetInventoryUI connects the inventory UI for ESC key closing.
func (s *InputSystem) SetInventoryUI(inventoryUI *EbitenInventoryUI) {
	s.inventoryUI = inventoryUI
}

// SetCharacterUI connects the character UI for ESC key closing.
func (s *InputSystem) SetCharacterUI(characterUI *EbitenCharacterUI) {
	s.characterUI = characterUI
}

// SetSkillsUI connects the skills UI for ESC key closing.
func (s *InputSystem) SetSkillsUI(skillsUI *EbitenSkillsUI) {
	s.skillsUI = skillsUI
}

// SetQuestUI connects the quest UI for ESC key closing.
func (s *InputSystem) SetQuestUI(questUI *EbitenQuestUI) {
	s.questUI = questUI
}

// SetMapUI connects the map UI for ESC key closing.
func (s *InputSystem) SetMapUI(mapUI *EbitenMapUI) {
	s.mapUI = mapUI
}

// SetCraftingUI connects the crafting UI for ESC key closing.
func (s *InputSystem) SetCraftingUI(craftingUI *CraftingUI) {
	s.craftingUI = craftingUI
}

// SetShopUI connects the shop UI for ESC key closing.
func (s *InputSystem) SetShopUI(shopUI *ShopUI) {
	s.shopUI = shopUI
}

// SetTradeUI connects the trade UI for T key toggle and ESC key closing.
// Phase 3.3 (PLAN.md): Player-to-player trading UI integration.
func (s *InputSystem) SetTradeUI(tradeUI *TradeUI) {
	s.tradeUI = tradeUI
}

// SetAdvancedClassUI sets the advanced class UI reference for ESC key handling (Phase 4.2)
func (s *InputSystem) SetAdvancedClassUI(advancedClassUI *AdvancedClassUI) {
	s.advancedClassUI = advancedClassUI
}

// SetTerritoryUI sets the territory UI reference for Y key toggle and ESC key handling (Phase 4.3)
func (s *InputSystem) SetTerritoryUI(territoryUI *TerritoryUI) {
	s.territoryUI = territoryUI
}

// SetDialogUI sets the dialog UI reference for D key toggle and ESC key handling (Phase 6.2)
func (s *InputSystem) SetDialogUI(dialogUI *DialogUI) {
	s.dialogUI = dialogUI
}

// SetPrestigeUI sets the prestige UI reference for P key toggle and ESC key handling (Phase 1.3)
func (s *InputSystem) SetPrestigeUI(prestigeUI PrestigeUIProvider) {
	s.prestigeUI = prestigeUI
}

// SetHousingUI sets the housing UI reference for H key toggle and ESC key handling
// INTEGRATION FIX [Category B]: V8.0 Housing UI ESC key handling (Phase 49.1)
// Gap: InputSystem had no way to close HousingUI with ESC key
// Fix: Added SetHousingUI to enable ESC key closing of housing panel
func (s *InputSystem) SetHousingUI(housingUI HousingUIProvider) {
	s.housingUI = housingUI
}

// SetQuickSaveCallback sets the callback function for quick save (F5).
func (s *InputSystem) SetQuickSaveCallback(callback func() error) {
	s.onQuickSave = callback
}

// SetQuickLoadCallback sets the callback function for quick load (F9).
func (s *InputSystem) SetQuickLoadCallback(callback func() error) {
	s.onQuickLoad = callback
}

// SetFullscreenToggleCallback sets the callback function for fullscreen toggle (F11).
func (s *InputSystem) SetFullscreenToggleCallback(callback func() error) {
	s.onFullscreenToggle = callback
}

// SetInventoryCallback sets the callback function for opening inventory (I key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetInventoryCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("inventory callback cannot be nil")
	}
	s.onInventoryOpen = callback
	return nil
}

// SetCharacterCallback sets the callback function for opening character screen (C key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetCharacterCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("character callback cannot be nil")
	}
	s.onCharacterOpen = callback
	return nil
}

// SetSkillsCallback sets the callback function for opening skills screen (K key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetSkillsCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("skills callback cannot be nil")
	}
	s.onSkillsOpen = callback
	return nil
}

// SetQuestsCallback sets the callback function for opening quest log (J key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetQuestsCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("quests callback cannot be nil")
	}
	s.onQuestsOpen = callback
	return nil
}

// SetMapCallback sets the callback function for opening map (M key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetMapCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("map callback cannot be nil")
	}
	s.onMapOpen = callback
	return nil
}

// SetCraftingCallback sets the callback function for opening crafting UI (R key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetCraftingCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("crafting callback cannot be nil")
	}
	s.onCraftingOpen = callback
	return nil
}

// SetMailboxCallback sets the callback function for opening mailbox UI (L key).
func (s *InputSystem) SetMailboxCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("mailbox callback cannot be nil")
	}
	s.onMailboxOpen = callback
	return nil
}

// SetTradeCallback sets the callback function for opening trade UI (T key).
// Phase 3.3 (PLAN.md): Player-to-player trading UI integration.
func (s *InputSystem) SetTradeCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("trade callback cannot be nil")
	}
	s.onTradeOpen = callback
	return nil
}

// SetClassesCallback sets the callback function for opening advanced class UI (X key).
// Phase 4.2 (PLAN.md): Advanced classes (multi-classing, prestige, talents) UI integration.
// INPUT CONFLICT FIX: Changed from A key to X key to avoid WASD movement conflict.
func (s *InputSystem) SetClassesCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("classes callback cannot be nil")
	}
	s.onClassesOpen = callback
	return nil
}

// SetTerritoryCallback sets the callback function for opening territory UI (Y key).
// Phase 4.3 (PLAN.md): Territory control (guild warfare, territory capture) UI integration.
func (s *InputSystem) SetTerritoryCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("territory callback cannot be nil")
	}
	s.onTerritoryOpen = callback
	return nil
}

// SetGuildCallback sets the callback function for opening guild UI (U key).
// Phase 3.2 (PLAN.md): Guild federation UI integration for guild management.
func (s *InputSystem) SetGuildCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("guild callback cannot be nil")
	}
	s.onGuildOpen = callback
	return nil
}

// SetPrestigeCallback sets the callback function for opening prestige UI (P key).
// Phase 1.3 (PLAN.md): Prestige system UI integration for paragon point allocation.
func (s *InputSystem) SetPrestigeCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("prestige callback cannot be nil")
	}
	s.onPrestigeOpen = callback
	return nil
}

// SetDialogCallback sets the callback function for opening dialog UI (D key).
// Phase 6.2 (PLAN.md): Dialog system UI integration for NPC conversations.
func (s *InputSystem) SetDialogCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("dialog callback cannot be nil")
	}
	s.onDialogOpen = callback
	return nil
}

// INTEGRATION FIX [Category B]: V8.0 UI callback setters
// Gap: Housing and Gallery UIs need callback setter methods for key binding integration
// Fix: Added SetHousingCallback and SetGalleryCallback methods
// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4

// SetHousingCallback sets the callback function for opening housing UI (H key).
func (s *InputSystem) SetHousingCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("housing callback cannot be nil")
	}
	s.onHousingOpen = callback
	return nil
}

// SetGalleryCallback sets the callback function for opening gallery UI (G key).
func (s *InputSystem) SetGalleryCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("gallery callback cannot be nil")
	}
	s.onGalleryOpen = callback
	return nil
}

// SetCycleTargetsCallback sets the callback function for cycling targets (Tab key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetCycleTargetsCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("cycle targets callback cannot be nil")
	}
	s.onCycleTargets = callback
	return nil
}

// SetMenuToggleCallback sets the callback function for toggling the pause menu (ESC key).
// This is called when ESC is pressed and neither tutorial nor help system consume the event.
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetMenuToggleCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("menu toggle callback cannot be nil")
	}
	s.onMenuToggle = callback
	return nil
}

// SetInteractCallback sets the callback function for interacting with NPCs/merchants (F key).
// H-008 FIX: Returns error if callback is nil to catch initialization order bugs.
func (s *InputSystem) SetInteractCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("interact callback cannot be nil")
	}
	s.onInteract = callback
	return nil
}

// ===== KEYBOARD INPUT METHODS =====

// IsKeyPressed returns true if the specified key is currently held down.
// Wraps Ebiten's continuous state check for consistency.
func (s *InputSystem) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

// IsKeyJustPressed returns true only on the frame when the key was first pressed.
// This is edge-triggered: returns true once per key press, not continuously while held.
// Explicitly exposed for use in game code.
func (s *InputSystem) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

// IsKeyReleased returns true only on the frame when the key was released.
// BUG-001 fix: Missing method for detecting key release events.
// Essential for charge attacks, aim mechanics, and jump height control.
func (s *InputSystem) IsKeyReleased(key ebiten.Key) bool {
	return inpututil.IsKeyJustReleased(key)
}

// IsKeyJustReleased is an alias for IsKeyReleased for API consistency.
// BUG-002 fix: Matches naming convention of IsKeyJustPressed.
func (s *InputSystem) IsKeyJustReleased(key ebiten.Key) bool {
	return s.IsKeyReleased(key)
}

// GetPressedKeys returns a slice of all keys currently pressed.
// BUG-003 fix: Needed for key binding UI and "press any key" prompts.
// Note: Returns a copy to allow callers to safely store the result.
func (s *InputSystem) GetPressedKeys() []ebiten.Key {
	// Reuse buffer to reduce allocations, then copy for safe return
	s.pressedKeysBuffer = inpututil.AppendPressedKeys(s.pressedKeysBuffer[:0])
	if len(s.pressedKeysBuffer) == 0 {
		return []ebiten.Key{} // Return empty slice, not nil (test expects non-nil)
	}
	result := make([]ebiten.Key, len(s.pressedKeysBuffer))
	copy(result, s.pressedKeysBuffer)
	return result
}

// IsAnyKeyPressed returns true if any keyboard key is currently pressed.
// BUG-021 fix: Common pattern for "press any key to continue" scenarios.
// Uses reusable buffer to reduce allocations.
func (s *InputSystem) IsAnyKeyPressed() bool {
	s.pressedKeysBuffer = inpututil.AppendPressedKeys(s.pressedKeysBuffer[:0])
	return len(s.pressedKeysBuffer) > 0
}

// GetAnyPressedKey returns the first pressed key found, or (0, false) if none.
// BUG-021 fix: Useful for key binding configuration UI.
// Uses reusable buffer to reduce allocations.
func (s *InputSystem) GetAnyPressedKey() (ebiten.Key, bool) {
	s.pressedKeysBuffer = inpututil.AppendPressedKeys(s.pressedKeysBuffer[:0])
	if len(s.pressedKeysBuffer) > 0 {
		return s.pressedKeysBuffer[0], true
	}
	return 0, false
}

// IsShiftPressed returns true if either left or right Shift key is pressed.
// BUG-022 fix: Convenience method for modifier keys.
func (s *InputSystem) IsShiftPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)
}

// IsControlPressed returns true if either left or right Control key is pressed.
// BUG-022 fix: Convenience method for modifier keys.
func (s *InputSystem) IsControlPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
}

// IsAltPressed returns true if either left or right Alt key is pressed.
// BUG-022 fix: Convenience method for modifier keys.
func (s *InputSystem) IsAltPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight)
}

// IsSuperPressed returns true if either left or right Super/Meta key is pressed.
// BUG-022 fix: Super key is Windows/Command/Meta depending on OS.
func (s *InputSystem) IsSuperPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyMetaLeft) || ebiten.IsKeyPressed(ebiten.KeyMetaRight)
}

// ===== MOUSE INPUT METHODS =====

// IsMouseButtonPressed returns true if the mouse button is currently held down.
// This is continuous state - returns true every frame while held.
func (s *InputSystem) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return ebiten.IsMouseButtonPressed(button)
}

// IsMouseButtonJustPressed returns true only on the frame when the button was first pressed.
// BUG-004 fix: Edge-triggered click detection for single-action clicks.
// Critical for UI interactions that should fire once per click, not continuously.
func (s *InputSystem) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

// IsMouseButtonReleased returns true only on the frame when the button was released.
// BUG-005 fix: Essential for drag-and-drop and context menus.
func (s *InputSystem) IsMouseButtonReleased(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustReleased(button)
}

// IsMouseButtonJustReleased is an alias for IsMouseButtonReleased for consistency.
// BUG-006 fix: Matches naming convention of IsMouseButtonJustPressed.
func (s *InputSystem) IsMouseButtonJustReleased(button ebiten.MouseButton) bool {
	return s.IsMouseButtonReleased(button)
}

// GetMousePosition returns the current mouse cursor position.
// Alias for GetCursorPosition for API clarity.
func (s *InputSystem) GetMousePosition() (x, y int) {
	return ebiten.CursorPosition()
}

// GetCursorPosition returns the current mouse cursor position.
func (s *InputSystem) GetCursorPosition() (x, y int) {
	return ebiten.CursorPosition()
}

// GetMouseDelta returns the mouse movement since the last frame.
// BUG-008 fix: Essential for first-person camera control and aiming.
// Returns the change in X and Y coordinates from the previous frame.
func (s *InputSystem) GetMouseDelta() (dx, dy int) {
	return s.mouseDeltaX, s.mouseDeltaY
}

// GetMouseWheel returns the mouse wheel scroll delta for the current frame.
// BUG-007 fix: Explicitly documented feature for camera zoom.
// Positive Y = scroll up, negative Y = scroll down.
// X is for horizontal scroll wheels (less common).
func (s *InputSystem) GetMouseWheel() (deltaX, deltaY float64) {
	return ebiten.Wheel()
}

// ===== UNIFIED TOUCH/MOUSE INPUT HELPERS =====
// Touch support for WASM/mobile: These functions provide unified detection
// of both touch and mouse input, enabling full touchscreen gameplay.

// IsTouchOrMouseJustPressed returns true if either a touch or left mouse button
// was just pressed this frame. This unifies touch and mouse input for UI interactions.
// Touch support for WASM/mobile platforms.
func IsTouchOrMouseJustPressed() bool {
	// Check for mouse click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}

	// Check for new touch input
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	return len(touchIDs) > 0
}

// GetTouchOrMousePosition returns the position of either the first active touch
// or the mouse cursor. Prioritizes touch input when available.
// The hasActiveInput return value is true when there's an active touch or pressed mouse button,
// false when only cursor position is available without interaction.
// Touch support for WASM/mobile platforms.
func GetTouchOrMousePosition() (x, y int, hasActiveInput bool) {
	// Check for active touch first (priority on touch devices)
	touchIDs := ebiten.TouchIDs()
	if len(touchIDs) > 0 {
		// Return the first touch position with active input flag
		touchX, touchY := ebiten.TouchPosition(touchIDs[0])
		return touchX, touchY, true
	}

	// Fall back to mouse position
	x, y = ebiten.CursorPosition()
	// Return cursor position with active input flag based on mouse button state
	hasActiveInput = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	return x, y, hasActiveInput
}

// HasTouchOrMouseInput returns true if there is any active touch or mouse button press.
// Only checks left mouse button for consistency with touch input (no button distinction).
// Touch support for WASM/mobile platforms.
func HasTouchOrMouseInput() bool {
	// Check for active touches
	if len(ebiten.TouchIDs()) > 0 {
		return true
	}

	// Check for left mouse button press (consistent with touch tap behavior)
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// IsTouchOrMouseJustReleased returns true if either a touch or left mouse button
// was just released this frame. This is the release counterpart to IsTouchOrMouseJustPressed.
// Touch support for WASM/mobile platforms.
func IsTouchOrMouseJustReleased() bool {
	// Check for mouse release
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		return true
	}

	// Check for touch release (any touch that was active but is now gone)
	touchIDs := inpututil.AppendJustReleasedTouchIDs(nil)
	return len(touchIDs) > 0
}

// ===== KEY BINDING MANAGEMENT =====

// SetKeyBinding sets a specific key binding by action name.
// BUG-019 fix: Comprehensive key binding API supporting all 18 keys.
// Valid action names: "up", "down", "left", "right", "action", "useitem",
// "inventory", "character", "skills", "quests", "map", "crafting",
// "help", "quicksave", "quickload", "cycletargets"
func (s *InputSystem) SetKeyBinding(action string, key ebiten.Key) bool {
	switch action {
	// Movement
	case "up":
		s.KeyUp = key
	case "down":
		s.KeyDown = key
	case "left":
		s.KeyLeft = key
	case "right":
		s.KeyRight = key
	// Actions
	case "action":
		s.KeyAction = key
	case "useitem":
		s.KeyUseItem = key
	// UI
	case "inventory":
		s.KeyInventory = key
	case "character":
		s.KeyCharacter = key
	case "skills":
		s.KeySkills = key
	case "quests":
		s.KeyQuests = key
	case "map":
		s.KeyMap = key
	case "crafting":
		s.KeyCrafting = key
	// System
	case "help":
		s.KeyHelp = key
	case "quicksave":
		s.KeyQuickSave = key
	case "quickload":
		s.KeyQuickLoad = key
	case "fullscreen":
		s.KeyFullscreen = key
	case "cycletargets":
		s.KeyCycleTargets = key
	default:
		return false // Unknown action
	}
	return true
}

// GetKeyBinding returns the current key binding for the specified action.
// BUG-020 fix: Query API for displaying current bindings in settings UI.
// Returns (key, true) if action exists, (0, false) if unknown action.
func (s *InputSystem) GetKeyBinding(action string) (ebiten.Key, bool) {
	switch action {
	// Movement
	case "up":
		return s.KeyUp, true
	case "down":
		return s.KeyDown, true
	case "left":
		return s.KeyLeft, true
	case "right":
		return s.KeyRight, true
	// Actions
	case "action":
		return s.KeyAction, true
	case "useitem":
		return s.KeyUseItem, true
	// UI
	case "inventory":
		return s.KeyInventory, true
	case "character":
		return s.KeyCharacter, true
	case "skills":
		return s.KeySkills, true
	case "quests":
		return s.KeyQuests, true
	case "map":
		return s.KeyMap, true
	case "crafting":
		return s.KeyCrafting, true
	// System
	case "help":
		return s.KeyHelp, true
	case "quicksave":
		return s.KeyQuickSave, true
	case "quickload":
		return s.KeyQuickLoad, true
	case "fullscreen":
		return s.KeyFullscreen, true
	case "cycletargets":
		return s.KeyCycleTargets, true
	default:
		return 0, false
	}
}

// GetAllKeyBindings returns a map of all current key bindings.
// BUG-020 fix: Comprehensive query for settings UI display.
func (s *InputSystem) GetAllKeyBindings() map[string]ebiten.Key {
	return map[string]ebiten.Key{
		// Movement
		"up":    s.KeyUp,
		"down":  s.KeyDown,
		"left":  s.KeyLeft,
		"right": s.KeyRight,
		// Actions
		"action":  s.KeyAction,
		"useitem": s.KeyUseItem,
		// UI
		"inventory": s.KeyInventory,
		"character": s.KeyCharacter,
		"skills":    s.KeySkills,
		"quests":    s.KeyQuests,
		"map":       s.KeyMap,
		"crafting":  s.KeyCrafting,
		// System
		"help":         s.KeyHelp,
		"quicksave":    s.KeyQuickSave,
		"quickload":    s.KeyQuickLoad,
		"fullscreen":   s.KeyFullscreen,
		"cycletargets": s.KeyCycleTargets,
	}
}

// DrawVirtualControls renders virtual controls on screen (mobile and WASM).
// Should be called during the game's Draw phase.
func (s *InputSystem) DrawVirtualControls(screen *ebiten.Image) {
	if screen != nil && s.useTouchInput && s.virtualControls != nil {
		s.virtualControls.Draw(screen)
	}
}

// GetTouchHandler returns the touch input handler for advanced touch processing.
func (s *InputSystem) GetTouchHandler() *mobile.TouchInputHandler {
	return s.touchHandler
}

// GetGamepadHandler returns the gamepad input handler for advanced gamepad processing.
// Platform parity fix: Enables external access to gamepad state for custom handling.
func (s *InputSystem) GetGamepadHandler() *GamepadInputHandler {
	return s.gamepadHandler
}

// HasGamepad returns true if a gamepad is connected and active.
// Platform parity fix: Allows UI to show gamepad-specific prompts.
func (s *InputSystem) HasGamepad() bool {
	return s.useGamepadInput && s.gamepadHandler != nil && s.gamepadHandler.HasGamepad()
}

// SetVirtualControlsVisible controls visibility of virtual controls.
func (s *InputSystem) SetVirtualControlsVisible(visible bool) {
	if s.virtualControls != nil {
		s.virtualControls.SetVisible(visible)
	}
}

// SetSpellButtonsVisible controls visibility of spell buttons on virtual controls.
// Platform parity fix: Allows hiding spell buttons when not needed.
func (s *InputSystem) SetSpellButtonsVisible(visible bool) {
	if s.virtualControls != nil {
		s.virtualControls.SetSpellButtonsVisible(visible)
	}
}

// ResizeVirtualControls recalculates virtual control positions for new screen size.
// Platform parity fix: Handles orientation changes on mobile/tablet devices.
func (s *InputSystem) ResizeVirtualControls(screenWidth, screenHeight int) {
	if s.virtualControls != nil {
		s.virtualControls.Resize(screenWidth, screenHeight)
	}
}

// OnScreenResize implements the screenDimensionConsumer interface so that
// virtual controls are repositioned whenever the window or device orientation
// changes. This is called by EbitenGame.propagateScreenResize.
func (s *InputSystem) OnScreenResize(width, height int) {
	s.ResizeVirtualControls(width, height)
}
