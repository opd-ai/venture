// Package engine provides the main game loop and Ebiten integration.
// This file implements EbitenGame which ties together the ECS world, rendering
// systems, and the Ebiten game engine. EbitenGame implements both ebiten.Game
// and GameRunner interfaces.
package engine

import (
	"fmt"
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
		// Log error but continue (save/load won't work but game can run)
		if logEntry != nil {
			logEntry.WithError(err).Warn("failed to initialize menu system")
		}
		// Note: No fallback logging when logEntry is nil - silent initialization failure
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

<<<<<<< HEAD
	// Initialize frame time tracker (disabled by default, enabled via EnableFrameTimeProfiling)
	frameTimeTracker := NewFrameTimeTracker(1000) // Track last 1000 frames (~16 seconds at 60 FPS)
=======
	// Create lighting system with default configuration
	// Note: Will be enabled via command-line flag in client/main.go
	lightingConfig := NewLightingConfig()
	lightingConfig.Enabled = false // Disabled by default, enable via flag
	lightingSystem := NewLightingSystemWithLogger(world, lightingConfig, logger)

	// Create reusable scene buffer for lighting post-processing
	// Allocated once to avoid per-frame allocations (60+ FPS)
	sceneBuffer := ebiten.NewImage(screenWidth, screenHeight)
>>>>>>> 8228407969cbadf869d831fa554f2b04033bd6de

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

// handleSinglePlayerMenuLoadGame handles the Load Game selection (Phase 8.3).
func (g *EbitenGame) handleSinglePlayerMenuLoadGame() {
	// TODO: Phase 8.3 - Implement save/load system
	if g.logger != nil {
		g.logger.Info("load game selected (not yet implemented)")
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
		g.logger.Info("host game selected - starting local server")
	}

	// TODO: Start local server using pkg/hostplay
	// For now, automatically connect to localhost:8080
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
func (g *EbitenGame) Update() error {
	// Track frame time for performance monitoring
	frameStart := time.Now()
	defer func() {
		if g.profilingEnabled && g.frameTimeTracker != nil {
			g.frameTimeTracker.RecordFrame(time.Since(frameStart))
			g.frameCount++

			// Log stats every 300 frames (5 seconds at 60 FPS)
			if g.frameCount%300 == 0 && g.logger != nil {
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

				// Warn if stuttering detected
				if stats.IsStuttering() {
					fields["stuttering"] = true
					g.logger.WithFields(fields).Warn("frame time stuttering detected")
				} else {
					g.logger.WithFields(fields).Info("frame time stats")
				}
			}
		}
	}()

	// Calculate delta time
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Cap delta time to prevent spiral of death
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	// If in main menu state, only update main menu
	if g.StateManager.CurrentState() == AppStateMainMenu {
		g.MainMenuUI.Update()
		return nil
	}

	// If in single-player menu state, only update single-player menu
	if g.StateManager.CurrentState() == AppStateSinglePlayerMenu {
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Update()
		}
		return nil
	}

	// If in genre selection state, only update genre selection menu
	if g.StateManager.CurrentState() == AppStateGenreSelection {
		if g.GenreSelectionMenu != nil {
			g.GenreSelectionMenu.Update()
		}
		return nil
	}

	// If in multiplayer menu state, only update multiplayer menu
	if g.StateManager.CurrentState() == AppStateMultiPlayerMenu {
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Update()
		}
		return nil
	}

	// If in server address input state, only update server address input
	if g.StateManager.CurrentState() == AppStateServerAddressInput {
		if g.ServerAddressInput != nil {
			g.ServerAddressInput.Update()
		}
		return nil
	}

	// If in settings state, only update settings
	if g.StateManager.CurrentState() == AppStateSettings {
		g.SettingsUI.Update()
		return nil
	}

	// If in character creation state, update character creation
	if g.StateManager.CurrentState() == AppStateCharacterCreation {
		completed := g.CharacterCreation.Update()
		if completed {
			// Character creation finished, store data and transition to gameplay
			charData := g.CharacterCreation.GetCharacterData()
			g.pendingCharData = &charData

			if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
				if g.logger != nil {
					g.logger.WithError(err).Error("failed to transition to gameplay after character creation")
				}
				return err
			}

			// Trigger appropriate callback based on mode
			if g.isMultiplayerMode {
				// Multiplayer: connect to server with character data
				if g.onMultiplayerConnect != nil {
					if err := g.onMultiplayerConnect(""); err != nil {
						if g.logger != nil {
							g.logger.WithError(err).Error("multiplayer connect callback failed")
						}
						// Transition back to menu on error
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
			} else {
				// Single-player: start new game
				if g.onNewGame != nil {
					if err := g.onNewGame(); err != nil {
						if g.logger != nil {
							g.logger.WithError(err).Error("new game callback failed")
						}
						// Transition back to menu on error
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
			}
		}
		return nil
	}

	// If in any other menu state, only update menu
	if g.StateManager.IsInMenu() {
		g.MainMenuUI.Update()
		return nil
	}

	// From here on, we're in gameplay state

	// If menu is visible, pause game world (but allow menu input)
	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.Paused = true
		// Update menu even when paused
		g.MenuSystem.Update(g.World.GetEntities(), deltaTime)
		return nil
	}

	if g.Paused {
		return nil
	}

	// Update UI systems first (they capture input if visible)
	g.InventoryUI.Update(nil, deltaTime)
	g.QuestUI.Update(nil, deltaTime)
	g.CharacterUI.Update(nil, deltaTime)
	g.SkillsUI.Update(nil, deltaTime)
	g.MapUI.Update(nil, deltaTime)

	// Update shop UI (if initialized)
	if g.ShopUI != nil {
		g.ShopUI.Update(g.World.GetEntities(), deltaTime)
	}

	// Update crafting UI (if initialized)
	if g.CraftingUI != nil {
		g.CraftingUI.Update(nil, deltaTime)
	}

	// Gap #6: Always update tutorial system for progress tracking (even when UI visible)
	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Update(g.World.GetEntities(), deltaTime)
	}

	// Update the world (unless UI is blocking input)
	if !g.InventoryUI.IsVisible() && !g.QuestUI.IsVisible() && !g.CharacterUI.IsVisible() && !g.SkillsUI.IsVisible() && !g.MapUI.IsFullScreen() && (g.ShopUI == nil || !g.ShopUI.IsVisible()) && (g.CraftingUI == nil || !g.CraftingUI.IsVisible()) {
		g.World.Update(deltaTime)
	}

	// Update camera system
	g.CameraSystem.Update(g.World.GetEntities(), deltaTime)

	return nil
}

// Draw implements ebiten.Game interface. Called every frame.
func (g *EbitenGame) Draw(screen *ebiten.Image) {
	// If in main menu state, only draw main menu
	if g.StateManager.CurrentState() == AppStateMainMenu {
		g.MainMenuUI.Draw(screen)
		return
	}

	// If in single-player menu state, only draw single-player menu
	if g.StateManager.CurrentState() == AppStateSinglePlayerMenu {
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Draw(screen)
		}
		return
	}

	// If in genre selection state, only draw genre selection menu
	if g.StateManager.CurrentState() == AppStateGenreSelection {
		if g.GenreSelectionMenu != nil {
			g.GenreSelectionMenu.Draw(screen)
		}
		return
	}

	// If in multiplayer menu state, only draw multiplayer menu
	if g.StateManager.CurrentState() == AppStateMultiPlayerMenu {
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Draw(screen)
		}
		return
	}

	// If in server address input state, only draw server address input
	if g.StateManager.CurrentState() == AppStateServerAddressInput {
		if g.ServerAddressInput != nil {
			g.ServerAddressInput.Draw(screen)
		}
		return
	}

	// If in settings state, only draw settings
	if g.StateManager.CurrentState() == AppStateSettings {
		g.SettingsUI.Draw(screen)
		return
	}

	// If in character creation state, only draw character creation
	if g.StateManager.CurrentState() == AppStateCharacterCreation {
		g.CharacterCreation.Draw(screen)
		return
	}

	// If in any other menu state, draw main menu
	if g.StateManager.IsInMenu() {
		g.MainMenuUI.Draw(screen)
		return
	}

	// From here on, we're in gameplay state and render the full game

	// If lighting is enabled, use post-processing pipeline
	if g.LightingSystem != nil && g.LightingSystem.IsEnabled() {
		// Clear and reuse scene buffer (avoid per-frame allocation)
		g.sceneBuffer.Clear()

		// Render terrain to buffer (if available)
		if g.TerrainRenderSystem != nil {
			g.TerrainRenderSystem.Draw(g.sceneBuffer, g.CameraSystem)
		}

		// Render all entities to buffer
		g.RenderSystem.Draw(g.sceneBuffer, g.World.GetEntities())

		// Update lighting system viewport based on camera
		if g.CameraSystem != nil {
			camX, camY := g.CameraSystem.GetPosition()
			g.LightingSystem.SetViewport(camX, camY, g.ScreenWidth, g.ScreenHeight)
		}

		// Apply lighting as post-processing (renders sceneBuffer with lighting to screen)
		entities := g.World.GetEntities()
		g.LightingSystem.ApplyLighting(screen, g.sceneBuffer, entities)
	} else {
		// Standard rendering pipeline (no lighting)
		// Render terrain (if available)
		if g.TerrainRenderSystem != nil {
			g.TerrainRenderSystem.Draw(screen, g.CameraSystem)
		}

		// Render all entities
		g.RenderSystem.Draw(screen, g.World.GetEntities())
	}

	// Render HUD overlay
	g.HUDSystem.Draw(screen)

	// Render tutorial overlay (if active)
	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Draw(screen)
	}

	// Render help overlay (if visible)
	if g.HelpSystem != nil && g.HelpSystem.Visible {
		g.HelpSystem.Draw(screen)
	}

	// Render menu overlay (if active)
	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.MenuSystem.Draw(screen)
	}

	// Render UI overlays (drawn last so they're on top)
	g.InventoryUI.Draw(screen)
	g.QuestUI.Draw(screen)
	g.CharacterUI.Draw(screen)
	g.SkillsUI.Draw(screen)
	g.MapUI.Draw(screen) // Map UI draws last to be on top of everything

	// Render shop UI (if initialized)
	if g.ShopUI != nil {
		g.ShopUI.Draw(screen)
	}

	// Render crafting UI (if initialized)
	if g.CraftingUI != nil {
		g.CraftingUI.Draw(screen)
	}

	// Render virtual controls (mobile only, drawn last to be on top of everything)
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
func (g *EbitenGame) SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) {
	// Connect inventory toggle
	inputSystem.SetInventoryCallback(func() {
		g.InventoryUI.Toggle()
		// GAP-014 REPAIR: Track inventory UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "inventory")
		}
	})

	// Connect quest log toggle
	inputSystem.SetQuestsCallback(func() {
		g.QuestUI.Toggle()
		// GAP-014 REPAIR: Track quest log UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "quest_log")
		}
	})

	// Connect character screen toggle
	inputSystem.SetCharacterCallback(func() {
		g.CharacterUI.Toggle()
		// GAP-014 REPAIR: Track character UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "character")
		}
	})

	// Connect skills screen toggle
	inputSystem.SetSkillsCallback(func() {
		g.SkillsUI.Toggle()
		// GAP-014 REPAIR: Track skills UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "skills")
		}
	})

	// Connect map toggle
	inputSystem.SetMapCallback(func() {
		g.MapUI.ToggleFullScreen()
		// GAP-014 REPAIR: Track map UI opens for tutorial objectives
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, "map")
		}
	})

	// Connect crafting toggle (Category 1.3 - Commerce & Crafting Integration)
	inputSystem.SetCraftingCallback(func() {
		if g.CraftingUI != nil {
			g.CraftingUI.Toggle()
			// Track crafting UI opens for tutorial objectives
			if objectiveTracker != nil && g.PlayerEntity != nil {
				objectiveTracker.OnUIOpened(g.PlayerEntity, "crafting")
			}
		}
	})

	// Connect pause menu toggle (ESC key)
	if g.MenuSystem != nil {
		inputSystem.SetMenuToggleCallback(func() {
			g.MenuSystem.Toggle()
		})
	}
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
