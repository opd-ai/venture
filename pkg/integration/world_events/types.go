package world_events

import (
	"fmt"
	"time"
)

// EventType represents the category of world event.
type EventType string

const (
	EventGuildWarfare    EventType = "guild_warfare"
	EventFactionResponse EventType = "faction_response"
	EventEconomic        EventType = "economic"
	EventWeatherDisaster EventType = "weather_disaster"
	EventCrossServer     EventType = "cross_server"
	EventChained         EventType = "chained"
)

// TriggerType represents the action that spawned an event.
type TriggerType string

const (
	TriggerGuildWar       TriggerType = "guild_war"
	TriggerGuildPeace     TriggerType = "guild_peace"
	TriggerTradeVolume    TriggerType = "trade_volume"
	TriggerPoliticalShift TriggerType = "political_shift"
	TriggerWeatherChange  TriggerType = "weather_change"
	TriggerPlayerChoice   TriggerType = "player_choice"
)

// Severity represents the magnitude of an event's impact.
type Severity int

const (
	SeverityMinor    Severity = 1
	SeverityModerate Severity = 2
	SeverityMajor    Severity = 3
	SeverityCritical Severity = 4
)

// ImpactType represents what aspect of the world is affected.
type ImpactType string

const (
	ImpactNPCReputation ImpactType = "npc_reputation"
	ImpactPriceChange   ImpactType = "price_change"
	ImpactWeather       ImpactType = "weather"
	ImpactTerrain       ImpactType = "terrain"
	ImpactSpawnRate     ImpactType = "spawn_rate"
)

// WorldEvent represents a dynamic world event.
type WorldEvent struct {
	ID          string
	Type        EventType
	Trigger     TriggerType
	Severity    Severity
	Title       string
	Description string
	Location    string
	ServerID    string
	StartTime   time.Time
	Duration    time.Duration
	Impacts     []Impact
	ChainEvents []string
	Permanent   bool
	// CenterX is the X coordinate of the event's center location.
	CenterX float64
	// CenterY is the Y coordinate of the event's center location.
	CenterY float64
}

// Impact represents a specific effect of an event.
type Impact struct {
	Type     ImpactType
	Target   string
	Modifier float64
	Duration time.Duration
}

// GetImpacts returns the impacts of this event.
func (e *WorldEvent) GetImpacts() []Impact {
	return e.Impacts
}

// IsActive checks if the event is currently active.
func (e *WorldEvent) IsActive(currentTime time.Time) bool {
	if e.Permanent {
		return true
	}
	endTime := e.StartTime.Add(e.Duration)
	return currentTime.After(e.StartTime) && currentTime.Before(endTime)
}

// FactionResponse represents NPC faction reaction to player actions.
type FactionResponse struct {
	FactionID        string
	ResponseType     string
	ReputationChange float64
	HostilityChange  float64
	TradeBonus       float64
	Message          string
}

// EconomicEvent represents a market or economic change.
type EconomicEvent struct {
	EventID       string
	ItemType      string
	PriceModifier float64
	SupplyChange  int
	Duration      time.Duration
	AffectedZones []string
}

// WeatherDisaster represents a severe weather event.
type WeatherDisaster struct {
	DisasterType string
	Intensity    float64
	Radius       float64
	CenterX      float64
	CenterY      float64
	Duration     time.Duration
	Damage       float64
}

// EventChain represents a sequence of linked events.
type EventChain struct {
	ChainID       string
	Events        []string
	CurrentIndex  int
	BranchChoices map[string][]string
}

// TriggerParams holds parameters for event generation.
type TriggerParams struct {
	TriggerType TriggerType
	Severity    Severity
	Location    string
	ServerID    string
	GuildID     string
	FactionID   string
	ItemType    string
	PlayerID    string
	ChoiceID    string
	Metadata    map[string]interface{}
	// CenterX is the X coordinate for the event's center location.
	CenterX float64
	// CenterY is the Y coordinate for the event's center location.
	CenterY float64
}

// Validate checks if trigger parameters are valid.
func (p *TriggerParams) Validate() error {
	if p.TriggerType == "" {
		return fmt.Errorf("trigger type required")
	}
	if p.Severity < SeverityMinor || p.Severity > SeverityCritical {
		return fmt.Errorf("severity must be 1-4, got %d", p.Severity)
	}
	if p.Location == "" {
		return fmt.Errorf("location required")
	}
	if p.ServerID == "" {
		return fmt.Errorf("server ID required")
	}
	return nil
}

// EventManagerConfig holds configuration for the event manager.
type EventManagerConfig struct {
	MaxActiveEvents      int
	EventFrequency       float64
	ChainProbability     float64
	CrossServerPropDelay time.Duration
	ResponseTimeMin      time.Duration
	ResponseTimeMax      time.Duration
}

// NewDefaultEventManagerConfig returns default configuration.
func NewDefaultEventManagerConfig() EventManagerConfig {
	return EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     0.3,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      1 * time.Minute,
		ResponseTimeMax:      5 * time.Minute,
	}
}
