// Package trade_routes implements automated AI merchant caravan systems for cross-server trading.
//
// Phase 57.3: Automated Trade Routes
//
// This package provides:
// - AI merchant caravans with NPC-controlled vehicle fleets
// - Route optimization calculating profitable trade paths
// - Risk/reward mechanics (dangerous routes offer higher profits)
// - Player escort missions protecting caravans
// - Procedural bandit attacks threatening shipments
// - Guild sponsorship for regional price manipulation
//
// Integration Points:
// - V4 Vehicles: pkg/procgen/vehicle (caravan vehicles)
// - V6 Federation Market: pkg/network/federation/market.go (pricing)
// - V4 AI: pkg/engine/ai_system.go (merchant pathfinding)
//
// Performance: <1ms per route calculation, <50 active routes per server
package trade_routes

import (
	"time"
)

// TradeRoute represents an automated merchant caravan route between regions.
type TradeRoute struct {
	// ID is the unique route identifier
	ID string

	// Name is the procedurally generated route name
	Name string

	// StartRegion is the origin region (server ID)
	StartRegion string

	// EndRegion is the destination region (server ID)
	EndRegion string

	// CaravanID is the vehicle fleet entity ID
	CaravanID uint64

	// Status represents the current state of the route
	Status RouteStatus

	// Cargo holds the items being transported
	Cargo []CargoItem

	// ProfitMargin is the expected profit percentage (10-50%)
	ProfitMargin float64

	// DangerLevel represents route risk (0.0-1.0)
	DangerLevel float64

	// Progress is the percentage complete (0.0-1.0)
	Progress float64

	// TravelTime is the real-time duration (30 minutes per region)
	TravelTime time.Duration

	// StartTime is when the caravan departed
	StartTime time.Time

	// EstimatedArrival is the expected completion time
	EstimatedArrival time.Time

	// EscortPlayers are the player entity IDs providing protection
	EscortPlayers []uint64

	// SponsoringGuild is the guild ID funding this route (optional)
	SponsoringGuild string

	// BanditAttacks is the count of attacks encountered
	BanditAttacks int

	// SuccessRate is the historical success rate (0.0-1.0)
	SuccessRate float64
}

// RouteStatus represents the current state of a trade route.
type RouteStatus int

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

// String returns the string representation of route status.
func (s RouteStatus) String() string {
	switch s {
	case StatusPlanning:
		return "Planning"
	case StatusActive:
		return "Active"
	case StatusUnderAttack:
		return "Under Attack"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// CargoItem represents a single item in the caravan cargo.
type CargoItem struct {
	// ItemID is the item identifier
	ItemID string

	// ItemName is the display name
	ItemName string

	// Quantity is the number of items
	Quantity int

	// PurchasePrice is the cost at origin
	PurchasePrice float64

	// TargetPrice is the expected sale price
	TargetPrice float64

	// Profit is the net profit (TargetPrice - PurchasePrice - ShippingCost)
	Profit float64
}

// RouteOptimization holds calculated route efficiency metrics.
type RouteOptimization struct {
	// Route is the trade route being optimized
	Route *TradeRoute

	// ExpectedProfit is the total profit in gold
	ExpectedProfit float64

	// RiskAdjustedProfit is profit adjusted for danger level
	RiskAdjustedProfit float64

	// ServerHops is the number of cross-server transitions
	ServerHops int

	// EstimatedTravelTime is the calculated duration
	EstimatedTravelTime time.Duration

	// DangerZones are the high-risk segments
	DangerZones []DangerZone

	// AlternateRoutes are safer/slower options
	AlternateRoutes []*TradeRoute
}

// DangerZone represents a high-risk segment of a trade route.
type DangerZone struct {
	// RegionID is the server/region with high danger
	RegionID string

	// DangerLevel is the risk factor (0.0-1.0)
	DangerLevel float64

	// BanditSpawnRate is attacks per hour (0.1-1.0)
	BanditSpawnRate float64

	// RecommendedEscorts is the suggested number of player escorts
	RecommendedEscorts int

	// Description is a flavor text explanation
	Description string
}

// BanditEncounter represents a procedural attack on a caravan.
type BanditEncounter struct {
	// ID is the unique encounter identifier
	ID string

	// RouteID is the affected trade route
	RouteID string

	// LocationX is the attack X coordinate
	LocationX float64

	// LocationY is the attack Y coordinate
	LocationY float64

	// BanditCount is the number of hostile entities
	BanditCount int

	// BanditStrength is the aggregate power level (0.0-1.0)
	BanditStrength float64

	// DefenseStrength is the caravan+escorts power (0.0-1.0)
	DefenseStrength float64

	// StartTime is when the attack began
	StartTime time.Time

	// Duration is the combat length
	Duration time.Duration

	// Outcome is the result of the encounter
	Outcome EncounterOutcome

	// LostCargo are items stolen by bandits
	LostCargo []CargoItem

	// PlayerRewards are bonuses for successful defense
	PlayerRewards map[uint64]float64 // PlayerID -> Gold reward
}

// EncounterOutcome represents the result of a bandit attack.
type EncounterOutcome int

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

// String returns the string representation of encounter outcome.
func (o EncounterOutcome) String() string {
	switch o {
	case OutcomePending:
		return "Pending"
	case OutcomeDefended:
		return "Defended"
	case OutcomeCompromised:
		return "Compromised"
	case OutcomeDestroyed:
		return "Destroyed"
	case OutcomeEvaded:
		return "Evaded"
	default:
		return "Unknown"
	}
}

// EscortMission represents a player quest to protect a caravan.
type EscortMission struct {
	// ID is the unique mission identifier
	ID string

	// RouteID is the trade route being escorted
	RouteID string

	// PlayerID is the accepting player
	PlayerID uint64

	// Reward is the gold payment for completion
	Reward float64

	// BonusReward is additional gold for defeating attacks
	BonusReward float64

	// Status is the current mission state
	Status MissionStatus

	// AcceptedAt is when the player took the mission
	AcceptedAt time.Time

	// CompletedAt is when the mission finished
	CompletedAt time.Time
}

// MissionStatus represents the state of an escort mission.
type MissionStatus int

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

// String returns the string representation of mission status.
func (s MissionStatus) String() string {
	switch s {
	case MissionAvailable:
		return "Available"
	case MissionActive:
		return "Active"
	case MissionCompleted:
		return "Completed"
	case MissionFailed:
		return "Failed"
	case MissionAbandoned:
		return "Abandoned"
	default:
		return "Unknown"
	}
}

// GuildSponsorship represents guild funding for trade manipulation.
type GuildSponsorship struct {
	// GuildID is the sponsoring guild
	GuildID string

	// RouteID is the funded trade route
	RouteID string

	// FundingAmount is the gold invested
	FundingAmount float64

	// TargetPriceChange is the desired market impact (-0.5 to +0.5)
	TargetPriceChange float64

	// ActiveRoutes is the count of simultaneous routes
	ActiveRoutes int

	// TotalProfit is the cumulative profit from all routes
	TotalProfit float64

	// StartDate is when sponsorship began
	StartDate time.Time

	// EndDate is when sponsorship expires (optional)
	EndDate time.Time
}
