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
	world       *World
	logger      *logrus.Entry
	carrySystem *CarrySystem // For pickup/throw mechanics
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

// Update checks for interaction key presses and processes player interactions.
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

	// Check if F key was just pressed (or touch input on mobile)
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
		s.handleOpenAction(entity)
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
// Phase 27.3: Added lock-picking mini-game requirement check
func (s *InteractionSystem) handleOpenAction(entity *Entity) {
	// Get context action to check for lock-picking requirement
	ctxCompRaw, ok := entity.GetComponent("contextAction")
	if !ok {
		return
	}
	ctx, ok := ctxCompRaw.(*ContextActionComponent)
	if !ok {
		return
	}

	// Check if lock-picking is required
	if ctx.RequiresLockPicking {
		// INTEGRATION FIX [Category F]: Lock-Picking Mini-Game Integration
		// Gap: ActionOpenLocked requires MiniGameSystem.StartGame() integration for lock-picking
		// Fix: Add MiniGameSystem field to InteractionSystem, call StartGame(TypeLockPicking, difficulty)
		// Roadmap: ROADMAP_V4.md Phase 27.3 - Mini-Game Integration (Quest mini-games)
		// Integration: MiniGameSystem available in world.GetSystems(), get via type assertion
		// For now, just log that lock-picking would be required
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entityID":       entity.ID,
				"lockDifficulty": ctx.LockDifficulty,
			}).Debug("lock-picking mini-game required (not implemented yet)")
		}
		// In full implementation, would call:
		// s.world.GetSystem("minigame").StartGame(playerID, MiniGameLockPicking, ctx.LockDifficulty)
		// and only proceed to open if mini-game succeeds
		return
	}

	// Normal door opening (no lock-picking required)
	ctx.ActionType = ActionClose
	ctx.ActionText = "Close"

	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("door opened")
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
func (s *InteractionSystem) handleBookshelfRead(player, bookshelfEntity *Entity, bookshelf *BookshelfComponent) {
	// Check if bookshelf is locked
	if bookshelf.IsLocked {
		if s.logger != nil {
			s.logger.WithField("bookshelfID", bookshelfEntity.ID).Debug("bookshelf is locked")
		}
		// INTEGRATION FIX [Category F]: Bookshelf Key Requirement
		// Gap: Locked bookshelf interaction needs inventory check for key items
		// Fix: Query InventoryComponent for item matching bookshelf.RequiredKeyID, reject if missing
		// Roadmap: ROADMAP_V4.md Phase 23.2 - Lore Integration (bookshelf interaction)
		// Integration: Add HasItem(requiredKeyID) check before showing bookshelf contents
		return
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
