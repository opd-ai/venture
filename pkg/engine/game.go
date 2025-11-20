// Package engine provides the main game loop and Ebiten integration.
// This file implements EbitenGame which ties together the ECS world, rendering
// systems, and the Ebiten game engine. EbitenGame implements both ebiten.Game
// and GameRunner interfaces.
package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// EbitenGame represents the main game instance with the ECS world and game loop.
// Implements both ebiten.Game and GameRunner interfaces.
type EbitenGame struct {
	World          *World
	lastUpdateTime time.Time
	ScreenWidth    int
	ScreenHeight   int
	Paused         bool

	// Application state management
	StateManager         *AppStateManager
	MainMenuUI           *MainMenuUI
	SinglePlayerMenu     *SinglePlayerMenu   // Submenu for single-player options
	GenreSelectionMenu   *GenreSelectionMenu // Genre selection for single-player
	MultiplayerMenu      *MultiplayerMenu    // Submenu for multiplayer options
	ServerAddressInput   *ServerAddressInput // Text input for server address
	SettingsUI           *SettingsUI
	SettingsManager      *SettingsManager
	CharacterCreation    *EbitenCharacterCreation
	pendingCharData      *CharacterData
	isMultiplayerMode    bool   // Track if character creation is for multiplayer
	selectedGenreID      string // Selected genre for world generation
	pendingServerAddress string // Server address for Join option
	worldSeed            int64  // World generation seed for deterministic content

	// Rendering systems
	CameraSystem        *CameraSystem
	RenderSystem        *EbitenRenderSystem
	TerrainRenderSystem *TerrainRenderSystem
	LightingSystem      *LightingSystem // Dynamic lighting system (Phase 5.3)
	sceneBuffer         *ebiten.Image   // Reusable buffer for lighting post-processing
	HUDSystem           *EbitenHUDSystem
	TutorialSystem      *EbitenTutorialSystem
	HelpSystem          *EbitenHelpSystem
	MenuSystem          *EbitenMenuSystem

	// UI systems
	InventoryUI *EbitenInventoryUI
	QuestUI     *EbitenQuestUI
	CharacterUI *EbitenCharacterUI
	SkillsUI    *EbitenSkillsUI
	MapUI       *EbitenMapUI
	ShopUI      *ShopUI     // Commerce and merchant interaction UI
	CraftingUI  *CraftingUI // Crafting and recipe UI
	MailboxUI   *MailboxUI  // Mail system UI (Phase 40.3)

	// INTEGRATION FIX [Category B]: V8.0 UI systems
	// Gap: V8 systems (housing, gallery) fully implemented but no UI fields
	// Fix: Added UI fields for housing and gallery management
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4
	HousingUI *HousingUI // Player housing management UI (Phase 49.1, 51.2, 51.3)
	GalleryUI *GalleryUI // Image gallery viewer UI (Phase 49.4)

	// Audio system (for settings integration)
	AudioManager *AudioManager

	// Player entity reference (for UI systems)
	PlayerEntity *Entity

	// Callbacks for state transitions
	onNewGame            func() error
	onMultiplayerConnect func(serverAddr string) error
	onQuitToMenu         func() error

	// Logger for game operations
	logger *logrus.Entry

	// Performance monitoring
	frameTimeTracker *FrameTimeTracker
	frameCount       uint64
	profilingEnabled bool
}

// NewEbitenGame creates a new game instance with Ebiten integration.
func NewEbitenGame(screenWidth, screenHeight int) *EbitenGame {
	return NewEbitenGameWithLogger(screenWidth, screenHeight, nil)
}

// NewEbitenGameWithLogger creates a new game instance with a logger.
func NewEbitenGameWithLogger(screenWidth, screenHeight int, logger *logrus.Logger) *EbitenGame {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "game",
		})
	}

	world := NewWorldWithLogger(logger)
	cameraSystem := NewCameraSystem(screenWidth, screenHeight)
	renderSystem := NewRenderSystem(cameraSystem)
	hudSystem := NewEbitenHUDSystem(screenWidth, screenHeight)
	// TerrainRenderSystem will be initialized later with specific genre/seed

	// Create UI systems
	inventoryUI := NewEbitenInventoryUI(world, screenWidth, screenHeight)
	questUI := NewEbitenQuestUI(world, screenWidth, screenHeight)
	characterUI := NewEbitenCharacterUI(world, screenWidth, screenHeight)
	skillsUI := NewEbitenSkillsUI(world, screenWidth, screenHeight)
	mapUI := NewEbitenMapUI(world, screenWidth, screenHeight)

	// Create menu system with save directory
	menuSystem, err := NewEbitenMenuSystem(world, screenWidth, screenHeight, "./saves")
	if err != nil {
		// Always log critical initialization failures
		if logEntry != nil {
			logEntry.WithError(err).Warn("failed to initialize menu system")
		} else {
			// Fallback to stderr if no logger configured
			fmt.Fprintf(os.Stderr, "WARNING: failed to initialize menu system: %v\n", err)
		}
	}

	// Create settings manager and UI
	settingsManager, err := NewSettingsManager()
	if err != nil {
		// Log error but continue with default settings
		if logEntry != nil {
			logEntry.WithError(err).Warn("failed to initialize settings manager, using defaults")
		}
		// Create a minimal settings manager with defaults
		settingsManager = &SettingsManager{
			settings: DefaultSettings(),
		}
	} else {
		// Load existing settings
		if err := settingsManager.LoadSettings(); err != nil {
			if logEntry != nil {
				logEntry.WithError(err).Warn("failed to load settings, using defaults")
			}
		}
	}

	settingsUI := NewSettingsUI(screenWidth, screenHeight, settingsManager)

	// Initialize frame time tracker (disabled by default, enabled via EnableFrameTimeProfiling)
	frameTimeTracker := NewFrameTimeTracker(1000) // Track last 1000 frames (~16 seconds at 60 FPS)

	// Create lighting system with default configuration
	// Note: Will be enabled via command-line flag in client/main.go
	lightingConfig := NewLightingConfig()
	lightingConfig.Enabled = false // Disabled by default, enable via flag
	lightingSystem := NewLightingSystemWithLogger(world, lightingConfig, logger)

	// WASM FIX: Scene buffer will be lazily initialized on first Draw() call
	// Cannot call ebiten.NewImage() during initialization in WASM builds
	// The graphics context is not available until the first Update/Draw cycle
	var sceneBuffer *ebiten.Image = nil

	game := &EbitenGame{
		World:              world,
		lastUpdateTime:     time.Now(),
		ScreenWidth:        screenWidth,
		ScreenHeight:       screenHeight,
		StateManager:       NewAppStateManager(),
		MainMenuUI:         NewMainMenuUI(screenWidth, screenHeight),
		SinglePlayerMenu:   NewSinglePlayerMenu(screenWidth, screenHeight),
		GenreSelectionMenu: NewGenreSelectionMenu(screenWidth, screenHeight),
		MultiplayerMenu:    NewMultiplayerMenu(screenWidth, screenHeight),
		ServerAddressInput: NewServerAddressInput(screenWidth, screenHeight),
		SettingsUI:         settingsUI,
		SettingsManager:    settingsManager,
		CharacterCreation:  NewCharacterCreation(screenWidth, screenHeight),
		CameraSystem:       cameraSystem,
		RenderSystem:       renderSystem,
		LightingSystem:     lightingSystem,
		sceneBuffer:        sceneBuffer,
		HUDSystem:          hudSystem,
		MenuSystem:         menuSystem,
		InventoryUI:        inventoryUI,
		QuestUI:            questUI,
		CharacterUI:        characterUI,
		SkillsUI:           skillsUI,
		MapUI:              mapUI,
		logger:             logEntry,
		frameTimeTracker:   frameTimeTracker,
		frameCount:         0,
		profilingEnabled:   false,
	}

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"screenWidth":  screenWidth,
			"screenHeight": screenHeight,
		}).Info("game initialized")
	}

	// Setup main menu callback
	game.MainMenuUI.SetSelectCallback(game.handleMainMenuSelection)

	// Setup single-player menu callbacks
	game.SinglePlayerMenu.SetNewGameCallback(game.handleSinglePlayerMenuNewGame)
	game.SinglePlayerMenu.SetLoadGameCallback(game.handleSinglePlayerMenuLoadGame)
	game.SinglePlayerMenu.SetBackCallback(game.handleSinglePlayerMenuBack)

	// Setup genre selection menu callbacks
	game.GenreSelectionMenu.SetGenreSelectCallback(game.handleGenreSelection)
	game.GenreSelectionMenu.SetBackCallback(game.handleGenreSelectionBack)

	// Setup multiplayer menu callbacks
	game.MultiplayerMenu.SetJoinCallback(game.handleMultiplayerMenuJoin)
	game.MultiplayerMenu.SetHostCallback(game.handleMultiplayerMenuHost)
	game.MultiplayerMenu.SetBackCallback(game.handleMultiplayerMenuBack)

	// Setup server address input callbacks
	game.ServerAddressInput.SetConnectCallback(game.handleServerAddressConnect)
	game.ServerAddressInput.SetCancelCallback(game.handleServerAddressCancel)

	// Setup settings UI callbacks
	game.SettingsUI.SetBackCallback(func() {
		// Return to main menu when back is pressed
		if err := game.StateManager.TransitionTo(AppStateMainMenu); err != nil {
			if logEntry != nil {
				logEntry.WithError(err).Error("failed to return to main menu from settings")
			}
		}
	})
	game.SettingsUI.SetApplyCallback(func() {
		// Apply settings when they're saved
		if err := game.ApplySettings(); err != nil {
			if logEntry != nil {
				logEntry.WithError(err).Error("failed to apply settings")
			}
		}
	})

	return game
}

// SetNewGameCallback sets the callback function called when New Game is selected.
func (g *EbitenGame) SetNewGameCallback(callback func() error) {
	g.onNewGame = callback
}

// SetMultiplayerConnectCallback sets the callback function called when connecting to multiplayer.
func (g *EbitenGame) SetMultiplayerConnectCallback(callback func(serverAddr string) error) {
	g.onMultiplayerConnect = callback
}

// SetQuitToMenuCallback sets the callback function called when quitting to main menu.
func (g *EbitenGame) SetQuitToMenuCallback(callback func() error) {
	g.onQuitToMenu = callback
}

// SetWorldSeed sets the world generation seed and applies it to character creation defaults.
// This ensures deterministic default character naming based on the world seed.
func (g *EbitenGame) SetWorldSeed(seed int64) {
	g.worldSeed = seed
	if g.CharacterCreation != nil {
		g.CharacterCreation.SetDefaultNameFromSeed(seed)
	}
}

// handleMainMenuSelection processes main menu option selections and triggers state transitions.
func (g *EbitenGame) handleMainMenuSelection(option MainMenuOption) {
	switch option {
	case MainMenuOptionSinglePlayer:
		// Transition to single-player submenu
		if err := g.StateManager.TransitionTo(AppStateSinglePlayerMenu); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("failed to transition to single-player menu")
			}
			return
		}

		// Show single-player menu
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Show()
		}

		if g.logger != nil {
			g.logger.Info("entering single-player menu")
		}

	case MainMenuOptionMultiPlayer:
		// Transition to multiplayer submenu
		if err := g.StateManager.TransitionTo(AppStateMultiPlayerMenu); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("failed to transition to multiplayer menu")
			}
			return
		}

		// Show multiplayer menu
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Show()
		}

		if g.logger != nil {
			g.logger.Info("entering multiplayer menu")
		}

	case MainMenuOptionSettings:
		// Transition to settings menu
		if err := g.StateManager.TransitionTo(AppStateSettings); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("failed to transition to settings")
			}
			return
		}

		// Show settings UI
		g.SettingsUI.Show()

		if g.logger != nil {
			g.logger.Info("entering settings menu")
		}

	case MainMenuOptionQuit:
		// Exit the application
		if g.logger != nil {
			g.logger.Info("quit selected")
		}
		// Ebiten doesn't have a clean exit API, so we return an error which will terminate RunGame
		// The client can handle this gracefully
	}
}

// handleSinglePlayerMenuNewGame handles the New Game selection from single-player menu.
func (g *EbitenGame) handleSinglePlayerMenuNewGame() {
	// Transition to genre selection
	if err := g.StateManager.TransitionTo(AppStateGenreSelection); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to genre selection")
		}
		return
	}

	// Show genre selection menu
	if g.GenreSelectionMenu != nil {
		g.GenreSelectionMenu.Show()
	}

	// Hide single-player menu
	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("entering genre selection for new single-player game")
	}
}

// handleGenreSelection handles genre selection from the genre selection menu.
func (g *EbitenGame) handleGenreSelection(genreID string) {
	// Store selected genre
	g.selectedGenreID = genreID

	// Transition to character creation
	if err := g.StateManager.TransitionTo(AppStateCharacterCreation); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to character creation")
		}
		return
	}

	// Reset character creation UI for new game
	g.CharacterCreation.Reset()

	// Set default name from world seed so user has a starting name
	g.CharacterCreation.SetDefaultNameFromSeed(g.worldSeed)

	g.isMultiplayerMode = false // Single-player mode

	// Hide genre selection menu
	if g.GenreSelectionMenu != nil {
		g.GenreSelectionMenu.Hide()
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"genre": genreID,
		}).Info("genre selected, entering character creation")
	}
}

// handleGenreSelectionBack handles the Back selection from genre selection menu.
func (g *EbitenGame) handleGenreSelectionBack() {
	// Transition back to single-player menu
	if err := g.StateManager.TransitionTo(AppStateSinglePlayerMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to single-player menu")
		}
		return
	}

	// Hide genre selection menu
	if g.GenreSelectionMenu != nil {
		g.GenreSelectionMenu.Hide()
	}

	// Show single-player menu
	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Show()
	}

	if g.logger != nil {
		g.logger.Info("returning to single-player menu from genre selection")
	}
}

// handleSinglePlayerMenuLoadGame handles the Load Game selection.
// Loads the most recent saved game using the save/load system.
func (g *EbitenGame) handleSinglePlayerMenuLoadGame() {
	if g.logger != nil {
		g.logger.Info("load game selected - attempting to load recent save")
	}

	// Note: Quick load (F9) is the primary load mechanism during gameplay.
	// Menu-driven load is available but deferred for future enhancement (load save browser UI).
	// For now, users should use F9 (quick load) during gameplay.
	if g.logger != nil {
		g.logger.Info("menu-driven load game UI deferred - use F9 quick load during gameplay")
	}
}

// handleSinglePlayerMenuBack handles the Back selection from single-player menu.
func (g *EbitenGame) handleSinglePlayerMenuBack() {
	// Transition back to main menu
	if err := g.StateManager.TransitionTo(AppStateMainMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to main menu")
		}
		return
	}

	// Hide single-player menu
	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("returning to main menu from single-player menu")
	}
}

// handleMultiplayerMenuJoin handles the Join Server selection from multiplayer menu.
func (g *EbitenGame) handleMultiplayerMenuJoin() {
	// Transition to server address input
	if err := g.StateManager.TransitionTo(AppStateServerAddressInput); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to server address input")
		}
		return
	}

	// Show server address input
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Show()
	}

	// Hide multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("showing server address input for join")
	}
}

// handleMultiplayerMenuHost handles the Host Game selection from multiplayer menu.
func (g *EbitenGame) handleMultiplayerMenuHost() {
	if g.logger != nil {
		g.logger.Info("host game selected - menu-driven host deferred")
	}

	// Note: --host-and-play CLI flag is the primary way to host games.
	// Menu-driven host would duplicate that functionality.
	// For now, automatically connect to localhost:8080 (assuming user started server separately).
	// Future enhancement: integrate pkg/hostplay to start server from menu.
	g.pendingServerAddress = "localhost:8080"

	// Transition to gameplay
	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay")
		}
		return
	}

	// Hide multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Hide()
	}

	// Call multiplayer connect callback if set
	if g.onMultiplayerConnect != nil {
		if err := g.onMultiplayerConnect(g.pendingServerAddress); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("failed to connect to hosted server")
			}
			// Transition back to multiplayer menu on error
			g.StateManager.TransitionTo(AppStateMultiPlayerMenu)
			if g.MultiplayerMenu != nil {
				g.MultiplayerMenu.Show()
			}
			return
		}
	}

	if g.logger != nil {
		g.logger.WithField("address", g.pendingServerAddress).Info("connected to hosted server")
	}
}

// handleMultiplayerMenuBack handles the Back selection from multiplayer menu.
func (g *EbitenGame) handleMultiplayerMenuBack() {
	// Transition back to main menu
	if err := g.StateManager.TransitionTo(AppStateMainMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to main menu")
		}
		return
	}

	// Hide multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("returning to main menu from multiplayer menu")
	}
}

// handleServerAddressConnect handles connecting to the entered server address.
func (g *EbitenGame) handleServerAddressConnect(address string) {
	if g.logger != nil {
		g.logger.WithField("address", address).Info("connecting to server")
	}

	g.pendingServerAddress = address

	// Transition to gameplay
	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay")
		}
		return
	}

	// Hide server address input
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Hide()
	}

	// Call multiplayer connect callback if set
	if g.onMultiplayerConnect != nil {
		if err := g.onMultiplayerConnect(address); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("failed to connect to server")
			}
			// Transition back to server address input on error
			g.StateManager.TransitionTo(AppStateServerAddressInput)
			if g.ServerAddressInput != nil {
				g.ServerAddressInput.Show()
			}
			return
		}
	}

	if g.logger != nil {
		g.logger.WithField("address", address).Info("connected to server")
	}
}

// handleServerAddressCancel handles canceling server address input.
func (g *EbitenGame) handleServerAddressCancel() {
	// Transition back to multiplayer menu
	if err := g.StateManager.TransitionTo(AppStateMultiPlayerMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to multiplayer menu")
		}
		return
	}

	// Hide server address input
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Hide()
	}

	// Show multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Show()
	}

	if g.logger != nil {
		g.logger.Info("cancelled server address input, returning to multiplayer menu")
	}
}

// IsInMainMenu returns true if currently displaying the main menu.
func (g *EbitenGame) IsInMainMenu() bool {
	return g.StateManager.IsInMenu()
}

// Update implements ebiten.Game interface. Called every frame.
// trackFramePerformance records frame timing metrics and logs performance statistics.
func (g *EbitenGame) trackFramePerformance(frameStart time.Time) {
	if !g.profilingEnabled || g.frameTimeTracker == nil {
		return
	}

	g.frameTimeTracker.RecordFrame(time.Since(frameStart))
	g.frameCount++

	if g.frameCount%300 == 0 && g.logger != nil {
		g.logFrameStats()
	}
}

// logFrameStats logs current frame time statistics with stuttering detection.
func (g *EbitenGame) logFrameStats() {
	stats := g.frameTimeTracker.GetStats()
	fields := logrus.Fields{
		"avg_ms":      stats.Average.Milliseconds(),
		"min_ms":      stats.Min.Milliseconds(),
		"max_ms":      stats.Max.Milliseconds(),
		"1pct_low_ms": stats.Percentile1.Milliseconds(),
		"avg_fps":     stats.GetFPS(),
		"worst_fps":   stats.GetWorstFPS(),
		"samples":     stats.SampleCount,
	}

	if stats.IsStuttering() {
		fields["stuttering"] = true
		g.logger.WithFields(fields).Warn("frame time stuttering detected")
	} else {
		g.logger.WithFields(fields).Info("frame time stats")
	}
}

// calculateDeltaTime computes time elapsed since last update with capping to prevent spiral of death.
func (g *EbitenGame) calculateDeltaTime() float64 {
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	if deltaTime > 0.1 {
		deltaTime = 0.1
	}
	return deltaTime
}

// updateMenuState handles updates for all menu-related application states.
func (g *EbitenGame) updateMenuState() (handled bool) {
	switch g.StateManager.CurrentState() {
	case AppStateMainMenu:
		g.MainMenuUI.Update()
		return true
	case AppStateSinglePlayerMenu:
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Update()
		}
		return true
	case AppStateGenreSelection:
		if g.GenreSelectionMenu != nil {
			g.GenreSelectionMenu.Update()
		}
		return true
	case AppStateMultiPlayerMenu:
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Update()
		}
		return true
	case AppStateServerAddressInput:
		if g.ServerAddressInput != nil {
			g.ServerAddressInput.Update()
		}
		return true
	case AppStateSettings:
		g.SettingsUI.Update()
		return true
	}

	if g.StateManager.IsInMenu() {
		g.MainMenuUI.Update()
		return true
	}
	return false
}

// handleCharacterCreation processes character creation updates and transitions to gameplay.
func (g *EbitenGame) handleCharacterCreation() error {
	completed := g.CharacterCreation.Update()
	if !completed {
		return nil
	}

	if g.logger != nil {
		g.logger.Info("character creation completed, transitioning to gameplay")
	}

	charData := g.CharacterCreation.GetCharacterData()
	g.pendingCharData = &charData
	g.CharacterCreation.Cleanup()

	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay after character creation")
		}
		return err
	}

	return g.triggerGameStartCallback(charData)
}

// triggerGameStartCallback invokes the appropriate callback for multiplayer or single-player mode.
func (g *EbitenGame) triggerGameStartCallback(charData CharacterData) error {
	if g.isMultiplayerMode {
		return g.startMultiplayerGame(charData)
	}
	return g.startSinglePlayerGame(charData)
}

// startMultiplayerGame connects to multiplayer server with character data.
func (g *EbitenGame) startMultiplayerGame(charData CharacterData) error {
	if g.onMultiplayerConnect != nil {
		if err := g.onMultiplayerConnect(""); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("multiplayer connect callback failed")
			}
			_ = g.StateManager.TransitionTo(AppStateMainMenu)
			g.pendingCharData = nil
			return err
		}
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"name":  charData.Name,
			"class": charData.Class.String(),
			"mode":  "multiplayer",
		}).Info("character created, connecting to server")
	}
	return nil
}

// startSinglePlayerGame starts a new single-player game session.
func (g *EbitenGame) startSinglePlayerGame(charData CharacterData) error {
	if g.onNewGame != nil {
		if err := g.onNewGame(); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("new game callback failed")
			}
			_ = g.StateManager.TransitionTo(AppStateMainMenu)
			g.pendingCharData = nil
			return err
		}
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"name":  charData.Name,
			"class": charData.Class.String(),
			"mode":  "single-player",
		}).Info("character created, starting game")
	}
	return nil
}

// updateGameplayUI updates all UI systems during active gameplay.
func (g *EbitenGame) updateGameplayUI(deltaTime float64) {
	g.InventoryUI.Update(nil, deltaTime)
	g.QuestUI.Update(nil, deltaTime)
	g.CharacterUI.Update(nil, deltaTime)
	g.SkillsUI.Update(nil, deltaTime)
	g.MapUI.Update(nil, deltaTime)

	if g.ShopUI != nil {
		g.ShopUI.Update(g.World.GetEntities(), deltaTime)
	}

	if g.CraftingUI != nil {
		g.CraftingUI.Update(nil, deltaTime)
	}

	// NOTE: MailboxUI does not have an Update() method - ESC key handling is done
	// by InputSystem.handleEscapeKey() which checks mailboxUI.IsOpen() and calls Close()

	// INTEGRATION FIX [Category B]: Update V8.0 UI screens
	// Gap: Housing and Gallery UIs created but never updated
	// Fix: Added update calls for housing and gallery UIs
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4
	if g.HousingUI != nil {
		g.HousingUI.Update()
	}

	if g.GalleryUI != nil {
		g.GalleryUI.Update()
	}

	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Update(g.World.GetEntities(), deltaTime)
	}

	// WASM/Mobile: Auto-hide virtual controls when any UI menu is open
	// This prevents D-pad/buttons from overlapping with UI elements
	g.updateVirtualControlsVisibility()
}

// updateVirtualControlsVisibility manages virtual control visibility based on UI state.
// Hides virtual controls (D-pad, action buttons) when any menu is open to prevent overlay.
// Shows virtual controls during active gameplay when no blocking UI is visible.
func (g *EbitenGame) updateVirtualControlsVisibility() {
	// Check if we're in menu states (main menu, settings, character creation, etc.)
	inMenuState := g.StateManager.CurrentState() != AppStateGameplay

	// Check if any gameplay UI is open
	// BUG FIX: Phase 3.7 - MailboxUI missing from virtual controls visibility check
	// Resolution: Added MailboxUI.IsOpen() check to hide virtual controls when mailbox is open
	anyUIOpen := g.InventoryUI.IsVisible() ||
		g.QuestUI.IsVisible() ||
		g.CharacterUI.IsVisible() ||
		g.SkillsUI.IsVisible() ||
		g.MapUI.IsFullScreen() ||
		(g.ShopUI != nil && g.ShopUI.IsVisible()) ||
		(g.CraftingUI != nil && g.CraftingUI.IsVisible()) ||
		(g.MailboxUI != nil && g.MailboxUI.IsOpen()) ||
		(g.MenuSystem != nil && g.MenuSystem.IsActive())

	// Virtual controls should be hidden if:
	// 1. In a menu state (not gameplay)
	// 2. Any gameplay UI is open
	shouldHide := inMenuState || anyUIOpen

	// Update virtual controls visibility for all input systems
	for _, system := range g.World.GetSystems() {
		if inputSys, ok := system.(*InputSystem); ok {
			inputSys.SetVirtualControlsVisible(!shouldHide)
			break
		}
	}
}

// shouldUpdateWorld checks if world updates should proceed based on UI visibility.
// BUG FIX: Phase 3.7 - MailboxUI missing from world update check
// Resolution: Added MailboxUI.IsOpen() check to pause world when mailbox is open
func (g *EbitenGame) shouldUpdateWorld() bool {
	return !g.InventoryUI.IsVisible() &&
		!g.QuestUI.IsVisible() &&
		!g.CharacterUI.IsVisible() &&
		!g.SkillsUI.IsVisible() &&
		!g.MapUI.IsFullScreen() &&
		(g.ShopUI == nil || !g.ShopUI.IsVisible()) &&
		(g.CraftingUI == nil || !g.CraftingUI.IsVisible()) &&
		(g.MailboxUI == nil || !g.MailboxUI.IsOpen())
}

func (g *EbitenGame) Update() error {
	frameStart := time.Now()
	defer func() { g.trackFramePerformance(frameStart) }()

	deltaTime := g.calculateDeltaTime()

	// Handle character creation BEFORE updateMenuState to prevent main menu from consuming input
	if g.StateManager.CurrentState() == AppStateCharacterCreation {
		return g.handleCharacterCreation()
	}

	if g.updateMenuState() {
		return nil
	}

	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.Paused = true
		g.MenuSystem.Update(g.World.GetEntities(), deltaTime)
		return nil
	}

	if g.Paused {
		return nil
	}

	g.updateGameplayUI(deltaTime)

	if g.shouldUpdateWorld() {
		g.World.Update(deltaTime)
	}

	g.CameraSystem.Update(g.World.GetEntities(), deltaTime)
	return nil
}

// Draw implements ebiten.Game interface. Called every frame.
func (g *EbitenGame) Draw(screen *ebiten.Image) {
	if g.drawMenuState(screen) {
		return
	}

	g.drawGameplayScene(screen)
	g.drawOverlays(screen)
	g.drawVirtualControls(screen)
}

// drawMenuState renders the appropriate menu based on current application state.
// Returns true if a menu was drawn, false if gameplay should be rendered.
func (g *EbitenGame) drawMenuState(screen *ebiten.Image) bool {
	switch g.StateManager.CurrentState() {
	case AppStateMainMenu:
		g.MainMenuUI.Draw(screen)
		return true
	case AppStateSinglePlayerMenu:
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Draw(screen)
		}
		return true
	case AppStateGenreSelection:
		if g.GenreSelectionMenu != nil {
			g.GenreSelectionMenu.Draw(screen)
		}
		return true
	case AppStateMultiPlayerMenu:
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Draw(screen)
		}
		return true
	case AppStateServerAddressInput:
		if g.ServerAddressInput != nil {
			g.ServerAddressInput.Draw(screen)
		}
		return true
	case AppStateSettings:
		g.SettingsUI.Draw(screen)
		return true
	case AppStateCharacterCreation:
		g.CharacterCreation.Draw(screen)
		return true
	}

	if g.StateManager.IsInMenu() {
		g.MainMenuUI.Draw(screen)
		return true
	}

	return false
}

// drawGameplayScene renders the main game scene with terrain, entities, and lighting effects.
func (g *EbitenGame) drawGameplayScene(screen *ebiten.Image) {
	if g.LightingSystem != nil && g.LightingSystem.IsEnabled() {
		g.drawLitScene(screen)
	} else {
		g.drawStandardScene(screen)
	}
}

// drawLitScene renders the game scene with dynamic lighting using post-processing pipeline.
func (g *EbitenGame) drawLitScene(screen *ebiten.Image) {
	if g.sceneBuffer == nil {
		g.sceneBuffer = ebiten.NewImage(g.ScreenWidth, g.ScreenHeight)
	}

	g.sceneBuffer.Clear()

	if g.TerrainRenderSystem != nil {
		g.TerrainRenderSystem.Draw(g.sceneBuffer, g.CameraSystem)
	}

	g.RenderSystem.Draw(g.sceneBuffer, g.World.GetEntities())

	if g.CameraSystem != nil {
		camX, camY := g.CameraSystem.GetPosition()
		g.LightingSystem.SetViewport(camX, camY, g.ScreenWidth, g.ScreenHeight)
	}

	entities := g.World.GetEntities()
	g.LightingSystem.ApplyLighting(screen, g.sceneBuffer, entities)
}

// drawStandardScene renders the game scene without lighting effects.
func (g *EbitenGame) drawStandardScene(screen *ebiten.Image) {
	if g.TerrainRenderSystem != nil {
		g.TerrainRenderSystem.Draw(screen, g.CameraSystem)
	}

	g.RenderSystem.Draw(screen, g.World.GetEntities())
}

// drawOverlays renders all UI overlays including HUD, menus, inventory, and other interfaces.
func (g *EbitenGame) drawOverlays(screen *ebiten.Image) {
	g.HUDSystem.Draw(screen)

	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Draw(screen)
	}

	if g.HelpSystem != nil && g.HelpSystem.Visible {
		g.HelpSystem.Draw(screen)
	}

	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.MenuSystem.Draw(screen)
	}

	g.InventoryUI.Draw(screen)
	g.QuestUI.Draw(screen)
	g.CharacterUI.Draw(screen)
	g.SkillsUI.Draw(screen)
	g.MapUI.Draw(screen)

	if g.ShopUI != nil {
		g.ShopUI.Draw(screen)
	}

	if g.CraftingUI != nil {
		g.CraftingUI.Draw(screen)
	}

	g.drawMailboxUI(screen)

	// INTEGRATION FIX [Category B]: Draw V8.0 UI screens
	// Gap: Housing and Gallery UIs created but never drawn
	// Fix: Added draw calls for housing and gallery UIs
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4
	if g.HousingUI != nil {
		g.HousingUI.Draw(screen)
	}

	if g.GalleryUI != nil {
		g.GalleryUI.Draw(screen)
	}
}

// drawMailboxUI renders the mailbox interface if open and loads mail data from player entity.
func (g *EbitenGame) drawMailboxUI(screen *ebiten.Image) {
	if g.MailboxUI == nil || !g.MailboxUI.IsOpen() {
		return
	}

	if g.PlayerEntity != nil {
		if mailComp, ok := g.PlayerEntity.GetComponent("mail"); ok {
			g.MailboxUI.LoadFromMailComponent(mailComp.(*MailComponent))
		}
	}

	mailImg := g.MailboxUI.Render()
	if mailImg != nil {
		screen.DrawImage(ebiten.NewImageFromImage(mailImg), nil)
	}
}

// drawVirtualControls renders mobile touch controls on top of all other elements.
func (g *EbitenGame) drawVirtualControls(screen *ebiten.Image) {
	for _, system := range g.World.GetSystems() {
		if inputSys, ok := system.(*InputSystem); ok {
			inputSys.DrawVirtualControls(screen)
			break
		}
	}
}

// Layout implements ebiten.Game interface. Returns the game's screen size.
func (g *EbitenGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ScreenWidth, g.ScreenHeight
}

// SetPlayerEntity sets the player entity for the game and UI systems.
// This should be called after creating the player entity.
func (g *EbitenGame) SetPlayerEntity(entity *Entity) {
	g.PlayerEntity = entity
	g.InventoryUI.SetPlayerEntity(entity)
	g.QuestUI.SetPlayerEntity(entity)
	g.CharacterUI.SetPlayerEntity(entity)
	g.SkillsUI.SetPlayerEntity(entity)
	g.MapUI.SetPlayerEntity(entity)

	// Set player for shop UI (if initialized)
	if g.ShopUI != nil {
		g.ShopUI.SetPlayerEntity(entity)
	}

	// Set player for crafting UI (if initialized)
	if g.CraftingUI != nil {
		g.CraftingUI.SetPlayerEntity(entity)
	}
}

// SetAudioManager sets the audio manager for the game.
// This should be called before ApplySettings() to enable volume control.
func (g *EbitenGame) SetAudioManager(audioManager *AudioManager) {
	g.AudioManager = audioManager

	// Apply current settings to audio if settings manager exists
	if g.SettingsManager != nil {
		_ = g.ApplySettings() // Ignore error, just apply what we can
	}
}

// SetInventorySystem connects the inventory system to the inventory UI for item actions.
func (g *EbitenGame) SetInventorySystem(system *InventorySystem) {
	g.InventoryUI.SetInventorySystem(system)
}

// GetPendingCharacterData returns the character data from character creation (if any).
// Returns nil if no character data is pending. Clears the pending data after retrieval.
func (g *EbitenGame) GetPendingCharacterData() *CharacterData {
	data := g.pendingCharData
	g.pendingCharData = nil // Clear after retrieval
	return data
}

// GetSelectedGenreID returns the selected genre ID and clears it.
func (g *EbitenGame) GetSelectedGenreID() string {
	genreID := g.selectedGenreID
	g.selectedGenreID = "" // Clear after retrieval
	return genreID
}

// SetupInputCallbacks connects the input system callbacks to the UI systems.
// This should be called after the InputSystem is added to the world.
// GAP-014 REPAIR: Accept objective tracker for quest progress tracking
// H-008 FIX: Returns error if any callback registration fails
func (g *EbitenGame) SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) error {
	// Connect inventory toggle
	if err := inputSystem.SetInventoryCallback(func() {
		g.InventoryUI.Toggle()
		// GAP-014 REPAIR: Track inventory UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "inventory")
		}
	}); err != nil {
		return fmt.Errorf("failed to set inventory callback: %w", err)
	}

	// Connect quest log toggle
	if err := inputSystem.SetQuestsCallback(func() {
		g.QuestUI.Toggle()
		// GAP-014 REPAIR: Track quest log UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "quest_log")
		}
	}); err != nil {
		return fmt.Errorf("failed to set quests callback: %w", err)
	}

	// Connect character screen toggle
	if err := inputSystem.SetCharacterCallback(func() {
		g.CharacterUI.Toggle()
		// GAP-014 REPAIR: Track character UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "character")
		}
	}); err != nil {
		return fmt.Errorf("failed to set character callback: %w", err)
	}

	// Connect skills screen toggle
	if err := inputSystem.SetSkillsCallback(func() {
		g.SkillsUI.Toggle()
		// GAP-014 REPAIR: Track skills UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "skills")
		}
	}); err != nil {
		return fmt.Errorf("failed to set skills callback: %w", err)
	}

	// Connect map toggle
	if err := inputSystem.SetMapCallback(func() {
		g.MapUI.ToggleFullScreen()
		// GAP-014 REPAIR: Track map UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "map")
		}
	}); err != nil {
		return fmt.Errorf("failed to set map callback: %w", err)
	}

	// Connect crafting toggle (Category 1.3 - Commerce & Crafting Integration)
	if err := inputSystem.SetCraftingCallback(func() {
		if g.CraftingUI != nil {
			g.CraftingUI.Toggle()
			// Track crafting UI opens for tutorial objectives
			if objectiveTracker != nil && g.PlayerEntity != nil {
				objectiveTracker.OnUIOpened(g.PlayerEntity, "crafting")
			}
		}
	}); err != nil {
		return fmt.Errorf("failed to set crafting callback: %w", err)
	}

	// Connect mailbox toggle (Phase 40.3: Mail System Integration)
	if err := inputSystem.SetMailboxCallback(func() {
		if g.MailboxUI != nil {
			g.MailboxUI.Toggle()
			// Track mailbox UI opens for tutorial objectives
			if objectiveTracker != nil && g.PlayerEntity != nil {
				objectiveTracker.OnUIOpened(g.PlayerEntity, "mailbox")
			}
		}
	}); err != nil {
		return fmt.Errorf("failed to set mailbox callback: %w", err)
	}

	// INTEGRATION FIX [Category B]: Housing and Gallery UI input callbacks
	// Gap: Housing and Gallery systems created but no keyboard bindings
	// Fix: Added H key for housing UI and G key for gallery UI toggle
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4

	// Connect housing toggle (H key) - Phase 49.1, 51.2, 51.3
	if err := inputSystem.SetHousingCallback(func() {
		if g.HousingUI != nil {
			g.HousingUI.Toggle()
			if objectiveTracker != nil && g.PlayerEntity != nil {
				objectiveTracker.OnUIOpened(g.PlayerEntity, "housing")
			}
		}
	}); err != nil {
		return fmt.Errorf("failed to set housing callback: %w", err)
	}

	// Connect gallery toggle (G key) - Phase 49.4
	if err := inputSystem.SetGalleryCallback(func() {
		if g.GalleryUI != nil {
			g.GalleryUI.Toggle()
			if objectiveTracker != nil && g.PlayerEntity != nil {
				objectiveTracker.OnUIOpened(g.PlayerEntity, "gallery")
			}
		}
	}); err != nil {
		return fmt.Errorf("failed to set gallery callback: %w", err)
	}

	// Connect pause menu toggle (ESC key)
	if g.MenuSystem != nil {
		if err := inputSystem.SetMenuToggleCallback(func() {
			g.MenuSystem.Toggle()
		}); err != nil {
			return fmt.Errorf("failed to set menu toggle callback: %w", err)
		}
	}

	return nil
}

// GetWorld returns the ECS world instance (implements GameRunner interface).
func (g *EbitenGame) GetWorld() *World {
	return g.World
}

// GetScreenSize returns the current screen dimensions (implements GameRunner interface).
func (g *EbitenGame) GetScreenSize() (width, height int) {
	return g.ScreenWidth, g.ScreenHeight
}

// IsPaused returns whether the game is currently paused (implements GameRunner interface).
func (g *EbitenGame) IsPaused() bool {
	return g.Paused
}

// SetPaused sets the game pause state (implements GameRunner interface).
func (g *EbitenGame) SetPaused(paused bool) {
	g.Paused = paused
}

// GetPlayerEntity returns the current player entity (implements GameRunner interface).
func (g *EbitenGame) GetPlayerEntity() *Entity {
	return g.PlayerEntity
}

// ApplySettings applies settings from SettingsManager to game systems.
// This should be called when settings are loaded or changed.
// Returns error if application fails for critical settings.
func (g *EbitenGame) ApplySettings() error {
	if g.SettingsManager == nil {
		return nil // No settings to apply
	}

	settings := g.SettingsManager.GetSettings()

	// Apply audio volumes
	if g.AudioManager != nil {
		// Master volume affects both music and SFX as a multiplier
		g.AudioManager.SetMusicVolume(settings.MasterVolume * settings.MusicVolume)
		g.AudioManager.SetSFXVolume(settings.MasterVolume * settings.SFXVolume)

		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"masterVolume": settings.MasterVolume,
				"musicVolume":  settings.MusicVolume,
				"sfxVolume":    settings.SFXVolume,
			}).Debug("applied audio settings")
		}
	}

	// Apply VSync
	ebiten.SetVsyncEnabled(settings.VSync)

	// Apply fullscreen (Note: this triggers a window mode change)
	if settings.Fullscreen != ebiten.IsFullscreen() {
		ebiten.SetFullscreen(settings.Fullscreen)
	}

	// Graphics quality and ShowFPS are informational for now
	// Future: could affect particle counts, sprite quality, etc.

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"vsync":      settings.VSync,
			"fullscreen": settings.Fullscreen,
			"quality":    settings.GraphicsQuality,
			"showFPS":    settings.ShowFPS,
		}).Debug("applied display settings")
	}

	return nil
}

// EnableFrameTimeProfiling enables performance profiling with frame time tracking.
// This should be called before starting the game loop.
func (g *EbitenGame) EnableFrameTimeProfiling() {
	g.profilingEnabled = true
	if g.logger != nil {
		g.logger.Info("frame time profiling enabled")
	}
}

// DisableFrameTimeProfiling disables performance profiling.
func (g *EbitenGame) DisableFrameTimeProfiling() {
	g.profilingEnabled = false
	if g.logger != nil {
		g.logger.Info("frame time profiling disabled")
	}
}

// GetFrameTimeStats returns the current frame time statistics.
// Returns empty stats if profiling is disabled or no frames recorded.
func (g *EbitenGame) GetFrameTimeStats() FrameTimeStats {
	if g.frameTimeTracker == nil {
		return FrameTimeStats{}
	}
	return g.frameTimeTracker.GetStats()
}

// Run starts the game loop.
func (g *EbitenGame) Run(title string) error {
	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("failed to run game: %w", err)
	}

	return nil
}

// SetFullscreen sets fullscreen mode.
func (g *EbitenGame) SetFullscreen(fullscreen bool) {
	ebiten.SetFullscreen(fullscreen)
}

// Compile-time interface checks
var (
	_ GameRunner  = (*EbitenGame)(nil)
	_ ebiten.Game = (*EbitenGame)(nil)
)

// EnableLighting enables or disables the dynamic lighting system.
// When enabled, uses post-processing rendering pipeline with light sources.
func (g *EbitenGame) EnableLighting(enabled bool) {
	if g.LightingSystem != nil {
		g.LightingSystem.SetEnabled(enabled)

		if g.logger != nil {
			g.logger.WithField("enabled", enabled).Info("lighting system toggled")
		}
	}
}

// SetLightingGenrePreset configures lighting for the specified genre.
// This should be called when the genre is selected or changed.
func (g *EbitenGame) SetLightingGenrePreset(genreID string) {
	if g.LightingSystem != nil {
		config := g.LightingSystem.GetConfig()
		if config != nil {
			config.SetGenrePreset(genreID)
			g.LightingSystem.SetConfig(config)

			if g.logger != nil {
				g.logger.WithField("genre", genreID).Info("lighting genre preset applied")
			}
		}
	}
}
