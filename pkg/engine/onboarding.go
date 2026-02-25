// Package engine provides onboarding flow management for new players.
// This file implements OnboardingManager which coordinates the three tutorial layers:
// CharacterCreationTutorial, EbitenTutorialSystem, and TutorialManager.
package engine

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// OnboardingState represents the current phase of the onboarding sequence.
type OnboardingState int

const (
	// StateCharacterCreation - Player is in character creation with tutorial overlay
	StateCharacterCreation OnboardingState = iota
	// StateInGameTutorial - Player has created character, in-game tutorial active
	StateInGameTutorial
	// StateContextHelp - In-game tutorial complete, context-sensitive help active
	StateContextHelp
	// StateComplete - All tutorials completed
	StateComplete
)

// String returns a human-readable name for the onboarding state.
func (s OnboardingState) String() string {
	switch s {
	case StateCharacterCreation:
		return "CharacterCreation"
	case StateInGameTutorial:
		return "InGameTutorial"
	case StateContextHelp:
		return "ContextHelp"
	case StateComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// OnboardingCompletionCallback is called when an onboarding phase completes.
type OnboardingCompletionCallback func(state OnboardingState)

// OnboardingManager coordinates transitions between tutorial layers.
// It manages the canonical onboarding sequence:
// 1. CharacterCreationTutorial overlays character creation UI
// 2. Character creation completes → player entity spawned
// 3. EbitenTutorialSystem activates for in-game guidance
// 4. TutorialManager provides context-sensitive help on first encounter
type OnboardingManager struct {
	mu               sync.RWMutex
	currentState     OnboardingState
	enabled          bool
	skipped          bool
	playerClass      CharacterClass
	playerEntity     *Entity // Player entity for saving onboarding state (Phase 3.2)
	completionCb     OnboardingCompletionCallback
	stateChangeCb    OnboardingCompletionCallback
	logger           *logrus.Entry
	tutorialSystem   *EbitenTutorialSystem
	creationTutorial *CharacterCreationTutorial
}

// NewOnboardingManager creates a new onboarding manager.
func NewOnboardingManager(logger *logrus.Entry) *OnboardingManager {
	return &OnboardingManager{
		currentState: StateCharacterCreation,
		enabled:      true,
		logger:       logger,
	}
}

// SetTutorialSystem connects the in-game tutorial system.
// Wires an OnCompleteCallback so the onboarding state machine advances
// when the in-game tutorial completes or is skipped.
func (om *OnboardingManager) SetTutorialSystem(ts *EbitenTutorialSystem) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.tutorialSystem = ts

	// Wire completion callback: when in-game tutorial finishes,
	// transition through ContextHelp to Complete.
	if ts != nil {
		ts.OnCompleteCallback = func() {
			if om.GetState() == StateInGameTutorial {
				om.TransitionToContextHelp()
				om.Complete()
			} else if om.GetState() == StateCharacterCreation {
				// Tutorial was skipped before character creation finished;
				// advance the full chain so the state machine doesn't get stuck.
				om.TransitionToInGameTutorial()
				om.TransitionToContextHelp()
				om.Complete()
			}
		}
	}
}

// SetCreationTutorial connects the character creation tutorial.
func (om *OnboardingManager) SetCreationTutorial(cct *CharacterCreationTutorial) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.creationTutorial = cct
}

// SetPlayerEntity sets the player entity for persisting onboarding state.
// Phase 3.2: Enables save/load to preserve onboarding progress via TutorialCompletionComponent.
func (om *OnboardingManager) SetPlayerEntity(player *Entity) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.playerEntity = player

	// Sync current state to player entity on attachment
	if player != nil {
		UpdateOnboardingState(player, om.currentState, om.playerClass)
	}

	if om.logger != nil {
		om.logger.WithField("player_attached", player != nil).Debug("player entity set for onboarding")
	}
}

// SetCompletionCallback sets the callback called when onboarding fully completes.
func (om *OnboardingManager) SetCompletionCallback(cb OnboardingCompletionCallback) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.completionCb = cb
}

// SetStateChangeCallback sets the callback called on any state transition.
func (om *OnboardingManager) SetStateChangeCallback(cb OnboardingCompletionCallback) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.stateChangeCb = cb
}

// GetState returns the current onboarding state.
func (om *OnboardingManager) GetState() OnboardingState {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.currentState
}

// IsEnabled returns whether onboarding is active.
func (om *OnboardingManager) IsEnabled() bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.enabled
}

// IsComplete returns whether onboarding has finished.
func (om *OnboardingManager) IsComplete() bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.currentState == StateComplete || om.skipped
}

// SetEnabled enables or disables the entire onboarding flow.
// When disabled, all tutorial layers are disabled.
// When re-enabled, child systems are restored to the correct state for the
// current onboarding phase (e.g. in-game tutorial stays disabled during
// character creation). Completed or skipped onboarding is never re-activated.
func (om *OnboardingManager) SetEnabled(enabled bool) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.enabled = enabled

	if !enabled {
		// Disable all connected tutorial systems
		if om.creationTutorial != nil {
			om.creationTutorial.SetEnabled(false)
		}
		if om.tutorialSystem != nil {
			om.tutorialSystem.Enabled = false
			om.tutorialSystem.ShowUI = false
		}
	} else if !om.skipped && om.currentState != StateComplete {
		// Re-enable child systems appropriate for the current onboarding phase.
		// This prevents premature activation of the in-game tutorial during
		// character creation (fixes ApplySettings/SetAudioManager race).
		switch om.currentState {
		case StateCharacterCreation:
			if om.creationTutorial != nil {
				om.creationTutorial.SetEnabled(true)
			}
			// In-game tutorial is not yet active in this phase
			if om.tutorialSystem != nil {
				om.tutorialSystem.Enabled = false
				om.tutorialSystem.ShowUI = false
			}
		case StateInGameTutorial:
			if om.creationTutorial != nil {
				om.creationTutorial.SetEnabled(false)
			}
			if om.tutorialSystem != nil {
				om.tutorialSystem.Enabled = true
				om.tutorialSystem.ShowUI = true
			}
		case StateContextHelp:
			// Both tutorials are past; contextual help is independent
			if om.creationTutorial != nil {
				om.creationTutorial.SetEnabled(false)
			}
			if om.tutorialSystem != nil {
				om.tutorialSystem.Enabled = false
				om.tutorialSystem.ShowUI = false
			}
		}
	}
	// When skipped or complete, re-enabling does not re-activate finished tutorials.

	if om.logger != nil {
		om.logger.WithField("enabled", enabled).Debug("onboarding enabled state changed")
	}
}

// SetPlayerClass stores the selected class for class-aware tutorial content.
// Propagates the class to the in-game tutorial system for class-specific hints.
func (om *OnboardingManager) SetPlayerClass(class CharacterClass) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.playerClass = class

	// Propagate class to tutorial system for class-aware content (Phase 3.4)
	if om.tutorialSystem != nil {
		om.tutorialSystem.SetPlayerClass(class)
	}

	if om.logger != nil {
		om.logger.WithField("class", class.String()).Debug("player class set for onboarding")
	}
}

// GetPlayerClass returns the stored player class.
func (om *OnboardingManager) GetPlayerClass() CharacterClass {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.playerClass
}

// SkipAll skips the entire onboarding flow.
// Child systems are disabled outside the lock to avoid a deadlock:
// DisableTutorial() fires OnCompleteCallback → GetState() → RLock while
// the write lock is already held by this method.
func (om *OnboardingManager) SkipAll() {
	om.mu.Lock()
	om.skipped = true
	om.currentState = StateComplete

	// Capture references before releasing the lock.
	creationTutorial := om.creationTutorial
	tutorialSystem := om.tutorialSystem
	logger := om.logger
	om.mu.Unlock()

	// Disable child systems outside the lock.
	if creationTutorial != nil {
		creationTutorial.SkipTutorial()
	}
	if tutorialSystem != nil {
		// Directly disable rather than calling DisableTutorial() to avoid
		// re-entering the OnCompleteCallback (state is already Complete).
		tutorialSystem.Enabled = false
		tutorialSystem.ShowUI = false
	}

	if logger != nil {
		logger.Debug("onboarding skipped entirely")
	}
}

// TransitionToInGameTutorial moves from character creation to in-game tutorial.
// This should be called when character creation completes.
func (om *OnboardingManager) TransitionToInGameTutorial() {
	om.mu.Lock()

	if !om.enabled || om.skipped {
		om.mu.Unlock()
		return
	}

	previousState := om.currentState
	om.currentState = StateInGameTutorial

	// Activate the in-game tutorial system
	if om.tutorialSystem != nil {
		om.tutorialSystem.Enabled = true
		om.tutorialSystem.ShowUI = true
	}

	// Persist onboarding state to player entity (Phase 3.2)
	if om.playerEntity != nil {
		UpdateOnboardingState(om.playerEntity, om.currentState, om.playerClass)
	}

	// Capture callbacks before unlocking
	stateChangeCb := om.stateChangeCb
	logger := om.logger

	om.mu.Unlock()

	// Call state change callback outside of lock
	if stateChangeCb != nil {
		stateChangeCb(om.currentState)
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"previous_state": previousState.String(),
			"current_state":  om.currentState.String(),
		}).Debug("transitioned to in-game tutorial")
	}
}

// TransitionToContextHelp moves from in-game tutorial to context-sensitive help.
// This should be called when the in-game tutorial completes.
func (om *OnboardingManager) TransitionToContextHelp() {
	om.mu.Lock()

	if !om.enabled || om.skipped {
		om.mu.Unlock()
		return
	}

	previousState := om.currentState
	om.currentState = StateContextHelp

	// Persist onboarding state to player entity (Phase 3.2)
	if om.playerEntity != nil {
		UpdateOnboardingState(om.playerEntity, om.currentState, om.playerClass)
	}

	// Capture callbacks before unlocking
	stateChangeCb := om.stateChangeCb
	logger := om.logger

	om.mu.Unlock()

	// Call state change callback outside of lock
	if stateChangeCb != nil {
		stateChangeCb(om.currentState)
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"previous_state": previousState.String(),
			"current_state":  om.currentState.String(),
		}).Debug("transitioned to context-sensitive help")
	}
}

// Complete marks the onboarding flow as finished.
// This should be called when context-sensitive help has been viewed
// or after a timeout period.
func (om *OnboardingManager) Complete() {
	om.mu.Lock()

	if om.currentState == StateComplete {
		om.mu.Unlock()
		return
	}

	previousState := om.currentState
	om.currentState = StateComplete

	// Persist onboarding state to player entity (Phase 3.2)
	if om.playerEntity != nil {
		UpdateOnboardingState(om.playerEntity, om.currentState, om.playerClass)
	}

	// Capture callbacks before unlocking
	completionCb := om.completionCb
	stateChangeCb := om.stateChangeCb
	logger := om.logger

	om.mu.Unlock()

	// Call callbacks outside of lock
	if stateChangeCb != nil {
		stateChangeCb(StateComplete)
	}
	if completionCb != nil {
		completionCb(StateComplete)
	}

	if logger != nil {
		logger.WithField("previous_state", previousState.String()).Debug("onboarding completed")
	}
}

// ExportState serializes the onboarding state for save/load.
func (om *OnboardingManager) ExportState() OnboardingStateData {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return OnboardingStateData{
		CurrentState: int(om.currentState),
		Enabled:      om.enabled,
		Skipped:      om.skipped,
		PlayerClass:  int(om.playerClass),
	}
}

// ImportState restores onboarding state from saved data.
func (om *OnboardingManager) ImportState(data OnboardingStateData) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.currentState = OnboardingState(data.CurrentState)
	om.enabled = data.Enabled
	om.skipped = data.Skipped
	om.playerClass = CharacterClass(data.PlayerClass)

	// Update connected systems based on loaded state
	if om.currentState >= StateInGameTutorial && om.tutorialSystem != nil {
		om.tutorialSystem.Enabled = om.enabled && om.currentState == StateInGameTutorial
		om.tutorialSystem.ShowUI = om.tutorialSystem.Enabled
	}
	if om.currentState >= StateCharacterCreation && om.creationTutorial != nil {
		// Creation tutorial should be disabled if we're past that phase
		if om.currentState > StateCharacterCreation {
			om.creationTutorial.SetEnabled(false)
		}
	}

	if om.logger != nil {
		om.logger.WithFields(logrus.Fields{
			"state":   om.currentState.String(),
			"enabled": om.enabled,
			"skipped": om.skipped,
		}).Debug("onboarding state imported")
	}
}

// OnboardingStateData holds serializable onboarding state.
type OnboardingStateData struct {
	CurrentState int  `json:"current_state"`
	Enabled      bool `json:"enabled"`
	Skipped      bool `json:"skipped"`
	PlayerClass  int  `json:"player_class"`
}
