// Package engine provides interaction system for Phase 11.2 and Phase 11.3.
// Puzzle Interaction System + Context Actions
//
// This file implements the InteractionSystem which handles player interactions:
// - Puzzle element interactions (Phase 11.2)
// - Context-sensitive actions (Phase 11.3): doors, levers, chests, NPCs
// - Carriable object pickup (Phase 11.3)
// All interactions use the F key (or touch input on mobile).
package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sirupsen/logrus"
)

// InteractionSystem manages player interactions with interactive entities.
type InteractionSystem struct {
	world          *World
	logger         *logrus.Entry
	carrySystem    *CarrySystem    // For pickup/throw mechanics
	inputSystem    *InputSystem    // INPUT CONFLICT FIX: Reference for checking if interaction allowed
	miniGameSystem *MiniGameSystem // For lock-picking and other mini-game integration
}

// NewInteractionSystem creates a new interaction system.
func NewInteractionSystem(world *World) *InteractionSystem {
	logger := world.GetLogger()
	return &InteractionSystem{
		world:  world,
		logger: logger,
	}
}

// SetCarrySystem sets the carry system reference for pickup mechanics.
func (s *InteractionSystem) SetCarrySystem(carrySystem *CarrySystem) {
	s.carrySystem = carrySystem
}

// SetInputSystem sets the input system reference for checking if interaction is allowed.
// INPUT CONFLICT FIX: Allows checking whether interactions are allowed based on current game state (e.g., UI open/closed) before processing.
func (s *InteractionSystem) SetInputSystem(inputSystem *InputSystem) {
	s.inputSystem = inputSystem
}

// SetMiniGameSystem sets the mini-game system reference for lock-picking and other mini-games.
func (s *InteractionSystem) SetMiniGameSystem(miniGameSystem *MiniGameSystem) {
	s.miniGameSystem = miniGameSystem
}

// Update checks for interaction key presses and processes player interactions.
// INPUT CONFLICT FIX: Now checks if interaction input is allowed before processing.
func (s *InteractionSystem) Update(entities []*Entity, deltaTime float64) {
	// Update context action cooldowns for all entities
	contextActions := s.world.GetEntitiesWith("contextAction")
	for _, entity := range contextActions {
		if comp, ok := entity.GetComponent("contextAction"); ok {
			if ctx, ok := comp.(*ContextActionComponent); ok {
				ctx.Update(deltaTime)
			}
		}
	}

	// INPUT CONFLICT FIX: Check if interaction input is allowed based on current game state
	if s.inputSystem != nil && !s.inputSystem.IsInteractionAllowed() {
		return
	}

	// Check if F key was just pressed
	// Note: Touch interaction is handled via virtual controls interact button (InputSystem callback)
	interactionPressed := inpututil.IsKeyJustPressed(ebiten.KeyF)

	if !interactionPressed {
		return
	}

	// Get player entity
	players := s.world.GetEntitiesWith("player")
	if len(players) == 0 {
		return
	}
	player := players[0]

	// Get player position
	playerPosComp, ok := player.GetComponent("position")
	if !ok {
		return
	}
	playerPos, ok := playerPosComp.(*PositionComponent)
	if !ok {
		return
	}

	// Priority order for interactions:
	// 1. Context actions (doors, levers, etc.)
	// 2. Carriable objects (pickup)
	// 3. Puzzle elements

	// Try context actions first
	if s.tryContextActions(player, playerPos) {
		return
	}

	// Try carriable object pickup
	if s.tryCarriablePickup(player, playerPos) {
		return
	}

	// Fall back to puzzle element interactions
	s.tryPuzzleInteraction(player, playerPos)
}

// tryContextActions attempts to interact with context action entities.
// Returns true if an interaction was performed.
func (s *InteractionSystem) tryContextActions(player *Entity, playerPos *PositionComponent) bool {
	// Find nearby context actions
	contextActions := s.world.GetEntitiesWith("contextAction")
	for _, entity := range contextActions {
		// Get entity position
		entPosComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entPos, ok := entPosComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Get context action component
		ctxComp, ok := entity.GetComponent("contextAction")
		if !ok {
			continue
		}
		contextAction, ok := ctxComp.(*ContextActionComponent)
		if !ok {
			continue
		}

		// Skip if not available or on cooldown
		if !contextAction.CanInteract() {
			continue
		}

		// Check distance to player
		dx := entPos.X - playerPos.X
		dy := entPos.Y - playerPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// If within interaction range, activate action
		if distance <= contextAction.InteractionRange {
			s.activateContextAction(player, entity, contextAction)
			return true
		}
	}

	return false
}

// activateContextAction handles context action activation.
func (s *InteractionSystem) activateContextAction(player, entity *Entity, contextAction *ContextActionComponent) {
	// Activate the action
	contextAction.Activate()

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":   entity.ID,
			"actionType": contextAction.ActionType.String(),
			"actionText": contextAction.ActionText,
		}).Debug("context action activated")
	}

	// Perform action-specific behavior
	switch contextAction.ActionType {
	case ActionOpen:
		s.handleOpenAction(player, entity)
	case ActionClose:
		s.handleCloseAction(entity)
	case ActionActivate:
		s.handleActivateAction(entity)
	case ActionRead:
		s.handleReadAction(player, entity)
	case ActionPlayGame:
		s.handlePlayGameAction(player, entity)
	case ActionInvestigate:
		s.handleInvestigateAction(player)
		// Other actions can be added here
	}
}

// handleOpenAction handles opening doors, chests, etc.
// Phase 27.3: Lock-picking mini-game integration
func (s *InteractionSystem) handleOpenAction(player, entity *Entity) {
	ctx := s.getContextAction(entity)
	if ctx == nil {
		return
	}

	if ctx.RequiresLockPicking {
		s.handleLockPickingOpen(player, entity, ctx)
	} else {
		s.handleNormalOpen(entity, ctx)
	}
}

// getContextAction retrieves and validates the context action component.
func (s *InteractionSystem) getContextAction(entity *Entity) *ContextActionComponent {
	ctxCompRaw, ok := entity.GetComponent("contextAction")
	if !ok {
		return nil
	}
	ctx, ok := ctxCompRaw.(*ContextActionComponent)
	if !ok {
		return nil
	}
	return ctx
}

// handleLockPickingOpen starts lock-picking mini-game or logs unavailability.
func (s *InteractionSystem) handleLockPickingOpen(player, entity *Entity, ctx *ContextActionComponent) {
	if s.miniGameSystem == nil {
		s.logLockPickingUnavailable(entity, ctx.LockDifficulty)
		return
	}

	difficulty := s.normalizeDifficulty(ctx.LockDifficulty)
	if err := s.startLockPickingGame(player, entity, difficulty); err != nil {
		s.logLockPickingError(player, err)
		return
	}

	s.storeLockPickingState(player, entity)
	s.logLockPickingStarted(player, entity, difficulty)
}

// normalizeDifficulty clamps difficulty to 0.0-1.0 range.
func (s *InteractionSystem) normalizeDifficulty(difficulty float64) float64 {
	if difficulty < 0.0 {
		return 0.0
	}
	if difficulty > 1.0 {
		return 1.0
	}
	return difficulty
}

// startLockPickingGame initiates the lock-picking mini-game.
func (s *InteractionSystem) startLockPickingGame(player, entity *Entity, difficulty float64) error {
	return s.miniGameSystem.StartGame(player.ID, MiniGameLockPicking, difficulty)
}

// storeLockPickingState stores the locked entity ID for completion callback.
func (s *InteractionSystem) storeLockPickingState(player, entity *Entity) {
	gameComp := s.miniGameSystem.GetGameComponent(player.ID)
	if gameComp == nil {
		return
	}
	stateMap := gameComp.State.(map[string]interface{})
	stateMap["lockedEntityID"] = entity.ID
}

// handleNormalOpen opens a door without lock-picking.
func (s *InteractionSystem) handleNormalOpen(entity *Entity, ctx *ContextActionComponent) {
	ctx.ActionType = ActionClose
	ctx.ActionText = "Close"
	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("door opened")
	}
}

// logLockPickingError logs lock-picking start error.
func (s *InteractionSystem) logLockPickingError(player *Entity, err error) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID": player.ID,
		"error":    err.Error(),
	}).Error("failed to start lock-picking mini-game")
}

// logLockPickingUnavailable logs when lock-picking system is not available.
func (s *InteractionSystem) logLockPickingUnavailable(entity *Entity, difficulty float64) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":       entity.ID,
		"lockDifficulty": difficulty,
	}).Debug("lock-picking mini-game required but system not available")
}

// logLockPickingStarted logs successful lock-picking game start.
func (s *InteractionSystem) logLockPickingStarted(player, entity *Entity, difficulty float64) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":       entity.ID,
		"playerID":       player.ID,
		"lockDifficulty": difficulty,
	}).Debug("lock-picking mini-game started")
}

// ProcessLockPickingCompletion checks if a lock-picking mini-game has completed
// and handles the result by opening the locked entity on success.
// This should be called by the game loop or mini-game system after game completion.
func (s *InteractionSystem) ProcessLockPickingCompletion(playerID uint64, success bool) {
	lockedEntityID, ctx := s.getLockPickingTarget(playerID)
	if ctx == nil {
		return
	}

	if success {
		s.unlockEntity(ctx, playerID, lockedEntityID)
	} else {
		s.logLockPickingFailure(playerID, lockedEntityID)
	}
}

// getLockPickingTarget retrieves the locked entity and its context action component from the mini-game state.
func (s *InteractionSystem) getLockPickingTarget(playerID uint64) (uint64, *ContextActionComponent) {
	gameComp := s.validateMiniGame(playerID)
	if gameComp == nil {
		return 0, nil
	}

	lockedEntityID := s.extractLockedEntityID(gameComp)
	if lockedEntityID == 0 {
		return 0, nil
	}

	ctx := s.getContextActionByEntityID(lockedEntityID)
	return lockedEntityID, ctx
}

// validateMiniGame checks if the mini-game system and component are valid for lock-picking.
func (s *InteractionSystem) validateMiniGame(playerID uint64) *MiniGameComponent {
	if s.miniGameSystem == nil {
		return nil
	}

	gameComp := s.miniGameSystem.GetGameComponent(playerID)
	if gameComp == nil || gameComp.GameType != MiniGameLockPicking {
		return nil
	}

	return gameComp
}

// extractLockedEntityID retrieves the locked entity ID from the mini-game state.
func (s *InteractionSystem) extractLockedEntityID(gameComp *MiniGameComponent) uint64 {
	stateMap, ok := gameComp.State.(map[string]interface{})
	if !ok {
		return 0
	}

	lockedEntityIDRaw, ok := stateMap["lockedEntityID"]
	if !ok {
		return 0
	}

	lockedEntityID, ok := lockedEntityIDRaw.(uint64)
	if !ok {
		return 0
	}

	return lockedEntityID
}

// getContextAction retrieves the context action component for a locked entity.
// getContextActionByEntityID retrieves the context action component for a locked entity.
func (s *InteractionSystem) getContextActionByEntityID(lockedEntityID uint64) *ContextActionComponent {
	lockedEntity, exists := s.world.GetEntity(lockedEntityID)
	if !exists || lockedEntity == nil {
		return nil
	}

	ctxCompRaw, ok := lockedEntity.GetComponent("contextAction")
	if !ok {
		return nil
	}

	ctx, ok := ctxCompRaw.(*ContextActionComponent)
	if !ok {
		return nil
	}

	return ctx
}

// unlockEntity updates the context action to reflect successful lock-picking.
func (s *InteractionSystem) unlockEntity(ctx *ContextActionComponent, playerID, lockedEntityID uint64) {
	ctx.ActionType = ActionClose
	ctx.ActionText = "Close"
	ctx.RequiresLockPicking = false

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":       playerID,
			"lockedEntityID": lockedEntityID,
		}).Info("lock picked successfully - door/chest opened")
	}
}

// logLockPickingFailure logs a failed lock-picking attempt.
func (s *InteractionSystem) logLockPickingFailure(playerID, lockedEntityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":       playerID,
			"lockedEntityID": lockedEntityID,
		}).Debug("lock-picking failed - door/chest remains locked")
	}
}

// handleCloseAction handles closing doors.
func (s *InteractionSystem) handleCloseAction(entity *Entity) {
	// Update action text to "Open"
	if ctxComp, ok := entity.GetComponent("contextAction"); ok {
		if ctx, ok := ctxComp.(*ContextActionComponent); ok {
			ctx.ActionType = ActionOpen
			ctx.ActionText = "Open"
		}
	}

	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("door closed")
	}
}

// handleActivateAction handles lever/switch activation.
func (s *InteractionSystem) handleActivateAction(entity *Entity) {
	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("lever activated")
	}
	// Can trigger puzzle state changes, open gates, etc.
}

// tryCarriablePickup attempts to pick up a carriable object.
// Returns true if a pickup was performed.
func (s *InteractionSystem) tryCarriablePickup(player *Entity, playerPos *PositionComponent) bool {
	if s.carrySystem == nil {
		return false
	}

	// Check if player is already carrying something
	if s.carrySystem.IsCarrying(player.ID) {
		// Drop the carried object
		s.carrySystem.DropObject(player.ID)
		return true
	}

	// Find nearby carriable object
	maxDistance := 48.0 // 1.5 tiles
	objectID, distance := s.carrySystem.FindNearbyCarriableObject(playerPos.X, playerPos.Y, maxDistance)

	if objectID != 0 {
		// Try to pick up the object
		if s.carrySystem.TryPickup(player.ID, objectID) {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"playerID": player.ID,
					"objectID": objectID,
					"distance": distance,
				}).Debug("object picked up")
			}
			return true
		}
	}

	return false
}

// tryPuzzleInteraction attempts to interact with puzzle elements.
func (s *InteractionSystem) tryPuzzleInteraction(player *Entity, playerPos *PositionComponent) {
	// Find nearby puzzle elements
	puzzleElements := s.world.GetEntitiesWith("puzzleElement")

	for _, element := range puzzleElements {
		// Get element position
		elemPosComp, ok := element.GetComponent("position")
		if !ok {
			continue
		}
		elemPos, ok := elemPosComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Get puzzle element component
		elemComp, ok := element.GetComponent("puzzleElement")
		if !ok {
			continue
		}
		puzzleElem, ok := elemComp.(*PuzzleElementComponent)
		if !ok {
			continue
		}

		// Skip if not interactable or on cooldown
		if !puzzleElem.IsInteractable || puzzleElem.CooldownElapsed > 0 {
			continue
		}

		// Check distance to player
		dx := elemPos.X - playerPos.X
		dy := elemPos.Y - playerPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// If within interaction range, activate element
		if distance <= puzzleElem.InteractionRange {
			s.activatePuzzleElement(player, element, puzzleElem)
			return
		}
	}
}

// activatePuzzleElement handles the interaction with a puzzle element.
func (s *InteractionSystem) activatePuzzleElement(player, element *Entity, puzzleElem *PuzzleElementComponent) {
	// Toggle element state
	puzzleElem.State = (puzzleElem.State + 1) % 2 // Toggle between 0 and 1

	// Set cooldown
	puzzleElem.CooldownElapsed = puzzleElem.CooldownTime

	// Record activation
	puzzleElem.ActivatedBy = player.ID

	// Find parent puzzle and record progress
	puzzles := s.world.GetEntitiesWith("puzzle")
	for _, puzzleEntity := range puzzles {
		puzzleComp, ok := puzzleEntity.GetComponent("puzzle")
		if !ok {
			continue
		}

		puzzle, ok := puzzleComp.(*PuzzleComponent)
		if !ok {
			continue
		}
		if puzzle.PuzzleID == puzzleElem.PuzzleID {
			// Record progress
			puzzle.RecordProgress(puzzleElem.ElementName)

			// Start solving if not yet started
			if puzzle.State == PuzzleStateUnsolved {
				puzzle.State = PuzzleStateSolving
			}

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"puzzleID":    puzzle.PuzzleID,
					"elementName": puzzleElem.ElementName,
					"elementType": puzzleElem.ElementType,
					"newState":    puzzleElem.State,
					"progress":    len(puzzle.CurrentProgress),
					"solution":    len(puzzle.Solution),
				}).Debug("puzzle element activated")
			}

			break
		}
	}
}

// handleReadAction handles reading books from bookshelves or direct book entities.
func (s *InteractionSystem) handleReadAction(player, entity *Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": player.ID,
			"entityID": entity.ID,
		}).Debug("read action initiated")
	}

	// Check if entity is a bookshelf
	if bookshelfComp, ok := entity.GetComponent("bookshelf"); ok {
		if bookshelf, ok := bookshelfComp.(*BookshelfComponent); ok {
			s.handleBookshelfRead(player, entity, bookshelf)
			return
		}
	}

	// Check if entity is a book directly
	if bookComp, ok := entity.GetComponent("book"); ok {
		if _, ok := bookComp.(*BookComponent); ok {
			// For direct book reading, we would trigger a UI to display the book
			// This would be handled by a BookReadingUI system
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"playerID": player.ID,
					"bookID":   entity.ID,
				}).Info("player interacting with book")
			}
			return
		}
	}
}

// handleBookshelfRead handles reading books from a bookshelf.
// INTEGRATION FIX [Category F] Gap F3: Bookshelf Key Requirement
// Locked bookshelves now require the player to have the correct key item in their inventory.
func (s *InteractionSystem) handleBookshelfRead(player, bookshelfEntity *Entity, bookshelf *BookshelfComponent) {
	// Check if bookshelf is locked
	if bookshelf.IsLocked {
		// Check if bookshelf requires a key
		if bookshelf.RequiresKey && bookshelf.KeyItemID != "" {
			// Check if player has the required key
			if !s.playerHasItemByID(player, bookshelf.KeyItemID) {
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"bookshelfID": bookshelfEntity.ID,
						"playerID":    player.ID,
						"requiredKey": bookshelf.KeyItemID,
					}).Info("player attempted to access locked bookshelf without key")
				}
				// Player doesn't have the key - bookshelf remains locked
				return
			}

			// Player has the key - unlock the bookshelf
			bookshelf.IsLocked = false
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"bookshelfID": bookshelfEntity.ID,
					"playerID":    player.ID,
					"keyItemID":   bookshelf.KeyItemID,
				}).Info("player unlocked bookshelf with key")
			}
		} else {
			// Bookshelf is locked but no key is required (should not happen in normal gameplay)
			if s.logger != nil {
				s.logger.WithField("bookshelfID", bookshelfEntity.ID).Debug("bookshelf is locked but no key configured")
			}
			return
		}
	}

	// Check if bookshelf is empty
	if bookshelf.IsEmpty() {
		if s.logger != nil {
			s.logger.WithField("bookshelfID", bookshelfEntity.ID).Debug("bookshelf is empty")
		}
		return
	}

	// For now, just log that the player can browse the bookshelf
	// A full implementation would open a BookshelfUI showing available books
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":    player.ID,
			"bookshelfID": bookshelfEntity.ID,
			"bookCount":   bookshelf.GetBookCount(),
		}).Info("player browsing bookshelf")
	}
}

// playerHasItemByID checks if the player has an item with the specified ID in their inventory.
// Returns true if the item is found, false otherwise.
func (s *InteractionSystem) playerHasItemByID(player *Entity, itemID string) bool {
	// Get player's inventory component
	invCompRaw, ok := player.GetComponent("inventory")
	if !ok || invCompRaw == nil {
		return false
	}

	invComp, ok := invCompRaw.(*InventoryComponent)
	if !ok || invComp == nil {
		return false
	}

	// Search through inventory items for matching ID
	for _, itm := range invComp.Items {
		if itm != nil && itm.ID == itemID {
			return true
		}
	}

	return false
}

// handlePlayGameAction handles interaction with mini-game stations (Phase 27.3).
// Validates entry cost/requirements and starts the mini-game if allowed.
func (s *InteractionSystem) handlePlayGameAction(player, stationEntity *Entity) {
	s.logPlayGameStart(player.ID, stationEntity.ID)

	station, ok := s.getMiniGameStation(stationEntity)
	if !ok {
		return
	}

	if !s.validateStationAvailable(station, stationEntity.ID) {
		return
	}

	playerLevel, playerGold := s.getPlayerResources(player)

	if !s.validatePlayerRequirements(station, playerLevel, playerGold, stationEntity.ID) {
		return
	}

	s.deductEntryCost(player, station.EntryCost)
	station.StartGame(player.ID)
	s.logGameStarted(player.ID, stationEntity.ID, station)
}

// getMiniGameStation retrieves and validates the mini-game station component.
func (s *InteractionSystem) getMiniGameStation(stationEntity *Entity) (*MiniGameStationComponent, bool) {
	stationCompRaw, ok := stationEntity.GetComponent("minigameStation")
	if !ok {
		if s.logger != nil {
			s.logger.WithField("stationID", stationEntity.ID).Warn("entity missing minigameStation component")
		}
		return nil, false
	}

	station, ok := stationCompRaw.(*MiniGameStationComponent)
	return station, ok
}

// validateStationAvailable checks if the station is currently available.
func (s *InteractionSystem) validateStationAvailable(station *MiniGameStationComponent, stationID uint64) bool {
	if station.IsOccupied {
		if s.logger != nil {
			s.logger.WithField("stationID", stationID).Debug("station is occupied")
		}
		return false
	}
	return true
}

// getPlayerResources extracts player level and gold from components.
func (s *InteractionSystem) getPlayerResources(player *Entity) (int, int) {
	playerLevel := 1
	playerGold := 0

	if expComp, ok := player.GetComponent("experience"); ok {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			playerLevel = exp.Level
		}
	}

	if invComp, ok := player.GetComponent("inventory"); ok {
		if inv, ok := invComp.(*InventoryComponent); ok {
			playerGold = inv.Gold
		}
	}

	return playerLevel, playerGold
}

// validatePlayerRequirements checks if player meets station requirements.
func (s *InteractionSystem) validatePlayerRequirements(station *MiniGameStationComponent, playerLevel, playerGold int, stationID uint64) bool {
	if !station.CanPlay(playerLevel, playerGold) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"playerLevel":   playerLevel,
				"playerGold":    playerGold,
				"requiredLevel": station.RequiresLevel,
				"entryCost":     station.EntryCost,
			}).Debug("player does not meet station requirements")
		}
		return false
	}
	return true
}

// deductEntryCost deducts the entry cost from player's gold.
func (s *InteractionSystem) deductEntryCost(player *Entity, entryCost int) {
	if entryCost > 0 {
		if invComp, ok := player.GetComponent("inventory"); ok {
			if inv, ok := invComp.(*InventoryComponent); ok {
				inv.Gold -= entryCost
			}
		}
	}
}

// logPlayGameStart logs the initiation of play game action.
func (s *InteractionSystem) logPlayGameStart(playerID, stationID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":  playerID,
			"stationID": stationID,
		}).Debug("play game action initiated")
	}
}

// logGameStarted logs successful game start with details.
func (s *InteractionSystem) logGameStarted(playerID, stationID uint64, station *MiniGameStationComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":   playerID,
			"stationID":  stationID,
			"gameType":   station.GameType.String(),
			"difficulty": station.Difficulty,
			"entryCost":  station.EntryCost,
		}).Info("player started mini-game")
	}
}

// handleInvestigateAction initiates environmental investigation (Phase 30.2).
// Starts an investigation action to reveal hidden story fragments and clues.
func (s *InteractionSystem) handleInvestigateAction(player *Entity) {
	if s.logger != nil {
		s.logger.WithField("playerID", player.ID).Debug("investigate action initiated")
	}

	// Get investigation component
	invComp, ok := player.GetComponent("investigation")
	if !ok {
		if s.logger != nil {
			s.logger.WithField("playerID", player.ID).Debug("player has no investigation component")
		}
		return
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		return
	}

	// Try to start investigation
	if !investigation.StartInvestigation() {
		if s.logger != nil {
			s.logger.WithField("playerID", player.ID).Debug("investigation on cooldown")
		}
		return
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": player.ID,
			"radius":   investigation.GetEffectiveRadius(),
			"duration": investigation.InvestigationDuration,
		}).Info("player started investigation")
	}

	// The actual investigation processing is handled by InvestigationSystem.Update()
}
