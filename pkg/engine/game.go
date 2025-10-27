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
	StateManager      *AppStateManager
	MainMenuUI        *MainMenuUI
	CharacterCreation *EbitenCharacterCreation
	pendingCharData   *CharacterData
	isMultiplayerMode bool // Track if character creation is for multiplayer

	// Rendering systems
	CameraSystem        *CameraSystem
	RenderSystem        *EbitenRenderSystem
	TerrainRenderSystem *TerrainRenderSystem
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

	// Player entity reference (for UI systems)
	PlayerEntity *Entity

	// Callbacks for state transitions
	onNewGame            func() error
	onMultiplayerConnect func(serverAddr string) error
	onQuitToMenu         func() error

	// Logger for game operations
	logger *logrus.Entry
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

	game := &EbitenGame{
		World:             world,
		lastUpdateTime:    time.Now(),
		ScreenWidth:       screenWidth,
		ScreenHeight:      screenHeight,
		StateManager:      NewAppStateManager(),
		MainMenuUI:        NewMainMenuUI(screenWidth, screenHeight),
		CharacterCreation: NewCharacterCreation(screenWidth, screenHeight),
		CameraSystem:      cameraSystem,
		RenderSystem:      renderSystem,
		HUDSystem:         hudSystem,
		MenuSystem:        menuSystem,
		InventoryUI:       inventoryUI,
		QuestUI:           questUI,
		CharacterUI:       characterUI,
		SkillsUI:          skillsUI,
		MapUI:             mapUI,
		logger:            logEntry,
	}

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"screenWidth":  screenWidth,
			"screenHeight": screenHeight,
		}).Info("game initialized")
	}

	// Setup main menu callback
	game.MainMenuUI.SetSelectCallback(game.handleMainMenuSelection)

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

		if g.logger != nil {
			g.logger.Info("entering character creation for single-player")
		}

	case MainMenuOptionMultiPlayer:
		// Transition to character creation (same as single-player)
		if err := g.StateManager.TransitionTo(AppStateCharacterCreation); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("failed to transition to character creation")
			}
			return
		}

		// Reset character creation UI for multiplayer
		g.CharacterCreation.Reset()
		g.isMultiplayerMode = true // Multiplayer mode

		if g.logger != nil {
			g.logger.Info("entering character creation for multiplayer")
		}

	case MainMenuOptionSettings:
		// Future: transition to settings menu
		if g.logger != nil {
			g.logger.Info("settings menu not yet implemented")
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

// IsInMainMenu returns true if currently displaying the main menu.
func (g *EbitenGame) IsInMainMenu() bool {
	return g.StateManager.IsInMenu()
}

// Update implements ebiten.Game interface. Called every frame.
func (g *EbitenGame) Update() error {
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

	// Render terrain (if available)
	if g.TerrainRenderSystem != nil {
		g.TerrainRenderSystem.Draw(screen, g.CameraSystem)
	}

	// Render all entities
	g.RenderSystem.Draw(screen, g.World.GetEntities())

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
