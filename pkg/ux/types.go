package ux

import (
	"time"
)

// JourneyType identifies a specific user experience journey.
type JourneyType string

const (
	JourneyNewPlayer           JourneyType = "new_player"
	JourneyCrafter             JourneyType = "crafter"
	JourneySocial              JourneyType = "social"
	JourneyExplorer            JourneyType = "explorer"
	JourneyTrader              JourneyType = "trader"
	JourneyBuilder             JourneyType = "builder"
	JourneyRaider              JourneyType = "raider"
	JourneyPvPer               JourneyType = "pvper"
	JourneyQuester             JourneyType = "quester"
	JourneyCompanionOwner      JourneyType = "companion_owner"
	JourneyVehicleUser         JourneyType = "vehicle_user"
	JourneyStoryteller         JourneyType = "storyteller"
	JourneyPrestigePlayer      JourneyType = "prestige_player"
	JourneyGuildLeader         JourneyType = "guild_leader"
	JourneyModder              JourneyType = "modder"
	JourneyCrossServerTraveler JourneyType = "cross_server_traveler"
	JourneyLegendaryQuester    JourneyType = "legendary_quester"
	JourneyHousingDecorator    JourneyType = "housing_decorator"
	JourneySiegeParticipant    JourneyType = "siege_participant"
	JourneyEconomyTycoon       JourneyType = "economy_tycoon"
)

// JourneyDefinition describes a user experience journey.
type JourneyDefinition struct {
	Type             JourneyType
	Name             string
	Description      string
	ExpectedDuration time.Duration
	Steps            []JourneyStep
	RequiredFeatures []string
}

// JourneyStep is a single action in a user journey.
type JourneyStep struct {
	Name        string
	Description string
	Action      func(*JourneyContext) error
}

// JourneyContext holds the state during journey execution.
type JourneyContext struct {
	PlayerID      int
	WorldSeed     int64
	StepIndex     int
	StepStartTime time.Time
	Data          map[string]interface{}
}

// JourneyResult contains the outcome of a journey validation.
type JourneyResult struct {
	Type            JourneyType
	Name            string
	Passed          bool
	CompletionRate  float64
	AverageDuration time.Duration
	Satisfaction    float64
	ErrorRate       float64
	Steps           []StepResult
	Error           error
}

// StepResult contains the outcome of a single step.
type StepResult struct {
	Name      string
	Completed bool
	Duration  time.Duration
	Error     error
}

// ValidationConfig configures journey validation behavior.
type ValidationConfig struct {
	Runs                 int
	TimeTolerancePercent float64
	MinCompletionRate    float64
	MinSatisfaction      float64
	MaxErrorRate         float64
}

// DefaultValidationConfig returns the standard validation configuration.
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		Runs:                 10,
		TimeTolerancePercent: 20.0,
		MinCompletionRate:    0.90,
		MinSatisfaction:      0.80,
		MaxErrorRate:         0.05,
	}
}
