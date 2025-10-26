// Package engine provides the dialog interaction system.
// This file implements DialogSystem which manages NPC conversations,
// dialog state, and player interaction with NPCs. The system supports
// extensible dialog providers for future branching conversations,
// quest dialogs, and dynamic content.
package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// DialogSystem manages NPC dialog interactions and state.
type DialogSystem struct {
	world              *World
	logger             *logrus.Entry
	activeDialogEntity uint64 // Entity currently engaged in dialog (0 = none)
}

// NewDialogSystem creates a new dialog system.
func NewDialogSystem(world *World) *DialogSystem {
	return NewDialogSystemWithLogger(world, nil)
}

// NewDialogSystemWithLogger creates a new dialog system with a logger.
func NewDialogSystemWithLogger(world *World, logger *logrus.Logger) *DialogSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "dialog")
	}
	return &DialogSystem{
		world:              world,
		logger:             logEntry,
		activeDialogEntity: 0,
	}
}

// StartDialog initiates a dialog interaction with an NPC.
// Returns true if dialog started successfully.
func (s *DialogSystem) StartDialog(playerID, npcID uint64) (bool, error) {
	// Check if already in dialog
	if s.activeDialogEntity != 0 {
		return false, fmt.Errorf("already in dialog with entity %d", s.activeDialogEntity)
	}

	// Verify NPC exists
	npcEntity, ok := s.world.GetEntity(npcID)
	if !ok {
		return false, fmt.Errorf("NPC entity %d not found", npcID)
	}

	// Get dialog component
	comp, ok := npcEntity.GetComponent("dialog")
	if !ok {
		return false, fmt.Errorf("entity %d does not have dialog component", npcID)
	}
	dialogComp, ok := comp.(*DialogComponent)
	if !ok {
		return false, fmt.Errorf("entity %d dialog component has wrong type", npcID)
	}

	// Activate dialog
	dialogComp.Activate()
	s.activeDialogEntity = npcID

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"npcID":    npcID,
		}).Info("dialog started")
	}

	return true, nil
}

// EndDialog closes the current dialog interaction.
func (s *DialogSystem) EndDialog() error {
	if s.activeDialogEntity == 0 {
		return fmt.Errorf("no active dialog to end")
	}

	// Get dialog component
	npcEntity, ok := s.world.GetEntity(s.activeDialogEntity)
	if !ok {
		// Entity was destroyed, just clear active dialog
		s.activeDialogEntity = 0
		return nil
	}

	comp, ok := npcEntity.GetComponent("dialog")
	if ok {
		if dialogComp, ok := comp.(*DialogComponent); ok {
			dialogComp.Deactivate()
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"npcID": s.activeDialogEntity,
		}).Info("dialog ended")
	}

	s.activeDialogEntity = 0
	return nil
}

// SelectDialogOption processes the player's choice in a dialog.
// Returns the selected action and any error.
func (s *DialogSystem) SelectDialogOption(optionIndex int) (DialogAction, error) {
	if s.activeDialogEntity == 0 {
		return ActionNone, fmt.Errorf("no active dialog")
	}

	// Get dialog component
	npcEntity, ok := s.world.GetEntity(s.activeDialogEntity)
	if !ok {
		s.activeDialogEntity = 0
		return ActionNone, fmt.Errorf("NPC entity %d not found", s.activeDialogEntity)
	}

	comp, ok := npcEntity.GetComponent("dialog")
	if !ok {
		return ActionNone, fmt.Errorf("entity %d does not have dialog component", s.activeDialogEntity)
	}
	dialogComp, ok := comp.(*DialogComponent)
	if !ok {
		return ActionNone, fmt.Errorf("entity %d dialog component has wrong type", s.activeDialogEntity)
	}

	// Validate option index
	if optionIndex < 0 || optionIndex >= len(dialogComp.Options) {
		return ActionNone, fmt.Errorf("invalid option index %d (available: 0-%d)", optionIndex, len(dialogComp.Options)-1)
	}

	option := dialogComp.Options[optionIndex]

	// Check if option is enabled
	if !option.Enabled {
		return ActionNone, fmt.Errorf("option %d is disabled", optionIndex)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"npcID":       s.activeDialogEntity,
			"optionIndex": optionIndex,
			"action":      option.Action.String(),
		}).Debug("dialog option selected")
	}

	// Close dialog if action requires it
	if option.Action == ActionCloseDialog {
		s.EndDialog()
	}

	return option.Action, nil
}

// GetActiveDialog returns the current dialog text and options.
// Returns empty string and nil if no dialog is active.
func (s *DialogSystem) GetActiveDialog() (text string, options []DialogOption, npcID uint64) {
	if s.activeDialogEntity == 0 {
		return "", nil, 0
	}

	npcEntity, ok := s.world.GetEntity(s.activeDialogEntity)
	if !ok {
		s.activeDialogEntity = 0
		return "", nil, 0
	}

	comp, ok := npcEntity.GetComponent("dialog")
	if !ok {
		return "", nil, 0
	}
	dialogComp, ok := comp.(*DialogComponent)
	if !ok {
		return "", nil, 0
	}

	if !dialogComp.IsActive {
		return "", nil, 0
	}

	return dialogComp.CurrentDialog, dialogComp.Options, s.activeDialogEntity
}

// IsDialogActive returns true if there's an active dialog.
func (s *DialogSystem) IsDialogActive() bool {
	return s.activeDialogEntity != 0
}

// GetActiveDialogEntity returns the entity ID of the NPC in active dialog (0 if none).
func (s *DialogSystem) GetActiveDialogEntity() uint64 {
	return s.activeDialogEntity
}

// Update processes dialog system state each frame.
// This method is reserved for future use (timed dialogs, auto-close, etc).
func (s *DialogSystem) Update(deltaTime float64) {
	// Future enhancement: auto-close dialogs after timeout
	// Future enhancement: update dynamic dialog content based on game state
}
