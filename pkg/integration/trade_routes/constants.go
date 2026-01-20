// File: constants.go
// Purpose: Centralized constants for trade route enums
//
// This file contains all constant definitions for the trade_routes package,
// extracted from types.go during reorganization for improved navigability.
package trade_routes

// Route status constants
// Originally defined in: types.go
const (
	// StatusPlanning indicates route is being calculated
	StatusPlanning RouteStatus = iota

	// StatusActive indicates caravan is traveling
	StatusActive

	// StatusUnderAttack indicates bandits are engaging
	StatusUnderAttack

	// StatusCompleted indicates successful arrival
	StatusCompleted

	// StatusFailed indicates route failed (attacked, vehicle destroyed)
	StatusFailed

	// StatusCancelled indicates route was manually cancelled
	StatusCancelled
)

// Encounter outcome constants
// Originally defined in: types.go
const (
	// OutcomePending indicates combat is ongoing
	OutcomePending EncounterOutcome = iota

	// OutcomeDefended indicates successful defense
	OutcomeDefended

	// OutcomeCompromised indicates partial cargo loss
	OutcomeCompromised

	// OutcomeDestroyed indicates total caravan loss
	OutcomeDestroyed

	// OutcomeEvaded indicates successful escape
	OutcomeEvaded
)

// Mission status constants
// Originally defined in: types.go
const (
	// MissionAvailable indicates mission can be accepted
	MissionAvailable MissionStatus = iota

	// MissionActive indicates player is escorting
	MissionActive

	// MissionCompleted indicates successful completion
	MissionCompleted

	// MissionFailed indicates route failed
	MissionFailed

	// MissionAbandoned indicates player cancelled
	MissionAbandoned
)
